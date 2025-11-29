package types

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ident"
)

// Type represents a DWScript type at compile-time.
// This is used for static type checking during semantic analysis,
// separate from runtime Value representations.
type Type interface {
	// String returns the string representation of the type (e.g., "Integer", "String")
	String() string

	// Equals checks if two types are identical
	Equals(other Type) bool

	// TypeKind returns a unique identifier for the type kind
	// Used for type discrimination without reflection
	TypeKind() string
}

// ============================================================================
// Basic Types
// ============================================================================

// IntegerType represents the Integer type
type IntegerType struct{}

func (t *IntegerType) String() string   { return "Integer" }
func (t *IntegerType) TypeKind() string { return "INTEGER" }
func (t *IntegerType) Equals(other Type) bool {
	// Resolve type aliases before comparison
	other = GetUnderlyingType(other)
	_, ok := other.(*IntegerType)
	return ok
}

// FloatType represents the Float type
type FloatType struct{}

func (t *FloatType) String() string   { return "Float" }
func (t *FloatType) TypeKind() string { return "FLOAT" }
func (t *FloatType) Equals(other Type) bool {
	// Resolve type aliases before comparison
	other = GetUnderlyingType(other)
	_, ok := other.(*FloatType)
	return ok
}

// StringType represents the String type
type StringType struct{}

func (t *StringType) String() string   { return "String" }
func (t *StringType) TypeKind() string { return "STRING" }
func (t *StringType) Equals(other Type) bool {
	// Resolve type aliases before comparison
	other = GetUnderlyingType(other)
	_, ok := other.(*StringType)
	return ok
}

// BooleanType represents the Boolean type
type BooleanType struct{}

func (t *BooleanType) String() string   { return "Boolean" }
func (t *BooleanType) TypeKind() string { return "BOOLEAN" }
func (t *BooleanType) Equals(other Type) bool {
	// Resolve type aliases before comparison
	other = GetUnderlyingType(other)
	_, ok := other.(*BooleanType)
	return ok
}

// DateTimeType represents the TDateTime type
// TDateTime is internally represented as a Float where:
// - Integer part = number of days since December 30, 1899
// - Fractional part = time of day (0.5 = noon, 0.25 = 6am)
// This matches Delphi's TDateTime representation exactly.
type DateTimeType struct{}

func (t *DateTimeType) String() string   { return "TDateTime" }
func (t *DateTimeType) TypeKind() string { return "DATETIME" }
func (t *DateTimeType) Equals(other Type) bool {
	// Resolve type aliases before comparison
	other = GetUnderlyingType(other)
	_, ok := other.(*DateTimeType)
	return ok
}

// NilType represents the nil/null type
type NilType struct{}

func (t *NilType) String() string   { return "Nil" }
func (t *NilType) TypeKind() string { return "NIL" }
func (t *NilType) Equals(other Type) bool {
	// Resolve type aliases before comparison
	other = GetUnderlyingType(other)
	_, ok := other.(*NilType)
	return ok
}

// VoidType represents the void type (for procedures with no return value)
type VoidType struct{}

func (t *VoidType) String() string   { return "Void" }
func (t *VoidType) TypeKind() string { return "VOID" }
func (t *VoidType) Equals(other Type) bool {
	// Resolve type aliases before comparison
	other = GetUnderlyingType(other)
	_, ok := other.(*VoidType)
	return ok
}

// ConstType represents the "const" type used in "array of const" parameters
// Similar to Pascal's "const" or Variant type - can hold any value
// This is used specifically for builtin functions like Format that accept heterogeneous arrays
// DEPRECATED: This is a temporary workaround. Use VariantType instead
type ConstType struct{}

func (t *ConstType) String() string   { return "Const" }
func (t *ConstType) TypeKind() string { return "CONST" }
func (t *ConstType) Equals(other Type) bool {
	// Resolve type aliases before comparison
	other = GetUnderlyingType(other)
	_, ok := other.(*ConstType)
	return ok
}

// VariantType represents the Variant type in DWScript.
// Variant is DWScript's dynamic type that can hold any value at runtime.
//
// The Variant type provides:
// - Dynamic, heterogeneous value storage with runtime type tracking
// - Automatic type conversions between compatible types (Integer ↔ Float, numeric ↔ String)
// - Support for all basic operations (arithmetic, comparison, concatenation)
// - Used in "array of const" parameters for variadic-style functions
// - Compatible with all DWScript types (Integer, Float, String, Boolean, objects, arrays, records)
//
// Unlike the temporary ConstType workaround, VariantType follows DWScript/Delphi semantics:
// - Variables can be declared as Variant: var v: Variant;
// - Any value can be assigned to a Variant (implicit boxing)
// - Variants can be assigned to typed variables with runtime checking (explicit unboxing)
// - Operations on Variants use runtime type information for type coercion
//
// Example usage:
//
//	var v: Variant;
//	v := 42;          // Stores Integer
//	v := 'hello';     // Stores String
//	v := 3.14;        // Stores Float
//	var i: Integer := v;  // Runtime type check and unboxing
//
// Similar to Delphi's TVarData and DWScript's TdwsVariant.
// See reference/dwscript-original/Source/dwsVariantFunctions.pas
type VariantType struct{}

func (t *VariantType) String() string   { return "Variant" }
func (t *VariantType) TypeKind() string { return "VARIANT" }
func (t *VariantType) Equals(other Type) bool {
	// Resolve type aliases before comparison
	other = GetUnderlyingType(other)
	_, ok := other.(*VariantType)
	return ok
}

// ============================================================================
// Singleton Type Constants
// ============================================================================

// Singleton instances of basic types
// These are used throughout the compiler for type checking
var (
	INTEGER  = &IntegerType{}
	FLOAT    = &FloatType{}
	STRING   = &StringType{}
	BOOLEAN  = &BooleanType{}
	DATETIME = &DateTimeType{}
	NIL      = &NilType{}
	VOID     = &VoidType{}
	CONST    = &ConstType{}
	VARIANT  = &VariantType{}
)

// ARRAY_OF_CONST is a special array type used for builtin functions like Format
// that accept heterogeneous arrays (array of const in Pascal)
var ARRAY_OF_CONST = NewDynamicArrayType(VARIANT)

// IINTERFACE is the base interface type (like IUnknown in COM)
// All interfaces can inherit from this root interface.
var IINTERFACE = &InterfaceType{
	Name:         "IInterface",
	Parent:       nil,
	Methods:      make(map[string]*FunctionType),
	IsExternal:   false,
	ExternalName: "",
}

// ============================================================================
// Type Aliases
// ============================================================================

// TypeAlias represents a type alias declaration.
// Type aliases create alternate names for existing types, improving code clarity
// and enabling domain-specific naming.
//
// Example: type TUserID = Integer; type TFileName = String;
//
// Type aliases are transparent - they are treated identically to their underlying type
// for type checking purposes. The alias name is preserved only for error messages
// and debugging.
type TypeAlias struct {
	AliasedType Type
	Name        string
}

// String returns the alias name (not the underlying type name).
// This preserves the alias name in error messages and debugging output.
func (t *TypeAlias) String() string {
	return t.Name
}

// TypeKind returns the underlying type's kind.
// For nested aliases (alias of an alias), this recursively resolves to the
// ultimate underlying type's kind.
//
// Example: type A = Integer; type B = A;
// Both A.TypeKind() and B.TypeKind() return "INTEGER"
func (t *TypeAlias) TypeKind() string {
	return GetUnderlyingType(t).TypeKind()
}

// Equals checks if two types are equal by comparing their underlying types.
// Type aliases are transparent for equality checking - an alias equals its
// underlying type and other aliases to the same type.
//
// Examples:
//   - type MyInt = Integer; MyInt.Equals(INTEGER) → true
//   - type A = Integer; type B = Integer; A.Equals(B) → true
//   - type MyInt = Integer; MyInt.Equals(STRING) → false
func (t *TypeAlias) Equals(other Type) bool {
	return GetUnderlyingType(t).Equals(GetUnderlyingType(other))
}

// GetUnderlyingType recursively resolves type aliases to find the ultimate underlying type.
// This handles nested aliases like: type A = Integer; type B = A; type C = B;
//
// For non-alias types, this simply returns the type itself.
// For alias types, this follows the chain until reaching a non-alias type.
func GetUnderlyingType(t Type) Type {
	for alias, ok := t.(*TypeAlias); ok; alias, ok = t.(*TypeAlias) {
		t = alias.AliasedType
	}
	return t
}

// ============================================================================
// Subrange Types
// ============================================================================

// SubrangeType represents a subrange type declaration.
// Subrange types restrict an ordinal type to a specific range of values,
// providing type safety and runtime validation.
//
// Example: type TDigit = 0..9; type TPercent = 0..100;
//
// Unlike type aliases, subranges are NOT transparent - they maintain their
// bounds and require validation at assignment time.
type SubrangeType struct {
	BaseType  Type   // Underlying ordinal type (Integer, Char, or enum)
	Name      string // Subrange type name (e.g., "TDigit")
	LowBound  int    // Inclusive lower bound
	HighBound int    // Inclusive upper bound
}

// String returns the range representation "LowBound..HighBound".
func (s *SubrangeType) String() string {
	return fmt.Sprintf("%d..%d", s.LowBound, s.HighBound)
}

// TypeKind returns "SUBRANGE" for subrange types.
func (s *SubrangeType) TypeKind() string {
	return "SUBRANGE"
}

// Equals checks if two subrange types are equal.
// Subrange types are equal if they have the same base type and the same bounds.
func (s *SubrangeType) Equals(other Type) bool {
	otherSubrange, ok := other.(*SubrangeType)
	if !ok {
		return false
	}
	return s.BaseType.Equals(otherSubrange.BaseType) &&
		s.LowBound == otherSubrange.LowBound &&
		s.HighBound == otherSubrange.HighBound
}

// Contains checks if a value is within the subrange bounds (inclusive).
func (s *SubrangeType) Contains(value int) bool {
	return value >= s.LowBound && value <= s.HighBound
}

// ValidateRange validates that a value is within the subrange bounds.
// Returns an error if the value is outside the allowed range.
func ValidateRange(value int, subrange *SubrangeType) error {
	if !subrange.Contains(value) {
		return fmt.Errorf("value %d is out of range for type %s (%d..%d)",
			value, subrange.Name, subrange.LowBound, subrange.HighBound)
	}
	return nil
}

// ============================================================================
// Type Utilities
// ============================================================================

// IsBasicType checks if a type is one of the basic types
func IsBasicType(t Type) bool {
	switch t.TypeKind() {
	case "INTEGER", "FLOAT", "STRING", "BOOLEAN":
		return true
	default:
		return false
	}
}

// IsNumericType checks if a type is numeric (Integer or Float)
func IsNumericType(t Type) bool {
	kind := t.TypeKind()
	return kind == "INTEGER" || kind == "FLOAT"
}

// IsOrdinalType checks if a type is an ordinal type.
// In DWScript, ordinal types are types with a well-defined ordering and can be used in:
//   - for loops (Integer, Boolean, Enum, Subrange)
//   - set elements (Integer, String/Char, Enum, Subrange)
//   - ranges (e.g., [1..10], ['a'..'z'])
func IsOrdinalType(t Type) bool {
	// Resolve type aliases before checking
	t = GetUnderlyingType(t)

	kind := t.TypeKind()
	return kind == "INTEGER" || kind == "BOOLEAN" || kind == "ENUM" ||
		kind == "STRING" || kind == "SUBRANGE"
}

// OrdinalBounds returns the lowest and highest ordinal values for bounded ordinal types.
// Supported types:
//   - Boolean: 0..1
//   - EnumType: MinOrdinal()..MaxOrdinal()
//   - SubrangeType: LowBound..HighBound
//
// Returns ok=false for types without finite bounds (e.g., Integer, String).
func OrdinalBounds(t Type) (low, high int, ok bool) {
	// Resolve aliases before inspecting the concrete type
	t = GetUnderlyingType(t)

	switch v := t.(type) {
	case *BooleanType:
		return 0, 1, true
	case *EnumType:
		return v.MinOrdinal(), v.MaxOrdinal(), true
	case *SubrangeType:
		return v.LowBound, v.HighBound, true
	default:
		return 0, 0, false
	}
}

// TypeFromString converts a type name string to a Type
// This is useful for parsing type annotations
// DWScript is case-insensitive, so this function normalizes the input
func TypeFromString(name string) (Type, error) {
	switch ident.Normalize(name) {
	case "integer":
		return INTEGER, nil
	case "float":
		return FLOAT, nil
	case "string":
		return STRING, nil
	case "boolean":
		return BOOLEAN, nil
	case "tdatetime":
		return DATETIME, nil
	case "void":
		return VOID, nil
	case "variant":
		return VARIANT, nil
	default:
		return nil, fmt.Errorf("unknown type: %s", name)
	}
}

// ============================================================================
// Object-Oriented Types (Stage 7)
// ============================================================================

// PropAccessKind represents how a property is accessed (read or write).
type PropAccessKind int

const (
	PropAccessNone       PropAccessKind = iota // No access (write-only has ReadKind=None, read-only has WriteKind=None)
	PropAccessField                            // Direct field access (e.g., FName)
	PropAccessMethod                           // Method call (e.g., GetName, SetName)
	PropAccessExpression                       // Expression-based getter (e.g., (FValue * 2))
	PropAccessBuiltin                          // Built-in property implemented in Go (e.g., array .Length, .High, .Low)
)

// PropertyInfo represents property metadata for a class.
// Fields: Name, Type, ReadSpec, WriteSpec, IsIndexed, IsDefault
// Properties provide syntactic sugar for getter/setter access.
type PropertyInfo struct {
	Type            Type
	ReadExpr        any
	Name            string
	ReadSpec        string
	WriteSpec       string
	ReadKind        PropAccessKind
	WriteKind       PropAccessKind
	HasIndexValue   bool
	IndexValue      interface{}
	IndexValueType  Type
	IsIndexed       bool
	IsDefault       bool
	IsClassProperty bool
}

// MethodInfo stores metadata about a single method or overload
// This allows tracking virtual/override/abstract/overload per method signature
type MethodInfo struct {
	Signature            *FunctionType
	IsVirtual            bool
	IsOverride           bool
	IsAbstract           bool
	IsReintroduce        bool
	IsForwarded          bool
	IsClassMethod        bool
	HasOverloadDirective bool
	Visibility           int
}

// ClassType represents a class type in DWScript.
// Classes support inheritance, fields, methods, and class variables (static fields).
type ClassType struct {
	OverrideMethods      map[string]bool
	AbstractMethods      map[string]bool
	ForwardedMethods     map[string]bool
	ReintroduceMethods   map[string]bool // Track methods marked with 'reintroduce'
	Fields               map[string]Type
	ClassVars            map[string]Type
	Constants            map[string]interface{}   // Class constants
	ConstantTypes        map[string]Type          // Types for class constants
	ConstantVisibility   map[string]int           // Visibility for class constants
	ClassVarVisibility   map[string]int           // Visibility for class variables
	Methods              map[string]*FunctionType // Primary method signature (first or only overload)
	MethodOverloads      map[string][]*MethodInfo // All overload variants
	FieldVisibility      map[string]int
	MethodVisibility     map[string]int
	VirtualMethods       map[string]bool
	Parent               *ClassType
	Properties           map[string]*PropertyInfo
	ClassMethodFlags     map[string]bool
	Constructors         map[string]*FunctionType // Primary constructor signature
	ConstructorOverloads map[string][]*MethodInfo // All constructor overload variants
	DefaultConstructor   string                   // Name of the constructor marked as 'default' (empty if none)
	Operators            *OperatorRegistry
	ExternalName         string
	Name                 string
	Interfaces           []*InterfaceType
	IsAbstract           bool
	IsExternal           bool
	IsForward            bool // True if this is a forward declaration only
	IsPartial            bool // True if this is a partial class
}

// String returns the string representation of the class type
func (ct *ClassType) String() string {
	if ct.Parent != nil {
		return fmt.Sprintf("%s(%s)", ct.Name, ct.Parent.Name)
	}
	return ct.Name
}

// TypeKind returns "CLASS" for class types
func (ct *ClassType) TypeKind() string {
	return "CLASS"
}

// Equals checks if two class types are equal.
// Class types are equal if they have the same name.
// Note: We use nominal typing (name-based) rather than structural typing.
func (ct *ClassType) Equals(other Type) bool {
	otherClass, ok := other.(*ClassType)
	if !ok {
		return false
	}
	return ct.Name == otherClass.Name
}

// HasField checks if the class or any of its ancestors has a field with the given name
func (ct *ClassType) HasField(name string) bool {
	// Case-insensitive field lookup
	normalizedName := ident.Normalize(name)
	if _, ok := ct.Fields[normalizedName]; ok {
		return true
	}
	if ct.Parent != nil {
		return ct.Parent.HasField(name)
	}
	return false
}

// GetField retrieves the type of a field by name, searching up the inheritance chain
func (ct *ClassType) GetField(name string) (Type, bool) {
	// Case-insensitive field lookup
	normalizedName := ident.Normalize(name)
	if fieldType, ok := ct.Fields[normalizedName]; ok {
		return fieldType, true
	}
	if ct.Parent != nil {
		return ct.Parent.GetField(name)
	}
	return nil, false
}

// HasMethod checks if the class or any of its ancestors has a method with the given name
func (ct *ClassType) HasMethod(name string) bool {
	methodName := ident.Normalize(name)
	if len(ct.MethodOverloads[methodName]) > 0 {
		return true
	}
	if ct.Parent != nil {
		return ct.Parent.HasMethod(methodName)
	}
	return false
}

// GetMethod retrieves the signature of a method by name, searching up the inheritance chain
func (ct *ClassType) GetMethod(name string) (*FunctionType, bool) {
	methodName := ident.Normalize(name)
	if overloads := ct.MethodOverloads[methodName]; len(overloads) > 0 {
		// Return the signature of the first overload
		// (For non-overloaded methods, there's only one)
		return overloads[0].Signature, true
	}
	if ct.Parent != nil {
		return ct.Parent.GetMethod(methodName)
	}
	return nil, false
}

// GetMethodOverloads retrieves all overload variants of a method by name
// Returns overloads from this class only (does not search parent)
func (ct *ClassType) GetMethodOverloads(name string) []*MethodInfo {
	// Normalize method names for case-insensitive lookup
	return ct.MethodOverloads[ident.Normalize(name)]
}

// GetConstructorOverloads retrieves all overload variants of a constructor by name
// Constructor names are case-insensitive, so we normalize for lookup
func (ct *ClassType) GetConstructorOverloads(name string) []*MethodInfo {
	return ct.ConstructorOverloads[ident.Normalize(name)]
}

// AddMethodOverload adds a method overload to the class
func (ct *ClassType) AddMethodOverload(name string, info *MethodInfo) {
	// Normalize method names for case-insensitive lookup
	normalizedName := ident.Normalize(name)
	ct.MethodOverloads[normalizedName] = append(ct.MethodOverloads[normalizedName], info)

	// Update the primary Methods map to point to the first overload
	// This maintains backward compatibility with code that uses Methods map directly
	if len(ct.MethodOverloads[normalizedName]) == 1 {
		ct.Methods[normalizedName] = info.Signature
	}
}

// AddConstructorOverload adds a constructor overload to the class
// Constructor names are case-insensitive, so we normalize for lookup
func (ct *ClassType) AddConstructorOverload(name string, info *MethodInfo) {
	normalizedName := ident.Normalize(name)
	ct.ConstructorOverloads[normalizedName] = append(ct.ConstructorOverloads[normalizedName], info)

	// Update the primary Constructors map
	if len(ct.ConstructorOverloads[normalizedName]) == 1 {
		ct.Constructors[normalizedName] = info.Signature
	}
}

// RegisterOperator adds a class operator overload to the class type.
func (ct *ClassType) RegisterOperator(signature *OperatorSignature) error {
	if ct.Operators == nil {
		ct.Operators = NewOperatorRegistry()
	}
	return ct.Operators.Register(signature)
}

// LookupOperator searches for a matching operator overload in the class hierarchy.
func (ct *ClassType) LookupOperator(operator string, operandTypes []Type) (*OperatorSignature, bool) {
	if ct == nil {
		return nil, false
	}
	if ct.Operators != nil {
		if sig, ok := ct.Operators.Lookup(operator, operandTypes); ok {
			return sig, true
		}
	}
	if ct.Parent != nil {
		return ct.Parent.LookupOperator(operator, operandTypes)
	}
	return nil, false
}

// HasConstructor checks if the class or any ancestor declares a constructor with the given name.
// Case-insensitive constructor name lookup
func (ct *ClassType) HasConstructor(name string) bool {
	if ct == nil {
		return false
	}
	// Case-insensitive lookup
	for ctorName := range ct.Constructors {
		if ident.Equal(ctorName, name) {
			return true
		}
	}
	if ct.Parent != nil {
		return ct.Parent.HasConstructor(name)
	}
	return false
}

// GetConstructor retrieves the signature of a constructor by name, searching up the inheritance chain.
// Returns (functionType, true) if found, or (nil, false) if not found.
// Supports inherited constructor calls in semantic analyzer
func (ct *ClassType) GetConstructor(name string) (*FunctionType, bool) {
	if ct == nil {
		return nil, false
	}
	// Constructors are stored with normalized keys for case-insensitive lookup
	normalizedName := ident.Normalize(name)
	if ctorType, ok := ct.Constructors[normalizedName]; ok {
		return ctorType, true
	}
	if ct.Parent != nil {
		return ct.Parent.GetConstructor(name)
	}
	return nil, false
}

// HasProperty checks if the class or any of its ancestors has a property with the given name.
func (ct *ClassType) HasProperty(name string) bool {
	if ct == nil {
		return false
	}
	for propName := range ct.Properties {
		if ident.Equal(propName, name) {
			return true
		}
	}
	if ct.Parent != nil {
		return ct.Parent.HasProperty(name)
	}
	return false
}

// GetProperty returns the property info for a given property name, searching up the inheritance chain.
// Returns (propertyInfo, true) if found, or (nil, false) if not found.
func (ct *ClassType) GetProperty(name string) (*PropertyInfo, bool) {
	if ct == nil {
		return nil, false
	}
	for propName, prop := range ct.Properties {
		if ident.Equal(propName, name) {
			return prop, true
		}
	}
	if ct.Parent != nil {
		return ct.Parent.GetProperty(name)
	}
	return nil, false
}

// HasConstant checks if the class or any of its ancestors has a constant with the given name.
func (ct *ClassType) HasConstant(name string) bool {
	if ct == nil {
		return false
	}
	if _, ok := ct.Constants[name]; ok {
		return true
	}
	if ct.Parent != nil {
		return ct.Parent.HasConstant(name)
	}
	return false
}

// GetConstant retrieves a constant expression by name, searching up the inheritance chain.
// Returns (constantExpr, true) if found, or (nil, false) if not found.
func (ct *ClassType) GetConstant(name string) (interface{}, bool) {
	if ct == nil {
		return nil, false
	}
	if constant, ok := ct.Constants[name]; ok {
		return constant, true
	}
	if ct.Parent != nil {
		return ct.Parent.GetConstant(name)
	}
	return nil, false
}

// GetClassVar retrieves a class variable type by name, searching up the inheritance chain.
// Class variables are static members that belong to the class itself rather than instances.
// Supports class variable lookup with inheritance.
func (ct *ClassType) GetClassVar(name string) (Type, bool) {
	if ct == nil {
		return nil, false
	}
	// Case-insensitive class variable lookup
	normalizedName := ident.Normalize(name)
	if classVarType, ok := ct.ClassVars[normalizedName]; ok {
		return classVarType, true
	}
	if ct.Parent != nil {
		return ct.Parent.GetClassVar(name)
	}
	return nil, false
}

// ImplementsInterface checks if this class implements the given interface.
// It checks both the class itself and its parent classes.
func (ct *ClassType) ImplementsInterface(iface *InterfaceType) bool {
	if ct == nil || iface == nil {
		return false
	}

	// Check direct implementation
	for _, implemented := range ct.Interfaces {
		if implemented == iface || implemented.Equals(iface) {
			return true
		}
		// Check if implemented interface inherits from target interface
		if implemented.InheritsFrom(iface) {
			return true
		}
	}

	// Check parent class
	if ct.Parent != nil {
		return ct.Parent.ImplementsInterface(iface)
	}

	return false
}

// NewClassType creates a new class type with the given name and optional parent.
// Fields, ClassVars, Methods, and visibility maps are initialized as empty.
func NewClassType(name string, parent *ClassType) *ClassType {
	return &ClassType{
		Name:                 name,
		Parent:               parent,
		Fields:               make(map[string]Type),
		ClassVars:            make(map[string]Type),
		Constants:            make(map[string]interface{}),
		ConstantTypes:        make(map[string]Type),
		ConstantVisibility:   make(map[string]int),
		ClassVarVisibility:   make(map[string]int),
		Methods:              make(map[string]*FunctionType),
		MethodOverloads:      make(map[string][]*MethodInfo),
		FieldVisibility:      make(map[string]int),
		MethodVisibility:     make(map[string]int),
		VirtualMethods:       make(map[string]bool),
		OverrideMethods:      make(map[string]bool),
		AbstractMethods:      make(map[string]bool),
		ForwardedMethods:     make(map[string]bool),
		ReintroduceMethods:   make(map[string]bool),
		Operators:            NewOperatorRegistry(),
		Constructors:         make(map[string]*FunctionType),
		ConstructorOverloads: make(map[string][]*MethodInfo),
		ClassMethodFlags:     make(map[string]bool),
		Properties:           make(map[string]*PropertyInfo),
	}
}

// ClassOfType represents a metaclass type in DWScript.
// A metaclass type is a type that holds a reference to a class type itself,
// not an instance of the class.
//
// In DWScript/Delphi, metaclass types are declared as "class of ClassName".
// They allow:
// - Storing class references in variables
// - Passing class types as parameters
// - Calling constructors polymorphically through the metaclass
// - Virtual constructor dispatch
//
// Example:
//
//	type
//	  TAnimalClass = class of TAnimal;
//
//	var
//	  AnimalType: TAnimalClass;
//
//	AnimalType := TDog;      // Assign a class type
//	obj := AnimalType.Create; // Call constructor through metaclass
type ClassOfType struct {
	// ClassType is the class type that this metaclass references
	// For "class of TAnimal", this would be the TAnimal ClassType
	ClassType *ClassType
}

// String returns the string representation of the metaclass type
func (ct *ClassOfType) String() string {
	if ct.ClassType != nil {
		return fmt.Sprintf("class of %s", ct.ClassType.Name)
	}
	return "class of <unknown>"
}

// TypeKind returns "CLASSOF" for metaclass types
func (ct *ClassOfType) TypeKind() string {
	return "CLASSOF"
}

// Equals checks if two metaclass types are equal.
// Metaclass types are equal if their underlying class types are equal.
//
// Assignment compatibility rules:
// - "class of TDog" can be assigned a TDog reference
// - "class of TAnimal" can be assigned TDog if TDog inherits from TAnimal
// - The metaclass type determines what class types can be stored
func (ct *ClassOfType) Equals(other Type) bool {
	otherClassOf, ok := other.(*ClassOfType)
	if !ok {
		return false
	}
	if ct.ClassType == nil || otherClassOf.ClassType == nil {
		return ct.ClassType == otherClassOf.ClassType
	}
	return ct.ClassType.Equals(otherClassOf.ClassType)
}

// IsAssignableFrom checks if a class reference can be assigned to this metaclass variable.
// For example:
// - "class of TAnimal" can hold TAnimal or any derived class (TDog, TCat)
// - "class of TDog" can only hold TDog or classes derived from TDog
//
// This is used for type checking metaclass assignments like:
//
//	var cls: class of TAnimal;
//	cls := TDog;  // Valid if TDog derives from TAnimal
func (ct *ClassOfType) IsAssignableFrom(classType *ClassType) bool {
	if ct.ClassType == nil || classType == nil {
		return false
	}

	// Check if classType is the exact type or a descendant
	current := classType
	for current != nil {
		if current.Name == ct.ClassType.Name {
			return true
		}
		current = current.Parent
	}
	return false
}

// NewClassOfType creates a new metaclass type referencing the given class type.
// This is used for "class of T" type declarations in DWScript.
func NewClassOfType(classType *ClassType) *ClassOfType {
	return &ClassOfType{
		ClassType: classType,
	}
}

// InterfaceType represents an interface type in DWScript.
// Interfaces define a contract of methods that implementing classes must provide.
type InterfaceType struct {
	Parent       *InterfaceType
	Methods      map[string]*FunctionType
	Name         string
	ExternalName string
	IsExternal   bool
}

// String returns the string representation of the interface type
func (it *InterfaceType) String() string {
	return it.Name
}

// TypeKind returns "INTERFACE" for interface types
func (it *InterfaceType) TypeKind() string {
	return "INTERFACE"
}

// Equals checks if two interface types are equal.
// Interface types are equal if they have the same name.
func (it *InterfaceType) Equals(other Type) bool {
	otherInterface, ok := other.(*InterfaceType)
	if !ok {
		return false
	}
	return it.Name == otherInterface.Name
}

// HasMethod checks if the interface has a method with the given name
func (it *InterfaceType) HasMethod(name string) bool {
	_, exists := it.Methods[ident.Normalize(name)]
	return exists
}

// GetMethod retrieves the signature of a method by name
func (it *InterfaceType) GetMethod(name string) (*FunctionType, bool) {
	methodType, ok := it.Methods[ident.Normalize(name)]
	return methodType, ok
}

// InheritsFrom checks if this interface inherits from (extends) another interface.
// It checks the entire parent chain.
func (it *InterfaceType) InheritsFrom(parent *InterfaceType) bool {
	if it == nil || parent == nil {
		return false
	}

	// Check direct parent
	if it.Parent != nil {
		if it.Parent == parent || it.Parent.Equals(parent) {
			return true
		}
		// Recursively check parent's parents
		return it.Parent.InheritsFrom(parent)
	}

	return false
}

// NewInterfaceType creates a new interface type with the given name.
// Parent is set to nil (use explicit struct initialization for interface inheritance).
func NewInterfaceType(name string) *InterfaceType {
	return &InterfaceType{
		Name:         name,
		Parent:       nil,
		Methods:      make(map[string]*FunctionType),
		IsExternal:   false,
		ExternalName: "",
	}
}

// ============================================================================
// Interface Inheritance and Compatibility
// ============================================================================

// IsSubinterfaceOf checks if 'child' is a subinterface of 'parent'.
// This includes checking the entire inheritance chain.
func IsSubinterfaceOf(child, parent *InterfaceType) bool {
	if child == nil || parent == nil {
		return false
	}

	// Walk up the inheritance chain
	current := child
	for current != nil {
		if current.Name == parent.Name {
			return true
		}
		current = current.Parent
	}

	return false
}

// GetAllInterfaceMethods returns all methods of an interface, including inherited methods.
func GetAllInterfaceMethods(iface *InterfaceType) map[string]*FunctionType {
	if iface == nil {
		return make(map[string]*FunctionType)
	}

	// Start with parent methods (if any)
	allMethods := make(map[string]*FunctionType)
	if iface.Parent != nil {
		parentMethods := GetAllInterfaceMethods(iface.Parent)
		for name, method := range parentMethods {
			allMethods[name] = method
		}
	}

	// Add/override with own methods
	for name, method := range iface.Methods {
		allMethods[name] = method
	}

	return allMethods
}

// ============================================================================
// Type Compatibility and Checking (Stage 7.4 and 7.5)
// ============================================================================

// IsAssignableFrom checks if a value of type 'source' can be assigned to a variable of type 'target'.
// This includes:
// - Exact type match
// - Numeric coercion (Integer -> Float)
// - Subclass to superclass (covariance)
// - Class to interface (if class implements the interface)
// - Interface to interface (if source is subinterface of target)
func IsAssignableFrom(target, source Type) bool {
	// Exact match
	if target.Equals(source) {
		return true
	}

	// Numeric coercion: Integer -> Float
	if target.TypeKind() == "FLOAT" && source.TypeKind() == "INTEGER" {
		return true
	}

	// Subclass to superclass assignment
	if targetClass, ok := target.(*ClassType); ok {
		if sourceClass, ok := source.(*ClassType); ok {
			return IsSubclassOf(sourceClass, targetClass)
		}
	}

	// Class to interface assignment (check if class implements interface)
	if targetInterface, ok := target.(*InterfaceType); ok {
		if sourceClass, ok := source.(*ClassType); ok {
			return ImplementsInterface(sourceClass, targetInterface)
		}
	}

	// A derived interface can be assigned to a base interface
	if targetInterface, ok := target.(*InterfaceType); ok {
		if sourceInterface, ok := source.(*InterfaceType); ok {
			return IsSubinterfaceOf(sourceInterface, targetInterface)
		}
	}

	return false
}

// IsSubclassOf checks if 'child' is a subclass of 'parent'.
// This includes checking the entire inheritance chain.
func IsSubclassOf(child, parent *ClassType) bool {
	if child == nil || parent == nil {
		return false
	}

	// Walk up the inheritance chain
	current := child
	for current != nil {
		if current.Name == parent.Name {
			return true
		}
		current = current.Parent
	}

	return false
}

// IsClassRelated checks if two classes are in the same inheritance hierarchy.
// Returns true if one class is a subclass of the other (in either direction).
func IsClassRelated(class1, class2 *ClassType) bool {
	if class1 == nil || class2 == nil {
		return false
	}

	// Check if class1 is derived from class2
	if IsSubclassOf(class1, class2) {
		return true
	}

	// Check if class2 is derived from class1
	if IsSubclassOf(class2, class1) {
		return true
	}

	return false
}

// ImplementsInterface checks if a class implements all methods required by an interface.
// This uses structural typing - the class must have all methods with compatible signatures.
func ImplementsInterface(class *ClassType, iface *InterfaceType) bool {
	if class == nil || iface == nil {
		return false
	}

	// Check each method in the interface
	for methodName, ifaceMethodType := range iface.Methods {
		// Class must have the method
		classMethodType, found := class.GetMethod(methodName)
		if !found {
			return false
		}

		// Method signatures must match exactly
		if !classMethodType.Equals(ifaceMethodType) {
			return false
		}
	}

	return true
}

// IsClassType checks if a type is a class type
func IsClassType(t Type) bool {
	return t.TypeKind() == "CLASS"
}

// IsInterfaceType checks if a type is an interface type
func IsInterfaceType(t Type) bool {
	return t.TypeKind() == "INTERFACE"
}

// ============================================================================
// External Variable Support
// ============================================================================

// ExternalVarSymbol represents a variable that is implemented externally
// (outside of DWScript). This is used for future FFI (Foreign Function Interface)
// and JavaScript codegen compatibility.
//
// External variables are declared with the 'external' keyword:
//
//	var x: Integer; external;
//	var y: String; external 'customName';
//
// The interpreter will raise errors when attempting to access external variables
// until getter/setter functions are provided.
type ExternalVarSymbol struct {
	Type         Type
	ReadFunc     func() (any, error)
	WriteFunc    func(any) error
	Name         string
	ExternalName string
}
