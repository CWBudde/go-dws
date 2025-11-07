package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

func TestExitStatementWithValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		source string
		want   string
	}{
		{
			name: "FunctionExitReturnsBoolean",
			source: `
				function Test: Boolean;
				begin
					Exit False;
				end;

				PrintLn(Test());
			`,
			want: "false\n",
		},
		{
			name: "ConditionalExitBranches",
			source: `
				function Choose(condition: Boolean): Integer;
				begin
					if condition then
					begin
						Exit(5);
					end
					else
					begin
						Exit(10);
					end;
				end;

				PrintLn(Choose(True));
				PrintLn(Choose(False));
			`,
			want: "5\n10\n",
		},
		{
			name: "ExitFromNestedBlocks",
			source: `
				function ExitFromNestedBlocks: Integer;
				var
					i: Integer;
					j: Integer;
				begin
					for i := 1 to 3 do
					begin
						j := 0;
						while j < 3 do
						begin
							j := j + 1;
							if (i = 2) and (j = 2) then
							begin
								Exit i * 100 + j;
							end;
						end;
					end;
					Exit -1;
				end;

				PrintLn(ExitFromNestedBlocks());
			`,
			want: "202\n",
		},
		{
			name: "ExitOverridesResultValue",
			source: `
				function ExitOverridesResultValue: Integer;
				begin
					Result := 50;
					Exit 123;
					Result := 0;
				end;

				PrintLn(ExitOverridesResultValue());
			`,
			want: "123\n",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := lexer.New(tc.source)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
			}

			var buf bytes.Buffer
			interp := New(&buf)
			interp.Eval(program)

			if got := buf.String(); got != tc.want {
				t.Errorf("expected output %q, got %q", tc.want, got)
			}
		})
	}
}
