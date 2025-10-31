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
	typeName := getTypeExpressionName(classVar.Type)
	varType, err := a.resolveType(typeName)
	if err != nil {
		a.addError("unknown type '%s' for class variable '%s' in helper '%s' at %s",
			typeName, varName, helperName, classVar.Token.Pos.String())
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
	helpers := a.helpers[typeName]

	// Task 9.171: For array types, also include generic ARRAY helpers
	if _, isArray := typ.(*types.ArrayType); isArray {
		arrayHelpers := a.helpers["ARRAY"]
		if arrayHelpers != nil {
			// Combine type-specific helpers with generic array helpers
			helpers = append(helpers, arrayHelpers...)
		}
	}

	return helpers
}

// hasHelperMethod checks if any helper for the given type defines the specified method.
// Returns the helper type and method if found.
// Task 9.83: Helper method resolution
func (a *Analyzer) hasHelperMethod(typ types.Type, methodName string) (*types.HelperType, *types.FunctionType) {
	helpers := a.getHelpersForType(typ)
	if helpers == nil {
		return nil, nil
	}

	// Check each helper in reverse order so user-defined helpers (added later)
	// take precedence over built-in helpers registered during initialization.
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
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
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
		if prop, ok := helper.Properties[propName]; ok {
			return helper, prop
		}
	}

	return nil, nil
}

// ============================================================================
// Built-in Array Helpers (Task 9.171)
// ============================================================================

// initArrayHelpers registers built-in helper properties for arrays
// Task 9.171.6: Semantic analyzer support for array helpers
func (a *Analyzer) initArrayHelpers() {
	// Create a helper for the generic ARRAY type
	// Since we need to support all array types, we'll register this for "ARRAY"
	// and modify getHelpersForType to check for array types
	arrayHelper := types.NewHelperType("TArrayHelper", nil, false)

	// Task 9.171.4: Register .Length property
	arrayHelper.Properties["Length"] = &types.PropertyInfo{
		Name:      "Length",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_length",
		WriteKind: types.PropAccessNone,
	}

	// Task 9.171.2: Register .High property
	arrayHelper.Properties["High"] = &types.PropertyInfo{
		Name:      "High",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_high",
		WriteKind: types.PropAccessNone,
	}

	// Task 9.171.3: Register .Low property
	arrayHelper.Properties["Low"] = &types.PropertyInfo{
		Name:      "Low",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_low",
		WriteKind: types.PropAccessNone,
	}

	// Register helper for ARRAY type (generic catch-all)
	a.helpers["ARRAY"] = append(a.helpers["ARRAY"], arrayHelper)
}

// initIntrinsicHelpers registers built-in helpers for primitive types (Integer, Float, Boolean).
// Task 9.205: Provide default helpers used by DWScript for core types.
func (a *Analyzer) initIntrinsicHelpers() {
	// Helper registration helper to reduce duplication.
	register := func(typeName string, helper *types.HelperType) {
		if a.helpers[typeName] == nil {
			a.helpers[typeName] = make([]*types.HelperType, 0)
		}
		a.helpers[typeName] = append(a.helpers[typeName], helper)
	}

	// Integer helper: provides ToString method/property
	intHelper := types.NewHelperType("__TIntegerIntrinsicHelper", types.INTEGER, false)
	intHelper.Properties["ToString"] = &types.PropertyInfo{
		Name:      "ToString",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__integer_tostring",
		WriteKind: types.PropAccessNone,
	}
	intHelper.Methods["ToString"] = types.NewFunctionType([]types.Type{}, types.STRING)
	intHelper.BuiltinMethods["ToString"] = "__integer_tostring"
	register(types.INTEGER.String(), intHelper)

	// Float helper: default ToString property and precision-aware method
	floatHelper := types.NewHelperType("__TFloatIntrinsicHelper", types.FLOAT, false)
	floatHelper.Properties["ToString"] = &types.PropertyInfo{
		Name:      "ToString",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__float_tostring_default",
		WriteKind: types.PropAccessNone,
	}
	floatHelper.Methods["ToString"] = types.NewFunctionType([]types.Type{types.INTEGER}, types.STRING)
	floatHelper.BuiltinMethods["ToString"] = "__float_tostring_prec"
	register(types.FLOAT.String(), floatHelper)

	// Boolean helper: ToString method/property returning 'True'/'False'
	boolHelper := types.NewHelperType("__TBooleanIntrinsicHelper", types.BOOLEAN, false)
	boolHelper.Properties["ToString"] = &types.PropertyInfo{
		Name:      "ToString",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__boolean_tostring",
		WriteKind: types.PropAccessNone,
	}
	boolHelper.Methods["ToString"] = types.NewFunctionType([]types.Type{}, types.STRING)
	boolHelper.BuiltinMethods["ToString"] = "__boolean_tostring"
	register(types.BOOLEAN.String(), boolHelper)

	// String dynamic array helper: Join method
	stringArrayHelper := types.NewHelperType("__TStringDynArrayIntrinsicHelper", nil, true)
	stringArrayHelper.TargetType = types.NewDynamicArrayType(types.STRING)
	stringArrayHelper.Methods["Join"] = types.NewFunctionType([]types.Type{types.STRING}, types.STRING)
	stringArrayHelper.BuiltinMethods["Join"] = "__string_array_join"
	register(stringArrayHelper.TargetType.String(), stringArrayHelper)
}
