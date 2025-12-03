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

	// Migrated from adapter.CheckType() to direct ClassMetadata usage
	result := e.checkType(left, targetTypeName)
	return &runtime.BooleanValue{Value: result}
}

// VisitAsExpression evaluates an 'as' type casting expression.
// Runtime type casting with interface wrapping/unwrapping.
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

	// Use evaluator's castType helper which handles:
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
// Interface implementation verification.
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

	// Use evaluator's checkImplements helper (no adapter)
	result, err := e.checkImplements(left, targetInterfaceName)
	if err != nil {
		return e.newError(node, "%s", err.Error())
	}

	return &runtime.BooleanValue{Value: result}
}

// checkType checks if a value is an instance of a specific type.
// Migrated from adapter.CheckType() to use ClassMetadata directly.
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
func (e *Evaluator) getClassMetadataFromValue(obj Value) *runtime.ClassMetadata {
	// Inline metadata extraction (was e.adapter.GetClassMetadataFromValue)
	// Try ObjectValue interface first (most common case)
	if objVal, ok := obj.(ObjectValue); ok {
		// Use ClassName to get the class name, then lookup via TypeSystem
		className := objVal.ClassName()
		if classInfo := e.typeSystem.LookupClass(className); classInfo != nil {
			// ClassInfo in TypeSystem should have GetMetadata() method
			if metadataProvider, ok := classInfo.(interface{ GetMetadata() *runtime.ClassMetadata }); ok {
				return metadataProvider.GetMetadata()
			}
		}
	}
	return nil
}

// classImplementsInterface checks if a class implements an interface.
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

// checkImplements checks if a value implements an interface.
//
// This implements the 'implements' operator which checks EXPLICIT interface implementations:
// - Does NOT check interface inheritance (differs from 'is' operator)
// - Returns (bool, error) where error is only for unknown interfaces
func (e *Evaluator) checkImplements(obj Value, interfaceName string) (bool, error) {
	// 1. Handle nil - nil implements no interfaces
	if obj == nil || obj.Type() == "NIL" {
		return false, nil
	}

	// 2. Extract ClassMetadata using adapter (extraction only)
	classMeta := e.getClassMetadataFromValue(obj)
	if classMeta == nil {
		// Guard against nil metadata (e.g., uninitialized metaclass variables)
		// Return false, not error - this matches adapter behavior
		return false, nil
	}

	// 3. Check if interface exists in TypeSystem
	if !e.typeSystem.HasInterface(interfaceName) {
		return false, fmt.Errorf("interface '%s' not found", interfaceName)
	}

	// 4. Check implementation using metadata traversal
	return e.classImplementsInterfaceExplicitly(classMeta, interfaceName), nil
}

// classImplementsInterfaceExplicitly checks if a class explicitly implements an interface.
//
// This is separate from classImplementsInterface() because:
// - classImplementsInterface() (for 'is' operator) - WILL check interface inheritance when implemented
// - classImplementsInterfaceExplicitly() (for 'implements' operator) - will NOT check interface inheritance
//
// Currently both are identical because interface inheritance isn't implemented yet.
// They will diverge when interface inheritance is added.
func (e *Evaluator) classImplementsInterfaceExplicitly(classMeta *runtime.ClassMetadata, interfaceName string) bool {
	if classMeta == nil {
		return false
	}

	// Check if this class explicitly declares the interface
	for _, ifaceName := range classMeta.Interfaces {
		if ident.Equal(ifaceName, interfaceName) {
			return true
		}
		// Note: Does NOT check interface inheritance (explicit declarations only)
		// This differs from classImplementsInterface() used by 'is' operator
	}

	// Recursively check parent class
	if classMeta.Parent != nil {
		return e.classImplementsInterfaceExplicitly(classMeta.Parent, interfaceName)
	}

	return false
}

// castType performs type casting for the 'as' operator.
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
	// Task 3.5.31: Inline GetInterfaceInstanceFromValue - use InterfaceInstanceValue type assertion
	if ifaceVal, ok := obj.(InterfaceInstanceValue); ok {
		underlyingObjVal := ifaceVal.GetUnderlyingObjectValue()

		// Check if target is a class via TypeSystem
		if e.typeSystem.HasClass(typeName) {
			// Interface-to-class casting: extract the underlying object
			if underlyingObjVal == nil {
				return nil, fmt.Errorf("cannot cast nil interface to class '%s'", typeName)
			}

			// Get the underlying object's class metadata
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
			if underlyingObjVal == nil {
				// DWScript: nil interface cast to interface yields nil interface wrapper
				// Task 3.5.31: Use inlined createInterfaceWrapper instead of adapter
				nilWrapper, err := e.createInterfaceWrapper(typeName, nil)
				if err != nil {
					return nil, err
				}
				return nilWrapper, nil
			}

			// Check if the underlying object's class implements the target interface
			underlyingClassMeta := e.getClassMetadataFromValue(underlyingObjVal)
			if underlyingClassMeta == nil {
				return nil, fmt.Errorf("cannot extract class metadata from interface underlying object")
			}

			if !e.classImplementsInterface(underlyingClassMeta, typeName) {
				return nil, fmt.Errorf("cannot cast interface of '%s' to interface '%s'", underlyingClassMeta.Name, typeName)
			}

			// Create and return new interface instance
			// Task 3.5.31: Use inlined createInterfaceWrapper instead of adapter
			wrapper, err := e.createInterfaceWrapper(typeName, underlyingObjVal)
			if err != nil {
				return nil, err
			}
			return wrapper, nil
		}

		return nil, fmt.Errorf("type '%s' not found (neither class nor interface)", typeName)
	}

	// Handle object casting - inline type assertion (was e.adapter.GetObjectInstanceFromValue)
	if _, ok := obj.(ObjectValue); !ok {
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
		// Task 3.5.31: Use inlined createInterfaceWrapper instead of adapter
		wrapper, err := e.createInterfaceWrapper(typeName, obj)
		if err != nil {
			return nil, err
		}
		return wrapper, nil
	}

	return nil, fmt.Errorf("type '%s' not found (neither class nor interface)", typeName)
}

// isClassHierarchyCompatible checks if a class is compatible with a target class.
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

// createInterfaceWrapper creates an InterfaceInstance wrapper for the given interface name and object.
// Task 3.5.31: Inlined from adapter.CreateInterfaceWrapper.
// Returns the InterfaceInstance wrapper or error if interface not found.
func (e *Evaluator) createInterfaceWrapper(interfaceName string, obj Value) (Value, error) {
	// Look up the interface via TypeSystem
	ifaceInfoAny := e.typeSystem.LookupInterface(interfaceName)
	if ifaceInfoAny == nil {
		return nil, fmt.Errorf("interface '%s' not found", interfaceName)
	}

	// Cast to runtime.IInterfaceInfo
	ifaceInfo, ok := ifaceInfoAny.(runtime.IInterfaceInfo)
	if !ok {
		return nil, fmt.Errorf("interface '%s' does not implement IInterfaceInfo", interfaceName)
	}

	// Handle nil object case
	if obj == nil {
		return runtime.NewInterfaceInstance(ifaceInfo, nil), nil
	}

	// Extract ObjectInstance from the object
	objInstance, ok := obj.(*runtime.ObjectInstance)
	if !ok {
		return nil, fmt.Errorf("cannot create interface wrapper for non-object type: %s", obj.Type())
	}

	return runtime.NewInterfaceInstance(ifaceInfo, objInstance), nil
}
