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
