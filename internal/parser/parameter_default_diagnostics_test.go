package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestParameterDefaults_ModifierDiagnosticsRecover(t *testing.T) {
	for _, tt := range []struct {
		modifier string
		message  string
		column   int
	}{
		{"lazy", "lazy parameter cannot have a default value", 35},
		{"var", "var parameter cannot have a default value", 34},
		{"const", "const parameter cannot have a default value", 36},
	} {
		t.Run(tt.modifier, func(t *testing.T) {
			input := "procedure Dummy1(" + tt.modifier + " a : Integer = 1 + 2; b: Integer); external;\nprocedure Next; external;"
			p := testParser(input)
			program := p.ParseProgram()
			if len(p.errors) != 1 {
				t.Fatalf("errors = %v, want one modifier diagnostic", p.errors)
			}
			err := p.errors[0]
			if err.Message != tt.message || err.Pos.Line != 1 || err.Pos.Column != tt.column || err.Stop {
				t.Errorf("diagnostic = %+v, want ordinary %q at 1:%d", err, tt.message, tt.column)
			}
			if len(program.Statements) != 2 {
				t.Fatalf("statements = %d, want both declarations", len(program.Statements))
			}
			fn := requireBodylessExternalRoutine(t, program.Statements[0], "Dummy1")
			if len(fn.Parameters) != 2 {
				t.Fatalf("parameters = %d, want 2", len(fn.Parameters))
			}
			param := fn.Parameters[0]
			if param.DefaultValueSeparatorPos.Line != 1 || param.DefaultValueSeparatorPos.Column != tt.column {
				t.Errorf("default separator = %v, want 1:%d", param.DefaultValueSeparatorPos, tt.column)
			}
			assertParameterModifier(t, param, tt.modifier)
			if _, ok := param.DefaultValue.(*ast.BinaryExpression); !ok {
				t.Errorf("default = %T, want parsed binary expression", param.DefaultValue)
			}
			requireBodylessExternalRoutine(t, program.Statements[1], "Next")
		})
	}
}

func TestParameterGroup_ModifierDefaultDiagnosticsRecover(t *testing.T) {
	for _, tt := range []struct {
		modifier string
		message  string
		column   int
	}{
		{"lazy", "lazy parameter cannot have a default value", 17},
		{"var", "var parameter cannot have a default value", 16},
		{"const", "const parameter cannot have a default value", 18},
	} {
		t.Run(tt.modifier, func(t *testing.T) {
			p := newTestParser(tt.modifier + " a: Integer = 1 + 2; b: Integer")
			params := p.ParameterGroup(ParameterGroupConfig{AllowModifiers: true, AllowDefaults: true})
			if len(p.errors) != 1 {
				t.Fatalf("errors = %v, want one modifier diagnostic", p.errors)
			}
			err := p.errors[0]
			if err.Message != tt.message || err.Pos.Line != 1 || err.Pos.Column != tt.column || err.Stop {
				t.Errorf("diagnostic = %+v, want ordinary %q at 1:%d", err, tt.message, tt.column)
			}
			if len(params) != 1 {
				t.Fatalf("parameters = %d, want retained parameter", len(params))
			}
			if params[0].DefaultValueSeparatorPos.Line != 1 || params[0].DefaultValueSeparatorPos.Column != tt.column {
				t.Errorf("default separator = %v, want 1:%d", params[0].DefaultValueSeparatorPos, tt.column)
			}
			if _, ok := params[0].DefaultValue.(*ast.BinaryExpression); !ok {
				t.Errorf("default = %T, want parsed binary expression", params[0].DefaultValue)
			}
			if p.cursor.Current().Literal != "2" || p.cursor.Peek(1).Type != lexer.SEMICOLON {
				t.Errorf("cursor = %v, want last default token before semicolon", p.cursor.Current())
			}
		})
	}
}

func TestParameterDefaults_SeparatorPositionAcrossLines(t *testing.T) {
	input := "procedure Test(a, b: Integer = // initializer follows\n  1; c: Integer); external;"
	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	fn, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok || len(fn.Parameters) != 3 {
		t.Fatalf("declaration = %+v, want three parameters", program.Statements[0])
	}
	for _, param := range fn.Parameters[:2] {
		if param.DefaultValueSeparatorPos.Line != 1 || param.DefaultValueSeparatorPos.Column != 30 {
			t.Errorf("parameter %s separator = %v, want 1:30", param.Name.Value, param.DefaultValueSeparatorPos)
		}
		if param.DefaultValue.Pos().Line != 2 {
			t.Errorf("parameter %s initializer line = %d, want 2", param.Name.Value, param.DefaultValue.Pos().Line)
		}
	}
	if fn.Parameters[2].DefaultValueSeparatorPos.Line != 0 {
		t.Errorf("required parameter separator = %v, want no position", fn.Parameters[2].DefaultValueSeparatorPos)
	}

	p = newTestParser("a, b: Integer = // initializer follows\n  1")
	params := p.ParameterGroup(ParameterGroupConfig{AllowDefaults: true})
	checkParserErrors(t, p)
	if len(params) != 2 {
		t.Fatalf("combinator parameters = %d, want 2", len(params))
	}
	for _, param := range params {
		if param.DefaultValueSeparatorPos.Line != 1 || param.DefaultValueSeparatorPos.Column != 15 {
			t.Errorf("combinator parameter %s separator = %v, want 1:15", param.Name.Value, param.DefaultValueSeparatorPos)
		}
		if param.DefaultValue.Pos().Line != 2 {
			t.Errorf("combinator parameter %s initializer line = %d, want 2", param.Name.Value, param.DefaultValue.Pos().Line)
		}
	}
}

func TestExternalRoutine_FollowingGlobalConstantAndParameterStop(t *testing.T) {
	input := "procedure Dummy5(a : Float = Random); external;\n" +
		"const f : Float = 1.5;\n" +
		"procedure Dummy6(a : Integer = f); external;\n" +
		"procedure Dummy7(a : Integer"
	p := testParser(input)
	program := p.ParseProgram()
	if len(p.errors) != 1 {
		t.Fatalf("errors = %v, want final missing parenthesis", p.errors)
	}
	err := p.errors[0]
	if err.Message != `")" expected` || err.Pos.Line != 4 || err.Pos.Column != 22 || !err.Stop {
		t.Errorf("diagnostic = %+v, want missing ')' stop at 4:22", err)
	}
	if len(program.Statements) != 3 {
		t.Fatalf("statements = %d, want two external routines and a global constant", len(program.Statements))
	}
	requireBodylessExternalRoutine(t, program.Statements[0], "Dummy5")
	constant, ok := program.Statements[1].(*ast.ConstDecl)
	if !ok || constant.Name.Value != "f" {
		t.Fatalf("second statement = %+v, want global constant f", program.Statements[1])
	}
	next := requireBodylessExternalRoutine(t, program.Statements[2], "Dummy6")
	if len(next.Parameters) != 1 {
		t.Fatalf("Dummy6 parameters = %d, want 1", len(next.Parameters))
	}
	defaultValue, ok := next.Parameters[0].DefaultValue.(*ast.Identifier)
	if !ok || defaultValue.Value != "f" {
		t.Errorf("default = %+v, want identifier f", next.Parameters[0].DefaultValue)
	}
}

func requireBodylessExternalRoutine(t *testing.T, statement ast.Statement, name string) *ast.FunctionDecl {
	t.Helper()
	fn, ok := statement.(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("statement = %T, want external routine %s", statement, name)
	}
	if fn.Name.Value != name || !fn.IsExternal || fn.Body != nil {
		t.Fatalf("routine = %+v, want bodyless external %s", fn, name)
	}
	return fn
}

func assertParameterModifier(t *testing.T, param *ast.Parameter, modifier string) {
	t.Helper()
	if param.IsLazy != (modifier == "lazy") || param.ByRef != (modifier == "var") || param.IsConst != (modifier == "const") {
		t.Errorf("parameter lost its %s modifier", modifier)
	}
}
