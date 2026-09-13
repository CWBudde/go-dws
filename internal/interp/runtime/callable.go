package runtime

import "github.com/cwbudde/go-dws/pkg/ast"

// MethodDeclaration returns the executable AST payload of a runtime callable.
func MethodDeclaration(method *MethodMetadata) *ast.FunctionDecl {
	if method == nil {
		return nil
	}
	return method.Declaration
}

// BindImplementation updates a callable in place, preserving references held by
// inherited dispatch entries and method pointers.
func (method *MethodMetadata) BindImplementation(decl *ast.FunctionDecl) {
	if method == nil || decl == nil {
		return
	}
	implementation := *decl
	implementation.Parameters = make([]*ast.Parameter, len(decl.Parameters))
	for index, parameter := range decl.Parameters {
		copyParameter := *parameter
		implementation.Parameters[index] = &copyParameter
	}
	if original := method.Declaration; original != nil {
		inheritCallableFlags(&implementation, original)
		implementation.Visibility = original.Visibility
		if implementation.PreConditions == nil {
			implementation.PreConditions = original.PreConditions
		}
		if implementation.PostConditions == nil {
			implementation.PostConditions = original.PostConditions
		}
		inheritParameterModifiers(implementation.Parameters, original.Parameters)
		if implementation.ClassName == nil {
			implementation.ClassName = original.ClassName
		}
	}
	mergeParameterDefaults(&implementation, method.Declaration)
	updated := MethodMetadataFromAST(&implementation)
	updated.ID = method.ID
	updated.Owner = method.Owner
	updated.UnitName = method.UnitName
	updated.SourceDeclaration = decl
	updated.ReturnType = method.ReturnType
	for index := range updated.Parameters {
		if index < len(method.Parameters) {
			updated.Parameters[index].Type = method.Parameters[index].Type
		}
	}
	*method = *updated
}

func inheritCallableFlags(implementation, original *ast.FunctionDecl) {
	implementation.IsVirtual = implementation.IsVirtual || original.IsVirtual
	implementation.IsOverride = implementation.IsOverride || original.IsOverride
	implementation.IsReintroduce = implementation.IsReintroduce || original.IsReintroduce
	implementation.IsClassMethod = implementation.IsClassMethod || original.IsClassMethod
	implementation.IsStatic = implementation.IsStatic || original.IsStatic
	implementation.IsConstructor = implementation.IsConstructor || original.IsConstructor
	implementation.IsDestructor = implementation.IsDestructor || original.IsDestructor
	implementation.IsOverload = implementation.IsOverload || original.IsOverload
	implementation.IsDefault = implementation.IsDefault || original.IsDefault
}

func inheritParameterModifiers(parameters, original []*ast.Parameter) {
	for index, parameter := range parameters {
		if index >= len(original) {
			break
		}
		parameter.ByRef = parameter.ByRef || original[index].ByRef
		parameter.IsLazy = parameter.IsLazy || original[index].IsLazy
		parameter.IsConst = parameter.IsConst || original[index].IsConst
	}
}
