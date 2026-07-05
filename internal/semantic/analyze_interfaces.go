package semantic

import (
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Interface Analysis
// ============================================================================

// analyzeInterfaceDecl analyzes an interface declaration
func (a *Analyzer) analyzeInterfaceDecl(decl *ast.InterfaceDecl) {
	interfaceName := decl.Name.Value

	// Check if interface is already declared (use lowercase for case-insensitive duplicate check)
	if a.hasType(interfaceName) {
		a.addError("%s", errors.FormatNameAlreadyExists(interfaceName, decl.Token.Pos.Line, decl.Token.Pos.Column))
		return
	}

	// Resolve parent interface if specified
	var parentInterface *types.InterfaceType
	if decl.Parent != nil {
		parentName := decl.Parent.Value
		parentInterface = a.getInterfaceType(parentName)
		if parentInterface == nil {
			a.addError("parent interface '%s' not found at %s", parentName, decl.Token.Pos.String())
			return
		}
	}

	// Create new interface type
	interfaceType := types.NewInterfaceType(interfaceName)
	interfaceType.Parent = parentInterface

	// Set external flag and name if specified
	if decl.IsExternal {
		interfaceType.IsExternal = true
		interfaceType.ExternalName = decl.ExternalName
	}

	// Analyze each method in the interface
	for _, method := range decl.Methods {
		a.analyzeInterfaceMethodDecl(method, interfaceType)
	}

	// Analyze each property in the interface
	for _, property := range decl.Properties {
		a.analyzeInterfacePropertyDecl(property, interfaceType)
	}

	// Register interface in the registry
	// Use lowercase key for case-insensitive lookup
	a.registerTypeWithPos(interfaceName, interfaceType, decl.Token.Pos)
}

// analyzeInterfaceMethodDecl analyzes an interface method declaration
func (a *Analyzer) analyzeInterfaceMethodDecl(method *ast.InterfaceMethodDecl, iface *types.InterfaceType) {
	methodName := method.Name.Value

	// Build parameter types list
	var paramTypes []types.Type
	for _, param := range method.Parameters {
		paramType, err := a.resolveType(getTypeExpressionName(param.Type))
		if err != nil {
			a.addError("unknown parameter type '%s' in interface method '%s' at %s",
				getTypeExpressionName(param.Type), methodName, method.Token.Pos.String())
			return
		}
		paramTypes = append(paramTypes, paramType)
	}

	// Determine return type
	var returnType types.Type = types.VOID
	if method.ReturnType != nil {
		var err error
		returnType, err = a.resolveType(getTypeExpressionName(method.ReturnType))
		if err != nil {
			a.addError("unknown return type '%s' in interface method '%s' at %s",
				getTypeExpressionName(method.ReturnType), methodName, method.Token.Pos.String())
			return
		}
	}

	// Create function type for this interface method
	funcType := types.NewFunctionType(paramTypes, returnType)

	// Check for duplicate method (case-insensitive)
	methodKey := ident.Normalize(methodName)

	// Check if method already exists in this interface
	if _, exists := iface.Methods[methodKey]; exists {
		a.addError("%s", errors.FormatNameAlreadyExists(methodName, method.Token.Pos.Line, method.Token.Pos.Column))
		return
	}

	// Check if method exists in parent interface (inherited methods cannot be redeclared)
	if iface.Parent != nil {
		parentMethods := types.GetAllInterfaceMethods(iface.Parent)
		if _, exists := parentMethods[methodKey]; exists {
			a.addError("%s", errors.FormatNameAlreadyExists(methodName, method.Token.Pos.Line, method.Token.Pos.Column))
			return
		}
	}

	// Add method to interface
	iface.Methods[methodKey] = funcType
}

// analyzeInterfacePropertyDecl analyzes a property declared on an interface and
// registers its PropertyInfo. Interface property accessors (method names or
// inline expressions) reference members that are validated against the
// implementing class, so only the declared property type and access kinds are
// recorded here; the expression accessors are executed at runtime against the
// underlying object.
func (a *Analyzer) analyzeInterfacePropertyDecl(prop *ast.PropertyDecl, iface *types.InterfaceType) {
	if prop == nil || prop.Name == nil {
		return
	}
	propName := prop.Name.Value
	propKey := ident.Normalize(propName)

	if _, exists := iface.Properties[propKey]; exists {
		a.addError("%s", errors.FormatNameAlreadyExists(propName, prop.Token.Pos.Line, prop.Token.Pos.Column))
		return
	}

	if prop.Type == nil {
		a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
			"property '"+propName+"' missing type annotation in interface '"+iface.Name+"'"))
		return
	}
	propType, err := a.resolveType(getTypeExpressionName(prop.Type))
	if err != nil {
		a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
			"unknown type '"+getTypeExpressionName(prop.Type)+"' for property '"+propName+"' in interface '"+iface.Name+"'"))
		return
	}

	propInfo := &types.PropertyInfo{
		Name:            propName,
		Type:            propType,
		IsIndexed:       len(prop.IndexParams) > 0,
		IsDefault:       prop.IsDefault,
		IsClassProperty: prop.IsClassProperty,
	}

	// Record read access kind.
	if prop.ReadSpec != nil {
		if id, ok := prop.ReadSpec.(*ast.Identifier); ok {
			propInfo.ReadSpec = id.Value
			propInfo.ReadKind = types.PropAccessMethod
		} else {
			propInfo.ReadKind = types.PropAccessExpression
			propInfo.ReadExpr = prop.ReadSpec
		}
	} else {
		propInfo.ReadKind = types.PropAccessNone
	}

	// Record write access kind.
	switch {
	case prop.WriteSpec != nil:
		// A bare write specifier must be a field or method name; parenthesized
		// expression setters arrive via WriteStmt instead. Reject other shapes
		// rather than silently degrading the property to read-only.
		id, ok := prop.WriteSpec.(*ast.Identifier)
		if !ok {
			a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
				"property '"+propName+"' write specifier must be a field or method name in interface '"+iface.Name+"'"))
			return
		}
		propInfo.WriteSpec = id.Value
		propInfo.WriteKind = types.PropAccessMethod
	case prop.WriteStmt != nil:
		propInfo.WriteKind = types.PropAccessExpression
		propInfo.WriteExpr = prop.WriteStmt
	default:
		propInfo.WriteKind = types.PropAccessNone
	}

	iface.Properties[propKey] = propInfo
}

// validateInterfaceImplementation validates that a class implements all required interface methods
func (a *Analyzer) validateInterfaceImplementation(classType *types.ClassType, decl *ast.ClassDecl) {
	// For each interface declared on the class
	for _, ifaceIdent := range decl.Interfaces {
		ifaceName := ifaceIdent.Value

		// Lookup the interface type (use lowercase for case-insensitive lookup)
		ifaceType := a.getInterfaceType(ifaceName)
		if ifaceType == nil {
			a.addError("interface '%s' not found at %s", ifaceName, decl.Token.Pos.String())
			continue
		}

		// Store interface in class type's Interfaces list
		classType.Interfaces = append(classType.Interfaces, ifaceType)

		// Check that class implements all interface methods
		allMethods := types.GetAllInterfaceMethods(ifaceType)
		for methodName, ifaceMethod := range allMethods {
			// Check if class has this method
			classMethod, hasMethod := classType.GetMethod(methodName)
			if !hasMethod {
				a.addError("class '%s' does not implement interface method '%s' from interface '%s' at %s",
					classType.Name, methodName, ifaceName, decl.Token.Pos.String())
				continue
			}

			// Check that signatures match
			// Use existing methodSignaturesMatch from analyzer.go:1038
			if !a.methodSignaturesMatch(classMethod, ifaceMethod) {
				a.addError("method '%s' in class '%s' does not match interface signature from '%s' at %s",
					methodName, classType.Name, ifaceName, decl.Token.Pos.String())
			} else {
				// Clear the forward flag since this method implements the interface
				// Methods implementing interfaces are complete implementations, not forward declarations
				forwardKey := ident.Normalize(classType.Name) + "." + ident.Normalize(methodName)
				delete(classType.ForwardedMethods, ident.Normalize(methodName))
				delete(a.forwardMethodPos, forwardKey)
				delete(a.forwardMethodNames, forwardKey)
			}
		}
	}
}
