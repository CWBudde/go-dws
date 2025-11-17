package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
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

	// This is a regular function (not a method implementation)
	// Convert parameter types and return type
	// Task 9.102: Use resolveType to support user-defined types like subranges
	// Task 9.136: Extract parameter metadata (names, lazy, var, const flags)
	// Task 9.1: Extract default value expressions for optional parameters
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

		// Task 9.1b: Validate optional parameter rules
		if param.DefaultValue != nil {
			// This is an optional parameter
			foundOptional = true

			// Optional parameters cannot have modifiers
			if param.IsLazy || param.ByRef || param.IsConst {
				a.addError("optional parameter '%s' cannot have lazy, var, or const modifiers in function '%s' at %s",
					param.Name.Value, decl.Name.Value, param.Token.Pos.String())
				return
			}
		} else if foundOptional {
			// Required parameter after optional parameter
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

		// Task 9.1b: Type-check default value if present
		if param.DefaultValue != nil {
			// Analyze default value expression to get its type
			// Use context-aware analysis so nil can infer the correct type
			defaultType := a.analyzeExpressionWithExpectedType(param.DefaultValue, paramType)
			if defaultType == nil {
				a.addError("invalid default value for parameter '%s' in function '%s'",
					param.Name.Value, decl.Name.Value)
				return
			}

			// Check if default value type matches parameter type
			// Use Equals for type checking (allows implicit conversions handled elsewhere)
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

	// Create function type with metadata
	// Task 9.136: Use NewFunctionTypeWithMetadata to include lazy, var, and const parameter info
	// Task 9.1: Include default value expressions for optional parameters
	// Task 9.21.4.3: Detect variadic parameters (last parameter is array type)
	var funcType *types.FunctionType
	if len(paramTypes) > 0 {
		// Check if last parameter is a dynamic array (variadic)
		lastParamType := paramTypes[len(paramTypes)-1]
		if arrayType, ok := lastParamType.(*types.ArrayType); ok && arrayType.IsDynamic() {
			// This is a variadic parameter
			variadicType := arrayType.ElementType
			funcType = types.NewVariadicFunctionTypeWithMetadata(
				paramTypes, paramNames, defaultValues, lazyParams, varParams, constParams,
				variadicType, returnType,
			)
		} else {
			// Regular (non-variadic) function
			funcType = types.NewFunctionTypeWithMetadata(
				paramTypes, paramNames, defaultValues, lazyParams, varParams, constParams, returnType,
			)
		}
	} else {
		// No parameters - create regular function type
		funcType = types.NewFunctionTypeWithMetadata(
			paramTypes, paramNames, defaultValues, lazyParams, varParams, constParams, returnType,
		)
	}

	// Task 9.58-9.62: Use DefineOverload to handle overload validation
	// Task 9.60: Pass IsForward to handle forward declarations
	// DefineOverload handles both single functions and overload sets
	// Task 9.63: Include position information in error messages
	if err := a.symbols.DefineOverload(decl.Name.Value, funcType, decl.IsOverload, decl.IsForward); err != nil {
		a.addError("Syntax Error: %s [line: %d, column: %d]", err.Error(), decl.Token.Pos.Line, decl.Token.Pos.Column)
		return
	}

	// Task 9.60: Skip body analysis for forward declarations (they have no body)
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
			a.symbols.DefineReadOnly(param.Name.Value, paramTypes[i])
		} else {
			a.symbols.Define(param.Name.Value, paramTypes[i])
		}
	}

	// For functions (not procedures), add Result variable
	if returnType != types.VOID {
		a.symbols.Define("Result", returnType)
		// Note: In Pascal, you can assign to the function name, but we don't add it
		// as a separate variable to avoid shadowing the function itself.
		// Assignments to the function name should be treated as assignments to Result.
	}

	// Set current function for return statement checking
	previousFunc := a.currentFunction
	a.currentFunction = decl
	defer func() { a.currentFunction = previousFunc }()

	// Analyze function body
	if decl.Body != nil {
		a.analyzeBlock(decl.Body)
	}

	// Task 9.140-9.143: Analyze contract conditions (require/ensure)
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
		// In lambda context - just analyze the return value
		// The type checking will be done during lambda return type inference
		// Task 9.19: Use context-aware analysis for type inference (e.g., nested lambdas)
		if stmt.ReturnValue != nil {
			a.analyzeExpression(stmt.ReturnValue)
		}
		return
	}

	// Check return value
	if stmt.ReturnValue != nil {
		if expectedType == types.VOID {
			a.addError("procedure cannot return a value at %s", stmt.Token.Pos.String())
			return
		}

		// Task 9.19: Use context-aware analysis for type inference (e.g., lambdas)
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

// checkPreconditions validates precondition (require) expressions
// Task 9.140: Validate precondition expressions are boolean type
func (a *Analyzer) checkPreconditions(preconds *ast.PreConditions, funcName string) {
	if preconds == nil {
		return
	}

	for _, cond := range preconds.Conditions {
		// Task 9.140: Validate test expression is boolean
		testType := a.analyzeExpression(cond.Test)
		if testType != nil && !isBooleanCompatible(testType) {
			a.addError("precondition must be boolean expression in function '%s', got %s at %s",
				funcName, testType.String(), cond.Token.Pos.String())
		}

		// Task 9.142: Validate message expression is string (if present)
		if cond.Message != nil {
			msgType := a.analyzeExpression(cond.Message)
			if msgType != nil && msgType != types.STRING {
				a.addError("condition message must be string expression in function '%s', got %s at %s",
					funcName, msgType.String(), cond.Token.Pos.String())
			}
		}
	}
}

// checkPostconditions validates postcondition (ensure) expressions
// Task 9.141: Validate postcondition expressions are boolean type
// Task 9.143: Validate 'old' keyword usage
func (a *Analyzer) checkPostconditions(postconds *ast.PostConditions, funcName string) {
	if postconds == nil {
		return
	}

	for _, cond := range postconds.Conditions {
		// Task 9.141: Validate test expression is boolean
		testType := a.analyzeExpression(cond.Test)
		if testType != nil && !isBooleanCompatible(testType) {
			a.addError("postcondition must be boolean expression in function '%s', got %s at %s",
				funcName, testType.String(), cond.Token.Pos.String())
		}

		// Task 9.142: Validate message expression is string (if present)
		if cond.Message != nil {
			msgType := a.analyzeExpression(cond.Message)
			if msgType != nil && msgType != types.STRING {
				a.addError("condition message must be string expression in function '%s', got %s at %s",
					funcName, msgType.String(), cond.Token.Pos.String())
			}
		}

		// Task 9.143: Validate 'old' expressions (check undefined identifiers)
		a.validateOldExpressions(cond.Test, funcName)
	}
}

// validateOldExpressions recursively validates 'old' keyword usage
// Task 9.143: Ensure referenced identifiers exist in scope
func (a *Analyzer) validateOldExpressions(expr ast.Expression, funcName string) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.OldExpression:
		// Validate that the referenced identifier is defined
		if e.Identifier != nil {
			if _, ok := a.symbols.Resolve(e.Identifier.Value); !ok {
				a.addError("old() references undefined identifier '%s' in function '%s' at %s",
					e.Identifier.Value, funcName, e.Token.Pos.String())
			}
		}

	case *ast.BinaryExpression:
		// Recursively check both sides
		a.validateOldExpressions(e.Left, funcName)
		a.validateOldExpressions(e.Right, funcName)

	case *ast.UnaryExpression:
		// Recursively check operand
		a.validateOldExpressions(e.Right, funcName)

	case *ast.GroupedExpression:
		// Recursively check grouped expression
		a.validateOldExpressions(e.Expression, funcName)

	case *ast.CallExpression:
		// Check all arguments
		for _, arg := range e.Arguments {
			a.validateOldExpressions(arg, funcName)
		}

	case *ast.IndexExpression:
		// Check index expression
		a.validateOldExpressions(e.Left, funcName)
		if e.Index != nil {
			a.validateOldExpressions(e.Index, funcName)
		}

	// For literals and identifiers, no further validation needed
	default:
		// No old expressions in other node types
	}
}
