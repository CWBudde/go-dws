// Package dwsfmt renders scalars the way DWScript (and the Delphi RTL it is
// built on) does, so the interpreter, the JSON connector and the builtins all
// agree on one spelling for a given value.
package dwsfmt

import (
	"math"
	"strconv"
	"strings"
)

// FloatSignificantDigits is the precision Delphi's FloatToStr uses (ffGeneral
// with 15 significant digits).
const FloatSignificantDigits = 15

// FloatToStr formats a float the way Delphi's FloatToStr does: general format
// with at most 15 significant digits, and an uppercase exponent without a plus
// sign or leading zeros ("4.61168601842739E18", "1E99", "1E-5"). Non-finite
// values use Delphi's spellings.
func FloatToStr(f float64) string {
	switch {
	case math.IsInf(f, 1):
		return "INF"
	case math.IsInf(f, -1):
		return "-INF"
	case math.IsNaN(f):
		return "NAN"
	}
	return NormalizeExponent(strconv.FormatFloat(f, 'G', FloatSignificantDigits, 64))
}

// NormalizeExponent rewrites Go's exponent form (E+99, E-05) into DWScript's
// (E99, E-5): drop a leading '+', keep '-', strip leading zeros of the exponent.
func NormalizeExponent(s string) string {
	i := strings.IndexAny(s, "eE")
	if i < 0 {
		return s
	}
	mantissa, exp := s[:i], s[i+1:]
	sign := ""
	if len(exp) > 0 && (exp[0] == '+' || exp[0] == '-') {
		if exp[0] == '-' {
			sign = "-"
		}
		exp = exp[1:]
	}
	exp = strings.TrimLeft(exp, "0")
	if exp == "" {
		exp = "0"
	}
	return mantissa + "E" + sign + exp
}
