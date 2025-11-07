package lexer

// This file provides type aliases for backwards compatibility
// while we transition to using pkg/token

import "github.com/cwbudde/go-dws/pkg/token"

// Type aliases to pkg/token types
type Position = token.Position
type Token = token.Token
type TokenType = token.TokenType

// Re-export all token type constants
const (
	ILLEGAL = token.ILLEGAL
	EOF     = token.EOF
	COMMENT = token.COMMENT

	IDENT  = token.IDENT
	INT    = token.INT
	FLOAT  = token.FLOAT
	STRING = token.STRING
	CHAR   = token.CHAR

	TRUE  = token.TRUE
	FALSE = token.FALSE
	NIL   = token.NIL

	BEGIN    = token.BEGIN
	END      = token.END
	IF       = token.IF
	THEN     = token.THEN
	ELSE     = token.ELSE
	CASE     = token.CASE
	OF       = token.OF
	WHILE    = token.WHILE
	REPEAT   = token.REPEAT
	UNTIL    = token.UNTIL
	FOR      = token.FOR
	TO       = token.TO
	DOWNTO   = token.DOWNTO
	STEP     = token.STEP
	DO       = token.DO
	BREAK    = token.BREAK
	CONTINUE = token.CONTINUE
	EXIT     = token.EXIT
	WITH     = token.WITH
	ASM      = token.ASM

	VAR            = token.VAR
	CONST          = token.CONST
	TYPE           = token.TYPE
	RECORD         = token.RECORD
	ARRAY          = token.ARRAY
	SET            = token.SET
	ENUM           = token.ENUM
	FLAGS          = token.FLAGS
	RESOURCESTRING = token.RESOURCESTRING
	NAMESPACE      = token.NAMESPACE
	UNIT           = token.UNIT
	USES           = token.USES
	PROGRAM        = token.PROGRAM
	LIBRARY        = token.LIBRARY
	IMPLEMENTATION = token.IMPLEMENTATION
	INITIALIZATION = token.INITIALIZATION
	FINALIZATION   = token.FINALIZATION

	CLASS       = token.CLASS
	OBJECT      = token.OBJECT
	INTERFACE   = token.INTERFACE
	IMPLEMENTS  = token.IMPLEMENTS
	FUNCTION    = token.FUNCTION
	PROCEDURE   = token.PROCEDURE
	CONSTRUCTOR = token.CONSTRUCTOR
	DESTRUCTOR  = token.DESTRUCTOR
	METHOD      = token.METHOD
	PROPERTY    = token.PROPERTY
	VIRTUAL     = token.VIRTUAL
	OVERRIDE    = token.OVERRIDE
	ABSTRACT    = token.ABSTRACT
	SEALED      = token.SEALED
	STATIC      = token.STATIC
	FINAL       = token.FINAL
	NEW         = token.NEW
	INHERITED   = token.INHERITED
	REINTRODUCE = token.REINTRODUCE
	OPERATOR    = token.OPERATOR
	HELPER      = token.HELPER
	PARTIAL     = token.PARTIAL
	LAZY        = token.LAZY
	INDEX       = token.INDEX

	TRY     = token.TRY
	EXCEPT  = token.EXCEPT
	RAISE   = token.RAISE
	FINALLY = token.FINALLY
	ON      = token.ON

	NOT = token.NOT
	AND = token.AND
	OR  = token.OR
	XOR = token.XOR

	IS   = token.IS
	AS   = token.AS
	IN   = token.IN
	DIV  = token.DIV
	MOD  = token.MOD
	SHL  = token.SHL
	SHR  = token.SHR
	SAR  = token.SAR
	IMPL = token.IMPL

	INLINE     = token.INLINE
	EXTERNAL   = token.EXTERNAL
	FORWARD    = token.FORWARD
	OVERLOAD   = token.OVERLOAD
	DEPRECATED = token.DEPRECATED
	READONLY   = token.READONLY
	EXPORT     = token.EXPORT
	REGISTER   = token.REGISTER
	PASCAL     = token.PASCAL
	CDECL      = token.CDECL
	SAFECALL   = token.SAFECALL
	STDCALL    = token.STDCALL
	FASTCALL   = token.FASTCALL
	REFERENCE  = token.REFERENCE

	PRIVATE   = token.PRIVATE
	PROTECTED = token.PROTECTED
	PUBLIC    = token.PUBLIC
	PUBLISHED = token.PUBLISHED
	STRICT    = token.STRICT

	READ        = token.READ
	WRITE       = token.WRITE
	DEFAULT     = token.DEFAULT
	DESCRIPTION = token.DESCRIPTION

	OLD        = token.OLD
	REQUIRE    = token.REQUIRE
	ENSURE     = token.ENSURE
	INVARIANTS = token.INVARIANTS

	ASYNC    = token.ASYNC
	AWAIT    = token.AWAIT
	LAMBDA   = token.LAMBDA
	IMPLIES  = token.IMPLIES
	EMPTY    = token.EMPTY
	IMPLICIT = token.IMPLICIT
	EXPLICIT = token.EXPLICIT

	LPAREN    = token.LPAREN
	RPAREN    = token.RPAREN
	LBRACK    = token.LBRACK
	RBRACK    = token.RBRACK
	LBRACE    = token.LBRACE
	RBRACE    = token.RBRACE
	SEMICOLON = token.SEMICOLON
	COMMA     = token.COMMA
	DOT       = token.DOT
	COLON     = token.COLON
	DOTDOT    = token.DOTDOT

	PLUS     = token.PLUS
	MINUS    = token.MINUS
	ASTERISK = token.ASTERISK
	SLASH    = token.SLASH
	PERCENT  = token.PERCENT
	CARET    = token.CARET
	POWER    = token.POWER

	EQ         = token.EQ
	NOT_EQ     = token.NOT_EQ
	LESS       = token.LESS
	GREATER    = token.GREATER
	LESS_EQ    = token.LESS_EQ
	GREATER_EQ = token.GREATER_EQ
	EQ_EQ      = token.EQ_EQ
	EQ_EQ_EQ   = token.EQ_EQ_EQ
	EXCL_EQ    = token.EXCL_EQ

	ASSIGN         = token.ASSIGN
	PLUS_ASSIGN    = token.PLUS_ASSIGN
	MINUS_ASSIGN   = token.MINUS_ASSIGN
	TIMES_ASSIGN   = token.TIMES_ASSIGN
	DIVIDE_ASSIGN  = token.DIVIDE_ASSIGN
	PERCENT_ASSIGN = token.PERCENT_ASSIGN
	CARET_ASSIGN   = token.CARET_ASSIGN
	AT_ASSIGN      = token.AT_ASSIGN
	TILDE_ASSIGN   = token.TILDE_ASSIGN

	INC = token.INC
	DEC = token.DEC

	LESS_LESS       = token.LESS_LESS
	GREATER_GREATER = token.GREATER_GREATER
	PIPE            = token.PIPE
	PIPE_PIPE       = token.PIPE_PIPE
	AMP             = token.AMP
	AMP_AMP         = token.AMP_AMP

	AT                = token.AT
	TILDE             = token.TILDE
	BACKSLASH         = token.BACKSLASH
	DOLLAR            = token.DOLLAR
	EXCLAMATION       = token.EXCLAMATION
	QUESTION          = token.QUESTION
	QUESTION_QUESTION = token.QUESTION_QUESTION
	QUESTION_DOT      = token.QUESTION_DOT
	FAT_ARROW         = token.FAT_ARROW

	SWITCH = token.SWITCH
)

// Re-export token functions
var (
	NewToken          = token.NewToken
	LookupIdent       = token.LookupIdent
	IsKeyword         = token.IsKeyword
	GetKeywordLiteral = token.GetKeywordLiteral
)
