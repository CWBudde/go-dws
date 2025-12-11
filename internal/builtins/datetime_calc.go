package builtins

import (
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

// NOTE: DecodeDate and DecodeTime are implemented in internal/interp/builtins_datetime_calc.go
// as var-param functions (taking []ast.Expression). They cannot be migrated to the builtins
// package because they need direct AST access to modify variables in-place.
//
// DecodeDate(dt: TDateTime; var year, month, day: Integer)
// DecodeTime(dt: TDateTime; var hour, minute, second, msec: Integer)
//
// These functions are registered in callBuiltinWithVarParam() in functions_builtins.go.

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
