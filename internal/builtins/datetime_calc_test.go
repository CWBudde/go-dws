package builtins

import (
	"math"
	"testing"
	"time"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Date Encoding Functions Tests
// =============================================================================

func TestEncodeDate(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name    string
		year    int64
		month   int64
		day     int64
		isError bool
	}{
		{
			name:    "valid date",
			year:    2023,
			month:   3,
			day:     15,
			isError: false,
		},
		{
			name:    "leap year Feb 29",
			year:    2020,
			month:   2,
			day:     29,
			isError: false,
		},
		{
			name:    "invalid Feb 29 non-leap year",
			year:    2023,
			month:   2,
			day:     29,
			isError: true,
		},
		{
			name:    "invalid month",
			year:    2023,
			month:   13,
			day:     15,
			isError: true,
		},
		{
			name:    "invalid day",
			year:    2023,
			month:   3,
			day:     32,
			isError: true,
		},
		{
			name:    "invalid year (too high)",
			year:    10000,
			month:   1,
			day:     1,
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeDate(ctx, []Value{
				&runtime.IntegerValue{Value: tt.year},
				&runtime.IntegerValue{Value: tt.month},
				&runtime.IntegerValue{Value: tt.day},
			})

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			floatVal, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}

			// Verify by decoding back
			dt := delphiDateTimeToGoTime(floatVal.Value)
			if dt.Year() != int(tt.year) || int(dt.Month()) != int(tt.month) || dt.Day() != int(tt.day) {
				t.Errorf("EncodeDate(%d,%d,%d) decoded to %v", tt.year, tt.month, tt.day, dt)
			}
		})
	}

	// Test argument count errors
	result := EncodeDate(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("EncodeDate with 0 args should error")
	}

	result = EncodeDate(ctx, []Value{
		&runtime.IntegerValue{Value: 2023},
		&runtime.IntegerValue{Value: 3},
	})
	if result.Type() != "ERROR" {
		t.Errorf("EncodeDate with 2 args should error")
	}

	// Test type errors
	result = EncodeDate(ctx, []Value{
		&runtime.StringValue{Value: "2023"},
		&runtime.IntegerValue{Value: 3},
		&runtime.IntegerValue{Value: 15},
	})
	if result.Type() != "ERROR" {
		t.Errorf("EncodeDate with wrong type should error")
	}
}

func TestEncodeTime(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name    string
		hour    int64
		minute  int64
		second  int64
		msec    int64
		isError bool
	}{
		{
			name:    "valid time",
			hour:    12,
			minute:  30,
			second:  45,
			msec:    123,
			isError: false,
		},
		{
			name:    "midnight",
			hour:    0,
			minute:  0,
			second:  0,
			msec:    0,
			isError: false,
		},
		{
			name:    "almost midnight (23:59:59.999)",
			hour:    23,
			minute:  59,
			second:  59,
			msec:    999,
			isError: false,
		},
		{
			name:    "invalid hour",
			hour:    24,
			minute:  0,
			second:  0,
			msec:    0,
			isError: true,
		},
		{
			name:    "invalid minute",
			hour:    12,
			minute:  60,
			second:  0,
			msec:    0,
			isError: true,
		},
		{
			name:    "invalid second",
			hour:    12,
			minute:  30,
			second:  60,
			msec:    0,
			isError: true,
		},
		{
			name:    "invalid millisecond",
			hour:    12,
			minute:  30,
			second:  45,
			msec:    1000,
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeTime(ctx, []Value{
				&runtime.IntegerValue{Value: tt.hour},
				&runtime.IntegerValue{Value: tt.minute},
				&runtime.IntegerValue{Value: tt.second},
				&runtime.IntegerValue{Value: tt.msec},
			})

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			floatVal, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}

			// Verify by decoding back
			dt := delphiDateTimeToGoTime(floatVal.Value)
			if dt.Hour() != int(tt.hour) || dt.Minute() != int(tt.minute) || dt.Second() != int(tt.second) {
				t.Errorf("EncodeTime(%d,%d,%d,%d) decoded to %v", tt.hour, tt.minute, tt.second, tt.msec, dt)
			}
		})
	}
}

func TestEncodeDateTime(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name    string
		year    int64
		month   int64
		day     int64
		hour    int64
		minute  int64
		second  int64
		msec    int64
		isError bool
	}{
		{
			name:    "valid datetime",
			year:    2023,
			month:   3,
			day:     15,
			hour:    12,
			minute:  30,
			second:  45,
			msec:    123,
			isError: false,
		},
		{
			name:    "invalid date component",
			year:    2023,
			month:   13,
			day:     15,
			hour:    12,
			minute:  30,
			second:  45,
			msec:    0,
			isError: true,
		},
		{
			name:    "invalid time component",
			year:    2023,
			month:   3,
			day:     15,
			hour:    25,
			minute:  30,
			second:  45,
			msec:    0,
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeDateTime(ctx, []Value{
				&runtime.IntegerValue{Value: tt.year},
				&runtime.IntegerValue{Value: tt.month},
				&runtime.IntegerValue{Value: tt.day},
				&runtime.IntegerValue{Value: tt.hour},
				&runtime.IntegerValue{Value: tt.minute},
				&runtime.IntegerValue{Value: tt.second},
				&runtime.IntegerValue{Value: tt.msec},
			})

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			floatVal, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}

			// Verify by decoding back
			dt := delphiDateTimeToGoTime(floatVal.Value)
			if dt.Year() != int(tt.year) || int(dt.Month()) != int(tt.month) || dt.Day() != int(tt.day) ||
				dt.Hour() != int(tt.hour) || dt.Minute() != int(tt.minute) || dt.Second() != int(tt.second) {
				t.Errorf("EncodeDateTime decoded to %v", dt)
			}
		})
	}
}

// =============================================================================
// Incrementing Functions Tests
// =============================================================================

func TestIncYear(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name         string
		years        int64
		expectedYear int
	}{
		{
			name:         "add 1 year",
			years:        1,
			expectedYear: 2024,
		},
		{
			name:         "add 10 years",
			years:        10,
			expectedYear: 2033,
		},
		{
			name:         "subtract 1 year",
			years:        -1,
			expectedYear: 2022,
		},
		{
			name:         "add 0 years",
			years:        0,
			expectedYear: 2023,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IncYear(ctx, []Value{
				&runtime.FloatValue{Value: testDate},
				&runtime.IntegerValue{Value: tt.years},
			})

			if result.Type() == "ERROR" {
				t.Errorf("IncYear() returned error: %v", result)
				return
			}

			floatVal, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}

			dt := delphiDateTimeToGoTime(floatVal.Value)
			if dt.Year() != tt.expectedYear {
				t.Errorf("IncYear() year = %d, want %d", dt.Year(), tt.expectedYear)
			}
		})
	}
}

func TestIncMonth(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name          string
		months        int64
		expectedMonth int
		expectedYear  int
	}{
		{
			name:          "add 1 month",
			months:        1,
			expectedMonth: 4,
			expectedYear:  2023,
		},
		{
			name:          "add 12 months",
			months:        12,
			expectedMonth: 3,
			expectedYear:  2024,
		},
		{
			name:          "subtract 1 month",
			months:        -1,
			expectedMonth: 2,
			expectedYear:  2023,
		},
		{
			name:          "add 0 months",
			months:        0,
			expectedMonth: 3,
			expectedYear:  2023,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IncMonth(ctx, []Value{
				&runtime.FloatValue{Value: testDate},
				&runtime.IntegerValue{Value: tt.months},
			})

			if result.Type() == "ERROR" {
				t.Errorf("IncMonth() returned error: %v", result)
				return
			}

			floatVal, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}

			dt := delphiDateTimeToGoTime(floatVal.Value)
			if int(dt.Month()) != tt.expectedMonth || dt.Year() != tt.expectedYear {
				t.Errorf("IncMonth() = %v, want month=%d year=%d", dt, tt.expectedMonth, tt.expectedYear)
			}
		})
	}
}

func TestIncDay(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name        string
		days        int64
		expectedDay int
	}{
		{
			name:        "add 1 day",
			days:        1,
			expectedDay: 16,
		},
		{
			name:        "add 7 days",
			days:        7,
			expectedDay: 22,
		},
		{
			name:        "subtract 1 day",
			days:        -1,
			expectedDay: 14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IncDay(ctx, []Value{
				&runtime.FloatValue{Value: testDate},
				&runtime.IntegerValue{Value: tt.days},
			})

			if result.Type() == "ERROR" {
				t.Errorf("IncDay() returned error: %v", result)
				return
			}

			floatVal, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}

			dt := delphiDateTimeToGoTime(floatVal.Value)
			if dt.Day() != tt.expectedDay {
				t.Errorf("IncDay() day = %d, want %d", dt.Day(), tt.expectedDay)
			}
		})
	}
}

func TestIncHour(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15 12:00:00
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 0, 0, 0, time.UTC))

	tests := []struct {
		name         string
		hours        int64
		expectedHour int
	}{
		{
			name:         "add 1 hour",
			hours:        1,
			expectedHour: 13,
		},
		{
			name:         "add 12 hours",
			hours:        12,
			expectedHour: 0, // wraps to next day
		},
		{
			name:         "subtract 1 hour",
			hours:        -1,
			expectedHour: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IncHour(ctx, []Value{
				&runtime.FloatValue{Value: testDate},
				&runtime.IntegerValue{Value: tt.hours},
			})

			if result.Type() == "ERROR" {
				t.Errorf("IncHour() returned error: %v", result)
				return
			}

			floatVal, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}

			dt := delphiDateTimeToGoTime(floatVal.Value)
			if dt.Hour() != tt.expectedHour {
				t.Errorf("IncHour() hour = %d, want %d", dt.Hour(), tt.expectedHour)
			}
		})
	}
}

func TestIncMinute(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15 12:30:00
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 30, 0, 0, time.UTC))

	tests := []struct {
		name           string
		minutes        int64
		expectedMinute int
	}{
		{
			name:           "add 1 minute",
			minutes:        1,
			expectedMinute: 31,
		},
		{
			name:           "add 30 minutes",
			minutes:        30,
			expectedMinute: 0, // wraps to next hour
		},
		{
			name:           "subtract 1 minute",
			minutes:        -1,
			expectedMinute: 29,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IncMinute(ctx, []Value{
				&runtime.FloatValue{Value: testDate},
				&runtime.IntegerValue{Value: tt.minutes},
			})

			if result.Type() == "ERROR" {
				t.Errorf("IncMinute() returned error: %v", result)
				return
			}

			floatVal, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}

			dt := delphiDateTimeToGoTime(floatVal.Value)
			if dt.Minute() != tt.expectedMinute {
				t.Errorf("IncMinute() minute = %d, want %d", dt.Minute(), tt.expectedMinute)
			}
		})
	}
}

func TestIncSecond(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15 12:30:45
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 30, 45, 0, time.UTC))

	tests := []struct {
		name           string
		seconds        int64
		expectedSecond int
	}{
		{
			name:           "add 1 second",
			seconds:        1,
			expectedSecond: 46,
		},
		{
			name:           "add 15 seconds",
			seconds:        15,
			expectedSecond: 0, // wraps to next minute
		},
		{
			name:           "subtract 1 second",
			seconds:        -1,
			expectedSecond: 44,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IncSecond(ctx, []Value{
				&runtime.FloatValue{Value: testDate},
				&runtime.IntegerValue{Value: tt.seconds},
			})

			if result.Type() == "ERROR" {
				t.Errorf("IncSecond() returned error: %v", result)
				return
			}

			floatVal, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}

			dt := delphiDateTimeToGoTime(floatVal.Value)
			if dt.Second() != tt.expectedSecond {
				t.Errorf("IncSecond() second = %d, want %d", dt.Second(), tt.expectedSecond)
			}
		})
	}
}

// =============================================================================
// Date Difference Functions Tests
// =============================================================================

func TestDaysBetween(t *testing.T) {
	ctx := newMockContext()

	// Create test dates
	date1 := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC))
	date2 := goTimeToDelphiDateTime(time.Date(2023, 3, 20, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		dt1      float64
		dt2      float64
		expected int64
	}{
		{
			name:     "5 days apart",
			dt1:      date1,
			dt2:      date2,
			expected: 5,
		},
		{
			name:     "same date",
			dt1:      date1,
			dt2:      date1,
			expected: 0,
		},
		{
			name:     "reversed order",
			dt1:      date2,
			dt2:      date1,
			expected: 5, // absolute value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DaysBetween(ctx, []Value{
				&runtime.FloatValue{Value: tt.dt1},
				&runtime.FloatValue{Value: tt.dt2},
			})

			if result.Type() == "ERROR" {
				t.Errorf("DaysBetween() returned error: %v", result)
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("DaysBetween() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestHoursBetween(t *testing.T) {
	ctx := newMockContext()

	// Create test datetimes
	dt1 := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 10, 0, 0, 0, time.UTC))
	dt2 := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 15, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		dt1      float64
		dt2      float64
		expected int64
	}{
		{
			name:     "5 hours apart",
			dt1:      dt1,
			dt2:      dt2,
			expected: 5,
		},
		{
			name:     "same time",
			dt1:      dt1,
			dt2:      dt1,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HoursBetween(ctx, []Value{
				&runtime.FloatValue{Value: tt.dt1},
				&runtime.FloatValue{Value: tt.dt2},
			})

			if result.Type() == "ERROR" {
				t.Errorf("HoursBetween() returned error: %v", result)
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("HoursBetween() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestMinutesBetween(t *testing.T) {
	ctx := newMockContext()

	// Create test datetimes
	dt1 := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 10, 0, 0, 0, time.UTC))
	dt2 := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 10, 30, 0, 0, time.UTC))

	tests := []struct {
		name     string
		dt1      float64
		dt2      float64
		expected int64
	}{
		{
			name:     "30 minutes apart",
			dt1:      dt1,
			dt2:      dt2,
			expected: 30,
		},
		{
			name:     "same time",
			dt1:      dt1,
			dt2:      dt1,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MinutesBetween(ctx, []Value{
				&runtime.FloatValue{Value: tt.dt1},
				&runtime.FloatValue{Value: tt.dt2},
			})

			if result.Type() == "ERROR" {
				t.Errorf("MinutesBetween() returned error: %v", result)
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("MinutesBetween() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestSecondsBetween(t *testing.T) {
	ctx := newMockContext()

	// Create test datetimes
	dt1 := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 10, 0, 0, 0, time.UTC))
	dt2 := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 10, 1, 30, 0, time.UTC))

	tests := []struct {
		name     string
		dt1      float64
		dt2      float64
		expected int64
	}{
		{
			name:     "90 seconds apart",
			dt1:      dt1,
			dt2:      dt2,
			expected: 90,
		},
		{
			name:     "same time",
			dt1:      dt1,
			dt2:      dt1,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SecondsBetween(ctx, []Value{
				&runtime.FloatValue{Value: tt.dt1},
				&runtime.FloatValue{Value: tt.dt2},
			})

			if result.Type() == "ERROR" {
				t.Errorf("SecondsBetween() returned error: %v", result)
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			// Allow small tolerance for floating point precision
			if math.Abs(float64(intVal.Value-tt.expected)) > 1 {
				t.Errorf("SecondsBetween() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}
