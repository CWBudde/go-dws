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

See [PLAN.md](PLAN.md) for the complete implementation roadmap and current progress.

## 🚀 Try It Online - Web Playground

**Try DWScript right now in your browser!** No installation needed.

👉 **[Open the DWScript Playground](https://cwbudde.github.io/go-dws/)** 👈

The Web Playground features:
- **Monaco Editor** (VS Code's editor) with DWScript syntax highlighting
- **WebAssembly-powered execution** - Run DWScript code at native speeds in your browser
- **Interactive examples** - Fibonacci, Factorial, Classes, and more
- **Code sharing** - Share your code via URL
- **Auto-save** - Your code persists in localStorage
- **Light & Dark themes** - Choose your preferred editor theme

Perfect for learning DWScript, testing code snippets, or experimenting with the language!

**Quick Start:**
1. Visit the [playground](https://cwbudde.github.io/go-dws/)
2. Try one of the example programs from the dropdown
3. Click "Run" or press `Ctrl+Enter`
4. See the output in real-time!

For local development or running the playground offline, see [playground/README.md](playground/README.md).

## Installation

### Option 1: Use the Web Playground (Recommended for Quick Start)

No installation needed! Just visit **[https://cwbudde.github.io/go-dws/](https://cwbudde.github.io/go-dws/)** and start coding.

### Option 2: Install the CLI Tool

```bash
# Clone the repository
git clone https://github.com/cwbudde/go-dws.git
cd go-dws

# Build the CLI tool
go build -o bin/dwscript ./cmd/dwscript

# Run a DWScript program
./bin/dwscript run script.dws
```

### Option 3: Use as a Go Library

```bash
# Add to your Go project
go get github.com/cwbudde/go-dws/pkg/dwscript
```

Then import and use in your Go code:

```go
import "github.com/cwbudde/go-dws/pkg/dwscript"

engine, _ := dwscript.New()
result, _ := engine.Eval(`PrintLn('Hello from Go!');`)
fmt.Print(result.Output)
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

## Embedding in Go Applications

go-dws can be used as a library to embed the DWScript interpreter in your Go applications:

```go
package main

import (
    "fmt"
    "log"

    "github.com/cwbudde/go-dws/pkg/dwscript"
)

func main() {
    // Create a new engine
    engine, err := dwscript.New()
    if err != nil {
        log.Fatal(err)
    }

    // Evaluate DWScript code
    result, err := engine.Eval(`
        var x: Integer := 42;
        PrintLn('The answer is ' + IntToStr(x));
    `)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Print(result.Output) // "The answer is 42"
}
```

For more examples and API documentation, see the [pkg/dwscript](https://pkg.go.dev/github.com/cwbudde/go-dws/pkg/dwscript) package documentation.

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

**Object-Oriented Example** (`examples/oop.dws`):

```pascal
type
  TShape = class abstract
  protected
    FColor: Integer;
  public
    constructor Create(color: Integer);
    function GetColor: Integer; virtual;
    function GetArea: Float; virtual; abstract;
  end;

  TCircle = class(TShape)
  private
    FRadius: Float;
  public
    constructor Create(color: Integer; radius: Float);
    function GetArea: Float; override;
  end;

  TRectangle = class(TShape)
  private
    FWidth, FHeight: Float;
  public
    constructor Create(color: Integer; width, height: Float);
    function GetArea: Float; override;
  end;

implementation

constructor TShape.Create(color: Integer);
begin
  FColor := color;
end;

function TShape.GetColor: Integer;
begin
  Result := FColor;
end;

constructor TCircle.Create(color: Integer; radius: Float);
begin
  inherited Create(color);
  FRadius := radius;
end;

function TCircle.GetArea: Float;
begin
  Result := 3.14159 * FRadius * FRadius;
end;

constructor TRectangle.Create(color: Integer; width, height: Float);
begin
  inherited Create(color);
  FWidth := width;
  FHeight := height;
end;

function TRectangle.GetArea: Float;
begin
  Result := FWidth * FHeight;
end;

// Main program
var
  shapes: array[0..1] of TShape;
begin
  shapes[0] := TCircle.Create(255, 5.0);
  shapes[1] := TRectangle.Create(128, 10.0, 20.0);

  for var i := 0 to 1 do
  begin
    PrintLn('Shape color: ', shapes[i].GetColor);
    PrintLn('Shape area: ', shapes[i].GetArea);
  end;
end.
```

**Interface Example** (`examples/interfaces.dws`):

```pascal
type
  IDrawable = interface
    procedure Draw;
  end;

  IPrintable = interface
    function ToString: String;
  end;

  TDocument = class(IDrawable, IPrintable)
  private
    FTitle: String;
  public
    constructor Create(title: String);
    procedure Draw; virtual;
    function ToString: String; virtual;
  end;

implementation

constructor TDocument.Create(title: String);
begin
  FTitle := title;
end;

procedure TDocument.Draw;
begin
  PrintLn('Drawing: ', FTitle);
end;

function TDocument.ToString: String;
begin
  Result := 'Document: ' + FTitle;
end;

// Main program
var
  doc: TDocument;
  drawable: IDrawable;
  printable: IPrintable;
begin
  doc := TDocument.Create('My Document');

  // Use as object
  doc.Draw;

  // Cast to interfaces
  drawable := doc as IDrawable;
  drawable.Draw;

  printable := doc as IPrintable;
  PrintLn(printable.ToString);
end.
```

More examples available in the `testdata/` directory.

## Object-Oriented Programming Features

go-dws includes a complete implementation of DWScript's object-oriented programming capabilities:

### Classes

- **Declaration**: `type TClassName = class(TParent) ... end;`
- **Fields**: Private, protected, and public members
- **Methods**: Procedures and functions with `Self` reference
- **Constructors**: Standard `Create` method or custom constructors
- **Inheritance**: Single inheritance from parent classes
- **Visibility**: `public`, `protected`, and `private` access modifiers

### Advanced OOP Features

- **Virtual Methods**: Methods marked with `virtual` can be overridden
- **Method Override**: Use `override` keyword to replace parent implementation
- **Abstract Classes**: Classes marked `abstract` cannot be instantiated
- **Abstract Methods**: Virtual methods with no implementation (must be overridden)
- **Static Members**: Class-level fields (`class var`) and methods (`class function`/`class procedure`)
- **Polymorphism**: Dynamic method dispatch based on actual object type

### Interfaces

- **Declaration**: `type IInterfaceName = interface ... end;`
- **Inheritance**: Interfaces can inherit from other interfaces
- **Implementation**: Classes can implement multiple interfaces
- **Casting**: Safe casting between objects and interfaces using `as` operator
- **Type Checking**: Use `is` operator to check interface compatibility
- **Method Dispatch**: Polymorphic method calls through interface references

### External Integration

- **External Classes**: Declare classes implemented in Go runtime
- **External Interfaces**: Interface to external Go code
- **FFI Preparation**: Foundation for Foreign Function Interface support

For detailed documentation on OOP features, see:
- [Stage 7 Completion Summary](docs/stage7-complete.md)
- [Delphi-to-Go Mapping Guide](docs/delphi-to-go-mapping.md)
- [Interfaces Implementation Guide](docs/interfaces-guide.md)

## Project Structure

```text
go-dws/
├── internal/
│   ├── lexer/          # Lexical analyzer (tokenizer)
│   ├── parser/         # Parser and AST builder
│   ├── ast/            # Abstract Syntax Tree node definitions
│   ├── types/          # Type system implementation
│   ├── semantic/       # Semantic analyzer
│   └── interp/         # Interpreter/runtime engine
├── pkg/
│   ├── dwscript/       # Public embedding API
│   ├── platform/       # Platform abstraction (native/WASM)
│   └── wasm/           # WebAssembly bridge code
├── cmd/
│   ├── dwscript/       # CLI application
│   └── dwscript-wasm/  # WASM entry point
├── playground/         # Web playground (Monaco Editor + WASM)
├── build/wasm/         # WASM build scripts and output
├── docs/               # Documentation
│   ├── wasm/          # WASM-specific docs (API.md, BUILD.md, PLAYGROUND.md)
│   └── plans/         # Design documents
├── testdata/           # Test scripts and data
├── reference/          # DWScript original source (read-only reference)
├── PLAN.md             # Detailed implementation roadmap
└── goal.md             # High-level project goals and strategy
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

### Multi-Platform Support

go-dws runs on multiple platforms:

**Native (Go):**
```text
CLI Tool (cmd/dwscript)
    ↓
DWScript Engine (pkg/dwscript)
    ↓
Native Platform (pkg/platform/native)
    ↓
OS (filesystem, console, etc.)
```

**WebAssembly (Browser):**
```text
Web Playground (playground/)
    ↓
Monaco Editor → JavaScript API (pkg/wasm/api.go)
                    ↓
            DWScript WASM Module
                    ↓
            WASM Platform (pkg/platform/wasm)
                    ↓
            Browser APIs (console.log, etc.)
```

The platform abstraction layer (`pkg/platform/`) enables DWScript to run seamlessly in both native Go environments and in the browser via WebAssembly, with identical language behavior.

### Future Optimizations

Planned enhancements may include:

- Bytecode compilation for better performance
- JIT compilation (if feasible in Go)
- JavaScript transpilation backend
- Additional platform targets (mobile, embedded)

## WebAssembly & Browser Support

go-dws compiles to WebAssembly, enabling it to run in any modern web browser at near-native speeds.

### Web Playground Features

The [DWScript Playground](https://cwbudde.github.io/go-dws/) provides:

- **Monaco Editor**: Full VS Code-style editor experience
  - Syntax highlighting for DWScript
  - Auto-indentation and code formatting
  - Find/replace, multi-cursor editing
  - Error markers for compilation errors
  - Minimap and line numbers

- **WASM Execution**: Run DWScript code in your browser
  - No server required - everything runs client-side
  - 50-80% of native Go performance
  - Instant feedback on code execution
  - Full access to DWScript language features

- **Developer Tools**:
  - 7 built-in example programs
  - URL-based code sharing (base64-encoded)
  - localStorage auto-save and restore
  - Light and dark themes
  - Keyboard shortcuts (Ctrl+Enter to run)
  - Split-pane UI with resizable panels

### Building for WebAssembly

```bash
# Build WASM module
just wasm

# Build with optimization
just wasm-opt

# Run playground locally
cd playground
python3 -m http.server 8080
```

For detailed WASM build instructions, see [docs/wasm/BUILD.md](docs/wasm/BUILD.md).

### JavaScript API

Embed DWScript in your web applications:

```javascript
// Load WASM module
const go = new Go();
const result = await WebAssembly.instantiateStreaming(
    fetch('dwscript.wasm'),
    go.importObject
);
go.run(result.instance);

// Create DWScript instance
const dws = new DWScript();
await dws.init();

// Run code
const result = dws.eval(`
    var x: Integer := 42;
    PrintLn('The answer is ' + IntToStr(x));
`);

console.log(result.output); // "The answer is 42"
```

For complete API documentation, see [docs/wasm/API.md](docs/wasm/API.md).

## License

**To be determined** - Pending review of DWScript's license and agreement with original author.

This project is a port/reimplementation and will respect the original DWScript license.

## Credits

- **Original DWScript**: [Eric Grange](https://github.com/EricGrange) and contributors
- **go-dws Port**: Christian Budde and contributors

## References

### DWScript Resources
- [DWScript Original Repository](https://github.com/EricGrange/DWScript)
- [DWScript Website](https://www.delphitools.info/dwscript/)

### go-dws Documentation
- [Implementation Plan](PLAN.md)
- [Project Goals](goal.md)

### Web Playground & WebAssembly
- [🚀 Try the Playground](https://cwbudde.github.io/go-dws/)
- [Playground Documentation](docs/wasm/PLAYGROUND.md)
- [JavaScript API Reference](docs/wasm/API.md)
- [WASM Build Guide](docs/wasm/BUILD.md)
- [Playground Quick Start](playground/README.md)

### OOP Features
- [Stage 7 Completion Summary](docs/stage7-complete.md)
- [Delphi-to-Go Mapping Guide](docs/delphi-to-go-mapping.md)
- [Interfaces Implementation Guide](docs/interfaces-guide.md)

## Contact

- GitHub Issues: [Report bugs or request features](https://github.com/cwbudde/go-dws/issues)

---

**Status**: 🚧 In Development - Not yet ready for production use

✅ **Web Playground**: [Try it now!](https://cwbudde.github.io/go-dws/) - Fully functional with WebAssembly execution
