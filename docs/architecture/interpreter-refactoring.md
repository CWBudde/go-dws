# Interpreter Architecture Refactoring Design

**Version**: 1.0
**Status**: Draft
**Author**: Claude Code
**Date**: 2025-11-16
**Related**: Phase 3 of PLAN.md

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current Architecture Analysis](#current-architecture-analysis)
3. [Pain Points and Problems](#pain-points-and-problems)
4. [Proposed Architecture](#proposed-architecture)
5. [Migration Strategy](#migration-strategy)
6. [New Patterns and Conventions](#new-patterns-and-conventions)
7. [Success Metrics](#success-metrics)
8. [Timeline and Phases](#timeline-and-phases)
9. [Risks and Mitigations](#risks-and-mitigations)

---

## Executive Summary

The DWScript interpreter has grown to **149 files** and **~81,500 lines** of code with significant architectural debt. This document proposes a comprehensive refactoring to improve maintainability, testability, and extensibility while maintaining 100% backward compatibility and zero performance regression.

**Key Goals**:
- Decompose the God object (Interpreter with 27 fields)
- Replace the 228-line Eval() switch with visitor pattern
- Establish clear component boundaries and separation of concerns
- Reduce tight coupling and heavy type assertion usage
- Enable unit testing of individual components
- Foundation for future optimizations

**Non-Goals**:
- Changing interpreter semantics or behavior
- Breaking existing tests (all ~2,100 tests must pass)
- Changing public API surface
- Performance regression

---

## Current Architecture Analysis

### Package Structure

```
internal/interp/
├── interpreter.go (461 lines, 7 methods)
├── value.go (1,248 lines, 30+ value types)
├── environment.go (environment/scope management)
├── builtins_*.go (20+ files, 1,182-680 lines each)
├── expressions_*.go (expression evaluation)
├── statements_*.go (statement execution)
├── objects_*.go (OOP support)
├── functions_*.go (function calls and built-ins)
├── array.go (768 lines)
├── class.go (532 lines)
├── record.go (509 lines)
├── enum.go
├── exceptions.go
├── errors.go
└── *_test.go (98 test files)
```

**Statistics**:
- **Total files**: 149 (51 implementation, 98 tests)
- **Total lines**: ~81,500 (30,602 implementation, rest tests)
- **Test coverage**: 65.8% (target: 90%+)
- **Methods on Interpreter**: Scattered across 56 files
- **Value types**: 30+ types spread across 9 files

### Interpreter Struct (God Object)

The `Interpreter` struct has **27 fields** mixing multiple concerns:

```go
type Interpreter struct {
    // Execution State (8 fields)
    currentNode          ast.Node
    env                  *Environment
    callStack            errors.StackTrace
    oldValuesStack       []map[string]Value
    breakSignal          bool
    continueSignal       bool
    exitSignal           bool
    propContext          *PropertyEvalContext

    // Type System (9 fields)
    classes              map[string]*ClassInfo
    records              map[string]*RecordTypeValue
    interfaces           map[string]*InterfaceInfo
    classTypeIDRegistry  map[string]int
    recordTypeIDRegistry map[string]int
    enumTypeIDRegistry   map[string]int
    nextClassTypeID      int
    nextRecordTypeID     int
    nextEnumTypeID       int

    // Runtime Services (5 fields)
    output               io.Writer
    rand                 *rand.Rand
    randSeed             int64
    externalFunctions    *ExternalFunctionRegistry
    exception            *ExceptionValue
    handlerException     *ExceptionValue

    // Function/Operator Registries (4 fields)
    functions            map[string][]*ast.FunctionDecl
    globalOperators      *runtimeOperatorRegistry
    conversions          *runtimeConversionRegistry
    helpers              map[string][]*HelperInfo

    // Unit System (3 fields)
    unitRegistry         *units.UnitRegistry
    initializedUnits     map[string]bool
    loadedUnits          []string

    // Configuration & Metadata (3 fields)
    maxRecursionDepth    int
    semanticInfo         *pkgast.SemanticInfo
    sourceCode           string
    sourceFile           string
}
```

### Eval() Method (Giant Switch)

The central `Eval()` method is a **228-line switch statement** with **40+ cases**:

```go
func (i *Interpreter) Eval(node ast.Node) Value {
    switch node := node.(type) {
    case *ast.Program:              // Statements (15 cases)
    case *ast.ExpressionStatement:
    case *ast.VarDeclStatement:
    case *ast.ConstDecl:
    case *ast.AssignmentStatement:
    case *ast.BlockStatement:
    case *ast.IfStatement:
    case *ast.WhileStatement:
    case *ast.RepeatStatement:
    case *ast.ForStatement:
    case *ast.ForInStatement:
    case *ast.CaseStatement:
    case *ast.TryStatement:
    case *ast.RaiseStatement:
    case *ast.BreakStatement:
    case *ast.ContinueStatement:
    case *ast.ExitStatement:
    case *ast.ReturnStatement:
    case *ast.UsesClause:

    case *ast.FunctionDecl:         // Declarations (8 cases)
    case *ast.ClassDecl:
    case *ast.InterfaceDecl:
    case *ast.OperatorDecl:
    case *ast.EnumDecl:
    case *ast.RecordDecl:
    case *ast.HelperDecl:
    case *ast.ArrayDecl:
    case *ast.TypeDeclaration:

    case *ast.IntegerLiteral:       // Expressions (20+ cases)
    case *ast.FloatLiteral:
    case *ast.StringLiteral:
    case *ast.BooleanLiteral:
    case *ast.CharLiteral:
    case *ast.NilLiteral:
    case *ast.Identifier:
    case *ast.BinaryExpression:
    case *ast.UnaryExpression:
    case *ast.AddressOfExpression:
    case *ast.GroupedExpression:
    case *ast.CallExpression:
    case *ast.NewExpression:
    case *ast.MemberAccessExpression:
    case *ast.MethodCallExpression:
    case *ast.InheritedExpression:
    case *ast.SelfExpression:
    case *ast.EnumLiteral:
    case *ast.RecordLiteralExpression:
    case *ast.SetLiteral:
    case *ast.ArrayLiteralExpression:
    case *ast.IndexExpression:
    case *ast.NewArrayExpression:
    case *ast.LambdaExpression:
    case *ast.IsExpression:
    case *ast.AsExpression:
    case *ast.ImplementsExpression:
    case *ast.IfExpression:
    case *ast.OldExpression:

    default:
        return newError("unknown node type: %T", node)
    }
}
```

Most cases delegate to `evalXxx()` methods, but the dispatch logic is centralized and monolithic.

### Method Distribution

Interpreter methods are scattered across **56 files**. Top files by method count:

| File | Methods | Lines |
|------|---------|-------|
| builtins_strings_basic.go | 34 | 1,182 |
| builtins_math_basic.go | 28 | 680 |
| builtins_strings_advanced.go | 21 | 839 |
| builtins_datetime_info.go | 21 | - |
| builtins_datetime_format.go | 18 | - |
| builtins_math_trig.go | 17 | - |
| array.go | 16 | 768 |
| builtins_datetime_calc.go | 15 | - |
| functions_typecast.go | 14 | 608 |
| declarations.go | 13 | 932 |

This makes it difficult to understand component boundaries and responsibilities.

---

## Pain Points and Problems

### 1. God Object Anti-Pattern

**Problem**: The `Interpreter` struct violates Single Responsibility Principle by managing:
- Execution state (env, callStack, control flow signals)
- Type registries (classes, records, interfaces, type IDs)
- Runtime services (output, random, exceptions)
- Configuration (recursion limits)
- Unit loading system

**Impact**:
- Difficult to test components in isolation
- Changes in one area affect unrelated areas
- High coupling between all interpreter subsystems
- Cannot parallelize execution (shared mutable state)

**Example**: To add a new built-in function, you need to understand the entire Interpreter struct even though you only need `output` and `env`.

### 2. Giant Switch Statement

**Problem**: The 228-line `Eval()` switch is difficult to maintain and extend.

**Impact**:
- Adding new AST nodes requires modifying central dispatch
- No compile-time safety for missing cases
- Performance overhead from repeated type switching
- Difficult to add instrumentation or optimization

**Example**: Adding a new expression type requires touching the core Eval() method and risking regressions in unrelated code.

### 3. Poor Separation of Concerns

**Problem**: All interpreter code is in one flat package (internal/interp) with 149 files.

**Impact**:
- No clear module boundaries
- Difficult to understand dependencies
- Easy to create circular dependencies
- Hard to navigate codebase
- Cannot evolve subsystems independently

**Breakdown by concern** (current):
- Value system: 9 files, ~2,500 lines
- Built-in functions: 20+ files, ~8,000 lines
- Expression evaluation: 5+ files, ~2,000 lines
- Statement execution: 5+ files, ~2,000 lines
- OOP (classes, objects): 5+ files, ~3,000 lines
- Type system (records, enums): 3+ files, ~1,500 lines
- Unit system: 2 files
- Environment/scope: 1 file

### 4. Heavy Type Assertion Usage

**Problem**: Code is littered with type assertions (`val.(*IntegerValue)`).

**Statistics**:
- lambda_test.go: 61 type assertions
- builtins_strings_basic.go: 56 type assertions
- helpers_conversion.go: 51 type assertions
- statements_assignments.go: 43 type assertions

**Impact**:
- Runtime panics if type assumptions are wrong
- Difficult to refactor value types
- No compile-time type safety
- Verbose and repetitive code

**Example**:
```go
// Common pattern repeated throughout
intVal, ok := val.(*IntegerValue)
if !ok {
    return newError("expected integer, got %s", val.Type())
}
result := intVal.Value + 1
```

### 5. Reflection Hacks for Circular Dependencies

**Problem**: `interpreter.go:107-132` uses reflection to extract fields from options to avoid importing `pkg/dwscript`.

**Impact**:
- No compile-time type safety
- Fragile coupling to struct field names
- Difficult to refactor
- Hidden dependencies

**Example**:
```go
// Current reflection hack
val := reflect.ValueOf(opts)
field := val.FieldByName("ExternalFunctions")
if registry, ok := field.Interface().(*ExternalFunctionRegistry); ok {
    interp.externalFunctions = registry
}
```

### 6. Inconsistent Error Handling

**Problem**: Multiple error handling patterns coexist:
- Return `ErrorValue` (most common)
- Return `ExceptionValue`
- Set `i.exception` field
- Use `raiseException()` method
- Check `i.exception != nil` after calls

**Impact**:
- Confusing for contributors
- Easy to miss error cases
- Difficult to add structured error context
- Exception state can leak between calls

### 7. Tight Coupling to AST

**Problem**: Interpreter directly works with internal AST types everywhere.

**Impact**:
- Cannot optimize AST representation
- Difficult to add bytecode compiler
- Cannot have multiple AST versions
- Hard to add instrumentation or profiling

### 8. Limited Testability

**Problem**: Current coverage is only **65.8%**. Many components cannot be tested in isolation.

**Impact**:
- Bugs in edge cases
- Fear of refactoring
- Difficult to add regression tests
- Manual integration testing required

**Missing tests for**:
- Complex error paths
- Edge cases in type coercion
- Exception propagation
- Unit initialization ordering
- Property getter/setter edge cases

---

## Proposed Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────────┐
│                    pkg/dwscript                             │
│  (Public API - Engine, Program, Options)                    │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│              internal/interp                                │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Interpreter (Facade/Coordinator)                      │ │
│  │  - Coordinates subsystems                              │ │
│  │  - Minimal state (just references to subsystems)       │ │
│  └────────────────────────────────────────────────────────┘ │
│                          │                                  │
│      ┌───────────────────┼───────────────────┐             │
│      ▼                   ▼                   ▼             │
│  ┌────────┐      ┌──────────────┐     ┌──────────┐        │
│  │ Exec   │      │   Runtime    │     │  Types   │        │
│  │ Context│      │   Services   │     │ Registry │        │
│  └────────┘      └──────────────┘     └──────────┘        │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │          runtime/ (sub-package)                      │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────────┐   │  │
│  │  │  Values  │  │ Builtins │  │   Environment    │   │  │
│  │  └──────────┘  └──────────┘  └──────────────────┘   │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │          eval/ (sub-package)                         │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────────┐ │  │
│  │  │ Statements │  │Expressions │  │ Declarations   │ │  │
│  │  │  Evaluator │  │  Evaluator │  │   Evaluator    │ │  │
│  │  └────────────┘  └────────────┘  └────────────────┘ │  │
│  │                                                       │  │
│  │  Uses Visitor Pattern from pkg/ast                   │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Component Breakdown

#### 1. Execution Context (`ExecutionContext`)

**Purpose**: Manages execution state for a single evaluation.

**Responsibilities**:
- Environment/scope chain
- Call stack tracking
- Control flow signals (break, continue, exit, return)
- Exception state
- Source location tracking

**Benefits**:
- Can be created per evaluation (enables parallelism)
- Easy to test execution state transitions
- Clear ownership of execution state

```go
type ExecutionContext struct {
    env            *Environment
    callStack      *CallStack
    controlFlow    *ControlFlow
    exception      *Exception
    sourceLocation *SourceLocation
}

type ControlFlow struct {
    breakSignal    bool
    continueSignal bool
    exitSignal     bool
    returnValue    Value
}
```

#### 2. Runtime Services (`RuntimeServices`)

**Purpose**: Provides runtime services needed during execution.

**Responsibilities**:
- I/O (output writer)
- Random number generation
- External function registry
- Time/date services

**Benefits**:
- Easy to mock for testing
- Clear interface for runtime dependencies
- Can be configured independently

```go
type RuntimeServices struct {
    output            io.Writer
    random            *RandomService
    externalFunctions *ExternalFunctionRegistry
    timeProvider      TimeProvider // for testability
}
```

#### 3. Type Registry (`TypeRegistry`)

**Purpose**: Manages type metadata and type IDs.

**Responsibilities**:
- Class/interface/record/enum type registration
- Type ID allocation
- Type lookup by name
- Helper method registration

**Benefits**:
- Centralized type management
- Easy to query type relationships
- Clear type system boundaries

```go
type TypeRegistry struct {
    classes   map[string]*ClassInfo
    records   map[string]*RecordTypeValue
    interfaces map[string]*InterfaceInfo
    enums     map[string]*EnumTypeValue
    helpers   map[string][]*HelperInfo

    typeIDs   *TypeIDAllocator
}

type TypeIDAllocator struct {
    classIDs  map[string]int
    recordIDs map[string]int
    enumIDs   map[string]int
    nextClass  int
    nextRecord int
    nextEnum   int
}
```

#### 4. Value System (`internal/interp/runtime/`)

**Purpose**: Define and manage runtime values.

**Changes**:
- Move to sub-package `internal/interp/runtime/`
- Introduce value interfaces for operations
- Separate primitive, reference, and composite types
- Add value pooling for performance

**Structure**:
```
internal/interp/runtime/
├── value.go           // Value interface, base types
├── primitives.go      // Integer, Float, String, Boolean
├── reference.go       // Nil, Reference, Variant
├── composite.go       // Array, Record, Set
├── object.go          // Class instances, interfaces
├── function.go        // Function pointers, lambdas
├── special.go         // Error, Exception, Type metadata
├── interfaces.go      // NumericValue, ComparableValue, etc.
├── pool.go            // Object pooling for performance
└── conversions.go     // Type conversion helpers
```

**Value Interfaces**:
```go
// Base interface (unchanged)
type Value interface {
    Type() string
    String() string
}

// Optional operation interfaces
type NumericValue interface {
    Value
    AsInteger() (int64, bool)
    AsFloat() (float64, bool)
}

type ComparableValue interface {
    Value
    CompareTo(other Value) (int, error)
}

type OrderableValue interface {
    Value
    LessThan(other Value) bool
}

type CopyableValue interface {
    Value
    Copy() Value
}

type ReferenceValue interface {
    Value
    IsReference() bool
}
```

#### 5. Evaluator Pattern (`internal/interp/eval/`)

**Purpose**: Replace giant switch with modular evaluators.

**Structure**:
```
internal/interp/eval/
├── evaluator.go       // Main Evaluator interface
├── statements.go      // Statement evaluator (implements Visitor)
├── expressions.go     // Expression evaluator (implements Visitor)
├── declarations.go    // Declaration evaluator
├── literals.go        // Literal evaluation (simple cases)
└── dispatch.go        // Top-level dispatch (replaces Eval())
```

**Pattern**:
```go
type Evaluator interface {
    Eval(node ast.Node, ctx *ExecutionContext) (Value, error)
}

type StatementEvaluator struct {
    exprEval *ExpressionEvaluator
    declEval *DeclarationEvaluator
    runtime  *RuntimeServices
    types    *TypeRegistry
}

// Use visitor pattern from pkg/ast
func (e *StatementEvaluator) VisitIfStatement(node *ast.IfStatement) (Value, error) {
    // Evaluate condition
    cond, err := e.exprEval.Eval(node.Condition, ctx)
    if err != nil {
        return nil, err
    }

    // Check if true
    boolVal, ok := cond.(*BooleanValue)
    if !ok {
        return nil, fmt.Errorf("condition must be boolean")
    }

    if boolVal.Value {
        return e.Eval(node.Consequence, ctx)
    } else if node.Alternative != nil {
        return e.Eval(node.Alternative, ctx)
    }
    return nil, nil
}
```

#### 6. Built-in Function Organization

**Current**: 20+ files in flat structure, all as Interpreter methods.

**Proposed**: Group into sub-packages by category:

```
internal/interp/runtime/builtins/
├── strings/
│   ├── basic.go       // Length, Pos, Copy, etc.
│   ├── transform.go   // UpperCase, LowerCase, Trim, etc.
│   ├── format.go      // Format, FormatFloat, etc.
│   └── compare.go     // CompareStr, SameText, etc.
├── math/
│   ├── basic.go       // Abs, Min, Max, etc.
│   ├── trig.go        // Sin, Cos, Tan, etc.
│   ├── convert.go     // IntToStr, FloatToStr, etc.
│   └── advanced.go    // Power, Sqrt, Exp, etc.
├── datetime/
│   ├── info.go        // Now, Date, Time, etc.
│   ├── format.go      // FormatDateTime, etc.
│   └── calc.go        // DateTimeAdd, etc.
├── collections/
│   ├── arrays.go      // Array helpers
│   ├── sets.go        // Set operations
│   └── strings.go     // String as collection
├── io/
│   └── output.go      // Print, PrintLn, Write
├── conversion/
│   └── convert.go     // StrToInt, StrToFloat, etc.
├── type_ops/
│   └── typeinfo.go    // TypeOf, ClassName, etc.
└── registry.go        // BuiltinRegistry for registration
```

**Benefits**:
- Clear organization by domain
- Easy to find and modify functions
- Can test each group independently
- Easier to document
- Reduce coupling (functions use BuiltinContext instead of full Interpreter)

```go
type BuiltinContext struct {
    Output   io.Writer
    Env      *Environment
    Random   *RandomService
    Types    *TypeRegistry
}

type BuiltinFunc func(ctx *BuiltinContext, args []Value) (Value, error)

type BuiltinRegistry struct {
    funcs map[string]BuiltinFunc
}
```

#### 7. Error Handling Strategy

**Unified approach**: Always return `(Value, error)` instead of special ErrorValue.

**Benefits**:
- Idiomatic Go
- Can use standard error wrapping
- Better stack traces
- Clear error propagation

**Exception handling**: Keep ExceptionValue for DWScript try-except, but store in ExecutionContext.

```go
// New pattern (preferred)
func (e *ExpressionEvaluator) Eval(node ast.Node, ctx *ExecutionContext) (Value, error) {
    switch n := node.(type) {
    case *ast.BinaryExpression:
        left, err := e.Eval(n.Left, ctx)
        if err != nil {
            return nil, fmt.Errorf("evaluating left operand: %w", err)
        }
        // ...
    }
}

// Exception handling (for DWScript semantics)
func (e *StatementEvaluator) EvalTryStatement(node *ast.TryStatement, ctx *ExecutionContext) (Value, error) {
    // Save exception state
    oldException := ctx.GetException()
    ctx.ClearException()

    // Execute try block
    _, err := e.Eval(node.TryBlock, ctx)

    // Check if exception was raised
    if exception := ctx.GetException(); exception != nil {
        // Handle exception in except blocks
        // ...
    }

    return nil, err
}
```

### Refactored Interpreter

After refactoring, the Interpreter becomes a **thin facade/coordinator**:

```go
type Interpreter struct {
    // Subsystems (composition over inheritance)
    runtime    *RuntimeServices
    types      *TypeRegistry
    units      *UnitManager
    evaluator  *Evaluator

    // Configuration
    config     *Config

    // Semantic analysis metadata (from parser)
    semanticInfo *pkgast.SemanticInfo
}

type Config struct {
    MaxRecursionDepth int
    SourceFile        string
    SourceCode        string
}

// Main entry point (simplified)
func (i *Interpreter) Eval(node ast.Node) (Value, error) {
    // Create execution context for this evaluation
    ctx := NewExecutionContext(
        NewEnvironment(),
        i.runtime,
        i.types,
    )

    // Delegate to evaluator
    return i.evaluator.Eval(node, ctx)
}
```

**Field count reduction**: 27 fields → 5 fields (81% reduction)

---

## Migration Strategy

### Principles

1. **Incremental refactoring** - No big-bang rewrite
2. **Tests always pass** - Every commit must pass all tests
3. **Backwards compatible** - Public API unchanged
4. **Measure progress** - Track metrics at each step
5. **Parallel development** - Don't block other work

### Phase-by-Phase Migration

#### Phase 3.2: Value System Refactoring (Week 1-2)

**Goal**: Move value types to sub-package with interfaces.

**Steps**:
1. Create `internal/interp/runtime/` package
2. Define value interfaces (NumericValue, ComparableValue, etc.)
3. Move value type definitions from value.go to runtime/
4. Update all imports (`interp.IntegerValue` → `runtime.IntegerValue`)
5. Add type assertions helpers to reduce verbosity
6. Run full test suite after each sub-step

**Validation**:
- All tests pass
- No performance regression (benchmarks)
- Value types implement new interfaces

**Files affected**: ~100 files (all files importing value types)

#### Phase 3.3: Execution Context Separation (Week 2-3)

**Goal**: Extract execution state into ExecutionContext.

**Steps**:
1. Create ExecutionContext struct with env, callStack, controlFlow
2. Thread ExecutionContext through all eval methods (alongside Interpreter)
3. Move state fields from Interpreter to ExecutionContext
4. Update all methods to use `ctx.Env` instead of `i.env`, etc.
5. Create new ExecutionContext per Eval() call

**Validation**:
- All tests pass
- Interpreter struct has fewer fields
- Execution state is isolated

**Files affected**: ~50 files (all files with eval methods)

#### Phase 3.4: Runtime Services Extraction (Week 3)

**Goal**: Extract I/O, random, external functions into RuntimeServices.

**Steps**:
1. Create RuntimeServices struct
2. Move output, rand, externalFunctions to RuntimeServices
3. Update Interpreter to hold RuntimeServices reference
4. Update all usages to `i.runtime.Output()` instead of `i.output`
5. Remove reflection hack by using proper interface

**Validation**:
- All tests pass
- Runtime dependencies are isolated
- Easy to mock for testing

**Files affected**: ~30 files (files using output, random, external functions)

#### Phase 3.5: Type Registry Extraction (Week 4)

**Goal**: Extract type system into TypeRegistry.

**Steps**:
1. Create TypeRegistry struct
2. Move classes, records, interfaces, helpers, type IDs to TypeRegistry
3. Add methods: RegisterClass, LookupClass, AllocateTypeID, etc.
4. Update Interpreter to hold TypeRegistry reference
5. Update all type registration and lookup code

**Validation**:
- All tests pass
- Type system is centralized
- Type queries are clear and consistent

**Files affected**: ~40 files (files declaring or using types)

#### Phase 3.6: Built-in Function Reorganization (Week 4-5)

**Goal**: Organize built-ins into sub-packages.

**Steps**:
1. Create `internal/interp/runtime/builtins/` structure
2. Move built-in functions to category sub-packages
3. Convert from Interpreter methods to standalone functions using BuiltinContext
4. Create BuiltinRegistry for registration
5. Update function calls to use registry

**Validation**:
- All tests pass
- Built-ins organized by category
- No Interpreter dependency for built-ins

**Files affected**: 20+ builtin files

#### Phase 3.7: Evaluator Pattern (Week 5-6)

**Goal**: Replace giant switch with modular evaluators.

**Steps**:
1. Create `internal/interp/eval/` package
2. Implement StatementEvaluator using visitor pattern
3. Implement ExpressionEvaluator using visitor pattern
4. Implement DeclarationEvaluator
5. Create top-level Evaluator that dispatches to sub-evaluators
6. Replace Interpreter.Eval() switch with evaluator delegation
7. Move eval methods to appropriate evaluators

**Validation**:
- All tests pass
- No giant switch statement
- Clear evaluator boundaries

**Files affected**: ~60 files (all files with eval logic)

#### Phase 3.8: Error Handling Unification (Week 6)

**Goal**: Standardize error handling.

**Steps**:
1. Update evaluators to return `(Value, error)` instead of ErrorValue
2. Update error creation to use standard errors
3. Add error wrapping for context
4. Keep ExceptionValue for DWScript try-except semantics
5. Update all error checks

**Validation**:
- All tests pass
- Consistent error handling
- Better error messages

**Files affected**: ~80 files (all files with error handling)

### Rollback Strategy

If any phase introduces regressions:

1. **Revert** the phase's commits (git revert)
2. **Analyze** test failures and benchmark regressions
3. **Fix** issues in a branch
4. **Re-attempt** merge when fixed

Each phase is small enough to revert without losing significant work.

---

## New Patterns and Conventions

### 1. Value Interface Pattern

**Use interfaces for operations instead of type assertions:**

```go
// OLD (bad - type assertion)
func add(a, b Value) Value {
    aInt, ok := a.(*IntegerValue)
    bInt, ok2 := b.(*IntegerValue)
    if ok && ok2 {
        return &IntegerValue{Value: aInt.Value + bInt.Value}
    }
    // ... more cases
}

// NEW (good - interface)
func add(a, b Value) (Value, error) {
    aNum, ok := a.(NumericValue)
    bNum, ok2 := b.(NumericValue)
    if !ok || !ok2 {
        return nil, fmt.Errorf("operands must be numeric")
    }

    // Let values convert themselves
    aFloat, _ := aNum.AsFloat()
    bFloat, _ := bNum.AsFloat()
    return &FloatValue{Value: aFloat + bFloat}, nil
}
```

### 2. Context Passing

**Always pass ExecutionContext explicitly:**

```go
// OLD (bad - Interpreter holds state)
func (i *Interpreter) evalBinaryExpression(node *ast.BinaryExpression) Value {
    left := i.Eval(node.Left)
    right := i.Eval(node.Right)
    // ...
}

// NEW (good - explicit context)
func (e *ExpressionEvaluator) EvalBinaryExpression(
    node *ast.BinaryExpression,
    ctx *ExecutionContext,
) (Value, error) {
    left, err := e.Eval(node.Left, ctx)
    if err != nil {
        return nil, err
    }
    right, err := e.Eval(node.Right, ctx)
    if err != nil {
        return nil, err
    }
    // ...
}
```

### 3. Error Wrapping

**Use error wrapping for context:**

```go
// OLD (bad - error loses context)
if err != nil {
    return newError("type mismatch")
}

// NEW (good - preserves context)
if err != nil {
    return nil, fmt.Errorf("evaluating binary expression: %w", err)
}
```

### 4. Built-in Function Pattern

**Built-ins use BuiltinContext instead of full Interpreter:**

```go
// OLD (bad - couples to Interpreter)
func (i *Interpreter) builtinLength(args []Value) Value {
    fmt.Fprintln(i.output, "calling Length")
    // ...
}

// NEW (good - minimal context)
func builtinLength(ctx *BuiltinContext, args []Value) (Value, error) {
    if len(args) != 1 {
        return nil, fmt.Errorf("Length expects 1 argument, got %d", len(args))
    }

    str, ok := args[0].(*StringValue)
    if !ok {
        return nil, fmt.Errorf("Length expects string, got %s", args[0].Type())
    }

    return &IntegerValue{Value: int64(len(str.Value))}, nil
}
```

### 5. Component Creation

**Use factory functions for consistency:**

```go
// Create execution context
func NewExecutionContext(env *Environment, runtime *RuntimeServices, types *TypeRegistry) *ExecutionContext {
    return &ExecutionContext{
        env:         env,
        callStack:   NewCallStack(),
        controlFlow: &ControlFlow{},
        runtime:     runtime,
        types:       types,
    }
}

// Create evaluator
func NewEvaluator(runtime *RuntimeServices, types *TypeRegistry) *Evaluator {
    return &Evaluator{
        statements:   NewStatementEvaluator(runtime, types),
        expressions:  NewExpressionEvaluator(runtime, types),
        declarations: NewDeclarationEvaluator(runtime, types),
    }
}
```

### 6. Testing Pattern

**Test components in isolation with mocks:**

```go
func TestStringBuiltins(t *testing.T) {
    // Create minimal context (no full interpreter needed)
    ctx := &BuiltinContext{
        Output: io.Discard,
    }

    // Test Length
    result, err := builtinLength(ctx, []Value{&StringValue{Value: "hello"}})
    require.NoError(t, err)

    intVal, ok := result.(*IntegerValue)
    require.True(t, ok)
    assert.Equal(t, int64(5), intVal.Value)
}
```

---

## Success Metrics

### Code Quality Metrics

| Metric | Current | Target | Improvement |
|--------|---------|--------|-------------|
| Interpreter fields | 27 | 5 | -81% |
| Eval() switch lines | 228 | 0 (delegated) | -100% |
| Package structure | 1 flat | 3 layered | Modular |
| Largest file | 1,248 lines | <500 lines | -60% |
| Files with Interpreter methods | 56 | ~10 | -82% |

### Test Metrics

| Metric | Current | Target |
|--------|---------|--------|
| Test coverage | 65.8% | 90%+ |
| Unit tests | Few | Many (isolated components) |
| Benchmark coverage | ~10 benchmarks | 50+ benchmarks |
| Tests passing | ~85% | 100% |

### Performance Metrics

| Metric | Target |
|--------|--------|
| Execution time | No regression (within 5%) |
| Memory allocations | Reduce by 10-20% (via pooling) |
| Benchmark suite | All benchmarks stable |

### Maintainability Metrics

| Metric | How to Measure |
|--------|----------------|
| Time to add built-in | <30 minutes (vs ~2 hours current) |
| Time to add AST node | <1 hour (vs ~4 hours current) |
| Cyclomatic complexity | <15 per function (vs 30+ current) |
| Package dependencies | Clear DAG (no cycles) |

---

## Timeline and Phases

### Overview

**Total estimated time**: 6 weeks
**Team size**: 1 developer
**Work style**: Incremental, test-driven

### Week-by-Week Breakdown

| Week | Phase | Tasks | Validation |
|------|-------|-------|------------|
| **1** | 3.1 Preparation | Analysis, design doc, benchmarks, tests | 90% coverage, 50+ benchmarks |
| **2** | 3.2 Values | Move to runtime/, add interfaces | Tests pass, no regression |
| **2-3** | 3.3 Context | Extract ExecutionContext | Tests pass, state isolated |
| **3** | 3.4 Runtime | Extract RuntimeServices | Tests pass, easy mocking |
| **4** | 3.5 Types | Extract TypeRegistry | Tests pass, types centralized |
| **4-5** | 3.6 Built-ins | Reorganize to sub-packages | Tests pass, clear organization |
| **5-6** | 3.7 Evaluators | Replace switch with visitor | Tests pass, no switch |
| **6** | 3.8 Errors | Unify error handling | Tests pass, consistent errors |

### Milestones

- **End of Week 1**: Design approved, baseline established
- **End of Week 3**: Core refactoring done (values, context, runtime)
- **End of Week 5**: Architecture transformation complete (types, built-ins, evaluators)
- **End of Week 6**: Polish and error handling unified

---

## Risks and Mitigations

### Risk 1: Test Failures

**Risk**: Refactoring breaks existing tests.

**Likelihood**: High
**Impact**: High

**Mitigation**:
- Run full test suite after every change
- Use TDD: write tests first, then refactor
- Keep changes small and atomic
- Use git bisect to find breaking commits

### Risk 2: Performance Regression

**Risk**: New architecture is slower.

**Likelihood**: Medium
**Impact**: High

**Mitigation**:
- Establish baseline benchmarks (Phase 3.1.2)
- Run benchmarks after each phase
- Profile before and after
- Add value pooling to reduce allocations (Phase 3.2.3)
- Set performance gates in CI

### Risk 3: Scope Creep

**Risk**: Refactoring expands beyond plan.

**Likelihood**: Medium
**Impact**: Medium

**Mitigation**:
- Stick to defined phases
- Document any new issues in backlog, don't fix immediately
- Time-box each phase
- Focus on non-goals (no behavior changes)

### Risk 4: Merge Conflicts

**Risk**: Other development conflicts with refactoring.

**Likelihood**: Low (currently solo project)
**Impact**: Low

**Mitigation**:
- Communicate refactoring plan
- Freeze feature development during refactoring
- Use feature branch and rebase frequently
- Break into small PRs that can be reviewed and merged independently

### Risk 5: Incomplete Migration

**Risk**: Some code uses old patterns, some uses new.

**Likelihood**: Medium
**Impact**: Medium

**Mitigation**:
- Complete each phase fully before moving on
- Use automated tools to find old patterns (grep, linters)
- Add linter rules to enforce new patterns
- Document patterns clearly in this doc

### Risk 6: Hidden Dependencies

**Risk**: Circular dependencies or hidden coupling emerges.

**Likelihood**: Medium
**Impact**: Medium

**Mitigation**:
- Draw dependency diagrams before each phase
- Use `go list -deps` to check actual dependencies
- Enforce layering with architecture tests
- Refactor incrementally to expose issues early

---

## Appendix A: Current Package Statistics

Generated 2025-11-16:

```
$ find internal/interp -name "*.go" | wc -l
149

$ find internal/interp -name "*.go" ! -name "*_test.go" | wc -l
51

$ find internal/interp -name "*_test.go" | wc -l
98

$ find internal/interp -name "*.go" ! -name "*_test.go" -exec wc -l {} + | tail -1
30602 total

$ go test -coverprofile=coverage.out ./internal/interp
coverage: 65.8% of statements
```

**Largest files**:
1. value.go: 1,248 lines
2. builtins_strings_basic.go: 1,182 lines
3. objects_methods.go: 1,114 lines
4. objects_hierarchy.go: 976 lines
5. declarations.go: 932 lines

**Files with most type assertions**:
1. lambda_test.go: 61
2. builtins_strings_basic.go: 56
3. helpers_conversion.go: 51
4. math_basic_test.go: 44
5. builtins_strings_advanced.go: 44

**Files with most Interpreter methods**:
1. builtins_strings_basic.go: 34
2. builtins_math_basic.go: 28
3. builtins_strings_advanced.go: 21
4. builtins_datetime_info.go: 21
5. builtins_datetime_format.go: 18

---

## Appendix B: Dependency Graph (Current)

```
pkg/dwscript
    ↓ (reflection hack)
internal/interp ← Everything in one package!
    ├── Interpreter (God object)
    ├── Environment
    ├── Value (30+ types)
    ├── Eval methods (scattered across 56 files)
    ├── Built-in functions (20+ files)
    ├── Type system (classes, records, enums)
    └── Unit loader
    ↓
internal/ast, internal/types, internal/units
```

**Problems**:
- Circular dependency risk (avoided via reflection)
- No clear layers
- Everything couples to Interpreter

## Appendix C: Dependency Graph (Proposed)

```
pkg/dwscript (public API)
    ↓
internal/interp (facade)
    ├── Interpreter (thin coordinator)
    ├── Config
    ↓
    ├── internal/interp/runtime (value system)
    │   ├── Values (primitives, reference, composite)
    │   ├── Environment
    │   ├── CallStack
    │   ├── Value interfaces
    │   └── builtins/ (organized by category)
    │
    ├── internal/interp/eval (evaluation logic)
    │   ├── Evaluator (dispatcher)
    │   ├── StatementEvaluator
    │   ├── ExpressionEvaluator
    │   └── DeclarationEvaluator
    │
    └── internal/interp (core types)
        ├── ExecutionContext
        ├── RuntimeServices
        ├── TypeRegistry
        └── UnitManager
    ↓
internal/ast, internal/types, internal/units
pkg/ast (visitor pattern)
```

**Benefits**:
- Clear layers (API → facade → subsystems → runtime)
- No circular dependencies
- Each layer has defined responsibilities
- Easy to test each layer independently

---

## Appendix D: Example Refactoring (Before/After)

### Before: Evaluating If Statement

```go
// In interpreter.go (part of 228-line switch)
func (i *Interpreter) Eval(node ast.Node) Value {
    switch node := node.(type) {
    // ... 40+ other cases
    case *ast.IfStatement:
        return i.evalIfStatement(node)
    // ... more cases
    }
}

// In statements.go
func (i *Interpreter) evalIfStatement(node *ast.IfStatement) Value {
    condition := i.Eval(node.Condition)
    if isError(condition) {
        return condition
    }

    if isTruthy(condition) {
        return i.Eval(node.Consequence)
    } else if node.Alternative != nil {
        return i.Eval(node.Alternative)
    }
    return nil
}
```

**Problems**:
- Giant switch in one place
- Returns special ErrorValue
- Uses global `i.env` state
- Hard to test in isolation
- Heavy coupling to Interpreter

### After: Evaluating If Statement

```go
// In eval/evaluator.go (top-level dispatcher)
func (e *Evaluator) Eval(node ast.Node, ctx *ExecutionContext) (Value, error) {
    // Use visitor pattern from pkg/ast
    return pkgast.Walk(node, &evaluatorVisitor{
        statements:   e.statements,
        expressions:  e.expressions,
        declarations: e.declarations,
        ctx:          ctx,
    })
}

// In eval/statements.go (visitor implementation)
func (e *StatementEvaluator) VisitIfStatement(
    node *ast.IfStatement,
    ctx *ExecutionContext,
) (Value, error) {
    // Evaluate condition
    condition, err := e.exprEval.Eval(node.Condition, ctx)
    if err != nil {
        return nil, fmt.Errorf("evaluating if condition: %w", err)
    }

    // Check if condition is boolean
    boolVal, ok := condition.(runtime.BooleanValue)
    if !ok {
        return nil, fmt.Errorf("if condition must be boolean, got %s", condition.Type())
    }

    // Execute appropriate branch
    if boolVal.Value {
        return e.Eval(node.Consequence, ctx)
    } else if node.Alternative != nil {
        return e.Eval(node.Alternative, ctx)
    }

    return nil, nil
}
```

**Benefits**:
- No giant switch (visitor pattern)
- Returns standard Go errors
- Uses explicit ExecutionContext
- Easy to test (mock context)
- Minimal coupling (only to ExpressionEvaluator and ExecutionContext)

### Testing Comparison

**Before** (can't test in isolation):
```go
func TestIfStatement(t *testing.T) {
    // Must create full Interpreter with all subsystems
    interp := interp.New(os.Stdout)

    // Must parse full program
    program := parseProgram(t, `
        var x := 0;
        if true then
            x := 1;
    `)

    // Eval entire program
    result := interp.Eval(program)

    // Check global state (ugh)
    val, _ := interp.env.Get("x")
    assert.Equal(t, int64(1), val.(*IntegerValue).Value)
}
```

**After** (clean unit test):
```go
func TestIfStatement(t *testing.T) {
    // Create minimal context for test
    ctx := eval.NewExecutionContext(
        runtime.NewEnvironment(),
        &runtime.RuntimeServices{Output: io.Discard},
        runtime.NewTypeRegistry(),
    )

    // Create evaluator
    evaluator := eval.NewStatementEvaluator(/* ... */)

    // Test if-then
    ifStmt := &ast.IfStatement{
        Condition:   &ast.BooleanLiteral{Value: true},
        Consequence: &ast.BlockStatement{/* ... */},
    }

    result, err := evaluator.VisitIfStatement(ifStmt, ctx)
    require.NoError(t, err)
    // Assert on result
}
```

---

## Appendix E: References

- Original DWScript source: `reference/dwscript-original/`
- PLAN.md: Phase 3 tasks
- Existing architecture docs:
  - `docs/architecture/bytecode-vm-design.md`
  - `docs/architecture/implicit-self-resolution.md`
- Go best practices:
  - [Effective Go](https://golang.org/doc/effective_go)
  - [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
  - [Standard Go Project Layout](https://github.com/golang-standards/project-layout)

---

## Conclusion

This refactoring will transform the interpreter from a monolithic 81K-line package with a 27-field God object into a clean, modular architecture with clear separation of concerns. The incremental migration strategy ensures we maintain stability while making steady progress over 6 weeks.

**Key outcomes**:
- **Maintainability**: 81% reduction in Interpreter fields, elimination of giant switch
- **Testability**: 90%+ coverage, isolated component testing
- **Extensibility**: Clear patterns for adding features
- **Performance**: No regression, potential 10-20% allocation improvement
- **Foundation**: Ready for future optimizations and language features

The refactoring enables confident development going forward and sets the stage for completing the remaining phases of the project (OOP features, advanced language features, standard library).
