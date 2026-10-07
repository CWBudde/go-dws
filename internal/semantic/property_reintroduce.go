package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// propertyForCompatibilityCall respects a descendant's ordinary member shadow.
func propertyForCompatibilityCall(class *types.ClassType, name string) *types.PropertyInfo {
	key := ident.Normalize(name)
	for current := class; current != nil; current = current.Parent {
		for propName, prop := range current.Properties {
			if ident.Equal(propName, name) {
				return prop
			}
		}
		if _, exists := current.Methods[key]; exists {
			return nil
		}
		if _, exists := current.Fields[key]; exists {
			return nil
		}
		if _, exists := current.ClassVars[key]; exists {
			return nil
		}
	}
	return nil
}

func (a *Analyzer) analyzePropertyCompatibilityRead(expr *ast.MethodCallExpression, class *types.ClassType) (types.Type, bool) {
	prop := propertyForCompatibilityCall(class, expr.Method.Value)
	if prop == nil || prop.IsIndexed || isFunctionPointerType(prop.Type) {
		return nil, false
	}
	read := &ast.MemberAccessExpression{BaseNode: expr.BaseNode, Object: expr.Object, Member: expr.Method}
	a.addIdentifierCaseHint(expr.Method, prop.Name)
	if !prop.IsReintroduce {
		a.addPunctuationStop(expr.ParenPos, "Not a method")
		return nil, true
	}
	// ReadPropertyExpr consumes a compatibility pair before performing the read.
	a.addHintAt(expr.ParenPos, "Property %q reintroduced a method, you should remove empty brackets () [line: %d, column: %d]", prop.Name, expr.ParenPos.Line, expr.ParenPos.Column)
	if expr.FirstArgumentToken.Type != token.RPAREN {
		a.addStructuredError(NewGenericError(expr.FirstArgumentToken.Pos, `")" expected`))
	}
	result := a.analyzeMemberAccessExpression(read)
	a.semanticInfo.SetPropertyRead(expr, read)
	return result, true
}

func (a *Analyzer) analyzeIncompleteMemberCall(_ *ast.MethodCallExpression) types.Type {
	// The parser retains the authoritative stop, including through enclosing
	// unfinished calls. Never validate a truncated argument list or run end checks.
	a.compileStopped = true
	return nil
}

// analyzeImplicitPropertyCompatibilityRead handles only a directly written empty
// pair. Nonempty lists and malformed boundaries need separate parser recovery.
func (a *Analyzer) analyzeImplicitPropertyCompatibilityRead(call *ast.CallExpression, name *ast.Identifier) (types.Type, bool) {
	if a.currentClass == nil || call.ParenPos.Line == 0 || len(call.Arguments) != 0 {
		return nil, false
	}
	scope := a.implicitPropertyMethodScope(name.Value)
	if scope == nil || a.hasHelperMethod(a.currentImplicitSelfType(), name.Value) != nil {
		return nil, false
	}
	prop := propertyForCompatibilityCall(a.currentClass, name.Value)
	if prop == nil || prop.IsIndexed || isFunctionPointerType(prop.Type) {
		return nil, false
	}
	a.addIdentifierCaseHint(name, prop.Name)
	if scope.classMethodStatic {
		a.addPunctuationStop(call.ParenPos, "Object reference needed to read/write an object field")
		return nil, true
	}
	if !prop.IsReintroduce {
		a.addPunctuationStop(call.ParenPos, "Not a method")
		return nil, true
	}
	a.addHintAt(call.ParenPos, "Property %q reintroduced a method, you should remove empty brackets () [line: %d, column: %d]", prop.Name, call.ParenPos.Line, call.ParenPos.Column)
	a.warnDeprecatedPropertyUsage(prop, name.Token.Pos)
	if prop.ReadKind == types.PropAccessNone {
		a.addStructuredError(NewWriteOnlyPropertyError(name.Token.Pos, name.Value))
		return nil, true
	}
	if a.inClassMethod && !prop.IsClassProperty {
		a.addStructuredError(NewObjectReferenceNeededError(name.Token.Pos))
		return nil, true
	}
	// Runtime Self retains the dynamic metaclass in class methods. Spelling the
	// declaring class as a receiver would lose virtual class-getter dispatch.
	read := &ast.MemberAccessExpression{BaseNode: call.BaseNode, Object: &ast.SelfExpression{BaseNode: name.BaseNode}, Member: name}
	a.semanticInfo.SetImplicitPropertyRead(call, &ast.ImplicitPropertyReadBinding{
		Owner: compatibilityPropertyOwner(a.currentClass, prop), Read: read,
	})
	return prop.Type, true
}

// implicitPropertyMethodScope ignores synthesized fields/class variables so the
// ordinary class hierarchy can decide whether they hide a property. Real local
// bindings and parameters still stop the lookup.
func (a *Analyzer) implicitPropertyMethodScope(name string) *SymbolTable {
	for scope := a.symbols; scope != nil; scope = scope.outer {
		if sym, local := scope.findLocal(name); local && !sym.methodScopeMember {
			return nil
		}
		if scope.classMethodOwner != nil {
			if scope.classMethodOwner == a.currentClass {
				return scope
			}
			return nil
		}
	}
	return nil
}

func compatibilityPropertyOwner(class *types.ClassType, prop *types.PropertyInfo) string {
	for current := class; current != nil; current = current.Parent {
		for _, candidate := range current.Properties {
			if candidate == prop {
				return current.Name
			}
		}
	}
	return ""
}
