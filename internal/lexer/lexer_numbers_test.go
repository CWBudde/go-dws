package lexer

import (
	"testing"
)

func TestIntegerLiterals(t *testing.T) {
	input := `123 0 $FF $ff $10 0xFF 0x10 %1010 %0`

	tests := []struct {
		expectedLiteral string
		expectedType    TokenType
	}{
		{"123", INT},
		{"0", INT},
		{"$FF", INT},
		{"$ff", INT},
		{"$10", INT},
		{"0xFF", INT},
		{"0x10", INT},
		{"%1010", INT},
		{"%0", INT},
		{"", EOF},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestFloatLiterals(t *testing.T) {
	input := `123.45 0.5 3.14 1.5e10 1.5E10 1.5e-5 2.0E+3`

	tests := []struct {
		expectedLiteral string
		expectedType    TokenType
	}{
		{"123.45", FLOAT},
		{"0.5", FLOAT},
		{"3.14", FLOAT},
		{"1.5e10", FLOAT},
		{"1.5E10", FLOAT},
		{"1.5e-5", FLOAT},
		{"2.0E+3", FLOAT},
		{"", EOF},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNumericLiteralsWithUnderscores(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedLiteral string
		expectedType    TokenType
	}{
		// Decimal integers with underscores
		{"decimal with underscores", "123_456", "123_456", INT},
		{"decimal leading underscore", "1_000_000", "1_000_000", INT},
		{"decimal multiple underscores", "1_2_3_4", "1_2_3_4", INT},

		// Binary literals with underscores (% prefix)
		{"binary % with underscores", "%1010_1010", "%1010_1010", INT},
		{"binary % multiple underscores", "%1_0_1_0", "%1_0_1_0", INT},
		{"binary % trailing underscore", "%1010_", "%1010_", INT},

		// Binary literals with underscores (0b prefix)
		{"binary 0b with underscores", "0b1010_1010", "0b1010_1010", INT},
		{"binary 0b leading underscore", "0b_1010", "0b_1010", INT},
		{"binary 0b multiple underscores", "0b1_0_1_0", "0b1_0_1_0", INT},

		// Hexadecimal literals with underscores ($ prefix)
		{"hex $ with underscores", "$FF_00", "$FF_00", INT},
		{"hex $ multiple underscores", "$A_B_C_D", "$A_B_C_D", INT},
		{"hex $ trailing underscore", "$DEAD_", "$DEAD_", INT},

		// Hexadecimal literals with underscores (0x prefix)
		{"hex 0x with underscores", "0xFF_00", "0xFF_00", INT},
		{"hex 0x leading underscore", "0x_DEAD_BEEF", "0x_DEAD_BEEF", INT},
		{"hex 0x multiple underscores", "0xA_B_C_D", "0xA_B_C_D", INT},

		// Float literals with underscores
		{"float with underscores", "123_456.789", "123_456.789", FLOAT},
		{"float with underscores in decimal", "3.14_159", "3.14_159", FLOAT},
		{"float with exponent underscores", "1.5e1_0", "1.5e1_0", FLOAT},
		{"float with all underscores", "1_234.567_89e1_2", "1_234.567_89e1_2", FLOAT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != tt.expectedType {
				t.Fatalf("tokentype wrong. expected=%q, got=%q (literal=%q)",
					tt.expectedType, tok.Type, tok.Literal)
			}

			if tok.Literal != tt.expectedLiteral {
				t.Fatalf("literal wrong. expected=%q, got=%q",
					tt.expectedLiteral, tok.Literal)
			}
		})
	}
}
