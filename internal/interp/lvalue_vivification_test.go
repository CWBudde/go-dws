package interp

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// TestNestedLValueVivification covers PLAN.md §3.3 "nested lvalue vivification
// through a key or index": reaching a missing associative-array key from an
// lvalue (or from a receiver whose method mutates it in place) must insert the
// slot, so the nested write lands in the map instead of in a throwaway zero
// value. A pure rvalue read must still leave the map untouched.
func TestNestedLValueVivification(t *testing.T) {
	tests := []struct {
		name   string
		script string
		want   string
	}{
		{
			name: "nested index write vivifies static array element",
			script: `type TStaticArray = array [0..2] of Integer;
var sa : array [Integer] of TStaticArray;
sa[1][1] := 123;
PrintLn(sa[1][1]);
PrintLn(sa.Length);
`,
			want: "123\n1\n",
		},
		{
			name: "nested member write vivifies record element",
			script: `type TRecord = record S : String; F : Float; end;
var ra : array [Integer] of TRecord;
ra[2].S := 'hello';
PrintLn(ra[2].S);
PrintLn(ra.Length);
`,
			want: "hello\n1\n",
		},
		{
			name: "mutating method receiver vivifies dynamic array element",
			script: `var a : array [String] of array of String;
a['alpha'].Add('beta');
PrintLn(a['alpha'].Join(','));
PrintLn(a.Keys.Sort.Join(','));
`,
			want: "beta\nalpha\n",
		},
		{
			name: "vivified slot is reused, not replaced",
			script: `var a : array [String] of array of String;
a['alpha'].Add('beta');
a['alpha'].Add('gamma');
PrintLn(a['alpha'].Join(','));
PrintLn(a.Length);
`,
			want: "beta,gamma\n1\n",
		},
		{
			name: "rvalue read of a missing key does not insert",
			script: `var a : array [String] of Integer;
a['present'] := 1;
PrintLn(a['missing']);
PrintLn(a.Length);
PrintLn(a.Keys.Join(','));
`,
			want: "0\n1\npresent\n",
		},
		{
			name: "rvalue read of a missing key through a nested index does not insert",
			script: `type TStaticArray = array [0..2] of Integer;
var sa : array [Integer] of TStaticArray;
PrintLn(sa[7][1]);
PrintLn(sa.Length);
`,
			want: "0\n0\n",
		},
		{
			name: "existing key is updated rather than duplicated",
			script: `type TRecord = record S : String; end;
var ra : array [Integer] of TRecord;
ra[2].S := 'first';
ra[2].S := 'second';
PrintLn(ra[2].S);
PrintLn(ra.Length);
`,
			want: "second\n1\n",
		},
		{
			name: "whole-element assignment still replaces the slot",
			script: `var a : array [String] of Integer;
a['k'] := 1;
a['k'] := 2;
PrintLn(a['k']);
PrintLn(a.Length);
`,
			want: "2\n1\n",
		},
		{
			name: "compound assignment on an existing key reads and writes back",
			script: `var a : array [String] of Integer;
a['k'] := 1;
a['k'] += 41;
PrintLn(a['k']);
PrintLn(a.Length);
`,
			want: "42\n1\n",
		},
		{
			name: "plain nested array indexing is unaffected",
			script: `var m : array of array of Integer;
m.SetLength(2);
m[0].SetLength(2);
m[0][1] := 7;
PrintLn(m[0][1]);
`,
			want: "7\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, out := testEvalWithOutput(tt.script)
			if errVal, ok := val.(*runtime.ErrorValue); ok {
				t.Fatalf("script raised an error: %s\noutput so far:\n%s", errVal.Message, out)
			}
			if out != tt.want {
				t.Errorf("output mismatch\n got: %q\nwant: %q", out, tt.want)
			}
		})
	}
}

// TestNestedJSONLValue covers the JSON half of the same defect: a JSON child
// reached through an index or a member in an lvalue position must be the live
// node, so that writing through it mutates the tree the variable refers to.
func TestNestedJSONLValue(t *testing.T) {
	tests := []struct {
		name   string
		script string
		want   string
	}{
		{
			name: "member write through an indexed JSON element",
			script: `var a := JSON.Parse('[]');
a[0] := JSON.Parse('{}');
a[0].TEST := 3;
PrintLn(JSON.Stringify(a));
`,
			want: `[{"TEST":3}]` + "\n",
		},
		{
			name: "member write through a JSON member",
			script: `var v := JSON.Parse('{}');
v.Flags := JSON.NewObject;
v.Flags.Bool := True;
PrintLn(JSON.Stringify(v));
`,
			want: `{"Flags":{"Bool":true}}` + "\n",
		},
		{
			name: "index write through a JSON member auto-extends the array",
			script: `var v := JSON.Parse('{}');
v.List := JSON.NewArray;
v.List[0] := 'zero';
v.List[2] := 2;
PrintLn(JSON.Stringify(v));
`,
			want: `{"List":["zero",null,2]}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, out := testEvalWithOutput(tt.script)
			if errVal, ok := val.(*runtime.ErrorValue); ok {
				t.Fatalf("script raised an error: %s\noutput so far:\n%s", errVal.Message, out)
			}
			if strings.TrimRight(out, "\n") != strings.TrimRight(tt.want, "\n") {
				t.Errorf("output mismatch\n got: %q\nwant: %q", out, tt.want)
			}
		})
	}
}
