package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
