package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cwbudde/go-dws/pkg/dwscript"
)

func main() {
	// Create engine
	engine, err := dwscript.New(dwscript.WithTypeCheck(false))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create engine: %v\n", err)
		os.Exit(1)
	}

	// Register example functions
	registerMathFunctions(engine)
	registerStringFunctions(engine)
	registerArrayFunctions(engine)
	registerErrorFunctions(engine)
	registerUtilityFunctions(engine)

	// Run the demo script
	scriptPath := "examples/ffi/demo.dws"
	if len(os.Args) > 1 {
		scriptPath = os.Args[1]
	}

	data, err := os.ReadFile(scriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read script: %v\n", err)
		os.Exit(1)
	}

	result, err := engine.Eval(string(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Fprintf(os.Stderr, "Script failed\n")
		os.Exit(1)
	}
}

func registerMathFunctions(engine *dwscript.Engine) {
	// Simple arithmetic
	engine.RegisterFunction("Add", func(a, b int64) int64 {
		return a + b
	})

	engine.RegisterFunction("Multiply", func(a, b int64) int64 {
		return a * b
	})

	// Safe division with error
	engine.RegisterFunction("SafeDivide", func(a, b int64) (int64, error) {
		if b == 0 {
			return 0, errors.New("division by zero")
		}
		return a / b, nil
	})

	// Power function
	engine.RegisterFunction("Power", func(base, exp int64) int64 {
		result := int64(1)
		for i := int64(0); i < exp; i++ {
			result *= base
		}
		return result
	})
}

func registerStringFunctions(engine *dwscript.Engine) {
	// String manipulation
	engine.RegisterFunction("ToUpper", func(s string) string {
		return strings.ToUpper(s)
	})

	engine.RegisterFunction("ToLower", func(s string) string {
		return strings.ToLower(s)
	})

	engine.RegisterFunction("Reverse", func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	})

	// String utilities
	engine.RegisterFunction("Contains", func(s, substr string) bool {
		return strings.Contains(s, substr)
	})

	engine.RegisterFunction("Split", func(s, sep string) []string {
		return strings.Split(s, sep)
	})

	engine.RegisterFunction("Join", func(parts []string, sep string) string {
		return strings.Join(parts, sep)
	})
}

func registerArrayFunctions(engine *dwscript.Engine) {
	// Sum array
	engine.RegisterFunction("SumArray", func(numbers []int64) int64 {
		sum := int64(0)
		for _, n := range numbers {
			sum += n
		}
		return sum
	})

	// Find max
	engine.RegisterFunction("MaxArray", func(numbers []int64) (int64, error) {
		if len(numbers) == 0 {
			return 0, errors.New("empty array")
		}
		max := numbers[0]
		for _, n := range numbers[1:] {
			if n > max {
				max = n
			}
		}
		return max, nil
	})

	// Filter even numbers
	engine.RegisterFunction("FilterEven", func(numbers []int64) []int64 {
		result := []int64{}
		for _, n := range numbers {
			if n%2 == 0 {
				result = append(result, n)
			}
		}
		return result
	})

	// Map function (double values)
	engine.RegisterFunction("DoubleAll", func(numbers []int64) []int64 {
		result := make([]int64, len(numbers))
		for i, n := range numbers {
			result[i] = n * 2
		}
		return result
	})
}

func registerErrorFunctions(engine *dwscript.Engine) {
	// Function that always fails
	engine.RegisterFunction("AlwaysFails", func() (string, error) {
		return "", errors.New("this function always fails")
	})

	// Function that might fail
	engine.RegisterFunction("MightFail", func(shouldFail bool) (string, error) {
		if shouldFail {
			return "", errors.New("operation failed")
		}
		return "success", nil
	})

	// Function that panics
	engine.RegisterFunction("MightPanic", func(trigger bool) string {
		if trigger {
			panic("intentional panic for demonstration")
		}
		return "no panic"
	})
}

func registerUtilityFunctions(engine *dwscript.Engine) {
	// Get environment variable
	engine.RegisterFunction("GetEnv", func(key string) string {
		return os.Getenv(key)
	})

	// Create map
	engine.RegisterFunction("MakeConfig", func() map[string]string {
		return map[string]string{
			"version": "1.0.0",
			"name":    "FFI Demo",
			"author":  "go-dws",
		}
	})

	// Format string
	engine.RegisterFunction("Format", func(template string, args []string) string {
		result := template
		for i, arg := range args {
			placeholder := fmt.Sprintf("{%d}", i)
			result = strings.ReplaceAll(result, placeholder, arg)
		}
		return result
	})

	// Repeat string
	engine.RegisterFunction("RepeatStr", func(s string, count int64) string {
		return strings.Repeat(s, int(count))
	})
}
