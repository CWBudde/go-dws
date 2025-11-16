package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Exception Statement Analysis
// ============================================================================

// Analyze raise statement
func (a *Analyzer) analyzeRaiseStatement(stmt *ast.RaiseStatement) {
	// Bare raise (re-raise current exception)
	if stmt.Exception == nil {
		// Bare raise is only valid inside an exception handler
		if !a.inExceptionHandler {
			a.addError("bare raise statement is only valid inside an exception handler")
		}
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

// Analyze try statement
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

// Analyze except clause
func (a *Analyzer) analyzeExceptClause(clause *ast.ExceptClause) {
	// Track exception types to detect duplicates
	seenTypes := make(map[string]bool)

	// Analyze each exception handler
	for _, handler := range clause.Handlers {
		// Check for duplicate exception types before analyzing
		if handler.ExceptionType != nil {
			typeName := getTypeExpressionName(handler.ExceptionType)
			if seenTypes[typeName] {
				a.addError("duplicate exception handler for type '%s'", typeName)
			}
			seenTypes[typeName] = true
		}

		a.analyzeExceptionHandler(handler)
	}

	// Analyze else block if present
	if clause.ElseBlock != nil {
		a.analyzeBlock(clause.ElseBlock)
	}
}

// Analyze exception handler
func (a *Analyzer) analyzeExceptionHandler(handler *ast.ExceptionHandler) {
	// For bare except handlers (handler.ExceptionType == nil), we don't need to validate the type
	// Bare except catches all exceptions
	var excType types.Type
	if handler.ExceptionType != nil {
		var err error
		excType, err = a.resolveType(getTypeExpressionName(handler.ExceptionType))
		if err != nil {
			a.addError("unknown exception type '%s'", getTypeExpressionName(handler.ExceptionType))
			return
		}

		// Validate that the type is Exception-compatible
		if !a.isExceptionType(excType) {
			a.addError("exception handler type must be Exception or derived class, got %s", excType.String())
			return
		}
	} else {
		// Bare except handler - catches all exceptions
		// Use Exception as the type for the scope
		if exceptionClass, exists := a.classes["exception"]; exists {
			excType = exceptionClass
		}
	}

	// Create new scope for exception variable
	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(a.symbols)

	// Add exception variable to scope as read-only
	if handler.Variable != nil {
		a.symbols.DefineReadOnly(handler.Variable.Value, excType)
	}

	// Set exception handler context for bare raise validation
	oldInExceptionHandler := a.inExceptionHandler
	a.inExceptionHandler = true

	// Analyze handler statement in exception variable scope
	if handler.Statement != nil {
		a.analyzeStatement(handler.Statement)
	}

	// Restore previous context and scope
	a.inExceptionHandler = oldInExceptionHandler
	a.symbols = oldSymbols
}

// Analyze finally clause
func (a *Analyzer) analyzeFinallyClause(clause *ast.FinallyClause) {
	if clause.Block != nil {
		// Set finally block context for control flow validation
		oldInFinallyBlock := a.inFinallyBlock
		a.inFinallyBlock = true

		a.analyzeBlock(clause.Block)

		// Restore previous context
		a.inFinallyBlock = oldInFinallyBlock
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
