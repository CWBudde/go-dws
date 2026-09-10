package encoding

import (
	"fmt"
	"html"
	"strings"
)

// This file implements the web-oriented codecs shared by DWScript's
// EncodingLib encoder classes (URLEncodedEncoder, HTMLTextEncoder,
// HTMLAttributeEncoder) and the StrToHtml* built-in functions.

// ============================================================================
// URL encoding
// ============================================================================

// URLEncodedEncode percent-encodes a script string. The string is first
// converted to UTF-8; every byte outside the RFC 3986 unreserved set
// (A-Z a-z 0-9 - . _ ~) is emitted as %XX with uppercase hexadecimal digits.
func URLEncodedEncode(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	for _, b := range []byte(s) {
		if isURLUnreserved(b) {
			sb.WriteByte(b)
			continue
		}
		sb.WriteByte('%')
		sb.WriteByte(upperHexDigits[b>>4])
		sb.WriteByte(upperHexDigits[b&0x0F])
	}
	return sb.String()
}

const upperHexDigits = "0123456789ABCDEF"

// isURLUnreserved reports whether a byte may appear literally in a URL.
func isURLUnreserved(b byte) bool {
	switch {
	case b >= 'A' && b <= 'Z', b >= 'a' && b <= 'z', b >= '0' && b <= '9':
		return true
	case b == '-', b == '.', b == '_', b == '~':
		return true
	}
	return false
}

// URLEncodedDecode reverses URLEncodedEncode. A '%' escape whose two hex
// digits are invalid yields U+FFFD, a truncated escape at the end of the input
// ends decoding, and byte sequences that are not valid UTF-8 also become
// U+FFFD. Decoding never fails.
func URLEncodedDecode(s string) string {
	chars := []rune(s)
	out := make([]byte, 0, len(chars))

	for i := 0; i < len(chars); {
		c := chars[i]
		if c != '%' {
			out = appendScriptRuneAsUTF8(out, c)
			i++
			continue
		}
		if i+2 >= len(chars) {
			// Truncated escape: DWScript stops decoding here.
			break
		}
		high, okHigh := hexDigitValue(chars[i+1])
		low, okLow := hexDigitValue(chars[i+2])
		if !okHigh || !okLow {
			out = append(out, "�"...)
		} else {
			out = append(out, byte(high<<4|low))
		}
		i += 3
	}

	return decodeUTF8Lossy(out)
}

// appendScriptRuneAsUTF8 appends the UTF-8 encoding of a literal character.
func appendScriptRuneAsUTF8(dst []byte, r rune) []byte {
	return append(dst, string(r)...)
}

// ============================================================================
// HTML encoding
// ============================================================================

// HTMLTextEncode escapes a script string for use as HTML text content.
// It encodes &, <, >, " and ' plus the non-breaking space (U+00A0).
func HTMLTextEncode(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	for _, r := range s {
		switch r {
		case '&':
			sb.WriteString("&amp;")
		case '<':
			sb.WriteString("&lt;")
		case '>':
			sb.WriteString("&gt;")
		case '"':
			sb.WriteString("&quot;")
		case '\'':
			sb.WriteString("&#39;")
		case '\u00A0': // non-breaking space
			sb.WriteString("&nbsp;")
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// HTMLTextDecode turns HTML markup back into plain text: tags are dropped and
// character references are resolved. Unrecognized references are left as-is,
// so "&&&bug;&;" decodes to itself.
func HTMLTextDecode(s string) string {
	return html.UnescapeString(stripHTMLTags(s))
}

// stripHTMLTags removes <...> markup, honouring quoted attribute values so a
// '>' inside an attribute does not terminate the tag early. An unterminated
// tag consumes the rest of the input, as a browser would.
func stripHTMLTags(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))

	inTag := false
	var quote rune
	for _, r := range s {
		switch {
		case !inTag && r == '<':
			inTag = true
			quote = 0
		case !inTag:
			sb.WriteRune(r)
		case quote != 0:
			if r == quote {
				quote = 0
			}
		case r == '\'' || r == '"':
			quote = r
		case r == '>':
			inTag = false
		}
	}
	return sb.String()
}

// HTMLAttributeEncode escapes a script string for use inside an HTML attribute
// value. Following OWASP rule #2 every character that is not alphanumeric (and
// not above U+00FF) is emitted as a numeric character reference.
func HTMLAttributeEncode(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	for _, r := range s {
		if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') || r > 255 {
			sb.WriteRune(r)
			continue
		}
		code := int(r)
		// Decimal for the common ASCII range, hexadecimal otherwise
		// (matches the DWScript reference implementation).
		if code >= 10 && code <= 99 {
			fmt.Fprintf(&sb, "&#%d;", code)
		} else {
			fmt.Fprintf(&sb, "&#x%X;", code)
		}
	}
	return sb.String()
}

// HTMLAttributeDecode reverses HTMLAttributeEncode. Attribute values use the
// same character references as HTML text, so decoding is shared.
func HTMLAttributeDecode(s string) string {
	return HTMLTextDecode(s)
}
