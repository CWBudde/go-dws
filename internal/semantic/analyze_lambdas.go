package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Lambda Expression Analysis
// ============================================================================

// analyzeLambdaExpression analyzes a lambda expression.
// It validates parameter types, creates a new scope for the lambda body,
// analyzes the body in that scope, and infers the return type if not specified.
//
// Examples:
//   - lambda(x: Integer): Integer begin Result := x * 2; end
//   - lambda(x: Integer) => x * 2 (shorthand, desugared to block)
//   - lambda(a, b: Integer): Integer begin Result := a + b; end
//
// Task 9.216: Analysis steps:
//   - Validate all parameter types exist (require explicit types for now)
//   - Check for duplicate parameter names
//   - Create new scope for lambda parameters
//   - Add parameters to the scope
//   - Add Result variable if return type specified
//   - Analyze lambda body in nested scope
//   - Infer return type from body if not specified
//   - Create and return a FunctionPointerType matching the signature
//   - Set the type annotation on the expression
func (a *Analyzer) analyzeLambdaExpression(expr *ast.LambdaExpression) types.Type {
	if expr == nil {
		return nil
	}

	// Task 9.216: Check for duplicate parameter names
	paramNames := make(map[string]bool)
	for _, param := range expr.Parameters {
		if paramNames[param.Name.Value] {
			a.addError("duplicate parameter name '%s' in lambda at %s",
				param.Name.Value, param.Name.Token.Pos.String())
			return nil
		}
		paramNames[param.Name.Value] = true
	}

	// Task 9.216/9.218: Validate and infer parameter types
	paramTypes := make([]types.Type, 0, len(expr.Parameters))
	hasUntypedParams := false

	for _, param := range expr.Parameters {
		if param.Type == nil {
			hasUntypedParams = true
			// Task 9.218: Try to infer parameter type from context
			// For now, mark as needing inference
			paramTypes = append(paramTypes, nil)
		} else {
			paramType, err := a.resolveType(param.Type.Name)
			if err != nil {
				a.addError("unknown parameter type '%s' in lambda at %s",
					param.Type.Name, param.Type.Token.Pos.String())
				return nil
			}
			paramTypes = append(paramTypes, paramType)
		}
	}

	// Task 9.218: If we have untyped parameters, attempt type inference
	if hasUntypedParams {
		// Try to infer from context
		// This would require passing expected type from parent expression
		// For now, report an error
		a.addError("lambda parameter type inference not fully implemented at %s - please provide explicit types",
			expr.Token.Pos.String())
		return nil
	}

	// Task 9.216: Create new scope for lambda body
	oldSymbols := a.symbols
	lambdaScope := NewEnclosedSymbolTable(oldSymbols)
	a.symbols = lambdaScope
	defer func() { a.symbols = oldSymbols }()

	// Add parameters to lambda scope
	for i, param := range expr.Parameters {
		a.symbols.Define(param.Name.Value, paramTypes[i])
	}

	// Track that we're in a lambda to allow return statements
	// Note: Shorthand lambdas desugar to ReturnStatement, which needs lambda context
	previousInLambda := a.inLambda
	a.inLambda = true
	defer func() { a.inLambda = previousInLambda }()

	// Task 9.216: Determine or infer return type
	var returnType types.Type
	if expr.ReturnType != nil {
		// Explicit return type specified
		var err error
		returnType, err = a.resolveType(expr.ReturnType.Name)
		if err != nil {
			a.addError("unknown return type '%s' in lambda at %s",
				expr.ReturnType.Name, expr.ReturnType.Token.Pos.String())
			return nil
		}
		// Add Result variable for lambdas with explicit return type
		if returnType != types.VOID {
			a.symbols.Define("Result", returnType)
		}
	} else {
		// No explicit return type - infer from body
		// Note: Inference analyzes the body to determine the type
		returnType = a.inferReturnTypeFromBody(expr.Body)
		if returnType == nil {
			a.addError("cannot infer return type for lambda at %s",
				expr.Token.Pos.String())
			return nil
		}
		// Add Result variable now that we know the type
		// Note: The body was already analyzed during inference, so we don't need to analyze it again
		if returnType != types.VOID {
			a.symbols.Define("Result", returnType)
		}
	}

	// Analyze lambda body (only if we had an explicit return type)
	// If return type was inferred, the body was already analyzed during inference
	if expr.ReturnType != nil && expr.Body != nil {
		a.analyzeBlock(expr.Body)
	}

	// Task 9.217: Perform closure capture analysis
	// Identify variables from outer scopes used in the lambda
	capturedVars := a.analyzeCapturedVariables(expr.Body, lambdaScope, oldSymbols)
	expr.CapturedVars = capturedVars

	// Task 9.216: Create function pointer type matching the lambda signature
	// For procedures (no return type or VOID), pass nil as return type
	var funcPtrReturnType types.Type
	if returnType != nil && returnType != types.VOID {
		funcPtrReturnType = returnType
	}
	funcPtrType := types.NewFunctionPointerType(paramTypes, funcPtrReturnType)

	// Task 9.216: Set the type annotation on the expression
	// Create a TypeAnnotation for the AST node
	typeAnnotation := &ast.TypeAnnotation{
		Name: fmt.Sprintf("lambda%s", funcPtrType.String()),
	}
	expr.Type = typeAnnotation

	return funcPtrType
}

// analyzeLambdaExpressionWithContext analyzes a lambda expression with expected function pointer type.
// This enables parameter type inference from the context where the lambda is used.
//
// Task 9.19: Parameter type inference from context.
// Examples:
//   - var f: TFunc := lambda(x) => x * 2  (x type inferred from TFunc)
//   - Apply(5, lambda(n) => n * 2)        (n type inferred from Apply's signature)
//   - return lambda(x) => x * 2           (x type inferred from function return type)
//
// The function validates that:
//   - Parameter count matches expected type
//   - Explicit parameter types are compatible with expected types
//   - Inferred parameter types are used for untyped parameters
func (a *Analyzer) analyzeLambdaExpressionWithContext(expr *ast.LambdaExpression, expectedFuncType *types.FunctionPointerType) types.Type {
	if expr == nil || expectedFuncType == nil {
		return nil
	}

	// Task 9.19.3: Validate parameter count compatibility
	if len(expr.Parameters) != len(expectedFuncType.Parameters) {
		a.addError("lambda has %d parameters but expected function type has %d at %s",
			len(expr.Parameters), len(expectedFuncType.Parameters), expr.Token.Pos.String())
		return nil
	}

	// Task 9.216: Check for duplicate parameter names
	paramNames := make(map[string]bool)
	for _, param := range expr.Parameters {
		if paramNames[param.Name.Value] {
			a.addError("duplicate parameter name '%s' in lambda at %s",
				param.Name.Value, param.Name.Token.Pos.String())
			return nil
		}
		paramNames[param.Name.Value] = true
	}

	// Task 9.19.3: Infer types for untyped parameters, validate explicit types
	paramTypes := make([]types.Type, 0, len(expr.Parameters))

	for i, param := range expr.Parameters {
		expectedParamType := expectedFuncType.Parameters[i]

		if param.Type == nil {
			// No explicit type - infer from expected type
			paramTypes = append(paramTypes, expectedParamType)

			// Create TypeAnnotation and attach to parameter for later use
			// This allows the interpreter to know the parameter type
			param.Type = &ast.TypeAnnotation{
				Token: param.Token,
				Name:  expectedParamType.String(),
			}
		} else {
			// Explicit type provided - validate it's compatible with expected type
			paramType, err := a.resolveType(param.Type.Name)
			if err != nil {
				a.addError("unknown parameter type '%s' in lambda at %s",
					param.Type.Name, param.Type.Token.Pos.String())
				return nil
			}

			// Check compatibility with expected type
			if !a.canAssign(paramType, expectedParamType) {
				a.addError("parameter '%s' has type %s but expected type requires %s at %s",
					param.Name.Value, paramType.String(), expectedParamType.String(),
					param.Name.Token.Pos.String())
				return nil
			}

			paramTypes = append(paramTypes, paramType)
		}
	}

	// Now we have all parameter types (either inferred or explicit)
	// Continue with standard lambda analysis

	// Task 9.216: Create new scope for lambda body
	oldSymbols := a.symbols
	lambdaScope := NewEnclosedSymbolTable(oldSymbols)
	a.symbols = lambdaScope
	defer func() { a.symbols = oldSymbols }()

	// Add parameters to lambda scope
	for i, param := range expr.Parameters {
		a.symbols.Define(param.Name.Value, paramTypes[i])
	}

	// Track that we're in a lambda to allow return statements
	previousInLambda := a.inLambda
	a.inLambda = true
	defer func() { a.inLambda = previousInLambda }()

	// Task 9.216: Determine or infer return type
	var returnType types.Type
	if expr.ReturnType != nil {
		// Explicit return type specified
		var err error
		returnType, err = a.resolveType(expr.ReturnType.Name)
		if err != nil {
			a.addError("unknown return type '%s' in lambda at %s",
				expr.ReturnType.Name, expr.ReturnType.Token.Pos.String())
			return nil
		}

		// Check compatibility with expected return type if available
		if expectedFuncType.ReturnType != nil {
			if !a.canAssign(returnType, expectedFuncType.ReturnType) {
				a.addError("lambda return type %s incompatible with expected return type %s at %s",
					returnType.String(), expectedFuncType.ReturnType.String(),
					expr.Token.Pos.String())
				return nil
			}
		}

		// Add Result variable for lambdas with explicit return type
		if returnType != types.VOID {
			a.symbols.Define("Result", returnType)
		}
	} else {
		// No explicit return type - infer from body
		returnType = a.inferReturnTypeFromBody(expr.Body)
		if returnType == nil {
			a.addError("cannot infer return type for lambda at %s",
				expr.Token.Pos.String())
			return nil
		}

		// Check compatibility with expected return type if available
		if expectedFuncType.ReturnType != nil {
			if !a.canAssign(returnType, expectedFuncType.ReturnType) {
				a.addError("inferred lambda return type %s incompatible with expected return type %s at %s",
					returnType.String(), expectedFuncType.ReturnType.String(),
					expr.Token.Pos.String())
				return nil
			}
		}

		// Add Result variable now that we know the type
		if returnType != types.VOID {
			a.symbols.Define("Result", returnType)
		}
	}

	// Analyze lambda body (only if we had an explicit return type)
	// If return type was inferred, the body was already analyzed during inference
	if expr.ReturnType != nil && expr.Body != nil {
		a.analyzeBlock(expr.Body)
	}

	// Task 9.217: Perform closure capture analysis
	capturedVars := a.analyzeCapturedVariables(expr.Body, lambdaScope, oldSymbols)
	expr.CapturedVars = capturedVars

	// Create function pointer type matching the lambda signature
	var funcPtrReturnType types.Type
	if returnType != nil && returnType != types.VOID {
		funcPtrReturnType = returnType
	}
	funcPtrType := types.NewFunctionPointerType(paramTypes, funcPtrReturnType)

	// Set the type annotation on the expression
	typeAnnotation := &ast.TypeAnnotation{
		Name: fmt.Sprintf("lambda%s", funcPtrType.String()),
	}
	expr.Type = typeAnnotation

	return funcPtrType
}

// inferReturnTypeFromBody attempts to infer the return type from a lambda body.
// It walks through the body statements looking for return statements and Result assignments.
//
// Task 9.216: Inference strategy:
//   - If body contains explicit return statements, infer from their expressions
//   - If body contains Result assignments, infer from the assigned values
//   - If shorthand lambda (single expression), infer from that expression
//   - If no returns/Result assignments found, infer as procedure (VOID)
//   - If multiple return statements with conflicting types, report error
func (a *Analyzer) inferReturnTypeFromBody(body *ast.BlockStatement) types.Type {
	if body == nil || len(body.Statements) == 0 {
		// Empty body - treat as procedure
		return types.VOID
	}

	// Collect return types from all return statements
	var returnTypes []types.Type

	// Walk through statements looking for returns and Result assignments
	for _, stmt := range body.Statements {
		if retStmt, ok := stmt.(*ast.ReturnStatement); ok {
			if retStmt.ReturnValue != nil {
				// Analyze the return value expression to get its type
				retType := a.analyzeExpression(retStmt.ReturnValue)
				if retType != nil {
					returnTypes = append(returnTypes, retType)
				}
			} else {
				// Return with no value - procedure
				returnTypes = append(returnTypes, types.VOID)
			}
		} else if assignStmt, ok := stmt.(*ast.AssignmentStatement); ok {
			// Check if this is a Result assignment
			if ident, ok := assignStmt.Target.(*ast.Identifier); ok {
				if ident.Value == "Result" {
					// This is a Result assignment - infer type from RHS
					rhsType := a.analyzeExpression(assignStmt.Value)
					if rhsType != nil {
						returnTypes = append(returnTypes, rhsType)
					}
				}
			}
		} else if ifStmt, ok := stmt.(*ast.IfStatement); ok {
			// Recursively check if/else blocks for Result assignments
			if ifStmt.Consequence != nil {
				var consequenceType types.Type
				if consequenceBlock, ok := ifStmt.Consequence.(*ast.BlockStatement); ok {
					consequenceType = a.inferReturnTypeFromBody(consequenceBlock)
				} else if assignStmt, ok := ifStmt.Consequence.(*ast.AssignmentStatement); ok {
					// Single assignment statement - check if it's Result
					if ident, ok := assignStmt.Target.(*ast.Identifier); ok && ident.Value == "Result" {
						consequenceType = a.analyzeExpression(assignStmt.Value)
					}
				}
				if consequenceType != nil && consequenceType != types.VOID {
					returnTypes = append(returnTypes, consequenceType)
				}
			}
			if ifStmt.Alternative != nil {
				var alternativeType types.Type
				if alternativeBlock, ok := ifStmt.Alternative.(*ast.BlockStatement); ok {
					alternativeType = a.inferReturnTypeFromBody(alternativeBlock)
				} else if assignStmt, ok := ifStmt.Alternative.(*ast.AssignmentStatement); ok {
					// Single assignment statement - check if it's Result
					if ident, ok := assignStmt.Target.(*ast.Identifier); ok && ident.Value == "Result" {
						alternativeType = a.analyzeExpression(assignStmt.Value)
					}
				}
				if alternativeType != nil && alternativeType != types.VOID {
					returnTypes = append(returnTypes, alternativeType)
				}
			}
		}
	}

	if len(returnTypes) == 0 {
		// No return statements found - treat as procedure
		return types.VOID
	}

	// Check if all return types are compatible
	firstType := returnTypes[0]
	for i, retType := range returnTypes {
		if i == 0 {
			continue
		}
		// Check if types are compatible
		if !a.canAssign(retType, firstType) && !a.canAssign(firstType, retType) {
			a.addError("lambda has conflicting return types: %s and %s",
				firstType.String(), retType.String())
			return nil
		}
	}

	return firstType
}

// ============================================================================
// Closure Capture Analysis
// ============================================================================

// analyzeCapturedVariables identifies variables from outer scopes used in the lambda body.
// Task 9.217: Closure capture analysis
//
// Strategy:
//   - Walk through all identifiers in the lambda body
//   - Check if they're defined in the lambda scope (parameters or locals)
//   - If not, check if they're in an outer scope
//   - If yes, add them to the captured variables list
//   - Validate that captured variables are accessible
func (a *Analyzer) analyzeCapturedVariables(body *ast.BlockStatement, lambdaScope, outerScope *SymbolTable) []string {
	if body == nil {
		return nil
	}

	// Collect all identifiers used in the body
	identifiers := a.collectIdentifiers(body)

	// Track captured variables (use map to avoid duplicates)
	capturedMap := make(map[string]bool)

	for _, ident := range identifiers {
		// Skip special identifiers
		if ident == "Result" {
			continue
		}

		// Check if it's defined in the lambda scope (parameter or local var)
		if lambdaScope.IsDeclaredInCurrentScope(ident) {
			continue
		}

		// Check if it's defined in an outer scope
		if _, ok := outerScope.Resolve(ident); ok {
			capturedMap[ident] = true
		}
		// Note: If not found in outer scope, it will be reported as undefined
		// during normal semantic analysis, so we don't need to report it here
	}

	// Convert map to slice
	captured := make([]string, 0, len(capturedMap))
	for varName := range capturedMap {
		captured = append(captured, varName)
	}

	return captured
}

// collectIdentifiers walks through the AST and collects all identifier names used in expressions.
// This is a helper for closure capture analysis.
func (a *Analyzer) collectIdentifiers(node ast.Node) []string {
	identifiers := []string{}

	switch n := node.(type) {
	case *ast.BlockStatement:
		for _, stmt := range n.Statements {
			identifiers = append(identifiers, a.collectIdentifiers(stmt)...)
		}

	case *ast.AssignmentStatement:
		identifiers = append(identifiers, a.collectIdentifiers(n.Value)...)
		// Don't collect from Target if it's a simple identifier (that's a definition, not a use)
		// But do collect from complex targets like array[i] or obj.field
		if _, ok := n.Target.(*ast.Identifier); !ok {
			identifiers = append(identifiers, a.collectIdentifiers(n.Target)...)
		}

	case *ast.VarDeclStatement:
		// Variable declaration - collect from the value expression
		if n.Value != nil {
			identifiers = append(identifiers, a.collectIdentifiers(n.Value)...)
		}

	case *ast.IfStatement:
		identifiers = append(identifiers, a.collectIdentifiers(n.Condition)...)
		if n.Consequence != nil {
			identifiers = append(identifiers, a.collectIdentifiers(n.Consequence)...)
		}
		if n.Alternative != nil {
			identifiers = append(identifiers, a.collectIdentifiers(n.Alternative)...)
		}

	case *ast.WhileStatement:
		identifiers = append(identifiers, a.collectIdentifiers(n.Condition)...)
		if n.Body != nil {
			identifiers = append(identifiers, a.collectIdentifiers(n.Body)...)
		}

	case *ast.ForStatement:
		if n.Start != nil {
			identifiers = append(identifiers, a.collectIdentifiers(n.Start)...)
		}
		if n.End != nil {
			identifiers = append(identifiers, a.collectIdentifiers(n.End)...)
		}
		if n.Body != nil {
			identifiers = append(identifiers, a.collectIdentifiers(n.Body)...)
		}

	case *ast.ReturnStatement:
		if n.ReturnValue != nil {
			identifiers = append(identifiers, a.collectIdentifiers(n.ReturnValue)...)
		}

	case *ast.ExpressionStatement:
		identifiers = append(identifiers, a.collectIdentifiers(n.Expression)...)

	case *ast.Identifier:
		identifiers = append(identifiers, n.Value)

	case *ast.BinaryExpression:
		identifiers = append(identifiers, a.collectIdentifiers(n.Left)...)
		identifiers = append(identifiers, a.collectIdentifiers(n.Right)...)

	case *ast.UnaryExpression:
		identifiers = append(identifiers, a.collectIdentifiers(n.Right)...)

	case *ast.CallExpression:
		identifiers = append(identifiers, a.collectIdentifiers(n.Function)...)
		for _, arg := range n.Arguments {
			identifiers = append(identifiers, a.collectIdentifiers(arg)...)
		}

	case *ast.IndexExpression:
		identifiers = append(identifiers, a.collectIdentifiers(n.Left)...)
		identifiers = append(identifiers, a.collectIdentifiers(n.Index)...)

	case *ast.MemberAccessExpression:
		identifiers = append(identifiers, a.collectIdentifiers(n.Object)...)

	case *ast.GroupedExpression:
		identifiers = append(identifiers, a.collectIdentifiers(n.Expression)...)

	case *ast.LambdaExpression:
		// Nested lambda - collect from its body, but it has its own scope
		// For now, we'll just note that nested lambdas may capture from this lambda
		// The nested lambda's own capture analysis will handle this
		// We don't recurse into nested lambdas for capture analysis
		// because they have their own independent capture lists

	// Literals don't contain identifiers
	case *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral, *ast.BooleanLiteral, *ast.NilLiteral:
		// No identifiers

		// Default: ignore other node types
	}

	return identifiers
}
