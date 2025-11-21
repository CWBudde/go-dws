package passes

import (
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
//   type TFoo = class; // Forward declaration - registered as incomplete
//   type TBar = class(TFoo) ... end; // Registered with parent name "TFoo" (not resolved yet)
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
	isForward := (decl.Fields == nil &&
		decl.Methods == nil &&
		decl.Properties == nil &&
		decl.Operators == nil &&
		decl.Constants == nil)

	// TODO: For now, we just note that this exists
	// In Pass 2, we'll create the actual ClassType and resolve parent references
	// For now, we just register the name with a placeholder
	_ = className
	_ = isForward

	// TODO: Register in TypeRegistry once we figure out how to handle incomplete types
}

// registerInterfaceDecl registers an interface declaration
func (w *declarationWalker) registerInterfaceDecl(decl *ast.InterfaceDecl) {
	if decl == nil || decl.Name == nil {
		return
	}

	// TODO: Implement interface registration
}

// registerEnumDecl registers an enum declaration
func (w *declarationWalker) registerEnumDecl(decl *ast.EnumDecl) {
	if decl == nil || decl.Name == nil {
		return
	}

	// TODO: Implement enum registration
}

// registerRecordDecl registers a record declaration
func (w *declarationWalker) registerRecordDecl(decl *ast.RecordDecl) {
	if decl == nil || decl.Name == nil {
		return
	}

	// TODO: Implement record registration
}

// registerFunctionDecl registers a function/procedure declaration
func (w *declarationWalker) registerFunctionDecl(decl *ast.FunctionDecl) {
	if decl == nil || decl.Name == nil {
		return
	}

	// TODO: Implement function registration
}

// registerTypeDeclaration registers a type alias/subrange/function pointer declaration
func (w *declarationWalker) registerTypeDeclaration(decl *ast.TypeDeclaration) {
	if decl == nil || decl.Name == nil {
		return
	}

	// TODO: Implement type declaration registration
}

// registerVarStatement registers a variable declaration
func (w *declarationWalker) registerVarStatement(stmt *ast.VarDeclStatement) {
	if stmt == nil {
		return
	}

	// TODO: Implement variable registration
}

// registerConstDecl registers a constant declaration
func (w *declarationWalker) registerConstDecl(decl *ast.ConstDecl) {
	if decl == nil || decl.Name == nil {
		return
	}

	// TODO: Implement constant registration
}
