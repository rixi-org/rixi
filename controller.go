package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
	"text/template"
)

// ControllerConfig holds the template data for generating a controller file.
type ControllerConfig struct {
	Name   string
	Module string
	Plural string
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

	if err := tmpl.Execute(f, ControllerConfig{Name: ctrlName, Module: module, Plural: strings.ToLower(name) + "s"}); err != nil {
		fmt.Printf("error: failed to execute template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("created %s\n", ctrlPath)

	// Add route to routes.go
	addRoute(name)
}

// addRoute adds the handler route to routes.go using AST.
func addRoute(name string) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "routes.go", nil, parser.ParseComments)
	if err != nil {
		return
	}

	routePath := "/" + strings.ToLower(name) + "s/"

	// Skip if route already exists
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "registerRoutes" {
			continue
		}
		for _, stmt := range fn.Body.List {
			call, ok := stmt.(*ast.ExprStmt)
			if !ok {
				continue
			}
			callExpr, ok := call.X.(*ast.CallExpr)
			if !ok {
				continue
			}
			if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "HandleFunc" && len(callExpr.Args) >= 1 {
					if lit, ok := callExpr.Args[0].(*ast.BasicLit); ok {
						if lit.Value == fmt.Sprintf(`"%s"`, routePath) {
							return
						}
					}
				}
			}
		}
	}

	// Find registerRoutes function and add route before closing brace
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "registerRoutes" {
			continue
		}

		// Build: mux.HandleFunc("/name/s", controller.NameHandler(db))
		route := &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "mux"},
					Sel: &ast.Ident{Name: "HandleFunc"},
				},
				Args: []ast.Expr{
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: fmt.Sprintf(`"%s"`, routePath),
					},
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "controller"},
							Sel: &ast.Ident{Name: toPascal(name) + "Handler"},
						},
						Args: []ast.Expr{&ast.Ident{Name: "db"}},
					},
				},
			},
		}

		// Insert at the end (before closing brace)
		fn.Body.List = append(fn.Body.List[:len(fn.Body.List)-1], route, fn.Body.List[len(fn.Body.List)-1])
	}

	// Write back
	var buf strings.Builder
	printer.Fprint(&buf, fset, file)
	os.WriteFile("routes.go", []byte(buf.String()), 0644)
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
