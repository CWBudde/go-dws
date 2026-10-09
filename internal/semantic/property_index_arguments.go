package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// propertyIndexArgument records the expression actually reached by ReadTerm,
// separately from the selected accessor's later argument validation.
type propertyIndexArgument struct {
	expression ast.Expression
	typ        types.Type
}

func propertyHasVarIndices(prop *types.PropertyInfo) bool {
	for i := range prop.IndexParamTypes {
		if prop.IndexMode(i) == types.PropertyIndexVar {
			return true
		}
	}
	return false
}

// propertyTermPrefix projects only the surface left spine. Grouping, call
// arguments and nested index arguments are full expression readers of their own.
func propertyTermPrefix(expr ast.Expression) (ast.Expression, token.Position) {
	if left, boundary := propertySurfaceOperator(expr); left != nil {
		if reached, earlier := propertyTermPrefix(left); earlier.IsValid() {
			return reached, earlier
		}
		return left, boundary
	}
	if unary, ok := expr.(*ast.UnaryExpression); ok {
		return propertyUnaryTermPrefix(unary)
	}
	var base ast.Expression
	switch e := expr.(type) {
	case *ast.MemberAccessExpression:
		base = e.Object
	case *ast.IndexExpression:
		base = e.Left
	case *ast.CallExpression:
		base = e.Function
	case *ast.MethodCallExpression:
		base = e.Object
	}
	if base != nil {
		if reached, stop := propertyTermPrefix(base); stop.IsValid() {
			return reached, stop
		}
	}
	return expr, token.Position{}
}

func propertySurfaceOperator(expr ast.Expression) (ast.Expression, token.Position) {
	switch e := expr.(type) {
	case *ast.BinaryExpression:
		return e.Left, e.Token.Pos
	case *ast.IsExpression:
		return e.Left, e.Token.Pos
	case *ast.AsExpression:
		return e.Left, e.Token.Pos
	case *ast.ImplementsExpression:
		return e.Left, e.Token.Pos
	}
	return nil, token.Position{}
}

func propertyUnaryTermPrefix(expr *ast.UnaryExpression) (ast.Expression, token.Position) {
	reached, boundary := propertyTermPrefix(expr.Right)
	// The parser also wraps infix 'not in/is/as' in UnaryExpression.
	if propertyPositionBefore(propertySurfaceStart(expr.Right), expr.Token.Pos) {
		if boundary.IsValid() && propertyPositionBefore(boundary, expr.Token.Pos) {
			return reached, boundary
		}
		return reached, expr.Token.Pos
	}
	if boundary.IsValid() {
		projected := *expr
		projected.Right = reached
		return &projected, boundary
	}
	return expr, token.Position{}
}

func propertySurfaceStart(expr ast.Expression) token.Position {
	switch e := expr.(type) {
	case *ast.BinaryExpression:
		return propertySurfaceStart(e.Left)
	case *ast.IsExpression:
		return propertySurfaceStart(e.Left)
	case *ast.AsExpression:
		return propertySurfaceStart(e.Left)
	case *ast.ImplementsExpression:
		return propertySurfaceStart(e.Left)
	case *ast.MemberAccessExpression:
		return propertySurfaceStart(e.Object)
	case *ast.IndexExpression:
		return propertySurfaceStart(e.Left)
	case *ast.CallExpression:
		return propertySurfaceStart(e.Function)
	case *ast.MethodCallExpression:
		return propertySurfaceStart(e.Object)
	}
	if expr == nil {
		return token.Position{}
	}
	return expr.Pos()
}

func propertyPositionBefore(left, right token.Position) bool {
	if !left.IsValid() || !right.IsValid() {
		return false
	}
	if left.Offset != 0 && right.Offset != 0 {
		return left.Offset < right.Offset
	}
	return left.Line < right.Line || left.Line == right.Line && left.Column < right.Column
}

func (a *Analyzer) propertyArgumentsStopped(mark int) bool {
	for _, diagnostic := range a.structuredErrors[mark:] {
		if diagnostic.Stop {
			return true
		}
	}
	return false
}

func (a *Analyzer) readPropertyIndexArguments(prop *types.PropertyInfo, indices []ast.Expression, list *ast.IndexExpression) ([]propertyIndexArgument, bool) {
	args := make([]propertyIndexArgument, 0, len(indices))
	for i, index := range indices {
		reached, boundary := index, token.Position{}
		if prop.IndexMode(i) == types.PropertyIndexVar {
			reached, boundary = propertyTermPrefix(index)
		}
		mark := len(a.structuredErrors)
		previous := a.propertyTermMark
		if propertyHasVarIndices(prop) {
			a.propertyTermMark = &mark
		}
		var typ types.Type
		if i < len(prop.IndexParamTypes) {
			typ = a.analyzeArgumentForParameter(reached, prop.IndexParamTypes[i], false)
		} else {
			typ = a.analyzeExpression(reached)
		}
		a.propertyTermMark = previous
		if a.propertyArgumentsStopped(mark) {
			return args, true
		}
		args = append(args, propertyIndexArgument{expression: reached, typ: typ})
		if boundary.IsValid() {
			a.addPunctuationStop(boundary, `")" expected`)
			return args, true
		}
	}
	if list != nil && list.MissingClosePos.IsValid() {
		a.addPunctuationStop(list.MissingClosePos, `")" expected`)
		return args, true
	}
	return args, false
}

func (a *Analyzer) checkPropertyReadArguments(prop *types.PropertyInfo, args []propertyIndexArgument, signature *types.FunctionType, pos token.Position) {
	if prop.ReadKind == types.PropAccessField {
		return
	}
	a.checkPropertyAccessorArguments(propertyAccessorArguments(prop, args), propertyAccessorSignature(prop, signature, false), pos)
}

// Selected method parameters own checking. Expression accessors use the
// declaration signature, including the setter value when present.
func propertyAccessorSignature(prop *types.PropertyInfo, selected *types.FunctionType, write bool) *types.FunctionType {
	if selected != nil {
		return selected
	}
	parameters := append([]types.Type(nil), prop.IndexParamTypes...)
	if prop.HasIndexValue {
		parameters = append(parameters, prop.IndexValueType)
	}
	if write {
		parameters = append(parameters, prop.Type)
	}
	signature := &types.FunctionType{Parameters: parameters, VarParams: make([]bool, len(parameters)), ParamNames: make([]string, len(parameters))}
	for i := range prop.IndexParamTypes {
		signature.VarParams[i] = prop.IndexMode(i) == types.PropertyIndexVar
		if i < len(prop.IndexParamNames) {
			signature.ParamNames[i] = prop.IndexParamNames[i]
		}
	}
	return signature
}

func propertyAccessorArguments(prop *types.PropertyInfo, indices []propertyIndexArgument) []propertyIndexArgument {
	args := append([]propertyIndexArgument(nil), indices...)
	if prop.HasIndexValue {
		// The directive is an already analyzed constant, after written indices.
		args = append(args, propertyIndexArgument{typ: prop.IndexValueType})
	}
	return args
}

func (a *Analyzer) checkPropertyAccessorArguments(args []propertyIndexArgument, signature *types.FunctionType, pos token.Position) {
	mark := len(a.errors)
	for i, arg := range args {
		if i >= len(signature.Parameters) {
			break
		}
		a.checkPropertyIndexArgument(arg, i, signature, pos)
	}

	if !a.errorsSince(mark) {
		a.addMemberCallCountError(signature, len(args), pos)
	}
}

func (a *Analyzer) addPropertyIndexTypeError(index int, expected, actual types.Type, pos token.Position) {
	message := fmt.Sprintf("Argument %d expects type %q instead of %q", index, semanticTypeNameForDiagnostic(expected), semanticTypeNameForDiagnostic(actual))
	if _, void := actual.(*types.VoidType); void {
		message = fmt.Sprintf("Argument %d expects type %q", index, semanticTypeNameForDiagnostic(expected))
	}
	diagnostic := NewGenericError(pos, message)
	diagnostic.AfterChildren = true
	a.addStructuredError(diagnostic)
}

func (a *Analyzer) checkPropertyIndexArgument(arg propertyIndexArgument, index int, signature *types.FunctionType, pos token.Position) {
	if arg.typ == nil {
		return
	}
	expected := signature.Parameters[index]
	name := functionParameterName(signature, index)
	if !a.argumentMatchesParameter(arg.typ, expected, false) {
		a.addPropertyIndexTypeError(index, expected, arg.typ, pos)
		return
	}
	if index >= len(signature.VarParams) || !signature.VarParams[index] {
		return
	}
	actualBase, expectedBase := types.GetUnderlyingType(arg.typ), types.GetUnderlyingType(expected)
	// Compatible class assignment can still fail the var parameter's IsOfType.
	// Storage is checked afterward even when that additional type check fails.
	_, actualClass := actualBase.(*types.ClassType)
	_, expectedClass := expectedBase.(*types.ClassType)
	if actualClass && expectedClass && !expectedBase.Equals(actualBase) {
		a.addPropertyIndexTypeError(index, expected, arg.typ, pos)
	}
	// Implicit Integer -> Float conversion is a value, even for writable input.
	widened := actualBase.Equals(types.INTEGER) && expectedBase.Equals(types.FLOAT)
	if !widened && a.isWritableVarArgument(arg.expression, arg.typ) {
		a.markVarArgumentWritten(arg.expression)
		return
	}
	diagnostic := NewGenericError(pos, fmt.Sprintf("Argument %d (%s) cannot be passed as Var-parameter", index, name))
	diagnostic.AfterChildren = true
	a.addStructuredError(diagnostic)
}

func propertyIndexClassReceiver(typ types.Type) (*types.ClassType, bool) {
	resolved := types.GetUnderlyingType(typ)
	if meta, ok := resolved.(*types.ClassOfType); ok {
		return meta.ClassType, true
	}
	class, ok := resolved.(*types.ClassType)
	if !ok {
		return nil, false
	}
	return class, false
}

func (a *Analyzer) propertyClassAccessor(class *types.ClassType, prop *types.PropertyInfo, write bool) *types.FunctionType {
	name, kind := prop.ReadSpec, prop.ReadKind
	if write {
		if signature := a.propertyWriterSignatures[prop]; signature != nil {
			return signature
		}
		name, kind = prop.WriteSpec, prop.WriteKind
	}
	if kind != types.PropAccessMethod {
		return nil
	}
	if !write && prop.ReadOwner != "" {
		for owner := class; owner != nil; owner = owner.Parent {
			if ident.Equal(owner.Name, prop.ReadOwner) {
				signature, _ := owner.GetMethod(name)
				return signature
			}
		}
	}
	signature, _ := class.GetMethod(name)
	return signature
}

// analyzeVarClassPropertyArguments is entered after receiver/property lookup.
// Writes consume the RHS before the selected method checks its indices.
func (a *Analyzer) analyzeVarClassPropertyArguments(expr *ast.IndexExpression, prop *types.PropertyInfo, class *types.ClassType, indices []ast.Expression, member *ast.MemberAccessExpression, metaclass bool, stmt *ast.AssignmentStatement) types.Type {
	pos := expr.Token.Pos
	if member != nil {
		pos = member.Member.Token.Pos
	}
	write := stmt != nil
	a.warnDeprecatedPropertyUsage(prop, pos)
	if !write && prop.ReadKind == types.PropAccessNone {
		a.addStructuredError(NewWriteOnlyPropertyError(pos, prop.Name))
		a.addStructuredError(NewGenericError(expr.Token.Pos, "Array expected"))
		return prop.Type
	}
	args, stopped := a.readPropertyIndexArguments(prop, indices, expr)
	if stopped {
		return nil
	}
	if !a.checkVarClassPropertyAccess(prop, class, member, metaclass, write, pos) {
		return prop.Type
	}

	if write {
		value, stopped := a.readPropertyAssignmentValue(stmt, prop.Type)
		if stopped {
			return nil
		}
		a.checkPropertyWriteArguments(stmt, prop, args, value, a.propertyClassAccessor(class, prop, true), pos)
	} else {
		a.checkPropertyReadArguments(prop, args, a.propertyClassAccessor(class, prop, false), pos)
	}
	return prop.Type
}

func (a *Analyzer) checkVarClassPropertyAccess(prop *types.PropertyInfo, class *types.ClassType, member *ast.MemberAccessExpression, metaclass, write bool, pos token.Position) bool {
	if write {
		if member != nil {
			return a.checkIndexedPropertyWriteTarget(member, prop)
		}
		if prop.WriteKind == types.PropAccessNone {
			a.addStructuredError(NewReadOnlyPropertyError(pos, prop.Name))
			return false
		}
	} else if metaclass && member != nil {
		return a.checkIndexedPropertyMetaclassAccess(class, prop, member)
	}
	return true
}

func (a *Analyzer) analyzeVarIndexedClassAssignment(expr *ast.IndexExpression, stmt *ast.AssignmentStatement) bool {
	nodes := []*ast.IndexExpression{expr}
	root := expr.Left
	for nodes[len(nodes)-1].CommaPos.IsValid() {
		inner, ok := root.(*ast.IndexExpression)
		if !ok {
			return false
		}
		nodes = append(nodes, inner)
		root = inner.Left
	}
	var member *ast.MemberAccessExpression
	receiver := root
	if m, ok := root.(*ast.MemberAccessExpression); ok {
		member = m
		receiver = m.Object
	}
	// Compatibility/inherited writes keep their deliberately separate boundary.
	if _, inherited := root.(*ast.InheritedExpression); inherited {
		return false
	}
	if !a.varClassAssignmentCandidate(receiver, member) {
		return false
	}
	objectType := a.analyzeExpression(receiver)
	resolved := types.GetUnderlyingType(objectType)
	class, ok := resolved.(*types.ClassType)
	metaclass := false
	if meta, isMeta := resolved.(*types.ClassOfType); isMeta {
		class, ok, metaclass = meta.ClassType, true, true
	}
	if !ok {
		return false
	}
	var prop *types.PropertyInfo
	if member != nil {
		prop, _ = class.GetProperty(member.Member.Value)
	} else {
		prop = a.getDefaultClassProperty(class)
	}
	if prop == nil || !prop.IsIndexed || !propertyHasVarIndices(prop) {
		return false
	}
	a.analyzeVarClassPropertyArguments(expr, prop, class, indexedPropertyArguments(nodes), member, metaclass, stmt)
	return true
}

// A setter's RHS is read before TypeCheckArguments, and a child compiler stop
// aborts that later phase just as a stopped index term does.
func (a *Analyzer) readPropertyAssignmentValue(stmt *ast.AssignmentStatement, expected types.Type) (types.Type, bool) {
	mark := len(a.structuredErrors)
	previous := a.propertyTermMark
	a.propertyTermMark = &mark
	defer func() { a.propertyTermMark = previous }()
	value := a.analyzeAssignmentValue(stmt, expected)
	return value, a.propertyArgumentsStopped(mark)
}

func (a *Analyzer) checkPropertyWriteArguments(stmt *ast.AssignmentStatement, prop *types.PropertyInfo, indices []propertyIndexArgument, value types.Type, signature *types.FunctionType, pos token.Position) {
	if prop.WriteKind == types.PropAccessField {
		if value != nil && !a.canAssign(value, prop.Type) {
			a.reportPropertyAssignmentMismatch(stmt.Value, prop, value, pos, stmt.Token.Pos)
		}
		return
	}
	args := propertyAccessorArguments(prop, indices)
	args = append(args, propertyIndexArgument{expression: stmt.Value, typ: value})
	a.checkPropertyAccessorArguments(args, propertyAccessorSignature(prop, signature, true), pos)
}

func (a *Analyzer) varClassAssignmentCandidate(receiver ast.Expression, member *ast.MemberAccessExpression) bool {
	inferred := a.inferMemberObjectType(receiver)
	if inferred == nil {
		return true
	}
	class, _ := propertyIndexClassReceiver(inferred)
	if class == nil {
		return false
	}
	var prop *types.PropertyInfo
	if member != nil {
		prop, _ = class.GetProperty(member.Member.Value)
	} else {
		prop = a.getDefaultClassProperty(class)
	}
	return prop != nil && prop.IsIndexed && propertyHasVarIndices(prop)
}

// Copies declaration-selected immutable signatures, including forwarding and
// visibility-only promotion. This metadata belongs to this analysis run.
func (a *Analyzer) copyPropertyWriterSignature(target, source *types.PropertyInfo) {
	if signature := a.propertyWriterSignatures[source]; signature != nil {
		a.propertyWriterSignatures[target] = signature
	}
}

func (a *Analyzer) analyzeDeferredIndexPrefix(index *ast.IndexExpression) bool {
	binary, ok := index.Index.(*ast.BinaryExpression)
	if !ok {
		return false
	}
	invalid, ok := binary.Right.(*ast.InvalidExpression)
	if !ok || invalid.Token.Type != token.RBRACK {
		return false
	}
	deferred := false
	for _, candidate := range a.deferredIndexStops {
		if candidate == index {
			deferred = true
			break
		}
	}
	if !deferred {
		return false
	}
	mark := len(a.structuredErrors)
	previous := a.propertyTermMark
	a.propertyTermMark = &mark
	defer func() { a.propertyTermMark = previous }()
	a.analyzeExpression(binary.Left)
	a.compileStopped = true
	return true
}
