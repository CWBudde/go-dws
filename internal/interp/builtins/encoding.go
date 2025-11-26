package builtins

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/lexer"
	pkgast "github.com/cwbudde/go-dws/pkg/ast"
	pkgtoken "github.com/cwbudde/go-dws/pkg/token"
)

// ============================================================================
// Encoding/Escaping Built-in Functions
// ============================================================================
//
// This file contains encoding and escaping functions that have been migrated
// from internal/interp to use the Context interface pattern.
//
// Functions in this file:
//   - StrToHtml: Encode string for HTML content
//   - StrToHtmlAttribute: Encode string for HTML attributes
//   - StrToJSON: Encode string for JSON
//   - StrToCSSText: Encode string for CSS text
//   - StrToXML: Encode string for XML (with optional mode)
//
// These functions provide safe encoding for various output formats to prevent
// injection attacks and ensure proper rendering.

// StrToHtml encodes a string for safe use in HTML content.
// StrToHtml(str: String): String
func StrToHtml(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("StrToHtml() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrToHtml() expects string argument, got %s", args[0].Type())
	}

	// HTML encode the string
	result := htmlEncode(strVal.Value)
	return &runtime.StringValue{Value: result}
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

// StrToHtmlAttribute encodes a string for safe use in HTML attributes.
// StrToHtmlAttribute(str: String): String
func StrToHtmlAttribute(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("StrToHtmlAttribute() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrToHtmlAttribute() expects string argument, got %s", args[0].Type())
	}

	// HTML attribute encode the string (more restrictive than content encoding)
	result := htmlAttributeEncode(strVal.Value)
	return &runtime.StringValue{Value: result}
}

// htmlAttributeEncode encodes a string for safe use in HTML attributes.
// More restrictive than htmlEncode - encodes more characters.
func htmlAttributeEncode(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, r := range s {
		// As per OWASP rule #2: encode everything except alphanumerics
		if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') || r > 255 {
			b.WriteRune(r)
			continue
		}

		code := int(r)
		// Use decimal for common ASCII characters, hex otherwise (matches DWScript reference)
		if code >= 10 && code <= 99 {
			fmt.Fprintf(&b, "&#%d;", code)
		} else {
			fmt.Fprintf(&b, "&#x%X;", code)
		}
	}

	return b.String()
}

// StrToJSON encodes a string for safe use in JSON (escapes special characters).
// StrToJSON(str: String): String
func StrToJSON(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("StrToJSON() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrToJSON() expects string argument, got %s", args[0].Type())
	}

	// JSON encode the string using the standard library for correctness
	jsonEncoded, err := json.Marshal(strVal.Value)
	if err != nil {
		return ctx.NewError("StrToJSON() failed: %v", err)
	}

	return &runtime.StringValue{Value: string(jsonEncoded)}
}

// StrToCSSText encodes a string for safe use in CSS text.
// StrToCSSText(str: String): String
func StrToCSSText(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("StrToCSSText() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrToCSSText() expects string argument, got %s", args[0].Type())
	}

	// CSS encode the string
	result := cssEncode(strVal.Value)
	return &runtime.StringValue{Value: result}
}

// cssEncode encodes a string for safe use in CSS text.
// Escapes special CSS characters using CSS escape sequences.
func cssEncode(s string) string {
	if s == "" {
		return ""
	}

	var b strings.Builder
	// Worst case every character is escaped with a leading backslash
	b.Grow(len(s) * 2)

	for _, r := range s {
		if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') || r > 255 {
			b.WriteRune(r)
			continue
		}

		b.WriteRune('\\')
		b.WriteRune(r)
	}

	return b.String()
}

// StrToXML encodes a string for safe use in XML.
// StrToXML(str: String): String (default mode)
// StrToXML(str: String, mode: Integer): String (with mode parameter)
//
// Mode values:
//   - 0: Standard XML encoding (content)
//   - 1: XML attribute encoding (more restrictive)
//   - 2: XML text encoding (preserves whitespace)
func StrToXML(ctx Context, args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return ctx.NewError("StrToXML() expects 1 or 2 arguments, got %d", len(args))
	}

	// First argument must be a string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrToXML() expects string as first argument, got %s", args[0].Type())
	}

	// Default mode is 0 (standard XML encoding)
	mode := 0

	// Optional second argument: mode
	if len(args) == 2 {
		modeValue, ok := ctx.ToInt64(args[1])
		if !ok {
			return ctx.NewError("StrToXML() expects integer as second argument (mode), got %s", args[1].Type())
		}
		mode = int(modeValue)
	}

	// XML encode the string
	result, err := xmlEncode(strVal.Value, mode)
	if err != nil {
		// Enrich error with source position if available
		msg := err.Error()
		var excPos *lexer.Position
		if node := ctx.CurrentNode(); node != nil {
			switch n := node.(type) {
			case *pkgast.MethodCallExpression:
				pos := n.Method.Pos()
				msg = fmt.Sprintf("%s [line: %d, column: %d]", msg, pos.Line, pos.Column)
				excPos = &lexer.Position{Line: pos.Line, Column: pos.Column}
			default:
				if posNode, ok := node.(interface{ Pos() pkgtoken.Position }); ok {
					pos := posNode.Pos()
					msg = fmt.Sprintf("%s [line: %d, column: %d]", msg, pos.Line, pos.Column)
					excPos = &lexer.Position{Line: pos.Line, Column: pos.Column}
				}
			}
		}

		// Try to raise a DWScript-style exception so try/except can catch it
		if raiser, ok := ctx.(interface {
			RaiseException(className, message string, pos any)
		}); ok {
			raiser.RaiseException("Exception", msg, excPos)
		}

		// Also return an error value to abort the current evaluation path
		return ctx.NewError(msg)
	}
	return &runtime.StringValue{Value: result}
}

// xmlEncode encodes a string for safe use in XML.
// Mode 0: ignore unsupported XML characters
// Mode 1: encode unsupported XML characters as numeric entities
// Other modes: raise an error on unsupported characters
func xmlEncode(s string, mode int) (string, error) {
	var b strings.Builder
	b.Grow(len(s))

	for _, r := range s {
		// Unsupported XML 1.0 characters (control chars except CR/LF/TAB)
		if (r >= 1 && r <= 8) || (r >= 11 && r <= 12) || (r >= 14 && r <= 31) {
			switch mode {
			case 0:
				// Ignore (drop character)
				continue
			case 1:
				fmt.Fprintf(&b, "&#%d;", r)
				continue
			default:
				return "", fmt.Errorf("Unsupported character #%d", r)
			}
		}

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
			b.WriteString("&apos;")
		default:
			b.WriteRune(r)
		}
	}

	return b.String(), nil
}
