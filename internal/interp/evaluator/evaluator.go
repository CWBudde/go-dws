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
// Phase 3.5.1: This is the main entry point for evaluation.
// In Phase 3.5.2, this will be refactored to use the visitor pattern.
// For now, it delegates back to the Interpreter for backward compatibility.
func (e *Evaluator) Eval(node ast.Node, ctx *ExecutionContext) Value {
	// Track current node for error reporting
	e.currentNode = node

	// Phase 3.5.1: Delegate to the legacy Interpreter during migration.
	// The actual evaluation logic is still in the Interpreter.
	// In the next phase (3.5.2), we'll move the evaluation logic here
	// and implement the visitor pattern.
	if e.adapter != nil {
		return e.adapter.EvalNode(node)
	}

	// If no adapter is set, this is a programming error
	panic("Evaluator.Eval: adapter not set - call SetAdapter first")
}
