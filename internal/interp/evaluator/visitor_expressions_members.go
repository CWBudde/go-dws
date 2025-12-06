package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// This file contains visitor methods for member access expression AST nodes.
// Member access includes field access, property access, and method references on objects.

// VisitMemberAccessExpression evaluates member access (obj.field, obj.method).
//
// **11 DISTINCT ACCESS MODES** (evaluated in this order):
//
// === SIMPLE MODES ===
//
// **1. UNIT-QUALIFIED ACCESS** (UnitName.Symbol)
//   - Pattern: `Math.PI`, `System.Print`
//   - Evaluation order:
//     a. Check if left side is a registered unit name (via unitRegistry)
//     b. Try to resolve as qualified variable/constant (ResolveQualifiedVariable)
//     c. If not a variable, it might be a function reference (handled in VisitCallExpression)
//   - Error: "qualified name 'Unit.Symbol' cannot be used as a value (functions must be called)"
//
// **2. STATIC CLASS ACCESS** (TClass.Member)
//   - Pattern: `TMyClass.ClassVar`, `TMyClass.Create`, `TMyClass.ClassName`
//   - Lookup order (case-insensitive):
//     a. Built-in properties: `ClassName` (string), `ClassType` (metaclass reference)
//     b. Class variables (inherited from parent classes)
//     c. Class constants (lazy evaluation with caching)
//     d. Class properties (if IsClassProperty or uses class-level read specs)
//     e. Constructors (auto-invoke with 0 arguments if no parentheses)
//     f. Class methods (static methods)
//   - Error: "member 'X' not found in class 'Y'"
//
// **3. ENUM TYPE ACCESS** (TColor.Red, TColor.Low, TColor.High)
//   - Pattern: `TColor.Red`, `TMyEnum.Low`, `TMyEnum.High`
//   - Lookup via TypeSystem
//   - For scoped enums:
//     a. Look up enum value in EnumType.Values (takes precedence over properties)
//     b. Check for special properties: `Low` (lowest ordinal), `High` (highest ordinal)
//   - For unscoped enums: also check environment for value name
//   - Returns: EnumValue or IntegerValue (for Low/High)
//   - Error: "enum value 'X' not found in enum 'Y'"
//
// **4. RECORD TYPE STATIC ACCESS** (TPoint.cOrigin, TPoint.Count)
//   - Pattern: `TPoint.cOrigin`, `TRecord.ClassMethod()`
//   - Lookup order (case-insensitive):
//     a. Constants (RecordTypeValue.Constants)
//     b. Class variables (RecordTypeValue.ClassVars)
//     c. Class methods (RecordTypeValue.ClassMethods)
//   - Error: "member 'X' not found in record type 'Y'"
//
// **5. RECORD INSTANCE ACCESS** (record.Field, record.Method)
//   - Pattern: `point.X`, `point.GetLength()`, `point.Prop`
//   - Object type: RecordValue
//   - Lookup order (case-insensitive):
//     a. Direct field access (RecordValue.Fields)
//     b. Properties (RecordType.Properties)
//     c. Instance methods (RecordValue.Methods)
//     d. Class methods (from RecordTypeValue, accessible via instance)
//     e. Constants (from RecordTypeValue, accessible via instance)
//     f. Class variables (from RecordTypeValue, accessible via instance)
//     g. Helper properties
//   - Error: "field 'X' not found in record 'Y'"
//
// **10. ENUM VALUE PROPERTIES** (enumVal.Value)
//   - Pattern: `TColor.Red.Value` (returns ordinal as integer)
//   - Object type: EnumValue
//   - Supported properties:
//     a. `.Value` (case-insensitive): returns OrdinalValue as IntegerValue
//     b. `.ToString`: handled by helpers (if available)
//   - Fallback: Check helpers for additional properties
//
// === COMPLEX MODES ===
//
// **6. CLASS/METACLASS ACCESS** (ClassInfoValue/ClassValue.Member)
//   - Pattern: When a class name is evaluated to ClassInfoValue or ClassValue
//   - Example: `var c := TMyClass; c.Create()`
//   - Lookup order (same as static class access #2):
//     a. Built-in properties: `ClassName`, `ClassType`
//     b. Class variables, constants, properties, constructors, class methods
//   - Returns: String/ClassValue/field value/method pointer
//
// **7. INTERFACE INSTANCE ACCESS** (interface.Method, interface.Property)
//   - Pattern: `intfVar.Hello`, `intfVar.SomeMethod`
//   - Object type: InterfaceInstance
//   - Validation: Verify member exists in interface definition (HasMethod)
//   - For methods:
//     - Look up implementation in underlying object's class
//     - Return FunctionPointerValue bound to underlying object (NO auto-invoke)
//     - Enables method delegate assignment: `var h : procedure := i.Hello;`
//   - For properties/fields: delegate to underlying object
//   - Error: "Interface is nil" or "method 'X' declared in interface 'Y' but not implemented"
//
// **8. TYPE CAST VALUE HANDLING** (TBase(child).ClassVar)
//   - Pattern: Accessing members through a type cast expression
//   - Object type: TypeCastValue
//   - Purpose: Class variables use static type, not runtime type
//   - Unwraps to actual object and continues evaluation with static type context
//
// **9. NIL OBJECT HANDLING** (nil.ClassVar)
//   - Pattern: `var o: TMyClass := nil; o.ClassVar`
//   - Object type: NilValue (with ClassType field) or nil evaluation result
//   - Special case: Accessing class variables on nil instances is allowed
//   - Error: "Object not instantiated" (for instance members)
//
// **11. OBJECT INSTANCE ACCESS** (obj.Field, obj.Method, obj.Property)
//   - Pattern: `myObj.Name`, `myObj.GetValue()`, `myObj.Count`
//   - Object type: ObjectInstance
//   - Built-in properties (inherited from TObject, case-insensitive):
//     a. `ClassName`: returns obj.Class.Name (runtime type)
//     b. `ClassType`: returns ClassValue (metaclass for runtime type)
//   - Lookup order (case-insensitive):
//     a. Properties - takes precedence over fields
//     b. Direct field access - instance fields
//     c. Class variables - accessible from instance
//     d. Class constants - accessible from instance
//     e. Instance methods
//     f. Class methods
//     g. Helper properties
//   - Error: "field 'X' not found in class 'Y'"
//
// **SPECIAL BEHAVIORS**:
// - **Auto-invocation**: Parameterless methods/properties auto-invoke when accessed without ()
// - **Case-insensitive**: All name lookups are case-insensitive (DWScript spec)
// - **Inheritance**: Class variables, constants, properties, methods searched up hierarchy
// - **Helper support**: Type helpers can add properties/methods to any type
// - **Function pointers**: Methods with parameters return FunctionPointerValue
// - **Lazy evaluation**: Class constants evaluated once and cached on first access
// - **Type safety**: Static types respected for class variable access through casts
func (e *Evaluator) VisitMemberAccessExpression(node *ast.MemberAccessExpression, ctx *ExecutionContext) Value {
	if node.Object == nil {
		return e.newError(node, "member access missing object")
	}
	if node.Member == nil {
		return e.newError(node, "member access missing member")
	}

	// Evaluate the object first
	obj := e.Eval(node.Object, ctx)
	if isError(obj) {
		return obj
	}

	memberName := node.Member.Value

	// Route based on object type
	switch obj.Type() {
	case "OBJECT":
		// Object instance access (Mode 11): obj.Field, obj.Property, obj.Method
		// Lookup order: Properties -> Fields -> Class Variables

		objVal, ok := obj.(ObjectValue)
		if !ok {
			return e.newError(node, "internal error: OBJECT value does not implement ObjectValue interface")
		}

		// Try property access first (with recursion protection)
		propCtx := ctx.PropContext()
		if propCtx == nil || (!propCtx.InPropertyGetter && !propCtx.InPropertySetter) {
			if objVal.HasProperty(memberName) {
				propValue := objVal.ReadProperty(memberName, func(propInfo any) Value {
					return e.executePropertyRead(obj, propInfo, node, ctx)
				})
				if propValue != nil && propValue.Type() != "ERROR" {
					return propValue
				}
			}
		}

		// Direct field access
		if fieldValue := objVal.GetField(memberName); fieldValue != nil {
			return fieldValue
		}

		// Class variable access
		if classVarValue, found := objVal.GetClassVar(memberName); found {
			return classVarValue
		}

		// Check for helper properties
		if helper, propInfo := e.FindHelperProperty(obj, memberName); propInfo != nil {
			return e.executeHelperPropertyRead(helper, propInfo, obj, node, ctx)
		}

		// Method or other member access via adapter
		return e.adapter.EvalNode(node)

	case "INTERFACE":
		// Interface instance access (Mode 7): intf.Method, intf.Property

		ifaceVal, ok := obj.(InterfaceInstanceValue)
		if !ok {
			return e.newError(node, "internal error: INTERFACE value does not implement InterfaceInstanceValue interface")
		}

		// Get underlying object - nil check is critical
		underlying := ifaceVal.GetUnderlyingObjectValue()
		if underlying == nil {
			return e.newError(node, "Interface is nil")
		}

		// Verify the member is part of the interface contract
		if ifaceVal.HasInterfaceProperty(memberName) {
			if objVal, ok := underlying.(ObjectValue); ok {
				propValue := objVal.ReadProperty(memberName, func(propInfo any) Value {
					return e.executePropertyRead(underlying, propInfo, node, ctx)
				})
				if propValue != nil && propValue.Type() != "ERROR" {
					return propValue
				}
				return e.newError(node, "failed to read property '%s' on interface '%s': %v", memberName, ifaceVal.InterfaceName(), propValue)
			}
			return e.newError(node, "internal error: interface underlying value does not implement ObjectValue")
		}

		// Method access or unknown member - delegate to adapter
		return e.adapter.EvalNode(node)

	case "CLASSINFO":
		// Metaclass access (Mode 6): ClassInfoValue.Member

		classMetaVal, ok := obj.(ClassMetaValue)
		if !ok {
			return e.newError(node, "internal error: %s value does not implement ClassMetaValue interface", obj.Type())
		}

		// Handle built-in properties
		if ident.Equal(memberName, "ClassName") {
			return &runtime.StringValue{Value: classMetaVal.GetClassName()}
		}
		if ident.Equal(memberName, "ClassType") {
			return obj
		}

		// Try class variables
		if val, found := classMetaVal.GetClassVar(memberName); found {
			return val
		}

		// Try class constants
		if val, found := classMetaVal.GetClassConstant(memberName); found {
			return val
		}

		// Complex cases (constructors, class methods, properties) via adapter
		return e.adapter.EvalNode(node)

	case "TYPE_CAST":
		// Type cast value handling (Mode 8): TBase(child).ClassVar
		// Uses static type from cast for class variable lookup, not runtime type

		typeCastVal, ok := obj.(TypeCastAccessor)
		if !ok {
			return e.newError(node, "internal error: TYPE_CAST value does not implement TypeCastAccessor interface")
		}

		// Try class variable lookup using the static type
		if classVarValue, found := typeCastVal.GetStaticClassVar(memberName); found {
			return classVarValue
		}

		// Get the wrapped value for further processing
		wrappedValue := typeCastVal.GetWrappedValue()

		// If wrapped value is an object, try field access and property reading
		if objVal, ok := wrappedValue.(ObjectValue); ok {
			if fieldValue := objVal.GetField(memberName); fieldValue != nil {
				return fieldValue
			}

			if objVal.HasProperty(memberName) {
				propValue := objVal.ReadProperty(memberName, func(propInfo any) Value {
					return e.executePropertyRead(wrappedValue, propInfo, node, ctx)
				})
				if propValue != nil && propValue.Type() != "ERROR" {
					return propValue
				}
			}
		}

		// Method calls and complex cases via adapter
		return e.adapter.EvalNode(node)

	case "NIL":
		// Nil object handling (Mode 9): typed nil values can access class variables

		nilVal, ok := obj.(NilAccessor)
		if !ok {
			return e.newError(node, "Object not instantiated")
		}

		typedClassName := nilVal.GetTypedClassName()
		if typedClassName == "" {
			return e.newError(node, "Object not instantiated")
		}

		// Typed nil: delegate to adapter for class variable lookup
		return e.adapter.EvalNode(node)

	case "RECORD":
		// Record instance access (Mode 5): record.Field, record.Method

		recVal, ok := obj.(RecordInstanceValue)
		if !ok {
			return e.newError(node, "internal error: RECORD value does not implement RecordInstanceValue interface")
		}

		// Direct field access - most common case
		if fieldVal, found := recVal.GetRecordField(memberName); found {
			return fieldVal
		}

		// Method reference via adapter
		if recVal.HasRecordMethod(memberName) {
			return e.adapter.EvalNode(node)
		}

		// Property access (rare, if supported)
		if recVal.HasRecordProperty(memberName) {
			return e.newError(node, "property access on records not supported")
		}

		// Member not found
		return e.newError(node, "field '%s' not found in record '%s'", memberName, recVal.GetRecordTypeName())

	case "ENUM":
		// Enum value properties (Mode 10): enumVal.Value

		enumVal, ok := obj.(EnumAccessor)
		if !ok {
			return e.newError(node, "internal error: ENUM value does not implement EnumAccessor interface")
		}

		// Handle built-in .Value property
		if ident.Equal(memberName, "Value") {
			return &runtime.IntegerValue{Value: int64(enumVal.GetOrdinal())}
		}

		// Check for helper methods (auto-invoke if parameterless)
		helperResult := e.FindHelperMethod(obj, memberName)
		if helperResult != nil {
			if helperResult.Method != nil && len(helperResult.Method.Parameters) == 0 {
				return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
			}
			if helperResult.BuiltinSpec != "" {
				return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
			}
		}

		// Check for helper properties
		helpers := e.getHelpersForValue(obj)
		for idx := len(helpers) - 1; idx >= 0; idx-- {
			helper := helpers[idx]
			if propInfo, ownerHelperAny, found := helper.GetPropertyAny(memberName); found && propInfo != nil {
				pInfo, ok := propInfo.(*types.PropertyInfo)
				if ok {
					ownerHelper, _ := ownerHelperAny.(HelperInfo)
					return e.executeHelperPropertyRead(ownerHelper, pInfo, obj, node, ctx)
				}
			}
		}

		// Unknown enum member
		return e.newError(node, "member '%s' not found on enum value", memberName)

	default:
		// Helper methods/properties for other types (STRING, INTEGER, FLOAT, BOOLEAN, ARRAY)

		// Check for helper methods (auto-invoke if parameterless)
		helperResult := e.FindHelperMethod(obj, memberName)
		if helperResult != nil {
			if helperResult.Method != nil && len(helperResult.Method.Parameters) == 0 {
				return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
			}
			if helperResult.BuiltinSpec != "" {
				return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
			}
		}

		// Check for helper properties
		helpers := e.getHelpersForValue(obj)
		for idx := len(helpers) - 1; idx >= 0; idx-- {
			helper := helpers[idx]
			if propInfo, ownerHelperAny, found := helper.GetPropertyAny(memberName); found && propInfo != nil {
				pInfo, ok := propInfo.(*types.PropertyInfo)
				if ok {
					ownerHelper, _ := ownerHelperAny.(HelperInfo)
					return e.executeHelperPropertyRead(ownerHelper, pInfo, obj, node, ctx)
				}
			}
		}

		// Member not found
		return e.newError(node, "member '%s' not found on value of type '%s'", memberName, obj.Type())
	}
}
