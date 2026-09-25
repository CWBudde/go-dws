package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

func TestFunctionPointerParameterModifiers(t *testing.T) {
	for _, modifier := range []string{"const", "var", "lazy"} {
		for _, named := range []bool{false, true} {
			for _, matching := range []bool{false, true} {
				t.Run(modifier+"/"+map[bool]string{true: "named", false: "inline"}[named]+"/"+map[bool]string{true: "match", false: "mismatch"}[matching], func(t *testing.T) {
					targetModifier := ""
					if matching {
						targetModifier = modifier + " "
					}
					pointer := "procedure(" + targetModifier + "x: Integer)"
					prefix := ""
					if named {
						prefix = "type TCallback = " + pointer + "; "
						pointer = "TCallback"
					}
					source := prefix + "procedure Test(" + modifier + " x: Integer); begin end; var p: " + pointer + " := @Test;"
					p := parser.New(lexer.New(source))
					program := p.ParseProgram()
					if len(p.Errors()) != 0 {
						t.Fatal(p.Errors())
					}
					a := NewAnalyzer()
					err := a.Analyze(program)
					if matching && err != nil {
						t.Fatalf("matching modifiers rejected: %v", err)
					}
					if !matching && err == nil {
						t.Fatal("mismatched modifiers accepted")
					}
				})
			}
		}
	}
}

func TestFunctionPointerInferredMethodModifiers(t *testing.T) {
	for _, addressOf := range []string{"", "@"} {
		t.Run(addressOf+"method", func(t *testing.T) {
			source := `type TTest = class
     procedure Test(const x: Integer); begin end;
    end;
    var obj := TTest.Create;
    var inferred := ` + addressOf + `obj.Test;
    var p: procedure(x: Integer) := inferred;`
			p := parser.New(lexer.New(source))
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatal(p.Errors())
			}
			a := NewAnalyzer()
			if err := a.Analyze(program); err == nil || !strings.Contains(err.Error(), "Incompatible types") || !strings.Contains(err.Error(), "const Integer") {
				t.Fatalf("expected mismatch retaining const parameter, got %v", err)
			}
		})
	}
}

func TestFunctionPointerInterfaceModifiers(t *testing.T) {
	for _, modifier := range []string{"const", "var", "lazy"} {
		t.Run(modifier, func(t *testing.T) {
			source := `type ITest = interface procedure Take(` + modifier + ` x: Integer); end;
    var obj: ITest;
    var p: procedure(` + modifier + ` x: Integer) := obj.Take;`
			p := parser.New(lexer.New(source))
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatal(p.Errors())
			}
			a := NewAnalyzer()
			if err := a.Analyze(program); err != nil {
				t.Fatalf("matching interface signature rejected: %v", err)
			}
		})
	}
}
