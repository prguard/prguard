.PHONY: build clean install test lint help bump-patch bump-minor bump-major show-version release fetch-tags

# Build variables
BINARY_NAME=prguard
# Get current version from GitHub tags (source of truth for Go modules)
CURRENT_VERSION=$(shell git ls-remote --tags origin 2>/dev/null | grep -o 'v[0-9]*\.[0-9]*\.[0-9]*$$' | sort -V | tail -n1 | sed 's/^v//' | grep . || cat VERSION 2>/dev/null || echo "0.0.0")
VERSION?=$(CURRENT_VERSION)
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/prguard

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f *.db *.db-shm *.db-wal
	@rm -rf exports/

install: build ## Build and install to $GOPATH/bin
	@echo "Installing to $(GOPATH)/bin/$(BINARY_NAME)..."
	@go install $(LDFLAGS) ./cmd/prguard

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

lint: ## Run linters
	@echo "Running linters..."
	@go fmt ./...
	@go vet ./...

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

run: build ## Build and run with example config
	@./$(BINARY_NAME) --help

# Version Management
# Usage:
#   make bump-patch  - Increment patch version (1.0.0 → 1.0.1)
#   make bump-minor  - Increment minor version (1.0.0 → 1.1.0)
#   make bump-major  - Increment major version (1.0.0 → 2.0.0)
#   make release     - Create a release (runs tests, builds, tags)

fetch-tags: ## Fetch latest tags from GitHub
	@echo "🔄 Fetching latest tags from GitHub..."
	@git fetch --tags origin 2>/dev/null || true

show-version: ## Show current version
	@echo "🔄 Checking GitHub for latest version..."
	@LATEST_REMOTE=$$(git ls-remote --tags origin 2>/dev/null | grep -o 'v[0-9]*\.[0-9]*\.[0-9]*$$' | sort -V | tail -n1 | sed 's/^v//' || echo "none"); \
	LATEST_LOCAL=$$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "none"); \
	VERSION_FILE=$$(cat VERSION 2>/dev/null || echo "none"); \
	echo ""; \
	echo "📦 Version Status:"; \
	echo "  GitHub (remote):     $$LATEST_REMOTE"; \
	echo "  Local (git tags):    $$LATEST_LOCAL"; \
	echo "  VERSION file:        $$VERSION_FILE"; \
	echo ""; \
	echo "Current version (source of truth): $(CURRENT_VERSION)"

bump-patch: ## Bump patch version (1.0.0 → 1.0.1)
	@$(MAKE) bump-version BUMP_TYPE=patch

bump-minor: ## Bump minor version (1.0.0 → 1.1.0)
	@$(MAKE) bump-version BUMP_TYPE=minor

bump-major: ## Bump major version (1.0.0 → 2.0.0)
	@$(MAKE) bump-version BUMP_TYPE=major

# Internal target for version bumping (don't call directly)
bump-version:
	@if [ -z "$(BUMP_TYPE)" ]; then \
		echo "❌ Error: Don't call 'make bump-version' directly!"; \
		echo ""; \
		echo "Use one of these commands instead:"; \
		echo "  make bump-patch  - Increment patch version (1.0.0 → 1.0.1)"; \
		echo "  make bump-minor  - Increment minor version (1.0.0 → 1.1.0)"; \
		echo "  make bump-major  - Increment major version (1.0.0 → 2.0.0)"; \
		echo ""; \
		exit 1; \
	fi
	@echo "📦 Current version: $(CURRENT_VERSION)"
	@IFS='.' read -r major minor patch <<< "$(CURRENT_VERSION)"; \
	case "$(BUMP_TYPE)" in \
		patch) patch=$$((patch + 1)) ;; \
		minor) minor=$$((minor + 1)); patch=0 ;; \
		major) major=$$((major + 1)); minor=0; patch=0 ;; \
	esac; \
	NEW_VERSION="$$major.$$minor.$$patch"; \
	echo ""; \
	echo "🔍 Checking if v$$NEW_VERSION already exists on GitHub..."; \
	if git ls-remote --tags origin 2>/dev/null | grep -q "refs/tags/v$$NEW_VERSION$$"; then \
		echo "❌ Error: Tag v$$NEW_VERSION already exists on GitHub!"; \
		echo ""; \
		echo "Recent tags on GitHub:"; \
		git ls-remote --tags origin 2>/dev/null | grep -o 'v[0-9]*\.[0-9]*\.[0-9]*$$' | sort -V | tail -5; \
		echo ""; \
		echo "Run 'make show-version' to see detailed version status."; \
		exit 1; \
	fi; \
	echo "✅ Tag v$$NEW_VERSION is available"; \
	echo ""; \
	echo "$$NEW_VERSION" > VERSION; \
	echo "✨ New version: $$NEW_VERSION"; \
	$(MAKE) commit-version VERSION=$$NEW_VERSION

# Commit version changes and create tag
commit-version:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Error: VERSION is empty!"; \
		exit 1; \
	fi
	@echo "📝 Committing version $(VERSION)"
	@git add VERSION
	@git commit -m "Release v$(VERSION)"
	@git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@echo "✅ Created tag v$(VERSION)"
	@echo ""
	@echo "To push the release, run:"
	@echo "  git push origin main && git push origin v$(VERSION)"
	@echo ""
	@echo "Or run: make release-push"

release-push: ## Push the latest release tag
	@LATEST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo ""); \
	if [ -z "$$LATEST_TAG" ]; then \
		echo "❌ No tags found. Create a release first with 'make bump-patch/minor/major'"; \
		exit 1; \
	fi; \
	echo "🚀 Pushing $$LATEST_TAG to GitHub..."; \
	git push origin main && git push origin $$LATEST_TAG
	@echo "✅ Release pushed! GitHub Actions will build and publish."

release: test lint ## Full release workflow (test, lint, build, ready for tag)
	@echo "🔍 Running pre-release checks..."
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "❌ Working directory is not clean. Commit or stash changes first."; \
		exit 1; \
	fi
	@echo "✅ Working directory is clean"
	@echo "🧪 Tests passed"
	@echo "✨ Linting passed"
	@echo ""
	@echo "📦 Current version: $(CURRENT_VERSION)"
	@echo ""
	@IFS='.' read -r major minor patch <<< "$(CURRENT_VERSION)"; \
	patch_preview="$$major.$$minor.$$((patch + 1))"; \
	minor_preview="$$major.$$((minor + 1)).0"; \
	major_preview="$$((major + 1)).0.0"; \
	echo "Ready to create a release!"; \
	echo "Run one of these commands:"; \
	echo "  make bump-patch  - for bug fixes ($(CURRENT_VERSION) → $$patch_preview)"; \
	echo "  make bump-minor  - for new features ($(CURRENT_VERSION) → $$minor_preview)"; \
	echo "  make bump-major  - for breaking changes ($(CURRENT_VERSION) → $$major_preview)"
