package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// Phase 3.5.4 - Phase 2D: Pointer, reference, and operator adapter methods
// These methods implement the InterpreterAdapter interface for references and operators.

// ===== Function Pointers =====

// CreateFunctionPointer creates a function pointer value from a function declaration.
// Task 3.5.8: Adapter method for function pointer creation.
func (i *Interpreter) CreateFunctionPointer(fn *ast.FunctionDecl, closure any) evaluator.Value {
	// Convert closure to Environment
	var env *Environment
	if closure != nil {
		env = closure.(*Environment)
	}

	return &FunctionPointerValue{
		Function: fn,
		Closure:  env,
	}
}

// CreateLambda creates a lambda/closure value from a lambda expression.
// Task 3.5.8: Adapter method for lambda creation.
func (i *Interpreter) CreateLambda(lambda *ast.LambdaExpression, closure any) evaluator.Value {
	// Convert closure to Environment
	var env *Environment
	if closure != nil {
		env = closure.(*Environment)
	}

	return &FunctionPointerValue{
		Lambda:  lambda,
		Closure: env,
	}
}

// IsFunctionPointer checks if a value is a function pointer.
// Task 3.5.8: Adapter method for function pointer type checking.
func (i *Interpreter) IsFunctionPointer(value evaluator.Value) bool {
	_, ok := value.(*FunctionPointerValue)
	return ok
}

// GetFunctionPointerParamCount returns the number of parameters a function pointer expects.
// Task 3.5.8: Adapter method for function pointer parameter count.
func (i *Interpreter) GetFunctionPointerParamCount(funcPtr evaluator.Value) int {
	fp, ok := funcPtr.(*FunctionPointerValue)
	if !ok {
		return 0
	}

	if fp.Function != nil {
		return len(fp.Function.Parameters)
	} else if fp.Lambda != nil {
		return len(fp.Lambda.Parameters)
	}

	return 0
}

// IsFunctionPointerNil checks if a function pointer is nil (unassigned).
// Task 3.5.8: Adapter method for function pointer nil checking.
func (i *Interpreter) IsFunctionPointerNil(funcPtr evaluator.Value) bool {
	fp, ok := funcPtr.(*FunctionPointerValue)
	if !ok {
		return false
	}

	// A function pointer is nil if both Function and Lambda are nil
	return fp.Function == nil && fp.Lambda == nil
}

// CreateMethodPointer creates a method pointer value bound to a specific object.
// Task 3.5.37: Adapter method for method pointer creation from @object.MethodName expressions.
func (i *Interpreter) CreateMethodPointer(objVal evaluator.Value, methodName string, closure any) (evaluator.Value, error) {
	// Extract the object instance
	obj, ok := AsObject(objVal)
	if !ok {
		return nil, fmt.Errorf("method pointer requires an object instance, got %s", objVal.Type())
	}

	// Look up the method in the class hierarchy (case-insensitive)
	method := obj.Class.lookupMethod(methodName)
	if method == nil {
		return nil, fmt.Errorf("undefined method: %s.%s", obj.Class.Name, methodName)
	}

	// Convert closure to Environment
	// Handle both direct *Environment and *EnvironmentAdapter (from evaluator)
	var env *Environment
	if closure != nil {
		if adapter, ok := closure.(*evaluator.EnvironmentAdapter); ok {
			env = adapter.Underlying().(*Environment)
		} else if envVal, ok := closure.(*Environment); ok {
			env = envVal
		}
	}

	// Build parameter types for the function pointer type
	paramTypes := make([]types.Type, len(method.Parameters))
	for idx, param := range method.Parameters {
		if param.Type != nil {
			paramTypes[idx] = i.getTypeFromAnnotation(param.Type)
		} else {
			paramTypes[idx] = &types.IntegerType{} // Default fallback
		}
	}

	// Get return type
	var returnType types.Type
	if method.ReturnType != nil {
		returnType = i.getTypeFromAnnotation(method.ReturnType)
	}

	// Create the method pointer type
	methodPtr := types.NewMethodPointerType(paramTypes, returnType)
	pointerType := &methodPtr.FunctionPointerType

	// Create and return the function pointer value with SelfObject bound
	return NewFunctionPointerValue(method, env, objVal, pointerType), nil
}

// CreateFunctionPointerFromName creates a function pointer for a named function.
// Task 3.5.37: Adapter method for function pointer creation from @FunctionName expressions.
func (i *Interpreter) CreateFunctionPointerFromName(funcName string, closure any) (evaluator.Value, error) {
	// Look up the function in the function registry (case-insensitive)
	overloads, exists := i.functions[ident.Normalize(funcName)]
	if !exists || len(overloads) == 0 {
		return nil, fmt.Errorf("undefined function or procedure: %s", funcName)
	}

	// For overloaded functions, use the first overload
	// Note: Function pointers cannot represent overload sets, only single functions
	function := overloads[0]

	// Convert closure to Environment
	// Handle both direct *Environment and *EnvironmentAdapter (from evaluator)
	var env *Environment
	if closure != nil {
		if adapter, ok := closure.(*evaluator.EnvironmentAdapter); ok {
			env = adapter.Underlying().(*Environment)
		} else if envVal, ok := closure.(*Environment); ok {
			env = envVal
		}
	}

	// Build parameter types for the function pointer type
	paramTypes := make([]types.Type, len(function.Parameters))
	for idx, param := range function.Parameters {
		if param.Type != nil {
			paramTypes[idx] = i.getTypeFromAnnotation(param.Type)
		} else {
			paramTypes[idx] = &types.IntegerType{} // Default fallback
		}
	}

	// Get return type
	var returnType types.Type
	if function.ReturnType != nil {
		returnType = i.getTypeFromAnnotation(function.ReturnType)
	}

	// Create the function pointer type
	var pointerType *types.FunctionPointerType
	if returnType != nil {
		pointerType = types.NewFunctionPointerType(paramTypes, returnType)
	} else {
		pointerType = types.NewProcedurePointerType(paramTypes)
	}

	// Create and return the function pointer value (no SelfObject)
	return NewFunctionPointerValue(function, env, nil, pointerType), nil
}

// ===== Environment and Exceptions =====

// RaiseException raises an exception during execution.
// Task 3.5.8: Adapter method for exception raising.
func (i *Interpreter) RaiseException(className string, message string, pos any) {
	// Convert pos to lexer.Position if provided
	var position *lexer.Position
	if pos != nil {
		if p, ok := pos.(*lexer.Position); ok {
			position = p
		}
	}

	// Call the internal raiseException method
	i.raiseException(className, message, position)
}

// Task 3.5.70: GetVariable removed - evaluator now uses ctx.Env().Get() directly

// DefineVariable defines a new variable in the execution context.
// Task 3.5.9: Adapter method for environment access.
func (i *Interpreter) DefineVariable(name string, value evaluator.Value, ctx *evaluator.ExecutionContext) {
	// Convert to internal Value type
	internalValue := value.(Value)

	// Define in the context's environment
	ctx.Env().Define(name, internalValue)
}

// ===== Binary Operator Adapters =====
// Task 3.5.19: Binary Operator Adapter Methods (Fix for PR #219)
//
// These adapter methods delegate to the Interpreter's binary operator implementation
// WITHOUT re-evaluating the operands. This fixes the double-evaluation bug where
// operands with side effects (function calls, increments, etc.) were executed twice.

// EvalVariantBinaryOp handles binary operations with Variant operands using pre-evaluated values.
func (i *Interpreter) EvalVariantBinaryOp(op string, left, right evaluator.Value, node ast.Node) evaluator.Value {
	// The Interpreter's evalVariantBinaryOp already works with pre-evaluated values
	return i.evalVariantBinaryOp(op, left, right, node)
}

// EvalInOperator evaluates the 'in' operator for membership testing using pre-evaluated values.
func (i *Interpreter) EvalInOperator(value, container evaluator.Value, node ast.Node) evaluator.Value {
	// The Interpreter's evalInOperator already works with pre-evaluated values
	return i.evalInOperator(value, container, node)
}

// EvalEqualityComparison handles = and <> operators for complex types using pre-evaluated values.
func (i *Interpreter) EvalEqualityComparison(op string, left, right evaluator.Value, node ast.Node) evaluator.Value {
	// This is extracted from eval BinaryExpression to handle complex type comparisons
	// with pre-evaluated operands (fixing double-evaluation bug in PR #219)

	// Check if either operand is nil or an object instance
	_, leftIsNil := left.(*NilValue)
	_, rightIsNil := right.(*NilValue)
	_, leftIsObj := left.(*ObjectInstance)
	_, rightIsObj := right.(*ObjectInstance)
	leftIntf, leftIsIntf := left.(*InterfaceInstance)
	rightIntf, rightIsIntf := right.(*InterfaceInstance)
	leftClass, leftIsClass := left.(*ClassValue)
	rightClass, rightIsClass := right.(*ClassValue)

	// Handle RTTITypeInfoValue comparisons (for TypeOf results)
	leftRTTI, leftIsRTTI := left.(*RTTITypeInfoValue)
	rightRTTI, rightIsRTTI := right.(*RTTITypeInfoValue)
	if leftIsRTTI && rightIsRTTI {
		// Compare by TypeID (unique identifier for each type)
		result := leftRTTI.TypeID == rightRTTI.TypeID
		if op == "=" {
			return &BooleanValue{Value: result}
		}
		return &BooleanValue{Value: !result}
	}

	// Handle ClassValue (metaclass) comparisons
	if leftIsClass || rightIsClass {
		// Both are ClassValue - compare by ClassInfo identity
		if leftIsClass && rightIsClass {
			result := leftClass.ClassInfo == rightClass.ClassInfo
			if op == "=" {
				return &BooleanValue{Value: result}
			}
			return &BooleanValue{Value: !result}
		}
		// One is ClassValue, one is nil
		if leftIsNil || rightIsNil {
			if op == "=" {
				return &BooleanValue{Value: false}
			}
			return &BooleanValue{Value: true}
		}
	}

	// Handle InterfaceInstance comparisons
	if leftIsIntf || rightIsIntf {
		// Both are interfaces - compare underlying objects
		if leftIsIntf && rightIsIntf {
			result := leftIntf.Object == rightIntf.Object
			if op == "=" {
				return &BooleanValue{Value: result}
			}
			return &BooleanValue{Value: !result}
		}
		// One is interface, one is nil
		if leftIsNil || rightIsNil {
			var intfIsNil bool
			if leftIsIntf {
				intfIsNil = leftIntf.Object == nil
			} else {
				intfIsNil = rightIntf.Object == nil
			}
			if op == "=" {
				return &BooleanValue{Value: intfIsNil}
			}
			return &BooleanValue{Value: !intfIsNil}
		}
	}

	// If either is nil or an object, do object identity comparison
	if leftIsNil || rightIsNil || leftIsObj || rightIsObj {
		// Both nil
		if leftIsNil && rightIsNil {
			if op == "=" {
				return &BooleanValue{Value: true}
			}
			return &BooleanValue{Value: false}
		}

		// One is nil, one is not
		if leftIsNil || rightIsNil {
			if op == "=" {
				return &BooleanValue{Value: false}
			}
			return &BooleanValue{Value: true}
		}

		// Both are objects - compare by identity
		if op == "=" {
			return &BooleanValue{Value: left == right}
		}
		return &BooleanValue{Value: left != right}
	}

	// Check if both are records
	if _, leftIsRecord := left.(*RecordValue); leftIsRecord {
		if _, rightIsRecord := right.(*RecordValue); rightIsRecord {
			return i.evalRecordBinaryOp(op, left, right)
		}
	}

	// Not a supported equality comparison type
	return i.newErrorWithLocation(node, "type mismatch: %s %s %s", left.Type(), op, right.Type())
}

// ===== Reference Values =====
// Task 3.5.21: Complex Value Retrieval Adapter Method Implementations
//
// These adapter methods allow the Evaluator to handle complex value types
// (ReferenceValue) that require special processing when accessed as identifiers.
//
// Task 3.5.71: IsReferenceValue removed - evaluator uses val.Type() == "REFERENCE" directly
// Task 3.5.73: IsExternalVar, IsLazyThunk, EvaluateLazyThunk, GetExternalVarName removed
//              - evaluator uses ExternalVarAccessor and LazyEvaluator interfaces directly
// Task 3.5.132: DereferenceValue removed - evaluator uses ReferenceAccessor interface directly

// ===== Property and Method References =====
// Task 3.5.22: Property & Method Reference Adapter Method Implementations
//
// These adapter methods allow the Evaluator to access object fields, properties,
// methods, and class metadata when handling identifier lookups in method contexts.

// Task 3.5.71: IsObjectInstance removed - evaluator uses val.Type() == "OBJECT" directly

// GetObjectFieldValue retrieves a field value from an object instance.
func (i *Interpreter) GetObjectFieldValue(obj evaluator.Value, fieldName string) (evaluator.Value, bool) {
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return nil, false
	}
	fieldValue := objInst.GetField(fieldName)
	if fieldValue == nil {
		return nil, false
	}
	return fieldValue, true
}

// GetClassVariableValue retrieves a class variable value from an object's class.
func (i *Interpreter) GetClassVariableValue(obj evaluator.Value, varName string) (evaluator.Value, bool) {
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return nil, false
	}
	// Case-insensitive lookup to match DWScript semantics
	for name, value := range objInst.Class.ClassVars {
		if ident.Equal(name, varName) {
			return value, true
		}
	}
	return nil, false
}

// Task 3.5.72: HasProperty removed - ObjectInstance implements evaluator.ObjectValue directly

// ReadPropertyValue reads a property value from an object.
func (i *Interpreter) ReadPropertyValue(obj evaluator.Value, propName string, node any) (evaluator.Value, error) {
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return nil, fmt.Errorf("cannot read property from non-object value")
	}

	propInfo := objInst.Class.lookupProperty(propName)
	if propInfo == nil {
		return nil, fmt.Errorf("property '%s' not found", propName)
	}

	// Use the existing evalPropertyRead method
	astNode, ok := node.(ast.Node)
	if !ok {
		astNode = nil
	}
	return i.evalPropertyRead(objInst, propInfo, astNode), nil
}

// ExecutePropertyRead executes property reading with a resolved PropertyInfo.
// Task 3.5.116: Low-level method for property getter execution.
// This is the callback implementation for ObjectValue.ReadProperty().
func (i *Interpreter) ExecutePropertyRead(obj evaluator.Value, propInfo any, node any) evaluator.Value {
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return i.NewError("cannot read property from non-object value")
	}

	pInfo, ok := propInfo.(*types.PropertyInfo)
	if !ok {
		return i.NewError("invalid property info type")
	}

	astNode, _ := node.(ast.Node)
	return i.evalPropertyRead(objInst, pInfo, astNode)
}

// Task 3.5.72: HasMethod removed - ObjectInstance implements evaluator.ObjectValue directly

// IsMethodParameterless checks if a method has zero parameters.
func (i *Interpreter) IsMethodParameterless(obj evaluator.Value, methodName string) bool {
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return false
	}
	method, exists := objInst.Class.Methods[strings.ToLower(methodName)]
	if !exists {
		return false
	}
	return len(method.Parameters) == 0
}

// CreateMethodCall creates a synthetic method call expression for auto-invocation.
func (i *Interpreter) CreateMethodCall(obj evaluator.Value, methodName string, node any) evaluator.Value {
	// Create a synthetic method call and evaluate it
	// We create identifiers without token information since this is synthetic
	selfIdent := &ast.Identifier{Value: "Self"}
	methodIdent := &ast.Identifier{Value: methodName}

	// Copy token information from the original node if available
	if astNode, ok := node.(*ast.Identifier); ok {
		selfIdent.Token = astNode.Token
		methodIdent.Token = astNode.Token
	}

	syntheticCall := &ast.MethodCallExpression{
		Object:    selfIdent,
		Method:    methodIdent,
		Arguments: []ast.Expression{},
	}

	return i.evalMethodCall(syntheticCall)
}

// CreateMethodPointerFromObject creates a method pointer for a method with parameters.
func (i *Interpreter) CreateMethodPointerFromObject(obj evaluator.Value, methodName string) (evaluator.Value, error) {
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return nil, fmt.Errorf("cannot create method pointer from non-object value")
	}

	method, exists := objInst.Class.Methods[strings.ToLower(methodName)]
	if !exists {
		return nil, fmt.Errorf("method '%s' not found", methodName)
	}

	// Build the pointer type
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

	return NewFunctionPointerValue(method, i.env, objInst, pointerType), nil
}

// CreateBoundMethodPointer creates a FunctionPointerValue for a method bound to an object.
// Task 3.5.120: Low-level adapter method for method pointer creation.
func (i *Interpreter) CreateBoundMethodPointer(obj evaluator.Value, methodDecl any) evaluator.Value {
	method := methodDecl.(*ast.FunctionDecl)
	objInst := obj.(*ObjectInstance)

	// Build the pointer type
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

	return NewFunctionPointerValue(method, i.env, objInst, pointerType)
}
