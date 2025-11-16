package runtime

import (
	"strings"
	"testing"
)

// ============================================================================
// ConversionError Tests
// ============================================================================

func TestConversionError(t *testing.T) {
	val := NewInteger(42)
	err := NewConversionError(val, "STRING", "invalid conversion")

	// Check error message
	msg := err.Error()
	if !strings.Contains(msg, "INTEGER") {
		t.Errorf("Error message should contain source type: %s", msg)
	}
	if !strings.Contains(msg, "STRING") {
		t.Errorf("Error message should contain target type: %s", msg)
	}
	if !strings.Contains(msg, "invalid conversion") {
		t.Errorf("Error message should contain reason: %s", msg)
	}

	// Check type assertion
	if !IsConversionError(err) {
		t.Error("IsConversionError() should return true")
	}
}

func TestConversionErrorNilValue(t *testing.T) {
	err := NewConversionError(nil, "INTEGER", "nil source")
	msg := err.Error()

	if !strings.Contains(msg, "nil") {
		t.Errorf("Error message should mention nil: %s", msg)
	}
}

// ============================================================================
// ArithmeticError Tests
// ============================================================================

func TestArithmeticError(t *testing.T) {
	err := NewArithmeticError("division by zero")

	// Check error message
	msg := err.Error()
	if !strings.Contains(msg, "division by zero") {
		t.Errorf("Error message should contain operation: %s", msg)
	}

	// Check type assertion
	if !IsArithmeticError(err) {
		t.Error("IsArithmeticError() should return true")
	}
}

// ============================================================================
// ComparisonError Tests
// ============================================================================

func TestComparisonError(t *testing.T) {
	left := NewInteger(42)
	right := NewString("hello")
	err := NewComparisonError(left, right, "<")

	// Check error message
	msg := err.Error()
	if !strings.Contains(msg, "INTEGER") {
		t.Errorf("Error message should contain left type: %s", msg)
	}
	if !strings.Contains(msg, "STRING") {
		t.Errorf("Error message should contain right type: %s", msg)
	}
	if !strings.Contains(msg, "<") {
		t.Errorf("Error message should contain operator: %s", msg)
	}

	// Check type assertion
	if !IsComparisonError(err) {
		t.Error("IsComparisonError() should return true")
	}
}

// ============================================================================
// IndexError Tests
// ============================================================================

func TestIndexError(t *testing.T) {
	err := NewIndexError(10, 0, 5, "array")

	// Check error message
	msg := err.Error()
	if !strings.Contains(msg, "10") {
		t.Errorf("Error message should contain index: %s", msg)
	}
	if !strings.Contains(msg, "array") {
		t.Errorf("Error message should contain type: %s", msg)
	}
	if !strings.Contains(msg, "[0..5]") {
		t.Errorf("Error message should contain range: %s", msg)
	}

	// Check type assertion
	if !IsIndexError(err) {
		t.Error("IsIndexError() should return true")
	}
}

// ============================================================================
// NilError Tests
// ============================================================================

func TestNilError(t *testing.T) {
	err := NewNilError("access field", "object")

	// Check error message
	msg := err.Error()
	if !strings.Contains(msg, "access field") {
		t.Errorf("Error message should contain operation: %s", msg)
	}
	if !strings.Contains(msg, "object") {
		t.Errorf("Error message should contain type: %s", msg)
	}

	// Check type assertion
	if !IsNilError(err) {
		t.Error("IsNilError() should return true")
	}
}

// ============================================================================
// TypeError Tests
// ============================================================================

func TestTypeError(t *testing.T) {
	val := NewInteger(42)
	err := NewTypeError("STRING", val, "function parameter")

	// Check error message
	msg := err.Error()
	if !strings.Contains(msg, "STRING") {
		t.Errorf("Error message should contain expected type: %s", msg)
	}
	if !strings.Contains(msg, "INTEGER") {
		t.Errorf("Error message should contain actual type: %s", msg)
	}
	if !strings.Contains(msg, "function parameter") {
		t.Errorf("Error message should contain context: %s", msg)
	}

	// Check type assertion
	if !IsTypeError(err) {
		t.Error("IsTypeError() should return true")
	}
}

func TestTypeErrorNoContext(t *testing.T) {
	val := NewString("hello")
	err := NewTypeError("INTEGER", val, "")

	msg := err.Error()
	// Should still work without context
	if !strings.Contains(msg, "INTEGER") || !strings.Contains(msg, "STRING") {
		t.Errorf("Error message should contain types: %s", msg)
	}
}

// ============================================================================
// Error Type Checking Tests
// ============================================================================

func TestErrorTypeChecking(t *testing.T) {
	// Create one error of each type
	convErr := NewConversionError(NewInteger(42), "STRING", "test")
	arithErr := NewArithmeticError("test")
	compErr := NewComparisonError(NewInteger(1), NewInteger(2), "=")
	idxErr := NewIndexError(10, 0, 5, "test")
	nilErr := NewNilError("test", "test")
	typeErr := NewTypeError("test", NewInteger(42), "test")

	// Test that each checker only returns true for its own type
	t.Run("ConversionError checks", func(t *testing.T) {
		if !IsConversionError(convErr) {
			t.Error("Should identify ConversionError")
		}
		if IsConversionError(arithErr) {
			t.Error("Should not identify ArithmeticError as ConversionError")
		}
	})

	t.Run("ArithmeticError checks", func(t *testing.T) {
		if !IsArithmeticError(arithErr) {
			t.Error("Should identify ArithmeticError")
		}
		if IsArithmeticError(convErr) {
			t.Error("Should not identify ConversionError as ArithmeticError")
		}
	})

	t.Run("ComparisonError checks", func(t *testing.T) {
		if !IsComparisonError(compErr) {
			t.Error("Should identify ComparisonError")
		}
		if IsComparisonError(convErr) {
			t.Error("Should not identify ConversionError as ComparisonError")
		}
	})

	t.Run("IndexError checks", func(t *testing.T) {
		if !IsIndexError(idxErr) {
			t.Error("Should identify IndexError")
		}
		if IsIndexError(convErr) {
			t.Error("Should not identify ConversionError as IndexError")
		}
	})

	t.Run("NilError checks", func(t *testing.T) {
		if !IsNilError(nilErr) {
			t.Error("Should identify NilError")
		}
		if IsNilError(convErr) {
			t.Error("Should not identify ConversionError as NilError")
		}
	})

	t.Run("TypeError checks", func(t *testing.T) {
		if !IsTypeError(typeErr) {
			t.Error("Should identify TypeError")
		}
		if IsTypeError(convErr) {
			t.Error("Should not identify ConversionError as TypeError")
		}
	})
}
