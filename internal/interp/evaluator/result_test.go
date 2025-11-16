package evaluator

import (
	"testing"
)

// Mock value types for testing
type mockValue struct {
	val string
}

func (m *mockValue) Type() string   { return "MOCK" }
func (m *mockValue) String() string { return m.val }

type mockError struct {
	msg string
}

func (m *mockError) Type() string   { return "ERROR" }
func (m *mockError) String() string { return "ERROR: " + m.msg }

func TestNewResult(t *testing.T) {
	val := &mockValue{val: "test"}
	result := NewResult(val)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Val() != val {
		t.Error("Expected value to be preserved")
	}
}

func TestError(t *testing.T) {
	t.Run("returns nil for non-error value", func(t *testing.T) {
		val := &mockValue{val: "test"}
		result := NewResult(val)

		if err := result.Error(); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("returns error for error value", func(t *testing.T) {
		err := &mockError{msg: "test error"}
		result := NewResult(err)

		if result.Error() == nil {
			t.Error("Expected error to be returned")
		}
		if result.Error() != err {
			t.Error("Expected same error object")
		}
	})

	t.Run("handles nil value", func(t *testing.T) {
		result := NewResult(nil)

		if err := result.Error(); err != nil {
			t.Errorf("Expected no error for nil value, got %v", err)
		}
	})
}

func TestIsError(t *testing.T) {
	t.Run("returns false for non-error", func(t *testing.T) {
		val := &mockValue{val: "test"}
		result := NewResult(val)

		if result.IsError() {
			t.Error("Expected IsError to return false")
		}
	})

	t.Run("returns true for error", func(t *testing.T) {
		err := &mockError{msg: "test error"}
		result := NewResult(err)

		if !result.IsError() {
			t.Error("Expected IsError to return true")
		}
	})
}

func TestIsOk(t *testing.T) {
	t.Run("returns true for non-error", func(t *testing.T) {
		val := &mockValue{val: "test"}
		result := NewResult(val)

		if !result.IsOk() {
			t.Error("Expected IsOk to return true")
		}
	})

	t.Run("returns false for error", func(t *testing.T) {
		err := &mockError{msg: "test error"}
		result := NewResult(err)

		if result.IsOk() {
			t.Error("Expected IsOk to return false")
		}
	})
}

func TestOrReturn(t *testing.T) {
	t.Run("returns value and nil for success", func(t *testing.T) {
		val := &mockValue{val: "test"}
		result := NewResult(val)

		value, err := result.OrReturn()

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if value != val {
			t.Error("Expected original value")
		}
	})

	t.Run("returns error twice for error", func(t *testing.T) {
		errVal := &mockError{msg: "test error"}
		result := NewResult(errVal)

		value, err := result.OrReturn()

		if err == nil {
			t.Error("Expected error")
		}
		if value != errVal {
			t.Error("Expected value to be error")
		}
		if err != errVal {
			t.Error("Expected err to be error")
		}
	})
}

func TestUnwrap(t *testing.T) {
	t.Run("returns value for success", func(t *testing.T) {
		val := &mockValue{val: "test"}
		result := NewResult(val)

		unwrapped := result.Unwrap()

		if unwrapped != val {
			t.Error("Expected original value")
		}
	})

	t.Run("returns error for error", func(t *testing.T) {
		errVal := &mockError{msg: "test error"}
		result := NewResult(errVal)

		unwrapped := result.Unwrap()

		if unwrapped != errVal {
			t.Error("Expected error value")
		}
	})
}

func TestMap(t *testing.T) {
	t.Run("applies function to success value", func(t *testing.T) {
		val := &mockValue{val: "test"}
		result := NewResult(val)

		mapped := result.Map(func(v Value) Value {
			return &mockValue{val: "mapped"}
		})

		if mapped.IsError() {
			t.Error("Expected success after map")
		}
		if mapped.Val().String() != "mapped" {
			t.Errorf("Expected 'mapped', got %s", mapped.Val().String())
		}
	})

	t.Run("skips function for error", func(t *testing.T) {
		errVal := &mockError{msg: "test error"}
		result := NewResult(errVal)

		called := false
		mapped := result.Map(func(v Value) Value {
			called = true
			return &mockValue{val: "should not see this"}
		})

		if called {
			t.Error("Expected function not to be called on error")
		}
		if !mapped.IsError() {
			t.Error("Expected error to be propagated")
		}
		if mapped.Error() != errVal {
			t.Error("Expected same error")
		}
	})
}

func TestAndThen(t *testing.T) {
	t.Run("chains evaluation on success", func(t *testing.T) {
		val := &mockValue{val: "test"}
		result := NewResult(val)

		chained := result.AndThen(func(v Value) Value {
			return &mockValue{val: "chained"}
		})

		if chained.IsError() {
			t.Error("Expected success after AndThen")
		}
		if chained.Val().String() != "chained" {
			t.Errorf("Expected 'chained', got %s", chained.Val().String())
		}
	})

	t.Run("skips evaluation for error", func(t *testing.T) {
		errVal := &mockError{msg: "test error"}
		result := NewResult(errVal)

		called := false
		chained := result.AndThen(func(v Value) Value {
			called = true
			return &mockValue{val: "should not see this"}
		})

		if called {
			t.Error("Expected function not to be called on error")
		}
		if !chained.IsError() {
			t.Error("Expected error to be propagated")
		}
		if chained.Error() != errVal {
			t.Error("Expected same error")
		}
	})

	t.Run("propagates error from chained evaluation", func(t *testing.T) {
		val := &mockValue{val: "test"}
		result := NewResult(val)

		chainedErr := &mockError{msg: "chained error"}
		chained := result.AndThen(func(v Value) Value {
			return chainedErr
		})

		if !chained.IsError() {
			t.Error("Expected error from chained evaluation")
		}
		if chained.Error() != chainedErr {
			t.Error("Expected chained error")
		}
	})
}

func TestFirstError(t *testing.T) {
	t.Run("returns nil for all successes", func(t *testing.T) {
		r1 := NewResult(&mockValue{val: "1"})
		r2 := NewResult(&mockValue{val: "2"})
		r3 := NewResult(&mockValue{val: "3"})

		err := FirstError(r1, r2, r3)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("returns first error", func(t *testing.T) {
		r1 := NewResult(&mockValue{val: "1"})
		err1 := &mockError{msg: "error 1"}
		r2 := NewResult(err1)
		err2 := &mockError{msg: "error 2"}
		r3 := NewResult(err2)

		err := FirstError(r1, r2, r3)

		if err == nil {
			t.Error("Expected error")
		}
		if err != err1 {
			t.Error("Expected first error")
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		err := FirstError()

		if err != nil {
			t.Errorf("Expected no error for empty list, got %v", err)
		}
	})
}

func TestAllValues(t *testing.T) {
	t.Run("returns all values for successes", func(t *testing.T) {
		v1 := &mockValue{val: "1"}
		v2 := &mockValue{val: "2"}
		v3 := &mockValue{val: "3"}

		values, err := AllValues(
			NewResult(v1),
			NewResult(v2),
			NewResult(v3),
		)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(values) != 3 {
			t.Errorf("Expected 3 values, got %d", len(values))
		}
		if values[0] != v1 || values[1] != v2 || values[2] != v3 {
			t.Error("Expected values in order")
		}
	})

	t.Run("returns first error", func(t *testing.T) {
		v1 := &mockValue{val: "1"}
		err1 := &mockError{msg: "error 1"}
		err2 := &mockError{msg: "error 2"}

		values, err := AllValues(
			NewResult(v1),
			NewResult(err1),
			NewResult(err2),
		)

		if err == nil {
			t.Error("Expected error")
		}
		if err != err1 {
			t.Error("Expected first error")
		}
		if values != nil {
			t.Error("Expected nil values on error")
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		values, err := AllValues()

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(values) != 0 {
			t.Errorf("Expected empty values, got %d", len(values))
		}
	})
}

func TestCollect(t *testing.T) {
	t.Run("collects all values for successes", func(t *testing.T) {
		v1 := &mockValue{val: "1"}
		v2 := &mockValue{val: "2"}
		v3 := &mockValue{val: "3"}

		values, err := Collect(func(collect func(Value)) {
			collect(v1)
			collect(v2)
			collect(v3)
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(values) != 3 {
			t.Errorf("Expected 3 values, got %d", len(values))
		}
		if values[0] != v1 || values[1] != v2 || values[2] != v3 {
			t.Error("Expected values in order")
		}
	})

	t.Run("stops collecting on first error", func(t *testing.T) {
		v1 := &mockValue{val: "1"}
		err1 := &mockError{msg: "error 1"}
		v3 := &mockValue{val: "3"}

		collected := 0
		values, err := Collect(func(collect func(Value)) {
			collected++
			collect(v1)
			collected++
			collect(err1)
			collected++
			collect(v3) // Should not collect this
		})

		if err == nil {
			t.Error("Expected error")
		}
		if err != err1 {
			t.Error("Expected first error")
		}
		if values != nil {
			t.Error("Expected nil values on error")
		}
		// Function should still be called 3 times (doesn't short-circuit the function itself)
		if collected != 3 {
			t.Errorf("Expected function to be called 3 times, got %d", collected)
		}
	})

	t.Run("handles empty collection", func(t *testing.T) {
		values, err := Collect(func(collect func(Value)) {
			// Don't collect anything
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(values) != 0 {
			t.Errorf("Expected empty values, got %d", len(values))
		}
	})
}

// Integration test: Demonstrate typical usage patterns
func TestUsagePatterns(t *testing.T) {
	t.Run("Pattern 1: OrReturn early exit", func(t *testing.T) {
		// Simulate evaluating left and right operands
		leftVal := &mockValue{val: "left"}
		rightErr := &mockError{msg: "right failed"}

		// This simulates what would happen in evalBinaryExpression
		_, err := NewResult(leftVal).OrReturn()
		if err != nil {
			t.Error("Left should succeed")
		}

		rightResult, err := NewResult(rightErr).OrReturn()
		if err == nil {
			t.Error("Right should fail")
		}

		// In real code, we would return err here
		if rightResult.Type() != "ERROR" {
			t.Error("Expected error type")
		}
	})

	t.Run("Pattern 2: FirstError for multiple values", func(t *testing.T) {
		left := NewResult(&mockValue{val: "left"})
		right := NewResult(&mockError{msg: "right failed"})

		if err := FirstError(left, right); err == nil {
			t.Error("Expected error from FirstError")
		}
	})

	t.Run("Pattern 3: Map for transformations", func(t *testing.T) {
		result := NewResult(&mockValue{val: "5"}).
			Map(func(v Value) Value {
				// Simulate converting string to number
				return &mockValue{val: "transformed"}
			})

		if result.IsError() {
			t.Error("Expected success")
		}
		if result.Val().String() != "transformed" {
			t.Error("Expected transformed value")
		}
	})

	t.Run("Pattern 4: AndThen for chaining", func(t *testing.T) {
		result := NewResult(&mockValue{val: "first"}).
			AndThen(func(v Value) Value {
				return &mockValue{val: "second"}
			}).
			AndThen(func(v Value) Value {
				return &mockValue{val: "third"}
			})

		if result.IsError() {
			t.Error("Expected success")
		}
		if result.Val().String() != "third" {
			t.Error("Expected final chained value")
		}
	})
}
