package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ident"
)

// Note: This file uses ident.Normalize() for type name normalization (case-insensitive lookups)
// and ident.Equal() for case-insensitive string comparisons

// ============================================================================
// Helper Discovery and Lookup
// ============================================================================

// getHelpersForValue returns all helpers that apply to the given value's type
func (i *Interpreter) getHelpersForValue(val Value) []*HelperInfo {
	// Get the type name from the value
	var typeName string
	switch v := val.(type) {
	case *IntegerValue:
		typeName = "Integer"
	case *FloatValue:
		typeName = "Float"
	case *StringValue:
		typeName = "String"
	case *BooleanValue:
		typeName = "Boolean"
	case *ObjectInstance:
		typeName = v.Class.GetName()
	case *RecordValue:
		typeName = v.RecordType.Name
	case *ArrayValue:
		// First try specific array type (e.g., "array of String"), then generic array helpers
		specific := ident.Normalize(v.ArrayType.String())
		var combined []*HelperInfo
		if h := i.typeSystem.LookupHelpers(specific); len(h) > 0 {
			for _, helper := range h {
				if hi, ok := helper.(*HelperInfo); ok {
					combined = append(combined, hi)
				}
			}
		}
		// If it's a static array, also try the dynamic equivalent ("array of <elem>")
		if v.ArrayType.IsStatic() && v.ArrayType.ElementType != nil {
			dynKey := ident.Normalize(fmt.Sprintf("array of %s", v.ArrayType.ElementType.String()))
			if h := i.typeSystem.LookupHelpers(dynKey); len(h) > 0 {
				for _, helper := range h {
					if hi, ok := helper.(*HelperInfo); ok {
						combined = append(combined, hi)
					}
				}
			}
		}
		if h := i.typeSystem.LookupHelpers("array"); len(h) > 0 {
			for _, helper := range h {
				if hi, ok := helper.(*HelperInfo); ok {
					combined = append(combined, hi)
				}
			}
		}
		return combined
	case *EnumValue:
		// First try specific enum type (e.g., "TColor"), then generic enum helpers
		specific := ident.Normalize(v.TypeName)
		var combined []*HelperInfo
		if h := i.typeSystem.LookupHelpers(specific); len(h) > 0 {
			for _, helper := range h {
				if hi, ok := helper.(*HelperInfo); ok {
					combined = append(combined, hi)
				}
			}
		}
		if h := i.typeSystem.LookupHelpers("enum"); len(h) > 0 {
			for _, helper := range h {
				if hi, ok := helper.(*HelperInfo); ok {
					combined = append(combined, hi)
				}
			}
		}
		return combined
	case *TypeMetaValue:
		if v.TypeName != "" {
			typeName = v.TypeName
		} else if v.TypeInfo != nil {
			typeName = v.TypeInfo.String()
		}
	default:
		// For other types, try to extract type name from Type() method
		typeName = v.Type()
	}

	// Look up helpers for this type
	helpers := i.typeSystem.LookupHelpers(ident.Normalize(typeName))
	result := make([]*HelperInfo, 0, len(helpers))
	for _, h := range helpers {
		if hi, ok := h.(*HelperInfo); ok {
			result = append(result, hi)
		}
	}
	return result
}
