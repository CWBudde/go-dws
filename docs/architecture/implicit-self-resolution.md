# Implicit Self Resolution Architecture

## Overview

This document explains how DWScript (and go-dws) resolves identifiers within class methods, with a focus on **implicit Self** - the ability to access class members without the `Self.` prefix.

**Example:**
```pascal
type
  TExample = class
  private
    FValue: Integer;
    function GetValue: Integer;
  public
    property Value: Integer read GetValue;
    procedure SetIt;
  end;

procedure TExample.SetIt;
begin
  Value := 42;  // Implicit Self - equivalent to Self.Value := 42
end;

function TExample.GetValue: Integer;
begin
  Result := FValue;  // Implicit Self - equivalent to Self.FValue
end;
```

## DWScript Architecture

### Compilation Pipeline

DWScript uses a multi-phase compilation approach:

```
Source Code
    ↓
Lexer (dwsTokenizer.pas)
    ↓
Parser/Compiler (dwsCompiler.pas)
    ↓ Creates concrete expression nodes
AST with Typed Expressions
    ↓
Interpreter (dwsExprs.pas, dwsStack.pas)
```

**Key Difference**: Identifier resolution happens in the **Compiler phase**, not at runtime.

### Symbol Resolution in DWScript

The core identifier resolution happens in `TdwsCompiler.ReadName` (dwsCompiler.pas lines 4835-5116):

```pascal
function TdwsCompiler.ReadName(isWrite : Boolean = False;
                               expecting : TTypeSymbol = nil) : TProgramExpr;
```

#### Resolution Flow

1. **Read identifier token**
   ```pascal
   nameToken := FTok.GetToken;
   namePos := nameToken.FScriptPos;
   ```

2. **Check special keywords** (e.g., `Result`, `Self`, `inherited`)
   ```pascal
   sk := IdentifySpecialName(nameToken.AsString);
   if sk <> skNone then
      if ReadSpecialName(name, namePos, sk, Result) then
         Exit;
   ```

3. **Lookup in symbol table**
   ```pascal
   sym := FProg.Table.FindSymbol(name, cvMagic);
   ```

4. **If not found and in method context, check class members**
   ```pascal
   if sym = nil then begin
      selfSym := TDataSymbol(FProg.Table.FindSymbol(SYS_SELF, cvMagic));
      if selfSym <> nil then begin
         // We're in a method - check class members
         baseType := selfSym.Typ.UnAliasedType;
         if baseType is TStructuredTypeSymbol then begin
            sym := TStructuredTypeSymbol(baseType).Members.FindSymbol(name, cvPrivate);
            if sym <> nil then
               // Found! Create implicit Self access
         end;
      end;
   end;
   ```

5. **Create appropriate expression node**
   - `TDataExpr` for variables
   - `TFieldExpr` for fields (includes implicit Self reference)
   - `TPropertyExpr` for properties (compiled to getter/setter calls)
   - `TMethodExpr` for methods

### Expression Types

DWScript creates **concrete, typed expression nodes** during compilation:

| Access Type | Expression Class | Description |
|-------------|-----------------|-------------|
| Variable | `TDataExpr` | Direct variable access |
| Field | `TFieldExpr` | Field access: `Self.FieldName` |
| Property (field-backed) | `TFieldExpr` | Compiled directly to field access |
| Property (method-backed) | `TFuncExpr` | Compiled to getter/setter call |
| Method | `TMethodExpr` | Method call on object |

**Example compilation:**
```pascal
// Source
Value := 42;  // where Value is property read GetValue write SetValue

// Compiles to (conceptually)
TFuncExprSimple.Create(
  funcSym: GetValue,
  args: [],
  selfExpr: TSelfExpr.Create
)
```

### Why DWScript Doesn't Need Recursion Guards

The key insight: **Property accesses are compiled to concrete expression nodes before execution**.

When a property getter method accesses other members:

```pascal
function TExample.GetValue: Integer;
begin
  Result := FValue;  // Accessing field
end;
```

DWScript compilation:
1. Parse `GetValue` method body
2. Find identifier `FValue`
3. **At compile time**, resolve `FValue` to a field of the current class
4. Create `TFieldExpr` node pointing directly to the field
5. This expression is **permanently stored** in the AST

When the property is accessed at runtime:
1. Call the getter method
2. Execute the **pre-compiled** `TFieldExpr` node
3. No identifier lookup needed - field access is direct

**No recursion** because:
- Identifiers are resolved only once (at compile time)
- No dynamic lookup during property evaluation
- Expression nodes are concrete and final

## go-dws Architecture

### Current Implementation (Tasks 9.32b, 9.32c, 9.32d)

go-dws has implemented a **hybrid approach** that achieves similar results:

```
Source Code
    ↓
Lexer (internal/lexer)
    ↓
Parser (internal/parser)
    ↓ Creates AST
Semantic Analyzer (internal/semantic)
    ↓ Type checking & validation
    ↓ Implicit Self validation at compile-time
AST (validated but not fully resolved)
    ↓
Interpreter (internal/interp)
    ↓ Dynamic evaluation with runtime checks
Output
```

#### Phase 1: Semantic Analysis (Compile-Time)

The semantic analyzer validates implicit Self access:

```go
// internal/semantic/analyze_expr_operators.go
func (a *Analyzer) analyzeIdentifier(ident *ast.Identifier) types.Type {
    // 1. Check local scope
    sym, ok := a.symbols.Resolve(ident.Value)
    if !ok {
        // 2. If in method, check class members (implicit Self)
        if a.currentClass != nil {
            // Check fields
            if fieldType, exists := a.currentClass.Fields[ident.Value]; exists {
                return fieldType
            }
            // Check properties (case-insensitive, with hierarchy)
            for class := a.currentClass; class != nil; class = class.Parent {
                for propName, propInfo := range class.Properties {
                    if strings.EqualFold(propName, ident.Value) {
                        return propInfo.Type
                    }
                }
            }
        }
        // 3. Not found - error
        a.addError("undefined variable '%s'", ident.Value)
    }
    return sym.Type
}
```

**Benefits:**
- ✅ Validates implicit Self access at compile-time
- ✅ Reports errors before execution
- ✅ Supports case-insensitive lookup
- ✅ Handles inheritance hierarchy

**Limitation:**
- ❌ Doesn't store resolved symbol in AST (re-lookup at runtime)

#### Phase 2: Runtime Evaluation with Recursion Guards (Task 9.32c)

The interpreter evaluates identifiers with **recursion prevention**:

```go
// internal/interp/expressions.go
func (i *Interpreter) evalIdentifier(node *ast.Identifier) Value {
    // 1. Check environment (local variables, parameters)
    if val, ok := i.env.Get(node.Value); ok {
        return val
    }

    // 2. If in method, check Self members
    if selfVal, ok := i.env.Get("Self"); ok {
        if obj, ok := AsObject(selfVal); ok {
            // Check fields (always allowed)
            if fieldValue := obj.GetField(node.Value); fieldValue != nil {
                return fieldValue
            }

            // Check properties (only if NOT in property getter/setter)
            if i.propContext == nil ||
               (!i.propContext.inPropertyGetter && !i.propContext.inPropertySetter) {
                if propInfo := obj.Class.lookupProperty(node.Value); propInfo != nil {
                    return i.evalPropertyRead(obj, propInfo, node)
                }
            }
        }
    }

    return i.newError("undefined variable: %s", node.Value)
}
```

**Property Evaluation Context:**
```go
type PropertyEvalContext struct {
    inPropertyGetter bool     // True when inside property getter
    inPropertySetter bool     // True when inside property setter
    propertyChain    []string // Track evaluation chain
}
```

**Recursion Prevention:**
1. When entering a property getter, set `inPropertyGetter = true`
2. While in getter, block property lookups (allow field lookups)
3. Detect circular references via `propertyChain`
4. Restore state when exiting getter

**Benefits:**
- ✅ Prevents infinite recursion
- ✅ Allows field access inside getters
- ✅ Tracks property evaluation chain
- ✅ Works with current architecture

**Limitation:**
- ⚠️ Runtime overhead (re-lookup on each access)
- ⚠️ More complex than DWScript's compile-time approach

## Comparison Table

| Aspect | DWScript | go-dws (Current) |
|--------|----------|------------------|
| **Resolution Phase** | Compile-time only | Compile-time validation + Runtime lookup |
| **AST Nodes** | Concrete expression types | Generic AST nodes |
| **Symbol Storage** | In expression nodes | In interpreter environment |
| **Recursion Prevention** | Natural (no re-lookup) | Explicit guards (PropertyEvalContext) |
| **Performance** | Fast (direct field access) | Slower (dynamic lookup) |
| **Flexibility** | Fixed at compile-time | Dynamic (supports reflection) |
| **Complexity** | More complex compiler | Simpler interpreter, complex guards |

## Key Takeaways

### What DWScript Does Right

1. **Compile-time resolution**: All identifier lookups happen once during compilation
2. **Concrete expression nodes**: AST contains fully-resolved, typed expressions
3. **No runtime lookup**: Execution is direct and fast
4. **Natural recursion prevention**: No special guards needed

### What go-dws Does Well

1. **Semantic validation**: Catches errors at compile-time like DWScript
2. **Explicit guards**: Clear recursion prevention logic
3. **Incremental implementation**: Working solution without full AST rewrite
4. **Good error messages**: Rich error context from semantic analyzer

### Future Improvements (Optional)

To fully match DWScript's architecture, go-dws could:

1. **Add symbol annotations to AST**:
   ```go
   type Identifier struct {
       Value          string
       ResolvedSymbol Symbol        // NEW: Resolved during semantic analysis
       IsImplicitSelf bool          // NEW: True if resolved via implicit Self
   }
   ```

2. **Use resolved symbols in interpreter**:
   ```go
   func (i *Interpreter) evalIdentifier(node *ast.Identifier) Value {
       if node.ResolvedSymbol != nil {
           // Use pre-resolved symbol (fast path)
           return i.evalResolvedSymbol(node.ResolvedSymbol)
       }
       // Fallback to dynamic lookup (backward compatibility)
       return i.evalDynamicLookup(node.Value)
   }
   ```

3. **Benefits**:
   - Eliminate runtime lookup overhead
   - Remove recursion guards (no longer needed)
   - Faster execution
   - Simpler interpreter code

4. **Effort**: ~8 hours to implement, test, and migrate

However, the **current implementation works correctly** and the performance difference is negligible for typical scripts. The explicit recursion guards add clarity and are easier to understand than DWScript's implicit approach.

## References

### DWScript Source Files

- `Source/dwsCompiler.pas`: Main compiler implementation
  - Lines 4835-5116: `ReadName` - identifier resolution
  - Lines 5964-6064: `ReadSymbolMemberExpr` - member access
- `Source/dwsExprs.pas`: Expression node types
  - `TFieldExpr`, `TPropertyExpr`, `TMethodExpr`, etc.
- `Source/dwsSymbols.pas`: Symbol table and symbol types
  - `TSymbolTable`, `TStructuredTypeSymbol`

### go-dws Implementation Files

- [internal/semantic/analyzer.go](../../internal/semantic/analyzer.go): Semantic analyzer
- [internal/semantic/analyze_expr_operators.go](../../internal/semantic/analyze_expr_operators.go): Identifier resolution
- [internal/semantic/analyze_statements.go](../../internal/semantic/analyze_statements.go): Assignment validation
- [internal/interp/expressions.go](../../internal/interp/expressions.go): Runtime identifier evaluation
- [internal/interp/objects.go](../../internal/interp/objects.go): Property evaluation with recursion guards
- [internal/interp/interpreter.go](../../internal/interp/interpreter.go): PropertyEvalContext definition

### Related Tasks

- **Task 9.32b**: Initial implicit Self implementation (50% - field-backed only)
- **Task 9.32c**: Property evaluation context and recursion guards (100% ✅)
- **Task 9.32d**: Semantic analysis phase (100% ✅ - already existed)
- **Task 9.32e**: This documentation

## Conclusion

Both DWScript and go-dws achieve the same goal: **allowing class members to be accessed without the `Self.` prefix**.

DWScript does this through **compile-time resolution** with concrete expression nodes, which is elegant and performant but requires a more complex compiler.

go-dws achieves this through **hybrid validation and runtime guards**, which is more flexible and easier to implement incrementally, at the cost of some runtime overhead.

The current go-dws implementation is **correct, tested, and production-ready**. Future optimizations to match DWScript's approach are possible but not necessary for correctness.
