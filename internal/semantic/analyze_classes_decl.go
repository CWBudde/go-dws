package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// ============================================================================
// Class Declaration Analysis
// ============================================================================

// isForwardDeclaration checks if a class declaration is a forward declaration (has no body).
func (a *Analyzer) isForwardDeclaration(decl *ast.ClassDecl) bool {
	return decl.Fields == nil &&
		decl.Methods == nil &&
		decl.Properties == nil &&
		decl.Operators == nil &&
		decl.Constants == nil
}

// handleExistingClass handles conflicts when a class is redeclared.
// It returns flags indicating if it's resolving a forward declaration, merging a partial class, or if analysis should stop.
func (a *Analyzer) handleExistingClass(
	existingClass *types.ClassType,
	decl *ast.ClassDecl,
	className string,
	isForwardDecl bool,
) (bool, bool, bool) {
	if existingClass == nil {
		return false, false, false
	}

	resolvingForwardDecl := false
	mergingPartialClass := false

	// Handle different redeclaration scenarios.
	if existingClass.IsPartial && decl.IsPartial {
		mergingPartialClass = true
		if !a.validatePartialClassParent(existingClass, decl, className) {
			return false, false, true
		}
	} else if existingClass.IsPartial && !decl.IsPartial && !isForwardDecl {
		a.addHint("Previous declaration of class was \"partial\" at %s", decl.Token.Pos.String())
		mergingPartialClass = true
	} else if !existingClass.IsPartial && decl.IsPartial {
		a.addError("%s", errors.FormatTypeAlreadyDefined(className, "Class", decl.Token.Pos.Line, decl.Token.Pos.Column))
		return false, false, true
	} else if existingClass.IsForward && !isForwardDecl {
		if !a.validateForwardDeclParent(existingClass, decl, className) {
			return false, false, true
		}
		resolvingForwardDecl = true
	} else if existingClass.IsForward && isForwardDecl {
		a.addError("%s", errors.FormatTypeAlreadyDefined(className, "Class", decl.Token.Pos.Line, decl.Token.Pos.Column))
		return false, false, true
	} else {
		a.addError("%s", errors.FormatTypeAlreadyDefined(className, "Class", decl.Token.Pos.Line, decl.Token.Pos.Column))
		return false, false, true
	}

	return resolvingForwardDecl, mergingPartialClass, false
}

// validatePartialClassParent ensures partial class declarations have matching parents.
func (a *Analyzer) validatePartialClassParent(
	existingClass *types.ClassType,
	decl *ast.ClassDecl,
	className string,
) bool {
	if decl.Parent != nil && existingClass.Parent != nil {
		if !ident.Equal(decl.Parent.Value, existingClass.Parent.Name) {
			a.addError("partial class '%s' has conflicting parent classes at %s",
				className, decl.Token.Pos.String())
			return false
		}
	}
	return true
}

// validateForwardDeclParent ensures a forward declaration and its implementation have matching parents.
func (a *Analyzer) validateForwardDeclParent(
	existingClass *types.ClassType,
	decl *ast.ClassDecl,
	className string,
) bool {
	var fullImplParent *types.ClassType
	if decl.Parent != nil {
		parentName := decl.Parent.Value
		fullImplParent = a.getClassType(parentName)
		if fullImplParent == nil {
			a.addError("parent class '%s' not found at %s", parentName, decl.Token.Pos.String())
			return false
		}
	}

	// If forward decl had a parent, implementation must match.
	if existingClass.Parent != nil {
		if fullImplParent == nil {
			a.addError("class '%s' forward declared with parent '%s', but implementation has no parent at %s",
				className, existingClass.Parent.Name, decl.Token.Pos.String())
			return false
		} else if existingClass.Parent.Name != fullImplParent.Name {
			a.addError("class '%s' forward declared with parent '%s', but implementation specifies different parent '%s' at %s",
				className, existingClass.Parent.Name, fullImplParent.Name, decl.Token.Pos.String())
			return false
		}
	}
	return true
}

// createForwardDeclaration creates a minimal class type for a forward declaration.
func (a *Analyzer) createForwardDeclaration(decl *ast.ClassDecl, className string) {
	var parentClass *types.ClassType
	if decl.Parent != nil {
		parentName := decl.Parent.Value
		parentClass = a.getClassType(parentName)
		if parentClass == nil {
			a.addError("parent class '%s' not found at %s", parentName, decl.Token.Pos.String())
			return
		}
	}

	classType := types.NewClassType(className, parentClass)
	classType.IsForward = true
	classType.IsAbstract = decl.IsAbstract
	classType.IsExternal = decl.IsExternal
	classType.ExternalName = decl.ExternalName

	a.registerTypeWithPos(className, classType, decl.Token.Pos)
}

// initializeClassType initializes or reuses a class type, resolving its parent.
// Returns the class type and its parent. Returns nil classType on error.
func (a *Analyzer) initializeClassType(
	existingClass *types.ClassType,
	decl *ast.ClassDecl,
	className string,
	resolvingForwardDecl bool,
	mergingPartialClass bool,
) (*types.ClassType, *types.ClassType) {
	var parentClass *types.ClassType
	var classType *types.ClassType

	if resolvingForwardDecl || mergingPartialClass {
		// Reuse existing class type.
		classType = existingClass
		parentClass = classType.Parent

		// Update parent if specified in a partial declaration.
		if decl.Parent != nil && parentClass == nil {
			parentName := decl.Parent.Value
			parentClass = a.getClassType(parentName)
			if parentClass == nil {
				a.addError("parent class '%s' not found at %s", parentName, decl.Token.Pos.String())
				return nil, nil
			}
			classType.Parent = parentClass
		}

		// Handle implicit TObject parent.
		if parentClass == nil && !ident.Equal(className, "TObject") && !decl.IsExternal {
			parentClass = a.getClassType("TObject")
			if parentClass == nil {
				a.addError("implicit parent class 'TObject' not found at %s", decl.Token.Pos.String())
				return nil, nil
			}
			classType.Parent = parentClass
		}
	} else {
		// Create a new class.
		parentClass = a.resolveParentClass(decl, className)
		if parentClass == nil && decl.Parent != nil {
			return nil, nil
		}
		classType = types.NewClassType(className, parentClass)
	}

	return classType, parentClass
}

// resolveParentClass finds the parent class for a new class declaration.
// Returns nil if no parent, which is valid for TObject or external classes.
func (a *Analyzer) resolveParentClass(decl *ast.ClassDecl, className string) *types.ClassType {
	if decl.Parent != nil {
		parentName := decl.Parent.Value
		parentClass := a.getClassType(parentName)
		if parentClass == nil {
			a.addError("parent class '%s' not found at %s", parentName, decl.Token.Pos.String())
			return nil
		}
		return parentClass
	}

	// Implicitly inherit from TObject if no explicit parent.
	if !ident.Equal(className, "TObject") && !decl.IsExternal {
		parentClass := a.getClassType("TObject")
		if parentClass == nil {
			a.addError("implicit parent class 'TObject' not found at %s", decl.Token.Pos.String())
			return nil
		}
		return parentClass
	}

	return nil
}

// setupNestedTypes builds a map of nested type aliases and analyzes nested declarations.
func (a *Analyzer) setupNestedTypes(decl *ast.ClassDecl, className string) {
	nestedAliases := a.buildNestedAliasMap(decl)
	a.nestedTypeAliases[ident.Normalize(className)] = nestedAliases

	for _, nested := range decl.NestedTypes {
		a.analyzeStatement(nested)
	}

	a.currentNestedTypes = nestedAliases
}

// updateClassFlags updates flags for partial, abstract, and external classes.
func (a *Analyzer) updateClassFlags(classType *types.ClassType, decl *ast.ClassDecl, isForwardDecl bool) {
	classType.IsForward = false

	if decl.IsPartial {
		classType.IsPartial = true
	} else if !isForwardDecl {
		classType.IsPartial = false
	}

	classType.IsAbstract = decl.IsAbstract || classType.IsAbstract
	classType.IsExternal = decl.IsExternal || classType.IsExternal
	if decl.ExternalName != "" {
		classType.ExternalName = decl.ExternalName
	}
}

// validateClassInheritance checks for valid inheritance rules (e.g., external, circular).
func (a *Analyzer) validateClassInheritance(
	classType *types.ClassType,
	parentClass *types.ClassType,
	decl *ast.ClassDecl,
	className string,
) bool {
	// Validate external class inheritance rules.
	if decl.IsExternal {
		if parentClass != nil && !parentClass.IsExternal {
			a.addError("external class '%s' cannot inherit from non-external class '%s' at %s",
				className, parentClass.Name, decl.Token.Pos.String())
			return false
		}
	} else {
		if parentClass != nil && parentClass.IsExternal {
			a.addError("non-external class '%s' cannot inherit from external class '%s' at %s",
				className, parentClass.Name, decl.Token.Pos.String())
			return false
		}
	}

	// Check for circular inheritance.
	if parentClass != nil && a.hasCircularInheritance(classType) {
		a.addError("circular inheritance detected in class '%s' at %s", className, decl.Token.Pos.String())
		return false
	}

	return true
}

// classFullName returns the fully qualified name of a class, including its enclosing class.
func classFullName(decl *ast.ClassDecl) string {
	if decl == nil || decl.Name == nil {
		return ""
	}
	if decl.EnclosingClass != nil && decl.EnclosingClass.Value != "" {
		return decl.EnclosingClass.Value + "." + decl.Name.Value
	}
	return decl.Name.Value
}

// buildNestedAliasMap creates a map of simple names to fully qualified names for nested types.
func (a *Analyzer) buildNestedAliasMap(decl *ast.ClassDecl) map[string]string {
	aliases := make(map[string]string)
	outer := classFullName(decl)
	for _, nested := range decl.NestedTypes {
		a.collectNestedAliases(aliases, nested, outer)
	}
	return aliases
}

// collectNestedAliases recursively builds the alias map from nested type statements.
func (a *Analyzer) collectNestedAliases(aliases map[string]string, stmt ast.Statement, outer string) {
	switch n := stmt.(type) {
	case *ast.BlockStatement:
		for _, inner := range n.Statements {
			a.collectNestedAliases(aliases, inner, outer)
		}
	case *ast.ClassDecl:
		enclosing := outer
		if n.EnclosingClass != nil && n.EnclosingClass.Value != "" {
			enclosing = n.EnclosingClass.Value
		}
		aliases[ident.Normalize(n.Name.Value)] = enclosing + "." + n.Name.Value
	}
}

// analyzeClassDecl analyzes a class declaration.
func (a *Analyzer) analyzeClassDecl(decl *ast.ClassDecl) {
	className := classFullName(decl)
	isForwardDecl := a.isForwardDeclaration(decl)

	// Handle existing class declarations (forward/partial).
	existingClass := a.getClassType(className)
	resolvingForwardDecl, mergingPartialClass, shouldReturn := a.handleExistingClass(
		existingClass, decl, className, isForwardDecl,
	)
	if shouldReturn {
		return
	}

	// Create a forward declaration if applicable.
	if isForwardDecl {
		a.createForwardDeclaration(decl, className)
		return
	}

	// Initialize or reuse the class type.
	classType, parentClass := a.initializeClassType(
		existingClass, decl, className, resolvingForwardDecl, mergingPartialClass,
	)
	if classType == nil {
		return
	}

	// Setup nested types, flags, and validate inheritance.
	a.setupNestedTypes(decl, className)
	defer func() { a.currentNestedTypes = nil }()
	a.updateClassFlags(classType, decl, isForwardDecl)
	if !a.validateClassInheritance(classType, parentClass, decl, className) {
		return
	}

	// Analyze constants in two passes to allow forward references.
	constantNames := make(map[string]bool)
	type constInfo struct {
		decl     *ast.ConstDecl
		explType types.Type
	}
	constList := make([]*constInfo, 0, len(decl.Constants))

	// First pass: Register constant names and resolve explicit types.
	for _, constant := range decl.Constants {
		constantName := constant.Name.Value
		if _, exists := classType.Constants[constantName]; exists || constantNames[constantName] {
			a.addError("%s", errors.FormatNameAlreadyExists(constantName, constant.Token.Pos.Line, constant.Token.Pos.Column))
			continue
		}
		constantNames[constantName] = true

		info := &constInfo{decl: constant}
		if constant.Type != nil {
			typeName := getTypeExpressionName(constant.Type)
			var err error
			info.explType, err = a.resolveType(typeName)
			if err != nil {
				a.addError("unknown type '%s' for constant '%s' at %s", typeName, constantName, constant.Token.Pos.String())
				continue
			}
		}
		// Register constant with a placeholder type.
		classType.Constants[constantName] = nil
		if info.explType != nil {
			classType.ConstantTypes[constantName] = info.explType
		}
		classType.ConstantVisibility[constantName] = int(constant.Visibility)
		constList = append(constList, info)
	}

	// Register the class if it's new.
	if !mergingPartialClass && !resolvingForwardDecl {
		a.registerTypeWithPos(className, classType, decl.Token.Pos)
	}

	// Set the current class context for member analysis.
	previousClass := a.currentClass
	a.currentClass = classType
	defer func() { a.currentClass = previousClass }()

	// Second pass: Analyze constant values.
	for _, info := range constList {
		constant := info.decl
		constantName := constant.Name.Value
		constType := info.explType

		if constType == nil && constant.Value != nil {
			constType = a.analyzeExpression(constant.Value)
			if constType == nil {
				a.addError("unable to determine type for constant '%s' at %s", constantName, constant.Token.Pos.String())
				continue
			}
			classType.ConstantTypes[constantName] = constType
		} else if constType == nil {
			a.addError("constant '%s' must have a value or type annotation at %s", constantName, constant.Token.Pos.String())
			continue
		}
	}

	// Analyze fields (instance and class variables).
	fieldNames := make(map[string]bool)
	classVarNames := make(map[string]bool)
	for _, field := range decl.Fields {
		originalFieldName := field.Name.Value
		normalizedFieldName := ident.Normalize(originalFieldName)

		if field.IsClassVar {
			// Check for duplicate class variable names (case-insensitive)
			// When merging partial classes, check if already exists in ClassType
			_, existsInClass := classType.ClassVars[normalizedFieldName]
			if existsInClass {
				a.addError("%s", errors.FormatNameAlreadyExists(originalFieldName, field.Token.Pos.Line, field.Token.Pos.Column))
				continue
			}
			if classVarNames[normalizedFieldName] {
				a.addError("%s", errors.FormatNameAlreadyExists(originalFieldName, field.Token.Pos.Line, field.Token.Pos.Column))
				continue
			}
			classVarNames[normalizedFieldName] = true

			var fieldType types.Type
			if field.Type != nil {
				typeName := getTypeExpressionName(field.Type)
				resolvedType, err := a.resolveType(typeName)
				if err != nil {
					a.addError("unknown type '%s' for class var '%s' at %s", typeName, originalFieldName, field.Token.Pos.String())
					continue
				}
				fieldType = resolvedType
			} else if field.InitValue != nil {
				initType := a.analyzeExpression(field.InitValue)
				if initType == nil {
					a.addError("cannot infer type for class var '%s' at %s", originalFieldName, field.Token.Pos.String())
					continue
				}
				fieldType = initType
			} else {
				a.addError("class var '%s' missing type annotation", originalFieldName)
				continue
			}

			if field.InitValue != nil && field.Type != nil {
				initType := a.analyzeExpression(field.InitValue)
				if initType != nil && fieldType != nil && !types.IsAssignableFrom(fieldType, initType) {
					a.addError("type mismatch for class var '%s' at %s", originalFieldName, field.Token.Pos.String())
				}
			}

			classType.ClassVars[normalizedFieldName] = fieldType
			classType.ClassVarVisibility[normalizedFieldName] = int(field.Visibility)
		} else {
			// Handle instance fields.
			fieldExists := false
			for existingName := range classType.Fields {
				if ident.Equal(existingName, normalizedFieldName) {
					fieldExists = true
					break
				}
			}
			if fieldExists || fieldNames[normalizedFieldName] {
				a.addError("%s", errors.FormatNameAlreadyExists(originalFieldName, field.Token.Pos.Line, field.Token.Pos.Column))
				continue
			}
			fieldNames[normalizedFieldName] = true

			var fieldType types.Type
			if field.Type != nil {
				typeName := getTypeExpressionName(field.Type)
				resolvedType, err := a.resolveType(typeName)
				if err != nil {
					a.addError("unknown type '%s' for field '%s' at %s", typeName, originalFieldName, field.Token.Pos.String())
					continue
				}
				fieldType = resolvedType
			} else if field.InitValue != nil {
				initType := a.analyzeExpression(field.InitValue)
				if initType == nil {
					a.addError("cannot infer type for field '%s' at %s", originalFieldName, field.Token.Pos.String())
					continue
				}
				fieldType = initType
			} else {
				a.addError("field '%s' missing type annotation", originalFieldName)
				continue
			}

			a.validateFieldInitializer(field, originalFieldName, fieldType)
			classType.Fields[originalFieldName] = fieldType
			classType.FieldVisibility[normalizedFieldName] = int(field.Visibility)
		}
	}

	// Analyze methods and constructors.
	for _, method := range decl.Methods {
		a.analyzeMethodDecl(method, classType)
	}
	if decl.Constructor != nil {
		a.analyzeMethodDecl(decl.Constructor, classType)
	}

	// Handle constructor inheritance and implicit constructors.
	if len(classType.Constructors) == 0 && len(classType.ConstructorOverloads) == 0 {
		if parentClass != nil && len(parentClass.Constructors) > 0 {
			a.inheritParentConstructors(classType, parentClass)
		} else {
			a.synthesizeDefaultConstructor(classType)
		}
	}
	a.synthesizeImplicitParameterlessConstructor(classType)

	// Analyze properties, operators, and validate class structure.
	for _, property := range decl.Properties {
		a.analyzePropertyDecl(property, classType)
	}
	a.registerClassOperators(classType, decl)
	if parentClass != nil {
		a.checkMethodOverriding(classType, parentClass)
	}
	if len(decl.Interfaces) > 0 {
		a.validateInterfaceImplementation(classType, decl)
	}
	a.validateAbstractClass(classType, decl)
}

// analyzeMethodImplementation analyzes an out-of-line method implementation.
func (a *Analyzer) analyzeMethodImplementation(decl *ast.FunctionDecl) {
	typeName := decl.ClassName.Value

	// Check if it's a class method.
	if classType := a.getClassType(typeName); classType != nil {
		a.analyzeClassMethodImplementation(decl, classType, typeName)
		return
	}

	// Check if it's a record method.
	if recordType := a.getRecordType(typeName); recordType != nil {
		a.analyzeRecordMethodImplementation(decl, recordType, typeName)
		return
	}

	a.addError("unknown type '%s' at %s", typeName, decl.Token.Pos.String())
}

// analyzeClassMethodImplementation handles out-of-line class method implementations.
func (a *Analyzer) analyzeClassMethodImplementation(decl *ast.FunctionDecl, classType *types.ClassType, className string) {
	methodName := decl.Name.Value
	var declaredMethod *types.FunctionType
	var methodExists bool

	// Find the declared method, handling overloads.
	if decl.IsConstructor {
		overloads := classType.GetConstructorOverloads(methodName)
		if len(overloads) > 0 {
			declaredMethod, methodExists = a.findMatchingOverloadForImplementation(decl, overloads, className)
		}
	} else {
		overloads := classType.GetMethodOverloads(methodName)
		if len(overloads) > 0 {
			declaredMethod, methodExists = a.findMatchingOverloadForImplementation(decl, overloads, className)
		} else {
			declaredMethod, methodExists = classType.Methods[methodName]
		}
	}

	if !methodExists {
		a.addError("method '%s' not declared in class '%s' at %s", methodName, className, decl.Token.Pos.String())
		return
	}

	// Set up the class context for analysis.
	previousClass := a.currentClass
	a.currentClass = classType
	prevNested := a.currentNestedTypes
	if aliases, ok := a.nestedTypeAliases[ident.Normalize(decl.ClassName.Value)]; ok {
		a.currentNestedTypes = aliases
	} else if aliases, ok := a.nestedTypeAliases[ident.Normalize(classType.Name)]; ok {
		a.currentNestedTypes = aliases
	}
	defer func() {
		a.currentClass = previousClass
		a.currentNestedTypes = prevNested
	}()

	// Validate method signature.
	isOverloaded := len(classType.GetMethodOverloads(methodName)) > 1 || len(classType.GetConstructorOverloads(methodName)) > 1
	if !isOverloaded {
		if err := a.validateMethodSignature(decl, declaredMethod, className); err != nil {
			a.addError("%s at %s", err.Error(), decl.Token.Pos.String())
			return
		}
	}

	// Mark the forward-declared method as implemented.
	delete(classType.ForwardedMethods, ident.Normalize(methodName))

	// Analyze the method body.
	a.analyzeMethodDecl(decl, classType)
}

// analyzeRecordMethodImplementation handles out-of-line record method implementations.
func (a *Analyzer) analyzeRecordMethodImplementation(decl *ast.FunctionDecl, recordType *types.RecordType, recordName string) {
	methodName := decl.Name.Value
	lowerMethodName := ident.Normalize(methodName)

	// Find the declared method in the record, handling overloads.
	var declaredMethod *types.FunctionType
	var methodExists bool
	if decl.IsClassMethod {
		overloads := recordType.GetClassMethodOverloads(lowerMethodName)
		if len(overloads) > 0 {
			declaredMethod, methodExists = a.findMatchingOverloadForImplementation(decl, overloads, recordName)
		} else {
			declaredMethod, methodExists = recordType.ClassMethods[lowerMethodName]
		}
	} else {
		overloads := recordType.GetMethodOverloads(lowerMethodName)
		if len(overloads) > 0 {
			declaredMethod, methodExists = a.findMatchingOverloadForImplementation(decl, overloads, recordName)
		} else {
			declaredMethod, methodExists = recordType.Methods[lowerMethodName]
		}
	}

	if !methodExists {
		a.addError("%s '%s' not declared in record '%s'", "method", methodName, recordName)
		return
	}

	// Validate the signature of non-overloaded methods.
	isOverloaded := (decl.IsClassMethod && len(recordType.GetClassMethodOverloads(lowerMethodName)) > 1) ||
		(!decl.IsClassMethod && len(recordType.GetMethodOverloads(lowerMethodName)) > 1)
	if !isOverloaded {
		if err := a.validateMethodSignature(decl, declaredMethod, recordName); err != nil {
			a.addError("%s at %s", err.Error(), decl.Token.Pos.String())
			return
		}
	}

	// Analyze the method body with the correct record scope.
	a.analyzeRecordMethodBody(decl, recordType)
}

// analyzeRecordMethodBody analyzes the body of a record method.
func (a *Analyzer) analyzeRecordMethodBody(decl *ast.FunctionDecl, recordType *types.RecordType) {
	previousRecord := a.currentRecord
	a.currentRecord = recordType
	defer func() { a.currentRecord = previousRecord }()

	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(oldSymbols)
	defer func() { a.symbols = oldSymbols }()

	// Bind 'Self', fields, properties, constants, and methods to scope.
	a.symbols.Define("Self", recordType, decl.Token.Pos)
	for fieldName, fieldType := range recordType.Fields {
		a.symbols.Define(recordType.FieldNames[fieldName], fieldType, token.Position{})
	}
	for _, propInfo := range recordType.Properties {
		a.symbols.Define(propInfo.Name, propInfo.Type, token.Position{})
	}
	for _, constInfo := range recordType.Constants {
		a.symbols.Define(constInfo.Name, constInfo.Type, token.Position{})
	}
	for varName, varType := range recordType.ClassVars {
		a.symbols.Define(recordType.ClassVarNames[varName], varType, token.Position{})
	}
	for methodName, methodType := range recordType.Methods {
		a.symbols.DefineFunction(recordType.MethodNames[methodName], methodType, token.Position{})
	}
	for methodName, methodType := range recordType.ClassMethods {
		a.symbols.DefineFunction(recordType.ClassMethodNames[methodName], methodType, token.Position{})
	}

	// Bind parameters to scope.
	for _, param := range decl.Parameters {
		paramTypeName := getTypeExpressionName(param.Type)
		paramType, err := a.resolveType(paramTypeName)
		if err != nil {
			a.addError("unknown parameter type '%s' at %s", paramTypeName, param.Token.Pos.String())
			continue
		}
		a.symbols.Define(param.Name.Value, paramType, param.Name.Token.Pos)
	}

	// Bind 'Result' variable for functions.
	if decl.ReturnType != nil {
		returnType, err := a.resolveType(getTypeExpressionName(decl.ReturnType))
		if err != nil {
			a.addError("unknown return type at %s", decl.Token.Pos.String())
		} else {
			a.symbols.Define("Result", returnType, decl.Name.Token.Pos)
			a.symbols.Define(decl.Name.Value, returnType, decl.Name.Token.Pos)
		}
	}

	previousFunc := a.currentFunction
	a.currentFunction = decl
	defer func() { a.currentFunction = previousFunc }()

	if decl.Body != nil {
		a.analyzeBlock(decl.Body)
	}
}

// findMatchingOverloadForImplementation finds the declared overload matching an implementation's signature.
func (a *Analyzer) findMatchingOverloadForImplementation(implDecl *ast.FunctionDecl, overloads []*types.MethodInfo, className string) (*types.FunctionType, bool) {
	implParamCount := len(implDecl.Parameters)

	// Filter overloads by parameter count.
	matchingCount := make([]*types.MethodInfo, 0)
	for _, overload := range overloads {
		if len(overload.Signature.Parameters) == implParamCount {
			matchingCount = append(matchingCount, overload)
		}
	}

	if len(matchingCount) == 0 {
		return nil, false
	}
	if len(matchingCount) == 1 {
		return matchingCount[0].Signature, true
	}

	// If count is ambiguous, match by parameter types.
	for _, overload := range matchingCount {
		matches := true
		for i, param := range implDecl.Parameters {
			if param.Type == nil {
				continue // Allow omitting types in implementation.
			}
			paramType, err := a.resolveType(getTypeExpressionName(param.Type))
			if err != nil || !paramType.Equals(overload.Signature.Parameters[i]) {
				matches = false
				break
			}
		}
		if matches {
			return overload.Signature, true
		}
	}

	// Return first match by count if types don't resolve ambiguity; validation will catch it.
	return matchingCount[0].Signature, true
}

// analyzeMethodDecl analyzes a method declaration within a class.
func (a *Analyzer) analyzeMethodDecl(method *ast.FunctionDecl, classType *types.ClassType) {
	if method.CallingConvention != "" {
		a.addHint("Calling convention \"%s\" is ignored at %s", method.CallingConvention, method.CallingConventionPos.String())
	}

	// Process parameters and build metadata.
	paramTypes := make([]types.Type, 0, len(method.Parameters))
	paramNames := make([]string, 0, len(method.Parameters))
	defaultValues := make([]interface{}, 0, len(method.Parameters))
	lazyParams := make([]bool, 0, len(method.Parameters))
	varParams := make([]bool, 0, len(method.Parameters))
	constParams := make([]bool, 0, len(method.Parameters))

	for _, param := range method.Parameters {
		if param.Type == nil {
			a.addError("parameter '%s' missing type in method '%s'", param.Name.Value, method.Name.Value)
			return
		}
		paramTypeName := getTypeExpressionName(param.Type)
		if aliases, ok := a.nestedTypeAliases[ident.Normalize(classType.Name)]; ok && a.currentNestedTypes == nil {
			if qualified, ok := aliases[ident.Normalize(paramTypeName)]; ok {
				paramTypeName = qualified
			}
		}
		paramType, err := a.resolveType(paramTypeName)
		if err != nil {
			a.addError("unknown parameter type '%s' in method '%s'", paramTypeName, method.Name.Value)
			return
		}
		paramTypes = append(paramTypes, paramType)
		paramNames = append(paramNames, param.Name.Value)
		defaultValues = append(defaultValues, param.DefaultValue)
		lazyParams = append(lazyParams, param.IsLazy)
		varParams = append(varParams, param.ByRef)
		constParams = append(constParams, param.IsConst)
	}

	// Auto-detect constructors and validate signatures.
	wasExplicitConstructor := method.IsConstructor
	if !method.IsConstructor && ident.Equal(method.Name.Value, "Create") && method.ReturnType != nil {
		if returnTypeName := getTypeExpressionName(method.ReturnType); ident.Equal(returnTypeName, classType.Name) {
			method.IsConstructor = true
		}
	}
	if method.IsConstructor && method.ReturnType != nil {
		if wasExplicitConstructor {
			a.addError("constructor '%s' cannot have an explicit return type at %s", method.Name.Value, method.Token.Pos.String())
			return
		}
		returnTypeName := getTypeExpressionName(method.ReturnType)
		if !ident.Equal(returnTypeName, classType.Name) {
			a.addError("constructor '%s' must return '%s', not '%s' at %s",
				method.Name.Value, classType.Name, returnTypeName, method.Token.Pos.String())
			return
		}
	}

	// Determine return type. Constructors implicitly return the class type.
	var returnType types.Type
	if method.ReturnType != nil {
		var err error
		returnType, err = a.resolveType(getTypeExpressionName(method.ReturnType))
		if err != nil {
			a.addError("unknown return type '%s' in method '%s'", getTypeExpressionName(method.ReturnType), method.Name.Value)
			return
		}
	} else if method.IsConstructor {
		returnType = classType
	} else {
		returnType = types.VOID
	}

	// Create the function type, handling variadic parameters.
	var funcType *types.FunctionType
	if len(paramTypes) > 0 {
		lastParamType := paramTypes[len(paramTypes)-1]
		if arrayType, ok := lastParamType.(*types.ArrayType); ok && arrayType.IsDynamic() {
			variadicType := arrayType.ElementType
			funcType = types.NewVariadicFunctionTypeWithMetadata(
				paramTypes, paramNames, defaultValues, lazyParams, varParams, constParams, variadicType, returnType)
		} else {
			funcType = types.NewFunctionTypeWithMetadata(
				paramTypes, paramNames, defaultValues, lazyParams, varParams, constParams, returnType)
		}
	} else {
		funcType = types.NewFunctionTypeWithMetadata(
			paramTypes, paramNames, defaultValues, lazyParams, varParams, constParams, returnType)
	}

	// Create method info and check for duplicate/ambiguous overloads.
	methodInfo := &types.MethodInfo{
		Signature:            funcType,
		IsVirtual:            method.IsVirtual,
		IsOverride:           method.IsOverride,
		IsAbstract:           method.IsAbstract,
		IsReintroduce:        method.IsReintroduce,
		IsForwarded:          method.Body == nil,
		IsClassMethod:        method.IsClassMethod,
		HasOverloadDirective: method.IsOverload,
		Visibility:           int(method.Visibility),
	}

	existingOverloads := classType.GetMethodOverloads(method.Name.Value)
	if method.IsConstructor {
		existingOverloads = classType.GetConstructorOverloads(method.Name.Value)
	}

	isImplementationOfForward := false
	for _, existing := range existingOverloads {
		if a.methodSignaturesMatch(funcType, existing.Signature) {
			// This is an implementation for a forward declaration.
			if existing.IsForwarded && method.Body != nil {
				existing.IsForwarded = false
				existing.Signature = funcType
				isImplementationOfForward = true
				break
			}
			a.addError("duplicate method signature for '%s' at %s", method.Name.Value, method.Token.Pos.String())
			return
		}
		if a.parametersMatch(funcType, existing.Signature) && !funcType.ReturnType.Equals(existing.Signature.ReturnType) {
			a.addError("ambiguous overload for '%s' at %s", method.Name.Value, method.Token.Pos.String())
			return
		}
	}

	// Add new overload if it's not an implementation of a forward declaration.
	if !isImplementationOfForward {
		if method.IsConstructor {
			classType.AddConstructorOverload(method.Name.Value, methodInfo)
			if method.IsDefault {
				if classType.DefaultConstructor != "" {
					a.addError(
						"class '%s' already has default constructor '%s'; cannot declare another default constructor '%s'",
						classType.Name, classType.DefaultConstructor, method.Name.Value)
					return
				}
				classType.DefaultConstructor = method.Name.Value
			}
		} else {
			classType.AddMethodOverload(method.Name.Value, methodInfo)
		}
		// Store metadata for new declarations.
		methodKey := ident.Normalize(method.Name.Value)
		classType.ClassMethodFlags[methodKey] = method.IsClassMethod
		classType.VirtualMethods[methodKey] = method.IsVirtual
		classType.OverrideMethods[methodKey] = method.IsOverride
		classType.ReintroduceMethods[methodKey] = method.IsReintroduce
		classType.AbstractMethods[methodKey] = method.IsAbstract
	}

	if method.Body == nil {
		classType.ForwardedMethods[ident.Normalize(method.Name.Value)] = true
	}
	methodKey := ident.Normalize(method.Name.Value)
	if _, exists := classType.MethodVisibility[methodKey]; !exists {
		classType.MethodVisibility[methodKey] = int(method.Visibility)
	}

	// Analyze method body in a new scope.
	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(oldSymbols)
	defer func() { a.symbols = oldSymbols }()

	if method.IsClassMethod {
		// Static methods only access class variables.
		for classVarName, classVarType := range classType.ClassVars {
			a.symbols.Define(classVarName, classVarType, token.Position{})
		}
		if classType.Parent != nil {
			a.addParentClassVarsToScope(classType.Parent)
		}
	} else {
		// Instance methods have 'Self' and access to all members.
		a.symbols.Define("Self", classType, method.Token.Pos)
		for fieldName, fieldType := range classType.Fields {
			a.symbols.Define(fieldName, fieldType, token.Position{})
		}
		for classVarName, classVarType := range classType.ClassVars {
			a.symbols.Define(classVarName, classVarType, token.Position{})
		}
		if classType.Parent != nil {
			a.addParentFieldsToScope(classType.Parent)
			a.addParentClassVarsToScope(classType.Parent)
		}
	}

	// Add parameters and 'Result' variable to scope.
	for i, param := range method.Parameters {
		a.symbols.Define(param.Name.Value, paramTypes[i], param.Name.Token.Pos)
	}
	if returnType != types.VOID {
		a.symbols.Define("Result", returnType, method.Name.Token.Pos)
		a.symbols.Define(method.Name.Value, returnType, method.Name.Token.Pos)
	}

	// Set context for body analysis.
	previousFunc := a.currentFunction
	a.currentFunction = method
	defer func() { a.currentFunction = previousFunc }()
	previousInClassMethod := a.inClassMethod
	a.inClassMethod = method.IsClassMethod
	defer func() { a.inClassMethod = previousInClassMethod }()

	a.validateVirtualOverride(method, classType, funcType)

	if method.Body != nil {
		a.analyzeBlock(method.Body)
	}
}

// Suppress unused import error for fmt
var _ = fmt.Sprint
