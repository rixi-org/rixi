package main

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"text/template"
)

//go:embed templates/*
//go:embed templates/controller/*
//go:embed templates/favicon/*
//go:embed templates/view/*
var tmplFS embed.FS

type Project struct {
	Name string
}

func createProject(name string) {
	os.MkdirAll(name+"/controller", 0755)
	os.MkdirAll(name+"/favicon", 0755)

	files := []struct{ Out, Tmpl string }{
		{"README.md", "templates/README.md.tmpl"},
		{"AGENTS.md", "templates/AGENTS.md.tmpl"},
		{".gitignore", "templates/gitignore.tmpl"},
		{"CLAUDE.md", "templates/CLAUDE.md.tmpl"},
		{"GEMINI.md", "templates/GEMINI.md.tmpl"},
		{"main.go", "templates/main.go.tmpl"},
		{"middleware.go", "templates/middleware.go.tmpl"},
		{"routes.go", "templates/routes.go.tmpl"},
		{"index.html", "templates/index.html.tmpl"},
		{"controller/health.go", "templates/controller/health.go.tmpl"},
		{"controller/home.go", "templates/controller/home.go.tmpl"},
	}

	for _, f := range files {
		t := template.Must(template.ParseFS(tmplFS, f.Tmpl))
		var buf bytes.Buffer
		t.Execute(&buf, Project{Name: name})
		os.WriteFile(name+"/"+f.Out, buf.Bytes(), 0644)
	}

	// Copy favicon files (binary, no template processing)
	faviconFiles := []string{
		"android-chrome-192x192.png",
		"android-chrome-512x512.png",
		"apple-touch-icon.png",
		"favicon-16x16.png",
		"favicon-32x32.png",
		"favicon.ico",
		"site.webmanifest",
	}
	for _, f := range faviconFiles {
		data, err := tmplFS.ReadFile("templates/favicon/" + f)
		if err != nil {
			fmt.Printf("warning: could not read favicon %s: %v\n", f, err)
			continue
		}
		os.WriteFile(name+"/favicon/"+f, data, 0644)
	}

	exec.Command("git", "init", name).Run()

	cmd := exec.Command("go", "mod", "init", name)
	cmd.Dir = name
	cmd.Run()

	fmt.Print("installing pkgsite-cli... ")
	exec.Command("go", "install", "golang.org/x/pkgsite/cmd/internal/pkgsite-cli@latest").Run()
	fmt.Println("done")

	fmt.Printf("created project %s\n", name)
}
