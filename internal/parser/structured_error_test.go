package parser

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestStructuredErrorBuilder(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() *StructuredParserError
		wantCode string
		wantKind ErrorKind
		checks   func(*testing.T, *StructuredParserError)
	}{
		{
			name: "basic error with message",
			builder: func() *StructuredParserError {
				return NewStructuredError(ErrKindSyntax).
					WithMessage("test error").
					WithCode(ErrInvalidSyntax).
					WithPosition(lexer.Position{Line: 1, Column: 5}, 3).
					Build()
			},
			wantCode: ErrInvalidSyntax,
			wantKind: ErrKindSyntax,
			checks: func(t *testing.T, err *StructuredParserError) {
				if err.Message != "test error" {
					t.Errorf("expected message 'test error', got %q", err.Message)
				}
				if err.Pos.Line != 1 || err.Pos.Column != 5 {
					t.Errorf("expected position 1:5, got %d:%d", err.Pos.Line, err.Pos.Column)
				}
				if err.Length != 3 {
					t.Errorf("expected length 3, got %d", err.Length)
				}
			},
		},
		{
			name: "unexpected token error",
			builder: func() *StructuredParserError {
				return NewStructuredError(ErrKindUnexpected).
					WithCode(ErrUnexpectedToken).
					WithPosition(lexer.Position{Line: 2, Column: 10}, 1).
					WithExpected(lexer.RPAREN).
					WithActual(lexer.SEMICOLON, ";").
					Build()
			},
			wantCode: ErrUnexpectedToken,
			wantKind: ErrKindUnexpected,
			checks: func(t *testing.T, err *StructuredParserError) {
				if len(err.Expected) != 1 {
					t.Errorf("expected 1 expected value, got %d", len(err.Expected))
				}
				if err.Actual == "" {
					t.Error("expected Actual to be set")
				}
				if err.ActualType != lexer.SEMICOLON {
					t.Errorf("expected ActualType SEMICOLON, got %v", err.ActualType)
				}
			},
		},
		{
			name: "missing token error",
			builder: func() *StructuredParserError {
				return NewStructuredError(ErrKindMissing).
					WithCode(ErrMissingRParen).
					WithPosition(lexer.Position{Line: 3, Column: 15}, 1).
					WithExpected(lexer.RPAREN).
					Build()
			},
			wantCode: ErrMissingRParen,
			wantKind: ErrKindMissing,
			checks: func(t *testing.T, err *StructuredParserError) {
				if len(err.Expected) != 1 || err.Expected[0] != lexer.RPAREN.String() {
					t.Errorf("expected RPAREN in Expected, got %v", err.Expected)
				}
			},
		},
		{
			name: "error with suggestions",
			builder: func() *StructuredParserError {
				return NewStructuredError(ErrKindMissing).
					WithMessage("missing closing brace").
					WithCode(ErrMissingRBrace).
					WithPosition(lexer.Position{Line: 10, Column: 1}, 1).
					WithSuggestion("add '}' to close the block").
					WithSuggestion("check that all opening braces have matching closing braces").
					Build()
			},
			wantCode: ErrMissingRBrace,
			wantKind: ErrKindMissing,
			checks: func(t *testing.T, err *StructuredParserError) {
				if len(err.Suggestions) != 2 {
					t.Errorf("expected 2 suggestions, got %d", len(err.Suggestions))
				}
			},
		},
		{
			name: "error with block context",
			builder: func() *StructuredParserError {
				ctx := &BlockContext{
					BlockType: "begin",
					StartPos:  lexer.Position{Line: 5, Column: 1},
					StartLine: 5,
				}
				return NewStructuredError(ErrKindMissing).
					WithMessage("missing 'end'").
					WithCode(ErrMissingEnd).
					WithPosition(lexer.Position{Line: 20, Column: 1}, 1).
					WithBlockContext(ctx).
					Build()
			},
			wantCode: ErrMissingEnd,
			wantKind: ErrKindMissing,
			checks: func(t *testing.T, err *StructuredParserError) {
				if err.BlockContext == nil {
					t.Error("expected BlockContext to be set")
				}
				if err.BlockContext != nil && err.BlockContext.BlockType != "begin" {
					t.Errorf("expected block type 'begin', got %q", err.BlockContext.BlockType)
				}
			},
		},
		{
			name: "error with related positions",
			builder: func() *StructuredParserError {
				return NewStructuredError(ErrKindMissing).
					WithMessage("missing closing brace").
					WithCode(ErrMissingRBrace).
					WithPosition(lexer.Position{Line: 20, Column: 1}, 1).
					WithRelatedPosition(lexer.Position{Line: 5, Column: 10}, "opening brace here").
					Build()
			},
			wantCode: ErrMissingRBrace,
			wantKind: ErrKindMissing,
			checks: func(t *testing.T, err *StructuredParserError) {
				if len(err.RelatedPos) != 1 {
					t.Errorf("expected 1 related position, got %d", len(err.RelatedPos))
				}
				if len(err.RelatedMessages) != 1 {
					t.Errorf("expected 1 related message, got %d", len(err.RelatedMessages))
				}
			},
		},
		{
			name: "error with parse phase",
			builder: func() *StructuredParserError {
				return NewStructuredError(ErrKindInvalid).
					WithMessage("invalid expression").
					WithCode(ErrInvalidExpression).
					WithPosition(lexer.Position{Line: 8, Column: 5}, 3).
					WithParsePhase("binary expression").
					Build()
			},
			wantCode: ErrInvalidExpression,
			wantKind: ErrKindInvalid,
			checks: func(t *testing.T, err *StructuredParserError) {
				if err.ParsePhase != "binary expression" {
					t.Errorf("expected parse phase 'binary expression', got %q", err.ParsePhase)
				}
			},
		},
		{
			name: "error with notes",
			builder: func() *StructuredParserError {
				return NewStructuredError(ErrKindInvalid).
					WithMessage("invalid syntax").
					WithCode(ErrInvalidSyntax).
					WithPosition(lexer.Position{Line: 12, Column: 8}, 4).
					WithNote("DWScript uses ':=' for assignment, not '='").
					WithNote("Use '=' for comparison").
					Build()
			},
			wantCode: ErrInvalidSyntax,
			wantKind: ErrKindInvalid,
			checks: func(t *testing.T, err *StructuredParserError) {
				if len(err.Notes) != 2 {
					t.Errorf("expected 2 notes, got %d", len(err.Notes))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.builder()

			if err.Code != tt.wantCode {
				t.Errorf("expected code %q, got %q", tt.wantCode, err.Code)
			}

			if err.Kind != tt.wantKind {
				t.Errorf("expected kind %q, got %q", tt.wantKind, err.Kind)
			}

			if tt.checks != nil {
				tt.checks(t, err)
			}
		})
	}
}

func TestStructuredError_Error(t *testing.T) {
	tests := []struct {
		name      string
		err       *StructuredParserError
		wantSubstr []string // Substrings that should appear in error message
	}{
		{
			name: "basic error message",
			err: NewStructuredError(ErrKindSyntax).
				WithMessage("test error").
				WithPosition(lexer.Position{Line: 1, Column: 5}, 3).
				Build(),
			wantSubstr: []string{"test error", "1:5"},
		},
		{
			name: "error with block context",
			err: NewStructuredError(ErrKindMissing).
				WithMessage("missing 'end'").
				WithPosition(lexer.Position{Line: 10, Column: 1}, 1).
				WithBlockContext(&BlockContext{
					BlockType: "begin",
					StartLine: 5,
				}).
				Build(),
			wantSubstr: []string{"missing 'end'", "10:1", "begin block", "line 5"},
		},
		{
			name: "error with parse phase",
			err: NewStructuredError(ErrKindInvalid).
				WithMessage("invalid expression").
				WithPosition(lexer.Position{Line: 8, Column: 3}, 2).
				WithParsePhase("binary expression").
				Build(),
			wantSubstr: []string{"invalid expression", "8:3", "binary expression"},
		},
		{
			name: "auto-generated message for missing token",
			err: NewStructuredError(ErrKindMissing).
				WithPosition(lexer.Position{Line: 5, Column: 10}, 1).
				WithExpected(lexer.RPAREN).
				Build(),
			wantSubstr: []string{"missing", "RPAREN", "5:10"},
		},
		{
			name: "auto-generated message for unexpected token",
			err: NewStructuredError(ErrKindUnexpected).
				WithPosition(lexer.Position{Line: 7, Column: 2}, 1).
				WithExpected(lexer.RPAREN).
				WithActual(lexer.SEMICOLON, ";").
				Build(),
			wantSubstr: []string{"expected", ")", "got", ";", "7:2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()

			for _, substr := range tt.wantSubstr {
				if !strings.Contains(errMsg, substr) {
					t.Errorf("expected error message to contain %q, got: %s", substr, errMsg)
				}
			}
		})
	}
}

func TestStructuredError_DetailedError(t *testing.T) {
	tests := []struct {
		name       string
		err        *StructuredParserError
		wantSubstr []string // Substrings that should appear in detailed error
	}{
		{
			name: "detailed error with suggestions",
			err: NewStructuredError(ErrKindMissing).
				WithMessage("missing closing brace").
				WithPosition(lexer.Position{Line: 10, Column: 1}, 1).
				WithExpected(lexer.RBRACE).
				WithSuggestion("add '}' to close the block").
				WithSuggestion("check brace balance").
				Build(),
			wantSubstr: []string{
				"missing closing brace",
				"10:1",
				"Expected:",
				"Suggestions:",
				"add '}' to close the block",
				"check brace balance",
			},
		},
		{
			name: "detailed error with related positions",
			err: NewStructuredError(ErrKindMissing).
				WithMessage("unclosed block").
				WithPosition(lexer.Position{Line: 20, Column: 1}, 1).
				WithRelatedPosition(lexer.Position{Line: 5, Column: 10}, "block starts here").
				Build(),
			wantSubstr: []string{
				"unclosed block",
				"20:1",
				"Related:",
				"5:10",
				"block starts here",
			},
		},
		{
			name: "detailed error with notes",
			err: NewStructuredError(ErrKindInvalid).
				WithMessage("invalid assignment").
				WithPosition(lexer.Position{Line: 8, Column: 5}, 1).
				WithNote("DWScript uses ':=' for assignment").
				WithNote("Use '=' for comparison").
				Build(),
			wantSubstr: []string{
				"invalid assignment",
				"8:5",
				"Notes:",
				"DWScript uses ':=' for assignment",
				"Use '=' for comparison",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detailed := tt.err.DetailedError()

			for _, substr := range tt.wantSubstr {
				if !strings.Contains(detailed, substr) {
					t.Errorf("expected detailed error to contain %q, got:\n%s", substr, detailed)
				}
			}
		})
	}
}

func TestStructuredError_ToParserError(t *testing.T) {
	structErr := NewStructuredError(ErrKindUnexpected).
		WithMessage("test error").
		WithCode(ErrUnexpectedToken).
		WithPosition(lexer.Position{Line: 5, Column: 10}, 3).
		Build()

	parserErr := structErr.ToParserError()

	if parserErr.Message != "test error" {
		t.Errorf("expected message 'test error', got %q", parserErr.Message)
	}

	if parserErr.Code != ErrUnexpectedToken {
		t.Errorf("expected code %q, got %q", ErrUnexpectedToken, parserErr.Code)
	}

	if parserErr.Pos.Line != 5 || parserErr.Pos.Column != 10 {
		t.Errorf("expected position 5:10, got %d:%d", parserErr.Pos.Line, parserErr.Pos.Column)
	}

	if parserErr.Length != 3 {
		t.Errorf("expected length 3, got %d", parserErr.Length)
	}
}

func TestNewUnexpectedTokenError(t *testing.T) {
	err := NewUnexpectedTokenError(
		lexer.Position{Line: 3, Column: 7},
		1,
		lexer.RPAREN,
		lexer.SEMICOLON,
		";",
	)

	if err.Kind != ErrKindUnexpected {
		t.Errorf("expected kind ErrKindUnexpected, got %v", err.Kind)
	}

	if err.Code != ErrUnexpectedToken {
		t.Errorf("expected code %q, got %q", ErrUnexpectedToken, err.Code)
	}

	if len(err.Expected) != 1 {
		t.Errorf("expected 1 expected value, got %d", len(err.Expected))
	}

	if err.ActualType != lexer.SEMICOLON {
		t.Errorf("expected ActualType SEMICOLON, got %v", err.ActualType)
	}

	// Error message should be auto-generated
	errMsg := err.Error()
	if !strings.Contains(errMsg, "expected") || !strings.Contains(errMsg, "got") {
		t.Errorf("expected auto-generated message with 'expected' and 'got', got: %s", errMsg)
	}
}

func TestNewMissingTokenError(t *testing.T) {
	err := NewMissingTokenError(
		lexer.Position{Line: 5, Column: 12},
		1,
		lexer.RPAREN,
		ErrMissingRParen,
	)

	if err.Kind != ErrKindMissing {
		t.Errorf("expected kind ErrKindMissing, got %v", err.Kind)
	}

	if err.Code != ErrMissingRParen {
		t.Errorf("expected code %q, got %q", ErrMissingRParen, err.Code)
	}

	if len(err.Expected) != 1 {
		t.Errorf("expected 1 expected value, got %d", len(err.Expected))
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "missing") {
		t.Errorf("expected auto-generated message with 'missing', got: %s", errMsg)
	}
}

func TestNewInvalidExpressionError(t *testing.T) {
	tests := []struct {
		name   string
		reason string
	}{
		{
			name:   "with reason",
			reason: "binary operator requires two operands",
		},
		{
			name:   "without reason",
			reason: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewInvalidExpressionError(
				lexer.Position{Line: 8, Column: 3},
				5,
				tt.reason,
			)

			if err.Kind != ErrKindInvalid {
				t.Errorf("expected kind ErrKindInvalid, got %v", err.Kind)
			}

			if err.Code != ErrInvalidExpression {
				t.Errorf("expected code %q, got %q", ErrInvalidExpression, err.Code)
			}

			errMsg := err.Error()
			if tt.reason != "" && !strings.Contains(errMsg, tt.reason) {
				t.Errorf("expected error message to contain reason %q, got: %s", tt.reason, errMsg)
			}
		})
	}
}

func TestAutoGenerateMessage(t *testing.T) {
	tests := []struct {
		name       string
		err        *StructuredParserError
		wantSubstr string
	}{
		{
			name: "missing with single expected",
			err: &StructuredParserError{
				Kind:     ErrKindMissing,
				Expected: []string{")"},
			},
			wantSubstr: "missing )",
		},
		{
			name: "missing with multiple expected",
			err: &StructuredParserError{
				Kind:     ErrKindMissing,
				Expected: []string{")", "end"},
			},
			wantSubstr: "missing one of:",
		},
		{
			name: "unexpected with expected and actual",
			err: &StructuredParserError{
				Kind:     ErrKindUnexpected,
				Expected: []string{")"},
				Actual:   ";",
			},
			wantSubstr: "expected",
		},
		{
			name: "unexpected with multiple expected",
			err: &StructuredParserError{
				Kind:     ErrKindUnexpected,
				Expected: []string{")", "end", ";"},
				Actual:   "begin",
			},
			wantSubstr: "expected one of",
		},
		{
			name: "invalid with actual",
			err: &StructuredParserError{
				Kind:   ErrKindInvalid,
				Actual: "expression",
			},
			wantSubstr: "invalid expression",
		},
		{
			name: "ambiguous",
			err: &StructuredParserError{
				Kind: ErrKindAmbiguous,
			},
			wantSubstr: "ambiguous syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.autoGenerateMessage()
			if !strings.Contains(msg, tt.wantSubstr) {
				t.Errorf("expected message to contain %q, got: %s", tt.wantSubstr, msg)
			}
		})
	}
}
