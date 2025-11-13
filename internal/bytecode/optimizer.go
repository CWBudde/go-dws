package bytecode

// chunkOptimizer rewrites a chunk in-place after running optimization passes.
type OptimizationPass string

const (
	PassLiteralDiscard   OptimizationPass = "literal-push-pop"
	PassStackShuffle     OptimizationPass = "stack-shuffle"
	PassInlineSmall      OptimizationPass = "inline-small"
	PassConstPropagation OptimizationPass = "const-prop"
	PassDeadCode         OptimizationPass = "dead-code"
)

// OptimizeOption toggles optimizer behavior.
type OptimizeOption func(*optimizeConfig)

type optimizeConfig struct {
	enabled map[OptimizationPass]bool
}

func defaultOptimizeConfig() optimizeConfig {
	return optimizeConfig{
		enabled: map[OptimizationPass]bool{
			PassLiteralDiscard:   true,
			PassStackShuffle:     true,
			PassInlineSmall:      true,
			PassConstPropagation: true,
			PassDeadCode:         true,
		},
	}
}

func (cfg optimizeConfig) isEnabled(pass OptimizationPass) bool {
	if cfg.enabled == nil {
		return true
	}
	enabled, ok := cfg.enabled[pass]
	if !ok {
		return true
	}
	return enabled
}

// WithOptimizationPass enables or disables an optimization pass.
func WithOptimizationPass(pass OptimizationPass, enabled bool) OptimizeOption {
	return func(cfg *optimizeConfig) {
		if cfg.enabled == nil {
			cfg.enabled = make(map[OptimizationPass]bool)
		}
		cfg.enabled[pass] = enabled
	}
}

type optimizerPass struct {
	run func(*chunkOptimizer) bool
	id  OptimizationPass
}

type chunkOptimizer struct {
	chunk             *Chunk
	jumpTargets       map[int]int
	tryInfos          map[int]TryInfo
	config            optimizeConfig
	currentCode       []Instruction
	currentLines      []int
	currentToOriginal []int
	passes            []optimizerPass
	originalCount     int
	hasLineInfo       bool
	changed           bool
}

func newChunkOptimizer(chunk *Chunk, cfg optimizeConfig) *chunkOptimizer {
	lineTable := chunk.lineNumberTable()
	currentCode := make([]Instruction, len(chunk.Code))
	copy(currentCode, chunk.Code)

	currentLines := make([]int, len(lineTable))
	copy(currentLines, lineTable)

	currentToOriginal := make([]int, len(chunk.Code))
	for i := range currentToOriginal {
		currentToOriginal[i] = i
	}

	opt := &chunkOptimizer{
		chunk:             chunk,
		currentCode:       currentCode,
		currentLines:      currentLines,
		currentToOriginal: currentToOriginal,
		originalCount:     len(chunk.Code),
		jumpTargets:       captureJumpTargets(chunk.Code),
		tryInfos:          copyTryInfos(chunk.tryInfos),
		hasLineInfo:       len(chunk.Lines) > 0,
		config:            cfg,
	}

	opt.passes = []optimizerPass{
		{id: PassLiteralDiscard, run: (*chunkOptimizer).peepholeLiteralPop},
		{id: PassStackShuffle, run: (*chunkOptimizer).collapseStackShuffles},
		{id: PassInlineSmall, run: (*chunkOptimizer).inlineSmallFunctions},
		{id: PassConstPropagation, run: (*chunkOptimizer).propagateConstants},
		{id: PassDeadCode, run: (*chunkOptimizer).eliminateDeadCode},
	}

	return opt
}

func (o *chunkOptimizer) run() {
	for _, pass := range o.passes {
		if !o.config.isEnabled(pass.id) {
			continue
		}
		if pass.run(o) {
			o.changed = true
		}
	}

	if o.changed {
		o.apply()
	}
}

// peepholeLiteralPop removes literal push instructions that are immediately discarded.
func (o *chunkOptimizer) peepholeLiteralPop() bool {
	if len(o.currentCode) == 0 {
		return false
	}

	newCode := make([]Instruction, 0, len(o.currentCode))
	newLines := make([]int, 0, len(o.currentLines))
	newMapping := make([]int, 0, len(o.currentToOriginal))

	removed := false

	for i := 0; i < len(o.currentCode); {
		inst := o.currentCode[i]
		if i+1 < len(o.currentCode) &&
			isLiteralPush(inst.OpCode()) &&
			o.currentCode[i+1].OpCode() == OpPop {
			removed = true
			i += 2
			continue
		}

		newCode = append(newCode, inst)
		newLines = append(newLines, o.currentLines[i])
		newMapping = append(newMapping, o.currentToOriginal[i])
		i++
	}

	if removed {
		o.currentCode = newCode
		o.currentLines = newLines
		o.currentToOriginal = newMapping
	}

	return removed
}

func (o *chunkOptimizer) collapseStackShuffles() bool {
	if len(o.currentCode) == 0 {
		return false
	}

	newCode := make([]Instruction, 0, len(o.currentCode))
	newLines := make([]int, 0, len(o.currentLines))
	newMapping := make([]int, 0, len(o.currentToOriginal))

	changed := false

	for i := 0; i < len(o.currentCode); {
		inst := o.currentCode[i]
		switch inst.OpCode() {
		case OpDup:
			if i+1 < len(o.currentCode) && o.currentCode[i+1].OpCode() == OpPop {
				changed = true
				i += 2
				continue
			}
		case OpDup2:
			if i+2 < len(o.currentCode) &&
				o.currentCode[i+1].OpCode() == OpPop &&
				o.currentCode[i+2].OpCode() == OpPop {
				changed = true
				i += 3
				continue
			}
		case OpSwap:
			runLen := 1
			for j := i + 1; j < len(o.currentCode) && o.currentCode[j].OpCode() == OpSwap; j++ {
				runLen++
			}

			if runLen > 1 {
				changed = true
			}

			if runLen%2 == 1 {
				idx := i + runLen - 1
				newCode = append(newCode, o.currentCode[idx])
				newLines = append(newLines, o.currentLines[idx])
				newMapping = append(newMapping, o.currentToOriginal[idx])
			}

			i += runLen
			continue
		case OpRotate3:
			runLen := 1
			for j := i + 1; j < len(o.currentCode) && o.currentCode[j].OpCode() == OpRotate3; j++ {
				runLen++
			}

			if runLen >= 3 {
				changed = true
			}

			remaining := runLen % 3
			if remaining > 0 {
				start := i + runLen - remaining
				for k := start; k < i+runLen; k++ {
					newCode = append(newCode, o.currentCode[k])
					newLines = append(newLines, o.currentLines[k])
					newMapping = append(newMapping, o.currentToOriginal[k])
				}
			}

			i += runLen
			continue
		}

		newCode = append(newCode, inst)
		newLines = append(newLines, o.currentLines[i])
		newMapping = append(newMapping, o.currentToOriginal[i])
		i++
	}

	if changed {
		o.currentCode = newCode
		o.currentLines = newLines
		o.currentToOriginal = newMapping
	}

	return changed
}

const inlineInstructionLimit = 10

func (o *chunkOptimizer) inlineSmallFunctions() bool {
	if len(o.currentCode) == 0 {
		return false
	}

	newCode := make([]Instruction, 0, len(o.currentCode))
	newLines := make([]int, 0, len(o.currentLines))
	newMapping := make([]int, 0, len(o.currentToOriginal))

	changed := false

	for i, inst := range o.currentCode {
		line := 0
		if i < len(o.currentLines) {
			line = o.currentLines[i]
		}
		orig := -1
		if i < len(o.currentToOriginal) {
			orig = o.currentToOriginal[i]
		}

		if inst.OpCode() == OpCall {
			if seq := o.inlineSequenceForCall(inst, line, orig); len(seq) > 0 {
				for _, inlineInst := range seq {
					newCode = append(newCode, inlineInst.inst)
					newLines = append(newLines, inlineInst.line)
					newMapping = append(newMapping, inlineInst.original)
				}
				changed = true
				continue
			}
		}

		newCode = append(newCode, inst)
		newLines = append(newLines, line)
		newMapping = append(newMapping, orig)
	}

	if changed {
		o.currentCode = newCode
		o.currentLines = newLines
		o.currentToOriginal = newMapping
	}

	return changed
}

type inlineInstruction struct {
	inst     Instruction
	line     int
	original int
}

func (o *chunkOptimizer) inlineSequenceForCall(callInst Instruction, line, original int) []inlineInstruction {
	argCount := int(callInst.A())
	if argCount != 0 {
		return nil
	}
	constIdx := int(callInst.B())
	if constIdx >= len(o.chunk.Constants) {
		return nil
	}
	fnVal := o.chunk.Constants[constIdx]
	if fnVal.Type != ValueFunction && fnVal.Type != ValueClosure {
		return nil
	}

	var fn *FunctionObject
	switch fnVal.Type {
	case ValueFunction:
		fn = fnVal.AsFunction()
	case ValueClosure:
		if closure := fnVal.AsClosure(); closure != nil {
			fn = closure.Function
		}
	}
	if fn == nil || fn.Chunk == nil {
		return nil
	}
	if fn.Arity != 0 || fn.UpvalueCount() != 0 {
		return nil
	}
	if fn.Chunk.LocalCount != 0 {
		return nil
	}
	if len(fn.Chunk.Code) == 0 || len(fn.Chunk.Code) > inlineInstructionLimit {
		return nil
	}
	seq := make([]inlineInstruction, 0, len(fn.Chunk.Code))
	for _, inst := range fn.Chunk.Code {
		switch inst.OpCode() {
		case OpLoadConst:
			value := fn.Chunk.Constants[int(inst.B())]
			idx := o.chunk.AddConstant(value)
			seq = append(seq, inlineInstruction{inst: MakeInstruction(OpLoadConst, 0, uint16(idx)), line: line, original: original})
		case OpLoadConst0:
			if len(fn.Chunk.Constants) == 0 {
				return nil
			}
			value := fn.Chunk.Constants[0]
			idx := o.chunk.AddConstant(value)
			seq = append(seq, inlineInstruction{inst: MakeInstruction(OpLoadConst, 0, uint16(idx)), line: line, original: original})
		case OpLoadConst1:
			if len(fn.Chunk.Constants) < 2 {
				return nil
			}
			value := fn.Chunk.Constants[1]
			idx := o.chunk.AddConstant(value)
			seq = append(seq, inlineInstruction{inst: MakeInstruction(OpLoadConst, 0, uint16(idx)), line: line, original: original})
		case OpLoadTrue, OpLoadFalse, OpLoadNil,
			OpAddInt, OpSubInt, OpMulInt, OpDivInt, OpModInt,
			OpAddFloat, OpSubFloat, OpMulFloat, OpDivFloat,
			OpEqual, OpNotEqual, OpGreater, OpGreaterEqual, OpLess, OpLessEqual,
			OpStringConcat:
			seq = append(seq, inlineInstruction{inst: inst, line: line, original: original})
		case OpReturn:
			if inst.A() == 0 {
				seq = append(seq, inlineInstruction{inst: MakeSimpleInstruction(OpLoadNil), line: line, original: original})
			}
			return seq
		case OpHalt:
			return seq
		default:
			return nil
		}
	}
	return seq
}

type valueState struct {
	value    Value
	producer int
	known    bool
}

func (o *chunkOptimizer) propagateConstants() bool {
	if len(o.currentCode) == 0 {
		return false
	}

	ctx := &constPropContext{
		optimizer:  o,
		newCode:    make([]Instruction, 0, len(o.currentCode)),
		newLines:   make([]int, 0, len(o.currentLines)),
		newMapping: make([]int, 0, len(o.currentToOriginal)),
		stack:      make([]valueState, 0, 16),
		locals:     make(map[uint16]valueState),
		changed:    false,
	}

	for i, inst := range o.currentCode {
		line := 0
		if i < len(o.currentLines) {
			line = o.currentLines[i]
		}
		orig := -1
		if i < len(o.currentToOriginal) {
			orig = o.currentToOriginal[i]
		}

		ctx.processInstruction(inst, line, orig)
	}

	if ctx.changed {
		o.currentCode = ctx.newCode
		o.currentLines = ctx.newLines
		o.currentToOriginal = ctx.newMapping
	}

	return ctx.changed
}

// constPropContext holds state for constant propagation optimization.
type constPropContext struct {
	optimizer  *chunkOptimizer
	newCode    []Instruction
	newLines   []int
	newMapping []int
	stack      []valueState
	locals     map[uint16]valueState
	changed    bool
}

func (ctx *constPropContext) processInstruction(inst Instruction, line, orig int) {
	op := inst.OpCode()

	switch op {
	case OpLoadConst, OpLoadConst0, OpLoadConst1, OpLoadTrue, OpLoadFalse, OpLoadNil:
		ctx.handleConstantLoad(inst, op, line, orig)
	case OpLoadLocal, OpStoreLocal:
		ctx.handleLocalVar(inst, op, line, orig)
	case OpLoadGlobal, OpStoreGlobal:
		ctx.handleGlobalVar(inst, op, line, orig)
	case OpPop:
		ctx.popStack()
		ctx.emit(inst, line, orig)
	case OpAddInt, OpSubInt, OpMulInt, OpDivInt, OpModInt,
		OpAddFloat, OpSubFloat, OpMulFloat, OpDivFloat,
		OpEqual, OpNotEqual, OpGreater, OpGreaterEqual, OpLess, OpLessEqual:
		ctx.handleBinaryOp(inst, op, line, orig)
	case OpJump, OpJumpIfTrue, OpJumpIfFalse, OpJumpIfTrueNoPop, OpJumpIfFalseNoPop,
		OpLoop, OpTry, OpCatch, OpFinally, OpCall, OpCallMethod, OpCallIndirect,
		OpReturn, OpThrow:
		ctx.resetAll()
		ctx.emit(inst, line, orig)
	default:
		ctx.resetStack()
		ctx.emit(inst, line, orig)
	}
}

// handleConstantLoad processes constant loading instructions and tracks their values in the stack.
func (ctx *constPropContext) handleConstantLoad(inst Instruction, op OpCode, line, orig int) {
	var val Value
	known := false

	switch op {
	case OpLoadConst:
		constIdx := int(inst.B())
		val = ctx.optimizer.chunk.Constants[constIdx]
		known = true
	case OpLoadConst0:
		if len(ctx.optimizer.chunk.Constants) > 0 {
			val = ctx.optimizer.chunk.Constants[0]
			known = true
		}
	case OpLoadConst1:
		if len(ctx.optimizer.chunk.Constants) > 1 {
			val = ctx.optimizer.chunk.Constants[1]
			known = true
		}
	case OpLoadTrue:
		val = BoolValue(true)
		known = true
	case OpLoadFalse:
		val = BoolValue(false)
		known = true
	case OpLoadNil:
		val = NilValue()
		known = true
	}

	prod := ctx.emit(inst, line, orig)
	if known {
		ctx.pushState(valueState{known: true, value: val, producer: prod})
	} else {
		ctx.pushState(valueState{producer: prod})
	}
}

// handleLocalVar processes local variable load and store operations, tracking known constant values.
func (ctx *constPropContext) handleLocalVar(inst Instruction, op OpCode, line, orig int) {
	slot := uint16(inst.B())

	if op == OpLoadLocal {
		if state, ok := ctx.locals[slot]; ok && state.known {
			prod := ctx.emitValue(state.value, line, orig)
			ctx.pushState(valueState{known: true, value: state.value, producer: prod})
			ctx.changed = true
		} else {
			prod := ctx.emit(inst, line, orig)
			ctx.pushState(valueState{producer: prod})
		}
	} else { // OpStoreLocal
		val := ctx.popStack()
		ctx.emit(inst, line, orig)
		if val.known {
			ctx.locals[slot] = val
		} else {
			delete(ctx.locals, slot)
		}
	}
}

// handleGlobalVar processes global variable load and store operations.
func (ctx *constPropContext) handleGlobalVar(inst Instruction, op OpCode, line, orig int) {
	if op == OpLoadGlobal {
		prod := ctx.emit(inst, line, orig)
		ctx.pushState(valueState{producer: prod})
	} else { // OpStoreGlobal
		ctx.popStack()
		ctx.emit(inst, line, orig)
	}
}

// handleBinaryOp processes binary operations, performing constant folding when both operands are known.
func (ctx *constPropContext) handleBinaryOp(inst Instruction, op OpCode, line, orig int) {
	right := ctx.popStack()
	left := ctx.popStack()

	if result, ok := foldBinaryOp(op, left, right); ok {
		if ctx.isSuffixProducer(right, 1) && ctx.isSuffixProducer(left, 2) {
			ctx.removeTail(2)
			prod := ctx.emitValue(result, line, orig)
			ctx.pushState(valueState{known: true, value: result, producer: prod})
			ctx.changed = true
			return
		}
		prod := ctx.emit(inst, line, orig)
		ctx.pushState(valueState{known: true, value: result, producer: prod})
	} else {
		prod := ctx.emit(inst, line, orig)
		ctx.pushState(valueState{producer: prod})
	}
}

func (ctx *constPropContext) emit(inst Instruction, line int, originalIdx int) int {
	ctx.newCode = append(ctx.newCode, inst)
	ctx.newLines = append(ctx.newLines, line)
	ctx.newMapping = append(ctx.newMapping, originalIdx)
	return len(ctx.newCode) - 1
}

func (ctx *constPropContext) emitValue(val Value, line int, originalIdx int) int {
	idx := ctx.optimizer.chunk.AddConstant(val)
	var inst Instruction
	switch idx {
	case 0:
		inst = MakeSimpleInstruction(OpLoadConst0)
	case 1:
		inst = MakeSimpleInstruction(OpLoadConst1)
	default:
		inst = MakeInstruction(OpLoadConst, 0, uint16(idx))
	}
	return ctx.emit(inst, line, originalIdx)
}

func (ctx *constPropContext) popStack() valueState {
	if len(ctx.stack) == 0 {
		return valueState{}
	}
	val := ctx.stack[len(ctx.stack)-1]
	ctx.stack = ctx.stack[:len(ctx.stack)-1]
	return val
}

func (ctx *constPropContext) pushState(state valueState) {
	ctx.stack = append(ctx.stack, state)
}

func (ctx *constPropContext) resetStack() {
	ctx.stack = ctx.stack[:0]
}

func (ctx *constPropContext) resetAll() {
	ctx.resetStack()
	ctx.locals = make(map[uint16]valueState)
}

func (ctx *constPropContext) isSuffixProducer(state valueState, offsetFromEnd int) bool {
	if !state.known || state.producer < 0 {
		return false
	}
	expected := len(ctx.newCode) - offsetFromEnd
	if expected < 0 {
		return false
	}
	return state.producer == expected
}

func (ctx *constPropContext) removeTail(count int) {
	if count <= 0 {
		return
	}
	if count > len(ctx.newCode) {
		count = len(ctx.newCode)
	}
	ctx.newCode = ctx.newCode[:len(ctx.newCode)-count]
	ctx.newLines = ctx.newLines[:len(ctx.newLines)-count]
	ctx.newMapping = ctx.newMapping[:len(ctx.newMapping)-count]
}

func (o *chunkOptimizer) eliminateDeadCode() bool {
	if len(o.currentCode) == 0 {
		return false
	}

	targetedOriginal := make([]bool, o.originalCount)
	if len(targetedOriginal) > 0 {
		targetedOriginal[0] = true
	}
	for _, target := range o.jumpTargets {
		if target >= 0 && target < len(targetedOriginal) {
			targetedOriginal[target] = true
		}
	}

	newCode := make([]Instruction, 0, len(o.currentCode))
	newLines := make([]int, 0, len(o.currentLines))
	newMapping := make([]int, 0, len(o.currentToOriginal))

	dead := false
	changed := false

	for i := 0; i < len(o.currentCode); i++ {
		origIdx := -1
		if i < len(o.currentToOriginal) {
			origIdx = o.currentToOriginal[i]
		}
		if origIdx >= 0 && origIdx < len(targetedOriginal) && targetedOriginal[origIdx] {
			dead = false
		}

		if dead {
			changed = true
			continue
		}

		inst := o.currentCode[i]
		newCode = append(newCode, inst)
		newLines = append(newLines, o.currentLines[i])
		newMapping = append(newMapping, o.currentToOriginal[i])

		if isTerminator(inst.OpCode()) {
			dead = true
		}
	}

	if changed {
		o.currentCode = newCode
		o.currentLines = newLines
		o.currentToOriginal = newMapping
	}

	return changed
}

func (o *chunkOptimizer) apply() {
	oldToNew := o.buildOldToNew()

	o.chunk.Code = o.currentCode

	if o.hasLineInfo {
		o.chunk.Lines = lineInfoFromTable(o.currentLines)
	} else {
		o.chunk.Lines = nil
	}

	o.chunk.tryInfos = o.rebuildTryInfos(oldToNew)
	o.rewriteJumpOffsets(oldToNew)
}

func (o *chunkOptimizer) buildOldToNew() []int {
	mapping := make([]int, o.originalCount)
	for i := range mapping {
		mapping[i] = -1
	}

	for newIdx, originalIdx := range o.currentToOriginal {
		if originalIdx >= 0 && originalIdx < len(mapping) {
			mapping[originalIdx] = newIdx
		}
	}

	return mapping
}

func (o *chunkOptimizer) rebuildTryInfos(oldToNew []int) map[int]TryInfo {
	if len(o.tryInfos) == 0 {
		return nil
	}

	nextKept := buildNextKept(oldToNew)
	newCount := len(o.currentCode)
	oldCount := len(oldToNew)

	updated := make(map[int]TryInfo, len(o.tryInfos))
	for oldIdx, info := range o.tryInfos {
		if oldIdx < 0 || oldIdx >= len(oldToNew) {
			continue
		}
		newIdx := oldToNew[oldIdx]
		if newIdx == -1 {
			continue
		}
		newInfo := info
		if info.HasCatch {
			newInfo.CatchTarget = remapTargetIndex(info.CatchTarget, oldToNew, nextKept, newCount, oldCount)
		}
		if info.HasFinally {
			newInfo.FinallyTarget = remapTargetIndex(info.FinallyTarget, oldToNew, nextKept, newCount, oldCount)
		}
		updated[newIdx] = newInfo
	}

	if len(updated) == 0 {
		return nil
	}

	return updated
}

func (o *chunkOptimizer) rewriteJumpOffsets(oldToNew []int) {
	if len(o.jumpTargets) == 0 {
		return
	}

	nextKept := buildNextKept(oldToNew)
	newCount := len(o.chunk.Code)
	oldCount := len(oldToNew)

	for oldIdx, oldTarget := range o.jumpTargets {
		if oldIdx < 0 || oldIdx >= len(oldToNew) {
			continue
		}

		jumpNewIdx := oldToNew[oldIdx]
		if jumpNewIdx == -1 {
			continue
		}

		newTarget := mapTargetIndex(oldTarget, nextKept, newCount, oldCount)
		if newTarget < 0 {
			continue
		}

		inst := o.chunk.Code[jumpNewIdx]
		offset := newTarget - jumpNewIdx - 1
		if offset > 32767 || offset < -32768 {
			continue
		}

		o.chunk.Code[jumpNewIdx] = MakeInstruction(inst.OpCode(), inst.A(), uint16(offset))
	}
}

func isLiteralPush(op OpCode) bool {
	switch op {
	case OpLoadConst, OpLoadConst0, OpLoadConst1, OpLoadNil, OpLoadTrue, OpLoadFalse:
		return true
	default:
		return false
	}
}

func buildNextKept(oldToNew []int) []int {
	next := -1
	nextKept := make([]int, len(oldToNew))
	for i := len(oldToNew) - 1; i >= 0; i-- {
		if oldToNew[i] != -1 {
			next = oldToNew[i]
		}
		nextKept[i] = next
	}
	return nextKept
}

func mapTargetIndex(oldTarget int, nextKept []int, newCount, oldCount int) int {
	if oldTarget < 0 {
		oldTarget = 0
	}
	if oldTarget >= oldCount {
		return newCount
	}
	mapped := nextKept[oldTarget]
	if mapped == -1 {
		return newCount
	}
	return mapped
}

func captureJumpTargets(code []Instruction) map[int]int {
	if len(code) == 0 {
		return nil
	}

	targets := make(map[int]int)
	for idx, inst := range code {
		if !usesRelativeOffset(inst.OpCode()) {
			continue
		}
		target := idx + 1 + int(inst.SignedB())
		targets[idx] = target
	}
	return targets
}

func usesRelativeOffset(op OpCode) bool {
	switch op {
	case OpJump,
		OpJumpIfTrue,
		OpJumpIfFalse,
		OpJumpIfTrueNoPop,
		OpJumpIfFalseNoPop,
		OpLoop,
		OpTry,
		OpCatch:
		return true
	default:
		return false
	}
}

func copyTryInfos(src map[int]TryInfo) map[int]TryInfo {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[int]TryInfo, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (c *Chunk) lineNumberTable() []int {
	table := make([]int, len(c.Code))
	if len(c.Code) == 0 || len(c.Lines) == 0 {
		return table
	}

	lineIdx := 0
	currentLine := c.Lines[lineIdx].Line

	for i := 0; i < len(c.Code); i++ {
		for lineIdx+1 < len(c.Lines) && i >= c.Lines[lineIdx+1].InstructionOffset {
			lineIdx++
			currentLine = c.Lines[lineIdx].Line
		}
		table[i] = currentLine
	}

	return table
}

func lineInfoFromTable(table []int) []LineInfo {
	if len(table) == 0 {
		return nil
	}

	result := make([]LineInfo, 0, len(table))
	var lastLine int
	hasLast := false

	for i, line := range table {
		if !hasLast || line != lastLine {
			result = append(result, LineInfo{
				InstructionOffset: i,
				Line:              line,
			})
			lastLine = line
			hasLast = true
		}
	}

	return result
}

func remapTargetIndex(oldIdx int, oldToNew, nextKept []int, newCount, oldCount int) int {
	if oldIdx < 0 {
		return oldIdx
	}
	if oldIdx >= oldCount {
		return newCount
	}
	if newIdx := oldToNew[oldIdx]; newIdx != -1 {
		return newIdx
	}
	return mapTargetIndex(oldIdx, nextKept, newCount, oldCount)
}

func isTerminator(op OpCode) bool {
	switch op {
	case OpReturn, OpHalt, OpThrow, OpJump, OpTailCall:
		return true
	default:
		return false
	}
}

func foldBinaryOp(op OpCode, left, right valueState) (Value, bool) {
	if !left.known || !right.known {
		return Value{}, false
	}

	switch op {
	case OpAddInt, OpSubInt, OpMulInt, OpDivInt, OpModInt:
		return foldIntegerOp(op, left.value, right.value)
	case OpAddFloat, OpSubFloat, OpMulFloat, OpDivFloat:
		return foldFloatOp(op, left.value, right.value)
	case OpEqual, OpNotEqual:
		return foldEqualityOp(op, left.value, right.value)
	case OpGreater, OpGreaterEqual, OpLess, OpLessEqual:
		return foldComparisonOp(op, left.value, right.value)
	}

	return Value{}, false
}

// foldIntegerOp performs constant folding for integer arithmetic operations.
func foldIntegerOp(op OpCode, left, right Value) (Value, bool) {
	if left.Type != ValueInt || right.Type != ValueInt {
		return Value{}, false
	}

	leftInt := left.AsInt()
	rightInt := right.AsInt()

	switch op {
	case OpAddInt:
		return IntValue(leftInt + rightInt), true
	case OpSubInt:
		return IntValue(leftInt - rightInt), true
	case OpMulInt:
		return IntValue(leftInt * rightInt), true
	case OpDivInt:
		if rightInt == 0 {
			return Value{}, false
		}
		return IntValue(leftInt / rightInt), true
	case OpModInt:
		if rightInt == 0 {
			return Value{}, false
		}
		return IntValue(leftInt % rightInt), true
	}

	return Value{}, false
}

// foldFloatOp performs constant folding for floating-point arithmetic operations.
func foldFloatOp(op OpCode, left, right Value) (Value, bool) {
	if !left.IsNumber() || !right.IsNumber() {
		return Value{}, false
	}

	leftFloat := left.AsFloat()
	rightFloat := right.AsFloat()

	switch op {
	case OpAddFloat:
		return FloatValue(leftFloat + rightFloat), true
	case OpSubFloat:
		return FloatValue(leftFloat - rightFloat), true
	case OpMulFloat:
		return FloatValue(leftFloat * rightFloat), true
	case OpDivFloat:
		if rightFloat == 0 {
			return Value{}, false
		}
		return FloatValue(leftFloat / rightFloat), true
	}

	return Value{}, false
}

// foldEqualityOp performs constant folding for equality operations.
func foldEqualityOp(op OpCode, left, right Value) (Value, bool) {
	switch op {
	case OpEqual:
		return BoolValue(valuesEqual(left, right)), true
	case OpNotEqual:
		return BoolValue(!valuesEqual(left, right)), true
	}

	return Value{}, false
}

// foldComparisonOp performs constant folding for comparison operations.
func foldComparisonOp(op OpCode, left, right Value) (Value, bool) {
	if !left.IsNumber() || !right.IsNumber() {
		return Value{}, false
	}

	leftFloat := left.AsFloat()
	rightFloat := right.AsFloat()

	switch op {
	case OpGreater:
		return BoolValue(leftFloat > rightFloat), true
	case OpGreaterEqual:
		return BoolValue(leftFloat >= rightFloat), true
	case OpLess:
		return BoolValue(leftFloat < rightFloat), true
	case OpLessEqual:
		return BoolValue(leftFloat <= rightFloat), true
	}

	return Value{}, false
}

func valuesEqual(a, b Value) bool {
	if a.Type != b.Type {
		if a.IsNumber() && b.IsNumber() {
			return a.AsFloat() == b.AsFloat()
		}
		return false
	}

	switch a.Type {
	case ValueNil:
		return true
	case ValueBool:
		return a.AsBool() == b.AsBool()
	case ValueInt:
		return a.AsInt() == b.AsInt()
	case ValueFloat:
		return a.AsFloat() == b.AsFloat()
	case ValueString:
		return a.AsString() == b.AsString()
	default:
		return false
	}
}
