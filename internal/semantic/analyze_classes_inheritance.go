package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Class Inheritance Analysis
// ============================================================================

// inheritParentConstructors copies accessible parent constructors to a child class (Task 9.2).
// In DWScript, child classes inherit parent constructors if the child doesn't declare any.
// Only public and protected constructors are inherited (private constructors are not).
func (a *Analyzer) inheritParentConstructors(childClass *types.ClassType, parentClass *types.ClassType) {
	// Iterate through all parent constructor overloads
	for ctorName, overloads := range parentClass.ConstructorOverloads {
		for _, parentCtor := range overloads {
			// Task 9.2: Private constructors are not inherited
			visibility := ast.Visibility(parentCtor.Visibility)
			if visibility == ast.VisibilityPrivate {
				continue
			}

			// Copy the constructor signature, updating return type to child class
			// Create new function type with same parameters but child class return type
			childCtorType := types.NewFunctionTypeWithMetadata(
				parentCtor.Signature.Parameters,
				parentCtor.Signature.ParamNames,
				parentCtor.Signature.DefaultValues,
				parentCtor.Signature.LazyParams,
				parentCtor.Signature.VarParams,
				parentCtor.Signature.ConstParams,
				childClass, // Returns instance of the child class, not parent
			)

			// Create method info for the inherited constructor
			childCtorInfo := &types.MethodInfo{
				Signature:            childCtorType,
				IsVirtual:            parentCtor.IsVirtual,
				IsOverride:           false, // Inherited constructors are not marked as override
				IsAbstract:           parentCtor.IsAbstract,
				IsForwarded:          false,
				IsClassMethod:        parentCtor.IsClassMethod,
				HasOverloadDirective: parentCtor.HasOverloadDirective,
				Visibility:           parentCtor.Visibility,
			}

			// Add inherited constructor to child class
			// Use lowercase for case-insensitive lookup
			lowerCtorName := strings.ToLower(ctorName)
			if _, exists := childClass.Constructors[lowerCtorName]; !exists {
				childClass.Constructors[lowerCtorName] = childCtorType
			}
			childClass.AddConstructorOverload(ctorName, childCtorInfo)
		}
	}

	// If no constructors were inherited (all were private), generate implicit default
	if len(childClass.Constructors) == 0 && len(childClass.ConstructorOverloads) == 0 {
		a.synthesizeDefaultConstructor(childClass)
	}
}

// synthesizeDefaultConstructor generates an implicit default constructor for a class (Task 9.1).
// DWScript automatically provides a parameterless `Create` constructor for classes
// that don't declare any explicit constructors.
func (a *Analyzer) synthesizeDefaultConstructor(classType *types.ClassType) {
	constructorName := "Create"

	// Create function type: no parameters, returns the class type
	funcType := types.NewFunctionTypeWithMetadata(
		[]types.Type{},  // No parameters
		[]string{},      // No parameter names
		[]interface{}{}, // No default values
		[]bool{},        // No lazy params
		[]bool{},        // No var params
		[]bool{},        // No const params
		classType,       // Returns instance of the class
	)

	// Create method info for the implicit constructor
	methodInfo := &types.MethodInfo{
		Signature:            funcType,
		IsVirtual:            false,
		IsOverride:           false,
		IsAbstract:           false,
		IsForwarded:          false,
		IsClassMethod:        false,
		HasOverloadDirective: false,
		Visibility:           int(ast.VisibilityPublic), // Public access
	}

	// Add to class constructor maps
	// Use lowercase for case-insensitive lookup
	classType.Constructors[strings.ToLower(constructorName)] = funcType
	classType.AddConstructorOverload(constructorName, methodInfo)
}

// checkMethodOverriding checks if overridden methods have compatible signatures
func (a *Analyzer) checkMethodOverriding(class, parent *types.ClassType) {
	for methodName, childMethodType := range class.Methods {
		// Check if method exists in parent
		parentMethodType, found := parent.GetMethod(methodName)
		if !found {
			// New method in child class - OK
			continue
		}

		// Method exists in parent - check signature compatibility
		if !childMethodType.Equals(parentMethodType) {
			a.addError("method '%s' signature mismatch in class '%s': expected %s, got %s",
				methodName, class.Name, parentMethodType.String(), childMethodType.String())
		}
	}
}
