package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// This file contains visitor methods for declaration AST nodes.
// Phase 3.5.2: Visitor pattern implementation for declarations.
//
// Declarations define types, functions, classes, etc. and register them
// in the appropriate registries.

// VisitFunctionDecl evaluates a function declaration.
func (e *Evaluator) VisitFunctionDecl(node *ast.FunctionDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Function registry available via adapter.LookupFunction()
	// TODO: Move function registration logic here (use adapter or typeSystem.FunctionRegistry)
	return e.adapter.EvalNode(node)
}

// VisitClassDecl evaluates a class declaration.
func (e *Evaluator) VisitClassDecl(node *ast.ClassDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Class registry available via adapter.LookupClass()
	// TODO: Move class registration logic here (use adapter type system methods)
	return e.adapter.EvalNode(node)
}

// VisitInterfaceDecl evaluates an interface declaration.
func (e *Evaluator) VisitInterfaceDecl(node *ast.InterfaceDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Interface registry available via adapter.LookupInterface()
	// TODO: Move interface registration logic here (use adapter type system methods)
	return e.adapter.EvalNode(node)
}

// VisitOperatorDecl evaluates an operator declaration (operator overloading).
// Task 3.5.14: Migrated from Interpreter.evalOperatorDeclaration to Evaluator visitor.
//
// This method handles global and conversion operator declarations:
//   - Global operators: registered via TypeSystem.Operators()
//   - Conversion operators: registered via TypeSystem.Conversions()
//
// Class operators (Kind == OperatorKindClass) are handled separately during
// class declaration evaluation, so this method returns NilValue for them.
func (e *Evaluator) VisitOperatorDecl(node *ast.OperatorDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil operator declaration")
	}

	// Class operators are registered during class declaration evaluation
	if node.Kind == ast.OperatorKindClass {
		return &runtime.NilValue{}
	}

	// Validate binding exists
	if node.Binding == nil {
		return e.newError(node, "operator '%s' missing binding", node.OperatorSymbol)
	}

	// Normalize operand types for consistent lookup
	operandTypes := make([]string, len(node.OperandTypes))
	for idx, operand := range node.OperandTypes {
		opRand := operand.String()
		operandTypes[idx] = interptypes.NormalizeTypeAnnotation(opRand)
	}

	// Handle conversion operators
	if node.Kind == ast.OperatorKindConversion {
		if len(operandTypes) != 1 {
			return e.newError(node, "conversion operator '%s' requires exactly one operand", node.OperatorSymbol)
		}
		if node.ReturnType == nil {
			return e.newError(node, "conversion operator '%s' requires a return type", node.OperatorSymbol)
		}

		targetType := interptypes.NormalizeTypeAnnotation(node.ReturnType.String())
		entry := &interptypes.ConversionEntry{
			From:        operandTypes[0],
			To:          targetType,
			BindingName: ident.Normalize(node.Binding.Value),
			Implicit:    ident.Equal(node.OperatorSymbol, "implicit"),
		}

		if err := e.typeSystem.Conversions().Register(entry); err != nil {
			return e.newError(node, "conversion from %s to %s already defined", operandTypes[0], targetType)
		}
		return &runtime.NilValue{}
	}

	// Handle global operators
	entry := &interptypes.OperatorEntry{
		Operator:     node.OperatorSymbol,
		OperandTypes: operandTypes,
		BindingName:  ident.Normalize(node.Binding.Value),
	}

	if err := e.typeSystem.Operators().Register(entry); err != nil {
		return e.newError(node, "operator '%s' already defined for operand types (%s)", node.OperatorSymbol, strings.Join(operandTypes, ", "))
	}

	return &runtime.NilValue{}
}

// VisitEnumDecl evaluates an enum declaration.
// Task 3.5.11: Migrated from Interpreter.evalEnumDeclaration to Evaluator visitor.
//
// This method:
//  1. Builds the enum type from the AST declaration
//  2. Calculates ordinal values (explicit or implicit)
//  3. For flags enums, validates values are powers of 2
//  4. For unscoped enums, registers values in the environment
//  5. Registers enum type metadata in the TypeSystem
//  6. Creates a TypeMetaValue for the enum type name
func (e *Evaluator) VisitEnumDecl(node *ast.EnumDecl, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil enum declaration")
	}

	enumName := node.Name.Value

	// Build the enum type from the declaration
	enumValues := make(map[string]int)
	orderedNames := make([]string, 0, len(node.Values))

	// Calculate ordinal values (explicit or implicit)
	currentOrdinal := 0
	flagBitPosition := 0 // For flags enums, track the bit position (2^n)

	for _, enumValue := range node.Values {
		valueName := enumValue.Name

		// Determine ordinal value (explicit or implicit)
		var ordinalValue int
		if enumValue.Value != nil {
			// Explicit value provided
			ordinalValue = *enumValue.Value
			if node.Flags {
				// For flags, explicit values must be powers of 2
				if ordinalValue <= 0 || (ordinalValue&(ordinalValue-1)) != 0 {
					return e.newError(node, "enum '%s' value '%s' (%d) must be a power of 2 for flags enum",
						enumName, valueName, ordinalValue)
				}
				// For flags, update bit position based on explicit value
				for bitPos := 0; bitPos < 64; bitPos++ {
					if (1 << bitPos) == ordinalValue {
						flagBitPosition = bitPos + 1
						break
					}
				}
			} else {
				// For regular enums, update current ordinal
				currentOrdinal = ordinalValue + 1
			}
		} else {
			// Implicit value
			if node.Flags {
				// Flags use power-of-2 values: 1, 2, 4, 8, 16, ...
				ordinalValue = 1 << flagBitPosition
				flagBitPosition++
			} else {
				// Regular enums use sequential values
				ordinalValue = currentOrdinal
				currentOrdinal++
			}
		}

		// Store the value
		enumValues[valueName] = ordinalValue
		orderedNames = append(orderedNames, valueName)
	}

	// Create the enum type
	var enumType *types.EnumType
	if node.Scoped || node.Flags {
		enumType = types.NewScopedEnumType(enumName, enumValues, orderedNames, node.Flags)
	} else {
		enumType = types.NewEnumType(enumName, enumValues, orderedNames)
	}

	// Register each enum value in the symbol table as a constant
	// For scoped enums (enum/flags keyword), skip global registration -
	// values are only accessible via qualified access (Type.Value)
	if !node.Scoped {
		for valueName, ordinalValue := range enumValues {
			enumVal := &runtime.EnumValue{
				TypeName:     enumName,
				ValueName:    valueName,
				OrdinalValue: ordinalValue,
			}
			ctx.Env().Define(valueName, enumVal)
		}
	}

	// Store enum type metadata in environment with special key
	// This allows variable declarations to resolve the type
	enumTypeKey := "__enum_type_" + ident.Normalize(enumName)
	enumTypeValue := runtime.NewEnumTypeValue(enumType)
	ctx.Env().Define(enumTypeKey, enumTypeValue)

	// Register in TypeSystem for consistent type lookups
	e.typeSystem.RegisterEnumType(enumName, enumTypeValue)

	// Register enum type name as a TypeMetaValue
	// This allows the type name to be used as a runtime value in expressions
	// like High(TColor) or Low(TColor), just like built-in types (Integer, Float, etc.)
	typeMetaValue := &runtime.TypeMetaValue{
		TypeInfo: enumType,
		TypeName: enumName,
	}
	ctx.Env().Define(enumName, typeMetaValue)

	return &runtime.NilValue{}
}

// VisitRecordDecl evaluates a record declaration.
func (e *Evaluator) VisitRecordDecl(node *ast.RecordDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Record registry available via adapter.LookupRecord()
	// TODO: Move record registration logic here (use adapter type system methods)
	return e.adapter.EvalNode(node)
}

// VisitHelperDecl evaluates a helper declaration (type extension).
func (e *Evaluator) VisitHelperDecl(node *ast.HelperDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Helper registry available via adapter.LookupHelpers()
	// TODO: Move helper registration logic here (use adapter type system methods)
	return e.adapter.EvalNode(node)
}

// VisitArrayDecl evaluates an array type declaration.
func (e *Evaluator) VisitArrayDecl(node *ast.ArrayDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Type system available for array type registration
	// TODO: Move array type registration logic here
	return e.adapter.EvalNode(node)
}

// VisitTypeDeclaration evaluates a type alias declaration.
func (e *Evaluator) VisitTypeDeclaration(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Type system available for type alias handling
	// TODO: Move type alias registration logic here
	return e.adapter.EvalNode(node)
}

// VisitSetDecl evaluates a set declaration.
func (e *Evaluator) VisitSetDecl(node *ast.SetDecl, ctx *ExecutionContext) Value {
	// Set type already registered by semantic analyzer
	// Delegate to adapter for now (Phase 3 migration)
	return e.adapter.EvalNode(node)
}
