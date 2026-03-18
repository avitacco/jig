package scaffold

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// fakeRenderer satisfies the Renderer interface and returns configurable output.
type fakeRenderer struct {
	output string
	err    error
}

func (f *fakeRenderer) Render(_ string, _ any) (string, error) {
	return f.output, f.err
}

// makeModuleDir creates a temp directory with a minimal valid metadata.json,
// suitable for use as a WorkDir in tests.
func makeModuleDir(t *testing.T, forgeUser, moduleName string) string {
	t.Helper()
	dir := t.TempDir()
	meta := map[string]any{
		"name":                    forgeUser + "-" + moduleName,
		"version":                 "0.1.0",
		"author":                  forgeUser,
		"license":                 "Apache-2.0",
		"summary":                 "test",
		"source":                  "https://example.com",
		"dependencies":            []any{},
		"requirements":            []any{},
		"operatingsystem_support": []any{},
		"tags":                    []any{},
		"pdk-version":             "3.4.0",
	}
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// --- GetMetadata ---

func TestGetMetadata(t *testing.T) {
	t.Run("valid module directory", func(t *testing.T) {
		dir := makeModuleDir(t, "myuser", "mymodule")
		m, err := GetMetadata(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m.Name != "myuser-mymodule" {
			t.Errorf("Name: got %q, want %q", m.Name, "myuser-mymodule")
		}
	})

	t.Run("missing metadata.json", func(t *testing.T) {
		dir := t.TempDir()
		_, err := GetMetadata(dir)
		if err == nil {
			t.Error("expected error for missing metadata.json, got nil")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "metadata.json"), []byte("{{{"), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := GetMetadata(dir)
		if err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
	})
}

// --- RenderTemplates ---

func TestRenderTemplates(t *testing.T) {
	t.Run("renders files to disk", func(t *testing.T) {
		dir := t.TempDir()
		renderer := &fakeRenderer{output: "rendered content"}
		templates := []TemplateFile{
			{FileName: "foo.pp", Destination: filepath.Join(dir, "manifests", "foo.pp")},
			{FileName: "bar.pp", Destination: filepath.Join(dir, "manifests", "bar.pp")},
		}

		if err := RenderTemplates(renderer, templates, nil, false); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, tf := range templates {
			content, err := os.ReadFile(tf.Destination)
			if err != nil {
				t.Errorf("expected file %s to exist: %v", tf.Destination, err)
				continue
			}
			if string(content) != "rendered content" {
				t.Errorf("file %s: got %q, want %q", tf.Destination, string(content), "rendered content")
			}
		}
	})

	t.Run("returns error when file exists and overwrite is false", func(t *testing.T) {
		dir := t.TempDir()
		existing := filepath.Join(dir, "init.pp")
		if err := os.WriteFile(existing, []byte("original"), 0644); err != nil {
			t.Fatal(err)
		}

		renderer := &fakeRenderer{output: "new content"}
		templates := []TemplateFile{
			{FileName: "init.pp", Destination: existing},
		}

		err := RenderTemplates(renderer, templates, nil, false)
		if err == nil {
			t.Error("expected error for existing file with overwrite=false, got nil")
		}

		// Original file should be untouched
		content, _ := os.ReadFile(existing)
		if string(content) != "original" {
			t.Errorf("existing file was modified despite overwrite=false")
		}
	})

	t.Run("overwrites file when overwrite is true", func(t *testing.T) {
		dir := t.TempDir()
		existing := filepath.Join(dir, "init.pp")
		if err := os.WriteFile(existing, []byte("original"), 0644); err != nil {
			t.Fatal(err)
		}

		renderer := &fakeRenderer{output: "new content"}
		templates := []TemplateFile{
			{FileName: "init.pp", Destination: existing},
		}

		if err := RenderTemplates(renderer, templates, nil, true); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, _ := os.ReadFile(existing)
		if string(content) != "new content" {
			t.Errorf("got %q, want %q", string(content), "new content")
		}
	})

	t.Run("returns error when renderer fails", func(t *testing.T) {
		dir := t.TempDir()
		renderer := &fakeRenderer{err: errors.New("render failed")}
		templates := []TemplateFile{
			{FileName: "foo.pp", Destination: filepath.Join(dir, "foo.pp")},
		}

		if err := RenderTemplates(renderer, templates, nil, false); err == nil {
			t.Error("expected error from renderer, got nil")
		}
	})
}

// --- NewModule ---

func TestNewModule(t *testing.T) {
	t.Run("creates module with expected files", func(t *testing.T) {
		dir := t.TempDir()
		opts := Options{
			ForgeUser: "myuser",
			Name:      "mymodule",
			Author:    "My Name",
			License:   "Apache-2.0",
			Summary:   "A test module",
			Source:    "https://example.com",
			TargetDir: dir,
		}

		if err := NewModule(opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		moduleDir := filepath.Join(dir, "mymodule")
		expectedFiles := []string{
			"metadata.json",
			"manifests/init.pp",
			"README.md",
			"CHANGELOG.md",
			"Gemfile",
			"Rakefile",
			".gitignore",
			".pdkignore",
			"hiera.yaml",
			"spec/classes/init_spec.rb",
			"spec/spec_helper.rb",
			"spec/default_facts.yml",
			"data/common.yaml",
			"data/.gitkeep",
			"examples/.gitkeep",
			"files/.gitkeep",
			"tasks/.gitkeep",
			"templates/.gitkeep",
		}
		for _, f := range expectedFiles {
			path := filepath.Join(moduleDir, f)
			if _, err := os.Stat(path); err != nil {
				t.Errorf("expected file %s to exist: %v", f, err)
			}
		}
	})

	t.Run("returns error when directory exists without force", func(t *testing.T) {
		dir := t.TempDir()
		moduleDir := filepath.Join(dir, "mymodule")
		if err := os.Mkdir(moduleDir, 0755); err != nil {
			t.Fatal(err)
		}

		opts := Options{
			Name:      "mymodule",
			TargetDir: dir,
		}

		if err := NewModule(opts); err == nil {
			t.Error("expected error for existing directory without --force, got nil")
		}
	})

	t.Run("backs up existing directory when force is set", func(t *testing.T) {
		dir := t.TempDir()
		moduleDir := filepath.Join(dir, "mymodule")
		if err := os.Mkdir(moduleDir, 0755); err != nil {
			t.Fatal(err)
		}
		// Put a sentinel file in the existing directory
		if err := os.WriteFile(filepath.Join(moduleDir, "sentinel"), []byte("old"), 0644); err != nil {
			t.Fatal(err)
		}

		opts := Options{
			ForgeUser: "myuser",
			Name:      "mymodule",
			Author:    "My Name",
			License:   "Apache-2.0",
			Summary:   "A test module",
			Source:    "https://example.com",
			Force:     true,
			TargetDir: dir,
		}

		if err := NewModule(opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// New module dir should exist and have metadata.json
		if _, err := os.Stat(filepath.Join(moduleDir, "metadata.json")); err != nil {
			t.Error("expected metadata.json in new module dir")
		}

		// A backup should exist somewhere in dir
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		var backupFound bool
		for _, e := range entries {
			if e.Name() != "mymodule" && len(e.Name()) > len("mymodule") {
				backupFound = true
				break
			}
		}
		if !backupFound {
			t.Error("expected a backup directory to exist after --force")
		}
	})

	t.Run("metadata.json contains correct name", func(t *testing.T) {
		dir := t.TempDir()
		opts := Options{
			ForgeUser: "myuser",
			Name:      "mymodule",
			Author:    "My Name",
			License:   "Apache-2.0",
			Summary:   "A test module",
			Source:    "https://example.com",
			TargetDir: dir,
		}

		if err := NewModule(opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m, err := GetMetadata(filepath.Join(dir, "mymodule"))
		if err != nil {
			t.Fatalf("could not read generated metadata: %v", err)
		}
		if m.Name != "myuser-mymodule" {
			t.Errorf("Name: got %q, want %q", m.Name, "myuser-mymodule")
		}
	})
}

// --- NewClass ---

func TestNewClass(t *testing.T) {
	t.Run("creates class and spec files", func(t *testing.T) {
		dir := makeModuleDir(t, "myuser", "mymodule")

		opts := ComponentOptions{
			Name:    "myclass",
			WorkDir: dir,
		}

		if err := NewClass(opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedFiles := []string{
			filepath.Join(dir, "manifests", "myclass.pp"),
			filepath.Join(dir, "spec", "classes", "myclass_spec.rb"),
		}
		for _, f := range expectedFiles {
			if _, err := os.Stat(f); err != nil {
				t.Errorf("expected file %s to exist: %v", f, err)
			}
		}
	})

	t.Run("creates namespaced class and spec files", func(t *testing.T) {
		dir := makeModuleDir(t, "myuser", "mymodule")

		opts := ComponentOptions{
			Name:    "myclass::sub",
			WorkDir: dir,
		}

		if err := NewClass(opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedFiles := []string{
			filepath.Join(dir, "manifests", "myclass", "sub.pp"),
			filepath.Join(dir, "spec", "classes", "myclass", "sub_spec.rb"),
		}
		for _, f := range expectedFiles {
			if _, err := os.Stat(f); err != nil {
				t.Errorf("expected file %s to exist: %v", f, err)
			}
		}
	})

	t.Run("returns error when class file already exists", func(t *testing.T) {
		dir := makeModuleDir(t, "myuser", "mymodule")

		classPath := filepath.Join(dir, "manifests", "myclass.pp")
		if err := os.MkdirAll(filepath.Dir(classPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(classPath, []byte("existing"), 0644); err != nil {
			t.Fatal(err)
		}

		opts := ComponentOptions{
			Name:    "myclass",
			WorkDir: dir,
		}

		if err := NewClass(opts); err == nil {
			t.Error("expected error for existing class file, got nil")
		}
	})

	t.Run("returns error when name includes module name", func(t *testing.T) {
		dir := makeModuleDir(t, "myuser", "mymodule")

		opts := ComponentOptions{
			Name:    "mymodule::myclass",
			WorkDir: dir,
		}

		if err := NewClass(opts); err == nil {
			t.Error("expected error when class name includes module name, got nil")
		}
	})

	t.Run("returns error for invalid module directory", func(t *testing.T) {
		opts := ComponentOptions{
			Name:    "myclass",
			WorkDir: t.TempDir(), // no metadata.json
		}

		if err := NewClass(opts); err == nil {
			t.Error("expected error for missing metadata.json, got nil")
		}
	})
}

// --- NewDefinedType ---

func TestNewDefinedType(t *testing.T) {
	t.Run("creates defined type and spec files", func(t *testing.T) {
		dir := makeModuleDir(t, "myuser", "mymodule")

		opts := ComponentOptions{
			Name:    "mytype",
			WorkDir: dir,
		}

		if err := NewDefinedType(opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedFiles := []string{
			filepath.Join(dir, "manifests", "mytype.pp"),
			filepath.Join(dir, "spec", "defines", "mytype_spec.rb"),
		}
		for _, f := range expectedFiles {
			if _, err := os.Stat(f); err != nil {
				t.Errorf("expected file %s to exist: %v", f, err)
			}
		}
	})

	t.Run("returns error when type file already exists", func(t *testing.T) {
		dir := makeModuleDir(t, "myuser", "mymodule")

		typePath := filepath.Join(dir, "manifests", "mytype.pp")
		if err := os.MkdirAll(filepath.Dir(typePath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(typePath, []byte("existing"), 0644); err != nil {
			t.Fatal(err)
		}

		opts := ComponentOptions{
			Name:    "mytype",
			WorkDir: dir,
		}

		if err := NewDefinedType(opts); err == nil {
			t.Error("expected error for existing type file, got nil")
		}
	})

	t.Run("returns error when spec file already exists", func(t *testing.T) {
		dir := makeModuleDir(t, "myuser", "mymodule")

		specPath := filepath.Join(dir, "spec", "defines", "mytype_spec.rb")
		if err := os.MkdirAll(filepath.Dir(specPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(specPath, []byte("existing"), 0644); err != nil {
			t.Fatal(err)
		}

		opts := ComponentOptions{
			Name:    "mytype",
			WorkDir: dir,
		}

		if err := NewDefinedType(opts); err == nil {
			t.Error("expected error for existing spec file, got nil")
		}
	})
}
func TestConstructDestinationFilename_Adversarial(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty name",
			input:   "",
			wantErr: true,
		},
		{
			name:    "path traversal with ..",
			input:   "..",
			wantErr: true,
		},
		{
			name:    "dot component",
			input:   ".",
			wantErr: true,
		},
		{
			name:    "slash in name",
			input:   "foo/bar",
			wantErr: true,
		},
		{
			name:    "backslash in name",
			input:   `foo\bar`,
			wantErr: true,
		},
		{
			name:    "leading double colon",
			input:   "::foo",
			wantErr: true,
		},
		{
			name:    "trailing double colon",
			input:   "foo::",
			wantErr: true,
		},
		{
			name:    "consecutive double colons",
			input:   "foo::::bar",
			wantErr: true,
		},
		{
			name:    "valid simple name",
			input:   "myclass",
			wantErr: false,
		},
		{
			name:    "valid namespaced name",
			input:   "myclass::sub",
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ConstructDestinationFilename(tc.input, "mymodule", "manifests", ".pp")
			if tc.wantErr && err == nil {
				t.Errorf("expected error for input %q, got nil", tc.input)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error for input %q: %v", tc.input, err)
			}
		})
	}
}

func TestValidateComponentName(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "mymodule", false},
		{"valid with numbers", "mymodule2", false},
		{"valid with underscores", "my_module", false},
		{"empty", "", true},
		{"slash", "foo/bar", true},
		{"backslash", `foo\bar`, true},
		{"dot dot", "..", true},
		{"single dot", ".", true},
		{"path traversal embedded", "foo/../bar", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateComponentName(tc.input)
			if tc.wantErr && err == nil {
				t.Errorf("expected error for %q, got nil", tc.input)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error for %q: %v", tc.input, err)
			}
		})
	}
}

func TestNewModule_PathTraversal(t *testing.T) {
	dir := t.TempDir()

	cases := []struct {
		name       string
		moduleName string
	}{
		{"path separator", "foo/bar"},
		{"dot dot", ".."},
		{"dot dot slash", "../evil"},
		{"empty", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := Options{
				ForgeUser: "myuser",
				Name:      tc.moduleName,
				Author:    "My Name",
				License:   "Apache-2.0",
				Summary:   "test",
				Source:    "https://example.com",
				TargetDir: dir,
			}
			err := NewModule(opts)
			if err == nil {
				t.Errorf("expected error for module name %q, got nil", tc.moduleName)
			}
		})
	}
}

func TestNewClass_EmptyName(t *testing.T) {
	dir := makeModuleDir(t, "myuser", "mymodule")
	opts := ComponentOptions{
		Name:    "",
		WorkDir: dir,
	}
	if err := NewClass(opts); err == nil {
		t.Error("expected error for empty class name, got nil")
	}
	// Verify no ".pp" file was created
	if _, err := os.Stat(filepath.Join(dir, "manifests", ".pp")); err == nil {
		t.Error("a file named '.pp' was created from an empty class name")
	}
}

func TestNewDefinedType_EmptyName(t *testing.T) {
	dir := makeModuleDir(t, "myuser", "mymodule")
	opts := ComponentOptions{
		Name:    "",
		WorkDir: dir,
	}
	if err := NewDefinedType(opts); err == nil {
		t.Error("expected error for empty defined type name, got nil")
	}
}
