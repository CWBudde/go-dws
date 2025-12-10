package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// Member access visitor methods for field, property, and method references.

// VisitMemberAccessExpression evaluates member access (obj.field, obj.method).
//
// 11 ACCESS MODES (evaluated in order):
//  1. UNIT-QUALIFIED: Math.PI, System.Print
//  2. STATIC CLASS: TMyClass.ClassVar, TMyClass.Create
//  3. ENUM TYPE: TColor.Red, TColor.Low/High
//  4. RECORD TYPE STATIC: TPoint.cOrigin
//  5. RECORD INSTANCE: point.X, point.GetLength()
//  6. CLASS/METACLASS: classVar.Create()
//  7. INTERFACE: intf.Method, intf.Property
//  8. TYPE CAST: TBase(child).ClassVar
//  9. NIL OBJECT: nil.ClassVar (allowed), nil.Field (error)
//
// 10. ENUM VALUE: enumVal.Value
// 11. OBJECT INSTANCE: obj.Field, obj.Property, obj.Method
//
// Key behaviors:
// - Parameterless methods auto-invoke without ()
// - All lookups are case-insensitive
// - Inheritance chain searched for class vars/consts/methods
// - Type helpers can extend any type with properties/methods
func (e *Evaluator) VisitMemberAccessExpression(node *ast.MemberAccessExpression, ctx *ExecutionContext) Value {
	if node.Object == nil {
		return e.newError(node, "member access missing object")
	}
	if node.Member == nil {
		return e.newError(node, "member access missing member")
	}

	obj := e.Eval(node.Object, ctx)
	if isError(obj) {
		return obj
	}

	memberName := node.Member.Value

	// Record instance check via type assertion (RecordValue.Type() returns specific names like "TPoint")
	if recVal, ok := obj.(RecordInstanceValue); ok {
		// Direct field access
		if fieldVal, found := recVal.GetRecordField(memberName); found {
			return fieldVal
		}

		// Method reference: parameterless auto-invokes, with params requires ()
		if recVal.HasRecordMethod(memberName) {
			methodDecl, found := recVal.GetRecordMethod(memberName)
			if !found {
				return e.newError(node, "internal error: method '%s' not retrievable", memberName)
			}
			if len(methodDecl.Parameters) > 0 {
				return e.newError(node,
					"method '%s' of record '%s' requires %d parameter(s); use parentheses to call",
					memberName, recVal.GetRecordTypeName(), len(methodDecl.Parameters))
			}
			return e.callRecordMethod(recVal, methodDecl, []Value{}, node, ctx)
		}

		if recVal.HasRecordProperty(memberName) {
			return e.newError(node, "property access on records not supported")
		}

		return e.newError(node, "field '%s' not found in record '%s'", memberName, recVal.GetRecordTypeName())
	}

	// Route based on object type
	switch obj.Type() {
	case "OBJECT":
		// Object instance: Properties -> Fields -> Class Variables -> Helpers -> Methods
		objVal, ok := obj.(ObjectValue)
		if !ok {
			return e.newError(node, "internal error: OBJECT value does not implement ObjectValue interface")
		}

		// Built-in TObject properties (ClassName, ClassType)
		if ident.Equal(memberName, "ClassName") {
			return &runtime.StringValue{Value: objVal.ClassName()}
		}
		if ident.Equal(memberName, "ClassType") {
			// GetClassType() returns classTypeProxy, convert to proper ClassValue
			className := objVal.ClassName()
			classVal, err := e.typeSystem.CreateClassValue(className)
			if err != nil {
				return e.newError(node, "%s", err.Error())
			}
			if val, ok := classVal.(Value); ok {
				return val
			}
			return e.newError(node, "internal error: ClassValue conversion failed")
		}

		// Property access (with recursion protection)
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

		// Field access
		if fieldValue := objVal.GetField(memberName); fieldValue != nil {
			return fieldValue
		}

		// Class variable access
		if classVarValue, found := objVal.GetClassVar(memberName); found {
			return classVarValue
		}

		// Helper properties
		if helper, propInfo := e.FindHelperProperty(obj, memberName); propInfo != nil {
			return e.executeHelperPropertyRead(helper, propInfo, obj, node, ctx)
		}

		// Method access: auto-invoke if parameterless, else return function pointer
		if objVal.HasMethod(memberName) {
			// Try parameterless auto-invoke first
			result, invoked := objVal.InvokeParameterlessMethod(memberName, func(methodDecl any) Value {
				// Create synthetic method call
				methodCall := &ast.MethodCallExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{Token: node.Token},
					},
					Object:    node.Object,
					Method:    node.Member,
					Arguments: []ast.Expression{},
				}
				return e.VisitMethodCallExpression(methodCall, ctx)
			})
			if invoked {
				return result
			}

			// Return function pointer for methods with parameters
			result, created := objVal.CreateMethodPointer(memberName, func(methodDecl any) Value {
				return e.createFunctionPointerFromDecl(methodDecl, obj, ctx)
			})
			if created {
				return result
			}
		}

		// Helper methods (parameterless auto-invoke)
		helperResult := e.FindHelperMethod(obj, memberName)
		if helperResult != nil {
			if helperResult.Method != nil && len(helperResult.Method.Parameters) == 0 {
				return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
			}
			if helperResult.BuiltinSpec != "" {
				return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
			}
		}

		return e.newError(node, "member '%s' not found on object of class '%s'", memberName, objVal.ClassName())

	case "INTERFACE":
		// Interface instance: verify member exists, access underlying object
		ifaceVal, ok := obj.(InterfaceInstanceValue)
		if !ok {
			return e.newError(node, "internal error: INTERFACE value does not implement InterfaceInstanceValue interface")
		}

		underlying := ifaceVal.GetUnderlyingObjectValue()
		if underlying == nil {
			return e.newError(node, "Interface is nil")
		}

		// Property access via underlying object
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

		// Method access via underlying object
		if ifaceVal.HasInterfaceMethod(memberName) {
			if objVal, ok := underlying.(ObjectValue); ok {
				// Try parameterless auto-invoke first
				result, invoked := objVal.InvokeParameterlessMethod(memberName, func(methodDecl any) Value {
					// Create synthetic method call - use original node.Object (the interface expression)
					methodCall := &ast.MethodCallExpression{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{Token: node.Token},
						},
						Object:    node.Object,
						Method:    node.Member,
						Arguments: []ast.Expression{},
					}
					return e.VisitMethodCallExpression(methodCall, ctx)
				})
				if invoked {
					return result
				}

				// Return function pointer for methods with parameters
				result, created := objVal.CreateMethodPointer(memberName, func(methodDecl any) Value {
					return e.createFunctionPointerFromDecl(methodDecl, underlying, ctx)
				})
				if created {
					return result
				}
			}
			return e.newError(node, "internal error: interface underlying value does not implement ObjectValue")
		}

		return e.newError(node, "member '%s' not found on interface '%s'", memberName, ifaceVal.InterfaceName())

	case "CLASS":
		// CLASS and CLASSINFO both implement ClassMetaValue
		fallthrough

	case "CLASSINFO":
		// Metaclass access: ClassName, ClassType, class vars/consts
		classMetaVal, ok := obj.(ClassMetaValue)
		if !ok {
			return e.newError(node, "internal error: %s value does not implement ClassMetaValue interface", obj.Type())
		}

		// Built-in properties
		if ident.Equal(memberName, "ClassName") {
			return &runtime.StringValue{Value: classMetaVal.GetClassName()}
		}
		if ident.Equal(memberName, "ClassType") {
			return obj
		}

		// Class variables and constants
		if val, found := classMetaVal.GetClassVar(memberName); found {
			return val
		}
		if val, found := classMetaVal.GetClassConstant(memberName); found {
			return val
		}

		// Class properties (class property Counter: Integer read FCounter)
		if result, found := classMetaVal.ReadClassProperty(memberName, func(propInfo any) Value {
			return e.adapter.EvalClassPropertyRead(classMetaVal.GetClassInfo(), propInfo, node)
		}); found {
			return result
		}

		// Constructors: auto-invoke without parentheses
		if classMetaVal.HasConstructor(memberName) {
			result, invoked := classMetaVal.InvokeConstructor(memberName, func(methodDecl any) Value {
				// Create synthetic method call and route to VisitMethodCallExpression
				methodCall := &ast.MethodCallExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{Token: node.Token},
					},
					Object:    node.Object,
					Method:    node.Member,
					Arguments: []ast.Expression{},
				}
				return e.VisitMethodCallExpression(methodCall, ctx)
			})
			if invoked {
				return result
			}
		}

		// Class methods: auto-invoke if parameterless, else return function pointer
		if classMetaVal.HasClassMethod(memberName) {
			// Try parameterless auto-invoke
			result, invoked := classMetaVal.InvokeParameterlessClassMethod(memberName, func(methodDecl any) Value {
				// Create synthetic method call and route to VisitMethodCallExpression
				methodCall := &ast.MethodCallExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{Token: node.Token},
					},
					Object:    node.Object,
					Method:    node.Member,
					Arguments: []ast.Expression{},
				}
				return e.VisitMethodCallExpression(methodCall, ctx)
			})
			if invoked {
				return result
			}

			// Return function pointer for class methods with parameters
			result, created := classMetaVal.CreateClassMethodPointer(memberName, func(methodDecl any) Value {
				return e.createFunctionPointerFromDecl(methodDecl, nil, ctx)
			})
			if created {
				return result
			}
		}

		// Nested class access
		if nestedClass := classMetaVal.GetNestedClass(memberName); nestedClass != nil {
			return nestedClass
		}

		return e.newError(node, "member '%s' not found in class '%s'", memberName, classMetaVal.GetClassName())

	case "TYPE_CAST":
		// Type cast: use static type for class var lookup (TBase(child).ClassVar)
		typeCastVal, ok := obj.(TypeCastAccessor)
		if !ok {
			return e.newError(node, "internal error: TYPE_CAST value does not implement TypeCastAccessor interface")
		}

		// Static class variable lookup (key behavior of type casts)
		if classVarValue, found := typeCastVal.GetStaticClassVar(memberName); found {
			return classVarValue
		}

		// Field and property access on wrapped object
		wrappedValue := typeCastVal.GetWrappedValue()
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

			// Method dispatch on wrapped object
			if objVal.HasMethod(memberName) {
				// Try parameterless auto-invoke
				result, invoked := objVal.InvokeParameterlessMethod(memberName, func(methodDecl any) Value {
					// Create synthetic method call - use original node.Object (the cast expression)
					// so that the method call sees the cast wrapper
					methodCall := &ast.MethodCallExpression{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{Token: node.Token},
						},
						Object:    node.Object,
						Method:    node.Member,
						Arguments: []ast.Expression{},
					}
					return e.VisitMethodCallExpression(methodCall, ctx)
				})
				if invoked {
					return result
				}

				// Return function pointer for methods with parameters
				result, created := objVal.CreateMethodPointer(memberName, func(methodDecl any) Value {
					return e.createFunctionPointerFromDecl(methodDecl, wrappedValue, ctx)
				})
				if created {
					return result
				}
			}
		}

		// Helper methods on wrapped value
		if wrappedValue != nil {
			helperResult := e.FindHelperMethod(wrappedValue, memberName)
			if helperResult != nil {
				if helperResult.Method != nil && len(helperResult.Method.Parameters) == 0 {
					return e.CallHelperMethod(helperResult, wrappedValue, []Value{}, node, ctx)
				}
				if helperResult.BuiltinSpec != "" {
					return e.CallHelperMethod(helperResult, wrappedValue, []Value{}, node, ctx)
				}
			}
		}

		return e.newError(node, "member '%s' not found", memberName)

	case "TYPE_META":
		// Enum type meta access (TColor.Red, TColor.Low/High)
		enumMeta, ok := obj.(EnumTypeMetaDispatcher)
		if !ok {
			return e.newError(node, "internal error: TYPE_META value does not implement EnumTypeMetaDispatcher")
		}

		// Non-enum type meta: check helpers
		if !enumMeta.IsEnumTypeMeta() {
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

		// Low/High properties
		normalizedMember := ident.Normalize(memberName)
		if normalizedMember == "low" {
			return &runtime.IntegerValue{Value: int64(enumMeta.EnumLow())}
		}
		if normalizedMember == "high" {
			return &runtime.IntegerValue{Value: int64(enumMeta.EnumHigh())}
		}

		// Enum value by name
		if enumVal := enumMeta.GetEnumValue(memberName); enumVal != nil {
			return enumVal
		}

		return e.newError(node, "enum value '%s' not found in enum type", memberName)

	case "NIL":
		// Typed nil can access class vars, but not instance members
		nilVal, ok := obj.(NilAccessor)
		if !ok {
			return e.newError(node, "Object not instantiated")
		}
		typedClassName := nilVal.GetTypedClassName()
		if typedClassName == "" {
			return e.newError(node, "Object not instantiated")
		}

		// nil.Free is allowed (no-op)
		if ident.Equal(memberName, "Free") {
			return &runtime.NilValue{}
		}

		// Look up class and access class variable
		classMetaVal := e.adapter.LookupClassByName(typedClassName)
		if classMetaVal != nil {
			if classVarValue, found := classMetaVal.GetClassVar(memberName); found {
				return classVarValue
			}
		}

		// Instance member access on nil is an error
		return e.newError(node, "Object not instantiated")

	case "ENUM":
		// Enum value properties (.Value, helpers)
		enumVal, ok := obj.(EnumAccessor)
		if !ok {
			return e.newError(node, "internal error: ENUM value does not implement EnumAccessor interface")
		}

		// Built-in .Value property
		if ident.Equal(memberName, "Value") {
			return &runtime.IntegerValue{Value: int64(enumVal.GetOrdinal())}
		}

		// Helper methods (.Name, .ToString, etc.)
		helperResult := e.FindHelperMethod(obj, memberName)
		if helperResult != nil {
			if helperResult.Method != nil && len(helperResult.Method.Parameters) == 0 {
				return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
			}
			if helperResult.BuiltinSpec != "" {
				return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
			}
		}

		// Helper properties
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

		return e.newError(node, "member '%s' not found on enum value", memberName)

	default:
		// Other types (STRING, INTEGER, FLOAT, BOOLEAN, ARRAY): check helpers

		// Helper methods (parameterless auto-invoke)
		helperResult := e.FindHelperMethod(obj, memberName)
		if helperResult != nil {
			if helperResult.Method != nil && len(helperResult.Method.Parameters) == 0 {
				return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
			}
			if helperResult.BuiltinSpec != "" {
				return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
			}
		}

		// Helper properties
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

		return e.newError(node, "member '%s' not found on value of type '%s'", memberName, obj.Type())
	}
}
