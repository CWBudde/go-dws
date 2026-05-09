package runtime

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

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

func (h *MutableHelperInfo) GetName() string {
	if h == nil {
		return ""
	}
	return h.Name
}

func (h *MutableHelperInfo) GetTargetType() types.Type {
	if h == nil {
		return nil
	}
	return h.TargetType
}

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

func (h *MutableHelperInfo) GetClassVars() map[string]Value {
	return h.ClassVars
}

func (h *MutableHelperInfo) GetClassConsts() map[string]Value {
	return h.ClassConsts
}

func (h *MutableHelperInfo) GetParentHelper() *MutableHelperInfo {
	return h.ParentHelper
}

func (h *MutableHelperInfo) GetMethodAny(name string) (*ast.FunctionDecl, any, bool) {
	method, owner, found := h.GetMethod(name)
	return method, owner, found
}

func (h *MutableHelperInfo) GetMethodOverloadsAny(name string) ([]*ast.FunctionDecl, any, bool) {
	methods, owner, found := h.GetMethodOverloads(name)
	return methods, owner, found
}

func (h *MutableHelperInfo) GetBuiltinMethodAny(name string) (string, any, bool) {
	spec, owner, found := h.GetBuiltinMethod(name)
	return spec, owner, found
}

func (h *MutableHelperInfo) GetPropertyAny(name string) (any, any, bool) {
	prop, owner, found := h.GetProperty(name)
	return prop, owner, found
}

func (h *MutableHelperInfo) GetParentHelperAny() any {
	return h.ParentHelper
}
