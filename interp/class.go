package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/types"
)

// ClassInfo represents runtime class metadata.
// It stores information about a class's structure including fields, methods,
// parent class, and constructor/destructor.
type ClassInfo struct {
	// Name is the class name (e.g., "TPoint")
	Name string

	// Parent is the parent class (nil for root classes)
	Parent *ClassInfo

	// Fields maps field names to their types
	Fields map[string]types.Type

	// Methods maps method names to their AST declarations
	Methods map[string]*ast.FunctionDecl

	// Constructor is the constructor method (usually "Create")
	Constructor *ast.FunctionDecl

	// Destructor is the destructor method (if present)
	Destructor *ast.FunctionDecl
}

// NewClassInfo creates a new ClassInfo with the given name.
// Fields and Methods maps are initialized as empty.
func NewClassInfo(name string) *ClassInfo {
	return &ClassInfo{
		Name:    name,
		Parent:  nil,
		Fields:  make(map[string]types.Type),
		Methods: make(map[string]*ast.FunctionDecl),
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
// - If a child class defines a method with the same name as a parent class method,
//   the child's method is returned (overriding).
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
