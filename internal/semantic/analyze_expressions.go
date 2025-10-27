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

		// Copy built-in function (Task 8.183, Task 9.67)
		if funcIdent.Value == "Copy" {
			// Copy has two overloads:
			// - Copy(arr) - returns copy of array
			// - Copy(str, index, count) - returns substring

			if len(expr.Arguments) == 1 {
				// Copy(arr) - array copy overload
				arrType := a.analyzeExpression(expr.Arguments[0])
				if arrType != nil {
					if arrayType, ok := arrType.(*types.ArrayType); ok {
						// Return the same array type
						return arrayType
					}
					a.addError("function 'Copy' with 1 argument expects array, got %s at %s",
						arrType.String(), expr.Token.Pos.String())
				}
				// Return a generic array type as fallback
				return types.NewDynamicArrayType(types.INTEGER)
			}

			if len(expr.Arguments) != 3 {
				a.addError("function 'Copy' expects either 1 argument (array) or 3 arguments (string), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}

			// Copy(str, index, count) - string copy overload
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

		// Trim built-in function (Task 9.40)
		if funcIdent.Value == "Trim" {
			// Trim takes one string argument and returns a string
			if len(expr.Arguments) != 1 {
				a.addError("function 'Trim' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// Analyze the argument and verify it's a string
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && argType != types.STRING {
				a.addError("function 'Trim' expects string as argument, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.STRING
		}

		// TrimLeft built-in function (Task 9.41)
		if funcIdent.Value == "TrimLeft" {
			// TrimLeft takes one string argument and returns a string
			if len(expr.Arguments) != 1 {
				a.addError("function 'TrimLeft' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// Analyze the argument and verify it's a string
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && argType != types.STRING {
				a.addError("function 'TrimLeft' expects string as argument, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.STRING
		}

		// TrimRight built-in function (Task 9.41)
		if funcIdent.Value == "TrimRight" {
			// TrimRight takes one string argument and returns a string
			if len(expr.Arguments) != 1 {
				a.addError("function 'TrimRight' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// Analyze the argument and verify it's a string
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && argType != types.STRING {
				a.addError("function 'TrimRight' expects string as argument, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.STRING
		}

		// StringReplace built-in function (Task 9.46)
		if funcIdent.Value == "StringReplace" {
			// StringReplace takes 3-4 arguments: str, old, new, [count]
			if len(expr.Arguments) < 3 || len(expr.Arguments) > 4 {
				a.addError("function 'StringReplace' expects 3 or 4 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// First argument: string to search in
			arg1Type := a.analyzeExpression(expr.Arguments[0])
			if arg1Type != nil && arg1Type != types.STRING {
				a.addError("function 'StringReplace' expects string as first argument, got %s at %s",
					arg1Type.String(), expr.Token.Pos.String())
			}
			// Second argument: old substring
			arg2Type := a.analyzeExpression(expr.Arguments[1])
			if arg2Type != nil && arg2Type != types.STRING {
				a.addError("function 'StringReplace' expects string as second argument, got %s at %s",
					arg2Type.String(), expr.Token.Pos.String())
			}
			// Third argument: new substring
			arg3Type := a.analyzeExpression(expr.Arguments[2])
			if arg3Type != nil && arg3Type != types.STRING {
				a.addError("function 'StringReplace' expects string as third argument, got %s at %s",
					arg3Type.String(), expr.Token.Pos.String())
			}
			// Optional fourth argument: count (integer)
			if len(expr.Arguments) == 4 {
				arg4Type := a.analyzeExpression(expr.Arguments[3])
				if arg4Type != nil && arg4Type != types.INTEGER {
					a.addError("function 'StringReplace' expects integer as fourth argument, got %s at %s",
						arg4Type.String(), expr.Token.Pos.String())
				}
			}
			return types.STRING
		}

		// Format built-in function (Task 9.51a)
		if funcIdent.Value == "Format" {
			// Format takes exactly 2 arguments: format string and array of values
			if len(expr.Arguments) != 2 {
				a.addError("Format() expects exactly 2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// First argument: format string (must be String)
			fmtType := a.analyzeExpression(expr.Arguments[0])
			if fmtType != nil && fmtType != types.STRING {
				a.addError("Format() expects string as first argument, got %s at %s",
					fmtType.String(), expr.Token.Pos.String())
			}
			// Second argument: array of values (must be Array type)
			arrType := a.analyzeExpression(expr.Arguments[1])
			if arrType != nil {
				if _, isArray := arrType.(*types.ArrayType); !isArray {
					a.addError("Format() expects array as second argument, got %s at %s",
						arrType.String(), expr.Token.Pos.String())
				}
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

		// Min built-in function (Task 9.54)
		if funcIdent.Value == "Min" {
			// Min takes two numeric arguments and returns the smaller value
			if len(expr.Arguments) != 2 {
				a.addError("function 'Min' expects 2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER // Default to INTEGER on error
			}
			// Analyze both arguments and verify they're Integer or Float
			arg1Type := a.analyzeExpression(expr.Arguments[0])
			arg2Type := a.analyzeExpression(expr.Arguments[1])

			if arg1Type != nil && arg2Type != nil {
				// Verify both are numeric
				if (arg1Type != types.INTEGER && arg1Type != types.FLOAT) ||
					(arg2Type != types.INTEGER && arg2Type != types.FLOAT) {
					a.addError("function 'Min' expects Integer or Float arguments, got %s and %s at %s",
						arg1Type.String(), arg2Type.String(), expr.Token.Pos.String())
					return types.INTEGER
				}
				// If both Integer, return Integer; otherwise return Float
				if arg1Type == types.INTEGER && arg2Type == types.INTEGER {
					return types.INTEGER
				}
				return types.FLOAT
			}
			return types.INTEGER // Default to INTEGER if type is unknown
		}

		// Max built-in function (Task 9.55)
		if funcIdent.Value == "Max" {
			// Max takes two numeric arguments and returns the larger value
			if len(expr.Arguments) != 2 {
				a.addError("function 'Max' expects 2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER // Default to INTEGER on error
			}
			// Analyze both arguments and verify they're Integer or Float
			arg1Type := a.analyzeExpression(expr.Arguments[0])
			arg2Type := a.analyzeExpression(expr.Arguments[1])

			if arg1Type != nil && arg2Type != nil {
				// Verify both are numeric
				if (arg1Type != types.INTEGER && arg1Type != types.FLOAT) ||
					(arg2Type != types.INTEGER && arg2Type != types.FLOAT) {
					a.addError("function 'Max' expects Integer or Float arguments, got %s and %s at %s",
						arg1Type.String(), arg2Type.String(), expr.Token.Pos.String())
					return types.INTEGER
				}
				// If both Integer, return Integer; otherwise return Float
				if arg1Type == types.INTEGER && arg2Type == types.INTEGER {
					return types.INTEGER
				}
				return types.FLOAT
			}
			return types.INTEGER // Default to INTEGER if type is unknown
		}

		// Sqr built-in function (Task 9.57)
		if funcIdent.Value == "Sqr" {
			// Sqr takes one numeric argument and returns x*x, preserving type
			if len(expr.Arguments) != 1 {
				a.addError("function 'Sqr' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER // Default to INTEGER on error
			}
			// Analyze the argument and verify it's Integer or Float
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil {
				if argType != types.INTEGER && argType != types.FLOAT {
					a.addError("function 'Sqr' expects Integer or Float as argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
					return types.INTEGER
				}
				// Return the same type as the input
				return argType
			}
			return types.INTEGER // Default to INTEGER if type is unknown
		}

		// Power built-in function (Task 9.58)
		if funcIdent.Value == "Power" {
			// Power takes two numeric arguments and always returns Float
			if len(expr.Arguments) != 2 {
				a.addError("function 'Power' expects 2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.FLOAT // Default to FLOAT on error
			}
			// Analyze both arguments and verify they're Integer or Float
			arg1Type := a.analyzeExpression(expr.Arguments[0])
			arg2Type := a.analyzeExpression(expr.Arguments[1])

			if arg1Type != nil && arg2Type != nil {
				// Verify both are numeric
				if (arg1Type != types.INTEGER && arg1Type != types.FLOAT) ||
					(arg2Type != types.INTEGER && arg2Type != types.FLOAT) {
					a.addError("function 'Power' expects Integer or Float arguments, got %s and %s at %s",
						arg1Type.String(), arg2Type.String(), expr.Token.Pos.String())
				}
			}
			// Always returns Float
			return types.FLOAT
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

		// RandomInt built-in function (Task 9.63)
		if funcIdent.Value == "RandomInt" {
			// RandomInt takes one Integer argument and returns random Integer in [0, max)
			if len(expr.Arguments) != 1 {
				a.addError("function 'RandomInt' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER // Default to INTEGER on error
			}
			// Analyze argument and verify it's Integer
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && argType != types.INTEGER {
				a.addError("function 'RandomInt' expects Integer argument, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			// Always returns Integer
			return types.INTEGER
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

		// Ceil built-in function (Task 9.60)
		if funcIdent.Value == "Ceil" {
			// Ceil takes one numeric argument and always returns Integer
			if len(expr.Arguments) != 1 {
				a.addError("function 'Ceil' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the argument and verify it's Integer or Float
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil {
				if argType != types.INTEGER && argType != types.FLOAT {
					a.addError("function 'Ceil' expects Integer or Float as argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			// Always returns Integer
			return types.INTEGER
		}

		// Floor built-in function (Task 9.61)
		if funcIdent.Value == "Floor" {
			// Floor takes one numeric argument and always returns Integer
			if len(expr.Arguments) != 1 {
				a.addError("function 'Floor' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the argument and verify it's Integer or Float
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil {
				if argType != types.INTEGER && argType != types.FLOAT {
					a.addError("function 'Floor' expects Integer or Float as argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			// Always returns Integer
			return types.INTEGER
		}

		// Low built-in function (Task 8.132, extended in Task 9.31)
		if funcIdent.Value == "Low" {
			// Low takes one argument (array or enum) and returns an integer (for arrays) or enum value (for enums)
			if len(expr.Arguments) != 1 {
				a.addError("function 'Low' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the argument
			argType := a.analyzeExpression(expr.Arguments[0])
			// Verify it's an array or enum
			if argType != nil {
				if _, isArray := argType.(*types.ArrayType); isArray {
					// For arrays, return Integer
					return types.INTEGER
				}
				if enumType, isEnum := argType.(*types.EnumType); isEnum {
					// For enums, return the same enum type
					return enumType
				}
				// Neither array nor enum
				a.addError("function 'Low' expects array or enum, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.INTEGER
		}

		// High built-in function (Task 8.133, extended in Task 9.32)
		if funcIdent.Value == "High" {
			// High takes one argument (array or enum) and returns an integer (for arrays) or enum value (for enums)
			if len(expr.Arguments) != 1 {
				a.addError("function 'High' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the argument
			argType := a.analyzeExpression(expr.Arguments[0])
			// Verify it's an array or enum
			if argType != nil {
				if _, isArray := argType.(*types.ArrayType); isArray {
					// For arrays, return Integer
					return types.INTEGER
				}
				if enumType, isEnum := argType.(*types.EnumType); isEnum {
					// For enums, return the same enum type
					return enumType
				}
				// Neither array nor enum
				a.addError("function 'High' expects array or enum, got %s at %s",
					argType.String(), expr.Token.Pos.String())
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

		// Delete built-in function (Tasks 8.135, 9.44 - overloaded)
		// Delete(array, index) - for arrays (2 args)
		// Delete(string, pos, count) - for strings (3 args)
		if funcIdent.Value == "Delete" {
			if len(expr.Arguments) == 2 {
				// Array delete: Delete(array, index)
				argType := a.analyzeExpression(expr.Arguments[0])
				if argType != nil {
					if _, isArray := argType.(*types.ArrayType); !isArray {
						a.addError("function 'Delete' expects array as first argument for 2-argument form, got %s at %s",
							argType.String(), expr.Token.Pos.String())
					}
				}
				indexType := a.analyzeExpression(expr.Arguments[1])
				if indexType != nil && indexType != types.INTEGER {
					a.addError("function 'Delete' expects integer as second argument, got %s at %s",
						indexType.String(), expr.Token.Pos.String())
				}
				return types.VOID
			} else if len(expr.Arguments) == 3 {
				// String delete: Delete(string, pos, count)
				if _, ok := expr.Arguments[0].(*ast.Identifier); !ok {
					a.addError("function 'Delete' first argument must be a variable at %s",
						expr.Token.Pos.String())
				} else {
					strType := a.analyzeExpression(expr.Arguments[0])
					if strType != nil && strType != types.STRING {
						a.addError("function 'Delete' first argument must be String for 3-argument form, got %s at %s",
							strType.String(), expr.Token.Pos.String())
					}
				}
				posType := a.analyzeExpression(expr.Arguments[1])
				if posType != nil && posType != types.INTEGER {
					a.addError("function 'Delete' second argument must be Integer, got %s at %s",
						posType.String(), expr.Token.Pos.String())
				}
				countType := a.analyzeExpression(expr.Arguments[2])
				if countType != nil && countType != types.INTEGER {
					a.addError("function 'Delete' third argument must be Integer, got %s at %s",
						countType.String(), expr.Token.Pos.String())
				}
				return types.VOID
			} else {
				a.addError("function 'Delete' expects 2 or 3 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
		}

		// IntToStr built-in function (Task 8.187)
		if funcIdent.Value == "IntToStr" {
			// IntToStr takes one integer argument and returns a string
			if len(expr.Arguments) != 1 {
				a.addError("function 'IntToStr' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// Analyze the argument and verify it's Integer or a subrange of Integer
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && argType != types.INTEGER {
				// Check if it's a subrange type with Integer base
				if subrange, ok := argType.(*types.SubrangeType); ok {
					if subrange.BaseType != types.INTEGER {
						a.addError("function 'IntToStr' expects Integer as argument, got %s at %s",
							argType.String(), expr.Token.Pos.String())
					}
				} else {
					a.addError("function 'IntToStr' expects Integer as argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
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

		// Inc built-in procedure (Task 9.24)
		if funcIdent.Value == "Inc" {
			// Inc takes 1-2 arguments: variable and optional delta
			if len(expr.Arguments) < 1 || len(expr.Arguments) > 2 {
				a.addError("function 'Inc' expects 1-2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			// First argument must be a variable (Identifier)
			if _, ok := expr.Arguments[0].(*ast.Identifier); !ok {
				a.addError("function 'Inc' first argument must be a variable at %s",
					expr.Token.Pos.String())
			} else {
				// Analyze the variable to get its type
				varType := a.analyzeExpression(expr.Arguments[0])
				// Must be Integer or Enum
				if varType != nil {
					if varType != types.INTEGER {
						if _, isEnum := varType.(*types.EnumType); !isEnum {
							a.addError("function 'Inc' expects Integer or Enum variable, got %s at %s",
								varType.String(), expr.Token.Pos.String())
						}
					}
				}
			}
			// If there's a second argument (delta), it must be Integer
			if len(expr.Arguments) == 2 {
				deltaType := a.analyzeExpression(expr.Arguments[1])
				if deltaType != nil && deltaType != types.INTEGER {
					a.addError("function 'Inc' delta must be Integer, got %s at %s",
						deltaType.String(), expr.Token.Pos.String())
				}
			}
			return types.VOID
		}

		// Dec built-in procedure (Task 9.25 - not yet implemented in interpreter)
		if funcIdent.Value == "Dec" {
			// Dec takes 1-2 arguments: variable and optional delta
			if len(expr.Arguments) < 1 || len(expr.Arguments) > 2 {
				a.addError("function 'Dec' expects 1-2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			// First argument must be a variable (Identifier)
			if _, ok := expr.Arguments[0].(*ast.Identifier); !ok {
				a.addError("function 'Dec' first argument must be a variable at %s",
					expr.Token.Pos.String())
			} else {
				// Analyze the variable to get its type
				varType := a.analyzeExpression(expr.Arguments[0])
				// Must be Integer or Enum
				if varType != nil {
					if varType != types.INTEGER {
						if _, isEnum := varType.(*types.EnumType); !isEnum {
							a.addError("function 'Dec' expects Integer or Enum variable, got %s at %s",
								varType.String(), expr.Token.Pos.String())
						}
					}
				}
			}
			// If there's a second argument (delta), it must be Integer
			if len(expr.Arguments) == 2 {
				deltaType := a.analyzeExpression(expr.Arguments[1])
				if deltaType != nil && deltaType != types.INTEGER {
					a.addError("function 'Dec' delta must be Integer, got %s at %s",
						deltaType.String(), expr.Token.Pos.String())
				}
			}
			return types.VOID
		}

		// Succ built-in function (Task 9.28)
		if funcIdent.Value == "Succ" {
			// Succ takes 1 argument: ordinal value
			if len(expr.Arguments) != 1 {
				a.addError("function 'Succ' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the argument to get its type
			argType := a.analyzeExpression(expr.Arguments[0])
			// Must be Integer or Enum
			if argType != nil {
				if argType == types.INTEGER {
					return types.INTEGER
				}
				if enumType, isEnum := argType.(*types.EnumType); isEnum {
					return enumType
				}
				a.addError("function 'Succ' expects Integer or Enum, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.INTEGER
		}

		// Pred built-in function (Task 9.29)
		if funcIdent.Value == "Pred" {
			// Pred takes 1 argument: ordinal value
			if len(expr.Arguments) != 1 {
				a.addError("function 'Pred' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the argument to get its type
			argType := a.analyzeExpression(expr.Arguments[0])
			// Must be Integer or Enum
			if argType != nil {
				if argType == types.INTEGER {
					return types.INTEGER
				}
				if enumType, isEnum := argType.(*types.EnumType); isEnum {
					return enumType
				}
				a.addError("function 'Pred' expects Integer or Enum, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.INTEGER
		}

		// Assert built-in procedure (Task 9.36)
		if funcIdent.Value == "Assert" {
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

		// Insert built-in procedure (Task 9.43)
		if funcIdent.Value == "Insert" {
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
