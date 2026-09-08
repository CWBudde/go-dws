// Package runtime provides the core runtime value system for the DWScript interpreter.
// It contains:
//   - Value type definitions (primitives, composites, references)
//   - Value interfaces for type-safe operations (NumericValue, ComparableValue, etc.)
//   - Value creation helpers and utilities
//   - Object pooling for performance optimization
//
// Values and execution state are grouped by responsibility, including primitives,
// arrays, records, objects, callable values and execution contexts.
// Shared metadata, environments, call stacks and reference counting live here
// alongside the value containers used by the evaluator.
//
// Design Goals:
//   - Type safety through interfaces instead of type assertions
//   - Clear separation between value types and reference types
//   - Reduced allocations via object pooling
//   - Easy to test components in isolation
//
// For more details, see docs/architecture/interp-evaluator-steady-state.md
package runtime
