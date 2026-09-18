package interp

import "testing"

func TestStringReplaceHelper(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"all matches", `PrintLn('aba aba'.Replace('aba', 'longer'));`, "longer longer\n"},
		{"nonoverlapping matches", `PrintLn('aaaaa'.Replace('aa', 'b'));`, "bba\n"},
		{"case sensitive contents", `PrintLn('aAa'.Replace('a', 'b'));`, "bAb\n"},
		{"no match", `PrintLn('abc'.Replace('z', 'x'));`, "abc\n"},
		{"empty source", `PrintLn(''.Replace('a', 'b'));`, "\n"},
		{"empty needle", `PrintLn('abc'.Replace('', 'x'));`, "abc\n"},
		{"empty replacement", `PrintLn('banana'.Replace('an', ''));`, "ba\n"},
		{"Unicode", `PrintLn('é🙂é'.Replace('é', '雪'));`, "雪🙂雪\n"},
		{"receiver unchanged", `var s := 'aba'; PrintLn(s.Replace('a', 'x')); PrintLn(s);`, "xbx\naba\n"},
		{"alias case and chaining", `type TText = String; var s: TText := 'aba'; PrintLn(s.rEpLaCe('a', 'x').Replace('x', 'z'));`, "zbz\n"},
		{"ignored result", `var s := 'aba'; s.Replace('a', 'x'); PrintLn(s);`, "aba\n"},
		{"ignored result evaluates once", `
var calls := 0;
function Receiver: String;
begin
  Inc(calls);
  PrintLn('receiver');
  Result := 'aba';
end;
function Needle: String;
begin
  Inc(calls);
  PrintLn('needle');
  Result := 'a';
end;
function Replacement: String;
begin
  Inc(calls);
  PrintLn('replacement');
  Result := 'x';
end;
Receiver().Replace(Needle(), Replacement());
PrintLn(calls);`, "receiver\nneedle\nreplacement\n3\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runHelperScript(t, tt.source); got != tt.want {
				t.Fatalf("output mismatch: want %q, got %q", tt.want, got)
			}
		})
	}
}
