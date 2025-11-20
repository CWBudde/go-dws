package semantic

import (
	"testing"
)

// TestAnalyzeStrToHtml tests the StrToHtml() built-in function
func TestAnalyzeStrToHtml(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "StrToHtml with correct argument",
			code:        `var x := StrToHtml('<div>Test</div>');`,
			expectError: false,
		},
		{
			name:        "StrToHtml with wrong argument type",
			code:        `var x := StrToHtml(123);`,
			expectError: true,
			errorMsg:    "expects String as argument",
		},
		{
			name:        "StrToHtml with no arguments",
			code:        `var x := StrToHtml();`,
			expectError: true,
			errorMsg:    "expects 1 argument",
		},
		{
			name:        "StrToHtml with too many arguments",
			code:        `var x := StrToHtml('<div>', '</div>');`,
			expectError: true,
			errorMsg:    "expects 1 argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBuiltinTest(t, tt.code, tt.expectError, tt.errorMsg)
		})
	}
}

// TestAnalyzeStrToHtmlAttribute tests the StrToHtmlAttribute() built-in function
func TestAnalyzeStrToHtmlAttribute(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "StrToHtmlAttribute with correct argument",
			code:        `var x := StrToHtmlAttribute('class="test"');`,
			expectError: false,
		},
		{
			name:        "StrToHtmlAttribute with wrong argument type",
			code:        `var x := StrToHtmlAttribute(123);`,
			expectError: true,
			errorMsg:    "expects String as argument",
		},
		{
			name:        "StrToHtmlAttribute with no arguments",
			code:        `var x := StrToHtmlAttribute();`,
			expectError: true,
			errorMsg:    "expects 1 argument",
		},
		{
			name:        "StrToHtmlAttribute with too many arguments",
			code:        `var x := StrToHtmlAttribute('attr1', 'attr2');`,
			expectError: true,
			errorMsg:    "expects 1 argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBuiltinTest(t, tt.code, tt.expectError, tt.errorMsg)
		})
	}
}

// TestAnalyzeStrToJSON tests the StrToJSON() built-in function
func TestAnalyzeStrToJSON(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "StrToJSON with correct argument",
			code:        `var x := StrToJSON('{"key": "value"}');`,
			expectError: false,
		},
		{
			name:        "StrToJSON with wrong argument type",
			code:        `var x := StrToJSON(123);`,
			expectError: true,
			errorMsg:    "expects String as argument",
		},
		{
			name:        "StrToJSON with no arguments",
			code:        `var x := StrToJSON();`,
			expectError: true,
			errorMsg:    "expects 1 argument",
		},
		{
			name:        "StrToJSON with too many arguments",
			code:        `var x := StrToJSON('test', 'extra');`,
			expectError: true,
			errorMsg:    "expects 1 argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBuiltinTest(t, tt.code, tt.expectError, tt.errorMsg)
		})
	}
}

// TestAnalyzeStrToCSSText tests the StrToCSSText() built-in function
func TestAnalyzeStrToCSSText(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "StrToCSSText with correct argument",
			code:        `var x := StrToCSSText('color: red;');`,
			expectError: false,
		},
		{
			name:        "StrToCSSText with wrong argument type",
			code:        `var x := StrToCSSText(123);`,
			expectError: true,
			errorMsg:    "expects String as argument",
		},
		{
			name:        "StrToCSSText with no arguments",
			code:        `var x := StrToCSSText();`,
			expectError: true,
			errorMsg:    "expects 1 argument",
		},
		{
			name:        "StrToCSSText with too many arguments",
			code:        `var x := StrToCSSText('css', 'extra');`,
			expectError: true,
			errorMsg:    "expects 1 argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBuiltinTest(t, tt.code, tt.expectError, tt.errorMsg)
		})
	}
}

// TestAnalyzeStrToXML tests the StrToXML() built-in function
func TestAnalyzeStrToXML(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "StrToXML with 1 argument",
			code:        `var x := StrToXML('<tag>content</tag>');`,
			expectError: false,
		},
		{
			name:        "StrToXML with 2 arguments (mode)",
			code:        `var x := StrToXML('<tag>content</tag>', 0);`,
			expectError: false,
		},
		{
			name:        "StrToXML with wrong first argument type",
			code:        `var x := StrToXML(123);`,
			expectError: true,
			errorMsg:    "expects String as first argument",
		},
		{
			name:        "StrToXML with wrong second argument type",
			code:        `var x := StrToXML('<tag>', 'mode');`,
			expectError: true,
			errorMsg:    "expects Integer as second argument",
		},
		{
			name:        "StrToXML with no arguments",
			code:        `var x := StrToXML();`,
			expectError: true,
			errorMsg:    "expects 1 or 2 arguments",
		},
		{
			name:        "StrToXML with too many arguments",
			code:        `var x := StrToXML('<tag>', 0, 99);`,
			expectError: true,
			errorMsg:    "expects 1 or 2 arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBuiltinTest(t, tt.code, tt.expectError, tt.errorMsg)
		})
	}
}
