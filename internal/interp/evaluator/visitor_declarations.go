package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// This file contains visitor methods for declaration AST nodes.
// Phase 3.5.2: Visitor pattern implementation for declarations.
//
// Declarations define types, functions, classes, etc. and register them
// in the appropriate registries.

// VisitFunctionDecl evaluates a function declaration.
// Task 3.5.7: Migrated from Interpreter.evalFunctionDeclaration to Evaluator visitor.
//
// This method handles two cases:
//   - Global functions: Registered in TypeSystem.FunctionRegistry
//   - Method implementations (node.ClassName != nil): Delegated to adapter for ClassInfo handling
//
// The split design is intentional: method implementations require ClassInfo internals
// (VMT rebuild, descendant propagation) which belong in the Interpreter. Global functions
// can be fully handled via the TypeSystem's FunctionRegistry.
func (e *Evaluator) VisitFunctionDecl(node *ast.FunctionDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil function declaration")
	}

	// Method implementations (TClass.Method syntax) are delegated to adapter
	// They require ClassInfo internals: VMT rebuild, descendant propagation
	if node.ClassName != nil {
		return e.adapter.EvalMethodImplementation(node)
	}

	// Global function - register in FunctionRegistry via TypeSystem
	// RegisterFunctionOrReplace handles the interface/implementation pattern:
	// - Forward declarations (no body) are appended
	// - Implementations (with body) replace matching declarations by signature
	e.typeSystem.RegisterFunctionOrReplace(node.Name.Value, node)

	return &runtime.NilValue{}
}

// fullClassNameFromDecl returns the fully qualified class name for nested classes.
// For nested classes, it combines the enclosing class name with the class name (e.g., "Outer.Inner").
// For top-level classes, it returns just the class name.
func (e *Evaluator) fullClassNameFromDecl(cd *ast.ClassDecl) string {
	if cd.EnclosingClass != nil && cd.EnclosingClass.Value != "" {
		return cd.EnclosingClass.Value + "." + cd.Name.Value
	}
	return cd.Name.Value
}

// VisitClassDecl evaluates a class declaration.
// Task 3.5.8: Migrated from Interpreter.evalClassDeclaration to Evaluator visitor.
//
// This method handles all aspects of class declaration including:
// - Basic class creation and partial class merging
// - Inheritance resolution (explicit parent or implicit TObject)
// - Interface implementation
// - Constants, fields, and nested types
// - Methods, constructors, destructors, and properties
// - Operator registration
// - Virtual method table building
// - TypeSystem registration
func (e *Evaluator) VisitClassDecl(node *ast.ClassDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil class declaration")
	}

	className := e.fullClassNameFromDecl(node)

	// Phase 2.1: Handle partial class merging or create new class
	var classInfo interface{}
	existingClass := e.typeSystem.LookupClass(className)

	if existingClass != nil {
		ci, ok := e.adapter.CastToClassInfo(existingClass)
		if !ok {
			return e.newError(node, "internal error: invalid class type for '%s'", className)
		}

		existingPartial := e.adapter.IsClassPartial(ci)
		if existingPartial && node.IsPartial {
			// Merge into existing partial class
			classInfo = ci
		} else {
			return e.newError(node, "class '%s' already declared", className)
		}
	} else {
		// Create new class
		classInfo = e.adapter.NewClassInfoAdapter(className)
	}

	// Set class flags
	if node.IsPartial {
		e.adapter.SetClassPartial(classInfo, true)
	}
	if node.IsAbstract {
		e.adapter.SetClassAbstract(classInfo, true)
	}
	if node.IsExternal {
		e.adapter.SetClassExternal(classInfo, true, node.ExternalName)
	}

	// Phase 2.2: Setup temporary environment for nested class context
	tempEnv := ctx.Env().NewEnclosedEnvironment()
	e.adapter.DefineCurrentClassMarker(tempEnv, classInfo)
	savedEnv := ctx.Env()
	ctx.SetEnv(tempEnv)
	defer ctx.SetEnv(savedEnv)

	// Phase 3: Inheritance resolution
	// Handle inheritance if parent class is specified
	var parentClass interface{}
	if node.Parent != nil {
		// Explicit parent specified
		parentName := node.Parent.Value
		parentClass = e.typeSystem.LookupClass(parentName)
		if parentClass == nil {
			return e.newError(node, "parent class '%s' not found", parentName)
		}
	} else {
		// If no explicit parent, implicitly inherit from TObject
		// (unless this IS TObject or it's an external class)
		if !ident.Equal(className, "TObject") && !node.IsExternal {
			parentClass = e.typeSystem.LookupClass("TObject")
			if parentClass == nil {
				return e.newError(node, "implicit parent class 'TObject' not found")
			}
		}
	}

	// Set parent reference and inherit members (only if not already set for partial classes)
	if parentClass != nil && e.adapter.ClassHasNoParent(classInfo) {
		e.adapter.SetClassParent(classInfo, parentClass)
	}

	// Phase 4: Interface implementation
	// Process implemented interfaces - validate existence and store references
	for _, ifaceIdent := range node.Interfaces {
		ifaceName := ifaceIdent.Value
		// Look up interface in TypeSystem (case-insensitive)
		iface := e.typeSystem.LookupInterface(ifaceName)
		if iface == nil {
			return e.newError(node, "interface '%s' not found", ifaceName)
		}

		// Add interface to class's interface list (updates both ClassInfo and Metadata)
		e.adapter.AddInterfaceToClass(classInfo, iface, ifaceName)

		// Note: Method implementation validation is deferred until class methods
		// are fully processed. For now, we just register that the class claims to
		// implement the interface. Full validation happens in semantic analysis.
	}

	// TODO: Phase 5 - Constants, fields, nested types

	// Phase 6: Methods, properties, operators

	// Add own methods to ClassInfo (these override parent methods if same name)
	// Support method overloading by storing multiple methods per name
	for _, method := range node.Methods {
		if !e.adapter.AddClassMethod(classInfo, method, className) {
			return e.newError(method, "failed to add method '%s' to class '%s'", method.Name.Value, className)
		}
	}

	// Process explicit constructor if declared separately
	if node.Constructor != nil {
		if !e.adapter.AddClassMethod(classInfo, node.Constructor, className) {
			return e.newError(node.Constructor, "failed to add constructor to class '%s'", className)
		}
	}

	// Identify constructor (method named "Create")
	// Legacy behavior from old implementation
	if constructor, exists := e.adapter.LookupClassMethod(classInfo, "create", false); exists {
		e.adapter.SetClassConstructor(classInfo, constructor)
	}

	// Identify destructor (method named "Destroy")
	// Legacy behavior - destructor metadata is already set during AddClassMethod if marked as destructor
	if destructor, exists := e.adapter.LookupClassMethod(classInfo, "destroy", false); exists {
		e.adapter.SetClassDestructor(classInfo, destructor)
	}

	// Inherit destructor from parent if no local destructor declared
	e.adapter.InheritDestructorIfMissing(classInfo)

	// Synthesize implicit parameterless constructor if any constructor has 'overload'
	e.adapter.SynthesizeDefaultConstructor(classInfo)

	// Register properties
	// Properties are registered after fields and methods so they can reference them
	for _, propDecl := range node.Properties {
		if propDecl == nil {
			continue
		}
		if !e.adapter.AddClassProperty(classInfo, propDecl) {
			return e.newError(propDecl, "failed to add property '%s' to class '%s'", propDecl.Name.Value, className)
		}
	}

	// Copy parent properties (child inherits all parent properties)
	e.adapter.InheritParentProperties(classInfo)

	// Register class operators (Stage 8)
	for _, opDecl := range node.Operators {
		if opDecl == nil {
			continue
		}
		if errVal := e.adapter.RegisterClassOperator(classInfo, opDecl); isError(errVal) {
			return errVal
		}
	}

	// Phase 7: Virtual method table and TypeSystem registration

	// Build virtual method table after all methods and fields are processed
	// This implements proper virtual/override/reintroduce semantics
	e.adapter.BuildVirtualMethodTable(classInfo)

	// Register class in TypeSystem after VMT is built
	// Note: Legacy map registration already done early (for field initializers)
	parentName := ""
	if parentClass != nil {
		parentName = e.adapter.GetClassNameFromClassInfoInterface(parentClass)
	}
	e.adapter.RegisterClassInTypeSystem(classInfo, parentName)

	return &runtime.NilValue{}
}

// VisitInterfaceDecl evaluates an interface declaration.
// Task 3.5.9: Migrated from Interpreter.evalInterfaceDeclaration to Evaluator visitor.
//
// This method:
//  1. Creates InterfaceInfo from the AST declaration
//  2. Resolves parent interface via TypeSystem.LookupInterface()
//  3. Converts InterfaceMethodDecl nodes to FunctionDecl (interface methods have no body)
//  4. Converts PropertyDecl nodes using convertPropertyDecl()
//  5. Registers interface via TypeSystem.RegisterInterface()
func (e *Evaluator) VisitInterfaceDecl(node *ast.InterfaceDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil interface declaration")
	}

	// Create new InterfaceInfo
	interfaceInfo := e.adapter.NewInterfaceInfoAdapter(node.Name.Value)

	// Handle inheritance if parent interface is specified
	if node.Parent != nil {
		parentName := node.Parent.Value
		parentInterface := e.typeSystem.LookupInterface(parentName)

		if parentInterface == nil {
			return e.newError(node.Parent, "parent interface '%s' not found", parentName)
		}

		// Type assert from any to *InterfaceInfo
		parentInfo, ok := e.adapter.CastToInterfaceInfo(parentInterface)
		if !ok {
			return e.newError(node.Parent, "invalid parent interface '%s'", parentName)
		}

		// Set parent reference (hierarchy traversal happens in InterfaceInfo methods)
		e.adapter.SetInterfaceParent(interfaceInfo, parentInfo)

		// Check for circular inheritance
		if e.hasCircularInterfaceInheritance(interfaceInfo) {
			return e.newError(node.Parent,
				"circular inheritance detected in interface '%s'", node.Name.Value)
		}
	}

	// Add methods to InterfaceInfo
	// Convert InterfaceMethodDecl nodes to FunctionDecl nodes for consistency
	for _, methodDecl := range node.Methods {
		// Create a FunctionDecl from the InterfaceMethodDecl
		// Interface methods are declarations only (no body)
		funcDecl := &ast.FunctionDecl{
			BaseNode: ast.BaseNode{
				Token: methodDecl.Token,
			},
			Name:       methodDecl.Name,
			Parameters: methodDecl.Parameters,
			ReturnType: methodDecl.ReturnType,
			Body:       nil, // Interface methods have no body
		}

		// Use normalized key for case-insensitive method lookups
		e.adapter.AddInterfaceMethod(interfaceInfo, ident.Normalize(methodDecl.Name.Value), funcDecl)
	}

	// Register properties declared on the interface
	for _, propDecl := range node.Properties {
		if propDecl == nil {
			continue
		}
		propInfo, err := e.convertPropertyDecl(propDecl)
		if err != nil {
			return e.newError(propDecl, "%v", err)
		}
		if propInfo != nil {
			e.adapter.AddInterfaceProperty(interfaceInfo, ident.Normalize(propDecl.Name.Value), propInfo)
		}
	}

	// Register interface in TypeSystem
	e.typeSystem.RegisterInterface(e.adapter.GetInterfaceName(interfaceInfo), interfaceInfo)

	return &runtime.NilValue{}
}

// convertPropertyDecl converts an AST property declaration to a PropertyInfo struct.
// This extracts the property metadata for runtime property access handling.
// This method is used by interface, class, and record declaration evaluation.
//
// Note: We mark all identifiers as field access for now and will check at runtime
// whether they're actually fields or methods.
func (e *Evaluator) convertPropertyDecl(propDecl *ast.PropertyDecl) (*types.PropertyInfo, error) {
	// Resolve property type
	var propType types.Type
	switch propDecl.Type.String() {
	case "Integer":
		propType = types.INTEGER
	case "Float":
		propType = types.FLOAT
	case "String":
		propType = types.STRING
	case "Boolean":
		propType = types.BOOLEAN
	default:
		// Try to resolve known class types; fall back to NIL if unknown
		if classInfo := e.adapter.ResolveClassInfoByName(propDecl.Type.String()); classInfo != nil {
			propType = types.NewClassType(e.adapter.GetClassNameFromInfo(classInfo), nil)
		} else {
			propType = types.NIL
		}
	}

	propInfo := &types.PropertyInfo{
		Name:            propDecl.Name.Value,
		Type:            propType,
		IsIndexed:       len(propDecl.IndexParams) > 0,
		IsDefault:       propDecl.IsDefault,
		IsClassProperty: propDecl.IsClassProperty,
	}

	if propDecl.IndexValue != nil {
		if val, ok := ast.ExtractIntegerLiteral(propDecl.IndexValue); ok {
			propInfo.HasIndexValue = true
			propInfo.IndexValue = val
			propInfo.IndexValueType = types.INTEGER
		}
	}

	// Determine read access kind and spec
	if propDecl.ReadSpec != nil {
		if ident, ok := propDecl.ReadSpec.(*ast.Identifier); ok {
			// It's an identifier - store the name, we'll check if it's a field or method at access time
			propInfo.ReadSpec = ident.Value
			// Mark as field for now - evalPropertyRead will check both fields and methods
			propInfo.ReadKind = types.PropAccessField
		} else {
			// It's an expression
			propInfo.ReadKind = types.PropAccessExpression
			propInfo.ReadSpec = propDecl.ReadSpec.String()
			propInfo.ReadExpr = propDecl.ReadSpec // Store AST node for evaluation
		}
	} else {
		propInfo.ReadKind = types.PropAccessNone
	}

	// Determine write access kind and spec
	if propDecl.WriteSpec != nil {
		if ident, ok := propDecl.WriteSpec.(*ast.Identifier); ok {
			// It's an identifier - store the name, we'll check if it's a field or method at access time
			propInfo.WriteSpec = ident.Value
			// Mark as field for now - evalPropertyWrite will check both fields and methods
			propInfo.WriteKind = types.PropAccessField
		} else {
			propInfo.WriteKind = types.PropAccessNone
		}
	} else {
		propInfo.WriteKind = types.PropAccessNone
	}

	return propInfo, nil
}

// hasCircularInterfaceInheritance checks if an interface has circular inheritance.
// This prevents infinite loops when traversing the interface hierarchy.
func (e *Evaluator) hasCircularInterfaceInheritance(iface any) bool {
	seen := make(map[string]bool)
	current := iface

	for current != nil {
		normalizedName := ident.Normalize(e.adapter.GetInterfaceName(current))
		if seen[normalizedName] {
			return true
		}
		seen[normalizedName] = true
		current = e.adapter.GetInterfaceParent(current)
	}

	return false
}

// VisitOperatorDecl evaluates an operator declaration (operator overloading).
// Task 3.5.14: Migrated from Interpreter.evalOperatorDeclaration to Evaluator visitor.
//
// This method handles global and conversion operator declarations:
//   - Global operators: registered via TypeSystem.Operators()
//   - Conversion operators: registered via TypeSystem.Conversions()
//
// Class operators (Kind == OperatorKindClass) are handled separately during
// class declaration evaluation, so this method returns NilValue for them.
func (e *Evaluator) VisitOperatorDecl(node *ast.OperatorDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil operator declaration")
	}

	// Class operators are registered during class declaration evaluation
	if node.Kind == ast.OperatorKindClass {
		return &runtime.NilValue{}
	}

	// Validate binding exists
	if node.Binding == nil {
		return e.newError(node, "operator '%s' missing binding", node.OperatorSymbol)
	}

	// Normalize operand types for consistent lookup
	operandTypes := make([]string, len(node.OperandTypes))
	for idx, operand := range node.OperandTypes {
		opRand := operand.String()
		operandTypes[idx] = interptypes.NormalizeTypeAnnotation(opRand)
	}

	// Handle conversion operators
	if node.Kind == ast.OperatorKindConversion {
		if len(operandTypes) != 1 {
			return e.newError(node, "conversion operator '%s' requires exactly one operand", node.OperatorSymbol)
		}
		if node.ReturnType == nil {
			return e.newError(node, "conversion operator '%s' requires a return type", node.OperatorSymbol)
		}

		targetType := interptypes.NormalizeTypeAnnotation(node.ReturnType.String())
		entry := &interptypes.ConversionEntry{
			From:        operandTypes[0],
			To:          targetType,
			BindingName: ident.Normalize(node.Binding.Value),
			Implicit:    ident.Equal(node.OperatorSymbol, "implicit"),
		}

		if err := e.typeSystem.Conversions().Register(entry); err != nil {
			return e.newError(node, "conversion from %s to %s already defined", operandTypes[0], targetType)
		}
		return &runtime.NilValue{}
	}

	// Handle global operators
	entry := &interptypes.OperatorEntry{
		Operator:     node.OperatorSymbol,
		OperandTypes: operandTypes,
		BindingName:  ident.Normalize(node.Binding.Value),
	}

	if err := e.typeSystem.Operators().Register(entry); err != nil {
		return e.newError(node, "operator '%s' already defined for operand types (%s)", node.OperatorSymbol, strings.Join(operandTypes, ", "))
	}

	return &runtime.NilValue{}
}

// VisitEnumDecl evaluates an enum declaration.
// Task 3.5.11: Migrated from Interpreter.evalEnumDeclaration to Evaluator visitor.
//
// This method:
//  1. Builds the enum type from the AST declaration
//  2. Calculates ordinal values (explicit or implicit)
//  3. For flags enums, validates values are powers of 2
//  4. For unscoped enums, registers values in the environment
//  5. Registers enum type metadata in the TypeSystem
//  6. Creates a TypeMetaValue for the enum type name
func (e *Evaluator) VisitEnumDecl(node *ast.EnumDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil enum declaration")
	}

	enumName := node.Name.Value

	// Build the enum type from the declaration
	enumValues := make(map[string]int)
	orderedNames := make([]string, 0, len(node.Values))

	// Calculate ordinal values (explicit or implicit)
	currentOrdinal := 0
	flagBitPosition := 0 // For flags enums, track the bit position (2^n)

	for _, enumValue := range node.Values {
		valueName := enumValue.Name

		// Determine ordinal value (explicit or implicit)
		var ordinalValue int
		if enumValue.Value != nil {
			// Explicit value provided
			ordinalValue = *enumValue.Value
			if node.Flags {
				// For flags, explicit values must be powers of 2
				if ordinalValue <= 0 || (ordinalValue&(ordinalValue-1)) != 0 {
					return e.newError(node, "enum '%s' value '%s' (%d) must be a power of 2 for flags enum",
						enumName, valueName, ordinalValue)
				}
				// For flags, update bit position based on explicit value
				for bitPos := 0; bitPos < 64; bitPos++ {
					if (1 << bitPos) == ordinalValue {
						flagBitPosition = bitPos + 1
						break
					}
				}
			} else {
				// For regular enums, update current ordinal
				currentOrdinal = ordinalValue + 1
			}
		} else {
			// Implicit value
			if node.Flags {
				// Flags use power-of-2 values: 1, 2, 4, 8, 16, ...
				ordinalValue = 1 << flagBitPosition
				flagBitPosition++
			} else {
				// Regular enums use sequential values
				ordinalValue = currentOrdinal
				currentOrdinal++
			}
		}

		// Store the value
		enumValues[valueName] = ordinalValue
		orderedNames = append(orderedNames, valueName)
	}

	// Create the enum type
	var enumType *types.EnumType
	if node.Scoped || node.Flags {
		enumType = types.NewScopedEnumType(enumName, enumValues, orderedNames, node.Flags)
	} else {
		enumType = types.NewEnumType(enumName, enumValues, orderedNames)
	}

	// Register each enum value in the symbol table as a constant
	// For scoped enums (enum/flags keyword), skip global registration -
	// values are only accessible via qualified access (Type.Value)
	if !node.Scoped {
		for valueName, ordinalValue := range enumValues {
			enumVal := &runtime.EnumValue{
				TypeName:     enumName,
				ValueName:    valueName,
				OrdinalValue: ordinalValue,
			}
			ctx.Env().Define(valueName, enumVal)
		}
	}

	// Store enum type metadata in environment with special key
	// This allows variable declarations to resolve the type
	enumTypeKey := "__enum_type_" + ident.Normalize(enumName)
	enumTypeValue := runtime.NewEnumTypeValue(enumType)
	ctx.Env().Define(enumTypeKey, enumTypeValue)

	// Register in TypeSystem for consistent type lookups
	e.typeSystem.RegisterEnumType(enumName, enumTypeValue)

	// Register enum type name as a TypeMetaValue
	// This allows the type name to be used as a runtime value in expressions
	// like High(TColor) or Low(TColor), just like built-in types (Integer, Float, etc.)
	typeMetaValue := &runtime.TypeMetaValue{
		TypeInfo: enumType,
		TypeName: enumName,
	}
	ctx.Env().Define(enumName, typeMetaValue)

	return &runtime.NilValue{}
}

// VisitRecordDecl evaluates a record declaration.
func (e *Evaluator) VisitRecordDecl(node *ast.RecordDecl, ctx *ExecutionContext) Value {
	// Task 3.5.10: Full implementation moved from Interpreter to Evaluator.
	// Eliminates adapter.EvalNode() call - direct registration using evaluator methods.
	if node == nil {
		return e.newError(nil, "nil record declaration")
	}

	recordName := node.Name.Value

	// Build the record type from the declaration
	fields := make(map[string]types.Type)
	// Task 9.5: Store field declarations for initializer access
	fieldDecls := make(map[string]*ast.FieldDecl)

	for _, field := range node.Fields {
		fieldName := field.Name.Value

		// Task 9.12.1: Handle type inference for fields
		var fieldType types.Type
		if field.Type != nil {
			// Explicit type - use evaluator's type resolution with context for environment access
			var err error
			typeName := field.Type.String()
			fieldType, err = e.resolveTypeName(typeName, ctx)
			if err != nil || fieldType == nil {
				return e.newError(node, "unknown or invalid type for field '%s' in record '%s'", fieldName, recordName)
			}
		} else if field.InitValue != nil {
			// Type inference from initializer - use evaluator's Eval
			initValue := e.Eval(field.InitValue, ctx)
			if isError(initValue) {
				return initValue
			}
			fieldType = e.getValueType(initValue)
			if fieldType == nil {
				return e.newError(node, "cannot infer type for field '%s' in record '%s'", fieldName, recordName)
			}
		} else {
			return e.newError(node, "field '%s' in record '%s' must have either a type or initializer", fieldName, recordName)
		}

		// Keep original casing for field map (NewRecordType normalizes keys)
		fields[fieldName] = fieldType
		// Task 9.5: Store field declaration (use lowercase key)
		fieldDecls[ident.Normalize(fieldName)] = field
	}

	// Create the record type
	recordType := types.NewRecordType(recordName, fields)

	// Task 9.7: Store method AST nodes for runtime invocation
	// Build maps for instance methods and static methods (class function/procedure)
	// Task 9.7f: Separate static methods from instance methods
	// Note: Use lowercase keys for case-insensitive lookup
	methods := make(map[string]*ast.FunctionDecl)
	staticMethods := make(map[string]*ast.FunctionDecl)
	for _, method := range node.Methods {
		methodKey := ident.Normalize(method.Name.Value)
		if method.IsClassMethod {
			staticMethods[methodKey] = method
		} else {
			methods[methodKey] = method
		}
	}

	// Task 9.12.2: Evaluate record constants - use evaluator's Eval
	constants := make(map[string]Value)
	for _, constant := range node.Constants {
		constName := constant.Name.Value
		constValue := e.Eval(constant.Value, ctx)
		if isError(constValue) {
			return constValue
		}
		// Normalize to lowercase for case-insensitive access
		constants[ident.Normalize(constName)] = constValue
	}

	// Task 9.12.2: Initialize class variables - use evaluator's Eval and GetDefaultValue
	classVars := make(map[string]Value)
	for _, classVar := range node.ClassVars {
		varName := classVar.Name.Value
		var varValue Value

		if classVar.InitValue != nil {
			// Evaluate the initializer
			varValue = e.Eval(classVar.InitValue, ctx)
			if isError(varValue) {
				return varValue
			}
		} else {
			// Use type to determine zero value
			var varType types.Type
			if classVar.Type != nil {
				var err error
				typeName := classVar.Type.String()
				varType, err = e.resolveTypeName(typeName, ctx)
				if err != nil || varType == nil {
					return e.newError(node, "unknown type for class variable '%s' in record '%s'", varName, recordName)
				}
			}
			varValue = e.GetDefaultValue(varType)
		}

		// Normalize to lowercase for case-insensitive access
		classVars[ident.Normalize(varName)] = varValue
	}

	// Process properties
	for _, prop := range node.Properties {
		propName := prop.Name.Value
		propNameLower := ident.Normalize(propName)

		// Resolve property type - use evaluator's type resolution with context
		typeName := prop.Type.String()
		propType, err := e.resolveTypeName(typeName, ctx)
		if err != nil || propType == nil {
			return e.newError(node, "unknown type for property '%s' in record '%s'", propName, recordName)
		}

		// Create property info
		propInfo := &types.RecordPropertyInfo{
			Name:       propName,
			Type:       propType,
			ReadField:  prop.ReadField,
			WriteField: prop.WriteField,
			IsDefault:  prop.IsDefault,
		}

		// Store in recordType.Properties (case-insensitive)
		recordType.Properties[propNameLower] = propInfo
	}

	// Task 3.5.42: Build RecordMetadata from AST declarations
	// Task 3.5.10: Use evaluator's buildRecordMetadata (no adapter)
	metadata := e.buildRecordMetadata(recordName, recordType, methods, staticMethods, constants, classVars)

	// Store record type metadata in environment with special key
	// This allows variable declarations to resolve the type
	recordTypeKey := "__record_type_" + ident.Normalize(recordName)
	recordTypeValue := &RecordTypeValue{
		RecordType:           recordType,
		FieldDecls:           fieldDecls, // Task 9.5: Include field declarations
		Metadata:             metadata,   // Task 3.5.42: AST-free metadata
		Methods:              methods,
		StaticMethods:        staticMethods,
		ClassMethods:         make(map[string]*ast.FunctionDecl),
		ClassMethodOverloads: make(map[string][]*ast.FunctionDecl),
		MethodOverloads:      make(map[string][]*ast.FunctionDecl),
		Constants:            constants, // Task 9.12.2: Record constants
		ClassVars:            classVars, // Task 9.12.2: Class variables
	}

	// Initialize ClassMethods with StaticMethods for compatibility
	for k, v := range staticMethods {
		recordTypeValue.ClassMethods[k] = v
	}

	// Initialize overload lists from method declarations
	// Note: methodName is already lowercase from the maps above
	for methodName, methodDecl := range methods {
		recordTypeValue.MethodOverloads[methodName] = []*ast.FunctionDecl{methodDecl}
	}
	for methodName, methodDecl := range staticMethods {
		recordTypeValue.ClassMethodOverloads[methodName] = []*ast.FunctionDecl{methodDecl}
	}

	// Task 3.5.10: Direct environment and TypeSystem access (NO ADAPTER)
	ctx.Env().Define(recordTypeKey, recordTypeValue)
	e.typeSystem.RegisterRecord(recordName, recordTypeValue)

	return &runtime.NilValue{}
}

// VisitHelperDecl evaluates a helper declaration (type extension).
// Task 3.5.12: Migrated from Interpreter.evalHelperDeclaration to Evaluator visitor.
//
// This method handles helper declarations of the form:
//
//	type TStringHelper = helper for String ... end;
//	type TPointHelper = record helper for TPoint ... end;
//	type TChildHelper = helper(TParentHelper) for String ... end;
//
// It resolves the target type, handles parent helper inheritance,
// registers methods/properties, and registers the helper in the TypeSystem.
func (e *Evaluator) VisitHelperDecl(node *ast.HelperDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return &runtime.NilValue{}
	}

	// 3.5.12.2: Resolve the target type using evaluator's type resolution
	targetType, err := e.ResolveTypeFromAnnotation(node.ForType)
	if err != nil {
		return e.newError(node, "unknown target type '%s' for helper '%s'",
			node.ForType.String(), node.Name.Value)
	}

	// Create helper info via adapter
	helperInfo := e.adapter.CreateHelperInfo(node.Name.Value, targetType, node.IsRecordHelper)
	if helperInfo == nil {
		return e.newError(node, "failed to create helper info for '%s'", node.Name.Value)
	}

	// 3.5.12.3: Resolve parent helper if specified
	if node.ParentHelper != nil {
		parentHelperName := node.ParentHelper.Value

		// Search all registered helpers for parent by name
		var foundParent interface{}
		for _, helpers := range e.typeSystem.AllHelpers() {
			for _, helper := range helpers {
				if ident.Equal(e.adapter.GetHelperName(helper), parentHelperName) {
					foundParent = helper
					break
				}
			}
			if foundParent != nil {
				break
			}
		}

		if foundParent == nil {
			return e.newError(node.ParentHelper,
				"unknown parent helper '%s' for helper '%s'",
				parentHelperName, node.Name.Value)
		}

		// Verify target type compatibility
		if !e.adapter.VerifyHelperTargetTypeMatch(foundParent, targetType) {
			return e.newError(node.ParentHelper,
				"parent helper '%s' extends different type than child helper '%s'",
				parentHelperName, node.Name.Value)
		}

		// Set up inheritance chain
		e.adapter.SetHelperParent(helperInfo, foundParent)
	}

	// Register methods (case-insensitive keys)
	for _, method := range node.Methods {
		e.adapter.AddHelperMethod(helperInfo, ident.Normalize(method.Name.Value), method)
	}

	// Register properties
	for _, prop := range node.Properties {
		propType, propErr := e.ResolveTypeFromAnnotation(prop.Type)
		if propErr != nil {
			return e.newError(prop, "unknown type '%s' for property '%s'",
				prop.Type.String(), prop.Name.Value)
		}
		e.adapter.AddHelperProperty(helperInfo, prop, propType)
	}

	// Initialize class variables
	for _, classVar := range node.ClassVars {
		var (
			varType      types.Type
			initialValue runtime.Value
		)

		if classVar.Type != nil {
			resolvedType, err := e.resolveTypeName(classVar.Type.String(), ctx)
			if err != nil {
				return e.newError(classVar, "unknown type for class variable '%s'",
					classVar.Name.Value)
			}
			varType = resolvedType
		}

		if classVar.InitValue != nil {
			val := e.Eval(classVar.InitValue, ctx)
			if isError(val) {
				return val
			}
			initialValue = val

			if varType == nil {
				switch v := val.(type) {
				case *runtime.IntegerValue:
					varType = types.INTEGER
				case *runtime.FloatValue:
					varType = types.FLOAT
				case *runtime.StringValue:
					varType = types.STRING
				case *runtime.BooleanValue:
					varType = types.BOOLEAN
				case *runtime.ArrayValue:
					varType = v.ArrayType
				}
				if varType == nil {
					return e.newError(classVar, "cannot infer type for class variable '%s'",
						classVar.Name.Value)
				}
			}
		}

		if varType == nil {
			return e.newError(classVar, "unknown type for class variable '%s'",
				classVar.Name.Value)
		}

		if initialValue == nil {
			initialValue = e.GetDefaultValue(varType)
		}

		e.adapter.AddHelperClassVar(helperInfo, classVar.Name.Value, initialValue)
	}

	// Initialize class constants (evaluate values)
	for _, classConst := range node.ClassConsts {
		constValue := e.Eval(classConst.Value, ctx)
		if isError(constValue) {
			return constValue
		}
		e.adapter.AddHelperClassConst(helperInfo, classConst.Name.Value, constValue)
	}

	// 3.5.12.4: Register via TypeSystem.RegisterHelper()
	typeName := ident.Normalize(targetType.String())
	e.typeSystem.RegisterHelper(typeName, helperInfo)

	// Also maintain legacy map for backward compatibility during migration
	e.adapter.RegisterHelperLegacy(typeName, helperInfo)

	// Also register by simple type name for lookup compatibility
	simpleTypeName := ident.Normalize(extractSimpleTypeName(targetType.String()))
	if simpleTypeName != typeName {
		e.typeSystem.RegisterHelper(simpleTypeName, helperInfo)
		e.adapter.RegisterHelperLegacy(simpleTypeName, helperInfo)
	}

	// Expose helper name as a type meta value for static access (e.g., THelper.Member)
	ctx.Env().Define(node.Name.Value, &runtime.TypeMetaValue{
		TypeInfo: targetType,
		TypeName: targetType.String(),
	})

	return &runtime.NilValue{}
}

// VisitArrayDecl evaluates an array type declaration.
// Task 3.5.13: Migrated from Interpreter.evalArrayDeclaration to Evaluator visitor.
//
// This method handles array type declarations of the form:
//
//	type TMyArray = array[1..10] of Integer;  // static array
//	type TDynArray = array of String;         // dynamic array
//
// It resolves the element type, evaluates bounds for static arrays,
// creates the appropriate ArrayType, and registers it in the TypeSystem.
func (e *Evaluator) VisitArrayDecl(node *ast.ArrayDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil array declaration")
	}

	arrayName := node.Name.Value

	// Build the array type from the declaration
	arrayTypeAnnotation := node.ArrayType
	if arrayTypeAnnotation == nil {
		return e.newError(node, "invalid array type declaration")
	}

	// Resolve the element type using evaluator's type resolution
	elementTypeName := arrayTypeAnnotation.ElementType.String()
	elementType, err := e.resolveTypeName(elementTypeName, ctx)
	if err != nil {
		return e.newError(node, "unknown element type '%s': %v", elementTypeName, err)
	}

	// Create the array type
	var arrayType *types.ArrayType
	if arrayTypeAnnotation.IsDynamic() {
		arrayType = types.NewDynamicArrayType(elementType)
	} else {
		// Evaluate bound expressions using evaluator
		lowBoundVal := e.Eval(arrayTypeAnnotation.LowBound, ctx)
		if isError(lowBoundVal) {
			return lowBoundVal
		}
		highBoundVal := e.Eval(arrayTypeAnnotation.HighBound, ctx)
		if isError(highBoundVal) {
			return highBoundVal
		}

		// Extract integer values
		lowBound, ok := lowBoundVal.(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "array lower bound must be an integer")
		}
		highBound, ok := highBoundVal.(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "array upper bound must be an integer")
		}

		arrayType = types.NewStaticArrayType(elementType, int(lowBound.Value), int(highBound.Value))
	}

	// Register array type in TypeSystem
	e.typeSystem.RegisterArrayType(arrayName, arrayType)

	return &runtime.NilValue{}
}

// VisitTypeDeclaration evaluates a type declaration.
// Task 3.5.16: Migrated from Interpreter.evalTypeDeclaration to Evaluator visitor.
//
// This method handles:
//  1. Subrange types: type TDigit = 0..9;
//  2. Function pointer types: type TCallback = procedure(x: Integer);
//  3. Type aliases: type TUserID = Integer;
//
// Inline/complex type expressions (ClassOfTypeNode, SetTypeNode, ArrayTypeNode,
// FunctionPointerTypeNode) are handled by the semantic analyzer and return NilValue.
func (e *Evaluator) VisitTypeDeclaration(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil type declaration")
	}

	// Handle subrange types
	if node.IsSubrange {
		return e.evalSubrangeType(node, ctx)
	}

	// Handle function pointer type declarations
	if node.IsFunctionPointer {
		return e.evalFunctionPointerType(node, ctx)
	}

	// Handle type aliases
	if node.IsAlias {
		return e.evalTypeAlias(node, ctx)
	}

	// Non-alias type declarations (future)
	return e.newError(node, "non-alias type declarations not yet supported")
}

// evalSubrangeType evaluates a subrange type declaration.
// Example: type TDigit = 0..9;
func (e *Evaluator) evalSubrangeType(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	// Evaluate low bound
	lowBoundVal := e.Eval(node.LowBound, ctx)
	if isError(lowBoundVal) {
		return lowBoundVal
	}
	lowBoundIntVal, ok := lowBoundVal.(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "subrange low bound must be an integer")
	}
	lowBoundInt := int(lowBoundIntVal.Value)

	// Evaluate high bound
	highBoundVal := e.Eval(node.HighBound, ctx)
	if isError(highBoundVal) {
		return highBoundVal
	}
	highBoundIntVal, ok := highBoundVal.(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "subrange high bound must be an integer")
	}
	highBoundInt := int(highBoundIntVal.Value)

	// Validate bounds
	if lowBoundInt > highBoundInt {
		return e.newError(node, "subrange low bound (%d) cannot be greater than high bound (%d)", lowBoundInt, highBoundInt)
	}

	// Create SubrangeType
	subrangeType := &types.SubrangeType{
		BaseType:  types.INTEGER,
		Name:      node.Name.Value,
		LowBound:  lowBoundInt,
		HighBound: highBoundInt,
	}

	// Register in TypeSystem
	e.typeSystem.RegisterSubrangeType(node.Name.Value, subrangeType)

	return &runtime.NilValue{}
}

// evalFunctionPointerType evaluates a function pointer type declaration.
// Example: type TCallback = procedure(x: Integer);
func (e *Evaluator) evalFunctionPointerType(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	if node.FunctionPointerType == nil {
		return e.newError(node, "function pointer type declaration has no type information")
	}

	funcPtrType := node.FunctionPointerType

	// Resolve parameter types
	paramTypes := make([]types.Type, len(funcPtrType.Parameters))
	for idx, param := range funcPtrType.Parameters {
		var paramType types.Type
		if param.Type != nil {
			var err error
			paramType, err = e.ResolveTypeFromAnnotation(param.Type)
			if err != nil || paramType == nil {
				return e.newError(node, "unknown parameter type '%s' in function pointer '%s'", param.Type.String(), node.Name.Value)
			}
		} else {
			paramType = types.INTEGER
		}
		paramTypes[idx] = paramType
	}

	// Resolve return type (nil for procedures)
	var returnType types.Type
	if funcPtrType.ReturnType != nil {
		var err error
		returnType, err = e.ResolveTypeFromAnnotation(funcPtrType.ReturnType)
		if err != nil {
			return e.newError(node, "unknown return type '%s' in function pointer '%s'", funcPtrType.ReturnType.String(), node.Name.Value)
		}
	}

	// Create function or method pointer type
	var resolvedType types.Type
	if funcPtrType.OfObject {
		resolvedType = types.NewMethodPointerType(paramTypes, returnType)
	} else {
		resolvedType = types.NewFunctionPointerType(paramTypes, returnType)
	}

	// Register in the type system for runtime resolution
	if e.typeSystem != nil {
		e.typeSystem.RegisterFunctionPointerType(node.Name.Value, resolvedType)
	}

	// Legacy marker for compatibility
	typeKey := "__funcptr_type_" + node.Name.Value
	ctx.Env().Define(typeKey, &runtime.StringValue{Value: "function_pointer_type"})

	return &runtime.NilValue{}
}

// evalTypeAlias evaluates a type alias declaration.
// Example: type TUserID = Integer;
func (e *Evaluator) evalTypeAlias(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	// Check for inline/complex type expressions
	// For inline types, we don't need to resolve them at runtime because:
	// 1. The semantic analyzer already validated the types during analysis phase
	// 2. Inline type aliases are purely semantic constructs (no runtime storage needed)
	// 3. The interpreter's resolveType() doesn't support complex inline syntax
	switch node.AliasedType.(type) {
	case *ast.ClassOfTypeNode:
		// Metaclass types (class of TBase) - semantic analyzer handles them
		return &runtime.NilValue{}
	case *ast.SetTypeNode:
		// Set types (set of TEnum) - semantic analyzer handles them
		return &runtime.NilValue{}
	case *ast.ArrayTypeNode:
		// Inline array types (array of Integer, array[1..10] of String)
		// Note: These could potentially need runtime storage, but semantic analyzer
		// already validated and stored them.
		return &runtime.NilValue{}
	case *ast.FunctionPointerTypeNode:
		// Function pointer types - already handled earlier in this function
		return &runtime.NilValue{}
	}

	// For TypeAnnotation with InlineType, also skip runtime resolution
	if typeAnnot, ok := node.AliasedType.(*ast.TypeAnnotation); ok && typeAnnot.InlineType != nil {
		// TypeAnnotation wrapping an inline type expression
		return &runtime.NilValue{}
	}

	// Resolve the aliased type by name (handles simple named types only)
	aliasedType, err := e.resolveTypeName(node.AliasedType.String(), ctx)
	if err != nil {
		return e.newError(node, "unknown type '%s' in type alias", node.AliasedType.String())
	}

	// Create TypeAliasValue and register it
	typeAlias := &runtime.TypeAliasValue{
		Name:        node.Name.Value,
		AliasedType: aliasedType,
	}

	// Store in environment with special prefix (case-insensitive)
	typeKey := "__type_alias_" + ident.Normalize(node.Name.Value)
	ctx.Env().Define(typeKey, typeAlias)

	return &runtime.NilValue{}
}

// VisitSetDecl evaluates a set declaration.
func (e *Evaluator) VisitSetDecl(node *ast.SetDecl, ctx *ExecutionContext) Value {
	// Set type already registered by semantic analyzer
	// Delegate to adapter for now (Phase 3 migration)
	return e.adapter.EvalNode(node)
}

// extractSimpleTypeName extracts the simple type name from a possibly qualified type string.
// Task 3.5.12: Used for helper registration with simplified type names.
// Example: "array of Integer" -> "array"
func extractSimpleTypeName(typeName string) string {
	if idx := strings.Index(typeName, " "); idx != -1 {
		return typeName[:idx]
	}
	return typeName
}
