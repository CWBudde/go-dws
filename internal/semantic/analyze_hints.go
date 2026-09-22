package semantic

import (
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// This file collects the hints DWScript raises about declarations that compile
// but say nothing: a private method marked virtual, a visibility section that
// repeats the visibility already in effect, an assignment of a variable to
// itself, a `begin … end` wrapped around a case..of else clause, and a `var`
// parameter of a reference type the routine never writes to.
//
// Each one carries upstream's own hint level, which decides whether it survives
// at a fixture suite's level (only UScriptTests raises the compiler to pedantic;
// every other runner leaves it at the hlStrict default — see
// internal/fixtureconfig):
//
//	private virtual     hlNormal    (dwsCompiler.pas ReadMethodDecl)
//	redundant section   hlStrict for a class body, hlNormal for a record or
//	                    helper body (ReadClassDecl / ReadRecordDecl / ReadHelperDecl)
//	assigning to itself hlNormal    (ReadAssign.CheckAssigningToSelf)
//	redundant begin     hlPedantic  (ReadCaseStatement)
//	unwritten var param hlPedantic  (HintReferenceConstVarParams)

// ============================================================================
// Private virtual methods
// ============================================================================

// checkPrivateVirtualMethod hints that `virtual` buys a private method nothing:
// a descendant cannot see the method, so it can never override it. Upstream
// raises it while reading the directive, so the anchor is the `virtual` keyword
// rather than the method name.
func (a *Analyzer) checkPrivateVirtualMethod(method *ast.FunctionDecl) {
	if a == nil || method == nil || !method.IsVirtual {
		return
	}
	if method.Visibility != ast.VisibilityPrivate {
		return
	}
	pos := method.VirtualPos
	if pos.Line == 0 || pos.Column == 0 {
		return
	}
	a.addHintAt(pos, "Private virtual methods cannot be overridden [line: %d, column: %d]",
		pos.Line, pos.Column)
}

// ============================================================================
// Redundant visibility sections
// ============================================================================

// visibilityInEffectInRecord is the visibility a named record or helper body
// starts out in, so an opening `public` section restates it and is redundant
// (FailureScripts/record_visibility_redundant line 3).
const visibilityInEffectInRecord = "public"

// checkClassVisibilitySections diagnoses the visibility specifiers written in a
// class body.
//
// A class body differs from a record's in two ways, both of which upstream
// spells out in ReadClassDecl: its opening specifier is never redundant
// (`firstVisibilityToken`), whichever one it names — OverloadsPass/
// meth_private_public opens two classes with `private` and with `public` and
// expects nothing — and the hint is hlStrict rather than hlNormal, which is why
// the class under {$HINTS NORMAL} in FailureScripts/class_visibility_redundant
// stays silent.
func (a *Analyzer) checkClassVisibilitySections(decl *ast.ClassDecl) {
	if a == nil || decl == nil {
		return
	}
	a.checkVisibilitySectionRedundancy(decl.VisibilitySections, "", HintsLevelStrict, nil)
}

// checkVisibilitySectionRedundancy walks a type body's visibility specifiers in
// source order and hints on each one that restates the visibility already in
// effect. `published` counts as a specifier of its own here even where it
// behaves like `public` for member access — `published` followed by `public` is
// a change, not a repetition.
//
// initial is the visibility the body starts out in, or "" when the body has none
// and its opening section can never be redundant. minLevel is the hint level the
// body's rule carries.
//
// onUnsupported, when non-nil, takes over a specifier the body does not accept
// (`protected` in a record); such a section is neither reported as redundant nor
// allowed to become the visibility in effect.
func (a *Analyzer) checkVisibilitySectionRedundancy(
	sections []ast.RecordVisibilitySection,
	initial string,
	minLevel HintsLevel,
	onUnsupported func(ast.RecordVisibilitySection),
) {
	current := initial
	for _, section := range sections {
		if onUnsupported != nil && section.Specifier == "protected" {
			onUnsupported(section)
			continue
		}
		if section.Specifier != current {
			current = section.Specifier
			continue
		}
		if a.hintsLevelAt(section.Pos) < minLevel {
			continue
		}
		a.addHintAt(section.Pos, "Redundant specifier, visibility is already %q [line: %d, column: %d]",
			section.Specifier, section.Pos.Line, section.Pos.Column)
	}
}

// ============================================================================
// Assigning a variable to itself
// ============================================================================

// checkSelfAssignment hints when both sides of a plain `:=` name the same
// entity. A compound assignment reads the old value before storing, so upstream
// restricts the check to ttASSIGN, and it skips the check entirely once an error
// has been reported ("too sensitive otherwise"). Upstream reads a program in one
// pass, so "reported" there means reported earlier in the source; this analyzer
// defers top-level routine bodies to a second pass, and
// errorsPrecedeCurrentStatement restores that source order.
//
// Scope: both sides must be bare identifiers resolving to the same symbol. That
// covers a local, a parameter, a global and a field reached by implicit Self,
// which is the set the fixtures pin. Upstream also compares two explicitly
// qualified field accesses (`o.F := o.F`, `Self.F := Self.F`) by field symbol
// and receiver; no fixture pins one, so those stay unmeasured here.
func (a *Analyzer) checkSelfAssignment(stmt *ast.AssignmentStatement) {
	if a == nil || a.symbols == nil || stmt == nil {
		return
	}
	if stmt.Operator != lexer.ASSIGN && stmt.Operator != lexer.TokenType(0) {
		return
	}
	if a.parseHadErrors || a.errorsPrecedeCurrentStatement() {
		return
	}
	pos := stmt.Token.Pos
	if pos.Line == 0 || pos.Column == 0 {
		return
	}
	sym, name := a.selfAssignedSymbol(stmt)
	if sym == nil {
		return
	}
	a.addHintAt(pos, "Assigning %s to itself [line: %d, column: %d]",
		qualifiedSymbolName(sym, name), pos.Line, pos.Column)
}

// selfAssignedSymbol returns the symbol both sides of an assignment name, and
// the spelling the left-hand side used, or nil when the two sides are not the
// same bare identifier.
func (a *Analyzer) selfAssignedSymbol(stmt *ast.AssignmentStatement) (*Symbol, string) {
	target, ok := stmt.Target.(*ast.Identifier)
	if !ok {
		return nil, ""
	}
	value, ok := stmt.Value.(*ast.Identifier)
	if !ok || !ident.Equal(target.Value, value.Value) {
		return nil, ""
	}
	// Assigning to the enclosing routine's own name sets the result; it is a
	// write to Result, not to the symbol the right-hand side reads.
	if a.currentFunction != nil && a.currentFunction.Name != nil &&
		ident.Equal(target.Value, a.currentFunction.Name.Value) {
		return nil, ""
	}
	// A routine's implicit Result and a local spelled "result" collapse into one
	// symbol here, where DWScript rejects the redeclaration outright, so a
	// `Result := result` between them is an artifact of that binding rather than
	// an assignment to itself. No fixture pins `Result := Result`, so the whole
	// name is left alone.
	if ident.Equal(target.Value, "Result") {
		return nil, ""
	}
	sym, found := a.symbols.Resolve(target.Value)
	if !found || sym == nil {
		return nil, ""
	}
	if valueSym, ok := a.symbols.Resolve(value.Value); !ok || valueSym != sym {
		return nil, ""
	}
	return sym, target.Value
}

// qualifiedSymbolName spells a symbol the way DWScript's QualifiedName does: a
// field is qualified with its declaring class ("TTest.F"), a local, parameter or
// global stays bare.
func qualifiedSymbolName(sym *Symbol, written string) string {
	name := sym.Name
	if name == "" {
		name = written
	}
	if sym.ClassFieldOwner == nil {
		return name
	}
	return sym.ClassFieldOwner.Name + "." + name
}

// ============================================================================
// Write tracking for var parameters
// ============================================================================

// markStatementLeadingNameWritten records a write against the symbol named at
// the head of a statement.
//
// It mirrors how upstream's symbol dictionary comes by its suWrite entries: a
// statement that starts with a name is read with ReadName(isWrite=True), because
// it may turn out to be an assignment, and that optimistic flag is what
// HintReferenceConstVarParams later reads back. So `o := …` counts as a write,
// and so do `o.Method;` and `o.Field := …` — SimpleScripts/var_param_parent
// expects no hint for a `var` parameter whose only use is `a.Inc;`.
//
// A variable handed on as another routine's `var` argument is recorded by
// markVarArgumentWritten instead, from the call analyzer.
func (a *Analyzer) markStatementLeadingNameWritten(stmt ast.Statement) {
	if a == nil || a.symbols == nil {
		return
	}
	var leading ast.Expression
	switch s := stmt.(type) {
	case *ast.AssignmentStatement:
		leading = s.Target
	case *ast.ExpressionStatement:
		leading = s.Expression
	default:
		return
	}
	name := leadingIdentifier(leading)
	if name == nil {
		return
	}
	if sym, ok := a.symbols.Resolve(name.Value); ok && sym != nil {
		sym.Written = true
	}
}

// markVarArgumentWritten records a write against the variable a call passes as
// another routine's `var` argument.
//
// Upstream reads such an argument with ReadName(isWrite=True) as well, because
// the callee may rebind it, so `Sink(o)` leaves an suWrite entry on `o` exactly
// as `o := …` would. Without it, forwarding a `var` parameter on would draw the
// "never written to" hint even though the routine does give up control of it.
func (a *Analyzer) markVarArgumentWritten(arg ast.Expression) {
	if a == nil || a.symbols == nil {
		return
	}
	name := leadingIdentifier(arg)
	if name == nil {
		return
	}
	if sym, ok := a.symbols.Resolve(name.Value); ok && sym != nil {
		sym.Written = true
	}
}

// leadingIdentifier unwraps member, index and call expressions down to the bare
// name a statement starts with, or nil when it does not start with one.
func leadingIdentifier(expr ast.Expression) *ast.Identifier {
	for {
		switch e := expr.(type) {
		case *ast.Identifier:
			return e
		case *ast.MemberAccessExpression:
			expr = e.Object
		case *ast.IndexExpression:
			expr = e.Left
		case *ast.CallExpression:
			expr = e.Function
		default:
			return nil
		}
	}
}

// ============================================================================
// Reference-type var parameters that are never written
// ============================================================================

// emitReferenceVarParamHints hints on every `var` parameter of a class type that
// the routine body never writes to: passing a reference type by reference is
// only meaningful when the callee rebinds it, so an unwritten one should have
// been a plain value parameter.
//
// Virtual methods are exempt — upstream checks TMethodSymbol.IsVirtual, which an
// override also sets, because a method bound by the hierarchy cannot drop the
// `var`. Routines without a body have nothing to write in.
//
// Scope: only class-typed parameters are reported, matching upstream's
// `param.Typ.IsClassSymbol`. Interfaces and dynamic arrays are reference types
// too but are not class symbols, so they stay out.
func (a *Analyzer) emitReferenceVarParamHints() {
	if !a.symbolDictionaryDiagnosticsEnabled() || a.symbols == nil {
		return
	}
	fn := a.currentFunction
	if fn == nil || fn.Body == nil {
		return
	}
	if fn.IsVirtual || fn.IsOverride || fn.IsAbstract || fn.IsExternal {
		return
	}

	for _, param := range fn.Parameters {
		if !a.isUnwrittenReferenceVarParam(param) {
			continue
		}
		pos := param.Name.Token.Pos
		a.addHintAt(pos, "%q parameter is a reference type passed as VAR, but never written to [line: %d, column: %d]",
			param.Name.Value, pos.Line, pos.Column)
	}
}

// isUnwrittenReferenceVarParam reports whether one parameter of the routine
// being closed is a class-typed `var` parameter its body never wrote to, and the
// hint level at its declaration admits the hint.
func (a *Analyzer) isUnwrittenReferenceVarParam(param *ast.Parameter) bool {
	if param == nil || !param.ByRef || param.IsConst || param.IsLazy || param.Name == nil {
		return false
	}
	pos := param.Name.Token.Pos
	if pos.Line == 0 || pos.Column == 0 {
		return false
	}
	sym, ok := a.symbols.symbols.Get(param.Name.Value)
	if !ok || sym == nil || sym.Written {
		return false
	}
	// The scope must be this routine's own: the unused-warning hook this runs
	// from is deferred twice per body, and the outer run still sees the inner
	// scope while currentFunction has already been restored.
	if !samePosition(sym.DeclPosition, pos) {
		return false
	}
	// Deliberately not unwrapped. Upstream asks `param.Typ.IsClassSymbol`, and
	// GetIsClassSymbol is overridden (final) only on TClassSymbol, so the
	// TAliasSymbol that `type TObjAlias = TObject` installs answers False and the
	// parameter is skipped. Unwrapping here would emit a hint upstream does not.
	// (dwsCompiler.pas:13401 HintReferenceConstVarParams, dwsSymbols.pas:1873/5775.)
	if _, isClass := sym.Type.(*types.ClassType); !isClass {
		return false
	}
	return a.hintsLevelAt(pos) >= HintsLevelPedantic
}

// ============================================================================
// Redundant begin in a case..of else clause
// ============================================================================

// checkRedundantCaseElseBegin hints on a `begin … end` opening the else clause of
// a case..of: the clause already accepts a statement list, so the block adds
// nothing. Upstream tests for the `begin` token the moment it has read `else`,
// before parsing the clause, so the hint fires whenever the clause starts with
// `begin`, however many statements follow, and never for a case *branch*.
//
// The parser folds a single-statement else clause into that statement, so an
// explicit block arrives as the else branch itself; a multi-statement clause
// becomes a synthetic block whose token is the clause's first token. Testing for
// a BEGIN token covers both.
func (a *Analyzer) checkRedundantCaseElseBegin(stmt *ast.CaseStatement) {
	if a == nil || stmt == nil {
		return
	}
	block, ok := stmt.Else.(*ast.BlockStatement)
	if !ok || block.Token.Type != lexer.BEGIN {
		return
	}
	pos := block.Token.Pos
	if pos.Line == 0 || pos.Column == 0 {
		return
	}
	if a.hintsLevelAt(pos) < HintsLevelPedantic {
		return
	}
	a.addHintAt(pos, `Redundant "begin" in clause of a case..of [line: %d, column: %d]`,
		pos.Line, pos.Column)
}

func samePosition(left, right token.Position) bool {
	return left.Line == right.Line && left.Column == right.Column
}
