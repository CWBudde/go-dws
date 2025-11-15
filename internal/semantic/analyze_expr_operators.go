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

	// Task 9.6: Handle ClassName identifier in method contexts
	// ClassName is a built-in property available on all objects (inherited from TObject)
	// When used as an identifier, it returns the class name as a String
	if strings.EqualFold(ident.Value, "ClassName") && a.currentClass != nil {
		return types.STRING
	}

	// Task 9.7: Handle ClassType identifier in method contexts
	// ClassType is a built-in property that returns the metaclass reference
	if strings.EqualFold(ident.Value, "ClassType") && a.currentClass != nil {
		return types.NewClassOfType(a.currentClass)
	}

	sym, ok := a.symbols.Resolve(ident.Value)
	if !ok {
		// Task 9.285: Use lowercase for case-insensitive lookup
		// Task 9.73.5: When a class name is used as an identifier in expressions,
		// it should be treated as a metaclass reference (class of ClassName)
		if classType, exists := a.classes[strings.ToLower(ident.Value)]; exists {
			// Return ClassOfType (metaclass) instead of ClassType
			// This allows: var meta: class of TBase; meta := TBase;
			return &types.ClassOfType{ClassType: classType}
		}
		if a.currentClass != nil && !a.inClassMethod {
			// Task 9.32b/9.32c: Check if identifier is a field of the current class (implicit Self, includes inherited)
			// NOTE: This only applies to instance methods, NOT class methods (static methods)
			if fieldType, exists := a.currentClass.GetField(ident.Value); exists {
				// Check field visibility
				fieldOwner := a.getFieldOwner(a.currentClass, ident.Value)
				if fieldOwner != nil {
					// Use lowercase for case-insensitive lookup
					lowerFieldName := strings.ToLower(ident.Value)
					visibility, hasVisibility := fieldOwner.FieldVisibility[lowerFieldName]
					if hasVisibility && !a.checkVisibility(fieldOwner, visibility, ident.Value, "field") {
						visibilityStr := ast.Visibility(visibility).String()
						a.addError("cannot access %s field '%s' at %s",
							visibilityStr, ident.Value, ident.Token.Pos.String())
						return nil
					}
				}
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
				// In DWScript/Pascal, parameterless methods can be called without parentheses
				// When referenced as an identifier, they should be treated as implicit calls
				if len(methodType.Parameters) == 0 {
					// Implicit call - return the method's return type
					if methodType.ReturnType == nil {
						// Procedure (no return value)
						return types.VOID
					}
					return methodType.ReturnType
				}

				// Methods with parameters cannot be called without parentheses
				// This is an error - the user must provide arguments
				a.addError("method '%s' requires %d argument(s) at %s",
					ident.Value, len(methodType.Parameters), ident.Token.Pos.String())
				return nil
			}
		}

		// Task 9.2: Check if identifier is a class constant (accessible from both instance and class methods)
		// Class constants should be accessible from anywhere within the class, unlike fields which are
		// only accessible from instance methods (not class methods)
		if a.currentClass != nil {
			// Check current class and all parent classes for constants
			for class := a.currentClass; class != nil; class = class.Parent {
				for constName, constType := range class.ConstantTypes {
					if strings.EqualFold(constName, ident.Value) {
						// Check visibility
						constantOwner := a.getConstantOwner(a.currentClass, constName)
						if constantOwner != nil {
							visibility, hasVisibility := constantOwner.ConstantVisibility[constName]
							if hasVisibility && !a.checkVisibility(constantOwner, visibility, ident.Value, "constant") {
								visibilityStr := ast.Visibility(visibility).String()
								a.addError("cannot access %s constant '%s' at %s",
									visibilityStr, ident.Value, ident.Token.Pos.String())
								return nil
							}
						}
						return constType
					}
				}
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

	// Special handling for coalesce operator (??)
	// Both operands must have compatible types; result type is the common type
	if operator == "??" {
		leftType := a.analyzeExpression(expr.Left)
		rightType := a.analyzeExpression(expr.Right)

		if leftType == nil || rightType == nil {
			return nil
		}

		// Check if types are compatible (either equal or one can be assigned to the other)
		if leftType.Equals(rightType) {
			return leftType
		}

		// Check if right can be assigned to left
		if a.canAssign(leftType, rightType) {
			return leftType
		}

		// Check if left can be assigned to right
		if a.canAssign(rightType, leftType) {
			return rightType
		}

		// Handle numeric type promotion (Integer ?? Float -> Float)
		if types.IsNumericType(leftType) && types.IsNumericType(rightType) {
			return types.PromoteTypes(leftType, rightType)
		}

		a.addError("incompatible types in coalesce operator: %s and %s at %s",
			leftType.String(), rightType.String(), expr.Token.Pos.String())
		return nil
	}

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
	if operator == "shl" || operator == "shr" || operator == "sar" {
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

	// Handle 'in' operator for set/string/array membership
	if operator == "in" {
		// Right operand can be:
		// 1. Set type - for set membership
		// 2. String type - for character/substring membership
		// 3. Array type - for array element membership

		// Check for set type
		if rightSetType, isSet := rightType.(*types.SetType); isSet {
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

		// Check for string type - allows character/substring membership
		if rightType == types.STRING {
			// Left operand should be String (character or substring)
			if leftType != types.STRING {
				a.addError("'in' operator with String requires String as left operand, got %s at %s",
					leftType.String(), expr.Token.Pos.String())
				return nil
			}
			// 'in' operator returns Boolean
			return types.BOOLEAN
		}

		// Check for array type - allows element membership
		if _, isArray := rightType.(*types.ArrayType); isArray {
			// Left operand can be any type - will be checked at runtime
			// 'in' operator returns Boolean
			return types.BOOLEAN
		}

		// If none of the above, error
		a.addError("'in' operator requires set, string, or array as right operand, got %s at %s",
			rightType.String(), expr.Token.Pos.String())
		return nil
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
	// In DWScript, 'not' works on Boolean (logical NOT), Integer (bitwise NOT), and Variant
	if operator == "not" {
		if operandType.Equals(types.BOOLEAN) {
			return types.BOOLEAN
		}
		if operandType.Equals(types.INTEGER) {
			return types.INTEGER
		}
		// Task 9.35: Support Variant→Boolean implicit conversion in not operator
		if operandType.Equals(types.VARIANT) {
			return types.VARIANT
		}
		a.addError("unary not requires Boolean, Integer, or Variant operand, got %s at %s",
			operandType.String(), expr.Token.Pos.String())
		return nil
	}

	a.addError("unknown unary operator: %s at %s", operator, expr.Token.Pos.String())
	return nil
}

// analyzeCallExpression analyzes a function call and returns its type
// Task 9.161: Semantic analysis for inherited keyword
