package token

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Position represents a location in the source code.
// Line and Column are 1-indexed for human readability.
// Offset is 0-indexed for programmatic access.
//
// # Column Position and Unicode
//
// The Column field represents the count of Unicode code points (runes) from the
// start of the line, NOT the byte offset or visual display width.
//
// For example:
//   - ASCII text: "var x" → 'x' is at column 5
//   - Unicode text: "var Δ" → 'Δ' is at column 5 (Δ is one rune, even though it's multi-byte)
//   - Emoji: "// 🚀" → '🚀' is at column 4 (even though it may display as 2 cells wide)
//
// This means:
//   - Multi-byte UTF-8 characters count as 1 column
//   - Display width (terminal cells) may differ from column number
//   - Error markers may not align visually for wide characters
//   - But positions are consistent and reproducible across all systems
type Position struct {
	Line   int // Line number (1-indexed)
	Column int // Column number (1-indexed, rune count not display width or byte offset)
	Offset int // Byte offset (0-indexed)
}

// String returns a string representation of the position in the format "line:column".
func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// IsValid reports whether the position is valid.
// A valid position has Line > 0.
func (p Position) IsValid() bool {
	return p.Line > 0
}

// Token represents a lexical token with its type, literal value, and position.
// Every piece of DWScript source code is represented as a sequence of tokens.
type Token struct {
	Literal string
	Pos     Position
	Type    TokenType
}

// String returns a string representation of the token for debugging.
// The format is: TYPE("literal") at line:column
func (t Token) String() string {
	if t.Type == EOF {
		return fmt.Sprintf("%s at %d:%d", t.Type, t.Pos.Line, t.Pos.Column)
	}
	if len(t.Literal) > 20 {
		return fmt.Sprintf("%s(%q...) at %d:%d", t.Type, t.Literal[:20], t.Pos.Line, t.Pos.Column)
	}
	return fmt.Sprintf("%s(%q) at %d:%d", t.Type, t.Literal, t.Pos.Line, t.Pos.Column)
}

// Length returns the length of the token in characters (runes).
// This is useful for error reporting and LSP integration, allowing tools
// to highlight the exact span of code represented by this token.
// For byte length, use len(t.Literal).
func (t Token) Length() int {
	return utf8.RuneCountInString(t.Literal)
}

// End returns the position immediately after this token.
// Column is calculated using rune count to match the lexer's rune-based column tracking.
// Offset uses byte length for correct byte position in the source.
func (t Token) End() Position {
	return Position{
		Line:   t.Pos.Line,
		Column: t.Pos.Column + utf8.RuneCountInString(t.Literal),
		Offset: t.Pos.Offset + len(t.Literal),
	}
}

// NewToken creates a new token with the given type, literal, and position.
func NewToken(tokenType TokenType, literal string, pos Position) Token {
	return Token{
		Type:    tokenType,
		Literal: literal,
		Pos:     pos,
	}
}

// TokenType represents the type of a token in DWScript source code.
// The token types are organized into logical groups for clarity.
type TokenType int

// Token type constants organized by category
const (
	// Special tokens
	ILLEGAL TokenType = iota // Unexpected character
	EOF                      // End of file
	COMMENT                  // Comment (line or block)

	// Identifiers and literals
	IDENT  // identifiers: x, myVar, MyClass
	INT    // integer literals: 123, $FF, %1010, 0xFF
	FLOAT  // float literals: 123.45, 1.5e10
	STRING // string literals: 'hello', "world", '''multiline'''
	CHAR   // character literals: #65, #$41

	literalEnd // marker for end of literals section

	// Keywords - Boolean literals (also keywords)
	TRUE  // true boolean literal
	FALSE // false boolean literal
	NIL   // nil literal

	// Keywords - Variant special values
	NULL       // Null - special variant value
	UNASSIGNED // Unassigned - default uninitialized variant value

	// Keywords - Control flow
	BEGIN    // begin
	END      // end
	IF       // if
	THEN     // then
	ELSE     // else
	CASE     // case
	OF       // of
	WHILE    // while
	REPEAT   // repeat
	UNTIL    // until
	FOR      // for
	TO       // to
	DOWNTO   // downto
	STEP     // step
	DO       // do
	BREAK    // break
	CONTINUE // continue
	EXIT     // exit
	WITH     // with
	ASM      // asm

	// Keywords - Declaration
	VAR            // var
	CONST          // const
	TYPE           // type
	RECORD         // record
	ARRAY          // array
	SET            // set
	ENUM           // enum
	FLAGS          // flags
	RESOURCESTRING // resourcestring
	NAMESPACE      // namespace
	UNIT           // unit
	USES           // uses
	PROGRAM        // program
	LIBRARY        // library
	IMPLEMENTATION // implementation
	INITIALIZATION // initialization
	FINALIZATION   // finalization

	// Keywords - Object-oriented
	CLASS       // class
	OBJECT      // object
	INTERFACE   // interface
	IMPLEMENTS  // implements
	FUNCTION    // function
	PROCEDURE   // procedure
	CONSTRUCTOR // constructor
	DESTRUCTOR  // destructor
	METHOD      // method
	PROPERTY    // property
	VIRTUAL     // virtual
	OVERRIDE    // override
	ABSTRACT    // abstract
	SEALED      // sealed
	STATIC      // static
	FINAL       // final
	NEW         // new
	INHERITED   // inherited
	SELF        // self
	REINTRODUCE // reintroduce
	OPERATOR    // operator
	HELPER      // helper
	PARTIAL     // partial
	LAZY        // lazy
	INDEX       // index

	// Keywords - Exception handling
	TRY     // try
	EXCEPT  // except
	RAISE   // raise
	FINALLY // finally
	ON      // on

	// Keywords - Boolean/Logical
	NOT // not
	AND // and
	OR  // or
	XOR // xor

	// Keywords - Special
	IS   // is (type checking)
	AS   // as (type casting)
	IN   // in (set membership)
	DIV  // div (integer division)
	MOD  // mod (modulo)
	SHL  // shl (shift left)
	SHR  // shr (shift right)
	SAR  // sar (arithmetic shift right)
	IMPL // impl (short for implementation)

	// Keywords - Function modifiers
	INLINE     // inline
	EXTERNAL   // external
	FORWARD    // forward
	OVERLOAD   // overload
	DEPRECATED // deprecated
	READONLY   // readonly
	EXPORT     // export
	REGISTER   // register (calling convention)
	PASCAL     // pascal (calling convention)
	CDECL      // cdecl (calling convention)
	SAFECALL   // safecall (calling convention)
	STDCALL    // stdcall (calling convention)
	FASTCALL   // fastcall (calling convention)
	REFERENCE  // reference (calling convention)

	// Keywords - Access modifiers
	PRIVATE   // private
	PROTECTED // protected
	PUBLIC    // public
	PUBLISHED // published
	STRICT    // strict

	// Keywords - Property access
	READ        // read
	WRITE       // write
	DEFAULT     // default
	DESCRIPTION // description

	// Keywords - Contracts (Design by Contract)
	OLD        // old (contracts)
	REQUIRE    // require (precondition)
	ENSURE     // ensure (postcondition)
	INVARIANTS // invariants

	// Keywords - Modern features
	ASYNC    // async
	AWAIT    // await
	LAMBDA   // lambda
	IMPLIES  // implies
	EMPTY    // empty
	IMPLICIT // implicit
	EXPLICIT // explicit

	keywordEnd // marker for end of keywords section

	// Delimiters
	LPAREN    // (
	RPAREN    // )
	LBRACK    // [
	RBRACK    // ]
	LBRACE    // {
	RBRACE    // }
	SEMICOLON // ;
	COMMA     // ,
	DOT       // .
	COLON     // :
	DOTDOT    // .. (range operator)

	// Arithmetic operators
	PLUS     // +
	MINUS    // -
	ASTERISK // *
	SLASH    // /
	PERCENT  // %
	CARET    // ^ (power/pointer dereference)
	POWER    // ** (alternative power operator)

	// Comparison operators
	EQ         // =
	NOT_EQ     // <>
	LESS       // <
	GREATER    // >
	LESS_EQ    // <=
	GREATER_EQ // >=
	EQ_EQ      // == (JavaScript-style equality)
	EQ_EQ_EQ   // === (JavaScript-style strict equality)
	EXCL_EQ    // != (C-style not equal)

	// Assignment operators
	ASSIGN         // :=
	PLUS_ASSIGN    // +=
	MINUS_ASSIGN   // -=
	TIMES_ASSIGN   // *=
	DIVIDE_ASSIGN  // /=
	PERCENT_ASSIGN // %=
	CARET_ASSIGN   // ^=
	AT_ASSIGN      // @=
	TILDE_ASSIGN   // ~=

	// Increment/Decrement
	INC // ++ (increment)
	DEC // -- (decrement)

	// Bitwise/Boolean operators
	LESS_LESS       // << (left shift)
	GREATER_GREATER // >> (right shift)
	PIPE            // |
	PIPE_PIPE       // ||
	AMP             // &
	AMP_AMP         // &&

	// Special operators
	AT                // @ (address of)
	TILDE             // ~
	BACKSLASH         // \
	DOLLAR            // $
	EXCLAMATION       // !
	QUESTION          // ?
	QUESTION_QUESTION // ?? (null coalescing)
	QUESTION_DOT      // ?. (safe navigation)
	FAT_ARROW         // => (lambda arrow)

	// Compiler directives
	SWITCH // {$directive} compiler switch
)

// String returns the string representation of a TokenType.
func (tt TokenType) String() string {
	if int(tt) < len(tokenTypeStrings) {
		return tokenTypeStrings[tt]
	}
	return "UNKNOWN"
}

// IsLiteral returns true if the token type is a literal value.
func (tt TokenType) IsLiteral() bool {
	return tt > EOF && tt < literalEnd
}

// IsKeyword returns true if the token type is a keyword.
func (tt TokenType) IsKeyword() bool {
	return tt > literalEnd && tt < keywordEnd
}

// IsOperator returns true if the token type is an operator.
func (tt TokenType) IsOperator() bool {
	return tt >= PLUS && tt <= FAT_ARROW
}

// IsDelimiter returns true if the token type is a delimiter.
func (tt TokenType) IsDelimiter() bool {
	return tt >= LPAREN && tt <= DOTDOT
}

// tokenTypeStrings maps TokenType values to their string representations.
var tokenTypeStrings = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	COMMENT: "COMMENT",

	// Identifiers and literals
	IDENT:  "IDENT",
	INT:    "INT",
	FLOAT:  "FLOAT",
	STRING: "STRING",
	CHAR:   "CHAR",
	TRUE:   "TRUE",
	FALSE:  "FALSE",
	NIL:    "NIL",

	// Variant special values
	NULL:       "NULL",
	UNASSIGNED: "UNASSIGNED",

	// Keywords - Control flow
	BEGIN:    "BEGIN",
	END:      "END",
	IF:       "IF",
	THEN:     "THEN",
	ELSE:     "ELSE",
	CASE:     "CASE",
	OF:       "OF",
	WHILE:    "WHILE",
	REPEAT:   "REPEAT",
	UNTIL:    "UNTIL",
	FOR:      "FOR",
	TO:       "TO",
	DOWNTO:   "DOWNTO",
	STEP:     "STEP",
	DO:       "DO",
	BREAK:    "BREAK",
	CONTINUE: "CONTINUE",
	EXIT:     "EXIT",
	WITH:     "WITH",
	ASM:      "ASM",

	// Keywords - Declaration
	VAR:            "VAR",
	CONST:          "CONST",
	TYPE:           "TYPE",
	RECORD:         "RECORD",
	ARRAY:          "ARRAY",
	SET:            "SET",
	ENUM:           "ENUM",
	FLAGS:          "FLAGS",
	RESOURCESTRING: "RESOURCESTRING",
	NAMESPACE:      "NAMESPACE",
	UNIT:           "UNIT",
	USES:           "USES",
	PROGRAM:        "PROGRAM",
	LIBRARY:        "LIBRARY",
	IMPLEMENTATION: "IMPLEMENTATION",
	INITIALIZATION: "INITIALIZATION",
	FINALIZATION:   "FINALIZATION",

	// Keywords - Object-oriented
	CLASS:       "CLASS",
	OBJECT:      "OBJECT",
	INTERFACE:   "INTERFACE",
	IMPLEMENTS:  "IMPLEMENTS",
	FUNCTION:    "FUNCTION",
	PROCEDURE:   "PROCEDURE",
	CONSTRUCTOR: "CONSTRUCTOR",
	DESTRUCTOR:  "DESTRUCTOR",
	METHOD:      "METHOD",
	PROPERTY:    "PROPERTY",
	VIRTUAL:     "VIRTUAL",
	OVERRIDE:    "OVERRIDE",
	ABSTRACT:    "ABSTRACT",
	SEALED:      "SEALED",
	STATIC:      "STATIC",
	FINAL:       "FINAL",
	NEW:         "NEW",
	INHERITED:   "INHERITED",
	SELF:        "SELF",
	REINTRODUCE: "REINTRODUCE",
	OPERATOR:    "OPERATOR",
	HELPER:      "HELPER",
	PARTIAL:     "PARTIAL",
	LAZY:        "LAZY",
	INDEX:       "INDEX",

	// Keywords - Exception handling
	TRY:     "TRY",
	EXCEPT:  "EXCEPT",
	RAISE:   "RAISE",
	FINALLY: "FINALLY",
	ON:      "ON",

	// Keywords - Boolean/Logical
	NOT: "NOT",
	AND: "AND",
	OR:  "OR",
	XOR: "XOR",

	// Keywords - Special
	IS:   "IS",
	AS:   "AS",
	IN:   "IN",
	DIV:  "DIV",
	MOD:  "MOD",
	SHL:  "SHL",
	SHR:  "SHR",
	SAR:  "SAR",
	IMPL: "IMPL",

	// Keywords - Function modifiers
	INLINE:     "INLINE",
	EXTERNAL:   "EXTERNAL",
	FORWARD:    "FORWARD",
	OVERLOAD:   "OVERLOAD",
	DEPRECATED: "DEPRECATED",
	READONLY:   "READONLY",
	EXPORT:     "EXPORT",
	REGISTER:   "REGISTER",
	PASCAL:     "PASCAL",
	CDECL:      "CDECL",
	SAFECALL:   "SAFECALL",
	STDCALL:    "STDCALL",
	FASTCALL:   "FASTCALL",
	REFERENCE:  "REFERENCE",

	// Keywords - Access modifiers
	PRIVATE:   "PRIVATE",
	PROTECTED: "PROTECTED",
	PUBLIC:    "PUBLIC",
	PUBLISHED: "PUBLISHED",
	STRICT:    "STRICT",

	// Keywords - Property access
	READ:        "READ",
	WRITE:       "WRITE",
	DEFAULT:     "DEFAULT",
	DESCRIPTION: "DESCRIPTION",

	// Keywords - Contracts
	OLD:        "OLD",
	REQUIRE:    "REQUIRE",
	ENSURE:     "ENSURE",
	INVARIANTS: "INVARIANTS",

	// Keywords - Modern features
	ASYNC:    "ASYNC",
	AWAIT:    "AWAIT",
	LAMBDA:   "LAMBDA",
	IMPLIES:  "IMPLIES",
	EMPTY:    "EMPTY",
	IMPLICIT: "IMPLICIT",
	EXPLICIT: "EXPLICIT",

	// Delimiters
	LPAREN:    "LPAREN",
	RPAREN:    "RPAREN",
	LBRACK:    "LBRACK",
	RBRACK:    "RBRACK",
	LBRACE:    "LBRACE",
	RBRACE:    "RBRACE",
	SEMICOLON: "SEMICOLON",
	COMMA:     "COMMA",
	DOT:       "DOT",
	COLON:     "COLON",
	DOTDOT:    "DOTDOT",

	// Arithmetic operators
	PLUS:     "PLUS",
	MINUS:    "MINUS",
	ASTERISK: "ASTERISK",
	SLASH:    "SLASH",
	PERCENT:  "PERCENT",
	CARET:    "CARET",
	POWER:    "POWER",

	// Comparison operators
	EQ:         "EQ",
	NOT_EQ:     "NOT_EQ",
	LESS:       "LESS",
	GREATER:    "GREATER",
	LESS_EQ:    "LESS_EQ",
	GREATER_EQ: "GREATER_EQ",
	EQ_EQ:      "EQ_EQ",
	EQ_EQ_EQ:   "EQ_EQ_EQ",
	EXCL_EQ:    "EXCL_EQ",

	// Assignment operators
	ASSIGN:         "ASSIGN",
	PLUS_ASSIGN:    "PLUS_ASSIGN",
	MINUS_ASSIGN:   "MINUS_ASSIGN",
	TIMES_ASSIGN:   "TIMES_ASSIGN",
	DIVIDE_ASSIGN:  "DIVIDE_ASSIGN",
	PERCENT_ASSIGN: "PERCENT_ASSIGN",
	CARET_ASSIGN:   "CARET_ASSIGN",
	AT_ASSIGN:      "AT_ASSIGN",
	TILDE_ASSIGN:   "TILDE_ASSIGN",

	// Increment/Decrement
	INC: "INC",
	DEC: "DEC",

	// Bitwise/Boolean operators
	LESS_LESS:       "LESS_LESS",
	GREATER_GREATER: "GREATER_GREATER",
	PIPE:            "PIPE",
	PIPE_PIPE:       "PIPE_PIPE",
	AMP:             "AMP",
	AMP_AMP:         "AMP_AMP",

	// Special operators
	AT:                "AT",
	TILDE:             "TILDE",
	BACKSLASH:         "BACKSLASH",
	DOLLAR:            "DOLLAR",
	EXCLAMATION:       "EXCLAMATION",
	QUESTION:          "QUESTION",
	QUESTION_QUESTION: "QUESTION_QUESTION",
	QUESTION_DOT:      "QUESTION_DOT",
	FAT_ARROW:         "FAT_ARROW",

	// Compiler directives
	SWITCH: "SWITCH",
}

// keywords maps DWScript keyword strings to their TokenType.
// DWScript keywords are case-insensitive, so we store them in lowercase.
var keywords = map[string]TokenType{
	// Control flow
	"begin":    BEGIN,
	"end":      END,
	"if":       IF,
	"then":     THEN,
	"else":     ELSE,
	"case":     CASE,
	"of":       OF,
	"while":    WHILE,
	"repeat":   REPEAT,
	"until":    UNTIL,
	"for":      FOR,
	"to":       TO,
	"downto":   DOWNTO,
	"step":     STEP,
	"do":       DO,
	"break":    BREAK,
	"continue": CONTINUE,
	"exit":     EXIT,
	"with":     WITH,
	"asm":      ASM,

	// Declaration
	"var":            VAR,
	"const":          CONST,
	"type":           TYPE,
	"record":         RECORD,
	"array":          ARRAY,
	"set":            SET,
	"enum":           ENUM,
	"flags":          FLAGS,
	"resourcestring": RESOURCESTRING,
	"namespace":      NAMESPACE,
	"unit":           UNIT,
	"uses":           USES,
	"program":        PROGRAM,
	"library":        LIBRARY,
	"implementation": IMPLEMENTATION,
	"initialization": INITIALIZATION,
	"finalization":   FINALIZATION,

	// Object-oriented
	"class":       CLASS,
	"object":      OBJECT,
	"interface":   INTERFACE,
	"implements":  IMPLEMENTS,
	"function":    FUNCTION,
	"procedure":   PROCEDURE,
	"constructor": CONSTRUCTOR,
	"destructor":  DESTRUCTOR,
	"method":      METHOD,
	"property":    PROPERTY,
	"virtual":     VIRTUAL,
	"override":    OVERRIDE,
	"abstract":    ABSTRACT,
	"sealed":      SEALED,
	"static":      STATIC,
	"final":       FINAL,
	"new":         NEW,
	"inherited":   INHERITED,
	"self":        SELF,
	"reintroduce": REINTRODUCE,
	"operator":    OPERATOR,
	"helper":      HELPER,
	"partial":     PARTIAL,
	"lazy":        LAZY,
	"index":       INDEX,

	// Exception handling
	"try":     TRY,
	"except":  EXCEPT,
	"raise":   RAISE,
	"finally": FINALLY,
	"on":      ON,

	// Boolean/Logical operators (keywords)
	"not": NOT,
	"and": AND,
	"or":  OR,
	"xor": XOR,

	// Special keywords
	"true":  TRUE,
	"false": FALSE,
	"nil":   NIL,

	// Variant special values
	"null":       NULL,
	"unassigned": UNASSIGNED,

	"is":   IS,
	"as":   AS,
	"in":   IN,
	"div":  DIV,
	"mod":  MOD,
	"shl":  SHL,
	"shr":  SHR,
	"sar":  SAR,
	"impl": IMPL,

	// Function modifiers
	"inline":     INLINE,
	"external":   EXTERNAL,
	"forward":    FORWARD,
	"overload":   OVERLOAD,
	"deprecated": DEPRECATED,
	"readonly":   READONLY,
	"export":     EXPORT,

	// Note: Calling convention keywords (register, pascal, cdecl, safecall, stdcall,
	// fastcall, reference) are NOT reserved keywords in DWScript. They are contextual
	// tokens only recognized by the parser in function/procedure declarations.
	// Therefore, they are NOT included in this keywords map and are treated as
	// regular identifiers by the lexer.

	// Access modifiers
	"private":   PRIVATE,
	"protected": PROTECTED,
	"public":    PUBLIC,
	"published": PUBLISHED,
	"strict":    STRICT,

	// Property access
	"read":        READ,
	"write":       WRITE,
	"default":     DEFAULT,
	"description": DESCRIPTION,

	// Contracts (Design by Contract)
	"old":        OLD,
	"require":    REQUIRE,
	"ensure":     ENSURE,
	"invariants": INVARIANTS,

	// Modern features
	"async":    ASYNC,
	"await":    AWAIT,
	"lambda":   LAMBDA,
	"implies":  IMPLIES,
	"empty":    EMPTY,
	"implicit": IMPLICIT,
	"explicit": EXPLICIT,
}

// LookupIdent returns the TokenType for a given identifier.
// If the identifier is a keyword, it returns the keyword's TokenType.
// Otherwise, it returns IDENT. Keywords are case-insensitive in DWScript.
func LookupIdent(ident string) TokenType {
	// Convert to lowercase for case-insensitive keyword lookup
	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}
	return IDENT
}

// IsKeyword returns true if the given string is a DWScript keyword.
// The comparison is case-insensitive.
func IsKeyword(ident string) bool {
	_, ok := keywords[strings.ToLower(ident)]
	return ok
}

// GetKeywordLiteral returns the canonical (lowercase) form of a keyword.
// If the string is not a keyword, it returns the original string.
func GetKeywordLiteral(ident string) string {
	lower := strings.ToLower(ident)
	if _, ok := keywords[lower]; ok {
		return lower
	}
	return ident
}
