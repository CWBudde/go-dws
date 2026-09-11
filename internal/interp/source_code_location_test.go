package interp

import (
	"testing"
)

// The call-stack introspection built-ins are position-sensitive, so each case
// pins both the value and the line it is written on. The expected shapes are
// recorded in testdata/fixtures/FunctionsDebug.
func TestSourceCodeLocationBuiltins(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{
			name: "current location in main program",
			source: `
				var loc := CurrentSourceCodeLocation;
				PrintLn(loc.File);
				PrintLn(loc.Line);
				PrintLn('[' + loc.Name + ']');
			`,
			expected: "*MainModule*\n2\n[]\n",
		},
		{
			name: "current location names the enclosing routine",
			source: `
				function Test : TSourceCodeLocation;
				begin
					Result := CurrentSourceCodeLocation;
				end;
				var loc := Test;
				PrintLn(loc.Line);
				PrintLn(loc.Name);
			`,
			expected: "4\nTest\n",
		},
		{
			name: "caller location is empty in the main program",
			source: `
				var loc := CallerSourceCodeLocation;
				PrintLn('[' + loc.File + ']');
				PrintLn(loc.Line);
				PrintLn('[' + loc.Name + ']');
			`,
			expected: "[]\n0\n[]\n",
		},
		{
			name: "caller location points at the call site",
			source: `
				function Test : TSourceCodeLocation;
				begin
					Result := CallerSourceCodeLocation;
				end;
				var loc := Test;
				PrintLn(loc.Line);
				PrintLn('[' + loc.Name + ']');
			`,
			expected: "6\n[]\n",
		},
		{
			name: "caller location names the calling routine",
			source: `
				function Test : TSourceCodeLocation;
				begin
					Result := CallerSourceCodeLocation;
				end;
				function SubTest : TSourceCodeLocation;
				begin
					Result := Test;
				end;
				var loc := SubTest;
				PrintLn(loc.Line);
				PrintLn(loc.Name);
			`,
			expected: "8\nSubTest\n",
		},
		{
			name: "stack trace lists every live call site",
			source: `procedure Test2;
begin
   PrintLn(CurrentStackTrace);
end;

procedure Test;
begin
   Test2;
end;

Test;
`,
			expected: "Test2 [line: 3, column: 12]\nTest [line: 8, column: 4]\n [line: 11, column: 1]\n",
		},
		{
			name:     "stack trace at program level has one line",
			source:   "PrintLn(CurrentStackTrace);",
			expected: " [line: 1, column: 9]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output := testEvalWithOutput(tt.source)
			if isError(result) {
				t.Fatalf("unexpected error: %v", result)
			}
			if output != tt.expected {
				t.Errorf("output:\n%q\nwant:\n%q", output, tt.expected)
			}
		})
	}
}
