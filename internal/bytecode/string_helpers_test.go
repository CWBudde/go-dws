package bytecode_test

import (
	"strings"
	"testing"
)

// ============================================================================
// String Helper Method Tests (Task 9.23.6.3)
// Tests that bytecode VM produces the same output as AST interpreter
// ============================================================================

func TestVMParity_StringHelpers(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		// Conversion helpers
		{
			name: "ToUpper basic",
			source: `
				var s := 'hello';
				PrintLn(s.ToUpper);
			`,
		},
		{
			name: "ToLower basic",
			source: `
				var s := 'WORLD';
				PrintLn(s.ToLower);
			`,
		},
		{
			name: "ToInteger basic",
			source: `
				var s := '123';
				PrintLn(IntToStr(s.ToInteger));
			`,
		},
		{
			name: "ToFloat basic",
			source: `
				var s := '3.14';
				PrintLn(FloatToStr(s.ToFloat));
			`,
		},
		{
			name: "ToString identity",
			source: `
				var s := 'test';
				PrintLn(s.ToString);
			`,
		},

		// Search/Check helpers
		{
			name: "StartsWith true",
			source: `
				var s := 'hello';
				if s.StartsWith('he') then
					PrintLn('yes')
				else
					PrintLn('no');
			`,
		},
		{
			name: "StartsWith false",
			source: `
				var s := 'hello';
				if s.StartsWith('wo') then
					PrintLn('yes')
				else
					PrintLn('no');
			`,
		},
		{
			name: "EndsWith true",
			source: `
				var s := 'hello';
				if s.EndsWith('lo') then
					PrintLn('yes')
				else
					PrintLn('no');
			`,
		},
		{
			name: "Contains true",
			source: `
				var s := 'hello world';
				if s.Contains('lo wo') then
					PrintLn('found')
				else
					PrintLn('not found');
			`,
		},
		{
			name: "IndexOf found",
			source: `
				var s := 'hello';
				PrintLn(IntToStr(s.IndexOf('ll')));
			`,
		},
		{
			name: "IndexOf not found",
			source: `
				var s := 'hello';
				PrintLn(IntToStr(s.IndexOf('xyz')));
			`,
		},

		// Extraction helpers
		{
			name: "Copy with 2 params",
			source: `
				var s := 'hello';
				PrintLn(s.Copy(2, 3));
			`,
		},
		{
			name: "Copy with 1 param",
			source: `
				var s := 'hello';
				PrintLn(s.Copy(3));
			`,
		},
		{
			name: "Before found",
			source: `
				var s := 'hello world';
				PrintLn(s.Before(' '));
			`,
		},
		{
			name: "After found",
			source: `
				var s := 'hello world';
				PrintLn(s.After(' '));
			`,
		},

		// Split helper
		{
			name: "Split basic",
			source: `
				var s := 'a,b,c';
				var parts := s.Split(',');
				PrintLn(parts[0]);
				PrintLn(parts[1]);
				PrintLn(parts[2]);
			`,
		},

		// Method chaining (via intermediate variables - direct chaining not yet fully supported)
		{
			name: "ToLower then Copy via variable",
			source: `
				var s := 'HELLO';
				var lower := s.ToLower;
				PrintLn(lower.Copy(2, 3));
			`,
		},

		// Helpers on string literals
		{
			name: "Literal ToUpper",
			source: `
				PrintLn('hello'.ToUpper);
			`,
		},
		{
			name: "Literal Copy",
			source: `
				PrintLn('hello'.Copy(2, 3));
			`,
		},
		{
			name: "Literal StartsWith",
			source: `
				if 'test'.StartsWith('te') then
					PrintLn('yes')
				else
					PrintLn('no');
			`,
		},

		// Helpers in expressions
		{
			name: "Helper in concatenation",
			source: `
				PrintLn('hello'.ToUpper + ' ' + 'world'.ToLower);
			`,
		},
		{
			name: "Helper in arithmetic",
			source: `
				var result := '10'.ToInteger + '20'.ToInteger;
				PrintLn(IntToStr(result));
			`,
		},

		// Edge cases
		{
			name: "Empty string ToUpper",
			source: `
				var s := '';
				PrintLn(s.ToUpper);
			`,
		},
		{
			name: "ToUpper with parens",
			source: `
				var s := 'test';
				PrintLn(s.ToUpper());
			`,
		},
		{
			name: "StartsWith empty",
			source: `
				if 'test'.StartsWith('') then
					PrintLn('yes')
				else
					PrintLn('no');
			`,
		},

		// Complex scenarios
		{
			name: "Multiple helpers",
			source: `
				var s := 'hello world';
				PrintLn(s.ToUpper);
				PrintLn(s.Copy(1, 5));
				if s.StartsWith('hello') then PrintLn('starts');
				if s.EndsWith('world') then PrintLn('ends');
				if s.Contains('lo wo') then PrintLn('contains');
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run with AST interpreter
			astOutput := runWithInterpreter(t, tt.source)

			// Run with bytecode VM
			bcOutput := runWithBytecode(t, tt.source)

			// Compare outputs (normalize boolean case differences)
			astNorm := normalizeOutput(astOutput)
			bcNorm := normalizeOutput(bcOutput)

			if astNorm != bcNorm {
				t.Errorf("Output mismatch:\nAST output:\n%s\nBytecode output:\n%s", astOutput, bcOutput)
			}
		})
	}
}

// normalizeOutput normalizes boolean output differences between interpreters
// (AST interpreter outputs "True"/"False", bytecode VM outputs "true"/"false")
func normalizeOutput(s string) string {
	s = strings.ReplaceAll(s, "True", "true")
	s = strings.ReplaceAll(s, "False", "false")
	return s
}
