package types

// ============================================================================
// Type Comparison
// ============================================================================

// IsIdentical checks if two types are strictly identical.
// This is a wrapper around Type.Equals() for clarity.
func IsIdentical(a, b Type) bool {
	return a.Equals(b)
}

// ============================================================================
// Type Compatibility (Assignment Compatibility)
// ============================================================================

func IsCompatible(from, to Type) bool {
	if from == nil || to == nil {
		return false
	}

	from = GetUnderlyingType(from)
	to = GetUnderlyingType(to)

	// Identical types are always compatible
	if from.Equals(to) {
		return true
	}

	// Variant is a universal type - any type can be assigned to it
	if to.TypeKind() == "VARIANT" {
		return true
	}

	// Nil is compatible with reference types
	if from.TypeKind() == "NIL" {
		switch to.TypeKind() {
		case "NIL", "CLASS", "INTERFACE", "CLASSOF":
			return true
		default:
			return false
		}
	}

	// Integer can be implicitly converted to Float
	if from.TypeKind() == "INTEGER" && to.TypeKind() == "FLOAT" {
		return true
	}

	// Class inheritance: derived class -> base class
	if fromClass, ok := from.(*ClassType); ok {
		if toClass, ok := to.(*ClassType); ok {
			if fromClass.Equals(toClass) || isClassDescendantOf(fromClass, toClass) {
				return true
			}
		}
	}

	// Dynamic arrays are compatible with static arrays of same element type
	fromArray, fromIsArray := from.(*ArrayType)
	toArray, toIsArray := to.(*ArrayType)
	if fromIsArray && toIsArray {
		// Array types are invariant: element types must match exactly
		if !fromArray.ElementType.Equals(toArray.ElementType) {
			return false
		}

		// Allow mixing static/dynamic arrays when element types are compatible
		if fromArray.IsDynamic() || toArray.IsDynamic() {
			return true
		}

		// Static array can be assigned if bounds match
		return fromArray.Equals(toArray)
	}

	// No other implicit conversions
	return false
}

// ============================================================================
// Type Coercion Rules
// ============================================================================

// CanCoerce checks if a value of type 'from' can be implicitly converted to type 'to'.
// This is used to determine if automatic type conversion should occur.
//
// DWScript coercion rules:
//   - Integer -> Float: yes (widening conversion)
//   - Float -> Integer: no (narrowing, requires explicit conversion)
//   - String concatenation: numeric types can be implicitly converted to string in some contexts
//
// Returns true if implicit coercion is allowed, false otherwise.
func CanCoerce(from, to Type) bool {
	// Same type needs no coercion
	if from.Equals(to) {
		return true
	}

	// Integer can be coerced to Float (widening)
	if from.TypeKind() == "INTEGER" && to.TypeKind() == "FLOAT" {
		return true
	}

	// No other implicit coercions
	return false
}

// NeedsCoercion checks if a value of type 'from' requires coercion to type 'to'.
// This is used by code generators to insert conversion operations.
//
// Returns true if coercion is needed and allowed, false otherwise.
func NeedsCoercion(from, to Type) bool {
	// If types are identical, no coercion needed
	if from.Equals(to) {
		return false
	}

	// If coercion is possible, it's needed
	return CanCoerce(from, to)
}

// ============================================================================
// Type Promotion (for Binary Operations)
// ============================================================================

// PromoteTypes determines the result type for a binary operation on two types.
// This implements DWScript's type promotion rules for arithmetic and comparison operations.
//
// Rules:
//   - Integer op Integer -> Integer
//   - Float op Float -> Float
//   - Integer op Float -> Float (promote Integer to Float)
//   - Float op Integer -> Float (promote Integer to Float)
//   - String op String -> String (for concatenation)
//   - Boolean op Boolean -> Boolean (for logical operations)
//
// Returns the promoted type, or nil if the operation is invalid.
func PromoteTypes(left, right Type) Type {
	// If both types are the same, no promotion needed
	if left.Equals(right) {
		return left
	}

	leftKind := left.TypeKind()
	rightKind := right.TypeKind()

	// Numeric type promotion: Integer + Float -> Float
	if (leftKind == "INTEGER" && rightKind == "FLOAT") ||
		(leftKind == "FLOAT" && rightKind == "INTEGER") {
		return FLOAT
	}

	// No valid promotion
	return nil
}

// IsComparableType checks if values of this type can be compared with =, <>, <, >, etc.
// In DWScript:
//   - All basic types are comparable
//   - Enum types are comparable (ordinal values)
//   - Arrays and records may have limited comparison support
//   - Functions are not comparable
func IsComparableType(t Type) bool {
	switch t.TypeKind() {
	case "INTEGER", "FLOAT", "STRING", "BOOLEAN", "NIL", "ENUM", "CLASS", "INTERFACE", "CLASSOF":
		return true
	case "FUNCTION", "VOID":
		return false
	case "ARRAY", "RECORD":
		// Arrays and records have limited comparison support
		// For now, we'll say they're not comparable
		// This can be refined later based on DWScript's actual rules
		return false
	default:
		return false
	}
}

// IsOrderedType checks if values of this type support ordering comparisons (<, >, <=, >=).
// Numeric types, strings, and enums support ordering in DWScript.
func IsOrderedType(t Type) bool {
	switch t.TypeKind() {
	case "INTEGER", "FLOAT", "STRING", "ENUM":
		return true
	default:
		return false
	}
}

// isClassDescendantOf checks if child derives from parent (or is the same class).
func isClassDescendantOf(child, parent *ClassType) bool {
	if child == nil || parent == nil {
		return false
	}

	for current := child; current != nil; current = current.Parent {
		if current.Equals(parent) {
			return true
		}
	}
	return false
}

// SupportsOperation checks if a type supports a given operation.
// This is used to validate operations during semantic analysis.
func SupportsOperation(t Type, operation string) bool {
	kind := t.TypeKind()

	switch operation {
	case "+", "-", "*", "/":
		// Arithmetic operations: numeric types and string concatenation for +
		if kind == "INTEGER" || kind == "FLOAT" {
			return true
		}
		if operation == "+" && kind == "STRING" {
			return true
		}
		return false

	case "div", "mod":
		// Integer division and modulo: integers only
		return kind == "INTEGER"

	case "=", "<>":
		// Equality/inequality: most types
		return IsComparableType(t)

	case "<", ">", "<=", ">=":
		// Ordering: numeric types and strings
		return IsOrderedType(t)

	case "and", "or", "xor", "not":
		// Logical operations: booleans only
		return kind == "BOOLEAN"

	default:
		return false
	}
}

// ============================================================================
// Type Validation
// ============================================================================

// IsValidType checks if a type is valid in the current context.
// This is used to validate type annotations during parsing.
func IsValidType(t Type) bool {
	if t == nil {
		return false
	}

	// All concrete types are valid
	switch t.TypeKind() {
	case "INTEGER", "FLOAT", "STRING", "BOOLEAN", "NIL", "VOID":
		return true
	case "FUNCTION":
		// Check that function type has valid parameter and return types
		ft := t.(*FunctionType)
		for _, param := range ft.Parameters {
			if !IsValidType(param) {
				return false
			}
		}
		return IsValidType(ft.ReturnType)
	case "ARRAY":
		// Check that array element type is valid
		at := t.(*ArrayType)
		return IsValidType(at.ElementType)
	case "RECORD":
		// Check that all field types are valid
		rt := t.(*RecordType)
		for _, fieldType := range rt.Fields {
			if !IsValidType(fieldType) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
