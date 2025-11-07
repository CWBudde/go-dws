package semantic

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
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
	arrays             map[string]*types.ArrayType
	typeAliases        map[string]*types.TypeAlias
	subranges          map[string]*types.SubrangeType
	functionPointers   map[string]*types.FunctionPointerType // Task 9.159: Function pointer types
	helpers            map[string][]*types.HelperType        // Task 9.82: Helper types (type name -> list of helpers)
	currentFunction    *ast.FunctionDecl
	classes            map[string]*types.ClassType
	interfaces         map[string]*types.InterfaceType
	enums              map[string]*types.EnumType
	records            map[string]*types.RecordType
	sets               map[string]*types.SetType
	conversionRegistry *types.ConversionRegistry
	currentClass       *types.ClassType
	symbols            *SymbolTable
	unitSymbols        map[string]*SymbolTable // Unit name -> symbol table for qualified access
	globalOperators    *types.OperatorRegistry
	errors             []string
	structuredErrors   []*SemanticError // Task 9.110: Structured errors with rich context
	sourceCode         string           // Task 9.110: Source code for error display
	sourceFile         string           // Task 9.110: Source filename for error display
	loopDepth          int
	inExceptionHandler bool
	inFinallyBlock     bool
	inLoop             bool
	inLambda           bool // Task 9.216: Track if we're analyzing a lambda body
	inClassMethod      bool // Track if we're analyzing a class method (static method)
}

// NewAnalyzer creates a new semantic analyzer
func NewAnalyzer() *Analyzer {
	a := &Analyzer{
		symbols:            NewSymbolTable(),
		unitSymbols:        make(map[string]*SymbolTable),
		errors:             make([]string, 0),
		structuredErrors:   make([]*SemanticError, 0), // Task 9.110
		classes:            make(map[string]*types.ClassType),
		interfaces:         make(map[string]*types.InterfaceType),
		enums:              make(map[string]*types.EnumType),
		records:            make(map[string]*types.RecordType),
		sets:               make(map[string]*types.SetType),
		arrays:             make(map[string]*types.ArrayType),
		typeAliases:        make(map[string]*types.TypeAlias),
		subranges:          make(map[string]*types.SubrangeType),
		functionPointers:   make(map[string]*types.FunctionPointerType), // Task 9.159
		helpers:            make(map[string][]*types.HelperType),        // Task 9.82
		globalOperators:    types.NewOperatorRegistry(),
		conversionRegistry: types.NewConversionRegistry(),
	}

	// Task 8.203: Register built-in Exception base class
	a.registerBuiltinExceptionTypes()

	// Task 9.171: Register built-in array helpers
	a.initArrayHelpers()

	// Task 9.205: Register built-in helpers for primitive types
	a.initIntrinsicHelpers()

	// Task 9.31: Register built-in enum helpers
	a.initEnumHelpers()

	return a
}

// registerBuiltinExceptionTypes registers Exception and standard exception types
// Task 8.203-8.204
func (a *Analyzer) registerBuiltinExceptionTypes() {
	// Register TObject as the root base class for all classes
	// Required for DWScript compatibility
	objectClass := &types.ClassType{
		Name:             "TObject",
		Parent:           nil, // Root of the class hierarchy
		Fields:           make(map[string]types.Type),
		Methods:          make(map[string]*types.FunctionType),
		FieldVisibility:  make(map[string]int),
		MethodVisibility: make(map[string]int),
		VirtualMethods:   make(map[string]bool),
		OverrideMethods:  make(map[string]bool),
		AbstractMethods:  make(map[string]bool),
		Constructors:     make(map[string]*types.FunctionType),
		Interfaces:       make([]*types.InterfaceType, 0),
		Properties:       make(map[string]*types.PropertyInfo),
		ClassMethodFlags: make(map[string]bool),
	}

	// Add basic Create constructor
	objectClass.Constructors["Create"] = &types.FunctionType{
		Parameters: []types.Type{}, // no parameters
		ReturnType: objectClass,
	}

	// Add ClassName method (returns the runtime type name)
	objectClass.Methods["ClassName"] = &types.FunctionType{
		Parameters: []types.Type{},
		ReturnType: types.STRING,
	}

	// Task 9.285: Use lowercase for case-insensitive lookup
	a.classes["tobject"] = objectClass

	// Task 8.203: Define Exception base class
	exceptionClass := &types.ClassType{
		Name:             "Exception",
		Parent:           objectClass, // Exception inherits from TObject
		Fields:           make(map[string]types.Type),
		Methods:          make(map[string]*types.FunctionType),
		FieldVisibility:  make(map[string]int),
		MethodVisibility: make(map[string]int),
		VirtualMethods:   make(map[string]bool),
		OverrideMethods:  make(map[string]bool),
		AbstractMethods:  make(map[string]bool),
		Constructors:     make(map[string]*types.FunctionType),
		Interfaces:       make([]*types.InterfaceType, 0),
		Properties:       make(map[string]*types.PropertyInfo),
		ClassMethodFlags: make(map[string]bool),
	}

	// Add Message field to Exception
	exceptionClass.Fields["Message"] = types.STRING

	// Add Create constructor
	exceptionClass.Constructors["Create"] = &types.FunctionType{
		Parameters: []types.Type{types.STRING}, // message parameter
		ReturnType: exceptionClass,
	}

	// Task 9.285: Use lowercase for case-insensitive lookup
	a.classes["exception"] = exceptionClass

	// Task 8.204: Define standard exception types
	standardExceptions := []string{
		"EConvertError",    // Type conversion failures
		"ERangeError",      // Array bounds, invalid ranges
		"EDivByZero",       // Division by zero
		"EAssertionFailed", // Failed assertions
		"EInvalidOp",       // Invalid operations
	}

	for _, excName := range standardExceptions {
		excClass := &types.ClassType{
			Name:             excName,
			Parent:           exceptionClass, // All inherit from Exception
			Fields:           make(map[string]types.Type),
			Methods:          make(map[string]*types.FunctionType),
			FieldVisibility:  make(map[string]int),
			MethodVisibility: make(map[string]int),
			VirtualMethods:   make(map[string]bool),
			OverrideMethods:  make(map[string]bool),
			AbstractMethods:  make(map[string]bool),
			Constructors:     make(map[string]*types.FunctionType),
			Interfaces:       make([]*types.InterfaceType, 0),
			Properties:       make(map[string]*types.PropertyInfo),
			ClassMethodFlags: make(map[string]bool),
		}

		// Inherit Message field from Exception
		excClass.Fields["Message"] = types.STRING

		// Inherit Create constructor
		excClass.Constructors["Create"] = &types.FunctionType{
			Parameters: []types.Type{types.STRING},
			ReturnType: excClass,
		}

		// Task 9.285: Use lowercase for case-insensitive lookup
		a.classes[strings.ToLower(excName)] = excClass
	}
}

// Analyze performs semantic analysis on a program.
// Returns nil if analysis succeeds, or an error if there are semantic errors.
func (a *Analyzer) Analyze(program *ast.Program) error {
	if program == nil {
		return fmt.Errorf("cannot analyze nil program")
	}

	// Analyze each statement in the program
	for _, stmt := range program.Statements {
		a.analyzeStatement(stmt)
	}

	// Task 9.284: Validate that all forward-declared methods have implementations
	a.validateMethodImplementations()

	// If we accumulated errors, return them
	if len(a.errors) > 0 {
		return &AnalysisError{Errors: a.errors}
	}

	return nil
}

// validateMethodImplementations checks that all forward-declared methods have implementations
// Task 9.284: Post-analysis validation for missing method implementations
func (a *Analyzer) validateMethodImplementations() {
	// Iterate through all classes
	// Note: className is the lowercase map key, use classType.Name for original case
	for _, classType := range a.classes {
		// Check each method in the class
		for methodName := range classType.Methods {
			// Check if method is still marked as forward (not implemented)
			if classType.ForwardedMethods[methodName] {
				// Skip abstract methods - they don't need implementations
				if classType.AbstractMethods[methodName] {
					continue
				}

				// Skip external methods - they are implemented externally
				if classType.IsExternal {
					continue
				}

				// This method was declared but never implemented
				// Use classType.Name to preserve original case in error messages
				a.addError("method '%s.%s' declared but not implemented",
					classType.Name, methodName)
			}
		}

		// Also check constructors
		for ctorName := range classType.Constructors {
			// Check if constructor is still marked as forward (not implemented)
			if classType.ForwardedMethods[ctorName] {
				// Skip if class is external
				if classType.IsExternal {
					continue
				}

				// This constructor was declared but never implemented
				// Use classType.Name to preserve original case in error messages
				a.addError("constructor '%s.%s' declared but not implemented",
					classType.Name, ctorName)
			}
		}
	}
}

// Errors returns all accumulated semantic errors
func (a *Analyzer) Errors() []string {
	return a.errors
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
	if fromClass, ok := from.(*types.ClassType); ok {
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
