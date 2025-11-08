package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// PropertyDecl Tests
// ============================================================================

func TestPropertyDeclBasic(t *testing.T) {
	t.Run("field-backed property", func(t *testing.T) {
		// property Name: String read FName write FName;
		prop := &PropertyDecl{
			Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Name"},
				Value: "Name",
			},
			Type: &TypeAnnotation{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "String"},
				Name:  "String",
			},
			ReadSpec: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "FName"},
				Value: "FName",
			},
			WriteSpec: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "FName"},
				Value: "FName",
			},
			IndexParams: nil,
			IsDefault:   false,
		}

		if prop.Name.Value != "Name" {
			t.Errorf("Expected Name='Name', got '%s'", prop.Name.Value)
		}
		if prop.Type.Name != "String" {
			t.Errorf("Expected Type='String', got '%s'", prop.Type.Name)
		}
		if prop.ReadSpec.(*Identifier).Value != "FName" {
			t.Errorf("Expected ReadSpec='FName', got '%s'", prop.ReadSpec.(*Identifier).Value)
		}
		if prop.WriteSpec.(*Identifier).Value != "FName" {
			t.Errorf("Expected WriteSpec='FName', got '%s'", prop.WriteSpec.(*Identifier).Value)
		}
	})

	t.Run("method-backed property", func(t *testing.T) {
		// property Count: Integer read GetCount write SetCount;
		prop := &PropertyDecl{
			Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Count"},
				Value: "Count",
			},
			Type: &TypeAnnotation{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
				Name:  "Integer",
			},
			ReadSpec: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "GetCount"},
				Value: "GetCount",
			},
			WriteSpec: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "SetCount"},
				Value: "SetCount",
			},
			IndexParams: nil,
			IsDefault:   false,
		}

		if prop.ReadSpec.(*Identifier).Value != "GetCount" {
			t.Errorf("Expected ReadSpec='GetCount', got '%s'", prop.ReadSpec.(*Identifier).Value)
		}
		if prop.WriteSpec.(*Identifier).Value != "SetCount" {
			t.Errorf("Expected WriteSpec='SetCount', got '%s'", prop.WriteSpec.(*Identifier).Value)
		}
	})
}

func TestPropertyDeclReadOnly(t *testing.T) {
	// property Size: Integer read FSize;
	prop := &PropertyDecl{
		Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
		Name: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "Size"},
			Value: "Size",
		},
		Type: &TypeAnnotation{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
			Name:  "Integer",
		},
		ReadSpec: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "FSize"},
			Value: "FSize",
		},
		WriteSpec:   nil, // Read-only: no write spec
		IndexParams: nil,
		IsDefault:   false,
	}

	if prop.ReadSpec == nil {
		t.Error("Read-only property should have ReadSpec")
	}
	if prop.WriteSpec != nil {
		t.Error("Read-only property should not have WriteSpec")
	}
}

func TestPropertyDeclWriteOnly(t *testing.T) {
	// property Output: String write SetOutput;
	prop := &PropertyDecl{
		Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
		Name: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "Output"},
			Value: "Output",
		},
		Type: &TypeAnnotation{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "String"},
			Name:  "String",
		},
		ReadSpec: nil, // Write-only: no read spec
		WriteSpec: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "SetOutput"},
			Value: "SetOutput",
		},
		IndexParams: nil,
		IsDefault:   false,
	}

	if prop.ReadSpec != nil {
		t.Error("Write-only property should not have ReadSpec")
	}
	if prop.WriteSpec == nil {
		t.Error("Write-only property should have WriteSpec")
	}
}

func TestPropertyDeclIndexed(t *testing.T) {
	t.Run("indexed property with single parameter", func(t *testing.T) {
		// property Items[index: Integer]: String read GetItem write SetItem;
		prop := &PropertyDecl{
			Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Items"},
				Value: "Items",
			},
			Type: &TypeAnnotation{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "String"},
				Name:  "String",
			},
			ReadSpec: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "GetItem"},
				Value: "GetItem",
			},
			WriteSpec: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "SetItem"},
				Value: "SetItem",
			},
			IndexParams: []*Parameter{
				{
					Name: &Identifier{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "index"},
						Value: "index",
					},
					Type: &TypeAnnotation{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
						Name:  "Integer",
					},
				},
			},
			IsDefault: false,
		}

		if prop.IndexParams == nil || len(prop.IndexParams) != 1 {
			t.Error("Indexed property should have 1 index parameter")
		}
		if prop.IndexParams[0].Name.Value != "index" {
			t.Errorf("Expected parameter name 'index', got '%s'", prop.IndexParams[0].Name.Value)
		}
	})

	t.Run("indexed property with multiple parameters", func(t *testing.T) {
		// property Data[x, y: Integer]: Float read GetData write SetData;
		prop := &PropertyDecl{
			Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Data"},
				Value: "Data",
			},
			Type: &TypeAnnotation{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Float"},
				Name:  "Float",
			},
			ReadSpec: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "GetData"},
				Value: "GetData",
			},
			WriteSpec: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "SetData"},
				Value: "SetData",
			},
			IndexParams: []*Parameter{
				{
					Name: &Identifier{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
						Value: "x",
					},
					Type: &TypeAnnotation{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
						Name:  "Integer",
					},
				},
				{
					Name: &Identifier{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "y"},
						Value: "y",
					},
					Type: &TypeAnnotation{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
						Name:  "Integer",
					},
				},
			},
			IsDefault: false,
		}

		if prop.IndexParams == nil || len(prop.IndexParams) != 2 {
			t.Errorf("Expected 2 index parameters, got %d", len(prop.IndexParams))
		}
	})
}

func TestPropertyDeclDefault(t *testing.T) {
	// property Items[index: Integer]: String read GetItem write SetItem; default;
	prop := &PropertyDecl{
		Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
		Name: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "Items"},
			Value: "Items",
		},
		Type: &TypeAnnotation{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "String"},
			Name:  "String",
		},
		ReadSpec: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "GetItem"},
			Value: "GetItem",
		},
		WriteSpec: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "SetItem"},
			Value: "SetItem",
		},
		IndexParams: []*Parameter{
			{
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "index"},
					Value: "index",
				},
				Type: &TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
					Name:  "Integer",
				},
			},
		},
		IsDefault: true,
	}

	if !prop.IsDefault {
		t.Error("Expected IsDefault=true")
	}
	if prop.IndexParams == nil {
		t.Error("Default property must be indexed")
	}
}

func TestPropertyDeclString(t *testing.T) {
	tests := []struct {
		name     string
		prop     *PropertyDecl
		expected string
	}{
		{
			name: "field-backed property",
			prop: &PropertyDecl{
				Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Name"},
					Value: "Name",
				},
				Type: &TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "String"},
					Name:  "String",
				},
				ReadSpec: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "FName"},
					Value: "FName",
				},
				WriteSpec: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "FName"},
					Value: "FName",
				},
			},
			expected: "property Name: String read FName write FName;",
		},
		{
			name: "read-only property",
			prop: &PropertyDecl{
				Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Size"},
					Value: "Size",
				},
				Type: &TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
					Name:  "Integer",
				},
				ReadSpec: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "FSize"},
					Value: "FSize",
				},
				WriteSpec: nil,
			},
			expected: "property Size: Integer read FSize;",
		},
		{
			name: "write-only property",
			prop: &PropertyDecl{
				Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Output"},
					Value: "Output",
				},
				Type: &TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "String"},
					Name:  "String",
				},
				ReadSpec: nil,
				WriteSpec: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "SetOutput"},
					Value: "SetOutput",
				},
			},
			expected: "property Output: String write SetOutput;",
		},
		{
			name: "indexed property",
			prop: &PropertyDecl{
				Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Items"},
					Value: "Items",
				},
				Type: &TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "String"},
					Name:  "String",
				},
				ReadSpec: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "GetItem"},
					Value: "GetItem",
				},
				WriteSpec: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "SetItem"},
					Value: "SetItem",
				},
				IndexParams: []*Parameter{
					{
						Name: &Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "index"},
							Value: "index",
						},
						Type: &TypeAnnotation{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
							Name:  "Integer",
						},
					},
				},
			},
			expected: "property Items[index: Integer]: String read GetItem write SetItem;",
		},
		{
			name: "default indexed property",
			prop: &PropertyDecl{
				Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Items"},
					Value: "Items",
				},
				Type: &TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "String"},
					Name:  "String",
				},
				ReadSpec: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "GetItem"},
					Value: "GetItem",
				},
				WriteSpec: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "SetItem"},
					Value: "SetItem",
				},
				IndexParams: []*Parameter{
					{
						Name: &Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "index"},
							Value: "index",
						},
						Type: &TypeAnnotation{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
							Name:  "Integer",
						},
					},
				},
				IsDefault: true,
			},
			expected: "property Items[index: Integer]: String read GetItem write SetItem; default;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.prop.String()
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestPropertyDeclTokenLiteral(t *testing.T) {
	prop := &PropertyDecl{
		Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
		Name: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "Name"},
			Value: "Name",
		},
		Type: &TypeAnnotation{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "String"},
			Name:  "String",
		},
		ReadSpec: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "FName"},
			Value: "FName",
		},
	}

	if prop.TokenLiteral() != "property" {
		t.Errorf("Expected TokenLiteral='property', got '%s'", prop.TokenLiteral())
	}
}

// ============================================================================
// Class Property Tests (Task 9.10)
// ============================================================================

func TestClassProperty(t *testing.T) {
	t.Run("basic class property", func(t *testing.T) {
		// class property Count: Integer read GetCount write SetCount;
		prop := &PropertyDecl{
			Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Count"},
				Value: "Count",
			},
			Type: &TypeAnnotation{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
				Name:  "Integer",
			},
			ReadSpec: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "GetCount"},
				Value: "GetCount",
			},
			WriteSpec: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "SetCount"},
				Value: "SetCount",
			},
			IsClassProperty: true,
		}

		if !prop.IsClassProperty {
			t.Error("Expected IsClassProperty=true")
		}

		expected := "class property Count: Integer read GetCount write SetCount;"
		result := prop.String()
		if result != expected {
			t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
		}
	})

	t.Run("class property read-only", func(t *testing.T) {
		// class property Version: String read GetVersion;
		prop := &PropertyDecl{
			Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Version"},
				Value: "Version",
			},
			Type: &TypeAnnotation{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "String"},
				Name:  "String",
			},
			ReadSpec: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "GetVersion"},
				Value: "GetVersion",
			},
			WriteSpec:       nil,
			IsClassProperty: true,
		}

		expected := "class property Version: String read GetVersion;"
		result := prop.String()
		if result != expected {
			t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
		}
	})

	t.Run("instance property (IsClassProperty=false)", func(t *testing.T) {
		// Regular instance property for comparison
		prop := &PropertyDecl{
			Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
			Name: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Name"},
				Value: "Name",
			},
			Type: &TypeAnnotation{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "String"},
				Name:  "String",
			},
			ReadSpec: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "FName"},
				Value: "FName",
			},
			WriteSpec: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "FName"},
				Value: "FName",
			},
			IsClassProperty: false,
		}

		if prop.IsClassProperty {
			t.Error("Expected IsClassProperty=false for instance property")
		}

		expected := "property Name: String read FName write FName;"
		result := prop.String()
		if result != expected {
			t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
		}
	})
}
