package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Operator Analysis (Stage 8)
// ============================================================================

func (a *Analyzer) analyzeOperatorDecl(decl *ast.OperatorDecl) {
	if decl == nil {
		return
	}

	if decl.Kind == ast.OperatorKindClass {
		// Class operators are processed as part of class analysis
		return
	}

	operandTypes := make([]types.Type, decl.Arity)
	for i, operand := range decl.OperandTypes {
		typ, err := a.resolveOperatorType(operand.String())
		if err != nil {
			a.addError("unknown type '%s' in operator declaration at %s", operand.String(), decl.Token.Pos.String())
			return
		}
		operandTypes[i] = typ
	}

	var resultType types.Type = types.VOID
	if decl.ReturnType != nil {
		var err error
		resultType, err = a.resolveOperatorType(decl.ReturnType.String())
		if err != nil {
			a.addError("unknown return type '%s' in operator declaration at %s", decl.ReturnType.String(), decl.Token.Pos.String())
			return
		}
	}

	if decl.Binding == nil {
		a.addError("operator '%s' missing binding at %s", decl.OperatorSymbol, decl.Token.Pos.String())
		return
	}

	sym, ok := a.symbols.Resolve(decl.Binding.Value)
	if !ok {
		a.addError("binding '%s' for operator '%s' not found at %s", decl.Binding.Value, decl.OperatorSymbol, decl.Token.Pos.String())
		return
	}

	funcType, ok := sym.Type.(*types.FunctionType)
	if !ok {
		a.addError("binding '%s' for operator '%s' is not a function at %s", decl.Binding.Value, decl.OperatorSymbol, decl.Token.Pos.String())
		return
	}

	if len(funcType.Parameters) != len(operandTypes) {
		a.addError("binding '%s' for operator '%s' expects %d parameters, got %d at %s",
			decl.Binding.Value, decl.OperatorSymbol, len(operandTypes), len(funcType.Parameters), decl.Token.Pos.String())
		return
	}

	for i, paramType := range funcType.Parameters {
		if !paramType.Equals(operandTypes[i]) {
			a.addError("binding '%s' parameter %d type %s does not match operator operand type %s at %s",
				decl.Binding.Value, i+1, paramType.String(), operandTypes[i].String(), decl.Token.Pos.String())
			return
		}
	}

	if decl.Kind == ast.OperatorKindConversion {
		if len(operandTypes) != 1 {
			a.addError("conversion operator '%s' must have exactly one operand at %s", decl.OperatorSymbol, decl.Token.Pos.String())
			return
		}
		if resultType == types.VOID {
			a.addError("conversion operator '%s' must specify a return type at %s", decl.OperatorSymbol, decl.Token.Pos.String())
			return
		}

		kind := types.ConversionExplicit
		if ident.Equal(decl.OperatorSymbol, "implicit") {
			kind = types.ConversionImplicit
		}

		sig := &types.ConversionSignature{
			From:    operandTypes[0],
			To:      resultType,
			Binding: decl.Binding.Value,
			Kind:    kind,
		}

		if err := a.conversionRegistry.Register(sig); err != nil {
			a.addError("conversion from %s to %s already defined at %s", operandTypes[0].String(), resultType.String(), decl.Token.Pos.String())
		}
		return
	}

	sig := &types.OperatorSignature{
		Operator:     decl.OperatorSymbol,
		OperandTypes: operandTypes,
		ResultType:   resultType,
		Binding:      decl.Binding.Value,
	}

	if err := a.globalOperators.Register(sig); err != nil {
		opSignatures := make([]string, len(operandTypes))
		for i, typ := range operandTypes {
			opSignatures[i] = typ.String()
		}
		a.addError("operator '%s' already defined for operand types (%s) at %s",
			decl.OperatorSymbol, strings.Join(opSignatures, ", "), decl.Token.Pos.String())
	}
}

func (a *Analyzer) resolveBinaryOperator(operator string, leftType, rightType types.Type) (*types.OperatorSignature, bool) {
	if classType, ok := leftType.(*types.ClassType); ok {
		if sig, found := classType.LookupOperator(operator, []types.Type{leftType, rightType}); found {
			return sig, true
		}
	}
	if classType, ok := rightType.(*types.ClassType); ok {
		if sig, found := classType.LookupOperator(operator, []types.Type{leftType, rightType}); found {
			return sig, true
		}
	}
	if sig, found := a.globalOperators.Lookup(operator, []types.Type{leftType, rightType}); found {
		return sig, true
	}
	return nil, false
}

func (a *Analyzer) resolveUnaryOperator(operator string, operand types.Type) (*types.OperatorSignature, bool) {
	if classType, ok := operand.(*types.ClassType); ok {
		if sig, found := classType.LookupOperator(operator, []types.Type{operand}); found {
			return sig, true
		}
	}
	if sig, found := a.globalOperators.Lookup(operator, []types.Type{operand}); found {
		return sig, true
	}
	return nil, false
}

func (a *Analyzer) registerClassOperators(classType *types.ClassType, decl *ast.ClassDecl) {
	for _, opDecl := range decl.Operators {
		if opDecl == nil {
			continue
		}

		if opDecl.Binding == nil {
			a.addError("class operator '%s' missing binding in class '%s' at %s",
				opDecl.OperatorSymbol, classType.Name, opDecl.Token.Pos.String())
			continue
		}

		// Look up method using overload system
		methodOverloads := classType.GetMethodOverloads(opDecl.Binding.Value)
		if len(methodOverloads) == 0 {
			a.addError("binding '%s' for class operator '%s' not found in class '%s' at %s",
				opDecl.Binding.Value, opDecl.OperatorSymbol, classType.Name, opDecl.Token.Pos.String())
			continue
		}

		// For class operators, use the first matching overload
		// (In the future, we could support overloaded operators with different parameter types)
		methodInfo := methodOverloads[0]
		methodType := methodInfo.Signature

		isClassMethod := methodInfo.IsClassMethod

		operandTypes := make([]types.Type, 0, len(opDecl.OperandTypes)+1)
		includesClass := false
		for _, operand := range opDecl.OperandTypes {
			resolved, err := a.resolveOperatorType(operand.String())
			if err != nil {
				a.addError("unknown type '%s' in class operator declaration at %s", operand.String(), opDecl.Token.Pos.String())
				includesClass = false
				operandTypes = nil
				break
			}
			if resolved.Equals(classType) {
				includesClass = true
			}
			operandTypes = append(operandTypes, resolved)
		}
		if operandTypes == nil {
			continue
		}

		if !includesClass {
			if ident.Equal(opDecl.OperatorSymbol, "in") {
				operandTypes = append(operandTypes, classType)
			} else {
				operandTypes = append([]types.Type{classType}, operandTypes...)
			}
		}

		expectedParams := len(operandTypes)
		if !isClassMethod {
			expectedParams--
		}
		if len(methodType.Parameters) != expectedParams {
			a.addError("binding '%s' for class operator '%s' expects %d parameters, got %d at %s",
				opDecl.Binding.Value, opDecl.OperatorSymbol, expectedParams, len(methodType.Parameters), opDecl.Token.Pos.String())
			continue
		}

		paramIdx := 0
		mismatch := false
		for _, operandType := range operandTypes {
			if !isClassMethod && operandType.Equals(classType) {
				continue
			}
			if paramIdx >= len(methodType.Parameters) {
				a.addError("binding '%s' parameter count mismatch for operator '%s' at %s",
					opDecl.Binding.Value, opDecl.OperatorSymbol, opDecl.Token.Pos.String())
				mismatch = true
				break
			}
			if !methodType.Parameters[paramIdx].Equals(operandType) {
				a.addError("binding '%s' parameter %d type %s does not match operator operand type %s at %s",
					opDecl.Binding.Value, paramIdx+1, methodType.Parameters[paramIdx].String(), operandType.String(), opDecl.Token.Pos.String())
				mismatch = true
				break
			}
			paramIdx++
		}
		if mismatch {
			continue
		}

		resultType := methodType.ReturnType
		if opDecl.ReturnType != nil {
			var err error
			resultType, err = a.resolveOperatorType(opDecl.ReturnType.String())
			if err != nil {
				a.addError("unknown return type '%s' in class operator declaration at %s", opDecl.ReturnType.String(), opDecl.Token.Pos.String())
				continue
			}
			if !methodType.ReturnType.Equals(resultType) {
				a.addError("binding '%s' return type %s does not match operator return type %s at %s",
					opDecl.Binding.Value, methodType.ReturnType.String(), resultType.String(), opDecl.Token.Pos.String())
				continue
			}
		}

		sig := &types.OperatorSignature{
			Operator:     opDecl.OperatorSymbol,
			OperandTypes: operandTypes,
			ResultType:   resultType,
			Binding:      opDecl.Binding.Value,
		}

		if err := classType.RegisterOperator(sig); err != nil {
			a.addError("class operator '%s' already defined for class '%s' at %s",
				opDecl.OperatorSymbol, classType.Name, opDecl.Token.Pos.String())
		}
	}
}
