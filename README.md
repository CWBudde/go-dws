# go-dws

A port of [DWScript](https://github.com/EricGrange/DWScript) (Delphi Web Script) from Delphi to Go.

## Overview

go-dws is a faithful implementation of the DWScript scripting language in Go, preserving 100% of DWScript's syntax and semantics while leveraging Go's modern language features and ecosystem.

**DWScript** is a full-featured Object Pascal-based scripting language featuring:
- Strong static typing with type inference
- Object-oriented programming (classes, interfaces, inheritance)
- Functions and procedures with nested scopes
- Operator overloading
- Exception handling
- Comprehensive built-in functions
- And much more...

## Project Status

🚧 **Early Development** - Stage 3 in progress

This project is being developed incrementally following a detailed implementation plan. See [PLAN.md](PLAN.md) for the complete roadmap.

### Recent Milestones

- ✅ **Stage 1 – Lexer complete** (token definitions, full lexer, 97% coverage, CLI `lex` command)
- ✅ **Stage 2 – Parser & AST complete** (Pratt parser for expressions, AST nodes, CLI `parse` dump)

### Current Milestone: Stage 3 - Parse and Execute Simple Statements

- [ ] Extend AST with statement nodes (variable declarations, assignments, blocks)
- [ ] Implement statement parsing and sequencing rules
- [ ] Introduce interpreter scaffolding for executing parsed statements
- [ ] Enhance CLI `run` command to evaluate scripts end-to-end
- [ ] Expand regression tests with DWScript samples covering statements

## Installation

**Note:** Not yet ready for installation. This section will be updated when the first working version is released.

```bash
# Future installation (not yet available)
go install github.com/cwbudde/go-dws/cmd/dwscript@latest
```

## Usage

**Note:** Not yet functional. This section shows the planned CLI interface.

```bash
# Run a DWScript file
dwscript run script.dws

# Evaluate an expression
dwscript -e "PrintLn('Hello, World!');"

# Show version
dwscript version
```

## Project Structure

```
go-dws/
├── lexer/          # Lexical analyzer (tokenizer)
├── parser/         # Parser and AST builder
├── ast/            # Abstract Syntax Tree node definitions
├── types/          # Type system implementation
├── interp/         # Interpreter/runtime engine
├── cmd/dwscript/   # CLI application
├── testdata/       # Test scripts and data
├── reference/      # DWScript original source (read-only reference)
├── PLAN.md         # Detailed implementation roadmap
└── goal.md         # High-level project goals and strategy
```

## Development Roadmap

The project follows a 10-stage incremental development plan:

1. ✅ **Stage 1**: Implement the Lexer (Tokenization)
2. ✅ **Stage 2**: Build a Minimal Parser and AST (Expressions Only)
3. 🔄 **Stage 3**: Parse and Execute Simple Statements ⬅️ *Current*
4. **Stage 4**: Control Flow (Conditions and Loops)
5. **Stage 5**: Functions, Procedures, and Scope Management
6. **Stage 6**: Static Type Checking and Semantic Analysis
7. **Stage 7**: Object-Oriented Features (Classes, Interfaces, Methods)
8. **Stage 8**: Additional DWScript Features and Polishing
9. **Stage 9**: Long-Term Evolution

See [PLAN.md](PLAN.md) for the complete task breakdown (~511 tasks).

### Estimated Timeline

- **Core compiler** (Stages 0-5): 3-6 months
- **Full feature parity** (All stages): 1-3 years
- Timeline depends on development pace and team size

## Design Philosophy

1. **100% Language Compatibility**: Preserve all DWScript syntax and semantics
2. **Incremental Development**: Each stage produces testable, working components
3. **Idiomatic Go**: Use Go best practices while honoring the original design
4. **Comprehensive Testing**: Mirror DWScript's extensive test suite
5. **Clear Documentation**: Maintain thorough docs for users and contributors

## Contributing

Contributions are welcome! This project is in very early stages.

### Getting Started

1. Read [PLAN.md](PLAN.md) to understand the implementation roadmap
2. Review the [reference/](reference/) directory for DWScript original source
3. Check open issues for tasks marked "good first issue"
4. Follow Go best practices and write tests for all changes

### Development Setup

```bash
# Clone the repository
git clone https://github.com/cwbudde/go-dws.git
cd go-dws

# Install dependencies
go mod download

# Run tests (when available)
go test ./...

# Build CLI (when available)
go build ./cmd/dwscript
```

## Architecture

The compiler/interpreter follows a traditional architecture:

```
Source Code → Lexer → Parser → AST → Semantic Analyzer → Interpreter
                                                            ↓
                                                         Output
```

Future optimizations may include:
- Bytecode compilation for better performance
- JIT compilation (if feasible in Go)
- JavaScript transpilation backend

## License

**To be determined** - Pending review of DWScript's license and agreement with original author.

This project is a port/reimplementation and will respect the original DWScript license.

## Credits

- **Original DWScript**: [Eric Grange](https://github.com/EricGrange) and contributors
- **go-dws Port**: Christian Budde and contributors

## References

- [DWScript Original Repository](https://github.com/EricGrange/DWScript)
- [DWScript Website](https://www.delphitools.info/dwscript/)
- [Implementation Plan](PLAN.md)
- [Project Goals](goal.md)

## Contact

- GitHub Issues: [Report bugs or request features](https://github.com/cwbudde/go-dws/issues)
- Organization: [MeKo-Tech](https://github.com/MeKo-Tech)

---

**Status**: 🚧 In Development - Not yet ready for production use
