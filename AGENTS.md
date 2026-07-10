# AI Agent Instructions for pires-cli

## Purpose
This repository is a Go CLI tool whose main application lives under `app/`. Agents should use this file to understand project structure, build/test commands, and repository conventions before making code changes.

## Key project areas
- `app/` is the main Go module (`go.mod`) and contains the CLI source.
- `app/cmd/` contains Cobra commands and root command definitions.
- `app/internal/config/` manages runtime configuration and environment loading.
- `app/pkg/pireslib/` contains reusable library code used by the CLI.
- `app/CHANGELOG.md` is the release changelog and should be updated for user-facing changes.

## Build and development commands
- Prepare the workspace: `cd app && make prepare`
  - installs required system dependencies and downloads the platform-specific `yq` binary used by the project.
- Build release artifacts: `cd app && make build`
  - builds cross-platform binaries under `app/bin/`.
- For local Go validation, use `cd app && go test ./...`.

## Configuration and runtime behavior
- Configuration is loaded in `app/cmd/root.go` using Viper.
- Supported config file: `app/.env`, or a custom file passed via `--config-file` / `-C`.
- Environment variables use the `CLI_` prefix and mirror long flag names with underscores, e.g. `CLI_GCP_REGION`.
- Root flags are required together: `--environment`, `--gcp-project`, `--gcp-region`.
- `app/main.go` currently performs global checks before Cobra executes, and includes a `TODO` about moving some logic into the root command run function. Treat this as a known code smell rather than a required fix unless the change is clearly scoped.

## Conventions
- Keep the CLI implementation idiomatic Go and follow Cobra command patterns.
- Use `go fmt` before committing Go changes.
- Update `app/CHANGELOG.md` for behavioral or feature changes.
- Update the `CLIVersion` variable in `app/internal/config/config.go` when bumping the CLI version, per `CONTRIBUTING.md`.
- Prefer small, focused changes with clear commit messages.

## Useful documentation
- `CONTRIBUTING.md` for repo setup, release process, and dependency requirements.
- `README.md` and `app/README.md` for CLI usage and runtime expectations.

## Notes for agents
- There are currently no Go test files under `app/` (no `_test.go` files found during inspection).
- Validate any new environment variable handling against the existing Viper config logic in `app/cmd/root.go`.
- Do not assume the repository root contains the Go module; the Go project is under `app/`.
