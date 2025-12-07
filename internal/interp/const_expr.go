package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// evaluateConstantExpression evaluates a constant expression at compile-time.
// This is used for enum value initialization with expressions like Ord('A'), 1+2, etc.
//
// Returns the integer value of the expression, or an error if:
// - The expression is not a valid constant expression
// - The expression cannot be evaluated at compile-time
// - The expression does not produce an integer result
//
// Supported expressions:
// - Integer literals: 42, -5
// - Unary expressions: -x
// - Binary expressions: 1+2, 5*3, 10 div 2
// - Function calls: Ord('A'), Ord('Z')
func (i *Interpreter) evaluateConstantExpression(expr ast.Expression) (int, error) {
	if expr == nil {
		return 0, fmt.Errorf("nil expression")
	}

	switch node := expr.(type) {
	case *ast.IntegerLiteral:
		// Simple integer literal
		return int(node.Value), nil

	case *ast.UnaryExpression:
		// Handle unary operators (-, +, not)
		operand, err := i.evaluateConstantExpression(node.Right)
		if err != nil {
			return 0, err
		}

		switch node.Operator {
		case "-":
			return -operand, nil
		case "+":
			return operand, nil
		case "not":
			// Bitwise NOT for integer
			return ^operand, nil
		default:
			return 0, fmt.Errorf("unsupported unary operator in constant expression: %s", node.Operator)
		}

	case *ast.BinaryExpression:
		// Handle binary operators (+, -, *, div, mod, etc.)
		left, err := i.evaluateConstantExpression(node.Left)
		if err != nil {
			return 0, fmt.Errorf("left operand: %w", err)
		}

		right, err := i.evaluateConstantExpression(node.Right)
		if err != nil {
			return 0, fmt.Errorf("right operand: %w", err)
		}

		switch node.Operator {
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
		case "and":
			// Bitwise AND
			return left & right, nil
		case "or":
			// Bitwise OR
			return left | right, nil
		case "xor":
			// Bitwise XOR
			return left ^ right, nil
		case "shl":
			// Bitwise shift left
			return left << uint(right), nil
		case "shr":
			// Bitwise shift right
			return left >> uint(right), nil
		default:
			return 0, fmt.Errorf("unsupported binary operator in constant expression: %s", node.Operator)
		}

	case *ast.CallExpression:
		// Handle function calls like Ord('A'), Chr(65)
		funcIdent, ok := node.Function.(*ast.Identifier)
		if !ok {
			return 0, fmt.Errorf("only simple function calls are supported in constant expressions")
		}

		funcName := funcIdent.Value

		switch funcName {
		case "Ord":
			// Ord('A') -> 65
			if len(node.Arguments) != 1 {
				return 0, fmt.Errorf("Ord() expects exactly 1 argument, got %d", len(node.Arguments))
			}

			arg := node.Arguments[0]

			// Check if it's a character literal (string with length 1)
			if strLit, ok := arg.(*ast.StringLiteral); ok {
				if len(strLit.Value) != 1 {
					return 0, fmt.Errorf("Ord() argument must be a single character, got string of length %d", len(strLit.Value))
				}
				return int(strLit.Value[0]), nil
			}

			// Otherwise, try to evaluate the argument as an integer expression
			return i.evaluateConstantExpression(arg)

		case "Chr":
			// Chr(65) -> 'A' (but we return the integer value)
			if len(node.Arguments) != 1 {
				return 0, fmt.Errorf("Chr() expects exactly 1 argument, got %d", len(node.Arguments))
			}

			return i.evaluateConstantExpression(node.Arguments[0])

		default:
			return 0, fmt.Errorf("function '%s' is not supported in constant expressions (only Ord and Chr are allowed)", funcName)
		}

	case *ast.StringLiteral:
		// Single character string can be treated as its ordinal value
		if len(node.Value) == 1 {
			return int(node.Value[0]), nil
		}
		return 0, fmt.Errorf("string literals in constant expressions must be single characters (use Ord('x'))")

	default:
		return 0, fmt.Errorf("unsupported expression type in constant expression: %T", expr)
	}
}
