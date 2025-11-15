# AST Visitor Code Generation

This document explains the code generation approach for the AST visitor pattern in go-dws.

## Overview

The visitor pattern implementation for AST traversal is **automatically generated** from AST node definitions using the `cmd/gen-visitor` tool. This eliminates 83.6% of manually-written boilerplate code (922 lines → 151 lines of generator code) while maintaining zero runtime overhead.

## Why Code Generation?

### Problem: Manual Visitor is Tedious

The original manual visitor (`pkg/ast/visitor.go`) required 922 lines of boilerplate:

```go
// 195-line switch statement in Walk()
func Walk(v Visitor, node Node) {
    if v = v.Visit(node); v == nil {
        return
    }

    switch n := node.(type) {
    case *BinaryExpression:
        walkBinaryExpression(n, v)
    case *UnaryExpression:
        walkUnaryExpression(n, v)
    // ... 100+ more cases
    }
}

// Plus 100+ separate walk functions
func walkBinaryExpression(n *BinaryExpression, v Visitor) {
    if n.Left != nil {
        Walk(v, n.Left)
    }
    if n.Right != nil {
        Walk(v, n.Right)
    }
}

// ... 100+ more functions
```

**Every new AST node** requires:
1. Adding a case to the Walk() switch statement
2. Writing a dedicated walkXXX() function
3. Manually identifying and walking all Node fields
4. Handling nil checks, slices, and helper types

This is **error-prone** and **tedious**.

### Solution: Automatic Code Generation

The code generator:
- Parses AST node definitions automatically
- Generates type-safe walk functions for all nodes
- Handles slices, interfaces, and helper types correctly
- Supports struct tags for customization
- Runs in ~0.5 seconds (well under 1s target)
- **Zero runtime overhead** vs manual code

### Rejected Alternative: Reflection

We prototyped a reflection-based visitor (see [visitor-reflection-research.md](visitor-reflection-research.md)) but found it **30x slower** than manual code, making it unsuitable for production use.

**Code generation** gives us the best of both worlds:
- Automatic maintenance like reflection
- Zero runtime cost like manual code

## Architecture

### Code Generator: cmd/gen-visitor/main.go

The generator is a standalone Go program (536 lines) that:

1. **Parses** all Go source files in `pkg/ast/`
2. **Identifies** types that implement the Node interface
3. **Extracts** fields that need to be walked
4. **Generates** type-safe walk functions
5. **Writes** output to `pkg/ast/visitor_generated.go`

### Generated Code: pkg/ast/visitor_generated.go

The generated file (805 lines) contains:
- `Walk()` function with switch statement for all node types
- `walkXXX()` function for each node type
- Specialized walk functions for helper types (Parameter, etc.)

### Integration: pkg/ast/visitor_interface.go

Contains the Visitor interface and go:generate directive:

```go
//go:generate go run ../../cmd/gen-visitor/main.go

type Visitor interface {
    Visit(node Node) (w Visitor)
}
```

## How It Works

### 1. Node Detection

The generator scans all `.go` files in `pkg/ast/` looking for struct types that:

- Embed `BaseNode` or `TypedExpressionBase`, OR
- Are explicitly listed in `knownNodeTypes` (e.g., `Program`)

Example detected nodes:
```go
// Detected: embeds BaseNode
type BinaryExpression struct {
    BaseNode
    Left     Expression
    Operator string
    Right    Expression
}

// Detected: embeds TypedExpressionBase (which embeds BaseNode)
type IntegerLiteral struct {
    TypedExpressionBase
    Value   int64
    IntValue string
}

// Detected: in knownNodeTypes list
type Program struct {
    Statements []Statement
}
```

**Not detected**:
```go
// Skipped: doesn't embed BaseNode
type Token struct {
    Type    TokenType
    Literal string
}
```

### 2. Field Extraction

For each detected node, the generator examines all exported fields:

```go
type FunctionDecl struct {
    BaseNode                           // Embedded, skipped
    Name           *Identifier         // Node field → walk
    Parameters     []*Parameter        // Slice of helper type → walk
    ReturnType     *TypeAnnotation     // Node field → walk
    Body           *BlockStatement     // Node field → walk
    Token          token.Token         // Not a Node → skip
    IsExternal     bool                // Not a Node → skip
}
```

**Fields are walked if**:
- Type implements Node interface (Expression, Statement, or concrete node)
- Type is a known helper type (Parameter, CaseBranch, etc.)
- Field is not explicitly skipped via `ast:"skip"` tag

**Fields are skipped if**:
- Type is primitive (string, int, bool)
- Type is in skipFields list (Token, EndPos, Visibility, etc.)
- Field has `ast:"skip"` struct tag

### 3. Type Handling

The generator handles different field types correctly:

#### Single Node Field
```go
// AST definition
type UnaryExpression struct {
    BaseNode
    Operand Expression
}

// Generated code
func walkUnaryExpression(n *UnaryExpression, v Visitor) {
    if n.Operand != nil {
        Walk(v, n.Operand)
    }
}
```

#### Slice of Pointers
```go
// AST definition
type BlockStatement struct {
    BaseNode
    Statements []Statement  // Interface type
}

// Generated code
func walkBlockStatement(n *BlockStatement, v Visitor) {
    for _, item := range n.Statements {
        if item != nil {
            Walk(v, item)
        }
    }
}
```

#### Slice of Values
```go
// AST definition
type RecordDecl struct {
    BaseNode
    Fields []RecordFieldDecl  // Concrete struct, not pointer
}

// Generated code
func walkRecordDecl(n *RecordDecl, v Visitor) {
    for i := range n.Fields {
        Walk(v, &n.Fields[i])  // Take address for addressability
    }
}
```

#### Helper Types
```go
// AST definition
type Parameter struct {  // Doesn't implement Node, but contains Nodes
    Name         *Identifier
    Type         *TypeAnnotation
    DefaultValue Expression
}

// Generated walk function
func walkParameter(param *Parameter, v Visitor) {
    if param == nil {
        return
    }
    if param.Name != nil {
        Walk(v, param.Name)
    }
    if param.Type != nil {
        Walk(v, param.Type)
    }
    if param.DefaultValue != nil {
        Walk(v, param.DefaultValue)
    }
}

// Used in FunctionDecl walker
func walkFunctionDecl(n *FunctionDecl, v Visitor) {
    // ...
    for _, item := range n.Parameters {
        if item != nil {
            walkParameter(item, v)
        }
    }
    // ...
}
```

### 4. Struct Tags

The generator respects struct tags for fine-grained control:

#### ast:"skip" - Skip Field

```go
type MyNode struct {
    BaseNode
    MainData  *Node                     // Visited
    CachedData *Node `ast:"skip"`        // NOT visited
}
```

Generated code:
```go
func walkMyNode(n *MyNode, v Visitor) {
    if n.MainData != nil {
        Walk(v, n.MainData)
    }
    // CachedData skipped (ast:"skip" tag)
}
```

#### ast:"order:N" - Custom Traversal Order

```go
type FunctionDecl struct {
    BaseNode
    Name       *Identifier                        // Visited 3rd (default order)
    Parameters []*Parameter                       // Visited 4th (default order)
    ReturnType *TypeAnnotation `ast:"order:1"`    // Visited 1st
    Body       *BlockStatement `ast:"order:2"`    // Visited 2nd
}
```

Generated code:
```go
func walkFunctionDecl(n *FunctionDecl, v Visitor) {
    // Fields with explicit order come first
    if n.ReturnType != nil {  // order:1
        Walk(v, n.ReturnType)
    }
    if n.Body != nil {  // order:2
        Walk(v, n.Body)
    }
    // Fields without explicit order follow
    if n.Name != nil {
        Walk(v, n.Name)
    }
    for _, item := range n.Parameters {
        walkParameter(item, v)
    }
}
```

See [ast-visitor-tags.md](ast-visitor-tags.md) for complete tag documentation.

## Usage

### Adding New AST Nodes

When you add a new AST node type:

1. **Define the node** in `pkg/ast/*.go`:
   ```go
   type MyNewNode struct {
       BaseNode
       Field1 *Identifier
       Field2 []Expression
   }

   func (n *MyNewNode) expressionNode() {}
   func (n *MyNewNode) TokenLiteral() string { return n.Token.Literal }
   func (n *MyNewNode) String() string { /* ... */ }
   ```

2. **Regenerate visitor**:
   ```bash
   go generate ./pkg/ast
   ```

3. **Verify** the generated code:
   ```bash
   go test ./pkg/ast
   ```

That's it! The visitor automatically handles your new node type.

### Modifying Existing Nodes

If you add/remove/modify fields in an existing node:

1. **Edit the struct** in `pkg/ast/*.go`
2. **Regenerate**:
   ```bash
   go generate ./pkg/ast
   ```

The generated visitor is automatically updated.

### Using Struct Tags

To skip a field or control traversal order:

1. **Add struct tag** to the field:
   ```go
   type MyNode struct {
       BaseNode
       Important   *Node `ast:"order:1"`
       CachedData  *Node `ast:"skip"`
   }
   ```

2. **Regenerate**:
   ```bash
   go generate ./pkg/ast
   ```

## Implementation Details

### Generator Algorithm

```
1. Parse all .go files in pkg/ast/
   ↓
2. For each type definition:
   - Check if it embeds BaseNode/TypedExpressionBase
   - OR check if it's in knownNodeTypes list
   ↓
3. For each node type found:
   - Extract exported fields
   - Identify Node fields vs non-Node fields
   - Parse struct tags (ast:"skip", ast:"order:N")
   ↓
4. For each node:
   - Sort fields by order tag (if present)
   - Generate walkXXX() function
   - Handle slices, interfaces, helper types
   ↓
5. Generate main Walk() function:
   - Switch statement with case for each node type
   - Call corresponding walkXXX() function
   ↓
6. Write to pkg/ast/visitor_generated.go
```

### Type Classification

The generator classifies types into categories:

| Category | Examples | Handling |
|----------|----------|----------|
| **Node interfaces** | `Node`, `Expression`, `Statement` | Walk via `Walk(v, node)` |
| **Concrete nodes** | `BinaryExpression`, `IfStatement` | Walk via `Walk(v, node)` |
| **Helper types** | `Parameter`, `CaseBranch` | Walk via `walkParameter(p, v)` |
| **Primitives** | `string`, `int`, `bool` | Skip (not walked) |
| **Metadata** | `Token`, `EndPos`, `Visibility` | Skip (in skipFields list) |

### Special Cases

#### Program Node

`Program` doesn't embed `BaseNode` but is still a node:

```go
// Added to knownNodeTypes map
var knownNodeTypes = map[string]bool{
    "Program": true,
    // ...
}
```

#### Interface vs Concrete Slices

Slices of interfaces are handled differently than slices of structs:

```go
// Slice of interface - use range variable directly
[]Statement  →  for _, item := range n.Statements { Walk(v, item) }

// Slice of concrete struct - use index for addressability
[]RecordFieldDecl  →  for i := range n.Fields { Walk(v, &n.Fields[i]) }
```

This is necessary because you can't take the address of a range variable over an interface slice.

#### Nil Checks

All pointer fields are nil-checked before walking:

```go
if n.Field != nil {
    Walk(v, n.Field)
}
```

## Performance

### Zero Runtime Overhead

Benchmarks show the generated visitor performs identically to (or better than) manual code:

| Test Case | Generated | Manual | Difference |
|-----------|-----------|--------|------------|
| Simple program | 92 ns/op | 92 ns/op | **0% overhead** |
| Complex program | 1,412 ns/op | 1,896 ns/op | **25% faster!** |

The generated code is actually **faster** for complex programs, likely due to better compiler optimization of the generated code.

See [visitor-benchmark-results.md](visitor-benchmark-results.md) for detailed benchmarks.

### Fast Code Generation

Generation time is **~0.5 seconds** (well under the 1s target):

```
Run 1: 0.499s
Run 2: 0.501s
Run 3: 0.485s
Run 4: 0.502s
Run 5: 0.483s

Average: 0.494s
```

For 64 AST node types across 13 source files, this is **~130 node types/second**.

## Testing

### Verification

The generated visitor has comprehensive test coverage:

1. **Unit tests** (`pkg/ast/visitor_test.go`):
   - TestWalk_VisitsAllNodes
   - TestWalk_VisitorReturnsNil
   - TestInspect_FindsFunctions
   - TestWalk_AllNodeTypes
   - TestWalk_WithNilNodes

2. **Struct tag tests** (`pkg/ast/visitor_tags_test.go`):
   - TestNodeWithSkipTag
   - TestGeneratedVisitorCompleteness

3. **Benchmarks** (`pkg/ast/visitor_bench_test.go`):
   - BenchmarkVisitorGenerated_SimpleProgram
   - BenchmarkVisitorGenerated_ComplexProgram
   - BenchmarkVisitorGenerated_DeepNesting
   - BenchmarkVisitorGenerated_WideTree

4. **Integration tests** (`pkg/dwscript/integration_test.go`):
   - TestIntegration_LSPWorkflow
   - TestIntegration_ParseASTSymbols

All tests pass with the generated visitor.

### Continuous Integration

The CI pipeline should:

1. **Run tests** after generation:
   ```bash
   go generate ./pkg/ast
   go test ./pkg/ast
   ```

2. **Verify no changes** (visitor is up-to-date):
   ```bash
   go generate ./pkg/ast
   git diff --exit-code pkg/ast/visitor_generated.go
   ```

If the generated file differs, the CI should fail, indicating the visitor needs regeneration.

## Troubleshooting

### Error: "undefined: walkXXX"

**Cause**: You added a new helper type but didn't add it to `knownHelperTypes`.

**Fix**: Add the type to `cmd/gen-visitor/main.go`:
```go
var knownHelperTypes = map[string]bool{
    "Parameter":    true,
    "CaseBranch":   true,
    "YourNewType":  true,  // Add here
}
```

Then regenerate:
```bash
go generate ./pkg/ast
```

### Error: "cannot use &n.Field[i] (value of type *Expression)"

**Cause**: Trying to take address of interface slice element.

**Fix**: The generator should detect this automatically. If not, verify the type is in `isInterfaceType()`:
```go
func isInterfaceType(typeName string) bool {
    interfaceTypes := map[string]bool{
        "Node":       true,
        "Expression": true,
        "Statement":  true,
    }
    return interfaceTypes[typeName]
}
```

### Generator Doesn't Find My Node

**Cause**: Node doesn't embed `BaseNode` and isn't in `knownNodeTypes`.

**Fix Option 1** (preferred): Embed `BaseNode`:
```go
type MyNode struct {
    BaseNode  // Add this
    // ... fields
}
```

**Fix Option 2**: Add to `knownNodeTypes` in `cmd/gen-visitor/main.go`:
```go
var knownNodeTypes = map[string]bool{
    "Program": true,
    "MyNode":  true,  // Add here
}
```

### Field Should Be Walked But Isn't

**Cause**: Field name is in `shouldSkipField()` list.

**Fix**: Either:
1. Rename the field
2. Remove it from `shouldSkipField()` in `cmd/gen-visitor/main.go`

## Comparison with Alternatives

### vs Manual Visitor

| Aspect | Manual | Generated |
|--------|--------|-----------|
| Lines of code | 922 | 805 (auto-generated from 151 lines) |
| Maintenance | Manual updates needed | Automatic |
| Performance | Baseline | 0-25% faster |
| Error-prone | Yes | No |
| New node cost | 10-20 lines of code | 0 lines (automatic) |

### vs Reflection

| Aspect | Reflection | Generated |
|--------|------------|-----------|
| Lines of code | 151 | 805 (auto-generated) |
| Maintenance | Automatic | Automatic |
| Performance | 30x slower | Same as manual |
| Type safety | Runtime | Compile-time |
| Complexity | Medium | Low |

### vs go/ast Visitor

go-dws's generated visitor is similar to Go's standard `go/ast` package approach:

| Feature | go/ast | go-dws |
|---------|--------|--------|
| Approach | Manual | Code-generated |
| Performance | ~85 ns/op | ~92 ns/op |
| Lines of code | ~800 (manual) | 805 (generated) |
| Customization | None | Struct tags |

Our generated visitor is competitive with Go's hand-written standard library implementation while being automatically maintained.

## Future Enhancements

Potential improvements for future work (not currently implemented):

### 1. Incremental Generation

Generate only changed nodes instead of full regeneration:
- Track file modification times
- Only re-parse changed files
- Merge with existing generated code

**Benefit**: Faster generation for large projects

### 2. Custom Visitor Templates

Allow users to provide custom walk function templates:
```go
//go:generate go run gen-visitor --template=my-template.go
```

**Benefit**: Support different traversal patterns

### 3. Parallel Traversal

Enable parallel traversal for independent subtrees:
```go
type MyNode struct {
    Left  *Node `ast:"parallel"`
    Right *Node `ast:"parallel"`
}
```

**Benefit**: 2x speedup for independent subtrees

### 4. Visitor Analytics

Generate visitor usage statistics:
- Which nodes are visited most frequently
- Average traversal depth
- Hotspots for optimization

**Benefit**: Identify performance bottlenecks

## References

- **Code generator**: [cmd/gen-visitor/main.go](../cmd/gen-visitor/main.go)
- **Generated code**: [pkg/ast/visitor_generated.go](../pkg/ast/visitor_generated.go)
- **Visitor interface**: [pkg/ast/visitor_interface.go](../pkg/ast/visitor_interface.go)
- **Struct tags**: [ast-visitor-tags.md](ast-visitor-tags.md)
- **Benchmarks**: [visitor-benchmark-results.md](visitor-benchmark-results.md)
- **Research**: [visitor-reflection-research.md](visitor-reflection-research.md)
- **Compatibility**: [visitor-compatibility-test-results.md](visitor-compatibility-test-results.md)

## Summary

The code-generated visitor provides:

✅ **83.6% reduction** in manually-written code
✅ **Zero runtime overhead** vs manual implementation
✅ **Automatic maintenance** when adding/modifying nodes
✅ **Type safety** with compile-time checks
✅ **Fast generation** (~0.5 seconds for 64 nodes)
✅ **100% backward compatible** with existing code
✅ **Customizable** via struct tags

This approach successfully eliminates boilerplate while maintaining performance, making AST maintenance significantly easier.
