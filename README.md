<div align="center">

# `goproject`

[![License](https://img.shields.io/github/license/l50/goproject?label=License&style=flat&color=blue&logo=github)](https://github.com/l50/goproject/blob/main/LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/l50/goproject?logo=go)](https://go.dev/)
[![Go CI](https://github.com/l50/goproject/actions/workflows/go-ci.yaml/badge.svg)](https://github.com/l50/goproject/actions/workflows/go-ci.yaml)
[![Semgrep](https://github.com/l50/goproject/actions/workflows/semgrep.yaml/badge.svg)](https://github.com/l50/goproject/actions/workflows/semgrep.yaml)
[![Renovate](https://github.com/l50/goproject/actions/workflows/renovate.yaml/badge.svg)](https://github.com/l50/goproject/actions/workflows/renovate.yaml)

</div>

---

`goproject` is a template for new Go command line tools. It ships a working
[Cobra](https://github.com/spf13/cobra) CLI with
[Viper](https://github.com/spf13/viper) configuration, leveled logging, tests,
and the build, lint, CI, release, and dependency tooling used by
[warpgate](https://github.com/CowDogMoo/warpgate) and
[squad](https://github.com/CowDogMoo/squad).

## Table of Contents

- [Using This Template](#using-this-template)
- [Repository Layout](#repository-layout)
- [Configuration](#configuration)
- [Development](#development)
- [Releasing](#releasing)
- [License](#license)

## Using This Template

1. Create a new repository from this template and clone it.

1. Rename the module, binary, and environment prefix. Replace `OWNER/NAME`
   and `NAME` with your repository and binary names:

   ```bash
   git grep -lz -e goproject -e GOPROJECT | xargs -0 perl -pi -e \
     's|l50/goproject|OWNER/NAME|g; s|goproject|NAME|g; s|GOPROJECT|NAME_IN_CAPS|g'
   git mv cmd/goproject cmd/NAME
   ```

1. Update the owner in `.github/CODEOWNERS` and the copyright holder in
   `LICENSE`.

1. Confirm everything still builds, lints, and tests clean:

   ```bash
   task
   task pre-commit:run-hooks
   ```

1. Add your commands under `cmd/NAME` and your packages at the repository
   root, next to `config/` and `logging/`.

## Repository Layout

| Path                 | Purpose                                                              |
| -------------------- | -------------------------------------------------------------------- |
| `cmd/goproject`      | Cobra CLI: root command, `version`, and `completion`                 |
| `config`             | Viper-backed config loading with defaults, XDG lookup, env overrides |
| `logging`            | Leveled console logger with text, color, and JSON output             |
| `.hooks`             | Scripts run by pre-commit and CI                                     |
| `.github/workflows`  | Go CI, govulncheck, Semgrep, GoReleaser, Renovate, and label sync    |
| `Taskfile.yaml`      | Task runner entry point, built on the shared CowDogMoo templates     |
| `.goreleaser.yaml`   | Release builds and changelog                                         |

## Configuration

Settings resolve in this order, highest first:

1. CLI flags (`--log-level`, `--log-format`, `--quiet`, `--verbose`)
1. Environment variables prefixed with `GOPROJECT_` (`GOPROJECT_LOG_LEVEL`)
1. A config file: `--config <path>`, or the first `config.yaml` found in
   `$XDG_CONFIG_HOME/goproject`, `~/.goproject`, `$XDG_CONFIG_DIRS/goproject`,
   or the current directory
1. Built-in defaults

```yaml
log:
  level: info # debug, info, warn, error
  format: text # text, color, json
```

## Development

Install [Go](https://go.dev/doc/install) (the version in `go.mod`),
[Task](https://taskfile.dev/installation/), and
[pre-commit](https://pre-commit.com/#install), then:

```bash
task pre-commit:install-pc-hooks # install the git hooks
task                             # clean, build, and test
task go:lint                     # golangci-lint
task pre-commit:run-hooks        # every hook against every file
```

Run `task --list` for every available task. Tests run with
`bash .hooks/run-go-tests.sh all`; CI uses `coverage` to write
`coverage-all.out` for Codecov.

## Releasing

Pushing a tag runs GoReleaser, which builds the binaries and publishes a
GitHub release with a grouped changelog:

```bash
task go:release TAG=v1.0.0
```

Use `task go:release-test` for a local snapshot build that publishes nothing.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
