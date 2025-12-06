package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// Phase 3.5.4 - Phase 2C: Object operations and type casting adapter methods
// These methods implement the InterpreterAdapter interface for object operations.

// ===== Object Creation =====

// CreateObject creates a new object instance of the specified class with constructor arguments.
func (i *Interpreter) CreateObject(className string, args []evaluator.Value) (evaluator.Value, error) {
	// Convert arguments
	internalArgs := convertEvaluatorArgs(args)

	// Look up class via TypeSystem (case-insensitive)
	classInfoIface := i.typeSystem.LookupClass(className)
	if classInfoIface == nil {
		return nil, fmt.Errorf("class '%s' not found", className)
	}
	classInfo, ok := classInfoIface.(*ClassInfo)
	if !ok {
		return nil, fmt.Errorf("class '%s' has invalid type", className)
	}

	// Check if trying to instantiate an abstract class
	if classInfo.IsAbstractFlag {
		return nil, fmt.Errorf("Trying to create an instance of an abstract class")
	}

	// Check if trying to instantiate an external class
	if classInfo.IsExternalFlag {
		return nil, fmt.Errorf("cannot instantiate external class '%s' - external classes are not supported", className)
	}

	// Create new object instance
	obj := NewObjectInstance(classInfo)

	// Initialize fields with default values
	savedEnv := i.env
	i.PushEnvironment(i.env)

	for fieldName, fieldType := range classInfo.Fields {
		var fieldValue Value
		if fieldDecl, hasDecl := classInfo.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
			fieldValue = i.Eval(fieldDecl.InitValue)
			if isError(fieldValue) {
				i.RestoreEnvironment(savedEnv)
				return nil, fmt.Errorf("failed to initialize field '%s': %v", fieldName, fieldValue)
			}
		} else {
			fieldValue = getZeroValueForType(fieldType, nil)
		}
		obj.SetField(fieldName, fieldValue)
	}

	i.RestoreEnvironment(savedEnv)

	// Call constructor if it exists
	constructorNameLower := ident.Normalize("Create")
	if constructor, exists := classInfo.Constructors[constructorNameLower]; exists {
		ctorEnv := i.PushEnvironment(i.env)
		ctorEnv.Define("Self", obj)

		result := i.executeUserFunctionViaEvaluator(constructor, internalArgs)

		i.RestoreEnvironment(savedEnv)

		// Propagate constructor errors
		if isError(result) {
			return nil, fmt.Errorf("constructor failed: %v", result)
		}
	} else if len(internalArgs) > 0 {
		return nil, fmt.Errorf("no constructor found for class '%s' with %d arguments", className, len(internalArgs))
	}

	return obj, nil
}

// ExecuteConstructor executes a constructor method on an already-created object instance.
func (i *Interpreter) ExecuteConstructor(obj evaluator.Value, constructorName string, args []evaluator.Value) error {
	// Convert arguments
	internalArgs := convertEvaluatorArgs(args)
	internalObj := obj.(Value)

	// Get the object's class
	objectInstance, ok := internalObj.(*ObjectInstance)
	if !ok {
		return fmt.Errorf("value is not an object instance")
	}

	classInfo := objectInstance.Class

	// Look up constructor - need concrete ClassInfo for Constructors map
	concreteClass, ok := classInfo.(*ClassInfo)
	if !ok {
		return fmt.Errorf("class information is not a ClassInfo")
	}

	constructorNameNorm := ident.Normalize(constructorName)
	constructor, exists := concreteClass.Constructors[constructorNameNorm]
	if !exists {
		if len(internalArgs) > 0 {
			return fmt.Errorf("no constructor '%s' found for class '%s' with %d arguments", constructorName, classInfo.GetName(), len(internalArgs))
		}
		// No constructor and no args - OK
		return nil
	}

	// Execute constructor in a new environment with Self bound
	savedEnv := i.env
	ctorEnv := i.PushEnvironment(i.env)
	ctorEnv.Define("Self", objectInstance)

	result := i.executeUserFunctionViaEvaluator(constructor, internalArgs)

	i.RestoreEnvironment(savedEnv)

	// Propagate constructor errors
	if isError(result) {
		return fmt.Errorf("constructor failed: %v", result)
	}

	return nil
}

// ===== Type Checking and Casting =====

// Task 3.5.29: GetClassMetadataFromValue REMOVED - use evaluator.getClassMetadataFromValue() helper
// Replacement: evaluator uses ObjectValue.ClassName() + TypeSystem.LookupClass() + GetMetadata()
//
// Task 3.5.29: CheckType REMOVED - zero callers
// Note: CheckType was already migrated away in earlier phases
//
// CastType performs type casting (implements 'as' operator).
//
// Handles the following cases:
// 1. nil → any type: returns nil
// 2. interface → class: extracts underlying object (with type check)
// 3. interface → interface: creates new interface wrapper (with implementation check)
// 4. object → class: validates class hierarchy
// 5. object → interface: creates interface wrapper (with implementation check)
func (i *Interpreter) CastType(obj evaluator.Value, typeName string) (evaluator.Value, error) {
	// Convert to internal type
	internalObj := obj.(Value)
	targetLower := ident.Normalize(typeName)

	// Variant-specific casting to primitive types
	if variantVal, ok := internalObj.(*VariantValue); ok {
		switch targetLower {
		case "integer":
			result := i.castToInteger(variantVal)
			if isError(result) {
				return nil, fmt.Errorf("%s", result.String())
			}
			return result, nil
		case "float":
			result := i.castToFloat(variantVal)
			if isError(result) {
				return nil, fmt.Errorf("%s", result.String())
			}
			return result, nil
		case "string":
			result := i.castToString(variantVal)
			return result, nil
		case "boolean":
			result := i.castToBoolean(variantVal)
			if isError(result) {
				return nil, fmt.Errorf("%s", result.String())
			}
			return result, nil
		case "variant":
			return variantVal, nil
		}

		// For class/interface targets, unwrap and continue
		internalObj = variantVal.Value
		if internalObj == nil {
			internalObj = &UnassignedValue{}
		}
	}

	// Handle nil - nil can be cast to any type
	if _, isNil := internalObj.(*NilValue); isNil {
		return &NilValue{}, nil
	}

	// Handle interface-to-object/interface casting
	if intfInst, ok := internalObj.(*InterfaceInstance); ok {
		// Check if target is a class via TypeSystem
		if targetClassIface := i.typeSystem.LookupClass(typeName); targetClassIface != nil {
			targetClass, _ := targetClassIface.(*ClassInfo)
			// Interface-to-class casting: extract the underlying object
			underlyingObj := intfInst.Object
			if underlyingObj == nil {
				return nil, fmt.Errorf("cannot cast nil interface to class '%s'", targetClass.Name)
			}

			// Check if the underlying object's class is compatible with the target class
			// Need concrete ClassInfo for isClassCompatible
			if concreteClass, ok := underlyingObj.Class.(*ClassInfo); ok {
				if !isClassCompatible(concreteClass, targetClass) {
					return nil, fmt.Errorf("cannot cast interface of '%s' to class '%s'", underlyingObj.Class.GetName(), targetClass.Name)
				}
			} else {
				return nil, fmt.Errorf("underlying object has invalid class type")
			}

			// Cast is valid - return the underlying object
			return underlyingObj, nil
		}

		// Check if target is an interface
		if targetIface := i.lookupInterfaceInfo(typeName); targetIface != nil {
			// Interface-to-interface casting
			underlyingObj := intfInst.Object
			if underlyingObj == nil {
				// DWScript: nil interface cast to interface yields nil interface wrapper
				return &InterfaceInstance{Interface: targetIface, Object: nil}, nil
			}

			// Check if the underlying object's class implements the target interface
			// Need concrete ClassInfo for classImplementsInterface
			if concreteClass, ok := underlyingObj.Class.(*ClassInfo); ok {
				if !classImplementsInterface(concreteClass, targetIface) {
					return nil, fmt.Errorf("cannot cast interface of '%s' to interface '%s'", underlyingObj.Class.GetName(), targetIface.Name)
				}
			} else {
				return nil, fmt.Errorf("underlying object has invalid class type")
			}

			// Create and return new interface instance
			return NewInterfaceInstance(targetIface, underlyingObj), nil
		}

		return nil, fmt.Errorf("type '%s' not found (neither class nor interface)", typeName)
	}

	// Handle object casting
	objVal, ok := internalObj.(*ObjectInstance)
	if !ok {
		return nil, fmt.Errorf("'as' operator requires object instance, got %s", internalObj.Type())
	}

	// Look up class-to-class casting first via TypeSystem
	if targetClassIface := i.typeSystem.LookupClass(typeName); targetClassIface != nil {
		targetClass, _ := targetClassIface.(*ClassInfo)
		// Validate that the object's actual runtime type is compatible with the target
		// Need concrete ClassInfo for isClassCompatible
		if concreteClass, ok := objVal.Class.(*ClassInfo); ok {
			if !isClassCompatible(concreteClass, targetClass) {
				return nil, fmt.Errorf("instance of type '%s' cannot be cast to class '%s'", objVal.Class.GetName(), targetClass.Name)
			}
		} else {
			return nil, fmt.Errorf("object has invalid class type")
		}

		// Cast is valid - return the same object
		return objVal, nil
	}

	// Try interface casting
	if iface := i.lookupInterfaceInfo(typeName); iface != nil {
		// Validate that the object's class implements the interface
		// Need concrete ClassInfo for classImplementsInterface
		if concreteClass, ok := objVal.Class.(*ClassInfo); ok {
			if !classImplementsInterface(concreteClass, iface) {
				return nil, fmt.Errorf("class '%s' does not implement interface '%s'", objVal.Class.GetName(), iface.Name)
			}
		} else {
			return nil, fmt.Errorf("object has invalid class type")
		}

		// Create and return the interface instance
		return NewInterfaceInstance(iface, objVal), nil
	}

	return nil, fmt.Errorf("type '%s' not found (neither class nor interface)", typeName)
}

// CastToClass performs class type casting for TypeName(expr) expressions.
func (i *Interpreter) CastToClass(val evaluator.Value, className string, node ast.Expression) evaluator.Value {
	// Convert to internal type
	internalVal := val.(Value)

	// Look up the class
	classInfo := i.lookupClassInfo(className)
	if classInfo == nil {
		return nil // Not a class type, caller will try other options
	}

	// Use the existing castToClass method
	return i.castToClass(internalVal, classInfo, node)
}

// GetObjectInstanceFromValue extracts ObjectInstance from a Value.
// Returns the ObjectInstance interface{} if the value is an ObjectInstance, nil otherwise.
// Task 3.5.29: GetObjectInstanceFromValue REMOVED - use ObjectValue interface type assertion
// Replacement: if _, ok := val.(evaluator.ObjectValue); ok { ... }

// CreateTypeCastWrapper creates a TypeCastValue wrapper.
// Returns the TypeCastValue wrapper or nil if class not found.
func (i *Interpreter) CreateTypeCastWrapper(className string, obj evaluator.Value) evaluator.Value {
	// Convert to internal type
	internalObj := obj.(Value)

	// Look up the class
	classInfo := i.lookupClassInfo(className)
	if classInfo == nil {
		return nil // Class not found
	}

	// Create and return the TypeCastValue wrapper
	return &TypeCastValue{
		Object:     internalObj,
		StaticType: classInfo,
	}
}

// RaiseTypeCastException raises a type cast exception.
func (i *Interpreter) RaiseTypeCastException(message string, node ast.Node) {
	pos := node.Pos()
	fullMessage := fmt.Sprintf("%s [line: %d, column: %d]", message, pos.Line, pos.Column)
	i.raiseException("Exception", fullMessage, &pos)
}

// ===== Metaclass Operations =====

// Task 3.5.27: CreateClassValue REMOVED - zero callers
// Task 3.5.27: GetClassName(obj) REMOVED - zero callers (use ObjectValue.ClassName())
// Task 3.5.27: GetClassType(obj) REMOVED - zero callers (use ObjectValue.GetClassType())
// Task 3.5.27: GetClassNameFromClassInfo REMOVED - zero callers
// Task 3.5.27: GetClassTypeFromClassInfo REMOVED - zero callers
// Task 3.5.27: GetClassVariableFromClassInfo REMOVED - zero callers
