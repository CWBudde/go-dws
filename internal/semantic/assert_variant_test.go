package semantic

import "testing"

func TestAssertConditionTypes(t *testing.T) {
	for _, source := range []string{
		`Assert(True);`,
		`var v: Variant; Assert(v);`,
		`var v: Variant := 1; aSsErT(v, 'condition');`,
		`type TCondition = Variant; var v: TCondition := True; Assert(v);`,
		`type TCondition = Boolean; var v: TCondition := True; Assert(v);`,
		`function Condition: Variant; begin Result := True; end; Assert(Condition());`,
	} {
		t.Run(source, func(t *testing.T) { expectNoErrors(t, source) })
	}
}

func TestAssertRejectsInvalidArguments(t *testing.T) {
	for _, tt := range []struct{ source, diagnostic string }{
		{`Assert(1);`, "first argument must be Boolean"},
		{`Assert('true');`, "first argument must be Boolean"},
		{`Assert(1.0);`, "first argument must be Boolean"},
		{`Assert(True, 1);`, "second argument must be String"},
		{`var v: Variant := True; Assert(v, 1);`, "second argument must be String"},
		{`Assert();`, "expects 1-2 arguments"},
		{`Assert(True, 'message', 1);`, "expects 1-2 arguments"},
	} {
		t.Run(tt.source, func(t *testing.T) { expectError(t, tt.source, tt.diagnostic) })
	}
}
