package template

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	tmplpkg "text/template"
)

//go:embed templates
var embeddedTemplates embed.FS

type Renderer struct {
	externalDir string
}

func NewRenderer() *Renderer {
	return &Renderer{}
}

func NewRendererWithExternalDir(dir string) *Renderer {
	return &Renderer{
		externalDir: dir,
	}
}

func (r Renderer) Render(templateName string, data any) (string, error) {
	var content []byte
	var err error

	if r.externalDir != "" {
		content, err = os.ReadFile(filepath.Join(r.externalDir, templateName))
		if err != nil {
			if !os.IsNotExist(err) {
				return "", fmt.Errorf("failed to read external template %s: %w", templateName, err)
			}
			// file not found in external dir, fall through to embedded templates
			content, err = embeddedTemplates.ReadFile("templates/" + templateName)
			if err != nil {
				return "", fmt.Errorf("failed to read embedded template %s: %w", templateName, err)
			}
		}

	} else {
		content, err = embeddedTemplates.ReadFile("templates/" + templateName)
		if err != nil {
			return "", fmt.Errorf("failed to read embedded template %s: %w", templateName, err)
		}
	}

	t, err := tmplpkg.New(templateName).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", templateName, err)
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", templateName, err)
	}
	return buf.String(), nil
}
