package lexer

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/token"
)

func TestHintLevel_RestoredAfterSpeculativeRead(t *testing.T) {
	l := New("first {$HINTS OFF} second {$HINTS PEDANTIC} third {$HINTS ON} fourth")
	saved := l.SaveState()
	expected := []token.HintLevel{token.HintLevelDefault, token.HintLevelDisabled, token.HintLevelPedantic, token.HintLevelDefault}
	for attempt := 0; attempt < 2; attempt++ {
		for index, want := range expected {
			got := l.NextToken()
			if got.Pos.Hints != want {
				t.Fatalf("attempt %d token %d: hints %v, want %v", attempt, index, got.Pos.Hints, want)
			}
			if got.End().Hints != want {
				t.Fatalf("token End lost hint setting: %v", got)
			}
		}
		l.RestoreState(saved)
	}
}
