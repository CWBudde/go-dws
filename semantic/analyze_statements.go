package semantic

import (
	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/types"
)

// ============================================================================
// Statement Analysis
// ============================================================================

// analyzeStatement analyzes a statement node
func (a *Analyzer) analyzeStatement(stmt ast.Statement) {
	if stmt == nil {
		return
	}

	switch s := stmt.(type) {
	case *ast.VarDeclStatement:
		a.analyzeVarDecl(s)
	case *ast.AssignmentStatement:
		a.analyzeAssignment(s)
	case *ast.ExpressionStatement:
		a.analyzeExpression(s.Expression)
	case *ast.BlockStatement:
		a.analyzeBlock(s)
	case *ast.IfStatement:
		a.analyzeIf(s)
	case *ast.WhileStatement:
		a.analyzeWhile(s)
	case *ast.RepeatStatement:
		a.analyzeRepeat(s)
	case *ast.ForStatement:
		a.analyzeFor(s)
	case *ast.CaseStatement:
		a.analyzeCase(s)
	case *ast.FunctionDecl:
		a.analyzeFunctionDecl(s)
	case *ast.ReturnStatement:
		a.analyzeReturn(s)
	case *ast.ClassDecl:
		a.analyzeClassDecl(s)
	case *ast.InterfaceDecl:
		a.analyzeInterfaceDecl(s)
	case *ast.OperatorDecl:
		a.analyzeOperatorDecl(s)
	case *ast.EnumDecl:
		a.analyzeEnumDecl(s)
	case *ast.RecordDecl:
		a.analyzeRecordDecl(s)
	default:
		// Unknown statement type - this shouldn't happen
		a.addError("unknown statement type: %T", stmt)
	}
}

// analyzeVarDecl analyzes a variable declaration
func (a *Analyzer) analyzeVarDecl(stmt *ast.VarDeclStatement) {
	// Check if variable is already declared in current scope
	if a.symbols.IsDeclaredInCurrentScope(stmt.Name.Value) {
		a.addError("variable '%s' already declared at %s", stmt.Name.Value, stmt.Token.Pos.String())
		return
	}

	// Determine the type of the variable
	var varType types.Type
	var err error

	if stmt.Type != nil {
		// Explicit type annotation - use resolveType to handle both basic and class types
		varType, err = a.resolveType(stmt.Type.Name)
		if err != nil {
			a.addError("unknown type '%s' at %s", stmt.Type.Name, stmt.Token.Pos.String())
			return
		}
	}

	// If there's an initializer, check its type
	if stmt.Value != nil {
		var initType types.Type

		// Special handling for record literals - they need the expected type
		if recordLit, ok := stmt.Value.(*ast.RecordLiteral); ok {
			if varType == nil {
				a.addError("record literal requires explicit type annotation at %s", stmt.Token.Pos.String())
				return
			}
			initType = a.analyzeRecordLiteral(recordLit, varType)
			if initType == nil {
				// Error already reported by analyzeRecordLiteral
				return
			}
		} else {
			initType = a.analyzeExpression(stmt.Value)
			if initType == nil {
				// Error already reported by analyzeExpression
				return
			}
		}

		if varType == nil {
			// Type inference: use initializer's type
			varType = initType
		} else {
			// Check that initializer type is compatible with declared type
			if !a.canAssign(initType, varType) {
				a.addError("cannot assign %s to %s in variable declaration at %s",
					initType.String(), varType.String(), stmt.Token.Pos.String())
				return
			}
		}
	}

	// If we still don't have a type, that's an error
	if varType == nil {
		a.addError("variable '%s' must have either a type annotation or an initializer at %s",
			stmt.Name.Value, stmt.Token.Pos.String())
		return
	}

	// Add variable to symbol table
	a.symbols.Define(stmt.Name.Value, varType)
}

// analyzeAssignment analyzes an assignment statement
func (a *Analyzer) analyzeAssignment(stmt *ast.AssignmentStatement) {
	// Look up the variable
	sym, ok := a.symbols.Resolve(stmt.Name.Value)
	if !ok {
		a.addError("undefined variable '%s' at %s", stmt.Name.Value, stmt.Token.Pos.String())
		return
	}

	// Check the type of the value being assigned
	valueType := a.analyzeExpression(stmt.Value)
	if valueType == nil {
		// Error already reported
		return
	}

	// Check type compatibility
	if !a.canAssign(valueType, sym.Type) {
		a.addError("cannot assign %s to %s at %s",
			valueType.String(), sym.Type.String(), stmt.Token.Pos.String())
	}
}

// analyzeBlock analyzes a block statement
func (a *Analyzer) analyzeBlock(stmt *ast.BlockStatement) {
	// Create a new scope for the block
	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(oldSymbols)
	defer func() { a.symbols = oldSymbols }()

	// Analyze each statement in the block
	for _, s := range stmt.Statements {
		a.analyzeStatement(s)
	}
}

// analyzeIf analyzes an if statement
func (a *Analyzer) analyzeIf(stmt *ast.IfStatement) {
	// Check condition type
	condType := a.analyzeExpression(stmt.Condition)
	if condType != nil && !condType.Equals(types.BOOLEAN) {
		a.addError("if condition must be boolean, got %s at %s",
			condType.String(), stmt.Token.Pos.String())
	}

	// Analyze consequence
	a.analyzeStatement(stmt.Consequence)

	// Analyze alternative if present
	if stmt.Alternative != nil {
		a.analyzeStatement(stmt.Alternative)
	}
}

// analyzeWhile analyzes a while statement
func (a *Analyzer) analyzeWhile(stmt *ast.WhileStatement) {
	// Check condition type
	condType := a.analyzeExpression(stmt.Condition)
	if condType != nil && !condType.Equals(types.BOOLEAN) {
		a.addError("while condition must be boolean, got %s at %s",
			condType.String(), stmt.Token.Pos.String())
	}

	// Analyze body
	a.analyzeStatement(stmt.Body)
}

// analyzeRepeat analyzes a repeat-until statement
func (a *Analyzer) analyzeRepeat(stmt *ast.RepeatStatement) {
	// Analyze body
	a.analyzeStatement(stmt.Body)

	// Check condition type
	condType := a.analyzeExpression(stmt.Condition)
	if condType != nil && !condType.Equals(types.BOOLEAN) {
		a.addError("repeat-until condition must be boolean, got %s at %s",
			condType.String(), stmt.Token.Pos.String())
	}
}

// analyzeFor analyzes a for statement
func (a *Analyzer) analyzeFor(stmt *ast.ForStatement) {
	// Create a new scope for the loop variable
	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(oldSymbols)
	defer func() { a.symbols = oldSymbols }()

	// Analyze start and end expressions
	startType := a.analyzeExpression(stmt.Start)
	endType := a.analyzeExpression(stmt.End)

	// Check that both are ordinal types (Integer or Boolean)
	if startType != nil && !types.IsOrdinalType(startType) {
		a.addError("for loop start must be ordinal type, got %s at %s",
			startType.String(), stmt.Token.Pos.String())
	}
	if endType != nil && !types.IsOrdinalType(endType) {
		a.addError("for loop end must be ordinal type, got %s at %s",
			endType.String(), stmt.Token.Pos.String())
	}

	// Define loop variable (typically Integer)
	var loopVarType types.Type = types.INTEGER
	if startType != nil && types.IsOrdinalType(startType) {
		loopVarType = startType
	}
	a.symbols.Define(stmt.Variable.Value, loopVarType)

	// Analyze body
	a.analyzeStatement(stmt.Body)
}

// analyzeCase analyzes a case statement
func (a *Analyzer) analyzeCase(stmt *ast.CaseStatement) {
	// Analyze the case expression
	caseType := a.analyzeExpression(stmt.Expression)

	// Analyze each case branch
	for _, branch := range stmt.Cases {
		// Check that case values are compatible with case expression
		for _, value := range branch.Values {
			valueType := a.analyzeExpression(value)
			if caseType != nil && valueType != nil {
				if !a.canAssign(valueType, caseType) {
					a.addError("case value type %s incompatible with case expression type %s",
						valueType.String(), caseType.String())
				}
			}
		}
		// Analyze the branch statement
		a.analyzeStatement(branch.Statement)
	}

	// Analyze else branch if present
	if stmt.Else != nil {
		a.analyzeStatement(stmt.Else)
	}
}
