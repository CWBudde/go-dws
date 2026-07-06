package jsonvalue

import (
	"math"
	"strconv"
	"strings"
	"unicode/utf16"
)

// Stringify renders v as a compact JSON string using DWScript's serialization
// rules (see reference/dwscript-original/Source/dwsJSON.pas). It is byte-compatible
// with TdwsJSONWriter: no spaces, forward slashes escaped as \/, lowercase boolean
// literals, and DWScript's number formatting. Object keys are emitted in insertion
// order.
func Stringify(v *Value) string {
	var sb strings.Builder
	writeCompact(&sb, v)
	return sb.String()
}

// StringifyPretty renders v as a beautified JSON string. Each nesting level is
// prefixed by indent repeated by depth, the key/value separator is " : ", and
// members are placed on their own lines (matching TdwsJSONBeautifiedWriter). The
// newline is CRLF, as in DWScript; the fixture harness normalizes trailing \r.
func StringifyPretty(v *Value, indent string) string {
	var sb strings.Builder
	writePretty(&sb, v, indent, 0)
	return sb.String()
}

func writeCompact(sb *strings.Builder, v *Value) {
	switch v.Kind() {
	case KindUndefined, KindNull:
		sb.WriteString("null")
	case KindBoolean:
		if v.BoolValue() {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case KindInt64:
		sb.WriteString(strconv.FormatInt(v.Int64Value(), 10))
	case KindNumber:
		sb.WriteString(FormatNumber(v.NumberValue()))
	case KindString:
		WriteJSONString(sb, v.StringValue())
	case KindObject:
		sb.WriteByte('{')
		for i, key := range v.objKeys {
			if i > 0 {
				sb.WriteByte(',')
			}
			WriteJSONString(sb, key)
			sb.WriteByte(':')
			writeCompact(sb, v.objEntries[key])
		}
		sb.WriteByte('}')
	case KindArray:
		sb.WriteByte('[')
		for i, elem := range v.arrElems {
			if i > 0 {
				sb.WriteByte(',')
			}
			writeCompact(sb, elem)
		}
		sb.WriteByte(']')
	default:
		sb.WriteString("null")
	}
}

func writePretty(sb *strings.Builder, v *Value, indent string, depth int) {
	switch v.Kind() {
	case KindObject:
		if len(v.objKeys) == 0 {
			sb.WriteString("{ }")
			return
		}
		sb.WriteByte('{')
		for i, key := range v.objKeys {
			if i > 0 {
				sb.WriteByte(',')
			}
			sb.WriteString("\r\n")
			sb.WriteString(strings.Repeat(indent, depth+1))
			WriteJSONString(sb, key)
			sb.WriteString(" : ")
			writePretty(sb, v.objEntries[key], indent, depth+1)
		}
		sb.WriteString("\r\n")
		sb.WriteString(strings.Repeat(indent, depth))
		sb.WriteByte('}')
	case KindArray:
		if len(v.arrElems) == 0 {
			sb.WriteString("[ ]")
			return
		}
		sb.WriteByte('[')
		for i, elem := range v.arrElems {
			if i > 0 {
				sb.WriteByte(',')
			}
			sb.WriteString("\r\n")
			sb.WriteString(strings.Repeat(indent, depth+1))
			writePretty(sb, elem, indent, depth+1)
		}
		sb.WriteString("\r\n")
		sb.WriteString(strings.Repeat(indent, depth))
		sb.WriteByte(']')
	default:
		writeCompact(sb, v)
	}
}

// FormatNumber renders a float64 with DWScript's WriteNumber rules: zero (and
// negative zero) as "0", NaN/Infinity as "null", integral values within the Int64
// range as plain integers, and everything else via a general 15-significant-digit
// format with a normalized exponent (uppercase E, no plus sign, no leading zeros).
func FormatNumber(n float64) string {
	if n == 0 {
		return "0"
	}
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return "null"
	}
	if math.Abs(n) <= float64(math.MaxInt64) && math.Round(n) == n {
		return strconv.FormatInt(int64(math.Round(n)), 10)
	}
	return normalizeExponent(strconv.FormatFloat(n, 'G', 15, 64))
}

// normalizeExponent rewrites Go's exponent form (E+99, E-05) into DWScript's
// (E99, E-5): drop a leading '+', keep '-', strip leading zeros of the exponent.
func normalizeExponent(s string) string {
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

const jsonHexDigits = "0123456789ABCDEF"

// WriteJSONString writes s as a JSON string literal (including the surrounding
// quotes) using DWScript's WriteJavaScriptString escaping: \b \t \n \f \r for the
// named control characters, \uXXXX (uppercase) for other controls and for code
// units >= U+0100, \" \\ \/ for quote/backslash/slash, and literal bytes for
// everything in 0x20..0xFF.
func WriteJSONString(sb *strings.Builder, s string) {
	sb.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '\b':
			sb.WriteString(`\b`)
		case r == '\t':
			sb.WriteString(`\t`)
		case r == '\n':
			sb.WriteString(`\n`)
		case r == '\f':
			sb.WriteString(`\f`)
		case r == '\r':
			sb.WriteString(`\r`)
		case r < 0x20:
			writeUnicodeEscape(sb, uint16(r))
		case r == '"':
			sb.WriteString(`\"`)
		case r == '\\':
			sb.WriteString(`\\`)
		case r == '/':
			sb.WriteString(`\/`)
		case r < 0x100:
			sb.WriteRune(r)
		case r <= 0xFFFF:
			writeUnicodeEscape(sb, uint16(r))
		default:
			// Astral plane: DWScript operates on UTF-16 code units, so emit a
			// surrogate pair as two \uXXXX escapes.
			hi, lo := utf16.EncodeRune(r)
			writeUnicodeEscape(sb, uint16(hi))
			writeUnicodeEscape(sb, uint16(lo))
		}
	}
	sb.WriteByte('"')
}

func writeUnicodeEscape(sb *strings.Builder, c uint16) {
	sb.WriteString(`\u`)
	sb.WriteByte(jsonHexDigits[(c>>12)&0xF])
	sb.WriteByte(jsonHexDigits[(c>>8)&0xF])
	sb.WriteByte(jsonHexDigits[(c>>4)&0xF])
	sb.WriteByte(jsonHexDigits[c&0xF])
}
