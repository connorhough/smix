# Default binary name
BINARY_NAME := smix

# Version information
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse --short HEAD)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# LDFLAGS for version injection
LDFLAGS := -X 'github.com/connorhough/smix/internal/version.Version=$(VERSION)' \
           -X 'github.com/connorhough/smix/internal/version.GitCommit=$(COMMIT)' \
           -X 'github.com/connorhough/smix/internal/version.BuildDate=$(DATE)'

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOINSTALL := $(GOCMD) install

# Build target
build:
	$(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) .

# Install target
install:
	$(GOINSTALL) -ldflags="$(LDFLAGS)" .

# Clean target
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Test target
test:
	$(GOTEST) -v ./...

# Lint target (optional)
lint:
	golangci-lint run

# Cross-compilation targets
build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-darwin-arm64 .

build-linux-amd64:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-linux-amd64 .

.PHONY: build install clean test lint build-darwin-arm64 build-linux-amd64