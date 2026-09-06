# docs/archive

Superseded design notes, audits, migration guides, and measurement snapshots. They were moved here in 2026-09 because the code they describe has changed, the numbers are stale, or the work they planned is finished or parked. Nothing here is maintained, and links inside these files are not kept working.

Why keep them at all: they hold the reasoning behind decisions (adapter removal, EvalNode elimination, builtin registry design, visitor generation, bytecode VM design) that is still useful when revisiting an area.

**Task and stage IDs** in these files (`Task 3.5.38`, `Phase 4.10.2`, `Stage 11`, …) refer to the pre-2026-07 `PLAN.md`, which exists only in git history (`git show <sha>:PLAN.md`).

Parked backlogs kept here on purpose, gated on non-host fixture categories reaching ≥ 80% (`PLAN.md` §5): `CodeGenTODO.md`, `CodeGenJSGoal.md`, `bytecode-vm-design.md`, `bytecode-vm-quick-reference.md`, `bytecode-value-optimization.md`.

Still-open questions referenced from `PLAN.md`: `phase-4.12.1-lifetime-inventory.md` (two competing scope/binding lifetime models, see PLAN.md A6), `semantic-legacy-hotspots-5.3.10.md` (raw `addError` sites, see PLAN.md F8), `failure-scripts-next-phase-plan.md` (its work families were folded into PLAN.md §4).

Current open work: `PLAN.md`. Documentation index: `docs/README.md`.

## Files

- [`adapter-consolidation-plan.md`](adapter-consolidation-plan.md) — InterpreterAdapter Consolidation Plan
- [`adapter-method-audit.md`](adapter-method-audit.md) — Adapter Method Usage Audit (Task 3.4.1)
- [`arrays-type-compatibility-research.md`](arrays-type-compatibility-research.md) — Array Type Compatibility Research
- [`arrayvalue-migration-analysis.md`](arrayvalue-migration-analysis.md) — ArrayValue Migration Analysis
- [`benchmark_baseline_phase3.txt`](benchmark_baseline_phase3.txt) — benchmark_baseline_phase3.txt
- [`benchmark_summary.md`](benchmark_summary.md) — Benchmark Summary - Phase 3.1
- [`builtin-migration-roadmap.md`](builtin-migration-roadmap.md) — Built-in Function Migration Roadmap
- [`builtin-registry-summary.md`](builtin-registry-summary.md) — Built-in Function Registry Summary
- [`builtin-registry-proposal.md`](builtin-registry-proposal.md) — Builtin Function Registry Patterns
- [`bytecode-value-optimization.md`](bytecode-value-optimization.md) — Bytecode Value Optimization Design
- [`bytecode-vm-design.md`](bytecode-vm-design.md) — Bytecode VM Instruction Set Design Research
- [`bytecode-vm-quick-reference.md`](bytecode-vm-quick-reference.md) — Bytecode VM Quick Reference
- [`CodeGenJSGoal.md`](CodeGenJSGoal.md) — Implementing a JSCodeGen for DWScript in Go (go-dws)
- [`CodeGenTODO.md`](CodeGenTODO.md) — Stage 11: Code Generation - Multi-Backend Architecture
- [`dual-mode-parser.md`](dual-mode-parser.md) — Dual-Mode Parser Architecture
- [`environment-audit.md`](environment-audit.md) — Environment Dual System Audit
- [`environment-migration-guide.md`](environment-migration-guide.md) — Environment Synchronization Migration Guide
- [`evalnode-audit-final.md`](evalnode-audit-final.md) — EvalNode Calls Audit - Final Report
- [`evalnode-audit.md`](evalnode-audit.md) — EvalNode Calls Audit - Post-Consolidation Report
- [`evaluator-architecture.md`](evaluator-architecture.md) — Evaluator Architecture: Separation of Concerns
- [`evaluator.md`](evaluator.md) — Evaluator Architecture - Consolidated Summary
- [`failure-scripts-5.4.1.md`](failure-scripts-5.4.1.md) — FailureScripts Classification (5.4.1)
- [`failure-scripts-5.4.4.md`](failure-scripts-5.4.4.md) — FailureScripts Delta After 5.4.2.x / 5.4.3 (5.4.4)
- [`failure-scripts-next-phase-plan.md`](failure-scripts-next-phase-plan.md) — FailureScripts Next-Phase Plan
- [`feature-matrix.md`](feature-matrix.md) — DWScript vs go-dws Feature Comparison Matrix
- [`FIXTURE_FAILURES_ANALYSIS.md`](FIXTURE_FAILURES_ANALYSIS.md) — SimpleScripts Test Failures Analysis
- [`FIXTURE_FAILURES_BY_CAUSE.md`](FIXTURE_FAILURES_BY_CAUSE.md) — Test Failures Grouped by Implementation Cause
- [`implemented-features.md`](implemented-features.md) — go-dws Implemented Features Catalog
- [`interface-usage-audit.md`](interface-usage-audit.md) — Interface{} Usage Audit Report
- [`interpreter.md`](interpreter.md) — Interpreter Architecture
- [`interpreter-refactoring.md`](interpreter-refactoring.md) — Interpreter Architecture Refactoring Design
- [`method_overload_audit.md`](method_overload_audit.md) — Method Overload Implementation Audit (Task 9.20.1)
- [`missing-features-recommendations.md`](missing-features-recommendations.md) — Missing Features: Recommendations and Roadmap
- [`overloadspass_test_results.md`](overloadspass_test_results.md) — OverloadsPass Test Results - Phase 9 Stage 6
- [`panic_fix_overload_func_ptr_param.md`](panic_fix_overload_func_ptr_param.md) — Panic Fix: overload_func_ptr_param.pas
- [`parse-expression-migration.md`](parse-expression-migration.md) — parseExpression Migration Design Document
- [`phase3-adapter-strategy.md`](phase3-adapter-strategy.md) — Phase 3 Adapter Strategy: General Facilities vs Operation-Specific Interfaces
- [`phase-4.10.1-live-interp-execution-inventory.md`](phase-4.10.1-live-interp-execution-inventory.md) — Phase 4.10.1 Live Interp Execution Inventory
- [`phase-4.10.2-seam-decisions.md`](phase-4.10.2-seam-decisions.md) — Phase 4.10.2 Seam Decisions
- [`phase-4.10.5-interp-allowed-responsibilities.md`](phase-4.10.5-interp-allowed-responsibilities.md) — Phase 4.10.5 Allowed `internal/interp` Responsibilities
- [`phase-4.11-neutral-boundary-audit.md`](phase-4.11-neutral-boundary-audit.md) — Phase 4.11 Neutral Boundary Audit
- [`phase-4.12.1-lifetime-inventory.md`](phase-4.12.1-lifetime-inventory.md) — Phase 4.12.1 Lifetime Inventory
- [`phase-4.1-boundary.md`](phase-4.1-boundary.md) — Phase 4.1 Boundary
- [`phase-4.9.1-interpreter-execution-inventory.md`](phase-4.9.1-interpreter-execution-inventory.md) — Phase 4.9.1 Interpreter Execution Inventory
- [`plan_md_updates_summary.md`](plan_md_updates_summary.md) — PLAN.md Updates Summary - Phase 9 Overloading
- [`semantic-legacy-hotspots-5.3.10.md`](semantic-legacy-hotspots-5.3.10.md) — Semantic Legacy Hotspot Audit (5.3.10)
- [`test-coverage.md`](test-coverage.md) — Test Coverage Report
- [`visitor-benchmark-results.md`](visitor-benchmark-results.md) — Visitor Pattern Performance Benchmarks
- [`visitor-compatibility-test-results.md`](visitor-compatibility-test-results.md) — Visitor Compatibility Test Results (Task 9.17.6)
- [`visitor-reflection-research.md`](visitor-reflection-research.md) — Reflection-Based Visitor Pattern Research (Task 9.17.1)
