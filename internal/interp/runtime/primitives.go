package runtime

import (
	"fmt"
	"math"
	"strconv"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Primitive Value Types
// ============================================================================

// IntegerValue represents an integer value in DWScript.
type IntegerValue struct {
	Value int64
}

// Type returns "INTEGER".
func (i *IntegerValue) Type() string {
	return "INTEGER"
}

// String returns the string representation of the integer.
func (i *IntegerValue) String() string {
	return strconv.FormatInt(i.Value, 10)
}

// AsInteger returns the integer value directly.
func (i *IntegerValue) AsInteger() (int64, bool) {
	return i.Value, true
}

// AsFloat converts the integer to a float.
func (i *IntegerValue) AsFloat() (float64, bool) {
	return float64(i.Value), true
}

// Equals checks if this integer equals another value.
func (i *IntegerValue) Equals(other Value) (bool, error) {
	switch v := other.(type) {
	case *IntegerValue:
		return i.Value == v.Value, nil
	case *FloatValue:
		return float64(i.Value) == v.Value, nil
	case NumericValue:
		// Try numeric comparison
		if otherInt, ok := v.AsInteger(); ok {
			return i.Value == otherInt, nil
		}
		if otherFloat, ok := v.AsFloat(); ok {
			return float64(i.Value) == otherFloat, nil
		}
	}
	return false, fmt.Errorf("cannot compare INTEGER with %s", other.Type())
}

// CompareTo compares this integer with another value.
func (i *IntegerValue) CompareTo(other Value) (int, error) {
	switch v := other.(type) {
	case *IntegerValue:
		if i.Value < v.Value {
			return -1, nil
		} else if i.Value > v.Value {
			return 1, nil
		}
		return 0, nil
	case *FloatValue:
		f := float64(i.Value)
		if f < v.Value {
			return -1, nil
		} else if f > v.Value {
			return 1, nil
		}
		return 0, nil
	case NumericValue:
		// Try numeric comparison
		if otherInt, ok := v.AsInteger(); ok {
			if i.Value < otherInt {
				return -1, nil
			} else if i.Value > otherInt {
				return 1, nil
			}
			return 0, nil
		}
		if otherFloat, ok := v.AsFloat(); ok {
			f := float64(i.Value)
			if f < otherFloat {
				return -1, nil
			} else if f > otherFloat {
				return 1, nil
			}
			return 0, nil
		}
	}
	return 0, fmt.Errorf("cannot compare INTEGER with %s", other.Type())
}

// Copy returns a copy of the integer (primitives are copied by value).
func (i *IntegerValue) Copy() Value {
	return &IntegerValue{Value: i.Value}
}

// ConvertTo converts the integer to the target type.
func (i *IntegerValue) ConvertTo(targetType string) (Value, error) {
	switch targetType {
	case "INTEGER":
		return i, nil
	case "FLOAT":
		return &FloatValue{Value: float64(i.Value)}, nil
	case "STRING":
		return &StringValue{Value: i.String()}, nil
	case "BOOLEAN":
		return &BooleanValue{Value: i.Value != 0}, nil
	default:
		return nil, fmt.Errorf("cannot convert INTEGER to %s", targetType)
	}
}

// ============================================================================

// FloatValue represents a floating-point value in DWScript.
type FloatValue struct {
	Value float64
}

// Type returns "FLOAT".
func (f *FloatValue) Type() string {
	return "FLOAT"
}

// String returns the string representation of the float.
func (f *FloatValue) String() string {
	if math.IsInf(f.Value, 1) {
		return "INF"
	}
	if math.IsInf(f.Value, -1) {
		return "-INF"
	}
	if math.IsNaN(f.Value) {
		return "NaN"
	}
	return strconv.FormatFloat(f.Value, 'g', -1, 64)
}

// AsInteger converts the float to an integer (truncates).
func (f *FloatValue) AsInteger() (int64, bool) {
	return int64(f.Value), true
}

// AsFloat returns the float value directly.
func (f *FloatValue) AsFloat() (float64, bool) {
	return f.Value, true
}

// Equals checks if this float equals another value.
func (f *FloatValue) Equals(other Value) (bool, error) {
	switch v := other.(type) {
	case *FloatValue:
		return f.Value == v.Value, nil
	case *IntegerValue:
		return f.Value == float64(v.Value), nil
	case NumericValue:
		if otherFloat, ok := v.AsFloat(); ok {
			return f.Value == otherFloat, nil
		}
	}
	return false, fmt.Errorf("cannot compare FLOAT with %s", other.Type())
}

// CompareTo compares this float with another value.
func (f *FloatValue) CompareTo(other Value) (int, error) {
	switch v := other.(type) {
	case *FloatValue:
		if f.Value < v.Value {
			return -1, nil
		} else if f.Value > v.Value {
			return 1, nil
		}
		return 0, nil
	case *IntegerValue:
		otherFloat := float64(v.Value)
		if f.Value < otherFloat {
			return -1, nil
		} else if f.Value > otherFloat {
			return 1, nil
		}
		return 0, nil
	case NumericValue:
		if otherFloat, ok := v.AsFloat(); ok {
			if f.Value < otherFloat {
				return -1, nil
			} else if f.Value > otherFloat {
				return 1, nil
			}
			return 0, nil
		}
	}
	return 0, fmt.Errorf("cannot compare FLOAT with %s", other.Type())
}

// Copy returns a copy of the float (primitives are copied by value).
func (f *FloatValue) Copy() Value {
	return &FloatValue{Value: f.Value}
}

// ConvertTo converts the float to the target type.
func (f *FloatValue) ConvertTo(targetType string) (Value, error) {
	switch targetType {
	case "FLOAT":
		return f, nil
	case "INTEGER":
		return &IntegerValue{Value: int64(f.Value)}, nil
	case "STRING":
		return &StringValue{Value: f.String()}, nil
	default:
		return nil, fmt.Errorf("cannot convert FLOAT to %s", targetType)
	}
}

// ============================================================================

// StringValue represents a string value in DWScript.
type StringValue struct {
	Value string
}

// Type returns "STRING".
func (s *StringValue) Type() string {
	return "STRING"
}

// String returns the string value itself.
func (s *StringValue) String() string {
	return s.Value
}

// Equals checks if this string equals another value.
func (s *StringValue) Equals(other Value) (bool, error) {
	if v, ok := other.(*StringValue); ok {
		return s.Value == v.Value, nil
	}
	return false, fmt.Errorf("cannot compare STRING with %s", other.Type())
}

// CompareTo compares this string with another value lexicographically.
func (s *StringValue) CompareTo(other Value) (int, error) {
	if v, ok := other.(*StringValue); ok {
		if s.Value < v.Value {
			return -1, nil
		} else if s.Value > v.Value {
			return 1, nil
		}
		return 0, nil
	}
	return 0, fmt.Errorf("cannot compare STRING with %s", other.Type())
}

// Copy returns a copy of the string (primitives are copied by value).
func (s *StringValue) Copy() Value {
	return &StringValue{Value: s.Value}
}

// GetIndex retrieves a character at the specified index (1-based, DWScript convention).
func (s *StringValue) GetIndex(index int64) (Value, error) {
	// DWScript uses 1-based indexing
	if index < 1 || index > int64(len(s.Value)) {
		return nil, fmt.Errorf("string index %d out of range [1..%d]", index, len(s.Value))
	}
	return &StringValue{Value: string(s.Value[index-1])}, nil
}

// SetIndex is not supported for strings (they are immutable).
func (s *StringValue) SetIndex(index int64, value Value) error {
	return fmt.Errorf("cannot modify string: strings are immutable")
}

// Length returns the length of the string.
func (s *StringValue) Length() int64 {
	return int64(len(s.Value))
}

// ConvertTo converts the string to the target type.
func (s *StringValue) ConvertTo(targetType string) (Value, error) {
	switch targetType {
	case "STRING":
		return s, nil
	case "INTEGER":
		if val, err := strconv.ParseInt(s.Value, 10, 64); err == nil {
			return &IntegerValue{Value: val}, nil
		}
		return nil, fmt.Errorf("cannot convert '%s' to INTEGER", s.Value)
	case "FLOAT":
		if val, err := strconv.ParseFloat(s.Value, 64); err == nil {
			return &FloatValue{Value: val}, nil
		}
		return nil, fmt.Errorf("cannot convert '%s' to FLOAT", s.Value)
	default:
		return nil, fmt.Errorf("cannot convert STRING to %s", targetType)
	}
}

// ============================================================================

// BooleanValue represents a boolean value in DWScript.
type BooleanValue struct {
	Value bool
}

// Type returns "BOOLEAN".
func (b *BooleanValue) Type() string {
	return "BOOLEAN"
}

// String returns "True" or "False".
func (b *BooleanValue) String() string {
	if b.Value {
		return "True"
	}
	return "False"
}

// Equals checks if this boolean equals another value.
func (b *BooleanValue) Equals(other Value) (bool, error) {
	if v, ok := other.(*BooleanValue); ok {
		return b.Value == v.Value, nil
	}
	return false, fmt.Errorf("cannot compare BOOLEAN with %s", other.Type())
}

// Copy returns a copy of the boolean (primitives are copied by value).
func (b *BooleanValue) Copy() Value {
	return &BooleanValue{Value: b.Value}
}

// ConvertTo converts the boolean to the target type.
func (b *BooleanValue) ConvertTo(targetType string) (Value, error) {
	switch targetType {
	case "BOOLEAN":
		return b, nil
	case "STRING":
		return &StringValue{Value: b.String()}, nil
	case "INTEGER":
		if b.Value {
			return &IntegerValue{Value: 1}, nil
		}
		return &IntegerValue{Value: 0}, nil
	default:
		return nil, fmt.Errorf("cannot convert BOOLEAN to %s", targetType)
	}
}

// ============================================================================

// NilValue represents a nil/null value in DWScript.
type NilValue struct {
	// ClassType stores the expected class type for this nil value (if any).
	// This allows accessing class variables via nil instances: var b: TBase; b.ClassVar
	ClassType string // e.g., "TBase"
}

// Type returns "NIL".
func (n *NilValue) Type() string {
	return "NIL"
}

// String returns "nil".
func (n *NilValue) String() string {
	return "nil"
}

// IsNil always returns true for NilValue.
func (n *NilValue) IsNil() bool {
	return true
}

// Equals checks if another value is also nil.
func (n *NilValue) Equals(other Value) (bool, error) {
	_, isNil := other.(*NilValue)
	return isNil, nil
}

// Copy returns a copy of the nil value.
func (n *NilValue) Copy() Value {
	return &NilValue{ClassType: n.ClassType}
}

// GetTypedClassName returns the class type name for typed nil values.
// Returns "" for untyped nil values.
func (n *NilValue) GetTypedClassName() string {
	if n == nil {
		return ""
	}
	return n.ClassType
}

// ============================================================================

// NullValue represents an uninitialized variant.
// This is different from NilValue (which is an explicit nil object reference).
type NullValue struct{}

// Type returns "NULL".
func (n *NullValue) Type() string {
	return "NULL"
}

// String returns "null".
func (n *NullValue) String() string {
	return "null"
}

// Equals checks if another value is also null.
func (n *NullValue) Equals(other Value) (bool, error) {
	_, isNull := other.(*NullValue)
	return isNull, nil
}

// Copy returns a copy of the null value.
func (n *NullValue) Copy() Value {
	return &NullValue{}
}

// ============================================================================

// UnassignedValue represents a variable that has been declared but not initialized.
type UnassignedValue struct{}

// Type returns "UNASSIGNED".
func (u *UnassignedValue) Type() string {
	return "UNASSIGNED"
}

// String returns "unassigned".
func (u *UnassignedValue) String() string {
	return "unassigned"
}

// Equals checks if another value is also unassigned.
func (u *UnassignedValue) Equals(other Value) (bool, error) {
	_, isUnassigned := other.(*UnassignedValue)
	return isUnassigned, nil
}

// Copy returns a copy of the unassigned value.
func (u *UnassignedValue) Copy() Value {
	return &UnassignedValue{}
}

// ============================================================================
// FunctionPointerValue - Runtime representation for function/method pointers
// ============================================================================
// Task 3.7.7: Moved from internal/interp/value.go to consolidate runtime types.

// FunctionPointerValue represents a function or procedure pointer in DWScript.
// Task 9.164: Create runtime representation for function pointers.
// Task 9.221: Extended to support lambda expressions/anonymous methods.
//
// Function pointers store a reference to a callable function/procedure along with
// its closure environment. Method pointers additionally capture the Self object.
// Lambdas are also represented using this type.
//
// Examples:
//   - Function pointer: var f: TFunc; f := @MyFunction;
//   - Method pointer: var m: TMethod; m := @obj.MyMethod; (captures obj as Self)
//   - Lambda: var f := lambda(x: Integer): Integer begin Result := x * 2; end;
//
// NOTE: Closure field uses interface{} to avoid circular import with interp.Environment.
// At runtime, this will be *Environment.
type FunctionPointerValue struct {
	Closure     interface{}                    // Environment where function was defined
	SelfObject  Value                          // Object instance for method pointers
	PointerType *types.FunctionPointerType     // Function pointer type info
	Function    *ast.FunctionDecl              // AST node of function (legacy)
	Lambda      *ast.LambdaExpression          // AST node of lambda (legacy)
	BuiltinName string                         // Built-in function identifier
	MethodID    MethodID                       // Unique ID in MethodRegistry
}

// Type returns "FUNCTION_POINTER", "METHOD_POINTER", or "LAMBDA" (closure).
// Task 9.221: Updated to distinguish lambdas.
func (f *FunctionPointerValue) Type() string {
	if f.SelfObject != nil {
		return "METHOD_POINTER"
	}
	if f.Lambda != nil {
		return "LAMBDA"
	}
	return "FUNCTION_POINTER"
}

// ===== FunctionPointerCallable interface implementation =====

// IsNil returns true if this function pointer has no function or lambda assigned.
// Used to check before invocation to raise appropriate DWScript exceptions.
func (f *FunctionPointerValue) IsNil() bool {
	return f.Function == nil && f.Lambda == nil && f.MethodID == InvalidMethodID
}

// ParamCount returns the number of parameters this function pointer expects.
// For lambdas, returns the lambda parameter count.
// For regular functions, returns the function parameter count.
// Returns 0 if neither is set.
func (f *FunctionPointerValue) ParamCount() int {
	if f.Lambda != nil {
		return len(f.Lambda.Parameters)
	}
	if f.Function != nil {
		return len(f.Function.Parameters)
	}
	return 0
}

// IsLambda returns true if this is a lambda/closure, false for regular function pointers.
func (f *FunctionPointerValue) IsLambda() bool {
	return f.Lambda != nil
}

// HasSelfObject returns true if this is a method pointer with a bound Self object.
func (f *FunctionPointerValue) HasSelfObject() bool {
	return f.SelfObject != nil
}

// GetFunctionDecl returns the function AST node (*ast.FunctionDecl) for regular function pointers.
// Returns nil for lambda closures.
func (f *FunctionPointerValue) GetFunctionDecl() any {
	return f.Function
}

// GetLambdaExpr returns the lambda AST node (*ast.LambdaExpression) for lambda closures.
// Returns nil for regular function pointers.
func (f *FunctionPointerValue) GetLambdaExpr() any {
	return f.Lambda
}

// GetClosure returns the captured environment (type: *Environment).
func (f *FunctionPointerValue) GetClosure() any {
	return f.Closure
}

// GetSelfObject returns the bound Self for method pointers.
// Returns nil for non-method pointers.
func (f *FunctionPointerValue) GetSelfObject() Value {
	return f.SelfObject
}

// String returns the string representation of the function pointer.
// Format: @FunctionName, @Object.MethodName, or <lambda> for closures
// Task 9.221: Updated to handle lambdas.
func (f *FunctionPointerValue) String() string {
	// Lambda closures (check legacy field first for backward compatibility)
	if f.Lambda != nil {
		return "<lambda>"
	}

	// Check if we have MethodID (AST-free path)
	if f.MethodID != InvalidMethodID {
		// For method pointers, show object + method
		if f.SelfObject != nil {
			return "@" + f.SelfObject.String() + ".<method>"
		}
		// For regular function pointers, show generic representation
		// (We don't have access to MethodRegistry here to look up name)
		return "@<function>"
	}

	if f.BuiltinName != "" {
		return "@" + f.BuiltinName
	}

	// Legacy path: use AST nodes
	// Regular function/method pointers
	if f.Function == nil {
		return "@<nil>"
	}

	if f.SelfObject != nil {
		return "@" + f.SelfObject.String() + "." + f.Function.Name.Value
	}

	return "@" + f.Function.Name.Value
}

// ============================================================================
// ErrorValue - Runtime representation for error values
// ============================================================================
// Task 3.7.7: Simple error value type for builtin functions.

// ErrorValue represents an error runtime value returned by builtin functions.
type ErrorValue struct {
	Message string
}

// Type returns "ERROR".
func (e *ErrorValue) Type() string {
	return "ERROR"
}

// String returns the error message.
func (e *ErrorValue) String() string {
	return e.Message
}
