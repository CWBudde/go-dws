package types

import "github.com/cwbudde/go-dws/pkg/ident"

// BuiltinHelperOperation identifies an intrinsic operation independently of the
// receiver's member name. Aliases share an operation; registration and execution
// use the same identifiers.
type BuiltinHelperOperation string

// Builtin helper operations used by the semantic and execution engines.
const (
	HelperBuiltinNormalizeString BuiltinHelperOperation = "NormalizeString"
	HelperBuiltinPadLeft         BuiltinHelperOperation = "PadLeft"
	HelperBuiltinPadRight        BuiltinHelperOperation = "PadRight"
	HelperBuiltinStrDeleteLeft   BuiltinHelperOperation = "StrDeleteLeft"
	HelperBuiltinStrDeleteRight  BuiltinHelperOperation = "StrDeleteRight"
	HelperBuiltinStripAccents    BuiltinHelperOperation = "StripAccents"
	HelperArrayAdd               BuiltinHelperOperation = "__array_add"
	HelperArrayClear             BuiltinHelperOperation = "__array_clear"
	HelperArrayContains          BuiltinHelperOperation = "__array_contains"
	HelperArrayCopy              BuiltinHelperOperation = "__array_copy"
	HelperArrayCount             BuiltinHelperOperation = "__array_count"
	HelperArrayDelete            BuiltinHelperOperation = "__array_delete"
	HelperArrayFilter            BuiltinHelperOperation = "__array_filter"
	HelperArrayForEach           BuiltinHelperOperation = "__array_foreach"
	HelperArrayHigh              BuiltinHelperOperation = "__array_high"
	HelperArrayIndexOf           BuiltinHelperOperation = "__array_indexof"
	HelperArrayInsert            BuiltinHelperOperation = "__array_insert"
	HelperArrayJoin              BuiltinHelperOperation = "__array_join"
	HelperArrayLength            BuiltinHelperOperation = "__array_length"
	HelperArrayLow               BuiltinHelperOperation = "__array_low"
	HelperArrayMap               BuiltinHelperOperation = "__array_map"
	HelperArrayMove              BuiltinHelperOperation = "__array_move"
	HelperArrayPeek              BuiltinHelperOperation = "__array_peek"
	HelperArrayPop               BuiltinHelperOperation = "__array_pop"
	HelperArrayPush              BuiltinHelperOperation = "__array_push"
	HelperArrayRemove            BuiltinHelperOperation = "__array_remove"
	HelperArrayReverse           BuiltinHelperOperation = "__array_reverse"
	HelperArraySetLength         BuiltinHelperOperation = "__array_setlength"
	HelperArraySort              BuiltinHelperOperation = "__array_sort"
	HelperArraySwap              BuiltinHelperOperation = "__array_swap"
	HelperBooleanToString        BuiltinHelperOperation = "__boolean_tostring"
	HelperEnumName               BuiltinHelperOperation = "__enum_name"
	HelperEnumQualifiedName      BuiltinHelperOperation = "__enum_qualifiedname"
	HelperEnumValue              BuiltinHelperOperation = "__enum_value"
	HelperFloatToStringDefault   BuiltinHelperOperation = "__float_tostring_default"
	HelperFloatToStringPrec      BuiltinHelperOperation = "__float_tostring_prec"
	HelperIntegerToHexString     BuiltinHelperOperation = "__integer_tohexstring"
	HelperIntegerToString        BuiltinHelperOperation = "__integer_tostring"
	HelperStringAfter            BuiltinHelperOperation = "__string_after"
	HelperStringArrayJoin        BuiltinHelperOperation = "__string_array_join"
	HelperStringBefore           BuiltinHelperOperation = "__string_before"
	HelperStringContains         BuiltinHelperOperation = "__string_contains"
	HelperStringCopy             BuiltinHelperOperation = "__string_copy"
	HelperStringEndsWith         BuiltinHelperOperation = "__string_endswith"
	HelperStringIndexOf          BuiltinHelperOperation = "__string_indexof"
	HelperStringIsASCII          BuiltinHelperOperation = "__string_isascii"
	HelperStringLength           BuiltinHelperOperation = "__string_length"
	HelperStringMatches          BuiltinHelperOperation = "__string_matches"
	HelperStringSplit            BuiltinHelperOperation = "__string_split"
	HelperStringStartsWith       BuiltinHelperOperation = "__string_startswith"
	HelperStringToCSSText        BuiltinHelperOperation = "__string_tocsstext"
	HelperStringToFloat          BuiltinHelperOperation = "__string_tofloat"
	HelperStringToHTML           BuiltinHelperOperation = "__string_tohtml"
	HelperStringToHTMLAttribute  BuiltinHelperOperation = "__string_tohtmlattribute"
	HelperStringToInteger        BuiltinHelperOperation = "__string_tointeger"
	HelperStringToJSON           BuiltinHelperOperation = "__string_tojson"
	HelperStringToLower          BuiltinHelperOperation = "__string_tolower"
	HelperStringToString         BuiltinHelperOperation = "__string_tostring"
	HelperStringToUpper          BuiltinHelperOperation = "__string_toupper"
	HelperStringToXML            BuiltinHelperOperation = "__string_toxml"
	HelperStringTrim             BuiltinHelperOperation = "__string_trim"
	HelperStringTrimLeft         BuiltinHelperOperation = "__string_trimleft"
	HelperStringTrimRight        BuiltinHelperOperation = "__string_trimright"
)

// BuiltinHelperMember describes one intrinsic member. Signature constructs fresh
// semantic metadata because helper registrations may be specialized by callers.
// A method with no Signature uses receiver-dependent array analysis.
type BuiltinHelperMember struct {
	Name              string
	Operation         BuiltinHelperOperation
	Method            bool
	Signature         func() *FunctionType
	PropertyType      Type
	PropertyOperation BuiltinHelperOperation
}

// builtinHelperMembers is the sole catalog of names, aliases, property access,
// signatures, defaults and execution operations for builtin helpers.
var builtinHelperMembers = map[string][]BuiltinHelperMember{
	"array": {
		{Name: "Length", Operation: HelperArrayLength, Method: true, PropertyType: INTEGER},
		{Name: "High", Operation: HelperArrayHigh, Method: true, PropertyType: INTEGER},
		{Name: "Low", Operation: HelperArrayLow, Method: true, PropertyType: INTEGER},
		{Name: "Count", Operation: HelperArrayCount, Method: true, PropertyType: INTEGER},
		{Name: "Add", Operation: HelperArrayAdd, Method: true, Signature: func() *FunctionType { return NewProcedureType([]Type{nil}) }},
		{Name: "Delete", Operation: HelperArrayDelete, Method: true, Signature: func() *FunctionType {
			return NewFunctionTypeWithMetadata(
				[]Type{INTEGER, INTEGER},
				[]string{"index", "count"},
				[]interface{}{nil, int64(1)},
				[]bool{false, false},
				[]bool{false, false},
				[]bool{false, false},
				nil,
			)
		}},
		{Name: "IndexOf", Operation: HelperArrayIndexOf, Method: true, Signature: func() *FunctionType {
			return NewFunctionTypeWithMetadata(
				[]Type{nil, INTEGER},
				[]string{"value", "startIndex"},
				[]interface{}{nil, int64(0)},
				[]bool{false, false},
				[]bool{false, false},
				[]bool{false, false},
				INTEGER,
			)
		}},
		{Name: "SetLength", Operation: HelperArraySetLength, Method: true, Signature: func() *FunctionType { return NewProcedureType([]Type{INTEGER}) }},
		{Name: "Swap", Operation: HelperArraySwap, Method: true, Signature: func() *FunctionType { return NewProcedureType([]Type{INTEGER, INTEGER}) }},
		{Name: "Push", Operation: HelperArrayPush, Method: true, Signature: func() *FunctionType { return NewProcedureType([]Type{nil}) }},
		{Name: "Pop", Operation: HelperArrayPop, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{}, VARIANT) }},
		{Name: "Map", Operation: HelperArrayMap, Method: true, Signature: func() *FunctionType {
			return NewFunctionType([]Type{NewFunctionPointerType([]Type{VARIANT}, VARIANT)}, NewDynamicArrayType(VARIANT))
		}},
		{Name: "Join", Operation: HelperArrayJoin, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{STRING}, STRING) }},
		{Name: "Clear", Operation: HelperArrayClear, Method: true},
		{Name: "Contains", Operation: HelperArrayContains, Method: true},
		{Name: "Copy", Operation: HelperArrayCopy, Method: true},
		{Name: "Filter", Operation: HelperArrayFilter, Method: true},
		{Name: "ForEach", Operation: HelperArrayForEach, Method: true},
		{Name: "Insert", Operation: HelperArrayInsert, Method: true},
		{Name: "Move", Operation: HelperArrayMove, Method: true},
		{Name: "Peek", Operation: HelperArrayPeek, Method: true},
		{Name: "Remove", Operation: HelperArrayRemove, Method: true},
		{Name: "Reverse", Operation: HelperArrayReverse, Method: true},
		{Name: "Sort", Operation: HelperArraySort, Method: true},
	},
	"integer": {
		{Name: "ToString", Operation: HelperIntegerToString, Method: true, Signature: func() *FunctionType {
			return NewFunctionTypeWithMetadata(
				[]Type{INTEGER},
				[]string{"base"},
				[]interface{}{int64(10)},
				[]bool{false},
				[]bool{false},
				[]bool{false},
				STRING,
			)
		}, PropertyType: STRING},
		{Name: "ToHexString", Operation: HelperIntegerToHexString, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{INTEGER}, STRING) }},
	},
	"float": {
		{Name: "ToString", Operation: HelperFloatToStringPrec, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{INTEGER}, STRING) }, PropertyType: STRING, PropertyOperation: HelperFloatToStringDefault},
	},
	"boolean": {
		{Name: "ToString", Operation: HelperBooleanToString, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{}, STRING) }, PropertyType: STRING},
	},
	"string": {
		{Name: "Length", Operation: HelperStringLength, PropertyType: INTEGER},
		{Name: "ToUpper", Operation: HelperStringToUpper, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{}, STRING) }},
		{Name: "ToLower", Operation: HelperStringToLower, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{}, STRING) }},
		{Name: "PadLeft", Operation: HelperBuiltinPadLeft, Method: true, Signature: func() *FunctionType {
			return NewFunctionTypeWithMetadata(
				[]Type{INTEGER, STRING},
				[]string{"count", "char"},
				[]interface{}{nil, " "},
				[]bool{false, false},
				[]bool{false, false},
				[]bool{false, false},
				STRING,
			)
		}},
		{Name: "PadRight", Operation: HelperBuiltinPadRight, Method: true, Signature: func() *FunctionType {
			return NewFunctionTypeWithMetadata(
				[]Type{INTEGER, STRING},
				[]string{"count", "char"},
				[]interface{}{nil, " "},
				[]bool{false, false},
				[]bool{false, false},
				[]bool{false, false},
				STRING,
			)
		}},
		{Name: "DeleteLeft", Operation: HelperBuiltinStrDeleteLeft, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{INTEGER}, STRING) }},
		{Name: "DeleteRight", Operation: HelperBuiltinStrDeleteRight, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{INTEGER}, STRING) }},
		{Name: "Normalize", Operation: HelperBuiltinNormalizeString, Method: true, Signature: func() *FunctionType {
			return NewFunctionTypeWithMetadata(
				[]Type{STRING},
				[]string{"form"},
				[]interface{}{"NFC"},
				[]bool{false},
				[]bool{false},
				[]bool{false},
				STRING,
			)
		}},
		{Name: "StripAccents", Operation: HelperBuiltinStripAccents, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{}, STRING) }, PropertyType: STRING},
		{Name: "ToInteger", Operation: HelperStringToInteger, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{}, INTEGER) }},
		{Name: "ToFloat", Operation: HelperStringToFloat, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{}, FLOAT) }},
		{Name: "ToString", Operation: HelperStringToString, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{}, STRING) }},
		{Name: "StartsWith", Operation: HelperStringStartsWith, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{STRING}, BOOLEAN) }},
		{Name: "EndsWith", Operation: HelperStringEndsWith, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{STRING}, BOOLEAN) }},
		{Name: "Contains", Operation: HelperStringContains, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{STRING}, BOOLEAN) }},
		{Name: "IndexOf", Operation: HelperStringIndexOf, Method: true, Signature: func() *FunctionType {
			return NewFunctionTypeWithMetadata(
				[]Type{STRING, INTEGER},
				[]string{"substring", "startIndex"},
				[]interface{}{nil, int64(1)},
				[]bool{false, false},
				[]bool{false, false},
				[]bool{false, false},
				INTEGER,
			)
		}},
		{Name: "Matches", Operation: HelperStringMatches, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{STRING}, BOOLEAN) }},
		{Name: "IsASCII", Operation: HelperStringIsASCII, PropertyType: BOOLEAN},
		{Name: "Trim", Operation: HelperStringTrim, Method: true, Signature: func() *FunctionType {
			return NewFunctionTypeWithMetadata(
				[]Type{INTEGER, INTEGER},
				[]string{"left", "right"},
				[]interface{}{int64(0), int64(0)},
				[]bool{false, false},
				[]bool{false, false},
				[]bool{false, false},
				STRING,
			)
		}, PropertyType: STRING},
		{Name: "TrimLeft", Operation: HelperStringTrimLeft, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{INTEGER}, STRING) }, PropertyType: STRING},
		{Name: "TrimRight", Operation: HelperStringTrimRight, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{INTEGER}, STRING) }, PropertyType: STRING},
		{Name: "Copy", Operation: HelperStringCopy, Method: true, Signature: func() *FunctionType {
			return NewFunctionTypeWithMetadata(
				[]Type{INTEGER, INTEGER},
				[]string{"start", "length"},
				[]interface{}{nil, int64(2147483647)},
				[]bool{false, false},
				[]bool{false, false},
				[]bool{false, false},
				STRING,
			)
		}},
		{Name: "Before", Operation: HelperStringBefore, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{STRING}, STRING) }},
		{Name: "After", Operation: HelperStringAfter, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{STRING}, STRING) }},
		{Name: "Split", Operation: HelperStringSplit, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{STRING}, NewDynamicArrayType(STRING)) }},
		{Name: "ToJSON", Operation: HelperStringToJSON, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{}, STRING) }},
		{Name: "ToHTML", Operation: HelperStringToHTML, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{}, STRING) }},
		{Name: "ToHTMLAttribute", Operation: HelperStringToHTMLAttribute, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{}, STRING) }},
		{Name: "ToCSSText", Operation: HelperStringToCSSText, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{}, STRING) }},
		{Name: "ToXML", Operation: HelperStringToXML, Method: true, Signature: func() *FunctionType {
			return NewFunctionTypeWithMetadata(
				[]Type{INTEGER},
				[]string{"mode"},
				[]interface{}{int64(0)},
				[]bool{false},
				[]bool{false},
				[]bool{false},
				STRING,
			)
		}},
		{Name: "UpperCase", Operation: HelperStringToUpper, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{}, STRING) }},
		{Name: "LowerCase", Operation: HelperStringToLower, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{}, STRING) }},
	},
	"array of string": {
		{Name: "Join", Operation: HelperStringArrayJoin, Method: true, Signature: func() *FunctionType { return NewFunctionType([]Type{STRING}, STRING) }},
	},
	"enum": {
		{Name: "Value", Operation: HelperEnumValue, PropertyType: INTEGER},
		{Name: "Name", Operation: HelperEnumName, PropertyType: STRING},
		{Name: "QualifiedName", Operation: HelperEnumQualifiedName, PropertyType: STRING},
	},
}

// LookupBuiltinHelper resolves a member with DWScript's case-insensitive rules.
func LookupBuiltinHelper(target, name string) (BuiltinHelperMember, bool) {
	for _, member := range builtinHelperMembers[ident.Normalize(target)] {
		if ident.Equal(member.Name, name) {
			return member, true
		}
	}
	return BuiltinHelperMember{}, false
}

// NewBuiltinHelper returns an independent registration populated from the shared
// catalog. Generic array and enum registrations have no concrete target type.
func NewBuiltinHelper(target string) *HelperType {
	target = ident.Normalize(target)
	var helper *HelperType
	switch target {
	case "array":
		helper = NewHelperType("TArrayHelper", nil, false)
	case "integer":
		helper = NewHelperType("__TIntegerIntrinsicHelper", INTEGER, false)
	case "float":
		helper = NewHelperType("__TFloatIntrinsicHelper", FLOAT, false)
	case "boolean":
		helper = NewHelperType("__TBooleanIntrinsicHelper", BOOLEAN, false)
	case "string":
		helper = NewHelperType("__TStringIntrinsicHelper", STRING, false)
	case "array of string":
		helper = NewHelperType("__TStringDynArrayIntrinsicHelper", NewDynamicArrayType(STRING), true)
	case "enum":
		helper = NewHelperType("__TEnumIntrinsicHelper", nil, false)
	default:
		return nil
	}
	for _, member := range builtinHelperMembers[target] {
		key := ident.Normalize(member.Name)
		if member.Method {
			helper.BuiltinMethods[key] = string(member.Operation)
			if member.Signature != nil {
				helper.Methods[key] = member.Signature()
			}
		}
		if member.PropertyType != nil {
			operation := member.PropertyOperation
			if operation == "" {
				operation = member.Operation
			}
			helper.Properties[key] = &PropertyInfo{Name: member.Name, Type: member.PropertyType,
				ReadKind: PropAccessBuiltin, ReadSpec: string(operation), WriteKind: PropAccessNone}
		}
	}
	return helper
}
