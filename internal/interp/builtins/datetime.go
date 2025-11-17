package builtins

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Date Encoding Functions
// =============================================================================

// EncodeDate implements the EncodeDate() built-in function.
// Creates a TDateTime from year, month, day components.
// EncodeDate(year, month, day: Integer): TDateTime
func EncodeDate(ctx Context, args []Value) Value {
	if len(args) != 3 {
		return ctx.NewError("EncodeDate() expects 3 arguments (year, month, day), got %d", len(args))
	}

	// Extract year
	yearVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("EncodeDate() year must be Integer, got %s", args[0].Type())
	}

	// Extract month
	monthVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("EncodeDate() month must be Integer, got %s", args[1].Type())
	}

	// Extract day
	dayVal, ok := args[2].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("EncodeDate() day must be Integer, got %s", args[2].Type())
	}

	year := int(yearVal.Value)
	month := int(monthVal.Value)
	day := int(dayVal.Value)

	// Validate date
	if !isValidDate(year, month, day) {
		return ctx.NewError("EncodeDate() invalid date: %d-%02d-%02d", year, month, day)
	}

	// Create date (time = 00:00:00)
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	dtValue := goTimeToDelphiDateTime(t)

	return &runtime.FloatValue{Value: dtValue}
}

// EncodeTime implements the EncodeTime() built-in function.
// Creates a TDateTime from hour, minute, second, millisecond components.
// EncodeTime(hour, minute, second, msec: Integer): TDateTime
func EncodeTime(ctx Context, args []Value) Value {
	if len(args) != 4 {
		return ctx.NewError("EncodeTime() expects 4 arguments (hour, minute, second, msec), got %d", len(args))
	}

	// Extract hour
	hourVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("EncodeTime() hour must be Integer, got %s", args[0].Type())
	}

	// Extract minute
	minuteVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("EncodeTime() minute must be Integer, got %s", args[1].Type())
	}

	// Extract second
	secondVal, ok := args[2].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("EncodeTime() second must be Integer, got %s", args[2].Type())
	}

	// Extract millisecond
	msecVal, ok := args[3].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("EncodeTime() msec must be Integer, got %s", args[3].Type())
	}

	hour := int(hourVal.Value)
	minute := int(minuteVal.Value)
	second := int(secondVal.Value)
	msec := int(msecVal.Value)

	// Validate time
	if !isValidTime(hour, minute, second, msec) {
		return ctx.NewError("EncodeTime() invalid time: %02d:%02d:%02d.%03d", hour, minute, second, msec)
	}

	// Create time on epoch date
	nanoseconds := msec * 1000000
	t := time.Date(1899, 12, 30, hour, minute, second, nanoseconds, time.UTC)
	dtValue := goTimeToDelphiDateTime(t)

	return &runtime.FloatValue{Value: dtValue}
}

// EncodeDateTime implements the EncodeDateTime() built-in function.
// Creates a TDateTime from full date and time components.
// EncodeDateTime(year, month, day, hour, minute, second, msec: Integer): TDateTime
func EncodeDateTime(ctx Context, args []Value) Value {
	if len(args) != 7 {
		return ctx.NewError("EncodeDateTime() expects 7 arguments (year, month, day, hour, minute, second, msec), got %d", len(args))
	}

	// Extract all components
	yearVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("EncodeDateTime() year must be Integer, got %s", args[0].Type())
	}

	monthVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("EncodeDateTime() month must be Integer, got %s", args[1].Type())
	}

	dayVal, ok := args[2].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("EncodeDateTime() day must be Integer, got %s", args[2].Type())
	}

	hourVal, ok := args[3].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("EncodeDateTime() hour must be Integer, got %s", args[3].Type())
	}

	minuteVal, ok := args[4].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("EncodeDateTime() minute must be Integer, got %s", args[4].Type())
	}

	secondVal, ok := args[5].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("EncodeDateTime() second must be Integer, got %s", args[5].Type())
	}

	msecVal, ok := args[6].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("EncodeDateTime() msec must be Integer, got %s", args[6].Type())
	}

	year := int(yearVal.Value)
	month := int(monthVal.Value)
	day := int(dayVal.Value)
	hour := int(hourVal.Value)
	minute := int(minuteVal.Value)
	second := int(secondVal.Value)
	msec := int(msecVal.Value)

	// Validate date and time
	if !isValidDate(year, month, day) {
		return ctx.NewError("EncodeDateTime() invalid date: %d-%02d-%02d", year, month, day)
	}

	if !isValidTime(hour, minute, second, msec) {
		return ctx.NewError("EncodeDateTime() invalid time: %02d:%02d:%02d.%03d", hour, minute, second, msec)
	}

	// Create full datetime
	nanoseconds := msec * 1000000
	t := time.Date(year, time.Month(month), day, hour, minute, second, nanoseconds, time.UTC)
	dtValue := goTimeToDelphiDateTime(t)

	return &runtime.FloatValue{Value: dtValue}
}

// =============================================================================
// Date Decoding Functions (Var Parameters)
// =============================================================================

// TODO: DecodeDate - Requires special handling (takes []ast.Expression, modifies variables in-place)
// Original signature: func (i *Interpreter) builtinDecodeDate(args []ast.Expression) Value
// DecodeDate(dt: TDateTime; var year, month, day: Integer)

// TODO: DecodeTime - Requires special handling (takes []ast.Expression, modifies variables in-place)
// Original signature: func (i *Interpreter) builtinDecodeTime(args []ast.Expression) Value
// DecodeTime(dt: TDateTime; var hour, minute, second, msec: Integer)

// =============================================================================
// Incrementing Functions
// =============================================================================

// IncYear implements the IncYear() built-in function.
// Adds years to a TDateTime.
func IncYear(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("IncYear() expects 2 arguments (dt, years), got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("IncYear() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	yearsVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("IncYear() expects Integer as second argument, got %s", args[1].Type())
	}

	result := incYears(dtVal.Value, int(yearsVal.Value))
	return &runtime.FloatValue{Value: result}
}

// IncMonth implements the IncMonth() built-in function.
// Adds months to a TDateTime.
func IncMonth(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("IncMonth() expects 2 arguments (dt, months), got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("IncMonth() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	monthsVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("IncMonth() expects Integer as second argument, got %s", args[1].Type())
	}

	result := incMonths(dtVal.Value, int(monthsVal.Value))
	return &runtime.FloatValue{Value: result}
}

// IncDay implements the IncDay() built-in function.
// Adds days to a TDateTime.
func IncDay(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("IncDay() expects 2 arguments (dt, days), got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("IncDay() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	daysVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("IncDay() expects Integer as second argument, got %s", args[1].Type())
	}

	// Simple addition since TDateTime stores days as integer part
	result := dtVal.Value + float64(daysVal.Value)
	return &runtime.FloatValue{Value: result}
}

// IncHour implements the IncHour() built-in function.
// Adds hours to a TDateTime.
func IncHour(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("IncHour() expects 2 arguments (dt, hours), got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("IncHour() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	hoursVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("IncHour() expects Integer as second argument, got %s", args[1].Type())
	}

	// 1 hour = 1/24 day
	result := dtVal.Value + (float64(hoursVal.Value) / 24.0)
	return &runtime.FloatValue{Value: result}
}

// IncMinute implements the IncMinute() built-in function.
// Adds minutes to a TDateTime.
func IncMinute(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("IncMinute() expects 2 arguments (dt, minutes), got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("IncMinute() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	minutesVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("IncMinute() expects Integer as second argument, got %s", args[1].Type())
	}

	// 1 minute = 1/(24*60) day
	result := dtVal.Value + (float64(minutesVal.Value) / (24.0 * 60.0))
	return &runtime.FloatValue{Value: result}
}

// IncSecond implements the IncSecond() built-in function.
// Adds seconds to a TDateTime.
func IncSecond(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("IncSecond() expects 2 arguments (dt, seconds), got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("IncSecond() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	secondsVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("IncSecond() expects Integer as second argument, got %s", args[1].Type())
	}

	// 1 second = 1/86400 day
	result := dtVal.Value + (float64(secondsVal.Value) / 86400.0)
	return &runtime.FloatValue{Value: result}
}

// =============================================================================
// Date Difference Functions
// =============================================================================

// DaysBetween implements the DaysBetween() built-in function.
// Calculates whole days between two TDateTime values.
func DaysBetween(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("DaysBetween() expects 2 arguments (dt1, dt2), got %d", len(args))
	}

	dt1Val, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("DaysBetween() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	dt2Val, ok := args[1].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("DaysBetween() expects Float/TDateTime as second argument, got %s", args[1].Type())
	}

	days := daysBetween(dt1Val.Value, dt2Val.Value)
	return &runtime.IntegerValue{Value: int64(days)}
}

// HoursBetween implements the HoursBetween() built-in function.
// Calculates whole hours between two TDateTime values.
func HoursBetween(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("HoursBetween() expects 2 arguments (dt1, dt2), got %d", len(args))
	}

	dt1Val, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("HoursBetween() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	dt2Val, ok := args[1].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("HoursBetween() expects Float/TDateTime as second argument, got %s", args[1].Type())
	}

	hours := hoursBetween(dt1Val.Value, dt2Val.Value)
	return &runtime.IntegerValue{Value: int64(hours)}
}

// MinutesBetween implements the MinutesBetween() built-in function.
// Calculates whole minutes between two TDateTime values.
func MinutesBetween(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("MinutesBetween() expects 2 arguments (dt1, dt2), got %d", len(args))
	}

	dt1Val, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("MinutesBetween() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	dt2Val, ok := args[1].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("MinutesBetween() expects Float/TDateTime as second argument, got %s", args[1].Type())
	}

	minutes := minutesBetween(dt1Val.Value, dt2Val.Value)
	return &runtime.IntegerValue{Value: int64(minutes)}
}

// SecondsBetween implements the SecondsBetween() built-in function.
// Calculates whole seconds between two TDateTime values.
func SecondsBetween(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("SecondsBetween() expects 2 arguments (dt1, dt2), got %d", len(args))
	}

	dt1Val, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("SecondsBetween() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	dt2Val, ok := args[1].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("SecondsBetween() expects Float/TDateTime as second argument, got %s", args[1].Type())
	}

	seconds := secondsBetween(dt1Val.Value, dt2Val.Value)
	return &runtime.IntegerValue{Value: int64(seconds)}
}

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

// =============================================================================
// Current Date/Time Functions
// =============================================================================

// Now implements the Now() built-in function.
// Returns the current date and time as TDateTime.
func Now(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("Now() expects 0 arguments, got %d", len(args))
	}

	now := time.Now().UTC()
	dtValue := goTimeToDelphiDateTime(now)

	return &runtime.FloatValue{Value: dtValue}
}

// Date implements the Date() built-in function.
// Returns the current date (time part = 0.0) as TDateTime.
func Date(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("Date() expects 0 arguments, got %d", len(args))
	}

	now := time.Now().UTC()
	// Zero out the time component
	dateOnly := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	dtValue := goTimeToDelphiDateTime(dateOnly)

	return &runtime.FloatValue{Value: dtValue}
}

// Time implements the Time() built-in function.
// Returns the current time (date part = 0.0) as TDateTime.
func Time(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("Time() expects 0 arguments, got %d", len(args))
	}

	now := time.Now().UTC()
	// Use epoch date, only keep time
	timeOnly := time.Date(1899, 12, 30, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), time.UTC)
	dtValue := goTimeToDelphiDateTime(timeOnly)

	return &runtime.FloatValue{Value: dtValue}
}

// UTCDateTime implements the UTCDateTime() built-in function.
// Returns the current UTC date and time as TDateTime.
func UTCDateTime(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("UTCDateTime() expects 0 arguments, got %d", len(args))
	}

	now := time.Now().UTC()
	dtValue := goTimeToDelphiDateTime(now)

	return &runtime.FloatValue{Value: dtValue}
}

// =============================================================================
// Component Extraction Functions
// =============================================================================

// YearOf implements the YearOf() built-in function.
// Returns the year component of a TDateTime.
func YearOf(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("YearOf() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("YearOf() expects Float/TDateTime, got %s", args[0].Type())
	}

	year, _, _ := extractDateComponents(floatVal.Value)
	return &runtime.IntegerValue{Value: int64(year)}
}

// MonthOf implements the MonthOf() built-in function.
// Returns the month component of a TDateTime (1-12).
func MonthOf(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("MonthOf() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("MonthOf() expects Float/TDateTime, got %s", args[0].Type())
	}

	_, month, _ := extractDateComponents(floatVal.Value)
	return &runtime.IntegerValue{Value: int64(month)}
}

// DayOf implements the DayOf() built-in function.
// Returns the day component of a TDateTime (1-31).
func DayOf(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DayOf() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("DayOf() expects Float/TDateTime, got %s", args[0].Type())
	}

	_, _, day := extractDateComponents(floatVal.Value)
	return &runtime.IntegerValue{Value: int64(day)}
}

// HourOf implements the HourOf() built-in function.
// Returns the hour component of a TDateTime (0-23).
func HourOf(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("HourOf() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("HourOf() expects Float/TDateTime, got %s", args[0].Type())
	}

	hour, _, _, _ := extractTimeComponents(floatVal.Value)
	return &runtime.IntegerValue{Value: int64(hour)}
}

// MinuteOf implements the MinuteOf() built-in function.
// Returns the minute component of a TDateTime (0-59).
func MinuteOf(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("MinuteOf() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("MinuteOf() expects Float/TDateTime, got %s", args[0].Type())
	}

	_, minute, _, _ := extractTimeComponents(floatVal.Value)
	return &runtime.IntegerValue{Value: int64(minute)}
}

// SecondOf implements the SecondOf() built-in function.
// Returns the second component of a TDateTime (0-59).
func SecondOf(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("SecondOf() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("SecondOf() expects Float/TDateTime, got %s", args[0].Type())
	}

	_, _, second, _ := extractTimeComponents(floatVal.Value)
	return &runtime.IntegerValue{Value: int64(second)}
}

// DayOfWeek implements the DayOfWeek() built-in function.
// Returns the day of week (1=Sunday, 7=Saturday) like Delphi.
func DayOfWeek(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DayOfWeek() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("DayOfWeek() expects Float/TDateTime, got %s", args[0].Type())
	}

	dow := dayOfWeek(floatVal.Value)
	return &runtime.IntegerValue{Value: int64(dow)}
}

// DayOfTheWeek implements the DayOfTheWeek() built-in function.
// Returns the ISO day of week (1=Monday, 7=Sunday).
func DayOfTheWeek(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DayOfTheWeek() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("DayOfTheWeek() expects Float/TDateTime, got %s", args[0].Type())
	}

	dow := dayOfTheWeek(floatVal.Value)
	return &runtime.IntegerValue{Value: int64(dow)}
}

// DayOfYear implements the DayOfYear() built-in function.
// Returns the day number within the year (1-366).
func DayOfYear(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DayOfYear() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("DayOfYear() expects Float/TDateTime, got %s", args[0].Type())
	}

	doy := dayOfYear(floatVal.Value)
	return &runtime.IntegerValue{Value: int64(doy)}
}

// WeekNumber implements the WeekNumber() built-in function.
// Returns the ISO 8601 week number (1-53).
func WeekNumber(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("WeekNumber() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("WeekNumber() expects Float/TDateTime, got %s", args[0].Type())
	}

	wn := weekNumber(floatVal.Value)
	return &runtime.IntegerValue{Value: int64(wn)}
}

// YearOfWeek implements the YearOfWeek() built-in function.
// Returns the year of the ISO 8601 week.
func YearOfWeek(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("YearOfWeek() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("YearOfWeek() expects Float/TDateTime, got %s", args[0].Type())
	}

	yow := yearOfWeek(floatVal.Value)
	return &runtime.IntegerValue{Value: int64(yow)}
}

// =============================================================================
// Special Date Functions
// =============================================================================

// IsLeapYear implements the IsLeapYear() built-in function.
// Determines if a year is a leap year.
func IsLeapYear(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("IsLeapYear() expects 1 argument, got %d", len(args))
	}

	yearVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("IsLeapYear() expects Integer, got %s", args[0].Type())
	}

	result := isLeapYear(int(yearVal.Value))
	return &runtime.BooleanValue{Value: result}
}

// FirstDayOfYear implements the FirstDayOfYear() built-in function.
// Returns the first day of the year for a given TDateTime.
func FirstDayOfYear(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("FirstDayOfYear() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("FirstDayOfYear() expects Float/TDateTime, got %s", args[0].Type())
	}

	result := firstDayOfYear(dtVal.Value)
	return &runtime.FloatValue{Value: result}
}

// FirstDayOfNextYear implements the FirstDayOfNextYear() built-in function.
// Returns the first day of the next year.
func FirstDayOfNextYear(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("FirstDayOfNextYear() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("FirstDayOfNextYear() expects Float/TDateTime, got %s", args[0].Type())
	}

	result := firstDayOfNextYear(dtVal.Value)
	return &runtime.FloatValue{Value: result}
}

// FirstDayOfMonth implements the FirstDayOfMonth() built-in function.
// Returns the first day of the month for a given TDateTime.
func FirstDayOfMonth(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("FirstDayOfMonth() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("FirstDayOfMonth() expects Float/TDateTime, got %s", args[0].Type())
	}

	result := firstDayOfMonth(dtVal.Value)
	return &runtime.FloatValue{Value: result}
}

// FirstDayOfNextMonth implements the FirstDayOfNextMonth() built-in function.
// Returns the first day of the next month.
func FirstDayOfNextMonth(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("FirstDayOfNextMonth() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("FirstDayOfNextMonth() expects Float/TDateTime, got %s", args[0].Type())
	}

	result := firstDayOfNextMonth(dtVal.Value)
	return &runtime.FloatValue{Value: result}
}

// FirstDayOfWeek implements the FirstDayOfWeek() built-in function.
// Returns the first day (Monday) of the week containing the given TDateTime.
func FirstDayOfWeek(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("FirstDayOfWeek() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("FirstDayOfWeek() expects Float/TDateTime, got %s", args[0].Type())
	}

	result := firstDayOfWeek(dtVal.Value)
	return &runtime.FloatValue{Value: result}
}

// =============================================================================
// Helper Functions (copied from datetime_utils.go)
// =============================================================================

// DelphiEpoch is the reference date for TDateTime calculations.
// Delphi's TDateTime uses December 30, 1899 as day 0.
// This matches Microsoft's OLE Automation date format.
var delphiEpoch = time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)

// Constants for time calculations
const (
	secondsPerDay      = 86400.0
	millisecondsPerDay = 86400000.0
	nanosecondsPerDay  = 86400000000000.0
)

// goTimeToDelphiDateTime converts a Go time.Time to Delphi TDateTime format.
// TDateTime is a float64 where:
//   - Integer part = number of days since December 30, 1899
//   - Fractional part = time of day (0.5 = noon, 0.25 = 6am)
//
// Core conversion function for DateTime support
func goTimeToDelphiDateTime(t time.Time) float64 {
	// Calculate days since Delphi epoch
	duration := t.Sub(delphiEpoch)
	days := duration.Seconds() / secondsPerDay
	return days
}

// delphiDateTimeToGoTime converts a Delphi TDateTime float64 to Go time.Time.
// The result is in UTC timezone.
//
// Core conversion function for DateTime support
func delphiDateTimeToGoTime(dt float64) time.Time {
	// Calculate duration from epoch
	seconds := dt * secondsPerDay
	duration := time.Duration(seconds * float64(time.Second))
	return delphiEpoch.Add(duration)
}

// isValidDate checks if the given year, month, day constitutes a valid date.
// Date validation for EncodeDate
func isValidDate(year, month, day int) bool {
	if year < 1 || year > 9999 {
		return false
	}
	if month < 1 || month > 12 {
		return false
	}
	if day < 1 {
		return false
	}

	// Check days in month
	daysInMonth := []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

	// Handle leap years
	if isLeapYear(year) {
		daysInMonth[1] = 29
	}

	return day <= daysInMonth[month-1]
}

// isValidTime checks if the given hour, minute, second, millisecond constitutes valid time.
func isValidTime(hour, minute, second, millisecond int) bool {
	return hour >= 0 && hour < 24 &&
		minute >= 0 && minute < 60 &&
		second >= 0 && second < 60 &&
		millisecond >= 0 && millisecond < 1000
}

// isLeapYear determines if a year is a leap year.
// A year is a leap year if it's divisible by 4, except for years divisible by 100,
// unless they're also divisible by 400.
func isLeapYear(year int) bool {
	return (year%4 == 0 && year%100 != 0) || year%400 == 0
}

// daysInMonth returns the number of days in a given month for a given year.
func daysInMonth(year, month int) int {
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	case 2:
		if isLeapYear(year) {
			return 29
		}
		return 28
	default:
		return 0
	}
}

// extractDateComponents extracts year, month, day from a TDateTime value.
func extractDateComponents(dt float64) (year, month, day int) {
	t := delphiDateTimeToGoTime(dt)
	return t.Year(), int(t.Month()), t.Day()
}

// extractTimeComponents extracts hour, minute, second, millisecond from a TDateTime value.
func extractTimeComponents(dt float64) (hour, minute, second, millisecond int) {
	t := delphiDateTimeToGoTime(dt)
	return t.Hour(), t.Minute(), t.Second(), t.Nanosecond() / 1000000
}

// formatDateTime formats a TDateTime value according to a format string.
// This implements DWScript's FormatDateTime function with Delphi-style format specifiers.
//
// Supported format specifiers:
//
//	yyyy - 4-digit year (e.g., 2023)
//	yy   - 2-digit year (e.g., 23)
//	mm   - 2-digit month (01-12)
//	m    - month without leading zero (1-12)
//	dd   - 2-digit day (01-31)
//	d    - day without leading zero (1-31)
//	hh   - 2-digit hour, 24h format (00-23)
//	h    - hour without leading zero, 24h format (0-23)
//	nn   - 2-digit minute (00-59)
//	n    - minute without leading zero (0-59)
//	ss   - 2-digit second (00-59)
//	s    - second without leading zero (0-59)
//	zzz  - 3-digit millisecond (000-999)
//	z    - millisecond without leading zeros (0-999)
//
// Note: Format specifiers are case-sensitive.
func formatDateTime(format string, dt float64) string {
	t := delphiDateTimeToGoTime(dt)

	year := t.Year()
	month := int(t.Month())
	day := t.Day()
	hour := t.Hour()
	minute := t.Minute()
	second := t.Second()
	millisecond := t.Nanosecond() / 1000000

	result := format

	// Process format specifiers in order of longest to shortest to avoid conflicts
	// e.g., 'yyyy' should be processed before 'yy'

	// Year formats
	result = strings.ReplaceAll(result, "yyyy", fmt.Sprintf("%04d", year))
	result = strings.ReplaceAll(result, "yy", fmt.Sprintf("%02d", year%100))

	// Month formats
	result = strings.ReplaceAll(result, "mm", fmt.Sprintf("%02d", month))
	result = strings.ReplaceAll(result, "m", fmt.Sprintf("%d", month))

	// Day formats
	result = strings.ReplaceAll(result, "dd", fmt.Sprintf("%02d", day))
	result = strings.ReplaceAll(result, "d", fmt.Sprintf("%d", day))

	// Hour formats (24h)
	result = strings.ReplaceAll(result, "hh", fmt.Sprintf("%02d", hour))
	result = strings.ReplaceAll(result, "h", fmt.Sprintf("%d", hour))

	// Minute formats (nn for minute to avoid conflict with month m)
	result = strings.ReplaceAll(result, "nn", fmt.Sprintf("%02d", minute))
	result = strings.ReplaceAll(result, "n", fmt.Sprintf("%d", minute))

	// Second formats
	result = strings.ReplaceAll(result, "ss", fmt.Sprintf("%02d", second))
	result = strings.ReplaceAll(result, "s", fmt.Sprintf("%d", second))

	// Millisecond formats
	result = strings.ReplaceAll(result, "zzz", fmt.Sprintf("%03d", millisecond))
	result = strings.ReplaceAll(result, "z", fmt.Sprintf("%d", millisecond))

	return result
}

// parseDateTime attempts to parse a date/time string in various common formats.
// This is a simplified implementation that handles common ISO-8601 and locale-neutral formats.
func parseDateTime(s string) (float64, error) {
	s = strings.TrimSpace(s)

	// Try various common formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		"02/01/2006 15:04:05",
		"02/01/2006",
		"01/02/2006 15:04:05",
		"01/02/2006",
		"15:04:05",
		"15:04",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return goTimeToDelphiDateTime(t), nil
		}
	}

	return 0, fmt.Errorf("unable to parse date/time: %s", s)
}

// parseDate attempts to parse a date string.
func parseDate(s string) (float64, error) {
	s = strings.TrimSpace(s)

	formats := []string{
		"2006-01-02",
		"02/01/2006",
		"01/02/2006",
		"2006/01/02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			// Set time component to 0
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
			return goTimeToDelphiDateTime(t), nil
		}
	}

	return 0, fmt.Errorf("unable to parse date: %s", s)
}

// parseTime attempts to parse a time string.
func parseTime(s string) (float64, error) {
	s = strings.TrimSpace(s)

	formats := []string{
		"15:04:05",
		"15:04",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			// Use epoch date, only keep time
			t = time.Date(1899, 12, 30, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
			return goTimeToDelphiDateTime(t), nil
		}
	}

	return 0, fmt.Errorf("unable to parse time: %s", s)
}

// formatISO8601 formats a TDateTime as ISO 8601 string (YYYY-MM-DDTHH:MM:SS).
func formatISO8601(dt float64) string {
	t := delphiDateTimeToGoTime(dt)
	return t.Format("2006-01-02T15:04:05")
}

// formatDateISO8601 formats just the date part as ISO 8601 (YYYY-MM-DD).
func formatDateISO8601(dt float64) string {
	t := delphiDateTimeToGoTime(dt)
	return t.Format("2006-01-02")
}

// parseISO8601 parses an ISO 8601 date/time string.
// Supports formats: YYYY-MM-DD, YYYY-MM-DDTHH:MM:SS, YYYY-MM-DDTHH:MM:SSZ
func parseISO8601(s string) (float64, error) {
	s = strings.TrimSpace(s)

	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return goTimeToDelphiDateTime(t), nil
		}
	}

	return 0, fmt.Errorf("unable to parse ISO 8601 date/time: %s", s)
}

// formatRFC822 formats a TDateTime as RFC 822 string.
func formatRFC822(dt float64) string {
	t := delphiDateTimeToGoTime(dt)
	return t.Format(time.RFC822)
}

// parseRFC822 parses an RFC 822 date/time string.
func parseRFC822(s string) (float64, error) {
	s = strings.TrimSpace(s)

	if t, err := time.Parse(time.RFC822, s); err == nil {
		return goTimeToDelphiDateTime(t), nil
	}
	if t, err := time.Parse(time.RFC822Z, s); err == nil {
		return goTimeToDelphiDateTime(t), nil
	}

	return 0, fmt.Errorf("unable to parse RFC 822 date/time: %s", s)
}

// unixTimeToDateTime converts Unix timestamp (seconds since 1970-01-01) to TDateTime.
func unixTimeToDateTime(unixTime int64) float64 {
	t := time.Unix(unixTime, 0).UTC()
	return goTimeToDelphiDateTime(t)
}

// unixTimeMSecToDateTime converts Unix timestamp in milliseconds to TDateTime.
func unixTimeMSecToDateTime(unixTimeMS int64) float64 {
	seconds := unixTimeMS / 1000
	nanoseconds := (unixTimeMS % 1000) * 1000000
	t := time.Unix(seconds, nanoseconds).UTC()
	return goTimeToDelphiDateTime(t)
}

// dateTimeToUnixTime converts TDateTime to Unix timestamp (seconds since 1970-01-01).
func dateTimeToUnixTime(dt float64) int64 {
	t := delphiDateTimeToGoTime(dt)
	return t.Unix()
}

// dateTimeToUnixTimeMSec converts TDateTime to Unix timestamp in milliseconds.
func dateTimeToUnixTimeMSec(dt float64) int64 {
	t := delphiDateTimeToGoTime(dt)
	return t.UnixMilli()
}

// incMonths adds a number of months to a TDateTime value.
// This correctly handles month boundaries and leap years.
func incMonths(dt float64, months int) float64 {
	t := delphiDateTimeToGoTime(dt)

	year := t.Year()
	month := int(t.Month())
	day := t.Day()

	// Add months
	month += months

	// Handle year overflow/underflow
	for month > 12 {
		month -= 12
		year++
	}
	for month < 1 {
		month += 12
		year--
	}

	// Adjust day if it exceeds days in target month
	maxDay := daysInMonth(year, month)
	if day > maxDay {
		day = maxDay
	}

	result := time.Date(year, time.Month(month), day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
	return goTimeToDelphiDateTime(result)
}

// incYears adds a number of years to a TDateTime value.
func incYears(dt float64, years int) float64 {
	return incMonths(dt, years*12)
}

// daysBetween calculates the number of whole days between two TDateTime values.
func daysBetween(dt1, dt2 float64) int {
	diff := math.Abs(dt1 - dt2)
	return int(math.Floor(diff))
}

// hoursBetween calculates the number of whole hours between two TDateTime values.
func hoursBetween(dt1, dt2 float64) int {
	diff := math.Abs(dt1 - dt2)
	return int(math.Floor(diff * 24.0))
}

// minutesBetween calculates the number of whole minutes between two TDateTime values.
func minutesBetween(dt1, dt2 float64) int {
	diff := math.Abs(dt1 - dt2)
	return int(math.Floor(diff * 24.0 * 60.0))
}

// secondsBetween calculates the number of whole seconds between two TDateTime values.
func secondsBetween(dt1, dt2 float64) int {
	diff := math.Abs(dt1 - dt2)
	return int(math.Floor(diff * secondsPerDay))
}

// firstDayOfYear returns the first day of the year for a given TDateTime.
func firstDayOfYear(dt float64) float64 {
	t := delphiDateTimeToGoTime(dt)
	result := time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	return goTimeToDelphiDateTime(result)
}

// firstDayOfNextYear returns the first day of the next year.
func firstDayOfNextYear(dt float64) float64 {
	t := delphiDateTimeToGoTime(dt)
	result := time.Date(t.Year()+1, time.January, 1, 0, 0, 0, 0, time.UTC)
	return goTimeToDelphiDateTime(result)
}

// firstDayOfMonth returns the first day of the month for a given TDateTime.
func firstDayOfMonth(dt float64) float64 {
	t := delphiDateTimeToGoTime(dt)
	result := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	return goTimeToDelphiDateTime(result)
}

// firstDayOfNextMonth returns the first day of the next month.
func firstDayOfNextMonth(dt float64) float64 {
	t := delphiDateTimeToGoTime(dt)

	year := t.Year()
	month := int(t.Month()) + 1

	if month > 12 {
		month = 1
		year++
	}

	result := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	return goTimeToDelphiDateTime(result)
}

// firstDayOfWeek returns the first day (Monday) of the week containing the given TDateTime.
func firstDayOfWeek(dt float64) float64 {
	t := delphiDateTimeToGoTime(dt)

	// Go's time.Weekday: Sunday=0, Monday=1, ..., Saturday=6
	// We want Monday as first day of week
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday becomes 7
	}

	daysToSubtract := weekday - 1 // Days to go back to Monday
	result := t.AddDate(0, 0, -daysToSubtract)
	result = time.Date(result.Year(), result.Month(), result.Day(), 0, 0, 0, 0, time.UTC)

	return goTimeToDelphiDateTime(result)
}

// dayOfYear returns the day number within the year (1-366).
func dayOfYear(dt float64) int {
	t := delphiDateTimeToGoTime(dt)
	return t.YearDay()
}

// weekNumber calculates the ISO 8601 week number for a given TDateTime.
func weekNumber(dt float64) int {
	t := delphiDateTimeToGoTime(dt)
	_, week := t.ISOWeek()
	return week
}

// yearOfWeek returns the year associated with the ISO 8601 week containing the date.
func yearOfWeek(dt float64) int {
	t := delphiDateTimeToGoTime(dt)
	year, _ := t.ISOWeek()
	return year
}

// dayOfWeek returns the day of week (1=Sunday, 7=Saturday) like Delphi's DayOfWeek.
func dayOfWeek(dt float64) int {
	t := delphiDateTimeToGoTime(dt)
	// Go: Sunday=0, Monday=1, ..., Saturday=6
	// Delphi: Sunday=1, Monday=2, ..., Saturday=7
	return int(t.Weekday()) + 1
}

// dayOfTheWeek returns the day of week (1=Monday, 7=Sunday) for ISO 8601 compatibility.
func dayOfTheWeek(dt float64) int {
	t := delphiDateTimeToGoTime(dt)
	weekday := int(t.Weekday())
	if weekday == 0 {
		return 7 // Sunday = 7
	}
	return weekday // Monday=1, ..., Saturday=6
}

// parseCustomDateTime parses a datetime string according to a custom format specifier.
// This is a simplified implementation of DWScript's ParseDateTime function.
func parseCustomDateTime(format, s string) (float64, error) {
	// This is a simplified regex-based parser
	// Convert format specifiers to regex patterns
	pattern := format

	// Escape regex special characters except our format specifiers
	pattern = regexp.QuoteMeta(pattern)

	// Replace format specifiers with capturing groups
	replacements := map[string]string{
		`yyyy`: `(\d{4})`,
		`yy`:   `(\d{2})`,
		`mm`:   `(\d{1,2})`,
		`dd`:   `(\d{1,2})`,
		`hh`:   `(\d{1,2})`,
		`nn`:   `(\d{1,2})`,
		`ss`:   `(\d{1,2})`,
		`zzz`:  `(\d{1,3})`,
	}

	// Apply replacements (longest first to avoid conflicts)
	for spec, regex := range replacements {
		pattern = strings.ReplaceAll(pattern, regexp.QuoteMeta(spec), regex)
	}

	// Try to match
	re := regexp.MustCompile("^" + pattern + "$")
	matches := re.FindStringSubmatch(s)

	if matches == nil {
		return 0, fmt.Errorf("string '%s' does not match format '%s'", s, format)
	}

	// Extract matched values
	// This is a simplified implementation - a complete one would track which
	// groups correspond to which format specifiers

	// For now, return an error indicating this needs the simpler parse functions
	return parseDateTime(s)
}
