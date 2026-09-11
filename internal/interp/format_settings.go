package interp

import (
	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// registerDateTimeFormatSettings creates the runtime counterparts of the
// script-visible DateTimeZone enumeration and FormatSettings static class.
//
// FormatSettings is a class whose class variables are the only storage for the
// running script's date/time locale settings: scripts read and write them
// through the ordinary class-variable paths, and the date/time built-ins read
// them back through builtins.Context.DateTimeFormatSettings.
func (i *Interpreter) registerDateTimeFormatSettings(env *runtime.Environment) {
	zone := semantic.NewDateTimeZoneType()
	zoneTypeValue := runtime.NewEnumTypeValue(zone)
	env.Define("__enum_type_"+ident.Normalize(semantic.DateTimeZoneTypeName), zoneTypeValue)
	i.typeSystem.RegisterEnumType(semantic.DateTimeZoneTypeName, zoneTypeValue)
	env.Define(semantic.DateTimeZoneTypeName, &TypeMetaValue{
		TypeInfo: zone,
		TypeName: semantic.DateTimeZoneTypeName,
	})

	class := NewClassInfo(semantic.FormatSettingsTypeName)
	defaults := builtins.DefaultDateTimeFormatSettings()
	// The names come from the analyzer so the two sides of the symbol cannot
	// drift apart: a variable the analyzer knows about but the runtime does not
	// would resolve at compile time and then read back as nil.
	for _, name := range semantic.FormatSettingsClassVars {
		class.ClassVars[name] = formatSettingsDefault(name, &defaults)
	}
	i.typeSystem.RegisterClass(semantic.FormatSettingsTypeName, class)
	// Binding the name lets assignments resolve FormatSettings as an lvalue
	// object, the same way a declared class does.
	class.DefineInEnv(env)
}

// formatSettingsDefault returns the initial runtime value of one FormatSettings
// class variable. Zone is stored as its plain ordinal; a script that assigns
// DateTimeZone.Local or DateTimeZone.UTC replaces it with an enum value, and
// the reader in the evaluator accepts either form.
func formatSettingsDefault(name string, defaults *builtins.DateTimeFormatSettings) runtime.Value {
	switch {
	case ident.Equal(name, "ShortDateFormat"):
		return &StringValue{Value: defaults.ShortDateFormat}
	case ident.Equal(name, "LongDateFormat"):
		return &StringValue{Value: defaults.LongDateFormat}
	case ident.Equal(name, "ShortTimeFormat"):
		return &StringValue{Value: defaults.ShortTimeFormat}
	case ident.Equal(name, "LongTimeFormat"):
		return &StringValue{Value: defaults.LongTimeFormat}
	case ident.Equal(name, "TimeAMString"):
		return &StringValue{Value: defaults.TimeAMString}
	case ident.Equal(name, "TimePMString"):
		return &StringValue{Value: defaults.TimePMString}
	case ident.Equal(name, "Zone"):
		return &IntegerValue{Value: int64(defaults.Zone)}
	default:
		return &NilValue{}
	}
}
