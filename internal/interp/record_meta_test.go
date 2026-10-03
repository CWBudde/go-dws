package interp

import (
	"fmt"
	"testing"
)

func TestRecordMeta_CompiledDispatch(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type R = record
 x: Integer;
 class var Count: Integer;
 class const Increment = 2;
 class function Get: Integer; begin Result := Count; end;
 class procedure Add(n: Integer); begin Count += n; end;
 property Total: Integer read Count write Count;
end;
type A = R;
type H = helper for R
 class function Read: Integer; begin Result := Self.Count; end;
 function Instance: Integer; begin Result := Self.x; end;
end;
var m := R; var copy := m; var aliasMeta := A;
m.Count := 1;
A.Add(R.Increment);
copy.Total += 4;
PrintLn(R.Get); PrintLn(A.Get()); PrintLn(m.Get); PrintLn(copy.Get());
PrintLn(aliasMeta.Read); PrintLn(H.Read(m)); PrintLn(H.Read(A));
var item: A; item.x := 9; var itemCopy := item; item.x := 8;
PrintLn(itemCopy.Instance);
`, "record_meta.dws", "7\n7\n7\n7\n7\n7\n7\n9\n")
}

func TestRecordMeta_CompiledReceivingPaths(t *testing.T) {
	for _, tt := range []struct{ name, source, want string }{
		{"helper alias ownership", `
type R = record class var Count: Integer; end; type A = R;
type H = helper for R class function Label: String; begin Result := 'record'; end; end;
type HA = helper for A class function Label: String; begin Result := 'alias'; end; end;
var m := A; var copy := m;
PrintLn(R.Label); PrintLn(A.Label); PrintLn(m.Label()); PrintLn(copy.Label);
m.Count := 4; PrintLn(R.Count);
`, "record\nalias\nalias\nalias\n4\n"},
		{"helper class property", `
type R = record class var Count: Integer; end;
type H = helper for R
 class function Get: Integer; begin Result := Self.Count; end;
 class procedure Put(v: Integer); begin Self.Count := v; end;
 class property Total: Integer read Get write Put;
end;
var m := R; m.Total := 3; m.Total += 4; PrintLn(R.Total);
`, "7\n"},
		{"declared inferred meta", `type R = record class var Count: Integer; x: Integer; end;
var m := R; PrintLn(Declared('m.Count')); PrintLn(Declared('m.x'));`, "True\nFalse\n"},
		{"record method property", `
type R = record
 class var Count: Integer;
 class function Get: Integer; begin Result := Count; end;
 class procedure Put(v: Integer); begin Count := v; end;
 property Total: Integer read Get write Put;
end;
var m := R; m.Total := 3; m.Total += 4; PrintLn(R.Total);
`, "7\n"},
		{"overloads retain value category", `
type R = record x: Integer; end;
function Pick(v: R): String; overload; begin Result := 'instance'; end;
function Pick(v: Variant): String; overload; begin Result := 'meta'; end;
type Host = record
 class function Pick(v: R): String; overload; begin Result := 'instance'; end;
 class function Pick(v: Variant): String; overload; begin Result := 'meta'; end;
end;
type A = R; var m := A; var item: R;
PrintLn(Pick(m)); PrintLn(Pick(item)); PrintLn(Host.Pick(A)); PrintLn(Host.Pick(item));
`, "meta\ninstance\nmeta\ninstance\n"},
	} {
		t.Run(tt.name, func(t *testing.T) { compileAndRunWithHelperTransfer(t, tt.source, "record_receivers.dws", tt.want) })
	}
}

func TestRecordMeta_IndexedPropertyCapture(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type R = record
 class var Count: Integer;
 class function Get(i: Integer): Integer; begin Result := Count + i; end;
 class procedure Put(i, v: Integer); begin Count := v - i; end;
 property Items[i: Integer]: Integer read Get write Put; default;
end;
type A = R; var m := A;
var calls := 0;
function NextIndex: Integer; begin calls += 1; Result := 2; end;
m.Items[NextIndex()] := 5;
m.Items[NextIndex()] += 4;
PrintLn(m.Items[2]); PrintLn(m[2]);
m[NextIndex()] += 1;
PrintLn(R.Count); PrintLn(calls);
`, "record_indexed_meta.dws", "9\n9\n8\n3\n")
}

func TestRecordMeta_HelperBareClassMembers(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type R = record class var Count: Integer; class function Get: Integer; begin Result := Count; end; end;
type H = helper for R
 class function Add: Integer; begin Count += 1; Result := Get; end;
end;
PrintLn(R.Add); PrintLn(R.Count);
`, "record_helper_scope.dws", "1\n1\n")
}

func TestRecordMeta_ClassPropertyExpressions(t *testing.T) {
	for i, declaration := range []string{
		`type R = record class var Count: Integer; class property Double: Integer read (Count*2) write (Count:=Value div 2); end;`,
		`type R = record class var Count: Integer; end; type H = helper for R class property Double: Integer read (Count*2) write (Count:=Value div 2); end;`,
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			compileAndRunWithHelperTransfer(t, declaration+`
type A = R; var m := A;
m.Double := 6; m.Double += 4; PrintLn(R.Double); PrintLn(A.Count);
`, "record_class_expression.dws", "10\n5\n")
		})
	}
}

func TestRecordMeta_ClassHelperOnInstanceHasMetaSelf(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type R = record class var Count: Integer; end;
type H = helper for R
 class function Add: Integer;
 begin var m := R; m := Self; m.Count += 1; Result := Self.Count; end;
end;
var item: R; PrintLn(item.Add); PrintLn(R.Count);
`, "record_instance_class_helper.dws", "1\n1\n")
}

func TestRecordMeta_ExplicitHelperOverloads(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type R = record x: Integer; end; type A = R;
type H = helper for R
 class function Pick(n: Integer): String; overload; begin Result := 'integer'; end;
 class function Pick(s: String): String; overload; begin Result := 'string'; end;
end;
var m := A; var copy := m;
PrintLn(H.Pick(R,1)); PrintLn(H.Pick(A,'s')); PrintLn(H.Pick(m,1)); PrintLn(H.Pick(copy,'s'));
`, "record_explicit_overloads.dws", "integer\nstring\ninteger\nstring\n")
}

func TestRecordMeta_ArrayCarrier(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type R = record class var Count: Integer; end; type A = R;
var values := [R, A]; values[0].Count := 3; PrintLn(values[1].Count);
`, "record_meta_array.dws", "3\n")
}

func TestRecordMeta_HelperOverloadDispatch(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `type R=record x: Integer; end;
type H=helper for R
 class function Pick(n: Integer): String; overload; begin Result:='class'; end;
 function Pick(s: String): String; overload; begin Result:='instance'; end;
end;
PrintLn(R.Pick(1)); var item:R; PrintLn(item.Pick('ok'));
`, "record_helper_overloads.dws", "class\ninstance\n")
}

func TestRecordMeta_HelperPrecedenceAndInheritance(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type R=record x:Integer; end;
type First=helper for R class function Pick: Integer; begin Result:=4; end; end;
type Later=helper for R class function Pick: String; begin Result:='later'; end; end;
var n:=R.Pick; PrintLn(n+1);
type Child=helper(First) for R class function Twice:Integer; begin Result:=2*Pick; end; end;
PrintLn(R.Twice); PrintLn(Child.Pick(R));
`, "record_helper_precedence.dws", "5\n8\n4\n")
}

func TestRecordMeta_DefaultPropertyIndexOwnership(t *testing.T) {
	const declaration = `
var trace := '';
type R = record
 class var Count: Integer;
 class function Get(i,j: Integer): Integer;
 begin trace += 'g'; Result := Count + 10*i + j; end;
 class procedure Put(i,j,v: Integer);
 begin trace += 'w'; Count := v - 10*i - j; end;
 property Items[i: Integer; j: Integer]: Integer read Get write Put; default;
end;
var m := R; var values := [R]; var grid := [values];
function Mark(labelText: String; value: Integer): Integer;
begin trace += labelText; Result := value; end;
function Increment: Integer; begin trace += 'r'; Result := 5; end;
`
	for _, receiver := range []struct{ name, expression, trace string }{
		{"direct", "m", ""},
		{"aggregate", "values[Mark('a',0)]", "a"},
		{"nested aggregate", "grid[Mark('a',0)][Mark('b',0)]", "ab"},
		{"grouped aggregate", "(values[Mark('a',0)])", "a"},
		{"named control", "values[Mark('a',0)].Items", "a"},
	} {
		for _, operation := range []struct{ name, source, output, trace string }{
			{"read", "PrintLn(%s);", "12\n", "ijg"},
			{"write", "%s := 40; PrintLn(R.Count);", "28\n", "ijw"},
			{"compound", "%s += Increment(); PrintLn(R.Count);", "5\n", "ijgrw"},
		} {
			if receiver.name == "grouped aggregate" && operation.name != "read" {
				continue // Parenthesized expressions are not statement assignment targets.
			}
			t.Run(receiver.name+"/"+operation.name, func(t *testing.T) {
				target := receiver.expression + "[Mark('i',1),Mark('j',2)]"
				source := declaration + fmt.Sprintf(operation.source, target) + "PrintLn(trace);"
				want := operation.output + receiver.trace + operation.trace + "\n"
				compileAndRunWithHelperTransfer(t, source, "record_default_indices.dws", want)
			})
		}
	}
}

func TestRecordMeta_DefaultPropertyAfterArraySelection(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type R = record
 class var Count: Integer;
 class function Get(i: Integer): Integer; begin Result := Count + i; end;
 class procedure Put(i,v: Integer); begin Count := v - i; end;
 property Items[i: Integer]: Integer read Get write Put; default;
end;
var values := [R]; var calls := 0;
function NextIndex: Integer; begin calls += 1; Result := 0; end;
values[NextIndex()][1] := 8;
values[NextIndex()][1] += 2;
PrintLn(values[NextIndex()][1]); PrintLn(calls);
`, "record_default_array.dws", "10\n3\n")
}
