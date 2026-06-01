package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: rixi <command>")
		fmt.Println("Commands:")
		fmt.Println("  create <name>   scaffold a new Go project")
		fmt.Println("  generate <name>  generate model, view, and controller for a resource")
		fmt.Println("  model <name>    generate a model in the current project")
		fmt.Println("  view <name>     generate a view with templates for a model")
		fmt.Println("  controller <name>  generate a controller that uses model and view")
		fmt.Println("  serve           run the project (go run .)")
		fmt.Println("  dev             run with auto-reload on file changes")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "create":
		if len(os.Args) < 3 {
			fmt.Println("Usage: rixi create <project-name>")
			os.Exit(1)
		}
		createProject(os.Args[2])
	case "generate":
		if len(os.Args) < 3 {
			fmt.Println("Usage: rixi generate <name>")
			os.Exit(1)
		}
		generate(os.Args[2])
	case "model":
		if len(os.Args) < 3 {
			fmt.Println("Usage: rixi model <name>")
			os.Exit(1)
		}
		generateModel(os.Args[2])
	case "view":
		if len(os.Args) < 3 {
			fmt.Println("Usage: rixi view <name>")
			os.Exit(1)
		}
		generateView(os.Args[2])
	case "controller":
		if len(os.Args) < 3 {
			fmt.Println("Usage: rixi controller <name>")
			os.Exit(1)
		}
		generateController(os.Args[2])
	case "serve":
		serve()
	case "dev":
		dev()
	default:
		fmt.Printf("unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
