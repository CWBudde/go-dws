package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Exception Value Representation
// ============================================================================

// ExceptionValue is now in runtime package.
type ExceptionValue = runtime.ExceptionValue

// ============================================================================
// Built-in Exception Classes Registration
// ============================================================================

// registerBuiltinExceptions registers the Exception base class and standard exception types.
func (i *Interpreter) registerBuiltinExceptions() {
	// Register TObject as the root base class for all classes
	// This is required for DWScript compatibility - all classes ultimately inherit from TObject
	objectClass := NewClassInfo("TObject")
	objectClass.Parent = nil // Root of the class hierarchy
	objectClass.IsAbstractFlag = false
	objectClass.IsExternalFlag = false

	// Add basic TObject constructor
	// Create a minimal Create constructor AST node
	// The nil body means it just initializes fields with defaults
	createConstructor := &ast.FunctionDecl{
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Create"},
				},
			},
			Value: "Create",
		},
		Parameters:    []*ast.Parameter{},                                 // No parameters
		ReturnType:    nil,                                                // Constructors don't have explicit return types
		Body:          &ast.BlockStatement{Statements: []ast.Statement{}}, // Empty body
		IsConstructor: true,
	}
	// Use lowercase key for case-insensitive constructor matching
	objectClass.Constructors["create"] = createConstructor
	objectClass.ConstructorOverloads["create"] = []*ast.FunctionDecl{createConstructor}

	// Add default Destroy destructor (virtual) and Free method
	destroyMethod := &ast.FunctionDecl{
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Destroy"},
				},
			},
			Value: "Destroy",
		},
		Parameters:    []*ast.Parameter{},
		Body:          &ast.BlockStatement{Statements: []ast.Statement{}},
		IsDestructor:  true,
		IsVirtual:     true,
		Visibility:    ast.VisibilityPublic,
		IsConstructor: false,
	}
	freeMethod := &ast.FunctionDecl{
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Free"},
				},
			},
			Value: "Free",
		},
		Parameters: []*ast.Parameter{},
		Body:       &ast.BlockStatement{Statements: []ast.Statement{}},
		Visibility: ast.VisibilityPublic,
	}

	// Store methods with lowercase keys for case-insensitive lookup
	objectClass.Methods["destroy"] = destroyMethod
	objectClass.MethodOverloads["destroy"] = []*ast.FunctionDecl{destroyMethod}
	objectClass.Methods["free"] = freeMethod
	objectClass.MethodOverloads["free"] = []*ast.FunctionDecl{freeMethod}
	objectClass.Destructor = destroyMethod

	// Populate metadata for Destroy/Free (AST-free path)
	runtime.AddMethodToClass(objectClass.Metadata, &runtime.MethodMetadata{
		Name:           "Destroy",
		Parameters:     []runtime.ParameterMetadata{},
		ReturnType:     nil,
		ReturnTypeName: "",
		Body:           destroyMethod.Body,
		IsVirtual:      true,
		IsDestructor:   true,
		Visibility:     runtime.VisibilityPublic,
	}, false)
	runtime.AddMethodToClass(objectClass.Metadata, &runtime.MethodMetadata{
		Name:           "Free",
		Parameters:     []runtime.ParameterMetadata{},
		ReturnType:     nil,
		ReturnTypeName: "",
		Body:           freeMethod.Body,
		Visibility:     runtime.VisibilityPublic,
	}, false)
	objectClass.buildVirtualMethodTable()

	// Use lowercase key for O(1) case-insensitive lookup
	i.typeSystem.RegisterClass("TObject", objectClass)

	// Register Exception base class
	exceptionClass := NewClassInfo("Exception")
	exceptionClass.Parent = objectClass // Exception inherits from TObject
	exceptionClass.Fields["Message"] = types.STRING
	exceptionClass.IsAbstractFlag = false
	exceptionClass.IsExternalFlag = false

	// Set parent metadata for hierarchy checks
	exceptionClass.Metadata.Parent = objectClass.Metadata
	exceptionClass.Metadata.ParentName = "TObject"

	// Populate metadata for exception fields
	messageMeta := &runtime.FieldMetadata{
		Name:       "Message",
		TypeName:   "String",
		Type:       types.STRING,
		Visibility: runtime.FieldVisibilityPublic,
	}
	runtime.AddFieldToClass(exceptionClass.Metadata, messageMeta)

	// Add Create constructor - just a placeholder, will be handled specially
	exceptionClass.Constructors["Create"] = nil

	// Use lowercase key for O(1) case-insensitive lookup
	i.typeSystem.RegisterClassWithParent("Exception", exceptionClass, "TObject")

	// Register standard exception types
	standardExceptions := []string{
		"EConvertError",
		"ERangeError",
		"EDivByZero",
		"EAssertionFailed",
		"EInvalidOp",
		"EScriptStackOverflow",
		"EDelphi", // For Format() and other Delphi-compatible runtime errors
	}

	for _, excName := range standardExceptions {
		excClass := NewClassInfo(excName)
		excClass.Parent = exceptionClass
		excClass.Fields["Message"] = types.STRING
		excClass.IsAbstractFlag = false
		excClass.IsExternalFlag = false

		// Set parent metadata for hierarchy checks
		excClass.Metadata.Parent = exceptionClass.Metadata
		excClass.Metadata.ParentName = "Exception"

		// Populate metadata for exception fields
		messageMeta := &runtime.FieldMetadata{
			Name:       "Message",
			TypeName:   "String",
			Type:       types.STRING,
			Visibility: runtime.FieldVisibilityPublic,
		}
		runtime.AddFieldToClass(excClass.Metadata, messageMeta)

		// Inherit Create constructor
		excClass.Constructors["Create"] = nil

		// Use lowercase key for O(1) case-insensitive lookup
		i.typeSystem.RegisterClassWithParent(excName, excClass, "Exception")
	}

	// Register EHost exception wrapper for host runtime errors.
	eHostClass := NewClassInfo("EHost")
	eHostClass.Parent = exceptionClass
	eHostClass.Fields["Message"] = types.STRING
	eHostClass.Fields["ExceptionClass"] = types.STRING
	eHostClass.IsAbstractFlag = false
	eHostClass.IsExternalFlag = false

	// Set parent metadata for hierarchy checks
	eHostClass.Metadata.Parent = exceptionClass.Metadata
	eHostClass.Metadata.ParentName = "Exception"

	// Populate metadata for EHost fields
	messageMeta2 := &runtime.FieldMetadata{
		Name:       "Message",
		TypeName:   "String",
		Type:       types.STRING,
		Visibility: runtime.FieldVisibilityPublic,
	}
	runtime.AddFieldToClass(eHostClass.Metadata, messageMeta2)

	exceptionClassMeta := &runtime.FieldMetadata{
		Name:       "ExceptionClass",
		TypeName:   "String",
		Type:       types.STRING,
		Visibility: runtime.FieldVisibilityPublic,
	}
	runtime.AddFieldToClass(eHostClass.Metadata, exceptionClassMeta)

	eHostClass.Constructors["Create"] = nil

	// Use lowercase key for O(1) case-insensitive lookup
	i.typeSystem.RegisterClassWithParent("EHost", eHostClass, "Exception")
}

// raiseMaxRecursionExceeded raises an EScriptStackOverflow exception when the
// maximum recursion depth is exceeded. This prevents infinite recursion and
// stack overflow errors.
func (i *Interpreter) raiseMaxRecursionExceeded() Value {
	return i.raiseMaxRecursionExceededInContext(i.ctx)
}

func (i *Interpreter) raiseMaxRecursionExceededInContext(ctx *runtime.ExecutionContext) Value {
	if ctx == nil {
		ctx = i.ctx
	}

	message := fmt.Sprintf("Maximal recursion exceeded (%d)", i.engineState.MaxRecursionDepth)

	// Capture current call stack
	callStack := ctx.CallStack()

	stackOverflowClass := i.lookupRegisteredClassInfo("EScriptStackOverflow")
	if stackOverflowClass == nil {
		stackOverflowClass = i.lookupRegisteredClassInfo("Exception")
	}
	if stackOverflowClass == nil {
		// As a last resort, return NilValue without setting exception
		return &NilValue{}
	}

	// Create exception instance
	instance := NewObjectInstance(stackOverflowClass)
	instance.SetField("Message", &StringValue{Value: message})

	// Set the exception (Position is nil for internally-raised exceptions like recursion overflow)
	ctx.SetException(&runtime.ExceptionValue{
		Metadata:  stackOverflowClass.Metadata,
		Instance:  instance,
		Message:   message,
		Position:  nil,
		CallStack: callStack,
		ClassInfo: stackOverflowClass, // Deprecated: backward compatibility
	})

	return &NilValue{}
}

// matchesExceptionType checks if an exception matches a handler's type.
func (i *Interpreter) matchesExceptionType(exc *ExceptionValue, typeExpr ast.TypeExpression) bool {
	if typeExpr == nil {
		return true // Bare handler catches all
	}

	typeName := typeExpr.String()

	// Prefer metadata if available
	if exc.Metadata != nil {
		// Check if exception class matches or inherits from handler type
		currentMetadata := exc.Metadata
		for currentMetadata != nil {
			if currentMetadata.Name == typeName {
				return true
			}
			// Check parent class metadata
			currentMetadata = currentMetadata.Parent
		}
		return false
	}

	// Fallback to ClassInfo for backward compatibility
	if exc.ClassInfo != nil {
		// Check if exception class matches or inherits from handler type
		if classInfo, ok := exc.ClassInfo.(*ClassInfo); ok {
			currentClass := classInfo
			for currentClass != nil {
				if currentClass.Name == typeName {
					return true
				}
				// Check parent class
				currentClass = currentClass.Parent
			}
		}
	}

	return false
}

// isExceptionClass checks if a class is an Exception or inherits from Exception.
func (i *Interpreter) isExceptionClass(classInfo *ClassInfo) bool {
	current := classInfo
	for current != nil {
		if current.Name == "Exception" {
			return true
		}
		current = current.Parent
	}
	return false
}
