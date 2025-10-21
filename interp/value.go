// Package interp provides the interpreter and runtime for DWScript.
package interp

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/types"
)

// Value represents a runtime value in the DWScript interpreter.
// All runtime values must implement this interface.
// This interface does NOT use interface{} to ensure type safety.
type Value interface {
	// Type returns the type name of the value (e.g., "INTEGER", "STRING")
	Type() string
	// String returns the string representation of the value
	String() string
}

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
	return strconv.FormatFloat(f.Value, 'g', -1, 64)
}

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

// BooleanValue represents a boolean value in DWScript.
type BooleanValue struct {
	Value bool
}

// Type returns "BOOLEAN".
func (b *BooleanValue) Type() string {
	return "BOOLEAN"
}

// String returns "true" or "false".
func (b *BooleanValue) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

// NilValue represents a nil/null value in DWScript.
type NilValue struct{}

// Type returns "NIL".
func (n *NilValue) Type() string {
	return "NIL"
}

// String returns "nil".
func (n *NilValue) String() string {
	return "nil"
}

// EnumValue represents an enum value in DWScript.
// Task 8.49: Store enum values with their ordinal value and type name.
type EnumValue struct {
	TypeName     string // Enum type name (e.g., "TColor")
	ValueName    string // Enum value name (e.g., "Red")
	OrdinalValue int    // The ordinal value (e.g., 0 for Red if implicit)
}

// Type returns "ENUM".
func (e *EnumValue) Type() string {
	return "ENUM"
}

// String returns the enum value name.
func (e *EnumValue) String() string {
	return e.ValueName
}

// RecordValue represents a record value in DWScript.
// Task 8.73: Store record type metadata and field values.
// Records are value types (like structs) with fields.
type RecordValue struct {
	RecordType *types.RecordType  // The record type metadata
	Fields     map[string]Value    // Field name -> runtime value mapping
}

// Type returns "RECORD".
func (r *RecordValue) Type() string {
	return "RECORD"
}

// String returns the string representation of the record.
func (r *RecordValue) String() string {
	var sb strings.Builder

	// Show type name if available
	if r.RecordType != nil && r.RecordType.Name != "" {
		sb.WriteString(r.RecordType.Name)
		sb.WriteString("(")
	} else {
		sb.WriteString("record(")
	}

	// Sort field names for consistent output
	fieldNames := make([]string, 0, len(r.Fields))
	for name := range r.Fields {
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)

	// Add field values
	for i, name := range fieldNames {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(name)
		sb.WriteString(": ")
		if val := r.Fields[name]; val != nil {
			sb.WriteString(val.String())
		} else {
			sb.WriteString("nil")
		}
	}

	sb.WriteString(")")
	return sb.String()
}

// Copy creates a deep copy of the record value (Task 8.77).
// Records have value semantics in DWScript, so assignment should copy.
func (r *RecordValue) Copy() *RecordValue {
	copiedFields := make(map[string]Value, len(r.Fields))

	// Deep copy all fields
	for name, val := range r.Fields {
		// Check if the value is also a record that needs copying
		if recVal, ok := val.(*RecordValue); ok {
			copiedFields[name] = recVal.Copy()
		} else {
			// For basic types (Integer, String, etc.), they're already immutable or copied by value
			copiedFields[name] = val
		}
	}

	return &RecordValue{
		RecordType: r.RecordType,
		Fields:     copiedFields,
	}
}

// ExternalVarValue represents an external variable marker.
// Task 7.144: This is a special marker stored in the environment to indicate
// that a variable is external. Attempting to read or write this value raises an error.
type ExternalVarValue struct {
	Name         string // The variable name in DWScript
	ExternalName string // The external name for FFI binding (may be empty)
}

// Type returns "EXTERNAL_VAR".
func (e *ExternalVarValue) Type() string {
	return "EXTERNAL_VAR"
}

// String returns a description of the external variable.
func (e *ExternalVarValue) String() string {
	if e.ExternalName != "" {
		return fmt.Sprintf("external(%s -> %s)", e.Name, e.ExternalName)
	}
	return fmt.Sprintf("external(%s)", e.Name)
}

// Helper functions to create values from Go types

// NewIntegerValue creates a new IntegerValue from an int64.
func NewIntegerValue(v int64) Value {
	return &IntegerValue{Value: v}
}

// NewFloatValue creates a new FloatValue from a float64.
func NewFloatValue(v float64) Value {
	return &FloatValue{Value: v}
}

// NewStringValue creates a new StringValue from a string.
func NewStringValue(v string) Value {
	return &StringValue{Value: v}
}

// NewBooleanValue creates a new BooleanValue from a bool.
func NewBooleanValue(v bool) Value {
	return &BooleanValue{Value: v}
}

// NewNilValue creates a new NilValue.
func NewNilValue() Value {
	return &NilValue{}
}

// NewRecordValue creates a new RecordValue with the given record type.
// Task 8.73: Initialize the fields map.
func NewRecordValue(recordType *types.RecordType) Value {
	return &RecordValue{
		RecordType: recordType,
		Fields:     make(map[string]Value),
	}
}

// ClassInfoValue is a special internal value type used to track the current class context
// in class methods. It wraps a ClassInfo pointer and is stored as "__CurrentClass__"
// in the environment when executing class methods.
type ClassInfoValue struct {
	ClassInfo *ClassInfo
}

// Type returns "CLASSINFO".
func (c *ClassInfoValue) Type() string {
	return "CLASSINFO"
}

// String returns the class name.
func (c *ClassInfoValue) String() string {
	return "class " + c.ClassInfo.Name
}

// GoInt converts a Value to a Go int64. Returns error if not an IntegerValue.
func GoInt(v Value) (int64, error) {
	if iv, ok := v.(*IntegerValue); ok {
		return iv.Value, nil
	}
	return 0, fmt.Errorf("value is not an integer: %s", v.Type())
}

// GoFloat converts a Value to a Go float64. Returns error if not a FloatValue.
func GoFloat(v Value) (float64, error) {
	if fv, ok := v.(*FloatValue); ok {
		return fv.Value, nil
	}
	return 0, fmt.Errorf("value is not a float: %s", v.Type())
}

// GoString converts a Value to a Go string. Returns error if not a StringValue.
func GoString(v Value) (string, error) {
	if sv, ok := v.(*StringValue); ok {
		return sv.Value, nil
	}
	return "", fmt.Errorf("value is not a string: %s", v.Type())
}

// GoBool converts a Value to a Go bool. Returns error if not a BooleanValue.
func GoBool(v Value) (bool, error) {
	if bv, ok := v.(*BooleanValue); ok {
		return bv.Value, nil
	}
	return false, fmt.Errorf("value is not a boolean: %s", v.Type())
}
