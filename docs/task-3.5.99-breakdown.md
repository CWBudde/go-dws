# Task 3.5.99: Advanced Index Expression Cases - Breakdown

**Date**: 2025-11-25
**Status**: Split into 7 subtasks (3.5.99a-g)
**Total Estimated Time**: 5-7 hours across 7 focused tasks

## Overview

Task 3.5.99 eliminates the remaining 4 `EvalNode` delegation points in `VisitIndexExpression` for advanced indexing scenarios. The original task was too large (~400+ lines of complex logic) and has been split into 7 granular subtasks following the successful pattern from tasks 3.5.93, 3.5.97, and 3.5.98.

## Current EvalNode Delegation Points

Location: `internal/interp/evaluator/visitor_expressions.go` in `VisitIndexExpression`

1. **Line 2014**: Indexed property access via `MemberAccessExpression`
   - Example: `obj.Property[x, y]`
   - Complexity: HIGH (~300 lines in `objects_properties.go`)

2. **Lines 2041-2043**: Default property access for Object/Interface/Record
   - Example: `obj[index]` → `obj.DefaultProperty[index]`
   - Complexity: MEDIUM (~100 lines in `array.go`)

3. **Line 2046**: JSON object/array indexing
   - Example: `jsonObj["key"]`, `jsonArr[5]`
   - Complexity: LOW (~40 lines in `array.go`)

## Blocking Factors

### 1. Circular Import Issue
- `JSONValue` and `VariantValue` are defined in `internal/interp/value.go`
- Evaluator is in `internal/interp/evaluator/` (child package)
- Cannot import parent package → need interface-based solution

### 2. Complex Dependencies
- `evalIndexedPropertyRead()` requires:
  - Environment swapping (`i.env` management)
  - Type resolution (`resolveTypeFromAnnotation`)
  - Default value creation (`getDefaultValue`)
  - Method lookup on classes (`lookupMethod`)
  - All currently in Interpreter

### 3. Property Metadata Access
- Need `ClassInfo.lookupProperty()` and `getDefaultProperty()`
- Need `InterfaceInfo.getDefaultProperty()`
- Need `RecordType.Properties` access

## Subtask Breakdown

### **3.5.99a: Property Access Infrastructure** ⏱️ 30-45 min
**Priority**: Must be done first (foundation for 3.5.99c-e)

**Objective**: Create interfaces for property access on runtime types

**Tasks**:
- Define `PropertyAccessor` interface in evaluator
  ```go
  type PropertyAccessor interface {
      GetProperty(name string) *PropertyInfo
      GetDefaultProperty() *PropertyInfo
      GetIndexedProperty(name string, indices []Value) (Value, error)
  }
  ```
- Implement in `ObjectInstance` (runtime/object.go)
- Implement in `InterfaceInstanceValue` (runtime/interface.go)
- Implement in `RecordValue` (runtime/record.go)

**Files**:
- `internal/interp/evaluator/evaluator.go`
- `internal/interp/runtime/object.go`
- `internal/interp/runtime/interface.go`
- `internal/interp/runtime/record.go`

**Acceptance**: All runtime types implement PropertyAccessor interface

---

### **3.5.99b: JSON Indexing Migration** ⏱️ 45-60 min
**Priority**: Can be done early (independent of other subtasks)

**Objective**: Move JSON indexing logic to evaluator, avoiding circular import

**Tasks**:
- Create `internal/interp/evaluator/json_helpers.go`
- Implement `indexJSON(base Value, index Value) (Value, error)`
  - Use type assertion to detect JSONValue without importing parent
  - Handle JSON objects (string index)
  - Handle JSON arrays (integer index)
  - Return as `VariantValue` (also via interface)
- Update `VisitIndexExpression` to call `indexJSON` directly
- Remove EvalNode delegation at line 2046

**Files**:
- `internal/interp/evaluator/json_helpers.go` (new)
- `internal/interp/evaluator/visitor_expressions.go`

**Tests**:
- `testdata/fixtures/JSONConnectorPass/`

**Acceptance**: JSON indexing works without EvalNode, all JSON tests pass

---

### **3.5.99c: Object Default Property Access** ⏱️ 45-60 min
**Priority**: After 3.5.99a (depends on PropertyAccessor)

**Objective**: Implement object default property indexing in evaluator

**Tasks**:
- Create `evalObjectDefaultPropertyRead(obj *ObjectInstance, index Value) (Value, error)` in evaluator
- Use `obj.GetDefaultProperty()` to get property info
- Handle single-index case
- Add error handling for missing default property
- Update `VisitIndexExpression` to handle OBJECT case directly
- Remove EvalNode delegation at lines 2041-2043

**Files**:
- `internal/interp/evaluator/visitor_expressions.go`
- `internal/interp/evaluator/index_ops.go` (new or existing)

**Tests**:
- `testdata/properties/default_property.dws`

**Acceptance**: Object default property indexing works, tests pass

---

### **3.5.99d: Interface Default Property Access** ⏱️ 30-45 min
**Priority**: After 3.5.99a and 3.5.99c

**Objective**: Handle interface default property access via object unwrapping

**Tasks**:
- Implement interface unwrapping in `VisitIndexExpression`
- Check for nil interface → error
- Extract underlying object from `InterfaceInstanceValue`
- Delegate to object's default property access (reuse 3.5.99c logic)
- Remove EvalNode delegation at line 2041

**Files**:
- `internal/interp/evaluator/visitor_expressions.go`

**Tests**:
- Interface property tests in `testdata/fixtures/`

**Acceptance**: Interface default property indexing works

---

### **3.5.99e: Record Default Property Access** ⏱️ 45-60 min
**Priority**: After 3.5.99a

**Objective**: Implement record default property with getter method calls

**Tasks**:
- Create `evalRecordDefaultPropertyRead(rec *RecordValue, index Value) (Value, error)`
- Use `rec.GetDefaultProperty()` to get property info
- Handle property getter method invocation
- Add parameter passing for index value
- Update `VisitIndexExpression` to handle RECORD case directly
- Remove EvalNode delegation at line 2043

**Files**:
- `internal/interp/evaluator/visitor_expressions.go`
- `internal/interp/evaluator/index_ops.go`

**Tests**:
- Record property tests

**Acceptance**: Record default property indexing works

---

### **3.5.99f: Indexed Property Infrastructure** ⏱️ 45-60 min
**Priority**: Before 3.5.99g (prepares for indexed property access)

**Objective**: Create adapter method for property getter invocation

**Tasks**:
- Add `CallIndexedPropertyGetter(property *PropertyInfo, self Value, indices []Value) (Value, error)` to adapter interface
- Implement in Interpreter (delegate to existing `evalIndexedPropertyRead` for now)
- Add environment management infrastructure to evaluator
  - May need `PushEnvironment(env *Environment)` / `PopEnvironment()`
  - Or `WithEnvironment(env *Environment, fn func()) (Value, error)`
- Document that this is temporary infrastructure (will be migrated later)

**Files**:
- `internal/interp/evaluator/evaluator.go` (adapter interface)
- `internal/interp/interpreter.go` (implementation)

**Acceptance**: Adapter method works, ready for 3.5.99g

---

### **3.5.99g: Indexed Property Access via MemberAccessExpression** ⏱️ 60-90 min
**Priority**: Last (most complex, depends on 3.5.99f)

**Objective**: Migrate indexed property access for `obj.Property[x, y]` syntax

**Tasks**:
- Detect when base is `MemberAccessExpression` in `VisitIndexExpression`
- Use existing `CollectIndices` helper to flatten multi-dimensional indices
- Lookup property on class via member name
- Call `CallIndexedPropertyGetter` adapter method with property, object, indices
- Handle return value and type
- Remove EvalNode delegation at line 2014

**Files**:
- `internal/interp/evaluator/visitor_expressions.go`
- `internal/interp/evaluator/property_helpers.go` (new, if needed)

**Tests**:
- `testdata/properties/indexed_property.dws`
- `testdata/properties/multi_index_property.dws`

**Acceptance**: Indexed property access works, all property tests pass

---

## Task Dependencies

```
3.5.99a (Infrastructure)
    ├──> 3.5.99c (Object default props)
    ├──> 3.5.99d (Interface default props - also needs 3.5.99c)
    └──> 3.5.99e (Record default props)

3.5.99b (JSON) - Independent

3.5.99f (Indexed prop infra)
    └──> 3.5.99g (Indexed prop access)
```

## Recommended Order

1. **3.5.99a** - Foundation (property interfaces)
2. **3.5.99b** - Quick win (JSON indexing)
3. **3.5.99c** - Object default properties
4. **3.5.99d** - Interface default properties
5. **3.5.99e** - Record default properties
6. **3.5.99f** - Indexed property infrastructure
7. **3.5.99g** - Indexed property access (finale)

## Success Metrics

- ✅ All 4 EvalNode delegation points removed from `VisitIndexExpression`
- ✅ All existing tests pass
- ✅ Property tests pass: default properties, indexed properties
- ✅ JSON tests pass: `testdata/fixtures/JSONConnectorPass/`
- ✅ No new adapter methods added (except temporary `CallIndexedPropertyGetter`)
- ✅ Zero circular imports
- ✅ Clean, maintainable code following evaluator visitor pattern

## Notes

- Each subtask should be a separate PR
- All tests must pass before merging each subtask
- Follow the established pattern from 3.5.93 (var params) and 3.5.97 (user functions)
- Document any temporary adapter methods with TODO comments
- Update PLAN.md as each subtask completes
