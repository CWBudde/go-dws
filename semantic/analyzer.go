package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/types"
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

	// Current class being analyzed (for field/method access)
	currentClass *types.ClassType

	// Operator registries (Stage 8)
	globalOperators    *types.OperatorRegistry
	conversionRegistry *types.ConversionRegistry
}

// NewAnalyzer creates a new semantic analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		symbols:            NewSymbolTable(),
		errors:             make([]string, 0),
		classes:            make(map[string]*types.ClassType),
		interfaces:         make(map[string]*types.InterfaceType),
		enums:              make(map[string]*types.EnumType),
		records:            make(map[string]*types.RecordType),
		globalOperators:    types.NewOperatorRegistry(),
		conversionRegistry: types.NewConversionRegistry(),
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
