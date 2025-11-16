// Package runtime provides the core runtime value system for the DWScript interpreter.
//
// This package is part of the Phase 3.2 refactoring to improve the interpreter architecture.
// It contains:
//   - Value type definitions (primitives, composites, references)
//   - Value interfaces for type-safe operations (NumericValue, ComparableValue, etc.)
//   - Value creation helpers and utilities
//   - Object pooling for performance optimization
//
// The package is organized into:
//   - value_interfaces.go: Interface definitions for value operations
//   - primitives.go: Basic value types (Integer, Float, String, Boolean, Nil)
//   - composite.go: Composite types (Array, Record, Set) - TODO
//   - object.go: Object types (Class instances, Interfaces) - TODO
//   - function.go: Callable types (Function pointers, Lambdas) - TODO
//   - special.go: Special types (Error, Exception, Variant, TypeInfo) - TODO
//   - pool.go: Object pooling for frequently allocated types
//
// Design Goals:
//   - Type safety through interfaces instead of type assertions
//   - Clear separation between value types and reference types
//   - Reduced allocations via object pooling
//   - Easy to test components in isolation
//
// Migration Status:
//
//	Phase 3.2.1: Value interfaces defined ✓
//	Phase 3.2.2: Primitives moved to runtime/ ✓
//	Phase 3.2.3: Object pooling ✓
//
// For more details, see docs/architecture/interpreter-refactoring.md
package runtime
