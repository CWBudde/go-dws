package dwscript

import (
	"bytes"
	"testing"
)

func TestDefaultNamespace_Execution(t *testing.T) {
	tests := []struct{ name, source, want string }{
		{"output forms", "Default.Print('a'); Default.PrintLn('b'); Default.Print; Default.PrintLn; Default.Print(); Default.PrintLn();", "ab\n\n\n"},
		{"global routine shadow", "procedure PrintLn(V: Variant); begin Default.Print('shadow:'); Default.Print(V); end; Default.PrintLn('builtin'); PrintLn('user');", "builtin\nshadow:user"},
		{"local routine shadow", "procedure Test; begin procedure PrintLn(V: Variant); begin Default.Print('shadow:'); Default.Print(V); end; Default.PrintLn('builtin'); PrintLn('user'); end; Test;", "builtin\nshadow:user"},
		{"parameter pointer shadow", "type P = procedure(V: Variant); procedure User(V: Variant); begin Default.Print('user:'); Default.Print(V); end; procedure Test(PrintLn: P); begin Default.PrintLn('builtin'); PrintLn('pointer'); end; Test(@User);", "builtin\nuser:pointer"},
		{"argument uses lexical value", "procedure Test(PrintLn: Integer); begin Default.PrintLn(PrintLn); end; Test(7);", "7\n"},
		{"argument implicit call once", "var N := 0; function Next: Integer; begin N += 1; Result := N; end; Default.PrintLn(Next); Default.PrintLn(N);", "1\n1\n"},
		{"typed bare callback", "type P = procedure(V: Variant); procedure PrintLn(V: Variant); begin Default.Print('user'); end; var F: P := Default.PrintLn; F('builtin');", "builtin\n"},
		{"address callback", "var F := @Default.PrintLn; F('address');", "address\n"},
		{"typed address callback", "type P = procedure(V: Variant); var F: P := @Default.Print; F('address');", "address"},
		{"local receiver", "type T = class procedure PrintLn(V: Variant); begin Print('member:'); Print(V); end; end; procedure Test; begin var &Default := T.Create; Default.PrintLn('local'); end; Test;", "member:local"},
		{"receiver parameter", "type T = class procedure PrintLn(V: Variant); begin Print('member:'); Print(V); end; end; procedure Test(&Default: T); begin Default.PrintLn('parameter'); end; Test(T.Create);", "member:parameter"},
		{"implicit receiver field", "type T = class procedure PrintLn(V: Variant); begin Print('member:'); Print(V); end; end; type Holder = class &Default: T; procedure Test; begin Default.PrintLn('field'); end; end; var H := Holder.Create; H.Default := T.Create; H.Test;", "member:field"},
		{"implicit output member", "type T = class procedure PrintLn(V: Variant); begin Print('member:'); Print(V); end; procedure Test; begin Default.PrintLn('builtin'); Self.PrintLn('user'); end; end; T.Create.Test;", "builtin\nmember:user"},
		{"intrinsic Default remains", "Default.PrintLn(Default(Integer));", "0\n"},
		{"mixed case", "dEfAuLt.pRiNt('a'); DEFAULT.PRINTLN('b');", "ab\n"},
		{"caller receiver cannot change namespace binding", "type T = class procedure PrintLn(V: Variant); begin Default.Print('member:'); Default.Print(V); end; end; procedure Test; begin var &Default := T.Create; Default.PrintLn('local'); end; Test;", "member:local"},
		{"routine definition retains namespace", "type T = class procedure PrintLn(V: Variant); begin Print('member:'); Print(V); end; end; procedure Output; begin Default.PrintLn('builtin'); end; procedure Test; begin var &Default := T.Create; Output; end; Test;", "builtin\n"},
		{"local type receiver", "procedure Test; begin type &Default = class class procedure PrintLn(V: Variant); begin Print('class:'); Print(V); end; end; Default.PrintLn('local type'); end; Test;", "class:local type"},
		{"implicit receiver property", "type T = class procedure PrintLn(V: Variant); begin Print('member:'); Print(V); end; end; type Holder = class FValue: T; property &Default: T read FValue write FValue; procedure Test; begin Default.PrintLn('property'); end; end; var H := Holder.Create; H.Default := T.Create; H.Test;", "member:property"},
		{"variadic output", "Default.Print('a', 1); Default.PrintLn('b', 2);", "a1b2\n"},
		{"caller helper cannot change namespace binding", "procedure Output; begin Default.PrintLn('builtin'); end; procedure Test; begin type &Default = helper for Integer class procedure PrintLn(V: Variant); begin Print('helper:'); Print(V); end; end; Output; end; Test;", "builtin\n"},
		{"caller helper cannot change bare output binding", "procedure Output; begin Default.PrintLn; end; procedure Test; begin type &Default = helper for Integer class procedure PrintLn; begin Print('helper'); end; end; Output; end; Test;", "\n"},
		{"caller helper cannot change callback binding", "type P = procedure(V: Variant); procedure Output; begin var F: P := Default.PrintLn; F('builtin'); end; procedure Test; begin type &Default = helper for Integer class procedure PrintLn; begin Print('helper'); end; end; Output; end; Test;", "builtin\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			engine, err := New(WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := engine.Eval(tt.source); err != nil {
				t.Fatalf("execution failed: %v", err)
			}
			if got := output.String(); got != tt.want {
				t.Fatalf("output = %q; want %q", got, tt.want)
			}
		})
	}
}
