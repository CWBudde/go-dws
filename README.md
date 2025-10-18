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

🚧 **Work in Progress** - This project is under active development.

**Current Capabilities**:

- Variables and expressions
- Control flow (if/else, while, for, repeat, case)
- User-defined functions and procedures
- Recursion support
- Built-in functions (PrintLn, Print)

See [PLAN.md](PLAN.md) for the complete implementation roadmap and current progress.

## Installation

**Note:** Not yet ready for installation. This section will be updated when the first working version is released.

```bash
# Future installation (not yet available)
go install github.com/cwbudde/go-dws/cmd/dwscript@latest
```

## Usage

The CLI tool is functional for running DWScript programs with variables, expressions, control flow, and functions.

```bash
# Build the CLI tool
go build -o bin/dwscript ./cmd/dwscript

# Run a DWScript file
./bin/dwscript run script.dws

# Evaluate inline code
./bin/dwscript run -e "PrintLn('Hello, World!');"

# Parse and display AST (for debugging)
./bin/dwscript parse script.dws

# Tokenize source code
./bin/dwscript lex script.dws

# Show version
./bin/dwscript version
```

### Quick Examples

**Hello World**:

```bash
./bin/dwscript run -e "PrintLn('Hello, World!');"
```

**Variables and Arithmetic**:

```bash
./bin/dwscript run -e "var x := 5; var y := 10; PrintLn('Sum: ', x + y);"
```

**Control Flow**:

```bash
./bin/dwscript run -e "for var i := 1 to 5 do PrintLn(i);"
```

**Functions**:

```bash
./bin/dwscript run -e "function Add(a, b: Integer): Integer; begin Result := a + b; end; begin PrintLn(Add(5, 3)); end."
```

### Example Programs

**Factorial Calculator** (`examples/factorial.dws`):

```pascal
function Factorial(n: Integer): Integer;
begin
    if n <= 1 then
        Result := 1
    else
        Result := n * Factorial(n - 1);
end;

begin
    PrintLn('Factorial(5) = ', Factorial(5));
    PrintLn('Factorial(10) = ', Factorial(10));
end.
```

Run it:

```bash
./bin/dwscript run examples/factorial.dws
```

**FizzBuzz** (`examples/fizzbuzz.dws`):

```pascal
begin
    for var i := 1 to 20 do
    begin
        if (i mod 15) = 0 then
            PrintLn('FizzBuzz')
        else if (i mod 3) = 0 then
            PrintLn('Fizz')
        else if (i mod 5) = 0 then
            PrintLn('Buzz')
        else
            PrintLn(i);
    end;
end.
```

More examples available in the `testdata/` directory.

## Project Structure

```text
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

The project follows a 10-stage incremental development plan covering ~511 tasks. For detailed progress and task breakdown, see [PLAN.md](PLAN.md).

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

```text
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
