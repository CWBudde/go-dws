package semantic

import (
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	ident "github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Expression Analysis
// ============================================================================

func (a *Analyzer) analyzeIdentifier(identifier *ast.Identifier) types.Type {
	// Task 9.133: Handle built-in type names as type meta-values
	// These identifiers represent type names that can be used as runtime values
	// (e.g., High(Integer), Low(Boolean))
	switch identifier.Value {
	case "Integer":
		return types.INTEGER
	case "Float":
		return types.FLOAT
	case "Boolean":
		return types.BOOLEAN
	case "String":
		return types.STRING
	}

	// This allows enum type names to be used in expressions like High(TColor)
	if enumType := a.getEnumType(identifier.Value); enumType != nil {
		return enumType
	}

	// Handle built-in ExceptObject variable
	// ExceptObject is a global variable that holds the current exception (or nil)
	if identifier.Value == "ExceptObject" {
		// ExceptObject is always of type Exception (the base exception class)
		// Task 6.1.1.3: Use TypeRegistry for unified type lookup
		if exceptionClass := a.getClassType("Exception"); exceptionClass != nil {
			return exceptionClass
		}
		// If Exception class doesn't exist (shouldn't happen), return nil
		a.addError("internal error: Exception class not found")
		return nil
	}

	// Task 9.6: Handle ClassName identifier in method contexts
	// ClassName is a built-in property available on all objects (inherited from TObject)
	// When used as an identifier, it returns the class name as a String
	if ident.Equal(identifier.Value, "ClassName") && a.currentClass != nil {
		if identifier.Value != "ClassName" {
			a.addCaseMismatchHint(identifier.Value, "ClassName", identifier.Token.Pos)
		}
		return types.STRING
	}

	// Task 9.7: Handle ClassType identifier in method contexts
	// ClassType is a built-in property that returns the metaclass reference
	if ident.Equal(identifier.Value, "ClassType") && a.currentClass != nil {
		if identifier.Value != "ClassType" {
			a.addCaseMismatchHint(identifier.Value, "ClassType", identifier.Token.Pos)
		}
		return types.NewClassOfType(a.currentClass)
	}

	sym, ok := a.symbols.Resolve(identifier.Value)
	if !ok {
		// Task 9.73.5: When a class name is used as an identifier in expressions,
		// it should be treated as a metaclass reference (class of ClassName)
		// Task 6.1.1.3: Use TypeRegistry for unified type lookup
		if classType := a.getClassType(identifier.Value); classType != nil {
			// Return ClassOfType (metaclass) instead of ClassType
			// This allows: var meta: class of TBase; meta := TBase;
			return &types.ClassOfType{ClassType: classType}
		}
		if a.currentClass != nil && !a.inClassMethod {
			// Task 9.32b/9.32c: Check if identifier is a field of the current class (implicit Self, includes inherited)
			// NOTE: This only applies to instance methods, NOT class methods (static methods)
			if fieldType, exists := a.currentClass.GetField(identifier.Value); exists {
				// Check field visibility
				fieldOwner := a.getFieldOwner(a.currentClass, identifier.Value)
				if fieldOwner != nil {
					lowerFieldName := ident.Normalize(identifier.Value)
					visibility, hasVisibility := fieldOwner.FieldVisibility[lowerFieldName]
					if hasVisibility && !a.checkVisibility(fieldOwner, visibility, identifier.Value, "field") {
						visibilityStr := ast.Visibility(visibility).String()
						a.addError("cannot access %s field '%s' at %s",
							visibilityStr, identifier.Value, identifier.Token.Pos.String())
						return nil
					}
				}
				return fieldType
			}

			// Check if identifier is a property of the current class (implicit Self)
			// DWScript is case-insensitive, so we need to search all properties
			// Also search parent class hierarchy
			for class := a.currentClass; class != nil; class = class.Parent {
				for propName, propInfo := range class.Properties {
					if ident.Equal(propName, identifier.Value) {
						// Task 9.49: Check for circular reference in property expressions
						if a.inPropertyExpr && ident.Equal(propName, a.currentProperty) {
							a.addError("property '%s' cannot be read-accessed at %s", identifier.Value, identifier.Token.Pos.String())
							return nil
						}

						// For write-only properties, check if read access is defined
						if propInfo.ReadKind == types.PropAccessNone {
							a.addError("property '%s' is write-only at %s", identifier.Value, identifier.Token.Pos.String())
							return nil
						}
						return propInfo.Type
					}
				}
			}

			// Task 9.173: Check if identifier refers to a method of the current class
			// This allows method pointers to be passed as function arguments
			methodType, found := a.currentClass.GetMethod(identifier.Value)
			if found {
				// Check method visibility
				methodOwner := a.getMethodOwner(a.currentClass, identifier.Value)
				if methodOwner != nil {
					visibility, hasVisibility := methodOwner.MethodVisibility[identifier.Value]
					if hasVisibility && !a.checkVisibility(methodOwner, visibility, identifier.Value, "method") {
						visibilityStr := ast.Visibility(visibility).String()
						a.addError("cannot access %s method '%s' of class '%s' at %s",
							visibilityStr, identifier.Value, methodOwner.Name, identifier.Token.Pos.String())
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
				// Return a method pointer type for deferred invocation
				return types.NewMethodPointerType(methodType.Parameters, methodType.ReturnType)
			}
		}

		// Task 9.2: Check if identifier is a class constant (accessible from both instance and class methods)
		// Class constants should be accessible from anywhere within the class, unlike fields which are
		// only accessible from instance methods (not class methods)
		if a.currentClass != nil {
			if constType := a.findClassConstantWithVisibility(a.currentClass, identifier.Value, identifier.Token.Pos.String()); constType != nil {
				return constType
			}
		}

		// Task 9.132: Check if this is a built-in function used without parentheses
		// In DWScript, built-in functions like PrintLn can be called without parentheses
		// The semantic analyzer should allow this and treat them as procedure calls
		if a.isBuiltinFunction(identifier.Value) {
			// Return Void type for built-in procedures (or appropriate type for functions)
			// For simplicity, we'll return VOID type which means "any" - the interpreter will handle it
			return types.VOID
		}

		a.addError("%s", errors.FormatUnknownName(identifier.Value, identifier.Token.Pos.Line, identifier.Token.Pos.Column))
		return nil
	}

	// Emit a hint when the identifier casing doesn't match its declaration.
	if sym.Name != "" && sym.Name != identifier.Value && ident.Equal(sym.Name, identifier.Value) {
		a.addCaseMismatchHint(identifier.Value, sym.Name, identifier.Token.Pos)
	}

	// Task 9.228 + Function Name Alias: Handle function/procedure references
	// When a function is referenced as a value (not called), there are two cases:
	// 1. Inside the function's own body: function name is an alias for Result variable
	// 2. Outside the function body: convert to function pointer type
	if funcType, ok := sym.Type.(*types.FunctionType); ok {
		// Check if we're inside the function's own body (function name = Result alias)
		// In DWScript, the function name can be used as an alias for the Result variable
		if a.currentFunction != nil && ident.Equal(a.currentFunction.Name.Value, identifier.Value) {
			// Inside the function's own body - function name is an alias for Result

			// Procedures (no return type) don't have a Result variable
			// Return nil to trigger "unknown name" error for reads
			if funcType.IsProcedure() {
				return nil
			}

			// For functions with return types, return the return type
			// This allows: GetValue := GetValue + 1
			return funcType.ReturnType
		}

		// Outside the function body - convert to function pointer type (existing behavior)
		// This allows functions to be passed as arguments to higher-order functions.
		// Example: PrintLn(First(Second)) where Second is a function
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

		// Task 9.4.2: Check if either operand is Variant
		leftIsVariant := leftType == types.VARIANT
		rightIsVariant := rightType == types.VARIANT

		// Special case: + can also concatenate strings
		if operator == "+" && (leftType.Equals(types.STRING) || rightType.Equals(types.STRING)) {
			// Task 9.4.2: Allow Variant in string concatenation
			if !leftIsVariant && !rightIsVariant {
				// String concatenation
				if !leftType.Equals(types.STRING) || !rightType.Equals(types.STRING) {
					a.addError("string concatenation requires both operands to be strings at %s",
						expr.Token.Pos.String())
					return nil
				}
			}
			// If Variant is involved, return Variant; otherwise return STRING
			if leftIsVariant || rightIsVariant {
				return types.VARIANT
			}
			return types.STRING
		}

		// Numeric arithmetic
		// Task 9.4.2: Allow Variant in numeric operations
		if !leftIsVariant && !rightIsVariant {
			if !types.IsNumericType(leftType) || !types.IsNumericType(rightType) {
				a.addError("arithmetic operator %s requires numeric operands, got %s and %s at %s",
					operator, leftType.String(), rightType.String(), expr.Token.Pos.String())
				return nil
			}
		}

		// Type promotion: Integer + Float -> Float
		// Task 9.4.2: If Variant is involved, return Variant
		if leftIsVariant || rightIsVariant {
			return types.VARIANT
		}
		return types.PromoteTypes(leftType, rightType)
	}

	// Handle integer division and modulo
	if operator == "div" || operator == "mod" {
		// Task 9.4.2: Allow Variant in div/mod operations
		leftIsVariant := leftType == types.VARIANT
		rightIsVariant := rightType == types.VARIANT

		if !leftIsVariant && !rightIsVariant {
			if !leftType.Equals(types.INTEGER) || !rightType.Equals(types.INTEGER) {
				a.addError("operator %s requires integer operands, got %s and %s at %s",
					operator, leftType.String(), rightType.String(), expr.Token.Pos.String())
				return nil
			}
		}

		// If Variant is involved, return Variant; otherwise return INTEGER
		if leftIsVariant || rightIsVariant {
			return types.VARIANT
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
		// Task 9.4.1: Allow Variant to be compared with any type
		leftIsVariant := leftType == types.VARIANT
		rightIsVariant := rightType == types.VARIANT

		// For equality, types must be comparable
		if operator == "=" || operator == "<>" {
			// If either operand is Variant, allow the comparison
			if !leftIsVariant && !rightIsVariant {
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
			}
		} else {
			// For ordering, types must be orderable
			// Task 9.4.1: Allow Variant in ordering comparisons
			if !leftIsVariant && !rightIsVariant {
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
		}
		return types.BOOLEAN
	}

	// Handle logical/bitwise operators (and, or, xor)
	// These operators work on Boolean (logical), Integer (bitwise), and Enum types
	if operator == "and" || operator == "or" || operator == "xor" {
		// Task 9.4.2: Allow Variant in logical/bitwise operations
		leftIsVariant := leftType == types.VARIANT
		rightIsVariant := rightType == types.VARIANT

		// If Variant is involved, allow the operation and return Variant
		if leftIsVariant || rightIsVariant {
			return types.VARIANT
		}

		// Both operands must be Boolean or both must be Integer
		if leftType.Equals(types.BOOLEAN) && rightType.Equals(types.BOOLEAN) {
			return types.BOOLEAN
		}
		if leftType.Equals(types.INTEGER) && rightType.Equals(types.INTEGER) {
			return types.INTEGER
		}

		// Task 1.6: Allow boolean operations on enum types (especially flags enums)
		// Check if both operands are the same enum type
		leftEnum, leftIsEnum := leftType.(*types.EnumType)
		rightEnum, rightIsEnum := rightType.(*types.EnumType)
		if leftIsEnum && rightIsEnum {
			// Both operands must be the same enum type
			if leftEnum.Equals(rightEnum) {
				// Return the enum type
				return leftEnum
			}
			a.addError("operator %s requires operands of the same enum type, got %s and %s at %s",
				operator, leftType.String(), rightType.String(), expr.Token.Pos.String())
			return nil
		}

		a.addError("operator %s requires both operands to be Boolean, both Integer, or both the same enum type, got %s and %s at %s",
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
		// Task 9.4.2: Allow Variant in unary numeric operations
		if operandType == types.VARIANT {
			return types.VARIANT
		}

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
