package interp

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

// This file contains built-in functions that require AST-level access to modify
// variables in-place (var parameters). These functions cannot be migrated to the
// builtins package because they need access to AST nodes to identify and modify
// the target variables directly in the environment.
//
// All other built-in functions have been migrated to internal/interp/builtins/
// and are registered in builtins.DefaultRegistry.
//
// Functions in this file are routed through callBuiltinWithVarParam() in
// functions_builtins.go.

// builtinInsert implements the Insert() built-in function.
// It inserts a source string into a target string at the specified position.
// Insert(source, target, pos) - modifies target in-place (1-based position)
func (i *Interpreter) builtinInsert(args []ast.Expression) Value {
	// Validate argument count (3 arguments)
	if len(args) != 3 {
		return i.newErrorWithLocation(i.currentNode, "Insert() expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: source string to insert (evaluate it)
	sourceVal := i.Eval(args[0])
	if isError(sourceVal) {
		return sourceVal
	}
	sourceStr, ok := sourceVal.(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Insert() expects String as first argument (source), got %s", sourceVal.Type())
	}

	// Second argument: target string variable (must be an identifier)
	targetIdent, ok := args[1].(*ast.Identifier)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Insert() second argument (target) must be a variable, got %T", args[1])
	}

	targetName := targetIdent.Value

	// Get current target value from environment
	currentVal, exists := i.env.Get(targetName)
	if !exists {
		return i.newErrorWithLocation(i.currentNode, "undefined variable: %s", targetName)
	}

	targetStr, ok := currentVal.(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Insert() expects target to be String, got %s", currentVal.Type())
	}

	// Third argument: position (1-based index)
	posVal := i.Eval(args[2])
	if isError(posVal) {
		return posVal
	}
	posInt, ok := posVal.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Insert() expects Integer as third argument (position), got %s", posVal.Type())
	}

	pos := int(posInt.Value)
	target := targetStr.Value
	source := sourceStr.Value

	// Use rune-based insertion to handle UTF-8 correctly
	newStr := runeInsert(source, target, pos)

	// Update the target variable with the new string
	newValue := &StringValue{Value: newStr}
	if err := i.env.Set(targetName, newValue); err != nil {
		return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", targetName, err)
	}

	return &NilValue{}
}

// builtinDeleteString implements the Delete() built-in function for strings.
// It deletes count characters from a string starting at the specified position.
// Delete(s, pos, count) - modifies s in-place (1-based position)
func (i *Interpreter) builtinDeleteString(args []ast.Expression) Value {
	// Validate argument count (3 arguments)
	if len(args) != 3 {
		return i.newErrorWithLocation(i.currentNode, "Delete() for strings expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: string variable (must be an identifier)
	strIdent, ok := args[0].(*ast.Identifier)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Delete() first argument must be a variable, got %T", args[0])
	}

	strName := strIdent.Value

	// Get current string value from environment
	currentVal, exists := i.env.Get(strName)
	if !exists {
		return i.newErrorWithLocation(i.currentNode, "undefined variable: %s", strName)
	}

	strVal, ok := currentVal.(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Delete() expects first argument to be String, got %s", currentVal.Type())
	}

	// Second argument: position (1-based index)
	posVal := i.Eval(args[1])
	if isError(posVal) {
		return posVal
	}
	posInt, ok := posVal.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Delete() expects Integer as second argument (position), got %s", posVal.Type())
	}

	// Third argument: count (number of characters to delete)
	countVal := i.Eval(args[2])
	if isError(countVal) {
		return countVal
	}
	countInt, ok := countVal.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Delete() expects Integer as third argument (count), got %s", countVal.Type())
	}

	pos := int(posInt.Value)
	count := int(countInt.Value)
	str := strVal.Value

	// Use rune-based deletion to handle UTF-8 correctly
	newStr := runeDelete(str, pos, count)

	// Update the string variable with the new value
	newValue := &StringValue{Value: newStr}
	if err := i.env.Set(strName, newValue); err != nil {
		return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", strName, err)
	}

	return &NilValue{}
}
