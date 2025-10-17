package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/types"
)

// Analyzer performs semantic analysis on a DWScript program.
// It validates types, checks for undefined variables, and ensures
// type compatibility in expressions and statements.
type Analyzer struct {
	// Symbol table for tracking variables and functions
	symbols *SymbolTable

	// Accumulated errors during analysis
	errors []string

	// Current function being analyzed (for return type checking)
	currentFunction *ast.FunctionDecl
}

// NewAnalyzer creates a new semantic analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		symbols: NewSymbolTable(),
		errors:  make([]string, 0),
	}
}

// Analyze performs semantic analysis on a program.
// Returns nil if analysis succeeds, or an error if there are semantic errors.
func (a *Analyzer) Analyze(program *ast.Program) error {
	if program == nil {
		return fmt.Errorf("cannot analyze nil program")
	}

	// Analyze each statement in the program
	for _, stmt := range program.Statements {
		a.analyzeStatement(stmt)
	}

	// If we accumulated errors, return them
	if len(a.errors) > 0 {
		return &AnalysisError{Errors: a.errors}
	}

	return nil
}

// Errors returns all accumulated semantic errors
func (a *Analyzer) Errors() []string {
	return a.errors
}

// addError adds a semantic error to the error list
func (a *Analyzer) addError(format string, args ...interface{}) {
	a.errors = append(a.errors, fmt.Sprintf(format, args...))
}

// ============================================================================
// Statement Analysis
// ============================================================================

// analyzeStatement analyzes a statement node
func (a *Analyzer) analyzeStatement(stmt ast.Statement) {
	if stmt == nil {
		return
	}

	switch s := stmt.(type) {
	case *ast.VarDeclStatement:
		a.analyzeVarDecl(s)
	case *ast.AssignmentStatement:
		a.analyzeAssignment(s)
	case *ast.ExpressionStatement:
		a.analyzeExpression(s.Expression)
	case *ast.BlockStatement:
		a.analyzeBlock(s)
	case *ast.IfStatement:
		a.analyzeIf(s)
	case *ast.WhileStatement:
		a.analyzeWhile(s)
	case *ast.RepeatStatement:
		a.analyzeRepeat(s)
	case *ast.ForStatement:
		a.analyzeFor(s)
	case *ast.CaseStatement:
		a.analyzeCase(s)
	case *ast.FunctionDecl:
		a.analyzeFunctionDecl(s)
	case *ast.ReturnStatement:
		a.analyzeReturn(s)
	default:
		// Unknown statement type - this shouldn't happen
		a.addError("unknown statement type: %T", stmt)
	}
}

// analyzeVarDecl analyzes a variable declaration
func (a *Analyzer) analyzeVarDecl(stmt *ast.VarDeclStatement) {
	// Check if variable is already declared in current scope
	if a.symbols.IsDeclaredInCurrentScope(stmt.Name.Value) {
		a.addError("variable '%s' already declared at %s", stmt.Name.Value, stmt.Token.Pos.String())
		return
	}

	// Determine the type of the variable
	var varType types.Type
	var err error

	if stmt.Type != nil {
		// Explicit type annotation
		varType, err = types.TypeFromString(stmt.Type.Name)
		if err != nil {
			a.addError("unknown type '%s' at %s", stmt.Type.Name, stmt.Token.Pos.String())
			return
		}
	}

	// If there's an initializer, check its type
	if stmt.Value != nil {
		initType := a.analyzeExpression(stmt.Value)
		if initType == nil {
			// Error already reported by analyzeExpression
			return
		}

		if varType == nil {
			// Type inference: use initializer's type
			varType = initType
		} else {
			// Check that initializer type is compatible with declared type
			if !types.IsCompatible(initType, varType) {
				a.addError("cannot assign %s to %s in variable declaration at %s",
					initType.String(), varType.String(), stmt.Token.Pos.String())
				return
			}
		}
	}

	// If we still don't have a type, that's an error
	if varType == nil {
		a.addError("variable '%s' must have either a type annotation or an initializer at %s",
			stmt.Name.Value, stmt.Token.Pos.String())
		return
	}

	// Add variable to symbol table
	a.symbols.Define(stmt.Name.Value, varType)
}

// analyzeAssignment analyzes an assignment statement
func (a *Analyzer) analyzeAssignment(stmt *ast.AssignmentStatement) {
	// Look up the variable
	sym, ok := a.symbols.Resolve(stmt.Name.Value)
	if !ok {
		a.addError("undefined variable '%s' at %s", stmt.Name.Value, stmt.Token.Pos.String())
		return
	}

	// Check the type of the value being assigned
	valueType := a.analyzeExpression(stmt.Value)
	if valueType == nil {
		// Error already reported
		return
	}

	// Check type compatibility
	if !types.IsCompatible(valueType, sym.Type) {
		a.addError("cannot assign %s to %s at %s",
			valueType.String(), sym.Type.String(), stmt.Token.Pos.String())
	}
}

// analyzeBlock analyzes a block statement
func (a *Analyzer) analyzeBlock(stmt *ast.BlockStatement) {
	// Create a new scope for the block
	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(oldSymbols)
	defer func() { a.symbols = oldSymbols }()

	// Analyze each statement in the block
	for _, s := range stmt.Statements {
		a.analyzeStatement(s)
	}
}

// analyzeIf analyzes an if statement
func (a *Analyzer) analyzeIf(stmt *ast.IfStatement) {
	// Check condition type
	condType := a.analyzeExpression(stmt.Condition)
	if condType != nil && !condType.Equals(types.BOOLEAN) {
		a.addError("if condition must be boolean, got %s at %s",
			condType.String(), stmt.Token.Pos.String())
	}

	// Analyze consequence
	a.analyzeStatement(stmt.Consequence)

	// Analyze alternative if present
	if stmt.Alternative != nil {
		a.analyzeStatement(stmt.Alternative)
	}
}

// analyzeWhile analyzes a while statement
func (a *Analyzer) analyzeWhile(stmt *ast.WhileStatement) {
	// Check condition type
	condType := a.analyzeExpression(stmt.Condition)
	if condType != nil && !condType.Equals(types.BOOLEAN) {
		a.addError("while condition must be boolean, got %s at %s",
			condType.String(), stmt.Token.Pos.String())
	}

	// Analyze body
	a.analyzeStatement(stmt.Body)
}

// analyzeRepeat analyzes a repeat-until statement
func (a *Analyzer) analyzeRepeat(stmt *ast.RepeatStatement) {
	// Analyze body
	a.analyzeStatement(stmt.Body)

	// Check condition type
	condType := a.analyzeExpression(stmt.Condition)
	if condType != nil && !condType.Equals(types.BOOLEAN) {
		a.addError("repeat-until condition must be boolean, got %s at %s",
			condType.String(), stmt.Token.Pos.String())
	}
}

// analyzeFor analyzes a for statement
func (a *Analyzer) analyzeFor(stmt *ast.ForStatement) {
	// Create a new scope for the loop variable
	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(oldSymbols)
	defer func() { a.symbols = oldSymbols }()

	// Analyze start and end expressions
	startType := a.analyzeExpression(stmt.Start)
	endType := a.analyzeExpression(stmt.End)

	// Check that both are ordinal types (Integer or Boolean)
	if startType != nil && !types.IsOrdinalType(startType) {
		a.addError("for loop start must be ordinal type, got %s at %s",
			startType.String(), stmt.Token.Pos.String())
	}
	if endType != nil && !types.IsOrdinalType(endType) {
		a.addError("for loop end must be ordinal type, got %s at %s",
			endType.String(), stmt.Token.Pos.String())
	}

	// Define loop variable (typically Integer)
	var loopVarType types.Type = types.INTEGER
	if startType != nil && types.IsOrdinalType(startType) {
		loopVarType = startType
	}
	a.symbols.Define(stmt.Variable.Value, loopVarType)

	// Analyze body
	a.analyzeStatement(stmt.Body)
}

// analyzeCase analyzes a case statement
func (a *Analyzer) analyzeCase(stmt *ast.CaseStatement) {
	// Analyze the case expression
	caseType := a.analyzeExpression(stmt.Expression)

	// Analyze each case branch
	for _, branch := range stmt.Cases {
		// Check that case values are compatible with case expression
		for _, value := range branch.Values {
			valueType := a.analyzeExpression(value)
			if caseType != nil && valueType != nil {
				if !types.IsCompatible(valueType, caseType) {
					a.addError("case value type %s incompatible with case expression type %s",
						valueType.String(), caseType.String())
				}
			}
		}
		// Analyze the branch statement
		a.analyzeStatement(branch.Statement)
	}

	// Analyze else branch if present
	if stmt.Else != nil {
		a.analyzeStatement(stmt.Else)
	}
}

// analyzeFunctionDecl analyzes a function declaration
func (a *Analyzer) analyzeFunctionDecl(decl *ast.FunctionDecl) {
	// Convert parameter types and return type
	paramTypes := make([]types.Type, 0, len(decl.Parameters))
	for _, param := range decl.Parameters {
		if param.Type == nil {
			a.addError("parameter '%s' missing type annotation in function '%s'",
				param.Name.Value, decl.Name.Value)
			return
		}
		paramType, err := types.TypeFromString(param.Type.Name)
		if err != nil {
			a.addError("unknown parameter type '%s' in function '%s': %v",
				param.Type.Name, decl.Name.Value, err)
			return
		}
		paramTypes = append(paramTypes, paramType)
	}

	// Determine return type
	var returnType types.Type
	if decl.ReturnType != nil {
		var err error
		returnType, err = types.TypeFromString(decl.ReturnType.Name)
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

	// Create function type and add to symbol table
	funcType := types.NewFunctionType(paramTypes, returnType)
	a.symbols.DefineFunction(decl.Name.Value, funcType)

	// Analyze function body in new scope
	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(oldSymbols)
	defer func() { a.symbols = oldSymbols }()

	// Add parameters to function scope
	for i, param := range decl.Parameters {
		a.symbols.Define(param.Name.Value, paramTypes[i])
	}

	// For functions (not procedures), add Result variable
	if returnType != types.VOID {
		a.symbols.Define("Result", returnType)
		// Also allow function name as return variable (Pascal style)
		a.symbols.Define(decl.Name.Value, returnType)
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
	if a.currentFunction == nil {
		a.addError("return statement outside of function at %s", stmt.Token.Pos.String())
		return
	}

	// Get expected return type
	var expectedType types.Type
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

	// Check return value
	if stmt.ReturnValue != nil {
		if expectedType == types.VOID {
			a.addError("procedure cannot return a value at %s", stmt.Token.Pos.String())
			return
		}

		returnType := a.analyzeExpression(stmt.ReturnValue)
		if returnType != nil && !types.IsCompatible(returnType, expectedType) {
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
// Expression Analysis
// ============================================================================

// analyzeExpression analyzes an expression and returns its type.
// Returns nil if the expression is invalid.
func (a *Analyzer) analyzeExpression(expr ast.Expression) types.Type {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return types.INTEGER
	case *ast.FloatLiteral:
		return types.FLOAT
	case *ast.StringLiteral:
		return types.STRING
	case *ast.BooleanLiteral:
		return types.BOOLEAN
	case *ast.NilLiteral:
		return types.NIL
	case *ast.Identifier:
		return a.analyzeIdentifier(e)
	case *ast.BinaryExpression:
		return a.analyzeBinaryExpression(e)
	case *ast.UnaryExpression:
		return a.analyzeUnaryExpression(e)
	case *ast.GroupedExpression:
		return a.analyzeExpression(e.Expression)
	case *ast.CallExpression:
		return a.analyzeCallExpression(e)
	default:
		a.addError("unknown expression type: %T", expr)
		return nil
	}
}

// analyzeIdentifier analyzes an identifier and returns its type
func (a *Analyzer) analyzeIdentifier(ident *ast.Identifier) types.Type {
	sym, ok := a.symbols.Resolve(ident.Value)
	if !ok {
		a.addError("undefined variable '%s' at %s", ident.Value, ident.Token.Pos.String())
		return nil
	}
	return sym.Type
}

// analyzeBinaryExpression analyzes a binary expression and returns its type
func (a *Analyzer) analyzeBinaryExpression(expr *ast.BinaryExpression) types.Type {
	// Analyze left and right operands
	leftType := a.analyzeExpression(expr.Left)
	rightType := a.analyzeExpression(expr.Right)

	if leftType == nil || rightType == nil {
		// Errors already reported
		return nil
	}

	operator := expr.Operator

	// Handle arithmetic operators
	if operator == "+" || operator == "-" || operator == "*" || operator == "/" {
		// Special case: + can also concatenate strings
		if operator == "+" && (leftType.Equals(types.STRING) || rightType.Equals(types.STRING)) {
			// String concatenation
			if !leftType.Equals(types.STRING) || !rightType.Equals(types.STRING) {
				a.addError("string concatenation requires both operands to be strings at %s",
					expr.Token.Pos.String())
				return nil
			}
			return types.STRING
		}

		// Numeric arithmetic
		if !types.IsNumericType(leftType) || !types.IsNumericType(rightType) {
			a.addError("arithmetic operator %s requires numeric operands, got %s and %s at %s",
				operator, leftType.String(), rightType.String(), expr.Token.Pos.String())
			return nil
		}

		// Type promotion: Integer + Float -> Float
		return types.PromoteTypes(leftType, rightType)
	}

	// Handle integer division and modulo
	if operator == "div" || operator == "mod" {
		if !leftType.Equals(types.INTEGER) || !rightType.Equals(types.INTEGER) {
			a.addError("operator %s requires integer operands, got %s and %s at %s",
				operator, leftType.String(), rightType.String(), expr.Token.Pos.String())
			return nil
		}
		return types.INTEGER
	}

	// Handle comparison operators
	if operator == "=" || operator == "<>" || operator == "<" || operator == ">" || operator == "<=" || operator == ">=" {
		// For equality, types must be comparable
		if operator == "=" || operator == "<>" {
			if !types.IsComparableType(leftType) || !types.IsComparableType(rightType) {
				a.addError("operator %s requires comparable types at %s",
					operator, expr.Token.Pos.String())
				return nil
			}
			// Types must be compatible
			if !leftType.Equals(rightType) && !types.IsCompatible(leftType, rightType) && !types.IsCompatible(rightType, leftType) {
				a.addError("cannot compare %s with %s at %s",
					leftType.String(), rightType.String(), expr.Token.Pos.String())
				return nil
			}
		} else {
			// For ordering, types must be orderable
			if !types.IsOrderedType(leftType) || !types.IsOrderedType(rightType) {
				a.addError("operator %s requires ordered types, got %s and %s at %s",
					operator, leftType.String(), rightType.String(), expr.Token.Pos.String())
				return nil
			}
			// Types must match (or be compatible)
			if !leftType.Equals(rightType) && !types.IsCompatible(leftType, rightType) && !types.IsCompatible(rightType, leftType) {
				a.addError("cannot compare %s with %s at %s",
					leftType.String(), rightType.String(), expr.Token.Pos.String())
				return nil
			}
		}
		return types.BOOLEAN
	}

	// Handle logical operators
	if operator == "and" || operator == "or" || operator == "xor" {
		if !leftType.Equals(types.BOOLEAN) || !rightType.Equals(types.BOOLEAN) {
			a.addError("logical operator %s requires boolean operands, got %s and %s at %s",
				operator, leftType.String(), rightType.String(), expr.Token.Pos.String())
			return nil
		}
		return types.BOOLEAN
	}

	a.addError("unknown binary operator: %s at %s", operator, expr.Token.Pos.String())
	return nil
}

// analyzeUnaryExpression analyzes a unary expression and returns its type
func (a *Analyzer) analyzeUnaryExpression(expr *ast.UnaryExpression) types.Type {
	// Analyze operand
	operandType := a.analyzeExpression(expr.Right)
	if operandType == nil {
		// Error already reported
		return nil
	}

	operator := expr.Operator

	// Handle negation
	if operator == "-" || operator == "+" {
		if !types.IsNumericType(operandType) {
			a.addError("unary %s requires numeric operand, got %s at %s",
				operator, operandType.String(), expr.Token.Pos.String())
			return nil
		}
		return operandType
	}

	// Handle logical not
	if operator == "not" {
		if !operandType.Equals(types.BOOLEAN) {
			a.addError("unary not requires boolean operand, got %s at %s",
				operandType.String(), expr.Token.Pos.String())
			return nil
		}
		return types.BOOLEAN
	}

	a.addError("unknown unary operator: %s at %s", operator, expr.Token.Pos.String())
	return nil
}

// analyzeCallExpression analyzes a function call and returns its type
func (a *Analyzer) analyzeCallExpression(expr *ast.CallExpression) types.Type {
	// Get function name
	funcIdent, ok := expr.Function.(*ast.Identifier)
	if !ok {
		a.addError("function call must use identifier at %s", expr.Token.Pos.String())
		return nil
	}

	// Look up function
	sym, ok := a.symbols.Resolve(funcIdent.Value)
	if !ok {
		// Check if it's a built-in function
		if funcIdent.Value == "PrintLn" || funcIdent.Value == "Print" {
			// Built-in functions - allow any arguments
			// Analyze arguments for side effects
			for _, arg := range expr.Arguments {
				a.analyzeExpression(arg)
			}
			return types.VOID
		}

		a.addError("undefined function '%s' at %s", funcIdent.Value, expr.Token.Pos.String())
		return nil
	}

	// Check that symbol is a function
	funcType, ok := sym.Type.(*types.FunctionType)
	if !ok {
		a.addError("'%s' is not a function at %s", funcIdent.Value, expr.Token.Pos.String())
		return nil
	}

	// Check argument count
	if len(expr.Arguments) != len(funcType.Parameters) {
		a.addError("function '%s' expects %d arguments, got %d at %s",
			funcIdent.Value, len(funcType.Parameters), len(expr.Arguments),
			expr.Token.Pos.String())
		return nil
	}

	// Check argument types
	for i, arg := range expr.Arguments {
		argType := a.analyzeExpression(arg)
		expectedType := funcType.Parameters[i]
		if argType != nil && !types.IsCompatible(argType, expectedType) {
			a.addError("argument %d to function '%s' has type %s, expected %s at %s",
				i+1, funcIdent.Value, argType.String(), expectedType.String(),
				expr.Token.Pos.String())
		}
	}

	return funcType.ReturnType
}
