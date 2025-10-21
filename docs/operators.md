# Operator Overloading Research

This note consolidates the DWScript operator overloading behavior that we need to mirror in go-dws. It combines observations from the upstream source tree under `reference/dwscript-original/` and community guidance from [StackOverflow: Does DWScript support operator overloading?](https://stackoverflow.com/questions/6203402/does-dwscript-support-operator-overloading).

## Primary References

- `reference/dwscript-original/Test/OperatorOverloadPass/*.pas` — canonical positive coverage for standalone, implicit, and symbolic operators.
- `reference/dwscript-original/Test/SimpleScripts/class_operator*.pas` — class operator usage and inheritance examples.
- `reference/dwscript-original/Test/FailureScripts/class_operator*.pas`, `in_operator3.pas`, `class_operator3.txt` — negative cases defining expected diagnostics.
- `reference/dwscript-original/Source/dwsOperators.pas`, `dwsCompiler.pas` — operator registry plumbing inside the DWScript compiler.

These fixtures illustrate the exact syntax surface and runtime semantics we should reproduce.

## DWScript Syntax Surface

### Global Operator Declarations

Standalone operators live at the unit level and bind to global functions through the `uses` clause.

```pascal
function StrPlusInt(s : String; i : Integer) : String;
begin
   Result := s + '[' + IntToStr(i) + ']';
end;

operator + (String, Integer) : String uses StrPlusInt;
```

Key characteristics:

- Token selection uses the operator literal (`+`, `-`, `*`, `IN`, `==`, `!=`, `<<`, `>>`, etc.).
- Argument lists appear inside parentheses without parameter names.
- Return type is mandatory.
- `uses Identifier` associates the operator with an existing function symbol.
- Multiple overloads for the same token and different operand types are permitted.

### Symbolic Operator Tokens

`operator_overloading2.pas` confirms support for symbolic names beyond the Pascal core:

- `<<`, `>>` for custom stream piping.
- `==`, `!=` for equality.
- `+=` and other compound tokens (see class operators below).
- `IN` for membership tests with overload hooks.

These map back to specific `TokenKind` entries. We must normalize symbolic strings during parsing so overload lookup can key off the existing lexer token representations.

### Unary Operators

Unary declarations reuse the same grammar, with a single operand type inside the parentheses. Upstream tests cover unary `-` and `+`, and the runtime falls back to the default operator when no overload matches.

### Conversion Operators

Implicit/explicit conversions are expressed with the `implicit` / `explicit` keywords replacing the token:

```pascal
operator implicit (Integer) : String uses IntToStr;
operator explicit (TFoo) : Integer uses FooToInt;
```

- Only one operand type is supplied.
- Return type indicates the target type of the conversion.
- `implicit` conversions participate automatically during assignment/call resolution; `explicit` conversions require an explicit cast.

### Class Operators

Within a class body, `class operator` exposes operators bound to static methods:

```pascal
type TTest = class
   Field : String;
   procedure AppendString(str : String);
   class operator += String uses AppendString;
end;
```

Class operator traits:

- Declared in the class scope; `uses` must reference a class method (static or class procedure/function).
- Support compound assignment tokens (`+=`, `-=`), the membership operator `IN`, and standard arithmetic operators.
- Inheritance-aware: subclasses inherit parent class operators (`class_operator1.pas`).
- Failure scripts demonstrate duplicate registration errors and invalid signatures (`class_operator2.pas`, `class_operator3.txt`).

### Forward Declarations and External Binding

`class_operator3.pas` and `class_operator3.txt` show diagnostics for missing operator bodies. There is no standalone forward syntax for operators, so our parser should reject missing `uses` clauses or body references.

## Semantic Expectations

- **Registration**: Operators populate tables keyed by token + operand types. `dwsOperators.pas` keeps a catalog of expression classes used by the runtime.
- **Resolution Order**: The compiler first attempts class operators when operands include class instances. Fallback to global operators occurs when class matches fail.
- **Ambiguity Handling**: Failure fixtures emit diagnostics like “Class operator already defined for type `String`” and “Overloadable operator expected.” We should mirror these messages where practical.
- **Implicit Conversions**: Semantic analysis inserts implicit conversion calls when a compatible `operator implicit` exists. Tests such as `implicit_record1.pas` demonstrate chains of conversions.
- **Runtime Execution**: Interpreter invokes the bound function/method identified by `uses`. For `class operator`, the receiver instance participates as `Self` when applicable (e.g., compound assignments in `class_operator1.pas`).

## Implementation Notes for go-dws

1. **AST Enhancements**
   - Introduce node variants for global operator declarations, class operator declarations, and conversion operators.
   - Capture operator token, arity, operand types, return type, `uses` identifier, and source span for diagnostics.

2. **Parser Coverage**
   - Reuse expression token tables to validate the token after `operator`.
   - Support keywords `implicit`, `explicit`, and `IN`.
   - Ensure parentheses contain either one type (conversion) or two types (binary); enforce trailing semicolons.

3. **Type System Registries**
   - Add per-scope operator registries: global, class, and conversion.
   - Define lookup routines that mirror DWScript precedence (class before global; implicit conversion search during type compatibility).

4. **Semantic Analyzer**
   - Validate `uses` reference (function for global, class method for class operator).
   - Check signatures: operand and return types must match the declaration.
   - Enforce duplicate detection and ambiguous overload errors.
   - Insert implicit conversion nodes when needed.

5. **Interpreter**
   - Dispatch global operators via stored function references.
   - Invoke class operators with proper `Self` binding.
   - Run implicit conversions automatically where inserted by the analyzer.

6. **Testing Strategy**
   - Port positive/negative fixtures from the upstream `Test/OperatorOverloadPass/`, `Test/SimpleScripts/`, and `Test/FailureScripts/` directories into `testdata/operators/`.
   - Provide unit tests for parser, semantic analysis, and interpreter layers that cover the syntax combinations above.
   - Add CLI integration scripts exercising stream-style operators (`<<`, `>>`), implicit conversions, and compound assignments.

## Outstanding Questions

- **Unary vs. Binary Token Collision**: Need to confirm whether DWScript differentiates unary `-` and binary `-` in operator tables or relies on arity alone.
- **Generic Constraints**: `Test/GenericsFail/binop_constraint.txt` indicates operators participate in generic constraint resolution; ensure future generics work references this behavior.
- **External Operators**: Upstream runtime exposes mechanisms for Delphi-side operator registration (`dwsByteBufferFunctions.pas`). Evaluate whether go-dws needs an equivalent hook for Go-native extensions.

Capturing these findings satisfies Stage 8 task 8.1 and seeds the downstream work (tasks 8.2 onward) with concrete inputs from the reference implementation.
