package interp

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Helper Transfer from Semantic Analyzer to Runtime
// ============================================================================

// TransferHelpersFromSemanticAnalysis converts and registers helpers from the
// semantic analyzer's type system into the interpreter's runtime structures.
// This bridges the gap between compile-time type analysis and runtime execution.
//
// The function handles both user-defined helpers (with AST declarations) and
// built-in helpers (array, enum, intrinsic) which are registered separately.
func (i *Interpreter) TransferHelpersFromSemanticAnalysis(semanticHelpers map[string][]*types.HelperType) {
	if semanticHelpers == nil {
		return
	}

	// Initialize maps if needed
	if i.helpers == nil {
		i.helpers = make(map[string][]*HelperInfo)
	}

	// First pass: Convert all helpers (without parent references)
	// This ensures all helpers are registered before we try to link parents
	helperMap := make(map[string]*HelperInfo) // Map from helper name to runtime info

	for typeName, helperList := range semanticHelpers {
		for _, semanticHelper := range helperList {
			// Skip built-in helpers (array, enum, intrinsic)
			// These are registered by init*Helpers() methods
			if semanticHelper.Decl == nil {
				continue
			}

			// Convert to runtime HelperInfo (without parent resolution)
			runtimeHelper := convertHelperTypeToHelperInfoNoParent(semanticHelper)
			if runtimeHelper == nil {
				// Conversion failed (shouldn't happen in practice)
				continue
			}

			// Store in map for parent resolution
			helperMap[ident.Normalize(semanticHelper.Name)] = runtimeHelper

			// Register in legacy map (case-insensitive key)
			norm := ident.Normalize(typeName)
			i.helpers[norm] = append(i.helpers[norm], runtimeHelper)

			// Also register in TypeSystem for evaluator access
			if i.typeSystem != nil {
				i.typeSystem.RegisterHelper(typeName, runtimeHelper)
			}
		}
	}

	// Second pass: Resolve parent helper references
	for _, helperList := range semanticHelpers {
		for _, semanticHelper := range helperList {
			if semanticHelper.Decl == nil || semanticHelper.ParentHelper == nil {
				continue
			}

			// Find the runtime helper we just created
			runtimeHelper := helperMap[ident.Normalize(semanticHelper.Name)]
			if runtimeHelper == nil {
				continue
			}

			// Find the parent runtime helper
			parentName := ident.Normalize(semanticHelper.ParentHelper.Name)
			parentRuntime := helperMap[parentName]
			if parentRuntime != nil {
				runtimeHelper.ParentHelper = parentRuntime
			}
		}
	}
}

// convertHelperTypeToHelperInfoNoParent converts a semantic analyzer HelperType
// to an interpreter HelperInfo structure for runtime execution.
// Parent helper references are NOT resolved (must be done in a second pass).
//
// Returns nil if the helper has no AST declaration (built-in helper).
func convertHelperTypeToHelperInfoNoParent(semanticHelper *types.HelperType) *HelperInfo {
	if semanticHelper == nil {
		return nil
	}

	// Only convert user-defined helpers (with AST declarations)
	// Built-in helpers are registered separately via init*Helpers()
	decl, ok := semanticHelper.Decl.(*ast.HelperDecl)
	if !ok || decl == nil {
		return nil
	}

	// Create runtime HelperInfo
	runtimeHelper := NewHelperInfo(
		semanticHelper.Name,
		semanticHelper.TargetType,
		semanticHelper.IsRecordHelper,
	)

	// Parent helper references are resolved in a second pass
	// (after all helpers are converted)
	runtimeHelper.ParentHelper = nil

	// Transfer methods from AST declaration
	// The semantic analyzer validated these, so we just copy them
	for _, method := range decl.Methods {
		methodName := ident.Normalize(method.Name.Value)
		runtimeHelper.Methods[methodName] = method
	}

	// Transfer properties from types.HelperType
	// Properties are already in runtime format (types.PropertyInfo)
	for propName, propInfo := range semanticHelper.Properties {
		runtimeHelper.Properties[propName] = propInfo
	}

	// Transfer class variables - need to convert from Type to initial Value
	// For now, just store the type information
	// Actual values will be initialized during first use
	for varName := range semanticHelper.ClassVars {
		// Initialize class vars to nil/zero values
		// They'll be properly initialized when first accessed
		runtimeHelper.ClassVars[varName] = &NilValue{}
	}

	// Transfer class constants from AST declaration
	// Evaluate constant expressions at transfer time
	for _, constDecl := range decl.ClassConsts {
		constName := ident.Normalize(constDecl.Name.Value)
		// Store the constant value from semantic analysis
		if constValue, ok := semanticHelper.ClassConsts[constName]; ok {
			// Convert interface{} to Value
			runtimeHelper.ClassConsts[constName] = interfaceToValue(constValue)
		}
	}

	// Transfer built-in method mappings
	for methodName, builtinName := range semanticHelper.BuiltinMethods {
		runtimeHelper.BuiltinMethods[methodName] = builtinName
	}

	return runtimeHelper
}

// interfaceToValue converts a semantic constant value (interface{}) to a runtime Value.
func interfaceToValue(v interface{}) Value {
	if v == nil {
		return &NilValue{}
	}

	switch val := v.(type) {
	case int:
		return &IntegerValue{Value: int64(val)}
	case int64:
		return &IntegerValue{Value: val}
	case float64:
		return &FloatValue{Value: val}
	case string:
		return &StringValue{Value: val}
	case bool:
		return &BooleanValue{Value: val}
	default:
		// Unknown type - return nil
		return &NilValue{}
	}
}
