# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PRGuard is a CLI tool for detecting and managing spam pull requests on GitHub. It uses configurable heuristics to identify spam PRs, maintains a portable blocklist of problematic users, and provides automation for closing spam PRs.

**Key Technologies:**
- Go 1.21+
- SQLite (via go-sqlite3) for local blocklist storage
- GitHub REST API (via google/go-github/v57)
- Cobra for CLI framework
- YAML for configuration

## Build & Development Commands

```bash
# Build the binary
make build
# or
go build -o prguard ./cmd/prguard

# Run tests
make test
# or
go test ./...

# Run linters
make lint

# Install to $GOPATH/bin
make install

# Clean build artifacts
make clean
```

## Architecture

### Core Components Flow

1. **Config Loading** (`internal/config/config.go`)
   - Searches for config in: `~/.config/prguard/config.yaml`, `./config.yaml`, or custom path
   - Supports env variable overrides with `PRGUARD_*` prefix
   - Validates GitHub token, database settings, and filter thresholds

2. **Command Initialization** (`internal/commands/common.go`)
   - All commands use `initClients()` which returns: Config, GitHub client, Blocklist manager, and Database
   - Database auto-creates directories (e.g., `~/.local/prguard/`)

3. **PR Scanning Flow** (`internal/scanner/scanner.go`)
   - Fetches PR details and file changes via GitHub API
   - Applies heuristics: README-only edits, account age, minimal changes, spam phrases
   - Returns `ScanResult` with `IsSpam`, `IsUncertain`, or clean status
   - `scan-all` command iterates over `repositories` list in config

4. **Blocklist Management** (`internal/blocklist/blocklist.go`)
   - Uses UUID as primary key for deduplication across imports
   - Import/export supports JSON and CSV formats
   - Import merge strategy: union of entries, keeps highest severity on conflicts
   - All entries track: username, reason, evidence URL, severity (low/medium/high), source (manual/imported/auto-detected)

5. **Database Layer** (`internal/database/database.go`)
   - SQLite with schema defined in `schema.go`
   - CRUD operations: Add, Get, List, Remove, Update entries
   - `IsBlocked()` checks username existence

6. **GitHub Integration** (`internal/github/client.go`)
   - Wraps go-github client with OAuth2 token auth
   - Key methods: `GetPullRequests()`, `GetUser()`, `ClosePullRequest()`, `AddLabel()`, `BlockUser()`
   - Returns structured `PullRequest` and `User` types

### Directory Structure (XDG Conventions)

- **Config**: `~/.config/prguard/config.yaml`
- **Database**: `~/.local/prguard/prguard.db` (auto-created)
- **Exports**: Configurable via `blocklist.export_path` (default: `./exports`)

### Spam Detection Logic

The scanner (`internal/scanner/scanner.go`) applies these rules:

- **Definite Spam**: Single README-only file edit + (new account OR spam phrases)
- **Uncertain**: Meets some criteria (e.g., new account + minimal changes, but not README-only)
- **Whitelisted users bypass all checks** (e.g., `dependabot[bot]`)

Key methods:
- `ScanPR()` - Analyzes single PR
- `ScanRepository()` - Scans all open PRs in a repo
- Heuristics: `isSingleFileReadmeEdit()`, `isNewAccount()`, `isMinimalChanges()`, `containsSpamPhrases()`

### Configuration Schema

The config struct (`internal/config/config.go`) includes:

- `GitHub`: token, org/user
- `Database`: type (sqlite/turso), path/url
- `Repositories`: list of owner/name pairs (for `scan-all`)
- `Filters`: min_files, min_lines, account_age_days, whitelist, spam_phrases
- `Blocklist`: auto_export, export_path, sources (remote blocklists)
- `Actions`: close_prs, add_spam_label, comment_template

All config values can be overridden via `PRGUARD_*` env vars.

### Adding New Commands

1. Create command file in `internal/commands/` (e.g., `new_command.go`)
2. Implement `NewXCommand(configPath *string) *cobra.Command`
3. Use `initClients()` to get config, GitHub client, blocklist manager, and database
4. Register in `cmd/prguard/main.go` via `rootCmd.AddCommand()`

### Testing Strategy (TODO)

Per `docs/tasks.md`, tests are planned for:
- Database CRUD operations
- Scanner heuristics
- Import/export functionality
- GitHub API mocking

When implementing tests, mock the GitHub client and use in-memory SQLite (`:memory:`).

## Important Patterns

- **Error Handling**: Commands return errors, main handles exit codes
- **Database Cleanup**: Always `defer db.Close()` after `initClients()`
- **User Feedback**: Commands print structured output with status indicators (✓)
- **Repo Format**: Commands expect `owner/repo` format, parsed by `parseRepo()` in `internal/commands/scan.go`

## Development Roadmap

See `docs/tasks.md` for current status. MVP is complete. Next priorities:
- Unit and integration tests
- Turso database support
- CSV import functionality
- GitHub Actions integration
