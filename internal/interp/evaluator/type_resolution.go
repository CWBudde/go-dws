package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// resolvedSemanticType returns an immutable analyzer-owned type when available.
func (e *Evaluator) resolvedSemanticType(node ast.Node) types.Type {
	if e.engineState == nil || e.SemanticInfo() == nil {
		return nil
	}
	return e.SemanticInfo().GetResolvedType(node)
}

// resolvedExpressionType reads the analyzer's identity, with a compatibility
// fallback for annotation-only callers.
func (e *Evaluator) resolvedExpressionType(node ast.Expression, ctx *ExecutionContext) types.Type {
	if resolved := e.resolvedSemanticType(node); resolved != nil {
		return resolved
	}
	if e.SemanticInfo() != nil {
		if annotation := e.SemanticInfo().GetType(node); annotation != nil {
			if resolved, err := e.ResolveTypeFromAnnotation(annotation, ctx); err == nil {
				return resolved
			}
		}
	}
	return nil
}

// resolvedExpressionTypeKind reads an explicit semantic annotation's type kind.
// Inferred callable types alone do not request a pointer: bare routines with
// default arguments must still be invoked when no contextual annotation exists.
func (e *Evaluator) resolvedExpressionTypeKind(node ast.Expression, ctx *ExecutionContext) string {
	if e.SemanticInfo() == nil || e.SemanticInfo().GetType(node) == nil {
		return ""
	}
	if resolved := e.resolvedExpressionType(node, ctx); resolved != nil {
		return resolved.TypeKind()
	}
	return ""
}

// resolveTypeReference preserves semantic identity for identifiers that name a
// type; unchecked callers fall back to the runtime's declared-name registry.
func (e *Evaluator) resolveTypeReference(node ast.Node, name string, ctx *ExecutionContext) (types.Type, error) {
	if resolved := e.resolvedSemanticType(node); resolved != nil {
		return resolved, nil
	}
	return e.resolveTypeName(name, ctx)
}

// ============================================================================
// Type Annotation Resolution
// ============================================================================

// ResolveTypeFromAnnotation resolves a type from an AST TypeExpression.
// This is used for function return types, parameter types, and variable declarations.
func (e *Evaluator) ResolveTypeFromAnnotation(typeExpr ast.TypeExpression, ctx *ExecutionContext) (types.Type, error) {
	if typeExpr == nil {
		return nil, nil
	}

	if e.engineState != nil && e.SemanticInfo() != nil {
		if resolved := e.SemanticInfo().GetResolvedType(typeExpr); resolved != nil {
			return resolved, nil
		}
	}

	switch node := typeExpr.(type) {
	case *ast.TypeAnnotation:
		if node.InlineType != nil {
			return e.ResolveTypeFromAnnotation(node.InlineType, ctx)
		}
		return e.resolveTypeName(node.String(), ctx)
	case *ast.SetTypeNode:
		element, err := e.ResolveTypeFromAnnotation(node.ElementType, ctx)
		if err != nil {
			return nil, fmt.Errorf("invalid set element type: %w", err)
		}
		if element == nil {
			return nil, fmt.Errorf("set element type is missing")
		}
		return types.NewSetType(element), nil
	case *ast.FunctionPointerTypeNode:
		return e.resolveFunctionPointerTypeNode(node, ctx)
	case *ast.ClassOfTypeNode:
		if node.ClassType == nil {
			return nil, fmt.Errorf("metaclass base type is missing")
		}
		base, err := e.ResolveTypeFromAnnotation(node.ClassType, ctx)
		if err != nil {
			return nil, err
		}
		class, ok := types.GetUnderlyingType(base).(*types.ClassType)
		if !ok {
			return nil, fmt.Errorf("invalid metaclass base type '%s'", node.ClassType.String())
		}
		return types.NewClassOfType(class), nil
	case *ast.RecordTypeNode:
		return e.resolveRecordTypeNode(node, ctx)
	case *ast.ArrayTypeAnnotation:
		return e.ResolveTypeFromAnnotation(&ast.ArrayTypeNode{
			ElementType: node.ElementType,
			LowBound:    node.LowBound,
			HighBound:   node.HighBound,
			Token:       node.Token,
		}, ctx)
	case *ast.ArrayTypeNode:
		if ctx == nil {
			ctx = &ExecutionContext{}
		}
		if assocType := e.resolveAssociativeArrayTypeNode(node, ctx); assocType != nil {
			return assocType, nil
		}
		arrayType := e.resolveArrayTypeNode(node, ctx)
		if arrayType == nil {
			return nil, fmt.Errorf("invalid array type")
		}
		return arrayType, nil
	}

	return nil, fmt.Errorf("unsupported type expression %T", typeExpr)
}

func (e *Evaluator) resolveFunctionPointerTypeNode(node *ast.FunctionPointerTypeNode, ctx *ExecutionContext) (types.Type, error) {
	params := make([]types.Type, len(node.Parameters))
	for index, param := range node.Parameters {
		var paramType types.Type = types.INTEGER
		if param.Type != nil {
			resolved, err := e.ResolveTypeFromAnnotation(param.Type, ctx)
			if err != nil {
				return nil, err
			}
			paramType = resolved
		}
		params[index] = paramType
	}
	result, err := e.ResolveTypeFromAnnotation(node.ReturnType, ctx)
	if err != nil {
		return nil, err
	}
	if node.OfObject {
		return types.NewMethodPointerType(params, result), nil
	}
	return types.NewFunctionPointerType(params, result), nil
}

func (e *Evaluator) resolveRecordTypeNode(recordNode *ast.RecordTypeNode, ctx *ExecutionContext) (types.Type, error) {
	recordType := types.NewRecordType("", make(map[string]types.Type))

	for _, field := range recordNode.Fields {
		fieldName := field.Name.Value
		fieldKey := ident.Normalize(fieldName)
		if _, exists := recordType.Fields[fieldKey]; exists {
			return nil, fmt.Errorf("field '%s' already exists in inline record", fieldName)
		}

		var fieldType types.Type
		if field.Type != nil {
			resolved, err := e.ResolveTypeFromAnnotation(field.Type, ctx)
			if err != nil {
				return nil, err
			}
			fieldType = resolved
		} else if field.InitValue != nil {
			value := e.Eval(field.InitValue, ctx)
			if isError(value) {
				return nil, fmt.Errorf("%s", value.String())
			}
			fieldType = e.getValueType(value)
		}
		if fieldType == nil {
			return nil, fmt.Errorf("field '%s' in inline record must have either a type or initializer", fieldName)
		}

		recordType.AddField(fieldName, fieldType, field.InitValue != nil)
	}

	for _, prop := range recordNode.Properties {
		propType, err := e.ResolveTypeFromAnnotation(prop.Type, ctx)
		if err != nil {
			return nil, err
		}
		propKey := ident.Normalize(prop.Name.Value)
		recordType.Properties[propKey] = &types.RecordPropertyInfo{
			Name:       prop.Name.Value,
			Type:       propType,
			ReadField:  prop.ReadField,
			WriteField: prop.WriteField,
			IsDefault:  prop.IsDefault,
			IsIndexed:  len(prop.IndexParams) > 0,

			IndexParamTypes: e.resolveRecordPropertyIndexParamTypes(prop.IndexParams, ctx),
		}
	}

	return recordType, nil
}

// ============================================================================
// Default Value Creation
// ============================================================================

// GetDefaultValue returns the default/zero value for a given type.
// This is used for Result variable initialization in functions.
func (e *Evaluator) GetDefaultValue(typ types.Type, ctx *ExecutionContext) Value {
	if typ == nil {
		return e.nilValue()
	}
	typ = types.GetUnderlyingType(typ)

	switch typ.TypeKind() {
	case "STRING":
		return &runtime.StringValue{Value: ""}
	case "INTEGER":
		return &runtime.IntegerValue{Value: 0}
	case "FLOAT":
		return &runtime.FloatValue{Value: 0.0}
	case "BOOLEAN":
		return &runtime.BooleanValue{Value: false}
	case "CLASS", "INTERFACE", "FUNCTION_POINTER", "METHOD_POINTER":
		return e.nilValue()
	case "CLASSOF":
		// An unassigned `class of X` variable is nil, but a nil metaclass is a
		// distinct mistake from a nil object and is reported as such.
		return &runtime.NilValue{IsMetaclass: true}
	case "BYTE_BUFFER":
		// ByteBuffer is auto-instantiated rather than nil-defaulted.
		return runtime.NewByteBufferValue()
	case "ARRAY":
		// Arrays should default to an empty array value of the correct element type.
		if arrType, ok := typ.(*types.ArrayType); ok {
			return runtime.NewArrayValue(arrType, nil)
		}
		return e.nilValue()
	case "RECORD", "SET", "ASSOCIATIVE_ARRAY":
		// Records and sets are value types and must be zero-initialized,
		// especially when dynamic arrays grow via SetLength and allocate new
		// elements. A set-typed Result must start as the empty set, not nil.
		return e.getZeroValueForType(typ, ctx)
	case "VARIANT":
		// Variants default to Unassigned (nil-like)
		return e.nilValue()
	default:
		// Unknown types default to NIL
		return e.nilValue()
	}
}

// nilValue returns a nil value.
func (e *Evaluator) nilValue() Value {
	return &runtime.NilValue{}
}

func (e *Evaluator) recordTypeFromAnnotation(annotation ast.TypeExpression, ctx *ExecutionContext) *types.RecordType {
	resolved, err := e.ResolveTypeFromAnnotation(annotation, ctx)
	if err != nil {
		return nil
	}
	if record, ok := types.GetUnderlyingType(resolved).(*types.RecordType); ok {
		return record
	}
	return nil
}

// resolveRecordPropertyIndexParamTypes resolves the declared index parameter
// types of a record property (`property Items[i : Integer] : String`).
//
// It returns nil when the property is not indexed or when any index parameter
// lacks a resolvable type annotation, matching the semantic analyzer so that
// runtime and compile-time record metadata agree.
func (e *Evaluator) resolveRecordPropertyIndexParamTypes(params []*ast.Parameter, ctx *ExecutionContext) []types.Type {
	if len(params) == 0 {
		return nil
	}
	resolved := make([]types.Type, 0, len(params))
	for _, param := range params {
		if param == nil || param.Type == nil {
			return nil
		}
		paramType, err := e.ResolveTypeFromAnnotation(param.Type, ctx)
		if err != nil || paramType == nil {
			return nil
		}
		resolved = append(resolved, paramType)
	}
	return resolved
}
