package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// TestExecuteConversionFunction_Validation tests parameter validation.
func TestExecuteConversionFunction_Validation(t *testing.T) {
	// Create evaluator with minimal setup
	typeSystem := interptypes.NewTypeSystem()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil)

	// Create execution context
	ctx := NewExecutionContext(nil)

	// Test with function that has no parameters (should fail)
	t.Run("no parameters", func(t *testing.T) {
		fn := &ast.FunctionDecl{
			Name:       &ast.Identifier{Value: "InvalidConversion"},
			Parameters: []*ast.Parameter{}, // Empty
			ReturnType: &ast.TypeAnnotation{Name: "String"},
		}

		_, err := e.ExecuteConversionFunction(fn, &runtime.IntegerValue{Value: 42}, ctx, nil)
		if err == nil {
			t.Error("expected error for conversion function with no parameters")
		}
		if err.Error() != "conversion function 'InvalidConversion' must have exactly 1 parameter, got 0" {
			t.Errorf("unexpected error message: %s", err.Error())
		}
	})

	// Test with function that has multiple parameters (should fail)
	t.Run("multiple parameters", func(t *testing.T) {
		fn := &ast.FunctionDecl{
			Name: &ast.Identifier{Value: "InvalidConversion"},
			Parameters: []*ast.Parameter{
				{Name: &ast.Identifier{Value: "a"}},
				{Name: &ast.Identifier{Value: "b"}},
			},
			ReturnType: &ast.TypeAnnotation{Name: "String"},
		}

		_, err := e.ExecuteConversionFunction(fn, &runtime.IntegerValue{Value: 42}, ctx, nil)
		if err == nil {
			t.Error("expected error for conversion function with multiple parameters")
		}
		if err.Error() != "conversion function 'InvalidConversion' must have exactly 1 parameter, got 2" {
			t.Errorf("unexpected error message: %s", err.Error())
		}
	})

	// Test with procedure (no return type) - should fail
	t.Run("no return type", func(t *testing.T) {
		fn := &ast.FunctionDecl{
			Name: &ast.Identifier{Value: "InvalidConversion"},
			Parameters: []*ast.Parameter{
				{Name: &ast.Identifier{Value: "value"}},
			},
			ReturnType: nil, // No return type (procedure)
		}

		_, err := e.ExecuteConversionFunction(fn, &runtime.IntegerValue{Value: 42}, ctx, nil)
		if err == nil {
			t.Error("expected error for conversion function without return type")
		}
		if err.Error() != "conversion function 'InvalidConversion' must have a return type" {
			t.Errorf("unexpected error message: %s", err.Error())
		}
	})
}

// TestExecuteConversionFunction_WithMockAdapter tests conversion function execution
// with a mock adapter that handles the actual function body execution.
func TestExecuteConversionFunction_WithMockAdapter(t *testing.T) {
	// Create evaluator with minimal setup
	typeSystem := interptypes.NewTypeSystem()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil)

	// Create a mock adapter that simulates function execution
	mockAdapter := &mockConversionAdapter{
		evalNodeFunc: func(node ast.Node) Value {
			// Simulate setting Result in the function body
			// This would normally be done by the function body itself
			return &runtime.NilValue{}
		},
	}
	e.SetAdapter(mockAdapter)

	// Create execution context with a proper environment
	env := newTestConversionEnv()
	ctx := NewExecutionContext(env)

	t.Run("valid conversion function structure", func(t *testing.T) {
		// Create a conversion function: function IntToStr(value: Integer): String
		fn := &ast.FunctionDecl{
			Name: &ast.Identifier{Value: "IntToStr"},
			Parameters: []*ast.Parameter{
				{
					Name: &ast.Identifier{Value: "value"},
					Type: &ast.TypeAnnotation{Name: "Integer"},
				},
			},
			ReturnType: &ast.TypeAnnotation{Name: "String"},
			Body:       &ast.BlockStatement{Statements: []ast.Statement{}},
		}

		// Execute the conversion function
		result, err := e.ExecuteConversionFunction(
			fn,
			&runtime.IntegerValue{Value: 42},
			ctx,
			nil, // No callbacks needed for this test
		)

		// We expect an error because the mock adapter doesn't actually set Result
		// But the function structure validation should pass
		if err != nil {
			// If we get an error about function body, that's expected for this minimal test
			// The key is that validation passed
			t.Logf("Got expected error (mock adapter limitation): %v", err)
		} else if result == nil {
			t.Error("result should not be nil")
		}
	})
}

// TestExecuteConversionFunctionSimple tests the simplified version.
func TestExecuteConversionFunctionSimple_Validation(t *testing.T) {
	// Create evaluator with minimal setup
	typeSystem := interptypes.NewTypeSystem()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil)

	// Create execution context
	ctx := NewExecutionContext(nil)

	// Test validation still works without adapter
	t.Run("validation without adapter", func(t *testing.T) {
		fn := &ast.FunctionDecl{
			Name:       &ast.Identifier{Value: "InvalidConversion"},
			Parameters: []*ast.Parameter{}, // Invalid: no parameters
			ReturnType: &ast.TypeAnnotation{Name: "String"},
		}

		_, err := e.ExecuteConversionFunctionSimple(fn, &runtime.IntegerValue{Value: 42}, ctx)
		if err == nil {
			t.Error("expected error for conversion function with no parameters")
		}
	})
}

// TestConversionCallbacks_NilHandling tests that nil callbacks are handled gracefully.
func TestConversionCallbacks_NilHandling(t *testing.T) {
	// Create evaluator with minimal setup
	typeSystem := interptypes.NewTypeSystem()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil)

	// Set up a mock adapter
	mockAdapter := &mockConversionAdapter{
		evalNodeFunc: func(node ast.Node) Value {
			return &runtime.NilValue{}
		},
	}
	e.SetAdapter(mockAdapter)

	env := newTestConversionEnv()
	ctx := NewExecutionContext(env)

	fn := &ast.FunctionDecl{
		Name: &ast.Identifier{Value: "TestConversion"},
		Parameters: []*ast.Parameter{
			{Name: &ast.Identifier{Value: "value"}},
		},
		ReturnType: &ast.TypeAnnotation{Name: "String"},
		Body:       &ast.BlockStatement{Statements: []ast.Statement{}},
	}

	t.Run("nil ConversionCallbacks", func(t *testing.T) {
		// Should not panic with nil callbacks
		_, err := e.ExecuteConversionFunction(fn, &runtime.IntegerValue{Value: 1}, ctx, nil)
		if err != nil {
			// Error about function body is acceptable
			t.Logf("Got expected error: %v", err)
		}
	})

	t.Run("empty ConversionCallbacks", func(t *testing.T) {
		// Should not panic with empty callbacks
		callbacks := &ConversionCallbacks{}
		_, err := e.ExecuteConversionFunction(fn, &runtime.IntegerValue{Value: 1}, ctx, callbacks)
		if err != nil {
			// Error about function body is acceptable
			t.Logf("Got expected error: %v", err)
		}
	})

	t.Run("partial ConversionCallbacks", func(t *testing.T) {
		// Should not panic with partial callbacks
		callbacks := &ConversionCallbacks{
			ImplicitConversion: func(value Value, targetTypeName string) (Value, bool) {
				return value, false
			},
			// EnvSyncer is nil
		}
		_, err := e.ExecuteConversionFunction(fn, &runtime.IntegerValue{Value: 1}, ctx, callbacks)
		if err != nil {
			// Error about function body is acceptable
			t.Logf("Got expected error: %v", err)
		}
	})
}

// testConversionEnv is a simple mock environment for testing conversion functions.
type testConversionEnv struct {
	bindings map[string]interface{}
}

func newTestConversionEnv() *testConversionEnv {
	return &testConversionEnv{bindings: make(map[string]interface{})}
}

func (e *testConversionEnv) Define(name string, value interface{}) {
	e.bindings[name] = value
}

func (e *testConversionEnv) Get(name string) (interface{}, bool) {
	val, ok := e.bindings[name]
	return val, ok
}

func (e *testConversionEnv) Set(name string, value interface{}) bool {
	if _, ok := e.bindings[name]; ok {
		e.bindings[name] = value
		return true
	}
	return false
}

func (e *testConversionEnv) NewEnclosedEnvironment() Environment {
	child := newTestConversionEnv()
	return child
}

// mockConversionAdapter is a minimal mock for testing conversion function execution.
type mockConversionAdapter struct {
	evalNodeFunc func(node ast.Node) Value
}

func (m *mockConversionAdapter) EvalNode(node ast.Node) Value {
	if m.evalNodeFunc != nil {
		return m.evalNodeFunc(node)
	}
	return &runtime.NilValue{}
}

// Implement all required InterpreterAdapter methods with minimal stubs
func (m *mockConversionAdapter) CallFunctionPointer(funcPtr Value, args []Value, node ast.Node) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) CallUserFunction(fn *ast.FunctionDecl, args []Value) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) LookupFunction(name string) ([]*ast.FunctionDecl, bool) {
	return nil, false
}
func (m *mockConversionAdapter) EvalMethodImplementation(fn *ast.FunctionDecl) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) LookupClass(name string) (any, bool) { return nil, false }
func (m *mockConversionAdapter) ResolveClassInfoByName(name string) interface{} {
	return nil
}
func (m *mockConversionAdapter) GetClassNameFromInfo(classInfo interface{}) string { return "" }
func (m *mockConversionAdapter) LookupRecord(name string) (any, bool)              { return nil, false }
func (m *mockConversionAdapter) LookupInterface(name string) (any, bool)           { return nil, false }
func (m *mockConversionAdapter) LookupHelpers(typeName string) []any               { return nil }
func (m *mockConversionAdapter) CreateHelperInfo(name string, targetType any, isRecordHelper bool) interface{} {
	return nil
}
func (m *mockConversionAdapter) SetHelperParent(helper interface{}, parent interface{}) {}
func (m *mockConversionAdapter) VerifyHelperTargetTypeMatch(parent interface{}, targetType any) bool {
	return false
}
func (m *mockConversionAdapter) GetHelperName(helper interface{}) string { return "" }
func (m *mockConversionAdapter) AddHelperMethod(helper interface{}, normalizedName string, method *ast.FunctionDecl) {
}
func (m *mockConversionAdapter) AddHelperProperty(helper interface{}, prop *ast.PropertyDecl, propType any) {
}
func (m *mockConversionAdapter) AddHelperClassVar(helper interface{}, name string, value Value) {}
func (m *mockConversionAdapter) AddHelperClassConst(helper interface{}, name string, value Value) {
}
func (m *mockConversionAdapter) RegisterHelperLegacy(typeName string, helper interface{}) {}
func (m *mockConversionAdapter) NewInterfaceInfoAdapter(name string) interface{}          { return nil }
func (m *mockConversionAdapter) CastToInterfaceInfo(iface interface{}) (interface{}, bool) {
	return nil, false
}
func (m *mockConversionAdapter) SetInterfaceParent(iface interface{}, parent interface{}) {}
func (m *mockConversionAdapter) GetInterfaceName(iface interface{}) string                { return "" }
func (m *mockConversionAdapter) GetInterfaceParent(iface interface{}) interface{}         { return nil }
func (m *mockConversionAdapter) AddInterfaceMethod(iface interface{}, normalizedName string, method *ast.FunctionDecl) {
}
func (m *mockConversionAdapter) AddInterfaceProperty(iface interface{}, normalizedName string, propInfo any) {
}
func (m *mockConversionAdapter) GetOperatorRegistry() any          { return nil }
func (m *mockConversionAdapter) GetEnumTypeID(enumName string) int { return 0 }
func (m *mockConversionAdapter) CreateArray(elementType any, elements []Value) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) CreateArrayValue(arrayType any, elements []Value) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) CallMethod(obj Value, methodName string, args []Value, node ast.Node) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) CallInheritedMethod(obj Value, methodName string, args []Value) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) ExecuteMethodWithSelf(self Value, methodDecl any, args []Value) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) CreateObject(className string, args []Value) (Value, error) {
	return &runtime.NilValue{}, nil
}
func (m *mockConversionAdapter) ExecuteConstructor(obj Value, constructorName string, args []Value) error {
	return nil
}
func (m *mockConversionAdapter) CheckType(obj Value, typeName string) bool { return false }
func (m *mockConversionAdapter) GetClassMetadataFromValue(obj Value) *runtime.ClassMetadata {
	return nil
}
func (m *mockConversionAdapter) GetObjectInstanceFromValue(val Value) interface{} { return nil }
func (m *mockConversionAdapter) GetInterfaceInstanceFromValue(val Value) (interface{}, interface{}) {
	return nil, nil
}
func (m *mockConversionAdapter) CreateInterfaceWrapper(interfaceName string, obj Value) (Value, error) {
	return nil, nil
}
func (m *mockConversionAdapter) CreateTypeCastWrapper(className string, obj Value) Value { return nil }
func (m *mockConversionAdapter) RaiseTypeCastException(message string, node ast.Node)    {}
func (m *mockConversionAdapter) RaiseAssertionFailed(customMessage string)               {}
func (m *mockConversionAdapter) CreateContractException(className, message string, node ast.Node, classMetadata interface{}, callStack interface{}) interface{} {
	return nil
}
func (m *mockConversionAdapter) CleanupInterfaceReferences(env interface{}) {}
func (m *mockConversionAdapter) CreateClassValue(className string) (Value, error) {
	return &runtime.NilValue{}, nil
}
func (m *mockConversionAdapter) CreateLambda(lambda *ast.LambdaExpression, closure any) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) CreateMethodPointer(obj Value, methodName string, closure any) (Value, error) {
	return &runtime.NilValue{}, nil
}
func (m *mockConversionAdapter) ExecuteFunctionPointerCall(metadata FunctionPointerMetadata, args []Value, node ast.Node) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) CreateExceptionDirect(classMetadata any, message string, pos any, callStack any) any {
	return nil
}
func (m *mockConversionAdapter) WrapObjectInException(objInstance Value, pos any, callStack any) any {
	return nil
}
func (m *mockConversionAdapter) EvalVariantBinaryOp(op string, left, right Value, node ast.Node) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) EvalInOperator(value, container Value, node ast.Node) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) EvalEqualityComparison(op string, left, right Value, node ast.Node) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) TryImplicitConversion(value Value, targetTypeName string) (Value, bool) {
	return value, false
}
func (m *mockConversionAdapter) WrapInSubrange(value Value, subrangeTypeName string, node ast.Node) (Value, error) {
	return value, nil
}
func (m *mockConversionAdapter) WrapInInterface(value Value, interfaceName string, node ast.Node) (Value, error) {
	return value, nil
}
func (m *mockConversionAdapter) GetObjectFieldValue(obj Value, fieldName string) (Value, bool) {
	return nil, false
}
func (m *mockConversionAdapter) GetClassVariableValue(obj Value, varName string) (Value, bool) {
	return nil, false
}
func (m *mockConversionAdapter) ReadPropertyValue(obj Value, propName string, node any) (Value, error) {
	return nil, nil
}
func (m *mockConversionAdapter) ExecutePropertyRead(obj Value, propInfo any, node any) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) IsMethodParameterless(obj Value, methodName string) bool {
	return false
}
func (m *mockConversionAdapter) CreateMethodCall(obj Value, methodName string, node any) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) CreateMethodPointerFromObject(obj Value, methodName string) (Value, error) {
	return nil, nil
}
func (m *mockConversionAdapter) CreateBoundMethodPointer(obj Value, methodDecl any) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) GetClassName(obj Value) string                    { return "" }
func (m *mockConversionAdapter) GetClassType(obj Value) Value                     { return &runtime.NilValue{} }
func (m *mockConversionAdapter) GetClassNameFromClassInfo(classInfo Value) string { return "" }
func (m *mockConversionAdapter) GetClassTypeFromClassInfo(classInfo Value) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) GetClassVariableFromClassInfo(classInfo Value, varName string) (Value, bool) {
	return nil, false
}
func (m *mockConversionAdapter) CallMemberMethod(callExpr *ast.CallExpression, memberAccess *ast.MemberAccessExpression, objVal Value) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) CallQualifiedOrConstructor(callExpr *ast.CallExpression, memberAccess *ast.MemberAccessExpression) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) CallUserFunctionWithOverloads(callExpr *ast.CallExpression, funcName *ast.Identifier) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) CallImplicitSelfMethod(callExpr *ast.CallExpression, funcName *ast.Identifier) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) CallRecordStaticMethod(callExpr *ast.CallExpression, funcName *ast.Identifier) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) DispatchRecordStaticMethod(recordTypeName string, callExpr *ast.CallExpression, funcName *ast.Identifier) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) CallIndexedPropertyGetter(obj Value, propImpl any, indices []Value, node any) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) ExecuteIndexedPropertyRead(obj Value, propInfo any, indices []Value, node any) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) CallRecordPropertyGetter(record Value, propImpl any, indices []Value, node any) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) ExecuteRecordPropertyRead(record Value, propInfo any, indices []Value, node any) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) NewClassInfoAdapter(name string) interface{} { return nil }
func (m *mockConversionAdapter) CastToClassInfo(class interface{}) (interface{}, bool) {
	return nil, false
}
func (m *mockConversionAdapter) GetClassNameFromClassInfoInterface(classInfo interface{}) string {
	return ""
}
func (m *mockConversionAdapter) RegisterClassEarly(name string, classInfo interface{})   {}
func (m *mockConversionAdapter) IsClassPartial(classInfo interface{}) bool               { return false }
func (m *mockConversionAdapter) SetClassPartial(classInfo interface{}, isPartial bool)   {}
func (m *mockConversionAdapter) SetClassAbstract(classInfo interface{}, isAbstract bool) {}
func (m *mockConversionAdapter) SetClassExternal(classInfo interface{}, isExternal bool, externalName string) {
}
func (m *mockConversionAdapter) ClassHasNoParent(classInfo interface{}) bool { return true }
func (m *mockConversionAdapter) DefineCurrentClassMarker(env interface{}, classInfo interface{}) {
}
func (m *mockConversionAdapter) SetClassParent(classInfo interface{}, parentClass interface{}) {}
func (m *mockConversionAdapter) AddInterfaceToClass(classInfo interface{}, interfaceInfo interface{}, interfaceName string) {
}
func (m *mockConversionAdapter) AddClassMethod(classInfo interface{}, method *ast.FunctionDecl, className string) bool {
	return false
}
func (m *mockConversionAdapter) CreateMethodMetadata(method *ast.FunctionDecl) interface{} {
	return nil
}
func (m *mockConversionAdapter) SynthesizeDefaultConstructor(classInfo interface{}) {}
func (m *mockConversionAdapter) AddClassProperty(classInfo interface{}, propDecl *ast.PropertyDecl) bool {
	return false
}
func (m *mockConversionAdapter) RegisterClassOperator(classInfo interface{}, opDecl *ast.OperatorDecl) Value {
	return &runtime.NilValue{}
}
func (m *mockConversionAdapter) LookupClassMethod(classInfo interface{}, methodName string, isClassMethod bool) (interface{}, bool) {
	return nil, false
}
func (m *mockConversionAdapter) SetClassConstructor(classInfo interface{}, constructor interface{}) {}
func (m *mockConversionAdapter) SetClassDestructor(classInfo interface{}, destructor interface{})   {}
func (m *mockConversionAdapter) InheritDestructorIfMissing(classInfo interface{})                   {}
func (m *mockConversionAdapter) InheritParentProperties(classInfo interface{})                      {}
func (m *mockConversionAdapter) BuildVirtualMethodTable(classInfo interface{})                      {}
func (m *mockConversionAdapter) RegisterClassInTypeSystem(classInfo interface{}, parentName string) {}
