package semantic

import (
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
		a.addError("interface method '%s' already declared in interface '%s' at %s",
			methodName, iface.Name, method.Token.Pos.String())
		return
	}

	// Check if method exists in parent interface (inherited methods cannot be redeclared)
	if iface.Parent != nil {
		parentMethods := types.GetAllInterfaceMethods(iface.Parent)
		if _, exists := parentMethods[methodKey]; exists {
			a.addError("interface method '%s' already declared in interface '%s' at %s",
				methodName, iface.Name, method.Token.Pos.String())
			return
		}
	}

	// Add method to interface
	iface.Methods[methodKey] = funcType
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
				delete(classType.ForwardedMethods, ident.Normalize(methodName))
			}
		}
	}
}
