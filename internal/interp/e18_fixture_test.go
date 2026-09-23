package interp

import (
	"path/filepath"
	"testing"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
	"github.com/cwbudde/go-dws/internal/frontend"
)

func TestE18_Fixtures(t *testing.T) {
	for _, tc := range []struct {
		category string
		name     string
		fails    bool
	}{
		{"PropertyExpressionsPass", "read_write_other_property", false},
		{"PropertyExpressionsFail", "read_write_other_property", true},
		{"FunctionsString", "toxml", false},
	} {
		t.Run(tc.category+"/"+tc.name, func(t *testing.T) {
			path := filepath.Join(fixturesRoot, tc.category, tc.name+".pas")
			result, detail := runFixtureTest(path, tc.fails, fixtureconfig.HintsLevel(tc.category))
			if result != testResultPassed {
				t.Fatalf("fixture failed: %s", detail)
			}
		})
	}
}

func TestE18_PropertyAccessorPrefersNearestOverride(t *testing.T) {
	source := `type
  TBase = class
    BaseField: Integer = 1;
    property Prop: Integer read BaseField;
  end;
  TChild = class(TBase)
    ChildField: Integer = 2;
    property Prop: Integer read ChildField;
    property Mapped: Integer read Prop;
  end;
PrintLn(TChild.Create.Mapped);`
	compiled := frontend.CompileWithOptions(source, frontend.Options{})
	if !compiled.SemanticSuccessful {
		t.Fatalf("compile failed: %v", compiled.DiagnosticStrings())
	}
	output, value := evalFixture(compiled, "")
	if value != nil && value.Type() == "ERROR" {
		t.Fatalf("runtime error: %v", value)
	}
	if got := output.String(); got != "2\n" {
		t.Fatalf("output = %q, want %q", got, "2\n")
	}
}
