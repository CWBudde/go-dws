package bytecode

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestCompiler_AnalyzedClassIsRejected(t *testing.T) {
	for _, tt := range []struct{ name, declarations, rhs string }{
		{"variable", "var Meta: TClass;", "Meta"},
		{"group", "var Meta: TClass;", "(Meta)"},
		{"qualified type", "", "System.TObject"},
		{"qualified value", "type THolder = record Meta: TClass; end; var Holder: THolder;", "Holder.Meta"},
		{"factory", "function Target: TClass; begin end;", "Target()"},
		{"grouped factory", "function Target: TClass; begin end;", "(Target())"},
	} {
		for _, metadata := range []bool{false, true} {
			t.Run(tt.name+map[bool]string{false: "/without metadata", true: "/with metadata"}[metadata], func(t *testing.T) {
				p := parser.New(lexer.New(tt.declarations + " var Obj: TObject; PrintLn(Obj is " + tt.rhs + ");"))
				program := p.ParseProgram()
				if len(p.Errors()) != 0 {
					t.Fatal(p.Errors())
				}
				a := semantic.NewAnalyzer()
				if err := a.Analyze(program); err != nil {
					t.Fatal(err)
				}
				compiler := NewCompiler("test")
				if metadata {
					compiler.SetSemanticInfo(a.GetSemanticInfo())
				}
				_, err := compiler.Compile(program)
				if err == nil || !strings.Contains(err.Error(), "type checking with 'is' operator not yet supported in bytecode mode") {
					t.Fatalf("wanted explicit unsupported IS, got %v", err)
				}
			})
		}
	}
}

func TestCompiler_BooleanIsStillSupported(t *testing.T) {
	for _, analyze := range []bool{false, true} {
		p := parser.New(lexer.New("PrintLn(false is False); PrintLn(true is False); PrintLn(true is (1 = 1));"))
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			t.Fatal(p.Errors())
		}
		compiler := NewCompiler("test")
		if analyze {
			a := semantic.NewAnalyzer()
			if err := a.Analyze(program); err != nil {
				t.Fatal(err)
			}
			compiler.SetSemanticInfo(a.GetSemanticInfo())
		}
		chunk, err := compiler.Compile(program)
		if err != nil {
			t.Fatal(err)
		}
		var output bytes.Buffer
		if _, err := NewVMWithOutput(&output).Run(chunk); err != nil {
			t.Fatal(err)
		}
		if got := output.String(); got != "true\nfalse\ntrue\n" {
			t.Fatalf("output = %q", got)
		}
	}
}

func TestCompiler_BooleanImplicitIsRejected(t *testing.T) {
	for _, tt := range []struct{ name, declarations, rhs string }{
		{"routine", "function Check: Boolean; begin Result := false; end;", "Check"},
		{"pointer", "type TCheck = function: Boolean; function Check: Boolean; begin Result := false; end; var P: TCheck := @Check;", "P"},
		{"grouped pointer", "type TCheck = function: Boolean; function Check: Boolean; begin Result := false; end; var P: TCheck := @Check;", "(P)"},
		{"method pointer", "type TCheck = function: Boolean of object; type TChecker = class function Check: Boolean; begin Result := false; end; end; var Obj: TChecker; var P: TCheck := @Obj.Check;", "P"},
		{"grouped method pointer", "type TCheck = function: Boolean of object; type TChecker = class function Check: Boolean; begin Result := false; end; end; var Obj: TChecker; var P: TCheck := @Obj.Check;", "(P)"},
		{"bound method", "type TChecker = class function Check: Boolean; begin Result := false; end; end; var Obj: TChecker;", "Obj.Check"},
		{"grouped bound method", "type TChecker = class function Check: Boolean; begin Result := false; end; end; var Obj: TChecker;", "(Obj.Check)"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.declarations + " PrintLn(false is " + tt.rhs + ");"))
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatal(p.Errors())
			}
			a := semantic.NewAnalyzer()
			if err := a.Analyze(program); err != nil {
				t.Fatal(err)
			}
			compiler := NewCompiler("test")
			compiler.SetSemanticInfo(a.GetSemanticInfo())
			// Compile the analyzed IS node with real declared symbol types. Avoid
			// unrelated VM gaps in pointer type declarations and routine bodies.
			for _, name := range []string{"Check", "P", "Obj"} {
				if symbol, found := a.GetSymbolTable().Resolve(name); found {
					if _, err := compiler.declareGlobal(&ast.Identifier{Value: name}, symbol.Type); err != nil {
						t.Fatal(err)
					}
				}
			}
			var check *ast.IsExpression
			ast.Inspect(program, func(node ast.Node) bool {
				if expr, ok := node.(*ast.IsExpression); ok {
					check = expr
				}
				return true
			})
			if check == nil {
				t.Fatal("missing IS expression")
			}
			err := compiler.compileIsExpression(check)
			if err == nil || !strings.Contains(err.Error(), "implicit calls in 'is' operator not yet supported in bytecode mode") {
				t.Fatalf("wanted unsupported implicit IS call, got %v", err)
			}
		})
	}
}

func TestCompiler_ClassMethodImplicitIsRejected(t *testing.T) {
	p := parser.New(lexer.New("type TChecker = class class function Check: Boolean; begin Result := false; end; end; PrintLn(false is TChecker.Check);"))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatal(p.Errors())
	}
	a := semantic.NewAnalyzer()
	if err := a.Analyze(program); err != nil {
		t.Fatal(err)
	}
	compiler := NewCompiler("test")
	compiler.SetSemanticInfo(a.GetSemanticInfo())
	_, err := compiler.Compile(program)
	if err == nil || !strings.Contains(err.Error(), "implicit calls in 'is' operator not yet supported in bytecode mode") {
		t.Fatalf("wanted unsupported implicit class-method call, got %v", err)
	}
}
