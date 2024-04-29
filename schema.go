package main

import (
	"fmt"
	"os"
	"strings"

	_ "embed"

	"github.com/authzed/spicedb/pkg/namespace"
	core "github.com/authzed/spicedb/pkg/proto/core/v1"
	"github.com/authzed/spicedb/pkg/schemadsl/compiler"
	"github.com/authzed/spicedb/pkg/schemadsl/generator"
	"github.com/authzed/spicedb/pkg/schemadsl/input"
	"gopkg.in/yaml.v3"
)

type schema struct {
	definitions   map[string]*core.NamespaceDefinition
	permissionSet map[string]bool
	raw           []compiler.SchemaDefinition
	yml           map[string]interface{}
}

func load(f string) (*schema, error) {
	reader, err := os.OpenFile(f, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}

	file := make(map[string]interface{})
	dec := yaml.NewDecoder(reader)
	err = dec.Decode(file)
	if err != nil {
		return nil, err
	}

	dsl := file["schema"].(string)
	source := input.Source(f)

	compiled, err := compiler.Compile(compiler.InputSchema{Source: source, SchemaString: dsl}, compiler.AllowUnprefixedObjectType())
	if err != nil {
		return nil, err
	}

	definitions := make(map[string]*core.NamespaceDefinition)
	for _, ns := range compiled.ObjectDefinitions {
		definitions[ns.Name] = ns
	}

	perms := make(map[string]bool)
	role := definitions["role"]
	for _, rel := range role.Relation {
		perms[rel.Name] = true
	}

	return &schema{
		definitions:   definitions,
		permissionSet: perms,
		raw:           compiled.OrderedDefinitions,
		yml:           file,
	}, nil
}

func (s *schema) store(f string) error {
	source, _, err := generator.GenerateSchema(s.raw)
	if err != nil {
		return err
	}

	s.yml["schema"] = source

	writer, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(0644))
	if err != nil {
		return err
	}

	enc := yaml.NewEncoder(writer)
	return enc.Encode(s.yml)
}

func (s *schema) hasPermission(permission string) bool {
	return s.permissionSet[permission]
}

func (s *schema) addPermission(permission string) {
	if s.permissionSet[permission] {
		return
	}

	role := s.definitions["role"]
	role.Relation = append(role.Relation, namespace.MustRelation(permission, nil, namespace.AllowedRelation("user:*", compiler.Ellipsis)))

	rolebinding := s.definitions["role_binding"]
	rolebinding.Relation = append(rolebinding.Relation, namespace.MustRelation(permission, namespace.Intersection(namespace.ComputedUserset("subject"), namespace.TupleToUserset("granted", permission))))

	workspace := s.definitions["workspace"]
	workspace.Relation = append(workspace.Relation, namespace.MustRelation(permission, namespace.Union(namespace.TupleToUserset("user_grant", permission), namespace.TupleToUserset("parent", permission))))

	s.permissionSet[permission] = true
}

func (s *schema) addDefinition(definition *core.NamespaceDefinition) {
	s.definitions[definition.Name] = definition
	s.raw = append(s.raw, definition)
}

func (s *schema) addOrExtendResourceType(app string, resourceType string, permissions []string) {
	app = cleanNameForSchemaCompatibility(app)
	resourceType = cleanNameForSchemaCompatibility(resourceType)

	definitionName := fmt.Sprintf("%s/%s", app, resourceType)
	definition, ok := s.definitions[definitionName]
	if !ok {
		definition = namespace.Namespace(definitionName)
		definition.Relation = append(definition.Relation, namespace.MustRelation("workspace", nil, namespace.AllowedRelation("workspace", compiler.Ellipsis)))
		s.addDefinition(definition)
	}

	appWildcard := fmt.Sprintf("%s_all_all", app)
	s.addPermission(appWildcard)
	resourceWildcard := fmt.Sprintf("%s_%s_all", app, resourceType)
	s.addPermission(resourceWildcard)

	for _, permission := range permissions {
		permission = cleanNameForSchemaCompatibility(permission)
		qualifiedPermission := fmt.Sprintf("%s_%s_%s", app, resourceType, permission)
		if s.hasPermission(qualifiedPermission) {
			continue
		}

		s.addPermission(qualifiedPermission)
		permissionWildcard := fmt.Sprintf("%s_all_%s", app, permission)
		s.addPermission(permissionWildcard)

		definition.Relation = append(definition.Relation, namespace.MustRelation(permission, namespace.Union(
			namespace.TupleToUserset("workspace", qualifiedPermission),
			namespace.TupleToUserset("workspace", permissionWildcard),
			namespace.TupleToUserset("workspace", resourceWildcard),
			namespace.TupleToUserset("workspace", appWildcard),
		)))
	}
}

func (s *schema) addWildcardResourceType(app string, wildcard *Resource) {
	app = cleanNameForSchemaCompatibility(app)
	s.addPermission(fmt.Sprintf("%s_all_all", app))
	for _, permission := range wildcard.Permissions {
		if permission == "*" {
			continue
		}

		s.addPermission(fmt.Sprintf("%s_all_%s", app, cleanNameForSchemaCompatibility(permission)))
	}
}

func cleanNameForSchemaCompatibility(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "*", "all")

	return name
}

//go:embed empty_bootstrap.yaml
var emptySchema string

func storeEmpySchema(path string) error {
	return os.WriteFile(path, []byte(emptySchema), 0644)
}
