package passes

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// DeclarationPass implements Pass 1: Declaration Collection
//
// **Purpose**: Register all type and function names without resolving their references.
// This pass creates "forward declarations" for everything so that Pass 2 can resolve
// references in any order.
//
// **Responsibilities**:
// - Register class names (including forward declarations)
// - Register interface names
// - Register enum names
// - Register record names
// - Register type aliases
// - Register function/procedure signatures (names, parameter names, return types as strings)
// - Register global variables and constants (types as unresolved strings)
// - Mark forward-declared types as incomplete
//
// **What it does NOT do**:
// - Resolve type references (e.g., "Integer" → types.INTEGER_TYPE)
// - Resolve parent classes or interfaces
// - Analyze function bodies
// - Type-check expressions
// - Validate type compatibility
//
// **Dependencies**: None (this is always the first pass)
//
// **Outputs**:
// - TypeRegistry populated with type names (types may be incomplete)
// - Symbols populated with function/variable names (with unresolved type references)
//
// **Example**:
//
//	type TFoo = class; // Forward declaration - registered as incomplete
//	type TBar = class(TFoo) ... end; // Registered with parent name "TFoo" (not resolved yet)
type DeclarationPass struct{}

// NewDeclarationPass creates a new declaration collection pass.
func NewDeclarationPass() *DeclarationPass {
	return &DeclarationPass{}
}

// Name returns the name of this pass.
func (p *DeclarationPass) Name() string {
	return "Pass 1: Declaration Collection"
}

// Run executes the declaration collection pass.
func (p *DeclarationPass) Run(program *ast.Program, ctx *PassContext) error {
	walker := &declarationWalker{
		ctx: ctx,
	}

	// Walk all statements in the program
	for _, stmt := range program.Statements {
		walker.walkStatement(stmt)
	}

	return nil
}

// declarationWalker walks the AST and registers declarations
type declarationWalker struct {
	ctx *PassContext
}

// walkStatement processes a single statement
func (w *declarationWalker) walkStatement(stmt ast.Statement) {
	if stmt == nil {
		return
	}

	switch s := stmt.(type) {
	case *ast.ClassDecl:
		w.registerClassDecl(s)
	case *ast.InterfaceDecl:
		w.registerInterfaceDecl(s)
	case *ast.EnumDecl:
		w.registerEnumDecl(s)
	case *ast.RecordDecl:
		w.registerRecordDecl(s)
	case *ast.FunctionDecl:
		w.registerFunctionDecl(s)
	case *ast.TypeDeclaration:
		w.registerTypeDeclaration(s)
	case *ast.VarDeclStatement:
		w.registerVarStatement(s)
	case *ast.ConstDecl:
		w.registerConstDecl(s)
	// Other statements don't declare types or functions, skip them
	default:
		// Skip expression statements, control flow, etc.
	}
}

// registerClassDecl registers a class declaration
func (w *declarationWalker) registerClassDecl(decl *ast.ClassDecl) {
	if decl == nil || decl.Name == nil {
		return
	}

	className := decl.Name.Value

	// Check if this is a forward declaration
	// A forward declaration has no body - the slices are nil (not initialized)
	// An empty class has initialized but empty slices
	isForward := (decl.Fields == nil &&
		decl.Methods == nil &&
		decl.Properties == nil &&
		decl.Operators == nil &&
		decl.Constants == nil)

	// Check if class is already declared
	existingDesc, _ := w.ctx.TypeRegistry.ResolveDescriptor(className)

	if existingDesc != nil {
		// Type already exists
		if existingClass, ok := existingDesc.Type.(*types.ClassType); ok {
			// Handle forward declaration completion
			if existingClass.IsForward && !isForward {
				// This is the full implementation of a forward-declared class
				// Clear the forward flag so Pass 2 knows it has an implementation
				existingClass.IsForward = false
				return
			}

			// Handle partial class merging
			if existingClass.IsPartial && decl.IsPartial {
				// Both are partial - we'll merge them in Pass 2
				return
			}
		}

		// Duplicate declaration (neither forward nor partial case)
		w.ctx.AddError("type '%s' already declared at %s", className, existingDesc.Position)
		return
	}

	// Create a skeletal ClassType
	// In Pass 2, we'll resolve parent types and field types
	classType := &types.ClassType{
		Name:                 className,
		IsForward:            isForward,
		IsPartial:            decl.IsPartial,
		IsAbstract:           decl.IsAbstract,
		IsExternal:           decl.IsExternal,
		ExternalName:         decl.ExternalName,
		Fields:               make(map[string]types.Type),
		Methods:              make(map[string]*types.FunctionType),
		MethodOverloads:      make(map[string][]*types.MethodInfo),
		Properties:           make(map[string]*types.PropertyInfo),
		ClassVars:            make(map[string]types.Type),
		Constants:            make(map[string]interface{}),
		ConstantTypes:        make(map[string]types.Type),
		ConstantVisibility:   make(map[string]int),
		ClassVarVisibility:   make(map[string]int),
		FieldVisibility:      make(map[string]int),
		MethodVisibility:     make(map[string]int),
		VirtualMethods:       make(map[string]bool),
		OverrideMethods:      make(map[string]bool),
		AbstractMethods:      make(map[string]bool),
		ForwardedMethods:     make(map[string]bool),
		ReintroduceMethods:   make(map[string]bool),
		ClassMethodFlags:     make(map[string]bool),
		Constructors:         make(map[string]*types.FunctionType),
		ConstructorOverloads: make(map[string][]*types.MethodInfo),
		Operators:            types.NewOperatorRegistry(),
		// Parent and Interfaces will be resolved in Pass 2
		Parent:     nil,
		Interfaces: nil,
	}

	// Register in TypeRegistry
	err := w.ctx.TypeRegistry.Register(
		className,
		classType,
		decl.Token.Pos,
		0, // Visibility will be handled later
	)

	if err != nil {
		w.ctx.AddError("failed to register class '%s': %v", className, err)
	}
}

// registerInterfaceDecl registers an interface declaration
func (w *declarationWalker) registerInterfaceDecl(decl *ast.InterfaceDecl) {
	if decl == nil || decl.Name == nil {
		return
	}

	interfaceName := decl.Name.Value

	// Check if interface is already declared
	existingDesc, _ := w.ctx.TypeRegistry.ResolveDescriptor(interfaceName)
	if existingDesc != nil {
		w.ctx.AddError("type '%s' already declared at %s", interfaceName, existingDesc.Position)
		return
	}

	// Create a skeletal InterfaceType
	// Parent will be resolved in Pass 2
	interfaceType := &types.InterfaceType{
		Name:         interfaceName,
		ExternalName: decl.ExternalName,
		IsExternal:   decl.IsExternal,
		Methods:      make(map[string]*types.FunctionType),
		Parent:       nil, // Will be resolved in Pass 2
	}

	// Register in TypeRegistry
	err := w.ctx.TypeRegistry.Register(
		interfaceName,
		interfaceType,
		decl.Token.Pos,
		0, // Visibility will be handled later
	)

	if err != nil {
		w.ctx.AddError("failed to register interface '%s': %v", interfaceName, err)
	}
}

// registerEnumDecl registers an enum declaration
func (w *declarationWalker) registerEnumDecl(decl *ast.EnumDecl) {
	if decl == nil || decl.Name == nil {
		return
	}

	enumName := decl.Name.Value

	// Check if enum is already declared
	existingDesc, _ := w.ctx.TypeRegistry.ResolveDescriptor(enumName)
	if existingDesc != nil {
		w.ctx.AddError("type '%s' already declared at %s", enumName, existingDesc.Position)
		return
	}

	// Create enum type with values
	// We can populate values now since they don't have type dependencies
	values := make(map[string]int)
	orderedNames := make([]string, 0, len(decl.Values))

	for i, ev := range decl.Values {
		valueName := ev.Name
		var valueInt int

		if ev.Value != nil {
			valueInt = *ev.Value
		} else {
			// Auto-assign value
			if decl.Flags {
				// Flags enum: power of 2 (1, 2, 4, 8, ...)
				valueInt = 1 << i
			} else {
				// Regular enum: sequential (0, 1, 2, 3, ...)
				valueInt = i
			}
		}

		values[valueName] = valueInt
		orderedNames = append(orderedNames, valueName)
	}

	enumType := &types.EnumType{
		Name:         enumName,
		Values:       values,
		OrderedNames: orderedNames,
		Scoped:       decl.Scoped,
		Flags:        decl.Flags,
	}

	// Register in TypeRegistry
	err := w.ctx.TypeRegistry.Register(
		enumName,
		enumType,
		decl.Token.Pos,
		0, // Visibility will be handled later
	)

	if err != nil {
		w.ctx.AddError("failed to register enum '%s': %v", enumName, err)
	}
}

// registerRecordDecl registers a record declaration
func (w *declarationWalker) registerRecordDecl(decl *ast.RecordDecl) {
	if decl == nil || decl.Name == nil {
		return
	}

	recordName := decl.Name.Value

	// Check if record is already declared
	existingDesc, _ := w.ctx.TypeRegistry.ResolveDescriptor(recordName)
	if existingDesc != nil {
		w.ctx.AddError("type '%s' already declared at %s", recordName, existingDesc.Position)
		return
	}

	// Create a skeletal RecordType
	// Field types will be resolved in Pass 2
	recordType := &types.RecordType{
		Name:                 recordName,
		Fields:               make(map[string]types.Type),
		Methods:              make(map[string]*types.FunctionType),
		MethodOverloads:      make(map[string][]*types.MethodInfo),
		ClassMethods:         make(map[string]*types.FunctionType),
		ClassMethodOverloads: make(map[string][]*types.MethodInfo),
		Properties:           make(map[string]*types.RecordPropertyInfo),
		Constants:            make(map[string]*types.ConstantInfo),
		ClassVars:            make(map[string]types.Type),
		FieldsWithInit:       make(map[string]bool),
	}

	// Register in TypeRegistry
	err := w.ctx.TypeRegistry.Register(
		recordName,
		recordType,
		decl.Token.Pos,
		0, // Visibility will be handled later
	)

	if err != nil {
		w.ctx.AddError("failed to register record '%s': %v", recordName, err)
	}
}

// registerFunctionDecl registers a function/procedure declaration
func (w *declarationWalker) registerFunctionDecl(decl *ast.FunctionDecl) {
	if decl == nil || decl.Name == nil {
		return
	}

	funcName := decl.Name.Value

	// Skip class methods - they're registered as part of their class
	if decl.ClassName != nil {
		return
	}

	// Check if function is already declared
	existing, exists := w.ctx.Symbols.Resolve(funcName)

	if exists {
		// Handle forward declarations
		if existing.IsForward && decl.Body != nil {
			// This is the implementation of a forward-declared function
			// We'll validate and complete it in Pass 2
			return
		}

		// Handle overloaded functions
		if existing.HasOverloadDirective || decl.IsOverload {
			// Overloaded function - we'll register it in Pass 2
			return
		}

		// Duplicate declaration
		w.ctx.AddError("function '%s' already declared", funcName)
		return
	}

	// Create a placeholder FunctionType
	// Parameter and return types will be resolved in Pass 2
	// For now, we just mark it as forward if it has no body
	isForward := decl.Body == nil

	// Register in symbol table with a placeholder type
	// We'll create the full FunctionType in Pass 2
	placeholderType := &types.FunctionType{
		Parameters: nil, // Will be populated in Pass 2
		ReturnType: nil, // Will be set in Pass 2
	}

	// Use DefineOverload to support forward declarations and overloading
	err := w.ctx.Symbols.DefineOverload(funcName, placeholderType, decl.IsOverload, isForward)
	if err != nil {
		w.ctx.AddError("failed to register function '%s': %v", funcName, err)
	}
}

// registerTypeDeclaration registers a type alias/subrange/function pointer declaration
func (w *declarationWalker) registerTypeDeclaration(decl *ast.TypeDeclaration) {
	if decl == nil || decl.Name == nil {
		return
	}

	typeName := decl.Name.Value

	// Check if type is already declared
	existingDesc, _ := w.ctx.TypeRegistry.ResolveDescriptor(typeName)
	if existingDesc != nil {
		w.ctx.AddError("type '%s' already declared at %s", typeName, existingDesc.Position)
		return
	}

	// For type aliases, subranges, and function pointers, we need to resolve
	// the referenced types in Pass 2. For now, just register a placeholder.
	// The actual type will be created in Pass 2.

	// We'll store type declarations differently - maybe in a separate map
	// in PassContext for Pass 2 to process. For now, skip registration.
	// TODO: Implement in Pass 2
}

// registerVarStatement registers a variable declaration
func (w *declarationWalker) registerVarStatement(stmt *ast.VarDeclStatement) {
	if stmt == nil || len(stmt.Names) == 0 {
		return
	}

	// Register each variable name
	// Type resolution happens in Pass 2
	for _, nameIdent := range stmt.Names {
		if nameIdent == nil {
			continue
		}

		varName := nameIdent.Value

		// Check if variable already exists
		_, exists := w.ctx.Symbols.Resolve(varName)
		if exists {
			w.ctx.AddError("variable '%s' already declared", varName)
			continue
		}

		// Register with nil type - will be resolved in Pass 2
		w.ctx.Symbols.Define(varName, nil)
	}
}

// registerConstDecl registers a constant declaration
func (w *declarationWalker) registerConstDecl(decl *ast.ConstDecl) {
	if decl == nil || decl.Name == nil {
		return
	}

	constName := decl.Name.Value

	// Skip class constants - they're registered as part of their class
	if decl.IsClassConst {
		return
	}

	// Check if constant already exists
	_, exists := w.ctx.Symbols.Resolve(constName)
	if exists {
		w.ctx.AddError("constant '%s' already declared", constName)
		return
	}

	// Register with nil type and value - will be resolved in Pass 2
	w.ctx.Symbols.DefineConst(constName, nil, nil)
}
