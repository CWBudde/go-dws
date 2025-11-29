package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
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

	// Task 3.5.46: Look up class via TypeSystem (case-insensitive)
	classInfoIface := i.typeSystem.LookupClass(className)
	if classInfoIface == nil {
		return nil, fmt.Errorf("class '%s' not found", className)
	}
	classInfo, ok := classInfoIface.(*ClassInfo)
	if !ok {
		return nil, fmt.Errorf("class '%s' has invalid type", className)
	}

	// Check if trying to instantiate an abstract class
	if classInfo.IsAbstract {
		return nil, fmt.Errorf("Trying to create an instance of an abstract class")
	}

	// Check if trying to instantiate an external class
	if classInfo.IsExternal {
		return nil, fmt.Errorf("cannot instantiate external class '%s' - external classes are not supported", className)
	}

	// Create new object instance
	obj := NewObjectInstance(classInfo)

	// Initialize fields with default values
	savedEnv := i.env
	tempEnv := NewEnclosedEnvironment(i.env)
	i.env = tempEnv

	for fieldName, fieldType := range classInfo.Fields {
		var fieldValue Value
		if fieldDecl, hasDecl := classInfo.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
			fieldValue = i.Eval(fieldDecl.InitValue)
			if isError(fieldValue) {
				i.env = savedEnv
				return nil, fmt.Errorf("failed to initialize field '%s': %v", fieldName, fieldValue)
			}
		} else {
			fieldValue = getZeroValueForType(fieldType, nil)
		}
		obj.SetField(fieldName, fieldValue)
	}

	i.env = savedEnv

	// Call constructor if it exists
	constructorNameLower := ident.Normalize("Create")
	if constructor, exists := classInfo.Constructors[constructorNameLower]; exists {
		tempEnv := NewEnclosedEnvironment(i.env)
		tempEnv.Define("Self", obj)
		i.env = tempEnv

		result := i.callUserFunction(constructor, internalArgs)

		i.env = savedEnv

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
// Task 3.5.126f: Callback for complex constructor execution (method body + Self binding).
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

	// Look up constructor
	constructorNameNorm := ident.Normalize(constructorName)
	constructor, exists := classInfo.Constructors[constructorNameNorm]
	if !exists {
		if len(internalArgs) > 0 {
			return fmt.Errorf("no constructor '%s' found for class '%s' with %d arguments", constructorName, classInfo.Name, len(internalArgs))
		}
		// No constructor and no args - OK
		return nil
	}

	// Execute constructor in a new environment with Self bound
	savedEnv := i.env
	tempEnv := NewEnclosedEnvironment(i.env)
	tempEnv.Define("Self", objectInstance)
	i.env = tempEnv

	result := i.callUserFunction(constructor, internalArgs)

	i.env = savedEnv

	// Propagate constructor errors
	if isError(result) {
		return fmt.Errorf("constructor failed: %v", result)
	}

	return nil
}

// ===== Type Checking and Casting =====

// CheckType checks if an object is of a specified type (implements 'is' operator).
// Task 3.5.34: Extended to support both class hierarchy and interface implementation checking.
// GetClassMetadataFromValue extracts ClassMetadata from an object value.
// Task 3.5.140: Helper method to extract metadata from ObjectInstance or InterfaceInstance.
func (i *Interpreter) GetClassMetadataFromValue(obj evaluator.Value) *runtime.ClassMetadata {
	// Convert to internal type
	internalObj := obj.(Value)

	// Handle nil
	if internalObj == nil {
		return nil
	}

	// Check for ObjectInstance
	if objVal, ok := internalObj.(*ObjectInstance); ok {
		if objVal.Class != nil {
			return objVal.Class.Metadata
		}
		return nil
	}

	// Check for InterfaceInstance - extract the underlying object's class
	if ifaceVal, ok := internalObj.(*InterfaceInstance); ok {
		if ifaceVal.Object != nil && ifaceVal.Object.Class != nil {
			return ifaceVal.Object.Class.Metadata
		}
		return nil
	}

	// Check for TypeCastValue - extract the wrapped object's class
	if typeCastVal, ok := internalObj.(*TypeCastValue); ok {
		if typeCastVal.Object != nil {
			// Recursively extract from wrapped value
			return i.GetClassMetadataFromValue(typeCastVal.Object)
		}
		return nil
	}

	return nil
}

func (i *Interpreter) CheckType(obj evaluator.Value, typeName string) bool {
	// Convert to internal type
	internalObj := obj.(Value)

	// Handle nil - nil is not an instance of any type
	if _, isNil := internalObj.(*NilValue); isNil {
		return false
	}

	// Check if it's an object
	objVal, ok := internalObj.(*ObjectInstance)
	if !ok {
		return false
	}

	// Get class info
	classInfo := objVal.Class
	if classInfo == nil {
		return false
	}

	// Check if the object's class matches (case-insensitive)
	if ident.Equal(classInfo.Name, typeName) {
		return true
	}

	// Check parent class hierarchy
	current := classInfo.Parent
	for current != nil {
		if ident.Equal(current.Name, typeName) {
			return true
		}
		current = current.Parent
	}

	// Task 3.5.34: Check if the target is an interface and if the object's class implements it
	if iface, exists := i.interfaces[ident.Normalize(typeName)]; exists {
		return classImplementsInterface(objVal.Class, iface)
	}

	return false
}

// CastType performs type casting (implements 'as' operator).
// Task 3.5.35: Extended to fully support type casting with interface wrapping/unwrapping.
// Task 3.5.141: DEPRECATED - Migrated to Evaluator.castType(). Kept for backwards compatibility with old interpreter path.
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
		// Task 3.5.46: Check if target is a class via TypeSystem
		if targetClassIface := i.typeSystem.LookupClass(typeName); targetClassIface != nil {
			targetClass, _ := targetClassIface.(*ClassInfo)
			// Interface-to-class casting: extract the underlying object
			underlyingObj := intfInst.Object
			if underlyingObj == nil {
				return nil, fmt.Errorf("cannot cast nil interface to class '%s'", targetClass.Name)
			}

			// Check if the underlying object's class is compatible with the target class
			if !isClassCompatible(underlyingObj.Class, targetClass) {
				return nil, fmt.Errorf("cannot cast interface of '%s' to class '%s'", underlyingObj.Class.Name, targetClass.Name)
			}

			// Cast is valid - return the underlying object
			return underlyingObj, nil
		}

		// Check if target is an interface
		if targetIface, isInterface := i.interfaces[ident.Normalize(typeName)]; isInterface {
			// Interface-to-interface casting
			underlyingObj := intfInst.Object
			if underlyingObj == nil {
				// DWScript: nil interface cast to interface yields nil interface wrapper
				return &InterfaceInstance{Interface: targetIface, Object: nil}, nil
			}

			// Check if the underlying object's class implements the target interface
			if !classImplementsInterface(underlyingObj.Class, targetIface) {
				return nil, fmt.Errorf("cannot cast interface of '%s' to interface '%s'", underlyingObj.Class.Name, targetIface.Name)
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

	// Task 3.5.46: Try class-to-class casting first via TypeSystem
	if targetClassIface := i.typeSystem.LookupClass(typeName); targetClassIface != nil {
		targetClass, _ := targetClassIface.(*ClassInfo)
		// Validate that the object's actual runtime type is compatible with the target
		if !isClassCompatible(objVal.Class, targetClass) {
			return nil, fmt.Errorf("instance of type '%s' cannot be cast to class '%s'", objVal.Class.Name, targetClass.Name)
		}

		// Cast is valid - return the same object
		return objVal, nil
	}

	// Try interface casting
	if iface, exists := i.interfaces[ident.Normalize(typeName)]; exists {
		// Validate that the object's class implements the interface
		if !classImplementsInterface(objVal.Class, iface) {
			return nil, fmt.Errorf("class '%s' does not implement interface '%s'", objVal.Class.Name, iface.Name)
		}

		// Create and return the interface instance
		return NewInterfaceInstance(iface, objVal), nil
	}

	return nil, fmt.Errorf("type '%s' not found (neither class nor interface)", typeName)
}

// CastToClass performs class type casting for TypeName(expr) expressions.
// Task 3.5.94: Adapter method for type cast migration. Uses existing castToClass logic.
// Task 3.5.141: DEPRECATED - Migrated to Evaluator.castToClassType(). Kept for backwards compatibility with old interpreter path.
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
// Task 3.5.141: Helper to extract ObjectInstance for type casting.
// Returns the ObjectInstance interface{} if the value is an ObjectInstance, nil otherwise.
func (i *Interpreter) GetObjectInstanceFromValue(val evaluator.Value) interface{} {
	// Convert to internal type
	internalVal := val.(Value)

	// Type assert to ObjectInstance
	if objInst, ok := internalVal.(*ObjectInstance); ok {
		return objInst
	}

	return nil
}

// GetInterfaceInstanceFromValue extracts InterfaceInstance from a Value.
// Task 3.5.141: Helper to extract InterfaceInstance for interface casting.
// Returns (interfaceInfo, underlyingObject) if the value is an InterfaceInstance, (nil, nil) otherwise.
func (i *Interpreter) GetInterfaceInstanceFromValue(val evaluator.Value) (interfaceInfo interface{}, underlyingObject interface{}) {
	// Convert to internal type
	internalVal := val.(Value)

	// Type assert to InterfaceInstance
	if intfInst, ok := internalVal.(*InterfaceInstance); ok {
		return intfInst.Interface, intfInst.Object
	}

	return nil, nil
}

// CreateInterfaceWrapper creates an InterfaceInstance wrapper.
// Task 3.5.141: Helper to create interface wrappers for 'as' operator.
// Returns the InterfaceInstance wrapper or error if interface not found.
func (i *Interpreter) CreateInterfaceWrapper(interfaceName string, obj evaluator.Value) (evaluator.Value, error) {
	// Convert to internal type
	var internalObj *ObjectInstance
	if obj != nil {
		if o, ok := obj.(*ObjectInstance); ok {
			internalObj = o
		} else {
			return nil, fmt.Errorf("cannot create interface wrapper for non-object type: %s", obj.Type())
		}
	}

	// Look up the interface
	iface, exists := i.interfaces[ident.Normalize(interfaceName)]
	if !exists {
		return nil, fmt.Errorf("interface '%s' not found", interfaceName)
	}

	// Create and return the interface instance
	return NewInterfaceInstance(iface, internalObj), nil
}

// CreateTypeCastWrapper creates a TypeCastValue wrapper.
// Task 3.5.141: Helper to create TypeCastValue for TypeName(expr) casts.
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
// Task 3.5.141: Helper to raise exceptions for invalid TypeName(expr) casts.
// This matches the behavior of castToClass which raises exceptions.
func (i *Interpreter) RaiseTypeCastException(message string, node ast.Node) {
	pos := node.Pos()
	fullMessage := fmt.Sprintf("%s [line: %d, column: %d]", message, pos.Line, pos.Column)
	i.raiseException("Exception", fullMessage, &pos)
}

// CheckImplements checks if an object/class implements an interface.
// Task 3.5.36: Adapter method for 'implements' operator.
// Supports ObjectInstance, ClassValue, and ClassInfoValue inputs.
func (i *Interpreter) CheckImplements(obj evaluator.Value, interfaceName string) (bool, error) {
	// Convert to internal type
	internalObj := obj.(Value)

	// Handle nil - nil implements no interfaces
	if _, isNil := internalObj.(*NilValue); isNil {
		return false, nil
	}

	// Extract ClassInfo from different value types
	var classInfo *ClassInfo

	if objInst, ok := internalObj.(*ObjectInstance); ok {
		// Object instance - extract class
		classInfo = objInst.Class
	} else if classVal, ok := internalObj.(*ClassValue); ok {
		// Class reference (e.g., from metaclass variable: var cls: class of TParent)
		classInfo = classVal.ClassInfo
	} else if classInfoVal, ok := internalObj.(*ClassInfoValue); ok {
		// Class type identifier (e.g., TMyImplementation in: if TMyImplementation implements IMyInterface then)
		classInfo = classInfoVal.ClassInfo
	} else {
		return false, fmt.Errorf("'implements' operator requires object instance or class type, got %s", internalObj.Type())
	}

	// Guard against nil ClassInfo (e.g., uninitialized metaclass variables)
	if classInfo == nil {
		return false, nil
	}

	// Look up the interface in the registry
	iface, exists := i.interfaces[ident.Normalize(interfaceName)]
	if !exists {
		return false, fmt.Errorf("interface '%s' not found", interfaceName)
	}

	// Check if the class implements the interface
	// 'implements' operator in DWScript only considers explicitly declared interfaces,
	// not interfaces inherited through other interfaces.
	return classExplicitlyImplementsInterface(classInfo, iface), nil
}

// ===== Metaclass Operations =====

// CreateClassValue creates a ClassValue (metaclass reference) from a class name.
// Task 3.5.85: Adapter method for returning metaclass references from VisitIdentifier.
func (i *Interpreter) CreateClassValue(className string) (evaluator.Value, error) {
	// Look up the class in the registry (case-insensitive)
	for name, classInfo := range i.classes {
		if ident.Equal(name, className) {
			return &ClassValue{ClassInfo: classInfo}, nil
		}
	}
	return nil, fmt.Errorf("class '%s' not found", className)
}

// GetClassName returns the class name for an object instance.
func (i *Interpreter) GetClassName(obj evaluator.Value) string {
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return ""
	}
	return objInst.Class.Name
}

// GetClassType returns the ClassValue (metaclass) for an object instance.
func (i *Interpreter) GetClassType(obj evaluator.Value) evaluator.Value {
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return nil
	}
	return &ClassValue{ClassInfo: objInst.Class}
}

// Task 3.5.71: IsClassInfoValue removed - evaluator uses val.Type() == "CLASSINFO" directly

// GetClassNameFromClassInfo returns the class name from a ClassInfoValue.
func (i *Interpreter) GetClassNameFromClassInfo(classInfo evaluator.Value) string {
	classInfoVal, ok := classInfo.(*ClassInfoValue)
	if !ok {
		panic("GetClassNameFromClassInfo called on non-ClassInfoValue value")
	}
	return classInfoVal.ClassInfo.Name
}

// GetClassTypeFromClassInfo returns the ClassValue from a ClassInfoValue.
func (i *Interpreter) GetClassTypeFromClassInfo(classInfo evaluator.Value) evaluator.Value {
	classInfoVal, ok := classInfo.(*ClassInfoValue)
	if !ok {
		panic("GetClassTypeFromClassInfo called on non-ClassInfoValue value")
	}
	return &ClassValue{ClassInfo: classInfoVal.ClassInfo}
}

// GetClassVariableFromClassInfo retrieves a class variable from ClassInfoValue.
func (i *Interpreter) GetClassVariableFromClassInfo(classInfo evaluator.Value, varName string) (evaluator.Value, bool) {
	classInfoVal, ok := classInfo.(*ClassInfoValue)
	if !ok {
		panic("GetClassVariableFromClassInfo called on non-ClassInfoValue value")
	}
	// Case-insensitive lookup to match DWScript semantics
	for name, value := range classInfoVal.ClassInfo.ClassVars {
		if ident.Equal(name, varName) {
			return value, true
		}
	}
	return nil, false
}
