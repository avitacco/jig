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

// makeFunctionTemplateDir creates a temp directory with minimal function
// templates. This is necessary because NewFunction calls newRenderer
// internally and cannot accept an injected renderer.
func makeFunctionTemplateDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	funcDir := filepath.Join(dir, "function")
	if err := os.MkdirAll(funcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(funcDir, "function.pp"), []byte("# {{.Name}}\nfunction {{.Name}}() >> Any {\n}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(funcDir, "function_spec.rb"), []byte("# frozen_string_literal: true\nrequire 'spec_helper'\ndescribe '{{.Name}}' do\nend\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// --- NewFunction ---

func TestNewFunction_CreatesFilesAtExpectedLocations(t *testing.T) {
	dir := makeModuleDir(t, "myuser", "mymodule")
	err := NewFunction(ComponentOptions{
		Name:        "myfunc",
		WorkDir:     dir,
		TemplateDir: makeFunctionTemplateDir(t),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFiles := []string{
		filepath.Join(dir, "functions", "myfunc.pp"),
		filepath.Join(dir, "spec", "functions", "myfunc_spec.rb"),
	}
	for _, f := range expectedFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("expected file %s to exist: %v", f, err)
		}
	}
}

// TestNewFunction_OutputContainsFullyQualifiedName verifies that the rendered
// files reference "module::function", not just the bare name passed in opts.Name.
// This tests the contract between NewFunction and the templates: if functionName
// were accidentally replaced with opts.Name in the data struct, this catches it.
func TestNewFunction_OutputContainsFullyQualifiedName(t *testing.T) {
	dir := makeModuleDir(t, "myuser", "mymodule")
	err := NewFunction(ComponentOptions{
		Name:        "myfunc",
		WorkDir:     dir,
		TemplateDir: makeFunctionTemplateDir(t),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	funcContent, err := os.ReadFile(filepath.Join(dir, "functions", "myfunc.pp"))
	if err != nil {
		t.Fatalf("could not read function file: %v", err)
	}
	if !contains(string(funcContent), "mymodule::myfunc") {
		t.Errorf("function file content %q does not contain fully-qualified name %q", string(funcContent), "mymodule::myfunc")
	}

	specContent, err := os.ReadFile(filepath.Join(dir, "spec", "functions", "myfunc_spec.rb"))
	if err != nil {
		t.Fatalf("could not read spec file: %v", err)
	}
	if !contains(string(specContent), "mymodule::myfunc") {
		t.Errorf("spec file content %q does not contain fully-qualified name %q", string(specContent), "mymodule::myfunc")
	}
}

// TestNewFunction_RejectsInvalidModuleDirectory verifies that NewFunction
// fails before creating any files when run outside a valid module directory.
func TestNewFunction_RejectsInvalidModuleDirectory(t *testing.T) {
	emptyDir := t.TempDir()
	err := NewFunction(ComponentOptions{
		Name:        "myfunc",
		WorkDir:     emptyDir,
		TemplateDir: makeFunctionTemplateDir(t),
	})
	if err == nil {
		t.Error("expected error for missing metadata.json, got nil")
	}
	if _, statErr := os.Stat(filepath.Join(emptyDir, "functions", "myfunc.pp")); statErr == nil {
		t.Error("function file should not have been created in an invalid module directory")
	}
}

// TestNewFunction_RejectsCorruptMetadata verifies that malformed JSON in
// metadata.json is caught and not silently ignored.
func TestNewFunction_RejectsCorruptMetadata(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), []byte("{{{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	err := NewFunction(ComponentOptions{
		Name:        "myfunc",
		WorkDir:     dir,
		TemplateDir: makeFunctionTemplateDir(t),
	})
	if err == nil {
		t.Error("expected error for corrupt metadata.json, got nil")
	}
}

// TestNewFunction_RejectsEmptyName verifies that an empty name is rejected
// before any filesystem work is done.
func TestNewFunction_RejectsEmptyName(t *testing.T) {
	dir := makeModuleDir(t, "myuser", "mymodule")
	err := NewFunction(ComponentOptions{
		Name:        "",
		WorkDir:     dir,
		TemplateDir: makeFunctionTemplateDir(t),
	})
	if err == nil {
		t.Error("expected error for empty function name, got nil")
	}
	if _, statErr := os.Stat(filepath.Join(dir, "functions")); statErr == nil {
		t.Error("functions directory should not have been created for empty name")
	}
}

// TestNewFunction_RejectsNameEqualToModuleName verifies that passing the
// module name as the function name is caught. ConstructDestinationFilename
// rejects it, but this test pins the behavior at the NewFunction level.
func TestNewFunction_RejectsNameEqualToModuleName(t *testing.T) {
	dir := makeModuleDir(t, "myuser", "mymodule")
	err := NewFunction(ComponentOptions{
		Name:        "mymodule",
		WorkDir:     dir,
		TemplateDir: makeFunctionTemplateDir(t),
	})
	if err == nil {
		t.Error("expected error when function name equals module name, got nil")
	}
}

// TestNewFunction_RejectsNameWithModulePrefix verifies that a user passing the
// fully-qualified name (e.g. "mymodule::myfunc") instead of just the
// unqualified name ("myfunc") is rejected.
func TestNewFunction_RejectsNameWithModulePrefix(t *testing.T) {
	dir := makeModuleDir(t, "myuser", "mymodule")
	err := NewFunction(ComponentOptions{
		Name:        "mymodule::myfunc",
		WorkDir:     dir,
		TemplateDir: makeFunctionTemplateDir(t),
	})
	if err == nil {
		t.Error("expected error when function name includes the module prefix, got nil")
	}
}

// TestNewFunction_RejectsPathSeparatorInName verifies that a name containing
// a slash or backslash cannot be used to write files outside the functions
// directory.
func TestNewFunction_RejectsPathSeparatorInName(t *testing.T) {
	cases := []string{"foo/bar", `foo\bar`}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			dir := makeModuleDir(t, "myuser", "mymodule")
			err := NewFunction(ComponentOptions{
				Name:        name,
				WorkDir:     dir,
				TemplateDir: makeFunctionTemplateDir(t),
			})
			if err == nil {
				t.Errorf("expected error for name %q with path separator, got nil", name)
			}
		})
	}
}

// TestNewFunction_RefusesIfFunctionFileExists_NoSpecCreated verifies that an
// existing function file causes an early error and that the spec file is not
// created as a side effect, leaving no partial state.
func TestNewFunction_RefusesIfFunctionFileExists_NoSpecCreated(t *testing.T) {
	dir := makeModuleDir(t, "myuser", "mymodule")
	funcPath := filepath.Join(dir, "functions", "myfunc.pp")
	if err := os.MkdirAll(filepath.Dir(funcPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(funcPath, []byte("# existing"), 0644); err != nil {
		t.Fatal(err)
	}

	err := NewFunction(ComponentOptions{
		Name:        "myfunc",
		WorkDir:     dir,
		TemplateDir: makeFunctionTemplateDir(t),
	})
	if err == nil {
		t.Error("expected error for existing function file, got nil")
	}

	specPath := filepath.Join(dir, "spec", "functions", "myfunc_spec.rb")
	if _, statErr := os.Stat(specPath); statErr == nil {
		t.Error("spec file should not have been created when function file already exists")
	}
}

// TestNewFunction_RefusesIfSpecFileExists_NoFunctionFileCreated verifies the
// inverse: an existing spec file causes an error and the function file is not
// written. This guards against partial state when only the spec was left behind
// by a previous failed run.
func TestNewFunction_RefusesIfSpecFileExists_NoFunctionFileCreated(t *testing.T) {
	dir := makeModuleDir(t, "myuser", "mymodule")
	specPath := filepath.Join(dir, "spec", "functions", "myfunc_spec.rb")
	if err := os.MkdirAll(filepath.Dir(specPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(specPath, []byte("# existing"), 0644); err != nil {
		t.Fatal(err)
	}

	err := NewFunction(ComponentOptions{
		Name:        "myfunc",
		WorkDir:     dir,
		TemplateDir: makeFunctionTemplateDir(t),
	})
	if err == nil {
		t.Error("expected error for existing spec file, got nil")
	}

	funcPath := filepath.Join(dir, "functions", "myfunc.pp")
	if _, statErr := os.Stat(funcPath); statErr == nil {
		t.Error("function file should not have been created when spec file already exists")
	}
}

// TestNewFunction_NamespacedCreatesCorrectDirectoryStructure verifies that a
// namespaced name like "sub::myfunc" produces the correct nested paths under
// functions/ and spec/functions/, mirroring the behavior of NewClass.
func TestNewFunction_NamespacedCreatesCorrectDirectoryStructure(t *testing.T) {
	dir := makeModuleDir(t, "myuser", "mymodule")
	err := NewFunction(ComponentOptions{
		Name:        "sub::myfunc",
		WorkDir:     dir,
		TemplateDir: makeFunctionTemplateDir(t),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFiles := []string{
		filepath.Join(dir, "functions", "sub", "myfunc.pp"),
		filepath.Join(dir, "spec", "functions", "sub", "myfunc_spec.rb"),
	}
	for _, f := range expectedFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("expected file %s to exist: %v", f, err)
		}
	}
}

// TestNewFunction_NamespacedOutputContainsFullyQualifiedName verifies that a
// namespaced function's output uses the full name "module::sub::function", not
// "module::function" with the intermediate namespace silently dropped.
func TestNewFunction_NamespacedOutputContainsFullyQualifiedName(t *testing.T) {
	dir := makeModuleDir(t, "myuser", "mymodule")
	err := NewFunction(ComponentOptions{
		Name:        "sub::myfunc",
		WorkDir:     dir,
		TemplateDir: makeFunctionTemplateDir(t),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "functions", "sub", "myfunc.pp"))
	if err != nil {
		t.Fatalf("could not read function file: %v", err)
	}
	if !contains(string(content), "mymodule::sub::myfunc") {
		t.Errorf("function file content %q does not contain fully-qualified name %q", string(content), "mymodule::sub::myfunc")
	}
}

// contains is a small helper to avoid importing strings in test context where
// it may not already be imported.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

// makeTaskTemplateDir creates a temp directory with minimal task templates,
// necessary because NewTask calls newRenderer internally and cannot accept an
// injected renderer.
func makeTaskTemplateDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	taskDir := filepath.Join(dir, "task")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(taskDir, "task.sh"), []byte("#!/bin/bash\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(taskDir, "metadata.json"), []byte(`{"description":"","parameters":{}}`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestNewTask_HappyPath verifies that a valid name produces both expected files.
func TestNewTask_HappyPath(t *testing.T) {
	dir := makeModuleDir(t, "myuser", "mymodule")
	err := NewTask(ComponentOptions{
		Name:        "mytask",
		WorkDir:     dir,
		TemplateDir: makeTaskTemplateDir(t),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFiles := []string{
		filepath.Join(dir, "tasks", "mytask.sh"),
		filepath.Join(dir, "tasks", "mytask.json"),
	}
	for _, f := range expectedFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("expected file %s to exist: %v", f, err)
		}
	}
}

// TestNewTask_InitIsValid verifies that the special name "init" is accepted,
// since it maps to the module itself and is a valid PDK task name.
func TestNewTask_InitIsValid(t *testing.T) {
	dir := makeModuleDir(t, "myuser", "mymodule")
	err := NewTask(ComponentOptions{
		Name:        "init",
		WorkDir:     dir,
		TemplateDir: makeTaskTemplateDir(t),
	})
	if err != nil {
		t.Errorf("expected init to be a valid task name, got error: %v", err)
	}
}

// TestNewTask_RejectsInvalidModuleDirectory verifies that NewTask fails before
// touching the filesystem when run outside a valid module directory.
func TestNewTask_RejectsInvalidModuleDirectory(t *testing.T) {
	emptyDir := t.TempDir()
	err := NewTask(ComponentOptions{
		Name:        "mytask",
		WorkDir:     emptyDir,
		TemplateDir: makeTaskTemplateDir(t),
	})
	if err == nil {
		t.Error("expected error for missing metadata.json, got nil")
	}
	if _, statErr := os.Stat(filepath.Join(emptyDir, "tasks")); statErr == nil {
		t.Error("tasks directory should not have been created in an invalid module directory")
	}
}

// TestNewTask_RejectsCorruptMetadata verifies that malformed JSON in
// metadata.json is caught rather than silently ignored.
func TestNewTask_RejectsCorruptMetadata(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), []byte("{{{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	err := NewTask(ComponentOptions{
		Name:        "mytask",
		WorkDir:     dir,
		TemplateDir: makeTaskTemplateDir(t),
	})
	if err == nil {
		t.Error("expected error for corrupt metadata.json, got nil")
	}
}

// TestNewTask_NameValidation covers the full range of invalid name inputs
// against the [a-z][a-z0-9_]* pattern, plus valid edge cases.
func TestNewTask_NameValidation(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "mytask", false},
		{"valid with numbers", "task2", false},
		{"valid with underscores", "my_task", false},
		{"valid with mixed", "my_task_2", false},
		{"init", "init", false},
		{"empty", "", true},
		{"uppercase start", "MyTask", true},
		{"all uppercase", "MYTASK", true},
		{"starts with digit", "2task", true},
		{"starts with underscore", "_task", true},
		{"starts with hyphen", "-task", true},
		{"contains double colon", "my::task", true},
		{"contains colon", "my:task", true},
		{"contains hyphen", "my-task", true},
		{"contains dot", "my.task", true},
		{"contains space", "my task", true},
		{"path separator slash", "my/task", true},
		{"path separator backslash", `my\task`, true},
		{"dot dot", "..", true},
		{"single dot", ".", true},
		{"unicode", "tâche", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := makeModuleDir(t, "myuser", "mymodule")
			err := NewTask(ComponentOptions{
				Name:        tc.input,
				WorkDir:     dir,
				TemplateDir: makeTaskTemplateDir(t),
			})
			if tc.wantErr && err == nil {
				t.Errorf("expected error for name %q, got nil", tc.input)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error for name %q: %v", tc.input, err)
			}
		})
	}
}

// TestNewTask_RefusesIfScriptFileExists_NoMetadataCreated verifies that a
// pre-existing .sh file causes an early error and the .json file is not created
// as a side effect, leaving no partial state on disk.
func TestNewTask_RefusesIfScriptFileExists_NoMetadataCreated(t *testing.T) {
	dir := makeModuleDir(t, "myuser", "mymodule")
	taskDir := filepath.Join(dir, "tasks")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(taskDir, "mytask.sh"), []byte("# existing"), 0644); err != nil {
		t.Fatal(err)
	}

	err := NewTask(ComponentOptions{
		Name:        "mytask",
		WorkDir:     dir,
		TemplateDir: makeTaskTemplateDir(t),
	})
	if err == nil {
		t.Error("expected error for existing task script, got nil")
	}
	if _, statErr := os.Stat(filepath.Join(taskDir, "mytask.json")); statErr == nil {
		t.Error("metadata file should not have been created when script file already exists")
	}
}

// TestNewTask_RefusesIfMetadataFileExists_NoScriptCreated verifies the inverse:
// a pre-existing .json file causes an error and the .sh file is not written.
// This guards against partial state when only the metadata was left behind.
func TestNewTask_RefusesIfMetadataFileExists_NoScriptCreated(t *testing.T) {
	dir := makeModuleDir(t, "myuser", "mymodule")
	taskDir := filepath.Join(dir, "tasks")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(taskDir, "mytask.json"), []byte(`{"description":""}`), 0644); err != nil {
		t.Fatal(err)
	}

	err := NewTask(ComponentOptions{
		Name:        "mytask",
		WorkDir:     dir,
		TemplateDir: makeTaskTemplateDir(t),
	})
	if err == nil {
		t.Error("expected error for existing task metadata, got nil")
	}
	if _, statErr := os.Stat(filepath.Join(taskDir, "mytask.sh")); statErr == nil {
		t.Error("script file should not have been created when metadata file already exists")
	}
}

// TestNewTask_ExistingScriptFileIsUntouched verifies that the content of a
// pre-existing .sh file is not modified when NewTask returns an error.
func TestNewTask_ExistingScriptFileIsUntouched(t *testing.T) {
	dir := makeModuleDir(t, "myuser", "mymodule")
	taskDir := filepath.Join(dir, "tasks")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatal(err)
	}
	original := []byte("# do not touch")
	if err := os.WriteFile(filepath.Join(taskDir, "mytask.sh"), original, 0644); err != nil {
		t.Fatal(err)
	}

	_ = NewTask(ComponentOptions{
		Name:        "mytask",
		WorkDir:     dir,
		TemplateDir: makeTaskTemplateDir(t),
	})

	content, err := os.ReadFile(filepath.Join(taskDir, "mytask.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != string(original) {
		t.Errorf("existing script file was modified: got %q, want %q", string(content), string(original))
	}
}
