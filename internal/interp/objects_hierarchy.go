package interp

import (
	"github.com/cwbudde/go-dws/pkg/ast"
	pkgident "github.com/cwbudde/go-dws/pkg/ident"
)

// lookupConstructorOverloadsInHierarchy returns all constructor overloads with the given name
// by searching the class hierarchy. Case-insensitive.
func (i *Interpreter) lookupConstructorOverloadsInHierarchy(classInfo *ClassInfo, name string) []*ast.FunctionDecl {
	for current := classInfo; current != nil; current = current.Parent {
		for ctorName, overloads := range current.ConstructorOverloads {
			if pkgident.Equal(ctorName, name) && len(overloads) > 0 {
				return overloads
			}
		}
		for ctorName, constructor := range current.Constructors {
			if pkgident.Equal(ctorName, name) {
				return []*ast.FunctionDecl{constructor}
			}
		}
	}
	return nil
}

// lookupClassMethodInHierarchy searches for a class method by name in the class hierarchy.
// Case-insensitive.
func (i *Interpreter) lookupClassMethodInHierarchy(classInfo *ClassInfo, name string) *ast.FunctionDecl {
	normalizedName := pkgident.Normalize(name)
	for current := classInfo; current != nil; current = current.Parent {
		if method, exists := current.ClassMethods[normalizedName]; exists {
			return method
		}
	}
	return nil
}

// bindClassConstantsToEnv adds all class constants to the current environment,
// allowing methods to access them directly without qualification.
func (i *Interpreter) bindClassConstantsToEnv(classInfo *ClassInfo) {
	for constName, constValue := range classInfo.ConstantValues {
		i.Env().Define(constName, constValue)
	}
}

// getClassConstant retrieves and caches a class constant value by name.
// Evaluates lazily on first access and supports inheritance.
func (i *Interpreter) getClassConstant(classInfo *ClassInfo, constantName string) Value {
	constDecl, ownerClass := classInfo.lookupConstant(constantName)
	if constDecl == nil {
		return nil
	}

	if cachedValue, cached := ownerClass.ConstantValues[constantName]; cached {
		return cachedValue
	}

	// Evaluate constant in temporary environment with other evaluated constants
	defer i.PushScope()()

	for constName, constVal := range ownerClass.ConstantValues {
		if constName != constantName && constVal != nil {
			i.Env().Define(constName, constVal)
		}
	}

	constValue := i.Eval(constDecl.Value)

	if isError(constValue) {
		return constValue
	}

	ownerClass.ConstantValues[constantName] = constValue
	return constValue
}
