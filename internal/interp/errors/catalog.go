package errors

import (
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// Error Message Catalog
//
// This file provides standardized error messages and helper functions for
// creating consistent, well-formatted errors throughout the interpreter.
//
// Error Message Format:
//   - Type errors: "operation failed: reason" (e.g., "binary operation failed: type mismatch")
//   - Runtime errors: "operation failed: reason" (e.g., "division failed: division by zero")
//   - Undefined errors: "entity type not found: name" (e.g., "variable not found: x")
//   - Contract errors: "contract type failed: details" (e.g., "precondition failed: x > 0")
//   - Internal errors: "internal error: details" (e.g., "internal error: unknown node type")
//
// All error messages:
//   - Start with lowercase (except for proper nouns)
//   - Use present tense
//   - Include relevant context (types, values, names)
//   - Are concise but informative

// ============================================================================
// Type Error Messages
// ============================================================================

const (
	// Binary operation errors
	ErrMsgTypeMismatch     = "type mismatch: %s %s %s"
	ErrMsgUnknownOperator  = "unknown operator: %s %s %s"
	ErrMsgInvalidOperation = "invalid operation: %s on type %s"

	// Type conversion errors
	ErrMsgCannotConvert      = "cannot convert %s to %s"
	ErrMsgCannotConvertValue = "cannot convert %s '%v' to %s"
	ErrMsgCannotCast         = "cannot cast %s to %s"
	ErrMsgCannotCastValue    = "cannot cast %s '%v' to %s: %s"

	// Type expectation errors
	ErrMsgExpectedType      = "expected %s, got %s"
	ErrMsgExpectedTypes     = "expected %s or %s, got %s"
	ErrMsgIncompatibleTypes = "incompatible types: %s and %s"
)

// ============================================================================
// Runtime Error Messages
// ============================================================================

const (
	// Arithmetic errors
	ErrMsgDivisionByZero = "division by zero"
	ErrMsgDivByZero      = "division by zero: %v / %v"
	ErrMsgIntegerDivByZero = "integer division by zero: %v div %v"
	ErrMsgModByZero        = "modulo by zero: %v mod %v"
	ErrMsgOverflow         = "arithmetic overflow"
	ErrMsgUnderflow        = "arithmetic underflow"

	// Index errors
	ErrMsgIndexOutOfBounds       = "index out of bounds: %d"
	ErrMsgIndexOutOfBoundsArray  = "index out of bounds: %d (array length is %d)"
	ErrMsgIndexOutOfBoundsRange  = "index out of bounds: %d (bounds are %d..%d)"
	ErrMsgIndexOutOfBoundsString = "string index out of bounds: %d (string length is %d)"

	// Nil/null errors
	ErrMsgNilDereference = "nil dereference"
	ErrMsgNilObject      = "object is nil"
	ErrMsgNilInterface   = "interface is nil"

	// Function call errors
	ErrMsgWrongArgCount      = "wrong number of arguments: expected %d, got %d"
	ErrMsgWrongArgCountMin   = "wrong number of arguments: expected at least %d, got %d"
	ErrMsgWrongArgCountMax   = "wrong number of arguments: expected at most %d, got %d"
	ErrMsgWrongArgCountRange = "wrong number of arguments: expected %d-%d, got %d"
	ErrMsgWrongArgCountFor   = "wrong number of arguments for %s: expected %d, got %d"
)

// ============================================================================
// Undefined Error Messages
// ============================================================================

const (
	// Variable errors
	ErrMsgUndefinedVariable = "undefined variable: %s"
	ErrMsgVariableNotFound  = "variable not found: %s"

	// Function errors
	ErrMsgUndefinedFunction  = "undefined function: %s"
	ErrMsgUndefinedProcedure = "undefined procedure: %s"
	ErrMsgFunctionNotFound   = "function or procedure not found: %s"

	// Type errors
	ErrMsgUndefinedType  = "undefined type: %s"
	ErrMsgTypeNotFound   = "type not found: %s"
	ErrMsgUnknownType    = "unknown type: %s"

	// Member errors
	ErrMsgUndefinedMember = "undefined member: %s"
	ErrMsgMemberNotFound  = "member not found: %s in %s"
	ErrMsgUndefinedMethod = "undefined method: %s"
	ErrMsgMethodNotFound  = "method not found: %s in class %s"

	// Enum errors
	ErrMsgUndefinedEnumValue = "undefined enum value: %s"
	ErrMsgEnumValueNotFound  = "enum value not found: %s in enum %s"
)

// ============================================================================
// Contract Error Messages
// ============================================================================

const (
	// Precondition errors
	ErrMsgPreconditionFailed    = "precondition failed: %s"
	ErrMsgPreconditionNonBool   = "precondition must evaluate to boolean, got %s"
	ErrMsgPreconditionEvalError = "precondition evaluation failed: %s"

	// Postcondition errors
	ErrMsgPostconditionFailed    = "postcondition failed: %s"
	ErrMsgPostconditionNonBool   = "postcondition must evaluate to boolean, got %s"
	ErrMsgPostconditionEvalError = "postcondition evaluation failed: %s"

	// Invariant errors
	ErrMsgInvariantFailed    = "invariant failed: %s"
	ErrMsgInvariantNonBool   = "invariant must evaluate to boolean, got %s"
	ErrMsgInvariantEvalError = "invariant evaluation failed: %s"

	// Assert errors
	ErrMsgAssertionFailed = "assertion failed: %s"
)

// ============================================================================
// Internal Error Messages
// ============================================================================

const (
	// Interpreter errors
	ErrMsgInternalError  = "internal error: %s"
	ErrMsgUnknownNode    = "internal error: unknown node type: %T"
	ErrMsgInvalidState   = "internal error: invalid interpreter state: %s"
	ErrMsgMissingValue   = "internal error: missing value for: %s"

	// Unimplemented features
	ErrMsgNotImplemented       = "not implemented: %s"
	ErrMsgFeatureNotSupported  = "feature not supported: %s"
)

// ============================================================================
// Helper Functions for Creating Standardized Errors
// ============================================================================

// TypeMismatchError creates a type mismatch error for binary operations.
// Example: "type mismatch: INTEGER + STRING"
func TypeMismatchError(pos *token.Position, expr string, leftType, op, rightType string) *InterpreterError {
	return NewTypeErrorf(pos, expr, ErrMsgTypeMismatch, leftType, op, rightType)
}

// UnknownOperatorError creates an unknown operator error.
// Example: "unknown operator: INTEGER ++ STRING"
func UnknownOperatorError(pos *token.Position, expr string, leftType, op, rightType string) *InterpreterError {
	return NewTypeErrorf(pos, expr, ErrMsgUnknownOperator, leftType, op, rightType)
}

// CannotConvertError creates a type conversion error.
// Example: "cannot convert STRING to INTEGER"
func CannotConvertError(pos *token.Position, expr string, fromType, toType string) *InterpreterError {
	return NewTypeErrorf(pos, expr, ErrMsgCannotConvert, fromType, toType)
}

// CannotConvertValueError creates a type conversion error with the value.
// Example: "cannot convert STRING 'abc' to INTEGER"
func CannotConvertValueError(pos *token.Position, expr string, fromType string, value interface{}, toType string) *InterpreterError {
	return NewTypeErrorf(pos, expr, ErrMsgCannotConvertValue, fromType, value, toType)
}

// CannotCastError creates a type cast error.
// Example: "cannot cast VARIANT to MyClass"
func CannotCastError(pos *token.Position, expr string, fromType, toType string) *InterpreterError {
	return NewTypeErrorf(pos, expr, ErrMsgCannotCast, fromType, toType)
}

// ExpectedTypeError creates an expected type error.
// Example: "expected INTEGER, got STRING"
func ExpectedTypeError(pos *token.Position, expr string, expectedType, actualType string) *InterpreterError {
	return NewTypeErrorf(pos, expr, ErrMsgExpectedType, expectedType, actualType)
}

// DivisionByZeroError creates a division by zero error.
// Example: "division by zero: 10 / 0"
func DivisionByZeroError(pos *token.Position, expr string, left, right interface{}) *InterpreterError {
	return NewRuntimeErrorf(pos, expr, ErrMsgDivByZero, left, right)
}

// IntegerDivByZeroError creates an integer division by zero error.
// Example: "integer division by zero: 10 div 0"
func IntegerDivByZeroError(pos *token.Position, expr string, left, right interface{}) *InterpreterError {
	return NewRuntimeErrorf(pos, expr, ErrMsgIntegerDivByZero, left, right)
}

// IndexOutOfBoundsError creates an index out of bounds error.
// Example: "index out of bounds: 10 (array length is 5)"
func IndexOutOfBoundsError(pos *token.Position, expr string, index, length int) *InterpreterError {
	return NewRuntimeErrorf(pos, expr, ErrMsgIndexOutOfBoundsArray, index, length)
}

// IndexOutOfBoundsRangeError creates an index out of bounds error with range.
// Example: "index out of bounds: 10 (bounds are 1..5)"
func IndexOutOfBoundsRangeError(pos *token.Position, expr string, index, low, high int) *InterpreterError {
	return NewRuntimeErrorf(pos, expr, ErrMsgIndexOutOfBoundsRange, index, low, high)
}

// StringIndexOutOfBoundsError creates a string index out of bounds error.
// Example: "string index out of bounds: 10 (string length is 5)"
func StringIndexOutOfBoundsError(pos *token.Position, expr string, index, length int) *InterpreterError {
	return NewRuntimeErrorf(pos, expr, ErrMsgIndexOutOfBoundsString, index, length)
}

// WrongArgumentCountError creates a wrong argument count error.
// Example: "wrong number of arguments: expected 2, got 3"
func WrongArgumentCountError(pos *token.Position, expr string, expected, actual int) *InterpreterError {
	return NewRuntimeErrorf(pos, expr, ErrMsgWrongArgCount, expected, actual)
}

// WrongArgumentCountForError creates a wrong argument count error for a named function.
// Example: "wrong number of arguments for Sqrt: expected 1, got 2"
func WrongArgumentCountForError(pos *token.Position, expr string, name string, expected, actual int) *InterpreterError {
	return NewRuntimeErrorf(pos, expr, ErrMsgWrongArgCountFor, name, expected, actual)
}

// UndefinedVariableError creates an undefined variable error.
// Example: "undefined variable: x"
func UndefinedVariableError(pos *token.Position, expr string, name string) *InterpreterError {
	return NewUndefinedErrorf(pos, expr, ErrMsgUndefinedVariable, name)
}

// UndefinedFunctionError creates an undefined function error.
// Example: "undefined function: DoSomething"
func UndefinedFunctionError(pos *token.Position, expr string, name string) *InterpreterError {
	return NewUndefinedErrorf(pos, expr, ErrMsgUndefinedFunction, name)
}

// FunctionNotFoundError creates a function/procedure not found error.
// Example: "function or procedure not found: DoSomething"
func FunctionNotFoundError(pos *token.Position, expr string, name string) *InterpreterError {
	return NewUndefinedErrorf(pos, expr, ErrMsgFunctionNotFound, name)
}

// UndefinedTypeError creates an undefined type error.
// Example: "undefined type: MyClass"
func UndefinedTypeError(pos *token.Position, expr string, name string) *InterpreterError {
	return NewUndefinedErrorf(pos, expr, ErrMsgUndefinedType, name)
}

// MethodNotFoundError creates a method not found error.
// Example: "method not found: ToString in class MyClass"
func MethodNotFoundError(pos *token.Position, expr string, methodName, className string) *InterpreterError {
	return NewUndefinedErrorf(pos, expr, ErrMsgMethodNotFound, methodName, className)
}

// PreconditionFailedError creates a precondition failure error.
// Example: "precondition failed: x > 0"
func PreconditionFailedError(pos *token.Position, expr string, condition string) *InterpreterError {
	return NewContractErrorf(pos, expr, ErrMsgPreconditionFailed, condition)
}

// PreconditionNonBoolError creates a precondition non-boolean error.
// Example: "precondition must evaluate to boolean, got INTEGER"
func PreconditionNonBoolError(pos *token.Position, expr string, actualType string) *InterpreterError {
	return NewContractErrorf(pos, expr, ErrMsgPreconditionNonBool, actualType)
}

// PostconditionFailedError creates a postcondition failure error.
// Example: "postcondition failed: result > 0"
func PostconditionFailedError(pos *token.Position, expr string, condition string) *InterpreterError {
	return NewContractErrorf(pos, expr, ErrMsgPostconditionFailed, condition)
}

// PostconditionNonBoolError creates a postcondition non-boolean error.
// Example: "postcondition must evaluate to boolean, got INTEGER"
func PostconditionNonBoolError(pos *token.Position, expr string, actualType string) *InterpreterError {
	return NewContractErrorf(pos, expr, ErrMsgPostconditionNonBool, actualType)
}

// AssertionFailedError creates an assertion failure error.
// Example: "assertion failed: x > 0"
func AssertionFailedError(pos *token.Position, expr string, condition string) *InterpreterError {
	return NewRuntimeErrorf(pos, expr, ErrMsgAssertionFailed, condition)
}

// UnknownNodeError creates an unknown node type error.
// Example: "internal error: unknown node type: *ast.SomeNode"
func UnknownNodeError(pos *token.Position, expr string, node ast.Node) *InterpreterError {
	return NewInternalErrorf(pos, expr, ErrMsgUnknownNode, node)
}

// NotImplementedError creates a not implemented error.
// Example: "not implemented: feature XYZ"
func NotImplementedError(pos *token.Position, expr string, feature string) *InterpreterError {
	return NewInternalErrorf(pos, expr, ErrMsgNotImplemented, feature)
}

// ============================================================================
// Convenience Functions for Common Error Patterns
// ============================================================================

// ErrTypeMismatch creates a type mismatch error from types.
func ErrTypeMismatch(node ast.Node, leftType, op, rightType string) *InterpreterError {
	pos := PositionFromNode(node)
	expr := ExpressionFromNode(node)
	return TypeMismatchError(pos, expr, leftType, op, rightType)
}

// ErrUnknownOperator creates an unknown operator error from types.
func ErrUnknownOperator(node ast.Node, leftType, op, rightType string) *InterpreterError {
	pos := PositionFromNode(node)
	expr := ExpressionFromNode(node)
	return UnknownOperatorError(pos, expr, leftType, op, rightType)
}

// ErrCannotConvert creates a type conversion error from types.
func ErrCannotConvert(node ast.Node, fromType, toType string) *InterpreterError {
	pos := PositionFromNode(node)
	expr := ExpressionFromNode(node)
	return CannotConvertError(pos, expr, fromType, toType)
}

// ErrCannotConvertValue creates a type conversion error with value from types.
func ErrCannotConvertValue(node ast.Node, fromType string, value interface{}, toType string) *InterpreterError {
	pos := PositionFromNode(node)
	expr := ExpressionFromNode(node)
	return CannotConvertValueError(pos, expr, fromType, value, toType)
}

// ErrDivByZero creates a division by zero error.
func ErrDivByZero(node ast.Node, left, right interface{}) *InterpreterError {
	pos := PositionFromNode(node)
	expr := ExpressionFromNode(node)
	return DivisionByZeroError(pos, expr, left, right)
}

// ErrIndexOutOfBounds creates an index out of bounds error.
func ErrIndexOutOfBounds(node ast.Node, index, length int) *InterpreterError {
	pos := PositionFromNode(node)
	expr := ExpressionFromNode(node)
	return IndexOutOfBoundsError(pos, expr, index, length)
}

// ErrUndefinedVariable creates an undefined variable error.
func ErrUndefinedVariable(node ast.Node, name string) *InterpreterError {
	pos := PositionFromNode(node)
	expr := ExpressionFromNode(node)
	return UndefinedVariableError(pos, expr, name)
}

// ErrUndefinedFunction creates an undefined function error.
func ErrUndefinedFunction(node ast.Node, name string) *InterpreterError {
	pos := PositionFromNode(node)
	expr := ExpressionFromNode(node)
	return UndefinedFunctionError(pos, expr, name)
}

// ErrWrongArgCount creates a wrong argument count error.
func ErrWrongArgCount(node ast.Node, expected, actual int) *InterpreterError {
	pos := PositionFromNode(node)
	expr := ExpressionFromNode(node)
	return WrongArgumentCountError(pos, expr, expected, actual)
}

// ErrNotImplemented creates a not implemented error.
func ErrNotImplemented(node ast.Node, feature string) *InterpreterError {
	pos := PositionFromNode(node)
	expr := ExpressionFromNode(node)
	return NotImplementedError(pos, expr, feature)
}
