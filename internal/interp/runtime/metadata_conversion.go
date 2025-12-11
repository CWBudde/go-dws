package runtime

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// MethodMetadataFromAST converts an AST function declaration to MethodMetadata.
// This is used during the migration period to convert existing AST-based code
// to use the new metadata structures.
//
// Phase 9: Creates MethodMetadata with AST Body, PreConditions, PostConditions.
// Phase 10+: Will pre-compile Body/conditions to bytecode.
func MethodMetadataFromAST(fn *ast.FunctionDecl) *MethodMetadata {
	if fn == nil {
		return nil
	}

	metadata := &MethodMetadata{
		Name:       fn.Name.Value,
		Parameters: make([]ParameterMetadata, len(fn.Parameters)),
		Body:       fn.Body,
	}

	// Convert parameters
	for i, param := range fn.Parameters {
		metadata.Parameters[i] = ParameterMetadataFromAST(param)
	}

	// Set return type information
	if fn.ReturnType != nil {
		metadata.ReturnTypeName = fn.ReturnType.String()
		// ReturnType will be resolved during semantic analysis
	}

	// Copy validation conditions
	metadata.PreConditions = fn.PreConditions
	metadata.PostConditions = fn.PostConditions

	// Set method characteristics from AST flags
	metadata.IsVirtual = fn.IsVirtual
	metadata.IsAbstract = fn.IsAbstract
	metadata.IsOverride = fn.IsOverride
	metadata.IsReintroduce = fn.IsReintroduce
	metadata.IsClassMethod = fn.IsClassMethod
	metadata.IsConstructor = fn.IsConstructor
	metadata.IsDestructor = fn.IsDestructor

	return metadata
}

// ParameterMetadataFromAST converts an AST parameter to ParameterMetadata.
func ParameterMetadataFromAST(param *ast.Parameter) ParameterMetadata {
	metadata := ParameterMetadata{
		Name:         param.Name.Value,
		ByRef:        param.ByRef,
		DefaultValue: param.DefaultValue,
	}

	if param.Type != nil {
		metadata.TypeName = param.Type.String()
		// Type will be resolved during semantic analysis
	}

	return metadata
}

// FieldMetadataFromAST converts an AST field declaration to FieldMetadata.
func FieldMetadataFromAST(field *ast.FieldDecl) *FieldMetadata {
	if field == nil {
		return nil
	}

	metadata := &FieldMetadata{
		Name:      field.Name.Value,
		InitValue: field.InitValue,
	}

	if field.Type != nil {
		metadata.TypeName = field.Type.String()
		// Type will be resolved during semantic analysis
	}

	// Default visibility is public
	metadata.Visibility = FieldVisibilityPublic

	return metadata
}

// ClassMetadataFromAST creates ClassMetadata from a class declaration.
// This is a partial conversion - methods and fields are added separately
// as they are processed.
func ClassMetadataFromAST(decl *ast.ClassDecl) *ClassMetadata {
	if decl == nil {
		return nil
	}

	metadata := NewClassMetadata(decl.Name.Value)

	// Set parent name if specified
	if decl.Parent != nil {
		metadata.ParentName = decl.Parent.String()
	}

	// Set interfaces
	for _, intf := range decl.Interfaces {
		metadata.Interfaces = append(metadata.Interfaces, intf.String())
	}

	// Set class flags
	metadata.IsAbstract = decl.IsAbstract
	metadata.IsExternal = decl.IsExternal
	metadata.IsPartial = decl.IsPartial
	metadata.ExternalName = decl.ExternalName

	return metadata
}

// RecordMetadataFromAST creates RecordMetadata from a record declaration.
// This is a partial conversion - methods and fields are added separately.
func RecordMetadataFromAST(decl *ast.RecordDecl, recordType *types.RecordType) *RecordMetadata {
	if decl == nil {
		return nil
	}

	metadata := NewRecordMetadata(decl.Name.Value, recordType)

	return metadata
}

// AddMethodToClass adds a method to ClassMetadata, handling overloads.
func AddMethodToClass(class *ClassMetadata, method *MethodMetadata, isClassMethod bool) {
	if class == nil || method == nil {
		return
	}

	normalizedName := normalizeIdentifier(method.Name)

	if isClassMethod {
		// Class method (static)
		if existing, ok := class.ClassMethods[normalizedName]; ok {
			// Method exists - add to overloads
			if len(class.ClassMethodOverloads[normalizedName]) == 0 {
				// First overload - add existing as first overload
				class.ClassMethodOverloads[normalizedName] = append(
					class.ClassMethodOverloads[normalizedName],
					existing,
				)
			}
			class.ClassMethodOverloads[normalizedName] = append(
				class.ClassMethodOverloads[normalizedName],
				method,
			)
		} else {
			// First declaration
			class.ClassMethods[normalizedName] = method
		}
	} else {
		// Instance method
		if existing, ok := class.Methods[normalizedName]; ok {
			// Method exists - add to overloads
			if len(class.MethodOverloads[normalizedName]) == 0 {
				// First overload - add existing as first overload
				class.MethodOverloads[normalizedName] = append(
					class.MethodOverloads[normalizedName],
					existing,
				)
			}
			class.MethodOverloads[normalizedName] = append(
				class.MethodOverloads[normalizedName],
				method,
			)
		} else {
			// First declaration
			class.Methods[normalizedName] = method
		}
	}
}

// AddConstructorToClass adds a constructor to ClassMetadata, handling overloads.
func AddConstructorToClass(class *ClassMetadata, constructor *MethodMetadata) {
	if class == nil || constructor == nil {
		return
	}

	normalizedName := normalizeIdentifier(constructor.Name)
	constructor.IsConstructor = true

	if existing, ok := class.Constructors[normalizedName]; ok {
		// Constructor exists - add to overloads
		if len(class.ConstructorOverloads[normalizedName]) == 0 {
			// First overload - add existing as first overload
			class.ConstructorOverloads[normalizedName] = append(
				class.ConstructorOverloads[normalizedName],
				existing,
			)
		}
		class.ConstructorOverloads[normalizedName] = append(
			class.ConstructorOverloads[normalizedName],
			constructor,
		)
	} else {
		// First declaration
		class.Constructors[normalizedName] = constructor

		// Set as default constructor if none exists or if name is "Create"
		if class.DefaultConstructor == "" || normalizedName == "create" {
			class.DefaultConstructor = constructor.Name
		}
	}
}

// AddFieldToClass adds a field to ClassMetadata.
func AddFieldToClass(class *ClassMetadata, field *FieldMetadata) {
	if class == nil || field == nil {
		return
	}

	normalizedName := normalizeIdentifier(field.Name)
	class.Fields[normalizedName] = field
}

// AddMethodToRecord adds a method to RecordMetadata, handling overloads.
func AddMethodToRecord(record *RecordMetadata, method *MethodMetadata, isStatic bool) {
	if record == nil || method == nil {
		return
	}

	normalizedName := normalizeIdentifier(method.Name)

	if isStatic {
		// Static method
		if existing, ok := record.StaticMethods[normalizedName]; ok {
			// Method exists - add to overloads
			if len(record.StaticMethodOverloads[normalizedName]) == 0 {
				// First overload - add existing as first overload
				record.StaticMethodOverloads[normalizedName] = append(
					record.StaticMethodOverloads[normalizedName],
					existing,
				)
			}
			record.StaticMethodOverloads[normalizedName] = append(
				record.StaticMethodOverloads[normalizedName],
				method,
			)
		} else {
			// First declaration
			record.StaticMethods[normalizedName] = method
		}
	} else {
		// Instance method
		if existing, ok := record.Methods[normalizedName]; ok {
			// Method exists - add to overloads
			if len(record.MethodOverloads[normalizedName]) == 0 {
				// First overload - add existing as first overload
				record.MethodOverloads[normalizedName] = append(
					record.MethodOverloads[normalizedName],
					existing,
				)
			}
			record.MethodOverloads[normalizedName] = append(
				record.MethodOverloads[normalizedName],
				method,
			)
		} else {
			// First declaration
			record.Methods[normalizedName] = method
		}
	}
}

// AddFieldToRecord adds a field to RecordMetadata.
func AddFieldToRecord(record *RecordMetadata, field *FieldMetadata) {
	if record == nil || field == nil {
		return
	}

	normalizedName := normalizeIdentifier(field.Name)
	record.Fields[normalizedName] = field
}

// normalizeIdentifier converts an identifier to lowercase for case-insensitive lookup.
// This uses the same normalization as pkg/ident.Normalize.
//
// TODO: Once we can import pkg/ident without circular dependencies,
// replace this with ident.Normalize.
func normalizeIdentifier(name string) string {
	// Simple lowercase conversion for now
	// This matches the behavior of ident.Normalize
	result := make([]rune, 0, len(name))
	for _, r := range name {
		if r >= 'A' && r <= 'Z' {
			result = append(result, r+('a'-'A'))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
