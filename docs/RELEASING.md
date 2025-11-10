# Release Process

This document describes how to create and publish new releases of PRGuard.

## Overview

PRGuard uses **semantic versioning** (MAJOR.MINOR.PATCH) and **manual releases** via the Makefile. Releases are NOT automated in CI to maintain full control over when versions are published.

**Version Source of Truth:** PRGuard uses **GitHub git tags** as the authoritative source for version numbers, following Go module conventions. The VERSION file is maintained for convenience but tags are queried from GitHub when determining the current version.

## Prerequisites

Before creating a release:

1. All tests must pass: `make test`
2. Code must be linted: `make lint`
3. Working directory must be clean (all changes committed)
4. You must be on the `main` branch

## Release Workflow

### 1. Check Current Status

```bash
# Show current version (compares GitHub tags, local tags, and VERSION file)
make show-version

# Run pre-release checks (tests + lint + clean working dir)
make release
```

The `make show-version` command displays:
- 📦 GitHub remote tags (source of truth)
- 📦 Local git tags
- 📦 VERSION file contents
- ✅ Current version being used for builds

The `make release` command will:
- ✅ Run all tests
- ✅ Run linters
- ✅ Verify working directory is clean
- ✅ Show current version
- ✅ Preview what the next version would be for each bump type

### 2. Bump Version

Choose the appropriate version bump based on your changes:

#### Patch Release (Bug fixes)
```bash
make bump-patch
# Example: 1.0.0 → 1.0.1
```

Use for:
- Bug fixes
- Documentation updates
- Minor tweaks
- Performance improvements

#### Minor Release (New features)
```bash
make bump-minor
# Example: 1.0.0 → 1.1.0
```

Use for:
- New features
- New commands
- Backward-compatible changes

#### Major Release (Breaking changes)
```bash
make bump-major
# Example: 1.0.0 → 2.0.0
```

Use for:
- Breaking API changes
- Major refactoring
- Incompatible changes

### 3. What Happens During Version Bump

When you run `make bump-patch/minor/major`, the Makefile will:

1. 📦 Show current version (from GitHub tags)
2. 🔍 Check if the new version tag already exists on GitHub
3. ❌ Exit with error if tag already exists (prevents duplicates)
4. ✨ Update the `VERSION` file
5. 📝 Create a git commit: `Release v<version>`
6. 🏷️  Create an annotated git tag: `v<version>`
7. 💡 Show push instructions

**Note:** The tag existence check prevents accidentally creating duplicate version tags, which is important since Go modules use git tags as the source of truth.

### 4. Push the Release

After the version is bumped, you have two options:

#### Option A: Manual Push (recommended for first-timers)
```bash
# Push the commit and tag separately
git push origin main
git push origin v1.2.3  # Replace with your version
```

#### Option B: Quick Push
```bash
# Automatically pushes the latest tag
make release-push
```

### 5. GitHub Actions Takes Over

Once you push the tag, GitHub Actions will:
- ✅ Run all tests
- ✅ Build binaries for multiple platforms
- ✅ Create a GitHub Release (if configured)
- ✅ Publish to Homebrew (when Homebrew tap is set up)

## Example: Full Release Flow

```bash
# 1. Make sure you're on main with latest changes
git checkout main
git pull origin main

# 2. Make your changes and commit them
# ... (edit code, add features, fix bugs) ...
git add .
git commit -m "Add new spam detection feature"

# 3. Run pre-release checks
make release

# Output shows:
# ✅ Working directory is clean
# 🧪 Tests passed
# ✨ Linting passed
# 📦 Current version: 0.1.0
# Ready to create a release!
# Run one of these commands:
#   make bump-patch  - for bug fixes (0.1.0 → 0.1.1)
#   make bump-minor  - for new features (0.1.0 → 0.2.0)
#   make bump-major  - for breaking changes (0.1.0 → 1.0.0)

# 4. Bump version (choosing minor for new feature)
make bump-minor

# Output shows:
# 📦 Current version: 0.1.0
# ✨ New version: 0.2.0
# 📝 Committing version 0.2.0
# ✅ Created tag v0.2.0
#
# To push the release, run:
#   git push origin main && git push origin v0.2.0
#
# Or run: make release-push

# 5. Push the release
make release-push

# Output shows:
# 🚀 Pushing v0.2.0 to GitHub...
# ✅ Release pushed! GitHub Actions will build and publish.
```

## Versioning Guidelines

### When to Bump PATCH (x.y.Z)
- Bug fixes that don't change the API
- Documentation improvements
- Code refactoring with no user-facing changes
- Performance improvements
- Dependency updates (minor/patch)

### When to Bump MINOR (x.Y.0)
- New commands added
- New flags added to existing commands
- New configuration options
- New features that are backward-compatible
- Deprecation warnings (but not removals)

### When to Bump MAJOR (X.0.0)
- Removing commands or flags
- Changing command behavior in incompatible ways
- Changing configuration file format
- Removing deprecated features
- Database schema changes requiring migration

## Troubleshooting

### "Tag already exists on GitHub"
```bash
# The version you're trying to create already exists on GitHub
# Check current status
make show-version

# This will show you GitHub tags vs local tags
# If GitHub has a newer version, fetch and sync:
git fetch --tags origin
git pull origin main

# Then try bumping again with the next version
make bump-patch  # or bump-minor/major as appropriate
```

**Common causes:**
- Someone else created a release
- You already pushed this version
- Local repo is out of sync with GitHub

### "Working directory is not clean"
```bash
# You have uncommitted changes
git status

# Either commit them or stash them
git add .
git commit -m "Your changes"
# OR
git stash
```

### "Tests failed"
```bash
# Fix the failing tests first
make test

# Then try again
make release
```

### "No tags found" when running release-push
```bash
# You need to create a version bump first
make bump-patch  # or bump-minor or bump-major
```

### Want to see what a version bump would do without committing?
```bash
# Run 'make release' to see previews of all bump types
make release
```

### Made a mistake with version bump?
```bash
# If you haven't pushed yet, you can undo
git reset --hard HEAD~1  # Remove the commit
git tag -d v1.2.3        # Remove the tag (replace with your version)

# Then try again
make bump-patch  # or bump-minor or bump-major
```

## GoReleaser Integration

PRGuard is configured with GoReleaser for automated binary builds. When you push a tag:

1. GitHub Actions detects the tag
2. Runs tests (must pass)
3. Runs GoReleaser to build binaries
4. Creates GitHub Release with artifacts
5. Publishes to Homebrew tap (when configured)

See `.goreleaser.yml` for configuration details.

## Future: Homebrew Distribution

When ready to publish to Homebrew:

1. Create a repository: `prguard/homebrew-tap`
2. Uncomment the `brews:` section in `.goreleaser.yml`
3. Follow GitHub's Homebrew tap setup guide
4. Next release will automatically update Homebrew formula

## Questions?

- Check the Makefile: `make help`
- Review CI workflows: `.github/workflows/ci.yml`
- GoReleaser config: `.goreleaser.yml`
- Open an issue if you need help
