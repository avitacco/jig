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
| `new`              | `defined_type` | 🔲 Planned |
| `new`              | `fact`         | 🔲 Planned |
| `new`              | `function`     | 🔲 Planned |
| `new`              | `provider`     | 🔲 Planned |
| `new`              | `task`         | 🔲 Planned |
| `new`              | `test`         | 🔲 Planned |
| `new`              | `transport`    | 🔲 Planned |
| `--skip-interview` |                | ✅ Working  |
| Template override  |                | ✅ Working  |
| `templates`        | `dump`         | ✅ Working  |
| `templates`        | `resolve`      | 🔲 Planned |
| `build`            |                | 🔲 Planned |
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
    manifests/
      class.pp
    spec/
      classes/
        class_spec.rb
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

### Design notes for contributors

- Templates are embedded via `go:embed`. External templates take precedence over embedded ones, with per-file fallback to embedded defaults when a custom template is not found.
- `--force` never deletes existing files outright. It creates a timestamped backup of the target directory first.
- Module name validation uses a `ValidationResult` type with an iota-based `Severity`. Violations at the `Warning` level do not halt execution.
- Config is handled with [Viper](https://github.com/spf13/viper).

## NOTICE

Some default template files included in this project are derived from the
[pdk-templates](https://github.com/puppetlabs/pdk-templates) project,
copyright Puppet Labs, and are used under the terms of the
[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0).

## License

See [LICENSE](LICENSE).