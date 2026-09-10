// Package lexer provides lexical analysis for DWScript source code.
// This file contains the declaration tracker that backs {$IF Declared('X')}.
package lexer

import (
	"github.com/cwbudde/go-dws/pkg/ident"
)

// The declaration tracker is a deliberate heuristic, in the same spirit as trackConst.
//
// {$IF Declared('X')} asks a symbol-table question, but compiler directives are resolved
// while tokenizing — long before the parser has built an AST and before the semantic
// analyzer has a symbol table. The lexer therefore watches the token stream it is already
// producing and records the names it sees being declared. Because the lexer scans
// strictly forward, this yields exactly the point-of-use visibility DWScript has: a name
// declared further down the file is not visible to an earlier {$IF}.
//
// It is explicitly NOT a second parser. It recognizes the shapes DWScript scripts
// actually use and ignores everything else; a name it fails to recognize simply reports
// as not declared. Anything needing real scope, overload or visibility rules belongs in
// internal/semantic, which implements the runtime Declared() intrinsic.
//
// Recognized shapes:
//
//	type T = class ... end;              -> T, T.Member
//	type T = record ... end;             -> T, T.Member
//	type T = interface ... end;          -> T, T.Member
//	type T = helper for U ... end;       -> T, T.Member and U.Member
//	type T = class helper for U ... end; -> likewise
//	type T = <anything else>;            -> T
//	const/var NAME[, NAME]               -> NAME (file scope only)
//	function/procedure NAME              -> NAME (file scope only)
//
// Member names are recorded only in dotted form: a record field Dummy makes
// TMyRecord.Dummy declared but leaves the bare name dummy undeclared, matching
// SimpleScripts/declared.pas.

// declState is the position of the declaration tracker inside a type declaration.
type declState int

const (
	// declStateIdle is outside any type declaration.
	declStateIdle declState = iota
	// declStateTypeName expects the name of a type being declared.
	declStateTypeName
	// declStateTypeEq expects the '=' of a type declaration.
	declStateTypeEq
	// declStateTypeKind expects the type's right-hand side.
	declStateTypeKind
	// declStateClassSeen expects the token after "class" or "record", which decides
	// between a body, a metaclass ("class of X") and a forward declaration ("class;").
	declStateClassSeen
	// declStateHelperFor expects the "for" of a helper declaration.
	declStateHelperFor
	// declStateHelperTarget expects the helped type's name.
	declStateHelperTarget
	// declStateBody is inside a class/record/interface/helper body.
	declStateBody
)

// declTracker records the names the lexer has seen declared so far.
//
// It is part of the lexer's mutable state and is deep-copied by SaveState/RestoreState:
// unlike constValues (whose map inserts are idempotent), the state-machine fields below
// are position-dependent, so parser backtracking must rewind them along with the input
// position or a replayed token run would be interpreted in the wrong state.
type declTracker struct {
	// names holds every declared name, normalized via ident.Normalize. Members are
	// stored dotted ("thelper.proc"); bare member names are never inserted.
	names map[string]struct{}
	// typeName is the type currently being declared.
	typeName string
	// helperTarget is the helped type of a "helper for" declaration, if any.
	helperTarget string
	// tentativeName holds an identifier seen just after a section semicolon. DWScript
	// allows statements between file-scope declarations, so the name is committed only
	// once a ':', '=' or ',' confirms it really is a declaration and not the start of a
	// statement such as "PrintLn('x');" or "Total := 1;".
	tentativeName string
	// fieldNames buffers identifiers that may turn out to be field names, i.e. the
	// "A, B" of "A, B : Integer;". They are committed when the ':' arrives.
	fieldNames []string
	state      declState
	// bodyDepth counts open block constructs inside a type body; the body ends when it
	// falls back to zero.
	bodyDepth int
	// blockDepth counts open blocks at file scope, so that a routine's locals are not
	// mistaken for file-scope declarations.
	blockDepth int
	// inTypeSection records that a "type" keyword is still in effect, so further names
	// in the same section are recognized without repeating the keyword.
	inTypeSection bool
	// memberPending marks that the next identifier names a method or property.
	memberPending bool
	// declPending marks that the next identifier names a file-scope const/var/routine.
	declPending bool
	// declList marks that a file-scope declaration list is open, so a comma introduces
	// another name.
	declList bool
	// inRoutine marks that a routine declaration is open, so its parameters and locals
	// are not recorded as file-scope names. It is cleared by the routine's closing
	// "end". A routine declared "forward" therefore keeps it set until the
	// implementation is reached, which under-reports the names in between; that is an
	// accepted limit of the heuristic.
	inRoutine bool
	// atDeclStart marks a position inside a type body where a field declaration may begin.
	atDeclStart bool
	// inVarSection records that a file-scope "var" or "const" keyword is still in
	// effect, so further names in the same section are recognized without repeating it
	// ("var A : Integer; B : Integer;").
	inVarSection bool
	// expectSectionName marks the position just after a semicolon inside such a
	// section, where another declaration may begin.
	expectSectionName bool
}

// wellKnownTObjectMembers are the TObject members DWScript's internal unit always
// declares. They are seeded so {$IF Declared('TObject.Create')} resolves without the
// lexer ever having seen the internal unit's source.
var wellKnownTObjectMembers = []string{"Create", "Destroy", "Free", "ClassName", "ClassType"}

// newDeclTracker creates a declaration tracker seeded with the built-in symbols that are
// always in scope.
func newDeclTracker() *declTracker {
	t := &declTracker{
		names: make(map[string]struct{}),
		state: declStateIdle,
	}
	t.add("TObject")
	for _, m := range wellKnownTObjectMembers {
		t.addMember("TObject", m)
	}
	return t
}

// clone returns a deep copy of the tracker, used by SaveState/RestoreState.
func (t *declTracker) clone() *declTracker {
	if t == nil {
		return nil
	}
	names := make(map[string]struct{}, len(t.names))
	for k, v := range t.names {
		names[k] = v
	}
	fields := make([]string, len(t.fieldNames))
	copy(fields, t.fieldNames)

	c := *t
	c.names = names
	c.fieldNames = fields
	return &c
}

// add records a plain declared name.
func (t *declTracker) add(name string) {
	if name == "" {
		return
	}
	t.names[ident.Normalize(name)] = struct{}{}
}

// addMember records a member of owner as the dotted name "owner.member". The bare member
// name is deliberately not recorded.
func (t *declTracker) addMember(owner, member string) {
	if owner == "" || member == "" {
		return
	}
	t.names[ident.Normalize(owner)+"."+ident.Normalize(member)] = struct{}{}
}

// isDeclared reports whether name (plain or dotted) has been declared at this point in
// the token stream. Comparison is case-insensitive.
func (t *declTracker) isDeclared(name string) bool {
	if t == nil || name == "" {
		return false
	}
	_, ok := t.names[ident.Normalize(name)]
	return ok
}

// trackDeclaration feeds a freshly produced token to the declaration tracker.
//
// It must be called exactly once per generated token, in source order — see the call
// sites in lexer.go, which track at token-generation time only, never when a buffered
// token is later handed out by NextToken.
func (l *Lexer) trackDeclaration(tok Token) {
	if l.decls == nil {
		l.decls = newDeclTracker()
	}
	// Tokens inside an inactive conditional branch are not part of the program.
	if l.isSkippingTokens() {
		return
	}
	l.decls.feed(tok)
}

// feed advances the tracker's state machine by one token.
func (t *declTracker) feed(tok Token) {
	switch t.state {
	case declStateBody:
		t.feedBody(tok)
	case declStateTypeName:
		if tok.Type == IDENT {
			t.typeName = tok.Literal
			t.state = declStateTypeEq
			return
		}
		t.feedFileScope(tok)
	case declStateTypeEq:
		if tok.Type == EQ {
			t.state = declStateTypeKind
			return
		}
		t.resetType()
		t.feedFileScope(tok)
	case declStateTypeKind:
		t.feedTypeKind(tok)
	case declStateClassSeen:
		t.feedClassSeen(tok)
	case declStateHelperFor:
		if tok.Type == FOR {
			t.state = declStateHelperTarget
			return
		}
		t.openBody()
		t.feedBody(tok)
	case declStateHelperTarget:
		if tok.Type == IDENT {
			t.helperTarget = tok.Literal
			t.openBody()
			return
		}
		t.openBody()
		t.feedBody(tok)
	case declStateIdle:
		t.feedFileScope(tok)
	}
}

// feedFileScope handles tokens outside any type declaration body.
//
//nolint:gocyclo // A flat switch over the declaration shapes is clearer than dispatch indirection.
func (t *declTracker) feedFileScope(tok Token) {
	switch tok.Type {
	case TYPE:
		if t.blockDepth == 0 {
			t.inTypeSection = true
			t.state = declStateTypeName
		}
		t.endVarSection()
		t.declPending = false
		t.declList = false
	case VAR, CONST:
		t.inTypeSection = false
		t.state = declStateIdle
		t.declPending = t.blockDepth == 0 && !t.inRoutine
		t.declList = false
		t.endVarSection()
		t.inVarSection = t.declPending
	case FUNCTION, PROCEDURE:
		t.inTypeSection = false
		t.state = declStateIdle
		t.declPending = t.blockDepth == 0 && !t.inRoutine
		t.declList = false
		t.endVarSection()
		t.inRoutine = true
	case IDENT:
		switch {
		case t.declPending:
			t.add(tok.Literal)
			t.declPending = false
			t.declList = true
		case t.expectSectionName:
			t.tentativeName = tok.Literal
			t.expectSectionName = false
		}
	case COLON, EQ:
		t.commitTentative()
	case ASSIGN:
		// "NAME := value" at file scope is an assignment statement, not a continuation
		// of a var section: the inference form "var x := 1" carries its own keyword.
		t.endVarSection()
	case COMMA:
		t.commitTentative()
		if t.declList {
			t.declPending = true
		}
	case SEMICOLON:
		t.declPending = false
		t.declList = false
		t.tentativeName = ""
		t.expectSectionName = t.inVarSection && t.blockDepth == 0 && !t.inRoutine
		if t.inTypeSection && t.blockDepth == 0 {
			t.state = declStateTypeName
		}
	case BEGIN, CASE, TRY, ASM:
		t.inTypeSection = false
		t.declPending = false
		t.declList = false
		t.endVarSection()
		t.blockDepth++
	case END:
		t.declPending = false
		t.declList = false
		if t.blockDepth > 0 {
			t.blockDepth--
		}
		if t.blockDepth == 0 {
			t.inRoutine = false
		}
	case IMPLEMENTATION, INITIALIZATION, FINALIZATION, UNIT, INTERFACE:
		t.inTypeSection = false
		t.declPending = false
		t.declList = false
		t.endVarSection()
		t.state = declStateIdle
	default:
		t.declPending = false
		t.declList = false
		if t.tentativeName != "" {
			// The identifier was the start of a statement, not a declaration, so the
			// section is over.
			t.endVarSection()
		}
	}
}

// commitTentative accepts a pending section name once a token confirms it introduces a
// declaration rather than a statement.
func (t *declTracker) commitTentative() {
	if t.tentativeName == "" {
		return
	}
	t.add(t.tentativeName)
	t.tentativeName = ""
	t.declList = true
}

// endVarSection closes a file-scope var/const section and drops any unconfirmed name.
func (t *declTracker) endVarSection() {
	t.inVarSection = false
	t.expectSectionName = false
	t.tentativeName = ""
}

// feedTypeKind inspects the first token of the right-hand side of "type NAME = ...".
func (t *declTracker) feedTypeKind(tok Token) {
	// The type itself is declared regardless of what its definition turns out to be.
	t.add(t.typeName)

	switch tok.Type {
	case CLASS, RECORD:
		t.state = declStateClassSeen
	case INTERFACE:
		t.openBody()
	case HELPER:
		t.state = declStateHelperFor
	case SEMICOLON:
		// A forward declaration: no body follows.
		t.resetType()
		t.feedFileScope(tok)
	default:
		// An alias, enum, array, set or similar: the name is declared, nothing more.
		t.resetType()
	}
}

// feedClassSeen decides what follows "class"/"record" in a type declaration.
func (t *declTracker) feedClassSeen(tok Token) {
	switch tok.Type {
	case HELPER:
		t.state = declStateHelperFor
	case OF:
		// "class of X" is a metaclass: no body.
		t.resetType()
	case SEMICOLON:
		// "class;" is a forward declaration.
		t.resetType()
		t.feedFileScope(tok)
	default:
		// Anything else opens a body: "class(TBase)", "class end", "record Dummy : ...".
		t.openBody()
		t.feedBody(tok)
	}
}

// resetType leaves the current type declaration and returns to the enclosing section.
func (t *declTracker) resetType() {
	t.typeName = ""
	t.helperTarget = ""
	t.memberPending = false
	t.fieldNames = nil
	t.bodyDepth = 0
	t.atDeclStart = false
	if t.inTypeSection {
		t.state = declStateTypeName
	} else {
		t.state = declStateIdle
	}
}

// openBody enters the body of the type currently being declared.
func (t *declTracker) openBody() {
	t.add(t.typeName)
	t.state = declStateBody
	t.bodyDepth = 1
	t.memberPending = false
	t.atDeclStart = true
	t.fieldNames = nil
}

// registerMember records a member under the declared type and, for a helper, also under
// the helped type so Declared('TObject.Proc') resolves through THelper.
func (t *declTracker) registerMember(name string) {
	t.addMember(t.typeName, name)
	if t.helperTarget != "" {
		t.addMember(t.helperTarget, name)
	}
}

// feedBody handles tokens inside a class/record/interface/helper body.
//
//nolint:gocyclo // A flat switch over the member shapes is clearer than dispatch indirection.
func (t *declTracker) feedBody(tok Token) {
	switch tok.Type {
	case FUNCTION, PROCEDURE, CONSTRUCTOR, DESTRUCTOR, METHOD, PROPERTY, OPERATOR:
		if t.bodyDepth == 1 {
			t.memberPending = true
		}
		t.fieldNames = nil
		t.atDeclStart = false
	case CLASS:
		// Inside a body "class" is a modifier ("class procedure", "class var"), never a
		// nested type, so it neither opens a block nor interrupts a declaration.
	case IDENT:
		t.feedBodyIdent(tok.Literal)
	case COMMA:
		if t.bodyDepth == 1 && len(t.fieldNames) > 0 {
			t.atDeclStart = true
		}
	case COLON:
		if t.bodyDepth == 1 {
			for _, name := range t.fieldNames {
				t.registerMember(name)
			}
		}
		t.fieldNames = nil
		t.atDeclStart = false
	case SEMICOLON:
		t.fieldNames = nil
		t.memberPending = false
		t.atDeclStart = t.bodyDepth == 1
	case VAR, CONST:
		t.fieldNames = nil
		t.atDeclStart = t.bodyDepth == 1
	case PRIVATE, PROTECTED, PUBLIC, PUBLISHED, STRICT:
		// A visibility section introduces members rather than interrupting them, so the
		// next identifier still starts a field declaration.
		t.fieldNames = nil
		t.atDeclStart = t.bodyDepth == 1
	case BEGIN, CASE, TRY, ASM, RECORD:
		t.bodyDepth++
		t.fieldNames = nil
		t.memberPending = false
		t.atDeclStart = false
	case END:
		t.bodyDepth--
		t.fieldNames = nil
		t.memberPending = false
		if t.bodyDepth <= 0 {
			t.resetType()
			return
		}
		t.atDeclStart = t.bodyDepth == 1
	default:
		t.fieldNames = nil
		t.atDeclStart = false
	}
}

// feedBodyIdent handles an identifier inside a type body.
func (t *declTracker) feedBodyIdent(literal string) {
	if t.bodyDepth != 1 {
		return
	}
	if t.memberPending {
		t.registerMember(literal)
		t.memberPending = false
		t.atDeclStart = false
		return
	}
	if t.atDeclStart {
		t.fieldNames = append(t.fieldNames, literal)
		t.atDeclStart = false
	}
}
