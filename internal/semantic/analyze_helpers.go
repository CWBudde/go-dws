package semantic

import (
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// ============================================================================
// Helper Type Analysis
// ============================================================================

// findPropertyCaseInsensitive searches for a property by name using case-insensitive comparison.
func findPropertyCaseInsensitive(props map[string]*types.PropertyInfo, name string) *types.PropertyInfo {
	for key, prop := range props {
		if ident.Equal(key, name) {
			return prop
		}
	}
	return nil
}

// findMethodCaseInsensitive searches for a method by name using case-insensitive comparison.
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
	helperType.IsClassHelper = decl.IsClassHelper
	helperType.IsStrict = decl.IsStrict
	helperType.Decl = decl // Store AST declaration for runtime

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

	// First collect the complete helper surface. Bodies are analyzed only after
	// every member is registered so inline methods can call later helper members.
	for _, method := range decl.Methods {
		a.analyzeHelperMethod(method, helperType, helperName)
	}
	for _, prop := range decl.Properties {
		a.analyzeHelperProperty(prop, helperType, helperName)
	}
	for _, classVar := range decl.ClassVars {
		a.analyzeHelperClassVar(classVar, helperType, helperName)
	}
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
	declaredTargetName := ident.Normalize(getTypeExpressionName(decl.ForType))
	if declaredTargetName != "" && declaredTargetName != targetTypeName {
		a.helpers[declaredTargetName] = append(a.helpers[declaredTargetName], helperType)
	}

	// Also register the helper type itself in the symbol table
	// so it can be referenced by name (e.g., TStringHelper.PI)
	a.symbols.Define(helperName, helperType, decl.Token.Pos)

	for _, method := range decl.Methods {
		a.analyzeHelperMethodBody(method, helperType)
	}
}

func (a *Analyzer) analyzeFunctionHelperDecl(decl *ast.FunctionDecl, paramTypes []types.Type, returnType types.Type) {
	if decl == nil {
		return
	}
	if len(paramTypes) == 0 {
		a.addError("helper function '%s' must declare at least one parameter", decl.Name.Value)
		return
	}

	targetType := paramTypes[0]
	helperName := "__" + decl.Name.Value + "FunctionHelper"
	if decl.HelperName != nil {
		helperName = "__" + decl.HelperName.Value + "FunctionHelper"
	}

	helperType := types.NewHelperType(helperName, targetType, false)
	helperType.Decl = decl

	methodName := decl.Name.Value
	if decl.HelperName != nil {
		methodName = decl.HelperName.Value
	}
	methodParams := append([]types.Type(nil), paramTypes[1:]...)
	var funcType *types.FunctionType
	if returnType == types.VOID {
		funcType = types.NewProcedureType(methodParams)
	} else {
		funcType = types.NewFunctionType(methodParams, returnType)
	}
	helperType.Methods[ident.Normalize(methodName)] = funcType

	targetTypeName := ident.Normalize(targetType.String())
	if a.helpers[targetTypeName] == nil {
		a.helpers[targetTypeName] = make([]*types.HelperType, 0)
	}
	a.helpers[targetTypeName] = append(a.helpers[targetTypeName], helperType)
}

func (a *Analyzer) getHelperType(name string) *types.HelperType {
	if sym, ok := a.symbols.Resolve(name); ok {
		if helperType, ok := sym.Type.(*types.HelperType); ok {
			return helperType
		}
	}
	for _, helpers := range a.helpers {
		for _, helper := range helpers {
			if ident.Equal(helper.Name, name) {
				return helper
			}
		}
	}
	return nil
}

func (a *Analyzer) analyzeHelperMethodImplementation(decl *ast.FunctionDecl, helperType *types.HelperType) {
	if !a.helperMethodImplementationMatchesDeclaration(decl, helperType) {
		a.analyzeHelperMethod(decl, helperType, helperType.Name)
	}
	a.analyzeHelperMethodBody(decl, helperType)
}

func (a *Analyzer) helperMethodImplementationMatchesDeclaration(decl *ast.FunctionDecl, helperType *types.HelperType) bool {
	if decl == nil || helperType == nil {
		return false
	}
	overloads := helperType.MethodOverloads[ident.Normalize(decl.Name.Value)]
	if len(overloads) == 0 {
		return false
	}

	paramTypes := make([]types.Type, 0, len(decl.Parameters))
	for _, param := range decl.Parameters {
		paramType, err := a.resolveTypeExpression(param.Type)
		if err != nil || paramType == nil {
			return false
		}
		paramTypes = append(paramTypes, paramType)
	}

	var returnType types.Type = types.VOID
	if decl.ReturnType != nil {
		resolvedReturn, err := a.resolveTypeExpression(decl.ReturnType)
		if err != nil || resolvedReturn == nil {
			return false
		}
		returnType = resolvedReturn
	}

	for _, overload := range overloads {
		if overload == nil || len(overload.Parameters) != len(paramTypes) {
			continue
		}
		matches := true
		for i, paramType := range paramTypes {
			if !types.IsIdentical(overload.Parameters[i], paramType) {
				matches = false
				break
			}
		}
		if matches && types.IsIdentical(overload.ReturnType, returnType) {
			return true
		}
	}
	return false
}

func (a *Analyzer) analyzeHelperMethodBody(decl *ast.FunctionDecl, helperType *types.HelperType) {
	if decl == nil || decl.Body == nil || helperType == nil {
		return
	}

	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(oldSymbols)
	defer func() { a.symbols = oldSymbols }()
	defer a.emitUnusedWarningsForCurrentScope()

	a.symbols.Define("Self", helperType.TargetType, decl.Token.Pos)
	for name, varType := range helperType.ClassVars {
		a.symbols.Define(name, varType, token.Position{})
	}
	for name, constType := range helperType.ClassConsts {
		if typ, ok := constType.(types.Type); ok {
			a.symbols.DefineConst(name, typ, nil, token.Position{})
		}
	}
	for name, overloads := range helperType.MethodOverloads {
		for _, methodType := range overloads {
			if err := a.symbols.DefineOverload(name, methodType, true, false, token.Position{}); err != nil {
				a.symbols.DefineFunction(name, methodType, token.Position{})
			}
		}
	}

	if classType, ok := types.GetUnderlyingType(helperType.TargetType).(*types.ClassType); ok && !decl.IsClassMethod {
		for fieldName, fieldType := range classType.Fields {
			a.symbols.Define(fieldName, fieldType, token.Position{})
		}
	}
	if recordType, ok := types.GetUnderlyingType(helperType.TargetType).(*types.RecordType); ok {
		for fieldName, fieldType := range recordType.Fields {
			a.symbols.Define(recordType.FieldNames[fieldName], fieldType, token.Position{})
		}
	}

	for _, param := range decl.Parameters {
		if param.Type == nil {
			continue
		}
		paramType, err := a.resolveTypeExpression(param.Type)
		if err != nil || paramType == nil {
			continue
		}
		a.symbols.DefineParameter(param.Name.Value, paramType, param.Name.Token.Pos, param.IsConst)
	}

	var returnType types.Type = types.VOID
	if decl.ReturnType != nil {
		if resolved, err := a.resolveTypeExpression(decl.ReturnType); err == nil && resolved != nil {
			returnType = resolved
		}
	}
	if returnType != types.VOID {
		a.symbols.Define("Result", returnType, decl.Name.Token.Pos)
	}

	prevFunc := a.currentFunction
	prevSelf := a.currentSelfType
	prevClass := a.currentClass
	prevRecord := a.currentRecord
	prevInClassMethod := a.inClassMethod
	a.currentFunction = decl
	a.currentSelfType = helperType.TargetType
	a.inClassMethod = decl.IsClassMethod
	if classType, ok := types.GetUnderlyingType(helperType.TargetType).(*types.ClassType); ok {
		a.currentClass = classType
		a.currentRecord = nil
	}
	if recordType, ok := types.GetUnderlyingType(helperType.TargetType).(*types.RecordType); ok {
		a.currentRecord = recordType
		a.currentClass = nil
	}
	defer func() {
		a.currentFunction = prevFunc
		a.currentSelfType = prevSelf
		a.currentClass = prevClass
		a.currentRecord = prevRecord
		a.inClassMethod = prevInClassMethod
	}()

	a.analyzeBlock(decl.Body)
}

// analyzeHelperMethod analyzes a method in a helper.
// Note: In helper methods, 'Self' refers to the target type instance.
func (a *Analyzer) analyzeHelperMethod(method *ast.FunctionDecl, helperType *types.HelperType, helperName string) {
	if method == nil {
		return
	}

	methodName := method.Name.Value

	methodNameLower := ident.Normalize(methodName)
	if _, exists := helperType.Methods[methodNameLower]; exists && !method.IsOverload {
		a.addError("%s", errors.FormatNameAlreadyExists(methodName, method.Token.Pos.Line, method.Token.Pos.Column))
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
	helperType.Methods[methodNameLower] = funcType
	helperType.MethodOverloads[methodNameLower] = append(helperType.MethodOverloads[methodNameLower], funcType)
}

// analyzeHelperProperty analyzes a property in a helper.
func (a *Analyzer) analyzeHelperProperty(prop *ast.PropertyDecl, helperType *types.HelperType, helperName string) {
	if prop == nil {
		return
	}

	propName := prop.Name.Value

	// Check for duplicate properties
	if _, exists := helperType.Properties[propName]; exists {
		a.addError("%s", errors.FormatNameAlreadyExists(propName, prop.Token.Pos.Line, prop.Token.Pos.Column))
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
		a.addError("%s", errors.FormatNameAlreadyExists(varName, classVar.Token.Pos.Line, classVar.Token.Pos.Column))
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
		a.addError("%s", errors.FormatNameAlreadyExists(constName, classConst.Token.Pos.Line, classConst.Token.Pos.Column))
		return
	}

	var constType types.Type
	if classConst.Type != nil {
		expectedType, err := a.resolveType(getTypeExpressionName(classConst.Type))
		if err != nil {
			a.addError("unknown type '%s' for constant '%s' in helper '%s' at %s",
				getTypeExpressionName(classConst.Type), constName, helperName, classConst.Token.Pos.String())
			return
		}

		constType = a.analyzeExpressionWithExpectedType(classConst.Value, expectedType)
		if constType == nil {
			a.addError("invalid constant value for '%s' in helper '%s' at %s",
				constName, helperName, classConst.Token.Pos.String())
			return
		}
		if !a.canAssign(constType, expectedType) {
			a.addError("constant '%s' type mismatch: cannot assign %s to %s in helper '%s' at %s",
				constName, constType.String(), expectedType.String(), helperName, classConst.Token.Pos.String())
			return
		}
		constType = expectedType
	} else {
		constType = a.analyzeExpression(classConst.Value)
		if constType == nil {
			a.addError("invalid constant value for '%s' in helper '%s' at %s",
				constName, helperName, classConst.Token.Pos.String())
			return
		}
	}

	// Store the constant type
	constNameLower := ident.Normalize(constName)
	helperType.ClassConsts[constNameLower] = constType
}

// getHelpersForType returns all helpers that extend the given type.
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

	// For array types, also include generic array helpers
	if _, isArray := typ.(*types.ArrayType); isArray {
		arrayHelpers := a.helpers["array"]
		if arrayHelpers != nil {
			// Combine type-specific helpers with generic array helpers
			helpers = append(helpers, arrayHelpers...)
		}
	}

	// For enum types, also include generic enum helpers
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
// Returns the method if found (helper type not used by callers).
func (a *Analyzer) hasHelperMethod(typ types.Type, methodName string) *types.FunctionType {
	helpers := a.getHelpersForType(typ)
	if helpers == nil {
		return nil
	}

	// Check helpers in reverse order (user-defined take precedence over built-ins)
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
					return specialized
				}
			}
			return method
		}
	}

	return nil
}

func (a *Analyzer) resolveHelperMethodForCall(typ types.Type, methodName string, args []ast.Expression) *types.FunctionType {
	helpers := a.getHelpersForType(typ)
	if helpers == nil {
		return nil
	}

	methodNameLower := ident.Normalize(methodName)
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
		overloads := helper.MethodOverloads[methodNameLower]
		if len(overloads) == 0 {
			if method := findMethodCaseInsensitive(helper.Methods, methodName); method != nil {
				return method
			}
			continue
		}
		if len(overloads) == 1 {
			return overloads[0]
		}

		argTypes := make([]types.Type, len(args))
		for i, arg := range args {
			argType := a.analyzeExpression(arg)
			if argType == nil {
				return nil
			}
			argTypes[i] = argType
		}
		candidates := make([]*Symbol, len(overloads))
		for i, overload := range overloads {
			candidates[i] = &Symbol{Type: overload}
		}
		selected, err := ResolveOverload(candidates, argTypes)
		if err != nil {
			a.addStructuredError(NewNoOverloadMatchError(argsPositionFallback(args), methodName))
			return nil
		}
		if methodType, ok := selected.Type.(*types.FunctionType); ok {
			return methodType
		}
	}

	return nil
}

func argsPositionFallback(args []ast.Expression) token.Position {
	if len(args) > 0 {
		return args[0].Pos()
	}
	return token.Position{}
}

// hasHelperProperty checks if any helper for the given type defines the specified property.
// Returns the property if found (helper type not used by callers).
func (a *Analyzer) hasHelperProperty(typ types.Type, propName string) *types.PropertyInfo {
	helpers := a.getHelpersForType(typ)
	if helpers == nil {
		return nil
	}

	// Check helpers in reverse order (user-defined take precedence over built-ins)
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
		if prop := findPropertyCaseInsensitive(helper.Properties, propName); prop != nil {
			return prop
		}
	}

	return nil
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
func (a *Analyzer) hasHelperClassConst(typ types.Type, constName string) (*types.HelperType, interface{}) {
	helpers := a.getHelpersForType(typ)
	if helpers == nil {
		return nil, nil
	}

	// Check helpers in reverse order (user-defined take precedence over built-ins)
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

// initArrayHelpers registers built-in helper properties and methods for arrays.
func (a *Analyzer) initArrayHelpers() {
	// Create a helper for the generic ARRAY type
	arrayHelper := types.NewHelperType("TArrayHelper", nil, false)

	// Register .Length property
	arrayHelper.Properties["length"] = &types.PropertyInfo{
		Name:      "Length",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_length",
		WriteKind: types.PropAccessNone,
	}

	// Register .High property
	arrayHelper.Properties["high"] = &types.PropertyInfo{
		Name:      "High",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_high",
		WriteKind: types.PropAccessNone,
	}

	// Register .Low property
	arrayHelper.Properties["low"] = &types.PropertyInfo{
		Name:      "Low",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_low",
		WriteKind: types.PropAccessNone,
	}

	// Register .Count property (alias for .Length)
	arrayHelper.Properties["count"] = &types.PropertyInfo{
		Name:      "Count",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_count",
		WriteKind: types.PropAccessNone,
	}

	// Register .Add() method for dynamic arrays
	arrayHelper.Methods["add"] = types.NewProcedureType([]types.Type{nil})
	arrayHelper.BuiltinMethods["add"] = "__array_add"

	// Register .Delete() method: Delete(index) or Delete(index, count)
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

	// Register .IndexOf() method: IndexOf(value) or IndexOf(value, startIndex)
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

	// Register .SetLength() method for dynamic arrays
	arrayHelper.Methods["setlength"] = types.NewProcedureType([]types.Type{types.INTEGER})
	arrayHelper.BuiltinMethods["setlength"] = "__array_setlength"

	// Register .Swap() method: Swap(i, j) - swaps elements at indices i and j
	arrayHelper.Methods["swap"] = types.NewProcedureType([]types.Type{types.INTEGER, types.INTEGER})
	arrayHelper.BuiltinMethods["swap"] = "__array_swap"

	// Register .Push() method: Push(value) - appends element (alias for Add)
	arrayHelper.Methods["push"] = types.NewProcedureType([]types.Type{nil})
	arrayHelper.BuiltinMethods["push"] = "__array_push"

	// Register .Pop() method: Pop() - removes and returns last element
	// Use VARIANT as placeholder - will be specialized to array's element type by hasHelperMethod
	arrayHelper.Methods["pop"] = types.NewFunctionType([]types.Type{}, types.VARIANT)
	arrayHelper.BuiltinMethods["pop"] = "__array_pop"

	// Register helper for array type
	a.helpers["array"] = append(a.helpers["array"], arrayHelper)

	// Additional array helper methods
	arrayHelper.Methods["map"] = types.NewFunctionType([]types.Type{types.NewFunctionPointerType([]types.Type{types.VARIANT}, types.VARIANT)}, types.NewDynamicArrayType(types.VARIANT))
	arrayHelper.BuiltinMethods["map"] = "__array_map"
	arrayHelper.Methods["join"] = types.NewFunctionType([]types.Type{types.STRING}, types.STRING)
	arrayHelper.BuiltinMethods["join"] = "__array_join"
}

// initIntrinsicHelpers registers built-in helpers for primitive types (Integer, Float, Boolean, String).
func (a *Analyzer) initIntrinsicHelpers() {
	// Helper function to register helpers
	register := func(typeName string, helper *types.HelperType) {
		key := ident.Normalize(typeName)
		if a.helpers[key] == nil {
			a.helpers[key] = make([]*types.HelperType, 0)
		}
		a.helpers[key] = append(a.helpers[key], helper)
	}

	// Integer helper: provides ToString method/property
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

	// Float helper: default ToString property and precision-aware method
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

	// Boolean helper: ToString method/property returning 'True'/'False'
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

	// Conversion helper methods
	stringHelper.Methods["tointeger"] = types.NewFunctionType([]types.Type{}, types.INTEGER)
	stringHelper.BuiltinMethods["tointeger"] = "__string_tointeger"
	stringHelper.Methods["tofloat"] = types.NewFunctionType([]types.Type{}, types.FLOAT)
	stringHelper.BuiltinMethods["tofloat"] = "__string_tofloat"
	stringHelper.Methods["tostring"] = types.NewFunctionType([]types.Type{}, types.STRING)
	stringHelper.BuiltinMethods["tostring"] = "__string_tostring"

	// Search/check helper methods
	stringHelper.Methods["startswith"] = types.NewFunctionType([]types.Type{types.STRING}, types.BOOLEAN)
	stringHelper.BuiltinMethods["startswith"] = "__string_startswith"
	stringHelper.Methods["endswith"] = types.NewFunctionType([]types.Type{types.STRING}, types.BOOLEAN)
	stringHelper.BuiltinMethods["endswith"] = "__string_endswith"
	stringHelper.Methods["contains"] = types.NewFunctionType([]types.Type{types.STRING}, types.BOOLEAN)
	stringHelper.BuiltinMethods["contains"] = "__string_contains"
	// Register .IndexOf() method: IndexOf(substring) or IndexOf(substring, startIndex)
	// Second parameter (startIndex) is optional with default value 1 (1-based indexing in DWScript)
	stringHelper.Methods["indexof"] = types.NewFunctionTypeWithMetadata(
		[]types.Type{types.STRING, types.INTEGER},
		[]string{"substring", "startIndex"},
		[]interface{}{nil, int64(1)}, // Default startIndex = 1 (1-based)
		[]bool{false, false},
		[]bool{false, false},
		[]bool{false, false},
		types.INTEGER,
	)
	stringHelper.BuiltinMethods["indexof"] = "__string_indexof"
	stringHelper.Methods["matches"] = types.NewFunctionType([]types.Type{types.STRING}, types.BOOLEAN)
	stringHelper.BuiltinMethods["matches"] = "__string_matches"
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

	// Extraction helper methods
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
	stringHelper.Methods["before"] = types.NewFunctionType([]types.Type{types.STRING}, types.STRING)
	stringHelper.BuiltinMethods["before"] = "__string_before"
	stringHelper.Methods["after"] = types.NewFunctionType([]types.Type{types.STRING}, types.STRING)
	stringHelper.BuiltinMethods["after"] = "__string_after"

	// Modification helper methods
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

	// Split/join helper methods
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

	// Case aliases for DWScript compatibility
	stringHelper.Methods["uppercase"] = types.NewFunctionType([]types.Type{}, types.STRING)
	stringHelper.BuiltinMethods["uppercase"] = "__string_toupper"
	stringHelper.Methods["lowercase"] = types.NewFunctionType([]types.Type{}, types.STRING)
	stringHelper.BuiltinMethods["lowercase"] = "__string_tolower"

	register(types.STRING.String(), stringHelper)

	// String dynamic array helper: Join method
	stringArrayHelper := types.NewHelperType("__TStringDynArrayIntrinsicHelper", nil, true)
	stringArrayHelper.TargetType = types.NewDynamicArrayType(types.STRING)
	stringArrayHelper.Methods["join"] = types.NewFunctionType([]types.Type{types.STRING}, types.STRING)
	stringArrayHelper.BuiltinMethods["join"] = "__string_array_join"
	register(stringArrayHelper.TargetType.String(), stringArrayHelper)
}

// initEnumHelpers registers built-in helpers for enumerated types.
// Implements .Value, .Name and .QualifiedName properties for all enums.
func (a *Analyzer) initEnumHelpers() {
	// Create a helper for the generic ENUM type
	enumHelper := types.NewHelperType("__TEnumIntrinsicHelper", nil, false)

	// Register .Value property (returns ordinal value)
	enumHelper.Properties["value"] = &types.PropertyInfo{
		Name:      "Value",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__enum_value",
		WriteKind: types.PropAccessNone,
	}

	// Register .Name property (returns enum value name as string)
	enumHelper.Properties["name"] = &types.PropertyInfo{
		Name:      "Name",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__enum_name",
		WriteKind: types.PropAccessNone,
	}

	// Register .QualifiedName property (returns TypeName.ValueName)
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
