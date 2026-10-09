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
		default:
			var value types.Type
			if write {
				value = prop.Type
			}
			if !a.checkPropertyAccessorParameters(decl, method, prop.IndexParamTypes, value, propertyAccessorDiagnosticPos(decl, write)) {
				message = `Method "%s" has incompatible parameters`
			}
		}

		if message != "" {
			a.addError("Syntax Error: "+message+" [line: %d, column: %d]", id.Value, id.Pos().Line, id.Pos().Column)
		}
	}
}

// interfaceIndexedProperty resolves the whole index chain without treating its
// property name as a read. Assignments use this path for write-only properties.
func (a *Analyzer) interfaceIndexedProperty(expr *ast.IndexExpression) (*types.PropertyInfo, *types.InterfaceType, []ast.Expression, bool) {
	root, indices := interfacePropertyIndexChain(expr)
	if a.isNonInterfaceIndexVariable(root) {
		return nil, nil, nil, false
	}
	var prop *types.PropertyInfo
	var contract *types.InterfaceType
	if member, ok := root.(*ast.MemberAccessExpression); ok {
		var classProperty bool
		prop, contract, classProperty = a.namedInterfaceIndexProperty(member)
		if classProperty {
			return nil, nil, nil, false
		}
	}

	if prop == nil || !prop.IsIndexed {
		receiver := a.analyzeIndexBase(root)
		receiver = a.interfacePropertyReceiverType(root, receiver)
		if iface, ok := types.GetUnderlyingType(receiver).(*types.InterfaceType); ok {
			prop = iface.GetDefaultProperty()
			contract = iface
		} else {
			return nil, nil, nil, false
		}
	}
	if prop == nil || !prop.IsIndexed {
		return nil, nil, nil, false
	}
	return prop, contract, indices, true
}

func (a *Analyzer) namedInterfaceIndexProperty(member *ast.MemberAccessExpression) (*types.PropertyInfo, *types.InterfaceType, bool) {
	if class, _ := propertyIndexClassReceiver(a.inferMemberObjectType(member.Object)); class != nil {
		if property, found := class.GetProperty(member.Member.Value); found && property.IsIndexed {
			return nil, nil, true
		}
	}
	receiver := a.analyzeExpression(member.Object)
	receiver = a.interfacePropertyReceiverType(member.Object, receiver)
	if iface, ok := types.GetUnderlyingType(receiver).(*types.InterfaceType); ok {
		return iface.GetProperty(member.Member.Value), iface, false
	} else if class, _ := propertyIndexClassReceiver(receiver); class != nil {
		// Dynamic receivers already read here are consumed by the named class path.
		if property, found := class.GetProperty(member.Member.Value); found && property.IsIndexed {
			if propertyHasVarIndices(property) {
				if a.probedReceivers == nil {
					a.probedReceivers = make(map[ast.Expression]types.Type)
				}
				a.probedReceivers[member.Object] = receiver
			}
			return nil, nil, true
		}
	}
	return nil, nil, false
}

func (a *Analyzer) analyzeInterfaceIndexedProperty(expr *ast.IndexExpression, write, compound bool, stmt *ast.AssignmentStatement) (types.Type, bool) {
	prop, contract, indices, found := a.interfaceIndexedProperty(expr)
	if !found {
		return nil, false
	}
	if len(indices) > len(prop.IndexParamTypes) {
		return nil, false
	}
	if propertyHasVarIndices(prop) {
		return a.analyzeVarInterfacePropertyArguments(expr, prop, contract, indices, write, compound, stmt), true
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

func (a *Analyzer) analyzeVarInterfacePropertyArguments(expr *ast.IndexExpression, prop *types.PropertyInfo, contract *types.InterfaceType, indices []ast.Expression, write, compound bool, stmt *ast.AssignmentStatement) types.Type {
	pos := expr.Token.Pos
	root, _ := interfacePropertyIndexChain(expr)
	if member, ok := root.(*ast.MemberAccessExpression); ok {
		pos = member.Member.Token.Pos
	}
	if !write && prop.ReadKind == types.PropAccessNone {
		a.addStructuredError(NewWriteOnlyPropertyError(pos, prop.Name))
		a.addStructuredError(NewGenericError(expr.Token.Pos, "Array expected"))
		return prop.Type
	}
	args, stopped := a.readPropertyIndexArguments(prop, indices, expr)
	if stopped {
		return nil
	}
	if !a.checkInterfacePropertyAccess(prop, expr, write, compound) {
		return nil
	}
	name := prop.ReadSpec
	if write {
		name = prop.WriteSpec
	}
	signature := types.GetAllInterfaceMethods(contract)[ident.Normalize(name)]
	if write && stmt != nil {
		value, stopped := a.readPropertyAssignmentValue(stmt, prop.Type)
		if stopped {
			return nil
		}
		a.checkPropertyWriteArguments(stmt, prop, args, value, signature, pos)
		return nil
	}
	a.checkPropertyReadArguments(prop, args, signature, pos)
	return prop.Type
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

func (a *Analyzer) resolveInterfacePropertyIndices(prop *ast.PropertyDecl, propInfo *types.PropertyInfo) bool {
	propInfo.IndexParamModes = propertyIndexParamModes(prop.IndexParams)
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
	// An empty member call can be compatibility punctuation. Preserve the
	// current bracket group so its inner property read is analyzed with its
	// own indices before probing a subsequent array/default-property index.
	emptyCall := false
	switch call := root.(type) {
	case *ast.MethodCallExpression:
		emptyCall = len(call.Arguments) == 0
	case *ast.CallExpression:
		emptyCall = len(call.Arguments) == 0
	case *ast.InheritedExpression:
		emptyCall = call.Method != nil && len(call.Arguments) == 0
	}
	if emptyCall {
		indices = []ast.Expression{expr.Index}
		first := expr
		for first.CommaPos.Line != 0 {
			inner, ok := first.Left.(*ast.IndexExpression)
			if !ok {
				break
			}
			indices = append(indices, inner.Index)
			first = inner
		}
		root = first.Left
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
