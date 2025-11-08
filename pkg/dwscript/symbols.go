package dwscript

import (
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// Symbol represents a symbol in the program's symbol table.
// A symbol can be a variable, constant, function, class, parameter, or other declaration.
//
// Symbols are extracted from the semantic analyzer's symbol table after compilation
// and are useful for IDE features like code completion, go-to-definition, and hover information.
type Symbol struct {
	Name       string
	Kind       string
	Type       string
	Scope      string
	Position   token.Position
	IsReadOnly bool
	IsConst    bool
}

// Symbols returns all symbols declared in the program.
// This includes variables, constants, functions, classes, and other declarations.
//
// Symbols are organized hierarchically by scope. Global symbols appear first,
// followed by symbols from inner scopes.
//
// If the program was not type-checked (e.g., compiled with TypeCheck: false),
// this method returns an empty slice as symbol information is not available.
//
// Example usage:
//
//	program, _ := engine.Compile(`
//	    var x: Integer := 42;
//	    function Add(a, b: Integer): Integer;
//	    begin
//	        Result := a + b;
//	    end;
//	`)
//
//	symbols := program.Symbols()
//	for _, sym := range symbols {
//	    fmt.Printf("Symbol: %s (%s) at %s\n", sym.Name, sym.Kind, sym.Position)
//	}
func (p *Program) Symbols() []Symbol {
	// If no analyzer (e.g., type checking was disabled), return empty slice
	if p.analyzer == nil {
		return []Symbol{}
	}

	// Extract symbols from the analyzer
	return extractSymbols(p.analyzer)
}

// extractSymbols walks through the analyzer's symbol table and extracts all symbols.
func extractSymbols(analyzer *semantic.Analyzer) []Symbol {
	result := []Symbol{}

	// Extract variables and functions from symbol table
	symbolTable := analyzer.GetSymbolTable()
	if symbolTable != nil {
		for _, sym := range symbolTable.AllSymbols() {
			// Determine kind based on symbol type
			kind := determineSymbolKind(sym)

			result = append(result, Symbol{
				Name:       sym.Name,
				Kind:       kind,
				Type:       sym.Type.String(),
				Position:   token.Position{}, // Position info not stored in symbol table
				Scope:      "global",         // TODO: Track actual scope level
				IsReadOnly: sym.ReadOnly,
				IsConst:    sym.IsConst,
			})
		}
	}

	// Extract type declarations (classes, interfaces, enums, records)
	// These are stored separately from the symbol table

	// Extract classes
	for name, classType := range analyzer.GetClasses() {
		result = append(result, Symbol{
			Name:       name,
			Kind:       "class",
			Type:       classType.String(),
			Position:   token.Position{},
			Scope:      "global",
			IsReadOnly: false,
			IsConst:    false,
		})
	}

	// Extract interfaces
	for name, interfaceType := range analyzer.GetInterfaces() {
		result = append(result, Symbol{
			Name:       name,
			Kind:       "interface",
			Type:       interfaceType.String(),
			Position:   token.Position{},
			Scope:      "global",
			IsReadOnly: false,
			IsConst:    false,
		})
	}

	// Extract enums
	for name, enumType := range analyzer.GetEnums() {
		result = append(result, Symbol{
			Name:       name,
			Kind:       "enum",
			Type:       enumType.String(),
			Position:   token.Position{},
			Scope:      "global",
			IsReadOnly: false,
			IsConst:    false,
		})
	}

	// Extract records
	for name, recordType := range analyzer.GetRecords() {
		result = append(result, Symbol{
			Name:       name,
			Kind:       "record",
			Type:       recordType.String(),
			Position:   token.Position{},
			Scope:      "global",
			IsReadOnly: false,
			IsConst:    false,
		})
	}

	// Extract array types
	for name, arrayType := range analyzer.GetArrayTypes() {
		result = append(result, Symbol{
			Name:       name,
			Kind:       "type",
			Type:       arrayType.String(),
			Position:   token.Position{},
			Scope:      "global",
			IsReadOnly: false,
			IsConst:    false,
		})
	}

	// Extract type aliases
	for name, typeAlias := range analyzer.GetTypeAliases() {
		result = append(result, Symbol{
			Name:       name,
			Kind:       "type",
			Type:       typeAlias.String(),
			Position:   token.Position{},
			Scope:      "global",
			IsReadOnly: false,
			IsConst:    false,
		})
	}

	return result
}

// determineSymbolKind determines the kind of a symbol based on its type.
func determineSymbolKind(sym *semantic.Symbol) string {
	if sym.IsConst {
		return "constant"
	}

	// Check if it's a function by looking at the type
	switch sym.Type.(type) {
	case *types.FunctionType:
		return "function"
	default:
		return "variable"
	}
}

// TypeAt returns the type of the expression at the given position in the source code.
//
// This method is useful for IDE features like hover information, where you want to
// show the type of an expression when the user hovers over it.
//
// The position should match a position from the AST (1-indexed line and column).
// If the position doesn't map to a typed expression, returns ("", false).
//
// If the program was not type-checked (e.g., compiled with TypeCheck: false),
// this method returns ("", false) as type information is not available.
//
// Example usage:
//
//	program, _ := engine.Compile(`
//	    var x: Integer := 42;
//	    var y := x + 10;
//	`)
//
//	// Get type at position of 'x' in second line
//	typ, ok := program.TypeAt(token.Position{Line: 2, Column: 14})
//	if ok {
//	    fmt.Printf("Type: %s\n", typ) // Output: Type: Integer
//	}
func (p *Program) TypeAt(pos token.Position) (string, bool) {
	// If no analyzer (e.g., type checking was disabled), return empty
	if p.analyzer == nil {
		return "", false
	}

	// Walk the AST to find the node at the given position
	node := findNodeAtPosition(p.ast, pos)
	if node == nil {
		return "", false
	}

	// Get type information from the analyzer for this node
	return getTypeForNode(p.analyzer, node)
}

// findNodeAtPosition walks the AST to find the node at the given position.
// Returns the deepest (most specific) node that contains the position.
func findNodeAtPosition(program *ast.Program, pos token.Position) ast.Node {
	var result ast.Node

	// Use the AST visitor pattern to walk the tree
	ast.Inspect(program, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		// Check if this node contains the position
		nodeStart := node.Pos()
		nodeEnd := node.End()

		// Check if position is within this node's range
		if positionInRange(pos, nodeStart, nodeEnd) {
			// This node contains the position
			// Keep going deeper to find the most specific node
			result = node
			return true
		}

		// Position is not in this node, skip its children
		return false
	})

	return result
}

// positionInRange checks if pos is within the range [start, end].
func positionInRange(pos, start, end token.Position) bool {
	// Check if pos is after or equal to start
	if pos.Line < start.Line {
		return false
	}
	if pos.Line == start.Line && pos.Column < start.Column {
		return false
	}

	// Check if pos is before or equal to end
	if pos.Line > end.Line {
		return false
	}
	if pos.Line == end.Line && pos.Column > end.Column {
		return false
	}

	return true
}

// getTypeForNode retrieves type information for a given AST node.
func getTypeForNode(analyzer *semantic.Analyzer, node ast.Node) (string, bool) {
	// Try to get type based on node type
	switch n := node.(type) {
	case *ast.Identifier:
		// Look up in symbol table
		sym, ok := analyzer.GetSymbolTable().Resolve(n.Value)
		if ok && sym.Type != nil {
			return sym.Type.String(), true
		}

		// Check if it's a type name (class, enum, etc.)
		if classType, ok := analyzer.GetClasses()[n.Value]; ok {
			return classType.String(), true
		}
		if enumType, ok := analyzer.GetEnums()[n.Value]; ok {
			return enumType.String(), true
		}
		if recordType, ok := analyzer.GetRecords()[n.Value]; ok {
			return recordType.String(), true
		}

	case *ast.IntegerLiteral:
		return "Integer", true

	case *ast.FloatLiteral:
		return "Float", true

	case *ast.StringLiteral:
		return "String", true

	case *ast.BooleanLiteral:
		return "Boolean", true

	case *ast.CharLiteral:
		return "Char", true

	case *ast.NilLiteral:
		return "nil", true

	case *ast.BinaryExpression:
		// For binary expressions, we'd need to analyze the operator and operands
		// This requires more complex type inference
		// For now, we'll return false and let the caller try parent nodes
		return "", false

	case *ast.UnaryExpression:
		// Similar to binary expressions
		return "", false

	case *ast.CallExpression:
		// For function calls, we'd need to look up the function and return its return type
		// This is more complex and would require tracking function signatures
		return "", false

	case ast.Expression:
		// Generic expression - we don't have type info
		return "", false

	case ast.Statement:
		// Statements don't have types
		return "", false
	}

	return "", false
}
