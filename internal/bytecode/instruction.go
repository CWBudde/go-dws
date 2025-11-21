// Package bytecode implements a stack-based bytecode virtual machine for DWScript.
//
// This package provides bytecode instruction definitions, encoding/decoding utilities,
// and execution logic for a bytecode VM that offers 2-3x performance improvement
// over AST interpretation.
//
// Architecture: Stack-based VM with 32-bit instructions
// Format: [8-bit opcode][8-bit operand A][16-bit operand B]
// Design: See docs/architecture/bytecode-vm-design.md
package bytecode

// OpCode represents a bytecode instruction opcode.
// The bytecode VM uses a stack-based architecture with 116 opcodes
// optimized for Go's switch statement performance (< 128 opcodes = 7 comparisons).
type OpCode byte

// Instruction format:
// - 32-bit fixed-size instructions for Go optimization
// - Format: [8-bit opcode][8-bit A][16-bit B]
// - Alternative formats supported for specific opcodes:
//   - [opcode][unused][unused][unused] - no operands
//   - [opcode][A][B high][B low] - one byte + one short
//   - [opcode][A][B][C] - three bytes (rare)

const (
	// ========================================
	// Constants and Variables (12 opcodes)
	// ========================================

	// OpLoadConst pushes a constant from the constant pool onto the stack.
	// Format: [OpLoadConst][unused][index]
	// Stack: [] -> [constant]
	OpLoadConst OpCode = iota

	// OpLoadConst0 pushes constant at index 0 (common optimization).
	// Format: [OpLoadConst0][unused][unused]
	// Stack: [] -> [constant[0]]
	OpLoadConst0

	// OpLoadConst1 pushes constant at index 1 (common optimization).
	// Format: [OpLoadConst1][unused][unused]
	// Stack: [] -> [constant[1]]
	OpLoadConst1

	// OpLoadLocal loads a local variable onto the stack.
	// Format: [OpLoadLocal][unused][index]
	// Stack: [] -> [local[index]]
	OpLoadLocal

	// OpStoreLocal pops stack and stores to local variable.
	// Format: [OpStoreLocal][unused][index]
	// Stack: [value] -> []
	OpStoreLocal

	// OpLoadGlobal loads a global variable onto the stack.
	// Format: [OpLoadGlobal][unused][index]
	// Stack: [] -> [global[index]]
	OpLoadGlobal

	// OpStoreGlobal pops stack and stores to global variable.
	// Format: [OpStoreGlobal][unused][index]
	// Stack: [value] -> []
	OpStoreGlobal

	// OpLoadUpvalue loads a closure variable (upvalue) onto the stack.
	// Format: [OpLoadUpvalue][unused][index]
	// Stack: [] -> [upvalue[index]]
	OpLoadUpvalue

	// OpStoreUpvalue pops stack and stores to closure variable.
	// Format: [OpStoreUpvalue][unused][index]
	// Stack: [value] -> []
	OpStoreUpvalue

	// OpLoadNil pushes nil onto the stack.
	// Format: [OpLoadNil][unused][unused]
	// Stack: [] -> [nil]
	OpLoadNil

	// OpLoadTrue pushes boolean true onto the stack.
	// Format: [OpLoadTrue][unused][unused]
	// Stack: [] -> [true]
	OpLoadTrue

	// OpLoadFalse pushes boolean false onto the stack.
	// Format: [OpLoadFalse][unused][unused]
	// Stack: [] -> [false]
	OpLoadFalse

	// ========================================
	// Integer Arithmetic (12 opcodes)
	// ========================================

	// OpAddInt pops two integers, adds them, pushes result.
	// Format: [OpAddInt][unused][unused]
	// Stack: [a, b] -> [a + b]
	OpAddInt

	// OpSubInt pops two integers, subtracts them, pushes result.
	// Format: [OpSubInt][unused][unused]
	// Stack: [a, b] -> [a - b]
	OpSubInt

	// OpMulInt pops two integers, multiplies them, pushes result.
	// Format: [OpMulInt][unused][unused]
	// Stack: [a, b] -> [a * b]
	OpMulInt

	// OpDivInt pops two integers, divides them (DWScript 'div'), pushes result.
	// Format: [OpDivInt][unused][unused]
	// Stack: [a, b] -> [a div b]
	// Note: This is integer division, different from OpDivFloat
	OpDivInt

	// OpModInt pops two integers, computes modulo, pushes result.
	// Format: [OpModInt][unused][unused]
	// Stack: [a, b] -> [a mod b]
	OpModInt

	// OpNegateInt pops integer, negates it, pushes result.
	// Format: [OpNegateInt][unused][unused]
	// Stack: [a] -> [-a]
	OpNegateInt

	// OpIncInt increments top of stack integer in-place.
	// Format: [OpIncInt][unused][unused]
	// Stack: [a] -> [a + 1]
	OpIncInt

	// OpDecInt decrements top of stack integer in-place.
	// Format: [OpDecInt][unused][unused]
	// Stack: [a] -> [a - 1]
	OpDecInt

	// OpBitAnd pops two integers, performs bitwise AND, pushes result.
	// Format: [OpBitAnd][unused][unused]
	// Stack: [a, b] -> [a & b]
	OpBitAnd

	// OpBitOr pops two integers, performs bitwise OR, pushes result.
	// Format: [OpBitOr][unused][unused]
	// Stack: [a, b] -> [a | b]
	OpBitOr

	// OpBitXor pops two integers, performs bitwise XOR, pushes result.
	// Format: [OpBitXor][unused][unused]
	// Stack: [a, b] -> [a ^ b]
	OpBitXor

	// OpBitNot pops integer, performs bitwise NOT, pushes result.
	// Format: [OpBitNot][unused][unused]
	// Stack: [a] -> [^a]
	OpBitNot

	// ========================================
	// Float Arithmetic (12 opcodes)
	// ========================================

	// OpAddFloat pops two floats, adds them, pushes result.
	// Format: [OpAddFloat][unused][unused]
	// Stack: [a, b] -> [a + b]
	OpAddFloat

	// OpSubFloat pops two floats, subtracts them, pushes result.
	// Format: [OpSubFloat][unused][unused]
	// Stack: [a, b] -> [a - b]
	OpSubFloat

	// OpMulFloat pops two floats, multiplies them, pushes result.
	// Format: [OpMulFloat][unused][unused]
	// Stack: [a, b] -> [a * b]
	OpMulFloat

	// OpDivFloat pops two floats, divides them (DWScript '/'), pushes result.
	// Format: [OpDivFloat][unused][unused]
	// Stack: [a, b] -> [a / b]
	OpDivFloat

	// OpNegateFloat pops float, negates it, pushes result.
	// Format: [OpNegateFloat][unused][unused]
	// Stack: [a] -> [-a]
	OpNegateFloat

	// OpPower pops two floats, computes a^b, pushes result.
	// Format: [OpPower][unused][unused]
	// Stack: [a, b] -> [a ** b]
	OpPower

	// OpSqrt pops float, computes square root, pushes result.
	// Format: [OpSqrt][unused][unused]
	// Stack: [a] -> [sqrt(a)]
	OpSqrt

	// OpAbs pops numeric value, computes absolute value, pushes result.
	// Format: [OpAbs][unused][unused]
	// Stack: [a] -> [abs(a)]
	OpAbs

	// OpFloor pops float, computes floor, pushes result.
	// Format: [OpFloor][unused][unused]
	// Stack: [a] -> [floor(a)]
	OpFloor

	// OpCeil pops float, computes ceiling, pushes result.
	// Format: [OpCeil][unused][unused]
	// Stack: [a] -> [ceil(a)]
	OpCeil

	// OpRound pops float, rounds to nearest integer, pushes result.
	// Format: [OpRound][unused][unused]
	// Stack: [a] -> [round(a)]
	OpRound

	// OpTrunc pops float, truncates to integer part, pushes result.
	// Format: [OpTrunc][unused][unused]
	// Stack: [a] -> [trunc(a)]
	OpTrunc

	// ========================================
	// Comparison Operations (8 opcodes)
	// ========================================

	// OpEqual pops two values, compares for equality, pushes boolean.
	// Format: [OpEqual][unused][unused]
	// Stack: [a, b] -> [a = b]
	OpEqual

	// OpNotEqual pops two values, compares for inequality, pushes boolean.
	// Format: [OpNotEqual][unused][unused]
	// Stack: [a, b] -> [a <> b]
	OpNotEqual

	// OpLess pops two values, compares less-than, pushes boolean.
	// Format: [OpLess][unused][unused]
	// Stack: [a, b] -> [a < b]
	OpLess

	// OpLessEqual pops two values, compares less-or-equal, pushes boolean.
	// Format: [OpLessEqual][unused][unused]
	// Stack: [a, b] -> [a <= b]
	OpLessEqual

	// OpGreater pops two values, compares greater-than, pushes boolean.
	// Format: [OpGreater][unused][unused]
	// Stack: [a, b] -> [a > b]
	OpGreater

	// OpGreaterEqual pops two values, compares greater-or-equal, pushes boolean.
	// Format: [OpGreaterEqual][unused][unused]
	// Stack: [a, b] -> [a >= b]
	OpGreaterEqual

	// OpCompareInt specialized integer comparison for performance.
	// Format: [OpCompareInt][unused][unused]
	// Stack: [a, b] -> [-1 if a<b, 0 if a=b, 1 if a>b]
	OpCompareInt

	// OpCompareFloat specialized float comparison for performance.
	// Format: [OpCompareFloat][unused][unused]
	// Stack: [a, b] -> [-1 if a<b, 0 if a=b, 1 if a>b]
	OpCompareFloat

	// ========================================
	// Logical Operations (4 opcodes)
	// ========================================

	// OpNot pops boolean, performs logical NOT, pushes result.
	// Format: [OpNot][unused][unused]
	// Stack: [a] -> [not a]
	OpNot

	// OpAnd pops two booleans, performs logical AND, pushes result.
	// Format: [OpAnd][unused][unused]
	// Stack: [a, b] -> [a and b]
	// Note: For short-circuit evaluation, use OpJumpIfFalse instead
	OpAnd

	// OpOr pops two booleans, performs logical OR, pushes result.
	// Format: [OpOr][unused][unused]
	// Stack: [a, b] -> [a or b]
	// Note: For short-circuit evaluation, use OpJumpIfTrue instead
	OpOr

	// OpXor pops two booleans, performs logical XOR, pushes result.
	// Format: [OpXor][unused][unused]
	// Stack: [a, b] -> [a xor b]
	OpXor

	// OpIsFalsey pops a value, checks if it's falsey (0, 0.0, "", false, nil, empty array), pushes boolean result.
	// Format: [OpIsFalsey][unused][unused]
	// Stack: [value] -> [boolean]
	// Used for coalesce operator (??) to check if left operand is falsey
	OpIsFalsey

	// ========================================
	// Bit Shift Operations (4 opcodes)
	// ========================================

	// OpShl pops count and value, performs left shift, pushes result.
	// Format: [OpShl][unused][unused]
	// Stack: [value, count] -> [value << count]
	OpShl

	// OpShr pops count and value, performs logical right shift, pushes result.
	// Format: [OpShr][unused][unused]
	// Stack: [value, count] -> [value >> count]
	OpShr

	// OpSar pops count and value, performs arithmetic right shift, pushes result.
	// Format: [OpSar][unused][unused]
	// Stack: [value, count] -> [value >>> count]
	OpSar

	// OpRotl pops count and value, performs rotate left, pushes result.
	// Format: [OpRotl][unused][unused]
	// Stack: [value, count] -> [rotate_left(value, count)]
	OpRotl

	// ========================================
	// Control Flow (12 opcodes)
	// ========================================

	// OpJump unconditionally jumps to offset.
	// Format: [OpJump][unused][offset]
	// Stack: [] -> []
	// Note: offset is signed 16-bit relative to current IP
	OpJump

	// OpJumpIfTrue pops boolean, jumps if true.
	// Format: [OpJumpIfTrue][unused][offset]
	// Stack: [bool] -> []
	OpJumpIfTrue

	// OpJumpIfFalse pops boolean, jumps if false.
	// Format: [OpJumpIfFalse][unused][offset]
	// Stack: [bool] -> []
	OpJumpIfFalse

	// OpJumpIfTrueNoPop peeks boolean, jumps if true without popping.
	// Format: [OpJumpIfTrueNoPop][unused][offset]
	// Stack: [bool] -> [bool]
	// Note: Used for short-circuit evaluation of 'or'
	OpJumpIfTrueNoPop

	// OpJumpIfFalseNoPop peeks boolean, jumps if false without popping.
	// Format: [OpJumpIfFalseNoPop][unused][offset]
	// Stack: [bool] -> [bool]
	// Note: Used for short-circuit evaluation of 'and'
	OpJumpIfFalseNoPop

	// OpLoop jumps backward for loop iteration.
	// Format: [OpLoop][unused][offset]
	// Stack: [] -> []
	// Note: offset is negative, jumps backward
	OpLoop

	// OpForPrep prepares for-loop, checks condition.
	// Format: [OpForPrep][loopVar][jumpOffset]
	// Stack: [start, end, step] -> [current]
	OpForPrep

	// OpForLoop continues for-loop iteration.
	// Format: [OpForLoop][loopVar][jumpOffset]
	// Stack: [current] -> [current+step] or exits loop
	OpForLoop

	// OpCase implements case/switch dispatch.
	// Format: [OpCase][unused][jumpTableIndex]
	// Stack: [value] -> []
	OpCase

	// OpBreak breaks out of current loop.
	// Format: [OpBreak][unused][unused]
	// Stack: [] -> []
	OpBreak

	// OpContinue continues to next loop iteration.
	// Format: [OpContinue][unused][unused]
	// Stack: [] -> []
	OpContinue

	// OpExit exits current function with return value.
	// Format: [OpExit][unused][unused]
	// Stack: [returnValue] -> [] (and returns)
	OpExit

	// ========================================
	// Function Calls (8 opcodes)
	// ========================================

	// OpCall calls a function with arguments.
	// Format: [OpCall][argCount][funcIndex]
	// Stack: [arg1, arg2, ..., argN] -> [returnValue]
	OpCall

	// OpCallMethod calls a method on an object.
	// Format: [OpCallMethod][argCount][methodIndex]
	// Stack: [obj, arg1, arg2, ..., argN] -> [returnValue]
	OpCallMethod

	// OpCallVirtual calls a virtual method (dynamic dispatch).
	// Format: [OpCallVirtual][argCount][methodIndex]
	// Stack: [obj, arg1, arg2, ..., argN] -> [returnValue]
	OpCallVirtual

	// OpReturn returns from current function.
	// Format: [OpReturn][hasReturnValue][unused]
	// Stack: [returnValue?] -> [] (and returns)
	OpReturn

	// OpClosure creates a closure with captured variables.
	// Format: [OpClosure][upvalueCount][funcIndex]
	// Stack: [upvalue1, upvalue2, ..., upvalueN] -> [closure]
	OpClosure

	// OpCallBuiltin calls a built-in function.
	// Format: [OpCallBuiltin][argCount][builtinIndex]
	// Stack: [arg1, arg2, ..., argN] -> [returnValue]
	OpCallBuiltin

	// OpCallIndirect calls a function pointer/lambda.
	// Format: [OpCallIndirect][argCount][unused]
	// Stack: [funcPtr, arg1, arg2, ..., argN] -> [returnValue]
	OpCallIndirect

	// OpTailCall performs tail call optimization.
	// Format: [OpTailCall][argCount][funcIndex]
	// Stack: [arg1, arg2, ..., argN] -> [returnValue]
	OpTailCall

	// ========================================
	// Stack Operations (5 opcodes)
	// ========================================

	// OpPop pops and discards top of stack.
	// Format: [OpPop][unused][unused]
	// Stack: [value] -> []
	OpPop

	// OpDup duplicates top of stack.
	// Format: [OpDup][unused][unused]
	// Stack: [value] -> [value, value]
	OpDup

	// OpDup2 duplicates top two stack values.
	// Format: [OpDup2][unused][unused]
	// Stack: [a, b] -> [a, b, a, b]
	OpDup2

	// OpSwap swaps top two stack values.
	// Format: [OpSwap][unused][unused]
	// Stack: [a, b] -> [b, a]
	OpSwap

	// OpRotate3 rotates top three stack values.
	// Format: [OpRotate3][unused][unused]
	// Stack: [a, b, c] -> [b, c, a]
	OpRotate3

	// ========================================
	// Array Operations (8 opcodes)
	// ========================================

	// OpNewArray creates a new array.
	// Format: [OpNewArray][unused][elementCount]
	// Stack: [elem1, elem2, ..., elemN] -> [array]
	OpNewArray

	// OpNewArraySized creates a new array with size.
	// Format: [OpNewArraySized][unused][typeIndex]
	// Stack: [size] -> [array]
	// typeIndex: constant pool index for element type name (for zero-value initialization)
	OpNewArraySized

	// OpNewArrayMultiDim creates a multi-dimensional array.
	// Format: [OpNewArrayMultiDim][dimCount][typeIndex]
	// Stack: [dim1, dim2, ..., dimN] -> [array]
	// Creates nested arrays: new T[d1, d2] creates array of d1 elements, each is array of d2 elements
	// typeIndex: constant pool index for element type name (for zero-value initialization)
	OpNewArrayMultiDim

	// OpArrayLength pushes array length onto stack.
	// Format: [OpArrayLength][unused][unused]
	// Stack: [array] -> [length]
	OpArrayLength

	// OpArrayGet loads array element.
	// Format: [OpArrayGet][unused][unused]
	// Stack: [array, index] -> [element]
	OpArrayGet

	// OpArraySet stores to array element.
	// Format: [OpArraySet][unused][unused]
	// Stack: [array, index, value] -> []
	OpArraySet

	// OpArraySetLength sets array length (dynamic arrays).
	// Format: [OpArraySetLength][unused][unused]
	// Stack: [array, newLength] -> []
	OpArraySetLength

	// OpArrayHigh returns high bound of array.
	// Format: [OpArrayHigh][unused][unused]
	// Stack: [array] -> [highIndex]
	OpArrayHigh

	// OpArrayLow returns low bound of array.
	// Format: [OpArrayLow][unused][unused]
	// Stack: [array] -> [lowIndex]
	OpArrayLow

	// OpArrayCount returns array count (alias for length).
	// Format: [OpArrayCount][unused][unused]
	// Stack: [array] -> [count]
	OpArrayCount

	// OpArrayDelete removes elements from array.
	// Format: [OpArrayDelete][unused][unused]
	// Stack: [array, index, count] -> []
	OpArrayDelete

	// OpArrayIndexOf finds element index in array.
	// Format: [OpArrayIndexOf][unused][unused]
	// Stack: [array, value, startIndex] -> [index]
	OpArrayIndexOf

	// ========================================
	// Set Operations (1 opcode)
	// ========================================

	// OpNewSet creates a new set from elements on stack.
	// Format: [OpNewSet][unused][elementCount]
	// Stack: [elem1, elem2, ..., elemN] -> [set]
	// Elements can be individual values or expanded from ranges
	OpNewSet

	// ========================================
	// Object Operations (10 opcodes)
	// ========================================

	// OpNewObject creates a new object instance.
	// Format: [OpNewObject][unused][classIndex]
	// Stack: [] -> [object]
	OpNewObject

	// OpNewRecord creates a new record instance (Task 9.7).
	// Format: [OpNewRecord][unused][typeIndex]
	// Stack: [] -> [record]
	OpNewRecord

	// OpGetField loads object/record field.
	// Format: [OpGetField][unused][fieldIndex]
	// Stack: [object|record] -> [value]
	OpGetField

	// OpSetField stores to object/record field.
	// Format: [OpSetField][unused][fieldIndex]
	// Stack: [object|record, value] -> []
	OpSetField

	// OpGetProperty calls property getter.
	// Format: [OpGetProperty][unused][propertyIndex]
	// Stack: [object] -> [value]
	OpGetProperty

	// OpSetProperty calls property setter.
	// Format: [OpSetProperty][unused][propertyIndex]
	// Stack: [object, value] -> []
	OpSetProperty

	// OpGetClass gets object's class.
	// Format: [OpGetClass][unused][unused]
	// Stack: [object] -> [class]
	OpGetClass

	// OpInstanceOf checks if object is instance of class.
	// Format: [OpInstanceOf][unused][classIndex]
	// Stack: [object] -> [boolean]
	OpInstanceOf

	// OpCastObject casts object to class (with check).
	// Format: [OpCastObject][unused][classIndex]
	// Stack: [object] -> [object] or raises exception
	OpCastObject

	// OpGetSelf pushes current 'Self' reference onto stack.
	// Format: [OpGetSelf][unused][unused]
	// Stack: [] -> [self]
	OpGetSelf

	// OpInvoke invokes method with dynamic dispatch.
	// Format: [OpInvoke][argCount][methodNameIndex]
	// Stack: [object, arg1, ..., argN] -> [returnValue]
	OpInvoke

	// ========================================
	// String Operations (4 opcodes)
	// ========================================

	// OpStringConcat concatenates strings.
	// Format: [OpStringConcat][unused][unused]
	// Stack: [str1, str2] -> [str1 + str2]
	OpStringConcat

	// OpStringLength returns string length.
	// Format: [OpStringLength][unused][unused]
	// Stack: [string] -> [length]
	OpStringLength

	// OpStringGet gets character at index.
	// Format: [OpStringGet][unused][unused]
	// Stack: [string, index] -> [char]
	OpStringGet

	// OpStringSlice extracts substring.
	// Format: [OpStringSlice][unused][unused]
	// Stack: [string, start, length] -> [substring]
	OpStringSlice

	// ========================================
	// Type Operations (7 opcodes)
	// ========================================

	// OpIntToFloat converts integer to float.
	// Format: [OpIntToFloat][unused][unused]
	// Stack: [int] -> [float]
	OpIntToFloat

	// OpFloatToInt converts float to integer (rounds to nearest).
	// Format: [OpFloatToInt][unused][unused]
	// Stack: [float] -> [int]
	OpFloatToInt

	// OpIntToString converts integer to string.
	// Format: [OpIntToString][unused][unused]
	// Stack: [int] -> [string]
	OpIntToString

	// OpFloatToString converts float to string.
	// Format: [OpFloatToString][unused][unused]
	// Stack: [float] -> [string]
	OpFloatToString

	// OpBoolToString converts boolean to string.
	// Format: [OpBoolToString][unused][unused]
	// Stack: [bool] -> [string]
	OpBoolToString

	// OpVariantToType converts variant to specific type.
	// Format: [OpVariantToType][unused][typeIndex]
	// Stack: [variant] -> [typedValue]
	OpVariantToType

	// OpToBool converts any value to boolean.
	// Format: [OpToBool][unused][unused]
	// Stack: [value] -> [boolean]
	// Note: Uses same conversion rules as isTruthy() in interpreter
	OpToBool

	// ========================================
	// Exception Handling (4 opcodes)
	// ========================================

	// OpTry begins try block.
	// Format: [OpTry][unused][catchOffset]
	// Stack: [] -> []
	OpTry

	// OpCatch begins catch block.
	// Format: [OpCatch][unused][finallyOffset]
	// Stack: [exception] -> []
	OpCatch

	// OpFinally begins finally block.
	// Format: [OpFinally][unused][unused]
	// Stack: [] -> []
	OpFinally

	// OpThrow throws an exception.
	// Format: [OpThrow][unused][unused]
	// Stack: [exception] -> (unwinds stack)
	OpThrow

	// ========================================
	// Miscellaneous (4 opcodes)
	// ========================================

	// OpHalt terminates program execution.
	// Format: [OpHalt][unused][unused]
	// Stack: [] -> []
	OpHalt

	// OpPrint prints value (debug/output).
	// Format: [OpPrint][unused][unused]
	// Stack: [value] -> []
	OpPrint

	// OpAssert checks assertion.
	// Format: [OpAssert][unused][unused]
	// Stack: [condition] -> [] or raises assertion failure
	OpAssert

	// OpDebugger triggers debugger breakpoint (if debugger attached).
	// Format: [OpDebugger][unused][unused]
	// Stack: [] -> []
	OpDebugger

	// Total: 115 opcodes
	// This keeps us well under 128 (7 binary search comparisons in Go switch)
)

// OpCodeNames maps opcodes to their string names for debugging and disassembly.
var OpCodeNames = [...]string{
	OpLoadConst:        "LOAD_CONST",
	OpLoadConst0:       "LOAD_CONST_0",
	OpLoadConst1:       "LOAD_CONST_1",
	OpLoadLocal:        "LOAD_LOCAL",
	OpStoreLocal:       "STORE_LOCAL",
	OpLoadGlobal:       "LOAD_GLOBAL",
	OpStoreGlobal:      "STORE_GLOBAL",
	OpLoadUpvalue:      "LOAD_UPVALUE",
	OpStoreUpvalue:     "STORE_UPVALUE",
	OpLoadNil:          "LOAD_NIL",
	OpLoadTrue:         "LOAD_TRUE",
	OpLoadFalse:        "LOAD_FALSE",
	OpAddInt:           "ADD_INT",
	OpSubInt:           "SUB_INT",
	OpMulInt:           "MUL_INT",
	OpDivInt:           "DIV_INT",
	OpModInt:           "MOD_INT",
	OpNegateInt:        "NEGATE_INT",
	OpIncInt:           "INC_INT",
	OpDecInt:           "DEC_INT",
	OpBitAnd:           "BIT_AND",
	OpBitOr:            "BIT_OR",
	OpBitXor:           "BIT_XOR",
	OpBitNot:           "BIT_NOT",
	OpAddFloat:         "ADD_FLOAT",
	OpSubFloat:         "SUB_FLOAT",
	OpMulFloat:         "MUL_FLOAT",
	OpDivFloat:         "DIV_FLOAT",
	OpNegateFloat:      "NEGATE_FLOAT",
	OpPower:            "POWER",
	OpSqrt:             "SQRT",
	OpAbs:              "ABS",
	OpFloor:            "FLOOR",
	OpCeil:             "CEIL",
	OpRound:            "ROUND",
	OpTrunc:            "TRUNC",
	OpEqual:            "EQUAL",
	OpNotEqual:         "NOT_EQUAL",
	OpLess:             "LESS",
	OpLessEqual:        "LESS_EQUAL",
	OpGreater:          "GREATER",
	OpGreaterEqual:     "GREATER_EQUAL",
	OpCompareInt:       "COMPARE_INT",
	OpCompareFloat:     "COMPARE_FLOAT",
	OpNot:              "NOT",
	OpAnd:              "AND",
	OpOr:               "OR",
	OpXor:              "XOR",
	OpIsFalsey:         "IS_FALSEY",
	OpShl:              "SHL",
	OpShr:              "SHR",
	OpSar:              "SAR",
	OpRotl:             "ROTL",
	OpJump:             "JUMP",
	OpJumpIfTrue:       "JUMP_IF_TRUE",
	OpJumpIfFalse:      "JUMP_IF_FALSE",
	OpJumpIfTrueNoPop:  "JUMP_IF_TRUE_NO_POP",
	OpJumpIfFalseNoPop: "JUMP_IF_FALSE_NO_POP",
	OpLoop:             "LOOP",
	OpForPrep:          "FOR_PREP",
	OpForLoop:          "FOR_LOOP",
	OpCase:             "CASE",
	OpBreak:            "BREAK",
	OpContinue:         "CONTINUE",
	OpExit:             "EXIT",
	OpCall:             "CALL",
	OpCallMethod:       "CALL_METHOD",
	OpCallVirtual:      "CALL_VIRTUAL",
	OpReturn:           "RETURN",
	OpClosure:          "CLOSURE",
	OpCallBuiltin:      "CALL_BUILTIN",
	OpCallIndirect:     "CALL_INDIRECT",
	OpTailCall:         "TAIL_CALL",
	OpPop:              "POP",
	OpDup:              "DUP",
	OpDup2:             "DUP2",
	OpSwap:             "SWAP",
	OpRotate3:          "ROTATE3",
	OpNewArray:         "NEW_ARRAY",
	OpNewArraySized:    "NEW_ARRAY_SIZED",
	OpNewArrayMultiDim: "NEW_ARRAY_MULTIDIM",
	OpArrayLength:      "ARRAY_LENGTH",
	OpArrayGet:         "ARRAY_GET",
	OpArraySet:         "ARRAY_SET",
	OpArraySetLength:   "ARRAY_SET_LENGTH",
	OpArrayHigh:        "ARRAY_HIGH",
	OpArrayLow:         "ARRAY_LOW",
	OpArrayCount:       "ARRAY_COUNT",
	OpArrayDelete:      "ARRAY_DELETE",
	OpArrayIndexOf:     "ARRAY_INDEX_OF",
	OpNewSet:           "NEW_SET",
	OpNewObject:        "NEW_OBJECT",
	OpNewRecord:        "NEW_RECORD", // Task 9.7
	OpGetField:         "GET_FIELD",
	OpSetField:         "SET_FIELD",
	OpGetProperty:      "GET_PROPERTY",
	OpSetProperty:      "SET_PROPERTY",
	OpGetClass:         "GET_CLASS",
	OpInstanceOf:       "INSTANCE_OF",
	OpCastObject:       "CAST_OBJECT",
	OpGetSelf:          "GET_SELF",
	OpInvoke:           "INVOKE",
	OpStringConcat:     "STRING_CONCAT",
	OpStringLength:     "STRING_LENGTH",
	OpStringGet:        "STRING_GET",
	OpStringSlice:      "STRING_SLICE",
	OpIntToFloat:       "INT_TO_FLOAT",
	OpFloatToInt:       "FLOAT_TO_INT",
	OpIntToString:      "INT_TO_STRING",
	OpFloatToString:    "FLOAT_TO_STRING",
	OpBoolToString:     "BOOL_TO_STRING",
	OpVariantToType:    "VARIANT_TO_TYPE",
	OpToBool:           "TO_BOOL",
	OpTry:              "TRY",
	OpCatch:            "CATCH",
	OpFinally:          "FINALLY",
	OpThrow:            "THROW",
	OpHalt:             "HALT",
	OpPrint:            "PRINT",
	OpAssert:           "ASSERT",
	OpDebugger:         "DEBUGGER",
}

// Instruction represents a single bytecode instruction.
// Format: 32-bit instruction [8-bit opcode][8-bit A][16-bit B]
type Instruction uint32

// MakeInstruction creates a new instruction from opcode and operands.
func MakeInstruction(op OpCode, a byte, b uint16) Instruction {
	return Instruction(uint32(op) | uint32(a)<<8 | uint32(b)<<16)
}

// MakeSimpleInstruction creates an instruction with no operands.
func MakeSimpleInstruction(op OpCode) Instruction {
	return Instruction(op)
}

// MakeInstructionABC creates an instruction with three byte operands.
func MakeInstructionABC(op OpCode, a, b, c byte) Instruction {
	return Instruction(uint32(op) | uint32(a)<<8 | uint32(b)<<16 | uint32(c)<<24)
}

// OpCode returns the opcode of this instruction.
func (inst Instruction) OpCode() OpCode {
	return OpCode(inst & 0xFF)
}

// A returns the A operand (8 bits).
func (inst Instruction) A() byte {
	return byte((inst >> 8) & 0xFF)
}

// B returns the B operand (16 bits).
func (inst Instruction) B() uint16 {
	return uint16((inst >> 16) & 0xFFFF)
}

// SignedB returns the B operand as a signed 16-bit integer (for jump offsets).
func (inst Instruction) SignedB() int16 {
	return int16(inst.B())
}

// C returns the C operand (8 bits, from high byte of B field).
func (inst Instruction) C() byte {
	return byte((inst >> 24) & 0xFF)
}

// String returns a human-readable string representation of the instruction.
func (inst Instruction) String() string {
	op := inst.OpCode()
	if int(op) < len(OpCodeNames) && OpCodeNames[op] != "" {
		return OpCodeNames[op]
	}
	return "UNKNOWN"
}
