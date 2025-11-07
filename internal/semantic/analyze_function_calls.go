package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Expression Analysis
// ============================================================================
func (a *Analyzer) analyzeCallExpression(expr *ast.CallExpression) types.Type {
	// Handle member access expressions (method calls like obj.Method())
	if memberAccess, ok := expr.Function.(*ast.MemberAccessExpression); ok {
		// Analyze the member access to get the method type
		methodType := a.analyzeMemberAccessExpression(memberAccess)
		if methodType == nil {
			// Error already reported by analyzeMemberAccessExpression
			return nil
		}

		// Verify it's a function type
		funcType, ok := methodType.(*types.FunctionType)
		if !ok {
			a.addError("cannot call non-function type %s at %s",
				methodType.String(), expr.Token.Pos.String())
			return nil
		}

		// Validate argument count
		if len(expr.Arguments) != len(funcType.Parameters) {
			a.addError("method call expects %d argument(s), got %d at %s",
				len(funcType.Parameters), len(expr.Arguments), expr.Token.Pos.String())
		}

		// Validate argument types
		for i, arg := range expr.Arguments {
			if i >= len(funcType.Parameters) {
				break // Already reported count mismatch
			}

			// Task 9.2b: Validate var parameter receives an lvalue
			isVar := len(funcType.VarParams) > i && funcType.VarParams[i]
			if isVar && !a.isLValue(arg) {
				a.addError("var parameter %d requires a variable (identifier, array element, or field), got %s at %s",
					i+1, arg.String(), arg.Pos().String())
			}

			paramType := funcType.Parameters[i]
			argType := a.analyzeExpressionWithExpectedType(arg, paramType)
			if argType != nil && !a.canAssign(argType, paramType) {
				a.addError("argument %d has type %s, expected %s at %s",
					i+1, argType.String(), paramType.String(),
					expr.Token.Pos.String())
			}
		}

		return funcType.ReturnType
	}

	// Handle regular function calls (identifier-based)
	funcIdent, ok := expr.Function.(*ast.Identifier)
	if !ok {
		a.addError("function call must use identifier or member access at %s", expr.Token.Pos.String())
		return nil
	}

	// Look up function
	sym, ok := a.symbols.Resolve(funcIdent.Value)
	if !ok {
		// Check if it's a built-in function (using new dispatcher)
		if resultType, isBuiltin := a.analyzeBuiltinFunction(funcIdent.Value, expr.Arguments, expr); isBuiltin {
			return resultType
		}

		// Low built-in function

		// Assert built-in procedure
		if strings.EqualFold(funcIdent.Value, "Assert") {
			// Assert takes 1-2 arguments: Boolean condition and optional String message
			if len(expr.Arguments) < 1 || len(expr.Arguments) > 2 {
				a.addError("function 'Assert' expects 1-2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			// First argument must be Boolean
			condType := a.analyzeExpression(expr.Arguments[0])
			if condType != nil && condType != types.BOOLEAN {
				a.addError("function 'Assert' first argument must be Boolean, got %s at %s",
					condType.String(), expr.Token.Pos.String())
			}
			// If there's a second argument (message), it must be String
			if len(expr.Arguments) == 2 {
				msgType := a.analyzeExpression(expr.Arguments[1])
				if msgType != nil && msgType != types.STRING {
					a.addError("function 'Assert' second argument must be String, got %s at %s",
						msgType.String(), expr.Token.Pos.String())
				}
			}
			return types.VOID
		}

		// Insert built-in procedure
		if strings.EqualFold(funcIdent.Value, "Insert") {
			if len(expr.Arguments) != 3 {
				a.addError("function 'Insert' expects 3 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			sourceType := a.analyzeExpression(expr.Arguments[0])
			if sourceType != nil && sourceType != types.STRING {
				a.addError("function 'Insert' first argument must be String, got %s at %s",
					sourceType.String(), expr.Token.Pos.String())
			}
			if _, ok := expr.Arguments[1].(*ast.Identifier); !ok {
				a.addError("function 'Insert' second argument must be a variable at %s",
					expr.Token.Pos.String())
			} else {
				targetType := a.analyzeExpression(expr.Arguments[1])
				if targetType != nil && targetType != types.STRING {
					a.addError("function 'Insert' second argument must be String, got %s at %s",
						targetType.String(), expr.Token.Pos.String())
				}
			}
			posType := a.analyzeExpression(expr.Arguments[2])
			if posType != nil && posType != types.INTEGER {
				a.addError("function 'Insert' third argument must be Integer, got %s at %s",
					posType.String(), expr.Token.Pos.String())
			}
			return types.VOID
		}

		// Task 9.227: Higher-order functions for working with lambdas
		if strings.EqualFold(funcIdent.Value, "Map") {
			// Map(array, lambda) -> array
			if len(expr.Arguments) != 2 {
				a.addError("function 'Map' expects 2 arguments (array, lambda), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			arrayType := a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])

			// Verify first argument is an array
			if arrType, ok := arrayType.(*types.ArrayType); ok {
				return arrType // Return same array type
			}
			return types.VOID
		}

		if strings.EqualFold(funcIdent.Value, "Filter") {
			// Filter(array, predicate) -> array
			if len(expr.Arguments) != 2 {
				a.addError("function 'Filter' expects 2 arguments (array, predicate), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			arrayType := a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])

			// Verify first argument is an array
			if arrType, ok := arrayType.(*types.ArrayType); ok {
				return arrType // Return same array type
			}
			return types.VOID
		}

		if strings.EqualFold(funcIdent.Value, "Reduce") {
			// Reduce(array, lambda, initial) -> value
			if len(expr.Arguments) != 3 {
				a.addError("function 'Reduce' expects 3 arguments (array, lambda, initial), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])
			initialType := a.analyzeExpression(expr.Arguments[2])

			// Return type is the same as initial value type
			return initialType
		}

		if strings.EqualFold(funcIdent.Value, "ForEach") {
			// ForEach(array, lambda) -> void
			if len(expr.Arguments) != 2 {
				a.addError("function 'ForEach' expects 2 arguments (array, lambda), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])

			return types.VOID
		}

		// Task 9.95-9.97: Current date/time functions

		// Allow calling methods within the current class without explicit Self
		if a.currentClass != nil {
			if methodType, found := a.currentClass.GetMethod(funcIdent.Value); found {
				if len(expr.Arguments) != len(methodType.Parameters) {
					a.addError("method '%s' expects %d arguments, got %d at %s",
						funcIdent.Value, len(methodType.Parameters), len(expr.Arguments), expr.Token.Pos.String())
					return methodType.ReturnType
				}
				for i, arg := range expr.Arguments {
					argType := a.analyzeExpression(arg)
					expectedType := methodType.Parameters[i]
					if argType != nil && !a.canAssign(argType, expectedType) {
						a.addError("argument %d to method '%s' has type %s, expected %s at %s",
							i+1, funcIdent.Value, argType.String(), expectedType.String(), expr.Token.Pos.String())
					}
				}
				return methodType.ReturnType
			}
		}

		// Task 9.232: Variant introspection functions

		// Task 9.114: GetStackTrace() built-in function
		if strings.EqualFold(funcIdent.Value, "GetStackTrace") {
			if len(expr.Arguments) != 0 {
				a.addError("function 'GetStackTrace' expects 0 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
			}
			// Returns String
			return types.STRING
		}

		// Task 9.116: GetCallStack() built-in function
		if strings.EqualFold(funcIdent.Value, "GetCallStack") {
			if len(expr.Arguments) != 0 {
				a.addError("function 'GetCallStack' expects 0 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
			}
			// Returns dynamic array of records
			// Each record has: FunctionName: String, FileName: String, Line: Integer, Column: Integer
			// For simplicity in semantic analysis, we return a generic dynamic array type
			return types.NewDynamicArrayType(types.VARIANT)
		}

		a.addError("undefined function '%s' at %s", funcIdent.Value, expr.Token.Pos.String())
		return nil
	}

	// Task 9.65-9.66: Check if this is an overloaded function
	// If so, resolve the overload set to select the best match
	var funcType *types.FunctionType
	if sym.IsOverloadSet {
		// Get all overload candidates
		candidates := a.symbols.GetOverloadSet(funcIdent.Value)
		if candidates == nil || len(candidates) == 0 {
			a.addError("no overload candidates found for '%s' at %s", funcIdent.Value, expr.Token.Pos.String())
			return nil
		}

		// Analyze argument types first
		argTypes := make([]types.Type, len(expr.Arguments))
		for i, arg := range expr.Arguments {
			argType := a.analyzeExpression(arg)
			if argType == nil {
				return nil // Error already reported
			}
			argTypes[i] = argType
		}

		// Resolve overload based on argument types
		selected, err := ResolveOverload(candidates, argTypes)
		if err != nil {
			// Task 9.63: Provide DWScript-compatible error message for failed overload resolution
			a.addError("Syntax Error: There is no overloaded version of \"%s\" that can be called with these arguments [line: %d, column: %d]",
				funcIdent.Value, expr.Token.Pos.Line, expr.Token.Pos.Column)
			return nil
		}

		// Use the selected overload's function type
		var ok bool
		funcType, ok = selected.Type.(*types.FunctionType)
		if !ok {
			a.addError("selected overload for '%s' is not a function type at %s", funcIdent.Value, expr.Token.Pos.String())
			return nil
		}
	} else {
		// Task 9.162: Check if it's a function pointer type first
		if funcPtrType := a.analyzeFunctionPointerCall(expr, sym.Type); funcPtrType != nil {
			return funcPtrType
		}

		// Check that symbol is a function
		var ok bool
		funcType, ok = sym.Type.(*types.FunctionType)
		if !ok {
			a.addError("'%s' is not a function at %s", funcIdent.Value, expr.Token.Pos.String())
			return nil
		}
	}

	// Task 9.1: Check argument count with optional parameters support
	// Count required parameters (those without defaults)
	requiredParams := 0
	for _, defaultVal := range funcType.DefaultValues {
		if defaultVal == nil {
			requiredParams++
		}
	}

	// Check argument count is within valid range
	if len(expr.Arguments) < requiredParams {
		// Use more precise error message based on whether function has optional parameters
		if requiredParams == len(funcType.Parameters) {
			// All parameters are required
			a.addError("function '%s' expects %d arguments, got %d at %s",
				funcIdent.Value, requiredParams, len(expr.Arguments),
				expr.Token.Pos.String())
		} else {
			// Function has optional parameters
			a.addError("function '%s' expects at least %d arguments, got %d at %s",
				funcIdent.Value, requiredParams, len(expr.Arguments),
				expr.Token.Pos.String())
		}
		return nil
	}
	if len(expr.Arguments) > len(funcType.Parameters) {
		a.addError("function '%s' expects at most %d arguments, got %d at %s",
			funcIdent.Value, len(funcType.Parameters), len(expr.Arguments),
			expr.Token.Pos.String())
		return nil
	}

	// Check argument types
	// Task 9.137: Handle lazy parameters - validate expression type without evaluating
	// Task 9.2b: Handle var parameters - validate that argument is an lvalue
	for i, arg := range expr.Arguments {
		expectedType := funcType.Parameters[i]

		// Check if this parameter is lazy
		isLazy := len(funcType.LazyParams) > i && funcType.LazyParams[i]

		// Check if this parameter is var (by-reference)
		isVar := len(funcType.VarParams) > i && funcType.VarParams[i]

		// Task 9.2b: Validate var parameter receives an lvalue
		if isVar && !a.isLValue(arg) {
			a.addError("var parameter %d to function '%s' requires a variable (identifier, array element, or field), got %s at %s",
				i+1, funcIdent.Value, arg.String(), arg.Pos().String())
		}

		if isLazy {
			// For lazy parameters, check expression type but don't evaluate
			// The expression will be passed as-is to the interpreter for deferred evaluation
			argType := a.analyzeExpressionWithExpectedType(arg, expectedType)
			if argType != nil && !a.canAssign(argType, expectedType) {
				a.addError("lazy argument %d to function '%s' has type %s, expected %s at %s",
					i+1, funcIdent.Value, argType.String(), expectedType.String(),
					expr.Token.Pos.String())
			}
		} else {
			// Regular parameter: validate type normally
			argType := a.analyzeExpressionWithExpectedType(arg, expectedType)
			if argType != nil && !a.canAssign(argType, expectedType) {
				a.addError("argument %d to function '%s' has type %s, expected %s at %s",
					i+1, funcIdent.Value, argType.String(), expectedType.String(),
					expr.Token.Pos.String())
			}
		}
	}

	return funcType.ReturnType
}

// analyzeOldExpression analyzes an 'old' expression in a postcondition
// Task 9.143: Return the type of the referenced identifier
func (a *Analyzer) analyzeOldExpression(expr *ast.OldExpression) types.Type {
	if expr.Identifier == nil {
		return nil
	}

	// Look up the identifier in the symbol table
	sym, ok := a.symbols.Resolve(expr.Identifier.Value)
	if !ok {
		// Error already reported in validateOldExpressions
		return nil
	}

	// Return the type of the identifier
	return sym.Type
}
