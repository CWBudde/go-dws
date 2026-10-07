# Error Message Format Documentation

**Tasks**: 9.110-9.118
**Status**: ✅ COMPLETE
**Date**: 2025-11-04

## Overview

This document describes the error message format and improvements implemented in go-dws. The enhanced error messages provide clear, actionable information to help developers quickly identify and fix issues.

## Error Message Structure

All errors in go-dws follow a consistent format with rich contextual information:

```
[Error Type] in <filename>:<line>:<column>
  <line-number> | <source code line>
                  <position marker>
<detailed error message>
[optional: additional context]
```

### Components

1. **Error Type**: Classification (Error, Runtime Error, etc.)
2. **Location**: Filename, line number, column number
3. **Source Context**: The actual line of code that caused the error
4. **Position Marker**: Visual indicator (`^`) pointing to the error location
5. **Detailed Message**: Clear explanation of what went wrong
6. **Additional Context**: Optional extra information (values, types, suggestions)

## Error Categories

### Routine signatures

Routine type diagnostics retain the declared name, routine kind, and parameter
modifiers. Examples include `class function ClassType: TClass`,
`function IntToHex(Integer, Integer): String`, `destructor Destroy`, and
`procedure Test(const String)`. Parameterless routines omit parentheses; unnamed
types retain the space after their kind, as in `procedure (String)`.

Assignments between incompatible routine pointers report both signatures, target
first, at the supplied expression. Parameter passing modes (`const`, `var`, and
`lazy`) participate in compatibility. An incompatible reference with required
parameters can also report `More arguments expected` before the type mismatch,
matching DWScript's diagnostic recovery.

### Call arguments and duplicate methods

Call argument mismatches report `Argument N expects type "X" instead of "Y"`
at the supplied expression, with zero-based numbering. A procedure call has no
value, so its diagnostic ends after the expected type. An argument with its own
error does not receive an additional type-mismatch diagnostic. Strict parameters
declared with `type` reject implicit conversions, including in class methods and
constructors; ordinary parameters retain expected-type inference and conversions.

Record instance methods and helper instance methods reserve argument zero for their receiver.
Their first written argument is numbered one, and the shifted index selects the
next written argument's position, or the method name when none remains. Methods on classes,
record class methods, constructors, and set `Include`/`Exclude` use ordinary
argument numbering.

Regular named routines reject nonvariables and immutable storage with
`Argument N (name) cannot be passed as Var-parameter`, using the parameter's
declared casing, a zero-based index, and the written argument's position.
Constants, `const` and `lazy` parameters, and object `Self` cannot supply a
writable slot. Record `Self` fields, mutable array elements, JSON members,
helper class variables and field-backed properties retain their storage semantics. A
type mismatch or an error inside the argument suppresses an additional
reference diagnostic; array compatibility recovery still checks storage.

`Inc` and `Dec` use the same reference sentence for their first argument,
with index zero and parameter name `a`. Assignments to object `Self` report
`Cannot assign a value to the left-side argument` at the assignment operator
and continue semantic analysis, preserving surrounding diagnostics and hints.
Broader method/constructor/helper/indirect-call reference checks remain under
audit. Explicit `@Routine` var arguments retain their existing rejection:
upstream permits writable routine-reference data, but Go-dws still needs
runtime support for those temporary reference slots.

A named routine marked `overload` reports
`There is no overloaded version of "X" that can be called with these arguments`
for an argument-count mismatch, even if only one overload is currently visible.
Unmarked routines retain `More arguments expected`, `Too many arguments`, or
`No arguments expected`. A duplicate class-method signature reports
`There is already a method with name "X"` at the completed header: its semicolon
when no directive follows, otherwise the token after the last directive.

Class member calls, including implicit `Self`, named `inherited` calls, and
constructors, use `More arguments expected` or `Too many arguments` for unmarked
signatures. Parameterless class members use `Too many arguments`; specialized
helper and regular-routine policies are separate. A marked class-member or
constructor set uses the overload sentence above, even for a sole marked method.

These class calls read all written arguments before validating types and count.
A supplied parameter's type mismatch suppresses the count error. An error inside
an argument is reported first and can still be followed by the count error, even
when the argument is on a later line. Omitted default parameters remain optional
and execute normally. Constructor dispatch respects an unmarked constructor's
hiding of same-named ancestor constructors. Inherited constructor sets can still
be ambiguous when a synthetic parameterless constructor competes with an
all-optional signature; that candidate-policy issue remains open in PLAN.md.

Count and overload diagnostics anchor at the method or constructor name,
including the name following `inherited`. For `new T(...)`, the anchor is `T`
(or the written metaclass expression). Existing result-type recovery is retained.
Parenthesized native record calls follow the same type-before-count policy,
including implicit `Self` calls and class methods. Unmarked record signatures
use the count sentences; marked overloads use the overload sentence even when
only one declaration exists. These calls read every supplied argument, retain
defaulted parameters, and anchor count/overload errors at the method name.

A record instance method's receiver is argument 0 in DWScript, so its first
written parameter is reported as argument 1. For a type error, DWScript uses
that shifted index to look up the written argument positions: the next written
argument, or the method name when no next argument exists. Record class methods
have no receiver argument and report ordinary indices and argument positions.

Direct parenthesized native interface calls, including inherited methods, use
`More arguments expected` or `Too many arguments` at the method name. Their
receiver stays outside the argument list: argument indices start at 0 and type
errors anchor at the supplied expression. All supplied children are read before
count validation, and supplied type mismatches suppress count errors. The method's
return type remains available for subsequent assignment checks.

Unmarked parenthesized helper calls through a receiver use the same child-before-count
and supplied-type-before-count policy. Count errors anchor at the member name
and parameterless helpers use `Too many arguments`. Instance helpers, including
function helpers, reserve argument 0 for Self; nonstatic class helpers for
records/classes do so too. Their type errors use the shifted position described
above. Static helper methods and class helpers for primitive/interface targets
have no Self argument and use ordinary indices and supplied-expression anchors.
Receiver ownership follows the selected overload, even when instance and static
methods share a name. Implicit Self helper fallback uses the same policy.

Explicit helper-name calls use the same child-before-count and type-before-count
policy. An explicitly supplied Self uses its written argument index and position.
Record class helpers require the matching record metatype, including aliases and
inferred metatype copies; a Variant cannot supply that receiver.

Calls bound to a helper member inside its body use the selected signature's
implicit receiver role, including class/static methods and out-of-line bodies.
Local bindings that shadow the member keep their own call diagnostics.
Inline declaration visibility is retained. A marked helper declaration uses overload
selection even when it is the only visible candidate. Failed receiver, explicit-name,
and bound-body matches report `There is no overloaded version of "X" that can be
called with these arguments` at the member, using its declared name. A found
helper's failed match does not add an inaccessible-member error. Recoverable
child errors precede the enclosing no-match; declaration-case hints and independent
instance-versus-type receiver errors remain visible before ranking.

Successful selection proceeds to ordinary argument checking and execution of that
same declaration and declaring helper. Lazy arguments remain deferred, var arguments
retain their caller's storage, and a statically typed Variant does not choose a new
overload from its contained value. Receiver and ordinary arguments are evaluated once.
Inline helper-body calls retain the declarations visible where the body was written.
For unrelated helpers on the same target, the first declaring helper supplies the
candidate set; a failed match does not try a later helper. Alias-specific lookup
retains precedence. The port's parent-helper extension preserves the complete
inherited overload set and its declaring storage owner when a child has no local
member of that name; a local declaration hides the parent set.

General mixed/inherited candidate collection and ambiguity/Variant ranking remain
under audit; details and bounded runtime acceptance evidence are in the
[helper overload measurement](../architecture/helper-overload-measurement-2026-10.md).

Bare user-helper members are invocations unless a compatible callable type is
expected. Missing required arguments report `More arguments expected` at the
member name while retaining the result type. Compatible references capture the
receiver and helper owner and defer execution; helper-name references include
written Self, while references inside helper bodies bind it implicitly.

Grouping reads the inner helper name as a value before the outer call. A scalar
result reports `Not a method` at the outer opening parenthesis as a compiler stop;
the outer arguments and later diagnostics are suppressed. A callable result is
invoked once by the outer call, with supplied types and children checked before
argument count. A returned callable stored in a variable is retained as a value.

Helper overload selection, helper/interface declaration defaults, bare native
record/interface member reference contexts, and bare inherited-call arity remain
under audit. Imported-helper and metaclass-target
lookup also remain open. Specialized intrinsic helpers retain
their own argument policies.

### Parameter defaults and Swap

A `lazy`, `var`, or `const` parameter with a default reports
`lazy/var/const parameter cannot have a default value` at `=`. Parsing continues
through the initializer and following declarations. The rejected default does
not make the parameter optional. Regular-routine defaults must be constant;
`Constant expression expected` uses the cursor after the initializer, while
incompatible folded scalar constants report their type pair at `=`. Pure calls
and constant aggregates/casts remain valid defaults. An open `array of const`
parameter must use the `const` parameter modifier.

`Swap` checks each argument for writable data and reports `Variable expected`
at the invalid argument. Ordinary variables, array elements, and fields retain
their storage semantics; getter calls and immutable data are rejected. Type
mismatches report `Incompatible types: "X" and "Y"` at `Swap`, with the first
argument's type first. Both types must match, including Variant; invalid
arguments suppress the enclosing pair. A single argument reports `"," expected`
at the closing parenthesis. Explicit `@Print` and `@PrintLn` references expose
DWScript's one-Variant procedure signature.

Assigning to a constant or read-only binding reports
`Cannot assign a value to the left-side argument` at the assignment operator.

### Type-directed punctuation

Ordinary and routine-local record-typed constants require parenthesized field
initializers. Missing `(`
reports `"(" expected` at the first initializer token and stops compilation,
including through record aliases and local constants. Scalar constants and ordinary
calls such as `F(;` retain `Expression expected`. An initializer interrupted after
its opening parenthesis retains the parser's own stop. Retained scalar initializer
fragments do not acquire additional declaration checks.

Unqualified special functions require `(`: `Assert`, `Assigned`, `High`,
`Length`, `Low`, `Ord`, `SizeOf`, `Defined`, `Declared`, `Inc`, `Dec`, `Succ`,
`Pred`, `Include`, `Exclude`, `Swap`, and `ConditionalDefined`. A casing hint
precedes the punctuation stop, anchored at the next token after comments or
whitespace, or at the last real token at EOF. This applies in value, statement,
callback, grouped-callee, range-bound, and unresolved assignment-target contexts.
Existing lexical values/routines and implicit members keep their lookup priority;
qualified user members and explicit calls retain their existing resolution.
Later parser punctuation at the semantic stop is suppressed.

Taking the address of an unshadowed bare special name reads that child's opening
parenthesis first. `@Length` stops with `"(" expected` before creating a callback
or reporting a result-level address error. Grouped/indexed operands preserve this
reading and stop before reading indices. A scalar expected type can emit
`unexpected "@"` before the child is read, as upstream does. Bare `Default`
follows ordinary name lookup without an intrinsic casing hint; its parenthesized
intrinsic and namespace forms retain their existing behavior.

Taking the address of a bare scalar type reports its missing `(` first, followed
by `unexpected "@"` at the address operator. Scalar variables report the latter
sentence alone. These errors allow subsequent statements to be checked; declared
routines, ordinary builtin references, and callable variables remain valid.

`DebugBreak` is a reserved, valueless intrinsic. Both `DebugBreak;` and
`DebugBreak()` execute as a no-op. If `(` is present, the next token must be `)`;
otherwise compilation stops with `")" expected` at that token before reading an
argument. Comments/newlines do not change this rule; EOF uses the opening `(`.
A direct malformed call retains its earlier pedantic casing hint. Ordinary
qualified member calls such as `obj.DebugBreak(1)` retain method lookup.

Value and grouped uses report `Expression expected` at the following token.
Taking its address reports that child error before `unexpected "@"`; it does not
create a callback. Constant recovery substitutes null and keeps the declared name
available to later checks, including through unary/binary constant operators.
Nested malformed-call casing hints remain gated by Phase 2.4's stopped-call AST
recovery; reserved declaration-name validation remains in Phase 3. Other
qualified namespace forms, Default resolution, and remaining reintroduced-property contexts
remain open in Phase 2.1; completed-call address recovery remains in Phase 2.2.

### Class, record and helper member headers

An invalid member after `class` reports `PROCEDURE or FUNCTION expected` at
the offending token and stops compilation. This also applies to top-level
`class` routine headers. At EOF, the diagnostic anchors at the final real
token, including when comments or whitespace follow it.

Unterminated record and helper bodies report `END expected` as compiler
stops. A helper body also expects `END` where an unsupported member starts.
Earlier declaration errors and visibility hints remain before the stop;
later declarations and deferred unimplemented-method checks are suppressed.
Record fields missing a colon retain Variant recovery types. Duplicate record
fields report `There is already a field with name "Name"` at the duplicate,
using the earlier declaration's casing.

Helpers begin with public visibility. Repeated visibility sections produce
normal-level `Redundant specifier` hints, and `protected` reports
`Helpers do not supported "protected" visibility specifier` without changing
the visibility in effect.

### Reintroduced property reads

A scalar class property declared with `reintroduce` accepts an empty `()` pair
immediately after its name. The explicit read `obj.Prop()` returns the same value
as `obj.Prop`. Within instance and class methods, a directly written `Prop()`
also reads the scalar property, including class properties in class methods.
At normal hint level it reports
`Property "Prop" reintroduced a method, you should remove empty brackets ()`
at the opening `(`. Bare reads have no compatibility hint. Pedantic hints report
wrong member casing before the compatibility hint, using declaration casing.
`{$HINTS OFF}` and disabled hints suppress the hint.

For `var A := obj.Prop(;`, `")" expected` is a recoverable error at `;`.
The declaration retains the property's type, and compilation can report later
semantic errors. Comments/newlines preserve the next-token anchor. An ordinary scalar
property called with `()` instead stops with `Not a method` at `(` before reading
arguments. Actual methods and ordinary function-valued properties retain their
existing call paths.

Checked scalar reads cover explicit receivers and unqualified empty calls in
class methods, including aliases/implicit routine receivers, inherited descriptors
and class properties. Local callable variables and parameters hide unqualified
properties; class members take priority over same-named global routines. Descendant
methods and helper methods retain their call paths. Getters execute once; virtual
instance/class getters keep dynamic dispatch and original exceptions. Unqualified
compatibility calls preserve the declaring descriptor when descendants redeclare
the property; reads also work inside another getter/setter. Static class methods
have no implicit receiver and stop before a compatibility hint. Deprecated
property warning order/anchors remain a diagnostic follow-up.
The source printer preserves the declaration marker. Nonempty compatibility-token recovery,
EOF/end boundaries, lexer-directive reach after recovery, discarded enclosing calls,
malformed/nonempty unqualified calls and explicit inherited syntax,
indexed/default properties, writes, flagged function-valued properties and unchecked
execution remain open in PLAN.md. These need their own parsing/dispatch acceptance;
passing the scalar fixtures does not close the full property item. Parser-only
clients retain the provisional `Expression expected` stop for a malformed boundary
call until semantic analysis resolves its property identity.

### Property and operator declarations

A missing property `read` or `write` name reports `Name expected` at the token
found instead (the final real token at EOF) and stops compilation. Earlier
accessor errors survive the stop; later declarations and unimplemented-method
checks do not add diagnostics.

Named accessor validation anchors at the accessor name, using the member's
**declared casing** in messages. A getter's result type is checked before its
parameters. A function used as a setter reports `Procedure expected` first.
Index type mismatches report `Parameter N - Type "X" expected (instead of "Y")`
with a zero-based index, followed by `Method "Name" has incompatible parameters`.
Getter type errors use `Field/method "Name" has an incompatible type`; writer
field type errors use `Symbol "Name" has an incompatible type`.

An invalid class operator reports `Overloadable operator expected` at its symbol
and stops immediately. An invalid global operator reports the same ordinary
error, then attempts the operand list: `operator dummy ;` also reports
`"(" expected` at `dummy` and stops.

Global operator operand-count errors anchor at the closing parenthesis. Binding
validation checks result type, parameter count, then parameter types at the
binding name. An earlier rejected binding still participates in duplicate
checks in its local scope, but never becomes an executable overload. Existing
unary-minus overloads remain supported.

### Incompatible type pairs

Ordinary scalar and interface assignments report
`Incompatible types: Cannot assign "X" to "Y"` at the right-hand expression.
Field-backed property writers use the same assignment sentence and position;
method-backed writers retain their setter-argument diagnostic at the property.
These rules also apply to inherited fields and implicit `Self` properties.
An assignment from a class that does not implement the destination interface
reports `Class "X" does not implement interface "Y"` at `:=`. Compatibility uses
the source's declared type, including when its current object implements the
interface. Method-backed property writers retain the setter-argument diagnostic.
Array assignments retain their context-specific positions: array expressions can
anchor at their operator or literal, while indexed writes normally anchor at `:=`.
For an allocation expression (`new T[...]`), an assignment mismatch anchors at
the opening `[` even when whitespace or a comment separates it from the type.
Each invalid allocation dimension reports `Integer expression expected`; analysis
continues through the remaining dimensions and retains the nested array type.

Incompatible array constructors retain their inferred element type and size in
assignment diagnostics. A dynamic-array target reports the mismatch at `:=`;
static-array targets retain the constructor's position. Constant-array values
also report assignment mismatches at `:=`. A write to a constant array element
reports `Cannot assign a value to the left-side argument` and still checks the
supplied value, reporting an incompatible RHS at the same operator.

An untyped empty constant constructor can initialize any dynamic array. Each use
gets independent storage with the destination's element type. This does not allow
an ordinary `array of Variant` variable to be assigned to `array of Integer`.

Compound assignments check both operand types. An unsupported scalar pair reports
`Incompatible operands` at the operator, then checks the RHS assignment and reports
`Cannot assign` at the supplied expression when needed. Registered operators still
require RHS assignment compatibility: `Integer += Float` remains invalid. A missing
class operator stops compilation at the operator. Numeric targets convert a Variant
RHS before the operation. A raised conversion preserves the target and original
exception, and runs the converter once.

Routine-pointer assignments preserve compatible bare routine names and pointer
copies as references. An incompatible bare routine is read as one call, retaining
its declared result even after a missing-argument error. A procedure result normally
reports `Assignment's right-side-argument has no return type` at the assignment
operator, after child diagnostics. General scalar-to-pointer assignments report
`Incompatible operands` at the operator, followed by `Cannot assign` at the
supplying expression, using the pointer's canonical diagnostic name. Simple
dynamic-array slot writes use the element-type mismatch at `:=`, including
`Cannot assign "void"` for procedure results.

A factory returning a compatible callable runs once; its result is stored for a
later call. This also applies to typed initializers and var/lazy factory parameters.
Lazy arguments evaluate on each read. A raising supplier preserves its original
exception and destination, and coalesce skips its fallback after that raise.

Typed constants and incompatible coalesce operands report
`Incompatible types: "X" and "Y"`, with the declared or left-hand type first.
Dynamic-array coalescing accepts derived-class elements on the right when the
left array holds their base class; the reverse direction is rejected.

For-in loops report class narrowing as
`Incompatible types: Cannot assign "TBase" to "TChild"`, with the collection's
element type first and the loop variable's type second, anchored at `do`.
Other loop-variable type mismatches report `Incompatible types: "X" and "Y"`,
with the variable's type first, anchored at `in`. Widening and same-class
assignments remain accepted.

Array constructors report incompatible row sizes with the current row's type
first and the previously inferred row's type second. A procedure used as an
element where its reference does not fit is read as a call: any
`More arguments expected` diagnostic precedes the `"void" and "nil"` (first
element) or `"void" and "Integer"` (after an integer) type pair. Explicit
`@Routine` references remain references when the array's type is inferred.

These constructor errors stop compilation. Their position is the scanner cursor
before the element is read: after the first element's first token, or immediately
after a separating comma. Consequently the position may be at the closing `]`,
in whitespace, or on the previous line. Later diagnostics are omitted.

An explicit `@Routine` passed as a set element reports `unexpected "@"`, the
incompatible routine signature, and then the enclosing parameter mismatch
against the recovered `nil` value.

### Array and range diagnostics

With `--diagnostics=plain`, array diagnostics use DWScript's sentences and
`[line: N, column: N]` positions. Indexing a non-indexable value reports
`Array expected`; each excess comma-separated index reports `Too many indices`
at its comma. Constant indices outside a static array's bounds report
`Lower bound exceeded! Index N` or `Upper bound exceeded! Index N`.

Incompatible range endpoints report
`Range start and range stop are of incompatible types: "X" and "Y"` in array
constructors, sets, membership tests, and case branches. Numeric case ranges may
mix Integer and Float. A reversed constant case range produces the hint
`Case range condition lower bound is greater than higher bound`.

Malformed array declarations and unfinished bracket literals retain diagnostics
from bounds or elements parsed before the error. An error inside an expression
is reported before the enclosing bound or range error, even when its displayed
column is later.

### 1. Syntax Errors (Parser)

**When**: During parsing, when source code doesn't match DWScript grammar

**Format**:
```
Error in <file>:<line>:<column>
  <line> | <source code>
           ^
<syntax error message>
```

**Example**:
```
Error in script.dws:5:10
   5 | var x Integer := 42;
              ^
expected ':' after identifier, got 'Integer'
```

**Features**:
- Shows exactly where the parser got confused
- Indicates what was expected vs what was found
- Highlights the problematic token

### 2. Type Errors (Semantic Analysis)

**When**: During semantic analysis, when types don't match

**Format**:
```
Error in <file>:<line>:<column>
  <line> | <source code>
           ^
<type error message>
  Expected: <type>
  Actual: <type>
```

**Example**:
```
Error in script.dws:12:15
  12 | count := price;
                    ^
Cannot assign Float to Integer variable 'count'
  Expected: Integer
  Actual: Float
```

**Features**
- Variable names included in error messages
- Expected and actual types clearly shown
- Context-aware messages (assignment, parameter, return value, etc.)
- Suggestions when applicable

**Common Type Errors**:

| Error | Message Format |
|-------|----------------|
| Assignment mismatch | `Cannot assign <type> to <type> variable '<name>'` |
| Parameter mismatch | `Function '<name>' expects <type> as parameter N, got <type>` |
| Return type mismatch | `Function '<name>' returns <type>, expected <type>` |
| Undefined variable | `Undefined variable '<name>'` |
| Undefined function | `Undefined function '<name>'` |

### 3. Runtime Errors

**When**: During execution, when operations fail at runtime

**Format**:
```
Runtime Error: <error type>: <message> [line: N, column: M]
<source code snippet with highlighting>
  <value context>
```

**Example**
```
Runtime Error: Division by zero [line: 15, column: 12]
  15 | result := a div b;
                      ^
Division by zero: 10 / 0
  Left operand: 10
  Right operand: 0
```

**Features**:
- Shows runtime values of operands
- Source code snippet with position highlighting
- Color-coded output (red for errors)
- Contextual information about the operation

**Common Runtime Errors**:

| Error Type | Message Format | Values Shown |
|------------|----------------|--------------|
| Division by zero | `Division by zero: <left> / <right>` | Both operands |
| Array index out of bounds | `Index <N> out of bounds for array of length <M>` | Index, length |
| Nil dereference | `Nil object reference at <location>` | Object name |
| Type conversion failure | `Cannot convert <value> to <type>` | Value, types |

### 4. Exceptions

**When**: Raised explicitly via `raise` statement or by runtime errors

**Format**:
```
Runtime Error: <ExceptionClass>: <message> [line: N, column: M]
<stack trace>
```

**Example** (Tasks 9.113, 9.114):
```
Runtime Error: Exception: Error from deep in the call stack [line: 7, column: 3]
DeepFunction [line: 13, column: 3]
MiddleFunction [line: 19, column: 3]
TopFunction [line: 26, column: 1]
```

**Features**:
- Exception class name (Exception, EDivByZero, etc.)
- Exception message
- Position where exception was raised
- Complete stack trace with all function calls
- Each stack frame shows: function name, line, column

**Stack Trace Format**:
```
FunctionName [line: N, column: M]
```

Each frame is listed from most recent (where the error occurred) to oldest (entry point).

A frame pairs a **name** with a **position**, and the two do not belong to the same routine:

- the name is the routine that *contains* the call site — the caller, not the callee;
- the position is the **name token of the thing being called**. `obj.Bar` reports the column of
  `Bar`, not of `obj`; `Exception.Create('x')` reports `Create`, neither `Exception` nor the
  closing paren.

A method is named with its class, whether its implementation is written inline in the class body
or out of line: `TMyClass.Boom`, never bare `Boom`.

The outermost frame's call site lies in the main program, which has no name, so its label is
empty — the line begins with a space:

```
TMyClass.Boom [line: 46, column: 20]
 [line: 51, column: 6]
```

Read that as: the raise happened at 46:20 inside `TMyClass.Boom`, and the main program called
`Boom` at 51:6.

## Color Coding

When output to a color-capable terminal, errors use ANSI color codes:

- **Red (`\x1b[1;31m`)**: Error messages, position markers
- **Reset (`\x1b[0m`)**: Back to normal text

Example colored output:
```
Error in script.dws:12:15
  12 | count := price;
                    ^        (in red)
Cannot assign Float to Integer variable 'count'  (in red)
```

## Error Message Improvements

### Type Error Messages

**Before**:
```
Type mismatch
```

**After**:
```
Error in script.dws:12:15
  12 | count := price;
                    ^
Cannot assign Float to Integer variable 'count'
  Expected: Integer
  Actual: Float
```

**Improvements**:
- Variable name included
- Both types shown
- Exact location
- Source code context

### Runtime Error Messages

**Before**:
```
ERROR: division by zero
```

**After**:
```
Runtime Error: Division by zero [line: 15, column: 12]
  15 | result := a div b;
                      ^
Division by zero: 10 / 0
  Left operand: 10
  Right operand: 0
```

**Improvements**:
- Operand values shown
- Source code snippet
- Position highlighting
- Visual error marker

### Source Code Snippets

All errors include source code context:
- The line containing the error
- Position marker (`^`) pointing to exact location
- Optional: 1-2 lines of surrounding context

## Stack Traces (Tasks 9.113-9.116)

### Automatic Stack Traces

Unhandled exceptions automatically display stack traces:

```bash
./bin/dwscript run script.dws
```

Output:
```
Runtime Error: Exception: Error message [line: 7, column: 3]
Level3 [line: 7, column: 3]
Level2 [line: 13, column: 3]
Level1 [line: 19, column: 3]
```

Note that the position on the `Runtime Error:` line and the position of the innermost stack frame
answer different questions and can differ. The message reports the `raise` statement — just past
the raised expression — while the innermost frame reports where the exception object was
*constructed*, at the constructor's name token.

### Programmatic Stack Traces

Two built-in functions provide stack trace access:

#### GetStackTrace()

Returns formatted string:

```pascal
var trace: String;
trace := GetStackTrace();
PrintLn(trace);
```

Output:
```
FunctionName [line: 10, column: 5]
CallerFunction [line: 20, column: 3]
```

#### GetCallStack()

Returns structured array:

```pascal
var stack: array of Variant;
stack := GetCallStack();
PrintLn('Depth: ' + IntToStr(Length(stack)));
```

Each element is a record with:
- `FunctionName: String`
- `FileName: String`
- `Line: Integer`
- `Column: Integer`

## Error Context

### Variable Names

Type errors include variable names when applicable:

```
Cannot assign Float to Integer variable 'count'
```

### Function Names

Function-related errors include function names:

```
Function 'ProcessNumber' expects Integer as parameter 1, got Float
```

### Type Information

Type errors show both expected and actual types:

```
Expected: Integer
Actual: Float
```

### Runtime Values

Runtime errors show relevant values:

```
Division by zero: 10 / 0
  Left operand: 10
  Right operand: 0
```

## Best Practices

### For Developers

1. **Read the full error message**: Don't just look at the line number
2. **Check the source snippet**: The `^` marker shows exactly where the error is
3. **Review type information**: Expected vs actual types guide the fix
4. **Use stack traces**: Understand the call path that led to the error
5. **Look for suggestions**: Some errors include hints for fixes

### For Error Message Design

1. **Be specific**: Include variable/function names
2. **Show context**: Display source code and values
3. **Be actionable**: Suggest what to fix
4. **Be consistent**: Follow the standard format
5. **Provide details**: Types, values, locations

## Error Reporting Infrastructure

### CompilerError Type

Defined in `internal/errors/compiler_error.go`:

```go
type CompilerError struct {
    Position   lexer.Position
    Message    string
    SourceCode string
    FileName   string
}
```

**Methods**:
- `Format(useColor bool) string`: Formats error with source snippet
- `String() string`: Simple string representation

### RuntimeError Type

Defined in `internal/interp/errors.go`:

```go
type RuntimeError struct {
    Message    string
    Pos        *lexer.Position
    Expression string
    Values     map[string]string
    SourceCode string
    SourceFile string
    ErrorType  string
    CallStack  errors.StackTrace
}
```

**Methods**:
- `ToCompilerError() *CompilerError`: Converts to formatted error
- `String() string`: Simple string representation

### ExceptionValue Type

Defined in `internal/interp/exceptions.go`:

```go
type ExceptionValue struct {
    ClassInfo *ClassInfo
    Instance  *ObjectInstance
    Message   string
    Position  *lexer.Position
    CallStack errors.StackTrace
}
```

## Testing Error Messages

Test fixtures are available in `testdata/error_messages/`:

1. **Type errors**: 01-03
2. **Runtime errors**: 04-05
3. **Exception handling**: 06-08
4. **Combined**: 09

Run with:
```bash
./bin/dwscript run testdata/error_messages/<fixture>.dws
```

See `testdata/error_messages/README.md` for details.

## Command-Line Interface

### Error Display

The CLI (`cmd/dwscript/cmd/run.go`) compiles through `internal/frontend` and presents its
diagnostics in one of two styles, selected with `run --diagnostics=pretty|plain`:

1. **pretty** (default): compile errors as `CompilerError` blocks with a source excerpt via
   `errors.FormatErrors()`; runtime errors with source; unhandled exceptions with class, message,
   position and stack trace. ANSI colors are used only on a terminal and never when `NO_COLOR` is set.
   Runtime-generated messages that already end with the reported position display
   that position once. User-authored exception messages and distinct re-raise
   positions are preserved.
2. **plain**: the DWScript wire format below, one message per line, all severities in the
   compiler's order (`frontend.Result.DiagnosticStrings()`); runtime errors via
   `interp.FormatRuntimeErrorValue`. Exit code 1 on any error, nothing else printed.

Two further flags exist for test harnesses and imply `--diagnostics=plain`:

- `--test-envelope`: program output is wrapped in DWScript's test-runner framing when there is at
  least one message (`Errors >>>>`, hints/warnings, the runtime error, `Result >>>>`, output).
- `--compile-only`: parse and type-check without executing; prints the compiler's full message
  list (hints included at the chosen `--hints` level), which is what the `*Fail` fixture suites expect.

`cmd/fixture-report` uses exactly these flags, so its numbers match the Go harness.
Both runners use pedantic hints for the categories upstream's `UScriptTests` collects, and
strict hints (the compiler default) everywhere else; see
[the fixture README](../../testdata/fixtures/README.md#evidence-for-case-mismatch-hint-settings). Missing `.txt` expectations mean empty output except for the documented
[category exclusions](../../testdata/fixtures/README.md#a-missing-txt-means-must-print-nothing-upstream);
unexpected diagnostics still fail those checks.

### Hint levels and source controls

`run --hints off|normal|strict|pedantic` sets the initial hint level. Case-mismatch
hints require `pedantic`; name lookup remains case-insensitive at every level. Hints use
the resolved declaration's spelling, including assignments, routine references and calls,
class and helper members, and built-in registry names. Repeated analysis of the same
identifier during overload selection emits its case hint once.

Source directives control hints from that point onward:

| Directive | Effect |
| --- | --- |
| `{$HINTS OFF}` | Suppress hints |
| `{$HINTS ON}` | Restore the compiler's configured initial level |
| `{$HINTS NORMAL}` | Select normal hints |
| `{$HINTS STRICT}` | Select strict hints |
| `{$HINTS PEDANTIC}` | Select pedantic hints, including case mismatches |

These settings apply to semantic hints and explicit `{$HINT 'text'}` messages. An
include inherits the current setting and can change it for following source; a separately
compiled unit starts with the configured initial level. Inactive conditional branches do
not change settings. Case hints retain the setting at the identifier's source location even
when analysis of a routine body is deferred. Warnings and errors are independent of
`{$HINTS}`. The CLI's `--hints off` also hides hint/warning output during execution;
`--compile-only` always prints the resulting diagnostic list.

### Flow and contract warnings

`Warning: Unreachable code` identifies the first unreachable statement in a
statement list after an unconditional `exit`, `raise`, `break`, or `continue`,
including branches that all interrupt execution. Analysis continues through the
remaining statements so their errors are still reported. This follows DWScript's
unoptimized flow rules; an `if True then exit` without an `else` does not make the
next statement unreachable for this warning.

`Warning: Constant condition` identifies a constant `require` or `ensure` test.
Both warnings are independent of the hint level. Diagnostics from inline methods
and deferred routines retain declaration order, including when another statement
or routine appears between class declarations.

A function's implicit `Result` cannot be redeclared as a local variable, even in
a nested block. The error points to the redeclared name. A separate procedure or
global scope may still declare a variable named `result`; identifiers remain case
insensitive. `Result := Result` receives the ordinary self-assignment hint.

### Symbol-dictionary diagnostics

`run --symbol-dictionary-diagnostics=false` disables hints that depend on upstream's
symbol dictionary: unused locals, unused `Result`, unused private members, and
reference `var` parameters that are never written. The default is `true`, preserving
ordinary CLI behavior. This setting also applies to imported units; it does not
disable case hints, source-message directives, or other diagnostics.

Methods that implement an interface are exempt from unused-private hints, including
inherited implementations. Other unused private methods still produce hints when
dictionary diagnostics and the appropriate hint level are enabled.

The fixture harness and CLI report set this option explicitly from the same category
policy. Execution categories collected by `UScriptTests` disable it; that runner's
nonoptimized failure categories enable it. Other categories retain their existing
setting. See [the fixture README](../../testdata/fixtures/README.md#symbol-dictionary-diagnostics).

### Example Output

```bash
$ ./bin/dwscript run script.dws
Error in script.dws:12:15
  12 | count := price;
                    ^
Cannot assign Float to Integer variable 'count'
Error: semantic analysis failed with 1 error(s)
```

## Future Enhancements

Potential improvements for error messages:

1. **Suggestions**: "Did you mean `cout` instead of `count`?"
2. **Error codes**: `E0001: Type mismatch`
3. **Multi-line errors**: Show errors spanning multiple lines
4. **Error recovery**: Continue after errors to show multiple issues
5. **IDE integration**: Machine-readable error format (JSON)
6. **Documentation links**: Link to docs for error types

## Related Documentation

- `docs/guide/exceptions.md` - Exception handling reference
- `testdata/error_messages/README.md` - Test fixtures overview

## Conclusion

The enhanced error messages in go-dws provide developers with:
- **Clear context**: Source code, positions, values
- **Actionable information**: Types, names, suggestions
- **Complete traces**: Full call stack when needed
- **Consistent format**: Easy to read and understand

These improvements significantly enhance the developer experience and make debugging faster and more effective.

## Appendix: DWScript wire format for diagnostics

The fixture corpus and the `*Fail` suites expect the exact DWScript serialization, one diagnostic per line:

```
Syntax Error: <message> [line: X, column: Y]
Error: <message> [line: X, column: Y]         (runtime errors)
Compile Error: <message> [line: X, column: Y] ({$ERROR} and {$FATAL} directives)
Hint: <message> [line: X, column: Y]
Warning: <message> [line: X, column: Y]
```

`internal/frontend.Result.DiagnosticStrings()` (via `Diagnostic.Render()` → `dwserrors.FormatDWScriptError`) produces this shape and is what the fixture harness compares against. `cmd/dwscript run --diagnostics=plain` prints exactly this; the default pretty blocks are a presentation layer on top of it. Historical standardization work: `docs/history/task-6.3-summary.md`.

### Diagnostics from compiler directives

The message directives emit diagnostics straight from the lexer:

| Directive | Output |
| --- | --- |
| `{$HINT 'text'}` | `Hint: text [line: X, column: Y]` |
| `{$WARNING 'text'}` | `Warning: text [line: X, column: Y]` |
| `{$ERROR 'text'}` | `Compile Error: text [line: X, column: Y]` |
| `{$FATAL 'text'}` | `Compile Error: text [line: X, column: Y]`, then compilation stops |

`{$ERROR}` and `{$FATAL}` share a prefix and differ only in whether compilation continues.
`{$FATAL}` halts tokenization immediately, so nothing after it is compiled, while every
message recorded before it is still reported. Both `'…'` and `"…"` quoting is accepted.

The position is the directive *name* — the column of `{` plus two. Argument diagnostics
(`String expected`, `ON/OFF expected`) anchor at the argument when one is present and at the
closing brace otherwise. A directive inside an inactive `{$IF}`/`{$IFDEF}` branch emits nothing.

`{$HINTS}` and `{$WARNINGS}` take `ON`, `OFF`, `NORMAL`, `STRICT` or `PEDANTIC`;
`{$R}`/`{$RESOURCE}` require a quoted string. An unrecognized switch reports
`Compiler switch "NAME" unknown`, and a directive missing its closing brace reports
`"}" expected`.

Unlike a failed `{$INCLUDE}`, these do not block semantic analysis: the source before a
`{$FATAL}` parsed correctly, so its errors are still reported alongside the fatal.

### `Declared()` and `ConditionalDefined()`

`Declared('Name')` is a compile-time predicate available both as an ordinary expression and
inside `{$IF}`. It takes a constant string, resolves it case-insensitively, and folds to a
boolean. Dotted names reach members (`Declared('TObject.Create')`), a leading `Internal.`
segment denotes the built-in unit (`Declared('Internal.Sin')`), and helper methods are visible
through the helped type. Bare member names do not resolve: a record field `Dummy` does not make
`Declared('dummy')` true. A non-String argument reports `String expected`, and a non-constant
argument inside `{$IF}` reports `Constant expression expected`.

`Defined('X')` is different and narrower: it asks only whether `X` is a `{$DEFINE}` symbol.
`dwscript run --define X script.dws` adds that symbol before scanning both the script and
its imported units; repeat `--define` for additional symbols.

Two limitations: `ConditionalDefined()` always folds to `False` because `{$DEFINE}` symbols are
not reachable from the analyzer, and inside `{$IF}` the set of declared names comes from a
forward-only scan of the token stream, so it sees only what precedes the directive.
