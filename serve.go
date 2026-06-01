package main

import (
	"log/slog"
	"os"
	"os/exec"
)

func serve() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	cmd := exec.Command("go", "run", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	slog.Info("running project")

	if err := cmd.Run(); err != nil {
		slog.Error("serve failed", "error", err)
		os.Exit(1)
	}
}
