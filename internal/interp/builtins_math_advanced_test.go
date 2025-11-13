package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// Helper function to run a test case
func runTestCase(t *testing.T, input string) (string, *ErrorValue) {
	t.Helper()
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		strs := make([]string, len(p.Errors()))
		for i, err := range p.Errors() {
			strs[i] = err.Error()
		}
		return "", &ErrorValue{Message: strings.Join(strs, ", ")}
	}

	var buf bytes.Buffer
	interp := New(&buf)
	val := interp.Eval(program)

	// Check for errors
	if errVal, ok := val.(*ErrorValue); ok {
		return "", errVal
	}

	return buf.String(), nil
}

func TestBuiltinFactorial(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		isError  bool
	}{
		// Basic cases
		{"factorial of 0", "PrintLn(Factorial(0));", "1\n", false},
		{"factorial of 1", "PrintLn(Factorial(1));", "1\n", false},
		{"factorial of 5", "PrintLn(Factorial(5));", "120\n", false},
		{"factorial of 10", "PrintLn(Factorial(10));", "3628800\n", false},
		{"factorial of 20", "PrintLn(Factorial(20));", "2432902008176640000\n", false},

		// Error cases
		{"negative number", "PrintLn(Factorial(-1));", "", true},
		{"overflow", "PrintLn(Factorial(21));", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runTestCase(t, tt.input)

			if tt.isError {
				if err == nil {
					t.Errorf("expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if output != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, output)
				}
			}
		})
	}
}

func TestBuiltinGcd(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		isError  bool
	}{
		// Basic cases
		{"gcd(12, 8)", "PrintLn(Gcd(12, 8));", "4\n", false},
		{"gcd(48, 18)", "PrintLn(Gcd(48, 18));", "6\n", false},
		{"gcd(100, 50)", "PrintLn(Gcd(100, 50));", "50\n", false},
		{"gcd(17, 19)", "PrintLn(Gcd(17, 19));", "1\n", false}, // Coprime
		{"gcd(0, 5)", "PrintLn(Gcd(0, 5));", "5\n", false},
		{"gcd(5, 0)", "PrintLn(Gcd(5, 0));", "5\n", false},

		// Negative numbers
		{"gcd(-12, 8)", "PrintLn(Gcd(-12, 8));", "4\n", false},
		{"gcd(12, -8)", "PrintLn(Gcd(12, -8));", "4\n", false},
		{"gcd(-12, -8)", "PrintLn(Gcd(-12, -8));", "4\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runTestCase(t, tt.input)

			if tt.isError {
				if err == nil {
					t.Errorf("expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if output != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, output)
				}
			}
		})
	}
}

func TestBuiltinLcm(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		isError  bool
	}{
		// Basic cases
		{"lcm(12, 8)", "PrintLn(Lcm(12, 8));", "24\n", false},
		{"lcm(4, 6)", "PrintLn(Lcm(4, 6));", "12\n", false},
		{"lcm(3, 7)", "PrintLn(Lcm(3, 7));", "21\n", false}, // Coprime
		{"lcm(0, 5)", "PrintLn(Lcm(0, 5));", "0\n", false},
		{"lcm(5, 0)", "PrintLn(Lcm(5, 0));", "0\n", false},

		// Negative numbers
		{"lcm(-12, 8)", "PrintLn(Lcm(-12, 8));", "24\n", false},
		{"lcm(12, -8)", "PrintLn(Lcm(12, -8));", "24\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runTestCase(t, tt.input)

			if tt.isError {
				if err == nil {
					t.Errorf("expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if output != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, output)
				}
			}
		})
	}
}

func TestBuiltinIsPrime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		isError  bool
	}{
		// Primes
		{"2 is prime", "PrintLn(IsPrime(2));", "True\n", false},
		{"3 is prime", "PrintLn(IsPrime(3));", "True\n", false},
		{"5 is prime", "PrintLn(IsPrime(5));", "True\n", false},
		{"7 is prime", "PrintLn(IsPrime(7));", "True\n", false},
		{"11 is prime", "PrintLn(IsPrime(11));", "True\n", false},
		{"97 is prime", "PrintLn(IsPrime(97));", "True\n", false},

		// Non-primes
		{"0 is not prime", "PrintLn(IsPrime(0));", "False\n", false},
		{"1 is not prime", "PrintLn(IsPrime(1));", "False\n", false},
		{"4 is not prime", "PrintLn(IsPrime(4));", "False\n", false},
		{"6 is not prime", "PrintLn(IsPrime(6));", "False\n", false},
		{"8 is not prime", "PrintLn(IsPrime(8));", "False\n", false},
		{"9 is not prime", "PrintLn(IsPrime(9));", "False\n", false},
		{"100 is not prime", "PrintLn(IsPrime(100));", "False\n", false},

		// Negative numbers
		{"negative number", "PrintLn(IsPrime(-5));", "False\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runTestCase(t, tt.input)

			if tt.isError {
				if err == nil {
					t.Errorf("expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if output != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, output)
				}
			}
		})
	}
}

func TestBuiltinLeastFactor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		isError  bool
	}{
		// Basic cases
		{"least factor of 0", "PrintLn(LeastFactor(0));", "1\n", false},
		{"least factor of 1", "PrintLn(LeastFactor(1));", "1\n", false},
		{"least factor of 2", "PrintLn(LeastFactor(2));", "2\n", false},
		{"least factor of 15", "PrintLn(LeastFactor(15));", "3\n", false},
		{"least factor of 21", "PrintLn(LeastFactor(21));", "3\n", false},
		{"least factor of 35", "PrintLn(LeastFactor(35));", "5\n", false},
		{"least factor of 49", "PrintLn(LeastFactor(49));", "7\n", false},
		{"least factor of prime", "PrintLn(LeastFactor(17));", "17\n", false},

		// Even numbers
		{"least factor of 100", "PrintLn(LeastFactor(100));", "2\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runTestCase(t, tt.input)

			if tt.isError {
				if err == nil {
					t.Errorf("expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if output != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, output)
				}
			}
		})
	}
}

func TestBuiltinPopCount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		isError  bool
	}{
		// Basic cases
		{"popcount of 0", "PrintLn(PopCount(0));", "0\n", false},
		{"popcount of 1", "PrintLn(PopCount(1));", "1\n", false},
		{"popcount of 3", "PrintLn(PopCount(3));", "2\n", false},     // 11 in binary
		{"popcount of 7", "PrintLn(PopCount(7));", "3\n", false},     // 111 in binary
		{"popcount of 15", "PrintLn(PopCount(15));", "4\n", false},   // 1111 in binary
		{"popcount of 255", "PrintLn(PopCount(255));", "8\n", false}, // 11111111 in binary
		{"popcount of 256", "PrintLn(PopCount(256));", "1\n", false}, // 100000000 in binary

		// Negative numbers (counted as two's complement)
		{"popcount of -1", "PrintLn(PopCount(-1));", "64\n", false}, // All bits set in int64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runTestCase(t, tt.input)

			if tt.isError {
				if err == nil {
					t.Errorf("expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if output != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, output)
				}
			}
		})
	}
}

func TestBuiltinTestBit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		isError  bool
	}{
		// Basic cases
		{"bit 0 of 1", "PrintLn(TestBit(1, 0));", "True\n", false},
		{"bit 1 of 1", "PrintLn(TestBit(1, 1));", "False\n", false},
		{"bit 0 of 2", "PrintLn(TestBit(2, 0));", "False\n", false},
		{"bit 1 of 2", "PrintLn(TestBit(2, 1));", "True\n", false},
		{"bit 0 of 7", "PrintLn(TestBit(7, 0));", "True\n", false},
		{"bit 1 of 7", "PrintLn(TestBit(7, 1));", "True\n", false},
		{"bit 2 of 7", "PrintLn(TestBit(7, 2));", "True\n", false},
		{"bit 3 of 7", "PrintLn(TestBit(7, 3));", "False\n", false},

		// Edge cases
		{"bit 0 of 0", "PrintLn(TestBit(0, 0));", "False\n", false},

		// Error cases
		{"bit position too large", "PrintLn(TestBit(1, 64));", "", true},
		{"bit position negative", "PrintLn(TestBit(1, -1));", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runTestCase(t, tt.input)

			if tt.isError {
				if err == nil {
					t.Errorf("expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if output != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, output)
				}
			}
		})
	}
}

func TestBuiltinHaversine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		isError  bool
	}{
		// Same location (distance should be 0)
		{"same location", "PrintLn(Haversine(0.0, 0.0, 0.0, 0.0));", "0\n", false},

		// Known distances (approximate)
		// NYC (40.7128, -74.0060) to LA (34.0522, -118.2437) is ~3936 km
		{"NYC to LA", `
			var d := Haversine(40.7128, -74.0060, 34.0522, -118.2437);
			PrintLn(d > 3900.0 and d < 4000.0);
		`, "True\n", false},

		// London (51.5074, -0.1278) to Paris (48.8566, 2.3522) is ~344 km
		{"London to Paris", `
			var d := Haversine(51.5074, -0.1278, 48.8566, 2.3522);
			PrintLn(d > 340.0 and d < 350.0);
		`, "True\n", false},

		// Works with integers too
		{"integer coordinates", `
			var d := Haversine(0, 0, 1, 1);
			PrintLn(d > 0.0);
		`, "True\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runTestCase(t, tt.input)

			if tt.isError {
				if err == nil {
					t.Errorf("expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if output != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, output)
				}
			}
		})
	}
}

func TestBuiltinCompareNum(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		isError  bool
	}{
		// Basic comparisons
		{"equal numbers", "PrintLn(CompareNum(5.0, 5.0));", "0\n", false},
		{"first less than second", "PrintLn(CompareNum(3.0, 5.0));", "-1\n", false},
		{"first greater than second", "PrintLn(CompareNum(7.0, 5.0));", "1\n", false},

		// Integer arguments
		{"integer equal", "PrintLn(CompareNum(5, 5));", "0\n", false},
		{"integer less", "PrintLn(CompareNum(3, 5));", "-1\n", false},
		{"integer greater", "PrintLn(CompareNum(7, 5));", "1\n", false},

		// Mixed types
		{"mixed int and float", "PrintLn(CompareNum(5, 5.0));", "0\n", false},
		{"mixed less", "PrintLn(CompareNum(3, 5.0));", "-1\n", false},

		// NaN handling
		{"both NaN", "PrintLn(CompareNum(NaN, NaN));", "0\n", false},
		{"first is NaN", "PrintLn(CompareNum(NaN, 5.0));", "-1\n", false},
		{"second is NaN", "PrintLn(CompareNum(5.0, NaN));", "1\n", false},

		// Infinity
		{"both infinity", "PrintLn(CompareNum(Infinity, Infinity));", "0\n", false},
		{"infinity vs number", "PrintLn(CompareNum(Infinity, 5.0));", "1\n", false},
		{"number vs infinity", "PrintLn(CompareNum(5.0, Infinity));", "-1\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runTestCase(t, tt.input)

			if tt.isError {
				if err == nil {
					t.Errorf("expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if output != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, output)
				}
			}
		})
	}
}

// Test all functions together in a single program
func TestBuiltinMathAdvancedIntegration(t *testing.T) {
	input := `
// Factorial tests
PrintLn('Factorial:');
PrintLn(Factorial(5));
PrintLn(Factorial(10));

// GCD and LCM tests
PrintLn('GCD and LCM:');
PrintLn(Gcd(48, 18));
PrintLn(Lcm(12, 8));

// Prime tests
PrintLn('Prime tests:');
PrintLn(IsPrime(17));
PrintLn(IsPrime(18));
PrintLn(LeastFactor(15));

// Bit operations
PrintLn('Bit operations:');
PrintLn(PopCount(15));
PrintLn(TestBit(7, 1));

// Haversine
PrintLn('Haversine:');
var d := Haversine(0.0, 0.0, 0.0, 0.0);
PrintLn(d);

// CompareNum
PrintLn('CompareNum:');
PrintLn(CompareNum(5.0, 3.0));
PrintLn(CompareNum(3.0, 5.0));
PrintLn(CompareNum(5.0, 5.0));
`

	expected := `Factorial:
120
3628800
GCD and LCM:
6
24
Prime tests:
True
False
3
Bit operations:
4
True
Haversine:
0
CompareNum:
1
-1
0
`

	output, err := runTestCase(t, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, output)
	}
}
