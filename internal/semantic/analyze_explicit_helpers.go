package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

type explicitHelperSignature struct {
	typ    *types.FunctionType
	method *ast.FunctionDecl
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
			var lazy, byRef, constant []bool
			if receiver := explicitHelperReceiverType(owner.TargetType, method); receiver != nil && !method.IsHelper {
				params = append(params, receiver)
				names = append(names, "Self")
				defaults = append(defaults, nil)
				lazy, byRef, constant = append(lazy, false), append(byRef, false), append(constant, false)
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
			}
			signatures = append(signatures, explicitHelperSignature{
				typ: types.NewFunctionTypeWithMetadata(params, names, defaults,
					lazy, byRef, constant, a.helperMethodReturnType(method)),
				method: method,
			})
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
	selected, ok := a.selectExplicitHelperOverload(signatures, member, args)
	if !ok {
		return nil, true
	}
	signature := signatures[selected].typ
	required := 0
	for _, defaultValue := range signature.DefaultValues {
		if defaultValue == nil {
			required++
		}
	}
	if len(args) < required || len(args) > len(signature.Parameters) {
		a.addArgumentCountError(member.Token.Pos, len(args), required, len(signature.Parameters))
		return signature.ReturnType, true
	}
	if !a.validateExplicitHelperReceiver(helper, signatures[selected].method, args) {
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
		a.analyzeCallArgument(i, arg, signature.Parameters[i])
	}
	return signature.ReturnType, true
}

func (a *Analyzer) validateExplicitHelperReceiver(helper *types.HelperType, method *ast.FunctionDecl, args []ast.Expression) bool {
	if method.IsClassMethod && !method.IsStatic && len(args) > 0 {
		if _, record := types.GetUnderlyingType(helper.TargetType).(*types.RecordType); record && !a.isRecordTypeReceiver(args[0]) {
			a.addError("record type expected at %s", args[0].Pos().String())
			return false
		}
	}
	return true
}

func (a *Analyzer) selectExplicitHelperOverload(signatures []explicitHelperSignature, member *ast.Identifier, args []ast.Expression) (int, bool) {
	selected := 0
	if len(signatures) > 1 {
		candidates := make([]types.Type, len(signatures))
		for i, signature := range signatures {
			candidates[i] = signature.typ
		}
		argTypes := make([]types.Type, len(args))
		for i, arg := range args {
			argTypes[i] = a.analyzeOverloadArgument(arg)
			if argTypes[i] == nil {
				return 0, false
			}
		}
		var err error
		selected, err = types.ResolveOverload(candidates, argTypes)
		if err != nil {
			a.addStructuredError(NewNoOverloadMatchError(member.Token.Pos, member.Value))
			return 0, false
		}
	}
	return selected, true
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
		return target
	default:
		return nil
	}
}

// Record names share their semantic type with instances. The declaration
// position distinguishes the type's synthesized binding from a shadowing local.
func (a *Analyzer) isRecordTypeReceiver(expr ast.Expression) bool {
	name, ok := expr.(*ast.Identifier)
	if !ok {
		return false
	}
	descriptor, found := a.typeRegistry.ResolveDescriptor(name.Value)
	if !found {
		return false
	}
	if symbol, found := a.symbols.Resolve(name.Value); found {
		return symbol.DeclPosition == descriptor.Position
	}
	return true
}
