# Bytecode VM: experimental and parked

## Status

- **Owner decision (2026-07-04):** keep `internal/bytecode` in the tree, unmaintained and opt-in. Do not extend it.
- The delete-versus-rebuild question is open. Revisit it after the type-system collapse (`PLAN.md` item A6), because a rebuilt VM would have to sit on the shared `internal/interp/runtime` values and the `internal/builtins` registry rather than its own forks.
- The AST evaluator (`internal/interp/evaluator`) is the only production execution path. Every fixture number in `testdata/fixtures/TEST_STATUS.md` is measured on it.

## What exists

`internal/bytecode` is about 12,900 non-test lines across 21 files (plus about 17,000 lines of tests):

| Area | Files |
|---|---|
| Compiler | `compiler_core.go`, `compiler_statements.go`, `compiler_expressions.go`, `compiler_functions.go`, `optimizer.go` |
| Bytecode format | `bytecode.go` (32-bit instructions: 8-bit opcode, 24 bits of operands; constant pools; line info), `instruction.go` |
| VM | `vm_core.go`, `vm_exec.go`, `vm_ops.go`, `vm_calls.go`, `vm_stack.go`, `runtime_error.go` |
| Forked builtins | `vm_builtins*.go`, `builtins.go` (about 2,600 lines) |
| Tooling | `disasm.go`, `serializer.go` (`.dwc` files) |

Two design choices make it a fork rather than a backend:

- It has its own tagged-union `Value`, `ObjectInstance`, and `Closure` types instead of the `internal/interp/runtime` value model.
- It reimplements builtins by hand and does not import `internal/builtins` (which registers about 241 functions).

The `.dwc` container is a small binary format: an 8-byte header (`DWC\0`, major/minor/patch, reserved) followed by the chunk name, local count, instructions, constants, line info, exception-handler table, and helper metadata. The serializer enforces semantic versioning (major must match; older minor versions load on newer VMs). Only nil, bool, int64, float64, string, function objects, and builtin references can appear in the constant pool.

## What it cannot do

Verified on 2026-09-06 with `dwscript run --bytecode -e ...`:

- `for` statements: `bytecode compile error: unsupported statement type *ast.ForStatement`.
- `case` statements: `unsupported statement type *ast.CaseStatement`.
- Classes: a `TA.Create` call fails with `unknown identifier "TA"`.
- Local function declarations: `local function declarations are not supported yet` (`compiler_statements.go:439`).
- Constructors with arguments: `constructors with arguments are not supported in bytecode yet` (`compiler_expressions.go:444`).

The compiler handles 18 of the 44 statement node types in `pkg/ast`; the evaluator's dispatch handles 35. Interfaces, virtual methods, external functions, and most of the OOP surface are absent.

## Performance

There is no verified speedup. The "5 to 6 times faster" figure that older documents quoted came from a benchmark that re-bootstrapped the whole interpreter (`runner.New(nil)`) inside the interpreter-side loop, which manufactured the ratio. A fair comparison on the subset of programs the VM can run showed roughly 2.4x, and that excludes compile time. See `docs/history/CODEBASE_REVIEW_2026-07.md`, section 4.6.

Do not cite a speed figure for the VM until it runs the fixture corpus.

## How to use it anyway

Only for experiments. Expect compile errors on ordinary programs.

```bash
dwscript run --bytecode script.dws              # execute via the VM
dwscript run --bytecode --trace script.dws      # print the disassembly while running
dwscript compile script.dws                     # write script.dwc
dwscript compile script.dws -o out.dwc          # custom output path
dwscript compile script.dws --disassemble       # show bytecode instead of writing
dwscript compile script.dws --skip-type-check   # skip semantic analysis
dwscript run script.dwc                         # run a precompiled file
```

Library API:

```go
engine, _ := dwscript.New(dwscript.WithCompileMode(dwscript.CompileModeBytecode))
```

Tests: `go test ./internal/bytecode` passes; `TestVMParity` covers the subset the VM supports.

## Public surfaces that still expose it

- `cmd/dwscript run --bytecode` (`cmd/dwscript/cmd/run.go`)
- the `dwscript compile` subcommand (`cmd/dwscript/cmd/compile.go`)
- `pkg/dwscript.CompileModeBytecode` and `WithCompileMode` (`pkg/dwscript/options.go`, `dwscript.go`)

`PLAN.md` item A11 tracks marking these as experimental in help text and godoc. They must not be presented as a production mode.

## Design history

The original design documents are archived and describe the intended instruction set and value layout, not the current state:

- [`../archive/bytecode-vm-design.md`](../archive/bytecode-vm-design.md)
- [`../archive/bytecode-vm-quick-reference.md`](../archive/bytecode-vm-quick-reference.md)
- [`../archive/bytecode-value-optimization.md`](../archive/bytecode-value-optimization.md)
