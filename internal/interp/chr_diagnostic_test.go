package interp

import (
	"fmt"
	"testing"
)

func TestChr_InvalidCodepointDiagnostic(t *testing.T) {
	for _, code := range []int64{-1, 0x110000} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			source := fmt.Sprintf(`var code: Integer := %d;
try
  Chr(code);
except
  on e: Exception do PrintLn(e.Message);
end;
PrintLn('continued');`, code)
			want := fmt.Sprintf("Invalid codepoint: %d [line: 3, column: 3]\ncontinued\n", code)
			if got := runHelperScript(t, source); got != want {
				t.Fatalf("want %q, got %q", want, got)
			}
		})
	}
}

func TestChr_ValidCodepointBoundaries(t *testing.T) {
	for _, code := range []int64{0, 0xFFFF, 0x10000, 0x10FFFF} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			source := fmt.Sprintf("var code: Integer := %d; PrintLn(Chr(code));", code)
			want := string(rune(code)) + "\n"
			if got := runHelperScript(t, source); got != want {
				t.Fatalf("want %q, got %q", want, got)
			}
		})
	}
}
