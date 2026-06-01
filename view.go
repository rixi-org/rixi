package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

// ViewConfig holds the template data for generating a view file.
type ViewConfig struct {
	Name string
}

func generateView(name string) {
	// Check we're inside a Go project
	if _, err := os.Stat("go.mod"); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("error: no go.mod found — run this command inside a Rixi project")
			os.Exit(1)
		}
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	// Validate name is non-empty
	if name == "" {
		fmt.Println("error: view name cannot be empty")
		os.Exit(1)
	}

	// PascalCase the name
	viewName := toPascal(name)

	// Create view directory
	viewDir := "view"
	if err := os.MkdirAll(viewDir, 0755); err != nil {
		fmt.Printf("error: could not create view directory: %v\n", err)
		os.Exit(1)
	}

	// Generate shared files (view.go, render.go, base.html) only once
	sharedFiles := []struct {
		FileName string
		TmplPath string
	}{
		{viewDir + "/view.go", "templates/view/view.go.tmpl"},
		{viewDir + "/render.go", "templates/view/render.go.tmpl"},
		{viewDir + "/base.html", "templates/view/base.html.tmpl"},
	}

	for _, f := range sharedFiles {
		if _, err := os.Stat(f.FileName); err == nil {
			continue // already exists, skip
		}
		tmpl, err := template.ParseFS(tmplFS, f.TmplPath)
		if err != nil {
			fmt.Printf("error: failed to parse template %s: %v\n", f.TmplPath, err)
			os.Exit(1)
		}
		out, err := os.Create(f.FileName)
		if err != nil {
			fmt.Printf("error: could not create %s: %v\n", f.FileName, err)
			os.Exit(1)
		}
		if err := tmpl.Execute(out, nil); err != nil {
			out.Close()
			fmt.Printf("error: failed to execute template %s: %v\n", f.TmplPath, err)
			os.Exit(1)
		}
		out.Close()
		fmt.Printf("created %s\n", f.FileName)
	}

	// Generate the per-model view template
	viewPath := viewDir + "/" + strings.ToLower(name) + ".html"
	if _, err := os.Stat(viewPath); err == nil {
		fmt.Printf("error: %s already exists — remove it first or use a different name\n", viewPath)
		os.Exit(1)
	}

	tmpl, err := template.ParseFS(tmplFS, "templates/view/name.html.tmpl")
	if err != nil {
		fmt.Printf("error: failed to parse template: %v\n", err)
		os.Exit(1)
	}

	f, err := os.Create(viewPath)
	if err != nil {
		fmt.Printf("error: could not create %s: %v\n", viewPath, err)
		os.Exit(1)
	}
	defer f.Close()

	if err := tmpl.Execute(f, ViewConfig{Name: viewName}); err != nil {
		fmt.Printf("error: failed to execute template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("created %s\n", viewPath)
}
