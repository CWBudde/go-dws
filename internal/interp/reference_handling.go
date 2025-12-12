package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ===== Function Pointers =====

// CreateFunctionPointer creates a function pointer value from a function declaration.
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

// CreateFunctionPointerFromName creates a function pointer for a named function.
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
	var env *Environment
	if closure != nil {
		if envVal, ok := closure.(*Environment); ok {
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

// CreateExceptionDirect creates an exception with pre-resolved dependencies.
//
// The evaluator resolves the exception class via TypeSystem, then calls this method
// to construct the ExceptionValue without doing class lookup itself.
func (i *Interpreter) CreateExceptionDirect(classMetadata any, message string, pos any, callStack any) any {
	// Convert position
	var position *lexer.Position
	if pos != nil {
		if p, ok := pos.(*lexer.Position); ok {
			position = p
		}
	}

	// Convert call stack
	var stack errors.StackTrace
	if callStack != nil {
		if s, ok := callStack.(errors.StackTrace); ok {
			stack = s
		}
	}

	// Convert ClassMetadata (from runtime) to ClassInfo (interp)
	var excClass *ClassInfo
	if classMetadata != nil {
		if meta, ok := classMetadata.(*runtime.ClassMetadata); ok {
			// Look up ClassInfo using normalized name
			excClass = i.classes[ident.Normalize(meta.Name)]
		}
	}

	// Fallback to base Exception if class not found
	if excClass == nil {
		excClass = i.classes[ident.Normalize("Exception")]
	}

	// Create instance with Message field
	var instance *ObjectInstance
	if excClass != nil {
		instance = NewObjectInstance(excClass)
		instance.SetField("Message", &StringValue{Value: message})
	}

	// Convert ClassMetadata safely (may be nil)
	var metadata *runtime.ClassMetadata
	if classMetadata != nil {
		if meta, ok := classMetadata.(*runtime.ClassMetadata); ok {
			metadata = meta
		}
	}

	// Create ExceptionValue with both Metadata and ClassInfo (backward compatibility)
	return &runtime.ExceptionValue{
		Metadata:  metadata,
		ClassInfo: excClass,
		Instance:  instance,
		Message:   message,
		Position:  position,
		CallStack: stack,
	}
}

// WrapObjectInException wraps an existing ObjectInstance in an ExceptionValue.
//
// The evaluator handles nil checking and validation, this just wraps a valid object.
func (i *Interpreter) WrapObjectInException(objInstance evaluator.Value, pos any, callStack any) any {
	// Convert position
	var position *lexer.Position
	if pos != nil {
		if p, ok := pos.(*lexer.Position); ok {
			position = p
		}
	}

	// Convert call stack
	var stack errors.StackTrace
	if callStack != nil {
		if s, ok := callStack.(errors.StackTrace); ok {
			stack = s
		}
	}

	// Cast to ObjectInstance (caller must ensure this is valid)
	objInst, ok := objInstance.(*ObjectInstance)
	if !ok {
		panic(fmt.Sprintf("runtime error: WrapObjectInException requires ObjectInstance, got %s", objInstance.Type()))
	}

	// Get the class info
	classInfo := objInst.Class

	// Extract message from the object's Message field
	message := ""
	if msgVal, ok := objInst.Fields["Message"]; ok {
		if strVal, ok := msgVal.(*StringValue); ok {
			message = strVal.Value
		}
	}

	// Create ExceptionValue - need concrete ClassInfo
	concreteClass, ok := classInfo.(*ClassInfo)
	if !ok {
		return &runtime.ExceptionValue{
			Message:   message,
			Position:  position,
			CallStack: stack,
		}
	}

	return &runtime.ExceptionValue{
		Metadata:  classInfo.GetMetadata(),
		ClassInfo: concreteClass,
		Message:   message,
		Instance:  objInst,
		Position:  position,
		CallStack: stack,
	}
}

// ===== Binary Operator Adapters =====

// CreateBoundMethodPointer creates a FunctionPointerValue for a method bound to an object.
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

	return NewFunctionPointerValue(method, i.Env(), objInst, pointerType)
}
