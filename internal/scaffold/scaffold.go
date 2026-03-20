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

// Renderer is the interface satisfied by *template.Renderer. Declared here
// so that RenderTemplates can be tested with a fake implementation.
type Renderer interface {
	Render(templateName string, data any) (string, error)
}

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
	WorkDir     string
}

type TemplateFile struct {
	FileName    string
	Destination string
}

func newRenderer(templateDir string) Renderer {
	if templateDir != "" {
		return template.NewRendererWithExternalDir(templateDir)
	}
	return template.NewRenderer()
}

func BackupDir(path string) error {
	backupName := fmt.Sprintf("%s.bak.%s", path, time.Now().Format("20060102150405"))
	return os.Rename(path, backupName)
}

func GetMetadata(dir string) (module.Metadata, error) {
	if _, err := os.Stat(filepath.Join(dir, "metadata.json")); err != nil {
		return module.Metadata{}, fmt.Errorf("%s is not a valid module directory", dir)
	}

	metadata, err := module.ReadMetadata(filepath.Join(dir, "metadata.json"))
	if err != nil {
		return module.Metadata{}, fmt.Errorf("failed to read module metadata: %w", err)
	}

	return metadata, nil
}

func ConstructDestinationFilename(name string, moduleName string, prefix string, suffix string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name cannot be empty")
	}

	parts := strings.Split(name, "::")

	for _, part := range parts {
		if part == "" {
			return "", fmt.Errorf("name %q contains an empty component (check for leading, trailing, or consecutive '::')", name)
		}
		if strings.ContainsAny(part, "/\\") {
			return "", fmt.Errorf("name %q contains an invalid path separator in component %q", name, part)
		}
		if part == ".." || part == "." {
			return "", fmt.Errorf("name %q contains an invalid component %q", name, part)
		}
	}

	if parts[0] == moduleName {
		return "", fmt.Errorf("module name should not be included in class name")
	}

	fileName := parts[len(parts)-1]
	filePath := parts[:len(parts)-1]

	pathParts := append([]string{prefix}, filePath...)
	pathParts = append(pathParts, fileName+suffix)
	return filepath.Join(pathParts...), nil
}

func validateComponentName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("name %q contains an invalid path separator", name)
	}
	if name == ".." || name == "." {
		return fmt.Errorf("name %q is not a valid component name", name)
	}
	// Guard against names that would escape the target directory when joined.
	// filepath.Join cleans the path, so "foo/../bar" becomes "bar" -- we catch
	// this by checking the cleaned result stays within a known base.
	cleaned := filepath.Join("base", name)
	if !strings.HasPrefix(cleaned, filepath.Join("base", "")) {
		return fmt.Errorf("name %q would escape the target directory", name)
	}
	return nil
}

func RenderTemplates(renderer Renderer, templateFiles []TemplateFile, data any, overwrite bool) error {
	for _, t := range templateFiles {
		if !overwrite {
			if _, err := os.Stat(t.Destination); err == nil {
				return fmt.Errorf("file %s already exists", t.Destination)
			}
		}
		rendered, err := renderer.Render(t.FileName, data)
		if err != nil {
			return fmt.Errorf("failed to render template %s: %w", t.FileName, err)
		}
		if err := os.MkdirAll(filepath.Dir(t.Destination), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(t.Destination), err)
		}
		if err := os.WriteFile(t.Destination, []byte(rendered), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", t.Destination, err)
		}
	}
	return nil
}

func NewModule(opts Options) error {
	if err := validateComponentName(opts.Name); err != nil {
		return fmt.Errorf("invalid module name: %w", err)
	}

	baseDir := opts.TargetDir
	if baseDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		baseDir = cwd
	}

	moduleDir := filepath.Join(baseDir, opts.Name)

	if _, err := os.Stat(moduleDir); err == nil {
		if !opts.Force {
			return fmt.Errorf("directory %s already exists, use --force to replace it", moduleDir)
		}
		if err := BackupDir(moduleDir); err != nil {
			return fmt.Errorf("failed to back up existing directory: %w", err)
		}
		fmt.Printf("Renamed existing directory %s\n", moduleDir)
	}

	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", moduleDir, err)
	}

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
		{FileName: "module/data/common.yaml", Destination: filepath.Join(moduleDir, "data", "common.yaml")},
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
	metadata, err := GetMetadata(opts.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	moduleName := metadata.ModuleName()

	classFile, err := ConstructDestinationFilename(
		opts.Name,
		moduleName,
		filepath.Join(opts.WorkDir, "manifests"),
		".pp",
	)
	if err != nil {
		return fmt.Errorf("failed to construct class file path: %w", err)
	}

	specFile, err := ConstructDestinationFilename(
		opts.Name,
		moduleName,
		filepath.Join(opts.WorkDir, "spec", "classes"),
		"_spec.rb",
	)
	if err != nil {
		return fmt.Errorf("failed to construct spec file path: %w", err)
	}

	className := fmt.Sprintf("%s::%s", moduleName, opts.Name)

	if _, err := os.Stat(classFile); err == nil {
		return fmt.Errorf("class %s already exists: %s", className, classFile)
	}

	renderer := newRenderer(opts.TemplateDir)

	templates := []TemplateFile{
		{FileName: "class/class.pp", Destination: classFile},
		{FileName: "class/class_spec.rb", Destination: specFile},
	}

	data := struct{ Name string }{Name: className}

	fmt.Printf("creating class %s...\n", className)

	if err := RenderTemplates(renderer, templates, data, false); err != nil {
		return fmt.Errorf("failed to render templates: %w", err)
	}

	return nil
}

func NewDefinedType(opts ComponentOptions) error {
	metadata, err := GetMetadata(opts.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	moduleName := metadata.ModuleName()

	typeFile, err := ConstructDestinationFilename(
		opts.Name,
		moduleName,
		filepath.Join(opts.WorkDir, "manifests"),
		".pp",
	)
	if err != nil {
		return fmt.Errorf("failed to construct defined_type file path: %w", err)
	}

	specFile, err := ConstructDestinationFilename(
		opts.Name,
		moduleName,
		filepath.Join(opts.WorkDir, "spec", "defines"),
		"_spec.rb",
	)
	if err != nil {
		return fmt.Errorf("failed to construct defined_type test file path: %w", err)
	}

	typeName := fmt.Sprintf("%s::%s", moduleName, opts.Name)

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

	data := struct{ Name string }{Name: typeName}

	fmt.Printf("creating defined_type %s...\n", typeName)

	if err := RenderTemplates(renderer, templates, data, false); err != nil {
		return fmt.Errorf("failed to render templates: %w", err)
	}

	return nil
}

func NewFact(opts ComponentOptions) error {
	// We only run this to see that we're in a valid module directory.
	// Nothing is ever done with the metadata.
	_, err := GetMetadata(opts.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	// Check to ensure opts.Name doesn't contain '::' as that would be invalid
	// for a fact name.
	if strings.Contains(opts.Name, "::") {
		return fmt.Errorf("fact name cannot contain '::'")
	}

	// Check to see if a fact by given name already exists in the module, and
	// return an error if it does.
	factFileName := filepath.Join(opts.WorkDir, "lib", "facter", opts.Name+".rb")
	if _, err := os.Stat(factFileName); err == nil {
		return fmt.Errorf("fact %s already exists: %s", opts.Name, factFileName)
	}

	factTestFileName := filepath.Join(opts.WorkDir, "spec", "unit", "facter", opts.Name+"_spec.rb")
	if _, err := os.Stat(factTestFileName); err == nil {
		return fmt.Errorf("fact %s test already exists: %s", opts.Name, factTestFileName)
	}

	// Set up the templates and render them to disk at the appropriate locations.
	fmt.Printf("creating fact %s...\n", opts.Name)
	renderer := newRenderer(opts.TemplateDir)

	templates := []TemplateFile{
		{FileName: "fact/fact.rb", Destination: factFileName},
		{FileName: "fact/fact_spec.rb", Destination: factTestFileName},
	}

	data := struct{ Name string }{Name: opts.Name}

	if err := RenderTemplates(renderer, templates, data, false); err != nil {
		return fmt.Errorf("failed to render templates: %w", err)
	}

	return nil
}

func NewFunction(opts ComponentOptions) error {
	// Get metadata, we need the module name to construct the function name
	metadata, err := GetMetadata(opts.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	moduleName := metadata.ModuleName()
	functionName := fmt.Sprintf("%s::%s", moduleName, opts.Name)

	functionFile, err := ConstructDestinationFilename(
		opts.Name,
		moduleName,
		filepath.Join(opts.WorkDir, "functions"),
		".pp",
	)
	if err != nil {
		return fmt.Errorf("failed to construct function file path: %w", err)
	}
	if _, err := os.Stat(functionFile); err == nil {
		return fmt.Errorf("function %s already exists: %s", functionName, functionFile)
	}

	specFile, err := ConstructDestinationFilename(
		opts.Name,
		moduleName,
		filepath.Join(opts.WorkDir, "spec", "functions"),
		"_spec.rb",
	)
	if err != nil {
		return fmt.Errorf("failed to construct function test file path: %w", err)
	}
	if _, err := os.Stat(specFile); err == nil {
		return fmt.Errorf("function %s test already exists: %s", functionName, specFile)
	}

	renderer := newRenderer(opts.TemplateDir)

	templates := []TemplateFile{
		{FileName: "function/function.pp", Destination: functionFile},
		{FileName: "function/function_spec.rb", Destination: specFile},
	}

	data := struct{ Name string }{Name: functionName}

	fmt.Printf("creating function %s...\n", functionName)

	if err := RenderTemplates(renderer, templates, data, false); err != nil {
		return fmt.Errorf("failed to render templates: %w", err)
	}

	return nil
}
