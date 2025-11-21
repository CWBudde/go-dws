package evaluator

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// This file contains visitor methods for expression AST nodes.
// Visitor pattern implementation for expressions.
//
// Expressions evaluate to values and can be nested (e.g., binary expressions
// contain left and right sub-expressions).

// ErrorValue represents a runtime error.
type ErrorValue struct {
	Message string
}

func (e *ErrorValue) Type() string   { return "ERROR" }
func (e *ErrorValue) String() string { return "ERROR: " + e.Message }

// newError creates a new error value with optional formatting.
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
func (e *Evaluator) VisitIdentifier(node *ast.Identifier, ctx *ExecutionContext) Value {
	// Self keyword refers to current object instance
	if node.Value == "Self" {
		val, ok := ctx.Env().Get("Self")
		if !ok {
			return e.newError(node, "Self used outside method context")
		}
		// Environment stores interface{}, cast to Value
		if selfVal, ok := val.(Value); ok {
			return selfVal
		}
		return e.newError(node, "Self has invalid type")
	}

	// Try to find identifier in current environment (variables, parameters, constants)
	val, ok := e.adapter.GetVariable(node.Value, ctx)
	if ok {
		// Variable found - return immediately for basic primitives
		switch val.(type) {
		case *runtime.IntegerValue, *runtime.FloatValue, *runtime.StringValue, *runtime.BooleanValue, *runtime.NilValue:
			return val
		}

		// For complex value types (ExternalVarValue, LazyThunk, ReferenceValue, arrays, objects, records),
		// delegate to adapter for full processing
		return e.adapter.EvalNode(node)
	}

	// Check if we're in an instance method context (Self is bound)
	// When Self is bound, identifiers can refer to instance fields, class variables,
	// properties, methods (auto-invoked if zero params), or ClassName/ClassType
	if selfRaw, selfOk := ctx.Env().Get("Self"); selfOk {
		if _, ok := selfRaw.(Value); ok {
			return e.adapter.EvalNode(node)
		}
	}

	// Check if we're in a class method context (__CurrentClass__ is bound)
	// Identifiers can refer to ClassName, ClassType, or class variables
	if currentClassRaw, hasCurrentClass := ctx.Env().Get("__CurrentClass__"); hasCurrentClass {
		if _, ok := currentClassRaw.(Value); ok {
			return e.adapter.EvalNode(node)
		}
	}

	// Check if this identifier is a user-defined function name
	// Functions are auto-invoked if they have zero parameters, or converted to function pointers if they have parameters
	funcNameLower := strings.ToLower(node.Value)
	if overloads, exists := e.adapter.LookupFunction(funcNameLower); exists && len(overloads) > 0 {
		return e.adapter.EvalNode(node)
	}

	// Check if this identifier is a class name (metaclass reference)
	if e.adapter.HasClass(node.Value) {
		return e.adapter.EvalNode(node)
	}

	// Final check: delegate to adapter for built-in functions or error if undefined
	return e.adapter.EvalNode(node)
}

// VisitBinaryExpression evaluates a binary expression (e.g., a + b, x == y).
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
// Creates function/method pointers that can be called later or assigned to variables.
func (e *Evaluator) VisitAddressOfExpression(node *ast.AddressOfExpression, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitGroupedExpression evaluates a grouped expression (parenthesized).
func (e *Evaluator) VisitGroupedExpression(node *ast.GroupedExpression, ctx *ExecutionContext) Value {
	// Grouped expressions just evaluate their inner expression
	// Parentheses are only for precedence, they don't change the value
	return e.Eval(node.Expression, ctx)
}

// VisitCallExpression evaluates a function call expression.
func (e *Evaluator) VisitCallExpression(node *ast.CallExpression, ctx *ExecutionContext) Value {
	// Check for function pointer calls (require special handling for lazy and var parameters)
	if funcIdent, ok := node.Function.(*ast.Identifier); ok {
		if val, exists := e.adapter.GetVariable(funcIdent.Value, ctx); exists {
			if val.Type() == "FUNCTION_POINTER" || val.Type() == "LAMBDA" {
				return e.adapter.EvalNode(node)
			}
		}
	}

	// Check for member access calls: obj.Method(), UnitName.Func(), TClass.Create()
	if memberAccess, ok := node.Function.(*ast.MemberAccessExpression); ok {
		objVal := e.Eval(memberAccess.Object, ctx)
		if isError(objVal) {
			return objVal
		}

		// Delegate record, interface, and object method calls to adapter
		if objVal.Type() == "RECORD" || objVal.Type() == "INTERFACE" || objVal.Type() == "OBJECT" {
			return e.adapter.EvalNode(node)
		}

		// Check for unit-qualified or class constructor calls
		if ident, ok := memberAccess.Object.(*ast.Identifier); ok {
			if e.unitRegistry != nil || e.adapter.HasClass(ident.Value) {
				return e.adapter.EvalNode(node)
			}
		}

		return e.newError(node, "cannot call member expression that is not a method or unit-qualified function")
	}

	// Remaining call types require a simple identifier
	funcName, ok := node.Function.(*ast.Identifier)
	if !ok {
		return e.newError(node, "function call requires identifier or qualified name, got %T", node.Function)
	}

	// Check for user-defined functions (with potential overloading)
	funcNameLower := strings.ToLower(funcName.Value)
	if overloads, exists := e.adapter.LookupFunction(funcNameLower); exists && len(overloads) > 0 {
		return e.adapter.EvalNode(node)
	}

	// Check for implicit Self method calls (MethodName() is shorthand for Self.MethodName())
	if selfRaw, ok := ctx.Env().Get("Self"); ok {
		if selfVal, ok := selfRaw.(Value); ok {
			if selfVal.Type() == "OBJECT" || selfVal.Type() == "CLASS" {
				return e.adapter.EvalNode(node)
			}
		}
	}

	// Check for record static method calls
	if recordRaw, ok := ctx.Env().Get("__CurrentRecord__"); ok {
		if recordVal, ok := recordRaw.(Value); ok {
			if recordVal.Type() == "RECORD_TYPE" {
				return e.adapter.EvalNode(node)
			}
		}
	}

	// Check for built-in functions that need var parameter handling (modify arguments in place)
	switch funcNameLower {
	case "inc", "dec", "insert", "decodedate", "decodetime",
		"swap", "divmod", "trystrtoint", "trystrtofloat", "setlength":
		return e.adapter.EvalNode(node)
	case "delete":
		// Only the 3-parameter form needs var parameter handling
		if len(node.Arguments) == 3 {
			return e.adapter.EvalNode(node)
		}
	}

	// Check for external (Go) functions that may need var parameter handling
	if e.externalFunctions != nil {
		return e.adapter.EvalNode(node)
	}

	// Check for Default(TypeName) function which expects unevaluated type identifier
	if funcNameLower == "default" && len(node.Arguments) == 1 {
		return e.adapter.EvalNode(node)
	}

	// Try type cast for single-argument calls: TypeName(expression)
	if len(node.Arguments) == 1 {
		result := e.adapter.EvalNode(node)
		// If type cast succeeded or there's a real error (not "unknown function"), return it
		if result != nil && !isError(result) {
			return result
		}
		if isError(result) {
			if !strings.Contains(result.String(), "unknown function") &&
				!strings.Contains(result.String(), "undefined identifier") {
				return result
			}
		}
	}

	// Try built-in functions - evaluate all arguments first
	args := make([]Value, len(node.Arguments))
	for idx, arg := range node.Arguments {
		val := e.Eval(arg, ctx)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	// Call built-in function via adapter
	return e.adapter.CallBuiltinFunction(funcName.Value, args)
}

// VisitNewExpression evaluates a 'new' expression (object instantiation).
// Handles class lookup, field initialization, constructor dispatch, and interface wrapping.
func (e *Evaluator) VisitNewExpression(node *ast.NewExpression, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitMemberAccessExpression evaluates member access (obj.field, obj.method).
//
// **COMPLEXITY**: Very High (700+ lines in original implementation)
// **STATUS**: Documentation-only migration with full adapter delegation
//
// **11 DISTINCT ACCESS MODES** (evaluated in this order):
//
// **1. UNIT-QUALIFIED ACCESS** (UnitName.Symbol)
//   - Pattern: `Math.PI`, `System.Print`
//   - Evaluation order:
//     a. Check if left side is a registered unit name (via unitRegistry)
//     b. Try to resolve as qualified variable/constant (ResolveQualifiedVariable)
//     c. If not a variable, it might be a function reference (handled in VisitCallExpression)
//   - Error: "qualified name 'Unit.Symbol' cannot be used as a value (functions must be called)"
//   - Implementation: ~14 lines in original
//
// **2. STATIC CLASS ACCESS** (TClass.Member)
//   - Pattern: `TMyClass.ClassVar`, `TMyClass.Create`, `TMyClass.ClassName`
//   - Lookup order (case-insensitive):
//     a. Built-in properties: `ClassName` (string), `ClassType` (metaclass reference)
//     b. Class variables (lookupClassVar) - inherited from parent classes
//     c. Class constants (getClassConstant) - lazy evaluation with caching
//     d. Class properties (lookupProperty) - if IsClassProperty or uses class-level read specs
//     e. Constructors (HasConstructor) - auto-invoke with 0 arguments if no parentheses
//   - Supports constructor overloading and inheritance
//   - Falls back to implicit parameterless constructor if needed
//     f. Class methods (lookupClassMethodInHierarchy) - static methods
//   - Parameterless: auto-invoke via VisitMethodCallExpression
//   - With parameters: return as FunctionPointerValue
//   - Error: "member 'X' not found in class 'Y'"
//   - Implementation: ~100 lines in original
//
// **3. ENUM TYPE ACCESS** (TColor.Red, TColor.Low, TColor.High)
//   - Pattern: `TColor.Red`, `TMyEnum.Low`, `TMyEnum.High`
//   - Lookup in environment: `__enum_type_` + lowercase(enumTypeName)
//   - For scoped enums:
//     a. Look up enum value in EnumType.Values (takes precedence over properties)
//     b. Check for special properties: `Low` (lowest ordinal), `High` (highest ordinal)
//   - For unscoped enums: also check environment for value name
//   - Returns: EnumValue or IntegerValue (for Low/High)
//   - Error: "enum value 'X' not found in enum 'Y'"
//   - Implementation: ~45 lines in original
//
// **4. RECORD TYPE STATIC ACCESS** (TPoint.cOrigin, TPoint.Count)
//   - Pattern: `TPoint.cOrigin`, `TRecord.ClassMethod()`
//   - Lookup in environment: `__record_type_` + lowercase(recordTypeName)
//   - Lookup order (case-insensitive):
//     a. Constants (RecordTypeValue.Constants)
//     b. Class variables (RecordTypeValue.ClassVars)
//     c. Class methods (RecordTypeValue.ClassMethods)
//   - Parameterless: auto-invoke via VisitMethodCallExpression
//   - With parameters: error (requires parentheses to call)
//   - Error: "member 'X' not found in record type 'Y'"
//   - Implementation: ~40 lines in original
//
// **5. RECORD INSTANCE ACCESS** (record.Field, record.Method)
//   - Pattern: `point.X`, `point.GetLength()`, `point.Prop`
//   - Object type: RecordValue
//   - Lookup order (case-insensitive):
//     a. Direct field access (RecordValue.Fields)
//     b. Properties (RecordType.Properties):
//   - ReadField: field name → direct access, method name → call getter
//   - Write-only: error "property 'X' is write-only"
//     c. Instance methods (RecordValue.Methods):
//   - Parameterless: auto-invoke via VisitMethodCallExpression
//   - With parameters: error "method 'X' requires N parameter(s); use parentheses"
//     d. Class methods (from RecordTypeValue, accessible via instance)
//     e. Constants (from RecordTypeValue, accessible via instance)
//     f. Class variables (from RecordTypeValue, accessible via instance)
//     g. Helper properties (findHelperProperty → evalHelperPropertyRead)
//   - Error: "field 'X' not found in record 'Y'"
//   - Implementation: ~115 lines in original
//
// **6. CLASS/METACLASS ACCESS** (ClassInfoValue/ClassValue.Member)
//   - Pattern: When a class name is evaluated to ClassInfoValue or ClassValue
//   - Example: `var c := TMyClass; c.Create()`
//   - Lookup order (same as static class access #2):
//     a. Built-in properties: `ClassName`, `ClassType`
//     b. Class variables, constants, properties, constructors, class methods
//   - Returns: String/ClassValue/field value/method pointer
//   - Implementation: ~95 lines in original
//
// **7. INTERFACE INSTANCE ACCESS** (interface.Method, interface.Property)
//   - Pattern: `intfVar.Hello`, `intfVar.SomeMethod`
//   - Object type: InterfaceInstance
//   - Validation: Verify member exists in interface definition (HasMethod)
//   - For methods:
//   - Look up implementation in underlying object's class (getMethodOverloadsInHierarchy)
//   - Return FunctionPointerValue bound to underlying object (NO auto-invoke)
//   - Enables method delegate assignment: `var h : procedure := i.Hello;`
//   - For properties/fields: delegate to underlying object (without validation currently)
//   - Unwrap interface to underlying object and continue evaluation
//   - Error: "Interface is nil" or "method 'X' declared in interface 'Y' but not implemented"
//   - Implementation: ~50 lines in original
//
// **8. TYPE CAST VALUE HANDLING** (TBase(child).ClassVar)
//   - Pattern: Accessing members through a type cast expression
//   - Object type: TypeCastValue
//   - Extracts: StaticType (for class variable lookup), Object (actual instance)
//   - Purpose: Class variables use static type, not runtime type
//   - Unwraps to actual object and continues evaluation with static type context
//   - Implementation: ~5 lines in original
//
// **9. NIL OBJECT HANDLING** (nil.ClassVar)
//   - Pattern: `var o: TMyClass := nil; o.ClassVar`
//   - Object type: NilValue (with ClassType field) or nil evaluation result
//   - Special case: Accessing class variables on nil instances is allowed
//   - Lookup:
//     a. If staticClassType from cast (TBase(nil).ClassVar): use static type
//     b. If NilValue.ClassType set: look up class and check for class variable
//   - Success: Return class variable value
//   - Failure: Error "Object not instantiated" (for instance members)
//   - Implementation: ~35 lines in original
//
// **10. ENUM VALUE PROPERTIES** (enumVal.Value)
//   - Pattern: `TColor.Red.Value` (returns ordinal as integer)
//   - Object type: EnumValue
//   - Supported properties:
//     a. `.Value` (case-insensitive): returns OrdinalValue as IntegerValue
//     b. `.ToString`: handled by helpers (if available)
//   - Fallback: Check helpers for additional properties
//   - Implementation: ~10 lines in original
//
// **11. OBJECT INSTANCE ACCESS** (obj.Field, obj.Method, obj.Property)
//   - Pattern: `myObj.Name`, `myObj.GetValue()`, `myObj.Count`
//   - Object type: ObjectInstance
//   - Built-in properties (inherited from TObject, case-insensitive):
//     a. `ClassName`: returns obj.Class.Name (runtime type)
//     b. `ClassType`: returns ClassValue (metaclass for runtime type)
//   - Lookup order (case-insensitive):
//     a. Properties (Class.lookupProperty) - takes precedence over fields
//   - Call evalPropertyRead for read accessor (field, method, or expression)
//     b. Direct field access (obj.GetField) - instance fields
//     c. Class variables (lookupClassVar) - accessible from instance
//   - Uses static type from cast if available (e.g., TBase(child).ClassVar)
//     d. Class constants (getClassConstant) - accessible from instance
//     e. Instance methods (getMethodOverloadsInHierarchy):
//   - Check all overloads for parameterless variants
//   - Parameterless: auto-invoke via VisitMethodCallExpression
//   - With parameters: return first overload as FunctionPointerValue
//     f. Class methods (getMethodOverloadsInHierarchy with classMethod=true)
//   - Same logic as instance methods (auto-invoke or function pointer)
//     g. Helper properties (findHelperProperty → evalHelperPropertyRead)
//   - Error: "field 'X' not found in class 'Y'"
//   - Implementation: ~115 lines in original
//
// **SPECIAL BEHAVIORS**:
// - **Auto-invocation**: Parameterless methods/properties auto-invoke when accessed without ()
// - **Case-insensitive**: All name lookups are case-insensitive (DWScript spec)
// - **Inheritance**: Class variables, constants, properties, methods searched up hierarchy
// - **Helper support**: Type helpers can add properties/methods to any type
// - **Function pointers**: Methods with parameters return FunctionPointerValue
// - **Lazy evaluation**: Class constants evaluated once and cached on first access
// - **Type safety**: Static types respected for class variable access through casts
//
// **DEPENDENCIES** (blockers for full migration):
// - RecordValue, RecordTypeValue - in internal/interp (needs migration to runtime)
// - ObjectInstance, ClassInfo, ClassValue - in internal/interp (needs migration to runtime)
// - InterfaceInstance - in internal/interp (needs migration to runtime)
// - EnumValue, EnumTypeValue - in internal/interp (needs migration to runtime)
// - FunctionPointerValue - in internal/interp (needs migration to runtime)
// - TypeCastValue - in internal/interp (needs migration to runtime)
// - ClassInfoValue - in internal/interp (needs migration to runtime)
// - Helper infrastructure - findHelperProperty, findHelperMethod (needs adapter methods)
// - Method call infrastructure - evalMethodCall (delegated to VisitMethodCallExpression)
// - Property read infrastructure - evalPropertyRead, evalHelperPropertyRead (needs adapter)
// - Class hierarchy - lookupClassVar, lookupProperty, getMethodOverloadsInHierarchy (needs adapter)
// - Unit registry - ResolveQualifiedVariable (already in Evaluator via unitRegistry)
//
// **TESTING**:
// - Unit-qualified access (Math.PI, System.WriteLine)
// - Static class access (TMyClass.ClassVar, TMyClass.Create, TMyClass.ClassName)
// - Enum type access (TColor.Red, TColor.Low, TColor.High)
// - Record type static access (TPoint.cOrigin, TPoint.Count)
// - Record instance access (point.X, point.GetLength())
// - Object instance access (obj.Name, obj.GetValue(), obj.Count)
// - Interface method access (intf.Hello)
// - Type cast access (TBase(child).ClassVar)
// - Nil object access (nil.ClassVar, nil.Name → error)
// - Enum value properties (TColor.Red.Value)
// - Helper properties/methods (arr.Length, str.ToUpper)
// - Auto-invocation (obj.Method without parentheses for parameterless)
// - Function pointers (obj.Method with parameters returns pointer)
//
// **IMPLEMENTATION SUMMARY**:
// - Original implementation: 706 lines (objects_hierarchy.go:13-719)
// - Handles 11 distinct access modes with complex precedence rules
// - Supports case-insensitive lookups, inheritance, helpers, auto-invocation
// - Requires extensive value type infrastructure not yet in runtime package
// - Full migration deferred - will be broken into category-specific sub-tasks
//
// **MIGRATION STRATEGY**:
// - Phase 1 (this task): Comprehensive documentation of all access modes
// - Phase 2 (future): Migrate simple cases (built-in properties, direct field access)
// - Phase 3 (future): Migrate class/record static access after type system migration
// - Phase 4 (future): Migrate helper infrastructure after helper system migration
// - Phase 5 (future): Migrate method/property dispatch after OOP infrastructure migration
func (e *Evaluator) VisitMemberAccessExpression(node *ast.MemberAccessExpression, ctx *ExecutionContext) Value {
	// All 11 access modes delegated to adapter for now
	// See comprehensive documentation above for detailed behavior
	return e.adapter.EvalNode(node)
}

// VisitMethodCallExpression evaluates a method call (obj.Method(args)).
//
// **COMPLEXITY**: Very High (1,116 lines in original implementation)
// **STATUS**: Documentation-only migration with full adapter delegation
//
// **15 DISTINCT METHOD CALL MODES** (evaluated in this order):
//
// **1. UNIT-QUALIFIED FUNCTION CALLS** (UnitName.FunctionName())
//   - Pattern: `Math.Sin(x)`, `System.WriteLine(s)`
//   - Detection: Object is Identifier referring to registered unit name
//   - Process:
//     a. Check if identifier is registered unit (via unitRegistry.GetUnit)
//     b. Resolve qualified function (ResolveQualifiedFunction)
//     c. Evaluate all arguments
//     d. Call user function (callUserFunction)
//   - Error: "function 'X' not found in unit 'Y'"
//   - Implementation: ~20 lines in original
//
// **2. STATIC CLASS METHOD CALLS** (TClass.Method())
//   - Pattern: `TMyClass.ClassMethod()`, `TMyClass.Create()`
//   - Detection: Object is Identifier referring to registered class name
//   - Lookup order:
//     a. Collect class method overloads (getMethodOverloadsInHierarchy with isClassMethod=true)
//     b. Collect instance method overloads including constructors (isClassMethod=false)
//     c. Special: If constructor with 0 args and no parameterless constructor exists,
//     create object with implicit parameterless constructor (just initialize fields)
//     d. Resolve overload based on argument types (resolveMethodOverload)
//     e. If class method: execute with Self bound to ClassInfoValue (executeClassMethod)
//     f. If instance method/constructor: create new object, initialize fields,
//     execute method with Self bound to new instance
//   - Overload resolution: Uses semantic.ResolveOverload with type matching
//   - Virtual dispatch: NOT used for static calls (static binding)
//   - Field initialization: Field initializers evaluated, then default values for remaining fields
//   - Result: For constructors, always return object (not Result variable)
//   - Error: "wrong number of arguments", "There is no overloaded version that can be called with these arguments"
//   - Implementation: ~245 lines in original
//
// **3. RECORD TYPE STATIC METHOD CALLS** (TRecord.Method())
//   - Pattern: `TPoint.Create()`, `TRecord.ClassMethod()`
//   - Detection: Object is Identifier with `__record_type_` + lowercase(name) in environment
//   - Lookup: Check RecordTypeValue.ClassMethodOverloads (case-insensitive)
//   - Overload resolution: Same as class methods (resolveMethodOverload)
//   - Execution: Call callRecordStaticMethod WITHOUT Self binding
//   - Error: "static method 'X' not found in record type 'Y'"
//   - Implementation: ~30 lines in original
//
// **4. CLASSINFO VALUE METHOD CALLS** (ClassInfoValue.Method())
//   - Pattern: `Self.ClassMethod()` where Self is ClassInfoValue in class method
//   - Detection: Object evaluates to ClassInfoValue
//   - Lookup: Only class methods (getMethodOverloadsInHierarchy with isClassMethod=true)
//   - Execution: executeClassMethod with Self bound to ClassInfoValue
//   - Error: "class method 'X' not found in class 'Y'"
//   - Implementation: ~15 lines in original
//
// **5. METACLASS CONSTRUCTOR CALLS** (ClassValue.Create())
//   - Pattern: `var cls: class of TParent; cls := TChild; obj := cls.Create()`
//   - Detection: Object evaluates to ClassValue
//   - Purpose: Virtual constructor dispatch via metaclass variables
//   - Process:
//     a. Extract runtime class from ClassValue.ClassInfo
//     b. Look up constructor overloads in runtime class (getMethodOverloadsInHierarchy)
//     c. Resolve constructor overload based on arguments
//     d. Create new instance of runtime class (virtual dispatch - uses actual class, not declared type)
//     e. Initialize all fields with default values
//     f. Execute constructor with Self bound to new instance
//   - Key feature: Creates instance of RUNTIME type, not static type
//   - Error: "constructor 'X' not found in class 'Y'"
//   - Implementation: ~95 lines in original
//
// **6. SET VALUE BUILT-IN METHODS** (SetValue.Include/Exclude())
//   - Pattern: `mySet.Include(x)`, `mySet.Exclude(y)`
//   - Detection: Object evaluates to SetValue
//   - Supported methods (case-insensitive):
//     a. Include(element): Add element to set
//     b. Exclude(element): Remove element from set
//   - Error: "method 'X' not found for set type"
//   - Implementation: ~30 lines in original
//
// **7. RECORD INSTANCE METHOD CALLS** (RecordValue.Method())
//   - Pattern: `point.GetLength()`, `record.DoSomething(x)`
//   - Detection: Object evaluates to RecordValue
//   - Process: Convert to member access, delegate to evalRecordMethodCall
//   - Supports: Instance methods, properties, class methods (via instance)
//   - Implementation: ~10 lines in original (delegates to evalRecordMethodCall)
//
// **8. INTERFACE INSTANCE METHOD CALLS** (InterfaceInstance.Method())
//   - Pattern: `intfVar.Hello()`, `intf.DoSomething(x)`
//   - Detection: Object evaluates to InterfaceInstance
//   - Validation: Verify method exists in interface definition (Interface.HasMethod)
//   - Process: Unwrap to underlying object, continue with object method dispatch
//   - Error: "Interface is nil", "method 'X' not found in interface 'Y'"
//   - Implementation: ~15 lines in original
//
// **9. NIL OBJECT ERROR HANDLING**
//   - Pattern: `var o: TClass := nil; o.Method()`
//   - Detection: Object evaluates to NilValue
//   - Result: Always raise "Object not instantiated"
//   - Note: Class methods can only be called on class name, not nil instance
//   - Implementation: ~5 lines in original
//
// **10. ENUM TYPE META METHODS** (TypeMetaValue.Low/High/ByName())
//   - Pattern: `TColor.Low()`, `TColor.High()`, `TColor.ByName('Red')`
//   - Detection: Object evaluates to TypeMetaValue with EnumType
//   - Supported methods (case-insensitive):
//     a. Low(): Returns lowest ordinal value as Integer
//     b. High(): Returns highest ordinal value as Integer
//     c. ByName(name: string): Returns ordinal for enum value name (0 if not found)
//   - Supports qualified names (TypeName.ValueName)
//   - Case-insensitive lookup
//   - Returns 0 for empty string or not found (DWScript behavior)
//   - Implementation: ~50 lines in original
//
// **11. HELPER METHOD CALLS** (any_type.HelperMethod())
//   - Pattern: `"hello".ToUpper()`, `arr.Push(x)`, `123.ToString()`
//   - Detection: Object is not an object/record, but helpers provide this method
//   - Process:
//     a. Find helper method (findHelperMethod) - returns AST method or builtin spec
//     b. Evaluate all arguments
//     c. Call helper method (callHelperMethod)
//   - Supports: String, Array, Integer, Float helpers, etc.
//   - Error: "cannot call method 'X' on type 'Y' (no helper found)"
//   - Implementation: ~20 lines in original
//
// **12. OBJECT INSTANCE METHOD CALLS** (ObjectInstance.Method())
//   - Pattern: `obj.DoSomething()`, `obj.GetValue(x)`
//   - Detection: Object evaluates to ObjectInstance
//   - Built-in methods: `ClassName()` returns obj.Class.Name
//   - Process:
//     a. Collect instance method overloads (getMethodOverloadsInHierarchy with isClassMethod=false)
//     b. Collect class method overloads (isClassMethod=true) - can be called on instances
//     c. Resolve overload based on argument types (resolveMethodOverload)
//     d. Apply VIRTUAL DISPATCH for virtual/override methods (use VirtualMethodTable)
//   - Only for methods with IsVirtual or IsOverride (NOT reintroduce)
//   - Look up method signature in obj.Class.VirtualMethodTable
//   - Use most derived implementation if found
//     e. Check recursion depth (WillOverflow) before execution
//     f. Push method name to call stack for stack traces
//     g. Bind Self to object (or ClassInfoValue for class methods)
//     h. Add class constants to method scope
//     i. Bind parameters to arguments with implicit type conversion
//     j. Initialize Result variable (or method name alias)
//     k. Execute method body
//     l. Extract return value (Result or method name variable)
//     m. Apply implicit return type conversion
//   - Virtual constructor handling: If calling constructor on instance (o.Create),
//     create NEW instance of runtime type with virtual dispatch
//   - Error: "method 'X' not found in class 'Y'"
//   - Implementation: ~290 lines in original
//
// **13. VIRTUAL CONSTRUCTOR DISPATCH** (obj.Create())
//   - Pattern: `var o: TParent; o := TChild.Create; newObj := o.Create()`
//   - Detection: Resolved method is constructor (method.IsConstructor)
//   - Purpose: Create new instance of object's RUNTIME type (virtual dispatch)
//   - Process:
//     a. Find constructor in object's runtime class hierarchy (start from obj.Class)
//     b. Create NEW instance of runtime class (not existing object)
//     c. Initialize all fields with default values
//     d. Execute constructor with Self bound to new instance
//   - Key feature: Always creates NEW object, doesn't modify existing object
//   - Returns: New ObjectInstance (not the one the method was called on)
//   - Implementation: ~85 lines in original
//
// **14. CLASS METHOD EXECUTION** (executeClassMethod)
//   - Pattern: All class methods (called on class or instance)
//   - Self binding: Bound to ClassInfoValue (not instance)
//   - Environment: New environment with class constants, parameters, Result
//   - Call stack: Tracks method name for recursion detection and stack traces
//   - Recursion checking: Validates against max recursion limit
//   - Result handling: Checks Result and method name variables, implicit conversion
//   - Implementation: ~105 lines in original
//
// **15. OVERLOAD RESOLUTION** (resolveMethodOverload)
//   - Purpose: Select correct method from multiple overloads based on argument types
//   - Process:
//     a. Fast path: Single overload → return immediately
//     b. Evaluate all arguments to get runtime types (getValueType)
//     c. Convert method declarations to semantic.Symbol
//     d. Use semantic.ResolveOverload with type matching
//     e. Find method declaration corresponding to selected symbol
//   - Signature matching: Compares parameter types AND return type
//   - Inheritance: Child methods with same signature hide parent methods
//   - Error: "There is no overloaded version of 'X.Y' that can be called with these arguments"
//   - Implementation: ~50 lines in original + ~70 lines for getMethodOverloadsInHierarchy
//
// **SPECIAL BEHAVIORS**:
// - **Virtual dispatch**: Methods marked virtual/override use VirtualMethodTable for polymorphism
// - **Overload resolution**: Multiple methods with same name resolved by argument types
// - **Case-insensitive**: All name lookups are case-insensitive (DWScript spec)
// - **Inheritance**: Methods searched up class hierarchy, child signatures hide parent
// - **Recursion tracking**: Call stack monitored, max depth enforced
// - **Self binding**: Varies by context (object, ClassInfoValue, or nil)
// - **Result variable**: Functions initialize Result, can also use method name as alias
// - **Implicit conversion**: Parameters and return values converted to match declared types
// - **Helper support**: Types without native methods can have helper methods
// - **Constructor semantics**: Always return new object (not Result), initialize all fields
// - **Virtual constructors**: Constructor calls on instances use runtime type (virtual dispatch)
// - **Field initialization**: Field initializers evaluated in temporary environment with class constants
// - **Class constants**: Added to method scope for direct access without qualification
//
// **DEPENDENCIES** (blockers for full migration):
// - ObjectInstance, ClassInfo, ClassValue, ClassInfoValue - in internal/interp (needs migration)
// - RecordValue, RecordTypeValue - in internal/interp (needs migration)
// - InterfaceInstance, Interface - in internal/interp (needs migration)
// - SetValue - in internal/interp (needs migration)
// - TypeMetaValue, EnumType - in internal/interp (needs migration)
// - ReferenceValue - in internal/interp (needs migration)
// - VirtualMethodTable infrastructure - in internal/interp (needs migration)
// - Helper infrastructure - findHelperMethod, callHelperMethod (needs adapter)
// - Overload resolution - resolveMethodOverload, getMethodOverloadsInHierarchy (needs adapter)
// - User function calls - callUserFunction, callRecordStaticMethod (needs adapter)
// - Type system - getValueType, extractFunctionType, resolveTypeFromAnnotation (needs adapter)
// - Unit registry - ResolveQualifiedFunction (already in Evaluator)
// - Call stack - pushCallStack, popCallStack, WillOverflow (needs adapter)
// - Environment management - NewEnclosedEnvironment, bindClassConstantsToEnv (needs adapter)
//
// **TESTING**:
// - Unit-qualified function calls (Math.Sin, System.WriteLine)
// - Static class method calls (TClass.ClassMethod, TClass.Create)
// - Implicit parameterless constructor (TClass.Create with no constructor defined)
// - Constructor overloading (TClass.Create(), TClass.Create(x))
// - Record static method calls (TRecord.Create, TRecord.Count)
// - ClassInfoValue method calls (Self.ClassMethod in class context)
// - Metaclass constructor calls (cls.Create where cls is class of T)
// - Set methods (mySet.Include(x), mySet.Exclude(y))
// - Record instance methods (point.GetLength, record.DoSomething)
// - Interface method calls (intf.Hello, intf.Process)
// - Nil object error (nil.Method → error)
// - Enum meta methods (TColor.Low(), TColor.ByName('Red'))
// - Helper methods (str.ToUpper, arr.Push, num.ToString)
// - Object instance methods (obj.GetValue, obj.DoSomething)
// - Virtual dispatch (parent ref to child object calls child's override)
// - Virtual constructor dispatch (obj.Create creates new instance of runtime type)
// - Method overloading (multiple methods with same name, different parameters)
// - Overload inheritance (child methods hide parent methods with same signature)
// - Recursion limit enforcement (stack overflow detection)
// - Self binding variations (object, ClassInfoValue, record)
// - Result variable handling (Result, method name alias, implicit conversion)
// - Class constant access in methods (direct access without qualification)
// - Field initialization (initializers, default values, constructor execution)
//
// **IMPLEMENTATION SUMMARY**:
// - Original implementation: 1,116 lines (objects_methods.go:12-1116)
// - Handles 15 distinct method call modes with complex dispatch logic
// - Supports virtual dispatch, overload resolution, recursion tracking
// - Requires extensive OOP infrastructure not yet in runtime package
// - Full migration deferred - will be broken into category-specific sub-tasks
//
// **MIGRATION STRATEGY**:
// - Phase 1 (this task): Comprehensive documentation of all call modes
// - Phase 2 (future): Migrate simple cases (built-in methods, direct calls)
// - Phase 3 (future): Migrate overload resolution after type system migration
// - Phase 4 (future): Migrate virtual dispatch after VMT migration
// - Phase 5 (future): Migrate constructor dispatch after object creation migration
// - Phase 6 (future): Migrate helper methods after helper system migration
func (e *Evaluator) VisitMethodCallExpression(node *ast.MethodCallExpression, ctx *ExecutionContext) Value {
	// All 15 method call modes delegated to adapter for now
	// See comprehensive documentation above for detailed behavior
	return e.adapter.EvalNode(node)
}

// VisitInheritedExpression evaluates an 'inherited' expression.
// Calls parent class method with proper context and Self preservation.
func (e *Evaluator) VisitInheritedExpression(node *ast.InheritedExpression, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitSelfExpression evaluates a 'Self' expression.
// Self refers to the current instance (in instance methods) or the current class (in class methods).
func (e *Evaluator) VisitSelfExpression(node *ast.SelfExpression, ctx *ExecutionContext) Value {
	selfVal, exists := ctx.Env().Get("Self")
	if !exists {
		return e.newError(node, "Self used outside method context")
	}

	val, ok := selfVal.(Value)
	if !ok {
		return e.newError(node, "Self has invalid type")
	}

	return val
}

// VisitEnumLiteral evaluates an enum literal (EnumType.Value).
func (e *Evaluator) VisitEnumLiteral(node *ast.EnumLiteral, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil enum literal")
	}

	valueName := node.ValueName
	val, ok := ctx.Env().Get(valueName)
	if !ok {
		return e.newError(node, "undefined enum value '%s'", valueName)
	}

	if value, ok := val.(Value); ok {
		return value
	}

	return e.newError(node, "enum value '%s' has invalid type", valueName)
}

// VisitRecordLiteralExpression evaluates a record literal expression.
// Handles typed and anonymous record literals with field initialization and default values.
func (e *Evaluator) VisitRecordLiteralExpression(node *ast.RecordLiteralExpression, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitSetLiteral evaluates a set literal [value1, value2, ...].
// Handles simple elements, ranges, and mixed sets with proper type inference.
func (e *Evaluator) VisitSetLiteral(node *ast.SetLiteral, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitArrayLiteralExpression evaluates an array literal [1, 2, 3].
// Handles type inference, element coercion, and bounds validation for static and dynamic arrays.
func (e *Evaluator) VisitArrayLiteralExpression(node *ast.ArrayLiteralExpression, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil array literal")
	}

	// Empty arrays need type annotation
	if len(node.Elements) == 0 {
		return e.adapter.EvalNode(node)
	}

	// Evaluate all element expressions
	elementCount := len(node.Elements)
	evaluatedElements := make([]Value, elementCount)

	for idx, elem := range node.Elements {
		val := e.Eval(elem, ctx)
		if isError(val) {
			return val
		}
		evaluatedElements[idx] = val
	}

	// Determine array element type
	arrayType := e.getArrayElementType(node, evaluatedElements)
	if arrayType == nil {
		return e.newError(node, "cannot infer type of array literal")
	}

	// Validate element type compatibility
	if err := e.coerceArrayElements(evaluatedElements, arrayType.ElementType, node); err != nil {
		return err
	}

	// Validate static array bounds
	if err := e.validateArrayLiteralSize(arrayType, elementCount, node); err != nil {
		return err
	}

	// Delegate final construction to adapter
	return e.adapter.EvalNode(node)
}

// VisitIndexExpression evaluates an index expression array[index].
// Handles array, string, property, and JSON indexing with bounds checking.
func (e *Evaluator) VisitIndexExpression(node *ast.IndexExpression, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil index expression")
	}

	if node.Left == nil {
		return e.newError(node, "index expression missing base")
	}

	base := e.Eval(node.Left, ctx)
	if isError(base) {
		return base
	}

	if node.Index == nil {
		return e.newError(node, "index expression missing index")
	}

	index := e.Eval(node.Index, ctx)
	if isError(index) {
		return index
	}

	// Delegate indexing logic to adapter
	return e.adapter.EvalNode(node)
}

// VisitNewArrayExpression evaluates a new array expression.
// Handles dimension evaluation and multi-dimensional array construction with default values.
func (e *Evaluator) VisitNewArrayExpression(node *ast.NewArrayExpression, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil new array expression")
	}

	if node.ElementTypeName == nil {
		return e.newError(node, "new array expression missing element type")
	}

	// Evaluate and validate dimensions
	_, err := e.evaluateDimensions(node.Dimensions, ctx, node)
	if err != nil {
		return err
	}

	// Delegate array construction to adapter
	return e.adapter.EvalNode(node)
}

// VisitLambdaExpression evaluates a lambda expression (closure).
// Creates a lambda that captures the current scope.
func (e *Evaluator) VisitLambdaExpression(node *ast.LambdaExpression, ctx *ExecutionContext) Value {
	return e.adapter.CreateLambda(node, ctx.Env())
}

// VisitIsExpression evaluates an 'is' type checking expression.
// Performs runtime type checking with class hierarchy and interface support.
func (e *Evaluator) VisitIsExpression(node *ast.IsExpression, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitAsExpression evaluates an 'as' type casting expression.
// Performs runtime type checking and casts, raising exception on failure.
func (e *Evaluator) VisitAsExpression(node *ast.AsExpression, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitImplementsExpression evaluates an 'implements' interface checking expression.
// Checks if an object's class implements a specified interface.
func (e *Evaluator) VisitImplementsExpression(node *ast.ImplementsExpression, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitIfExpression evaluates an inline if-then-else expression.
func (e *Evaluator) VisitIfExpression(node *ast.IfExpression, ctx *ExecutionContext) Value {
	condition := e.Eval(node.Condition, ctx)
	if isError(condition) {
		return condition
	}

	// Evaluate consequence if condition is truthy
	if IsTruthy(condition) {
		result := e.Eval(node.Consequence, ctx)
		if isError(result) {
			return result
		}
		return result
	}

	// Evaluate alternative if present
	if node.Alternative != nil {
		result := e.Eval(node.Alternative, ctx)
		if isError(result) {
			return result
		}
		return result
	}

	// No else clause - return default value for the consequence type
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
		return &runtime.NilValue{}
	}
}

// VisitOldExpression evaluates an 'old' expression (used in postconditions).
func (e *Evaluator) VisitOldExpression(node *ast.OldExpression, ctx *ExecutionContext) Value {
	identName := node.Identifier.Value
	oldValue, found := ctx.GetOldValue(identName)
	if !found {
		return e.newError(node, "old value for '%s' not captured (internal error)", identName)
	}
	return oldValue.(Value)
}
