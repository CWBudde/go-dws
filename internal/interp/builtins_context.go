package interp

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/interp/builtins"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Ensure Interpreter implements builtins.Context interface at compile time.
var _ builtins.Context = (*Interpreter)(nil)

// NewError creates an error value with location information from the current node.
// This implements the builtins.Context interface.
func (i *Interpreter) NewError(format string, args ...interface{}) builtins.Value {
	return i.newErrorWithLocation(i.currentNode, format, args...)
}

// CurrentNode returns the AST node currently being evaluated.
// This implements the builtins.Context interface.
func (i *Interpreter) CurrentNode() ast.Node {
	return i.currentNode
}

// RandSource returns the random number generator for built-in functions.
// This implements the builtins.Context interface.
func (i *Interpreter) RandSource() *rand.Rand {
	return i.rand
}

// GetRandSeed returns the current random number generator seed value.
// This implements the builtins.Context interface.
func (i *Interpreter) GetRandSeed() int64 {
	return i.randSeed
}

// SetRandSeed sets the random number generator seed.
// This implements the builtins.Context interface.
func (i *Interpreter) SetRandSeed(seed int64) {
	i.randSeed = seed
	i.rand = rand.New(rand.NewSource(seed))
}

// UnwrapVariant returns the underlying value if input is a Variant, otherwise returns input as-is.
// This implements the builtins.Context interface.
func (i *Interpreter) UnwrapVariant(value builtins.Value) builtins.Value {
	if value != nil {
		// Check if it's a VariantValue and unwrap it
		if variant, ok := value.(*VariantValue); ok {
			if variant.Value == nil {
				return &UnassignedValue{}
			}
			return variant.Value
		}
	}
	return value
}

// ToInt64 converts a Value to int64, handling SubrangeValue and EnumValue.
// This implements the builtins.Context interface.
// Task 3.7.3: Type helper for conversion functions.
func (i *Interpreter) ToInt64(value builtins.Value) (int64, bool) {
	switch v := value.(type) {
	case *IntegerValue:
		return v.Value, true
	case *SubrangeValue:
		return int64(v.Value), true
	case *EnumValue:
		return int64(v.OrdinalValue), true
	case *BooleanValue:
		if v.Value {
			return 1, true
		}
		return 0, true
	case *FloatValue:
		return int64(v.Value), true
	default:
		return 0, false
	}
}

// ToBool converts a Value to bool.
// This implements the builtins.Context interface.
// Task 3.7.3: Type helper for conversion functions.
func (i *Interpreter) ToBool(value builtins.Value) (bool, bool) {
	switch v := value.(type) {
	case *BooleanValue:
		return v.Value, true
	case *IntegerValue:
		return v.Value != 0, true
	case *SubrangeValue:
		return v.Value != 0, true
	case *EnumValue:
		return v.OrdinalValue != 0, true
	default:
		return false, false
	}
}

// ToFloat64 converts a Value to float64, handling integer types.
// This implements the builtins.Context interface.
// Task 3.7.3: Type helper for conversion functions.
func (i *Interpreter) ToFloat64(value builtins.Value) (float64, bool) {
	switch v := value.(type) {
	case *FloatValue:
		return v.Value, true
	case *IntegerValue:
		return float64(v.Value), true
	case *SubrangeValue:
		return float64(v.Value), true
	case *EnumValue:
		return float64(v.OrdinalValue), true
	default:
		return 0.0, false
	}
}

// ParseJSONString parses a JSON string and returns a Value (Variant containing JSONValue).
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for ParseJSON function.
func (i *Interpreter) ParseJSONString(jsonStr string) (builtins.Value, error) {
	// Parse JSON using the existing helper function
	jsonVal, err := parseJSONString(jsonStr)
	if err != nil {
		return nil, err
	}

	// Convert to Variant containing JSONValue
	variant := jsonValueToVariant(jsonVal)
	return variant, nil
}

// ValueToJSON converts a DWScript Value to a JSON string.
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for ToJSON and ToJSONFormatted functions.
func (i *Interpreter) ValueToJSON(value builtins.Value, formatted bool) (string, error) {
	return i.ValueToJSONWithIndent(value, formatted, 2)
}

// ValueToJSONWithIndent converts a DWScript Value to a JSON string with custom indentation.
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for ToJSONFormatted function with custom indent.
func (i *Interpreter) ValueToJSONWithIndent(value builtins.Value, formatted bool, indent int) (string, error) {
	// Convert Value to jsonvalue.Value using existing helper
	jsonVal := valueToJSONValue(value)

	// Serialize to JSON string using encoding/json
	var jsonBytes []byte
	var err error
	if formatted {
		// Build indent string
		indentStr := ""
		for j := 0; j < indent; j++ {
			indentStr += " "
		}
		jsonBytes, err = json.MarshalIndent(jsonVal, "", indentStr)
	} else {
		jsonBytes, err = json.Marshal(jsonVal)
	}

	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// GetTypeOf returns the type name of a value.
// This implements the builtins.Context interface.
// Task 3.7.6: Type introspection helper for TypeOf function.
func (i *Interpreter) GetTypeOf(value builtins.Value) string {
	if value == nil {
		return "Null"
	}

	// Special handling for ObjectInstance - return the class name
	if objVal, ok := value.(*ObjectInstance); ok {
		if objVal.Class != nil {
			return objVal.Class.GetName()
		}
		return "Object"
	}

	// Special handling for ClassValue - return the class name
	// ClassValue represents a class reference (e.g., TMyClass as a value)
	if classVal, ok := value.(*ClassValue); ok {
		if classVal.ClassInfo != nil {
			return classVal.ClassInfo.Name
		}
		return "Class"
	}

	// Special handling for ClassInfoValue - return the class name
	if classInfoVal, ok := value.(*ClassInfoValue); ok {
		if classInfoVal.ClassInfo != nil {
			return classInfoVal.ClassInfo.Name
		}
		return "Class"
	}

	typeName := value.Type()

	// Convert internal type names to DWScript format (proper capitalization)
	// Internal names are uppercase (INTEGER, FLOAT, etc.)
	// DWScript expects: Integer, Float, String, Boolean, etc.
	switch typeName {
	case "INTEGER":
		return "Integer"
	case "FLOAT":
		return "Float"
	case "STRING":
		return "String"
	case "BOOLEAN":
		return "Boolean"
	case "NIL", "NULL":
		return "Null"
	case "ARRAY":
		return "Array"
	case "RECORD":
		return "Record"
	default:
		// For other types (enum names, etc.),
		// return as-is since they already have the proper capitalization
		return typeName
	}
}

// GetClassOf returns the class name of an object value.
// This implements the builtins.Context interface.
// Task 3.7.6: Type introspection helper for TypeOfClass function.
func (i *Interpreter) GetClassOf(value builtins.Value) string {
	// Handle ObjectInstance - return the class name
	if objVal, ok := value.(*ObjectInstance); ok {
		if objVal.Class != nil {
			return objVal.Class.GetName()
		}
	}

	// Handle ClassValue - return the class name (for class references like TMyClass)
	if classVal, ok := value.(*ClassValue); ok {
		if classVal.ClassInfo != nil {
			return classVal.ClassInfo.Name
		}
	}

	// Handle ClassInfoValue - return the class name
	if classInfoVal, ok := value.(*ClassInfoValue); ok {
		if classInfoVal.ClassInfo != nil {
			return classInfoVal.ClassInfo.Name
		}
	}

	return ""
}

// JSONHasField checks if a JSON object value has a given field.
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for JSONHasField function.
func (i *Interpreter) JSONHasField(value builtins.Value, fieldName string) bool {
	// Unwrap variant
	val := unwrapVariant(value)

	// Check if it's a JSON value
	jsonVal, ok := val.(*JSONValue)
	if !ok {
		return false
	}

	// Check if it's an object
	if jsonVal.Value == nil || jsonVal.Value.Kind() != 2 { // KindObject = 2
		return false
	}

	// Check if field exists
	fieldValue := jsonVal.Value.ObjectGet(fieldName)
	return fieldValue != nil
}

// JSONGetKeys returns the keys of a JSON object in insertion order.
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for JSONKeys function.
func (i *Interpreter) JSONGetKeys(value builtins.Value) []string {
	// Unwrap variant
	val := unwrapVariant(value)

	// Check if it's a JSON value
	jsonVal, ok := val.(*JSONValue)
	if !ok {
		return []string{}
	}

	// Check if it's an object
	if jsonVal.Value == nil || jsonVal.Value.Kind() != 2 { // KindObject = 2
		return []string{}
	}

	// Get keys
	return jsonVal.Value.ObjectKeys()
}

// JSONGetValues returns the values of a JSON object/array.
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for JSONValues function.
func (i *Interpreter) JSONGetValues(value builtins.Value) []builtins.Value {
	// Unwrap variant
	val := unwrapVariant(value)

	// Check if it's a JSON value
	jsonVal, ok := val.(*JSONValue)
	if !ok {
		return []builtins.Value{}
	}

	if jsonVal.Value == nil {
		return []builtins.Value{}
	}

	// Handle objects
	if jsonVal.Value.Kind() == 2 { // KindObject = 2
		keys := jsonVal.Value.ObjectKeys()
		values := make([]builtins.Value, len(keys))
		for idx, key := range keys {
			fieldVal := jsonVal.Value.ObjectGet(key)
			// Wrap in JSONValue and then Variant
			values[idx] = jsonValueToVariant(fieldVal)
		}
		return values
	}

	// Handle arrays
	if jsonVal.Value.Kind() == 3 { // KindArray = 3
		arrayLen := jsonVal.Value.ArrayLen()
		values := make([]builtins.Value, arrayLen)
		for idx := 0; idx < arrayLen; idx++ {
			elemVal := jsonVal.Value.ArrayGet(idx)
			// Wrap in JSONValue and then Variant
			values[idx] = jsonValueToVariant(elemVal)
		}
		return values
	}

	return []builtins.Value{}
}

// JSONGetLength returns the length of a JSON array or object.
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for JSONLength function.
func (i *Interpreter) JSONGetLength(value builtins.Value) int {
	// Unwrap variant
	val := unwrapVariant(value)

	// Check if it's a JSON value
	jsonVal, ok := val.(*JSONValue)
	if !ok {
		return 0
	}

	if jsonVal.Value == nil {
		return 0
	}

	// Handle objects - return number of keys
	if jsonVal.Value.Kind() == 2 { // KindObject = 2
		return len(jsonVal.Value.ObjectKeys())
	}

	// Handle arrays - return length
	if jsonVal.Value.Kind() == 3 { // KindArray = 3
		return jsonVal.Value.ArrayLen()
	}

	return 0
}

// CreateStringArray creates an array of strings from a slice of string values.
// This implements the builtins.Context interface.
// Task 3.7.6: Helper for creating string arrays in JSON functions.
func (i *Interpreter) CreateStringArray(values []string) builtins.Value {
	// Convert strings to StringValue elements
	elements := make([]Value, len(values))
	for idx, str := range values {
		elements[idx] = &StringValue{Value: str}
	}

	// Create array type: array of String
	arrayType := types.NewDynamicArrayType(types.STRING)

	return &ArrayValue{
		Elements:  elements,
		ArrayType: arrayType,
	}
}

// CreateVariantArray creates an array of Variants from a slice of values.
// This implements the builtins.Context interface.
// Task 3.7.6: Helper for creating variant arrays in JSON functions.
func (i *Interpreter) CreateVariantArray(values []builtins.Value) builtins.Value {
	// Create array type: array of Variant
	arrayType := types.NewDynamicArrayType(types.VARIANT)

	return &ArrayValue{
		Elements:  values,
		ArrayType: arrayType,
	}
}

// Write writes a string to the output without a newline.
// This implements the builtins.Context interface.
// Task 3.7.4: I/O helper for Print function.
func (i *Interpreter) Write(s string) {
	// If output is nil, silently discard output (some tests use New(nil))
	if i.output != nil {
		i.output.Write([]byte(s))
	}
}

// WriteLine writes a string to the output followed by a newline.
// This implements the builtins.Context interface.
// Task 3.7.4: I/O helper for PrintLn function.
func (i *Interpreter) WriteLine(s string) {
	// If output is nil, silently discard output (some tests use New(nil))
	if i.output != nil {
		i.output.Write([]byte(s + "\n"))
	}
}

// GetEnumOrdinal returns the ordinal value of an enum Value.
// This implements the builtins.Context interface.
// Task 3.7.5: Helper for Ord() function.
func (i *Interpreter) GetEnumOrdinal(value builtins.Value) (int64, bool) {
	if enumVal, ok := value.(*EnumValue); ok {
		return int64(enumVal.OrdinalValue), true
	}
	return 0, false
}

// GetJSONVarType returns the VarType code for a JSON value based on its kind.
// This implements the builtins.Context interface.
// Task 3.7.5: Helper for VarType() function to handle JSON values.
func (i *Interpreter) GetJSONVarType(value builtins.Value) (int64, bool) {
	// Check if it's a JSON value
	jsonVal, ok := value.(*JSONValue)
	if !ok {
		return 0, false
	}

	// Return VarType code based on JSON kind
	if jsonVal.Value == nil {
		return varEmpty, true
	}

	typeCode := jsonKindToVarType(jsonVal.Value.Kind())
	return typeCode, true
}

// GetBuiltinArrayLength returns the length of an array for builtin functions.
// This implements the builtins.Context interface.
// Task 3.7.7: Helper for Length() function on arrays.
func (i *Interpreter) GetBuiltinArrayLength(value builtins.Value) (int64, bool) {
	arrayVal, ok := value.(*ArrayValue)
	if !ok {
		return 0, false
	}
	return int64(len(arrayVal.Elements)), true
}

// SetArrayLength resizes a dynamic array to the specified length.
// This implements the builtins.Context interface.
// Task 3.7.7: Helper for SetLength() function on arrays.
func (i *Interpreter) SetArrayLength(array builtins.Value, newLength int) error {
	// Handle arrays
	arrayVal, ok := array.(*ArrayValue)
	if !ok {
		return fmt.Errorf("SetArrayLength() expects array, got %s", array.Type())
	}

	// Check that it's a dynamic array
	if arrayVal.ArrayType == nil {
		return fmt.Errorf("array has no type information")
	}

	if arrayVal.ArrayType.IsStatic() {
		return fmt.Errorf("SetArrayLength() can only be used with dynamic arrays, not static arrays")
	}

	currentLength := len(arrayVal.Elements)

	if newLength != currentLength {
		if newLength < currentLength {
			// Truncate the slice
			arrayVal.Elements = arrayVal.Elements[:newLength]
		} else {
			// Extend the slice with nil values
			additional := make([]Value, newLength-currentLength)
			arrayVal.Elements = append(arrayVal.Elements, additional...)
		}
	}

	return nil
}

// ArrayCopy creates a deep copy of an array value.
// This implements the builtins.Context interface.
// Task 3.7.7: Helper for Copy() function on arrays.
func (i *Interpreter) ArrayCopy(array builtins.Value) builtins.Value {
	arrayVal, ok := array.(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "ArrayCopy() expects array, got %s", array.Type())
	}

	return i.builtinArrayCopy(arrayVal)
}

// ArrayReverse reverses the elements of an array in place.
// This implements the builtins.Context interface.
// Task 3.7.7: Helper for Reverse() function on arrays.
func (i *Interpreter) ArrayReverse(array builtins.Value) builtins.Value {
	arrayVal, ok := array.(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "ArrayReverse() expects array, got %s", array.Type())
	}

	return i.builtinArrayReverse(arrayVal)
}

// ArraySort sorts the elements of an array in place using default comparison.
// This implements the builtins.Context interface.
// Task 3.7.7: Helper for Sort() function on arrays.
func (i *Interpreter) ArraySort(array builtins.Value) builtins.Value {
	arrayVal, ok := array.(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "ArraySort() expects array, got %s", array.Type())
	}

	return i.builtinArraySort(arrayVal)
}

// EvalFunctionPointer calls a function pointer with the given arguments.
// This implements the builtins.Context interface.
// Task 3.7.7: Helper for collection functions (Map, Filter, Reduce, etc.).
func (i *Interpreter) EvalFunctionPointer(funcPtr builtins.Value, args []builtins.Value) builtins.Value {
	lambdaVal, ok := funcPtr.(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "EvalFunctionPointer() expects function pointer, got %s", funcPtr.Type())
	}

	return i.callFunctionPointer(lambdaVal, args, i.currentNode)
}

// GetCallStackString returns a formatted string representation of the current call stack.
// This implements the builtins.Context interface.
// Task 3.7.8: Helper for GetStackTrace() function.
func (i *Interpreter) GetCallStackString() string {
	return i.callStack.String()
}

// GetCallStackArray returns the current call stack as an array of records.
// This implements the builtins.Context interface.
// Task 3.7.8: Helper for GetCallStack() function.
func (i *Interpreter) GetCallStackArray() builtins.Value {
	// Create an array of records
	elements := make([]Value, len(i.callStack))

	for idx, frame := range i.callStack {
		// Create a record with FunctionName, Line, Column fields
		fields := make(map[string]Value)
		fields["FunctionName"] = &StringValue{Value: frame.FunctionName}

		// Extract line and column from Position
		if frame.Position != nil {
			fields["Line"] = &IntegerValue{Value: int64(frame.Position.Line)}
			fields["Column"] = &IntegerValue{Value: int64(frame.Position.Column)}
		} else {
			fields["Line"] = &IntegerValue{Value: 0}
			fields["Column"] = &IntegerValue{Value: 0}
		}

		record := &RecordValue{
			Fields:     fields,
			RecordType: nil, // No type metadata needed for this simple record
		}

		elements[idx] = record
	}

	// Helper to create int pointers
	lowBound := 0
	highBound := len(elements) - 1

	// Create and return the array
	return &ArrayValue{
		Elements: elements,
		ArrayType: &types.ArrayType{
			ElementType: nil, // Variant or unspecified type
			LowBound:    &lowBound,
			HighBound:   &highBound,
		},
	}
}

// IsAssigned checks if a value is assigned (not nil).
// This implements the builtins.Context interface.
// Task 3.7.8: Helper for Assigned() function.
func (i *Interpreter) IsAssigned(value builtins.Value) bool {
	// Handle nil
	if value == nil {
		return false
	}

	// Handle NilValue
	if _, ok := value.(*NilValue); ok {
		return false
	}

	// Handle interfaces
	if intfVal, ok := value.(*InterfaceInstance); ok {
		return intfVal.Object != nil
	}

	// Handle objects
	if objVal, ok := value.(*ObjectInstance); ok {
		return objVal != nil
	}

	// Handle Variant values - unwrap and check
	if varVal, ok := value.(*VariantValue); ok {
		// Unwrap the variant and recursively check
		return i.IsAssigned(varVal.Value)
	}

	// All other values are considered assigned
	return true
}

// RaiseException raises a DWScript exception so try/except blocks can handle it.
// Builtins call this optionally when they need to surface script-level exceptions.
func (i *Interpreter) RaiseException(className, message string, pos any) {
	var lexerPos *lexer.Position

	switch p := pos.(type) {
	case *lexer.Position:
		lexerPos = p
	case lexer.Position:
		lexerPos = &p
	}

	i.raiseException(className, message, lexerPos)
}

// RaiseAssertionFailed raises an EAssertionFailed exception with an optional custom message.
// This implements the builtins.Context interface.
// Task 3.7.8: Helper for Assert() function.
func (i *Interpreter) RaiseAssertionFailed(customMessage string) {
	// Build the assertion message with position information
	var message string
	if i.currentNode != nil {
		pos := i.currentNode.Pos()
		message = fmt.Sprintf("Assertion failed [line: %d, column: %d]", pos.Line, pos.Column)
	} else {
		message = "Assertion failed"
	}

	// If custom message provided, append it
	if customMessage != "" {
		message = message + " : " + customMessage
	}

	// Create EAssertionFailed exception
	// PR #147: Use lowercase key for O(1) case-insensitive lookup
	assertClass, ok := i.classes[strings.ToLower("EAssertionFailed")]
	if !ok {
		// Fallback: raise EAssertionFailed as a simple error if class not found
		// This should not happen in normal execution
		i.exception = &runtime.ExceptionValue{
			Metadata:  nil,
			ClassInfo: nil,
			Message:   message,
			Instance:  nil,
			Position:  nil,
			CallStack: nil,
		}
		return
	}

	// Create exception instance
	instance := NewObjectInstance(assertClass)

	// Set the Message field
	instance.SetField("Message", &StringValue{Value: message})

	// Create exception value and set it
	i.exception = &runtime.ExceptionValue{
		Metadata:  assertClass.Metadata,
		ClassInfo: assertClass,
		Message:   message,
		Instance:  instance,
		Position:  nil,
		CallStack: nil,
	}
}

// CreateContractException creates an exception value for contract violations.
// This implements the InterpreterAdapter interface.
func (i *Interpreter) CreateContractException(className, message string, node ast.Node, classMetadata interface{}, callStack interface{}) interface{} {
	// Get position information from the AST node
	var pos *lexer.Position
	if node != nil {
		nodePos := node.Pos()
		pos = &nodePos
	}

	// Extract ClassMetadata from interface{} (passed from TypeSystem)
	var metadata *runtime.ClassMetadata
	if classMetadata != nil {
		if md, ok := classMetadata.(*runtime.ClassMetadata); ok {
			metadata = md
		}
	}

	// Extract call stack from interface{}
	var stack errors.StackTrace
	if callStack != nil {
		if st, ok := callStack.(errors.StackTrace); ok {
			stack = st
		}
	}

	// Create ExceptionValue
	exc := &runtime.ExceptionValue{
		Metadata:  metadata,
		Instance:  nil, // Contract exceptions don't need full instance
		Message:   message,
		Position:  pos,
		CallStack: stack,
	}

	return exc
}

// CleanupInterfaceReferences implements evaluator.InterpreterAdapter.
func (i *Interpreter) CleanupInterfaceReferences(env interface{}) {
	if envTyped, ok := env.(*Environment); ok {
		i.cleanupInterfaceReferences(envTyped)
	}
}

// GetEnumSuccessor returns the successor of an enum value.
// This implements the builtins.Context interface.
// Task 3.7.8: Helper for Succ() function.
func (i *Interpreter) GetEnumSuccessor(enumVal builtins.Value) (builtins.Value, error) {
	val, ok := enumVal.(*EnumValue)
	if !ok {
		return nil, fmt.Errorf("expected EnumValue, got %T", enumVal)
	}

	// Get enum type metadata via TypeSystem
	enumMetadata := i.typeSystem.LookupEnumMetadata(val.TypeName)
	if enumMetadata == nil {
		return nil, fmt.Errorf("enum type metadata not found for %s", val.TypeName)
	}

	etv, ok := enumMetadata.(*EnumTypeValue)
	if !ok {
		return nil, fmt.Errorf("invalid enum type metadata for %s", val.TypeName)
	}
	enumType := etv.EnumType

	// Find current value's position in OrderedNames
	currentPos := -1
	for idx, name := range enumType.OrderedNames {
		if name == val.ValueName {
			currentPos = idx
			break
		}
	}

	if currentPos == -1 {
		return nil, fmt.Errorf("enum value '%s' not found in type '%s'", val.ValueName, val.TypeName)
	}

	// Check if we can increment (not at the end)
	if currentPos >= len(enumType.OrderedNames)-1 {
		return nil, fmt.Errorf("cannot get successor of maximum enum value")
	}

	// Get next value
	nextValueName := enumType.OrderedNames[currentPos+1]
	nextOrdinal := enumType.Values[nextValueName]

	// Create new enum value
	return &EnumValue{
		TypeName:     val.TypeName,
		ValueName:    nextValueName,
		OrdinalValue: nextOrdinal,
	}, nil
}

// GetEnumPredecessor returns the predecessor of an enum value.
// This implements the builtins.Context interface.
func (i *Interpreter) GetEnumPredecessor(enumVal builtins.Value) (builtins.Value, error) {
	val, ok := enumVal.(*EnumValue)
	if !ok {
		return nil, fmt.Errorf("expected EnumValue, got %T", enumVal)
	}

	// Get enum type metadata via TypeSystem
	enumMetadata := i.typeSystem.LookupEnumMetadata(val.TypeName)
	if enumMetadata == nil {
		return nil, fmt.Errorf("enum type metadata not found for %s", val.TypeName)
	}

	etv, ok := enumMetadata.(*EnumTypeValue)
	if !ok {
		return nil, fmt.Errorf("invalid enum type metadata for %s", val.TypeName)
	}
	enumType := etv.EnumType

	// Find current value's position in OrderedNames
	currentPos := -1
	for idx, name := range enumType.OrderedNames {
		if name == val.ValueName {
			currentPos = idx
			break
		}
	}

	if currentPos == -1 {
		return nil, fmt.Errorf("enum value '%s' not found in type '%s'", val.ValueName, val.TypeName)
	}

	// Check if we can decrement (not at the beginning)
	if currentPos <= 0 {
		return nil, fmt.Errorf("cannot get predecessor of minimum enum value")
	}

	// Get previous value
	prevValueName := enumType.OrderedNames[currentPos-1]
	prevOrdinal := enumType.Values[prevValueName]

	// Create new enum value
	return &EnumValue{
		TypeName:     val.TypeName,
		ValueName:    prevValueName,
		OrdinalValue: prevOrdinal,
	}, nil
}

// ParseInt parses a string to an integer with the specified base (2-36).
// This implements the builtins.Context interface.
// Task 3.7.8: Helper for StrToIntDef() function.
func (i *Interpreter) ParseInt(s string, base int) (int64, bool) {
	// Use strings.TrimSpace to handle leading/trailing whitespace
	s = strings.TrimSpace(s)

	// Use strconv.ParseInt for strict parsing
	intValue, err := strconv.ParseInt(s, base, 64)
	if err != nil {
		return 0, false
	}

	return intValue, true
}

// ParseFloat parses a string to a float64.
// This implements the builtins.Context interface.
// Task 3.7.8: Helper for StrToFloatDef() function.
func (i *Interpreter) ParseFloat(s string) (float64, bool) {
	// Use strings.TrimSpace to handle leading/trailing whitespace
	s = strings.TrimSpace(s)

	// Use strconv.ParseFloat for strict parsing
	floatValue, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, false
	}

	return floatValue, true
}

// FormatString formats a string using Go fmt.Sprintf semantics with DWScript values.
// This implements the builtins.Context interface.
// Task 3.7.8: Helper for Format() function.
func (i *Interpreter) FormatString(format string, args []builtins.Value) (string, error) {
	// Parse format string to extract format specifiers
	type formatSpec struct {
		verb  rune
		index int
	}
	normalizeFloat := func(f float64) float64 {
		// Clamp tiny magnitudes to zero to avoid "-0.00" artifacts when formatting.
		if math.Abs(f) < 1e-12 || (f == 0 && math.Signbit(f)) {
			return 0
		}
		return f
	}
	var specs []formatSpec
	argIndex := 0

	iStr := 0
	for iStr < len(format) {
		ch := format[iStr]
		if ch == '%' {
			if iStr+1 < len(format) && format[iStr+1] == '%' {
				// %% - literal percent sign
				iStr += 2
				continue
			}
			// Parse format specifier
			iStr++
			// Skip width/precision/flags
			for iStr < len(format) {
				b := format[iStr]
				if (b >= '0' && b <= '9') || b == '.' || b == '+' || b == '-' || b == ' ' || b == '#' {
					iStr++
					continue
				}
				break
			}
			// Get the verb
			if iStr < len(format) {
				verb := rune(format[iStr])
				if verb == 's' || verb == 'd' || verb == 'f' || verb == 'v' || verb == 'x' || verb == 'X' || verb == 'o' {
					specs = append(specs, formatSpec{verb: verb, index: argIndex})
					argIndex++
				}
				iStr++
			}
		} else {
			iStr++
		}
	}

	// Validate that we have the right number of arguments
	if len(specs) != len(args) {
		return "", fmt.Errorf("expects %d arguments for format string, got %d", len(specs), len(args))
	}

	// Validate types and convert DWScript values to Go interface{} values
	goArgs := make([]interface{}, len(args))
	for idx, elem := range args {
		if idx >= len(specs) {
			break
		}
		spec := specs[idx]

		// Unbox Variant values for Format() function
		unwrapped := unwrapVariant(elem)

		switch v := unwrapped.(type) {
		case *IntegerValue:
			// %d, %x, %X, %o, %v are valid for integers
			switch spec.verb {
			case 'd', 'x', 'X', 'o', 'v':
				goArgs[idx] = v.Value
			case 'f':
				// Allow integers with %f by promoting to float64 (Delphi-compatible)
				goArgs[idx] = normalizeFloat(float64(v.Value))
			case 's':
				// Allow integer to string conversion for %s
				goArgs[idx] = fmt.Sprintf("%d", v.Value)
			default:
				return "", fmt.Errorf("cannot use %%%c with Integer value at index %d", spec.verb, idx)
			}
		case *FloatValue:
			// %f, %v are valid for floats
			switch spec.verb {
			case 'f', 'v':
				goArgs[idx] = normalizeFloat(v.Value)
			case 's':
				// Allow float to string conversion for %s
				goArgs[idx] = fmt.Sprintf("%f", v.Value)
			default:
				return "", fmt.Errorf("cannot use %%%c with Float value at index %d", spec.verb, idx)
			}
		case *StringValue:
			// %s, %v are valid for strings
			switch spec.verb {
			case 's', 'v':
				goArgs[idx] = v.Value
			case 'd', 'x', 'X', 'o':
				// String cannot be used with integer format specifiers
				return "", fmt.Errorf("cannot use %%%c with String value at index %d", spec.verb, idx)
			case 'f':
				// String cannot be used with float format specifiers
				return "", fmt.Errorf("cannot use %%%c with String value at index %d", spec.verb, idx)
			default:
				goArgs[idx] = v.Value
			}
		case *BooleanValue:
			goArgs[idx] = v.Value
		default:
			return "", fmt.Errorf("cannot format value of type %s at index %d", unwrapped.Type(), idx)
		}
	}

	// Format the string
	result := fmt.Sprintf(format, goArgs...)

	return result, nil
}

// GetLowBound returns the lower bound for arrays, enums, or type meta-values.
// This implements the builtins.Context interface.
// Task 3.7.9: Support for polymorphic Low() function.
func (i *Interpreter) GetLowBound(value builtins.Value) (builtins.Value, error) {
	// Handle type meta-values (type names as values)
	if typeMetaVal, ok := value.(*TypeMetaValue); ok {
		// Handle built-in types
		switch typeMetaVal.TypeInfo {
		case types.INTEGER:
			return &IntegerValue{Value: math.MinInt64}, nil
		case types.FLOAT:
			return &FloatValue{Value: -math.MaxFloat64}, nil
		case types.BOOLEAN:
			return &BooleanValue{Value: false}, nil
		}

		// Handle enum types
		if enumType, ok := typeMetaVal.TypeInfo.(*types.EnumType); ok {
			if len(enumType.OrderedNames) == 0 {
				return nil, fmt.Errorf("enum type '%s' has no values", typeMetaVal.TypeName)
			}
			firstValueName := enumType.OrderedNames[0]
			firstOrdinal := enumType.Values[firstValueName]
			return &EnumValue{
				TypeName:     typeMetaVal.TypeName,
				ValueName:    firstValueName,
				OrdinalValue: firstOrdinal,
			}, nil
		}

		return nil, fmt.Errorf("Low() not supported for type %s", typeMetaVal.TypeName)
	}

	// Handle array values
	if arrayVal, ok := value.(*ArrayValue); ok {
		if arrayVal.ArrayType == nil {
			return &IntegerValue{Value: 0}, nil // Dynamic array default
		}
		if arrayVal.ArrayType.IsStatic() {
			return &IntegerValue{Value: int64(*arrayVal.ArrayType.LowBound)}, nil
		}
		return &IntegerValue{Value: 0}, nil
	}

	// Handle enum values
	if enumVal, ok := value.(*EnumValue); ok {
		// Get enum type metadata via TypeSystem
		enumMetadata := i.typeSystem.LookupEnumMetadata(enumVal.TypeName)
		if enumMetadata == nil {
			return nil, fmt.Errorf("enum type '%s' not found", enumVal.TypeName)
		}
		etv, ok := enumMetadata.(*EnumTypeValue)
		if !ok {
			return nil, fmt.Errorf("invalid enum type metadata for '%s'", enumVal.TypeName)
		}
		enumType := etv.EnumType
		if len(enumType.OrderedNames) == 0 {
			return nil, fmt.Errorf("enum type '%s' has no values", enumVal.TypeName)
		}
		firstValueName := enumType.OrderedNames[0]
		firstOrdinal := enumType.Values[firstValueName]
		return &EnumValue{
			TypeName:     enumVal.TypeName,
			ValueName:    firstValueName,
			OrdinalValue: firstOrdinal,
		}, nil
	}

	return nil, fmt.Errorf("Low() expects array, enum, or type name, got %s", value.Type())
}

// GetHighBound returns the upper bound for arrays, enums, or type meta-values.
// This implements the builtins.Context interface.
// Task 3.7.9: Support for polymorphic High() function.
func (i *Interpreter) GetHighBound(value builtins.Value) (builtins.Value, error) {
	// Handle type meta-values (type names as values)
	if typeMetaVal, ok := value.(*TypeMetaValue); ok {
		// Handle built-in types
		switch typeMetaVal.TypeInfo {
		case types.INTEGER:
			return &IntegerValue{Value: math.MaxInt64}, nil
		case types.FLOAT:
			return &FloatValue{Value: math.MaxFloat64}, nil
		case types.BOOLEAN:
			return &BooleanValue{Value: true}, nil
		}

		// Handle enum types
		if enumType, ok := typeMetaVal.TypeInfo.(*types.EnumType); ok {
			if len(enumType.OrderedNames) == 0 {
				return nil, fmt.Errorf("enum type '%s' has no values", typeMetaVal.TypeName)
			}
			lastValueName := enumType.OrderedNames[len(enumType.OrderedNames)-1]
			lastOrdinal := enumType.Values[lastValueName]
			return &EnumValue{
				TypeName:     typeMetaVal.TypeName,
				ValueName:    lastValueName,
				OrdinalValue: lastOrdinal,
			}, nil
		}

		return nil, fmt.Errorf("High() not supported for type %s", typeMetaVal.TypeName)
	}

	// Handle array values
	if arrayVal, ok := value.(*ArrayValue); ok {
		if arrayVal.ArrayType == nil {
			return &IntegerValue{Value: int64(len(arrayVal.Elements) - 1)}, nil
		}
		if arrayVal.ArrayType.IsStatic() {
			return &IntegerValue{Value: int64(*arrayVal.ArrayType.HighBound)}, nil
		}
		return &IntegerValue{Value: int64(len(arrayVal.Elements) - 1)}, nil
	}

	// Handle enum values
	if enumVal, ok := value.(*EnumValue); ok {
		// Get enum type metadata via TypeSystem
		enumMetadata := i.typeSystem.LookupEnumMetadata(enumVal.TypeName)
		if enumMetadata == nil {
			return nil, fmt.Errorf("enum type '%s' not found", enumVal.TypeName)
		}
		etv, ok := enumMetadata.(*EnumTypeValue)
		if !ok {
			return nil, fmt.Errorf("invalid enum type metadata for '%s'", enumVal.TypeName)
		}
		enumType := etv.EnumType
		if len(enumType.OrderedNames) == 0 {
			return nil, fmt.Errorf("enum type '%s' has no values", enumVal.TypeName)
		}
		lastValueName := enumType.OrderedNames[len(enumType.OrderedNames)-1]
		lastOrdinal := enumType.Values[lastValueName]
		return &EnumValue{
			TypeName:     enumVal.TypeName,
			ValueName:    lastValueName,
			OrdinalValue: lastOrdinal,
		}, nil
	}

	return nil, fmt.Errorf("High() expects array, enum, or type name, got %s", value.Type())
}

// ConcatStrings concatenates multiple string values into a single string.
// This implements the builtins.Context interface.
// Task 3.7.9: Support for polymorphic Concat() function.
func (i *Interpreter) ConcatStrings(args []builtins.Value) builtins.Value {
	// Build the concatenated string
	var result strings.Builder

	for idx, arg := range args {
		strVal, ok := arg.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "Concat() expects string as argument %d, got %s", idx+1, arg.Type())
		}
		result.WriteString(strVal.Value)
	}

	return &StringValue{Value: result.String()}
}
