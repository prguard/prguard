# Software Specification Document: PRGuard

**Project Name:** PRGuard  
**Version:** 1.0  
**Date:** November 10, 2025  
**Status:** Draft

---

## 1. Executive Summary

PRGuard is a tool that helps open source maintainers detect, block, and manage spam pull requests on GitHub. It analyzes PR quality using configurable heuristics, maintains a portable blocklist, and provides automation for closing spam PRs and blocking abusive users.

---

## 2. Goals and Objectives

### Primary Goals
- Detect low-quality and spam pull requests automatically
- Maintain a decentralized, shareable blocklist of problematic users
- Provide maintainers with tools to quickly close spam PRs and block users
- Enable community-driven blocklist sharing without centralized control

### Non-Goals
- This is not a moderation platform or social network
- Not designed to replace GitHub's built-in abuse reporting
- Not intended for general-purpose PR review automation

---

## 3. Technical Architecture

### 3.1 Language & Runtime
- **Language:** Go (1.21+)
- **Distribution:** Single compiled binary
- **Configuration:** External YAML or TOML file (not embedded in binary)

### 3.2 Data Storage
- **Primary Database:** libSQL/Turso
  - Local SQLite mode for single users
  - Remote Turso mode for teams/organizations
- **Blocklist Schema:**
  - UUID (primary key for deduplication)
  - GitHub username
  - Reason for blocking
  - Evidence URL (link to problematic PR/issue)
  - Timestamp
  - Blocked by (maintainer who added entry)
  - Severity level (low/medium/high)
  - Source (manual/imported/auto-detected)
  - Metadata JSON field for extensibility

### 3.3 External Integrations
- **GitHub API:** REST API for PR analysis, user data, blocking
- **Authentication:** Personal Access Token or GitHub App

---

## 4. Feature Specifications

### 4.1 PR Quality Detection

#### Low-Quality PR Heuristics

**Automatic Block Triggers:**
1. **Single-file README edits**
   - Only one file modified
   - File is README.md or similar documentation
   - Changes are minimal (< 10 lines or configurable threshold)

2. **Account age check**
   - GitHub account created within last 7 days (configurable)
   - Combined with other suspicious indicators

3. **Pattern matching**
   - Common spam phrases (configurable)
   - Link-only changes without context
   - Duplicate/copy-paste descriptions

**Uncertain Case Handling:**
- Generate a "review needed" report for maintainers
- Flag PRs that meet some but not all criteria
- Prompt maintainer with evidence and recommendation
- Maintainer can approve/block with single command

**Configuration Options:**
- Threshold for minimum file changes
- Threshold for minimum line changes
- Account age threshold
- Whitelist for trusted contributors
- Custom pattern rules

### 4.2 Blocklist Management

#### Core Functionality
- **Add Entry:** Manually block a user with reason and evidence
- **Check Status:** Query if a user is blocked
- **List Entries:** View all blocked users with filters
- **Remove Entry:** Unblock a user
- **Update Entry:** Modify reason or severity

#### Import/Export
- **Export Formats:**
  - JSON (full metadata, recommended for sharing)
  - CSV (simple format for spreadsheets)
- **Import Sources:**
  - Local file (JSON/CSV)
  - Remote URL (HTTP/HTTPS)
  - GitHub Gist
- **Deduplication:** Use UUID to prevent duplicate entries
- **Merge Strategy:** Union of entries, keep most severe rating on conflicts

#### Blocklist Sources
- Subscribe to trusted maintainer blocklists
- Auto-sync on schedule or manual trigger
- Track source provenance for each entry
- Enable/disable sources without removing data

### 4.3 PR Management

#### Close Spam PRs
- Close one or multiple PRs by number
- Add standardized comment explaining closure
- Optionally label PR as "spam"
- Batch operation for multiple PRs from same user

#### Block User Workflow
1. Detect or identify spam PR
2. Add user to blocklist
3. Close all open PRs from that user (optional)
4. Block user at organization level via GitHub API (optional)
5. Generate audit log

### 4.4 CLI Commands

#### Required Commands
- `scan` - Analyze PRs in a repository
- `block` - Add user to blocklist
- `unblock` - Remove user from blocklist
- `check` - Check if user is blocked
- `list` - List blocklist entries
- `export` - Export blocklist to file
- `import` - Import blocklist from file
- `close-pr` - Close spam PR(s)
- `sync` - Sync with remote blocklist sources
- `review` - Show PRs needing manual review

#### Optional Commands
- `stats` - Show blocklist statistics
- `audit` - Show action history
- `config` - Validate or display config

---

## 5. User Workflows

### 5.1 Initial Setup
1. Install PRGuard binary
2. Create configuration file from template
3. Add GitHub token
4. Initialize database (local or Turso)
5. Optionally import community blocklist

### 5.2 Daily Usage - Reactive
1. Receive notification of new PR
2. Run `prguard scan` command on repository
3. Review flagged PRs
4. For confirmed spam:
   - Run `prguard block` command
   - Run `prguard close-pr` command
   - Optionally export updated blocklist

### 5.3 Daily Usage - Proactive
1. Run scheduled scan via cron/GitHub Actions
2. Receive report of suspicious PRs
3. Review uncertain cases
4. Block confirmed spammers
5. Close spam PRs in batch

### 5.4 Blocklist Sharing
1. Maintainer A exports blocklist: `prguard export --format json`
2. Maintainer A shares file via GitHub/Gist/email
3. Maintainer B imports: `prguard import --file blocklist.json`
4. Both maintainers benefit from shared knowledge

---

## 6. Configuration Specification

### 6.1 Configuration File Structure

**Required Sections:**
- `github` - API token, org/user, rate limiting
- `database` - Type (sqlite/turso), connection details
- `filters` - PR quality heuristics and thresholds
- `blocklist` - Import sources, auto-sync settings
- `actions` - Default behaviors (close PRs, add labels)

**Environment Variable Overrides:**
- All config values should support env var overrides
- Prefix: `PRGUARD_` 
- Useful for CI/CD and secrets management

**Example Config Keys:**
- `github.token` (or `PRGUARD_GITHUB_TOKEN`)
- `github.org` or `github.user`
- `filters.min_files` (default: 2)
- `filters.min_lines` (default: 10)
- `filters.account_age_days` (default: 7)
- `filters.readme_only_block` (default: true)
- `database.type` (sqlite or turso)
- `database.path` or `database.url`
- `blocklist.auto_export` (boolean)
- `blocklist.sources` (array of URLs)

### 6.2 Configuration Validation
- Validate on startup
- Clear error messages for missing/invalid values
- Fail fast if critical config is missing

---

## 7. GitHub API Integration

### 7.1 Required Permissions
- `repo` - Read PR data, close PRs
- `admin:org` - Block users at org level (optional)
- `write:discussion` - Comment on PRs

### 7.2 Rate Limiting
- Respect GitHub API rate limits
- Implement exponential backoff
- Cache data where appropriate
- Allow configurable rate limit settings

### 7.3 API Operations
- Fetch PR details (files changed, lines, author)
- Fetch user account data (creation date, contributions)
- Close PR with comment
- Add labels to PR
- Block user at org level
- Check if user is already blocked

---

## 8. Data Privacy & Security

### 8.1 Token Storage
- Never store tokens in database
- Read from config file or environment variables
- Warn if token has excessive permissions
- Support GitHub App authentication for better security

### 8.2 Blocklist Data
- Store only public GitHub usernames
- Include evidence URLs (public PR/issue links)
- No personal information beyond public GitHub data
- UUID ensures entries are portable without conflicts

### 8.3 Export Security
- Exports contain no authentication tokens
- Exports are shareable by default
- Include metadata about export source for transparency

---

## 9. Extensibility

### 9.1 Plugin Architecture (Future)
- Custom heuristics via plugins
- External ML models for spam detection
- Integration with other platforms (GitLab, Bitbucket)

### 9.2 Webhook Support (Future)
- Listen for GitHub webhooks
- Automatic real-time PR scanning
- Deploy as service for continuous monitoring

### 9.3 API (Future)
- REST API for programmatic access
- Integration with other tools
- Dashboard for visualization

---

## 10. Testing Strategy

### 10.1 Unit Tests
- Database operations (CRUD for blocklist)
- PR quality heuristics
- Import/export functionality
- Configuration parsing

### 10.2 Integration Tests
- GitHub API mocking
- End-to-end command testing
- Database migrations

### 10.3 Manual Testing
- Test against real repositories (with permission)
- Test import/export between instances
- Verify GitHub API blocking works correctly

---

## 11. Documentation Requirements

### 11.1 User Documentation
- Installation guide
- Quick start tutorial
- Command reference
- Configuration guide
- Troubleshooting guide

### 11.2 Developer Documentation
- Architecture overview
- Database schema
- Adding new heuristics
- Contributing guide

### 11.3 Community Guidelines
- How to share blocklists responsibly
- Evidence requirements for blocklist entries
- Appeal process for false positives

---

## 12. Deployment & Distribution

### 12.1 Binary Distribution
- Single executable for major platforms (Linux, macOS, Windows)
- GitHub Releases with checksums
- Homebrew formula (macOS/Linux)
- Installation script

### 12.2 GitHub Action
- Pre-built action for CI/CD workflows
- Automatic PR scanning on open
- Comment on suspicious PRs

### 12.3 Docker Image (Optional)
- Containerized version for server deployment
- Useful for webhook-based monitoring

---

## 13. Success Metrics

### 13.1 User Metrics
- Number of installations
- Number of PRs scanned
- Number of spam PRs detected and closed
- False positive rate (via user feedback)

### 13.2 Community Metrics
- Number of shared blocklists
- Number of imported entries
- Community contributions to heuristics

---

## 14. Risks & Mitigations

### 14.1 False Positives
**Risk:** Legitimate contributors blocked by mistake  
**Mitigation:**
- Conservative default thresholds
- Manual review for uncertain cases
- Easy unblock process
- Track source of blocks for accountability

### 14.2 Blocklist Abuse
**Risk:** Malicious actors sharing poisoned blocklists  
**Mitigation:**
- Provenance tracking for all imports
- Trust levels for blocklist sources
- Easy way to audit and remove entries
- Community reporting for bad blocklists

### 14.3 GitHub API Changes
**Risk:** Breaking changes to GitHub API  
**Mitigation:**
- Use stable API versions
- Comprehensive error handling
- Version compatibility checks
- Active maintenance commitment

### 14.4 Privacy Concerns
**Risk:** Blocklists could be used for harassment  
**Mitigation:**
- Only public GitHub data
- Require evidence URLs
- Clear community guidelines
- Focus on behavior, not identity

---

## 15. Future Enhancements

### Phase 2 (After Initial Release)
- Machine learning model for spam detection
- Web dashboard for visualization
- GitHub App for easier authentication
- Webhook support for real-time monitoring
- Integration with GitHub Discussions

### Phase 3 (Long-term)
- Support for GitLab, Bitbucket
- Federated blocklist protocol
- Browser extension for maintainers
- Analytics and reporting features

---

## 16. Implementation Phases

### Phase 1: Core Functionality (MVP)
- CLI commands (scan, block, check, list, close-pr)
- Local SQLite database
- Basic PR quality heuristics
- Manual review workflow
- Export/import JSON
- Configuration file support

### Phase 2: Sharing & Automation
- Remote Turso database support
- Blocklist source subscriptions
- CSV export
- Batch PR operations
- GitHub Actions integration

### Phase 3: Advanced Features
- Account age analysis
- Pattern matching system
- Audit logging
- Statistics and reporting
- Enhanced configuration options

---

## Appendix A: Open Questions

1. **Default Thresholds:** What values are conservative enough to avoid false positives?
2. **Comment Templates:** What should automated PR closure comments say?
3. **Severity Levels:** How should different severity levels affect behavior?
4. **Appeal Process:** Should there be a formal process for unblocking?
5. **Maintainer Prompts:** What information helps maintainers decide on uncertain cases?
6. **Trademark Verification:** Confirm PRGuard name availability before public release

---

## Appendix B: Name Selection Rationale

**Selected Name:** PRGuard

**Reasoning:**
- Clear and descriptive of function
- Lowest trademark risk among candidates
- Professional tone appropriate for open source tooling
- Easy to remember and type
- Works well as both project name and CLI command (`prguard`)

**Rejected Names:**
- Bouncer - High trademark risk (security/club industry)
- SpamShield - Potential conflicts with spam filtering products
- GateKeeper - Possible trademark issues, negative connotations
- PRFilter - Too generic, less memorable

---

## Appendix C: Example Configuration File

```yaml
# config.example.yaml - PRGuard Configuration Template

github:
  token: "YOUR_GITHUB_TOKEN_HERE"
  org: "your-org-name"  # or use 'user' instead
  # user: "your-username"

database:
  type: "sqlite"  # or "turso"
  path: "./prguard.db"
  # For Turso:
  # type: "turso"
  # url: "libsql://your-db.turso.io"
  # auth_token: "${TURSO_AUTH_TOKEN}"

filters:
  min_files: 2
  min_lines: 10
  account_age_days: 7
  readme_only_block: true
  
  # Whitelist trusted contributors
  whitelist:
    - "dependabot[bot]"
    - "renovate[bot]"

blocklist:
  auto_export: true
  export_path: "./exports"
  
  # Subscribe to trusted blocklist sources
  sources:
    - name: "Community Maintainers"
      url: "https://example.com/blocklist.json"
      trusted: true
      auto_sync: false

actions:
  close_prs: true
  add_spam_label: true
  comment_template: |
    This PR has been automatically closed due to low quality indicators.
    If you believe this is an error, please contact the maintainers.
```

---

**End of Specification Document**
