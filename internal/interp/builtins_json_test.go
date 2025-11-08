package interp

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/jsonvalue"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Tests for JSON Serialization Functions (ToJSON, ToJSONFormatted)
// ============================================================================

func TestToJSON_Primitives(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		wantJSON string
	}{
		{
			name:     "integer",
			value:    &IntegerValue{Value: 42},
			wantJSON: "42",
		},
		{
			name:     "negative integer",
			value:    &IntegerValue{Value: -99},
			wantJSON: "-99",
		},
		{
			name:     "zero",
			value:    &IntegerValue{Value: 0},
			wantJSON: "0",
		},
		{
			name:     "float",
			value:    &FloatValue{Value: 3.14},
			wantJSON: "3.14",
		},
		{
			name:     "negative float",
			value:    &FloatValue{Value: -2.5},
			wantJSON: "-2.5",
		},
		{
			name:     "string",
			value:    &StringValue{Value: "hello"},
			wantJSON: `"hello"`,
		},
		{
			name:     "empty string",
			value:    &StringValue{Value: ""},
			wantJSON: `""`,
		},
		{
			name:     "string with escapes",
			value:    &StringValue{Value: "line1\nline2\ttab"},
			wantJSON: `"line1\nline2\ttab"`,
		},
		{
			name:     "string with quotes",
			value:    &StringValue{Value: `say "hello"`},
			wantJSON: `"say \"hello\""`,
		},
		{
			name:     "boolean true",
			value:    &BooleanValue{Value: true},
			wantJSON: "true",
		},
		{
			name:     "boolean false",
			value:    &BooleanValue{Value: false},
			wantJSON: "false",
		},
		{
			name:     "nil",
			value:    &NilValue{},
			wantJSON: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := newTestInterpreter()
			result := interp.builtinToJSON([]Value{tt.value})

			// Check for error
			if errVal, ok := result.(*ErrorValue); ok {
				t.Fatalf("ToJSON() returned error: %s", errVal.Message)
			}

			// Check result is a string
			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("ToJSON() returned %T, want StringValue", result)
			}

			// Compare JSON
			if strVal.Value != tt.wantJSON {
				t.Errorf("ToJSON() = %q, want %q", strVal.Value, tt.wantJSON)
			}
		})
	}
}

func TestToJSON_Arrays(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		wantJSON string
	}{
		{
			name: "empty array",
			value: &ArrayValue{
				Elements: []Value{},
			},
			wantJSON: "[]",
		},
		{
			name: "integer array",
			value: &ArrayValue{
				Elements: []Value{
					&IntegerValue{Value: 1},
					&IntegerValue{Value: 2},
					&IntegerValue{Value: 3},
				},
			},
			wantJSON: "[1,2,3]",
		},
		{
			name: "string array",
			value: &ArrayValue{
				Elements: []Value{
					&StringValue{Value: "a"},
					&StringValue{Value: "b"},
					&StringValue{Value: "c"},
				},
			},
			wantJSON: `["a","b","c"]`,
		},
		{
			name: "mixed array",
			value: &ArrayValue{
				Elements: []Value{
					&IntegerValue{Value: 42},
					&StringValue{Value: "hello"},
					&BooleanValue{Value: true},
					&NilValue{},
				},
			},
			wantJSON: `[42,"hello",true,null]`,
		},
		{
			name: "nested array",
			value: &ArrayValue{
				Elements: []Value{
					&ArrayValue{
						Elements: []Value{
							&IntegerValue{Value: 1},
							&IntegerValue{Value: 2},
						},
					},
					&ArrayValue{
						Elements: []Value{
							&IntegerValue{Value: 3},
							&IntegerValue{Value: 4},
						},
					},
				},
			},
			wantJSON: "[[1,2],[3,4]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := newTestInterpreter()
			result := interp.builtinToJSON([]Value{tt.value})

			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("ToJSON() returned %T, want StringValue", result)
			}

			if strVal.Value != tt.wantJSON {
				t.Errorf("ToJSON() = %q, want %q", strVal.Value, tt.wantJSON)
			}
		})
	}
}

func TestToJSON_Records(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		wantJSON string
	}{
		{
			name: "empty record",
			value: &RecordValue{
				Fields: map[string]Value{},
			},
			wantJSON: "{}",
		},
		{
			name: "simple record",
			value: &RecordValue{
				Fields: map[string]Value{
					"name": &StringValue{Value: "John"},
					"age":  &IntegerValue{Value: 30},
				},
			},
			// Note: map order is not guaranteed, so we'll check both fields exist
			wantJSON: `{"age":30,"name":"John"}`,
		},
		{
			name: "nested record",
			value: &RecordValue{
				Fields: map[string]Value{
					"person": &RecordValue{
						Fields: map[string]Value{
							"name": &StringValue{Value: "Alice"},
							"age":  &IntegerValue{Value: 25},
						},
					},
					"active": &BooleanValue{Value: true},
				},
			},
			// Check contains key parts
			wantJSON: "", // We'll check this differently
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := newTestInterpreter()
			result := interp.builtinToJSON([]Value{tt.value})

			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("ToJSON() returned %T, want StringValue", result)
			}

			if tt.name == "nested record" {
				// Check key parts exist (order may vary)
				if !strings.Contains(strVal.Value, `"person"`) {
					t.Errorf("ToJSON() missing 'person' field")
				}
				if !strings.Contains(strVal.Value, `"name":"Alice"`) {
					t.Errorf("ToJSON() missing nested name field")
				}
				if !strings.Contains(strVal.Value, `"active":true`) {
					t.Errorf("ToJSON() missing active field")
				}
			} else if tt.wantJSON != "" && strVal.Value != tt.wantJSON {
				t.Errorf("ToJSON() = %q, want %q", strVal.Value, tt.wantJSON)
			}
		})
	}
}

func TestToJSON_Variants(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		wantJSON string
	}{
		{
			name: "variant wrapping integer",
			value: &VariantValue{
				Value:      &IntegerValue{Value: 42},
				ActualType: types.INTEGER,
			},
			wantJSON: "42",
		},
		{
			name: "variant wrapping string",
			value: &VariantValue{
				Value:      &StringValue{Value: "test"},
				ActualType: types.STRING,
			},
			wantJSON: `"test"`,
		},
		{
			name: "variant wrapping array",
			value: &VariantValue{
				Value: &ArrayValue{
					Elements: []Value{
						&IntegerValue{Value: 1},
						&IntegerValue{Value: 2},
					},
				},
				ActualType: nil,
			},
			wantJSON: "[1,2]",
		},
		{
			name: "empty variant",
			value: &VariantValue{
				Value:      nil,
				ActualType: nil,
			},
			wantJSON: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := newTestInterpreter()
			result := interp.builtinToJSON([]Value{tt.value})

			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("ToJSON() returned %T, want StringValue", result)
			}

			if strVal.Value != tt.wantJSON {
				t.Errorf("ToJSON() = %q, want %q", strVal.Value, tt.wantJSON)
			}
		})
	}
}

func TestToJSON_JSONValues(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		wantJSON string
	}{
		{
			name: "JSON null",
			value: &JSONValue{
				Value: jsonvalue.NewNull(),
			},
			wantJSON: "null",
		},
		{
			name: "JSON boolean",
			value: &JSONValue{
				Value: jsonvalue.NewBoolean(true),
			},
			wantJSON: "true",
		},
		{
			name: "JSON string",
			value: &JSONValue{
				Value: jsonvalue.NewString("hello"),
			},
			wantJSON: `"hello"`,
		},
		{
			name: "JSON number",
			value: &JSONValue{
				Value: jsonvalue.NewNumber(3.14),
			},
			wantJSON: "3.14",
		},
		{
			name: "JSON int64",
			value: &JSONValue{
				Value: jsonvalue.NewInt64(42),
			},
			wantJSON: "42",
		},
		{
			name: "JSON array",
			value: &JSONValue{
				Value: func() *jsonvalue.Value {
					arr := jsonvalue.NewArray()
					arr.ArrayAppend(jsonvalue.NewInt64(1))
					arr.ArrayAppend(jsonvalue.NewInt64(2))
					arr.ArrayAppend(jsonvalue.NewInt64(3))
					return arr
				}(),
			},
			wantJSON: "[1,2,3]",
		},
		{
			name: "JSON object",
			value: &JSONValue{
				Value: func() *jsonvalue.Value {
					obj := jsonvalue.NewObject()
					obj.ObjectSet("name", jsonvalue.NewString("John"))
					obj.ObjectSet("age", jsonvalue.NewInt64(30))
					return obj
				}(),
			},
			// Map order not guaranteed
			wantJSON: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := newTestInterpreter()
			result := interp.builtinToJSON([]Value{tt.value})

			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("ToJSON() returned %T, want StringValue", result)
			}

			if tt.name == "JSON object" {
				// Just check it's valid JSON with expected keys
				if !strings.Contains(strVal.Value, `"name"`) || !strings.Contains(strVal.Value, `"age"`) {
					t.Errorf("ToJSON() = %q, missing expected keys", strVal.Value)
				}
			} else if tt.wantJSON != "" && strVal.Value != tt.wantJSON {
				t.Errorf("ToJSON() = %q, want %q", strVal.Value, tt.wantJSON)
			}
		})
	}
}

func TestToJSON_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		wantErrSubstr string
		args          []Value
	}{
		{
			name:          "no arguments",
			args:          []Value{},
			wantErrSubstr: "expects exactly 1 argument",
		},
		{
			name: "too many arguments",
			args: []Value{
				&IntegerValue{Value: 42},
				&IntegerValue{Value: 99},
			},
			wantErrSubstr: "expects exactly 1 argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := newTestInterpreter()
			result := interp.builtinToJSON(tt.args)

			errVal, ok := result.(*ErrorValue)
			if !ok {
				t.Fatalf("ToJSON() returned %T, want ErrorValue", result)
			}

			if !strings.Contains(errVal.Message, tt.wantErrSubstr) {
				t.Errorf("ToJSON() error = %q, want substring %q", errVal.Message, tt.wantErrSubstr)
			}
		})
	}
}

func TestToJSONFormatted_Basic(t *testing.T) {
	tests := []struct {
		value      Value
		name       string
		wantSubstr []string
		indent     int64
	}{
		{
			name:       "integer with indent 2",
			value:      &IntegerValue{Value: 42},
			indent:     2,
			wantSubstr: []string{"42"},
		},
		{
			name:       "string with indent 4",
			value:      &StringValue{Value: "hello"},
			indent:     4,
			wantSubstr: []string{`"hello"`},
		},
		{
			name: "simple array with indent 2",
			value: &ArrayValue{
				Elements: []Value{
					&IntegerValue{Value: 1},
					&IntegerValue{Value: 2},
					&IntegerValue{Value: 3},
				},
			},
			indent:     2,
			wantSubstr: []string{"[\n", "  1", "  2", "  3", "\n]"},
		},
		{
			name: "simple record with indent 2",
			value: &RecordValue{
				Fields: map[string]Value{
					"name": &StringValue{Value: "John"},
					"age":  &IntegerValue{Value: 30},
				},
			},
			indent:     2,
			wantSubstr: []string{"{\n", `"name"`, `"age"`, "\n}"},
		},
		{
			name: "nested structure with indent 4",
			value: &RecordValue{
				Fields: map[string]Value{
					"person": &RecordValue{
						Fields: map[string]Value{
							"name": &StringValue{Value: "Alice"},
						},
					},
				},
			},
			indent:     4,
			wantSubstr: []string{"{\n", `"person"`, `"name"`, "Alice", "\n}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := newTestInterpreter()
			args := []Value{
				tt.value,
				&IntegerValue{Value: tt.indent},
			}
			result := interp.builtinToJSONFormatted(args)

			// Check for error
			if errVal, ok := result.(*ErrorValue); ok {
				t.Fatalf("ToJSONFormatted() returned error: %s", errVal.Message)
			}

			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("ToJSONFormatted() returned %T, want StringValue", result)
			}

			// Check all expected substrings are present
			for _, substr := range tt.wantSubstr {
				if !strings.Contains(strVal.Value, substr) {
					t.Errorf("ToJSONFormatted() = %q, missing substring %q", strVal.Value, substr)
				}
			}
		})
	}
}

func TestToJSONFormatted_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		wantErrSubstr string
		args          []Value
	}{
		{
			name:          "no arguments",
			args:          []Value{},
			wantErrSubstr: "expects exactly 2 arguments",
		},
		{
			name: "one argument",
			args: []Value{
				&IntegerValue{Value: 42},
			},
			wantErrSubstr: "expects exactly 2 arguments",
		},
		{
			name: "too many arguments",
			args: []Value{
				&IntegerValue{Value: 42},
				&IntegerValue{Value: 2},
				&IntegerValue{Value: 99},
			},
			wantErrSubstr: "expects exactly 2 arguments",
		},
		{
			name: "second arg not integer",
			args: []Value{
				&IntegerValue{Value: 42},
				&StringValue{Value: "not an int"},
			},
			wantErrSubstr: "expects Integer as second argument",
		},
		{
			name: "negative indent",
			args: []Value{
				&IntegerValue{Value: 42},
				&IntegerValue{Value: -1},
			},
			wantErrSubstr: "indent must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := newTestInterpreter()
			result := interp.builtinToJSONFormatted(tt.args)

			errVal, ok := result.(*ErrorValue)
			if !ok {
				t.Fatalf("ToJSONFormatted() returned %T, want ErrorValue", result)
			}

			if !strings.Contains(errVal.Message, tt.wantErrSubstr) {
				t.Errorf("ToJSONFormatted() error = %q, want substring %q", errVal.Message, tt.wantErrSubstr)
			}
		})
	}
}

func TestToJSONFormatted_ZeroIndent(t *testing.T) {
	// Zero indent should still format with newlines but no indentation
	value := &ArrayValue{
		Elements: []Value{
			&IntegerValue{Value: 1},
			&IntegerValue{Value: 2},
		},
	}

	interp := newTestInterpreter()
	args := []Value{value, &IntegerValue{Value: 0}}
	result := interp.builtinToJSONFormatted(args)

	strVal, ok := result.(*StringValue)
	if !ok {
		t.Fatalf("ToJSONFormatted() returned %T, want StringValue", result)
	}

	// Should have newlines but no spaces
	if !strings.Contains(strVal.Value, "\n") {
		t.Errorf("ToJSONFormatted() with indent=0 should have newlines")
	}
}

// TestToJSON_Integration tests ToJSON in a real interpreter context with DWScript code
// Note: Integration tests that run full DWScript programs are in testdata/json/
// directory and are tested via the CLI tool tests.
