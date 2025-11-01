package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Interface Analysis
// ============================================================================

// analyzeInterfaceDecl analyzes an interface declaration
func (a *Analyzer) analyzeInterfaceDecl(decl *ast.InterfaceDecl) {
	interfaceName := decl.Name.Value

	// Check if interface is already declared
	if _, exists := a.interfaces[interfaceName]; exists {
		a.addError("interface '%s' already declared at %s", interfaceName, decl.Token.Pos.String())
		return
	}

	// Resolve parent interface if specified
	var parentInterface *types.InterfaceType
	if decl.Parent != nil {
		parentName := decl.Parent.Value
		var found bool
		parentInterface, found = a.interfaces[parentName]
		if !found {
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
	a.interfaces[interfaceName] = interfaceType
}

// analyzeInterfaceMethodDecl analyzes an interface method declaration
func (a *Analyzer) analyzeInterfaceMethodDecl(method *ast.InterfaceMethodDecl, iface *types.InterfaceType) {
	methodName := method.Name.Value

	// Build parameter types list
	var paramTypes []types.Type
	for _, param := range method.Parameters {
		paramType, err := a.resolveType(param.Type.Name)
		if err != nil {
			a.addError("unknown parameter type '%s' in interface method '%s' at %s",
				param.Type.Name, methodName, method.Token.Pos.String())
			return
		}
		paramTypes = append(paramTypes, paramType)
	}

	// Determine return type
	var returnType types.Type = types.VOID
	if method.ReturnType != nil {
		var err error
		returnType, err = a.resolveType(method.ReturnType.Name)
		if err != nil {
			a.addError("unknown return type '%s' in interface method '%s' at %s",
				method.ReturnType.Name, methodName, method.Token.Pos.String())
			return
		}
	}

	// Create function type for this interface method
	funcType := types.NewFunctionType(paramTypes, returnType)

	// Add method to interface
	iface.Methods[methodName] = funcType
}

// validateInterfaceImplementation validates that a class implements all required interface methods
func (a *Analyzer) validateInterfaceImplementation(classType *types.ClassType, decl *ast.ClassDecl) {
	// For each interface declared on the class
	for _, ifaceIdent := range decl.Interfaces {
		ifaceName := ifaceIdent.Value

		// Lookup the interface type
		ifaceType, found := a.interfaces[ifaceName]
		if !found {
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
			}
		}
	}
}
