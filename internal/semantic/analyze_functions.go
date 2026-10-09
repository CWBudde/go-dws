package semantic

import (
	"strconv"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// ============================================================================
// Function Analysis
// ============================================================================

// analyzeFunctionDecl analyzes a function declaration.
//
// It is the single-pass entry point (used by nested contexts, unit sections, and
// the fallback path in Analyze). For top-level program declarations the driver
// splits this work: registerFunctionSignature runs in an early pass so functions
// are visible regardless of source order, and analyzeFunctionBody runs afterwards.
func (a *Analyzer) analyzeFunctionDecl(decl *ast.FunctionDecl) {
	// Check if this is a method implementation (has ClassName)
	if decl.ClassName != nil {
		a.analyzeMethodImplementation(decl)
		return
	}

	paramTypes, returnType, ok := a.registerFunctionSignature(decl)
	if !ok {
		// Registration failed, or the decl was a helper that fully handles itself.
		return
	}

	// Skip body analysis for forward declarations
	if decl.IsForward {
		return
	}

	a.analyzeFunctionBody(decl, paramTypes, returnType)
}

// registerFunctionSignature resolves a regular (non-method) function's parameter
// and return types and registers its overload in the current symbol table. It does
// NOT analyze the body. It returns the resolved parameter types, the return type,
// and ok=true when the signature was registered and a body pass should follow.
//
// ok=false means either registration failed (an error was already recorded) or the
// declaration was a helper function that is fully analyzed by analyzeFunctionHelperDecl;
// in both cases the caller must not run a body pass.
func (a *Analyzer) registerFunctionSignature(decl *ast.FunctionDecl) (paramTypes []types.Type, returnType types.Type, ok bool) {
	// Export qualifiers are reached only after forward binding allows EXPORT.
	// Preserve convention hints on ordinary routines and early header failures.
	conventionHintPending := decl.IsExport || decl.IsForward
	var forwardHintInsertion *diagnosticInsertion
	if decl.IsForward {
		// Binding may stop at FORWARD before its calling qualifier is reached.
		// Keep a reached hint at its original emission point on other branches.
		forwardHintInsertion = a.newDiagnosticInsertion()
	}
	if !conventionHintPending {
		a.addCallConventionHint(decl)
	}
	defer func() {
		if conventionHintPending {
			if forwardHintInsertion != nil {
				a.analyzeAtDiagnosticInsertion(forwardHintInsertion, func() { a.addCallConventionHint(decl) })
			} else {
				a.addCallConventionHint(decl)
			}
		}
	}()

	// Regular function (not method): resolve parameter and return types
	paramTypes = make([]types.Type, 0, len(decl.Parameters))
	paramNames := make([]string, 0, len(decl.Parameters))
	paramTypeNames := make([]string, 0, len(decl.Parameters))
	defaultValues := make([]interface{}, 0, len(decl.Parameters))
	lazyParams := make([]bool, 0, len(decl.Parameters))
	varParams := make([]bool, 0, len(decl.Parameters))
	constParams := make([]bool, 0, len(decl.Parameters))
	strictParams := make([]bool, 0, len(decl.Parameters))
	foundOptional := false // Track if we've seen an optional parameter

	for _, param := range decl.Parameters {
		// Validate that lazy, var, and const are mutually exclusive
		exclusiveCount := 0
		if param.IsLazy {
			exclusiveCount++
		}
		if param.ByRef {
			exclusiveCount++
		}
		if param.IsConst {
			exclusiveCount++
		}
		if exclusiveCount > 1 {
			a.addError("parameter '%s' cannot have multiple modifiers (lazy/var/const) in function '%s' at %s",
				param.Name.Value, decl.Name.Value, param.Token.Pos.String())
			return nil, nil, false
		}

		// Optional parameters must come last, without modifiers. A parameter
		// whose default was rejected stays required (see below).
		if param.DefaultValue == nil && foundOptional {
			a.addError("required parameter '%s' cannot come after optional parameters in function '%s' at %s",
				param.Name.Value, decl.Name.Value, param.Token.Pos.String())
			return nil, nil, false
		}

		if param.Type == nil {
			a.addError("parameter '%s' missing type annotation in function '%s'",
				param.Name.Value, decl.Name.Value)
			return nil, nil, false
		}
		if isRefusedTypeExpression(param.Type) {
			// The parser already reported "Type expected" for it.
			return nil, nil, false
		}
		paramType, err := a.resolveTypeExpression(param.Type)
		if err == nil && paramType == nil {
			return nil, nil, false
		}
		if err != nil {
			a.addError("unknown parameter type '%s' in function '%s': %v",
				getTypeExpressionName(param.Type), decl.Name.Value, err)
			return nil, nil, false
		}
		if param.IsConst {
			if arrayType, ok := paramType.(*types.ArrayType); ok && arrayType.ElementType == types.VARIANT {
				paramType = types.ARRAY_OF_CONST
			}
		}

		defaultValue := a.analyzeParameterDefault(param, paramType)
		if defaultValue != nil {
			foundOptional = true
		}
		if !param.IsConst && declaresArrayOfConst(param.Type) {
			pos := param.Type.Pos()
			if array, ok := param.Type.(*ast.ArrayTypeNode); ok {
				pos = array.ElementType.Pos()
			}
			a.addError("open array parameter must be const at %s", pos.String())
		}

		paramTypes = append(paramTypes, paramType)
		paramNames = append(paramNames, param.Name.Value)
		paramTypeNames = append(paramTypeNames, semanticDeclaredTypeName(param.Type, paramType))
		defaultValues = append(defaultValues, defaultValue)
		lazyParams = append(lazyParams, param.IsLazy)
		varParams = append(varParams, param.ByRef)
		constParams = append(constParams, param.IsConst)
		strictParams = append(strictParams, isStrictTypeAnnotation(param.Type))
	}

	// Determine return type
	if decl.ReturnType != nil {
		var err error
		returnType, err = a.resolveTypeExpression(decl.ReturnType)
		if err != nil {
			a.addError("unknown return type '%s' in function '%s': %v",
				getTypeExpressionName(decl.ReturnType), decl.Name.Value, err)
			return nil, nil, false
		}
	} else {
		returnType = types.VOID
	}

	// Create function type with metadata (handles lazy, var, const, defaults)
	var funcType *types.FunctionType
	if len(paramTypes) > 0 {
		// Check if last parameter is dynamic array (variadic)
		lastParamType := paramTypes[len(paramTypes)-1]
		if arrayType, ok := lastParamType.(*types.ArrayType); ok && arrayType.IsDynamic() {
			variadicType := arrayType.ElementType
			funcType = types.NewVariadicFunctionTypeWithMetadata(
				paramTypes, paramNames, defaultValues, lazyParams, varParams, constParams,
				variadicType, returnType,
			)
		} else {
			funcType = types.NewFunctionTypeWithMetadata(
				paramTypes, paramNames, defaultValues, lazyParams, varParams, constParams, returnType,
			)
		}
	} else {
		funcType = types.NewFunctionTypeWithMetadata(
			paramTypes, paramNames, defaultValues, lazyParams, varParams, constParams, returnType,
		)
	}
	funcType.Name = decl.Name.Value
	funcType.IsClassMethod = decl.IsClassMethod
	funcType.IsConstructor = decl.IsConstructor
	funcType.IsDestructor = decl.IsDestructor
	funcType.ParamTypeNames = paramTypeNames
	funcType.StrictParams = strictParams

	// Register function/overload with position info for error messages
	// `forward` is meaningless on an external routine (the host implements it),
	// so it is not left awaiting an implementation.
	isForward := decl.IsForward && !decl.IsExternal
	var matchedForward *Symbol
	if !isForward {
		matchedForward = a.symbols.matchingExplicitForward(decl.Name.Value, funcType)
	}
	repeatedExport := decl.IsExport && !isForward && a.symbols.exportImplementsForward(decl.Name.Value, funcType)
	if conventionHintPending && !decl.IsForward {
		conventionHintPending = false
		if !repeatedExport {
			a.addCallConventionHint(decl)
		}
	}
	if decl.IsHelper && !repeatedExport {
		// Helper functions are fully analyzed here (signature + body); there is no
		// separate body pass for them.
		a.analyzeFunctionHelperDecl(decl, paramTypes, returnType)
		return nil, nil, false
	}
	if err := a.symbols.DefineOverload(decl.Name.Value, funcType, decl.IsOverload, isForward, decl.Name.Token.Pos); err != nil {
		pos := decl.Token.Pos
		switch declarationAnchor(err) {
		case anchorHeaderEnd:
			if decl.HeaderEndPos.Line > 0 {
				pos = decl.HeaderEndPos
			}
		case anchorForward:
			if decl.ForwardPos.Line > 0 {
				pos = decl.ForwardPos
			}
		case anchorDecl:
		}
		if declarationStopsCompilation(err) {
			conventionHintPending = false
			a.addPunctuationStop(pos, err.Error())
		} else {
			a.addError("Syntax Error: %s [line: %d, column: %d]", err.Error(), pos.Line, pos.Column)
		}
		if repeatedExport {
			a.addPunctuationStop(decl.ExportPos, "BEGIN expected")
		}
		return nil, nil, false
	}
	if repeatedExport {
		a.addPunctuationStop(decl.ExportPos, "BEGIN expected")
		return nil, nil, false
	}
	if matchedForward != nil {
		// Keep the source header unchanged. Runtime registration uses a separate
		// view of this successfully bound declaration with the original defaults.
		if matchedForward.forwardDefaultSignature != nil {
			a.semanticInfo.SetResolvedType(decl, matchedForward.forwardDefaultSignature)
		}
	}
	if isForward && decl.IsOverload {
		if forward := a.symbols.matchingExplicitForward(decl.Name.Value, funcType); forward != nil {
			forward.forwardDefaultSignature = a.bindForwardDefaultSignature(funcType)
		}
	}

	if decl.IsDeprecated {
		a.symbols.MarkDeprecated(decl.Name.Value, decl.DeprecatedMessage)
	}

	return paramTypes, returnType, true
}

// bindForwardDefaultSignature captures scalar defaults at their declaration,
// rather than allowing a caller's local names to rebind constant references.
// The symbol's original signature and the source expressions stay unchanged.
func (a *Analyzer) bindForwardDefaultSignature(signature *types.FunctionType) *types.FunctionType {
	bound := *signature
	bound.DefaultValues = append([]interface{}(nil), signature.DefaultValues...)
	for i, value := range signature.DefaultValues {
		expression, ok := value.(ast.Expression)
		if !ok || expression == nil {
			continue
		}
		constant, err := a.evaluateConstant(expression)
		if err != nil {
			continue
		}
		position := expression.Pos()
		switch value := constant.(type) {
		case int:
			bound.DefaultValues[i] = &ast.IntegerLiteral{Value: int64(value),
				BaseNode: ast.BaseNode{Token: token.NewToken(token.INT, strconv.Itoa(value), position)}}
		case bool:
			kind := token.FALSE
			if value {
				kind = token.TRUE
			}
			bound.DefaultValues[i] = &ast.BooleanLiteral{Value: value,
				BaseNode: ast.BaseNode{Token: token.NewToken(kind, strconv.FormatBool(value), position)}}
		case string:
			bound.DefaultValues[i] = &ast.StringLiteral{Value: value,
				BaseNode: ast.BaseNode{Token: token.NewToken(token.STRING, value, position)}}
		case float64:
			bound.DefaultValues[i] = &ast.FloatLiteral{Value: value,
				BaseNode: ast.BaseNode{Token: token.NewToken(token.FLOAT, strconv.FormatFloat(value, 'g', -1, 64), position)}}
		}
	}
	return &bound
}

// analyzeFunctionBody analyzes a regular function's body in a fresh scope, using the
// parameter and return types already resolved by registerFunctionSignature. Callers
// must skip forward declarations before invoking this.
func (a *Analyzer) analyzeFunctionBody(decl *ast.FunctionDecl, paramTypes []types.Type, returnType types.Type) {
	defer a.enterResultScope(returnType != nil && returnType != types.VOID)()
	// Analyze function body in new scope
	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(oldSymbols)
	a.retainScope(a.symbols, decl.Name.Value)
	defer func() { a.symbols = oldSymbols }()

	// Add parameters to function scope
	for i, param := range decl.Parameters {
		// Const and lazy parameters cannot provide writable storage.
		if param.IsConst || param.IsLazy {
			a.symbols.DefineParameter(param.Name.Value, paramTypes[i], param.Name.Token.Pos, true)
		} else {
			a.symbols.DefineParameter(param.Name.Value, paramTypes[i], param.Name.Token.Pos, false)
		}
	}

	// Add Result variable for functions (not procedures)
	if returnType != types.VOID {
		resultPos := decl.Name.Token.Pos
		if decl.End().Line != 0 {
			resultPos = blockEndStart(decl.End())
		}
		a.symbols.defineInternal("Result", returnType, resultPos)
		// Inside a unit, an empty implementation body deliberately leaves
		// Result at its default; do not hint "Result is never used" for it.
		// DWScript also skips the hint for overloaded functions with real
		// bodies (see fixture OverloadsPass/arrays).
		if decl.Body != nil && ((a.inUnitDecl && len(decl.Body.Statements) == 0) ||
			(decl.IsOverload && len(decl.Body.Statements) > 0)) {
			a.recordSymbolUsage("Result", resultPos)
		}
	}

	previousFunc := a.currentFunction
	a.currentFunction = decl
	defer func() { a.currentFunction = previousFunc }()
	defer a.emitUnusedWarningsForCurrentScope()

	if decl.PreConditions != nil {
		a.checkPreconditions(decl.PreConditions, decl.Name.Value)
	}
	if decl.Body != nil {
		a.analyzeRootBlock(decl.Body)
	}
	if decl.PostConditions != nil {
		a.checkPostconditions(decl.PostConditions, decl.Name.Value)
	}
}

// analyzeReturn analyzes a return statement
func (a *Analyzer) analyzeReturn(stmt *ast.ReturnStatement) {
	// Check if return statement is inside a finally block
	if a.inFinallyBlock {
		a.addError("return statement not allowed in finally block at %s", stmt.Token.Pos.String())
		return
	}

	if a.currentFunction == nil && !a.inLambda {
		a.addError("return statement outside of function at %s", stmt.Token.Pos.String())
		return
	}

	// Get expected return type
	var expectedType types.Type
	if a.currentFunction != nil {
		if a.currentFunction.ReturnType != nil {
			var err error
			expectedType, err = types.TypeFromString(getTypeExpressionName(a.currentFunction.ReturnType))
			if err != nil {
				// Error already reported during function declaration analysis
				return
			}
		} else {
			expectedType = types.VOID
		}
	} else if a.inLambda {
		// In lambda: analyze return value (type checking done during lambda inference)
		if stmt.ReturnValue != nil {
			a.analyzeExpression(stmt.ReturnValue)
		}
		return
	}

	// Validate return value type matches function return type
	if stmt.ReturnValue != nil {
		if expectedType == types.VOID {
			a.addError("procedure cannot return a value at %s", stmt.Token.Pos.String())
			return
		}
		returnType := a.analyzeExpressionWithExpectedType(stmt.ReturnValue, expectedType)
		if returnType != nil && !a.canAssign(returnType, expectedType) {
			a.addError("return type %s incompatible with function return type %s at %s",
				returnType.String(), expectedType.String(), stmt.Token.Pos.String())
		}
	} else {
		if expectedType != types.VOID {
			a.addError("function must return a value at %s", stmt.Token.Pos.String())
		}
	}
}

// ============================================================================
// Contract Analysis (Design by Contract)
// ============================================================================

// checkPreconditions validates precondition (require) expressions are boolean
func (a *Analyzer) checkPreconditions(preconds *ast.PreConditions, funcName string) {
	if preconds == nil {
		return
	}

	for _, cond := range preconds.Conditions {
		testType := a.analyzeExpression(cond.Test)
		if testType != nil && !isBooleanCompatible(testType) {
			a.addBooleanExpected(cond.Token.Pos)
		}
		a.warnConstantCondition(cond)

		// Message must be string (if present)
		if cond.Message != nil {
			msgType := a.analyzeExpression(cond.Message)
			if msgType != nil && msgType != types.STRING {
				// The message error re-uses the condition's anchor, not the
				// message expression's: contracts_types reports both at the
				// clause's first token.
				a.addStringExpected(cond.Token.Pos)
			}
		}
	}
}

// checkPostconditions validates postcondition (ensure) expressions are boolean
func (a *Analyzer) checkPostconditions(postconds *ast.PostConditions, funcName string) {
	if postconds == nil {
		return
	}

	for _, cond := range postconds.Conditions {
		testType := a.analyzeExpression(cond.Test)
		if testType != nil && !isBooleanCompatible(testType) {
			a.addBooleanExpected(cond.Token.Pos)
		}
		a.warnConstantCondition(cond)

		// Message must be string (if present)
		if cond.Message != nil {
			msgType := a.analyzeExpression(cond.Message)
			if msgType != nil && msgType != types.STRING {
				// The message error re-uses the condition's anchor, not the
				// message expression's: contracts_types reports both at the
				// clause's first token.
				a.addStringExpected(cond.Token.Pos)
			}
		}

		// Validate 'old' expressions (check undefined identifiers)
		a.validateOldExpressions(cond.Test, funcName)
	}
}

// warnConstantCondition follows DWScript's contract compilation order: the test
// is type checked first, its constancy is reported next, and its message is
// checked afterward. Even a constant test of the wrong type receives the warning.
func (a *Analyzer) warnConstantCondition(cond *ast.Condition) {
	if a.isConstantInstruction(cond.Test) {
		pos := cond.Token.Pos
		a.addWarning("Constant condition [line: %d, column: %d]", pos.Line, pos.Column)
	}
}

// validateOldExpressions recursively validates 'old' expression identifiers exist in scope
func (a *Analyzer) validateOldExpressions(expr ast.Expression, funcName string) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.OldExpression:
		if e.Identifier != nil {
			if _, ok := a.symbols.Resolve(e.Identifier.Value); !ok {
				a.addError("old() references undefined identifier '%s' in function '%s' at %s",
					e.Identifier.Value, funcName, e.Token.Pos.String())
			}
		}
	case *ast.BinaryExpression:
		a.validateOldExpressions(e.Left, funcName)
		a.validateOldExpressions(e.Right, funcName)
	case *ast.UnaryExpression:
		a.validateOldExpressions(e.Right, funcName)
	case *ast.GroupedExpression:
		a.validateOldExpressions(e.Expression, funcName)
	case *ast.CallExpression:
		for _, arg := range e.Arguments {
			a.validateOldExpressions(arg, funcName)
		}
	case *ast.IndexExpression:
		a.validateOldExpressions(e.Left, funcName)
		if e.Index != nil {
			a.validateOldExpressions(e.Index, funcName)
		}
	}
}

// addCallConventionHint reports a calling convention the port ignores. Methods and
// free routines share it so they share the sentence and the anchor: every other
// diagnostic renders its position as "[line: L, column: C]", and a hint that
// formats its own "at L:C" cannot match any fixture (SimpleScripts/call_conventions).
func (a *Analyzer) addCallConventionHint(decl *ast.FunctionDecl) {
	if decl == nil || decl.CallingConvention == "" {
		return
	}
	a.addHintAt(decl.CallingConventionPos, "Call convention %q is not supported and ignored [line: %d, column: %d]",
		decl.CallingConvention, decl.CallingConventionPos.Line, decl.CallingConventionPos.Column)
}

// isRefusedTypeExpression reports a type expression the parser refused and has
// already reported ("Type expected"); the analyzer does not report it again.
func isRefusedTypeExpression(expr ast.TypeExpression) bool {
	invalid, ok := expr.(*ast.InvalidTypeExpression)
	return ok && invalid != nil
}

// isConstantParameterDefault includes aggregate and nil constants, which do
// not draw the scalar constant-instruction hint but remain valid defaults.
func (a *Analyzer) isConstantParameterDefault(value ast.Expression) bool {
	if a.isConstantScalar(value) || a.isConstantInstruction(value) {
		return true
	}
	switch expr := value.(type) {
	case *ast.GroupedExpression:
		return a.isConstantParameterDefault(expr.Expression)
	case *ast.CallExpression:
		return a.isConstantParameterDefaultCall(expr)
	case *ast.ArrayLiteralExpression:
		for _, element := range expr.Elements {
			if !a.isConstantParameterDefault(element) {
				return false
			}
		}
		return true
	case *ast.RecordLiteralExpression:
		for _, field := range expr.Fields {
			if !a.isConstantParameterDefault(field.Value) {
				return false
			}
		}
		return true
	}
	_, err := a.evaluateConstant(value)
	return err == nil
}

func (a *Analyzer) isConstantParameterDefaultCall(expr *ast.CallExpression) bool {
	name, named := expr.Function.(*ast.Identifier)
	if !named {
		return false
	}
	if symbol, shadowed := a.symbols.Resolve(name.Value); shadowed {
		switch symbol.Type.(type) {
		case *types.FunctionType, *types.FunctionPointerType, *types.MethodPointerType:
			return false
		}
	} else if statelessBuiltins[ident.Normalize(name.Value)] {
		for _, arg := range expr.Arguments {
			if !a.isConstantParameterDefault(arg) {
				return false
			}
		}
		return true
	}
	if len(expr.Arguments) == 1 {
		if _, err := a.resolveType(name.Value); err == nil {
			return a.isConstantParameterDefault(expr.Arguments[0])
		}
	}
	return false
}

// analyzeParameterDefault validates the retained initializer without making a
// rejected or modified default optional. Errors remain after child diagnostics
// and before calls in subsequent declarations/instructions on the same line.
func (a *Analyzer) analyzeParameterDefault(param *ast.Parameter, paramType types.Type) interface{} {
	var defaultValue interface{}
	if param.DefaultValue != nil {
		mark := len(a.errors)
		defaultType := a.analyzeExpressionWithExpectedType(param.DefaultValue, paramType)
		if defaultType != nil && !a.errorsSince(mark) {
			switch {
			case !a.isConstantParameterDefault(param.DefaultValue):
				a.addStructuredError(&SemanticError{
					Type: ErrorInvalidOperation, Message: "Syntax Error: Constant expression expected",
					Pos: param.DefaultValue.End(), Severity: SeverityError, AfterChildren: true,
				})
			case !a.canAssign(defaultType, paramType):
				pos := param.DefaultValue.Pos()
				if param.DefaultValueSeparatorPos.Line > 0 && a.isConstantScalar(param.DefaultValue) {
					pos = param.DefaultValueSeparatorPos
				}
				err := NewIncompatibleTypesPairError(pos,
					semanticTypeNameForDiagnostic(paramType), semanticTypeNameForDiagnostic(defaultType))
				err.AfterChildren = true
				a.addStructuredError(err)
			case !param.IsLazy && !param.ByRef && !param.IsConst:
				defaultValue = param.DefaultValue
			}
		}
	}
	return defaultValue
}
