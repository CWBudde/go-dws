# Semantic Analysis: Multi-Pass Architecture

**Created**: Task 6.1.2.1
**Status**: Design Complete
**Implementation**: In Progress

## Overview

The semantic analyzer uses a **multi-pass architecture** to handle DWScript's complex type system, forward declarations, and circular dependencies. The analyzer executes four distinct passes in sequence, each building on the results of the previous pass.

### Why Multi-Pass?

The original single-pass approach had critical limitations:

1. **Forward declarations don't work reliably** - Class A can't reference Class B if B is declared later
2. **Complex type dependencies fail** - Recursive types and circular references cause errors
3. **No parallelization** - Everything must be done sequentially in source order
4. **Difficult to cache** - Can't cache intermediate results for incremental analysis
5. **"Fix-up" validations scattered** - Post-hoc validation methods band-aid architectural issues

DWScript's language features **require** multi-pass analysis:

```pascal
// Forward declaration of TNode references TNode itself
type
  TNode = class;  // Pass 1: Register name

type
  TNode = class   // Pass 2: Resolve type references
    Next: TNode;  // Recursive type reference
  end;

// Class referencing a type declared later
type
  TFoo = class
    Bar: TBaz;    // TBaz not declared yet!
  end;

type
  TBaz = class
    Foo: TFoo;    // Circular dependency
  end;
```

### Architecture Principles

1. **Separation of Concerns**: Each pass has a single, well-defined responsibility
2. **Progressive Refinement**: Each pass adds more information to the shared context
3. **Early Error Detection**: Errors are detected as early as possible in the pipeline
4. **No AST Modification**: Passes only annotate the AST, never modify its structure
5. **Shared State**: All passes communicate through PassContext

## The Four Passes

### Pass Dependency Graph

```
Source Code
    ↓
    ↓ Parse
    ↓
AST (Program)
    ↓
    ├─→ Pass 1: Declaration Collection
    │   └─→ Output: Type names, Function signatures (unresolved)
    │
    ├─→ Pass 2: Type Resolution
    │   └─→ Output: Resolved types, Inheritance hierarchies
    │
    ├─→ Pass 3: Semantic Validation
    │   └─→ Output: Type-checked AST, Annotated expressions
    │
    └─→ Pass 4: Contract Validation
        └─→ Output: Validated contracts, Final errors
            ↓
    Bytecode Compilation / Interpretation
```

### Pass 1: Declaration Collection

**File**: `internal/semantic/passes/declaration_pass.go`

**Purpose**: Register all type and function names without resolving their references.

**Inputs**:
- AST (Program)
- Empty PassContext

**Responsibilities**:
- Register class names (including forward declarations)
- Register interface names
- Register enum names and values
- Register record names
- Register type aliases
- Register function/procedure signatures (with unresolved type references)
- Register global variables and constants
- Mark forward-declared types as "incomplete"

**Does NOT**:
- Resolve type references (e.g., "Integer" → types.INTEGER_TYPE)
- Resolve parent classes or interfaces
- Analyze function bodies
- Type-check expressions

**Outputs**:
- `PassContext.TypeRegistry` populated with type names
- `PassContext.Symbols` populated with function/variable names
- Types marked as "incomplete" for forward declarations

**Example**:

```pascal
type TFoo = class; // Forward declaration

type TBar = class(TFoo)  // Parent is "TFoo" (string, not resolved)
  Field: Integer;        // Type is "Integer" (string, not resolved)
end;
```

After Pass 1:
```
TypeRegistry:
  "tfoo" → ClassType{Name: "TFoo", Incomplete: true}
  "tbar" → ClassType{Name: "TBar", Parent: "TFoo" (unresolved string)}

Symbols:
  (none - no functions/variables yet)
```

### Pass 2: Type Resolution

**File**: `internal/semantic/passes/type_resolution_pass.go`

**Purpose**: Resolve all type references into concrete type objects.

**Inputs**:
- AST (Program)
- PassContext with registered but incomplete types

**Responsibilities**:
- Resolve type references: "Integer" → `types.INTEGER_TYPE`
- Resolve class parent types (build inheritance chains)
- Resolve interface parent interfaces
- Resolve field types in classes and records
- Resolve parameter types in function signatures
- Resolve return types in function signatures
- Build complete type hierarchies
- Detect circular type dependencies
- Validate that forward-declared types have implementations
- Resolve array element types
- Resolve set element types
- Resolve function pointer types

**Does NOT**:
- Type-check expressions or statements
- Analyze function bodies
- Validate abstract method implementations

**Outputs**:
- `PassContext.TypeRegistry` with fully resolved types
- `PassContext.Symbols` with resolved type references
- Complete inheritance hierarchies
- Errors for undefined types, circular dependencies

**Example**:

After Pass 2 (continuing from Pass 1):
```
TypeRegistry:
  "tfoo" → ClassType{Name: "TFoo", Parent: *ClassType{TObject}, Incomplete: false}
  "tbar" → ClassType{
    Name: "TBar",
    Parent: *ClassType{TFoo},  // Resolved pointer to TFoo
    Fields: {
      "Field": FieldType{Type: types.INTEGER_TYPE}  // Resolved to built-in type
    }
  }
```

**Circular Dependency Detection**:

```pascal
type A = B;  // Error: B not declared
type B = A;  // Error: Circular type dependency A → B → A
```

### Pass 3: Semantic Validation

**File**: `internal/semantic/passes/validation_pass.go`

**Purpose**: Type-check all expressions and statements now that types are resolved.

**Inputs**:
- AST (Program)
- PassContext with fully resolved types

**Responsibilities**:
- Type-check all expressions (binary, unary, calls, indexing, member access)
- Validate variable declarations
- Validate assignments (type compatibility, const violations)
- Type-check function calls (argument types, count, named args)
- Validate return statements (type, presence in all code paths)
- Validate control flow (break/continue only in loops, etc.)
- Check abstract method implementations in concrete classes
- Validate interface method implementations
- Validate visibility rules (private, protected, public)
- Check constructor/destructor rules
- Validate operator overloads
- Check property getter/setter compatibility
- Validate exception handling
- Annotate AST nodes with resolved types (store in `SemanticInfo`)

**Does NOT**:
- Resolve type names (already done in Pass 2)
- Validate contracts (done in Pass 4)

**Outputs**:
- `PassContext.SemanticInfo` with type annotations
- Errors for type mismatches, undefined variables, invalid operations

**Example**:

```pascal
var x: Integer;
var y: String;
x := y; // ERROR: Cannot assign String to Integer
```

```pascal
class TFoo = class
  procedure Bar; virtual; abstract;
end;

class TBaz = class(TFoo)
end; // ERROR: TBaz must implement abstract method Bar
```

### Pass 4: Contract Validation

**File**: `internal/semantic/passes/contract_pass.go`

**Purpose**: Validate Design-by-Contract constructs (requires, ensures, invariant).

**Inputs**:
- AST (Program)
- PassContext with fully type-checked expressions

**Responsibilities**:
- Validate `require` (precondition) clauses
  - Check expressions are boolean
  - Validate only parameters/constants referenced
- Validate `ensure` (postcondition) clauses
  - Check expressions are boolean
  - Validate `old(expr)` expressions
  - Check no side effects in old()
- Validate `invariant` clauses
  - Check expressions are boolean
  - Validate only fields/constants referenced
  - Check maintained after constructors
- Validate `assert` statements
- Check contract inheritance rules
  - Preconditions can be weakened (OR)
  - Postconditions can be strengthened (AND)

**Does NOT**:
- Type-check general expressions (already done in Pass 3)

**Outputs**:
- Errors for invalid contracts
- Warnings for detectable contract violations

**Example**:

```pascal
function Divide(x, y: Integer): Integer;
require
  y <> 0  // Precondition: divisor must be non-zero
ensure
  Result * y = x  // Postcondition: result * divisor = dividend
begin
  Result := x div y;
end;
```

```pascal
class TStack = class
private
  FCount: Integer;
invariant
  FCount >= 0  // Invariant: count never negative
public
  procedure Push(Item: Integer);
end;
```

## Implementation Details

### PassContext

**File**: `internal/semantic/passes/pass_context.go`

The `PassContext` is the shared state container for all passes:

```go
type PassContext struct {
    // Core Registries
    Symbols          *semantic.SymbolTable
    TypeRegistry     *semantic.TypeRegistry
    GlobalOperators  *types.OperatorRegistry
    ConversionRegistry *types.ConversionRegistry
    SemanticInfo     *pkgast.SemanticInfo

    // Secondary Registries
    UnitSymbols      map[string]*semantic.SymbolTable
    Helpers          map[string][]*types.HelperType
    Subranges        map[string]*types.SubrangeType
    FunctionPointers map[string]*types.FunctionPointerType

    // Error Collection
    Errors           []string
    StructuredErrors []*semantic.SemanticError

    // Execution Context
    CurrentFunction  interface{}
    CurrentClass     *types.ClassType
    CurrentRecord    *types.RecordType
    CurrentProperty  string

    // Source Context
    SourceCode       string
    SourceFile       string

    // State Flags
    LoopDepth        int
    InExceptionHandler bool
    InFinallyBlock   bool
    InLoop           bool
    InLambda         bool
    InClassMethod    bool
    InPropertyExpr   bool
}
```

### Pass Interface

**File**: `internal/semantic/passes/pass.go`

```go
type Pass interface {
    Name() string
    Run(program *ast.Program, ctx *PassContext) error
}
```

### PassManager

Coordinates execution of passes:

```go
type PassManager struct {
    passes []Pass
}

func (pm *PassManager) RunAll(program *ast.Program, ctx *PassContext) error {
    for _, pass := range pm.passes {
        if err := pass.Run(program, ctx); err != nil {
            return err
        }
        if ctx.HasCriticalErrors() {
            break  // Stop on critical errors
        }
    }
    return nil
}
```

## Usage Example

```go
// Create passes
declarationPass := passes.NewDeclarationPass()
typeResolutionPass := passes.NewTypeResolutionPass()
validationPass := passes.NewValidationPass()
contractPass := passes.NewContractPass()

// Create manager
manager := passes.NewPassManager(
    declarationPass,
    typeResolutionPass,
    validationPass,
    contractPass,
)

// Create context
ctx := passes.NewPassContext()

// Run all passes
if err := manager.RunAll(program, ctx); err != nil {
    return err
}

// Check for semantic errors
if ctx.HasCriticalErrors() {
    return fmt.Errorf("semantic analysis failed with %d errors", ctx.CriticalErrorCount())
}
```

## Benefits

### 1. Forward Declarations Work

```pascal
type TFoo = class;  // Pass 1: Register
type TBar = class(TFoo) ... end;  // Pass 1: Register, Pass 2: Resolve parent
type TFoo = class ... end;  // Pass 1: Complete, Pass 2: Resolve
```

### 2. Circular Dependencies Detected

```pascal
type A = B;  // Pass 1: Register
type B = A;  // Pass 1: Register, Pass 2: ERROR - cycle detected
```

### 3. Clear Error Messages

Each pass provides context-specific errors:
- Pass 1: "Type 'Foo' declared multiple times"
- Pass 2: "Type 'Bar' not found in scope"
- Pass 3: "Cannot assign String to Integer"
- Pass 4: "Invariant expression must be boolean"

### 4. Incremental Analysis (Future)

Pass results can be cached:
- Cache Pass 1 results for unchanged files
- Cache Pass 2 for unchanged type definitions
- Invalidate only affected passes when files change

### 5. Parallelization (Future)

- Pass 1 can analyze multiple files in parallel
- Pass 2 can resolve independent type hierarchies in parallel
- Pass 3 can validate independent functions in parallel

## Migration from Single-Pass

The current `Analyzer.Analyze()` method will be refactored to use the pass manager:

**Before** (analyzer.go:269-299):
```go
func (a *Analyzer) Analyze(program *ast.Program) error {
    for _, stmt := range program.Statements {
        a.analyzeStatement(stmt)  // Does everything at once
    }

    // Post-hoc fix-ups
    a.validateMethodImplementations()
    a.validateFunctionImplementations()
    a.validateClassForwardDeclarations()

    return ...
}
```

**After** (analyzer.go - to be implemented in task 6.1.2.6):
```go
func (a *Analyzer) Analyze(program *ast.Program) error {
    // Create pass manager
    manager := passes.NewPassManager(
        passes.NewDeclarationPass(),
        passes.NewTypeResolutionPass(),
        passes.NewValidationPass(),
        passes.NewContractPass(),
    )

    // Create context from analyzer state
    ctx := a.createPassContext()

    // Run all passes
    if err := manager.RunAll(program, ctx); err != nil {
        return err
    }

    // Update analyzer state from context
    a.updateFromContext(ctx)

    return a.buildAnalysisError()
}
```

## Implementation Tasks

- [x] **6.1.2.1**: Design pass architecture ✓
  - [x] Create `internal/semantic/passes/` package
  - [x] Define `Pass` interface
  - [x] Create `PassContext` struct
  - [x] Design pass dependency graph
  - [x] Write architecture documentation

- [ ] **6.1.2.2**: Implement Pass 1 (Declaration Collection)
- [ ] **6.1.2.3**: Implement Pass 2 (Type Resolution)
- [ ] **6.1.2.4**: Implement Pass 3 (Semantic Validation)
- [ ] **6.1.2.5**: Implement Pass 4 (Contract Validation)
- [ ] **6.1.2.6**: Integrate passes into Analyzer
- [ ] **6.1.2.7**: Add pass result caching

## Performance Considerations

### Memory

- PassContext reuses existing registries (no duplication)
- AST is annotated, not copied
- Type resolution is done once, cached in TypeRegistry

### Speed

- Four passes are slower than one pass initially
- But enables:
  - Incremental analysis (only re-run affected passes)
  - Parallel processing (within passes)
  - Better error recovery (stop early on critical errors)

### Benchmarks (Target)

- Single-pass baseline: 100%
- Multi-pass (no caching): 110-120% (acceptable overhead)
- Multi-pass (with caching): 80-90% (faster on incremental changes)

## Testing Strategy

1. **Unit Tests**: Each pass tested independently with mock PassContext
2. **Integration Tests**: Full pipeline tested with complex DWScript programs
3. **Regression Tests**: Existing fixture tests (~2,100) must pass
4. **Performance Tests**: Benchmarks for each pass and full pipeline
5. **Error Tests**: Verify correct error messages from each pass

## References

- **PLAN.md**: Task 6.1.2 (lines 1602-1712)
- **analyzer.go**: Current single-pass implementation (lines 267-310)
- **DWScript Original**: Multi-pass compiler in `reference/dwscript-original/`

## Future Enhancements

1. **Parallel Pass Execution**: Analyze multiple files in Pass 1 concurrently
2. **Incremental Analysis**: Cache pass results, invalidate on file changes
3. **LSP Integration**: Use PassContext for go-to-definition, hover, etc.
4. **Optimization Passes**: Add Pass 5 for constant folding, dead code elimination
5. **LLVM Backend**: Add Pass 6 for LLVM IR generation
