package ast

import "github.com/cwbudde/go-dws/internal/lexer"

// TypeAnnotation represents a type annotation in the AST.
// This is used for variable declarations, parameters, and return types.
// Example: `: Integer` in `var x: Integer := 5;`
type TypeAnnotation struct {
	Token lexer.Token // The ':' token or type name token
	Name  string      // The type name (e.g., "Integer", "String")
}

// String returns the string representation of the type annotation
func (ta *TypeAnnotation) String() string {
	if ta == nil {
		return ""
	}
	return ta.Name
}

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
// Type Declarations (Task 9.15)
// ============================================================================

// TypeDeclaration represents a type declaration statement in DWScript.
// This can be either a type alias or a full type definition.
//
// Type alias examples:
//
//	type TUserID = Integer;
//	type TFileName = String;
//	type TIntArray = array of Integer;
//
// Full type definitions (enums, records, classes) will use specialized
// declaration nodes (EnumDecl, RecordDecl, ClassDecl) but may eventually
// be unified under this node.
//
// Task 9.15: This node currently supports type aliases. Future tasks will
// extend it to handle all type declarations.
type TypeDeclaration struct {
	Token       lexer.Token     // The 'type' token
	Name        *Identifier     // The type name (e.g., "TUserID", "TFileName")
	IsAlias     bool            // True if this is a type alias (type A = B)
	AliasedType *TypeAnnotation // For aliases: the type being aliased (nil if not an alias)
	// Future: Add fields for full type definitions (record, class, enum, etc.)
}

func (td *TypeDeclaration) statementNode()       {}
func (td *TypeDeclaration) TokenLiteral() string { return td.Token.Literal }
func (td *TypeDeclaration) Pos() lexer.Position  { return td.Token.Pos }

// String returns the string representation of the type declaration.
// For type aliases, this returns: "type Name = Type;"
// For full type definitions, this will be extended in future tasks.
func (td *TypeDeclaration) String() string {
	if td.IsAlias {
		// Type alias: type TUserID = Integer;
		return "type " + td.Name.String() + " = " + td.AliasedType.String()
	}

	// For now, only aliases are supported
	// Future: Handle full type definitions
	return "type " + td.Name.String()
}
