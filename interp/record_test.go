package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/lexer"
	"github.com/cwbudde/go-dws/parser"
	"github.com/cwbudde/go-dws/types"
)

// ============================================================================
// Task 8.73: RecordValue Creation Tests
// ============================================================================

func TestRecordValueCreation(t *testing.T) {
	t.Run("Create simple record value", func(t *testing.T) {
		// Create a TPoint record type with X and Y fields
		fields := map[string]types.Type{
			"X": types.INTEGER,
			"Y": types.INTEGER,
		}
		recordType := types.NewRecordType("TPoint", fields)

		// Create a record value
		recordVal := NewRecordValue(recordType)

		// Verify the value implements Value interface
		if recordVal == nil {
			t.Fatal("NewRecordValue returned nil")
		}

		// Verify type
		if recordVal.Type() != "RECORD" {
			t.Errorf("Type() = %v, want RECORD", recordVal.Type())
		}

		// Verify the record type is stored
		if rv, ok := recordVal.(*RecordValue); ok {
			if rv.RecordType != recordType {
				t.Errorf("RecordType mismatch")
			}
			if rv.Fields == nil {
				t.Error("Fields map should be initialized")
			}
		} else {
			t.Error("Value is not a *RecordValue")
		}
	})

	t.Run("Record value string representation", func(t *testing.T) {
		fields := map[string]types.Type{
			"X": types.INTEGER,
			"Y": types.INTEGER,
		}
		recordType := types.NewRecordType("TPoint", fields)
		recordVal := NewRecordValue(recordType)

		// Set some field values
		if rv, ok := recordVal.(*RecordValue); ok {
			rv.Fields["X"] = &IntegerValue{Value: 10}
			rv.Fields["Y"] = &IntegerValue{Value: 20}
		}

		// String should show record type and fields
		str := recordVal.String()
		if str == "" {
			t.Error("String() should not be empty")
		}
		// We don't enforce exact format, just check it's not empty
	})
}

// ============================================================================
// Task 8.74: Record Literal Evaluation Tests
// ============================================================================

func TestEvalRecordLiteral(t *testing.T) {
	t.Run("Named field initialization", func(t *testing.T) {
		// DWScript: var p: TPoint; p := (X: 10, Y: 20);
		input := `
			type TPoint = record
				X: Integer;
				Y: Integer;
			end;

			var p: TPoint;
			p := (X: 10, Y: 20);
		`

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Fatalf("Parser errors: %v", p.Errors())
		}

		var buf bytes.Buffer
		interp := New(&buf)
		result := interp.Eval(program)

		// Should not be an error
		if isError(result) {
			t.Fatalf("Eval error: %v", result)
		}

		// Get the variable 'p'
		pVal, ok := interp.env.Get("p")
		if !ok {
			t.Fatal("Variable 'p' not found")
		}

		// Should be a RecordValue
		recordVal, ok := pVal.(*RecordValue)
		if !ok {
			t.Fatalf("Expected RecordValue, got %T", pVal)
		}

		// Check field values
		xVal, ok := recordVal.Fields["X"]
		if !ok {
			t.Fatal("Field 'X' not found")
		}
		if intVal, ok := xVal.(*IntegerValue); ok {
			if intVal.Value != 10 {
				t.Errorf("X = %d, want 10", intVal.Value)
			}
		} else {
			t.Errorf("X is not an IntegerValue, got %T", xVal)
		}

		yVal, ok := recordVal.Fields["Y"]
		if !ok {
			t.Fatal("Field 'Y' not found")
		}
		if intVal, ok := yVal.(*IntegerValue); ok {
			if intVal.Value != 20 {
				t.Errorf("Y = %d, want 20", intVal.Value)
			}
		} else {
			t.Errorf("Y is not an IntegerValue, got %T", yVal)
		}
	})

	t.Run("Expression in record field", func(t *testing.T) {
		// DWScript: var p: TPoint; p := (X: 5 + 5, Y: 10 * 2);
		input := `
			type TPoint = record
				X: Integer;
				Y: Integer;
			end;

			var p: TPoint;
			p := (X: 5 + 5, Y: 10 * 2);
		`

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Fatalf("Parser errors: %v", p.Errors())
		}

		var buf bytes.Buffer
		interp := New(&buf)
		result := interp.Eval(program)

		if isError(result) {
			t.Fatalf("Eval error: %v", result)
		}

		pVal, _ := interp.env.Get("p")
		recordVal := pVal.(*RecordValue)

		xVal := recordVal.Fields["X"].(*IntegerValue)
		if xVal.Value != 10 {
			t.Errorf("X = %d, want 10", xVal.Value)
		}

		yVal := recordVal.Fields["Y"].(*IntegerValue)
		if yVal.Value != 20 {
			t.Errorf("Y = %d, want 20", yVal.Value)
		}
	})
}

// ============================================================================
// Task 8.75: Record Field Access (Read) Tests
// ============================================================================

func TestRecordFieldAccess(t *testing.T) {
	t.Run("Read record field", func(t *testing.T) {
		// DWScript: var p: TPoint; p := (X: 10, Y: 20); var x: Integer; x := p.X;
		input := `
			type TPoint = record
				X: Integer;
				Y: Integer;
			end;

			var p: TPoint;
			p := (X: 10, Y: 20);
			var x: Integer;
			x := p.X;
		`

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Fatalf("Parser errors: %v", p.Errors())
		}

		var buf bytes.Buffer
		interp := New(&buf)
		result := interp.Eval(program)

		if isError(result) {
			t.Fatalf("Eval error: %v", result)
		}

		// Get the variable 'x'
		xVal, ok := interp.env.Get("x")
		if !ok {
			t.Fatal("Variable 'x' not found")
		}

		// Should be an IntegerValue with value 10
		if intVal, ok := xVal.(*IntegerValue); ok {
			if intVal.Value != 10 {
				t.Errorf("x = %d, want 10", intVal.Value)
			}
		} else {
			t.Errorf("x is not an IntegerValue, got %T", xVal)
		}
	})

	t.Run("Read multiple record fields", func(t *testing.T) {
		input := `
			type TPoint = record
				X: Integer;
				Y: Integer;
			end;

			var p: TPoint;
			p := (X: 5, Y: 10);
			var sum: Integer;
			sum := p.X + p.Y;
		`

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Fatalf("Parser errors: %v", p.Errors())
		}

		var buf bytes.Buffer
		interp := New(&buf)
		result := interp.Eval(program)

		if isError(result) {
			t.Fatalf("Eval error: %v", result)
		}

		sumVal, _ := interp.env.Get("sum")
		if intVal, ok := sumVal.(*IntegerValue); ok {
			if intVal.Value != 15 {
				t.Errorf("sum = %d, want 15", intVal.Value)
			}
		} else {
			t.Errorf("sum is not an IntegerValue, got %T", sumVal)
		}
	})
}

// ============================================================================
// Task 8.76: Record Field Assignment (Write) Tests
// ============================================================================

func TestRecordFieldAssignment(t *testing.T) {
	t.Run("Assign to record field", func(t *testing.T) {
		// DWScript: var p: TPoint; p := (X: 10, Y: 20); p.X := 30;
		input := `
			type TPoint = record
				X: Integer;
				Y: Integer;
			end;

			var p: TPoint;
			p := (X: 10, Y: 20);
			p.X := 30;
		`

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Fatalf("Parser errors: %v", p.Errors())
		}

		var buf bytes.Buffer
		interp := New(&buf)
		result := interp.Eval(program)

		if isError(result) {
			t.Fatalf("Eval error: %v", result)
		}

		// Get the record and check X was updated
		pVal, ok := interp.env.Get("p")
		if !ok {
			t.Fatal("Variable 'p' not found")
		}

		recordVal, ok := pVal.(*RecordValue)
		if !ok {
			t.Fatalf("Expected RecordValue, got %T", pVal)
		}

		xVal := recordVal.Fields["X"].(*IntegerValue)
		if xVal.Value != 30 {
			t.Errorf("p.X = %d, want 30", xVal.Value)
		}

		yVal := recordVal.Fields["Y"].(*IntegerValue)
		if yVal.Value != 20 {
			t.Errorf("p.Y = %d, want 20 (should be unchanged)", yVal.Value)
		}
	})

	t.Run("Assign expression to record field", func(t *testing.T) {
		input := `
			type TPoint = record
				X: Integer;
				Y: Integer;
			end;

			var p: TPoint;
			p := (X: 5, Y: 10);
			p.X := p.X * 2;
			p.Y := p.Y + 5;
		`

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Fatalf("Parser errors: %v", p.Errors())
		}

		var buf bytes.Buffer
		interp := New(&buf)
		result := interp.Eval(program)

		if isError(result) {
			t.Fatalf("Eval error: %v", result)
		}

		pVal, _ := interp.env.Get("p")
		recordVal := pVal.(*RecordValue)

		xVal := recordVal.Fields["X"].(*IntegerValue)
		if xVal.Value != 10 {
			t.Errorf("p.X = %d, want 10", xVal.Value)
		}

		yVal := recordVal.Fields["Y"].(*IntegerValue)
		if yVal.Value != 15 {
			t.Errorf("p.Y = %d, want 15", yVal.Value)
		}
	})
}

// ============================================================================
// Task 8.77: Record Copying (Value Semantics) Tests
// ============================================================================

func TestRecordCopying(t *testing.T) {
	t.Run("Record assignment creates copy (value semantics)", func(t *testing.T) {
		// DWScript: var p1, p2: TPoint; p1 := (X: 10, Y: 20); p2 := p1; p2.X := 99;
		// After this, p1.X should still be 10 (not 99), proving value semantics
		input := `
			type TPoint = record
				X: Integer;
				Y: Integer;
			end;

			var p1: TPoint;
			var p2: TPoint;
			p1 := (X: 10, Y: 20);
			p2 := p1;
			p2.X := 99;
		`

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Fatalf("Parser errors: %v", p.Errors())
		}

		var buf bytes.Buffer
		interp := New(&buf)
		result := interp.Eval(program)

		if isError(result) {
			t.Fatalf("Eval error: %v", result)
		}

		// Get p1 - should be unchanged
		p1Val, _ := interp.env.Get("p1")
		p1Record := p1Val.(*RecordValue)
		p1X := p1Record.Fields["X"].(*IntegerValue)
		if p1X.Value != 10 {
			t.Errorf("p1.X = %d, want 10 (should not be affected by p2 modification)", p1X.Value)
		}

		// Get p2 - should have modified value
		p2Val, _ := interp.env.Get("p2")
		p2Record := p2Val.(*RecordValue)
		p2X := p2Record.Fields["X"].(*IntegerValue)
		if p2X.Value != 99 {
			t.Errorf("p2.X = %d, want 99", p2X.Value)
		}
	})

	t.Run("Nested record copying", func(t *testing.T) {
		// Test that copying works correctly when records contain other records
		// For now, we'll test with basic fields
		input := `
			type TPoint = record
				X: Integer;
				Y: Integer;
			end;

			var p1: TPoint;
			p1 := (X: 1, Y: 2);
			var p2: TPoint;
			p2 := p1;
			var p3: TPoint;
			p3 := p2;

			p1.X := 100;
			p2.Y := 200;
		`

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Fatalf("Parser errors: %v", p.Errors())
		}

		var buf bytes.Buffer
		interp := New(&buf)
		result := interp.Eval(program)

		if isError(result) {
			t.Fatalf("Eval error: %v", result)
		}

		// p3 should be unaffected by changes to p1 and p2
		p3Val, _ := interp.env.Get("p3")
		p3Record := p3Val.(*RecordValue)
		p3X := p3Record.Fields["X"].(*IntegerValue)
		p3Y := p3Record.Fields["Y"].(*IntegerValue)

		if p3X.Value != 1 {
			t.Errorf("p3.X = %d, want 1 (original value)", p3X.Value)
		}
		if p3Y.Value != 2 {
			t.Errorf("p3.Y = %d, want 2 (original value)", p3Y.Value)
		}
	})
}

// ============================================================================
// Task 8.79: Record Comparison Tests
// ============================================================================

func TestRecordComparison(t *testing.T) {
	t.Run("Equal records with = operator", func(t *testing.T) {
		input := `
			type TPoint = record
				X: Integer;
				Y: Integer;
			end;

			var p1: TPoint;
			var p2: TPoint;
			p1 := (X: 10, Y: 20);
			p2 := (X: 10, Y: 20);
			var result: Boolean;
			result := p1 = p2;
		`

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Fatalf("Parser errors: %v", p.Errors())
		}

		var buf bytes.Buffer
		interp := New(&buf)
		result := interp.Eval(program)

		if isError(result) {
			t.Fatalf("Eval error: %v", result)
		}

		resultVal, _ := interp.env.Get("result")
		boolVal, ok := resultVal.(*BooleanValue)
		if !ok {
			t.Fatalf("result is not a BooleanValue, got %T", resultVal)
		}

		if !boolVal.Value {
			t.Error("p1 = p2 should be true (records with same field values)")
		}
	})

	t.Run("Unequal records with <> operator", func(t *testing.T) {
		input := `
			type TPoint = record
				X: Integer;
				Y: Integer;
			end;

			var p1: TPoint;
			var p2: TPoint;
			p1 := (X: 10, Y: 20);
			p2 := (X: 10, Y: 99);
			var result: Boolean;
			result := p1 <> p2;
		`

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Fatalf("Parser errors: %v", p.Errors())
		}

		var buf bytes.Buffer
		interp := New(&buf)
		result := interp.Eval(program)

		if isError(result) {
			t.Fatalf("Eval error: %v", result)
		}

		resultVal, _ := interp.env.Get("result")
		boolVal := resultVal.(*BooleanValue)

		if !boolVal.Value {
			t.Error("p1 <> p2 should be true (records with different Y values)")
		}
	})

	t.Run("Copied records are equal", func(t *testing.T) {
		input := `
			type TPoint = record
				X: Integer;
				Y: Integer;
			end;

			var p1: TPoint;
			var p2: TPoint;
			p1 := (X: 5, Y: 10);
			p2 := p1;
			var result: Boolean;
			result := p1 = p2;
		`

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Fatalf("Parser errors: %v", p.Errors())
		}

		var buf bytes.Buffer
		interp := New(&buf)
		result := interp.Eval(program)

		if isError(result) {
			t.Fatalf("Eval error: %v", result)
		}

		resultVal, _ := interp.env.Get("result")
		boolVal := resultVal.(*BooleanValue)

		if !boolVal.Value {
			t.Error("p1 = p2 should be true (p2 is a copy of p1 with same values)")
		}
	})
}
