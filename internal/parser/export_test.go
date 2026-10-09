package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestParse_ExportDeclarations(t *testing.T) {
	for _, tt := range []struct {
		directive, name string
		present         bool
	}{
		{"export", "", false}, {"export 'PublicName'", "PublicName", true}, {"ExPoRt ''", "", true}, {"export 'can''t'", "can't", true},
		{"export \"first\nsecond\"", "first\nsecond", true}, {"overload; export; inline", "", false}, {"export; deprecated 'old'", "", false},
	} {
		t.Run(tt.directive, func(t *testing.T) {
			p := New(lexer.New("procedure P; " + tt.directive + "; begin end; function F: Integer; export; begin Result := 42; end;"))
			program := p.ParseProgram()
			if errs := p.Errors(); len(errs) != 0 {
				t.Fatalf("parse: %v", errs)
			}
			if len(program.Statements) != 2 {
				t.Fatalf("lost following declaration: %v", program)
			}
			for _, stmt := range program.Statements {
				fn, ok := stmt.(*ast.FunctionDecl)
				if !ok || fn.Body == nil || fn.IsExternal || !fn.IsExport {
					t.Fatalf("lost ordinary body: %v", stmt)
				}
			}
			fn := program.Statements[0].(*ast.FunctionDecl)
			if fn.ExternalName != tt.name || fn.HasExportName != tt.present {
				t.Fatalf("export name %q present %v; want %q present %v", fn.ExternalName, fn.HasExportName, tt.name, tt.present)
			}
		})
	}
}

func TestParse_ExportKeywordAnchorPrecedesOptionalName(t *testing.T) {
	p := New(lexer.New("procedure P; {header}\n  ExPoRt {name}\n 'PublicName'; begin end;"))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("parse: %v", errs)
	}
	fn := program.Statements[0].(*ast.FunctionDecl)
	if fn.ExportPos.Line != 2 || fn.ExportPos.Column != 3 {
		t.Fatalf("export anchor is not the keyword: %+v", fn.ExportPos)
	}
	if fn.ExternalName != "PublicName" || !fn.HasExportName {
		t.Fatalf("optional name metadata lost: %+v", fn)
	}
}

func TestParse_ExportRejectedTokenStops(t *testing.T) {
	for _, name := range []string{"Foo", "42", "'name' extra"} {
		t.Run(name, func(t *testing.T) {
			p := New(lexer.New("procedure P; export " + name + "; begin Bad; end; Later;"))
			_ = p.parseFunctionDeclaration()
			errs := p.Errors()
			col, rejected := 21, name
			if name == "'name' extra" {
				col, rejected = 28, "extra"
			}
			if len(errs) != 2 || errs[0].Message != `";" expected` || errs[0].Stop || errs[1].Message != "BEGIN expected" || !errs[1].Stop {
				t.Fatalf("errors: %v", errs)
			}
			for _, e := range errs {
				if e.Pos.Line != 1 || e.Pos.Column != col {
					t.Fatalf("position: %+v", e)
				}
			}
			if errs[0].Code != ErrMissingSemicolon || errs[1].Code != ErrUnexpectedToken {
				t.Fatalf("error codes: %v", errs)
			}
			if got := p.cursor.Peek(1).Literal; got != rejected {
				t.Fatalf("rejected token consumed: %q; want %q", got, rejected)
			}
		})
	}
}

func TestParse_ExportEOFIsCompilerStop(t *testing.T) {
	for _, source := range []string{"procedure P; export;", "procedure P; export", "procedure P; export 'name'", "procedure P;\n export; {comment}\n // trailing\n"} {
		t.Run(source, func(t *testing.T) {
			p := New(lexer.New(source))
			_ = p.parseFunctionDeclaration()
			errs := p.Errors()
			if len(errs) == 0 || errs[len(errs)-1].Message != "BEGIN expected" || !errs[len(errs)-1].Stop || errs[len(errs)-1].Code != ErrUnexpectedToken {
				t.Fatalf("missing real EOF body stop: %v", errs)
			}
			if p.cursor.Peek(1).Type != lexer.EOF {
				t.Fatalf("EOF cursor changed: %v", p.cursor.Current())
			}
		})
	}
}

func TestParse_ExportDirectivePhaseStopsAtRejectedToken(t *testing.T) {
	for _, tt := range []struct {
		directives, rejected string
		column               int
	}{
		{"export; export", "export", 22}, {"export; external", "external", 22},
		{"export; forward", "forward", 22}, {"export; overload", "overload", 22},
	} {
		t.Run(tt.directives, func(t *testing.T) {
			p := New(lexer.New("procedure P; " + tt.directives + "; begin end; Later;"))
			_ = p.parseFunctionDeclaration()
			errs := p.Errors()
			if len(errs) != 1 || errs[0].Message != "BEGIN expected" || !errs[0].Stop || errs[0].Pos.Line != 1 || errs[0].Pos.Column != tt.column {
				t.Fatalf("directive phase errors: %v", errs)
			}
			if got := p.cursor.Peek(1).Literal; got != tt.rejected {
				t.Fatalf("rejected directive consumed: %q; want %q", got, tt.rejected)
			}
		})
	}
}

func TestParse_ExportAfterLaterPhaseIsNotDirective(t *testing.T) {
	for _, directives := range []string{"inline; export", "cdecl; export", "deprecated 'old'; export"} {
		t.Run(directives, func(t *testing.T) {
			p := New(lexer.New("procedure P; " + directives + "; begin end;"))
			fn := p.parseFunctionDeclaration()
			if fn == nil || fn.IsExport || fn.HasExportName {
				t.Fatalf("misplaced export was accepted as a directive: %+v", fn)
			}
			if errs := p.Errors(); len(errs) != 0 {
				t.Fatalf("ordinary bodyless header recovery changed: %v", errs)
			}
			if p.cursor.Peek(1).Type != lexer.EXPORT {
				t.Fatalf("misplaced EXPORT consumed: %v", p.cursor.Peek(1))
			}
		})
	}
}

func TestParse_ExportMethodContextsRemainRejected(t *testing.T) {
	for _, source := range []string{
		"type T = class procedure P; export; begin end; end;",
		"type T = class class procedure P; export; begin end; end;",
		"type T = record procedure P; export; begin end; end;",
		"type T = helper for Integer procedure P; export; begin end; end;",
		"type T = class procedure P; end; procedure T.P; export; begin end;",
		"method P; export; begin end;", "constructor P; export; begin end;",
		"type T = procedure; export;", "var R := record procedure P; export; begin end; end;",
	} {
		t.Run(source, func(t *testing.T) {
			p := New(lexer.New(source))
			program := p.ParseProgram()
			if len(p.Errors()) == 0 {
				t.Fatal("method export newly accepted")
			}
			ast.Inspect(program, func(node ast.Node) bool {
				if fn, ok := node.(*ast.FunctionDecl); ok && fn.IsExport {
					t.Fatalf("member acquired export metadata: %v", fn)
				}
				return true
			})
		})
	}
}

func TestParse_ExportUnitInterfaceIsForward(t *testing.T) {
	p := New(lexer.New("unit U; interface function F: Integer; export 'PublicF'; implementation function F: Integer; begin Result := 42; end; end."))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("parse: %v", errs)
	}
	u := program.Statements[0].(*ast.UnitDeclaration)
	fn := u.InterfaceSection.Statements[0].(*ast.FunctionDecl)
	if !fn.IsExport || !fn.HasExportName || fn.ExternalName != "PublicF" || !fn.IsForward || fn.IsExternal || fn.Body != nil {
		t.Fatalf("interface metadata: %+v", fn)
	}
	impl := u.ImplementationSection.Statements[0].(*ast.FunctionDecl)
	if impl.IsExport || impl.IsForward || impl.Body == nil {
		t.Fatalf("implementation metadata: %+v", impl)
	}
}

func TestParse_ExportUnitInterfaceDoesNotBypassStop(t *testing.T) {
	p := New(lexer.New("unit U; interface var X := ; procedure P; export; implementation end."))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 1 || !errs[0].Stop {
		t.Fatalf("earliest stop: %v", errs)
	}
	ast.Inspect(program, func(node ast.Node) bool {
		if fn, ok := node.(*ast.FunctionDecl); ok && fn.Name.Value == "P" {
			t.Fatal("unreached exported routine built after compiler stop")
		}
		return true
	})
}

func TestParse_ExportUnitInterfaceHoistsImplicitEnum(t *testing.T) {
	p := New(lexer.New("unit U; interface function F: set of (A, B); export; implementation end."))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("parse: %v", errs)
	}
	u := program.Statements[0].(*ast.UnitDeclaration)
	block, ok := u.InterfaceSection.Statements[0].(*ast.BlockStatement)
	if !ok || !block.SharesEnclosingScope || len(block.Statements) != 2 {
		t.Fatalf("implicit enum was not hoisted: %v", u.InterfaceSection)
	}
	if _, ok := block.Statements[0].(*ast.EnumDecl); !ok {
		t.Fatalf("missing implicit enum: %T", block.Statements[0])
	}
	fn := block.Statements[1].(*ast.FunctionDecl)
	if !fn.IsForward || !fn.IsExport || fn.Body != nil {
		t.Fatalf("exported interface metadata: %+v", fn)
	}
}

// The separate interface-method parser already discards unknown directives.
// Keep its method signature instead of introducing ordinary export metadata.
func TestParse_ExportInterfaceMethodPreservesSignature(t *testing.T) {
	p := New(lexer.New("type T = interface procedure P; export; end;"))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("preexisting interface recovery changed: %v", errs)
	}
	iface := program.Statements[0].(*ast.InterfaceDecl)
	if len(iface.Methods) != 1 || iface.Methods[0].String() != "procedure P" {
		t.Fatalf("interface signature: %v", iface)
	}
	ast.Inspect(program, func(node ast.Node) bool {
		if fn, ok := node.(*ast.FunctionDecl); ok && fn.IsExport {
			t.Fatal("ordinary export metadata in interface")
		}
		return true
	})
}
