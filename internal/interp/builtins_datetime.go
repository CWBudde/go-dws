package interp

import (
	"time"

	"github.com/cwbudde/go-dws/internal/ast"
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
// Formatting Functions
// ============================================================================

// builtinFormatDateTime implements the FormatDateTime() built-in function.
// Formats a TDateTime according to a format string.
// FormatDateTime(format: String, dt: TDateTime): String
func (i *Interpreter) builtinFormatDateTime(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "FormatDateTime() expects 2 arguments (format, dt), got %d", len(args))
	}

	// First argument: format string
	formatVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "FormatDateTime() expects String as first argument, got %s", args[0].Type())
	}

	// Second argument: TDateTime value
	dtVal, ok := args[1].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "FormatDateTime() expects Float/TDateTime as second argument, got %s", args[1].Type())
	}

	result := formatDateTime(formatVal.Value, dtVal.Value)
	return &StringValue{Value: result}
}

// builtinDateTimeToStr implements the DateTimeToStr() built-in function.
// Converts a TDateTime to a string using default format.
func (i *Interpreter) builtinDateTimeToStr(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "DateTimeToStr() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "DateTimeToStr() expects Float/TDateTime, got %s", args[0].Type())
	}

	// Use default format: YYYY-MM-DD HH:MM:SS
	result := formatDateTime("yyyy-mm-dd hh:nn:ss", dtVal.Value)
	return &StringValue{Value: result}
}

// builtinDateToStr implements the DateToStr() built-in function.
// Converts a TDateTime to a date string.
func (i *Interpreter) builtinDateToStr(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "DateToStr() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "DateToStr() expects Float/TDateTime, got %s", args[0].Type())
	}

	// Use default date format: YYYY-MM-DD
	result := formatDateTime("yyyy-mm-dd", dtVal.Value)
	return &StringValue{Value: result}
}

// builtinTimeToStr implements the TimeToStr() built-in function.
// Converts a TDateTime to a time string.
func (i *Interpreter) builtinTimeToStr(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "TimeToStr() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "TimeToStr() expects Float/TDateTime, got %s", args[0].Type())
	}

	// Use default time format: HH:MM:SS
	result := formatDateTime("hh:nn:ss", dtVal.Value)
	return &StringValue{Value: result}
}

// builtinDateToISO8601 implements the DateToISO8601() built-in function.
// Formats date as ISO 8601 string (YYYY-MM-DD).
func (i *Interpreter) builtinDateToISO8601(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "DateToISO8601() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "DateToISO8601() expects Float/TDateTime, got %s", args[0].Type())
	}

	result := formatDateISO8601(dtVal.Value)
	return &StringValue{Value: result}
}

// builtinDateTimeToISO8601 implements the DateTimeToISO8601() built-in function.
// Formats datetime as ISO 8601 string (YYYY-MM-DDTHH:MM:SS).
func (i *Interpreter) builtinDateTimeToISO8601(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "DateTimeToISO8601() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "DateTimeToISO8601() expects Float/TDateTime, got %s", args[0].Type())
	}

	result := formatISO8601(dtVal.Value)
	return &StringValue{Value: result}
}

// builtinDateTimeToRFC822 implements the DateTimeToRFC822() built-in function.
// Formats datetime as RFC 822 string.
func (i *Interpreter) builtinDateTimeToRFC822(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "DateTimeToRFC822() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "DateTimeToRFC822() expects Float/TDateTime, got %s", args[0].Type())
	}

	result := formatRFC822(dtVal.Value)
	return &StringValue{Value: result}
}

// ============================================================================
// Parsing Functions
// ============================================================================

// builtinStrToDate implements the StrToDate() built-in function.
// Parses a date string to TDateTime.
func (i *Interpreter) builtinStrToDate(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "StrToDate() expects 1 argument, got %d", len(args))
	}

	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToDate() expects String, got %s", args[0].Type())
	}

	dt, err := parseDate(strVal.Value)
	if err != nil {
		return i.newErrorWithLocation(i.currentNode, "StrToDate() %s", err)
	}

	return &FloatValue{Value: dt}
}

// builtinStrToDateTime implements the StrToDateTime() built-in function.
// Parses a datetime string to TDateTime.
func (i *Interpreter) builtinStrToDateTime(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "StrToDateTime() expects 1 argument, got %d", len(args))
	}

	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToDateTime() expects String, got %s", args[0].Type())
	}

	dt, err := parseDateTime(strVal.Value)
	if err != nil {
		return i.newErrorWithLocation(i.currentNode, "StrToDateTime() %s", err)
	}

	return &FloatValue{Value: dt}
}

// builtinStrToTime implements the StrToTime() built-in function.
// Parses a time string to TDateTime.
func (i *Interpreter) builtinStrToTime(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "StrToTime() expects 1 argument, got %d", len(args))
	}

	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToTime() expects String, got %s", args[0].Type())
	}

	dt, err := parseTime(strVal.Value)
	if err != nil {
		return i.newErrorWithLocation(i.currentNode, "StrToTime() %s", err)
	}

	return &FloatValue{Value: dt}
}

// builtinISO8601ToDateTime implements the ISO8601ToDateTime() built-in function.
// Parses an ISO 8601 string to TDateTime.
func (i *Interpreter) builtinISO8601ToDateTime(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "ISO8601ToDateTime() expects 1 argument, got %d", len(args))
	}

	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "ISO8601ToDateTime() expects String, got %s", args[0].Type())
	}

	dt, err := parseISO8601(strVal.Value)
	if err != nil {
		return i.newErrorWithLocation(i.currentNode, "ISO8601ToDateTime() %s", err)
	}

	return &FloatValue{Value: dt}
}

// builtinRFC822ToDateTime implements the RFC822ToDateTime() built-in function.
// Parses an RFC 822 string to TDateTime.
func (i *Interpreter) builtinRFC822ToDateTime(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "RFC822ToDateTime() expects 1 argument, got %d", len(args))
	}

	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "RFC822ToDateTime() expects String, got %s", args[0].Type())
	}

	dt, err := parseRFC822(strVal.Value)
	if err != nil {
		return i.newErrorWithLocation(i.currentNode, "RFC822ToDateTime() %s", err)
	}

	return &FloatValue{Value: dt}
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

// ============================================================================
// Unix Time Functions
// ============================================================================

// builtinUnixTime implements the UnixTime() built-in function.
// Returns Unix timestamp (seconds since 1970-01-01) for current time.
func (i *Interpreter) builtinUnixTime(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "UnixTime() expects 0 arguments, got %d", len(args))
	}

	now := time.Now().UTC()
	return &IntegerValue{Value: now.Unix()}
}

// builtinUnixTimeMSec implements the UnixTimeMSec() built-in function.
// Returns Unix timestamp in milliseconds for current time.
func (i *Interpreter) builtinUnixTimeMSec(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "UnixTimeMSec() expects 0 arguments, got %d", len(args))
	}

	now := time.Now().UTC()
	return &IntegerValue{Value: now.UnixMilli()}
}

// builtinUnixTimeToDateTime implements the UnixTimeToDateTime() built-in function.
// Converts Unix timestamp to TDateTime.
func (i *Interpreter) builtinUnixTimeToDateTime(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "UnixTimeToDateTime() expects 1 argument, got %d", len(args))
	}

	unixTimeVal, ok := args[0].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "UnixTimeToDateTime() expects Integer, got %s", args[0].Type())
	}

	dt := unixTimeToDateTime(unixTimeVal.Value)
	return &FloatValue{Value: dt}
}

// builtinDateTimeToUnixTime implements the DateTimeToUnixTime() built-in function.
// Converts TDateTime to Unix timestamp.
func (i *Interpreter) builtinDateTimeToUnixTime(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "DateTimeToUnixTime() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "DateTimeToUnixTime() expects Float/TDateTime, got %s", args[0].Type())
	}

	unixTime := dateTimeToUnixTime(dtVal.Value)
	return &IntegerValue{Value: unixTime}
}

// builtinUnixTimeMSecToDateTime implements the UnixTimeMSecToDateTime() built-in function.
// Converts Unix timestamp in milliseconds to TDateTime.
func (i *Interpreter) builtinUnixTimeMSecToDateTime(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "UnixTimeMSecToDateTime() expects 1 argument, got %d", len(args))
	}

	unixTimeMSVal, ok := args[0].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "UnixTimeMSecToDateTime() expects Integer, got %s", args[0].Type())
	}

	dt := unixTimeMSecToDateTime(unixTimeMSVal.Value)
	return &FloatValue{Value: dt}
}

// builtinDateTimeToUnixTimeMSec implements the DateTimeToUnixTimeMSec() built-in function.
// Converts TDateTime to Unix timestamp in milliseconds.
func (i *Interpreter) builtinDateTimeToUnixTimeMSec(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "DateTimeToUnixTimeMSec() expects 1 argument, got %d", len(args))
	}

	dtVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "DateTimeToUnixTimeMSec() expects Float/TDateTime, got %s", args[0].Type())
	}

	unixTimeMS := dateTimeToUnixTimeMSec(dtVal.Value)
	return &IntegerValue{Value: unixTimeMS}
}
