package template

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

func DumpTemplates(dest string) error {
	return fs.WalkDir(embeddedTemplates, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Strip the "templates/" prefix to get the relative path
		relPath := strings.TrimPrefix(path, "templates/")
		if relPath == "" {
			return nil
		}

		destPath := filepath.Join(dest, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		content, err := embeddedTemplates.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded template %s: %w", path, err)
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
		}

		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", destPath, err)
		}

		fmt.Printf("wrote %s\n", destPath)
		return nil
	})
}
