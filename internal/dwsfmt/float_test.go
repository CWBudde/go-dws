package dwsfmt

import (
	"math"
	"testing"
)

func TestFloatToStr(t *testing.T) {
	tests := []struct {
		name string
		want string
		in   float64
	}{
		{name: "zero", in: 0, want: "0"},
		{name: "integral", in: 100, want: "100"},
		{name: "simple fraction", in: 1.25, want: "1.25"},
		{name: "fifteen significant digits", in: 1.0 / 3.0, want: "0.333333333333333"},
		{name: "rounds to fifteen digits", in: 0.1 + 0.2, want: "0.3"},
		{name: "large int64 as exponent", in: 4611686018427387905, want: "4.61168601842739E18"},
		{name: "big exponent drops plus", in: 1e99, want: "1E99"},
		{name: "small exponent keeps minus", in: 1e-5, want: "1E-5"},
		{name: "exponent leading zeros stripped", in: 1e-7, want: "1E-7"},
		{name: "fixed below exponent threshold", in: 1e14, want: "100000000000000"},
		{name: "negative", in: -0.5, want: "-0.5"},
		{name: "positive infinity", in: math.Inf(1), want: "INF"},
		{name: "negative infinity", in: math.Inf(-1), want: "-INF"},
		{name: "not a number", in: math.NaN(), want: "NAN"},
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
