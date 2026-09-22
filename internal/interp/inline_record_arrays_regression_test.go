package interp

import (
	"path/filepath"
	"testing"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
)

func TestInlineRecordArrays_JSONFixtures(t *testing.T) {
	const category = "JSONConnectorPass"
	for _, name := range []string{"const_array", "stringify_array_of_array"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(fixturesRoot, category, name+".pas")
			if got, detail := runFixtureTest(path, false, fixtureconfig.HintsLevel(category)); got != testResultPassed {
				t.Fatalf("%v: %s", got, detail)
			}
		})
	}
}

func TestInlineRecordArrays_NestedMutation(t *testing.T) {
	const source = `
type TRows = array[2..3] of record value: Integer; end;
var rows: TRows;
rows[2].value := 42;
rows[3].value := 43;
PrintLn(JSON.Stringify(rows));
var nested: array of array of record values: array of Integer; end;
nested.SetLength(1);
nested[0].SetLength(2);
nested[0][0].values.Add(7);
nested[0][1].values.Add(8);
PrintLn(JSON.Stringify(nested));`
	compileAndRunWithHelperTransfer(t, source, "inline_record_arrays.dws", "[{\"value\":42},{\"value\":43}]\n[[{\"values\":[7]},{\"values\":[8]}]]\n")
}
