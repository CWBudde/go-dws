package interp

import (
	"bytes"
	"testing"
)

// TestAssociativeArrayARC pins the destructor timing of an
// `array [TObject] of TObject`: overwriting a slot destroys the displaced
// value, Delete destroys the removed value (the key survives while a variable
// still holds it), Clear destroys the remaining values and object keys, and
// program end destroys whatever the map still owns.
func TestAssociativeArrayARC(t *testing.T) {
	const decl = `
type TTest = class
   Field : String;
   constructor Create(f : String); begin Field := f end;
   destructor Destroy; override; begin PrintLn(Field) end;
end;
var a : array [TTest] of TTest;
`

	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name: "slot replace destroys the displaced value",
			body: `var k := new TTest('k');
a[k] := new TTest('alpha');
a[k] := new TTest('beta');
PrintLn('after');`,
			// 'beta' is still in the map at program end, so it dies there; the
			// key variable k outlives the map.
			expected: "alpha\nafter\nbeta\n",
		},
		{
			name: "delete destroys the value but not a still-referenced key",
			body: `var k := new TTest('k');
a[k] := new TTest('alpha');
a.Delete(k);
PrintLn('after');`,
			expected: "alpha\nafter\n",
		},
		{
			name: "clear destroys remaining values and object keys",
			body: `a[new TTest('k')] := new TTest('alpha');
a.Clear;
PrintLn('after');`,
			expected: "alpha\nk\nafter\n",
		},
		{
			name: "program end destroys what the map still owns",
			body: `a[new TTest('k')] := new TTest('alpha');
PrintLn('after');`,
			expected: "after\nk\nalpha\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			interp := New(&out)
			result := interpret(interp, decl+tt.body)
			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}
			if out.String() != tt.expected {
				t.Errorf("wrong output. expected=%q, got=%q", tt.expected, out.String())
			}
		})
	}
}

// TestAssociativeArrayVariantKeyCoercion checks that a Variant index is
// converted to the array's declared key type, so a value written through a
// Variant is found again under a literal key of the declared type.
func TestAssociativeArrayVariantKeyCoercion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "integer variant against a string key",
			input: `var a : array [String] of String;
var v : Variant = 123;
a[v] := 'bar';
PrintLn(a['123']);
PrintLn(a[v]);`,
			expected: "bar\nbar\n",
		},
		{
			name: "string variant against a string key",
			input: `var a : array [String] of String;
a['foo'] := 'bar';
var v : Variant = 'foo';
PrintLn(a[v]);`,
			expected: "bar\n",
		},
		{
			name: "numeric string variant against an integer key",
			input: `var a : array [Integer] of String;
a[7] := 'seven';
var v : Variant = '7';
PrintLn(a[v]);`,
			expected: "seven\n",
		},
		{
			name: "delete through a coerced variant key",
			input: `var a : array [String] of String;
a['123'] := 'bar';
var v : Variant = 123;
PrintLn(a.Delete(v));
PrintLn(a.Length);`,
			expected: "True\n0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			interp := New(&out)
			result := interpret(interp, tt.input)
			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}
			if out.String() != tt.expected {
				t.Errorf("wrong output. expected=%q, got=%q", tt.expected, out.String())
			}
		})
	}
}
