package semantic

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
	pkgast "github.com/cwbudde/go-dws/pkg/ast" // Task 9.18
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
	subranges          map[string]*types.SubrangeType
	functionPointers   map[string]*types.FunctionPointerType
	arrays             map[string]*types.ArrayType
	currentFunction    *ast.FunctionDecl
	classes            map[string]*types.ClassType
	currentClass       *types.ClassType
	currentRecord      *types.RecordType
	enums              map[string]*types.EnumType
	records            map[string]*types.RecordType
	sets               map[string]*types.SetType
	conversionRegistry *types.ConversionRegistry
	typeAliases        map[string]*types.TypeAlias
	interfaces         map[string]*types.InterfaceType
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
		semanticInfo:       pkgast.NewSemanticInfo(), // Task 9.18
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

// registerBuiltinExceptionTypes registers Exception and standard exception types
func (a *Analyzer) registerBuiltinExceptionTypes() {
	// Register TObject as the root base class for all classes
	// Required for DWScript compatibility
	objectClass := &types.ClassType{
		Name:                 "TObject",
		Parent:               nil, // Root of the class hierarchy
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

	// Add basic Create constructor
	objectClass.AddConstructorOverload("Create", &types.MethodInfo{
		Signature: &types.FunctionType{
			Parameters: []types.Type{}, // no parameters
			ReturnType: objectClass,
		},
		Visibility: int(ast.VisibilityPublic),
	})

	// Add ClassName method (returns the runtime type name)
	objectClass.Methods["ClassName"] = &types.FunctionType{
		Parameters: []types.Type{},
		ReturnType: types.STRING,
	}

	// Task 9.285: Use lowercase for case-insensitive lookup
	a.classes["tobject"] = objectClass

	// Register TClass as a type alias for "class of TObject"
	// TClass is a built-in metaclass type in DWScript
	tclassAlias := &types.TypeAlias{
		Name:        "TClass",
		AliasedType: types.NewClassOfType(objectClass),
	}
	a.typeAliases["tclass"] = tclassAlias

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
	// Task 9.285: Use lowercase for case-insensitive lookup
	exceptionClass.Fields["message"] = types.STRING

	// Add Create constructor
	exceptionClass.AddConstructorOverload("Create", &types.MethodInfo{
		Signature: &types.FunctionType{
			Parameters: []types.Type{types.STRING}, // message parameter
			ReturnType: exceptionClass,
		},
		Visibility: int(ast.VisibilityPublic),
	})

	// Task 9.285: Use lowercase for case-insensitive lookup
	a.classes["exception"] = exceptionClass

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

		// Task 9.285: Use lowercase for case-insensitive lookup
		a.classes[strings.ToLower(excName)] = excClass
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

	// Task 9.285: Use lowercase for case-insensitive lookup
	a.interfaces[strings.ToLower("IInterface")] = iinterface
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

	// Task 9.64: Validate that all forward-declared functions have implementations
	a.validateFunctionImplementations()

	// Task 9.11: Validate that all forward-declared classes have implementations
	a.validateClassForwardDeclarations()

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

// validateMethodImplementations checks that all forward-declared methods have implementations
// Task 9.284: Post-analysis validation for missing method implementations
func (a *Analyzer) validateMethodImplementations() {
	// Iterate through all classes
	// Note: className is the lowercase map key, use classType.Name for original case
	for _, classType := range a.classes {
		// Check each method and constructor in ForwardedMethods
		// Task 9.16.1: ForwardedMethods now uses lowercase keys
		for methodName, isForwarded := range classType.ForwardedMethods {
			if !isForwarded {
				continue
			}

			// Skip abstract methods - they don't need implementations
			// Task 9.16.1: AbstractMethods also uses lowercase keys now
			if classType.AbstractMethods[methodName] {
				continue
			}

			// Skip external methods - they are implemented externally
			if classType.IsExternal {
				continue
			}

			// Determine if this is a constructor or regular method
			// Task 9.16.1: ConstructorOverloads uses lowercase keys
			isConstructor := len(classType.ConstructorOverloads[methodName]) > 0

			// This method/constructor was declared but never implemented
			// Use classType.Name to preserve original case in error messages
			if isConstructor {
				a.addError("constructor '%s.%s' declared but not implemented",
					classType.Name, methodName)
			} else {
				a.addError("method '%s.%s' declared but not implemented",
					classType.Name, methodName)
			}
		}
	}
}

// validateFunctionImplementations checks that all forward-declared functions have implementations
// Task 9.64: Post-analysis validation for missing function implementations
func (a *Analyzer) validateFunctionImplementations() {
	// Walk through all symbols in the global scope and nested scopes
	a.validateFunctionImplementationsInScope(a.symbols)
}

// validateFunctionImplementationsInScope recursively checks a scope and its nested scopes
func (a *Analyzer) validateFunctionImplementationsInScope(scope *SymbolTable) {
	if scope == nil {
		return
	}

	// Check all symbols in this scope
	for _, symbol := range scope.symbols {
		// Check overload sets first (their Type is nil, so must check before type assertion)
		if symbol.IsOverloadSet {
			for _, overload := range symbol.Overloads {
				if overload.IsForward {
					// Get function type from the overload
					overloadFuncType, ok := overload.Type.(*types.FunctionType)
					if !ok {
						continue
					}

					// Format error message to match DWScript
					kind := "function"
					if overloadFuncType.ReturnType == nil || overloadFuncType.ReturnType.String() == "Void" {
						kind = "function" // DWScript uses "function" for both
					}
					a.addError("Syntax Error: The %s \"%s\" was forward declared but not implemented",
						kind, overload.Name)
				}
			}
			continue // Skip to next symbol (overload sets don't have individual Type)
		}

		// Check non-overload functions
		funcType, ok := symbol.Type.(*types.FunctionType)
		if !ok {
			continue // Not a function
		}

		// Check if this is a non-overloaded forward function
		if symbol.IsForward {
			// Format error message to match DWScript: "The function "X" was forward declared but not implemented"
			kind := "function"
			if funcType.ReturnType == nil || funcType.ReturnType.String() == "Void" {
				kind = "function" // DWScript uses "function" for both
			}
			// Note: We don't have position info in Symbol, so we'll use a simple error message
			a.addError("Syntax Error: The %s \"%s\" was forward declared but not implemented",
				kind, symbol.Name)
		}
	}

	// Recursively check nested scopes (parent scope)
	if scope.outer != nil {
		a.validateFunctionImplementationsInScope(scope.outer)
	}
}

// validateClassForwardDeclarations checks that all forward-declared classes have implementations
// Task 9.11: Post-analysis validation for missing class implementations
// Task 9.4: Updated error message to match DWScript format
func (a *Analyzer) validateClassForwardDeclarations() {
	// Iterate through all classes
	for _, classType := range a.classes {
		// Check if this class is still marked as forward (not implemented)
		if classType.IsForward {
			// This class was forward declared but never implemented
			// Use DWScript format: Class "Name" isn't defined completely
			a.addError("Class \"%s\" isn't defined completely", classType.Name)
		}
	}
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
// Symbol Table Accessors (Task 10.15)
// ============================================================================

// GetSymbolTable returns the current symbol table.
// This is used for extracting symbol information for LSP features.
func (a *Analyzer) GetSymbolTable() *SymbolTable {
	return a.symbols
}

// GetClasses returns the analyzer's class type map.
func (a *Analyzer) GetClasses() map[string]*types.ClassType {
	return a.classes
}

// GetInterfaces returns the analyzer's interface type map.
func (a *Analyzer) GetInterfaces() map[string]*types.InterfaceType {
	return a.interfaces
}

// GetEnums returns the analyzer's enum type map.
func (a *Analyzer) GetEnums() map[string]*types.EnumType {
	return a.enums
}

// GetRecords returns the analyzer's record type map.
func (a *Analyzer) GetRecords() map[string]*types.RecordType {
	return a.records
}

// GetArrayTypes returns the analyzer's array type map.
func (a *Analyzer) GetArrayTypes() map[string]*types.ArrayType {
	return a.arrays
}

// GetTypeAliases returns the analyzer's type alias map.
func (a *Analyzer) GetTypeAliases() map[string]*types.TypeAlias {
	return a.typeAliases
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
