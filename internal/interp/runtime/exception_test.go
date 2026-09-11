package runtime

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/lexer"
)

func at(line, column int) *lexer.Position {
	return &lexer.Position{Line: line, Column: column}
}

// TestExceptionValue_StackTraceString covers how the innermost frame of
// Exception.StackTrace is chosen: OriginPos (the construction site) when set,
// the reported raise position otherwise, and no extra frame when both are nil.
func TestExceptionValue_StackTraceString(t *testing.T) {
	callStack := errors.StackTrace{
		errors.NewStackFrame("ThisOneBombs", "", at(22, 4)),
	}

	tests := []struct {
		name string
		exc  *ExceptionValue
		want string
	}{
		{
			name: "origin position wins over the reported raise position",
			exc: &ExceptionValue{
				CallStack: callStack,
				Position:  at(3, 35),
				OriginPos: at(3, 20),
			},
			want: "ThisOneBombs [line: 3, column: 20]\n [line: 22, column: 4]",
		},
		{
			name: "runtime errors fall back to the reported position",
			exc: &ExceptionValue{
				CallStack: callStack,
				Position:  at(3, 12),
			},
			want: "ThisOneBombs [line: 3, column: 12]\n [line: 22, column: 4]",
		},
		{
			name: "no position contributes no innermost frame",
			exc: &ExceptionValue{
				CallStack: errors.StackTrace{
					errors.NewStackFrame("TestProc", "", at(26, 4)),
					errors.NewStackFrame("RequirePositive", "", at(15, 4)),
				},
			},
			want: "TestProc [line: 15, column: 4]\n [line: 26, column: 4]",
		},
		{
			name: "empty stack renders nothing",
			exc:  &ExceptionValue{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.exc.StackTraceString(); got != tt.want {
				t.Fatalf("StackTraceString() =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}
