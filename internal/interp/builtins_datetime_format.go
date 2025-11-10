package interp

import (
	"time"
)

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
