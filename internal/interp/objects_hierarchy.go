package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	pkgident "github.com/cwbudde/go-dws/pkg/ident"
)

// evalMemberAccess evaluates obj.field or TClass.Variable access.
// Handles instance field access, static class access, unit-qualified symbols, enum values, and record types.
func (i *Interpreter) evalMemberAccess(ma *ast.MemberAccessExpression) Value {
	// Check for static access patterns (TClass.Member, UnitName.Symbol, TEnum.Value)
	if ident, ok := ma.Object.(*ast.Identifier); ok {
		// Unit-qualified access: UnitName.Symbol
		if i.unitRegistry != nil {
			if _, exists := i.unitRegistry.GetUnit(ident.Value); exists {
				if val, err := i.ResolveQualifiedVariable(ident.Value, ma.Member.Value); err == nil {
					return val
				}
				return i.newErrorWithLocation(ma, "qualified name '%s.%s' cannot be used as a value (functions must be called)", ident.Value, ma.Member.Value)
			}
		}

		// Class static access: TClass.Variable/Method/Constructor
		classInfo := i.resolveClassInfoByName(ident.Value)
		if classInfo != nil {
			memberName := ma.Member.Value

			// Built-in class properties
			if pkgident.Equal(memberName, "ClassType") {
				return &ClassValue{ClassInfo: classInfo}
			}
			if pkgident.Equal(memberName, "ClassName") {
				return &StringValue{Value: classInfo.Name}
			}

			// Class variables, constants, and properties
			if classVarValue, ownerClass := classInfo.lookupClassVar(memberName); ownerClass != nil {
				return classVarValue
			}
			if constValue := i.getClassConstant(classInfo, memberName, ma); constValue != nil {
				return constValue
			}
			if propInfo := classInfo.lookupProperty(memberName); propInfo != nil {
				if propInfo.IsClassProperty {
					return i.evalClassPropertyRead(classInfo, propInfo, ma)
				}
				if result := i.canAccessInstancePropertyViaClass(classInfo, propInfo, ma); result != nil {
					return result
				}
			}

			// Constructors: auto-invoke without parentheses
			if classInfo.HasConstructor(memberName) {
				constructorOverloads := i.lookupConstructorOverloadsInHierarchy(classInfo, memberName)
				if len(constructorOverloads) > 0 {
					methodCall := &ast.MethodCallExpression{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: ma.Token,
							},
						},
						Object:    ma.Object,
						Method:    ma.Member,
						Arguments: []ast.Expression{},
					}
					return i.evalMethodCall(methodCall)
				}
			}

			// Class methods: auto-invoke if parameterless, else return function pointer
			if classMethod := i.lookupClassMethodInHierarchy(classInfo, memberName); classMethod != nil {
				if len(classMethod.Parameters) == 0 {
					methodCall := &ast.MethodCallExpression{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: ma.Token,
							},
						},
						Object:    ma.Object,
						Method:    ma.Member,
						Arguments: []ast.Expression{},
					}
					return i.evalMethodCall(methodCall)
				}
				paramTypes := make([]types.Type, len(classMethod.Parameters))
				for idx, param := range classMethod.Parameters {
					if param.Type != nil {
						paramTypes[idx] = i.getTypeFromAnnotation(param.Type)
					}
				}
				var returnType types.Type
				if classMethod.ReturnType != nil {
					returnType = i.getTypeFromAnnotation(classMethod.ReturnType)
				}
				pointerType := types.NewFunctionPointerType(paramTypes, returnType)
				return NewFunctionPointerValue(classMethod, i.env, nil, pointerType)
			}

			// Nested class access
			if nested := classInfo.lookupNestedClass(memberName); nested != nil {
				return &ClassInfoValue{ClassInfo: nested}
			}

			return i.newErrorWithLocation(ma, "member '%s' not found in class '%s'", memberName, classInfo.Name)
		}

		// Enum scoped access: TColor.Red
		if enumMetadata := i.typeSystem.LookupEnumMetadata(ident.Value); enumMetadata != nil {
			if etv, ok := enumMetadata.(*EnumTypeValue); ok {
				enumType := etv.EnumType
				valueName := ma.Member.Value

				// Enum values take precedence (allows shadowing Low/High)
				if ordinalValue, exists := enumType.Values[valueName]; exists {
					return &EnumValue{
						TypeName:     ident.Value,
						ValueName:    valueName,
						OrdinalValue: ordinalValue,
					}
				}

				// Special enum properties
				lowerMember := pkgident.Normalize(valueName)
				switch lowerMember {
				case "low":
					return &IntegerValue{Value: int64(enumType.Low())}
				case "high":
					return &IntegerValue{Value: int64(enumType.High())}
				}

				// Unscoped enums: fallback to environment lookup
				if !enumType.Scoped {
					if val, envExists := i.env.Get(valueName); envExists {
						if enumVal, isEnum := val.(*EnumValue); isEnum {
							if enumVal.TypeName == ident.Value {
								return enumVal
							}
						}
					}
				}

				return i.newErrorWithLocation(ma, "enum value '%s' not found in enum '%s'", ma.Member.Value, ident.Value)
			}
		}

		// Record type static access: TPoint.cOrigin, TPoint.Count
		recordTypeKey := "__record_type_" + pkgident.Normalize(ident.Value)
		if recordTypeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := recordTypeVal.(*RecordTypeValue); ok {
				memberName := ma.Member.Value

				if constValue, exists := rtv.Constants[pkgident.Normalize(memberName)]; exists {
					return constValue
				}
				if classVarValue, exists := rtv.ClassVars[pkgident.Normalize(memberName)]; exists {
					return classVarValue
				}
				if classMethod, exists := rtv.ClassMethods[pkgident.Normalize(memberName)]; exists {
					if len(classMethod.Parameters) == 0 {
						methodCall := &ast.MethodCallExpression{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: ma.Token,
								},
							},
							Object:    ma.Object,
							Method:    ma.Member,
							Arguments: []ast.Expression{},
						}
						return i.evalMethodCall(methodCall)
					}
					return i.newErrorWithLocation(ma, "class method '%s' of record '%s' requires %d parameter(s); use parentheses to call",
						memberName, ident.Value, len(classMethod.Parameters))
				}

				return i.newErrorWithLocation(ma, "member '%s' not found in record type '%s'", memberName, ident.Value)
			}
		}
	}

	// Helper class constants and variables via TypeMetaValue
	if objIdent, ok := ma.Object.(*ast.Identifier); ok {
		if typeMetaVal, exists := i.env.Get(objIdent.Value); exists {
			if tmv, ok := typeMetaVal.(*TypeMetaValue); ok {
				if constVal := i.findHelperClassConst(tmv, ma.Member.Value); constVal != nil {
					return constVal
				}
				if varVal := i.findHelperClassVar(tmv, ma.Member.Value); varVal != nil {
					return varVal
				}
			}
		}
	}

	// Instance access: evaluate object expression
	objVal := i.Eval(ma.Object)
	if isError(objVal) {
		return objVal
	}
	if i.exception != nil {
		return nil
	}

	// Record instance access
	if recordVal, ok := objVal.(*RecordValue); ok {
		fieldValue, exists := recordVal.Fields[pkgident.Normalize(ma.Member.Value)]
		if exists {
			return fieldValue
		}

		memberNameLower := pkgident.Normalize(ma.Member.Value)

		// Record properties
		if propInfo, exists := recordVal.RecordType.Properties[memberNameLower]; exists {
			if propInfo.ReadField != "" {
				if fieldVal, exists := recordVal.Fields[pkgident.Normalize(propInfo.ReadField)]; exists {
					return fieldVal
				}
				if getterMethod := GetRecordMethod(recordVal, propInfo.ReadField); getterMethod != nil {
					methodCall := &ast.MethodCallExpression{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: ma.Token,
							},
						},
						Object:    ma.Object,
						Method:    &ast.Identifier{Value: propInfo.ReadField, TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{Token: ma.Token}}},
						Arguments: []ast.Expression{},
					}
					return i.evalMethodCall(methodCall)
				}
				return i.newErrorWithLocation(ma, "property '%s' read accessor '%s' not found in record '%s'",
					ma.Member.Value, propInfo.ReadField, recordVal.RecordType.Name)
			}
			return i.newErrorWithLocation(ma, "property '%s' is write-only", ma.Member.Value)
		}

		// Record methods: auto-invoke if parameterless
		methodDecl := GetRecordMethod(recordVal, ma.Member.Value)
		if methodDecl != nil {
			if len(methodDecl.Parameters) == 0 {
				methodCall := &ast.MethodCallExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: ma.Token,
						},
					},
					Object:    ma.Object,
					Method:    ma.Member,
					Arguments: []ast.Expression{},
				}
				return i.evalMethodCall(methodCall)
			}
			return i.newErrorWithLocation(ma, "method '%s' of record '%s' requires %d parameter(s); use parentheses to call",
				ma.Member.Value, recordVal.RecordType.Name, len(methodDecl.Parameters))
		}

		// Record class-level members (accessible via instance)
		recordTypeKey := "__record_type_" + pkgident.Normalize(recordVal.RecordType.Name)
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				if classMethod, exists := rtv.ClassMethods[memberNameLower]; exists {
					if len(classMethod.Parameters) == 0 {
						methodCall := &ast.MethodCallExpression{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: ma.Token,
								},
							},
							Object:    ma.Object,
							Method:    ma.Member,
							Arguments: []ast.Expression{},
						}
						return i.evalMethodCall(methodCall)
					}
					return i.newErrorWithLocation(ma, "class method '%s' of record '%s' requires %d parameter(s); use parentheses to call",
						ma.Member.Value, recordVal.RecordType.Name, len(classMethod.Parameters))
				}
				if constValue, exists := rtv.Constants[memberNameLower]; exists {
					return constValue
				}
				if classVarValue, exists := rtv.ClassVars[memberNameLower]; exists {
					return classVarValue
				}
			}
		}

		helper, helperProp := i.findHelperProperty(recordVal, ma.Member.Value)
		if helperProp != nil {
			return i.evalHelperPropertyRead(helper, helperProp, recordVal, ma)
		}

		return i.newErrorWithLocation(ma, "field '%s' not found in record '%s'", ma.Member.Value, recordVal.RecordType.Name)
	}

	// ClassInfoValue or ClassValue (metaclass) access
	var classInfo *ClassInfo
	if classInfoVal, ok := objVal.(*ClassInfoValue); ok {
		classInfo = classInfoVal.ClassInfo
	} else if classVal, ok := objVal.(*ClassValue); ok {
		classInfo = classVal.ClassInfo
	}

	if classInfo != nil {
		memberName := ma.Member.Value

		// Built-in metaclass properties
		if pkgident.Equal(memberName, "ClassName") {
			return &StringValue{Value: classInfo.Name}
		}
		if pkgident.Equal(memberName, "ClassType") {
			return &ClassValue{ClassInfo: classInfo}
		}

		// Class-level members
		if classVarValue, ownerClass := classInfo.lookupClassVar(memberName); ownerClass != nil {
			return classVarValue
		}
		if constValue := i.getClassConstant(classInfo, memberName, ma); constValue != nil {
			return constValue
		}
		if propInfo := classInfo.lookupProperty(memberName); propInfo != nil {
			if propInfo.IsClassProperty {
				return i.evalClassPropertyRead(classInfo, propInfo, ma)
			}
			if result := i.canAccessInstancePropertyViaClass(classInfo, propInfo, ma); result != nil {
				return result
			}
		}

		// Constructors and class methods
		if classInfo.HasConstructor(memberName) {
			constructorOverloads := i.lookupConstructorOverloadsInHierarchy(classInfo, memberName)
			if len(constructorOverloads) > 0 {
				methodCall := &ast.MethodCallExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: ma.Token,
						},
					},
					Object:    ma.Object,
					Method:    ma.Member,
					Arguments: []ast.Expression{},
				}
				return i.evalMethodCall(methodCall)
			}
		}
		if classMethod := i.lookupClassMethodInHierarchy(classInfo, memberName); classMethod != nil {
			if len(classMethod.Parameters) == 0 {
				methodCall := &ast.MethodCallExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: ma.Token,
						},
					},
					Object:    ma.Object,
					Method:    ma.Member,
					Arguments: []ast.Expression{},
				}
				return i.evalMethodCall(methodCall)
			}
			paramTypes := make([]types.Type, len(classMethod.Parameters))
			for idx, param := range classMethod.Parameters {
				if param.Type != nil {
					paramTypes[idx] = i.getTypeFromAnnotation(param.Type)
				}
			}
			var returnType types.Type
			if classMethod.ReturnType != nil {
				returnType = i.getTypeFromAnnotation(classMethod.ReturnType)
			}
			pointerType := types.NewFunctionPointerType(paramTypes, returnType)
			return NewFunctionPointerValue(classMethod, i.env, nil, pointerType)
		}

		return i.newErrorWithLocation(ma, "member '%s' not found in class '%s'", memberName, classInfo.Name)
	}

	// Interface instance access: delegate to underlying object
	if intfInst, ok := objVal.(*InterfaceInstance); ok {
		if intfInst.Object == nil {
			return i.newErrorWithLocation(ma.Member, "Interface is nil")
		}

		memberName := ma.Member.Value
		propInfo := intfInst.Interface.GetProperty(memberName)
		hasMethod := intfInst.Interface.HasMethod(memberName)
		if !hasMethod && propInfo == nil {
			return i.newErrorWithLocation(ma, "member '%s' not found in interface '%s'", memberName, intfInst.Interface.GetName())
		}

		underlyingObj, isObj := AsObject(intfInst.Object)
		if !isObj {
			return i.newErrorWithLocation(ma, "interface underlying object is not a class instance")
		}

		// Interface properties
		if propInfo != nil {
			typesPropertyInfo, ok := propInfo.Impl.(*types.PropertyInfo)
			if !ok {
				return i.newErrorWithLocation(ma, "invalid property info type")
			}
			return i.evalPropertyRead(underlyingObj, typesPropertyInfo, ma)
		}

		// Interface methods: auto-invoke parameterless, else return function pointer
		concreteClass, ok := underlyingObj.Class.(*ClassInfo)
		if !ok {
			return i.newErrorWithLocation(ma, "interface wraps invalid object class")
		}
		methodOverloads := i.getMethodOverloadsInHierarchy(concreteClass, memberName, false)
		classMethodOverloads := i.getMethodOverloadsInHierarchy(concreteClass, memberName, true)

		if len(methodOverloads) == 0 && len(classMethodOverloads) == 0 {
			return i.newErrorWithLocation(ma, "method '%s' declared in interface '%s' but not implemented by class '%s'",
				memberName, intfInst.Interface.GetName(), underlyingObj.Class.GetName())
		}

		var method *ast.FunctionDecl
		if len(methodOverloads) > 0 {
			method = methodOverloads[0]
		} else {
			method = classMethodOverloads[0]
		}

		if method.ReturnType != nil && len(method.Parameters) == 0 {
			methodCall := &ast.MethodCallExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: ma.Token,
					},
				},
				Object:    ma.Object,
				Method:    ma.Member,
				Arguments: []ast.Expression{},
			}
			return i.evalMethodCall(methodCall)
		}

		paramTypes := make([]types.Type, len(method.Parameters))
		for idx, param := range method.Parameters {
			if param.Type != nil {
				paramTypes[idx] = i.getTypeFromAnnotation(param.Type)
			}
		}
		var returnType types.Type
		if method.ReturnType != nil {
			returnType = i.getTypeFromAnnotation(method.ReturnType)
		}
		pointerType := types.NewFunctionPointerType(paramTypes, returnType)
		return NewFunctionPointerValue(method, i.env, underlyingObj, pointerType)
	}

	// Type cast values: preserve static type for class variable access
	var staticClassType *ClassInfo
	if typeCast, isTypeCast := objVal.(*TypeCastValue); isTypeCast {
		staticClassType = typeCast.StaticType
		objVal = typeCast.Object
	}

	obj, ok := AsObject(objVal)

	// Nil object access: handle class variables and TObject.Free
	nilVal, isNilValue := objVal.(*NilValue)
	if !ok && (objVal.Type() == "NIL" || isNilValue) {
		memberName := ma.Member.Value

		if pkgident.Equal(memberName, "Free") {
			return &NilValue{}
		}

		// Try class variable lookup using static type or typed nil
		if staticClassType != nil {
			if classVarValue, ownerClass := staticClassType.lookupClassVar(memberName); ownerClass != nil {
				return classVarValue
			}
		} else if isNilValue && nilVal.ClassType != "" {
			for registeredClassName, classInfo := range i.classes {
				if pkgident.Equal(registeredClassName, nilVal.ClassType) {
					if classVarValue, ownerClass := classInfo.lookupClassVar(memberName); ownerClass != nil {
						return classVarValue
					}
					break
				}
			}
		}

		message := fmt.Sprintf("Object not instantiated [line: %d, column: %d]", ma.Token.Pos.Line, ma.Token.Pos.Column+1)
		i.raiseException("Exception", message, &ma.Token.Pos)
		return &NilValue{}
	}

	if !ok {
		// Enum values: .Value property
		if enumVal, isEnum := objVal.(*EnumValue); isEnum {
			memberName := pkgident.Normalize(ma.Member.Value)
			if memberName == "value" {
				return &IntegerValue{Value: int64(enumVal.OrdinalValue)}
			}
		}

		// Non-object types: try helpers (methods take precedence over properties)
		helper, methodDecl, builtinSpec := i.findHelperMethod(objVal, ma.Member.Value)
		if helper != nil && (methodDecl != nil || builtinSpec != "") {
			isParameterless := false
			if methodDecl != nil {
				isParameterless = len(methodDecl.Parameters) == 0
			} else if builtinSpec != "" {
				isParameterless = i.isBuiltinMethodParameterless(builtinSpec)
			}

			if isParameterless {
				methodCall := &ast.MethodCallExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: ma.Token,
						},
					},
					Object:    ma.Object,
					Method:    ma.Member,
					Arguments: []ast.Expression{},
				}
				return i.evalMethodCall(methodCall)
			}
		}

		helper, helperProp := i.findHelperProperty(objVal, ma.Member.Value)
		if helperProp != nil {
			return i.evalHelperPropertyRead(helper, helperProp, objVal, ma)
		}
		if constVal := i.findHelperClassConst(objVal, ma.Member.Value); constVal != nil {
			return constVal
		}
		if varVal := i.findHelperClassVar(objVal, ma.Member.Value); varVal != nil {
			return varVal
		}

		return i.newErrorWithLocation(ma, "cannot access member '%s' of type '%s' (no helper found)",
			ma.Member.Value, objVal.Type())
	}

	memberName := ma.Member.Value

	if obj.Destroyed {
		message := fmt.Sprintf("Object already destroyed [line: %d, column: %d]", ma.Token.Pos.Line, ma.Token.Pos.Column)
		i.raiseException("Exception", message, &ma.Token.Pos)
		return &NilValue{}
	}

	// Built-in TObject properties
	if pkgident.Equal(memberName, "ClassName") {
		return &StringValue{Value: obj.Class.GetName()}
	}
	if pkgident.Equal(memberName, "ClassType") {
		concreteClass, ok := obj.Class.(*ClassInfo)
		if !ok {
			return i.newErrorWithLocation(ma, "object has invalid class type")
		}
		return &ClassValue{ClassInfo: concreteClass}
	}

	// Properties take precedence over fields
	if propDesc := obj.Class.LookupProperty(memberName); propDesc != nil {
		propInfo, ok := propDesc.Impl.(*types.PropertyInfo)
		if !ok {
			return i.newErrorWithLocation(ma, "invalid property descriptor")
		}
		return i.evalPropertyRead(obj, propInfo, ma)
	}

	// Direct field access
	fieldValue := obj.GetField(memberName)
	if fieldValue == nil {
		// Class variables (use static type from cast if available)
		classForLookup := obj.Class
		if staticClassType != nil {
			classForLookup = staticClassType
		}
		if classVarValue, ownerClass := classForLookup.LookupClassVar(memberName); ownerClass != nil {
			return classVarValue
		}

		// Class constants
		concreteClass, ok := obj.Class.(*ClassInfo)
		if !ok {
			return i.newErrorWithLocation(ma, "object has invalid class type")
		}
		if constValue := i.getClassConstant(concreteClass, memberName, ma); constValue != nil {
			return constValue
		}

		// Methods (instance and class methods)
		concreteClassForMethod, ok := obj.Class.(*ClassInfo)
		if !ok {
			return i.newErrorWithLocation(ma, "object has invalid class type")
		}
		methodOverloads := i.getMethodOverloadsInHierarchy(concreteClassForMethod, memberName, false)
		classMethodOverloads := i.getMethodOverloadsInHierarchy(concreteClassForMethod, memberName, true)

		if len(methodOverloads) > 0 || len(classMethodOverloads) > 0 {
			hasParameterlessOverload := false

			for _, overload := range methodOverloads {
				if len(overload.Parameters) == 0 {
					hasParameterlessOverload = true
					break
				}
			}
			if !hasParameterlessOverload {
				for _, overload := range classMethodOverloads {
					if len(overload.Parameters) == 0 {
						hasParameterlessOverload = true
						break
					}
				}
			}

			// Auto-invoke parameterless methods
			if hasParameterlessOverload {
				methodCall := &ast.MethodCallExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: ma.Token,
						},
					},
					Object:    ma.Object,
					Method:    ma.Member,
					Arguments: []ast.Expression{},
				}
				return i.evalMethodCall(methodCall)
			}

			// Return function pointer for methods with parameters
			var method *ast.FunctionDecl
			if len(methodOverloads) > 0 {
				method = methodOverloads[0]
			} else {
				method = classMethodOverloads[0]
			}

			paramTypes := make([]types.Type, len(method.Parameters))
			for idx, param := range method.Parameters {
				if param.Type != nil {
					paramTypes[idx] = i.getTypeFromAnnotation(param.Type)
				}
			}
			var returnType types.Type
			if method.ReturnType != nil {
				returnType = i.getTypeFromAnnotation(method.ReturnType)
			}
			pointerType := types.NewFunctionPointerType(paramTypes, returnType)
			return NewFunctionPointerValue(method, i.env, obj, pointerType)
		}

		helper, helperProp := i.findHelperProperty(obj, memberName)
		if helperProp != nil {
			return i.evalHelperPropertyRead(helper, helperProp, obj, ma)
		}
		return i.newErrorWithLocation(ma, "field '%s' not found in class '%s'", memberName, obj.Class.GetName())
	}

	return fieldValue
}

// lookupConstructorOverloadsInHierarchy returns all constructor overloads with the given name
// by searching the class hierarchy. Case-insensitive.
func (i *Interpreter) lookupConstructorOverloadsInHierarchy(classInfo *ClassInfo, name string) []*ast.FunctionDecl {
	for current := classInfo; current != nil; current = current.Parent {
		for ctorName, overloads := range current.ConstructorOverloads {
			if pkgident.Equal(ctorName, name) && len(overloads) > 0 {
				return overloads
			}
		}
		for ctorName, constructor := range current.Constructors {
			if pkgident.Equal(ctorName, name) {
				return []*ast.FunctionDecl{constructor}
			}
		}
	}
	return nil
}

// lookupClassMethodInHierarchy searches for a class method by name in the class hierarchy.
// Case-insensitive.
func (i *Interpreter) lookupClassMethodInHierarchy(classInfo *ClassInfo, name string) *ast.FunctionDecl {
	normalizedName := pkgident.Normalize(name)
	for current := classInfo; current != nil; current = current.Parent {
		if method, exists := current.ClassMethods[normalizedName]; exists {
			return method
		}
	}
	return nil
}

// bindClassConstantsToEnv adds all class constants to the current environment,
// allowing methods to access them directly without qualification.
func (i *Interpreter) bindClassConstantsToEnv(classInfo *ClassInfo) {
	for constName, constValue := range classInfo.ConstantValues {
		i.env.Define(constName, constValue)
	}
}

// evalSelfExpression returns the Self value from the current method context.
func (i *Interpreter) evalSelfExpression(se *ast.SelfExpression) Value {
	selfVal, exists := i.env.Get("Self")
	if !exists {
		return i.newErrorWithLocation(se, "Self used outside method context")
	}
	return selfVal
}

// evalInheritedExpression calls a parent class method.
// Supports explicit (inherited MethodName(args)) and bare (inherited) forms.
func (i *Interpreter) evalInheritedExpression(ie *ast.InheritedExpression) Value {
	selfVal, exists := i.env.Get("Self")
	if !exists {
		return i.newErrorWithLocation(ie, "inherited can only be used inside a method")
	}

	obj, ok := selfVal.(*ObjectInstance)
	if !ok {
		return i.newErrorWithLocation(ie, "inherited requires Self to be an object instance")
	}

	// Determine static class context (prefer __CurrentClass__, fall back to runtime class)
	var classInfo *ClassInfo
	if currentClassVal, has := i.env.Get("__CurrentClass__"); has {
		if civ, isClassVal := currentClassVal.(*ClassInfoValue); isClassVal && civ.ClassInfo != nil {
			classInfo = civ.ClassInfo
		}
	}
	if classInfo == nil {
		concreteClass, ok := obj.Class.(*ClassInfo)
		if !ok {
			return i.newErrorWithLocation(ie, "object has invalid class type")
		}
		classInfo = concreteClass
	}

	if classInfo.Parent == nil {
		return i.newErrorWithLocation(ie, "class '%s' has no parent class", classInfo.Name)
	}
	parentClass := classInfo.Parent

	// Determine method name (explicit or from __CurrentMethod__)
	var methodName string
	if ie.Method != nil {
		methodName = ie.Method.Value
	} else {
		currentMethodVal, exists := i.env.Get("__CurrentMethod__")
		if !exists {
			return i.newErrorWithLocation(ie, "bare 'inherited' requires method context")
		}
		currentMethodName, ok := currentMethodVal.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(ie, "invalid method context")
		}
		methodName = currentMethodName.Value
	}

	// Try method lookup in parent class
	methodOverloads := i.getMethodOverloadsInHierarchy(parentClass, methodName, false)
	if len(methodOverloads) > 0 {
		parentMethod, err := i.resolveMethodOverload(parentClass.Name, methodName, methodOverloads, ie.Arguments)
		if err != nil {
			return i.newErrorWithLocation(ie, "%s", err.Error())
		}

		// Evaluate arguments
		args := make([]Value, len(ie.Arguments))
		for idx, arg := range ie.Arguments {
			val := i.Eval(arg)
			if isError(val) {
				return val
			}
			args[idx] = val
		}

		if len(args) != len(parentMethod.Parameters) {
			return i.newErrorWithLocation(ie, "wrong number of arguments for method '%s': expected %d, got %d",
				methodName, len(parentMethod.Parameters), len(args))
		}

		// Create method environment
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		i.env.Define("Self", obj)
		i.env.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: parentClass})
		i.bindClassConstantsToEnv(parentClass)
		i.env.Define("__CurrentMethod__", &StringValue{Value: methodName})

		// Bind parameters with implicit conversion
		for idx, param := range parentMethod.Parameters {
			arg := args[idx]
			if param.Type != nil {
				paramTypeName := param.Type.String()
				if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
					arg = converted
				}
			}
			i.env.Define(param.Name.Value, arg)
		}

		// Initialize Result for functions
		if parentMethod.ReturnType != nil {
			returnType := i.resolveTypeFromAnnotation(parentMethod.ReturnType)
			defaultVal := i.getDefaultValue(returnType)
			i.env.Define("Result", defaultVal)
			i.env.Define(parentMethod.Name.Value, &ReferenceValue{Env: i.env, VarName: "Result"})
		}

		_ = i.Eval(parentMethod.Body)

		// Return value handling
		var returnValue Value
		if parentMethod.ReturnType != nil {
			if resultVal, ok := i.env.Get("Result"); ok {
				returnValue = resultVal
			} else if methodVal, ok := i.env.Get(parentMethod.Name.Value); ok {
				returnValue = methodVal
			} else {
				returnValue = &NilValue{}
			}
		} else {
			returnValue = &NilValue{}
		}

		i.env = savedEnv
		return returnValue
	}

	// Try properties in parent class
	var propInfo *types.PropertyInfo
	for name, prop := range parentClass.Properties {
		if pkgident.Equal(name, methodName) {
			propInfo = prop
			break
		}
	}
	if propInfo != nil {
		if len(ie.Arguments) > 0 || ie.IsCall {
			return i.newErrorWithLocation(ie, "cannot call property '%s' as a method", methodName)
		}
		return i.evalPropertyRead(obj, propInfo, ie)
	}

	// Try fields in parent class
	for name := range parentClass.Fields {
		if pkgident.Equal(name, methodName) {
			if len(ie.Arguments) > 0 || ie.IsCall {
				return i.newErrorWithLocation(ie, "cannot call field '%s' as a method", methodName)
			}
			fieldValue := obj.GetField(name)
			if fieldValue == nil {
				return &NilValue{}
			}
			return fieldValue
		}
	}

	return i.newErrorWithLocation(ie, "method, property, or field '%s' not found in parent class '%s'", methodName, parentClass.Name)
}

// getClassConstant retrieves and caches a class constant value by name.
// Evaluates lazily on first access and supports inheritance.
func (i *Interpreter) getClassConstant(classInfo *ClassInfo, constantName string, ma *ast.MemberAccessExpression) Value {
	constDecl, ownerClass := classInfo.lookupConstant(constantName)
	if constDecl == nil {
		return nil
	}

	if cachedValue, cached := ownerClass.ConstantValues[constantName]; cached {
		return cachedValue
	}

	// Evaluate constant in temporary environment with other evaluated constants
	savedEnv := i.env
	tempEnv := NewEnclosedEnvironment(i.env)

	for constName, constVal := range ownerClass.ConstantValues {
		if constName != constantName && constVal != nil {
			tempEnv.Define(constName, constVal)
		}
	}

	i.env = tempEnv
	constValue := i.Eval(constDecl.Value)
	i.env = savedEnv

	if isError(constValue) {
		return constValue
	}

	ownerClass.ConstantValues[constantName] = constValue
	return constValue
}

// canAccessInstancePropertyViaClass checks if an instance property can be accessed
// via the class (TClass.PropertyName) when the read spec is a class variable or constant.
func (i *Interpreter) canAccessInstancePropertyViaClass(classInfo *ClassInfo, propInfo *types.PropertyInfo, ma *ast.MemberAccessExpression) Value {
	if propInfo.ReadKind != types.PropAccessField {
		return nil
	}

	if _, ownerClass := classInfo.lookupClassVar(propInfo.ReadSpec); ownerClass != nil {
		tempObj := &ObjectInstance{Class: classInfo, Fields: make(map[string]Value)}
		return i.evalPropertyRead(tempObj, propInfo, ma)
	}

	if _, ownerClass := classInfo.lookupConstant(propInfo.ReadSpec); ownerClass != nil {
		tempObj := &ObjectInstance{Class: classInfo, Fields: make(map[string]Value)}
		return i.evalPropertyRead(tempObj, propInfo, ma)
	}

	return nil
}
