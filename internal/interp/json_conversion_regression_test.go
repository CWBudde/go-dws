package interp

import (
	"path/filepath"
	"testing"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
)

func TestJSONConversion_Fixtures(t *testing.T) {
	for _, name := range []string{"implicit_associative_key_cast", "implicit_from_cast", "write_immediate_prop"} {
		t.Run(name, func(t *testing.T) {
			const category = "JSONConnectorPass"
			path := filepath.Join(fixturesRoot, category, name+".pas")
			if got, detail := runFixtureTest(path, false, fixtureconfig.HintsLevel(category)); got != testResultPassed {
				t.Fatalf("%v: %s", got, detail)
			}
		})
	}
}

func TestJSONConversion_AssociativeKeys(t *testing.T) {
	const source = `type TKey = String;
var a: array [TKey] of String;
var key: JSONVariant := 123;
a[key] := 'first';
PrintLn(a['123']);
PrintLn(a[key]);
PrintLn(key in a);
a['123'] := 'second';
PrintLn(a[key]);
PrintLn(a.Delete(key));
PrintLn('123' in a);
key := 'word';
a[key] := 'third';
PrintLn(a['word']);
PrintLn(a.Delete('word'));
PrintLn(key in a);
var b: array [Variant] of String;
b[1] := 'integer';
b['1'] := 'string';
PrintLn(b.Count);
PrintLn(b[1]);
PrintLn(b['1']);`
	compileAndRunWithHelperTransfer(t, source, "json_keys.dws", "first\nfirst\nTrue\nsecond\nTrue\nFalse\nthird\nTrue\nFalse\n2\ninteger\nstring\n")
}

func TestJSONConversion_DeclaredAliases(t *testing.T) {
	const source = `type TJSON = JSONVariant;
type TAlias = TJSON;
procedure Show(const v: TAlias);
begin PrintLn(v.TypeName()); end;
function Identity(v: TJSON): TAlias;
begin Result := v; end;
type TTest = class
 procedure Show(v: TAlias);
 begin PrintLn(v.TypeName()); end;
end;
var uninitialized: Variant;
var missing: TAlias := uninitialized;
var explicitNull: TAlias := Null;
var text: TAlias := 'hello';
var integerValue: TAlias := 42;
PrintLn(missing.TypeName());
PrintLn(explicitNull.TypeName());
PrintLn(text.TypeName());
PrintLn(integerValue.TypeName());
Show(uninitialized);
Show(Unassigned);
Show(Null);
Show(False);
Show(0.5);
PrintLn(Identity('value').TypeName());
var obj := TTest.Create;
obj.Show(uninitialized);
obj.Show(Null);
obj.Show(1);
var showLambda := lambda(v: TAlias): String => v.TypeName();
PrintLn(showLambda(uninitialized));
PrintLn(showLambda(Null));
PrintLn(showLambda(True));
var original := JSON.NewObject;
var aliasValue: TAlias := Identity(original);
aliasValue.name := 'shared';
PrintLn(original.name);
missing := uninitialized;
PrintLn(missing.TypeName());`
	compileAndRunWithHelperTransfer(t, source, "json_aliases.dws", "Undefined\nNull\nString\nNumber\nUndefined\nUndefined\nNull\nBoolean\nNumber\nString\nUndefined\nNull\nNumber\nUndefined\nNull\nBoolean\nshared\nUndefined\n")
}

func TestJSONConversion_TypedStorage(t *testing.T) {
	for _, test := range []struct{ name, source string }{
		{"map alias", `type TInt = Integer; var a: array[String] of TInt; a['key'] := v; PrintLn(a['key'] + 1);`},
		{"member map", `type TRec = record a: array[String] of Integer; end; var r: TRec; r.a['key'] := v; PrintLn(r.a['key'] + 1);`},
		{"static array", `var a: array[0..0] of Integer; a[0] := v; PrintLn(a[0] + 1);`},
		{"dynamic array", `var a: array of Integer; a.Add(0); a[0] := v; PrintLn(a[0] + 1);`},
		{"array add", `var a: array of Integer; a.Add(v); PrintLn(a[0] + 1);`},
		{"record field", `type TRec = record x: Integer; end; var r: TRec; r.x := v; PrintLn(r.x + 1);`},
		{"object field", `type TObj = class x: Integer; end; var obj := TObj.Create; obj.x := v; PrintLn(obj.x + 1);`},
	} {
		t.Run(test.name, func(t *testing.T) {
			compileAndRunWithHelperTransfer(t, `var v: JSONVariant := 1; `+test.source, "json_typed_storage.dws", "2\n")
		})
	}
}

func TestJSONConversion_JSONStorage(t *testing.T) {
	for _, test := range []struct{ name, source string }{
		{"map", `var a: array[String] of TJSON; a['key'] := missing; PrintLn(a['key'].TypeName()); a['key'] := Null; PrintLn(a['key'].TypeName());`},
		{"array", `var a: array[0..0] of TJSON; a[0] := missing; PrintLn(a[0].TypeName()); a[0] := Null; PrintLn(a[0].TypeName());`},
		{"array add", `var a: array of TJSON; a.Add(missing, Null); PrintLn(a[0].TypeName()); PrintLn(a[1].TypeName());`},
		{"record field", `type TRec = record x: TJSON; end; var r: TRec; r.x := missing; PrintLn(r.x.TypeName()); r.x := Null; PrintLn(r.x.TypeName());`},
		{"object field", `type TObj = class x: TJSON; end; var obj := TObj.Create; obj.x := missing; PrintLn(obj.x.TypeName()); obj.x := Null; PrintLn(obj.x.TypeName());`},
	} {
		t.Run(test.name, func(t *testing.T) {
			compileAndRunWithHelperTransfer(t, `type TJSON = JSONVariant; var missing: Variant; `+test.source, "json_storage.dws", "Undefined\nNull\n")
		})
	}
}

func TestJSONConversion_ScalarStorage(t *testing.T) {
	const source = `type TRec = record f: Float; s: String; b: Boolean; end;
var r: TRec;
r.f := JSON.Parse('1.5');
r.s := JSON.Parse('"xy"');
r.b := JSON.Parse('true');
PrintLn(r.f + 1);
PrintLn(r.s[1]);
PrintLn(r.b and True);
r.b := JSON.Parse('"false"');
PrintLn(r.b);
r.b := JSON.Parse('0');
PrintLn(r.b);
var boolMap: array[String] of Boolean;
boolMap['flag'] := JSON.Parse('true');
PrintLn(boolMap['flag'] and True);
var texts: array of String;
texts.Add(JSON.Parse('"word"'));
PrintLn(texts[0][1]);
var numbers: array[0..0] of Float;
numbers[0] := JSON.Parse('1.5');
PrintLn(numbers[0] + 1);`
	compileAndRunWithHelperTransfer(t, source, "json_scalar_storage.dws", "2.5\nx\nTrue\nTrue\nFalse\nTrue\nw\n2.5\n")
}
