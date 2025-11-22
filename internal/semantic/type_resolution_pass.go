package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// TypeResolutionPass implements Pass 2: Type Resolution
//
// **Purpose**: Resolve all type references collected in Pass 1 into concrete type objects.
// This pass builds the complete type system with proper inheritance hierarchies and
// type relationships.
//
// **Responsibilities**:
// - Resolve type references (e.g., "Integer" string → types.INTEGER_TYPE)
// - Resolve class parent types (build inheritance chains)
// - Resolve interface parent interfaces
// - Resolve field types in classes and records
// - Resolve parameter types in function signatures
// - Resolve return types in function signatures
// - Build complete type hierarchies (inheritance chains)
// - Detect circular type dependencies (e.g., type A = B; type B = A;)
// - Validate that forward-declared types have implementations
// - Resolve array element types
// - Resolve set element types
// - Resolve function pointer types
//
// **What it does NOT do**:
// - Type-check expressions or statements
// - Analyze function bodies
// - Validate abstract method implementations
// - Check interface implementations
//
// **Dependencies**: Pass 1 (Declaration Collection)
//
// **Inputs**:
// - TypeRegistry with registered but potentially incomplete types
// - Symbols with unresolved type references
//
// **Outputs**:
// - TypeRegistry with fully resolved types
// - Symbols with resolved type references
// - Complete inheritance hierarchies
// - Errors for undefined types, circular dependencies
//
// **Example**:
//
//	// After Pass 1, we have:
//	// - TypeRegistry["tfoo"] = ClassType{Name: "TFoo", Parent: "TObject" (string)}
//	// - TypeRegistry["tbar"] = ClassType{Name: "TBar", Parent: "TFoo" (string)}
//	//
//	// After Pass 2, we have:
//	// - TypeRegistry["tfoo"] = ClassType{Name: "TFoo", Parent: *types.ClassType{TObject}}
//	// - TypeRegistry["tbar"] = ClassType{Name: "TBar", Parent: *types.ClassType{TFoo}}
type TypeResolutionPass struct{}

// NewTypeResolutionPass creates a new type resolution pass.
func NewTypeResolutionPass() *TypeResolutionPass {
	return &TypeResolutionPass{}
}

// Name returns the name of this pass.
func (p *TypeResolutionPass) Name() string {
	return "Pass 2: Type Resolution"
}

// Run executes the type resolution pass.
func (p *TypeResolutionPass) Run(program *ast.Program, ctx *PassContext) error {
	resolver := &typeResolver{
		ctx:     ctx,
		program: program,
		visited: make(map[string]bool),
	}

	// Step 1: Ensure built-in types are registered
	resolver.registerBuiltinTypes()

	// Step 2: Resolve type declarations (aliases, subranges, function pointers)
	resolver.resolveTypeDeclarations()

	// Step 3: Resolve all class parent types and build inheritance chains
	resolver.resolveClassHierarchies()

	// Step 4: Resolve interface parent types
	resolver.resolveInterfaceHierarchies()

	// Step 5: Resolve field types in all classes and records
	resolver.resolveFieldTypes()

	// Step 6: Resolve method signatures (parameters and return types)
	resolver.resolveMethodSignatures()

	// Step 7: Validate that all forward-declared types have implementations
	resolver.validateForwardDeclarations()

	return nil
}

// typeResolver handles the type resolution logic
type typeResolver struct {
	ctx     *PassContext
	program *ast.Program
	visited map[string]bool // For circular dependency detection
}

// registerBuiltinTypes ensures all built-in types are available in the TypeRegistry
func (r *typeResolver) registerBuiltinTypes() {
	// Register primitive types
	builtins := map[string]types.Type{
		"Integer": types.INTEGER,
		"Float":   types.FLOAT,
		"String":  types.STRING,
		"Boolean": types.BOOLEAN,
		"Variant": types.VARIANT,
		"Void":    types.VOID,
	}

	for name, typ := range builtins {
		if !r.ctx.TypeRegistry.Has(name) {
			_ = r.ctx.TypeRegistry.RegisterBuiltIn(name, typ)
		}
	}
}

// resolveTypeDeclarations resolves all type declarations (aliases, subranges, function pointers)
func (r *typeResolver) resolveTypeDeclarations() {
	// Walk through all statements in the program
	for _, stmt := range r.program.Statements {
		typeDecl, ok := stmt.(*ast.TypeDeclaration)
		if !ok {
			continue
		}

		r.resolveTypeDeclaration(typeDecl)
	}
}

// resolveTypeDeclaration resolves a single type declaration
func (r *typeResolver) resolveTypeDeclaration(decl *ast.TypeDeclaration) {
	if decl == nil || decl.Name == nil {
		return
	}

	typeName := decl.Name.Value

	// Check if type is already declared
	if r.ctx.TypeRegistry.Has(typeName) {
		r.ctx.AddError("type '%s' already declared", typeName)
		return
	}

	// Handle subrange types
	if decl.IsSubrange {
		r.resolveSubrangeType(decl)
		return
	}

	// Handle function pointer types
	if decl.IsFunctionPointer {
		r.resolveFunctionPointerType(decl)
		return
	}

	// Handle type aliases
	if decl.IsAlias {
		r.resolveTypeAlias(decl)
		return
	}
}

// resolveSubrangeType resolves a subrange type declaration
func (r *typeResolver) resolveSubrangeType(decl *ast.TypeDeclaration) {
	// Evaluate low bound (must be compile-time constant)
	lowBound, err := r.evaluateConstantInt(decl.LowBound)
	if err != nil {
		r.ctx.AddError("subrange low bound must be a compile-time constant integer: %v", err)
		return
	}

	// Evaluate high bound (must be compile-time constant)
	highBound, err := r.evaluateConstantInt(decl.HighBound)
	if err != nil {
		r.ctx.AddError("subrange high bound must be a compile-time constant integer: %v", err)
		return
	}

	// Validate low <= high
	if lowBound > highBound {
		r.ctx.AddError("subrange low bound (%d) cannot be greater than high bound (%d)",
			lowBound, highBound)
		return
	}

	// Create SubrangeType
	subrangeType := &types.SubrangeType{
		BaseType:  types.INTEGER, // Subranges are currently based on Integer
		Name:      decl.Name.Value,
		LowBound:  lowBound,
		HighBound: highBound,
	}

	// Register in TypeRegistry (visibility 0 = private/program-level)
	err = r.ctx.TypeRegistry.Register(decl.Name.Value, subrangeType, decl.Token.Pos, 0)
	if err != nil {
		r.ctx.AddError("failed to register subrange type '%s': %v", decl.Name.Value, err)
		return
	}

	// Also register in Subranges map for backward compatibility
	// Use normalized key for case-insensitive lookup (DWScript is case-insensitive)
	r.ctx.Subranges[ident.Normalize(decl.Name.Value)] = subrangeType
}

// resolveFunctionPointerType resolves a function pointer type declaration
// TODO: Implement this when needed (Task 6.1.2.1 already completed)
func (r *typeResolver) resolveFunctionPointerType(decl *ast.TypeDeclaration) {
	// Function pointer types are already handled by the old analyzer
	// This is a placeholder for future migration
}

// resolveTypeAlias resolves a type alias declaration
func (r *typeResolver) resolveTypeAlias(decl *ast.TypeDeclaration) {
	// Type aliases are already handled by the old analyzer
	// This is a placeholder for future migration
}

// evaluateConstantInt evaluates a compile-time constant integer expression
func (r *typeResolver) evaluateConstantInt(expr ast.Expression) (int, error) {
	if expr == nil {
		return 0, fmt.Errorf("nil expression")
	}

	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		// Direct integer literal
		return int(e.Value), nil

	case *ast.Identifier:
		// Constant identifier reference
		sym, ok := r.ctx.Symbols.Resolve(e.Value)
		if !ok {
			return 0, fmt.Errorf("undefined identifier '%s'", e.Value)
		}
		if !sym.IsConst {
			return 0, fmt.Errorf("identifier '%s' is not a constant", e.Value)
		}
		if sym.Value == nil {
			return 0, fmt.Errorf("constant '%s' has no value", e.Value)
		}
		intVal, ok := sym.Value.(int)
		if !ok {
			return 0, fmt.Errorf("constant '%s' is not an integer", e.Value)
		}
		return intVal, nil

	case *ast.UnaryExpression:
		// Handle negative numbers: -40
		if e.Operator == "-" {
			value, err := r.evaluateConstantInt(e.Right)
			if err != nil {
				return 0, err
			}
			return -value, nil
		}
		if e.Operator == "+" {
			// Unary plus: +5
			return r.evaluateConstantInt(e.Right)
		}
		return 0, fmt.Errorf("non-constant unary expression with operator %s", e.Operator)

	case *ast.BinaryExpression:
		// Handle binary expressions
		left, err := r.evaluateConstantInt(e.Left)
		if err != nil {
			return 0, err
		}
		right, err := r.evaluateConstantInt(e.Right)
		if err != nil {
			return 0, err
		}

		// Evaluate based on operator
		switch e.Operator {
		case "+":
			return left + right, nil
		case "-":
			return left - right, nil
		case "*":
			return left * right, nil
		case "div":
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		case "mod":
			if right == 0 {
				return 0, fmt.Errorf("modulo by zero")
			}
			return left % right, nil
		default:
			return 0, fmt.Errorf("non-constant binary operator '%s'", e.Operator)
		}

	default:
		return 0, fmt.Errorf("non-constant expression")
	}
}

// resolveClassHierarchies resolves parent types for all classes
func (r *typeResolver) resolveClassHierarchies() {
	// Get all class types from the registry
	allTypes := r.ctx.TypeRegistry.AllDescriptors()

	for _, desc := range allTypes {
		if classType, ok := desc.Type.(*types.ClassType); ok {
			r.resolveClassParent(classType)
		}
	}
}

// isForwardClassDecl returns true if the class declaration is a forward declaration
// (all member slices are nil, meaning no body was defined)
func isForwardClassDecl(decl *ast.ClassDecl) bool {
	return decl.Fields == nil &&
		decl.Methods == nil &&
		decl.Properties == nil &&
		decl.Operators == nil &&
		decl.Constants == nil
}

// findClassParentName finds the parent class name from the AST for the given class.
// It skips forward declarations to find the actual class definition with a parent.
func (r *typeResolver) findClassParentName(className string) string {
	for _, stmt := range r.program.Statements {
		classDecl, ok := stmt.(*ast.ClassDecl)
		if !ok || classDecl.Name == nil || classDecl.Name.Value != className {
			continue
		}

		if classDecl.Parent != nil {
			return classDecl.Parent.Value // Found the definition with a parent
		}
		if !isForwardClassDecl(classDecl) {
			return "" // Found a non-forward declaration without parent - this is the root
		}
		// Otherwise, this is a forward declaration - keep looking for the full definition
	}
	return ""
}

// resolveClassParent resolves the parent type for a single class
func (r *typeResolver) resolveClassParent(classType *types.ClassType) {
	// If parent is already resolved, we're done
	if classType.Parent != nil {
		return
	}

	parentName := r.findClassParentName(classType.Name)

	// If no parent specified, we're done (root class)
	if parentName == "" {
		return
	}

	// Detect circular dependencies
	if r.visited[classType.Name] {
		r.ctx.AddError("circular dependency detected in class hierarchy for '%s'", classType.Name)
		return
	}

	r.visited[classType.Name] = true
	defer func() { delete(r.visited, classType.Name) }()

	// Resolve parent type
	parentType, ok := r.ctx.TypeRegistry.Resolve(parentName)
	if !ok {
		r.ctx.AddError("parent class '%s' not found for class '%s'", parentName, classType.Name)
		return
	}

	// Verify parent is a class type
	parentClass, ok := parentType.(*types.ClassType)
	if !ok {
		r.ctx.AddError("parent '%s' is not a class type for class '%s'", parentName, classType.Name)
		return
	}

	// Recursively resolve parent's parent first
	r.resolveClassParent(parentClass)

	// Set the parent
	classType.Parent = parentClass
}

// resolveInterfaceHierarchies resolves parent types for all interfaces
func (r *typeResolver) resolveInterfaceHierarchies() {
	// Get all interface types from the registry
	allTypes := r.ctx.TypeRegistry.AllDescriptors()

	for _, desc := range allTypes {
		if interfaceType, ok := desc.Type.(*types.InterfaceType); ok {
			r.resolveInterfaceParent(interfaceType)
		}
	}
}

// resolveInterfaceParent resolves the parent type for a single interface
func (r *typeResolver) resolveInterfaceParent(interfaceType *types.InterfaceType) {
	// If parent is already resolved, we're done
	if interfaceType.Parent != nil {
		return
	}

	// Find the AST node for this interface to get the parent name
	var parentName string
	for _, stmt := range r.program.Statements {
		if interfaceDecl, ok := stmt.(*ast.InterfaceDecl); ok {
			if interfaceDecl.Name != nil && interfaceDecl.Name.Value == interfaceType.Name {
				if interfaceDecl.Parent != nil {
					parentName = interfaceDecl.Parent.Value
				}
				break
			}
		}
	}

	// If no parent specified, we're done
	if parentName == "" {
		return
	}

	// Detect circular dependencies
	if r.visited[interfaceType.Name] {
		r.ctx.AddError("circular dependency detected in interface hierarchy for '%s'", interfaceType.Name)
		return
	}

	r.visited[interfaceType.Name] = true
	defer func() { delete(r.visited, interfaceType.Name) }()

	// Resolve parent type
	parentType, ok := r.ctx.TypeRegistry.Resolve(parentName)
	if !ok {
		r.ctx.AddError("parent interface '%s' not found for interface '%s'", parentName, interfaceType.Name)
		return
	}

	// Verify parent is an interface type
	parentInterface, ok := parentType.(*types.InterfaceType)
	if !ok {
		r.ctx.AddError("parent '%s' is not an interface type for interface '%s'", parentName, interfaceType.Name)
		return
	}

	// Recursively resolve parent's parent first
	r.resolveInterfaceParent(parentInterface)

	// Set the parent
	interfaceType.Parent = parentInterface
}

// resolveFieldTypes resolves field types for all classes and records
func (r *typeResolver) resolveFieldTypes() {
	// TODO: Implement field type resolution
	// This will walk through all classes and records and resolve their field types
}

// resolveMethodSignatures resolves parameter and return types for all methods
func (r *typeResolver) resolveMethodSignatures() {
	// TODO: Implement method signature resolution
	// This will walk through all functions and methods and resolve their parameter/return types
}

// validateForwardDeclarations ensures all forward-declared types have implementations
func (r *typeResolver) validateForwardDeclarations() {
	allTypes := r.ctx.TypeRegistry.AllDescriptors()

	// Validate forward-declared classes have implementations
	for _, desc := range allTypes {
		if classType, ok := desc.Type.(*types.ClassType); ok {
			if classType.IsForward {
				// Use DWScript format: Class "Name" isn't defined completely
				r.ctx.AddError("Class \"%s\" isn't defined completely", classType.Name)
			}
		}
	}

	// Validate forward-declared methods have implementations
	r.validateMethodImplementations()

	// Validate forward-declared functions have implementations
	r.validateFunctionImplementations()
}

// validateMethodImplementations checks that all forward-declared methods have implementations
// Migrated from analyzer.go:validateMethodImplementations()
func (r *typeResolver) validateMethodImplementations() {
	// Iterate through all classes
	classNames := r.ctx.TypeRegistry.TypesByKind("CLASS")
	for _, className := range classNames {
		typ, ok := r.ctx.TypeRegistry.Resolve(className)
		if !ok {
			continue
		}
		classType, ok := typ.(*types.ClassType)
		if !ok {
			continue
		}

		// Check each method and constructor in ForwardedMethods
		for methodName, isForwarded := range classType.ForwardedMethods {
			if !isForwarded {
				continue
			}

			// Skip abstract methods - they don't need implementations
			if classType.AbstractMethods[methodName] {
				continue
			}

			// Skip external methods - they are implemented externally
			if classType.IsExternal {
				continue
			}

			// Determine if this is a constructor or regular method
			isConstructor := len(classType.ConstructorOverloads[methodName]) > 0

			// This method/constructor was declared but never implemented
			// Use classType.Name to preserve original case in error messages
			if isConstructor {
				r.ctx.AddError("constructor '%s.%s' declared but not implemented",
					classType.Name, methodName)
			} else {
				r.ctx.AddError("method '%s.%s' declared but not implemented",
					classType.Name, methodName)
			}
		}
	}
}

// validateFunctionImplementations checks that all forward-declared functions have implementations
// Migrated from analyzer.go:validateFunctionImplementations()
func (r *typeResolver) validateFunctionImplementations() {
	// Walk through all symbols in the global scope
	r.validateFunctionImplementationsInScope(r.ctx.Symbols)
}

// validateFunctionImplementationsInScope checks all function symbols in a scope for unimplemented forward declarations
func (r *typeResolver) validateFunctionImplementationsInScope(scope *SymbolTable) {
	if scope == nil {
		return
	}

	// Check all symbols in this scope
	scope.symbols.Range(func(_ string, symbol *Symbol) bool {
		// Check overload sets first (their Type is nil, so must check before type assertion)
		if symbol.IsOverloadSet {
			for _, overload := range symbol.Overloads {
				if overload.IsForward {
					// Format error message to match DWScript
					// DWScript uses "function" for both functions and procedures
					r.ctx.AddError("Syntax Error: The function \"%s\" was forward declared but not implemented",
						overload.Name)
				}
			}
			return true // continue to next symbol (overload sets don't have individual Type)
		}

		// Check non-overload functions
		_, ok := symbol.Type.(*types.FunctionType)
		if !ok {
			return true // Not a function - continue to next symbol
		}

		// Check if this is a non-overloaded forward function
		if symbol.IsForward {
			// Format error message to match DWScript
			// DWScript uses "function" for both functions and procedures
			r.ctx.AddError("Syntax Error: The function \"%s\" was forward declared but not implemented",
				symbol.Name)
		}
		return true // continue iteration
	})

	// Note: Forward declarations only exist at global scope in DWScript,
	// so no need to traverse parent scopes (scope.outer)
}
