package interp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJSONGlobalStorage_Strings(t *testing.T) {
	for _, tt := range []struct{ name, value, serialized string }{
		{"plain", "hello", `"hello"`},
		{"empty", "", `""`},
		{"escaped and Unicode", "quote\" apostrophe' slash\\ newline\n tab\t café 世界", `"quote\" apostrophe' slash\\ newline\n tab\t café \u4E16\u754C"`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			literal := strings.ReplaceAll(tt.serialized, "'", "''")
			source := `
var value := JSON.Parse('` + literal + `');
PrintLn(JSON.Stringify(value));
WriteGlobalVar('e6.string', value);
PrintLn(VarIsStr(ReadGlobalVar('e6.string')));
PrintLn('[' + String(ReadGlobalVar('e6.string')) + ']');
PrintLn(JSON.Stringify(value));
DeleteGlobalVar('e6.string');
`
			want := tt.serialized + "\nTrue\n[" + tt.value + "]\n" + tt.serialized + "\n"
			assertOutput(t, runQuickwinScript(t, source), want)
		})
	}
}

func TestJSONGlobalStorage_SharedConsumers(t *testing.T) {
	assertOutput(t, runQuickwinScript(t, `
CleanupGlobalQueues('e6.queue');
GlobalQueuePush('e6.queue', JSON.Parse('"hello"'));
GlobalQueueInsert('e6.queue', JSON.Parse('""'));
var value: Variant;
while GlobalQueuePull('e6.queue', value) do begin
  PrintLn(VarIsStr(value));
  PrintLn('[' + String(value) + ']');
end;
CleanupGlobalQueues('e6.queue');

WriteGlobalVar('e6.exchange', 'hello');
PrintLn(CompareExchangeGlobalVar('e6.exchange', JSON.Parse('"world"'), JSON.Parse('"hello"')));
PrintLn(ReadGlobalVar('e6.exchange'));
PrintLn(CompareExchangeGlobalVar('e6.exchange', JSON.Parse('""'), 'world'));
PrintLn('[' + String(ReadGlobalVar('e6.exchange')) + ']');
DeleteGlobalVar('e6.exchange');
`), "True\n[]\nTrue\n[hello]\nhello\nworld\nworld\n[]\n")
}

func TestJSONGlobalStorage_OtherKinds(t *testing.T) {
	for _, value := range []string{`["hello"]`, `{"a":"hello","b":[123,456]}`, `1`, `1.25`, `true`, `false`, `null`} {
		t.Run(value, func(t *testing.T) {
			source := `
var value := JSON.Parse('` + value + `');
WriteGlobalVar('e6.other', value);
PrintLn(VarIsStr(ReadGlobalVar('e6.other')));
PrintLn(ReadGlobalVar('e6.other'));
PrintLn(JSON.Stringify(value));
DeleteGlobalVar('e6.other');
`
			assertOutput(t, runQuickwinScript(t, source), "True\n"+value+"\n"+value+"\n")
		})
	}
}

func TestJSONGlobalStorage_Fixtures(t *testing.T) {
	for _, fixture := range []string{"FunctionsGlobalVars/write_json"} {
		t.Run(fixture, func(t *testing.T) {
			base := filepath.Join("..", "..", "testdata", "fixtures", fixture)
			source, err := os.ReadFile(base + ".pas")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(base + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			assertOutput(t, strings.TrimSpace(runQuickwinScript(t, string(source))), strings.TrimSpace(string(want)))
		})
	}
}
