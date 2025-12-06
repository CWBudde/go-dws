package interp

import (
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// evalRecordMethodCall evaluates a method call on a record value,
// including method resolution, argument evaluation, and helper method dispatch.
func (i *Interpreter) evalRecordMethodCall(recVal *RecordValue, memberAccess *ast.MemberAccessExpression, argExprs []ast.Expression, objExpr ast.Expression) Value {
	methodName := memberAccess.Member.Value

	// Method resolution - lookup in RecordValue.Methods (instance methods)
	// No inheritance needed for records (unlike classes)
	if !RecordHasMethod(recVal, methodName) {
		// Check for class methods (static methods can be called on instances)
		recordTypeKey := "__record_type_" + ident.Normalize(recVal.RecordType.Name)
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				// Check if this is a class method (case-insensitive)
				if classMethod, exists := rtv.ClassMethods[ident.Normalize(methodName)]; exists {
					// Call the class method (static method)
					return i.callRecordStaticMethod(rtv, classMethod, argExprs, memberAccess)
				}
			}
		}

		// Check if helpers provide this method
		helper, helperMethod, builtinSpec := i.findHelperMethod(recVal, methodName)
		if helperMethod == nil && builtinSpec == "" {
			if recVal.RecordType != nil {
				return i.newErrorWithLocation(memberAccess, "method '%s' not found in record type '%s' (no helper found)",
					methodName, recVal.RecordType.Name)
			}
			return i.newErrorWithLocation(memberAccess, "method '%s' not found (no helper found)", methodName)
		}

		// Evaluate method arguments
		args := make([]Value, len(argExprs))
		for idx, arg := range argExprs {
			val := i.Eval(arg)
			if isError(val) {
				return val
			}
			args[idx] = val
		}

		// Call the helper method
		return i.callHelperMethod(helper, helperMethod, builtinSpec, recVal, args, memberAccess)
	}

	method := GetRecordMethod(recVal, methodName)
	if method == nil {
		return i.newErrorWithLocation(memberAccess, "method '%s' not found in record type '%s'",
			methodName, recVal.RecordType.Name)
	}

	// Evaluate method arguments
	args := make([]Value, len(argExprs))
	for idx, arg := range argExprs {
		val := i.Eval(arg)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	// Check argument count matches parameter count
	if len(args) != len(method.Parameters) {
		return i.newErrorWithLocation(memberAccess, "wrong number of arguments for method '%s': expected %d, got %d",
			methodName, len(method.Parameters), len(args))
	}

	// Create method environment with Self bound to the record
	// IMPORTANT: Records are value types, so we need to work with a copy
	// For mutating methods, we'll need to copy back changes to the original
	// Phase 3.8.2.9: Use helper to sync both i.env and i.ctx.env
	savedEnv := i.env
	i.PushEnvironment(i.env)

	// Make a copy of the record for the method execution
	// This implements value semantics - the method works on a copy
	recordCopy := recVal.Copy()

	// Bind Self to the record copy
	i.env.Define("Self", recordCopy)

	// Bind all record fields to environment so they can be accessed directly
	// This allows code like "X := X + dx" to work without needing "Self.X"
	// Similar to how class property expressions bind fields (see objects.go:431-435)
	for fieldName, fieldValue := range recordCopy.Fields {
		i.env.Define(fieldName, fieldValue)
	}

	// Bind properties to environment (for simple field-backed properties)
	// Properties that use fields as read accessors should be accessible like fields
	for propName, propInfo := range recVal.RecordType.Properties {
		if propInfo.ReadField != "" {
			// Check if the read field is an actual field (use lowercase)
			if fval, exists := recordCopy.Fields[ident.Normalize(propInfo.ReadField)]; exists {
				// Bind the property name to the field value
				i.env.Define(propName, fval)
			}
		}
	}

	// Look up the record type value to get constants and class vars
	recordTypeKey := "__record_type_" + ident.Normalize(recVal.RecordType.Name)
	var recordTypeValue *RecordTypeValue // Track for class var write-back
	var boundClassVars map[string]bool   // Track which class vars we bound
	if typeVal, ok := i.env.Get(recordTypeKey); ok {
		if rtv, ok := typeVal.(*RecordTypeValue); ok {
			recordTypeValue = rtv
			boundClassVars = make(map[string]bool)

			// Bind constants (keys normalized to lowercase for case-insensitive access)
			for constName, constValue := range rtv.Constants {
				i.env.Define(constName, constValue)
			}
			// Bind class variables (keys normalized to lowercase for case-insensitive access)
			for varName, varValue := range rtv.ClassVars {
				i.env.Define(varName, varValue)
				boundClassVars[varName] = true
			}
		}
	}

	// Check recursion depth before pushing to call stack
	if i.ctx.GetCallStack().WillOverflow() {
		i.RestoreEnvironment(savedEnv)
		return i.raiseMaxRecursionExceeded()
	}

	// Push method name onto call stack for stack traces
	fullMethodName := recVal.RecordType.Name + "." + methodName
	i.pushCallStack(fullMethodName)
	defer i.popCallStack()

	// Bind method parameters to arguments with implicit conversion
	for idx, param := range method.Parameters {
		arg := args[idx]

		// Apply implicit conversion if parameter has a type and types don't match
		if param.Type != nil {
			paramTypeName := param.Type.String()
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}

		// TODO: implement proper by-ref support for parameters
		i.env.Define(param.Name.Value, arg)
	}

	// For functions (not procedures), initialize the Result variable
	if method.ReturnType != nil {
		// Initialize Result based on return type with appropriate defaults
		returnType := i.resolveTypeFromAnnotation(method.ReturnType)
		var resultValue = i.getDefaultValue(returnType)

		// Check if return type is a record (overrides default)
		returnTypeName := method.ReturnType.String()
		recordTypeKey := "__record_type_" + ident.Normalize(returnTypeName)
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				// Use createRecordValue for proper nested record initialization
				resultValue = i.createRecordValue(rtv.RecordType)
			}
		}

		i.env.Define("Result", resultValue)
		// Also define the method name as an alias for Result
		// In DWScript, assigning to either Result or the method name sets the return value
		i.env.Define(method.Name.Value, &ReferenceValue{Env: i.env, VarName: "Result"})
	}

	// Execute method body
	if method.Body == nil {
		i.RestoreEnvironment(savedEnv)
		return i.newErrorWithLocation(memberAccess, "method '%s' has no body", methodName)
	}

	bodyResult := i.Eval(method.Body)

	// If an error occurred during execution, propagate it
	if isError(bodyResult) {
		i.RestoreEnvironment(savedEnv)
		return bodyResult
	}

	// If an exception was raised during method execution, propagate it immediately
	if i.exception != nil {
		i.RestoreEnvironment(savedEnv)
		return &NilValue{}
	}

	// Handle exit signal
	if i.ctx.ControlFlow().IsExit() {
		i.ctx.ControlFlow().Clear()
	}

	// Extract return value
	var returnValue Value
	if method.ReturnType != nil {
		// Method has a return type - get the Result value
		resultVal, resultOk := i.env.Get("Result")
		methodNameVal, methodNameOk := i.env.Get(method.Name.Value)

		// Use whichever variable is not nil, preferring Result if both are set
		if resultOk && resultVal.Type() != "NIL" {
			returnValue = resultVal
		} else if methodNameOk && methodNameVal.Type() != "NIL" {
			returnValue = methodNameVal
		} else if resultOk {
			returnValue = resultVal
		} else if methodNameOk {
			returnValue = methodNameVal
		} else {
			returnValue = &NilValue{}
		}

		// Apply implicit conversion if return type doesn't match
		if returnValue.Type() != "NIL" {
			expectedReturnType := method.ReturnType.String()
			if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
				returnValue = converted
			}
		}
	} else {
		// Procedure - no return value
		// But we need to handle copy-back for mutating procedures
		returnValue = &NilValue{}

		// Copy-back semantics for procedures:
		// If the method is a procedure (no return type), it may have modified Self.
		// We need to update the original record with the modified fields.
		// However, since we evaluated the object expression already, we can't directly
		// modify the original. This is a limitation of the current approach.
		//
		// TODO: For full copy-back semantics, we would need to:
		// 1. Track the lvalue (variable) that holds the record
		// 2. Update that variable with the modified record copy
		//
		// For now, we return the modified copy and rely on assignment handling.
	}

	// Copy modified field values back from environment to record copy
	// This ensures that any field modifications made during method execution are preserved
	for fieldName := range recordCopy.Fields {
		if updatedVal, exists := i.env.Get(fieldName); exists {
			recordCopy.Fields[fieldName] = updatedVal
		}
	}

	// Also check for property name assignments and copy them back to backing fields
	// This handles cases where code writes to a property name instead of the field name
	for propName, propInfo := range recVal.RecordType.Properties {
		// Only process properties with a write accessor (field name)
		if propInfo.WriteField != "" {
			// Check if the property name was assigned in the environment
			if updatedVal, exists := i.env.Get(propName); exists {
				// Copy the value to the backing field (use lowercase for field lookup)
				backingFieldName := ident.Normalize(propInfo.WriteField)
				if _, fieldExists := recordCopy.Fields[backingFieldName]; fieldExists {
					recordCopy.Fields[backingFieldName] = updatedVal
				}
			}
		}
	}

	// Write back class variable changes to the record type (shared mutable state)
	if recordTypeValue != nil && boundClassVars != nil {
		for varName := range boundClassVars {
			if updatedVal, exists := i.env.Get(varName); exists {
				recordTypeValue.ClassVars[varName] = updatedVal
			}
		}
	}

	// Restore environment
	i.RestoreEnvironment(savedEnv)

	// Update the original variable with the modified record copy
	// This implements proper value semantics for records - mutations persist
	// Check if the object expression is a simple identifier (variable)
	if ident, ok := objExpr.(*ast.Identifier); ok {
		// Update the variable in the environment with the modified copy
		// This makes mutations visible: p.SetCoords(10, 20) updates p
		i.env.Set(ident.Value, recordCopy)
	}

	return returnValue
}

// callRecordStaticMethod executes a static record method (class function/procedure).
// Example: TPoint.Origin() where Origin is declared as "class function Origin: TPoint"
//
// Parameters:
//   - rtv: The RecordTypeValue containing the static method
//   - method: The FunctionDecl AST node for the static method
//   - argExprs: The argument expressions from the call site
//   - callNode: The call node for error reporting
//
// Static methods behave like regular functions but are scoped to the record type.
// They cannot access instance fields (no Self) but can return values of the record type.
func (i *Interpreter) callRecordStaticMethod(rtv *RecordTypeValue, method *ast.FunctionDecl, argExprs []ast.Expression, callNode ast.Node) Value {
	methodName := method.Name.Value

	// Evaluate method arguments
	args := make([]Value, len(argExprs))
	for idx, arg := range argExprs {
		val := i.Eval(arg)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	// Check argument count matches parameter count
	if len(args) != len(method.Parameters) {
		return i.newErrorWithLocation(callNode, "wrong number of arguments for static method '%s': expected %d, got %d",
			methodName, len(method.Parameters), len(args))
	}

	// Create method environment (NO Self binding for static methods)
	// Phase 3.8.2.9: Use helper to sync both i.env and i.ctx.env
	savedEnv := i.env
	i.PushEnvironment(i.env)

	// Check recursion depth before pushing to call stack
	if i.ctx.GetCallStack().WillOverflow() {
		i.RestoreEnvironment(savedEnv)
		return i.raiseMaxRecursionExceeded()
	}

	// Push method name onto call stack for stack traces
	fullMethodName := rtv.RecordType.Name + "." + methodName
	i.pushCallStack(fullMethodName)
	defer i.popCallStack()

	// Bind __CurrentRecord__ so record static methods can be called without qualification
	i.env.Define("__CurrentRecord__", rtv)

	// Track which class vars we bound for write-back
	boundClassVars := make(map[string]bool)

	// Bind constants (keys normalized to lowercase for case-insensitive access)
	for constName, constValue := range rtv.Constants {
		i.env.Define(constName, constValue)
	}
	// Bind class variables (keys normalized to lowercase for case-insensitive access)
	for varName, varValue := range rtv.ClassVars {
		i.env.Define(varName, varValue)
		boundClassVars[varName] = true
	}

	// Bind method parameters to arguments with implicit conversion
	for idx, param := range method.Parameters {
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
	if method.ReturnType != nil {
		// Initialize Result based on return type with appropriate defaults
		returnType := i.resolveTypeFromAnnotation(method.ReturnType)
		var resultValue = i.getDefaultValue(returnType)

		// Check if return type is a record (overrides default)
		returnTypeName := method.ReturnType.String()
		recordTypeKey := "__record_type_" + ident.Normalize(returnTypeName)
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if recordTV, ok := typeVal.(*RecordTypeValue); ok {
				// Return type is a record - create an instance
				resultValue = i.createRecordValue(recordTV.RecordType)
			}
		}

		i.env.Define("Result", resultValue)
		// Also define the method name as an alias for Result
		// In DWScript, assigning to either Result or the method name sets the return value
		i.env.Define(methodName, &ReferenceValue{Env: i.env, VarName: "Result"})
	}

	// Execute method body
	result := i.Eval(method.Body)
	if isError(result) {
		i.RestoreEnvironment(savedEnv)
		return result
	}

	// Extract return value (same logic as class methods)
	var returnValue Value
	if method.ReturnType != nil {
		// Check both Result and method name variable
		resultVal, resultOk := i.env.Get("Result")
		methodNameVal, methodNameOk := i.env.Get(methodName)

		// Use whichever variable is not nil, preferring Result if both are set
		if resultOk && resultVal.Type() != "NIL" {
			returnValue = resultVal
		} else if methodNameOk && methodNameVal.Type() != "NIL" {
			returnValue = methodNameVal
		} else if resultOk {
			returnValue = resultVal
		} else if methodNameOk {
			returnValue = methodNameVal
		} else {
			returnValue = &NilValue{}
		}

		// Apply implicit conversion if return type doesn't match
		if returnValue.Type() != "NIL" {
			expectedReturnType := method.ReturnType.String()
			if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
				returnValue = converted
			}
		}
	} else {
		// Procedure - no return value
		returnValue = &NilValue{}
	}

	// Write back class variable changes to the record type (shared mutable state)
	for varName := range boundClassVars {
		if updatedVal, exists := i.env.Get(varName); exists {
			rtv.ClassVars[varName] = updatedVal
		}
	}

	// Restore environment
	i.RestoreEnvironment(savedEnv)

	return returnValue
}
