package interp

import "testing"

// TestFormatSettings_DefaultsAreVisible checks that the engine-provided
// FormatSettings class exposes DWScript's default locale strings to a script.
func TestFormatSettings_DefaultsAreVisible(t *testing.T) {
	runScriptTestWithSemantic(t, `
PrintLn(FormatSettings.ShortDateFormat);
PrintLn(FormatSettings.ShortTimeFormat);
`, "yyyy-mm-dd\nhh:nn")
}

// TestFormatSettings_AssignmentDrivesFormatting checks that FormatSettings is
// mutable and that the date/time built-ins observe the assignment: the class
// variables are the only storage, so a write has to be visible on the very next
// call.
func TestFormatSettings_AssignmentDrivesFormatting(t *testing.T) {
	runScriptTestWithSemantic(t, `
var d := EncodeDate(2019, 1, 31);
PrintLn(DateToStr(d));
FormatSettings.ShortDateFormat := 'dd/mm/yyyy';
PrintLn(DateToStr(d));
`, "2019-01-31\n31/01/2019")
}

// TestFormatSettings_ZoneAcceptsEnumAssignment checks that FormatSettings.Zone
// selects the default zone used by a built-in that is not given one, and that
// it round-trips a DateTimeZone enum value: the bootstrap seeds the class
// variable with a plain ordinal, so both representations have to be readable.
//
// The assertions compare the implicit rendering against the explicit one rather
// than against a literal, so the test does not depend on the host's zone.
func TestFormatSettings_ZoneAcceptsEnumAssignment(t *testing.T) {
	runScriptTestWithSemantic(t, `
var d := EncodeDate(2019, 6, 1, DateTimeZone.UTC) + EncodeTime(12, 0, 0, 0);
FormatSettings.Zone := DateTimeZone.UTC;
PrintLn(DateTimeToStr(d) = DateTimeToStr(d, DateTimeZone.UTC));
FormatSettings.Zone := DateTimeZone.Local;
PrintLn(DateTimeToStr(d) = DateTimeToStr(d, DateTimeZone.Local));
`, "True\nTrue")
}
