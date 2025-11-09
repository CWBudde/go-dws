package bytecode

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
)

// Compiler converts AST nodes into bytecode chunks.
type Compiler struct {
	functions       map[string]functionInfo
	enclosing       *Compiler
	globals         map[string]globalVar
	chunk           *Chunk
	upvalues        []upvalue
	loopStack       []*loopContext
	locals          []local
	optimizeOptions []OptimizeOption
	scopeDepth      int
	lastLine        int
	nextSlot        uint16
	maxSlot         uint16
	nextGlobal      uint16
}

type local struct {
	typ   types.Type
	name  string
	depth int
	slot  uint16
}

type globalVar struct {
	typ   types.Type
	name  string
	index uint16
}

type upvalue struct {
	index   uint16
	isLocal bool
}

type functionInfo struct {
	fn         *FunctionObject
	constIndex uint16
	globalSlot uint16
}

type loopKind int

const (
	loopKindWhile loopKind = iota
	loopKindRepeat
)

type loopContext struct {
	breakJumps    []int
	continueJumps []int
	kind          loopKind
	loopStart     int
}

// CompilerOption configures a new compiler instance.
type CompilerOption func(*Compiler)

// NewCompiler creates a compiler for the given chunk name.
func NewCompiler(chunkName string, opts ...CompilerOption) *Compiler {
	return newCompiler(chunkName, nil, opts...)
}

func newCompiler(chunkName string, enclosing *Compiler, opts ...CompilerOption) *Compiler {
	globals := make(map[string]globalVar)
	functions := make(map[string]functionInfo)
	if enclosing != nil {
		globals = enclosing.globals
		functions = enclosing.functions
	}
	c := &Compiler{
		chunk:     NewChunk(chunkName),
		globals:   globals,
		functions: functions,
		enclosing: enclosing,
	}
	if enclosing != nil {
		c.optimizeOptions = enclosing.optimizeOptions
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Compiler) newChildCompiler(name string) *Compiler {
	return newCompiler(name, c)
}

// WithCompilerOptimizeOptions overrides the optimization passes used by this compiler.
func WithCompilerOptimizeOptions(opts ...OptimizeOption) CompilerOption {
	return func(c *Compiler) {
		c.optimizeOptions = opts
	}
}

// Compile compiles the provided program into bytecode.
func (c *Compiler) Compile(program *ast.Program) (*Chunk, error) {
	if program == nil {
		return nil, fmt.Errorf("bytecode compile error: nil program")
	}

	chunkName := c.chunk.Name
	c.chunk = NewChunk(chunkName)

	// Reset compiler state for reuse
	c.locals = c.locals[:0]
	c.upvalues = c.upvalues[:0]
	c.globals = make(map[string]globalVar)
	c.functions = make(map[string]functionInfo)
	c.loopStack = c.loopStack[:0]
	c.scopeDepth = 0
	c.nextSlot = 0
	c.maxSlot = 0
	c.nextGlobal = 0
	c.enclosing = nil

	c.initBuiltins()

	if err := c.compileProgram(program); err != nil {
		return nil, err
	}

	c.chunk.LocalCount = int(c.maxSlot)

	if c.needsHalt() {
		c.chunk.WriteSimple(OpHalt, c.lastLine)
	}

	c.chunk.Optimize(c.optimizeOptions...)

	return c.chunk, nil
}

// LocalCount returns the number of local slots allocated during compilation.
func (c *Compiler) LocalCount() int {
	return int(c.maxSlot)
}

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

func (c *Compiler) compileExpression(expr ast.Expression) error {
	if expr == nil {
		return nil
	}

	switch node := expr.(type) {
	case *ast.IntegerLiteral:
		return c.emitLoadConstant(IntValue(node.Value), lineOf(node))
	case *ast.FloatLiteral:
		return c.emitLoadConstant(FloatValue(node.Value), lineOf(node))
	case *ast.StringLiteral:
		return c.emitLoadConstant(StringValue(node.Value), lineOf(node))
	case *ast.BooleanLiteral:
		if node.Value {
			c.chunk.WriteSimple(OpLoadTrue, lineOf(node))
		} else {
			c.chunk.WriteSimple(OpLoadFalse, lineOf(node))
		}
		return nil
	case *ast.NilLiteral:
		c.chunk.WriteSimple(OpLoadNil, lineOf(node))
		return nil
	case *ast.Identifier:
		return c.compileIdentifier(node)
	case *ast.ArrayLiteralExpression:
		return c.compileArrayLiteral(node)
	case *ast.IndexExpression:
		return c.compileIndexExpression(node)
	case *ast.MemberAccessExpression:
		return c.compileMemberAccess(node)
	case *ast.NewExpression:
		return c.compileNewExpression(node)
	case *ast.LambdaExpression:
		return c.compileLambdaExpression(node)
	case *ast.MethodCallExpression:
		return c.compileMethodCallExpression(node)
	case *ast.CallExpression:
		return c.compileCallExpression(node)
	case *ast.BinaryExpression:
		return c.compileBinaryExpression(node)
	case *ast.UnaryExpression:
		return c.compileUnaryExpression(node)
	default:
		return c.errorf(expr, "unsupported expression type %T", expr)
	}
}

func (c *Compiler) compileIdentifier(ident *ast.Identifier) error {
	localInfo, ok := c.resolveLocal(ident.Value)
	if !ok {
		if uvIndex, ok, err := c.resolveUpvalue(ident.Value); err != nil {
			return err
		} else if ok {
			c.chunk.Write(OpLoadUpvalue, 0, uvIndex, lineOf(ident))
			return nil
		}
		if globalInfo, found := c.resolveGlobal(ident.Value); found {
			c.chunk.Write(OpLoadGlobal, 0, globalInfo.index, lineOf(ident))
			return nil
		}
		if strings.EqualFold(ident.Value, "Self") {
			c.chunk.WriteSimple(OpGetSelf, lineOf(ident))
			return nil
		}
		return c.errorf(ident, "unknown identifier %q", ident.Value)
	}

	c.chunk.Write(OpLoadLocal, 0, localInfo.slot, lineOf(ident))
	return nil
}

func (c *Compiler) compileIdentifierAssignment(ident *ast.Identifier, value ast.Expression) error {
	if err := c.compileExpression(value); err != nil {
		return err
	}

	if strings.EqualFold(ident.Value, builtinExceptObjectName) {
		return c.errorf(ident, "cannot assign to %s", builtinExceptObjectName)
	}

	if localInfo, ok := c.resolveLocal(ident.Value); ok {
		c.chunk.Write(OpStoreLocal, 0, localInfo.slot, lineOf(ident))
		return nil
	}

	if uvIndex, ok, err := c.resolveUpvalue(ident.Value); err != nil {
		return err
	} else if ok {
		c.chunk.Write(OpStoreUpvalue, 0, uvIndex, lineOf(ident))
		return nil
	}

	if globalInfo, ok := c.resolveGlobal(ident.Value); ok {
		c.chunk.Write(OpStoreGlobal, 0, globalInfo.index, lineOf(ident))
		return nil
	}

	if strings.EqualFold(ident.Value, "Self") {
		return c.errorf(ident, "cannot assign to Self")
	}

	return c.errorf(ident, "unknown variable %q", ident.Value)
}

func (c *Compiler) compileMemberAccess(expr *ast.MemberAccessExpression) error {
	if expr == nil || expr.Member == nil {
		return c.errorf(expr, "invalid member access expression")
	}

	if err := c.compileExpression(expr.Object); err != nil {
		return err
	}

	nameIndex, err := c.propertyNameIndex(expr.Member.Value, expr)
	if err != nil {
		return err
	}

	c.chunk.Write(OpGetProperty, 0, nameIndex, lineOf(expr))
	return nil
}

func (c *Compiler) compileMemberAssignment(target *ast.MemberAccessExpression, value ast.Expression) error {
	if target == nil || target.Member == nil {
		return c.errorf(target, "invalid member assignment target")
	}

	if err := c.compileExpression(target.Object); err != nil {
		return err
	}

	if err := c.compileExpression(value); err != nil {
		return err
	}

	nameIndex, err := c.propertyNameIndex(target.Member.Value, target)
	if err != nil {
		return err
	}

	c.chunk.Write(OpSetProperty, 0, nameIndex, lineOf(target))
	return nil
}

func (c *Compiler) compileArrayLiteral(expr *ast.ArrayLiteralExpression) error {
	if expr == nil {
		return c.errorf(nil, "nil array literal expression")
	}
	for _, element := range expr.Elements {
		if err := c.compileExpression(element); err != nil {
			return err
		}
	}
	count := len(expr.Elements)
	if count > 0xFFFF {
		return c.errorf(expr, "array literal has too many elements (%d)", count)
	}
	c.chunk.Write(OpNewArray, 0, uint16(count), lineOf(expr))
	return nil
}

func (c *Compiler) compileIndexExpression(expr *ast.IndexExpression) error {
	if expr == nil {
		return c.errorf(nil, "nil index expression")
	}
	if err := c.compileExpression(expr.Left); err != nil {
		return err
	}
	if err := c.compileExpression(expr.Index); err != nil {
		return err
	}
	c.chunk.WriteSimple(OpArrayGet, lineOf(expr))
	return nil
}

func (c *Compiler) compileIndexAssignment(target *ast.IndexExpression, value ast.Expression) error {
	if target == nil {
		return c.errorf(nil, "nil index assignment target")
	}
	if value != nil {
		if err := c.compileExpression(value); err != nil {
			return err
		}
	} else {
		c.chunk.WriteSimple(OpLoadNil, lineOf(target))
	}
	if err := c.compileExpression(target.Left); err != nil {
		return err
	}
	if err := c.compileExpression(target.Index); err != nil {
		return err
	}
	c.chunk.WriteSimple(OpRotate3, lineOf(target))
	c.chunk.WriteSimple(OpArraySet, lineOf(target))
	return nil
}

func (c *Compiler) compileNewExpression(expr *ast.NewExpression) error {
	if expr == nil || expr.ClassName == nil {
		return c.errorf(expr, "invalid new expression")
	}
	if len(expr.Arguments) > 0 {
		return c.errorf(expr, "constructors with arguments are not supported in bytecode yet")
	}
	constIdx := c.chunk.AddConstant(StringValue(expr.ClassName.Value))
	if constIdx > 0xFFFF {
		return c.errorf(expr, "constant pool overflow")
	}
	c.chunk.Write(OpNewObject, 0, uint16(constIdx), lineOf(expr))
	return nil
}

func (c *Compiler) compileMethodCallExpression(expr *ast.MethodCallExpression) error {
	if expr == nil || expr.Method == nil {
		return c.errorf(expr, "invalid method call expression")
	}

	if err := c.compileExpression(expr.Object); err != nil {
		return err
	}

	for _, arg := range expr.Arguments {
		if err := c.compileExpression(arg); err != nil {
			return err
		}
	}

	argCount := len(expr.Arguments)
	if argCount > 0xFF {
		return c.errorf(expr, "too many arguments in method call: %d", argCount)
	}

	nameIndex, err := c.propertyNameIndex(expr.Method.Value, expr)
	if err != nil {
		return err
	}

	c.chunk.Write(OpCallMethod, byte(argCount), nameIndex, lineOf(expr))
	return nil
}

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

func (c *Compiler) compileBinaryExpression(expr *ast.BinaryExpression) error {
	if folded, err := c.tryFoldBinaryExpression(expr); folded {
		return err
	}

	if err := c.compileExpression(expr.Left); err != nil {
		return err
	}
	if err := c.compileExpression(expr.Right); err != nil {
		return err
	}

	resultType := c.inferExpressionType(expr)
	if resultType == nil {
		resultType = c.inferExpressionType(expr.Left)
	}
	if resultType == nil {
		resultType = c.inferExpressionType(expr.Right)
	}

	line := lineOf(expr)
	switch strings.ToLower(expr.Operator) {
	case "+":
		return c.emitNumericBinaryOp(resultType, line, OpAddInt, OpAddFloat, OpStringConcat)
	case "-":
		return c.emitNumericBinaryOp(resultType, line, OpSubInt, OpSubFloat, 0)
	case "*":
		return c.emitNumericBinaryOp(resultType, line, OpMulInt, OpMulFloat, 0)
	case "div":
		return c.emitNumericBinaryOp(resultType, line, OpDivInt, 0, 0)
	case "/":
		return c.emitNumericBinaryOp(resultType, line, 0, OpDivFloat, 0)
	case "mod":
		return c.emitNumericBinaryOp(resultType, line, OpModInt, 0, 0)
	case "=":
		c.chunk.WriteSimple(OpEqual, line)
	case "<>":
		c.chunk.WriteSimple(OpNotEqual, line)
	case "<":
		c.chunk.WriteSimple(OpLess, line)
	case "<=":
		c.chunk.WriteSimple(OpLessEqual, line)
	case ">":
		c.chunk.WriteSimple(OpGreater, line)
	case ">=":
		c.chunk.WriteSimple(OpGreaterEqual, line)
	case "and":
		c.chunk.WriteSimple(OpAnd, line)
	case "or":
		c.chunk.WriteSimple(OpOr, line)
	default:
		return c.errorf(expr, "unsupported binary operator %q", expr.Operator)
	}

	return nil
}

func (c *Compiler) compileUnaryExpression(expr *ast.UnaryExpression) error {
	if folded, err := c.tryFoldUnaryExpression(expr); folded {
		return err
	}

	if err := c.compileExpression(expr.Right); err != nil {
		return err
	}

	exprType := c.inferExpressionType(expr)
	if exprType == nil {
		exprType = c.inferExpressionType(expr.Right)
	}

	line := lineOf(expr)
	switch strings.ToLower(expr.Operator) {
	case "+":
		return nil
	case "-":
		if isFloatType(exprType) {
			c.chunk.WriteSimple(OpNegateFloat, line)
		} else {
			c.chunk.WriteSimple(OpNegateInt, line)
		}
	case "not":
		c.chunk.WriteSimple(OpNot, line)
	default:
		return c.errorf(expr, "unsupported unary operator %q", expr.Operator)
	}

	return nil
}

func (c *Compiler) tryFoldBinaryExpression(expr *ast.BinaryExpression) (bool, error) {
	leftVal, leftOk := literalValue(expr.Left)
	rightVal, rightOk := literalValue(expr.Right)
	if !leftOk || !rightOk {
		return false, nil
	}

	result, ok := evaluateBinary(expr.Operator, leftVal, rightVal)
	if !ok {
		return false, nil
	}

	return true, c.emitValue(result, lineOf(expr))
}

func (c *Compiler) tryFoldUnaryExpression(expr *ast.UnaryExpression) (bool, error) {
	operandVal, ok := literalValue(expr.Right)
	if !ok {
		return false, nil
	}

	result, ok := evaluateUnary(expr.Operator, operandVal)
	if !ok {
		return false, nil
	}

	return true, c.emitValue(result, lineOf(expr))
}

func (c *Compiler) emitNumericBinaryOp(resultType types.Type, line int, intOp, floatOp OpCode, stringOp OpCode) error {
	switch {
	case stringOp != 0 && isStringType(resultType):
		c.chunk.WriteSimple(stringOp, line)
	case isFloatType(resultType):
		if floatOp == 0 {
			return c.errorf(nil, "no float opcode available for operation")
		}
		c.chunk.WriteSimple(floatOp, line)
	default:
		if intOp == 0 {
			return c.errorf(nil, "no integer opcode available for operation")
		}
		c.chunk.WriteSimple(intOp, line)
	}
	return nil
}

func (c *Compiler) emitLoadConstant(value Value, line int) error {
	index := c.chunk.AddConstant(value)
	if index > 0xFFFF {
		return c.errorf(nil, "constant pool overflow")
	}

	switch index {
	case 0:
		c.chunk.WriteSimple(OpLoadConst0, line)
	case 1:
		c.chunk.WriteSimple(OpLoadConst1, line)
	default:
		c.chunk.Write(OpLoadConst, 0, uint16(index), line)
	}

	return nil
}

func (c *Compiler) emitValue(value Value, line int) error {
	switch value.Type {
	case ValueNil:
		c.chunk.WriteSimple(OpLoadNil, line)
		return nil
	case ValueBool:
		if value.AsBool() {
			c.chunk.WriteSimple(OpLoadTrue, line)
		} else {
			c.chunk.WriteSimple(OpLoadFalse, line)
		}
		return nil
	default:
		return c.emitLoadConstant(value, line)
	}
}

func (c *Compiler) declareLocal(ident *ast.Identifier, typ types.Type) (uint16, error) {
	if _, exists := c.resolveLocalInCurrentScope(ident.Value); exists {
		return 0, c.errorf(ident, "duplicate variable %q in current scope", ident.Value)
	}

	slot := c.nextSlot
	c.nextSlot++
	if slot+1 > c.maxSlot {
		c.maxSlot = slot + 1
	}

	c.locals = append(c.locals, local{
		name:  ident.Value,
		depth: c.scopeDepth,
		slot:  slot,
		typ:   typ,
	})

	return slot, nil
}

func (c *Compiler) declareSyntheticLocal(name string) (uint16, error) {
	ident := &ast.Identifier{Value: name}
	return c.declareLocal(ident, nil)
}

func (c *Compiler) declareGlobal(ident *ast.Identifier, typ types.Type) (uint16, error) {
	if c.globals == nil {
		c.globals = make(map[string]globalVar)
	}

	key := strings.ToLower(ident.Value)
	if _, exists := c.globals[key]; exists {
		return 0, c.errorf(ident, "duplicate global variable %q", ident.Value)
	}

	index := c.nextGlobal
	c.nextGlobal++

	c.globals[key] = globalVar{
		name:  ident.Value,
		index: index,
		typ:   typ,
	}

	return index, nil
}

func (c *Compiler) resolveLocal(name string) (local, bool) {
	for i := len(c.locals) - 1; i >= 0; i-- {
		if strings.EqualFold(c.locals[i].name, name) {
			return c.locals[i], true
		}
	}
	return local{}, false
}

func (c *Compiler) resolveLocalInCurrentScope(name string) (local, bool) {
	for i := len(c.locals) - 1; i >= 0; i-- {
		loc := c.locals[i]
		if loc.depth != c.scopeDepth {
			break
		}
		if strings.EqualFold(loc.name, name) {
			return loc, true
		}
	}
	return local{}, false
}

func (c *Compiler) resolveGlobal(name string) (globalVar, bool) {
	if c.globals == nil {
		return globalVar{}, false
	}
	g, ok := c.globals[strings.ToLower(name)]
	return g, ok
}

func (c *Compiler) resolveUpvalue(name string) (uint16, bool, error) {
	if c.enclosing == nil {
		return 0, false, nil
	}

	if localInfo, ok := c.enclosing.resolveLocal(name); ok {
		return c.addUpvalue(localInfo.slot, true)
	}

	upvalueIndex, ok, err := c.enclosing.resolveUpvalue(name)
	if err != nil || !ok {
		return 0, ok, err
	}
	return c.addUpvalue(upvalueIndex, false)
}

func (c *Compiler) addUpvalue(index uint16, isLocal bool) (uint16, bool, error) {
	for i, uv := range c.upvalues {
		if uv.index == index && uv.isLocal == isLocal {
			return uint16(i), true, nil
		}
	}

	if len(c.upvalues) >= 0xFF {
		return 0, false, c.errorf(nil, "too many upvalues (max 255)")
	}

	c.upvalues = append(c.upvalues, upvalue{
		index:   index,
		isLocal: isLocal,
	})

	return uint16(len(c.upvalues) - 1), true, nil
}

func (c *Compiler) propertyNameIndex(name string, node ast.Node) (uint16, error) {
	index := c.chunk.AddConstant(StringValue(name))
	if index > 0xFFFF {
		return 0, c.errorf(node, "constant pool overflow")
	}
	return uint16(index), nil
}

func (c *Compiler) pushLoop(kind loopKind, loopStart int) *loopContext {
	ctx := &loopContext{
		kind:      kind,
		loopStart: loopStart,
	}
	c.loopStack = append(c.loopStack, ctx)
	return ctx
}

func (c *Compiler) popLoop() {
	if len(c.loopStack) == 0 {
		return
	}
	c.loopStack = c.loopStack[:len(c.loopStack)-1]
}

func (c *Compiler) currentLoop() *loopContext {
	if len(c.loopStack) == 0 {
		return nil
	}
	return c.loopStack[len(c.loopStack)-1]
}

func (c *Compiler) patchLoopBreaks(ctx *loopContext) error {
	if ctx == nil || len(ctx.breakJumps) == 0 {
		return nil
	}
	target := len(c.chunk.Code)
	for _, idx := range ctx.breakJumps {
		if err := c.patchJumpToTarget(idx, target); err != nil {
			return err
		}
	}
	ctx.breakJumps = ctx.breakJumps[:0]
	return nil
}

func (c *Compiler) patchLoopContinues(ctx *loopContext, target int) error {
	if ctx == nil || len(ctx.continueJumps) == 0 {
		return nil
	}
	for _, idx := range ctx.continueJumps {
		if err := c.patchJumpToTarget(idx, target); err != nil {
			return err
		}
	}
	ctx.continueJumps = ctx.continueJumps[:0]
	return nil
}

func (c *Compiler) patchJumpToTarget(jumpIndex, target int) error {
	offset := target - jumpIndex - 1
	if offset > 32767 || offset < -32768 {
		return c.errorf(nil, "jump offset too large: %d", offset)
	}

	inst := c.chunk.Code[jumpIndex]
	c.chunk.Code[jumpIndex] = MakeInstruction(inst.OpCode(), inst.A(), uint16(offset))
	return nil
}

func (c *Compiler) buildUpvalueDefs() []UpvalueDef {
	if len(c.upvalues) == 0 {
		return nil
	}
	defs := make([]UpvalueDef, len(c.upvalues))
	for i, uv := range c.upvalues {
		defs[i] = UpvalueDef{
			IsLocal: uv.isLocal,
			Index:   int(uv.index),
		}
	}
	return defs
}

func (c *Compiler) ensureFunctionReturn(line int) {
	if len(c.chunk.Code) == 0 {
		c.chunk.Write(OpReturn, 0, 0, line)
		return
	}
	last := c.chunk.Code[len(c.chunk.Code)-1]
	if last.OpCode() != OpReturn {
		c.chunk.Write(OpReturn, 0, 0, line)
	}
}

func (c *Compiler) loadExceptObject(line int) error {
	global, ok := c.resolveGlobal(builtinExceptObjectName)
	if !ok {
		return c.errorf(nil, "%s global not initialized", builtinExceptObjectName)
	}
	c.chunk.Write(OpLoadGlobal, 0, global.index, line)
	return nil
}

func (c *Compiler) beginScope() {
	c.scopeDepth++
}

func (c *Compiler) endScope() {
	if c.scopeDepth == 0 {
		return
	}

	for len(c.locals) > 0 && c.locals[len(c.locals)-1].depth == c.scopeDepth {
		c.locals = c.locals[:len(c.locals)-1]
	}
	c.scopeDepth--
}

func (c *Compiler) isGlobalScope() bool {
	return c.enclosing == nil && c.scopeDepth == 0
}

func (c *Compiler) initBuiltins() {
	c.addBuiltinGlobal(builtinExceptObjectName)
	// Register built-in functions as globals
	c.addBuiltinGlobal("PrintLn")
	c.addBuiltinGlobal("Print")
	c.addBuiltinGlobal("IntToStr")
	c.addBuiltinGlobal("FloatToStr")
	c.addBuiltinGlobal("StrToInt")
	c.addBuiltinGlobal("StrToFloat")
	c.addBuiltinGlobal("Length")
	c.addBuiltinGlobal("Copy")
	c.addBuiltinGlobal("Ord")
	c.addBuiltinGlobal("Chr")
	// Type cast functions
	c.addBuiltinGlobal("Integer")
	c.addBuiltinGlobal("Float")
	c.addBuiltinGlobal("String")
	c.addBuiltinGlobal("Boolean")
}

func (c *Compiler) addBuiltinGlobal(name string) {
	if c.globals == nil {
		c.globals = make(map[string]globalVar)
	}
	key := strings.ToLower(name)
	if _, exists := c.globals[key]; exists {
		return
	}
	c.globals[key] = globalVar{
		name:  name,
		index: c.nextGlobal,
	}
	c.nextGlobal++
}

func (c *Compiler) hasEnclosingLocal(name string) bool {
	for env := c.enclosing; env != nil; env = env.enclosing {
		if _, ok := env.resolveLocal(name); ok {
			return true
		}
	}
	return false
}

func (c *Compiler) inferExpressionType(expr ast.Expression) types.Type {
	if expr == nil {
		return nil
	}

	switch node := expr.(type) {
	case ast.TypedExpression:
		return typeFromAnnotation(node.GetType())
	case *ast.IntegerLiteral:
		return types.INTEGER
	case *ast.FloatLiteral:
		return types.FLOAT
	case *ast.StringLiteral:
		return types.STRING
	case *ast.BooleanLiteral:
		return types.BOOLEAN
	case *ast.NilLiteral:
		return types.NIL
	case *ast.Identifier:
		if localInfo, ok := c.resolveLocal(node.Value); ok {
			return localInfo.typ
		}
		if globalInfo, ok := c.resolveGlobal(node.Value); ok {
			return globalInfo.typ
		}
		return typeFromAnnotation(node.GetType())
	default:
		return nil
	}
}

func (c *Compiler) needsHalt() bool {
	if len(c.chunk.Code) == 0 {
		return true
	}

	last := c.chunk.Code[len(c.chunk.Code)-1]
	return last.OpCode() != OpReturn && last.OpCode() != OpHalt
}

func (c *Compiler) errorf(node ast.Node, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	if node != nil {
		pos := node.Pos()
		if pos.Line > 0 {
			message = fmt.Sprintf("%s at %d:%d", message, pos.Line, pos.Column)
		}
	}
	return fmt.Errorf("bytecode compile error: %s", message)
}

func lineOf(node ast.Node) int {
	if node == nil {
		return 0
	}
	pos := node.Pos()
	return pos.Line
}

func typeFromAnnotation(annotation *ast.TypeAnnotation) types.Type {
	if annotation == nil {
		return nil
	}

	switch strings.ToLower(annotation.Name) {
	case "integer":
		return types.INTEGER
	case "float":
		return types.FLOAT
	case "string":
		return types.STRING
	case "boolean":
		return types.BOOLEAN
	case "variant":
		return types.VARIANT
	case "nil":
		return types.NIL
	case "void":
		return types.VOID
	default:
		return nil
	}
}

func lineOfExceptClause(clause *ast.ExceptClause) int {
	if clause == nil {
		return 0
	}
	for _, handler := range clause.Handlers {
		if handler != nil && handler.Statement != nil {
			if line := lineOf(handler.Statement); line != 0 {
				return line
			}
		}
	}
	if clause.ElseBlock != nil {
		return lineOf(clause.ElseBlock)
	}
	return 0
}

func isFloatType(t types.Type) bool {
	if t == nil {
		return false
	}
	return types.GetUnderlyingType(t).TypeKind() == types.FLOAT.TypeKind()
}

func isStringType(t types.Type) bool {
	if t == nil {
		return false
	}
	return types.GetUnderlyingType(t).TypeKind() == types.STRING.TypeKind()
}

func literalValue(expr ast.Expression) (Value, bool) {
	switch node := expr.(type) {
	case *ast.IntegerLiteral:
		return IntValue(node.Value), true
	case *ast.FloatLiteral:
		return FloatValue(node.Value), true
	case *ast.StringLiteral:
		return StringValue(node.Value), true
	case *ast.BooleanLiteral:
		return BoolValue(node.Value), true
	case *ast.NilLiteral:
		return NilValue(), true
	case *ast.UnaryExpression:
		operand, ok := literalValue(node.Right)
		if !ok {
			return Value{}, false
		}
		return evaluateUnary(node.Operator, operand)
	default:
		return Value{}, false
	}
}

func evaluateBinary(operator string, left, right Value) (Value, bool) {
	switch strings.ToLower(operator) {
	case "+":
		if left.IsString() && right.IsString() {
			return StringValue(left.AsString() + right.AsString()), true
		}
		if left.IsNumber() && right.IsNumber() {
			if left.Type == ValueFloat || right.Type == ValueFloat {
				return FloatValue(left.AsFloat() + right.AsFloat()), true
			}
			return IntValue(left.AsInt() + right.AsInt()), true
		}
	case "-":
		if left.IsNumber() && right.IsNumber() {
			if left.Type == ValueFloat || right.Type == ValueFloat {
				return FloatValue(left.AsFloat() - right.AsFloat()), true
			}
			return IntValue(left.AsInt() - right.AsInt()), true
		}
	case "*":
		if left.IsNumber() && right.IsNumber() {
			if left.Type == ValueFloat || right.Type == ValueFloat {
				return FloatValue(left.AsFloat() * right.AsFloat()), true
			}
			return IntValue(left.AsInt() * right.AsInt()), true
		}
	case "div":
		if left.Type == ValueInt && right.Type == ValueInt {
			divisor := right.AsInt()
			if divisor == 0 {
				return Value{}, false
			}
			return IntValue(left.AsInt() / divisor), true
		}
	case "/":
		if left.IsNumber() && right.IsNumber() {
			divisor := right.AsFloat()
			if divisor == 0 {
				return Value{}, false
			}
			return FloatValue(left.AsFloat() / divisor), true
		}
	case "mod":
		if left.Type == ValueInt && right.Type == ValueInt {
			divisor := right.AsInt()
			if divisor == 0 {
				return Value{}, false
			}
			return IntValue(left.AsInt() % divisor), true
		}
	case "=":
		if eq, ok := valuesEqualForFold(left, right); ok {
			return BoolValue(eq), true
		}
	case "<>":
		if eq, ok := valuesEqualForFold(left, right); ok {
			return BoolValue(!eq), true
		}
	case "<":
		if left.IsNumber() && right.IsNumber() {
			if left.Type == ValueFloat || right.Type == ValueFloat {
				return BoolValue(left.AsFloat() < right.AsFloat()), true
			}
			return BoolValue(left.AsInt() < right.AsInt()), true
		}
		if left.IsString() && right.IsString() {
			return BoolValue(left.AsString() < right.AsString()), true
		}
	case "<=":
		if left.IsNumber() && right.IsNumber() {
			if left.Type == ValueFloat || right.Type == ValueFloat {
				return BoolValue(left.AsFloat() <= right.AsFloat()), true
			}
			return BoolValue(left.AsInt() <= right.AsInt()), true
		}
		if left.IsString() && right.IsString() {
			return BoolValue(left.AsString() <= right.AsString()), true
		}
	case ">":
		if left.IsNumber() && right.IsNumber() {
			if left.Type == ValueFloat || right.Type == ValueFloat {
				return BoolValue(left.AsFloat() > right.AsFloat()), true
			}
			return BoolValue(left.AsInt() > right.AsInt()), true
		}
		if left.IsString() && right.IsString() {
			return BoolValue(left.AsString() > right.AsString()), true
		}
	case ">=":
		if left.IsNumber() && right.IsNumber() {
			if left.Type == ValueFloat || right.Type == ValueFloat {
				return BoolValue(left.AsFloat() >= right.AsFloat()), true
			}
			return BoolValue(left.AsInt() >= right.AsInt()), true
		}
		if left.IsString() && right.IsString() {
			return BoolValue(left.AsString() >= right.AsString()), true
		}
	case "and":
		if left.Type == ValueBool && right.Type == ValueBool {
			return BoolValue(left.AsBool() && right.AsBool()), true
		}
	case "or":
		if left.Type == ValueBool && right.Type == ValueBool {
			return BoolValue(left.AsBool() || right.AsBool()), true
		}
	}
	return Value{}, false
}

func evaluateUnary(operator string, operand Value) (Value, bool) {
	switch strings.ToLower(operator) {
	case "+":
		if operand.IsNumber() {
			return operand, true
		}
	case "-":
		if operand.Type == ValueFloat {
			return FloatValue(-operand.AsFloat()), true
		}
		if operand.Type == ValueInt {
			return IntValue(-operand.AsInt()), true
		}
	case "not":
		if operand.Type == ValueBool {
			return BoolValue(!operand.AsBool()), true
		}
	}
	return Value{}, false
}

func valuesEqualForFold(left, right Value) (bool, bool) {
	if left.Type == right.Type {
		switch left.Type {
		case ValueNil:
			return true, true
		case ValueBool:
			return left.AsBool() == right.AsBool(), true
		case ValueInt:
			return left.AsInt() == right.AsInt(), true
		case ValueFloat:
			return left.AsFloat() == right.AsFloat(), true
		case ValueString:
			return left.AsString() == right.AsString(), true
		default:
			return false, false
		}
	}

	if left.IsNumber() && right.IsNumber() {
		return left.AsFloat() == right.AsFloat(), true
	}

	return false, false
}
