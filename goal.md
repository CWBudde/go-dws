# Plan for Porting DWScript from Delphi to Go

## Overview and Strategy

DWScript is a full-featured Object Pascal-based scripting language with object-oriented features, strong static typing, and many extensions[\[1\]](https://www.delphitools.info/dwscript/#:~:text=DWScript%20is%20an%20object,An%20experimental%20JIT)[\[2\]](https://www.delphitools.info/dwscript/#:~:text=,programming%20%2A%20generalized%20helpers). Porting its compiler from Delphi to Go is a large undertaking, so we need to break it into incremental milestones. Each milestone will produce a **testable component or subset** of the compiler, ensuring we have working functionality at each step (avoiding the risk of "nothing works until everything is done"). We will preserve 100% of DWScript's syntax and semantics, even if some Delphi features (e.g. class-based OOP, function overloading) are not native to Go - we'll simulate those in the engine where needed. The end goal is a Go library that can compile/execute DWScript code, plus a convenient CLI (using Cobra) to run scripts for testing. We'll use idiomatic Go designs (packages, interfaces, structs, etc.) while mirroring the original compiler's logic closely. Crucially, **each stage will include thorough unit tests** to verify correctness and prevent regressions (mirroring DWScript's own extensive test suite[\[3\]](https://www.delphitools.info/dwscript/#:~:text=DWScript%20is%20an%20object,but%20also%20supports%20syntax%20and)). Below is the step-by-step plan:

## Stage 0: Project Setup and Familiarization

- **Study the DWScript source:** Begin by reviewing the DWScript Delphi code structure to identify its main components: the tokenizer (lexical scanner), the parser/AST builder, the semantic analyzer (type checker), the runtime or code generation, and any supporting utilities (e.g. symbol tables, memory manager). Understanding the original architecture will inform how to design the Go version.
- **Set up Go project:** Initialize a new Go module (e.g. github.com/CWBudde/dwscript-go). Plan a package structure to keep the project organized and idiomatic:
- For example, packages like lexer, parser, ast, interp (interpreter/runtime), and dws (higher-level API) could be used.
- This separation will allow working on components in isolation with targeted tests.
- **Create a CLI skeleton:** Use Cobra to scaffold a command-line application (e.g. dwscript tool) that uses the library. Initially, the CLI can just accept a script file or code string and print a placeholder (since the compiler isn't ready yet). For instance, implement a run command that reads a file and for now just echoes back or says "not implemented". This will ensure the Cobra framework is in place early.
- **Version control and CI:** Set up a git repository and CI pipeline (if possible) to run tests on each commit. This helps maintain quality as you progress through stages.
- **Define success metrics:** Before coding, define what "basic working compiler" means (e.g. _able to parse and execute simple scripts with variables, expressions, control flow, and functions_). This will be the milestone for the first complete iteration, with advanced features (classes, etc.) to follow.

## Stage 1: Implement the Lexer (Tokenization)

A lexical analyzer (tokenizer) reads the source text and produces a stream of tokens that the parser will consume. We'll implement this first since it's a self-contained component that can be tested independently.

- **Define token types:** Enumerate all token categories in DWScript's syntax. This includes keywords (begin, end, if, var, etc.), symbols (operators like + - \* /, punctuation like ; , ()), literals (strings, numbers), identifiers, comments, and so on. Using the DWScript language reference or source, make a comprehensive list of tokens up front.
- **Write the lexer logic:** Create a lexer.Lexer struct with an input (the script source string) and an API to produce the next token. This can be done via a method like NextToken() Token. Implement it in idiomatic Go (for example, using a simple state machine or switch over current rune). Ensure it handles:
- Skipping whitespace and comments (DWScript uses {...} or (\* ... \*) for comments, similar to Delphi).
- Recognizing multi-character operators (e.g. :=, >=, &lt;=, <&gt; for not-equal, etc.).
- Handling string literals (pay attention to quotes and escape sequences) and numeric literals (integers, floats, possibly hex if supported).
- Identifying identifiers and keywords (you can maintain a map of reserved words to token types for quick lookup after reading an identifier).
- **Testing the lexer:** For this stage, write extensive **unit tests** to validate the tokenizer. Craft sample input strings covering various cases:
- Basic sequences (e.g. var x := 42; should produce the tokens \[VAR\]\[IDENT "x"\]\[ASSIGN :=\]\[NUMBER 42\]\[SEMICOLON\]).
- Edge cases: strings with escaped quotes, comment blocks, unusual numeric formats, etc.
- Full-line examples from DWScript code (maybe take a snippet of a known simple DWScript program and lex it to ensure all tokens come out in order).
- **Iterate and refine:** Run tests and fix any issues (off-by-one errors in reading characters, wrong token for a given keyword, etc.). By the end of this stage, we should have a reliable lexer package that turns source text into a token stream. This forms the foundation for parsing.

## Stage 2: Build a Minimal Parser and AST (Expressions Only)

With tokenization in place, the next step is to start building the parser and Abstract Syntax Tree (AST) structures. To keep things incremental, begin with a **minimal subset of the language** - for example, just arithmetic expressions and perhaps a simple print statement - so that we can parse and evaluate something end-to-end early on.

- **Define AST node types:** In a new ast package (or within parser package), define Go structs or interfaces for the core AST nodes. For instance:
- ASTNode interface (or base struct) to represent any syntax node.
- Specific node types: NumberLiteral, StringLiteral, Identifier, BinaryExpr (for infix operations like X + Y), UnaryExpr, etc. Also define Statement nodes if needed, like ExpressionStmt (an expression used as a statement) or a placeholder PrintStmt if implementing a print for testing.
- Designing the AST to closely mirror DWScript's structure will help keep the translation accurate. For now, focus on expressions; you can add more node types later for statements, functions, classes, etc.
- **Implement a recursive-descent parser for expressions:** Create a parser.Parser that uses the Lexer to consume tokens and build an AST. Implement expression grammar (likely following operator precedence rules for Pascal):
- e.g. rules for factors (numbers, strings, identifiers, parentheses), terms (multiplication/division), expressions (addition/subtraction), etc., including unary operators (like - negation, not if Boolean, etc.).
- This should handle basic math and possibly string concatenation (+ might be used for strings in DWScript, given Pascal heritage).
- Also handle operator precedence and associativity correctly (you might model this on how Delphi handles it, or use a precedence table approach).
- **Parse minimal statements:** To allow testing a full line, implement parsing of a simple statement like an expression statement or an assignment. For example, support Identifier := Expression; as an assignment statement node. Also, if the DWScript language has a PrintLn built-in function (as seen in examples[\[4\]](https://www.delphitools.info/dwscript/#:~:text=type%20THelloWorld%20%3D%20class%20procedure,Hello%2C%20World%21%27%29%3B%20end%20end)), we could parse a function call as an expression/statement to use as output in tests.
- **Testing the parser (expressions):** Write unit tests for the parser that feed in tokens (or raw text combined with the lexer) and check that the AST produced is correct:
- For example, parsing 3 + 5 \* 2 should produce an AST with + at the root and 3 and (\*) subtree as children (with 5 and 2). We can verify the structure by traversing the AST or implementing a .String() method for AST nodes to get a normalized representation (like (3 + (5 \* 2))) for comparison in tests.
- Test assignment parsing: input like x := 10; should produce an Assignment node with Identifier("x") and a NumberLiteral(10).
- Test a simple built-in call if implemented: e.g. PrintLn('hi'); might produce a CallExpr node with callee "PrintLn" and argument "hi".
- **Integrate with CLI for a quick demo:** At this stage, although the language coverage is tiny, you can integrate with the CLI to demonstrate progress. For example, make the run command use the Lexer and Parser on an input string and then maybe print the AST or evaluate the expression. This could be as simple as evaluating an arithmetic expression and printing the result. This gives an early end-to-end test:
- e.g. running dwscript -e "3+5\*2" (if you add a flag to evaluate an expression) could print 13 as a proof of concept that lexing, parsing, and evaluating works on a basic level.
- This stage ensures the parser's foundation is laid and we have the ability to handle expressions, which will be part of all larger constructs. We also gain confidence that the whole pipeline (tokenize -> parse -> output) can work on a small scale.

## Stage 3: Parse and Execute Simple Statements (Sequential Execution)

Now we expand the supported syntax to be able to handle a sequence of statements (a basic script) and execute them. The idea is to achieve a **small but functioning interpreter** for a subset of DWScript (without control flow yet), so we can run simple scripts end-to-end.

- **Expand AST for statements:** Introduce node types for statements such as:
- VarDecl for variable declarations (e.g. parsing the var x: Integer; syntax - initially, we might skip type annotations or store them for type-checking).
- AssignmentStmt (already partly done with the assignment expression in Stage 2).
- BlockStmt or CompoundStmt to represent a series of statements (like the contents of a begin ... end block, or the entire script).
- Possibly CallStmt if function calls can be top-level statements (like calling a procedure).
- **Implement parsing of multiple statements and blocks:** Allow the parser to handle a sequence of statements terminated by semicolons. Likely, DWScript syntax is similar to Delphi:
- It might allow standalone statements at global scope (and begin...end blocks for compound statements). Implement parsing such that it can read a list of statements until EOF or a closing end.
- Parse variable declarations (var blocks) to introduce new variables. For now, record their names and optionally types in a symbol table or context (to be used later for type checking and execution).
- If not already, implement parsing of simple procedure calls or built-in procedures (like PrintLn()) as statements so we can perform output in scripts.
- **Create a basic runtime environment:** Build a minimal **execution engine** (an interpreter) for the subset of the language we have:
- Manage a symbol table or environment map for variable values. You can use a Go map (e.g. map\[string\]Value) to store variable names to their current values. Define a Value type (could be an interface{} or a custom union type for int/float/string, etc.). Initially, you can make it dynamic (e.g. store everything as interface{}) to get things running, and enforce types later.
- Implement evaluation of statements in order: e.g. for an AssignmentStmt, evaluate the right-hand expression and store it in the variables map; for a VarDecl, allocate an initial value (default zero or nil) for that variable; for a CallStmt like PrintLn, call the corresponding Go function to output the argument.
- You might want to represent this as an Interpreter struct with methods like ExecStmt(ASTNode) or ExecBlock(\[\]ASTNode) to walk the AST and perform actions.
- **Testing sequential execution:** Write tests where you feed a small multi-line script into the system and verify the final state or output:
- e.g. Script: var x: Integer; x := 5; x := x + 3; PrintLn(x); should result in output 8 and perhaps you can also inspect that the interpreter's variable map has x = 8.
- Test that multiple statements execute in order and affect the environment correctly.
- If an undeclared variable is used or other error, decide how to handle it: for now, you might allow dynamic creation (treat it as 0?) or better, have the parser/semantic stage catch it. Likely better to catch as an error - you can add a simple check: when assigning or referencing a variable, ensure it's in the symbol table (declared). For now, a runtime error or assertion in tests is fine; we will improve error handling and static checks later.
- **Interactive testing:** At this point, you can actually run simple scripts via the CLI. Enhance the Cobra run command to use the parser and interpreter to execute a file or code string:
- For example, running a file that contains a few variable assignments and a PrintLn should now produce real output on the console.
- This is a major milestone: a basic DWScript interpreter in Go for simple programs.

## Stage 4: Control Flow - Conditions and Loops

With basic sequential execution working, the next step is to introduce control flow constructs (which are essential in any non-trivial program). We will implement parsing and execution of if statements, case (if needed), loops like while, repeat...until, for loops, etc., as supported by DWScript.

- **Parsing control structures:** Expand the parser to handle DWScript's control-flow grammar:
- **If/Else:** Parse an if &lt;expr&gt; then &lt;stmt&gt; \[else &lt;stmt&gt;\]. The AST node could be IfStmt containing the condition expression and the then-branch (and optional else-branch statement or block).
- **Case statements:** DWScript likely supports Pascal-style case for enums/integers. If so, parse into a CaseStmt node with the controlling expression and a list of case branches (each branch having a constant value list and associated statement). This can be complex; you might choose to postpone case until after simpler loops, or implement a basic version.
- **Loops:** Handle while &lt;expr&gt; do &lt;stmt&gt;, repeat &lt;stmt(s)&gt; until &lt;expr&gt;, and for loops (for i := 1 to N do ...). Each of these becomes an AST node (WhileStmt, RepeatStmt, ForStmt) with the necessary parts (initialization, condition, increment, loop body, etc.). Pay attention to scope rules (e.g. in Pascal, the loop variable in a for loop might be local).
- **Begin/End blocks:** Ensure the parser can handle compound statements (multiple statements enclosed in begin ... end - often used as the body of if or loops).
- **Executing control structures:** Extend the interpreter to support these new AST nodes:
- **IfStmt:** Evaluate the condition; if true, execute the then-branch, otherwise execute the else-branch (if present).
- **While/Repeat:** Use a loop in Go to repeatedly execute the body until the condition fails (or in repeat...until, until condition is true). Take care to prevent infinite loops if the script logic is wrong (perhaps add a safeguard in tests).
- **ForStmt:** Evaluate the start and end values, then loop accordingly, modifying the loop variable each iteration. Make sure to handle the loop variable's scope and final value according to Pascal's rules (in Pascal, the loop variable is often read-only inside the loop and goes out of scope after, but some script engines treat it differently).
- **CaseStmt:** For each case branch, compare the expression value to the branch values and execute the matching branch. Include a default/else branch if the syntax allows (Delphi has an else for case for no matches).
- **Testing control flow:** Write tests for each construct:
- If/else: e.g. if x > 0 then PrintLn('Positive') else PrintLn('Non-positive'); - test both branches by varying x.
- While loop: e.g. a loop that sums numbers; verify the result.
- Repeat...until and for loops similarly. For loops can be tested with ascending and (if supported) descending direction (downto in Pascal).
- Case: test a case with a few branches.
- Also test nested control structures (an if inside a loop, etc.) to ensure the AST and interpreter handle nesting properly.
- Now the interpreter can handle most imperative logic. At this stage, we have a **functioning mini-compiler** for a significant subset of the language: variables, expressions, assignments, and control flow. This is enough to write non-trivial scripts, except that all code must be in one block (we haven't added functions yet). The CLI can now run more complex script files (you can try out some RosettaCode examples for Pascal to see if they work, adjusting syntax if needed).

## Stage 5: Functions, Procedures, and Scope Management

Functions and procedures (subroutines) are critical for structure and reusability in scripts. DWScript being a Pascal-like language supports procedure/function definitions, likely with their own local variables, parameters, and return values (functions). This stage will implement user-defined functions and the necessary infrastructure (call stack, scoped symbol tables, etc.) in the Go port.

- **Parse function/procedure declarations:** Extend the parser to recognize DWScript's way of defining functions. For example, in Pascal it could be:
- function Foo(a: Integer): String; begin ... end;
- procedure Bar; begin ... end;
- There may also be nested functions or just global ones; handle accordingly.
- Create AST node types like FunctionDecl (with fields for name, parameters, return type, and body (a Block of statements)). Also a Param struct for parameters (name and type).
- Store function declarations in a symbol table of sorts, so that calls to them can be resolved. Possibly maintain a map functions\[name\] = FunctionDecl node in the parser or a separate semantic phase.
- **Parse function calls (expressions):** We likely already handle built-in calls, but now ensure the parser can parse user-defined function calls as an expression (if it returns a value) or as a statement (if procedure). You might have a CallExpr AST node with a name (or a reference to the function symbol) and a list of argument expressions.
- **Manage scopes and symbol tables:** Introduce a structure to represent _environments_ or symbol tables for variables at runtime:
- When entering a function, create a new local symbol table (possibly chain it to a global table for global vars). Each function invocation will have its own frame with local variables and parameters.
- The parser or a semantic analysis step should also manage scopes for static checking: e.g. a local variable named x in a function should not conflict with a global x. We can maintain a stack of scopes during parsing or separate semantic pass to validate and set up symbol table entries.
- You may design a Symbol struct for each variable/function with type info, and a SymbolTable structure with methods to enter scope, leave scope, define and lookup symbols.
- **Interpreter: call stack and function execution:** Extend the interpreter to handle function calls and returns:
- Maintain a call stack. This can be implicit via recursion in Go (calling an ExecBlock for the function body), but it's good to conceptualize it as a stack of activation records. An activation record contains the local variable map (and perhaps a reference to the function's AST or return address).
- When a CallExpr/CallStmt is encountered:
  - Look up the function definition (from a global registry of functions parsed).
  - Evaluate the argument expressions.
  - Create a new environment for the function: map parameters to argument values, set up local vars (if any initializations).
  - Execute the function's body statements (you likely need to detect a return value for functions - DWScript might use the function name as a return variable or an explicit Result variable; Delphi uses a Result keyword or function name assignment).
  - When done, capture the return value (if any) and pop the environment back to the caller.
- Implement a mechanism for returning from function early if Exit or similar is used (could be by panicking internally and catching, or checking a flag).
- **Testing functions:** Write various tests:
- A simple function that returns a calculated value. For example, function Add(a, b: Integer): Integer; begin Result := a+b; end; PrintLn(Add(2,3)); should output 5.
- Test recursion: a recursive factorial or Fibonacci to ensure the call stack works and there's no leftover state between calls.
- Test a procedure (no return) that perhaps modifies a global variable or has side effects.
- Test passing arguments by value (DWScript likely only has value semantics, unless it supports var parameters for by-reference - which it might, being Pascal-derived. If so, that's another feature to consider handling).
- Ensure that local variables in one function do not interfere with others (scope isolation works).
- At this stage, we have a **basically working compiler/interpreter** for much of the core language: you can define variables, write functions, use loops and ifs, and so on. This meets the initial goal of a "basically working compiler" capable of running meaningful scripts. The CLI tool can now load a DWScript file and execute it fully. It's a good point to possibly release an alpha version of the library for feedback, as all further features are extensions on this foundation.

## Stage 6: Static Type Checking and Semantic Analysis

Up to now, we may have been lenient on type checking (perhaps even using dynamic types for convenience). Since DWScript is strongly typed, we should incorporate a semantic analysis phase that enforces the type rules at **compile time** (when scripts are compiled to AST, before execution). This will improve correctness and bring us closer to DWScript's true behavior.

- **Type system setup:** Define a representation for types in the language. You can create a Type interface or enum for base types (Integer, Float, Boolean, String, etc.) and structures to represent complex types (class types, record types, array types, interface types, etc.). For now, focus on primitive types and maybe arrays.
- You might model this with an enum for simple types and struct types for compound ones. Provide a way to compare types for equality (needed for checking assignments, etc.).
- **Attach types to AST nodes:** Modify AST node definitions to carry type information where relevant:
- Each expression node can have a field Type Type. For literals, that's known (number literal => Integer or Float type; string literal => String type; boolean literal => Boolean).
- For variable references, the type is whatever the variable was declared as.
- For binary operations, determine result types (e.g. Integer + Integer -> Integer; if mixed int/float, perhaps promote to float as Delphi does; string concatenation results in String, etc.). If an operation is not defined on given types (e.g. adding a number and a string without an overload), that's a compile error.
- For assignment, ensure the right-hand expression type is compatible with the variable's type.
- For function calls, check that the number and types of arguments match the function's parameters. Also assign a type to the CallExpr node equal to the function's return type (for usage in larger expressions).
- **Semantic analysis pass:** Implement a traversal of the AST (after parsing) to resolve and check all symbols and types:
- Ensure every identifier is declared (if not, report an undefined variable error at that point - this may involve storing errors or returning an error up the call stack).
- Check that types match in assignments, arguments, return statements, etc. Where possible, also check range or other constraints (for instance, if there's an in operator for sets, type check that the left side is of the set's element type).
- Insert any implicit conversions if DWScript allows (Delphi sometimes auto-converts integer to float in expressions, etc. If DWScript has similar rules, you might incorporate that).
- For control flow, ensure conditions are booleans, for-loops have an ordinal type for the loop var, etc.
- If the language supports forward declarations or external linking of functions, handle resolution of those as well.
- **Testing type checking:** Create tests that deliberately violate type rules to see if the compiler catches them:
- Assigning wrong types: e.g. var i: Integer; i := 'hello'; should produce a type error.
- Calling a function with wrong arg count or types should be an error.
- If possible, test some subtle cases like overflows or incompatible array element assignments, etc.
- Also test that correct code still passes. Ensure that the semantic analyzer doesn't wrongly reject valid constructs (e.g. type inference of var x := 5; results in x being Integer type, etc., if you implement type inference).
- **Error handling design:** Decide how to handle errors in the compiler. You might accumulate errors in a list and return them up, or panic/exception for fatal ones. It's user-friendly to gather multiple errors in one compile attempt if possible. Make sure the CLI can report compile errors clearly with line numbers (this implies tracking line/column positions in tokens and carrying that into AST nodes for error messages).
- By the end of this stage, your compiler is not only _functionally_ correct but also _semantically_ robust, catching mistakes in the scripts just like the original DWScript would. This matches DWScript's strong typing and compile-time checks[\[5\]](https://www.delphitools.info/dwscript/#:~:text=,type%20inference). The engine is now quite solid for real usage of the core language.

## Stage 7: Support Object-Oriented Features (Classes, Interfaces, Methods)

DWScript heavily supports object-oriented programming (classes with inheritance, interfaces, polymorphism, etc.[\[6\]](https://www.delphitools.info/dwscript/#:~:text=name%2C%20it%20is%20general%20purpose,available%20for%20Win32%20%26%20Win64)[\[2\]](https://www.delphitools.info/dwscript/#:~:text=,programming%20%2A%20generalized%20helpers)). This stage will be the most involved, as Go doesn't natively support class-based inheritance or method overloading. We will emulate these features within our engine to preserve the DWScript language semantics.

- **Class declarations parsing:** Extend the parser to handle class type definitions. In DWScript/Delphi, class syntax looks like:
- type  
    TPerson = class(TObject)  
    Name: String;  
    Age: Integer;  
    procedure SayHello;  
    end;
- We need to parse class types, including their inheritance (optional parent class), fields, methods (which themselves have bodies, possibly defined later or inline).
- Represent a class in the AST or symbol table as a ClassDecl (with name, a list of members, parent class reference, etc.).
- Each field can be treated like a variable (with a type) that will belong to instances.
- Each method is essentially a function tied to the class; you can reuse the FunctionDecl structure but mark it as a method of a class (and maybe store the class type as an attribute).
- Parse method implementations. In Delphi, methods can be declared in the class and implemented later (outside the class block). DWScript might allow inline method bodies (the example in docs shows an inline method body within the class[\[4\]](https://www.delphitools.info/dwscript/#:~:text=type%20THelloWorld%20%3D%20class%20procedure,Hello%2C%20World%21%27%29%3B%20end%20end)). Support at least inline implementation for now. If separate implementation is allowed, you might postpone that or treat it similarly to a forward-declared function that gets defined later.
- Also parse constructors, destructors if present (these are just special-named methods).
- **Note:** Even if Go doesn't have classes, we are just building an AST representation here; actual execution will come later.
- **Object instantiation and member access:** Parse the new keyword or constructor calls. E.g., var p := new TPerson; or TPerson.Create() depending on syntax. Represent object instantiation as an AST node (NewExpr or a call to a constructor method).
- Also allow member field access syntax in expressions (obj.Name) and method calls on objects (obj.SayHello()).
- These will likely be parsed as a special kind of expression: you might create a MemberAccessExpr (with fields for the object expression and the field/method name) and a MethodCallExpr similar to CallExpr but with a target object.
- **Runtime representation of objects:** Decide how to represent class instances and classes at runtime in Go:
- Since Go lacks built-in class inheritance, use an **object struct or map**: For instance, each script object could be represented as map\[string\]Value for fields plus a pointer to a class metadata. Or define a struct ObjectInstance with a reference to ClassInfo and a map of field names to values.
- ClassInfo (or Class metadata) can hold information about the class: its name, parent class (if any), a list of field names/types, list of methods (perhaps pointers to their AST or to function closures implementing them), and maybe a method table for quick lookup of methods (to handle overriding).
- When an object is created (via new), allocate an ObjectInstance: initialize the fields map (with default zero values or nils), store a reference to its ClassInfo. If the class has a parent, you may need to initialize parent fields as well (though if using a map this is implicit if you include inherited fields in the map keys).
- For method calls, we'll need to resolve the method by name, possibly following the class's inheritance chain. You could store in ClassInfo a map of method name to function (with the child class's methods overriding parent's).
- Overloading: If DWScript allows method overloading (same name, different params), then the method resolution also needs to consider parameter types count. This complicates lookup: possibly store methods as name+signature keys, or store a list of methods by name and decide which to call based on argument count/types at call time. (In this initial implementation, you might choose to forbid or not implement overloading to simplify, since Go can't do it at compile time - but you can simulate it by checking arg types in the interpreter and calling the matching one.)
- **Executing methods and field access:** Extend the interpreter to support class operations:
- **Field access:** When evaluating obj.Name, retrieve the Value from the object's field map. Also, ensure type correctness (e.g. the field exists and the object is of the right class type).
- **Method call:** When executing obj.SayHello(), the interpreter should:
  - Find the ClassInfo for obj. Look up SayHello in its method table (which might be a pointer to the AST/function for that method).
  - Set up a call similarly to a normal function, but with an implicit **Self parameter** for the object instance. In Delphi, inside a method, you refer to the instance as Self. We can mimic that by always passing the object as a hidden first argument or by storing it in a special variable accessible during method execution.
  - Execute the method's AST (which is essentially a function body) in a new environment. In that environment, resolve references to fields or other methods via Self. For example, if the script's method code accesses Name or calls another method, the interpreter should know it refers to the current object. One way: when entering a method, push the ObjectInstance as the current Self (perhaps an entry in the symbol table or a special global). Another way: capture Self as a context pointer in a closure representing the method, if you pre-bind methods.
  - Handle method return values (if not a procedure). It works like function returns.
- **Inheritance and polymorphism:** If class B inherits A and overrides a method, ensure the method table in B's ClassInfo points to B's implementation. So when calling a method on an object of class B stored in a variable of type A (if type-check allows that), the interpreter should still call B's override - this is dynamic dispatch. With our method table approach, as long as the object's ClassInfo is B, it will find B's method.
- **Interfaces:** If DWScript supports interfaces (it does[\[7\]](https://www.delphitools.info/dwscript/#:~:text=tests%20suite%20ensures%20high%20quality,available%20for%20Win32%20%26%20Win64)), implementing them fully might be complex. Possibly skip initially or treat them similar to abstract classes with only methods. You could mark in ClassInfo which interfaces it implements and ensure those method signatures exist. But unless needed immediately, this could be a later enhancement once classes work.
- **Testing OOP features:** This requires comprehensive tests:
- Class instantiation and field usage: e.g. define a class with some fields, create it, assign to fields, and then print them or compute something to verify fields hold values.
- Method invocation: define methods that do something visible (like modify a field or print output). Test calling them works.
- Inheritance: class B inherits A, overrides a method. If you call that method on an instance of B (even if typed as A), B's version should run (test for correct polymorphic dispatch).
- Method overloading (if implemented): define two methods same name different params, call each and ensure the correct one is invoked. (If choosing not to implement overloading at first, document that or have the compiler treat it as an error to avoid confusion.)
- Interface (if implemented): have a class implement an interface, and ensure an interface variable can hold the object and calling interface methods calls the actual class's method.
- **Go idiomatic adjustments:** Although Go doesn't have classes, our implementation is essentially a mini OOP system on top. Ensure the design is idiomatic in terms of code structure:
- Use structs to represent ClassInfo and ObjectInstance rather than overly generic maps if possible, to make code clearer (the fields map inside ObjectInstance can remain a map for flexibility).
- Use methods on those structs for operations (e.g. func (obj \*ObjectInstance) GetField(name string) Value).
- Use interface types in Go if it helps (for example, you might have Value as an interface and make ObjectInstance implement it so it can be a Value).
- This stage is possibly the most challenging, but once completed, we will have **full class support** akin to DWScript: the script can define classes with inheritance, create objects, and call methods, achieving parity with Delphi's object model (aside from things like Delphi's RTTI, which we can skip or handle later). We will have to carefully test and perhaps refine memory management (though Go's GC will handle garbage collection of objects as long as we don't create reference cycles outside its scope).

## Stage 8: Additional DWScript Features and Polishing

After implementing the core language including OOP, there may remain several DWScript-specific features and miscellaneous enhancements. This stage involves adding those and polishing the project for completeness:

- **Operator overloading:** DWScript allows operator overloading[\[8\]](https://www.delphitools.info/dwscript/#:~:text=,compound%20assignment%20operators), meaning classes or records can define how +, -, etc., work for them. Implementing this requires:
- Recognizing operator overload definitions in class/record declarations (if DWScript's syntax allows, e.g. class operator Add(a, b: TMyClass): TMyClass; begin ... end; or something similar).
- Storing these as special functions associated with a type and operator symbol.
- During expression parsing/semantic analysis, if an operator is used on user-defined types, resolve to the overloaded operator function instead of a built-in operation. Possibly create a pseudo-call in the AST to that operator function.
- At runtime, ensure that when evaluating such an expression, it calls the overload function.
- This is complex and can be deferred if not immediately needed, but keep it in plan for completeness.
- **Properties and property expressions:** If DWScript supports properties (like Delphi property X: Integer read FX write SetX), decide how to handle them. You might translate property access to method calls or field access in AST during semantic phase. This might be an advanced feature to add after the main OOP is working.
- **Set types, enumerated types, records:** DWScript likely supports sets and records as per Pascal tradition.
- **Records**: simpler than classes (no methods, value types). You can treat them similarly to classes but without inheritance. Implement parsing of record types and usage (accessing record fields).
- **Sets**: e.g. set of &lt;enum&gt; or set of Byte. Implement as a bitset or boolean map internally. Parsing could treat set literals like \[val1, val2\] and support the in operator. Execution would involve representing sets perhaps as a map or a Go bitset for small ranges.
- **Enums**: If present, parse enum type declarations and treat them as special constants.
- Add these as needed, with tests for each (like set membership, record field behavior).
- **Meta-class (class reference) support:** DWScript has "full support for meta-classes"[\[9\]](https://www.delphitools.info/dwscript/#:~:text=,support%20function%20%26%20methods%20pointers), meaning you can have a variable that is a class type itself (like TClass = class of TObject in Delphi). Supporting this means:
- Allowing class types as first-class values (so a variable can hold a reference to a ClassInfo).
- Possibly allowing calling static methods or constructors via that reference. This is advanced; implement if necessary by treating class references akin to an object that has no instance (just metadata).
- If not critical, this can be postponed.
- **Function pointers / delegates:** DWScript supports function and method pointers[\[10\]](https://www.delphitools.info/dwscript/#:~:text=,programming). To implement, you can allow taking the address of a function (creating a closure or a reference that can be stored in a variable and later called). In the interpreter, this could be represented by storing a reference to the function's AST or a wrapper that calls it. This is a nice-to-have feature for completeness.
- **Contracts (Design by Contract):** If the language has require/ensure clauses in functions (as "contracts-programming" suggests[\[11\]](https://www.delphitools.info/dwscript/#:~:text=%2A%20full%20support%20for%20meta,programming)), you can support by parsing those and executing checks at runtime (throwing an error if contract fails). This is optional and can be added later in the function execution logic.
- **Other extensions:** Partial classes, inline assembly, COM integration, JavaScript output, etc., are listed in DWScript's features[\[12\]](https://www.delphitools.info/dwscript/#:~:text=,etc). These can be considered out of scope for the initial port:
- **Partial classes** (splitting class definition across multiple places) might not be needed unless supporting multi-unit compilation.
- **Inline assembly** and COM/RTTI connectors are very platform-specific and can be ignored in Go implementation.
- **JavaScript codegen** is a separate goal (compiling DWScript to JS); not needed for a basic working compiler, but after everything, one could consider adding an alternative backend to output JavaScript instead of executing (this could reuse the AST and generate JS code as a string).
- **Testing & validation:** For each added feature, expand your test suite. Also, consider running the original DWScript's unit tests (if available publicly) against your implementation:
- You could write DWScript test scripts (from their suite or blog examples) and run them with your Go engine to compare outcomes.
- This is a great way to catch any semantic differences or missing features.
- Pay special attention to edge cases in advanced features (like calling an overloaded operator on null object, or interface casting).
- By end of this stage, the goal is to have **feature parity with DWScript** in terms of language capabilities, modulo any platform-specific bits. The compiler library should be robust and cover essentially all syntax that DWScript advertises.

## Stage 9: Performance Tuning and Refactoring

With correctness and completeness achieved, we can look at improving performance and code maintainability:

- **Optimize the interpreter (or add a bytecode compiler):** The AST-walking interpreter, while simple, might be slow for heavy scripts. Consider translating the AST into a bytecode and implementing a bytecode VM in Go for faster execution. This could significantly speed up loops and function calls:
- Design an instruction set for common operations (load variable, set variable, binary op, call function, etc.).
- Write a compiler that traverses AST and emits bytecode instructions.
- Implement a VM loop to execute the bytecode. Go is quite fast at running loops, so this could be efficient. However, care with Go interface{} and type assertions in VM to keep it optimized.
- This is a complex project by itself, so gauge if performance is a concern. The experimental JIT in DWScript is an analog - we may or may not need that depending on use cases.
- **Memory management considerations:** DWScript has automatic memory management[\[13\]](https://www.delphitools.info/dwscript/#:~:text=,strong%20typing). In our implementation, Go's garbage collector handles most of it. Just ensure that we don't hold onto objects longer than needed (e.g. if we implement a global pool, etc.). If we simulate reference counting (Delphi might use reference counts for interfaces/strings), we can ignore that because Go GC does the job. Just double-check that cyclic structures (object graphs) pose no issue - in Go they don't, the GC collects them.
- **Refactor and clean up:** As the codebase grew with new features, some earlier design decisions might need revisiting:
- Maybe unify the representation of values now that we have many types (perhaps introduce an interface Value with methods or a single struct with kind and union).
- Ensure error handling is consistent (the parser vs semantic vs runtime errors are reported in a uniform way).
- Review performance of critical sections (hot loops in interpreter, large nested ASTs in recursion, etc.) and optimize (e.g. use slices instead of maps where applicable, preallocate lists, etc.).
- Apply Go best practices (run golint, go vet, etc.) to fix any non-idiomatic patterns that slipped in when mirroring Delphi logic.
- **Documentation and examples:** Write clear documentation for the Go library (public GoDoc comments) and for the CLI usage. Provide examples of how to use the library to embed the DWScript engine in a Go application, and how to use the CLI to run scripts. This will help others (and your future self) understand the design.
- **Increase test coverage:** By now you should have a broad test suite. Consider adding property-based tests or fuzz testing for the parser (generate random but valid syntax to stress-test the parser's error recovery and ensure no panics). Given the complexity, a fuzzer might find edge cases in parsing or execution.
- At the end of this stage, you should have a **production-ready DWScript compiler in Go**: it's complete, tested, documented, and reasonably efficient.

## Stage 10: Long-Term Evolution

This final "stage" is more about the project's future beyond the initial port:

- **Feature parity verification:** Continuously track DWScript's upstream changes (if any, since it's actively maintained[\[14\]](https://github.com/EricGrange/DWScript#:~:text=Releases%208)) to see if new language features or fixes should be ported. Plan updates accordingly to keep parity with DWScript's syntax and behavior.
- **Potential enhancements:** Since the project is now in Go, consider leveraging Go's ecosystem:
- Perhaps offer a WASM compilation of the engine for running DWScript in web browsers (taking advantage of Go's WebAssembly target).
- Provide an interactive REPL mode (read-eval-print loop) using the interpreter, which could be handy for testing or educational use.
- If performance is critical, explore integrating Go's unsafe or syscall to implement a JIT (though Go is not great for self-modifying code generation, you might skip JIT and rely on bytecode VM efficiency).
- **Community and testing:** Encourage usage by others to get feedback. Use the CLI to run DWScript example scripts available on Rosetta Code[\[15\]](https://www.delphitools.info/dwscript/#:~:text=Much%20of%20your%20Pascal%20or,report%20more%20specific%C2%A0issues%20and%20suggestions) or elsewhere, to validate behavior against expected outputs (these can serve as high-level integration tests).
- **Maintainability:** Since this is a large project, ensure there's a clear CONTRIBUTING guide if open source, and modularize the code so future contributors (or yourself) can easily modify one part (e.g. parser) without breaking others. High test coverage will support this.
- **Edge-case audit:** DWScript likely has some tricky corners (like short-circuit boolean evaluation, operator precedence quirks, etc.). Do a sweep to verify those against the original (for example, confirm that boolean and/or are short-circuited, division by zero handling, etc., match DWScript's behavior). Add tests or fixes for any discrepancies.

Throughout all these stages, we emphasize **incremental development and testing**. By always having a working subset of the compiler, you reduce risk and can validate each component thoroughly. Given that DWScript itself has a comprehensive test suite[\[16\]](https://www.delphitools.info/dwscript/#:~:text=DWScript%20is%20an%20object,class%20and%20interfaces%20support), you can aim to achieve similar robustness. Each stage builds upon the last, and by the time all are complete, you will have a Go-based DWScript compiler that is close to the original in functionality yet implemented in an idiomatic, maintainable Go style.

**Sources:**

- Eric Grange, _DelphiWebScript Project README_ - Overview of DWScript's purpose and features[\[17\]](https://github.com/EricGrange/DWScript#:~:text=DWScript%20is%20an%20object,of%20its%20own%20as%20well)[\[1\]](https://www.delphitools.info/dwscript/#:~:text=DWScript%20is%20an%20object,An%20experimental%20JIT).
- _DelphiTools - DWScript Overview_ - List of supported language features in DWScript (OOP, strong typing, extensions, etc.)[\[2\]](https://www.delphitools.info/dwscript/#:~:text=,programming%20%2A%20generalized%20helpers).

[\[1\]](https://www.delphitools.info/dwscript/#:~:text=DWScript%20is%20an%20object,An%20experimental%20JIT) [\[2\]](https://www.delphitools.info/dwscript/#:~:text=,programming%20%2A%20generalized%20helpers) [\[3\]](https://www.delphitools.info/dwscript/#:~:text=DWScript%20is%20an%20object,but%20also%20supports%20syntax%20and) [\[4\]](https://www.delphitools.info/dwscript/#:~:text=type%20THelloWorld%20%3D%20class%20procedure,Hello%2C%20World%21%27%29%3B%20end%20end) [\[5\]](https://www.delphitools.info/dwscript/#:~:text=,type%20inference) [\[6\]](https://www.delphitools.info/dwscript/#:~:text=name%2C%20it%20is%20general%20purpose,available%20for%20Win32%20%26%20Win64) [\[7\]](https://www.delphitools.info/dwscript/#:~:text=tests%20suite%20ensures%20high%20quality,available%20for%20Win32%20%26%20Win64) [\[8\]](https://www.delphitools.info/dwscript/#:~:text=,compound%20assignment%20operators) [\[9\]](https://www.delphitools.info/dwscript/#:~:text=,support%20function%20%26%20methods%20pointers) [\[10\]](https://www.delphitools.info/dwscript/#:~:text=,programming) [\[11\]](https://www.delphitools.info/dwscript/#:~:text=%2A%20full%20support%20for%20meta,programming) [\[12\]](https://www.delphitools.info/dwscript/#:~:text=,etc) [\[13\]](https://www.delphitools.info/dwscript/#:~:text=,strong%20typing) [\[15\]](https://www.delphitools.info/dwscript/#:~:text=Much%20of%20your%20Pascal%20or,report%20more%20specific%C2%A0issues%20and%20suggestions) [\[16\]](https://www.delphitools.info/dwscript/#:~:text=DWScript%20is%20an%20object,class%20and%20interfaces%20support) DWScript - DelphiTools <https://www.delphitools.info/dwscript/>
[\[14\]](https://github.com/EricGrange/DWScript#:~:text=Releases%208) [\[17\]](https://github.com/EricGrange/DWScript#:~:text=DWScript%20is%20an%20object,of%20its%20own%20as%20well) GitHub - EricGrange/DWScript: Delphi Web Script general purpose scripting engine <https://github.com/EricGrange/DWScript>