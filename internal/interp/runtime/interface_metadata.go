package runtime

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// MutableInterfaceInfo is the concrete runtime-owned interface metadata used during declaration
// and execution. It implements IInterfaceInfo while remaining directly mutable by the engine.
type MutableInterfaceInfo struct {
	Parent     *MutableInterfaceInfo
	Methods    map[string]*ast.FunctionDecl
	Properties map[string]*types.PropertyInfo
	Name       string
}

var _ IInterfaceInfo = (*MutableInterfaceInfo)(nil)

func NewMutableInterfaceInfo(name string) *MutableInterfaceInfo {
	return &MutableInterfaceInfo{
		Name:       name,
		Methods:    make(map[string]*ast.FunctionDecl),
		Properties: make(map[string]*types.PropertyInfo),
	}
}

func (ii *MutableInterfaceInfo) GetName() string {
	return ii.Name
}

func (ii *MutableInterfaceInfo) GetParent() IInterfaceInfo {
	if ii.Parent == nil {
		return nil
	}
	return ii.Parent
}

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

func (ii *MutableInterfaceInfo) HasMethod(name string) bool {
	return ii.GetMethod(name) != nil
}

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

func (ii *MutableInterfaceInfo) HasProperty(name string) bool {
	return ii.GetProperty(name) != nil
}

func (ii *MutableInterfaceInfo) GetDefaultProperty() *PropertyInfo {
	for _, prop := range ii.AllProperties() {
		if prop.IsDefault {
			return prop
		}
	}
	return nil
}

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
