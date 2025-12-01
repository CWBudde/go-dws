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

	if field := interp.exception.Instance.GetField("ExceptionClass"); field != nil {
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

// TestGoErrorToExceptionConversion tests that Go errors are properly converted to DWScript exceptions.
func TestGoErrorToExceptionConversion(t *testing.T) {
	t.Run("BasicErrorConversion", func(t *testing.T) {
		interp := newTestInterpreter()
		testErr := errors.New("test error message")

		// Simulate what happens when an external function returns an error
		interp.raiseGoErrorAsException(testErr)

		if interp.exception == nil {
			t.Fatal("expected exception to be set")
		}

		if interp.exception.Message != "test error message" {
			t.Errorf("expected message 'test error message', got %q", interp.exception.Message)
		}

		if interp.exception.ClassInfo.Name != "EHost" {
			t.Errorf("expected exception class 'EHost', got %q", interp.exception.ClassInfo.Name)
		}
	})

	t.Run("ExceptionClassField", func(t *testing.T) {
		interp := newTestInterpreter()
		testErr := errors.New("custom error")

		interp.raiseGoErrorAsException(testErr)

		// Verify ExceptionClass field contains the Go type
		exceptionClassField := interp.exception.Instance.GetField("ExceptionClass")
		if exceptionClassField == nil {
			t.Fatal("expected ExceptionClass field to exist")
		}

		exceptionClassStr, ok := exceptionClassField.(*StringValue)
		if !ok {
			t.Fatalf("expected ExceptionClass to be StringValue, got %T", exceptionClassField)
		}

		if exceptionClassStr.Value != "*errors.errorString" {
			t.Errorf("expected ExceptionClass '*errors.errorString', got %q", exceptionClassStr.Value)
		}
	})

	t.Run("CallStackCaptured", func(t *testing.T) {
		interp := newTestInterpreter()

		// Push some call frames
		interp.pushCallStack("function1")
		interp.pushCallStack("function2")

		testErr := errors.New("error in nested call")
		interp.raiseGoErrorAsException(testErr)

		if interp.exception == nil {
			t.Fatal("expected exception to be set")
		}

		if len(interp.exception.CallStack) != 2 {
			t.Errorf("expected call stack length 2, got %d", len(interp.exception.CallStack))
		}

		if interp.exception.CallStack[0].FunctionName != "function1" {
			t.Errorf("expected first call stack entry 'function1', got %q", interp.exception.CallStack[0].FunctionName)
		}
	})

	t.Run("NilErrorDoesNothing", func(t *testing.T) {
		interp := newTestInterpreter()

		// Calling with nil error should not set exception
		interp.raiseGoErrorAsException(nil)

		if interp.exception != nil {
			t.Error("expected no exception for nil error")
		}
	})
}

// TestErrorPropagationNestedCalls tests that errors from Go functions propagate correctly
// through nested DWScript call stacks.
func TestErrorPropagationNestedCalls(t *testing.T) {
	t.Run("ErrorInNestedGoFunction", func(t *testing.T) {
		interp := newTestInterpreter()

		// Simulate nested call stack: DWScript → Go func A → DWScript → Go func B (errors)
		// Set up call stack to simulate this
		interp.pushCallStack("outerDWScriptFunction")
		interp.pushCallStack("goFunctionA")
		interp.pushCallStack("innerDWScriptFunction")

		// Function B returns an error
		testErr := errors.New("error in nested function")
		interp.raiseGoErrorAsException(testErr)

		if interp.exception == nil {
			t.Fatal("expected exception to be raised")
		}

		// Verify exception message
		if interp.exception.Message != "error in nested function" {
			t.Errorf("expected message 'error in nested function', got %q", interp.exception.Message)
		}

		// Verify call stack is captured with all frames
		if len(interp.exception.CallStack) != 3 {
			t.Errorf("expected call stack length 3, got %d", len(interp.exception.CallStack))
		}

		// Verify call stack frames are correct
		expectedFrames := []string{"outerDWScriptFunction", "goFunctionA", "innerDWScriptFunction"}
		for i, expected := range expectedFrames {
			if i >= len(interp.exception.CallStack) {
				break
			}
			if interp.exception.CallStack[i].FunctionName != expected {
				t.Errorf("call stack frame %d: expected %q, got %q", i, expected, interp.exception.CallStack[i].FunctionName)
			}
		}
	})

	t.Run("MultipleNestedErrors", func(t *testing.T) {
		interp := newTestInterpreter()

		// First error in nested call
		interp.pushCallStack("level1")
		interp.pushCallStack("level2")

		firstErr := errors.New("first error")
		interp.raiseGoErrorAsException(firstErr)

		if interp.exception == nil {
			t.Fatal("expected first exception to be set")
		}

		firstException := interp.exception
		if !strings.Contains(firstException.Message, "first error") {
			t.Errorf("expected first error message, got %q", firstException.Message)
		}

		// Clear exception and test second error
		interp.exception = nil
		interp.pushCallStack("level3")

		secondErr := errors.New("second error")
		interp.raiseGoErrorAsException(secondErr)

		if interp.exception == nil {
			t.Fatal("expected second exception to be set")
		}

		if !strings.Contains(interp.exception.Message, "second error") {
			t.Errorf("expected second error message, got %q", interp.exception.Message)
		}

		// Call stack should have 3 frames for second error
		if len(interp.exception.CallStack) != 3 {
			t.Errorf("expected call stack length 3, got %d", len(interp.exception.CallStack))
		}
	})

	t.Run("ErrorPropagationWithEmptyCallStack", func(t *testing.T) {
		interp := newTestInterpreter()

		// No call stack frames
		testErr := errors.New("top-level error")
		interp.raiseGoErrorAsException(testErr)

		if interp.exception == nil {
			t.Fatal("expected exception to be set")
		}

		// Call stack should be empty but not nil
		if interp.exception.CallStack == nil {
			t.Error("expected call stack to be non-nil even when empty")
		}

		if len(interp.exception.CallStack) != 0 {
			t.Errorf("expected empty call stack, got %d frames", len(interp.exception.CallStack))
		}
	})
}

// TestEHostExceptionSpecificFeatures tests the specific features of EHost exceptions.
func TestEHostExceptionSpecificFeatures(t *testing.T) {
	t.Run("EHostInheritsFromException", func(t *testing.T) {
		interp := newTestInterpreter()

		// Verify EHost class exists
		eHostClass, exists := interp.classes["ehost"]
		if !exists {
			t.Fatal("expected EHost class to be registered")
		}

		// Verify EHost inherits from Exception
		if eHostClass.Parent == nil {
			t.Fatal("expected EHost to have a parent class")
		}

		if eHostClass.Parent.Name != "Exception" {
			t.Errorf("expected EHost parent to be 'Exception', got %q", eHostClass.Parent.Name)
		}

		// Verify InheritsFrom works correctly
		if !eHostClass.InheritsFrom("Exception") {
			t.Error("expected EHost to inherit from Exception")
		}

		if !eHostClass.InheritsFrom("EHost") {
			t.Error("expected EHost to inherit from itself")
		}
	})

	t.Run("EHostHasExceptionClassField", func(t *testing.T) {
		interp := newTestInterpreter()
		testErr := errors.New("test error")

		interp.raiseGoErrorAsException(testErr)

		if interp.exception == nil {
			t.Fatal("expected exception to be set")
		}

		// Verify ExceptionClass field exists and is populated
		exceptionClassField := interp.exception.Instance.GetField("ExceptionClass")
		if exceptionClassField == nil {
			t.Fatal("expected ExceptionClass field to exist in EHost exception")
		}

		exceptionClassStr, ok := exceptionClassField.(*StringValue)
		if !ok {
			t.Fatalf("expected ExceptionClass to be StringValue, got %T", exceptionClassField)
		}

		// Should contain the Go error type
		if exceptionClassStr.Value == "" {
			t.Error("expected ExceptionClass to be non-empty")
		}

		if !strings.Contains(exceptionClassStr.Value, "error") {
			t.Errorf("expected ExceptionClass to contain 'error', got %q", exceptionClassStr.Value)
		}
	})

	t.Run("EHostMessageField", func(t *testing.T) {
		interp := newTestInterpreter()
		testErr := errors.New("detailed error message")

		interp.raiseGoErrorAsException(testErr)

		if interp.exception == nil {
			t.Fatal("expected exception to be set")
		}

		// Verify Message field exists
		messageField := interp.exception.Instance.GetField("Message")
		if messageField == nil {
			t.Fatal("expected Message field to exist in EHost exception")
		}

		messageStr, ok := messageField.(*StringValue)
		if !ok {
			t.Fatalf("expected Message to be StringValue, got %T", messageField)
		}

		if messageStr.Value != "detailed error message" {
			t.Errorf("expected Message 'detailed error message', got %q", messageStr.Value)
		}

		// Also verify the exception's Message field
		if interp.exception.Message != "detailed error message" {
			t.Errorf("expected exception.Message 'detailed error message', got %q", interp.exception.Message)
		}
	})

	t.Run("EHostWithDifferentErrorTypes", func(t *testing.T) {
		interp := newTestInterpreter()

		// Test with different error types to verify ExceptionClass captures them correctly
		testCases := []struct {
			name          string
			err           error
			expectedClass string
		}{
			{
				name:          "errorString",
				err:           errors.New("simple error"),
				expectedClass: "*errors.errorString",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Clear previous exception
				interp.exception = nil

				interp.raiseGoErrorAsException(tc.err)

				if interp.exception == nil {
					t.Fatal("expected exception to be set")
				}

				exceptionClassField := interp.exception.Instance.GetField("ExceptionClass")
				if exceptionClassField == nil {
					t.Fatal("expected ExceptionClass field")
				}

				exceptionClassStr := exceptionClassField.(*StringValue)
				if exceptionClassStr.Value != tc.expectedClass {
					t.Errorf("expected ExceptionClass %q, got %q", tc.expectedClass, exceptionClassStr.Value)
				}
			})
		}
	})

	t.Run("EHostFallbackToBaseException", func(t *testing.T) {
		interp := newTestInterpreter()

		// Temporarily remove EHost class to test fallback
		eHostClass := interp.classes["ehost"]
		delete(interp.classes, "ehost")
		defer func() {
			interp.classes["ehost"] = eHostClass
		}()

		testErr := errors.New("error without EHost")
		interp.raiseGoErrorAsException(testErr)

		if interp.exception == nil {
			t.Fatal("expected exception to be set even without EHost class")
		}

		// Should fall back to Exception class
		if interp.exception.ClassInfo.Name != "Exception" {
			t.Errorf("expected fallback to 'Exception' class, got %q", interp.exception.ClassInfo.Name)
		}

		// Message should still be set
		if interp.exception.Message != "error without EHost" {
			t.Errorf("expected message to be preserved, got %q", interp.exception.Message)
		}
	})

	t.Run("EHostExceptionInstanceType", func(t *testing.T) {
		interp := newTestInterpreter()
		testErr := errors.New("instance type test")

		interp.raiseGoErrorAsException(testErr)

		if interp.exception == nil {
			t.Fatal("expected exception to be set")
		}

		// Verify exception.Instance is an ObjectInstance
		if interp.exception.Instance == nil {
			t.Fatal("expected exception instance to be non-nil")
		}

		// Verify instance has Class pointing to EHost
		if interp.exception.Instance.Class.Name != "EHost" {
			t.Errorf("expected instance Class to be 'EHost', got %q", interp.exception.Instance.Class.Name)
		}

		// Verify instance has Fields map
		if interp.exception.Instance.Fields == nil {
			t.Fatal("expected instance Fields to be non-nil")
		}
	})
}
