package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// builtinConcat implements the Concat() built-in function.
// It concatenates multiple strings or arrays together.
// Concat(str1, str2, ...) - variable number of string arguments
// Concat(arr1, arr2, ...) - variable number of array arguments
func (i *Interpreter) builtinConcat(args []Value) Value {
	if len(args) == 0 {
		return i.newErrorWithLocation(i.currentNode, "Concat() expects at least 1 argument, got 0")
	}

	// Check if first argument is an array - if so, dispatch to array concatenation
	if _, ok := args[0].(*ArrayValue); ok {
		return i.builtinConcatArrays(args)
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
	if runeLength(charVal.Value) == 0 {
		return &StringValue{Value: ""}
	}

	// Get the first character (rune-based to handle UTF-8)
	firstRune, _ := runeAt(charVal.Value, 1)
	ch := string(firstRune)

	// Use strings.Repeat to create the repeated string
	result := strings.Repeat(ch, count)

	return &StringValue{Value: result}
}

// builtinSubStr implements the SubStr() built-in function.
// It extracts a substring from a string with a length parameter.
// SubStr(str, start) - returns substring from start to end (1-based)
// SubStr(str, start, length) - returns length characters starting at start (1-based)
// Note: Different from SubString which takes an end position instead of length.
func (i *Interpreter) builtinSubStr(args []Value) Value {
	// Accept 2 or 3 arguments
	if len(args) < 2 || len(args) > 3 {
		return i.newErrorWithLocation(i.currentNode, "SubStr() expects 2 or 3 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "SubStr() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: start position (1-based)
	startVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "SubStr() expects integer as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	start := startVal.Value // 1-based

	// Third argument (optional): length
	// Default is MaxInt (meaning "to end of string")
	length := int64(1<<31 - 1) // MaxInt for 32-bit (matches DWScript behavior)
	if len(args) == 3 {
		lengthVal, ok := args[2].(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "SubStr() expects integer as third argument, got %s", args[2].Type())
		}
		length = lengthVal.Value
	}

	// Use rune-based slicing to handle UTF-8 correctly
	// This is the same logic as Copy()
	result := runeSliceFrom(str, int(start), int(length))
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

// builtinChr implements the Chr() built-in function.
// It converts an integer character code to a single-character string.
// Chr(code: Integer): String
func (i *Interpreter) builtinChr(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Chr() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be Integer
	intVal, ok := args[0].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Chr() expects Integer argument, got %s", args[0].Type())
	}

	// Check if the code is in valid range (0-1114111 for Unicode)
	if intVal.Value < 0 || intVal.Value > 0x10FFFF {
		return i.newErrorWithLocation(i.currentNode, "Chr() code %d out of valid Unicode range (0-1114111)", intVal.Value)
	}

	// Convert to rune and then to string
	return &StringValue{Value: string(rune(intVal.Value))}
}

// builtinIntToHex implements the IntToHex() built-in function.
// It converts an integer to a hexadecimal string with specified minimum number of digits.
// IntToHex(value: Integer, digits: Integer): String
func (i *Interpreter) builtinIntToHex(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "IntToHex() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument must be Integer (the value to convert)
	intVal, ok := args[0].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IntToHex() first argument must be Integer, got %s", args[0].Type())
	}

	// Second argument must be Integer (minimum number of digits)
	digitsVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IntToHex() second argument must be Integer, got %s", args[1].Type())
	}

	// Convert to hexadecimal string with uppercase letters
	hexStr := fmt.Sprintf("%X", uint64(intVal.Value))

	// Pad with zeros if necessary to reach minimum digit count
	if digitsVal.Value > 0 && int64(len(hexStr)) < digitsVal.Value {
		// Pad with leading zeros
		hexStr = strings.Repeat("0", int(digitsVal.Value)-len(hexStr)) + hexStr
	}

	return &StringValue{Value: hexStr}
}

// builtinStrToBool implements the StrToBool() built-in function.
// It converts a string to a boolean value.
// Accepts: 'True', 'False', '1', '0', 'Yes', 'No' (case-insensitive)
// StrToBool(s: String): Boolean
func (i *Interpreter) builtinStrToBool(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "StrToBool() expects exactly 1 argument, got %d", len(args))
	}

	// First argument must be String
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToBool() expects String, got %s", args[0].Type())
	}

	// Normalize to lowercase for case-insensitive matching
	s := strings.ToLower(strings.TrimSpace(strVal.Value))

	// Check for true values
	switch s {
	case "true", "1", "yes", "t", "y":
		return &BooleanValue{Value: true}
	case "false", "0", "no", "f", "n":
		return &BooleanValue{Value: false}
	default:
		return i.newErrorWithLocation(i.currentNode, "StrToBool() invalid boolean string: '%s'", strVal.Value)
	}
}

// builtinSubString implements the SubString() built-in function.
// It extracts a substring from a string using start and end positions.
// SubString(str, start, end) - returns substring from start to end (1-based, inclusive)
// Note: Different from SubStr which takes a length parameter instead of end position.
func (i *Interpreter) builtinSubString(args []Value) Value {
	if len(args) != 3 {
		return i.newErrorWithLocation(i.currentNode, "SubString() expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "SubString() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: start position (1-based)
	startVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "SubString() expects integer as second argument, got %s", args[1].Type())
	}

	// Third argument: end position (1-based, inclusive)
	endVal, ok := args[2].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "SubString() expects integer as third argument, got %s", args[2].Type())
	}

	str := strVal.Value
	start := int(startVal.Value) // 1-based
	end := int(endVal.Value)     // 1-based, inclusive

	// Calculate length from start and end positions
	// SubString(str, 3, 7) should return 5 characters (positions 3, 4, 5, 6, 7)
	length := end - start + 1

	// Handle edge cases
	if length <= 0 {
		return &StringValue{Value: ""}
	}

	// Use rune-based slicing to handle UTF-8 correctly
	result := runeSliceFrom(str, start, length)
	return &StringValue{Value: result}
}

// builtinLeftStr implements the LeftStr() built-in function.
// It returns the leftmost N characters of a string.
// LeftStr(str, count) - returns first count characters
func (i *Interpreter) builtinLeftStr(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "LeftStr() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "LeftStr() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: count
	countVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "LeftStr() expects integer as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	count := int(countVal.Value)

	// Handle edge cases
	if count <= 0 {
		return &StringValue{Value: ""}
	}

	// Use rune-based slicing to handle UTF-8 correctly
	// LeftStr is equivalent to SubStr(str, 1, count)
	result := runeSliceFrom(str, 1, count)
	return &StringValue{Value: result}
}

// builtinRightStr implements the RightStr() built-in function.
// It returns the rightmost N characters of a string.
// RightStr(str, count) - returns last count characters
func (i *Interpreter) builtinRightStr(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "RightStr() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "RightStr() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: count
	countVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "RightStr() expects integer as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	count := int(countVal.Value)

	// Handle edge cases
	if count <= 0 {
		return &StringValue{Value: ""}
	}

	// Get the length of the string in runes (not bytes)
	strLen := runeLength(str)

	// If count >= length, return the whole string
	if count >= strLen {
		return &StringValue{Value: str}
	}

	// Calculate start position (1-based)
	// For a string of length 10, RightStr(str, 3) should return positions 8, 9, 10
	start := strLen - count + 1

	// Use rune-based slicing to handle UTF-8 correctly
	result := runeSliceFrom(str, start, count)
	return &StringValue{Value: result}
}

// builtinMidStr implements the MidStr() built-in function.
// It is an alias for SubStr - extracts a substring with a length parameter.
// MidStr(str, start, count) - returns count characters starting at start (1-based)
func (i *Interpreter) builtinMidStr(args []Value) Value {
	// MidStr is just an alias for SubStr
	return i.builtinSubStr(args)
}

// builtinStrBeginsWith implements the StrBeginsWith() built-in function.
// It checks if a string starts with a given prefix.
// StrBeginsWith(str, prefix) - returns true if str starts with prefix
func (i *Interpreter) builtinStrBeginsWith(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrBeginsWith() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string to check
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrBeginsWith() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: prefix
	prefixVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrBeginsWith() expects string as second argument, got %s", args[1].Type())
	}

	result := strings.HasPrefix(strVal.Value, prefixVal.Value)
	return &BooleanValue{Value: result}
}

// builtinStrEndsWith implements the StrEndsWith() built-in function.
// It checks if a string ends with a given suffix.
// StrEndsWith(str, suffix) - returns true if str ends with suffix
func (i *Interpreter) builtinStrEndsWith(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrEndsWith() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string to check
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrEndsWith() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: suffix
	suffixVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrEndsWith() expects string as second argument, got %s", args[1].Type())
	}

	result := strings.HasSuffix(strVal.Value, suffixVal.Value)
	return &BooleanValue{Value: result}
}

// builtinStrContains implements the StrContains() built-in function.
// It checks if a string contains a given substring.
// StrContains(str, substring) - returns true if str contains substring
func (i *Interpreter) builtinStrContains(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrContains() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string to check
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrContains() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: substring
	substrVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrContains() expects string as second argument, got %s", args[1].Type())
	}

	result := strings.Contains(strVal.Value, substrVal.Value)
	return &BooleanValue{Value: result}
}

// builtinPosEx implements the PosEx() built-in function.
// It finds the position of a substring within a string, starting from an offset.
// PosEx(needle, haystack, offset) - returns 1-based position (0 if not found)
func (i *Interpreter) builtinPosEx(args []Value) Value {
	if len(args) != 3 {
		return i.newErrorWithLocation(i.currentNode, "PosEx() expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: substring to find (needle)
	needleVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "PosEx() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string to search in (haystack)
	haystackVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "PosEx() expects string as second argument, got %s", args[1].Type())
	}

	// Third argument: offset (1-based starting position)
	offsetVal, ok := args[2].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "PosEx() expects integer as third argument, got %s", args[2].Type())
	}

	needle := needleVal.Value
	haystack := haystackVal.Value
	offset := int(offsetVal.Value) // 1-based

	// Handle invalid offset first (before empty needle check)
	// This prevents returning negative positions
	if offset < 1 {
		return &IntegerValue{Value: 0}
	}

	// Handle empty needle - returns 0 (not found)
	// This matches the original DWScript behavior
	if len(needle) == 0 {
		return &IntegerValue{Value: 0}
	}

	// Convert to rune-based indexing for UTF-8 support
	haystackRunes := []rune(haystack)
	needleRunes := []rune(needle)

	// Adjust offset to 0-based
	startIdx := offset - 1

	// If offset is beyond the string length, not found
	if startIdx >= len(haystackRunes) {
		return &IntegerValue{Value: 0}
	}

	// Search for the needle starting from offset
	for i := startIdx; i <= len(haystackRunes)-len(needleRunes); i++ {
		match := true
		for j := 0; j < len(needleRunes); j++ {
			if haystackRunes[i+j] != needleRunes[j] {
				match = false
				break
			}
		}
		if match {
			// Return 1-based position
			return &IntegerValue{Value: int64(i + 1)}
		}
	}

	// Not found
	return &IntegerValue{Value: 0}
}

// builtinRevPos implements the RevPos() built-in function.
// It finds the last position of a substring within a string.
// RevPos(needle, haystack) - returns 1-based position of last occurrence (0 if not found)
func (i *Interpreter) builtinRevPos(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "RevPos() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: substring to find (needle)
	needleVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "RevPos() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string to search in (haystack)
	haystackVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "RevPos() expects string as second argument, got %s", args[1].Type())
	}

	needle := needleVal.Value
	haystack := haystackVal.Value

	// Handle empty needle - returns length + 1 (not found behavior for DWScript)
	if len(needle) == 0 {
		return &IntegerValue{Value: int64(runeLength(haystack) + 1)}
	}

	// Find the last occurrence using strings.LastIndex
	index := strings.LastIndex(haystack, needle)

	// Convert to 1-based index (or 0 if not found)
	if index == -1 {
		return &IntegerValue{Value: 0}
	}

	// Convert byte index to rune index (for UTF-8 support)
	runeIndex := len([]rune(haystack[:index])) + 1
	return &IntegerValue{Value: int64(runeIndex)}
}

// builtinStrFind implements the StrFind() built-in function.
// It is an alias for PosEx - finds substring with starting index.
// StrFind(str, substr, fromIndex) - returns 1-based position (0 if not found)
func (i *Interpreter) builtinStrFind(args []Value) Value {
	if len(args) != 3 {
		return i.newErrorWithLocation(i.currentNode, "StrFind() expects exactly 3 arguments, got %d", len(args))
	}

	// StrFind(str, substr, fromIndex) maps to PosEx(substr, str, fromIndex)
	// Need to reorder arguments
	reorderedArgs := []Value{
		args[1], // substr becomes first arg (needle)
		args[0], // str becomes second arg (haystack)
		args[2], // fromIndex stays as third arg (offset)
	}

	return i.builtinPosEx(reorderedArgs)
}

// builtinStrSplit implements the StrSplit() built-in function.
// It splits a string into an array using a delimiter.
// StrSplit(str, delimiter) - returns array of strings
func (i *Interpreter) builtinStrSplit(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrSplit() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string to split
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrSplit() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrSplit() expects string as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	delim := delimVal.Value

	// Handle empty delimiter - return array with single element (the original string)
	if len(delim) == 0 {
		elements := []Value{&StringValue{Value: str}}
		return &ArrayValue{
			Elements:  elements,
			ArrayType: types.NewDynamicArrayType(types.STRING),
		}
	}

	// Split the string
	parts := strings.Split(str, delim)

	// Convert to array of StringValue
	elements := make([]Value, len(parts))
	for idx, part := range parts {
		elements[idx] = &StringValue{Value: part}
	}

	return &ArrayValue{
		Elements:  elements,
		ArrayType: types.NewDynamicArrayType(types.STRING),
	}
}

// builtinStrJoin implements the StrJoin() built-in function.
// It joins an array of strings into a single string using a delimiter.
// StrJoin(array, delimiter) - returns joined string
func (i *Interpreter) builtinStrJoin(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrJoin() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: array of strings
	arrVal, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrJoin() expects array as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrJoin() expects string as second argument, got %s", args[1].Type())
	}

	delim := delimVal.Value

	// Convert array elements to strings
	parts := make([]string, len(arrVal.Elements))
	for idx, elem := range arrVal.Elements {
		strElem, ok := elem.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "StrJoin() expects array of strings, got %s at index %d", elem.Type(), idx)
		}
		parts[idx] = strElem.Value
	}

	// Join the strings
	result := strings.Join(parts, delim)
	return &StringValue{Value: result}
}

// builtinStrArrayPack implements the StrArrayPack() built-in function.
// It removes empty strings from an array.
// StrArrayPack(array) - returns array with empty strings removed
func (i *Interpreter) builtinStrArrayPack(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "StrArrayPack() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: array of strings
	arrVal, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrArrayPack() expects array as argument, got %s", args[0].Type())
	}

	// Filter out empty strings
	var packed []Value
	for _, elem := range arrVal.Elements {
		strElem, ok := elem.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "StrArrayPack() expects array of strings, got %s", elem.Type())
		}
		if strElem.Value != "" {
			packed = append(packed, strElem)
		}
	}

	return &ArrayValue{
		Elements:  packed,
		ArrayType: types.NewDynamicArrayType(types.STRING),
	}
}

// builtinStrBefore implements the StrBefore() built-in function.
// It returns the substring before the first occurrence of a delimiter.
// StrBefore(str, delimiter) - returns substring before first delimiter (empty if not found)
func (i *Interpreter) builtinStrBefore(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrBefore() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrBefore() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrBefore() expects string as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	delim := delimVal.Value

	// Handle empty delimiter - return empty string
	if len(delim) == 0 {
		return &StringValue{Value: ""}
	}

	// Find the first occurrence of delimiter
	index := strings.Index(str, delim)
	if index == -1 {
		// Delimiter not found - return empty string
		return &StringValue{Value: ""}
	}

	// Return substring before delimiter
	return &StringValue{Value: str[:index]}
}

// builtinStrBeforeLast implements the StrBeforeLast() built-in function.
// It returns the substring before the last occurrence of a delimiter.
// StrBeforeLast(str, delimiter) - returns substring before last delimiter (empty if not found)
func (i *Interpreter) builtinStrBeforeLast(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrBeforeLast() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrBeforeLast() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrBeforeLast() expects string as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	delim := delimVal.Value

	// Handle empty delimiter - return empty string
	if len(delim) == 0 {
		return &StringValue{Value: ""}
	}

	// Find the last occurrence of delimiter
	index := strings.LastIndex(str, delim)
	if index == -1 {
		// Delimiter not found - return empty string
		return &StringValue{Value: ""}
	}

	// Return substring before last delimiter
	return &StringValue{Value: str[:index]}
}

// builtinStrAfter implements the StrAfter() built-in function.
// It returns the substring after the first occurrence of a delimiter.
// StrAfter(str, delimiter) - returns substring after first delimiter (empty if not found)
func (i *Interpreter) builtinStrAfter(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrAfter() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrAfter() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrAfter() expects string as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	delim := delimVal.Value

	// Handle empty delimiter - return empty string
	if len(delim) == 0 {
		return &StringValue{Value: ""}
	}

	// Find the first occurrence of delimiter
	index := strings.Index(str, delim)
	if index == -1 {
		// Delimiter not found - return empty string
		return &StringValue{Value: ""}
	}

	// Return substring after delimiter
	return &StringValue{Value: str[index+len(delim):]}
}

// builtinStrAfterLast implements the StrAfterLast() built-in function.
// It returns the substring after the last occurrence of a delimiter.
// StrAfterLast(str, delimiter) - returns substring after last delimiter (empty if not found)
func (i *Interpreter) builtinStrAfterLast(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrAfterLast() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrAfterLast() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrAfterLast() expects string as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	delim := delimVal.Value

	// Handle empty delimiter - return empty string
	if len(delim) == 0 {
		return &StringValue{Value: ""}
	}

	// Find the last occurrence of delimiter
	index := strings.LastIndex(str, delim)
	if index == -1 {
		// Delimiter not found - return empty string
		return &StringValue{Value: ""}
	}

	// Return substring after last delimiter
	return &StringValue{Value: str[index+len(delim):]}
}

// builtinStrBetween implements the StrBetween() built-in function.
// It returns the substring between first occurrence of start and first occurrence of stop after start.
// StrBetween(str, start, stop) - returns substring between start and stop delimiters
func (i *Interpreter) builtinStrBetween(args []Value) Value {
	if len(args) != 3 {
		return i.newErrorWithLocation(i.currentNode, "StrBetween() expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrBetween() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: start delimiter
	startVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrBetween() expects string as second argument, got %s", args[1].Type())
	}

	// Third argument: stop delimiter
	stopVal, ok := args[2].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrBetween() expects string as third argument, got %s", args[2].Type())
	}

	str := strVal.Value
	start := startVal.Value
	stop := stopVal.Value

	// Handle empty delimiters - return empty string
	if len(start) == 0 || len(stop) == 0 {
		return &StringValue{Value: ""}
	}

	// Find the first occurrence of start delimiter
	startIdx := strings.Index(str, start)
	if startIdx == -1 {
		// Start delimiter not found - return empty string
		return &StringValue{Value: ""}
	}

	// Search for stop delimiter after the start delimiter
	searchFrom := startIdx + len(start)
	if searchFrom >= len(str) {
		// No room for stop delimiter - return empty string
		return &StringValue{Value: ""}
	}

	stopIdx := strings.Index(str[searchFrom:], stop)
	if stopIdx == -1 {
		// Stop delimiter not found - return empty string
		return &StringValue{Value: ""}
	}

	// Adjust stopIdx to be relative to the original string
	stopIdx += searchFrom

	// Return substring between start and stop delimiters
	return &StringValue{Value: str[searchFrom:stopIdx]}
}

// builtinIsDelimiter implements the IsDelimiter() built-in function.
// It checks if the character at a given position is one of the specified delimiters.
// IsDelimiter(delims, str, index) - returns true if char at index is a delimiter (1-based index)
func (i *Interpreter) builtinIsDelimiter(args []Value) Value {
	if len(args) != 3 {
		return i.newErrorWithLocation(i.currentNode, "IsDelimiter() expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: delimiter characters
	delimsVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IsDelimiter() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string to check
	strVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IsDelimiter() expects string as second argument, got %s", args[1].Type())
	}

	// Third argument: index (1-based)
	indexVal, ok := args[2].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IsDelimiter() expects integer as third argument, got %s", args[2].Type())
	}

	delims := delimsVal.Value
	str := strVal.Value
	index := int(indexVal.Value) // 1-based

	// Handle invalid index
	if index < 1 {
		return &BooleanValue{Value: false}
	}

	// Convert to rune-based indexing for UTF-8 support
	strRunes := []rune(str)

	// Check if index is within bounds (1-based)
	if index > len(strRunes) {
		return &BooleanValue{Value: false}
	}

	// Get the character at the specified position (convert to 0-based)
	ch := strRunes[index-1]

	// Check if the character is in the delimiter string
	result := strings.ContainsRune(delims, ch)
	return &BooleanValue{Value: result}
}

// builtinLastDelimiter implements the LastDelimiter() built-in function.
// It finds the position of the last occurrence of any delimiter character.
// LastDelimiter(delims, str) - returns 1-based position of last delimiter (0 if not found)
func (i *Interpreter) builtinLastDelimiter(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "LastDelimiter() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: delimiter characters
	delimsVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "LastDelimiter() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string to search
	strVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "LastDelimiter() expects string as second argument, got %s", args[1].Type())
	}

	delims := delimsVal.Value
	str := strVal.Value

	// Convert to rune-based for UTF-8 support
	strRunes := []rune(str)

	// Search from the end for any delimiter character
	for i := len(strRunes) - 1; i >= 0; i-- {
		if strings.ContainsRune(delims, strRunes[i]) {
			// Return 1-based position
			return &IntegerValue{Value: int64(i + 1)}
		}
	}

	// No delimiter found
	return &IntegerValue{Value: 0}
}

// builtinFindDelimiter implements the FindDelimiter() built-in function.
// It finds the position of the first occurrence of any delimiter character, starting from an index.
// FindDelimiter(delims, str, startIndex) - returns 1-based position (0 if not found)
func (i *Interpreter) builtinFindDelimiter(args []Value) Value {
	if len(args) != 3 {
		return i.newErrorWithLocation(i.currentNode, "FindDelimiter() expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: delimiter characters
	delimsVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "FindDelimiter() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string to search
	strVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "FindDelimiter() expects string as second argument, got %s", args[1].Type())
	}

	// Third argument: start index (1-based)
	startIndexVal, ok := args[2].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "FindDelimiter() expects integer as third argument, got %s", args[2].Type())
	}

	delims := delimsVal.Value
	str := strVal.Value
	startIndex := int(startIndexVal.Value) // 1-based

	// Handle invalid start index
	if startIndex < 1 {
		return &IntegerValue{Value: 0}
	}

	// Convert to rune-based for UTF-8 support
	strRunes := []rune(str)

	// Adjust to 0-based index
	startIdx := startIndex - 1

	// Check if start index is within bounds
	if startIdx >= len(strRunes) {
		return &IntegerValue{Value: 0}
	}

	// Search from startIdx for any delimiter character
	for i := startIdx; i < len(strRunes); i++ {
		if strings.ContainsRune(delims, strRunes[i]) {
			// Return 1-based position
			return &IntegerValue{Value: int64(i + 1)}
		}
	}

	// No delimiter found
	return &IntegerValue{Value: 0}
}
