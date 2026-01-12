# Binary name
BINARY_NAME=pdfdarkmode

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
VERSION?=1.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Version ldflags
VERSION_LDFLAGS=-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)

.PHONY: all build build-release build-small clean deps test help

# Default target
all: build

# Standard build (with debug info)
build:
	$(GOBUILD) -o $(BINARY_NAME) .

# Release build (optimized, stripped)
build-release:
	CGO_ENABLED=0 $(GOBUILD) -ldflags="$(LDFLAGS) $(VERSION_LDFLAGS)" -trimpath -o $(BINARY_NAME) .

# Smallest possible binary (aggressive optimization)
build-small:
	CGO_ENABLED=0 $(GOBUILD) -ldflags="$(LDFLAGS) $(VERSION_LDFLAGS)" -trimpath -o $(BINARY_NAME) .
	@which upx > /dev/null && upx --best --lzma $(BINARY_NAME) || echo "Install upx for further compression: brew install upx"

# Cross-compilation targets
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -trimpath -o $(BINARY_NAME)-linux-amd64 .

build-linux-arm:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) -ldflags="$(LDFLAGS)" -trimpath -o $(BINARY_NAME)-linux-arm64 .

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -trimpath -o $(BINARY_NAME)-windows-amd64.exe .

build-mac-intel:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -trimpath -o $(BINARY_NAME)-darwin-amd64 .

build-mac-arm:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags="$(LDFLAGS)" -trimpath -o $(BINARY_NAME)-darwin-arm64 .

# Build all platforms
build-all: build-linux build-linux-arm build-windows build-mac-intel build-mac-arm

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run tests
test:
	$(GOTEST) -v ./...

# Install to $GOPATH/bin
install: build-release
	cp $(BINARY_NAME) $(GOPATH)/bin/

# Show binary size
size: build-release
	@ls -lh $(BINARY_NAME)
	@file $(BINARY_NAME)

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build with debug info (development)"
	@echo "  build-release  - Build optimized, stripped binary (recommended)"
	@echo "  build-small    - Build smallest binary (uses upx if available)"
	@echo "  build-all      - Build for all platforms"
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
