package interp

import "testing"

func TestAssertVariant_CompiledConversions(t *testing.T) {
	for _, tt := range []struct {
		name, initializer string
		passes            bool
	}{
		{"true", "True", true},
		{"positive integer", "1", true},
		{"negative integer", "-1", true},
		{"float", "0.5", true},
		{"true string", "'true'", true},
		{"numeric string", "'-2'", true},
		{"false", "False", false},
		{"zero integer", "0", false},
		{"zero float", "0.0", false},
		{"uninitialized", "", false},
		{"unassigned", "Unassigned", false},
		{"null", "Null", false},
		{"false string", "'false'", false},
		{"empty string", "''", false},
		{"invalid boolean string", "'not a boolean'", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			declaration := "var condition: Variant"
			if tt.initializer != "" {
				declaration += " := " + tt.initializer
			}
			source := declaration + `;
try
  Assert(condition, 'condition failed');
  PrintLn('passed');
except
  on E: EAssertionFailed do PrintLn(E.Message);
end;`
			want := "Assertion failed [line: 3, column: 3] : condition failed\n"
			if tt.passes {
				want = "passed\n"
			}
			assertOutput(t, runQuickwinScript(t, source), want)
		})
	}
}

func TestAssertVariant_ConditionEvaluatedOnce(t *testing.T) {
	source := `var calls := 0;
function Condition: Variant;
begin
  Inc(calls);
  Result := calls = 1;
end;
Assert(Condition());
try
  Assert(Condition());
except
  on E: EAssertionFailed do PrintLn(E.Message);
end;
PrintLn(calls);`
	assertOutput(t, runQuickwinScript(t, source), "Assertion failed [line: 9, column: 3]\n2\n")
}

func TestAssertVariant_ConditionAliases(t *testing.T) {
	for _, conditionType := range []string{"Variant", "Boolean"} {
		t.Run(conditionType, func(t *testing.T) {
			source := `type TCondition = ` + conditionType + `;
type TAlias = TCondition;
var condition: TAlias := True;
Assert(condition);
function ConditionResult: TAlias;
begin Result := False; end;
try
  Assert(ConditionResult(), 'alias');
except
  on E: EAssertionFailed do PrintLn(E.Message);
end;
PrintLn('passed');`
			assertOutput(t, runQuickwinScript(t, source), "Assertion failed [line: 8, column: 3] : alias\npassed\n")
		})
	}
}
