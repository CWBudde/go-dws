package interp

import "testing"

func TestBareFunctionReceiver_WriteOnlyProperty(t *testing.T) {
	for _, receiver := range []string{"Make", "Make()"} {
		t.Run(receiver, func(t *testing.T) {
			assertOutput(t, runQuickwinScript(t, `
var calls := 0;
type TTarget = class
  procedure SetValue(value: Integer);
  begin PrintLn(value); end;
  property Prop: Integer write SetValue;
end;
var target := TTarget.Create;
function Make: TTarget;
begin calls += 1; Result := target; end;
`+receiver+`.Prop := 42;
PrintLn(calls);
`), "42\n1\n")
		})
	}
}

func TestBareFunctionReceiver_PropertyAccessValidation(t *testing.T) {
	for _, receiver := range []string{"Make", "Make()"} {
		for _, test := range []struct{ operation, want string }{
			{"PrintLn(" + receiver + ".OnlyWrite);", "Cannot read a write only property"},
			{receiver + ".OnlyWrite += 1;", "Cannot read a write only property"},
			{receiver + ".OnlyRead := 1;", "Cannot set a value for a read-only property"},
		} {
			t.Run(test.operation, func(t *testing.T) {
				assertCompileError(t, `
type TTarget = class
  Value: Integer;
  procedure SetValue(value: Integer); begin end;
  property OnlyWrite: Integer write SetValue;
  property OnlyRead: Integer read Value;
end;
function Make: TTarget;
begin Result := TTarget.Create; end;
`+test.operation, test.want)
			})
		}
	}
}
