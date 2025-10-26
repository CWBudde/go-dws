package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// Helper function to evaluate DWScript code and return the result
func testEvalClass(input string) Value {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		panic("Parser errors: " + strings.Join(p.Errors(), "\n"))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	return interp.Eval(program)
}

// Test 7.45: TestObjectCreation
func TestObjectCreation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // Expected object class name
	}{
		{
			name: "Simple class with no constructor",
			input: `
				type TPoint = class
					X: Integer;
					Y: Integer;
				end;

				var p: TPoint;
				p := TPoint.Create();
			`,
			expected: "TPoint",
		},
		{
			name: "Class with fields - no Self assignment yet",
			input: `
				type TPerson = class
					Name: String;
					Age: Integer;
				end;

				var person: TPerson;
				person := TPerson.Create();
			`,
			expected: "TPerson",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEvalClass(tt.input)

			if isError(result) {
				t.Fatalf("Evaluation error: %s", result.String())
			}

			obj, ok := AsObject(result)
			if !ok {
				t.Fatalf("Expected object, got %s", result.Type())
			}

			if obj.Class.Name != tt.expected {
				t.Errorf("Expected class name %s, got %s", tt.expected, obj.Class.Name)
			}
		})
	}
}

// Test 7.46: TestFieldAccess
func TestFieldAccess(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name: "Read field values",
			input: `
				type TPoint = class
					X: Integer;
					Y: Integer;

					function Create(x: Integer; y: Integer): TPoint;
					begin
						Self.X := x;
						Self.Y := y;
					end;

					function GetX(): Integer;
					begin
						Result := Self.X;
					end;
				end;

				var p: TPoint;
				p := TPoint.Create(10, 20);
				p.GetX();
			`,
			expected: int64(10),
		},
		{
			name: "Access field directly",
			input: `
				type TPoint = class
					X: Integer;
					Y: Integer;

					function Create(x: Integer; y: Integer): TPoint;
					begin
						Self.X := x;
						Self.Y := y;
					end;
				end;

				var p: TPoint;
				p := TPoint.Create(42, 99);
				p.X;
			`,
			expected: int64(42),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEvalClass(tt.input)

			if isError(result) {
				t.Fatalf("Evaluation error: %s", result.String())
			}

			switch expected := tt.expected.(type) {
			case int64:
				intVal, ok := result.(*IntegerValue)
				if !ok {
					t.Fatalf("Expected IntegerValue, got %s", result.Type())
				}
				if intVal.Value != expected {
					t.Errorf("Expected %d, got %d", expected, intVal.Value)
				}
			case string:
				strVal, ok := result.(*StringValue)
				if !ok {
					t.Fatalf("Expected StringValue, got %s", result.Type())
				}
				if strVal.Value != expected {
					t.Errorf("Expected %s, got %s", expected, strVal.Value)
				}
			}
		})
	}
}

// Test 7.47: TestMethodCalls
func TestMethodCalls(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name: "Method with no arguments",
			input: `
				type TCounter = class
					Count: Integer;

					function Create(): TCounter;
					begin
						Self.Count := 0;
					end;

					procedure Increment();
					begin
						Self.Count := Self.Count + 1;
					end;

					function GetCount(): Integer;
					begin
						Result := Self.Count;
					end;
				end;

				var c: TCounter;
				c := TCounter.Create();
				c.Increment();
				c.Increment();
				c.GetCount();
			`,
			expected: int64(2),
		},
		{
			name: "Method with arguments",
			input: `
				type TCalculator = class
					function Add(a: Integer; b: Integer): Integer;
					begin
						Result := a + b;
					end;
				end;

				var calc: TCalculator;
				calc := TCalculator.Create();
				calc.Add(5, 7);
			`,
			expected: int64(12),
		},
		{
			name: "Method accessing Self fields",
			input: `
				type TRectangle = class
					Width: Integer;
					Height: Integer;

					function Create(w: Integer; h: Integer): TRectangle;
					begin
						Self.Width := w;
						Self.Height := h;
					end;

					function Area(): Integer;
					begin
						Result := Self.Width * Self.Height;
					end;
				end;

				var rect: TRectangle;
				rect := TRectangle.Create(5, 10);
				rect.Area();
			`,
			expected: int64(50),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEvalClass(tt.input)

			if isError(result) {
				t.Fatalf("Evaluation error: %s", result.String())
			}

			switch expected := tt.expected.(type) {
			case int64:
				intVal, ok := result.(*IntegerValue)
				if !ok {
					t.Fatalf("Expected IntegerValue, got %s", result.Type())
				}
				if intVal.Value != expected {
					t.Errorf("Expected %d, got %d", expected, intVal.Value)
				}
			}
		})
	}
}

// Test 7.48: TestInheritance
func TestInheritance(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name: "Child inherits parent fields",
			input: `
				type TAnimal = class
					Name: String;

					function Create(n: String): TAnimal;
					begin
						Self.Name := n;
					end;

					function GetName(): String;
					begin
						Result := Self.Name;
					end;
				end;

				type TDog = class(TAnimal)
					Breed: String;

					function Create(n: String; b: String): TDog;
					begin
						Self.Name := n;
						Self.Breed := b;
					end;
				end;

				var dog: TDog;
				dog := TDog.Create('Buddy', 'Golden Retriever');
				dog.GetName();
			`,
			expected: "Buddy",
		},
		{
			name: "Child overrides parent method",
			input: `
				type TAnimal = class
					function Speak(): String;
					begin
						Result := 'Some sound';
					end;
				end;

				type TDog = class(TAnimal)
					function Speak(): String;
					begin
						Result := 'Woof!';
					end;
				end;

				var dog: TDog;
				dog := TDog.Create();
				dog.Speak();
			`,
			expected: "Woof!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEvalClass(tt.input)

			if isError(result) {
				t.Fatalf("Evaluation error: %s", result.String())
			}

			switch expected := tt.expected.(type) {
			case string:
				strVal, ok := result.(*StringValue)
				if !ok {
					t.Fatalf("Expected StringValue, got %s", result.Type())
				}
				if strVal.Value != expected {
					t.Errorf("Expected %s, got %s", expected, strVal.Value)
				}
			}
		})
	}
}

// Test 7.49: TestPolymorphism
func TestPolymorphism(t *testing.T) {
	input := `
		type TShape = class
			function Area(): Integer;
			begin
				Result := 0;
			end;
		end;

		type TSquare = class(TShape)
			Side: Integer;

			function Create(s: Integer): TSquare;
			begin
				Self.Side := s;
			end;

			function Area(): Integer;
			begin
				Result := Self.Side * Self.Side;
			end;
		end;

		var square: TSquare;
		square := TSquare.Create(5);
		square.Area();
	`

	result := testEvalClass(input)

	if isError(result) {
		t.Fatalf("Evaluation error: %s", result.String())
	}

	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %s", result.Type())
	}

	if intVal.Value != 25 {
		t.Errorf("Expected 25, got %d", intVal.Value)
	}
}

// Test 7.50: TestConstructors
func TestConstructors(t *testing.T) {
	input := `
		type TBox = class
			Width: Integer;
			Height: Integer;
			Depth: Integer;

			function Create(w: Integer; h: Integer; d: Integer): TBox;
			begin
				Self.Width := w;
				Self.Height := h;
				Self.Depth := d;
			end;

			function Volume(): Integer;
			begin
				Result := Self.Width * Self.Height * Self.Depth;
			end;
		end;

		var box: TBox;
		box := TBox.Create(2, 3, 4);
		box.Volume();
	`

	result := testEvalClass(input)

	if isError(result) {
		t.Fatalf("Evaluation error: %s", result.String())
	}

	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %s", result.Type())
	}

	if intVal.Value != 24 {
		t.Errorf("Expected 24, got %d", intVal.Value)
	}
}

// Test 7.51: TestSelfReference
func TestSelfReference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name: "Self.field access",
			input: `
				type TValue = class
					Val: Integer;

					function Create(v: Integer): TValue;
					begin
						Self.Val := v;
					end;

					function Double(): Integer;
					begin
						Result := Self.Val * 2;
					end;
				end;

				var v: TValue;
				v := TValue.Create(21);
				v.Double();
			`,
			expected: int64(42),
		},
		{
			name: "Self.method() calls",
			input: `
				type TChain = class
					Value: Integer;

					function Create(v: Integer): TChain;
					begin
						Self.Value := v;
					end;

					function GetValue(): Integer;
					begin
						Result := Self.Value;
					end;

					function DoubleValue(): Integer;
					begin
						Result := Self.GetValue() * 2;
					end;
				end;

				var chain: TChain;
				chain := TChain.Create(10);
				chain.DoubleValue();
			`,
			expected: int64(20),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEvalClass(tt.input)

			if isError(result) {
				t.Fatalf("Evaluation error: %s", result.String())
			}

			switch expected := tt.expected.(type) {
			case int64:
				intVal, ok := result.(*IntegerValue)
				if !ok {
					t.Fatalf("Expected IntegerValue, got %s", result.Type())
				}
				if intVal.Value != expected {
					t.Errorf("Expected %d, got %d", expected, intVal.Value)
				}
			}
		})
	}
}

// Test error cases
func TestClassErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "Parent class not found",
			input: `
				type TDog = class(TAnimal)
					Name: String;
				end;
			`,
			expectedError: "parent class 'TAnimal' not found",
		},
		{
			name: "Class not found during instantiation",
			input: `
				var obj: TUnknown;
				obj := TUnknown.Create();
			`,
			expectedError: "class 'TUnknown' not found",
		},
		{
			name: "Method not found",
			input: `
				type TSimple = class
					Value: Integer;
				end;

				var obj: TSimple;
				obj := TSimple.Create();
				obj.NonExistentMethod();
			`,
			expectedError: "method 'NonExistentMethod' not found",
		},
		{
			name: "Field not found",
			input: `
				type TSimple = class
					Value: Integer;
				end;

				var obj: TSimple;
				obj := TSimple.Create();
				obj.NonExistentField;
			`,
			expectedError: "field 'NonExistentField' not found",
		},
		{
			name: "Self used outside method",
			input: `
				type TSimple = class
					Value: Integer;
				end;

				var x: Integer;
				x := Self.Value;
			`,
			expectedError: "Self used outside method context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEvalClass(tt.input)

			if !isError(result) {
				t.Fatalf("Expected error, got %s", result.String())
			}

			errVal, ok := result.(*ErrorValue)
			if !ok {
				t.Fatalf("Expected ErrorValue, got %T", result)
			}

			if !strings.Contains(errVal.Message, tt.expectedError) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, errVal.Message)
			}
		})
	}
}

// ============================================================================

// ============================================================================
// Virtual/Override Method Tests (Task 7.64)
// ============================================================================

func TestVirtualMethodPolymorphism(t *testing.T) {
	input := `
		type TBase = class
			function GetValue(): Integer; virtual;
			begin
				Result := 1;
			end;
		end;

		type TChild = class(TBase)
			function GetValue(): Integer; override;
			begin
				Result := 2;
			end;
		end;

		var obj: TBase;
		begin
			obj := TChild.Create();
			PrintLn(obj.GetValue());
		end
	`

	_, output := testEvalWithOutput(input)
	expected := "2\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

func TestVirtualMethodThreeLevels(t *testing.T) {
	input := `
		type TBase = class
			function GetValue(): Integer; virtual;
			begin
				Result := 1;
			end;
		end;

		type TMiddle = class(TBase)
			function GetValue(): Integer; override;
			begin
				Result := 2;
			end;
		end;

		type TLeaf = class(TMiddle)
			function GetValue(): Integer; override;
			begin
				Result := 3;
			end;
		end;

		var obj: TBase;
		begin
			obj := TLeaf.Create();
			PrintLn(obj.GetValue());
		end
	`

	_, output := testEvalWithOutput(input)
	expected := "3\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

func TestNonVirtualMethodDynamicDispatch(t *testing.T) {
	input := `
		type TBase = class
			function GetValue(): Integer;
			begin
				Result := 1;
			end;
		end;

		type TChild = class(TBase)
			function GetValue(): Integer;
			begin
				Result := 2;
			end;
		end;

		var obj: TBase;
		begin
			obj := TChild.Create();
			PrintLn(obj.GetValue());
		end
	`

	// DWScript uses dynamic dispatch for all methods
	// The actual runtime type determines which method is called
	_, output := testEvalWithOutput(input)
	expected := "2\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}
