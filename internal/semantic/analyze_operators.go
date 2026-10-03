package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
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

	expectedCount := 2
	if decl.Kind == ast.OperatorKindConversion || (decl.OperatorSymbol == "-" && len(operandTypes) == 1) {
		expectedCount = 1
	}
	validCount := len(operandTypes) == expectedCount
	if !validCount {
		pos := decl.OperandValidationPos
		if pos.Line == 0 {
			pos = decl.Pos()
		}
		a.reportOperatorDeclarationError(pos, fmt.Sprintf("Expected %d parameters (instead of %d)", expectedCount, len(operandTypes)))
	}
	duplicate := false
	if validCount && decl.OperatorSymbol != "" && decl.Kind == ast.OperatorKindGlobal {
		duplicate = a.recordOperatorDeclaration(decl.OperatorSymbol, operandTypes)
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
	if decl.BindingHelper != nil {
		if validCount && decl.OperatorSymbol != "" {
			a.analyzeHelperOperatorBinding(decl, operandTypes, resultType, duplicate)
		}
		return
	}

	sym, ok := a.symbols.Resolve(decl.Binding.Value)
	if !ok {
		if sig, builtin := a.builtinRegistry.GetSignature(decl.Binding.Value); builtin &&
			len(operandTypes) >= sig.MinArgs && len(operandTypes) <= sig.MaxArgs && len(sig.ParamTypes) >= len(operandTypes) &&
			sig.ReturnType != nil && types.OperatorTypesEqual(sig.ReturnType, resultType) {
			matching := true
			for i, operand := range operandTypes {
				if !types.OperatorTypesEqual(sig.ParamTypes[i], operand) {
					matching = false
					break
				}
			}
			if matching {
				if decl.Kind == ast.OperatorKindConversion {
					if !validCount {
						return
					}
					kind := types.ConversionExplicit
					if ident.Equal(decl.OperatorSymbol, "implicit") {
						kind = types.ConversionImplicit
					}
					if err := a.conversionRegistry.Register(&types.ConversionSignature{From: operandTypes[0], To: resultType, Binding: decl.Binding.Value, Kind: kind}); err != nil {
						a.addError("conversion from %s to %s already defined at %s", operandTypes[0], resultType, decl.Token.Pos.String())
					}
					return
				}
			}
		}
		a.addError("binding '%s' for operator '%s' not found at %s", decl.Binding.Value, decl.OperatorSymbol, decl.Token.Pos.String())
		return
	}

	funcType, ok := sym.Type.(*types.FunctionType)
	if !ok {
		a.addError("binding '%s' for operator '%s' is not a function at %s", decl.Binding.Value, decl.OperatorSymbol, decl.Token.Pos.String())
		return
	}

	pos := decl.Binding.Pos()
	// DWScript validates result type before binding parameters. Its expected
	// parameter count comes from the operator, even after operand-count errors.
	if funcType.ReturnType == nil || !operatorBindingTypesCompatible(funcType.ReturnType, resultType) {
		a.reportOperatorDeclarationError(pos, fmt.Sprintf(`Result type should be "%s"`, semanticTypeNameForDiagnostic(resultType)))
		return
	}
	if len(funcType.Parameters) != expectedCount {
		a.reportOperatorDeclarationError(pos, fmt.Sprintf("Expected %d parameters (instead of %d)", expectedCount, len(funcType.Parameters)))
		return
	}
	for i, paramType := range funcType.Parameters {
		if i >= len(operandTypes) {
			break
		}
		if !operatorBindingTypesCompatible(paramType, operandTypes[i]) {
			a.reportOperatorDeclarationError(pos, fmt.Sprintf(`Parameter %d - Type "%s" expected (instead of "%s")`,
				i, semanticTypeNameForDiagnostic(operandTypes[i]), semanticTypeNameForDiagnostic(paramType)))
			return
		}
		if i < len(funcType.VarParams) && funcType.VarParams[i] {
			a.reportOperatorDeclarationError(pos, fmt.Sprintf("Parameter %d - Var-parameter forbidden", i))
			return
		}
	}
	if !validCount || decl.OperatorSymbol == "" {
		return
	}
	if duplicate {
		a.reportOperatorDeclarationError(decl.Pos(), "An overload already exists for this operator and types")
		return
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
		a.reportOperatorDeclarationError(decl.Pos(), "An overload already exists for this operator and types")
	}
}

// operatorBindingTypesCompatible is a signature relation, not an assignment:
// primitive coercions (notably Integer to Float/Variant) are forbidden, while
// aliases, class ancestry, and interface ancestry retain their type identity.
func operatorBindingTypesCompatible(actual, expected types.Type) bool {
	if types.OperatorTypesEqual(actual, expected) {
		return true
	}
	actual, expected = types.GetUnderlyingType(actual), types.GetUnderlyingType(expected)
	if actual == nil || expected == nil {
		return false
	}
	switch actual := actual.(type) {
	case *types.ClassType:
		_, sameKind := expected.(*types.ClassType)
		return sameKind && types.IsCompatible(actual, expected)
	case *types.InterfaceType:
		_, sameKind := expected.(*types.InterfaceType)
		return sameKind && types.IsCompatible(actual, expected)
	case *types.ClassOfType:
		_, sameKind := expected.(*types.ClassOfType)
		return sameKind && types.IsCompatible(actual, expected)
	case *types.ArrayType:
		right, sameKind := expected.(*types.ArrayType)
		return sameKind && actual.IsDynamic() && right.IsDynamic() && operatorBindingTypesCompatible(right.ElementType, actual.ElementType)
	}
	return false
}

func (a *Analyzer) reportOperatorDeclarationError(pos token.Position, message string) {
	a.addStructuredError(&SemanticError{Type: ErrorInvalidOperation, Pos: pos, Message: message, Severity: SeverityError})
}

// recordOperatorDeclaration tracks source declarations separately from executable
// overloads. Upstream's duplicate check includes earlier rejected bindings in
// the same local symbol table, compares unaliased operand types, and only checks
// binary declarations.
func (a *Analyzer) recordOperatorDeclaration(operator string, operands []types.Type) bool {
	if a.operatorDeclarations == nil {
		a.operatorDeclarations = make(map[*SymbolTable][]*types.OperatorSignature)
	}
	duplicate := false
	if len(operands) == 2 {
		for _, previous := range a.operatorDeclarations[a.symbols] {
			if ident.Equal(previous.Operator, operator) && types.OperatorOperandsEqual(previous.OperandTypes, operands) {
				duplicate = true
				break
			}
		}
	}
	a.operatorDeclarations[a.symbols] = append(a.operatorDeclarations[a.symbols], &types.OperatorSignature{Operator: operator, OperandTypes: operands})
	return duplicate
}

func (a *Analyzer) analyzeHelperOperatorBinding(decl *ast.OperatorDecl, operands []types.Type, result types.Type, duplicate bool) {
	sym, found := a.symbols.Resolve(decl.BindingHelper.Value)
	if !found {
		a.addError("binding '%s.%s' for operator '%s' not found at %s", decl.BindingHelper.Value, decl.Binding.Value, decl.OperatorSymbol, decl.Token.Pos.String())
		return
	}
	helper, ok := sym.Type.(*types.HelperType)
	if !ok || len(operands) < 1 || !types.OperatorTypesEqual(helper.TargetType, operands[0]) {
		a.addError("binding '%s.%s' for operator '%s' has incompatible helper target at %s", decl.BindingHelper.Value, decl.Binding.Value, decl.OperatorSymbol, decl.Token.Pos.String())
		return
	}
	overloads := helper.MethodOverloads[ident.Normalize(decl.Binding.Value)]
	for idx, method := range overloads {
		if len(method.Parameters) != len(operands)-1 || !types.OperatorTypesEqual(method.ReturnType, result) {
			continue
		}
		match := true
		for i, param := range method.Parameters {
			if !types.OperatorTypesEqual(param, operands[i+1]) {
				match = false
				break
			}
		}
		if !match {
			continue
		}
		decl.BindingOverload = idx
		if duplicate {
			a.reportOperatorDeclarationError(decl.Pos(), "An overload already exists for this operator and types")
			return
		}
		if err := a.globalOperators.Register(&types.OperatorSignature{Operator: decl.OperatorSymbol, OperandTypes: operands, ResultType: result, Binding: decl.BindingHelper.Value + "." + decl.Binding.Value}); err != nil {
			a.addError("operator '%s' already defined for operand types (%s) at %s", decl.OperatorSymbol, types.FormatTypeList(operands), decl.Token.Pos.String())
		}
		return
	}
	a.addError("binding '%s.%s' for operator '%s' has no matching overload at %s", decl.BindingHelper.Value, decl.Binding.Value, decl.OperatorSymbol, decl.Token.Pos.String())
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
