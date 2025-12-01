package evaluator

import (
	"fmt"
	"math"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// ============================================================================
// String Formatting Methods
// ============================================================================
//
// This file implements the string formatting method of the builtins.Context
// interface for the Evaluator:
// - FormatString(): Format strings using fmt.Sprintf semantics
//
// Supports format verbs: %s, %d, %f, %v, %x, %X, %o, %%
// Supports width/precision modifiers: %5d, %.2f, %8.2f
//
// Phase 3.5.143 - Phase IV: Complex Methods
// ============================================================================

// FormatString formats a string using Go fmt.Sprintf semantics with DWScript values.
// This implements the builtins.Context interface.
// Task 3.7.8: Helper for Format() function.
func (e *Evaluator) FormatString(format string, args []Value) (string, error) {
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
		unwrapped := e.UnwrapVariant(elem)

		switch v := unwrapped.(type) {
		case *runtime.IntegerValue:
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
		case *runtime.FloatValue:
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
		case *runtime.StringValue:
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
		case *runtime.BooleanValue:
			goArgs[idx] = v.Value
		default:
			return "", fmt.Errorf("cannot format value of type %s at index %d", unwrapped.Type(), idx)
		}
	}

	// Format the string
	result := fmt.Sprintf(format, goArgs...)

	return result, nil
}
