package semantic

import (
	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/types"
)

// ============================================================================
// Exception Statement Analysis (Tasks 8.205-8.210)
// ============================================================================

// Task 8.208: Analyze raise statement
func (a *Analyzer) analyzeRaiseStatement(stmt *ast.RaiseStatement) {
	// Bare raise (re-raise current exception)
	if stmt.Exception == nil {
		// Bare raise is only valid inside an exception handler
		// For now, we allow it (runtime will check context)
		// TODO: Track if we're inside an except block during analysis
		return
	}

	// Analyze the exception expression
	excType := a.analyzeExpression(stmt.Exception)
	if excType == nil {
		// Error already reported
		return
	}

	// Validate that the expression evaluates to an Exception type
	if !a.isExceptionType(excType) {
		a.addError("raise statement requires Exception type, got %s", excType.String())
	}
}

// Task 8.205: Analyze try statement
func (a *Analyzer) analyzeTryStatement(stmt *ast.TryStatement) {
	// Analyze try block
	if stmt.TryBlock != nil {
		a.analyzeBlock(stmt.TryBlock)
	}

	// Analyze except clause if present
	if stmt.ExceptClause != nil {
		a.analyzeExceptClause(stmt.ExceptClause)
	}

	// Analyze finally clause if present
	if stmt.FinallyClause != nil {
		a.analyzeFinallyClause(stmt.FinallyClause)
	}

	// Validate that at least one of except or finally is present
	if stmt.ExceptClause == nil && stmt.FinallyClause == nil {
		a.addError("try statement must have either except or finally clause")
	}
}

// Task 8.206: Analyze except clause
func (a *Analyzer) analyzeExceptClause(clause *ast.ExceptClause) {
	// Analyze each exception handler
	for _, handler := range clause.Handlers {
		a.analyzeExceptionHandler(handler)
	}

	// Analyze else block if present
	if clause.ElseBlock != nil {
		a.analyzeBlock(clause.ElseBlock)
	}
}

// Task 8.207: Analyze exception handler
func (a *Analyzer) analyzeExceptionHandler(handler *ast.ExceptionHandler) {
	// Validate exception type
	if handler.ExceptionType == nil {
		a.addError("exception handler must specify exception type")
		return
	}

	excType, err := a.resolveType(handler.ExceptionType.Name)
	if err != nil {
		a.addError("unknown exception type '%s'", handler.ExceptionType.Name)
		return
	}

	// Validate that the type is Exception-compatible
	if !a.isExceptionType(excType) {
		a.addError("exception handler type must be Exception or derived class, got %s", excType.String())
		return
	}

	// Create new scope for exception variable
	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(a.symbols)

	// Add exception variable to scope
	if handler.Variable != nil {
		a.symbols.Define(handler.Variable.Value, excType)
	}

	// Analyze handler statement in exception variable scope
	if handler.Statement != nil {
		a.analyzeStatement(handler.Statement)
	}

	// Restore previous scope
	a.symbols = oldSymbols
}

// Task 8.200: Analyze finally clause
func (a *Analyzer) analyzeFinallyClause(clause *ast.FinallyClause) {
	if clause.Block != nil {
		a.analyzeBlock(clause.Block)
	}
}

// isExceptionType checks if a type is Exception or derived from Exception
func (a *Analyzer) isExceptionType(t types.Type) bool {
	classType, ok := t.(*types.ClassType)
	if !ok {
		return false
	}

	// Check if this is Exception or inherits from Exception
	for classType != nil {
		if classType.Name == "Exception" {
			return true
		}
		classType = classType.Parent
	}

	return false
}
