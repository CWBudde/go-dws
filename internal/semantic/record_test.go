package semantic

import (
	"testing"
)

// ============================================================================
// Record Declaration Tests
// ============================================================================

func TestRecordDeclaration(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "simple record",
			input: `
				type TPoint = record
					X, Y: Integer;
				end;
			`,
		},
		{
			name: "record with multiple field types",
			input: `
				type TPerson = record
					Name: String;
					Age: Integer;
					Score: Float;
				end;
			`,
		},
		{
			name: "record type in variable declaration",
			input: `
				type TPoint = record
					X, Y: Integer;
				end;
				var point: TPoint;
			`,
		},
		{
			name: "record with visibility sections",
			input: `
				type TPoint = record
				private
					FX, FY: Integer;
				public
					X, Y: Integer;
				end;
			`,
		},
		{
			name: "multiple records",
			input: `
				type TPoint = record
					X, Y: Integer;
				end;
				type TRect = record
					Left, Top, Right, Bottom: Integer;
				end;
				var p: TPoint;
				var r: TRect;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

func TestRecordErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "duplicate field names",
			input: `
				type TPoint = record
					X: Integer;
					X: Float;
				end;
			`,
			expectedError: "duplicate field 'X'",
		},
		{
			name: "undefined field type",
			input: `
				type TPoint = record
					X: UnknownType;
				end;
			`,
			expectedError: "unknown type 'UnknownType'",
		},
		{
			name: "undefined record type in variable",
			input: `
				var point: TPoint;
			`,
			expectedError: "unknown type 'TPoint'",
		},
		{
			name: "duplicate record type declaration",
			input: `
				type TPoint = record
					X, Y: Integer;
				end;
				type TPoint = record
					A, B: Float;
				end;
			`,
			expectedError: "already declared",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, tt.expectedError)
		})
	}
}

// Task 8.70: Type-check record literals
func TestRecordLiterals(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "named field record literal",
			input: `
				type TPoint = record
					X, Y: Integer;
				end;
				var p: TPoint := (X: 10, Y: 20);
			`,
		},
		{
			name: "record literal with all fields",
			input: `
				type TPerson = record
					Name: String;
					Age: Integer;
				end;
				var person: TPerson := (Name: 'Alice', Age: 30);
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

func TestRecordLiteralErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "undefined field in literal",
			input: `
				type TPoint = record
					X, Y: Integer;
				end;
				var p: TPoint := (X: 10, Z: 20);
			`,
			expectedError: "field 'Z' does not exist",
		},
		{
			name: "type mismatch in field value",
			input: `
				type TPoint = record
					X, Y: Integer;
				end;
				var p: TPoint := (X: 10, Y: 'hello');
			`,
			expectedError: "cannot assign String to Integer",
		},
		{
			name: "missing required fields",
			input: `
				type TPoint = record
					X, Y: Integer;
				end;
				var p: TPoint := (X: 10);
			`,
			expectedError: "missing required field 'Y'",
		},
		{
			name: "duplicate field in literal",
			input: `
				type TPoint = record
					X, Y: Integer;
				end;
				var p: TPoint := (X: 10, X: 20, Y: 30);
			`,
			expectedError: "duplicate field 'X'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, tt.expectedError)
		})
	}
}

// Task 8.71: Type-check record field access
func TestRecordFieldAccess(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "read field value",
			input: `
				type TPoint = record
					X, Y: Integer;
				end;
				var p: TPoint;
				var x: Integer := p.X;
			`,
		},
		// Note: Field assignment via member access (p.X := 10) requires
		// parser support for AssignmentStatement with Expression target
		// Currently the parser only supports simple identifier assignment
		{
			name: "field access in expression",
			input: `
				type TPoint = record
					X, Y: Integer;
				end;
				var p: TPoint;
				var sum: Integer := p.X + p.Y;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

func TestRecordFieldAccessErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "undefined field access",
			input: `
				type TPoint = record
					X, Y: Integer;
				end;
				var p: TPoint;
				var z: Integer := p.Z;
			`,
			expectedError: "field 'Z' does not exist",
		},
		{
			name: "field access on non-record type",
			input: `
				var x: Integer := 42;
				var y: Integer := x.SomeField;
			`,
			// Task 9.83: With helper support, the error message changed
			expectedError: "requires a helper",
		},
		// Note: Field assignment tests removed - require parser support
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, tt.expectedError)
		})
	}
}

// Test record type compatibility
func TestRecordTypeCompatibility(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "assign record to same type",
			input: `
				type TPoint = record
					X, Y: Integer;
				end;
				var p1: TPoint;
				var p2: TPoint;
				p1 := p2;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

func TestRecordTypeCompatibilityErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "assign different record types",
			input: `
				type TPoint = record
					X, Y: Integer;
				end;
				type TVector = record
					X, Y: Integer;
				end;
				var p: TPoint;
				var v: TVector;
				p := v;
			`,
			expectedError: "cannot assign TVector to TPoint",
		},
		{
			name: "assign integer to record",
			input: `
				type TPoint = record
					X, Y: Integer;
				end;
				var p: TPoint := 42;
			`,
			expectedError: "cannot assign Integer to TPoint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, tt.expectedError)
		})
	}
}
