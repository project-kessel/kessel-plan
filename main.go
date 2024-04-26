package main

import (
	"flag"
	"fmt"
	"os"
)

func addResourcePermissions(app string, resourceType string, permissions []string) error {
	s, err := load(inputSchemaFile)
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
	return os.WriteFile(path, []byte(baselineSchema), 0600)
}

var inputSchemaFile string
var outputSchemaFile string

func main() {
	addPermissions := flag.NewFlagSet("add-permissions", flag.ExitOnError)
	app := addPermissions.String("svc", "", "The service the permission applies to, used as a prefix in definition names.")
	res := addPermissions.String("res", "", "The resource type the permission applies to. Will be created if not present.")
	addGlobalParameters(addPermissions)

	importService := flag.NewFlagSet("import-service", flag.ExitOnError)
	rbacPath := importService.String("rbac-config", "", "The path to an rbac-config 'permissions' directory containing service-specific JSON files (ex: configs/prod/permissions)")
	rbacSvc := importService.String("svc", "", "The service to be imported. Must match the name of an rbac-config JSON file.")
	addGlobalParameters(importService)

	newFile := flag.NewFlagSet("new", flag.ExitOnError)
	newFilePath := newFile.String("path", "bootstrap.yaml", "The location to store the empty bootstrap file.")

	if len(os.Args) < 2 {
		print("Please specify a subcommand: new, add-permissions, import-service")
		return
	}

	switch os.Args[1] {
	case "new":
		newFile.Parse(os.Args[2:])
		createNewFile(*newFilePath)
	case "add-permissions":
		addPermissions.Parse(os.Args[2:])
		err := addResourcePermissions(*app, *res, addPermissions.Args())
		if err != nil {
			fmt.Print(err)
			return
		}
	case "import-service":
		importService.Parse(os.Args[2:])
		err := importRBACService(*rbacPath, *rbacSvc)
		if err != nil {
			fmt.Print(err)
			return
		}
	}
}

func addGlobalParameters(fs *flag.FlagSet) {
	fs.StringVar(&inputSchemaFile, "input", "bootstrap.yaml", "The bootstrap yaml file to load for editing.")
	fs.StringVar(&outputSchemaFile, "output", "bootstrap.yaml", "Where to store the modifed bootstrap yaml. Set to the input path to overwrite.")
}
