package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/template"
)

// ControllerConfig holds the template data for generating a controller file.
type ControllerConfig struct {
	Name   string
	Module string
}

func generateController(name string) {
	// Check we're inside a Go project
	module, err := readModuleName()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("error: no go.mod found — run this command inside a Rixi project")
		} else {
			fmt.Printf("error: %v\n", err)
		}
		os.Exit(1)
	}

	// Validate name is non-empty
	if name == "" {
		fmt.Println("error: controller name cannot be empty")
		os.Exit(1)
	}

	// PascalCase the name
	ctrlName := toPascal(name)

	// Path for the output file
	ctrlDir := "controller"
	ctrlPath := ctrlDir + "/" + strings.ToLower(name) + ".go"

	// Check if file already exists
	if _, err := os.Stat(ctrlPath); err == nil {
		fmt.Printf("error: %s already exists — remove it first or use a different name\n", ctrlPath)
		os.Exit(1)
	}

	// Create controller directory if needed
	if err := os.MkdirAll(ctrlDir, 0755); err != nil {
		fmt.Printf("error: could not create controller directory: %v\n", err)
		os.Exit(1)
	}

	// Parse and execute the template
	tmpl, err := template.ParseFS(tmplFS, "templates/controller/name.go.tmpl")
	if err != nil {
		fmt.Printf("error: failed to parse template: %v\n", err)
		os.Exit(1)
	}

	f, err := os.Create(ctrlPath)
	if err != nil {
		fmt.Printf("error: could not create %s: %v\n", ctrlPath, err)
		os.Exit(1)
	}
	defer f.Close()

	if err := tmpl.Execute(f, ControllerConfig{Name: ctrlName, Module: module}); err != nil {
		fmt.Printf("error: failed to execute template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("created %s\n", ctrlPath)
}

// readModuleName extracts the module path from go.mod.
func readModuleName() (string, error) {
	f, err := os.Open("go.mod")
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(line[7:]), nil
		}
	}
	return "", fmt.Errorf("module declaration not found in go.mod")
}
