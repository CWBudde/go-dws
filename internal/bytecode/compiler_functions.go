package bytecode

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
)

func (c *Compiler) compileLambdaExpression(expr *ast.LambdaExpression) error {
	if expr == nil {
		return c.errorf(nil, "nil lambda expression")
	}

	name := fmt.Sprintf("lambda@%d", lineOf(expr))
	child := c.newChildCompiler(name)
	child.beginScope()

	for _, param := range expr.Parameters {
		if param == nil || param.Name == nil {
			return c.errorf(expr, "lambda parameter missing identifier")
		}
		paramType := typeFromAnnotation(param.Type)
		if _, err := child.declareLocal(param.Name, paramType); err != nil {
			return err
		}
	}

	if expr.Body != nil {
		if err := child.compileBlock(expr.Body); err != nil {
			return err
		}
	} else {
		child.chunk.WriteSimple(OpLoadNil, lineOf(expr))
		child.chunk.Write(OpReturn, 1, 0, lineOf(expr))
	}

	child.endScope()
	child.chunk.LocalCount = int(child.maxSlot)
	child.ensureFunctionReturn(lineOf(expr))
	child.chunk.Optimize()

	fn := NewFunctionObject(name, child.chunk, len(expr.Parameters))
	fn.UpvalueDefs = child.buildUpvalueDefs()

	fnIndex := c.chunk.AddConstant(FunctionValue(fn))
	if fnIndex > 0xFFFF {
		return c.errorf(expr, "constant pool overflow")
	}

	upvalueCount := len(fn.UpvalueDefs)
	if upvalueCount > 0xFF {
		return c.errorf(expr, "too many upvalues in lambda (max 255)")
	}

	c.chunk.Write(OpClosure, byte(upvalueCount), uint16(fnIndex), lineOf(expr))
	return nil
}

func (c *Compiler) compileCallExpression(expr *ast.CallExpression) error {
	argCount := len(expr.Arguments)
	if argCount > 0xFF {
		return c.errorf(expr, "too many arguments in function call: %d", argCount)
	}

	if ident, ok := expr.Function.(*ast.Identifier); ok {
		if info, ok := c.directCallInfo(ident); ok {
			for _, arg := range expr.Arguments {
				if err := c.compileExpression(arg); err != nil {
					return err
				}
			}
			c.chunk.Write(OpCall, byte(argCount), info.constIndex, lineOf(expr))
			return nil
		}
	}

	if err := c.compileExpression(expr.Function); err != nil {
		return err
	}

	for _, arg := range expr.Arguments {
		if err := c.compileExpression(arg); err != nil {
			return err
		}
	}

	c.chunk.Write(OpCallIndirect, byte(argCount), 0, lineOf(expr))
	return nil
}

func (c *Compiler) directCallInfo(ident *ast.Identifier) (functionInfo, bool) {
	if ident == nil || c.functions == nil {
		return functionInfo{}, false
	}

	if _, ok := c.resolveLocal(ident.Value); ok {
		return functionInfo{}, false
	}
	if c.hasEnclosingLocal(ident.Value) {
		return functionInfo{}, false
	}

	info, ok := c.functions[strings.ToLower(ident.Value)]
	return info, ok
}
