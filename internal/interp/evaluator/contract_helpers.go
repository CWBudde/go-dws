package evaluator

import "strings"

// cleanContractMessage removes unnecessary parentheses from contract expressions
// for display in error messages. The AST's String() method adds structural parentheses
// that make error messages harder to read.
func cleanContractMessage(message string) string {
	// Strip all outer parentheses pairs
	for len(message) > 2 && message[0] == '(' && message[len(message)-1] == ')' {
		message = message[1 : len(message)-1]
	}
	// Remove parentheses after binary operators to make messages more readable
	// e.g., "Result = (old val + 1)" -> "Result = old val + 1"
	message = strings.ReplaceAll(message, " = (", " = ")
	message = strings.ReplaceAll(message, " <> (", " <> ")
	message = strings.ReplaceAll(message, " < (", " < ")
	message = strings.ReplaceAll(message, " > (", " > ")
	message = strings.ReplaceAll(message, " <= (", " <= ")
	message = strings.ReplaceAll(message, " >= (", " >= ")
	message = strings.ReplaceAll(message, " + (", " + ")
	message = strings.ReplaceAll(message, " - (", " - ")
	message = strings.ReplaceAll(message, " * (", " * ")
	message = strings.ReplaceAll(message, " / (", " / ")
	message = strings.ReplaceAll(message, " div (", " div ")
	message = strings.ReplaceAll(message, " mod (", " mod ")
	message = strings.ReplaceAll(message, " and (", " and ")
	message = strings.ReplaceAll(message, " or (", " or ")
	message = strings.ReplaceAll(message, " xor (", " xor ")
	// Remove matching trailing parentheses that were left over
	for strings.Count(message, "(") < strings.Count(message, ")") {
		lastParen := strings.LastIndex(message, ")")
		if lastParen >= 0 {
			message = message[:lastParen] + message[lastParen+1:]
		} else {
			break
		}
	}
	return message
}
