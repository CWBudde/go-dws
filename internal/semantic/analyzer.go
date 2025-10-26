package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// Analyzer performs semantic analysis on a DWScript program.
// It validates types, checks for undefined variables, and ensures
// type compatibility in expressions and statements.
type Analyzer struct {
	// Symbol table for tracking variables and functions
	symbols *SymbolTable

	// Accumulated errors during analysis
	errors []string

	// Current function being analyzed (for return type checking)
	currentFunction *ast.FunctionDecl

	// Class registry for tracking declared classes (Task 7.54)
	classes map[string]*types.ClassType

	// Interface registry for tracking declared interfaces (Task 7.97)
	interfaces map[string]*types.InterfaceType

	// Enum registry for tracking declared enums (Task 8.43)
	enums map[string]*types.EnumType

	// Record registry for tracking declared records (Task 8.68)
	records map[string]*types.RecordType

	// Set registry for tracking declared sets (Task 8.99)
	sets map[string]*types.SetType

	// Array registry for tracking declared arrays (Task 8.126)
	arrays map[string]*types.ArrayType

	// Current class being analyzed (for field/method access)
	currentClass *types.ClassType

	// Operator registries (Stage 8)
	globalOperators    *types.OperatorRegistry
	conversionRegistry *types.ConversionRegistry

	// Exception handling context tracking (Task 8.208, 8.209)
	inExceptionHandler bool // Track if we're inside an exception handler (for bare raise validation)
	inFinallyBlock     bool // Track if we're inside a finally block (for control flow validation)

	// Loop control context tracking (Task 8.235c)
	inLoop    bool // Track if we're inside a loop body (for break/continue validation)
	loopDepth int  // Track loop nesting level
}

// NewAnalyzer creates a new semantic analyzer
func NewAnalyzer() *Analyzer {
	a := &Analyzer{
		symbols:            NewSymbolTable(),
		errors:             make([]string, 0),
		classes:            make(map[string]*types.ClassType),
		interfaces:         make(map[string]*types.InterfaceType),
		enums:              make(map[string]*types.EnumType),
		records:            make(map[string]*types.RecordType),
		sets:               make(map[string]*types.SetType),
		arrays:             make(map[string]*types.ArrayType),
		globalOperators:    types.NewOperatorRegistry(),
		conversionRegistry: types.NewConversionRegistry(),
	}

	// Task 8.203: Register built-in Exception base class
	a.registerBuiltinExceptionTypes()

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

	a.classes["TObject"] = objectClass

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

	a.classes["Exception"] = exceptionClass

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

		a.classes[excName] = excClass
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

	// If we accumulated errors, return them
	if len(a.errors) > 0 {
		return &AnalysisError{Errors: a.errors}
	}

	return nil
}

// Errors returns all accumulated semantic errors
func (a *Analyzer) Errors() []string {
	return a.errors
}

// addError adds a semantic error to the error list
func (a *Analyzer) addError(format string, args ...interface{}) {
	a.errors = append(a.errors, fmt.Sprintf(format, args...))
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
	}
	if sig, ok := a.conversionRegistry.FindImplicit(from, to); ok && sig != nil {
		return true
	}
	return false
}
