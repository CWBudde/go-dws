package interp

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/types"
)

func TestEmptyArrayConstantAssignmentRuntime(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `const empty = [];
var integers: array of Integer := [1, 2];
integers := empty;
integers.Add(3);
PrintLn(integers[0]);
PrintLn(Length(integers));
var strings: array of String := ['old'];
strings := empty;
strings.Add('ok');
PrintLn(strings[0]);`), "3\n1\nok\n")
}

func TestEmptyArrayConstantInitializerRuntime(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `const empty = [];
var integers: array of Integer := empty;
integers.Add(3);
var strings: array of String := empty;
strings.Add('ok');
PrintLn(integers[0]);
PrintLn(strings[0]);
PrintLn(Length(empty));`), "3\nok\n0\n")
}

func TestEmptyArrayConstantVarParameterRuntime(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `const empty = [];
procedure Reset(var values: array of Integer);
begin
   values := empty;
   values.Add(3);
end;
var integers: array of Integer := [1, 2];
Reset(integers);
PrintLn(integers[0]);
PrintLn(Length(integers));
PrintLn(Length(empty));`), "3\n1\n0\n")
}

func TestStaticArrayConstantConversionIsolationRuntime(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `const original = [1, 2];
var first: array of Integer := original;
first[0] := 9;
first.Add(3);
var second: array of Integer := original;
second.Add(4);
var shared: array of Integer := second;
shared[0] := 7;
PrintLn(original[0]);
PrintLn(Length(original));
PrintLn(first[0]);
PrintLn(Length(first));
PrintLn(second[0]);
PrintLn(shared[0]);`), "1\n2\n9\n3\n7\n7\n")
}

// Exercise conversion directly so an incompatible array cannot be retyped
// even when a caller bypasses compile-time assignment checks.
func TestImplicitStaticArrayConversionCompatibility(t *testing.T) {
	destination := types.NewDynamicArrayType(types.INTEGER)
	tests := []struct {
		source      *runtime.ArrayValue
		name        string
		wantConvert bool
	}{
		{&runtime.ArrayValue{ArrayType: types.NewStaticArrayType(types.VARIANT, 0, -1)}, "empty constant", true},
		{&runtime.ArrayValue{ArrayType: types.NewStaticArrayType(types.INTEGER, 0, 0), Elements: []runtime.Value{&runtime.IntegerValue{Value: 1}}}, "matching static", true},
		{&runtime.ArrayValue{ArrayType: types.NewStaticArrayType(types.STRING, 0, 0), Elements: []runtime.Value{&runtime.StringValue{Value: "text"}}}, "incompatible element", false},
		{&runtime.ArrayValue{ArrayType: types.NewStaticArrayType(types.FLOAT, 0, 0), Elements: []runtime.Value{&runtime.FloatValue{Value: 1}}}, "numeric element invariance", false},
		{&runtime.ArrayValue{ArrayType: destination, Elements: []runtime.Value{&runtime.IntegerValue{Value: 1}}}, "dynamic reference", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := evaluator.NewEvaluator(interptypes.NewTypeSystem(), nil, nil, nil, nil, runtime.NewRefCountManager())
			got, converted := e.TryImplicitConversion(tt.source, destination, evaluator.NewExecutionContext(nil))
			if converted != tt.wantConvert {
				t.Fatalf("conversion = %v, want %v", converted, tt.wantConvert)
			}
			if !converted {
				if got != tt.source {
					t.Fatal("unconverted array did not retain its identity")
				}
				return
			}
			array, ok := got.(*runtime.ArrayValue)
			if !ok || array == tt.source || array.ArrayType != destination {
				t.Fatalf("converted array must be a fresh value with destination metadata: %#v", got)
			}
			if len(array.Elements) != len(tt.source.Elements) {
				t.Fatal("conversion changed the number of elements")
			}
			if len(array.Elements) != 0 {
				array.Elements[0] = &runtime.IntegerValue{Value: 8}
				if tt.source.Elements[0].String() != "1" {
					t.Fatal("converted array shares constant element storage")
				}
			}
		})
	}
}

func TestArrayConstantAggregateStorageRuntime(t *testing.T) {
	tests := []struct {
		name   string
		setup  string
		target string
	}{
		{
			name:   "class field",
			setup:  "type THolder = class Values: array of Integer; end; var holder := THolder.Create;",
			target: "holder.Values",
		},
		{
			name:   "class variable",
			setup:  "type THolder = class class var Values: array of Integer; end;",
			target: "THolder.Values",
		},
		{
			name: "class property",
			setup: `type THolder = class
   class var FValues: array of Integer;
   class property Values: array of Integer read FValues write FValues;
end;`,
			target: "THolder.Values",
		},
		{
			name: "metaclass property",
			setup: `type THolder = class
   class var FValues: array of Integer;
   class property Values: array of Integer read FValues write FValues;
end;
type THolderClass = class of THolder;
var holderClass: THolderClass := THolder;`,
			target: "holderClass.Values",
		},
		{
			name:   "record field",
			setup:  "type THolder = record Values: array of Integer; end; var holder: THolder;",
			target: "holder.Values",
		},
		{
			name: "inherited field-backed property",
			setup: `type TBase = class
   FValues: array of Integer;
   property Values: array of Integer read FValues write FValues;
end;
type THolder = class(TBase) end;
var holder := THolder.Create;`,
			target: "holder.Values",
		},
		{
			name: "record field-backed property",
			setup: `type THolder = record
   FValues: array of Integer;
   property Values: array of Integer read FValues write FValues;
end;
var holder: THolder;`,
			target: "holder.Values",
		},
		{
			name:   "nested array slot",
			setup:  "var rows: array of array of Integer; rows.SetLength(1);",
			target: "rows[0]",
		},
		{
			name:   "associative array slot",
			setup:  "var rows: array[String] of array of Integer;",
			target: "rows['key']",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("empty", func(t *testing.T) {
				source := "const original = [];\n" + tt.setup + "\n" +
					tt.target + " := original;\n" + tt.target + ".Add(3);\n" +
					"var stored: array of Integer := " + tt.target + ";\nPrintLn(stored[0]);\nPrintLn(Length(original));"
				assertOutput(t, runQuickwinScript(t, source), "3\n0\n")
			})
			t.Run("populated", func(t *testing.T) {
				source := "const original = [1, 2];\n" + tt.setup + "\n" +
					tt.target + " := original;\n" + tt.target + "[0] := 9;\n" + tt.target + ".Add(3);\n" +
					"var stored: array of Integer := " + tt.target + ";\nPrintLn(original[0]);\nPrintLn(Length(original));\nPrintLn(stored[0]);\nPrintLn(Length(stored));"
				assertOutput(t, runQuickwinScript(t, source), "1\n2\n9\n3\n")
			})
		})
	}
}

func TestArrayConstantNestedAddRuntime(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `const empty = [];
const original = [1, 2];
var rows: array of array of Integer;
rows.Add(empty);
rows[0].Add(3);
rows.Add(original);
rows[1][0] := 9;
rows[1].Add(4);
PrintLn(rows[0][0]);
PrintLn(rows[1][0]);
PrintLn(Length(rows[1]));
PrintLn(Length(empty));
PrintLn(original[0]);
PrintLn(Length(original));`), "3\n9\n3\n0\n1\n2\n")
}

func TestArrayConstantNestedInsertionRuntime(t *testing.T) {
	for _, operation := range []struct {
		name  string
		setup string
	}{
		{"Insert", "var rows: array of array of Integer; rows.Add([99]); rows.Insert(0, original);"},
		{"typed constructor", "var rows: array of array of Integer := [original];"},
	} {
		t.Run(operation.name, func(t *testing.T) {
			t.Run("empty", func(t *testing.T) {
				source := "const original = [];\n" + operation.setup + `
rows[0].Add(3);
PrintLn(rows[0][0]);
PrintLn(Length(rows[0]));
PrintLn(Length(original));`
				assertOutput(t, runQuickwinScript(t, source), "3\n1\n0\n")
			})
			t.Run("populated", func(t *testing.T) {
				source := "const original = [1, 2];\n" + operation.setup + `
rows[0][0] := 9;
rows[0].Add(3);
PrintLn(rows[0][0]);
PrintLn(Length(rows[0]));
PrintLn(original[0]);
PrintLn(Length(original));`
				assertOutput(t, runQuickwinScript(t, source), "9\n3\n1\n2\n")
			})
			t.Run("dynamic reference", func(t *testing.T) {
				source := "var original: array of Integer := [1, 2];\n" + operation.setup + `
rows[0][0] := 9;
rows[0].Add(3);
PrintLn(original[0]);
PrintLn(Length(original));`
				assertOutput(t, runQuickwinScript(t, source), "9\n3\n")
			})
		})
	}
}
