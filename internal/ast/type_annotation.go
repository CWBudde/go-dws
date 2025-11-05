package ast

import "github.com/cwbudde/go-dws/internal/lexer"

// TypeAnnotation represents a type annotation in the AST.
// This is used for variable declarations, parameters, and return types.
// Example: `: Integer` in `var x: Integer := 5;`
type TypeAnnotation struct {
	Name       string
	Token      lexer.Token
	InlineType TypeExpression // For complex inline types (arrays, function pointers) that need AST evaluation
}

// String returns the string representation of the type annotation
func (ta *TypeAnnotation) String() string {
	if ta == nil {
		return ""
	}
	return ta.Name
}

// TokenLiteral returns the literal value of the token
func (ta *TypeAnnotation) TokenLiteral() string {
	return ta.Token.Literal
}

// Pos returns the position of the type annotation
func (ta *TypeAnnotation) Pos() lexer.Position {
	return ta.Token.Pos
}

// typeExpressionNode marks this as a type expression
func (ta *TypeAnnotation) typeExpressionNode() {}

// TypedExpression is an interface that all expressions with type information must implement.
// This allows the semantic analyzer to attach type information to expressions.
type TypedExpression interface {
	Expression
	// GetType returns the type of this expression (nil if not yet determined)
	GetType() *TypeAnnotation
	// SetType sets the type of this expression
	SetType(typ *TypeAnnotation)
}

// ============================================================================
// Type Declarations
// ============================================================================

// TypeDeclaration represents a type declaration statement in DWScript.
// This can be either a type alias, subrange type, or a full type definition.
//
// Type alias examples:
//
//	type TUserID = Integer;
//	type TFileName = String;
//	type TIntArray = array of Integer;
//
// Subrange type examples
//
//	type TDigit = 0..9;
//	type TPercent = 0..100;
//	type TTemperature = -40..50;
//
// Full type definitions (enums, records, classes) will use specialized
// declaration nodes (EnumDecl, RecordDecl, ClassDecl) but may eventually
// be unified under this node.
//
// Task 9.15: This node currently supports type aliases. Future tasks will
// extend it to handle all type declarations.
// Task 9.94: Extended to support subrange types with IsSubrange, LowBound, and HighBound.
// Task 9.155: Extended to support function pointer types with FunctionPointerType and IsFunctionPointer.
type TypeDeclaration struct {
	Name                *Identifier
	AliasedType         *TypeAnnotation
	LowBound            Expression               // For subrange types
	HighBound           Expression               // For subrange types
	FunctionPointerType *FunctionPointerTypeNode // For function/procedure pointer types
	Token               lexer.Token
	IsAlias             bool
	IsSubrange          bool // For subrange types
	IsFunctionPointer   bool // For function/procedure pointer types
}

func (td *TypeDeclaration) statementNode()       {}
func (td *TypeDeclaration) TokenLiteral() string { return td.Token.Literal }
func (td *TypeDeclaration) Pos() lexer.Position  { return td.Token.Pos }

// String returns the string representation of the type declaration.
// For type aliases, this returns: "type Name = Type;"
// For subrange types, this returns: "type Name = LowBound..HighBound"
// For function pointer types, this returns: "type Name = function(...): ReturnType"
// For full type definitions, this will be extended in future tasks.
func (td *TypeDeclaration) String() string {
	if td.IsSubrange {
		// Subrange type: type TDigit = 0..9;
		// Task 9.94: Format as "type Name = Low..High"
		return "type " + td.Name.String() + " = " + td.LowBound.String() + ".." + td.HighBound.String()
	}

	if td.IsFunctionPointer {
		// Function pointer type: type TFunc = function(x: Integer): Boolean;
		// Task 9.155: Format as "type Name = function/procedure..."
		return "type " + td.Name.String() + " = " + td.FunctionPointerType.String()
	}

	if td.IsAlias {
		// Type alias: type TUserID = Integer;
		return "type " + td.Name.String() + " = " + td.AliasedType.String()
	}

	// For now, only aliases, subranges, and function pointers are supported
	// Future: Handle full type definitions
	return "type " + td.Name.String()
}
