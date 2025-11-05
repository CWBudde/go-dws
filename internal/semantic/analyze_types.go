package semantic

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Type Declaration Analysis
// ============================================================================

// evaluateConstant evaluates a compile-time constant expression.
// Returns the constant value and an error if the expression is not a constant.
// Task 9.205: Used for const declarations to store compile-time values.
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
		// Delegate to evaluateConstantInt for integer binary ops
		val, err := a.evaluateConstantInt(expr)
		if err != nil {
			return nil, err
		}
		return val, nil

	default:
		return nil, fmt.Errorf("expression is not a compile-time constant")
	}
}

// evaluateConstantInt evaluates a compile-time constant integer expression.
// Returns the integer value and an error if the expression is not a constant.
// Task 9.98: Used for subrange bound evaluation.
// Task 9.205: Extended to handle identifiers (const references) and binary expressions.
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

	default:
		return 0, fmt.Errorf("expression is not a compile-time constant integer")
	}
}

// analyzeTypeDeclaration analyzes a type declaration statement
// Handles type aliases: type TUserID = Integer;
// Handles subrange types: type TDigit = 0..9;
// Task 9.159: Handles function pointer types: type TFunc = function(x: Integer): Boolean;
func (a *Analyzer) analyzeTypeDeclaration(decl *ast.TypeDeclaration) {
	if decl == nil {
		return
	}

	// Check if type name already exists
	if _, err := a.resolveType(decl.Name.Value); err == nil {
		a.addError("type '%s' already declared at %s", decl.Name.Value, decl.Token.Pos.String())
		return
	}

	// Task 9.159: Handle function pointer types
	if decl.FunctionPointerType != nil {
		a.analyzeFunctionPointerTypeDeclaration(decl)
		return
	}

	// Task 9.98: Handle subrange types
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

		// Task 9.98: Validate low <= high
		if lowBound > highBound {
			a.addError("subrange low bound (%d) cannot be greater than high bound (%d) at %s",
				lowBound, highBound, decl.Token.Pos.String())
			return
		}

		// Task 9.98: Create SubrangeType and register in type environment
		subrangeType := &types.SubrangeType{
			BaseType:  types.INTEGER, // Subranges are currently based on Integer
			Name:      decl.Name.Value,
			LowBound:  lowBound,
			HighBound: highBound,
		}

		// Use lowercase key for case-insensitive lookup
		a.subranges[strings.ToLower(decl.Name.Value)] = subrangeType
		return
	}

	// Handle type aliases
	if decl.IsAlias {
		// Resolve the aliased type
		aliasedType, err := a.resolveType(decl.AliasedType.Name)
		if err != nil {
			a.addError("unknown type '%s' in type alias at %s", decl.AliasedType.Name, decl.Token.Pos.String())
			return
		}

		// Create TypeAlias and register it
		typeAlias := &types.TypeAlias{
			Name:        decl.Name.Value,
			AliasedType: aliasedType,
		}

		// Use lowercase key for case-insensitive lookup
		a.typeAliases[strings.ToLower(decl.Name.Value)] = typeAlias
	}
}
