package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Helper Type Analysis (Task 9.82-9.85)
// ============================================================================

// analyzeHelperDecl analyzes a helper type declaration.
// Helpers extend existing types with additional methods, properties, and class members.
//
// Task 9.82: Analyze helper declarations
// Task 9.83: Register helper in type environment
// Task 9.84: Implement Self binding in helper methods
func (a *Analyzer) analyzeHelperDecl(decl *ast.HelperDecl) {
	if decl == nil {
		return
	}

	helperName := decl.Name.Value

	// Resolve the target type (the type being extended)
	targetTypeName := decl.ForType.Name
	targetType, err := a.resolveType(targetTypeName)
	if err != nil {
		a.addError("unknown target type '%s' for helper '%s' at %s",
			targetTypeName, helperName, decl.Token.Pos.String())
		return
	}

	// Create the helper type
	helperType := types.NewHelperType(helperName, targetType, decl.IsRecordHelper)

	// Process methods
	for _, method := range decl.Methods {
		a.analyzeHelperMethod(method, helperType, helperName)
	}

	// Process properties
	for _, prop := range decl.Properties {
		a.analyzeHelperProperty(prop, helperType, helperName)
	}

	// Process class variables
	for _, classVar := range decl.ClassVars {
		a.analyzeHelperClassVar(classVar, helperType, helperName)
	}

	// Process class constants
	for _, classConst := range decl.ClassConsts {
		a.analyzeHelperClassConst(classConst, helperType, helperName)
	}

	// Register the helper
	// Multiple helpers can extend the same type, so we store them in a list
	targetTypeName = targetType.String()
	if a.helpers[targetTypeName] == nil {
		a.helpers[targetTypeName] = make([]*types.HelperType, 0)
	}
	a.helpers[targetTypeName] = append(a.helpers[targetTypeName], helperType)

	// Also register the helper type itself in the symbol table
	// so it can be referenced by name (e.g., TStringHelper.PI)
	a.symbols.Define(helperName, helperType)
}

// analyzeHelperMethod analyzes a method in a helper.
// Task 9.84: Self binding - Self refers to the target type instance
func (a *Analyzer) analyzeHelperMethod(method *ast.FunctionDecl, helperType *types.HelperType, helperName string) {
	if method == nil {
		return
	}

	methodName := method.Name.Value

	// Check for duplicate methods
	if _, exists := helperType.Methods[methodName]; exists {
		a.addError("duplicate method '%s' in helper '%s' at %s",
			methodName, helperName, method.Token.Pos.String())
		return
	}

	// Create function type for the method
	var paramTypes []types.Type
	for _, param := range method.Parameters {
		paramType, err := a.resolveType(param.Type.Name)
		if err != nil {
			a.addError("unknown type '%s' for parameter '%s' in helper method '%s.%s' at %s",
				param.Type.Name, param.Name.Value, helperName, methodName, param.Token.Pos.String())
			continue
		}
		paramTypes = append(paramTypes, paramType)
	}

	var returnType types.Type
	if method.ReturnType != nil {
		rt, err := a.resolveType(method.ReturnType.Name)
		if err != nil {
			a.addError("unknown return type '%s' for helper method '%s.%s' at %s",
				method.ReturnType.Name, helperName, methodName, method.Token.Pos.String())
		} else {
			returnType = rt
		}
	}

	var funcType *types.FunctionType
	if returnType != nil {
		funcType = types.NewFunctionType(paramTypes, returnType)
	} else {
		funcType = types.NewProcedureType(paramTypes)
	}

	// Add method to helper
	helperType.Methods[methodName] = funcType

	// Note: In helper methods, 'Self' implicitly refers to an instance of the target type.
	// This is validated when analyzing the method body (not implemented in this stage).
}

// analyzeHelperProperty analyzes a property in a helper.
func (a *Analyzer) analyzeHelperProperty(prop *ast.PropertyDecl, helperType *types.HelperType, helperName string) {
	if prop == nil {
		return
	}

	propName := prop.Name.Value

	// Check for duplicate properties
	if _, exists := helperType.Properties[propName]; exists {
		a.addError("duplicate property '%s' in helper '%s' at %s",
			propName, helperName, prop.Token.Pos.String())
		return
	}

	// Resolve property type
	propType, err := a.resolveType(prop.Type.Name)
	if err != nil {
		a.addError("unknown type '%s' for property '%s' in helper '%s' at %s",
			prop.Type.Name, propName, helperName, prop.Token.Pos.String())
		return
	}

	// Create property info
	propInfo := &types.PropertyInfo{
		Name: propName,
		Type: propType,
		// ReadSpec and WriteSpec analysis would go here
		// For now, we just track the basic property info
	}

	helperType.Properties[propName] = propInfo
}

// analyzeHelperClassVar analyzes a class variable in a helper.
func (a *Analyzer) analyzeHelperClassVar(classVar *ast.FieldDecl, helperType *types.HelperType, helperName string) {
	if classVar == nil {
		return
	}

	varName := classVar.Name.Value

	// Check for duplicate class vars
	if _, exists := helperType.ClassVars[varName]; exists {
		a.addError("duplicate class variable '%s' in helper '%s' at %s",
			varName, helperName, classVar.Token.Pos.String())
		return
	}

	// Resolve variable type
	varType, err := a.resolveType(classVar.Type.Name)
	if err != nil {
		a.addError("unknown type '%s' for class variable '%s' in helper '%s' at %s",
			classVar.Type.Name, varName, helperName, classVar.Token.Pos.String())
		return
	}

	helperType.ClassVars[varName] = varType
}

// analyzeHelperClassConst analyzes a class constant in a helper.
func (a *Analyzer) analyzeHelperClassConst(classConst *ast.ConstDecl, helperType *types.HelperType, helperName string) {
	if classConst == nil {
		return
	}

	constName := classConst.Name.Value

	// Check for duplicate class consts
	if _, exists := helperType.ClassConsts[constName]; exists {
		a.addError("duplicate class constant '%s' in helper '%s' at %s",
			constName, helperName, classConst.Token.Pos.String())
		return
	}

	// Analyze the constant value expression
	constType := a.analyzeExpression(classConst.Value)
	if constType == nil {
		a.addError("invalid constant value for '%s' in helper '%s' at %s",
			constName, helperName, classConst.Token.Pos.String())
		return
	}

	// Type checking: if a type is specified, validate it matches
	if classConst.Type != nil {
		expectedType, err := a.resolveType(classConst.Type.Name)
		if err != nil {
			a.addError("unknown type '%s' for constant '%s' in helper '%s' at %s",
				classConst.Type.Name, constName, helperName, classConst.Token.Pos.String())
			return
		}

		if !a.canAssign(constType, expectedType) {
			a.addError("constant '%s' type mismatch: cannot assign %s to %s in helper '%s' at %s",
				constName, constType.String(), expectedType.String(), helperName, classConst.Token.Pos.String())
			return
		}
	}

	// Store the constant (value would be evaluated by interpreter)
	// For now, we just track that it exists with its type
	helperType.ClassConsts[constName] = constType
}

// getHelpersForType returns all helpers that extend the given type.
// Task 9.83: Helper method resolution
func (a *Analyzer) getHelpersForType(typ types.Type) []*types.HelperType {
	if typ == nil {
		return nil
	}

	// Look up helpers by the type's string representation
	typeName := typ.String()
	return a.helpers[typeName]
}

// hasHelperMethod checks if any helper for the given type defines the specified method.
// Returns the helper type and method if found.
// Task 9.83: Helper method resolution
func (a *Analyzer) hasHelperMethod(typ types.Type, methodName string) (*types.HelperType, *types.FunctionType) {
	helpers := a.getHelpersForType(typ)
	if helpers == nil {
		return nil, nil
	}

	// Check each helper in order (first match wins)
	for _, helper := range helpers {
		if method, ok := helper.Methods[methodName]; ok {
			return helper, method
		}
	}

	return nil, nil
}

// hasHelperProperty checks if any helper for the given type defines the specified property.
// Returns the helper type and property if found.
func (a *Analyzer) hasHelperProperty(typ types.Type, propName string) (*types.HelperType, *types.PropertyInfo) {
	helpers := a.getHelpersForType(typ)
	if helpers == nil {
		return nil, nil
	}

	// Check each helper in order (first match wins)
	for _, helper := range helpers {
		if prop, ok := helper.Properties[propName]; ok {
			return helper, prop
		}
	}

	return nil, nil
}
