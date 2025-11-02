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

	// This is a regular function (not a method implementation)
	// Convert parameter types and return type
	// Task 9.102: Use resolveType to support user-defined types like subranges
	// Task 9.136: Extract parameter metadata (names, lazy, var, const flags)
	paramTypes := make([]types.Type, 0, len(decl.Parameters))
	paramNames := make([]string, 0, len(decl.Parameters))
	lazyParams := make([]bool, 0, len(decl.Parameters))
	varParams := make([]bool, 0, len(decl.Parameters))
	constParams := make([]bool, 0, len(decl.Parameters))

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

		if param.Type == nil {
			a.addError("parameter '%s' missing type annotation in function '%s'",
				param.Name.Value, decl.Name.Value)
			return
		}
		paramType, err := a.resolveType(param.Type.Name)
		if err != nil {
			a.addError("unknown parameter type '%s' in function '%s': %v",
				param.Type.Name, decl.Name.Value, err)
			return
		}
		paramTypes = append(paramTypes, paramType)
		paramNames = append(paramNames, param.Name.Value)
		lazyParams = append(lazyParams, param.IsLazy)
		varParams = append(varParams, param.ByRef)
		constParams = append(constParams, param.IsConst)
	}

	// Determine return type
	var returnType types.Type
	if decl.ReturnType != nil {
		var err error
		returnType, err = a.resolveType(decl.ReturnType.Name)
		if err != nil {
			a.addError("unknown return type '%s' in function '%s': %v",
				decl.ReturnType.Name, decl.Name.Value, err)
			return
		}
	} else {
		returnType = types.VOID
	}

	// Check if function is already declared
	if a.symbols.IsDeclaredInCurrentScope(decl.Name.Value) {
		a.addError("function '%s' already declared", decl.Name.Value)
		return
	}

	// Create function type with metadata and add to symbol table
	// Task 9.136: Use NewFunctionTypeWithMetadata to include lazy, var, and const parameter info
	funcType := types.NewFunctionTypeWithMetadata(paramTypes, paramNames, lazyParams, varParams, constParams, returnType)
	a.symbols.DefineFunction(decl.Name.Value, funcType)

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
			expectedType, err = types.TypeFromString(a.currentFunction.ReturnType.Name)
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

		returnType := a.analyzeExpression(stmt.ReturnValue)
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
