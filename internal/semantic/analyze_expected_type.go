package semantic

import (
	"github.com/cwbudde/go-dws/pkg/token"
)

// DWScript names the type a context required and nothing else. It does not name
// the type it got, and it does not name the construct that wanted it, because
// the message is raised where the value is read rather than where the construct
// is assembled:
//
//	Syntax Error: Boolean expected [line: 1, column: 1]
//
// The anchor is the first token of the smallest syntactic unit that owns the
// value. Where that unit has an introducer of its own the introducer wins over
// the expression: `while` for a while loop (loop_nonbool, column 1, not the
// column 7 where the condition starts), `until` for a repeat loop (repeat2,
// column 8), `if` for the if-then-else expression (ifthenelse_expression1,
// column 11). Where it has none — a call argument, a contract clause — the unit
// begins at the expression, so the two coincide.
func (a *Analyzer) addBooleanExpected(pos token.Position) {
	a.addStructuredError(NewTypeExpectedError(pos, "Boolean"))
}

func (a *Analyzer) addStringExpected(pos token.Position) {
	a.addStructuredError(NewTypeExpectedError(pos, "String"))
}
