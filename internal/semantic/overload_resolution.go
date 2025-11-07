package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
)

// SignaturesEqual checks if two function signatures are identical for overload detection.
// Two signatures are equal if they have:
// - Same parameter count (excluding optional parameters with defaults)
// - Same parameter types
// - Same parameter modifiers (var/const/lazy)
// - Same variadic status
//
// Return type is NOT considered for signature equality (DWScript allows same signature
// with different return types if marked with overload directive).
//
// This is used to detect duplicate overload declarations and ensure each overload
// has a unique signature.
func SignaturesEqual(sig1, sig2 *types.FunctionType) bool {
	// Check parameter count
	if len(sig1.Parameters) != len(sig2.Parameters) {
		return false
	}

	// Check variadic status
	if sig1.IsVariadic != sig2.IsVariadic {
		return false
	}

	// Check variadic element type if both are variadic
	if sig1.IsVariadic {
		if sig1.VariadicType == nil || sig2.VariadicType == nil {
			if sig1.VariadicType != sig2.VariadicType {
				return false
			}
		} else if !sig1.VariadicType.Equals(sig2.VariadicType) {
			return false
		}
	}

	// Check each parameter type and modifiers
	for i := 0; i < len(sig1.Parameters); i++ {
		// Check parameter type
		if !sig1.Parameters[i].Equals(sig2.Parameters[i]) {
			return false
		}

		// Check parameter modifiers (var/const/lazy)
		// NOTE: This is the fix for the limitation documented in overload_test.go:440-475
		if i < len(sig1.VarParams) && i < len(sig2.VarParams) {
			if sig1.VarParams[i] != sig2.VarParams[i] {
				return false
			}
		}
		if i < len(sig1.ConstParams) && i < len(sig2.ConstParams) {
			if sig1.ConstParams[i] != sig2.ConstParams[i] {
				return false
			}
		}
		if i < len(sig1.LazyParams) && i < len(sig2.LazyParams) {
			if sig1.LazyParams[i] != sig2.LazyParams[i] {
				return false
			}
		}
	}

	return true
}

// SignatureDistance calculates the "distance" between argument types and parameter types.
// Lower distance means better match. Returns -1 if incompatible.
//
// Distance levels:
//
//	 0 = Exact match (same type)
//	 1 = Implicit conversion (Integer -> Float, derived class -> base class)
//	 2 = Variant conversion (any type -> Variant, Variant -> any type)
//	-1 = Incompatible (no conversion possible)
//
// This is used to rank overload candidates when multiple overloads could match
// the provided arguments.
func SignatureDistance(argTypes []types.Type, signature *types.FunctionType) int {
	// Handle variadic functions: minimum number of parameters is len(Parameters)-1
	minParams := len(signature.Parameters)
	if signature.IsVariadic {
		minParams = len(signature.Parameters) - 1
	}

	// Check argument count compatibility
	if len(argTypes) < minParams {
		return -1 // Too few arguments
	}

	// Check if too many arguments for non-variadic function
	if !signature.IsVariadic && len(argTypes) > len(signature.Parameters) {
		return -1 // Too many arguments
	}

	totalDistance := 0

	// Calculate distance for each argument
	for i, argType := range argTypes {
		var paramType types.Type

		// For variadic functions, arguments beyond the last parameter use variadic element type
		if signature.IsVariadic && i >= len(signature.Parameters)-1 {
			paramType = signature.VariadicType
		} else if i < len(signature.Parameters) {
			paramType = signature.Parameters[i]
		} else {
			// Too many arguments (shouldn't happen due to check above)
			return -1
		}

		// Calculate distance for this argument-parameter pair
		dist := typeDistance(argType, paramType)
		if dist < 0 {
			return -1 // Incompatible
		}
		totalDistance += dist
	}

	return totalDistance
}

// typeDistance calculates the conversion distance between two types.
// Returns -1 if no conversion is possible.
func typeDistance(from, to types.Type) int {
	// Exact match
	if from.Equals(to) {
		return 0
	}

	fromKind := from.TypeKind()
	toKind := to.TypeKind()

	// Integer -> Float conversion (common implicit conversion)
	if fromKind == "INTEGER" && toKind == "FLOAT" {
		return 1
	}

	// Class inheritance: derived class -> base class
	// TODO: Implement class hierarchy checking when class types are available
	// For now, we only support exact matches and basic conversions

	// Variant conversions: any type <-> Variant
	if toKind == "VARIANT" {
		return 2 // Any type can be converted to Variant
	}
	if fromKind == "VARIANT" {
		return 2 // Variant can be converted to any type (runtime checked)
	}

	// String conversions (if needed)
	// DWScript allows implicit string concatenation and conversions in some contexts
	// For now, we require explicit conversions

	// No conversion available
	return -1
}

// ResolveOverload selects the best-fit overload from a set of candidates.
//
// Parameters:
//   - candidates: List of function symbols representing overloaded functions
//   - argTypes: Actual argument types from the call site
//
// Returns:
//   - The best matching function symbol
//   - Error if no match found or ambiguous (multiple equally good matches)
//
// Algorithm:
//  1. Filter candidates by signature compatibility (SignatureDistance >= 0)
//  2. Find candidate(s) with minimum distance
//  3. If exactly one candidate has minimum distance, return it
//  4. If multiple candidates tie for minimum distance, return ambiguity error
//  5. If no compatible candidates, return no match error
func ResolveOverload(candidates []*Symbol, argTypes []types.Type) (*Symbol, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no overload candidates provided")
	}

	// Single candidate - no resolution needed
	if len(candidates) == 1 {
		// Still need to verify compatibility
		funcType, ok := candidates[0].Type.(*types.FunctionType)
		if !ok {
			return nil, fmt.Errorf("candidate is not a function type")
		}
		dist := SignatureDistance(argTypes, funcType)
		if dist < 0 {
			return nil, fmt.Errorf("no matching overload for argument types")
		}
		return candidates[0], nil
	}

	// Calculate distance for each candidate
	type candidateWithDistance struct {
		symbol   *Symbol
		distance int
	}
	var compatible []candidateWithDistance

	for _, candidate := range candidates {
		funcType, ok := candidate.Type.(*types.FunctionType)
		if !ok {
			continue // Skip non-function types
		}

		dist := SignatureDistance(argTypes, funcType)
		if dist >= 0 {
			compatible = append(compatible, candidateWithDistance{
				symbol:   candidate,
				distance: dist,
			})
		}
	}

	// No compatible overloads
	if len(compatible) == 0 {
		return nil, fmt.Errorf("no matching overload for argument types: %s", formatArgTypes(argTypes))
	}

	// Find minimum distance
	minDist := compatible[0].distance
	for _, c := range compatible[1:] {
		if c.distance < minDist {
			minDist = c.distance
		}
	}

	// Collect all candidates with minimum distance
	var bestMatches []*Symbol
	for _, c := range compatible {
		if c.distance == minDist {
			bestMatches = append(bestMatches, c.symbol)
		}
	}

	// Exactly one best match
	if len(bestMatches) == 1 {
		return bestMatches[0], nil
	}

	// Ambiguous: multiple equally good matches
	return nil, fmt.Errorf("ambiguous overload call: %d candidates with equal distance %d for argument types: %s",
		len(bestMatches), minDist, formatArgTypes(argTypes))
}

// formatArgTypes formats argument types for error messages
func formatArgTypes(argTypes []types.Type) string {
	if len(argTypes) == 0 {
		return "()"
	}

	result := "("
	for i, t := range argTypes {
		if i > 0 {
			result += ", "
		}
		result += t.String()
	}
	result += ")"
	return result
}
