package interp

import "testing"

func TestCompoundAssignmentVariantRHSRuntime(t *testing.T) {
	for _, op := range []struct{ symbol, result string }{
		{"+=", "14"}, {"-=", "10"}, {"*=", "24"}, {"/=", "6"},
	} {
		for _, targetType := range []string{"Integer", "Float"} {
			for _, value := range []string{"2", "2.0", "'2'"} {
				t.Run(targetType+op.symbol+value, func(t *testing.T) {
					source := "var target: " + targetType + " := 12;\nvar value: Variant := " + value + ";\ntarget " + op.symbol + " value;\nPrintLn(target = " + op.result + ");"
					assertOutput(t, runQuickwinScript(t, source), "True\n")
				})
			}
		}
	}
}

func TestCompoundAssignmentVariantRHSConvertsBeforeOperation(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `var target: Integer := 10;
var value: Variant := 2.4;
target *= value;
PrintLn(target);
PrintLn(value);`), "20\n2.4\n")
}

func TestCompoundAssignmentVariantRHSFailurePreservesTarget(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `var target: Integer := 10;
var value: Variant := 'bad';
try target += value;
except PrintLn(target); end;
PrintLn(target);`), "10\n10\n")
}

func TestCompoundAssignmentRaisingConversionPreservesTarget(t *testing.T) {
	const conversion = `var original := Exception.Create('conversion raised');
var conversions: Integer;
var rhs: Variant := 2;
function ConvertValue(value: Variant): Integer;
begin
   Inc(conversions);
   raise original;
end;
operator implicit (Variant): Integer uses ConvertValue;
`
	for _, tt := range []struct{ name, setup, action, read string }{
		{"variable", "var target: Integer := 10;", "target += rhs", "target"},
		{"var parameter", `procedure Update(var target: Integer);
begin target += rhs; end;
var target: Integer := 10;`, "Update(target)", "target"},
		{"implicit field", `type THolder = class
   Value: Integer;
   procedure Update;
   begin Value += rhs; end;
end;
var holder := THolder.Create;
holder.Value := 10;`, "holder.Update", "holder.Value"},
		{"implicit property", `var reads, writes: Integer;
type THolder = class
   Value: Integer;
   function GetItem: Integer;
   begin Inc(reads); Result := Value; end;
   procedure SetItem(value: Integer);
   begin Inc(writes); Self.Value := value; end;
   property Item: Integer read GetItem write SetItem;
   procedure Update;
   begin Item += rhs; end;
end;
var holder := THolder.Create;
holder.Value := 10;`, "holder.Update", "holder.Value"},
		{"class variable via Self", `type THolder = class
   class var Value: Integer;
   procedure Update;
   begin Value += rhs; end;
end;
var holder := THolder.Create;
THolder.Value := 10;`, "holder.Update", "THolder.Value"},
		{"captured member receiver", `var receivers: Integer;
type THolder = class Value: Integer; end;
var holder := THolder.Create;
holder.Value := 10;
function GetHolder: THolder;
begin Inc(receivers); Result := holder; end;`, "GetHolder.Value += rhs", "holder.Value"},
		{"captured array index", `var indices: Integer;
var targets: array of Integer := [10];
function NextIndex: Integer;
begin Inc(indices); Result := 0; end;`, "targets[NextIndex()] += rhs", "targets[0]"},
		{"class variable via class context", `type THolder = class
   class var Value: Integer;
   class procedure Update;
   begin Value += rhs; end;
end;
THolder.Value := 10;`, "THolder.Update", "THolder.Value"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := conversion + tt.setup + "\ntry " + tt.action + `;
except on E: Exception do begin
   PrintLn(E = original);
   PrintLn(E.Message);
end; end;
PrintLn(` + tt.read + `);
PrintLn(conversions);`
			if tt.name == "implicit property" {
				source += "\nPrintLn(reads);\nPrintLn(writes);"
				assertOutput(t, runQuickwinScript(t, source), "True\nconversion raised\n10\n1\n1\n0\n")
				return
			}
			if tt.name == "captured member receiver" || tt.name == "captured array index" {
				counter := "receivers"
				if tt.name == "captured array index" {
					counter = "indices"
				}
				source += "\nPrintLn(" + counter + ");"
				assertOutput(t, runQuickwinScript(t, source), "True\nconversion raised\n10\n1\n1\n")
				return
			}
			assertOutput(t, runQuickwinScript(t, source), "True\nconversion raised\n10\n1\n")
		})
	}
}
