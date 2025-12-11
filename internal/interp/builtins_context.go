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

// NewError creates an error value with location from the current node.
func (i *Interpreter) NewError(format string, args ...interface{}) builtins.Value {
	return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), format, args...)
}

// CurrentNode returns the AST node currently being evaluated.
// This implements the builtins.Context interface.
func (i *Interpreter) CurrentNode() ast.Node {
	return i.evaluatorInstance.CurrentNode()
}

// RandSource returns the random number generator for built-in functions.
// This implements the builtins.Context interface.
func (i *Interpreter) RandSource() *rand.Rand {
	return i.evaluatorInstance.Random()
}

// GetRandSeed returns the current random number generator seed value.
// This implements the builtins.Context interface.
func (i *Interpreter) GetRandSeed() int64 {
	return i.evaluatorInstance.RandomSeed()
}

// SetRandSeed sets the random number generator seed.
// This implements the builtins.Context interface.
func (i *Interpreter) SetRandSeed(seed int64) {
	i.evaluatorInstance.SetRandomSeed(seed)
}

// UnwrapVariant returns the underlying value if input is a Variant.
func (i *Interpreter) UnwrapVariant(value builtins.Value) builtins.Value {
	if value != nil {
		if variant, ok := value.(*VariantValue); ok {
			if variant.Value == nil {
				return &UnassignedValue{}
			}
			return variant.Value
		}
	}
	return value
}

// Type conversion helpers for builtin functions

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
func (i *Interpreter) ParseJSONString(jsonStr string) (builtins.Value, error) {
	jsonVal, err := parseJSONString(jsonStr)
	if err != nil {
		return nil, err
	}
	return jsonValueToVariant(jsonVal), nil
}

// ValueToJSON converts a DWScript Value to a JSON string.
func (i *Interpreter) ValueToJSON(value builtins.Value, formatted bool) (string, error) {
	return i.ValueToJSONWithIndent(value, formatted, 2)
}

// ValueToJSONWithIndent converts a DWScript Value to a JSON string with custom indentation.
func (i *Interpreter) ValueToJSONWithIndent(value builtins.Value, formatted bool, indent int) (string, error) {
	jsonVal := valueToJSONValue(value)

	var jsonBytes []byte
	var err error
	if formatted {
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

// Type introspection helpers

// GetTypeOf returns the type name of a value.
func (i *Interpreter) GetTypeOf(value builtins.Value) string {
	if value == nil {
		return "Null"
	}

	// Handle object/class types - return class name
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

	// Convert internal uppercase names to DWScript format
	switch typeName := value.Type(); typeName {
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
func (i *Interpreter) GetClassOf(value builtins.Value) string {
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
func (i *Interpreter) JSONHasField(value builtins.Value, fieldName string) bool {
	val := unwrapVariant(value)

	// Check if it's a JSON value
	jsonVal, ok := val.(*JSONValue)
	if !ok || jsonVal.Value == nil || jsonVal.Value.Kind() != 2 { // KindObject = 2
		return false
	}
	return jsonVal.Value.ObjectGet(fieldName) != nil
}

// JSONGetKeys returns the keys of a JSON object in insertion order.
func (i *Interpreter) JSONGetKeys(value builtins.Value) []string {
	val := unwrapVariant(value)
	jsonVal, ok := val.(*JSONValue)
	if !ok || jsonVal.Value == nil || jsonVal.Value.Kind() != 2 {
		return []string{}
	}
	return jsonVal.Value.ObjectKeys()
}

// JSONGetValues returns the values of a JSON object/array.
func (i *Interpreter) JSONGetValues(value builtins.Value) []builtins.Value {
	val := unwrapVariant(value)
	jsonVal, ok := val.(*JSONValue)
	if !ok || jsonVal.Value == nil {
		return []builtins.Value{}
	}

	// Handle objects (KindObject = 2)
	if jsonVal.Value.Kind() == 2 {
		keys := jsonVal.Value.ObjectKeys()
		values := make([]builtins.Value, len(keys))
		for idx, key := range keys {
			values[idx] = jsonValueToVariant(jsonVal.Value.ObjectGet(key))
		}
		return values
	}

	// Handle arrays (KindArray = 3)
	if jsonVal.Value.Kind() == 3 {
		arrayLen := jsonVal.Value.ArrayLen()
		values := make([]builtins.Value, arrayLen)
		for idx := 0; idx < arrayLen; idx++ {
			values[idx] = jsonValueToVariant(jsonVal.Value.ArrayGet(idx))
		}
		return values
	}

	return []builtins.Value{}
}

// JSONGetLength returns the length of a JSON array or object.
func (i *Interpreter) JSONGetLength(value builtins.Value) int {
	val := unwrapVariant(value)
	jsonVal, ok := val.(*JSONValue)
	if !ok || jsonVal.Value == nil {
		return 0
	}

	switch jsonVal.Value.Kind() {
	case 2: // KindObject
		return len(jsonVal.Value.ObjectKeys())
	case 3: // KindArray
		return jsonVal.Value.ArrayLen()
	default:
		return 0
	}
}

// CreateStringArray creates an array of strings from a slice of string values.
func (i *Interpreter) CreateStringArray(values []string) builtins.Value {
	elements := make([]Value, len(values))
	for idx, str := range values {
		elements[idx] = &StringValue{Value: str}
	}
	return &ArrayValue{
		Elements:  elements,
		ArrayType: types.NewDynamicArrayType(types.STRING),
	}
}

// CreateVariantArray creates an array of Variants from a slice of values.
func (i *Interpreter) CreateVariantArray(values []builtins.Value) builtins.Value {
	return &ArrayValue{
		Elements:  values,
		ArrayType: types.NewDynamicArrayType(types.VARIANT),
	}
}

// Write writes a string to the output without a newline.
func (i *Interpreter) Write(s string) {
	if i.output != nil {
		i.output.Write([]byte(s))
	}
}

// WriteLine writes a string to the output followed by a newline.

func (i *Interpreter) WriteLine(s string) {
	if i.output != nil {
		i.output.Write([]byte(s + "\n"))
	}
}

// GetEnumOrdinal returns the ordinal value of an enum Value.
func (i *Interpreter) GetEnumOrdinal(value builtins.Value) (int64, bool) {
	if enumVal, ok := value.(*EnumValue); ok {
		return int64(enumVal.OrdinalValue), true
	}
	return 0, false
}

// GetJSONVarType returns the VarType code for a JSON value based on its kind.
func (i *Interpreter) GetJSONVarType(value builtins.Value) (int64, bool) {
	jsonVal, ok := value.(*JSONValue)
	if !ok {
		return 0, false
	}

	// Return VarType code based on JSON kind
	if jsonVal.Value == nil {
		return varEmpty, true
	}
	return jsonKindToVarType(jsonVal.Value.Kind()), true
}

// GetBuiltinArrayLength returns the length of an array for builtin functions.
func (i *Interpreter) GetBuiltinArrayLength(value builtins.Value) (int64, bool) {
	arrayVal, ok := value.(*ArrayValue)
	if !ok {
		return 0, false
	}
	return int64(len(arrayVal.Elements)), true
}

// SetArrayLength resizes a dynamic array to the specified length.
func (i *Interpreter) SetArrayLength(array builtins.Value, newLength int) error {
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
	if newLength < currentLength {
		arrayVal.Elements = arrayVal.Elements[:newLength]
	} else if newLength > currentLength {
		additional := make([]Value, newLength-currentLength)
		arrayVal.Elements = append(arrayVal.Elements, additional...)
	}
	return nil
}

// ArrayCopy creates a deep copy of an array value.
func (i *Interpreter) ArrayCopy(array builtins.Value) builtins.Value {
	arrayVal, ok := array.(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "ArrayCopy() expects array, got %s", array.Type())
	}

	return i.builtinArrayCopy(arrayVal)
}

// ArrayReverse reverses the elements of an array in place.
func (i *Interpreter) ArrayReverse(array builtins.Value) builtins.Value {
	arrayVal, ok := array.(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "ArrayReverse() expects array, got %s", array.Type())
	}
	return i.builtinArrayReverse(arrayVal)
}

// ArraySort sorts the elements of an array in place using default comparison.
func (i *Interpreter) ArraySort(array builtins.Value) builtins.Value {
	arrayVal, ok := array.(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "ArraySort() expects array, got %s", array.Type())
	}

	return i.builtinArraySort(arrayVal)
}

// EvalFunctionPointer calls a function pointer with the given arguments.
func (i *Interpreter) EvalFunctionPointer(funcPtr builtins.Value, args []builtins.Value) builtins.Value {
	lambdaVal, ok := funcPtr.(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "EvalFunctionPointer() expects function pointer, got %s", funcPtr.Type())
	}
	return i.callFunctionPointer(lambdaVal, args, i.evaluatorInstance.CurrentNode())
}

// GetCallStackString returns a formatted string representation of the current call stack.
func (i *Interpreter) GetCallStackString() string {
	return i.callStack.String()
}

// GetCallStackArray returns the current call stack as an array of records.
func (i *Interpreter) GetCallStackArray() builtins.Value {
	elements := make([]Value, len(i.callStack))

	for idx, frame := range i.callStack {
		fields := make(map[string]Value)
		fields["FunctionName"] = &StringValue{Value: frame.FunctionName}
		if frame.Position != nil {
			fields["Line"] = &IntegerValue{Value: int64(frame.Position.Line)}
			fields["Column"] = &IntegerValue{Value: int64(frame.Position.Column)}
		} else {
			fields["Line"] = &IntegerValue{Value: 0}
			fields["Column"] = &IntegerValue{Value: 0}
		}
		elements[idx] = &RecordValue{Fields: fields, RecordType: nil}
	}

	// Helper to create int pointers
	lowBound := 0
	highBound := len(elements) - 1

	// Create and return the array
	return &ArrayValue{
		Elements: elements,
		ArrayType: &types.ArrayType{
			ElementType: nil,
			LowBound:    &lowBound,
			HighBound:   &highBound,
		},
	}
}

// IsAssigned checks if a value is assigned (not nil, NilValue, or unassigned interface/variant).
func (i *Interpreter) IsAssigned(value builtins.Value) bool {
	if value == nil {
		return false
	}
	if _, ok := value.(*NilValue); ok {
		return false
	}

	// Handle interfaces (runtime.InterfaceInstance)
	if intfVal, ok := value.(*runtime.InterfaceInstance); ok {
		return intfVal.Object != nil
	}

	// Handle objects
	if objVal, ok := value.(*ObjectInstance); ok {
		return objVal != nil
	}

	// Handle Variant values - unwrap and check
	if varVal, ok := value.(*VariantValue); ok {
		return i.IsAssigned(varVal.Value)
	}

	// All other values are considered assigned
	return true
}

// RaiseException raises a DWScript exception so try/except blocks can handle it.
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
func (i *Interpreter) RaiseAssertionFailed(customMessage string) {
	var message string
	if i.evaluatorInstance.CurrentNode() != nil {
		pos := i.evaluatorInstance.CurrentNode().Pos()
		message = fmt.Sprintf("Assertion failed [line: %d, column: %d]", pos.Line, pos.Column)
	} else {
		message = "Assertion failed"
	}

	// If custom message provided, append it
	if customMessage != "" {
		message = message + " : " + customMessage
	}

	assertClass, ok := i.classes[strings.ToLower("EAssertionFailed")]
	if !ok {
		// Fallback if class not found
		i.exception = &runtime.ExceptionValue{Message: message}
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
	}
}

// CreateContractException creates an exception value for contract violations.
func (i *Interpreter) CreateContractException(className, message string, node ast.Node, classMetadata interface{}, callStack interface{}) interface{} {
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

	return &runtime.ExceptionValue{
		Metadata:  metadata,
		Message:   message,
		Position:  pos,
		CallStack: stack,
	}
}

// CleanupInterfaceReferences implements evaluator.InterpreterAdapter.
func (i *Interpreter) CleanupInterfaceReferences(env interface{}) {
	if envTyped, ok := env.(*Environment); ok {
		i.cleanupInterfaceReferences(envTyped)
	}
}

// GetEnumSuccessor returns the successor of an enum value.
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

	// Find current position
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
	return &EnumValue{
		TypeName:     val.TypeName,
		ValueName:    nextValueName,
		OrdinalValue: enumType.Values[nextValueName],
	}, nil
}

// GetEnumPredecessor returns the predecessor of an enum value.
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

	// Find current position
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
	return &EnumValue{
		TypeName:     val.TypeName,
		ValueName:    prevValueName,
		OrdinalValue: enumType.Values[prevValueName],
	}, nil
}

// ParseInt parses a string to an integer with the specified base (2-36).
func (i *Interpreter) ParseInt(s string, base int) (int64, bool) {
	s = strings.TrimSpace(s)

	// Use strconv.ParseInt for strict parsing
	intValue, err := strconv.ParseInt(s, base, 64)
	if err != nil {
		return 0, false
	}
	return intValue, true
}

// ParseFloat parses a string to a float64.
func (i *Interpreter) ParseFloat(s string) (float64, bool) {
	s = strings.TrimSpace(s)

	// Use strconv.ParseFloat for strict parsing
	floatValue, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, false
	}
	return floatValue, true
}

// FormatString formats a string using Go fmt.Sprintf semantics with DWScript values.
func (i *Interpreter) FormatString(format string, args []builtins.Value) (string, error) {
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

	// Parse format specifiers
	var specs []formatSpec
	argIndex := 0
	iStr := 0
	for iStr < len(format) {
		ch := format[iStr]
		if ch == '%' {
			if iStr+1 < len(format) && format[iStr+1] == '%' {
				iStr += 2
				continue
			}
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

	// Convert DWScript values to Go interface{} values
	goArgs := make([]interface{}, len(args))
	for idx, elem := range args {
		if idx >= len(specs) {
			break
		}
		spec := specs[idx]
		unwrapped := unwrapVariant(elem)

		switch v := unwrapped.(type) {
		case *IntegerValue:
			switch spec.verb {
			case 'd', 'x', 'X', 'o', 'v':
				goArgs[idx] = v.Value
			case 'f':
				goArgs[idx] = normalizeFloat(float64(v.Value))
			case 's':
				goArgs[idx] = fmt.Sprintf("%d", v.Value)
			default:
				return "", fmt.Errorf("cannot use %%%c with Integer value at index %d", spec.verb, idx)
			}
		case *FloatValue:
			switch spec.verb {
			case 'f', 'v':
				goArgs[idx] = normalizeFloat(v.Value)
			case 's':
				goArgs[idx] = fmt.Sprintf("%f", v.Value)
			default:
				return "", fmt.Errorf("cannot use %%%c with Float value at index %d", spec.verb, idx)
			}
		case *StringValue:
			switch spec.verb {
			case 's', 'v':
				goArgs[idx] = v.Value
			case 'd', 'x', 'X', 'o', 'f':
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

	return fmt.Sprintf(format, goArgs...), nil
}

// GetLowBound returns the lower bound for arrays, enums, or type meta-values.
func (i *Interpreter) GetLowBound(value builtins.Value) (builtins.Value, error) {
	// Type meta-values
	if typeMetaVal, ok := value.(*TypeMetaValue); ok {
		switch typeMetaVal.TypeInfo {
		case types.INTEGER:
			return &IntegerValue{Value: math.MinInt64}, nil
		case types.FLOAT:
			return &FloatValue{Value: -math.MaxFloat64}, nil
		case types.BOOLEAN:
			return &BooleanValue{Value: false}, nil
		}
		if enumType, ok := typeMetaVal.TypeInfo.(*types.EnumType); ok {
			if len(enumType.OrderedNames) == 0 {
				return nil, fmt.Errorf("enum type '%s' has no values", typeMetaVal.TypeName)
			}
			firstValueName := enumType.OrderedNames[0]
			return &EnumValue{
				TypeName:     typeMetaVal.TypeName,
				ValueName:    firstValueName,
				OrdinalValue: enumType.Values[firstValueName],
			}, nil
		}
		return nil, fmt.Errorf("Low() not supported for type %s", typeMetaVal.TypeName)
	}

	// Arrays
	if arrayVal, ok := value.(*ArrayValue); ok {
		if arrayVal.ArrayType != nil && arrayVal.ArrayType.IsStatic() {
			return &IntegerValue{Value: int64(*arrayVal.ArrayType.LowBound)}, nil
		}
		return &IntegerValue{Value: 0}, nil
	}

	// Enum values
	if enumVal, ok := value.(*EnumValue); ok {
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
		return &EnumValue{
			TypeName:     enumVal.TypeName,
			ValueName:    firstValueName,
			OrdinalValue: enumType.Values[firstValueName],
		}, nil
	}

	// Strings are 1-indexed in DWScript
	if _, ok := value.(*StringValue); ok {
		return &IntegerValue{Value: 1}, nil
	}

	return nil, fmt.Errorf("Low() expects array, enum, string, or type name, got %s", value.Type())
}

// GetHighBound returns the upper bound for arrays, enums, or type meta-values.
func (i *Interpreter) GetHighBound(value builtins.Value) (builtins.Value, error) {
	// Type meta-values
	if typeMetaVal, ok := value.(*TypeMetaValue); ok {
		switch typeMetaVal.TypeInfo {
		case types.INTEGER:
			return &IntegerValue{Value: math.MaxInt64}, nil
		case types.FLOAT:
			return &FloatValue{Value: math.MaxFloat64}, nil
		case types.BOOLEAN:
			return &BooleanValue{Value: true}, nil
		}
		if enumType, ok := typeMetaVal.TypeInfo.(*types.EnumType); ok {
			if len(enumType.OrderedNames) == 0 {
				return nil, fmt.Errorf("enum type '%s' has no values", typeMetaVal.TypeName)
			}
			lastValueName := enumType.OrderedNames[len(enumType.OrderedNames)-1]
			return &EnumValue{
				TypeName:     typeMetaVal.TypeName,
				ValueName:    lastValueName,
				OrdinalValue: enumType.Values[lastValueName],
			}, nil
		}
		return nil, fmt.Errorf("High() not supported for type %s", typeMetaVal.TypeName)
	}

	// Arrays
	if arrayVal, ok := value.(*ArrayValue); ok {
		if arrayVal.ArrayType != nil && arrayVal.ArrayType.IsStatic() {
			return &IntegerValue{Value: int64(*arrayVal.ArrayType.HighBound)}, nil
		}
		return &IntegerValue{Value: int64(len(arrayVal.Elements) - 1)}, nil
	}

	// Enum values
	if enumVal, ok := value.(*EnumValue); ok {
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
		return &EnumValue{
			TypeName:     enumVal.TypeName,
			ValueName:    lastValueName,
			OrdinalValue: enumType.Values[lastValueName],
		}, nil
	}

	// Strings - return length (Unicode code points)
	if strVal, ok := value.(*StringValue); ok {
		return &IntegerValue{Value: int64(runeLength(strVal.Value))}, nil
	}

	return nil, fmt.Errorf("High() expects array, enum, string, or type name, got %s", value.Type())
}

// ConcatStrings concatenates multiple string values.
func (i *Interpreter) ConcatStrings(args []builtins.Value) builtins.Value {
	var result strings.Builder
	for idx, arg := range args {
		strVal, ok := arg.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Concat() expects string as argument %d, got %s", idx+1, arg.Type())
		}
		result.WriteString(strVal.Value)
	}
	return &StringValue{Value: result.String()}
}

// =============================================================================
// VarParamContext interface implementation
// =============================================================================

// Ensure Interpreter implements builtins.VarParamContext interface at compile time.
var _ builtins.VarParamContext = (*Interpreter)(nil)

// IsError checks if a value is an error value.
func (i *Interpreter) IsError(value builtins.Value) bool {
	return isError(value)
}

// GetVariable retrieves a variable's value by name from the current environment.
func (i *Interpreter) GetVariable(name string) (builtins.Value, bool) {
	return i.Env().Get(name)
}

// SetVariable sets a variable's value by name in the current environment.
func (i *Interpreter) SetVariable(name string, value builtins.Value) error {
	return i.Env().Set(name, value)
}

// EvaluateLValue evaluates an lvalue expression once and returns the current value
// and a closure function to assign a new value to that lvalue.
func (i *Interpreter) EvaluateLValue(lvalue ast.Expression) (builtins.Value, builtins.LValueAssignFunc, error) {
	return i.evaluateLValue(lvalue)
}

// DereferenceValue unwraps a ReferenceValue to get the actual value.
func (i *Interpreter) DereferenceValue(value builtins.Value) (builtins.Value, error) {
	if ref, isRef := value.(*ReferenceValue); isRef {
		return ref.Dereference()
	}
	return value, nil
}

// AssignToReference assigns a value to a ReferenceValue.
func (i *Interpreter) AssignToReference(ref builtins.Value, value builtins.Value) error {
	if refVal, isRef := ref.(*ReferenceValue); isRef {
		return refVal.Assign(value)
	}
	return fmt.Errorf("cannot assign to non-reference value")
}

// IsReference checks if a value is a ReferenceValue.
func (i *Interpreter) IsReference(value builtins.Value) bool {
	_, isRef := value.(*ReferenceValue)
	return isRef
}

// CreateIntegerValue creates a new IntegerValue with the given value.
func (i *Interpreter) CreateIntegerValue(value int64) builtins.Value {
	return &IntegerValue{Value: value}
}

// CreateStringValue creates a new StringValue with the given value.
func (i *Interpreter) CreateStringValue(value string) builtins.Value {
	return &StringValue{Value: value}
}

// CreateNilValue creates a new NilValue.
func (i *Interpreter) CreateNilValue() builtins.Value {
	return &NilValue{}
}

// GetEnumMetadata retrieves enum type metadata by type name.
func (i *Interpreter) GetEnumMetadata(typeName string) builtins.Value {
	metadata := i.typeSystem.LookupEnumMetadata(typeName)
	if metadata == nil {
		return nil
	}
	if val, ok := metadata.(builtins.Value); ok {
		return val
	}
	return nil
}

// CreateEnumValue creates a new EnumValue with the given type and value information.
func (i *Interpreter) CreateEnumValue(typeName, valueName string, ordinal int64) builtins.Value {
	return &EnumValue{
		TypeName:     typeName,
		ValueName:    valueName,
		OrdinalValue: int(ordinal),
	}
}

// RuneInsert inserts source into target at position (1-based).
func (i *Interpreter) RuneInsert(source, target string, pos int) string {
	return runeInsert(source, target, pos)
}

// RuneDelete deletes count characters from str starting at pos (1-based).
func (i *Interpreter) RuneDelete(str string, pos, count int) string {
	return runeDelete(str, pos, count)
}

// RuneSetLength resizes a string to newLength characters.
func (i *Interpreter) RuneSetLength(str string, newLength int) string {
	return runeSetLength(str, newLength)
}
