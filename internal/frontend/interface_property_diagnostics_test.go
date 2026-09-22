package frontend

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestInterfacePropertyDiagnostics_WithLaterParserError(t *testing.T) {
	const source = `type IItems = interface
 function GetItem: String;
 property Item: Integer read GetItem;
 end;
 type IBroken = interface(IItems end;`
	result := Compile(source, "interface_property_parser_error.dws", semantic.HintsLevelDisabled)
	diagnostics := strings.Join(result.DiagnosticStrings(), "\n")
	for _, want := range []string{
		`Field/method "GetItem" has an incompatible type`,
		`")" expected`,
	} {
		if !strings.Contains(diagnostics, want) {
			t.Errorf("diagnostics %q do not contain %q", diagnostics, want)
		}
	}
}
