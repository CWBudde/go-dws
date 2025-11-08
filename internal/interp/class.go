package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ClassInfo represents runtime class metadata.
// It stores information about a class's structure including fields, methods,
// parent class, and constructor/destructor.
type ClassInfo struct {
	Constructor          *ast.FunctionDecl
	Constructors         map[string]*ast.FunctionDecl
	ConstructorOverloads map[string][]*ast.FunctionDecl // Task 9.67: Overloaded constructors
	Fields               map[string]types.Type
	ClassVars            map[string]Value
	Constants            map[string]*ast.ConstDecl      // Task 9.20-9.22: Class constants
	ConstantValues       map[string]Value               // Task 9.22: Evaluated constant values
	Methods              map[string]*ast.FunctionDecl
	MethodOverloads      map[string][]*ast.FunctionDecl // Task 9.67: Overloaded instance methods
	ClassMethods         map[string]*ast.FunctionDecl
	ClassMethodOverloads map[string][]*ast.FunctionDecl // Task 9.67: Overloaded class methods
	Properties           map[string]*types.PropertyInfo
	Destructor           *ast.FunctionDecl
	Parent               *ClassInfo
	Operators            *runtimeOperatorRegistry
	ExternalName         string
	Name                 string
	IsExternal           bool
	IsAbstract           bool
}

// NewClassInfo creates a new ClassInfo with the given name.
// Fields, Methods, ClassVars, ClassMethods, and Properties maps are initialized as empty.
func NewClassInfo(name string) *ClassInfo {
	return &ClassInfo{
		Name:                 name,
		Parent:               nil,
		Fields:               make(map[string]types.Type),
		ClassVars:            make(map[string]Value),
		Constants:            make(map[string]*ast.ConstDecl),
		ConstantValues:       make(map[string]Value),
		Methods:              make(map[string]*ast.FunctionDecl),
		MethodOverloads:      make(map[string][]*ast.FunctionDecl),
		ClassMethods:         make(map[string]*ast.FunctionDecl),
		ClassMethodOverloads: make(map[string][]*ast.FunctionDecl),
		Operators:            newRuntimeOperatorRegistry(),
		Constructors:         make(map[string]*ast.FunctionDecl),
		ConstructorOverloads: make(map[string][]*ast.FunctionDecl),
		Properties:           make(map[string]*types.PropertyInfo),
	}
}

// ObjectInstance represents a runtime instance of a class.
// It implements the Value interface so it can be used as a runtime value.
type ObjectInstance struct {
	// Class points to the class metadata
	Class *ClassInfo

	// Fields maps field names to their runtime values
	Fields map[string]Value
}

// NewObjectInstance creates a new object instance of the given class.
// Field values are initialized as an empty map.
func NewObjectInstance(class *ClassInfo) *ObjectInstance {
	return &ObjectInstance{
		Class:  class,
		Fields: make(map[string]Value),
	}
}

// GetField retrieves the value of a field by name.
// Returns nil if the field doesn't exist or hasn't been set.
func (o *ObjectInstance) GetField(name string) Value {
	// Check if field is defined in class
	if _, exists := o.Class.Fields[name]; !exists {
		// Field not defined in class
		return nil
	}

	// Return the field value (may be nil if not yet set)
	return o.Fields[name]
}

// SetField sets the value of a field by name.
// The field must be defined in the class's field map.
func (o *ObjectInstance) SetField(name string, value Value) {
	// Only set if field is defined in class
	if _, exists := o.Class.Fields[name]; exists {
		o.Fields[name] = value
	}
}

// GetMethod looks up a method by name in this object's class.
// It searches the class hierarchy, starting with the object's class
// and walking up through parent classes until the method is found.
// Returns nil if the method is not found in the class hierarchy.
//
// This implements method resolution order (MRO) and supports method overriding:
//   - If a child class defines a method with the same name as a parent class method,
//     the child's method is returned (overriding).
func (o *ObjectInstance) GetMethod(name string) *ast.FunctionDecl {
	return o.Class.lookupMethod(name)
}

// lookupMethod searches for a method in the class hierarchy.
// It starts with the current class and walks up the parent chain.
// Returns the first method found, or nil if not found.
func (c *ClassInfo) lookupMethod(name string) *ast.FunctionDecl {
	// Check current class
	if method, exists := c.Methods[name]; exists {
		return method
	}

	// Check parent class (recursive)
	if c.Parent != nil {
		return c.Parent.lookupMethod(name)
	}

	// Not found
	return nil
}

// lookupProperty searches for a property in the class hierarchy.
// It starts with the current class and walks up the parent chain.
// Returns the first property found, or nil if not found.
func (c *ClassInfo) lookupProperty(name string) *types.PropertyInfo {
	// Check current class
	if prop, exists := c.Properties[name]; exists {
		return prop
	}

	// Check parent class (recursive)
	if c.Parent != nil {
		return c.Parent.lookupProperty(name)
	}

	// Not found
	return nil
}

// lookupConstant searches for a constant in the class hierarchy.
// It starts with the current class and walks up the parent chain.
// Returns the ConstDecl and the ClassInfo that owns it, or (nil, nil) if not found.
func (c *ClassInfo) lookupConstant(name string) (*ast.ConstDecl, *ClassInfo) {
	// Check current class
	if constDecl, exists := c.Constants[name]; exists {
		return constDecl, c
	}

	// Check parent class (recursive)
	if c.Parent != nil {
		return c.Parent.lookupConstant(name)
	}

	// Not found
	return nil, nil
}

// getDefaultProperty searches for the default property in the class hierarchy.
// Returns the default property if found, or nil if no default property exists.
// Task 9.16: Support obj[index] routing to obj.DefaultProperty[index]
func (c *ClassInfo) getDefaultProperty() *types.PropertyInfo {
	// Check current class
	for _, prop := range c.Properties {
		if prop.IsDefault {
			return prop
		}
	}

	// Check parent class (recursive)
	if c.Parent != nil {
		return c.Parent.getDefaultProperty()
	}

	// Not found
	return nil
}

// lookupOperator searches for a class operator in the hierarchy.
func (c *ClassInfo) lookupOperator(operator string, operandTypes []string) (*runtimeOperatorEntry, bool) {
	if c == nil {
		return nil, false
	}
	if c.Operators != nil {
		if entry, ok := c.Operators.lookup(operator, operandTypes); ok {
			return entry, true
		}
	}
	if c.Parent != nil {
		return c.Parent.lookupOperator(operator, operandTypes)
	}
	return nil, false
}

// HasConstructor checks whether the class or its ancestors declare a constructor with the given name.
// Task 9.82: Check both single constructors and overloaded constructors
// Task 9.82: Case-insensitive lookup (DWScript is case-insensitive)
func (c *ClassInfo) HasConstructor(name string) bool {
	if c == nil {
		return false
	}
	// Case-insensitive search through constructors
	for ctorName := range c.Constructors {
		if strings.EqualFold(ctorName, name) {
			return true
		}
	}
	// Task 9.82: Also check constructor overloads (case-insensitive)
	for ctorName, overloads := range c.ConstructorOverloads {
		if strings.EqualFold(ctorName, name) && len(overloads) > 0 {
			return true
		}
	}
	if c.Parent != nil {
		return c.Parent.HasConstructor(name)
	}
	return false
}

// InheritsFrom reports whether the class or any of its ancestors has the given name.
// The check is case-sensitive and includes the class itself.
func (c *ClassInfo) InheritsFrom(name string) bool {
	for current := c; current != nil; current = current.Parent {
		if current.Name == name {
			return true
		}
	}
	return false
}

// ============================================================================
// Value Interface Implementation
// ============================================================================

// Type returns "OBJECT" to indicate this is an object instance.
func (o *ObjectInstance) Type() string {
	return "OBJECT"
}

// String returns a string representation of the object instance.
// Format: "TClassName instance"
func (o *ObjectInstance) String() string {
	return fmt.Sprintf("%s instance", o.Class.Name)
}

// IsInstanceOf checks whether the object derives from the given class.
func (o *ObjectInstance) IsInstanceOf(target *ClassInfo) bool {
	if o == nil || o.Class == nil || target == nil {
		return false
	}
	current := o.Class
	for current != nil {
		if current.Name == target.Name {
			return true
		}
		current = current.Parent
	}
	return false
}

// Helper function to check if a value is an ObjectInstance
func isObject(v Value) bool {
	_, ok := v.(*ObjectInstance)
	return ok
}

// AsObject attempts to cast a Value to an ObjectInstance.
// Returns the ObjectInstance and true if successful, or nil and false if not.
func AsObject(v Value) (*ObjectInstance, bool) {
	obj, ok := v.(*ObjectInstance)
	return obj, ok
}

// ============================================================================
// ClassValue - Metaclass Runtime Value
// ============================================================================

// ClassValue represents a class reference (metaclass value) at runtime.
// In DWScript, when you write "TMyClass" in an expression, it represents
// a reference to the class type itself, not an instance.
//
// Example usage:
//   var cls: class of TAnimal;
//   cls := TDog;              // Assign class reference
//   obj := cls.Create;        // Call constructor through metaclass
//
// Task 9.72: Metaclass runtime values
type ClassValue struct {
	// ClassInfo points to the class metadata
	ClassInfo *ClassInfo
}

// Type returns "CLASS" to indicate this is a class reference.
func (c *ClassValue) Type() string {
	return "CLASS"
}

// String returns a string representation of the class reference.
// Format: "class TClassName"
func (c *ClassValue) String() string {
	if c.ClassInfo != nil {
		return fmt.Sprintf("class %s", c.ClassInfo.Name)
	}
	return "class <nil>"
}

// IsAssignableTo checks if this class reference can be assigned to a variable
// of the given metaclass type. This implements the assignment compatibility
// rules for metaclasses.
//
// For example:
// - TDog (ClassValue) can be assigned to "class of TAnimal" if TDog inherits from TAnimal
// - TDog cannot be assigned to "class of TCat"
//
// Returns true if assignment is allowed, false otherwise.
func (c *ClassValue) IsAssignableTo(targetClass *ClassInfo) bool {
	if c.ClassInfo == nil || targetClass == nil {
		return false
	}

	// Check if c.ClassInfo is targetClass or derives from it
	current := c.ClassInfo
	for current != nil {
		if current.Name == targetClass.Name {
			return true
		}
		current = current.Parent
	}
	return false
}

// Helper function to check if a value is a ClassValue
func isClassValue(v Value) bool {
	_, ok := v.(*ClassValue)
	return ok
}

// AsClassValue attempts to cast a Value to a ClassValue.
// Returns the ClassValue and true if successful, or nil and false if not.
func AsClassValue(v Value) (*ClassValue, bool) {
	cls, ok := v.(*ClassValue)
	return cls, ok
}
