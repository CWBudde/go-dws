# Evaluator Architecture: Separation of Concerns

**Created**: 2025-12-04
**Task**: 3.5.43 - Documenting architectural boundaries
**Purpose**: Define evaluator vs interpreter responsibilities and explain why the adapter interface is essential

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architecture Overview](#architecture-overview)
3. [Evaluator Responsibilities](#evaluator-responsibilities)
4. [Interpreter Responsibilities](#interpreter-responsibilities)
5. [The Adapter Interface](#the-adapter-interface)
6. [The Two Essential EvalNode Calls](#the-two-essential-evalnode-calls)
7. [Migration Pattern: When to Migrate vs Preserve](#migration-pattern-when-to-migrate-vs-preserve)
8. [Comparison with Reference Counting Migration](#comparison-with-reference-counting-migration)
9. [Alternative Approaches Considered](#alternative-approaches-considered)
10. [Conclusion](#conclusion)

---

## Executive Summary

The evaluator and interpreter have **distinct, complementary responsibilities**:

- **Evaluator**: Executes AST nodes (expressions, statements, declarations) using runtime values
- **Interpreter**: Manages OOP semantics (classes, objects, Self context, method dispatch)

The adapter interface provides a **clean architectural boundary** between these concerns. While Phase 3.5 successfully eliminated 95%+ of adapter usage, **2 EvalNode calls remain essential** for Self/class context handling.

**Key Decision** (Task 3.5.43): Keep these 2 calls as architectural boundaries rather than forcing evaluator to handle OOP semantics.

---

## Architecture Overview

### Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                          Interpreter                            │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ OOP Semantics Layer                                        │ │
│  │ - ClassInfo registry                                       │ │
│  │ - ObjectInstance lifecycle                                 │ │
│  │ - Self context management (env.Define("Self", obj))        │ │
│  │ - Field/property/class var lookup                          │ │
│  │ - Method dispatch & virtual method tables                  │ │
│  │ - Constructor/destructor invocation                        │ │
│  └────────────────────────────────────────────────────────────┘ │
│                              ▲                                  │
│                              │                                  │
│                              │ Adapter Interface                │
│                              │ (2 essential EvalNode calls)     │
│                              │                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ Evaluator                                                  │ │
│  │ - AST node evaluation (visitor pattern)                    │ │
│  │ - Expression/statement execution                           │ │
│  │ - Environment-based variable resolution                    │ │
│  │ - Control flow (break, continue, exit, return)             │ │
│  │ - Binary/unary operations                                  │ │
│  │ - Array/record access                                      │ │
│  └────────────────────────────────────────────────────────────┘ │
│                              ▲                                  │
│                              │                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ Runtime Package                                            │ │
│  │ - Value types (Integer, Float, String, Boolean, etc.)      │ │
│  │ - ObjectInstance, InterfaceInstance                        │ │
│  │ - RefCountManager (lifecycle management)                   │ │
│  │ - Environment (variable storage)                           │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### Separation Principle

**Clean Architecture**: Each layer depends only on layers below it, with well-defined interfaces.

- **Runtime**: Pure value types, no AST dependencies, no interpreter dependencies
- **Evaluator**: Depends on runtime, operates on AST nodes, uses ExecutionContext
- **Interpreter**: Depends on evaluator + runtime, owns OOP semantics and global state

---

## Evaluator Responsibilities

The evaluator's job is to **execute AST nodes and produce runtime values**.

### Core Duties

1. **AST Node Evaluation** (48+ visitor methods)
   - Expressions: Binary/unary ops, literals, identifiers, member access, array indexing
   - Statements: Assignments, if/else, loops (for/while/repeat), try/except
   - Declarations: Variables, functions, classes, records (via callbacks)

2. **Environment-Based Variable Resolution**
   - `ctx.Env().Get(name)` - Retrieve variable value from environment
   - `ctx.Env().Set(name, value)` - Update existing variable
   - `ctx.Env().Define(name, value)` - Declare new variable
   - Handles 99% of variable access through environment scope chain

3. **Control Flow Management**
   - Break/continue signals (loop control)
   - Exit signals (program termination)
   - Return signals (function return)
   - Exception handling (try/except/finally)

4. **Value Operations**
   - Arithmetic: `+`, `-`, `*`, `/`, `div`, `mod`
   - Comparison: `=`, `<>`, `<`, `>`, `<=`, `>=`
   - Boolean: `and`, `or`, `xor`, `not`
   - String: Concatenation, comparison

5. **Type Conversions**
   - Implicit: Integer → Float, String ↔ other types
   - Explicit: Type casts, range checking

### What Evaluator Does NOT Handle

- ❌ Class/object creation (ObjectInstance allocation)
- ❌ Method dispatch (virtual method table lookup)
- ❌ Self context management (binding Self to environment)
- ❌ Field/property/class var lookup (requires ClassInfo access)
- ❌ Constructor/destructor invocation (lifecycle events)

These are **interpreter responsibilities**.

---

## Interpreter Responsibilities

The interpreter owns **OOP semantics and global state management**.

### Core Duties

1. **Class Registry**
   - Maintains map of all classes: `map[string]*ClassInfo`
   - Class hierarchy (parent/child relationships)
   - Class metadata (fields, methods, properties, class vars)

2. **Object Lifecycle**
   - Constructor invocation: `TMyClass.Create(args)`
   - Object allocation: `NewObjectInstance(class)`
   - Destructor invocation: `callDestructorIfNeeded(obj)`
   - Reference counting coordination (via RefCountManager)

3. **Self Context Management**
   - **Instance methods**: `env.Define("Self", obj)` before executing method body
   - **Class methods**: `env.Define("__CurrentClass__", classInfo)` before executing
   - **Cleanup**: `env.Undefine("Self")` after method execution

4. **Field/Property/Class Var Access**
   - **Field lookup**: `obj.Class.LookupField(name)` walks inheritance chain
   - **Property access**: Resolves getter/setter methods, evaluates expressions
   - **Class var access**: `classInfo.ClassVars[name]` with hierarchy search

5. **Method Dispatch**
   - **Virtual dispatch**: Resolves overridden methods via class hierarchy
   - **Inherited calls**: `inherited MethodName` walks parent chain
   - **Method pointers**: Creates function pointers with bound Self

### Why Interpreter Owns OOP Semantics

**Architectural Reason**: OOP semantics require **global knowledge** that evaluator shouldn't have:

- Class hierarchies span entire program
- Method dispatch requires virtual method tables
- Self context is stateful (changes per method invocation)
- Property evaluation can execute arbitrary code

**Separating these concerns keeps evaluator simple and testable.**

---

## The Adapter Interface

### Purpose

The adapter provides a **callback boundary** between evaluator and interpreter for cases where:

1. Evaluator needs interpreter-owned OOP logic
2. Circular imports must be avoided (evaluator can't import interpreter)
3. Clean separation of concerns is desired

### Essential Methods (Post-Phase 3.5)

After Phase 3.5 migration, the adapter interface has **~5 essential methods**:

```go
type InterpreterAdapter interface {
    // EvalNode evaluates any AST node using the interpreter's full context.
    // Used for: Self/class context resolution (2 calls in assignment_helpers.go)
    EvalNode(node ast.Node) Value

    // ExecuteUserFunctionBody executes a user-defined function body.
    // Used for: Function calls, method invocation
    ExecuteUserFunctionBody(funcDecl *ast.FunctionDecl, args []Value, ctx *ExecutionContext) Value

    // Additional methods for class/object operations (~3 methods)
    // - Constructor invocation
    // - Destructor callbacks
    // - Class registration
}
```

### Adapter is NOT Technical Debt

**Common Misconception**: "Adapter interface is temporary, should be eliminated"

**Reality**: Adapter is a **permanent architectural feature** that provides clean separation:

- Evaluator operates on **local context** (ExecutionContext, Environment)
- Interpreter operates on **global state** (class registry, Self binding)
- Adapter bridges these concerns **without violating separation**

**Precedent**: Similar to how RefCountManager provides callback interface for destructors.

---

## The Two Essential EvalNode Calls

### Location

**File**: `internal/interp/evaluator/assignment_helpers.go`

**Lines**: 93 (simple assignment), 368 (compound assignment)

### The Problem

When the evaluator encounters `x := value` or `x += value`:

1. **First check**: `ctx.Env().Get("x")` - Is `x` in the environment?
2. **If found**: Normal variable assignment (99% of cases) ✅
3. **If NOT found**: Could be one of:
   - Instance field: `Self.Field := value` (implicit Self in method)
   - Class variable: `TClass.ClassVar := value` (static variable)
   - Property: `Self.PropName := value` (property setter)

**The evaluator cannot distinguish these cases** because:

- It doesn't have access to `ClassInfo.ClassVars` map
- It doesn't have `obj.Class.LookupField()` method
- It doesn't have `evalPropertyWrite()` logic
- All of these live in interpreter package

### Code Example

**Line 93** - Simple assignment to non-env variable:

```go
func (e *Evaluator) evalSimpleAssignmentDirect(
    stmt *ast.AssignmentStatement,
    target *ast.Identifier,
    value Value,
    ctx *ExecutionContext,
) Value {
    targetName := target.Value

    // Get existing value to check for special types
    existingValRaw, exists := ctx.Env().Get(targetName)
    if !exists {
        // ARCHITECTURAL BOUNDARY: Self/class context is owned by interpreter.
        // The evaluator operates on the environment (ctx.Env()) and cannot access:
        // - Instance fields (Self.Field) without interpreter's field lookup logic
        // - Class variables (TClass.ClassVar) without interpreter's ClassInfo access
        // - Property setters without interpreter's property execution logic
        // This delegation is intentional and represents a clean separation of concerns.
        return e.adapter.EvalNode(stmt)
    }

    // ... rest of assignment logic for environment variables
}
```

**Line 368** - Compound assignment to non-env variable:

```go
func (e *Evaluator) evalCompoundIdentifierAssignment(
    stmt *ast.AssignmentStatement,
    target *ast.Identifier,
    ctx *ExecutionContext,
) Value {
    targetName := target.Value

    // Get current value from environment
    currentValRaw, exists := ctx.Env().Get(targetName)
    if !exists {
        // ARCHITECTURAL BOUNDARY: Self/class context is owned by interpreter.
        // (Same rationale as line 93)
        return e.adapter.EvalNode(stmt)
    }

    // ... rest of compound assignment logic for environment variables
}
```

### Why These Calls Remain

**Option 1** (Callbacks): Just shifts the adapter call to `SetVar()` - doesn't eliminate it

**Option 2** (Explicit Context): Requires evaluator to duplicate interpreter logic:
- Add `ctx.selfObject` and `ctx.currentClass` fields
- Implement field/property/class var lookup in evaluator
- Create `SetClassVar()` method on ObjectValue interface
- Violates separation of concerns (evaluator shouldn't know OOP internals)

**Option 3** (Keep as Essential): Accept architectural boundary
- Acknowledges that Self/class context is **OOP semantics** (interpreter concern)
- Preserves clean separation (evaluator evaluates, interpreter manages OOP)
- Minimal effort (2-3h documentation vs 8-12h risky code changes)

**Decision**: Task 3.5.43 chose **Option 3** after thorough analysis.

---

## Migration Pattern: When to Migrate vs Preserve

Phase 3.5 successfully migrated **~50+ adapter methods** while preserving **2 essential calls**. The pattern:

### ✅ Migrate When: Value Lifecycle Concerns

**Example**: Reference counting (tasks 3.5.39-3.5.42)

**Characteristics**:
- Operates on **runtime values** (ObjectInstance, InterfaceInstance)
- Doesn't require **class hierarchy knowledge**
- Can be abstracted via **clean interface** (RefCountManager)
- Affects **value lifetime**, not OOP semantics

**Migration Strategy**:
1. Design interface in runtime package (`RefCountManager`)
2. Move logic from interpreter to runtime implementation
3. Pass interface to evaluator via ExecutionContext
4. Replace adapter calls with direct interface calls

**Result**: 7 EvalNode calls eliminated ✅

### ⛔ Preserve When: OOP Semantics Concerns

**Example**: Self/class context (task 3.5.43)

**Characteristics**:
- Requires **global knowledge** (class registry, hierarchy)
- Involves **stateful management** (Self binding/unbinding)
- Needs **interpreter methods** (evalPropertyWrite, lookupField)
- Represents **architectural boundary** between concerns

**Preservation Strategy**:
1. Document why calls remain (architectural rationale)
2. Update comments to reference documentation
3. Accept adapter as **essential** (not technical debt)

**Result**: 2 EvalNode calls preserved as architectural boundary ✅

### Decision Matrix

| Concern Type | Evaluator Responsibility? | Migration Action |
|--------------|---------------------------|------------------|
| Value operations | ✅ Yes | Migrate to evaluator |
| Type conversions | ✅ Yes | Migrate to evaluator |
| Control flow | ✅ Yes | Migrate to evaluator |
| Reference counting | ✅ Yes (value lifecycle) | Migrate via interface |
| Property callbacks | ✅ Yes (with callback) | Migrate via callback pattern |
| Self/class context | ❌ No (OOP semantics) | Keep in interpreter |
| Method dispatch | ❌ No (requires VMT) | Keep in interpreter |
| Class registration | ❌ No (global state) | Keep in interpreter |

---

## Comparison with Reference Counting Migration

The successful **ref counting migration** (3.5.39-3.5.42) provides a **template** for when to migrate:

### Reference Counting (Migrated) ✅

**Nature**: Value lifecycle management

**Requirements**:
- `IncrementRef(obj)` - Increment reference count
- `DecrementRef(obj)` - Decrement reference count, invoke destructor if 0
- `ReleaseObject(obj)` - Decrement and optionally invoke destructor
- `SetDestructorCallback(func)` - Register destructor callback

**Implementation**:
- Interface in `runtime/refcount.go`
- No access to ClassInfo or class hierarchy needed
- Destructor callback avoids circular import
- Clean separation: runtime manages value lifecycle

**Result**: 7 adapter calls eliminated, 0 violations of separation of concerns

### Self Context (Preserved) ✅

**Nature**: OOP semantics management

**Requirements**:
- `env.Get("Self")` → extract ObjectInstance
- `obj.Class.LookupField(name)` → walk inheritance chain
- `obj.Class.LookupClassVar(name)` → search hierarchy
- `evalPropertyWrite(obj, prop, value)` → execute setter logic

**Why NOT Migratable**:
- Requires access to `ClassInfo.ClassVars` map (interpreter package)
- Requires `obj.Class.LookupField()` (interpreter method)
- Requires `evalPropertyWrite()` (interpreter method)
- Would duplicate interpreter logic into evaluator

**Result**: 2 adapter calls preserved, clean separation maintained

### The Key Difference

| Aspect | Reference Counting | Self Context |
|--------|-------------------|--------------|
| Abstraction Level | Value lifecycle | OOP semantics |
| Scope | Local (per value) | Global (class registry) |
| Dependencies | None (self-contained) | ClassInfo, hierarchy, methods |
| Interface Complexity | Simple (6 methods) | Complex (would need 20+ methods) |
| Separation Violation? | No | Yes (if migrated) |

**Lesson**: Migrate when abstraction is **clean and self-contained**. Preserve when migration would **violate architectural boundaries**.

---

## Alternative Approaches Considered

### Option 1: Callback Pattern (Like RefCountManager)

**Idea**: Add `TrySetSelfOrClassVar()` callback to adapter

```go
func (e *Evaluator) SetVar(ctx *ExecutionContext, name string, value Value) bool {
    if ctx.Env().Set(name, value) {
        return true
    }
    // Callback to interpreter for Self/class scope
    if e.adapter != nil {
        return e.adapter.TrySetSelfOrClassVar(name, value)
    }
    return false
}
```

**Analysis**:
- ✅ Minimal changes
- ❌ Doesn't eliminate adapter (just shifts location)
- ❌ Doesn't achieve stated goal

**Verdict**: Rejected (doesn't solve the problem)

### Option 2: Explicit Self Context in ExecutionContext

**Idea**: Add `selfObject` and `currentClass` fields to ExecutionContext

```go
type ExecutionContext struct {
    // ... existing fields
    selfObject   Value
    currentClass Value
}

func (e *Evaluator) SetVar(ctx *ExecutionContext, name string, value Value) bool {
    // 1. Try environment
    if ctx.Env().Set(name, value) {
        return true
    }

    // 2. Try Self context
    if ctx.selfObject != nil {
        if obj, ok := ctx.selfObject.(ObjectValue); ok {
            // Check field, class var, property
            // ← PROBLEM: Needs ClassInfo.ClassVars access
        }
    }

    return false
}
```

**Problems Discovered**:

1. **Missing interface methods**:
   - `ObjectValue.SetClassVar(name, value)` doesn't exist
   - `ClassMetaValue` interface doesn't exist

2. **Requires duplicating interpreter logic**:
   - Field lookup: `runtime.LookupFieldInHierarchy(obj.Class.GetMetadata(), name)`
   - Class var lookup: `classInfo.ClassVars[name]`
   - Property write: `evalPropertyWrite(obj, propInfo, value)`

3. **Circular import risk**:
   - Evaluator would need access to `ClassInfo.ClassVars`
   - Evaluator would need `obj.Class.LookupField()` method
   - Both live in interpreter package

**Analysis**:
- ✅ Eliminates 2 adapter calls
- ❌ Violates separation of concerns
- ❌ Duplicates interpreter logic
- ❌ High complexity (8-12 hours)
- ❌ High bug risk

**Verdict**: Rejected (architecturally problematic)

### Option 3: Keep as Essential (CHOSEN)

**Idea**: Accept 2 EvalNode calls as architectural boundaries

**Analysis**:
- ✅ Preserves separation of concerns
- ✅ Minimal effort (2-3h documentation)
- ✅ No risk of bugs
- ✅ Clean architecture
- ❌ Adapter interface remains (but minimal)

**Verdict**: **ACCEPTED** (task 3.5.43)

---

## Conclusion

### Phase 3.5 Achievements

**Evaluator Independence**: ✅ Achieved

- All declaration evaluation migrated
- All value types migrated to runtime
- Reference counting migrated
- Property callbacks implemented
- 95%+ adapter reduction (75 → ~5 essential methods)

**Architectural Integrity**: ✅ Preserved

- Evaluator evaluates AST nodes
- Interpreter manages OOP semantics
- Adapter provides clean boundary
- Separation of concerns maintained

### The Two Essential Calls

**Lines 93, 368** in `assignment_helpers.go` are **not technical debt** - they represent:

1. **Architectural boundary** between evaluator (AST execution) and interpreter (OOP semantics)
2. **Clean separation** of concerns (evaluator shouldn't know ClassInfo internals)
3. **Pragmatic engineering** (2-3h documentation vs 8-12h risky migration)

### Lessons Learned

**When to Migrate**:
- Value lifecycle concerns (reference counting)
- Self-contained abstractions (RefCountManager)
- No violation of separation of concerns

**When to Preserve**:
- OOP semantics (Self context, method dispatch)
- Global state management (class registry)
- Would violate architectural boundaries

### Final Architecture

```
Evaluator (Independent)
    ↓ Operates on: AST nodes, runtime values, ExecutionContext
    ↓ Handles: 99% of expressions, statements, assignments
    ↓
Adapter (Essential Boundary)
    ↓ Provides: 2 EvalNode calls for Self/class context
    ↓ Provides: User function execution
    ↓
Interpreter (OOP Semantics)
    ↓ Owns: ClassInfo registry, Self management, method dispatch
    ↓ Handles: 1% of cases requiring global OOP knowledge
```

**Result**: Clean, maintainable architecture with clear responsibilities.

---

## References

- [docs/evalnode-audit-final.md](evalnode-audit-final.md) - Complete audit of all EvalNode calls
- [docs/refcounting-design.md](refcounting-design.md) - Reference counting migration pattern
- [PLAN.md](../PLAN.md) - Task 3.5.43 decision and rationale
- [internal/interp/evaluator/assignment_helpers.go](../internal/interp/evaluator/assignment_helpers.go) - The 2 essential calls
- [internal/interp/statements_assignments.go](../internal/interp/statements_assignments.go) - Interpreter's Self context logic
