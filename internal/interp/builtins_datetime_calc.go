package interp

import (
	"time"

	"github.com/cwbudde/go-dws/internal/ast"
)

// ============================================================================
// Date Encoding Functions
// ============================================================================

// builtinEncodeDate implements the EncodeDate() built-in function.
// Creates a TDateTime from year, month, day components.
// EncodeDate(year, month, day: Integer): TDateTime
func (i *Interpreter) builtinEncodeDate(args []Value) Value {
	if len(args) != 3 {
		return i.newErrorWithLocation(i.currentNode, "EncodeDate() expects 3 arguments (year, month, day), got %d", len(args))
	}

	// Extract year
	yearVal, ok := args[0].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "EncodeDate() year must be Integer, got %s", args[0].Type())
	}

	// Extract month
	monthVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "EncodeDate() month must be Integer, got %s", args[1].Type())
	}

	// Extract day
	dayVal, ok := args[2].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "EncodeDate() day must be Integer, got %s", args[2].Type())
	}

	year := int(yearVal.Value)
	month := int(monthVal.Value)
	day := int(dayVal.Value)

	// Validate date
	if !isValidDate(year, month, day) {
		return i.newErrorWithLocation(i.currentNode, "EncodeDate() invalid date: %d-%02d-%02d", year, month, day)
	}

	// Create date (time = 00:00:00)
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	dtValue := goTimeToDelphiDateTime(t)

	return &FloatValue{Value: dtValue}
}

// builtinEncodeTime implements the EncodeTime() built-in function.
// Creates a TDateTime from hour, minute, second, millisecond components.
// EncodeTime(hour, minute, second, msec: Integer): TDateTime
func (i *Interpreter) builtinEncodeTime(args []Value) Value {
	if len(args) != 4 {
		return i.newErrorWithLocation(i.currentNode, "EncodeTime() expects 4 arguments (hour, minute, second, msec), got %d", len(args))
	}

	// Extract hour
	hourVal, ok := args[0].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "EncodeTime() hour must be Integer, got %s", args[0].Type())
	}

	// Extract minute
	minuteVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "EncodeTime() minute must be Integer, got %s", args[1].Type())
	}

	// Extract second
	secondVal, ok := args[2].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "EncodeTime() second must be Integer, got %s", args[2].Type())
	}

	// Extract millisecond
	msecVal, ok := args[3].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "EncodeTime() msec must be Integer, got %s", args[3].Type())
	}

	hour := int(hourVal.Value)
	minute := int(minuteVal.Value)
	second := int(secondVal.Value)
	msec := int(msecVal.Value)

	// Validate time
	if !isValidTime(hour, minute, second, msec) {
		return i.newErrorWithLocation(i.currentNode, "EncodeTime() invalid time: %02d:%02d:%02d.%03d", hour, minute, second, msec)
	}

	// Create time on epoch date
	nanoseconds := msec * 1000000
	t := time.Date(1899, 12, 30, hour, minute, second, nanoseconds, time.UTC)
	dtValue := goTimeToDelphiDateTime(t)

	return &FloatValue{Value: dtValue}
}

// builtinEncodeDateTime implements the EncodeDateTime() built-in function.
// Creates a TDateTime from full date and time components.
// EncodeDateTime(year, month, day, hour, minute, second, msec: Integer): TDateTime
func (i *Interpreter) builtinEncodeDateTime(args []Value) Value {
	if len(args) != 7 {
		return i.newErrorWithLocation(i.currentNode, "EncodeDateTime() expects 7 arguments (year, month, day, hour, minute, second, msec), got %d", len(args))
	}

	// Extract all components
	yearVal, ok := args[0].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "EncodeDateTime() year must be Integer, got %s", args[0].Type())
	}

	monthVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "EncodeDateTime() month must be Integer, got %s", args[1].Type())
	}

	dayVal, ok := args[2].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "EncodeDateTime() day must be Integer, got %s", args[2].Type())
	}

	hourVal, ok := args[3].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "EncodeDateTime() hour must be Integer, got %s", args[3].Type())
	}

	minuteVal, ok := args[4].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "EncodeDateTime() minute must be Integer, got %s", args[4].Type())
	}

	secondVal, ok := args[5].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "EncodeDateTime() second must be Integer, got %s", args[5].Type())
	}

	msecVal, ok := args[6].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "EncodeDateTime() msec must be Integer, got %s", args[6].Type())
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
		return i.newErrorWithLocation(i.currentNode, "EncodeDateTime() invalid date: %d-%02d-%02d", year, month, day)
	}

	if !isValidTime(hour, minute, second, msec) {
		return i.newErrorWithLocation(i.currentNode, "EncodeDateTime() invalid time: %02d:%02d:%02d.%03d", hour, minute, second, msec)
	}

	// Create full datetime
	nanoseconds := msec * 1000000
	t := time.Date(year, time.Month(month), day, hour, minute, second, nanoseconds, time.UTC)
	dtValue := goTimeToDelphiDateTime(t)

	return &FloatValue{Value: dtValue}
}

// ============================================================================
// Date Decoding Functions (Var Parameters)
// ============================================================================

// builtinDecodeDate implements the DecodeDate() built-in function.
// Extracts year, month, day components from a TDateTime.
// DecodeDate(dt: TDateTime; var year, month, day: Integer)
func (i *Interpreter) builtinDecodeDate(args []ast.Expression) Value {
	if len(args) != 4 {
		return i.newErrorWithLocation(i.currentNode, "DecodeDate() expects 4 arguments (dt, var year, var month, var day), got %d", len(args))
	}

	// Evaluate the first argument (the TDateTime value)
	dtVal := i.Eval(args[0])
	if isError(dtVal) {
		return dtVal
	}

	floatVal, ok := dtVal.(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "DecodeDate() expects Float/TDateTime as first argument, got %s", dtVal.Type())
	}

	// Extract date components
	year, month, day := extractDateComponents(floatVal.Value)

	// Set the var parameters (args 1, 2, 3)
	for idx, val := range []int{year, month, day} {
		varIdent, ok := args[idx+1].(*ast.Identifier)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "DecodeDate() argument %d must be a variable, got %T", idx+2, args[idx+1])
		}

		if err := i.env.Set(varIdent.Value, &IntegerValue{Value: int64(val)}); err != nil {
			return i.newErrorWithLocation(i.currentNode, "DecodeDate() failed to set variable %s: %s", varIdent.Value, err)
		}
	}

	return &NilValue{}
}

// builtinDecodeTime implements the DecodeTime() built-in function.
// Extracts hour, minute, second, millisecond components from a TDateTime.
// DecodeTime(dt: TDateTime; var hour, minute, second, msec: Integer)
func (i *Interpreter) builtinDecodeTime(args []ast.Expression) Value {
	if len(args) != 5 {
		return i.newErrorWithLocation(i.currentNode, "DecodeTime() expects 5 arguments (dt, var hour, var minute, var second, var msec), got %d", len(args))
	}

	// Evaluate the first argument (the TDateTime value)
	dtVal := i.Eval(args[0])
	if isError(dtVal) {
		return dtVal
	}

	floatVal, ok := dtVal.(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "DecodeTime() expects Float/TDateTime as first argument, got %s", dtVal.Type())
	}

	// Extract time components
	hour, minute, second, msec := extractTimeComponents(floatVal.Value)

	// Set the var parameters (args 1, 2, 3, 4)
	for idx, val := range []int{hour, minute, second, msec} {
		varIdent, ok := args[idx+1].(*ast.Identifier)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "DecodeTime() argument %d must be a variable, got %T", idx+2, args[idx+1])
		}

		if err := i.env.Set(varIdent.Value, &IntegerValue{Value: int64(val)}); err != nil {
			return i.newErrorWithLocation(i.currentNode, "DecodeTime() failed to set variable %s: %s", varIdent.Value, err)
		}
	}

	return &NilValue{}
}

// ============================================================================
// Incrementing Functions
// ============================================================================

// builtinIncYear implements the IncYear() built-in function.
// Adds years to a TDateTime.
func (i *Interpreter) builtinIncYear(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "IncYear() expects 2 arguments (dt, years), got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IncYear() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	yearsVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IncYear() expects Integer as second argument, got %s", args[1].Type())
	}

	result := incYears(dtVal.Value, int(yearsVal.Value))
	return &FloatValue{Value: result}
}

// builtinIncMonth implements the IncMonth() built-in function.
// Adds months to a TDateTime.
func (i *Interpreter) builtinIncMonth(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "IncMonth() expects 2 arguments (dt, months), got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IncMonth() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	monthsVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IncMonth() expects Integer as second argument, got %s", args[1].Type())
	}

	result := incMonths(dtVal.Value, int(monthsVal.Value))
	return &FloatValue{Value: result}
}

// builtinIncDay implements the IncDay() built-in function.
// Adds days to a TDateTime.
func (i *Interpreter) builtinIncDay(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "IncDay() expects 2 arguments (dt, days), got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IncDay() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	daysVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IncDay() expects Integer as second argument, got %s", args[1].Type())
	}

	// Simple addition since TDateTime stores days as integer part
	result := dtVal.Value + float64(daysVal.Value)
	return &FloatValue{Value: result}
}

// builtinIncHour implements the IncHour() built-in function.
// Adds hours to a TDateTime.
func (i *Interpreter) builtinIncHour(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "IncHour() expects 2 arguments (dt, hours), got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IncHour() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	hoursVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IncHour() expects Integer as second argument, got %s", args[1].Type())
	}

	// 1 hour = 1/24 day
	result := dtVal.Value + (float64(hoursVal.Value) / 24.0)
	return &FloatValue{Value: result}
}

// builtinIncMinute implements the IncMinute() built-in function.
// Adds minutes to a TDateTime.
func (i *Interpreter) builtinIncMinute(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "IncMinute() expects 2 arguments (dt, minutes), got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IncMinute() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	minutesVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IncMinute() expects Integer as second argument, got %s", args[1].Type())
	}

	// 1 minute = 1/(24*60) day
	result := dtVal.Value + (float64(minutesVal.Value) / (24.0 * 60.0))
	return &FloatValue{Value: result}
}

// builtinIncSecond implements the IncSecond() built-in function.
// Adds seconds to a TDateTime.
func (i *Interpreter) builtinIncSecond(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "IncSecond() expects 2 arguments (dt, seconds), got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IncSecond() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	secondsVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IncSecond() expects Integer as second argument, got %s", args[1].Type())
	}

	// 1 second = 1/86400 day
	result := dtVal.Value + (float64(secondsVal.Value) / 86400.0)
	return &FloatValue{Value: result}
}

// ============================================================================
// Date Difference Functions
// ============================================================================

// builtinDaysBetween implements the DaysBetween() built-in function.
// Calculates whole days between two TDateTime values.
func (i *Interpreter) builtinDaysBetween(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "DaysBetween() expects 2 arguments (dt1, dt2), got %d", len(args))
	}

	dt1Val, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "DaysBetween() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	dt2Val, ok := args[1].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "DaysBetween() expects Float/TDateTime as second argument, got %s", args[1].Type())
	}

	days := daysBetween(dt1Val.Value, dt2Val.Value)
	return &IntegerValue{Value: int64(days)}
}

// builtinHoursBetween implements the HoursBetween() built-in function.
// Calculates whole hours between two TDateTime values.
func (i *Interpreter) builtinHoursBetween(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "HoursBetween() expects 2 arguments (dt1, dt2), got %d", len(args))
	}

	dt1Val, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "HoursBetween() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	dt2Val, ok := args[1].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "HoursBetween() expects Float/TDateTime as second argument, got %s", args[1].Type())
	}

	hours := hoursBetween(dt1Val.Value, dt2Val.Value)
	return &IntegerValue{Value: int64(hours)}
}

// builtinMinutesBetween implements the MinutesBetween() built-in function.
// Calculates whole minutes between two TDateTime values.
func (i *Interpreter) builtinMinutesBetween(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "MinutesBetween() expects 2 arguments (dt1, dt2), got %d", len(args))
	}

	dt1Val, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "MinutesBetween() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	dt2Val, ok := args[1].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "MinutesBetween() expects Float/TDateTime as second argument, got %s", args[1].Type())
	}

	minutes := minutesBetween(dt1Val.Value, dt2Val.Value)
	return &IntegerValue{Value: int64(minutes)}
}

// builtinSecondsBetween implements the SecondsBetween() built-in function.
// Calculates whole seconds between two TDateTime values.
func (i *Interpreter) builtinSecondsBetween(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "SecondsBetween() expects 2 arguments (dt1, dt2), got %d", len(args))
	}

	dt1Val, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "SecondsBetween() expects Float/TDateTime as first argument, got %s", args[0].Type())
	}

	dt2Val, ok := args[1].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "SecondsBetween() expects Float/TDateTime as second argument, got %s", args[1].Type())
	}

	seconds := secondsBetween(dt1Val.Value, dt2Val.Value)
	return &IntegerValue{Value: int64(seconds)}
}
