// Package runtime provides runtime value types for the DWScript interpreter.
// This file contains the ClassMetadata interface that allows ObjectInstance
// to live in the runtime package without importing the interp package
// (which would create a circular dependency).
package runtime

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

// IClassInfo provides read-only access to class information.
// This interface allows ObjectInstance (in runtime package) to reference
// ClassInfo (in interp package) without creating a circular import dependency.
//
// ClassInfo (internal/interp/class.go) implements this interface.
type IClassInfo interface {
	// GetName returns the class name
	GetName() string

	// GetParent returns the parent class metadata, or nil if this is a root class
	GetParent() IClassInfo

	// GetMetadata returns the AST-free metadata (ClassMetadata), when available.
	// Returns nil if metadata is not yet available (during migration).
	GetMetadata() *ClassMetadata

	// IsAbstract returns true if this class is declared as abstract.
	// Abstract classes cannot be instantiated directly.
	// Task 3.5.22k: Added for CreateObject migration to evaluator.
	IsAbstract() bool

	// IsExternal returns true if this class is declared as external.
	// External classes cannot be instantiated (not supported in go-dws).
	// Task 3.5.22k: Added for CreateObject migration to evaluator.
	IsExternal() bool

	// LookupMethod finds a method by name in the class hierarchy.
	// Searches the current class first, then walks up the parent chain.
	// Returns the method declaration or nil if not found.
	// Name comparison is case-insensitive.
	LookupMethod(name string) *ast.FunctionDecl

	// LookupProperty finds a property by name in the class hierarchy.
	// Searches the current class first, then walks up the parent chain.
	// Returns property metadata or nil if not found.
	// Name comparison is case-insensitive.
	LookupProperty(name string) *PropertyInfo

	// GetDefaultProperty returns the default property for this class, if any.
	// Default properties allow indexing syntax: obj[index] instead of obj.Property[index].
	// Returns nil if no default property is defined.
	GetDefaultProperty() *PropertyInfo

	// FieldExists checks if a field with the given (normalized) name exists
	// in this class's field map. This is used for legacy field access.
	// The normalizedName parameter should already be normalized via ident.Normalize().
	FieldExists(normalizedName string) bool

	// GetFieldsMap returns the legacy field map for direct field access.
	// Used during migration period. Returns map[string]*ast.FieldDecl.
	// In the future, this will be replaced by metadata-only field access.
	GetFieldsMap() map[string]*ast.FieldDecl

	// GetMethodsMap returns the legacy method map for direct method access.
	// Used during migration period. Returns map[string]*ast.FunctionDecl.
	// In the future, this will be replaced by metadata-only method access.
	GetMethodsMap() map[string]*ast.FunctionDecl

	// LookupClassVar retrieves a class variable by name (case-insensitive).
	// Returns the value and the owning class (may be a parent class).
	// Returns (nil, nil) if the class variable is not found.
	LookupClassVar(name string) (Value, IClassInfo)

	// GetClassVarsMap returns the map of class variables for this class.
	// Used for iteration and initialization. Returns map[string]Value.
	GetClassVarsMap() map[string]Value

	// GetVirtualMethodTable returns the virtual method dispatch table.
	// Returns nil if no virtual methods are defined.
	// The table maps method signatures to virtual method entries.
	GetVirtualMethodTable() map[string]*VirtualMethodEntry

	// LookupOperator finds an operator overload for the given operator and operand types.
	// Returns the operator entry and true if found, or nil and false if not found.
	LookupOperator(operator string, operandTypes []string) (*OperatorEntry, bool)

	// GetInterfaces returns the list of interfaces this class implements.
	// Used for interface casting and type checking.
	GetInterfaces() []*InterfaceInfo

	// GetConstructor returns a constructor declaration by name (case-insensitive).
	// Returns nil if no constructor with that name exists.
	// Task 3.5.22k: Added for CreateObject migration to evaluator.
	GetConstructor(name string) *ast.FunctionDecl

	// GetFieldTypesMap returns the field name to type mapping for this class.
	// Used for field initialization during object creation.
	// Task 3.5.22k: Added for CreateObject migration to evaluator.
	GetFieldTypesMap() map[string]any
}

// PropertyInfo wraps property metadata for runtime access.
// This avoids importing internal/types into the runtime package.
type PropertyInfo struct {
	Name      string // Property name
	IsIndexed bool   // True if this is an indexed property
	IsDefault bool   // True if this is the default property
	ReadSpec  string // Field name or getter method name
	WriteSpec string // Field name or setter method name
	Impl      any    // *types.PropertyInfo - stored as any to avoid import cycle
}

// VirtualMethodEntry tracks virtual method dispatch information.
// Copied from internal/interp/class.go to avoid import cycle.
type VirtualMethodEntry struct {
	Method       *ast.FunctionDecl // The method declaration
	OwningClass  IClassInfo        // The class that first declared this virtual method
	IsVirtual    bool              // True if declared as 'virtual'
	IsOverride   bool              // True if declared as 'override'
	IsReintroduce bool             // True if declared as 'reintroduce'
}

// OperatorEntry represents an operator overload registration.
// Copied from internal/interp/operators.go to avoid import cycle.
type OperatorEntry struct {
	Operator      string            // Operator symbol (+, -, *, etc.)
	OperandTypes  []string          // Type names of operands
	Class         IClassInfo        // Class that owns the operator
	Method        *ast.FunctionDecl // Method implementing the operator (may be nil)
	BindingName   string            // Normalized method name binding
	SelfIndex     int               // Index of the 'self' operand (0 for unary/left, 1 for right, -1 for global)
	IsClassMethod bool              // True if this is a class method operator
}

// IInterfaceInfo provides read-only access to interface information.
// This interface allows InterfaceInstance (in runtime package) to reference
// InterfaceInfo (in interp package) without creating a circular import dependency.
//
// InterfaceInfo (internal/interp/interface.go) implements this interface.
type IInterfaceInfo interface {
	// GetName returns the interface name
	GetName() string

	// GetParent returns the parent interface, or nil if this is a root interface
	GetParent() IInterfaceInfo

	// GetMethod looks up a method by name in this interface (case-insensitive).
	// Searches the interface hierarchy, starting with this interface
	// and walking up through parent interfaces until the method is found.
	// Returns the method declaration (*ast.FunctionDecl passed as any) or nil if not found.
	GetMethod(name string) any

	// HasMethod checks if this interface (or any parent) has a method with the given name.
	// Name comparison is case-insensitive.
	HasMethod(name string) bool

	// GetProperty looks up a property by name in this interface (case-insensitive).
	// Searches the interface hierarchy until found.
	// Returns property metadata or nil if not found.
	GetProperty(name string) *PropertyInfo

	// HasProperty checks if the interface (or any parent) declares a property.
	// Name comparison is case-insensitive.
	HasProperty(name string) bool

	// GetDefaultProperty returns the default property defined on the interface hierarchy, if any.
	// Returns nil if no default property exists.
	GetDefaultProperty() *PropertyInfo

	// AllMethods returns all methods in this interface, including inherited methods.
	// Returns a new map containing all methods from this interface and its parents.
	AllMethods() map[string]any

	// AllProperties returns all properties declared on this interface and its parents.
	AllProperties() map[string]*PropertyInfo
}

// InterfaceInfo is a legacy type alias maintained for backward compatibility.
// New code should use IInterfaceInfo.
// TODO: Remove this alias after all code is migrated to use IInterfaceInfo.
type InterfaceInfo = IInterfaceInfo
