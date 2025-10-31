package interp

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func newTestInterpreter() *Interpreter {
	var buf bytes.Buffer
	return New(&buf)
}

func TestCallExternalFunctionSafeRecoversPanicError(t *testing.T) {
	interp := newTestInterpreter()
	interp.callExternalFunctionSafe(func() (Value, error) {
		panic(errors.New("boom"))
	})

	if interp.exception == nil {
		t.Fatalf("expected exception to be raised")
	}

	if interp.exception.ClassInfo.Name != "EHost" {
		t.Fatalf("expected exception class EHost, got %s", interp.exception.ClassInfo.Name)
	}

	if !strings.Contains(interp.exception.Message, "panic: boom") {
		t.Fatalf("expected panic message, got %q", interp.exception.Message)
	}

	if field, ok := interp.exception.Instance.Fields["ExceptionClass"]; ok {
		if str, isString := field.(*StringValue); !isString || str.Value == "" {
			t.Fatalf("expected ExceptionClass to contain panic type, got %#v", field)
		}
	} else {
		t.Fatalf("expected ExceptionClass field to be set")
	}
}

func TestCallExternalFunctionSafePropagatesError(t *testing.T) {
	interp := newTestInterpreter()
	err := errors.New("call failed")

	result := interp.callExternalFunctionSafe(func() (Value, error) {
		return &NilValue{}, err
	})

	if result.Type() != "NIL" {
		t.Fatalf("expected nil result, got %s", result.Type())
	}

	if interp.exception == nil {
		t.Fatalf("expected exception to be raised for error path")
	}

	if !strings.Contains(interp.exception.Message, err.Error()) {
		t.Fatalf("expected error message to be reflected, got %q", interp.exception.Message)
	}
}
