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
	// Type is the resolved type (nil if not yet resolved).
	Type types.Type

	// DefaultValue is the expression to evaluate for optional parameters.
	// Nil for required parameters.
	DefaultValue ast.Expression

	// Name is the parameter name for binding arguments.
	Name string

	// ByRef indicates if this is a var parameter (pass-by-reference).
	ByRef bool
	// IsLazy delays evaluation until the parameter is accessed.
	IsLazy bool
	// IsConst marks a read-only parameter binding.
	IsConst bool
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

// NativeClassMethod implements a built-in class method body in Go.
//
// It receives the runtime class the call dispatched to (which lets one shared
// implementation serve several classes) and the already-evaluated arguments.
// A returned error is turned into a catchable DWScript exception carrying the
// error text as its Message, so implementations should use DWScript's wording.
type NativeClassMethod func(class IClassInfo, args []Value) (Value, error)

// MethodMetadata is the canonical runtime identity and resolved signature of a
// callable. It retains the executable AST payload and original declaration for
// semantic bindings; implementation binding updates this identity in place.
type MethodMetadata struct {
	// Declaration is the executable AST payload for this canonical runtime callable.
	Declaration *ast.FunctionDecl
	// Native implements the method body in Go instead of in AST. It is set only
	// for built-in classes (see the EncodingLib encoders) and, when present,
	// takes precedence over Declaration/Body.
	Native NativeClassMethod
	// SourceDeclaration identifies the original source node for semantic bindings.
	SourceDeclaration *ast.FunctionDecl
	// Owner is the runtime class that declared this callable.
	Owner          IClassInfo
	ReturnType     types.Type          // Resolved return type (nil for procedures)
	Body           ast.Statement       // AST statement block to execute
	PreConditions  *ast.PreConditions  // Assertions checked before execution
	PostConditions *ast.PostConditions // Assertions checked after execution
	Name           string              // Method/function name
	Parameters     []ParameterMetadata // Method parameters
	BytecodeID     int                 // ID of compiled bytecode if pre-compiled
	ID             MethodID            // Unique method identifier in registry
	Visibility     MethodVisibility    // Access control level
	IsStatic       bool                // Binds the defining class when a class method is invoked
	IsVirtual      bool                // Uses virtual dispatch
	IsAbstract     bool                // No implementation (abstract)
	IsOverride     bool                // Overrides parent's virtual method
	IsReintroduce  bool                // Breaks virtual dispatch chain
	IsClassMethod  bool                // Static method
	IsConstructor  bool                // Constructor method
	IsDestructor   bool                // Destructor method
}

// IsFunction returns true if this method has a return value.
func (m *MethodMetadata) IsFunction() bool {
	if m.ReturnType != nil {
		return m.ReturnType != types.VOID
	}
	// A forward declaration can precede resolution of its return type.
	return m.Declaration != nil && m.Declaration.ReturnType != nil
}

// IsProcedure returns true if this method has no return value.
func (m *MethodMetadata) IsProcedure() bool {
	return !m.IsFunction()
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
	Type       types.Type      // Resolved field type
	InitValue  ast.Expression  // Initializer expression
	Name       string          // Field name
	Visibility FieldVisibility // Access control level
}

// VirtualMethodMetadata tracks virtual method dispatch information.
// This replaces VirtualMethodEntry without AST dependencies.
type VirtualMethodMetadata = VirtualMethodEntry

// ClassMetadata contains runtime metadata for a class.
// This replaces the AST-dependent fields in ClassInfo.
//
// Design rationale:
//   - All methods stored as MethodMetadata (not *ast.FunctionDecl)
//   - All fields stored as FieldMetadata (not *ast.FieldDecl)
//   - Constants/ClassVars remain as Values (already runtime values)
//   - Enables independent evolution of runtime and AST representations
type ClassMetadata struct {
	Operators            *ClassOperatorRegistry            // Operator overload registry
	ConstructorOverloads map[string][]*MethodMetadata      // All constructor overload variants
	Destructor           *MethodMetadata                   // Class destructor
	Properties           map[string]*types.PropertyInfo    // Property metadata
	Fields               map[string]*FieldMetadata         // Instance fields
	Methods              map[string]*MethodMetadata        // Instance methods
	MethodOverloads      map[string][]*MethodMetadata      // Instance method overloads
	ClassMethods         map[string]*MethodMetadata        // Static methods
	ClassMethodOverloads map[string][]*MethodMetadata      // Static method overloads
	Constructors         map[string]*MethodMetadata        // Constructors
	ClassVars            map[string]Value                  // Class variable values
	Parent               *ClassMetadata                    // Parent class metadata
	VirtualMethods       map[string]*VirtualMethodMetadata // Virtual dispatch info
	Constants            map[string]Value                  // Evaluated constant values
	DefaultConstructor   string                            // Default constructor name
	Name                 string                            // Class name
	ExternalName         string                            // External implementation name
	Interfaces           []string                          // Implemented interface names
	IsAbstract           bool                              // Cannot be instantiated
	IsExternal           bool                              // Externally implemented
	IsPartial            bool                              // Partial class declaration
}

// NewClassMetadata creates a new ClassMetadata with initialized maps.
func NewClassMetadata(name string) *ClassMetadata {
	return &ClassMetadata{
		Name:                 name,
		Operators:            NewClassOperatorRegistry(),
		Fields:               make(map[string]*FieldMetadata),
		Methods:              make(map[string]*MethodMetadata),
		MethodOverloads:      make(map[string][]*MethodMetadata),
		ClassMethods:         make(map[string]*MethodMetadata),
		ClassMethodOverloads: make(map[string][]*MethodMetadata),
		Constructors:         make(map[string]*MethodMetadata),
		ConstructorOverloads: make(map[string][]*MethodMetadata),
		VirtualMethods:       make(map[string]*VirtualMethodMetadata),
		Constants:            make(map[string]Value),
		ClassVars:            make(map[string]Value),
		Properties:           make(map[string]*types.PropertyInfo),
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
	RecordType            interface{}                  // Underlying type information
	Fields                map[string]*FieldMetadata    // Record fields
	Methods               map[string]*MethodMetadata   // Instance methods
	MethodOverloads       map[string][]*MethodMetadata // Instance method overloads
	StaticMethods         map[string]*MethodMetadata   // Static methods
	StaticMethodOverloads map[string][]*MethodMetadata // Static method overloads
	Constants             map[string]interface{}       // Evaluated constant values
	ClassVars             map[string]interface{}       // Class variable values
	Name                  string                       // Record type name
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
	TargetType     types.Type                 // Type this helper extends
	ParentHelper   *HelperMetadata            // Parent helper (inheritance)
	Methods        map[string]*MethodMetadata // Helper methods
	Properties     map[string]interface{}     // Property metadata
	ClassVars      map[string]interface{}     // Class variable values
	ClassConsts    map[string]interface{}     // Class constant values
	BuiltinMethods map[string]string          // Built-in method mappings
	Name           string                     // Helper name
	IsRecordHelper bool                       // Record helper vs class helper
	IsClassHelper  bool                       // Class helper syntax was used
	IsStrict       bool                       // Strict helper lookup
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
