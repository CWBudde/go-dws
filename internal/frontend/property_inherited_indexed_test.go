package frontend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// ReadInherited bypasses the member case hint and selects the lexical parent's
// descriptor before ReadPropertyExpr consumes () and ReadArguments consumes [].
func TestCompile_InheritedIndexedCompatibility(t *testing.T) {
	const prefix = "type TBase = class\n function Get(I: Integer): Integer; begin Result := I; end;\n property Prop[I: Integer]: Integer read Get reintroduce; default;\n property Plain[I: Integer]: Integer read Get;\nend;\ntype TChild = class(TBase) function Read: Integer; end;\nfunction TChild.Read: Integer;\n"
	const hint = "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 8, column: 31]"
	tests := []struct {
		name, body, want string
		hints            semantic.HintsLevel
	}{
		{"normal", "begin Result := inherited Prop()[3]; end;", hint, semantic.HintsLevelNormal},
		{"pedantic bypasses case", "begin Result := inherited pRoP()[3]; end;", hint, semantic.HintsLevelPedantic},
		{"disabled", "begin Result := inherited Prop()[3]; end;", "", semantic.HintsLevelDisabled},
		{"directive", "{$HINTS OFF}\nbegin Result := inherited Prop()[3]; end;", "", semantic.HintsLevelNormal},
		{"comment", "begin Result := inherited Prop {comment}\n()[3]; end;", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 9, column: 1]", semantic.HintsLevelNormal},
		{"ordinary bare", "begin Result := inherited Plain[3]; end;", "", semantic.HintsLevelNormal},
		{"flagged bare", "begin Result := inherited Prop[3]; end;", "", semantic.HintsLevelNormal},
		{"ordinary call stop", "begin Result := inherited Plain()[Missing]; Other; end;", "Syntax Error: More arguments expected [line: 8, column: 27]\nSyntax Error: Not a method [line: 8, column: 32]", semantic.HintsLevelNormal},
		{"type", "begin Result := inherited Prop()[True]; end;", hint + "\nSyntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 8, column: 32]", semantic.HintsLevelNormal},
		{"extra", "begin Result := inherited Prop()[1, 2]; end;", hint + "\nSyntax Error: Too many arguments [line: 8, column: 32]", semantic.HintsLevelNormal},
		{"type suppresses count", "begin Result := inherited Prop()[True, 2]; end;", hint + "\nSyntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 8, column: 32]", semantic.HintsLevelNormal},
		{"child before count", "begin Result := inherited Prop()[1, Missing]; end;", hint + "\nSyntax Error: Unknown name \"Missing\" [line: 8, column: 37]\nSyntax Error: Too many arguments [line: 8, column: 32]", semantic.HintsLevelNormal},
		{"trailing call stop", "begin Result := inherited Prop[3](); Other; end;", "Syntax Error: Not a method [line: 8, column: 34]", semantic.HintsLevelNormal},
		{"bare ordinary type", "begin Result := inherited Plain[True]; end;", "Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 8, column: 27]", semantic.HintsLevelNormal},
		{"bare flagged type", "begin Result := inherited Prop[True]; end;", "Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 8, column: 31]", semantic.HintsLevelNormal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(prefix+tt.body, Options{Filename: "<test>", HintsLevel: tt.hints, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_InheritedIndexedAccessorContexts(t *testing.T) {
	const prefix = "type TBase = class\n function Get(I: Integer): Integer; begin Result := I; end; procedure SetValue(I, V: Integer); begin end;\n property Prop[I: Integer]: Integer read Get reintroduce;\nend;\ntype TChild = class(TBase) function Read: Integer; end;\nfunction TChild.Read: Integer;\n"
	const hint = "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 7, column: 31]"
	tests := []struct{ name, source, want string }{
		{"write only empty", strings.Replace(prefix, "read Get reintroduce", "write SetValue reintroduce", 1) + "begin Result := inherited Prop()[Missing]; end;", hint + "\nSyntax Error: Cannot read a write only property [line: 7, column: 32]\nSyntax Error: Array expected [line: 7, column: 33]"},
		{"write only flagged bare", strings.Replace(prefix, "read Get reintroduce", "write SetValue reintroduce", 1) + "begin Result := inherited Prop[Missing]; end;", "Syntax Error: Cannot read a write only property [line: 7, column: 31]\nSyntax Error: Array expected [line: 7, column: 31]"},
		{"write only ordinary bare", strings.Replace(prefix, "read Get reintroduce", "write SetValue", 1) + "begin Result := inherited Prop[Missing]; end;", "Syntax Error: Cannot read a write only property [line: 7, column: 27]\nSyntax Error: Array expected [line: 7, column: 31]"},
		{"deprecated before hint no case", strings.Replace(prefix, "read Get reintroduce;", "read Get reintroduce; deprecated 'old';", 1) + "begin Result := inherited pRoP()[3]; end;", "Warning: \"Prop\" has been deprecated: old [line: 7, column: 27]\n" + hint},
		{"class getter eligibility", strings.Replace(strings.Replace(prefix, " function Read", " class function Read", 1), "function TChild.Read", "class function TChild.Read", 1) + "begin Result := inherited Prop()[3]; end;", hint + "\nSyntax Error: Read access of property should be a static method [line: 7, column: 32]\nSyntax Error: Class method or constructor expected [line: 7, column: 32]"},
		{"child before class eligibility", strings.Replace(strings.Replace(prefix, " function Read", " class function Read", 1), "function TChild.Read", "class function TChild.Read", 1) + "begin Result := inherited Prop()[Missing]; end;", hint + "\nSyntax Error: Unknown name \"Missing\" [line: 7, column: 34]\nSyntax Error: Read access of property should be a static method [line: 7, column: 32]\nSyntax Error: Class method or constructor expected [line: 7, column: 32]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(tt.source, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_InheritedIndexedMultipleArguments(t *testing.T) {
	const prefix = "type TBase = class function Get(I, J: Integer): Integer; begin Result := I+J; end; property Pair[I, J: Integer]: Integer read Get reintroduce; end;\ntype TChild = class(TBase) function Read: Integer; end;\nfunction TChild.Read: Integer;\n"
	const hint = "Hint: Property \"Pair\" reintroduced a method, you should remove empty brackets () [line: 4, column: 31]\n"
	tests := []struct{ name, body, want string }{
		{"missing", "begin Result := inherited Pair()[1]; end;", "Syntax Error: More arguments expected [line: 4, column: 32]"},
		{"type suppresses missing", "begin Result := inherited Pair()[True]; end;", "Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 4, column: 32]"},
		{"two types order", "begin Result := inherited Pair()[True, False]; end;", "Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 4, column: 32]\nSyntax Error: Argument 1 expects type \"Integer\" instead of \"Boolean\" [line: 4, column: 32]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(prefix+tt.body, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != hint+tt.want {
				t.Fatalf("got %q; want %q", got, hint+tt.want)
			}
		})
	}
}

func TestCompile_InheritedIndexedReadBoundaries(t *testing.T) {
	const prefix = "type TBase = class function Get(I: Integer): Integer; begin Result := I; end; property Prop[I: Integer]: Integer read Get reintroduce; end;\ntype TChild = class(TBase) procedure Read; end;\nprocedure TChild.Read;\n"
	tests := []struct{ name, body, want string }{
		{"nonempty call", "begin PrintLn(inherited Prop(1)[2]); end;", "Syntax Error: cannot call property 'Prop' as a method [line: 4, column: 15]"},
		{"compatibility write remains unsupported", "begin inherited Prop()[1] := 2; end;", "Syntax Error: cannot call property 'Prop' as a method [line: 4, column: 7]"},
		{"bare write remains unsupported", "begin inherited Prop[1] := 2; end;", "Syntax Error: Array expected [line: 4, column: 21]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(prefix+tt.body, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_InheritedIndexedParentMarker(t *testing.T) {
	const source = "type TBase = class function Get(I: Integer): Integer; begin Result := I; end; property Prop[I: Integer]: Integer read Get; end;\ntype TChild = class(TBase)\n property Prop[I: Integer]: Integer read Get reintroduce;\n function Read: Integer; begin Result := inherited Prop()[Missing]; Other; end;\nend;"
	result := CompileWithOptions(source, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
	const want = "Syntax Error: More arguments expected [line: 4, column: 52]\nSyntax Error: Not a method [line: 4, column: 56]"
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
		t.Fatalf("got %q; want %q", got, want)
	}
}

func TestCompile_InheritedIndexedUnreadableArrayRecovery(t *testing.T) {
	const prefix = "type TInts = array of Integer;\ntype TBase = class procedure SetValue(I: Integer; V: TInts); begin end; property Prop[I: Integer]: TInts write SetValue reintroduce; end;\ntype TChild = class(TBase) function Read: Integer; end;\nfunction TChild.Read: Integer;\n"
	const hint = "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 5, column: 31]\n"
	tests := []struct{ name, body, want string }{
		{"empty pair index child", "begin Result := inherited Prop()[Missing]; end;", hint + "Syntax Error: Cannot read a write only property [line: 5, column: 32]\nSyntax Error: Unknown name \"Missing\" [line: 5, column: 34]"},
		{"bare index child", "begin Result := inherited Prop[Missing]; end;", "Syntax Error: Cannot read a write only property [line: 5, column: 31]\nSyntax Error: Unknown name \"Missing\" [line: 5, column: 32]"},
		{"empty pair array element recovery", "begin Result := inherited Prop()[2]; end;", hint + "Syntax Error: Cannot read a write only property [line: 5, column: 32]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(prefix+tt.body, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

// Static inherited property reads remain on their existing unsupported path;
// this pins scope preservation, not an upstream static diagnostic contract.
func TestCompile_InheritedIndexedStaticPreservation(t *testing.T) {
	const prefix = "type TBase = class\n class function Get(I: Integer): Integer; begin Result := I; end;\n class property Prop[I: Integer]: Integer read Get reintroduce;\nend;\ntype TChild = class(TBase) class function Read(Prop: Integer): Integer; static; end;\nclass function TChild.Read(Prop: Integer): Integer;\n"
	tests := []struct{ name, body, want string }{
		{"empty pair", "begin Result := inherited Prop()[Prop]; end;", "Syntax Error: cannot call property 'Prop' as a method [line: 7, column: 17]"},
		{"bare", "begin Result := inherited Prop[Prop]; end;", "Syntax Error: Array expected [line: 7, column: 31]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(prefix+tt.body, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_InheritedIndexedOmittedArguments(t *testing.T) {
	const prefix = "type TBase = class\n function Get(I: Integer): Integer; begin Result := I; end;\n property Prop[I: Integer]: Integer read Get reintroduce;\n property Plain[I: Integer]: Integer read Get;\nend;\ntype TChild = class(TBase) function Read: Integer; end;\nfunction TChild.Read: Integer;\n"
	const hint = "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 8, column: 31]\n"
	tests := []struct{ name, body, want string }{
		{"compatibility pair", "begin Result := inherited Prop(); end;", hint + "Syntax Error: More arguments expected [line: 8, column: 32]"},
		{"continues after count", "begin Result := inherited Prop(); Other; end;", hint + "Syntax Error: More arguments expected [line: 8, column: 32]\nSyntax Error: Unknown name \"Other\" [line: 8, column: 35]"},
		{"bare marked", "begin Result := inherited Prop; end;", "Syntax Error: More arguments expected [line: 8, column: 31]"},
		{"bare ordinary", "begin Result := inherited Plain; end;", "Syntax Error: More arguments expected [line: 8, column: 27]"},
		{"ordinary call stops", "begin Result := inherited Plain(); Other; end;", "Syntax Error: More arguments expected [line: 8, column: 27]\nSyntax Error: Not a method [line: 8, column: 32]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(prefix+tt.body, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_IndexedPropertyEmptyArguments(t *testing.T) {
	const inheritedPrefix = "type TBase = class\n function Get(I: Integer): Integer; begin Result := I; end;\n property Prop[I: Integer]: Integer read Get reintroduce;\n property Plain[I: Integer]: Integer read Get;\nend;\ntype TChild = class(TBase) function Read: Integer; end;\nfunction TChild.Read: Integer;\n"
	tests := []struct{ name, source, want string }{
		{"inherited empty pair", inheritedPrefix + "begin Result := inherited Prop()[]; end;", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 8, column: 31]\nSyntax Error: More arguments expected [line: 8, column: 32]"},
		{"inherited bare marked", inheritedPrefix + "begin Result := inherited Prop[]; end;", "Syntax Error: More arguments expected [line: 8, column: 31]"},
		{"inherited bare ordinary", inheritedPrefix + "begin Result := inherited Plain[]; end;", "Syntax Error: More arguments expected [line: 8, column: 27]"},
		{"explicit empty pair", "type T = class function Get(I: Integer): Integer; begin Result := I; end; property Prop[I: Integer]: Integer read Get reintroduce; end;\nvar O := T.Create;\nvar A := O.Prop()[];", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 3, column: 16]\nSyntax Error: More arguments expected [line: 3, column: 17]"},
		{"implicit empty pair", "type T = class function Get(I: Integer): Integer; begin Result := I; end; property Prop[I: Integer]: Integer read Get reintroduce; function Read: Integer; end;\nfunction T.Read: Integer;\nbegin Result := Prop()[]; end;", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 3, column: 21]\nSyntax Error: More arguments expected [line: 3, column: 22]"},
		{"ordinary array unchanged", "var A: array of Integer;\nvar B := A[]; Missing;", "Syntax Error: Expression expected [line: 2, column: 12]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(tt.source, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_IndexedPropertyUnfinishedArguments(t *testing.T) {
	const prefix = "type TBase = class\n function Get(I: Integer): Integer; begin Result := I; end;\n property Prop[I: Integer]: Integer read Get reintroduce;\nend;\ntype TChild = class(TBase) function Read: Integer; end;\nfunction TChild.Read: Integer;\n"
	const hint = "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 7, column: 31]\n"
	tests := []struct{ name, source, want string }{
		{"type check never reached", prefix + "begin Result := inherited Prop()[True; end;", hint + "Syntax Error: \")\" expected [line: 7, column: 38]"},
		{"child error precedes stop", prefix + "begin Result := inherited Prop()[Missing; Other; end;", hint + "Syntax Error: Unknown name \"Missing\" [line: 7, column: 34]\nSyntax Error: \")\" expected [line: 7, column: 41]"},
		{"extra and type checks never reached", prefix + "begin Result := inherited Prop()[True, Missing; end;", hint + "Syntax Error: Unknown name \"Missing\" [line: 7, column: 40]\nSyntax Error: \")\" expected [line: 7, column: 47]"},
		{"eligibility never reached", strings.Replace(strings.Replace(prefix, " function Read", " class function Read", 1), "function TChild.Read", "class function TChild.Read", 1) + "begin Result := inherited Prop()[True; end;", hint + "Syntax Error: \")\" expected [line: 7, column: 38]"},
		{"explicit group", "type T = class function Get(I: Integer): Integer; begin Result := I; end; property Prop[I: Integer]: Integer read Get reintroduce; end;\nvar O := T.Create;\nvar A := O.Prop()[True;", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 3, column: 16]\nSyntax Error: \")\" expected [line: 3, column: 23]"},
		{"implicit group", "type T = class function Get(I: Integer): Integer; begin Result := I; end; property Prop[I: Integer]: Integer read Get reintroduce; function Read: Integer; end;\nfunction T.Read: Integer;\nbegin Result := Prop()[True; end;", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 3, column: 21]\nSyntax Error: \")\" expected [line: 3, column: 28]"},
		{"ordinary array recovery unchanged", "var A: array of Integer;\nvar B := A[2;", "Syntax Error: \"]\" expected [line: 2, column: 13]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(tt.source, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_InheritedIndexedArgumentStopCutoff(t *testing.T) {
	const prefix = "type TBase = class\n function Get(I: Integer): Integer; begin Result := I; end;\n property Prop[I: Integer]: Integer read Get reintroduce;\nend;\ntype TChild = class(TBase) function Read: Integer; end;\nfunction TChild.Read: Integer;\n"
	tests := []struct{ name, body, want string }{
		{"prior semantic diagnostic survives", "begin Earlier; Result := inherited Prop()[True; Later; end;", "Syntax Error: Unknown name \"Earlier\" [line: 7, column: 7]\nHint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 7, column: 40]\nSyntax Error: \")\" expected [line: 7, column: 47]"},
		{"later lexer directive is cut off", "begin Result := inherited Prop()[True; {$FOOBAR}\nend;", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 7, column: 31]\nSyntax Error: \")\" expected [line: 7, column: 38]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(prefix+tt.body, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_DeferredIndexDiagnosticBoundaries(t *testing.T) {
	const inheritedPrefix = "type TBase = class\n function Get(I: Integer): Integer; begin Result := I; end;\n property Prop[I: Integer]: Integer read Get reintroduce;\nend;\ntype TChild = class(TBase) function Read: Integer; end;\nfunction TChild.Read: Integer;\n"
	tests := []struct{ name, source, want string }{
		{"inherited empty group keeps later directive", inheritedPrefix + "begin Result := inherited Prop()[]; end;\n{$ERROR 'late'}", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 7, column: 31]\nSyntax Error: More arguments expected [line: 7, column: 32]\nCompile Error: late [line: 8, column: 3]"},
		{"explicit empty group keeps later directive", "type T = class function Get(I: Integer): Integer; begin Result := I; end; property Prop[I: Integer]: Integer read Get reintroduce; end;\nvar O := T.Create;\nvar A := O.Prop()[];\n{$ERROR 'late'}", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 3, column: 16]\nSyntax Error: More arguments expected [line: 3, column: 17]\nCompile Error: late [line: 4, column: 3]"},
		{"ordinary empty group suppresses forward end check", "procedure Pending; forward;\nvar A: array of Integer;\nvar B := A[];\n", "Syntax Error: Expression expected [line: 3, column: 12]"},
		{"ordinary empty group suppresses incomplete class end check", "type TForward = class;\nvar A: array of Integer;\nvar B := A[];\n", "Syntax Error: Expression expected [line: 3, column: 12]"},
		{"ordinary empty group cuts off later directive", "var A: array of Integer;\nvar B := A[];\n{$ERROR 'late'}", "Syntax Error: Expression expected [line: 2, column: 12]"},
		{"ordinary enclosing call suppresses end checks", "procedure Pending; forward;\ntype TForward = class;\nvar A: array of Integer;\nPrintLn(A[]);", "Syntax Error: Expression expected [line: 4, column: 11]"},
		{"enclosing stop cuts off directive after recovered property", inheritedPrefix + "begin Result := inherited Prop()[]; var X := ; end;\n{$ERROR 'late'}", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 7, column: 31]\nSyntax Error: More arguments expected [line: 7, column: 32]\nSyntax Error: Expression expected [line: 7, column: 46]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(tt.source, Options{Filename: "<test>", HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
	t.Run("parse only retains original cutoff", func(t *testing.T) {
		result := ParseWithOptions(inheritedPrefix+"begin Result := inherited Prop()[]; end;\n{$ERROR 'late'}", Options{})
		const want = "Syntax Error: Expression expected [line: 7, column: 34]"
		if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
			t.Fatalf("got %q; want %q", got, want)
		}
	})
}

func TestCompile_DeferredIndexImportedForwardChecks(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Fwd.pas"), []byte("unit Fwd;\ninterface\nprocedure Pending;\nimplementation\nend."), 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct{ name, source, want string }{
		{"ordinary empty index suppresses unit forward", "uses Fwd;\nvar A: array of Integer;\nvar B := A[];", "Syntax Error: Expression expected [line: 3, column: 12]"},
		{"recovered property retains unit forward", "uses Fwd;\ntype T = class function Get(I: Integer): Integer; begin Result := I; end; property Prop[I: Integer]: Integer read Get reintroduce; end;\nvar O := T.Create;\nvar A := O.Prop()[];", "Hint: Property \"Prop\" reintroduced a method, you should remove empty brackets () [line: 4, column: 16]\nSyntax Error: More arguments expected [line: 4, column: 17]\nSyntax Error: The function \"Pending\" was forward declared but not implemented [line: 3, column: 11]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompileWithOptions(tt.source, Options{Filename: filepath.Join(dir, "main.dws"), UnitSearchPaths: []string{dir}, HintsLevel: semantic.HintsLevelNormal, DisableSymbolDictionaryDiagnostics: true})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}
