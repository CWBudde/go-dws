# Task 9.4: Virtual Constructors - Implementation Analysis

**Date**: 2025-01-08
**Status**: 70% Complete - BLOCKED by Task 9.4.5
**Author**: Claude Code

## Summary

Task 9.4 aims to implement virtual constructors in go-dws, allowing polymorphic instantiation through metaclass variables and instance constructor calls. The semantic analysis and most runtime dispatch code is complete, but a critical pre-existing bug blocks final testing.

## Completed Work (70%)

### 1. Semantic Analysis Layer ✅ (100%)

**Files Modified**: `internal/semantic/analyze_classes.go`, `internal/semantic/type_resolution.go`

**Changes**:
- Updated `validateVirtualOverride()` (lines 709-785) to handle constructors
- Added `findMatchingConstructorInParent()` helper (lines 526-542 in type_resolution.go)
- Added `hasConstructorWithName()` helper (lines 546-558 in type_resolution.go)
- Validates virtual/override keywords on constructors
- Warns when hiding virtual constructor without override keyword

**Test Coverage**: Passes semantic validation for:
```pascal
type TClassA = class
  constructor Create; virtual;
end;

type TClassB = class(TClassA)
  constructor Create; override;  // Properly validated ✓
end;
```

### 2. Constructor Inheritance ✅ (100%)

**File Modified**: `internal/semantic/analyze_classes.go`

**Changes**:
- `inheritParentConstructors()` (lines 389-438) copies virtual metadata
- Line 417: `IsVirtual: parentCtor.IsVirtual` - Preserves virtual flag
- Line 418: `IsOverride: false` - Inherited constructors not marked override (correct)

**Behavior**: When TClassB extends TClassA, TClassB inherits TClassA's constructors with correct virtual flags.

### 3. Instance Constructor Virtual Dispatch ✅ (100%)

**File Modified**: `internal/interp/objects.go`

**Location**: Lines 1787-1873 in `evalMethodCall()`

**Implementation**:
```go
// When calling o.Create where o is an instance
if method.IsConstructor {
    // Find constructor in runtime class hierarchy
    for class := obj.Class; class != nil; class = class.Parent {
        if ctor, exists := class.Constructors[constructorName]; exists {
            actualConstructor = ctor
            break
        }
    }
    // Create NEW instance of runtime type
    newObj := NewObjectInstance(obj.Class)
    // ... initialize and execute constructor
    return newObj
}
```

**Test**: Works correctly for:
```pascal
var o, o2: TClassA;
o := TClassA.Create;    // Works ✓
o2 := o.Create;         // Creates new TClassA instance ✓

o := TClassB.Create;    // BLOCKED - see bug below
o2 := o.Create;         // Would create new TClassB instance (virtual dispatch)
```

### 4. Metaclass Constructor Virtual Dispatch ✅ (100%)

**File Modified**: `internal/interp/objects.go`

**Location**: Lines 1596-1689 in `evalMethodCall()`

**Implementation**:
```go
// Check if it's a metaclass (ClassValue) calling a constructor
if classVal, ok := objVal.(*ClassValue); ok {
    runtimeClass := classVal.ClassInfo

    // Get all constructor overloads from hierarchy
    constructorOverloads := i.getMethodOverloadsInHierarchy(runtimeClass, methodName, false)

    // Resolve overload
    constructor, err := i.resolveMethodOverload(runtimeClass.Name, methodName, constructorOverloads, mc.Arguments)

    // Create new instance of runtime class
    newInstance := NewObjectInstance(runtimeClass)

    // Execute constructor
    // ... (lines 1659-1686)

    return newInstance
}
```

**Test**: Would work for:
```pascal
var cls: class of TClassA;
var o: TClassA;
cls := TClassB;
o := cls.Create;  // Should create TClassB instance (virtual dispatch)
```

**Status**: Implementation complete, but BLOCKED by bug in TClassB.Create resolution.

## Critical Blocker: Task 9.4.5

### Bug Description

**Symptom**: Child class constructor calls fail with "There is no overloaded version" error

**Reproducible Test Case**:
```pascal
type TClassA = class constructor Create; virtual; end;
type TClassB = class(TClassA) constructor Create; override; end;
constructor TClassA.Create; begin end;
constructor TClassB.Create; begin end;

var o: TClassA;
o := TClassA.Create;  // ✓ Works
o := TClassB.Create;  // ✗ FAILS with "no overloaded version" error
```

**Error Message**:
```
Runtime error: ERROR: There is no overloaded version of "TClassB.Create"
that can be called with these arguments
```

### Bug Analysis

**What Works**:
- ✓ Parent class constructor calls (TClassA.Create)
- ✓ Instance constructor calls (o.Create where o is TClassA)
- ✓ Constructor declarations are parsed correctly
- ✓ Constructor implementations are parsed correctly
- ✓ Parent ConstructorOverloads are copied to child
- ✓ Child constructor implementation is added to ConstructorOverloads

**What Fails**:
- ✗ Direct child class constructor calls (TClassB.Create)
- ✗ Affects both parameterless and parameterized constructors
- ✗ Happens for any child class with override constructor

**Execution Flow for `TClassB.Create`**:

1. **Parser** → MethodCallExpression with `Object=Identifier("TClassB")`, `Method=Identifier("Create")`

2. **evalMethodCall()** (objects.go:1231)
   - Recognizes TClassB as a class name
   - Calls `getMethodOverloadsInHierarchy(TClassB, "Create", false)`

3. **getMethodOverloadsInHierarchy()** (objects.go:2035-2053)
   - Checks `TClassB.ConstructorOverloads["Create"]`
   - Should return: `[TClassA.Create, TClassB.Create]` (both are in list due to copying)
   - Returns this list

4. **resolveMethodOverload()** (objects.go:1981-2030)
   - Called with className="TClassB", methodName="Create", overloads=[...], arguments=[]
   - If len(overloads) == 1: Returns immediately ✓
   - If len(overloads) > 1: Calls `extractFunctionType()` for each
   - Calls `semantic.ResolveOverload(candidates, argTypes)`
   - **FAILS HERE** → Returns error

**Hypothesis**: The issue is in step 4. When there are 2 constructors in the list (parent + child), the overload resolution fails. Possible causes:

A. `extractFunctionType()` returns nil for one of the constructors
B. `semantic.ResolveOverload()` can't distinguish between them (same signature)
C. The constructor declarations/implementations have mismatched metadata
D. Override constructors have different handling requirements

### Files Involved

1. **internal/interp/declarations.go**
   - `evalFunctionDeclaration()` (lines 13-73) - Processes constructor implementations
   - `replaceMethodInOverloadList()` (lines 517-542) - Replaces declaration with implementation

2. **internal/interp/objects.go**
   - `evalMethodCall()` (lines 1231-1450) - Handles TClassB.Create calls
   - `getMethodOverloadsInHierarchy()` (lines 2035-2107) - Returns constructor overloads
   - `resolveMethodOverload()` (lines 1981-2030) - Fails here
   - `extractFunctionType()` - Extracts type from FunctionDecl

3. **internal/semantic/overload.go**
   - `ResolveOverload()` - Semantic overload resolution

### Investigation Steps (Task 9.4.5.1-9.4.5.3)

#### 9.4.5.1: Debug Constructor Overload Resolution

Add logging to understand what's happening:

```go
// In getMethodOverloadsInHierarchy() around line 2045
if !isClassMethod {
    for ctorName, constructorOverloads := range classInfo.ConstructorOverloads {
        if strings.EqualFold(ctorName, methodName) {
            // ADD LOGGING HERE
            fmt.Printf("DEBUG: Found %d constructor overloads for %s.%s\n",
                len(constructorOverloads), classInfo.Name, methodName)
            for i, ctor := range constructorOverloads {
                fmt.Printf("  [%d] %s.%s (%d params, IsVirtual=%v, IsOverride=%v)\n",
                    i, ctor.ClassName.Value, ctor.Name.Value,
                    len(ctor.Parameters), ctor.IsVirtual, ctor.IsOverride)
            }
            result = append(result, constructorOverloads...)
            return result
        }
    }
}
```

```go
// In resolveMethodOverload() around line 1998
candidates := make([]*semantic.Symbol, len(overloads))
for idx, method := range overloads {
    methodType := i.extractFunctionType(method)
    // ADD LOGGING HERE
    fmt.Printf("DEBUG: Extracting type for overload %d: %s.%s\n", idx, className, methodName)
    if methodType == nil {
        fmt.Printf("  ERROR: extractFunctionType returned nil!\n")
        return nil, fmt.Errorf("unable to extract method type for overload %d", idx)
    }
    fmt.Printf("  ParamTypes: %v, ReturnType: %v\n", methodType.ParamTypes, methodType.ReturnType)
    // ...
}
```

**Expected Output**:
- Should see 2 constructor overloads for TClassB.Create
- Both should have 0 parameters
- One should be IsVirtual=true (parent), one IsOverride=true (child)
- extractFunctionType should return valid types for both

**If extractFunctionType returns nil**: Bug is in type extraction (go to 9.4.5.2)
**If both have valid types**: Bug is in semantic.ResolveOverload (go to 9.4.5.3)

#### 9.4.5.2: Investigate extractFunctionType

Check the `extractFunctionType()` function:

```go
func (i *Interpreter) extractFunctionType(fn *ast.FunctionDecl) *types.FunctionType {
    // Find this function in the codebase
    // Check if it handles constructors correctly
    // Verify it resolves parameter types
    // Check return type handling for constructors
}
```

**Questions to Answer**:
1. Does it handle constructors differently than regular methods?
2. Does it properly resolve parameter types when parameters are empty?
3. What does it return for return type on constructors?
4. Does IsOverride or IsVirtual affect type extraction?

#### 9.4.5.3: Test semantic.ResolveOverload

Check if `semantic.ResolveOverload()` can handle two constructors with identical signatures:

```go
// In semantic/overload.go
func ResolveOverload(candidates []*Symbol, argTypes []types.Type) (*Symbol, error) {
    // When called with:
    // - candidates = [TClassA.Create, TClassB.Create]
    // - argTypes = [] (no arguments)
    // Both constructors have signature: () -> TClassA/TClassB

    // Should this:
    // A. Return error because ambiguous? (current behavior?)
    // B. Prefer the most derived class (TClassB)?
    // C. Return the first match?
}
```

**Possible Fix**: Override constructors should hide parent constructors, so only TClassB.Create should be in the candidates list.

### Proposed Solutions

#### Solution A: Filter Parent Constructors (Recommended)

When a child class has an override constructor, don't include the parent's constructor in the overload list.

**Location**: `getMethodOverloadsInHierarchy()` or `evalMethodCall()`

**Implementation**:
```go
// After getting constructorOverloads from classInfo.ConstructorOverloads
// Filter out parent constructors that are overridden
var filteredOverloads []*ast.FunctionDecl
for _, ctor := range constructorOverloads {
    // Skip if this is a parent constructor and child has override
    isParentCtor := ctor.ClassName != nil && ctor.ClassName.Value != classInfo.Name
    if isParentCtor {
        // Check if child has override for this signature
        hasOverride := false
        for _, childCtor := range constructorOverloads {
            if childCtor.ClassName != nil && childCtor.ClassName.Value == classInfo.Name {
                if signaturesMatch(ctor, childCtor) {
                    hasOverride = true
                    break
                }
            }
        }
        if hasOverride {
            continue // Skip parent constructor
        }
    }
    filteredOverloads = append(filteredOverloads, ctor)
}
return filteredOverloads
```

#### Solution B: Don't Copy Parent ConstructorOverloads

Follow the same pattern as MethodOverloads (Task 9.21.6) - don't copy from parent.

**Location**: `internal/interp/declarations.go` - evalClassDeclaration()

**Implementation**:
```go
// Remove lines 123-125 (don't copy parent ConstructorOverloads)
// Update getMethodOverloadsInHierarchy() to walk hierarchy for constructors
// Filter overridden constructors during hierarchy walk
```

**Pros**: Consistent with method handling, cleaner design
**Cons**: More changes required, need to update hierarchy walk logic

#### Solution C: Fix replaceMethodInOverloadList

Make `replaceMethodInOverloadList()` remove parent constructors when child overrides.

**Location**: `internal/interp/declarations.go:517-542`

**Implementation**:
```go
func (i *Interpreter) replaceMethodInOverloadList(list []*ast.FunctionDecl, impl *ast.FunctionDecl) []*ast.FunctionDecl {
    // ... existing matching logic ...
    if match {
        list[idx] = impl

        // If this is an override constructor, remove parent constructors with same signature
        if impl.IsConstructor && impl.IsOverride {
            return []*ast.FunctionDecl{impl} // Only keep the override
        }

        return list
    }
    // ...
}
```

### Test Cases to Validate Fix

```pascal
// Test 1: Simple override (currently failing)
type TClassA = class constructor Create; virtual; end;
type TClassB = class(TClassA) constructor Create; override; end;
constructor TClassA.Create; begin end;
constructor TClassB.Create; begin end;
var o: TClassA;
o := TClassB.Create;  // Must work

// Test 2: Parameterized override
type TClassA = class constructor Create(x: Integer); virtual; end;
type TClassB = class(TClassA) constructor Create(x: Integer); override; end;
constructor TClassA.Create(x: Integer); begin end;
constructor TClassB.Create(x: Integer); begin end;
var o: TClassA;
o := TClassB.Create(42);  // Must work

// Test 3: Multiple constructors
type TClassA = class
  constructor Create; virtual;
  constructor Create(x: Integer); virtual;
end;
type TClassB = class(TClassA)
  constructor Create; override;
  constructor Create(x: Integer); override;
end;
// ... implementations ...
var o: TClassA;
o := TClassB.Create;      // Must work
o := TClassB.Create(42);  // Must work
```

## Remaining Work (30%)

### Task 9.4.5: Fix Child Class Constructor Bug (Priority: CRITICAL)

Subtasks 9.4.5.1 through 9.4.5.8 detailed above.

**Estimated Effort**: 4-8 hours
- Investigation: 1-2 hours
- Fix implementation: 2-4 hours
- Testing: 1-2 hours

### Bytecode VM Support (Not Started)

**Files to Modify**:
- `internal/bytecode/compiler.go` - Compile virtual constructor calls
- `internal/bytecode/vm.go` - Execute virtual constructor dispatch
- `internal/bytecode/instruction.go` - Add opcodes if needed

**Implementation Notes**:
- Metaclass constructor calls need special opcode or handling
- Must look up constructor in runtime class, not compile-time class
- Instance constructor calls (o.Create) need virtual dispatch

**Estimated Effort**: 4-6 hours

### Final Testing

Once Task 9.4.5 is complete:

1. Run fixture tests:
   ```bash
   go test ./internal/interp -run TestDWScriptFixtures/SimpleScripts/virtual_constructor
   go test ./internal/interp -run TestDWScriptFixtures/SimpleScripts/virtual_constructor2
   ```

2. Expected output:
   - `virtual_constructor.pas`: "B\nTestA\nTestB"
   - `virtual_constructor2.pas`: "A\nA\nB\nB"

3. Mark Task 9.4 as complete in PLAN.md

## References

- PLAN.md: Task 9.4 (line 134)
- Fixtures: `testdata/fixtures/SimpleScripts/virtual_constructor*.pas`
- Related: Task 9.72 (Metaclasses), Task 9.73 (Metaclass dispatch)

## Notes for Future Work

- Virtual constructors are rare in DWScript code but essential for polymorphic factories
- The metaclass dispatch implementation (9.4.4) is solid and follows DWScript semantics
- Instance constructor calls (9.4.3) enable copy/clone patterns
- Once bug 9.4.5 is fixed, the feature should be production-ready
