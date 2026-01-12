# Binary name
BINARY_NAME=pdfdarkmode

# Output directory
BIN_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags for optimized binary
# -s: Omit symbol table and debug info
# -w: Omit DWARF symbol table
LDFLAGS=-s -w

# Version info (can be overridden)
VERSION?=1.0.1
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Version ldflags
VERSION_LDFLAGS=-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)

.PHONY: all build build-release build-small build-all release clean deps test help tag-release

# Default target
all: build

# Ensure bin directory exists
$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

# Standard build (with debug info)
build: $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME) .

# Release build (optimized, stripped)
build-release: $(BIN_DIR)
	CGO_ENABLED=0 $(GOBUILD) -ldflags="$(LDFLAGS) $(VERSION_LDFLAGS)" -trimpath -o $(BIN_DIR)/$(BINARY_NAME) .

# Smallest possible binary (aggressive optimization)
build-small: $(BIN_DIR)
	CGO_ENABLED=0 $(GOBUILD) -ldflags="$(LDFLAGS) $(VERSION_LDFLAGS)" -trimpath -o $(BIN_DIR)/$(BINARY_NAME) .
	@which upx > /dev/null && upx --best --lzma $(BIN_DIR)/$(BINARY_NAME) || echo "Install upx for further compression: brew install upx"

# Cross-compilation targets
build-linux: $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS) $(VERSION_LDFLAGS)" -trimpath -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 .

build-linux-arm: $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) -ldflags="$(LDFLAGS) $(VERSION_LDFLAGS)" -trimpath -o $(BIN_DIR)/$(BINARY_NAME)-linux-arm64 .

build-windows: $(BIN_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS) $(VERSION_LDFLAGS)" -trimpath -o $(BIN_DIR)/$(BINARY_NAME)-windows-amd64.exe .

build-mac-intel: $(BIN_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS) $(VERSION_LDFLAGS)" -trimpath -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 .

build-mac-arm: $(BIN_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags="$(LDFLAGS) $(VERSION_LDFLAGS)" -trimpath -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 .

# Build all platforms
build-all: build-linux build-linux-arm build-windows build-mac-intel build-mac-arm

# Create release packages
release: clean build-all
	@mkdir -p release
	@echo "Creating release packages..."
	@tar -czf release/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BIN_DIR) $(BINARY_NAME)-linux-amd64
	@tar -czf release/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz -C $(BIN_DIR) $(BINARY_NAME)-linux-arm64
	@cd $(BIN_DIR) && zip -q ../release/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@tar -czf release/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz -C $(BIN_DIR) $(BINARY_NAME)-darwin-amd64
	@tar -czf release/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz -C $(BIN_DIR) $(BINARY_NAME)-darwin-arm64
	@echo "Release packages created in release/ directory:"
	@ls -lh release/

# Create and push a release tag (triggers GitHub Actions release)
# Usage: make tag-release VERSION=1.0.0
tag-release:
	@if [ -z "$(VERSION)" ]; then echo "Usage: make tag-release VERSION=x.y.z"; exit 1; fi
	@echo "Creating release tag v$(VERSION)..."
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@echo "Pushing tag to origin..."
	git push origin v$(VERSION)
	@echo "Release v$(VERSION) triggered! Check GitHub Actions for progress."

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BIN_DIR)
	rm -rf release

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run tests
test:
	$(GOTEST) -v ./...

# Install to $GOPATH/bin
install: build-release
	cp $(BIN_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# Show binary size
size: build-release
	@ls -lh $(BIN_DIR)/$(BINARY_NAME)
	@file $(BIN_DIR)/$(BINARY_NAME)

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build with debug info (development)"
	@echo "  build-release  - Build optimized, stripped binary (recommended)"
	@echo "  build-small    - Build smallest binary (uses upx if available)"
	@echo "  build-all      - Build for all platforms"
	@echo "  release        - Build all platforms and create release packages"
	@echo "  tag-release    - Create and push a git tag to trigger GitHub release"
	@echo "                   Usage: make tag-release VERSION=1.0.0"
	@echo "  build-linux    - Build for Linux amd64"
	@echo "  build-linux-arm- Build for Linux arm64"
	@echo "  build-windows  - Build for Windows amd64"
	@echo "  build-mac-intel- Build for macOS Intel"
	@echo "  build-mac-arm  - Build for macOS Apple Silicon"
	@echo "  clean          - Remove build artifacts"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  test           - Run tests"
	@echo "  install        - Install to GOPATH/bin"
	@echo "  size           - Show binary size"
	@echo "  help           - Show this help"
	@echo ""
	@echo "Binaries are output to the $(BIN_DIR)/ directory"
