package semantic

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	pkgast "github.com/cwbudde/go-dws/pkg/ast" // Task 9.18
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token" // Task 6.1.1.3: for TypeRegistry
)

// ============================================================================
// Type Expression Helpers
// ============================================================================

// getTypeExpressionName extracts a type name string from a TypeExpression.
// For simple types, returns the identifier name.
// For complex types (arrays, function pointers), returns the string representation.
func getTypeExpressionName(typeExpr ast.TypeExpression) string {
	if typeExpr == nil {
		return ""
	}

	switch te := typeExpr.(type) {
	case *ast.TypeAnnotation:
		if te == nil {
			return ""
		}
		if te.InlineType != nil {
			return getTypeExpressionName(te.InlineType)
		}
		return te.Name
	case *ast.ArrayTypeNode:
		return te.String()
	case *ast.FunctionPointerTypeNode:
		return te.String()
	default:
		return typeExpr.String()
	}
}

// ============================================================================
// Analyzer
// ============================================================================

// Analyzer performs semantic analysis on a DWScript program.
// It validates types, checks for undefined variables, and ensures
// type compatibility in expressions and statements.
type Analyzer struct {
	symbols            *SymbolTable
	globalOperators    *types.OperatorRegistry
	typeRegistry       *TypeRegistry // Task 6.1.1: Unified type registry (replaces 7 scattered maps)
	subranges          map[string]*types.SubrangeType
	functionPointers   map[string]*types.FunctionPointerType
	currentFunction    *ast.FunctionDecl
	currentClass       *types.ClassType
	currentNestedTypes map[string]string
	currentRecord      *types.RecordType
	conversionRegistry *types.ConversionRegistry
	semanticInfo       *pkgast.SemanticInfo
	unitSymbols        map[string]*SymbolTable
	helpers            map[string][]*types.HelperType
	sourceCode         string
	currentProperty    string
	sourceFile         string
	errors             []string
	structuredErrors   []*SemanticError
	loopDepth          int
	inExceptionHandler bool
	inFinallyBlock     bool
	inLoop             bool
	inLambda           bool
	inClassMethod      bool
	inPropertyExpr     bool
	nestedTypeAliases  map[string]map[string]string

	// experimentalPasses enables the new multi-pass semantic analysis system.
	// When false (default), only the old analyzer runs, keeping behavior stable.
	// When true, Pass 2 (Type Resolution) and Pass 3 (Semantic Validation) also run.
	// Use NewAnalyzerWithExperimentalPasses() to enable for task 6.1.2 development.
	experimentalPasses bool
}

// NewAnalyzer creates a new semantic analyzer
func NewAnalyzer() *Analyzer {
	a := &Analyzer{
		symbols:            NewSymbolTable(),
		typeRegistry:       NewTypeRegistry(), // Task 6.1.1: Unified type registry
		unitSymbols:        make(map[string]*SymbolTable),
		errors:             make([]string, 0),
		structuredErrors:   make([]*SemanticError, 0), // Task 9.110
		subranges:          make(map[string]*types.SubrangeType),
		functionPointers:   make(map[string]*types.FunctionPointerType), // Task 9.159
		helpers:            make(map[string][]*types.HelperType),        // Task 9.82
		globalOperators:    types.NewOperatorRegistry(),
		conversionRegistry: types.NewConversionRegistry(),
		semanticInfo:       pkgast.NewSemanticInfo(), // Task 9.18
		nestedTypeAliases:  make(map[string]map[string]string),
	}

	// Register built-in Exception base class
	a.registerBuiltinExceptionTypes()

	// Register built-in interfaces
	a.registerBuiltinInterfaces()

	// Register built-in array helpers
	a.initArrayHelpers()

	// Register built-in helpers for primitive types
	a.initIntrinsicHelpers()

	// Register built-in enum helpers
	a.initEnumHelpers()

	// Register mathematical constants (for DWScript compatibility)
	a.symbols.DefineConst("Pi", types.FLOAT, math.Pi)

	// Task 9.4.1: Register Variant special values as built-in constants
	a.symbols.DefineConst("Null", types.VARIANT, nil)       // Null is a variant special value
	a.symbols.DefineConst("Unassigned", types.VARIANT, nil) // Unassigned is a variant special value

	return a
}

// NewAnalyzerWithExperimentalPasses creates a new semantic analyzer with experimental
// multi-pass analysis enabled. This runs both the old analyzer AND the new Pass 2/3
// system for comparison and development purposes.
//
// Use this constructor when working on task 6.1.2 (multi-pass semantic analysis).
// For normal usage, use NewAnalyzer() which only runs the stable old analyzer.
//
// Example usage in tests:
//
//	analyzer := semantic.NewAnalyzerWithExperimentalPasses()
//	err := analyzer.Analyze(program)
func NewAnalyzerWithExperimentalPasses() *Analyzer {
	a := NewAnalyzer()
	a.experimentalPasses = true
	return a
}

// registerBuiltinExceptionTypes registers Exception and standard exception types
func (a *Analyzer) registerBuiltinExceptionTypes() {
	// Register TObject as the root base class for all classes
	// Required for DWScript compatibility
	objectClass := types.NewClassType("TObject", nil)

	// Add basic Create constructor
	objectClass.AddConstructorOverload("Create", &types.MethodInfo{
		Signature: &types.FunctionType{
			Parameters: []types.Type{}, // no parameters
			ReturnType: objectClass,
		},
		Visibility: int(ast.VisibilityPublic),
	})

	// Add default virtual destructor Destroy
	destroySig := &types.FunctionType{
		Parameters: []types.Type{}, // no parameters
		ReturnType: types.VOID,
	}
	objectClass.AddMethodOverload("Destroy", &types.MethodInfo{
		Signature:  destroySig,
		IsVirtual:  true,
		Visibility: int(ast.VisibilityPublic),
	})
	// Add TObject.Free which delegates to Destroy at runtime
	freeSig := &types.FunctionType{
		Parameters: []types.Type{}, // no parameters
		ReturnType: types.VOID,
	}
	objectClass.AddMethodOverload("Free", &types.MethodInfo{
		Signature:  freeSig,
		Visibility: int(ast.VisibilityPublic),
	})

	// Track visibility/virtual flags for built-in methods
	objectClass.VirtualMethods[ident.Normalize("Destroy")] = true
	objectClass.MethodVisibility[ident.Normalize("Destroy")] = int(ast.VisibilityPublic)
	objectClass.MethodVisibility[ident.Normalize("Free")] = int(ast.VisibilityPublic)

	// Add ClassName method (returns the runtime type name)
	objectClass.Methods["ClassName"] = &types.FunctionType{
		Parameters: []types.Type{},
		ReturnType: types.STRING,
	}

	// Task 6.1.1.3: Use TypeRegistry instead of scattered maps
	a.registerBuiltinType("TObject", objectClass)

	// Register TClass as a type alias for "class of TObject"
	// TClass is a built-in metaclass type in DWScript
	tclassAlias := &types.TypeAlias{
		Name:        "TClass",
		AliasedType: types.NewClassOfType(objectClass),
	}
	a.registerBuiltinType("TClass", tclassAlias)

	// Define Exception base class
	exceptionClass := &types.ClassType{
		Name:                 "Exception",
		Parent:               objectClass, // Exception inherits from TObject
		Fields:               make(map[string]types.Type),
		Methods:              make(map[string]*types.FunctionType),
		FieldVisibility:      make(map[string]int),
		MethodVisibility:     make(map[string]int),
		VirtualMethods:       make(map[string]bool),
		OverrideMethods:      make(map[string]bool),
		AbstractMethods:      make(map[string]bool),
		ReintroduceMethods:   make(map[string]bool), // Task 9.2
		Constructors:         make(map[string]*types.FunctionType),
		ConstructorOverloads: make(map[string][]*types.MethodInfo),
		Interfaces:           make([]*types.InterfaceType, 0),
		Properties:           make(map[string]*types.PropertyInfo),
		ClassMethodFlags:     make(map[string]bool),
	}

	// Add Message field to Exception
	exceptionClass.Fields["Message"] = types.STRING

	// Add Create constructor
	exceptionClass.AddConstructorOverload("Create", &types.MethodInfo{
		Signature: &types.FunctionType{
			Parameters: []types.Type{types.STRING}, // message parameter
			ReturnType: exceptionClass,
		},
		Visibility: int(ast.VisibilityPublic),
	})

	// Task 6.1.1.3: Use TypeRegistry instead of scattered maps
	a.registerBuiltinType("Exception", exceptionClass)

	// Define standard exception types
	standardExceptions := []string{
		"EConvertError",    // Type conversion failures
		"ERangeError",      // Array bounds, invalid ranges
		"EDivByZero",       // Division by zero
		"EAssertionFailed", // Failed assertions
		"EInvalidOp",       // Invalid operations
	}

	for _, excName := range standardExceptions {
		excClass := &types.ClassType{
			Name:                 excName,
			Parent:               exceptionClass, // All inherit from Exception
			Fields:               make(map[string]types.Type),
			Methods:              make(map[string]*types.FunctionType),
			FieldVisibility:      make(map[string]int),
			MethodVisibility:     make(map[string]int),
			VirtualMethods:       make(map[string]bool),
			OverrideMethods:      make(map[string]bool),
			AbstractMethods:      make(map[string]bool),
			ReintroduceMethods:   make(map[string]bool), // Task 9.2
			Constructors:         make(map[string]*types.FunctionType),
			ConstructorOverloads: make(map[string][]*types.MethodInfo),
			Interfaces:           make([]*types.InterfaceType, 0),
			Properties:           make(map[string]*types.PropertyInfo),
			ClassMethodFlags:     make(map[string]bool),
		}

		// Inherit Message field from Exception
		excClass.Fields["Message"] = types.STRING

		// Inherit Create constructor
		excClass.AddConstructorOverload("Create", &types.MethodInfo{
			Signature: &types.FunctionType{
				Parameters: []types.Type{types.STRING},
				ReturnType: excClass,
			},
			Visibility: int(ast.VisibilityPublic),
		})

		// Task 6.1.1.3: Use TypeRegistry instead of scattered maps
		a.registerBuiltinType(excName, excClass)
	}
}

// registerBuiltinInterfaces registers IInterface, the root interface type.
// IInterface is the root interface available in DWScript for explicit implementation by classes.
// Note: Interfaces do NOT automatically inherit from IInterface unless explicitly declared.
func (a *Analyzer) registerBuiltinInterfaces() {
	// Register IInterface as the root interface type
	// This is a marker interface with no methods, similar to Go's empty interface
	iinterface := &types.InterfaceType{
		Name:    "IInterface",
		Parent:  nil, // Root of the interface hierarchy
		Methods: make(map[string]*types.FunctionType),
	}

	// Task 6.1.1.3: Use TypeRegistry instead of scattered maps
	a.registerBuiltinType("IInterface", iinterface)
}

// Analyze performs semantic analysis on a program.
// Returns nil if analysis succeeds, or an error if there are semantic errors.
//
// By default, only the stable old analyzer runs. To enable the experimental
// multi-pass system (for task 6.1.2 development), use NewAnalyzerWithExperimentalPasses().
//
// Multi-pass architecture (experimental, task 6.1.2.6):
// - Pass 1: Declaration Collection (register types and function names)
// - Pass 2: Type Resolution (resolve type references, build hierarchies)
// - Pass 3: Semantic Validation (type-check expressions, validate statements)
// - Pass 4: Contract Validation (validate requires/ensures/invariant)
func (a *Analyzer) Analyze(program *ast.Program) error {
	if program == nil {
		return fmt.Errorf("cannot analyze nil program")
	}

	// Task 6.1.2.6: PARTIAL MIGRATION TO PASSES
	// The old analyzer handles declaration collection and most validation.
	// Pass 2 (Type Resolution) performs full type resolution including forward declaration validation.
	// Note: Some work is duplicated between old analyzer and Pass 2 during transition period.

	// OLD IMPLEMENTATION: Analyze each statement (handles declaration collection and validation)
	// This is the stable, default behavior used in production.
	for _, stmt := range program.Statements {
		a.analyzeStatement(stmt)
	}

	// Validate forward declarations are all resolved
	// (previously done in Pass 2, now done here for old analyzer)
	a.validateForwardDeclarations()

	// EXPERIMENTAL: Multi-pass semantic analysis (task 6.1.2)
	// Only runs when experimentalPasses is enabled via NewAnalyzerWithExperimentalPasses().
	// This allows development on the new pass system without breaking main branch tests.
	if a.experimentalPasses {
		// NEW IMPLEMENTATION: Run Pass 2, Pass 3 (and eventually Pass 4)
		// Pass 2 performs: built-in type registration, class/interface hierarchy resolution,
		// field type resolution, method signature resolution, and forward declaration validation.
		// Pass 3 performs: full semantic validation (type checking, control flow, etc.)
		// Create PassContext from Analyzer state to share registries
		ctx := a.createPassContext()

		// Create Pass 2 (Type Resolution)
		// Skip Pass 1 (Declaration Collection) since old analyzer already does that
		pass2 := NewTypeResolutionPass()

		// Run Pass 2 directly
		if err := pass2.Run(program, ctx); err != nil {
			// Fatal error in pass execution (not a semantic error)
			return err
		}

		// Task 6.1.2.6.1: Enable Pass 3 (Semantic Validation) in dual mode
		// Run Pass 3 to validate types, expressions, and control flow
		// Note: Old analyzer still runs for comparison during transition
		pass3 := NewValidationPass()
		if err := pass3.Run(program, ctx); err != nil {
			// Fatal error in pass execution (not a semantic error)
			return err
		}

		// Sync PassContext state back to Analyzer
		// The passes may have updated the type registry, symbol table, etc.
		a.syncFromPassContext(ctx)

		// Collect errors from all passes
		a.mergePassErrors(ctx)
	}

	// If we accumulated errors (not hints), return them
	// Task 9.61.4: Hints don't prevent analysis from succeeding
	hasActualErrors := false
	for _, err := range a.errors {
		if !strings.HasPrefix(err, "Hint:") {
			hasActualErrors = true
			break
		}
	}

	if hasActualErrors {
		return &AnalysisError{Errors: a.errors}
	}

	return nil
}

// validateForwardDeclarations ensures all forward-declared types have implementations.
// This is called at the end of analysis to catch incomplete forward declarations.
func (a *Analyzer) validateForwardDeclarations() {
	// Check all registered classes for unresolved forward declarations
	for _, t := range a.typeRegistry.AllTypes() {
		if classType, ok := t.(*types.ClassType); ok {
			if classType.IsForward {
				// Use DWScript format: Class "Name" isn't defined completely
				a.addError("Class \"%s\" isn't defined completely", classType.Name)
			}
		}
	}
}

// createPassContext creates a PassContext from the Analyzer's current state.
// Task 6.1.2.6: Share context between passes (TypeRegistry, SymbolTable, etc.)
func (a *Analyzer) createPassContext() *PassContext {
	ctx := &PassContext{
		// Core registries - shared with passes
		Symbols:            a.symbols,
		TypeRegistry:       a.typeRegistry,
		GlobalOperators:    a.globalOperators,
		ConversionRegistry: a.conversionRegistry,
		SemanticInfo:       a.semanticInfo,
		BuiltinChecker:     a, // Analyzer implements BuiltinChecker interface

		// Secondary registries
		UnitSymbols:      a.unitSymbols,
		Helpers:          a.helpers,
		Subranges:        a.subranges,
		FunctionPointers: a.functionPointers,

		// Error collection
		Errors:           make([]string, 0),
		StructuredErrors: make([]*SemanticError, 0),

		// Pass execution context
		ScopeStack:      []*Scope{NewScope(ScopeGlobal, nil)}, // Initialize with global scope
		CurrentFunction: a.currentFunction,
		CurrentClass:    a.currentClass,
		CurrentRecord:   a.currentRecord,
		CurrentProperty: a.currentProperty,

		// Source code context
		SourceCode: a.sourceCode,
		SourceFile: a.sourceFile,

		// State flags
		LoopDepth:          a.loopDepth,
		InExceptionHandler: a.inExceptionHandler,
		InFinallyBlock:     a.inFinallyBlock,
		InLoop:             a.inLoop,
		InLambda:           a.inLambda,
		InClassMethod:      a.inClassMethod,
		InPropertyExpr:     a.inPropertyExpr,
	}

	return ctx
}

// AnalyzeBuiltin implements the BuiltinChecker interface.
// This allows passes to validate built-in function calls.
func (a *Analyzer) AnalyzeBuiltin(name string, args []ast.Expression, callExpr *ast.CallExpression) (types.Type, bool) {
	return a.analyzeBuiltinFunction(name, args, callExpr)
}

// IsBuiltinFunction implements the BuiltinChecker interface.
// Task 6.1.2.1: This method checks if a function is built-in WITHOUT analyzing arguments.
// Returns the return type (or VARIANT for unknown return types) and true if it's a built-in.
func (a *Analyzer) IsBuiltinFunction(name string) (types.Type, bool) {
	return a.getBuiltinReturnType(name)
}

// syncFromPassContext syncs state from PassContext back to Analyzer.
// Task 6.1.2.6: Passes may update shared state (symbol table, type registry, etc.)
func (a *Analyzer) syncFromPassContext(ctx *PassContext) {
	// Update execution context (may have changed during analysis)
	if funcDecl, ok := ctx.CurrentFunction.(*ast.FunctionDecl); ok {
		a.currentFunction = funcDecl
	}
	a.currentClass = ctx.CurrentClass
	a.currentRecord = ctx.CurrentRecord
	a.currentProperty = ctx.CurrentProperty

	// Update state flags
	a.loopDepth = ctx.LoopDepth
	a.inExceptionHandler = ctx.InExceptionHandler
	a.inFinallyBlock = ctx.InFinallyBlock
	a.inLoop = ctx.InLoop
	a.inLambda = ctx.InLambda
	a.inClassMethod = ctx.InClassMethod
	a.inPropertyExpr = ctx.InPropertyExpr

	// Core registries are shared by reference, no sync needed
	// (symbols, typeRegistry, globalOperators, etc.)
}

// mergePassErrors merges errors from PassContext into Analyzer.
// Task 6.1.2.6: Collect errors from all passes before returning
func (a *Analyzer) mergePassErrors(ctx *PassContext) {
	// Merge string errors
	a.errors = append(a.errors, ctx.Errors...)

	// Merge structured errors
	a.structuredErrors = append(a.structuredErrors, ctx.StructuredErrors...)
}

// Errors returns all accumulated semantic errors
func (a *Analyzer) Errors() []string {
	return a.errors
}

// GetSemanticInfo returns the semantic metadata table populated during analysis.
// This table maps AST nodes to their inferred types and resolved symbols.
// Task 9.18: Separate type metadata from AST nodes.
func (a *Analyzer) GetSemanticInfo() *pkgast.SemanticInfo {
	return a.semanticInfo
}

// StructuredErrors returns all accumulated structured semantic errors
// Task 9.110: New method for structured error access
func (a *Analyzer) StructuredErrors() []*SemanticError {
	return a.structuredErrors
}

// SetSource sets the source code and filename for error display
// Task 9.110: Allows rich error messages with source snippets
func (a *Analyzer) SetSource(source, filename string) {
	a.sourceCode = source
	a.sourceFile = filename
}

// addError adds a semantic error to the error list
func (a *Analyzer) addError(format string, args ...any) {
	a.errors = append(a.errors, fmt.Sprintf(format, args...))
}

// addHint adds a hint message
// Hints are less severe than errors and don't prevent compilation
func (a *Analyzer) addHint(format string, args ...any) {
	a.errors = append(a.errors, fmt.Sprintf("Hint: "+format, args...))
}

// addStructuredError adds a structured semantic error
// Task 9.110: New method for adding rich errors
func (a *Analyzer) addStructuredError(err *SemanticError) {
	a.structuredErrors = append(a.structuredErrors, err)
	// Also add to string errors for backward compatibility
	a.errors = append(a.errors, err.Error())
}

// canAssign checks assignment compatibility, accounting for implicit conversions.
func (a *Analyzer) canAssign(from, to types.Type) bool {
	if from == nil || to == nil {
		return false
	}
	if types.IsCompatible(from, to) {
		return true
	}

	// Allow assigning nil to class types (and vice versa for comparison)
	if from.TypeKind() == "NIL" && to.TypeKind() == "CLASS" {
		return true
	}
	if from.TypeKind() == "CLASS" && to.TypeKind() == "NIL" {
		return true
	}
	// Allow assigning nil to interface types (and vice versa for comparison)
	if from.TypeKind() == "NIL" && to.TypeKind() == "INTERFACE" {
		return true
	}
	if from.TypeKind() == "INTERFACE" && to.TypeKind() == "NIL" {
		return true
	}
	// Task 9.73.1: Allow assigning nil to metaclass types
	if from.TypeKind() == "NIL" && to.TypeKind() == "CLASSOF" {
		return true
	}
	if from.TypeKind() == "CLASSOF" && to.TypeKind() == "NIL" {
		return true
	}
	// Task 9.73.5: Handle ClassOfType to ClassOfType assignment
	// This handles: var meta: TBaseClass; meta := TBase;
	// where TBaseClass = class of TBase
	// Resolve type aliases to get the underlying types
	fromResolved := types.GetUnderlyingType(from)
	toResolved := types.GetUnderlyingType(to)

	if fromMetaclass, ok := fromResolved.(*types.ClassOfType); ok {
		if toMetaclass, ok := toResolved.(*types.ClassOfType); ok {
			// Check if the underlying class types are compatible
			fromClass := fromMetaclass.ClassType
			toClass := toMetaclass.ClassType
			if fromClass.Equals(toClass) || a.isDescendantOf(fromClass, toClass) {
				return true
			}
		}
	}

	if fromClass, ok := from.(*types.ClassType); ok {
		// Task 9.73.1: Allow assigning a class reference to a metaclass variable
		// When a class name is used as a value (e.g., TClassB), it's represented as a ClassType
		// and can be assigned to a variable of type "class of TBase" if it's compatible
		if toMetaclass, ok := to.(*types.ClassOfType); ok {
			// Check if fromClass is the same as or a descendant of the base class
			if fromClass.Equals(toMetaclass.ClassType) || a.isDescendantOf(fromClass, toMetaclass.ClassType) {
				return true
			}
		}
		if toClass, ok := to.(*types.ClassType); ok {
			if fromClass.Equals(toClass) || a.isDescendantOf(fromClass, toClass) {
				return true
			}
		}
		// Task 9.128: Allow assigning a class to an interface it implements
		if toInterface, ok := to.(*types.InterfaceType); ok {
			if fromClass.ImplementsInterface(toInterface) {
				return true
			}
		}
	}
	// Task 9.130: Allow assigning an interface to another compatible interface
	if fromInterface, ok := from.(*types.InterfaceType); ok {
		if toInterface, ok := to.(*types.InterfaceType); ok {
			// Same interface or derived interface can be assigned to base interface
			if fromInterface.Equals(toInterface) || fromInterface.InheritsFrom(toInterface) {
				return true
			}
		}
	}
	// Task 9.98: Check type compatibility for subrange ↔ base type assignments
	// Subrange values can be assigned to their base type (no check needed)
	if fromSubrange, ok := from.(*types.SubrangeType); ok {
		if fromSubrange.BaseType.Equals(to) {
			return true
		}
	}
	// Base type values can be assigned to subrange (runtime check in interpreter)
	if toSubrange, ok := to.(*types.SubrangeType); ok {
		if toSubrange.BaseType.Equals(from) {
			return true
		}
	}
	// Task 9.161: Check function pointer assignment compatibility
	// Task 9.173: Resolve type aliases before checking function pointer compatibility
	toUnderlying := types.GetUnderlyingType(to)
	fromUnderlying := types.GetUnderlyingType(from)

	// Task 9.224: Variant assignment rules
	// Any type can be assigned TO Variant (implicit boxing)
	if toUnderlying.TypeKind() == "VARIANT" {
		return true
	}

	// Variant can be assigned FROM Variant (preserves wrapped value)
	// Note: Variant to typed variable requires runtime checking (will be implemented in interpreter)
	if fromUnderlying.TypeKind() == "VARIANT" && toUnderlying.TypeKind() == "VARIANT" {
		return true
	}

	// Task 9.17.11b: Allow Variant to be assigned to any typed variable
	// This enables str[i] (Variant from array of const) to be assigned to String
	// The interpreter will perform runtime type checking and conversion
	if fromUnderlying.TypeKind() == "VARIANT" {
		// Allow Variant to be assigned to basic types, arrays, records, classes, etc.
		// Runtime will validate and convert
		return true
	}

	// Task 1.6: Allow implicit enum-to-integer conversion
	// Enums can be implicitly converted to Integer (their underlying representation)
	if fromUnderlying.TypeKind() == "ENUM" && toUnderlying.Equals(types.INTEGER) {
		return true
	}

	if toUnderlying.TypeKind() == "FUNCTION_POINTER" || toUnderlying.TypeKind() == "METHOD_POINTER" {
		// Use dedicated function pointer validation which provides detailed errors
		// Note: We don't call validateFunctionPointerAssignment here because it reports errors
		// Instead, we check compatibility directly

		// Task 9.173: Allow method pointers to be assigned to function pointers
		// Check if from (source) can be assigned to to (destination)
		if fromMethodPtr, ok := fromUnderlying.(*types.MethodPointerType); ok {
			return fromMethodPtr.IsCompatibleWith(toUnderlying)
		}

		if toFuncPtr, ok := toUnderlying.(*types.FunctionPointerType); ok {
			return toFuncPtr.IsCompatibleWith(fromUnderlying)
		}
		if toMethodPtr, ok := toUnderlying.(*types.MethodPointerType); ok {
			return toMethodPtr.IsCompatibleWith(fromUnderlying)
		}
	}
	if sig, ok := a.conversionRegistry.FindImplicit(from, to); ok && sig != nil {
		return true
	}
	return false
}

// ============================================================================
// Symbol Table Accessors
// ============================================================================

// GetSymbolTable returns the current symbol table.
// This is used for extracting symbol information for LSP features.
func (a *Analyzer) GetSymbolTable() *SymbolTable {
	return a.symbols
}

// GetClasses returns the analyzer's class type map.
func (a *Analyzer) GetClasses() map[string]*types.ClassType {
	result := make(map[string]*types.ClassType)
	classNames := a.typeRegistry.TypesByKind("CLASS")
	for _, name := range classNames {
		if classType := a.getClassType(name); classType != nil {
			result[strings.ToLower(name)] = classType
		}
	}
	return result
}

// GetInterfaces returns the analyzer's interface type map.
func (a *Analyzer) GetInterfaces() map[string]*types.InterfaceType {
	result := make(map[string]*types.InterfaceType)
	interfaceNames := a.typeRegistry.TypesByKind("INTERFACE")
	for _, name := range interfaceNames {
		if interfaceType := a.getInterfaceType(name); interfaceType != nil {
			result[strings.ToLower(name)] = interfaceType
		}
	}
	return result
}

// GetEnums returns the analyzer's enum type map.
func (a *Analyzer) GetEnums() map[string]*types.EnumType {
	result := make(map[string]*types.EnumType)
	enumNames := a.typeRegistry.TypesByKind("ENUM")
	for _, name := range enumNames {
		if enumType := a.getEnumType(name); enumType != nil {
			result[strings.ToLower(name)] = enumType
		}
	}
	return result
}

// GetRecords returns the analyzer's record type map.
func (a *Analyzer) GetRecords() map[string]*types.RecordType {
	result := make(map[string]*types.RecordType)
	recordNames := a.typeRegistry.TypesByKind("RECORD")
	for _, name := range recordNames {
		if recordType := a.getRecordType(name); recordType != nil {
			result[strings.ToLower(name)] = recordType
		}
	}
	return result
}

// GetArrayTypes returns the analyzer's array type map.
func (a *Analyzer) GetArrayTypes() map[string]*types.ArrayType {
	result := make(map[string]*types.ArrayType)
	arrayNames := a.typeRegistry.TypesByKind("ARRAY")
	for _, name := range arrayNames {
		if arrayType := a.getArrayType(name); arrayType != nil {
			result[strings.ToLower(name)] = arrayType
		}
	}
	return result
}

// GetTypeAliases returns the analyzer's type alias map.
//
// Note: We iterate through all types and check for type aliases via type assertion
// because TypeAlias.TypeKind() returns the underlying type's kind (for compatibility),
// not "ALIAS", so TypesByKind("ALIAS") would return nothing.
func (a *Analyzer) GetTypeAliases() map[string]*types.TypeAlias {
	result := make(map[string]*types.TypeAlias)
	for name, typ := range a.typeRegistry.AllTypes() {
		if aliasType, ok := typ.(*types.TypeAlias); ok {
			result[strings.ToLower(name)] = aliasType
		}
	}
	return result
}

// GetFunctionPointers returns the analyzer's function pointer type map.
func (a *Analyzer) GetFunctionPointers() map[string]*types.FunctionPointerType {
	return a.functionPointers
}

// areArrayTypesCompatibleForVarParam checks if an array type can be passed to a var parameter
// In DWScript, dynamic arrays can be passed to var parameters of static array type if element types match
func (a *Analyzer) areArrayTypesCompatibleForVarParam(argType, paramType types.Type) bool {
	// Check if both are array types
	argArray, argIsArray := argType.(*types.ArrayType)
	paramArray, paramIsArray := paramType.(*types.ArrayType)

	if !argIsArray || !paramIsArray {
		return false
	}

	// Check if element types are compatible
	if !types.IsCompatible(argArray.ElementType, paramArray.ElementType) {
		return false
	}

	// Allow passing dynamic array to static array var parameter
	// or static array to static array var parameter
	return true
}

// validateFieldInitializer validates that a field initializer expression is type-compatible
// with the field's declared type. Used for both class and record field initializers.
func (a *Analyzer) validateFieldInitializer(field *ast.FieldDecl, fieldName string, fieldType types.Type) {
	if field.InitValue != nil {
		initType := a.analyzeExpression(field.InitValue)
		if initType != nil && fieldType != nil {
			// Check type compatibility
			if !types.IsAssignableFrom(fieldType, initType) {
				a.addError("cannot initialize field '%s' of type '%s' with value of type '%s' at %s",
					fieldName, fieldType.String(), initType.String(), field.Token.Pos.String())
			}
		}
	}
}

// ============================================================================
// Type Registry Helper Methods
// ============================================================================
// These methods provide a convenient interface to the TypeRegistry for
// backward compatibility during migration from the old scattered type maps.

// registerType registers a user-defined type with the registry.
// Uses position 0:0 and private visibility (0) as defaults when position is not available.
func (a *Analyzer) registerType(name string, typ types.Type) {
	pos := token.Position{Line: 0, Column: 0, Offset: 0}
	visibility := 0 // private by default
	if err := a.typeRegistry.Register(name, typ, pos, visibility); err != nil {
		// If registration fails (e.g., duplicate), add an error
		a.addError("failed to register type '%s': %v", name, err)
	}
}

// registerTypeWithPos registers a type with explicit position information.
// Uses private visibility (0) as default.
func (a *Analyzer) registerTypeWithPos(name string, typ types.Type, pos token.Position) {
	visibility := 0 // private by default
	if err := a.typeRegistry.Register(name, typ, pos, visibility); err != nil {
		a.addError("failed to register type '%s' at %s: %v", name, pos, err)
	}
}

// registerBuiltinType registers a built-in type with public visibility.
// Uses the RegisterBuiltIn convenience method.
func (a *Analyzer) registerBuiltinType(name string, typ types.Type) {
	if err := a.typeRegistry.RegisterBuiltIn(name, typ); err != nil {
		a.addError("failed to register built-in type '%s': %v", name, err)
	}
}

// lookupType looks up a type by name (case-insensitive).
// Returns the type and true if found, nil and false otherwise.
func (a *Analyzer) lookupType(name string) (types.Type, bool) {
	return a.typeRegistry.Resolve(name)
}

// hasType checks if a type with the given name exists (case-insensitive).
func (a *Analyzer) hasType(name string) bool {
	return a.typeRegistry.Has(name)
}

// getClassType looks up a class type by name and returns it as *ClassType.
// Returns nil if not found or if the type is not a class.
func (a *Analyzer) getClassType(name string) *types.ClassType {
	if a.currentNestedTypes != nil {
		if qualified, ok := a.currentNestedTypes[ident.Normalize(name)]; ok {
			name = qualified
		}
	}

	typ, ok := a.typeRegistry.Resolve(name)
	if !ok {
		return nil
	}
	classType, ok := typ.(*types.ClassType)
	if !ok {
		return nil
	}
	return classType
}

// getInterfaceType looks up an interface type by name and returns it as *InterfaceType.
// Returns nil if not found or if the type is not an interface.
func (a *Analyzer) getInterfaceType(name string) *types.InterfaceType {
	typ, ok := a.typeRegistry.Resolve(name)
	if !ok {
		return nil
	}
	interfaceType, ok := typ.(*types.InterfaceType)
	if !ok {
		return nil
	}
	return interfaceType
}

// getEnumType looks up an enum type by name and returns it as *EnumType.
// Returns nil if not found or if the type is not an enum.
func (a *Analyzer) getEnumType(name string) *types.EnumType {
	typ, ok := a.typeRegistry.Resolve(name)
	if !ok {
		return nil
	}
	// Unwrap type aliases so aliases to enums behave like the enum itself.
	// This allows scoped access via an alias (e.g., TE2.Value when TE2 = TE1).
	if aliasType, ok := typ.(*types.TypeAlias); ok {
		typ = types.GetUnderlyingType(aliasType)
	}
	enumType, ok := typ.(*types.EnumType)
	if !ok {
		return nil
	}
	return enumType
}

// getRecordType looks up a record type by name and returns it as *RecordType.
// Returns nil if not found or if the type is not a record.
func (a *Analyzer) getRecordType(name string) *types.RecordType {
	typ, ok := a.typeRegistry.Resolve(name)
	if !ok {
		return nil
	}
	recordType, ok := typ.(*types.RecordType)
	if !ok {
		return nil
	}
	return recordType
}

// getSetType looks up a set type by name and returns it as *SetType.
// Returns nil if not found or if the type is not a set.
func (a *Analyzer) getSetType(name string) *types.SetType {
	typ, ok := a.typeRegistry.Resolve(name)
	if !ok {
		return nil
	}
	setType, ok := typ.(*types.SetType)
	if !ok {
		return nil
	}
	return setType
}

// getArrayType looks up an array type by name and returns it as *ArrayType.
// Returns nil if not found or if the type is not an array.
func (a *Analyzer) getArrayType(name string) *types.ArrayType {
	typ, ok := a.typeRegistry.Resolve(name)
	if !ok {
		return nil
	}
	arrayType, ok := typ.(*types.ArrayType)
	if !ok {
		return nil
	}
	return arrayType
}

// getTypeAlias looks up a type alias by name and returns it as *TypeAlias.
// Returns nil if not found or if the type is not an alias.
func (a *Analyzer) getTypeAlias(name string) *types.TypeAlias {
	typ, ok := a.typeRegistry.Resolve(name)
	if !ok {
		return nil
	}
	aliasType, ok := typ.(*types.TypeAlias)
	if !ok {
		return nil
	}
	return aliasType
}
