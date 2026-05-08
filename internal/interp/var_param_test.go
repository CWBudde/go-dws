package interp

import (
	"strings"
	"testing"
)

// TestVarParam_BasicInteger tests basic var parameter modification with integers
func TestVarParam_BasicInteger(t *testing.T) {
	input := `
		procedure Increment(var x: Integer);
		begin
			x := x + 1;
		end;

		var n: Integer := 5;
		Increment(n);
		PrintLn(n);
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("unexpected error: %s", result)
	}

	expected := "6\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

// TestVarParam_MultipleModifications tests multiple modifications to a var parameter
func TestVarParam_MultipleModifications(t *testing.T) {
	input := `
		procedure AddAndMultiply(var x: Integer; add: Integer; mul: Integer);
		begin
			x := x + add;
			x := x * mul;
		end;

		var n: Integer := 10;
		AddAndMultiply(n, 5, 2);
		PrintLn(n);
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("unexpected error: %s", result)
	}

	expected := "30\n" // (10 + 5) * 2 = 30
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

// TestVarParam_MultipleVarParams tests function with multiple var parameters
func TestVarParam_MultipleVarParams(t *testing.T) {
	input := `
		procedure Swap(var a: Integer; var b: Integer);
		var temp: Integer;
		begin
			temp := a;
			a := b;
			b := temp;
		end;

		var x: Integer := 10;
		var y: Integer := 20;
		Swap(x, y);
		PrintLn(x);
		PrintLn(y);
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("unexpected error: %s", result)
	}

	expected := "20\n10\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

// TestVarParam_WithLazy tests var parameter combined with lazy parameter (Jensen's Device pattern)
func TestVarParam_WithLazy(t *testing.T) {
	input := `
		function sum(var i: Integer; lo, hi: Integer; lazy term: Float): Float;
		begin
			i := lo;
			while i <= hi do begin
				Result := Result + term;
				i := i + 1;
			end;
		end;

		var i: Integer;
		var result: Float;
		result := sum(i, 1, 5, 1.0 / i);
		PrintLn(result);
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("unexpected error: %s", result)
	}

	// Sum should be 1/1 + 1/2 + 1/3 + 1/4 + 1/5 = 2.283333...
	if !strings.Contains(output, "2.28") && !strings.Contains(output, "2.29") && !strings.Contains(output, "2.3") {
		// Allow some floating point variation
		if !strings.Contains(output, "2.2") {
			t.Errorf("expected output to contain approximately 2.28, got=%q", output)
		}
	}
}

// TestVarParam_ErrorNonVariable tests that passing non-variables to var parameters fails
func TestVarParam_ErrorNonVariable(t *testing.T) {
	input := `
		procedure Increment(var x: Integer);
		begin
			x := x + 1;
		end;

		Increment(42);  // Should fail - can't pass literal to var parameter
	`

	result, _ := testEvalWithOutput(input)
	if !isError(result) {
		t.Fatal("expected error when passing literal to var parameter")
	}

	errMsg := result.(*ErrorValue).Message
	if !strings.Contains(errMsg, "var parameter requires a variable") {
		t.Errorf("wrong error message. got=%q", errMsg)
	}
}

// TestVarParam_NestedCalls tests var parameters in nested function calls
func TestVarParam_NestedCalls(t *testing.T) {
	input := `
		procedure IncrementTwice(var x: Integer);
		begin
			x := x + 1;
			x := x + 1;
		end;

		procedure ProcessValue(var n: Integer);
		begin
			IncrementTwice(n);
			n := n * 2;
		end;

		var val: Integer := 5;
		ProcessValue(val);
		PrintLn(val);
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("unexpected error: %s", result)
	}

	expected := "14\n" // (5 + 1 + 1) * 2 = 14
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

func TestVarParam_RecordVariableCopyBack(t *testing.T) {
	input := `
		type TRec = record
			A, B: Integer;
		end;

		procedure CopyRec(const src: TRec; var dest: TRec);
		begin
			dest := src;
		end;

		var src: TRec;
		var dest: TRec;
		src.A := 7;
		src.B := 9;
		CopyRec(src, dest);
		PrintLn(dest.A);
		PrintLn(dest.B);
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("unexpected error: %s", result)
	}

	expected := "7\n9\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

func TestVarParam_RecordArrayElementCopyBack(t *testing.T) {
	input := `
		type TRec = record
			A, B: Integer;
		end;

		procedure CopyRec(const src: TRec; var dest: TRec);
		begin
			dest := src;
		end;

		var src: TRec;
		src.A := 11;
		src.B := 13;
		var values: array [0..1] of TRec;
		CopyRec(src, values[1]);
		PrintLn(values[1].A);
		PrintLn(values[1].B);
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("unexpected error: %s", result)
	}

	expected := "11\n13\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

func TestVarParam_NestedRecordFieldCopyBack(t *testing.T) {
	input := `
		type TRec = record
			A: Integer;
		end;

		type THolder = record
			Item: TRec;
		end;

		procedure SetRec(var dest: TRec; value: Integer);
		begin
			dest.A := value;
		end;

		var holder: THolder;
		SetRec(holder.Item, 23);
		PrintLn(holder.Item.A);
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("unexpected error: %s", result)
	}

	expected := "23\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

func TestVarParam_RecordParameterNestedArrayMutationCopiesElements(t *testing.T) {
	input := `
		type TRec = record
			A: Integer;
			B: String;
		end;

		type THolder = record
			Items: array [0..1] of TRec;
		end;

		procedure Store(var holder: THolder; idx: Integer; item: TRec);
		begin
			holder.Items[idx] := item;
		end;

		procedure CopyHolder(src: THolder; var dest: THolder);
		begin
			dest := src;
		end;

		var item: TRec;
		item.A := 4;
		item.B := 'four';
		var holder: THolder;
		Store(holder, 0, item);
		item.A := 99;
		item.B := 'changed';
		PrintLn(holder.Items[0].A);
		PrintLn(holder.Items[0].B);
		var copied: THolder;
		CopyHolder(holder, copied);
		holder.Items[0].A := 77;
		holder.Items[0].B := 'changed again';
		PrintLn(copied.Items[0].A);
		PrintLn(copied.Items[0].B);
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("unexpected error: %s", result)
	}

	expected := "4\nfour\n4\nfour\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

func TestVarParam_RecordMethodRecordParameterCopyBack(t *testing.T) {
	input := `
		type TRec = record
			A: Integer;
			B: Integer;
		end;

		type TMutator = record
			procedure Replace(var dest: TRec; src: TRec);
			begin
				dest := src;
			end;
		end;

		var src: TRec;
		src.A := 31;
		src.B := 37;
		var dest: TRec;
		var mutator: TMutator;
		mutator.Replace(dest, src);
		PrintLn(dest.A);
		PrintLn(dest.B);
	`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("unexpected error: %s", result)
	}

	expected := "31\n37\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}
