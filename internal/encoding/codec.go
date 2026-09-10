package encoding

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// This file implements the byte-oriented codecs shared by DWScript's
// EncodingLib encoder classes (Base64Encoder, Base32Encoder, ...) and the
// function-shaped encoders in internal/builtins.
//
// Terminology
//
//	script string - a Go string whose runes are DWScript characters. DWScript
//	                strings are sequences of 16-bit characters, so a rune here
//	                is a UTF-16 code unit value.
//	byte string   - a script string used as a container for raw bytes: every
//	                rune holds one byte value in 0..255. This mirrors Delphi's
//	                RawByteString, which is what the original DWScript encoders
//	                consume and produce.
//
// Encoders therefore convert between the two representations explicitly rather
// than relying on Go's UTF-8 encoding of the string.

// ScriptStringToBytes reinterprets a script string as raw bytes by taking the
// low byte of every character. Characters above U+00FF are truncated, which
// matches Delphi's implicit UnicodeString-to-RawByteString conversion.
func ScriptStringToBytes(s string) []byte {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		out = append(out, byte(r))
	}
	return out
}

// BytesToScriptString promotes every byte to the character with the same
// ordinal value, producing a byte string (see the file comment).
func BytesToScriptString(b []byte) string {
	var sb strings.Builder
	sb.Grow(len(b))
	for _, c := range b {
		sb.WriteRune(rune(c))
	}
	return sb.String()
}

// ============================================================================
// Base64
// ============================================================================

// base64MIMELineWidth is the maximum number of Base64 characters per line in
// MIME output (RFC 2045).
const base64MIMELineWidth = 76

// base64MIMELineBreak separates MIME output lines.
const base64MIMELineBreak = "\r\n"

// Base64Encode encodes a byte string as standard, padded Base64.
func Base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString(ScriptStringToBytes(s))
}

// Base64EncodeMIME encodes a byte string as standard, padded Base64 and wraps
// the result at 76 characters per line as required by MIME (RFC 2045).
// The empty input yields the empty string (no trailing line break).
func Base64EncodeMIME(s string) string {
	encoded := Base64Encode(s)
	if len(encoded) <= base64MIMELineWidth {
		return encoded
	}

	var sb strings.Builder
	sb.Grow(len(encoded) + 2*(len(encoded)/base64MIMELineWidth+1))
	for start := 0; start < len(encoded); start += base64MIMELineWidth {
		if start > 0 {
			sb.WriteString(base64MIMELineBreak)
		}
		end := start + base64MIMELineWidth
		if end > len(encoded) {
			end = len(encoded)
		}
		sb.WriteString(encoded[start:end])
	}
	return sb.String()
}

// Base64Decode decodes standard Base64. Whitespace (spaces, tabs, CR and LF)
// is ignored anywhere in the input, and missing trailing padding is tolerated.
func Base64Decode(s string) (string, error) {
	cleaned := stripBase64Whitespace(s)
	if cleaned == "" {
		return "", nil
	}
	if pad := len(cleaned) % 4; pad != 0 {
		cleaned += strings.Repeat("=", 4-pad)
	}
	decoded, err := base64.StdEncoding.DecodeString(cleaned)
	if err != nil {
		return "", fmt.Errorf("Invalid Base64 input")
	}
	return BytesToScriptString(decoded), nil
}

// Base64URIEncode encodes a byte string using the URL and filename safe
// alphabet (RFC 4648 §5) without padding.
func Base64URIEncode(s string) string {
	return base64.RawURLEncoding.EncodeToString(ScriptStringToBytes(s))
}

// Base64URIDecode decodes the URL and filename safe Base64 alphabet.
// Padding is optional and whitespace is ignored.
func Base64URIDecode(s string) (string, error) {
	cleaned := stripBase64Whitespace(s)
	cleaned = strings.TrimRight(cleaned, "=")
	if cleaned == "" {
		return "", nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(cleaned)
	if err != nil {
		return "", fmt.Errorf("Invalid Base64 input")
	}
	return BytesToScriptString(decoded), nil
}

// stripBase64Whitespace removes the whitespace characters that Base64 streams
// may be wrapped with.
func stripBase64Whitespace(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	for _, r := range s {
		switch r {
		case ' ', '\t', '\r', '\n':
			continue
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

// ============================================================================
// Base32
// ============================================================================

// base32Alphabet is the RFC 4648 Base32 alphabet.
const base32Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

// Base32Encode encodes a byte string as unpadded RFC 4648 Base32.
// DWScript omits the '=' padding, so this does too.
func Base32Encode(s string) string {
	data := ScriptStringToBytes(s)
	var sb strings.Builder
	sb.Grow((len(data)*8 + 4) / 5)

	var buffer uint32
	var bits uint
	for _, b := range data {
		buffer = buffer<<8 | uint32(b)
		bits += 8
		for bits >= 5 {
			bits -= 5
			sb.WriteByte(base32Alphabet[(buffer>>bits)&0x1F])
		}
	}
	if bits > 0 {
		sb.WriteByte(base32Alphabet[(buffer<<(5-bits))&0x1F])
	}
	return sb.String()
}

// Base32Decode decodes RFC 4648 Base32 into a byte string. Decoding is
// case-insensitive, '=' padding is optional, and leftover bits that cannot
// form a whole byte are discarded. The digit '0' is accepted as an alias for
// the letter 'O', matching the original DWScript decoder's lenient table.
func Base32Decode(s string) (string, error) {
	var out []byte
	var buffer uint32
	var bits uint

	for _, r := range s {
		if r == '=' {
			continue
		}
		value, err := base32CharValue(r)
		if err != nil {
			return "", err
		}
		buffer = buffer<<5 | uint32(value)
		bits += 5
		if bits >= 8 {
			bits -= 8
			out = append(out, byte((buffer>>bits)&0xFF))
		}
	}
	return BytesToScriptString(out), nil
}

// base32CharValue maps a Base32 character to its 5-bit value.
func base32CharValue(r rune) (int, error) {
	switch {
	case r >= 'A' && r <= 'Z':
		return int(r - 'A'), nil
	case r >= 'a' && r <= 'z':
		return int(r - 'a'), nil
	case r >= '2' && r <= '7':
		return int(r-'2') + 26, nil
	case r == '0':
		// Lenient alias for the visually identical letter 'O'.
		return int('O' - 'A'), nil
	}
	return 0, fmt.Errorf("Invalid character (#%d) in Base32", int(r))
}

// ============================================================================
// Base58
// ============================================================================

// base58Alphabet is the Bitcoin Base58 alphabet.
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// Base58Encode encodes a byte string using the Bitcoin Base58 alphabet.
// Every leading zero byte is emitted as a leading '1'.
func Base58Encode(s string) string {
	data := ScriptStringToBytes(s)

	zeros := 0
	for zeros < len(data) && data[zeros] == 0 {
		zeros++
	}

	num := new(big.Int).SetBytes(data)
	radix := big.NewInt(58)
	mod := new(big.Int)
	var digits []byte
	for num.Sign() > 0 {
		num.DivMod(num, radix, mod)
		digits = append(digits, base58Alphabet[mod.Int64()])
	}

	var sb strings.Builder
	sb.Grow(zeros + len(digits))
	for i := 0; i < zeros; i++ {
		sb.WriteByte(base58Alphabet[0])
	}
	for i := len(digits) - 1; i >= 0; i-- {
		sb.WriteByte(digits[i])
	}
	return sb.String()
}

// Base58Decode decodes the Bitcoin Base58 alphabet into a byte string.
// Every leading '1' becomes a leading zero byte.
func Base58Decode(s string) (string, error) {
	num := new(big.Int)
	radix := big.NewInt(58)

	zeros := 0
	countingZeros := true
	for _, r := range s {
		index := strings.IndexRune(base58Alphabet, r)
		if index < 0 {
			return "", fmt.Errorf("Non-base58 character")
		}
		if index == 0 && countingZeros {
			zeros++
		} else {
			countingZeros = false
		}
		num.Mul(num, radix)
		num.Add(num, big.NewInt(int64(index)))
	}

	body := num.Bytes()
	out := make([]byte, zeros+len(body))
	copy(out[zeros:], body)
	return BytesToScriptString(out), nil
}

// ============================================================================
// Hexadecimal
// ============================================================================

const hexDigits = "0123456789abcdef"

// HexadecimalEncode encodes a byte string as lowercase hexadecimal.
func HexadecimalEncode(s string) string {
	var sb strings.Builder
	sb.Grow(2 * len(s))
	for _, r := range s {
		b := byte(r)
		sb.WriteByte(hexDigits[b>>4])
		sb.WriteByte(hexDigits[b&0x0F])
	}
	return sb.String()
}

// HexadecimalDecode decodes a hexadecimal string into a byte string.
// It reports an odd character count and invalid characters using DWScript's
// wording; the reported index is 1-based.
func HexadecimalDecode(s string) (string, error) {
	chars := []rune(s)
	if len(chars)%2 != 0 {
		return "", fmt.Errorf("Even hexadecimal character count expected")
	}

	out := make([]byte, 0, len(chars)/2)
	for i := 0; i < len(chars); i += 2 {
		high, ok := hexDigitValue(chars[i])
		if !ok {
			return "", fmt.Errorf("Invalid hexadecimal character at index %d", i+1)
		}
		low, ok := hexDigitValue(chars[i+1])
		if !ok {
			return "", fmt.Errorf("Invalid hexadecimal character at index %d", i+2)
		}
		out = append(out, byte(high<<4|low))
	}
	return BytesToScriptString(out), nil
}

// hexDigitValue returns the numeric value of a hexadecimal digit.
func hexDigitValue(r rune) (int, bool) {
	switch {
	case r >= '0' && r <= '9':
		return int(r - '0'), true
	case r >= 'a' && r <= 'f':
		return int(r-'a') + 10, true
	case r >= 'A' && r <= 'F':
		return int(r-'A') + 10, true
	}
	return 0, false
}

// ============================================================================
// UTF-8 / UTF-16
// ============================================================================

// UTF8Encode converts a script string into a byte string holding its UTF-8
// representation.
func UTF8Encode(s string) string {
	return BytesToScriptString([]byte(s))
}

// UTF8Decode interprets a byte string as UTF-8 and returns the script string
// it represents. Invalid sequences become U+FFFD.
func UTF8Decode(s string) string {
	data := ScriptStringToBytes(s)
	return decodeUTF8Lossy(data)
}

// UTF16Encode converts a script string into a byte string holding its UTF-16
// representation, in big-endian order when bigEndian is true.
func UTF16Encode(s string, bigEndian bool) string {
	units := utf16.Encode([]rune(s))
	out := make([]byte, 0, 2*len(units))
	for _, unit := range units {
		hi := byte(unit >> 8)
		lo := byte(unit)
		if bigEndian {
			out = append(out, hi, lo)
		} else {
			out = append(out, lo, hi)
		}
	}
	return BytesToScriptString(out)
}

// UTF16Decode interprets a byte string as UTF-16, in big-endian order when
// bigEndian is true. A trailing odd byte is ignored.
func UTF16Decode(s string, bigEndian bool) string {
	data := ScriptStringToBytes(s)
	units := make([]uint16, 0, len(data)/2)
	for i := 0; i+1 < len(data); i += 2 {
		if bigEndian {
			units = append(units, uint16(data[i])<<8|uint16(data[i+1]))
		} else {
			units = append(units, uint16(data[i+1])<<8|uint16(data[i]))
		}
	}
	return string(utf16.Decode(units))
}

// decodeUTF8Lossy decodes UTF-8 bytes, replacing every invalid byte with
// U+FFFD instead of failing.
func decodeUTF8Lossy(data []byte) string {
	var sb strings.Builder
	sb.Grow(len(data))
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		sb.WriteRune(r)
		data = data[size:]
	}
	return sb.String()
}
