package builtins

import (
	"math"
)

// =============================================================================
// Formatting Functions
// =============================================================================

// FormatDateTime implements the FormatDateTime() built-in function.
// FormatDateTime(format: String; dt: TDateTime; tz: DateTimeZone = Default): String
func FormatDateTime(ctx Context, args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		return ctx.NewError("FormatDateTime() expects 2 or 3 arguments (format, dt, [zone]), got %d", len(args))
	}
	format, errVal := stringArg(ctx, args, 0, "FormatDateTime")
	if errVal != nil {
		return errVal
	}
	dt, errVal := dateTimeArg(ctx, args, 1, "FormatDateTime")
	if errVal != nil {
		return errVal
	}
	result, err := formatDateTimeSettings(format, dt, settings(ctx), zoneArg(ctx, args, 2))
	if err != nil {
		return ctx.NewError("%s", err.Error())
	}
	return stringResult(result)
}

// DateTimeToStr implements the DateTimeToStr() built-in function, using
// FormatSettings.ShortDateFormat and FormatSettings.LongTimeFormat.
func DateTimeToStr(ctx Context, args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return ctx.NewError("DateTimeToStr() expects 1 or 2 arguments, got %d", len(args))
	}
	dt, errVal := dateTimeArg(ctx, args, 0, "DateTimeToStr")
	if errVal != nil {
		return errVal
	}
	result, err := dateTimeToStrSettings(dt, settings(ctx), zoneArg(ctx, args, 1))
	if err != nil {
		return ctx.NewError("%s", err.Error())
	}
	return stringResult(result)
}

// DateToStr implements the DateToStr() built-in function, using
// FormatSettings.ShortDateFormat.
func DateToStr(ctx Context, args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return ctx.NewError("DateToStr() expects 1 or 2 arguments, got %d", len(args))
	}
	dt, errVal := dateTimeArg(ctx, args, 0, "DateToStr")
	if errVal != nil {
		return errVal
	}
	result, err := dateToStrSettings(dt, settings(ctx), zoneArg(ctx, args, 1))
	if err != nil {
		return ctx.NewError("%s", err.Error())
	}
	return stringResult(result)
}

// TimeToStr implements the TimeToStr() built-in function, using
// FormatSettings.LongTimeFormat.
func TimeToStr(ctx Context, args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return ctx.NewError("TimeToStr() expects 1 or 2 arguments, got %d", len(args))
	}
	dt, errVal := dateTimeArg(ctx, args, 0, "TimeToStr")
	if errVal != nil {
		return errVal
	}
	result, err := timeToStrSettings(dt, settings(ctx), zoneArg(ctx, args, 1))
	if err != nil {
		return ctx.NewError("%s", err.Error())
	}
	return stringResult(result)
}

// DateToISO8601 implements the DateToISO8601() built-in function.
func DateToISO8601(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DateToISO8601() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArg(ctx, args, 0, "DateToISO8601")
	if errVal != nil {
		return errVal
	}
	return stringResult(formatDateISO8601(dt))
}

// DateTimeToISO8601 implements the DateTimeToISO8601() built-in function.
// DateTimeToISO8601(dt: TDateTime; precision: String = ”): String
// The precision argument accepts 'sec' and 'msec'; anything else selects the
// automatic precision, which omits zero seconds.
func DateTimeToISO8601(ctx Context, args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return ctx.NewError("DateTimeToISO8601() expects 1 or 2 arguments, got %d", len(args))
	}
	dt, errVal := dateTimeArg(ctx, args, 0, "DateTimeToISO8601")
	if errVal != nil {
		return errVal
	}
	prec := ISO8601PrecAuto
	if len(args) == 2 {
		precStr, errVal := stringArg(ctx, args, 1, "DateTimeToISO8601")
		if errVal != nil {
			return errVal
		}
		prec = parseISO8601Precision(precStr)
	}
	return stringResult(formatISO8601(dt, prec))
}

// DateTimeToRFC822 implements the DateTimeToRFC822() built-in function.
func DateTimeToRFC822(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DateTimeToRFC822() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArg(ctx, args, 0, "DateTimeToRFC822")
	if errVal != nil {
		return errVal
	}
	return stringResult(formatRFC822(dt))
}

// =============================================================================
// Parsing Functions
// =============================================================================

// StrToDate implements the StrToDate() built-in function.
// StrToDate(str: String; tz: DateTimeZone = Default): TDateTime
func StrToDate(ctx Context, args []Value) Value {
	return parseWith(ctx, args, "StrToDate", tryStrToDateSettings)
}

// StrToTime implements the StrToTime() built-in function.
// StrToTime(str: String; tz: DateTimeZone = Default): TDateTime
func StrToTime(ctx Context, args []Value) Value {
	return parseWith(ctx, args, "StrToTime", tryStrToTimeSettings)
}

// StrToDateTime implements the StrToDateTime() built-in function. It falls back
// to date-only parsing when the combined date and time formats do not match.
func StrToDateTime(ctx Context, args []Value) Value {
	return parseWith(ctx, args, "StrToDateTime", func(str string, s *DateTimeFormatSettings, tz TimeZone) (float64, bool) {
		if dt, ok := tryStrToDateTimeSettings(str, s, tz); ok {
			return dt, true
		}
		return tryStrToDateSettings(str, s, tz)
	})
}

// StrToDateDef implements the StrToDateDef() built-in function.
// StrToDateDef(str: String; def: TDateTime; tz: DateTimeZone = Default): TDateTime
func StrToDateDef(ctx Context, args []Value) Value {
	return parseWithDefault(ctx, args, "StrToDateDef", tryStrToDateSettings)
}

// StrToTimeDef implements the StrToTimeDef() built-in function.
func StrToTimeDef(ctx Context, args []Value) Value {
	return parseWithDefault(ctx, args, "StrToTimeDef", tryStrToTimeSettings)
}

// StrToDateTimeDef implements the StrToDateTimeDef() built-in function.
func StrToDateTimeDef(ctx Context, args []Value) Value {
	return parseWithDefault(ctx, args, "StrToDateTimeDef",
		func(str string, s *DateTimeFormatSettings, tz TimeZone) (float64, bool) {
			if dt, ok := tryStrToDateTimeSettings(str, s, tz); ok {
				return dt, true
			}
			return tryStrToDateSettings(str, s, tz)
		})
}

// dateTimeTryParser is the shared shape of the settings-driven parsers.
type dateTimeTryParser func(str string, s *DateTimeFormatSettings, tz TimeZone) (float64, bool)

// parseWith implements the raising StrTo* family.
func parseWith(ctx Context, args []Value, fn string, parse dateTimeTryParser) Value {
	if len(args) < 1 || len(args) > 2 {
		return ctx.NewError("%s() expects 1 or 2 arguments, got %d", fn, len(args))
	}
	str, errVal := stringArg(ctx, args, 0, fn)
	if errVal != nil {
		return errVal
	}
	dt, ok := parse(str, settings(ctx), zoneArg(ctx, args, 1))
	if !ok {
		return ctx.NewError("%s", dateTimeConversionError(str).Error())
	}
	return floatResult(dt)
}

// parseWithDefault implements the non-raising StrTo*Def family.
func parseWithDefault(ctx Context, args []Value, fn string, parse dateTimeTryParser) Value {
	if len(args) < 2 || len(args) > 3 {
		return ctx.NewError("%s() expects 2 or 3 arguments, got %d", fn, len(args))
	}
	str, errVal := stringArg(ctx, args, 0, fn)
	if errVal != nil {
		return errVal
	}
	def, errVal := dateTimeArg(ctx, args, 1, fn)
	if errVal != nil {
		return errVal
	}
	dt, ok := parse(str, settings(ctx), zoneArg(ctx, args, 2))
	if !ok {
		return floatResult(def)
	}
	return floatResult(dt)
}

// ParseDateTime implements the ParseDateTime() built-in function, parsing a
// string against an explicit format. It returns 0 when the string does not
// match, rather than raising.
// ParseDateTime(format, str: String; tz: DateTimeZone = Default): TDateTime
func ParseDateTime(ctx Context, args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		return ctx.NewError("ParseDateTime() expects 2 or 3 arguments (format, str, [zone]), got %d", len(args))
	}
	format, errVal := stringArg(ctx, args, 0, "ParseDateTime")
	if errVal != nil {
		return errVal
	}
	str, errVal := stringArg(ctx, args, 1, "ParseDateTime")
	if errVal != nil {
		return errVal
	}
	dt, ok := tryStrToDateTimeFormat(format, str, settings(ctx), zoneArg(ctx, args, 2))
	if !ok {
		return floatResult(0)
	}
	return floatResult(dt)
}

// ISO8601ToDateTime implements the ISO8601ToDateTime() built-in function.
func ISO8601ToDateTime(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("ISO8601ToDateTime() expects 1 argument, got %d", len(args))
	}
	str, errVal := stringArg(ctx, args, 0, "ISO8601ToDateTime")
	if errVal != nil {
		return errVal
	}
	dt, err := parseISO8601(str)
	if err != nil {
		return ctx.NewError("%s", err.Error())
	}
	return floatResult(dt)
}

// RFC822ToDateTime implements the RFC822ToDateTime() built-in function.
func RFC822ToDateTime(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("RFC822ToDateTime() expects 1 argument, got %d", len(args))
	}
	str, errVal := stringArg(ctx, args, 0, "RFC822ToDateTime")
	if errVal != nil {
		return errVal
	}
	return floatResult(parseRFC822(str))
}

// =============================================================================
// Unix Time Functions
// =============================================================================

// UnixTime implements the UnixTime() built-in function.
// With no argument it returns the current Unix timestamp in seconds.
func UnixTime(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("UnixTime() expects 0 arguments, got %d", len(args))
	}
	return int64Result(dateTimeToUnixTime(utcNowDateTime()))
}

// UnixTimeMSec implements the UnixTimeMSec() built-in function.
func UnixTimeMSec(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("UnixTimeMSec() expects 0 arguments, got %d", len(args))
	}
	return int64Result(dateTimeToUnixTimeMSec(utcNowDateTime()))
}

// UnixTimeToDateTime implements the UnixTimeToDateTime() built-in function,
// producing a UTC TDateTime.
func UnixTimeToDateTime(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("UnixTimeToDateTime() expects 1 argument, got %d", len(args))
	}
	ut, errVal := dateTimeArg(ctx, args, 0, "UnixTimeToDateTime")
	if errVal != nil {
		return errVal
	}
	return floatResult(ut/secondsPerDay + unixEpochDateTime)
}

// DateTimeToUnixTime implements the DateTimeToUnixTime() built-in function,
// interpreting its argument as a UTC TDateTime.
func DateTimeToUnixTime(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DateTimeToUnixTime() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArg(ctx, args, 0, "DateTimeToUnixTime")
	if errVal != nil {
		return errVal
	}
	return int64Result(dateTimeToUnixTime(dt))
}

// UnixTimeMSecToDateTime implements the UnixTimeMSecToDateTime() built-in function.
func UnixTimeMSecToDateTime(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("UnixTimeMSecToDateTime() expects 1 argument, got %d", len(args))
	}
	ut, errVal := integerArg(ctx, args, 0, "UnixTimeMSecToDateTime")
	if errVal != nil {
		return errVal
	}
	return floatResult(unixTimeMSecToDateTime(ut))
}

// DateTimeToUnixTimeMSec implements the DateTimeToUnixTimeMSec() built-in function.
func DateTimeToUnixTimeMSec(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DateTimeToUnixTimeMSec() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArg(ctx, args, 0, "DateTimeToUnixTimeMSec")
	if errVal != nil {
		return errVal
	}
	return int64Result(dateTimeToUnixTimeMSec(dt))
}

// LocalDateTimeToUnixTime implements the LocalDateTimeToUnixTime() built-in
// function, converting a local TDateTime to a Unix timestamp.
func LocalDateTimeToUnixTime(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("LocalDateTimeToUnixTime() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArg(ctx, args, 0, "LocalDateTimeToUnixTime")
	if errVal != nil {
		return errVal
	}
	return int64Result(int64(math.Round(localDateTimeToUTCDateTime(dt)*secondsPerDay)) -
		int64(unixEpochDateTime)*86400)
}

// UnixTimeToLocalDateTime implements the UnixTimeToLocalDateTime() built-in
// function, converting a Unix timestamp to a local TDateTime.
func UnixTimeToLocalDateTime(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("UnixTimeToLocalDateTime() expects 1 argument, got %d", len(args))
	}
	ut, errVal := integerArg(ctx, args, 0, "UnixTimeToLocalDateTime")
	if errVal != nil {
		return errVal
	}
	return floatResult(utcDateTimeToLocalDateTime(float64(ut)/secondsPerDay + unixEpochDateTime))
}
