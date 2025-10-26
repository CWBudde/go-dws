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

- ✅ **Variables and Expressions**: Full support for all DWScript types
- ✅ **Control Flow**: if/else, while, for, repeat, case statements
- ✅ **Functions**: User-defined functions and procedures with recursion
- ✅ **Object-Oriented Programming**: Classes, inheritance, polymorphism, interfaces (Stage 7 complete)
  - Classes with fields, methods, constructors
  - Single inheritance with method overriding
  - Virtual/abstract methods and abstract classes
  - Static fields and methods
  - Visibility control (public/protected/private)
  - Interfaces with multiple implementation
  - Interface casting and polymorphism
- ✅ **Type System**: Strong static typing with semantic analysis
- ✅ **Built-in Functions**: PrintLn, Print, Length, and more

**Completed Stages:**
- Stage 1: Lexer ✅
- Stage 2: Parser (expressions) ✅
- Stage 3: Statements and execution ✅
- Stage 4: Control flow ✅
- Stage 5: Functions and scope ✅
- Stage 6: Type checking ✅
- Stage 7: Classes and OOP ✅ (76.3% - all features complete)

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
