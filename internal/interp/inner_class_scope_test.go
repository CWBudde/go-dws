package interp

import (
	"testing"
)

// A nested type shadows a global type of the same name inside the enclosing
// class's own methods, and only there.
func TestNestedClassShadowsGlobalClass(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{
			name: "new resolves the nested type inside a method",
			source: `
				type TOuter = class
					type TSub = class
						procedure Hello; begin PrintLn('inner') end;
					end;
					procedure Run;
					begin
						var s := new TSub;
						s.Hello;
					end;
				end;
				type TSub = class
					procedure Hello; begin PrintLn('outer') end;
				end;
				(new TOuter).Run;
			`,
			expected: "inner\n",
		},
		{
			name: "the global type still wins outside the class",
			source: `
				type TOuter = class
					type TSub = class
						procedure Hello; begin PrintLn('inner') end;
					end;
				end;
				type TSub = class
					procedure Hello; begin PrintLn('outer') end;
				end;
				(new TSub).Hello;
			`,
			expected: "outer\n",
		},
		{
			name: "a global type with no nested counterpart still resolves",
			source: `
				type TOther = class
					procedure Hello; begin PrintLn('global') end;
				end;
				type TOuter = class
					procedure Run;
					begin
						(new TOther).Hello;
					end;
				end;
				(new TOuter).Run;
			`,
			expected: "global\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output := testEvalWithOutput(tt.source)
			if isError(result) {
				t.Fatalf("unexpected error: %v", result)
			}
			if output != tt.expected {
				t.Errorf("output = %q; want %q", output, tt.expected)
			}
		})
	}
}
