package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// Format specifies the output format for the printer.
type Format int

const (
	// FormatDWScript produces valid DWScript source code (default).
	FormatDWScript Format = iota

	// FormatTree produces a hierarchical AST structure visualization.
	FormatTree

	// FormatJSON produces a JSON representation of the AST.
	FormatJSON
)

// String returns the string representation of the format.
func (f Format) String() string {
	switch f {
	case FormatDWScript:
		return "dwscript"
	case FormatTree:
		return "tree"
	case FormatJSON:
		return "json"
	default:
		return "unknown"
	}
}

// Style specifies the formatting style for the printer.
type Style int

const (
	// StyleDetailed uses full indentation and spacing (default).
	StyleDetailed Style = iota

	// StyleCompact minimizes whitespace for single-line output.
	StyleCompact

	// StyleMultiline ensures every statement is on a new line.
	StyleMultiline
)

// String returns the string representation of the style.
func (s Style) String() string {
	switch s {
	case StyleDetailed:
		return "detailed"
	case StyleCompact:
		return "compact"
	case StyleMultiline:
		return "multiline"
	default:
		return "unknown"
	}
}

// Options configures the printer behavior.
type Options struct {
	// Format specifies the output format (DWScript, Tree, JSON).
	Format Format

	// Style specifies the formatting style (Detailed, Compact, Multiline).
	Style Style

	// IndentWidth specifies the number of spaces per indent level.
	// Default is 2 for DWScript format, 2 for Tree format.
	IndentWidth int

	// UseSpaces specifies whether to use spaces (true) or tabs (false) for indentation.
	// Default is true (spaces).
	UseSpaces bool

	// IncludePositions includes source position information in output (Tree and JSON formats only).
	IncludePositions bool

	// IncludeTypes includes type information in output (Tree and JSON formats only).
	IncludeTypes bool
}

// DefaultOptions returns the default printer options.
func DefaultOptions() Options {
	return Options{
		Format:           FormatDWScript,
		Style:            StyleDetailed,
		IndentWidth:      2,
		UseSpaces:        true,
		IncludePositions: false,
		IncludeTypes:     false,
	}
}

// Printer formats AST nodes into text output.
type Printer struct {
	opts   Options
	buf    bytes.Buffer
	indent int
}

// New creates a new Printer with the given options.
func New(opts Options) *Printer {
	// Apply defaults for unset values
	if opts.IndentWidth == 0 {
		opts.IndentWidth = 2
	}
	return &Printer{
		opts:   opts,
		indent: 0,
	}
}

// Print formats the given node and returns the output as a string.
// This is a convenience function that creates a printer with default options.
func Print(node ast.Node) string {
	p := New(DefaultOptions())
	return p.Print(node)
}

// Print formats the given node using the printer's configuration.
func (p *Printer) Print(node ast.Node) string {
	p.buf.Reset()
	p.indent = 0

	switch p.opts.Format {
	case FormatDWScript:
		p.printDWScript(node)
	case FormatTree:
		p.printTree(node)
	case FormatJSON:
		p.printJSON(node)
	default:
		p.printDWScript(node)
	}

	return p.buf.String()
}

// Helper methods for building output
// ============================================================================

// write writes a string to the buffer without indentation.
func (p *Printer) write(s string) {
	p.buf.WriteString(s)
}

// writeln writes a string followed by a newline (respects compact style).
func (p *Printer) writeln(s string) {
	if p.opts.Style == StyleCompact {
		p.buf.WriteString(s)
	} else {
		p.buf.WriteString(s)
		p.buf.WriteByte('\n')
	}
}

// writeIndent writes the current indentation.
func (p *Printer) writeIndent() {
	if p.opts.Style == StyleCompact {
		return // No indentation in compact mode
	}

	indentStr := p.getIndentString()
	p.buf.WriteString(indentStr)
}

// getIndentString returns the string for the current indent level.
func (p *Printer) getIndentString() string {
	if p.opts.Style == StyleCompact {
		return ""
	}

	width := p.indent * p.opts.IndentWidth
	if p.opts.UseSpaces {
		return strings.Repeat(" ", width)
	}
	return strings.Repeat("\t", p.indent)
}

// newline writes a newline (unless in compact mode).
func (p *Printer) newline() {
	if p.opts.Style != StyleCompact {
		p.buf.WriteByte('\n')
	}
}

// space writes a space (unless in compact mode).
func (p *Printer) space() {
	if p.opts.Style == StyleCompact {
		return
	}
	p.buf.WriteByte(' ')
}

// incIndent increases the indentation level.
func (p *Printer) incIndent() {
	p.indent++
}

// decIndent decreases the indentation level.
func (p *Printer) decIndent() {
	if p.indent > 0 {
		p.indent--
	}
}

// DWScript Format Printer
// ============================================================================

// printDWScript prints the node in DWScript source format.
func (p *Printer) printDWScript(node ast.Node) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.Program:
		p.printProgram(n)

	// Literals
	case *ast.IntegerLiteral:
		p.write(n.Token.Literal)
	case *ast.FloatLiteral:
		p.write(n.Token.Literal)
	case *ast.StringLiteral:
		p.write(fmt.Sprintf("\"%s\"", n.Value))
	case *ast.BooleanLiteral:
		p.write(n.Token.Literal)
	case *ast.CharLiteral:
		p.write(n.Token.Literal)
	case *ast.NilLiteral:
		p.write("nil")

	// Basic expressions
	case *ast.Identifier:
		p.write(n.Value)
	case *ast.BinaryExpression:
		p.printBinaryExpression(n)
	case *ast.UnaryExpression:
		p.printUnaryExpression(n)
	case *ast.GroupedExpression:
		p.write("(")
		p.printDWScript(n.Expression)
		p.write(")")
	case *ast.RangeExpression:
		p.printRangeExpression(n)

	// Statements
	case *ast.ExpressionStatement:
		p.printDWScript(n.Expression)
	case *ast.BlockStatement:
		p.printBlockStatement(n)
	case *ast.VarDeclStatement:
		p.printVarDeclStatement(n)
	case *ast.AssignmentStatement:
		p.printAssignmentStatement(n)
	case *ast.ConstDecl:
		p.printConstDecl(n)
	case *ast.ReturnStatement:
		p.printReturnStatement(n)

	// Control flow
	case *ast.IfStatement:
		p.printIfStatement(n)
	case *ast.IfExpression:
		p.printIfExpression(n)
	case *ast.WhileStatement:
		p.printWhileStatement(n)
	case *ast.RepeatStatement:
		p.printRepeatStatement(n)
	case *ast.ForStatement:
		p.printForStatement(n)
	case *ast.ForInStatement:
		p.printForInStatement(n)
	case *ast.CaseStatement:
		p.printCaseStatement(n)
	case *ast.BreakStatement:
		p.write("break")
	case *ast.ContinueStatement:
		p.write("continue")
	case *ast.ExitStatement:
		p.write("exit")

	// Exception handling
	case *ast.TryStatement:
		p.printTryStatement(n)
	case *ast.RaiseStatement:
		p.printRaiseStatement(n)

	// Declarations
	case *ast.FunctionDecl:
		p.printFunctionDecl(n)
	case *ast.ClassDecl:
		p.printClassDecl(n)
	case *ast.RecordDecl:
		p.printRecordDecl(n)
	case *ast.EnumDecl:
		p.printEnumDecl(n)
	case *ast.ArrayDecl:
		p.printArrayDecl(n)
	case *ast.SetDecl:
		p.printSetDecl(n)
	case *ast.InterfaceDecl:
		p.printInterfaceDecl(n)
	case *ast.TypeDeclaration:
		p.printTypeDeclaration(n)
	case *ast.UnitDeclaration:
		p.printUnitDeclaration(n)

	// Array and collection expressions
	case *ast.ArrayLiteralExpression:
		p.printArrayLiteral(n)
	case *ast.IndexExpression:
		p.printIndexExpression(n)
	case *ast.NewArrayExpression:
		p.printNewArrayExpression(n)
	case *ast.SetLiteral:
		p.printSetLiteral(n)

	// Object-oriented expressions
	case *ast.NewExpression:
		p.printNewExpression(n)
	case *ast.MemberAccessExpression:
		p.printMemberAccessExpression(n)
	case *ast.MethodCallExpression:
		p.printMethodCallExpression(n)
	case *ast.CallExpression:
		p.printCallExpression(n)
	case *ast.InheritedExpression:
		p.write("inherited")
		if n.Method != nil {
			p.space()
			p.printDWScript(n.Method)
		}

	// Type expressions
	case *ast.IsExpression:
		p.printIsExpression(n)
	case *ast.AsExpression:
		p.printAsExpression(n)
	case *ast.ImplementsExpression:
		p.printImplementsExpression(n)

	// Other
	case *ast.RecordLiteralExpression:
		p.printRecordLiteral(n)
	case *ast.LambdaExpression:
		p.printLambdaExpression(n)
	case *ast.AddressOfExpression:
		p.write("@")
		p.printDWScript(n.Operator)

	default:
		// Fallback: use the node's String() method
		p.write(fmt.Sprintf("%v", node))
	}
}

// printProgram prints a Program node.
func (p *Printer) printProgram(prog *ast.Program) {
	for i, stmt := range prog.Statements {
		p.printDWScript(stmt)
		// Add semicolon and newline between statements
		if i < len(prog.Statements)-1 {
			p.write(";")
			p.newline()
		}
	}
}

// Tree Format Printer
// ============================================================================

// printTree prints the node in tree structure format.
func (p *Printer) printTree(node ast.Node) {
	p.printTreeNode(node, "")
}

// printTreeNode recursively prints a node in tree format.
func (p *Printer) printTreeNode(node ast.Node, prefix string) {
	if node == nil {
		return
	}

	// Print node type and key information
	p.write(prefix)
	p.printTreeNodeInfo(node)
	p.newline()

	// Print children with increased indentation
	childPrefix := prefix + "  "

	switch n := node.(type) {
	case *ast.Program:
		for _, stmt := range n.Statements {
			p.printTreeNode(stmt, childPrefix)
		}
	case *ast.BinaryExpression:
		p.write(childPrefix + "Left:\n")
		p.printTreeNode(n.Left, childPrefix+"  ")
		p.write(childPrefix + "Right:\n")
		p.printTreeNode(n.Right, childPrefix+"  ")
	case *ast.UnaryExpression:
		p.printTreeNode(n.Right, childPrefix)
	case *ast.BlockStatement:
		for _, stmt := range n.Statements {
			p.printTreeNode(stmt, childPrefix)
		}
	case *ast.IfStatement:
		p.write(childPrefix + "Condition:\n")
		p.printTreeNode(n.Condition, childPrefix+"  ")
		p.write(childPrefix + "Consequence:\n")
		p.printTreeNode(n.Consequence, childPrefix+"  ")
		if n.Alternative != nil {
			p.write(childPrefix + "Alternative:\n")
			p.printTreeNode(n.Alternative, childPrefix+"  ")
		}
	case *ast.FunctionDecl:
		if n.Body != nil {
			p.write(childPrefix + "Body:\n")
			p.printTreeNode(n.Body, childPrefix+"  ")
		}
	case *ast.ClassDecl:
		for _, field := range n.Fields {
			p.printTreeNode(field, childPrefix)
		}
		for _, method := range n.Methods {
			p.printTreeNode(method, childPrefix)
		}
	// Add more cases as needed
	}
}

// printTreeNodeInfo prints basic information about a node for tree format.
func (p *Printer) printTreeNodeInfo(node ast.Node) {
	// Get the type name without the package prefix
	typeName := fmt.Sprintf("%T", node)
	// Remove "*ast." prefix
	if len(typeName) > 5 && typeName[:5] == "*ast." {
		typeName = typeName[5:]
	}

	switch n := node.(type) {
	case *ast.Program:
		p.write(fmt.Sprintf("Program (%d statements)", len(n.Statements)))
	case *ast.Identifier:
		p.write(fmt.Sprintf("Identifier: %s", n.Value))
	case *ast.IntegerLiteral:
		p.write(fmt.Sprintf("IntegerLiteral: %d", n.Value))
	case *ast.FloatLiteral:
		p.write(fmt.Sprintf("FloatLiteral: %g", n.Value))
	case *ast.StringLiteral:
		p.write(fmt.Sprintf("StringLiteral: %q", n.Value))
	case *ast.BooleanLiteral:
		p.write(fmt.Sprintf("BooleanLiteral: %v", n.Value))
	case *ast.BinaryExpression:
		p.write(fmt.Sprintf("BinaryExpression (%s)", n.Operator))
	case *ast.UnaryExpression:
		p.write(fmt.Sprintf("UnaryExpression (%s)", n.Operator))
	case *ast.FunctionDecl:
		p.write(fmt.Sprintf("FunctionDecl: %s", n.Name.Value))
	case *ast.ClassDecl:
		p.write(fmt.Sprintf("ClassDecl: %s", n.Name.Value))
	case *ast.FieldDecl:
		p.write(fmt.Sprintf("FieldDecl: %s", n.Name.Value))
	default:
		p.write(typeName)
	}

	// Optionally include position information
	if p.opts.IncludePositions {
		pos := node.Pos()
		if pos.Line > 0 {
			p.write(fmt.Sprintf(" [%d:%d]", pos.Line, pos.Column))
		}
	}
}

// JSON Format Printer
// ============================================================================

// printJSON prints the node in JSON format.
func (p *Printer) printJSON(node ast.Node) {
	data := p.nodeToMap(node)
	var output []byte
	var err error

	if p.opts.Style == StyleCompact {
		output, err = json.Marshal(data)
	} else {
		output, err = json.MarshalIndent(data, "", strings.Repeat(" ", p.opts.IndentWidth))
	}

	if err != nil {
		p.write(fmt.Sprintf(`{"error": "failed to marshal JSON: %v"}`, err))
		return
	}

	p.write(string(output))
}

// nodeToMap converts an AST node to a map for JSON serialization.
func (p *Printer) nodeToMap(node ast.Node) map[string]interface{} {
	if node == nil {
		return nil
	}

	result := make(map[string]interface{})

	// Add type information
	result["type"] = fmt.Sprintf("%T", node)[5:] // Remove "*ast." prefix

	// Add position if requested
	if p.opts.IncludePositions {
		pos := node.Pos()
		if pos.Line > 0 {
			result["position"] = map[string]interface{}{
				"line":   pos.Line,
				"column": pos.Column,
			}
		}
	}

	// Add node-specific fields
	switch n := node.(type) {
	case *ast.Identifier:
		result["value"] = n.Value
	case *ast.IntegerLiteral:
		result["value"] = n.Value
	case *ast.FloatLiteral:
		result["value"] = n.Value
	case *ast.StringLiteral:
		result["value"] = n.Value
	case *ast.BooleanLiteral:
		result["value"] = n.Value
	case *ast.BinaryExpression:
		result["operator"] = n.Operator
		result["left"] = p.nodeToMap(n.Left)
		result["right"] = p.nodeToMap(n.Right)
	case *ast.UnaryExpression:
		result["operator"] = n.Operator
		result["operand"] = p.nodeToMap(n.Right)
	case *ast.Program:
		stmts := make([]interface{}, len(n.Statements))
		for i, stmt := range n.Statements {
			stmts[i] = p.nodeToMap(stmt)
		}
		result["statements"] = stmts
	// Add more cases as needed
	}

	return result
}

// Helper function to convert position to string
func formatPosition(pos token.Position) string {
	if pos.Line == 0 {
		return ""
	}
	return fmt.Sprintf("%d:%d", pos.Line, pos.Column)
}
