package printer_test

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/dwscript"
	"github.com/cwbudde/go-dws/pkg/printer"
)

func TestPrint_ExportRoundTrip(t *testing.T) {
	for _, directive := range []string{"export", "export 'PublicName'", "export ''", "export 'can''t'", "export \"first\nsecond \"\"quoted\"\"\"", "overload; export; inline", "export; cdecl; inline; deprecated 'it''s old'"} {
		t.Run(directive, func(t *testing.T) {
			p := parser.New(lexer.New("procedure P; " + directive + "; begin PrintLn('P'); end; function F: Integer; export 'PublicF'; begin Result := 42; end; P; PrintLn(F());"))
			tree := p.ParseProgram()
			if errs := p.Errors(); len(errs) != 0 {
				t.Fatalf("parse: %v", errs)
			}
			rendered := printer.New(printer.DefaultOptions()).Print(tree)
			var output bytes.Buffer
			engine, err := dwscript.New(dwscript.WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(rendered)
			if err != nil {
				t.Fatalf("compile: %v\n%s", err, rendered)
			}
			original := tree.Statements[0].(*ast.FunctionDecl)
			roundTrip := program.AST().Statements[0].(*ast.FunctionDecl)
			assertExportRoundTrip(t, original, roundTrip, rendered)
			assertExportRoundTrip(t, tree.Statements[1].(*ast.FunctionDecl), program.AST().Statements[1].(*ast.FunctionDecl), rendered)
			if _, err := engine.Run(program); err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != "P\n42\n" {
				t.Fatalf("output %q\n%s", got, rendered)
			}
		})
	}
}

func TestPrint_ExportLinkageRoundTrip(t *testing.T) {
	for _, source := range []string{
		"procedure P; forward; export 'PublicName'; procedure P; begin end; P;",
		"procedure P; external 'host'; export;",
		"procedure P; external 'host'; export '';",
		"procedure P; external 'host'; export 'can''t'; cdecl; inline; deprecated 'old';",
	} {
		t.Run(source, func(t *testing.T) {
			p := parser.New(lexer.New(source))
			tree := p.ParseProgram()
			if errs := p.Errors(); len(errs) != 0 {
				t.Fatalf("parse: %v", errs)
			}
			rendered := printer.New(printer.DefaultOptions()).Print(tree)
			engine, err := dwscript.New()
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(rendered)
			if err != nil {
				t.Fatalf("compile: %v\n%s", err, rendered)
			}
			assertExportRoundTrip(t, tree.Statements[0].(*ast.FunctionDecl), program.AST().Statements[0].(*ast.FunctionDecl), rendered)
		})
	}
}

func TestPrint_ExportUnitInterfaceRoundTrip(t *testing.T) {
	const source = "unit U; interface function F: Integer; export ''; implementation function F: Integer; begin Result := 42; end; end."
	p := parser.New(lexer.New(source))
	tree := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("parse: %v", errs)
	}
	rendered := printer.New(printer.DefaultOptions()).Print(tree)
	p = parser.New(lexer.New(rendered))
	roundTrip := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("reparse: %v\n%s", errs, rendered)
	}
	original := tree.Statements[0].(*ast.UnitDeclaration).InterfaceSection.Statements[0].(*ast.FunctionDecl)
	got := roundTrip.Statements[0].(*ast.UnitDeclaration).InterfaceSection.Statements[0].(*ast.FunctionDecl)
	assertExportRoundTrip(t, original, got, rendered)
	impl := roundTrip.Statements[0].(*ast.UnitDeclaration).ImplementationSection.Statements[0].(*ast.FunctionDecl)
	if impl.IsExport || impl.Body == nil {
		t.Fatalf("implementation changed: %+v\n%s", impl, rendered)
	}
}

func assertExportRoundTrip(t *testing.T, original, got *ast.FunctionDecl, rendered string) {
	t.Helper()
	if !got.IsExport || exportMetadataOf(got) != exportMetadataOf(original) {
		t.Fatalf("routine metadata changed: original %+v got %+v\n%s", original, got, rendered)
	}
}

type exportMetadata struct {
	externalName, convention, deprecation, helperName                                         string
	export, named, external, forward, body, overload, inline, deprecated, helper, namedHelper bool
}

func exportMetadataOf(fn *ast.FunctionDecl) exportMetadata {
	helperName := ""
	if fn.HelperName != nil {
		helperName = fn.HelperName.Value
	}
	return exportMetadata{
		externalName: fn.ExternalName, convention: fn.CallingConvention, deprecation: fn.DeprecatedMessage,
		export: fn.IsExport, named: fn.HasExportName, external: fn.IsExternal, forward: fn.IsForward,
		body: fn.Body != nil, overload: fn.IsOverload, inline: fn.IsInline, deprecated: fn.IsDeprecated,
		helper: fn.IsHelper, namedHelper: fn.HelperName != nil, helperName: helperName,
	}
}

func TestPrint_ExportHelperRoundTrip(t *testing.T) {
	for _, tt := range []struct{ directive, method string }{
		{"export; helper", "Twice"}, {"export 'PublicTwice'; helper Double; cdecl; inline; deprecated 'old'", "Double"},
	} {
		t.Run(tt.directive, func(t *testing.T) {
			source := "function Twice(X: Integer): Integer; " + tt.directive + "; begin Result := X*2; end; var X := 5; PrintLn(X." + tt.method + "());"
			p := parser.New(lexer.New(source))
			tree := p.ParseProgram()
			if errs := p.Errors(); len(errs) != 0 {
				t.Fatalf("parse: %v", errs)
			}
			rendered := printer.New(printer.DefaultOptions()).Print(tree)
			var output bytes.Buffer
			engine, err := dwscript.New(dwscript.WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := engine.Eval(source); err != nil {
				t.Fatalf("original helper execution: %v", err)
			}
			if got := output.String(); got != "10\n" {
				t.Fatalf("original output %q", got)
			}
			output.Reset()
			program, err := engine.Compile(rendered)
			if err != nil {
				t.Fatalf("printed helper compile: %v\n%s", err, rendered)
			}
			assertExportRoundTrip(t, tree.Statements[0].(*ast.FunctionDecl), program.AST().Statements[0].(*ast.FunctionDecl), rendered)
			if _, err := engine.Run(program); err != nil {
				t.Fatalf("printed helper execution: %v", err)
			}
			if got := output.String(); got != "10\n" {
				t.Fatalf("printed output %q\n%s", got, rendered)
			}
		})
	}
}

func TestPrint_ExportCombinedLinkageUsesSourceOrder(t *testing.T) {
	const source = "procedure P; external 'host'; forward; export;"
	p := parser.New(lexer.New(source))
	tree := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("parse: %v", errs)
	}
	rendered := printer.New(printer.DefaultOptions()).Print(tree)
	// Pinned ReadProcDecl consumes external, then forward, then export exactly
	// once. The relaxed Go parser alone cannot validate that source grammar.
	if rendered != source {
		t.Fatalf("invalid linkage phase order: %q; want %q", rendered, source)
	}
	p = parser.New(lexer.New(rendered))
	roundTrip := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("reparse: %v", errs)
	}
	assertExportRoundTrip(t, tree.Statements[0].(*ast.FunctionDecl), roundTrip.Statements[0].(*ast.FunctionDecl), rendered)
}
