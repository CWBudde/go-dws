package semantic

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
)

// AnalysisError represents one or more semantic analysis errors
type AnalysisError struct {
	Errors []string
}

// Error returns a formatted error message containing all semantic errors
func (e *AnalysisError) Error() string {
	if len(e.Errors) == 0 {
		return "semantic analysis failed"
	}

	if len(e.Errors) == 1 {
		return fmt.Sprintf("semantic error: %s", e.Errors[0])
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("semantic analysis failed with %d errors:\n", len(e.Errors)))
	for i, err := range e.Errors {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err))
	}

	return sb.String()
}

// SemanticErrorType classifies the type of semantic error
type SemanticErrorType string

const (
	ErrorTypeMismatch       SemanticErrorType = "type_mismatch"
	ErrorUndefinedVariable  SemanticErrorType = "undefined_variable"
	ErrorUndefinedFunction  SemanticErrorType = "undefined_function"
	ErrorUndefinedType      SemanticErrorType = "undefined_type"
	ErrorRedeclaration      SemanticErrorType = "redeclaration"
	ErrorInvalidOperation   SemanticErrorType = "invalid_operation"
	ErrorVisibility         SemanticErrorType = "visibility"
	ErrorConstantModified   SemanticErrorType = "constant_modified"
	ErrorInvalidAssignment  SemanticErrorType = "invalid_assignment"
	ErrorInvalidReturn      SemanticErrorType = "invalid_return"
	ErrorInvalidBreak       SemanticErrorType = "invalid_break"
	ErrorInvalidContinue    SemanticErrorType = "invalid_continue"
	ErrorMissingReturn      SemanticErrorType = "missing_return"
	ErrorArgumentCount      SemanticErrorType = "argument_count"
	ErrorInheritance        SemanticErrorType = "inheritance"
	ErrorAbstractClass      SemanticErrorType = "abstract_class"
	ErrorInterface          SemanticErrorType = "interface"
	ErrorGeneric            SemanticErrorType = "generic"
)

// SemanticError represents a structured semantic/compile-time error
type SemanticError struct {
	Type         SemanticErrorType
	Message      string
	Pos          lexer.Position
	Expected     types.Type
	Got          types.Type
	VariableName string
	FunctionName string
	TypeName     string
	ClassName    string
	Context      map[string]interface{}
}

// Error implements the error interface
func (e *SemanticError) Error() string {
	return fmt.Sprintf("%s at %s", e.Message, e.Pos.String())
}

// ToCompilerError converts a SemanticError to a CompilerError for display
func (e *SemanticError) ToCompilerError(source, filename string) *errors.CompilerError {
	message := e.Message

	// Add additional context based on error type
	switch e.Type {
	case ErrorTypeMismatch:
		if e.Expected != nil && e.Got != nil {
			message = fmt.Sprintf("%s\nExpected: %s\nGot: %s",
				e.Message, e.Expected.String(), e.Got.String())
		}
	case ErrorUndefinedVariable:
		if e.VariableName != "" {
			message = fmt.Sprintf("Undefined variable '%s'", e.VariableName)
		}
	case ErrorUndefinedFunction:
		if e.FunctionName != "" {
			message = fmt.Sprintf("Undefined function '%s'", e.FunctionName)
		}
	case ErrorUndefinedType:
		if e.TypeName != "" {
			message = fmt.Sprintf("Undefined type '%s'", e.TypeName)
		}
	}

	return errors.NewCompilerError(e.Pos, message, source, filename)
}

// NewTypeMismatch creates a type mismatch error
func NewTypeMismatch(pos lexer.Position, varName string, expected, got types.Type) *SemanticError {
	message := fmt.Sprintf("Cannot assign %s to %s", got.String(), expected.String())
	if varName != "" {
		message = fmt.Sprintf("Cannot assign %s to %s variable '%s'",
			got.String(), expected.String(), varName)
	}

	return &SemanticError{
		Type:         ErrorTypeMismatch,
		Message:      message,
		Pos:          pos,
		Expected:     expected,
		Got:          got,
		VariableName: varName,
	}
}

// NewOperatorTypeMismatch creates an operator type mismatch error
func NewOperatorTypeMismatch(pos lexer.Position, operator string, left, right types.Type) *SemanticError {
	message := fmt.Sprintf("Invalid operation: %s %s %s",
		left.String(), operator, right.String())

	return &SemanticError{
		Type:     ErrorTypeMismatch,
		Message:  message,
		Pos:      pos,
		Expected: nil,
		Got:      nil,
	}
}

// NewUndefinedVariable creates an undefined variable error
func NewUndefinedVariable(pos lexer.Position, varName string) *SemanticError {
	return &SemanticError{
		Type:         ErrorUndefinedVariable,
		Message:      fmt.Sprintf("Undefined variable '%s'", varName),
		Pos:          pos,
		VariableName: varName,
	}
}

// NewUndefinedFunction creates an undefined function error
func NewUndefinedFunction(pos lexer.Position, funcName string) *SemanticError {
	return &SemanticError{
		Type:         ErrorUndefinedFunction,
		Message:      fmt.Sprintf("Undefined function '%s'", funcName),
		Pos:          pos,
		FunctionName: funcName,
	}
}

// NewUndefinedType creates an undefined type error
func NewUndefinedType(pos lexer.Position, typeName string) *SemanticError {
	return &SemanticError{
		Type:     ErrorUndefinedType,
		Message:  fmt.Sprintf("Undefined type '%s'", typeName),
		Pos:      pos,
		TypeName: typeName,
	}
}

// NewRedeclaration creates a redeclaration error
func NewRedeclaration(pos lexer.Position, name string) *SemanticError {
	return &SemanticError{
		Type:    ErrorRedeclaration,
		Message: fmt.Sprintf("'%s' is already declared", name),
		Pos:     pos,
	}
}

// NewInvalidOperation creates an invalid operation error
func NewInvalidOperation(pos lexer.Position, message string) *SemanticError {
	return &SemanticError{
		Type:    ErrorInvalidOperation,
		Message: message,
		Pos:     pos,
	}
}

// NewVisibilityError creates a visibility error
func NewVisibilityError(pos lexer.Position, member, className string) *SemanticError {
	return &SemanticError{
		Type:      ErrorVisibility,
		Message:   fmt.Sprintf("Cannot access private member '%s' of class '%s'", member, className),
		Pos:       pos,
		ClassName: className,
	}
}

// NewConstantModified creates a constant modification error
func NewConstantModified(pos lexer.Position, constName string) *SemanticError {
	return &SemanticError{
		Type:         ErrorConstantModified,
		Message:      fmt.Sprintf("Cannot assign to constant '%s'", constName),
		Pos:          pos,
		VariableName: constName,
	}
}

// NewInvalidAssignment creates an invalid assignment error
func NewInvalidAssignment(pos lexer.Position, message string) *SemanticError {
	return &SemanticError{
		Type:    ErrorInvalidAssignment,
		Message: message,
		Pos:     pos,
	}
}

// NewInvalidReturn creates an invalid return statement error
func NewInvalidReturn(pos lexer.Position, expected, got types.Type) *SemanticError {
	message := "Return type mismatch"
	if expected != nil && got != nil {
		message = fmt.Sprintf("Cannot return %s from function returning %s",
			got.String(), expected.String())
	}

	return &SemanticError{
		Type:     ErrorInvalidReturn,
		Message:  message,
		Pos:      pos,
		Expected: expected,
		Got:      got,
	}
}

// NewInvalidBreak creates an invalid break statement error
func NewInvalidBreak(pos lexer.Position) *SemanticError {
	return &SemanticError{
		Type:    ErrorInvalidBreak,
		Message: "Break statement outside of loop",
		Pos:     pos,
	}
}

// NewInvalidContinue creates an invalid continue statement error
func NewInvalidContinue(pos lexer.Position) *SemanticError {
	return &SemanticError{
		Type:    ErrorInvalidContinue,
		Message: "Continue statement outside of loop",
		Pos:     pos,
	}
}

// NewMissingReturn creates a missing return statement error
func NewMissingReturn(pos lexer.Position, funcName string) *SemanticError {
	return &SemanticError{
		Type:         ErrorMissingReturn,
		Message:      fmt.Sprintf("Function '%s' must return a value", funcName),
		Pos:          pos,
		FunctionName: funcName,
	}
}

// NewArgumentCountError creates an argument count mismatch error
func NewArgumentCountError(pos lexer.Position, funcName string, expected, got int) *SemanticError {
	return &SemanticError{
		Type:         ErrorArgumentCount,
		Message:      fmt.Sprintf("Function '%s' expects %d arguments, got %d", funcName, expected, got),
		Pos:          pos,
		FunctionName: funcName,
		Context: map[string]interface{}{
			"expected": expected,
			"got":      got,
		},
	}
}

// NewInheritanceError creates an inheritance error
func NewInheritanceError(pos lexer.Position, message string) *SemanticError {
	return &SemanticError{
		Type:    ErrorInheritance,
		Message: message,
		Pos:     pos,
	}
}

// NewAbstractClassError creates an abstract class instantiation error
func NewAbstractClassError(pos lexer.Position, className string) *SemanticError {
	return &SemanticError{
		Type:      ErrorAbstractClass,
		Message:   fmt.Sprintf("Cannot instantiate abstract class '%s'", className),
		Pos:       pos,
		ClassName: className,
	}
}

// NewInterfaceError creates an interface error
func NewInterfaceError(pos lexer.Position, message string) *SemanticError {
	return &SemanticError{
		Type:    ErrorInterface,
		Message: message,
		Pos:     pos,
	}
}

// NewGenericError creates a generic semantic error
func NewGenericError(pos lexer.Position, message string) *SemanticError {
	return &SemanticError{
		Type:    ErrorGeneric,
		Message: message,
		Pos:     pos,
	}
}
