package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

func (a *Analyzer) validateInterfacePropertyAccessors(decl *ast.PropertyDecl, prop *types.PropertyInfo, iface *types.InterfaceType) {
	methods := types.GetAllInterfaceMethods(iface)
	for _, write := range []bool{false, true} {
		accessor := decl.ReadSpec
		if write {
			accessor = decl.WriteSpec
		}
		id, ok := accessor.(*ast.Identifier)
		if !ok {
			continue
		}
		method := methods[ident.Normalize(id.Value)]
		// Parser recovery can omit an accessor declaration. Suppress only
		// that cascading missing-name error, retaining signature diagnostics
		// for declarations that survived unrelated syntax errors.
		if method == nil && a.parseHadErrors {
			continue
		}
		message := ""
		switch {
		case method == nil:
			message = `Field/method "%s" not found`
		case write && method.ReturnType != nil && !method.ReturnType.Equals(types.VOID):
			a.addError("Syntax Error: Procedure expected [line: %d, column: %d]", id.Pos().Line, id.Pos().Column)
			continue
		case !write && (method.ReturnType == nil || !method.ReturnType.Equals(prop.Type)):
			message = `Field/method "%s" has an incompatible type`
		case !interfacePropertyParametersMatch(method, prop, write):
			message = `Method "%s" has incompatible parameters`
		}

		if message != "" {
			a.addError("Syntax Error: "+message+" [line: %d, column: %d]", id.Value, id.Pos().Line, id.Pos().Column)
		}
	}
}

// interfaceIndexedProperty resolves the whole index chain without treating its
// property name as a read. Assignments use this path for write-only properties.
func (a *Analyzer) interfaceIndexedProperty(expr *ast.IndexExpression) (*types.PropertyInfo, []ast.Expression, bool) {
	root, indices := interfacePropertyIndexChain(expr)
	if a.isNonInterfaceIndexVariable(root) {
		return nil, nil, false
	}
	var prop *types.PropertyInfo
	if member, ok := root.(*ast.MemberAccessExpression); ok {
		receiver := a.analyzeExpression(member.Object)
		receiver = a.interfacePropertyReceiverType(member.Object, receiver)
		if iface, ok := types.GetUnderlyingType(receiver).(*types.InterfaceType); ok {
			prop = iface.GetProperty(member.Member.Value)
		} else if class, ok := types.GetUnderlyingType(receiver).(*types.ClassType); ok {
			// A class indexed property must retain the class analysis path;
			// its bare member would incorrectly report missing arguments here.
			if property, found := class.GetProperty(member.Member.Value); found && property.IsIndexed {
				return nil, nil, false
			}
		}
	}
	if prop == nil || !prop.IsIndexed {
		receiver := a.analyzeIndexBase(root)
		receiver = a.interfacePropertyReceiverType(root, receiver)
		if iface, ok := types.GetUnderlyingType(receiver).(*types.InterfaceType); ok {
			prop = iface.GetDefaultProperty()
		} else {
			return nil, nil, false
		}
	}
	if prop == nil || !prop.IsIndexed {
		return nil, nil, false
	}
	return prop, indices, true
}

func (a *Analyzer) analyzeInterfaceIndexedProperty(expr *ast.IndexExpression, write, compound bool) (types.Type, bool) {
	prop, indices, found := a.interfaceIndexedProperty(expr)
	if !found {
		return nil, false
	}
	if len(indices) > len(prop.IndexParamTypes) {
		return nil, false
	}
	if !a.checkInterfacePropertyAccess(prop, expr, write, compound) {
		return prop.Type, true
	}
	if len(indices) != len(prop.IndexParamTypes) {
		a.addStructuredError(NewPropertyDeclarationArgumentCountError(expr.Pos(), "property '"+prop.Name+"' expects "+formatInt(len(prop.IndexParamTypes))+" index arguments, got "+formatInt(len(indices))))
		return prop.Type, true
	}
	for n, index := range indices {
		got := a.analyzeExpressionWithExpectedType(index, prop.IndexParamTypes[n])
		if got != nil && !a.canAssign(got, prop.IndexParamTypes[n]) {
			a.addStructuredError(NewArrayIndexError(index.Pos(), prop.IndexParamTypes[n].String(), got.String()))
		}
	}
	return prop.Type, true
}

func (a *Analyzer) checkInterfacePropertyAccess(prop *types.PropertyInfo, node ast.Node, write, compound bool) bool {
	if write && prop.WriteKind == types.PropAccessNone {
		a.addStructuredError(NewReadOnlyPropertyError(node.Pos(), prop.Name))
		return false
	}
	if (!write || compound) && prop.ReadKind == types.PropAccessNone {
		a.addStructuredError(NewWriteOnlyPropertyError(node.Pos(), prop.Name))
		return false
	}
	return true
}

func interfacePropertyParametersMatch(method *types.FunctionType, prop *types.PropertyInfo, write bool) bool {
	expected := append([]types.Type{}, prop.IndexParamTypes...)
	if write {
		expected = append(expected, prop.Type)
	}
	if len(expected) != len(method.Parameters) {
		return false
	}
	for n, typ := range expected {
		if !typ.Equals(method.Parameters[n]) {
			return false
		}
	}
	return true
}

func (a *Analyzer) resolveInterfacePropertyIndices(prop *ast.PropertyDecl, propInfo *types.PropertyInfo) bool {
	for _, param := range prop.IndexParams {
		paramType, err := a.resolveTypeExpression(param.Type)
		if err != nil {
			a.addStructuredError(NewPropertyDeclarationError(param.Pos(), err.Error()))
			return false
		}
		propInfo.IndexParamTypes = append(propInfo.IndexParamTypes, paramType)
		propInfo.IndexParamNames = append(propInfo.IndexParamNames, param.Name.Value)
	}
	return true
}

func (a *Analyzer) validateInterfaceDefaultProperty(prop *ast.PropertyDecl, propInfo *types.PropertyInfo, iface *types.InterfaceType) {
	if prop.IsDefault {
		for _, existing := range iface.Properties {
			if existing.IsDefault {
				a.addError(`Syntax Error: "%s" already has "%s" as default property [line: %d, column: %d]`, iface.Name, existing.Name, prop.DefaultPos.Line, prop.DefaultPos.Column)
				propInfo.IsDefault = false
				break
			}
		}
	}
}

func interfacePropertyIndexChain(expr *ast.IndexExpression) (ast.Expression, []ast.Expression) {
	var indices []ast.Expression
	var root ast.Expression = expr
	for {
		idx, ok := root.(*ast.IndexExpression)
		if !ok {
			break
		}
		indices = append(indices, idx.Index)
		root = idx.Left
	}
	for i, j := 0, len(indices)-1; i < j; i, j = i+1, j-1 {
		indices[i], indices[j] = indices[j], indices[i]
	}
	return root, indices
}

func (a *Analyzer) isNonInterfaceIndexVariable(root ast.Expression) bool {
	// Ordinary array variables need no interface probe or repeated analysis.
	if id, ok := root.(*ast.Identifier); ok {
		if symbol, found := a.symbols.Resolve(id.Value); found {
			typ := types.GetUnderlyingType(symbol.Type)
			if result := implicitValueContextType(typ); result != nil {
				typ = types.GetUnderlyingType(result)
			}
			if _, ok := typ.(*types.InterfaceType); !ok {
				return true
			}
		}
	}
	return false
}

func (a *Analyzer) interfacePropertyReceiverType(expr ast.Expression, typ types.Type) types.Type {
	typ = a.applyImplicitCallType(expr, typ)
	if result := implicitValueContextType(typ); result != nil {
		return result
	}
	return typ
}
