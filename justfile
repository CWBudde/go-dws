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