package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
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
	// Check for unsupported calling conventions and emit hints
	if decl.CallingConvention != "" {
		a.addHint("Call convention \"%s\" is not supported and ignored [line: %d, column: %d]",
			decl.CallingConvention, decl.CallingConventionPos.Line, decl.CallingConventionPos.Column)
	}

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

		// Optional parameters must come last, without modifiers
		if param.DefaultValue != nil {
			foundOptional = true
			if param.IsLazy || param.ByRef || param.IsConst {
				a.addError("optional parameter '%s' cannot have lazy, var, or const modifiers in function '%s' at %s",
					param.Name.Value, decl.Name.Value, param.Token.Pos.String())
				return nil, nil, false
			}
		} else if foundOptional {
			a.addError("required parameter '%s' cannot come after optional parameters in function '%s' at %s",
				param.Name.Value, decl.Name.Value, param.Token.Pos.String())
			return nil, nil, false
		}

		if param.Type == nil {
			a.addError("parameter '%s' missing type annotation in function '%s'",
				param.Name.Value, decl.Name.Value)
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

		// Validate default value type matches parameter type
		if param.DefaultValue != nil {
			defaultType := a.analyzeExpressionWithExpectedType(param.DefaultValue, paramType)
			if defaultType == nil {
				a.addError("invalid default value for parameter '%s' in function '%s'",
					param.Name.Value, decl.Name.Value)
				return nil, nil, false
			}
			if !paramType.Equals(defaultType) && !defaultType.Equals(paramType) {
				a.addError("default value type '%s' does not match parameter type '%s' for parameter '%s' in function '%s'",
					defaultType.String(), paramType.String(), param.Name.Value, decl.Name.Value)
				return nil, nil, false
			}
		}

		paramTypes = append(paramTypes, paramType)
		paramNames = append(paramNames, param.Name.Value)
		paramTypeNames = append(paramTypeNames, semanticDeclaredTypeName(param.Type, paramType))
		defaultValues = append(defaultValues, param.DefaultValue) // Store AST expression (may be nil)
		lazyParams = append(lazyParams, param.IsLazy)
		varParams = append(varParams, param.ByRef)
		constParams = append(constParams, param.IsConst)
		strictParams = append(strictParams, isStrictTypeAnnotation(param.Type))
	}

	// Determine return type
	if decl.ReturnType != nil {
		var err error
		returnType, err = a.resolveType(getTypeExpressionName(decl.ReturnType))
		if err != nil {
			a.addError("unknown return type '%s' in function '%s': %v",
				getTypeExpressionName(decl.ReturnType), decl.Name.Value, err)
			return nil, nil, false
		}
	} else {
		returnType = types.VOID
	}

	if decl.IsHelper {
		// Helper functions are fully analyzed here (signature + body); there is no
		// separate body pass for them.
		a.analyzeFunctionHelperDecl(decl, paramTypes, returnType)
		return nil, nil, false
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
	funcType.ParamTypeNames = paramTypeNames
	funcType.StrictParams = strictParams

	// Register function/overload with position info for error messages
	if err := a.symbols.DefineOverload(decl.Name.Value, funcType, decl.IsOverload, decl.IsForward, decl.Name.Token.Pos); err != nil {
		a.addError("Syntax Error: %s [line: %d, column: %d]", err.Error(), decl.Token.Pos.Line, decl.Token.Pos.Column)
		return nil, nil, false
	}

	return paramTypes, returnType, true
}

// analyzeFunctionBody analyzes a regular function's body in a fresh scope, using the
// parameter and return types already resolved by registerFunctionSignature. Callers
// must skip forward declarations before invoking this.
func (a *Analyzer) analyzeFunctionBody(decl *ast.FunctionDecl, paramTypes []types.Type, returnType types.Type) {
	// Analyze function body in new scope
	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(oldSymbols)
	defer func() { a.symbols = oldSymbols }()

	// Add parameters to function scope
	for i, param := range decl.Parameters {
		// Const parameters are read-only
		if param.IsConst {
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
		a.symbols.Define("Result", returnType, resultPos)
		// Inside a unit, an empty implementation body deliberately leaves
		// Result at its default; do not hint "Result is never used" for it.
		if a.inUnitDecl && decl.Body != nil && len(decl.Body.Statements) == 0 {
			a.recordSymbolUsage("Result", resultPos)
		}
	}

	previousFunc := a.currentFunction
	a.currentFunction = decl
	defer func() { a.currentFunction = previousFunc }()
	defer a.emitUnusedWarningsForCurrentScope()

	if decl.Body != nil {
		a.analyzeBlock(decl.Body)
	}

	// Analyze contract conditions (require/ensure)
	if decl.PreConditions != nil {
		a.checkPreconditions(decl.PreConditions, decl.Name.Value)
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
			a.addError("precondition must be boolean expression in function '%s', got %s at %s",
				funcName, testType.String(), cond.Token.Pos.String())
		}

		// Message must be string (if present)
		if cond.Message != nil {
			msgType := a.analyzeExpression(cond.Message)
			if msgType != nil && msgType != types.STRING {
				a.addError("condition message must be string expression in function '%s', got %s at %s",
					funcName, msgType.String(), cond.Token.Pos.String())
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
			a.addError("postcondition must be boolean expression in function '%s', got %s at %s",
				funcName, testType.String(), cond.Token.Pos.String())
		}

		// Message must be string (if present)
		if cond.Message != nil {
			msgType := a.analyzeExpression(cond.Message)
			if msgType != nil && msgType != types.STRING {
				a.addError("condition message must be string expression in function '%s', got %s at %s",
					funcName, msgType.String(), cond.Token.Pos.String())
			}
		}

		// Validate 'old' expressions (check undefined identifiers)
		a.validateOldExpressions(cond.Test, funcName)
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
