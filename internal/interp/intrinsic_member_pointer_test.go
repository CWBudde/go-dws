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
