package semantic

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func unreachableWarnings(a *Analyzer) []string {
	var result []string
	for _, message := range a.Errors() {
		if strings.HasPrefix(message, "Warning: Unreachable code") {
			result = append(result, message)
		}
	}
	return result
}

func TestUnreachableWarnings_Fixtures(t *testing.T) {
	for _, name := range []string{"FailureScripts/unreachable", "FailureScripts/unreachable_case_of", "FailureScripts/break_continue", "FailureScripts/class_cast", "FailureScripts/raise_error", "SimpleScripts/exit", "SimpleScripts/exceptions2"} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile("../../testdata/fixtures/" + name + ".pas")
			if err != nil {
				t.Fatal(err)
			}
			golden, err := os.ReadFile("../../testdata/fixtures/" + name + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			var want []string
			for _, line := range strings.Split(strings.ReplaceAll(string(golden), "\r\n", "\n"), "\n") {
				if strings.HasPrefix(line, "Warning: Unreachable code") {
					want = append(want, line)
				}
			}
			got := unreachableWarnings(parseAndAnalyze(t, string(source)))
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("warnings\ngot: %v\nwant: %v", got, want)
			}
		})
	}
}

func TestUnreachableWarnings_ControlFlow(t *testing.T) {
	for _, test := range []struct {
		name, body string
		count      int
	}{
		{"once per list", "exit; PrintLn(1); exit; PrintLn(2);", 1},
		{"nested list", "exit; begin exit; PrintLn(1); end; PrintLn(2);", 2},
		{"if without else", "if True then exit; PrintLn(1);", 0},
		{"if with else", "if Random > 0 then exit else exit; PrintLn(1);", 1},
		{"case without else", "case Random(2) of 1: exit; end; PrintLn(1);", 0},
		{"while true exit", "while True do exit; PrintLn(1);", 1},
		{"while variable exit", "while Random > 0 do exit; PrintLn(1);", 0},
		{"repeat exit", "repeat exit until True; PrintLn(1);", 1},
		{"repeat break", "repeat break until False; PrintLn(1);", 0},
		{"mixed loop interruption", "repeat if Random > 0 then break else exit until False; PrintLn(1);", 0},
		{"for exit", "for var i := 1 to 2 do exit; PrintLn(1);", 0},
		{"try exit", "try exit finally PrintLn(1); end; PrintLn(2);", 0},
		{"invalid continue", "continue; PrintLn(1);", 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			a := parseAndAnalyze(t, "procedure Test; begin "+test.body+" end;")
			if got := unreachableWarnings(a); len(got) != test.count {
				t.Fatalf("got %v, want %d warnings", got, test.count)
			}
		})
	}
}

func TestUnreachableWarnings_PrecedeTargetErrorsAndLoopWarning(t *testing.T) {
	a := parseAndAnalyze(t, "procedure Test; begin while True do begin continue; Missing; end; end;")
	var relevant []string
	for _, message := range a.Errors() {
		if strings.Contains(message, "Unreachable code") || strings.Contains(message, "Missing") || strings.Contains(message, "Infinite loop") {
			relevant = append(relevant, message)
		}
	}
	if len(relevant) != 3 || !strings.Contains(relevant[0], "Unreachable code") || !strings.Contains(relevant[1], "Missing") || !strings.Contains(relevant[2], "Infinite loop") {
		t.Fatalf("unexpected emission order: %v", relevant)
	}
}
