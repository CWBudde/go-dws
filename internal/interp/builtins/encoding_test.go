package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

func TestStrToHtml(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
		isError  bool
	}{
		{
			name:     "basic string",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: "hello",
		},
		{
			name:     "with ampersand",
			args:     []Value{&runtime.StringValue{Value: "A & B"}},
			expected: "A &amp; B",
		},
		{
			name:     "with less than",
			args:     []Value{&runtime.StringValue{Value: "x < y"}},
			expected: "x &lt; y",
		},
		{
			name:     "with greater than",
			args:     []Value{&runtime.StringValue{Value: "x > y"}},
			expected: "x &gt; y",
		},
		{
			name:     "with quotes",
			args:     []Value{&runtime.StringValue{Value: `"hello"`}},
			expected: "&quot;hello&quot;",
		},
		{
			name:     "with single quote",
			args:     []Value{&runtime.StringValue{Value: "'hello'"}},
			expected: "&#39;hello&#39;",
		},
		{
			name:     "all special chars",
			args:     []Value{&runtime.StringValue{Value: `<tag attr="value">&'test'</tag>`}},
			expected: "&lt;tag attr=&quot;value&quot;&gt;&amp;&#39;test&#39;&lt;/tag&gt;",
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrToHtml(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, strVal.Value)
			}
		})
	}
}

func TestStrToHtmlAttribute(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
		isError  bool
	}{
		{
			name:     "basic string",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: "hello",
		},
		{
			name:     "with newline",
			args:     []Value{&runtime.StringValue{Value: "hello\nworld"}},
			expected: "hello&#10;world",
		},
		{
			name:     "with carriage return",
			args:     []Value{&runtime.StringValue{Value: "hello\rworld"}},
			expected: "hello&#13;world",
		},
		{
			name:     "with tab",
			args:     []Value{&runtime.StringValue{Value: "hello\tworld"}},
			expected: "hello&#9;world",
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrToHtmlAttribute(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, strVal.Value)
			}
		})
	}
}

func TestStrToJSON(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
		isError  bool
	}{
		{
			name:     "basic string",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: "hello",
		},
		{
			name:     "with backslash",
			args:     []Value{&runtime.StringValue{Value: `C:\path\file`}},
			expected: `C:\\path\\file`,
		},
		{
			name:     "with quotes",
			args:     []Value{&runtime.StringValue{Value: `"quoted"`}},
			expected: `\"quoted\"`,
		},
		{
			name:     "with newline",
			args:     []Value{&runtime.StringValue{Value: "line1\nline2"}},
			expected: `line1\nline2`,
		},
		{
			name:     "with tab",
			args:     []Value{&runtime.StringValue{Value: "col1\tcol2"}},
			expected: `col1\tcol2`,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrToJSON(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, strVal.Value)
			}
		})
	}
}

func TestStrToCSSText(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
		isError  bool
	}{
		{
			name:     "basic string",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: "hello",
		},
		{
			name:     "with special chars",
			args:     []Value{&runtime.StringValue{Value: "a:b;c"}},
			expected: "a\\3a b\\3b c",
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrToCSSText(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, strVal.Value)
			}
		})
	}
}

func TestStrToXML(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
		isError  bool
	}{
		{
			name:     "basic string",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: "hello",
		},
		{
			name:     "with ampersand",
			args:     []Value{&runtime.StringValue{Value: "A & B"}},
			expected: "A &amp; B",
		},
		{
			name:     "with less than",
			args:     []Value{&runtime.StringValue{Value: "x < y"}},
			expected: "x &lt; y",
		},
		{
			name:     "with greater than",
			args:     []Value{&runtime.StringValue{Value: "x > y"}},
			expected: "x &gt; y",
		},
		{
			name:     "attribute mode - quotes",
			args:     []Value{&runtime.StringValue{Value: `"hello"`}, &runtime.IntegerValue{Value: 1}},
			expected: "&quot;hello&quot;",
		},
		{
			name:     "attribute mode - single quote",
			args:     []Value{&runtime.StringValue{Value: "'hello'"}, &runtime.IntegerValue{Value: 1}},
			expected: "&apos;hello&apos;",
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrToXML(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, strVal.Value)
			}
		})
	}
}
