package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
)

// builtinConcat implements the Concat() built-in function.
// It concatenates multiple strings together.
// Concat(str1, str2, ...) - variable number of string arguments
func (i *Interpreter) builtinConcat(args []Value) Value {
	if len(args) == 0 {
		return i.newErrorWithLocation(i.currentNode, "Concat() expects at least 1 argument, got 0")
	}

	// Build the concatenated string
	var result strings.Builder

	for idx, arg := range args {
		strVal, ok := arg.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "Concat() expects string as argument %d, got %s", idx+1, arg.Type())
		}
		result.WriteString(strVal.Value)
	}

	return &StringValue{Value: result.String()}
}

// builtinPos implements the Pos() built-in function.
// It finds the position of a substring within a string.
// Pos(substr, str) - returns 1-based position (0 if not found)
func (i *Interpreter) builtinPos(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "Pos() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: substring to find
	substrVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Pos() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string to search in
	strVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Pos() expects string as second argument, got %s", args[1].Type())
	}

	substr := substrVal.Value
	str := strVal.Value

	// Handle empty substring - returns 1 (found at start)
	if len(substr) == 0 {
		return &IntegerValue{Value: 1}
	}

	// Find the substring
	index := strings.Index(str, substr)

	// Convert to 1-based index (or 0 if not found)
	if index == -1 {
		return &IntegerValue{Value: 0}
	}

	return &IntegerValue{Value: int64(index + 1)}
}

// builtinUpperCase implements the UpperCase() built-in function.
// It converts a string to uppercase.
// UpperCase(str) - returns uppercase version of the string
func (i *Interpreter) builtinUpperCase(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "UpperCase() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "UpperCase() expects string as argument, got %s", args[0].Type())
	}

	return &StringValue{Value: strings.ToUpper(strVal.Value)}
}

// builtinLowerCase implements the LowerCase() built-in function.
// It converts a string to lowercase.
// LowerCase(str) - returns lowercase version of the string
func (i *Interpreter) builtinLowerCase(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "LowerCase() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "LowerCase() expects string as argument, got %s", args[0].Type())
	}

	return &StringValue{Value: strings.ToLower(strVal.Value)}
}

// builtinTrim implements the Trim() built-in function.
// It removes leading and trailing whitespace from a string.
// Trim(str) - returns string with whitespace removed from both ends
func (i *Interpreter) builtinTrim(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Trim() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Trim() expects string as argument, got %s", args[0].Type())
	}

	return &StringValue{Value: strings.TrimSpace(strVal.Value)}
}

// builtinTrimLeft implements the TrimLeft() built-in function.
// It removes leading whitespace from a string.
// TrimLeft(str) - returns string with leading whitespace removed
func (i *Interpreter) builtinTrimLeft(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "TrimLeft() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "TrimLeft() expects string as argument, got %s", args[0].Type())
	}

	// Use TrimLeft to remove leading whitespace
	return &StringValue{Value: strings.TrimLeft(strVal.Value, " \t\n\r")}
}

// builtinTrimRight implements the TrimRight() built-in function.
// It removes trailing whitespace from a string.
// TrimRight(str) - returns string with trailing whitespace removed
func (i *Interpreter) builtinTrimRight(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "TrimRight() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "TrimRight() expects string as argument, got %s", args[0].Type())
	}

	// Use TrimRight to remove trailing whitespace
	return &StringValue{Value: strings.TrimRight(strVal.Value, " \t\n\r")}
}

// builtinStringReplace implements the StringReplace() built-in function.
// It replaces occurrences of a substring within a string.
// StringReplace(str, old, new) - replaces all occurrences of old with new
// StringReplace(str, old, new, count) - replaces count occurrences (count=-1 means all)
func (i *Interpreter) builtinStringReplace(args []Value) Value {
	// Accept 3 or 4 arguments
	if len(args) < 3 || len(args) > 4 {
		return i.newErrorWithLocation(i.currentNode, "StringReplace() expects 3 or 4 arguments, got %d", len(args))
	}

	// First argument: string to search in
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StringReplace() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: old substring
	oldVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StringReplace() expects string as second argument, got %s", args[1].Type())
	}

	// Third argument: new substring
	newVal, ok := args[2].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StringReplace() expects string as third argument, got %s", args[2].Type())
	}

	str := strVal.Value
	old := oldVal.Value
	new := newVal.Value

	// Default count: -1 means replace all
	count := -1

	// Optional fourth argument: count
	if len(args) == 4 {
		countVal, ok := args[3].(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "StringReplace() expects integer as fourth argument, got %s", args[3].Type())
		}
		count = int(countVal.Value)
	}

	// Handle edge cases
	// Empty old string: return original (can't replace nothing)
	if len(old) == 0 {
		return &StringValue{Value: str}
	}

	// Count is 0 or negative (except -1): no replacement
	if count == 0 || (count < 0 && count != -1) {
		return &StringValue{Value: str}
	}

	// Perform replacement
	var result string
	if count == -1 {
		result = strings.ReplaceAll(str, old, new)
	} else {
		result = strings.Replace(str, old, new, count)
	}

	return &StringValue{Value: result}
}

// builtinStringOfChar implements the StringOfChar() built-in function.
// It creates a string by repeating a character N times.
// StringOfChar(ch, count) - returns a string with ch repeated count times
func (i *Interpreter) builtinStringOfChar(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StringOfChar() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: character (string)
	charVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StringOfChar() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: count (integer)
	countVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StringOfChar() expects integer as second argument, got %s", args[1].Type())
	}

	count := int(countVal.Value)

	// Handle edge cases
	// If count <= 0, return empty string
	if count <= 0 {
		return &StringValue{Value: ""}
	}

	// Extract the first character from the string
	// If the string is empty, return empty string
	if len(charVal.Value) == 0 {
		return &StringValue{Value: ""}
	}

	// Get the first character
	ch := charVal.Value[0:1]

	// Use strings.Repeat to create the repeated string
	result := strings.Repeat(ch, count)

	return &StringValue{Value: result}
}

// builtinFormat implements the Format() built-in function.
//
// Format() function for string formatting
// Supports: %s (string), %d (integer), %f (float), %% (literal %)
// Optional: width and precision (%5d, %.2f, %8.2f)
func (i *Interpreter) builtinFormat(args []Value) Value {
	// Expect exactly 2 arguments: format string and array of values
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "Format() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: format string
	fmtVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Format() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: array of values
	arrVal, ok := args[1].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Format() expects array as second argument, got %s", args[1].Type())
	}

	formatStr := fmtVal.Value
	elements := arrVal.Elements

	// Parse format string to extract format specifiers
	type formatSpec struct {
		verb  rune
		index int
	}
	var specs []formatSpec
	argIndex := 0

	iStr := 0
	for iStr < len(formatStr) {
		ch := rune(formatStr[iStr])
		if ch == '%' {
			if iStr+1 < len(formatStr) && formatStr[iStr+1] == '%' {
				// %% - literal percent sign
				iStr += 2
				continue
			}
			// Parse format specifier
			iStr++
			// Skip width/precision/flags
			for iStr < len(formatStr) {
				ch := formatStr[iStr]
				if (ch >= '0' && ch <= '9') || ch == '.' || ch == '+' || ch == '-' || ch == ' ' || ch == '#' {
					iStr++
					continue
				}
				break
			}
			// Get the verb
			if iStr < len(formatStr) {
				verb := rune(formatStr[iStr])
				if verb == 's' || verb == 'd' || verb == 'f' || verb == 'v' || verb == 'x' || verb == 'X' || verb == 'o' {
					specs = append(specs, formatSpec{verb: verb, index: argIndex})
					argIndex++
				}
				iStr++
			}
		} else {
			iStr++
		}
	}

	// Validate that we have the right number of arguments
	if len(specs) != len(elements) {
		return i.newErrorWithLocation(i.currentNode, "Format() expects %d arguments for format string, got %d", len(specs), len(elements))
	}

	// Validate types and convert DWScript values to Go interface{} values
	goArgs := make([]interface{}, len(elements))
	for idx, elem := range elements {
		if idx >= len(specs) {
			break
		}
		spec := specs[idx]

		// Unbox Variant values for Format() function
		// Since ARRAY_OF_CONST now uses VARIANT element type
		// we need to unwrap Variant values before formatting
		unwrapped := unwrapVariant(elem)

		switch v := unwrapped.(type) {
		case *IntegerValue:
			// %d, %x, %X, %o, %v are valid for integers
			switch spec.verb {
			case 'd', 'x', 'X', 'o', 'v':
				goArgs[idx] = v.Value
			case 's':
				// Allow integer to string conversion for %s
				goArgs[idx] = fmt.Sprintf("%d", v.Value)
			default:
				return i.newErrorWithLocation(i.currentNode, "Format() cannot use %%%c with Integer value at index %d", spec.verb, idx)
			}
		case *FloatValue:
			// %f, %v are valid for floats
			switch spec.verb {
			case 'f', 'v':
				goArgs[idx] = v.Value
			case 's':
				// Allow float to string conversion for %s
				goArgs[idx] = fmt.Sprintf("%f", v.Value)
			default:
				return i.newErrorWithLocation(i.currentNode, "Format() cannot use %%%c with Float value at index %d", spec.verb, idx)
			}
		case *StringValue:
			// %s, %v are valid for strings
			switch spec.verb {
			case 's', 'v':
				goArgs[idx] = v.Value
			case 'd', 'x', 'X', 'o':
				// String cannot be used with integer format specifiers
				return i.newErrorWithLocation(i.currentNode, "Format() cannot use %%%c with String value at index %d", spec.verb, idx)
			case 'f':
				// String cannot be used with float format specifiers
				return i.newErrorWithLocation(i.currentNode, "Format() cannot use %%%c with String value at index %d", spec.verb, idx)
			default:
				goArgs[idx] = v.Value
			}
		case *BooleanValue:
			goArgs[idx] = v.Value
		default:
			return i.newErrorWithLocation(i.currentNode, "Format() cannot format value of type %s at index %d", unwrapped.Type(), idx)
		}
	}

	// Format the string
	result := fmt.Sprintf(formatStr, goArgs...)

	return &StringValue{Value: result}
}

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

	// Handle edge cases for position
	// If pos < 1, insert at beginning
	// If pos > length, insert at end
	if pos < 1 {
		pos = 1
	}
	if pos > len(target)+1 {
		pos = len(target) + 1
	}

	// Build new string by inserting source at position (1-based)
	// Convert to 0-based for Go string slicing
	insertPos := pos - 1

	var newStr string
	if insertPos <= 0 {
		newStr = source + target
	} else if insertPos >= len(target) {
		newStr = target + source
	} else {
		newStr = target[:insertPos] + source + target[insertPos:]
	}

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

	// Handle edge cases
	// If pos < 1 or pos > length, do nothing (no-op)
	// If count <= 0, do nothing (no-op)
	if pos < 1 || pos > len(str) || count <= 0 {
		// No modification needed
		return &NilValue{}
	}

	// Convert to 0-based index
	startPos := pos - 1

	// Calculate end position, clamping to string length
	endPos := startPos + count
	if endPos > len(str) {
		endPos = len(str)
	}

	// Build new string by removing the substring
	var newStr string
	if startPos == 0 {
		// Delete from beginning
		newStr = str[endPos:]
	} else if endPos >= len(str) {
		// Delete to end
		newStr = str[:startPos]
	} else {
		// Delete middle section
		newStr = str[:startPos] + str[endPos:]
	}

	// Update the string variable with the new value
	newValue := &StringValue{Value: newStr}
	if err := i.env.Set(strName, newValue); err != nil {
		return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", strName, err)
	}

	return &NilValue{}
}
