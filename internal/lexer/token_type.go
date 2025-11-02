package lexer

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
