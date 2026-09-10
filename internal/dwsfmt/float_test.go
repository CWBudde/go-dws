package dwsfmt

import (
	"math"
	"testing"
)

func TestFloatToStr(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want string
	}{
		{"zero", 0, "0"},
		{"integral", 100, "100"},
		{"simple fraction", 1.25, "1.25"},
		{"fifteen significant digits", 1.0 / 3.0, "0.333333333333333"},
		{"rounds to fifteen digits", 0.1 + 0.2, "0.3"},
		{"large int64 as exponent", 4611686018427387905, "4.61168601842739E18"},
		{"big exponent drops plus", 1e99, "1E99"},
		{"small exponent keeps minus", 1e-5, "1E-5"},
		{"exponent leading zeros stripped", 1e-7, "1E-7"},
		{"fixed below exponent threshold", 1e14, "100000000000000"},
		{"negative", -0.5, "-0.5"},
		{"positive infinity", math.Inf(1), "INF"},
		{"negative infinity", math.Inf(-1), "-INF"},
		{"not a number", math.NaN(), "NAN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FloatToStr(tt.in); got != tt.want {
				t.Errorf("FloatToStr(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalizeExponent(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"1E+99", "1E99"},
		{"1E-05", "1E-5"},
		{"1e+06", "1E6"},
		{"123.456", "123.456"},
		{"1E+00", "1E0"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := NormalizeExponent(tt.in); got != tt.want {
				t.Errorf("NormalizeExponent(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
