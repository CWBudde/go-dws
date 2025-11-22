package ast

import "github.com/cwbudde/go-dws/pkg/token"

// TypeAnnotation represents a type annotation in the AST.
// This is used for variable declarations, parameters, and return types.
// Example: `: Integer` in `var x: Integer := 5;`
type TypeAnnotation struct {
	InlineType TypeExpression
	Name       string
	Token      token.Token
	EndPos     token.Position
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
func (ta *TypeAnnotation) Pos() token.Position {
	return ta.Token.Pos
}

// End returns the end position of the type annotation
func (ta *TypeAnnotation) End() token.Position {
	if ta.EndPos.Line != 0 {
		return ta.EndPos
	}
	pos := ta.Token.Pos
	pos.Column += len(ta.Name)
	pos.Offset += len(ta.Name)
	return pos
}

// typeExpressionNode marks this as a type expression
func (ta *TypeAnnotation) typeExpressionNode() {}

// TypedExpression was previously an interface for expressions with type information.
// Type information is stored separately in SemanticInfo. TODO: Check if still needed!
// This interface is kept for backward compatibility but now just aliases Expression.
// Code that needs type information should use SemanticInfo.GetType(expr) instead.
type TypedExpression interface {
	Expression
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
type TypeDeclaration struct {
	AliasedType         TypeExpression
	LowBound            Expression
	HighBound           Expression
	Name                *Identifier
	FunctionPointerType *FunctionPointerTypeNode
	BaseNode
	IsAlias           bool
	IsSubrange        bool
	IsFunctionPointer bool
}

func (t *TypeDeclaration) End() token.Position {
	if t.EndPos.Line != 0 {
		return t.EndPos
	}
	return t.Token.Pos
}

func (td *TypeDeclaration) statementNode()       {}
func (td *TypeDeclaration) TokenLiteral() string { return td.Token.Literal }
func (td *TypeDeclaration) Pos() token.Position  { return td.Token.Pos }

// String returns the string representation of the type declaration.
// For type aliases, this returns: "type Name = Type;"
// For subrange types, this returns: "type Name = LowBound..HighBound"
// For function pointer types, this returns: "type Name = function(...): ReturnType"
// For full type definitions, this will be extended in future tasks.
func (td *TypeDeclaration) String() string {
	if td.IsSubrange {
		// Subrange type: type TDigit = 0..9;
		return "type " + td.Name.String() + " = " + td.LowBound.String() + ".." + td.HighBound.String()
	}

	if td.IsFunctionPointer {
		// Function pointer type: type TFunc = function(x: Integer): Boolean;
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
