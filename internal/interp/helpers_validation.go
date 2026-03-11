package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Helper Declaration, Validation, and Execution Infrastructure
// ============================================================================

type HelperInfo = runtime.MutableHelperInfo

var NewHelperInfo = runtime.NewMutableHelperInfo

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
		for _, helpers := range i.typeSystem.AllHelpers() {
			for _, helper := range helpers {
				if hi, ok := helper.(*HelperInfo); ok && ident.Equal(hi.Name, parentHelperName) {
					foundParent = hi
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

	// Register methods (case-insensitive lookup)
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
		var initialValue Value

		// Evaluate initializer if present (used both for value and potential type inference)
		if classVar.InitValue != nil {
			val := i.Eval(classVar.InitValue)
			if isError(val) {
				return val
			}
			initialValue = val

			// If no explicit type, infer from the initializer
			if varType == nil {
				varType = i.inferTypeFromValue(val)
				if varType == nil {
					return i.newErrorWithLocation(classVar, "cannot infer type for class variable '%s'",
						classVar.Name.Value)
				}
			}
		}

		if varType == nil {
			return i.newErrorWithLocation(classVar, "unknown or invalid type for class variable '%s'",
				classVar.Name.Value)
		}

		// Initialize with default value
		if initialValue == nil {
			switch varType {
			case types.INTEGER:
				initialValue = &IntegerValue{Value: 0}
			case types.FLOAT:
				initialValue = &FloatValue{Value: 0.0}
			case types.STRING:
				initialValue = &StringValue{Value: ""}
			case types.BOOLEAN:
				initialValue = &BooleanValue{Value: false}
			default:
				initialValue = &NilValue{}
			}
		}

		helperInfo.ClassVars[ident.Normalize(classVar.Name.Value)] = initialValue
	}

	// Initialize class constants
	for _, classConst := range decl.ClassConsts {
		// Evaluate the constant value
		constValue := i.Eval(classConst.Value)
		if isError(constValue) {
			return constValue
		}
		helperInfo.ClassConsts[ident.Normalize(classConst.Name.Value)] = constValue
	}

	// Register the helper
	// Get the type name for indexing
	typeName := ident.Normalize(targetType.String())
	// Register helper in TypeSystem
	i.typeSystem.RegisterHelper(typeName, helperInfo)

	// Also register by simple type name for lookup compatibility
	simpleTypeName := ident.Normalize(extractSimpleTypeName(targetType.String()))
	if simpleTypeName != typeName {
		// Register helper by simple name in TypeSystem
		i.typeSystem.RegisterHelper(simpleTypeName, helperInfo)
	}

	// Expose helper name as a type meta value so static access (e.g., TDummy.Hello) resolves
	helperTypeMeta := NewTypeMetaValue(targetType, targetType.String())
	i.Env().Define(decl.Name.Value, helperTypeMeta)

	return &NilValue{}
}

// ============================================================================
// Built-in Helper Initialization
// ============================================================================

// initArrayHelpers registers built-in helper properties for arrays
// Array Helper Properties (.High, .Low, .Length)
func (i *Interpreter) initArrayHelpers() {
	register := func(typeName string, helper *HelperInfo) {
		if i.typeSystem != nil {
			i.typeSystem.RegisterHelper(typeName, helper)
		}
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

	// Array properties: Length, High, Low, Count
	arrayHelper.Properties["Length"] = &types.PropertyInfo{
		Name:      "Length",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_length",
		WriteKind: types.PropAccessNone,
	}

	arrayHelper.Properties["High"] = &types.PropertyInfo{
		Name:      "High",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_high",
		WriteKind: types.PropAccessNone,
	}

	arrayHelper.Properties["Low"] = &types.PropertyInfo{
		Name:      "Low",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_low",
		WriteKind: types.PropAccessNone,
	}

	arrayHelper.Properties["Count"] = &types.PropertyInfo{
		Name:      "Count",
		Type:      types.INTEGER,
		ReadKind:  types.PropAccessBuiltin,
		ReadSpec:  "__array_count",
		WriteKind: types.PropAccessNone,
	}

	// Array methods: Add, Delete, IndexOf, SetLength, Swap, Push, Pop
	arrayHelper.BuiltinMethods["add"] = "__array_add"
	arrayHelper.BuiltinMethods["delete"] = "__array_delete"
	arrayHelper.BuiltinMethods["indexof"] = "__array_indexof"
	arrayHelper.BuiltinMethods["setlength"] = "__array_setlength"
	arrayHelper.BuiltinMethods["swap"] = "__array_swap"
	arrayHelper.BuiltinMethods["push"] = "__array_push"
	arrayHelper.BuiltinMethods["pop"] = "__array_pop"

	// Register helper for array type
	register("array", arrayHelper)
}

// initIntrinsicHelpers registers built-in helpers for primitive types (Integer, Float, Boolean).
func (i *Interpreter) initIntrinsicHelpers() {
	register := func(typeName string, helper *HelperInfo) {
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
	// Case conversion and matching methods
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

	// Modification methods (also available as properties)
	stringHelper.Methods["trim"] = nil
	stringHelper.BuiltinMethods["trim"] = "__string_trim"
	stringHelper.Methods["trimleft"] = nil
	stringHelper.BuiltinMethods["trimleft"] = "__string_trimleft"
	stringHelper.Methods["trimright"] = nil
	stringHelper.BuiltinMethods["trimright"] = "__string_trimright"
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

	// Split/join and encoding methods
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
	stringArrayHelper.Methods["join"] = nil
	stringArrayHelper.BuiltinMethods["join"] = "__string_array_join"
	register(stringArrayType.String(), stringArrayHelper)

	// Generic array helper additions
	arrayHelper := NewHelperInfo("TArrayHelper", nil, false)
	arrayHelper.Methods["map"] = nil
	arrayHelper.BuiltinMethods["map"] = "__array_map"
	arrayHelper.Methods["join"] = nil
	arrayHelper.BuiltinMethods["join"] = "__array_join"
	i.typeSystem.RegisterHelper("array", arrayHelper)
}

// initEnumHelpers registers built-in helpers for enumerated types.
func (i *Interpreter) initEnumHelpers() {
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
	i.typeSystem.RegisterHelper("enum", enumHelper)
}
