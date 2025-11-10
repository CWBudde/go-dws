package interp

import (
	"time"
)

// ============================================================================
// Current Date/Time Functions
// ============================================================================

// builtinNow implements the Now() built-in function.
// Returns the current date and time as TDateTime.
func (i *Interpreter) builtinNow(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "Now() expects 0 arguments, got %d", len(args))
	}

	now := time.Now().UTC()
	dtValue := goTimeToDelphiDateTime(now)

	return &FloatValue{Value: dtValue}
}

// builtinDate implements the Date() built-in function.
// Returns the current date (time part = 0.0) as TDateTime.
func (i *Interpreter) builtinDate(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "Date() expects 0 arguments, got %d", len(args))
	}

	now := time.Now().UTC()
	// Zero out the time component
	dateOnly := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	dtValue := goTimeToDelphiDateTime(dateOnly)

	return &FloatValue{Value: dtValue}
}

// builtinTime implements the Time() built-in function.
// Returns the current time (date part = 0.0) as TDateTime.
func (i *Interpreter) builtinTime(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "Time() expects 0 arguments, got %d", len(args))
	}

	now := time.Now().UTC()
	// Use epoch date, only keep time
	timeOnly := time.Date(1899, 12, 30, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), time.UTC)
	dtValue := goTimeToDelphiDateTime(timeOnly)

	return &FloatValue{Value: dtValue}
}

// builtinUTCDateTime implements the UTCDateTime() built-in function.
// Returns the current UTC date and time as TDateTime.
func (i *Interpreter) builtinUTCDateTime(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "UTCDateTime() expects 0 arguments, got %d", len(args))
	}

	now := time.Now().UTC()
	dtValue := goTimeToDelphiDateTime(now)

	return &FloatValue{Value: dtValue}
}

// ============================================================================
// Component Extraction Functions
// ============================================================================

// builtinYearOf implements the YearOf() built-in function.
// Returns the year component of a TDateTime.
func (i *Interpreter) builtinYearOf(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "YearOf() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "YearOf() expects Float/TDateTime, got %s", args[0].Type())
	}

	year, _, _ := extractDateComponents(floatVal.Value)
	return &IntegerValue{Value: int64(year)}
}

// builtinMonthOf implements the MonthOf() built-in function.
// Returns the month component of a TDateTime (1-12).
func (i *Interpreter) builtinMonthOf(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "MonthOf() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "MonthOf() expects Float/TDateTime, got %s", args[0].Type())
	}

	_, month, _ := extractDateComponents(floatVal.Value)
	return &IntegerValue{Value: int64(month)}
}

// builtinDayOf implements the DayOf() built-in function.
// Returns the day component of a TDateTime (1-31).
func (i *Interpreter) builtinDayOf(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "DayOf() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "DayOf() expects Float/TDateTime, got %s", args[0].Type())
	}

	_, _, day := extractDateComponents(floatVal.Value)
	return &IntegerValue{Value: int64(day)}
}

// builtinHourOf implements the HourOf() built-in function.
// Returns the hour component of a TDateTime (0-23).
func (i *Interpreter) builtinHourOf(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "HourOf() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "HourOf() expects Float/TDateTime, got %s", args[0].Type())
	}

	hour, _, _, _ := extractTimeComponents(floatVal.Value)
	return &IntegerValue{Value: int64(hour)}
}

// builtinMinuteOf implements the MinuteOf() built-in function.
// Returns the minute component of a TDateTime (0-59).
func (i *Interpreter) builtinMinuteOf(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "MinuteOf() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "MinuteOf() expects Float/TDateTime, got %s", args[0].Type())
	}

	_, minute, _, _ := extractTimeComponents(floatVal.Value)
	return &IntegerValue{Value: int64(minute)}
}

// builtinSecondOf implements the SecondOf() built-in function.
// Returns the second component of a TDateTime (0-59).
func (i *Interpreter) builtinSecondOf(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "SecondOf() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "SecondOf() expects Float/TDateTime, got %s", args[0].Type())
	}

	_, _, second, _ := extractTimeComponents(floatVal.Value)
	return &IntegerValue{Value: int64(second)}
}

// builtinDayOfWeek implements the DayOfWeek() built-in function.
// Returns the day of week (1=Sunday, 7=Saturday) like Delphi.
func (i *Interpreter) builtinDayOfWeek(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "DayOfWeek() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "DayOfWeek() expects Float/TDateTime, got %s", args[0].Type())
	}

	dow := dayOfWeek(floatVal.Value)
	return &IntegerValue{Value: int64(dow)}
}

// builtinDayOfTheWeek implements the DayOfTheWeek() built-in function.
// Returns the ISO day of week (1=Monday, 7=Sunday).
func (i *Interpreter) builtinDayOfTheWeek(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "DayOfTheWeek() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "DayOfTheWeek() expects Float/TDateTime, got %s", args[0].Type())
	}

	dow := dayOfTheWeek(floatVal.Value)
	return &IntegerValue{Value: int64(dow)}
}

// builtinDayOfYear implements the DayOfYear() built-in function.
// Returns the day number within the year (1-366).
func (i *Interpreter) builtinDayOfYear(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "DayOfYear() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "DayOfYear() expects Float/TDateTime, got %s", args[0].Type())
	}

	doy := dayOfYear(floatVal.Value)
	return &IntegerValue{Value: int64(doy)}
}

// builtinWeekNumber implements the WeekNumber() built-in function.
// Returns the ISO 8601 week number (1-53).
func (i *Interpreter) builtinWeekNumber(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "WeekNumber() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "WeekNumber() expects Float/TDateTime, got %s", args[0].Type())
	}

	wn := weekNumber(floatVal.Value)
	return &IntegerValue{Value: int64(wn)}
}

// builtinYearOfWeek implements the YearOfWeek() built-in function.
// Returns the year of the ISO 8601 week.
func (i *Interpreter) builtinYearOfWeek(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "YearOfWeek() expects 1 argument, got %d", len(args))
	}

	floatVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "YearOfWeek() expects Float/TDateTime, got %s", args[0].Type())
	}

	yow := yearOfWeek(floatVal.Value)
	return &IntegerValue{Value: int64(yow)}
}

// ============================================================================
// Special Date Functions
// ============================================================================

// builtinIsLeapYear implements the IsLeapYear() built-in function.
// Determines if a year is a leap year.
func (i *Interpreter) builtinIsLeapYear(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "IsLeapYear() expects 1 argument, got %d", len(args))
	}

	yearVal, ok := args[0].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IsLeapYear() expects Integer, got %s", args[0].Type())
	}

	result := isLeapYear(int(yearVal.Value))
	return &BooleanValue{Value: result}
}

// builtinFirstDayOfYear implements the FirstDayOfYear() built-in function.
// Returns the first day of the year for a given TDateTime.
func (i *Interpreter) builtinFirstDayOfYear(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "FirstDayOfYear() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "FirstDayOfYear() expects Float/TDateTime, got %s", args[0].Type())
	}

	result := firstDayOfYear(dtVal.Value)
	return &FloatValue{Value: result}
}

// builtinFirstDayOfNextYear implements the FirstDayOfNextYear() built-in function.
// Returns the first day of the next year.
func (i *Interpreter) builtinFirstDayOfNextYear(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "FirstDayOfNextYear() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "FirstDayOfNextYear() expects Float/TDateTime, got %s", args[0].Type())
	}

	result := firstDayOfNextYear(dtVal.Value)
	return &FloatValue{Value: result}
}

// builtinFirstDayOfMonth implements the FirstDayOfMonth() built-in function.
// Returns the first day of the month for a given TDateTime.
func (i *Interpreter) builtinFirstDayOfMonth(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "FirstDayOfMonth() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "FirstDayOfMonth() expects Float/TDateTime, got %s", args[0].Type())
	}

	result := firstDayOfMonth(dtVal.Value)
	return &FloatValue{Value: result}
}

// builtinFirstDayOfNextMonth implements the FirstDayOfNextMonth() built-in function.
// Returns the first day of the next month.
func (i *Interpreter) builtinFirstDayOfNextMonth(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "FirstDayOfNextMonth() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "FirstDayOfNextMonth() expects Float/TDateTime, got %s", args[0].Type())
	}

	result := firstDayOfNextMonth(dtVal.Value)
	return &FloatValue{Value: result}
}

// builtinFirstDayOfWeek implements the FirstDayOfWeek() built-in function.
// Returns the first day (Monday) of the week containing the given TDateTime.
func (i *Interpreter) builtinFirstDayOfWeek(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "FirstDayOfWeek() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "FirstDayOfWeek() expects Float/TDateTime, got %s", args[0].Type())
	}

	result := firstDayOfWeek(dtVal.Value)
	return &FloatValue{Value: result}
}
