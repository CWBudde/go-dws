# InterpreterAdapter Audit - Task 3.5.24

This document audits all remaining methods in the `InterpreterAdapter` interface and justifies why each still exists. The goal is to establish a roadmap for eventual complete adapter removal.

**Generated**: Phase 3.5.24 Final Cleanup
**Interface Location**: `internal/interp/evaluator/evaluator.go:349-1024`
**Total Methods**: 75 methods

---

## Method Categories

### 1. Core Evaluation Fallback (1 method)

| Method | Reason Still Exists | Removal Difficulty |
|--------|--------------------|--------------------|
| `EvalNode(node ast.Node) Value` | **Essential fallback** for complex operations not yet migrated to native visitor methods. Used ~30+ times for: member assignments, compound ops, property access, etc. | **Hard** - Requires migrating all remaining delegation cases |

**EvalNode Usage Sites** (by file):
- `member_assignment.go`: 6 calls - member/property/index assignments
- `assignment_helpers.go`: 10 calls - assignment to various targets
- `visitor_expressions_members.go`: 7 calls - complex member access patterns
- `visitor_expressions_functions.go`: 3 calls - qualified calls, edge cases
- `visitor_statements.go`: 2 calls - complex statement patterns
- `index_assignment.go`: 3 calls - index assignments
- `helper_methods.go`: 1 call - helper method fallback
- `compound_ops.go`: 1 call - compound operation fallback
- `method_dispatch.go`: 1 call - edge case dispatch
- `evaluator.go`: 1 call - exception handling
- `visitor_declarations.go`: 1 call - unit declaration

---

### 2. Function/Method Execution (6 methods)

| Method | Reason Still Exists | Removal Difficulty |
|--------|--------------------|--------------------|
| `CallFunctionPointer(funcPtr Value, args []Value, node ast.Node) Value` | Executes function pointers (lambdas, method pointers). Used in `context.go` and `visitor_expressions_functions.go`. Requires Interpreter's environment/call stack management. | **Medium** - Move call stack to Evaluator |
| `CallUserFunction(fn *ast.FunctionDecl, args []Value) Value` | Executes user-defined functions. Still used for function pointer calls, method calls, explicit calls with arguments. | **Medium** - Blocked by environment migration |
| `LookupFunction(name string) ([]*ast.FunctionDecl, bool)` | Function registry access. Could be replaced by direct TypeSystem access. | **Easy** - Direct TypeSystem access |
| `EvalMethodImplementation(fn *ast.FunctionDecl) Value` | Method implementation registration. Requires ClassInfo VMT rebuild and descendant propagation. | **Hard** - Deep ClassInfo integration |
| `ExecuteFunctionPointerCall(metadata FunctionPointerMetadata, args []Value, node ast.Node) Value` | Low-level function pointer execution. Used in `visitor_statements.go` and `visitor_expressions_functions.go`. | **Medium** - Part of call system |
| `CallUserFunctionWithOverloads(callExpr *ast.CallExpression, funcName *ast.Identifier) Value` | User function call with overload resolution. Used in `visitor_expressions_functions.go`. | **Medium** - Overload logic could move to Evaluator |

---

### 3. Type Registry Lookups (5 methods)

| Method | Reason Still Exists | Removal Difficulty |
|--------|--------------------|--------------------|
| `LookupClass(name string) (any, bool)` | Class registry lookup. Returns `interface{}` to avoid import cycles. | **Easy** - Use TypeSystem directly |
| `ResolveClassInfoByName(name string) interface{}` | Class resolution for type resolution. | **Easy** - Use TypeSystem directly |
| `GetClassNameFromInfo(classInfo interface{}) string` | Extract class name from ClassInfo interface{}. | **Easy** - Type assertion in Evaluator |
| `LookupRecord(name string) (any, bool)` | Record registry lookup. | **Easy** - Use TypeSystem directly |
| `LookupInterface(name string) (any, bool)` | Interface registry lookup. | **Easy** - Use TypeSystem directly |

---

### 4. Helper System (11 methods)

| Method | Reason Still Exists | Removal Difficulty |
|--------|--------------------|--------------------|
| `LookupHelpers(typeName string) []any` | Helper lookup. | **Easy** - TypeSystem access |
| `CreateHelperInfo(name string, targetType any, isRecordHelper bool) interface{}` | HelperInfo creation. Avoids `interp` import in evaluator. | **Medium** - Move HelperInfo to runtime |
| `SetHelperParent(helper interface{}, parent interface{})` | Helper inheritance. | **Medium** - Move HelperInfo to runtime |
| `VerifyHelperTargetTypeMatch(parent interface{}, targetType any) bool` | Type matching for helpers. | **Medium** - Move HelperInfo to runtime |
| `GetHelperName(helper interface{}) string` | Helper name extraction. | **Easy** - Type assertion |
| `AddHelperMethod(helper interface{}, normalizedName string, method *ast.FunctionDecl)` | Register helper method. | **Medium** - Move HelperInfo to runtime |
| `AddHelperProperty(helper interface{}, prop *ast.PropertyDecl, propType any)` | Register helper property. | **Medium** - Move HelperInfo to runtime |
| `AddHelperClassVar(helper interface{}, name string, value Value)` | Add class variable. | **Medium** - Move HelperInfo to runtime |
| `AddHelperClassConst(helper interface{}, name string, value Value)` | Add class constant. | **Medium** - Move HelperInfo to runtime |
| `RegisterHelperLegacy(typeName string, helper interface{})` | Legacy helper registration. | **Easy** - Direct TypeSystem access |

---

### 5. Interface System (8 methods)

| Method | Reason Still Exists | Removal Difficulty |
|--------|--------------------|--------------------|
| `NewInterfaceInfoAdapter(name string) interface{}` | InterfaceInfo creation. Avoids `interp` import. | **Medium** - Move InterfaceInfo to runtime |
| `CastToInterfaceInfo(iface interface{}) (interface{}, bool)` | Type assertion for interfaces. | **Easy** - Direct type assertion |
| `SetInterfaceParent(iface interface{}, parent interface{})` | Interface inheritance. | **Medium** - Move InterfaceInfo to runtime |
| `GetInterfaceName(iface interface{}) string` | Interface name extraction. | **Easy** - Type assertion |
| `GetInterfaceParent(iface interface{}) interface{}` | Interface parent access. | **Easy** - Type assertion |
| `AddInterfaceMethod(iface interface{}, normalizedName string, method *ast.FunctionDecl)` | Register interface method. | **Medium** - Move InterfaceInfo to runtime |
| `AddInterfaceProperty(iface interface{}, normalizedName string, propInfo any)` | Register interface property. | **Medium** - Move InterfaceInfo to runtime |
| `CreateInterfaceWrapper(interfaceName string, obj Value) (Value, error)` | Create InterfaceInstance wrapper. Used in `visitor_expressions_types.go`. | **Medium** - Move wrapper creation to runtime |

---

### 6. Class Declaration (20 methods)

| Method | Reason Still Exists | Removal Difficulty |
|--------|--------------------|--------------------|
| `NewClassInfoAdapter(name string) interface{}` | ClassInfo creation. Avoids `interp` import. | **Medium** - IClassInfo interface exists in runtime |
| `CastToClassInfo(class interface{}) (interface{}, bool)` | Type assertion for classes. | **Easy** - Direct type assertion |
| `GetClassNameFromClassInfoInterface(classInfo interface{}) string` | Extract name from ClassInfo. | **Easy** - Type assertion |
| `RegisterClassEarly(name string, classInfo interface{})` | Early class registration for field initializers. | **Medium** - TypeSystem direct access |
| `IsClassPartial(classInfo interface{}) bool` | Check partial class. | **Easy** - Type assertion |
| `SetClassPartial(classInfo interface{}, isPartial bool)` | Set partial flag. | **Easy** - Type assertion |
| `SetClassAbstract(classInfo interface{}, isAbstract bool)` | Set abstract flag. | **Easy** - Type assertion |
| `SetClassExternal(classInfo interface{}, isExternal bool, externalName string)` | Set external class. | **Easy** - Type assertion |
| `ClassHasNoParent(classInfo interface{}) bool` | Check if parent is nil. | **Easy** - Type assertion |
| `DefineCurrentClassMarker(env interface{}, classInfo interface{})` | Marker for nested type resolution. | **Medium** - Environment access |
| `SetClassParent(classInfo interface{}, parentClass interface{})` | Set class parent, copy inherited members. | **Medium** - Complex inheritance logic |
| `AddInterfaceToClass(classInfo interface{}, interfaceInfo interface{}, interfaceName string)` | Add interface implementation. | **Medium** - Type system integration |
| `AddClassMethod(classInfo interface{}, method *ast.FunctionDecl, className string) bool` | Register class method. | **Medium** - MethodRegistry integration |
| `CreateMethodMetadata(method *ast.FunctionDecl) interface{}` | Create MethodMetadata. | **Medium** - Move to runtime package |
| `SynthesizeDefaultConstructor(classInfo interface{})` | Create implicit constructors. | **Medium** - ClassInfo internals |
| `AddClassProperty(classInfo interface{}, propDecl *ast.PropertyDecl) bool` | Register class property. | **Medium** - PropertyInfo creation |
| `RegisterClassOperator(classInfo interface{}, opDecl *ast.OperatorDecl) Value` | Register operator overload. | **Medium** - OperatorRegistry integration |
| `LookupClassMethod(classInfo interface{}, methodName string, isClassMethod bool) (interface{}, bool)` | Method lookup. | **Easy** - Type assertion |
| `SetClassConstructor(classInfo interface{}, constructor interface{})` | Set constructor field. | **Easy** - Type assertion |
| `SetClassDestructor(classInfo interface{}, destructor interface{})` | Set destructor field. | **Easy** - Type assertion |
| `InheritDestructorIfMissing(classInfo interface{})` | Inherit parent destructor. | **Medium** - Inheritance logic |
| `InheritParentProperties(classInfo interface{})` | Copy parent properties. | **Medium** - Inheritance logic |
| `BuildVirtualMethodTable(classInfo interface{})` | Build VMT. | **Hard** - Complex virtual dispatch |
| `RegisterClassInTypeSystem(classInfo interface{}, parentName string)` | Final registration. | **Easy** - TypeSystem direct access |

---

### 7. Object Operations (14 methods)

| Method | Reason Still Exists | Removal Difficulty |
|--------|--------------------|--------------------|
| `CallMethod(obj Value, methodName string, args []Value, node ast.Node) Value` | Object method call. Used in `method_dispatch.go`. | **Medium** - Method dispatch logic |
| `CallInheritedMethod(obj Value, methodName string, args []Value) Value` | Inherited method call. DEPRECATED. | **Easy** - Remove deprecated usage |
| `ExecuteMethodWithSelf(self Value, methodDecl any, args []Value) Value` | Low-level method execution with Self binding. | **Medium** - Environment setup |
| `ExecuteConstructor(obj Value, constructorName string, args []Value) error` | Constructor execution on created object. | **Medium** - Constructor logic |
| `CheckType(obj Value, typeName string) bool` | `is` operator. | **Easy** - Can move to runtime |
| `GetClassMetadataFromValue(obj Value) *runtime.ClassMetadata` | Extract metadata from object. | **Easy** - Type assertion |
| `GetObjectInstanceFromValue(val Value) interface{}` | Extract ObjectInstance. | **Easy** - Type assertion |
| `GetInterfaceInstanceFromValue(val Value) (interfaceInfo interface{}, underlyingObject interface{})` | Extract InterfaceInstance components. | **Easy** - Type assertion |
| `CreateTypeCastWrapper(className string, obj Value) Value` | TypeCastValue for `TypeName(expr)`. | **Medium** - Runtime type creation |
| `RaiseTypeCastException(message string, node ast.Node)` | Raise type cast exception. | **Easy** - Exception creation |
| `RaiseAssertionFailed(customMessage string)` | Assert exception. | **Easy** - Exception creation |
| `CreateContractException(className, message string, node ast.Node, classMetadata interface{}, callStack interface{}) interface{}` | Contract violation exception. | **Easy** - Exception creation |
| `CleanupInterfaceReferences(env interface{})` | Reference cleanup when scope ends. | **Medium** - Reference counting |
| `CreateClassValue(className string) (Value, error)` | Create ClassValue (metaclass). | **Easy** - TypeSystem lookup |

---

### 8. Property/Field Access (12 methods)

| Method | Reason Still Exists | Removal Difficulty |
|--------|--------------------|--------------------|
| `GetObjectFieldValue(obj Value, fieldName string) (Value, bool)` | Field access. | **Easy** - ObjectInstance method |
| `GetClassVariableValue(obj Value, varName string) (Value, bool)` | Class variable access. | **Easy** - ClassInfo method |
| `ReadPropertyValue(obj Value, propName string, node any) (Value, error)` | Property read. DEPRECATED. | **Easy** - Remove deprecated |
| `ExecutePropertyRead(obj Value, propInfo any, node any) Value` | Property read with PropertyInfo. Used ~10 times. | **Medium** - Getter method execution |
| `IsMethodParameterless(obj Value, methodName string) bool` | Check method signature. | **Easy** - Method lookup |
| `CreateMethodCall(obj Value, methodName string, node any) Value` | Auto-invocation for parameterless. | **Easy** - Call method directly |
| `CreateMethodPointerFromObject(obj Value, methodName string) (Value, error)` | Create method pointer. | **Medium** - FunctionPointerValue |
| `CreateBoundMethodPointer(obj Value, methodDecl any) Value` | Bound method pointer. Used 2 times. | **Medium** - FunctionPointerValue |
| `GetClassName(obj Value) string` | Get object's class name. | **Easy** - Type assertion |
| `GetClassType(obj Value) Value` | Get metaclass. | **Easy** - Lookup |
| `GetClassNameFromClassInfo(classInfo Value) string` | ClassInfoValue name. | **Easy** - Type assertion |
| `GetClassTypeFromClassInfo(classInfo Value) Value` | ClassInfoValue metaclass. | **Easy** - Lookup |
| `GetClassVariableFromClassInfo(classInfo Value, varName string) (Value, bool)` | Class variable from metaclass. | **Easy** - ClassInfo method |

---

### 9. Indexed Property Access (5 methods)

| Method | Reason Still Exists | Removal Difficulty |
|--------|--------------------|--------------------|
| `CallIndexedPropertyGetter(obj Value, propImpl any, indices []Value, node any) Value` | Default property getter. DEPRECATED. | **Easy** - Remove deprecated |
| `ExecuteIndexedPropertyRead(obj Value, propInfo any, indices []Value, node any) Value` | Indexed property read. Used ~8 times. | **Medium** - Getter method execution |
| `CallRecordPropertyGetter(record Value, propImpl any, indices []Value, node any) Value` | Record property getter. DEPRECATED. | **Easy** - Remove deprecated |
| `ExecuteRecordPropertyRead(record Value, propInfo any, indices []Value, node any) Value` | Record property read. Used ~2 times. | **Medium** - Getter method execution |

---

### 10. Method/Qualified Calls (5 methods)

| Method | Reason Still Exists | Removal Difficulty |
|--------|--------------------|--------------------|
| `CallMemberMethod(callExpr, memberAccess, objVal) Value` | Member method dispatch. DEPRECATED. | **Easy** - Remove deprecated |
| `CallQualifiedOrConstructor(callExpr, memberAccess) Value` | Unit-qualified calls. Still used for unit functions. | **Medium** - Unit system |
| `CallImplicitSelfMethod(callExpr, funcName) Value` | Implicit Self method calls. | **Medium** - Self binding |
| `CallRecordStaticMethod(callExpr, funcName) Value` | Record static methods. DEPRECATED. | **Easy** - Remove deprecated |
| `DispatchRecordStaticMethod(recordTypeName string, callExpr, funcName) Value` | Record static dispatch. | **Medium** - Record method lookup |

---

### 11. Binary Operations (3 methods)

| Method | Reason Still Exists | Removal Difficulty |
|--------|--------------------|--------------------|
| `EvalVariantBinaryOp(op string, left, right Value, node ast.Node) Value` | Variant binary ops. Prevents double-evaluation. | **Medium** - Variant type handling |
| `EvalInOperator(value, container Value, node ast.Node) Value` | `in` operator. Prevents double-evaluation. | **Medium** - Collection type handling |
| `EvalEqualityComparison(op string, left, right Value, node ast.Node) Value` | `=` and `<>` for complex types. | **Medium** - Type-specific equality |

---

### 12. Array/Collection Creation (2 methods)

| Method | Reason Still Exists | Removal Difficulty |
|--------|--------------------|--------------------|
| `CreateArray(elementType any, elements []Value) Value` | Array from elements. | **Easy** - Runtime package |
| `CreateArrayValue(arrayType any, elements []Value) Value` | ArrayValue with type. | **Easy** - Runtime package |

---

### 13. Lambda/Function Pointers (2 methods)

| Method | Reason Still Exists | Removal Difficulty |
|--------|--------------------|--------------------|
| `CreateLambda(lambda *ast.LambdaExpression, closure any) Value` | Lambda creation with closure. | **Medium** - Environment capture |
| `CreateMethodPointer(obj Value, methodName string, closure any) (Value, error)` | Method pointer creation. | **Medium** - Bound method handling |

---

### 14. Exception Handling (2 methods - temporary)

| Method | Reason Still Exists | Removal Difficulty |
|--------|--------------------|--------------------|
| `CreateExceptionDirect(classMetadata any, message string, pos any, callStack any) any` | Exception with pre-resolved deps. **Temporary**. | **Easy** - Move to runtime |
| `WrapObjectInException(objInstance Value, pos any, callStack any) any` | Wrap object in exception. **Temporary**. | **Easy** - Move to runtime |

---

### 15. Variable/Subrange Wrapping (2 methods)

| Method | Reason Still Exists | Removal Difficulty |
|--------|--------------------|--------------------|
| `WrapInSubrange(value Value, subrangeTypeName string, node ast.Node) (Value, error)` | Subrange validation/wrapping. | **Medium** - Subrange lookup |
| `WrapInInterface(value Value, interfaceName string, node ast.Node) (Value, error)` | Interface wrapping with validation. | **Medium** - Interface lookup |

---

### 16. Operator Registry (2 methods)

| Method | Reason Still Exists | Removal Difficulty |
|--------|--------------------|--------------------|
| `GetOperatorRegistry() any` | Operator overload lookups. | **Easy** - TypeSystem direct access |
| `GetEnumTypeID(enumName string) int` | Enum type ID lookup. | **Easy** - TypeSystem direct access |

---

## Summary by Removal Difficulty

| Difficulty | Count | Methods |
|------------|-------|---------|
| **Easy** | 35 | Lookup methods, type assertions, deprecated methods, simple wrappers |
| **Medium** | 32 | Helper/Interface/Class system methods, method execution, property access |
| **Hard** | 3 | EvalNode, EvalMethodImplementation, BuildVirtualMethodTable |

---

## Recommended Removal Roadmap

### Phase 1: Remove Deprecated Methods (Easy, ~2-4 hours)
- Remove `CallInheritedMethod` (use ExecuteMethodWithSelf)
- Remove `ReadPropertyValue` (use ExecutePropertyRead)
- Remove `CallIndexedPropertyGetter` (use ExecuteIndexedPropertyRead)
- Remove `CallRecordPropertyGetter` (use ExecuteRecordPropertyRead)
- Remove `CallMemberMethod` (evaluator uses DispatchMethodCall)
- Remove `CallRecordStaticMethod` (use DispatchRecordStaticMethod)

### Phase 2: Simplify Type Assertions (Easy, ~4-6 hours)
- Move simple lookup/assertion methods to direct TypeSystem calls
- Remove: `LookupClass`, `LookupRecord`, `LookupInterface`, `LookupHelpers`
- Remove: `GetOperatorRegistry`, `GetEnumTypeID`, `LookupFunction`
- Remove: Simple `Get*` methods that just do type assertions

### Phase 3: Migrate Type Creation to Runtime (Medium, ~8-12 hours)
- Move `HelperInfo` to `runtime` package with interface
- Move `InterfaceInfo` to `runtime` package (extend IInterfaceInfo)
- Move exception creation helpers to `runtime` package
- Remove: All helper adapter methods, interface adapter methods, exception adapters

### Phase 4: Migrate Class System (Medium, ~12-16 hours)
- Extend `IClassInfo` interface with all needed methods
- Move class creation/configuration methods to use IClassInfo directly
- Remove: Class declaration adapter methods

### Phase 5: Migrate Method/Property Execution (Medium, ~16-20 hours)
- Move method execution logic to Evaluator (requires call stack access)
- Move property read execution to use runtime interfaces
- Remove: `ExecutePropertyRead`, `ExecuteMethodWithSelf`, `CallMethod`

### Phase 6: Eliminate EvalNode Fallback (Hard, ~40+ hours)
- Migrate remaining ~30+ EvalNode delegation sites to native visitor methods
- Focus areas: member_assignment.go, assignment_helpers.go, visitor_expressions_members.go
- Final removal of adapter interface

---

## Essential Methods (Cannot Be Removed Without Major Refactoring)

These methods represent core functionality that would require significant architectural changes:

1. **EvalNode** - The catch-all fallback. Requires completing migration of all visitor methods.
2. **EvalMethodImplementation** - VMT rebuild, descendant propagation. Requires moving VMT logic to runtime.
3. **BuildVirtualMethodTable** - Complex virtual dispatch logic. Core to class inheritance.
4. **ExecuteMethodWithSelf** - Method execution with Self binding. Core to OOP support.
5. **CallUserFunction** - Function execution. Core to the interpreter.

---

## Conclusion

The adapter interface contains 75 methods, of which:
- **35 (47%)** are **Easy** to remove - mostly lookups and deprecated methods
- **32 (43%)** are **Medium** difficulty - require moving logic to runtime package
- **3 (4%)** are **Hard** - require architectural changes

The recommended approach is incremental removal starting with deprecated methods and simple lookups, progressing to type system migration, and finally tackling the hard cases.

**Estimated total effort**: 80-100 hours for complete adapter removal.
