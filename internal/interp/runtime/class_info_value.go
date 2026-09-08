package runtime

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ClassInfoValue tracks current class context in class methods.
// Stored as "__CurrentClass__" in the environment during method execution.
type ClassInfoValue struct {
	ClassInfo *ClassInfo
}

// Type returns the runtime tag for a current-class context value.
func (c *ClassInfoValue) Type() string   { return "CLASSINFO" }
func (c *ClassInfoValue) String() string { return "class " + c.ClassInfo.Name }

// GetClassName returns the class name.
func (c *ClassInfoValue) GetClassName() string {
	if c == nil || c.ClassInfo == nil {
		return ""
	}
	return c.ClassInfo.Name
}

// GetClassType returns the class type (metaclass) as a ClassValue.
func (c *ClassInfoValue) GetClassType() Value {
	if c == nil || c.ClassInfo == nil {
		return nil
	}
	return &ClassValue{ClassInfo: c.ClassInfo}
}

// GetClassVar retrieves a class variable by name from the hierarchy.
func (c *ClassInfoValue) GetClassVar(name string) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	value, owningClass := c.ClassInfo.lookupClassVar(name)
	return value, owningClass != nil
}

// GetClassConstant retrieves a class constant by name from the hierarchy.
func (c *ClassInfoValue) GetClassConstant(name string) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	// Check current class (case-insensitive)
	for constName, value := range c.ClassInfo.ConstantValues {
		if ident.Equal(constName, name) {
			return value, true
		}
	}
	// Check parent hierarchy
	if c.ClassInfo.Parent != nil {
		parentCIV := &ClassInfoValue{ClassInfo: c.ClassInfo.Parent}
		return parentCIV.GetClassConstant(name)
	}
	return nil, false
}

// HasClassMethod checks if a class method exists in the hierarchy.
func (c *ClassInfoValue) HasClassMethod(name string) bool {
	if c == nil || c.ClassInfo == nil {
		return false
	}
	normalizedName := ident.Normalize(name)
	// Check single methods and overloads
	if _, exists := c.ClassInfo.ClassMethods[normalizedName]; exists {
		return true
	}
	if overloads, exists := c.ClassInfo.ClassMethodOverloads[normalizedName]; exists && len(overloads) > 0 {
		return true
	}
	// Check parent
	if c.ClassInfo.Parent != nil {
		parentCIV := &ClassInfoValue{ClassInfo: c.ClassInfo.Parent}
		return parentCIV.HasClassMethod(name)
	}
	return false
}

// HasConstructor checks if a constructor exists.
func (c *ClassInfoValue) HasConstructor(name string) bool {
	if c == nil || c.ClassInfo == nil {
		return false
	}
	return c.ClassInfo.HasConstructor(name)
}

// InvokeParameterlessClassMethod invokes a parameterless class method.
func (c *ClassInfoValue) InvokeParameterlessClassMethod(name string, executor func(methodDecl any) Value) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	// Look up class method in hierarchy
	normalizedName := ident.Normalize(name)
	for current := c.ClassInfo; current != nil; current = current.Parent {
		if method, exists := current.ClassMethods[normalizedName]; exists {
			if len(method.Parameters) == 0 {
				return executor(method), true
			}
			return nil, false // Has parameters
		}
		if overloads, exists := current.ClassMethodOverloads[normalizedName]; exists && len(overloads) > 0 {
			// Check if any overload is parameterless
			for _, m := range overloads {
				if len(m.Parameters) == 0 {
					return executor(m), true
				}
			}
			return nil, false // No parameterless overload
		}
	}
	return nil, false
}

// CreateClassMethodPointer creates a function pointer for a class method with parameters.
func (c *ClassInfoValue) CreateClassMethodPointer(name string, creator func(methodDecl any) Value) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	// Look up class method in hierarchy
	normalizedName := ident.Normalize(name)
	for current := c.ClassInfo; current != nil; current = current.Parent {
		if method, exists := current.ClassMethods[normalizedName]; exists {
			if len(method.Parameters) > 0 {
				return creator(method), true
			}
			return nil, false // Parameterless
		}
		if overloads, exists := current.ClassMethodOverloads[normalizedName]; exists && len(overloads) > 0 {
			// Return pointer for first overload with parameters
			for _, m := range overloads {
				if len(m.Parameters) > 0 {
					return creator(m), true
				}
			}
			return nil, false // All parameterless
		}
	}
	return nil, false
}

// InvokeConstructor invokes a constructor.
func (c *ClassInfoValue) InvokeConstructor(name string, executor func(methodDecl any) Value) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	// Look up constructor in hierarchy
	for current := c.ClassInfo; current != nil; current = current.Parent {
		// Check ConstructorOverloads first
		for ctorName, overloads := range current.ConstructorOverloads {
			if ident.Equal(ctorName, name) && len(overloads) > 0 {
				return executor(overloads[0]), true
			}
		}
		// Check single constructors
		for ctorName, ctor := range current.Constructors {
			if ident.Equal(ctorName, name) {
				return executor(ctor), true
			}
		}
	}
	return nil, false
}

// GetNestedClass returns a nested class by name.
func (c *ClassInfoValue) GetNestedClass(name string) Value {
	if c == nil || c.ClassInfo == nil {
		return nil
	}
	nested := c.ClassInfo.LookupNestedClass(name)
	if nested == nil {
		return nil
	}
	return &ClassInfoValue{ClassInfo: nested}
}

// ReadClassProperty reads a class property value using the executor callback.
func (c *ClassInfoValue) ReadClassProperty(name string, executor func(propInfo any) Value) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	// Look up property in hierarchy
	propInfo := c.ClassInfo.lookupProperty(name)
	if propInfo == nil || !propInfo.IsClassProperty {
		return nil, false
	}
	return executor(propInfo), true
}

// GetClassInfo returns the underlying ClassInfo.
func (c *ClassInfoValue) GetClassInfo() IClassInfo {
	if c == nil {
		return nil
	}
	return c.ClassInfo
}

// SetClassVar sets a class variable by name in the hierarchy.
func (c *ClassInfoValue) SetClassVar(name string, value Value) bool {
	if c == nil || c.ClassInfo == nil {
		return false
	}
	return c.ClassInfo.setClassVar(name, value)
}

// HasClassVar checks if a class variable exists in the hierarchy.
func (c *ClassInfoValue) HasClassVar(name string) bool {
	if c == nil || c.ClassInfo == nil {
		return false
	}
	return c.ClassInfo.hasClassVar(name)
}

// WriteClassProperty writes to a class property using the executor callback.
func (c *ClassInfoValue) WriteClassProperty(name string, value Value, executor func(propInfo any, value Value) Value) (Value, bool) {
	if c == nil || c.ClassInfo == nil {
		return nil, false
	}
	// Look up class property in hierarchy
	propDesc := c.ClassInfo.LookupProperty(name)
	if propDesc == nil {
		return nil, false
	}
	propInfo, ok := propDesc.Impl.(*types.PropertyInfo)
	if !ok || !propInfo.IsClassProperty {
		return nil, false
	}
	// Check if property is writable
	if propInfo.WriteKind == types.PropAccessNone {
		return nil, false
	}
	// Execute the write via callback
	result := executor(propInfo, value)
	return result, true
}
