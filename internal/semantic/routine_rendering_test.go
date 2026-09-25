package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/types"
)

func TestSemanticRoutineRendering_RegistryCallback(t *testing.T) {
	a := NewAnalyzer()
	a.builtinRegistry = builtins.NewRegistry()
	sig := builtins.SigOptional([]types.Type{types.INTEGER, types.INTEGER, types.INTEGER}, types.VOID, 1).
		WithVarParams(0).WithArgCounts(1, 3)
	a.builtinRegistry.RegisterWithSignature("RegisteredOnly", nil, builtins.CategorySystem, "test", sig)
	pointer := a.resolveNamedFunctionPointerType("rEgIsTeReDoNlY")
	if pointer == nil {
		t.Fatal("registered callback was not resolved")
	}
	want := "procedure RegisteredOnly(var Integer, Integer, Integer)"
	if got := semanticTypeNameForDiagnostic(pointer); got != want {
		t.Errorf("callback signature = %q, want %q", got, want)
	}
	if pointer.MinArgs != 1 || !pointer.AcceptsArgCount(1) || pointer.AcceptsArgCount(2) || !pointer.AcceptsArgCount(3) {
		t.Errorf("callback arity lost: min = %d, allowed = %v", pointer.MinArgs, pointer.AllowedArgCounts)
	}
}

func TestSemanticRoutineRendering_ParameterLists(t *testing.T) {
	for _, tt := range []struct {
		name string
		fn   *types.FunctionPointerType
		want string
	}{
		{"", types.NewProcedurePointerType(nil), "procedure "},
		{"TMyProc", types.NewProcedurePointerType(nil), "procedure TMyProc"},
		{"", types.NewFunctionPointerType(nil, types.INTEGER), "function : Integer"},
		{"Read", types.NewFunctionPointerType(nil, types.INTEGER), "function Read: Integer"},
		{"", types.NewProcedurePointerType([]types.Type{types.STRING}), "procedure (String)"},
		{"IntToHex", types.NewFunctionPointerType([]types.Type{types.INTEGER, types.INTEGER}, types.STRING), "function IntToHex(Integer, Integer): String"},
	} {
		t.Run(tt.want, func(t *testing.T) {
			if got := semanticNamedFunctionPointerName(tt.name, tt.fn); got != tt.want {
				t.Errorf("signature = %q, want %q", got, tt.want)
			}
			if tt.name == "" {
				if got := semanticTypeNameForDiagnostic(tt.fn); got != tt.want {
					t.Errorf("type name = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestSemanticRoutineRendering_Metadata(t *testing.T) {
	for _, tt := range []struct {
		want string
		fn   types.FunctionPointerType
	}{
		{"class function ClassName: String", types.FunctionPointerType{Name: "ClassName", IsClassMethod: true, ReturnType: types.STRING}},
		{"class function ClassType: TClass", types.FunctionPointerType{Name: "ClassType", IsClassMethod: true, ReturnType: &types.ClassOfType{}, ReturnTypeName: "TClass"}},
		{"class procedure Reset", types.FunctionPointerType{Name: "Reset", IsClassMethod: true}},
		{"constructor Create", types.FunctionPointerType{Name: "Create", IsConstructor: true}},
		{"destructor Destroy", types.FunctionPointerType{Name: "Destroy", IsDestructor: true}},
		{"procedure Test(const String)", types.FunctionPointerType{Name: "Test", Parameters: []types.Type{types.STRING}, ConstParams: []bool{true}}},
		{"procedure Test(var Integer, const String, lazy Boolean)", types.FunctionPointerType{Name: "Test", Parameters: []types.Type{types.INTEGER, types.STRING, types.BOOLEAN}, VarParams: []bool{true}, ConstParams: []bool{false, true}, LazyParams: []bool{false, false, true}}},
	} {
		t.Run(tt.want, func(t *testing.T) {
			if got := semanticTypeNameForDiagnostic(&tt.fn); got != tt.want {
				t.Errorf("type name = %q, want %q", got, tt.want)
			}
			method := &types.MethodPointerType{FunctionPointerType: tt.fn, OfObject: true}
			if got := semanticTypeNameForDiagnostic(method); got != tt.want {
				t.Errorf("method type name = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSemanticRoutineRendering_FunctionSignature(t *testing.T) {
	fn := types.NewFunctionType([]types.Type{types.STRING, types.INTEGER, types.BOOLEAN}, types.VOID)
	fn.ConstParams = []bool{true}
	fn.VarParams = []bool{false, true}
	fn.LazyParams = []bool{false, false, true}
	want := "procedure Test(const String, var Integer, lazy Boolean)"
	if got := semanticNamedFunctionSignature("Test", fn); got != want {
		t.Errorf("signature = %q, want %q", got, want)
	}
	if got := semanticNamedFunctionPointerName("Test", functionPointerFromFunctionType(fn)); got != want {
		t.Errorf("converted signature = %q, want %q", got, want)
	}
	if got := semanticNamedFunctionSignature("Test", types.NewFunctionType(nil, types.VOID)); got != "procedure Test" {
		t.Errorf("parameterless signature = %q, want procedure Test", got)
	}
}

func TestSemanticRoutineRendering_DeclaredParameter(t *testing.T) {
	proc := types.NewProcedurePointerType(nil)
	proc.Name = "TProc"
	fn := types.NewFunctionType([]types.Type{proc}, types.VOID)
	fn.ParamTypeNames = []string{"TProc"}
	if got := semanticFunctionParamTypeName(fn, 0, proc); got != "procedure TProc" {
		t.Errorf("parameter type = %q, want procedure TProc", got)
	}
}
