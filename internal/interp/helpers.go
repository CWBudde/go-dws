package interp

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Helper Support (Task 9.86-9.89)
// ============================================================================

// HelperInfo stores runtime information about a helper type
type HelperInfo struct {
	Name           string                               // Helper type name
	TargetType     types.Type                           // The type being extended
	Methods        map[string]*ast.FunctionDecl         // Helper methods
	Properties     map[string]*types.PropertyInfo       // Helper properties
	ClassVars      map[string]Value                     // Class variable values
	ClassConsts    map[string]Value                     // Class constant values
	IsRecordHelper bool                                 // true if "record helper"
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
		IsRecordHelper: isRecordHelper,
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

	// Initialize class variables (Task 9.87)
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

	// Initialize class constants (Task 9.87)
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
		// Use generic "ARRAY" type name for all arrays
		typeName = "ARRAY"
	default:
		// For other types, try to extract type name from Type() method
		typeName = v.Type()
	}

	// Look up helpers for this type
	return i.helpers[typeName]
}

// findHelperMethod searches all applicable helpers for a method with the given name
func (i *Interpreter) findHelperMethod(val Value, methodName string) (*HelperInfo, *ast.FunctionDecl) {
	helpers := i.getHelpersForValue(val)
	if helpers == nil {
		return nil, nil
	}

	// Search all helpers for the method
	for _, helper := range helpers {
		if method, exists := helper.Methods[methodName]; exists {
			return helper, method
		}
	}

	return nil, nil
}

// findHelperProperty searches all applicable helpers for a property with the given name
func (i *Interpreter) findHelperProperty(val Value, propName string) (*HelperInfo, *types.PropertyInfo) {
	helpers := i.getHelpersForValue(val)
	if helpers == nil {
		return nil, nil
	}

	// Search all helpers for the property
	for _, helper := range helpers {
		if prop, exists := helper.Properties[propName]; exists {
			return helper, prop
		}
	}

	return nil, nil
}

// callHelperMethod executes a helper method on a value
// Task 9.86: Implement helper method dispatch
func (i *Interpreter) callHelperMethod(helper *HelperInfo, method *ast.FunctionDecl,
	selfValue Value, args []Value, node ast.Node) Value {

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

	// Bind helper class vars and consts (Task 9.87)
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
	if method.ReturnType != nil {
		i.env.Define("Result", &NilValue{})
		i.env.Define(method.Name.Value, &NilValue{})
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
		if method, exists := helper.Methods[propInfo.ReadSpec]; exists {
			// Call the getter method with no arguments
			return i.callHelperMethod(helper, method, selfValue, []Value{}, node)
		}

		return i.newErrorWithLocation(node, "property '%s' read specifier '%s' not found",
			propInfo.Name, propInfo.ReadSpec)

	case types.PropAccessMethod:
		// Call getter method
		if method, exists := helper.Methods[propInfo.ReadSpec]; exists {
			return i.callHelperMethod(helper, method, selfValue, []Value{}, node)
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
		return &types.ArrayType{
			ElementType: elementType,
			LowBound:    arrayType.LowBound,
			HighBound:   arrayType.HighBound,
		}
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

	// Check basic types
	switch typeName {
	case "Integer":
		return types.INTEGER
	case "Float":
		return types.FLOAT
	case "String":
		return types.STRING
	case "Boolean":
		return types.BOOLEAN
	}

	// Check for class types (stored in i.classes map)
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
// Built-in Array Helpers (Task 9.171)
// ============================================================================

// evalBuiltinHelperProperty evaluates a built-in helper property
// Task 9.171: Implements .Length, .High, .Low for arrays
func (i *Interpreter) evalBuiltinHelperProperty(propSpec string, selfValue Value, node ast.Node) Value {
	// Check if the value is an array
	arrVal, ok := selfValue.(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(node, "built-in property '%s' can only be used on arrays", propSpec)
	}

	switch propSpec {
	case "__array_length":
		// Task 9.171.4: .Length property
		return &IntegerValue{Value: int64(len(arrVal.Elements))}

	case "__array_high":
		// Task 9.171.2: .High property
		// For static arrays, return the HighBound
		if arrVal.ArrayType.IsStatic() {
			return &IntegerValue{Value: int64(*arrVal.ArrayType.HighBound)}
		}
		// For dynamic arrays, return len(arr) - 1
		return &IntegerValue{Value: int64(len(arrVal.Elements) - 1)}

	case "__array_low":
		// Task 9.171.3: .Low property
		// For static arrays, return the LowBound
		if arrVal.ArrayType.IsStatic() {
			return &IntegerValue{Value: int64(*arrVal.ArrayType.LowBound)}
		}
		// For dynamic arrays, always return 0
		return &IntegerValue{Value: 0}

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

	// Register helper for ARRAY type
	i.helpers["ARRAY"] = append(i.helpers["ARRAY"], arrayHelper)
}
