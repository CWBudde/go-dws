# ADR: Generated AST visitor instead of reflection

**Status:** Accepted (November 2025, tasks 9.17.1 to 9.17.7 of the pre-2026-07 plan)

## Context

`pkg/ast` needs a `Walk`/`Inspect` traversal over roughly 80 node types. The original hand-written walker (`pkg/ast/visitor_legacy.go`) was a 900-line type switch that had to be edited by hand every time a node gained a child field, and it silently skipped fields that nobody remembered to add.

Two alternatives were prototyped and measured:

1. **Reflection-based walker.** 151 lines instead of 922. Discovers `Node` fields, slices of nodes, and helper structs at runtime and honours an `ast:"skip"` tag. Correct, but the cost is CPU time in `reflect.Value.Field`/`Interface` calls on every node.
2. **Code-generated walker.** A generator (`cmd/gen-visitor`) parses the node definitions in `pkg/ast/*.go` and emits a type switch equivalent to the manual one, so the maintenance benefit of reflection is kept without any runtime cost.

Benchmarks (`pkg/ast/visitor_bench_test.go`, ns/op, 24 B/op and 2 allocs/op in every variant):

| Program | Manual | Reflection | Generated |
|---|---:|---:|---:|
| Simple program | 92 to 113 | 2,537 to 2,632 | 92 |
| Standard program | 2,177 | 66,224 | n/a |
| Complex program (enum, record, functions, loops, try) | 1,896 to 2,148 | 65,609 to 67,367 | 1,412 |

Reflection is 22x to 31x slower than the manual walker in the first research round and 28x to 48x slower in the final round. The plan's acceptance threshold was under 10% overhead. The generated walker matches the manual one on simple input and is about 25% faster on complex input, most likely because the emitted code optimises better.

Compatibility testing of the generated walker ran the full `pkg/ast` visitor suite plus the LSP integration tests (`TestIntegration_ParseASTSymbols`, `TestIntegration_LSPWorkflow`, `TestIntegration_NoRegressions` and the rest) with no code changes to callers: 100% pass, no public API change.

## Decision

Use the generated visitor. `pkg/ast/visitor_generated.go` is produced by `cmd/gen-visitor/main.go` and is the traversal used by `Walk` and `Inspect`.

Regenerate after changing any AST node:

```bash
go run cmd/gen-visitor/main.go
# or
go generate ./pkg/ast
```

Generator features that callers rely on:

- **Type-aware field handling.** Only fields whose type is a `Node` (`Expression`, `Statement`, concrete node pointers, or slices of those) are walked. Primitive fields such as `BinaryExpression.Operator` (a string) are skipped automatically, while `AddressOfExpression.Operator` (an `Expression`) is walked.
- **Embedded base types.** A struct is recognised as a node when it embeds `BaseNode`; the embedded base itself holds no walkable children (type annotations live in `SemanticInfo`, not on the node), so it contributes no per-node code.
- **Helper types.** Non-`Node` helpers (`Parameter`, `CaseBranch`, `ExceptClause`) get their own walk functions, called from the parent node's walker.
- **Struct tags.** `ast:"skip"` excludes a field; `ast:"order:N"` overrides the default source-order traversal (tags can be combined as `ast:"skip,order:10"`).
- `TestGeneratedVisitorCompleteness` fails when a node type exists without a generated case, which is the guard against the "forgot to add the field" failure mode of the manual walker.

The reflection prototype (`visitor_reflect.go`) was removed.

## Consequences

- Adding a node type or a child field is a one-step change: edit the node struct, regenerate, done.
- Traversal cost is identical to hand-written code, so the LSP, formatter research, and semantic passes can walk freely.
- `pkg/ast/visitor_legacy.go` (a 68-case hand-written switch) still exists alongside the 83-case generated file. It is no longer the traversal used by `Walk` and is a cleanup candidate once nothing references it.
- The generator is one more tool to keep in sync with Go syntax changes to the AST package; it parses source with `go/ast`, so ordinary struct edits are safe.

## References

- [`../archive/visitor-reflection-research.md`](../archive/visitor-reflection-research.md): prototype, first benchmark round, recommendation to generate code.
- [`../history/TASK-9-17-1-SUMMARY.md`](../history/TASK-9-17-1-SUMMARY.md): research summary for the reflection prototype.
- [`../archive/visitor-benchmark-results.md`](../archive/visitor-benchmark-results.md): final benchmarks for manual, reflection, and generated walkers.
- [`../archive/visitor-compatibility-test-results.md`](../archive/visitor-compatibility-test-results.md): backward-compatibility test log.
- `cmd/gen-visitor/main.go`, `pkg/ast/visitor_generated.go`, `pkg/ast/visitor_tags_test.go`.
