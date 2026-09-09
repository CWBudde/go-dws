package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

func TestFunctionPointerEntry_Builtin(t *testing.T) {
	var output bytes.Buffer
	engine := New(&output)
	pointer := &FunctionPointerValue{BuiltinName: "pRiNtLn"}
	result := engine.EvalFunctionPointer(pointer, []Value{&StringValue{Value: "direct"}})
	if isError(result) {
		t.Fatalf("direct call: %s", result.String())
	}
	_, err := engine.callDWScriptFunction(pointer, []any{"callback"})
	if err != nil {
		t.Fatalf("host callback: %v", err)
	}
	if got := output.String(); got != "direct\ncallback\n" {
		t.Fatalf("output = %q", got)
	}
}

func TestFunctionPointerEntry_Invalid(t *testing.T) {
	engine := New(&bytes.Buffer{})
	if result := engine.EvalFunctionPointer(&IntegerValue{Value: 1}, nil); !isError(result) {
		t.Fatalf("non-callable result = %v", result)
	}
}

func TestFunctionPointerEntry_CapturedClosure(t *testing.T) {
	var output bytes.Buffer
	engine := New(&output)
	p := parser.New(lexer.New(`
 var captured := 40;
 var callback := lambda(value: Integer): Integer begin
   captured += value;
   Result := captured;
 end;
 `))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("parse: %v", errs)
	}
	if result := engine.Eval(program); isError(result) {
		t.Fatalf("declarations: %s", result.String())
	}
	value, ok := engine.Env().Get("callback")
	if !ok {
		t.Fatal("callback missing")
	}
	pointer, ok := value.(*FunctionPointerValue)
	if !ok {
		t.Fatalf("callback type = %T", value)
	}
	for _, want := range []int64{42, 44} {
		got, err := engine.callDWScriptFunction(pointer, []any{int64(2)})
		if err != nil {
			t.Fatalf("callback: %v", err)
		}
		if got != want {
			t.Fatalf("callback result = %v, want %d", got, want)
		}
	}
	captured, _ := engine.Env().Get("captured")
	if captured.(*IntegerValue).Value != 44 {
		t.Fatalf("captured = %v", captured)
	}
}

func TestInterpreterDoesNotImplementBuiltinContext(t *testing.T) {
	if _, ok := any(New(&bytes.Buffer{})).(builtins.Context); ok {
		t.Fatal("interpreter shell must delegate builtin dispatch to the evaluator")
	}
}

func TestFunctionPointerEntry_CapturedBeforeMethodImplementation(t *testing.T) {
	var output bytes.Buffer
	engine := New(&output)
	p := parser.New(lexer.New(`
 type TSample = class
  function Compute(value: Integer): Integer;
  function Answer: Integer;
 end;
 var instance := TSample.Create;
 var callback := @instance.Compute;
 var parameterless := @instance.Answer;
 function TSample.Compute(value: Integer): Integer;
 begin
  Result := value + 1;
 end;
 function TSample.Answer: Integer;
 begin
  Result := 42;
 end;
 PrintLn(callback(41));
 PrintLn(parameterless());
 `))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if result := engine.Eval(program); isError(result) {
		t.Fatalf("evaluation: %s", result.String())
	}
	if got := output.String(); got != "42\n42\n" {
		t.Fatalf("output = %q", got)
	}
	for _, name := range []string{"callback", "parameterless"} {
		captured, ok := engine.Env().Get(name)
		if !ok {
			t.Fatalf("missing %s", name)
		}
		pointer, ok := captured.(*FunctionPointerValue)
		if !ok || pointer.Callable == nil || pointer.Callable.Declaration.Body == nil {
			t.Fatalf("%s did not retain the canonical callable: %v", name, captured)
		}
	}
}
