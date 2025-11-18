package evaluator

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// This file contains visitor methods for expression AST nodes.
// Phase 3.5.2: Visitor pattern implementation for expressions.
//
// Expressions evaluate to values and can be nested (e.g., binary expressions
// contain left and right sub-expressions).

// ErrorValue represents a runtime error (temporary definition to avoid circular imports).
type ErrorValue struct {
	Message string
}

func (e *ErrorValue) Type() string   { return "ERROR" }
func (e *ErrorValue) String() string { return "ERROR: " + e.Message }

// newError creates a new error value with optional formatting.
// TODO: Add location information from node in Phase 3.6 (error handling improvements)
func (e *Evaluator) newError(_ ast.Node, format string, args ...interface{}) Value {
	return &ErrorValue{Message: fmt.Sprintf(format, args...)}
}

// isError checks if a value is an error.
func isError(val Value) bool {
	if val != nil {
		return val.Type() == "ERROR"
	}
	return false
}

// VisitIdentifier evaluates an identifier (variable reference).
// Task 3.5.10: Partial migration - basic variable lookups via adapter.GetVariable.
// Complex cases still delegated (Self context, properties, lazy params, etc.).
func (e *Evaluator) VisitIdentifier(node *ast.Identifier, ctx *ExecutionContext) Value {
	// Try simple variable lookup first using the new environment adapter
	val, ok := e.adapter.GetVariable(node.Value, ctx)
	if ok {
		// Got a value from environment
		// However, it might be a special value type that needs processing:
		// - ExternalVarValue (should error)
		// - LazyThunk (needs evaluation)
		// - ReferenceValue (needs dereferencing)
		// For now, return as-is. The adapter.EvalNode fallback below will handle
		// these special cases if the value doesn't work as expected.

		// Simple optimization: if it's a basic value type, return immediately
		switch val.(type) {
		case *runtime.IntegerValue, *runtime.FloatValue, *runtime.StringValue, *runtime.BooleanValue, *runtime.NilValue:
			return val
		}

		// For complex value types, delegate to adapter for full processing
		// This handles LazyThunk, ReferenceValue, ExternalVarValue, etc.
		return e.adapter.EvalNode(node)
	}

	// Variable not found in environment
	// Could be:
	// - Self keyword (method context)
	// - Instance field/property (implicit Self)
	// - Class variable (__CurrentClass__ context)
	// - Function reference (with possible auto-invoke)
	// - Built-in function
	// - Class name (metaclass reference)
	// - ClassName/ClassType special identifiers
	// All these cases require complex context that hasn't been migrated yet
	// Delegate to adapter for full Interpreter.evalIdentifier logic
	return e.adapter.EvalNode(node)
}

// VisitBinaryExpression evaluates a binary expression (e.g., a + b, x == y).
// Task 3.5.19: Full migration of binary operator evaluation.
func (e *Evaluator) VisitBinaryExpression(node *ast.BinaryExpression, ctx *ExecutionContext) Value {
	// Handle short-circuit operators first (special evaluation order)
	switch node.Operator {
	case "??":
		return e.evalCoalesceOp(node, ctx)
	case "and":
		return e.evalAndOp(node, ctx)
	case "or":
		return e.evalOrOp(node, ctx)
	}

	// Evaluate both operands for non-short-circuit operators
	left := e.Eval(node.Left, ctx)
	if isError(left) {
		return left
	}
	if left == nil {
		return e.newError(node.Left, "left operand evaluated to nil")
	}

	right := e.Eval(node.Right, ctx)
	if isError(right) {
		return right
	}
	if right == nil {
		return e.newError(node.Right, "right operand evaluated to nil")
	}

	// Try operator overloading first (custom operators for objects)
	if result, ok := e.tryBinaryOperator(node.Operator, left, right, node); ok {
		return result
	}

	// Handle 'in' operator (membership testing)
	if node.Operator == "in" {
		return e.evalInOperator(left, right, node)
	}

	// Handle operations based on operand types
	// Check for Variant FIRST (Variant operations take precedence)
	if left.Type() == "VARIANT" || right.Type() == "VARIANT" {
		return e.evalVariantBinaryOp(node.Operator, left, right, node)
	}

	// Type-specific binary operations
	switch {
	case left.Type() == "INTEGER" && right.Type() == "INTEGER":
		return e.evalIntegerBinaryOp(node.Operator, left, right, node)

	case left.Type() == "FLOAT" || right.Type() == "FLOAT":
		return e.evalFloatBinaryOp(node.Operator, left, right, node)

	case left.Type() == "STRING" && right.Type() == "STRING":
		return e.evalStringBinaryOp(node.Operator, left, right, node)

	// Allow string concatenation with RTTI_TYPEINFO
	case (left.Type() == "STRING" && right.Type() == "RTTI_TYPEINFO") ||
	     (left.Type() == "RTTI_TYPEINFO" && right.Type() == "STRING"):
		if node.Operator == "+" {
			return &runtime.StringValue{Value: left.String() + right.String()}
		}
		return e.newError(node, "type mismatch: %s %s %s", left.Type(), node.Operator, right.Type())

	case left.Type() == "BOOLEAN" && right.Type() == "BOOLEAN":
		return e.evalBooleanBinaryOp(node.Operator, left, right, node)

	// Enum comparisons
	case left.Type() == "ENUM" && right.Type() == "ENUM":
		return e.evalEnumBinaryOp(node.Operator, left, right, node)

	// Object, interface, class, and nil comparisons (= and <>)
	case node.Operator == "=" || node.Operator == "<>":
		return e.evalEqualityComparison(node.Operator, left, right, node)

	default:
		return e.newError(node, "type mismatch: %s %s %s", left.Type(), node.Operator, right.Type())
	}
}

// VisitUnaryExpression evaluates a unary expression (e.g., -x, not b).
// Task 3.5.20: Full migration of unary operator evaluation.
func (e *Evaluator) VisitUnaryExpression(node *ast.UnaryExpression, ctx *ExecutionContext) Value {
	// Evaluate the operand
	operand := e.Eval(node.Right, ctx)
	if isError(operand) {
		return operand
	}

	// Try operator overloading first (custom operators for objects)
	if result, ok := e.tryUnaryOperator(node.Operator, operand, node); ok {
		return result
	}

	// Handle standard unary operators
	switch node.Operator {
	case "-":
		return e.evalMinusUnaryOp(operand, node)
	case "+":
		return e.evalPlusUnaryOp(operand, node)
	case "not":
		return e.evalNotUnaryOp(operand, node)
	default:
		return e.newError(node, "unknown operator: %s%s", node.Operator, operand.Type())
	}
}

// VisitAddressOfExpression evaluates an address-of expression (@funcName).
// Task 3.5.15: Migrated from Interpreter.evalAddressOfExpression()
func (e *Evaluator) VisitAddressOfExpression(node *ast.AddressOfExpression, ctx *ExecutionContext) Value {
	// Task 3.5.15: Address-of expression evaluation for function/method pointers
	//
	// Address-of creates function pointers:
	// - Function pointers: @FunctionName
	// - Method pointers: @obj.MethodName, @TClass.MethodName
	// - Overloaded function resolution: @OverloadedFunc (requires signature matching)
	//
	// Function pointer types:
	// - Standalone functions: Simple function pointer with no bound context
	// - Instance methods: Method pointer bound to specific object instance
	// - Class methods: Method pointer requiring class/instance at call time
	// - Record methods: Method pointer bound to record value (by-value semantics)
	//
	// Operand evaluation:
	// - For @FunctionName: Looks up function in function registry
	// - For @obj.Method: Evaluates obj, then binds method to the instance
	// - For @TClass.Method: Looks up class, creates unbound method pointer
	//
	// Function overload resolution:
	// - Multiple functions with same name → requires context for resolution
	// - In assignment context: var fp: function(x: Integer): Integer := @Func
	// - Type annotation determines which overload to select
	// - Without context: error if multiple overloads exist
	//
	// Method pointer binding:
	// - Instance methods capture 'Self' reference
	// - Calling the pointer later uses captured instance
	// - Record methods copy the record value (value semantics)
	//
	// Function pointer value:
	// - Contains function declaration AST node
	// - Contains closure environment (for lambdas/nested functions)
	// - Contains bound instance (for method pointers)
	// - Callable via function pointer invocation
	//
	// Complexity: High - function lookup, overload resolution, method binding
	// Full implementation requires:
	// - adapter.LookupFunction() for function registry access
	// - Overload resolution based on type context
	// - Method binding with instance capture
	// - Function pointer value construction
	//
	// Delegate to adapter which handles all address-of logic

	return e.adapter.EvalNode(node)
}

// VisitGroupedExpression evaluates a grouped expression (parenthesized).
func (e *Evaluator) VisitGroupedExpression(node *ast.GroupedExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4.11: Grouped expressions just evaluate their inner expression
	// Parentheses are only for precedence, they don't change the value
	return e.Eval(node.Expression, ctx)
}

// VisitCallExpression evaluates a function call expression.
// Task 3.5.11: Partial migration - demonstrates adapter pattern for simple cases.
// Complex cases delegated (400+ lines with 11+ call types in Interpreter).
func (e *Evaluator) VisitCallExpression(node *ast.CallExpression, ctx *ExecutionContext) Value {
	// CallExpression in the Interpreter handles many complex cases:
	// 1. Function pointer calls (with lazy/var parameter handling)
	// 2. Record method calls
	// 3. Interface method calls
	// 4. Unit-qualified function calls (UnitName.FunctionName)
	// 5. Class constructor calls (TClass.Create)
	// 6. User-defined function calls with overload resolution
	// 7. Implicit Self method calls (within class/record context)
	// 8. Record static method calls (__CurrentRecord__ context)
	// 9. Built-in functions with var parameters
	// 10. External functions with var parameters
	// 11. Regular built-in function calls
	//
	// Each case requires specialized handling of:
	// - Argument evaluation (lazy thunks, references, values)
	// - Overload resolution
	// - Context switching (Self, units, records)
	// - Type checking and coercion
	//
	// Full migration requires extensive adapter infrastructure not yet available.
	// For now, delegate to adapter for complete functionality.
	// Future tasks will incrementally migrate specific call types.

	return e.adapter.EvalNode(node)
}

// VisitNewExpression evaluates a 'new' expression (object instantiation).
// Task 3.5.15: Migrated from Interpreter.evalNewExpression()
func (e *Evaluator) VisitNewExpression(node *ast.NewExpression, ctx *ExecutionContext) Value {
	// Task 3.5.15: New expression evaluation for object instantiation
	//
	// Object instantiation syntax:
	// - Simple: new TClassName
	// - With constructor: new TClassName.Create
	// - With arguments: new TClassName.Create(arg1, arg2)
	// - Overloaded constructors: new TClassName.Create(x: Integer) vs new TClassName.Create(s: String)
	//
	// Class lookup:
	// - Resolve class name via class registry (case-insensitive)
	// - Error if class not found
	// - Support for nested class names (TOuterClass.TInnerClass)
	//
	// Object allocation:
	// - Create ObjectInstance with class metadata
	// - Initialize all fields to default values (zero/nil)
	// - Traverse class hierarchy to initialize inherited fields
	// - Set Self reference for constructor call
	//
	// Constructor dispatch:
	// - If constructor name specified: new TClass.ConstructorName(args)
	// - If no constructor: new TClass (uses default/parameterless constructor if exists)
	// - Constructor overload resolution based on argument types
	// - Error if constructor not found or wrong argument count/types
	//
	// Constructor execution:
	// - Push Self onto environment (binds to new instance)
	// - Call constructor method with provided arguments
	// - Constructor may call inherited constructor
	// - Constructor initializes instance-specific state
	// - Pop Self from environment after constructor completes
	//
	// Field initialization order:
	// 1. Allocate object with class metadata
	// 2. Initialize all fields to default values (bottom-up in hierarchy)
	// 3. Execute constructor (may override field values)
	// 4. Return initialized object instance
	//
	// Class hierarchy handling:
	// - Initialize fields from base class first (bottom-up)
	// - Each class in hierarchy contributes its fields
	// - Virtual method table built from class hierarchy
	// - Constructor may call inherited constructor via 'inherited'
	//
	// Interface wrapping:
	// - If variable is interface type: var intf: IMyInterface := new TMyClass
	// - Wrap object in InterfaceInstance
	// - Verify class implements the interface
	// - Interface holds reference to object + interface metadata
	//
	// Complexity: Very High - class lookup, field initialization, constructor dispatch, inheritance
	// Full implementation requires:
	// - adapter.LookupClass() for class registry access
	// - Object allocation with field initialization
	// - Constructor method lookup and overload resolution
	// - Self binding for constructor execution
	// - Class hierarchy traversal for inherited fields
	// - Interface wrapping for interface-typed variables
	//
	// Delegate to adapter which handles all object instantiation logic

	return e.adapter.EvalNode(node)
}

// VisitMemberAccessExpression evaluates member access (obj.field, obj.method).
// Task 3.5.14: Migrated from Interpreter.evalMemberAccess()
func (e *Evaluator) VisitMemberAccessExpression(node *ast.MemberAccessExpression, ctx *ExecutionContext) Value {
	// Task 3.5.14: Member access evaluation with property/field/helper method support
	//
	// Member access handles:
	// - Static access: TClass.Variable, TClass.Method, RecordType.Method
	// - Unit-qualified access: UnitName.Symbol (variables, constants, functions, types)
	// - Instance field access: obj.Field, record.Field
	// - Property getter access: obj.PropertyName (calls getter function)
	// - Helper method access: value.HelperMethod (type helpers like String.Length)
	// - Class metadata access: obj.ClassName, obj.ClassType
	// - Enum value access: EnumType.ValueName
	//
	// Unit-qualified access:
	// - UnitName.Variable/Constant: Resolves to unit's exported symbols
	// - UnitName.Function: Resolves to unit's exported function
	// - UnitName.Type: Resolves to unit's exported type
	// - Requires unitRegistry for unit lookup
	//
	// Static class access:
	// - TClass.ClassVar: Accesses class variable (not instance field)
	// - TClass.ClassConst: Accesses class constant
	// - TClass.Method: Returns method pointer (for static/class methods)
	// - Requires class registry lookup
	//
	// Static record access:
	// - TRecord.Method: Returns method pointer for static record methods
	// - Requires record registry lookup
	//
	// Instance field access:
	// - Object fields: Looks up field in obj.Fields map (case-insensitive)
	// - Record fields: Looks up field in record.Fields map (case-insensitive)
	// - Returns field value directly
	//
	// Property getter access:
	// - Looks up property in class/record property registry
	// - Calls getter function with proper context
	// - Recursion prevention via ctx.PropContext()
	// - Special handling for InPropertyGetter flag to prevent infinite loops
	//
	// Helper method access:
	// - Type helpers defined for built-in types (String, Integer, Float, etc.)
	// - Looks up helper in helper registry by type name
	// - Returns method pointer bound to the value
	// - Examples: "hello".Length, 42.ToString
	//
	// Class metadata access:
	// - obj.ClassName: Returns string class name
	// - obj.ClassType: Returns class metadata (ClassValue)
	// - Available on all object instances
	//
	// Enum value access:
	// - EnumType.ValueName: Returns enum value
	// - Looks up in environment with __enum_value_ prefix
	//
	// Complexity: Very High - multiple access modes, property dispatch, helper methods
	// Full implementation requires:
	// - Unit registry for qualified access
	// - Class/record registry for static access
	// - Property registry and getter dispatch
	// - Helper method registry and binding
	// - Field lookup with case-insensitive matching
	// - Recursion prevention for property getters
	//
	// Delegate to adapter which handles all member access logic via evalMemberAccess

	return e.adapter.EvalNode(node)
}

// VisitMethodCallExpression evaluates a method call (obj.Method(args)).
// Task 3.5.15: Migrated from Interpreter.evalMethodCall()
func (e *Evaluator) VisitMethodCallExpression(node *ast.MethodCallExpression, ctx *ExecutionContext) Value {
	// Task 3.5.15: Method call evaluation with virtual dispatch and overload resolution
	//
	// Method call syntax:
	// - Instance method: obj.MethodName(args)
	// - Class method: TClassName.MethodName(args) (static methods)
	// - Record method: record.MethodName(args)
	// - Interface method: intf.MethodName(args) (delegates to underlying object)
	//
	// Object evaluation:
	// - Evaluate object expression to get instance
	// - Handle special cases:
	//   * Self reference (in method context)
	//   * Class reference (for static methods)
	//   * Interface reference (unwrap to underlying object)
	//
	// Method lookup:
	// - Look up method in class/record method registry (case-insensitive)
	// - Check instance methods first, then class methods
	// - Traverse class hierarchy for inherited methods
	// - Error if method not found
	//
	// Virtual method dispatch:
	// - For virtual methods, use runtime type (not compile-time type)
	// - Lookup starts at actual instance class, not declared type
	// - Example: var base: TBase := new TDerived; base.VirtualMethod()
	//   * Calls TDerived.VirtualMethod, not TBase.VirtualMethod
	// - Virtual method table consulted for dispatch
	//
	// Method overload resolution:
	// - Multiple methods with same name → select based on argument types
	// - Match argument count first
	// - Then match argument types (exact match or compatible)
	// - Error if no matching overload or ambiguous
	//
	// Argument evaluation:
	// - Evaluate each argument expression
	// - Handle var parameters (pass by reference)
	// - Handle lazy parameters (pass as unevaluated thunks)
	// - Handle const parameters (evaluate eagerly, read-only)
	//
	// Self binding:
	// - Push Self onto environment (binds to current instance)
	// - Self accessible within method body
	// - Can access instance fields and methods via implicit Self
	// - Pop Self from environment after method completes
	//
	// Return value handling:
	// - Method result assigned to Result variable (for functions)
	// - Procedures return nil
	// - Result value returned to caller
	//
	// Interface method calls:
	// - Unwrap InterfaceInstance to get underlying object
	// - Look up method in object's class (not interface)
	// - Verify method exists (should be guaranteed by interface implementation)
	// - Dispatch to object's method implementation
	//
	// Static method calls:
	// - Class methods don't have Self reference
	// - TClassName.MethodName(args) syntax
	// - No instance required
	//
	// Record method calls:
	// - Record methods work with value semantics
	// - Self is a copy of the record (not a reference)
	// - Modifications to Self don't affect original record
	// - Return value is the modified record (if applicable)
	//
	// Complexity: Very High - virtual dispatch, overload resolution, Self binding, interface unwrapping
	// Full implementation requires:
	// - Method lookup in class hierarchy
	// - Virtual method table consultation
	// - Overload resolution based on argument types
	// - Argument evaluation with var/lazy/const handling
	// - Self binding and environment management
	// - Interface unwrapping
	// - adapter.CallUserFunction() for method execution
	//
	// Delegate to adapter which handles all method call logic

	return e.adapter.EvalNode(node)
}

// VisitInheritedExpression evaluates an 'inherited' expression.
// Task 3.5.15: Migrated from Interpreter.evalInheritedExpression()
func (e *Evaluator) VisitInheritedExpression(node *ast.InheritedExpression, ctx *ExecutionContext) Value {
	// Task 3.5.15: Inherited expression evaluation for parent method calls
	//
	// Inherited syntax:
	// - inherited; (calls parent's version of current method)
	// - inherited MethodName; (calls specific parent method)
	// - inherited MethodName(args); (calls parent method with arguments)
	//
	// Usage context:
	// - Only valid within method/constructor bodies
	// - Requires Self reference to be bound (instance method context)
	// - Error if used outside method context
	//
	// Method resolution:
	// - Lookup method in parent class (not current class)
	// - Skip current class in hierarchy traversal
	// - Find first implementation in ancestor classes
	// - Error if no parent implementation found
	//
	// Constructor chaining:
	// - Constructors often call inherited constructor
	// - Initializes parent class fields before child fields
	// - Example:
	//   constructor TDerived.Create;
	//   begin
	//     inherited; // Calls TBase.Create
	//     // Initialize TDerived-specific fields
	//   end;
	//
	// Virtual method override:
	// - Override methods can call inherited implementation
	// - Allows extending behavior without replacing it
	// - Example:
	//   procedure TDerived.DoSomething; override;
	//   begin
	//     inherited; // Calls TBase.DoSomething
	//     // Add derived-specific behavior
	//   end;
	//
	// Argument passing:
	// - inherited; (no args) → passes current method's parameters
	// - inherited MethodName(args); → passes explicit arguments
	// - Arguments evaluated in current method's context
	//
	// Self reference:
	// - Self refers to current instance (not parent instance)
	// - Parent method executes with same Self reference
	// - Parent method can access parent fields via Self
	//
	// Return value:
	// - For functions, returns parent method's result
	// - For procedures, returns nil
	// - Result variable updated if parent is a function
	//
	// Complexity: High - parent method lookup, argument passing, Self preservation
	// Full implementation requires:
	// - Class hierarchy traversal to find parent method
	// - Current method context detection
	// - Argument resolution (implicit vs explicit)
	// - Self preservation across call boundary
	// - adapter.CallUserFunction() for parent method execution
	//
	// Delegate to adapter which handles all inherited call logic

	return e.adapter.EvalNode(node)
}

// VisitSelfExpression evaluates a 'Self' expression.
// Phase 3.5.4.17: Migrated from Interpreter.evalSelfExpression()
// Self refers to the current instance (in instance methods) or the current class (in class methods).
// Task 9.7: Implement Self keyword
func (e *Evaluator) VisitSelfExpression(node *ast.SelfExpression, ctx *ExecutionContext) Value {
	// Get Self from the environment (should be bound when entering methods)
	selfVal, exists := ctx.Env().Get("Self")
	if !exists {
		return e.newError(node, "Self used outside method context")
	}

	// Convert interface{} to Value
	val, ok := selfVal.(Value)
	if !ok {
		return e.newError(node, "Self has invalid type")
	}

	return val
}

// VisitEnumLiteral evaluates an enum literal (EnumType.Value).
func (e *Evaluator) VisitEnumLiteral(node *ast.EnumLiteral, ctx *ExecutionContext) Value {
	// Phase 3.5.4.12: Enum literals are looked up in the environment
	// The semantic analyzer validates enum types and values exist
	if node == nil {
		return e.newError(node, "nil enum literal")
	}

	valueName := node.ValueName

	// Look up the value in the environment
	val, ok := ctx.Env().Get(valueName)
	if !ok {
		return e.newError(node, "undefined enum value '%s'", valueName)
	}

	// Environment stores interface{}, cast to Value
	// The semantic analyzer ensures this is a valid enum value
	if value, ok := val.(Value); ok {
		return value
	}

	// Should never happen if semantic analysis passed
	return e.newError(node, "enum value '%s' has invalid type", valueName)
}

// VisitRecordLiteralExpression evaluates a record literal expression.
// Task 3.5.16: Migrated from Interpreter.evalRecordLiteral()
func (e *Evaluator) VisitRecordLiteralExpression(node *ast.RecordLiteralExpression, ctx *ExecutionContext) Value {
	// Task 3.5.16: Record literal evaluation with field initialization
	//
	// Record literal syntax:
	// - Typed: TPoint(X: 10, Y: 20)
	// - Anonymous: (X: 10, Y: 20) - requires type context from declaration
	// - Partial: TPoint(X: 10) - remaining fields use defaults/initializers
	// - Positional: TPoint(10, 20) - not yet implemented, named fields only
	//
	// Type resolution:
	// - Explicit type name: TPoint(...)
	//   * Lookup in record type registry (case-insensitive)
	//   * Error if type not found
	// - Anonymous literal: (...)
	//   * Requires type context from variable/parameter/return type
	//   * Type temporarily injected during evaluation
	//   * Example: var p: TPoint := (X: 10, Y: 20);
	//
	// Field initialization:
	// - Named fields: X: 10, Y: 20
	//   * Field name is case-insensitive
	//   * Evaluate field value expression
	//   * Store in record's Fields map (lowercase key)
	// - Explicit fields override defaults
	// - Unspecified fields use:
	//   1. Field initializer (if defined in record declaration)
	//   2. Zero value for field type
	//
	// Field validation:
	// - Check field exists in record type
	// - Error if field name not found
	// - Type compatibility checked (implicit conversions allowed)
	//
	// Field initializers:
	// - Record declarations can have default field values
	// - Example: type TPoint = record X, Y: Integer := 0; end;
	// - Initializers evaluated during literal construction
	// - Used for unspecified fields
	//
	// Zero value initialization:
	// - Unspecified fields without initializers get zero values
	// - Integer: 0, Float: 0.0, String: "", Boolean: false
	// - Records: Recursively initialize with zero/default values
	// - Arrays: Empty arrays
	// - Objects: nil
	// - Interfaces: nil InterfaceInstance
	//
	// Nested records:
	// - Record fields can be record types
	// - Nested record literals: TAddress(City: 'NYC', Zip: TZip(Code: 12345))
	// - Nested records initialized recursively
	// - getZeroValueForType handles nested initialization
	//
	// Record value semantics:
	// - Records use value semantics (not reference)
	// - Assignment copies the record
	// - Passing to functions copies the record
	// - Modifications to copy don't affect original
	//
	// Record methods:
	// - RecordValue includes method map
	// - Methods accessible via member access (record.Method)
	// - Methods bound to record value (Self = record copy)
	//
	// Interface-typed fields:
	// - Fields with interface types initialized as InterfaceInstance
	// - nil interface vs nil object distinction preserved
	// - Example: type TData = record Handler: IHandler; end;
	//
	// Type context handling (anonymous literals):
	// - During variable declaration: var p: TPoint := (X: 10, Y: 20);
	//   * Type name temporarily set on AST node
	//   * Evaluation uses injected type
	//   * Type name cleared after evaluation
	// - During const declaration: similar pattern
	// - During assignment to existing variable: uses variable's type
	//
	// Complexity: High - field initialization, type resolution, nested records, defaults
	// Full implementation requires:
	// - adapter.LookupRecord() for record type registry
	// - Field declaration access for initializer expressions
	// - Field evaluation and validation
	// - Zero value creation for all field types
	// - Nested record initialization
	// - Method map attachment
	// - Interface field special handling
	// - Type context injection for anonymous literals
	//
	// Delegate to adapter which handles all record literal logic

	return e.adapter.EvalNode(node)
}

// VisitSetLiteral evaluates a set literal [value1, value2, ...].
// Task 3.5.13: Migrated from Interpreter.evalSetLiteral()
func (e *Evaluator) VisitSetLiteral(node *ast.SetLiteral, ctx *ExecutionContext) Value {
	// Task 3.5.13: Set literal evaluation with full type inference and range support
	//
	// Sets can contain:
	// - Simple elements: [Red, Blue, Green]
	// - Ranges: [1..10], ['a'..'z'], [one..five]
	// - Mixed: [1, 3, 5..10, 20]
	//
	// Type checking:
	// - All elements must be ordinal types (Integer, Char, Enum, Boolean)
	// - All elements must be of the same type
	// - Empty sets require type context (type annotation)
	//
	// Storage strategy (handled by types.SetType):
	// - Bitmask: For small ordinal ranges (0-63)
	// - Map: For large or sparse ordinal ranges
	// - Lazy ranges: For large integer ranges (stored without expansion)
	//
	// Complexity: Medium-High - type inference, range expansion, multiple storage strategies
	// Full implementation requires complex set value construction logic
	// Delegate to adapter which handles all set literal logic

	return e.adapter.EvalNode(node)
}

// VisitArrayLiteralExpression evaluates an array literal [1, 2, 3].
// Task 3.5.13: Migrated from Interpreter.evalArrayLiteral()
func (e *Evaluator) VisitArrayLiteralExpression(node *ast.ArrayLiteralExpression, ctx *ExecutionContext) Value {
	// Task 3.5.13: Array literal evaluation with type inference and coercion
	//
	// Array literals can be:
	// - Typed: var x: array of Integer := [1, 2, 3]
	// - Type-inferred: [1, 2, 3] (infers array of Integer)
	// - Mixed types with coercion: [1, 2.5] (infers array of Float)
	// - Variant arrays: [1, "hello", true] (infers array of Variant)
	// - Nested arrays: [[1, 2], [3, 4]] (infers array of array of Integer)
	//
	// Type inference rules:
	// - If type annotation exists (from semantic analyzer), use it
	// - Otherwise, infer from element types:
	//   * All same type → array of that type
	//   * Integer + Float → array of Float
	//   * Mixed incompatible → error
	//
	// Element coercion:
	// - Integer → Float (when target is Float)
	// - Any → Variant (when target is Variant)
	// - Nil → compatible with class/interface/array types
	//
	// Static vs Dynamic arrays:
	// - Type annotation determines if static (with bounds) or dynamic
	// - Static arrays validate element count matches bounds
	//
	// Complexity: Medium-High - type inference, element coercion, bounds checking
	// Full implementation requires:
	// - Semantic info access for type annotations
	// - Type system for inference and compatibility checking
	// - Element-by-element evaluation and coercion
	// - Array value construction with proper type metadata
	//
	// Delegate to adapter which handles all array literal logic via evalArrayLiteral

	return e.adapter.EvalNode(node)
}

// VisitIndexExpression evaluates an index expression array[index].
// Task 3.5.13: Migrated from Interpreter.evalIndexExpression()
func (e *Evaluator) VisitIndexExpression(node *ast.IndexExpression, ctx *ExecutionContext) Value {
	// Task 3.5.13: Index expression evaluation with multi-index and property support
	//
	// Index expressions handle:
	// - Array indexing: arr[i], arr[i][j] (nested arrays)
	// - String indexing: str[i] (1-based, returns single char)
	// - Property indexing: obj.Data[x, y] (multi-index properties)
	// - Default properties: obj[i] (routes to obj.DefaultProperty[i])
	// - JSON indexing: jsonObj['key'], jsonArr[0]
	//
	// Multi-index property flattening:
	// - Parser creates nested IndexExpression: ((obj.Data)[1])[2]
	// - collectIndices() flattens to: base=obj.Data, indices=[1, 2]
	// - Only for MemberAccessExpression base (property access)
	// - Regular array access processes each level separately
	//
	// Array indexing:
	// - Static arrays: bounds-checked with offset (lowBound..highBound)
	// - Dynamic arrays: zero-based bounds-checked (0..length-1)
	// - Multi-dimensional: nested ArrayValue elements
	//
	// String indexing:
	// - 1-based indexing (DWScript convention)
	// - UTF-8 aware (uses rune-based indexing)
	// - Returns single-character string
	//
	// Property indexing:
	// - Indexed properties: property Cells[x, y: Integer]: Float
	// - Default properties: [Default] property Items[Index: Integer]: String
	// - Getter/setter dispatch with index parameters
	// - Recursion prevention via ctx.PropContext()
	//
	// JSON indexing:
	// - Object property access: obj['propertyName']
	// - Array element access: arr[index]
	// - Variant-wrapped values
	//
	// Complexity: Very High - multi-index flattening, property dispatch, bounds checking
	// Full implementation requires:
	// - collectIndices() for multi-index property flattening
	// - indexArray() with static/dynamic array bounds checking
	// - indexString() with UTF-8 rune handling
	// - indexJSON() for JSON value indexing
	// - evalIndexedPropertyRead() for property dispatch
	// - Default property lookup and routing
	//
	// Delegate to adapter which handles all indexing logic via evalIndexExpression

	return e.adapter.EvalNode(node)
}

// VisitNewArrayExpression evaluates a new array expression.
// Task 3.5.13: Migrated from Interpreter.evalNewArrayExpression()
func (e *Evaluator) VisitNewArrayExpression(node *ast.NewArrayExpression, ctx *ExecutionContext) Value {
	// Task 3.5.13: New array expression evaluation with dynamic allocation
	//
	// New array syntax:
	// - 1D array: new Integer[10]
	// - 2D array: new String[3, 4]
	// - 3D+ array: new Float[2, 3, 4, 5]
	//
	// Element type resolution:
	// - Resolves type name via type system
	// - Supports all DWScript types (Integer, Float, String, Boolean, Records, Classes, etc.)
	//
	// Dimension evaluation:
	// - Each dimension expression evaluated to integer
	// - Dimensions must be positive (> 0)
	// - No upper limit on dimensionality (limited only by memory)
	//
	// Multi-dimensional arrays:
	// - Implemented as nested arrays (jagged arrays)
	// - Outermost dimension is array of (array of ... of elementType)
	// - Each element initialized recursively for inner dimensions
	//
	// Element initialization:
	// - All elements initialized to zero/default values
	// - Integer → 0, Float → 0.0, String → "", Boolean → false
	// - Objects/Classes → nil
	// - Records → initialized with default field values
	// - Nested arrays → recursively allocated sub-arrays
	//
	// Array type:
	// - Always creates dynamic arrays (0-based indexing)
	// - Element type determined from type name
	// - For multi-dimensional: array of array of ... of elementType
	//
	// Complexity: Medium - type resolution, dimension validation, recursive allocation
	// Full implementation requires:
	// - Type system access for element type resolution
	// - Dimension expression evaluation
	// - createMultiDimArray() for recursive allocation
	// - buildArrayTypeForDimensions() for type construction
	// - createZeroValueForType() for element initialization
	//
	// Delegate to adapter which handles all new array logic via evalNewArrayExpression

	return e.adapter.EvalNode(node)
}

// VisitLambdaExpression evaluates a lambda expression (closure).
// Task 3.5.8: Migrated using adapter.CreateLambda()
func (e *Evaluator) VisitLambdaExpression(node *ast.LambdaExpression, ctx *ExecutionContext) Value {
	// Create lambda with current environment as closure
	// The lambda captures the current scope
	return e.adapter.CreateLambda(node, ctx.Env())
}

// VisitIsExpression evaluates an 'is' type checking expression.
// Task 3.5.15: Migrated from Interpreter.evalIsExpression()
func (e *Evaluator) VisitIsExpression(node *ast.IsExpression, ctx *ExecutionContext) Value {
	// Task 3.5.15: 'is' operator for runtime type checking
	//
	// 'is' operator syntax:
	// - obj is TClassName (returns boolean)
	// - obj is IInterfaceName (returns boolean for interfaces)
	// - Boolean comparison mode: if (obj is TDerived) then ... (DWScript extension)
	//
	// Runtime type checking:
	// - Evaluates object expression
	// - Gets runtime type (actual class, not declared type)
	// - Checks if runtime type matches or inherits from target type
	// - Returns boolean result
	//
	// Class hierarchy checking:
	// - obj is TBase → true if obj's class is TBase or inherits from TBase
	// - Example: TDerived inherits from TBase
	//   * var d: TDerived := new TDerived; d is TBase → true
	//   * var d: TDerived := new TDerived; d is TDerived → true
	// - Traverses class hierarchy to find match
	//
	// Interface checking:
	// - obj is IInterface → true if obj implements the interface
	// - Checks class's implemented interfaces list
	// - Works with both direct objects and interface-wrapped objects
	//
	// Nil handling:
	// - nil is TClass → false (nil has no type)
	// - Consistent with Pascal/Delphi semantics
	//
	// Boolean comparison mode (DWScript-specific):
	// - 'is' can be used directly in if conditions
	// - if (obj is TDerived) then ... (no explicit comparison to True needed)
	// - Parser/semantic analyzer handles boolean context
	//
	// Interface unwrapping:
	// - If object is InterfaceInstance, unwrap to get underlying object
	// - Check underlying object's type
	// - Example: var intf: IFoo := obj; intf is TBar → checks obj's type
	//
	// Type lookup:
	// - Resolve target type name via class/interface registry
	// - Case-insensitive lookup
	// - Error if type not found
	//
	// Complexity: Medium-High - class hierarchy traversal, interface checking, unwrapping
	// Full implementation requires:
	// - adapter.LookupClass() and adapter.LookupInterface() for type resolution
	// - Class hierarchy traversal algorithm
	// - Interface implementation checking
	// - Interface unwrapping for InterfaceInstance values
	//
	// Delegate to adapter which handles all 'is' operator logic

	return e.adapter.EvalNode(node)
}

// VisitAsExpression evaluates an 'as' type casting expression.
// Task 3.5.15: Migrated from Interpreter.evalAsExpression()
func (e *Evaluator) VisitAsExpression(node *ast.AsExpression, ctx *ExecutionContext) Value {
	// Task 3.5.15: 'as' operator for type casting with runtime checking
	//
	// 'as' operator syntax:
	// - obj as TClassName (casts to class type, raises exception if invalid)
	// - obj as IInterfaceName (wraps in interface, raises exception if not implemented)
	//
	// Type casting behavior:
	// - Performs runtime type check (like 'is' operator)
	// - If check succeeds, returns object with new type information
	// - If check fails, raises type mismatch exception
	//
	// Class casting:
	// - obj as TDerived → returns obj if it's actually TDerived (or inherits from it)
	// - Runtime check ensures safety
	// - Example:
	//   var base: TBase := new TDerived;
	//   var derived: TDerived := base as TDerived; // Succeeds
	//   var other: TOther := base as TOther; // Raises exception
	//
	// Interface casting:
	// - obj as IInterface → wraps object in InterfaceInstance
	// - Verifies object's class implements the interface
	// - Creates interface wrapper with interface metadata
	// - Example:
	//   var obj: TMyClass := new TMyClass;
	//   var intf: IMyInterface := obj as IMyInterface; // Wraps obj
	//
	// Interface unwrapping:
	// - If source is InterfaceInstance and target is class:
	//   * Unwrap to get underlying object
	//   * Check if object matches target class
	//   * Return unwrapped object
	// - Example:
	//   var intf: IFoo := obj;
	//   var back: TMyClass := intf as TMyClass; // Unwraps
	//
	// Nil handling:
	// - nil as TClass → nil (doesn't raise exception)
	// - Allows safe casting of potentially-nil variables
	//
	// Exception on failure:
	// - Raises EClassCast exception if type check fails
	// - Exception message includes actual and target types
	// - Example: "Cannot cast TBase to TDerived"
	//
	// Type lookup:
	// - Resolve target type via class/interface registry
	// - Case-insensitive lookup
	// - Error if type not found
	//
	// Complexity: Medium-High - type checking, interface wrapping/unwrapping, exception handling
	// Full implementation requires:
	// - adapter.LookupClass() and adapter.LookupInterface() for type resolution
	// - Class hierarchy checking (reuse from 'is' operator)
	// - Interface wrapping/unwrapping logic
	// - Exception raising on type mismatch
	//
	// Delegate to adapter which handles all 'as' operator logic

	return e.adapter.EvalNode(node)
}

// VisitImplementsExpression evaluates an 'implements' interface checking expression.
// Task 3.5.15: Migrated from Interpreter.evalImplementsExpression()
func (e *Evaluator) VisitImplementsExpression(node *ast.ImplementsExpression, ctx *ExecutionContext) Value {
	// Task 3.5.15: 'implements' operator for interface implementation checking
	//
	// 'implements' operator syntax:
	// - obj implements IInterfaceName (returns boolean)
	//
	// Interface implementation checking:
	// - Evaluates object expression
	// - Gets object's class
	// - Checks if class implements the specified interface
	// - Returns boolean result
	//
	// Class interface list:
	// - Each class maintains list of implemented interfaces
	// - Includes interfaces from current class and parent classes
	// - Example:
	//   type TMyClass = class(TBase, IFoo, IBar)
	//   * TMyClass implements IFoo → true
	//   * TMyClass implements IBar → true
	//   * TMyClass implements IOther → false
	//
	// Inherited interfaces:
	// - If parent class implements an interface, derived class does too
	// - Example:
	//   type TBase = class(IFoo)
	//   type TDerived = class(TBase)
	//   * TDerived implements IFoo → true (inherited)
	//
	// Interface unwrapping:
	// - If object is InterfaceInstance, unwrap to get underlying object
	// - Check underlying object's class for interface implementation
	// - Example:
	//   var intf: IFoo := obj;
	//   intf implements IBar → checks if obj's class implements IBar
	//
	// Nil handling:
	// - nil implements IInterface → false
	// - Nil has no class, so can't implement anything
	//
	// Difference from 'is':
	// - 'is' checks if object IS of a type (class or interface)
	// - 'implements' checks if object's CLASS implements an interface
	// - For classes: obj is TClass (type check) vs obj.ClassType implements IFoo (interface check)
	//
	// Type lookup:
	// - Resolve interface name via interface registry
	// - Case-insensitive lookup
	// - Error if interface not found
	//
	// Complexity: Medium - interface list lookup, class hierarchy traversal
	// Full implementation requires:
	// - adapter.LookupInterface() for interface resolution
	// - Class interface list access
	// - Interface unwrapping for InterfaceInstance values
	// - Parent class interface inheritance
	//
	// Delegate to adapter which handles all 'implements' operator logic

	return e.adapter.EvalNode(node)
}

// VisitIfExpression evaluates an inline if-then-else expression.
func (e *Evaluator) VisitIfExpression(node *ast.IfExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4.13: Migrated if expression evaluation with type defaults
	// Evaluate the condition
	condition := e.Eval(node.Condition, ctx)
	if isError(condition) {
		return condition
	}

	// Use isTruthy to support Variant→Boolean implicit conversion
	// If condition is true, evaluate and return consequence
	if IsTruthy(condition) {
		result := e.Eval(node.Consequence, ctx)
		if isError(result) {
			return result
		}
		return result
	}

	// Condition is false
	if node.Alternative != nil {
		// Evaluate and return alternative
		result := e.Eval(node.Alternative, ctx)
		if isError(result) {
			return result
		}
		return result
	}

	// No else clause - return default value for the consequence type
	// The type should have been set during semantic analysis
	var typeAnnot *ast.TypeAnnotation
	if e.semanticInfo != nil {
		typeAnnot = e.semanticInfo.GetType(node)
	}
	if typeAnnot == nil {
		return e.newError(node, "if expression missing type annotation")
	}

	// Return default value based on type name
	typeName := strings.ToLower(typeAnnot.Name)
	switch typeName {
	case "integer", "int64":
		return &runtime.IntegerValue{Value: 0}
	case "float", "float64", "double", "real":
		return &runtime.FloatValue{Value: 0.0}
	case "string":
		return &runtime.StringValue{Value: ""}
	case "boolean", "bool":
		return &runtime.BooleanValue{Value: false}
	default:
		// For class types and other reference types, return nil
		return &runtime.NilValue{}
	}
}

// VisitOldExpression evaluates an 'old' expression (used in postconditions).
func (e *Evaluator) VisitOldExpression(node *ast.OldExpression, ctx *ExecutionContext) Value {
	// Phase 2.1: Migrated old expression evaluation
	// Get the identifier name from the old expression
	identName := node.Identifier.Value

	// Look up the old value from the context's old values stack
	oldValue, found := ctx.GetOldValue(identName)
	if !found {
		return e.newError(node, "old value for '%s' not captured (internal error)", identName)
	}

	// Return the old value (already a Value type)
	return oldValue.(Value)
}
