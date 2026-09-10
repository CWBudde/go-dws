package builtins

import (
	"math"
	"time"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Argument helpers
// =============================================================================

// dateTimeArg reads a TDateTime argument. TDateTime is a Float alias, so an
// Integer argument converts implicitly, exactly as in DWScript.
func dateTimeArg(ctx Context, args []Value, index int, fn string) (float64, Value) {
	value, ok := ctx.ToFloat64(args[index])
	if !ok {
		return 0, ctx.NewError("%s() expects Float/TDateTime, got %s", fn, args[index].Type())
	}
	return value, nil
}

// dateTimeArgOrNow reads a TDateTime argument, substituting the current local
// date and time for a zero value. Several DWScript date functions use zero as
// "unspecified".
func dateTimeArgOrNow(ctx Context, args []Value, index int, fn string) (float64, Value) {
	dt, errVal := dateTimeArg(ctx, args, index, fn)
	if errVal != nil {
		return 0, errVal
	}
	if dt == 0 {
		dt = nowDateTime()
	}
	return dt, nil
}

// integerArg reads an Integer argument.
func integerArg(ctx Context, args []Value, index int, fn string) (int64, Value) {
	value, ok := ctx.ToInt64(args[index])
	if !ok {
		return 0, ctx.NewError("%s() expects Integer, got %s", fn, args[index].Type())
	}
	return value, nil
}

// stringArg reads a String argument.
func stringArg(ctx Context, args []Value, index int, fn string) (string, Value) {
	strVal, ok := ctx.UnwrapVariant(args[index]).(*runtime.StringValue)
	if !ok {
		return "", ctx.NewError("%s() expects String, got %s", fn, args[index].Type())
	}
	return strVal.Value, nil
}

// zoneArg reads an optional trailing DateTimeZone argument. An absent argument
// means TimeZoneDefault, which defers to FormatSettings.Zone.
func zoneArg(ctx Context, args []Value, index int) TimeZone {
	if index >= len(args) {
		return TimeZoneDefault
	}
	ordinal, ok := ctx.ToInt64(args[index])
	if !ok {
		if ordinal, ok = ctx.GetEnumOrdinal(args[index]); !ok {
			return TimeZoneDefault
		}
	}
	switch ordinal {
	case int64(TimeZoneLocal):
		return TimeZoneLocal
	case int64(TimeZoneUTC):
		return TimeZoneUTC
	default:
		return TimeZoneDefault
	}
}

// settings returns the running script's date/time format settings.
func settings(ctx Context) *DateTimeFormatSettings {
	if s := ctx.DateTimeFormatSettings(); s != nil {
		return s
	}
	fallback := DefaultDateTimeFormatSettings()
	return &fallback
}

func floatResult(v float64) Value { return &runtime.FloatValue{Value: v} }
func intResult(v int) Value       { return &runtime.IntegerValue{Value: int64(v)} }
func int64Result(v int64) Value   { return &runtime.IntegerValue{Value: v} }
func stringResult(v string) Value { return &runtime.StringValue{Value: v} }
func boolResult(v bool) Value     { return &runtime.BooleanValue{Value: v} }

// =============================================================================
// Current Date/Time Functions
// =============================================================================

// Now implements the Now() built-in function.
// Returns the current local date and time as TDateTime.
func Now(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("Now() expects 0 arguments, got %d", len(args))
	}
	return floatResult(nowDateTime())
}

// Date implements the Date() built-in function.
// Returns the current local date with a zero time component.
func Date(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("Date() expects 0 arguments, got %d", len(args))
	}
	return floatResult(math.Trunc(nowDateTime()))
}

// Time implements the Time() built-in function.
// Returns the current local time of day with a zero date component.
func Time(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("Time() expects 0 arguments, got %d", len(args))
	}
	now := nowDateTime()
	return floatResult(now - math.Trunc(now))
}

// UTCDateTime implements the UTCDateTime() built-in function.
// Returns the current UTC date and time as TDateTime.
func UTCDateTime(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("UTCDateTime() expects 0 arguments, got %d", len(args))
	}
	return floatResult(utcNowDateTime())
}

// LocalDateTimeToUTCDateTime implements the LocalDateTimeToUTCDateTime()
// built-in function, reinterpreting a local TDateTime as UTC.
func LocalDateTimeToUTCDateTime(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("LocalDateTimeToUTCDateTime() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArg(ctx, args, 0, "LocalDateTimeToUTCDateTime")
	if errVal != nil {
		return errVal
	}
	return floatResult(localDateTimeToUTCDateTime(dt))
}

// UTCDateTimeToLocalDateTime implements the UTCDateTimeToLocalDateTime()
// built-in function, reinterpreting a UTC TDateTime as local time.
func UTCDateTimeToLocalDateTime(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("UTCDateTimeToLocalDateTime() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArg(ctx, args, 0, "UTCDateTimeToLocalDateTime")
	if errVal != nil {
		return errVal
	}
	return floatResult(utcDateTimeToLocalDateTime(dt))
}

// Sleep implements the Sleep() built-in procedure, pausing the script for the
// given number of milliseconds.
func Sleep(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Sleep() expects 1 argument, got %d", len(args))
	}
	msec, errVal := integerArg(ctx, args, 0, "Sleep")
	if errVal != nil {
		return errVal
	}
	if msec > 0 {
		time.Sleep(time.Duration(msec) * time.Millisecond)
	}
	return nil
}

// =============================================================================
// Component Extraction Functions
// =============================================================================

// YearOf implements the YearOf() built-in function.
func YearOf(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("YearOf() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArgOrNow(ctx, args, 0, "YearOf")
	if errVal != nil {
		return errVal
	}
	return intResult(decodeDateTime(dt).Year)
}

// MonthOf implements the MonthOf() built-in function.
func MonthOf(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("MonthOf() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArgOrNow(ctx, args, 0, "MonthOf")
	if errVal != nil {
		return errVal
	}
	return intResult(decodeDateTime(dt).Month)
}

// MonthOfYear implements the MonthOfYear() built-in function.
func MonthOfYear(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("MonthOfYear() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArgOrNow(ctx, args, 0, "MonthOfYear")
	if errVal != nil {
		return errVal
	}
	return intResult(decodeDateTime(dt).Month)
}

// DayOf implements the DayOf() built-in function.
func DayOf(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DayOf() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArgOrNow(ctx, args, 0, "DayOf")
	if errVal != nil {
		return errVal
	}
	return intResult(decodeDateTime(dt).Day)
}

// DayOfMonth implements the DayOfMonth() built-in function.
func DayOfMonth(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DayOfMonth() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArgOrNow(ctx, args, 0, "DayOfMonth")
	if errVal != nil {
		return errVal
	}
	return intResult(decodeDateTime(dt).Day)
}

// HourOf implements the HourOf() built-in function.
func HourOf(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("HourOf() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArg(ctx, args, 0, "HourOf")
	if errVal != nil {
		return errVal
	}
	return intResult(decodeDateTime(dt).Hour)
}

// MinuteOf implements the MinuteOf() built-in function.
func MinuteOf(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("MinuteOf() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArg(ctx, args, 0, "MinuteOf")
	if errVal != nil {
		return errVal
	}
	return intResult(decodeDateTime(dt).Minute)
}

// SecondOf implements the SecondOf() built-in function.
func SecondOf(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("SecondOf() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArg(ctx, args, 0, "SecondOf")
	if errVal != nil {
		return errVal
	}
	return intResult(decodeDateTime(dt).Second)
}

// DayOfWeek implements the DayOfWeek() built-in function.
// Returns 1 for Sunday through 7 for Saturday, as in Delphi.
func DayOfWeek(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DayOfWeek() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArg(ctx, args, 0, "DayOfWeek")
	if errVal != nil {
		return errVal
	}
	return intResult(dayOfWeek(dt))
}

// DayOfTheWeek implements the DayOfTheWeek() built-in function.
// Returns 1 for Monday through 7 for Sunday, as in ISO 8601.
func DayOfTheWeek(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DayOfTheWeek() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArg(ctx, args, 0, "DayOfTheWeek")
	if errVal != nil {
		return errVal
	}
	return intResult((dayOfWeek(dt)+5)%7 + 1)
}

// DayOfYear implements the DayOfYear() built-in function.
func DayOfYear(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DayOfYear() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArgOrNow(ctx, args, 0, "DayOfYear")
	if errVal != nil {
		return errVal
	}
	return intResult(dayOfYear(dt))
}

// WeekNumber implements the WeekNumber() built-in function.
func WeekNumber(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("WeekNumber() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArgOrNow(ctx, args, 0, "WeekNumber")
	if errVal != nil {
		return errVal
	}
	return intResult(dateToWeekNumber(dt))
}

// DateToWeekNumber implements the DateToWeekNumber() built-in function.
func DateToWeekNumber(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DateToWeekNumber() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArgOrNow(ctx, args, 0, "DateToWeekNumber")
	if errVal != nil {
		return errVal
	}
	return intResult(dateToWeekNumber(dt))
}

// YearOfWeek implements the YearOfWeek() built-in function.
func YearOfWeek(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("YearOfWeek() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArgOrNow(ctx, args, 0, "YearOfWeek")
	if errVal != nil {
		return errVal
	}
	return intResult(yearOfWeek(dt))
}

// DateToYearOfWeek implements the DateToYearOfWeek() built-in function.
func DateToYearOfWeek(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DateToYearOfWeek() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArgOrNow(ctx, args, 0, "DateToYearOfWeek")
	if errVal != nil {
		return errVal
	}
	return intResult(yearOfWeek(dt))
}

// =============================================================================
// Date Calculation Functions
// =============================================================================

// IsLeapYear implements the IsLeapYear() built-in function.
func IsLeapYear(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("IsLeapYear() expects 1 argument, got %d", len(args))
	}
	year, errVal := integerArg(ctx, args, 0, "IsLeapYear")
	if errVal != nil {
		return errVal
	}
	return boolResult(isLeapYear(int(year)))
}

// FirstDayOfYear implements the FirstDayOfYear() built-in function.
func FirstDayOfYear(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("FirstDayOfYear() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArgOrNow(ctx, args, 0, "FirstDayOfYear")
	if errVal != nil {
		return errVal
	}
	c := decodeDateTime(dt)
	return floatResult(encodeDateOnly(c.Year, 1, 1))
}

// FirstDayOfNextYear implements the FirstDayOfNextYear() built-in function.
func FirstDayOfNextYear(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("FirstDayOfNextYear() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArgOrNow(ctx, args, 0, "FirstDayOfNextYear")
	if errVal != nil {
		return errVal
	}
	c := decodeDateTime(dt)
	return floatResult(encodeDateOnly(c.Year+1, 1, 1))
}

// FirstDayOfMonth implements the FirstDayOfMonth() built-in function.
func FirstDayOfMonth(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("FirstDayOfMonth() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArgOrNow(ctx, args, 0, "FirstDayOfMonth")
	if errVal != nil {
		return errVal
	}
	c := decodeDateTime(dt)
	return floatResult(encodeDateOnly(c.Year, c.Month, 1))
}

// FirstDayOfNextMonth implements the FirstDayOfNextMonth() built-in function.
func FirstDayOfNextMonth(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("FirstDayOfNextMonth() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArgOrNow(ctx, args, 0, "FirstDayOfNextMonth")
	if errVal != nil {
		return errVal
	}
	c := decodeDateTime(dt)
	year, month := c.Year, c.Month+1
	if month > 12 {
		month = 1
		year++
	}
	return floatResult(encodeDateOnly(year, month, 1))
}

// FirstDayOfWeek implements the FirstDayOfWeek() built-in function.
// The week starts on Monday.
func FirstDayOfWeek(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("FirstDayOfWeek() expects 1 argument, got %d", len(args))
	}
	dt, errVal := dateTimeArgOrNow(ctx, args, 0, "FirstDayOfWeek")
	if errVal != nil {
		return errVal
	}
	// Delphi DayOfWeek is 1=Sunday..7=Saturday; the offset table maps it to the
	// number of days back to the preceding Monday.
	offsets := [8]int{0, 6, 0, 1, 2, 3, 4, 5}
	return floatResult(math.Trunc(dt) - float64(offsets[dayOfWeek(dt)]))
}

// =============================================================================
// Helper Functions
// =============================================================================

// isValidDate reports whether year, month and day form a valid calendar date.
func isValidDate(year, month, day int) bool {
	if year < 1 || year > 9999 || month < 1 || month > 12 || day < 1 {
		return false
	}
	return day <= daysInMonth(year, month)
}

// isValidTime reports whether the clock fields form a valid time of day.
func isValidTime(hour, minute, second, millisecond int) bool {
	return hour >= 0 && hour < 24 &&
		minute >= 0 && minute < 60 &&
		second >= 0 && second < 60 &&
		millisecond >= 0 && millisecond < 1000
}

// isLeapYear reports whether a year is a leap year in the Gregorian calendar.
func isLeapYear(year int) bool {
	return (year%4 == 0 && year%100 != 0) || year%400 == 0
}

// daysInMonth returns the number of days in a month of a given year.
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

// incMonths adds a signed number of months to a TDateTime, clamping the day of
// month to the length of the target month.
func incMonths(dt float64, months int) float64 {
	c := decodeDateTime(dt)
	year := c.Year
	month := c.Month + months
	for month > 12 {
		month -= 12
		year++
	}
	for month < 1 {
		month += 12
		year--
	}
	if maxDay := daysInMonth(year, month); c.Day > maxDay {
		c.Day = maxDay
	}
	return encodeCivil(year, month, c.Day, c.Hour, c.Minute, c.Second, c.Millisecond)
}

// daysBetween returns the number of whole days between two TDateTime values.
func daysBetween(dt1, dt2 float64) int {
	return int(math.Floor(math.Abs(dt1 - dt2)))
}

// hoursBetween returns the number of whole hours between two TDateTime values.
func hoursBetween(dt1, dt2 float64) int {
	return int(math.Floor(math.Abs(dt1-dt2) * 24.0))
}

// minutesBetween returns the number of whole minutes between two TDateTime values.
func minutesBetween(dt1, dt2 float64) int {
	return int(math.Floor(math.Abs(dt1-dt2) * 1440.0))
}

// secondsBetween returns the number of whole seconds between two TDateTime values.
func secondsBetween(dt1, dt2 float64) int {
	return int(math.Floor(math.Abs(dt1-dt2) * secondsPerDay))
}

// dayOfWeek returns Delphi's day of week: 1 for Sunday through 7 for Saturday.
func dayOfWeek(dt float64) int {
	dayNumber, _ := dateTimeSplit(dt)
	return int(dateTimeEpoch.AddDate(0, 0, dayNumber).Weekday()) + 1
}

// dayOfYear returns the 1-based day number within the year.
func dayOfYear(dt float64) int {
	c := decodeDateTime(dt)
	return int(math.Trunc(dt)-encodeDateOnly(c.Year, 1, 1)) + 1
}

// dateToWeekNumber returns the ISO 8601 week number of a date.
func dateToWeekNumber(dt float64) int {
	const (
		isoFirstWeekDay  = 2 // Monday
		isoFirstWeekDays = 4 // the first week has at least four days
	)
	weekDay := (dayOfWeek(dt)-isoFirstWeekDay+7)%7 + 1
	shifted := math.Trunc(dt) - float64(weekDay) + 8 - isoFirstWeekDays
	c := decodeDateTime(shifted)
	return int(shifted-encodeDateOnly(c.Year, 1, 1))/7 + 1
}

// yearOfWeek returns the ISO 8601 week-numbering year of a date.
func yearOfWeek(dt float64) int {
	c := decodeDateTime(dt)
	switch {
	case c.Month == 1 && c.Day < 4:
		// The week of an early January day may belong to the previous year.
		if dateToWeekNumber(dt) == 1 {
			return c.Year
		}
		return c.Year - 1
	case c.Month == 12 && c.Day >= 29:
		// The week of a late December day may belong to the next year.
		if dateToWeekNumber(dt) == 1 {
			return c.Year + 1
		}
		return c.Year
	default:
		return c.Year
	}
}
