package interp

import (
	"strings"
	"testing"
)

// TestBuiltinRandom_BasicUsage tests that Random returns values in range [0, 1)
func TestBuiltinRandom_BasicUsage(t *testing.T) {
	input := `
var i: Integer;
var allInRange := true;
for i := 1 to 100 do begin
	var r := Random();
	if (r < 0.0) or (r >= 1.0) then
		allInRange := false;
end;
begin
	allInRange;
end
	`
	result := testEval(input)

	boolVal, ok := result.(*BooleanValue)
	if !ok {
		t.Fatalf("result is not *BooleanValue. got=%T (%+v)", result, result)
	}

	if !boolVal.Value {
		t.Errorf("Random() produced value outside [0, 1) range")
	}
}

// TestBuiltinRandom_ReturnType tests that Random() returns Float.
func TestBuiltinRandom_ReturnType(t *testing.T) {
	input := `
begin
	Random();
end
	`
	result := testEval(input)

	_, ok := result.(*FloatValue)
	if !ok {
		t.Fatalf("Random() did not return *FloatValue. got=%T (%+v)", result, result)
	}
}

// TestBuiltinRandom_Variation tests that Random() produces different values.
func TestBuiltinRandom_Variation(t *testing.T) {
	input := `
var r1 := Random();
var r2 := Random();
var r3 := Random();
begin
	(r1 <> r2) and (r2 <> r3);
end
	`
	result := testEval(input)

	boolVal, ok := result.(*BooleanValue)
	if !ok {
		t.Fatalf("result is not *BooleanValue. got=%T (%+v)", result, result)
	}

	if !boolVal.Value {
		t.Errorf("Random() produced identical consecutive values (very unlikely)")
	}
}

// TestBuiltinRandomize_BasicUsage tests Randomize() function.
// Randomize() seeds the random number generator
func TestBuiltinRandomize_BasicUsage(t *testing.T) {
	input := `
begin
	Randomize();
end
	`
	result := testEval(input)

	// Randomize returns nil/nothing
	_, ok := result.(*NilValue)
	if !ok {
		t.Fatalf("Randomize() did not return *NilValue. got=%T (%+v)", result, result)
	}
}

// TestBuiltinRandom_Errors tests Random() error cases.
func TestBuiltinRandom_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "Too many arguments",
			input: `
begin
	Random(5);
end
			`,
			expectedError: "Random() expects no arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			errVal, ok := result.(*ErrorValue)
			if !ok {
				t.Fatalf("expected error, got=%T (%+v)", result, result)
			}

			if !strings.Contains(errVal.Message, tt.expectedError) {
				t.Errorf("error message = %q, want to contain %q", errVal.Message, tt.expectedError)
			}
		})
	}
}

// TestBuiltinRandomize_Errors tests Randomize() error cases.
func TestBuiltinRandomize_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "Too many arguments",
			input: `
begin
	Randomize(5);
end
			`,
			expectedError: "Randomize() expects no arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			errVal, ok := result.(*ErrorValue)
			if !ok {
				t.Fatalf("expected error, got=%T (%+v)", result, result)
			}

			if !strings.Contains(errVal.Message, tt.expectedError) {
				t.Errorf("error message = %q, want to contain %q", errVal.Message, tt.expectedError)
			}
		})
	}
}

// TestBuiltinRandomInt_BasicUsage tests RandomInt() with range validation
func TestBuiltinRandomInt_BasicUsage(t *testing.T) {
	input := `
var i: Integer;
var allInRange := true;
for i := 1 to 100 do begin
	var r := RandomInt(10);
	if (r < 0) or (r >= 10) then
		allInRange := false;
end;
begin
	allInRange;
end
	`
	result := testEval(input)

	boolVal, ok := result.(*BooleanValue)
	if !ok {
		t.Fatalf("result is not *BooleanValue. got=%T (%+v)", result, result)
	}

	if !boolVal.Value {
		t.Errorf("RandomInt(10) produced value outside [0, 10) range")
	}
}

// TestBuiltinRandomInt_ReturnType tests that RandomInt() returns Integer.
func TestBuiltinRandomInt_ReturnType(t *testing.T) {
	input := `
begin
	RandomInt(10);
end
	`
	result := testEval(input)

	_, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("RandomInt() did not return *IntegerValue. got=%T (%+v)", result, result)
	}
}

// TestBuiltinRandomInt_Variation tests that RandomInt() produces different values.
func TestBuiltinRandomInt_Variation(t *testing.T) {
	input := `
var r1 := RandomInt(1000);
var r2 := RandomInt(1000);
var r3 := RandomInt(1000);
begin
	(r1 <> r2) or (r2 <> r3);
end
	`
	result := testEval(input)

	boolVal, ok := result.(*BooleanValue)
	if !ok {
		t.Fatalf("result is not *BooleanValue. got=%T (%+v)", result, result)
	}

	// With max=1000, it's extremely unlikely all three are the same
	if !boolVal.Value {
		t.Errorf("RandomInt() produced identical consecutive values (very unlikely)")
	}
}

// TestBuiltinRandomInt_MaxOne tests RandomInt(1) always returns 0.
func TestBuiltinRandomInt_MaxOne(t *testing.T) {
	input := `
var i: Integer;
var allZero := true;
for i := 1 to 10 do begin
	var r := RandomInt(1);
	if r <> 0 then
		allZero := false;
end;
begin
	allZero;
end
	`
	result := testEval(input)

	boolVal, ok := result.(*BooleanValue)
	if !ok {
		t.Fatalf("result is not *BooleanValue. got=%T (%+v)", result, result)
	}

	if !boolVal.Value {
		t.Errorf("RandomInt(1) should always return 0")
	}
}

// TestBuiltinRandomInt_Errors tests RandomInt() error cases.
func TestBuiltinRandomInt_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "No arguments",
			input: `
begin
	RandomInt();
end
			`,
			expectedError: "RandomInt() expects exactly 1 argument",
		},
		{
			name: "Too many arguments",
			input: `
begin
	RandomInt(10, 20);
end
			`,
			expectedError: "RandomInt() expects exactly 1 argument",
		},
		{
			name: "Max is zero",
			input: `
begin
	RandomInt(0);
end
			`,
			expectedError: "RandomInt() expects max > 0",
		},
		{
			name: "Max is negative",
			input: `
begin
	RandomInt(-5);
end
			`,
			expectedError: "RandomInt() expects max > 0",
		},
		{
			name: "String argument",
			input: `
begin
	RandomInt("hello");
end
			`,
			expectedError: "RandomInt() expects Integer",
		},
		{
			name: "Float argument",
			input: `
begin
	RandomInt(10.5);
end
			`,
			expectedError: "RandomInt() expects Integer",
		},
		{
			name: "Boolean argument",
			input: `
begin
	RandomInt(true);
end
			`,
			expectedError: "RandomInt() expects Integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			errVal, ok := result.(*ErrorValue)
			if !ok {
				t.Fatalf("expected error, got=%T (%+v)", result, result)
			}

			if !strings.Contains(errVal.Message, tt.expectedError) {
				t.Errorf("error message = %q, want to contain %q", errVal.Message, tt.expectedError)
			}
		})
	}
}
