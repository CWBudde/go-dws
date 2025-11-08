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

	newCode := make([]Instruction, 0, len(o.currentCode))
	newLines := make([]int, 0, len(o.currentLines))
	newMapping := make([]int, 0, len(o.currentToOriginal))

	stack := make([]valueState, 0, 16)
	locals := make(map[uint16]valueState)

	resetStack := func() {
		stack = stack[:0]
	}

	resetAll := func() {
		resetStack()
		locals = make(map[uint16]valueState)
	}

	emit := func(inst Instruction, line int, originalIdx int) int {
		newCode = append(newCode, inst)
		newLines = append(newLines, line)
		newMapping = append(newMapping, originalIdx)
		return len(newCode) - 1
	}

	emitValue := func(val Value, line int, originalIdx int) int {
		idx := o.chunk.AddConstant(val)
		switch idx {
		case 0:
			return emit(MakeSimpleInstruction(OpLoadConst0), line, originalIdx)
		case 1:
			return emit(MakeSimpleInstruction(OpLoadConst1), line, originalIdx)
		default:
			return emit(MakeInstruction(OpLoadConst, 0, uint16(idx)), line, originalIdx)
		}
	}

	popStack := func() valueState {
		if len(stack) == 0 {
			return valueState{}
		}
		val := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return val
	}

	pushState := func(state valueState) {
		stack = append(stack, state)
	}

	isSuffixProducer := func(state valueState, offsetFromEnd int) bool {
		if !state.known || state.producer < 0 {
			return false
		}
		expected := len(newCode) - offsetFromEnd
		if expected < 0 {
			return false
		}
		return state.producer == expected
	}

	removeTail := func(count int) {
		if count <= 0 {
			return
		}
		if count > len(newCode) {
			count = len(newCode)
		}
		newCode = newCode[:len(newCode)-count]
		newLines = newLines[:len(newLines)-count]
		newMapping = newMapping[:len(newMapping)-count]
	}

	var changed bool

	for i, inst := range o.currentCode {
		line := 0
		if i < len(o.currentLines) {
			line = o.currentLines[i]
		}
		orig := -1
		if i < len(o.currentToOriginal) {
			orig = o.currentToOriginal[i]
		}

		op := inst.OpCode()

		switch op {
		case OpLoadConst:
			constIdx := int(inst.B())
			val := o.chunk.Constants[constIdx]
			prod := emit(inst, line, orig)
			pushState(valueState{known: true, value: val, producer: prod})
		case OpLoadConst0:
			if len(o.chunk.Constants) > 0 {
				val := o.chunk.Constants[0]
				prod := emit(inst, line, orig)
				pushState(valueState{known: true, value: val, producer: prod})
			} else {
				prod := emit(inst, line, orig)
				pushState(valueState{producer: prod})
			}
		case OpLoadConst1:
			if len(o.chunk.Constants) > 1 {
				val := o.chunk.Constants[1]
				prod := emit(inst, line, orig)
				pushState(valueState{known: true, value: val, producer: prod})
			} else {
				prod := emit(inst, line, orig)
				pushState(valueState{producer: prod})
			}
		case OpLoadTrue:
			prod := emit(inst, line, orig)
			pushState(valueState{known: true, value: BoolValue(true), producer: prod})
		case OpLoadFalse:
			prod := emit(inst, line, orig)
			pushState(valueState{known: true, value: BoolValue(false), producer: prod})
		case OpLoadNil:
			prod := emit(inst, line, orig)
			pushState(valueState{known: true, value: NilValue(), producer: prod})
		case OpLoadLocal:
			slot := uint16(inst.B())
			if state, ok := locals[slot]; ok && state.known {
				prod := emitValue(state.value, line, orig)
				pushState(valueState{known: true, value: state.value, producer: prod})
				changed = true
			} else {
				prod := emit(inst, line, orig)
				pushState(valueState{producer: prod})
			}
		case OpStoreLocal:
			val := popStack()
			emit(inst, line, orig)
			slot := uint16(inst.B())
			if val.known {
				locals[slot] = val
			} else {
				delete(locals, slot)
			}
		case OpLoadGlobal:
			prod := emit(inst, line, orig)
			pushState(valueState{producer: prod})
		case OpStoreGlobal:
			popStack()
			emit(inst, line, orig)
		case OpPop:
			popStack()
			emit(inst, line, orig)
		case OpAddInt, OpSubInt, OpMulInt, OpDivInt, OpModInt,
			OpAddFloat, OpSubFloat, OpMulFloat, OpDivFloat,
			OpEqual, OpNotEqual, OpGreater, OpGreaterEqual, OpLess, OpLessEqual:
			right := popStack()
			left := popStack()
			if result, ok := foldBinaryOp(op, left, right); ok {
				if isSuffixProducer(right, 1) && isSuffixProducer(left, 2) {
					removeTail(2)
					prod := emitValue(result, line, orig)
					pushState(valueState{known: true, value: result, producer: prod})
					changed = true
					continue
				}
				prod := emit(inst, line, orig)
				pushState(valueState{known: true, value: result, producer: prod})
			} else {
				prod := emit(inst, line, orig)
				pushState(valueState{producer: prod})
			}
		case OpJump, OpJumpIfTrue, OpJumpIfFalse, OpJumpIfTrueNoPop, OpJumpIfFalseNoPop,
			OpLoop, OpTry, OpCatch, OpFinally, OpCall, OpCallMethod, OpCallIndirect,
			OpReturn, OpThrow:
			resetAll()
			emit(inst, line, orig)
		default:
			resetStack()
			emit(inst, line, orig)
		}
	}

	if changed {
		o.currentCode = newCode
		o.currentLines = newLines
		o.currentToOriginal = newMapping
	}

	return changed
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
	case OpAddInt:
		if left.value.Type == ValueInt && right.value.Type == ValueInt {
			return IntValue(left.value.AsInt() + right.value.AsInt()), true
		}
	case OpSubInt:
		if left.value.Type == ValueInt && right.value.Type == ValueInt {
			return IntValue(left.value.AsInt() - right.value.AsInt()), true
		}
	case OpMulInt:
		if left.value.Type == ValueInt && right.value.Type == ValueInt {
			return IntValue(left.value.AsInt() * right.value.AsInt()), true
		}
	case OpDivInt:
		if left.value.Type == ValueInt && right.value.Type == ValueInt {
			div := right.value.AsInt()
			if div == 0 {
				return Value{}, false
			}
			return IntValue(left.value.AsInt() / div), true
		}
	case OpModInt:
		if left.value.Type == ValueInt && right.value.Type == ValueInt {
			div := right.value.AsInt()
			if div == 0 {
				return Value{}, false
			}
			return IntValue(left.value.AsInt() % div), true
		}
	case OpAddFloat:
		if left.value.IsNumber() && right.value.IsNumber() {
			return FloatValue(left.value.AsFloat() + right.value.AsFloat()), true
		}
	case OpSubFloat:
		if left.value.IsNumber() && right.value.IsNumber() {
			return FloatValue(left.value.AsFloat() - right.value.AsFloat()), true
		}
	case OpMulFloat:
		if left.value.IsNumber() && right.value.IsNumber() {
			return FloatValue(left.value.AsFloat() * right.value.AsFloat()), true
		}
	case OpDivFloat:
		if left.value.IsNumber() && right.value.IsNumber() {
			div := right.value.AsFloat()
			if div == 0 {
				return Value{}, false
			}
			return FloatValue(left.value.AsFloat() / div), true
		}
	case OpEqual:
		return BoolValue(valuesEqual(left.value, right.value)), true
	case OpNotEqual:
		return BoolValue(!valuesEqual(left.value, right.value)), true
	case OpGreater:
		if left.value.IsNumber() && right.value.IsNumber() {
			return BoolValue(left.value.AsFloat() > right.value.AsFloat()), true
		}
	case OpGreaterEqual:
		if left.value.IsNumber() && right.value.IsNumber() {
			return BoolValue(left.value.AsFloat() >= right.value.AsFloat()), true
		}
	case OpLess:
		if left.value.IsNumber() && right.value.IsNumber() {
			return BoolValue(left.value.AsFloat() < right.value.AsFloat()), true
		}
	case OpLessEqual:
		if left.value.IsNumber() && right.value.IsNumber() {
			return BoolValue(left.value.AsFloat() <= right.value.AsFloat()), true
		}
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
