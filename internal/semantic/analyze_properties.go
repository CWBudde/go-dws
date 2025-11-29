package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	pkgident "github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Property Semantic Analysis
// ============================================================================

// pluralizeParam returns "parameter" or "parameters" based on count.
func pluralizeParam(count int) string {
	if count == 1 {
		return "parameter"
	}
	return "parameters"
}

// analyzePropertyDecl validates a property declaration and registers it in the class metadata.
func (a *Analyzer) analyzePropertyDecl(prop *ast.PropertyDecl, classType *types.ClassType) {
	propName := prop.Name.Value

	//  Check for duplicate property names within class
	for existingName := range classType.Properties {
		if pkgident.Equal(existingName, propName) {
			a.addError("duplicate property '%s' in class '%s' at %s",
				propName, classType.Name, prop.Token.Pos.String())
			return
		}
	}

	// Register properties with their declared casing
	if classType.Properties == nil {
		classType.Properties = make(map[string]*types.PropertyInfo)
	}
	if _, exists := classType.Properties[propName]; exists {
		a.addError("duplicate property '%s' in class '%s' at %s",
			propName, classType.Name, prop.Token.Pos.String())
		return
	}

	// Resolve property type
	if prop.Type == nil {
		a.addError("property '%s' missing type annotation in class '%s' at %s",
			propName, classType.Name, prop.Token.Pos.String())
		return
	}

	propType, err := a.resolveType(getTypeExpressionName(prop.Type))
	if err != nil {
		a.addError("unknown type '%s' for property '%s' in class '%s' at %s",
			getTypeExpressionName(prop.Type), propName, classType.Name, prop.Token.Pos.String())
		return
	}

	// Validate indexed property parameters have valid types
	isIndexed := prop.IndexParams != nil && len(prop.IndexParams) > 0
	var indexParamTypes []types.Type
	if isIndexed {
		for _, param := range prop.IndexParams {
			if param.Type == nil {
				a.addError("index parameter '%s' missing type annotation in property '%s' at %s",
					param.Name.Value, propName, prop.Token.Pos.String())
				return
			}
			paramType, err := a.resolveType(getTypeExpressionName(param.Type))
			if err != nil {
				a.addError("unknown type '%s' for index parameter '%s' in property '%s' at %s",
					getTypeExpressionName(param.Type), param.Name.Value, propName, prop.Token.Pos.String())
				return
			}
			indexParamTypes = append(indexParamTypes, paramType)
		}
	}

	// Create PropertyInfo to store in class metadata
	propInfo := &types.PropertyInfo{
		Name:            propName,
		Type:            propType,
		IsIndexed:       isIndexed,
		IsDefault:       prop.IsDefault,
		IsClassProperty: prop.IsClassProperty,
	}

	// Task 9.49: Register property stub before validating specs
	// This allows circular reference detection in property expressions
	classType.Properties[propName] = propInfo

	// Validate read specifier
	if prop.ReadSpec != nil {
		a.validateReadSpec(prop, classType, propInfo, indexParamTypes)
	}

	// Validate write specifier
	if prop.WriteSpec != nil {
		a.validateWriteSpec(prop, classType, propInfo, indexParamTypes)
	}

	// Validate default property restrictions
	if prop.IsDefault {
		// Default properties must be indexed
		if !isIndexed {
			a.addError("default property '%s' must be an indexed property at %s",
				propName, prop.Token.Pos.String())
			return
		}

		// Only one default property per class
		for existingPropName, existingProp := range classType.Properties {
			if existingProp.IsDefault && existingPropName != propName {
				a.addError("class '%s' already has default property '%s'; cannot declare another default property '%s' at %s",
					classType.Name, existingPropName, propName, prop.Token.Pos.String())
				return
			}
		}
	}

	// Property already registered at the beginning (line 68) for circular reference detection
}

// validateReadSpec validates the read specifier of a property.
// The read specifier can be:
//   - Field: A field name (identifier) - the field must exist and have matching type
//   - Constant: A constant name (identifier) - the constant must exist and have matching type
//   - Method: A method name (identifier) - the method must exist with correct signature
//   - Expression: An inline expression - the expression type must match property type
func (a *Analyzer) validateReadSpec(prop *ast.PropertyDecl, classType *types.ClassType, propInfo *types.PropertyInfo, indexParamTypes []types.Type) {
	propName := prop.Name.Value
	propType := propInfo.Type

	// Check if read spec is an identifier (field, constant, or method name)
	if ident, ok := prop.ReadSpec.(*ast.Identifier); ok {
		readSpecName := ident.Value

		// Task 9.17: Check class-level members first (class vars, then constants)
		// This ensures class-level storage takes precedence over instance fields

		// 1. Check if it's a class variable (only for class properties)
		if fieldType, found := classType.ClassVars[pkgident.Normalize(readSpecName)]; found {
			// Only class properties can read from class variables
			if propInfo.IsClassProperty {
				if !propType.Equals(fieldType) {
					a.addError("property '%s' read class variable '%s' has type %s, expected %s at %s",
						propName, readSpecName, fieldType.String(), propType.String(),
						prop.Token.Pos.String())
					return
				}
				propInfo.ReadKind = types.PropAccessField
				propInfo.ReadSpec = readSpecName
				return
			}
			// Instance property cannot read from class variable - skip this match
		}

		// 2. Check if it's a constant
		if constantType, constantFound := a.getConstantType(classType, readSpecName); constantFound {
			if !propType.Equals(constantType) {
				a.addError("property '%s' read constant '%s' has type %s, expected %s at %s",
					propName, readSpecName, constantType.String(), propType.String(),
					prop.Token.Pos.String())
				return
			}
			propInfo.ReadKind = types.PropAccessField // Constants are treated like fields
			propInfo.ReadSpec = readSpecName
			return
		}

		// 3. Check if it's an instance field (only for instance properties)
		if !propInfo.IsClassProperty {
			// Instance property can use instance field
			if fieldType, found := classType.GetField(pkgident.Normalize(readSpecName)); found {
				if !propType.Equals(fieldType) {
					a.addError("property '%s' read field '%s' has type %s, expected %s at %s",
						propName, readSpecName, fieldType.String(), propType.String(),
						prop.Token.Pos.String())
					return
				}
				propInfo.ReadKind = types.PropAccessField
				propInfo.ReadSpec = readSpecName
				return
			}
		}

		// If method, verify method exists with correct signature
		if methodType, found := classType.GetMethod(pkgident.Normalize(readSpecName)); found {
			// For class properties, verify the method is a class method
			// Task 9.16.1: Use lowercase key since ClassMethodFlags now uses lowercase keys
			if propInfo.IsClassProperty {
				isClassMethod := classType.ClassMethodFlags != nil && classType.ClassMethodFlags[pkgident.Normalize(readSpecName)]
				if !isClassMethod {
					a.addError("class property '%s' read method '%s' must be a class method at %s",
						propName, readSpecName, prop.Token.Pos.String())
					return
				}
			} else {
				// For instance properties, verify the method is NOT a class method
				// Task 9.16.1: Use lowercase key since ClassMethodFlags now uses lowercase keys
				isClassMethod := classType.ClassMethodFlags != nil && classType.ClassMethodFlags[pkgident.Normalize(readSpecName)]
				if isClassMethod {
					a.addError("instance property '%s' read method '%s' cannot be a class method at %s",
						propName, readSpecName, prop.Token.Pos.String())
					return
				}
			}

			// Getter signature: for indexed properties, method must accept index parameters
			// and return property type. For non-indexed, method must take no parameters
			// and return property type.

			expectedParamCount := len(indexParamTypes)
			if len(methodType.Parameters) != expectedParamCount {
				a.addError("property '%s' getter method '%s' has %d %s, expected %d %s at %s",
					propName, readSpecName, len(methodType.Parameters), pluralizeParam(len(methodType.Parameters)),
					expectedParamCount, pluralizeParam(expectedParamCount),
					prop.Token.Pos.String())
				return
			}

			// Verify getter signature includes index parameters
			for i, paramType := range indexParamTypes {
				if !methodType.Parameters[i].Equals(paramType) {
					a.addError("property '%s' getter method '%s' parameter %d has type %s, expected %s at %s",
						propName, readSpecName, i+1,
						methodType.Parameters[i].String(), paramType.String(),
						prop.Token.Pos.String())
					return
				}
			}

			// Verify return type matches property type
			if !methodType.ReturnType.Equals(propType) {
				a.addError("property '%s' getter method '%s' returns %s, expected %s at %s",
					propName, readSpecName,
					methodType.ReturnType.String(), propType.String(),
					prop.Token.Pos.String())
				return
			}

			propInfo.ReadKind = types.PropAccessMethod
			propInfo.ReadSpec = readSpecName
			return
		}

		// Neither field nor method found
		a.addError("property '%s' read specifier '%s' not found in class '%s' at %s",
			propName, readSpecName, classType.Name, prop.Token.Pos.String())
		return
	}

	// If expression, validate expression type matches property type
	// Set up class context for expression analysis to enable implicit self access
	savedClass := a.currentClass
	savedInClassMethod := a.inClassMethod
	savedInPropertyExpr := a.inPropertyExpr
	savedCurrentProperty := a.currentProperty

	a.currentClass = classType
	a.inClassMethod = false      // Property expressions are in instance context
	a.inPropertyExpr = true      // Flag to enable special property expression handling
	a.currentProperty = propName // Track current property for circular reference detection

	defer func() {
		a.currentClass = savedClass
		a.inClassMethod = savedInClassMethod
		a.inPropertyExpr = savedInPropertyExpr
		a.currentProperty = savedCurrentProperty
	}()

	// Analyze the expression with implicit self context
	exprType := a.analyzeExpression(prop.ReadSpec)
	if exprType == nil {
		// Error already reported by analyzeExpression
		return
	}

	// Validate expression type matches property type
	if !exprType.Equals(propType) {
		a.addError("property '%s' read expression has type %s, expected %s at %s",
			propName, exprType.String(), propType.String(),
			prop.Token.Pos.String())
		return
	}

	// Store the expression for runtime evaluation
	propInfo.ReadKind = types.PropAccessExpression
	propInfo.ReadSpec = prop.ReadSpec.String()
	propInfo.ReadExpr = prop.ReadSpec // Store AST node for interpreter
}

// validateWriteSpec validates the write specifier of a property.
// The write specifier can be:
//   - Field: A field name (identifier) - the field must exist and have matching type
//   - Method: A method name (identifier) - the method must exist with correct signature
func (a *Analyzer) validateWriteSpec(prop *ast.PropertyDecl, classType *types.ClassType, propInfo *types.PropertyInfo, indexParamTypes []types.Type) {
	propName := prop.Name.Value
	propType := propInfo.Type

	// Write spec must be an identifier (field or method name)
	ident, ok := prop.WriteSpec.(*ast.Identifier)
	if !ok {
		a.addError("property '%s' write specifier must be a field or method name at %s",
			propName, prop.Token.Pos.String())
		return
	}

	writeSpecName := ident.Value

	// Check if it's a field (instance or class field)
	// For class properties, look in ClassVars; for instance properties, look in Fields
	var fieldType types.Type
	var found bool

	if propInfo.IsClassProperty {
		// Class property must use class variable
		fieldType, found = classType.ClassVars[pkgident.Normalize(writeSpecName)]
		if found && !propType.Equals(fieldType) {
			a.addError("class property '%s' write field '%s' has type %s, expected %s at %s",
				propName, writeSpecName, fieldType.String(), propType.String(),
				prop.Token.Pos.String())
			return
		}
	} else {
		// Instance property can only use instance field
		fieldType, found = classType.GetField(pkgident.Normalize(writeSpecName))
		if found && !propType.Equals(fieldType) {
			a.addError("property '%s' write field '%s' has type %s, expected %s at %s",
				propName, writeSpecName, fieldType.String(), propType.String(),
				prop.Token.Pos.String())
			return
		}
	}

	if found {
		propInfo.WriteKind = types.PropAccessField
		propInfo.WriteSpec = writeSpecName
		return
	}

	// Task 9.17: Check if it's a constant (constants are read-only, so error if used as write spec)
	if _, constantFound := a.getConstantType(classType, writeSpecName); constantFound {
		a.addError("property '%s' write specifier '%s' is a constant and cannot be written to at %s",
			propName, writeSpecName, prop.Token.Pos.String())
		return
	}

	// If method, verify method exists with correct signature
	if methodType, found := classType.GetMethod(pkgident.Normalize(writeSpecName)); found {
		// For class properties, verify the method is a class method
		// Task 9.16.1: Use lowercase key since ClassMethodFlags now uses lowercase keys
		if propInfo.IsClassProperty {
			isClassMethod := classType.ClassMethodFlags != nil && classType.ClassMethodFlags[pkgident.Normalize(writeSpecName)]
			if !isClassMethod {
				a.addError("class property '%s' write method '%s' must be a class method at %s",
					propName, writeSpecName, prop.Token.Pos.String())
				return
			}
		} else {
			// For instance properties, verify the method is NOT a class method
			// Task 9.16.1: Use lowercase key since ClassMethodFlags now uses lowercase keys
			isClassMethod := classType.ClassMethodFlags != nil && classType.ClassMethodFlags[pkgident.Normalize(writeSpecName)]
			if isClassMethod {
				a.addError("instance property '%s' write method '%s' cannot be a class method at %s",
					propName, writeSpecName, prop.Token.Pos.String())
				return
			}
		}

		// Setter signature: for indexed properties, method must accept index parameters
		// plus the property value. For non-indexed, method must take only the value parameter.
		// Setter must return void.

		expectedParamCount := len(indexParamTypes) + 1 // index params + value param
		if len(methodType.Parameters) != expectedParamCount {
			a.addError("property '%s' setter method '%s' has %d %s, expected %d %s at %s",
				propName, writeSpecName, len(methodType.Parameters), pluralizeParam(len(methodType.Parameters)),
				expectedParamCount, pluralizeParam(expectedParamCount),
				prop.Token.Pos.String())
			return
		}

		// Verify setter signature includes index parameters
		for i, paramType := range indexParamTypes {
			if !methodType.Parameters[i].Equals(paramType) {
				a.addError("property '%s' setter method '%s' parameter %d has type %s, expected %s at %s",
					propName, writeSpecName, i+1,
					methodType.Parameters[i].String(), paramType.String(),
					prop.Token.Pos.String())
				return
			}
		}

		// Verify last parameter is the property value with matching type
		valueParamIndex := len(indexParamTypes)
		if !methodType.Parameters[valueParamIndex].Equals(propType) {
			a.addError("property '%s' setter method '%s' value parameter has type %s, expected %s at %s",
				propName, writeSpecName,
				methodType.Parameters[valueParamIndex].String(), propType.String(),
				prop.Token.Pos.String())
			return
		}

		// Verify return type is void
		if methodType.ReturnType != types.VOID {
			a.addError("property '%s' setter method '%s' must return void, not %s at %s",
				propName, writeSpecName, methodType.ReturnType.String(),
				prop.Token.Pos.String())
			return
		}

		propInfo.WriteKind = types.PropAccessMethod
		propInfo.WriteSpec = writeSpecName
		return
	}

	// Neither field nor method found
	a.addError("property '%s' write specifier '%s' not found in class '%s' at %s",
		propName, writeSpecName, classType.Name, prop.Token.Pos.String())
}

// getConstantType retrieves the type of a constant from a class.
// It searches the class and its ancestors for the constant and returns its type.
// Task 9.17: Helper for property expression validation
func (a *Analyzer) getConstantType(classType *types.ClassType, constantName string) (types.Type, bool) {
	if classType == nil {
		return nil, false
	}

	// Search for the constant in this class and ancestors
	current := classType
	for current != nil {
		// Case-insensitive constant lookup
		for existingName, constType := range current.ConstantTypes {
			if pkgident.Equal(existingName, constantName) {
				return constType, true
			}
		}
		current = current.Parent
	}

	return nil, false
}
