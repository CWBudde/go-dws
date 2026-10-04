package interp

import "testing"

func TestRecordArity_DefaultsThroughCompilePath(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type R = record
 procedure Take(v: Integer = 7); begin PrintLn(v); end;
 class procedure StaticTake(v: Integer = 8); begin PrintLn(v); end;
 procedure Run;
 begin
  Take();
  Self.Take();
  StaticTake();
 end;
end;
var item: R;
item.Take();
R.StaticTake();
var meta := R;
meta.StaticTake();
item.Run();
`, "record_arity_defaults.dws", "7\n8\n8\n7\n7\n8\n")
}

func TestRecordArity_OverloadsThroughCompilePath(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type R = record
 procedure Pick(v: Integer = 7); overload; begin PrintLn(v); end;
 procedure Pick(v: String); overload; begin PrintLn(v); end;
 class procedure StaticPick(v: Integer = 8); overload; begin PrintLn(v); end;
 class procedure StaticPick(v: String); overload; begin PrintLn(v); end;
 procedure Run; begin Pick('implicit'); StaticPick('static'); end;
end;
var item: R;
item.Pick();
item.Pick('explicit');
R.StaticPick();
R.StaticPick('class');
item.Run();
`, "record_arity_overloads.dws", "7\nexplicit\n8\nclass\nimplicit\nstatic\n")
}

func TestRecordArity_MixedAndConstantDefaultsThroughCompilePath(t *testing.T) {
	for _, tt := range []struct{ name, source, want string }{
		{"implicit mixed overloads", `
type R = record
 procedure Pick(v: Integer); overload; begin PrintLn('instance'); end;
 class procedure Pick(v: String); overload; begin PrintLn('static'); end;
 procedure Run; begin Pick('s'); end;
end;
var item: R; item.Run();`, "static\n"},
		{"class default through instance", `
type R = record
 procedure Pick(v: Integer); overload; begin PrintLn('instance'); end;
 class procedure Pick(v: String = 'default'); overload; begin PrintLn(v); end;
end;
var item: R; item.Pick();`, "default\n"},
		{"record constants", `
type R = record
 const D = 7;
 procedure Pick(v: Integer = D); begin PrintLn(v); end;
 class procedure StaticPick(v: Integer = D); begin PrintLn(v); end;
end;
var D := 99;
var item: R;
item.Pick();
R.StaticPick();`, "7\n7\n"},
		{"implicit instance in mixed set", `
type R = record
 procedure Pick(v: Integer); overload; begin PrintLn('instance'); end;
 class procedure Pick(v: String); overload; begin PrintLn('static'); end;
 procedure Run; begin Pick(1); end;
end;
var item: R; item.Run();`, "instance\n"},
		{"out of line defaults", `
type R = record
 procedure Pick(v: Integer = 7);
 class procedure StaticPick(v: Integer = 8);
end;
procedure R.Pick(v: Integer); begin PrintLn(v); end;
class procedure R.StaticPick(v: Integer); begin PrintLn(v); end;
var item: R; item.Pick(); R.StaticPick();`, "7\n8\n"},
		{"lexical constant defaults", `
const D = 7;
type R = record
 procedure Pick(v: Integer = D); begin PrintLn(v); end;
 class procedure StaticPick(v: Integer = D); begin PrintLn(v); end;
end;
procedure Run; begin var D := 99; var item: R; item.Pick(); R.StaticPick(); end;
Run();`, "7\n7\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			compileAndRunWithHelperTransfer(t, tt.source, "record_arity_mixed.dws", tt.want)
		})
	}
}

func TestRecordArity_MixedSuppliedArgumentsOnceThroughCompilePath(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
var calls := 0;
function Next: Integer; begin calls += 1; Result := calls; end;
type R = record
 procedure Pick(v: Integer; extra: Integer = 7); overload; begin PrintLn(v); PrintLn(extra); end;
 class procedure Pick(v: String; extra: Integer = 8); overload; begin PrintLn(v); end;
 procedure Run; begin Pick(Next()); end;
end;
var item: R;
item.Pick(Next());
function Make: R; begin Result := item; end;
Make().Pick(Next());
item.Run();
PrintLn(calls);
`, "record_arity_once.dws", "1\n7\n2\n7\n3\n7\n3\n")
}

func TestRecordArity_OutOfLineOverloadedDefaultsThroughCompilePath(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type R = record
 procedure Pick(v: Integer = 7); overload;
 procedure Pick(v: String); overload;
 class procedure StaticPick(v: Integer = 8); overload;
 class procedure StaticPick(v: String); overload;
end;
procedure R.Pick(v: Integer); begin PrintLn(v); end;
procedure R.Pick(v: String); begin PrintLn(v); end;
class procedure R.StaticPick(v: Integer); begin PrintLn(v); end;
class procedure R.StaticPick(v: String); begin PrintLn(v); end;
var item: R;
item.Pick(); item.Pick('instance');
R.StaticPick(); R.StaticPick('class');
`, "record_arity_outofline.dws", "7\ninstance\n8\nclass\n")
}
