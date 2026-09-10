package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// internalUnitName is the namespace prefix DWScript uses for its built-in
// declarations. `Internal.Sin` and `Sin` name the same built-in function, so the
// prefix is stripped before a name is resolved.
const internalUnitName = "Internal"

// analyzeDeclared analyzes a `Declared(s: String): Boolean` call.
//
// Declared is a compile-time intrinsic: the argument must be a constant string
// expression, and the call is folded to the boolean answer during semantic
// analysis. The folded value is published through the semantic metadata table so
// the evaluator can return it without repeating the lookup at run time.
func (a *Analyzer) analyzeDeclared(name string, args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	literal, ok := a.compileTimeStringArgument(name, args, callExpr)
	if !ok {
		a.foldCompileTimePredicate(callExpr, false)
		return types.BOOLEAN
	}
	a.foldCompileTimePredicate(callExpr, a.isDeclaredName(literal))
	return types.BOOLEAN
}

// analyzeConditionalDefined analyzes a `ConditionalDefined(s: String): Boolean`
// call.
//
// Limitation: the conditional symbols established by `{$DEFINE}` live in the
// lexer's preprocessor state, which the analyzer cannot reach. The call is
// therefore always folded to False. The argument is still validated, so a
// non-constant or non-String argument is reported exactly as DWScript does.
func (a *Analyzer) analyzeConditionalDefined(name string, args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	a.compileTimeStringArgument(name, args, callExpr)
	a.foldCompileTimePredicate(callExpr, false)
	return types.BOOLEAN
}

// compileTimeStringArgument validates the single argument of a compile-time
// predicate and returns its constant string value.
//
// Anything that is not a constant string expression is reported as
// "String expected" at the argument's own position, matching DWScript.
func (a *Analyzer) compileTimeStringArgument(name string, args []ast.Expression, callExpr *ast.CallExpression) (string, bool) {
	if len(args) != 1 {
		a.addError("function '%s' expects 1 argument, got %d at %s",
			name, len(args), callExpr.Token.Pos.String())
		for _, arg := range args {
			a.analyzeExpression(arg)
		}
		return "", false
	}

	arg := args[0]
	argType := a.analyzeExpression(arg)
	literal, ok := a.constantStringValue(arg)
	if !ok || (argType != nil && !argType.Equals(types.STRING)) {
		a.addError("String expected at %s", arg.Pos().String())
		return "", false
	}
	return literal, true
}

// constantStringValue returns the compile-time value of a constant string
// expression. String literals and constants declared with a string value are
// recognized; everything else is not a compile-time string.
func (a *Analyzer) constantStringValue(expr ast.Expression) (string, bool) {
	switch node := expr.(type) {
	case *ast.StringLiteral:
		return node.Value, true
	case *ast.Identifier:
		sym, ok := a.symbols.Resolve(node.Value)
		if !ok || !sym.IsConst {
			return "", false
		}
		value, ok := sym.Value.(string)
		return value, ok
	default:
		return "", false
	}
}

// foldCompileTimePredicate records the compile-time result of an intrinsic call
// so the evaluator can return it directly. The value is attached to the callee
// identifier, which is unique per call site.
func (a *Analyzer) foldCompileTimePredicate(callExpr *ast.CallExpression, value bool) {
	if a.semanticInfo == nil || callExpr == nil {
		return
	}
	if callee, ok := callExpr.Function.(*ast.Identifier); ok {
		a.semanticInfo.SetSymbol(callee, value)
	}
}

// isDeclaredName reports whether a possibly dotted name refers to something
// declared in the current compilation: a type, a variable, a function, or a
// member of a type. The lookup is case-insensitive and never emits diagnostics
// or hints.
//
// A bare member name (a record field or a method without its owning type) is
// deliberately invisible, because it is not reachable by that name alone.
func (a *Analyzer) isDeclaredName(name string) bool {
	parts := strings.Split(strings.TrimSpace(name), ".")
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return false
		}
	}

	// `Internal` is the built-in unit's namespace; it qualifies rather than
	// declares, so it is stripped before resolution.
	if len(parts) > 1 && ident.Equal(parts[0], internalUnitName) {
		parts = parts[1:]
	}

	head, rest := parts[0], parts[1:]

	if typ, ok := a.typeRegistry.Resolve(head); ok {
		return a.declaredMemberChain(typ, rest)
	}
	if sym, ok := a.symbols.Resolve(head); ok {
		if len(rest) == 0 {
			return true
		}
		return a.declaredMemberChain(sym.Type, rest)
	}
	if len(rest) == 0 && a.isBuiltinFunction(head) {
		return true
	}
	return false
}

// declaredMemberChain walks a dotted member path starting at typ.
func (a *Analyzer) declaredMemberChain(typ types.Type, parts []string) bool {
	for _, part := range parts {
		memberType, ok := a.declaredMemberType(typ, part)
		if !ok {
			return false
		}
		typ = memberType
	}
	return true
}

// declaredMemberType resolves a single member of a type and returns the member's
// type so that deeper member paths can be followed. The returned type may be nil
// when the member exists but carries no useful type (a constant, for example);
// a further path segment then fails to resolve.
func (a *Analyzer) declaredMemberType(typ types.Type, member string) (types.Type, bool) {
	if typ == nil {
		return nil, false
	}
	switch owner := types.GetUnderlyingType(typ).(type) {
	case *types.ClassType:
		return classMemberType(owner, member)
	case *types.ClassOfType:
		return classMemberType(owner.ClassType, member)
	case *types.RecordType:
		return recordMemberType(owner, member)
	case *types.InterfaceType:
		if method, ok := owner.GetMethod(member); ok {
			return method, true
		}
		if prop := owner.GetProperty(member); prop != nil {
			return prop.Type, true
		}
		return nil, false
	case *types.EnumType:
		for valueName := range owner.Values {
			if ident.Equal(valueName, member) {
				return owner, true
			}
		}
		return nil, false
	case *types.FunctionType:
		// A function's result members are reachable through its return type.
		return a.declaredMemberType(owner.ReturnType, member)
	default:
		return nil, false
	}
}

// classMemberType resolves a member of a class, searching the inheritance chain.
func classMemberType(class *types.ClassType, member string) (types.Type, bool) {
	if class == nil {
		return nil, false
	}
	if fieldType, ok := class.GetField(member); ok {
		return fieldType, true
	}
	if classVarType, ok := class.GetClassVar(member); ok {
		return classVarType, true
	}
	if method, ok := class.GetMethod(member); ok {
		return method, true
	}
	if ctor, ok := class.GetConstructor(member); ok {
		return ctor, true
	}
	if prop, ok := class.GetProperty(member); ok {
		return prop.Type, true
	}
	if class.HasConstant(member) {
		return nil, true
	}
	return nil, false
}

// recordMemberType resolves a member of a record type.
func recordMemberType(record *types.RecordType, member string) (types.Type, bool) {
	if record == nil {
		return nil, false
	}
	if record.HasField(member) {
		return record.GetFieldType(member), true
	}
	if classVarType, ok := record.ClassVars[ident.Normalize(member)]; ok {
		return classVarType, true
	}
	if record.HasMethod(member) {
		return record.GetMethod(member), true
	}
	if record.HasClassMethod(member) {
		return record.GetClassMethod(member), true
	}
	if record.HasProperty(member) {
		if prop := record.GetProperty(member); prop != nil {
			return prop.Type, true
		}
		return nil, true
	}
	for constName := range record.Constants {
		if ident.Equal(constName, member) {
			return nil, true
		}
	}
	return nil, false
}
