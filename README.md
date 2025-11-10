# PRGuard

PRGuard helps open source maintainers detect, block, and manage spam pull requests on GitHub. It analyzes PR quality using configurable heuristics, maintains a portable blocklist, and provides automation for closing spam PRs and blocking abusive users.

## Features

- 🔍 **Automatic Spam Detection**: Analyze PRs for spam indicators like single-file README edits, new accounts, and minimal changes
- 🚫 **Blocklist Management**: Maintain a local database of problematic users with reasons and evidence
- 📤 **Import/Export**: Share blocklists with other maintainers via JSON or CSV
- 🤖 **GitHub Integration**: Close PRs, add labels, and block users directly via the GitHub API
- ⚙️ **Configurable**: Customize detection thresholds and whitelists to fit your project

## Installation

### From Source

```bash
git clone https://github.com/prguard/prguard.git
cd prguard
go build -o prguard ./cmd/prguard
```

### Binary Releases

Coming soon - binaries will be available via GitHub Releases.

## Quick Start

1. **Create a configuration file**:

PRGuard looks for a config file in these locations (in order):
- `~/.config/prguard/config.yaml` (recommended)
- `./config.yaml` (current directory)
- Custom path via `--config` flag

```bash
# Create config directory
mkdir -p ~/.config/prguard

# Copy example config
cp config.example.yaml ~/.config/prguard/config.yaml
```

2. **Edit your config file** and add your GitHub token:

```yaml
github:
  token: "YOUR_GITHUB_TOKEN_HERE"
  org: "your-org-name"  # or use 'user' instead

database:
  type: "sqlite"
  path: "~/.local/prguard/prguard.db"

# Optional: List repositories to monitor
repositories:
  - owner: "your-org"
    name: "repo-1"
  - owner: "your-org"
    name: "repo-2"
```

3. **Scan a repository**:

```bash
./prguard scan owner/repo
```

Or scan all configured repositories at once:

```bash
./prguard scan-all
```

4. **Block a spammer**:

```bash
./prguard block username --reason "Spam PRs" --evidence https://github.com/owner/repo/pull/123
```

5. **Close spam PRs**:

```bash
./prguard close-pr owner/repo 123 456
```

## Commands

- `scan <owner>/<repo>` - Analyze PRs in a single repository for spam
- `scan-all` - Scan all repositories configured in config.yaml
- `block <username>` - Add a user to the blocklist
- `unblock <username>` - Remove a user from the blocklist
- `check <username>` - Check if a user is blocked
- `list` - List all blocklist entries
- `export` - Export blocklist to JSON or CSV
- `import` - Import blocklist from a file or URL
- `close-pr <owner>/<repo> <pr-number>...` - Close spam PRs
- `review <owner>/<repo>` - Show PRs needing manual review

Run `./prguard --help` or `./prguard <command> --help` for detailed usage information.

## Configuration

PRGuard uses a YAML configuration file. See [config.example.yaml](config.example.yaml) for a complete example.

### Key Configuration Options

- **GitHub Token**: Required for API access (needs `repo` and `write:discussion` permissions)
- **Database Path**: Defaults to `~/.local/prguard/prguard.db` for SQLite
- **Detection Thresholds**: Customize `min_files`, `min_lines`, `account_age_days`
- **Whitelist**: Trusted contributors who bypass spam detection
- **Default Actions**: Automatically close PRs and add labels

### Directory Structure

PRGuard follows XDG Base Directory conventions:
- **Config**: `~/.config/prguard/config.yaml`
- **Database**: `~/.local/prguard/prguard.db`
- **Exports**: Configurable via `blocklist.export_path` (default: `./exports`)

### Environment Variables

All config values can be overridden with environment variables prefixed with `PRGUARD_`:

```bash
export PRGUARD_GITHUB_TOKEN="your-token"
export PRGUARD_DATABASE_PATH="./prguard.db"
```

## Spam Detection Heuristics

PRGuard automatically flags PRs as spam if they meet these criteria:

1. **Single-file README edits**: Only one file modified and it's a README
2. **Account age**: GitHub account created within the last 7 days (configurable)
3. **Minimal changes**: Fewer than 2 files or 10 lines changed (configurable)
4. **Spam phrases**: Contains known spam patterns (configurable)

PRs with some but not all indicators are marked for manual review.

## Blocklist Sharing

Export your blocklist to share with other maintainers:

```bash
./prguard export --format json --output my-blocklist.json
```

Import a trusted blocklist:

```bash
./prguard import --file community-blocklist.json
# or from a URL
./prguard import --url https://example.com/blocklist.json
```

## Development

### Project Structure

```
prguard/
├── cmd/prguard/        # Main CLI entry point
├── internal/
│   ├── blocklist/      # Blocklist management
│   ├── commands/       # CLI command implementations
│   ├── config/         # Configuration parsing
│   ├── database/       # Database operations
│   ├── github/         # GitHub API client
│   └── scanner/        # PR quality detection
├── pkg/models/         # Data models
└── docs/               # Documentation
```

### Building

```bash
go build -o prguard ./cmd/prguard
```

### Testing

```bash
go test ./...
```

## Roadmap

- [ ] Unit and integration tests
- [ ] GitHub Actions integration
- [ ] Turso (remote database) support
- [ ] Pattern matching for spam phrases
- [ ] Batch operations for multiple PRs
- [ ] Audit logging
- [ ] Statistics and reporting

See [docs/tasks.md](docs/tasks.md) for detailed development progress.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

## Disclaimer

PRGuard is a tool to assist with spam management. Always review flagged PRs manually before taking action. False positives may occur, so use appropriate thresholds and whitelists for your project.
