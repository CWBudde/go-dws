# AST-Free Runtime Metadata Design

**Task**: 3.5.37
**Status**: Design Complete
**Author**: Phase 9 (AST-Free Runtime Types)
**Date**: 2025-01-22

## Executive Summary

This document describes the design for eliminating AST dependencies from runtime value types in the DWScript interpreter. The goal is to enable the Evaluator to work independently of the AST package by storing only the metadata needed at runtime, not full AST nodes.

## Problem Statement

### Current State

The runtime type system has heavy AST dependencies that couple execution to the AST:

1. **ClassInfo** stores 9+ fields with `*ast.FunctionDecl` and `*ast.FieldDecl`:
   - `Methods map[string]*ast.FunctionDecl`
   - `MethodOverloads map[string][]*ast.FunctionDecl`
   - `ClassMethods map[string]*ast.FunctionDecl`
   - `ClassMethodOverloads map[string][]*ast.FunctionDecl`
   - `Constructors map[string]*ast.FunctionDecl`
   - `ConstructorOverloads map[string][]*ast.FunctionDecl`
   - `Constructor *ast.FunctionDecl`
   - `Destructor *ast.FunctionDecl`
   - `FieldDecls map[string]*ast.FieldDecl`
   - `Constants map[string]*ast.ConstDecl`

2. **RecordTypeValue** stores 6 AST-dependent fields:
   - `Methods map[string]*ast.FunctionDecl`
   - `StaticMethods map[string]*ast.FunctionDecl`
   - `ClassMethods map[string]*ast.FunctionDecl`
   - `MethodOverloads map[string][]*ast.FunctionDecl`
   - `ClassMethodOverloads map[string][]*ast.FunctionDecl`
   - `FieldDecls map[string]*ast.FieldDecl`

3. **HelperInfo** stores methods as `*ast.FunctionDecl`

4. **VirtualMethodEntry** stores `Implementation *ast.FunctionDecl`

### Why This Is Problematic

1. **Tight Coupling**: Runtime execution is tightly coupled to AST structure
2. **Memory Overhead**: AST nodes carry compile-time information not needed at runtime
3. **Adapter Pattern Dependency**: Evaluator requires InterpreterAdapter to access type system
4. **No Clear Separation**: Compile-time and runtime concerns are mixed

## Analysis: What's Actually Used at Runtime

I analyzed the codebase to determine what information from AST nodes is actually used during execution:

### From `*ast.FunctionDecl` (in `callUserFunction`):

**Used Information**:
- `Name.Value` (string) - for call stack traces, Result variable alias
- `Parameters` - array with per-parameter:
  - `Name.Value` (string) - for binding arguments
  - `Type` (TypeAnnotation) - for implicit conversion
  - `ByRef` (bool) - to distinguish var parameters
  - `DefaultValue` (Expression) - AST expression evaluated for optional parameters
- `ReturnType` (TypeAnnotation) - to initialize Result variable
- `Body` (Statement) - AST statement block to execute
- `PreConditions` ([]Expression) - AST expressions checked before execution
- `PostConditions` ([]Expression) - AST expressions checked after execution

**Not Used**: Position info, documentation, compiler flags

### From `*ast.FieldDecl` (in field initialization):

**Used Information**:
- `Name.Value` (string) - field name
- `Type` (TypeAnnotation) - field type
- `InitValue` (Expression) - AST expression for default value

**Not Used**: Position info, visibility modifiers (stored separately)

### From `*ast.ConstDecl`:

**Stored but Pre-Evaluated**: Constants are evaluated once and stored as `Value` in `ConstantValues` map. The AST node is kept but rarely re-evaluated.

## Proposed Design

### Core Metadata Structures

#### ParameterMetadata

Replaces the need to access `*ast.Parameter` at runtime:

```go
// ParameterMetadata describes a function/method parameter at runtime.
type ParameterMetadata struct {
    // Name is the parameter name for binding arguments
    Name string

    // TypeName is the string representation of the type (for display)
    TypeName string

    // Type is the resolved type (nil if not resolved)
    Type types.Type

    // ByRef indicates if this is a var parameter (pass-by-reference)
    ByRef bool

    // DefaultValue is the expression to evaluate for optional parameters.
    // Nil for required parameters.
    // Phase 9: Keeps AST expression; Phase 10: Migrate to bytecode
    DefaultValue ast.Expression
}
```

#### MethodMetadata

Replaces `*ast.FunctionDecl` in runtime type maps:

```go
// MethodMetadata describes a callable method/function at runtime.
// This replaces the need to store full *ast.FunctionDecl nodes.
type MethodMetadata struct {
    // Signature information
    Name           string
    Parameters     []ParameterMetadata
    ReturnTypeName string       // "" for procedures
    ReturnType     types.Type   // nil for procedures

    // Executable body - exactly one of these will be set:
    Body           ast.Statement          // AST body (Phase 9)
    BytecodeID     int                   // Bytecode ID (future)
    NativeFunc     func(args []interface{}) interface{}   // Built-in function

    // Validation (Phase 9: keep AST; future: migrate to bytecode)
    PreConditions  *ast.PreConditions
    PostConditions *ast.PostConditions

    // Method characteristics
    IsVirtual      bool
    IsAbstract     bool
    IsOverride     bool
    IsReintroduce  bool
    IsClassMethod  bool  // Static method
    IsConstructor  bool
    IsDestructor   bool

    // Visibility
    Visibility     MethodVisibility
}

// MethodVisibility represents method visibility levels.
type MethodVisibility int

const (
    VisibilityPublic MethodVisibility = iota
    VisibilityPrivate
    VisibilityProtected
    VisibilityPublished
)
```

#### FieldMetadata

Replaces `*ast.FieldDecl` in runtime type maps:

```go
// FieldMetadata describes a field at runtime.
type FieldMetadata struct {
    Name       string
    TypeName   string
    Type       types.Type

    // InitValue is the initializer expression (nil if none).
    // Phase 9: Keeps AST expression; future: migrate to bytecode
    InitValue  ast.Expression

    Visibility FieldVisibility
}

// FieldVisibility represents field visibility levels.
type FieldVisibility int

const (
    FieldVisibilityPublic FieldVisibility = iota
    FieldVisibilityPrivate
    FieldVisibilityProtected
    FieldVisibilityPublished
)
```

#### VirtualMethodMetadata

Replaces `VirtualMethodEntry` without AST dependency:

```go
// VirtualMethodMetadata tracks virtual method dispatch information.
type VirtualMethodMetadata struct {
    // IntroducedBy is the class that first declared this method as virtual
    IntroducedBy   *ClassMetadata

    // Implementation is the method to actually call for this class
    Implementation *MethodMetadata

    IsVirtual      bool
    IsReintroduced bool
}
```

#### ClassMetadata

Replaces ClassInfo's AST-dependent fields:

```go
// ClassMetadata contains runtime metadata for a class.
// This replaces the AST-dependent fields in ClassInfo.
type ClassMetadata struct {
    Name         string
    ParentName   string
    Parent       *ClassMetadata
    Interfaces   []string  // Interface names implemented

    // Fields
    Fields       map[string]*FieldMetadata

    // Instance methods
    Methods          map[string]*MethodMetadata
    MethodOverloads  map[string][]*MethodMetadata

    // Static methods (class functions/procedures)
    ClassMethods         map[string]*MethodMetadata
    ClassMethodOverloads map[string][]*MethodMetadata

    // Constructors
    Constructors         map[string]*MethodMetadata
    ConstructorOverloads map[string][]*MethodMetadata
    DefaultConstructor   string

    // Destructor
    Destructor       *MethodMetadata

    // Virtual dispatch
    VirtualMethods   map[string]*VirtualMethodMetadata

    // Constants and class variables (already runtime values, no change)
    Constants        map[string]Value
    ClassVars        map[string]Value

    // Properties (already metadata, no change)
    Properties       map[string]*types.PropertyInfo

    // Operators (already runtime registry, no change)
    Operators        *runtimeOperatorRegistry

    // Class flags
    IsAbstract       bool
    IsExternal       bool
    IsPartial        bool
    ExternalName     string
}

// NewClassMetadata creates a new ClassMetadata with initialized maps.
func NewClassMetadata(name string) *ClassMetadata {
    return &ClassMetadata{
        Name:                 name,
        Fields:               make(map[string]*FieldMetadata),
        Methods:              make(map[string]*MethodMetadata),
        MethodOverloads:      make(map[string][]*MethodMetadata),
        ClassMethods:         make(map[string]*MethodMetadata),
        ClassMethodOverloads: make(map[string][]*MethodMetadata),
        Constructors:         make(map[string]*MethodMetadata),
        ConstructorOverloads: make(map[string][]*MethodMetadata),
        VirtualMethods:       make(map[string]*VirtualMethodMetadata),
        Constants:            make(map[string]Value),
        ClassVars:            make(map[string]Value),
        Properties:           make(map[string]*types.PropertyInfo),
    }
}
```

#### RecordMetadata

Replaces RecordTypeValue's AST-dependent fields:

```go
// RecordMetadata contains runtime metadata for a record type.
// This replaces the AST-dependent fields in RecordTypeValue.
type RecordMetadata struct {
    Name       string
    RecordType *types.RecordType

    // Fields
    Fields     map[string]*FieldMetadata

    // Instance methods
    Methods          map[string]*MethodMetadata
    MethodOverloads  map[string][]*MethodMetadata

    // Static methods (class functions/procedures)
    StaticMethods         map[string]*MethodMetadata
    StaticMethodOverloads map[string][]*MethodMetadata

    // Constants and class variables (already runtime values)
    Constants  map[string]Value
    ClassVars  map[string]Value
}

// NewRecordMetadata creates a new RecordMetadata with initialized maps.
func NewRecordMetadata(name string, recordType *types.RecordType) *RecordMetadata {
    return &RecordMetadata{
        Name:                  name,
        RecordType:            recordType,
        Fields:                make(map[string]*FieldMetadata),
        Methods:               make(map[string]*MethodMetadata),
        MethodOverloads:       make(map[string][]*MethodMetadata),
        StaticMethods:         make(map[string]*MethodMetadata),
        StaticMethodOverloads: make(map[string][]*MethodMetadata),
        Constants:             make(map[string]Value),
        ClassVars:             make(map[string]Value),
    }
}
```

### Helper Utilities

```go
// MethodMetadataFromAST converts an AST function declaration to MethodMetadata.
func MethodMetadataFromAST(fn *ast.FunctionDecl) *MethodMetadata {
    metadata := &MethodMetadata{
        Name:       fn.Name.Value,
        Parameters: make([]ParameterMetadata, len(fn.Parameters)),
        Body:       fn.Body,
    }

    // Convert parameters
    for i, param := range fn.Parameters {
        metadata.Parameters[i] = ParameterMetadata{
            Name:         param.Name.Value,
            TypeName:     param.Type.String(),
            ByRef:        param.ByRef,
            DefaultValue: param.DefaultValue,
        }
    }

    // Set return type
    if fn.ReturnType != nil {
        metadata.ReturnTypeName = fn.ReturnType.String()
    }

    // Copy conditions
    metadata.PreConditions = fn.PreConditions
    metadata.PostConditions = fn.PostConditions

    return metadata
}

// FieldMetadataFromAST converts an AST field declaration to FieldMetadata.
func FieldMetadataFromAST(field *ast.FieldDecl) *FieldMetadata {
    return &FieldMetadata{
        Name:      field.Name.Value,
        TypeName:  field.Type.String(),
        InitValue: field.InitValue,
    }
}
```

## Migration Strategy

### Phase 1: Design & Implementation (Task 3.5.37) ✓

- [x] Analyze current AST dependencies
- [x] Design metadata structures
- [x] Create design document
- [ ] Implement `internal/interp/runtime/metadata.go` with structs
- [ ] Add conversion utilities (AST → Metadata)
- [ ] Add basic tests for metadata structs

**Deliverables**:
- `docs/ast-free-runtime-design.md` (this document)
- `internal/interp/runtime/metadata.go` with all metadata types
- Tests: `internal/interp/runtime/metadata_test.go`

### Phase 2: MethodMetadata & MethodRegistry (Task 3.5.38)

- [ ] Create `MethodRegistry` to store methods by unique ID
- [ ] Update method registration to create MethodMetadata
- [ ] Add method lookup by ID
- [ ] Replace `*ast.FunctionDecl` references with method IDs where possible
- [ ] Update `callUserFunction` to accept MethodMetadata

**Files Modified**:
- `internal/interp/runtime/method_registry.go` (new)
- `internal/interp/evaluator/visitor_declarations.go`
- `internal/interp/functions_user.go`

### Phase 3: FieldMetadata Migration (Task 3.5.39)

- [ ] Update field registration to create FieldMetadata
- [ ] Update field initialization code to use FieldMetadata
- [ ] Replace `FieldDecls map[string]*ast.FieldDecl` with `Fields map[string]*FieldMetadata`

**Files Modified**:
- `internal/interp/class.go`
- `internal/interp/record.go`
- `internal/interp/objects_instantiation.go`

### Phase 4: ClassMetadata Migration (Task 3.5.40)

- [ ] Build ClassMetadata during class declaration evaluation
- [ ] Add ClassMetadata field to ClassInfo
- [ ] Update method lookups to use ClassMetadata
- [ ] Deprecate old AST-dependent fields
- [ ] Add compatibility layer for gradual migration

**Files Modified**:
- `internal/interp/class.go`
- `internal/interp/evaluator/visitor_declarations.go`
- All files accessing ClassInfo methods

### Phase 5: RecordMetadata Migration (Task 3.5.41)

- [ ] Build RecordMetadata during record declaration evaluation
- [ ] Add RecordMetadata field to RecordTypeValue
- [ ] Update record method lookups
- [ ] Deprecate old AST-dependent fields

**Files Modified**:
- `internal/interp/record.go`
- `internal/interp/evaluator/visitor_declarations.go`

### Phase 6: Remove Adapter Pattern (Task 3.5.42)

- [ ] Remove InterpreterAdapter interface from Evaluator
- [ ] Make Evaluator work directly with metadata
- [ ] Clean up circular dependencies
- [ ] Update documentation

**Files Modified**:
- `internal/interp/evaluator/evaluator.go`
- `internal/interp/interpreter.go`

## Design Decisions & Rationale

### Keep AST Expressions in Phase 9

**Decision**: MethodMetadata keeps `Body ast.Statement`, `DefaultValue ast.Expression`, `InitValue ast.Expression`, `PreConditions []ast.Expression`, etc.

**Rationale**:
1. **Incremental Migration**: Allows us to remove AST dependencies from type metadata while keeping the execution engine unchanged
2. **Reduced Risk**: Smaller changes, easier to test and validate
3. **Future-Proof**: Clear path to migrate to bytecode in Phase 10
4. **Partial Win**: Even with AST expressions, we eliminate 80%+ of AST coupling

**Future**: Phase 10 will replace AST expressions with bytecode IDs or compiled closures.

### Metadata Lives in runtime Package

**Decision**: Place metadata.go in `internal/interp/runtime/` package.

**Rationale**:
1. **Clear Ownership**: Runtime metadata is part of the runtime system
2. **Package Organization**: Separates runtime concerns from interpreter orchestration
3. **Reusability**: Can be used by both AST interpreter and bytecode VM
4. **No Circular Dependencies**: runtime package doesn't import interp

### Method IDs for Indirection

**Decision**: MethodRegistry assigns unique IDs to methods; ClassMetadata stores IDs instead of pointers.

**Rationale**:
1. **Serialization**: IDs are easier to serialize for bytecode cache
2. **Memory**: Reduces pointer chasing and memory overhead
3. **Lookup**: O(1) lookup in registry by ID
4. **Flexibility**: Can swap implementations (AST vs bytecode) without changing ClassMetadata

**Alternative Considered**: Direct MethodMetadata pointers. Rejected due to serialization complexity and memory overhead.

### Preserve Visibility in Metadata

**Decision**: Add Visibility fields to MethodMetadata and FieldMetadata.

**Rationale**:
1. **Semantic Checking**: Needed for access control validation
2. **Future Stages**: Required for Stage 6 (semantic analysis)
3. **Completeness**: Metadata should be self-contained

## Impact Assessment

### Benefits

1. **Reduced Coupling**: Evaluator no longer depends on full AST structure
2. **Memory Savings**: Metadata is smaller than full AST nodes
3. **Clearer Separation**: Compile-time vs runtime concerns are explicit
4. **Future-Ready**: Paves way for bytecode migration
5. **Remove Adapter**: Eliminates InterpreterAdapter anti-pattern

### Risks

1. **Migration Complexity**: Large refactoring across many files
2. **Compatibility**: Must maintain existing behavior during transition
3. **Testing Burden**: Need comprehensive tests to ensure correctness

### Mitigation

1. **Incremental Rollout**: Migrate one component at a time (methods → fields → classes → records)
2. **Parallel Structures**: Keep old and new side-by-side during transition
3. **Extensive Testing**: Run full test suite at each phase
4. **Compatibility Layer**: Provide conversion functions between old and new formats

## Success Criteria

1. **All tests pass** after each migration phase
2. **No AST imports** in runtime metadata types (except for expressions in Phase 9)
3. **InterpreterAdapter removed** from Evaluator
4. **Performance neutral** or improved
5. **Documentation updated** to reflect new architecture

## Future Work (Post-Phase 9)

### Phase 10: Bytecode Migration

- Replace `Body ast.Statement` with `BytecodeID int`
- Replace `DefaultValue ast.Expression` with pre-compiled bytecode
- Replace `InitValue ast.Expression` with pre-compiled bytecode
- Completely eliminate AST dependencies from runtime

### Performance Optimizations

- Inline method caches for virtual dispatch
- Method JIT compilation for hot paths
- Metadata sharing for identical signatures

## Conclusion

This design provides a clear path to eliminating AST dependencies from runtime types while maintaining compatibility with existing code. The incremental migration strategy reduces risk and allows validation at each step.

By the end of Phase 9, the Evaluator will work with clean, AST-free metadata, enabling removal of the adapter pattern and paving the way for future bytecode optimization.
