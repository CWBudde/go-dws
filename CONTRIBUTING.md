# Contributing to go-dws

Thank you for your interest in contributing to go-dws! This document provides guidelines and instructions for contributing.

## Getting Started

1. **Read the documentation**
   - [PLAN.md](PLAN.md) - Detailed implementation roadmap
   - [goal.md](goal.md) - High-level project goals
   - [README.md](README.md) - Project overview

2. **Study the reference**
   - Review the original DWScript source in `reference/dwscript-original/`
   - Understand the language features and semantics

3. **Set up your development environment**
   ```bash
   git clone https://github.com/cwbudde/go-dws.git
   cd go-dws
   go mod download
   ```

## Development Workflow

### 1. Choose a Task

- Check the [PLAN.md](PLAN.md) for uncompleted tasks
- Look for issues labeled "good first issue" or "help wanted"
- Coordinate with maintainers to avoid duplicate work

### 2. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/bug-description
```

### 3. Make Your Changes

- Follow the project structure and package organization
- Write idiomatic Go code
- Add comprehensive tests for all new functionality
- Update documentation as needed

### 4. Test Your Changes

```bash
# Run tests
go test ./...

# Run tests with race detector
go test -race ./...

# Run tests with coverage
go test -coverprofile=coverage.txt ./...

# Check formatting
gofmt -s -w .

# Run linter
golangci-lint run

# Run vet
go vet ./...
```

### 5. Commit Your Changes

Use clear, descriptive commit messages:

```
Add lexer support for hex literals

- Implement hex number parsing ($FF format)
- Add tests for various hex patterns
- Update lexer documentation

Relates to PLAN.md Stage 1 task 1.19
```

### 6. Submit a Pull Request

- Push your branch to your fork
- Create a pull request against the `main` branch
- Fill out the PR template completely
- Reference any related issues

## Code Style Guidelines

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` and `goimports` for formatting
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Write clear, self-documenting code with appropriate comments

### Documentation

- Add GoDoc comments for all exported types, functions, and methods
- Use complete sentences in comments
- Provide usage examples where helpful

### Testing

- Write table-driven tests where appropriate
- Test both success and error cases
- Use descriptive test names: `TestLexer_HexLiterals`
- Aim for >85% code coverage

### Error Handling

- Always handle errors explicitly
- Provide context in error messages
- Use `fmt.Errorf` with `%w` for error wrapping

## Project Structure

```
go-dws/
├── lexer/          # Tokenization
├── parser/         # Parsing and AST building
├── ast/            # AST node definitions
├── types/          # Type system
├── interp/         # Interpreter/runtime
├── cmd/dwscript/   # CLI application
├── testdata/       # Test scripts
└── reference/      # Original DWScript (read-only)
```

## Testing Strategy

### Unit Tests

- Place tests in `*_test.go` files in the same package
- Test individual functions and methods
- Mock dependencies where appropriate

### Integration Tests

- Test complete workflows (lex → parse → execute)
- Use real DWScript programs from `testdata/`
- Compare outputs with expected results

### Reference Tests

- Port tests from DWScript's test suite when possible
- Ensure behavior matches original implementation

## Incremental Development

This project follows an incremental development approach:

1. Each stage builds on previous stages
2. All features are tested thoroughly before moving forward
3. Maintain backward compatibility within stages
4. Don't break existing tests when adding features

## Communication

- **Issues**: Report bugs, request features, ask questions
- **Pull Requests**: Propose code changes
- **Discussions**: General discussions about design and architecture

## Code Review Process

1. Automated checks must pass (CI, tests, linters)
2. At least one maintainer review required
3. Address review feedback promptly
4. Maintain a collaborative and respectful tone

## Release Process

Releases follow [Semantic Versioning](https://semver.org/):

- **0.x.y**: Pre-1.0 development releases
- **1.0.0**: First stable release (after Stage 6 completion)
- **Patch (x.y.Z)**: Bug fixes only
- **Minor (x.Y.0)**: New features, backward compatible
- **Major (X.0.0)**: Breaking changes

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (TBD - pending original DWScript license review).

## Questions?

Don't hesitate to ask! Open an issue or discussion if you need help.

## Recognition

Contributors will be acknowledged in:
- README.md contributors section
- Release notes
- Git commit history

Thank you for contributing to go-dws! 🎉
