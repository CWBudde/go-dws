package semantic

import (
	"fmt"
	"strings"

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

	// Class registry for tracking declared classes (Task 7.54)
	classes map[string]*types.ClassType

	// Interface registry for tracking declared interfaces (Task 7.97)
	interfaces map[string]*types.InterfaceType

	// Current class being analyzed (for field/method access)
	currentClass *types.ClassType

	// Operator registries (Stage 8)
	globalOperators    *types.OperatorRegistry
	conversionRegistry *types.ConversionRegistry
}

// NewAnalyzer creates a new semantic analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		symbols:            NewSymbolTable(),
		errors:             make([]string, 0),
		classes:            make(map[string]*types.ClassType),
		interfaces:         make(map[string]*types.InterfaceType),
		globalOperators:    types.NewOperatorRegistry(),
		conversionRegistry: types.NewConversionRegistry(),
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

// canAssign checks assignment compatibility, accounting for implicit conversions.
func (a *Analyzer) canAssign(from, to types.Type) bool {
	if from == nil || to == nil {
		return false
	}
	if types.IsCompatible(from, to) {
		return true
	}
	if fromClass, ok := from.(*types.ClassType); ok {
		if toClass, ok := to.(*types.ClassType); ok {
			if fromClass.Equals(toClass) || a.isDescendantOf(fromClass, toClass) {
				return true
			}
		}
	}
	if sig, ok := a.conversionRegistry.FindImplicit(from, to); ok && sig != nil {
		return true
	}
	return false
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
	case *ast.ClassDecl:
		a.analyzeClassDecl(s)
	case *ast.InterfaceDecl:
		a.analyzeInterfaceDecl(s)
	case *ast.OperatorDecl:
		a.analyzeOperatorDecl(s)
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
		// Explicit type annotation - use resolveType to handle both basic and class types
		varType, err = a.resolveType(stmt.Type.Name)
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
			if !a.canAssign(initType, varType) {
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
	if !a.canAssign(valueType, sym.Type) {
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
		if strings.EqualFold(decl.OperatorSymbol, "implicit") {
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
				if !a.canAssign(valueType, caseType) {
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
	// Check if this is a method implementation (has ClassName)
	if decl.ClassName != nil {
		a.analyzeMethodImplementation(decl)
		return
	}

	// This is a regular function (not a method implementation)
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
	case *ast.NewExpression:
		return a.analyzeNewExpression(e)
	case *ast.MemberAccessExpression:
		return a.analyzeMemberAccessExpression(e)
	case *ast.MethodCallExpression:
		return a.analyzeMethodCallExpression(e)
	default:
		a.addError("unknown expression type: %T", expr)
		return nil
	}
}

// analyzeIdentifier analyzes an identifier and returns its type
func (a *Analyzer) analyzeIdentifier(ident *ast.Identifier) types.Type {
	sym, ok := a.symbols.Resolve(ident.Value)
	if !ok {
		if classType, exists := a.classes[ident.Value]; exists {
			return classType
		}
		if a.currentClass != nil {
			if owner := a.getFieldOwner(a.currentClass.Parent, ident.Value); owner != nil {
				if visibility, ok := owner.FieldVisibility[ident.Value]; ok && visibility == int(ast.VisibilityPrivate) {
					a.addError("cannot access private field '%s' of class '%s' at %s",
						ident.Value, owner.Name, ident.Token.Pos.String())
					return nil
				}
			}
		}
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

	if sig, ok := a.resolveBinaryOperator(operator, leftType, rightType); ok {
		if sig.ResultType != nil {
			return sig.ResultType
		}
		return types.VOID
	}

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
			if !leftType.Equals(rightType) && !a.canAssign(leftType, rightType) && !a.canAssign(rightType, leftType) {
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
			if !leftType.Equals(rightType) && !a.canAssign(leftType, rightType) && !a.canAssign(rightType, leftType) {
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

	if sig, ok := a.resolveUnaryOperator(operator, operandType); ok {
		if sig.ResultType != nil {
			return sig.ResultType
		}
		return types.VOID
	}

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
		if argType != nil && !a.canAssign(argType, expectedType) {
			a.addError("argument %d to function '%s' has type %s, expected %s at %s",
				i+1, funcIdent.Value, argType.String(), expectedType.String(),
				expr.Token.Pos.String())
		}
	}

	return funcType.ReturnType
}

// ============================================================================
// Class Analysis (Tasks 7.54-7.59)
// ============================================================================

// analyzeClassDecl analyzes a class declaration
func (a *Analyzer) analyzeClassDecl(decl *ast.ClassDecl) {
	className := decl.Name.Value

	// Check if class is already declared (Task 7.55)
	if _, exists := a.classes[className]; exists {
		a.addError("class '%s' already declared at %s", className, decl.Token.Pos.String())
		return
	}

	// Resolve parent class if specified (Task 7.55)
	var parentClass *types.ClassType
	if decl.Parent != nil {
		parentName := decl.Parent.Value
		var found bool
		parentClass, found = a.classes[parentName]
		if !found {
			a.addError("parent class '%s' not found at %s", parentName, decl.Token.Pos.String())
			return
		}
	}

	// Create new class type
	classType := types.NewClassType(className, parentClass)

	// Set abstract flag (Task 7.65e)
	classType.IsAbstract = decl.IsAbstract

	// Set external flags (Task 7.138)
	classType.IsExternal = decl.IsExternal
	classType.ExternalName = decl.ExternalName

	// Validate external class inheritance (Task 7.139)
	if decl.IsExternal {
		// External class must inherit from nil (Object) or another external class
		if parentClass != nil && !parentClass.IsExternal {
			a.addError("external class '%s' cannot inherit from non-external class '%s' at %s",
				className, parentClass.Name, decl.Token.Pos.String())
			return
		}
	} else {
		// Non-external class cannot inherit from external class
		if parentClass != nil && parentClass.IsExternal {
			a.addError("non-external class '%s' cannot inherit from external class '%s' at %s",
				className, parentClass.Name, decl.Token.Pos.String())
			return
		}
	}

	// Check for circular inheritance (Task 7.55)
	if parentClass != nil && a.hasCircularInheritance(classType) {
		a.addError("circular inheritance detected in class '%s' at %s", className, decl.Token.Pos.String())
		return
	}

	// Analyze and add fields (Task 7.55, 7.62)
	fieldNames := make(map[string]bool)
	classVarNames := make(map[string]bool)
	for _, field := range decl.Fields {
		fieldName := field.Name.Value

		// Check if this is a class variable (static field) - Task 7.62
		if field.IsClassVar {
			// Check for duplicate class variable names
			if classVarNames[fieldName] {
				a.addError("duplicate class variable '%s' in class '%s' at %s",
					fieldName, className, field.Token.Pos.String())
				continue
			}
			classVarNames[fieldName] = true

			// Verify class variable type exists
			if field.Type == nil {
				a.addError("class variable '%s' missing type annotation in class '%s'",
					fieldName, className)
				continue
			}

			fieldType, err := a.resolveType(field.Type.Name)
			if err != nil {
				a.addError("unknown type '%s' for class variable '%s' in class '%s' at %s",
					field.Type.Name, fieldName, className, field.Token.Pos.String())
				continue
			}

			// Store class variable type in ClassType - Task 7.62
			classType.ClassVars[fieldName] = fieldType
		} else {
			// Instance field
			// Check for duplicate field names
			if fieldNames[fieldName] {
				a.addError("duplicate field '%s' in class '%s' at %s",
					fieldName, className, field.Token.Pos.String())
				continue
			}
			fieldNames[fieldName] = true

			// Verify field type exists
			if field.Type == nil {
				a.addError("field '%s' missing type annotation in class '%s'",
					fieldName, className)
				continue
			}

			fieldType, err := a.resolveType(field.Type.Name)
			if err != nil {
				a.addError("unknown type '%s' for field '%s' in class '%s' at %s",
					field.Type.Name, fieldName, className, field.Token.Pos.String())
				continue
			}

			// Add instance field to class
			classType.Fields[fieldName] = fieldType

			// Store field visibility (Task 7.63f)
			classType.FieldVisibility[fieldName] = int(field.Visibility)
		}
	}

	// Register class before analyzing methods (so methods can reference the class)
	a.classes[className] = classType

	// Analyze methods (Task 7.56)
	previousClass := a.currentClass
	a.currentClass = classType
	defer func() { a.currentClass = previousClass }()

	for _, method := range decl.Methods {
		a.analyzeMethodDecl(method, classType)
	}

	// Analyze constructor if present (Task 7.56)
	if decl.Constructor != nil {
		a.analyzeMethodDecl(decl.Constructor, classType)
	}

	// Register class operators (Stage 8)
	a.registerClassOperators(classType, decl)

	// Check method overriding (Task 7.59)
	if parentClass != nil {
		a.checkMethodOverriding(classType, parentClass)
	}

	// Validate interface implementation (Task 7.100)
	if len(decl.Interfaces) > 0 {
		a.validateInterfaceImplementation(classType, decl)
	}

	// Validate abstract class rules (Task 7.65)
	a.validateAbstractClass(classType, decl)
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

		methodType, ok := classType.Methods[opDecl.Binding.Value]
		if !ok {
			a.addError("binding '%s' for class operator '%s' not found in class '%s' at %s",
				opDecl.Binding.Value, opDecl.OperatorSymbol, classType.Name, opDecl.Token.Pos.String())
			continue
		}

		if len(opDecl.OperandTypes) != len(methodType.Parameters) {
			a.addError("binding '%s' for class operator '%s' expects %d parameters, got %d at %s",
				opDecl.Binding.Value, opDecl.OperatorSymbol, len(opDecl.OperandTypes), len(methodType.Parameters), opDecl.Token.Pos.String())
			continue
		}

		extraTypes := make([]types.Type, len(opDecl.OperandTypes))
		errorFound := false
		for i, operand := range opDecl.OperandTypes {
			resolved, err := a.resolveOperatorType(operand.String())
			if err != nil {
				a.addError("unknown type '%s' in class operator declaration at %s", operand.String(), opDecl.Token.Pos.String())
				errorFound = true
				break
			}
			extraTypes[i] = resolved
			if !methodType.Parameters[i].Equals(resolved) {
				a.addError("binding '%s' parameter %d type %s does not match operator operand type %s at %s",
					opDecl.Binding.Value, i+1, methodType.Parameters[i].String(), resolved.String(), opDecl.Token.Pos.String())
				errorFound = true
				break
			}
		}
		if errorFound {
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

		operandTypes := make([]types.Type, 0, len(extraTypes)+1)
		if strings.EqualFold(opDecl.OperatorSymbol, "in") {
			operandTypes = append(operandTypes, extraTypes...)
			operandTypes = append(operandTypes, classType)
		} else {
			operandTypes = append(operandTypes, classType)
			operandTypes = append(operandTypes, extraTypes...)
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

// hasCircularInheritance checks if a class has circular inheritance
func (a *Analyzer) hasCircularInheritance(class *types.ClassType) bool {
	seen := make(map[string]bool)
	current := class

	for current != nil {
		if seen[current.Name] {
			return true
		}
		seen[current.Name] = true
		current = current.Parent
	}

	return false
}

// resolveType resolves a type name to a Type
// Handles basic types and class types
func (a *Analyzer) resolveType(typeName string) (types.Type, error) {
	// Try basic types first
	basicType, err := types.TypeFromString(typeName)
	if err == nil {
		return basicType, nil
	}

	// Try class types
	if classType, found := a.classes[typeName]; found {
		return classType, nil
	}

	return nil, fmt.Errorf("unknown type: %s", typeName)
}

// resolveOperatorType resolves type annotations used in operator declarations.
func (a *Analyzer) resolveOperatorType(typeName string) (types.Type, error) {
	name := strings.TrimSpace(typeName)
	if name == "" {
		return types.VOID, nil
	}

	if t, err := a.resolveType(name); err == nil {
		return t, nil
	}

	lower := strings.ToLower(name)
	if strings.HasPrefix(lower, "array of ") {
		elemName := strings.TrimSpace(name[len("array of "):])
		elemType, err := a.resolveOperatorType(elemName)
		if err != nil {
			return nil, err
		}
		return types.NewDynamicArrayType(elemType), nil
	}

	return nil, fmt.Errorf("unknown type: %s", name)
}

// analyzeMethodImplementation analyzes a method implementation outside a class (Task 7.63v-z)
// This handles code like: function TExample.GetValue: Integer; begin ... end;
func (a *Analyzer) analyzeMethodImplementation(decl *ast.FunctionDecl) {
	className := decl.ClassName.Value

	// Look up the class
	classType, exists := a.classes[className]
	if !exists {
		a.addError("unknown type '%s' at %s", className, decl.Token.Pos.String())
		return
	}

	// Set the current class context
	previousClass := a.currentClass
	a.currentClass = classType
	defer func() { a.currentClass = previousClass }()

	// Use analyzeMethodDecl to analyze the method body with proper scope
	// This will set up Self, fields, and all method scope correctly
	a.analyzeMethodDecl(decl, classType)
}

// analyzeMethodDecl analyzes a method declaration within a class (Task 7.56, 7.61)
func (a *Analyzer) analyzeMethodDecl(method *ast.FunctionDecl, classType *types.ClassType) {
	// Convert parameter types
	paramTypes := make([]types.Type, 0, len(method.Parameters))
	for _, param := range method.Parameters {
		if param.Type == nil {
			a.addError("parameter '%s' missing type annotation in method '%s'",
				param.Name.Value, method.Name.Value)
			return
		}

		paramType, err := a.resolveType(param.Type.Name)
		if err != nil {
			a.addError("unknown parameter type '%s' in method '%s': %v",
				param.Type.Name, method.Name.Value, err)
			return
		}
		paramTypes = append(paramTypes, paramType)
	}

	// Determine return type
	var returnType types.Type
	if method.ReturnType != nil {
		var err error
		returnType, err = a.resolveType(method.ReturnType.Name)
		if err != nil {
			a.addError("unknown return type '%s' in method '%s': %v",
				method.ReturnType.Name, method.Name.Value, err)
			return
		}
	} else {
		returnType = types.VOID
	}

	// Create function type and add to class methods
	funcType := types.NewFunctionType(paramTypes, returnType)
	classType.Methods[method.Name.Value] = funcType
	if method.IsConstructor {
		classType.Constructors[method.Name.Value] = funcType
	}

	// Store method visibility (Task 7.63f)
	// Only set visibility if this is the first time we're seeing this method (declaration in class body)
	// Method implementations outside the class shouldn't overwrite the visibility
	if _, exists := classType.MethodVisibility[method.Name.Value]; !exists {
		classType.MethodVisibility[method.Name.Value] = int(method.Visibility)
	}

	// Store virtual/override/abstract flags (Task 7.64, 7.65)
	classType.VirtualMethods[method.Name.Value] = method.IsVirtual
	classType.OverrideMethods[method.Name.Value] = method.IsOverride
	classType.AbstractMethods[method.Name.Value] = method.IsAbstract

	// Analyze method body in new scope
	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(oldSymbols)
	defer func() { a.symbols = oldSymbols }()

	// Task 7.61: Check if this is a class method (static method)
	if method.IsClassMethod {
		// Class methods (static methods) do NOT have access to Self or instance fields
		// They can only access class variables (static fields)
		// Do NOT add Self to scope
		// Do NOT add instance fields to scope

		// Add class variables to scope (Task 7.62)
		for classVarName, classVarType := range classType.ClassVars {
			a.symbols.Define(classVarName, classVarType)
		}

		// If class has parent, add parent class variables too
		if classType.Parent != nil {
			a.addParentClassVarsToScope(classType.Parent)
		}
	} else {
		// Instance method - add Self reference to method scope (Task 7.56)
		a.symbols.Define("Self", classType)

		// Add class fields to method scope (Task 7.56)
		for fieldName, fieldType := range classType.Fields {
			a.symbols.Define(fieldName, fieldType)
		}

		// Add class variables to method scope (Task 7.62)
		// Instance methods can also access class variables
		for classVarName, classVarType := range classType.ClassVars {
			a.symbols.Define(classVarName, classVarType)
		}

		// If class has parent, add parent fields and class variables too
		if classType.Parent != nil {
			a.addParentFieldsToScope(classType.Parent)
			a.addParentClassVarsToScope(classType.Parent)
		}
	}

	// Add parameters to method scope (both instance and class methods have parameters)
	for i, param := range method.Parameters {
		a.symbols.Define(param.Name.Value, paramTypes[i])
	}

	// For methods with return type, add Result variable
	if returnType != types.VOID {
		a.symbols.Define("Result", returnType)
		a.symbols.Define(method.Name.Value, returnType)
	}

	// Set current function for return statement checking
	previousFunc := a.currentFunction
	a.currentFunction = method
	defer func() { a.currentFunction = previousFunc }()

	// Task 7.64e-h: Validate virtual/override usage
	a.validateVirtualOverride(method, classType, funcType)

	// Analyze method body
	if method.Body != nil {
		a.analyzeBlock(method.Body)
	}
}

// validateVirtualOverride validates virtual/override method declarations (Task 7.64e-h)
func (a *Analyzer) validateVirtualOverride(method *ast.FunctionDecl, classType *types.ClassType, methodType *types.FunctionType) {
	methodName := method.Name.Value

	// Task 7.64f: If method is marked override, validate parent has virtual method
	if method.IsOverride {
		if classType.Parent == nil {
			a.addError("method '%s' marked as override, but class has no parent", methodName)
			return
		}

		// Find method in parent class hierarchy
		parentMethod := a.findMethodInParent(methodName, classType.Parent)
		if parentMethod == nil {
			a.addError("method '%s' marked as override, but no such method exists in parent class", methodName)
			return
		}

		// Task 7.64g: Check that parent method is virtual or override
		if !a.isMethodVirtualOrOverride(methodName, classType.Parent) {
			a.addError("method '%s' marked as override, but parent method is not virtual", methodName)
			return
		}

		// Task 7.64f: Ensure signatures match
		if !a.methodSignaturesMatch(methodType, parentMethod) {
			a.addError("method '%s' override signature does not match parent method signature", methodName)
			return
		}
	}

	// Task 7.64h: Warn if redefining virtual method without override keyword
	if !method.IsOverride && !method.IsVirtual && classType.Parent != nil {
		parentMethod := a.findMethodInParent(methodName, classType.Parent)
		if parentMethod != nil && a.isMethodVirtualOrOverride(methodName, classType.Parent) {
			a.addError("method '%s' hides virtual parent method; use 'override' keyword", methodName)
		}
	}
}

// findMethodInParent searches for a method in the parent class hierarchy
func (a *Analyzer) findMethodInParent(methodName string, parent *types.ClassType) *types.FunctionType {
	if parent == nil {
		return nil
	}

	// Check if method exists in parent
	if methodType, exists := parent.Methods[methodName]; exists {
		return methodType
	}

	// Recursively search in grandparent
	return a.findMethodInParent(methodName, parent.Parent)
}

// isMethodVirtualOrOverride checks if a method is marked virtual or override in class hierarchy
func (a *Analyzer) isMethodVirtualOrOverride(methodName string, classType *types.ClassType) bool {
	if classType == nil {
		return false
	}

	// Check if method exists in this class
	if _, exists := classType.Methods[methodName]; exists {
		// Check if method is virtual or override
		isVirtual := classType.VirtualMethods[methodName]
		isOverride := classType.OverrideMethods[methodName]
		return isVirtual || isOverride
	}

	// Recursively check parent
	return a.isMethodVirtualOrOverride(methodName, classType.Parent)
}

// methodSignaturesMatch compares two function signatures
func (a *Analyzer) methodSignaturesMatch(sig1, sig2 *types.FunctionType) bool {
	// Check parameter count
	if len(sig1.Parameters) != len(sig2.Parameters) {
		return false
	}

	// Check parameter types
	for i := range sig1.Parameters {
		if !sig1.Parameters[i].Equals(sig2.Parameters[i]) {
			return false
		}
	}

	// Check return type
	if !sig1.ReturnType.Equals(sig2.ReturnType) {
		return false
	}

	return true
}

// addParentFieldsToScope recursively adds parent class fields to current scope
func (a *Analyzer) addParentFieldsToScope(parent *types.ClassType) {
	if parent == nil {
		return
	}

	// Add parent's fields
	for fieldName, fieldType := range parent.Fields {
		// Don't override if already defined (shadowing)
		if !a.symbols.IsDeclaredInCurrentScope(fieldName) {
			if visibility, ok := parent.FieldVisibility[fieldName]; ok && visibility == int(ast.VisibilityPrivate) {
				continue
			}
			a.symbols.Define(fieldName, fieldType)
		}
	}

	// Recursively add grandparent fields
	if parent.Parent != nil {
		a.addParentFieldsToScope(parent.Parent)
	}
}

// checkMethodOverriding checks if overridden methods have compatible signatures (Task 7.59)
func (a *Analyzer) checkMethodOverriding(class, parent *types.ClassType) {
	for methodName, childMethodType := range class.Methods {
		// Check if method exists in parent
		parentMethodType, found := parent.GetMethod(methodName)
		if !found {
			// New method in child class - OK
			continue
		}

		// Method exists in parent - check signature compatibility
		if !childMethodType.Equals(parentMethodType) {
			a.addError("method '%s' signature mismatch in class '%s': expected %s, got %s",
				methodName, class.Name, parentMethodType.String(), childMethodType.String())
		}
	}
}

// checkVisibility checks if a member (field or method) is accessible from the current context (Task 7.63g-l).
// Returns true if accessible, false otherwise.
//
// Visibility rules:
//   - Private: only accessible from the same class
//   - Protected: accessible from the same class and all descendants
//   - Public: accessible from anywhere
//
// Parameters:
//   - memberClass: the class that owns the member
//   - visibility: the visibility level of the member (ast.Visibility as int)
//   - memberName: the name of the member (for error messages)
//   - memberType: "field" or "method" (for error messages)
func (a *Analyzer) checkVisibility(memberClass *types.ClassType, visibility int, memberName, memberType string) bool {
	// Public is always accessible (Task 7.63i)
	if visibility == int(ast.VisibilityPublic) {
		return true
	}

	// If we're analyzing code outside any class context, only public members are accessible
	if a.currentClass == nil {
		return false
	}

	// Private members are only accessible from the same class (Task 7.63g, 7.63l)
	if visibility == int(ast.VisibilityPrivate) {
		return a.currentClass.Name == memberClass.Name
	}

	// Protected members are accessible from the same class and descendants (Task 7.63h)
	if visibility == int(ast.VisibilityProtected) {
		// Same class?
		if a.currentClass.Name == memberClass.Name {
			return true
		}

		// Check if current class inherits from member's class
		return a.isDescendantOf(a.currentClass, memberClass)
	}

	// Should not reach here, but default to false for safety
	return false
}

// isDescendantOf checks if a class is a descendant of another class
func (a *Analyzer) isDescendantOf(class, ancestor *types.ClassType) bool {
	if class == nil || ancestor == nil {
		return false
	}

	// Walk up the inheritance chain
	current := class.Parent
	for current != nil {
		if current.Name == ancestor.Name {
			return true
		}
		current = current.Parent
	}

	return false
}

// getFieldOwner returns the class that declares a field, walking up the inheritance chain
func (a *Analyzer) getFieldOwner(class *types.ClassType, fieldName string) *types.ClassType {
	if class == nil {
		return nil
	}

	// Check if this class declares the field
	if _, found := class.Fields[fieldName]; found {
		return class
	}

	// Check parent classes
	return a.getFieldOwner(class.Parent, fieldName)
}

// getMethodOwner returns the class that declares a method, walking up the inheritance chain
func (a *Analyzer) getMethodOwner(class *types.ClassType, methodName string) *types.ClassType {
	if class == nil {
		return nil
	}

	// Check if this class declares the method
	if _, found := class.Methods[methodName]; found {
		return class
	}

	// Check parent classes
	return a.getMethodOwner(class.Parent, methodName)
}

// analyzeNewExpression analyzes object creation (Task 7.57, 7.65f)
func (a *Analyzer) analyzeNewExpression(expr *ast.NewExpression) types.Type {
	className := expr.ClassName.Value

	// Look up class in registry
	classType, found := a.classes[className]
	if !found {
		a.addError("undefined class '%s' at %s", className, expr.Token.Pos.String())
		return nil
	}

	// Check if trying to instantiate an abstract class (Task 7.65f)
	if classType.IsAbstract {
		a.addError("cannot instantiate abstract class '%s' at %s", className, expr.Token.Pos.String())
		return nil
	}

	// Check if class has a constructor
	constructorType, hasConstructor := classType.GetMethod("Create")
	if hasConstructor {
		if owner := a.getMethodOwner(classType, "Create"); owner != nil {
			if visibility, ok := owner.MethodVisibility["Create"]; ok {
				if !a.checkVisibility(owner, visibility, "Create", "method") {
					visibilityStr := ast.Visibility(visibility).String()
					a.addError("cannot access %s constructor 'Create' of class '%s' at %s",
						visibilityStr, owner.Name, expr.Token.Pos.String())
					return classType
				}
			}
		}

		// Validate constructor arguments
		if len(expr.Arguments) != len(constructorType.Parameters) {
			a.addError("constructor for class '%s' expects %d arguments, got %d at %s",
				className, len(constructorType.Parameters), len(expr.Arguments),
				expr.Token.Pos.String())
			return classType
		}

		// Check argument types
		for i, arg := range expr.Arguments {
			argType := a.analyzeExpression(arg)
			expectedType := constructorType.Parameters[i]
			if argType != nil && !a.canAssign(argType, expectedType) {
				a.addError("argument %d to constructor of '%s' has type %s, expected %s at %s",
					i+1, className, argType.String(), expectedType.String(),
					expr.Token.Pos.String())
			}
		}
	}
	// If no constructor but arguments provided, that's OK - default constructor

	return classType
}

// analyzeMemberAccessExpression analyzes member access (Task 7.58)
func (a *Analyzer) analyzeMemberAccessExpression(expr *ast.MemberAccessExpression) types.Type {
	// Analyze the object expression
	objectType := a.analyzeExpression(expr.Object)
	if objectType == nil {
		// Error already reported
		return nil
	}

	// Check if object is a class type
	classType, ok := objectType.(*types.ClassType)
	if !ok {
		a.addError("member access requires class type, got %s at %s",
			objectType.String(), expr.Token.Pos.String())
		return nil
	}

	memberName := expr.Member.Value

	// Look up field in class (including inherited fields)
	fieldType, found := classType.GetField(memberName)
	if found {
		// Check field visibility (Task 7.63j)
		fieldOwner := a.getFieldOwner(classType, memberName)
		if fieldOwner != nil {
			visibility, hasVisibility := fieldOwner.FieldVisibility[memberName]
			if hasVisibility && !a.checkVisibility(fieldOwner, visibility, memberName, "field") {
				visibilityStr := ast.Visibility(visibility).String()
				a.addError("cannot access %s field '%s' of class '%s' at %s",
					visibilityStr, memberName, fieldOwner.Name, expr.Token.Pos.String())
				return nil
			}
		}
		return fieldType
	}

	// Look up method in class (for method references)
	methodType, found := classType.GetMethod(memberName)
	if found {
		// Check method visibility (Task 7.63k)
		methodOwner := a.getMethodOwner(classType, memberName)
		if methodOwner != nil {
			visibility, hasVisibility := methodOwner.MethodVisibility[memberName]
			if hasVisibility && !a.checkVisibility(methodOwner, visibility, memberName, "method") {
				visibilityStr := ast.Visibility(visibility).String()
				a.addError("cannot access %s method '%s' of class '%s' at %s",
					visibilityStr, memberName, methodOwner.Name, expr.Token.Pos.String())
				return nil
			}
		}
		return methodType
	}

	// Member not found
	a.addError("class '%s' has no member '%s' at %s",
		classType.Name, memberName, expr.Token.Pos.String())
	return nil
}

// analyzeMethodCallExpression analyzes a method call on an object
func (a *Analyzer) analyzeMethodCallExpression(expr *ast.MethodCallExpression) types.Type {
	// Analyze the object expression
	objectType := a.analyzeExpression(expr.Object)
	if objectType == nil {
		// Error already reported
		return nil
	}

	// Check if object is a class type
	classType, ok := objectType.(*types.ClassType)
	if !ok {
		a.addError("method call requires class type, got %s at %s",
			objectType.String(), expr.Token.Pos.String())
		return nil
	}

	methodName := expr.Method.Value

	// Look up method in class (including inherited methods)
	methodType, found := classType.GetMethod(methodName)
	if !found {
		a.addError("class '%s' has no method '%s' at %s",
			classType.Name, methodName, expr.Token.Pos.String())
		return nil
	}

	// Check method visibility (Task 7.63k)
	methodOwner := a.getMethodOwner(classType, methodName)
	if methodOwner != nil {
		visibility, hasVisibility := methodOwner.MethodVisibility[methodName]
		if hasVisibility && !a.checkVisibility(methodOwner, visibility, methodName, "method") {
			visibilityStr := ast.Visibility(visibility).String()
			if methodOwner.HasConstructor(methodName) {
				a.addError("cannot access %s constructor '%s' of class '%s' at %s",
					visibilityStr, methodName, methodOwner.Name, expr.Token.Pos.String())
				return classType
			}
			a.addError("cannot call %s method '%s' of class '%s' at %s",
				visibilityStr, methodName, methodOwner.Name, expr.Token.Pos.String())
			return methodType.ReturnType
		}
	}

	// Check argument count
	if len(expr.Arguments) != len(methodType.Parameters) {
		a.addError("method '%s' of class '%s' expects %d arguments, got %d at %s",
			methodName, classType.Name, len(methodType.Parameters), len(expr.Arguments),
			expr.Token.Pos.String())
		return methodType.ReturnType
	}

	// Check argument types
	for i, arg := range expr.Arguments {
		argType := a.analyzeExpression(arg)
		expectedType := methodType.Parameters[i]
		if argType != nil && !a.canAssign(argType, expectedType) {
			a.addError("argument %d to method '%s' of class '%s' has type %s, expected %s at %s",
				i+1, methodName, classType.Name, argType.String(), expectedType.String(),
				expr.Token.Pos.String())
		}
	}

	if classType.HasConstructor(methodName) {
		return classType
	}
	return methodType.ReturnType
}

// addParentClassVarsToScope recursively adds parent class variables to current scope (Task 7.62)
func (a *Analyzer) addParentClassVarsToScope(parent *types.ClassType) {
	if parent == nil {
		return
	}

	// Add parent's class variables
	for classVarName, classVarType := range parent.ClassVars {
		// Don't override if already defined (shadowing)
		if !a.symbols.IsDeclaredInCurrentScope(classVarName) {
			a.symbols.Define(classVarName, classVarType)
		}
	}

	// Recursively add grandparent class variables
	if parent.Parent != nil {
		a.addParentClassVarsToScope(parent.Parent)
	}
}

// ============================================================================
// Abstract Class/Method Validation (Task 7.65)
// ============================================================================

// validateAbstractClass validates abstract class rules:
// 1. Abstract methods can only exist in abstract classes (Task 7.65i)
// 2. Concrete classes must implement all inherited abstract methods (Task 7.65g)
// 3. Abstract methods are implicitly virtual
func (a *Analyzer) validateAbstractClass(classType *types.ClassType, decl *ast.ClassDecl) {
	// Rule 1: Abstract methods can only exist in abstract classes
	for methodName, isAbstract := range classType.AbstractMethods {
		if isAbstract && !classType.IsAbstract {
			a.addError("abstract method '%s' can only be declared in an abstract class at %s",
				methodName, decl.Token.Pos.String())
		}

		// Abstract methods are implicitly virtual
		if isAbstract {
			classType.VirtualMethods[methodName] = true
		}
	}

	// Rule 2: Concrete classes must implement all inherited abstract methods
	if !classType.IsAbstract {
		unimplementedMethods := a.getUnimplementedAbstractMethods(classType)
		if len(unimplementedMethods) > 0 {
			for _, methodName := range unimplementedMethods {
				a.addError("concrete class '%s' must implement abstract method '%s' at %s",
					classType.Name, methodName, decl.Token.Pos.String())
			}
		}
	}
}

// getUnimplementedAbstractMethods returns a list of abstract methods from the inheritance chain
// that are not implemented in the given class or its ancestors.
func (a *Analyzer) getUnimplementedAbstractMethods(classType *types.ClassType) []string {
	unimplemented := []string{}

	// Collect all abstract methods from parent chain
	abstractMethods := a.collectAbstractMethods(classType.Parent)

	// Check which ones are not implemented in this class
	for methodName := range abstractMethods {
		// Check if this class implements the method (non-abstract)
		if classType.AbstractMethods[methodName] {
			// Still abstract in this class - not implemented
			unimplemented = append(unimplemented, methodName)
		} else if _, hasMethod := classType.Methods[methodName]; !hasMethod {
			// Method not defined in this class at all - not implemented
			unimplemented = append(unimplemented, methodName)
		}
		// Otherwise, method is implemented (exists and is not abstract)
	}

	return unimplemented
}

// collectAbstractMethods recursively collects all abstract methods from the parent chain
func (a *Analyzer) collectAbstractMethods(parent *types.ClassType) map[string]bool {
	abstractMethods := make(map[string]bool)

	if parent == nil {
		return abstractMethods
	}

	// Add parent's abstract methods
	for methodName, isAbstract := range parent.AbstractMethods {
		if isAbstract {
			abstractMethods[methodName] = true
		}
	}

	// Recursively collect from grandparent
	grandparentAbstract := a.collectAbstractMethods(parent.Parent)
	for methodName := range grandparentAbstract {
		// Only add if not already implemented (non-abstract) in parent
		if !parent.AbstractMethods[methodName] {
			if _, hasMethod := parent.Methods[methodName]; hasMethod {
				// Parent implemented it, don't add to abstract list
				continue
			}
		}
		abstractMethods[methodName] = true
	}

	return abstractMethods
}

// ============================================================================
// Interface Analysis (Task 7.96-7.103)
// ============================================================================

// analyzeInterfaceDecl analyzes an interface declaration (Task 7.98)
func (a *Analyzer) analyzeInterfaceDecl(decl *ast.InterfaceDecl) {
	interfaceName := decl.Name.Value

	// Check if interface is already declared (Task 7.98)
	if _, exists := a.interfaces[interfaceName]; exists {
		a.addError("interface '%s' already declared at %s", interfaceName, decl.Token.Pos.String())
		return
	}

	// Resolve parent interface if specified (Task 7.98)
	var parentInterface *types.InterfaceType
	if decl.Parent != nil {
		parentName := decl.Parent.Value
		var found bool
		parentInterface, found = a.interfaces[parentName]
		if !found {
			a.addError("parent interface '%s' not found at %s", parentName, decl.Token.Pos.String())
			return
		}
	}

	// Create new interface type
	interfaceType := types.NewInterfaceType(interfaceName)
	interfaceType.Parent = parentInterface

	// Set external flag and name if specified
	if decl.IsExternal {
		interfaceType.IsExternal = true
		interfaceType.ExternalName = decl.ExternalName
	}

	// Analyze each method in the interface
	for _, method := range decl.Methods {
		a.analyzeInterfaceMethodDecl(method, interfaceType)
	}

	// Register interface in the registry
	a.interfaces[interfaceName] = interfaceType
}

// analyzeInterfaceMethodDecl analyzes an interface method declaration (Task 7.99)
func (a *Analyzer) analyzeInterfaceMethodDecl(method *ast.InterfaceMethodDecl, iface *types.InterfaceType) {
	methodName := method.Name.Value

	// Build parameter types list
	var paramTypes []types.Type
	for _, param := range method.Parameters {
		paramType, err := a.resolveType(param.Type.Name)
		if err != nil {
			a.addError("unknown parameter type '%s' in interface method '%s' at %s",
				param.Type.Name, methodName, method.Token.Pos.String())
			return
		}
		paramTypes = append(paramTypes, paramType)
	}

	// Determine return type
	var returnType types.Type = types.VOID
	if method.ReturnType != nil {
		var err error
		returnType, err = a.resolveType(method.ReturnType.Name)
		if err != nil {
			a.addError("unknown return type '%s' in interface method '%s' at %s",
				method.ReturnType.Name, methodName, method.Token.Pos.String())
			return
		}
	}

	// Create function type for this interface method
	funcType := types.NewFunctionType(paramTypes, returnType)

	// Add method to interface
	iface.Methods[methodName] = funcType
}

// validateInterfaceImplementation validates that a class implements all required interface methods (Task 7.100)
func (a *Analyzer) validateInterfaceImplementation(classType *types.ClassType, decl *ast.ClassDecl) {
	// For each interface declared on the class
	for _, ifaceIdent := range decl.Interfaces {
		ifaceName := ifaceIdent.Value

		// Lookup the interface type
		ifaceType, found := a.interfaces[ifaceName]
		if !found {
			a.addError("interface '%s' not found at %s", ifaceName, decl.Token.Pos.String())
			continue
		}

		// Store interface in class type's Interfaces list
		classType.Interfaces = append(classType.Interfaces, ifaceType)

		// Check that class implements all interface methods
		allMethods := types.GetAllInterfaceMethods(ifaceType)
		for methodName, ifaceMethod := range allMethods {
			// Check if class has this method
			classMethod, hasMethod := classType.GetMethod(methodName)
			if !hasMethod {
				a.addError("class '%s' does not implement interface method '%s' from interface '%s' at %s",
					classType.Name, methodName, ifaceName, decl.Token.Pos.String())
				continue
			}

			// Check that signatures match (Task 7.103)
			// Use existing methodSignaturesMatch from analyzer.go:1038
			if !a.methodSignaturesMatch(classMethod, ifaceMethod) {
				a.addError("method '%s' in class '%s' does not match interface signature from '%s' at %s",
					methodName, classType.Name, ifaceName, decl.Token.Pos.String())
			}
		}
	}
}
