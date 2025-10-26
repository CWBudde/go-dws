# justfile for go-dws project
# Run with: just <command>
# List all commands: just --list

# Default recipe - show available commands
default:
    @just --list

# Build the dwscript CLI binary
build:
    go build -o bin/dwscript ./cmd/dwscript

# Format all Go code
fmt:
    go fmt ./...
    goimports -w .

# Run linter (golangci-lint)
lint:
    golangci-lint run

# Fix linter issues automatically where possible
lint-fix:
    golangci-lint run --fix

# Run all tests
test:
    go test ./...

# Run tests with coverage
test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Run tests with verbose output
test-verbose:
    go test -v ./...

# Clean build artifacts
clean:
    rm -rf bin/
    rm -f coverage.out coverage.html

# Tidy dependencies
tidy:
    go mod tidy

# Install development tools
install-tools:
    go install golang.org/x/tools/cmd/goimports@latest
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Run the CLI with help
help:
    ./bin/dwscript --help

# Full development setup
setup: tidy install-tools

# Quick development cycle: format, lint, test, build
dev: fmt lint test build

# CI pipeline: format check, lint, test with coverage
ci: lint test-coverage

# Run lexer on a test file
lex file="testdata/hello.dws":
    ./bin/dwscript lex {{file}}

# Run parser on a test file  
parse file="testdata/hello.dws":
    ./bin/dwscript parse {{file}}

# Run interpreter on a test file (Stage 3+)
run file="testdata/hello.dws":
    ./bin/dwscript run {{file}}

# === WebAssembly Build Targets (Stage 10.15) ===

# Build WASM binary (modes: monolithic, modular, hybrid)
wasm mode="monolithic":
    @echo "Building WASM ({{mode}} mode)..."
    @./build/wasm/build.sh {{mode}}

# Build WASM binary with optimization
wasm-opt mode="monolithic":
    @echo "Building optimized WASM ({{mode}} mode)..."
    @OPTIMIZE=true ./build/wasm/build.sh {{mode}}

# Test WASM build (compile only, no execution)
wasm-test:
    @echo "Testing WASM build..."
    @GOOS=js GOARCH=wasm go build -o /tmp/dwscript-test.wasm ./cmd/dwscript-wasm
    @echo "✓ WASM compiles successfully"
    @rm /tmp/dwscript-test.wasm

# Optimize existing WASM binary
wasm-optimize file="build/wasm/dist/dwscript.wasm":
    @echo "Optimizing WASM binary..."
    @./build/wasm/optimize.sh {{file}}

# Clean WASM build artifacts
wasm-clean:
    @echo "Cleaning WASM build artifacts..."
    @rm -rf build/wasm/dist
    @echo "✓ Clean complete"

# Show WASM binary size
wasm-size file="build/wasm/dist/dwscript.wasm":
    #!/usr/bin/env bash
    set -e
    if [ -f {{file}} ]; then
        SIZE=$(stat -f%z {{file}} 2>/dev/null || stat -c%s {{file}})
        SIZE_MB=$(echo "scale=2; $SIZE / 1048576" | bc)
        echo "WASM binary size: $SIZE_MB MB ($SIZE bytes)"
    else
        echo "Error: WASM binary not found at {{file}}"
        echo "Run 'just wasm' first"
        exit 1
    fi

# Full WASM build pipeline (build, optimize, show size)
wasm-all mode="monolithic": (wasm mode) (wasm-optimize "build/wasm/dist/dwscript.wasm") (wasm-size "build/wasm/dist/dwscript.wasm")