package cmd

import (
	"regexp"
	"strings"
	"testing"
)

func TestRun_PlainDiagnosticsAreWireFormat(t *testing.T) {
	out, err := captureRun(t, "var x: Integer := 'hello';", nil, func() { diagnosticsMode = "plain" })
	if err == nil {
		t.Fatal("expected a compile failure")
	}
	if strings.Contains(out, "\x1b[") || strings.Contains(out, "Compilation failed") {
		t.Fatalf("plain mode must not decorate: %q", out)
	}
	if !regexp.MustCompile(`(?m)^Syntax Error: .* \[line: 1, column: \d+\]$`).MatchString(out) {
		t.Fatalf("expected wire format, got %q", out)
	}
}

func TestRun_PlainRuntimeError(t *testing.T) {
	out, err := captureRun(t, "var a := 1 div 0;", nil, func() { diagnosticsMode = "plain" })
	if err == nil {
		t.Fatal("expected a runtime failure")
	}
	// The message text itself ("division by zero: 1 div 0" vs upstream's "Division by
	// zero") is a runtime-parity gap tracked in PLAN.md §3.3; here only the wire shape
	// matters: one line, "Runtime Error: " prefix, bracketed position, no "ERROR:" noise.
	line := strings.TrimSpace(out)
	if strings.Count(line, "\n") != 0 || !strings.HasPrefix(line, "Runtime Error: ") ||
		strings.Contains(line, "ERROR:") || !strings.HasSuffix(line, "[line: 1, column: 12]") {
		t.Fatalf("got %q", out)
	}
}

func TestRun_PrettyHonorsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out, _ := captureRun(t, "var x: Integer := 'hello';", nil, func() { diagnosticsMode = "pretty" })
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("NO_COLOR ignored: %q", out)
	}
	if !strings.Contains(out, "Error in <eval>:1:1") {
		t.Fatalf("expected the pretty block, got %q", out)
	}
}

func TestRun_RejectsUnknownDiagnosticsMode(t *testing.T) {
	_, err := captureRun(t, "PrintLn(1);", nil, func() { diagnosticsMode = "fancy" })
	if err == nil || !strings.Contains(err.Error(), "--diagnostics") {
		t.Fatalf("expected a flag validation error, got %v", err)
	}
}
