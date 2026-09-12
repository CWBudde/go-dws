package evaluator

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// resolveParentClass resolves a class declaration's ancestor: the named parent,
// or the implicit TObject when none is named. It returns the runtime class, the
// name it was found under, and an error value when the ancestor is unknown.
func (e *Evaluator) resolveParentClass(
	node *ast.ClassDecl,
	className string,
	ctx *ExecutionContext,
) (parentClass any, parentClassName string, errValue Value) {
	if node.Parent == nil {
		// Implicit TObject inheritance, unless this IS TObject or is external.
		if ident.Equal(className, "TObject") || node.IsExternal {
			return nil, "", nil
		}
		parent := e.typeSystem.LookupClass("TObject")
		if parent == nil {
			return nil, "", e.newError(node, "implicit parent class 'TObject' not found")
		}
		return parent, "TObject", nil
	}

	name := node.Parent.Value
	parent := e.typeSystem.LookupClass(name)
	if parent == nil {
		// The parent may be named through an alias (`type TMyControl =
		// TObject;`), which the class registry does not carry. Resolve the
		// alias and look the underlying class up under its own name.
		if resolved := e.resolveClassAliasName(name, ctx); resolved != "" {
			name = resolved
			parent = e.typeSystem.LookupClass(name)
		}
	}
	if parent == nil {
		return nil, "", e.newError(node, "parent class '%s' not found", node.Parent.Value)
	}
	return parent, name, nil
}

// A type alias for a class names that class everywhere a class name is
// accepted: as a parent, as a metaclass operand, and as a static receiver. The
// semantic analyzer resolves aliases through the type registry, but the
// runtime class registry is keyed by the declared class name, so the evaluator
// has to translate an alias before it can look a class up.

// resolveClassAliasName resolves a type name that may be an alias for a class
// and returns the underlying class's own name, or "" when the name is not an
// alias for a class. The runtime class registry is keyed by the declared class
// name, so an alias has to be translated before a lookup can succeed.
func (e *Evaluator) resolveClassAliasName(name string, ctx *ExecutionContext) string {
	resolved, err := e.resolveTypeName(name, ctx)
	if err != nil || resolved == nil {
		return ""
	}
	classType, ok := types.GetUnderlyingType(resolved).(*types.ClassType)
	if !ok || classType.Name == "" || ident.Equal(classType.Name, name) {
		return ""
	}
	return classType.Name
}

// attachImplementedInterfaces registers every interface a class declaration
// names, reporting the first unknown one.
func (e *Evaluator) attachImplementedInterfaces(node *ast.ClassDecl, classInfo classDeclarationInfo) Value {
	for _, ifaceIdent := range node.Interfaces {
		ifaceName := ifaceIdent.Value
		iface := e.typeSystem.LookupInterface(ifaceName)
		if iface == nil {
			return e.newError(node, "interface '%s' not found", ifaceName)
		}
		classInfo.AddImplementedInterface(iface, ifaceName)
	}
	return nil
}
