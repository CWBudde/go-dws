package interp

import (
	"path/filepath"
	"testing"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
)

func TestJSONMutation_InvalidDelete(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
var a := JSON.Parse('[1,2]');
try a.Delete(-1); except on E: Exception do PrintLn(E.Message); end;
try a.Delete(2); except on E: Exception do PrintLn(E.Message); end;
PrintLn(a);
a.Delete(0);
a.Delete(0);
try a.Delete(0); except on E: Exception do PrintLn(E.Message); end;
PrintLn(a);
`), "Array index (-1) out of range [0..1]\nArray index (2) out of range [0..1]\n[1,2]\nArray index (0) out of range (empty array)\n[]\n")
}

func TestJSONMutation_CyclePreservesOwnership(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
var root := JSON.Parse('{"before":1,"child":{"a":[],"kept":2},"after":3}');
var child := root.child;
var a := child.a;
try child.kept := root; except on E: Exception do PrintLn(E.Message); end;
try child['added'] := root; except on E: Exception do PrintLn(E.Message); end;
try a[5] := root; except on E: Exception do PrintLn(E.Message); end;
try a.Add(child); except on E: Exception do PrintLn(E.Message); end;
PrintLn(root);
PrintLn(child);
PrintLn(a);
`), "JSON circular reference\nJSON circular reference\nJSON circular reference\nJSON circular reference\n"+
		"{\"before\":1,\"child\":{\"a\":[],\"kept\":2},\"after\":3}\n{\"a\":[],\"kept\":2}\n[]\n")
}

func TestJSONMutation_AddSuccessfulPrefix(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
var a := JSON.NewArray;
try a.Add(1, a, 2); except on E: Exception do PrintLn(E.Message); end;
PrintLn(a);
var source := JSON.NewArray;
source.Add(7, a, 8);
try a.AddFrom(source); except on E: Exception do PrintLn(E.Message); end;
PrintLn(source);
PrintLn(a);
a.AddFrom(a);
PrintLn(a);
`), "JSON circular reference\n[1]\nJSON circular reference\n[[1,7],8]\n[1,7]\n[1,7]\n")
}

func TestJSONMutation_Fixtures(t *testing.T) {
	const category = "JSONConnectorPass"
	for _, name := range []string{"delete_array_index", "circular_references", "reparent", "reposition_node_in_array", "array_add_dupe", "array_add_push"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(fixturesRoot, category, name+".pas")
			if got, detail := runFixtureTest(path, false, fixtureconfig.HintsLevel(category)); got != testResultPassed {
				t.Fatalf("%s/%s: %v: %s", category, name, got, detail)
			}
		})
	}
}
