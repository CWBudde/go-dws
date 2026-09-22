package interp

import (
	"path/filepath"
	"testing"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
)

func TestE3cFixtures(t *testing.T) {
	for _, tc := range []struct {
		category string
		name     string
		fail     bool
	}{
		{"SimpleScripts", "class_var_dyn2", false},
		{"SimpleScripts", "string_builtin_methods", false},
		{"ArrayPass", "dynamic_anonymous_record", false},
		{"FailureScripts", "block_unfinished4", true},
	} {
		t.Run(tc.category+"/"+tc.name, func(t *testing.T) {
			path := filepath.Join(fixturesRoot, tc.category, tc.name+".pas")
			if got, detail := runFixtureTest(path, tc.fail, fixtureconfig.HintsLevel(tc.category)); got != testResultPassed {
				t.Fatalf("%s: %v: %s", tc.name, got, detail)
			}
		})
	}
}

func TestE3cStringBounds_UnicodeAndEmpty(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
var s := '';
PrintLn(s.Low, ':', s.High, ':', s.Length());
s := 'a😀';
PrintLn(s.Low(), ':', s.High(), ':', s.Length);
`, "e3c_string_bounds.dws", "1:0:0\n1:2:2\n")
}

func TestE3cAnonymousRecord_PropertyMethodAndCopy(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
var r := record
  Value := 3;
  property Current: Integer read Value;
  function Twice: Integer;
  begin
    Result := Current * 2;
  end;
end;
var copy := r;
r.Value := 5;
PrintLn(r.Current, ':', r.Twice());
PrintLn(copy.Current, ':', copy.Twice);
`, "e3c_anonymous_record.dws", "5:10\n3:6\n")
}

func TestE3cAnonymousRecord_AutoProperty(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
var r := record
  property Value: Integer;
end;
r.Value := 4;
PrintLn(r.Value);
`, "e3c_anonymous_auto_property.dws", "4\n")
}
