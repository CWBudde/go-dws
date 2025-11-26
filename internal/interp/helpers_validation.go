package interp

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Helper Declaration, Validation, and Execution Infrastructure
// ============================================================================

// HelperInfo stores runtime information about a helper type
type HelperInfo struct {
	TargetType     types.Type
	ParentHelper   *HelperInfo
	Methods        map[string]*ast.FunctionDecl
	Properties     map[string]*types.PropertyInfo
	ClassVars      map[string]Value
	ClassConsts    map[string]Value
	BuiltinMethods map[string]string
	Name           string
	IsRecordHelper bool
}

// NewHelperInfo creates a new HelperInfo
func NewHelperInfo(name string, targetType types.Type, isRecordHelper bool) *HelperInfo {
	return &HelperInfo{
		Name:           name,
		TargetType:     targetType,
		Methods:        make(map[string]*ast.FunctionDecl),
		Properties:     make(map[string]*types.PropertyInfo),
		ClassVars:      make(map[string]Value),
		ClassConsts:    make(map[string]Value),
		BuiltinMethods: make(map[string]string),
		IsRecordHelper: isRecordHelper,
	}
}

// ============================================================================
// Helper Declaration and Validation
// ============================================================================

// evalHelperDeclaration processes a helper declaration at runtime.
func (i *Interpreter) evalHelperDeclaration(decl *ast.HelperDecl) Value {
	if decl == nil {
		return &NilValue{}
	}

	// Resolve the target type
	targetType := i.resolveTypeFromAnnotation(decl.ForType)
	if targetType == nil {
		return i.newErrorWithLocation(decl, "unknown target type '%s' for helper '%s'",
			decl.ForType.String(), decl.Name.Value)
	}

	// Create helper info
	helperInfo := NewHelperInfo(decl.Name.Value, targetType, decl.IsRecordHelper)

	// Resolve parent helper if specified
	if decl.ParentHelper != nil {
		parentHelperName := decl.ParentHelper.Value

		// Look up the parent helper by searching all registered helpers
		var foundParent *HelperInfo
		for _, helpers := range i.helpers {
			for _, helper := range helpers {
				if ident.Equal(helper.Name, parentHelperName) {
					foundParent = helper
					break
				}
			}
			if foundParent != nil {
				break
			}
		}

		if foundParent == nil {
			return i.newErrorWithLocation(decl.ParentHelper,
				"unknown parent helper '%s' for helper '%s'",
				parentHelperName, decl.Name.Value)
		}

		// Verify target type compatibility
		if !foundParent.TargetType.Equals(targetType) {
			return i.newErrorWithLocation(decl.ParentHelper,
				"parent helper '%s' extends type '%s', but child helper '%s' extends type '%s'",
				parentHelperName, foundParent.TargetType.String(),
				decl.Name.Value, targetType.String())
		}

		// Set up inheritance chain
		helperInfo.ParentHelper = foundParent
	}

	// Register methods
	// Task 9.16.2: Normalize method names for case-insensitive lookup
	for _, method := range decl.Methods {
		helperInfo.Methods[ident.Normalize(method.Name.Value)] = method
	}

	// Register properties
	for _, prop := range decl.Properties {
		propType := i.resolveTypeFromAnnotation(prop.Type)
		if propType == nil {
			return i.newErrorWithLocation(prop, "unknown type '%s' for property '%s'",
				prop.Type.String(), prop.Name.Value)
		}

		propInfo := &types.PropertyInfo{
			Name: prop.Name.Value,
			Type: propType,
		}

		// Set up property access - extract identifier from expression
		if prop.ReadSpec != nil {
			if ident, ok := prop.ReadSpec.(*ast.Identifier); ok {
				propInfo.ReadKind = types.PropAccessMethod
				propInfo.ReadSpec = ident.Value
			}
		}
		if prop.WriteSpec != nil {
			if ident, ok := prop.WriteSpec.(*ast.Identifier); ok {
				propInfo.WriteKind = types.PropAccessMethod
				propInfo.WriteSpec = ident.Value
			}
		}

		helperInfo.Properties[prop.Name.Value] = propInfo
	}

	// Initialize class variables
	for _, classVar := range decl.ClassVars {
		varType := i.resolveTypeFromExpression(classVar.Type)
		if varType == nil {
			return i.newErrorWithLocation(classVar, "unknown or invalid type for class variable '%s'",
				classVar.Name.Value)
		}

		// Initialize with default value
		var defaultValue Value
		switch varType {
		case types.INTEGER:
			defaultValue = &IntegerValue{Value: 0}
		case types.FLOAT:
			defaultValue = &FloatValue{Value: 0.0}
		case types.STRING:
			defaultValue = &StringValue{Value: ""}
		case types.BOOLEAN:
			defaultValue = &BooleanValue{Value: false}
		default:
			defaultValue = &NilValue{}
		}

		helperInfo.ClassVars[classVar.Name.Value] = defaultValue
	}

	// Initialize class constants
	for _, classConst := range decl.ClassConsts {
		// Evaluate the constant value
		constValue := i.Eval(classConst.Value)
		if isError(constValue) {
			return constValue
		}
		helperInfo.ClassConsts[classConst.Name.Value] = constValue
	}

	// Register the helper
	if i.helpers == nil {
		i.helpers = make(map[string][]*HelperInfo)
	}

	// Get the type name for indexing
	typeName := ident.Normalize(targetType.String())
	// Register helper in TypeSystem
	i.typeSystem.RegisterHelper(typeName, helperInfo)
	// Also maintain legacy map for backward compatibility during migration
	i.helpers[typeName] = append(i.helpers[typeName], helperInfo)

	// Also register by simple type name for lookup compatibility
	simpleTypeName := ident.Normalize(extractSimpleTypeName(targetType.String()))
	if simpleTypeName != typeName {
		// Register helper by simple name in TypeSystem
		i.typeSystem.RegisterHelper(simpleTypeName, helperInfo)
		// Also maintain legacy map for backward compatibility during migration
		i.helpers[simpleTypeName] = append(i.helpers[simpleTypeName], helperInfo)
	}

	return &NilValue{}
}

// ============================================================================
// Helper Execution
// ============================================================================

// callHelperMethod executes a helper method (user-defined or built-in) on a value
func (i *Interpreter) callHelperMethod(helper *HelperInfo, method *ast.FunctionDecl,
	builtinSpec string, selfValue Value, args []Value, node ast.Node) Value {

	if builtinSpec != "" {
		return i.evalBuiltinHelperMethod(builtinSpec, selfValue, args, node)
	}

	if method == nil {
		return i.newErrorWithLocation(node, "helper method not implemented")
	}

	// Check argument count
	if len(args) != len(method.Parameters) {
		return i.newErrorWithLocation(node, "wrong number of arguments for helper method '%s': expected %d, got %d",
			method.Name.Value, len(method.Parameters), len(args))
	}

	// Create method environment
	methodEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = methodEnv

	// Bind Self to the target value (the value being extended)
	i.env.Define("Self", selfValue)

	// Bind helper class vars and consts from entire inheritance chain
	// Walk from root parent to current helper so child helpers override parents
	var helperChain []*HelperInfo
	for h := helper; h != nil; h = h.ParentHelper {
		helperChain = append([]*HelperInfo{h}, helperChain...)
	}
	for _, h := range helperChain {
		for name, value := range h.ClassVars {
			i.env.Define(name, value)
		}
		for name, value := range h.ClassConsts {
			i.env.Define(name, value)
		}
	}

	// Bind method parameters
	for idx, param := range method.Parameters {
		i.env.Define(param.Name.Value, args[idx])
	}

	// For functions, initialize the Result variable
	if method.ReturnType != nil {
		returnType := i.resolveTypeFromAnnotation(method.ReturnType)
		defaultVal := i.getDefaultValue(returnType)
		i.env.Define("Result", defaultVal)
		i.env.Define(method.Name.Value, defaultVal)
	}

	// Execute method body
	result := i.Eval(method.Body)
	if isError(result) {
		i.env = savedEnv
		return result
	}

	// Extract return value
	var returnValue Value
	if method.ReturnType != nil {
		resultVal, resultOk := i.env.Get("Result")
		methodNameVal, methodNameOk := i.env.Get(method.Name.Value)

		if resultOk && resultVal.Type() != "NIL" {
			returnValue = resultVal
		} else if methodNameOk && methodNameVal.Type() != "NIL" {
			returnValue = methodNameVal
		} else if resultOk {
			returnValue = resultVal
		} else if methodNameOk {
			returnValue = methodNameVal
		} else {
			returnValue = &NilValue{}
		}
	} else {
		returnValue = &NilValue{}
	}

	// Restore environment
	i.env = savedEnv

	return returnValue
}

// evalHelperPropertyRead evaluates a helper property read access
func (i *Interpreter) evalHelperPropertyRead(helper *HelperInfo, propInfo *types.PropertyInfo,
	selfValue Value, node ast.Node) Value {

	switch propInfo.ReadKind {
	case types.PropAccessField:
		// For helpers on records, try to access the field from the record
		if recordVal, ok := selfValue.(*RecordValue); ok {
			if fieldValue, exists := recordVal.Fields[propInfo.ReadSpec]; exists {
				return fieldValue
			}
		}

		// Otherwise, try as a method (getter)
		// Task 9.16.2: Method names are case-insensitive
		normalizedReadSpec := ident.Normalize(propInfo.ReadSpec)

		// Search for the getter method in the owner helper's inheritance chain
		if method, methodOwner, ok := helper.GetMethod(normalizedReadSpec); ok {
			// Get builtin spec if any
			var builtinSpec string
			if spec, _, ok := methodOwner.GetBuiltinMethod(normalizedReadSpec); ok {
				builtinSpec = spec
			}
			// Call the getter method with no arguments
			return i.callHelperMethod(methodOwner, method, builtinSpec, selfValue, []Value{}, node)
		}

		return i.newErrorWithLocation(node, "property '%s' read specifier '%s' not found",
			propInfo.Name, propInfo.ReadSpec)

	case types.PropAccessMethod:
		// Call getter method
		// Task 9.16.2: Method names are case-insensitive
		normalizedReadSpec := ident.Normalize(propInfo.ReadSpec)

		// Search for the getter method in the owner helper's inheritance chain
		if method, methodOwner, ok := helper.GetMethod(normalizedReadSpec); ok {
			// Get builtin spec if any
			var builtinSpec string
			if spec, _, ok := methodOwner.GetBuiltinMethod(normalizedReadSpec); ok {
				builtinSpec = spec
			}
			return i.callHelperMethod(methodOwner, method, builtinSpec, selfValue, []Value{}, node)
		}

		return i.newErrorWithLocation(node, "property '%s' getter method '%s' not found",
			propInfo.Name, propInfo.ReadSpec)

	case types.PropAccessBuiltin:
		// Built-in array helper properties
		return i.evalBuiltinHelperProperty(propInfo.ReadSpec, selfValue, node)

	case types.PropAccessNone:
		return i.newErrorWithLocation(node, "property '%s' is write-only", propInfo.Name)

	default:
		return i.newErrorWithLocation(node, "property '%s' has no read access", propInfo.Name)
	}
}

// ============================================================================
// Built-in Helper Initialization
// ============================================================================

// initArrayHelpers registers built-in helper properties for arrays
// Array Helper Properties (.High, .Low, .Length)
func (i *Interpreter) initArrayHelpers() {
	if i.helpers == nil {
		i.helpers = make(map[string][]*HelperInfo)
	}

	// Create a helper for the generic ARRAY type
	arrayHelper := &HelperInfo{
		Name:           "TArrayHelper",
		TargetType:     nil, // Generic - applies to all arrays
		Methods:        make(map[string]*ast.FunctionDecl),
		Properties:     make(map[string]*types.PropertyInfo),
		ClassVars:      make(map[string]Value),
		ClassConsts:    make(map[string]Value),
		IsRecordHelper: false,
		BuiltinMethods: make(map[string]string),
	}

	// Register .Length property
	arrayHelper.Properties["Length"] = &types.PropertyInfo{
		Name:      "Length",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_length",
		WriteKind: types.PropAccessNone,
	}

	// Register .High property
	arrayHelper.Properties["High"] = &types.PropertyInfo{
		Name:      "High",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_high",
		WriteKind: types.PropAccessNone,
	}

	// Register .Low property
	arrayHelper.Properties["Low"] = &types.PropertyInfo{
		Name:      "Low",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_low",
		WriteKind: types.PropAccessNone,
	}

	// Task 9.34: Register .Count property (alias for .Length)
	arrayHelper.Properties["Count"] = &types.PropertyInfo{
		Name:      "Count",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_count",
		WriteKind: types.PropAccessNone,
	}

	// Register .Add() method for dynamic arrays
	// This allows: arr.Add(value) syntax
	arrayHelper.BuiltinMethods["add"] = "__array_add"

	// Task 9.34: Register .Delete() method for dynamic arrays
	arrayHelper.BuiltinMethods["delete"] = "__array_delete"

	// Task 9.34: Register .IndexOf() method for dynamic arrays
	arrayHelper.BuiltinMethods["indexof"] = "__array_indexof"

	// Register .SetLength() method for dynamic arrays
	// This allows: arr.SetLength(newLength) syntax
	arrayHelper.BuiltinMethods["setlength"] = "__array_setlength"

	// Task 9.8: Register .Swap() method for arrays
	// This allows: arr.Swap(i, j) syntax
	arrayHelper.BuiltinMethods["swap"] = "__array_swap"

	// Task 9.8: Register .Push() method for dynamic arrays (alias for Add)
	// This allows: arr.Push(value) syntax
	arrayHelper.BuiltinMethods["push"] = "__array_push"

	// Task 9.8: Register .Pop() method for dynamic arrays
	// This allows: arr.Pop() syntax - removes and returns last element
	arrayHelper.BuiltinMethods["pop"] = "__array_pop"

	// Register helper for array type
	i.helpers["array"] = append(i.helpers["array"], arrayHelper)
}

// initIntrinsicHelpers registers built-in helpers for primitive types (Integer, Float, Boolean).
func (i *Interpreter) initIntrinsicHelpers() {
	if i.helpers == nil {
		i.helpers = make(map[string][]*HelperInfo)
	}

	register := func(typeName string, helper *HelperInfo) {
		norm := ident.Normalize(typeName)
		i.helpers[norm] = append(i.helpers[norm], helper)
		i.typeSystem.RegisterHelper(typeName, helper)
	}

	// Integer helper
	intHelper := NewHelperInfo("__TIntegerIntrinsicHelper", types.INTEGER, false)
	intHelper.Properties["ToString"] = &types.PropertyInfo{
		Name:      "ToString",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__integer_tostring",
		WriteKind: types.PropAccessNone,
	}
	// Task 9.16.2: Method names are case-insensitive, use lowercase keys
	intHelper.Methods["tostring"] = nil
	intHelper.BuiltinMethods["tostring"] = "__integer_tostring"
	intHelper.Methods["tohexstring"] = nil
	intHelper.BuiltinMethods["tohexstring"] = "__integer_tohexstring"
	register("Integer", intHelper)

	// Float helper
	floatHelper := NewHelperInfo("__TFloatIntrinsicHelper", types.FLOAT, false)
	floatHelper.Properties["ToString"] = &types.PropertyInfo{
		Name:      "ToString",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__float_tostring_default",
		WriteKind: types.PropAccessNone,
	}
	// Task 9.16.2: Method names are case-insensitive, use lowercase keys
	floatHelper.Methods["tostring"] = nil
	floatHelper.BuiltinMethods["tostring"] = "__float_tostring_prec"
	register("Float", floatHelper)

	// Boolean helper
	boolHelper := NewHelperInfo("__TBooleanIntrinsicHelper", types.BOOLEAN, false)
	boolHelper.Properties["ToString"] = &types.PropertyInfo{
		Name:      "ToString",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__boolean_tostring",
		WriteKind: types.PropAccessNone,
	}
	// Task 9.16.2: Method names are case-insensitive, use lowercase keys
	boolHelper.Methods["tostring"] = nil
	boolHelper.BuiltinMethods["tostring"] = "__boolean_tostring"
	register("Boolean", boolHelper)

	// String helper
	stringHelper := NewHelperInfo("__TStringIntrinsicHelper", types.STRING, false)
	stringHelper.Properties["Length"] = &types.PropertyInfo{
		Name:      "Length",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__string_length",
		WriteKind: types.PropAccessNone,
	}
	// Task 9.16.2: Method names are case-insensitive, use lowercase keys
	stringHelper.Methods["toupper"] = nil
	stringHelper.BuiltinMethods["toupper"] = "__string_toupper"
	stringHelper.Methods["tolower"] = nil
	stringHelper.BuiltinMethods["tolower"] = "__string_tolower"
	stringHelper.Methods["matches"] = nil
	stringHelper.BuiltinMethods["matches"] = "__string_matches"

	// String transformation methods
	stringHelper.Methods["padleft"] = nil
	stringHelper.BuiltinMethods["padleft"] = "PadLeft"
	stringHelper.Methods["padright"] = nil
	stringHelper.BuiltinMethods["padright"] = "PadRight"
	stringHelper.Methods["deleteleft"] = nil
	stringHelper.BuiltinMethods["deleteleft"] = "StrDeleteLeft"
	stringHelper.Methods["deleteright"] = nil
	stringHelper.BuiltinMethods["deleteright"] = "StrDeleteRight"
	stringHelper.Methods["normalize"] = nil
	stringHelper.BuiltinMethods["normalize"] = "NormalizeString"
	stringHelper.Methods["stripaccents"] = nil
	stringHelper.BuiltinMethods["stripaccents"] = "StripAccents"

	// Register StripAccents as a property (no-argument methods can be accessed without parentheses)
	stringHelper.Properties["stripaccents"] = &types.PropertyInfo{
		Name:      "StripAccents",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "StripAccents",
		WriteKind: types.PropAccessNone,
	}

	// Task 9.23: Additional string helper methods
	// Conversion methods
	stringHelper.Methods["tointeger"] = nil
	stringHelper.BuiltinMethods["tointeger"] = "__string_tointeger"
	stringHelper.Methods["tofloat"] = nil
	stringHelper.BuiltinMethods["tofloat"] = "__string_tofloat"
	stringHelper.Methods["tostring"] = nil
	stringHelper.BuiltinMethods["tostring"] = "__string_tostring"

	// Search/check methods
	stringHelper.Methods["startswith"] = nil
	stringHelper.BuiltinMethods["startswith"] = "__string_startswith"
	stringHelper.Methods["endswith"] = nil
	stringHelper.BuiltinMethods["endswith"] = "__string_endswith"
	stringHelper.Methods["contains"] = nil
	stringHelper.BuiltinMethods["contains"] = "__string_contains"
	stringHelper.Methods["indexof"] = nil
	stringHelper.BuiltinMethods["indexof"] = "__string_indexof"

	// Extraction methods
	stringHelper.Methods["copy"] = nil
	stringHelper.BuiltinMethods["copy"] = "__string_copy"
	stringHelper.Methods["before"] = nil
	stringHelper.BuiltinMethods["before"] = "__string_before"
	stringHelper.Methods["after"] = nil
	stringHelper.BuiltinMethods["after"] = "__string_after"

	// Modification methods
	stringHelper.Methods["trim"] = nil
	stringHelper.BuiltinMethods["trim"] = "__string_trim"
	stringHelper.Methods["trimleft"] = nil
	stringHelper.BuiltinMethods["trimleft"] = "__string_trimleft"
	stringHelper.Methods["trimright"] = nil
	stringHelper.BuiltinMethods["trimright"] = "__string_trimright"

	// Split/join methods
	stringHelper.Methods["split"] = nil
	stringHelper.BuiltinMethods["split"] = "__string_split"
	stringHelper.Methods["tojson"] = nil
	stringHelper.BuiltinMethods["tojson"] = "__string_tojson"
	stringHelper.Methods["tohtml"] = nil
	stringHelper.BuiltinMethods["tohtml"] = "__string_tohtml"
	stringHelper.Methods["tohtmlattribute"] = nil
	stringHelper.BuiltinMethods["tohtmlattribute"] = "__string_tohtmlattribute"
	stringHelper.Methods["tocsstext"] = nil
	stringHelper.BuiltinMethods["tocsstext"] = "__string_tocsstext"
	stringHelper.Methods["toxml"] = nil
	stringHelper.BuiltinMethods["toxml"] = "__string_toxml"
	stringHelper.Properties["isascii"] = &types.PropertyInfo{
		Name:      "IsASCII",
		Type:      types.BOOLEAN,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__string_isascii",
		WriteKind: types.PropAccessNone,
	}

	// Case-insensitive aliases for DWScript compatibility
	stringHelper.Methods["uppercase"] = nil
	stringHelper.BuiltinMethods["uppercase"] = "__string_toupper"
	stringHelper.Methods["lowercase"] = nil
	stringHelper.BuiltinMethods["lowercase"] = "__string_tolower"

	register("String", stringHelper)

	// String dynamic array helper
	stringArrayType := types.NewDynamicArrayType(types.STRING)
	stringArrayHelper := NewHelperInfo("__TStringDynArrayIntrinsicHelper", stringArrayType, true)
	// Task 9.16.2: Method names are case-insensitive, use lowercase keys
	stringArrayHelper.Methods["join"] = nil
	stringArrayHelper.BuiltinMethods["join"] = "__string_array_join"
	register(stringArrayType.String(), stringArrayHelper)

	// Generic array helper additions
	arrayHelper := NewHelperInfo("TArrayHelper", nil, false)
	arrayHelper.Methods["map"] = nil
	arrayHelper.BuiltinMethods["map"] = "__array_map"
	arrayHelper.Methods["join"] = nil
	arrayHelper.BuiltinMethods["join"] = "__array_join"
	i.helpers["array"] = append(i.helpers["array"], arrayHelper)
	i.typeSystem.RegisterHelper("array", arrayHelper)
}

// initEnumHelpers registers built-in helpers for enumerated types.
func (i *Interpreter) initEnumHelpers() {
	if i.helpers == nil {
		i.helpers = make(map[string][]*HelperInfo)
	}

	// Create a helper for the generic ENUM type
	enumHelper := NewHelperInfo("__TEnumIntrinsicHelper", nil, false)

	// Register .Value property (returns ordinal value)
	enumHelper.Properties["Value"] = &types.PropertyInfo{
		Name:      "Value",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__enum_value",
		WriteKind: types.PropAccessNone,
	}

	// Register .Name property (returns enum value name as string)
	enumHelper.Properties["Name"] = &types.PropertyInfo{
		Name:      "Name",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__enum_name",
		WriteKind: types.PropAccessNone,
	}

	// Register .QualifiedName property (returns TypeName.ValueName)
	enumHelper.Properties["QualifiedName"] = &types.PropertyInfo{
		Name:      "QualifiedName",
		Type:      types.STRING,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__enum_qualifiedname",
		WriteKind: types.PropAccessNone,
	}

	// Register helper for enum type (generic catch-all)
	i.helpers["enum"] = append(i.helpers["enum"], enumHelper)
}
