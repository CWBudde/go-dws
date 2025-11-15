package bytecode

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

func (c *Compiler) compileProgram(program *ast.Program) error {
	for _, stmt := range program.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileStatement(stmt ast.Statement) error {
	if stmt == nil {
		return nil
	}

	c.lastLine = lineOf(stmt)

	switch node := stmt.(type) {
	case *ast.BlockStatement:
		return c.compileBlock(node)
	case *ast.VarDeclStatement:
		return c.compileVarDecl(node)
	case *ast.AssignmentStatement:
		return c.compileAssignment(node)
	case *ast.ExpressionStatement:
		return c.compileExpressionStatement(node)
	case *ast.FunctionDecl:
		return c.compileFunctionDecl(node)
	case *ast.IfStatement:
		return c.compileIf(node)
	case *ast.WhileStatement:
		return c.compileWhile(node)
	case *ast.RepeatStatement:
		return c.compileRepeat(node)
	case *ast.TryStatement:
		return c.compileTryStatement(node)
	case *ast.RaiseStatement:
		return c.compileRaiseStatement(node)
	case *ast.ReturnStatement:
		return c.compileReturn(node)
	case *ast.BreakStatement:
		return c.compileBreak(node)
	case *ast.ContinueStatement:
		return c.compileContinue(node)
	case *ast.HelperDecl:
		return c.compileHelperDecl(node)
	// Task 9.7: Type declarations don't generate bytecode - they're handled by semantic analysis
	case *ast.RecordDecl:
		return nil // No bytecode needed for type declarations
	case *ast.ClassDecl:
		return nil // No bytecode needed for type declarations
	case *ast.EnumDecl:
		return nil // No bytecode needed for type declarations
	default:
		return c.errorf(stmt, "unsupported statement type %T", stmt)
	}
}

func (c *Compiler) compileBlock(block *ast.BlockStatement) error {
	c.beginScope()
	for _, stmt := range block.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}
	c.endScope()
	return nil
}

func (c *Compiler) compileBlockStatements(block *ast.BlockStatement) error {
	if block == nil {
		return nil
	}
	for _, stmt := range block.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileVarDecl(stmt *ast.VarDeclStatement) error {
	if len(stmt.Names) == 0 {
		return c.errorf(stmt, "variable declaration without names")
	}

	for _, name := range stmt.Names {
		localType := typeFromAnnotation(stmt.Type)
		if localType == nil && stmt.Value != nil {
			localType = c.inferExpressionType(stmt.Value)
		}
		if c.isGlobalScope() {
			index, err := c.declareGlobal(name, localType)
			if err != nil {
				return err
			}
			if err := c.emitInitializer(stmt.Value, stmt); err != nil {
				return err
			}
			c.chunk.Write(OpStoreGlobal, 0, index, lineOf(name))
			continue
		}

		slot, err := c.declareLocal(name, localType)
		if err != nil {
			return err
		}

		if err := c.emitInitializer(stmt.Value, stmt); err != nil {
			return err
		}

		c.chunk.Write(OpStoreLocal, 0, slot, lineOf(name))
	}

	return nil
}

func (c *Compiler) emitInitializer(value ast.Expression, stmt ast.Statement) error {
	if value != nil {
		return c.compileExpression(value)
	}
	c.chunk.WriteSimple(OpLoadNil, lineOf(stmt))
	return nil
}

func (c *Compiler) compileAssignment(stmt *ast.AssignmentStatement) error {
	if stmt.Operator != lexer.ASSIGN {
		return c.errorf(stmt, "unsupported assignment operator %s", stmt.Operator)
	}

	switch target := stmt.Target.(type) {
	case *ast.Identifier:
		return c.compileIdentifierAssignment(target, stmt.Value)
	case *ast.MemberAccessExpression:
		return c.compileMemberAssignment(target, stmt.Value)
	case *ast.IndexExpression:
		return c.compileIndexAssignment(target, stmt.Value)
	default:
		return c.errorf(stmt.Target, "unsupported assignment target %T", stmt.Target)
	}
}

func (c *Compiler) compileExpressionStatement(stmt *ast.ExpressionStatement) error {
	if stmt.Expression == nil {
		return nil
	}

	if err := c.compileExpression(stmt.Expression); err != nil {
		return err
	}

	c.chunk.WriteSimple(OpPop, lineOf(stmt))
	return nil
}

func (c *Compiler) compileIf(stmt *ast.IfStatement) error {
	if err := c.compileExpression(stmt.Condition); err != nil {
		return err
	}

	jumpIfFalse := c.chunk.EmitJump(OpJumpIfFalse, lineOf(stmt.Condition))

	if err := c.compileStatement(stmt.Consequence); err != nil {
		return err
	}

	if stmt.Alternative != nil {
		jumpToEnd := c.chunk.EmitJump(OpJump, lineOf(stmt))
		if err := c.chunk.PatchJump(jumpIfFalse); err != nil {
			return err
		}

		if err := c.compileStatement(stmt.Alternative); err != nil {
			return err
		}

		return c.chunk.PatchJump(jumpToEnd)
	}

	return c.chunk.PatchJump(jumpIfFalse)
}

func (c *Compiler) compileWhile(stmt *ast.WhileStatement) error {
	loopStart := len(c.chunk.Code)
	ctx := c.pushLoop(loopKindWhile, loopStart)
	defer c.popLoop()

	if err := c.compileExpression(stmt.Condition); err != nil {
		return err
	}

	exitJump := c.chunk.EmitJump(OpJumpIfFalse, lineOf(stmt.Condition))

	if err := c.compileStatement(stmt.Body); err != nil {
		return err
	}

	if err := c.chunk.EmitLoop(loopStart, lineOf(stmt)); err != nil {
		return err
	}

	if err := c.chunk.PatchJump(exitJump); err != nil {
		return err
	}

	return c.patchLoopBreaks(ctx)
}

func (c *Compiler) compileRepeat(stmt *ast.RepeatStatement) error {
	loopStart := len(c.chunk.Code)
	ctx := c.pushLoop(loopKindRepeat, loopStart)
	defer c.popLoop()

	if err := c.compileStatement(stmt.Body); err != nil {
		return err
	}

	conditionStart := len(c.chunk.Code)
	if err := c.patchLoopContinues(ctx, conditionStart); err != nil {
		return err
	}

	if err := c.compileExpression(stmt.Condition); err != nil {
		return err
	}

	exitJump := c.chunk.EmitJump(OpJumpIfTrue, lineOf(stmt.Condition))

	if err := c.chunk.EmitLoop(loopStart, lineOf(stmt)); err != nil {
		return err
	}

	if err := c.chunk.PatchJump(exitJump); err != nil {
		return err
	}

	return c.patchLoopBreaks(ctx)
}

func (c *Compiler) compileTryStatement(stmt *ast.TryStatement) error {
	if stmt == nil || stmt.TryBlock == nil {
		return c.errorf(stmt, "invalid try statement")
	}

	hasExcept := stmt.ExceptClause != nil
	line := lineOf(stmt)
	tryInst := c.chunk.Write(OpTry, 0, 0, line)

	if err := c.compileBlockStatements(stmt.TryBlock); err != nil {
		return err
	}

	jumpAfterTry := c.chunk.EmitJump(OpJump, line)

	catchStart := -1
	afterCatchJump := -1
	if hasExcept {
		catchStart = len(c.chunk.Code)
		c.chunk.Write(OpCatch, 0, 0, line)
		if err := c.compileExceptClause(stmt.ExceptClause); err != nil {
			return err
		}
		afterCatchJump = c.chunk.EmitJump(OpJump, line)
	}

	finallyStart := len(c.chunk.Code)
	c.chunk.Write(OpFinally, 0, 0, line)
	if stmt.FinallyClause != nil {
		if err := c.compileBlockStatements(stmt.FinallyClause.Block); err != nil {
			return err
		}
	}
	c.chunk.Write(OpFinally, 1, 0, line)

	if err := c.patchJumpToTarget(jumpAfterTry, finallyStart); err != nil {
		return err
	}
	if hasExcept {
		if err := c.patchJumpToTarget(afterCatchJump, finallyStart); err != nil {
			return err
		}
		if err := c.patchJumpToTarget(catchStart, finallyStart); err != nil {
			return err
		}
	}

	target := finallyStart
	if hasExcept {
		target = catchStart
	}
	if err := c.patchJumpToTarget(tryInst, target); err != nil {
		return err
	}

	info := TryInfo{
		CatchTarget:   catchStart,
		FinallyTarget: finallyStart,
		HasCatch:      hasExcept,
		HasFinally:    true,
	}
	c.chunk.SetTryInfo(tryInst, info)

	return nil
}

func (c *Compiler) compileExceptClause(clause *ast.ExceptClause) error {
	if clause == nil {
		return nil
	}

	c.beginScope()
	defer c.endScope()

	tmpSlot, err := c.declareSyntheticLocal("$exception")
	if err != nil {
		return err
	}
	line := lineOfExceptClause(clause)
	if line == 0 {
		line = c.lastLine
	}
	c.chunk.Write(OpStoreLocal, 0, tmpSlot, line)

	endJumps := make([]int, 0, len(clause.Handlers)+1)

	for _, handler := range clause.Handlers {
		if handler == nil {
			continue
		}
		handlerLine := lineOf(handler.Statement)
		if handlerLine == 0 {
			handlerLine = line
		}

		jumpIfNoMatch := -1
		if handler.ExceptionType != nil {
			typeConst := c.chunk.AddConstant(StringValue(handler.ExceptionType.Name))
			c.chunk.Write(OpLoadLocal, 0, tmpSlot, handlerLine)
			c.chunk.WriteSimple(OpGetClass, handlerLine)
			c.chunk.Write(OpLoadConst, 0, uint16(typeConst), handlerLine)
			c.chunk.WriteSimple(OpEqual, handlerLine)
			jumpIfNoMatch = c.chunk.EmitJump(OpJumpIfFalse, handlerLine)
		}

		beginHandlerScope := false
		if handler.Variable != nil {
			beginHandlerScope = true
			c.beginScope()
			slot, err := c.declareLocal(handler.Variable, typeFromAnnotation(handler.ExceptionType))
			if err != nil {
				return err
			}
			c.chunk.Write(OpLoadLocal, 0, tmpSlot, handlerLine)
			c.chunk.Write(OpStoreLocal, 0, slot, handlerLine)
		}

		if err := c.compileStatement(handler.Statement); err != nil {
			return err
		}

		if beginHandlerScope {
			c.endScope()
		}

		endJumps = append(endJumps, c.chunk.EmitJump(OpJump, handlerLine))
		if jumpIfNoMatch >= 0 {
			if err := c.patchJumpToTarget(jumpIfNoMatch, len(c.chunk.Code)); err != nil {
				return err
			}
		}
	}

	if clause.ElseBlock != nil {
		if err := c.compileBlock(clause.ElseBlock); err != nil {
			return err
		}
	} else {
		c.chunk.Write(OpLoadLocal, 0, tmpSlot, line)
		c.chunk.WriteSimple(OpThrow, line)
	}

	endTarget := len(c.chunk.Code)
	for _, jump := range endJumps {
		if err := c.patchJumpToTarget(jump, endTarget); err != nil {
			return err
		}
	}

	return nil
}

func (c *Compiler) compileRaiseStatement(stmt *ast.RaiseStatement) error {
	if stmt == nil {
		return c.errorf(nil, "invalid raise statement")
	}
	line := lineOf(stmt)
	if stmt.Exception == nil {
		if err := c.loadExceptObject(line); err != nil {
			return err
		}
	} else {
		if err := c.compileExpression(stmt.Exception); err != nil {
			return err
		}
	}
	c.chunk.WriteSimple(OpThrow, line)
	return nil
}

func (c *Compiler) compileReturn(stmt *ast.ReturnStatement) error {
	if stmt.ReturnValue != nil {
		if err := c.compileExpression(stmt.ReturnValue); err != nil {
			return err
		}
		c.chunk.Write(OpReturn, 1, 0, lineOf(stmt))
		return nil
	}

	c.chunk.Write(OpReturn, 0, 0, lineOf(stmt))
	return nil
}

func (c *Compiler) compileFunctionDecl(fn *ast.FunctionDecl) error {
	if fn.Name == nil {
		return c.errorf(fn, "function declaration missing name")
	}
	if !c.isGlobalScope() {
		return c.errorf(fn, "local function declarations are not supported yet")
	}

	globalSlot, err := c.declareGlobal(fn.Name, typeFromAnnotation(fn.ReturnType))
	if err != nil {
		return err
	}

	child := c.newChildCompiler(fn.Name.Value)
	child.beginScope()

	for _, param := range fn.Parameters {
		if param == nil || param.Name == nil {
			return c.errorf(fn, "function parameter missing identifier")
		}
		paramType := typeFromAnnotation(param.Type)
		if _, err := child.declareLocal(param.Name, paramType); err != nil {
			return err
		}
	}

	if fn.Body == nil {
		return c.errorf(fn, "function %s missing body", fn.Name.Value)
	}

	if err := child.compileBlock(fn.Body); err != nil {
		return err
	}

	child.endScope()
	child.chunk.LocalCount = int(child.maxSlot)
	child.ensureFunctionReturn(lineOf(fn))
	child.chunk.Optimize()

	functionObject := NewFunctionObject(fn.Name.Value, child.chunk, len(fn.Parameters))
	functionObject.UpvalueDefs = child.buildUpvalueDefs()

	fnConstIndex := c.chunk.AddConstant(FunctionValue(functionObject))
	if fnConstIndex > 0xFFFF {
		return c.errorf(fn, "constant pool overflow")
	}

	upvalueCount := len(functionObject.UpvalueDefs)
	if upvalueCount > 0xFF {
		return c.errorf(fn, "too many upvalues in function %s", fn.Name.Value)
	}

	info := functionInfo{
		constIndex: uint16(fnConstIndex),
		globalSlot: globalSlot,
		fn:         functionObject,
	}
	if c.functions == nil {
		c.functions = make(map[string]functionInfo)
	}
	c.functions[strings.ToLower(fn.Name.Value)] = info

	// Check if this is a helper method (ClassName is set and refers to a helper)
	if fn.ClassName != nil {
		helperKey := strings.ToLower(fn.ClassName.Value)
		if helper, ok := c.helpers[helperKey]; ok {
			// This is a helper method - associate it with the helper
			methodKey := strings.ToLower(fn.Name.Value)
			helper.Methods[methodKey] = globalSlot
		}
		// If ClassName is set but not a helper, it's a class method (handled elsewhere)
	}

	c.chunk.Write(OpClosure, byte(upvalueCount), uint16(fnConstIndex), lineOf(fn))
	c.chunk.Write(OpStoreGlobal, 0, globalSlot, lineOf(fn))
	return nil
}

func (c *Compiler) compileBreak(stmt *ast.BreakStatement) error {
	loop := c.currentLoop()
	if loop == nil {
		return c.errorf(stmt, "break outside of loop")
	}
	jumpIdx := c.chunk.EmitJump(OpJump, lineOf(stmt))
	loop.breakJumps = append(loop.breakJumps, jumpIdx)
	return nil
}

func (c *Compiler) compileContinue(stmt *ast.ContinueStatement) error {
	loop := c.currentLoop()
	if loop == nil {
		return c.errorf(stmt, "continue outside of loop")
	}

	switch loop.kind {
	case loopKindWhile:
		if err := c.chunk.EmitLoop(loop.loopStart, lineOf(stmt)); err != nil {
			return err
		}
	case loopKindRepeat:
		jumpIdx := c.chunk.EmitJump(OpJump, lineOf(stmt))
		loop.continueJumps = append(loop.continueJumps, jumpIdx)
	default:
		return c.errorf(stmt, "unsupported loop kind for continue")
	}

	return nil
}

func (c *Compiler) compileHelperDecl(decl *ast.HelperDecl) error {
	if decl == nil || decl.Name == nil {
		return c.errorf(decl, "invalid helper declaration")
	}

	if decl.ForType == nil {
		return c.errorf(decl, "helper '%s' missing target type", decl.Name.Value)
	}

	// Initialize helpers map if needed
	if c.helpers == nil {
		c.helpers = make(map[string]*HelperInfo)
	}

	// Create helper metadata
	helperInfo := &HelperInfo{
		Name:        decl.Name.Value,
		TargetType:  decl.ForType.Name,
		Methods:     make(map[string]uint16),
		Properties:  make([]string, 0),
		ClassVars:   make([]string, 0),
		ClassConsts: make(map[string]Value),
	}

	// Store parent helper name if present
	if decl.ParentHelper != nil {
		helperInfo.ParentHelper = decl.ParentHelper.Value
	}

	// Register property names
	for _, prop := range decl.Properties {
		if prop != nil && prop.Name != nil {
			helperInfo.Properties = append(helperInfo.Properties, prop.Name.Value)
		}
	}

	// Register class variable names
	for _, classVar := range decl.ClassVars {
		if classVar != nil && classVar.Name != nil {
			helperInfo.ClassVars = append(helperInfo.ClassVars, classVar.Name.Value)
		}
	}

	// Compile and register class constants
	for _, classConst := range decl.ClassConsts {
		if classConst == nil || classConst.Name == nil {
			continue
		}
		// Try to evaluate the constant at compile time
		if classConst.Value != nil {
			if value, ok := literalValue(classConst.Value); ok {
				helperInfo.ClassConsts[classConst.Name.Value] = value
			}
		}
	}

	// Note: Method implementations are separate FunctionDecl nodes with ClassName set.
	// They will be compiled via compileFunctionDecl and associated with this helper there.
	// We only store method declarations here for metadata.
	for _, method := range decl.Methods {
		if method != nil && method.Name != nil {
			// Method will be compiled separately, we'll update the slot mapping later
			helperInfo.Methods[strings.ToLower(method.Name.Value)] = 0 // Placeholder
		}
	}

	// Register the helper (case-insensitive key)
	key := strings.ToLower(decl.Name.Value)
	c.helpers[key] = helperInfo

	// Helper declaration itself doesn't generate runtime bytecode
	// The actual method implementations (separate FunctionDecl nodes) will generate bytecode
	return nil
}
