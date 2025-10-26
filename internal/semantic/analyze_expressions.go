package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

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
	case *ast.RecordLiteral:
		// RecordLiteral needs context to know the expected type
		// This will be handled in analyzeVarDecl or analyzeAssignment
		a.addError("record literal requires type context")
		return nil
	case *ast.SetLiteral:
		// SetLiteral needs context to know the expected type
		// This will be handled in analyzeVarDecl or analyzeAssignment
		return a.analyzeSetLiteralWithContext(e, nil)
	case *ast.IndexExpression:
		return a.analyzeIndexExpression(e)
	default:
		a.addError("unknown expression type: %T", expr)
		return nil
	}
}

// analyzeIdentifier analyzes an identifier and returns its type
func (a *Analyzer) analyzeIdentifier(ident *ast.Identifier) types.Type {
	// Handle built-in ExceptObject variable (Task 8.206)
	// ExceptObject is a global variable that holds the current exception (or nil)
	if ident.Value == "ExceptObject" {
		// ExceptObject is always of type Exception (the base exception class)
		if exceptionClass, exists := a.classes["Exception"]; exists {
			return exceptionClass
		}
		// If Exception class doesn't exist (shouldn't happen), return nil
		a.addError("internal error: Exception class not found")
		return nil
	}

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
		// Task 8.102: Check for set operations first
		leftSetType, leftIsSet := leftType.(*types.SetType)
		rightSetType, rightIsSet := rightType.(*types.SetType)

		if leftIsSet || rightIsSet {
			// At least one operand is a set, so this should be a set operation
			if operator == "/" {
				// Division is not a set operation
				a.addError("operator / is not supported for sets at %s", expr.Token.Pos.String())
				return nil
			}

			// Both operands must be sets
			if !leftIsSet || !rightIsSet {
				a.addError("set operator %s requires set operands, got %s and %s at %s",
					operator, leftType.String(), rightType.String(), expr.Token.Pos.String())
				return nil
			}

			// Element types must be compatible
			if !leftSetType.ElementType.Equals(rightSetType.ElementType) {
				a.addError("incompatible types in set operation: set of %s and set of %s at %s",
					leftSetType.ElementType.String(), rightSetType.ElementType.String(), expr.Token.Pos.String())
				return nil
			}

			// Return the set type (union, difference, intersection all return the same set type)
			return leftSetType
		}

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

	// Task 8.103: Handle 'in' operator for set membership
	if operator == "in" {
		// Right operand must be a set type
		rightSetType, isSet := rightType.(*types.SetType)
		if !isSet {
			a.addError("'in' operator requires set as right operand, got %s at %s",
				rightType.String(), expr.Token.Pos.String())
			return nil
		}

		// Left operand must be an enum type matching the set's element type
		leftEnumType, isEnum := leftType.(*types.EnumType)
		if !isEnum {
			a.addError("'in' operator requires enum value as left operand, got %s at %s",
				leftType.String(), expr.Token.Pos.String())
			return nil
		}

		// Element types must match
		if !leftEnumType.Equals(rightSetType.ElementType) {
			a.addError("type mismatch in 'in' operator: %s is not compatible with set of %s at %s",
				leftEnumType.String(), rightSetType.ElementType.String(), expr.Token.Pos.String())
			return nil
		}

		// 'in' operator returns Boolean
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

		// Ord and Integer built-in functions (Task 8.51, 8.52)
		if funcIdent.Value == "Ord" || funcIdent.Value == "Integer" {
			// These functions take one argument and return an integer
			if len(expr.Arguments) != 1 {
				a.addError("function '%s' expects 1 argument, got %d at %s",
					funcIdent.Value, len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the argument
			a.analyzeExpression(expr.Arguments[0])
			return types.INTEGER
		}

		// Length built-in function (Task 8.130)
		if funcIdent.Value == "Length" {
			// Length takes one argument (array or string) and returns an integer
			if len(expr.Arguments) != 1 {
				a.addError("function 'Length' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the argument
			argType := a.analyzeExpression(expr.Arguments[0])
			// Verify it's an array or string
			if argType != nil {
				if _, isArray := argType.(*types.ArrayType); !isArray {
					if argType != types.STRING {
						a.addError("function 'Length' expects array or string, got %s at %s",
							argType.String(), expr.Token.Pos.String())
					}
				}
			}
			return types.INTEGER
		}

		// Copy built-in function (Task 8.183)
		if funcIdent.Value == "Copy" {
			// Copy takes three arguments (string, index, count) and returns a string
			if len(expr.Arguments) != 3 {
				a.addError("function 'Copy' expects 3 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// Analyze the first argument (string)
			strType := a.analyzeExpression(expr.Arguments[0])
			if strType != nil && strType != types.STRING {
				a.addError("function 'Copy' expects string as first argument, got %s at %s",
					strType.String(), expr.Token.Pos.String())
			}
			// Analyze the second argument (index - integer)
			indexType := a.analyzeExpression(expr.Arguments[1])
			if indexType != nil && indexType != types.INTEGER {
				a.addError("function 'Copy' expects integer as second argument, got %s at %s",
					indexType.String(), expr.Token.Pos.String())
			}
			// Analyze the third argument (count - integer)
			countType := a.analyzeExpression(expr.Arguments[2])
			if countType != nil && countType != types.INTEGER {
				a.addError("function 'Copy' expects integer as third argument, got %s at %s",
					countType.String(), expr.Token.Pos.String())
			}
			return types.STRING
		}

		// Concat built-in function (Task 8.183)
		if funcIdent.Value == "Concat" {
			// Concat takes at least one argument (all strings) and returns a string
			if len(expr.Arguments) == 0 {
				a.addError("function 'Concat' expects at least 1 argument, got 0 at %s",
					expr.Token.Pos.String())
				return types.STRING
			}
			// Analyze all arguments and verify they're strings
			for i, arg := range expr.Arguments {
				argType := a.analyzeExpression(arg)
				if argType != nil && argType != types.STRING {
					a.addError("function 'Concat' expects string as argument %d, got %s at %s",
						i+1, argType.String(), expr.Token.Pos.String())
				}
			}
			return types.STRING
		}

		// Pos built-in function (Task 8.183)
		if funcIdent.Value == "Pos" {
			// Pos takes two string arguments and returns an integer
			if len(expr.Arguments) != 2 {
				a.addError("function 'Pos' expects 2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the first argument (substring)
			substrType := a.analyzeExpression(expr.Arguments[0])
			if substrType != nil && substrType != types.STRING {
				a.addError("function 'Pos' expects string as first argument, got %s at %s",
					substrType.String(), expr.Token.Pos.String())
			}
			// Analyze the second argument (string to search in)
			strType := a.analyzeExpression(expr.Arguments[1])
			if strType != nil && strType != types.STRING {
				a.addError("function 'Pos' expects string as second argument, got %s at %s",
					strType.String(), expr.Token.Pos.String())
			}
			return types.INTEGER
		}

		// UpperCase built-in function (Task 8.183)
		if funcIdent.Value == "UpperCase" {
			// UpperCase takes one string argument and returns a string
			if len(expr.Arguments) != 1 {
				a.addError("function 'UpperCase' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// Analyze the argument and verify it's a string
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && argType != types.STRING {
				a.addError("function 'UpperCase' expects string as argument, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.STRING
		}

		// LowerCase built-in function (Task 8.183)
		if funcIdent.Value == "LowerCase" {
			// LowerCase takes one string argument and returns a string
			if len(expr.Arguments) != 1 {
				a.addError("function 'LowerCase' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// Analyze the argument and verify it's a string
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && argType != types.STRING {
				a.addError("function 'LowerCase' expects string as argument, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.STRING
		}

		// Abs built-in function (Task 8.185)
		if funcIdent.Value == "Abs" {
			// Abs takes one numeric argument and returns the same type
			if len(expr.Arguments) != 1 {
				a.addError("function 'Abs' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER // Default to INTEGER on error
			}
			// Analyze the argument and verify it's Integer or Float
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil {
				if argType != types.INTEGER && argType != types.FLOAT {
					a.addError("function 'Abs' expects Integer or Float as argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
					return types.INTEGER
				}
				// Return the same type as the input
				return argType
			}
			return types.INTEGER // Default to INTEGER if type is unknown
		}

		// Sqrt built-in function (Task 8.185)
		if funcIdent.Value == "Sqrt" {
			// Sqrt takes one numeric argument and always returns Float
			if len(expr.Arguments) != 1 {
				a.addError("function 'Sqrt' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.FLOAT
			}
			// Analyze the argument and verify it's Integer or Float
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil {
				if argType != types.INTEGER && argType != types.FLOAT {
					a.addError("function 'Sqrt' expects Integer or Float as argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			// Always returns Float
			return types.FLOAT
		}

		// Sin built-in function (Task 8.185)
		if funcIdent.Value == "Sin" {
			// Sin takes one numeric argument and always returns Float
			if len(expr.Arguments) != 1 {
				a.addError("function 'Sin' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.FLOAT
			}
			// Analyze the argument and verify it's Integer or Float
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil {
				if argType != types.INTEGER && argType != types.FLOAT {
					a.addError("function 'Sin' expects Integer or Float as argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			// Always returns Float
			return types.FLOAT
		}

		// Cos built-in function (Task 8.185)
		if funcIdent.Value == "Cos" {
			// Cos takes one numeric argument and always returns Float
			if len(expr.Arguments) != 1 {
				a.addError("function 'Cos' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.FLOAT
			}
			// Analyze the argument and verify it's Integer or Float
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil {
				if argType != types.INTEGER && argType != types.FLOAT {
					a.addError("function 'Cos' expects Integer or Float as argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			// Always returns Float
			return types.FLOAT
		}

		// Tan built-in function (Task 8.185)
		if funcIdent.Value == "Tan" {
			// Tan takes one numeric argument and always returns Float
			if len(expr.Arguments) != 1 {
				a.addError("function 'Tan' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.FLOAT
			}
			// Analyze the argument and verify it's Integer or Float
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil {
				if argType != types.INTEGER && argType != types.FLOAT {
					a.addError("function 'Tan' expects Integer or Float as argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			// Always returns Float
			return types.FLOAT
		}

		// Random built-in function (Task 8.185)
		if funcIdent.Value == "Random" {
			// Random takes no arguments and always returns Float
			if len(expr.Arguments) != 0 {
				a.addError("function 'Random' expects no arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
			}
			// Always returns Float
			return types.FLOAT
		}

		// Randomize built-in procedure (Task 8.185)
		if funcIdent.Value == "Randomize" {
			// Randomize takes no arguments and returns nothing (nil/void)
			if len(expr.Arguments) != 0 {
				a.addError("function 'Randomize' expects no arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
			}
			// Returns nil/void (no meaningful return value)
			return nil
		}

		// Exp built-in function (Task 8.185)
		if funcIdent.Value == "Exp" {
			// Exp takes one numeric argument and always returns Float
			if len(expr.Arguments) != 1 {
				a.addError("function 'Exp' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.FLOAT
			}
			// Analyze the argument and verify it's Integer or Float
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil {
				if argType != types.INTEGER && argType != types.FLOAT {
					a.addError("function 'Exp' expects Integer or Float as argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			// Always returns Float
			return types.FLOAT
		}

		// Ln built-in function (Task 8.185)
		if funcIdent.Value == "Ln" {
			// Ln takes one numeric argument and always returns Float
			if len(expr.Arguments) != 1 {
				a.addError("function 'Ln' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.FLOAT
			}
			// Analyze the argument and verify it's Integer or Float
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil {
				if argType != types.INTEGER && argType != types.FLOAT {
					a.addError("function 'Ln' expects Integer or Float as argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			// Always returns Float
			return types.FLOAT
		}

		// Round built-in function (Task 8.185)
		if funcIdent.Value == "Round" {
			// Round takes one numeric argument and always returns Integer
			if len(expr.Arguments) != 1 {
				a.addError("function 'Round' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the argument and verify it's Integer or Float
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil {
				if argType != types.INTEGER && argType != types.FLOAT {
					a.addError("function 'Round' expects Integer or Float as argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			// Always returns Integer
			return types.INTEGER
		}

		// Trunc built-in function (Task 8.185)
		if funcIdent.Value == "Trunc" {
			// Trunc takes one numeric argument and always returns Integer
			if len(expr.Arguments) != 1 {
				a.addError("function 'Trunc' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the argument and verify it's Integer or Float
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil {
				if argType != types.INTEGER && argType != types.FLOAT {
					a.addError("function 'Trunc' expects Integer or Float as argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			// Always returns Integer
			return types.INTEGER
		}

		// Low built-in function (Task 8.132)
		if funcIdent.Value == "Low" {
			// Low takes one argument (array) and returns an integer
			if len(expr.Arguments) != 1 {
				a.addError("function 'Low' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the argument
			argType := a.analyzeExpression(expr.Arguments[0])
			// Verify it's an array
			if argType != nil {
				if _, isArray := argType.(*types.ArrayType); !isArray {
					a.addError("function 'Low' expects array, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			return types.INTEGER
		}

		// High built-in function (Task 8.133)
		if funcIdent.Value == "High" {
			// High takes one argument (array) and returns an integer
			if len(expr.Arguments) != 1 {
				a.addError("function 'High' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the argument
			argType := a.analyzeExpression(expr.Arguments[0])
			// Verify it's an array
			if argType != nil {
				if _, isArray := argType.(*types.ArrayType); !isArray {
					a.addError("function 'High' expects array, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			return types.INTEGER
		}

		// SetLength built-in function (Task 8.131)
		if funcIdent.Value == "SetLength" {
			// SetLength takes two arguments (array, integer) and returns void
			if len(expr.Arguments) != 2 {
				a.addError("function 'SetLength' expects 2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			// Analyze the first argument (array)
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil {
				if _, isArray := argType.(*types.ArrayType); !isArray {
					a.addError("function 'SetLength' expects array as first argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			// Analyze the second argument (integer)
			lengthType := a.analyzeExpression(expr.Arguments[1])
			if lengthType != nil && lengthType != types.INTEGER {
				a.addError("function 'SetLength' expects integer as second argument, got %s at %s",
					lengthType.String(), expr.Token.Pos.String())
			}
			return types.VOID
		}

		// Add built-in function (Task 8.134)
		if funcIdent.Value == "Add" {
			// Add takes two arguments (array, element) and returns void
			if len(expr.Arguments) != 2 {
				a.addError("function 'Add' expects 2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			// Analyze the first argument (array)
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil {
				if _, isArray := argType.(*types.ArrayType); !isArray {
					a.addError("function 'Add' expects array as first argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			// Analyze the second argument (element to add)
			a.analyzeExpression(expr.Arguments[1])
			return types.VOID
		}

		// Delete built-in function (Task 8.135)
		if funcIdent.Value == "Delete" {
			// Delete takes two arguments (array, index) and returns void
			if len(expr.Arguments) != 2 {
				a.addError("function 'Delete' expects 2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			// Analyze the first argument (array)
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil {
				if _, isArray := argType.(*types.ArrayType); !isArray {
					a.addError("function 'Delete' expects array as first argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			// Analyze the second argument (index - should be integer)
			indexType := a.analyzeExpression(expr.Arguments[1])
			if indexType != nil && indexType != types.INTEGER {
				a.addError("function 'Delete' expects integer as second argument, got %s at %s",
					indexType.String(), expr.Token.Pos.String())
			}
			return types.VOID
		}

		// IntToStr built-in function (Task 8.187)
		if funcIdent.Value == "IntToStr" {
			// IntToStr takes one integer argument and returns a string
			if len(expr.Arguments) != 1 {
				a.addError("function 'IntToStr' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// Analyze the argument and verify it's Integer
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && argType != types.INTEGER {
				a.addError("function 'IntToStr' expects Integer as argument, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.STRING
		}

		// StrToInt built-in function (Task 8.187)
		if funcIdent.Value == "StrToInt" {
			// StrToInt takes one string argument and returns an integer
			if len(expr.Arguments) != 1 {
				a.addError("function 'StrToInt' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the argument and verify it's String
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && argType != types.STRING {
				a.addError("function 'StrToInt' expects String as argument, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.INTEGER
		}

		// FloatToStr built-in function (Task 8.187)
		if funcIdent.Value == "FloatToStr" {
			// FloatToStr takes one float argument and returns a string
			if len(expr.Arguments) != 1 {
				a.addError("function 'FloatToStr' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// Analyze the argument and verify it's Float
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && argType != types.FLOAT {
				a.addError("function 'FloatToStr' expects Float as argument, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.STRING
		}

		// StrToFloat built-in function (Task 8.187)
		if funcIdent.Value == "StrToFloat" {
			// StrToFloat takes one string argument and returns a float
			if len(expr.Arguments) != 1 {
				a.addError("function 'StrToFloat' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.FLOAT
			}
			// Analyze the argument and verify it's String
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && argType != types.STRING {
				a.addError("function 'StrToFloat' expects String as argument, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.FLOAT
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
