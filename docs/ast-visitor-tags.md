# AST Visitor Struct Tags

This document describes how to use struct tags to control visitor behavior for AST nodes.

## Overview

The code-generated visitor supports struct tags that allow fine-grained control over how fields are traversed. This is useful when you need to:
- Skip certain fields during traversal
- Control the order in which fields are visited
- Optimize visitor performance for specific use cases

## Supported Tags

### `ast:"skip"` - Skip Field Traversal

Use this tag to prevent a field from being walked by the visitor, even if it implements the Node interface.

**Use cases:**
- Performance optimization: Skip expensive subtrees that aren't needed for your analysis
- Semantic boundaries: Skip certain relationships that shouldn't be traversed
- Temporary fields: Exclude cached or computed fields from traversal

**Example:**

```go
type MyNode struct {
    BaseNode
    MainChild  *SomeNode                    // Visited normally
    CachedData *SomeNode `ast:"skip"`      // Skipped - not visited
    Children   []Node                       // Visited normally
}
```

**Generated code:**

```go
func walkMyNode(n *MyNode, v Visitor) {
    if n.MainChild != nil {
        Walk(v, n.MainChild)
    }
    // CachedData skipped (ast:"skip" tag)
    for _, item := range n.Children {
        if item != nil {
            Walk(v, item)
        }
    }
}
```

### `ast:"order:N"` - Custom Traversal Order

Use this tag to specify the order in which fields should be visited. Fields are sorted by their order value before generation.

**Use cases:**
- Semantic ordering: Visit declarations before uses
- Dependency ordering: Process dependencies before dependents
- Analysis optimization: Visit most important nodes first

**Example:**

```go
type FunctionDecl struct {
    BaseNode
    Name       *Identifier                 // Visited 3rd (no explicit order)
    Parameters []*Parameter                // Visited 4th (no explicit order)
    ReturnType *TypeAnnotation `ast:"order:1"` // Visited 1st
    Body       *BlockStatement  `ast:"order:2"` // Visited 2nd
}
```

**Generated code:**

```go
func walkFunctionDecl(n *FunctionDecl, v Visitor) {
    // Fields with explicit order come first, sorted by order value
    if n.ReturnType != nil {  // order:1
        Walk(v, n.ReturnType)
    }
    if n.Body != nil {  // order:2
        Walk(v, n.Body)
    }
    // Fields without explicit order follow in original struct order
    if n.Name != nil {
        Walk(v, n.Name)
    }
    for _, item := range n.Parameters {
        walkParameter(item, v)
    }
}
```

**Ordering rules:**
1. Fields with `ast:"order:N"` are visited first, sorted by N (ascending)
2. Fields without explicit order are visited after, in their original struct order
3. If two fields have the same order value, they're visited in struct order

### Combined Tags

You can combine multiple tags (though skip and order together doesn't make semantic sense):

```go
type ComplexNode struct {
    BaseNode
    Field1 *Node `ast:"order:1"`           // Visited first
    Field2 *Node `ast:"order:2"`           // Visited second
    Field3 *Node                           // Visited third (original order)
    Field4 *Node `ast:"skip"`              // Never visited
}
```

## Implementation Details

### Tag Parsing

The code generator parses tags during AST analysis:

```go
// extractFields in cmd/gen-visitor/main.go
if field.Tag != nil {
    tag := field.Tag.Value

    // Check for skip
    if strings.Contains(tag, `ast:"skip"`) {
        skip = true
    }

    // Check for order
    if strings.Contains(tag, "order:") {
        // Parse order value (e.g., "order:10")
        order = parseOrderValue(tag)
    }
}
```

### Field Sorting

Fields are sorted before code generation:

```go
// sortFieldsByOrder in cmd/gen-visitor/main.go
func sortFieldsByOrder(fields []*FieldInfo) []*FieldInfo {
    sort.SliceStable(fields, func(i, j int) bool {
        fi, fj := fields[i], fields[j]

        // Fields with explicit order come first
        if fi.Order > 0 && fj.Order > 0 {
            return fi.Order < fj.Order
        }
        if fi.Order > 0 {
            return true
        }
        if fj.Order > 0 {
            return false
        }

        // Maintain original order for fields without explicit order
        return originalIndex[i] < originalIndex[j]
    })
}
```

## Examples

### Example 1: Type Checker Optimization

Process type annotations before analyzing expressions:

```go
type AssignmentStatement struct {
    BaseNode
    Left  Expression `ast:"order:2"`  // Process after type is known
    Right Expression `ast:"order:2"`  // Process after type is known
    Type  *TypeAnnotation `ast:"order:1"` // Process type hint first
}
```

### Example 2: Skipping Cached Data

Avoid visiting computed/cached fields:

```go
type Identifier struct {
    TypedExpressionBase
    Value string
    Type  *TypeAnnotation `ast:"skip"` // Type is set during analysis, don't walk it
}
```

### Example 3: Semantic Ordering

Visit preconditions and postconditions in the right order:

```go
type FunctionDecl struct {
    BaseNode
    Name           *Identifier
    Parameters     []*Parameter
    ReturnType     *TypeAnnotation
    PreConditions  *PreConditions `ast:"order:1"`  // Check before body
    Body           *BlockStatement `ast:"order:2"` // Main implementation
    PostConditions *PostConditions `ast:"order:3"` // Check after body
}
```

## Best Practices

### 1. Use Tags Sparingly

Don't over-optimize. The default traversal order (struct field order) is usually fine. Only use tags when:
- You have a specific semantic requirement (preconditions before body)
- You've identified a performance bottleneck
- You're implementing a specific algorithm that requires certain ordering

### 2. Document Your Tags

When you add tags, document why:

```go
type MyNode struct {
    BaseNode
    // Dependencies must be processed before the main body
    Imports []Import `ast:"order:1"`
    Body    *Block   `ast:"order:2"`
}
```

### 3. Keep Order Values Simple

Use multiples of 10 (10, 20, 30) to leave room for future additions:

```go
type Node struct {
    Phase1Field *Foo `ast:"order:10"`
    Phase2Field *Bar `ast:"order:20"`
    Phase3Field *Baz `ast:"order:30"`
    // Easy to add order:15 field later
}
```

### 4. Test Your Changes

After adding tags, regenerate the visitor and verify behavior:

```bash
go generate ./pkg/ast
go test ./pkg/ast -v
```

## Regenerating After Changes

Whenever you modify struct tags, regenerate the visitor:

```bash
# From project root
go generate ./pkg/ast

# Or run the generator directly
go run cmd/gen-visitor/main.go
```

## Performance Impact

- **Skip tag**: Can improve performance by avoiding unnecessary subtree traversal
- **Order tag**: Zero runtime cost - ordering happens at code generation time
- **Both tags**: Add zero runtime overhead compared to manual visitor

## Future Enhancements

Potential future tag features (not yet implemented):

- `ast:"visit:funcName"` - Custom visit function
- `ast:"depth:N"` - Maximum traversal depth
- `ast:"cache"` - Cache traversal results

## See Also

- [Visitor Pattern Documentation](ast-visitor-codegen.md)
- [Code Generator Implementation](../cmd/gen-visitor/main.go)
- [PLAN.md Task 9.17.4](../PLAN.md) - Struct tag implementation task
