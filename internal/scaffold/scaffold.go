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

func BackupDir(path string) error {
	backupName := fmt.Sprintf("%s.bak.%s", path, time.Now().Format("20060102150405"))
	return os.Rename(path, backupName)
}

func GetMetadata() (module.Metadata, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return module.Metadata{}, fmt.Errorf("failed to get current working directory: %w", err)
	}
	if _, err := os.Stat(filepath.Join(cwd, "metadata.json")); err != nil {
		return module.Metadata{}, fmt.Errorf("%s is not a valid module directory", cwd)
	}

	// Read the module metadata
	metadata, err := module.ReadMetadata(filepath.Join(cwd, "metadata.json"))
	if err != nil {
		return module.Metadata{}, fmt.Errorf("failed to read module metadata: %w", err)
	}

	return metadata, nil
}

func ConstructDestinationFilename(name string, moduleName string, prefix string, suffix string) (string, error) {
	parts := strings.Split(name, "::")
	if parts[0] == moduleName {
		return "", fmt.Errorf("module name should not be included in class name")
	}
	fileName := parts[len(parts)-1]
	filePath := parts[:len(parts)-1]

	pathParts := append([]string{prefix}, filePath...)
	pathParts = append(pathParts, fileName+suffix)
	return filepath.Join(pathParts...), nil
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
		err = BackupDir(moduleDir)
		fmt.Printf("Renamed existing directory %s\n", moduleDir)
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
	// Attempt to load the module metadata
	metadata, err := GetMetadata()
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	moduleName := metadata.ModuleName()

	// Construct the class and spec file paths and the class name
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	classFile, err := ConstructDestinationFilename(
		opts.Name,
		moduleName,
		filepath.Join(append([]string{cwd, "manifests"})...),
		".pp",
	)

	specFile, err := ConstructDestinationFilename(
		opts.Name,
		moduleName,
		filepath.Join(append([]string{cwd, "spec", "classes"})...),
		"_spec.rb",
	)

	className := fmt.Sprintf("%s::%s", moduleName, opts.Name)

	// Check if the class file already exists
	if _, err := os.Stat(classFile); err == nil {
		return fmt.Errorf("class %s already exists: %s", className, classFile)
	}

	// Render the class and spec templates
	renderer := newRenderer(opts.TemplateDir)

	templates := []TemplateFile{
		{FileName: "class/class.pp", Destination: classFile},
		{FileName: "class/class_spec.rb", Destination: specFile},
	}

	data := struct {
		Name string
	}{
		Name: className,
	}

	fmt.Printf("creating class %s...\n", className)

	err = RenderTemplates(renderer, templates, data, false)
	if err != nil {
		return fmt.Errorf("failed to render templates: %w", err)
	}

	return nil
}

func NewDefinedType(opts ComponentOptions) error {
	metadata, err := GetMetadata()
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	moduleName := metadata.ModuleName()

	// Construct the defined_type and spec file paths and the defined_type name
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	typeFile, err := ConstructDestinationFilename(
		opts.Name,
		moduleName,
		filepath.Join(append([]string{cwd, "manifests"})...),
		".pp",
	)
	if err != nil {
		return fmt.Errorf("failed to construct defined_type file path: %w", err)
	}

	specFile, err := ConstructDestinationFilename(
		opts.Name,
		moduleName,
		filepath.Join(append([]string{cwd, "spec", "defines"})...),
		"_spec.rb",
	)
	if err != nil {
		return fmt.Errorf("failed to construct defined_type test file path: %w", err)
	}

	typeName := fmt.Sprintf("%s::%s", moduleName, opts.Name)

	// Check if the defined_type or test file already exists
	if _, err := os.Stat(typeFile); err == nil {
		return fmt.Errorf("defined_type %s already exists: %s", typeName, typeFile)
	}
	if _, err := os.Stat(specFile); err == nil {
		return fmt.Errorf("defined_type %s test already exists: %s", typeName, specFile)
	}

	renderer := newRenderer(opts.TemplateDir)

	templates := []TemplateFile{
		{FileName: "type/defined_type.pp", Destination: typeFile},
		{FileName: "type/defined_type_spec.rb", Destination: specFile},
	}

	data := struct {
		Name string
	}{
		Name: typeName,
	}

	fmt.Printf("creating defined_type %s...\n", typeName)

	err = RenderTemplates(renderer, templates, data, false)
	if err != nil {
		return fmt.Errorf("failed to render templates: %w", err)
	}

	return nil
}
