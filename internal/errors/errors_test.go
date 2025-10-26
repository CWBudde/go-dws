package errors

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/lexer"
)

func TestCompilerError_Format(t *testing.T) {
	tests := []struct {
		name        string
		pos         lexer.Position
		message     string
		source      string
		file        string
		wantContain []string // Strings that should appear in output
	}{
		{
			name:    "simple error with file",
			pos:     lexer.Position{Line: 1, Column: 10},
			message: "undefined variable 'x'",
			source:  "var y := x + 5;",
			file:    "test.dws",
			wantContain: []string{
				"Error in test.dws:1:10",
				"   1 | var y := x + 5;",
				"^",
				"undefined variable 'x'",
			},
		},
		{
			name:    "error without file",
			pos:     lexer.Position{Line: 5, Column: 15},
			message: "type mismatch",
			source:  "line1\nline2\nline3\nline4\nline5 with error here\nline6",
			file:    "",
			wantContain: []string{
				"Error at line 5:15",
				"   5 | line5 with error here",
				"^",
				"type mismatch",
			},
		},
		{
			name:    "multi-line source",
			pos:     lexer.Position{Line: 2, Column: 5},
			message: "expected semicolon",
			source:  "begin\n  x := 10\n  y := 20;\nend;",
			file:    "script.dws",
			wantContain: []string{
				"Error in script.dws:2:5",
				"   2 |   x := 10",
				"^",
				"expected semicolon",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewCompilerError(tt.pos, tt.message, tt.source, tt.file)
			got := err.Format(false)

			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("Format() output missing expected string\nwant substring: %q\ngot:\n%s", want, got)
				}
			}
		})
	}
}

func TestCompilerError_FormatWithContext(t *testing.T) {
	source := `var x: Integer := 5;
var y: String;
y := 10;
PrintLn(y);`

	tests := []struct {
		name         string
		pos          lexer.Position
		message      string
		contextLines int
		wantContain  []string
	}{
		{
			name:         "error with 1 line context",
			pos:          lexer.Position{Line: 3, Column: 6},
			message:      "cannot assign Integer to String",
			contextLines: 1,
			wantContain: []string{
				"Error in test.dws:3:6",
				"   2 | var y: String;",
				"   3 | y := 10;",
				"   4 | PrintLn(y);",
				"^",
				"cannot assign Integer to String",
			},
		},
		{
			name:         "error with 2 lines context",
			pos:          lexer.Position{Line: 3, Column: 6},
			message:      "type mismatch",
			contextLines: 2,
			wantContain: []string{
				"   1 | var x: Integer := 5;",
				"   2 | var y: String;",
				"   3 | y := 10;",
				"   4 | PrintLn(y);",
				"^",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewCompilerError(tt.pos, tt.message, source, "test.dws")
			got := err.FormatWithContext(tt.contextLines, false)

			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("FormatWithContext() output missing expected string\nwant substring: %q\ngot:\n%s", want, got)
				}
			}
		})
	}
}

func TestCompilerError_getSourceLine(t *testing.T) {
	source := "line1\nline2\nline3\nline4"

	tests := []struct {
		name    string
		lineNum int
		want    string
	}{
		{
			name:    "first line",
			lineNum: 1,
			want:    "line1",
		},
		{
			name:    "middle line",
			lineNum: 2,
			want:    "line2",
		},
		{
			name:    "last line",
			lineNum: 4,
			want:    "line4",
		},
		{
			name:    "out of range (too high)",
			lineNum: 10,
			want:    "",
		},
		{
			name:    "out of range (zero)",
			lineNum: 0,
			want:    "",
		},
		{
			name:    "out of range (negative)",
			lineNum: -1,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewCompilerError(lexer.Position{}, "", source, "")
			got := err.getSourceLine(tt.lineNum)
			if got != tt.want {
				t.Errorf("getSourceLine(%d) = %q, want %q", tt.lineNum, got, tt.want)
			}
		})
	}
}

func TestCompilerError_getSourceContext(t *testing.T) {
	source := "line1\nline2\nline3\nline4\nline5"

	tests := []struct {
		name          string
		lineNum       int
		contextBefore int
		contextAfter  int
		want          []string
	}{
		{
			name:          "middle with 1 context",
			lineNum:       3,
			contextBefore: 1,
			contextAfter:  1,
			want:          []string{"line2", "line3", "line4"},
		},
		{
			name:          "first line with context",
			lineNum:       1,
			contextBefore: 1,
			contextAfter:  2,
			want:          []string{"line1", "line2", "line3"},
		},
		{
			name:          "last line with context",
			lineNum:       5,
			contextBefore: 2,
			contextAfter:  1,
			want:          []string{"line3", "line4", "line5"},
		},
		{
			name:          "no context",
			lineNum:       3,
			contextBefore: 0,
			contextAfter:  0,
			want:          []string{"line3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewCompilerError(lexer.Position{}, "", source, "")
			got := err.getSourceContext(tt.lineNum, tt.contextBefore, tt.contextAfter)

			if len(got) != len(tt.want) {
				t.Errorf("getSourceContext() returned %d lines, want %d", len(got), len(tt.want))
				return
			}

			for i, line := range got {
				if line != tt.want[i] {
					t.Errorf("getSourceContext() line %d = %q, want %q", i, line, tt.want[i])
				}
			}
		})
	}
}

func TestFormatErrors(t *testing.T) {
	tests := []struct {
		name        string
		errors      []*CompilerError
		wantContain []string
	}{
		{
			name:        "no errors",
			errors:      []*CompilerError{},
			wantContain: []string{},
		},
		{
			name: "single error",
			errors: []*CompilerError{
				NewCompilerError(
					lexer.Position{Line: 1, Column: 5},
					"syntax error",
					"var x",
					"test.dws",
				),
			},
			wantContain: []string{
				"Error in test.dws:1:5",
				"syntax error",
			},
		},
		{
			name: "multiple errors",
			errors: []*CompilerError{
				NewCompilerError(
					lexer.Position{Line: 1, Column: 5},
					"first error",
					"var x",
					"test.dws",
				),
				NewCompilerError(
					lexer.Position{Line: 3, Column: 10},
					"second error",
					"line1\nline2\ny := 10",
					"test.dws",
				),
			},
			wantContain: []string{
				"Compilation failed with 2 error(s)",
				"[Error 1 of 2]",
				"first error",
				"[Error 2 of 2]",
				"second error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatErrors(tt.errors, false)

			if len(tt.errors) == 0 && got != "" {
				t.Errorf("FormatErrors() with no errors should return empty string, got %q", got)
				return
			}

			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("FormatErrors() output missing expected string\nwant substring: %q\ngot:\n%s", want, got)
				}
			}
		})
	}
}

func TestFromStringErrors(t *testing.T) {
	tests := []struct {
		name         string
		stringErrors []string
		source       string
		file         string
		wantCount    int
		checkFirst   func(*testing.T, *CompilerError)
	}{
		{
			name: "errors with position info",
			stringErrors: []string{
				"undefined variable 'x' at 5:10",
				"type mismatch at 10:15",
			},
			source:    "some source code",
			file:      "test.dws",
			wantCount: 2,
			checkFirst: func(t *testing.T, err *CompilerError) {
				if err.Pos.Line != 5 {
					t.Errorf("First error line = %d, want 5", err.Pos.Line)
				}
				if err.Pos.Column != 10 {
					t.Errorf("First error column = %d, want 10", err.Pos.Column)
				}
				if !strings.Contains(err.Message, "undefined variable 'x'") {
					t.Errorf("First error message = %q, want to contain 'undefined variable'", err.Message)
				}
			},
		},
		{
			name: "errors without position info",
			stringErrors: []string{
				"general error message",
			},
			source:    "code",
			file:      "",
			wantCount: 1,
			checkFirst: func(t *testing.T, err *CompilerError) {
				if err.Pos.Line != 0 {
					t.Errorf("Error without position should have Line=0, got %d", err.Pos.Line)
				}
				if err.Message != "general error message" {
					t.Errorf("Message = %q, want %q", err.Message, "general error message")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromStringErrors(tt.stringErrors, tt.source, tt.file)

			if len(got) != tt.wantCount {
				t.Errorf("FromStringErrors() returned %d errors, want %d", len(got), tt.wantCount)
				return
			}

			if tt.wantCount > 0 && tt.checkFirst != nil {
				tt.checkFirst(t, got[0])
			}
		})
	}
}

func TestParseErrorString(t *testing.T) {
	tests := []struct {
		name        string
		errStr      string
		wantLine    int
		wantColumn  int
		wantMessage string
	}{
		{
			name:        "error with position",
			errStr:      "undefined variable 'x' at 10:15",
			wantLine:    10,
			wantColumn:  15,
			wantMessage: "undefined variable 'x'",
		},
		{
			name:        "error without position",
			errStr:      "general error message",
			wantLine:    0,
			wantColumn:  0,
			wantMessage: "general error message",
		},
		{
			name:        "error with 'at' in message",
			errStr:      "error at location in code at 5:8",
			wantLine:    5,
			wantColumn:  8,
			wantMessage: "error at location in code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, message := parseErrorString(tt.errStr)

			if pos.Line != tt.wantLine {
				t.Errorf("parseErrorString() line = %d, want %d", pos.Line, tt.wantLine)
			}
			if pos.Column != tt.wantColumn {
				t.Errorf("parseErrorString() column = %d, want %d", pos.Column, tt.wantColumn)
			}
			if message != tt.wantMessage {
				t.Errorf("parseErrorString() message = %q, want %q", message, tt.wantMessage)
			}
		})
	}
}

func TestCompilerError_ErrorInterface(t *testing.T) {
	err := NewCompilerError(
		lexer.Position{Line: 1, Column: 5},
		"test error",
		"var x",
		"test.dws",
	)

	// Test that it implements error interface
	var _ error = err

	// Test Error() method
	errStr := err.Error()
	if !strings.Contains(errStr, "test error") {
		t.Errorf("Error() should contain 'test error', got: %s", errStr)
	}
}

func TestFormatWithColor(t *testing.T) {
	err := NewCompilerError(
		lexer.Position{Line: 1, Column: 5},
		"test error",
		"var x := 10;",
		"test.dws",
	)

	// Test with color enabled
	colorOutput := err.Format(true)
	if !strings.Contains(colorOutput, "\033[") {
		t.Error("Format(true) should contain ANSI color codes")
	}

	// Test without color
	plainOutput := err.Format(false)
	if strings.Contains(plainOutput, "\033[") {
		t.Error("Format(false) should not contain ANSI color codes")
	}
}
