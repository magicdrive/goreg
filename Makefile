# Makefile for goreg
#
# Works with GNU make 3.81 (macOS) and later. Releases are published only by
# the release workflow on a v* tag push (see docs/development/releasing.md):
# `make preflight tag=vX.Y.Z` checks that the release can run, and
# `make publish tag=vX.Y.Z` runs that check, then creates and pushes the tag.

# Variables
BUILD_DIR = $(CURDIR)/build
CMD_DIRS = $(wildcard cmd/*)
BINARY_NAME = goreg
VERSION = $(shell git describe --tags --always)
LDFLAGS += -X "main.version=$(VERSION)"

PLATFORMS := linux/amd64 darwin/amd64 windows/amd64
GO := GO111MODULE=on CGO_ENABLED=0 go

# Tool versions. CI runs these targets, so the versions here are the ones CI
# uses; Renovate updates them through the `# renovate:` comments.
# renovate: datasource=go depName=github.com/magicdrive/hephaestus
HEPHAESTUS_VERSION ?= v1.1.0
HEPHAESTUS ?= go run github.com/magicdrive/hephaestus@$(HEPHAESTUS_VERSION)
# renovate: datasource=go depName=honnef.co/go/tools
STATICCHECK_VERSION ?= v0.8.2
# renovate: datasource=go depName=golang.org/x/vuln
GOVULNCHECK_VERSION ?= v1.9.0
# renovate: datasource=go depName=github.com/rhysd/actionlint
ACTIONLINT_VERSION ?= v1.7.12
# renovate: datasource=npm depName=renovate
RENOVATE_VERSION ?= 44.149.3

# Tag of the local, never published build used by the Homebrew end-to-end
# test (it must sort after every real release, for `brew upgrade`).
E2E_TAG ?= v99.0.0

# Argments
tag =

# Default target
.PHONY: all
all: help

# Build all artifacts
.PHONY: build
build: clean
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) -ldflags "$(LDFLAGS)"
	@chmod 755 $(BUILD_DIR)/$(BINARY_NAME)

# Build artifacts for all platforms and release
.PHONY: release-build
release-build: clean $(PLATFORMS)
	@echo "Release files are created in the $(BUILD_DIR) directory."

# Build each platform
.PHONY: $(PLATFORMS)
$(PLATFORMS):
	@mkdir -p $(BUILD_DIR)
	GOOS=$(word 1,$(subst /, ,$@)) GOARCH=$(word 2,$(subst /, ,$@)) \
		 $(GO) build -o $(BUILD_DIR)/$(word 1,$(subst /, ,$@))-$(word 2,$(subst /, ,$@))/$(BINARY_NAME)\
		 -ldflags "$(LDFLAGS)" .
	chmod 755 $(BUILD_DIR)/$(word 1,$(subst /, ,$@))-$(word 2,$(subst /, ,$@))/$(BINARY_NAME)

# Run go test for each directory
.PHONY: test
test:
	@$(GO) test $(CURDIR)/...

# Run go test with the race detector (needs cgo)
.PHONY: test-race
test-race:
	@GO111MODULE=on CGO_ENABLED=1 go test -race -count=1 $(CURDIR)/...

.PHONY: test-bench
test-bench:
	@$(GO) test $(CURDIR)/... -bench=.

# Run go test with verbose output and clear test cache
.PHONY: test-verbose
test-verbose:
	@$(GO) clean -testcache
	@$(GO) test -v $(CURDIR)/...

# Format the Go sources
.PHONY: fmt
fmt:
	@gofmt -w .

# Fail when a Go source is not gofmt-formatted
.PHONY: fmt-check
fmt-check:
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then echo "gofmt needed (run make fmt):"; echo "$$unformatted"; exit 1; fi

# Run go vet for the host and for Windows
.PHONY: vet
vet:
	@$(GO) vet ./...
	@GOOS=windows $(GO) vet ./...

# Fail when go.mod or go.sum is not tidy
.PHONY: tidy-check
tidy-check:
	@$(GO) mod tidy -diff

# Static analysis
.PHONY: lint
lint:
	@$(GO) run honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION) ./...

# Known vulnerabilities in the dependencies and the Go standard library
.PHONY: vuln
vuln:
	@$(GO) run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

# Build, then run the binary: --version, --help and one formatting run
.PHONY: smoke
smoke: build
	@$(BUILD_DIR)/$(BINARY_NAME) --version | grep -q '^goreg version .' || { echo "smoke: unexpected --version output"; exit 1; }
	@$(BUILD_DIR)/$(BINARY_NAME) --help | grep -q '^Usage: goreg ' || { echo "smoke: unexpected --help output"; exit 1; }
	@$(BUILD_DIR)/$(BINARY_NAME) --local github.com/magicdrive/goreg misc/example.go.txt | grep -q '^package main' || { echo "smoke: formatting failed"; exit 1; }
	@echo "smoke: ok ($$($(BUILD_DIR)/$(BINARY_NAME) --version))"

# Every local quality gate that CI runs on each pull request
.PHONY: check
check: fmt-check vet tidy-check lint test test-race vuln smoke hephaestus-check

# Lint the GitHub Actions workflows
.PHONY: actionlint
actionlint:
	@$(GO) run github.com/rhysd/actionlint/cmd/actionlint@$(ACTIONLINT_VERSION)

# Validate renovate.json
.PHONY: renovate-check
renovate-check:
	@npx --yes --package renovate@$(RENOVATE_VERSION) -- renovate-config-validator --strict renovate.json

# Validate hephaestus.yaml
.PHONY: hephaestus-check
hephaestus-check:
	@$(HEPHAESTUS) check

# Build every release target, archive and checksum into ./dist without a tag
# (version 0.0.0-SNAPSHOT), then check the layout the releases promise
.PHONY: snapshot
snapshot:
	@rm -rf dist
	@$(HEPHAESTUS) run --snapshot --stages build,archive,checksum --fresh
	@scripts/verify-dist.sh dist 0.0.0-SNAPSHOT

# Local build of $(E2E_TAG) with its Homebrew formula in ./dist (nothing is
# published: no release, no tap stage, no token)
.PHONY: e2e-dist
e2e-dist:
	@rm -rf dist
	@env -u GITHUB_TOKEN -u GH_TOKEN -u HOMEBREW_TAP_GITHUB_TOKEN \
		$(HEPHAESTUS) run --tag $(E2E_TAG) --stages build,archive,checksum,formula --fresh
	@scripts/verify-dist.sh dist $(patsubst v%,%,$(E2E_TAG))

# Real Homebrew end-to-end test: the tap's current formula, then an upgrade
# to the formula generated here. Installs into the local Homebrew prefix and
# removes everything on exit; intended for CI runners and containers (it
# refuses to run when goreg is already installed).
.PHONY: e2e-homebrew
e2e-homebrew: e2e-dist
	@scripts/homebrew-e2e.sh dist $(E2E_TAG)

# Read-only release checks for the tag about to be pushed. Needs a GitHub
# token in GITHUB_TOKEN, e.g. GITHUB_TOKEN="$$(gh auth token)".
.PHONY: preflight
preflight:
	@if [ -z "$(tag)" ]; then \
		echo "Error: tag is not set."; \
		echo "ex) make preflight tag=v0.0.1"; \
		exit 1; \
	fi
	@$(HEPHAESTUS) preflight --tag $(tag) --strict

# Install application. Use `go install`
.PHONY: install
install:
	@echo "Installing goreg..."
	@$(GO) install -ldflags "$(LDFLAGS)"

# Clean build artifacts
.PHONY: clean
clean:
	@rm -rf $(BUILD_DIR)

# Execute goreg -w to entire gofile.
.PHONY: goreg
goreg:
	git ls-files | grep -e '.go$$' | xargs -I GOFILE goreg -w GOFILE

# Publish to github.com: run the release preflight, then create and push the
# tag. The release workflow builds and publishes the release from the tag.
.PHONY: publish
publish:
	@if [ -z "$(tag)" ]; then \
		echo "Error: version is not set. Please set it and try again."; \
		echo "ex) make publish tag=v0.0.1"; \
		echo ; \
		echo ; \
		exit 1; \
	fi
	@$(MAKE) --no-print-directory preflight tag=$(tag)
	git tag $(tag)
	git push origin $(tag)


# Show help
.PHONY: help
help:
	@echo "Makefile commands:"
	@echo "  make build             - Build all artifacts"
	@echo "  make release-build     - Build artifacts for multiple platforms with version info"
	@echo "  make install           - Install application. Use 'go install'"
	@echo "  make test              - Run go test"
	@echo "  make test-race         - Run go test with the race detector"
	@echo "  make test-bench        - Run go test ./... bench=."
	@echo "  make test-verbose      - Run go test -v with go clean -testcache"
	@echo "  make fmt               - Format the Go sources with gofmt"
	@echo "  make fmt-check         - Fail when a Go source is not gofmt-formatted"
	@echo "  make vet               - Run go vet (host and windows)"
	@echo "  make tidy-check        - Fail when go.mod/go.sum are not tidy"
	@echo "  make lint              - Run staticcheck"
	@echo "  make vuln              - Run govulncheck"
	@echo "  make smoke             - Build and run --version, --help and a formatting run"
	@echo "  make check             - Run every local quality gate"
	@echo "  make actionlint        - Lint the GitHub Actions workflows"
	@echo "  make renovate-check    - Validate renovate.json"
	@echo "  make hephaestus-check  - Validate hephaestus.yaml"
	@echo "  make snapshot          - Build, archive and checksum every release target into ./dist"
	@echo "  make e2e-homebrew      - Homebrew install/upgrade end-to-end test (CI runners)"
	@echo "  make preflight tag=<tag> - Read-only release checks before tagging"
	@echo "  make clean             - Remove build artifacts"
	@echo "  make goreg             - Execute goreg -w to entire gofile"
	@echo "  make publish tag=<tag> - Run the preflight, then push the tag (the release workflow publishes)"
	@echo "  make help              - Show this message"
