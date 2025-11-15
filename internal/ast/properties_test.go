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
			BaseNode: BaseNode{
				Token: NewTestToken(lexer.PROPERTY, "property"),
			},
			Name:        NewTestIdentifier("Name"),
			Type:        NewTestTypeAnnotation("String"),
			ReadSpec:    NewTestIdentifier("FName"),
			WriteSpec:   NewTestIdentifier("FName"),
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
			BaseNode: BaseNode{
				Token: NewTestToken(lexer.PROPERTY, "property"),
			},
			Name:        NewTestIdentifier("Count"),
			Type:        NewTestTypeAnnotation("Integer"),
			ReadSpec:    NewTestIdentifier("GetCount"),
			WriteSpec:   NewTestIdentifier("SetCount"),
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
		BaseNode: BaseNode{
			Token: NewTestToken(lexer.PROPERTY, "property"),
		},
		Name:        NewTestIdentifier("Size"),
		Type:        NewTestTypeAnnotation("Integer"),
		ReadSpec:    NewTestIdentifier("FSize"),
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
		BaseNode: BaseNode{
			Token: NewTestToken(lexer.PROPERTY, "property"),
		},
		Name:        NewTestIdentifier("Output"),
		Type:        NewTestTypeAnnotation("String"),
		ReadSpec:    nil, // Write-only: no read spec
		WriteSpec:   NewTestIdentifier("SetOutput"),
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
			BaseNode: BaseNode{
				Token: NewTestToken(lexer.PROPERTY, "property"),
			},
			Name:      NewTestIdentifier("Items"),
			Type:      NewTestTypeAnnotation("String"),
			ReadSpec:  NewTestIdentifier("GetItem"),
			WriteSpec: NewTestIdentifier("SetItem"),
			IndexParams: []*Parameter{
				{
					Name: NewTestIdentifier("index"),
					Type: NewTestTypeAnnotation("Integer"),
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
			BaseNode: BaseNode{
				Token: NewTestToken(lexer.PROPERTY, "property"),
			},
			Name:      NewTestIdentifier("Data"),
			Type:      NewTestTypeAnnotation("Float"),
			ReadSpec:  NewTestIdentifier("GetData"),
			WriteSpec: NewTestIdentifier("SetData"),
			IndexParams: []*Parameter{
				{
					Name: NewTestIdentifier("x"),
					Type: NewTestTypeAnnotation("Integer"),
				},
				{
					Name: NewTestIdentifier("y"),
					Type: NewTestTypeAnnotation("Integer"),
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
		BaseNode: BaseNode{
			Token: NewTestToken(lexer.PROPERTY, "property"),
		},
		Name:      NewTestIdentifier("Items"),
		Type:      NewTestTypeAnnotation("String"),
		ReadSpec:  NewTestIdentifier("GetItem"),
		WriteSpec: NewTestIdentifier("SetItem"),
		IndexParams: []*Parameter{
			{
				Name: NewTestIdentifier("index"),
				Type: NewTestTypeAnnotation("Integer"),
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
				BaseNode: BaseNode{
					Token: NewTestToken(lexer.PROPERTY, "property"),
				},
				Name:      NewTestIdentifier("Name"),
				Type:      NewTestTypeAnnotation("String"),
				ReadSpec:  NewTestIdentifier("FName"),
				WriteSpec: NewTestIdentifier("FName"),
			},
			expected: "property Name: String read FName write FName;",
		},
		{
			name: "read-only property",
			prop: &PropertyDecl{
				BaseNode: BaseNode{
					Token: NewTestToken(lexer.PROPERTY, "property"),
				},
				Name:     NewTestIdentifier("Size"),
				Type:     NewTestTypeAnnotation("Integer"),
				ReadSpec: NewTestIdentifier("FSize"),
			},
			expected: "property Size: Integer read FSize;",
		},
		{
			name: "write-only property",
			prop: &PropertyDecl{
				BaseNode: BaseNode{
					Token: NewTestToken(lexer.PROPERTY, "property"),
				},
				Name:      NewTestIdentifier("Output"),
				Type:      NewTestTypeAnnotation("String"),
				WriteSpec: NewTestIdentifier("SetOutput"),
			},
			expected: "property Output: String write SetOutput;",
		},
		{
			name: "indexed property",
			prop: &PropertyDecl{
				BaseNode: BaseNode{
					Token: NewTestToken(lexer.PROPERTY, "property"),
				},
				Name:      NewTestIdentifier("Items"),
				Type:      NewTestTypeAnnotation("String"),
				ReadSpec:  NewTestIdentifier("GetItem"),
				WriteSpec: NewTestIdentifier("SetItem"),
				IndexParams: []*Parameter{
					{
						Name: NewTestIdentifier("index"),
						Type: NewTestTypeAnnotation("Integer"),
					},
				},
			},
			expected: "property Items[index: Integer]: String read GetItem write SetItem;",
		},
		{
			name: "default indexed property",
			prop: &PropertyDecl{
				BaseNode: BaseNode{
					Token: NewTestToken(lexer.PROPERTY, "property"),
				},
				Name:      NewTestIdentifier("Items"),
				Type:      NewTestTypeAnnotation("String"),
				ReadSpec:  NewTestIdentifier("GetItem"),
				WriteSpec: NewTestIdentifier("SetItem"),
				IndexParams: []*Parameter{
					{
						Name: NewTestIdentifier("index"),
						Type: NewTestTypeAnnotation("Integer"),
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
		BaseNode: BaseNode{
			Token: NewTestToken(lexer.PROPERTY, "property"),
		},
		Name:     NewTestIdentifier("Name"),
		Type:     NewTestTypeAnnotation("String"),
		ReadSpec: NewTestIdentifier("FName"),
	}

	if prop.TokenLiteral() != "property" {
		t.Errorf("Expected TokenLiteral='property', got '%s'", prop.TokenLiteral())
	}
}

// ============================================================================
// Class Property Tests
// ============================================================================

func TestClassProperty(t *testing.T) {
	t.Run("basic class property", func(t *testing.T) {
		// class property Count: Integer read GetCount write SetCount;
		prop := &PropertyDecl{
			BaseNode: BaseNode{
				Token: NewTestToken(lexer.PROPERTY, "property"),
			},
			Name:            NewTestIdentifier("Count"),
			Type:            NewTestTypeAnnotation("Integer"),
			ReadSpec:        NewTestIdentifier("GetCount"),
			WriteSpec:       NewTestIdentifier("SetCount"),
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
			BaseNode: BaseNode{
				Token: NewTestToken(lexer.PROPERTY, "property"),
			},
			Name:            NewTestIdentifier("Version"),
			Type:            NewTestTypeAnnotation("String"),
			ReadSpec:        NewTestIdentifier("GetVersion"),
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
			BaseNode: BaseNode{
				Token: NewTestToken(lexer.PROPERTY, "property"),
			},
			Name:            NewTestIdentifier("Name"),
			Type:            NewTestTypeAnnotation("String"),
			ReadSpec:        NewTestIdentifier("FName"),
			WriteSpec:       NewTestIdentifier("FName"),
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
