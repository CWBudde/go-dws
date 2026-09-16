package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// addIdentifierCaseHint reports a resolved spelling once per source identifier.
// Overload selection may analyze an argument more than once; AST identity keeps
// those visits distinct from separate uses at identical positions in includes.
func (a *Analyzer) addIdentifierCaseHint(identifier *ast.Identifier, declared string) {
	if identifier == nil || identifier.Value == declared || declared == "" || !ident.Equal(identifier.Value, declared) {
		return
	}
	if a.caseHintIdentifiers == nil {
		a.caseHintIdentifiers = make(map[*ast.Identifier]bool)
	}
	if a.caseHintIdentifiers[identifier] {
		return
	}
	a.caseHintIdentifiers[identifier] = true
	a.addCaseMismatchHint(identifier.Value, declared, identifier.Token.Pos)
}

func (a *Analyzer) declaredClassMemberName(classType *types.ClassType, name string) string {
	key := ident.Normalize(name)
	for current := classType; current != nil; current = current.Parent {
		if declared := current.FieldDeclNames[key]; declared != "" {
			return declared
		}
		if declared := current.ClassVarDeclNames[key]; declared != "" {
			return declared
		}
		if prop, ok := current.Properties[key]; ok {
			return prop.Name
		}
		if declared := current.MethodDeclNames[key]; declared != "" {
			return declared
		}
	}
	return ""
}

// declaredHelperMethodName follows the same helper precedence as method lookup.
func (a *Analyzer) declaredHelperMethodName(typ types.Type, name string) string {
	helpers := a.getHelpersForType(typ)
	key := ident.Normalize(name)
	for i := len(helpers) - 1; i >= 0; i-- {
		// Intrinsic members carry an operation but not always a signature, so
		// BuiltinMethods is consulted alongside the declared method table.
		_, isMethod := helpers[i].Methods[key]
		_, isBuiltin := helpers[i].BuiltinMethods[key]
		if isMethod || isBuiltin {
			return helpers[i].MethodDeclNames[key]
		}
	}
	return ""
}

// declaredInterfaceMethodName returns the source spelling of an interface
// method, following the parent chain the same way method lookup does.
func (a *Analyzer) declaredInterfaceMethodName(iface *types.InterfaceType, name string) string {
	key := ident.Normalize(name)
	for current := iface; current != nil; current = current.Parent {
		if declared := current.MethodDeclNames[key]; declared != "" {
			return declared
		}
	}
	return ""
}

// declaredRecordMethodName returns the source spelling of a record method,
// preferring instance methods over class methods as method lookup does.
func (a *Analyzer) declaredRecordMethodName(recordType *types.RecordType, name string) string {
	if recordType == nil {
		return ""
	}
	key := ident.Normalize(name)
	if declared := recordType.MethodNames[key]; declared != "" {
		return declared
	}
	return recordType.ClassMethodNames[key]
}
