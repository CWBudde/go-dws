package runtime

import "github.com/cwbudde/go-dws/internal/types"

// LanguageType returns the resolved language type carried by a runtime value.
// Execution-only containers without a language type return nil.
func LanguageType(value Value) types.Type {
	if value == nil {
		return types.NIL
	}
	if scalar := scalarLanguageType(value); scalar != nil {
		return scalar
	}
	switch v := value.(type) {
	case *ByteBufferValue:
		// The built-in buffer type is a singleton, so it carries no per-value
		// type information the way objects and records do.
		return types.BYTE_BUFFER
	case *ObjectInstance:
		if v.Class != nil {
			if classType := v.Class.GetClassType(); classType != nil {
				return classType
			}
		}
	case *EnumValue:
		if v.EnumType != nil {
			return v.EnumType
		}
		return &types.EnumType{Name: v.TypeName}
	case *InterfaceInstance:
		if iface, ok := v.Interface.(interface{ GetInterfaceType() *types.InterfaceType }); ok {
			if interfaceType := iface.GetInterfaceType(); interfaceType != nil {
				return interfaceType
			}
		}
	}
	return aggregateLanguageType(value)
}

func aggregateLanguageType(value Value) types.Type {
	switch v := value.(type) {
	case *ArrayValue:
		if v.ArrayType != nil {
			return v.ArrayType
		}
	case *RecordValue:
		if v.RecordType != nil {
			return v.RecordType
		}
	case *SetValue:
		if v.SetType != nil {
			return v.SetType
		}
	case *SubrangeValue:
		if v.SubrangeType != nil {
			return v.SubrangeType
		}
	case *AssociativeArrayValue:
		if v.AssocType != nil {
			return v.AssocType
		}
	case *FunctionPointerValue:
		if v.PointerType != nil {
			return v.PointerType
		}
	}
	return nil
}

func scalarLanguageType(value Value) types.Type {
	switch value.(type) {
	case *IntegerValue:
		return types.INTEGER
	case *FloatValue:
		return types.FLOAT
	case *StringValue:
		return types.STRING
	case *BooleanValue:
		return types.BOOLEAN
	case *NilValue, *NullValue, *UnassignedValue:
		return types.NIL
	case *VariantValue:
		return types.VARIANT
	case *JSONValue:
		return types.JSON_VARIANT
	}
	return nil
}
