package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func addResourcePermissions(definition string, permissions []string) error {
	s, err := load(inputSchemaFile)
	if err != nil {
		return err
	}

	app, resourceType, err := splitDefinition(definition)
	if err != nil {
		return err
	}

	s.addOrExtendResourceType(app, resourceType, permissions)

	return s.store(outputSchemaFile)
}

func importRBACService(rbacPath, svcName string) error {
	s, err := load(inputSchemaFile)
	if err != nil {
		return err
	}

	service, err := LoadService(rbacPath, svcName)
	if err != nil {
		return err
	}

	for _, resource := range service.Resources {
		s.addOrExtendResourceType(service.Name, resource.Name, resource.Permissions)
	}

	return s.store(outputSchemaFile)
}

func createNewFile(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	_, err = os.Stat(abs)
	if os.IsNotExist(err) { // Expect an error that the file doesn't exist, and if not, create it
		return storeEmpySchema(abs)
	} else if err != nil {
		return err // If there's a different error, something's wrong, pass it through
	} else { // If there's NO error, the file exists, print and abort
		fmt.Printf("Bootstrap file '%s' already exists. Use -output=<path/to/filename> to specify a different path to create a new bootstrap file.\n", abs)
		return nil
	}
}

var inputSchemaFile string
var outputSchemaFile string

func main() {
	addPermissions := flag.NewFlagSet("add-permissions", flag.ExitOnError)
	definition := addPermissions.String("res", "", "The resource type the permission applies to in the form <service_name>/<resource_type>. Will be created if not present.")
	addGlobalParameters(addPermissions)

	importService := flag.NewFlagSet("import-service", flag.ExitOnError)
	rbacPath := importService.String("rbac-config", "", "The path to an rbac-config 'permissions' directory containing service-specific JSON files (ex: configs/prod/permissions)")
	rbacSvc := importService.String("svc", "", "The service to be imported. Must match the name of an rbac-config JSON file.")
	addGlobalParameters(importService)

	newFile := flag.NewFlagSet("new", flag.ExitOnError)
	newFilePath := newFile.String("output", "bootstrap.yaml", "The location to store the empty bootstrap file.")

	if len(os.Args) < 2 {
		print("Please specify a subcommand: new, add-permissions, import-service")
		return
	}

	switch os.Args[1] {
	case "new":
		newFile.Parse(os.Args[2:])
		err := createNewFile(*newFilePath)
		if err != nil {
			fmt.Println(err)
			return
		}
	case "add-permissions":
		addPermissions.Parse(os.Args[2:])
		err := addResourcePermissions(*definition, addPermissions.Args())
		if err != nil {
			fmt.Println(err)
			return
		}
	case "import-service":
		importService.Parse(os.Args[2:])
		err := importRBACService(*rbacPath, *rbacSvc)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func addGlobalParameters(fs *flag.FlagSet) {
	fs.StringVar(&inputSchemaFile, "input", "bootstrap.yaml", "The bootstrap yaml file to load for editing.")
	fs.StringVar(&outputSchemaFile, "output", "bootstrap.yaml", "Where to store the modifed bootstrap yaml. Set to the input path to overwrite.")
}

func splitDefinition(definition string) (app string, resourcetype string, err error) {
	parts := strings.Split(definition, "/")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("resource type not well-formed, should be <service_name>/<resource_type>: %s", definition)
	}

	return parts[0], parts[1], nil
}
