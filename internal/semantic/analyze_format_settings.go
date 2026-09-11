package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// DateTimeZoneTypeName is the name of the engine-provided scoped enumeration
// that selects how date/time values are interpreted and rendered.
const DateTimeZoneTypeName = "DateTimeZone"

// FormatSettingsTypeName is the name of the engine-provided static class that
// exposes the running script's date/time locale settings.
const FormatSettingsTypeName = "FormatSettings"

// FormatSettingsClassVars lists the FormatSettings class variables in
// declaration order. The runtime bootstrap uses the same list, so the two
// sides cannot drift apart.
var FormatSettingsClassVars = []string{
	"ShortDateFormat",
	"LongDateFormat",
	"ShortTimeFormat",
	"LongTimeFormat",
	"TimeAMString",
	"TimePMString",
	"Zone",
}

// DateTimeZoneElements lists the DateTimeZone members in ordinal order.
// Default (0) defers to FormatSettings.Zone, Local (1) means local wall-clock
// time and UTC (2) means Coordinated Universal Time.
var DateTimeZoneElements = []string{"Default", "Local", "UTC"}

// NewDateTimeZoneType builds the DateTimeZone enumeration type.
func NewDateTimeZoneType() *types.EnumType {
	values := make(map[string]int, len(DateTimeZoneElements))
	ordered := make([]string, 0, len(DateTimeZoneElements))
	for ordinal, name := range DateTimeZoneElements {
		values[name] = ordinal
		ordered = append(ordered, name)
	}
	return types.NewScopedEnumType(DateTimeZoneTypeName, values, ordered, false)
}

// registerDateTimeFormatSettings registers the DateTimeZone enumeration and the
// FormatSettings static class, the two script-visible symbols behind DWScript's
// date/time locale handling. Their runtime counterparts are created by the
// interpreter bootstrap.
func (a *Analyzer) registerDateTimeFormatSettings() {
	zone := NewDateTimeZoneType()
	a.registerBuiltinType(DateTimeZoneTypeName, zone)
	a.symbols.Define(DateTimeZoneTypeName, zone, token.Position{})
	a.createEnumScopedAccessHelper(DateTimeZoneTypeName, zone)

	settings := types.NewClassType(FormatSettingsTypeName, nil)
	for _, name := range FormatSettingsClassVars {
		normalized := ident.Normalize(name)
		varType := types.Type(types.STRING)
		if ident.Equal(name, "Zone") {
			varType = zone
		}
		settings.ClassVars[normalized] = varType
		settings.ClassVarVisibility[normalized] = int(ast.VisibilityPublic)
		settings.ClassVarDeclNames[normalized] = name
	}
	a.registerBuiltinType(FormatSettingsTypeName, settings)
}
