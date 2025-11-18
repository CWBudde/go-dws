package interp

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/internal/units"

	// Task 3.8.2: pkg/ast is imported for SemanticInfo, which holds semantic analysis
	// metadata (type annotations, symbol resolutions). This is separate from the AST
	// structure itself and is not aliased in internal/ast.
	// Task 9.18: Separate type metadata from AST nodes.
	pkgast "github.com/cwbudde/go-dws/pkg/ast"
)

// DefaultMaxRecursionDepth is the default maximum recursion depth for function calls.
// This matches DWScript's default limit (see dwsCompiler.pas:39 cDefaultMaxRecursionDepth).
// When the call stack reaches this depth, the interpreter raises an EScriptStackOverflow exception
// to prevent infinite recursion and potential Go runtime stack overflow.
const DefaultMaxRecursionDepth = 1024

// PropertyEvalContext tracks the state during property getter/setter evaluation.
// Deprecated: Use evaluator.PropertyEvalContext instead. This is kept for backward compatibility.
type PropertyEvalContext = evaluator.PropertyEvalContext

// Interpreter executes DWScript AST nodes and manages the runtime environment.
//
// Phase 3.5.1: The Interpreter is being refactored to be a thin orchestrator.
// The evaluator field contains the evaluation logic and dependencies.
// Eventually, most of the fields below will be removed and accessed via the evaluator.
type Interpreter struct {
	// Phase 3.5.1: Evaluator - the new evaluation engine
	// This holds all the evaluation logic and dependencies (type system, runtime services, config)
	evaluatorInstance *evaluator.Evaluator

	// Phase 3.3.1: Execution context (gradually replacing individual state fields)
	ctx *evaluator.ExecutionContext

	// Execution state (Phase 3.3: will be moved to ExecutionContext)
	currentNode      ast.Node
	env              *Environment       // Phase 3.3: migrating to ctx.Env()
	exception        *ExceptionValue    // Phase 3.3: migrating to ctx.Exception()
	handlerException *ExceptionValue    // Phase 3.3: migrating to ctx.HandlerException()
	callStack        errors.StackTrace  // Phase 3.3: migrating to ctx.CallStack()
	oldValuesStack   []map[string]Value // Phase 3.3: migrating to ctx.PushOldValues/PopOldValues
	propContext      *PropertyEvalContext
	// Phase 3.3.2: Control flow now managed by ctx.ControlFlow() instead of boolean flags

	// Type System (Phase 3.4.1: Extracted to TypeSystem)
	typeSystem *interptypes.TypeSystem

	// Backward compatibility fields (Phase 3.4.1: point to typeSystem internals)
	// These will be gradually removed as code is migrated to use typeSystem directly
	classes              map[string]*ClassInfo
	records              map[string]*RecordTypeValue
	interfaces           map[string]*InterfaceInfo
	functions            map[string][]*ast.FunctionDecl
	globalOperators      *runtimeOperatorRegistry
	conversions          *runtimeConversionRegistry
	helpers              map[string][]*HelperInfo
	classTypeIDRegistry  map[string]int // Type ID registry for classes
	recordTypeIDRegistry map[string]int // Type ID registry for records
	enumTypeIDRegistry   map[string]int // Type ID registry for enums
	nextClassTypeID      int            // Next available class type ID
	nextRecordTypeID     int            // Next available record type ID
	nextEnumTypeID       int            // Next available enum type ID

	// Runtime Services (Phase 3.4: will be moved to RuntimeServices)
	output            io.Writer
	rand              *rand.Rand
	randSeed          int64
	externalFunctions *ExternalFunctionRegistry

	// Configuration
	maxRecursionDepth int
	sourceCode        string
	sourceFile        string

	// Unit System
	initializedUnits map[string]bool
	unitRegistry     *units.UnitRegistry
	loadedUnits      []string

	// Semantic Analysis
	semanticInfo *pkgast.SemanticInfo // Task 9.18: Type metadata from semantic analysis
}

// New creates a new Interpreter with a fresh global environment.
// The output writer is where built-in functions like PrintLn will write.
func New(output io.Writer) *Interpreter {
	return NewWithOptions(output, nil)
}

// NewWithOptions creates a new Interpreter with options.
// If options is nil, default options are used.
// Task 3.8.1: Uses Options interface to avoid circular dependency and remove reflection hack.
func NewWithOptions(output io.Writer, opts Options) *Interpreter {
	env := NewEnvironment()
	// Initialize random number generator with a default seed
	// Randomize() can be called to re-seed with current time
	const defaultSeed = int64(1)
	source := rand.NewSource(defaultSeed)

	// Phase 3.4.1: Initialize TypeSystem
	// The TypeSystem is the new centralized type registry that manages all type information
	// including classes, records, interfaces, functions, helpers, operators, and conversions.
	//
	// Migration Strategy (Gradual Transition):
	// - The old fields (functions, classes, records, etc.) are kept for backward compatibility
	// - Existing code continues to work unchanged during the transition period
	// - New code should use typeSystem methods (e.g., typeSystem.RegisterClass, typeSystem.LookupClass)
	// - Old code will be gradually refactored to use typeSystem in future tasks
	// - Once migration is complete, the old fields will be removed (future Phase 4+ work)
	ts := interptypes.NewTypeSystem()

	interp := &Interpreter{
		env:               env,
		output:            output,
		rand:              rand.New(source),
		randSeed:          defaultSeed,
		loadedUnits:       make([]string, 0),
		initializedUnits:  make(map[string]bool),
		maxRecursionDepth: DefaultMaxRecursionDepth,
		callStack:         errors.NewStackTrace(), // Initialize stack trace

		// Phase 3.4.1: TypeSystem (new centralized type registry)
		// This is the modern API - use this for new code
		typeSystem: ts,

		// Phase 3.4.1: Legacy fields for backward compatibility
		// These will be removed once migration to typeSystem is complete
		functions:            make(map[string][]*ast.FunctionDecl), // Task 9.66: Support overloading
		classes:              make(map[string]*ClassInfo),
		records:              make(map[string]*RecordTypeValue),
		interfaces:           make(map[string]*InterfaceInfo),
		globalOperators:      newRuntimeOperatorRegistry(),
		conversions:          newRuntimeConversionRegistry(),
		helpers:              make(map[string][]*HelperInfo),
		classTypeIDRegistry:  make(map[string]int), // Task 9.25: RTTI type ID registry
		recordTypeIDRegistry: make(map[string]int), // Task 9.25: RTTI type ID registry
		enumTypeIDRegistry:   make(map[string]int), // Task 9.25: RTTI type ID registry
		nextClassTypeID:      1000,                 // Task 9.25: Start class IDs at 1000
		nextRecordTypeID:     200000,               // Task 9.25: Start record IDs at 200000
		nextEnumTypeID:       300000,               // Task 9.25: Start enum IDs at 300000
	}

	// Task 3.8.1: Extract external functions and recursion depth from options using interface
	// This replaces the reflection hack with a clean interface-based approach
	if opts != nil {
		// Extract ExternalFunctions
		if registry := opts.GetExternalFunctions(); registry != nil {
			interp.externalFunctions = registry
		}

		// Extract MaxRecursionDepth
		if depth := opts.GetMaxRecursionDepth(); depth > 0 {
			interp.maxRecursionDepth = depth
		}
	}

	// Ensure we have a registry even if not provided
	if interp.externalFunctions == nil {
		interp.externalFunctions = NewExternalFunctionRegistry()
	}

	// Phase 3.3.1/3.3.3: Initialize execution context with call stack overflow detection
	// The context wraps the environment with a simple adapter to avoid circular dependencies.
	// The Environment is passed as interface{} to ExecutionContext to avoid import cycles.
	// Phase 3.3.3: Pass maxRecursionDepth to configure CallStack overflow detection.
	interp.ctx = evaluator.NewExecutionContextWithMaxDepth(
		evaluator.NewEnvironmentAdapter(env),
		interp.maxRecursionDepth,
	)

	// Phase 3.5.1: Initialize Evaluator
	// The Evaluator holds evaluation logic and dependencies (type system, runtime services, config)
	// Note: unitRegistry can be nil initially - it's set via SetUnitRegistry if needed

	// Create evaluator config
	evalConfig := &evaluator.Config{
		MaxRecursionDepth: interp.maxRecursionDepth,
		SourceCode:        "",
		SourceFile:        "",
	}

	// Create evaluator instance
	interp.evaluatorInstance = evaluator.NewEvaluator(
		ts,
		output,
		evalConfig,
		nil, // unitRegistry is set later via SetUnitRegistry if needed
	)

	// Set external functions if available
	if interp.externalFunctions != nil {
		interp.evaluatorInstance.SetExternalFunctions(interp.externalFunctions)
	}

	// Set the adapter so the evaluator can delegate back to the interpreter during migration
	interp.evaluatorInstance.SetAdapter(interp)

	// Register built-in exception classes
	interp.registerBuiltinExceptions()

	// Register built-in interfaces
	interp.registerBuiltinInterfaces()

	// Register built-in array helpers
	interp.initArrayHelpers()

	// Register built-in helpers for primitive types
	interp.initIntrinsicHelpers()

	// Register built-in enum helpers
	interp.initEnumHelpers()

	// Initialize ExceptObject to nil
	// ExceptObject is a built-in global variable that holds the current exception
	env.Define("ExceptObject", &NilValue{})

	// Register built-in type meta-values
	// These allow type names to be used as runtime values, e.g., High(Integer)
	env.Define("Integer", NewTypeMetaValue(types.INTEGER, "Integer"))
	env.Define("Float", NewTypeMetaValue(types.FLOAT, "Float"))
	env.Define("String", NewTypeMetaValue(types.STRING, "String"))
	env.Define("Boolean", NewTypeMetaValue(types.BOOLEAN, "Boolean"))

	// Register mathematical constants
	env.Define("PI", &FloatValue{Value: math.Pi})
	env.Define("NaN", &FloatValue{Value: math.NaN()})
	env.Define("Infinity", &FloatValue{Value: math.Inf(1)})

	// Task 9.4.1: Register Variant special values
	env.Define("Null", NewNullValue())
	env.Define("Unassigned", NewUnassignedValue())

	return interp
}

// GetException returns the current active exception, or nil if none.
// This is used by the CLI to detect and report unhandled exceptions.
func (i *Interpreter) GetException() *ExceptionValue {
	return i.exception
}

// SetSemanticInfo sets the semantic metadata table for this interpreter.
// The semantic info contains type annotations and symbol resolutions from analysis.
// Task 9.18: Separate type metadata from AST nodes.
func (i *Interpreter) SetSemanticInfo(info *pkgast.SemanticInfo) {
	i.semanticInfo = info

	// Phase 3.5.1: Also update the evaluator's semantic info
	if i.evaluatorInstance != nil {
		i.evaluatorInstance.SetSemanticInfo(info)
	}
}

// GetEvaluator returns the evaluator instance.
// Phase 3.5.1: This provides access to the evaluation engine for advanced use cases.
func (i *Interpreter) GetEvaluator() *evaluator.Evaluator {
	return i.evaluatorInstance
}

// EvalNode implements the evaluator.InterpreterAdapter interface.
// This allows the Evaluator to delegate back to the Interpreter during migration.
// Phase 3.5.1: This is temporary and will be removed once all evaluation logic
// is moved to the Evaluator.
func (i *Interpreter) EvalNode(node ast.Node) evaluator.Value {
	// Delegate to the legacy Eval method
	// The cast is safe because our Value type matches evaluator.Value interface
	return i.Eval(node)
}

// Phase 3.5.4 - Phase 2A: Function call system adapter methods
// These methods implement the InterpreterAdapter interface for function calls.

// convertEvaluatorArgs converts a slice of evaluator.Value to interp.Value.
// This is used by adapter methods when delegating to internal functions.
func convertEvaluatorArgs(args []evaluator.Value) []Value {
	interpArgs := make([]Value, len(args))
	for idx, arg := range args {
		interpArgs[idx] = arg
	}
	return interpArgs
}

// CallFunctionPointer executes a function pointer with given arguments.
func (i *Interpreter) CallFunctionPointer(funcPtr evaluator.Value, args []evaluator.Value, node ast.Node) evaluator.Value {
	// Convert evaluator.Value to interp.Value (they're the same interface)
	fp, ok := funcPtr.(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(node, "invalid function pointer type: expected FunctionPointerValue, got %T", funcPtr)
	}

	return i.callFunctionPointer(fp, convertEvaluatorArgs(args), node)
}

// CallUserFunction executes a user-defined function.
func (i *Interpreter) CallUserFunction(fn *ast.FunctionDecl, args []evaluator.Value) evaluator.Value {
	return i.callUserFunction(fn, convertEvaluatorArgs(args))
}

// CallBuiltinFunction executes a built-in function by name.
func (i *Interpreter) CallBuiltinFunction(name string, args []evaluator.Value) evaluator.Value {
	return i.callBuiltinFunction(name, convertEvaluatorArgs(args))
}

// LookupFunction finds a function by name in the function registry.
func (i *Interpreter) LookupFunction(name string) ([]*ast.FunctionDecl, bool) {
	// DWScript is case-insensitive, so normalize to lowercase
	normalizedName := strings.ToLower(name)
	functions, ok := i.functions[normalizedName]
	return functions, ok
}

// Phase 3.5.4 - Phase 2B: Type system access adapter methods
// These methods implement the InterpreterAdapter interface for type system access.

// ===== Class Registry =====

// LookupClass finds a class by name in the class registry.
func (i *Interpreter) LookupClass(name string) (any, bool) {
	normalizedName := strings.ToLower(name)
	class, ok := i.classes[normalizedName]
	if !ok {
		return nil, false
	}
	return class, true
}

// HasClass checks if a class with the given name exists.
func (i *Interpreter) HasClass(name string) bool {
	normalizedName := strings.ToLower(name)
	_, ok := i.classes[normalizedName]
	return ok
}

// GetClassTypeID returns the type ID for a class, or 0 if not found.
func (i *Interpreter) GetClassTypeID(className string) int {
	normalizedName := strings.ToLower(className)
	typeID, ok := i.classTypeIDRegistry[normalizedName]
	if !ok {
		return 0
	}
	return typeID
}

// ===== Record Registry =====

// LookupRecord finds a record type by name in the record registry.
func (i *Interpreter) LookupRecord(name string) (any, bool) {
	normalizedName := strings.ToLower(name)
	record, ok := i.records[normalizedName]
	if !ok {
		return nil, false
	}
	return record, true
}

// HasRecord checks if a record type with the given name exists.
func (i *Interpreter) HasRecord(name string) bool {
	normalizedName := strings.ToLower(name)
	_, ok := i.records[normalizedName]
	return ok
}

// GetRecordTypeID returns the type ID for a record type, or 0 if not found.
func (i *Interpreter) GetRecordTypeID(recordName string) int {
	normalizedName := strings.ToLower(recordName)
	typeID, ok := i.recordTypeIDRegistry[normalizedName]
	if !ok {
		return 0
	}
	return typeID
}

// ===== Interface Registry =====

// LookupInterface finds an interface by name in the interface registry.
func (i *Interpreter) LookupInterface(name string) (any, bool) {
	normalizedName := strings.ToLower(name)
	iface, ok := i.interfaces[normalizedName]
	if !ok {
		return nil, false
	}
	return iface, true
}

// HasInterface checks if an interface with the given name exists.
func (i *Interpreter) HasInterface(name string) bool {
	normalizedName := strings.ToLower(name)
	_, ok := i.interfaces[normalizedName]
	return ok
}

// ===== Helper Registry =====

// LookupHelpers finds helper methods for a type by name.
func (i *Interpreter) LookupHelpers(typeName string) []any {
	normalizedName := strings.ToLower(typeName)
	helpers, ok := i.helpers[normalizedName]
	if !ok {
		return nil
	}
	// Convert []*HelperInfo to []any
	result := make([]any, len(helpers))
	for idx, helper := range helpers {
		result[idx] = helper
	}
	return result
}

// HasHelpers checks if a type has helper methods defined.
func (i *Interpreter) HasHelpers(typeName string) bool {
	normalizedName := strings.ToLower(typeName)
	helpers, ok := i.helpers[normalizedName]
	return ok && len(helpers) > 0
}

// ===== Operator & Conversion Registries =====

// GetOperatorRegistry returns the operator registry for operator overload lookups.
func (i *Interpreter) GetOperatorRegistry() any {
	return i.globalOperators
}

// GetConversionRegistry returns the conversion registry for type conversion lookups.
func (i *Interpreter) GetConversionRegistry() any {
	return i.conversions
}

// ===== Enum Type IDs =====

// GetEnumTypeID returns the type ID for an enum type, or 0 if not found.
func (i *Interpreter) GetEnumTypeID(enumName string) int {
	normalizedName := strings.ToLower(enumName)
	typeID, ok := i.enumTypeIDRegistry[normalizedName]
	if !ok {
		return 0
	}
	return typeID
}

// ===== Task 3.5.5: Type System Adapter Method Implementations =====

// GetType resolves a type by name using the type system.
func (i *Interpreter) GetType(name string) (any, error) {
	typ, err := i.resolveType(name)
	if err != nil {
		return nil, err
	}
	return typ, nil
}

// ResolveType resolves a type from an AST type annotation.
func (i *Interpreter) ResolveType(typeAnnotation *ast.TypeAnnotation) (any, error) {
	if typeAnnotation == nil {
		return nil, fmt.Errorf("nil type annotation")
	}
	return i.resolveType(typeAnnotation.String())
}

// IsTypeCompatible checks if a value is compatible with a target type.
func (i *Interpreter) IsTypeCompatible(from evaluator.Value, toTypeName string) bool {
	// Convert from evaluator.Value to internal Value
	internalValue := from.(Value)

	// Try implicit conversion - if it succeeds, types are compatible
	_, ok := i.tryImplicitConversion(internalValue, toTypeName)
	return ok
}

// InferArrayElementType infers the element type from array literal elements.
func (i *Interpreter) InferArrayElementType(elements []evaluator.Value) (any, error) {
	if len(elements) == 0 {
		// Empty array - cannot infer type
		return nil, fmt.Errorf("cannot infer type from empty array")
	}

	// Convert first element to internal Value
	firstInternalValue := elements[0].(Value)

	// Use the type of the first element
	firstType := i.typeFromValue(firstInternalValue)

	// Verify all elements have compatible types
	for idx, elem := range elements[1:] {
		internalElem := elem.(Value)
		elemType := i.typeFromValue(internalElem)
		if elemType.String() != firstType.String() {
			return nil, fmt.Errorf("incompatible types in array: element 0 is %s, element %d is %s",
				firstType.String(), idx+1, elemType.String())
		}
	}

	return firstType, nil
}

// InferRecordType infers the record type name from field values.
func (i *Interpreter) InferRecordType(fields map[string]evaluator.Value) (string, error) {
	// This is complex - for now, we cannot infer record types from values alone
	// Record type inference typically requires explicit type annotations
	return "", fmt.Errorf("cannot infer record type from fields (explicit type required)")
}

// ConvertValue performs implicit or explicit type conversion.
func (i *Interpreter) ConvertValue(value evaluator.Value, targetTypeName string) (evaluator.Value, error) {
	// Convert from evaluator.Value to internal Value
	internalValue := value.(Value)

	// Try implicit conversion first
	if converted, ok := i.tryImplicitConversion(internalValue, targetTypeName); ok {
		return converted, nil
	}

	// Conversion failed
	return nil, fmt.Errorf("cannot convert %s to %s", value.Type(), targetTypeName)
}

// CreateDefaultValue creates a zero/default value for a given type name.
func (i *Interpreter) CreateDefaultValue(typeName string) evaluator.Value {
	normalizedName := strings.ToLower(typeName)

	// Check for basic types
	switch normalizedName {
	case "integer", "int64":
		return &IntegerValue{Value: 0}
	case "float", "float64", "double", "real":
		return &FloatValue{Value: 0.0}
	case "string":
		return &StringValue{Value: ""}
	case "boolean", "bool":
		return &BooleanValue{Value: false}
	case "variant":
		return &VariantValue{} // Unassigned variant
	}

	// Check for enum types
	if i.IsEnumType(typeName) {
		enumTypeKey := "__enum_type_" + normalizedName
		if typeVal, ok := i.env.Get(enumTypeKey); ok {
			if etv, ok := typeVal.(*EnumTypeValue); ok {
				// Return first enum value
				if len(etv.EnumType.OrderedNames) > 0 {
					firstValueName := etv.EnumType.OrderedNames[0]
					firstOrdinal := etv.EnumType.Values[firstValueName]
					return &EnumValue{
						TypeName:     etv.EnumType.Name,
						ValueName:    firstValueName,
						OrdinalValue: firstOrdinal,
					}
				}
			}
		}
	}

	// Check for record types
	if i.IsRecordType(typeName) {
		recordTypeKey := "__record_type_" + normalizedName
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				return i.createRecordValue(rtv.RecordType, rtv.Methods)
			}
		}
	}

	// Check for array types
	if i.IsArrayType(typeName) {
		arrayTypeKey := "__array_type_" + normalizedName
		if typeVal, ok := i.env.Get(arrayTypeKey); ok {
			if atv, ok := typeVal.(*ArrayTypeValue); ok {
				return NewArrayValue(atv.ArrayType)
			}
		}
	}

	// Check for set types
	if strings.HasPrefix(normalizedName, "set of ") {
		setType := i.parseInlineSetType(typeName)
		if setType != nil {
			return NewSetValue(setType)
		}
	}

	// For unknown types, return nil
	return &NilValue{}
}

// IsEnumType checks if a given name refers to an enum type.
func (i *Interpreter) IsEnumType(typeName string) bool {
	normalizedName := strings.ToLower(typeName)
	enumTypeKey := "__enum_type_" + normalizedName
	_, ok := i.env.Get(enumTypeKey)
	return ok
}

// IsRecordType checks if a given name refers to a record type.
func (i *Interpreter) IsRecordType(typeName string) bool {
	normalizedName := strings.ToLower(typeName)
	recordTypeKey := "__record_type_" + normalizedName
	_, ok := i.env.Get(recordTypeKey)
	return ok
}

// IsArrayType checks if a given name refers to an array type.
func (i *Interpreter) IsArrayType(typeName string) bool {
	normalizedName := strings.ToLower(typeName)
	arrayTypeKey := "__array_type_" + normalizedName
	_, ok := i.env.Get(arrayTypeKey)
	return ok
}

// ===== Task 3.5.6: Array and Collection Adapter Method Implementations =====

// CreateArray creates an array from a list of elements with a specified element type.
func (i *Interpreter) CreateArray(elementType any, elements []evaluator.Value) evaluator.Value {
	// Convert elementType to types.Type
	var typedElementType types.Type
	if elementType != nil {
		if t, ok := elementType.(types.Type); ok {
			typedElementType = t
		}
	}

	// Convert evaluator.Value slice to internal Value slice
	internalElements := make([]Value, len(elements))
	for idx, elem := range elements {
		internalElements[idx] = elem.(Value)
	}

	// Create array type (dynamic array has nil bounds)
	arrayType := &types.ArrayType{
		ElementType: typedElementType,
		LowBound:    nil,
		HighBound:   nil,
	}

	// Create array value
	arrayVal := NewArrayValue(arrayType)

	// Add elements (append to Elements slice)
	arrayVal.Elements = append(arrayVal.Elements, internalElements...)

	return arrayVal
}

// CreateDynamicArray allocates a new dynamic array of a given size and element type.
func (i *Interpreter) CreateDynamicArray(elementType any, size int) evaluator.Value {
	// Convert elementType to types.Type
	var typedElementType types.Type
	if elementType != nil {
		if t, ok := elementType.(types.Type); ok {
			typedElementType = t
		}
	}

	// Create array type (dynamic array has nil bounds)
	arrayType := &types.ArrayType{
		ElementType: typedElementType,
		LowBound:    nil,
		HighBound:   nil,
	}

	// Create array value
	arrayVal := NewArrayValue(arrayType)

	// Pre-fill with default values if size > 0
	if size > 0 {
		defaultVal := i.CreateDefaultValue(typedElementType.String()).(Value)
		for j := 0; j < size; j++ {
			arrayVal.Elements = append(arrayVal.Elements, defaultVal)
		}
	}

	return arrayVal
}

// CreateArrayWithExpectedType creates an array from elements with type-aware construction.
func (i *Interpreter) CreateArrayWithExpectedType(elements []evaluator.Value, expectedType any) evaluator.Value {
	// Convert expectedType to *types.ArrayType
	var arrayType *types.ArrayType
	if expectedType != nil {
		if at, ok := expectedType.(*types.ArrayType); ok {
			arrayType = at
		}
	}

	// Convert evaluator.Value slice to internal Value slice
	internalElements := make([]Value, len(elements))
	for idx, elem := range elements {
		internalElements[idx] = elem.(Value)
	}

	// Use existing method that handles type inference and conversion
	if arrayType != nil {
		return i.evalArrayLiteralWithExpected(&ast.ArrayLiteralExpression{Elements: nil}, arrayType)
	}

	// Fallback: create array without expected type
	if len(internalElements) > 0 {
		elemType := i.typeFromValue(internalElements[0])
		return i.CreateArray(elemType, elements)
	}

	// Empty array with unknown type (dynamic array)
	return NewArrayValue(&types.ArrayType{LowBound: nil, HighBound: nil})
}

// GetArrayElement retrieves an element from an array at the given index.
func (i *Interpreter) GetArrayElement(array evaluator.Value, index evaluator.Value) (evaluator.Value, error) {
	// Convert to internal types
	internalArray := array.(Value)
	internalIndex := index.(Value)

	// Check if it's an array
	arrayVal, ok := internalArray.(*ArrayValue)
	if !ok {
		return nil, fmt.Errorf("not an array: %s", internalArray.Type())
	}

	// Get index as integer
	indexInt, ok := internalIndex.(*IntegerValue)
	if !ok {
		return nil, fmt.Errorf("array index must be integer, got %s", internalIndex.Type())
	}

	// Convert to 0-based index (DWScript uses different conventions depending on array type)
	idx := int(indexInt.Value)

	// Handle low bound offset for static arrays
	if !arrayVal.ArrayType.IsDynamic() && arrayVal.ArrayType.LowBound != nil {
		idx = idx - *arrayVal.ArrayType.LowBound
	}

	// Bounds check
	if idx < 0 || idx >= len(arrayVal.Elements) {
		return nil, fmt.Errorf("array index %d out of bounds (0..%d)", idx, len(arrayVal.Elements)-1)
	}

	return arrayVal.Elements[idx], nil
}

// SetArrayElement sets an element in an array at the given index.
func (i *Interpreter) SetArrayElement(array evaluator.Value, index evaluator.Value, value evaluator.Value) error {
	// Convert to internal types
	internalArray := array.(Value)
	internalIndex := index.(Value)
	internalValue := value.(Value)

	// Check if it's an array
	arrayVal, ok := internalArray.(*ArrayValue)
	if !ok {
		return fmt.Errorf("not an array: %s", internalArray.Type())
	}

	// Get index as integer
	indexInt, ok := internalIndex.(*IntegerValue)
	if !ok {
		return fmt.Errorf("array index must be integer, got %s", internalIndex.Type())
	}

	// Convert to 0-based index
	idx := int(indexInt.Value)

	// Handle low bound offset for static arrays
	if !arrayVal.ArrayType.IsDynamic() && arrayVal.ArrayType.LowBound != nil {
		idx = idx - *arrayVal.ArrayType.LowBound
	}

	// Bounds check
	if idx < 0 || idx >= len(arrayVal.Elements) {
		return fmt.Errorf("array index %d out of bounds (0..%d)", idx, len(arrayVal.Elements)-1)
	}

	// Set element
	arrayVal.Elements[idx] = internalValue

	return nil
}

// GetArrayLength returns the length of an array (adapter method for evaluator).
func (i *Interpreter) GetArrayLength(array evaluator.Value) int {
	// Convert to internal type
	internalArray := array.(Value)

	// Check if it's an array
	arrayVal, ok := internalArray.(*ArrayValue)
	if !ok {
		return 0
	}

	return len(arrayVal.Elements)
}

// CreateSet creates a set from a list of elements with a specified element type.
func (i *Interpreter) CreateSet(elementType any, elements []evaluator.Value) evaluator.Value {
	// Convert elementType to types.Type
	var typedElementType types.Type
	if elementType != nil {
		if t, ok := elementType.(types.Type); ok {
			typedElementType = t
		}
	}

	// Create set type
	setType := &types.SetType{
		ElementType: typedElementType,
	}

	// Create set value
	setValue := NewSetValue(setType)

	// Add elements
	for _, elem := range elements {
		internalElem := elem.(Value)

		// Get ordinal value
		ordinal, err := GetOrdinalValue(internalElem)
		if err != nil {
			// Skip non-ordinal elements
			continue
		}

		setValue.AddElement(ordinal)
	}

	return setValue
}

// EvaluateSetRange expands a range expression (e.g., 1..10, 'a'..'z') into ordinal values.
func (i *Interpreter) EvaluateSetRange(start evaluator.Value, end evaluator.Value) ([]int, error) {
	// Convert to internal types
	internalStart := start.(Value)
	internalEnd := end.(Value)

	// Get ordinal values
	startOrd, err := GetOrdinalValue(internalStart)
	if err != nil {
		return nil, fmt.Errorf("range start must be ordinal type: %s", err.Error())
	}

	endOrd, err := GetOrdinalValue(internalEnd)
	if err != nil {
		return nil, fmt.Errorf("range end must be ordinal type: %s", err.Error())
	}

	// Expand range
	var ordinals []int
	if startOrd <= endOrd {
		for ord := startOrd; ord <= endOrd; ord++ {
			ordinals = append(ordinals, ord)
		}
	} else {
		// Reverse range
		for ord := startOrd; ord >= endOrd; ord-- {
			ordinals = append(ordinals, ord)
		}
	}

	return ordinals, nil
}

// AddToSet adds an element to a set.
func (i *Interpreter) AddToSet(set evaluator.Value, element evaluator.Value) error {
	// Convert to internal types
	internalSet := set.(Value)
	internalElement := element.(Value)

	// Check if it's a set
	setValue, ok := internalSet.(*SetValue)
	if !ok {
		return fmt.Errorf("not a set: %s", internalSet.Type())
	}

	// Get ordinal value of element
	ordinal, err := GetOrdinalValue(internalElement)
	if err != nil {
		return fmt.Errorf("cannot add non-ordinal element to set: %s", err.Error())
	}

	// Add to set
	setValue.AddElement(ordinal)

	return nil
}

// GetStringChar retrieves a character from a string at the given index (1-based).
func (i *Interpreter) GetStringChar(str evaluator.Value, index evaluator.Value) (evaluator.Value, error) {
	// Convert to internal types
	internalStr := str.(Value)
	internalIndex := index.(Value)

	// Get string value
	strVal, ok := internalStr.(*StringValue)
	if !ok {
		return nil, fmt.Errorf("not a string: %s", internalStr.Type())
	}

	// Get index as integer
	indexInt, ok := internalIndex.(*IntegerValue)
	if !ok {
		return nil, fmt.Errorf("string index must be integer, got %s", internalIndex.Type())
	}

	// DWScript uses 1-based indexing for strings
	idx := int(indexInt.Value) - 1

	// Bounds check
	runes := []rune(strVal.Value)
	if idx < 0 || idx >= len(runes) {
		return nil, fmt.Errorf("string index %d out of bounds (1..%d)", int(indexInt.Value), len(runes))
	}

	// Return character as string
	return &StringValue{Value: string(runes[idx])}, nil
}

// ===== Task 3.5.7: Property, Field, and Member Access Adapter Methods =====

// GetObjectField retrieves a field value from an object.
func (i *Interpreter) GetObjectField(obj evaluator.Value, fieldName string) (evaluator.Value, error) {
	// Convert to internal type
	internalObj := obj.(Value)

	// Get object instance
	objVal, ok := internalObj.(*ObjectInstance)
	if !ok {
		return nil, fmt.Errorf("not an object: %s", internalObj.Type())
	}

	// Look up field value
	fieldValue, exists := objVal.Fields[strings.ToLower(fieldName)]
	if !exists {
		return nil, fmt.Errorf("field '%s' not found in object", fieldName)
	}

	return fieldValue, nil
}

// SetObjectField sets a field value in an object.
func (i *Interpreter) SetObjectField(obj evaluator.Value, fieldName string, value evaluator.Value) error {
	// Convert to internal types
	internalObj := obj.(Value)
	internalValue := value.(Value)

	// Get object instance
	objVal, ok := internalObj.(*ObjectInstance)
	if !ok {
		return fmt.Errorf("not an object: %s", internalObj.Type())
	}

	// Verify field exists in class definition (case-insensitive)
	fieldNameLower := strings.ToLower(fieldName)
	if _, exists := objVal.Class.Fields[fieldNameLower]; !exists {
		return fmt.Errorf("field '%s' not found in class '%s'", fieldName, objVal.Class.Name)
	}

	// Set field value (case-insensitive)
	objVal.Fields[fieldNameLower] = internalValue
	return nil
}

// GetRecordField retrieves a field value from a record.
func (i *Interpreter) GetRecordField(record evaluator.Value, fieldName string) (evaluator.Value, error) {
	// Convert to internal type
	internalRecord := record.(Value)

	// Get record value
	recVal, ok := internalRecord.(*RecordValue)
	if !ok {
		return nil, fmt.Errorf("not a record: %s", internalRecord.Type())
	}

	// Look up field value
	fieldValue, exists := recVal.Fields[strings.ToLower(fieldName)]
	if !exists {
		return nil, fmt.Errorf("field '%s' not found in record", fieldName)
	}

	return fieldValue, nil
}

// SetRecordField sets a field value in a record.
func (i *Interpreter) SetRecordField(record evaluator.Value, fieldName string, value evaluator.Value) error {
	// Convert to internal types
	internalRecord := record.(Value)
	internalValue := value.(Value)

	// Get record value
	recVal, ok := internalRecord.(*RecordValue)
	if !ok {
		return fmt.Errorf("not a record: %s", internalRecord.Type())
	}

	// Verify field exists in record type definition (case-insensitive)
	fieldNameLower := strings.ToLower(fieldName)
	if recVal.RecordType != nil {
		if _, exists := recVal.RecordType.Fields[fieldNameLower]; !exists {
			return fmt.Errorf("field '%s' not found in record type '%s'", fieldName, recVal.RecordType.Name)
		}
	}

	// Set field value (case-insensitive)
	recVal.Fields[fieldNameLower] = internalValue
	return nil
}

// GetPropertyValue retrieves a property value from an object.
func (i *Interpreter) GetPropertyValue(obj evaluator.Value, propName string) (evaluator.Value, error) {
	// Convert to internal type
	internalObj := obj.(Value)

	// Get object instance
	objVal, ok := internalObj.(*ObjectInstance)
	if !ok {
		return nil, fmt.Errorf("not an object: %s", internalObj.Type())
	}

	// Get class info
	classInfo := objVal.Class
	if classInfo == nil {
		return nil, fmt.Errorf("object has no class information")
	}

	// Find property (case-insensitive)
	propNameLower := strings.ToLower(propName)
	prop, exists := classInfo.Properties[propNameLower]
	if !exists {
		return nil, fmt.Errorf("property '%s' not found in class '%s'", propName, classInfo.Name)
	}

	// Delegate to existing property read infrastructure
	result := i.evalPropertyRead(objVal, prop, nil)
	if isError(result) {
		return nil, fmt.Errorf("property read failed: %v", result)
	}

	return result, nil
}

// SetPropertyValue sets a property value in an object.
func (i *Interpreter) SetPropertyValue(obj evaluator.Value, propName string, value evaluator.Value) error {
	// Convert to internal types
	internalObj := obj.(Value)
	internalValue := value.(Value)

	// Get object instance
	objVal, ok := internalObj.(*ObjectInstance)
	if !ok {
		return fmt.Errorf("not an object: %s", internalObj.Type())
	}

	// Get class info
	classInfo := objVal.Class
	if classInfo == nil {
		return fmt.Errorf("object has no class information")
	}

	// Find property (case-insensitive)
	propNameLower := strings.ToLower(propName)
	prop, exists := classInfo.Properties[propNameLower]
	if !exists {
		return fmt.Errorf("property '%s' not found in class '%s'", propName, classInfo.Name)
	}

	// Check if property is read-only
	if prop.WriteKind == types.PropAccessNone {
		return fmt.Errorf("property '%s' is read-only", propName)
	}

	// Handle property write based on WriteKind
	switch prop.WriteKind {
	case types.PropAccessField:
		// Direct field assignment
		objVal.Fields[strings.ToLower(prop.WriteSpec)] = internalValue
		return nil

	case types.PropAccessMethod:
		// Call setter method
		method := objVal.Class.lookupMethod(prop.WriteSpec)
		if method == nil {
			return fmt.Errorf("property '%s' setter method '%s' not found", propName, prop.WriteSpec)
		}

		savedEnv := i.env
		tempEnv := NewEnclosedEnvironment(i.env)
		tempEnv.Define("Self", objVal)
		i.env = tempEnv

		args := []Value{internalValue}
		result := i.callUserFunction(method, args)

		i.env = savedEnv

		if isError(result) {
			return fmt.Errorf("property setter failed: %v", result)
		}
		return nil

	default:
		return fmt.Errorf("property '%s' has unsupported write access kind", propName)
	}
}

// GetIndexedProperty retrieves an indexed property value from an object.
func (i *Interpreter) GetIndexedProperty(obj evaluator.Value, propName string, indices []evaluator.Value) (evaluator.Value, error) {
	// Convert to internal type
	internalObj := obj.(Value)

	// Convert indices
	internalIndices := make([]Value, len(indices))
	for idx, index := range indices {
		internalIndices[idx] = index.(Value)
	}

	// Get object instance
	objVal, ok := internalObj.(*ObjectInstance)
	if !ok {
		return nil, fmt.Errorf("not an object: %s", internalObj.Type())
	}

	// Get class info
	classInfo := objVal.Class
	if classInfo == nil {
		return nil, fmt.Errorf("object has no class information")
	}

	// Find indexed property (case-insensitive)
	propNameLower := strings.ToLower(propName)
	prop, exists := classInfo.Properties[propNameLower]
	if !exists {
		return nil, fmt.Errorf("property '%s' not found in class '%s'", propName, classInfo.Name)
	}

	if !prop.IsIndexed {
		return nil, fmt.Errorf("property '%s' is not an indexed property", propName)
	}

	// Call getter method with indices
	if prop.ReadKind == types.PropAccessMethod {
		method := objVal.Class.lookupMethod(prop.ReadSpec)
		if method == nil {
			return nil, fmt.Errorf("indexed property '%s' getter method '%s' not found", propName, prop.ReadSpec)
		}

		savedEnv := i.env
		tempEnv := NewEnclosedEnvironment(i.env)
		tempEnv.Define("Self", objVal)
		i.env = tempEnv

		result := i.callUserFunction(method, internalIndices)

		i.env = savedEnv

		if isError(result) {
			return nil, fmt.Errorf("indexed property getter failed: %v", result)
		}
		return result, nil
	}

	return nil, fmt.Errorf("indexed property '%s' has unsupported read access kind", propName)
}

// SetIndexedProperty sets an indexed property value in an object.
func (i *Interpreter) SetIndexedProperty(obj evaluator.Value, propName string, indices []evaluator.Value, value evaluator.Value) error {
	// Convert to internal types
	internalObj := obj.(Value)
	internalValue := value.(Value)

	// Convert indices
	internalIndices := make([]Value, len(indices))
	for idx, index := range indices {
		internalIndices[idx] = index.(Value)
	}

	// Get object instance
	objVal, ok := internalObj.(*ObjectInstance)
	if !ok {
		return fmt.Errorf("not an object: %s", internalObj.Type())
	}

	// Get class info
	classInfo := objVal.Class
	if classInfo == nil {
		return fmt.Errorf("object has no class information")
	}

	// Find indexed property (case-insensitive)
	propNameLower := strings.ToLower(propName)
	prop, exists := classInfo.Properties[propNameLower]
	if !exists {
		return fmt.Errorf("property '%s' not found in class '%s'", propName, classInfo.Name)
	}

	if !prop.IsIndexed {
		return fmt.Errorf("property '%s' is not an indexed property", propName)
	}

	// Check if property is read-only
	if prop.WriteKind == types.PropAccessNone {
		return fmt.Errorf("indexed property '%s' is read-only", propName)
	}

	// Call setter method with indices + value
	if prop.WriteKind == types.PropAccessMethod {
		method := objVal.Class.lookupMethod(prop.WriteSpec)
		if method == nil {
			return fmt.Errorf("indexed property '%s' setter method '%s' not found", propName, prop.WriteSpec)
		}

		savedEnv := i.env
		tempEnv := NewEnclosedEnvironment(i.env)
		tempEnv.Define("Self", objVal)
		i.env = tempEnv

		// Append value to indices for setter call
		args := append(internalIndices, internalValue)
		result := i.callUserFunction(method, args)

		i.env = savedEnv

		if isError(result) {
			return fmt.Errorf("indexed property setter failed: %v", result)
		}
		return nil
	}

	return fmt.Errorf("indexed property '%s' has unsupported write access kind", propName)
}

// CallMethod executes a method on an object with the given arguments.
func (i *Interpreter) CallMethod(obj evaluator.Value, methodName string, args []evaluator.Value, node ast.Node) evaluator.Value {
	// Convert to internal types
	internalObj := obj.(Value)
	internalArgs := convertEvaluatorArgs(args)

	// Get object instance
	objVal, ok := internalObj.(*ObjectInstance)
	if !ok {
		panic(fmt.Sprintf("not an object: %s", internalObj.Type()))
	}

	// Get class info
	classInfo := objVal.Class
	if classInfo == nil {
		panic("object has no class information")
	}

	// Find method (case-insensitive) using the existing helper
	method := classInfo.lookupMethod(methodName)
	if method == nil {
		panic(fmt.Sprintf("method '%s' not found in class '%s'", methodName, classInfo.Name))
	}

	// Call the method using existing infrastructure
	savedEnv := i.env
	tempEnv := NewEnclosedEnvironment(i.env)
	tempEnv.Define("Self", objVal)
	i.env = tempEnv

	result := i.callUserFunction(method, internalArgs)

	i.env = savedEnv
	return result
}

// CallInheritedMethod executes an inherited (parent) method with the given arguments.
func (i *Interpreter) CallInheritedMethod(obj evaluator.Value, methodName string, args []evaluator.Value) evaluator.Value {
	// Convert to internal types
	internalObj := obj.(Value)
	internalArgs := convertEvaluatorArgs(args)

	// Get object instance
	objVal, ok := internalObj.(*ObjectInstance)
	if !ok {
		panic(fmt.Sprintf("not an object: %s", internalObj.Type()))
	}

	// Get class info
	classInfo := objVal.Class
	if classInfo == nil {
		panic("object has no class information")
	}

	// Check parent class
	if classInfo.Parent == nil {
		panic(fmt.Sprintf("class '%s' has no parent", classInfo.Name))
	}

	parentInfo := classInfo.Parent

	// Find method in parent (case-insensitive)
	methodNameLower := strings.ToLower(methodName)
	method, exists := parentInfo.Methods[methodNameLower]
	if !exists {
		panic(fmt.Sprintf("inherited method '%s' not found in parent class '%s'", methodName, parentInfo.Name))
	}

	// Call the method using existing infrastructure
	savedEnv := i.env
	tempEnv := NewEnclosedEnvironment(i.env)
	tempEnv.Define("Self", objVal)
	i.env = tempEnv

	result := i.callUserFunction(method, internalArgs)

	i.env = savedEnv
	return result
}

// CreateObject creates a new object instance of the specified class with constructor arguments.
func (i *Interpreter) CreateObject(className string, args []evaluator.Value) (evaluator.Value, error) {
	// Convert arguments
	internalArgs := convertEvaluatorArgs(args)

	// Look up class (case-insensitive)
	classInfo, exists := i.classes[strings.ToLower(className)]
	if !exists {
		return nil, fmt.Errorf("class '%s' not found", className)
	}

	// Create new object instance
	obj := NewObjectInstance(classInfo)

	// Initialize fields with default values
	savedEnv := i.env
	tempEnv := NewEnclosedEnvironment(i.env)
	i.env = tempEnv

	for fieldName, fieldType := range classInfo.Fields {
		var fieldValue Value
		if fieldDecl, hasDecl := classInfo.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
			fieldValue = i.Eval(fieldDecl.InitValue)
			if isError(fieldValue) {
				i.env = savedEnv
				return nil, fmt.Errorf("failed to initialize field '%s': %v", fieldName, fieldValue)
			}
		} else {
			fieldValue = getZeroValueForType(fieldType, nil)
		}
		obj.SetField(fieldName, fieldValue)
	}

	i.env = savedEnv

	// Call constructor if it exists
	constructorNameLower := strings.ToLower("Create")
	if constructor, exists := classInfo.Constructors[constructorNameLower]; exists {
		tempEnv := NewEnclosedEnvironment(i.env)
		tempEnv.Define("Self", obj)
		i.env = tempEnv

		result := i.callUserFunction(constructor, internalArgs)

		i.env = savedEnv

		// Propagate constructor errors
		if isError(result) {
			return nil, fmt.Errorf("constructor failed: %v", result)
		}
	} else if len(internalArgs) > 0 {
		return nil, fmt.Errorf("no constructor found for class '%s' with %d arguments", className, len(internalArgs))
	}

	return obj, nil
}

// CheckType checks if an object is of a specified type (implements 'is' operator).
func (i *Interpreter) CheckType(obj evaluator.Value, typeName string) bool {
	// Convert to internal type
	internalObj := obj.(Value)

	// Check if it's an object
	objVal, ok := internalObj.(*ObjectInstance)
	if !ok {
		return false
	}

	// Get class info
	classInfo := objVal.Class
	if classInfo == nil {
		return false
	}

	// Check if the object's class matches (case-insensitive)
	if strings.EqualFold(classInfo.Name, typeName) {
		return true
	}

	// Check parent class hierarchy
	current := classInfo.Parent
	for current != nil {
		if strings.EqualFold(current.Name, typeName) {
			return true
		}
		current = current.Parent
	}

	return false
}

// CastType performs type casting (implements 'as' operator).
// Task 3.5.8: Adapter method for type casting.
func (i *Interpreter) CastType(obj evaluator.Value, typeName string) (evaluator.Value, error) {
	// Convert to internal type
	internalObj := obj.(Value)

	// Check if the cast is valid
	if i.CheckType(obj, typeName) {
		return obj, nil
	}

	// Invalid cast - return error
	return nil, fmt.Errorf("cannot cast %s to %s", internalObj.Type(), typeName)
}

// CreateFunctionPointer creates a function pointer value from a function declaration.
// Task 3.5.8: Adapter method for function pointer creation.
func (i *Interpreter) CreateFunctionPointer(fn *ast.FunctionDecl, closure any) evaluator.Value {
	// Convert closure to Environment
	var env *Environment
	if closure != nil {
		env = closure.(*Environment)
	}

	return &FunctionPointerValue{
		Function: fn,
		Closure:  env,
	}
}

// CreateLambda creates a lambda/closure value from a lambda expression.
// Task 3.5.8: Adapter method for lambda creation.
func (i *Interpreter) CreateLambda(lambda *ast.LambdaExpression, closure any) evaluator.Value {
	// Convert closure to Environment
	var env *Environment
	if closure != nil {
		env = closure.(*Environment)
	}

	return &FunctionPointerValue{
		Lambda:  lambda,
		Closure: env,
	}
}

// IsFunctionPointer checks if a value is a function pointer.
// Task 3.5.8: Adapter method for function pointer type checking.
func (i *Interpreter) IsFunctionPointer(value evaluator.Value) bool {
	_, ok := value.(*FunctionPointerValue)
	return ok
}

// GetFunctionPointerParamCount returns the number of parameters a function pointer expects.
// Task 3.5.8: Adapter method for function pointer parameter count.
func (i *Interpreter) GetFunctionPointerParamCount(funcPtr evaluator.Value) int {
	fp, ok := funcPtr.(*FunctionPointerValue)
	if !ok {
		return 0
	}

	if fp.Function != nil {
		return len(fp.Function.Parameters)
	} else if fp.Lambda != nil {
		return len(fp.Lambda.Parameters)
	}

	return 0
}

// IsFunctionPointerNil checks if a function pointer is nil (unassigned).
// Task 3.5.8: Adapter method for function pointer nil checking.
func (i *Interpreter) IsFunctionPointerNil(funcPtr evaluator.Value) bool {
	fp, ok := funcPtr.(*FunctionPointerValue)
	if !ok {
		return false
	}

	// A function pointer is nil if both Function and Lambda are nil
	return fp.Function == nil && fp.Lambda == nil
}

// CreateRecord creates a record value from field values.
// Task 3.5.8: Adapter method for record construction.
func (i *Interpreter) CreateRecord(recordType string, fields map[string]evaluator.Value) (evaluator.Value, error) {
	// Look up the record type
	recordTypeKey := "__record_type_" + strings.ToLower(recordType)
	typeVal, ok := i.env.Get(recordTypeKey)
	if !ok {
		return nil, fmt.Errorf("unknown record type '%s'", recordType)
	}

	rtv, ok := typeVal.(*RecordTypeValue)
	if !ok {
		return nil, fmt.Errorf("'%s' is not a record type", recordType)
	}

	// Convert field values to internal types
	internalFields := make(map[string]Value)
	for name, val := range fields {
		internalFields[name] = val.(Value)
	}

	// Create record value
	return &RecordValue{
		RecordType: rtv.RecordType,
		Fields:     internalFields,
		Methods:    rtv.Methods,
	}, nil
}

// SetVariable assigns a value to a variable in the execution context.
// Task 3.5.7: Adapter method for variable assignment.
func (i *Interpreter) SetVariable(name string, value evaluator.Value, ctx *evaluator.ExecutionContext) error {
	// Convert to internal type
	internalValue := value.(Value)

	// Set the variable in the environment
	env := ctx.Env()
	envAdapter, ok := env.(interface{ SetInternal(string, Value) bool })
	if !ok {
		return fmt.Errorf("cannot set variable '%s'", name)
	}

	if !envAdapter.SetInternal(name, internalValue) {
		return fmt.Errorf("undefined variable '%s'", name)
	}

	return nil
}

// CanAssign checks if an AST node can be used as an lvalue (assignment target).
// Task 3.5.8: Adapter method for lvalue validation.
func (i *Interpreter) CanAssign(target ast.Node) bool {
	switch target.(type) {
	case *ast.Identifier:
		return true
	case *ast.MemberAccessExpression:
		return true
	case *ast.IndexExpression:
		return true
	default:
		return false
	}
}

// RaiseException raises an exception with the given class name and message.
// Task 3.5.8: Adapter method for exception handling.
func (i *Interpreter) RaiseException(className string, message string, pos any) {
	// Convert pos to lexer.Position if provided
	var position *lexer.Position
	if pos != nil {
		if p, ok := pos.(*lexer.Position); ok {
			position = p
		}
	}

	// Call the internal raiseException method
	i.raiseException(className, message, position)
}

// GetVariable retrieves a variable value from the execution context.
// Task 3.5.9: Adapter method for environment access.
func (i *Interpreter) GetVariable(name string, ctx *evaluator.ExecutionContext) (evaluator.Value, bool) {
	// Get the value from the context's environment
	val, found := ctx.Env().Get(name)
	if !found {
		return nil, false
	}

	// Convert to evaluator.Value type
	return val.(evaluator.Value), true
}

// DefineVariable defines a new variable in the execution context.
// Task 3.5.9: Adapter method for environment access.
func (i *Interpreter) DefineVariable(name string, value evaluator.Value, ctx *evaluator.ExecutionContext) {
	// Convert to internal Value type
	internalValue := value.(Value)

	// Define in the context's environment
	ctx.Env().Define(name, internalValue)
}

// CreateEnclosedEnvironment creates a new execution context with an enclosed environment.
// Task 3.5.9: Adapter method for environment access.
func (i *Interpreter) CreateEnclosedEnvironment(ctx *evaluator.ExecutionContext) *evaluator.ExecutionContext {
	// Get the current environment from the context and create an enclosed scope
	currentEnv := ctx.Env()
	newEnv := currentEnv.NewEnclosedEnvironment()

	// Create a new execution context with the enclosed environment
	newCtx := evaluator.NewExecutionContext(newEnv)

	return newCtx
}

// GetCallStack returns a copy of the current call stack.
// Returns stack frames in the order they were called (oldest to newest).
func (i *Interpreter) GetCallStack() errors.StackTrace {
	// Return a copy to prevent external modification
	stack := make(errors.StackTrace, len(i.callStack))
	copy(stack, i.callStack)
	return stack
}

// pushCallStack adds a new frame to the call stack with the given function name.
// The position is taken from the current node being evaluated.
// Phase 3.3.3: Delegates to ExecutionContext's CallStack.
func (i *Interpreter) pushCallStack(functionName string) {
	var pos *lexer.Position
	if i.currentNode != nil {
		nodePos := i.currentNode.Pos()
		pos = &nodePos
	}
	// Also push to the old callStack field for backward compatibility
	frame := errors.NewStackFrame(functionName, i.sourceFile, pos)
	i.callStack = append(i.callStack, frame)

	// Phase 3.3.3: Also push to context's CallStack
	// Ignore errors here for backward compatibility; callers should check WillOverflow first
	_ = i.ctx.GetCallStack().Push(functionName, i.sourceFile, pos)
}

// popCallStack removes the most recent frame from the call stack.
// Phase 3.3.3: Delegates to ExecutionContext's CallStack.
func (i *Interpreter) popCallStack() {
	if len(i.callStack) > 0 {
		i.callStack = i.callStack[:len(i.callStack)-1]
	}
	// Phase 3.3.3: Also pop from context's CallStack
	i.ctx.GetCallStack().Pop()
}

// Eval evaluates an AST node and returns its value.
// This is the main entry point for the interpreter.
func (i *Interpreter) Eval(node ast.Node) Value {
	// Track the current node for error reporting
	i.currentNode = node

	switch node := node.(type) {
	// Program
	case *ast.Program:
		return i.evalProgram(node)

	// Statements
	case *ast.ExpressionStatement:
		// Evaluate the expression
		val := i.Eval(node.Expression)
		if isError(val) {
			return val
		}

		// Auto-invoke parameterless function pointers stored in variables
		// In DWScript, when a variable holds a function pointer with no parameters
		// and is used as a statement, it's automatically invoked
		// Example: var fp := @SomeProc; fp; // auto-invokes SomeProc
		if funcPtr, isFuncPtr := val.(*FunctionPointerValue); isFuncPtr {
			// Determine parameter count
			paramCount := 0
			if funcPtr.Function != nil {
				paramCount = len(funcPtr.Function.Parameters)
			} else if funcPtr.Lambda != nil {
				paramCount = len(funcPtr.Lambda.Parameters)
			}

			// If it has zero parameters, auto-invoke it
			if paramCount == 0 {
				// Check if the function pointer is nil (not assigned)
				if funcPtr.Function == nil && funcPtr.Lambda == nil {
					// Raise an exception that can be caught by try-except
					i.raiseException("Exception", "Function pointer is nil", &node.Token.Pos)
					return &NilValue{}
				}
				return i.callFunctionPointer(funcPtr, []Value{}, node)
			}
		}

		return val

	case *ast.VarDeclStatement:
		return i.evalVarDeclStatement(node)

	case *ast.ConstDecl:
		return i.evalConstDecl(node)

	case *ast.AssignmentStatement:
		return i.evalAssignmentStatement(node)

	case *ast.BlockStatement:
		return i.evalBlockStatement(node)

	case *ast.IfStatement:
		return i.evalIfStatement(node)

	case *ast.WhileStatement:
		return i.evalWhileStatement(node)

	case *ast.RepeatStatement:
		return i.evalRepeatStatement(node)

	case *ast.ForStatement:
		return i.evalForStatement(node)

	case *ast.ForInStatement:
		return i.evalForInStatement(node)

	case *ast.CaseStatement:
		return i.evalCaseStatement(node)

	case *ast.TryStatement:
		return i.evalTryStatement(node)

	case *ast.RaiseStatement:
		return i.evalRaiseStatement(node)

	case *ast.BreakStatement:
		return i.evalBreakStatement(node)

	case *ast.ContinueStatement:
		return i.evalContinueStatement(node)

	case *ast.ExitStatement:
		return i.evalExitStatement(node)

	case *ast.ReturnStatement:
		// Handle return statements in lambda shorthand syntax
		return i.evalReturnStatement(node)

	case *ast.UsesClause:
		// Uses clauses are processed before execution by the CLI/loader
		// At runtime, they're no-ops since units are already loaded
		return nil

	case *ast.FunctionDecl:
		return i.evalFunctionDeclaration(node)

	case *ast.ClassDecl:
		return i.evalClassDeclaration(node)

	case *ast.InterfaceDecl:
		return i.evalInterfaceDeclaration(node)

	case *ast.OperatorDecl:
		return i.evalOperatorDeclaration(node)

	case *ast.EnumDecl:
		return i.evalEnumDeclaration(node)

	case *ast.RecordDecl:
		return i.evalRecordDeclaration(node)

	case *ast.HelperDecl:
		return i.evalHelperDeclaration(node)

	case *ast.ArrayDecl:
		return i.evalArrayDeclaration(node)

	case *ast.TypeDeclaration:
		return i.evalTypeDeclaration(node)

	// Expressions
	case *ast.IntegerLiteral:
		return &IntegerValue{Value: node.Value}

	case *ast.FloatLiteral:
		return &FloatValue{Value: node.Value}

	case *ast.StringLiteral:
		return &StringValue{Value: node.Value}

	case *ast.BooleanLiteral:
		return &BooleanValue{Value: node.Value}

	case *ast.CharLiteral:
		// Character literals are treated as single-character strings
		return &StringValue{Value: string(node.Value)}

	case *ast.NilLiteral:
		return &NilValue{}

	case *ast.Identifier:
		return i.evalIdentifier(node)

	case *ast.BinaryExpression:
		return i.evalBinaryExpression(node)

	case *ast.UnaryExpression:
		return i.evalUnaryExpression(node)

	case *ast.AddressOfExpression:
		return i.evalAddressOfExpression(node)

	case *ast.GroupedExpression:
		return i.Eval(node.Expression)

	case *ast.CallExpression:
		return i.evalCallExpression(node)

	case *ast.NewExpression:
		return i.evalNewExpression(node)

	case *ast.MemberAccessExpression:
		return i.evalMemberAccess(node)

	case *ast.MethodCallExpression:
		return i.evalMethodCall(node)

	case *ast.InheritedExpression:
		return i.evalInheritedExpression(node)

	case *ast.SelfExpression:
		return i.evalSelfExpression(node)

	case *ast.EnumLiteral:
		return i.evalEnumLiteral(node)

	case *ast.RecordLiteralExpression:
		return i.evalRecordLiteral(node)

	case *ast.SetLiteral:
		return i.evalSetLiteral(node)

	case *ast.ArrayLiteralExpression:
		return i.evalArrayLiteral(node)

	case *ast.IndexExpression:
		return i.evalIndexExpression(node)

	case *ast.NewArrayExpression:
		return i.evalNewArrayExpression(node)

	case *ast.LambdaExpression:
		// Evaluate lambda expression to create closure
		return i.evalLambdaExpression(node)

	case *ast.IsExpression:
		// Task 9.40: Evaluate 'is' type checking operator
		return i.evalIsExpression(node)

	case *ast.AsExpression:
		// Task 9.48: Evaluate 'as' type casting operator
		return i.evalAsExpression(node)

	case *ast.ImplementsExpression:
		// Task 9.48: Evaluate 'implements' interface checking operator
		return i.evalImplementsExpression(node)

	case *ast.IfExpression:
		// Task 9.217: Evaluate inline if-then-else expressions
		return i.evalIfExpression(node)

	case *ast.OldExpression:
		// Evaluate 'old' expressions in postconditions
		identName := node.Identifier.Value
		oldValue, found := i.getOldValue(identName)
		if !found {
			return newError("old value for '%s' not captured (internal error)", identName)
		}
		return oldValue

	default:
		return newError("unknown node type: %T", node)
	}
}

// EvalWithExpectedType evaluates a node with an expected type for better type inference.
// This is primarily used for array literals in function calls where the parameter type is known.
// If expectedType is nil, this falls back to regular Eval().
func (i *Interpreter) EvalWithExpectedType(node ast.Node, expectedType types.Type) Value {
	// Special handling for array literals with expected array type
	if arrayLit, ok := node.(*ast.ArrayLiteralExpression); ok {
		if arrayType, ok := expectedType.(*types.ArrayType); ok {
			return i.evalArrayLiteralWithExpected(arrayLit, arrayType)
		}
	}

	// For all other cases, use regular Eval
	return i.Eval(node)
}
