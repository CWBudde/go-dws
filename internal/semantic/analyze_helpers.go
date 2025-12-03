package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Helper Type Analysis
// ============================================================================

// findPropertyCaseInsensitive searches for a property by name using case-insensitive comparison.
// Task 9.217: Support case-insensitive helper property lookup
func findPropertyCaseInsensitive(props map[string]*types.PropertyInfo, name string) *types.PropertyInfo {
	for key, prop := range props {
		if ident.Equal(key, name) {
			return prop
		}
	}
	return nil
}

// findMethodCaseInsensitive searches for a method by name using case-insensitive comparison.
// Task 9.217: Support case-insensitive helper method lookup
func findMethodCaseInsensitive(methods map[string]*types.FunctionType, name string) *types.FunctionType {
	for key, method := range methods {
		if ident.Equal(key, name) {
			return method
		}
	}
	return nil
}

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
	targetTypeName := getTypeExpressionName(decl.ForType)
	targetType, err := a.resolveType(targetTypeName)
	if err != nil {
		a.addError("unknown target type '%s' for helper '%s' at %s",
			targetTypeName, helperName, decl.Token.Pos.String())
		return
	}

	// Create the helper type
	helperType := types.NewHelperType(helperName, targetType, decl.IsRecordHelper)

	// Resolve parent helper if specified
	if decl.ParentHelper != nil {
		parentHelperName := decl.ParentHelper.Value

		// Look up the parent helper in the symbol table
		parentSymbol, exists := a.symbols.Resolve(parentHelperName)
		if !exists {
			a.addError("unknown parent helper '%s' for helper '%s' at %s",
				parentHelperName, helperName, decl.ParentHelper.Token.Pos.String())
		} else {
			// Verify that the parent is actually a helper type
			parentHelper, ok := parentSymbol.Type.(*types.HelperType)
			if !ok {
				a.addError("'%s' is not a helper type (used as parent for helper '%s') at %s",
					parentHelperName, helperName, decl.ParentHelper.Token.Pos.String())
			} else {
				// Verify that the parent helper extends the same target type
				if !parentHelper.TargetType.Equals(targetType) {
					a.addError("parent helper '%s' extends type '%s', but child helper '%s' extends type '%s' at %s",
						parentHelperName, parentHelper.TargetType.String(),
						helperName, targetType.String(), decl.ParentHelper.Token.Pos.String())
				} else {
					// All validations passed - set up the inheritance chain
					helperType.ParentHelper = parentHelper
				}
			}
		}
	}

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
	targetTypeName = ident.Normalize(targetType.String())
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
		paramType, err := a.resolveType(getTypeExpressionName(param.Type))
		if err != nil {
			a.addError("unknown type '%s' for parameter '%s' in helper method '%s.%s' at %s",
				getTypeExpressionName(param.Type), param.Name.Value, helperName, methodName, param.Token.Pos.String())
			continue
		}
		paramTypes = append(paramTypes, paramType)
	}

	var returnType types.Type
	if method.ReturnType != nil {
		rt, err := a.resolveType(getTypeExpressionName(method.ReturnType))
		if err != nil {
			a.addError("unknown return type '%s' for helper method '%s.%s' at %s",
				getTypeExpressionName(method.ReturnType), helperName, methodName, method.Token.Pos.String())
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
	methodNameLower := ident.Normalize(methodName)
	helperType.Methods[methodNameLower] = funcType

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
	propType, err := a.resolveType(getTypeExpressionName(prop.Type))
	if err != nil {
		a.addError("unknown type '%s' for property '%s' in helper '%s' at %s",
			getTypeExpressionName(prop.Type), propName, helperName, prop.Token.Pos.String())
		return
	}

	// Create property info
	propInfo := &types.PropertyInfo{
		Name: propName,
		Type: propType,
		// ReadSpec and WriteSpec analysis would go here
		// For now, we just track the basic property info
	}

	propNameLower := ident.Normalize(propName)
	helperType.Properties[propNameLower] = propInfo
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

	var varType types.Type

	if classVar.Type != nil {
		// Explicit type annotation
		typeName := getTypeExpressionName(classVar.Type)
		resolvedType, err := a.resolveType(typeName)
		if err != nil {
			a.addError("unknown type '%s' for class variable '%s' in helper '%s' at %s",
				typeName, varName, helperName, classVar.Token.Pos.String())
			return
		}
		varType = resolvedType
	} else if classVar.InitValue != nil {
		// Infer type from initializer
		inferredType := a.analyzeExpression(classVar.InitValue)
		if inferredType == nil {
			a.addError("cannot infer type for class variable '%s' in helper '%s' at %s",
				varName, helperName, classVar.Token.Pos.String())
			return
		}
		varType = inferredType
	} else {
		a.addError("class variable '%s' missing type annotation in helper '%s'",
			varName, helperName)
		return
	}

	// Validate initializer compatibility when both are present
	if classVar.InitValue != nil && classVar.Type != nil {
		initType := a.analyzeExpression(classVar.InitValue)
		if initType != nil && varType != nil && !types.IsAssignableFrom(varType, initType) {
			a.addError("cannot initialize class variable '%s' of type '%s' with value of type '%s' in helper '%s' at %s",
				varName, varType.String(), initType.String(), helperName, classVar.Token.Pos.String())
			return
		}
	}

	varNameLower := ident.Normalize(varName)
	helperType.ClassVars[varNameLower] = varType
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
		expectedType, err := a.resolveType(getTypeExpressionName(classConst.Type))
		if err != nil {
			a.addError("unknown type '%s' for constant '%s' in helper '%s' at %s",
				getTypeExpressionName(classConst.Type), constName, helperName, classConst.Token.Pos.String())
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
	constNameLower := ident.Normalize(constName)
	helperType.ClassConsts[constNameLower] = constType
}

// getHelpersForType returns all helpers that extend the given type.
// Task 9.83: Helper method resolution
func (a *Analyzer) getHelpersForType(typ types.Type) []*types.HelperType {
	if typ == nil {
		return nil
	}

	// If the type itself is a helper type (e.g., TDummy.Hello), use its target type
	if helperType, ok := typ.(*types.HelperType); ok {
		if helperType.TargetType == nil {
			return nil
		}
		typ = helperType.TargetType
	}

	// Look up helpers by the type's string representation
	typeName := ident.Normalize(typ.String())
	helpers := a.helpers[typeName]

	// Task 9.171: For array types, also include generic array helpers
	if _, isArray := typ.(*types.ArrayType); isArray {
		arrayHelpers := a.helpers["array"]
		if arrayHelpers != nil {
			// Combine type-specific helpers with generic array helpers
			helpers = append(helpers, arrayHelpers...)
		}
	}

	// Task 9.31: For enum types, also include generic enum helpers
	if _, isEnum := typ.(*types.EnumType); isEnum {
		enumHelpers := a.helpers["enum"]
		if enumHelpers != nil {
			// Combine type-specific helpers with generic enum helpers
			helpers = append(helpers, enumHelpers...)
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
	// Task 9.217: Use case-insensitive lookup for DWScript compatibility
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
		if method := findMethodCaseInsensitive(helper.Methods, methodName); method != nil {
			// For array types, specialize the method signature if needed
			// (e.g., Pop() should return the array's element type, not VARIANT)
			if arrayType, isArray := typ.(*types.ArrayType); isArray {
				// Check if this is the Pop method that needs specialization
				if ident.Equal(methodName, "pop") && method.ReturnType == types.VARIANT {
					// Create a specialized version with the actual element type
					specialized := types.NewFunctionType(method.Parameters, arrayType.ElementType)
					specialized.ParamNames = method.ParamNames
					specialized.DefaultValues = method.DefaultValues
					specialized.VarParams = method.VarParams
					specialized.ConstParams = method.ConstParams
					specialized.LazyParams = method.LazyParams
					specialized.IsVariadic = method.IsVariadic
					specialized.VariadicType = method.VariadicType
					return helper, specialized
				}
			}
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
	// Task 9.217: Use case-insensitive lookup for DWScript compatibility
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
		if prop := findPropertyCaseInsensitive(helper.Properties, propName); prop != nil {
			return helper, prop
		}
	}

	return nil, nil
}

// hasHelperClassVar checks if any helper for the given type defines the specified class variable.
// Returns the helper type and variable type if found.
func (a *Analyzer) hasHelperClassVar(typ types.Type, varName string) (*types.HelperType, types.Type) {
	helpers := a.getHelpersForType(typ)
	if helpers == nil {
		return nil, nil
	}

	varNameLower := ident.Normalize(varName)
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
		if varType, ok := helper.ClassVars[varNameLower]; ok {
			return helper, varType
		}
	}

	return nil, nil
}

// hasHelperClassConst checks if any helper for the given type defines the specified class constant.
// Returns the helper type and constant value if found.
// Task 9.54: Support scoped enum access via helper class constants
func (a *Analyzer) hasHelperClassConst(typ types.Type, constName string) (*types.HelperType, interface{}) {
	helpers := a.getHelpersForType(typ)
	if helpers == nil {
		return nil, nil
	}

	// Check each helper in reverse order (most recent first)
	// Task 9.217: Use case-insensitive lookup for DWScript compatibility
	constNameLower := ident.Normalize(constName)
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
		if constVal, ok := helper.ClassConsts[constNameLower]; ok {
			return helper, constVal
		}
	}

	return nil, nil
}

// ============================================================================
// Built-in Array Helpers
// ============================================================================

// initArrayHelpers registers built-in helper properties for arrays
// Task 9.171.6: Semantic analyzer support for array helpers
func (a *Analyzer) initArrayHelpers() {
	// Create a helper for the generic ARRAY type
	// Since we need to support all array types, we'll register this for "ARRAY"
	// and modify getHelpersForType to check for array types
	arrayHelper := types.NewHelperType("TArrayHelper", nil, false)

	// Task 9.171.4: Register .Length property (lowercase key for case-insensitive lookup)
	arrayHelper.Properties["length"] = &types.PropertyInfo{
		Name:      "Length",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_length",
		WriteKind: types.PropAccessNone,
	}

	// Task 9.171.2: Register .High property (lowercase key for case-insensitive lookup)
	arrayHelper.Properties["high"] = &types.PropertyInfo{
		Name:      "High",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_high",
		WriteKind: types.PropAccessNone,
	}

	// Task 9.171.3: Register .Low property (lowercase key for case-insensitive lookup)
	arrayHelper.Properties["low"] = &types.PropertyInfo{
		Name:      "Low",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_low",
		WriteKind: types.PropAccessNone,
	}

	// Task 9.34: Register .Count property (alias for .Length) (lowercase key for case-insensitive lookup)
	arrayHelper.Properties["count"] = &types.PropertyInfo{
		Name:      "Count",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_count",
		WriteKind: types.PropAccessNone,
	}

	// Register .Add() method for dynamic arrays (lowercase key for case-insensitive lookup)
	arrayHelper.Methods["add"] = types.NewProcedureType([]types.Type{nil}) // Takes one parameter (element to add)
	arrayHelper.BuiltinMethods["add"] = "__array_add"

	// Task 9.34: Register .Delete() method for dynamic arrays (lowercase key for case-insensitive lookup)
	// Delete can take 1 or 2 parameters: Delete(index) or Delete(index, count)
	// Second parameter (count) is optional with default value 1
	arrayHelper.Methods["delete"] = types.NewFunctionTypeWithMetadata(
		[]types.Type{types.INTEGER, types.INTEGER},
		[]string{"index", "count"},
		[]interface{}{nil, int64(1)}, // Default count = 1
		[]bool{false, false},
		[]bool{false, false},
		[]bool{false, false},
		nil, // Procedure (no return value)
	)
	arrayHelper.BuiltinMethods["delete"] = "__array_delete"

	// Task 9.34: Register .IndexOf() method for dynamic arrays (lowercase key for case-insensitive lookup)
	// IndexOf(value) or IndexOf(value, startIndex)
	// Second parameter (startIndex) is optional with default value 0
	arrayHelper.Methods["indexof"] = types.NewFunctionTypeWithMetadata(
		[]types.Type{nil, types.INTEGER},
		[]string{"value", "startIndex"},
		[]interface{}{nil, int64(0)}, // Default startIndex = 0
		[]bool{false, false},
		[]bool{false, false},
		[]bool{false, false},
		types.INTEGER,
	)
	arrayHelper.BuiltinMethods["indexof"] = "__array_indexof"

	// Task 9.216: Register .SetLength() method for dynamic arrays (lowercase key for case-insensitive lookup)
	arrayHelper.Methods["setlength"] = types.NewProcedureType([]types.Type{types.INTEGER})
	arrayHelper.BuiltinMethods["setlength"] = "__array_setlength"

	// Task 9.8: Register .Swap() method for arrays (lowercase key for case-insensitive lookup)
	// Swap(i, j) - swaps elements at indices i and j
	arrayHelper.Methods["swap"] = types.NewProcedureType([]types.Type{types.INTEGER, types.INTEGER})
	arrayHelper.BuiltinMethods["swap"] = "__array_swap"

	// Task 9.8: Register .Push() method for dynamic arrays (lowercase key for case-insensitive lookup)
	// Push(value) - appends element (alias for Add)
	arrayHelper.Methods["push"] = types.NewProcedureType([]types.Type{nil}) // Takes one parameter (element to push)
	arrayHelper.BuiltinMethods["push"] = "__array_push"

	// Task 9.8: Register .Pop() method for dynamic arrays (lowercase key for case-insensitive lookup)
	// Pop() - removes and returns last element
	// Use VARIANT as placeholder - will be specialized to array's element type by hasHelperMethod
	arrayHelper.Methods["pop"] = types.NewFunctionType([]types.Type{}, types.VARIANT)
	arrayHelper.BuiltinMethods["pop"] = "__array_pop"

	// Register helper for array type (generic catch-all)
	a.helpers["array"] = append(a.helpers["array"], arrayHelper)

	// Generic array helper methods
	arrayHelper.Methods["map"] = types.NewFunctionType([]types.Type{types.NewFunctionPointerType([]types.Type{types.VARIANT}, types.VARIANT)}, types.NewDynamicArrayType(types.VARIANT))
	arrayHelper.BuiltinMethods["map"] = "__array_map"
	arrayHelper.Methods["join"] = types.NewFunctionType([]types.Type{types.STRING}, types.STRING)
	arrayHelper.BuiltinMethods["join"] = "__array_join"
}

// initIntrinsicHelpers registers built-in helpers for primitive types (Integer, Float, Boolean).
// Task 9.205: Provide default helpers used by DWScript for core types.
func (a *Analyzer) initIntrinsicHelpers() {
	// Helper registration helper to reduce duplication.
	register := func(typeName string, helper *types.HelperType) {
		key := ident.Normalize(typeName)
		if a.helpers[key] == nil {
			a.helpers[key] = make([]*types.HelperType, 0)
		}
		a.helpers[key] = append(a.helpers[key], helper)
	}

	// Integer helper: provides ToString method/property (lowercase keys for case-insensitive lookup)
	intHelper := types.NewHelperType("__TIntegerIntrinsicHelper", types.INTEGER, false)
	intHelper.Properties["tostring"] = &types.PropertyInfo{
		Name:      "ToString",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__integer_tostring",
		WriteKind: types.PropAccessNone,
	}
	// ToString([base]) - optional base (2..36), default 10
	intHelper.Methods["tostring"] = types.NewFunctionTypeWithMetadata(
		[]types.Type{types.INTEGER},
		[]string{"base"},
		[]interface{}{int64(10)},
		[]bool{false},
		[]bool{false},
		[]bool{false},
		types.STRING,
	)
	intHelper.BuiltinMethods["tostring"] = "__integer_tostring"
	// ToHexString method: converts integer to hex string with specified number of digits
	intHelper.Methods["tohexstring"] = types.NewFunctionType([]types.Type{types.INTEGER}, types.STRING)
	intHelper.BuiltinMethods["tohexstring"] = "__integer_tohexstring"
	register(types.INTEGER.String(), intHelper)

	// Float helper: default ToString property and precision-aware method (lowercase keys for case-insensitive lookup)
	floatHelper := types.NewHelperType("__TFloatIntrinsicHelper", types.FLOAT, false)
	floatHelper.Properties["tostring"] = &types.PropertyInfo{
		Name:      "ToString",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__float_tostring_default",
		WriteKind: types.PropAccessNone,
	}
	floatHelper.Methods["tostring"] = types.NewFunctionType([]types.Type{types.INTEGER}, types.STRING)
	floatHelper.BuiltinMethods["tostring"] = "__float_tostring_prec"
	register(types.FLOAT.String(), floatHelper)

	// Boolean helper: ToString method/property returning 'True'/'False' (lowercase keys for case-insensitive lookup)
	boolHelper := types.NewHelperType("__TBooleanIntrinsicHelper", types.BOOLEAN, false)
	boolHelper.Properties["tostring"] = &types.PropertyInfo{
		Name:      "ToString",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__boolean_tostring",
		WriteKind: types.PropAccessNone,
	}
	boolHelper.Methods["tostring"] = types.NewFunctionType([]types.Type{}, types.STRING)
	boolHelper.BuiltinMethods["tostring"] = "__boolean_tostring"
	register(types.BOOLEAN.String(), boolHelper)

	// String helper: provides Length property, ToUpper, and ToLower methods
	stringHelper := types.NewHelperType("__TStringIntrinsicHelper", types.STRING, false)
	stringHelper.Properties["length"] = &types.PropertyInfo{
		Name:      "Length",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__string_length",
		WriteKind: types.PropAccessNone,
	}
	stringHelper.Methods["toupper"] = types.NewFunctionType([]types.Type{}, types.STRING)
	stringHelper.BuiltinMethods["toupper"] = "__string_toupper"
	stringHelper.Methods["tolower"] = types.NewFunctionType([]types.Type{}, types.STRING)
	stringHelper.BuiltinMethods["tolower"] = "__string_tolower"
	// PadLeft, PadRight: (count: Integer, [char: String]) - char is optional with default ' '
	stringHelper.Methods["padleft"] = types.NewFunctionTypeWithMetadata(
		[]types.Type{types.INTEGER, types.STRING},
		[]string{"count", "char"},
		[]interface{}{nil, " "},
		[]bool{false, false},
		[]bool{false, false},
		[]bool{false, false},
		types.STRING,
	)
	stringHelper.BuiltinMethods["padleft"] = "PadLeft"
	stringHelper.Methods["padright"] = types.NewFunctionTypeWithMetadata(
		[]types.Type{types.INTEGER, types.STRING},
		[]string{"count", "char"},
		[]interface{}{nil, " "},
		[]bool{false, false},
		[]bool{false, false},
		[]bool{false, false},
		types.STRING,
	)
	stringHelper.BuiltinMethods["padright"] = "PadRight"
	// DeleteLeft, DeleteRight: (count: Integer)
	stringHelper.Methods["deleteleft"] = types.NewFunctionType([]types.Type{types.INTEGER}, types.STRING)
	stringHelper.BuiltinMethods["deleteleft"] = "StrDeleteLeft"
	stringHelper.Methods["deleteright"] = types.NewFunctionType([]types.Type{types.INTEGER}, types.STRING)
	stringHelper.BuiltinMethods["deleteright"] = "StrDeleteRight"
	// Normalize: ([form: String]) - form is optional with default "NFC"
	stringHelper.Methods["normalize"] = types.NewFunctionTypeWithMetadata(
		[]types.Type{types.STRING},
		[]string{"form"},
		[]interface{}{"NFC"},
		[]bool{false},
		[]bool{false},
		[]bool{false},
		types.STRING,
	)
	stringHelper.BuiltinMethods["normalize"] = "NormalizeString"
	// StripAccents: () - also register as property for no-argument access
	stringHelper.Methods["stripaccents"] = types.NewFunctionType([]types.Type{}, types.STRING)
	stringHelper.BuiltinMethods["stripaccents"] = "StripAccents"
	stringHelper.Properties["stripaccents"] = &types.PropertyInfo{
		Name:      "StripAccents",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "StripAccents",
		WriteKind: types.PropAccessNone,
	}

	// Task 9.23: Conversion helper methods
	// .ToInteger -> StrToInt(self)
	stringHelper.Methods["tointeger"] = types.NewFunctionType([]types.Type{}, types.INTEGER)
	stringHelper.BuiltinMethods["tointeger"] = "__string_tointeger"
	// .ToFloat -> StrToFloat(self)
	stringHelper.Methods["tofloat"] = types.NewFunctionType([]types.Type{}, types.FLOAT)
	stringHelper.BuiltinMethods["tofloat"] = "__string_tofloat"
	// .ToString -> identity (returns self)
	stringHelper.Methods["tostring"] = types.NewFunctionType([]types.Type{}, types.STRING)
	stringHelper.BuiltinMethods["tostring"] = "__string_tostring"

	// Task 9.23: Search/check helper methods
	// .StartsWith(str) -> StrBeginsWith(self, str)
	stringHelper.Methods["startswith"] = types.NewFunctionType([]types.Type{types.STRING}, types.BOOLEAN)
	stringHelper.BuiltinMethods["startswith"] = "__string_startswith"
	// .EndsWith(str) -> StrEndsWith(self, str)
	stringHelper.Methods["endswith"] = types.NewFunctionType([]types.Type{types.STRING}, types.BOOLEAN)
	stringHelper.BuiltinMethods["endswith"] = "__string_endswith"
	// .Contains(str) -> StrContains(self, str)
	stringHelper.Methods["contains"] = types.NewFunctionType([]types.Type{types.STRING}, types.BOOLEAN)
	stringHelper.BuiltinMethods["contains"] = "__string_contains"
	// .IndexOf(str) -> Pos(str, self) - note parameter order is reversed!
	stringHelper.Methods["indexof"] = types.NewFunctionType([]types.Type{types.STRING}, types.INTEGER)
	stringHelper.BuiltinMethods["indexof"] = "__string_indexof"
	// .Matches(mask) -> StrMatches(self, mask)
	stringHelper.Methods["matches"] = types.NewFunctionType([]types.Type{types.STRING}, types.BOOLEAN)
	stringHelper.BuiltinMethods["matches"] = "__string_matches"
	// .IsASCII property
	stringHelper.Properties["isascii"] = &types.PropertyInfo{
		Name:      "IsASCII",
		Type:      types.BOOLEAN,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__string_isascii",
		WriteKind: types.PropAccessNone,
	}
	// Allow property-style access for Trim helpers (e.g., s.Trim)
	stringHelper.Properties["trim"] = &types.PropertyInfo{
		Name:      "Trim",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__string_trim",
		WriteKind: types.PropAccessNone,
	}
	stringHelper.Properties["trimleft"] = &types.PropertyInfo{
		Name:      "TrimLeft",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__string_trimleft",
		WriteKind: types.PropAccessNone,
	}
	stringHelper.Properties["trimright"] = &types.PropertyInfo{
		Name:      "TrimRight",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__string_trimright",
		WriteKind: types.PropAccessNone,
	}

	// Task 9.23: Extraction helper methods
	// .Copy(start, len) -> Copy(self, start, len) - 2-parameter variant
	stringHelper.Methods["copy"] = types.NewFunctionTypeWithMetadata(
		[]types.Type{types.INTEGER, types.INTEGER},
		[]string{"start", "length"},
		[]interface{}{nil, int64(2147483647)}, // Default length = MaxInt (copy to end)
		[]bool{false, false},
		[]bool{false, false},
		[]bool{false, false},
		types.STRING,
	)
	stringHelper.BuiltinMethods["copy"] = "__string_copy"
	// .Before(str) -> StrBefore(self, str)
	stringHelper.Methods["before"] = types.NewFunctionType([]types.Type{types.STRING}, types.STRING)
	stringHelper.BuiltinMethods["before"] = "__string_before"
	// .After(str) -> StrAfter(self, str)
	stringHelper.Methods["after"] = types.NewFunctionType([]types.Type{types.STRING}, types.STRING)
	stringHelper.BuiltinMethods["after"] = "__string_after"

	// Task 9.23: Modification helper methods
	// .Trim([left, right]) -> Trim variations
	stringHelper.Methods["trim"] = types.NewFunctionTypeWithMetadata(
		[]types.Type{types.INTEGER, types.INTEGER},
		[]string{"left", "right"},
		[]interface{}{int64(0), int64(0)},
		[]bool{false, false},
		[]bool{false, false},
		[]bool{false, false},
		types.STRING,
	)
	stringHelper.BuiltinMethods["trim"] = "__string_trim"
	stringHelper.Methods["trimleft"] = types.NewFunctionType([]types.Type{types.INTEGER}, types.STRING)
	stringHelper.BuiltinMethods["trimleft"] = "__string_trimleft"
	stringHelper.Methods["trimright"] = types.NewFunctionType([]types.Type{types.INTEGER}, types.STRING)
	stringHelper.BuiltinMethods["trimright"] = "__string_trimright"

	// Task 9.23: Split/join helper methods
	// .Split(delimiter) -> StrSplit(self, delimiter)
	stringHelper.Methods["split"] = types.NewFunctionType([]types.Type{types.STRING}, types.NewDynamicArrayType(types.STRING))
	stringHelper.BuiltinMethods["split"] = "__string_split"
	// Encoding helpers
	stringHelper.Methods["tojson"] = types.NewFunctionType([]types.Type{}, types.STRING)
	stringHelper.BuiltinMethods["tojson"] = "__string_tojson"
	stringHelper.Methods["tohtml"] = types.NewFunctionType([]types.Type{}, types.STRING)
	stringHelper.BuiltinMethods["tohtml"] = "__string_tohtml"
	stringHelper.Methods["tohtmlattribute"] = types.NewFunctionType([]types.Type{}, types.STRING)
	stringHelper.BuiltinMethods["tohtmlattribute"] = "__string_tohtmlattribute"
	stringHelper.Methods["tocsstext"] = types.NewFunctionType([]types.Type{}, types.STRING)
	stringHelper.BuiltinMethods["tocsstext"] = "__string_tocsstext"
	stringHelper.Methods["toxml"] = types.NewFunctionTypeWithMetadata(
		[]types.Type{types.INTEGER},
		[]string{"mode"},
		[]interface{}{int64(0)}, // Default mode: ignore unsupported characters
		[]bool{false},
		[]bool{false},
		[]bool{false},
		types.STRING,
	)
	stringHelper.BuiltinMethods["toxml"] = "__string_toxml"

	// Register case-insensitive property/method aliases for DWScript compatibility
	// Task 9.23: .uppercase -> .ToUpper, .lowercase -> .ToLower
	stringHelper.Methods["uppercase"] = types.NewFunctionType([]types.Type{}, types.STRING)
	stringHelper.BuiltinMethods["uppercase"] = "__string_toupper"
	stringHelper.Methods["lowercase"] = types.NewFunctionType([]types.Type{}, types.STRING)
	stringHelper.BuiltinMethods["lowercase"] = "__string_tolower"

	register(types.STRING.String(), stringHelper)

	// String dynamic array helper: Join method (lowercase keys for case-insensitive lookup)
	stringArrayHelper := types.NewHelperType("__TStringDynArrayIntrinsicHelper", nil, true)
	stringArrayHelper.TargetType = types.NewDynamicArrayType(types.STRING)
	stringArrayHelper.Methods["join"] = types.NewFunctionType([]types.Type{types.STRING}, types.STRING)
	stringArrayHelper.BuiltinMethods["join"] = "__string_array_join"
	register(stringArrayHelper.TargetType.String(), stringArrayHelper)
}

// initEnumHelpers registers built-in helpers for enumerated types.
// Task 9.31: Implement enum .Value helper property
// Also implements .Name and .QualifiedName properties
func (a *Analyzer) initEnumHelpers() {
	// Create a helper for the generic ENUM type
	// Since we need to support all enum types, we'll register this for "ENUM"
	enumHelper := types.NewHelperType("__TEnumIntrinsicHelper", nil, false)

	// Task 9.31: Register .Value property (returns ordinal value) - lowercase key for case-insensitive lookup
	enumHelper.Properties["value"] = &types.PropertyInfo{
		Name:      "Value",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__enum_value",
		WriteKind: types.PropAccessNone,
	}

	// Register .Name property (returns enum value name as string) - lowercase key for case-insensitive lookup
	enumHelper.Properties["name"] = &types.PropertyInfo{
		Name:      "Name",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__enum_name",
		WriteKind: types.PropAccessNone,
	}

	// Register .QualifiedName property (returns TypeName.ValueName) - lowercase key for case-insensitive lookup
	enumHelper.Properties["qualifiedname"] = &types.PropertyInfo{
		Name:      "QualifiedName",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__enum_qualifiedname",
		WriteKind: types.PropAccessNone,
	}

	// Register helper for enum type (generic catch-all)
	// This will be checked in getHelpersForType for all enum types
	if a.helpers["enum"] == nil {
		a.helpers["enum"] = make([]*types.HelperType, 0)
	}
	a.helpers["enum"] = append(a.helpers["enum"], enumHelper)
}
