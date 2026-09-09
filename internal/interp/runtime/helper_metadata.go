package runtime

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// MutableHelperInfo holds runtime declarations and state for a type helper.
type MutableHelperInfo struct {
	TargetType      types.Type
	ParentHelper    *MutableHelperInfo
	Methods         map[string]*ast.FunctionDecl
	MethodOverloads map[string][]*ast.FunctionDecl
	Properties      map[string]*types.PropertyInfo
	ClassVars       map[string]Value
	ClassConsts     map[string]Value
	BuiltinMethods  map[string]string
	Name            string
	IsRecordHelper  bool
	IsClassHelper   bool
	IsStrict        bool
}

// NewMutableHelperInfo creates empty helper metadata for the target type.
func NewMutableHelperInfo(name string, targetType types.Type, isRecordHelper bool) *MutableHelperInfo {
	return &MutableHelperInfo{
		Name:            name,
		TargetType:      targetType,
		Methods:         make(map[string]*ast.FunctionDecl),
		MethodOverloads: make(map[string][]*ast.FunctionDecl),
		Properties:      make(map[string]*types.PropertyInfo),
		ClassVars:       make(map[string]Value),
		ClassConsts:     make(map[string]Value),
		BuiltinMethods:  make(map[string]string),
		IsRecordHelper:  isRecordHelper,
	}
}

// GetName returns the declared helper name.
func (h *MutableHelperInfo) GetName() string {
	if h == nil {
		return ""
	}
	return h.Name
}

// GetTargetType returns the type extended by this helper.
func (h *MutableHelperInfo) GetTargetType() types.Type {
	if h == nil {
		return nil
	}
	return h.TargetType
}

// GetMethod finds a method and the helper that declares it.
func (h *MutableHelperInfo) GetMethod(name string) (*ast.FunctionDecl, *MutableHelperInfo, bool) {
	for key, method := range h.Methods {
		if ident.Equal(key, name) {
			return method, h, true
		}
	}
	for key, overloads := range h.MethodOverloads {
		if ident.Equal(key, name) && len(overloads) > 0 {
			return overloads[len(overloads)-1], h, true
		}
	}
	if h.ParentHelper != nil {
		return h.ParentHelper.GetMethod(name)
	}
	return nil, nil, false
}

// GetMethodOverloads finds method overloads and their declaring helper.
func (h *MutableHelperInfo) GetMethodOverloads(name string) ([]*ast.FunctionDecl, *MutableHelperInfo, bool) {
	for key, overloads := range h.MethodOverloads {
		if ident.Equal(key, name) && len(overloads) > 0 {
			return overloads, h, true
		}
	}
	if method, owner, found := h.GetMethod(name); found {
		return []*ast.FunctionDecl{method}, owner, true
	}
	return nil, nil, false
}

// GetBuiltinMethod finds a builtin method binding and its declaring helper.
func (h *MutableHelperInfo) GetBuiltinMethod(name string) (string, *MutableHelperInfo, bool) {
	for key, spec := range h.BuiltinMethods {
		if ident.Equal(key, name) {
			return spec, h, true
		}
	}
	if h.ParentHelper != nil {
		return h.ParentHelper.GetBuiltinMethod(name)
	}
	return "", nil, false
}

// GetProperty finds a property and the helper that declares it.
func (h *MutableHelperInfo) GetProperty(name string) (*types.PropertyInfo, *MutableHelperInfo, bool) {
	for key, prop := range h.Properties {
		if ident.Equal(key, name) {
			return prop, h, true
		}
	}
	if h.ParentHelper != nil {
		return h.ParentHelper.GetProperty(name)
	}
	return nil, nil, false
}

// GetClassVars returns the mutable class variable table.
func (h *MutableHelperInfo) GetClassVars() map[string]Value {
	return h.ClassVars
}

// GetClassConsts returns the class constant table.
func (h *MutableHelperInfo) GetClassConsts() map[string]Value {
	return h.ClassConsts
}

// GetParentHelper returns the inherited helper, if any.
func (h *MutableHelperInfo) GetParentHelper() *MutableHelperInfo {
	return h.ParentHelper
}
