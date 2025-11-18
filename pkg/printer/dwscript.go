package printer

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// This file contains DWScript format printing methods for all AST node types.
// These methods produce valid DWScript source code from AST nodes.

// Expression printing methods
// ============================================================================

func (p *Printer) printBinaryExpression(be *ast.BinaryExpression) {
	p.printDWScript(be.Left)
	// Always add spaces around binary operators, even in compact mode
	p.buf.WriteByte(' ')
	p.write(be.Operator)
	p.buf.WriteByte(' ')
	p.printDWScript(be.Right)
}

func (p *Printer) printUnaryExpression(ue *ast.UnaryExpression) {
	p.write(ue.Operator)
	// Add space for keyword operators (required for parsing)
	if ue.Operator == "not" {
		p.requiredSpace()
	}
	p.printDWScript(ue.Right)
}

func (p *Printer) printRangeExpression(re *ast.RangeExpression) {
	p.printDWScript(re.Start)
	p.write("..")
	p.printDWScript(re.RangeEnd)
}

func (p *Printer) printCallExpression(ce *ast.CallExpression) {
	p.printDWScript(ce.Function)
	p.write("(")
	for i, arg := range ce.Arguments {
		p.printDWScript(arg)
		if i < len(ce.Arguments)-1 {
			p.write(",")
			p.space()
		}
	}
	p.write(")")
}

func (p *Printer) printArrayLiteral(al *ast.ArrayLiteralExpression) {
	p.write("[")
	for i, elem := range al.Elements {
		p.printDWScript(elem)
		if i < len(al.Elements)-1 {
			p.write(",")
			p.space()
		}
	}
	p.write("]")
}

func (p *Printer) printIndexExpression(ie *ast.IndexExpression) {
	p.printDWScript(ie.Left)
	p.write("[")
	p.printDWScript(ie.Index)
	p.write("]")
}

func (p *Printer) printNewArrayExpression(nae *ast.NewArrayExpression) {
	p.write("new")
	p.space()
	p.printDWScript(nae.ElementTypeName)
	p.write("[")
	for i, dim := range nae.Dimensions {
		p.printDWScript(dim)
		if i < len(nae.Dimensions)-1 {
			p.write(",")
			p.space()
		}
	}
	p.write("]")
}

func (p *Printer) printSetLiteral(sl *ast.SetLiteral) {
	p.write("[")
	for i, elem := range sl.Elements {
		p.printDWScript(elem)
		if i < len(sl.Elements)-1 {
			p.write(",")
			p.space()
		}
	}
	p.write("]")
}

func (p *Printer) printNewExpression(ne *ast.NewExpression) {
	p.write("new")
	p.space()
	p.printDWScript(ne.ClassName)
	if len(ne.Arguments) > 0 {
		p.write("(")
		for i, arg := range ne.Arguments {
			p.printDWScript(arg)
			if i < len(ne.Arguments)-1 {
				p.write(",")
				p.space()
			}
		}
		p.write(")")
	}
}

func (p *Printer) printMemberAccessExpression(mae *ast.MemberAccessExpression) {
	p.printDWScript(mae.Object)
	p.write(".")
	p.printDWScript(mae.Member)
}

func (p *Printer) printMethodCallExpression(mce *ast.MethodCallExpression) {
	p.printDWScript(mce.Object)
	p.write(".")
	p.printDWScript(mce.Method)
	p.write("(")
	for i, arg := range mce.Arguments {
		p.printDWScript(arg)
		if i < len(mce.Arguments)-1 {
			p.write(",")
			p.space()
		}
	}
	p.write(")")
}

func (p *Printer) printIsExpression(ie *ast.IsExpression) {
	p.printDWScript(ie.Left)
	p.space()
	p.write("is")
	p.space()
	if ie.TargetType != nil {
		p.printDWScript(ie.TargetType)
	} else if ie.Right != nil {
		p.printDWScript(ie.Right)
	}
}

func (p *Printer) printAsExpression(ae *ast.AsExpression) {
	p.printDWScript(ae.Left)
	p.space()
	p.write("as")
	p.space()
	p.printDWScript(ae.TargetType)
}

func (p *Printer) printImplementsExpression(ie *ast.ImplementsExpression) {
	p.printDWScript(ie.Left)
	p.space()
	p.write("implements")
	p.space()
	p.printDWScript(ie.TargetType)
}

func (p *Printer) printRecordLiteral(rl *ast.RecordLiteralExpression) {
	// Record literal syntax: [field1: value1, field2: value2]
	// Or constructor syntax: TRecord.Create(value1, value2)
	if rl.TypeName != nil {
		p.printDWScript(rl.TypeName)
		p.write(".Create")
	}
	p.write("(")
	for i, init := range rl.Fields {
		if init.Name != nil {
			p.printDWScript(init.Name)
			p.write(":")
			p.space()
		}
		p.printDWScript(init.Value)
		if i < len(rl.Fields)-1 {
			p.write(",")
			p.space()
		}
	}
	p.write(")")
}

func (p *Printer) printLambdaExpression(le *ast.LambdaExpression) {
	p.write("lambda")
	p.space()
	if len(le.Parameters) > 0 {
		p.write("(")
		for i, param := range le.Parameters {
			p.printParameter(param)
			if i < len(le.Parameters)-1 {
				p.write(";")
				p.space()
			}
		}
		p.write(")")
		p.space()
	}
	if le.Body != nil {
		p.printDWScript(le.Body)
	}
}

// Statement printing methods
// ============================================================================

func (p *Printer) printBlockStatement(bs *ast.BlockStatement) {
	p.write("begin")
	p.newline()
	p.incIndent()

	for _, stmt := range bs.Statements {
		p.writeIndent()
		p.printDWScript(stmt)
		p.write(";")
		p.newline()
	}

	p.decIndent()
	p.writeIndent()
	p.write("end")
}

func (p *Printer) printVarDeclStatement(vds *ast.VarDeclStatement) {
	p.write("var")
	p.requiredSpace()

	// Print all variable names
	for i, name := range vds.Names {
		p.printDWScript(name)
		if i < len(vds.Names)-1 {
			p.write(",")
			p.space()
		}
	}

	if vds.Type != nil {
		p.write(":")
		p.space()
		p.printDWScript(vds.Type)
	}

	if vds.Value != nil {
		p.space()
		p.write(":=")
		p.space()
		p.printDWScript(vds.Value)
	}

	if vds.IsExternal {
		p.write(";")
		p.requiredSpace()
		p.write("external")
		if vds.ExternalName != "" {
			p.requiredSpace()
			p.write(fmt.Sprintf("'%s'", vds.ExternalName))
		}
	}
}

func (p *Printer) printAssignmentStatement(as *ast.AssignmentStatement) {
	p.printDWScript(as.Target)
	p.space()
	// Convert token type to operator string
	op := ""
	switch as.Operator {
	case token.ASSIGN:
		op = ":="
	case token.PLUS_ASSIGN:
		op = "+="
	case token.MINUS_ASSIGN:
		op = "-="
	case token.TIMES_ASSIGN:
		op = "*="
	case token.DIVIDE_ASSIGN:
		op = "/="
	default:
		op = ":="
	}
	p.write(op)
	p.space()
	p.printDWScript(as.Value)
}

func (p *Printer) printConstDecl(cd *ast.ConstDecl) {
	p.write("const")
	p.space()
	p.printDWScript(cd.Name)

	if cd.Type != nil {
		p.write(":")
		p.space()
		p.printDWScript(cd.Type)
	}

	p.space()
	p.write("=")
	p.space()
	p.printDWScript(cd.Value)
}

func (p *Printer) printReturnStatement(rs *ast.ReturnStatement) {
	// Use the token literal to preserve the original return syntax
	// This handles: result, function name, or exit
	p.write(rs.Token.Literal)

	// Only add assignment if there's a return value
	if rs.ReturnValue != nil {
		p.requiredSpace()
		p.write(":=")
		p.requiredSpace()
		p.printDWScript(rs.ReturnValue)
	}
}

// Control flow printing methods
// ============================================================================

func (p *Printer) printIfStatement(is *ast.IfStatement) {
	p.write("if")
	p.space()
	p.printDWScript(is.Condition)
	p.space()
	p.write("then")

	// Handle consequence
	if _, isBlock := is.Consequence.(*ast.BlockStatement); isBlock {
		p.newline()
		p.writeIndent()
		p.printDWScript(is.Consequence)
	} else {
		p.newline()
		p.incIndent()
		p.writeIndent()
		p.printDWScript(is.Consequence)
		p.decIndent()
	}

	// Handle alternative
	if is.Alternative != nil {
		p.newline()
		p.writeIndent()
		p.write("else")

		if _, isBlock := is.Alternative.(*ast.BlockStatement); isBlock {
			p.newline()
			p.writeIndent()
			p.printDWScript(is.Alternative)
		} else if ifStmt, isIf := is.Alternative.(*ast.IfStatement); isIf {
			// else if chain
			p.space()
			p.printIfStatement(ifStmt)
		} else {
			p.newline()
			p.incIndent()
			p.writeIndent()
			p.printDWScript(is.Alternative)
			p.decIndent()
		}
	}
}

func (p *Printer) printIfExpression(ie *ast.IfExpression) {
	p.write("if")
	p.space()
	p.printDWScript(ie.Condition)
	p.space()
	p.write("then")
	p.space()
	p.printDWScript(ie.Consequence)
	if ie.Alternative != nil {
		p.space()
		p.write("else")
		p.space()
		p.printDWScript(ie.Alternative)
	}
}

func (p *Printer) printWhileStatement(ws *ast.WhileStatement) {
	p.write("while")
	p.space()
	p.printDWScript(ws.Condition)
	p.space()
	p.write("do")
	p.newline()
	p.incIndent()
	p.writeIndent()
	p.printDWScript(ws.Body)
	p.decIndent()
}

func (p *Printer) printRepeatStatement(rs *ast.RepeatStatement) {
	p.write("repeat")
	p.newline()
	p.incIndent()
	p.writeIndent()
	p.printDWScript(rs.Body)
	p.decIndent()
	p.newline()
	p.writeIndent()
	p.write("until")
	p.space()
	p.printDWScript(rs.Condition)
}

func (p *Printer) printForStatement(fs *ast.ForStatement) {
	p.write("for")
	p.space()
	p.printDWScript(fs.Variable)
	p.space()
	p.write(":=")
	p.space()
	p.printDWScript(fs.Start)
	p.space()

	if fs.Direction == ast.ForTo {
		p.write("to")
	} else {
		p.write("downto")
	}
	p.space()
	p.printDWScript(fs.EndValue)

	if fs.Step != nil {
		p.space()
		p.write("step")
		p.space()
		p.printDWScript(fs.Step)
	}

	p.space()
	p.write("do")
	p.newline()
	p.incIndent()
	p.writeIndent()
	p.printDWScript(fs.Body)
	p.decIndent()
}

func (p *Printer) printForInStatement(fis *ast.ForInStatement) {
	p.write("for")
	p.space()
	p.printDWScript(fis.Variable)
	p.space()
	p.write("in")
	p.space()
	p.printDWScript(fis.Collection)
	p.space()
	p.write("do")
	p.newline()
	p.incIndent()
	p.writeIndent()
	p.printDWScript(fis.Body)
	p.decIndent()
}

func (p *Printer) printCaseStatement(cs *ast.CaseStatement) {
	p.write("case")
	p.space()
	p.printDWScript(cs.Expression)
	p.space()
	p.write("of")
	p.newline()

	p.incIndent()
	for _, branch := range cs.Cases {
		p.writeIndent()
		for i, value := range branch.Values {
			p.printDWScript(value)
			if i < len(branch.Values)-1 {
				p.write(",")
				p.space()
			}
		}
		p.write(":")
		p.space()
		p.printDWScript(branch.Statement)
		p.write(";")
		p.newline()
	}
	p.decIndent()

	if cs.Else != nil {
		p.writeIndent()
		p.write("else")
		p.newline()
		p.incIndent()
		p.writeIndent()
		p.printDWScript(cs.Else)
		p.write(";")
		p.newline()
		p.decIndent()
	}

	p.writeIndent()
	p.write("end")
}

// Exception handling printing methods
// ============================================================================

func (p *Printer) printTryStatement(ts *ast.TryStatement) {
	p.write("try")
	p.newline()
	p.incIndent()

	if ts.TryBlock != nil {
		for _, stmt := range ts.TryBlock.Statements {
			p.writeIndent()
			p.printDWScript(stmt)
			p.write(";")
			p.newline()
		}
	}

	p.decIndent()

	if ts.ExceptClause != nil {
		p.writeIndent()
		p.write("except")
		p.newline()
		p.incIndent()

		for _, handler := range ts.ExceptClause.Handlers {
			p.writeIndent()
			p.write("on")
			p.space()
			if handler.Variable != nil {
				p.printDWScript(handler.Variable)
				p.write(":")
				p.space()
			}
			p.printDWScript(handler.ExceptionType)
			p.space()
			p.write("do")
			p.space()
			p.printDWScript(handler.Statement)
			p.write(";")
			p.newline()
		}

		if ts.ExceptClause.ElseBlock != nil {
			p.writeIndent()
			p.write("else")
			p.newline()
			p.incIndent()
			for _, stmt := range ts.ExceptClause.ElseBlock.Statements {
				p.writeIndent()
				p.printDWScript(stmt)
				p.write(";")
				p.newline()
			}
			p.decIndent()
		}

		p.decIndent()
	}

	if ts.FinallyClause != nil {
		p.writeIndent()
		p.write("finally")
		p.newline()
		p.incIndent()

		if ts.FinallyClause.Block != nil {
			for _, stmt := range ts.FinallyClause.Block.Statements {
				p.writeIndent()
				p.printDWScript(stmt)
				p.write(";")
				p.newline()
			}
		}

		p.decIndent()
	}

	p.writeIndent()
	p.write("end")
}

func (p *Printer) printRaiseStatement(rs *ast.RaiseStatement) {
	p.write("raise")
	if rs.Exception != nil {
		p.space()
		p.printDWScript(rs.Exception)
	}
}

// Declaration printing methods
// ============================================================================

func (p *Printer) printFunctionDecl(fd *ast.FunctionDecl) {
	// Print visibility for class methods (only if not public and part of a class)
	// Standalone functions don't have visibility modifiers
	if fd.ClassName != nil && fd.Visibility != ast.VisibilityPublic {
		switch fd.Visibility {
		case ast.VisibilityPrivate, ast.VisibilityProtected:
			p.write(fd.Visibility.String())
			p.requiredSpace()
		}
	}

	// Print class keyword for class methods
	if fd.IsClassMethod {
		p.write("class")
		p.requiredSpace()
	}

	// Print constructor/destructor or function/procedure keyword
	if fd.IsConstructor {
		p.write("constructor")
	} else if fd.IsDestructor {
		p.write("destructor")
	} else if fd.ReturnType != nil {
		p.write("function")
	} else {
		p.write("procedure")
	}
	p.requiredSpace()

	// Print name
	p.printDWScript(fd.Name)

	// Print parameters
	if len(fd.Parameters) > 0 {
		p.write("(")
		for i, param := range fd.Parameters {
			p.printParameter(param)
			if i < len(fd.Parameters)-1 {
				p.write(";")
				p.space()
			}
		}
		p.write(")")
	}

	// Print return type for functions
	if fd.ReturnType != nil {
		p.write(":")
		p.requiredSpace()
		p.printDWScript(fd.ReturnType)
	}

	// Print method modifiers (virtual, override, etc.)
	if fd.IsVirtual {
		p.write(";")
		p.requiredSpace()
		p.write("virtual")
	}
	if fd.IsOverride {
		p.write(";")
		p.requiredSpace()
		p.write("override")
	}
	if fd.IsReintroduce {
		p.write(";")
		p.requiredSpace()
		p.write("reintroduce")
	}
	if fd.IsAbstract {
		p.write(";")
		p.requiredSpace()
		p.write("abstract")
	}
	if fd.IsOverload {
		p.write(";")
		p.requiredSpace()
		p.write("overload")
	}

	// Print calling convention
	if fd.CallingConvention != "" {
		p.write(";")
		p.requiredSpace()
		p.write(fd.CallingConvention)
	}

	// Print deprecated
	if fd.IsDeprecated {
		p.write(";")
		p.requiredSpace()
		p.write("deprecated")
		if fd.DeprecatedMessage != "" {
			p.requiredSpace()
			p.write(fmt.Sprintf("'%s'", fd.DeprecatedMessage))
		}
	}

	// Print forward/external modifiers
	if fd.IsForward {
		p.write(";")
		p.requiredSpace()
		p.write("forward")
		return
	}
	if fd.IsExternal {
		p.write(";")
		p.requiredSpace()
		p.write("external")
		if fd.ExternalName != "" {
			p.requiredSpace()
			p.write(fmt.Sprintf("'%s'", fd.ExternalName))
		}
		return
	}

	// Print body
	if fd.Body != nil {
		p.write(";")
		p.newline()
		p.printDWScript(fd.Body)
	}
}

func (p *Printer) printParameter(param *ast.Parameter) {
	// Print parameter modifiers
	if param.IsConst {
		p.write("const")
		p.space()
	} else if param.ByRef {
		p.write("var")
		p.space()
	}

	// Print name
	p.printDWScript(param.Name)

	// Print type
	if param.Type != nil {
		p.write(":")
		p.space()
		p.printDWScript(param.Type)
	}

	// Print default value
	if param.DefaultValue != nil {
		p.space()
		p.write("=")
		p.space()
		p.printDWScript(param.DefaultValue)
	}
}

func (p *Printer) printClassDecl(cd *ast.ClassDecl) {
	p.write("type")
	p.space()
	p.printDWScript(cd.Name)
	p.space()
	p.write("=")
	p.space()

	if cd.IsPartial {
		p.write("partial")
		p.space()
	}

	p.write("class")

	// Print parent and interfaces
	if cd.Parent != nil || len(cd.Interfaces) > 0 {
		p.write("(")
		if cd.Parent != nil {
			p.printDWScript(cd.Parent)
			if len(cd.Interfaces) > 0 {
				p.write(",")
				p.space()
			}
		}
		for i, intf := range cd.Interfaces {
			p.printDWScript(intf)
			if i < len(cd.Interfaces)-1 {
				p.write(",")
				p.space()
			}
		}
		p.write(")")
	}

	if cd.IsAbstract {
		p.space()
		p.write("abstract")
	}

	if cd.IsExternal {
		p.space()
		p.write("external")
		if cd.ExternalName != "" {
			p.space()
			p.write(fmt.Sprintf("'%s'", cd.ExternalName))
		}
	}

	p.newline()

	// Print members
	p.incIndent()

	// Constants
	for _, constant := range cd.Constants {
		p.writeIndent()
		p.printConstDecl(constant)
		p.write(";")
		p.newline()
	}

	// Fields
	for _, field := range cd.Fields {
		p.writeIndent()
		p.printFieldDecl(field)
		p.write(";")
		p.newline()
	}

	// Properties
	for _, prop := range cd.Properties {
		p.writeIndent()
		p.printPropertyDecl(prop)
		p.write(";")
		p.newline()
	}

	// Constructor
	if cd.Constructor != nil {
		p.writeIndent()
		p.printFunctionDecl(cd.Constructor)
		p.write(";")
		p.newline()
	}

	// Destructor
	if cd.Destructor != nil {
		p.writeIndent()
		p.printFunctionDecl(cd.Destructor)
		p.write(";")
		p.newline()
	}

	// Methods
	for _, method := range cd.Methods {
		p.writeIndent()
		p.printFunctionDecl(method)
		p.write(";")
		p.newline()
	}

	// Operators
	for _, operator := range cd.Operators {
		p.writeIndent()
		p.printOperatorDecl(operator)
		p.write(";")
		p.newline()
	}

	p.decIndent()
	p.writeIndent()
	p.write("end")
}

func (p *Printer) printFieldDecl(fd *ast.FieldDecl) {
	// Print visibility if not public
	if fd.Visibility != ast.VisibilityPublic {
		p.write(fd.Visibility.String())
		p.space()
	}

	p.printDWScript(fd.Name)

	// Handle type inference syntax (Type is nil, only InitValue present)
	if fd.Type != nil {
		// Explicit type: Name: Type [= InitValue]
		p.write(":")
		p.space()
		p.printDWScript(fd.Type)

		if fd.InitValue != nil {
			p.space()
			p.write("=")
			p.space()
			p.printDWScript(fd.InitValue)
		}
	} else if fd.InitValue != nil {
		// Type inference: Name := InitValue
		p.space()
		p.write(":=")
		p.space()
		p.printDWScript(fd.InitValue)
	}
}

func (p *Printer) printPropertyDecl(pd *ast.PropertyDecl) {
	p.write("property")
	p.space()
	p.printDWScript(pd.Name)
	p.write(":")
	p.space()
	p.printDWScript(pd.Type)

	if pd.ReadSpec != nil {
		p.space()
		p.write("read")
		p.space()
		p.printDWScript(pd.ReadSpec)
	}

	if pd.WriteSpec != nil {
		p.space()
		p.write("write")
		p.space()
		p.printDWScript(pd.WriteSpec)
	}

	if pd.IsDefault {
		p.write(";")
		p.space()
		p.write("default")
	}
}

func (p *Printer) printRecordDecl(rd *ast.RecordDecl) {
	p.write("type")
	p.space()
	p.printDWScript(rd.Name)
	p.space()
	p.write("=")
	p.space()
	p.write("record")
	p.newline()

	p.incIndent()

	// Print constants
	for _, constant := range rd.Constants {
		p.writeIndent()
		if constant.IsClassConst {
			p.write("class")
			p.space()
		}
		p.write("const")
		p.space()
		p.printDWScript(constant.Name)
		if constant.Type != nil {
			p.write(":")
			p.space()
			p.printDWScript(constant.Type)
		}
		p.space()
		p.write("=")
		p.space()
		p.printDWScript(constant.Value)
		p.write(";")
		p.newline()
	}

	// Print class variables
	for _, classVar := range rd.ClassVars {
		p.writeIndent()
		p.write("class")
		p.space()
		p.write("var")
		p.space()
		p.printFieldDecl(classVar)
		p.write(";")
		p.newline()
	}

	// Print fields
	for _, field := range rd.Fields {
		p.writeIndent()
		p.printFieldDecl(field)
		p.write(";")
		p.newline()
	}

	// Print methods
	for _, method := range rd.Methods {
		p.writeIndent()
		p.printFunctionDecl(method)
		p.write(";")
		p.newline()
	}

	// Print properties
	for _, property := range rd.Properties {
		p.writeIndent()
		p.write("property")
		p.space()
		p.printDWScript(property.Name)
		p.write(":")
		p.space()
		p.printDWScript(property.Type)
		if property.ReadField != "" {
			p.space()
			p.write("read")
			p.space()
			p.write(property.ReadField)
		}
		if property.WriteField != "" {
			p.space()
			p.write("write")
			p.space()
			p.write(property.WriteField)
		}
		p.write(";")
		p.newline()
	}

	p.decIndent()

	p.writeIndent()
	p.write("end")
}

func (p *Printer) printEnumDecl(ed *ast.EnumDecl) {
	p.write("type")
	p.space()
	p.printDWScript(ed.Name)
	p.space()
	p.write("=")
	p.space()
	p.write("(")

	for i, value := range ed.Values {
		p.write(value.Name)
		if value.Value != nil {
			p.space()
			p.write("=")
			p.space()
			p.write(fmt.Sprintf("%d", *value.Value))
		}
		if i < len(ed.Values)-1 {
			p.write(",")
			p.space()
		}
	}

	p.write(")")
}

func (p *Printer) printArrayDecl(ad *ast.ArrayDecl) {
	p.write("type")
	p.space()
	p.printDWScript(ad.Name)
	p.space()
	p.write("=")
	p.space()
	if ad.ArrayType != nil {
		p.printDWScript(ad.ArrayType)
	} else {
		p.write("array")
	}
}

func (p *Printer) printSetDecl(sd *ast.SetDecl) {
	p.write("type")
	p.space()
	p.printDWScript(sd.Name)
	p.space()
	p.write("=")
	p.space()
	p.write("set")
	p.space()
	p.write("of")
	p.space()
	p.printDWScript(sd.ElementType)
}

func (p *Printer) printInterfaceDecl(id *ast.InterfaceDecl) {
	p.write("type")
	p.space()
	p.printDWScript(id.Name)
	p.space()
	p.write("=")
	p.space()
	p.write("interface")

	if id.Parent != nil {
		p.write("(")
		p.printDWScript(id.Parent)
		p.write(")")
	}

	p.newline()

	p.incIndent()
	for _, method := range id.Methods {
		p.writeIndent()
		p.printInterfaceMethodDecl(method)
		p.write(";")
		p.newline()
	}
	p.decIndent()

	p.writeIndent()
	p.write("end")
}

func (p *Printer) printInterfaceMethodDecl(imd *ast.InterfaceMethodDecl) {
	if imd.ReturnType != nil {
		p.write("function")
	} else {
		p.write("procedure")
	}
	p.space()
	p.printDWScript(imd.Name)

	if len(imd.Parameters) > 0 {
		p.write("(")
		for i, param := range imd.Parameters {
			p.printParameter(param)
			if i < len(imd.Parameters)-1 {
				p.write(";")
				p.space()
			}
		}
		p.write(")")
	}

	if imd.ReturnType != nil {
		p.write(":")
		p.space()
		p.printDWScript(imd.ReturnType)
	}
}

func (p *Printer) printTypeDeclaration(td *ast.TypeDeclaration) {
	p.write("type")
	p.space()
	p.printDWScript(td.Name)
	p.space()
	p.write("=")
	p.space()
	if td.AliasedType != nil {
		p.printDWScript(td.AliasedType)
	} else if td.FunctionPointerType != nil {
		p.printDWScript(td.FunctionPointerType)
	} else if td.IsSubrange {
		p.printDWScript(td.LowBound)
		p.write("..")
		p.printDWScript(td.HighBound)
	}
}

func (p *Printer) printUnitDeclaration(ud *ast.UnitDeclaration) {
	p.write("unit")
	p.space()
	p.printDWScript(ud.Name)
	p.write(";")
	p.newline()
	p.newline()

	// Print interface section
	p.write("interface")
	p.newline()
	p.newline()

	if ud.InterfaceSection != nil {
		p.incIndent()
		for _, stmt := range ud.InterfaceSection.Statements {
			p.writeIndent()
			p.printDWScript(stmt)
			p.write(";")
			p.newline()
		}
		p.decIndent()
	}

	p.newline()

	// Print implementation section
	p.write("implementation")
	p.newline()
	p.newline()

	if ud.ImplementationSection != nil {
		p.incIndent()
		for _, stmt := range ud.ImplementationSection.Statements {
			p.writeIndent()
			p.printDWScript(stmt)
			p.write(";")
			p.newline()
		}
		p.decIndent()
	}

	p.newline()
	p.write("end.")
}

// Additional node type printing methods
// ============================================================================

func (p *Printer) printEnumLiteral(el *ast.EnumLiteral) {
	if el.EnumName != "" {
		p.write(el.EnumName)
		p.write(".")
	}
	p.write(el.ValueName)
}

func (p *Printer) printOldExpression(oe *ast.OldExpression) {
	p.write("old")
	p.space()
	p.printDWScript(oe.Identifier)
}

func (p *Printer) printHelperDecl(hd *ast.HelperDecl) {
	p.write("type")
	p.space()
	p.printDWScript(hd.Name)
	p.space()
	p.write("=")
	p.space()
	p.write("helper")

	if hd.ParentHelper != nil {
		p.write("(")
		p.printDWScript(hd.ParentHelper)
		p.write(")")
	}

	p.space()
	p.write("for")
	p.space()
	p.printDWScript(hd.ForType)
	p.newline()

	p.incIndent()

	// Print class vars
	for _, classVar := range hd.ClassVars {
		p.writeIndent()
		p.write("class var")
		p.space()
		p.printDWScript(classVar)
		p.write(";")
		p.newline()
	}

	// Print class consts
	for _, classConst := range hd.ClassConsts {
		p.writeIndent()
		p.write("class const")
		p.space()
		p.printDWScript(classConst)
		p.write(";")
		p.newline()
	}

	// Print methods
	for _, method := range hd.Methods {
		p.writeIndent()
		p.printDWScript(method)
		p.write(";")
		p.newline()
	}

	// Print properties
	for _, prop := range hd.Properties {
		p.writeIndent()
		p.printDWScript(prop)
		p.write(";")
		p.newline()
	}

	p.decIndent()
	p.writeIndent()
	p.write("end")
}

func (p *Printer) printOperatorDecl(od *ast.OperatorDecl) {
	// Print visibility if not public
	if od.Visibility != ast.VisibilityPublic {
		p.write(od.Visibility.String())
		p.space()
	}

	// Print class keyword for class operators
	if od.Kind == ast.OperatorKindClass {
		p.write("class")
		p.space()
	}

	p.write("operator")
	p.space()
	p.write(od.OperatorSymbol)

	// Print operand types
	if len(od.OperandTypes) > 0 {
		p.write("(")
		for i, opType := range od.OperandTypes {
			p.printDWScript(opType)
			if i < len(od.OperandTypes)-1 {
				p.write(",")
				p.space()
			}
		}
		p.write(")")
	}

	// Print return type
	if od.ReturnType != nil {
		p.write(":")
		p.space()
		p.printDWScript(od.ReturnType)
	}

	// Print binding (implementation)
	if od.Binding != nil {
		p.space()
		p.write("uses")
		p.space()
		p.printDWScript(od.Binding)
	}
}

func (p *Printer) printRecordPropertyDecl(rpd *ast.RecordPropertyDecl) {
	p.write("property")
	p.space()
	p.printDWScript(rpd.Name)
	p.write(":")
	p.space()
	p.printDWScript(rpd.Type)

	if rpd.ReadField != "" {
		p.space()
		p.write("read")
		p.space()
		p.write(rpd.ReadField)
	}

	if rpd.WriteField != "" {
		p.space()
		p.write("write")
		p.space()
		p.write(rpd.WriteField)
	}
}

// Type annotation printing methods
// ============================================================================

func (p *Printer) printTypeAnnotation(ta *ast.TypeAnnotation) {
	if ta == nil {
		return
	}

	if ta.InlineType != nil {
		p.printDWScript(ta.InlineType)
	} else if ta.Name != "" {
		p.write(ta.Name)
	}
}

func (p *Printer) printArrayTypeAnnotation(ata *ast.ArrayTypeAnnotation) {
	p.write("array")

	// Print bounds if present
	if ata.LowBound != nil && ata.HighBound != nil {
		p.write("[")
		p.printDWScript(ata.LowBound)
		p.write("..")
		p.printDWScript(ata.HighBound)
		p.write("]")
	}

	p.space()
	p.write("of")
	p.space()
	p.printDWScript(ata.ElementType)
}

func (p *Printer) printArrayTypeNode(atn *ast.ArrayTypeNode) {
	p.write("array")

	// Print index type or bounds
	if atn.IndexType != nil {
		p.write("[")
		p.printDWScript(atn.IndexType)
		p.write("]")
	} else if atn.LowBound != nil && atn.HighBound != nil {
		p.write("[")
		p.printDWScript(atn.LowBound)
		p.write("..")
		p.printDWScript(atn.HighBound)
		p.write("]")
	}

	p.space()
	p.write("of")
	p.space()
	p.printDWScript(atn.ElementType)
}

func (p *Printer) printSetTypeNode(stn *ast.SetTypeNode) {
	p.write("set")
	p.space()
	p.write("of")
	p.space()
	p.printDWScript(stn.ElementType)
}

func (p *Printer) printClassOfTypeNode(cotn *ast.ClassOfTypeNode) {
	p.write("class")
	p.space()
	p.write("of")
	p.space()
	p.printDWScript(cotn.ClassType)
}

func (p *Printer) printFunctionPointerTypeNode(fptn *ast.FunctionPointerTypeNode) {
	if fptn.ReturnType != nil {
		p.write("function")
	} else {
		p.write("procedure")
	}

	if len(fptn.Parameters) > 0 {
		p.write("(")
		for i, param := range fptn.Parameters {
			p.printParameter(param)
			if i < len(fptn.Parameters)-1 {
				p.write(";")
				p.space()
			}
		}
		p.write(")")
	}

	if fptn.ReturnType != nil {
		p.write(":")
		p.space()
		p.printDWScript(fptn.ReturnType)
	}

	if fptn.OfObject {
		p.space()
		p.write("of object")
	}
}
