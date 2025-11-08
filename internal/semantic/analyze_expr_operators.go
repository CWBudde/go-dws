package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Expression Analysis
// ============================================================================
func (a *Analyzer) analyzeIdentifier(ident *ast.Identifier) types.Type {
	// Task 9.133: Handle built-in type names as type meta-values
	// These identifiers represent type names that can be used as runtime values
	// (e.g., High(Integer), Low(Boolean))
	switch ident.Value {
	case "Integer":
		return types.INTEGER
	case "Float":
		return types.FLOAT
	case "Boolean":
		return types.BOOLEAN
	case "String":
		return types.STRING
	}

	// Task 9.161: Handle user-defined enum type names as type meta-values
	// This allows enum type names to be used in expressions like High(TColor)
	if enumType, exists := a.enums[ident.Value]; exists {
		return enumType
	}

	// Handle built-in ExceptObject variable
	// ExceptObject is a global variable that holds the current exception (or nil)
	if ident.Value == "ExceptObject" {
		// ExceptObject is always of type Exception (the base exception class)
		// Task 9.285: Use lowercase for case-insensitive lookup
		if exceptionClass, exists := a.classes["exception"]; exists {
			return exceptionClass
		}
		// If Exception class doesn't exist (shouldn't happen), return nil
		a.addError("internal error: Exception class not found")
		return nil
	}

	sym, ok := a.symbols.Resolve(ident.Value)
	if !ok {
		// Task 9.285: Use lowercase for case-insensitive lookup
		if classType, exists := a.classes[strings.ToLower(ident.Value)]; exists {
			return classType
		}
		if a.currentClass != nil && !a.inClassMethod {
			// Task 9.32b/9.32c: Check if identifier is a field of the current class (implicit Self)
			// NOTE: This only applies to instance methods, NOT class methods (static methods)
			if fieldType, exists := a.currentClass.Fields[ident.Value]; exists {
				return fieldType
			}

			// Task 9.32b/9.32c: Check if identifier is a property of the current class (implicit Self)
			// DWScript is case-insensitive, so we need to search all properties
			// Also search parent class hierarchy
			for class := a.currentClass; class != nil; class = class.Parent {
				for propName, propInfo := range class.Properties {
					if strings.EqualFold(propName, ident.Value) {
						// Task 9.49: Check for circular reference in property expressions
						if a.inPropertyExpr && strings.EqualFold(propName, a.currentProperty) {
							a.addError("property '%s' cannot be read-accessed at %s", ident.Value, ident.Token.Pos.String())
							return nil
						}

						// For write-only properties, check if read access is defined
						if propInfo.ReadKind == types.PropAccessNone {
							a.addError("property '%s' is write-only at %s", ident.Value, ident.Token.Pos.String())
							return nil
						}
						return propInfo.Type
					}
				}
			}

			if owner := a.getFieldOwner(a.currentClass.Parent, ident.Value); owner != nil {
				if visibility, ok := owner.FieldVisibility[ident.Value]; ok && visibility == int(ast.VisibilityPrivate) {
					a.addError("cannot access private field '%s' of class '%s' at %s",
						ident.Value, owner.Name, ident.Token.Pos.String())
					return nil
				}
			}

			// Task 9.173: Check if identifier refers to a method of the current class
			// This allows method pointers to be passed as function arguments
			methodType, found := a.currentClass.GetMethod(ident.Value)
			if found {
				// Check method visibility
				methodOwner := a.getMethodOwner(a.currentClass, ident.Value)
				if methodOwner != nil {
					visibility, hasVisibility := methodOwner.MethodVisibility[ident.Value]
					if hasVisibility && !a.checkVisibility(methodOwner, visibility, ident.Value, "method") {
						visibilityStr := ast.Visibility(visibility).String()
						a.addError("cannot access %s method '%s' of class '%s' at %s",
							visibilityStr, ident.Value, methodOwner.Name, ident.Token.Pos.String())
						return nil
					}
				}
				// Return the method as a method pointer type (not just a function type)
				// This allows it to be passed as a function pointer parameter
				return types.NewMethodPointerType(methodType.Parameters, methodType.ReturnType)
			}
		}

		// Task 9.132: Check if this is a built-in function used without parentheses
		// In DWScript, built-in functions like PrintLn can be called without parentheses
		// The semantic analyzer should allow this and treat them as procedure calls
		if a.isBuiltinFunction(ident.Value) {
			// Return Void type for built-in procedures (or appropriate type for functions)
			// For simplicity, we'll return VOID type which means "any" - the interpreter will handle it
			return types.VOID
		}

		a.addError("undefined variable '%s' at %s", ident.Value, ident.Token.Pos.String())
		return nil
	}

	// Task 9.228: When a function is referenced as a value (not called),
	// implicitly convert it to a function pointer type.
	// This allows functions to be passed as arguments to higher-order functions.
	// Example: PrintLn(First(Second)) where Second is a function
	if funcType, ok := sym.Type.(*types.FunctionType); ok {
		// Convert function type to function pointer type
		// Note: FunctionType uses VOID for procedures, but FunctionPointerType uses nil
		returnType := funcType.ReturnType
		if funcType.IsProcedure() {
			returnType = nil
		}
		return types.NewFunctionPointerType(funcType.Parameters, returnType)
	}

	return sym.Type
}

// analyzeBinaryExpression analyzes a binary expression and returns its type
func (a *Analyzer) analyzeBinaryExpression(expr *ast.BinaryExpression) types.Type {
	operator := expr.Operator

	// Task 9.226: Special handling for IN operator
	// For the IN operator, analyze the right operand with expected set type context
	// so that array literals like [1, 2, 3] can be converted to set literals
	var leftType, rightType types.Type
	if operator == "in" {
		// Analyze left operand first to infer the set element type
		leftType = a.analyzeExpression(expr.Left)
		if leftType == nil {
			return nil
		}

		// For IN operator, the right operand should be a set
		// If left is an ordinal type, expect a set of that type
		var expectedSetType types.Type
		if types.IsOrdinalType(leftType) {
			expectedSetType = types.NewSetType(leftType)
		}

		// Analyze right operand with expected set type
		rightType = a.analyzeExpressionWithExpectedType(expr.Right, expectedSetType)
		if rightType == nil {
			return nil
		}
	} else {
		// For other operators, analyze both operands without type context
		leftType = a.analyzeExpression(expr.Left)
		rightType = a.analyzeExpression(expr.Right)

		if leftType == nil || rightType == nil {
			// Errors already reported
			return nil
		}
	}

	if sig, ok := a.resolveBinaryOperator(operator, leftType, rightType); ok {
		if sig.ResultType != nil {
			return sig.ResultType
		}
		return types.VOID
	}

	// Handle arithmetic operators
	if operator == "+" || operator == "-" || operator == "*" || operator == "/" {
		// Check for set operations first
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

	// Handle bitwise shift operators
	if operator == "shl" || operator == "shr" {
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

	// Handle logical/bitwise operators (and, or, xor)
	// These operators work on both Boolean (logical) and Integer (bitwise) types
	if operator == "and" || operator == "or" || operator == "xor" {
		// Both operands must be Boolean or both must be Integer
		if leftType.Equals(types.BOOLEAN) && rightType.Equals(types.BOOLEAN) {
			return types.BOOLEAN
		}
		if leftType.Equals(types.INTEGER) && rightType.Equals(types.INTEGER) {
			return types.INTEGER
		}
		a.addError("operator %s requires both operands to be Boolean or both Integer, got %s and %s at %s",
			operator, leftType.String(), rightType.String(), expr.Token.Pos.String())
		return nil
	}

	// Handle 'in' operator for set membership
	if operator == "in" {
		// Right operand must be a set type
		rightSetType, isSet := rightType.(*types.SetType)
		if !isSet {
			a.addError("'in' operator requires set as right operand, got %s at %s",
				rightType.String(), expr.Token.Pos.String())
			return nil
		}

		// Left operand must be an ordinal type matching the set's element type
		if !types.IsOrdinalType(leftType) {
			a.addError("'in' operator requires ordinal value as left operand, got %s at %s",
				leftType.String(), expr.Token.Pos.String())
			return nil
		}

		// Element types must match (resolve underlying types for comparison)
		leftResolved := types.GetUnderlyingType(leftType)
		rightResolved := types.GetUnderlyingType(rightSetType.ElementType)

		if !leftResolved.Equals(rightResolved) {
			a.addError("type mismatch in 'in' operator: %s is not compatible with set of %s at %s",
				leftType.String(), rightSetType.ElementType.String(), expr.Token.Pos.String())
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

	// Handle logical/bitwise not
	// In DWScript, 'not' works on both Boolean (logical NOT) and Integer (bitwise NOT)
	if operator == "not" {
		if operandType.Equals(types.BOOLEAN) {
			return types.BOOLEAN
		}
		if operandType.Equals(types.INTEGER) {
			return types.INTEGER
		}
		a.addError("unary not requires Boolean or Integer operand, got %s at %s",
			operandType.String(), expr.Token.Pos.String())
		return nil
	}

	a.addError("unknown unary operator: %s at %s", operator, expr.Token.Pos.String())
	return nil
}

// analyzeCallExpression analyzes a function call and returns its type
// Task 9.161: Semantic analysis for inherited keyword
