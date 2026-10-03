package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/bytecode"
)

func TestCompileScript_ClassIsRejected(t *testing.T) {
	for _, rhs := range []string{"Meta", "(Meta)", "System.TObject", "Target()", "(Target())"} {
		t.Run(rhs, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "input.dws")
			source := "var Obj: TObject; var Meta: TClass; function Target: TClass; begin end; PrintLn(Obj is " + rhs + ");"
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			savedOutput, savedSkip, savedDisassemble, savedVerbose := outputFile, skipTypeCheck, disassemble, compileVerbose
			t.Cleanup(func() {
				outputFile, skipTypeCheck, disassemble, compileVerbose = savedOutput, savedSkip, savedDisassemble, savedVerbose
			})
			outputFile, skipTypeCheck, disassemble, compileVerbose = "", false, false, false
			err := compileScript(nil, []string{path})
			if err == nil || !strings.Contains(err.Error(), "type checking with 'is' operator not yet supported in bytecode mode") {
				t.Fatalf("wanted unsupported IS, got %v", err)
			}
			if _, err := os.Stat(strings.TrimSuffix(path, ".dws") + ".dwc"); !os.IsNotExist(err) {
				t.Fatalf("unexpected bytecode artifact: %v", err)
			}
		})
	}
}

func TestCompileScript_BooleanIsSupported(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.dws")
	if err := os.WriteFile(path, []byte("var B: Boolean := false; PrintLn(false is B); PrintLn(true is False); PrintLn(true is (1 = 1));"), 0600); err != nil {
		t.Fatal(err)
	}
	savedOutput, savedSkip, savedDisassemble, savedVerbose := outputFile, skipTypeCheck, disassemble, compileVerbose
	t.Cleanup(func() {
		outputFile, skipTypeCheck, disassemble, compileVerbose = savedOutput, savedSkip, savedDisassemble, savedVerbose
	})
	outputFile, skipTypeCheck, disassemble, compileVerbose = "", false, false, false
	if err := compileScript(nil, []string{path}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(strings.TrimSuffix(path, ".dws") + ".dwc")
	if err != nil {
		t.Fatal(err)
	}
	chunk, err := bytecode.NewSerializer().DeserializeChunk(data)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if _, err := bytecode.NewVMWithOutput(&output).Run(chunk); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "true\nfalse\ntrue\n" {
		t.Fatalf("output=%q", got)
	}
}

func TestRunBytecode_ClassIsRejected(t *testing.T) {
	output, err := captureRun(t, "var Obj: TObject; var Meta: TClass; PrintLn(Obj is Meta);", nil, func() { bytecodeMode = true })
	if err == nil || !strings.Contains(err.Error(), "type checking with 'is' operator not yet supported in bytecode mode") {
		t.Fatalf("output %q; error %v", output, err)
	}
}

func TestBytecodeCLI_BooleanIsExpressions(t *testing.T) {
	for _, tt := range []struct{ name, source string }{
		{"unary", "PrintLn(true is (not false));"},
		{"nested unary", "PrintLn(false is (not (not false)));"},
		{"cast call", "PrintLn(true is (Boolean(1)));"},
		{"builtin call", "PrintLn(true is (StrBeginsWith('abc', 'a')));"},

		{"unary call", "PrintLn(true is (not Boolean(0)));"},
		{"routine call", "function Check: Boolean; begin end; PrintLn(false is (Check()));"},
	} {
		for _, checked := range []bool{false, true} {
			mode := map[bool]string{false: "unchecked", true: "checked"}[checked]
			t.Run(tt.name+"/run/"+mode, func(t *testing.T) {
				output, err := captureRun(t, tt.source, nil, func() { bytecodeMode = true; typeCheck = checked })
				if err != nil || output != "true\n" {
					t.Fatalf("output %q; error %v", output, err)
				}
			})
			t.Run(tt.name+"/compile/"+mode, func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "input.dws")
				if err := os.WriteFile(path, []byte(tt.source), 0600); err != nil {
					t.Fatal(err)
				}
				savedOutput, savedSkip, savedDisassemble, savedVerbose := outputFile, skipTypeCheck, disassemble, compileVerbose
				t.Cleanup(func() {
					outputFile, skipTypeCheck, disassemble, compileVerbose = savedOutput, savedSkip, savedDisassemble, savedVerbose
				})
				outputFile, skipTypeCheck, disassemble, compileVerbose = "", !checked, false, false
				if err := compileScript(nil, []string{path}); err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(strings.TrimSuffix(path, ".dws") + ".dwc")
				if err != nil {
					t.Fatal(err)
				}
				chunk, err := bytecode.NewSerializer().DeserializeChunk(data)
				if err != nil {
					t.Fatal(err)
				}
				var output bytes.Buffer
				if _, err := bytecode.NewVMWithOutput(&output).Run(chunk); err != nil {
					t.Fatal(err)
				}
				if got := output.String(); got != "true\n" {
					t.Fatalf("output=%q", got)
				}
			})
		}
	}
}

func TestBytecodeCLI_BooleanBuiltinShadowRejected(t *testing.T) {
	for _, name := range []string{"SameText", "StrBeginsWith"} {
		t.Run(name, func(t *testing.T) {
			source := "function " + name + ": TClass; begin end; var Obj: TObject; PrintLn(Obj is (" + name + "()));"
			output, err := captureRun(t, source, nil, func() { bytecodeMode = true; typeCheck = false })
			want := "type checking with 'is' operator not yet supported in bytecode mode"
			if name == "StrBeginsWith" {
				want = "duplicate global variable"
			}
			if err == nil || !strings.Contains(err.Error(), want) || output != "" {
				t.Fatalf("class-returning shadow compiled: output %q; error %v", output, err)
			}
		})
	}
}

func TestBytecodeCLI_UncheckedImplicitRoutineIsRejected(t *testing.T) {
	output, err := captureRun(t, "function Check: Boolean; begin end; PrintLn(false is (Check));", nil, func() { bytecodeMode = true; typeCheck = false })
	if err == nil || !strings.Contains(err.Error(), "implicit calls in 'is' operator not yet supported in bytecode mode") || output != "" {
		t.Fatalf("implicit routine compiled: output %q; error %v", output, err)
	}
}
