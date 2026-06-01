package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

// ModelConfig holds the template data for generating a model file.
type ModelConfig struct {
	Name string
}

func generateModel(name string) {
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
		fmt.Println("error: model name cannot be empty")
		os.Exit(1)
	}

	// PascalCase the name
	modelName := toPascal(name)

	// Path for the output file
	modelDir := "model"
	modelPath := modelDir + "/" + strings.ToLower(name) + ".go"

	// Check if file already exists
	if _, err := os.Stat(modelPath); err == nil {
		fmt.Printf("error: %s already exists — remove it first or use a different name\n", modelPath)
		os.Exit(1)
	}

	// Create model directory if needed
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		fmt.Printf("error: could not create model directory: %v\n", err)
		os.Exit(1)
	}

	// Parse and execute the template
	tmpl, err := template.ParseFS(tmplFS, "templates/model.go.tmpl")
	if err != nil {
		fmt.Printf("error: failed to parse template: %v\n", err)
		os.Exit(1)
	}

	f, err := os.Create(modelPath)
	if err != nil {
		fmt.Printf("error: could not create %s: %v\n", modelPath, err)
		os.Exit(1)
	}
	defer f.Close()

	if err := tmpl.Execute(f, ModelConfig{Name: modelName}); err != nil {
		fmt.Printf("error: failed to execute template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("created %s\n", modelPath)
}

// toPascal converts a name to PascalCase.
// Examples: "user" → "User", "user_model" → "UserModel", "user-model" → "UserModel"
func toPascal(s string) string {
	// Split on non-alphanumeric boundaries
	parts := splitName(s)

	// Title-case each part and join
	var b strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		r, size := utf8.DecodeRuneInString(part)
		b.WriteRune(unicode.ToUpper(r))
		b.WriteString(strings.ToLower(part[size:]))
	}
	return b.String()
}

// splitName splits on underscore, hyphen, or space.
func splitName(s string) []string {
	var parts []string
	var cur strings.Builder
	for _, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			if cur.Len() > 0 {
				parts = append(parts, cur.String())
				cur.Reset()
			}
		} else {
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
}
