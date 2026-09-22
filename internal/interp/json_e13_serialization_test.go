package interp

import (
	"strings"
	"testing"
)

func TestJSONE13DuplicateTextualKeys(t *testing.T) {
	output := runQuickwinScript(t, `
var m: array[Variant] of Integer;
m[1] := 11; m['1'] := 22;
PrintLn(m[1]); PrintLn(m['1']);
PrintLn(JSON.Stringify(m));
PrintLn(JSON.StringifyUTF8(m));
PrintLn(JSON.PrettyStringify(m, '  '));
var a: array of array[Variant] of Integer;
a.Add(m); PrintLn(JSON.Stringify(a));
type TData = record Items: array[Variant] of Integer; end;
var r: TData; r.Items := m; PrintLn(JSON.Stringify(r));
PrintLn(JSON.Stringify(JSON.Serialize(m)));
PrintLn(m[1]); PrintLn(m['1']);
`)
	// Bucket traversal determines member order. Both distinct keys must survive
	// text serialization; Serialize uses the last emitted value for lookup.
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 3 {
		t.Fatalf("unexpected output: %s", output)
	}
	compact := lines[2]
	if compact != `{"1":11,"1":22}` && compact != `{"1":22,"1":11}` {
		t.Fatalf("duplicate keys lost: %s", output)
	}
	first, last := "11", "22"
	if compact == `{"1":22,"1":11}` {
		first, last = last, first
	}
	pretty := "{\r\n  \"1\" : " + first + ",\r\n  \"1\" : " + last + "\r\n}"
	want := "11\n22\n" + compact + "\n" + compact + "\n" + pretty + "\n[" + compact + "]\n{\"Items\":" + compact + "}\n{\"1\":" + last + "}\n11\n22\n"
	assertOutput(t, output, want)
}

func TestJSONE13CustomStringifyDuplicateKeys(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
var calls := 0;
type TData = class
  function Stringify: String;
  begin Inc(calls); Result := '{"x":1,"x":2}' end;
end;
PrintLn(JSON.Stringify(new TData));
PrintLn(JSON.PrettyStringify(new TData, '  '));
PrintLn(JSON.Stringify(JSON.Serialize(new TData)));
PrintLn(calls);
`), "{\"x\":1,\"x\":2}\n{\r\n  \"x\" : 1,\r\n  \"x\" : 2\r\n}\n{\"x\":2}\n3\n")
}

func TestJSONE13SerializeDuplicateOrder(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
type TData = class
  function Stringify: String;
  begin Result := '{"x":1,"middle":0,"x":2}' end;
end;
PrintLn(JSON.Stringify(JSON.Serialize(new TData)));
PrintLn(JSON.Stringify(JSON.Parse('{"x":1,"middle":0,"x":2}')));
`), "{\"middle\":0,\"x\":2}\n{\"middle\":0,\"x\":2}\n")
}
