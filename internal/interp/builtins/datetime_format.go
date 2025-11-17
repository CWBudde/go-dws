package builtins

import (
	"time"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Formatting Functions
// =============================================================================

// FormatDateTime implements the FormatDateTime() built-in function.
// Formats a TDateTime according to a format string.
// FormatDateTime(format: String, dt: TDateTime): String
func FormatDateTime(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("FormatDateTime() expects 2 arguments (format, dt), got %d", len(args))
	}

	// First argument: format string
	formatVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("FormatDateTime() expects String as first argument, got %s", args[0].Type())
	}

	// Second argument: TDateTime value
	dtVal, ok := args[1].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("FormatDateTime() expects Float/TDateTime as second argument, got %s", args[1].Type())
	}

	result := formatDateTime(formatVal.Value, dtVal.Value)
	return &runtime.StringValue{Value: result}
}

// DateTimeToStr implements the DateTimeToStr() built-in function.
// Converts a TDateTime to a string using default format.
func DateTimeToStr(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DateTimeToStr() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("DateTimeToStr() expects Float/TDateTime, got %s", args[0].Type())
	}

	// Use default format: YYYY-MM-DD HH:MM:SS
	result := formatDateTime("yyyy-mm-dd hh:nn:ss", dtVal.Value)
	return &runtime.StringValue{Value: result}
}

// DateToStr implements the DateToStr() built-in function.
// Converts a TDateTime to a date string.
func DateToStr(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DateToStr() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("DateToStr() expects Float/TDateTime, got %s", args[0].Type())
	}

	// Use default date format: YYYY-MM-DD
	result := formatDateTime("yyyy-mm-dd", dtVal.Value)
	return &runtime.StringValue{Value: result}
}

// TimeToStr implements the TimeToStr() built-in function.
// Converts a TDateTime to a time string.
func TimeToStr(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("TimeToStr() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("TimeToStr() expects Float/TDateTime, got %s", args[0].Type())
	}

	// Use default time format: HH:MM:SS
	result := formatDateTime("hh:nn:ss", dtVal.Value)
	return &runtime.StringValue{Value: result}
}

// DateToISO8601 implements the DateToISO8601() built-in function.
// Formats date as ISO 8601 string (YYYY-MM-DD).
func DateToISO8601(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DateToISO8601() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("DateToISO8601() expects Float/TDateTime, got %s", args[0].Type())
	}

	result := formatDateISO8601(dtVal.Value)
	return &runtime.StringValue{Value: result}
}

// DateTimeToISO8601 implements the DateTimeToISO8601() built-in function.
// Formats datetime as ISO 8601 string (YYYY-MM-DDTHH:MM:SS).
func DateTimeToISO8601(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DateTimeToISO8601() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("DateTimeToISO8601() expects Float/TDateTime, got %s", args[0].Type())
	}

	result := formatISO8601(dtVal.Value)
	return &runtime.StringValue{Value: result}
}

// DateTimeToRFC822 implements the DateTimeToRFC822() built-in function.
// Formats datetime as RFC 822 string.
func DateTimeToRFC822(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DateTimeToRFC822() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("DateTimeToRFC822() expects Float/TDateTime, got %s", args[0].Type())
	}

	result := formatRFC822(dtVal.Value)
	return &runtime.StringValue{Value: result}
}

// =============================================================================
// Parsing Functions
// =============================================================================

// StrToDate implements the StrToDate() built-in function.
// Parses a date string to TDateTime.
func StrToDate(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("StrToDate() expects 1 argument, got %d", len(args))
	}

	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrToDate() expects String, got %s", args[0].Type())
	}

	dt, err := parseDate(strVal.Value)
	if err != nil {
		return ctx.NewError("StrToDate() %s", err)
	}

	return &runtime.FloatValue{Value: dt}
}

// StrToDateTime implements the StrToDateTime() built-in function.
// Parses a datetime string to TDateTime.
func StrToDateTime(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("StrToDateTime() expects 1 argument, got %d", len(args))
	}

	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrToDateTime() expects String, got %s", args[0].Type())
	}

	dt, err := parseDateTime(strVal.Value)
	if err != nil {
		return ctx.NewError("StrToDateTime() %s", err)
	}

	return &runtime.FloatValue{Value: dt}
}

// StrToTime implements the StrToTime() built-in function.
// Parses a time string to TDateTime.
func StrToTime(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("StrToTime() expects 1 argument, got %d", len(args))
	}

	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrToTime() expects String, got %s", args[0].Type())
	}

	dt, err := parseTime(strVal.Value)
	if err != nil {
		return ctx.NewError("StrToTime() %s", err)
	}

	return &runtime.FloatValue{Value: dt}
}

// ISO8601ToDateTime implements the ISO8601ToDateTime() built-in function.
// Parses an ISO 8601 string to TDateTime.
func ISO8601ToDateTime(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("ISO8601ToDateTime() expects 1 argument, got %d", len(args))
	}

	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("ISO8601ToDateTime() expects String, got %s", args[0].Type())
	}

	dt, err := parseISO8601(strVal.Value)
	if err != nil {
		return ctx.NewError("ISO8601ToDateTime() %s", err)
	}

	return &runtime.FloatValue{Value: dt}
}

// RFC822ToDateTime implements the RFC822ToDateTime() built-in function.
// Parses an RFC 822 string to TDateTime.
func RFC822ToDateTime(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("RFC822ToDateTime() expects 1 argument, got %d", len(args))
	}

	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("RFC822ToDateTime() expects String, got %s", args[0].Type())
	}

	dt, err := parseRFC822(strVal.Value)
	if err != nil {
		return ctx.NewError("RFC822ToDateTime() %s", err)
	}

	return &runtime.FloatValue{Value: dt}
}

// =============================================================================
// Unix Time Functions
// =============================================================================

// UnixTime implements the UnixTime() built-in function.
// Returns Unix timestamp (seconds since 1970-01-01) for current time.
func UnixTime(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("UnixTime() expects 0 arguments, got %d", len(args))
	}

	now := time.Now().UTC()
	return &runtime.IntegerValue{Value: now.Unix()}
}

// UnixTimeMSec implements the UnixTimeMSec() built-in function.
// Returns Unix timestamp in milliseconds for current time.
func UnixTimeMSec(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("UnixTimeMSec() expects 0 arguments, got %d", len(args))
	}

	now := time.Now().UTC()
	return &runtime.IntegerValue{Value: now.UnixMilli()}
}

// UnixTimeToDateTime implements the UnixTimeToDateTime() built-in function.
// Converts Unix timestamp to TDateTime.
func UnixTimeToDateTime(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("UnixTimeToDateTime() expects 1 argument, got %d", len(args))
	}

	unixTimeVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("UnixTimeToDateTime() expects Integer, got %s", args[0].Type())
	}

	dt := unixTimeToDateTime(unixTimeVal.Value)
	return &runtime.FloatValue{Value: dt}
}

// DateTimeToUnixTime implements the DateTimeToUnixTime() built-in function.
// Converts TDateTime to Unix timestamp.
func DateTimeToUnixTime(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DateTimeToUnixTime() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("DateTimeToUnixTime() expects Float/TDateTime, got %s", args[0].Type())
	}

	unixTime := dateTimeToUnixTime(dtVal.Value)
	return &runtime.IntegerValue{Value: unixTime}
}

// UnixTimeMSecToDateTime implements the UnixTimeMSecToDateTime() built-in function.
// Converts Unix timestamp in milliseconds to TDateTime.
func UnixTimeMSecToDateTime(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("UnixTimeMSecToDateTime() expects 1 argument, got %d", len(args))
	}

	unixTimeMSVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("UnixTimeMSecToDateTime() expects Integer, got %s", args[0].Type())
	}

	dt := unixTimeMSecToDateTime(unixTimeMSVal.Value)
	return &runtime.FloatValue{Value: dt}
}

// DateTimeToUnixTimeMSec implements the DateTimeToUnixTimeMSec() built-in function.
// Converts TDateTime to Unix timestamp in milliseconds.
func DateTimeToUnixTimeMSec(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DateTimeToUnixTimeMSec() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("DateTimeToUnixTimeMSec() expects Float/TDateTime, got %s", args[0].Type())
	}

	unixTimeMS := dateTimeToUnixTimeMSec(dtVal.Value)
	return &runtime.IntegerValue{Value: unixTimeMS}
}
