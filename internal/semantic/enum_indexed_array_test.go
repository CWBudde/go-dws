package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Enum-Indexed Array Tests
// Tests for Task 9.21.1: Enum-indexed arrays with non-zero ordinals
// ============================================================================

func TestEnumIndexedArrayWithNonZeroOrdinals(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "enum starting at 1",
			input: `
				type TDay = (Mon=1, Tue, Wed);
				const days: array[TDay] of String = ['Monday', 'Tuesday', 'Wednesday'];
			`,
		},
		{
			name: "enum starting at 0 (explicit)",
			input: `
				type TColor = (Red=0, Green, Blue);
				const colors: array[TColor] of String = ['Red', 'Green', 'Blue'];
			`,
		},
		{
			name: "enum with negative ordinals",
			input: `
				type TTemp = (Cold=-10, Warm=0, Hot=10);
				const temps: array[TTemp] of Integer = [-10, 0, 10];
			`,
		},
		{
			name: "enum with gaps in ordinals",
			input: `
				type TStatus = (Ok=200, BadRequest=400, ServerError=500);
				const messages: array[TStatus] of String = ['OK', 'Bad Request', 'Server Error'];
			`,
		},
		{
			name: "enum starting at large number",
			input: `
				type TCode = (First=100, Second, Third);
				const codes: array[TCode] of Integer = [100, 101, 102];
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

func TestEnumIndexedArrayErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "non-enum type as array index",
			input: `
				type TMyInt = Integer;
				const arr: array[TMyInt] of String = ['test'];
			`,
			expectedError: "unknown type",
		},
		{
			name: "undefined enum type as array index",
			input: `
				const arr: array[TUndefined] of String = ['test'];
			`,
			expectedError: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, tt.expectedError)
		})
	}
}

func TestEnumOrdinalHelperMethods(t *testing.T) {
	tests := []struct {
		name        string
		enumName    string
		enumValues  map[string]int
		orderedNames []string
		expectedMin int
		expectedMax int
	}{
		{
			name:        "enum starting at 1",
			enumName:    "TDay",
			enumValues:  map[string]int{"Mon": 1, "Tue": 2, "Wed": 3},
			orderedNames: []string{"Mon", "Tue", "Wed"},
			expectedMin: 1,
			expectedMax: 3,
		},
		{
			name:        "enum starting at 0",
			enumName:    "TColor",
			enumValues:  map[string]int{"Red": 0, "Green": 1, "Blue": 2},
			orderedNames: []string{"Red", "Green", "Blue"},
			expectedMin: 0,
			expectedMax: 2,
		},
		{
			name:        "enum with negative ordinals",
			enumName:    "TTemp",
			enumValues:  map[string]int{"Cold": -10, "Warm": 0, "Hot": 10},
			orderedNames: []string{"Cold", "Warm", "Hot"},
			expectedMin: -10,
			expectedMax: 10,
		},
		{
			name:        "enum with gaps",
			enumName:    "TStatus",
			enumValues:  map[string]int{"Ok": 200, "BadRequest": 400, "ServerError": 500},
			orderedNames: []string{"Ok", "BadRequest", "ServerError"},
			expectedMin: 200,
			expectedMax: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enumType := types.NewEnumType(tt.enumName, tt.enumValues, tt.orderedNames)

			minOrdinal := enumType.MinOrdinal()
			if minOrdinal != tt.expectedMin {
				t.Errorf("MinOrdinal() = %d, want %d", minOrdinal, tt.expectedMin)
			}

			maxOrdinal := enumType.MaxOrdinal()
			if maxOrdinal != tt.expectedMax {
				t.Errorf("MaxOrdinal() = %d, want %d", maxOrdinal, tt.expectedMax)
			}
		})
	}
}

func TestEnumIndexedArrayDeclaration(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "enum-indexed array in variable declaration",
			input: `
				type TDay = (Mon=1, Tue, Wed);
				var days: array[TDay] of String;
			`,
		},
		{
			name: "enum-indexed array in const declaration",
			input: `
				type TColor = (Red=0, Green, Blue);
				const colors: array[TColor] of String = ['Red', 'Green', 'Blue'];
			`,
		},
		{
			name: "multiple enum-indexed arrays",
			input: `
				type TDay = (Mon=1, Tue, Wed);
				type TMonth = (Jan=1, Feb, Mar);
				var days: array[TDay] of String;
				var months: array[TMonth] of String;
			`,
		},
		{
			name: "nested enum-indexed array",
			input: `
				type TDay = (Mon=1, Tue, Wed);
				type TColor = (Red=0, Green, Blue);
				var schedule: array[TDay] of array[TColor] of Integer;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}
