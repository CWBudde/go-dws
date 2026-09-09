package runtime

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// MutableInterfaceInfo is the concrete runtime-owned interface metadata used during declaration
// and execution. It implements IInterfaceInfo while remaining directly mutable by the engine.
type MutableInterfaceInfo struct {
	// Type is the immutable semantic identity retained during registration.
	Type       *types.InterfaceType
	Parent     *MutableInterfaceInfo
	Methods    map[string]*ast.FunctionDecl
	Properties map[string]*types.PropertyInfo
	Name       string
}

var _ IInterfaceInfo = (*MutableInterfaceInfo)(nil)

// NewMutableInterfaceInfo creates interface metadata with an independent type identity.
func NewMutableInterfaceInfo(name string) *MutableInterfaceInfo {
	return &MutableInterfaceInfo{
		Name:       name,
		Type:       types.NewInterfaceType(name),
		Methods:    make(map[string]*ast.FunctionDecl),
		Properties: make(map[string]*types.PropertyInfo),
	}
}

// GetInterfaceType returns the interface's resolved type identity.
func (ii *MutableInterfaceInfo) GetInterfaceType() *types.InterfaceType {
	return ii.Type
}

// GetName returns the declared interface name.
func (ii *MutableInterfaceInfo) GetName() string {
	return ii.Name
}

// GetParent returns the inherited interface, if any.
func (ii *MutableInterfaceInfo) GetParent() IInterfaceInfo {
	if ii.Parent == nil {
		return nil
	}
	return ii.Parent
}

// GetMethod finds a method in this interface or its ancestors.
func (ii *MutableInterfaceInfo) GetMethod(name string) any {
	normalizedName := ident.Normalize(name)
	if method, exists := ii.Methods[normalizedName]; exists {
		return method
	}
	if ii.Parent != nil {
		return ii.Parent.GetMethod(name)
	}
	return nil
}

// HasMethod reports whether the interface declares or inherits a method.
func (ii *MutableInterfaceInfo) HasMethod(name string) bool {
	return ii.GetMethod(name) != nil
}

// GetProperty finds a property in this interface or its ancestors.
func (ii *MutableInterfaceInfo) GetProperty(name string) *PropertyInfo {
	normalized := ident.Normalize(name)
	if prop, exists := ii.Properties[normalized]; exists {
		return &PropertyInfo{
			Name:      prop.Name,
			IsIndexed: prop.IsIndexed,
			IsDefault: prop.IsDefault,
			ReadSpec:  prop.ReadSpec,
			WriteSpec: prop.WriteSpec,
			Impl:      prop,
		}
	}
	if ii.Parent != nil {
		return ii.Parent.GetProperty(name)
	}
	return nil
}

// HasProperty reports whether the interface declares or inherits a property.
func (ii *MutableInterfaceInfo) HasProperty(name string) bool {
	return ii.GetProperty(name) != nil
}

// GetDefaultProperty returns the declared or inherited default property.
func (ii *MutableInterfaceInfo) GetDefaultProperty() *PropertyInfo {
	for _, prop := range ii.AllProperties() {
		if prop.IsDefault {
			return prop
		}
	}
	return nil
}

// AllMethods returns a copy of the method table including inherited methods.
func (ii *MutableInterfaceInfo) AllMethods() map[string]any {
	result := make(map[string]any)
	if ii.Parent != nil {
		for name, method := range ii.Parent.AllMethods() {
			result[name] = method
		}
	}
	for name, method := range ii.Methods {
		result[name] = method
	}
	return result
}

// AllProperties returns a copy of the property table including inherited properties.
func (ii *MutableInterfaceInfo) AllProperties() map[string]*PropertyInfo {
	result := make(map[string]*PropertyInfo)
	if ii.Parent != nil {
		for name, prop := range ii.Parent.AllProperties() {
			result[name] = prop
		}
	}
	for name, prop := range ii.Properties {
		result[name] = &PropertyInfo{
			Name:      prop.Name,
			IsIndexed: prop.IsIndexed,
			IsDefault: prop.IsDefault,
			ReadSpec:  prop.ReadSpec,
			WriteSpec: prop.WriteSpec,
			Impl:      prop,
		}
	}
	return result
}
