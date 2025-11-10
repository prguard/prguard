# Release Quick Reference

> **TL;DR:** Use `make` commands to create releases. No scripts needed!

## Quick Commands

```bash
# Show current version
make show-version

# Check if ready to release (runs tests, lint, checks git status)
make release

# Create a new release (choose one)
make bump-patch    # 0.1.0 → 0.1.1 (bug fixes)
make bump-minor    # 0.1.0 → 0.2.0 (new features)
make bump-major    # 0.1.0 → 1.0.0 (breaking changes)

# Push the release to GitHub
make release-push
```

## How It Works

1. **GitHub git tags** - Source of truth for current version (Go module convention)
2. **Makefile** - Queries GitHub tags, does version math, prevents duplicate tags
3. **VERSION file** - Convenience file, kept in sync with tags

## Complete Flow

```bash
# 1. Check everything is ready
make release
# Output shows: current version + previews of bump options

# 2. Bump version (example: patch release)
make bump-patch
# This:
#   - Checks GitHub for existing v0.1.1 tag
#   - Updates VERSION file
#   - Commits: "Release v0.1.1"
#   - Tags: v0.1.1

# 3. Push to GitHub
make release-push
# GitHub Actions builds and publishes automatically
```

## Comparison to Node.js

**Node.js (semantic-docs):**
```bash
npm version patch      # Uses package.json
```

**Go (prguard):**
```bash
make bump-patch        # Uses VERSION file
```

Both do the same thing - just different ecosystems!

## See Also

- Full documentation: [RELEASING.md](../RELEASING.md)
- Makefile targets: `make help`
- CI config: `.github/workflows/ci.yml`
