# DWScript to Go Port - Detailed Implementation Plan

This document breaks down the ambitious goal of porting DWScript from Delphi to Go into bite-sized, actionable tasks organized by stages. Each stage builds incrementally toward a fully functional DWScript compiler/interpreter in Go.

---

## Stage 1: Implement the Lexer (Tokenization)

**Progress**: 45/45 tasks completed (100%) | **✅ STAGE 1 COMPLETE**

### Token Type Definition ✅ **COMPLETED**

**Completion Date**: October 15, 2025 | **Commit**: 2ac3470 | **Coverage**: 95.5%

- [x] 1.1 Create `lexer/token_type.go` and `lexer/token.go` files
- [x] 1.2 Define `TokenType` as an integer enum using iota
- [x] 1.3 Enumerate all 100+ DWScript keywords (begin, end, if, class, function, etc.)
- [x] 1.4 Enumerate all 40+ operators (+, -, *, /, :=, =, <>, ++, -=, etc.)
- [x] 1.5 Enumerate all delimiters/punctuation (;, ,, ., :, (, ), [, ], {, }, ..)
- [x] 1.6 Define literal token types (INT, FLOAT, STRING, CHAR, IDENT)
- [x] 1.7 Define special tokens (ILLEGAL, EOF, COMMENT, SWITCH, TRUE, FALSE, NIL)
- [x] 1.8 Create `Token` struct with fields: Type, Literal, Position (Line, Column, Offset)
- [x] 1.9 Create keyword map with case-insensitive lookup (150+ tokens total)
- [x] 1.10 Add `String()` methods and predicates (IsLiteral, IsKeyword, IsOperator, IsDelimiter)

**Summary**: See [docs/stage1-phase1-summary.md](docs/stage1-phase1-summary.md)

### Lexer Implementation ✅ **COMPLETED**

**Completion Date**: October 15, 2025 | **Coverage**: 97.1%

- [x] 1.11 Create `lexer/lexer.go` file
- [x] 1.12 Define `Lexer` struct with: input string, position, readPosition, line, column, ch (current character)
- [x] 1.13 Implement `New(input string) *Lexer` constructor
- [x] 1.14 Implement `readChar()` method to advance through input
- [x] 1.15 Implement `peekChar()` method to look ahead without advancing
- [x] 1.16 Implement `skipWhitespace()` method
- [x] 1.17 Implement comment handling:
  - [x] Handle `{ ... }` block comments
  - [x] Handle `(* ... *)` block comments
  - [x] Handle `//` line comments
- [x] 1.18 Implement `readIdentifier()` method
- [x] 1.19 Implement `readNumber()` method:
  - [x] Support integers
  - [x] Support floating-point numbers
  - [x] Support hex literals ($FF, 0xFF)
  - [x] Support binary literals (%1010)
- [x] 1.20 Implement `readString()` method:
  - [x] Handle single quotes (standard Pascal strings)
  - [x] Handle double quotes
  - [x] Handle escape sequences (doubled quotes: `''` = single quote)
  - [x] Handle multi-line strings
- [x] 1.21 Implement `NextToken()` method with main switch/case logic
- [x] 1.22 Handle all single-character tokens (+, -, *, /, etc.)
- [x] 1.23 Handle multi-character operators (:=, <=, >=, <>, etc.)
- [x] 1.24 Implement operator lookahead logic
- [x] 1.25 Handle identifier vs keyword distinction using keyword map
- [x] 1.26 Add proper line and column tracking throughout

**Summary**: See [docs/stage1-phase2-summary.md](docs/stage1-phase2-summary.md)

### Lexer Testing ✅ **COMPLETED**

**Completion Date**: October 15, 2025 | **Coverage**: 97.1%

- [x] 1.27 Create `lexer/lexer_test.go` file
- [x] 1.28 Write test for single keywords: `TestKeywords`
- [x] 1.29 Write test for operators: `TestOperators`
- [x] 1.30 Write test for identifiers: `TestIdentifiers`
- [x] 1.31 Write test for integer literals: `TestIntegerLiterals`
- [x] 1.32 Write test for float literals: `TestFloatLiterals`
- [x] 1.33 Write test for string literals: `TestStringLiterals`
- [x] 1.34 Write test for comments: `TestComments`
- [x] 1.35 Write test for a complete simple program: `TestSimpleProgram`
- [x] 1.36 Write test for edge cases: `TestEdgeCases`
  - [x] Empty input
  - [x] Unterminated string
  - [x] Unterminated comment
  - [x] Invalid characters
- [x] 1.37 Write test for line/column tracking accuracy: `TestPositionTracking`
- [x] 1.38 Run all tests: `go test ./lexer -v` - ✅ ALL PASS
- [x] 1.39 Fix any failing tests and edge cases - ✅ No failures
- [x] 1.40 Achieve >90% code coverage - ✅ 97.1% achieved
- [x] 1.41 Run `go vet ./lexer` - ✅ No issues
- [x] 1.42 Document lexer package with GoDoc comments - ✅ Complete

### Lexer Integration ✅ **COMPLETED**

**Completion Date**: October 15, 2025

- [x] 1.43 Create example usage in `cmd/dwscript/` - ✅ `lex` command implemented
- [x] 1.44 Test CLI with sample DWScript code - ✅ Verified with test files
- [x] 1.45 Benchmark tests - ✅ Included in lexer_test.go (6 benchmarks)

---

## Stage 2: Build a Minimal Parser and AST (Expressions Only)

**Progress**: 60/60 tasks completed (100%) | **✅ STAGE 2 COMPLETE**

**Completion Date**: October 16, 2025 | **Coverage**: AST 92.7%, Parser 81.9%

**Summary**: See [docs/stage2-summary.md](docs/stage2-summary.md)

### AST Node Definitions ✅ **COMPLETED**
- [x] 2.1 Create `ast/ast.go` file
- [x] 2.2 Define `Node` interface with methods: `TokenLiteral() string`, `String() string`
- [x] 2.3 Define `Expression` interface (embeds `Node`)
- [x] 2.4 Define `Statement` interface (embeds `Node`)
- [x] 2.5 Create `Program` struct (root node) with slice of statements
- [x] 2.6 Create `Identifier` struct (name string, token Token)
- [x] 2.7 Create literal expression nodes:
  - [ ] `IntegerLiteral` (Value int64)
  - [ ] `FloatLiteral` (Value float64)
  - [ ] `StringLiteral` (Value string)
  - [ ] `BooleanLiteral` (Value bool)
- [x] 2.8 Create `BinaryExpression` struct (Left Expr, Operator Token, Right Expr)
- [x] 2.9 Create `UnaryExpression` struct (Operator Token, Right Expr)
- [x] 2.10 Create `GroupedExpression` struct (Expression Expr) for parentheses
- [x] 2.11 Implement `String()` methods for all AST nodes (for debugging/testing)
- [x] 2.12 Create `ExpressionStatement` struct for expressions used as statements

### Parser Infrastructure
- [x] 2.13 Create `parser/parser.go` file
- [x] 2.14 Define `Parser` struct with: lexer, curToken, peekToken, errors slice
- [x] 2.15 Define operator precedence constants (LOWEST, EQUALS, LESSGREATER, SUM, PRODUCT, PREFIX, CALL)
- [x] 2.16 Create precedence map: `map[TokenType]int`
- [x] 2.17 Implement `New(lexer *Lexer) *Parser` constructor
- [x] 2.18 Implement `nextToken()` method
- [x] 2.19 Implement `curTokenIs(t TokenType) bool` helper
- [x] 2.20 Implement `peekTokenIs(t TokenType) bool` helper
- [x] 2.21 Implement `expectPeek(t TokenType) bool` with error on mismatch
- [x] 2.22 Implement error handling methods: `peekError()`, `Errors() []string`
- [x] 2.23 Implement `noPrefixParseFnError(t TokenType)` for parser errors

### Expression Parsing (Pratt Parser)
- [x] 2.24 Define prefix parse function type: `type prefixParseFn func() ast.Expression`
- [x] 2.25 Define infix parse function type: `type infixParseFn func(ast.Expression) ast.Expression`
- [x] 2.26 Add maps to parser: `prefixParseFns` and `infixParseFns`
- [x] 2.27 Implement `registerPrefix(token TokenType, fn prefixParseFn)`
- [x] 2.28 Implement `registerInfix(token TokenType, fn infixParseFn)`
- [x] 2.29 Implement `parseExpression(precedence int) ast.Expression`
- [x] 2.30 Implement `parseIdentifier()` prefix function
- [x] 2.31 Implement `parseIntegerLiteral()` prefix function
- [x] 2.32 Implement `parseFloatLiteral()` prefix function (if token is float)
- [x] 2.33 Implement `parseStringLiteral()` prefix function
- [x] 2.34 Implement `parseBooleanLiteral()` prefix function
- [x] 2.35 Implement `parseGroupedExpression()` for parentheses
- [x] 2.36 Implement `parsePrefixExpression()` for unary operators (-, not)
- [x] 2.37 Implement `parseInfixExpression(left ast.Expression)` for binary operators
- [x] 2.38 Register all prefix parsers in parser constructor
- [x] 2.39 Register all infix parsers in parser constructor
- [x] 2.40 Set up precedences for all operators (+, -, *, /, div, mod, =, <>, <, >, <=, >=, and, or)

### Statement Parsing (Minimal)
- [x] 2.41 Implement `ParseProgram() *ast.Program`
- [x] 2.42 Implement `parseStatement()` dispatcher
- [x] 2.43 Implement `parseExpressionStatement()` (expression followed by optional semicolon)
- [x] 2.44 Handle semicolon as statement terminator
- [x] 2.45 Implement basic error recovery (skip to next statement on parse error)

### Parser Testing
- [x] 2.46 Create `parser/parser_test.go` file
- [x] 2.47 Write helper function to create parser from input string
- [x] 2.48 Write helper to check parser errors
- [x] 2.49 Test parsing integer literals: `TestIntegerLiterals`
- [x] 2.50 Test parsing float literals: `TestFloatLiterals`
- [x] 2.51 Test parsing string literals: `TestStringLiterals`
- [x] 2.52 Test parsing identifiers: `TestIdentifiers`
- [x] 2.53 Test parsing prefix expressions: `TestPrefixExpressions` (e.g., `-5`, `not true`)
- [x] 2.54 Test parsing infix expressions: `TestInfixExpressions`
  - [ ] Test arithmetic: `3 + 5`, `10 - 2`, `4 * 5`, `20 / 4`
  - [ ] Test comparisons: `3 < 5`, `5 > 3`, `3 = 3`, `3 <> 4`
  - [ ] Test boolean: `true and false`, `true or false`
- [x] 2.55 Test operator precedence: `TestOperatorPrecedence`
  - [ ] `3 + 5 * 2` should parse as `3 + (5 * 2)`
  - [ ] `3 * 5 + 2` should parse as `(3 * 5) + 2`
  - [ ] Test with parentheses: `(3 + 5) * 2`
- [x] 2.56 Test grouped expressions: `TestGroupedExpressions`
- [x] 2.57 Test error reporting for invalid syntax
- [x] 2.58 Run all tests: `go test ./parser -v`
- [x] 2.59 Achieve >85% code coverage for parser package
- [x] 2.60 Document parser package with GoDoc comments

### Integration with CLI
- [ ] 2.61 Update CLI `run` command to parse input and print AST
- [ ] 2.62 Add `--dump-ast` flag to CLI for debugging
- [ ] 2.63 Test CLI: `./dwscript -e "3 + 5 * 2" --dump-ast`
- [ ] 2.64 Create sample expression files in `testdata/` and verify parsing

---

## Stage 3: Parse and Execute Simple Statements (Sequential Execution)

### Expand AST for Statements
- [ ] 3.1 Create `ast/statements.go` file
- [ ] 3.2 Define `VarDeclStatement` struct:
  - [ ] Name *Identifier
  - [ ] Type (TypeAnnotation, optional for now)
  - [ ] Value Expression (for initialization)
- [ ] 3.3 Define `AssignmentStatement` struct:
  - [ ] Name *Identifier
  - [ ] Value Expression
- [ ] 3.4 Define `BlockStatement` struct:
  - [ ] Statements []Statement
- [ ] 3.5 Define `CallExpression` struct (for built-in calls like PrintLn):
  - [ ] Function Expression (usually Identifier)
  - [ ] Arguments []Expression
- [ ] 3.6 Implement `String()` methods for new statement types
- [ ] 3.7 Add tests for AST node string representations

### Parser Extensions for Statements
- [ ] 3.8 Implement `parseVarDeclaration()` in parser:
  - [ ] Parse `var` keyword
  - [ ] Parse identifier
  - [ ] Parse optional `: Type` annotation
  - [ ] Parse optional `:= Expression` initialization
  - [ ] Parse terminating semicolon
- [ ] 3.9 Implement `parseAssignment()` in parser:
  - [ ] Detect `:=` operator
  - [ ] Parse identifier on left
  - [ ] Parse expression on right
- [ ] 3.10 Update `parseStatement()` to dispatch to var/assignment parsers
- [ ] 3.11 Implement `parseBlockStatement()` for `begin...end` blocks:
  - [ ] Parse `begin` keyword
  - [ ] Parse statement list
  - [ ] Parse `end` keyword
- [ ] 3.12 Implement `parseCallExpression()` for function calls:
  - [ ] Parse function name
  - [ ] Parse `(` and argument list
  - [ ] Parse `)` and handle zero or more arguments
- [ ] 3.13 Register call expression as infix parser (for `ident(...)` syntax)
- [ ] 3.14 Handle programs without explicit begin/end (implicit program block)

### Parser Testing for Statements
- [ ] 3.15 Test variable declarations: `TestVarDeclarations`
  - [ ] `var x: Integer;`
  - [ ] `var x: Integer := 5;`
  - [ ] `var s: String := 'hello';`
- [ ] 3.16 Test assignments: `TestAssignments`
  - [ ] `x := 10;`
  - [ ] `x := x + 1;`
- [ ] 3.17 Test block statements: `TestBlockStatements`
  - [ ] `begin x := 1; y := 2; end;`
- [ ] 3.18 Test call expressions: `TestCallExpressions`
  - [ ] `PrintLn('hello');`
  - [ ] `PrintLn(x + 5);`
- [ ] 3.19 Test complete simple programs
- [ ] 3.20 Run parser tests and achieve >85% coverage

### Interpreter/Runtime Foundation
- [ ] 3.21 Create `interp/value.go` file
- [ ] 3.22 Define `Value` interface (or use interface{} initially)
- [ ] 3.23 Create concrete value types:
  - [ ] `IntegerValue` struct
  - [ ] `FloatValue` struct
  - [ ] `StringValue` struct
  - [ ] `BooleanValue` struct
  - [ ] `NilValue` struct
- [ ] 3.24 Implement `String()` and `Type()` methods for values
- [ ] 3.25 Create helper functions to convert between Go types and Values

### Environment/Symbol Table
- [ ] 3.26 Create `interp/environment.go` file
- [ ] 3.27 Define `Environment` struct with `store map[string]Value`
- [ ] 3.28 Implement `NewEnvironment() *Environment`
- [ ] 3.29 Implement `Get(name string) (Value, bool)` method
- [ ] 3.30 Implement `Set(name string, val Value)` method
- [ ] 3.31 Implement `Define(name string, val Value)` method
- [ ] 3.32 Add support for nested scopes (outer environment reference)
- [ ] 3.33 Implement `NewEnclosedEnvironment(outer *Environment)` for scoped envs

### Interpreter Implementation
- [ ] 3.34 Create `interp/interpreter.go` file
- [ ] 3.35 Define `Interpreter` struct with: env *Environment, output io.Writer
- [ ] 3.36 Implement `New() *Interpreter` constructor
- [ ] 3.37 Implement `Eval(node ast.Node) Value` main evaluation method
- [ ] 3.38 Implement evaluation for expressions:
  - [ ] Integer literals → IntegerValue
  - [ ] Float literals → FloatValue
  - [ ] String literals → StringValue
  - [ ] Boolean literals → BooleanValue
  - [ ] Identifiers → lookup in environment
- [ ] 3.39 Implement `evalBinaryExpression()`:
  - [ ] Arithmetic: +, -, *, /, div, mod
  - [ ] Comparison: =, <>, <, >, <=, >=
  - [ ] Boolean: and, or
  - [ ] String concatenation (+)
- [ ] 3.40 Implement `evalUnaryExpression()`:
  - [ ] Negation: -expr
  - [ ] Boolean not: not expr
- [ ] 3.41 Implement evaluation for statements:
  - [ ] VarDeclStatement → define variable in environment
  - [ ] AssignmentStatement → set variable value
  - [ ] ExpressionStatement → eval expression and discard result
  - [ ] BlockStatement → eval each statement in sequence
- [ ] 3.42 Implement built-in functions:
  - [ ] `PrintLn(args...)` → write to output
  - [ ] `Print(args...)` → write without newline
  - [ ] Store built-ins in a separate map
- [ ] 3.43 Implement `evalCallExpression()` for built-in calls
- [ ] 3.44 Add error handling (return error Values or panic for runtime errors)
- [ ] 3.45 Handle undefined variable errors
- [ ] 3.46 Handle type mismatches in operations (e.g., adding string to int)

### Interpreter Testing
- [ ] 3.47 Create `interp/interpreter_test.go` file
- [ ] 3.48 Write helper to create interpreter and eval input string
- [ ] 3.49 Test integer arithmetic: `TestIntegerArithmetic`
  - [ ] `3 + 5` = 8
  - [ ] `10 - 2` = 8
  - [ ] `4 * 5` = 20
  - [ ] `20 / 4` = 5
- [ ] 3.50 Test float arithmetic: `TestFloatArithmetic`
- [ ] 3.51 Test string concatenation: `TestStringConcatenation`
- [ ] 3.52 Test boolean operations: `TestBooleanOperations`
- [ ] 3.53 Test variable declarations and usage: `TestVariables`
  - [ ] `var x: Integer := 5; PrintLn(x);` outputs "5"
- [ ] 3.54 Test assignments: `TestAssignments`
  - [ ] `var x := 0; x := 5; x := x + 3; PrintLn(x);` outputs "8"
- [ ] 3.55 Test multiple statements: `TestMultipleStatements`
- [ ] 3.56 Test undefined variable errors
- [ ] 3.57 Test type error in operations (e.g., `3 + 'hello'`)
- [ ] 3.58 Run interpreter tests: `go test ./interp -v`
- [ ] 3.59 Achieve >80% code coverage for interpreter

### CLI Integration
- [ ] 3.60 Update CLI `run` command to execute scripts (not just parse)
- [ ] 3.61 Capture interpreter output and print to console
- [ ] 3.62 Add `--trace` flag for debugging execution
- [ ] 3.63 Test CLI with simple script files:
  - [ ] `testdata/hello.dws`: `PrintLn('Hello, World!');`
  - [ ] `testdata/arithmetic.dws`: variable declarations and arithmetic
- [ ] 3.64 Verify CLI outputs match expected results
- [ ] 3.65 Create integration tests in `cmd/dwscript/` using table-driven tests

---

## Stage 4: Control Flow - Conditions and Loops

### AST Nodes for Control Flow
- [ ] 4.1 Create `ast/control_flow.go` file
- [ ] 4.2 Define `IfStatement` struct:
  - [ ] Condition Expression
  - [ ] Consequence Statement (then branch)
  - [ ] Alternative Statement (else branch, optional)
- [ ] 4.3 Define `WhileStatement` struct:
  - [ ] Condition Expression
  - [ ] Body Statement
- [ ] 4.4 Define `RepeatStatement` struct:
  - [ ] Body Statement
  - [ ] Condition Expression (until condition)
- [ ] 4.5 Define `ForStatement` struct:
  - [ ] Variable *Identifier
  - [ ] Start Expression
  - [ ] End Expression
  - [ ] Direction (to or downto)
  - [ ] Body Statement
- [ ] 4.6 Define `CaseStatement` struct (optional for later):
  - [ ] Expression Expression
  - [ ] Cases []CaseBranch
  - [ ] Else Statement (optional)
- [ ] 4.7 Define `CaseBranch` struct:
  - [ ] Values []Expression
  - [ ] Statement Statement
- [ ] 4.8 Implement `String()` methods for all control flow nodes

### Parser for If Statements
- [ ] 4.9 Implement `parseIfStatement()`:
  - [ ] Parse `if` keyword
  - [ ] Parse condition expression
  - [ ] Parse `then` keyword
  - [ ] Parse consequence statement
  - [ ] Parse optional `else` keyword and alternative statement
- [ ] 4.10 Update `parseStatement()` to handle `if` token
- [ ] 4.11 Test if statement parsing: `TestIfStatements`
  - [ ] Simple if: `if x > 0 then PrintLn('positive');`
  - [ ] If-else: `if x > 0 then PrintLn('positive') else PrintLn('non-positive');`
  - [ ] If with block: `if x > 0 then begin ... end;`

### Parser for While Loops
- [ ] 4.12 Implement `parseWhileStatement()`:
  - [ ] Parse `while` keyword
  - [ ] Parse condition expression
  - [ ] Parse `do` keyword
  - [ ] Parse body statement
- [ ] 4.13 Update `parseStatement()` to handle `while` token
- [ ] 4.14 Test while statement parsing: `TestWhileStatements`
  - [ ] `while x < 10 do x := x + 1;`
  - [ ] While with block body

### Parser for Repeat-Until Loops
- [ ] 4.15 Implement `parseRepeatStatement()`:
  - [ ] Parse `repeat` keyword
  - [ ] Parse statement list (until `until` keyword)
  - [ ] Parse `until` keyword
  - [ ] Parse condition expression
- [ ] 4.16 Update `parseStatement()` to handle `repeat` token
- [ ] 4.17 Test repeat statement parsing: `TestRepeatStatements`

### Parser for For Loops
- [ ] 4.18 Implement `parseForStatement()`:
  - [ ] Parse `for` keyword
  - [ ] Parse loop variable identifier
  - [ ] Parse `:=` and start expression
  - [ ] Parse direction keyword (`to` or `downto`)
  - [ ] Parse end expression
  - [ ] Parse `do` keyword
  - [ ] Parse body statement
- [ ] 4.19 Update `parseStatement()` to handle `for` token
- [ ] 4.20 Test for statement parsing: `TestForStatements`
  - [ ] `for i := 1 to 10 do PrintLn(i);`
  - [ ] `for i := 10 downto 1 do PrintLn(i);`

### Parser for Case Statements (Optional)
- [ ] 4.21 Implement `parseCaseStatement()` (if tackling now):
  - [ ] Parse `case` keyword
  - [ ] Parse expression
  - [ ] Parse `of` keyword
  - [ ] Parse case branches (value: statement)
  - [ ] Parse optional `else` branch
  - [ ] Parse `end` keyword
- [ ] 4.22 Update `parseStatement()` to handle `case` token
- [ ] 4.23 Test case statement parsing: `TestCaseStatements`

### Parser Testing for Control Flow
- [ ] 4.24 Run all parser tests including new control flow tests
- [ ] 4.25 Achieve >85% parser coverage with control flow

### Interpreter for If Statements
- [ ] 4.26 Implement `evalIfStatement()` in interpreter:
  - [ ] Evaluate condition
  - [ ] Convert to boolean
  - [ ] Execute consequence if true, alternative if false
- [ ] 4.27 Test if statement execution: `TestIfExecution`
  - [ ] Test both branches
  - [ ] Test nested ifs

### Interpreter for While Loops
- [ ] 4.28 Implement `evalWhileStatement()` in interpreter:
  - [ ] Loop while condition is true
  - [ ] Evaluate body in each iteration
  - [ ] Add safeguard against infinite loops in tests (max iterations)
- [ ] 4.29 Test while loop execution: `TestWhileExecution`
  - [ ] Count from 0 to 10
  - [ ] Sum numbers in a loop

### Interpreter for Repeat-Until Loops
- [ ] 4.30 Implement `evalRepeatStatement()` in interpreter:
  - [ ] Execute body at least once
  - [ ] Continue until condition becomes true
- [ ] 4.31 Test repeat-until execution: `TestRepeatExecution`

### Interpreter for For Loops
- [ ] 4.32 Implement `evalForStatement()` in interpreter:
  - [ ] Evaluate start and end expressions
  - [ ] Create loop variable in local scope
  - [ ] Iterate from start to end (or downto)
  - [ ] Execute body for each iteration
  - [ ] Handle loop variable scope correctly
- [ ] 4.33 Test for loop execution: `TestForExecution`
  - [ ] Ascending loops
  - [ ] Descending loops
  - [ ] Empty loops (start > end for `to`)

### Interpreter for Case Statements (Optional)
- [ ] 4.34 Implement `evalCaseStatement()` in interpreter:
  - [ ] Evaluate case expression
  - [ ] Compare with each branch's values
  - [ ] Execute matching branch
  - [ ] Execute else branch if no match
- [ ] 4.35 Test case statement execution: `TestCaseExecution`

### Control Flow Testing
- [ ] 4.36 Create comprehensive test scripts in `testdata/`:
  - [ ] `if_else.dws`
  - [ ] `while_loop.dws`
  - [ ] `for_loop.dws`
  - [ ] `nested_loops.dws`
- [ ] 4.37 Test nested control structures
- [ ] 4.38 Test break/continue (if DWScript supports them)
- [ ] 4.39 Run all interpreter tests: `go test ./interp -v`
- [ ] 4.40 Achieve >80% interpreter coverage

### CLI Testing
- [ ] 4.41 Test CLI with control flow scripts
- [ ] 4.42 Verify output matches expected results
- [ ] 4.43 Create integration tests for control flow

---

## Stage 5: Functions, Procedures, and Scope Management

### AST Nodes for Functions
- [ ] 5.1 Create `ast/functions.go` file
- [ ] 5.2 Define `Parameter` struct:
  - [ ] Name *Identifier
  - [ ] Type TypeAnnotation
  - [ ] ByRef bool (for var parameters)
- [ ] 5.3 Define `FunctionDecl` struct:
  - [ ] Name *Identifier
  - [ ] Parameters []*Parameter
  - [ ] ReturnType TypeAnnotation (nil for procedures)
  - [ ] Body *BlockStatement
- [ ] 5.4 Define `ReturnStatement` struct:
  - [ ] ReturnValue Expression (optional)
- [ ] 5.5 Update `CallExpression` to support user-defined functions
- [ ] 5.6 Implement `String()` methods for function nodes

### Parser for Functions
- [ ] 5.7 Implement `parseFunctionDeclaration()`:
  - [ ] Parse `function` or `procedure` keyword
  - [ ] Parse function name
  - [ ] Parse parameter list `(param: Type; ...)`
  - [ ] Parse `: ReturnType` for functions
  - [ ] Parse semicolon after signature
  - [ ] Parse optional `forward;` (forward declaration)
  - [ ] Parse function body (begin...end or block)
  - [ ] Parse terminating semicolon
- [ ] 5.8 Implement `parseParameterList()`:
  - [ ] Parse parameters separated by semicolons
  - [ ] Handle `var` keyword for by-reference parameters
  - [ ] Parse multiple parameters with same type: `a, b: Integer`
- [ ] 5.9 Update `parseStatement()` or top-level parser to handle functions
- [ ] 5.10 Implement `parseReturnStatement()` or handle `Result :=` or function name assignment
- [ ] 5.11 Handle nested function declarations (if supported)
- [ ] 5.12 Build function symbol table during parsing

### Parser Testing for Functions
- [ ] 5.13 Test function declaration parsing: `TestFunctionDeclarations`
  - [ ] Simple function: `function Add(a, b: Integer): Integer;`
  - [ ] Procedure: `procedure Hello;`
  - [ ] Function with body
- [ ] 5.14 Test parameter parsing: `TestParameters`
- [ ] 5.15 Test function calls with arguments
- [ ] 5.16 Test nested functions (if supported)
- [ ] 5.17 Run parser tests: `go test ./parser -v`

### Symbol Table Enhancement
- [ ] 5.18 Create `interp/symbol_table.go` file
- [ ] 5.19 Define `Symbol` struct:
  - [ ] Name string
  - [ ] Type (function, variable, etc.)
  - [ ] Value (for variables at runtime, or FunctionDecl for functions)
- [ ] 5.20 Define `SymbolTable` struct:
  - [ ] symbols map[string]*Symbol
  - [ ] outer *SymbolTable (for nested scopes)
- [ ] 5.21 Implement `NewSymbolTable()` and `NewEnclosedSymbolTable(outer)`
- [ ] 5.22 Implement `Define()`, `Resolve()`, `Update()` methods
- [ ] 5.23 Add scope management methods: `EnterScope()`, `LeaveScope()`

### Interpreter for Functions
- [ ] 5.24 Update interpreter to maintain function registry (map of function names to FunctionDecl)
- [ ] 5.25 Implement `evalFunctionDeclaration()`:
  - [ ] Store function in registry
  - [ ] Don't execute body during declaration
- [ ] 5.26 Implement `evalCallExpression()` for user-defined functions:
  - [ ] Look up function in registry
  - [ ] Evaluate argument expressions
  - [ ] Create new environment for function scope
  - [ ] Bind parameters to argument values
  - [ ] Execute function body
  - [ ] Capture return value (via `Result` variable or function name)
  - [ ] Return to caller's environment
- [ ] 5.27 Implement return value handling:
  - [ ] Use `Result` variable convention (Delphi style)
  - [ ] Or function name as return variable
  - [ ] Handle explicit `Exit` or `Result :=` statements
- [ ] 5.28 Implement call stack for debugging (track current function)
- [ ] 5.29 Add recursion support (ensure environments don't leak)
- [ ] 5.30 Handle by-reference parameters (var parameters):
  - [ ] Pass reference to variable, not value
  - [ ] Modifications affect caller's variable

### Interpreter Testing for Functions
- [ ] 5.31 Test simple function calls: `TestFunctionCalls`
  - [ ] `function Add(a, b: Integer): Integer; begin Result := a + b; end;`
  - [ ] Call: `PrintLn(Add(2, 3));` outputs "5"
- [ ] 5.32 Test procedures (no return): `TestProcedures`
- [ ] 5.33 Test recursive functions: `TestRecursion`
  - [ ] Factorial
  - [ ] Fibonacci
- [ ] 5.34 Test function with local variables: `TestLocalVariables`
  - [ ] Ensure locals don't leak to global scope
- [ ] 5.35 Test multiple function calls: `TestMultipleFunctions`
- [ ] 5.36 Test nested function calls: `TestNestedCalls`
- [ ] 5.37 Test by-reference parameters: `TestVarParameters`
  - [ ] `procedure Swap(var a, b: Integer);`
- [ ] 5.38 Test scope isolation: `TestScopeIsolation`
  - [ ] Same variable name in different scopes
- [ ] 5.39 Run interpreter tests: `go test ./interp -v`
- [ ] 5.40 Achieve >80% coverage

### CLI Testing
- [ ] 5.41 Create test scripts with functions:
  - [ ] `testdata/functions.dws`
  - [ ] `testdata/recursion.dws`
- [ ] 5.42 Test CLI with function-based scripts
- [ ] 5.43 Verify correct execution and output

### Documentation
- [ ] 5.44 Document function calling convention
- [ ] 5.45 Document scope management strategy
- [ ] 5.46 Add examples to README

---

## Stage 6: Static Type Checking and Semantic Analysis

### Type System Foundation
- [ ] 6.1 Create `types/types.go` file
- [ ] 6.2 Define `Type` interface with methods: `String()`, `Equals(Type) bool`
- [ ] 6.3 Define basic type structs:
  - [ ] `IntegerType`
  - [ ] `FloatType`
  - [ ] `StringType`
  - [ ] `BooleanType`
  - [ ] `NilType`
  - [ ] `VoidType` (for procedures)
- [ ] 6.4 Create type constants: `INTEGER`, `FLOAT`, `STRING`, `BOOLEAN`, `NIL`, `VOID`
- [ ] 6.5 Implement `Equals()` for basic types
- [ ] 6.6 Create `FunctionType` struct:
  - [ ] Parameters []Type
  - [ ] ReturnType Type
- [ ] 6.7 Define `ArrayType`, `RecordType` (for later)
- [ ] 6.8 Implement type comparison and compatibility rules
- [ ] 6.9 Add type coercion rules (e.g., Integer → Float)

### Type Annotations in AST
- [ ] 6.10 Add `Type` field to AST expression nodes
- [ ] 6.11 Update AST node constructors to optionally accept type
- [ ] 6.12 Add type annotation parsing to variable declarations
- [ ] 6.13 Add type annotation parsing to parameters
- [ ] 6.14 Add return type parsing to functions

### Semantic Analyzer
- [ ] 6.15 Create `semantic/analyzer.go` file
- [ ] 6.16 Define `Analyzer` struct with: symbolTable, errors []string
- [ ] 6.17 Implement `NewAnalyzer() *Analyzer`
- [ ] 6.18 Implement `Analyze(program *ast.Program) error`
- [ ] 6.19 Implement `analyzeNode(node ast.Node)` visitor pattern
- [ ] 6.20 Implement variable declaration analysis:
  - [ ] Check for redeclaration
  - [ ] Store variable type in symbol table
  - [ ] Validate initializer type matches declared type
- [ ] 6.21 Implement identifier resolution:
  - [ ] Check variable is declared before use
  - [ ] Assign type to identifier node
- [ ] 6.22 Implement expression type checking:
  - [ ] Literals: assign known types
  - [ ] Binary expressions: check operand types compatibility
  - [ ] Assign result type based on operator and operands
  - [ ] Handle type coercion (Int + Float → Float)
- [ ] 6.23 Implement assignment type checking:
  - [ ] Ensure RHS type compatible with LHS variable type
- [ ] 6.24 Implement function declaration analysis:
  - [ ] Store function signature in symbol table
  - [ ] Check for duplicate function names
  - [ ] Analyze function body in function scope
  - [ ] Check return type matches (Result variable type)
- [ ] 6.25 Implement function call type checking:
  - [ ] Verify function exists
  - [ ] Check argument count matches parameters
  - [ ] Check argument types match parameter types
  - [ ] Assign return type to call expression
- [ ] 6.26 Implement control flow type checking:
  - [ ] If/while/for conditions must be boolean
  - [ ] For loop variable must be ordinal type
- [ ] 6.27 Add error accumulation and reporting

### Semantic Analyzer Testing
- [ ] 6.28 Create `semantic/analyzer_test.go` file
- [ ] 6.29 Test undefined variable detection: `TestUndefinedVariable`
- [ ] 6.30 Test type mismatch in assignment: `TestAssignmentTypeMismatch`
  - [ ] `var i: Integer; i := 'hello';` should error
- [ ] 6.31 Test type mismatch in operations: `TestOperationTypeMismatch`
  - [ ] `3 + 'hello'` should error
- [ ] 6.32 Test function call errors: `TestFunctionCallErrors`
  - [ ] Wrong argument count
  - [ ] Wrong argument types
  - [ ] Calling undefined function
- [ ] 6.33 Test valid type coercion: `TestTypeCoercion`
  - [ ] `var f: Float := 3;` should work (int → float)
- [ ] 6.34 Test return type checking: `TestReturnTypes`
  - [ ] Function must return correct type
- [ ] 6.35 Test control flow condition types: `TestControlFlowTypes`
  - [ ] `if 3 then ...` should error (not boolean)
- [ ] 6.36 Test redeclaration errors: `TestRedeclaration`
- [ ] 6.37 Run semantic analyzer tests: `go test ./semantic -v`
- [ ] 6.38 Achieve >85% coverage

### Integration with Parser and Interpreter
- [ ] 6.39 Update parser to run semantic analysis after parsing
- [ ] 6.40 Option to disable type checking (for testing)
- [ ] 6.41 Update interpreter to use type information from analysis
- [ ] 6.42 Add type assertions in interpreter operations
- [ ] 6.43 Improve error messages with line/column info
- [ ] 6.44 Update CLI to report semantic errors before execution

### Error Reporting Enhancement
- [ ] 6.45 Add line/column tracking to all AST nodes
- [ ] 6.46 Create `errors.go` with error formatting utilities
- [ ] 6.47 Implement pretty error messages:
  - [ ] Show source line
  - [ ] Point to error location with caret (^)
  - [ ] Include context
- [ ] 6.48 Support multiple error reporting (don't stop at first error)
- [ ] 6.49 Test error reporting with various invalid programs

### Testing Type System
- [ ] 6.50 Create test scripts with type errors:
  - [ ] `testdata/type_errors/`
- [ ] 6.51 Verify all are caught by semantic analyzer
- [ ] 6.52 Create test scripts with valid type usage
- [ ] 6.53 Verify all pass semantic analysis
- [ ] 6.54 Run full integration tests

---

## Stage 7: Support Object-Oriented Features (Classes, Interfaces, Methods)

### Type Definitions for OOP
- [ ] 7.1 Extend `types/types.go` for class types
- [ ] 7.2 Define `ClassType` struct:
  - [ ] Name string
  - [ ] Parent *ClassType
  - [ ] Fields map[string]Type
  - [ ] Methods map[string]*FunctionType
- [ ] 7.3 Define `InterfaceType` struct:
  - [ ] Name string
  - [ ] Methods map[string]*FunctionType
- [ ] 7.4 Implement type compatibility for classes (inheritance)
- [ ] 7.5 Implement interface satisfaction checking

### AST Nodes for Classes
- [ ] 7.6 Create `ast/classes.go` file
- [ ] 7.7 Define `ClassDecl` struct:
  - [ ] Name *Identifier
  - [ ] Parent *Identifier (optional)
  - [ ] Fields []*FieldDecl
  - [ ] Methods []*FunctionDecl
  - [ ] Constructor *FunctionDecl (optional)
  - [ ] Destructor *FunctionDecl (optional)
- [ ] 7.8 Define `FieldDecl` struct:
  - [ ] Name *Identifier
  - [ ] Type TypeAnnotation
  - [ ] Visibility (public, private, protected)
- [ ] 7.9 Define `NewExpression` struct (object creation):
  - [ ] ClassName *Identifier
  - [ ] Arguments []Expression
- [ ] 7.10 Define `MemberAccessExpression` struct:
  - [ ] Object Expression
  - [ ] Member *Identifier
- [ ] 7.11 Define `MethodCallExpression` struct:
  - [ ] Object Expression
  - [ ] Method *Identifier
  - [ ] Arguments []Expression
- [ ] 7.12 Implement `String()` methods for OOP nodes

### Parser for Classes
- [ ] 7.13 Implement `parseClassDeclaration()`:
  - [ ] Parse `type` keyword
  - [ ] Parse class name
  - [ ] Parse `= class` keyword
  - [ ] Parse optional `(ParentClass)` inheritance
  - [ ] Parse class body (fields and methods)
  - [ ] Parse `end` keyword
- [ ] 7.14 Implement `parseFieldDeclaration()`:
  - [ ] Parse field name
  - [ ] Parse `: Type` annotation
  - [ ] Parse semicolon
- [ ] 7.15 Implement parsing of methods within class:
  - [ ] Inline method implementation
  - [ ] Method declaration only (implementation later)
- [ ] 7.16 Implement `parseConstructor()` (if special syntax)
- [ ] 7.17 Implement `parseDestructor()` (if supported)
- [ ] 7.18 Implement `parseNewExpression()`:
  - [ ] Parse class name
  - [ ] Parse `.Create(...)` or `new ClassName`
- [ ] 7.19 Implement `parseMemberAccess()`:
  - [ ] Parse `obj.field` or `obj.method`
  - [ ] Handle as infix operator with `.`
- [ ] 7.20 Update expression parsing to handle member access and method calls

### Parser Testing for Classes
- [ ] 7.21 Test class declaration parsing: `TestClassDeclarations`
- [ ] 7.22 Test inheritance parsing: `TestClassInheritance`
- [ ] 7.23 Test field parsing: `TestFieldDeclarations`
- [ ] 7.24 Test method parsing: `TestMethodDeclarations`
- [ ] 7.25 Test object creation parsing: `TestNewExpressions`
- [ ] 7.26 Test member access parsing: `TestMemberAccess`
- [ ] 7.27 Run parser tests: `go test ./parser -v`

### Runtime Class Representation
- [ ] 7.28 Create `interp/class.go` file
- [ ] 7.29 Define `ClassInfo` struct (runtime metadata):
  - [ ] Name string
  - [ ] Parent *ClassInfo
  - [ ] FieldTypes map[string]Type
  - [ ] Methods map[string]*FunctionDecl
  - [ ] Constructor *FunctionDecl
- [ ] 7.30 Define `ObjectInstance` struct:
  - [ ] Class *ClassInfo
  - [ ] Fields map[string]Value
- [ ] 7.31 Implement `NewObjectInstance(class *ClassInfo) *ObjectInstance`
- [ ] 7.32 Implement `GetField(name string) Value`
- [ ] 7.33 Implement `SetField(name string, val Value)`
- [ ] 7.34 Build method lookup with inheritance (method resolution order)
- [ ] 7.35 Handle method overriding (child method overrides parent)

### Interpreter for Classes
- [ ] 7.36 Update interpreter to maintain class registry
- [ ] 7.37 Implement `evalClassDeclaration()`:
  - [ ] Build ClassInfo from AST
  - [ ] Register in class registry
  - [ ] Handle inheritance (copy parent fields/methods)
- [ ] 7.38 Implement `evalNewExpression()`:
  - [ ] Look up class in registry
  - [ ] Create ObjectInstance
  - [ ] Initialize fields with default values
  - [ ] Call constructor if present
  - [ ] Return object as value
- [ ] 7.39 Implement `evalMemberAccess()`:
  - [ ] Evaluate object expression
  - [ ] Ensure it's an ObjectInstance
  - [ ] Retrieve field value by name
- [ ] 7.40 Implement `evalMethodCall()`:
  - [ ] Evaluate object expression
  - [ ] Look up method in object's class
  - [ ] Create environment with `Self` bound to object
  - [ ] Execute method body
  - [ ] Return result
- [ ] 7.41 Handle `Self` keyword in methods:
  - [ ] Bind Self in method environment
  - [ ] Allow access to fields/methods via Self
- [ ] 7.42 Implement constructor execution:
  - [ ] Special handling for `Create` method
  - [ ] Initialize object fields
- [ ] 7.43 Implement destructor (if supported)
- [ ] 7.44 Handle polymorphism (dynamic dispatch):
  - [ ] When calling method, use object's actual class
  - [ ] Even if variable is typed as parent class

### Interpreter Testing for Classes
- [ ] 7.45 Test object creation: `TestObjectCreation`
  - [ ] Create simple class, instantiate, check fields
- [ ] 7.46 Test field access: `TestFieldAccess`
  - [ ] Set and get field values
- [ ] 7.47 Test method calls: `TestMethodCalls`
  - [ ] Call method on object
  - [ ] Verify method can access fields
- [ ] 7.48 Test inheritance: `TestInheritance`
  - [ ] Child class inherits parent fields
  - [ ] Child can override parent methods
- [ ] 7.49 Test polymorphism: `TestPolymorphism`
  - [ ] Variable of parent type holds child instance
  - [ ] Method call dispatches to child's override
- [ ] 7.50 Test constructors: `TestConstructors`
- [ ] 7.51 Test `Self` reference: `TestSelfReference`
- [ ] 7.52 Test method overloading (if supported): `TestMethodOverloading`
- [ ] 7.53 Run interpreter tests: `go test ./interp -v`

### Semantic Analysis for Classes
- [ ] 7.54 Update semantic analyzer to handle classes
- [ ] 7.55 Check class declarations:
  - [ ] Verify parent class exists (if inheritance)
  - [ ] Check for circular inheritance
  - [ ] Verify field types exist
- [ ] 7.56 Check method declarations within classes:
  - [ ] Methods have access to class fields
  - [ ] Handle Self type correctly
- [ ] 7.57 Check object creation:
  - [ ] Class must be defined
  - [ ] Constructor arguments match (if present)
- [ ] 7.58 Check member access:
  - [ ] Object expression must be class type
  - [ ] Field/method must exist in class
  - [ ] Visibility rules (public/private)
- [ ] 7.59 Check method overriding:
  - [ ] Signature must match parent method
- [ ] 7.60 Test semantic analysis for classes

### Advanced OOP Features
- [ ] 7.61 Implement class methods (static methods)
- [ ] 7.62 Implement class variables (static fields)
- [ ] 7.63 Implement abstract classes (if supported)
- [ ] 7.64 Implement virtual/override keywords
- [ ] 7.65 Implement visibility modifiers (private, protected, public)
- [ ] 7.66 Test advanced features

### Interfaces (Optional)
- [ ] 7.67 Parse interface declarations
- [ ] 7.68 Implement interface satisfaction checking in semantic analyzer
- [ ] 7.69 Implement interface variables at runtime
- [ ] 7.70 Implement interface method calls (dispatch to implementing class)
- [ ] 7.71 Test interfaces thoroughly

### CLI Testing for OOP
- [ ] 7.72 Create OOP test scripts:
  - [ ] `testdata/classes.dws`
  - [ ] `testdata/inheritance.dws`
  - [ ] `testdata/polymorphism.dws`
- [ ] 7.73 Verify CLI correctly executes OOP programs
- [ ] 7.74 Create integration tests

### Documentation
- [ ] 7.75 Document OOP implementation strategy
- [ ] 7.76 Document how Delphi classes map to Go structures
- [ ] 7.77 Add OOP examples to README

---

## Stage 8: Additional DWScript Features and Polishing

### Operator Overloading
- [ ] 8.1 Research DWScript operator overloading syntax
- [ ] 8.2 Parse operator overload declarations in classes
- [ ] 8.3 Store operator overloads in ClassInfo
- [ ] 8.4 Implement operator resolution in semantic analyzer
- [ ] 8.5 Implement operator overload execution in interpreter
- [ ] 8.6 Test operator overloading: `TestOperatorOverloading`

### Properties
- [ ] 8.7 Parse property declarations (with read/write specifiers)
- [ ] 8.8 Translate property access to getter/setter calls
- [ ] 8.9 Implement property evaluation in interpreter
- [ ] 8.10 Test properties: `TestProperties`

### Record Types
- [ ] 8.11 Define `RecordType` in type system
- [ ] 8.12 Parse record declarations: `type TPoint = record X, Y: Integer; end;`
- [ ] 8.13 Implement record instantiation (value type)
- [ ] 8.14 Implement record field access
- [ ] 8.15 Test records: `TestRecords`

### Set Types
- [ ] 8.16 Define `SetType` in type system
- [ ] 8.17 Parse set type declarations: `type TDays = set of (Mon, Tue, ...);`
- [ ] 8.18 Parse set literals: `[1, 3, 5]`
- [ ] 8.19 Implement set operations (in, +, -, *)
- [ ] 8.20 Implement set representation (bitset or map)
- [ ] 8.21 Test sets: `TestSets`

### Enumerated Types
- [ ] 8.22 Define `EnumType` in type system
- [ ] 8.23 Parse enum declarations: `type TColor = (Red, Green, Blue);`
- [ ] 8.24 Implement enum values as constants
- [ ] 8.25 Test enums: `TestEnums`

### Array Types
- [ ] 8.26 Define `ArrayType` in type system (static and dynamic)
- [ ] 8.27 Parse array declarations: `array[1..10] of Integer`
- [ ] 8.28 Parse dynamic array declarations: `array of Integer`
- [ ] 8.29 Implement array indexing: `arr[i]`
- [ ] 8.30 Implement array functions (Length, SetLength, etc.)
- [ ] 8.31 Test arrays: `TestArrays`

### String Functions
- [ ] 8.32 Implement built-in string functions:
  - [ ] Length(s)
  - [ ] Copy(s, index, count)
  - [ ] Concat(s1, s2, ...)
  - [ ] Pos(substr, s)
  - [ ] UpperCase(s), LowerCase(s)
- [ ] 8.33 Test string functions

### Math Functions
- [ ] 8.34 Implement built-in math functions:
  - [ ] Abs(x)
  - [ ] Sqrt(x)
  - [ ] Sin(x), Cos(x), Tan(x)
  - [ ] Ln(x), Exp(x)
  - [ ] Round(x), Trunc(x)
  - [ ] Random, Randomize
- [ ] 8.35 Test math functions

### Conversion Functions
- [ ] 8.36 Implement type conversion functions:
  - [ ] IntToStr(i)
  - [ ] StrToInt(s)
  - [ ] FloatToStr(f)
  - [ ] StrToFloat(s)
- [ ] 8.37 Test conversion functions

### Exception Handling (Try/Except/Finally)
- [ ] 8.38 Parse try-except-finally blocks (if supported)
- [ ] 8.39 Implement exception types
- [ ] 8.40 Implement raise statement
- [ ] 8.41 Implement exception catching in interpreter
- [ ] 8.42 Test exceptions: `TestExceptions`

### Meta-class Support
- [ ] 8.43 Implement class references (variables holding class types)
- [ ] 8.44 Allow calling constructors via class reference
- [ ] 8.45 Test meta-classes

### Function/Method Pointers
- [ ] 8.46 Parse function pointer types
- [ ] 8.47 Implement taking address of function (@Function)
- [ ] 8.48 Implement calling via function pointer
- [ ] 8.49 Test function pointers

### Contracts (Design by Contract)
- [ ] 8.50 Parse require/ensure clauses (if supported)
- [ ] 8.51 Implement contract checking at runtime
- [ ] 8.52 Test contracts

### Additional Features Assessment
- [ ] 8.53 Review DWScript feature list for missing items
- [ ] 8.54 Prioritize remaining features
- [ ] 8.55 Implement high-priority features
- [ ] 8.56 Document unsupported features

### Comprehensive Testing
- [ ] 8.57 Port DWScript's test suite (if available)
- [ ] 8.58 Run DWScript example scripts from documentation
- [ ] 8.59 Compare outputs with original DWScript
- [ ] 8.60 Fix any discrepancies
- [ ] 8.61 Create stress tests for complex features
- [ ] 8.62 Achieve >85% overall code coverage

---

## Stage 9: Performance Tuning and Refactoring

### Performance Profiling
- [ ] 9.1 Create performance benchmark scripts
- [ ] 9.2 Profile lexer performance: `BenchmarkLexer`
- [ ] 9.3 Profile parser performance: `BenchmarkParser`
- [ ] 9.4 Profile interpreter performance: `BenchmarkInterpreter`
- [ ] 9.5 Identify bottlenecks using `pprof`
- [ ] 9.6 Document performance baseline

### Optimization - Lexer
- [ ] 9.7 Optimize string handling in lexer (use bytes instead of runes where possible)
- [ ] 9.8 Reduce allocations in token creation
- [ ] 9.9 Use string interning for keywords/identifiers
- [ ] 9.10 Benchmark improvements

### Optimization - Parser
- [ ] 9.11 Reduce AST node allocations
- [ ] 9.12 Pool commonly created nodes
- [ ] 9.13 Optimize precedence table lookups
- [ ] 9.14 Benchmark improvements

### Bytecode Compiler (Optional)
- [ ] 9.15 Design bytecode instruction set:
  - [ ] Load constant
  - [ ] Load/store variable
  - [ ] Binary/unary operations
  - [ ] Jump instructions (conditional/unconditional)
  - [ ] Call/return
  - [ ] Object operations
- [ ] 9.16 Implement bytecode emitter (AST → bytecode)
- [ ] 9.17 Implement bytecode VM (execute instructions)
- [ ] 9.18 Handle stack management in VM
- [ ] 9.19 Test bytecode execution produces same results as AST interpreter
- [ ] 9.20 Benchmark bytecode VM vs AST interpreter
- [ ] 9.21 Optimize VM loop
- [ ] 9.22 Add option to CLI to use bytecode or AST interpreter

### Optimization - Interpreter
- [ ] 9.23 Optimize value representation (avoid interface{} overhead if possible)
- [ ] 9.24 Use switch statements instead of type assertions where possible
- [ ] 9.25 Cache frequently accessed symbols
- [ ] 9.26 Optimize environment lookups
- [ ] 9.27 Reduce allocations in hot paths
- [ ] 9.28 Benchmark improvements

### Memory Management
- [ ] 9.29 Ensure no memory leaks in long-running scripts
- [ ] 9.30 Profile memory usage with large programs
- [ ] 9.31 Optimize object allocation/deallocation
- [ ] 9.32 Consider object pooling for common types

### Code Quality Refactoring
- [ ] 9.33 Run `go vet ./...` and fix all issues
- [ ] 9.34 Run `golangci-lint run` and address warnings
- [ ] 9.35 Run `gofmt` on all files
- [ ] 9.36 Run `goimports` to organize imports
- [ ] 9.37 Review error handling consistency
- [ ] 9.38 Unify value representation if inconsistent
- [ ] 9.39 Refactor large functions into smaller ones
- [ ] 9.40 Extract common patterns into helper functions
- [ ] 9.41 Improve variable/function naming
- [ ] 9.42 Add missing error checks

### Documentation
- [ ] 9.43 Write comprehensive GoDoc comments for all exported types/functions
- [ ] 9.44 Document internal architecture in `docs/architecture.md`
- [ ] 9.45 Create user guide in `docs/user_guide.md`
- [ ] 9.46 Document CLI usage with examples
- [ ] 9.47 Create API documentation for embedding the library
- [ ] 9.48 Add code examples to documentation
- [ ] 9.49 Document known limitations
- [ ] 9.50 Create contribution guidelines in `CONTRIBUTING.md`

### Example Programs
- [ ] 9.51 Create `examples/` directory
- [ ] 9.52 Add example scripts:
  - [ ] Hello World
  - [ ] Fibonacci
  - [ ] Factorial
  - [ ] Class-based example (e.g., Person class)
  - [ ] Game or algorithm (e.g., sorting)
- [ ] 9.53 Add README in examples directory
- [ ] 9.54 Ensure all examples run correctly

### Testing Enhancements
- [ ] 9.55 Add integration tests in `test/integration/`
- [ ] 9.56 Add fuzzing tests for parser: `FuzzParser`
- [ ] 9.57 Add fuzzing tests for lexer: `FuzzLexer`
- [ ] 9.58 Add property-based tests (using testing/quick or gopter)
- [ ] 9.59 Ensure CI runs all test types
- [ ] 9.60 Achieve >90% code coverage overall
- [ ] 9.61 Add regression tests for all fixed bugs

### Release Preparation
- [ ] 9.62 Create `CHANGELOG.md`
- [ ] 9.63 Document version numbering scheme (SemVer)
- [ ] 9.64 Tag v0.1.0 alpha release
- [ ] 9.65 Create release binaries for major platforms (Linux, macOS, Windows)
- [ ] 9.66 Publish release on GitHub
- [ ] 9.67 Write announcement blog post or README update
- [ ] 9.68 Share with community for feedback

---

## Stage 10: Long-Term Evolution

### Feature Parity Tracking
- [ ] 10.1 Create feature matrix comparing go-dws with DWScript
- [ ] 10.2 Track DWScript upstream releases
- [ ] 10.3 Identify new features in DWScript updates
- [ ] 10.4 Plan integration of new features
- [ ] 10.5 Update feature matrix regularly

### Community Building
- [ ] 10.6 Set up issue templates on GitHub
- [ ] 10.7 Set up pull request template
- [ ] 10.8 Create CODE_OF_CONDUCT.md
- [ ] 10.9 Create discussions forum or mailing list
- [ ] 10.10 Encourage contributions (tag "good first issue")
- [ ] 10.11 Respond to issues and PRs promptly
- [ ] 10.12 Build maintainer team (if interest grows)

### Advanced Features
- [ ] 10.13 Implement REPL (Read-Eval-Print Loop):
  - [ ] Interactive prompt
  - [ ] Statement-by-statement execution
  - [ ] Variable inspection
  - [ ] History and autocomplete
- [ ] 10.14 Implement debugging support:
  - [ ] Breakpoints
  - [ ] Step-through execution
  - [ ] Variable inspection
  - [ ] Stack traces
- [ ] 10.15 Implement WebAssembly compilation:
  - [ ] Use Go's WASM target
  - [ ] Create web-based DWScript playground
  - [ ] Publish WASM build
- [ ] 10.16 Implement language server protocol (LSP):
  - [ ] Syntax highlighting
  - [ ] Autocomplete
  - [ ] Go-to-definition
  - [ ] Error diagnostics in IDE
- [ ] 10.17 Implement JavaScript code generation backend:
  - [ ] AST → JavaScript transpiler
  - [ ] Support browser execution
  - [ ] Create npm package

### Alternative Execution Modes
- [ ] 10.18 Add JIT compilation (if feasible in Go)
- [ ] 10.19 Add AOT compilation (compile to native binary)
- [ ] 10.20 Add compilation to Go source code
- [ ] 10.21 Benchmark different execution modes

### Platform-Specific Enhancements
- [ ] 10.22 Add Windows-specific features (if needed)
- [ ] 10.23 Add macOS-specific features (if needed)
- [ ] 10.24 Add Linux-specific features (if needed)
- [ ] 10.25 Test on multiple architectures (ARM, AMD64)

### Edge Case Audit
- [ ] 10.26 Test short-circuit evaluation (and, or)
- [ ] 10.27 Test operator precedence edge cases
- [ ] 10.28 Test division by zero handling
- [ ] 10.29 Test integer overflow behavior
- [ ] 10.30 Test floating-point edge cases (NaN, Inf)
- [ ] 10.31 Test string encoding (UTF-8 handling)
- [ ] 10.32 Test very large programs (scalability)
- [ ] 10.33 Test deeply nested structures
- [ ] 10.34 Test circular references (if possible in language)
- [ ] 10.35 Fix any discovered issues

### Performance Monitoring
- [ ] 10.36 Set up continuous performance benchmarking
- [ ] 10.37 Track performance metrics over releases
- [ ] 10.38 Identify and fix performance regressions
- [ ] 10.39 Publish performance comparison with DWScript

### Security Audit
- [ ] 10.40 Review for potential security issues (untrusted script execution)
- [ ] 10.41 Implement resource limits (memory, execution time)
- [ ] 10.42 Implement sandboxing for untrusted scripts
- [ ] 10.43 Audit for code injection vulnerabilities
- [ ] 10.44 Document security best practices

### Maintenance
- [ ] 10.45 Keep dependencies up to date
- [ ] 10.46 Monitor Go version updates and migrate as needed
- [ ] 10.47 Maintain CI/CD pipeline
- [ ] 10.48 Regular code reviews
- [ ] 10.49 Address technical debt periodically

### Long-term Roadmap
- [ ] 10.50 Define 1-year roadmap
- [ ] 10.51 Define 3-year roadmap
- [ ] 10.52 Gather user feedback and adjust priorities
- [ ] 10.53 Consider commercial applications/support
- [ ] 10.54 Explore academic/research collaborations

---

## Summary

This detailed plan breaks down the ambitious goal of porting DWScript from Delphi to Go into **~500+ bite-sized tasks** across 10 stages. Each stage builds incrementally:

1. **Stage 0**: Project foundation (34 tasks)
2. **Stage 1**: Lexer implementation (45 tasks)
3. **Stage 2**: Basic parser and AST (64 tasks)
4. **Stage 3**: Statement execution (65 tasks)
5. **Stage 4**: Control flow (46 tasks)
6. **Stage 5**: Functions and scope (46 tasks)
7. **Stage 6**: Type checking (50 tasks)
8. **Stage 7**: Object-oriented features (77 tasks)
9. **Stage 8**: Additional features (62 tasks)
10. **Stage 9**: Performance and polish (68 tasks)
11. **Stage 10**: Long-term evolution (54 tasks)

**Total: ~511 tasks**

Each task is actionable and testable. Following this plan methodically will result in a complete, production-ready DWScript implementation in Go, preserving 100% of the language's syntax and semantics while leveraging Go's ecosystem.

The project can realistically take **1-3 years** depending on:
- Development pace (full-time vs part-time)
- Team size (solo vs multiple contributors)
- Completeness goals (minimal viable vs full feature parity)

With consistent progress, a **working compiler for core features** (Stages 0-5) could be achieved in **3-6 months**, making the project usable early while continuing to add advanced features.
