package runtime

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

func (r *RecordTypeValue) RegisterMethodImplementation(fn *ast.FunctionDecl) {
	if r == nil || fn == nil {
		return
	}

	normalizedMethodName := ident.Normalize(fn.Name.Value)
	methodMeta := MethodMetadataFromAST(fn)

	if fn.IsClassMethod {
		r.ClassMethods[normalizedMethodName] = fn
		overloads := r.ClassMethodOverloads[normalizedMethodName]
		r.ClassMethodOverloads[normalizedMethodName] = replaceMethodOverloadList(overloads, fn)

		if r.Metadata != nil {
			r.Metadata.StaticMethods[normalizedMethodName] = methodMeta
			r.Metadata.StaticMethodOverloads[normalizedMethodName] = replaceMethodMetadataOverloadList(
				r.Metadata.StaticMethodOverloads[normalizedMethodName],
				methodMeta,
			)
		}
		return
	}

	r.Methods[normalizedMethodName] = fn
	overloads := r.MethodOverloads[normalizedMethodName]
	r.MethodOverloads[normalizedMethodName] = replaceMethodOverloadList(overloads, fn)

	if r.Metadata != nil {
		r.Metadata.Methods[normalizedMethodName] = methodMeta
		r.Metadata.MethodOverloads[normalizedMethodName] = replaceMethodMetadataOverloadList(
			r.Metadata.MethodOverloads[normalizedMethodName],
			methodMeta,
		)
	}
}

func replaceMethodOverloadList(list []*ast.FunctionDecl, impl *ast.FunctionDecl) []*ast.FunctionDecl {
	for idx, decl := range list {
		if parametersMatchAST(decl.Parameters, impl.Parameters) {
			list[idx] = impl
			return list
		}
	}
	return append(list, impl)
}

func replaceMethodMetadataOverloadList(list []*MethodMetadata, impl *MethodMetadata) []*MethodMetadata {
	if impl == nil {
		return list
	}
	for idx, decl := range list {
		if decl == nil {
			continue
		}
		if methodParametersMatchMetadata(decl, impl) {
			list[idx] = impl
			return list
		}
	}
	return append(list, impl)
}

func parametersMatchAST(params1, params2 []*ast.Parameter) bool {
	if len(params1) != len(params2) {
		return false
	}
	for i := range params1 {
		if params1[i].Type != nil && params2[i].Type != nil {
			if params1[i].Type.String() != params2[i].Type.String() {
				return false
			}
		} else if params1[i].Type != params2[i].Type {
			return false
		}
	}
	return true
}

func methodParametersMatchMetadata(first, second *MethodMetadata) bool {
	params1, params2 := first.Parameters, second.Parameters
	if len(params1) != len(params2) {
		return false
	}
	for i := range params1 {
		// Implementation binding can precede resolution of the new declaration.
		// Retain its original syntax until declaration registration supplies types.
		if params1[i].Type == nil || params2[i].Type == nil {
			if first.Declaration != nil && second.Declaration != nil {
				return parametersMatchAST(first.Declaration.Parameters, second.Declaration.Parameters)
			}
			return false
		}
		if !types.OperatorTypesEqual(params1[i].Type, params2[i].Type) {
			return false
		}
	}
	return true
}
