package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// HelperInfo is re-exported from helper_methods.go for use in this file.
// This import is necessary because the types package's PropertyInfo is used
// in type assertions on the result of GetPropertyAny.

// This file contains visitor methods for member access expression AST nodes.
// Member access includes field access, property access, and method references on objects.

// VisitMemberAccessExpression evaluates member access (obj.field, obj.method).
//
// **COMPLEXITY**: Very High (700+ lines in original implementation)
// **STATUS**: Documentation-only migration with full adapter delegation
// **Task 3.5.25**: Simple Modes (Unit-qualified, Enum, Static class, Record fields)
// **Task 3.5.26**: Complex Modes (Object instance, Interface, Metaclass, Type casts)
//
// **11 DISTINCT ACCESS MODES** (evaluated in this order):
//
// === SIMPLE MODES (Task 3.5.25) ===
//
// **1. UNIT-QUALIFIED ACCESS** (UnitName.Symbol) [Task 3.5.25]
//   - Pattern: `Math.PI`, `System.Print`
//   - Evaluation order:
//     a. Check if left side is a registered unit name (via unitRegistry)
//     b. Try to resolve as qualified variable/constant (ResolveQualifiedVariable)
//     c. If not a variable, it might be a function reference (handled in VisitCallExpression)
//   - Error: "qualified name 'Unit.Symbol' cannot be used as a value (functions must be called)"
//   - Implementation: ~14 lines in original
//
// **2. STATIC CLASS ACCESS** (TClass.Member) [Task 3.5.25]
//   - Pattern: `TMyClass.ClassVar`, `TMyClass.Create`, `TMyClass.ClassName`
//   - Lookup order (case-insensitive):
//     a. Built-in properties: `ClassName` (string), `ClassType` (metaclass reference)
//     b. Class variables (lookupClassVar) - inherited from parent classes
//     c. Class constants (getClassConstant) - lazy evaluation with caching
//     d. Class properties (lookupProperty) - if IsClassProperty or uses class-level read specs
//     e. Constructors (HasConstructor) - auto-invoke with 0 arguments if no parentheses
//   - Supports constructor overloading and inheritance
//   - Falls back to implicit parameterless constructor if needed
//     f. Class methods (lookupClassMethodInHierarchy) - static methods
//   - Parameterless: auto-invoke via VisitMethodCallExpression
//   - With parameters: return as FunctionPointerValue
//   - Error: "member 'X' not found in class 'Y'"
//   - Implementation: ~100 lines in original
//
// **3. ENUM TYPE ACCESS** (TColor.Red, TColor.Low, TColor.High) [Task 3.5.25]
//   - Pattern: `TColor.Red`, `TMyEnum.Low`, `TMyEnum.High`
//   - Lookup via TypeSystem (Task 3.5.143b)
//   - For scoped enums:
//     a. Look up enum value in EnumType.Values (takes precedence over properties)
//     b. Check for special properties: `Low` (lowest ordinal), `High` (highest ordinal)
//   - For unscoped enums: also check environment for value name
//   - Returns: EnumValue or IntegerValue (for Low/High)
//   - Error: "enum value 'X' not found in enum 'Y'"
//   - Implementation: ~45 lines in original
//
// **4. RECORD TYPE STATIC ACCESS** (TPoint.cOrigin, TPoint.Count) [Task 3.5.25]
//   - Pattern: `TPoint.cOrigin`, `TRecord.ClassMethod()`
//   - Lookup in environment: `__record_type_` + lowercase(recordTypeName)
//   - Lookup order (case-insensitive):
//     a. Constants (RecordTypeValue.Constants)
//     b. Class variables (RecordTypeValue.ClassVars)
//     c. Class methods (RecordTypeValue.ClassMethods)
//   - Parameterless: auto-invoke via VisitMethodCallExpression
//   - With parameters: error (requires parentheses to call)
//   - Error: "member 'X' not found in record type 'Y'"
//   - Implementation: ~40 lines in original
//
// **5. RECORD INSTANCE ACCESS** (record.Field, record.Method) [Task 3.5.25]
//   - Pattern: `point.X`, `point.GetLength()`, `point.Prop`
//   - Object type: RecordValue
//   - Lookup order (case-insensitive):
//     a. Direct field access (RecordValue.Fields)
//     b. Properties (RecordType.Properties):
//   - ReadField: field name -> direct access, method name -> call getter
//   - Write-only: error "property 'X' is write-only"
//     c. Instance methods (RecordValue.Methods):
//   - Parameterless: auto-invoke via VisitMethodCallExpression
//   - With parameters: error "method 'X' requires N parameter(s); use parentheses"
//     d. Class methods (from RecordTypeValue, accessible via instance)
//     e. Constants (from RecordTypeValue, accessible via instance)
//     f. Class variables (from RecordTypeValue, accessible via instance)
//     g. Helper properties (findHelperProperty -> evalHelperPropertyRead)
//   - Error: "field 'X' not found in record 'Y'"
//   - Implementation: ~115 lines in original
//
// **10. ENUM VALUE PROPERTIES** (enumVal.Value) [Task 3.5.25]
//   - Pattern: `TColor.Red.Value` (returns ordinal as integer)
//   - Object type: EnumValue
//   - Supported properties:
//     a. `.Value` (case-insensitive): returns OrdinalValue as IntegerValue
//     b. `.ToString`: handled by helpers (if available)
//   - Fallback: Check helpers for additional properties
//   - Implementation: ~10 lines in original
//   - NOTE: This is listed here (out of precedence order) for Task 3.5.25 grouping
//   - In actual implementation, this is checked at position 10
//
// === COMPLEX MODES (Task 3.5.26) ===
//
// **6. CLASS/METACLASS ACCESS** (ClassInfoValue/ClassValue.Member) [Task 3.5.26]
//   - Pattern: When a class name is evaluated to ClassInfoValue or ClassValue
//   - Example: `var c := TMyClass; c.Create()`
//   - Lookup order (same as static class access #2):
//     a. Built-in properties: `ClassName`, `ClassType`
//     b. Class variables, constants, properties, constructors, class methods
//   - Returns: String/ClassValue/field value/method pointer
//   - Implementation: ~95 lines in original
//
// **7. INTERFACE INSTANCE ACCESS** (interface.Method, interface.Property) [Task 3.5.26]
//   - Pattern: `intfVar.Hello`, `intfVar.SomeMethod`
//   - Object type: InterfaceInstance
//   - Validation: Verify member exists in interface definition (HasMethod)
//   - For methods:
//   - Look up implementation in underlying object's class (getMethodOverloadsInHierarchy)
//   - Return FunctionPointerValue bound to underlying object (NO auto-invoke)
//   - Enables method delegate assignment: `var h : procedure := i.Hello;`
//   - For properties/fields: delegate to underlying object (without validation currently)
//   - Unwrap interface to underlying object and continue evaluation
//   - Error: "Interface is nil" or "method 'X' declared in interface 'Y' but not implemented"
//   - Implementation: ~50 lines in original
//
// **8. TYPE CAST VALUE HANDLING** (TBase(child).ClassVar) [Task 3.5.26]
//   - Pattern: Accessing members through a type cast expression
//   - Object type: TypeCastValue
//   - Extracts: StaticType (for class variable lookup), Object (actual instance)
//   - Purpose: Class variables use static type, not runtime type
//   - Unwraps to actual object and continues evaluation with static type context
//   - Implementation: ~5 lines in original
//
// **9. NIL OBJECT HANDLING** (nil.ClassVar) [Task 3.5.26]
//   - Pattern: `var o: TMyClass := nil; o.ClassVar`
//   - Object type: NilValue (with ClassType field) or nil evaluation result
//   - Special case: Accessing class variables on nil instances is allowed
//   - Lookup:
//     a. If staticClassType from cast (TBase(nil).ClassVar): use static type
//     b. If NilValue.ClassType set: look up class and check for class variable
//   - Success: Return class variable value
//   - Failure: Error "Object not instantiated" (for instance members)
//   - Implementation: ~35 lines in original
//
// **11. OBJECT INSTANCE ACCESS** (obj.Field, obj.Method, obj.Property) [Task 3.5.26]
//   - Pattern: `myObj.Name`, `myObj.GetValue()`, `myObj.Count`
//   - Object type: ObjectInstance
//   - Built-in properties (inherited from TObject, case-insensitive):
//     a. `ClassName`: returns obj.Class.Name (runtime type)
//     b. `ClassType`: returns ClassValue (metaclass for runtime type)
//   - Lookup order (case-insensitive):
//     a. Properties (Class.lookupProperty) - takes precedence over fields
//   - Call evalPropertyRead for read accessor (field, method, or expression)
//     b. Direct field access (obj.GetField) - instance fields
//     c. Class variables (lookupClassVar) - accessible from instance
//   - Uses static type from cast if available (e.g., TBase(child).ClassVar)
//     d. Class constants (getClassConstant) - accessible from instance
//     e. Instance methods (getMethodOverloadsInHierarchy):
//   - Check all overloads for parameterless variants
//   - Parameterless: auto-invoke via VisitMethodCallExpression
//   - With parameters: return first overload as FunctionPointerValue
//     f. Class methods (getMethodOverloadsInHierarchy with classMethod=true)
//   - Same logic as instance methods (auto-invoke or function pointer)
//     g. Helper properties (findHelperProperty -> evalHelperPropertyRead)
//   - Error: "field 'X' not found in class 'Y'"
//   - Implementation: ~115 lines in original
//
// **SPECIAL BEHAVIORS**:
// - **Auto-invocation**: Parameterless methods/properties auto-invoke when accessed without ()
// - **Case-insensitive**: All name lookups are case-insensitive (DWScript spec)
// - **Inheritance**: Class variables, constants, properties, methods searched up hierarchy
// - **Helper support**: Type helpers can add properties/methods to any type
// - **Function pointers**: Methods with parameters return FunctionPointerValue
// - **Lazy evaluation**: Class constants evaluated once and cached on first access
// - **Type safety**: Static types respected for class variable access through casts
//
// **DEPENDENCIES** (blockers for full migration):
// - RecordValue, RecordTypeValue - in internal/interp (needs migration to runtime)
// - ObjectInstance, ClassInfo, ClassValue - in internal/interp (needs migration to runtime)
// - InterfaceInstance - in internal/interp (needs migration to runtime)
// - EnumValue, EnumTypeValue - in internal/interp (needs migration to runtime)
// - FunctionPointerValue - in internal/interp (needs migration to runtime)
// - TypeCastValue - in internal/interp (needs migration to runtime)
// - ClassInfoValue - in internal/interp (needs migration to runtime)
// - Helper infrastructure - findHelperProperty, findHelperMethod (needs adapter methods)
// - Method call infrastructure - evalMethodCall (delegated to VisitMethodCallExpression)
// - Property read infrastructure - evalPropertyRead, evalHelperPropertyRead (needs adapter)
// - Class hierarchy - lookupClassVar, lookupProperty, getMethodOverloadsInHierarchy (needs adapter)
// - Unit registry - ResolveQualifiedVariable (already in Evaluator via unitRegistry)
//
// **TESTING**:
// - Unit-qualified access (Math.PI, System.WriteLine)
// - Static class access (TMyClass.ClassVar, TMyClass.Create, TMyClass.ClassName)
// - Enum type access (TColor.Red, TColor.Low, TColor.High)
// - Record type static access (TPoint.cOrigin, TPoint.Count)
// - Record instance access (point.X, point.GetLength())
// - Object instance access (obj.Name, obj.GetValue(), obj.Count)
// - Interface method access (intf.Hello)
// - Type cast access (TBase(child).ClassVar)
// - Nil object access (nil.ClassVar, nil.Name -> error)
// - Enum value properties (TColor.Red.Value)
// - Helper properties/methods (arr.Length, str.ToUpper)
// - Auto-invocation (obj.Method without parentheses for parameterless)
// - Function pointers (obj.Method with parameters returns pointer)
//
// **IMPLEMENTATION SUMMARY**:
// - Original implementation: 706 lines (objects_hierarchy.go:13-719)
// - Handles 11 distinct access modes with complex precedence rules
// - Supports case-insensitive lookups, inheritance, helpers, auto-invocation
// - Requires extensive value type infrastructure not yet in runtime package
// - Full migration deferred - will be broken into category-specific sub-tasks
//
// **MIGRATION STRATEGY**:
// - Phase 1 (this task): Comprehensive documentation of all access modes
// - Phase 2 (future): Migrate simple cases (built-in properties, direct field access)
// - Phase 3 (future): Migrate class/record static access after type system migration
// - Phase 4 (future): Migrate helper infrastructure after helper system migration
// - Phase 5 (future): Migrate method/property dispatch after OOP infrastructure migration
func (e *Evaluator) VisitMemberAccessExpression(node *ast.MemberAccessExpression, ctx *ExecutionContext) Value {
	if node.Object == nil {
		return e.newError(node, "member access missing object")
	}
	if node.Member == nil {
		return e.newError(node, "member access missing member")
	}

	// Task 3.5.25 & 3.5.26: Implement member access with routing based on object type

	// Evaluate the object first
	obj := e.Eval(node.Object, ctx)
	if isError(obj) {
		return obj
	}

	memberName := node.Member.Value

	// Check for record instance using type assertion (before switch statement)
	// This handles ALL record types regardless of their Type() return value
	// (RecordValue.Type() returns the specific type name like "TPoint" instead of "RECORD")
	if recVal, ok := obj.(RecordInstanceValue); ok {
		// Direct field access - most common case
		if fieldVal, found := recVal.GetRecordField(memberName); found {
			return fieldVal
		}

		// Method reference - still uses adapter for method invocation
		if recVal.HasRecordMethod(memberName) {
			return e.adapter.EvalNode(node)
		}

		// Property access
		if recVal.HasRecordProperty(memberName) {
			// This should not happen for standard records, but if properties are added later,
			// this path would need proper property reading logic
			return e.newError(node, "property access on records not supported")
		}

		// Member not found - return proper error
		return e.newError(node, "field '%s' not found in record '%s'", memberName, recVal.GetRecordTypeName())
	}

	// Route based on object type
	switch obj.Type() {
	case "OBJECT":
		// Task 3.5.26: Object instance access (Mode 11)
		// Pattern: obj.Field, obj.Property, obj.Method
		// Lookup order per spec (lines 894-900): Properties -> Fields -> Class Variables

		// Task 3.5.86: Use ObjectValue interface for direct member access
		objVal, ok := obj.(ObjectValue)
		if !ok {
			// Task 3.5.101: Direct error instead of EvalNode delegation
			// If Type() returns "OBJECT" but ObjectValue interface is not implemented,
			// this is an internal inconsistency that should not happen
			return e.newError(node, "internal error: OBJECT value does not implement ObjectValue interface")
		}

		// Try property access first (with recursion protection)
		propCtx := ctx.PropContext()
		if propCtx == nil || (!propCtx.InPropertyGetter && !propCtx.InPropertySetter) {
			// Task 3.5.72: Use ObjectValue interface for direct property check
			// Task 3.5.116: Use ObjectValue.ReadProperty with callback pattern
			// Task 3.5.32: Use evaluator's executePropertyRead instead of adapter
			if objVal.HasProperty(memberName) {
				propValue := objVal.ReadProperty(memberName, func(propInfo any) Value {
					return e.executePropertyRead(obj, propInfo, node, ctx)
				})
				// Check if result is an error
				if propValue != nil && propValue.Type() != "ERROR" {
					return propValue
				}
			}
		}

		// Task 3.5.86: Direct field access via ObjectValue interface
		if fieldValue := objVal.GetField(memberName); fieldValue != nil {
			return fieldValue
		}

		// Task 3.5.86: Direct class variable access via ObjectValue interface
		if classVarValue, found := objVal.GetClassVar(memberName); found {
			return classVarValue
		}

		// Task 3.5.37: Check for helper properties before delegating to adapter
		if helper, propInfo := e.FindHelperProperty(obj, memberName); propInfo != nil {
			return e.executeHelperPropertyRead(helper, propInfo, obj, node, ctx)
		}

		// Try method or other member access via adapter
		return e.adapter.EvalNode(node)

	case "INTERFACE":
		// Task 3.5.26: Interface instance access (Mode 7)
		// Pattern: intf.Method, intf.Property

		// Task 3.5.87: Use InterfaceInstanceValue interface for direct member verification
		ifaceVal, ok := obj.(InterfaceInstanceValue)
		if !ok {
			// Task 3.5.101: Direct error instead of EvalNode delegation
			// If Type() returns "INTERFACE" but InterfaceInstanceValue interface is not implemented,
			// this is an internal inconsistency that should not happen
			return e.newError(node, "internal error: INTERFACE value does not implement InterfaceInstanceValue interface")
		}

		// Get underlying object - nil check is critical for interface access
		underlying := ifaceVal.GetUnderlyingObjectValue()
		if underlying == nil {
			// Task 3.5.101: Direct error for nil interface access
			return e.newError(node, "Interface is nil")
		}

		// Verify the member is part of the interface contract before delegating
		if ifaceVal.HasInterfaceProperty(memberName) {
			// Task 3.5.116: Use ObjectValue.ReadProperty with callback pattern
			// Task 3.5.32: Use evaluator's executePropertyRead instead of adapter
			if objVal, ok := underlying.(ObjectValue); ok {
				propValue := objVal.ReadProperty(memberName, func(propInfo any) Value {
					return e.executePropertyRead(underlying, propInfo, node, ctx)
				})
				// Check if result is an error
				if propValue != nil && propValue.Type() != "ERROR" {
					return propValue
				}
				// Task 3.5.101: Direct error instead of EvalNode delegation for property read failure
				return e.newError(node, "failed to read property '%s' on interface '%s': %v", memberName, ifaceVal.InterfaceName(), propValue)
			}
			// Underlying doesn't implement ObjectValue - should not happen for valid interfaces
			return e.newError(node, "internal error: interface underlying value does not implement ObjectValue")
		}

		// Method access or unknown member - delegate to adapter
		// Methods require complex dispatch logic (virtual method tables, etc.)
		return e.adapter.EvalNode(node)

	case "CLASS":
		// ClassValue from ClassType property or direct class reference
		// Uses same logic as CLASSINFO - both implement ClassMetaValue interface
		fallthrough

	case "CLASSINFO":
		// Task 3.5.88: Metaclass access (Mode 6)
		// Pattern: ClassInfoValue.Member or ClassValue.Member
		// Handle built-in properties and class variables/constants directly

		// Try to use ClassMetaValue interface for direct access
		classMetaVal, ok := obj.(ClassMetaValue)
		if !ok {
			// Task 3.5.101: Direct error instead of EvalNode delegation
			// If Type() returns "CLASSINFO" but ClassMetaValue interface is not implemented,
			// this is an internal inconsistency that should not happen
			return e.newError(node, "internal error: %s value does not implement ClassMetaValue interface", obj.Type())
		}

		// Handle built-in properties first
		if ident.Equal(memberName, "ClassName") {
			// ClassName property returns the class name as a string
			return &runtime.StringValue{Value: classMetaVal.GetClassName()}
		}
		if ident.Equal(memberName, "ClassType") {
			// ClassType property returns the class itself
			// For CLASSINFO, we return the same value
			return obj
		}

		// Try class variables - direct lookup without adapter
		if val, found := classMetaVal.GetClassVar(memberName); found {
			return val
		}

		// Try class constants - direct lookup without adapter
		if val, found := classMetaVal.GetClassConstant(memberName); found {
			return val
		}

		// Complex cases (constructors, class methods, properties) need adapter
		// because they require method invocation logic
		return e.adapter.EvalNode(node)

	case "TYPE_CAST":
		// Task 3.5.89: Type cast value handling (Mode 8)
		// Pattern: TBase(child).ClassVar
		// Uses static type from cast for class variable lookup, not runtime type.

		// Try TypeCastAccessor interface for direct access
		typeCastVal, ok := obj.(TypeCastAccessor)
		if !ok {
			// Task 3.5.101: Direct error instead of EvalNode delegation
			// If Type() returns "TYPE_CAST" but TypeCastAccessor interface is not implemented,
			// this is an internal inconsistency that should not happen
			return e.newError(node, "internal error: TYPE_CAST value does not implement TypeCastAccessor interface")
		}

		// First, try class variable lookup using the STATIC type
		// This is the key behavior: TBase(child).ClassVar accesses TBase's class var
		if classVarValue, found := typeCastVal.GetStaticClassVar(memberName); found {
			return classVarValue
		}

		// Get the wrapped value for further processing
		wrappedValue := typeCastVal.GetWrappedValue()

		// If wrapped value is an object, try field access and property reading
		if objVal, ok := wrappedValue.(ObjectValue); ok {
			// Try direct field access on the wrapped object
			if fieldValue := objVal.GetField(memberName); fieldValue != nil {
				return fieldValue
			}

			// Task 3.5.116: Try property access using ObjectValue.ReadProperty with callback pattern
			// Task 3.5.32: Use evaluator's executePropertyRead instead of adapter
			if objVal.HasProperty(memberName) {
				propValue := objVal.ReadProperty(memberName, func(propInfo any) Value {
					return e.executePropertyRead(wrappedValue, propInfo, node, ctx)
				})
				// Check if result is an error
				if propValue != nil && propValue.Type() != "ERROR" {
					return propValue
				}
			}
		}

		// For method calls and complex cases, delegate to adapter
		return e.adapter.EvalNode(node)

	case "TYPE_META":
		// Enum type meta-value access (TColor.Red for unscoped enums)
		// The object is a TypeMetaValue wrapping an enum type
		enumMeta, ok := obj.(EnumTypeMetaDispatcher)
		if !ok {
			return e.newError(node, "internal error: TYPE_META value does not implement EnumTypeMetaDispatcher")
		}

		// Only handle if this is an enum type meta
		if !enumMeta.IsEnumTypeMeta() {
			// Not an enum - fall through to default case
			// Check for helper methods/properties
			helperResult := e.FindHelperMethod(obj, memberName)
			if helperResult != nil {
				if helperResult.Method != nil && len(helperResult.Method.Parameters) == 0 {
					return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
				}
				if helperResult.BuiltinSpec != "" {
					return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
				}
			}
			return e.newError(node, "member '%s' not found on value of type '%s'", memberName, obj.Type())
		}

		// Handle Low/High special properties
		normalizedMember := ident.Normalize(memberName)
		if normalizedMember == "low" {
			return &runtime.IntegerValue{Value: int64(enumMeta.EnumLow())}
		}
		if normalizedMember == "high" {
			return &runtime.IntegerValue{Value: int64(enumMeta.EnumHigh())}
		}

		// Try to get enum value by name (TColor.Red)
		if enumVal := enumMeta.GetEnumValue(memberName); enumVal != nil {
			return enumVal
		}

		return e.newError(node, "enum value '%s' not found in enum type", memberName)

	case "NIL":
		// Task 3.5.90: Nil object handling (Mode 9)
		// Typed nil values can access class variables, but not instance members.

		// Try NilAccessor interface to get typed class name
		nilVal, ok := obj.(NilAccessor)
		if !ok {
			// Untyped nil - always error
			return e.newError(node, "Object not instantiated")
		}

		typedClassName := nilVal.GetTypedClassName()
		if typedClassName == "" {
			// Untyped nil - always error
			return e.newError(node, "Object not instantiated")
		}

		// Typed nil: Try to look up class variable via adapter
		// The adapter can access the class registry and lookup class variables
		// Delegate to adapter for class variable lookup with proper static type
		return e.adapter.EvalNode(node)

	// Note: RECORD case removed - now handled by type assertion before switch
	// This avoids issues with RecordValue.Type() returning specific type names like "TPoint"
	// instead of the generic "RECORD" string

	case "ENUM":
		// Task 3.5.89: Enum value properties (Mode 10)
		// Pattern: enumVal.Value
		// Handle .Value property directly via EnumAccessor interface
		enumVal, ok := obj.(EnumAccessor)
		if !ok {
			// Task 3.5.101: Direct error instead of EvalNode delegation
			// If Type() returns "ENUM" but EnumAccessor interface is not implemented,
			// this is an internal inconsistency that should not happen
			return e.newError(node, "internal error: ENUM value does not implement EnumAccessor interface")
		}

		// Handle built-in .Value property
		if ident.Equal(memberName, "Value") {
			return &runtime.IntegerValue{Value: int64(enumVal.GetOrdinal())}
		}

		// Task 3.5.101: Check for helper methods/properties on enums
		// Other properties like .Name, .ToString are handled by helper methods
		helperResult := e.FindHelperMethod(obj, memberName)
		if helperResult != nil {
			// Found a helper method - check if it's parameterless for auto-invoke
			if helperResult.Method != nil && len(helperResult.Method.Parameters) == 0 {
				// Auto-invoke parameterless helper method
				return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
			}
			if helperResult.BuiltinSpec != "" {
				// Builtin helper - auto-invoke
				return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
			}
		}

		// Task 3.5.37: Check for helper properties - native handling
		helpers := e.getHelpersForValue(obj)
		for idx := len(helpers) - 1; idx >= 0; idx-- {
			helper := helpers[idx]
			if propInfo, ownerHelperAny, found := helper.GetPropertyAny(memberName); found && propInfo != nil {
				// Found a helper property - use native executeHelperPropertyRead
				pInfo, ok := propInfo.(*types.PropertyInfo)
				if ok {
					ownerHelper, _ := ownerHelperAny.(HelperInfo)
					return e.executeHelperPropertyRead(ownerHelper, pInfo, obj, node, ctx)
				}
			}
		}

		// Task 3.5.101: Unknown enum member - return proper error
		return e.newError(node, "member '%s' not found on enum value", memberName)

	default:
		// Task 3.5.101: Handle helper methods/properties for other types
		// Types like STRING, INTEGER, FLOAT, BOOLEAN, ARRAY may have helper extensions

		// Check for helper methods first
		helperResult := e.FindHelperMethod(obj, memberName)
		if helperResult != nil {
			// Found a helper method - check if it's parameterless for auto-invoke
			if helperResult.Method != nil && len(helperResult.Method.Parameters) == 0 {
				// Auto-invoke parameterless helper method
				return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
			}
			if helperResult.BuiltinSpec != "" {
				// Builtin helper - auto-invoke
				return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
			}
			// Helper method has parameters - needs to be called explicitly
			// This case should be handled by VisitMethodCallExpression, not member access
		}

		// Task 3.5.37: Check for helper properties - native handling
		helpers := e.getHelpersForValue(obj)
		for idx := len(helpers) - 1; idx >= 0; idx-- {
			helper := helpers[idx]
			if propInfo, ownerHelperAny, found := helper.GetPropertyAny(memberName); found && propInfo != nil {
				// Found a helper property - use native executeHelperPropertyRead
				pInfo, ok := propInfo.(*types.PropertyInfo)
				if ok {
					ownerHelper, _ := ownerHelperAny.(HelperInfo)
					return e.executeHelperPropertyRead(ownerHelper, pInfo, obj, node, ctx)
				}
			}
		}

		// Task 3.5.101: Return proper error for unsupported member access
		return e.newError(node, "member '%s' not found on value of type '%s'", memberName, obj.Type())
	}
}
