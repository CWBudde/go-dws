package passes

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
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
//   // After Pass 1, we have:
//   // - TypeRegistry["tfoo"] = ClassType{Name: "TFoo", Parent: "TObject" (string)}
//   // - TypeRegistry["tbar"] = ClassType{Name: "TBar", Parent: "TFoo" (string)}
//   //
//   // After Pass 2, we have:
//   // - TypeRegistry["tfoo"] = ClassType{Name: "TFoo", Parent: *types.ClassType{TObject}}
//   // - TypeRegistry["tbar"] = ClassType{Name: "TBar", Parent: *types.ClassType{TFoo}}
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

	// Step 2: Resolve all class parent types and build inheritance chains
	resolver.resolveClassHierarchies()

	// Step 3: Resolve interface parent types
	resolver.resolveInterfaceHierarchies()

	// Step 4: Resolve field types in all classes and records
	resolver.resolveFieldTypes()

	// Step 5: Resolve method signatures (parameters and return types)
	resolver.resolveMethodSignatures()

	// Step 6: Validate that all forward-declared types have implementations
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

// resolveClassParent resolves the parent type for a single class
func (r *typeResolver) resolveClassParent(classType *types.ClassType) {
	// If parent is already resolved, we're done
	if classType.Parent != nil {
		return
	}

	// Find the AST node for this class to get the parent name
	var parentName string
	for _, stmt := range r.program.Statements {
		if classDecl, ok := stmt.(*ast.ClassDecl); ok {
			if classDecl.Name != nil && classDecl.Name.Value == classType.Name {
				if classDecl.Parent != nil {
					parentName = classDecl.Parent.Value
				}
				break
			}
		}
	}

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

	for _, desc := range allTypes {
		if classType, ok := desc.Type.(*types.ClassType); ok {
			if classType.IsForward {
				r.ctx.AddError("forward-declared class '%s' at %s has no implementation",
					classType.Name, desc.Position)
			}
		}
	}
}
