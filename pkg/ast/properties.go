package ast

import (
	"bytes"
	"strings"
)

// ============================================================================
// Property Declarations (Stage 8, Tasks 8.30-8.35)
// ============================================================================

// PropertyDecl represents a property declaration in a class.
// Properties provide syntactic sugar for getter/setter access.
//
// DWScript syntax examples:
//
//	property Name: String read FName write FName;                      // Field-backed
//	property Count: Integer read GetCount write SetCount;              // Method-backed
//	property Size: Integer read FSize;                                 // Read-only
//	property Items[index: Integer]: String read GetItem write SetItem; // Indexed
//	property Data[x, y: Integer]: Float read GetData;                  // Multi-index, read-only
//	property Items[i: Integer]: String read GetItem; default;          // Default property
//	class property Version: String read GetVersion;                    // Class property (static)
type PropertyDecl struct {
	BaseNode
	ReadSpec        Expression
	WriteSpec       Expression
	Name            *Identifier
	Type            *TypeAnnotation
	IndexParams     []*Parameter
	IsDefault       bool
	IsClassProperty bool
}

func (pd *PropertyDecl) statementNode() {}

// String returns the string representation of the property declaration.
func (pd *PropertyDecl) String() string {
	var out bytes.Buffer

	if pd.IsClassProperty {
		out.WriteString("class property ")
	} else {
		out.WriteString("property ")
	}
	out.WriteString(pd.Name.String())

	// Indexed property: property Items[index: Integer]
	if len(pd.IndexParams) > 0 {
		out.WriteString("[")
		params := make([]string, len(pd.IndexParams))
		for i, param := range pd.IndexParams {
			params[i] = param.Name.String() + ": " + param.Type.String()
		}
		out.WriteString(strings.Join(params, ", "))
		out.WriteString("]")
	}

	// Property type
	out.WriteString(": ")
	out.WriteString(pd.Type.String())

	// Read specifier
	if pd.ReadSpec != nil {
		out.WriteString(" read ")
		out.WriteString(pd.ReadSpec.String())
	}

	// Write specifier
	if pd.WriteSpec != nil {
		out.WriteString(" write ")
		out.WriteString(pd.WriteSpec.String())
	}

	out.WriteString(";")

	// Default keyword
	if pd.IsDefault {
		out.WriteString(" default;")
	}

	return out.String()
}
