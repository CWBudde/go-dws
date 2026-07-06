package jsonvalue

import (
	"math"
	"testing"
)

var negZero = math.Copysign(0, -1)

func TestStringifyCompact(t *testing.T) {
	tests := []struct {
		name string
		val  *Value
		want string
	}{
		{"int", NewInt64(1), "1"},
		{"float_half", NewNumber(1.5), "1.5"},
		{"neg_zero", NewNumber(negZero), "0"},
		{"exp", NewNumber(1e99), "1E99"},
		{"true", NewBoolean(true), "true"},
		{"false", NewBoolean(false), "false"},
		{"null", NewNull(), "null"},
		{"undefined", NewUndefined(), "null"},
		{"string", NewString("Hello"), `"Hello"`},
		{"slash", NewString("a/b"), `"a\/b"`},
		{"quote", NewString(`a"b`), `"a\"b"`},
		{"newline", NewString("a\nb"), `"a\nb"`},
		{"unicode_low", NewString("é"), `"é"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Stringify(tt.val); got != tt.want {
				t.Errorf("Stringify() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStringifyObjectOrder(t *testing.T) {
	obj := NewObject()
	obj.ObjectSet("hello", NewString("world"))
	obj.ObjectSet("one", NewInt64(1))
	obj.ObjectSet("half", NewNumber(0.5))
	obj.ObjectSet("yes", NewBoolean(true))
	want := `{"hello":"world","one":1,"half":0.5,"yes":true}`
	if got := Stringify(obj); got != want {
		t.Errorf("Stringify() = %q, want %q", got, want)
	}
}

func TestStringifyArrayNulls(t *testing.T) {
	arr := NewArray()
	arr.ArrayAppend(NewNumber(negZero))
	arr.ArrayAppend(NewNumber(nan()))
	arr.ArrayAppend(NewNumber(1e99))
	want := "[0,null,1E99]"
	if got := Stringify(arr); got != want {
		t.Errorf("Stringify() = %q, want %q", got, want)
	}
}

func TestParsePreservesOrder(t *testing.T) {
	v, err := Parse(`{"hello":"world","one":1,"half":0.5,"yes":true,"no":false}`)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if v.Kind() != KindObject {
		t.Fatalf("Kind = %v, want Object", v.Kind())
	}
	gotKeys := v.ObjectKeys()
	wantKeys := []string{"hello", "one", "half", "yes", "no"}
	if len(gotKeys) != len(wantKeys) {
		t.Fatalf("keys = %v, want %v", gotKeys, wantKeys)
	}
	for i := range wantKeys {
		if gotKeys[i] != wantKeys[i] {
			t.Errorf("key[%d] = %q, want %q", i, gotKeys[i], wantKeys[i])
		}
	}
	if v.ObjectGet("one").Kind() != KindInt64 {
		t.Errorf("one kind = %v, want Int64", v.ObjectGet("one").Kind())
	}
	if v.ObjectGet("half").Kind() != KindNumber {
		t.Errorf("half kind = %v, want Number", v.ObjectGet("half").Kind())
	}
}

func TestParseStringifyRoundTrip(t *testing.T) {
	src := `{"a":[1,2,3],"b":{"c":"d"}}`
	v, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if got := Stringify(v); got != src {
		t.Errorf("round-trip = %q, want %q", got, src)
	}
}

func nan() float64 {
	z := 0.0
	return z / z
}
