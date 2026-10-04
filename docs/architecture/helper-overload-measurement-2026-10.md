# Helper overload candidate measurement — October 2026

This note records the source evidence for the first candidate-selection batch in
PLAN.md §1.6. The acceptance matrix describes intended compile diagnostics;
verification results belong in the progress log. It does not establish complete
helper overload compatibility or runtime dispatch parity.

## Pinned source and provenance

The reference is DWScript commit
[`1dbf8a90329cc3f2638516e89c0668f916c1ddb9`](https://github.com/EricGrange/DWScript/tree/1dbf8a90329cc3f2638516e89c0668f916c1ddb9).
The locally inspected source files were checked against the Git blob hashes in
the saved recursive tree for that commit:

| Source | Git blob SHA |
| --- | --- |
| `Source/dwsCompiler.pas` | `cffe42f756f8f0a44eeae0e045d6bb2bbe7878b4` |
| `Source/dwsSymbols.pas` | `b10f40daf7130fc62c2bf5edf36d51d2c4e142fe` |
| `Source/dwsExprs.pas` | `4cd954122252e5a4b4e505acc4eed7a88b35fe84` |
| `Source/dwsCompilerUtils.pas` | `73f667cbf1050fff6415c24f8012640b2983b384` |

The compiler copy `/tmp/phase15-dwsCompiler.pas` matches this pin. The older
`/tmp/go-dws-13-pointer-upstream-compiler.pas` does not match it; its line numbers
must not be used as the pinned compiler evidence.

## Candidate selection and selected-signature checking

`ReadTypeHelper` first chooses a helper that supplies the member, minimizing
receiver-to-target distance and stopping at an exact target. It does not pool
same-name members from all applicable helpers.
[Source: helper lookup](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsCompiler.pas#L14827-L14839).

The looked-up method's `IsOverloaded` flag chooses the overload path, even when
only one marked declaration is visible. An unmarked singleton proceeds directly
to argument checking. Consequently, a marked singleton with wrong types or
counts reports no matching overload, rather than the ordinary argument-type or
count diagnostic.
[Source: helper overload dispatch](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsCompiler.pas#L14904-L14906).

`CollectMethodOverloads` gathers visible same-name methods from the declaring
structure. It can walk parents while the last encountered method is overloaded,
and excludes parent methods already represented by a child method. There is no
instance/class/static filter in this collector. Helpers use `ReadMethOverloaded`;
the native `ReadStaticMethOverloaded` filtering path is separate.
[Source: collection and call paths](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsCompiler.pas#L7251-L7328).

For parenthesized calls, arguments are read before ranking. Successful ranking
changes the function symbol and then checks the selected signature. Failed
ranking emits one no-overload diagnostic at the member's script position, returns an error
expression of `AnyType`, and skips selected-signature checking. Finding a helper
but failing to select an overload therefore does not justify another
inaccessible-member diagnostic.
[Source: failure and recovery](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsCompiler.pas#L7522-L7527),
[source: read, select, check](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsCompiler.pas#L7669-L7679).

## Receiver ownership and mixed sets

| Helper method | Upstream parameter role |
| --- | --- |
| Instance method | `Self` has the helper target type |
| Nonstatic class method for a class or record | `Self` has the target's metatype |
| Class method for a primitive or interface | No `Self` parameter |
| Static class method | No `Self` parameter |

These roles follow `THelperSymbol.CreateSelfParameter`; `SetIsStatic` removes any
previous Self parameter.
[Source: helper Self](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsSymbols.pas#L9652-L9673),
[source: static removal](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsSymbols.pas#L4939-L4947).

For homogeneous instance sets, the bound receiver participates as argument zero;
an explicit helper-name call writes that receiver as its first argument. For
mixed sets, upstream constructs the argument list using the initially looked-up
method: `TOverloadedExpr` inserts the receiver only when that method starts with
`Self`. Ranking then compares every candidate against this same argument list.
The initial helper-member read may also convert an object receiver to a metatype.
[Source: overloaded expression](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsExprs.pas#L7551-L7574),
[source: receiver conversion](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsCompiler.pas#L14871-L14902).

The initial method cannot be assumed to be the first source declaration: member
lookup uses a sorted table and binary search, and sorting can exchange equal-name
entries. Mixed receiver roles require separate concrete measurements; ranking
only the written arguments against signatures with Self removed is insufficient
to describe upstream behavior.
[Source: symbol lookup](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsSymbols.pas#L6966-L7002),
[source: sorting](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsSymbols.pas#L7043-L7073).

## Recoverable children and declaration order

The argument reader continues through recoverable child errors before outer
ranking. A common parameter expectation is available only when all candidates
support that position with compatible expected types.
[Source: argument reading](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsCompiler.pas#L7732-L7801),
[source: common expectation](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsExprs.pas#L7998-L8014).

For outer overloads `Take(Integer, String)` and `Take(String, Integer)`, consider
`Take(Inner(true), Other(true))`: `Inner` is a marked singleton Integer function
expecting Integer; `Other` is an ordinary Integer function expecting Integer.
The source-derived order is Inner no-overload, Other argument-type error, then
Take no-overload. Inner recovers as AnyType; both outer candidates reject it.
Ranking asks the parameter type whether it accepts the argument type, so
AnyType's permissive compatibility in the opposite direction does not help.
Error-expression skipping applies only after successful ranking.
[Source: ranking direction](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsCompiler.pas#L7363-L7368),
[source: ordinary type compatibility](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsSymbols.pas#L6203-L6207),
[source: Integer compatibility](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsSymbols.pas#L6243-L6249),
[source: AnyType compatibility](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsSymbols.pas#L8928-L8931),
[source: error-expression skip](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsCompilerUtils.pas#L683-L708).

Unknown identifiers are a separate recovery case: the pinned compiler calls
`AddCompilerStop`, so reading every later child is not established for them.
[Source: unknown-name stop](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsCompiler.pas#L4918-L4922).
Inline bodies are compiled immediately after their method is added, before later
declarations; the existing `HelpersFail/helper_overload_error` fixture covers this
boundary.
[Source: inline body timing](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsCompiler.pas#L3860-L3874).

## Bounded acceptance matrix

The compile acceptance assertions live in
[helper_overload_candidates_test.go](../../internal/frontend/helper_overload_candidates_test.go).
Verification results are recorded in the October progress log.

| Context | Candidate cases | Expected boundary |
| --- | --- | --- |
| Primitive, record, class, interface receivers | Marked singleton: valid, short, excess, wrong type | Valid selects; failures report one no-overload at the member |
| The same receivers | Two homogeneous instance overloads: valid, short, no-match | Selection and rejection occur before selected-signature checking |
| Singleton and multiple-candidate receiver calls | `[]` and `nil` | Preserve contextual array/reference arguments |
| Explicit helper name | Marked singleton: valid, short, wrong supplied/receiver type | Written Self participates in candidate matching |
| Bound helper-body call | Marked singleton: short, wrong type | Use the visible lexical declaration set |
| Receiver, explicit helper, bound body | Nested recoverable no-match plus failing ordinary sibling | Both child errors precede the enclosing no-overload |
| Receiver and explicit helper | Successful selection with a failing ordinary child | Emit that child diagnostic once |

## Remaining measurement and runtime work

The first batch does not replace the shared Go overload-ranking policy. Upstream
uses Variant-to-Integer cost 256 versus cost 1 for other compatible Variant
promotions, and boxing into Variant costs 128. It also favors deeper declaring
structures on equal parameter scores; a final tie produces an ambiguity hint
and retains the first ranked candidate. Go's shared resolver has different
weights and reports unresolved ties as errors. These require further measurements.
[Source: Variant ranking](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsCompiler.pas#L7375-L7385),
[source: boxing and ties](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsCompiler.pas#L7449-L7495).

Further work includes mixed instance/class/static sets, runtime dispatch agreeing
with the selected declaration, metaclass helper targets, imported-helper
availability, and contextual arguments across multiple differing signatures.
Go's `helper(ParentHelper)` syntax also needs its own extension contract: the
pinned `ReadHelperDecl` requires `for` immediately and does not parse a parent
helper. Helpers applicable through target-class inheritance can be measured
against this pin; parent-helper syntax cannot be presented as pinned support.
[Source: helper declaration grammar](https://github.com/EricGrange/DWScript/blob/1dbf8a90329cc3f2638516e89c0668f916c1ddb9/Source/dwsCompiler.pas#L10429-L10441).

## Current-source runtime reproductions left open

The source-built CLI confirms two execution defects, independent of the newly
checked compiler-failure cases. These are follow-ups, not accepted language behavior.

```pascal
type H = helper for Integer
 function Pick(v: Integer): String; overload; begin Result := 'integer'; end;
 function Pick(v: String): String; overload; begin Result := 'string'; end;
end;
var item: Integer;
PrintLn(item.Pick('x'));       // currently integer; selected signature is String
PrintLn(H.Pick(item, 'x'));    // string
```

Non-record `CallHelperMethod` chooses the first candidate of matching arity in
`internal/interp/evaluator/helper_methods.go`. The explicit helper-name path uses
signature ranking. Runtime execution should retain the compiler's selected
declaration and owner, including lazy and statically typed Variant arguments,
rather than independently selecting from evaluated values.

The port's parent-helper extension also collapses an inherited runtime overload
set to the parent's primary (last) method when the child has no own set:

```pascal
type H = helper for Integer
 function Pick(v: Integer): Integer; overload; begin Result := 11; end;
 function Pick(v: String): Integer; overload; begin Result := 12; end;
end;
type J = helper(H) for Integer end;
var item: Integer;
PrintLn(item.Pick(1));       // currently 12
PrintLn(J.Pick(item, 1));    // currently 12
PrintLn(H.Pick(item, 1));    // 11
```

`internal/interp/runtime/helper_metadata.go` falls back through `GetMethod` rather
than preserving the parent's complete set and owner. Agreement on the Integer
candidate is a port-extension correctness requirement; it is not evidence of
upstream parent-helper syntax support. Unrelated-helper precedence, helper-body
owner capture, and alias-specific implicit Self calls require their own runtime
acceptance cases as that follow-up proceeds.
