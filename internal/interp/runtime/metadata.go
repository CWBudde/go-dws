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

	// TypeName is the string representation of the type (for display/debugging).
	TypeName string

	// ByRef indicates if this is a var parameter (pass-by-reference).
	ByRef bool
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
	ReturnType     types.Type              // Resolved return type (nil for procedures)
	Body           ast.Statement           // AST statement block to execute (Phase 9)
	PreConditions  *ast.PreConditions      // Assertions checked before execution
	PostConditions *ast.PostConditions     // Assertions checked after execution
	NativeFunc     func(args []interface{}) interface{} // Built-in function implementation
	Name           string                  // Method/function name
	ReturnTypeName string                  // String representation of return type
	Parameters     []ParameterMetadata     // Method parameters
	BytecodeID     int                     // ID of compiled bytecode (Phase 10+)
	ID             MethodID                // Unique method identifier in registry
	Visibility     MethodVisibility        // Access control level
	IsVirtual      bool                    // Uses virtual dispatch
	IsAbstract     bool                    // No implementation (abstract)
	IsOverride     bool                    // Overrides parent's virtual method
	IsReintroduce  bool                    // Breaks virtual dispatch chain
	IsClassMethod  bool                    // Static method
	IsConstructor  bool                    // Constructor method
	IsDestructor   bool                    // Destructor method
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
	Type       types.Type        // Resolved field type
	InitValue  ast.Expression    // Initializer expression
	Name       string            // Field name
	TypeName   string            // String representation of type
	Visibility FieldVisibility   // Access control level
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
	Operators            interface{}                        // Operator overload registry
	ConstructorOverloads map[string][]*MethodMetadata       // All constructor overload variants
	Destructor           *MethodMetadata                    // Class destructor
	Properties           map[string]interface{}             // Property metadata
	Fields               map[string]*FieldMetadata          // Instance fields
	Methods              map[string]*MethodMetadata         // Instance methods
	MethodOverloads      map[string][]*MethodMetadata       // Instance method overloads
	ClassMethods         map[string]*MethodMetadata         // Static methods
	ClassMethodOverloads map[string][]*MethodMetadata       // Static method overloads
	Constructors         map[string]*MethodMetadata         // Constructors
	ClassVars            map[string]interface{}             // Class variable values
	Parent               *ClassMetadata                     // Parent class metadata
	VirtualMethods       map[string]*VirtualMethodMetadata  // Virtual dispatch info
	Constants            map[string]interface{}             // Evaluated constant values
	DefaultConstructor   string                             // Default constructor name
	Name                 string                             // Class name
	ParentName           string                             // Parent class name
	ExternalName         string                             // External implementation name
	Interfaces           []string                           // Implemented interface names
	IsAbstract           bool                               // Cannot be instantiated
	IsExternal           bool                               // Externally implemented
	IsPartial            bool                               // Partial class declaration
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
