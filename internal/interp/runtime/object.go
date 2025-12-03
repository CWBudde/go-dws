// Package runtime provides runtime value types for the DWScript interpreter.
// This file contains ObjectInstance, the runtime representation of class instances.
package runtime

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ObjectInstance represents a runtime instance of a class.
// It implements the Value interface so it can be used as a runtime value.
//
// Task 3.5.17: Moved from internal/interp/class.go to runtime package.
// Uses IClassInfo interface instead of *ClassInfo to avoid circular import.
type ObjectInstance struct {
	// Class points to the class metadata (via interface to avoid circular import)
	Class IClassInfo

	// Fields maps field names to their runtime values
	Fields map[string]Value

	// RefCount tracks interface references to this object for lifetime management
	// Task 9.1.5: Objects held by interfaces use reference counting
	// - Starts at 0 when object is created; incremented when assigned to variable or interface
	// - Increments when assigned to another variable or interface
	// - Decrements when variable is reassigned, set to nil, or goes out of scope
	// - Destructor called when RefCount reaches 0
	RefCount int

	// Destroyed indicates whether the object's destructor has completed.
	// DestroyCallDepth tracks nested Destroy calls during inherited dispatch.
	Destroyed        bool
	DestroyCallDepth int
}

// NewObjectInstance creates a new object instance of the given class.
// Field values are initialized as an empty map.
// Task 9.1.5: RefCount starts at 0; incremented when assigned to variable or interface
func NewObjectInstance(class IClassInfo) *ObjectInstance {
	return &ObjectInstance{
		Class:    class,
		Fields:   make(map[string]Value),
		RefCount: 0, // Start with reference count of 0
	}
}

// GetField retrieves the value of a field by name.
// Returns nil if the field doesn't exist or hasn't been set.
func (o *ObjectInstance) GetField(name string) Value {
	// Guard against nil class
	if o.Class == nil {
		return nil
	}

	normalizedName := ident.Normalize(name)

	// Try metadata first (AST-free path), walking up the inheritance chain.
	// This avoids missing inherited fields when the child class adds its own fields.
	if fieldMeta := lookupFieldMetadata(o.Class.GetMetadata(), normalizedName); fieldMeta != nil {
		if val, exists := o.Fields[normalizedName]; exists {
			return val
		}
		if val, exists := o.Fields[name]; exists {
			return val
		}
		return nil
	}

	// Legacy fallback: Check if field exists via interface method
	if !o.Class.FieldExists(normalizedName) {
		// Also check non-normalized key for backward compatibility
		if !o.Class.FieldExists(name) {
			// Final fallback: If neither metadata nor Fields map have this field,
			// but the field was directly set in o.Fields (e.g., in tests), return it
			if val, exists := o.Fields[normalizedName]; exists {
				return val
			}
			if val, exists := o.Fields[name]; exists {
				return val
			}
			return nil
		}
		// Return field with original name key
		return o.Fields[name]
	}

	// Return field with normalized name
	return o.Fields[normalizedName]
}

// SetField sets the value of a field by name.
// The field must be defined in the class's field map.
func (o *ObjectInstance) SetField(name string, value Value) {
	// Guard against nil class
	if o.Class == nil {
		return
	}

	normalizedName := ident.Normalize(name)

	// Try metadata first (AST-free path), walking up the inheritance chain.
	if lookupFieldMetadata(o.Class.GetMetadata(), normalizedName) != nil {
		o.Fields[normalizedName] = value
		return
	}

	// Legacy fallback: Check if field exists via interface method
	if o.Class.FieldExists(normalizedName) {
		o.Fields[normalizedName] = value
		return
	}

	// Also check non-normalized key for backward compatibility
	if o.Class.FieldExists(name) {
		o.Fields[name] = value
		return
	}

	// Final fallback: Allow setting fields even if not declared (for tests)
	// In production code, fields should always be declared in the class
	o.Fields[normalizedName] = value
}

// GetMethod looks up a method by name in this object's class.
// It searches the class hierarchy, starting with the object's class
// and walking up through parent classes until the method is found.
// Returns nil if the method is not found in the class hierarchy.
//
// This implements method resolution order (MRO) and supports method overriding:
//   - If a child class defines a method with the same name as a parent class method,
//     the child's method is returned (overriding).
//
// Note: This performs static method resolution (not virtual dispatch).
// Virtual dispatch is implemented inline in objects_methods.go where needed.
func (o *ObjectInstance) GetMethod(name string) *ast.FunctionDecl {
	if o.Class == nil {
		return nil
	}
	return o.Class.LookupMethod(name)
}

// LookupProperty searches for a property by name in the class hierarchy.
// Returns a PropertyDescriptor that can be used for property access.
//
// Note: Returns *PropertyDescriptor (not *evaluator.PropertyDescriptor) to avoid
// importing evaluator package (which would create circular import).
func (o *ObjectInstance) LookupProperty(name string) *PropertyDescriptor {
	if o.Class == nil {
		return nil
	}

	propInfo := o.Class.LookupProperty(name)
	if propInfo == nil {
		return nil
	}

	return &PropertyDescriptor{
		Name:      propInfo.Name,
		IsIndexed: propInfo.IsIndexed,
		IsDefault: propInfo.IsDefault,
		Impl:      propInfo.Impl, // Store the original PropertyInfo for later use
	}
}

// GetDefaultProperty returns the default property for this object's class, if any.
// Task 3.5.99a: Implements PropertyAccessor interface.
// Returns a PropertyDescriptor wrapping types.PropertyInfo, or nil if no default property exists.
func (o *ObjectInstance) GetDefaultProperty() *PropertyDescriptor {
	if o.Class == nil {
		return nil
	}

	propInfo := o.Class.GetDefaultProperty()
	if propInfo == nil {
		return nil
	}

	return &PropertyDescriptor{
		Name:      propInfo.Name,
		IsIndexed: propInfo.IsIndexed,
		IsDefault: propInfo.IsDefault,
		Impl:      propInfo.Impl, // Store the original PropertyInfo for later use
	}
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
	if o.Class == nil {
		return "<nil> instance"
	}
	return fmt.Sprintf("%s instance", o.Class.GetName())
}

// ============================================================================
// ObjectValue Interface Implementation (evaluator.ObjectValue)
// ============================================================================
// These methods allow the evaluator to work with ObjectInstance without
// needing an adapter.

// ClassName returns the class name of this object instance.
// Task 3.5.72: Implements evaluator.ObjectValue interface.
func (o *ObjectInstance) ClassName() string {
	if o == nil || o.Class == nil {
		return ""
	}
	return o.Class.GetName()
}

// GetClassType returns the class type (metaclass) for this object instance.
// Task 3.5.156: Implements evaluator.ObjectValue interface.
//
// Note: This returns a Value that should be a ClassValue. However, ClassValue
// is defined in interp package, so we can't construct it here. The calling code
// in interp package must handle this conversion.
func (o *ObjectInstance) GetClassType() Value {
	if o == nil || o.Class == nil {
		return nil
	}
	// Return a placeholder that interp package will convert to ClassValue
	// This is a temporary solution until ClassValue is also moved to runtime
	return &classTypeProxy{class: o.Class}
}

// classTypeProxy is a temporary placeholder for ClassValue.
// The interp package will convert this to a proper ClassValue.
type classTypeProxy struct {
	class IClassInfo
}

func (c *classTypeProxy) Type() string {
	return "CLASS"
}

func (c *classTypeProxy) String() string {
	if c.class != nil {
		return fmt.Sprintf("class %s", c.class.GetName())
	}
	return "class <nil>"
}

// HasProperty checks if this object's class has a property with the given name.
// The check includes the entire class hierarchy.
// Task 3.5.72: Implements evaluator.ObjectValue interface.
func (o *ObjectInstance) HasProperty(name string) bool {
	if o == nil || o.Class == nil {
		return false
	}
	return o.Class.LookupProperty(name) != nil
}

// HasMethod checks if this object's class has a method with the given name.
// Task 3.5.72: Implements evaluator.ObjectValue interface.
func (o *ObjectInstance) HasMethod(name string) bool {
	if o == nil || o.Class == nil {
		return false
	}
	methods := o.Class.GetMethodsMap()
	if methods == nil {
		return false
	}
	_, exists := methods[ident.Normalize(name)]
	return exists
}

// GetClassVar retrieves a class variable value by name from this object's class hierarchy.
// Returns the value and true if found, nil and false otherwise.
// Task 3.5.86: Implements evaluator.ObjectValue interface for direct class variable access.
func (o *ObjectInstance) GetClassVar(name string) (Value, bool) {
	if o == nil || o.Class == nil {
		return nil, false
	}
	value, owningClass := o.Class.LookupClassVar(name)
	if owningClass == nil {
		return nil, false
	}
	return value, true
}

// CallInheritedMethod calls a method from the parent class.
// Task 3.5.114: Implements evaluator.ObjectValue interface for direct inherited method calls.
// The methodExecutor callback is used to execute the method once resolved.
// Returns an error value if the class has no parent or the method is not found.
func (o *ObjectInstance) CallInheritedMethod(methodName string, args []Value, methodExecutor func(methodDecl any, args []Value) Value) Value {
	// Validate object state
	if o == nil || o.Class == nil {
		return newError("object has no class information")
	}

	// Check parent class exists
	parentInfo := o.Class.GetParent()
	if parentInfo == nil {
		return newError("class '%s' has no parent class", o.Class.GetName())
	}

	// Find method in parent (case-insensitive)
	methodNameLower := ident.Normalize(methodName)
	methods := parentInfo.GetMethodsMap()
	if methods == nil {
		return newError("parent class '%s' has no methods", parentInfo.GetName())
	}

	method, exists := methods[methodNameLower]
	if !exists {
		return newError("method, property, or field '%s' not found in parent class '%s'", methodName, parentInfo.GetName())
	}

	// Execute the method using the provided executor callback
	return methodExecutor(method, args)
}

// ReadProperty reads a property value from this object.
// Task 3.5.116: Enables direct property access without adapter.
// The propertyExecutor callback handles interpreter-dependent execution.
func (o *ObjectInstance) ReadProperty(propName string, propertyExecutor func(propInfo any) Value) Value {
	// Validate object state
	if o == nil || o.Class == nil {
		return newError("object has no class information")
	}

	// Look up the property in the class hierarchy
	propInfo := o.Class.LookupProperty(propName)
	if propInfo == nil {
		return newError("property '%s' not found", propName)
	}

	// Execute the property read using the provided executor callback
	// Pass the Impl field which contains the actual *types.PropertyInfo
	return propertyExecutor(propInfo.Impl)
}

// ReadIndexedProperty reads an indexed property value using the provided executor callback.
// Task 3.5.117: Enables direct indexed property access without going through adapter.
// The propInfo is already resolved by PropertyAccessor.LookupProperty or GetDefaultProperty.
func (o *ObjectInstance) ReadIndexedProperty(propInfo any, indices []Value, propertyExecutor func(propInfo any, indices []Value) Value) Value {
	// Validate object state
	if o == nil || o.Class == nil {
		return newError("object has no class information")
	}

	// The propInfo is already resolved by the caller (PropertyAccessor.LookupProperty or GetDefaultProperty)
	// Just call the executor with the property info and indices
	return propertyExecutor(propInfo, indices)
}

// InvokeParameterlessMethod invokes a method if it has zero parameters.
// Task 3.5.119: Implements evaluator.ObjectValue interface.
// Returns (result, true) if method exists and has 0 parameters,
// or (nil, false) if method has parameters (caller should create method pointer).
func (o *ObjectInstance) InvokeParameterlessMethod(methodName string,
	methodExecutor func(methodDecl any) Value) (Value, bool) {
	if o == nil || o.Class == nil {
		return nil, false
	}

	methods := o.Class.GetMethodsMap()
	if methods == nil {
		return nil, false
	}

	method, exists := methods[ident.Normalize(methodName)]
	if !exists {
		return nil, false // Method not found
	}

	if len(method.Parameters) > 0 {
		return nil, false // Has parameters - caller should create method pointer
	}

	// Parameterless method - invoke via callback
	return methodExecutor(method), true
}

// CreateMethodPointer creates a method pointer for a method with parameters.
// Task 3.5.120: Implements evaluator.ObjectValue interface.
// Returns (pointer, true) if method exists and has parameters,
// or (nil, false) if method doesn't exist or has no parameters.
func (o *ObjectInstance) CreateMethodPointer(methodName string,
	pointerCreator func(methodDecl any) Value) (Value, bool) {
	if o == nil || o.Class == nil {
		return nil, false
	}

	methods := o.Class.GetMethodsMap()
	if methods == nil {
		return nil, false
	}

	method, exists := methods[ident.Normalize(methodName)]
	if !exists {
		return nil, false // Method not found
	}

	if len(method.Parameters) == 0 {
		return nil, false // No parameters - caller should use InvokeParameterlessMethod
	}

	// Method has parameters - create pointer via callback
	return pointerCreator(method), true
}

// ============================================================================
// Helper Functions
// ============================================================================

// IsInstanceOf checks whether the object derives from the given class.
// Note: This method takes IClassInfo instead of concrete *ClassInfo type.
func (o *ObjectInstance) IsInstanceOf(target IClassInfo) bool {
	if o == nil || o.Class == nil || target == nil {
		return false
	}
	current := o.Class
	targetName := target.GetName()
	for current != nil {
		if current.GetName() == targetName {
			return true
		}
		current = current.GetParent()
	}
	return false
}

// Helper function to check if a value is an ObjectInstance
func IsObject(v Value) bool {
	_, ok := v.(*ObjectInstance)
	return ok
}

// AsObject attempts to cast a Value to an ObjectInstance.
// Returns the ObjectInstance and true if successful, or nil and false if not.
func AsObject(v Value) (*ObjectInstance, bool) {
	obj, ok := v.(*ObjectInstance)
	return obj, ok
}

// lookupFieldMetadata searches for a field in the class metadata hierarchy.
// Returns the metadata for the field if found, or nil otherwise.
func lookupFieldMetadata(meta *ClassMetadata, normalizedName string) *FieldMetadata {
	for current := meta; current != nil; current = current.Parent {
		if field, ok := current.Fields[normalizedName]; ok {
			return field
		}
	}
	return nil
}

// LookupFieldInHierarchy searches for a field in the class metadata hierarchy.
// This is the exported version of lookupFieldMetadata for use in other packages.
// Returns the metadata for the field if found, or nil otherwise.
func LookupFieldInHierarchy(meta *ClassMetadata, normalizedName string) *FieldMetadata {
	return lookupFieldMetadata(meta, normalizedName)
}

// newError creates an error value.
// This is a helper function to create error values without importing interp package.
func newError(format string, args ...interface{}) Value {
	return &ErrorValue{Message: fmt.Sprintf(format, args...)}
}

// ============================================================================
// Task 3.5.32: ClassMetaProvider and FieldBinder interface implementations
// ============================================================================

// GetClassConstantBySpec looks up a class constant by name in the class hierarchy.
// Task 3.5.32: Implements evaluator.ClassMetaProvider interface for property reads.
// Returns the constant value and true if found, nil and false otherwise.
//
// Note: This does NOT handle lazy evaluation of constant expressions.
// Lazy evaluation requires access to the Interpreter and must be done via a callback.
// For now, only pre-evaluated constants (stored in class info) are returned.
func (o *ObjectInstance) GetClassConstantBySpec(name string) (Value, bool) {
	if o == nil || o.Class == nil {
		return nil, false
	}

	// Check if the class info supports constant lookup
	if constProvider, ok := o.Class.(ClassConstantProvider); ok {
		return constProvider.GetClassConstant(name)
	}

	return nil, false
}

// BindFieldsToEnvironment iterates over all object fields and calls the binder function.
// Task 3.5.32: Implements evaluator.FieldBinder interface for expression-backed properties.
// This allows property expressions like (FWidth * FHeight) to access fields directly.
func (o *ObjectInstance) BindFieldsToEnvironment(binder func(name string, value Value)) {
	if o == nil || o.Fields == nil {
		return
	}

	for name, value := range o.Fields {
		binder(name, value)
	}
}

// ClassConstantProvider is an optional interface for IClassInfo implementations
// that can provide access to class constants.
type ClassConstantProvider interface {
	// GetClassConstant looks up a class constant by name in the class hierarchy.
	// Returns the constant value and true if found, nil and false otherwise.
	GetClassConstant(name string) (Value, bool)
}
