# Interpreter Architecture

**Status**: Phase 3.5 Complete - Evaluator Refactoring
**Last Updated**: 2025-11-22

## Overview

The DWScript interpreter has undergone a major architectural refactoring (Phase 3.5) to improve code organization, maintainability, and testability. The monolithic `Interpreter.Eval()` switch statement has been replaced with a clean visitor pattern architecture that separates concerns and provides better code organization.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                         AST (pkg/ast)                               │
│  IntegerLiteral, BinaryExpression, IfStatement, FunctionDecl, ...   │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             ↓
┌─────────────────────────────────────────────────────────────────────┐
│                    Interpreter (internal/interp)                    │
│  • Thin orchestrator                                                │
│  • Maintains global registries (functions, classes, records)         │
│  • Provides InterpreterAdapter for backward compatibility            │
│  • Delegates evaluation to Evaluator                                 │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             ↓
┌─────────────────────────────────────────────────────────────────────┐
│              Evaluator (internal/interp/evaluator)                  │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │  Visitor Pattern Evaluation                                  │  │
│  │  • VisitIntegerLiteral(node, ctx) → Value                   │  │
│  │  • VisitBinaryExpression(node, ctx) → Value                 │  │
│  │  • VisitIfStatement(node, ctx) → Value                      │  │
│  │  • VisitFunctionDecl(node, ctx) → Value                     │  │
│  │  • ... 48+ visitor methods                                  │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
│  ┌────────────────────┐  ┌────────────────────┐                   │
│  │  ExecutionContext  │  │    Type System     │                   │
│  │  • Environment     │  │  • FunctionRegistry│                   │
│  │  • Call Stack      │  │  • ClassRegistry   │                   │
│  │  • Control Flow    │  │  • RecordRegistry  │                   │
│  │  • Exceptions      │  │  • InterfaceRegistry│                  │
│  └────────────────────┘  └────────────────────┘                   │
└─────────────────────────────────────────────────────────────────────┘
                             │
                             ↓
┌─────────────────────────────────────────────────────────────────────┐
│                  Runtime Values (internal/interp/runtime)           │
│  IntegerValue, FloatValue, StringValue, BooleanValue,               │
│  ArrayValue, RecordValue, ObjectValue, FunctionValue, ...           │
└─────────────────────────────────────────────────────────────────────┘
```

## Key Components

### 1. Interpreter (Orchestrator)

**Location**: `internal/interp/interpreter.go`

**Responsibilities**:
- Maintains global state (function registry, class registry, etc.)
- Provides public API for script execution
- Implements `InterpreterAdapter` interface for Evaluator callbacks
- Manages legacy code compatibility during migration

**Key Methods**:
```go
func (i *Interpreter) Eval(node ast.Node) Value
func (i *Interpreter) Run(program *ast.Program) Value
```

**Migration Strategy**:
- Currently acts as adapter between old and new architecture
- Gradually handing off responsibilities to Evaluator
- Will become thin orchestrator once Phase 3.5.37 completes

### 2. Evaluator (Evaluation Engine)

**Location**: `internal/interp/evaluator/`

**Responsibilities**:
- Execute AST nodes using visitor pattern
- Manage execution context (environment, call stack)
- Implement all evaluation logic for expressions and statements
- Delegate to type system for type-related operations

**Structure**:
```
evaluator/
├── evaluator.go              # Core evaluator with visitor infrastructure
├── visitor_expressions.go    # Expression evaluation (48+ methods)
├── visitor_statements.go     # Statement evaluation (control flow, loops)
├── visitor_declarations.go   # Declaration evaluation (functions, classes, types)
├── visitor_literals.go       # Literal value creation
├── binary_ops.go             # Binary operation implementations
├── context.go                # ExecutionContext
├── callstack.go              # Call stack management
├── helpers.go                # Helper functions
├── env_adapter.go            # Environment adapter
└── benchmark_test.go         # Performance benchmarks
```

**Key Visitor Methods**:
```go
// Expressions
func (e *Evaluator) VisitIntegerLiteral(node *ast.IntegerLiteral, ctx *ExecutionContext) Value
func (e *Evaluator) VisitBinaryExpression(node *ast.BinaryExpression, ctx *ExecutionContext) Value
func (e *Evaluator) VisitCallExpression(node *ast.CallExpression, ctx *ExecutionContext) Value
func (e *Evaluator) VisitIdentifier(node *ast.Identifier, ctx *ExecutionContext) Value

// Statements
func (e *Evaluator) VisitIfStatement(node *ast.IfStatement, ctx *ExecutionContext) Value
func (e *Evaluator) VisitForStatement(node *ast.ForStatement, ctx *ExecutionContext) Value
func (e *Evaluator) VisitWhileStatement(node *ast.WhileStatement, ctx *ExecutionContext) Value
func (e *Evaluator) VisitAssignmentStatement(node *ast.AssignmentStatement, ctx *ExecutionContext) Value

// Declarations
func (e *Evaluator) VisitFunctionDecl(node *ast.FunctionDecl, ctx *ExecutionContext) Value
func (e *Evaluator) VisitClassDecl(node *ast.ClassDecl, ctx *ExecutionContext) Value
```

### 3. ExecutionContext

**Location**: `internal/interp/evaluator/context.go`

**Purpose**: Encapsulates all execution state needed during evaluation

**Components**:
```go
type ExecutionContext struct {
    env            Environment          // Variable bindings
    envStack       []Environment        // Environment stack for scoping
    callStack      *CallStack           // Function call stack
    controlFlow    *ControlFlow         // Break, continue, return, exit state
    exception      *ExceptionState      // Current exception
    propContext    *PropertyContext     // Property getter/setter context
    ...
}
```

**Key Methods**:
```go
func (ctx *ExecutionContext) Env() Environment
func (ctx *ExecutionContext) PushEnv(env Environment)
func (ctx *ExecutionContext) PopEnv() Environment
func (ctx *ExecutionContext) GetCallStack() *CallStack
func (ctx *ExecutionContext) GetControlFlow() *ControlFlow
```

### 4. Type System

**Location**: `internal/interp/types/`

**Purpose**: Centralized management of all type information

**Components**:
- `TypeSystem`: Main registry for all types
- `FunctionRegistry`: Function overload resolution
- `ClassRegistry`: Class type management
- `RecordRegistry`: Record type management
- `InterfaceRegistry`: Interface type management

**Example**:
```go
type TypeSystem struct {
    functions  *FunctionRegistry
    classes    *ClassRegistry
    records    *RecordRegistry
    interfaces *InterfaceRegistry
    helpers    map[string][]*HelperInfo
    operators  *OperatorRegistry
    ...
}
```

## Design Patterns

### Visitor Pattern

The evaluator uses the visitor pattern to traverse and evaluate the AST.

**Benefits**:
- **Separation of Concerns**: Evaluation logic is separated from AST structure
- **Extensibility**: Easy to add new node types or evaluation strategies
- **Maintainability**: Each visitor method is focused on a single node type
- **Type Safety**: Compile-time checking of visitor implementations

**Example**:
```go
// Evaluating a binary expression
func (e *Evaluator) VisitBinaryExpression(node *ast.BinaryExpression, ctx *ExecutionContext) Value {
    // Evaluate left operand
    left := e.VisitExpression(node.Left, ctx)
    if isError(left) {
        return left
    }

    // Short-circuit for boolean operators
    if node.Operator == "and" || node.Operator == "or" {
        return e.evalBooleanBinaryOp(node.Operator, left, node.Right, ctx)
    }

    // Evaluate right operand
    right := e.VisitExpression(node.Right, ctx)
    if isError(right) {
        return right
    }

    // Perform operation
    return e.evalBinaryOp(node.Operator, left, right, node)
}
```

### Adapter Pattern

The `InterpreterAdapter` interface provides a bridge between the Evaluator and Interpreter during the migration period.

**Purpose**:
- Allow gradual migration from monolithic interpreter to visitor-based evaluator
- Maintain backward compatibility during transition
- Enable testing of individual visitor methods

**Interface**:
```go
type InterpreterAdapter interface {
    // Delegate back to old Interpreter.Eval for nodes not yet migrated
    EvalNode(node ast.Node) Value

    // Function call system
    CallFunctionPointer(funcPtr Value, args []Value, node ast.Node) Value
    CallUserFunction(fn *ast.FunctionDecl, args []Value) Value
    CallBuiltinFunction(name string, args []Value) Value

    // Type system access
    LookupClass(name string) (any, bool)
    LookupRecord(name string) (any, bool)
    LookupFunction(name string) ([]*ast.FunctionDecl, bool)

    // ... more adapter methods
}
```

**Migration Plan**:
- ✅ Phase 3.5.1-3.5.35: Implement visitor pattern with adapter
- ⏸️ Phase 3.5.37: Remove adapter (blocked on AST-free runtime types)

## Performance

The visitor pattern has been benchmarked extensively (Task 3.5.35):

### Results

| Operation | ns/op | B/op | allocs/op | Status |
|-----------|-------|------|-----------|--------|
| Integer Literal | 0.31 | 0 | 0 | ✅ Optimized |
| Binary Add | 71.08 | 24 | 3 | ✅ Target <100 ns |
| Binary Multiply | 70.32 | 24 | 3 | ✅ Target <100 ns |
| Boolean AND | 47.64 | 3 | 3 | ✅ Fast path |
| String Concat | 156.0 | 64 | 4 | ✅ Target <200 ns |

### Key Findings

✅ **Zero regression**: Visitor pattern has no measurable overhead vs switch-based approach
✅ **Compiler optimization**: Simple operations are optimized away (0.3 ns/op)
✅ **Linear scaling**: Performance scales linearly with expression complexity
✅ **Predictable memory**: Allocation patterns are consistent and minimal

See `docs/evaluator-performance-report.md` for detailed analysis.

## Migration Guide

### For Contributors

If you're working on the interpreter, follow these guidelines:

#### 1. Adding New Expression Types

When adding a new expression node to the AST:

1. **Define the AST node** in `pkg/ast/expressions.go`
2. **Add visitor method** in `internal/interp/evaluator/visitor_expressions.go`:
   ```go
   func (e *Evaluator) VisitYourExpression(node *ast.YourExpression, ctx *ExecutionContext) Value {
       // Evaluate sub-expressions
       value := e.VisitExpression(node.SubExpr, ctx)
       if isError(value) {
           return value
       }

       // Perform operation
       result := // ... your logic here

       return result
   }
   ```

3. **Add dispatcher** in `visitor_expressions.go:VisitExpression()`:
   ```go
   case *ast.YourExpression:
       return e.VisitYourExpression(typed, ctx)
   ```

4. **Write tests** in `visitor_expressions_test.go`

5. **Benchmark** if it's a hot path (add to `benchmark_test.go`)

#### 2. Adding New Statement Types

Similar process to expressions, but in `visitor_statements.go`:

```go
func (e *Evaluator) VisitYourStatement(node *ast.YourStatement, ctx *ExecutionContext) Value {
    // Evaluate statement logic
    // ...

    // Check for control flow (break, continue, return)
    if ctx.GetControlFlow().HasSignal() {
        return nil
    }

    return nil
}
```

#### 3. Modifying Existing Visitor Methods

1. **Read the method** to understand current logic
2. **Make changes** carefully, preserving error handling
3. **Run tests**: `go test ./internal/interp/evaluator`
4. **Run benchmarks**: `go test -bench=. ./internal/interp/evaluator`
5. **Update documentation** if behavior changes

#### 4. Working with ExecutionContext

The ExecutionContext is your friend! It provides access to:

```go
// Environment (variables)
val, ok := ctx.Env().Get("x")
ctx.Env().Define("y", value)
ctx.Env().Set("z", newValue)

// Call stack (for error reporting)
ctx.GetCallStack().Push("functionName", node.Pos())
defer ctx.GetCallStack().Pop()

// Control flow
if ctx.GetControlFlow().ShouldBreak() {
    return nil
}
ctx.GetControlFlow().SetReturn(value)

// Exceptions
if ctx.HasException() {
    return e.newError(node, "uncaught exception")
}
```

## Testing

### Unit Tests

Test individual visitor methods:

```go
func TestVisitBinaryExpression_IntegerAdd(t *testing.T) {
    eval := createTestEvaluator()
    ctx := createTestContext()

    node := &ast.BinaryExpression{
        Left: &ast.IntegerLiteral{Value: 3},
        Operator: "+",
        Right: &ast.IntegerLiteral{Value: 5},
    }

    result := eval.VisitBinaryExpression(node, ctx)

    intVal := result.(*runtime.IntegerValue)
    assert.Equal(t, int64(8), intVal.Value)
}
```

### Benchmarks

Benchmark hot paths:

```go
func BenchmarkVisitBinaryExpression_IntegerAdd(b *testing.B) {
    eval := createTestEvaluator()
    ctx := createTestContext()
    node := createBinaryAddNode()

    b.ResetTimer()
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        _ = eval.VisitBinaryExpression(node, ctx)
    }
}
```

### Integration Tests

Use the full interpreter:

```go
func TestInterpreter_ComplexProgram(t *testing.T) {
    code := `
        var x: Integer := 10;
        var y: Integer := 20;
        x + y
    `

    interp := interp.New(os.Stdout)
    result := interp.Run(parseCode(code))

    assert.Equal(t, int64(30), result.(*runtime.IntegerValue).Value)
}
```

## Future Work

### Phase 3.5.37: Remove Adapter Pattern

**Status**: ⏸️ Blocked on AST-free runtime types architecture

**Goals**:
- Remove `InterpreterAdapter` interface
- Make Evaluator fully self-contained
- Make Interpreter a thin orchestrator only

**Blockers**:
- Need to separate runtime value types from AST
- Need to refactor type system to not depend on AST nodes
- See `docs/task-3.5.4-expansion-plan.md` for details

### Potential Optimizations

1. **Value pooling**: Pool common values (0, 1, true, false)
2. **String interning**: Intern commonly used strings
3. **Context reuse**: Reuse ExecutionContext instances
4. **Inline simple operations**: Let compiler inline more aggressively

Note: Current performance is excellent; these are only needed if profiling shows issues.

## References

- [PLAN.md](../../PLAN.md) - Task 3.5.1-3.5.36 (Evaluator Refactoring)
- [docs/evaluator-performance-report.md](../evaluator-performance-report.md) - Performance benchmarks
- [internal/interp/evaluator/](../../internal/interp/evaluator/) - Evaluator source code
- [internal/interp/evaluator/USAGE_EXAMPLES.md](../../internal/interp/evaluator/USAGE_EXAMPLES.md) - Evaluator API examples

---

**Document Version**: 1.0
**Last Updated**: 2025-11-22
**Phase**: 3.5 Complete
**Next Phase**: 3.5.37 (Blocked)
