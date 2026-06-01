package main

import (
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func dev() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	timestampFile := ".dev-timestamp"
	touch(timestampFile)

	var cmd *exec.Cmd
	restart := make(chan struct{}, 1)

	start := func() *exec.Cmd {
		slog.Info("starting dev server")
		c := exec.Command("go", "run", ".")
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if err := c.Start(); err != nil {
			slog.Error("failed to start", "error", err)
			return nil
		}
		return c
	}

	cmd = start()

	// Watch .go files for changes
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			changed, err := anyGoFileChangedSince(timestampFile)
			if err != nil {
				continue
			}
			if changed {
				select {
				case restart <- struct{}{}:
				default:
				}
			}
		}
	}()

	// Handle signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-restart:
			touch(timestampFile)
			slog.Info("changes detected, restarting...")
			if cmd != nil {
				syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
				waitDone := make(chan struct{})
				go func() {
					cmd.Wait()
					close(waitDone)
				}()
				select {
				case <-waitDone:
				case <-time.After(5 * time.Second):
					syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
					<-waitDone
				}
			}
			cmd = start()

		case <-sig:
			slog.Info("shutting down")
			if cmd != nil {
				syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
				go func() {
					time.Sleep(5 * time.Second)
					syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
				}()
				cmd.Wait()
			}
			os.Exit(0)
		}
	}
}

func touch(path string) {
	f, err := os.Create(path)
	if err != nil {
		return
	}
	f.Close()
}

// anyGoFileChangedSince walks the project tree and returns true
// if any .go file has been modified after timestampFile's mtime.
// Skips vendor/ and .git/ directories.
func anyGoFileChangedSince(timestampPath string) (bool, error) {
	ts, err := os.Stat(timestampPath)
	if err != nil {
		return false, err
	}
	threshold := ts.ModTime()

	changed := false
	err = filepath.Walk(".", func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible files
		}
		if info.IsDir() {
			if info.Name() == "vendor" || info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(p) == ".go" && info.ModTime().After(threshold) {
			changed = true
			return filepath.SkipAll
		}
		return nil
	})
	return changed, err
}
