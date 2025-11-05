package interp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Helper Support
// ============================================================================

// HelperInfo stores runtime information about a helper type
type HelperInfo struct {
	Name           string                         // Helper type name
	TargetType     types.Type                     // The type being extended
	Methods        map[string]*ast.FunctionDecl   // Helper methods
	Properties     map[string]*types.PropertyInfo // Helper properties
	ClassVars      map[string]Value               // Class variable values
	ClassConsts    map[string]Value               // Class constant values
	IsRecordHelper bool                           // true if "record helper"
	BuiltinMethods map[string]string              // Built-in method implementations (name -> spec)
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

// findPropertyCaseInsensitive searches for a property by name using case-insensitive comparison.
// Task 9.217: Support case-insensitive helper property lookup
func findPropertyCaseInsensitive(props map[string]*types.PropertyInfo, name string) *types.PropertyInfo {
	for key, prop := range props {
		if strings.EqualFold(key, name) {
			return prop
		}
	}
	return nil
}

// findMethodCaseInsensitive searches for a method by name using case-insensitive comparison.
// Task 9.217: Support case-insensitive helper method lookup
func findMethodCaseInsensitive(methods map[string]*ast.FunctionDecl, name string) *ast.FunctionDecl {
	for key, method := range methods {
		if strings.EqualFold(key, name) {
			return method
		}
	}
	return nil
}

// findBuiltinMethodCaseInsensitive searches for a builtin method spec by name using case-insensitive comparison.
// Task 9.217: Support case-insensitive builtin method lookup
func findBuiltinMethodCaseInsensitive(builtinMethods map[string]string, name string) (string, bool) {
	for key, spec := range builtinMethods {
		if strings.EqualFold(key, name) {
			return spec, true
		}
	}
	return "", false
}

// getDefaultValue returns the default/zero value for a given type.
// This is used for Result variable initialization in functions.
// Task 9.221: Fix STRING + NIL by initializing string results to empty string
func (i *Interpreter) getDefaultValue(typ types.Type) Value {
	if typ == nil {
		return &NilValue{}
	}

	switch typ.TypeKind() {
	case "STRING":
		return &StringValue{Value: ""}
	case "INTEGER":
		return &IntegerValue{Value: 0}
	case "FLOAT":
		return &FloatValue{Value: 0.0}
	case "BOOLEAN":
		return &BooleanValue{Value: false}
	case "CLASS", "INTERFACE", "FUNCTION_POINTER", "METHOD_POINTER":
		return &NilValue{}
	case "ARRAY":
		// Dynamic arrays default to NIL
		return &NilValue{}
	case "RECORD":
		// Records should be initialized with default field values
		// For now, return NIL (will be enhanced in future tasks if needed)
		return &NilValue{}
	default:
		// Unknown types default to NIL
		return &NilValue{}
	}
}

// evalHelperDeclaration processes a helper declaration at runtime.
// Task 9.86: Implement helper method dispatch
// Task 9.87: Implement helper method storage (class vars/consts)
func (i *Interpreter) evalHelperDeclaration(decl *ast.HelperDecl) Value {
	if decl == nil {
		return &NilValue{}
	}

	// Resolve the target type
	targetType := i.resolveTypeFromAnnotation(decl.ForType)
	if targetType == nil {
		return i.newErrorWithLocation(decl, "unknown target type '%s' for helper '%s'",
			decl.ForType.Name, decl.Name.Value)
	}

	// Create helper info
	helperInfo := NewHelperInfo(decl.Name.Value, targetType, decl.IsRecordHelper)

	// Register methods
	for _, method := range decl.Methods {
		helperInfo.Methods[method.Name.Value] = method
	}

	// Register properties
	for _, prop := range decl.Properties {
		propType := i.resolveTypeFromAnnotation(prop.Type)
		if propType == nil {
			return i.newErrorWithLocation(prop, "unknown type '%s' for property '%s'",
				prop.Type.Name, prop.Name.Value)
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
	typeName := targetType.String()
	i.helpers[typeName] = append(i.helpers[typeName], helperInfo)

	// Also register by simple type name for lookup compatibility
	simpleTypeName := extractSimpleTypeName(typeName)
	if simpleTypeName != typeName {
		i.helpers[simpleTypeName] = append(i.helpers[simpleTypeName], helperInfo)
	}

	return &NilValue{}
}

// getHelpersForValue returns all helpers that apply to the given value's type
func (i *Interpreter) getHelpersForValue(val Value) []*HelperInfo {
	if i.helpers == nil {
		return nil
	}

	// Get the type name from the value
	var typeName string
	switch v := val.(type) {
	case *IntegerValue:
		typeName = "Integer"
	case *FloatValue:
		typeName = "Float"
	case *StringValue:
		typeName = "String"
	case *BooleanValue:
		typeName = "Boolean"
	case *ObjectInstance:
		typeName = v.Class.Name
	case *RecordValue:
		typeName = v.RecordType.Name
	case *ArrayValue:
		// Task 9.171: Array helper properties support
		// First try specific array type (e.g., "array of String"), then generic ARRAY helpers
		specific := v.ArrayType.String()
		var combined []*HelperInfo
		if h, ok := i.helpers[specific]; ok {
			combined = append(combined, h...)
		}
		if h, ok := i.helpers["ARRAY"]; ok {
			combined = append(combined, h...)
		}
		return combined
	default:
		// For other types, try to extract type name from Type() method
		typeName = v.Type()
	}

	// Look up helpers for this type
	return i.helpers[typeName]
}

// findHelperMethod searches all applicable helpers for a method with the given name
// and returns the helper, method declaration (if any), and builtin specification identifier.
func (i *Interpreter) findHelperMethod(val Value, methodName string) (*HelperInfo, *ast.FunctionDecl, string) {
	helpers := i.getHelpersForValue(val)
	if helpers == nil {
		return nil, nil, ""
	}

	// Search helpers in reverse order so later (user-defined) helpers override earlier ones.
	// Task 9.217: Use case-insensitive lookup for DWScript compatibility
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
		if method := findMethodCaseInsensitive(helper.Methods, methodName); method != nil {
			if spec, ok := findBuiltinMethodCaseInsensitive(helper.BuiltinMethods, methodName); ok {
				return helper, method, spec
			}
			return helper, method, ""
		}
	}

	// If no declared method, check for builtin-only entries
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
		if spec, ok := findBuiltinMethodCaseInsensitive(helper.BuiltinMethods, methodName); ok {
			return helper, nil, spec
		}
	}

	return nil, nil, ""
}

// findHelperProperty searches all applicable helpers for a property with the given name
func (i *Interpreter) findHelperProperty(val Value, propName string) (*HelperInfo, *types.PropertyInfo) {
	helpers := i.getHelpersForValue(val)
	if helpers == nil {
		return nil, nil
	}

	// Search helpers in reverse order so later helpers override earlier ones
	// Task 9.217: Use case-insensitive lookup for DWScript compatibility
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
		if prop := findPropertyCaseInsensitive(helper.Properties, propName); prop != nil {
			return helper, prop
		}
	}

	return nil, nil
}

// callHelperMethod executes a helper method (user-defined or built-in) on a value
// Task 9.86: Implement helper method dispatch
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

	// Bind helper class vars and consts
	for name, value := range helper.ClassVars {
		i.env.Define(name, value)
	}
	for name, value := range helper.ClassConsts {
		i.env.Define(name, value)
	}

	// Bind method parameters
	for idx, param := range method.Parameters {
		i.env.Define(param.Name.Value, args[idx])
	}

	// For functions, initialize the Result variable
	// Task 9.221: Use appropriate default value based on return type
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

// evalBuiltinHelperMethod executes a built-in helper method implementation identified by spec.
func (i *Interpreter) evalBuiltinHelperMethod(spec string, selfValue Value, args []Value, node ast.Node) Value {
	switch spec {
	case "__integer_tostring":
		if len(args) != 0 {
			return i.newErrorWithLocation(node, "Integer.ToString does not take arguments")
		}
		intVal, ok := selfValue.(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(node, "Integer.ToString requires integer receiver")
		}
		return &StringValue{Value: strconv.FormatInt(intVal.Value, 10)}

	case "__float_tostring_prec":
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "Float.ToString expects exactly 1 argument")
		}
		floatVal, ok := selfValue.(*FloatValue)
		if !ok {
			return i.newErrorWithLocation(node, "Float.ToString requires float receiver")
		}
		precVal, ok := args[0].(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(node, "Float.ToString precision must be Integer, got %s", args[0].Type())
		}
		precision := int(precVal.Value)
		if precision < 0 {
			precision = 0
		}
		return &StringValue{Value: fmt.Sprintf("%.*f", precision, floatVal.Value)}

	case "__boolean_tostring":
		if len(args) != 0 {
			return i.newErrorWithLocation(node, "Boolean.ToString does not take arguments")
		}
		boolVal, ok := selfValue.(*BooleanValue)
		if !ok {
			return i.newErrorWithLocation(node, "Boolean.ToString requires boolean receiver")
		}
		if boolVal.Value {
			return &StringValue{Value: "True"}
		}
		return &StringValue{Value: "False"}

	case "__string_array_join":
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "String array Join expects exactly 1 argument")
		}
		separator, ok := args[0].(*StringValue)
		if !ok {
			return i.newErrorWithLocation(node, "Join separator must be String, got %s", args[0].Type())
		}
		arrVal, ok := selfValue.(*ArrayValue)
		if !ok {
			return i.newErrorWithLocation(node, "Join helper requires string array receiver")
		}
		var builder strings.Builder
		for idx, elem := range arrVal.Elements {
			strElem, ok := elem.(*StringValue)
			if !ok {
				return i.newErrorWithLocation(node, "Join requires elements of type String")
			}
			if idx > 0 {
				builder.WriteString(separator.Value)
			}
			builder.WriteString(strElem.Value)
		}
		return &StringValue{Value: builder.String()}

	case "__array_add":
		// Implements arr.Add(value) - adds an element to a dynamic array
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "Array.Add expects exactly 1 argument")
		}
		arrVal, ok := selfValue.(*ArrayValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.Add requires array receiver")
		}

		// Check if it's a dynamic array (static arrays cannot use Add)
		if !arrVal.ArrayType.IsDynamic() {
			return i.newErrorWithLocation(node, "Add() can only be used with dynamic arrays, not static arrays")
		}

		// For dynamic arrays, just append the element
		// Type checking should have been done at semantic analysis
		valueToAdd := args[0]
		arrVal.Elements = append(arrVal.Elements, valueToAdd)

		// Return nil (procedure, not a function)
		return &NilValue{}

	case "__array_setlength":
		// Task 9.216: Implements arr.SetLength(newLength) - resizes a dynamic array
		if len(args) != 1 {
			return i.newErrorWithLocation(node, "Array.SetLength expects exactly 1 argument")
		}
		arrVal, ok := selfValue.(*ArrayValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.SetLength requires array receiver")
		}

		// Check if it's a dynamic array (static arrays cannot use SetLength)
		if !arrVal.ArrayType.IsDynamic() {
			return i.newErrorWithLocation(node, "SetLength() can only be used with dynamic arrays, not static arrays")
		}

		// Get the new length
		lengthInt, ok := args[0].(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(node, "Array.SetLength expects integer argument, got %s", args[0].Type())
		}

		newLength := int(lengthInt.Value)
		if newLength < 0 {
			return i.newErrorWithLocation(node, "Array.SetLength expects non-negative length, got %d", newLength)
		}

		currentLength := len(arrVal.Elements)

		if newLength == currentLength {
			// No change
			return &NilValue{}
		}

		if newLength < currentLength {
			// Truncate the slice
			arrVal.Elements = arrVal.Elements[:newLength]
			return &NilValue{}
		}

		// Extend the slice with default values
		elementType := arrVal.ArrayType.ElementType
		for j := currentLength; j < newLength; j++ {
			arrVal.Elements = append(arrVal.Elements, getZeroValueForType(elementType, nil))
		}

		// Return nil (procedure, not a function)
		return &NilValue{}

	default:
		return i.newErrorWithLocation(node, "unknown built-in helper method '%s'", spec)
	}
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
		if method, exists := helper.Methods[propInfo.ReadSpec]; exists {
			builtinSpec := helper.BuiltinMethods[propInfo.ReadSpec]
			// Call the getter method with no arguments
			return i.callHelperMethod(helper, method, builtinSpec, selfValue, []Value{}, node)
		}

		return i.newErrorWithLocation(node, "property '%s' read specifier '%s' not found",
			propInfo.Name, propInfo.ReadSpec)

	case types.PropAccessMethod:
		// Call getter method
		if method, exists := helper.Methods[propInfo.ReadSpec]; exists {
			builtinSpec := helper.BuiltinMethods[propInfo.ReadSpec]
			return i.callHelperMethod(helper, method, builtinSpec, selfValue, []Value{}, node)
		}

		return i.newErrorWithLocation(node, "property '%s' getter method '%s' not found",
			propInfo.Name, propInfo.ReadSpec)

	case types.PropAccessBuiltin:
		// Task 9.171: Built-in array helper properties
		return i.evalBuiltinHelperProperty(propInfo.ReadSpec, selfValue, node)

	case types.PropAccessNone:
		return i.newErrorWithLocation(node, "property '%s' is write-only", propInfo.Name)

	default:
		return i.newErrorWithLocation(node, "property '%s' has no read access", propInfo.Name)
	}
}

// resolveTypeFromExpression resolves a type from any TypeExpression.
// Task 9.170.1: Added to support inline array types in class fields.
func (i *Interpreter) resolveTypeFromExpression(typeExpr ast.TypeExpression) types.Type {
	if typeExpr == nil {
		return nil
	}

	// For simple type annotations, delegate to existing function
	if typeAnnot, ok := typeExpr.(*ast.TypeAnnotation); ok {
		return i.resolveTypeFromAnnotation(typeAnnot)
	}

	// For array types, resolve the element type and construct an array type
	if arrayType, ok := typeExpr.(*ast.ArrayTypeNode); ok {
		elementType := i.resolveTypeFromExpression(arrayType.ElementType)
		if elementType == nil {
			return nil
		}

		// Task 9.205: Evaluate bound expressions if this is a static array
		if arrayType.IsDynamic() {
			return types.NewDynamicArrayType(elementType)
		}

		// Evaluate low bound
		lowBoundVal := i.Eval(arrayType.LowBound)
		if isError(lowBoundVal) {
			return nil
		}
		lowBound, ok := lowBoundVal.(*IntegerValue)
		if !ok {
			return nil
		}

		// Evaluate high bound
		highBoundVal := i.Eval(arrayType.HighBound)
		if isError(highBoundVal) {
			return nil
		}
		highBound, ok := highBoundVal.(*IntegerValue)
		if !ok {
			return nil
		}

		return types.NewStaticArrayType(elementType, int(lowBound.Value), int(highBound.Value))
	}

	// For function pointer types, we need full type information
	// For now, return a generic function type placeholder
	if _, ok := typeExpr.(*ast.FunctionPointerTypeNode); ok {
		// TODO: Properly construct function pointer type
		return types.NewFunctionType([]types.Type{}, nil)
	}

	return nil
}

// resolveTypeFromAnnotation resolves a type from an AST TypeAnnotation
func (i *Interpreter) resolveTypeFromAnnotation(typeAnnot *ast.TypeAnnotation) types.Type {
	if typeAnnot == nil {
		return nil
	}

	typeName := typeAnnot.Name

	// Normalize type name to lowercase for case-insensitive comparison
	// DWScript (like Pascal) is case-insensitive for all identifiers including type names
	lowerTypeName := strings.ToLower(typeName)

	// Check basic types (case-insensitive)
	switch lowerTypeName {
	case "integer":
		return types.INTEGER
	case "float":
		return types.FLOAT
	case "string":
		return types.STRING
	case "boolean":
		return types.BOOLEAN
	case "const":
		// Task 9.235: Migrate Const to Variant for proper dynamic typing
		// "Const" was a temporary workaround, now redirects to VARIANT
		return types.VARIANT
	case "variant":
		// Task 9.227: Support Variant type for dynamic values
		return types.VARIANT
	}

	// Check for class types (stored in i.classes map)
	// Preserve original case for custom type lookup
	if classInfo, ok := i.classes[typeName]; ok {
		return types.NewClassType(classInfo.Name, nil)
	}

	// Check for record types (stored with special prefix in environment)
	recordTypeKey := "__record_type_" + typeName
	if typeVal, ok := i.env.Get(recordTypeKey); ok {
		if recordTypeVal, ok := typeVal.(*RecordTypeValue); ok {
			return recordTypeVal.RecordType
		}
	}

	// Type not found
	return nil
}

// extractSimpleTypeName extracts the simple type name from a full type string
// e.g., "INTEGER" -> "Integer"
func extractSimpleTypeName(typeName string) string {
	// Just return the type name as-is for now
	return typeName
}

// ============================================================================
// Built-in Array Helpers
// ============================================================================

// evalBuiltinHelperProperty evaluates a built-in helper property
// Task 9.171: Implements .Length, .High, .Low for arrays
func (i *Interpreter) evalBuiltinHelperProperty(propSpec string, selfValue Value, node ast.Node) Value {
	switch propSpec {
	case "__array_length", "__array_high", "__array_low":
		arrVal, ok := selfValue.(*ArrayValue)
		if !ok {
			return i.newErrorWithLocation(node, "built-in property '%s' can only be used on arrays", propSpec)
		}
		var result Value
		switch propSpec {
		case "__array_length":
			result = &IntegerValue{Value: int64(len(arrVal.Elements))}
		case "__array_high":
			if arrVal.ArrayType.IsStatic() {
				result = &IntegerValue{Value: int64(*arrVal.ArrayType.HighBound)}
			} else {
				result = &IntegerValue{Value: int64(len(arrVal.Elements) - 1)}
			}
		case "__array_low":
			if arrVal.ArrayType.IsStatic() {
				result = &IntegerValue{Value: int64(*arrVal.ArrayType.LowBound)}
			} else {
				result = &IntegerValue{Value: 0}
			}
		}
		return result

	case "__integer_tostring":
		intVal, ok := selfValue.(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(node, "Integer.ToString property requires integer receiver")
		}
		return &StringValue{Value: strconv.FormatInt(intVal.Value, 10)}

	case "__float_tostring_default":
		floatVal, ok := selfValue.(*FloatValue)
		if !ok {
			return i.newErrorWithLocation(node, "Float.ToString property requires float receiver")
		}
		return &StringValue{Value: fmt.Sprintf("%g", floatVal.Value)}

	case "__boolean_tostring":
		boolVal, ok := selfValue.(*BooleanValue)
		if !ok {
			return i.newErrorWithLocation(node, "Boolean.ToString property requires boolean receiver")
		}
		if boolVal.Value {
			return &StringValue{Value: "True"}
		}
		return &StringValue{Value: "False"}

	default:
		return i.newErrorWithLocation(node, "unknown built-in property '%s'", propSpec)
	}
}

// initArrayHelpers registers built-in helper properties for arrays
// Task 9.171: Array Helper Properties (.High, .Low, .Length)
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

	// Register .Add() method for dynamic arrays
	// This allows: arr.Add(value) syntax
	arrayHelper.BuiltinMethods["Add"] = "__array_add"

	// Task 9.216: Register .SetLength() method for dynamic arrays
	// This allows: arr.SetLength(newLength) syntax
	arrayHelper.BuiltinMethods["SetLength"] = "__array_setlength"

	// Register helper for ARRAY type
	i.helpers["ARRAY"] = append(i.helpers["ARRAY"], arrayHelper)
}

// initIntrinsicHelpers registers built-in helpers for primitive types (Integer, Float, Boolean).
func (i *Interpreter) initIntrinsicHelpers() {
	if i.helpers == nil {
		i.helpers = make(map[string][]*HelperInfo)
	}

	register := func(typeName string, helper *HelperInfo) {
		i.helpers[typeName] = append(i.helpers[typeName], helper)
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
	intHelper.Methods["ToString"] = nil
	intHelper.BuiltinMethods["ToString"] = "__integer_tostring"
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
	floatHelper.Methods["ToString"] = nil
	floatHelper.BuiltinMethods["ToString"] = "__float_tostring_prec"
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
	boolHelper.Methods["ToString"] = nil
	boolHelper.BuiltinMethods["ToString"] = "__boolean_tostring"
	register("Boolean", boolHelper)

	// String dynamic array helper
	stringArrayType := types.NewDynamicArrayType(types.STRING)
	stringArrayHelper := NewHelperInfo("__TStringDynArrayIntrinsicHelper", stringArrayType, true)
	stringArrayHelper.Methods["Join"] = nil
	stringArrayHelper.BuiltinMethods["Join"] = "__string_array_join"
	register(stringArrayType.String(), stringArrayHelper)
}
