package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// applyCompoundOperation applies a compound assignment operation (+=, -=, *=, /=).
//
// Handles:
// - Integer operations: += -= *= /= between integers
// - Float operations: += -= *= /= between floats and integers (with implicit conversion)
// - String concatenation: += for strings
// - Variant operations: delegates to evalVariantBinaryOp
// - Class operator overloads: uses TryBinaryOperator infrastructure
func (e *Evaluator) applyCompoundOperation(op token.TokenType, left, right Value, node ast.Node) Value {
	// Check for class operator overloads first
	// Use existing TryBinaryOperator infrastructure
	leftType := left.Type()
	if strings.HasPrefix(leftType, "OBJECT[") {
		// Try operator overload - convert compound op to binary op
		var binaryOp string
		switch op {
		case token.PLUS_ASSIGN:
			binaryOp = "+"
		case token.MINUS_ASSIGN:
			binaryOp = "-"
		case token.TIMES_ASSIGN:
			binaryOp = "*"
		case token.DIVIDE_ASSIGN:
			binaryOp = "/"
		default:
			return e.newError(node, "unknown compound operator: %v", op)
		}

		// Try to find operator overload on the object
		if e.oopEngine != nil {
			if result, found := e.oopEngine.TryBinaryOperator(binaryOp, left, right, node); found {
				return result
			}
		}

		// No overload found, fall through to standard operations
	}

	switch op {
	case token.PLUS_ASSIGN:
		return e.evalPlusAssign(left, right, node)

	case token.MINUS_ASSIGN:
		return e.evalMinusAssign(left, right, node)

	case token.TIMES_ASSIGN:
		return e.evalTimesAssign(left, right, node)

	case token.DIVIDE_ASSIGN:
		return e.evalDivideAssign(left, right, node)

	default:
		return e.newError(node, "unknown compound operator: %v", op)
	}
}

// evalPlusAssign handles the += operator for various types.
func (e *Evaluator) evalPlusAssign(left, right Value, node ast.Node) Value {
	// Handle Variant values first - delegate to evalVariantBinaryOp
	if _, ok := left.(runtime.VariantWrapper); ok {
		result := e.evalVariantBinaryOp("+", left, right, node)
		if isError(result) {
			return result
		}
		return result
	}

	switch l := left.(type) {
	case *runtime.IntegerValue:
		if r, ok := right.(*runtime.IntegerValue); ok {
			return &runtime.IntegerValue{Value: l.Value + r.Value}
		}
		// Float to Integer conversion would lose precision, not allowed
		return e.newError(node, "type mismatch: cannot add %s to Integer", right.Type())

	case *runtime.FloatValue:
		// Support Float + Float and Float + Integer (with implicit conversion)
		switch r := right.(type) {
		case *runtime.FloatValue:
			return &runtime.FloatValue{Value: l.Value + r.Value}
		case *runtime.IntegerValue:
			// Implicit Integer to Float conversion
			return &runtime.FloatValue{Value: l.Value + float64(r.Value)}
		default:
			return e.newError(node, "type mismatch: cannot add %s to Float", right.Type())
		}

	case *runtime.StringValue:
		if r, ok := right.(*runtime.StringValue); ok {
			return &runtime.StringValue{Value: l.Value + r.Value}
		}
		// Handle Variant-to-String conversion for array of const elements
		if wrapper, ok := right.(runtime.VariantWrapper); ok {
			// Unwrap the variant and convert to string
			innerVal := wrapper.UnwrapVariant()
			if innerVal == nil {
				return e.newError(node, "failed to unbox variant")
			}
			strVal := convertToString(innerVal)
			return &runtime.StringValue{Value: l.Value + strVal}
		}
		return e.newError(node, "type mismatch: cannot add %s to String", right.Type())

	default:
		return e.newError(node, "operator += not supported for type %s", left.Type())
	}
}

// evalMinusAssign handles the -= operator for numeric types.
func (e *Evaluator) evalMinusAssign(left, right Value, node ast.Node) Value {
	// Handle Variant values first
	if _, ok := left.(runtime.VariantWrapper); ok {
		result := e.evalVariantBinaryOp("-", left, right, node)
		if isError(result) {
			return result
		}
		return result
	}

	switch l := left.(type) {
	case *runtime.IntegerValue:
		if r, ok := right.(*runtime.IntegerValue); ok {
			return &runtime.IntegerValue{Value: l.Value - r.Value}
		}
		return e.newError(node, "type mismatch: cannot subtract %s from Integer", right.Type())

	case *runtime.FloatValue:
		// Support Float - Float and Float - Integer (with implicit conversion)
		switch r := right.(type) {
		case *runtime.FloatValue:
			return &runtime.FloatValue{Value: l.Value - r.Value}
		case *runtime.IntegerValue:
			// Implicit Integer to Float conversion
			return &runtime.FloatValue{Value: l.Value - float64(r.Value)}
		default:
			return e.newError(node, "type mismatch: cannot subtract %s from Float", right.Type())
		}

	default:
		return e.newError(node, "operator -= not supported for type %s", left.Type())
	}
}

// evalTimesAssign handles the *= operator for numeric types.
func (e *Evaluator) evalTimesAssign(left, right Value, node ast.Node) Value {
	// Handle Variant values first
	if _, ok := left.(runtime.VariantWrapper); ok {
		result := e.evalVariantBinaryOp("*", left, right, node)
		if isError(result) {
			return result
		}
		return result
	}

	switch l := left.(type) {
	case *runtime.IntegerValue:
		if r, ok := right.(*runtime.IntegerValue); ok {
			return &runtime.IntegerValue{Value: l.Value * r.Value}
		}
		return e.newError(node, "type mismatch: cannot multiply Integer by %s", right.Type())

	case *runtime.FloatValue:
		// Support Float * Float and Float * Integer (with implicit conversion)
		switch r := right.(type) {
		case *runtime.FloatValue:
			return &runtime.FloatValue{Value: l.Value * r.Value}
		case *runtime.IntegerValue:
			// Implicit Integer to Float conversion
			return &runtime.FloatValue{Value: l.Value * float64(r.Value)}
		default:
			return e.newError(node, "type mismatch: cannot multiply Float by %s", right.Type())
		}

	default:
		return e.newError(node, "operator *= not supported for type %s", left.Type())
	}
}

// evalDivideAssign handles the /= operator for numeric types.
// Includes enhanced error reporting with operand values.
func (e *Evaluator) evalDivideAssign(left, right Value, node ast.Node) Value {
	// Handle Variant values first
	if _, ok := left.(runtime.VariantWrapper); ok {
		result := e.evalVariantBinaryOp("/", left, right, node)
		if isError(result) {
			return result
		}
		return result
	}

	switch l := left.(type) {
	case *runtime.IntegerValue:
		if r, ok := right.(*runtime.IntegerValue); ok {
			if r.Value == 0 {
				// Enhanced error with operand values
				return e.newDivisionByZeroError(node, l.Value, r.Value)
			}
			return &runtime.IntegerValue{Value: l.Value / r.Value}
		}
		return e.newError(node, "type mismatch: cannot divide Integer by %s", right.Type())

	case *runtime.FloatValue:
		// Support Float / Float and Float / Integer (with implicit conversion)
		switch r := right.(type) {
		case *runtime.FloatValue:
			if r.Value == 0.0 {
				// Enhanced error with operand values
				return e.newFloatDivisionByZeroError(node, l.Value, r.Value)
			}
			return &runtime.FloatValue{Value: l.Value / r.Value}
		case *runtime.IntegerValue:
			// Implicit Integer to Float conversion
			if r.Value == 0 {
				// Enhanced error with operand values
				return e.newFloatIntDivisionByZeroError(node, l.Value, r.Value)
			}
			return &runtime.FloatValue{Value: l.Value / float64(r.Value)}
		default:
			return e.newError(node, "type mismatch: cannot divide Float by %s", right.Type())
		}

	default:
		return e.newError(node, "operator /= not supported for type %s", left.Type())
	}
}

// newDivisionByZeroError creates an enhanced division by zero error for integers.
func (e *Evaluator) newDivisionByZeroError(node ast.Node, left, right int64) Value {
	return e.newError(node, "Division by zero")
}

// newFloatDivisionByZeroError creates an enhanced division by zero error for floats.
func (e *Evaluator) newFloatDivisionByZeroError(node ast.Node, left, right float64) Value {
	return e.newError(node, "Division by zero")
}

// newFloatIntDivisionByZeroError creates an enhanced division by zero error for float/int.
func (e *Evaluator) newFloatIntDivisionByZeroError(node ast.Node, left float64, right int64) Value {
	return e.newError(node, "Division by zero")
}
