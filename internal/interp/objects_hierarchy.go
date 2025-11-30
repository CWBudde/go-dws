package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	pkgident "github.com/cwbudde/go-dws/pkg/ident"
)

// evalMemberAccess evaluates field access (obj.field) or class variable access (TClass.Variable).
// It evaluates the object expression and retrieves the field value.
// For class variable access, it checks if the left side is a class name.
func (i *Interpreter) evalMemberAccess(ma *ast.MemberAccessExpression) Value {
	// Check if the left side is a class identifier (for static access: TClass.Variable)
	if ident, ok := ma.Object.(*ast.Identifier); ok {
		// First, check if this identifier refers to a unit (for qualified access: UnitName.Symbol)
		if i.unitRegistry != nil {
			if _, exists := i.unitRegistry.GetUnit(ident.Value); exists {
				// This is unit-qualified access: UnitName.Symbol
				// Try to resolve as a variable/constant first
				if val, err := i.ResolveQualifiedVariable(ident.Value, ma.Member.Value); err == nil {
					return val
				}
				// If not a variable, it might be a function being passed as a reference
				// For now, we'll return an error since function references aren't fully supported yet
				// The actual function call will be handled in evalCallExpression
				return i.newErrorWithLocation(ma, "qualified name '%s.%s' cannot be used as a value (functions must be called)", ident.Value, ma.Member.Value)
			}
		}

		// Task 9.68: Check if this identifier refers to a class (case-insensitive)
		classInfo := i.resolveClassInfoByName(ident.Value)
		if classInfo != nil {
			// This is static access: TClass.Variable
			memberName := ma.Member.Value

			// Check for built-in ClassType and ClassName properties (case-insensitive)
			if pkgident.Equal(memberName, "ClassType") {
				return &ClassValue{ClassInfo: classInfo}
			}
			if pkgident.Equal(memberName, "ClassName") {
				return &StringValue{Value: classInfo.Name}
			}

			// 1. Try class variables first (case-insensitive)
			if classVarValue, ownerClass := classInfo.lookupClassVar(memberName); ownerClass != nil {
				return classVarValue
			}

			// 2. Task 9.22: Try class constants
			if constValue := i.getClassConstant(classInfo, memberName, ma); constValue != nil {
				return constValue
			}

			// 3. Task 9.13: Try class properties
			// Task 9.17: Also allow instance properties that use class constants or class variables
			if propInfo := classInfo.lookupProperty(memberName); propInfo != nil {
				if propInfo.IsClassProperty {
					return i.evalClassPropertyRead(classInfo, propInfo, ma)
				}
				// Task 9.17: Allow instance properties accessed on class if they use class-level read specs
				if result := i.canAccessInstancePropertyViaClass(classInfo, propInfo, ma); result != nil {
					return result
				}
			}

			// 3. Task 9.32: Try constructors (with inheritance support)
			// Task 9.68: Also handle implicit parameterless constructor
			// Task 9.82: Handle constructor overloads properly
			if classInfo.HasConstructor(memberName) {
				// Find all constructor overloads in the hierarchy
				constructorOverloads := i.lookupConstructorOverloadsInHierarchy(classInfo, memberName)
				if len(constructorOverloads) > 0 {
					// Task 9.21: When accessing constructor without parentheses (TClass.Create),
					// invoke with 0 arguments. If no parameterless constructor exists,
					// the implicit parameterless constructor will be used.
					methodCall := &ast.MethodCallExpression{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: ma.Token,
							},
						},
						Object:    ma.Object, // TClassName identifier
						Method:    ma.Member, // Constructor name
						Arguments: []ast.Expression{},
					}
					return i.evalMethodCall(methodCall)
				}
			}

			// 3. Task 9.32: Try class methods (static methods)
			if classMethod := i.lookupClassMethodInHierarchy(classInfo, memberName); classMethod != nil {
				// Check if parameterless
				if len(classMethod.Parameters) == 0 {
					// Auto-invoke the class method
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
				// Class method has parameters - return as function pointer
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

			// 4. Nested class access (TOuter.TInner)
			if nested := classInfo.lookupNestedClass(memberName); nested != nil {
				return &ClassInfoValue{ClassInfo: nested}
			}

			// 4. Not found anywhere - error
			return i.newErrorWithLocation(ma, "member '%s' not found in class '%s'", memberName, classInfo.Name)
		}

		// Check if this identifier refers to an enum type (for scoped access: TColor.Red)
		// Look for enum type metadata via TypeSystem (Task 3.5.143b)
		if enumMetadata := i.typeSystem.LookupEnumMetadata(ident.Value); enumMetadata != nil {
			if etv, ok := enumMetadata.(*EnumTypeValue); ok {
				enumType := etv.EnumType
				// This is scoped enum access: TColor.Red
				// Look up the enum value in the enum type's values
				valueName := ma.Member.Value

				// For scoped enums, look up directly in the enum type's values FIRST
				// This allows enum values named "Low" or "High" to shadow the properties
				if ordinalValue, exists := enumType.Values[valueName]; exists {
					// Create and return the enum value
					return &EnumValue{
						TypeName:     ident.Value,
						ValueName:    valueName,
						OrdinalValue: ordinalValue,
					}
				}

				// Check for special enum properties: Low and High
				// These are only used if there's no enum value with that name
				lowerMember := pkgident.Normalize(valueName)
				switch lowerMember {
				case "low":
					return &IntegerValue{Value: int64(enumType.Low())}
				case "high":
					return &IntegerValue{Value: int64(enumType.High())}
				}

				// For unscoped enums, try to look up in environment as well
				if !enumType.Scoped {
					if val, envExists := i.env.Get(valueName); envExists {
						if enumVal, isEnum := val.(*EnumValue); isEnum {
							// Verify the value belongs to this enum type
							if enumVal.TypeName == ident.Value {
								return enumVal
							}
						}
					}
				}

				// Enum value not found
				return i.newErrorWithLocation(ma, "enum value '%s' not found in enum '%s'", ma.Member.Value, ident.Value)
			}
		}

		// Task 9.12.2: Check if this identifier refers to a record type (for static access: TPoint.cOrigin, TPoint.Count)
		recordTypeKey := "__record_type_" + pkgident.Normalize(ident.Value)
		if recordTypeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := recordTypeVal.(*RecordTypeValue); ok {
				memberName := ma.Member.Value

				// Check constants (case-insensitive)
				if constValue, exists := rtv.Constants[pkgident.Normalize(memberName)]; exists {
					return constValue
				}

				// Check class variables (case-insensitive)
				if classVarValue, exists := rtv.ClassVars[pkgident.Normalize(memberName)]; exists {
					return classVarValue
				}

				// Check class methods (case-insensitive)
				if classMethod, exists := rtv.ClassMethods[pkgident.Normalize(memberName)]; exists {
					// Check if parameterless
					if len(classMethod.Parameters) == 0 {
						// Auto-invoke the class method
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
					// Class method has parameters - error (cannot call without parentheses)
					return i.newErrorWithLocation(ma, "class method '%s' of record '%s' requires %d parameter(s); use parentheses to call",
						memberName, ident.Value, len(classMethod.Parameters))
				}

				// Member not found in record type
				return i.newErrorWithLocation(ma, "member '%s' not found in record type '%s'", memberName, ident.Value)
			}
		}
	}

	// Not static access - evaluate the object expression for instance access
	objVal := i.Eval(ma.Object)
	if isError(objVal) {
		return objVal
	}

	// Check if an exception was raised during object evaluation
	if i.exception != nil {
		return nil
	}

	// Check if it's a record value
	if recordVal, ok := objVal.(*RecordValue); ok {
		// Access record field (case-insensitive - use lowercase key)
		fieldValue, exists := recordVal.Fields[pkgident.Normalize(ma.Member.Value)]
		if exists {
			return fieldValue
		}

		// Check for properties, class methods, constants, and class variables
		memberNameLower := pkgident.Normalize(ma.Member.Value)

		// Check if this member is a property
		if propInfo, exists := recordVal.RecordType.Properties[memberNameLower]; exists {
			// Property found - evaluate read access
			if propInfo.ReadField != "" {
				// Check if ReadField is a field name or method name
				// First try as a field (use lowercase key)
				if fieldVal, exists := recordVal.Fields[pkgident.Normalize(propInfo.ReadField)]; exists {
					return fieldVal
				}

				// Not a field - try as a method (getter)
				// Task 3.5.128b: Use free function instead of method due to type alias
				if getterMethod := GetRecordMethod(recordVal, propInfo.ReadField); getterMethod != nil {
					// Call the getter method
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
			// Property is write-only
			return i.newErrorWithLocation(ma, "property '%s' is write-only", ma.Member.Value)
		}

		// Task 9.37: Check if it's a record method (via Metadata.Methods)
		// Task 3.5.128a: Use GetMethod which now only uses Metadata.Methods
		// Task 3.5.128b: Use free function instead of method due to type alias
		methodDecl := GetRecordMethod(recordVal, ma.Member.Value)
		if methodDecl != nil {
			// Only auto-invoke parameterless methods when accessed without parentheses
			if len(methodDecl.Parameters) == 0 {
				// Convert to a method call expression and evaluate it
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
			// Method has parameters - cannot auto-invoke without parentheses
			return i.newErrorWithLocation(ma, "method '%s' of record '%s' requires %d parameter(s); use parentheses to call",
				ma.Member.Value, recordVal.RecordType.Name, len(methodDecl.Parameters))
		}

		// Task 9.12.2: Check for class methods, constants, and class variables (accessible via instance)
		// Look up the record type value once for all checks
		recordTypeKey := "__record_type_" + pkgident.Normalize(recordVal.RecordType.Name)
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				// Check class methods (case-insensitive)
				if classMethod, exists := rtv.ClassMethods[memberNameLower]; exists {
					// Check if parameterless
					if len(classMethod.Parameters) == 0 {
						// Auto-invoke the class method
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
					// Class method has parameters - cannot auto-invoke without parentheses
					return i.newErrorWithLocation(ma, "class method '%s' of record '%s' requires %d parameter(s); use parentheses to call",
						ma.Member.Value, recordVal.RecordType.Name, len(classMethod.Parameters))
				}

				// Check constants (case-insensitive)
				if constValue, exists := rtv.Constants[memberNameLower]; exists {
					return constValue
				}

				// Check class variables (case-insensitive)
				if classVarValue, exists := rtv.ClassVars[memberNameLower]; exists {
					return classVarValue
				}
			}
		}

		// Check if helpers provide this member
		helper, helperProp := i.findHelperProperty(recordVal, ma.Member.Value)
		if helperProp != nil {
			return i.evalHelperPropertyRead(helper, helperProp, recordVal, ma)
		}

		return i.newErrorWithLocation(ma, "field '%s' not found in record '%s'", ma.Member.Value, recordVal.RecordType.Name)
	}

	// Task 9.68: Check if it's a ClassInfoValue (class type identifier)
	// This handles cases like TObj.Create where TObj was evaluated to a ClassInfoValue
	// Task 9.73.5: Also check for ClassValue (metaclass reference)
	var classInfo *ClassInfo
	if classInfoVal, ok := objVal.(*ClassInfoValue); ok {
		classInfo = classInfoVal.ClassInfo
	} else if classVal, ok := objVal.(*ClassValue); ok {
		classInfo = classVal.ClassInfo
	}

	if classInfo != nil {
		memberName := ma.Member.Value

		// Task 9.73: Check for ClassName property in class/metaclass context (case-insensitive)
		if pkgident.Equal(memberName, "ClassName") {
			return &StringValue{Value: classInfo.Name}
		}

		// Task 9.7.2: Check for ClassType property (returns metaclass reference)
		if pkgident.Equal(memberName, "ClassType") {
			return &ClassValue{ClassInfo: classInfo}
		}

		// Try class variables first (case-insensitive)
		if classVarValue, ownerClass := classInfo.lookupClassVar(memberName); ownerClass != nil {
			return classVarValue
		}

		// Task 9.22: Try class constants
		if constValue := i.getClassConstant(classInfo, memberName, ma); constValue != nil {
			return constValue
		}

		// Task 9.13: Try class properties
		// Task 9.17: Also allow instance properties that use class constants or class variables
		if propInfo := classInfo.lookupProperty(memberName); propInfo != nil {
			if propInfo.IsClassProperty {
				return i.evalClassPropertyRead(classInfo, propInfo, ma)
			}
			// Task 9.17: Allow instance properties accessed on class if they use class-level read specs
			if result := i.canAccessInstancePropertyViaClass(classInfo, propInfo, ma); result != nil {
				return result
			}
		}

		// Try constructors (same logic as above for identifier check)
		// Task 9.82: Handle constructor overloads properly
		if classInfo.HasConstructor(memberName) {
			constructorOverloads := i.lookupConstructorOverloadsInHierarchy(classInfo, memberName)
			if len(constructorOverloads) > 0 {
				// Auto-invoke constructor (with or without parameters)
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

		// Try class methods
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

	// Task 9.1.1: Check if it's an interface instance
	// If so, extract the underlying object and delegate member access to it
	if intfInst, ok := objVal.(*InterfaceInstance); ok {
		if intfInst.Object == nil {
			return i.newErrorWithLocation(ma.Member, "Interface is nil")
		}

		// Verify the member exists in the interface definition
		// This ensures we only access members that are part of the interface contract
		memberName := ma.Member.Value

		propInfo := intfInst.Interface.GetProperty(memberName)
		hasMethod := intfInst.Interface.HasMethod(memberName)
		if !hasMethod && propInfo == nil {
			return i.newErrorWithLocation(ma, "member '%s' not found in interface '%s'", memberName, intfInst.Interface.Name)
		}

		// Get the underlying object to find the implementation
		underlyingObj, isObj := AsObject(intfInst.Object)
		if !isObj {
			return i.newErrorWithLocation(ma, "interface underlying object is not a class instance")
		}

		// Handle interface properties explicitly (classes may not declare matching properties)
		if propInfo != nil {
			return i.evalPropertyRead(underlyingObj, propInfo, ma)
		}

		// Method path: reuse object logic (auto-invoke parameterless, otherwise return function pointer)
		methodOverloads := i.getMethodOverloadsInHierarchy(underlyingObj.Class, memberName, false)
		classMethodOverloads := i.getMethodOverloadsInHierarchy(underlyingObj.Class, memberName, true)

		if len(methodOverloads) == 0 && len(classMethodOverloads) == 0 {
			return i.newErrorWithLocation(ma, "method '%s' declared in interface '%s' but not implemented by class '%s'",
				memberName, intfInst.Interface.Name, underlyingObj.Class.Name)
		}

		var method *ast.FunctionDecl
		if len(methodOverloads) > 0 {
			method = methodOverloads[0]
		} else {
			method = classMethodOverloads[0]
		}

		// If the interface method is a function with no parameters and accessed without (),
		// auto-invoke it to match DWScript behavior for value-returning methods.
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

		// Return function pointer bound to the underlying object
		// RefCount handling is performed when storing the pointer (see evalSimpleAssignment)
		return NewFunctionPointerValue(method, i.env, underlyingObj, pointerType)
	}

	// Task 9.5: Check if this is a type cast value (e.g., TBase(child).ClassVar)
	// For class variables, we must use the static type, not the runtime type
	var staticClassType *ClassInfo
	if typeCast, isTypeCast := objVal.(*TypeCastValue); isTypeCast {
		staticClassType = typeCast.StaticType
		objVal = typeCast.Object // Unwrap to the actual object
	}

	// Check if it's an object instance
	obj, ok := AsObject(objVal)

	// Task 9.5.4: Handle class variable access via nil instance
	// When accessing o.Member where o is nil, check if Member is a class variable first
	nilVal, isNilValue := objVal.(*NilValue)
	if !ok && (objVal.Type() == "NIL" || isNilValue) {
		memberName := ma.Member.Value

		// TObject.Free is nil-safe even when accessed without parentheses
		if pkgident.Equal(memberName, "Free") {
			return &NilValue{}
		}

		// Task 9.5: If we have a static type from a cast (e.g., TBase(nil).ClassVar), use it
		if staticClassType != nil {
			// Use the static type from the cast
			if classVarValue, ownerClass := staticClassType.lookupClassVar(memberName); ownerClass != nil {
				return classVarValue
			}
		} else if isNilValue && nilVal.ClassType != "" {
			// Object is nil - check if this is a typed nil value with a ClassType
			className := nilVal.ClassType

			// Look up the class by name (case-insensitive)
			for registeredClassName, classInfo := range i.classes {
				if pkgident.Equal(registeredClassName, className) {
					// Check if this member is a class variable
					if classVarValue, ownerClass := classInfo.lookupClassVar(memberName); ownerClass != nil {
						return classVarValue
					}
					// Member exists but is not a class variable - fall through to error
					break
				}
			}
		}

		// Task 9.7: If we couldn't find a class variable, raise "Object not instantiated"
		// This happens when accessing instance members on nil objects
		message := fmt.Sprintf("Object not instantiated [line: %d, column: %d]", ma.Token.Pos.Line, ma.Token.Pos.Column+1)
		i.raiseException("Exception", message, &ma.Token.Pos)
		return &NilValue{}
	}

	if !ok {
		// Check if it's an enum value with .Value or .ToString property
		if enumVal, isEnum := objVal.(*EnumValue); isEnum {
			memberName := pkgident.Normalize(ma.Member.Value)
			if memberName == "value" {
				// Return the ordinal value as an integer
				return &IntegerValue{Value: int64(enumVal.OrdinalValue)}
			}
			// Note: .ToString will be handled by helpers if needed
		}

		// Not an object - check if helpers provide this member
		helper, helperProp := i.findHelperProperty(objVal, ma.Member.Value)
		if helperProp != nil {
			return i.evalHelperPropertyRead(helper, helperProp, objVal, ma)
		}

		// Task 9.8.5: Check if it's a parameterless helper method that can be auto-invoked
		// This allows arr.Pop to work the same as arr.Pop()
		// Note: The semantic analyzer already validated that parameterless methods accessed
		// without () are allowed. If we reach here with a helper method found, and the
		// semantic analysis passed, then this must be a parameterless method access.
		helper, methodDecl, builtinSpec := i.findHelperMethod(objVal, ma.Member.Value)
		if helper != nil && (methodDecl != nil || builtinSpec != "") {
			// Check if it's parameterless
			isParameterless := false
			if methodDecl != nil {
				// AST-declared method - check parameter count
				isParameterless = len(methodDecl.Parameters) == 0
			} else if builtinSpec != "" {
				// Builtin-only method - check parameter count from builtin spec
				isParameterless = i.isBuiltinMethodParameterless(builtinSpec)
			}

			if isParameterless {
				// Auto-invoke the parameterless method
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

		return i.newErrorWithLocation(ma, "cannot access member '%s' of type '%s' (no helper found)",
			ma.Member.Value, objVal.Type())
	}

	memberName := ma.Member.Value

	// Prevent member access on destroyed instances
	if obj.Destroyed {
		message := fmt.Sprintf("Object already destroyed [line: %d, column: %d]", ma.Token.Pos.Line, ma.Token.Pos.Column)
		i.raiseException("Exception", message, &ma.Token.Pos)
		return &NilValue{}
	}

	// Handle built-in properties/methods available on all objects (inherited from TObject)
	if pkgident.Equal(memberName, "ClassName") {
		// ClassName returns the runtime type name of the object
		return &StringValue{Value: obj.Class.Name}
	}
	// Task 9.7.2: ClassType returns metaclass reference for the object's runtime type
	if pkgident.Equal(memberName, "ClassType") {
		return &ClassValue{ClassInfo: obj.Class}
	}

	// Check if this is a property access (properties take precedence over fields)
	if propInfo := obj.Class.lookupProperty(memberName); propInfo != nil {
		return i.evalPropertyRead(obj, propInfo, ma)
	}

	// Not a property - try direct field access
	fieldValue := obj.GetField(memberName)
	if fieldValue == nil {
		// Task 9.5.4: Try class variables (accessible from instance)
		// Task 9.5: Use static type from cast if available (e.g., TBase(child).ClassVar)
		classForLookup := obj.Class
		if staticClassType != nil {
			classForLookup = staticClassType
		}
		if classVarValue, ownerClass := classForLookup.lookupClassVar(memberName); ownerClass != nil {
			return classVarValue
		}

		// Task 9.22: Try class constants (accessible from instance)
		if constValue := i.getClassConstant(obj.Class, memberName, ma); constValue != nil {
			return constValue
		}

		// Check if it's a method
		// Task 9.16.2: Method names are case-insensitive, normalize to lowercase
		// Task 9.67: Check MethodOverloads for overloaded methods, not just Methods map
		methodOverloads := i.getMethodOverloadsInHierarchy(obj.Class, memberName, false)

		// Task 9.7: Also check for class methods (which can be called on instances)
		classMethodOverloads := i.getMethodOverloadsInHierarchy(obj.Class, memberName, true)

		if len(methodOverloads) > 0 || len(classMethodOverloads) > 0 {
			// Check if any overload has 0 parameters and can be auto-invoked
			hasParameterlessOverload := false

			// Check instance methods first
			for _, overload := range methodOverloads {
				if len(overload.Parameters) == 0 {
					hasParameterlessOverload = true
					break
				}
			}

			// If no parameterless instance method, check class methods
			if !hasParameterlessOverload {
				for _, overload := range classMethodOverloads {
					if len(overload.Parameters) == 0 {
						hasParameterlessOverload = true
						break
					}
				}
			}

			// If there's a parameterless overload, auto-invoke it
			// This allows DWScript syntax: obj.Method instead of obj.Method()
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

			// No parameterless overload - return the first overload as method pointer
			// (In practice, DWScript requires explicit calls for methods with parameters)
			// Prefer instance methods over class methods
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

		// Check if helpers provide this member
		helper, helperProp := i.findHelperProperty(obj, memberName)
		if helperProp != nil {
			return i.evalHelperPropertyRead(helper, helperProp, obj, ma)
		}
		return i.newErrorWithLocation(ma, "field '%s' not found in class '%s'", memberName, obj.Class.Name)
	}

	return fieldValue
}

// lookupConstructorInHierarchy searches for a constructor by name in the class hierarchy.
// It walks the parent chain starting from the given class.
// Returns the constructor declaration, or nil if not found.
// Task 9.82: Updated to return all constructor overloads instead of just one
// Task 9.82: Case-insensitive lookup (DWScript is case-insensitive)
func (i *Interpreter) lookupConstructorOverloadsInHierarchy(classInfo *ClassInfo, name string) []*ast.FunctionDecl {
	for current := classInfo; current != nil; current = current.Parent {
		// Check overload set first (case-insensitive)
		for ctorName, overloads := range current.ConstructorOverloads {
			if pkgident.Equal(ctorName, name) && len(overloads) > 0 {
				return overloads
			}
		}
		// Fallback to single constructor (case-insensitive)
		for ctorName, constructor := range current.Constructors {
			if pkgident.Equal(ctorName, name) {
				return []*ast.FunctionDecl{constructor}
			}
		}
	}
	return nil
}

// Deprecated: Use lookupConstructorOverloadsInHierarchy instead
// Kept for backwards compatibility with existing code
func (i *Interpreter) lookupConstructorInHierarchy(classInfo *ClassInfo, name string) *ast.FunctionDecl {
	overloads := i.lookupConstructorOverloadsInHierarchy(classInfo, name)
	if len(overloads) > 0 {
		return overloads[0]
	}
	return nil
}

// lookupClassMethodInHierarchy searches for a class method by name in the class hierarchy.
// It walks the parent chain starting from the given class.
// Returns the method declaration, or nil if not found.
// Task 9.16.2: Method names are case-insensitive, so we normalize to lowercase
func (i *Interpreter) lookupClassMethodInHierarchy(classInfo *ClassInfo, name string) *ast.FunctionDecl {
	normalizedName := pkgident.Normalize(name)
	for current := classInfo; current != nil; current = current.Parent {
		if method, exists := current.ClassMethods[normalizedName]; exists {
			return method
		}
	}
	return nil
}

// bindClassConstantsToEnv adds all class constants from the given ClassInfo to the current environment.
// This allows methods to access class constants directly by name without qualification.
func (i *Interpreter) bindClassConstantsToEnv(classInfo *ClassInfo) {
	for constName, constValue := range classInfo.ConstantValues {
		i.env.Define(constName, constValue)
	}
}

// evalSelfExpression evaluates a Self expression.
// Self refers to the current instance (in instance methods) or
// the current class (in class methods).
// Task 9.7: Implement Self keyword
func (i *Interpreter) evalSelfExpression(se *ast.SelfExpression) Value {
	// Get Self from the environment (should be bound when entering methods)
	selfVal, exists := i.env.Get("Self")
	if !exists {
		return i.newErrorWithLocation(se, "Self used outside method context")
	}

	return selfVal
}

// evalInheritedExpression evaluates an inherited method call.
// Syntax: inherited MethodName(args) or inherited (bare, calls same method in parent)
// Task 9.164: Implement inherited keyword
func (i *Interpreter) evalInheritedExpression(ie *ast.InheritedExpression) Value {
	// Get current Self (must be in a method context)
	selfVal, exists := i.env.Get("Self")
	if !exists {
		return i.newErrorWithLocation(ie, "inherited can only be used inside a method")
	}

	obj, ok := selfVal.(*ObjectInstance)
	if !ok {
		return i.newErrorWithLocation(ie, "inherited requires Self to be an object instance")
	}

	// Determine the current static class context for inherited resolution.
	// Prefer __CurrentClass__ (set when entering a method), fall back to runtime class.
	classInfo := obj.Class
	if currentClassVal, has := i.env.Get("__CurrentClass__"); has {
		if civ, isClassVal := currentClassVal.(*ClassInfoValue); isClassVal && civ.ClassInfo != nil {
			classInfo = civ.ClassInfo
		}
	}

	// Get the parent class
	if classInfo.Parent == nil {
		return i.newErrorWithLocation(ie, "class '%s' has no parent class", classInfo.Name)
	}

	parentClass := classInfo.Parent

	// Determine which method to call
	var methodName string
	if ie.Method != nil {
		// Explicit method name provided: inherited MethodName(args)
		methodName = ie.Method.Value
	} else {
		// Bare inherited: need to get the current method name from environment
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

	// Task 9.16.4.2: Look up member in parent class (method, property, or field)
	// Try method first (case-insensitive)
	methodOverloads := i.getMethodOverloadsInHierarchy(parentClass, methodName, false)
	if len(methodOverloads) > 0 {
		parentMethod, err := i.resolveMethodOverload(parentClass.Name, methodName, methodOverloads, ie.Arguments)
		if err != nil {
			return i.newErrorWithLocation(ie, "%s", err.Error())
		}

		// Found a method - evaluate it
		// Evaluate arguments
		args := make([]Value, len(ie.Arguments))
		for idx, arg := range ie.Arguments {
			val := i.Eval(arg)
			if isError(val) {
				return val
			}
			args[idx] = val
		}

		// Check argument count matches parameter count
		if len(args) != len(parentMethod.Parameters) {
			return i.newErrorWithLocation(ie, "wrong number of arguments for method '%s': expected %d, got %d",
				methodName, len(parentMethod.Parameters), len(args))
		}

		// Create method environment (with Self binding)
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the current object
		i.env.Define("Self", obj)

		// Bind __CurrentClass__ to parent class
		i.env.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: parentClass})

		// Add class constants to method scope so they can be accessed directly
		i.bindClassConstantsToEnv(parentClass)

		// Bind __CurrentMethod__ for nested inherited calls
		i.env.Define("__CurrentMethod__", &StringValue{Value: methodName})

		// Bind method parameters to arguments with implicit conversion
		for idx, param := range parentMethod.Parameters {
			arg := args[idx]

			// Apply implicit conversion if parameter has a type and types don't match
			if param.Type != nil {
				paramTypeName := param.Type.String()
				if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
					arg = converted
				}
			}

			i.env.Define(param.Name.Value, arg)
		}

		// For functions (not procedures), initialize the Result variable
		// Task 9.221: Use appropriate default value based on return type
		if parentMethod.ReturnType != nil {
			returnType := i.resolveTypeFromAnnotation(parentMethod.ReturnType)
			defaultVal := i.getDefaultValue(returnType)
			i.env.Define("Result", defaultVal)
			// In DWScript, the method name can be used as an alias for Result
			i.env.Define(parentMethod.Name.Value, &ReferenceValue{Env: i.env, VarName: "Result"})
		}

		// Execute parent method body
		_ = i.Eval(parentMethod.Body)

		// Handle function return value
		var returnValue Value
		if parentMethod.ReturnType != nil {
			// For functions, check if Result was set
			if resultVal, ok := i.env.Get("Result"); ok {
				returnValue = resultVal
			} else {
				// Check if the method name was used as return value (DWScript style)
				if methodVal, ok := i.env.Get(parentMethod.Name.Value); ok {
					returnValue = methodVal
				} else {
					returnValue = &NilValue{}
				}
			}
		} else {
			// Procedure - no return value
			returnValue = &NilValue{}
		}

		// Restore environment
		i.env = savedEnv

		return returnValue
	}

	// Task 9.16.4.2: Try properties (case-insensitive)
	var propInfo *types.PropertyInfo
	for name, prop := range parentClass.Properties {
		if pkgident.Equal(name, methodName) {
			propInfo = prop
			break
		}
	}

	if propInfo != nil {
		// Found a property - read it
		if len(ie.Arguments) > 0 || ie.IsCall {
			return i.newErrorWithLocation(ie, "cannot call property '%s' as a method", methodName)
		}
		// Evaluate property read expression
		// The property's read expression is evaluated in the context of the parent class
		return i.evalPropertyRead(obj, propInfo, ie)
	}

	// Task 9.16.4.2: Try fields (case-insensitive)
	for name := range parentClass.Fields {
		if pkgident.Equal(name, methodName) {
			if len(ie.Arguments) > 0 || ie.IsCall {
				return i.newErrorWithLocation(ie, "cannot call field '%s' as a method", methodName)
			}
			// Return field value
			fieldValue := obj.GetField(name)
			if fieldValue == nil {
				return &NilValue{}
			}
			return fieldValue
		}
	}

	// Member not found in parent class
	return i.newErrorWithLocation(ie, "method, property, or field '%s' not found in parent class '%s'", methodName, parentClass.Name)
}

// getClassConstant retrieves a class constant value by name.
// It evaluates the constant expression lazily (on first access) and caches the result.
// Supports inheritance by searching up the class hierarchy.
// Returns nil if the constant doesn't exist.
// Task 9.22: Support class constant evaluation with visibility enforcement and inheritance.
func (i *Interpreter) getClassConstant(classInfo *ClassInfo, constantName string, ma *ast.MemberAccessExpression) Value {
	// Look up constant in hierarchy (supports inheritance)
	constDecl, ownerClass := classInfo.lookupConstant(constantName)
	if constDecl == nil {
		return nil
	}

	// Check if we've already evaluated this constant (check in the owner class)
	if cachedValue, cached := ownerClass.ConstantValues[constantName]; cached {
		return cachedValue
	}

	// Create a temporary environment for evaluating the constant expression
	// This allows constants to reference other already-evaluated constants in the same class
	savedEnv := i.env
	tempEnv := NewEnclosedEnvironment(i.env)

	// Add all ALREADY EVALUATED class constants to the temporary environment
	// This prevents infinite recursion
	for constName, constVal := range ownerClass.ConstantValues {
		if constName != constantName && constVal != nil {
			tempEnv.Define(constName, constVal)
		}
	}

	i.env = tempEnv

	// Evaluate the constant expression
	constValue := i.Eval(constDecl.Value)

	// Restore environment
	i.env = savedEnv

	if isError(constValue) {
		return constValue
	}

	// Cache the evaluated value for future access (in the owner class)
	ownerClass.ConstantValues[constantName] = constValue

	return constValue
}

// canAccessInstancePropertyViaClass checks if an instance property can be accessed
// through the class itself (e.g., TClass.PropertyName) when the property's read spec
// refers to a class variable or constant.
// Task 9.17: Extracted helper to avoid code duplication and ensure case-insensitive lookups.
func (i *Interpreter) canAccessInstancePropertyViaClass(classInfo *ClassInfo, propInfo *types.PropertyInfo, ma *ast.MemberAccessExpression) Value {
	// Only field-based properties can potentially use class-level read specs
	if propInfo.ReadKind != types.PropAccessField {
		return nil
	}

	// Check if read spec is a class variable (case-insensitive)
	if _, ownerClass := classInfo.lookupClassVar(propInfo.ReadSpec); ownerClass != nil {
		// Create a temporary object instance to evaluate the property
		tempObj := &ObjectInstance{Class: classInfo, Fields: make(map[string]Value)}
		return i.evalPropertyRead(tempObj, propInfo, ma)
	}

	// Check if read spec is a constant (case-insensitive)
	if _, ownerClass := classInfo.lookupConstant(propInfo.ReadSpec); ownerClass != nil {
		// Create a temporary object instance to evaluate the property
		tempObj := &ObjectInstance{Class: classInfo, Fields: make(map[string]Value)}
		return i.evalPropertyRead(tempObj, propInfo, ma)
	}

	// Property read spec is not a class-level member
	return nil
}
