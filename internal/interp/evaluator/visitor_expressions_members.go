package evaluator

import (
	"strings"

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

	expectedTypeKind := ""
	if e.SemanticInfo() != nil {
		if typeAnnot := e.SemanticInfo().GetType(node); typeAnnot != nil {
			if resolvedType, err := e.ResolveTypeFromAnnotation(typeAnnot); err == nil && resolvedType != nil {
				expectedTypeKind = resolvedType.TypeKind()
			}
		}
	}
	wantMethodPointer := expectedTypeKind == "FUNCTION_POINTER" || expectedTypeKind == "METHOD_POINTER"

	// JSON namespace bare access (JSON.NewObject / JSON.NewArray, invoked without
	// parentheses) must be handled before `JSON` is evaluated as an identifier.
	if e.isJSONNamespaceObject(node.Object, ctx) {
		return e.evalJSONNamespaceCall(node.Member.Value, nil, node, ctx)
	}

	// Unit-qualified access (UnitName.Symbol) should not evaluate the unit identifier.
	if identObj, ok := node.Object.(*ast.Identifier); ok {
		if _, exists := ctx.Env().Get(identObj.Value); !exists && e.UnitRegistry() != nil {
			if _, exists := e.UnitRegistry().GetUnit(identObj.Value); exists {
				if valRaw, ok := ctx.Env().Get(node.Member.Value); ok {
					if val, ok := valRaw.(Value); ok {
						return val
					}
				}
				return e.newError(node, "qualified name '%s.%s' cannot be used as a value (functions must be called)", identObj.Value, node.Member.Value)
			}
		}
	}

	obj := e.Eval(node.Object, ctx)
	if isError(obj) {
		return obj
	}
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}
	if refVal, isRef := obj.(ReferenceAccessor); isRef {
		deref, err := refVal.Dereference()
		if err != nil {
			if raised, handled := e.raiseBoundExceededError(err, ctx); handled {
				return raised
			}
			return e.newError(node, "failed to dereference: %s", err.Error())
		}
		obj = deref
	}

	memberName := node.Member.Value

	// Member access on a JSON value (v.foo, v.length) yields another JSON value.
	if isJSONBoxed(obj) {
		return e.evalJSONValueMember(jsonValueOf(obj), memberName)
	}

	// Associative array parameterless members (a.Keys, a.Length, a.Count, a.Clear).
	if assoc, ok := obj.(*runtime.AssociativeArrayValue); ok {
		if result, handled := e.evalAssociativeArrayMethod(assoc, memberName, nil, node); handled {
			return result
		}
	}

	// Alias/static-type helper binding: the analyzer records the receiver's
	// static type when a helper resolved against it (strict helper semantics
	// dispatch on the declared type, not the dynamic one).
	if node.Member != nil && e.SemanticInfo() != nil {
		if annot := e.SemanticInfo().GetType(node.Member); annot != nil && strings.HasPrefix(annot.Name, "__helper_receiver:") {
			target := strings.TrimPrefix(annot.Name, "__helper_receiver:")
			if helpersAny := e.typeSystem.LookupHelpers(ident.Normalize(target)); helpersAny != nil {
				for _, helper := range orderedHelpersForLookup(convertToHelperInfoSlice(helpersAny)) {
					if helperResult := e.findHelperMethodInHelper(helper, memberName); helperResult != nil {
						if zeroArg := zeroArgHelperOverload(helperResult); zeroArg != nil && helperResult.BuiltinSpec == "" {
							callResult := *helperResult
							callResult.Method = zeroArg
							return e.CallHelperMethod(&callResult, obj, []Value{}, node, ctx)
						}
					}
				}
			}
		}
	}

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

		if rec, ok := obj.(*runtime.RecordValue); ok && rec.RecordType != nil {
			recordTypeRaw := e.typeSystem.LookupRecord(rec.RecordType.Name)
			if recordType, ok := recordTypeRaw.(*RecordTypeValue); ok {
				normalizedMember := ident.Normalize(memberName)
				if val, found := recordType.Constants[normalizedMember]; found {
					return val
				}
				if val, found := recordType.ClassVars[normalizedMember]; found {
					return val
				}
				if recordType.HasStaticMethod(memberName) {
					methodCall := &ast.MethodCallExpression{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{Token: node.Token},
						},
						Object:    node.Object,
						Method:    node.Member,
						Arguments: []ast.Expression{},
					}
					return e.VisitMethodCallExpression(methodCall, ctx)
				}
			}
		}

		if recVal.HasRecordProperty(memberName) {
			if rec, ok := obj.(*runtime.RecordValue); ok {
				propDesc := rec.LookupProperty(memberName)
				if propDesc != nil {
					if propInfo, ok := propDesc.Impl.(*types.RecordPropertyInfo); ok {
						return e.executeRecordPropertyRead(obj, propInfo, node, ctx)
					}
				}
			}
			return e.newError(node, "property '%s' not found in record '%s'", memberName, recVal.GetRecordTypeName())
		}

		// Helper properties
		if helper, propInfo := e.FindHelperProperty(obj, memberName); propInfo != nil {
			return e.executeHelperPropertyRead(helper, propInfo, obj, node, ctx)
		}

		// Helper methods (parameterless auto-invoke)
		if !isCurrentHelperMethod(ctx, memberName) {
			helperResult := e.FindHelperMethod(obj, memberName)
			if helperResult != nil {
				if helperResult.Method != nil && helperASTMethodEffectiveParamCount(helperResult.Method) == 0 {
					return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
				}
				if helperResult.BuiltinSpec != "" {
					return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
				}
			}
		}

		// Helper class consts/vars accessed through a record instance
		if val, found := e.findHelperClassMember(obj, memberName); found {
			return val
		}

		return e.newError(node, "field '%s' not found in record '%s'", memberName, recVal.GetRecordTypeName())
	}

	// Route based on object type
	switch obj.Type() {
	case "OBJECT":
		// Object instance. Helpers are checked before TObject built-ins so user
		// helpers can deliberately override members such as ClassName.
		objVal, ok := obj.(ObjectValue)
		if !ok {
			return e.newError(node, "internal error: OBJECT value does not implement ObjectValue interface")
		}

		// Accessing members of an explicitly freed object is a catchable runtime
		// error, reported at the member's position. Objects reclaimed by the
		// reference counter are exempt: go-dws releases eagerly in places where
		// DWScript keeps objects alive, so only explicit Free/Destroy is enforced.
		if objInst, ok := obj.(*runtime.ObjectInstance); ok && objInst.ExplicitlyFreed {
			return e.newError(node.Member, "Object already destroyed")
		}

		helpers := orderedHelpersForLookup(e.getHelpersForValue(obj))
		for _, helper := range helpers {
			for name, value := range helper.GetClassConsts() {
				if ident.Equal(name, memberName) {
					return value
				}
			}
			for name, value := range helper.GetClassVars() {
				if ident.Equal(name, memberName) {
					return value
				}
			}
			if propInfo, ownerHelperAny, found := helper.GetPropertyAny(memberName); found && propInfo != nil {
				if pInfo, ok := propInfo.(*types.PropertyInfo); ok {
					ownerHelper, _ := ownerHelperAny.(HelperInfo)
					return e.executeHelperPropertyRead(ownerHelper, pInfo, obj, node, ctx)
				}
			}
			if !isCurrentHelperMethod(ctx, memberName) {
				helperResult := e.findHelperMethodInHelper(helper, memberName)
				if helperResult != nil {
					if helperResult.Method != nil && helperASTMethodEffectiveParamCount(helperResult.Method) == 0 {
						return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
					}
					if helperResult.BuiltinSpec != "" {
						return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
					}
				}
			}
		}

		// Built-in TObject properties (ClassName, ClassType).
		// A user-declared ClassName method callable with zero arguments hides
		// the builtin (falls through to the method auto-invoke below).
		if ident.Equal(memberName, "ClassName") && !e.userMethodHidesBuiltin(obj, memberName) {
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
		if ident.Equal(memberName, "ClassParent") {
			// ClassParent is callable on an instance too; return the parent class
			// reference, or nil for a root class (TObject).
			if meta := e.getClassMetadataFromValue(obj); meta != nil && meta.Parent != nil {
				return e.makeClassValue(node, meta.Parent.Name)
			}
			return &runtime.NilValue{}
		}

		// Property access (with recursion protection)
		propCtx := ctx.PropContext()
		if propCtx == nil || (!propCtx.InPropertyGetter && !propCtx.InPropertySetter) {
			if objVal.HasProperty(memberName) {
				propValue := objVal.ReadProperty(memberName, func(propInfo any) Value {
					return e.executePropertyRead(obj, propInfo, node, ctx)
				})
				if propValue != nil {
					return propValue
				}
			}
		}

		// Field access (resolved against the static class of the object
		// expression, which matters for shadowed fields)
		if fieldValue := getFieldWithStaticClass(objVal, memberName, e.staticClassNameOf(node.Object, ctx)); fieldValue != nil {
			return fieldValue
		}

		// Class variable access
		if classVarValue, found := objVal.GetClassVar(memberName); found {
			return classVarValue
		}

		// Class constant access
		if classMetaVal, ok := obj.(ClassMetaProvider); ok {
			if constValue, found := classMetaVal.GetClassConstantBySpec(memberName); found {
				return constValue
			}
		}

		// Method access: auto-invoke if parameterless, else return function pointer
		if objVal.HasMethod(memberName) {
			if wantMethodPointer {
				if methodDecl := objVal.GetMethodDecl(memberName); methodDecl != nil {
					return e.createFunctionPointerFromDecl(methodDecl, obj, ctx)
				}
				return e.newError(node, "method '%s' not found", memberName)
			}

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

		// Class (static) method accessed through an instance. DWScript allows
		// class methods to be invoked on instances; dispatch through the method-call
		// path (which resolves class methods and binds the metaclass as Self).
		if classMethodDecl, ok := objVal.GetClassMethodDecl(memberName).(*ast.FunctionDecl); ok && classMethodDecl != nil {
			// A class method observes Self as the metaclass, not the instance it was
			// reached through. Bind the receiver's class value so a class method using
			// Self (e.g. Self.Create or returning Self) behaves the same as the
			// TClass.Method() path. The zero-parameter auto-invoke below already routes
			// through VisitMethodCallExpression, which resolves the class value itself.
			classSelf := e.classSelfForInstance(objVal, obj)
			if wantMethodPointer {
				return e.createFunctionPointerFromDecl(classMethodDecl, classSelf, ctx)
			}
			if len(classMethodDecl.Parameters) == 0 {
				methodCall := &ast.MethodCallExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{Token: node.Token},
					},
					Object:    node.Object,
					Method:    node.Member,
					Arguments: []ast.Expression{},
				}
				return e.VisitMethodCallExpression(methodCall, ctx)
			}
			return e.createFunctionPointerFromDecl(classMethodDecl, classSelf, ctx)
		}

		// Helper methods (parameterless auto-invoke)
		if !isCurrentHelperMethod(ctx, memberName) {
			helperResult := e.FindHelperMethod(obj, memberName)
			if helperResult != nil {
				if helperResult.Method != nil && helperASTMethodEffectiveParamCount(helperResult.Method) == 0 {
					return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
				}
				if helperResult.BuiltinSpec != "" {
					return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
				}
			}
		}

		return e.newError(node, "field '%s' not found in class '%s'", memberName, objVal.ClassName())

	case "INTERFACE":
		// Interface instance: verify member exists, access underlying object
		ifaceVal, ok := obj.(InterfaceInstanceValue)
		if !ok {
			return e.newError(node, "internal error: INTERFACE value does not implement InterfaceInstanceValue interface")
		}

		underlying := ifaceVal.GetUnderlyingObjectValue()
		if underlying == nil {
			return e.newError(node.Member, "Interface is nil")
		}

		// Property access via interface metadata
		if accessor, ok := obj.(PropertyAccessor); ok {
			if propDesc := accessor.LookupProperty(memberName); propDesc != nil {
				if objVal, ok := underlying.(ObjectValue); ok {
					return e.executePropertyRead(objVal, propDesc.Impl, node, ctx)
				}
				return e.newError(node, "internal error: interface underlying value does not implement ObjectValue")
			}
		}

		// Method access via underlying object
		if ifaceVal.HasInterfaceMethod(memberName) {
			if objVal, ok := underlying.(ObjectValue); ok {
				if wantMethodPointer {
					if methodDecl := objVal.GetMethodDecl(memberName); methodDecl != nil {
						return e.createFunctionPointerFromDecl(methodDecl, underlying, ctx)
					}
					return e.newError(node, "method '%s' not found", memberName)
				}

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
		// Metaclass access: ClassName, ClassParent, ClassType, class vars/consts
		classMetaVal, ok := obj.(ClassMetaValue)
		if !ok {
			return e.newError(node, "internal error: %s value does not implement ClassMetaValue interface", obj.Type())
		}
		return e.resolveClassMetaMember(obj, classMetaVal, memberName, node, ctx)

	case "RECORD_TYPE":
		// Static record access: constants, class vars, static methods
		recTypeVal, ok := obj.(*RecordTypeValue)
		if !ok {
			return e.newError(node, "internal error: RECORD_TYPE value is not *RecordTypeValue")
		}

		normalizedMember := ident.Normalize(memberName)

		// Constants
		if val, found := recTypeVal.Constants[normalizedMember]; found {
			return val
		}

		// Class variables
		if val, found := recTypeVal.ClassVars[normalizedMember]; found {
			return val
		}

		if recTypeVal.RecordType != nil && recTypeVal.RecordType.Properties != nil {
			if propInfo, found := recTypeVal.RecordType.Properties[normalizedMember]; found {
				if value, ok := readRecordTypePropertyValue(recTypeVal, propInfo); ok {
					return value
				}
				return e.newError(node, "property '%s' has no readable record type accessor", memberName)
			}
		}

		// Static methods (Class Methods)
		if recTypeVal.HasStaticMethod(memberName) {
			// Try parameterless auto-invoke first
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
		}

		// Helper class consts/vars/methods declared for the record type
		if helpersAny := e.typeSystem.LookupHelpers(ident.Normalize(recTypeVal.GetRecordTypeName())); helpersAny != nil {
			for _, helper := range orderedHelpersForLookup(convertToHelperInfoSlice(helpersAny)) {
				for name, v := range helper.GetClassConsts() {
					if ident.Equal(name, memberName) {
						return v
					}
				}
				for name, v := range helper.GetClassVars() {
					if ident.Equal(name, memberName) {
						return v
					}
				}
				if helperResult := e.findHelperMethodInHelper(helper, memberName); helperResult != nil {
					if zeroArg := zeroArgHelperOverload(helperResult); zeroArg != nil && helperResult.BuiltinSpec == "" {
						callResult := *helperResult
						callResult.Method = zeroArg
						return e.CallHelperMethod(&callResult, obj, []Value{}, node, ctx)
					}
				}
			}
		}

		return e.newError(node, "member '%s' not found in record type '%s'", memberName, recTypeVal.GetRecordTypeName())

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
			if ident.Equal(memberName, "ClassName") {
				return &runtime.StringValue{Value: objVal.ClassName()}
			}
			if ident.Equal(memberName, "ClassType") {
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

			// Field access uses the cast's static type, which matters for
			// shadowed fields (TBase(child).Field reads TBase's slot).
			if fieldValue := getFieldWithStaticClass(objVal, memberName, typeCastVal.GetStaticTypeName()); fieldValue != nil {
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
				if helperResult.Method != nil && helperASTMethodEffectiveParamCount(helperResult.Method) == 0 {
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
			// Metaclass reference (e.g. `TClass = class of TObject`): a class-of type
			// used as a value. Resolve class members (ClassName, ClassParent, ClassType,
			// class methods/consts) against the referenced class.
			if tmv, ok := obj.(*runtime.TypeMetaValue); ok {
				if classOf, ok := tmv.TypeInfo.(*types.ClassOfType); ok && classOf.ClassType != nil {
					classVal := e.makeClassValue(node, classOf.ClassType.Name)
					if !isError(classVal) {
						if classMetaVal, ok := classVal.(ClassMetaValue); ok {
							return e.resolveClassMetaMember(classVal, classMetaVal, memberName, node, ctx)
						}
					}
				}
			}
			if helper, propInfo := e.FindHelperProperty(obj, memberName); propInfo != nil {
				return e.executeHelperPropertyRead(helper, propInfo, obj, node, ctx)
			}
			// Helper class consts and class vars accessed through a type's metaclass
			// (e.g. `String.Hello`, `TMyArray.ByeBye`).
			if val, found := e.findHelperClassMember(obj, memberName); found {
				return val
			}
			// Helper class methods reachable through the target type's meta
			// value (e.g. `IMy.SayHello`, `TObject.ClassName`).
			for _, helper := range orderedHelpersForLookup(e.getHelpersForValue(obj)) {
				if !isCurrentHelperMethod(ctx, memberName) {
					helperResult := e.findHelperMethodInHelper(helper, memberName)
					if helperResult != nil {
						if helperResult.Method != nil && helperASTMethodEffectiveParamCount(helperResult.Method) == 0 {
							return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
						}
						if helperResult.Method == nil && helperResult.BuiltinSpec != "" {
							return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
						}
					}
				}
			}
			// Helper type referenced by name (e.g. TStringHelper.CountIt):
			// resolve class members against the helper itself. The type-meta
			// resolves to the helper's target type, so prefer the source
			// identifier for the helper name.
			helperName := ""
			if objIdent, ok := node.Object.(*ast.Identifier); ok {
				helperName = objIdent.Value
			} else if tmv, ok := obj.(*runtime.TypeMetaValue); ok {
				helperName = tmv.TypeName
			}
			if helperName != "" {
				if h := e.lookupMutableHelper(helperName); h != nil {
					if helperResult := e.findHelperMethodInHelper(h, memberName); helperResult != nil {
						if (helperResult.Method != nil && helperASTMethodEffectiveParamCount(helperResult.Method) == 0) ||
							(helperResult.Method == nil && helperResult.BuiltinSpec != "") {
							return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
						}
					}
					for name, v := range h.GetClassConsts() {
						if ident.Equal(name, memberName) {
							return v
						}
					}
					for name, v := range h.GetClassVars() {
						if ident.Equal(name, memberName) {
							return v
						}
					}
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
		// nil.Free is allowed (no-op)
		if ident.Equal(memberName, "Free") {
			return &runtime.NilValue{}
		}

		// Typed nil can access class vars, but not instance members
		if nilVal, ok := obj.(NilAccessor); ok {
			if typedClassName := nilVal.GetTypedClassName(); typedClassName != "" {
				// Look up class and access class variable
				if cv, err := e.typeSystem.CreateClassValue(typedClassName); err == nil && cv != nil {
					if classMetaVal, ok := cv.(ClassMetaValue); ok {
						if classVarValue, found := classMetaVal.GetClassVar(memberName); found {
							return classVarValue
						}
					}
				}
			}
		}

		// DWScript statically dispatches non-virtual instance methods, so a
		// parameterless one can be auto-invoked on a nil receiver (the error
		// only surfaces if the body dereferences Self).
		if classInfo := e.staticClassInfoForNilReceiver(obj, node.Object); classInfo != nil {
			if method := classInfo.LookupMethod(memberName); method != nil &&
				isNonVirtualInstanceMethod(classInfo, method) && len(method.Parameters) == 0 {
				return e.executeMethodWithClassInfo(obj, classInfo, method, nil, ctx)
			}
		}

		// Instance member access on nil is an error, reported at the member's position
		return e.newError(node.Member, "Object not instantiated")

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
		if !isCurrentHelperMethod(ctx, memberName) {
			helperResult := e.FindHelperMethod(obj, memberName)
			if helperResult != nil {
				if helperResult.Method != nil && helperASTMethodEffectiveParamCount(helperResult.Method) == 0 {
					return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
				}
				if helperResult.BuiltinSpec != "" {
					return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
				}
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

		helpers := orderedHelpersForLookup(e.getHelpersForValue(obj))
		for _, helper := range helpers {
			for name, value := range helper.GetClassConsts() {
				if ident.Equal(name, memberName) {
					return value
				}
			}
			for name, value := range helper.GetClassVars() {
				if ident.Equal(name, memberName) {
					return value
				}
			}
			if propInfo, ownerHelperAny, found := helper.GetPropertyAny(memberName); found && propInfo != nil {
				pInfo, ok := propInfo.(*types.PropertyInfo)
				if ok {
					ownerHelper, _ := ownerHelperAny.(HelperInfo)
					return e.executeHelperPropertyRead(ownerHelper, pInfo, obj, node, ctx)
				}
			}
			if !isCurrentHelperMethod(ctx, memberName) {
				helperResult := e.findHelperMethodInHelper(helper, memberName)
				if helperResult != nil {
					if zeroArg := zeroArgHelperOverload(helperResult); zeroArg != nil && helperResult.BuiltinSpec == "" {
						callResult := *helperResult
						callResult.Method = zeroArg
						return e.CallHelperMethod(&callResult, obj, []Value{}, node, ctx)
					}
					if helperResult.BuiltinSpec != "" {
						return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
					}
				}
			}
		}

		return e.newError(node, "member '%s' not found on value of type '%s'", memberName, obj.Type())
	}
}

// makeClassValue builds a class (metaclass) reference value for the named class,
// returning an error value if the class cannot be resolved.
func (e *Evaluator) makeClassValue(node ast.Node, className string) Value {
	classValAny, err := e.typeSystem.CreateClassValue(className)
	if err != nil {
		return e.newError(node, "%s", err.Error())
	}
	if cv, ok := classValAny.(Value); ok && cv != nil {
		return cv
	}
	return e.newError(node, "internal error: class value factory returned non-Value type for '%s'", className)
}

// resolveClassMetaMember resolves a member access against a class/metaclass value
// (built-in properties, class vars/consts/properties, constructors, class methods,
// nested classes, and helpers). It is shared by the CLASS/CLASSINFO member path and
// by the metaclass (`class of X`) member path.
func (e *Evaluator) resolveClassMetaMember(obj Value, classMetaVal ClassMetaValue, memberName string, node *ast.MemberAccessExpression, ctx *ExecutionContext) Value {
	// Built-in properties
	if ident.Equal(memberName, "ClassName") {
		return &runtime.StringValue{Value: classMetaVal.GetClassName()}
	}
	if ident.Equal(memberName, "ClassType") {
		return obj
	}
	if ident.Equal(memberName, "ClassParent") {
		if classInfo := classMetaVal.GetClassInfo(); classInfo != nil {
			if parent := classInfo.GetParent(); parent != nil {
				return e.makeClassValue(node, parent.GetName())
			}
		}
		// Root class (TObject): ClassParent is nil.
		return &runtime.NilValue{}
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
		classInfo := classMetaVal.GetClassInfo()
		if classInfo == nil {
			return e.newError(node, "class metadata unavailable for class property read")
		}
		typedPropInfo, ok := propInfo.(*types.PropertyInfo)
		if !ok {
			return e.newError(node, "invalid property info type for class property read")
		}
		return e.evalClassPropertyRead(classInfo, typedPropInfo, node, ctx)
	}); found {
		return result
	}

	// Instance properties backed by class vars/consts accessed via class name
	if classInfo := classMetaVal.GetClassInfo(); classInfo != nil {
		if propDesc := classInfo.LookupProperty(memberName); propDesc != nil {
			if propInfo, ok := propDesc.Impl.(*types.PropertyInfo); ok && !propInfo.IsClassProperty && propInfo.ReadKind == types.PropAccessField {
				if val, found := classMetaVal.GetClassVar(propInfo.ReadSpec); found {
					return val
				}
				if val, found := classMetaVal.GetClassConstant(propInfo.ReadSpec); found {
					return val
				}
			}
		}
	}

	// Constructors: auto-invoke without parentheses. Construct directly from the
	// resolved metaclass value rather than re-evaluating node.Object — when this helper
	// is reached via the `class of X` (TYPE_META) delegation, node.Object evaluates to a
	// TypeMetaValue, not the class reference, which would break construction.
	if classMetaVal.HasConstructor(memberName) {
		result, invoked := classMetaVal.InvokeConstructor(memberName, func(methodDecl any) Value {
			return e.callClassConstructor(classMetaVal, memberName, []Value{}, node, ctx)
		})
		if invoked {
			return result
		}
	}
	if ident.Equal(memberName, "Create") && classMetaVal.GetClassInfo() != nil {
		return e.callClassConstructor(classMetaVal, memberName, []Value{}, node, ctx)
	}

	// Class methods: auto-invoke if parameterless, else return function pointer
	if classMetaVal.HasClassMethod(memberName) {
		// Try parameterless auto-invoke
		result, invoked := classMetaVal.InvokeParameterlessClassMethod(memberName, func(methodDecl any) Value {
			return e.executeClassMethodDirect(classMetaVal, methodDecl, nil, node, ctx)
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

	if !isCurrentHelperMethod(ctx, memberName) {
		if helperResult := e.FindHelperMethod(obj, memberName); helperResult != nil {
			if helperResult.Method != nil && helperASTMethodEffectiveParamCount(helperResult.Method) == 0 {
				return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
			}
			if helperResult.BuiltinSpec != "" {
				return e.CallHelperMethod(helperResult, obj, []Value{}, node, ctx)
			}
		}
	}

	return e.newError(node, "member '%s' not found in class '%s'", memberName, classMetaVal.GetClassName())
}

// classSelfForInstance resolves the class (metaclass) value for an instance, so a class
// method reached through that instance binds Self to the class rather than the object.
// If the class value cannot be resolved it falls back to the instance, preserving the
// previous behavior instead of failing the member access.
func (e *Evaluator) classSelfForInstance(objVal ObjectValue, obj Value) Value {
	if classValAny, err := e.typeSystem.CreateClassValue(objVal.ClassName()); err == nil {
		if cv, ok := classValAny.(Value); ok && cv != nil {
			return cv
		}
	}
	return obj
}
