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

// synthesizeImplicitParameterlessConstructor generates an implicit parameterless constructor
// when at least one constructor has the 'overload' directive (Task 9.19).
//
// In DWScript, when a constructor is marked with 'overload', the compiler implicitly provides
// a parameterless constructor if one doesn't already exist. This allows code like:
//
//	type TObj = class
//	  constructor Create(x: Integer); overload;
//	end;
//	var o := TObj.Create;  // Calls implicit parameterless constructor
//	var p := TObj.Create(5);  // Calls explicit overload with parameter
func (a *Analyzer) synthesizeImplicitParameterlessConstructor(classType *types.ClassType) {
	// For each constructor name, check if it has the 'overload' directive
	// If so, ensure there's a parameterless overload
	for ctorName, overloads := range classType.ConstructorOverloads {
		hasOverloadDirective := false
		hasParameterlessOverload := false

		// Check if any overload has the 'overload' directive
		// and if a parameterless overload already exists
		for _, methodInfo := range overloads {
			if methodInfo.HasOverloadDirective {
				hasOverloadDirective = true
			}
			if len(methodInfo.Signature.Parameters) == 0 {
				hasParameterlessOverload = true
			}
		}

		// If this constructor set has 'overload' but no parameterless version, synthesize one
		if hasOverloadDirective && !hasParameterlessOverload {
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
			// Mark it as having overload directive to be consistent with other constructors
			methodInfo := &types.MethodInfo{
				Signature:            funcType,
				IsVirtual:            false,
				IsOverride:           false,
				IsAbstract:           false,
				IsForwarded:          false,
				IsClassMethod:        false,
				HasOverloadDirective: true,                      // Mark as part of overload set
				Visibility:           int(ast.VisibilityPublic), // Public access
			}

			// Add to class constructor maps
			lowerName := strings.ToLower(ctorName)
			if _, exists := classType.Constructors[lowerName]; !exists {
				classType.Constructors[lowerName] = funcType
			}
			classType.AddConstructorOverload(ctorName, methodInfo)
		}
	}
}

// checkMethodOverriding checks if overridden methods have compatible signatures.
// Task 9.20.2: Methods with the 'overload' directive add to the overload set rather than
// replacing the parent method, so they don't need signature compatibility checks.
func (a *Analyzer) checkMethodOverriding(class, parent *types.ClassType) {
	// Check MethodOverloads instead of Methods to access overload directive information
	for methodName, childOverloads := range class.MethodOverloads {
		// Check if method exists in parent
		parentMethodType, found := parent.GetMethod(methodName)
		if !found {
			// New method in child class - OK
			continue
		}

		// Check each overload in the child class
		for _, childMethod := range childOverloads {
			// Task 9.20.2: If the child method has the 'overload' directive,
			// it's adding to the overload set, not replacing/overriding the parent.
			// So skip signature compatibility check.
			if childMethod.HasOverloadDirective {
				continue
			}

			// No 'overload' directive - this is a method override/hide.
			// Check signature compatibility with parent.
			childMethodType := childMethod.Signature
			if !childMethodType.Equals(parentMethodType) {
				a.addError("method '%s' signature mismatch in class '%s': expected %s, got %s",
					methodName, class.Name, parentMethodType.String(), childMethodType.String())
			}
		}
	}
}
