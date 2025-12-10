package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// MockClassMetaValue implements ClassMetaValue for testing.
type MockClassMetaValue struct {
	runtime.NilValue // Embed for Value interface
	Name string
}

func (m *MockClassMetaValue) Type() string { return "CLASSINFO" }
func (m *MockClassMetaValue) String() string { return "class " + m.Name }
func (m *MockClassMetaValue) GetClassName() string { return m.Name }
func (m *MockClassMetaValue) GetClassType() Value { return nil }
func (m *MockClassMetaValue) GetClassVar(name string) (Value, bool) { return nil, false }
func (m *MockClassMetaValue) GetClassConstant(name string) (Value, bool) { return nil, false }
func (m *MockClassMetaValue) HasClassMethod(name string) bool { return false }
func (m *MockClassMetaValue) HasConstructor(name string) bool { return false }
func (m *MockClassMetaValue) InvokeParameterlessClassMethod(name string, executor func(methodDecl any) Value) (Value, bool) { return nil, false }
func (m *MockClassMetaValue) CreateClassMethodPointer(name string, creator func(methodDecl any) Value) (Value, bool) { return nil, false }
func (m *MockClassMetaValue) InvokeConstructor(name string, executor func(methodDecl any) Value) (Value, bool) { return nil, false }
func (m *MockClassMetaValue) GetNestedClass(name string) Value { return nil }
func (m *MockClassMetaValue) ReadClassProperty(name string, executor func(propInfo any) Value) (Value, bool) { return nil, false }
func (m *MockClassMetaValue) GetClassInfo() any { return nil }
func (m *MockClassMetaValue) SetClassVar(name string, value Value) bool { return false }
func (m *MockClassMetaValue) WriteClassProperty(name string, value Value, executor func(propInfo any, value Value) Value) (Value, bool) { return nil, false }
func (m *MockClassMetaValue) HasClassVar(name string) bool { return false }

// TestMemberAssignment_ErrorCases tests error handling for member assignment.
func TestMemberAssignment_ErrorCases(t *testing.T) {
	// Setup
	typeSystem := interptypes.NewTypeSystem()
	refCountMgr := runtime.NewRefCountManager()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil, refCountMgr)
	
	// Create mock adapter
	mockAdapter := &mockConversionAdapter{
		evalNodeFunc: func(node ast.Node) Value {
			// Check if it's the specific fallback case we want to catch
			if _, ok := node.(*ast.AssignmentStatement); ok {
				return e.newError(node, "fallback adapter called")
			}
			return &runtime.NilValue{}
		},
	}
	e.SetAdapter(mockAdapter)

	ctx := NewExecutionContext(runtime.NewEnvironment())

	t.Run("assignment to unsupported type", func(t *testing.T) {
		stmt := &ast.AssignmentStatement{
			Target: &ast.MemberAccessExpression{
				Object: &ast.IntegerLiteral{Value: 42},
				Member: &ast.Identifier{Value: "SomeField"},
			},
			Value: &ast.IntegerLiteral{Value: 10},
		}

		result := e.Eval(stmt, ctx)
		
		if !isError(result) {
			t.Errorf("Expected error value, got %T", result)
		} else {
			if errVal, ok := result.(*runtime.ErrorValue); ok {
				expected := "member assignment not supported for type INTEGER"
				if errVal.Message != expected {
					t.Errorf("Expected error '%s', got '%s'", expected, errVal.Message)
				}
			} else {
				t.Errorf("Expected runtime.ErrorValue, got %T", result)
			}
		}
	})

	t.Run("assignment to non-existent class member", func(t *testing.T) {
		// Mock a class access that returns a ClassMetaValue
		mockClass := &MockClassMetaValue{Name: "TMyClass"}
		e.DefineVar(ctx, "TMyClass", mockClass)

		stmt := &ast.AssignmentStatement{
			Target: &ast.MemberAccessExpression{
				Object: &ast.Identifier{Value: "TMyClass"},
				Member: &ast.Identifier{Value: "NonExistent"},
			},
			Value: &ast.IntegerLiteral{Value: 10},
		}

		result := e.Eval(stmt, ctx)
		
		if !isError(result) {
			t.Errorf("Expected error, got %v", result)
		} else {
			if errVal, ok := result.(*runtime.ErrorValue); ok {
				// We expect delegation to adapter for non-existent class members
				if errVal.Message == "fallback adapter called" {
					// Expected
				} else {
					t.Errorf("Unexpected error message: %s", errVal.Message)
				}
			}
		}
	})
}

// TestMemberAssignment_AutoInit tests nil record array element auto-initialization.
func TestMemberAssignment_AutoInit(t *testing.T) {
	// Setup
	typeSystem := interptypes.NewTypeSystem()
	refCountMgr := runtime.NewRefCountManager()
	e := NewEvaluator(typeSystem, nil, nil, nil, nil, refCountMgr)
	
	// Adapter shouldn't be called for native auto-init
	mockAdapter := &mockConversionAdapter{
		evalNodeFunc: func(node ast.Node) Value {
			t.Fatal("EvalNode should not be called for auto-initialization")
			return nil
		},
	}
	e.SetAdapter(mockAdapter)

	ctx := NewExecutionContext(runtime.NewEnvironment())

	t.Run("auto-initialization of record array element", func(t *testing.T) {
		// 1. Define Record Type
		recordType := &types.RecordType{
			Name: "TPoint",
			Fields: map[string]types.Type{
				"X": types.INTEGER,
				"Y": types.INTEGER,
			},
		}

		// 2. Define Array Type
		arrayType := types.NewDynamicArrayType(recordType)

		// 3. Create Array Instance with 1 element (nil)
		arrayVal := &runtime.ArrayValue{
			ArrayType: arrayType,
			Elements:  []runtime.Value{nil}, 
		}

		e.DefineVar(ctx, "points", arrayVal)

		// 4. Construct Assignment: points[0].X := 10
		stmt := &ast.AssignmentStatement{
			Target: &ast.MemberAccessExpression{
				Object: &ast.IndexExpression{
					Left:  &ast.Identifier{Value: "points"},
					Index: &ast.IntegerLiteral{Value: 0},
				},
				Member: &ast.Identifier{Value: "X"},
			},
			Value: &ast.IntegerLiteral{Value: 10},
		}

		// 5. Eval
		result := e.Eval(stmt, ctx)

		if isError(result) {
			t.Fatalf("Unexpected error: %v", result)
		}

		// 6. Verify
		// points[0] should be a RecordValue
		elem := arrayVal.Elements[0]
		if elem == nil {
            t.Logf("Array address: %p", arrayVal)
            t.Logf("Elements: %v", arrayVal.Elements)
			t.Fatal("points[0] is still nil")
		}
		
		recVal, ok := elem.(*runtime.RecordValue)
		if !ok {
			t.Fatalf("points[0] is not a RecordValue, got %T", elem)
		}
		
		// Check X field
		xVal, found := recVal.GetRecordField("X")
		if !found {
			t.Fatal("Field X not found in new record")
		}
		
		if intVal, ok := xVal.(*runtime.IntegerValue); ok {
			if intVal.Value != 10 {
				t.Errorf("Expected X=10, got %d", intVal.Value)
			}
		} else {
			t.Errorf("Field X is not Integer, got %T", xVal)
		}
	})
}