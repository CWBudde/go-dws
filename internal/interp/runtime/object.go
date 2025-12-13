// Package runtime provides runtime value types for the DWScript interpreter.
// This file contains ObjectInstance, the runtime representation of class instances.
package runtime

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ObjectInstance represents a runtime instance of a class.
// Uses IClassInfo interface to avoid circular imports.
type ObjectInstance struct {
	Class            IClassInfo       // Class metadata
	Fields           map[string]Value // Field name to value mapping
	RefCount         int              // Interface reference count
	Destroyed        bool             // Destructor completed
	DestroyCallDepth int              // Nested Destroy call tracking
}

// NewObjectInstance creates a new object instance of the given class.
// RefCount starts at 0 and is incremented when assigned to variables or interfaces.
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
		ReadSpec:  propInfo.ReadSpec,
		WriteSpec: propInfo.WriteSpec,
		IsIndexed: propInfo.IsIndexed,
		IsDefault: propInfo.IsDefault,
		Impl:      propInfo.Impl, // Store the original PropertyInfo for later use
	}
}

// GetDefaultProperty returns the default property for this object's class, if any.
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
		ReadSpec:  propInfo.ReadSpec,
		WriteSpec: propInfo.WriteSpec,
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
func (o *ObjectInstance) ClassName() string {
	if o == nil || o.Class == nil {
		return ""
	}
	return o.Class.GetName()
}

// GetClassType returns the class type (metaclass) for this object instance.
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
func (o *ObjectInstance) HasProperty(name string) bool {
	if o == nil || o.Class == nil {
		return false
	}
	return o.Class.LookupProperty(name) != nil
}

// HasMethod checks if this object's class has a method with the given name.
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

// GetMethodDecl retrieves a method declaration by name from the class hierarchy.
func (o *ObjectInstance) GetMethodDecl(name string) any {
	if o == nil || o.Class == nil {
		return nil
	}
	method := o.Class.LookupMethod(name)
	if method == nil {
		return nil
	}
	return method
}

// GetClassVar retrieves a class variable value by name from the class hierarchy.
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

// CallInheritedMethod calls a method from the parent class using the provided executor callback.
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

// ReadProperty reads a property value using the provided executor callback.
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
func (o *ObjectInstance) ReadIndexedProperty(propInfo any, indices []Value, propertyExecutor func(propInfo any, indices []Value) Value) Value {
	// Validate object state
	if o == nil || o.Class == nil {
		return newError("object has no class information")
	}

	// The propInfo is already resolved by the caller (PropertyAccessor.LookupProperty or GetDefaultProperty)
	// Just call the executor with the property info and indices
	return propertyExecutor(propInfo, indices)
}

// WriteProperty writes a property value using the provided executor callback.
func (o *ObjectInstance) WriteProperty(propName string, value Value, propertyExecutor func(propInfo any, value Value) Value) Value {
	// Validate object state
	if o == nil || o.Class == nil {
		return newError("object has no class information")
	}

	// Look up the property in the class hierarchy
	propInfo := o.Class.LookupProperty(propName)
	if propInfo == nil {
		return newError("property '%s' not found", propName)
	}

	// Execute the property write using the provided executor callback
	// Pass the Impl field which contains the actual *types.PropertyInfo
	return propertyExecutor(propInfo.Impl, value)
}

// WriteIndexedProperty writes an indexed property value using the provided executor callback.
func (o *ObjectInstance) WriteIndexedProperty(propInfo any, indices []Value, value Value, propertyExecutor func(propInfo any, indices []Value, value Value) Value) Value {
	// Validate object state
	if o == nil || o.Class == nil {
		return newError("object has no class information")
	}

	// The propInfo is already resolved by the caller (PropertyAccessor.LookupProperty or GetDefaultProperty)
	// Just call the executor with the property info, indices, and value
	return propertyExecutor(propInfo, indices, value)
}

// InvokeParameterlessMethod invokes a method if it has zero parameters.
// Returns (result, true) if successful, or (nil, false) if method has parameters.
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
// Returns (pointer, true) if successful, or (nil, false) if method has no parameters.
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

// GetClassConstantBySpec looks up a class constant by name in the class hierarchy.
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
