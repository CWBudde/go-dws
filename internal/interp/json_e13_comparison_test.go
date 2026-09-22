package interp

import (
	"path/filepath"
	"testing"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
)

func TestJSONE13ComparisonFixtures(t *testing.T) {
	for _, name := range []string{"comparison2", "in_static", "undefined_vs_unassigned"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(fixturesRoot, "JSONConnectorPass", name+".pas")
			if got, detail := runFixtureTest(path, false, fixtureconfig.HintsLevel("JSONConnectorPass")); got != testResultPassed {
				t.Fatal(detail)
			}
		})
	}
}

func TestJSONE13ComparisonCoercion(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
var n := JSON.Parse('1');
var s := JSON.Parse('"01"');
PrintLn(n = s); PrintLn(s = n);
PrintLn(s = JSON.Parse('"1"'));
var big := JSON.Parse('9007199254740993');
PrintLn(big = '9007199254740993');
PrintLn(big = '9007199254740992');
PrintLn('9007199254740992' < big);
var bad := JSON.Parse('"abc"');
PrintLn(n = bad); PrintLn(n <> bad);
PrintLn(n < bad); PrintLn(n <= bad); PrintLn(n > bad); PrintLn(n >= bad);
PrintLn(bad < n); PrintLn(bad <= n); PrintLn(bad > n); PrintLn(bad >= n);
var v: Variant := 1;
var strings: array of String := ['01', '2'];
PrintLn(v in strings);
var b: Variant := False;
PrintLn(b in [False..True]);
`), "True\nTrue\nFalse\nTrue\nFalse\nTrue\nFalse\nTrue\nFalse\nFalse\nFalse\nFalse\nFalse\nFalse\nFalse\nFalse\nTrue\nTrue\n")
}

func TestJSONE13MembershipEvaluatesOnce(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
var calls := 0;
function Next: Variant;
begin Inc(calls); Result := 1 end;
var v: Variant := 'missing';
PrintLn(v in [Next()]);
PrintLn(calls);
var j := JSON.Parse('"1"');
PrintLn(j in [Next()]);
PrintLn(calls);
`), "False\n1\nTrue\n2\n")
}
