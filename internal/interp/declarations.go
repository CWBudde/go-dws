package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// evalClassMethodImplementation registers a class method implementation, replacing any declaration.
func (i *Interpreter) evalClassMethodImplementation(fn *ast.FunctionDecl, classInfo *ClassInfo) {
	normalizedMethodName := ident.Normalize(fn.Name.Value)

	// Replace declaration with implementation in method maps and overload lists
	if fn.IsClassMethod {
		classInfo.ClassMethods[normalizedMethodName] = fn
		overloads := classInfo.ClassMethodOverloads[normalizedMethodName]
		classInfo.ClassMethodOverloads[normalizedMethodName] = i.replaceMethodInOverloadList(overloads, fn)
	} else {
		classInfo.Methods[normalizedMethodName] = fn
		overloads := classInfo.MethodOverloads[normalizedMethodName]
		classInfo.MethodOverloads[normalizedMethodName] = i.replaceMethodInOverloadList(overloads, fn)
	}

	// Store constructors and destructors
	if fn.IsConstructor {
		normalizedCtorName := ident.Normalize(fn.Name.Value)
		classInfo.Constructors[normalizedCtorName] = fn
		overloads := classInfo.ConstructorOverloads[normalizedCtorName]
		classInfo.ConstructorOverloads[normalizedCtorName] = i.replaceMethodInOverloadList(overloads, fn)
		classInfo.Constructor = fn
		i.propagateConstructorImplementationToDescendants(classInfo, fn)
	}

	if fn.IsDestructor {
		classInfo.Destructor = fn
	}

	// Rebuild VMT and propagate to descendants
	classInfo.buildVirtualMethodTable()
	i.propagateMethodImplementationToDescendants(classInfo, normalizedMethodName, fn, fn.IsClassMethod)
	i.rebuildDescendantVMTs(classInfo)
}

// rebuildDescendantVMTs rebuilds VMTs for all descendant classes to pick up parent method changes.
func (i *Interpreter) rebuildDescendantVMTs(parentClass *ClassInfo) {
	for _, classInfo := range i.allRegisteredClassInfos() {
		if i.isDescendantOf(classInfo, parentClass) {
			classInfo.buildVirtualMethodTable()
		}
	}
}

// isDescendantOf returns true if childClass inherits from ancestorClass (directly or indirectly).
func (i *Interpreter) isDescendantOf(childClass, ancestorClass *ClassInfo) bool {
	current := childClass.Parent
	for current != nil {
		if current == ancestorClass {
			return true
		}
		current = current.Parent
	}
	return false
}

// propagateMethodImplementationToDescendants updates descendant method maps to latest parent implementation
// (unless the descendant provides its own override).
func (i *Interpreter) propagateMethodImplementationToDescendants(parentClass *ClassInfo, normalizedMethodName string, fn *ast.FunctionDecl, isClassMethod bool) {
	for _, classInfo := range i.allRegisteredClassInfos() {
		if !i.isDescendantOf(classInfo, parentClass) {
			continue
		}

		if isClassMethod {
			if existing, ok := classInfo.ClassMethods[normalizedMethodName]; ok {
				// Skip if descendant overrides the method
				if existing.ClassName != nil && ident.Equal(existing.ClassName.Value, classInfo.Name) {
					continue
				}
				classInfo.ClassMethods[normalizedMethodName] = fn
			}
		} else {
			if existing, ok := classInfo.Methods[normalizedMethodName]; ok {
				// Skip if descendant overrides the method
				if existing.ClassName != nil && ident.Equal(existing.ClassName.Value, classInfo.Name) {
					continue
				}
				classInfo.Methods[normalizedMethodName] = fn
			}
		}
	}
}

// propagateConstructorImplementationToDescendants updates inherited constructor overloads
// when the base class implementation becomes available.
func (i *Interpreter) propagateConstructorImplementationToDescendants(parentClass *ClassInfo, fn *ast.FunctionDecl) {
	normalizedCtorName := ident.Normalize(fn.Name.Value)

	for _, classInfo := range i.allRegisteredClassInfos() {
		if !i.isDescendantOf(classInfo, parentClass) {
			continue
		}

		// Update Constructors map if it still references the parent's declaration.
		if ctor, ok := classInfo.Constructors[normalizedCtorName]; ok && ctor != nil {
			if ctor.ClassName != nil && ident.Equal(ctor.ClassName.Value, classInfo.Name) {
				continue
			}
			if parametersMatch(ctor.Parameters, fn.Parameters) {
				classInfo.Constructors[normalizedCtorName] = fn
			}
		}

		// Update overload lists, but only for inherited constructor entries.
		if overloads, ok := classInfo.ConstructorOverloads[normalizedCtorName]; ok {
			for idx, decl := range overloads {
				if decl == nil {
					continue
				}
				if decl.ClassName != nil && ident.Equal(decl.ClassName.Value, classInfo.Name) {
					continue
				}
				if parametersMatch(decl.Parameters, fn.Parameters) {
					overloads[idx] = fn
				}
			}
			classInfo.ConstructorOverloads[normalizedCtorName] = overloads
		}
	}
}

// evalRecordMethodImplementation registers a record method implementation, replacing any declaration.
func (i *Interpreter) evalRecordMethodImplementation(fn *ast.FunctionDecl, recordInfo *RecordTypeValue) {
	normalizedMethodName := ident.Normalize(fn.Name.Value)
	methodMeta := runtime.MethodMetadataFromAST(fn)

	if fn.IsClassMethod {
		// Static method
		recordInfo.ClassMethods[normalizedMethodName] = fn
		overloads := recordInfo.ClassMethodOverloads[normalizedMethodName]
		recordInfo.ClassMethodOverloads[normalizedMethodName] = i.replaceMethodInOverloadList(overloads, fn)

		// Update metadata
		if recordInfo.Metadata != nil {
			recordInfo.Metadata.StaticMethods[normalizedMethodName] = methodMeta
			recordInfo.Metadata.StaticMethodOverloads[normalizedMethodName] = i.replaceMethodMetadataInOverloadList(
				recordInfo.Metadata.StaticMethodOverloads[normalizedMethodName],
				methodMeta,
			)
		}
	} else {
		// Instance method
		recordInfo.Methods[normalizedMethodName] = fn
		overloads := recordInfo.MethodOverloads[normalizedMethodName]
		recordInfo.MethodOverloads[normalizedMethodName] = i.replaceMethodInOverloadList(overloads, fn)

		// Update metadata
		if recordInfo.Metadata != nil {
			recordInfo.Metadata.Methods[normalizedMethodName] = methodMeta
			recordInfo.Metadata.MethodOverloads[normalizedMethodName] = i.replaceMethodMetadataInOverloadList(
				recordInfo.Metadata.MethodOverloads[normalizedMethodName],
				methodMeta,
			)
		}
	}
}

// convertPropertyDecl converts AST property to PropertyInfo for runtime access.
// classInfo is used to determine if a property spec is a field or method.
func (i *Interpreter) convertPropertyDecl(classInfo *ClassInfo, propDecl *ast.PropertyDecl) *types.PropertyInfo {
	// Resolve property type
	var propType types.Type
	switch propDecl.Type.String() {
	case "Integer":
		propType = types.INTEGER
	case "Float":
		propType = types.FLOAT
	case "String":
		propType = types.STRING
	case "Boolean":
		propType = types.BOOLEAN
	default:
		if classInfo := i.resolveClassInfoByName(propDecl.Type.String()); classInfo != nil {
			propType = types.NewClassType(classInfo.Name, nil)
		} else {
			propType = types.NIL
		}
	}

	propInfo := &types.PropertyInfo{
		Name:            propDecl.Name.Value,
		Type:            propType,
		IsIndexed:       len(propDecl.IndexParams) > 0,
		IsDefault:       propDecl.IsDefault,
		IsClassProperty: propDecl.IsClassProperty,
	}

	if propDecl.IndexValue != nil {
		if val, ok := ast.ExtractIntegerLiteral(propDecl.IndexValue); ok {
			propInfo.HasIndexValue = true
			propInfo.IndexValue = val
			propInfo.IndexValueType = types.INTEGER
		}
	}

	// Configure read access
	if propDecl.ReadSpec != nil {
		if ident, ok := propDecl.ReadSpec.(*ast.Identifier); ok {
			propInfo.ReadSpec = ident.Value
			propInfo.ReadKind = types.PropAccessField
		} else {
			propInfo.ReadKind = types.PropAccessExpression
			propInfo.ReadSpec = propDecl.ReadSpec.String()
			propInfo.ReadExpr = propDecl.ReadSpec
		}
	} else {
		propInfo.ReadKind = types.PropAccessNone
	}

	// Configure write access
	if propDecl.WriteSpec != nil {
		if ident, ok := propDecl.WriteSpec.(*ast.Identifier); ok {
			propInfo.WriteSpec = ident.Value
			propInfo.WriteKind = i.determinePropertyAccessKind(classInfo, ident.Value)
		} else {
			propInfo.WriteKind = types.PropAccessNone
		}
	} else {
		propInfo.WriteKind = types.PropAccessNone
	}

	return propInfo
}

// determinePropertyAccessKind determines whether a property spec name refers to a field, class var, or method.
func (i *Interpreter) determinePropertyAccessKind(classInfo *ClassInfo, specName string) types.PropAccessKind {
	if classInfo == nil {
		// For interfaces or when no class context, treat as method (conservative default)
		return types.PropAccessMethod
	}

	normalizedName := ident.Normalize(specName)

	// Check if it's a field in this class or any parent
	// Fields may be stored with original case or normalized key (inconsistency in codebase)
	current := classInfo
	for current != nil {
		// Try normalized key first
		if _, isField := current.Fields[normalizedName]; isField {
			return types.PropAccessField
		}
		// Try original case (fields are sometimes stored with original case)
		if _, isField := current.Fields[specName]; isField {
			return types.PropAccessField
		}
		// Check for class variables (used in class properties)
		if _, isClassVar := current.ClassVars[specName]; isClassVar {
			return types.PropAccessField
		}
		current = current.Parent
	}

	// Check if it's a method (including class methods) in this class or any parent
	current = classInfo
	for current != nil {
		if _, isMethod := current.Methods[normalizedName]; isMethod {
			return types.PropAccessMethod
		}
		if _, isClassMethod := current.ClassMethods[normalizedName]; isClassMethod {
			return types.PropAccessMethod
		}
		current = current.Parent
	}

	// Neither field nor method found - return None (error will be caught elsewhere)
	return types.PropAccessNone
}

// parametersMatch checks if two parameter lists have matching signatures.
func parametersMatch(params1, params2 []*ast.Parameter) bool {
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

// replaceMethodInOverloadList replaces a declaration with its implementation, or appends if not found.
func (i *Interpreter) replaceMethodInOverloadList(list []*ast.FunctionDecl, impl *ast.FunctionDecl) []*ast.FunctionDecl {
	for idx, decl := range list {
		if parametersMatch(decl.Parameters, impl.Parameters) {
			// Preserve flags from declaration
			impl.IsVirtual = decl.IsVirtual
			impl.IsOverride = decl.IsOverride
			impl.IsReintroduce = decl.IsReintroduce
			impl.IsAbstract = decl.IsAbstract

			list[idx] = impl
			return list
		}
	}
	return append(list, impl)
}

// replaceMethodMetadataInOverloadList replaces a declaration with its implementation.
func (i *Interpreter) replaceMethodMetadataInOverloadList(list []*runtime.MethodMetadata, impl *runtime.MethodMetadata) []*runtime.MethodMetadata {
	for idx, decl := range list {
		if methodMetadataSignatureMatch(decl, impl) {
			// Preserve flags from declaration
			impl.IsVirtual = decl.IsVirtual
			impl.IsOverride = decl.IsOverride
			impl.IsReintroduce = decl.IsReintroduce
			impl.IsAbstract = decl.IsAbstract
			impl.Visibility = decl.Visibility
			if impl.ReturnTypeName == "" {
				impl.ReturnTypeName = decl.ReturnTypeName
			}
			list[idx] = impl
			return list
		}
	}
	return append(list, impl)
}

// methodMetadataSignatureMatch checks if two MethodMetadata have matching signatures.
func methodMetadataSignatureMatch(a, b *runtime.MethodMetadata) bool {
	if a == nil || b == nil {
		return false
	}

	if !parameterMetadataMatch(a.Parameters, b.Parameters) {
		return false
	}

	if a.ReturnTypeName != "" && b.ReturnTypeName != "" && !ident.Equal(a.ReturnTypeName, b.ReturnTypeName) {
		return false
	}

	return true
}

// parameterMetadataMatch checks if two parameter lists match by type and ByRef flag.
func parameterMetadataMatch(params1, params2 []runtime.ParameterMetadata) bool {
	if len(params1) != len(params2) {
		return false
	}

	for i := range params1 {
		if params1[i].ByRef != params2[i].ByRef {
			return false
		}

		switch {
		case params1[i].TypeName != "" && params2[i].TypeName != "":
			if !ident.Equal(params1[i].TypeName, params2[i].TypeName) {
				return false
			}
		case params1[i].TypeName != params2[i].TypeName:
			return false
		}
	}

	return true
}
