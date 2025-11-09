package semantic

import (
	"testing"
)

// ============================================================================
// Built-in DateTime Functions Tests
// ============================================================================
// These tests cover the built-in date/time functions to improve
// coverage of analyze_builtin_datetime.go (currently at 0% coverage)

// Basic datetime retrieval functions
func TestBuiltinNow_Basic(t *testing.T) {
	input := `
		var dt := Now();
	`
	expectNoErrors(t, input)
}

func TestBuiltinDate_Basic(t *testing.T) {
	input := `
		var dt := Date();
	`
	expectNoErrors(t, input)
}

func TestBuiltinTime_Basic(t *testing.T) {
	input := `
		var dt := Time();
	`
	expectNoErrors(t, input)
}

func TestBuiltinUTCDateTime_Basic(t *testing.T) {
	input := `
		var dt := UTCDateTime();
	`
	expectNoErrors(t, input)
}

func TestBuiltinUnixTime_Basic(t *testing.T) {
	input := `
		var t := UnixTime();
	`
	expectNoErrors(t, input)
}

func TestBuiltinUnixTimeMSec_Basic(t *testing.T) {
	input := `
		var t := UnixTimeMSec();
	`
	expectNoErrors(t, input)
}

// Date/time encoding functions
func TestBuiltinEncodeDate_Basic(t *testing.T) {
	input := `
		var dt := EncodeDate(2024, 1, 15);
	`
	expectNoErrors(t, input)
}

func TestBuiltinEncodeDate_LeapYear(t *testing.T) {
	input := `
		var dt := EncodeDate(2024, 2, 29);
	`
	expectNoErrors(t, input)
}

func TestBuiltinEncodeTime_Basic(t *testing.T) {
	input := `
		var dt := EncodeTime(14, 30, 45, 0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinEncodeTime_Midnight(t *testing.T) {
	input := `
		var dt := EncodeTime(0, 0, 0, 0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinEncodeDateTime_Basic(t *testing.T) {
	input := `
		var dt := EncodeDateTime(2024, 1, 15, 14, 30, 45, 0);
	`
	expectNoErrors(t, input)
}

// Date/time decoding functions
func TestBuiltinDecodeDate_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var year, month, day: Integer;
		DecodeDate(dt, year, month, day);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDecodeTime_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var hour, min, sec, msec: Integer;
		DecodeTime(dt, hour, min, sec, msec);
	`
	expectNoErrors(t, input)
}

// Date/time component extraction functions
func TestBuiltinYearOf_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var year := YearOf(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinMonthOf_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var month := MonthOf(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDayOf_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var day := DayOf(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinHourOf_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var hour := HourOf(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinMinuteOf_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var minute := MinuteOf(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinSecondOf_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var second := SecondOf(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDayOfWeek_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var dow := DayOfWeek(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDayOfTheWeek_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var dow := DayOfTheWeek(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDayOfYear_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var doy := DayOfYear(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinWeekNumber_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var week := WeekNumber(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinYearOfWeek_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var year := YearOfWeek(dt);
	`
	expectNoErrors(t, input)
}

// Date/time formatting functions
func TestBuiltinFormatDateTime_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var s := FormatDateTime('yyyy-mm-dd', dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinFormatDateTime_WithTime(t *testing.T) {
	input := `
		var dt := Now();
		var s := FormatDateTime('yyyy-mm-dd hh:nn:ss', dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateTimeToStr_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var s := DateTimeToStr(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateToStr_Basic(t *testing.T) {
	input := `
		var dt := Date();
		var s := DateToStr(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinTimeToStr_Basic(t *testing.T) {
	input := `
		var dt := Time();
		var s := TimeToStr(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateToISO8601_Basic(t *testing.T) {
	input := `
		var dt := Date();
		var s := DateToISO8601(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateTimeToISO8601_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var s := DateTimeToISO8601(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateTimeToRFC822_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var s := DateTimeToRFC822(dt);
	`
	expectNoErrors(t, input)
}

// Date/time parsing functions
func TestBuiltinStrToDate_Basic(t *testing.T) {
	input := `
		var dt := StrToDate('2024-01-15');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStrToDateTime_Basic(t *testing.T) {
	input := `
		var dt := StrToDateTime('2024-01-15 14:30:45');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStrToTime_Basic(t *testing.T) {
	input := `
		var dt := StrToTime('14:30:45');
	`
	expectNoErrors(t, input)
}

func TestBuiltinISO8601ToDateTime_Basic(t *testing.T) {
	input := `
		var dt := ISO8601ToDateTime('2024-01-15T14:30:45Z');
	`
	expectNoErrors(t, input)
}

func TestBuiltinRFC822ToDateTime_Basic(t *testing.T) {
	input := `
		var dt := RFC822ToDateTime('Mon, 15 Jan 2024 14:30:45 GMT');
	`
	expectNoErrors(t, input)
}

// Date/time manipulation functions
func TestBuiltinIncYear_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var next := IncYear(dt, 1);
	`
	expectNoErrors(t, input)
}

func TestBuiltinIncYear_Negative(t *testing.T) {
	input := `
		var dt := Now();
		var prev := IncYear(dt, -1);
	`
	expectNoErrors(t, input)
}

func TestBuiltinIncMonth_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var next := IncMonth(dt, 3);
	`
	expectNoErrors(t, input)
}

func TestBuiltinIncDay_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var tomorrow := IncDay(dt, 1);
	`
	expectNoErrors(t, input)
}

func TestBuiltinIncHour_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var next := IncHour(dt, 2);
	`
	expectNoErrors(t, input)
}

func TestBuiltinIncMinute_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var next := IncMinute(dt, 30);
	`
	expectNoErrors(t, input)
}

func TestBuiltinIncSecond_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var next := IncSecond(dt, 45);
	`
	expectNoErrors(t, input)
}

// Date/time difference functions
func TestBuiltinDaysBetween_Basic(t *testing.T) {
	input := `
		var dt1 := EncodeDate(2024, 1, 1);
		var dt2 := EncodeDate(2024, 1, 15);
		var days := DaysBetween(dt1, dt2);
	`
	expectNoErrors(t, input)
}

func TestBuiltinHoursBetween_Basic(t *testing.T) {
	input := `
		var dt1 := Now();
		var dt2 := IncHour(dt1, 5);
		var hours := HoursBetween(dt1, dt2);
	`
	expectNoErrors(t, input)
}

func TestBuiltinMinutesBetween_Basic(t *testing.T) {
	input := `
		var dt1 := Now();
		var dt2 := IncMinute(dt1, 30);
		var minutes := MinutesBetween(dt1, dt2);
	`
	expectNoErrors(t, input)
}

func TestBuiltinSecondsBetween_Basic(t *testing.T) {
	input := `
		var dt1 := Now();
		var dt2 := IncSecond(dt1, 120);
		var seconds := SecondsBetween(dt1, dt2);
	`
	expectNoErrors(t, input)
}

// Leap year and special date functions
func TestBuiltinIsLeapYear_Basic(t *testing.T) {
	input := `
		var isLeap := IsLeapYear(2024);
	`
	expectNoErrors(t, input)
}

func TestBuiltinIsLeapYear_NotLeap(t *testing.T) {
	input := `
		var isLeap := IsLeapYear(2023);
	`
	expectNoErrors(t, input)
}

func TestBuiltinFirstDayOfYear_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var first := FirstDayOfYear(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinFirstDayOfNextYear_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var first := FirstDayOfNextYear(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinFirstDayOfMonth_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var first := FirstDayOfMonth(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinFirstDayOfNextMonth_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var first := FirstDayOfNextMonth(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinFirstDayOfWeek_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var first := FirstDayOfWeek(dt);
	`
	expectNoErrors(t, input)
}

// Unix time conversion functions
func TestBuiltinUnixTimeToDateTime_Basic(t *testing.T) {
	input := `
		var dt := UnixTimeToDateTime(1705329045);
	`
	expectNoErrors(t, input)
}

func TestBuiltinUnixTimeMSecToDateTime_Basic(t *testing.T) {
	input := `
		var dt := UnixTimeMSecToDateTime(1705329045000);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateTimeToUnixTime_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var unix := DateTimeToUnixTime(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateTimeToUnixTimeMSec_Basic(t *testing.T) {
	input := `
		var dt := Now();
		var unix := DateTimeToUnixTimeMSec(dt);
	`
	expectNoErrors(t, input)
}

// Combined datetime operations tests
func TestBuiltinDateTime_CompleteWorkflow(t *testing.T) {
	input := `
		var dt := Now();
		var year := YearOf(dt);
		var month := MonthOf(dt);
		var day := DayOf(dt);
		var str := DateTimeToStr(dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateTime_ManipulationChain(t *testing.T) {
	input := `
		var dt := Now();
		dt := IncDay(dt, 7);
		dt := IncHour(dt, -2);
		dt := IncMinute(dt, 30);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateTime_Comparison(t *testing.T) {
	input := `
		var dt1 := Now();
		var dt2 := IncDay(dt1, 1);
		var daysDiff := DaysBetween(dt1, dt2);
		var isAfter := dt2 > dt1;
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateTime_Formatting(t *testing.T) {
	input := `
		var dt := Now();
		var iso := DateTimeToISO8601(dt);
		var rfc := DateTimeToRFC822(dt);
		var custom := FormatDateTime('dd/mm/yyyy', dt);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateTime_RoundTrip(t *testing.T) {
	input := `
		var dt1 := Now();
		var str := DateTimeToStr(dt1);
		var dt2 := StrToDateTime(str);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateTime_UnixTimeRoundTrip(t *testing.T) {
	input := `
		var dt1 := Now();
		var unix := DateTimeToUnixTime(dt1);
		var dt2 := UnixTimeToDateTime(unix);
	`
	expectNoErrors(t, input)
}

// Edge cases
func TestBuiltinDateTime_LeapYearFeb29(t *testing.T) {
	input := `
		var dt := EncodeDate(2024, 2, 29);
		var isLeap := IsLeapYear(2024);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateTime_Epoch(t *testing.T) {
	input := `
		var dt := UnixTimeToDateTime(0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateTime_FarFuture(t *testing.T) {
	input := `
		var dt := EncodeDate(2099, 12, 31);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateTime_InFunction(t *testing.T) {
	input := `
		function GetAge(birthDate: Float): Integer;
		begin
			var today := Now();
			Result := YearOf(today) - YearOf(birthDate);
		end;

		var age := GetAge(EncodeDate(1990, 5, 15));
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateTime_DaysBetweenDates(t *testing.T) {
	input := `
		var start := EncodeDate(2024, 1, 1);
		var end := EncodeDate(2024, 12, 31);
		var days := DaysBetween(start, end);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateTime_TimeSpan(t *testing.T) {
	input := `
		var dt1 := Now();
		var dt2 := IncHour(IncDay(dt1, 5), 12);
		var hoursDiff := HoursBetween(dt1, dt2);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDateTime_WeekCalculations(t *testing.T) {
	input := `
		var dt := Now();
		var dow := DayOfWeek(dt);
		var week := WeekNumber(dt);
		var first := FirstDayOfWeek(dt);
	`
	expectNoErrors(t, input)
}
