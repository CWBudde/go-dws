package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestParse_RoutineDirectiveRecovery(t *testing.T) {
	for _, tt := range []struct {
		source, rejected string
		column           int
	}{
		{"procedure P; inline; export; begin Bad; end;", "export", 22},
		{"procedure P; inline; cdecl; begin end;", "cdecl", 22},
		{"procedure P; cdecl; cdecl; begin Bad; end;", "cdecl", 21},
		{"procedure P;", "", 12},
	} {
		t.Run(tt.source, func(t *testing.T) {
			p := New(lexer.New(tt.source))
			fn := p.parseFunctionDeclaration()
			errs := p.Errors()
			if len(errs) != 1 || !errs[0].Stop || errs[0].Message != "BEGIN expected" || errs[0].Pos.Column != tt.column {
				t.Fatalf("errors: %+v", errs)
			}
			if fn == nil || fn.Name.Value != "P" || fn.IsExport || fn.Body != nil || !fn.BodyMissingBegin {
				t.Fatalf("reached header: %+v", fn)
			}
			if p.cursor.Peek(1).Literal != tt.rejected {
				t.Fatalf("rejected token consumed: %v", p.cursor.Peek(1))
			}
		})
	}
}

func TestParse_RoutineRecoveryRetainsLocalPrefix(t *testing.T) {
	p := New(lexer.New("procedure P; const X = 1; var Y: Integer := X; export; begin Bad; end; Later;"))
	program := p.ParseProgram()
	if len(program.Statements) != 1 {
		t.Fatalf("reached statements: %v", program)
	}
	fn, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok || !fn.BodyMissingBegin || fn.Body == nil || !fn.Body.Truncated || len(fn.Body.Statements) != 2 {
		t.Fatalf("reached local prefix: %+v", program.Statements[0])
	}
	if _, ok := fn.Body.Statements[0].(*ast.ConstDecl); !ok {
		t.Fatalf("const lost: %v", fn.Body)
	}
	if _, ok := fn.Body.Statements[1].(*ast.VarDeclStatement); !ok {
		t.Fatalf("var lost: %v", fn.Body)
	}
	if len(p.Errors()) != 1 || !p.Errors()[0].Stop {
		t.Fatalf("stop: %v", p.Errors())
	}
	direct := New(lexer.New("procedure P; const X = 1; var Y: Integer := X; export; begin Bad; end; Later;"))
	if fn := direct.parseFunctionDeclaration(); fn == nil || direct.cursor.Peek(1).Type != lexer.EXPORT {
		t.Fatalf("direct cursor: %v", direct.cursor.Peek(1))
	}
}

func TestParse_InterfaceDirectiveRecovery(t *testing.T) {
	p := New(lexer.New("type T = interface procedure P; export; end; Later;"))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 1 || !errs[0].Stop || errs[0].Message != "END expected" || errs[0].Pos.Column != 33 {
		t.Fatalf("errors: %v", errs)
	}
	if len(program.Statements) != 1 {
		t.Fatalf("reached statements: %v", program)
	}
	iface, ok := program.Statements[0].(*ast.InterfaceDecl)
	if !ok || len(iface.Methods) != 1 || iface.Methods[0].Name.Value != "P" {
		t.Fatalf("reached interface: %v", program)
	}
	direct := New(lexer.New("interface procedure P; export; end; Later;"))
	if iface := direct.parseInterfaceDeclarationBody(&ast.Identifier{Value: "T"}); iface == nil || direct.cursor.Current().Type != lexer.EXPORT {
		t.Fatalf("direct cursor: %v", direct.cursor.Current())
	}
}

func TestParse_RoutineContractEOFRecovery(t *testing.T) {
	for _, tt := range []struct {
		source       string
		stop         bool
		locals       int
		line, column int
	}{
		{"procedure P; require True;", true, 0, 1, 26},
		{"procedure P;\nrequire True;\nconst X = 1;", true, 1, 3, 12},
		{"procedure P;\nrequire True;\nvar Y: Integer := 1;", true, 1, 3, 20},
		{"procedure P;\nrequire True;\nconst X = 1;\nvar Y: Integer := X;", true, 2, 4, 20},
		{"procedure Test(i : Integer);\nrequire\n   i>0", false, 0, 3, 6},
	} {
		t.Run(tt.source, func(t *testing.T) {
			p := New(lexer.New(tt.source))
			fn := p.parseFunctionDeclaration()
			if fn == nil || fn.PreConditions == nil || fn.BodyMissingBegin != tt.stop {
				t.Fatalf("contract carrier: %+v", fn)
			}
			checkRoutineContractEOFPrefix(t, fn, tt.locals)
			errs := p.Errors()
			want := "BEGIN expected"
			if !tt.stop {
				want = `";" expected`
			}
			if len(errs) != 1 || errs[0].Message != want || errs[0].Stop != tt.stop || errs[0].Pos.Line != tt.line || errs[0].Pos.Column != tt.column {
				t.Fatalf("errors: %+v", errs)
			}
			if p.cursor.Peek(1).Type != lexer.EOF {
				t.Fatalf("EOF cursor: %v", p.cursor.Peek(1))
			}
		})
	}
}

func checkRoutineContractEOFPrefix(t *testing.T, fn *ast.FunctionDecl, locals int) {
	t.Helper()
	if locals == 0 {
		if fn.Body != nil {
			t.Fatalf("unexpected locals: %v", fn.Body)
		}
		return
	}
	if fn.Body == nil || !fn.Body.Truncated || len(fn.Body.Statements) != locals {
		t.Fatalf("reached locals: %v", fn.Body)
	}
}
