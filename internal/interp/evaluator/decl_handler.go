package evaluator

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

// DeclHandler handles type declaration processing (classes, interfaces, helpers).
// Encapsulates logic for building type metadata, virtual method tables, and inheritance hierarchies.
type DeclHandler interface {
	// ===== Class Declaration (21 methods) =====
	// All methods used by: visitor_declarations.go (single caller each)

	// NewClassInfoAdapter creates a new class metadata object.
	NewClassInfoAdapter(name string) any

	// CastToClassInfo type-asserts to ClassInfo type.
	CastToClassInfo(class any) (any, bool)

	// IsClassPartial checks if class is partial (forward declaration).
	IsClassPartial(classInfo any) bool

	// SetClassPartial marks a class as partial.
	SetClassPartial(classInfo any, isPartial bool)

	// SetClassAbstract marks a class as abstract.
	SetClassAbstract(classInfo any, isAbstract bool)

	// SetClassExternal marks a class as external with optional name.
	SetClassExternal(classInfo any, isExternal bool, externalName string)

	// ClassHasNoParent checks if class has no parent (is root class).
	ClassHasNoParent(classInfo any) bool

	// DefineCurrentClassMarker defines environment marker for current class context.
	DefineCurrentClassMarker(env any, classInfo any)

	// SetClassParent sets the parent class for inheritance.
	SetClassParent(classInfo any, parentClass any)

	// AddInterfaceToClass adds an interface to class's implemented interfaces.
	AddInterfaceToClass(classInfo any, interfaceInfo any, interfaceName string)

	// AddClassMethod adds a method to a class.
	// Returns true if method was added, false if it already exists.
	// Used by: visitor_declarations.go (2 uses - regular methods + operators)
	AddClassMethod(classInfo any, method *ast.FunctionDecl, className string) bool

	// SynthesizeDefaultConstructor creates a default constructor if none exists.
	SynthesizeDefaultConstructor(classInfo any)

	// AddClassProperty adds a property to a class.
	// Returns true if property was added, false if it already exists.
	AddClassProperty(classInfo any, propDecl *ast.PropertyDecl) bool

	// RegisterClassOperator registers an operator overload for a class.
	RegisterClassOperator(classInfo any, opDecl *ast.OperatorDecl) Value

	// LookupClassMethod finds a method by name in class hierarchy.
	// Returns (method, true) if found, (nil, false) if not found.
	// Used by: visitor_declarations.go (2 uses - method lookup + validation)
	LookupClassMethod(classInfo any, methodName string, isClassMethod bool) (any, bool)

	// SetClassConstructor sets the constructor for a class.
	SetClassConstructor(classInfo any, constructor any)

	// SetClassDestructor sets the destructor for a class.
	SetClassDestructor(classInfo any, destructor any)

	// InheritDestructorIfMissing inherits parent's destructor if class has none.
	InheritDestructorIfMissing(classInfo any)

	// InheritParentProperties inherits properties from parent class.
	InheritParentProperties(classInfo any)

	// BuildVirtualMethodTable constructs the virtual method table (VMT).
	BuildVirtualMethodTable(classInfo any)

	// RegisterClassInTypeSystem registers class in global type system.
	RegisterClassInTypeSystem(classInfo any, parentName string)

	// ===== Interface Declaration (7 methods) =====
	// All methods used by: visitor_declarations.go (single caller each)

	// NewInterfaceInfoAdapter creates a new interface metadata object.
	NewInterfaceInfoAdapter(name string) any

	// CastToInterfaceInfo type-asserts to InterfaceInfo type.
	CastToInterfaceInfo(iface any) (any, bool)

	// SetInterfaceParent sets the parent interface for inheritance.
	SetInterfaceParent(iface any, parent any)

	// GetInterfaceName returns the interface name.
	GetInterfaceName(iface any) string

	// GetInterfaceParent returns the parent interface.
	GetInterfaceParent(iface any) any

	// AddInterfaceMethod adds a method signature to an interface.
	AddInterfaceMethod(iface any, normalizedName string, method *ast.FunctionDecl)

	// AddInterfaceProperty adds a property signature to an interface.
	AddInterfaceProperty(iface any, normalizedName string, propInfo any)

	// ===== Helper Declaration (9 methods) =====
	// All methods used by: visitor_declarations.go (single caller each)
	// Exception: RegisterHelperLegacy has 2 uses in visitor_declarations.go

	// CreateHelperInfo creates a new helper (extension) metadata object.
	CreateHelperInfo(name string, targetType any, isRecordHelper bool) any

	// SetHelperParent sets the parent helper for inheritance.
	SetHelperParent(helper any, parent any)

	// VerifyHelperTargetTypeMatch verifies helper target type matches parent.
	// Returns true if types match or no parent, false on mismatch.
	VerifyHelperTargetTypeMatch(parent any, targetType any) bool

	// GetHelperName returns the helper name.
	GetHelperName(helper any) string

	// AddHelperMethod adds a method to a helper.
	AddHelperMethod(helper any, normalizedName string, method *ast.FunctionDecl)

	// AddHelperProperty adds a property to a helper.
	AddHelperProperty(helper any, prop *ast.PropertyDecl, propType any)

	// AddHelperClassVar adds a class variable to a helper.
	AddHelperClassVar(helper any, name string, value Value)

	// AddHelperClassConst adds a class constant to a helper.
	AddHelperClassConst(helper any, name string, value Value)

	// RegisterHelperLegacy registers helper in type system (legacy method).
	// Used by: visitor_declarations.go (2 uses - different registration paths)
	RegisterHelperLegacy(typeName string, helper any)

	// ===== Method Implementation (1 method) =====

	// EvalMethodImplementation evaluates method implementation AST node.
	// Used by: visitor_declarations.go (1 use)
	// Handles class/record method body evaluation and VMT updates.
	EvalMethodImplementation(fn *ast.FunctionDecl) Value
}
