package evaluator

import (
	"reflect"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestRecordIndexDeclarations_AllBuilders(t *testing.T) {
	for _, form := range []string{"named", "inline", "anonymous"} {
		for _, checked := range []bool{false, true} {
			t.Run(form+map[bool]string{false: "/unchecked", true: "/checked"}[checked], func(t *testing.T) {
				program, info := parseRecordIndexMetadata(t, form, checked)
				e := NewEvaluator(interptypes.NewTypeSystem(), nil, nil, nil, info, runtime.NewRefCountManager())
				ctx := runtime.NewExecutionContext(runtime.NewEnvironment())
				if value := e.Eval(program, ctx); isError(value) {
					t.Fatal(value)
				}
				value, ok := ctx.Env().Get("r")
				if !ok {
					t.Fatal("r missing")
				}
				assertRecordIndexSignature(t, value.(*runtime.RecordValue).RecordType, form != "anonymous")
				// Source mutations cannot change registered signatures.
				params := recordIndexMetadataParameters(program, form)
				params[0].Name.Value = "Changed"
				params[0].ByRef = false
				params[2].IsConst = false
				assertRecordIndexSignature(t, value.(*runtime.RecordValue).RecordType, form != "anonymous")
			})
		}
	}
}

func parseRecordIndexMetadata(t *testing.T, form string, checked bool) (*ast.Program, *ast.SemanticInfo) {
	t.Helper()
	body := `class function Get(var x,y: Integer; const z: Integer; w: Integer): Integer; begin Result := 0; end;
 class property P[var a,b: Integer; const c: Integer; d: Integer]: Integer read Get; default;`
	source := "type TR = record " + body + " end; var r: TR;"
	if form == "inline" {
		source = "var r: record " + body + " end;"
	}
	if form == "anonymous" {
		source = "var r := record " + strings.ReplaceAll(body, "class ", "") + " end;"
	}
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatal(p.Errors())
	}
	if !checked {
		return program, nil
	}
	a := semantic.NewAnalyzer()
	if err := a.Analyze(program); err != nil {
		t.Fatal(err)
	}
	info := a.GetSemanticInfo()
	assertRecordIndexSignature(t, recordIndexMetadataType(program, info, a, form), form != "anonymous")
	return program, info
}

func recordIndexMetadataType(program *ast.Program, info *ast.SemanticInfo, a *semantic.Analyzer, form string) *types.RecordType {
	if form == "named" {
		return a.GetRecords()["tr"]
	}
	decl := program.Statements[0].(*ast.VarDeclStatement)
	node := ast.Node(decl.Type)
	if form == "anonymous" {
		node = decl.Value
	}
	typ, _ := info.GetResolvedType(node).(*types.RecordType)
	return typ
}

func recordIndexMetadataParameters(program *ast.Program, form string) []*ast.Parameter {
	if form == "named" {
		return program.Statements[0].(*ast.RecordDecl).Properties[0].IndexParams
	}
	decl := program.Statements[0].(*ast.VarDeclStatement)
	if form == "inline" {
		return decl.Type.(*ast.RecordTypeNode).Properties[0].IndexParams
	}
	return decl.Value.(*ast.AnonymousRecordExpression).Properties[0].IndexParams
}

func assertRecordIndexSignature(t *testing.T, typ *types.RecordType, class bool) {
	t.Helper()
	if typ == nil {
		t.Fatal("record type missing")
	}
	prop := typ.GetProperty("P")
	if prop == nil {
		t.Fatal("property missing")
	}
	if !prop.IsIndexed || !prop.IsDefault || prop.IsClassProperty != class {
		t.Fatalf("lost property flags: %+v", prop)
	}
	if !reflect.DeepEqual(prop.IndexParamTypes, []types.Type{types.INTEGER, types.INTEGER, types.INTEGER, types.INTEGER}) {
		t.Fatalf("types: %v", prop.IndexParamTypes)
	}
	if !reflect.DeepEqual(prop.IndexParamNames, []string{"a", "b", "c", "d"}) {
		t.Fatalf("names: %v", prop.IndexParamNames)
	}
	if !reflect.DeepEqual(prop.IndexParamModes, []types.PropertyIndexMode{types.PropertyIndexVar, types.PropertyIndexVar, types.PropertyIndexConst, types.PropertyIndexValue}) {
		t.Fatalf("modes: %v", prop.IndexParamModes)
	}
}

func TestRecordIndexDeclarations_UncheckedUnknownType(t *testing.T) {
	for _, source := range []string{
		"type TR = record property P[a: Missing]: Integer read (1); end;",
		"var r: record property P[a: Missing]: Integer read (1); end;",
		"var r := record property P[a: Missing]: Integer read (1); end;",
	} {
		p := parser.New(lexer.New(source))
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Fatal(p.Errors())
		}
		e := NewEvaluator(interptypes.NewTypeSystem(), nil, nil, nil, nil, runtime.NewRefCountManager())
		ctx := runtime.NewExecutionContext(runtime.NewEnvironment())
		value := e.Eval(program, ctx)
		if !isError(value) || !strings.Contains(value.String(), "record property index parameter 'a'") {
			t.Fatalf("source %s: unknown index type accepted: %v", source, value)
		}
	}
}
