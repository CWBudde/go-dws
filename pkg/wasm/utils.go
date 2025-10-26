//go:build js && wasm

package wasm

import (
	"fmt"
	"syscall/js"
)

// WrapError wraps a Go error as a JavaScript Error object.
func WrapError(err error) js.Value {
	if err == nil {
		return js.Null()
	}

	errorObj := js.Global().Get("Error").New(err.Error())
	return errorObj
}

// WrapValue wraps a Go value as a JavaScript value.
func WrapValue(v interface{}) js.Value {
	switch val := v.(type) {
	case string:
		return js.ValueOf(val)
	case int:
		return js.ValueOf(val)
	case int64:
		return js.ValueOf(int(val))
	case float64:
		return js.ValueOf(val)
	case bool:
		return js.ValueOf(val)
	case nil:
		return js.Null()
	default:
		return js.ValueOf(fmt.Sprint(val))
	}
}

// UnwrapValue converts a JavaScript value to a Go value.
func UnwrapValue(v js.Value) interface{} {
	switch v.Type() {
	case js.TypeString:
		return v.String()
	case js.TypeNumber:
		return v.Float()
	case js.TypeBoolean:
		return v.Bool()
	case js.TypeNull, js.TypeUndefined:
		return nil
	default:
		return v.String()
	}
}

// CreateObject creates a new JavaScript object.
func CreateObject() js.Value {
	return js.Global().Get("Object").New()
}

// CreateArray creates a new JavaScript array.
func CreateArray(length int) js.Value {
	return js.Global().Get("Array").New(length)
}

// SetProperty sets a property on a JavaScript object.
func SetProperty(obj js.Value, key string, value interface{}) {
	obj.Set(key, WrapValue(value))
}

// GetProperty gets a property from a JavaScript object.
func GetProperty(obj js.Value, key string) interface{} {
	return UnwrapValue(obj.Get(key))
}

// ConsoleLog logs a message to the JavaScript console.
func ConsoleLog(msg string) {
	js.Global().Get("console").Call("log", msg)
}

// ConsoleError logs an error to the JavaScript console.
func ConsoleError(msg string) {
	js.Global().Get("console").Call("error", msg)
}

// ConsoleWarn logs a warning to the JavaScript console.
func ConsoleWarn(msg string) {
	js.Global().Get("console").Call("warn", msg)
}

// CreateErrorObject creates a structured JavaScript error object.
// This provides detailed error information for DWScript runtime errors.
func CreateErrorObject(errorType, message string, details map[string]interface{}) js.Value {
	errObj := js.Global().Get("Error").New(message)
	errObj.Set("type", errorType)
	errObj.Set("message", message)

	// Add additional details if provided
	if details != nil {
		for key, value := range details {
			errObj.Set(key, WrapValue(value))
		}
	}

	return errObj
}
