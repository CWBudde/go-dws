package runtime

import (
	"reflect"

	"github.com/cwbudde/go-dws/internal/types"
)

// SameValueType compares representations while preserving named record and subrange identity.
func SameValueType(left, right Value) bool {
	leftKind, rightKind := KindOf(left), KindOf(right)
	if leftKind != rightKind {
		return false
	}
	switch leftKind {
	case KindRecord, KindSubrange:
		leftType, rightType := LanguageType(left), LanguageType(right)
		if leftType == nil || rightType == nil {
			return leftType == nil && rightType == nil
		}
		return types.OperatorTypesEqual(leftType, rightType)
	case KindUnknown:
		return reflect.TypeOf(left) == reflect.TypeOf(right)
	default:
		return true
	}
}

// ValueKind identifies a runtime representation, independently of its language type.
// Named records and subranges retain their declaration identity in LanguageType.
type ValueKind uint8

const (
	// KindUnknown identifies the UNKNOWN representation.
	KindUnknown ValueKind = 0
	// KindArray identifies the ARRAY representation.
	KindArray ValueKind = 1
	// KindAssociativeArray identifies the ASSOCIATIVE_ARRAY representation.
	KindAssociativeArray ValueKind = 2
	// KindBoolean identifies the BOOLEAN representation.
	KindBoolean ValueKind = 3
	// KindClass identifies the CLASS representation.
	KindClass ValueKind = 4
	// KindClassInfo identifies the CLASSINFO representation.
	KindClassInfo ValueKind = 5
	// KindEnum identifies the ENUM representation.
	KindEnum ValueKind = 6
	// KindEnumType identifies the ENUM_TYPE representation.
	KindEnumType ValueKind = 7
	// KindError identifies the ERROR representation.
	KindError ValueKind = 8
	// KindException identifies the EXCEPTION representation.
	KindException ValueKind = 9
	// KindExternalFunction identifies the EXTERNAL_FUNCTION representation.
	KindExternalFunction ValueKind = 10
	// KindExternalVar identifies the EXTERNAL_VAR representation.
	KindExternalVar ValueKind = 11
	// KindFloat identifies the FLOAT representation.
	KindFloat ValueKind = 12
	// KindFunctionPointer identifies the FUNCTION_POINTER representation.
	KindFunctionPointer ValueKind = 13
	// KindInteger identifies the INTEGER representation.
	KindInteger ValueKind = 14
	// KindInterface identifies the INTERFACE representation.
	KindInterface ValueKind = 15
	// KindJSON identifies the JSON representation.
	KindJSON ValueKind = 16
	// KindLambda identifies the LAMBDA representation.
	KindLambda ValueKind = 17
	// KindLazyThunk identifies the LAZY_THUNK representation.
	KindLazyThunk ValueKind = 18
	// KindLocalFunctionSet identifies the LOCAL_FUNCTION_SET representation.
	KindLocalFunctionSet ValueKind = 19
	// KindMethodPointer identifies the METHOD_POINTER representation.
	KindMethodPointer ValueKind = 20
	// KindNil identifies the NIL representation.
	KindNil ValueKind = 21
	// KindNull identifies the NULL representation.
	KindNull ValueKind = 22
	// KindObject identifies the OBJECT representation.
	KindObject ValueKind = 23
	// KindRecord identifies the RECORD representation.
	KindRecord ValueKind = 24
	// KindRecordType identifies the RECORD_TYPE representation.
	KindRecordType ValueKind = 25
	// KindReference identifies the REFERENCE representation.
	KindReference ValueKind = 26
	// KindRTTITypeInfo identifies the RTTI_TYPEINFO representation.
	KindRTTITypeInfo ValueKind = 27
	// KindSet identifies the SET representation.
	KindSet ValueKind = 28
	// KindSetType identifies the SET_TYPE representation.
	KindSetType ValueKind = 29
	// KindString identifies the STRING representation.
	KindString ValueKind = 30
	// KindSubrange identifies the SUBRANGE representation.
	KindSubrange ValueKind = 31
	// KindTypeAlias identifies the TYPE_ALIAS representation.
	KindTypeAlias ValueKind = 32
	// KindTypeCast identifies the TYPE_CAST representation.
	KindTypeCast ValueKind = 33
	// KindTypeMeta identifies the TYPE_META representation.
	KindTypeMeta ValueKind = 34
	// KindUnassigned identifies the UNASSIGNED representation.
	KindUnassigned ValueKind = 35
	// KindVariant identifies the VARIANT representation.
	KindVariant ValueKind = 36
	// KindByteBuffer identifies the BYTEBUFFER representation.
	KindByteBuffer ValueKind = 37
)

// KindOf identifies built-in runtime representations without reading display strings.
// Values outside the runtime may opt in with a ValueKind method.
func KindOf(value Value) ValueKind {
	if value == nil {
		return KindNil
	}
	if typed, ok := value.(interface{ ValueKind() ValueKind }); ok {
		return typed.ValueKind()
	}
	return KindUnknown
}

var valueKindNames = [...]string{
	KindArray:            "ARRAY",
	KindAssociativeArray: "ASSOCIATIVE_ARRAY",
	KindBoolean:          "BOOLEAN",
	KindByteBuffer:       "BYTEBUFFER",
	KindClass:            "CLASS",
	KindClassInfo:        "CLASSINFO",
	KindEnum:             "ENUM",
	KindEnumType:         "ENUM_TYPE",
	KindError:            "ERROR",
	KindException:        "EXCEPTION",
	KindExternalFunction: "EXTERNAL_FUNCTION",
	KindExternalVar:      "EXTERNAL_VAR",
	KindFloat:            "FLOAT",
	KindFunctionPointer:  "FUNCTION_POINTER",
	KindInteger:          "INTEGER",
	KindInterface:        "INTERFACE",
	KindJSON:             "JSON",
	KindLambda:           "LAMBDA",
	KindLazyThunk:        "LAZY_THUNK",
	KindLocalFunctionSet: "LOCAL_FUNCTION_SET",
	KindMethodPointer:    "METHOD_POINTER",
	KindNil:              "NIL",
	KindNull:             "NULL",
	KindObject:           "OBJECT",
	KindRecord:           "RECORD",
	KindRecordType:       "RECORD_TYPE",
	KindReference:        "REFERENCE",
	KindRTTITypeInfo:     "RTTI_TYPEINFO",
	KindSet:              "SET",
	KindSetType:          "SET_TYPE",
	KindString:           "STRING",
	KindSubrange:         "SUBRANGE",
	KindTypeAlias:        "TYPE_ALIAS",
	KindTypeCast:         "TYPE_CAST",
	KindTypeMeta:         "TYPE_META",
	KindUnassigned:       "UNASSIGNED",
	KindUnknown:          "UNKNOWN",
	KindVariant:          "VARIANT",
}

// String preserves legacy representation labels for diagnostics.
func (kind ValueKind) String() string {
	if int(kind) < len(valueKindNames) {
		return valueKindNames[kind]
	}
	return "UNKNOWN"
}

// ValueKind identifies the runtime representation.
func (*IntegerValue) ValueKind() ValueKind { return KindInteger }

// ValueKind identifies the runtime representation.
func (*FloatValue) ValueKind() ValueKind { return KindFloat }

// ValueKind identifies the runtime representation.
func (*StringValue) ValueKind() ValueKind { return KindString }

// ValueKind identifies the runtime representation.
func (*BooleanValue) ValueKind() ValueKind { return KindBoolean }

// ValueKind identifies the runtime representation.
func (*NilValue) ValueKind() ValueKind { return KindNil }

// ValueKind identifies the runtime representation.
func (*NullValue) ValueKind() ValueKind { return KindNull }

// ValueKind identifies the runtime representation.
func (*UnassignedValue) ValueKind() ValueKind { return KindUnassigned }

// ValueKind identifies the runtime representation.
func (*VariantValue) ValueKind() ValueKind { return KindVariant }

// ValueKind identifies the runtime representation.
func (*JSONValue) ValueKind() ValueKind { return KindJSON }

// ValueKind identifies the runtime representation.
func (*ObjectInstance) ValueKind() ValueKind { return KindObject }

// ValueKind identifies the runtime representation.
func (*ClassValue) ValueKind() ValueKind { return KindClass }

// ValueKind identifies the runtime representation.
func (*classTypeProxy) ValueKind() ValueKind { return KindClass }

// ValueKind identifies the runtime representation.
func (*ClassInfoValue) ValueKind() ValueKind { return KindClassInfo }

// ValueKind identifies the runtime representation.
func (*ArrayValue) ValueKind() ValueKind { return KindArray }

// ValueKind identifies the runtime representation.
func (*RecordValue) ValueKind() ValueKind { return KindRecord }

// ValueKind identifies the runtime representation.
func (*RecordTypeValue) ValueKind() ValueKind { return KindRecordType }

// ValueKind identifies the runtime representation.
func (*EnumValue) ValueKind() ValueKind { return KindEnum }

// ValueKind identifies the runtime representation.
func (*EnumTypeValue) ValueKind() ValueKind { return KindEnumType }

// ValueKind identifies the runtime representation.
func (*SetValue) ValueKind() ValueKind { return KindSet }

// ValueKind identifies the runtime representation.
func (*SetTypeValue) ValueKind() ValueKind { return KindSetType }

// ValueKind identifies the runtime representation.
func (*SubrangeValue) ValueKind() ValueKind { return KindSubrange }

// ValueKind identifies the runtime representation.
func (*AssociativeArrayValue) ValueKind() ValueKind { return KindAssociativeArray }

// ValueKind identifies the runtime representation.
func (*ErrorValue) ValueKind() ValueKind { return KindError }

// ValueKind identifies the runtime representation.
func (*ExceptionValue) ValueKind() ValueKind { return KindException }

// ValueKind identifies the runtime representation.
func (*TypeMetaValue) ValueKind() ValueKind { return KindTypeMeta }

// ValueKind identifies the runtime representation.
func (*TypeAliasValue) ValueKind() ValueKind { return KindTypeAlias }

// ValueKind identifies the runtime representation.
func (*ExternalVarValue) ValueKind() ValueKind { return KindExternalVar }

// ValueKind identifies the runtime representation.
func (*LazyThunk) ValueKind() ValueKind { return KindLazyThunk }

// ValueKind identifies the runtime representation.
func (*ReferenceValue) ValueKind() ValueKind { return KindReference }

// ValueKind identifies the runtime representation.
func (*InterfaceInstance) ValueKind() ValueKind { return KindInterface }

// ValueKind distinguishes unbound, bound, and lambda function pointers.
func (v *FunctionPointerValue) ValueKind() ValueKind {
	if v.SelfObject != nil {
		return KindMethodPointer
	}
	if v.Lambda != nil {
		return KindLambda
	}
	return KindFunctionPointer
}
