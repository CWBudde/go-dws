package ast_test

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/dwscript"
)

func TestFunctionDecl_ExportStringRoundTrip(t *testing.T) {
	for _, directive := range []string{"export", "export ''", "export 'PublicName'", "export 'can''t'", "export \"first\nsecond \"\"quoted\"\"\"", "overload; export; cdecl; inline; deprecated 'it''s old'", "export; helper", "export 'PublicName'; helper Double; cdecl; inline; deprecated 'old'", "forward; export 'PublicName'", "external 'host'; export 'can''t'"} {
		t.Run(directive, func(t *testing.T) {
			source := "procedure P; " + directive + ";"
			if directive != "forward; export 'PublicName'" && directive != "external 'host'; export 'can''t'" {
				source += " begin end;"
			}
			p := parser.New(lexer.New(source))
			tree := p.ParseProgram()
			if errs := p.Errors(); len(errs) != 0 {
				t.Fatalf("parse: %v", errs)
			}
			original := tree.Statements[0].(*ast.FunctionDecl)
			rendered := original.String() + ";"
			p = parser.New(lexer.New(rendered))
			tree = p.ParseProgram()
			if errs := p.Errors(); len(errs) != 0 {
				t.Fatalf("String reparse: %v\n%s", errs, rendered)
			}
			got := tree.Statements[0].(*ast.FunctionDecl)
			if !got.IsExport || exportStringMetadataOf(got) != exportStringMetadataOf(original) {
				t.Fatalf("metadata lost: %+v -> %+v\n%s", original, got, rendered)
			}
		})
	}
}

type exportStringMetadata struct {
	externalName, convention, deprecation, helperName                    string
	export, named, external, forward, body, overload, inline, deprecated bool
	helper, namedHelper                                                  bool
}

func exportStringMetadataOf(fn *ast.FunctionDecl) exportStringMetadata {
	helperName := ""
	if fn.HelperName != nil {
		helperName = fn.HelperName.Value
	}
	return exportStringMetadata{
		externalName: fn.ExternalName, convention: fn.CallingConvention, deprecation: fn.DeprecatedMessage,
		export: fn.IsExport, named: fn.HasExportName, external: fn.IsExternal, forward: fn.IsForward,
		body: fn.Body != nil, overload: fn.IsOverload, inline: fn.IsInline, deprecated: fn.IsDeprecated,
		helper: fn.IsHelper, namedHelper: fn.HelperName != nil, helperName: helperName,
	}
}

func TestFunctionDecl_ExportHelperStringExecution(t *testing.T) {
	for _, tt := range []struct{ directive, method string }{
		{"export; helper", "Twice"}, {"export 'PublicTwice'; helper Double; cdecl; inline; deprecated 'old'", "Double"},
	} {
		t.Run(tt.directive, func(t *testing.T) {
			source := "function Twice(X: Integer): Integer; " + tt.directive + "; begin Result := X*2; end;"
			p := parser.New(lexer.New(source))
			tree := p.ParseProgram()
			if errs := p.Errors(); len(errs) != 0 {
				t.Fatalf("parse: %v", errs)
			}
			fn := tree.Statements[0].(*ast.FunctionDecl)
			rendered := fn.String() + "; var X := 5; PrintLn(X." + tt.method + "());"
			var output bytes.Buffer
			engine, err := dwscript.New(dwscript.WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := engine.Eval(source + " var X := 5; PrintLn(X." + tt.method + "());"); err != nil {
				t.Fatalf("original helper execution: %v", err)
			}
			if got := output.String(); got != "10\n" {
				t.Fatalf("original helper output %q", got)
			}
			output.Reset()
			if _, err := engine.Eval(rendered); err != nil {
				t.Fatalf("String helper execution: %v\n%s", err, rendered)
			}
			if got := output.String(); got != "10\n" {
				t.Fatalf("String helper output %q\n%s", got, rendered)
			}
		})
	}
}
