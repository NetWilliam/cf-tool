# Makefile for cf-tool
# Modern Go build system with cross-platform support

# Variables
BINARY_NAME=cf
BINARY_PATH=./bin/$(BINARY_NAME)
SOURCE_DIR=.
MAIN_FILE=cf.go
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Go build flags
GO=go
GOFLAGS=-v -mod=mod
GOBUILD=$(GO) build $(GOFLAGS) $(LDFLAGS)
GOTEST=$(GO) test -v -race
GOGET=$(GO) get
GOMOD=$(GO) mod

# Build directories
BUILD_DIR=./bin
DIST_DIR=./dist

# Platform-specific settings
UNAME_S:=$(shell uname -s)
UNAME_M:=$(shell uname -m)

# Default target
.PHONY: all
all: clean build

# Build for current platform
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) for $(UNAME_S)/$(UNAME_M)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BINARY_PATH) $(MAIN_FILE)
	@echo "Build complete: $(BINARY_PATH)"

# Install to GOPATH/bin
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(LDFLAGS) $(MAIN_FILE)
	@echo "Installed $(BINARY_NAME)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@$(GO) clean
	@echo "Clean complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Run go vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	$(GOMOD) verify

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download

# Verify dependencies
.PHONY: verify
verify:
	@echo "Verifying dependencies..."
	$(GOMOD) verify

# Update dependencies
.PHONY: update-deps
update-deps:
	@echo "Updating dependencies..."
	$(GOMOD) tidy
	$(GO) get -u -t ./...
	$(GOMOD) tidy

# Cross-platform build targets
.PHONY: build-all
build-all: build-linux build-darwin build-windows
	@echo "Cross-platform build complete"

.PHONY: build-linux
build-linux:
	@mkdir -p $(DIST_DIR)/linux
	@echo "Building for Linux amd64..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(DIST_DIR)/linux/$(BINARY_NAME)-amd64 $(MAIN_FILE)
	@echo "Building for Linux arm64..."
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(DIST_DIR)/linux/$(BINARY_NAME)-arm64 $(MAIN_FILE)

.PHONY: build-darwin
build-darwin:
	@mkdir -p $(DIST_DIR)/darwin
	@echo "Building for macOS amd64..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(DIST_DIR)/darwin/$(BINARY_NAME)-amd64 $(MAIN_FILE)
	@echo "Building for macOS arm64 (Apple Silicon)..."
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(DIST_DIR)/darwin/$(BINARY_NAME)-arm64 $(MAIN_FILE)

.PHONY: build-windows
build-windows:
	@mkdir -p $(DIST_DIR)/windows
	@echo "Building for Windows amd64..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(DIST_DIR)/windows/$(BINARY_NAME)-amd64.exe $(MAIN_FILE)
	@echo "Building for Windows arm64..."
	GOOS=windows GOARCH=arm64 $(GOBUILD) -o $(DIST_DIR)/windows/$(BINARY_NAME)-arm64.exe $(MAIN_FILE)

# Development helpers
.PHONY: dev
dev: fmt vet build
	@echo "Development build complete"

.PHONY: check
check: fmt vet test
	@echo "All checks passed"

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all           - Clean and build (default)"
	@echo "  build         - Build for current platform"
	@echo "  install       - Install to GOPATH/bin"
	@echo "  clean         - Remove build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  tidy          - Tidy dependencies"
	@echo "  deps          - Download dependencies"
	@echo "  verify        - Verify dependencies"
	@echo "  update-deps   - Update all dependencies"
	@echo "  build-all     - Build for all platforms"
	@echo "  build-linux   - Build for Linux"
	@echo "  build-darwin  - Build for macOS"
	@echo "  build-windows - Build for Windows"
	@echo "  dev           - Format, vet, and build"
	@echo "  check         - Run all checks (fmt, vet, test)"
	@echo "  help          - Show this help message"
