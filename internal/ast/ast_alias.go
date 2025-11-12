package ast

// This file provides type aliases to pkg/ast for backwards compatibility.
// All internal code continues to use internal/ast, but now it maps to the
// public pkg/ast types.

import (
	pkgast "github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// Core interfaces
type (
	Node       = pkgast.Node
	Expression = pkgast.Expression
	Statement  = pkgast.Statement
)

// Type system interfaces
type (
	TypeExpression  = pkgast.TypeExpression
	TypedExpression = pkgast.TypedExpression
)

// Program and basic nodes
type (
	Program    = pkgast.Program
	Identifier = pkgast.Identifier
)

// Literals
type (
	IntegerLiteral = pkgast.IntegerLiteral
	FloatLiteral   = pkgast.FloatLiteral
	StringLiteral  = pkgast.StringLiteral
	BooleanLiteral = pkgast.BooleanLiteral
	CharLiteral    = pkgast.CharLiteral
	NilLiteral     = pkgast.NilLiteral
)

// Expressions
type (
	BinaryExpression        = pkgast.BinaryExpression
	UnaryExpression         = pkgast.UnaryExpression
	GroupedExpression       = pkgast.GroupedExpression
	RangeExpression         = pkgast.RangeExpression
	CallExpression          = pkgast.CallExpression
	IndexExpression         = pkgast.IndexExpression
	ArrayLiteralExpression  = pkgast.ArrayLiteralExpression
	NewArrayExpression      = pkgast.NewArrayExpression
	SetLiteral              = pkgast.SetLiteral
	EnumLiteral             = pkgast.EnumLiteral
	NewExpression           = pkgast.NewExpression
	MemberAccessExpression  = pkgast.MemberAccessExpression
	MethodCallExpression    = pkgast.MethodCallExpression
	InheritedExpression     = pkgast.InheritedExpression
	RecordLiteralExpression = pkgast.RecordLiteralExpression
	LambdaExpression        = pkgast.LambdaExpression
	AddressOfExpression     = pkgast.AddressOfExpression
	OldExpression           = pkgast.OldExpression
	IsExpression            = pkgast.IsExpression
	AsExpression            = pkgast.AsExpression
	ImplementsExpression    = pkgast.ImplementsExpression
)

// Statements
type (
	ExpressionStatement = pkgast.ExpressionStatement
	BlockStatement      = pkgast.BlockStatement
	VarDeclStatement    = pkgast.VarDeclStatement
	AssignmentStatement = pkgast.AssignmentStatement
	ConstDecl           = pkgast.ConstDecl
	ReturnStatement     = pkgast.ReturnStatement
)

// Control flow
type (
	IfStatement       = pkgast.IfStatement
	IfExpression      = pkgast.IfExpression
	WhileStatement    = pkgast.WhileStatement
	RepeatStatement   = pkgast.RepeatStatement
	ForStatement      = pkgast.ForStatement
	ForInStatement    = pkgast.ForInStatement
	CaseStatement     = pkgast.CaseStatement
	CaseBranch        = pkgast.CaseBranch
	BreakStatement    = pkgast.BreakStatement
	ContinueStatement = pkgast.ContinueStatement
	ExitStatement     = pkgast.ExitStatement
	ForDirection      = pkgast.ForDirection
)

// Exception handling
type (
	TryStatement     = pkgast.TryStatement
	ExceptClause     = pkgast.ExceptClause
	ExceptionHandler = pkgast.ExceptionHandler
	FinallyClause    = pkgast.FinallyClause
	RaiseStatement   = pkgast.RaiseStatement
)

// Declarations
type (
	FunctionDecl        = pkgast.FunctionDecl
	Parameter           = pkgast.Parameter
	ClassDecl           = pkgast.ClassDecl
	FieldDecl           = pkgast.FieldDecl
	RecordDecl          = pkgast.RecordDecl
	EnumDecl            = pkgast.EnumDecl
	EnumValue           = pkgast.EnumValue
	ArrayDecl           = pkgast.ArrayDecl
	SetDecl             = pkgast.SetDecl
	InterfaceDecl       = pkgast.InterfaceDecl
	InterfaceMethodDecl = pkgast.InterfaceMethodDecl
	OperatorDecl        = pkgast.OperatorDecl
	PropertyDecl        = pkgast.PropertyDecl
	HelperDecl          = pkgast.HelperDecl
	UnitDeclaration     = pkgast.UnitDeclaration
	UsesClause          = pkgast.UsesClause
	RecordPropertyDecl  = pkgast.RecordPropertyDecl
	FieldInitializer    = pkgast.FieldInitializer
)

// Type nodes
type (
	TypeAnnotation          = pkgast.TypeAnnotation
	ArrayTypeAnnotation     = pkgast.ArrayTypeAnnotation
	ArrayTypeNode           = pkgast.ArrayTypeNode
	SetTypeNode             = pkgast.SetTypeNode
	ClassOfTypeNode         = pkgast.ClassOfTypeNode
	FunctionPointerTypeNode = pkgast.FunctionPointerTypeNode
	TypeDeclaration         = pkgast.TypeDeclaration
)

// Supporting types
type (
	Condition      = pkgast.Condition
	PreConditions  = pkgast.PreConditions
	PostConditions = pkgast.PostConditions
)

// Enums
const (
	VisibilityPrivate   = pkgast.VisibilityPrivate
	VisibilityProtected = pkgast.VisibilityProtected
	VisibilityPublic    = pkgast.VisibilityPublic

	ForTo     = pkgast.ForTo
	ForDownto = pkgast.ForDownto

	OperatorKindGlobal     = pkgast.OperatorKindGlobal
	OperatorKindClass      = pkgast.OperatorKindClass
	OperatorKindConversion = pkgast.OperatorKindConversion
)

type (
	Visibility   = pkgast.Visibility
	OperatorKind = pkgast.OperatorKind
)

// Helper to maintain compatibility with existing code
// that might construct Position directly
type Position = token.Position
