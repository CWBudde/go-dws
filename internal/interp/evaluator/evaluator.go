package evaluator

import (
	"io"
	"math/rand"

	"github.com/cwbudde/go-dws/internal/ast"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/units"
	pkgast "github.com/cwbudde/go-dws/pkg/ast"
)

// Value represents a runtime value in the DWScript interpreter.
// This is temporarily defined here to avoid circular imports during the refactoring.
// In the final architecture, this will be properly organized.
type Value interface {
	Type() string
	String() string
}

// Config holds configuration options for the evaluator.
type Config struct {
	// MaxRecursionDepth is the maximum depth of the call stack.
	MaxRecursionDepth int

	// SourceCode is the original source code being executed (for error reporting).
	SourceCode string

	// SourceFile is the path to the source file (for error reporting).
	SourceFile string
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		MaxRecursionDepth: 1024, // Matches DWScript default
		SourceCode:        "",
		SourceFile:        "",
	}
}

// ExternalFunctionRegistry manages external (Go) functions that can be called from DWScript.
// This is temporarily defined here to avoid circular imports during the refactoring.
type ExternalFunctionRegistry interface {
	// Placeholder for now - will be properly defined later
}

// InterpreterAdapter is a temporary interface to allow the Evaluator to delegate
// back to the Interpreter during the migration process.
// Phase 3.5.1: This will be removed once all evaluation logic is moved to Evaluator.
type InterpreterAdapter interface {
	// EvalNode evaluates a node using the legacy Interpreter.Eval method.
	EvalNode(node ast.Node) Value
}

// Evaluator is responsible for evaluating DWScript AST nodes.
// It holds the necessary dependencies for evaluation (type system, runtime services, configuration)
// but does NOT hold execution state. Execution state is passed via ExecutionContext.
//
// Phase 3.5.1: This struct separates concerns:
// - TypeSystem: manages types, classes, records, interfaces, operators, conversions
// - Runtime services: I/O, random numbers, external functions
// - Configuration: recursion limits, source file information
// - Unit system: manages unit loading and initialization
//
// The Evaluator is stateless with respect to execution - all execution state
// (environment, call stack, exceptions, control flow) is in ExecutionContext.
type Evaluator struct {
	// Type System - manages all type metadata
	typeSystem *interptypes.TypeSystem

	// Runtime Services
	output            io.Writer
	rand              *rand.Rand
	randSeed          int64
	externalFunctions ExternalFunctionRegistry

	// Configuration
	config *Config

	// Unit System
	unitRegistry     *units.UnitRegistry
	initializedUnits map[string]bool
	loadedUnits      []string

	// Semantic Analysis metadata (from parser/semantic analyzer)
	semanticInfo *pkgast.SemanticInfo

	// currentNode tracks the current AST node being evaluated (for error reporting)
	// This is evaluation-local state (not execution state) and is safe to keep here
	currentNode ast.Node

	// Phase 3.5.1: Temporary adapter to delegate to legacy Interpreter during migration
	// This will be removed once all evaluation logic is moved here.
	adapter InterpreterAdapter
}

// NewEvaluator creates a new Evaluator with the given dependencies.
func NewEvaluator(
	typeSystem *interptypes.TypeSystem,
	output io.Writer,
	config *Config,
	unitRegistry *units.UnitRegistry,
) *Evaluator {
	if config == nil {
		config = DefaultConfig()
	}

	// Initialize random number generator with a default seed
	const defaultSeed = int64(1)
	source := rand.NewSource(defaultSeed)

	return &Evaluator{
		typeSystem:       typeSystem,
		output:           output,
		rand:             rand.New(source),
		randSeed:         defaultSeed,
		config:           config,
		unitRegistry:     unitRegistry,
		initializedUnits: make(map[string]bool),
		loadedUnits:      make([]string, 0),
	}
}

// TypeSystem returns the type system instance.
func (e *Evaluator) TypeSystem() *interptypes.TypeSystem {
	return e.typeSystem
}

// Output returns the output writer.
func (e *Evaluator) Output() io.Writer {
	return e.output
}

// SetOutput sets the output writer.
func (e *Evaluator) SetOutput(w io.Writer) {
	e.output = w
}

// Random returns the random number generator.
func (e *Evaluator) Random() *rand.Rand {
	return e.rand
}

// RandomSeed returns the current random seed.
func (e *Evaluator) RandomSeed() int64 {
	return e.randSeed
}

// SetRandomSeed sets the random seed and reinitializes the generator.
func (e *Evaluator) SetRandomSeed(seed int64) {
	e.randSeed = seed
	source := rand.NewSource(seed)
	e.rand = rand.New(source)
}

// ExternalFunctions returns the external function registry.
func (e *Evaluator) ExternalFunctions() ExternalFunctionRegistry {
	return e.externalFunctions
}

// SetExternalFunctions sets the external function registry.
func (e *Evaluator) SetExternalFunctions(reg ExternalFunctionRegistry) {
	e.externalFunctions = reg
}

// Config returns the configuration.
func (e *Evaluator) Config() *Config {
	return e.config
}

// SetConfig sets the configuration.
func (e *Evaluator) SetConfig(cfg *Config) {
	e.config = cfg
}

// MaxRecursionDepth returns the maximum recursion depth.
func (e *Evaluator) MaxRecursionDepth() int {
	return e.config.MaxRecursionDepth
}

// SourceCode returns the source code being executed.
func (e *Evaluator) SourceCode() string {
	return e.config.SourceCode
}

// SourceFile returns the source file path.
func (e *Evaluator) SourceFile() string {
	return e.config.SourceFile
}

// UnitRegistry returns the unit registry.
func (e *Evaluator) UnitRegistry() *units.UnitRegistry {
	return e.unitRegistry
}

// InitializedUnits returns the map of initialized units.
func (e *Evaluator) InitializedUnits() map[string]bool {
	return e.initializedUnits
}

// LoadedUnits returns the list of loaded units.
func (e *Evaluator) LoadedUnits() []string {
	return e.loadedUnits
}

// AddLoadedUnit adds a unit to the list of loaded units.
func (e *Evaluator) AddLoadedUnit(unitName string) {
	e.loadedUnits = append(e.loadedUnits, unitName)
}

// SemanticInfo returns the semantic analysis metadata.
func (e *Evaluator) SemanticInfo() *pkgast.SemanticInfo {
	return e.semanticInfo
}

// SetSemanticInfo sets the semantic analysis metadata.
func (e *Evaluator) SetSemanticInfo(info *pkgast.SemanticInfo) {
	e.semanticInfo = info
}

// CurrentNode returns the current AST node being evaluated (for error reporting).
func (e *Evaluator) CurrentNode() ast.Node {
	return e.currentNode
}

// SetCurrentNode sets the current AST node being evaluated (for error reporting).
func (e *Evaluator) SetCurrentNode(node ast.Node) {
	e.currentNode = node
}

// SetAdapter sets the interpreter adapter for delegation during migration.
// Phase 3.5.1: This is temporary and will be removed once migration is complete.
func (e *Evaluator) SetAdapter(adapter InterpreterAdapter) {
	e.adapter = adapter
}

// Eval evaluates an AST node and returns the result value.
// The execution context contains all execution state (environment, call stack, etc.).
//
// Phase 3.5.2: This uses the visitor pattern to dispatch to appropriate handler methods.
// The giant switch statement from Interpreter.Eval() is now here, but organized with
// visitor methods for better separation of concerns.
func (e *Evaluator) Eval(node ast.Node, ctx *ExecutionContext) Value {
	// Track current node for error reporting
	e.currentNode = node

	// Phase 3.5.2: Visitor pattern dispatch
	// Dispatch to the appropriate visitor method based on node type
	switch n := node.(type) {
	// Literals
	case *ast.IntegerLiteral:
		return e.VisitIntegerLiteral(n, ctx)
	case *ast.FloatLiteral:
		return e.VisitFloatLiteral(n, ctx)
	case *ast.StringLiteral:
		return e.VisitStringLiteral(n, ctx)
	case *ast.BooleanLiteral:
		return e.VisitBooleanLiteral(n, ctx)
	case *ast.CharLiteral:
		return e.VisitCharLiteral(n, ctx)
	case *ast.NilLiteral:
		return e.VisitNilLiteral(n, ctx)

	// Identifiers
	case *ast.Identifier:
		return e.VisitIdentifier(n, ctx)

	// Expressions
	case *ast.BinaryExpression:
		return e.VisitBinaryExpression(n, ctx)
	case *ast.UnaryExpression:
		return e.VisitUnaryExpression(n, ctx)
	case *ast.AddressOfExpression:
		return e.VisitAddressOfExpression(n, ctx)
	case *ast.GroupedExpression:
		return e.VisitGroupedExpression(n, ctx)
	case *ast.CallExpression:
		return e.VisitCallExpression(n, ctx)
	case *ast.NewExpression:
		return e.VisitNewExpression(n, ctx)
	case *ast.MemberAccessExpression:
		return e.VisitMemberAccessExpression(n, ctx)
	case *ast.MethodCallExpression:
		return e.VisitMethodCallExpression(n, ctx)
	case *ast.InheritedExpression:
		return e.VisitInheritedExpression(n, ctx)
	case *ast.SelfExpression:
		return e.VisitSelfExpression(n, ctx)
	case *ast.EnumLiteral:
		return e.VisitEnumLiteral(n, ctx)
	case *ast.RecordLiteralExpression:
		return e.VisitRecordLiteralExpression(n, ctx)
	case *ast.SetLiteral:
		return e.VisitSetLiteral(n, ctx)
	case *ast.ArrayLiteralExpression:
		return e.VisitArrayLiteralExpression(n, ctx)
	case *ast.IndexExpression:
		return e.VisitIndexExpression(n, ctx)
	case *ast.NewArrayExpression:
		return e.VisitNewArrayExpression(n, ctx)
	case *ast.LambdaExpression:
		return e.VisitLambdaExpression(n, ctx)
	case *ast.IsExpression:
		return e.VisitIsExpression(n, ctx)
	case *ast.AsExpression:
		return e.VisitAsExpression(n, ctx)
	case *ast.ImplementsExpression:
		return e.VisitImplementsExpression(n, ctx)
	case *ast.IfExpression:
		return e.VisitIfExpression(n, ctx)
	case *ast.OldExpression:
		return e.VisitOldExpression(n, ctx)

	// Statements
	case *ast.Program:
		return e.VisitProgram(n, ctx)
	case *ast.ExpressionStatement:
		return e.VisitExpressionStatement(n, ctx)
	case *ast.VarDeclStatement:
		return e.VisitVarDeclStatement(n, ctx)
	case *ast.ConstDecl:
		return e.VisitConstDecl(n, ctx)
	case *ast.AssignmentStatement:
		return e.VisitAssignmentStatement(n, ctx)
	case *ast.BlockStatement:
		return e.VisitBlockStatement(n, ctx)
	case *ast.IfStatement:
		return e.VisitIfStatement(n, ctx)
	case *ast.WhileStatement:
		return e.VisitWhileStatement(n, ctx)
	case *ast.RepeatStatement:
		return e.VisitRepeatStatement(n, ctx)
	case *ast.ForStatement:
		return e.VisitForStatement(n, ctx)
	case *ast.ForInStatement:
		return e.VisitForInStatement(n, ctx)
	case *ast.CaseStatement:
		return e.VisitCaseStatement(n, ctx)
	case *ast.TryStatement:
		return e.VisitTryStatement(n, ctx)
	case *ast.RaiseStatement:
		return e.VisitRaiseStatement(n, ctx)
	case *ast.BreakStatement:
		return e.VisitBreakStatement(n, ctx)
	case *ast.ContinueStatement:
		return e.VisitContinueStatement(n, ctx)
	case *ast.ExitStatement:
		return e.VisitExitStatement(n, ctx)
	case *ast.ReturnStatement:
		return e.VisitReturnStatement(n, ctx)
	case *ast.UsesClause:
		return e.VisitUsesClause(n, ctx)

	// Declarations
	case *ast.FunctionDecl:
		return e.VisitFunctionDecl(n, ctx)
	case *ast.ClassDecl:
		return e.VisitClassDecl(n, ctx)
	case *ast.InterfaceDecl:
		return e.VisitInterfaceDecl(n, ctx)
	case *ast.OperatorDecl:
		return e.VisitOperatorDecl(n, ctx)
	case *ast.EnumDecl:
		return e.VisitEnumDecl(n, ctx)
	case *ast.RecordDecl:
		return e.VisitRecordDecl(n, ctx)
	case *ast.HelperDecl:
		return e.VisitHelperDecl(n, ctx)
	case *ast.ArrayDecl:
		return e.VisitArrayDecl(n, ctx)
	case *ast.TypeDeclaration:
		return e.VisitTypeDeclaration(n, ctx)

	default:
		// Phase 3.5.2: Unknown node type - delegate to adapter if available
		// This provides a safety net during the migration
		if e.adapter != nil {
			return e.adapter.EvalNode(node)
		}
		// If no adapter, this is an error (unknown node type)
		panic("Evaluator.Eval: unknown node type and no adapter available")
	}
}
