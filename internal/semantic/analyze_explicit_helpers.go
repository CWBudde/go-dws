package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

type explicitHelperSignature struct {
	typ    *types.FunctionType
	owner  string
	index  int
	marked bool
}

// explicitHelperSignatures describes the written arguments of a call through a
// helper name. Instance methods take an explicit first receiver; nonstatic class
// methods on structured targets take a type receiver instead.
func (a *Analyzer) explicitHelperSignatures(helper *types.HelperType, name string) ([]explicitHelperSignature, string) {
	for owner := helper; owner != nil; owner = owner.ParentHelper {
		decl, ok := owner.Decl.(*ast.HelperDecl)
		if !ok {
			continue
		}
		var signatures []explicitHelperSignature
		var declaredName string
		for _, method := range decl.Methods {
			if !ident.Equal(method.Name.Value, name) {
				continue
			}
			declaredName = method.Name.Value
			var params []types.Type
			var names []string
			var defaults []interface{}
			var lazy, byRef, constant, strict []bool
			if receiver := explicitHelperReceiverType(owner.TargetType, method); receiver != nil && !method.IsHelper {
				params = append(params, receiver)
				names = append(names, "Self")
				defaults = append(defaults, nil)
				lazy, byRef, constant = append(lazy, false), append(byRef, false), append(constant, false)
				// A record class receiver must be the actual matching metatype;
				// the general Variant conversion policy cannot supply this role.
				_, recordMeta := receiver.(*types.RecordMetaType)
				strict = append(strict, recordMeta)
			}
			for _, param := range method.Parameters {
				paramType, err := a.resolveTypeExpression(param.Type)
				if err != nil {
					return nil, declaredName
				}
				params, names = append(params, paramType), append(names, param.Name.Value)
				// Keep an absent default as a nil interface, not a typed nil.
				var defaultValue interface{}
				if param.DefaultValue != nil {
					defaultValue = param.DefaultValue
				}
				defaults = append(defaults, defaultValue)
				lazy, byRef, constant = append(lazy, param.IsLazy), append(byRef, param.ByRef), append(constant, param.IsConst)
				strict = append(strict, isStrictTypeAnnotation(param.Type))
			}
			typ := types.NewFunctionTypeWithMetadata(params, names, defaults,
				lazy, byRef, constant, a.helperMethodReturnType(method))
			typ.StrictParams = strict
			signatures = append(signatures, explicitHelperSignature{typ: typ, owner: owner.Name, index: len(signatures), marked: method.IsOverload})
		}
		if len(signatures) != 0 {
			return signatures, declaredName
		}
	}
	return nil, ""
}

func (a *Analyzer) analyzeExplicitHelperCall(helper *types.HelperType, member *ast.Identifier, args []ast.Expression) (types.Type, bool) {
	signatures, declaredName := a.explicitHelperSignatures(helper, member.Value)
	if len(signatures) == 0 {
		return nil, false
	}
	a.addIdentifierCaseHint(member, declaredName)
	selected, ok := a.selectExplicitHelperOverload(signatures, member, args, declaredName)
	if !ok {
		return nil, true
	}
	signature := signatures[selected].typ
	a.annotateExplicitHelperCall(member, signatures[selected])
	// Self is a written argument through a helper name, so its index and
	// position participate in the ordinary argument list without a shift.
	a.analyzeMemberCallArguments(signature, args, member.Token.Pos, false)
	if len(args) < requiredParamCount(signature) || len(args) > len(signature.Parameters) {
		return signature.ReturnType, true
	}
	for i, arg := range args {
		if signature.VarParams[i] {
			a.markVarArgumentWritten(arg)
			if !a.isLValue(arg) {
				a.addError("var parameter %d to function '%s' requires a variable (identifier, array element, or field), got %s at %s",
					i+1, member.Value, arg.String(), arg.Pos().String())
			}
		}
	}
	return signature.ReturnType, true
}

func (a *Analyzer) selectExplicitHelperOverload(signatures []explicitHelperSignature, member *ast.Identifier, args []ast.Expression, declaredName string) (int, bool) {
	candidates := make([]*types.FunctionType, len(signatures))
	for i, signature := range signatures {
		candidates[i] = signature.typ
	}
	selected := a.selectHelperCallOverload(candidates, signatures[len(signatures)-1].marked, args, member.Token.Pos, declaredName)
	return selected, selected >= 0
}

// DWScript's CreateSelfParameter uses the target's metaclass for a nonstatic
// class helper method. Primitive and interface class helpers have no receiver.
func explicitHelperReceiverType(target types.Type, method *ast.FunctionDecl) types.Type {
	if !method.IsClassMethod {
		return target
	}
	if method.IsStatic {
		return nil
	}
	switch target := types.GetUnderlyingType(target).(type) {
	case *types.ClassType:
		return types.NewClassOfType(target)
	case *types.ClassOfType:
		return target
	case *types.RecordType:
		return types.NewRecordMetaType(target)
	default:
		return nil
	}
}
