package types

import "fmt"

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

// ============================================================================
// Singleton Type Constants
// ============================================================================

// Singleton instances of basic types
// These are used throughout the compiler for type checking
var (
	INTEGER = &IntegerType{}
	FLOAT   = &FloatType{}
	STRING  = &StringType{}
	BOOLEAN = &BooleanType{}
	NIL     = &NilType{}
	VOID    = &VoidType{}
)

// Task 7.75: IINTERFACE is the base interface type (like IUnknown in COM)
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
// Subrange Types (Task 9.91)
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
// Task 9.92
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

// IsOrdinalType checks if a type is an ordinal type (used for loop variables)
// In DWScript, ordinal types are Integer, Boolean, and enumerations
func IsOrdinalType(t Type) bool {
	kind := t.TypeKind()
	return kind == "INTEGER" || kind == "BOOLEAN" || kind == "ENUM"
}

// TypeFromString converts a type name string to a Type
// This is useful for parsing type annotations
func TypeFromString(name string) (Type, error) {
	switch name {
	case "Integer":
		return INTEGER, nil
	case "Float":
		return FLOAT, nil
	case "String":
		return STRING, nil
	case "Boolean":
		return BOOLEAN, nil
	case "Void":
		return VOID, nil
	default:
		return nil, fmt.Errorf("unknown type: %s", name)
	}
}

// ============================================================================
// Object-Oriented Types (Stage 7)
// ============================================================================

// PropAccessKind represents how a property is accessed (read or write).
// Task 8.26b: ReadKind/WriteKind enum
type PropAccessKind int

const (
	PropAccessNone       PropAccessKind = iota // No access (write-only has ReadKind=None, read-only has WriteKind=None)
	PropAccessField                            // Direct field access (e.g., FName)
	PropAccessMethod                           // Method call (e.g., GetName, SetName)
	PropAccessExpression                       // Expression-based getter (e.g., (FValue * 2))
)

// PropertyInfo represents property metadata for a class.
// Task 8.26a: Fields: Name, Type, ReadSpec, WriteSpec, IsIndexed, IsDefault
// Properties provide syntactic sugar for getter/setter access.
type PropertyInfo struct {
	Type      Type
	Name      string
	ReadSpec  string
	WriteSpec string
	ReadKind  PropAccessKind
	WriteKind PropAccessKind
	IsIndexed bool
	IsDefault bool
}

// ClassType represents a class type in DWScript.
// Classes support inheritance, fields, methods, and class variables (static fields).
type ClassType struct {
	OverrideMethods  map[string]bool
	AbstractMethods  map[string]bool
	Fields           map[string]Type
	ClassVars        map[string]Type
	Methods          map[string]*FunctionType
	FieldVisibility  map[string]int
	MethodVisibility map[string]int
	VirtualMethods   map[string]bool
	Parent           *ClassType
	Properties       map[string]*PropertyInfo
	ClassMethodFlags map[string]bool
	Constructors     map[string]*FunctionType
	Operators        *OperatorRegistry
	ExternalName     string
	Name             string
	Interfaces       []*InterfaceType
	IsAbstract       bool
	IsExternal       bool
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
	if _, ok := ct.Fields[name]; ok {
		return true
	}
	if ct.Parent != nil {
		return ct.Parent.HasField(name)
	}
	return false
}

// GetField retrieves the type of a field by name, searching up the inheritance chain
func (ct *ClassType) GetField(name string) (Type, bool) {
	if fieldType, ok := ct.Fields[name]; ok {
		return fieldType, true
	}
	if ct.Parent != nil {
		return ct.Parent.GetField(name)
	}
	return nil, false
}

// HasMethod checks if the class or any of its ancestors has a method with the given name
func (ct *ClassType) HasMethod(name string) bool {
	if _, ok := ct.Methods[name]; ok {
		return true
	}
	if ct.Parent != nil {
		return ct.Parent.HasMethod(name)
	}
	return false
}

// GetMethod retrieves the signature of a method by name, searching up the inheritance chain
func (ct *ClassType) GetMethod(name string) (*FunctionType, bool) {
	if methodType, ok := ct.Methods[name]; ok {
		return methodType, true
	}
	if ct.Parent != nil {
		return ct.Parent.GetMethod(name)
	}
	return nil, false
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
func (ct *ClassType) HasConstructor(name string) bool {
	if ct == nil {
		return false
	}
	if _, ok := ct.Constructors[name]; ok {
		return true
	}
	if ct.Parent != nil {
		return ct.Parent.HasConstructor(name)
	}
	return false
}

// HasProperty checks if the class or any of its ancestors has a property with the given name.
// Task 8.28
func (ct *ClassType) HasProperty(name string) bool {
	if ct == nil {
		return false
	}
	if _, ok := ct.Properties[name]; ok {
		return true
	}
	if ct.Parent != nil {
		return ct.Parent.HasProperty(name)
	}
	return false
}

// GetProperty returns the property info for a given property name, searching up the inheritance chain.
// Returns (propertyInfo, true) if found, or (nil, false) if not found.
// Task 8.28
func (ct *ClassType) GetProperty(name string) (*PropertyInfo, bool) {
	if ct == nil {
		return nil, false
	}
	if prop, ok := ct.Properties[name]; ok {
		return prop, true
	}
	if ct.Parent != nil {
		return ct.Parent.GetProperty(name)
	}
	return nil, false
}

// NewClassType creates a new class type with the given name and optional parent.
// Fields, ClassVars, Methods, and visibility maps are initialized as empty.
func NewClassType(name string, parent *ClassType) *ClassType {
	return &ClassType{
		Name:             name,
		Parent:           parent,
		Fields:           make(map[string]Type),
		ClassVars:        make(map[string]Type),
		Methods:          make(map[string]*FunctionType),
		FieldVisibility:  make(map[string]int),  // Task 7.63f
		MethodVisibility: make(map[string]int),  // Task 7.63f
		VirtualMethods:   make(map[string]bool), // Task 7.64
		OverrideMethods:  make(map[string]bool), // Task 7.64
		AbstractMethods:  make(map[string]bool), // Task 7.65
		Operators:        NewOperatorRegistry(),
		Constructors:     make(map[string]*FunctionType),
		ClassMethodFlags: make(map[string]bool),
		Properties:       make(map[string]*PropertyInfo), // Task 8.27
	}
}

// InterfaceType represents an interface type in DWScript.
// Interfaces define a contract of methods that implementing classes must provide.
// Task 7.73-7.74: Extended with Parent for inheritance, IsExternal/ExternalName for FFI
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
	_, ok := it.Methods[name]
	return ok
}

// GetMethod retrieves the signature of a method by name
func (it *InterfaceType) GetMethod(name string) (*FunctionType, bool) {
	methodType, ok := it.Methods[name]
	return methodType, ok
}

// NewInterfaceType creates a new interface type with the given name.
// Parent is set to nil (use explicit struct initialization for interface inheritance).
// Task 7.74: Initializes all fields including IsExternal (false) and ExternalName (empty).
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
// Task 7.77: Interface inheritance checking with circular detection.
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
// Task 7.78: Interface method inheritance - collects methods from entire hierarchy.
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
// - Interface to interface (if source is subinterface of target) - Task 7.79
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

	// Task 7.79: Interface to interface assignment (covariant)
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
