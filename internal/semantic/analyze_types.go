package semantic

import (
	"fmt"
	"math"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	pkgident "github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Type Declaration Analysis
// ============================================================================

// evaluateConstant evaluates a compile-time constant expression.
// Returns the constant value and an error if the expression is not a constant.
func (a *Analyzer) evaluateConstant(expr ast.Expression) (interface{}, error) {
	if expr == nil {
		return nil, fmt.Errorf("nil expression")
	}

	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return int(e.Value), nil

	case *ast.FloatLiteral:
		return e.Value, nil

	case *ast.StringLiteral:
		return e.Value, nil

	case *ast.BooleanLiteral:
		return e.Value, nil

	case *ast.CharLiteral:
		// Convert rune to string (single character)
		return string(e.Value), nil

	case *ast.Identifier:
		// Constant identifier reference
		sym, ok := a.symbols.Resolve(e.Value)
		if !ok {
			return nil, fmt.Errorf("undefined identifier '%s'", e.Value)
		}
		if !sym.IsConst {
			return nil, fmt.Errorf("identifier '%s' is not a constant", e.Value)
		}
		return sym.Value, nil

	case *ast.UnaryExpression:
		// Delegate to evaluateConstantInt for integer unary ops
		if e.Operator == "-" || e.Operator == "+" {
			val, err := a.evaluateConstantInt(expr)
			if err != nil {
				return nil, err
			}
			return val, nil
		}
		return nil, fmt.Errorf("non-constant unary expression")

	case *ast.BinaryExpression:
		// Handle binary expressions - need to check if operands are floats or ints
		leftVal, err := a.evaluateConstant(e.Left)
		if err != nil {
			return nil, err
		}
		rightVal, err := a.evaluateConstant(e.Right)
		if err != nil {
			return nil, err
		}

		// Check if either operand is a string and operator is '+'
		leftStr, leftIsStr := leftVal.(string)
		rightStr, rightIsStr := rightVal.(string)
		if (leftIsStr || rightIsStr) && e.Operator == "+" {
			// String concatenation
			if !leftIsStr {
				return nil, fmt.Errorf("cannot concatenate non-string with string")
			}
			if !rightIsStr {
				return nil, fmt.Errorf("cannot concatenate string with non-string")
			}
			return leftStr + rightStr, nil
		}

		// Check if either operand is a float
		leftFloat, leftIsFloat := leftVal.(float64)
		rightFloat, rightIsFloat := rightVal.(float64)
		leftInt, leftIsInt := leftVal.(int)
		rightInt, rightIsInt := rightVal.(int)

		// Convert to common type
		var left, right float64
		isFloat := leftIsFloat || rightIsFloat || e.Operator == "/"

		if leftIsFloat {
			left = leftFloat
		} else if leftIsInt {
			left = float64(leftInt)
		} else {
			return nil, fmt.Errorf("left operand is not numeric")
		}

		if rightIsFloat {
			right = rightFloat
		} else if rightIsInt {
			right = float64(rightInt)
		} else {
			return nil, fmt.Errorf("right operand is not numeric")
		}

		// Evaluate based on operator
		var result float64
		switch e.Operator {
		case "+":
			result = left + right
		case "-":
			result = left - right
		case "*":
			result = left * right
		case "/":
			if right == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			result = left / right
			isFloat = true // Division always returns float
		case "div":
			if right == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return int(left) / int(right), nil
		case "mod":
			if right == 0 {
				return nil, fmt.Errorf("modulo by zero")
			}
			return int(left) % int(right), nil
		default:
			return nil, fmt.Errorf("non-constant binary operator '%s'", e.Operator)
		}

		// Return appropriate type
		if isFloat {
			return result, nil
		}
		return int(result), nil

	case *ast.CallExpression:

		return a.evaluateConstantFunction(e)

	case *ast.RecordLiteralExpression:

		// Evaluate all field values recursively to ensure they're constants
		constFields := make(map[string]interface{})
		for _, field := range e.Fields {
			fieldVal, err := a.evaluateConstant(field.Value)
			if err != nil {
				return nil, fmt.Errorf("record field '%s' is not constant: %v", field.Name.Value, err)
			}
			constFields[field.Name.Value] = fieldVal
		}
		return constFields, nil

	case *ast.ArrayLiteralExpression:
		// Support array literals in const declarations
		// Evaluate all elements recursively to ensure they're constants
		constElements := make([]interface{}, len(e.Elements))
		for i, elem := range e.Elements {
			elemVal, err := a.evaluateConstant(elem)
			if err != nil {
				return nil, fmt.Errorf("array element %d is not constant: %v", i, err)
			}
			constElements[i] = elemVal
		}
		return constElements, nil

	default:
		return nil, fmt.Errorf("expression is not a compile-time constant")
	}
}

// evaluateConstantInt evaluates a compile-time constant integer expression.
// Returns the integer value and an error if the expression is not a constant.
func (a *Analyzer) evaluateConstantInt(expr ast.Expression) (int, error) {
	if expr == nil {
		return 0, fmt.Errorf("nil expression")
	}

	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		// Direct integer literal
		return int(e.Value), nil

	case *ast.Identifier:
		// Constant identifier reference: size, maxIndex, etc.
		// Look up the constant in the symbol table
		sym, ok := a.symbols.Resolve(e.Value)
		if !ok {
			return 0, fmt.Errorf("undefined identifier '%s'", e.Value)
		}
		if !sym.IsConst {
			return 0, fmt.Errorf("identifier '%s' is not a constant", e.Value)
		}
		// Get the constant value
		if sym.Value == nil {
			return 0, fmt.Errorf("constant '%s' has no value", e.Value)
		}
		// Convert to int
		intVal, ok := sym.Value.(int)
		if !ok {
			return 0, fmt.Errorf("constant '%s' is not an integer", e.Value)
		}
		return intVal, nil

	case *ast.UnaryExpression:
		// Handle negative numbers: -40, -size
		if e.Operator == "-" {
			value, err := a.evaluateConstantInt(e.Right)
			if err != nil {
				return 0, err
			}
			return -value, nil
		}
		if e.Operator == "+" {
			// Unary plus: +5
			return a.evaluateConstantInt(e.Right)
		}
		return 0, fmt.Errorf("non-constant unary expression with operator %s", e.Operator)

	case *ast.BinaryExpression:
		// Handle binary expressions: size - 1, maxIndex + 10, etc.
		// For division operator, we need to handle float conversion
		if e.Operator == "/" {
			// Division returns float, need to convert back to int
			val, err := a.evaluateConstant(expr)
			if err != nil {
				return 0, err
			}
			// Convert to int (truncate)
			switch v := val.(type) {
			case int:
				return v, nil
			case float64:
				return int(v), nil
			default:
				return 0, fmt.Errorf("division result is not numeric")
			}
		}

		left, err := a.evaluateConstantInt(e.Left)
		if err != nil {
			return 0, err
		}
		right, err := a.evaluateConstantInt(e.Right)
		if err != nil {
			return 0, err
		}

		// Evaluate based on operator
		switch e.Operator {
		case "+":
			return left + right, nil
		case "-":
			return left - right, nil
		case "*":
			return left * right, nil
		case "div":
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		case "mod":
			if right == 0 {
				return 0, fmt.Errorf("modulo by zero")
			}
			return left % right, nil
		default:
			return 0, fmt.Errorf("non-constant binary operator '%s'", e.Operator)
		}

	case *ast.CallExpression:

		val, err := a.evaluateConstantFunction(e)
		if err != nil {
			return 0, err
		}
		// Convert to int
		switch v := val.(type) {
		case int:
			return v, nil
		case float64:
			return int(v), nil
		default:
			return 0, fmt.Errorf("function result is not numeric")
		}

	default:
		return 0, fmt.Errorf("expression is not a compile-time constant integer")
	}
}

// evaluateConstantFunction evaluates a compile-time constant function call.
// like High(), Log2(), Floor().
func (a *Analyzer) evaluateConstantFunction(call *ast.CallExpression) (interface{}, error) {
	if call == nil || call.Function == nil {
		return nil, fmt.Errorf("nil function call")
	}

	// Get function name
	funcName := ""
	if ident, ok := call.Function.(*ast.Identifier); ok {
		funcName = ident.Value
	} else {
		return nil, fmt.Errorf("function call is not a compile-time constant")
	}

	funcNameLower := pkgident.Normalize(funcName)

	// Only support compile-time evaluable built-in functions
	switch funcNameLower {
	case "high":
		return a.evaluateConstantHigh(call.Arguments)
	case "low":
		return a.evaluateConstantLow(call.Arguments)
	case "log2":
		return a.evaluateConstantLog2(call.Arguments)
	case "floor":
		return a.evaluateConstantFloor(call.Arguments)
	case "ceil":
		return a.evaluateConstantCeil(call.Arguments)
	case "round":
		return a.evaluateConstantRound(call.Arguments)
	case "ord":
		return a.evaluateConstantOrd(call.Arguments)
	case "chr":
		return a.evaluateConstantChr(call.Arguments)
	default:
		// Type casts with exactly one argument can be compile-time constants
		if len(call.Arguments) == 1 {
			if castValue := a.evaluateConstantTypeCast(funcName, call.Arguments[0]); castValue != nil {
				return castValue, nil
			}
		}
		return nil, fmt.Errorf("function '%s' is not a compile-time constant", funcName)
	}
}

// evaluateConstantHigh evaluates High() at compile time.
func (a *Analyzer) evaluateConstantHigh(args []ast.Expression) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("High() expects exactly 1 argument")
	}

	// Check if it's a type identifier (e.g., High(Integer))
	if ident, ok := args[0].(*ast.Identifier); ok {
		typeName := pkgident.Normalize(ident.Value)
		switch typeName {
		case "integer":
			return math.MaxInt64, nil
		case "boolean":
			return true, nil
		default:
			// Check if it's a user-defined type
			sym, ok := a.symbols.Resolve(ident.Value)
			if !ok {
				return nil, fmt.Errorf("undefined type '%s'", ident.Value)
			}
			// Handle enum types
			if sym.Type != nil {
				if enumType, ok := sym.Type.(*types.EnumType); ok {
					if len(enumType.OrderedNames) == 0 {
						return nil, fmt.Errorf("enum type '%s' has no values", ident.Value)
					}
					lastValueName := enumType.OrderedNames[len(enumType.OrderedNames)-1]
					return enumType.Values[lastValueName], nil
				}
			}
			return nil, fmt.Errorf("High() not supported for type %s", ident.Value)
		}
	}

	return nil, fmt.Errorf("High() argument must be a type name")
}

// evaluateConstantLow evaluates Low() at compile time.
func (a *Analyzer) evaluateConstantLow(args []ast.Expression) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Low() expects exactly 1 argument")
	}

	// Check if it's a type identifier (e.g., Low(Integer))
	if ident, ok := args[0].(*ast.Identifier); ok {
		typeName := pkgident.Normalize(ident.Value)
		switch typeName {
		case "integer":
			return math.MinInt64, nil
		case "boolean":
			return false, nil
		default:
			// Check if it's a user-defined type
			sym, ok := a.symbols.Resolve(ident.Value)
			if !ok {
				return nil, fmt.Errorf("undefined type '%s'", ident.Value)
			}
			// Handle enum types
			if sym.Type != nil {
				if enumType, ok := sym.Type.(*types.EnumType); ok {
					if len(enumType.OrderedNames) == 0 {
						return nil, fmt.Errorf("enum type '%s' has no values", ident.Value)
					}
					firstValueName := enumType.OrderedNames[0]
					return enumType.Values[firstValueName], nil
				}
			}
			return nil, fmt.Errorf("Low() not supported for type %s", ident.Value)
		}
	}

	return nil, fmt.Errorf("Low() argument must be a type name")
}

// evaluateConstantLog2 evaluates Log2() at compile time.
func (a *Analyzer) evaluateConstantLog2(args []ast.Expression) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Log2() expects exactly 1 argument")
	}

	// Evaluate argument
	val, err := a.evaluateConstant(args[0])
	if err != nil {
		return nil, err
	}

	// Convert to float64
	var floatVal float64
	switch v := val.(type) {
	case int:
		floatVal = float64(v)
	case float64:
		floatVal = v
	default:
		return nil, fmt.Errorf("Log2() expects numeric argument")
	}

	if floatVal <= 0 {
		return nil, fmt.Errorf("Log2() of non-positive number")
	}

	return math.Log2(floatVal), nil
}

// evaluateConstantFloor evaluates Floor() at compile time.
func (a *Analyzer) evaluateConstantFloor(args []ast.Expression) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Floor() expects exactly 1 argument")
	}

	// Evaluate argument
	val, err := a.evaluateConstant(args[0])
	if err != nil {
		return nil, err
	}

	// Convert to float64 and apply floor
	switch v := val.(type) {
	case int:
		return v, nil // Already an integer
	case float64:
		return int(math.Floor(v)), nil
	default:
		return nil, fmt.Errorf("Floor() expects numeric argument")
	}
}

// evaluateConstantCeil evaluates Ceil() at compile time.
func (a *Analyzer) evaluateConstantCeil(args []ast.Expression) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Ceil() expects exactly 1 argument")
	}

	// Evaluate argument
	val, err := a.evaluateConstant(args[0])
	if err != nil {
		return nil, err
	}

	// Convert to float64 and apply ceil
	switch v := val.(type) {
	case int:
		return v, nil // Already an integer
	case float64:
		return int(math.Ceil(v)), nil
	default:
		return nil, fmt.Errorf("Ceil() expects numeric argument")
	}
}

// evaluateConstantRound evaluates Round() at compile time.
func (a *Analyzer) evaluateConstantRound(args []ast.Expression) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Round() expects exactly 1 argument")
	}

	// Evaluate argument
	val, err := a.evaluateConstant(args[0])
	if err != nil {
		return nil, err
	}

	// Convert to float64 and apply round
	switch v := val.(type) {
	case int:
		return v, nil // Already an integer
	case float64:
		return int(math.Round(v)), nil
	default:
		return nil, fmt.Errorf("Round() expects numeric argument")
	}
}

// evaluateConstantOrd evaluates Ord() at compile time.
// Ord() returns the ordinal value (Unicode code point) of a character.
func (a *Analyzer) evaluateConstantOrd(args []ast.Expression) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Ord() expects exactly 1 argument")
	}

	// Evaluate argument
	val, err := a.evaluateConstant(args[0])
	if err != nil {
		return nil, err
	}

	// Must be a string (character)
	strVal, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Ord() expects a character argument, got %T", val)
	}

	// Must be exactly one Unicode code point (rune)
	runes := []rune(strVal)
	if len(runes) != 1 {
		return nil, fmt.Errorf("Ord() expects a single character, got string with %d runes", len(runes))
	}

	// Return the Unicode code point
	return int(runes[0]), nil
}

// evaluateConstantChr evaluates Chr() at compile time.
// Chr() returns the character with the given ordinal value (Unicode code point).
func (a *Analyzer) evaluateConstantChr(args []ast.Expression) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Chr() expects exactly 1 argument")
	}

	// Evaluate argument
	val, err := a.evaluateConstant(args[0])
	if err != nil {
		return nil, err
	}

	// Convert to integer
	var intVal int
	switch v := val.(type) {
	case int:
		intVal = v
	case float64:
		intVal = int(v)
	default:
		return nil, fmt.Errorf("Chr() expects numeric argument, got %T", val)
	}

	// Validate code point range: full Unicode range
	if intVal < 0 || intVal > 0x10FFFF {
		return nil, fmt.Errorf("Chr() code point %d out of range (0-0x10FFFF)", intVal)
	}

	// Return single-character string
	return string(rune(intVal)), nil
}

// evaluateConstantTypeCast evaluates a type cast expression at compile time.
// Returns nil if the type cast cannot be evaluated at compile time.
func (a *Analyzer) evaluateConstantTypeCast(typeName string, arg ast.Expression) interface{} {
	// Try to resolve the type
	targetType, err := a.resolveType(typeName)
	if err != nil {
		return nil // Not a type name
	}

	// Unwrap TypeAlias to get the actual type
	if typeAlias, ok := targetType.(*types.TypeAlias); ok {
		targetType = typeAlias.AliasedType
	}

	// Evaluate the argument
	argVal, err := a.evaluateConstant(arg)
	if err != nil {
		return nil // Argument is not a compile-time constant
	}

	// Handle different type casts
	switch targetType.(type) {
	case *types.EnumType:
		// Cast to enum type
		switch v := argVal.(type) {
		case int:
			// Integer → Enum: Return the integer ordinal
			// The const will store the ordinal value, which will be converted to EnumValue at runtime
			return v
		default:
			return nil // Can't cast this type to enum at compile time
		}

	default:
		// For other types (Integer, Float, String), just return the value
		// The runtime will handle the actual casting
		typeLower := pkgident.Normalize(typeName)
		switch typeLower {
		case "integer":
			// Cast to Integer
			switch v := argVal.(type) {
			case int:
				return v
			case float64:
				return int(v)
			default:
				return nil
			}
		case "float":
			// Cast to Float
			switch v := argVal.(type) {
			case int:
				return float64(v)
			case float64:
				return v
			default:
				return nil
			}
		case "string":
			// Cast to String
			switch v := argVal.(type) {
			case int:
				return fmt.Sprintf("%d", v)
			case float64:
				return fmt.Sprintf("%g", v)
			case string:
				return v
			default:
				return nil
			}
		}
	}

	return nil // Type cast not supported at compile time
}

// analyzeTypeDeclaration analyzes a type declaration statement
// Handles type aliases: type TUserID = Integer;
// Handles subrange types: type TDigit = 0..9;
func (a *Analyzer) analyzeTypeDeclaration(decl *ast.TypeDeclaration) {
	if decl == nil {
		return
	}

	// Check if type name already exists
	if _, err := a.resolveType(decl.Name.Value); err == nil {
		a.addError("type '%s' already declared at %s", decl.Name.Value, decl.Token.Pos.String())
		return
	}

	// Handle function pointer types
	if decl.FunctionPointerType != nil {
		a.analyzeFunctionPointerTypeDeclaration(decl)
		return
	}

	// Handle subrange types
	if decl.IsSubrange {
		// Evaluate low bound (must be compile-time constant)
		lowBound, err := a.evaluateConstantInt(decl.LowBound)
		if err != nil {
			a.addError("subrange low bound must be a compile-time constant integer at %s: %v",
				decl.Token.Pos.String(), err)
			return
		}

		// Evaluate high bound (must be compile-time constant)
		highBound, err := a.evaluateConstantInt(decl.HighBound)
		if err != nil {
			a.addError("subrange high bound must be a compile-time constant integer at %s: %v",
				decl.Token.Pos.String(), err)
			return
		}

		if lowBound > highBound {
			a.addError("subrange low bound (%d) cannot be greater than high bound (%d) at %s",
				lowBound, highBound, decl.Token.Pos.String())
			return
		}

		subrangeType := &types.SubrangeType{
			BaseType:  types.INTEGER, // Subranges are currently based on Integer
			Name:      decl.Name.Value,
			LowBound:  lowBound,
			HighBound: highBound,
		}

		// Use lowercase key for case-insensitive lookup
		a.subranges[pkgident.Normalize(decl.Name.Value)] = subrangeType
		return
	}

	// Handle type aliases
	if decl.IsAlias {
		var aliasedType types.Type
		var err error

		// Resolve the aliased type expression
		aliasedType, err = a.resolveTypeExpression(decl.AliasedType)
		if err != nil {
			typeName := getTypeExpressionName(decl.AliasedType)
			a.addError("unknown type '%s' in type alias at %s", typeName, decl.Token.Pos.String())
			return
		}

		// Create TypeAlias and register it
		typeAlias := &types.TypeAlias{
			Name:        decl.Name.Value,
			AliasedType: aliasedType,
		}

		// Use lowercase key for case-insensitive lookup
		a.registerTypeWithPos(decl.Name.Value, typeAlias, decl.Token.Pos)
	}
}
