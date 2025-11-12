package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// Helper function to evaluate class var initialization tests with output capture
func testEvalClassVarInit(input string) (Value, string) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		panic("Parser errors: " + joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)
	return result, buf.String()
}

func TestClassVarInitialization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "class var with integer initialization",
			input: `
type TBase = class
	class var Counter: Integer := 42;
end;

PrintLn(IntToStr(TBase.Counter));
`,
			expected: "42\n",
		},
		{
			name: "class var with string initialization",
			input: `
type TBase = class
	class var Name: String := 'Hello';
end;

PrintLn(TBase.Name);
`,
			expected: "Hello\n",
		},
		// TODO: Support expressions with class constants
		// {
		// 	name: "class var with expression initialization",
		// 	input: `
		// type TBase = class
		// 	class const C = 10;
		// 	class var Value: Integer := 5 + C;
		// end;
		//
		// PrintLn(IntToStr(TBase.Value));
		// `,
		// 	expected: "15\n",
		// },
		{
			name: "class var with type inference",
			input: `
type TBase = class
	class var Test := 123;
end;

PrintLn(IntToStr(TBase.Test));
`,
			expected: "123\n",
		},
		{
			name: "class var with float initialization",
			input: `
type TBase = class
	class var Pi: Float := 3.14;
end;

PrintLn(FloatToStr(TBase.Pi));
`,
			expected: "3.14\n",
		},
		{
			name: "class var with boolean initialization",
			input: `
type TBase = class
	class var Enabled := True;
end;

if TBase.Enabled then
	PrintLn('Yes')
else
	PrintLn('No');
`,
			expected: "Yes\n",
		},
		{
			name: "multiple class vars with initialization",
			input: `
type TBase = class
	class var X: Integer := 10;
	class var Y: Integer := 20;
	class var Sum: Integer := 30;
end;

PrintLn(IntToStr(TBase.X));
PrintLn(IntToStr(TBase.Y));
PrintLn(IntToStr(TBase.Sum));
`,
			expected: "10\n20\n30\n",
		},
		{
			name: "class var with negative number",
			input: `
type TBase = class
	class var Offset: Integer := -10;
end;

PrintLn(IntToStr(TBase.Offset));
`,
			expected: "-10\n",
		},
		{
			name: "class var modification after initialization",
			input: `
type TBase = class
	class var Counter: Integer := 100;
end;

PrintLn(IntToStr(TBase.Counter));
TBase.Counter := 200;
PrintLn(IntToStr(TBase.Counter));
`,
			expected: "100\n200\n",
		},
		{
			name: "inherited class with own class var initialization",
			input: `
type TBase = class
	class var BaseVar: Integer := 10;
end;

type TChild = class(TBase)
	class var ChildVar: Integer := 20;
end;

PrintLn(IntToStr(TBase.BaseVar));
PrintLn(IntToStr(TChild.ChildVar));
PrintLn(IntToStr(TChild.BaseVar));
`,
			expected: "10\n20\n10\n",
		},
		{
			name: "class var used in constructor",
			input: `
type TBase = class
	Field: Integer;
	class var DefaultValue: Integer := 42;
	constructor Create;
	begin
		Field := DefaultValue;
	end;
end;

var obj: TBase;
obj := TBase.Create;
PrintLn(IntToStr(obj.Field));
`,
			expected: "42\n",
		},
		{
			name: "class var accessed from class method",
			input: `
type TBase = class
	class var Counter: Integer := 5;
	class procedure PrintCounter;
	begin
		PrintLn(IntToStr(Counter));
	end;
end;

TBase.PrintCounter;
`,
			expected: "5\n",
		},
		{
			name: "class var with complex expression",
			input: `
type TBase = class
	class const A = 10;
	class const B = 5;
	class var Result: Integer := (A + B) * 2;
end;

PrintLn(IntToStr(TBase.Result));
`,
			expected: "30\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output := testEvalClassVarInit(tt.input)
			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}
			if output != tt.expected {
				t.Errorf("wrong output. expected=%q, got=%q", tt.expected, output)
			}
		})
	}
}

