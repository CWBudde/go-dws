package runtime

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

type MutableHelperInfo struct {
	TargetType     types.Type
	ParentHelper   *MutableHelperInfo
	Methods        map[string]*ast.FunctionDecl
	Properties     map[string]*types.PropertyInfo
	ClassVars      map[string]Value
	ClassConsts    map[string]Value
	BuiltinMethods map[string]string
	Name           string
	IsRecordHelper bool
}

func NewMutableHelperInfo(name string, targetType types.Type, isRecordHelper bool) *MutableHelperInfo {
	return &MutableHelperInfo{
		Name:           name,
		TargetType:     targetType,
		Methods:        make(map[string]*ast.FunctionDecl),
		Properties:     make(map[string]*types.PropertyInfo),
		ClassVars:      make(map[string]Value),
		ClassConsts:    make(map[string]Value),
		BuiltinMethods: make(map[string]string),
		IsRecordHelper: isRecordHelper,
	}
}

func (h *MutableHelperInfo) GetMethod(name string) (*ast.FunctionDecl, *MutableHelperInfo, bool) {
	for key, method := range h.Methods {
		if ident.Equal(key, name) {
			return method, h, true
		}
	}
	if h.ParentHelper != nil {
		return h.ParentHelper.GetMethod(name)
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
