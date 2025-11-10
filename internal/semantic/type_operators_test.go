package semantic

import (
	"testing"
)

// ============================================================================
// Type Operator Tests (is, as, implements)
// ============================================================================
// These tests cover type checking and casting operators to improve
// coverage of analyze_expressions.go (currently at 0% for these operators)

// 'is' operator tests
func TestTypeOperator_Is_ClassInheritance(t *testing.T) {
	input := `
		type TBase = class
		end;

		type TChild = class(TBase)
		end;

		var obj: TBase := TChild.Create();
		var isChild := obj is TChild;
	`
	expectNoErrors(t, input)
}

func TestTypeOperator_Is_DirectType(t *testing.T) {
	input := `
		type TMyClass = class
		end;

		var obj := TMyClass.Create();
		var isMyClass := obj is TMyClass;
	`
	expectNoErrors(t, input)
}

func TestTypeOperator_Is_NilObject(t *testing.T) {
	input := `
		type TBase = class
		end;

		var obj: TBase := nil;
		var isBase := obj is TBase;
	`
	expectNoErrors(t, input)
}

func TestTypeOperator_Is_ParentType(t *testing.T) {
	input := `
		type TBase = class
		end;

		type TChild = class(TBase)
		end;

		var child := TChild.Create();
		var isBase := child is TBase;
	`
	expectNoErrors(t, input)
}

func TestTypeOperator_Is_UnrelatedClass(t *testing.T) {
	input := `
		type TClassA = class
		end;

		type TClassB = class
		end;

		var objA := TClassA.Create();
		var isB := objA is TClassB;
	`
	expectNoErrors(t, input)
}

func TestTypeOperator_Is_InCondition(t *testing.T) {
	input := `
		type TBase = class
		end;

		type TChild = class(TBase)
		end;

		var obj: TBase := TChild.Create();
		if obj is TChild then
			PrintLn('It is a child');
	`
	expectNoErrors(t, input)
}

func TestTypeOperator_Is_InvalidLeftOperand(t *testing.T) {
	input := `
		type TBase = class
		end;

		var n := 42;
		var result := n is TBase;
	`
	expectError(t, input, "class")
}

func TestTypeOperator_Is_InvalidRightOperand(t *testing.T) {
	input := `
		type TBase = class
		end;

		var obj := TBase.Create();
		var result := obj is Integer;
	`
	expectError(t, input, "class type")
}

// 'as' operator tests
func TestTypeOperator_As_DowncastValid(t *testing.T) {
	input := `
		type TBase = class
		end;

		type TChild = class(TBase)
			procedure DoSomething();
		end;

		procedure TChild.DoSomething();
		begin
		end;

		var obj: TBase := TChild.Create();
		var child := obj as TChild;
		child.DoSomething();
	`
	expectNoErrors(t, input)
}

func TestTypeOperator_As_UpcastValid(t *testing.T) {
	input := `
		type TBase = class
		end;

		type TChild = class(TBase)
		end;

		var child := TChild.Create();
		var base := child as TBase;
	`
	expectNoErrors(t, input)
}

func TestTypeOperator_As_WithCheck(t *testing.T) {
	input := `
		type TBase = class
		end;

		type TChild = class(TBase)
			procedure DoSomething();
		end;

		procedure TChild.DoSomething();
		begin
		end;

		var obj: TBase := TChild.Create();
		if obj is TChild then
		begin
			var child := obj as TChild;
			child.DoSomething();
		end;
	`
	expectNoErrors(t, input)
}

func TestTypeOperator_As_InvalidLeftOperand(t *testing.T) {
	input := `
		type TBase = class
		end;

		var n := 42;
		var obj := n as TBase;
	`
	expectError(t, input, "class")
}

func TestTypeOperator_As_InvalidRightOperand(t *testing.T) {
	input := `
		type TBase = class
		end;

		var obj := TBase.Create();
		var result := obj as Integer;
	`
	expectError(t, input, "class or interface")
}

func TestTypeOperator_As_NilValue(t *testing.T) {
	input := `
		type TBase = class
		end;

		type TChild = class(TBase)
		end;

		var obj: TBase := nil;
		var child := obj as TChild;
	`
	expectNoErrors(t, input)
}

// 'implements' operator tests
func TestTypeOperator_Implements_ClassImplementsInterface(t *testing.T) {
	input := `
		type IMyInterface = interface
			procedure DoSomething();
		end;

		type TMyClass = class(TObject, IMyInterface)
			procedure DoSomething();
		end;

		procedure TMyClass.DoSomething();
		begin
		end;

		var obj := TMyClass.Create();
		var doesImplement := obj implements IMyInterface;
	`
	expectNoErrors(t, input)
}

func TestTypeOperator_Implements_ClassDoesNotImplement(t *testing.T) {
	input := `
		type IMyInterface = interface
			procedure DoSomething();
		end;

		type TMyClass = class
		end;

		var obj := TMyClass.Create();
		var doesImplement := obj implements IMyInterface;
	`
	expectNoErrors(t, input)
}

func TestTypeOperator_Implements_InCondition(t *testing.T) {
	input := `
		type IMyInterface = interface
			procedure DoSomething();
		end;

		type TMyClass = class(TObject, IMyInterface)
			procedure DoSomething();
		end;

		procedure TMyClass.DoSomething();
		begin
		end;

		var obj: TObject := TMyClass.Create();
		if obj implements IMyInterface then
			PrintLn('Object implements interface');
	`
	expectNoErrors(t, input)
}

func TestTypeOperator_Implements_InvalidLeftOperand(t *testing.T) {
	input := `
		type IMyInterface = interface
			procedure DoSomething();
		end;

		var n := 42;
		var result := n implements IMyInterface;
	`
	expectError(t, input, "class")
}

func TestTypeOperator_Implements_InvalidRightOperand(t *testing.T) {
	input := `
		type TBase = class
		end;

		var obj := TBase.Create();
		var result := obj implements TBase;
	`
	expectError(t, input, "interface")
}

func TestTypeOperator_Implements_NilObject(t *testing.T) {
	input := `
		type IMyInterface = interface
			procedure DoSomething();
		end;

		type TMyClass = class(TObject, IMyInterface)
			procedure DoSomething();
		end;

		procedure TMyClass.DoSomething();
		begin
		end;

		var obj: TMyClass := nil;
		var doesImplement := obj implements IMyInterface;
	`
	expectNoErrors(t, input)
}

// Combined type operator tests
func TestTypeOperators_IsAndAs_Together(t *testing.T) {
	input := `
		type TBase = class
		end;

		type TChild = class(TBase)
			procedure Special();
		end;

		procedure TChild.Special();
		begin
			PrintLn('Special method');
		end;

		var obj: TBase := TChild.Create();
		if obj is TChild then
		begin
			var child := obj as TChild;
			child.Special();
		end;
	`
	expectNoErrors(t, input)
}

func TestTypeOperators_AllThreeTogether(t *testing.T) {
	input := `
		type IMyInterface = interface
			procedure DoIt();
		end;

		type TBase = class
		end;

		type TChild = class(TBase, IMyInterface)
			procedure DoIt();
		end;

		procedure TChild.DoIt();
		begin
		end;

		var obj: TBase := TChild.Create();
		if obj is TChild then
		begin
			if obj implements IMyInterface then
			begin
				var child := obj as TChild;
				child.DoIt();
			end;
		end;
	`
	expectNoErrors(t, input)
}

func TestTypeOperators_InFunction(t *testing.T) {
	input := `
		type TBase = class
		end;

		type TChild = class(TBase)
		end;

		function IsChildType(obj: TBase): Boolean;
		begin
			Result := obj is TChild;
		end;

		var obj := TChild.Create();
		var result := IsChildType(obj);
	`
	expectNoErrors(t, input)
}

func TestTypeOperators_WithPolymorphism(t *testing.T) {
	input := `
		type TAnimal = class
			procedure MakeSound(); virtual;
		end;

		type TDog = class(TAnimal)
			procedure MakeSound(); override;
			procedure Bark();
		end;

		procedure TAnimal.MakeSound();
		begin
			PrintLn('Animal sound');
		end;

		procedure TDog.MakeSound();
		begin
			PrintLn('Woof!');
		end;

		procedure TDog.Bark();
		begin
			PrintLn('Bark bark!');
		end;

		var animal: TAnimal := TDog.Create();
		if animal is TDog then
		begin
			var dog := animal as TDog;
			dog.Bark();
		end;
	`
	expectNoErrors(t, input)
}

// Edge cases
func TestTypeOperators_ChainedIs(t *testing.T) {
	input := `
		type TBase = class
		end;

		type TChild1 = class(TBase)
		end;

		type TChild2 = class(TBase)
		end;

		var obj: TBase := TChild1.Create();
		var isChild1 := obj is TChild1;
		var isChild2 := obj is TChild2;
	`
	expectNoErrors(t, input)
}

func TestTypeOperators_AsWithoutCheck(t *testing.T) {
	// Should analyze without error (runtime check for invalid cast)
	input := `
		type TBase = class
		end;

		type TChild = class(TBase)
		end;

		var obj: TBase := TBase.Create();
		var child := obj as TChild;
	`
	expectNoErrors(t, input)
}

func TestTypeOperators_InArrayIteration(t *testing.T) {
	input := `
		type TBase = class
		end;

		type TChild = class(TBase)
		end;

		var objects: array of TBase;
		SetLength(objects, 3);

		for i := 0 to 2 do
		begin
			if objects[i] is TChild then
				PrintLn('Found a child');
		end;
	`
	expectNoErrors(t, input)
}

func TestTypeOperators_MultipleInterfaces(t *testing.T) {
	input := `
		type IInterface1 = interface
			procedure Method1();
		end;

		type IInterface2 = interface
			procedure Method2();
		end;

		type TMyClass = class(TObject, IInterface1, IInterface2)
			procedure Method1();
			procedure Method2();
		end;

		procedure TMyClass.Method1();
		begin
		end;

		procedure TMyClass.Method2();
		begin
		end;

		var obj := TMyClass.Create();
		var impl1 := obj implements IInterface1;
		var impl2 := obj implements IInterface2;
	`
	expectNoErrors(t, input)
}

func TestTypeOperators_InCase(t *testing.T) {
	input := `
		type TBase = class
		end;

		type TChild1 = class(TBase)
		end;

		type TChild2 = class(TBase)
		end;

		var obj: TBase;
		if obj is TChild1 then
			PrintLn('Child1')
		else if obj is TChild2 then
			PrintLn('Child2')
		else
			PrintLn('Base');
	`
	expectNoErrors(t, input)
}
