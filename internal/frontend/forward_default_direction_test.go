package frontend

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestCompile_ForwardDefaultDirectionDiagnostics(t *testing.T) {
	for _, level := range []semantic.HintsLevel{semantic.HintsLevelNormal, semantic.HintsLevelPedantic} {
		for _, tt := range []struct {
			name, source, want string
			stop               bool
		}{
			{"matched export", "procedure P(X: Integer = 1); overload; forward;\nprocedure P(X: Integer); export; begin end;", "Syntax Error: BEGIN expected [line: 2, column: 26]", true},
			{"matched export EOF", "procedure P(X: Integer = 1); overload; forward;\nprocedure P(X: Integer); export", "Syntax Error: BEGIN expected [line: 2, column: 26]", true},
			{"matched malformed export", "procedure P(X: Integer = 1); overload; forward;\nprocedure P(X: Integer); export 42; cdecl; begin Bad; end; {$ERROR 'late'} Later;", "Syntax Error: BEGIN expected [line: 2, column: 26]", true},
			{"reached hints cutoff", "procedure Before; cdecl; begin end;\nprocedure P(X: Integer = 1); overload; forward;\nprocedure P(X: String); overload; export; cdecl; begin end;\nprocedure P(X: Integer); export; cdecl; begin Bad; end; {$ERROR 'late'} Later;", "Hint: Call convention \"cdecl\" is not supported and ignored [line: 1, column: 19]\nHint: Call convention \"cdecl\" is not supported and ignored [line: 3, column: 43]\nSyntax Error: BEGIN expected [line: 4, column: 26]", true},
			{"reverse addition keeps original pending", "procedure P(X: Integer); overload; forward;\nprocedure P(X: Integer = 1); overload; export; begin end;", "Syntax Error: Overload of \"P\" will be ambiguous with a previously declared version [line: 2, column: 1]\nSyntax Error: The function \"P\" was forward declared but not implemented [line: 1, column: 11]", false},
			{"reverse addition original resolved later", "procedure P(X: Integer); overload; forward;\nprocedure P(X: Integer = 1); overload; export; begin end;\nprocedure P(X: Integer); begin end;", "Syntax Error: Overload of \"P\" will be ambiguous with a previously declared version [line: 2, column: 1]", false},
			{"reverse addition missing directive", "procedure P(X: Integer); overload; forward;\nprocedure P(X: Integer = 1); export; begin end;", "Syntax Error: Overloaded procedure \"P\" must be marked with the \"overload\" directive [line: 2, column: 1]\nSyntax Error: The function \"P\" was forward declared but not implemented [line: 1, column: 11]", false},
			{"unmarked different type", "procedure P(X: Integer = 1); overload; forward;\nprocedure P(X: String); export; begin end;\nprocedure P(X: Integer); begin end;", "Syntax Error: Overloaded procedure \"P\" must be marked with the \"overload\" directive [line: 2, column: 1]", false},
			{"default conflict original resolved later", "procedure P(X: Integer = 1); overload; forward;\nprocedure P(X: String = 's'); overload; export; begin end;\nprocedure P(X: Integer); begin end;", "Syntax Error: Overload of \"P\" will be ambiguous with a previously declared version [line: 2, column: 1]", false},
			{"another forward terminal", "procedure P(X: Integer = 1); overload; forward;\nprocedure P(X: Boolean); overload; forward;\nprocedure P(X: Integer); begin end;", "Syntax Error: The function \"P\" was forward declared but not implemented [line: 2, column: 11]", false},
			{"ordinary default omission policy", "procedure P(X: Integer = 1); forward;\nprocedure P(X: Integer); export; begin Bad; end; Later;", "Syntax Error: implementation signature for 'P' does not match forward declaration [line: 2, column: 1]\nSyntax Error: BEGIN expected [line: 2, column: 26]", true},
		} {
			t.Run(fmt.Sprintf("%s/%d", tt.name, level), func(t *testing.T) {
				result := CompileWithOptions(tt.source, Options{HintsLevel: level, DisableSymbolDictionaryDiagnostics: true})
				if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
					t.Fatalf("complete diagnostics:\n%s\nwant:\n%s", got, tt.want)
				}
				if len(result.Diagnostics) == 0 || result.Diagnostics[len(result.Diagnostics)-1].Stop != tt.stop || !result.HasFatalDiagnostics() {
					t.Fatalf("typed stop/fatal state: %+v", result.Diagnostics)
				}
			})
		}
	}
}

func TestCompile_ForwardDefaultDirectionSelectedState(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		implemented  bool
		count        int
	}{
		{"selected only", "procedure P(X: Integer = 1); overload; forward;\nprocedure P(X: Boolean); overload; forward;\nprocedure P(X: Integer); begin end;", true, 2},
		{"reverse rejected", "procedure P(X: Integer); overload; forward;\nprocedure P(X: Integer = 1); overload; export; begin end;", false, 1},
		{"new directive rejected", "procedure P(X: Integer = 1); overload; forward;\nprocedure P(X: String); export; begin end;", false, 1},
		{"ambiguous member rejected", "procedure P(X: Integer = 1); overload; forward;\nprocedure P(X: String = 's'); overload; export; begin end;", false, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(tt.source, Options{DisableSymbolDictionaryDiagnostics: true})
			symbol, ok := result.Analyzer.GetSymbolTable().Resolve("P")
			if !ok {
				t.Fatal("missing original declaration")
			}
			candidates := symbol.Overloads
			if !symbol.IsOverloadSet {
				candidates = []*semantic.Symbol{symbol}
			}
			if len(candidates) != tt.count || candidates[0].IsForward == tt.implemented {
				t.Fatalf("selected state: %+v", candidates)
			}
			if tt.count == 2 && !candidates[1].IsForward {
				t.Fatal("unselected Boolean forward consumed")
			}
			signature, ok := candidates[0].Type.(*types.FunctionType)
			if !ok {
				t.Fatalf("original signature: %T", candidates[0].Type)
			}
			original := result.Program.Statements[0].(*ast.FunctionDecl)
			if signature.DefaultValues[0] != original.Parameters[0].DefaultValue {
				t.Fatal("original default expression identity changed")
			}
			assertForwardDefaultDirectionBinding(t, result, original, signature, tt.implemented)
		})
	}
}

func assertForwardDefaultDirectionBinding(t *testing.T, result *Result, original *ast.FunctionDecl, signature *types.FunctionType, implemented bool) {
	t.Helper()
	implementationIndex := 1
	if implemented {
		implementationIndex = 2
	}
	implementation := result.Program.Statements[implementationIndex].(*ast.FunctionDecl)
	metadata := result.SemanticInfo.GetResolvedType(implementation)
	if !implemented && metadata != nil {
		t.Fatal("rejected candidate received successful binding metadata")
	}
	if implemented {
		bound, ok := metadata.(*types.FunctionType)
		if !ok || bound == signature || len(bound.DefaultValues) != 1 {
			t.Fatalf("bound declaration snapshot: %T", metadata)
		}
		expression, ok := bound.DefaultValues[0].(ast.Expression)
		if !ok || expression == original.Parameters[0].DefaultValue || expression.Pos() != original.Parameters[0].DefaultValue.Pos() {
			t.Fatalf("default snapshot provenance: %+v", expression)
		}
	}
}

func TestCompile_ForwardDefaultDirectionAcceptedDiagnostics(t *testing.T) {
	for _, level := range []semantic.HintsLevel{semantic.HintsLevelNormal, semantic.HintsLevelPedantic} {
		for _, source := range []string{
			"procedure P(X: Integer = 1); overload; forward; procedure P(X: Integer); begin PrintLn(X); end; P; P(7);",
			"procedure P(X: Integer = 1); overload; forward; procedure P(X: Integer); overload; begin PrintLn(X); end; P();",
			"procedure P(X: Integer = 1); overload; forward; procedure P(X: Boolean); overload; forward; procedure P(X: Integer); begin PrintLn(X); end; procedure P(X: Boolean); begin PrintLn(X); end; P; P(True);",
		} {
			result := CompileWithOptions(source, Options{HintsLevel: level, DisableSymbolDictionaryDiagnostics: true})
			if len(result.Diagnostics) != 0 || !result.SemanticSuccessful || result.HasFatalDiagnostics() {
				t.Fatalf("accepted complete diagnostics: %+v", result.Diagnostics)
			}
		}
	}
}
