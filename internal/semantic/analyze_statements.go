package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
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
	case *ast.HelperDecl:
		a.analyzeHelperDecl(s)
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

	firstName := ""
	if len(stmt.Names) > 0 {
		firstName = stmt.Names[0].Value
	}

	// Determine the type of the variable
	var varType types.Type
	var err error

	if stmt.Type != nil {
		// Explicit type annotation - check for inline type first (arrays, function pointers)
		// Task: Fix negative array bounds like array[-5..5]
		if stmt.Type.InlineType != nil {
			// Use the stored AST node for complex types to avoid string conversion issues
			varType, err = a.resolveTypeExpression(stmt.Type.InlineType)
			if err != nil {
				// Get type name for error message
				typeName := getTypeExpressionName(stmt.Type.InlineType)
				a.addError("unknown type '%s' at %s", typeName, stmt.Token.Pos.String())
				return
			}
		} else {
			// Simple type name (Integer, String, etc.)
			varType, err = a.resolveType(stmt.Type.Name)
			if err != nil {
				a.addError("unknown type '%s' at %s", stmt.Type.Name, stmt.Token.Pos.String())
				return
			}
		}
	}

	// If there's an initializer, check its type
	// Note: Parser already validates that multi-name declarations cannot have initializers
	if stmt.Value != nil {
		initType := a.analyzeExpressionWithExpectedType(stmt.Value, varType)
		if initType == nil {
			if stmt.Type == nil {
				a.addError("cannot infer type for variable '%s' from initializer at %s",
					firstName, stmt.Token.Pos.String())
			}
			// Error already reported
			return
		}

		if varType == nil {
			underlying := types.GetUnderlyingType(initType)
			if _, isNil := underlying.(*types.NilType); isNil {
				a.addError("cannot infer type for variable '%s' from nil initializer at %s",
					firstName, stmt.Token.Pos.String())
				return
			}
			// Type inference: use initializer's type
			varType = initType
		} else {
			// Check that initializer type is compatible with declared type
			if !a.canAssign(initType, varType) {
				// Task 9.110: Use structured error for type mismatch
				a.addStructuredError(NewTypeMismatch(
					stmt.Token.Pos,
					firstName, // Variable name
					varType,   // Expected type
					initType,  // Got type
				))
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

	valueType = a.analyzeExpressionWithExpectedType(stmt.Value, constType)
	if valueType == nil {
		// Error already reported
		return
	}

	if constType == nil {
		// Type inference: use value's type
		constType = valueType
	} else {
		// Check that value type is compatible with declared type
		if !a.canAssign(valueType, constType) {
			// Task 9.110: Use structured error for type mismatch
			a.addStructuredError(NewTypeMismatch(
				stmt.Token.Pos,
				stmt.Name.Value, // Constant name
				constType,       // Expected type
				valueType,       // Got type
			))
			return
		}
	}

	// Task 9.205: Evaluate the constant value at compile time
	constValue, err := a.evaluateConstant(stmt.Value)
	if err != nil {
		a.addError("constant '%s' value must be a compile-time constant at %s: %v",
			stmt.Name.Value, stmt.Token.Pos.String(), err)
		return
	}

	// Add constant to symbol table with its compile-time value
	a.symbols.DefineConst(stmt.Name.Value, constType, constValue)
}

// analyzeAssignment analyzes an assignment statement
func (a *Analyzer) analyzeAssignment(stmt *ast.AssignmentStatement) {
	// Determine if this is a compound assignment
	isCompound := stmt.Operator != lexer.ASSIGN && stmt.Operator != lexer.TokenType(0)

	// Handle different target types
	switch target := stmt.Target.(type) {
	case *ast.Identifier:
		// Simple variable assignment: x := value or x += value

		// Special case: In DWScript, you can assign to the function name to set the return value
		// Check if we're inside a function and the target matches the function name
		if a.currentFunction != nil && strings.EqualFold(target.Value, a.currentFunction.Name.Value) {
			// Assigning to function name - treat it as assigning to Result
			if a.currentFunction.ReturnType == nil {
				a.addError("cannot assign to procedure name '%s' (procedures have no return value) at %s",
					target.Value, stmt.Token.Pos.String())
				return
			}

			// Get the return type
			returnType, err := a.resolveType(a.currentFunction.ReturnType.Name)
			if err != nil || returnType == nil {
				a.addError("cannot resolve return type '%s' at %s",
					a.currentFunction.ReturnType.Name, stmt.Token.Pos.String())
				return
			}

			// Analyze the value being assigned
			valueType := a.analyzeExpressionWithExpectedType(stmt.Value, returnType)
			if valueType == nil {
				return
			}

			// Check type compatibility
			if !a.canAssign(valueType, returnType) {
				a.addError("cannot assign %s to %s at %s",
					valueType.String(), returnType.String(), stmt.Token.Pos.String())
			}
			return
		}

		sym, ok := a.symbols.Resolve(target.Value)

		// Task 9.32b/9.32c: If variable not found, check for implicit Self field/property
		if !ok && a.currentClass != nil {
			// Check if it's a field of the current class
			if fieldType, exists := a.currentClass.Fields[target.Value]; exists {
				valueType := a.analyzeExpressionWithExpectedType(stmt.Value, fieldType)
				if valueType == nil {
					return
				}
				if isCompound {
					if !a.isCompoundOperatorValid(stmt.Operator, fieldType, valueType, stmt.Token.Pos) {
						return
					}
				}
				if !a.canAssign(valueType, fieldType) {
					a.addError("cannot assign %s to %s at %s",
						valueType.String(), fieldType.String(), stmt.Token.Pos.String())
				}
				return
			}

			// Check if it's a property of the current class
			// DWScript is case-insensitive, so we need to search all properties
			// Also search parent class hierarchy
			for class := a.currentClass; class != nil; class = class.Parent {
				for propName, propInfo := range class.Properties {
					if strings.EqualFold(propName, target.Value) {
						// Check if property is writable
						if propInfo.WriteKind == types.PropAccessNone {
							a.addError("property '%s' is read-only at %s", target.Value, stmt.Token.Pos.String())
							return
						}
						valueType := a.analyzeExpressionWithExpectedType(stmt.Value, propInfo.Type)
						if valueType == nil {
							return
						}
						if isCompound {
							if !a.isCompoundOperatorValid(stmt.Operator, propInfo.Type, valueType, stmt.Token.Pos) {
								return
							}
						}
						if !a.canAssign(valueType, propInfo.Type) {
							a.addError("cannot assign %s to %s at %s",
								valueType.String(), propInfo.Type.String(), stmt.Token.Pos.String())
						}
						return
					}
				}
			}
		}

		if !ok {
			// Task 9.110: Use structured error for undefined variable
			a.addStructuredError(NewUndefinedVariable(stmt.Token.Pos, target.Value))
			return
		}

		// Check if variable is read-only
		if sym.ReadOnly {
			if sym.IsConst {
				a.addError("Cannot assign to constant '%s' at %s", target.Value, stmt.Token.Pos.String())
			} else {
				a.addError("cannot assign to read-only variable '%s' at %s", target.Value, stmt.Token.Pos.String())
			}
			return
		}

		// For compound assignments, validate operator compatibility
		valueType := a.analyzeExpressionWithExpectedType(stmt.Value, sym.Type)
		if valueType == nil {
			return
		}

		if isCompound {
			if !a.isCompoundOperatorValid(stmt.Operator, sym.Type, valueType, stmt.Token.Pos) {
				return
			}
		}

		// Check type compatibility
		if !a.canAssign(valueType, sym.Type) {
			a.addError("cannot assign %s to %s at %s",
				valueType.String(), sym.Type.String(), stmt.Token.Pos.String())
		}

	case *ast.MemberAccessExpression:
		// Member assignment: obj.field := value or obj.field += value
		// Analyze the target to ensure it's valid
		targetType := a.analyzeExpression(target)
		if targetType == nil {
			return
		}

		valueType := a.analyzeExpressionWithExpectedType(stmt.Value, targetType)
		if valueType == nil {
			return
		}

		// For compound assignments, validate operator compatibility
		if isCompound {
			if !a.isCompoundOperatorValid(stmt.Operator, targetType, valueType, stmt.Token.Pos) {
				return
			}
		}

		// Check type compatibility
		if !a.canAssign(valueType, targetType) {
			a.addError("cannot assign %s to %s at %s",
				valueType.String(), targetType.String(), stmt.Token.Pos.String())
		}

	case *ast.IndexExpression:
		// Array index assignment: arr[i] := value or arr[i] += value
		// Analyze the target to ensure it's valid
		targetType := a.analyzeExpression(target)
		if targetType == nil {
			return
		}

		valueType := a.analyzeExpressionWithExpectedType(stmt.Value, targetType)
		if valueType == nil {
			return
		}

		// For compound assignments, validate operator compatibility
		if isCompound {
			if !a.isCompoundOperatorValid(stmt.Operator, targetType, valueType, stmt.Token.Pos) {
				return
			}
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

// isCompoundOperatorValid checks if a compound operator is valid for the given types.
func (a *Analyzer) isCompoundOperatorValid(op lexer.TokenType, targetType, valueType types.Type, pos lexer.Position) bool {
	switch op {
	case lexer.PLUS_ASSIGN:
		// += works with Integer, Float, String (concatenation)
		if targetType.Equals(types.INTEGER) || targetType.Equals(types.FLOAT) || targetType.Equals(types.STRING) {
			return true
		}
		a.addError("operator += not supported for type %s at %s", targetType.String(), pos.String())
		return false

	case lexer.MINUS_ASSIGN, lexer.TIMES_ASSIGN, lexer.DIVIDE_ASSIGN:
		// -=, *=, /= work with Integer, Float only
		if targetType.Equals(types.INTEGER) || targetType.Equals(types.FLOAT) {
			return true
		}
		opStr := "operator"
		switch op {
		case lexer.MINUS_ASSIGN:
			opStr = "operator -="
		case lexer.TIMES_ASSIGN:
			opStr = "operator *="
		case lexer.DIVIDE_ASSIGN:
			opStr = "operator /="
		}
		a.addError("%s not supported for type %s at %s", opStr, targetType.String(), pos.String())
		return false

	default:
		return true
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

// analyzeRepeat analyzes a repeat-until statement
func (a *Analyzer) analyzeRepeat(stmt *ast.RepeatStatement) {
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

	// Task 9.152: Analyze step expression if present
	if stmt.Step != nil {
		stepType := a.analyzeExpression(stmt.Step)

		// Validate step expression type is Integer
		if stepType != nil && stepType != types.INTEGER {
			a.addError("for loop step must be Integer, got %s at %s",
				stepType.String(), stmt.Token.Pos.String())
		}

		// Optional optimization: check constant step values at compile time
		if stepLiteral, ok := stmt.Step.(*ast.IntegerLiteral); ok {
			if stepLiteral.Value <= 0 {
				a.addError("for loop step must be strictly positive, got %d at %s",
					stepLiteral.Value, stmt.Token.Pos.String())
			}
		} else if unaryExpr, ok := stmt.Step.(*ast.UnaryExpression); ok {
			// Check for negative integer literals: -1, -5, etc.
			if unaryExpr.Operator == "-" {
				if innerLiteral, ok := unaryExpr.Right.(*ast.IntegerLiteral); ok {
					a.addError("for loop step must be strictly positive, got %d at %s",
						-innerLiteral.Value, stmt.Token.Pos.String())
				}
			}
		}
	}

	// Define loop variable (typically Integer)
	var loopVarType types.Type = types.INTEGER
	if startType != nil && types.IsOrdinalType(startType) {
		loopVarType = startType
	}
	a.symbols.Define(stmt.Variable.Value, loopVarType)

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

		case *types.EnumType:
			// Task 9.213: Enum types are enumerable
			// When iterating over an enum type directly (e.g., for var e in TColor do),
			// we iterate over all values of the enum type
			// The element type is the enum type itself
			elementType = ct

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
			case *types.EnumType:
				// Task 9.213: Aliased enum types are also enumerable
				elementType = ut
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
			// Check if this is a range expression
			if rangeExpr, isRange := value.(*ast.RangeExpression); isRange {
				// Analyze both start and end of range
				startType := a.analyzeExpression(rangeExpr.Start)
				endType := a.analyzeExpression(rangeExpr.End)

				// Check start is compatible with case expression
				if caseType != nil && startType != nil {
					if !a.canAssign(startType, caseType) {
						a.addError("case range start type %s incompatible with case expression type %s at %s",
							startType.String(), caseType.String(), rangeExpr.Start.Pos().String())
					}
				}

				// Check end is compatible with case expression
				if caseType != nil && endType != nil {
					if !a.canAssign(endType, caseType) {
						a.addError("case range end type %s incompatible with case expression type %s at %s",
							endType.String(), caseType.String(), rangeExpr.End.Pos().String())
					}
				}

				// Check start and end are compatible with each other
				if startType != nil && endType != nil {
					if !a.canAssign(startType, endType) && !a.canAssign(endType, startType) {
						a.addError("case range start type %s and end type %s are incompatible at %s",
							startType.String(), endType.String(), rangeExpr.Pos().String())
					}
				}
			} else {
				// Regular value (not a range)
				valueType := a.analyzeExpression(value)
				if caseType != nil && valueType != nil {
					if !a.canAssign(valueType, caseType) {
						a.addError("case value type %s incompatible with case expression type %s",
							valueType.String(), caseType.String())
					}
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

// analyzeBreakStatement analyzes a break statement
func (a *Analyzer) analyzeBreakStatement(stmt *ast.BreakStatement) {
	// Check if we're inside a finally block
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

// analyzeContinueStatement analyzes a continue statement
func (a *Analyzer) analyzeContinueStatement(stmt *ast.ContinueStatement) {
	// Check if we're inside a finally block
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

// analyzeExitStatement analyzes an exit statement
func (a *Analyzer) analyzeExitStatement(stmt *ast.ExitStatement) {
	// Check if we're inside a finally block
	if a.inFinallyBlock {
		a.addError("exit statement not allowed in finally block at %s", stmt.Token.Pos.String())
		return
	}

	// If we're at the top level (not in a function), only allow exit without a value
	if a.currentFunction == nil {
		if stmt.ReturnValue != nil {
			a.addError("exit with value not allowed at program level at %s", stmt.Token.Pos.String())
		}
		// exit without value is allowed at program level (exits the program)
		return
	}

	// Determine expected return type for the current function/procedure
	var expectedType types.Type = types.VOID
	if a.currentFunction.ReturnType != nil {
		var err error
		expectedType, err = a.resolveType(a.currentFunction.ReturnType.Name)
		if err != nil {
			a.addError("cannot resolve function return type: %v at %s", err, stmt.Token.Pos.String())
			return
		}
		if expectedType == nil {
			a.addError("function has unknown return type at %s", stmt.Token.Pos.String())
			return
		}
	}

	// If exit has a return value, analyze it
	if stmt.ReturnValue != nil {
		if expectedType == types.VOID {
			// Procedure (no return type) - exit should not have a value
			a.addError("exit with value not allowed in procedure at %s", stmt.Token.Pos.String())
			return
		}

		valueType := a.analyzeExpression(stmt.ReturnValue)
		if valueType != nil && !a.canAssign(valueType, expectedType) {
			a.addError("exit value type %s incompatible with function return type %s at %s",
				valueType.String(), expectedType.String(), stmt.Token.Pos.String())
		}
	}
	// Exit without an explicit return value is allowed. Functions rely on the current
	// Result variable (or their default) in that case, matching DWScript semantics.
}
