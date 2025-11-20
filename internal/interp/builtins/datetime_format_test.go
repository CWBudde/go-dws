package builtins

import (
	"testing"
	"time"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Formatting Functions Tests
// =============================================================================

func TestFormatDateTime(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15 12:30:45.123
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 30, 45, 123000000, time.UTC))

	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:     "full datetime",
			format:   "yyyy-mm-dd hh:nn:ss",
			expected: "2023-03-15 12:30:45",
		},
		{
			name:     "date only",
			format:   "yyyy-mm-dd",
			expected: "2023-03-15",
		},
		{
			name:     "time only",
			format:   "hh:nn:ss",
			expected: "12:30:45",
		},
		{
			name:     "with milliseconds",
			format:   "yyyy-mm-dd hh:nn:ss.zzz",
			expected: "2023-03-15 12:30:45.123",
		},
		{
			name:     "short year",
			format:   "yy-m-d",
			expected: "23-3-15",
		},
		{
			name:     "no leading zeros",
			format:   "d/m/yyyy h:n:s",
			expected: "15/3/2023 12:30:45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDateTime(ctx, []Value{
				&runtime.StringValue{Value: tt.format},
				&runtime.FloatValue{Value: testDate},
			})

			if result.Type() == "ERROR" {
				t.Errorf("FormatDateTime() returned error: %v", result)
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("FormatDateTime(%q) = %q, want %q", tt.format, strVal.Value, tt.expected)
			}
		})
	}

	// Test errors
	errorTests := []struct {
		name string
		args []Value
	}{
		{
			name: "wrong argument count - 0 args",
			args: []Value{},
		},
		{
			name: "wrong argument count - 1 arg",
			args: []Value{&runtime.StringValue{Value: "yyyy-mm-dd"}},
		},
		{
			name: "wrong argument count - 3 args",
			args: []Value{
				&runtime.StringValue{Value: "yyyy-mm-dd"},
				&runtime.FloatValue{Value: testDate},
				&runtime.IntegerValue{Value: 1},
			},
		},
		{
			name: "wrong first argument type",
			args: []Value{
				&runtime.IntegerValue{Value: 123},
				&runtime.FloatValue{Value: testDate},
			},
		},
		{
			name: "wrong second argument type",
			args: []Value{
				&runtime.StringValue{Value: "yyyy-mm-dd"},
				&runtime.StringValue{Value: "2023-03-15"},
			},
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDateTime(ctx, tt.args)
			if result.Type() != "ERROR" {
				t.Errorf("expected error, got %v", result)
			}
		})
	}
}

func TestDateTimeToStr(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15 12:30:45
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 30, 45, 0, time.UTC))

	tests := []struct {
		name     string
		expected string
		args     []Value
		isError  bool
	}{
		{
			name:     "valid datetime",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: "2023-03-15 12:30:45",
		},
		{
			name:    "wrong argument count - 0 args",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong argument count - 2 args",
			args:    []Value{&runtime.FloatValue{Value: testDate}, &runtime.IntegerValue{Value: 1}},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.StringValue{Value: "2023-03-15"}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DateTimeToStr(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("DateTimeToStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestDateToStr(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		expected string
		args     []Value
		isError  bool
	}{
		{
			name:     "valid date",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: "2023-03-15",
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.IntegerValue{Value: 20230315}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DateToStr(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("DateToStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestTimeToStr(t *testing.T) {
	ctx := newMockContext()

	// Create test time: 12:30:45
	testTime := goTimeToDelphiDateTime(time.Date(1899, 12, 30, 12, 30, 45, 0, time.UTC))

	tests := []struct {
		name     string
		expected string
		args     []Value
		isError  bool
	}{
		{
			name:     "valid time",
			args:     []Value{&runtime.FloatValue{Value: testTime}},
			expected: "12:30:45",
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.StringValue{Value: "12:30:45"}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TimeToStr(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("TimeToStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestDateToISO8601(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		expected string
		args     []Value
		isError  bool
	}{
		{
			name:     "valid date",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: "2023-03-15",
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
			result := DateToISO8601(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("DateToISO8601() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestDateTimeToISO8601(t *testing.T) {
	ctx := newMockContext()

	// Create test datetime: 2023-03-15 12:30:45
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 30, 45, 0, time.UTC))

	tests := []struct {
		name     string
		expected string
		args     []Value
		isError  bool
	}{
		{
			name:     "valid datetime",
			args:     []Value{&runtime.FloatValue{Value: testDate}},
			expected: "2023-03-15T12:30:45",
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
			result := DateTimeToISO8601(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("DateTimeToISO8601() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestDateTimeToRFC822(t *testing.T) {
	ctx := newMockContext()

	// Create test datetime: 2023-03-15 12:30:45
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 30, 0, 0, time.UTC))

	tests := []struct {
		name    string
		args    []Value
		isError bool
	}{
		{
			name:    "valid datetime",
			args:    []Value{&runtime.FloatValue{Value: testDate}},
			isError: false,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.StringValue{Value: "Wed, 15 Mar 2023 12:30:00 UTC"}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DateTimeToRFC822(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			// Just verify it returns a non-empty string
			if strVal.Value == "" {
				t.Errorf("DateTimeToRFC822() returned empty string")
			}
		})
	}
}

// =============================================================================
// Parsing Functions Tests
// =============================================================================

func TestStrToDate(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name    string
		args    []Value
		isError bool
	}{
		{
			name:    "ISO format",
			args:    []Value{&runtime.StringValue{Value: "2023-03-15"}},
			isError: false,
		},
		{
			name:    "US format",
			args:    []Value{&runtime.StringValue{Value: "03/15/2023"}},
			isError: false,
		},
		{
			name:    "invalid format",
			args:    []Value{&runtime.StringValue{Value: "not a date"}},
			isError: true,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.IntegerValue{Value: 20230315}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrToDate(ctx, tt.args)

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
			// Just verify it returns a value
			if floatVal.Value == 0 {
				t.Errorf("StrToDate() returned zero value")
			}
		})
	}
}

func TestStrToDateTime(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name    string
		args    []Value
		isError bool
	}{
		{
			name:    "ISO format with time",
			args:    []Value{&runtime.StringValue{Value: "2023-03-15 12:30:45"}},
			isError: false,
		},
		{
			name:    "ISO format with T separator",
			args:    []Value{&runtime.StringValue{Value: "2023-03-15T12:30:45"}},
			isError: false,
		},
		{
			name:    "invalid format",
			args:    []Value{&runtime.StringValue{Value: "not a datetime"}},
			isError: true,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.FloatValue{Value: 123.45}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrToDateTime(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			_, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
		})
	}
}

func TestStrToTime(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name    string
		args    []Value
		isError bool
	}{
		{
			name:    "time with seconds",
			args:    []Value{&runtime.StringValue{Value: "12:30:45"}},
			isError: false,
		},
		{
			name:    "time without seconds",
			args:    []Value{&runtime.StringValue{Value: "12:30"}},
			isError: false,
		},
		{
			name:    "invalid format",
			args:    []Value{&runtime.StringValue{Value: "not a time"}},
			isError: true,
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
			result := StrToTime(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			_, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
		})
	}
}

func TestISO8601ToDateTime(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name    string
		args    []Value
		isError bool
	}{
		{
			name:    "basic ISO 8601",
			args:    []Value{&runtime.StringValue{Value: "2023-03-15"}},
			isError: false,
		},
		{
			name:    "ISO 8601 with time",
			args:    []Value{&runtime.StringValue{Value: "2023-03-15T12:30:45"}},
			isError: false,
		},
		{
			name:    "ISO 8601 with Z suffix",
			args:    []Value{&runtime.StringValue{Value: "2023-03-15T12:30:45Z"}},
			isError: false,
		},
		{
			name:    "invalid format",
			args:    []Value{&runtime.StringValue{Value: "03/15/2023"}},
			isError: true,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.IntegerValue{Value: 20230315}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ISO8601ToDateTime(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			_, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
		})
	}
}

func TestRFC822ToDateTime(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name    string
		args    []Value
		isError bool
	}{
		{
			name:    "valid RFC 822",
			args:    []Value{&runtime.StringValue{Value: "15 Mar 23 12:30 UTC"}},
			isError: false,
		},
		{
			name:    "invalid format",
			args:    []Value{&runtime.StringValue{Value: "2023-03-15"}},
			isError: true,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.FloatValue{Value: 123.45}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RFC822ToDateTime(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			_, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
		})
	}
}

// =============================================================================
// Unix Time Functions Tests
// =============================================================================

func TestUnixTime(t *testing.T) {
	ctx := newMockContext()

	// Test normal call
	result := UnixTime(ctx, []Value{})
	if result.Type() != "INTEGER" {
		t.Errorf("UnixTime() should return INTEGER, got %s", result.Type())
	}

	intVal, ok := result.(*runtime.IntegerValue)
	if !ok {
		t.Fatalf("UnixTime() should return IntegerValue, got %T", result)
	}
	if intVal.Value <= 0 {
		t.Errorf("UnixTime() should return positive value, got %d", intVal.Value)
	}

	// Test error: too many arguments
	result = UnixTime(ctx, []Value{&runtime.IntegerValue{Value: 1}})
	if result.Type() != "ERROR" {
		t.Errorf("UnixTime() with arguments should return ERROR")
	}
}

func TestUnixTimeMSec(t *testing.T) {
	ctx := newMockContext()

	// Test normal call
	result := UnixTimeMSec(ctx, []Value{})
	if result.Type() != "INTEGER" {
		t.Errorf("UnixTimeMSec() should return INTEGER, got %s", result.Type())
	}

	intVal, ok := result.(*runtime.IntegerValue)
	if !ok {
		t.Fatalf("UnixTimeMSec() should return IntegerValue, got %T", result)
	}
	if intVal.Value <= 0 {
		t.Errorf("UnixTimeMSec() should return positive value, got %d", intVal.Value)
	}

	// Test error: too many arguments
	result = UnixTimeMSec(ctx, []Value{&runtime.IntegerValue{Value: 1}})
	if result.Type() != "ERROR" {
		t.Errorf("UnixTimeMSec() with arguments should return ERROR")
	}
}

func TestUnixTimeToDateTime(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name    string
		args    []Value
		isError bool
	}{
		{
			name:    "valid unix time",
			args:    []Value{&runtime.IntegerValue{Value: 1678886400}}, // 2023-03-15 12:00:00 UTC
			isError: false,
		},
		{
			name:    "zero timestamp",
			args:    []Value{&runtime.IntegerValue{Value: 0}},
			isError: false,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.FloatValue{Value: 1678886400.0}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnixTimeToDateTime(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			_, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
		})
	}
}

func TestDateTimeToUnixTime(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15 12:00:00 UTC
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 0, 0, 0, time.UTC))

	tests := []struct {
		name    string
		args    []Value
		isError bool
	}{
		{
			name:    "valid datetime",
			args:    []Value{&runtime.FloatValue{Value: testDate}},
			isError: false,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.IntegerValue{Value: 1678886400}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DateTimeToUnixTime(ctx, tt.args)

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
			// Verify reasonable unix time (after year 2000)
			if intVal.Value < 946684800 {
				t.Errorf("DateTimeToUnixTime() returned unreasonable value: %d", intVal.Value)
			}
		})
	}
}

func TestUnixTimeMSecToDateTime(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name    string
		args    []Value
		isError bool
	}{
		{
			name:    "valid unix time ms",
			args:    []Value{&runtime.IntegerValue{Value: 1678886400000}}, // 2023-03-15 12:00:00 UTC in ms
			isError: false,
		},
		{
			name:    "zero timestamp",
			args:    []Value{&runtime.IntegerValue{Value: 0}},
			isError: false,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.StringValue{Value: "1678886400000"}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnixTimeMSecToDateTime(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			_, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
		})
	}
}

func TestDateTimeToUnixTimeMSec(t *testing.T) {
	ctx := newMockContext()

	// Create test date: 2023-03-15 12:00:00 UTC
	testDate := goTimeToDelphiDateTime(time.Date(2023, 3, 15, 12, 0, 0, 0, time.UTC))

	tests := []struct {
		name    string
		args    []Value
		isError bool
	}{
		{
			name:    "valid datetime",
			args:    []Value{&runtime.FloatValue{Value: testDate}},
			isError: false,
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
			result := DateTimeToUnixTimeMSec(ctx, tt.args)

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
			// Verify reasonable unix time ms (after year 2000)
			if intVal.Value < 946684800000 {
				t.Errorf("DateTimeToUnixTimeMSec() returned unreasonable value: %d", intVal.Value)
			}
		})
	}
}
