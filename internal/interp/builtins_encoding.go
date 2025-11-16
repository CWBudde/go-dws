package interp

import (
	"fmt"
	"strings"
	"unicode"
)

// ============================================================================
// Encoding/Escaping Built-in Functions (Phase 9.17.6)
// ============================================================================

// builtinStrToHtml implements the StrToHtml() built-in function.
// It encodes a string for safe use in HTML content.
// StrToHtml(str: String): String
func (i *Interpreter) builtinStrToHtml(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "StrToHtml() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToHtml() expects string argument, got %s", args[0].Type())
	}

	// HTML encode the string
	result := htmlEncode(strVal.Value)
	return &StringValue{Value: result}
}

// htmlEncode encodes a string for safe use in HTML content.
// Encodes: & < > " '
func htmlEncode(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&#39;")
		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}

// builtinStrToHtmlAttribute implements the StrToHtmlAttribute() built-in function.
// It encodes a string for safe use in HTML attributes.
// StrToHtmlAttribute(str: String): String
func (i *Interpreter) builtinStrToHtmlAttribute(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "StrToHtmlAttribute() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToHtmlAttribute() expects string argument, got %s", args[0].Type())
	}

	// HTML attribute encode the string (more restrictive than content encoding)
	result := htmlAttributeEncode(strVal.Value)
	return &StringValue{Value: result}
}

// htmlAttributeEncode encodes a string for safe use in HTML attributes.
// More restrictive than htmlEncode - encodes more characters.
func htmlAttributeEncode(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&#39;")
		case '\n':
			b.WriteString("&#10;")
		case '\r':
			b.WriteString("&#13;")
		case '\t':
			b.WriteString("&#9;")
		default:
			// Encode other control characters
			if r < 32 || r == 127 {
				fmt.Fprintf(&b, "&#%d;", r)
			} else {
				b.WriteRune(r)
			}
		}
	}

	return b.String()
}

// builtinStrToJSON implements the StrToJSON() built-in function.
// It encodes a string for safe use in JSON (escapes special characters).
// StrToJSON(str: String): String
func (i *Interpreter) builtinStrToJSON(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "StrToJSON() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToJSON() expects string argument, got %s", args[0].Type())
	}

	// JSON encode the string
	result := jsonEncode(strVal.Value)
	return &StringValue{Value: result}
}

// jsonEncode encodes a string for safe use in JSON.
// Escapes: \ " and control characters
func jsonEncode(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 10) // Extra space for escape sequences

	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString("\\\\")
		case '"':
			b.WriteString("\\\"")
		case '\b':
			b.WriteString("\\b")
		case '\f':
			b.WriteString("\\f")
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		default:
			// Escape other control characters
			if r < 32 || r == 127 {
				fmt.Fprintf(&b, "\\u%04x", r)
			} else {
				b.WriteRune(r)
			}
		}
	}

	return b.String()
}

// builtinStrToCSSText implements the StrToCSSText() built-in function.
// It encodes a string for safe use in CSS text.
// StrToCSSText(str: String): String
func (i *Interpreter) builtinStrToCSSText(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "StrToCSSText() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToCSSText() expects string argument, got %s", args[0].Type())
	}

	// CSS encode the string
	result := cssEncode(strVal.Value)
	return &StringValue{Value: result}
}

// cssEncode encodes a string for safe use in CSS text.
// Escapes special CSS characters using CSS escape sequences.
func cssEncode(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 10)

	for _, r := range s {
		// Check if character needs escaping
		// CSS requires escaping for: \ " ' ( ) { } [ ] < > ; : , . / ? ! @ # $ % ^ & * = + | ~
		// Also escape newlines, tabs, and control characters
		needsEscape := false

		switch r {
		case '\\', '"', '\'', '(', ')', '{', '}', '[', ']',
			'<', '>', ';', ':', ',', '.', '/', '?', '!',
			'@', '#', '$', '%', '^', '&', '*', '=', '+',
			'|', '~', '\n', '\r', '\t', '\f':
			needsEscape = true
		default:
			// Escape control characters and non-ASCII characters that might be problematic
			if r < 32 || r == 127 {
				needsEscape = true
			}
		}

		if needsEscape {
			// CSS hex escape: \HH or \HHHHHH (up to 6 hex digits)
			fmt.Fprintf(&b, "\\%x ", r)
		} else {
			b.WriteRune(r)
		}
	}

	return b.String()
}

// builtinStrToXML implements the StrToXML() built-in function.
// It encodes a string for safe use in XML.
// StrToXML(str: String): String (default mode)
// StrToXML(str: String, mode: Integer): String (with mode parameter)
func (i *Interpreter) builtinStrToXML(args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return i.newErrorWithLocation(i.currentNode, "StrToXML() expects 1 or 2 arguments, got %d", len(args))
	}

	// First argument must be a string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToXML() expects string as first argument, got %s", args[0].Type())
	}

	// Default mode is 0 (standard XML encoding)
	mode := 0

	// Optional second argument: mode
	if len(args) == 2 {
		switch v := args[1].(type) {
		case *IntegerValue:
			mode = int(v.Value)
		case *SubrangeValue:
			mode = int(v.Value)
		default:
			return i.newErrorWithLocation(i.currentNode, "StrToXML() expects integer as second argument (mode), got %s", args[1].Type())
		}
	}

	// XML encode the string
	result := xmlEncode(strVal.Value, mode)
	return &StringValue{Value: result}
}

// xmlEncode encodes a string for safe use in XML.
// Mode 0: Standard XML encoding (content)
// Mode 1: XML attribute encoding (more restrictive)
// Mode 2: XML text encoding (preserves whitespace)
func xmlEncode(s string, mode int) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			if mode == 1 { // Attribute mode
				b.WriteString("&quot;")
			} else {
				b.WriteRune(r)
			}
		case '\'':
			if mode == 1 { // Attribute mode
				b.WriteString("&apos;")
			} else {
				b.WriteRune(r)
			}
		case '\n':
			switch mode {
			case 1: // Attribute mode - encode newlines
				b.WriteString("&#10;")
			case 2: // Text mode - preserve
				b.WriteRune(r)
			default:
				b.WriteRune(r)
			}
		case '\r':
			switch mode {
			case 1: // Attribute mode
				b.WriteString("&#13;")
			case 2: // Text mode - preserve
				b.WriteRune(r)
			default:
				b.WriteRune(r)
			}
		case '\t':
			if mode == 1 { // Attribute mode
				b.WriteString("&#9;")
			} else {
				b.WriteRune(r)
			}
		default:
			// Encode control characters in all modes
			if !unicode.IsPrint(r) && r != '\n' && r != '\r' && r != '\t' {
				fmt.Fprintf(&b, "&#%d;", r)
			} else {
				b.WriteRune(r)
			}
		}
	}

	return b.String()
}
