package interp

import (
	"github.com/cwbudde/go-dws/internal/ast"
)

func (i *Interpreter) invokeRuntimeOperator(entry *runtimeOperatorEntry, operands []Value, node ast.Node) Value {
	if entry == nil {
		return i.newErrorWithLocation(node, "operator not registered")
	}

	if entry.Class != nil {
		if entry.IsClassMethod {
			return i.invokeClassOperatorMethod(entry.Class, entry.BindingName, operands, node)
		}

		if entry.SelfIndex < 0 || entry.SelfIndex >= len(operands) {
			return i.newErrorWithLocation(node, "invalid operator configuration for '%s'", entry.Operator)
		}

		selfVal := operands[entry.SelfIndex]
		obj, ok := selfVal.(*ObjectInstance)
		if !ok {
			return i.newErrorWithLocation(node, "operator '%s' requires object operand", entry.Operator)
		}
		if !obj.IsInstanceOf(entry.Class) {
			return i.newErrorWithLocation(node, "operator '%s' requires instance of '%s'", entry.Operator, entry.Class.Name)
		}

		args := make([]Value, 0, len(operands)-1)
		for idx, val := range operands {
			if idx == entry.SelfIndex {
				continue
			}
			args = append(args, val)
		}

		return i.invokeInstanceOperatorMethod(obj, entry.BindingName, args, node)
	}

	return i.invokeGlobalOperator(entry, operands, node)
}

func (i *Interpreter) invokeGlobalOperator(entry *runtimeOperatorEntry, operands []Value, node ast.Node) Value {
	fn, exists := i.functions[entry.BindingName]
	if !exists {
		return i.newErrorWithLocation(node, "operator binding '%s' not found", entry.BindingName)
	}
	if len(fn.Parameters) != len(operands) {
		return i.newErrorWithLocation(node, "operator '%s' expects %d operands, got %d", entry.Operator, len(fn.Parameters), len(operands))
	}
	return i.callUserFunction(fn, operands)
}

func (i *Interpreter) invokeInstanceOperatorMethod(obj *ObjectInstance, methodName string, args []Value, node ast.Node) Value {
	method := obj.GetMethod(methodName)
	if method == nil {
		return i.newErrorWithLocation(node, "method '%s' not found in class '%s'", methodName, obj.Class.Name)
	}

	if len(args) != len(method.Parameters) {
		return i.newErrorWithLocation(node, "method '%s' expects %d arguments, got %d", methodName, len(method.Parameters), len(args))
	}

	methodEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = methodEnv

	i.env.Define("Self", obj)

	// Bind parameters to arguments with implicit conversion
	for idx, param := range method.Parameters {
		arg := args[idx]

		// Task 8.19b: Apply implicit conversion if parameter has a type and types don't match
		if param.Type != nil {
			paramTypeName := param.Type.Name
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}

		i.env.Define(param.Name.Value, arg)
	}

	if method.ReturnType != nil {
		i.env.Define("Result", &NilValue{})
		i.env.Define(method.Name.Value, &NilValue{})
	}

	result := i.Eval(method.Body)
	if isError(result) {
		i.env = savedEnv
		return result
	}

	var returnValue Value = &NilValue{}
	if method.ReturnType != nil {
		returnValue = i.extractReturnValue(method, methodEnv)
	}

	i.env = savedEnv
	return returnValue
}

func (i *Interpreter) invokeClassOperatorMethod(classInfo *ClassInfo, methodName string, args []Value, node ast.Node) Value {
	method, exists := classInfo.ClassMethods[methodName]
	if !exists {
		return i.newErrorWithLocation(node, "class method '%s' not found in class '%s'", methodName, classInfo.Name)
	}
	if len(args) != len(method.Parameters) {
		return i.newErrorWithLocation(node, "class method '%s' expects %d arguments, got %d", methodName, len(method.Parameters), len(args))
	}

	methodEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = methodEnv

	i.env.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: classInfo})

	// Bind parameters to arguments with implicit conversion
	for idx, param := range method.Parameters {
		arg := args[idx]

		// Task 8.19b: Apply implicit conversion if parameter has a type and types don't match
		if param.Type != nil {
			paramTypeName := param.Type.Name
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}

		i.env.Define(param.Name.Value, arg)
	}

	if method.ReturnType != nil {
		i.env.Define("Result", &NilValue{})
		i.env.Define(method.Name.Value, &NilValue{})
	}

	result := i.Eval(method.Body)
	if isError(result) {
		i.env = savedEnv
		return result
	}

	var returnValue Value = &NilValue{}
	if method.ReturnType != nil {
		returnValue = i.extractReturnValue(method, methodEnv)
	}

	i.env = savedEnv
	return returnValue
}

func (i *Interpreter) extractReturnValue(method *ast.FunctionDecl, env *Environment) Value {
	resultVal, resultOk := env.Get("Result")
	methodNameVal, methodNameOk := env.Get(method.Name.Value)

	var returnValue Value
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

	// Task 8.19c: Apply implicit conversion if return type doesn't match
	if method.ReturnType != nil && returnValue.Type() != "NIL" {
		expectedReturnType := method.ReturnType.Name
		if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
			return converted
		}
	}

	return returnValue
}

// tryImplicitConversion attempts to apply an implicit conversion from source to target type.
// Returns (convertedValue, true) if conversion found and applied, (original, false) otherwise.
// Task 8.19a: Apply implicit conversions automatically at runtime.
// Task 8.19d: Support chained implicit conversions (e.g., Integer -> String -> Custom).
func (i *Interpreter) tryImplicitConversion(value Value, targetTypeName string) (Value, bool) {
	// Handle nil value
	if value == nil {
		return nil, false
	}

	sourceTypeName := value.Type()

	// No conversion needed if types already match
	if sourceTypeName == targetTypeName {
		return value, false
	}

	// Normalize type names for conversion lookup (to match how they're registered)
	normalizedSource := normalizeTypeAnnotation(sourceTypeName)
	normalizedTarget := normalizeTypeAnnotation(targetTypeName)

	// Task 8.19a: Try direct conversion first
	entry, found := i.conversions.findImplicit(normalizedSource, normalizedTarget)
	if found {
		// Look up the conversion function
		fn, exists := i.functions[entry.BindingName]
		if !exists {
			// This should not happen if semantic analysis passed
			return value, false
		}

		// Call the conversion function with the value as argument
		args := []Value{value}
		result := i.callUserFunction(fn, args)

		if isError(result) {
			return result, false
		}

		return result, true
	}

	// Task 8.19d: Try chained conversion if direct conversion not found
	const maxConversionChainDepth = 3
	path := i.conversions.findConversionPath(normalizedSource, normalizedTarget, maxConversionChainDepth)
	if len(path) < 2 {
		return value, false
	}

	// Apply conversions sequentially along the path
	currentValue := value
	for idx := 0; idx < len(path)-1; idx++ {
		fromType := path[idx]
		toType := path[idx+1]

		// Find the conversion function for this step
		stepEntry, stepFound := i.conversions.findImplicit(fromType, toType)
		if !stepFound {
			// Path is broken - this shouldn't happen if findConversionPath worked correctly
			return value, false
		}

		// Look up the conversion function
		fn, exists := i.functions[stepEntry.BindingName]
		if !exists {
			return value, false
		}

		// Apply this conversion step
		args := []Value{currentValue}
		result := i.callUserFunction(fn, args)

		if isError(result) {
			return result, false
		}

		currentValue = result
	}

	return currentValue, true
}
