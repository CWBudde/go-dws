# Evaluator Migration Guide

**For Contributors**: This guide explains how to work with the new visitor-based evaluator architecture (Phase 3.5).

**Last Updated**: 2025-11-22
**Phase**: 3.5 Complete

## Overview

The DWScript interpreter has been refactored from a monolithic `Interpreter.Eval()` switch statement to a clean visitor pattern architecture. This guide helps contributors understand the new structure and how to work with it.

## What Changed?

### Before (Old Architecture)

```go
// All evaluation logic in one giant switch statement
func (i *Interpreter) Eval(node ast.Node) Value {
    switch node := node.(type) {
    case *ast.IntegerLiteral:
        return &IntegerValue{Value: node.Value}
    case *ast.BinaryExpression:
        left := i.Eval(node.Left)
        right := i.Eval(node.Right)
        // ... 100+ lines of binary operation logic
    case *ast.IfStatement:
        // ... if statement logic
    // ... 50+ more cases ...
    }
}
```

**Problems**:
- Single 3000+ line function
- Hard to test individual operations
- Poor code organization
- Difficult to extend

### After (New Architecture)

```go
// Evaluator with focused visitor methods
func (e *Evaluator) VisitIntegerLiteral(node *ast.IntegerLiteral, ctx *ExecutionContext) Value {
    return &runtime.IntegerValue{Value: node.Value}
}

func (e *Evaluator) VisitBinaryExpression(node *ast.BinaryExpression, ctx *ExecutionContext) Value {
    // Focused logic for binary expressions
    // Delegates to specialized helpers
    return e.evalBinaryOp(node.Operator, left, right, node)
}

func (e *Evaluator) VisitIfStatement(node *ast.IfStatement, ctx *ExecutionContext) Value {
    // Focused logic for if statements
}
```

**Benefits**:
- ✅ Focused, testable methods
- ✅ Clear code organization
- ✅ Easy to extend
- ✅ Zero performance overhead

## Quick Start

### Adding a New Expression Type

**Step 1**: Define the AST node in `pkg/ast/expressions.go`

```go
// ConditionalExpression represents a ternary expression
// Example: x > 5 ? 10 : 20
type ConditionalExpression struct {
    TypedExpressionBase
    Condition   Expression
    TrueValue   Expression
    FalseValue  Expression
}
```

**Step 2**: Add visitor method in `internal/interp/evaluator/visitor_expressions.go`

```go
// VisitConditionalExpression evaluates ternary conditional expressions
func (e *Evaluator) VisitConditionalExpression(node *ast.ConditionalExpression, ctx *ExecutionContext) Value {
    // Evaluate condition
    condition := e.VisitExpression(node.Condition, ctx)
    if isError(condition) {
        return condition
    }

    // Check if condition is true
    if isTruthy(condition) {
        return e.VisitExpression(node.TrueValue, ctx)
    }
    return e.VisitExpression(node.FalseValue, ctx)
}
```

**Step 3**: Add dispatcher in `VisitExpression()` (same file)

```go
func (e *Evaluator) VisitExpression(expr ast.Expression, ctx *ExecutionContext) Value {
    switch typed := expr.(type) {
    // ... existing cases ...
    case *ast.ConditionalExpression:
        return e.VisitConditionalExpression(typed, ctx)
    default:
        return e.newError(expr, "unknown expression type: %T", expr)
    }
}
```

**Step 4**: Write tests

```go
func TestVisitConditionalExpression(t *testing.B) {
    eval := createTestEvaluator()
    ctx := createTestContext()

    // Test true branch
    node := &ast.ConditionalExpression{
        Condition: &ast.BooleanLiteral{Value: true},
        TrueValue: &ast.IntegerLiteral{Value: 10},
        FalseValue: &ast.IntegerLiteral{Value: 20},
    }

    result := eval.VisitConditionalExpression(node, ctx)
    assert.Equal(t, int64(10), result.(*runtime.IntegerValue).Value)
}
```

**Step 5**: Add benchmark if it's a hot path

```go
func BenchmarkVisitConditionalExpression(b *testing.B) {
    eval := createTestEvaluator()
    ctx := createTestContext()
    node := createConditionalNode()

    b.ResetTimer()
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        _ = eval.VisitConditionalExpression(node, ctx)
    }
}
```

### Adding a New Statement Type

Similar process, but in `visitor_statements.go`:

```go
// VisitWithStatement evaluates a 'with' statement
func (e *Evaluator) VisitWithStatement(node *ast.WithStatement, ctx *ExecutionContext) Value {
    // Evaluate the object
    obj := e.VisitExpression(node.Object, ctx)
    if isError(obj) {
        return obj
    }

    // Push new environment with object fields accessible
    newEnv := ctx.Env().NewEnclosedEnvironment()
    // ... populate environment with object fields ...
    ctx.PushEnv(newEnv)
    defer ctx.PopEnv()

    // Execute body
    result := e.VisitStatement(node.Body, ctx)

    // Check for control flow signals
    if ctx.GetControlFlow().HasSignal() {
        return result
    }

    return nil
}
```

## Working with ExecutionContext

The `ExecutionContext` is your primary interface for managing evaluation state.

### Environment (Variables)

```go
// Get a variable
val, ok := ctx.Env().Get("x")
if !ok {
    return e.newError(node, "undefined variable: x")
}

// Define a new variable
ctx.Env().Define("y", &runtime.IntegerValue{Value: 42})

// Set an existing variable
if !ctx.Env().Set("z", newValue) {
    return e.newError(node, "undefined variable: z")
}

// Create enclosed scope
newEnv := ctx.Env().NewEnclosedEnvironment()
ctx.PushEnv(newEnv)
defer ctx.PopEnv()
```

### Call Stack (Function Calls)

```go
// Push function call onto stack
ctx.GetCallStack().Push("myFunction", node.Pos())
defer ctx.GetCallStack().Pop()

// Check recursion depth
if ctx.GetCallStack().Depth() > e.config.MaxRecursionDepth {
    return e.newError(node, "stack overflow: max recursion depth exceeded")
}
```

### Control Flow (Break, Continue, Return, Exit)

```go
// In a loop statement
for {
    result := e.VisitStatement(stmt, ctx)

    // Check for break
    if ctx.GetControlFlow().ShouldBreak() {
        ctx.GetControlFlow().ClearBreak()
        break
    }

    // Check for continue
    if ctx.GetControlFlow().ShouldContinue() {
        ctx.GetControlFlow().ClearContinue()
        continue
    }

    // Check for return or exit
    if ctx.GetControlFlow().ShouldReturn() || ctx.GetControlFlow().ShouldExit() {
        return result
    }
}

// Set control flow signals
ctx.GetControlFlow().SetBreak()
ctx.GetControlFlow().SetContinue()
ctx.GetControlFlow().SetReturn(value)
ctx.GetControlFlow().SetExit(value)
```

### Exception Handling

```go
// Check for exception
if ctx.HasException() {
    return e.newError(node, "uncaught exception: %v", ctx.GetException())
}

// In try-catch blocks
ctx.SetException(exceptionValue)
ctx.ClearException()
```

## Common Patterns

### Error Handling

Always check for errors from sub-evaluations:

```go
func (e *Evaluator) VisitSomeExpression(node *ast.SomeExpression, ctx *ExecutionContext) Value {
    // Evaluate sub-expression
    value := e.VisitExpression(node.SubExpr, ctx)

    // Check for error
    if isError(value) {
        return value // Propagate error up
    }

    // Continue with logic
    // ...
}
```

### Type Checking

```go
// Check value type
intVal, ok := value.(*runtime.IntegerValue)
if !ok {
    return e.newError(node, "expected integer, got %s", value.Type())
}

// Use type assertions with helper functions
if !isInteger(value) {
    return e.newError(node, "expected integer")
}
```

### Delegating to Adapter

During the migration phase, some operations still need to delegate to the old interpreter:

```go
// Delegate to adapter for operations not yet migrated
if e.adapter != nil {
    return e.adapter.CallBuiltinFunction(name, args)
}
return e.newError(node, "function call not supported without adapter")
```

**Note**: The adapter will be removed in Phase 3.5.37.

## File Organization

The evaluator package is organized by responsibility:

```
internal/interp/evaluator/
├── evaluator.go              # Core evaluator, visitor infrastructure
├── visitor_expressions.go    # Expression visitor methods (48+)
├── visitor_statements.go     # Statement visitor methods
├── visitor_declarations.go   # Declaration visitor methods
├── visitor_literals.go       # Literal value creation
├── binary_ops.go             # Binary operation implementations
├── context.go                # ExecutionContext
├── callstack.go              # Call stack management
├── helpers.go                # Helper functions (isTruthy, isError, etc.)
├── env_adapter.go            # Environment adapter
├── benchmark_test.go         # Performance benchmarks
└── *_test.go                 # Unit tests
```

**Where to add your code**:
- **New expression**: `visitor_expressions.go`
- **New statement**: `visitor_statements.go`
- **New declaration**: `visitor_declarations.go`
- **New binary operator**: `binary_ops.go`
- **Helper function**: `helpers.go`
- **Tests**: Create `visitor_<category>_test.go`
- **Benchmarks**: Add to `benchmark_test.go`

## Testing Your Changes

### 1. Unit Tests

Test individual visitor methods in isolation:

```bash
# Run all evaluator tests
go test ./internal/interp/evaluator

# Run specific test
go test -run TestVisitBinaryExpression ./internal/interp/evaluator

# Run with verbose output
go test -v ./internal/interp/evaluator
```

### 2. Benchmarks

Measure performance of your changes:

```bash
# Run all benchmarks
go test -bench=. ./internal/interp/evaluator -run=^$

# Run specific benchmark
go test -bench=BenchmarkVisitBinaryExpression ./internal/interp/evaluator

# With memory profiling
go test -bench=. -benchmem ./internal/interp/evaluator
```

### 3. Integration Tests

Test with full interpreter:

```bash
# Run interpreter tests
go test ./internal/interp

# Run DWScript fixture tests
go test -v ./internal/interp -run TestDWScriptFixtures
```

## Performance Guidelines

### Performance Targets

Based on benchmarks (Task 3.5.35):

| Operation Type | Target | Current |
|---------------|--------|---------|
| Literals | < 5 ns/op | 0.3-0.6 ns ✅ |
| Binary ops | < 100 ns/op | ~70 ns ✅ |
| Unary ops | < 50 ns/op | 29-40 ns ✅ |
| String concat | < 200 ns/op | 156 ns ✅ |

### Optimization Tips

1. **Avoid unnecessary allocations**:
   ```go
   // Bad: Creates new error for every call
   func (e *Evaluator) validate(val Value) Value {
       if val == nil {
           return &ErrorValue{Message: "nil value"}
       }
       return val
   }

   // Good: Return existing value
   func (e *Evaluator) validate(val Value) Value {
       if val == nil {
           return e.newError(nil, "nil value")
       }
       return val
   }
   ```

2. **Use short-circuit evaluation**:
   ```go
   // Short-circuit boolean operators
   if node.Operator == "and" {
       if !isTruthy(left) {
           return left // Don't evaluate right
       }
   }
   ```

3. **Cache frequently used values**:
   ```go
   // Cache common values in evaluator
   e.trueValue = &runtime.BooleanValue{Value: true}
   e.falseValue = &runtime.BooleanValue{Value: false}
   ```

4. **Profile before optimizing**:
   ```bash
   # CPU profiling
   go test -bench=. -cpuprofile=cpu.prof ./internal/interp/evaluator
   go tool pprof cpu.prof

   # Memory profiling
   go test -bench=. -memprofile=mem.prof ./internal/interp/evaluator
   go tool pprof mem.prof
   ```

## Common Mistakes to Avoid

### 1. Forgetting Error Propagation

```go
// ❌ BAD: Doesn't check for errors
func (e *Evaluator) VisitBadExample(node *ast.BadExample, ctx *ExecutionContext) Value {
    value := e.VisitExpression(node.Expr, ctx)
    return doSomething(value) // Error from Expr is lost!
}

// ✅ GOOD: Checks and propagates errors
func (e *Evaluator) VisitGoodExample(node *ast.GoodExample, ctx *ExecutionContext) Value {
    value := e.VisitExpression(node.Expr, ctx)
    if isError(value) {
        return value // Propagate error
    }
    return doSomething(value)
}
```

### 2. Not Checking Control Flow

```go
// ❌ BAD: Ignores control flow in loops
func (e *Evaluator) VisitBadLoop(node *ast.BadLoop, ctx *ExecutionContext) Value {
    for {
        e.VisitStatement(node.Body, ctx)
        // Missing control flow checks!
    }
}

// ✅ GOOD: Checks control flow
func (e *Evaluator) VisitGoodLoop(node *ast.GoodLoop, ctx *ExecutionContext) Value {
    for {
        e.VisitStatement(node.Body, ctx)

        if ctx.GetControlFlow().ShouldBreak() {
            ctx.GetControlFlow().ClearBreak()
            break
        }
    }
    return nil
}
```

### 3. Modifying Shared State

```go
// ❌ BAD: Modifies evaluator state
func (e *Evaluator) VisitBadState(node *ast.BadState, ctx *ExecutionContext) Value {
    e.currentValue = someValue // Don't modify evaluator fields!
    return someValue
}

// ✅ GOOD: Uses context for state
func (e *Evaluator) VisitGoodState(node *ast.GoodState, ctx *ExecutionContext) Value {
    ctx.Env().Define("temp", someValue) // Use environment
    return someValue
}
```

### 4. Not Using defer for Cleanup

```go
// ❌ BAD: Forgets to pop environment on error
func (e *Evaluator) VisitBadScope(node *ast.BadScope, ctx *ExecutionContext) Value {
    ctx.PushEnv(newEnv)
    result := e.VisitStatement(node.Body, ctx)
    ctx.PopEnv() // Never called if error!
    return result
}

// ✅ GOOD: Uses defer for cleanup
func (e *Evaluator) VisitGoodScope(node *ast.GoodScope, ctx *ExecutionContext) Value {
    ctx.PushEnv(newEnv)
    defer ctx.PopEnv() // Always called
    return e.VisitStatement(node.Body, ctx)
}
```

## Migration Checklist

When migrating code from old Interpreter to new Evaluator:

- [ ] Create visitor method with correct signature
- [ ] Add case to dispatcher (VisitExpression/VisitStatement/etc.)
- [ ] Implement evaluation logic
- [ ] Add error handling for sub-evaluations
- [ ] Add control flow checking (for statements)
- [ ] Write unit tests
- [ ] Add benchmarks for hot paths
- [ ] Run existing tests to ensure no regressions
- [ ] Update documentation if behavior changes

## Getting Help

### Documentation

- [CLAUDE.md](../CLAUDE.md) - Project overview and guidelines
- [docs/architecture/interpreter.md](architecture/interpreter.md) - Detailed architecture
- [docs/evaluator-performance-report.md](evaluator-performance-report.md) - Performance analysis
- [PLAN.md](../PLAN.md) - Phase 3.5 task breakdown

### Code Examples

Look at existing visitor methods for patterns:

- **Simple expression**: `VisitIntegerLiteral` in `visitor_literals.go`
- **Binary operation**: `VisitBinaryExpression` in `visitor_expressions.go`
- **Control flow**: `VisitIfStatement` in `visitor_statements.go`
- **Loops**: `VisitForStatement` in `visitor_statements.go`
- **Scoping**: `VisitBlockStatement` in `visitor_statements.go`

### Tests as Examples

The test files show how to use the evaluator API:

- `visitor_literals_test.go` - Simple visitor tests
- `binary_ops_test.go` - Complex operation tests
- `context_test.go` - ExecutionContext usage
- `benchmark_test.go` - Performance testing

## Future Changes

### Phase 3.5.37: Adapter Removal

**Status**: ⏸️ Blocked on AST-free runtime types

When this phase completes:
- Remove all `e.adapter` calls
- Evaluator becomes fully self-contained
- All evaluation logic in visitor methods
- No more delegation to old Interpreter

**What this means for you**:
- Code using adapter will need migration
- All visitor methods must be self-sufficient
- Type system must not depend on AST nodes

---

**Questions?** Check the documentation or look at existing visitor methods for examples.

**Found a bug?** File an issue with details of the problem and a minimal reproduction case.

**Want to contribute?** Follow this guide and submit a PR!
