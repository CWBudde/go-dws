package semantic

import (
	"slices"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// helperSignatureDeclaration locates a selected user signature without changing
// helper lookup or overload selection. Builtins have no user declaration.
func (a *Analyzer) helperSignatureDeclaration(receiver types.Type, name string, signature *types.FunctionType) (*types.HelperType, *ast.FunctionDecl) {
	key := ident.Normalize(name)
	for _, helper := range a.getHelpersForType(receiver) {
		for owner := helper; owner != nil; owner = owner.ParentHelper {
			if owner.BuiltinMethods[key] != "" {
				continue
			}
			switch declaration := owner.Decl.(type) {
			case *ast.FunctionDecl:
				if owner.Methods[key] == signature {
					return owner, declaration
				}
			case *ast.HelperDecl:
				index := slices.Index(owner.MethodOverloads[key], signature)
				if index < 0 {
					continue
				}
				for _, method := range declaration.Methods {
					if ident.Equal(method.Name.Value, name) {
						if index == 0 {
							return owner, method
						}
						index--
					}
				}
			}
		}
	}
	return nil, nil
}

// dispatchHelperReference selects a referenced user helper method as runtime
// dispatch does: the first declared helper declaring the name wins, unless a
// later helper inherits from it. Only overloads assignable to expected qualify.
func (a *Analyzer) dispatchHelperReference(receiver types.Type, name string, expected types.Type) (*types.HelperType, *types.FunctionType) {
	key := ident.Normalize(name)
	var winner *types.HelperType
	for _, helper := range a.getHelpersForType(receiver) {
		if !helperChainDeclares(helper, key) {
			continue
		}
		if winner == nil || helperInherits(helper, winner) {
			winner = helper
		}
	}
	for owner := winner; owner != nil; owner = owner.ParentHelper {
		if owner.BuiltinMethods[key] != "" {
			continue
		}
		for _, candidate := range helperOwnOverloads(owner, key) {
			pointer := types.FunctionPointerFromFunctionType(candidate)
			if a.canAssign(pointer, expected) {
				return owner, candidate
			}
		}
	}
	return nil, nil
}

func helperChainDeclares(helper *types.HelperType, key string) bool {
	for owner := helper; owner != nil; owner = owner.ParentHelper {
		if owner.BuiltinMethods[key] == "" && len(helperOwnOverloads(owner, key)) > 0 {
			return true
		}
	}
	return false
}

func helperOwnOverloads(owner *types.HelperType, key string) []*types.FunctionType {
	if overloads := owner.MethodOverloads[key]; len(overloads) > 0 {
		return overloads
	}
	if method := owner.Methods[key]; method != nil {
		return []*types.FunctionType{method}
	}
	return nil
}

func helperInherits(helper, ancestor *types.HelperType) bool {
	for owner := helper.ParentHelper; owner != nil; owner = owner.ParentHelper {
		if owner == ancestor {
			return true
		}
	}
	return false
}

// analyzeBareHelperMethod selects a compatible reference or one implicit call.
// Builtin helpers keep their specialized argument readers and existing policy.
func (a *Analyzer) analyzeBareHelperMethod(expr *ast.MemberAccessExpression, receiver types.Type, signature *types.FunctionType, expected types.Type) types.Type {
	owner, declaration := a.helperSignatureDeclaration(receiver, expr.Member.Value, signature)
	if declaration == nil || a.isHelperNativeSelfRead(expr) {
		if len(signature.Parameters) == 0 {
			return signature.ReturnType
		}
		return signature
	}
	if expected != nil && types.IsPointerType(expected) {
		if dispatchOwner, reference := a.dispatchHelperReference(receiver, expr.Member.Value, expected); reference != nil {
			pointer := types.FunctionPointerFromFunctionType(reference)
			pointer.Name = a.declaredHelperMethodName(receiver, expr.Member.Value)
			a.annotateMemberPointerType(expr, pointer)
			a.semanticInfo.SetType(expr.Member, &ast.TypeAnnotation{Token: expr.Member.Token, Name: helperMemberAnnotation(dispatchOwner, expr.Member.Value, reference)})
			return pointer
		}
		pointer := types.FunctionPointerFromFunctionType(signature)
		pointer.Name = a.declaredHelperMethodName(receiver, expr.Member.Value)
		if a.canAssign(pointer, expected) {
			a.annotateMemberPointerType(expr, pointer)
			a.semanticInfo.SetType(expr.Member, &ast.TypeAnnotation{Token: expr.Member.Token, Name: helperMemberAnnotation(owner, expr.Member.Value, signature)})
			return pointer
		}
	}
	for _, candidate := range owner.MethodOverloads[ident.Normalize(expr.Member.Value)] {
		if len(candidate.Parameters) == 0 {
			signature = candidate
			_, declaration = a.helperSignatureDeclaration(receiver, expr.Member.Value, candidate)
			break
		}
	}
	required := requiredHelperDeclarationParams(declaration)
	if required > 0 {
		a.addMoreArgumentsExpected(expr.Member.Token.Pos)
	}
	a.semanticInfo.SetType(expr.Member, &ast.TypeAnnotation{Token: expr.Member.Token, Name: helperMemberAnnotation(owner, expr.Member.Value, signature)})
	// The evaluator captures the original helper and calls it once. The result
	// may itself be a pointer and must remain a value of this member read.
	a.semanticInfo.SetResolvedType(expr.Member, signature)
	a.semanticInfo.SetImplicitCall(expr)
	if signature.ReturnType == nil {
		return types.VOID
	}
	return signature.ReturnType
}

func (a *Analyzer) analyzeBareBoundHelper(identifier *ast.Identifier, expected types.Type) (types.Type, bool) {
	symbol, found := a.symbols.Resolve(identifier.Value)
	if !found || (a.currentFunction != nil && ident.Equal(a.currentFunction.Name.Value, identifier.Value)) {
		return nil, false
	}
	signature, function := symbol.Type.(*types.FunctionType)
	if !function || !a.isBoundHelperMethod(identifier.Value, signature) {
		return nil, false
	}
	a.analyzeIdentifier(identifier) // Preserve usage, casing and deprecation.
	member := &ast.MemberAccessExpression{BaseNode: identifier.BaseNode, Member: identifier, Object: &ast.SelfExpression{BaseNode: identifier.BaseNode}}
	result := a.analyzeBareHelperMethod(member, a.currentHelperType.TargetType, signature, expected)
	if annotation := a.semanticInfo.GetType(identifier); annotation != nil {
		annotation.Name = "__helper_body_member:" + annotation.Name[len("__helper_member:"):]
	}
	if a.semanticInfo.IsImplicitCall(member) {
		a.semanticInfo.SetImplicitCall(identifier)
	}
	return result, true
}

// isGroupedHelperValue limits the returned-callable argument policy to helper
// reads. Other indirect-call contexts retain their separate existing reader.
func (a *Analyzer) isGroupedHelperValue(expr ast.Expression) bool {
	group, grouped := expr.(*ast.GroupedExpression)
	if !grouped {
		return false
	}
	inner := group.Expression
	for {
		nested, ok := inner.(*ast.GroupedExpression)
		if !ok {
			break
		}
		inner = nested.Expression
	}
	var member ast.Expression
	switch value := inner.(type) {
	case *ast.MemberAccessExpression:
		if _, explicit := a.semanticInfo.GetResolvedType(value.Object).(*types.HelperType); explicit {
			return true
		}
		member = value.Member
	case *ast.Identifier:
		member = value
	default:
		return false
	}
	annotation := a.semanticInfo.GetType(member)
	return annotation != nil && (strings.HasPrefix(annotation.Name, "__helper_member:") || strings.HasPrefix(annotation.Name, "__helper_body_member:"))
}

func (a *Analyzer) analyzeGroupedHelperValueCall(expr *ast.CallExpression, callee types.Type) (types.Type, bool) {
	if !a.isGroupedHelperValue(expr.Function) {
		return nil, false
	}
	var pointer *types.FunctionPointerType
	switch value := types.GetUnderlyingType(callee).(type) {
	case *types.FunctionPointerType:
		pointer = value
	case *types.MethodPointerType:
		pointer = &value.FunctionPointerType
	default:
		return nil, false
	}
	signature := types.NewFunctionTypeWithMetadata(pointer.Parameters, nil, nil, pointer.LazyParams, pointer.VarParams, pointer.ConstParams, functionPointerCallResult(pointer))
	a.analyzeMemberCallArguments(signature, expr.Arguments, expr.Token.Pos, false)
	return signature.ReturnType, true
}

func helperMemberAnnotation(owner *types.HelperType, name string, signature *types.FunctionType) string {
	index := slices.Index(owner.MethodOverloads[ident.Normalize(name)], signature)
	if index < 0 {
		index = 0
	}
	return "__helper_member:" + owner.Name + ":" + strconv.Itoa(index)
}

func requiredHelperDeclarationParams(declaration *ast.FunctionDecl) int {
	required := 0
	for i, parameter := range declaration.Parameters {
		if declaration.IsHelper && i == 0 {
			continue
		}
		if parameter.DefaultValue == nil {
			required++
		}
	}
	return required
}

// isHelperNativeSelfRead preserves Self.ClassName's native-member escape inside
// a same-named helper method, rather than recursively reading that helper.
func (a *Analyzer) isHelperNativeSelfRead(expr *ast.MemberAccessExpression) bool {
	if a.currentHelperType == nil || a.currentFunction == nil || !ident.Equal(a.currentFunction.Name.Value, expr.Member.Value) {
		return false
	}
	_, self := expr.Object.(*ast.SelfExpression)
	return self
}
