package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/avitacco/hammer/internal/module"
	"github.com/avitacco/hammer/internal/template"
)

type Options struct {
	Name      string
	Author    string
	License   string
	Summary   string
	Source    string
	Force     bool
	TargetDir string
}

func NewModule(opts Options) error {
	// Figure out the target directory
	baseDir := opts.TargetDir
	if baseDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		baseDir = cwd
	}

	moduleDir := filepath.Join(baseDir, opts.Name)

	// Check if the module directory already exists, if it does and the force
	// flag is not set, return an error. If the force flag IS set, rename the
	// existing directory before creating the new one.
	if _, err := os.Stat(moduleDir); err == nil {
		if !opts.Force {
			return fmt.Errorf("directory %s already exists, use --force to replace it", moduleDir)
		}
		backupName := fmt.Sprintf("%s.bak.%s", moduleDir, time.Now().Format("20060102150405"))
		if err := os.Rename(moduleDir, backupName); err != nil {
			return fmt.Errorf("failed to rename existing directory %s to %s: %w", moduleDir, backupName, err)
		}
		fmt.Printf("Renamed existing directory %s to %s\n", moduleDir, backupName)
	}

	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", moduleDir, err)
	}

	// Actually render the templates and write them to the module directory
	meta := module.NewMetadata(opts.Name, opts.Author)
	meta.License = opts.License
	meta.Summary = opts.Summary
	meta.Source = opts.Source

	if err := meta.Write(filepath.Join(moduleDir, "metadata.json")); err != nil {
		return fmt.Errorf("failed to write metadata.json: %w", err)
	}

	renderer := template.NewRenderer()

	templates := map[string]string{
		"module/manifests/init.pp": filepath.Join(moduleDir, "manifests", "init.pp"),
	}

	data := struct {
		ModuleName string
		Author     string
		License    string
	}{
		ModuleName: opts.Name,
		Author:     opts.Author,
		License:    opts.License,
	}

	for tmplName, destPath := range templates {
		rendered, err := renderer.Render(tmplName, data)
		if err != nil {
			return fmt.Errorf("failed to render template %s: %w", tmplName, err)
		}
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
		}
		if err := os.WriteFile(destPath, []byte(rendered), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", destPath, err)
		}
	}

	fmt.Printf("Created new module %s in %s\n", opts.Name, moduleDir)
	return nil
}
