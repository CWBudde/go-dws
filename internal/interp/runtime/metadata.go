// Package runtime provides runtime metadata structures for the DWScript interpreter.
// This file contains AST-free metadata types that replace AST node dependencies
// in runtime type information.
package runtime

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ParameterMetadata describes a function/method parameter at runtime.
// This replaces the need to access *ast.Parameter at runtime.
type ParameterMetadata struct {
	// Name is the parameter name for binding arguments.
	Name string

	// TypeName is the string representation of the type (for display/debugging).
	TypeName string

	// Type is the resolved type (nil if not yet resolved).
	Type types.Type

	// ByRef indicates if this is a var parameter (pass-by-reference).
	ByRef bool

	// DefaultValue is the expression to evaluate for optional parameters.
	// Nil for required parameters.
	// Phase 9: Keeps AST expression; Phase 10+: Migrate to bytecode.
	DefaultValue ast.Expression
}

// MethodVisibility represents method visibility levels in DWScript.
type MethodVisibility int

const (
	// VisibilityPublic means the method is accessible from anywhere.
	VisibilityPublic MethodVisibility = iota

	// VisibilityPrivate means the method is only accessible within the class.
	VisibilityPrivate

	// VisibilityProtected means the method is accessible within the class and descendants.
	VisibilityProtected

	// VisibilityPublished means the method is public and also published for RTTI.
	VisibilityPublished
)

// String returns the string representation of MethodVisibility.
func (v MethodVisibility) String() string {
	switch v {
	case VisibilityPublic:
		return "public"
	case VisibilityPrivate:
		return "private"
	case VisibilityProtected:
		return "protected"
	case VisibilityPublished:
		return "published"
	default:
		return "unknown"
	}
}

// MethodMetadata describes a callable method/function at runtime.
// This replaces the need to store full *ast.FunctionDecl nodes in runtime types.
//
// Design rationale:
//   - Stores only information needed at runtime (signature, flags, visibility)
//   - Executable body can be AST (Phase 9), bytecode ID (Phase 10+), or native func
//   - Reduces memory overhead compared to full AST nodes
//   - Enables serialization for bytecode cache
type MethodMetadata struct {
	// === Identity ===

	// ID is the unique method identifier in the MethodRegistry.
	// Set by MethodRegistry.RegisterMethod().
	ID MethodID

	// === Signature Information ===

	// Name is the method/function name.
	Name string

	// Parameters describes the method parameters.
	Parameters []ParameterMetadata

	// ReturnTypeName is the string representation of the return type.
	// Empty string for procedures (no return value).
	ReturnTypeName string

	// ReturnType is the resolved return type.
	// Nil for procedures.
	ReturnType types.Type

	// === Executable Body ===
	// Exactly one of Body, BytecodeID, or NativeFunc should be set.

	// Body is the AST statement block to execute.
	// Phase 9: Used for user-defined functions/methods.
	// Phase 10+: Will be replaced with BytecodeID.
	Body ast.Statement

	// BytecodeID is the ID of compiled bytecode in the bytecode registry.
	// Future: Phase 10+ will use this instead of Body.
	BytecodeID int

	// NativeFunc is a built-in function implementation.
	// Used for standard library functions (Print, Length, etc.).
	NativeFunc func(args []interface{}) interface{}

	// === Validation ===
	// Phase 9: Keeps AST structures; Phase 10+: Migrate to bytecode.

	// PreConditions are assertions checked before method execution.
	PreConditions *ast.PreConditions

	// PostConditions are assertions checked after method execution.
	PostConditions *ast.PostConditions

	// === Method Characteristics ===

	// IsVirtual indicates this method uses virtual dispatch.
	IsVirtual bool

	// IsAbstract indicates this method has no implementation (abstract).
	IsAbstract bool

	// IsOverride indicates this method overrides a parent's virtual method.
	IsOverride bool

	// IsReintroduce indicates this method breaks the virtual dispatch chain.
	IsReintroduce bool

	// IsClassMethod indicates this is a static method (class function/procedure).
	IsClassMethod bool

	// IsConstructor indicates this is a constructor.
	IsConstructor bool

	// IsDestructor indicates this is a destructor.
	IsDestructor bool

	// === Visibility ===

	// Visibility controls access to this method.
	Visibility MethodVisibility
}

// IsFunction returns true if this method has a return value.
func (m *MethodMetadata) IsFunction() bool {
	return m.ReturnTypeName != ""
}

// IsProcedure returns true if this method has no return value.
func (m *MethodMetadata) IsProcedure() bool {
	return m.ReturnTypeName == ""
}

// RequiredParamCount returns the number of required (non-optional) parameters.
func (m *MethodMetadata) RequiredParamCount() int {
	count := 0
	for _, param := range m.Parameters {
		if param.DefaultValue == nil {
			count++
		}
	}
	return count
}

// ParamCount returns the total number of parameters.
func (m *MethodMetadata) ParamCount() int {
	return len(m.Parameters)
}

// FieldVisibility represents field visibility levels in DWScript.
type FieldVisibility int

const (
	// FieldVisibilityPublic means the field is accessible from anywhere.
	FieldVisibilityPublic FieldVisibility = iota

	// FieldVisibilityPrivate means the field is only accessible within the class.
	FieldVisibilityPrivate

	// FieldVisibilityProtected means the field is accessible within the class and descendants.
	FieldVisibilityProtected

	// FieldVisibilityPublished means the field is public and also published for RTTI.
	FieldVisibilityPublished
)

// String returns the string representation of FieldVisibility.
func (v FieldVisibility) String() string {
	switch v {
	case FieldVisibilityPublic:
		return "public"
	case FieldVisibilityPrivate:
		return "private"
	case FieldVisibilityProtected:
		return "protected"
	case FieldVisibilityPublished:
		return "published"
	default:
		return "unknown"
	}
}

// FieldMetadata describes a field at runtime.
// This replaces the need to store *ast.FieldDecl in runtime types.
type FieldMetadata struct {
	// Name is the field name.
	Name string

	// TypeName is the string representation of the field type.
	TypeName string

	// Type is the resolved field type.
	Type types.Type

	// InitValue is the initializer expression evaluated when creating instances.
	// Nil if the field has no initializer.
	// Phase 9: Keeps AST expression; Phase 10+: Migrate to bytecode.
	InitValue ast.Expression

	// Visibility controls access to this field.
	Visibility FieldVisibility
}

// VirtualMethodMetadata tracks virtual method dispatch information.
// This replaces VirtualMethodEntry without AST dependencies.
type VirtualMethodMetadata struct {
	// IntroducedBy is the class that first declared this method as virtual.
	IntroducedBy *ClassMetadata

	// Implementation is the method to actually call for this class.
	Implementation *MethodMetadata

	// IsVirtual indicates this method participates in virtual dispatch.
	IsVirtual bool

	// IsReintroduced indicates this method breaks the virtual dispatch chain.
	IsReintroduced bool
}

// ClassMetadata contains runtime metadata for a class.
// This replaces the AST-dependent fields in ClassInfo.
//
// Design rationale:
//   - All methods stored as MethodMetadata (not *ast.FunctionDecl)
//   - All fields stored as FieldMetadata (not *ast.FieldDecl)
//   - Constants/ClassVars remain as Values (already runtime values)
//   - Enables independent evolution of runtime and AST representations
type ClassMetadata struct {
	// === Basic Information ===

	// Name is the class name.
	Name string

	// ParentName is the parent class name (empty for TObject root).
	ParentName string

	// Parent is a pointer to the parent class metadata.
	Parent *ClassMetadata

	// Interfaces is the list of interface names this class implements.
	Interfaces []string

	// === Fields ===

	// Fields maps field names (normalized) to field metadata.
	Fields map[string]*FieldMetadata

	// === Instance Methods ===

	// Methods maps method names (normalized) to primary method.
	// For overloaded methods, this contains the first declaration.
	Methods map[string]*MethodMetadata

	// MethodOverloads maps method names (normalized) to all overload variants.
	MethodOverloads map[string][]*MethodMetadata

	// === Static Methods (Class Methods) ===

	// ClassMethods maps class method names (normalized) to primary method.
	ClassMethods map[string]*MethodMetadata

	// ClassMethodOverloads maps class method names (normalized) to all overload variants.
	ClassMethodOverloads map[string][]*MethodMetadata

	// === Constructors ===

	// Constructors maps constructor names (normalized) to primary constructor.
	Constructors map[string]*MethodMetadata

	// ConstructorOverloads maps constructor names (normalized) to all overload variants.
	ConstructorOverloads map[string][]*MethodMetadata

	// DefaultConstructor is the name of the default constructor (usually "Create").
	DefaultConstructor string

	// === Destructor ===

	// Destructor is the class destructor (usually "Destroy").
	Destructor *MethodMetadata

	// === Virtual Dispatch ===

	// VirtualMethods maps method names (normalized) to virtual dispatch info.
	VirtualMethods map[string]*VirtualMethodMetadata

	// === Constants and Class Variables ===
	// These are already runtime values, so no change from ClassInfo.

	// Constants maps constant names (normalized) to evaluated values.
	Constants map[string]interface{}

	// ClassVars maps class variable names (normalized) to current values.
	ClassVars map[string]interface{}

	// === Properties ===
	// PropertyInfo is already metadata, so no change from ClassInfo.

	// Properties maps property names (normalized) to property metadata.
	Properties map[string]interface{} // Will be *types.PropertyInfo

	// === Operators ===
	// Operator registry is already runtime metadata, no change.

	// Operators is the operator overload registry for this class.
	Operators interface{} // Will be *runtimeOperatorRegistry

	// === Class Flags ===

	// IsAbstract indicates this class cannot be instantiated.
	IsAbstract bool

	// IsExternal indicates this class is implemented externally.
	IsExternal bool

	// IsPartial indicates this is a partial class declaration.
	IsPartial bool

	// ExternalName is the external implementation name (if IsExternal).
	ExternalName string
}

// NewClassMetadata creates a new ClassMetadata with initialized maps.
func NewClassMetadata(name string) *ClassMetadata {
	return &ClassMetadata{
		Name:                 name,
		Fields:               make(map[string]*FieldMetadata),
		Methods:              make(map[string]*MethodMetadata),
		MethodOverloads:      make(map[string][]*MethodMetadata),
		ClassMethods:         make(map[string]*MethodMetadata),
		ClassMethodOverloads: make(map[string][]*MethodMetadata),
		Constructors:         make(map[string]*MethodMetadata),
		ConstructorOverloads: make(map[string][]*MethodMetadata),
		VirtualMethods:       make(map[string]*VirtualMethodMetadata),
		Constants:            make(map[string]interface{}),
		ClassVars:            make(map[string]interface{}),
		Properties:           make(map[string]interface{}),
	}
}

// RecordMetadata contains runtime metadata for a record type.
// This replaces the AST-dependent fields in RecordTypeValue.
//
// Design rationale:
//   - Similar to ClassMetadata but simpler (no inheritance, constructors, virtual dispatch)
//   - All methods stored as MethodMetadata
//   - All fields stored as FieldMetadata
type RecordMetadata struct {
	// === Basic Information ===

	// Name is the record type name.
	Name string

	// RecordType is the underlying type information.
	RecordType interface{} // Will be *types.RecordType

	// === Fields ===

	// Fields maps field names (normalized) to field metadata.
	Fields map[string]*FieldMetadata

	// === Instance Methods ===

	// Methods maps method names (normalized) to primary method.
	Methods map[string]*MethodMetadata

	// MethodOverloads maps method names (normalized) to all overload variants.
	MethodOverloads map[string][]*MethodMetadata

	// === Static Methods ===

	// StaticMethods maps static method names (normalized) to primary method.
	StaticMethods map[string]*MethodMetadata

	// StaticMethodOverloads maps static method names (normalized) to all overload variants.
	StaticMethodOverloads map[string][]*MethodMetadata

	// === Constants and Class Variables ===

	// Constants maps constant names (normalized) to evaluated values.
	Constants map[string]interface{}

	// ClassVars maps class variable names (normalized) to current values.
	ClassVars map[string]interface{}
}

// NewRecordMetadata creates a new RecordMetadata with initialized maps.
func NewRecordMetadata(name string, recordType interface{}) *RecordMetadata {
	return &RecordMetadata{
		Name:                  name,
		RecordType:            recordType,
		Fields:                make(map[string]*FieldMetadata),
		Methods:               make(map[string]*MethodMetadata),
		MethodOverloads:       make(map[string][]*MethodMetadata),
		StaticMethods:         make(map[string]*MethodMetadata),
		StaticMethodOverloads: make(map[string][]*MethodMetadata),
		Constants:             make(map[string]interface{}),
		ClassVars:             make(map[string]interface{}),
	}
}

// HelperMetadata contains runtime metadata for a helper type.
// This replaces HelperInfo's AST-dependent fields.
type HelperMetadata struct {
	// Name is the helper name.
	Name string

	// TargetType is the type this helper extends.
	TargetType types.Type

	// ParentHelper is the parent helper (for helper inheritance).
	ParentHelper *HelperMetadata

	// Methods maps method names (normalized) to method metadata.
	Methods map[string]*MethodMetadata

	// Properties maps property names (normalized) to property metadata.
	Properties map[string]interface{} // Will be *types.PropertyInfo

	// ClassVars maps class variable names to values.
	ClassVars map[string]interface{}

	// ClassConsts maps class constant names to values.
	ClassConsts map[string]interface{}

	// BuiltinMethods maps method names to built-in method names.
	// Used for array helpers and other built-in types.
	BuiltinMethods map[string]string

	// IsRecordHelper indicates if this is a record helper (vs class helper).
	IsRecordHelper bool
}

// NewHelperMetadata creates a new HelperMetadata with initialized maps.
func NewHelperMetadata(name string, targetType types.Type, isRecordHelper bool) *HelperMetadata {
	return &HelperMetadata{
		Name:           name,
		TargetType:     targetType,
		Methods:        make(map[string]*MethodMetadata),
		Properties:     make(map[string]interface{}),
		ClassVars:      make(map[string]interface{}),
		ClassConsts:    make(map[string]interface{}),
		BuiltinMethods: make(map[string]string),
		IsRecordHelper: isRecordHelper,
	}
}
