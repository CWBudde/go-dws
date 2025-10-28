package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
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
	case *ast.ConstDecl:
		a.analyzeConstDecl(s)
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
	case *ast.ForInStatement:
		a.analyzeForIn(s)
	case *ast.CaseStatement:
		a.analyzeCase(s)
	case *ast.BreakStatement:
		a.analyzeBreakStatement(s)
	case *ast.ContinueStatement:
		a.analyzeContinueStatement(s)
	case *ast.ExitStatement:
		a.analyzeExitStatement(s)
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
	case *ast.SetDecl:
		a.analyzeSetDecl(s)
	case *ast.ArrayDecl:
		a.analyzeArrayDecl(s)
	case *ast.TypeDeclaration:
		a.analyzeTypeDeclaration(s)
	case *ast.RaiseStatement:
		a.analyzeRaiseStatement(s)
	case *ast.TryStatement:
		a.analyzeTryStatement(s)
	case *ast.UsesClause:
		// Uses clauses are handled at runtime by the interpreter
		// Semantic analyzer just ignores them
		return
	default:
		// Unknown statement type - this shouldn't happen
		a.addError("unknown statement type: %T", stmt)
	}
}

// analyzeVarDecl analyzes a variable declaration
func (a *Analyzer) analyzeVarDecl(stmt *ast.VarDeclStatement) {
	// Task 9.63: Check each name for duplicates in current scope
	for _, name := range stmt.Names {
		if a.symbols.IsDeclaredInCurrentScope(name.Value) {
			a.addError("variable '%s' already declared at %s", name.Value, stmt.Token.Pos.String())
			return
		}
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
	// Note: Parser already validates that multi-name declarations cannot have initializers
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
		} else if setLit, ok := stmt.Value.(*ast.SetLiteral); ok {
			// Special handling for set literals - they need the expected type
			initType = a.analyzeSetLiteralWithContext(setLit, varType)
			if initType == nil {
				// Error already reported by analyzeSetLiteralWithContext
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
		// Use first name for error message
		a.addError("variable '%s' must have either a type annotation or an initializer at %s",
			stmt.Names[0].Value, stmt.Token.Pos.String())
		return
	}

	// Task 9.63: Add all variables to symbol table with the same type
	for _, name := range stmt.Names {
		a.symbols.Define(name.Value, varType)
	}
}

// analyzeConstDecl analyzes a const declaration statement
func (a *Analyzer) analyzeConstDecl(stmt *ast.ConstDecl) {
	// Check if constant is already declared in current scope
	if a.symbols.IsDeclaredInCurrentScope(stmt.Name.Value) {
		a.addError("constant '%s' already declared at %s", stmt.Name.Value, stmt.Token.Pos.String())
		return
	}

	// Constants must have a value
	if stmt.Value == nil {
		a.addError("constant '%s' must have a value at %s", stmt.Name.Value, stmt.Token.Pos.String())
		return
	}

	// Determine the type of the constant
	var constType types.Type
	var err error

	if stmt.Type != nil {
		// Explicit type annotation
		constType, err = a.resolveType(stmt.Type.Name)
		if err != nil {
			a.addError("unknown type '%s' at %s", stmt.Type.Name, stmt.Token.Pos.String())
			return
		}
	}

	// Analyze the value expression
	var valueType types.Type

	// Special handling for record literals - they need the expected type
	if recordLit, ok := stmt.Value.(*ast.RecordLiteral); ok {
		if constType == nil {
			a.addError("record literal requires explicit type annotation at %s", stmt.Token.Pos.String())
			return
		}
		valueType = a.analyzeRecordLiteral(recordLit, constType)
		if valueType == nil {
			// Error already reported by analyzeRecordLiteral
			return
		}
	} else if setLit, ok := stmt.Value.(*ast.SetLiteral); ok {
		// Special handling for set literals - they need the expected type
		valueType = a.analyzeSetLiteralWithContext(setLit, constType)
		if valueType == nil {
			// Error already reported by analyzeSetLiteralWithContext
			return
		}
	} else {
		valueType = a.analyzeExpression(stmt.Value)
		if valueType == nil {
			// Error already reported by analyzeExpression
			return
		}
	}

	if constType == nil {
		// Type inference: use value's type
		constType = valueType
	} else {
		// Check that value type is compatible with declared type
		if !a.canAssign(valueType, constType) {
			a.addError("cannot assign %s to %s in constant declaration at %s",
				valueType.String(), constType.String(), stmt.Token.Pos.String())
			return
		}
	}

	// Add constant to symbol table as read-only const
	a.symbols.DefineConst(stmt.Name.Value, constType)
}

// analyzeAssignment analyzes an assignment statement
func (a *Analyzer) analyzeAssignment(stmt *ast.AssignmentStatement) {
	// Check the type of the value being assigned first
	valueType := a.analyzeExpression(stmt.Value)
	if valueType == nil {
		// Error already reported
		return
	}

	// Handle different target types
	switch target := stmt.Target.(type) {
	case *ast.Identifier:
		// Simple variable assignment: x := value
		sym, ok := a.symbols.Resolve(target.Value)
		if !ok {
			a.addError("undefined variable '%s' at %s", target.Value, stmt.Token.Pos.String())
			return
		}

		// Check if variable is read-only (Task 8.207, 8.255)
		if sym.ReadOnly {
			if sym.IsConst {
				a.addError("Cannot assign to constant '%s' at %s", target.Value, stmt.Token.Pos.String())
			} else {
				a.addError("cannot assign to read-only variable '%s' at %s", target.Value, stmt.Token.Pos.String())
			}
			return
		}

		// Check type compatibility
		if !a.canAssign(valueType, sym.Type) {
			a.addError("cannot assign %s to %s at %s",
				valueType.String(), sym.Type.String(), stmt.Token.Pos.String())
		}

	case *ast.MemberAccessExpression:
		// Member assignment: obj.field := value
		// Analyze the target to ensure it's valid
		targetType := a.analyzeExpression(target)
		if targetType == nil {
			return
		}

		// Check type compatibility
		if !a.canAssign(valueType, targetType) {
			a.addError("cannot assign %s to %s at %s",
				valueType.String(), targetType.String(), stmt.Token.Pos.String())
		}

	case *ast.IndexExpression:
		// Array index assignment: arr[i] := value
		// Analyze the target to ensure it's valid
		targetType := a.analyzeExpression(target)
		if targetType == nil {
			return
		}

		// Check type compatibility
		if !a.canAssign(valueType, targetType) {
			a.addError("cannot assign %s to %s at %s",
				valueType.String(), targetType.String(), stmt.Token.Pos.String())
		}

	default:
		a.addError("invalid assignment target at %s", stmt.Token.Pos.String())
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

	// Task 8.235g: Set loop context before analyzing body
	oldInLoop := a.inLoop
	a.inLoop = true
	a.loopDepth++
	defer func() {
		a.inLoop = oldInLoop
		a.loopDepth--
	}()

	// Analyze body
	a.analyzeStatement(stmt.Body)
}

// analyzeRepeat analyzes a repeat-until statement
func (a *Analyzer) analyzeRepeat(stmt *ast.RepeatStatement) {
	// Task 8.235g: Set loop context before analyzing body
	oldInLoop := a.inLoop
	a.inLoop = true
	a.loopDepth++
	defer func() {
		a.inLoop = oldInLoop
		a.loopDepth--
	}()

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

	// Task 8.235g: Set loop context before analyzing body
	oldInLoop := a.inLoop
	a.inLoop = true
	a.loopDepth++
	defer func() {
		a.inLoop = oldInLoop
		a.loopDepth--
	}()

	// Analyze body
	a.analyzeStatement(stmt.Body)
}

// analyzeForIn analyzes a for-in loop statement
func (a *Analyzer) analyzeForIn(stmt *ast.ForInStatement) {
	// Create a new scope for the loop variable
	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(oldSymbols)
	defer func() { a.symbols = oldSymbols }()

	// Analyze collection expression
	collectionType := a.analyzeExpression(stmt.Collection)

	// Determine the element type and validate the collection is enumerable
	var elementType types.Type

	if collectionType != nil {
		switch ct := collectionType.(type) {
		case *types.ArrayType:
			// Arrays are enumerable, element type is the array's element type
			elementType = ct.ElementType

		case *types.SetType:
			// Sets are enumerable, element type is the set's element type
			elementType = ct.ElementType

		case *types.StringType:
			// Strings are enumerable, iterates character by character
			// In DWScript, characters are represented as strings
			elementType = types.STRING

		case *types.TypeAlias:
			// Unwrap type alias and check the underlying type
			underlyingType := ct.AliasedType
			// Re-check with underlying type
			switch ut := underlyingType.(type) {
			case *types.ArrayType:
				elementType = ut.ElementType
			case *types.SetType:
				elementType = ut.ElementType
			case *types.StringType:
				elementType = types.STRING
			default:
				a.addError("for-in collection type %s (alias of %s) is not enumerable at %s",
					ct.String(), underlyingType.String(), stmt.Token.Pos.String())
				elementType = types.VOID
			}

		default:
			// Not an enumerable type
			a.addError("for-in collection type %s is not enumerable at %s",
				collectionType.String(), stmt.Token.Pos.String())
			elementType = types.VOID
		}
	} else {
		// Collection type could not be determined, use VOID
		elementType = types.VOID
	}

	// Define loop variable with the element type
	a.symbols.Define(stmt.Variable.Value, elementType)

	// Set loop context before analyzing body
	oldInLoop := a.inLoop
	a.inLoop = true
	a.loopDepth++
	defer func() {
		a.inLoop = oldInLoop
		a.loopDepth--
	}()

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

// analyzeBreakStatement analyzes a break statement (Task 8.235d)
func (a *Analyzer) analyzeBreakStatement(stmt *ast.BreakStatement) {
	// Task 8.235h: Check if we're inside a finally block
	if a.inFinallyBlock {
		a.addError("break statement not allowed in finally block at %s", stmt.Token.Pos.String())
		return
	}

	// Check if we're inside a loop
	if !a.inLoop {
		a.addError("break statement not allowed outside loop at %s", stmt.Token.Pos.String())
		return
	}
	// Valid break statement - no further analysis needed
}

// analyzeContinueStatement analyzes a continue statement (Task 8.235e)
func (a *Analyzer) analyzeContinueStatement(stmt *ast.ContinueStatement) {
	// Task 8.235h: Check if we're inside a finally block
	if a.inFinallyBlock {
		a.addError("continue statement not allowed in finally block at %s", stmt.Token.Pos.String())
		return
	}

	// Check if we're inside a loop
	if !a.inLoop {
		a.addError("continue statement not allowed outside loop at %s", stmt.Token.Pos.String())
		return
	}
	// Valid continue statement - no further analysis needed
}

// analyzeExitStatement analyzes an exit statement (Task 8.235f)
func (a *Analyzer) analyzeExitStatement(stmt *ast.ExitStatement) {
	// Task 8.235h: Check if we're inside a finally block
	if a.inFinallyBlock {
		a.addError("exit statement not allowed in finally block at %s", stmt.Token.Pos.String())
		return
	}

	// If we're at the top level (not in a function), only allow exit without a value
	if a.currentFunction == nil {
		if stmt.Value != nil {
			a.addError("exit with value not allowed at program level at %s", stmt.Token.Pos.String())
		}
		// exit without value is allowed at program level (exits the program)
		return
	}

	// If exit has a return value, analyze it
	if stmt.Value != nil {
		valueType := a.analyzeExpression(stmt.Value)

		// Check that the return value type matches the function's return type
		if a.currentFunction.ReturnType != nil {
			expectedType, err := a.resolveType(a.currentFunction.ReturnType.Name)
			if err != nil {
				a.addError("cannot resolve function return type: %v at %s", err, stmt.Token.Pos.String())
				return
			}
			if expectedType != nil && valueType != nil && !a.canAssign(valueType, expectedType) {
				a.addError("exit value type %s incompatible with function return type %s at %s",
					valueType.String(), expectedType.String(), stmt.Token.Pos.String())
			}
		} else {
			// Procedure (no return type) - exit should not have a value
			a.addError("exit with value not allowed in procedure at %s", stmt.Token.Pos.String())
		}
	} else {
		// Exit without value - check if function expects a return value
		if a.currentFunction.ReturnType != nil {
			// Function expects a return value but exit doesn't provide one
			// This is actually allowed in DWScript - it will return a default value
			// So we don't emit an error here
		}
	}
}
