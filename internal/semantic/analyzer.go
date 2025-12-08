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

// HintsLevel mirrors DWScript's hint levels for compatibility.
type HintsLevel int

const (
	HintsLevelDisabled HintsLevel = iota
	HintsLevelNormal
	HintsLevelStrict
	HintsLevelPedantic
)

// LoopExitability represents whether a loop can exit normally
type LoopExitability int

const (
	LoopNotExitable LoopExitability = iota // Loop has no exit mechanism (infinite loop)
	LoopExitBreak                          // Loop can exit via Break
	LoopExitExit                           // Loop can exit via Exit (or has other exits)
)

// Analyzer performs semantic analysis on a DWScript program.
type Analyzer struct {
	currentClass         *types.ClassType                      // Current class being analyzed
	helpers              map[string][]*types.HelperType        // Helper type registry
	typeRegistry         *TypeRegistry                         // Unified type registry
	subranges            map[string]*types.SubrangeType        // Subrange type registry
	functionPointers     map[string]*types.FunctionPointerType // Function pointer type registry
	currentFunction      *ast.FunctionDecl                     // Current function being analyzed
	currentRecord        *types.RecordType                     // Current record being analyzed
	symbols              *SymbolTable                          // Symbol table
	globalOperators      *types.OperatorRegistry               // Operator overload registry
	conversionRegistry   *types.ConversionRegistry             // Type conversion registry
	semanticInfo         *pkgast.SemanticInfo                  // AST annotations
	unitSymbols          map[string]*SymbolTable               // Unit symbol tables
	currentNestedTypes   map[string]string                     // Nested type tracking
	nestedTypeAliases    map[string]map[string]string          // Nested type aliases
	currentProperty      string                                // Current property being analyzed
	sourceFile           string                                // Source file path
	sourceCode           string                                // Original source text
	loopPosStack         []token.Position                      // Stack tracking loop positions for warnings
	structuredErrors     []*SemanticError                      // Structured error objects
	loopExitabilityStack []LoopExitability                     // Stack tracking loop exitability
	errors               []string                              // Error messages (legacy)
	loopDepth            int                                   // Loop nesting level
	hintsLevel           HintsLevel                            // Hints emission level
	inLoop               bool                                  // Inside loop construct
	inLambda             bool                                  // Inside lambda/anonymous function
	inClassMethod        bool                                  // Inside class method
	inPropertyExpr       bool                                  // Inside property expression
	inFinallyBlock       bool                                  // Inside finally block
	experimentalPasses   bool                                  // Enable experimental passes
	inExceptionHandler   bool                                  // Inside try/except block
}

// NewAnalyzer creates a new semantic analyzer
func NewAnalyzer() *Analyzer {
	a := &Analyzer{
		symbols:            NewSymbolTable(),
		typeRegistry:       NewTypeRegistry(),
		unitSymbols:        make(map[string]*SymbolTable),
		errors:             make([]string, 0),
		structuredErrors:   make([]*SemanticError, 0),
		subranges:          make(map[string]*types.SubrangeType),
		functionPointers:   make(map[string]*types.FunctionPointerType),
		helpers:            make(map[string][]*types.HelperType),
		globalOperators:    types.NewOperatorRegistry(),
		conversionRegistry: types.NewConversionRegistry(),
		semanticInfo:       pkgast.NewSemanticInfo(),
		nestedTypeAliases:  make(map[string]map[string]string),
		hintsLevel:         HintsLevelPedantic,
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

	// Register mathematical constants
	a.symbols.DefineConst("Pi", types.FLOAT, math.Pi)

	// Register Variant special values
	a.symbols.DefineConst("Null", types.VARIANT, nil)
	a.symbols.DefineConst("Unassigned", types.VARIANT, nil)

	return a
}

// NewAnalyzerWithExperimentalPasses creates a new semantic analyzer with experimental
// multi-pass analysis enabled. Use NewAnalyzer() for normal usage.
func NewAnalyzerWithExperimentalPasses() *Analyzer {
	a := NewAnalyzer()
	a.experimentalPasses = true
	return a
}

// registerBuiltinExceptionTypes registers Exception and standard exception types
func (a *Analyzer) registerBuiltinExceptionTypes() {
	// TObject is the root base class for all classes
	objectClass := types.NewClassType("TObject", nil)

	objectClass.AddConstructorOverload("Create", &types.MethodInfo{
		Signature: &types.FunctionType{
			Parameters: []types.Type{},
			ReturnType: objectClass,
		},
		Visibility: int(ast.VisibilityPublic),
	})

	objectClass.AddMethodOverload("Destroy", &types.MethodInfo{
		Signature: &types.FunctionType{
			Parameters: []types.Type{},
			ReturnType: types.VOID,
		},
		IsVirtual:  true,
		Visibility: int(ast.VisibilityPublic),
	})

	objectClass.AddMethodOverload("Free", &types.MethodInfo{
		Signature: &types.FunctionType{
			Parameters: []types.Type{},
			ReturnType: types.VOID,
		},
		Visibility: int(ast.VisibilityPublic),
	})

	objectClass.VirtualMethods[ident.Normalize("Destroy")] = true
	objectClass.MethodVisibility[ident.Normalize("Destroy")] = int(ast.VisibilityPublic)
	objectClass.MethodVisibility[ident.Normalize("Free")] = int(ast.VisibilityPublic)

	objectClass.Methods["ClassName"] = &types.FunctionType{
		Parameters: []types.Type{},
		ReturnType: types.STRING,
	}

	a.registerBuiltinType("TObject", objectClass)

	// TClass is the built-in metaclass type
	tclassAlias := &types.TypeAlias{
		Name:        "TClass",
		AliasedType: types.NewClassOfType(objectClass),
	}
	a.registerBuiltinType("TClass", tclassAlias)

	exceptionClass := &types.ClassType{
		Name:                 "Exception",
		Parent:               objectClass,
		Fields:               make(map[string]types.Type),
		Methods:              make(map[string]*types.FunctionType),
		FieldVisibility:      make(map[string]int),
		MethodVisibility:     make(map[string]int),
		VirtualMethods:       make(map[string]bool),
		OverrideMethods:      make(map[string]bool),
		AbstractMethods:      make(map[string]bool),
		ReintroduceMethods:   make(map[string]bool),
		Constructors:         make(map[string]*types.FunctionType),
		ConstructorOverloads: make(map[string][]*types.MethodInfo),
		Interfaces:           make([]*types.InterfaceType, 0),
		Properties:           make(map[string]*types.PropertyInfo),
		ClassMethodFlags:     make(map[string]bool),
	}

	exceptionClass.Fields["Message"] = types.STRING

	exceptionClass.AddConstructorOverload("Create", &types.MethodInfo{
		Signature: &types.FunctionType{
			Parameters: []types.Type{types.STRING},
			ReturnType: exceptionClass,
		},
		Visibility: int(ast.VisibilityPublic),
	})

	a.registerBuiltinType("Exception", exceptionClass)

	// Standard exception types
	standardExceptions := []string{
		"EConvertError",
		"ERangeError",
		"EDivByZero",
		"EAssertionFailed",
		"EInvalidOp",
	}

	for _, excName := range standardExceptions {
		excClass := &types.ClassType{
			Name:                 excName,
			Parent:               exceptionClass,
			Fields:               make(map[string]types.Type),
			Methods:              make(map[string]*types.FunctionType),
			FieldVisibility:      make(map[string]int),
			MethodVisibility:     make(map[string]int),
			VirtualMethods:       make(map[string]bool),
			OverrideMethods:      make(map[string]bool),
			AbstractMethods:      make(map[string]bool),
			ReintroduceMethods:   make(map[string]bool),
			Constructors:         make(map[string]*types.FunctionType),
			ConstructorOverloads: make(map[string][]*types.MethodInfo),
			Interfaces:           make([]*types.InterfaceType, 0),
			Properties:           make(map[string]*types.PropertyInfo),
			ClassMethodFlags:     make(map[string]bool),
		}

		excClass.Fields["Message"] = types.STRING

		excClass.AddConstructorOverload("Create", &types.MethodInfo{
			Signature: &types.FunctionType{
				Parameters: []types.Type{types.STRING},
				ReturnType: excClass,
			},
			Visibility: int(ast.VisibilityPublic),
		})

		a.registerBuiltinType(excName, excClass)
	}
}

// registerBuiltinInterfaces registers IInterface, the root interface type.
func (a *Analyzer) registerBuiltinInterfaces() {
	// IInterface is a marker interface with no methods
	iinterface := &types.InterfaceType{
		Name:    "IInterface",
		Parent:  nil,
		Methods: make(map[string]*types.FunctionType),
	}

	a.registerBuiltinType("IInterface", iinterface)
}

// Analyze performs semantic analysis on a program.
func (a *Analyzer) Analyze(program *ast.Program) error {
	if program == nil {
		return fmt.Errorf("cannot analyze nil program")
	}

	// Analyze each statement
	for _, stmt := range program.Statements {
		a.analyzeStatement(stmt)
	}

	a.validateForwardDeclarations()

	// Run experimental multi-pass analysis if enabled
	if a.experimentalPasses {
		ctx := a.createPassContext()

		pass2 := NewTypeResolutionPass()
		if err := pass2.Run(program, ctx); err != nil {
			return err
		}

		pass3 := NewValidationPass()
		if err := pass3.Run(program, ctx); err != nil {
			return err
		}

		a.syncFromPassContext(ctx)
		a.mergePassErrors(ctx)
	}

	// Return errors if any (hints and warnings don't prevent success)
	hasActualErrors := false
	for _, err := range a.errors {
		if !strings.HasPrefix(err, "Hint:") && !strings.HasPrefix(err, "Warning:") {
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
func (a *Analyzer) validateForwardDeclarations() {
	for _, t := range a.typeRegistry.AllTypes() {
		if classType, ok := t.(*types.ClassType); ok {
			if classType.IsForward {
				// Treat short-form empty classes (class with parent and no body) as complete.
				if classType.Parent != nil {
					classType.IsForward = false
					continue
				}
				a.addError("Class \"%s\" isn't defined completely", classType.Name)
			}
		}
	}
}

// createPassContext creates a PassContext from the Analyzer's current state.
func (a *Analyzer) createPassContext() *PassContext {
	return &PassContext{
		Symbols:            a.symbols,
		TypeRegistry:       a.typeRegistry,
		GlobalOperators:    a.globalOperators,
		ConversionRegistry: a.conversionRegistry,
		SemanticInfo:       a.semanticInfo,
		BuiltinChecker:     a,
		UnitSymbols:        a.unitSymbols,
		Helpers:            a.helpers,
		Subranges:          a.subranges,
		FunctionPointers:   a.functionPointers,
		Errors:             make([]string, 0),
		StructuredErrors:   make([]*SemanticError, 0),
		ScopeStack:         []*Scope{NewScope(ScopeGlobal, nil)},
		CurrentFunction:    a.currentFunction,
		CurrentClass:       a.currentClass,
		CurrentRecord:      a.currentRecord,
		CurrentProperty:    a.currentProperty,
		SourceCode:         a.sourceCode,
		SourceFile:         a.sourceFile,
		LoopDepth:          a.loopDepth,
		InExceptionHandler: a.inExceptionHandler,
		InFinallyBlock:     a.inFinallyBlock,
		InLoop:             a.inLoop,
		InLambda:           a.inLambda,
		InClassMethod:      a.inClassMethod,
		InPropertyExpr:     a.inPropertyExpr,
	}
}

// AnalyzeBuiltin implements the BuiltinChecker interface.
func (a *Analyzer) AnalyzeBuiltin(name string, args []ast.Expression, callExpr *ast.CallExpression) (types.Type, bool) {
	return a.analyzeBuiltinFunction(name, args, callExpr)
}

// IsBuiltinFunction implements the BuiltinChecker interface.
func (a *Analyzer) IsBuiltinFunction(name string) (types.Type, bool) {
	return a.getBuiltinReturnType(name)
}

// syncFromPassContext syncs state from PassContext back to Analyzer.
func (a *Analyzer) syncFromPassContext(ctx *PassContext) {
	if funcDecl, ok := ctx.CurrentFunction.(*ast.FunctionDecl); ok {
		a.currentFunction = funcDecl
	}
	a.currentClass = ctx.CurrentClass
	a.currentRecord = ctx.CurrentRecord
	a.currentProperty = ctx.CurrentProperty
	a.loopDepth = ctx.LoopDepth
	a.inExceptionHandler = ctx.InExceptionHandler
	a.inFinallyBlock = ctx.InFinallyBlock
	a.inLoop = ctx.InLoop
	a.inLambda = ctx.InLambda
	a.inClassMethod = ctx.InClassMethod
	a.inPropertyExpr = ctx.InPropertyExpr
}

// mergePassErrors merges errors from PassContext into Analyzer.
func (a *Analyzer) mergePassErrors(ctx *PassContext) {
	a.errors = append(a.errors, ctx.Errors...)
	a.structuredErrors = append(a.structuredErrors, ctx.StructuredErrors...)
}

// Errors returns all accumulated semantic errors
func (a *Analyzer) Errors() []string {
	return a.errors
}

// GetSemanticInfo returns the semantic metadata table.
func (a *Analyzer) GetSemanticInfo() *pkgast.SemanticInfo {
	return a.semanticInfo
}

// GetHelpers returns all registered helper types.
func (a *Analyzer) GetHelpers() map[string][]*types.HelperType {
	return a.helpers
}

// StructuredErrors returns all accumulated structured semantic errors.
func (a *Analyzer) StructuredErrors() []*SemanticError {
	return a.structuredErrors
}

// SetSource sets the source code and filename for error display.
func (a *Analyzer) SetSource(source, filename string) {
	a.sourceCode = source
	a.sourceFile = filename
}

// SetHintsLevel configures which hints should be emitted.
func (a *Analyzer) SetHintsLevel(level HintsLevel) {
	a.hintsLevel = level
}

func (a *Analyzer) addError(format string, args ...any) {
	a.errors = append(a.errors, fmt.Sprintf(format, args...))
}

func (a *Analyzer) addHint(format string, args ...any) {
	a.errors = append(a.errors, fmt.Sprintf("Hint: "+format, args...))
}

func (a *Analyzer) addWarning(format string, args ...any) {
	a.errors = append(a.errors, fmt.Sprintf("Warning: "+format, args...))
}

func (a *Analyzer) addCaseMismatchHint(actual, declared string, pos token.Position) {
	if a.hintsLevel < HintsLevelPedantic {
		return
	}
	a.addHint("\"%s\" does not match case of declaration (\"%s\") [line: %d, column: %d]",
		actual, declared, pos.Line, pos.Column)
}

func (a *Analyzer) addStructuredError(err *SemanticError) {
	a.structuredErrors = append(a.structuredErrors, err)
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

	// Allow nil assignment to/from class, interface, and metaclass types
	if from.TypeKind() == "NIL" && to.TypeKind() == "CLASS" {
		return true
	}
	if from.TypeKind() == "CLASS" && to.TypeKind() == "NIL" {
		return true
	}
	if from.TypeKind() == "NIL" && to.TypeKind() == "INTERFACE" {
		return true
	}
	if from.TypeKind() == "INTERFACE" && to.TypeKind() == "NIL" {
		return true
	}
	if from.TypeKind() == "NIL" && to.TypeKind() == "CLASSOF" {
		return true
	}
	if from.TypeKind() == "CLASSOF" && to.TypeKind() == "NIL" {
		return true
	}

	// Handle metaclass to metaclass assignment
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

	// Handle class type assignments
	if fromClass, ok := from.(*types.ClassType); ok {
		if toMetaclass, ok := to.(*types.ClassOfType); ok {
			if fromClass.Equals(toMetaclass.ClassType) || a.isDescendantOf(fromClass, toMetaclass.ClassType) {
				return true
			}
		}
		if toClass, ok := to.(*types.ClassType); ok {
			if fromClass.Equals(toClass) || a.isDescendantOf(fromClass, toClass) {
				return true
			}
		}
		if toInterface, ok := to.(*types.InterfaceType); ok {
			if fromClass.ImplementsInterface(toInterface) {
				return true
			}
		}
	}

	// Handle interface assignments
	if fromInterface, ok := from.(*types.InterfaceType); ok {
		if toInterface, ok := to.(*types.InterfaceType); ok {
			if fromInterface.Equals(toInterface) || fromInterface.InheritsFrom(toInterface) {
				return true
			}
		}
	}

	// Handle subrange type assignments
	if fromSubrange, ok := from.(*types.SubrangeType); ok {
		if fromSubrange.BaseType.Equals(to) {
			return true
		}
	}
	if toSubrange, ok := to.(*types.SubrangeType); ok {
		if toSubrange.BaseType.Equals(from) {
			return true
		}
	}

	toUnderlying := types.GetUnderlyingType(to)
	fromUnderlying := types.GetUnderlyingType(from)

	// Variant assignment rules
	if toUnderlying.TypeKind() == "VARIANT" {
		return true
	}
	if fromUnderlying.TypeKind() == "VARIANT" && toUnderlying.TypeKind() == "VARIANT" {
		return true
	}
	if fromUnderlying.TypeKind() == "VARIANT" {
		return true
	}

	// Allow implicit enum-to-integer conversion
	if fromUnderlying.TypeKind() == "ENUM" && toUnderlying.Equals(types.INTEGER) {
		return true
	}

	// Handle function and method pointer assignments
	if toUnderlying.TypeKind() == "FUNCTION_POINTER" || toUnderlying.TypeKind() == "METHOD_POINTER" {
		if fromMethodPtr, ok := fromUnderlying.(*types.MethodPointerType); ok {
			return fromMethodPtr.IsCompatibleWith(toUnderlying)
		}
		if toFuncPtr, ok := toUnderlying.(*types.FunctionPointerType); ok {
			// First try exact compatibility
			if toFuncPtr.IsCompatibleWith(fromUnderlying) {
				return true
			}
			// Task: For helper methods like Map that use Variant parameters,
			// allow function pointers with compatible concrete types.
			// E.g., function(Integer): String should be assignable to function(Variant): Variant
			if fromFuncPtr, ok := fromUnderlying.(*types.FunctionPointerType); ok {
				if a.isFunctionPointerVariantCompatible(fromFuncPtr, toFuncPtr) {
					return true
				}
			}
			return false
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

// isFunctionPointerVariantCompatible checks if a function pointer with concrete types
// can be assigned to a function pointer that uses Variant parameters/return type.
// This enables higher-order functions like Map, Filter, etc. to accept functions
// with specific types when the helper declares Variant-based signatures.
//
// Rules:
// - Parameter count must match
// - Target must have at least one Variant parameter OR Variant return type
// - Each target Variant parameter accepts any concrete type from source
// - Non-Variant parameters must match exactly (use Equals, not assignability)
// - Return types: Variant target accepts any source, otherwise must match exactly
func (a *Analyzer) isFunctionPointerVariantCompatible(from, to *types.FunctionPointerType) bool {
	// Parameter count must match
	if len(from.Parameters) != len(to.Parameters) {
		return false
	}

	// Track whether target has ANY Variant usage (parameter or return)
	// This function is specifically for Variant-based compatibility, not general assignability
	hasVariantUsage := false

	// Check each parameter
	for i := range from.Parameters {
		toParam := to.Parameters[i]
		fromParam := from.Parameters[i]

		// If target expects Variant, any concrete type is acceptable
		if toParam.Equals(types.VARIANT) {
			hasVariantUsage = true
			continue
		}

		// Otherwise, types must match exactly (not just be assignable)
		// This prevents Integer→Float implicit conversion from making function pointers compatible
		if !fromParam.Equals(toParam) {
			return false
		}
	}

	// Check return type compatibility
	// If target returns Variant, any concrete return type is acceptable
	if to.ReturnType != nil && to.ReturnType.Equals(types.VARIANT) {
		hasVariantUsage = true
		return true
	}

	// Both have nil return type (procedures)
	if from.ReturnType == nil && to.ReturnType == nil {
		// Only allow if there was Variant usage in parameters
		return hasVariantUsage
	}

	// One has return type, the other doesn't
	if from.ReturnType == nil || to.ReturnType == nil {
		return false
	}

	// Both have return types - must match exactly (not just be assignable)
	if !from.ReturnType.Equals(to.ReturnType) {
		return false
	}

	// Only return true if there was Variant usage somewhere
	return hasVariantUsage
}

// ============================================================================
// Symbol Table Accessors
// ============================================================================

// GetSymbolTable returns the current symbol table.
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

// areArrayTypesCompatibleForVarParam checks if an array type can be passed to a var parameter.
func (a *Analyzer) areArrayTypesCompatibleForVarParam(argType, paramType types.Type) bool {
	argArray, argIsArray := argType.(*types.ArrayType)
	paramArray, paramIsArray := paramType.(*types.ArrayType)

	if !argIsArray || !paramIsArray {
		return false
	}

	if !types.IsCompatible(argArray.ElementType, paramArray.ElementType) {
		return false
	}

	return true
}

// validateFieldInitializer validates field initializer type compatibility.
func (a *Analyzer) validateFieldInitializer(field *ast.FieldDecl, fieldName string, fieldType types.Type) {
	if field.InitValue != nil {
		initType := a.analyzeExpression(field.InitValue)
		if initType != nil && fieldType != nil {
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

// registerType registers a user-defined type with the registry.
func (a *Analyzer) registerType(name string, typ types.Type) {
	pos := token.Position{Line: 0, Column: 0, Offset: 0}
	visibility := 0
	if err := a.typeRegistry.Register(name, typ, pos, visibility); err != nil {
		a.addError("failed to register type '%s': %v", name, err)
	}
}

// registerTypeWithPos registers a type with explicit position information.
func (a *Analyzer) registerTypeWithPos(name string, typ types.Type, pos token.Position) {
	visibility := 0
	if err := a.typeRegistry.Register(name, typ, pos, visibility); err != nil {
		a.addError("failed to register type '%s' at %s: %v", name, pos, err)
	}
}

// registerBuiltinType registers a built-in type with public visibility.
func (a *Analyzer) registerBuiltinType(name string, typ types.Type) {
	if err := a.typeRegistry.RegisterBuiltIn(name, typ); err != nil {
		a.addError("failed to register built-in type '%s': %v", name, err)
	}
}

func (a *Analyzer) lookupType(name string) (types.Type, bool) {
	return a.typeRegistry.Resolve(name)
}

func (a *Analyzer) hasType(name string) bool {
	return a.typeRegistry.Has(name)
}

// getClassType looks up a class type by name.
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

func (a *Analyzer) getInterfaceType(name string) *types.InterfaceType {
	typ, ok := a.typeRegistry.Resolve(name)
	if !ok {
		return nil
	}
	interfaceType, _ := typ.(*types.InterfaceType)
	return interfaceType
}

func (a *Analyzer) getEnumType(name string) *types.EnumType {
	typ, ok := a.typeRegistry.Resolve(name)
	if !ok {
		return nil
	}
	// Unwrap type aliases for scoped access (e.g., TE2.Value when TE2 = TE1)
	if aliasType, ok := typ.(*types.TypeAlias); ok {
		typ = types.GetUnderlyingType(aliasType)
	}
	enumType, _ := typ.(*types.EnumType)
	return enumType
}

func (a *Analyzer) getRecordType(name string) *types.RecordType {
	typ, ok := a.typeRegistry.Resolve(name)
	if !ok {
		return nil
	}
	recordType, _ := typ.(*types.RecordType)
	return recordType
}

func (a *Analyzer) getSetType(name string) *types.SetType {
	typ, ok := a.typeRegistry.Resolve(name)
	if !ok {
		return nil
	}
	setType, _ := typ.(*types.SetType)
	return setType
}

func (a *Analyzer) getArrayType(name string) *types.ArrayType {
	typ, ok := a.typeRegistry.Resolve(name)
	if !ok {
		return nil
	}
	arrayType, _ := typ.(*types.ArrayType)
	return arrayType
}

// ============================================================================
// Infinite Loop Detection
// ============================================================================

// enterLoop marks the start of a loop and pushes it onto the exitability stack
func (a *Analyzer) enterLoop(pos token.Position) {
	a.loopExitabilityStack = append(a.loopExitabilityStack, LoopNotExitable)
	a.loopPosStack = append(a.loopPosStack, pos)
}

// markLoopExitable marks the current loop(s) as exitable
func (a *Analyzer) markLoopExitable(level LoopExitability) {
	if len(a.loopExitabilityStack) == 0 {
		return
	}

	switch level {
	case LoopExitBreak:
		// Break only marks the current loop as exitable
		if a.loopExitabilityStack[len(a.loopExitabilityStack)-1] == LoopNotExitable {
			a.loopExitabilityStack[len(a.loopExitabilityStack)-1] = level
		}
	case LoopExitExit:
		// Exit marks all loops in the stack as exitable
		for i := range a.loopExitabilityStack {
			if a.loopExitabilityStack[i] != LoopExitExit {
				a.loopExitabilityStack[i] = level
			}
		}
	}
}

// leaveLoop pops the loop from the exitability stack and emits warning if infinite
func (a *Analyzer) leaveLoop() {
	if len(a.loopExitabilityStack) == 0 {
		return
	}

	// Check if the loop is still marked as not exitable
	idx := len(a.loopExitabilityStack) - 1
	if a.loopExitabilityStack[idx] == LoopNotExitable {
		// Emit infinite loop warning in DWScript format
		pos := a.loopPosStack[idx]
		a.addWarning("Infinite loop [line: %d, column: %d]", pos.Line, pos.Column)
	}

	// Pop from stacks
	a.loopExitabilityStack = a.loopExitabilityStack[:idx]
	a.loopPosStack = a.loopPosStack[:idx]
}
