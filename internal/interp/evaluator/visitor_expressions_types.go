package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	pkgident "github.com/cwbudde/go-dws/pkg/ident"
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

	// Task 3.5.141: Use evaluator's castType helper which handles:
	// - nil casting
	// - interface-to-class casting (unwrapping)
	// - interface-to-interface casting (re-wrapping)
	// - object-to-class casting (hierarchy validation)
	// - object-to-interface casting (wrapping)
	result, err := e.castType(left, targetTypeName, node)
	if err != nil {
		// Cast failed - return error that should be raised as exception
		// The error message from castType includes the specific failure reason
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

// castType performs type casting for the 'as' operator.
// Task 3.5.141: Migrated from adapter.CastType().
//
// Handles:
// 1. Variant → primitive types (using existing cast helpers)
// 2. nil → any type (returns nil)
// 3. interface → class (extract + validate)
// 4. interface → interface (rewrap + validate)
// 5. object → class (validate hierarchy)
// 6. object → interface (wrap + validate)
//
// Returns (Value, error) - does NOT raise exceptions.
func (e *Evaluator) castType(obj Value, typeName string, node ast.Node) (Value, error) {
	targetLower := pkgident.Normalize(typeName)

	// Handle variant-specific casting to primitive types
	if variantVal, ok := obj.(VariantAccessor); ok {
		switch targetLower {
		case "integer":
			result := e.castToInteger(variantVal.GetVariantValue())
			if isError(result) {
				return nil, fmt.Errorf("%s", result.String())
			}
			return result, nil
		case "float":
			result := e.castToFloat(variantVal.GetVariantValue())
			if isError(result) {
				return nil, fmt.Errorf("%s", result.String())
			}
			return result, nil
		case "string":
			result := e.castToString(variantVal.GetVariantValue())
			return result, nil
		case "boolean":
			result := e.castToBoolean(variantVal.GetVariantValue())
			if isError(result) {
				return nil, fmt.Errorf("%s", result.String())
			}
			return result, nil
		case "variant":
			return obj, nil
		}

		// For class/interface targets, unwrap and continue
		obj = variantVal.GetVariantValue()
		if obj == nil {
			obj = &runtime.NilValue{}
		}
	}

	// Handle nil - nil can be cast to any type
	if _, isNil := obj.(*runtime.NilValue); isNil {
		return &runtime.NilValue{}, nil
	}

	// Handle interface-to-object/interface casting
	interfaceInfo, underlyingObj := e.adapter.GetInterfaceInstanceFromValue(obj)
	if interfaceInfo != nil {
		// Check if target is a class via TypeSystem
		if e.typeSystem.HasClass(typeName) {
			// Interface-to-class casting: extract the underlying object
			if underlyingObj == nil {
				return nil, fmt.Errorf("cannot cast nil interface to class '%s'", typeName)
			}

			// Get the underlying object's class metadata
			underlyingObjVal := underlyingObj.(Value)
			underlyingClassMeta := e.getClassMetadataFromValue(underlyingObjVal)
			if underlyingClassMeta == nil {
				return nil, fmt.Errorf("cannot extract class metadata from interface underlying object")
			}

			// Check if the underlying object's class is compatible with the target class
			targetClassMeta := e.typeSystem.LookupClass(typeName)
			if !e.isClassHierarchyCompatible(underlyingClassMeta, targetClassMeta) {
				return nil, fmt.Errorf("cannot cast interface of '%s' to class '%s'", underlyingClassMeta.Name, typeName)
			}

			// Cast is valid - return the underlying object
			return underlyingObjVal, nil
		}

		// Check if target is an interface
		if e.typeSystem.HasInterface(typeName) {
			// Interface-to-interface casting
			if underlyingObj == nil {
				// DWScript: nil interface cast to interface yields nil interface wrapper
				nilWrapper, err := e.adapter.CreateInterfaceWrapper(typeName, nil)
				if err != nil {
					return nil, err
				}
				return nilWrapper, nil
			}

			// Check if the underlying object's class implements the target interface
			underlyingObjVal := underlyingObj.(Value)
			underlyingClassMeta := e.getClassMetadataFromValue(underlyingObjVal)
			if underlyingClassMeta == nil {
				return nil, fmt.Errorf("cannot extract class metadata from interface underlying object")
			}

			if !e.classImplementsInterface(underlyingClassMeta, typeName) {
				return nil, fmt.Errorf("cannot cast interface of '%s' to interface '%s'", underlyingClassMeta.Name, typeName)
			}

			// Create and return new interface instance
			wrapper, err := e.adapter.CreateInterfaceWrapper(typeName, underlyingObjVal)
			if err != nil {
				return nil, err
			}
			return wrapper, nil
		}

		return nil, fmt.Errorf("type '%s' not found (neither class nor interface)", typeName)
	}

	// Handle object casting
	objVal := e.adapter.GetObjectInstanceFromValue(obj)
	if objVal == nil {
		return nil, fmt.Errorf("'as' operator requires object instance, got %s", obj.Type())
	}

	// Try class-to-class casting first via TypeSystem
	if e.typeSystem.HasClass(typeName) {
		// Get the object's class metadata
		objClassMeta := e.getClassMetadataFromValue(obj)
		if objClassMeta == nil {
			return nil, fmt.Errorf("cannot extract class metadata from object")
		}

		// Validate that the object's actual runtime type is compatible with the target
		targetClassMeta := e.typeSystem.LookupClass(typeName)
		if !e.isClassHierarchyCompatible(objClassMeta, targetClassMeta) {
			return nil, fmt.Errorf("instance of type '%s' cannot be cast to class '%s'", objClassMeta.Name, typeName)
		}

		// Cast is valid - return the same object
		return obj, nil
	}

	// Try interface casting
	if e.typeSystem.HasInterface(typeName) {
		// Get the object's class metadata
		objClassMeta := e.getClassMetadataFromValue(obj)
		if objClassMeta == nil {
			return nil, fmt.Errorf("cannot extract class metadata from object")
		}

		// Validate that the object's class implements the interface
		if !e.classImplementsInterface(objClassMeta, typeName) {
			return nil, fmt.Errorf("class '%s' does not implement interface '%s'", objClassMeta.Name, typeName)
		}

		// Create and return the interface instance
		wrapper, err := e.adapter.CreateInterfaceWrapper(typeName, obj)
		if err != nil {
			return nil, err
		}
		return wrapper, nil
	}

	return nil, fmt.Errorf("type '%s' not found (neither class nor interface)", typeName)
}

// isClassHierarchyCompatible checks if a class is compatible with a target class.
// Task 3.5.141: Helper for class hierarchy validation during type casting.
// Returns true if sourceClass is the same as or a descendant of targetClass.
func (e *Evaluator) isClassHierarchyCompatible(sourceClass, targetClass interface{}) bool {
	// Extract ClassMetadata if we have a generic interface
	var sourceMeta *runtime.ClassMetadata
	var targetMeta *runtime.ClassMetadata

	if sm, ok := sourceClass.(*runtime.ClassMetadata); ok {
		sourceMeta = sm
	} else {
		return false
	}

	if tm, ok := targetClass.(*runtime.ClassMetadata); ok {
		targetMeta = tm
	} else {
		return false
	}

	// Check if source class is the same as or descended from target class
	current := sourceMeta
	for current != nil {
		if pkgident.Equal(current.Name, targetMeta.Name) {
			return true
		}
		current = current.Parent
	}

	return false
}
