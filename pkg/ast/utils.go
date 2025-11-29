package ast

// ExtractIntegerLiteral extracts an integer literal value from an AST expression.
// It supports both plain integer literals and unary-negated integer literals.
//
// Returns the extracted integer value and true if successful, or 0 and false if
// the expression is not a supported pattern.
//
// Supported patterns:
//   - *IntegerLiteral: Returns the literal value directly
//   - *UnaryExpression with "-" operator and IntegerLiteral operand: Returns the negated value
//
// All other expression types return (0, false).
//
// This is commonly used for extracting compile-time constant integer values from
// property index directives and array bounds.
func ExtractIntegerLiteral(expr Expression) (int64, bool) {
	switch v := expr.(type) {
	case *IntegerLiteral:
		return v.Value, true
	case *UnaryExpression:
		if v.Operator == "-" {
			if lit, ok := v.Right.(*IntegerLiteral); ok {
				return -lit.Value, true
			}
		}
	}
	return 0, false
}
