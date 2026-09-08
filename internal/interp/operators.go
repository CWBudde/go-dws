package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"

	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// runtimeOperatorEntry is the runtime representation of a class operator binding.
type runtimeOperatorEntry = runtime.ClassOperatorEntry

// NormalizeTypeAnnotation normalizes a type annotation string for operator lookup.
// Primitive types (integer, float, string, boolean, variant, nil) and array types
// are returned normalized. All other types get a "class:" prefix.
// This is a convenience wrapper around interptypes.NormalizeTypeAnnotation.
func NormalizeTypeAnnotation(name string) string {
	return interptypes.NormalizeTypeAnnotation(name)
}

func valueTypeKey(val Value) string {
	if val == nil {
		return "nil"
	}
	switch v := val.(type) {
	case *ObjectInstance:
		if v.Class != nil {
			return "class:" + ident.Normalize(v.Class.GetName())
		}
		return "class:"
	case *RecordValue:
		if v.RecordType != nil && v.RecordType.Name != "" {
			return "class:" + ident.Normalize(v.RecordType.Name)
		}
		return "record"
	case *ArrayValue:
		// Include array element type for operator overload matching
		if v.ArrayType != nil && v.ArrayType.ElementType != nil {
			elemTypeStr := v.ArrayType.ElementType.String()
			return "array of " + ident.Normalize(elemTypeStr)
		}
		return "array"
	default:
		return ident.Normalize(val.Type())
	}
}
