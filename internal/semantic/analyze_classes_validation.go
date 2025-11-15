package semantic

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// validateMethodSignature validates that an out-of-line method implementation
// signature matches the forward declaration.
func (a *Analyzer) validateMethodSignature(implDecl *ast.FunctionDecl, declaredType *types.FunctionType, className string) error {
	// Resolve parameter types from implementation
	implParamTypes := make([]types.Type, 0, len(implDecl.Parameters))
	for _, param := range implDecl.Parameters {
		if param.Type == nil {
			// DWScript allows omitting parameter types in implementation if they match declaration
			// We'll accept this and rely on the declared types
			continue
		}
		paramType, err := a.resolveType(param.Type.Name)
		if err != nil {
			return fmt.Errorf("unknown parameter type '%s' in method '%s.%s'",
				param.Type.Name, className, implDecl.Name.Value)
		}
		implParamTypes = append(implParamTypes, paramType)
	}

	// If implementation specifies parameter types, validate count matches
	if len(implParamTypes) > 0 && len(implParamTypes) != len(declaredType.Parameters) {
		return fmt.Errorf("method '%s.%s' implementation has %d %s, but declaration has %d %s",
			className, implDecl.Name.Value, len(implParamTypes), pluralizeParam(len(implParamTypes)),
			len(declaredType.Parameters), pluralizeParam(len(declaredType.Parameters)))
	}

	// Validate parameter types match (if implementation specifies them)
	for i, implType := range implParamTypes {
		if i >= len(declaredType.Parameters) {
			break
		}
		declType := declaredType.Parameters[i]
		if !implType.Equals(declType) {
			return fmt.Errorf("method '%s.%s' parameter %d has type %s in implementation, but %s in declaration",
				className, implDecl.Name.Value, i+1, implType.String(), declType.String())
		}
	}

	// Resolve return type from implementation (if specified)
	var implReturnType types.Type
	if implDecl.ReturnType != nil {
		var err error
		implReturnType, err = a.resolveType(implDecl.ReturnType.Name)
		if err != nil {
			return fmt.Errorf("unknown return type '%s' in method '%s.%s'",
				implDecl.ReturnType.Name, className, implDecl.Name.Value)
		}

		// Validate return type matches
		if !implReturnType.Equals(declaredType.ReturnType) {
			return fmt.Errorf("method '%s.%s' has return type %s in implementation, but %s in declaration",
				className, implDecl.Name.Value, implReturnType.String(), declaredType.ReturnType.String())
		}
	}

	return nil
}

// validateVirtualOverride validates virtual/override method declarations
// Task 9.4.1: Updated to support virtual/override on constructors
func (a *Analyzer) validateVirtualOverride(method *ast.FunctionDecl, classType *types.ClassType, methodType *types.FunctionType) {
	methodName := method.Name.Value
	isConstructor := method.IsConstructor

	// If method is marked override, validate parent has virtual method with matching signature
	if method.IsOverride {
		if classType.Parent == nil {
			a.addError("method '%s' marked as override, but class has no parent", methodName)
			return
		}

		// Task 9.4.1: Find matching overload in parent class hierarchy (check both methods and constructors)
		var parentOverload *types.MethodInfo
		if isConstructor {
			parentOverload = a.findMatchingConstructorInParent(methodName, methodType, classType.Parent)
		} else {
			parentOverload = a.findMatchingOverloadInParent(methodName, methodType, classType.Parent)
		}

		if parentOverload == nil {
			// Check if method/constructor with this name exists at all
			var hasParentMember bool
			if isConstructor {
				hasParentMember = a.hasConstructorWithName(methodName, classType.Parent)
			} else {
				hasParentMember = a.hasMethodWithName(methodName, classType.Parent)
			}

			if hasParentMember {
				// Method/constructor name exists but signature doesn't match any parent overload
				a.addError("method '%s' marked as override, but no matching signature exists in parent class", methodName)
			} else {
				// Method/constructor name doesn't exist at all in parent
				a.addError("method '%s' marked as override, but no such method exists in parent class", methodName)
			}
			return
		}

		// Check that parent method/constructor is virtual, override, or abstract
		// Abstract methods are implicitly virtual and can be overridden
		if !parentOverload.IsVirtual && !parentOverload.IsOverride && !parentOverload.IsAbstract {
			a.addError("method '%s' marked as override, but parent method is not virtual", methodName)
			return
		}

		// Task 9.61.4: Add hint if override is part of an overload set but doesn't have overload directive
		// Check if there are other overloads of this method/constructor in the current class
		var currentOverloads []*types.MethodInfo
		if isConstructor {
			currentOverloads = classType.GetConstructorOverloads(methodName)
		} else {
			currentOverloads = classType.GetMethodOverloads(methodName)
		}

		if len(currentOverloads) > 1 && !method.IsOverload {
			a.addHint("Overloaded method \"%s\" should be marked with the \"overload\" directive [line: %d, column: %d]",
				methodName, method.Token.Pos.Line, method.Token.Pos.Column)
		}
	}

	// Warn if redefining virtual method without override or reintroduce keyword
	// Note: Constructors can be marked as virtual, so this check applies to both methods and constructors
	// Task 9.6: Check class metadata instead of AST node, since implementations don't have override keyword
	// Task 9.16.1: Use lowercase key for case-insensitive lookups
	// Task 9.2: Allow reintroduce keyword to explicitly hide virtual parent methods (check class metadata, not AST)
	methodNameLower := strings.ToLower(methodName)
	isOverrideInClass := classType.OverrideMethods[methodNameLower]
	isVirtualInClass := classType.VirtualMethods[methodNameLower]
	isReintroduceInClass := classType.ReintroduceMethods[methodNameLower]
	if !isOverrideInClass && !isVirtualInClass && !isReintroduceInClass && classType.Parent != nil {
		// Task 9.4.1: Check if any parent overload with matching signature is virtual
		var parentOverload *types.MethodInfo
		if isConstructor {
			parentOverload = a.findMatchingConstructorInParent(methodName, methodType, classType.Parent)
		} else {
			parentOverload = a.findMatchingOverloadInParent(methodName, methodType, classType.Parent)
		}

		if parentOverload != nil && (parentOverload.IsVirtual || parentOverload.IsOverride) {
			a.addError("method '%s' hides virtual parent method; use 'override' or 'reintroduce' keyword", methodName)
		}
	}
}

// checkVisibility checks if a member (field or method) is accessible from the current context
// Returns true if accessible, false otherwise.
//
// Visibility rules:
//   - Private: only accessible from the same class
//   - Protected: accessible from the same class and all descendants
//   - Public: accessible from anywhere
//
// Parameters:
//   - memberClass: the class that owns the member
//   - visibility: the visibility level of the member (ast.Visibility as int)
//   - memberName: the name of the member (for error messages)
//   - memberType: "field" or "method" (for error messages)
func (a *Analyzer) checkVisibility(memberClass *types.ClassType, visibility int, _, _ string) bool {
	// Public is always accessible
	if visibility == int(ast.VisibilityPublic) {
		return true
	}

	// If we're analyzing code outside any class context, only public members are accessible
	if a.currentClass == nil {
		return false
	}

	// Private members are only accessible from the same class
	if visibility == int(ast.VisibilityPrivate) {
		return a.currentClass.Name == memberClass.Name
	}

	// Protected members are accessible from the same class and descendants
	if visibility == int(ast.VisibilityProtected) {
		// Same class?
		if a.currentClass.Name == memberClass.Name {
			return true
		}

		// Check if current class inherits from member's class
		return a.isDescendantOf(a.currentClass, memberClass)
	}

	// Should not reach here, but default to false for safety
	return false
}

// validateAbstractClass validates abstract class rules:
// 1. Abstract methods can only exist in abstract classes
// 2. Concrete classes must implement all inherited abstract methods
// 3. Abstract methods are implicitly virtual
func (a *Analyzer) validateAbstractClass(classType *types.ClassType, decl *ast.ClassDecl) {
	// Rule 1: Classes with abstract methods are implicitly abstract
	// If a class has abstract methods, mark it as abstract automatically
	hasAbstractMethods := false
	for _, isAbstract := range classType.AbstractMethods {
		if isAbstract {
			hasAbstractMethods = true
			break
		}
	}
	if hasAbstractMethods {
		classType.IsAbstract = true
	}

	// Abstract methods are implicitly virtual
	for methodName, isAbstract := range classType.AbstractMethods {
		if isAbstract {
			classType.VirtualMethods[methodName] = true
		}
	}

	// Rule 2: Concrete classes must implement all inherited abstract methods
	// NOTE: We don't report this error during class declaration anymore.
	// Instead, we report it during instantiation (see analyzeNewExpression).
	// This matches DWScript behavior where the error is reported when trying to
	// create an instance, not when declaring the class.
	// The check is still performed in analyzeNewExpression via getUnimplementedAbstractMethods.
}
