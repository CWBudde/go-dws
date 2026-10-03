package frontend

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_AnonymousCallableJSONFixture(t *testing.T) {
	base := filepath.Join("..", "..", "testdata", "fixtures", "JSONConnectorFail", "autobox")
	source, err := os.ReadFile(base + ".pas")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(base + ".txt")
	if err != nil {
		t.Fatal(err)
	}
	got := Compile(string(source), "autobox.pas", semantic.HintsLevelPedantic).DiagnosticStrings()
	if expected := strings.Split(strings.TrimSpace(string(want)), "\n"); !reflect.DeepEqual(got, expected) {
		t.Fatalf("diagnostics = %q, want %q", got, expected)
	}
}

func TestCompile_AnonymousCallableJSONTargets(t *testing.T) {
	for _, supplier := range []struct{ name, declaration, rhs, caption string }{
		{"procedure result", "var source: function: procedure;", "source", "procedure "},
		{"explicit procedure result", "var source: function: procedure;", "source()", "procedure "},
		{"function result", "var source: function: function: Integer;", "source", "function : Integer"},
		{"method result", "var source: function: procedure of object;", "source", "procedure "},
		{"field result", "type THolder = class Factory: function: procedure; end; var holder: THolder;", "holder.Factory", "procedure "},
		{"index result", "type TFactory = function: procedure; var factories: array[0..0] of TFactory;", "factories[0]", "procedure "},
		{"class literal", "", "TObject", "class of TObject"},
		{"class alias", "type TMetaAlias = class of TObject; var source: TMetaAlias;", "source", "TMetaAlias"},
	} {
		for _, target := range []struct {
			name, declaration, lhs string
			column                 int
		}{
			{"variable", "var target: JSONVariant;", "target", 8},
			{"field", "type THost = class Value: JSONVariant; end; var host: THost;", "host.Value", 12},
			{"field property", "type THost = class F: JSONVariant; property Value: JSONVariant write F; end; var host: THost;", "host.Value", 12},
			{"method property", "type THost = class procedure Put(v: JSONVariant); begin end; property Value: JSONVariant write Put; end; var host: THost;", "host.Value", 12},
			{"array index", "var target: array[0..0] of JSONVariant;", "target[0]", 11},
			{"dynamic index", "var target: array of JSONVariant;", "target[0]", 11},
			{"indexed property", "type THost = class procedure Put(i: Integer; v: JSONVariant); begin end; property Items[i: Integer]: JSONVariant write Put; end; var host: THost;", "host.Items[0]", 15},
		} {
			t.Run(supplier.name+"/"+target.name, func(t *testing.T) {
				source := supplier.declaration + "\n" + target.declaration + "\n" + target.lhs + " := " + supplier.rhs + ";"
				got := Compile(source, "json.pas", semantic.HintsLevelDisabled).DiagnosticStrings()
				want := []string{fmt.Sprintf("Syntax Error: Incompatible types: Cannot assign %q to \"JSONVariant\" [line: 3, column: %d]", supplier.caption, target.column)}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("diagnostics = %q, want %q", got, want)
				}
			})
		}
	}
}

func TestCompile_AnonymousCallableJSONScalars(t *testing.T) {
	source := `var j: JSONVariant; var v: Variant;
j := 12; j := 1.5; j := 'text'; j := True; j := v; j := j; j := nil;`
	if got := Compile(source, "scalars.pas", semantic.HintsLevelDisabled).DiagnosticStrings(); len(got) != 0 {
		t.Fatal(got)
	}
}

func TestCompile_AnonymousCallableArrayParameter(t *testing.T) {
	// This valid nested signature formerly belonged to the malformed-input
	// crash suite because string-based type resolution could not represent it.
	result := Compile("procedure t(const a : Array Of function : procedure);\nbegin\nend;\n", "array_callback.pas", semantic.HintsLevelPedantic)
	if got := result.DiagnosticStrings(); len(got) != 0 || !result.SemanticSuccessful {
		t.Fatalf("nested signature rejected: %v", got)
	}
	if result.Program == nil {
		t.Fatal("missing compiled program")
	}
	_ = result.Program.String()
}

func TestCompile_AnonymousCallableRecordStaticJSON(t *testing.T) {
	for _, supplier := range []string{"R.Stored", "R.Factory", "item.Stored", "item.Factory", "meta.Stored", "meta.Factory"} {
		t.Run(supplier, func(t *testing.T) {
			source := "type TProc = procedure; type TFactory = function: TProc;\n" +
				"type R = record class var Stored: TFactory; property Factory: TFactory read Stored write Stored; end;\n" +
				"var item: R; var meta := R; var target: JSONVariant;\n" +
				"target := " + supplier + ";"
			want := []string{`Syntax Error: Incompatible types: Cannot assign "procedure TProc" to "JSONVariant" [line: 4, column: 8]`}
			got := Compile(source, "record_static_json.pas", semantic.HintsLevelDisabled).DiagnosticStrings()
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("diagnostics = %q, want %q", got, want)
			}
		})
	}
}

func TestCompile_AnonymousCallablePointerReceiverJSON(t *testing.T) {
	source := "type TProc = procedure; type TFactory = function: TProc;\n" +
		"type THolder = class Factory: TFactory; end;\n" +
		"var receiver: function: THolder; var target: JSONVariant;\n" +
		"target := receiver.Factory;"
	want := []string{`Syntax Error: Incompatible types: Cannot assign "procedure TProc" to "JSONVariant" [line: 4, column: 8]`}
	got := Compile(source, "pointer_receiver_json.pas", semantic.HintsLevelDisabled).DiagnosticStrings()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("diagnostics = %q, want %q", got, want)
	}
}
