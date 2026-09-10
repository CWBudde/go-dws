package interp

import (
	"strings"
	"testing"
)

// TestIntrinsicMemberPointer_Runtime exercises pointers formed from TObject's
// intrinsic parameterless members (ClassName, ClassType): capturing them with
// `@`, coercing them into a parameterless-function element type, and invoking
// them both explicitly and by auto-invoke in a value position.
func TestIntrinsicMemberPointer_Runtime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "address-of class type auto-invokes in receiver position",
			input: `
				var proc := @TObject.ClassType;
				PrintLn(proc.ClassName);
			`,
			expected: []string{"TObject"},
		},
		{
			name: "address-of class name is invoked explicitly",
			input: `
				type TThing = class(TObject);
				var proc := @TThing.ClassName;
				PrintLn(proc());
			`,
			expected: []string{"TThing"},
		},
		{
			name: "address-of class name on an instance",
			input: `
				type TThing = class(TObject);
				var obj := TThing.Create;
				var proc := @obj.ClassName;
				PrintLn(proc());
			`,
			expected: []string{"TThing"},
		},
		{
			name: "class type coerces into an array of parameterless functions",
			input: `
				type TClassA = class(TObject);
				var a : array of function : TClass;
				a.Add(TObject.ClassType);
				a.Add(TClassA.ClassType);
				var i : Integer;
				for i:=0 to High(a) do
					PrintLn(a[i]().ClassName);
			`,
			expected: []string{"TObject", "TClassA"},
		},
		{
			name: "class name coerces and auto-invokes on read",
			input: `
				type TClassB = class end;
				var a : array of function : String;
				a.Add(TObject.ClassName);
				a.Add(TClassB.ClassName);
				var i : Integer;
				for i:=0 to High(a) do
					PrintLn(a[i]);
			`,
			expected: []string{"TObject", "TClassB"},
		},
		{
			name: "a user-declared class method still wins over the intrinsic",
			input: `
				type TThing = class
					class function ClassName : String;
				end;
				class function TThing.ClassName : String;
				begin
					Result := 'overridden';
				end;
				var a : array of function : String;
				a.Add(TThing.ClassName);
				PrintLn(a[0]);
			`,
			expected: []string{"overridden"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testMethodPointerCase(tt.input, t)
			for _, want := range tt.expected {
				if !strings.Contains(output, want) {
					t.Errorf("expected output containing %q, got %q", want, output)
				}
			}
		})
	}
}

// TestIntrinsicMemberPointer_TypedBinding covers the pointer contexts that do
// not go through the plain member-read path: typed declarations, assignments to
// a declared function-pointer variable, and instance receivers. It also pins the
// precedence rule that a user-declared method of the same name owns the
// reference, and that the analyzer and the runtime agree on which overload a
// class-method pointer binds to.
func TestIntrinsicMemberPointer_TypedBinding(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "typed declaration initializer captures the intrinsic",
			input: `
				type TThing = class(TObject);
				var p : function : String := TThing.ClassName;
				PrintLn(p());
			`,
			expected: []string{"TThing"},
		},
		{
			name: "assignment to a declared pointer captures the intrinsic",
			input: `
				type TThing = class(TObject);
				var p : function : TClass;
				p := TThing.ClassType;
				PrintLn(p().ClassName);
			`,
			expected: []string{"TThing"},
		},
		{
			name: "assignment from an instance receiver captures the intrinsic",
			input: `
				type TThing = class(TObject);
				var obj := TThing.Create;
				var p : function : String;
				p := obj.ClassName;
				PrintLn(p());
			`,
			expected: []string{"TThing"},
		},
		{
			name: "instance receiver coerces into an array of parameterless functions",
			input: `
				type TThing = class(TObject);
				var obj := TThing.Create;
				var a : array of function : String;
				a.Add(obj.ClassName);
				PrintLn(a[0]);
			`,
			expected: []string{"TThing"},
		},
		{
			name: "a user class method wins over the intrinsic on an instance",
			input: `
				type TThing = class
					class function ClassName : String;
				end;
				class function TThing.ClassName : String;
				begin
					Result := 'overridden';
				end;
				var obj := TThing.Create;
				var p := @obj.ClassName;
				PrintLn(p());
			`,
			expected: []string{"overridden"},
		},
		{
			name: "a parameterless overload declared first is the one the pointer binds",
			input: `
				type TThing = class
					class function Make : String; overload;
					class function Make(const s : String) : String; overload;
				end;
				class function TThing.Make : String;
				begin
					Result := 'none';
				end;
				class function TThing.Make(const s : String) : String;
				begin
					Result := s;
				end;
				var p := @TThing.Make;
				PrintLn(p());
			`,
			expected: []string{"none"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testMethodPointerCase(tt.input, t)
			for _, want := range tt.expected {
				if !strings.Contains(output, want) {
					t.Errorf("expected output containing %q, got %q", want, output)
				}
			}
		})
	}
}
