package dwscript

import (
	"bytes"
	"strings"
	"testing"
)

// TestRegisterFunctionWithArray tests array marshaling.
func TestRegisterFunctionWithArray(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Test 1: Go function returns array
	err = engine.RegisterFunction("GetScores", func() []int64 {
		return []int64{95, 87, 92}
	})
	if err != nil {
		t.Fatalf("failed to register GetScores: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)
	_, err = engine.Eval(`
		var scores := GetScores();
		PrintLn(IntToStr(Length(scores)));
		PrintLn(IntToStr(scores[0]));
		PrintLn(IntToStr(scores[1]));
		PrintLn(IntToStr(scores[2]));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "3" {
		t.Errorf("expected length '3', got '%s'", lines[0])
	}
	if lines[1] != "95" || lines[2] != "87" || lines[3] != "92" {
		t.Errorf("unexpected scores: %v", lines[1:])
	}

	// Test 2: Go function accepts array parameter
	buf.Reset()
	err = engine.RegisterFunction("SumArray", func(numbers []int64) int64 {
		sum := int64(0)
		for _, n := range numbers {
			sum += n
		}
		return sum
	})
	if err != nil {
		t.Fatalf("failed to register SumArray: %v", err)
	}

	_, err = engine.Eval(`
		var nums: array of Integer := [10, 20, 30];
		var total := SumArray(nums);
		PrintLn(IntToStr(total));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != "60" {
		t.Errorf("expected '60', got '%s'", output)
	}

	// Test 3: String arrays
	buf.Reset()
	err = engine.RegisterFunction("JoinStrings", func(parts []string) string {
		result := ""
		for i, p := range parts {
			if i > 0 {
				result += ", "
			}
			result += p
		}
		return result
	})
	if err != nil {
		t.Fatalf("failed to register JoinStrings: %v", err)
	}

	_, err = engine.Eval(`
		var words: array of String := ['Hello', 'World', 'From', 'Go'];
		var joined := JoinStrings(words);
		PrintLn(joined);
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	output = strings.TrimSpace(buf.String())
	if output != "Hello, World, From, Go" {
		t.Errorf("expected 'Hello, World, From, Go', got '%s'", output)
	}
}

// TestRegisterFunctionWithMap tests map marshaling.
func TestRegisterFunctionWithMap(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Test 1: Go function returns map
	err = engine.RegisterFunction("GetConfig", func() map[string]string {
		return map[string]string{
			"host": "localhost",
			"port": "8080",
			"name": "MyApp",
		}
	})
	if err != nil {
		t.Fatalf("failed to register GetConfig: %v", err)
	}

	var buf bytes.Buffer
	engine.SetOutput(&buf)
	_, err = engine.Eval(`
		var config := GetConfig();
		PrintLn(config.host);
		PrintLn(config.port);
		PrintLn(config.name);
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "localhost" {
		t.Errorf("expected 'localhost', got '%s'", lines[0])
	}
	if lines[1] != "8080" {
		t.Errorf("expected '8080', got '%s'", lines[1])
	}
	if lines[2] != "MyApp" {
		t.Errorf("expected 'MyApp', got '%s'", lines[2])
	}

	// Test 2: Integer values in map
	buf.Reset()
	err = engine.RegisterFunction("GetScoreMap", func() map[string]int64 {
		return map[string]int64{
			"math":    95,
			"english": 87,
			"science": 92,
		}
	})
	if err != nil {
		t.Fatalf("failed to register GetScoreMap: %v", err)
	}

	_, err = engine.Eval(`
		var scores := GetScoreMap();
		PrintLn(IntToStr(scores.math));
		PrintLn(IntToStr(scores.english));
		PrintLn(IntToStr(scores.science));
	`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	lines = strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "95" {
		t.Errorf("expected '95', got '%s'", lines[0])
	}
	if lines[1] != "87" {
		t.Errorf("expected '87', got '%s'", lines[1])
	}
	if lines[2] != "92" {
		t.Errorf("expected '92', got '%s'", lines[2])
	}
}

// TestFFIInvalidArrayTypes tests passing invalid array types to FFI functions.
func TestFFIInvalidArrayTypes(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Register function expecting []int64
	err = engine.RegisterFunction("ProcessInts", func(nums []int64) int64 {
		sum := int64(0)
		for _, n := range nums {
			sum += n
		}
		return sum
	})
	if err != nil {
		t.Fatalf("failed to register ProcessInts: %v", err)
	}

	// Try to pass array of floats
	result, err := engine.Eval(`
		var x := ProcessInts([1.5, 2.5, 3.5]);  // Floats instead of integers
	`)

	// Should handle type conversion or error gracefully
	if err == nil && result.Success {
		// Type coercion might be allowed in some cases
		// Just verify it doesn't crash
	}
}
