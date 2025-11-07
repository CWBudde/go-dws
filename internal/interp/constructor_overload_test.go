package interp

import (
	"testing"
)

// TestConstructorOverload tests Task 9.68 - constructor overload dispatch
func TestConstructorOverload(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "exact fixture test case",
			input: `
				type tobj = class
					field : Integer;

					constructor create(a : Integer); overload;
					begin
						field:=a;
					end;
				end;

				var o := TObj.Create;
				PrintLn(o.field);
				o := TObj.Create(1);
				PrintLn(o.field);
				o := new TObj;
				PrintLn(o.field);
				o := new TObj(2);
				PrintLn(o.field);
			`,
			expected: "0\n1\n0\n2\n",
		},
		{
			name: "constructor with parameter",
			input: `
				type TObj = class
					field: Integer;
					constructor Create(a: Integer); overload;
					begin
						field := a;
					end;
				end;

				var o := TObj.Create(42);
				PrintLn(o.field);
			`,
			expected: "42\n",
		},
		{
			name: "implicit parameterless constructor",
			input: `
				type TObj = class
					field: Integer;
					constructor Create(a: Integer); overload;
					begin
						field := a;
					end;
				end;

				var o := TObj.Create;
				PrintLn(o.field);
			`,
			expected: "0\n", // Default value
		},
		{
			name: "new with implicit parameterless constructor",
			input: `
				type TObj = class
					field: Integer;
					constructor Create(a: Integer); overload;
					begin
						field := a;
					end;
				end;

				var o := new TObj;
				PrintLn(o.field);
			`,
			expected: "0\n", // Default value
		},
		{
			name: "new with parameter",
			input: `
				type TObj = class
					field: Integer;
					constructor Create(a: Integer); overload;
					begin
						field := a;
					end;
				end;

				var o := new TObj(99);
				PrintLn(o.field);
			`,
			expected: "99\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output := testEvalWithOutput(tt.input)
			// Check if there was an error
			if errVal, ok := result.(*ErrorValue); ok {
				t.Fatalf("evaluation error: %s", errVal.Message)
			}
			if output != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, output)
			}
		})
	}
}
