package interp

import (
	"bytes"
	"testing"
)

// TestHelperPropertyExpressionAccessors covers helper properties whose accessor is
// an expression rather than a method name — `read (2*Field)`, `write (F := Value)`
// and the lvalue shorthand `write (F.P)` — for class helpers, record helpers, and
// both instance and class-name receivers. executeHelperPropertyRead/Write used to
// have no PropAccessExpression case at all, so every one of these failed with
// "property '...' has no read access" or silently swallowed the write.
func TestHelperPropertyExpressionAccessors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "instance helper, expression read on an instance",
			input: `
type TBase = class
	Field : Integer;
end;
type TBaseHelper = class helper for TBase
	property Doubled : Integer read (2*Field);
end;
var b := TBase.Create;
b.Field := 4;
PrintLn(b.Doubled);
`,
			expected: "8\n",
		},
		{
			name: "class helper, class property read through an instance",
			input: `
type TBase = class
	class var Field : Integer = 3;
end;
type TBaseHelper = class helper for TBase
	class property Doubled : Integer read (2*Field);
end;
var b := TBase.Create;
PrintLn(b.Doubled);
`,
			expected: "6\n",
		},
		{
			name: "class helper, class property read through the class name",
			input: `
type TBase = class
	class var Field : Integer = 3;
end;
type TBaseHelper = class helper for TBase
	class property Doubled : Integer read (2*Field);
end;
PrintLn(TBase.Doubled);
`,
			expected: "6\n",
		},
		{
			name: "class helper, class property write through the class name",
			input: `
type TBase = class
	class var Field : Integer = 1;
end;
type TBaseHelper = class helper for TBase
	class property Doubled : Integer read (2*Field) write (Field := Value div 2);
end;
TBase.Doubled := 10;
PrintLn(TBase.Field);
`,
			expected: "5\n",
		},
		{
			name: "class helper, class property write through an instance",
			input: `
type TBase = class
	class var Field : Integer = 1;
end;
type TBaseHelper = class helper for TBase
	class property Doubled : Integer read (2*Field) write (Field := Value div 2);
end;
var b := TBase.Create;
b.Doubled := 10;
PrintLn(TBase.Field);
`,
			expected: "5\n",
		},
		{
			name: "lvalue write specifier is shorthand for assigning Value",
			input: `
type TBase = class
	class var Field : Integer = 1;
	class property Half : Integer read (Field) write (Field := Value div 2);
end;
type TOuter = class
	class var FBase : TBase;
end;
type TOuterHelper = class helper for TOuter
	class property Forwarded : Integer read (FBase.Half) write (FBase.Half);
end;
var o := TOuter.Create;
o.FBase := TBase.Create;
o.Forwarded := 10;
PrintLn(TBase.Field);
`,
			expected: "5\n",
		},
		{
			name: "helper class property reached through a type cast binds the static class",
			input: `
type TBase = class
	class var Field : Integer = 4;
end;
type TBaseHelper = class helper for TBase
	class property Doubled : Integer read (2*Field);
end;
type TSub = class (TBase)
end;
var s := TSub.Create;
PrintLn(TBase(s).Doubled);
`,
			expected: "8\n",
		},
		{
			name: "record helper, class property read through an instance",
			input: `
type TRec = record
	Dummy : Integer;
	class var Field : Integer = 5;
end;
type TRecHelper = record helper for TRec
	class property Doubled : Integer read (2*Field);
end;
var r : TRec;
PrintLn(r.Doubled);
`,
			expected: "10\n",
		},
		{
			name: "record helper, class property write through an instance",
			input: `
type TRec = record
	Dummy : Integer;
	class var Field : Integer = 1;
end;
type TRecHelper = record helper for TRec
	class property Doubled : Integer read (2*Field) write (Field := Value div 2);
end;
var r : TRec;
r.Doubled := 10;
PrintLn(r.Doubled);
`,
			expected: "10\n",
		},
		{
			name: "record class var written through an instance hits shared storage",
			input: `
type TRec = record
	Dummy : Integer;
	class var Field : Integer = 1;
end;
var a : TRec;
var b : TRec;
a.Field := 7;
PrintLn(b.Field);
PrintLn(TRec.Field);
`,
			expected: "7\n7\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			interp := New(&out)
			if result := interpret(interp, tt.input); isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}
			if out.String() != tt.expected {
				t.Errorf("got %q, want %q", out.String(), tt.expected)
			}
		})
	}
}
