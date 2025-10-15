package lexer

import (
	"fmt"
	"strings"
)

// Position represents a location in the source code.
type Position struct {
	Line   int // Line number (1-indexed)
	Column int // Column number (1-indexed)
	Offset int // Byte offset (0-indexed)
}

// Token represents a lexical token with its type, literal value, and position.
type Token struct {
	Type    TokenType // The type of the token
	Literal string    // The literal value of the token as it appears in source
	Pos     Position  // Position in the source code
}

// String returns a string representation of the token for debugging.
func (t Token) String() string {
	if t.Type == EOF {
		return fmt.Sprintf("%s at %d:%d", t.Type, t.Pos.Line, t.Pos.Column)
	}
	if len(t.Literal) > 20 {
		return fmt.Sprintf("%s(%q...) at %d:%d", t.Type, t.Literal[:20], t.Pos.Line, t.Pos.Column)
	}
	return fmt.Sprintf("%s(%q) at %d:%d", t.Type, t.Literal, t.Pos.Line, t.Pos.Column)
}

// NewToken creates a new token with the given type, literal, and position.
func NewToken(tokenType TokenType, literal string, pos Position) Token {
	return Token{
		Type:    tokenType,
		Literal: literal,
		Pos:     pos,
	}
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
	"is":    IS,
	"as":    AS,
	"in":    IN,
	"div":   DIV,
	"mod":   MOD,
	"shl":   SHL,
	"shr":   SHR,
	"sar":   SAR,
	"impl":  IMPL,

	// Function modifiers
	"inline":     INLINE,
	"external":   EXTERNAL,
	"forward":    FORWARD,
	"overload":   OVERLOAD,
	"deprecated": DEPRECATED,
	"readonly":   READONLY,
	"export":     EXPORT,
	"register":   REGISTER,
	"pascal":     PASCAL,
	"cdecl":      CDECL,
	"safecall":   SAFECALL,
	"stdcall":    STDCALL,
	"fastcall":   FASTCALL,
	"reference":  REFERENCE,

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
	"async":   ASYNC,
	"await":   AWAIT,
	"lambda":  LAMBDA,
	"implies": IMPLIES,
	"empty":   EMPTY,
	"implicit": IMPLICIT,
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
