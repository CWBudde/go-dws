package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// buildIndexDirectiveArgs converts a property's index directive into runtime arguments.
// Currently only integer index directives are supported.
func (i *Interpreter) buildIndexDirectiveArgs(propInfo *types.PropertyInfo) ([]Value, error) {
	if propInfo == nil || !propInfo.HasIndexValue {
		return nil, nil
	}

	if propInfo.IndexValueType != nil && propInfo.IndexValueType.Equals(types.INTEGER) {
		if intVal, ok := propInfo.IndexValue.(int64); ok {
			return []Value{NewIntegerValue(intVal)}, nil
		}
	}

	return nil, fmt.Errorf("property '%s' has unsupported index directive", propInfo.Name)
}

// evalPropertyRead evaluates a property read access.
// Handles field-backed, method-backed, and expression-backed properties.
func (i *Interpreter) evalPropertyRead(obj *ObjectInstance, propInfo *types.PropertyInfo, node ast.Node) Value {
	// Initialize property evaluation context if needed
	if i.propContext == nil {
		i.propContext = &PropertyEvalContext{
			PropertyChain: make([]string, 0),
		}
	}

	// Check for circular property references
	for _, prop := range i.propContext.PropertyChain {
		if prop == propInfo.Name {
			return i.newErrorWithLocation(node, "circular property reference detected: %s", propInfo.Name)
		}
	}

	// Push property onto chain
	i.propContext.PropertyChain = append(i.propContext.PropertyChain, propInfo.Name)
	defer func() {
		// Pop property from chain when done
		if len(i.propContext.PropertyChain) > 0 {
			i.propContext.PropertyChain = i.propContext.PropertyChain[:len(i.propContext.PropertyChain)-1]
		}
		// Clear context if chain is empty
		if len(i.propContext.PropertyChain) == 0 {
			i.propContext = nil
		}
	}()

	switch propInfo.ReadKind {
	case types.PropAccessField:
		// Field, constant, class var, or method access - check at runtime which it is
		// Task 9.17: Check in order: class vars → constants → instance fields
		// This matches the semantic analyzer's lookup order

		// 1. Try as a class variable (case-insensitive)
		if classVarValue, ownerClass := obj.Class.LookupClassVar(propInfo.ReadSpec); ownerClass != nil {
			return classVarValue
		}

		// 2. Try as a constant (case-insensitive, with lazy evaluation)
		// Note: getClassConstant doesn't currently use the ma parameter for error reporting
		concreteClass, ok := obj.Class.(*ClassInfo)
		if ok {
			if constValue := i.getClassConstant(concreteClass, propInfo.ReadSpec, nil); constValue != nil {
				return constValue
			}
		}

		// 3. Try as an instance field
		fields := obj.Class.GetFieldsMap()
		if fields != nil && fields[propInfo.ReadSpec] != nil {
			if propInfo.HasIndexValue {
				return i.newErrorWithLocation(node, "property '%s' uses an index directive and cannot be field-backed", propInfo.Name)
			}
			fieldValue := obj.GetField(propInfo.ReadSpec)
			if fieldValue == nil {
				return i.newErrorWithLocation(node, "property '%s' read field '%s' not found", propInfo.Name, propInfo.ReadSpec)
			}
			return fieldValue
		}

		// Not a field, class var, or constant - try as a method (getter)
		method := obj.Class.LookupMethod(propInfo.ReadSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "property '%s' read specifier '%s' not found as field, constant, class var, or method", propInfo.Name, propInfo.ReadSpec)
		}

		// Indexed properties must be accessed with index syntax
		if propInfo.IsIndexed {
			return i.newErrorWithLocation(node, "indexed property '%s' requires index arguments (e.g., obj.%s[index])", propInfo.Name, propInfo.Name)
		}

		// Build implicit index arguments from directive (if any)
		indexArgs, err := i.buildIndexDirectiveArgs(propInfo)
		if err != nil {
			return i.newErrorWithLocation(node, "%s", err.Error())
		}

		if len(method.Parameters) != len(indexArgs) {
			return i.newErrorWithLocation(node, "property '%s' getter method '%s' expects %d parameter(s), but index directive supplies %d",
				propInfo.Name, propInfo.ReadSpec, len(method.Parameters), len(indexArgs))
		}

		// Call the getter method
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the object
		i.env.Define("Self", obj)

		// Bind implicit index directive arguments, if present
		for idx, param := range method.Parameters {
			if idx < len(indexArgs) {
				i.env.Define(param.Name.Value, indexArgs[idx])
			}
		}

		// For functions, initialize the Result variable
		// Use appropriate default value based on return type
		if method.ReturnType != nil {
			returnType := i.resolveTypeFromAnnotation(method.ReturnType)
			defaultVal := i.getDefaultValue(returnType)
			i.env.Define("Result", defaultVal)
			// In DWScript, the method name can be used as an alias for Result
			i.env.Define(method.Name.Value, &ReferenceValue{Env: i.env, VarName: "Result"})
		}

		// Set flag to indicate we're inside a property getter
		savedInGetter := i.propContext.InPropertyGetter
		i.propContext.InPropertyGetter = true
		defer func() {
			i.propContext.InPropertyGetter = savedInGetter
		}()

		// Execute method body
		i.Eval(method.Body)

		// Get return value
		var returnValue Value
		if method.ReturnType != nil {
			if resultVal, ok := i.env.Get("Result"); ok {
				// Dereference ReferenceValue if needed
				if refVal, isRef := resultVal.(*ReferenceValue); isRef {
					if derefVal, err := refVal.Dereference(); err == nil {
						resultVal = derefVal
					}
				}
				returnValue = resultVal
			} else if methodNameVal, ok := i.env.Get(method.Name.Value); ok {
				// Dereference ReferenceValue if needed
				if refVal, isRef := methodNameVal.(*ReferenceValue); isRef {
					if derefVal, err := refVal.Dereference(); err == nil {
						methodNameVal = derefVal
					}
				}
				returnValue = methodNameVal
			} else {
				returnValue = &NilValue{}
			}
		} else {
			returnValue = &NilValue{}
		}

		// Restore environment
		i.env = savedEnv

		return returnValue

	case types.PropAccessMethod:
		// Indexed properties must be accessed with index syntax
		if propInfo.IsIndexed {
			return i.newErrorWithLocation(node, "indexed property '%s' requires index arguments (e.g., obj.%s[index])", propInfo.Name, propInfo.Name)
		}

		// Check if method exists
		method := obj.Class.LookupMethod(propInfo.ReadSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "property '%s' getter method '%s' not found", propInfo.Name, propInfo.ReadSpec)
		}

		// Build implicit index directive arguments, if any
		indexArgs, err := i.buildIndexDirectiveArgs(propInfo)
		if err != nil {
			return i.newErrorWithLocation(node, "%s", err.Error())
		}
		if len(method.Parameters) != len(indexArgs) {
			return i.newErrorWithLocation(node, "property '%s' getter method '%s' expects %d parameter(s), but index directive supplies %d",
				propInfo.Name, propInfo.ReadSpec, len(method.Parameters), len(indexArgs))
		}

		// Call the getter method with no arguments
		// Create method environment with Self bound to object
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the object
		i.env.Define("Self", obj)

		// Bind implicit index directive arguments, if present
		for idx, param := range method.Parameters {
			if idx < len(indexArgs) {
				i.env.Define(param.Name.Value, indexArgs[idx])
			}
		}

		// For functions, initialize the Result variable
		// Task 9.221: Use appropriate default value based on return type
		if method.ReturnType != nil {
			returnType := i.resolveTypeFromAnnotation(method.ReturnType)
			defaultVal := i.getDefaultValue(returnType)
			i.env.Define("Result", defaultVal)
			// In DWScript, the method name can be used as an alias for Result
			i.env.Define(method.Name.Value, &ReferenceValue{Env: i.env, VarName: "Result"})
		}

		// Task 9.32c: Set flag to indicate we're inside a property getter
		savedInGetter := i.propContext.InPropertyGetter
		i.propContext.InPropertyGetter = true
		defer func() {
			i.propContext.InPropertyGetter = savedInGetter
		}()

		// Execute method body
		i.Eval(method.Body)

		// Get return value
		var returnValue Value
		if method.ReturnType != nil {
			if resultVal, ok := i.env.Get("Result"); ok {
				// Dereference ReferenceValue if needed
				if refVal, isRef := resultVal.(*ReferenceValue); isRef {
					if derefVal, err := refVal.Dereference(); err == nil {
						resultVal = derefVal
					}
				}
				returnValue = resultVal
			} else if methodNameVal, ok := i.env.Get(method.Name.Value); ok {
				// Dereference ReferenceValue if needed
				if refVal, isRef := methodNameVal.(*ReferenceValue); isRef {
					if derefVal, err := refVal.Dereference(); err == nil {
						methodNameVal = derefVal
					}
				}
				returnValue = methodNameVal
			} else {
				returnValue = &NilValue{}
			}
		} else {
			returnValue = &NilValue{}
		}

		// Restore environment
		i.env = savedEnv

		return returnValue

	case types.PropAccessExpression:
		// Task 9.3c: Expression access - evaluate expression in context of object
		// Retrieve the AST expression from PropertyInfo
		if propInfo.ReadExpr == nil {
			return i.newErrorWithLocation(node, "property '%s' has expression-based getter but no expression stored", propInfo.Name)
		}

		// Type-assert to ast.Expression
		exprNode, ok := propInfo.ReadExpr.(ast.Expression)
		if !ok {
			return i.newErrorWithLocation(node, "property '%s' has invalid expression type", propInfo.Name)
		}

		// Unwrap GroupedExpression if present (parser wraps expressions in parentheses)
		if groupedExpr, ok := exprNode.(*ast.GroupedExpression); ok {
			exprNode = groupedExpr.Expression
		}

		// Create new environment with Self bound to object
		exprEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = exprEnv

		// Bind Self to the object instance
		i.env.Define("Self", obj)

		// Bind all object fields to environment so they can be accessed directly
		// This allows expressions like (FWidth * FHeight) to work
		for fieldName, fieldValue := range obj.Fields {
			i.env.Define(fieldName, fieldValue)
		}

		// Evaluate the expression AST node
		result := i.Eval(exprNode)

		// Restore environment
		i.env = savedEnv

		return result

	default:
		return i.newErrorWithLocation(node, "property '%s' has no read access", propInfo.Name)
	}
}

// evalClassPropertyRead evaluates a class property read operation: TClass.PropertyName
// Task 9.13: Handles reading class (static) properties.
func (i *Interpreter) evalClassPropertyRead(classInfo *ClassInfo, propInfo *types.PropertyInfo, node ast.Node) Value {
	// Indexed properties must be accessed with index syntax
	if propInfo.IsIndexed {
		return i.newErrorWithLocation(node, "indexed class property '%s' requires index arguments", propInfo.Name)
	}

	switch propInfo.ReadKind {
	case types.PropAccessField:
		// Field or method access - check at runtime which it is
		// First try as a class variable
		if classVarValue, exists := classInfo.ClassVars[propInfo.ReadSpec]; exists {
			return classVarValue
		}

		// Not a class variable - try as a class method
		method := i.lookupClassMethodInHierarchy(classInfo, propInfo.ReadSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "class property '%s' read specifier '%s' not found as class variable or class method", propInfo.Name, propInfo.ReadSpec)
		}

		// Call the class method getter
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind all class variables to environment so they can be accessed directly
		for classVarName, classVarValue := range classInfo.ClassVars {
			i.env.Define(classVarName, classVarValue)
		}

		// For functions, initialize the Result variable
		if method.ReturnType != nil {
			returnType := i.resolveTypeFromAnnotation(method.ReturnType)
			defaultVal := i.getDefaultValue(returnType)
			i.env.Define("Result", defaultVal)
			// In DWScript, the method name can be used as an alias for Result
			i.env.Define(method.Name.Value, &ReferenceValue{Env: i.env, VarName: "Result"})
		}

		// Execute method body
		i.Eval(method.Body)

		// Get return value
		var returnValue Value
		if method.ReturnType != nil {
			if resultVal, ok := i.env.Get("Result"); ok {
				// Dereference ReferenceValue if needed
				if refVal, isRef := resultVal.(*ReferenceValue); isRef {
					if derefVal, err := refVal.Dereference(); err == nil {
						resultVal = derefVal
					}
				}
				returnValue = resultVal
			} else if methodNameVal, ok := i.env.Get(method.Name.Value); ok {
				// Dereference ReferenceValue if needed
				if refVal, isRef := methodNameVal.(*ReferenceValue); isRef {
					if derefVal, err := refVal.Dereference(); err == nil {
						methodNameVal = derefVal
					}
				}
				returnValue = methodNameVal
			} else {
				returnValue = &NilValue{}
			}
		} else {
			returnValue = &NilValue{}
		}

		// Restore environment
		i.env = savedEnv

		return returnValue

	case types.PropAccessMethod:
		// Call the class method getter
		method := i.lookupClassMethodInHierarchy(classInfo, propInfo.ReadSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "class property '%s' getter method '%s' not found", propInfo.Name, propInfo.ReadSpec)
		}

		// Create method environment (no Self binding for class methods)
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind all class variables to environment so they can be accessed directly
		for classVarName, classVarValue := range classInfo.ClassVars {
			i.env.Define(classVarName, classVarValue)
		}

		// For functions, initialize the Result variable
		if method.ReturnType != nil {
			returnType := i.resolveTypeFromAnnotation(method.ReturnType)
			defaultVal := i.getDefaultValue(returnType)
			i.env.Define("Result", defaultVal)
			// In DWScript, the method name can be used as an alias for Result
			i.env.Define(method.Name.Value, &ReferenceValue{Env: i.env, VarName: "Result"})
		}

		// Execute method body
		i.Eval(method.Body)

		// Get return value
		var returnValue Value
		if method.ReturnType != nil {
			if resultVal, ok := i.env.Get("Result"); ok {
				// Dereference ReferenceValue if needed
				if refVal, isRef := resultVal.(*ReferenceValue); isRef {
					if derefVal, err := refVal.Dereference(); err == nil {
						resultVal = derefVal
					}
				}
				returnValue = resultVal
			} else if methodNameVal, ok := i.env.Get(method.Name.Value); ok {
				// Dereference ReferenceValue if needed
				if refVal, isRef := methodNameVal.(*ReferenceValue); isRef {
					if derefVal, err := refVal.Dereference(); err == nil {
						methodNameVal = derefVal
					}
				}
				returnValue = methodNameVal
			} else {
				returnValue = &NilValue{}
			}
		} else {
			returnValue = &NilValue{}
		}

		// Restore environment
		i.env = savedEnv

		return returnValue

	default:
		return i.newErrorWithLocation(node, "class property '%s' has no read access", propInfo.Name)
	}
}

// evalClassPropertyWrite evaluates a class property write operation: TClass.PropertyName := value
// Task 9.14: Handles writing to class (static) properties.
func (i *Interpreter) evalClassPropertyWrite(classInfo *ClassInfo, propInfo *types.PropertyInfo, value Value, node ast.Node) Value {
	// Indexed properties must be written with index syntax
	if propInfo.IsIndexed {
		return i.newErrorWithLocation(node, "indexed class property '%s' requires index arguments", propInfo.Name)
	}

	// Check if property has write access
	if propInfo.WriteKind == types.PropAccessNone {
		return i.newErrorWithLocation(node, "class property '%s' is read-only", propInfo.Name)
	}

	switch propInfo.WriteKind {
	case types.PropAccessField:
		// Field or method access - check at runtime which it is
		// First try as a class variable
		if _, exists := classInfo.ClassVars[propInfo.WriteSpec]; exists {
			classInfo.ClassVars[propInfo.WriteSpec] = value
			return value
		}

		// Not a class variable - try as a class method
		method := i.lookupClassMethodInHierarchy(classInfo, propInfo.WriteSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "class property '%s' write specifier '%s' not found as class variable or class method", propInfo.Name, propInfo.WriteSpec)
		}

		// Call the class method setter
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind all class variables to environment so they can be accessed directly
		for classVarName, classVarValue := range classInfo.ClassVars {
			i.env.Define(classVarName, classVarValue)
		}

		// Bind the value parameter
		if len(method.Parameters) > 0 {
			i.env.Define(method.Parameters[0].Name.Value, value)
		}

		// Execute method body
		i.Eval(method.Body)

		// Update class variables from environment (in case they were modified)
		for classVarName := range classInfo.ClassVars {
			if val, ok := i.env.Get(classVarName); ok {
				classInfo.ClassVars[classVarName] = val
			}
		}

		// Restore environment
		i.env = savedEnv

		return value

	case types.PropAccessMethod:
		// Call the class method setter
		method := i.lookupClassMethodInHierarchy(classInfo, propInfo.WriteSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "class property '%s' setter method '%s' not found", propInfo.Name, propInfo.WriteSpec)
		}

		// Create method environment (no Self binding for class methods)
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind all class variables to environment so they can be accessed directly
		for classVarName, classVarValue := range classInfo.ClassVars {
			i.env.Define(classVarName, classVarValue)
		}

		// Bind the value parameter
		if len(method.Parameters) > 0 {
			i.env.Define(method.Parameters[0].Name.Value, value)
		}

		// Execute method body
		i.Eval(method.Body)

		// Update class variables from environment (in case they were modified)
		for classVarName := range classInfo.ClassVars {
			if val, ok := i.env.Get(classVarName); ok {
				classInfo.ClassVars[classVarName] = val
			}
		}

		// Restore environment
		i.env = savedEnv

		return value

	default:
		return i.newErrorWithLocation(node, "class property '%s' has no write access", propInfo.Name)
	}
}

// evalIndexedPropertyRead evaluates an indexed property read operation: obj.Property[index]
// Support indexed property reads end-to-end.
// Calls the property getter method with index parameter(s).
func (i *Interpreter) evalIndexedPropertyRead(obj *ObjectInstance, propInfo *types.PropertyInfo, indices []Value, node ast.Node) Value {
	// Note: PropAccessKind is set to PropAccessField at registration time for both fields and methods
	// We need to check at runtime whether it's actually a field or method
	switch propInfo.ReadKind {
	case types.PropAccessField, types.PropAccessMethod:
		// Check if it's actually a field (not allowed for indexed properties)
		fields := obj.Class.GetFieldsMap()
		if fields != nil && fields[propInfo.ReadSpec] != nil {
			return i.newErrorWithLocation(node, "indexed property '%s' requires a getter method, not a field", propInfo.Name)
		}

		// Look up the getter method
		method := obj.Class.LookupMethod(propInfo.ReadSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "indexed property '%s' getter method '%s' not found", propInfo.Name, propInfo.ReadSpec)
		}

		// Verify method has correct number of parameters (index params, no value param)
		expectedParamCount := len(indices)
		if len(method.Parameters) != expectedParamCount {
			return i.newErrorWithLocation(node, "indexed property '%s' getter method '%s' expects %d parameter(s), got %d index argument(s)",
				propInfo.Name, propInfo.ReadSpec, len(method.Parameters), len(indices))
		}

		// Create method environment
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the object
		i.env.Define("Self", obj)

		// Bind index parameters
		for idx, param := range method.Parameters {
			if idx < len(indices) {
				i.env.Define(param.Name.Value, indices[idx])
			}
		}

		// For functions, initialize the Result variable
		// Task 9.221: Use appropriate default value based on return type
		if method.ReturnType != nil {
			returnType := i.resolveTypeFromAnnotation(method.ReturnType)
			defaultVal := i.getDefaultValue(returnType)
			i.env.Define("Result", defaultVal)
			// In DWScript, the method name can be used as an alias for Result
			i.env.Define(method.Name.Value, &ReferenceValue{Env: i.env, VarName: "Result"})
		}

		// Execute method body
		i.Eval(method.Body)

		// Get return value
		var returnValue Value
		if method.ReturnType != nil {
			if resultVal, ok := i.env.Get("Result"); ok {
				// Dereference ReferenceValue if needed
				if refVal, isRef := resultVal.(*ReferenceValue); isRef {
					if derefVal, err := refVal.Dereference(); err == nil {
						resultVal = derefVal
					}
				}
				returnValue = resultVal
			} else if methodNameVal, ok := i.env.Get(method.Name.Value); ok {
				// Dereference ReferenceValue if needed
				if refVal, isRef := methodNameVal.(*ReferenceValue); isRef {
					if derefVal, err := refVal.Dereference(); err == nil {
						methodNameVal = derefVal
					}
				}
				returnValue = methodNameVal
			} else {
				returnValue = &NilValue{}
			}
		} else {
			returnValue = &NilValue{}
		}

		// Restore environment
		i.env = savedEnv

		return returnValue

	case types.PropAccessExpression:
		// Expression-based indexed properties not supported yet
		return i.newErrorWithLocation(node, "expression-based indexed property getters not yet supported")

	default:
		return i.newErrorWithLocation(node, "indexed property '%s' has no read access", propInfo.Name)
	}
}

// evalIndexedPropertyWrite evaluates an indexed property write operation: obj.Property[index] := value
// Task 9.2b: Support indexed property writes.
// Calls the property setter method with index parameter(s) followed by the value.
func (i *Interpreter) evalIndexedPropertyWrite(obj *ObjectInstance, propInfo *types.PropertyInfo, indices []Value, value Value, node ast.Node) Value {
	// Note: PropAccessKind is set to PropAccessField at registration time for both fields and methods
	// We need to check at runtime whether it's actually a field or method
	switch propInfo.WriteKind {
	case types.PropAccessField, types.PropAccessMethod:
		// Check if it's actually a field (not allowed for indexed properties)
		fields := obj.Class.GetFieldsMap()
		if fields != nil && fields[propInfo.WriteSpec] != nil {
			return i.newErrorWithLocation(node, "indexed property '%s' requires a setter method, not a field", propInfo.Name)
		}

		// Look up the setter method
		method := obj.Class.LookupMethod(propInfo.WriteSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "indexed property '%s' setter method '%s' not found", propInfo.Name, propInfo.WriteSpec)
		}

		// Verify method has correct number of parameters (index params + value param)
		expectedParamCount := len(indices) + 1 // indices + value
		if len(method.Parameters) != expectedParamCount {
			return i.newErrorWithLocation(node, "indexed property '%s' setter method '%s' expects %d parameter(s) (indices + value), got %d",
				propInfo.Name, propInfo.WriteSpec, expectedParamCount, len(method.Parameters))
		}

		// Create method environment
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the object
		i.env.Define("Self", obj)

		// Bind index parameters (all but the last parameter)
		for idx := 0; idx < len(indices); idx++ {
			if idx < len(method.Parameters) {
				i.env.Define(method.Parameters[idx].Name.Value, indices[idx])
			}
		}

		// Bind value parameter (last parameter)
		if len(method.Parameters) > 0 {
			lastParamIdx := len(method.Parameters) - 1
			i.env.Define(method.Parameters[lastParamIdx].Name.Value, value)
		}

		// Execute method body
		i.Eval(method.Body)

		// Restore environment
		i.env = savedEnv

		// DWScript assignment is an expression that returns the assigned value
		return value

	case types.PropAccessNone:
		// Read-only property
		return i.newErrorWithLocation(node, "indexed property '%s' is read-only", propInfo.Name)

	default:
		return i.newErrorWithLocation(node, "indexed property '%s' has no write access", propInfo.Name)
	}
}

// evalPropertyWrite evaluates a property write access.
// Handles field-backed and method-backed property setters.
func (i *Interpreter) evalPropertyWrite(obj *ObjectInstance, propInfo *types.PropertyInfo, value Value, node ast.Node) Value {
	// Task 9.32c: Initialize property evaluation context if needed
	if i.propContext == nil {
		i.propContext = &PropertyEvalContext{
			PropertyChain: make([]string, 0),
		}
	}

	// Task 9.32c: Check for circular property references
	for _, prop := range i.propContext.PropertyChain {
		if prop == propInfo.Name {
			return i.newErrorWithLocation(node, "circular property reference detected: %s", propInfo.Name)
		}
	}

	// Task 9.32c: Push property onto chain
	i.propContext.PropertyChain = append(i.propContext.PropertyChain, propInfo.Name)
	defer func() {
		// Pop property from chain when done
		if len(i.propContext.PropertyChain) > 0 {
			i.propContext.PropertyChain = i.propContext.PropertyChain[:len(i.propContext.PropertyChain)-1]
		}
		// Clear context if chain is empty
		if len(i.propContext.PropertyChain) == 0 {
			i.propContext = nil
		}
	}()

	switch propInfo.WriteKind {
	case types.PropAccessField:
		// Field or method access - check at runtime which it is
		// First try as a field
		fields := obj.Class.GetFieldsMap()
		if fields != nil && fields[propInfo.WriteSpec] != nil {
			if propInfo.HasIndexValue {
				return i.newErrorWithLocation(node, "property '%s' uses an index directive and cannot be field-backed", propInfo.Name)
			}
			obj.SetField(propInfo.WriteSpec, value)
			return value
		}

		// Not a field - try as a method (setter)
		method := obj.Class.LookupMethod(propInfo.WriteSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "property '%s' write specifier '%s' not found as field or method", propInfo.Name, propInfo.WriteSpec)
		}

		indexArgs, err := i.buildIndexDirectiveArgs(propInfo)
		if err != nil {
			return i.newErrorWithLocation(node, "%s", err.Error())
		}
		expectedParams := len(indexArgs) + 1 // value parameter
		if len(method.Parameters) != expectedParams {
			return i.newErrorWithLocation(node, "property '%s' setter method '%s' expects %d parameter(s), but index directive supplies %d and value provides 1",
				propInfo.Name, propInfo.WriteSpec, len(method.Parameters), len(indexArgs))
		}

		// Call the setter method with the value as argument
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the object
		i.env.Define("Self", obj)

		// Bind implicit index arguments first
		for idx := 0; idx < len(indexArgs); idx++ {
			i.env.Define(method.Parameters[idx].Name.Value, indexArgs[idx])
		}

		// Bind the value parameter (setter should have exactly one parameter after index args)
		if len(method.Parameters) > 0 {
			paramName := method.Parameters[len(method.Parameters)-1].Name.Value
			i.env.Define(paramName, value)
		}

		// Task 9.32c: Set flag to indicate we're inside a property setter
		savedInSetter := i.propContext.InPropertySetter
		i.propContext.InPropertySetter = true
		defer func() {
			i.propContext.InPropertySetter = savedInSetter
		}()

		// Execute method body
		i.Eval(method.Body)

		// Restore environment
		i.env = savedEnv

		return value

	case types.PropAccessMethod:
		// Check if method exists
		method := obj.Class.LookupMethod(propInfo.WriteSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "property '%s' setter method '%s' not found", propInfo.Name, propInfo.WriteSpec)
		}

		indexArgs, err := i.buildIndexDirectiveArgs(propInfo)
		if err != nil {
			return i.newErrorWithLocation(node, "%s", err.Error())
		}
		expectedParams := len(indexArgs) + 1 // value parameter
		if len(method.Parameters) != expectedParams {
			return i.newErrorWithLocation(node, "property '%s' setter method '%s' expects %d parameter(s), but index directive supplies %d and value provides 1",
				propInfo.Name, propInfo.WriteSpec, len(method.Parameters), len(indexArgs))
		}

		// Call the setter method with the value as argument
		// Create method environment with Self bound to object
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the object
		i.env.Define("Self", obj)

		// Bind implicit index arguments first
		for idx := 0; idx < len(indexArgs); idx++ {
			i.env.Define(method.Parameters[idx].Name.Value, indexArgs[idx])
		}

		// Bind the value parameter (setter should have exactly one parameter after index args)
		if len(method.Parameters) > 0 {
			i.env.Define(method.Parameters[len(method.Parameters)-1].Name.Value, value)
		}

		// Task 9.32c: Set flag to indicate we're inside a property setter
		savedInSetter := i.propContext.InPropertySetter
		i.propContext.InPropertySetter = true
		defer func() {
			i.propContext.InPropertySetter = savedInSetter
		}()

		// Execute method body
		i.Eval(method.Body)

		// Restore environment
		i.env = savedEnv

		return value

	case types.PropAccessNone:
		// Read-only property
		return i.newErrorWithLocation(node, "property '%s' is read-only", propInfo.Name)

	default:
		return i.newErrorWithLocation(node, "property '%s' has no write access", propInfo.Name)
	}
}
