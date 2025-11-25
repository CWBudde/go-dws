package evaluator

import (
	"reflect"

	"github.com/cwbudde/go-dws/internal/jsonvalue"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// indexJSON performs JSON value indexing without importing the parent interp package.
// Supports both object property access (string index) and array element access (integer index).
//
// Task 3.5.99b: Migrated from interpreter.indexJSON to avoid EvalNode delegation.
// Uses reflection to access the internal jsonvalue.Value field to avoid circular imports.
//
// For JSON objects: obj['propertyName'] returns the value or nil if not found
// For JSON arrays: arr[index] returns the element or nil if out of bounds
func (e *Evaluator) indexJSON(base Value, index Value, node ast.Node) Value {
	// Extract the underlying jsonvalue.Value using reflection
	// This avoids importing the parent interp package
	jv := extractJSONValueViaReflection(base)

	// If we couldn't extract a JSON value, it's not a JSON type
	if jv == nil && base.Type() != "JSON" {
		return e.newError(node, "cannot index non-JSON value of type %s", base.Type())
	}

	// Handle nil/null JSON value
	if jv == nil {
		// nil/null JSON value - delegate to adapter for proper Variant wrapping
		return e.adapter.WrapJSONValueInVariant(nil)
	}

	kind := jv.Kind()

	// JSON Object: support string indexing
	if kind == jsonvalue.KindObject {
		// Index must be a string for object property access
		// Check via Type() method to avoid importing interp package
		if index.Type() != "STRING" {
			return e.newError(node, "JSON object index must be a string, got %s", index.Type())
		}

		// Extract string value via String() method
		indexStr := index.String()

		// Get the property value (returns nil if not found)
		propValue := jv.ObjectGet(indexStr)

		// Delegate to adapter to wrap in Variant
		return e.adapter.WrapJSONValueInVariant(propValue)
	}

	// JSON Array: support integer indexing
	if kind == jsonvalue.KindArray {
		// Extract integer index
		indexInt, ok := ExtractIntegerIndex(index)
		if !ok {
			return e.newError(node, "JSON array index must be an integer, got %s", index.Type())
		}

		// Get the array element (returns nil if out of bounds)
		elemValue := jv.ArrayGet(indexInt)

		// Delegate to adapter to wrap in Variant
		return e.adapter.WrapJSONValueInVariant(elemValue)
	}

	// Not an object or array
	return e.newError(node, "cannot index JSON %s", kind.String())
}

// extractJSONValueViaReflection uses reflection to extract the internal jsonvalue.Value
// from a JSONValue struct, avoiding the need to import the parent interp package.
// Task 3.5.99b: Reflection-based extraction to avoid circular imports.
func extractJSONValueViaReflection(val Value) *jsonvalue.Value {
	if val == nil {
		return nil
	}

	// Check if this is a "JSON" type
	if val.Type() != "JSON" {
		return nil
	}

	// Use reflection to access the Value field
	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil
	}

	// Look for a field named "Value" of type *jsonvalue.Value
	valueField := rv.FieldByName("Value")
	if !valueField.IsValid() {
		return nil
	}

	// Try to convert to *jsonvalue.Value
	if jv, ok := valueField.Interface().(*jsonvalue.Value); ok {
		return jv
	}

	return nil
}
