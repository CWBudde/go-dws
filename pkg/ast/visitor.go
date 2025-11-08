package ast

// Visitor is the interface for AST traversal using the visitor pattern.
// Implementations should define a Visit method that is called for each node.
// If Visit returns nil, the node's children are not traversed.
// Otherwise, Visit is called recursively for all child nodes.
type Visitor interface {
	Visit(node Node) (w Visitor)
}

// Walk traverses an AST in depth-first order, starting at the given node.
// It calls v.Visit(node) for each node encountered. If v.Visit returns nil,
// traversal of that node's children is skipped. Otherwise, Walk is called
// recursively for each child with the visitor returned by Visit.
//
// Walk follows the standard Go AST pattern from the go/ast package.
func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	// Walk children based on node type
	switch n := node.(type) {
	// Program
	case *Program:
		for _, stmt := range n.Statements {
			Walk(v, stmt)
		}

	// Literals
	case *Identifier:
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *IntegerLiteral:
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *FloatLiteral:
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *StringLiteral:
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *BooleanLiteral:
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *CharLiteral:
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *NilLiteral:
		if n.Type != nil {
			Walk(v, n.Type)
		}

	// Expressions
	case *AsExpression:
		Walk(v, n.Left)
		if n.TargetType != nil {
			Walk(v, n.TargetType)
		}
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *ImplementsExpression:
		Walk(v, n.Left)
		if n.TargetType != nil {
			Walk(v, n.TargetType)
		}
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *BinaryExpression:
		Walk(v, n.Left)
		Walk(v, n.Right)
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *UnaryExpression:
		Walk(v, n.Right)
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *GroupedExpression:
		Walk(v, n.Expression)
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *RangeExpression:
		Walk(v, n.Start)
		Walk(v, n.RangeEnd)
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *CallExpression:
		Walk(v, n.Function)
		for _, arg := range n.Arguments {
			Walk(v, arg)
		}
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *OldExpression:
		Walk(v, n.Identifier)
		if n.Type != nil {
			Walk(v, n.Type)
		}

	// Statements
	case *ExpressionStatement:
		if n.Expression != nil {
			Walk(v, n.Expression)
		}

	case *BlockStatement:
		for _, stmt := range n.Statements {
			Walk(v, stmt)
		}

	case *VarDeclStatement:
		for _, name := range n.Names {
			Walk(v, name)
		}
		if n.Type != nil {
			Walk(v, n.Type)
		}
		if n.Value != nil {
			Walk(v, n.Value)
		}

	case *AssignmentStatement:
		if n.Target != nil {
			Walk(v, n.Target)
		}
		if n.Value != nil {
			Walk(v, n.Value)
		}

	case *ConstDecl:
		Walk(v, n.Name)
		if n.Type != nil {
			Walk(v, n.Type)
		}
		Walk(v, n.Value)

	// Control Flow
	case *IfStatement:
		Walk(v, n.Condition)
		Walk(v, n.Consequence)
		if n.Alternative != nil {
			Walk(v, n.Alternative)
		}

	case *WhileStatement:
		Walk(v, n.Condition)
		Walk(v, n.Body)

	case *RepeatStatement:
		Walk(v, n.Body)
		Walk(v, n.Condition)

	case *ForStatement:
		Walk(v, n.Variable)
		Walk(v, n.Start)
		Walk(v, n.EndValue)
		if n.Step != nil {
			Walk(v, n.Step)
		}
		Walk(v, n.Body)

	case *ForInStatement:
		Walk(v, n.Variable)
		Walk(v, n.Collection)
		Walk(v, n.Body)

	case *CaseStatement:
		Walk(v, n.Expression)
		for _, branch := range n.Cases {
			// Walk branch values and statement (CaseBranch is not a Node)
			for _, value := range branch.Values {
				Walk(v, value)
			}
			Walk(v, branch.Statement)
		}
		if n.Else != nil {
			Walk(v, n.Else)
		}

	case *BreakStatement:
		// No children

	case *ContinueStatement:
		// No children

	case *ExitStatement:
		if n.ReturnValue != nil {
			Walk(v, n.ReturnValue)
		}

	// Functions
	case *FunctionDecl:
		Walk(v, n.Name)
		if n.ClassName != nil {
			Walk(v, n.ClassName)
		}
		for _, param := range n.Parameters {
			// Walk parameter children (Parameter is not a Node)
			if param.Name != nil {
				Walk(v, param.Name)
			}
			if param.Type != nil {
				Walk(v, param.Type)
			}
			if param.DefaultValue != nil {
				Walk(v, param.DefaultValue)
			}
		}
		if n.ReturnType != nil {
			Walk(v, n.ReturnType)
		}
		if n.PreConditions != nil {
			Walk(v, n.PreConditions)
		}
		if n.Body != nil {
			Walk(v, n.Body)
		}
		if n.PostConditions != nil {
			Walk(v, n.PostConditions)
		}

	case *ReturnStatement:
		if n.ReturnValue != nil {
			Walk(v, n.ReturnValue)
		}

	// Contracts
	case *Condition:
		if n.Test != nil {
			Walk(v, n.Test)
		}
		if n.Message != nil {
			Walk(v, n.Message)
		}

	case *PreConditions:
		for _, cond := range n.Conditions {
			Walk(v, cond)
		}

	case *PostConditions:
		for _, cond := range n.Conditions {
			Walk(v, cond)
		}

	// Arrays
	case *ArrayDecl:
		Walk(v, n.Name)
		if n.ArrayType != nil {
			Walk(v, n.ArrayType)
		}

	case *ArrayTypeAnnotation:
		if n.LowBound != nil {
			Walk(v, n.LowBound)
		}
		if n.HighBound != nil {
			Walk(v, n.HighBound)
		}
		if n.ElementType != nil {
			Walk(v, n.ElementType)
		}

	case *ArrayLiteralExpression:
		for _, elem := range n.Elements {
			Walk(v, elem)
		}
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *IndexExpression:
		Walk(v, n.Left)
		Walk(v, n.Index)
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *NewArrayExpression:
		if n.ElementTypeName != nil {
			Walk(v, n.ElementTypeName)
		}
		for _, dim := range n.Dimensions {
			Walk(v, dim)
		}
		if n.Type != nil {
			Walk(v, n.Type)
		}

	// Classes
	case *ClassDecl:
		Walk(v, n.Name)
		if n.Parent != nil {
			Walk(v, n.Parent)
		}
		for _, iface := range n.Interfaces {
			Walk(v, iface)
		}
		for _, field := range n.Fields {
			Walk(v, field)
		}
		for _, method := range n.Methods {
			Walk(v, method)
		}
		for _, operator := range n.Operators {
			Walk(v, operator)
		}
		for _, prop := range n.Properties {
			Walk(v, prop)
		}
		if n.Constructor != nil {
			Walk(v, n.Constructor)
		}
		if n.Destructor != nil {
			Walk(v, n.Destructor)
		}

	case *FieldDecl:
		Walk(v, n.Name)
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *NewExpression:
		Walk(v, n.ClassName)
		for _, arg := range n.Arguments {
			Walk(v, arg)
		}
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *MemberAccessExpression:
		Walk(v, n.Object)
		Walk(v, n.Member)
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *MethodCallExpression:
		Walk(v, n.Object)
		Walk(v, n.Method)
		for _, arg := range n.Arguments {
			Walk(v, arg)
		}
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *InheritedExpression:
		if n.Method != nil {
			Walk(v, n.Method)
		}
		for _, arg := range n.Arguments {
			Walk(v, arg)
		}
		if n.Type != nil {
			Walk(v, n.Type)
		}

	// Enums
	case *EnumDecl:
		Walk(v, n.Name)

	case *EnumLiteral:
		// No children to walk

	// Records
	case *RecordDecl:
		Walk(v, n.Name)
		for _, field := range n.Fields {
			Walk(v, field)
		}
		for _, method := range n.Methods {
			Walk(v, method)
		}

	case *RecordLiteralExpression:
		if n.TypeName != nil {
			Walk(v, n.TypeName)
		}
		for _, field := range n.Fields {
			// Walk field initializer children (FieldInitializer is not a Node)
			if field.Name != nil {
				Walk(v, field.Name)
			}
			Walk(v, field.Value)
		}

	// Lambda
	case *LambdaExpression:
		for _, param := range n.Parameters {
			// Walk parameter children (Parameter is not a Node)
			if param.Name != nil {
				Walk(v, param.Name)
			}
			if param.Type != nil {
				Walk(v, param.Type)
			}
			if param.DefaultValue != nil {
				Walk(v, param.DefaultValue)
			}
		}
		if n.ReturnType != nil {
			Walk(v, n.ReturnType)
		}
		if n.Body != nil {
			Walk(v, n.Body)
		}
		if n.Type != nil {
			Walk(v, n.Type)
		}

	// Sets
	case *SetDecl:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		if n.ElementType != nil {
			Walk(v, n.ElementType)
		}

	case *SetLiteral:
		for _, elem := range n.Elements {
			Walk(v, elem)
		}
		if n.Type != nil {
			Walk(v, n.Type)
		}

	// Properties
	case *PropertyDecl:
		Walk(v, n.Name)
		if n.Type != nil {
			Walk(v, n.Type)
		}
		for _, param := range n.IndexParams {
			// Walk parameter children (Parameter is not a Node)
			if param.Name != nil {
				Walk(v, param.Name)
			}
			if param.Type != nil {
				Walk(v, param.Type)
			}
			if param.DefaultValue != nil {
				Walk(v, param.DefaultValue)
			}
		}
		if n.ReadSpec != nil {
			Walk(v, n.ReadSpec)
		}
		if n.WriteSpec != nil {
			Walk(v, n.WriteSpec)
		}

	// Operators
	case *OperatorDecl:
		if n.Binding != nil {
			Walk(v, n.Binding)
		}
		for _, operandType := range n.OperandTypes {
			Walk(v, operandType)
		}
		if n.ReturnType != nil {
			Walk(v, n.ReturnType)
		}

	// Exceptions
	case *TryStatement:
		if n.TryBlock != nil {
			Walk(v, n.TryBlock)
		}
		if n.ExceptClause != nil {
			// Walk except clause children (ExceptClause is not a Node)
			for _, handler := range n.ExceptClause.Handlers {
				// Walk exception handler children (ExceptionHandler is not a Node)
				if handler.Variable != nil {
					Walk(v, handler.Variable)
				}
				if handler.ExceptionType != nil {
					Walk(v, handler.ExceptionType)
				}
				if handler.Statement != nil {
					Walk(v, handler.Statement)
				}
			}
			if n.ExceptClause.ElseBlock != nil {
				Walk(v, n.ExceptClause.ElseBlock)
			}
		}
		if n.FinallyClause != nil {
			// Walk finally clause children (FinallyClause is not a Node)
			if n.FinallyClause.Block != nil {
				Walk(v, n.FinallyClause.Block)
			}
		}

	case *RaiseStatement:
		if n.Exception != nil {
			Walk(v, n.Exception)
		}

	// Interfaces
	case *InterfaceDecl:
		Walk(v, n.Name)
		if n.Parent != nil {
			Walk(v, n.Parent)
		}
		for _, method := range n.Methods {
			// Walk interface method children (InterfaceMethodDecl is not a Node)
			Walk(v, method.Name)
			for _, param := range method.Parameters {
				// Walk parameter children (Parameter is not a Node)
				if param.Name != nil {
					Walk(v, param.Name)
				}
				if param.Type != nil {
					Walk(v, param.Type)
				}
				if param.DefaultValue != nil {
					Walk(v, param.DefaultValue)
				}
			}
			if method.ReturnType != nil {
				Walk(v, method.ReturnType)
			}
		}

	// Units
	case *UnitDeclaration:
		Walk(v, n.Name)
		if n.InterfaceSection != nil {
			Walk(v, n.InterfaceSection)
		}
		if n.ImplementationSection != nil {
			Walk(v, n.ImplementationSection)
		}
		if n.InitSection != nil {
			Walk(v, n.InitSection)
		}
		if n.FinalSection != nil {
			Walk(v, n.FinalSection)
		}

	case *UsesClause:
		for _, unit := range n.Units {
			Walk(v, unit)
		}

	// Type Annotations and Expressions
	case *TypeAnnotation:
		if n.InlineType != nil {
			Walk(v, n.InlineType)
		}

	case *TypeDeclaration:
		Walk(v, n.Name)
		if n.AliasedType != nil {
			Walk(v, n.AliasedType)
		}
		if n.LowBound != nil {
			Walk(v, n.LowBound)
		}
		if n.HighBound != nil {
			Walk(v, n.HighBound)
		}
		if n.FunctionPointerType != nil {
			Walk(v, n.FunctionPointerType)
		}

	case *FunctionPointerTypeNode:
		for _, param := range n.Parameters {
			// Walk parameter children (Parameter is not a Node)
			if param.Name != nil {
				Walk(v, param.Name)
			}
			if param.Type != nil {
				Walk(v, param.Type)
			}
			if param.DefaultValue != nil {
				Walk(v, param.DefaultValue)
			}
		}
		if n.ReturnType != nil {
			Walk(v, n.ReturnType)
		}

	case *AddressOfExpression:
		if n.Operator != nil {
			Walk(v, n.Operator)
		}
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *ArrayTypeNode:
		if n.LowBound != nil {
			Walk(v, n.LowBound)
		}
		if n.HighBound != nil {
			Walk(v, n.HighBound)
		}
		if n.ElementType != nil {
			Walk(v, n.ElementType)
		}

	case *SetTypeNode:
		if n.ElementType != nil {
			Walk(v, n.ElementType)
		}
	}
}

// Inspect traverses an AST in depth-first order, calling f for each node.
// If f returns false, traversal of that node's children is skipped.
// Otherwise, Inspect is called recursively for each child.
//
// This is a convenience wrapper around Walk for simple inspection tasks.
func Inspect(node Node, f func(Node) bool) {
	Walk(inspector(f), node)
}

// inspector is a helper type that implements Visitor for the Inspect function.
type inspector func(Node) bool

func (f inspector) Visit(node Node) Visitor {
	if f(node) {
		return f
	}
	return nil
}
