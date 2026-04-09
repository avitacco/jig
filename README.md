# jig

A Go-based reimplementation of the [Puppet Development Kit (PDK)](https://github.com/puppetlabs/pdk), built to be fast, self-contained, and free of Ruby runtime dependencies.

## Why jig?

PDK has been an essential tool for Puppet module authors for years. When
Perforce moved PDK to a closed-source model, it created a real problem for teams
and individuals who depend on open tooling for their workflows. On top of that,
PDK carries a heavy Ruby runtime footprint, which adds friction to CI
environments and developer machines alike.

jig aims to replace the parts of PDK that matter most: scaffolding new modules,
building module packages, and cutting releases. It ships as a single static
binary with no external runtime required.

## Status

jig is under active development. The table below reflects the current state of
planned functionality.

| Command            | Subcommand     | Status     |
|--------------------|----------------|------------|
| `new`              | `module`       | ✅ Working  |
| `new`              | `class`        | ✅ Working  |
| `new`              | `defined_type` | ✅ Working  |
| `new`              | `fact`         | ✅ Working  |
| `new`              | `function`     | ✅ Working  |
| `new`              | `provider`     | ✅ Working  |
| `new`              | `task`         | ✅ Working  |
| `new`              | `test`         | 🔲 Planned |
| `new`              | `transport`    | 🔲 Planned |
| `--skip-interview` |                | ✅ Working  |
| Template override  |                | ✅ Working  |
| `templates`        | `dump`         | ✅ Working  |
| `templates`        | `resolve`      | 🔲 Planned |
| `build`            |                | ✅ Working  |
| `release`          |                | 🔲 Planned |

## Installation

### Build from source

Requires Go 1.21 or later.
```bash
git clone https://github.com/avitacco/jig.git
cd jig
go build -o jig .
```

Move the resulting binary somewhere in your `$PATH`:
```bash
mv jig /usr/local/bin/
```

No other dependencies or runtimes needed.

## Usage

### `jig new module`

Scaffolds a new Puppet module with the standard directory structure and
metadata.
```
jig new module <n> [flags]
```

jig will walk you through an interactive interview to collect module metadata.
Values from your config file are used as defaults. If no config is present,
jig falls back to your system username and full name.

**Flags:**

| Flag | Description |
|------|-------------|
| `-u, --forge-user` | Your Puppet Forge username |
| `-a, --author` | Your full name |
| `-l, --license` | License type (default: from config, then `Apache-2.0`) |
| `-s, --summary` | One-line module summary |
| `-S, --source` | Source URL for the module |
| `-f, --force` | Overwrite an existing module directory. The existing directory is backed up with a timestamp before any files are written. |
| `-i, --skip-interview` | Skip the interactive interview and use flag values or defaults. |

### `jig new class`

Generates a new Puppet class manifest and its rspec-puppet spec file inside
the current module directory.
```
jig new class <n>
```

The class name follows standard Puppet naming conventions. Namespaced names
like `foo::bar` are supported and will generate the correct directory structure
under `manifests/`. The module name prefix must not be included in the name.

### `jig new defined_type`

Generates a new Puppet defined type manifest and its rspec-puppet spec file
inside the current module directory.
```
jig new defined_type <n>
```

Defined type names follow the same conventions as class names. Namespaced
names like `foo::bar` are supported and will generate the correct directory
structure under `manifests/`. The module name prefix must not be included in
the name.

### `jig new fact`

Generates a new custom Facter fact and its spec file inside the current module
directory.
```
jig new fact <n>
```

Fact names may not contain `::`. The generated fact is placed in
`lib/facter/<name>.rb` and its spec in `spec/unit/facter/<name>_spec.rb`.

### `jig new function`

Generates a new Puppet language function and its spec file inside the current
module directory.
```
jig new function <n>
```

Function names follow standard Puppet naming conventions. The module name is
automatically prepended to form the fully qualified function name
(`<module>::<name>`). The generated function is placed in
`functions/<name>.pp` and its spec in `spec/functions/<name>_spec.rb`.

### `jig new provider`

Generates a new Puppet resource type and provider using the
[Resource API](https://github.com/puppetlabs/puppet-resource_api), along with
spec files for both, inside the current module directory.
```
jig new provider <n>
```

Provider names must start with a lowercase letter and contain only lowercase
letters, numbers, and underscores (`[a-z][a-z0-9_]*`). Four files are
generated:

- `lib/puppet/type/<name>.rb` — the Resource API type definition
- `lib/puppet/provider/<name>/<name>.rb` — the Resource API simple provider
- `spec/unit/puppet/type/<name>_spec.rb` — spec file for the type
- `spec/unit/puppet/provider/<name>/<name>_spec.rb` — spec file for the provider

### `jig new task`

Generates a new Puppet task and its metadata file inside the current module
directory.
```
jig new task <n>
```

Task names must start with a lowercase letter and contain only lowercase
letters, numbers, and underscores (`[a-z][a-z0-9_]*`). The special name
`init` is valid and maps the task to the module itself. Namespaced names
using `::` are not valid for tasks.

**Flags on `jig new`:**

The following flag is available on all `jig new` subcommands:

| Flag | Description |
|------|-------------|
| `-t, --template-dir` | Path to a custom template directory. See [Template Overrides](#template-overrides) below. |

**Global flags:**

| Flag | Description |
|------|-------------|
| `--config` | Path to config file |
| `--debug` | Enable debug output |

**Module naming:** jig validates module names against Puppet's naming
conventions. Violations produce a warning but do not stop scaffolding.

### `jig templates dump`

Extracts all embedded default templates to a directory on disk. This is useful
as a starting point for creating your own custom templates. If the destination
directory already exists it will be renamed with a timestamp suffix before
writing.
```
jig templates dump <destination>
```

For example:
```bash
jig templates dump ~/.config/jig/templates
```

You can then edit the files in the destination directory and point jig at them
using `--template-dir` or the `template_dir` config key.

## Template Overrides

jig embeds default templates for all generated files. If you want to customise
them, you can point jig at a directory of your own templates. Any template
found in your custom directory takes precedence over the embedded default.
Templates not present in your custom directory fall back to the embedded
defaults automatically, so you only need to include the files you want to
change.

The easiest way to get started is to run `jig templates dump` to extract the
default templates, then edit the ones you want to change.

### Template directory structure

Your custom template directory must mirror the structure of jig's embedded
templates:
```
templates/
  common/
    gitkeep
  module/
    manifests/
      init.pp
    spec/
      class_spec.rb
      spec_helper.rb
      default_facts.yml
    Gemfile
    Rakefile
    README.md
    CHANGELOG.md
    gitignore
    pdkignore
    rubocop.yml
    hiera.yaml
  class/
    class.pp
    class_spec.rb
  type/
    defined_type.pp
    defined_type_spec.rb
  fact/
    fact.rb
    fact_spec.rb
  function/
    function.pp
    function_spec.rb
  provider/
    type.rb
    type_spec.rb
    provider.rb
    provider_spec.rb
  task/
    task.sh
    metadata.json
```

### Configuring the template directory

There are three ways to tell jig where your custom templates live, in order
of precedence:

**Command line flag:**
```bash
jig new --template-dir /path/to/templates module mymodule
```

**Environment variable:**
```bash
export JIG_TEMPLATE_DIR=/path/to/templates
jig new module mymodule
```

**Config file** (`~/.config/jig/config.toml`):
```toml
template_dir = "/path/to/templates"
```

## Configuration

jig looks for a config file at `~/.config/jig/config.toml`. All fields are
optional. If the file does not exist, jig falls back to sensible defaults.
```toml
forge_username = "avitacco"
author         = "John Doe"
license        = "Apache-2.0"
forge_token    = "your-forge-token"
template_dir   = "/path/to/templates"
```

The config path can be overridden with the `--config` flag or the
`JIG_CONFIG` environment variable.

## Contributing

Contributions are welcome. The project is in early stages, so the best place to
start is by opening an issue to discuss what you want to work on before sending
a PR.

### Project layout
```
.
├── main.go
├── commands/        # Cobra command definitions
└── internal/
    ├── build/
    ├── config/
    ├── forge/
    ├── module/      # Module metadata and validation
    ├── release/
    ├── scaffold/    # Scaffolding orchestration
    └── template/   # Template rendering with fallback logic
        └── templates/  # Embedded default templates
```

### Testing

Run the full test suite with:
```bash
go test ./...
```

Tests live alongside the source files they cover (`*_test.go`), which is
the standard Go convention. The `commands/` and `internal/config/` packages
do not currently have tests -- the former is thin Cobra wiring and the latter
is thin Viper wiring, so the internal packages are where the meaningful
coverage lives.

A few patterns used throughout the test suite that contributors should follow:

- **Table-driven tests** for functions with multiple input variations. Use a
  `cases := []struct{...}` slice and `t.Run` for each case.
- **`t.TempDir()`** for any test that touches the filesystem. It is cleaned up
  automatically after the test and requires no `defer os.Remove`.
- **`fakeRenderer`** in `internal/scaffold` implements the `scaffold.Renderer`
  interface and can be used to test template rendering paths without hitting
  the real embedded templates.
- **`makeBuildDir`** in `internal/build` and **`makeModuleDir`** in
  `internal/scaffold` are shared helpers that create realistic on-disk module
  structures for tests that need them.
- Both characterization tests (pinning current behavior) and adversarial tests
  (checking rejection of invalid or malicious input) are expected. When adding
  a new feature, include both.

### Design notes for contributors

- Templates are embedded via `go:embed`. External templates take precedence
  over embedded ones, with per-file fallback to embedded defaults when a custom
  template is not found. Template names are validated to prevent path traversal
  before any file is read.
- `--force` never deletes existing files outright. It creates a timestamped
  backup of the target directory first.
- Module name validation uses a `ValidationResult` type with an iota-based
  `Severity`. Violations at the `Warning` level do not halt execution. Version
  strings must be valid semver (`MAJOR.MINOR.PATCH`).
- Component names (module names, class names, defined type names) are validated
  to reject empty strings, path separators, and traversal sequences before they
  are used to construct filesystem paths.
- `os.Getwd()` is called only in the `commands/` layer. Internal packages
  receive directory paths as arguments, which keeps them testable without
  manipulating the process working directory.
- Config is handled with [Viper](https://github.com/spf13/viper).

## NOTICE

Some default template files included in this project are derived from the
[pdk-templates](https://github.com/puppetlabs/pdk-templates) project,
copyright Puppet Labs, and are used under the terms of the
[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0).

## License

See [LICENSE](LICENSE).