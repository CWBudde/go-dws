// Package types implements the DWScript type system.
//
// This package defines the types supported by DWScript and provides
// type checking, type compatibility checking, and type inference.
//
// Supported types:
//   - Basic types: Integer, Float, String, Boolean
//   - Compound types: Arrays, Records, Sets
//   - Object types: Classes, Interfaces
//   - Function types: Function and procedure signatures
//   - Special types: Nil, Void
//
// The type system enforces DWScript's strong static typing rules during
// semantic analysis, catching type errors before execution.
//
// Example usage:
//
//	intType := types.INTEGER
//	floatType := types.FLOAT
//	if types.IsCompatible(intType, floatType) {
//	    // can assign int to float
//	}
package types
