package semantic

import "testing"

func TestInterfaceProperties_AccessValidation(t *testing.T) {
	const prefix = `type IItems = interface
 function GetItem(x: String): Integer;
 procedure SetItem(x: String; v: Integer);
 property RO[x: String]: Integer read GetItem;
 property WO[x: String]: Integer write SetItem; default;
 end;
 var i: IItems;
 `
	for _, tc := range []struct{ name, source, want string }{
		{"readonly", `i.RO['x'] := 1;`, "read-only"},
		{"writeonly", `PrintLn(i.WO['x']);`, "write only"},
		{"default writeonly", `PrintLn(i['x']);`, "write only"},
		{"compound writeonly", `i['x'] += 1;`, "write only"},
		{"wrong index", `i[2] := 1;`, "Array index expected"},
		{"wrong value", `i['x'] := 'str';`, "expects type"},
	} {
		t.Run(tc.name, func(t *testing.T) { expectError(t, prefix+tc.source, tc.want) })
	}
}

func TestInterfaceProperties_IndexArity(t *testing.T) {
	const prefix = `type IItems = interface
 function GetItem(x, y: Integer): Integer;
 property Items[x, y: Integer]: Integer read GetItem; default;
 end;
 var i: IItems;
 `
	for _, source := range []string{`PrintLn(i[1]);`, `PrintLn(i.Items[1]);`} {
		expectError(t, prefix+source, "expects 2 index arguments, got 1")
	}
}
