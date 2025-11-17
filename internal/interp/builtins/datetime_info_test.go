package builtins

import (
	"testing"
	"time"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Current Date/Time Functions Tests
// =============================================================================

func TestNow(t *testing.T) {
	ctx := newMockContext()

	// Test normal call
	result := Now(ctx, []Value{})
	if result.Type() != "FLOAT" {
		t.Errorf("Now() should return FLOAT, got %s", result.Type())
	}

	// Verify it returns a reasonable TDateTime value (should be positive)
	floatVal, ok := result.(*runtime.FloatValue)
	if !ok {
		t.Fatalf("Now() should return FloatValue, got %T", result)
	}
	if floatVal.Value <= 0 {
		t.Errorf("Now() should return positive TDateTime value, got %f", floatVal.Value)
	}

	// Test error: too many arguments
	result = Now(ctx, []Value{&runtime.IntegerValue{Value: 1}})
	if result.Type() != "ERROR" {
		t.Errorf("Now() with arguments should return ERROR, got %s", result.Type())
	}
}

func TestDate(t *testing.T) {
	ctx := newMockContext()

	// Test normal call
	result := Date(ctx, []Value{})
	if result.Type() != "FLOAT" {
		t.Errorf("Date() should return FLOAT, got %s", result.Type())
	}

	// Verify it returns a value (time part should be ~0.0)
	floatVal, ok := result.(*runtime.FloatValue)
	if !ok {
		t.Fatalf("Date() should return FloatValue, got %T", result)
	}

	// The fractional part should be very small (representing midnight)
	fractionalPart := floatVal.Value - float64(int(floatVal.Value))
	if fractionalPart > 0.001 {
		t.Errorf("Date() should have minimal time component, got fractional part %f", fractionalPart)
	}

	// Test error: too many arguments
	result = Date(ctx, []Value{&runtime.IntegerValue{Value: 1}})
	if result.Type() != "ERROR" {
		t.Errorf("Date() with arguments should return ERROR, got %s", result.Type())
	}
}

func TestTime(t *testing.T) {
	ctx := newMockContext()

	// Test normal call
	result := Time(ctx, []Value{})
	if result.Type() != "FLOAT" {
		t.Errorf("Time() should return FLOAT, got %s", result.Type())
	}

	// Verify it returns a positive value
	floatVal, ok := result.(*runtime.FloatValue)
	if !ok {
		t.Fatalf("Time() should return FloatValue, got %T", result)
	}

	// Time component should be >= 0 and < 1 (representing time within a day)
	if floatVal.Value < 0 || floatVal.Value >= 1.0 {
		t.Errorf("Time() should return value in [0, 1), got %f", floatVal.Value)
	}

	// Test error: too many arguments
	result = Time(ctx, []Value{&runtime.IntegerValue{Value: 1}})
	if result.Type() != "ERROR" {
		t.Errorf("Time() with arguments should return ERROR, got %s", result.Type())
	}
}

func TestUTCDateTime(t *testing.T) {
	ctx := newMockContext()

	// Test normal call
	result := UTCDateTime(ctx, []Value{})
	if result.Type() != "FLOAT" {
		t.Errorf("UTCDateTime() should return FLOAT, got %s", result.Type())
	}

	// Verify it returns a reasonable TDateTime value
	floatVal, ok := result.(*runtime.FloatValue)
	if !ok {
		t.Fatalf("UTCDateTime() should return FloatValue, got %T", result)
	}
	if floatVal.Value <= 0 {
		t.Errorf("UTCDateTime() should return positive TDateTime value, got %f", floatVal.Value)
	}

	// Test error: too many arguments
	result = UTCDateTime(ctx, []Value{&runtime.IntegerValue{Value: 1}})
	if result.Type() != "ERROR" {
		t.Errorf("UTCDateTime() with arguments should return ERROR, got %s", result.Type())
	}
}

// =============================================================================
// Component Extraction Functions Tests
// =============================================================================

func TestYearOf(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15 12:30:45
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 30, 45, 0, time.UTC))

	tests := []struct {
		name     string
		args     []Value
		expected int64
		isError  bool
	}{
		{
			name:     "valid date",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: 2023,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.StringValue{Value: "hello"}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := YearOf(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("YearOf() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestMonthOf(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15 12:30:45
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 30, 45, 0, time.UTC))

	tests := []struct {
		name     string
		args     []Value
		expected int64
		isError  bool
	}{
		{
			name:     "valid date - March",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: 3,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.IntegerValue{Value: 123}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MonthOf(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("MonthOf() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestDayOf(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15 12:30:45
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 30, 45, 0, time.UTC))

	tests := []struct {
		name     string
		args     []Value
		expected int64
		isError  bool
	}{
		{
			name:     "valid date - 15th",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: 15,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.BooleanValue{Value: true}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DayOf(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("DayOf() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestHourOf(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15 12:30:45
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 30, 45, 0, time.UTC))

	tests := []struct {
		name     string
		args     []Value
		expected int64
		isError  bool
	}{
		{
			name:     "valid time - 12 hours",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: 12,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.StringValue{Value: "12:30"}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HourOf(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("HourOf() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestMinuteOf(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15 12:30:45
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 30, 45, 0, time.UTC))

	tests := []struct {
		name     string
		args     []Value
		expected int64
		isError  bool
	}{
		{
			name:     "valid time - 30 minutes",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: 30,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.IntegerValue{Value: 30}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MinuteOf(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("MinuteOf() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestSecondOf(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15 12:30:45
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 30, 45, 0, time.UTC))

	tests := []struct {
		name     string
		args     []Value
		expected int64
		isError  bool
	}{
		{
			name:     "valid time - 45 seconds",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: 45,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.BooleanValue{Value: false}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SecondOf(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("SecondOf() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestDayOfWeek(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15 (Wednesday)
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		args     []Value
		expected int64
		isError  bool
	}{
		{
			name:     "Wednesday - should be 4",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: 4, // Delphi: 1=Sunday, so Wednesday=4
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.StringValue{Value: "Wednesday"}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DayOfWeek(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("DayOfWeek() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestDayOfTheWeek(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15 (Wednesday)
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		args     []Value
		expected int64
		isError  bool
	}{
		{
			name:     "Wednesday - ISO 8601 should be 3",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: 3, // ISO 8601: 1=Monday, so Wednesday=3
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.IntegerValue{Value: 3}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DayOfTheWeek(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("DayOfTheWeek() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestDayOfYear(t *testing.T) {
	ctx := newMockContext()

	// Create test dates
	jan1 := goTimeToDelphiDateTime(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
	mar15 := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		args     []Value
		expected int64
		isError  bool
	}{
		{
			name:     "January 1 - should be 1",
			args:     []Value{&runtime.FloatValue{Value: jan1}},
			expected: 1,
		},
		{
			name:     "March 15 - should be 74",
			args:     []Value{&runtime.FloatValue{Value: mar15}},
			expected: 74, // 31 (Jan) + 28 (Feb) + 15 (Mar)
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.StringValue{Value: "day 74"}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DayOfYear(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("DayOfYear() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestWeekNumber(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		args     []Value
		expected int64
		isError  bool
	}{
		{
			name:     "March 15, 2023",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: 11, // ISO week 11
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.IntegerValue{Value: 11}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WeekNumber(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("WeekNumber() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestYearOfWeek(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		args     []Value
		expected int64
		isError  bool
	}{
		{
			name:     "March 15, 2023 - year should be 2023",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: 2023,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.BooleanValue{Value: true}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := YearOfWeek(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("YearOfWeek() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// =============================================================================
// Special Date Functions Tests
// =============================================================================

func TestIsLeapYear(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected bool
		isError  bool
	}{
		{
			name:     "2020 is leap year",
			args:     []Value{&runtime.IntegerValue{Value: 2020}},
			expected: true,
		},
		{
			name:     "2023 is not leap year",
			args:     []Value{&runtime.IntegerValue{Value: 2023}},
			expected: false,
		},
		{
			name:     "2000 is leap year (divisible by 400)",
			args:     []Value{&runtime.IntegerValue{Value: 2000}},
			expected: true,
		},
		{
			name:     "1900 is not leap year (divisible by 100 but not 400)",
			args:     []Value{&runtime.IntegerValue{Value: 1900}},
			expected: false,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.FloatValue{Value: 2020.0}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLeapYear(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("IsLeapYear() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}
}

func TestFirstDayOfYear(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 30, 45, 0, time.UTC))
	expectedDate := goTimeToDelphiDateTime(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		args     []Value
		expected float64
		isError  bool
	}{
		{
			name:     "March 15 -> January 1",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: expectedDate,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.IntegerValue{Value: 2023}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FirstDayOfYear(ctx, tt.args)

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
			if floatVal.Value != tt.expected {
				t.Errorf("FirstDayOfYear() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

func TestFirstDayOfNextYear(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 30, 45, 0, time.UTC))
	expectedDate := goTimeToDelphiDateTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		args     []Value
		expected float64
		isError  bool
	}{
		{
			name:     "2023-03-15 -> 2024-01-01",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: expectedDate,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.StringValue{Value: "2024"}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FirstDayOfNextYear(ctx, tt.args)

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
			if floatVal.Value != tt.expected {
				t.Errorf("FirstDayOfNextYear() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

func TestFirstDayOfMonth(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 30, 45, 0, time.UTC))
	expectedDate := goTimeToDelphiDateTime(time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		args     []Value
		expected float64
		isError  bool
	}{
		{
			name:     "March 15 -> March 1",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: expectedDate,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.BooleanValue{Value: false}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FirstDayOfMonth(ctx, tt.args)

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
			if floatVal.Value != tt.expected {
				t.Errorf("FirstDayOfMonth() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

func TestFirstDayOfNextMonth(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 30, 45, 0, time.UTC))
	expectedDate := goTimeToDelphiDateTime(time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		args     []Value
		expected float64
		isError  bool
	}{
		{
			name:     "March 15 -> April 1",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: expectedDate,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.IntegerValue{Value: 4}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FirstDayOfNextMonth(ctx, tt.args)

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
			if floatVal.Value != tt.expected {
				t.Errorf("FirstDayOfNextMonth() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

func TestFirstDayOfWeek(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15 (Wednesday)
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 30, 45, 0, time.UTC))
	expectedDate := goTimeToDelphiDateTime(time.Date(2023, 3, 13, 0, 0, 0, 0, time.UTC)) // Monday

	tests := []struct {
		name     string
		args     []Value
		expected float64
		isError  bool
	}{
		{
			name:     "Wednesday March 15 -> Monday March 13",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: expectedDate,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.StringValue{Value: "Monday"}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FirstDayOfWeek(ctx, tt.args)

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
			if floatVal.Value != tt.expected {
				t.Errorf("FirstDayOfWeek() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}
