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
| Template override  |                | 🔲 Planned |
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
jig new module <name> [flags]
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

**Global flags:**

| Flag | Description |
|------|-------------|
| `--config` | Path to config file |
| `--debug` | Enable debug output |

**Module naming:** jig validates module names against Puppet's naming
conventions. Violations produce a warning but do not stop scaffolding.

## Configuration

jig looks for a config file at `~/.config/jig/config.toml`. All fields are
optional. If the file does not exist, jig falls back to sensible defaults.

```toml
forge_username = "avitacco"
author         = "John Doe"
license        = "Apache-2.0"
forge_token    = "your-forge-token"
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

- Templates are embedded via `go:embed`. External templates in `~/.config/jig/templates/` take precedence.
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
