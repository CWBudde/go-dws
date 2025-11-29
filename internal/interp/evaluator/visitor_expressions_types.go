package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// This file contains visitor methods for type operation expression AST nodes.
// These include 'is' type checking, 'as' type casting, and 'implements' interface checking.

// VisitIsExpression evaluates an 'is' type checking expression.
// Task 3.5.34: Runtime type checking with class hierarchy and interface support.
//
// The 'is' operator has two modes:
// 1. Type checking: `obj is TMyClass` or `obj is IMyInterface`
//   - Returns true if obj is an instance of the class (or derived class)
//   - Returns true if obj's class implements the interface
//   - Returns false for nil objects
//
// 2. Boolean value comparison: `boolExpr is True` or `boolExpr is False`
//   - This variant uses Right expression instead of TargetType
//   - Compares two values as booleans using variant-to-bool coercion
func (e *Evaluator) VisitIsExpression(node *ast.IsExpression, ctx *ExecutionContext) Value {
	// Check if this is a boolean value comparison (expr.Right is set)
	// or a type check (expr.TargetType is set)
	if node.Right != nil {
		// Boolean value comparison: left is right
		// This is essentially checking if left == right for boolean values
		left := e.Eval(node.Left, ctx)
		if isError(left) {
			return left
		}

		right := e.Eval(node.Right, ctx)
		if isError(right) {
			return right
		}

		// Convert both to boolean values using VariantToBool
		leftBool := VariantToBool(left)
		rightBool := VariantToBool(right)

		return &runtime.BooleanValue{Value: leftBool == rightBool}
	}

	// Type checking mode
	// Evaluate the left expression (the object to check)
	left := e.Eval(node.Left, ctx)
	if isError(left) {
		return left
	}

	// Handle nil - nil is not an instance of any type
	if left == nil || left.Type() == "NIL" {
		return &runtime.BooleanValue{Value: false}
	}

	// Get the target type name from the type expression
	targetTypeName := ""
	if typeAnnotation, ok := node.TargetType.(*ast.TypeAnnotation); ok {
		targetTypeName = typeAnnotation.Name
	} else {
		return e.newError(node, "cannot determine target type")
	}

	// Task 3.5.140: Migrated from adapter.CheckType() to direct ClassMetadata usage
	result := e.checkType(left, targetTypeName)
	return &runtime.BooleanValue{Value: result}
}

// VisitAsExpression evaluates an 'as' type casting expression.
// Task 3.5.35: Runtime type casting with interface wrapping/unwrapping.
//
// The 'as' operator performs type casting with the following behaviors:
// 1. nil -> any type: returns nil (nil can be cast to any class or interface)
// 2. interface -> class: extracts underlying object (validates class hierarchy)
// 3. interface -> interface: creates new interface wrapper (validates implementation)
// 4. object -> class: validates class hierarchy (returns same object if valid)
// 5. object -> interface: creates interface wrapper (validates implementation)
//
// If the cast is invalid, returns an error that should be raised as an exception.
func (e *Evaluator) VisitAsExpression(node *ast.AsExpression, ctx *ExecutionContext) Value {
	// Evaluate the left expression (the object to cast)
	left := e.Eval(node.Left, ctx)
	if isError(left) {
		return left
	}

	// Get the target type name from the type expression
	targetTypeName := ""
	if typeAnnotation, ok := node.TargetType.(*ast.TypeAnnotation); ok {
		targetTypeName = typeAnnotation.Name
	} else {
		return e.newError(node, "cannot determine target type")
	}

	// Use the adapter's CastType method which handles:
	// - nil casting
	// - interface-to-class casting (unwrapping)
	// - interface-to-interface casting (re-wrapping)
	// - object-to-class casting (hierarchy validation)
	// - object-to-interface casting (wrapping)
	result, err := e.adapter.CastType(left, targetTypeName)
	if err != nil {
		// Cast failed - return error that should be raised as exception
		// The error message from CastType includes the specific failure reason
		return e.newError(node, "%s", err.Error())
	}

	return result
}

// VisitImplementsExpression evaluates an 'implements' interface checking expression.
// Task 3.5.36: Interface implementation verification.
//
// The 'implements' operator checks if a class/object implements an interface:
// - obj implements IInterface -> Boolean (object instance)
// - TClass implements IInterface -> Boolean (class type reference)
// - ClassVar implements IInterface -> Boolean (metaclass variable)
//
// Returns true if the class implements the interface, false otherwise.
// Unlike 'is' which checks class hierarchy too, 'implements' only checks interfaces.
func (e *Evaluator) VisitImplementsExpression(node *ast.ImplementsExpression, ctx *ExecutionContext) Value {
	// Evaluate the left expression (the object or class to check)
	left := e.Eval(node.Left, ctx)
	if isError(left) {
		return left
	}

	// Get the target interface name from the type expression
	targetInterfaceName := ""
	if typeAnnotation, ok := node.TargetType.(*ast.TypeAnnotation); ok {
		targetInterfaceName = typeAnnotation.Name
	} else {
		return e.newError(node, "cannot determine target interface type")
	}

	// Use the adapter's CheckImplements method which handles:
	// - nil objects (return false)
	// - ObjectInstance (extract class)
	// - ClassValue (metaclass variable)
	// - ClassInfoValue (class type identifier)
	result, err := e.adapter.CheckImplements(left, targetInterfaceName)
	if err != nil {
		return e.newError(node, "%s", err.Error())
	}

	return &runtime.BooleanValue{Value: result}
}

// checkType checks if a value is an instance of a specific type.
// Task 3.5.140: Migrated from adapter.CheckType() to use ClassMetadata directly.
//
// This implements the 'is' operator type checking with:
// - Class hierarchy traversal (obj is TMyClass checks class and parent classes)
// - Interface implementation checking (obj is IMyInterface checks if class implements it)
func (e *Evaluator) checkType(obj Value, typeName string) bool {
	// Handle nil - nil is not an instance of any type
	if obj == nil || obj.Type() == "NIL" {
		return false
	}

	// Get the class metadata from the object
	// For ObjectInstance, we access Class.Metadata
	// For InterfaceInstance, we access Object.Class.Metadata
	classMeta := e.getClassMetadataFromValue(obj)
	if classMeta == nil {
		return false
	}

	// Check if the object's class matches (case-insensitive)
	if ident.Equal(classMeta.Name, typeName) {
		return true
	}

	// Check parent class hierarchy
	current := classMeta.Parent
	for current != nil {
		if ident.Equal(current.Name, typeName) {
			return true
		}
		current = current.Parent
	}

	// Check if the target is an interface and if the object's class implements it
	if e.typeSystem.HasInterface(typeName) {
		return e.classImplementsInterface(classMeta, typeName)
	}

	return false
}

// getClassMetadataFromValue extracts ClassMetadata from a value.
// Task 3.5.140: Helper to extract metadata from ObjectInstance or InterfaceInstance.
func (e *Evaluator) getClassMetadataFromValue(obj Value) *runtime.ClassMetadata {
	// Use adapter to extract ClassMetadata from the object
	// The adapter has access to internal types and can safely extract the metadata
	return e.adapter.GetClassMetadataFromValue(obj)
}

// classImplementsInterface checks if a class implements an interface.
// Task 3.5.140: Helper for checkType to verify interface implementation.
func (e *Evaluator) classImplementsInterface(classMeta *runtime.ClassMetadata, interfaceName string) bool {
	if classMeta == nil {
		return false
	}

	// Check if this class explicitly declares the interface
	for _, ifaceName := range classMeta.Interfaces {
		if ident.Equal(ifaceName, interfaceName) {
			return true
		}
		// TODO: Check interface inheritance when that's implemented
	}

	// Check parent class (interfaces are inherited)
	if classMeta.Parent != nil {
		return e.classImplementsInterface(classMeta.Parent, interfaceName)
	}

	return false
}
