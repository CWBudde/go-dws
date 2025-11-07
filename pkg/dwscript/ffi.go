package dwscript

import (
	"fmt"
	"reflect"

	"github.com/cwbudde/go-dws/internal/interp"
)

// FunctionSignature describes the type signature of an external function.
// It specifies the parameter types and return type for validation and marshaling.
type FunctionSignature struct {
	Name       string   // Function name as it appears in DWScript
	ParamTypes []string // Parameter type names (e.g., "Integer", "String", "Boolean")
	ReturnType string   // Return type name (e.g., "Integer", "Void")
	IsVariadic bool     // True if the function accepts a variable number of arguments
}

// String returns a human-readable representation of the signature.
func (sig *FunctionSignature) String() string {
	if len(sig.ParamTypes) == 0 {
		return fmt.Sprintf("%s(): %s", sig.Name, sig.ReturnType)
	}

	// Format params, marking last param as variadic if applicable
	paramsStr := fmt.Sprintf("%v", sig.ParamTypes)
	if sig.IsVariadic && len(sig.ParamTypes) > 0 {
		// Show the last param as "...Type" instead of "array of Type"
		lastParam := sig.ParamTypes[len(sig.ParamTypes)-1]
		modifiedParams := make([]string, len(sig.ParamTypes))
		copy(modifiedParams, sig.ParamTypes)
		modifiedParams[len(modifiedParams)-1] = "..." + lastParam
		paramsStr = fmt.Sprintf("%v", modifiedParams)
	}

	return fmt.Sprintf("%s(%s): %s", sig.Name, paramsStr, sig.ReturnType)
}

// ExternalFunction represents a Go function that can be called from DWScript.
// Implementations wrap Go functions with appropriate marshaling and error handling.
type ExternalFunction interface {
	// Call invokes the function with DWScript values and returns a DWScript value.
	Call(args []interp.Value) (interp.Value, error)

	// Signature returns the function's type signature for validation.
	Signature() *FunctionSignature
}

// RegisterFunction registers a Go function to be callable from DWScript.
// The function's signature is automatically detected via reflection.
//
// Calling Convention:
// Go functions are called with type-safe marshaling. Arguments are automatically
// converted from DWScript values to the appropriate Go types based on the
// function signature. This is safer and more idiomatic than using []any.
//
// The wrapper:
//  1. Validates argument count at call time
//  2. Marshals each DWScript Value to the expected Go type
//  3. Calls the Go function with native types
//  4. Marshals the return value(s) back to DWScript Values
//  5. Converts errors to EHost exceptions automatically
//
// Type Mapping (Go ↔ DWScript):
//
//	Primitives:
//	- int, int32, int64, int16, int8 ↔ Integer
//	- float64, float32 ↔ Float
//	- string ↔ String
//	- bool ↔ Boolean
//
//	Collections:
//	- []T ↔ array of T (dynamic arrays, also supports variadic-like behavior)
//	- map[string]T ↔ record-like structure (associative array)
//
//	Error Handling:
//	- error ↔ EHost exception (Go errors are raised as DWScript exceptions)
//	- Go panics are also caught and converted to EHost exceptions
//
// Function Signatures:
//   - func(params...) T              - returns value
//   - func(params...) (T, error)     - returns value, error becomes exception
//   - func(params...) error          - procedure with error handling
//   - func(params...)                - void procedure
//   - func(params, ...T)             - variadic function (Go's ...T syntax supported)
//
// Example:
//
//	engine.RegisterFunction("Add", func(a, b int64) int64 {
//	    return a + b
//	})
//
//	engine.RegisterFunction("GetScores", func() []int64 {
//	    return []int64{95, 87, 92}
//	})
//
// DWScript code can then call:
//
//	var sum := Add(40, 2);
//	var scores := GetScores();
func (e *Engine) RegisterFunction(name string, fn any) error {
	if fn == nil {
		return fmt.Errorf("cannot register nil function")
	}

	// Use reflection to analyze the function
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	// Validate it's actually a function
	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("expected function, got %s", fnType.Kind())
	}

	// Detect the signature
	sig, err := detectSignature(name, fnType)
	if err != nil {
		return fmt.Errorf("invalid function signature for %s: %w", name, err)
	}

	// Create a wrapper that handles marshaling
	wrapper := &externalFunctionWrapper{
		name:      name,
		goFunc:    fnValue,
		signature: sig,
	}

	// Register with the engine's registry
	if e.externalFunctions == nil {
		e.externalFunctions = interp.NewExternalFunctionRegistry()
	}

	return e.externalFunctions.Register(name, wrapper)
}

// externalFunctionWrapper wraps a Go function with marshaling logic.
type externalFunctionWrapper struct {
	name      string
	goFunc    reflect.Value
	signature *FunctionSignature
}

// Call implements ExternalFunction.Call
func (w *externalFunctionWrapper) Call(args []interp.Value) (interp.Value, error) {
	fnType := w.goFunc.Type()
	numParams := fnType.NumIn()

	// Validate argument count
	if w.signature.IsVariadic {
		// For variadic functions, we need at least (numParams - 1) arguments
		// The last parameter is the variadic slice, which can be empty
		minArgs := numParams - 1
		if len(args) < minArgs {
			return nil, fmt.Errorf("function %s expects at least %d arguments, got %d",
				w.name, minArgs, len(args))
		}
	} else {
		// Non-variadic functions must have exact argument count
		if len(args) != numParams {
			return nil, fmt.Errorf("function %s expects %d arguments, got %d",
				w.name, numParams, len(args))
		}
	}

	// Marshal DWScript values to Go values
	var goArgs []reflect.Value

	if w.signature.IsVariadic {
		// Handle variadic function: pack extra arguments into a slice
		numRequiredParams := numParams - 1
		goArgs = make([]reflect.Value, numParams)

		// Marshal required (non-variadic) parameters
		for i := 0; i < numRequiredParams; i++ {
			goArg, err := interp.MarshalToGo(args[i], fnType.In(i))
			if err != nil {
				return nil, fmt.Errorf("argument %d: %w", i, err)
			}
			goArgs[i] = reflect.ValueOf(goArg)
		}

		// Pack variadic arguments into a slice
		variadicType := fnType.In(numRequiredParams).Elem() // Get element type of slice
		numVariadicArgs := len(args) - numRequiredParams
		variadicSlice := reflect.MakeSlice(fnType.In(numRequiredParams), numVariadicArgs, numVariadicArgs)

		for i := 0; i < numVariadicArgs; i++ {
			argIdx := numRequiredParams + i
			goArg, err := interp.MarshalToGo(args[argIdx], variadicType)
			if err != nil {
				return nil, fmt.Errorf("variadic argument %d: %w", i, err)
			}
			variadicSlice.Index(i).Set(reflect.ValueOf(goArg))
		}

		goArgs[numRequiredParams] = variadicSlice
	} else {
		// Non-variadic function: marshal arguments normally
		goArgs = make([]reflect.Value, numParams)
		for i := 0; i < numParams; i++ {
			goArg, err := interp.MarshalToGo(args[i], fnType.In(i))
			if err != nil {
				return nil, fmt.Errorf("argument %d: %w", i, err)
			}
			goArgs[i] = reflect.ValueOf(goArg)
		}
	}

	// Call the Go function
	var results []reflect.Value
	if w.signature.IsVariadic {
		// Use CallSlice for variadic functions - this unpacks the last argument
		results = w.goFunc.CallSlice(goArgs)
	} else {
		results = w.goFunc.Call(goArgs)
	}

	// Handle return values
	return handleReturnValues(results)
}

// Signature implements ExternalFunction.Signature
func (w *externalFunctionWrapper) Signature() *FunctionSignature {
	return w.signature
}

// detectSignature analyzes a Go function's type and creates a FunctionSignature.
func detectSignature(name string, fnType reflect.Type) (*FunctionSignature, error) {
	sig := &FunctionSignature{
		Name:       name,
		ParamTypes: make([]string, 0, fnType.NumIn()),
		ReturnType: "Void",
		IsVariadic: fnType.IsVariadic(),
	}

	// Analyze parameters
	for i := 0; i < fnType.NumIn(); i++ {
		paramType := fnType.In(i)
		dwsType, err := goTypeToDWS(paramType)
		if err != nil {
			return nil, fmt.Errorf("parameter %d: %w", i, err)
		}
		sig.ParamTypes = append(sig.ParamTypes, dwsType)
	}

	// Analyze return type(s)
	numOut := fnType.NumOut()
	if numOut > 2 {
		return nil, fmt.Errorf("functions can return at most 2 values (result, error)")
	}

	if numOut >= 1 {
		// Check if last return is error
		lastType := fnType.Out(numOut - 1)
		isError := lastType.Implements(reflect.TypeOf((*error)(nil)).Elem())

		if numOut == 1 {
			if isError {
				// func() error -> Void (error raised as exception)
				sig.ReturnType = "Void"
			} else {
				// func() T -> T
				dwsType, err := goTypeToDWS(lastType)
				if err != nil {
					return nil, fmt.Errorf("return type: %w", err)
				}
				sig.ReturnType = dwsType
			}
		} else if numOut == 2 {
			if !isError {
				return nil, fmt.Errorf("second return value must be error type")
			}
			// func() (T, error) -> T
			dwsType, err := goTypeToDWS(fnType.Out(0))
			if err != nil {
				return nil, fmt.Errorf("return type: %w", err)
			}
			sig.ReturnType = dwsType
		}
	}

	return sig, nil
}

// goTypeToDWS maps Go types to DWScript type names.
func goTypeToDWS(goType reflect.Type) (string, error) {
	switch goType.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return "Integer", nil
	case reflect.Float64, reflect.Float32:
		return "Float", nil
	case reflect.String:
		return "String", nil
	case reflect.Bool:
		return "Boolean", nil
	case reflect.Slice:
		// []T -> "array of T"
		elemType, err := goTypeToDWS(goType.Elem())
		if err != nil {
			return "", fmt.Errorf("slice element: %w", err)
		}
		return "array of " + elemType, nil
	case reflect.Map:
		// map[string]T -> "record" (associative array)
		if goType.Key().Kind() != reflect.String {
			return "", fmt.Errorf("only map[string]T is supported, got %s", goType)
		}
		// Could include value type info, but "record" is generic enough
		return "record", nil
	default:
		return "", fmt.Errorf("unsupported Go type: %s", goType)
	}
}

// handleReturnValues processes the results from a Go function call and converts to DWScript values.
func handleReturnValues(results []reflect.Value) (interp.Value, error) {
	if len(results) == 0 {
		// No return value -> Void
		return interp.NewNilValue(), nil
	}

	if len(results) == 1 {
		// Check if it's an error
		if results[0].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if results[0].IsNil() {
				return interp.NewNilValue(), nil
			}
			return nil, results[0].Interface().(error)
		}
		// Single non-error return value
		return interp.MarshalToDWS(results[0].Interface())
	}

	// Two return values: (result, error)
	if len(results) == 2 {
		// Check error first
		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}
		// Return the result value
		return interp.MarshalToDWS(results[0].Interface())
	}

	return nil, fmt.Errorf("unexpected number of return values: %d", len(results))
}
