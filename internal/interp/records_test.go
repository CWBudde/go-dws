package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestRecordMethodCall tests calling methods on records
func TestRecordMethodCall(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "record method with return value",
			input: `
type TPoint = record
	X, Y: Integer;
	function Distance: Float;
	begin
		Result := Sqrt(X*X + Y*Y);
	end;
end;

var p: TPoint;
p.X := 3;
p.Y := 4;
PrintLn(p.Distance());
`,
			expected: "5\n",
		},
		{
			name: "record method accessing Self fields",
			input: `
type TCounter = record
	Value: Integer;
	function GetValue: Integer;
	begin
		Result := Value;
	end;
end;

var c: TCounter;
c.Value := 42;
PrintLn(c.GetValue());
`,
			expected: "42\n",
		},
		{
			name: "record method with parameters",
			input: `
type TPoint = record
	X, Y: Integer;
	function DistanceFrom(otherX, otherY: Integer): Float;
	begin
		var dx := X - otherX;
		var dy := Y - otherY;
		Result := Sqrt(dx*dx + dy*dy);
	end;
end;

var p: TPoint;
p.X := 3;
p.Y := 4;
PrintLn(p.DistanceFrom(0, 0));
`,
			expected: "5\n",
		},
		{
			name: "record procedure (no return value)",
			input: `
type TPoint = record
	X, Y: Integer;
	procedure SetCoords(newX, newY: Integer);
	begin
		X := newX;
		Y := newY;
	end;
	function GetX: Integer;
	begin
		Result := X;
	end;
end;

var p: TPoint;
p.SetCoords(10, 20);
PrintLn(p.GetX());
`,
			expected: "10\n",
		},
		{
			name: "multiple method calls",
			input: `
type TCounter = record
	Value: Integer;
	function Increment: Integer;
	begin
		Value := Value + 1;
		Result := Value;
	end;
	function GetValue: Integer;
	begin
		Result := Value;
	end;
end;

var c: TCounter;
c.Value := 0;
PrintLn(c.Increment());
PrintLn(c.Increment());
PrintLn(c.GetValue());
`,
			expected: "1\n2\n2\n",
		},
		{
			name: "nested record with methods",
			input: `
type TInner = record
	Value: Integer;
	function Double: Integer;
	begin
		Result := Value * 2;
	end;
end;

type TOuter = record
	Inner: TInner;
	function GetInnerDouble: Integer;
	begin
		Result := Inner.Double();
	end;
end;

var outer: TOuter;
outer.Inner.Value := 5;
PrintLn(outer.GetInnerDouble());
`,
			expected: "10\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}
			interp := New(&out)
			result := interp.Eval(program)
			if exc := interp.GetException(); exc != nil {
				t.Fatalf("unexpected exception: %v", exc.Message)
			}
			if err, ok := result.(*ErrorValue); ok {
				t.Fatalf("unexpected error: %v", err.Message)
			}
			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestRecordMethodErrors tests error cases for record method calls
func TestRecordMethodErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "method not found",
			input: `
type TPoint = record
	X, Y: Integer;
end;

var p: TPoint;
p.DoesNotExist();
`,
			expectedErr: "method 'DoesNotExist' not found",
		},
		{
			name: "wrong number of arguments",
			input: `
type TPoint = record
	X, Y: Integer;
	function Add(a, b: Integer): Integer;
	begin
		Result := a + b;
	end;
end;

var p: TPoint;
p.Add(1);
`,
			expectedErr: "wrong number of arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				// Check if parser error matches expected error
				for _, err := range p.Errors() {
					if contains(err, tt.expectedErr) {
						return // Test passed
					}
				}
				t.Fatalf("parser errors but none match expected: %v", p.Errors())
			}
			interp := New(&out)
			result := interp.Eval(program)

			if !isError(result) {
				t.Fatalf("expected error containing %q, got %s", tt.expectedErr, result.String())
			}

			errVal, ok := result.(*ErrorValue)
			if !ok {
				t.Fatalf("expected ErrorValue, got %T", result)
			}

			if !contains(errVal.Message, tt.expectedErr) {
				t.Errorf("expected error containing %q, got %q", tt.expectedErr, errVal.Message)
			}
		})
	}
}

// TestRecordMethodSelfBinding tests that Self is properly bound in record methods
func TestRecordMethodSelfBinding(t *testing.T) {
	input := `
type TTest = record
	Value: Integer;
	function AccessSelf: Integer;
	begin
		// Should be able to access fields through Self
		Result := Self.Value;
	end;
end;

var t: TTest;
t.Value := 100;
PrintLn(t.AccessSelf());
`

	var out bytes.Buffer
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	interp := New(&out)
	interp.Eval(program)
	if exc := interp.GetException(); exc != nil {
		t.Fatalf("unexpected exception: %v", exc.Message)
	}
	expected := "100\n"
	if out.String() != expected {
		t.Errorf("expected output %q, got %q", expected, out.String())
	}
}

// TestStaticRecordMethods tests static record methods (class function/procedure)
func TestStaticRecordMethods(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "static function returning record",
			input: `
type TPoint = record
	X, Y: Integer;

	class function Origin: TPoint;
	begin
		Result.X := 0;
		Result.Y := 0;
	end;
end;

var p: TPoint;
p := TPoint.Origin();
PrintLn(p.X);
PrintLn(p.Y);
`,
			expected: "0\n0\n",
		},
		{
			name: "static procedure",
			input: `
type TCounter = record
	Value: Integer;

	class procedure PrintInfo;
	begin
		PrintLn('TCounter record');
	end;
end;

TCounter.PrintInfo();
`,
			expected: "TCounter record\n",
		},
		{
			name: "mix of static and instance methods",
			input: `
type TPoint = record
	X, Y: Integer;

	class function Origin: TPoint;
	begin
		Result.X := 0;
		Result.Y := 0;
	end;

	function Sum: Integer;
	begin
		Result := Self.X + Self.Y;
	end;
end;

var p: TPoint;
p := TPoint.Origin();
PrintLn(p.Sum());
p.X := 3;
p.Y := 7;
PrintLn(p.Sum());
`,
			expected: "0\n10\n",
		},
		{
			name: "static function returning integer",
			input: `
type TMath = record
	class function Add(a, b: Integer): Integer;
	begin
		Result := a + b;
	end;
end;

PrintLn(TMath.Add(5, 7));
`,
			expected: "12\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}
			interp := New(&out)
			result := interp.Eval(program)
			if isError(result) {
				t.Fatalf("interpreter error: %v", result.String())
			}
			if exc := interp.GetException(); exc != nil {
				t.Fatalf("unexpected exception: %v", exc.Message)
			}
			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestStaticRecordMethodErrors tests error cases for static record methods
func TestStaticRecordMethodErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "static method not found",
			input: `
type TPoint = record
	X, Y: Integer;
end;

TPoint.DoesNotExist();
`,
			expectedErr: "static method 'DoesNotExist' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				// Check if parser error matches expected error
				for _, err := range p.Errors() {
					if contains(err, tt.expectedErr) {
						return // Test passed
					}
				}
				t.Fatalf("parser errors but none match expected: %v", p.Errors())
			}
			interp := New(&out)
			result := interp.Eval(program)

			if !isError(result) {
				t.Fatalf("expected error containing %q, got %s", tt.expectedErr, result.String())
			}

			errVal, ok := result.(*ErrorValue)
			if !ok {
				t.Fatalf("expected ErrorValue, got %T", result)
			}

			if !contains(errVal.Message, tt.expectedErr) {
				t.Errorf("expected error containing %q, got %q", tt.expectedErr, errVal.Message)
			}
		})
	}
}
