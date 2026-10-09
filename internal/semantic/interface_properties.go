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

// interfaceIndexedProperty resolves one bracket group without treating its
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
			if prop == nil || !prop.IsIndexed {
				if a.probedReceivers == nil {
					a.probedReceivers = make(map[ast.Expression]types.Type)
				}
				a.probedReceivers[root] = receiver
			}
		} else {
			// The ordinary index path must reuse this declined probe, including
			// its nil result, rather than reading the receiver a second time.
			if a.probedReceivers == nil {
				a.probedReceivers = make(map[ast.Expression]types.Type)
			}
			a.probedReceivers[root] = receiver
			return nil, nil, nil, false
		}
	}
	if prop == nil || !prop.IsIndexed {
		return nil, nil, nil, false
	}
	return prop, contract, indices, true
}

func (a *Analyzer) namedInterfaceIndexProperty(member *ast.MemberAccessExpression) (*types.PropertyInfo, *types.InterfaceType, bool) {
	inferred := a.inferMemberObjectType(member.Object)
	if class, _ := propertyIndexClassReceiver(inferred); class != nil {
		if property, found := class.GetProperty(member.Member.Value); found && property.IsIndexed {
			return nil, nil, true
		}
	}
	if inferred != nil {
		if _, iface := types.GetUnderlyingType(a.interfacePropertyReceiverType(member.Object, inferred)).(*types.InterfaceType); !iface {
			return nil, nil, false
		}
	}
	analyzed := a.analyzeProbedReceiver(member.Object)
	receiver := a.interfacePropertyReceiverType(member.Object, analyzed)
	if iface, ok := types.GetUnderlyingType(receiver).(*types.InterfaceType); ok {
		if prop := iface.GetProperty(member.Member.Value); prop != nil && prop.IsIndexed {
			return prop, iface, false
		}
	} else if class, _ := propertyIndexClassReceiver(receiver); class != nil {
		// Dynamic receivers already read here are consumed by the named class path.
		if property, found := class.GetProperty(member.Member.Value); found && property.IsIndexed {
			if a.probedReceivers == nil {
				a.probedReceivers = make(map[ast.Expression]types.Type)
			}
			a.probedReceivers[member.Object] = analyzed
			return nil, nil, true
		}
	}
	if a.probedReceivers == nil {
		a.probedReceivers = make(map[ast.Expression]types.Type)
	}
	a.probedReceivers[member.Object] = analyzed
	return nil, nil, false
}

func (a *Analyzer) analyzeInterfaceIndexedProperty(expr *ast.IndexExpression, write, compound bool, stmt *ast.AssignmentStatement) (types.Type, bool) {
	mark := len(a.structuredErrors)
	prop, contract, indices, found := a.interfaceIndexedProperty(expr)
	if !found {
		if a.propertyArgumentsStopped(mark) {
			return nil, true
		}
		root, _ := interfacePropertyIndexChain(expr)
		if receiver, analyzed := a.probedReceivers[root]; analyzed && receiver == nil {
			// A failed receiver read has no type for another index owner to
			// probe. Keep its original diagnostics and leave operands unread.
			delete(a.probedReceivers, root)
			return nil, true
		}
		return nil, false
	}
	a.semanticInfo.MarkResolvedIndexedProperty(expr)
	if propertyHasVarIndices(prop) || len(indices) != len(prop.IndexParamTypes) {
		return a.analyzeInterfacePropertyArguments(expr, prop, contract, indices, write, compound, stmt), true
	}

	if !a.checkInterfacePropertyAccess(prop, expr, write, compound) {
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

func (a *Analyzer) analyzeInterfacePropertyArguments(expr *ast.IndexExpression, prop *types.PropertyInfo, contract *types.InterfaceType, indices []ast.Expression, write, compound bool, stmt *ast.AssignmentStatement) types.Type {
	root, _ := interfacePropertyIndexChain(expr)
	first := expr
	for first.CommaPos.IsValid() {
		inner, ok := first.Left.(*ast.IndexExpression)
		if !ok {
			break
		}
		first = inner
	}
	pos := first.Token.Pos
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
	nodes := []*ast.IndexExpression{expr}
	root := expr.Left
	for nodes[len(nodes)-1].CommaPos.IsValid() {
		inner, ok := root.(*ast.IndexExpression)
		if !ok {
			break
		}
		nodes = append(nodes, inner)
		root = inner.Left
	}
	return root, indexedPropertyArguments(nodes)
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
