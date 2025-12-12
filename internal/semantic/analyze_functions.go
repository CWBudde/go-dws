package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Function Analysis
// ============================================================================

// analyzeFunctionDecl analyzes a function declaration
func (a *Analyzer) analyzeFunctionDecl(decl *ast.FunctionDecl) {
	// Check if this is a method implementation (has ClassName)
	if decl.ClassName != nil {
		a.analyzeMethodImplementation(decl)
		return
	}

	// Check for unsupported calling conventions and emit hints
	if decl.CallingConvention != "" {
		a.addHint("Call convention \"%s\" is not supported and ignored [line: %d, column: %d]",
			decl.CallingConvention, decl.CallingConventionPos.Line, decl.CallingConventionPos.Column)
	}

	// Regular function (not method): resolve parameter and return types
	paramTypes := make([]types.Type, 0, len(decl.Parameters))
	paramNames := make([]string, 0, len(decl.Parameters))
	defaultValues := make([]interface{}, 0, len(decl.Parameters))
	lazyParams := make([]bool, 0, len(decl.Parameters))
	varParams := make([]bool, 0, len(decl.Parameters))
	constParams := make([]bool, 0, len(decl.Parameters))
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
			return
		}

		// Optional parameters must come last, without modifiers
		if param.DefaultValue != nil {
			foundOptional = true
			if param.IsLazy || param.ByRef || param.IsConst {
				a.addError("optional parameter '%s' cannot have lazy, var, or const modifiers in function '%s' at %s",
					param.Name.Value, decl.Name.Value, param.Token.Pos.String())
				return
			}
		} else if foundOptional {
			a.addError("required parameter '%s' cannot come after optional parameters in function '%s' at %s",
				param.Name.Value, decl.Name.Value, param.Token.Pos.String())
			return
		}

		if param.Type == nil {
			a.addError("parameter '%s' missing type annotation in function '%s'",
				param.Name.Value, decl.Name.Value)
			return
		}
		paramType, err := a.resolveType(getTypeExpressionName(param.Type))
		if err != nil {
			a.addError("unknown parameter type '%s' in function '%s': %v",
				getTypeExpressionName(param.Type), decl.Name.Value, err)
			return
		}

		// Validate default value type matches parameter type
		if param.DefaultValue != nil {
			defaultType := a.analyzeExpressionWithExpectedType(param.DefaultValue, paramType)
			if defaultType == nil {
				a.addError("invalid default value for parameter '%s' in function '%s'",
					param.Name.Value, decl.Name.Value)
				return
			}
			if !paramType.Equals(defaultType) && !defaultType.Equals(paramType) {
				a.addError("default value type '%s' does not match parameter type '%s' for parameter '%s' in function '%s'",
					defaultType.String(), paramType.String(), param.Name.Value, decl.Name.Value)
				return
			}
		}

		paramTypes = append(paramTypes, paramType)
		paramNames = append(paramNames, param.Name.Value)
		defaultValues = append(defaultValues, param.DefaultValue) // Store AST expression (may be nil)
		lazyParams = append(lazyParams, param.IsLazy)
		varParams = append(varParams, param.ByRef)
		constParams = append(constParams, param.IsConst)
	}

	// Determine return type
	var returnType types.Type
	if decl.ReturnType != nil {
		var err error
		returnType, err = a.resolveType(getTypeExpressionName(decl.ReturnType))
		if err != nil {
			a.addError("unknown return type '%s' in function '%s': %v",
				getTypeExpressionName(decl.ReturnType), decl.Name.Value, err)
			return
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

	// Register function/overload with position info for error messages
	if err := a.symbols.DefineOverload(decl.Name.Value, funcType, decl.IsOverload, decl.IsForward, decl.Name.Token.Pos); err != nil {
		a.addError("Syntax Error: %s [line: %d, column: %d]", err.Error(), decl.Token.Pos.Line, decl.Token.Pos.Column)
		return
	}

	// Skip body analysis for forward declarations
	if decl.IsForward {
		return
	}

	// Analyze function body in new scope
	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(oldSymbols)
	defer func() { a.symbols = oldSymbols }()

	// Add parameters to function scope
	for i, param := range decl.Parameters {
		// Const parameters are read-only
		if param.IsConst {
			a.symbols.DefineReadOnly(param.Name.Value, paramTypes[i], param.Name.Token.Pos)
		} else {
			a.symbols.Define(param.Name.Value, paramTypes[i], param.Name.Token.Pos)
		}
	}

	// Add Result variable for functions (not procedures)
	if returnType != types.VOID {
		a.symbols.Define("Result", returnType, decl.Name.Token.Pos)
	}

	previousFunc := a.currentFunction
	a.currentFunction = decl
	defer func() { a.currentFunction = previousFunc }()

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
