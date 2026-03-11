package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/avitacco/jig/internal/module"
	"github.com/avitacco/jig/internal/template"
)

type Options struct {
	ForgeUser   string
	Name        string
	Author      string
	License     string
	Summary     string
	Source      string
	Force       bool
	TargetDir   string
	TemplateDir string
}

type ComponentOptions struct {
	Name        string
	TemplateDir string
}

type TemplateFile struct {
	FileName    string
	Destination string
}

func newRenderer(templateDir string) *template.Renderer {
	if templateDir != "" {
		return template.NewRendererWithExternalDir(templateDir)
	}
	return template.NewRenderer()
}

func RenderTemplates(renderer *template.Renderer, templateFiles []TemplateFile, data any, overwrite bool) error {
	for _, template := range templateFiles {
		// Check if the destination file already exists and if it should be overwritten.
		if !overwrite {
			if _, err := os.Stat(template.Destination); err == nil {
				return fmt.Errorf("file %s already exists", template.Destination)
			}
		}
		// Render the template and write it to the destination file.
		rendered, err := renderer.Render(template.FileName, data)
		if err != nil {
			return fmt.Errorf("failed to render template %s: %w", template.FileName, err)
		}
		if err := os.MkdirAll(filepath.Dir(template.Destination), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(template.Destination), err)
		}
		if err := os.WriteFile(template.Destination, []byte(rendered), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", template.Destination, err)
		}
	}
	return nil
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
	meta := module.NewMetadata(opts.Name, opts.ForgeUser, opts.Author)
	meta.License = opts.License
	meta.Summary = opts.Summary
	meta.Source = opts.Source

	if err := meta.Write(filepath.Join(moduleDir, "metadata.json")); err != nil {
		return fmt.Errorf("failed to write metadata.json: %w", err)
	}

	renderer := newRenderer(opts.TemplateDir)

	templates := []TemplateFile{
		{FileName: "module/manifests/init.pp", Destination: filepath.Join(moduleDir, "manifests", "init.pp")},
		{FileName: "module/README.md", Destination: filepath.Join(moduleDir, "README.md")},
		{FileName: "module/CHANGELOG.md", Destination: filepath.Join(moduleDir, "CHANGELOG.md")},
		{FileName: "module/spec/class_spec.rb", Destination: filepath.Join(moduleDir, "spec", "classes", "init_spec.rb")},
		{FileName: "module/Gemfile", Destination: filepath.Join(moduleDir, "Gemfile")},
		{FileName: "module/Rakefile", Destination: filepath.Join(moduleDir, "Rakefile")},
		{FileName: "module/gitignore", Destination: filepath.Join(moduleDir, ".gitignore")},
		{FileName: "module/pdkignore", Destination: filepath.Join(moduleDir, ".pdkignore")},
		{FileName: "module/rubocop.yml", Destination: filepath.Join(moduleDir, ".rubocop.yml")},
		{FileName: "module/hiera.yaml", Destination: filepath.Join(moduleDir, "hiera.yaml")},
		{FileName: "module/spec/spec_helper.rb", Destination: filepath.Join(moduleDir, "spec", "spec_helper.rb")},
		{FileName: "module/spec/default_facts.yml", Destination: filepath.Join(moduleDir, "spec", "default_facts.yml")},
		{FileName: "common/gitkeep", Destination: filepath.Join(moduleDir, "data", ".gitkeep")},
		{FileName: "common/gitkeep", Destination: filepath.Join(moduleDir, "examples", ".gitkeep")},
		{FileName: "common/gitkeep", Destination: filepath.Join(moduleDir, "files", ".gitkeep")},
		{FileName: "common/gitkeep", Destination: filepath.Join(moduleDir, "tasks", ".gitkeep")},
		{FileName: "common/gitkeep", Destination: filepath.Join(moduleDir, "templates", ".gitkeep")},
	}

	data := struct {
		ModuleName string
		Author     string
		License    string
		ClassName  string
	}{
		ModuleName: opts.Name,
		Author:     opts.Author,
		License:    opts.License,
		ClassName:  opts.Name,
	}

	if err := RenderTemplates(renderer, templates, data, opts.Force); err != nil {
		return fmt.Errorf("failed to render templates: %w", err)
	}

	fmt.Printf("Created new module %s in %s\n", opts.Name, moduleDir)
	return nil
}

func NewClass(opts ComponentOptions) error {
	// Get the cwd and check if it's a module directory (contains a metadata.json file)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	if _, err := os.Stat(filepath.Join(cwd, "metadata.json")); err != nil {
		return fmt.Errorf("%s is not a valid module directory", cwd)
	}

	// Read the module metadata
	metadata, err := module.ReadMetadata(filepath.Join(cwd, "metadata.json"))
	if err != nil {
		return fmt.Errorf("failed to read module metadata: %w", err)
	}

	moduleName := metadata.ModuleName()

	// Figure out the filename for the class
	// Needs to handle several cases,
	// 1. module::classname (should raise an error, don't give module opts)
	// 2. sub::module::class::names (should be converted to sub/module/class/names.pp)
	// 3. classname (should be converted to classname.pp)
	parts := strings.Split(opts.Name, "::")
	if parts[0] == moduleName {
		return fmt.Errorf("module opts cannot be included in class opts")
	}
	fileName := parts[len(parts)-1]
	filePath := parts[:len(parts)-1]

	classFile := filepath.Join(append([]string{cwd, "manifests"}, append(filePath, fileName+".pp")...)...)
	className := fmt.Sprintf("%s::%s", moduleName, opts.Name)

	specFile := filepath.Join(append([]string{cwd, "spec", "classes"}, append(filePath, fileName+"_spec.rb")...)...)

	// Check if the class file already exists
	if _, err := os.Stat(classFile); err == nil {
		return fmt.Errorf("class %s already exists", fileName)
	}

	// Render the class and spec templates
	renderer := newRenderer(opts.TemplateDir)

	templates := []TemplateFile{
		{FileName: "class/manifests/class.pp", Destination: classFile},
		{FileName: "class/spec/classes/class_spec.rb", Destination: specFile},
	}

	data := struct {
		ClassName string
	}{
		ClassName: className,
	}

	fmt.Printf("creating class %s...\n", className)

	err = RenderTemplates(renderer, templates, data, false)
	if err != nil {
		return fmt.Errorf("failed to render templates: %w", err)
	}

	return nil
}
