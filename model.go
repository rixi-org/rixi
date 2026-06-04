package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

// ModelConfig holds the template data for generating a model file.
type ModelConfig struct {
	Name  string
	Table string
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

	if err := tmpl.Execute(f, ModelConfig{Name: modelName, Table: strings.ToLower(name) + "s"}); err != nil {
		fmt.Printf("error: failed to execute template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("created %s\n", modelPath)

	// Add migration to db.go
	addMigration(name)
}

// addMigration adds the migration call to db.go using AST.
func addMigration(name string) {
	module, _ := readModuleName()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "db.go", nil, parser.ParseComments)
	if err != nil {
		return
	}

	// Skip if already migrated
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "migrate" {
			continue
		}
		for _, stmt := range fn.Body.List {
			ifCall, ok := stmt.(*ast.IfStmt)
			if !ok {
				continue
			}
			call, ok := ifCall.Init.(*ast.AssignStmt)
			if !ok {
				continue
			}
			if expr, ok := call.Rhs[0].(*ast.CallExpr); ok {
				if sel, ok := expr.Fun.(*ast.SelectorExpr); ok {
					if sel.Sel.Name == "Migrate"+toPascal(name) {
						return
					}
				}
			}
		}
	}

	// Add model import if not present
	importExists := false
	for _, imp := range file.Imports {
		if imp.Path.Value == fmt.Sprintf(`"%s/model"`, module) {
			importExists = true
			break
		}
	}
	if !importExists {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.IMPORT {
				continue
			}
			gd.Specs = append(gd.Specs, &ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf(`"%s/model"`, module),
				},
			})
			break
		}
	}

	// Find migrate function and add call
	migrateFunc := "Migrate" + toPascal(name)
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "migrate" {
			continue
		}

		// Build: if err := model.MigrateX(db); err != nil { return err }
		ifStmt := &ast.IfStmt{
			Init: &ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "model"},
							Sel: &ast.Ident{Name: migrateFunc},
						},
						Args: []ast.Expr{&ast.Ident{Name: "db"}},
					},
				},
			},
			Cond: &ast.BinaryExpr{
				X:  &ast.Ident{Name: "err"},
				Op: token.NEQ,
				Y:  &ast.Ident{Name: "nil"},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{&ast.Ident{Name: "err"}},
					},
				},
			},
		}

		// Insert before return nil
		for i, stmt := range fn.Body.List {
			ret, ok := stmt.(*ast.ReturnStmt)
			if !ok {
				continue
			}
			if len(ret.Results) == 1 {
				if ident, ok := ret.Results[0].(*ast.Ident); ok && ident.Name == "nil" {
					fn.Body.List = append(fn.Body.List[:i], append([]ast.Stmt{ifStmt}, fn.Body.List[i:]...)...)
					break
				}
			}
		}
	}

	// Write back
	var buf strings.Builder
	printer.Fprint(&buf, fset, file)
	os.WriteFile("db.go", []byte(buf.String()), 0644)
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
