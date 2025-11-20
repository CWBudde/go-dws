package ident

import (
	"sort"
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "variable", "variable"},
		{"uppercase", "VARIABLE", "variable"},
		{"mixed case", "MyVariable", "myvariable"},
		{"camelCase", "myVariableName", "myvariablename"},
		{"PascalCase", "MyVariableName", "myvariablename"},
		{"with numbers", "Var123", "var123"},
		{"with underscores", "My_Var_Name", "my_var_name"},
		{"empty string", "", ""},
		{"single char lower", "x", "x"},
		{"single char upper", "X", "x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normalize(tt.input)
			if result != tt.expected {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeIdempotent(t *testing.T) {
	// Normalizing twice should produce the same result
	inputs := []string{"Variable", "VARIABLE", "variable", "MyVar"}

	for _, input := range inputs {
		first := Normalize(input)
		second := Normalize(first)
		if first != second {
			t.Errorf("Normalize not idempotent: Normalize(%q) = %q, Normalize(%q) = %q",
				input, first, first, second)
		}
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{"exact match lowercase", "variable", "variable", true},
		{"exact match uppercase", "VARIABLE", "VARIABLE", true},
		{"lowercase vs uppercase", "variable", "VARIABLE", true},
		{"mixed case match", "MyVariable", "myvariable", true},
		{"camelCase vs PascalCase", "myVariable", "MyVariable", true},
		{"all caps vs lowercase", "FUNCTION", "function", true},
		{"different words", "variable", "function", false},
		{"different lengths", "var", "variable", false},
		{"substring", "var", "variable", false},
		{"empty vs empty", "", "", true},
		{"empty vs non-empty", "", "x", false},
		{"single char equal", "x", "X", true},
		{"single char different", "x", "y", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Equal(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Equal(%q, %q) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}

			// Test symmetry: Equal(a, b) should equal Equal(b, a)
			reverse := Equal(tt.b, tt.a)
			if result != reverse {
				t.Errorf("Equal not symmetric: Equal(%q, %q) = %v, but Equal(%q, %q) = %v",
					tt.a, tt.b, result, tt.b, tt.a, reverse)
			}
		})
	}
}

func TestEqualTransitivity(t *testing.T) {
	// If Equal(a, b) and Equal(b, c), then Equal(a, c) should be true
	a := "Variable"
	b := "variable"
	c := "VARIABLE"

	if !Equal(a, b) {
		t.Errorf("Equal(%q, %q) should be true", a, b)
	}
	if !Equal(b, c) {
		t.Errorf("Equal(%q, %q) should be true", b, c)
	}
	if !Equal(a, c) {
		t.Errorf("Equal(%q, %q) should be true (transitivity)", a, c)
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int // <0 if a<b, 0 if a==b, >0 if a>b
	}{
		{"equal lowercase", "abc", "abc", 0},
		{"equal different case", "ABC", "abc", 0},
		{"less than", "abc", "def", -1},
		{"greater than", "def", "abc", 1},
		{"case insensitive less", "ABC", "def", -1},
		{"case insensitive greater", "XYZ", "abc", 1},
		{"prefix", "abc", "abcd", -1},
		{"empty vs non-empty", "", "x", -1},
		{"non-empty vs empty", "x", "", 1},
		{"empty vs empty", "", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compare(tt.a, tt.b)

			// Check sign matches expected
			var resultSign int
			if result < 0 {
				resultSign = -1
			} else if result > 0 {
				resultSign = 1
			} else {
				resultSign = 0
			}

			if resultSign != tt.expected {
				t.Errorf("Compare(%q, %q) = %d (sign: %d), want sign %d",
					tt.a, tt.b, result, resultSign, tt.expected)
			}

			// Test antisymmetry: Compare(a, b) = -Compare(b, a)
			reverse := Compare(tt.b, tt.a)
			if result != -reverse && (result != 0 || reverse != 0) {
				t.Errorf("Compare not antisymmetric: Compare(%q, %q) = %d, Compare(%q, %q) = %d",
					tt.a, tt.b, result, tt.b, tt.a, reverse)
			}
		})
	}
}

func TestCompareSort(t *testing.T) {
	// Test that Compare works correctly with sort.Slice
	names := []string{
		"zebra", "Apple", "BANANA", "cherry", "Date",
	}

	expected := []string{
		"Apple", "BANANA", "cherry", "Date", "zebra",
	}

	sort.Slice(names, func(i, j int) bool {
		return Compare(names[i], names[j]) < 0
	})

	for i, name := range names {
		if !Equal(name, expected[i]) {
			t.Errorf("After sort, names[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		search   string
		slice    []string
		expected bool
	}{
		{"found exact", []string{"abc", "def", "ghi"}, "def", true},
		{"found case insensitive", []string{"abc", "def", "ghi"}, "DEF", true},
		{"not found", []string{"abc", "def", "ghi"}, "xyz", false},
		{"empty slice", []string{}, "abc", false},
		{"empty search in empty", []string{}, "", false},
		{"empty search in non-empty", []string{"abc"}, "", false},
		{"found first", []string{"abc", "def"}, "ABC", true},
		{"found last", []string{"abc", "def"}, "DEF", true},
		{"partial match not found", []string{"variable"}, "var", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Contains(tt.slice, tt.search)
			if result != tt.expected {
				t.Errorf("Contains(%v, %q) = %v, want %v",
					tt.slice, tt.search, result, tt.expected)
			}
		})
	}
}

func TestIndex(t *testing.T) {
	tests := []struct {
		name     string
		search   string
		slice    []string
		expected int
	}{
		{"found at 0", []string{"abc", "def", "ghi"}, "abc", 0},
		{"found at 1", []string{"abc", "def", "ghi"}, "def", 1},
		{"found at 2", []string{"abc", "def", "ghi"}, "ghi", 2},
		{"case insensitive", []string{"abc", "def", "ghi"}, "DEF", 1},
		{"not found", []string{"abc", "def", "ghi"}, "xyz", -1},
		{"empty slice", []string{}, "abc", -1},
		{"duplicates returns first", []string{"abc", "def", "abc"}, "ABC", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Index(tt.slice, tt.search)
			if result != tt.expected {
				t.Errorf("Index(%v, %q) = %d, want %d",
					tt.slice, tt.search, result, tt.expected)
			}
		})
	}
}

func TestIsKeyword(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		keywords []string
		expected bool
	}{
		{"is keyword lowercase", "begin", []string{"begin", "end", "if"}, true},
		{"is keyword uppercase", "BEGIN", []string{"begin", "end", "if"}, true},
		{"is keyword mixed", "Begin", []string{"begin", "end", "if"}, true},
		{"not keyword", "variable", []string{"begin", "end", "if"}, false},
		{"empty keywords", "begin", []string{}, false},
		{"single keyword match", "end", []string{"end"}, true},
		{"single keyword no match", "begin", []string{"end"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsKeyword(tt.s, tt.keywords...)
			if result != tt.expected {
				t.Errorf("IsKeyword(%q, %v) = %v, want %v",
					tt.s, tt.keywords, result, tt.expected)
			}
		})
	}
}

// Benchmarks

func BenchmarkNormalize(b *testing.B) {
	identifiers := []string{
		"MyVariable", "CONSTANT", "functionName", "ClassType",
		"x", "veryLongIdentifierNameThatRepresentsAVariable",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Normalize(identifiers[i%len(identifiers)])
	}
}

func BenchmarkEqual(b *testing.B) {
	pairs := [][2]string{
		{"MyVariable", "myvariable"},
		{"FUNCTION", "function"},
		{"ClassType", "classtype"},
		{"x", "X"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pair := pairs[i%len(pairs)]
		_ = Equal(pair[0], pair[1])
	}
}

func BenchmarkEqualVsToLower(b *testing.B) {
	a := "MyVariableName"
	bLower := "myvariablename"

	b.Run("Equal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Equal(a, bLower)
		}
	})

	b.Run("ToLower", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Normalize(a) == bLower
		}
	})
}

func BenchmarkCompare(b *testing.B) {
	pairs := [][2]string{
		{"abc", "def"},
		{"MyVar", "MYVAR"},
		{"function", "PROCEDURE"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pair := pairs[i%len(pairs)]
		_ = Compare(pair[0], pair[1])
	}
}

func BenchmarkContains(b *testing.B) {
	keywords := []string{
		"begin", "end", "if", "then", "else", "while", "for", "do",
		"var", "const", "function", "procedure", "class",
	}

	searches := []string{"FUNCTION", "variable", "BEGIN", "xyz"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Contains(keywords, searches[i%len(searches)])
	}
}

func BenchmarkIndex(b *testing.B) {
	items := []string{
		"begin", "end", "if", "then", "else", "while", "for", "do",
	}

	searches := []string{"WHILE", "END", "notfound"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Index(items, searches[i%len(searches)])
	}
}
