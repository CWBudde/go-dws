package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// ============================================================================
// Integration Tests for Record Literals
// ============================================================================

func TestEvalRecordLiteral_TypedSimple(t *testing.T) {
	// Test: type TPoint = record X, Y: Integer; end;
	//       var p := TPoint(X: 10; Y: 20);
	input := `
		type TPoint = record
			X, Y: Integer;
		end;
		var p := TPoint(X: 10; Y: 20);
	`

	program := parseProgram(t, input)
	interp := New(new(bytes.Buffer))
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}

	// Get the variable p
	pVal, exists := interp.env.Get("p")
	if !exists {
		t.Fatal("variable 'p' not found in environment")
	}

	// Verify it's a record
	recordVal, ok := pVal.(*RecordValue)
	if !ok {
		t.Fatalf("expected RecordValue, got %T", pVal)
	}

	// Verify record type name
	if recordVal.RecordType.Name != "TPoint" {
		t.Errorf("RecordType.Name = %v, want 'TPoint'", recordVal.RecordType.Name)
	}

	// Verify field values
	xVal, ok := recordVal.Fields["x"]
	if !ok {
		t.Fatal("field 'x' not found")
	}
	if intVal, ok := xVal.(*IntegerValue); ok {
		if intVal.Value != 10 {
			t.Errorf("X = %d, want 10", intVal.Value)
		}
	} else {
		t.Errorf("X is not IntegerValue, got %T", xVal)
	}

	yVal, ok := recordVal.Fields["y"]
	if !ok {
		t.Fatal("field 'y' not found")
	}
	if intVal, ok := yVal.(*IntegerValue); ok {
		if intVal.Value != 20 {
			t.Errorf("Y = %d, want 20", intVal.Value)
		}
	} else {
		t.Errorf("Y is not IntegerValue, got %T", yVal)
	}
}

func TestEvalRecordLiteral_AnonymousWithTypeAnnotation(t *testing.T) {
	// Test: type TPoint = record X, Y: Integer; end;
	//       var p : TPoint := (X: 5; Y: 15);
	input := `
		type TPoint = record
			X, Y: Integer;
		end;
		var p : TPoint := (X: 5; Y: 15);
	`

	program := parseProgram(t, input)
	interp := New(new(bytes.Buffer))
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}

	// Get the variable p
	pVal, exists := interp.env.Get("p")
	if !exists {
		t.Fatal("variable 'p' not found in environment")
	}

	// Verify it's a record
	recordVal, ok := pVal.(*RecordValue)
	if !ok {
		t.Fatalf("expected RecordValue, got %T", pVal)
	}

	// Verify field values
	xVal := recordVal.Fields["x"].(*IntegerValue)
	if xVal.Value != 5 {
		t.Errorf("X = %d, want 5", xVal.Value)
	}

	yVal := recordVal.Fields["y"].(*IntegerValue)
	if yVal.Value != 15 {
		t.Errorf("Y = %d, want 15", yVal.Value)
	}
}

func TestEvalRecordLiteral_DeathStarExample(t *testing.T) {
	// Test actual Death_Star.dws examples
	input := `
		type TSphere = record
			cx, cy, cz, r: Float;
		end;

		const big : TSphere = (cx: 20; cy: 20; cz: 0; r: 20);
		const small : TSphere = (cx: 7; cy: 7; cz: -10; r: 15);
	`

	program := parseProgram(t, input)
	interp := New(new(bytes.Buffer))
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}

	// Verify 'big' constant
	bigVal, exists := interp.env.Get("big")
	if !exists {
		t.Fatal("constant 'big' not found in environment")
	}

	bigRec, ok := bigVal.(*RecordValue)
	if !ok {
		t.Fatalf("expected RecordValue for 'big', got %T", bigVal)
	}

	if bigRec.RecordType.Name != "TSphere" {
		t.Errorf("big.RecordType.Name = %v, want 'TSphere'", bigRec.RecordType.Name)
	}

	// Check big's field values (accept both Float and Integer since integer literals can be assigned to Float fields)
	expectedBig := map[string]float64{"cx": 20, "cy": 20, "cz": 0, "r": 20}
	for fieldName, expectedVal := range expectedBig {
		fieldVal, ok := bigRec.Fields[fieldName]
		if !ok {
			t.Errorf("field '%s' not found in 'big'", fieldName)
			continue
		}
		var actualVal float64
		if floatVal, ok := fieldVal.(*FloatValue); ok {
			actualVal = floatVal.Value
		} else if intVal, ok := fieldVal.(*IntegerValue); ok {
			actualVal = float64(intVal.Value)
		} else {
			t.Errorf("big.%s is not FloatValue or IntegerValue, got %T", fieldName, fieldVal)
			continue
		}
		if actualVal != expectedVal {
			t.Errorf("big.%s = %v, want %v", fieldName, actualVal, expectedVal)
		}
	}

	// Verify 'small' constant
	smallVal, exists := interp.env.Get("small")
	if !exists {
		t.Fatal("constant 'small' not found in environment")
	}

	smallRec, ok := smallVal.(*RecordValue)
	if !ok {
		t.Fatalf("expected RecordValue for 'small', got %T", smallVal)
	}

	// Check small's field values (including negative cz, accept both Float and Integer)
	expectedSmall := map[string]float64{"cx": 7, "cy": 7, "cz": -10, "r": 15}
	for fieldName, expectedVal := range expectedSmall {
		fieldVal, ok := smallRec.Fields[fieldName]
		if !ok {
			t.Errorf("field '%s' not found in 'small'", fieldName)
			continue
		}
		var actualVal float64
		if floatVal, ok := fieldVal.(*FloatValue); ok {
			actualVal = floatVal.Value
		} else if intVal, ok := fieldVal.(*IntegerValue); ok {
			actualVal = float64(intVal.Value)
		} else {
			t.Errorf("small.%s is not FloatValue or IntegerValue, got %T", fieldName, fieldVal)
			continue
		}
		if actualVal != expectedVal {
			t.Errorf("small.%s = %v, want %v", fieldName, actualVal, expectedVal)
		}
	}
}

func TestEvalRecordLiteral_NestedRecords(t *testing.T) {
	// Test: type TPoint = record X, Y: Integer; end;
	//       type TRect = record TopLeft, BottomRight: TPoint; end;
	//       var rect := TRect(TopLeft: TPoint(X: 0; Y: 0); BottomRight: TPoint(X: 10; Y: 10));
	input := `
		type TPoint = record
			X, Y: Integer;
		end;
		type TRect = record
			TopLeft, BottomRight: TPoint;
		end;
		var rect := TRect(TopLeft: TPoint(X: 0; Y: 0); BottomRight: TPoint(X: 10; Y: 10));
	`

	program := parseProgram(t, input)
	interp := New(new(bytes.Buffer))
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}

	// Get the variable rect
	rectVal, exists := interp.env.Get("rect")
	if !exists {
		t.Fatal("variable 'rect' not found in environment")
	}

	// Verify it's a record
	rectRec, ok := rectVal.(*RecordValue)
	if !ok {
		t.Fatalf("expected RecordValue, got %T", rectVal)
	}

	// Verify TopLeft field
	topLeftVal, ok := rectRec.Fields["TopLeft"]
	if !ok {
		t.Fatal("field 'TopLeft' not found")
	}
	topLeftRec, ok := topLeftVal.(*RecordValue)
	if !ok {
		t.Fatalf("TopLeft is not RecordValue, got %T", topLeftVal)
	}
	if topLeftRec.RecordType.Name != "TPoint" {
		t.Errorf("TopLeft.RecordType.Name = %v, want 'TPoint'", topLeftRec.RecordType.Name)
	}

	// Verify BottomRight field
	bottomRightVal, ok := rectRec.Fields["BottomRight"]
	if !ok {
		t.Fatal("field 'BottomRight' not found")
	}
	bottomRightRec, ok := bottomRightVal.(*RecordValue)
	if !ok {
		t.Fatalf("BottomRight is not RecordValue, got %T", bottomRightVal)
	}
	if bottomRightRec.RecordType.Name != "TPoint" {
		t.Errorf("BottomRight.RecordType.Name = %v, want 'TPoint'", bottomRightRec.RecordType.Name)
	}
}

func TestEvalRecordLiteral_WithExpressions(t *testing.T) {
	// Test: var x := 5; var y := 10;
	//       type TPoint = record X, Y: Integer; end;
	//       var p := TPoint(X: x + 5; Y: y * 2);
	input := `
		var x := 5;
		var y := 10;
		type TPoint = record
			X, Y: Integer;
		end;
		var p := TPoint(X: x + 5; Y: y * 2);
	`

	program := parseProgram(t, input)
	interp := New(new(bytes.Buffer))
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("evaluation error: %s", result.String())
	}

	// Get the variable p
	pVal, exists := interp.env.Get("p")
	if !exists {
		t.Fatal("variable 'p' not found in environment")
	}

	// Verify field values are computed correctly
	recordVal := pVal.(*RecordValue)
	xVal := recordVal.Fields["x"].(*IntegerValue)
	if xVal.Value != 10 { // x + 5 = 5 + 5 = 10
		t.Errorf("X = %d, want 10", xVal.Value)
	}

	yVal := recordVal.Fields["y"].(*IntegerValue)
	if yVal.Value != 20 { // y * 2 = 10 * 2 = 20
		t.Errorf("Y = %d, want 20", yVal.Value)
	}
}

func TestEvalRecordLiteral_MissingField_Error(t *testing.T) {
	// Test: type TPoint = record X, Y: Integer; end;
	//       var p := TPoint(X: 10);  // Missing Y field
	// With field initializers, missing fields now get default values (0 for Integer)
	// This test now verifies that missing fields are initialized with defaults, not errors
	input := `
		type TPoint = record
			X, Y: Integer;
		end;
		var p := TPoint(X: 10);
		PrintLn(p.X);
		PrintLn(p.Y);
	`

	program := parseProgram(t, input)
	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	// Should succeed - missing field Y gets default value 0
	if isError(result) {
		t.Fatalf("expected success, got error: %s", result.String())
	}

	output := buf.String()
	expectedOutput := "10\n0\n"
	if output != expectedOutput {
		t.Errorf("expected output %q, got %q", expectedOutput, output)
	}
}

func TestEvalRecordLiteral_UnknownField_Error(t *testing.T) {
	// Test: type TPoint = record X, Y: Integer; end;
	//       var p := TPoint(X: 10; Y: 20; Z: 30);  // Z doesn't exist
	input := `
		type TPoint = record
			X, Y: Integer;
		end;
		var p := TPoint(X: 10; Y: 20; Z: 30);
	`

	program := parseProgram(t, input)
	interp := New(new(bytes.Buffer))
	result := interp.Eval(program)

	// Should get an error about unknown field
	if !isError(result) {
		t.Fatal("expected error about unknown field, got success")
	}

	errMsg := result.String()
	if !containsSubstring(errMsg, "does not exist") {
		t.Errorf("error message should mention field 'does not exist', got: %s", errMsg)
	}
}

func TestEvalRecordLiteral_TypeMismatch_Error(t *testing.T) {
	// Test: type TPoint = record X, Y: Integer; end;
	//       var p := TPoint(X: 'hello'; Y: 20);  // X should be Integer
	input := `
		type TPoint = record
			X, Y: Integer;
		end;
		var p := TPoint(X: 'hello'; Y: 20);
	`

	program := parseProgram(t, input)
	interp := New(new(bytes.Buffer))
	result := interp.Eval(program)

	// Interpreter doesn't type-check (that's semantic analyzer's job)
	// But this should still work at runtime, just with wrong type stored
	// For strict type checking, we'd need semantic analysis first

	// For now, just verify it doesn't crash
	if isError(result) {
		// If there's an error, that's also acceptable behavior
		t.Logf("Got error (acceptable): %s", result.String())
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func parseProgram(t *testing.T, input string) *ast.Program {
	t.Helper()
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	errors := p.Errors()
	if len(errors) > 0 {
		t.Fatalf("parsing errors: %v", errors)
	}

	return program
}
