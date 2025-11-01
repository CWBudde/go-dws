package types

import (
	"strings"
	"testing"
)

// ============================================================================
// SubrangeType Tests
// ============================================================================

// Test creating subrange types
func TestSubrangeTypeCreation(t *testing.T) {
	t.Run("Create basic integer subrange", func(t *testing.T) {
		digitType := &SubrangeType{
			BaseType:  INTEGER,
			Name:      "TDigit",
			LowBound:  0,
			HighBound: 9,
		}

		if digitType.Name != "TDigit" {
			t.Errorf("Name = %v, want TDigit", digitType.Name)
		}
		if !digitType.BaseType.Equals(INTEGER) {
			t.Errorf("BaseType = %v, want Integer", digitType.BaseType)
		}
		if digitType.LowBound != 0 {
			t.Errorf("LowBound = %v, want 0", digitType.LowBound)
		}
		if digitType.HighBound != 9 {
			t.Errorf("HighBound = %v, want 9", digitType.HighBound)
		}
	})

	t.Run("Create percentage subrange", func(t *testing.T) {
		percentType := &SubrangeType{
			BaseType:  INTEGER,
			Name:      "TPercent",
			LowBound:  0,
			HighBound: 100,
		}

		if percentType.LowBound != 0 || percentType.HighBound != 100 {
			t.Errorf("Bounds = %v..%v, want 0..100", percentType.LowBound, percentType.HighBound)
		}
		if percentType.Name != "TPercent" {
			t.Errorf("Name = %v, want TPercent", percentType.Name)
		}
		if !percentType.BaseType.Equals(INTEGER) {
			t.Errorf("BaseType = %v, want Integer", percentType.BaseType)
		}
	})

	t.Run("Create negative range subrange", func(t *testing.T) {
		tempType := &SubrangeType{
			BaseType:  INTEGER,
			Name:      "TTemperature",
			LowBound:  -40,
			HighBound: 50,
		}

		if tempType.LowBound != -40 || tempType.HighBound != 50 {
			t.Errorf("Bounds = %v..%v, want -40..50", tempType.LowBound, tempType.HighBound)
		}
		if tempType.Name != "TTemperature" {
			t.Errorf("Name = %v, want TTemperature", tempType.Name)
		}
		if !tempType.BaseType.Equals(INTEGER) {
			t.Errorf("BaseType = %v, want Integer", tempType.BaseType)
		}
	})
}

// Test TypeKind returns "SUBRANGE"
func TestSubrangeTypeKind(t *testing.T) {
	subrange := &SubrangeType{
		BaseType:  INTEGER,
		Name:      "TDigit",
		LowBound:  0,
		HighBound: 9,
	}

	if subrange.TypeKind() != "SUBRANGE" {
		t.Errorf("TypeKind() = %v, want SUBRANGE", subrange.TypeKind())
	}
}

// Test String() returns proper format
func TestSubrangeString(t *testing.T) {
	tests := []struct {
		name      string
		subrange  *SubrangeType
		wantParts []string // Parts that should be in the string
	}{
		{
			name: "Digit range",
			subrange: &SubrangeType{
				BaseType:  INTEGER,
				Name:      "TDigit",
				LowBound:  0,
				HighBound: 9,
			},
			wantParts: []string{"0", "9", ".."},
		},
		{
			name: "Percentage range",
			subrange: &SubrangeType{
				BaseType:  INTEGER,
				Name:      "TPercent",
				LowBound:  0,
				HighBound: 100,
			},
			wantParts: []string{"0", "100", ".."},
		},
		{
			name: "Negative range",
			subrange: &SubrangeType{
				BaseType:  INTEGER,
				Name:      "TTemperature",
				LowBound:  -40,
				HighBound: 50,
			},
			wantParts: []string{"-40", "50", ".."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.subrange.String()
			for _, part := range tt.wantParts {
				if !strings.Contains(result, part) {
					t.Errorf("String() = %v, should contain %v", result, part)
				}
			}
		})
	}
}

// Test Equals() method
func TestSubrangeEquals(t *testing.T) {
	digit1 := &SubrangeType{
		BaseType:  INTEGER,
		Name:      "TDigit",
		LowBound:  0,
		HighBound: 9,
	}

	digit2 := &SubrangeType{
		BaseType:  INTEGER,
		Name:      "TDigit2",
		LowBound:  0,
		HighBound: 9,
	}

	differentBounds := &SubrangeType{
		BaseType:  INTEGER,
		Name:      "TSmallDigit",
		LowBound:  0,
		HighBound: 5,
	}

	t.Run("Same base type and bounds should be equal", func(t *testing.T) {
		if !digit1.Equals(digit2) {
			t.Error("Subranges with same base type and bounds should be equal")
		}
	})

	t.Run("Different bounds should not be equal", func(t *testing.T) {
		if digit1.Equals(differentBounds) {
			t.Error("Subranges with different bounds should not be equal")
		}
	})

	t.Run("Subrange should not equal base type", func(t *testing.T) {
		if digit1.Equals(INTEGER) {
			t.Error("Subrange should not equal its base type")
		}
	})

	t.Run("Subrange should not equal other types", func(t *testing.T) {
		if digit1.Equals(STRING) {
			t.Error("Subrange should not equal String type")
		}
		if digit1.Equals(FLOAT) {
			t.Error("Subrange should not equal Float type")
		}
	})
}

// Test Contains() method
func TestSubrangeContains(t *testing.T) {
	digit := &SubrangeType{
		BaseType:  INTEGER,
		Name:      "TDigit",
		LowBound:  0,
		HighBound: 9,
	}

	tests := []struct {
		name     string
		value    int
		expected bool
	}{
		{"Value at low bound", 0, true},
		{"Value at high bound", 9, true},
		{"Value in middle", 5, true},
		{"Value below range", -1, false},
		{"Value above range", 10, false},
		{"Value far below", -100, false},
		{"Value far above", 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := digit.Contains(tt.value)
			if result != tt.expected {
				t.Errorf("Contains(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

// Test range validation function
func TestValidateRange(t *testing.T) {
	percent := &SubrangeType{
		BaseType:  INTEGER,
		Name:      "TPercent",
		LowBound:  0,
		HighBound: 100,
	}

	t.Run("Valid value in range", func(t *testing.T) {
		err := ValidateRange(50, percent)
		if err != nil {
			t.Errorf("ValidateRange(50) returned error: %v", err)
		}
	})

	t.Run("Valid value at low bound", func(t *testing.T) {
		err := ValidateRange(0, percent)
		if err != nil {
			t.Errorf("ValidateRange(0) returned error: %v", err)
		}
	})

	t.Run("Valid value at high bound", func(t *testing.T) {
		err := ValidateRange(100, percent)
		if err != nil {
			t.Errorf("ValidateRange(100) returned error: %v", err)
		}
	})

	t.Run("Invalid value below range", func(t *testing.T) {
		err := ValidateRange(-1, percent)
		if err == nil {
			t.Error("ValidateRange(-1) should return error")
		}
		if err != nil && !strings.Contains(err.Error(), "out of range") {
			t.Errorf("Error message should mention 'out of range', got: %v", err)
		}
	})

	t.Run("Invalid value above range", func(t *testing.T) {
		err := ValidateRange(101, percent)
		if err == nil {
			t.Error("ValidateRange(101) should return error")
		}
		if err != nil && !strings.Contains(err.Error(), "out of range") {
			t.Errorf("Error message should mention 'out of range', got: %v", err)
		}
	})

	t.Run("Error message includes value and bounds", func(t *testing.T) {
		err := ValidateRange(150, percent)
		if err == nil {
			t.Error("ValidateRange(150) should return error")
		}
		if err != nil {
			errMsg := err.Error()
			if !strings.Contains(errMsg, "150") {
				t.Errorf("Error should contain value 150, got: %v", errMsg)
			}
		}
	})
}

// Test type compatibility
func TestSubrangeTypeCompatibility(t *testing.T) {
	digit := &SubrangeType{
		BaseType:  INTEGER,
		Name:      "TDigit",
		LowBound:  0,
		HighBound: 9,
	}

	t.Run("Subrange has correct base type", func(t *testing.T) {
		if !digit.BaseType.Equals(INTEGER) {
			t.Error("Subrange base type should be Integer")
		}
	})

	t.Run("Subrange is not directly assignable to base type", func(t *testing.T) {
		// Subranges are NOT transparent like type aliases
		// They require explicit validation
		if digit.Equals(INTEGER) {
			t.Error("Subrange should not equal its base type")
		}
	})
}

// Test nested subranges
func TestNestedSubranges(t *testing.T) {
	t.Run("Create nested subrange of subrange", func(t *testing.T) {
		// type TSmallDigit = 0..5;
		smallDigit := &SubrangeType{
			BaseType:  INTEGER,
			Name:      "TSmallDigit",
			LowBound:  0,
			HighBound: 5,
		}

		// type TTinyDigit: TSmallDigit = 0..3;
		// Nested subrange still uses Integer as base type
		tinyDigit := &SubrangeType{
			BaseType:  INTEGER, // Base type is still Integer, not TSmallDigit
			Name:      "TTinyDigit",
			LowBound:  0,
			HighBound: 3,
		}

		// Verify bounds are properly restricted
		if tinyDigit.HighBound > smallDigit.HighBound {
			t.Error("Nested subrange should have tighter bounds")
		}

		// Both should have Integer as base type
		if !tinyDigit.BaseType.Equals(INTEGER) {
			t.Error("Nested subrange should have Integer as base type")
		}
	})

	t.Run("Nested subrange validation", func(t *testing.T) {
		tinyDigit := &SubrangeType{
			BaseType:  INTEGER,
			Name:      "TTinyDigit",
			LowBound:  0,
			HighBound: 3,
		}

		// Value 3 is valid for TTinyDigit
		if !tinyDigit.Contains(3) {
			t.Error("Value 3 should be valid for TTinyDigit (0..3)")
		}

		// Value 4 is not valid for TTinyDigit
		if tinyDigit.Contains(4) {
			t.Error("Value 4 should not be valid for TTinyDigit (0..3)")
		}
	})
}

// Test edge cases
func TestSubrangeEdgeCases(t *testing.T) {
	t.Run("Single value range", func(t *testing.T) {
		single := &SubrangeType{
			BaseType:  INTEGER,
			Name:      "TSingleValue",
			LowBound:  42,
			HighBound: 42,
		}

		if !single.Contains(42) {
			t.Error("Should contain the single allowed value")
		}
		if single.Contains(41) || single.Contains(43) {
			t.Error("Should only contain the single value")
		}
	})

	t.Run("Large range", func(t *testing.T) {
		large := &SubrangeType{
			BaseType:  INTEGER,
			Name:      "TLargeRange",
			LowBound:  -1000000,
			HighBound: 1000000,
		}

		if !large.Contains(0) {
			t.Error("Should contain middle value")
		}
		if !large.Contains(-1000000) {
			t.Error("Should contain low bound")
		}
		if !large.Contains(1000000) {
			t.Error("Should contain high bound")
		}
	})
}
