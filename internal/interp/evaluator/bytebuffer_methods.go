package evaluator

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// byteBufferReceiver unwraps a value to the ByteBuffer it denotes.
func byteBufferReceiver(value Value) (*runtime.ByteBufferValue, bool) {
	buffer, ok := value.(*runtime.ByteBufferValue)
	return buffer, ok
}

// raiseByteBufferError turns a ByteBuffer domain error into a catchable script
// Exception positioned at the member name, matching DWScript's diagnostics.
func (e *Evaluator) raiseByteBufferError(node ast.Node, err error, ctx *ExecutionContext) Value {
	pos := arrayMethodNamePos(node)
	message := fmt.Sprintf("%s [line: %d, column: %d]", err.Error(), pos.Line, pos.Column)
	if ctx != nil {
		exc := e.createException("Exception", message, &pos, ctx)
		ctx.SetException(exc)
	}
	return e.nilValue()
}

// byteBufferIntArg coerces the i-th argument to an Integer.
func byteBufferIntArg(args []Value, i int) (int64, bool) {
	if i >= len(args) {
		return 0, false
	}
	value, err := runtime.ToInteger(args[i])
	if err != nil {
		return 0, false
	}
	return value, true
}

// byteBufferFloatArg coerces the i-th argument to a Float.
func byteBufferFloatArg(args []Value, i int) (float64, bool) {
	if i >= len(args) {
		return 0, false
	}
	value, err := runtime.ToFloat(args[i])
	if err != nil {
		return 0, false
	}
	return value, true
}

// byteBufferStringArg coerces the i-th argument to a String.
func byteBufferStringArg(args []Value, i int) (string, bool) {
	if i >= len(args) {
		return "", false
	}
	if str, ok := args[i].(*runtime.StringValue); ok {
		return str.Value, true
	}
	return runtime.ToString(args[i]), true
}

// byteBufferBoolArg coerces the i-th argument to a Boolean.
func byteBufferBoolArg(args []Value, i int) (bool, bool) {
	if i >= len(args) {
		return false, false
	}
	value, err := runtime.ToBoolean(args[i])
	if err != nil {
		return false, false
	}
	return value, true
}

// DispatchByteBufferMethod executes a ByteBuffer intrinsic. It serves both the
// method-call form (b.SetLength(4)) and the parameterless member form
// (b.ToJSON), which DWScript treats interchangeably.
//
// Accessors come in two arities: the cursor form (b.GetByte, b.SetByte(v))
// reads or writes at Position and advances it, while the indexed form
// (b.GetByte(i), b.SetByte(i, v)) addresses the buffer directly and leaves the
// cursor alone.
func (e *Evaluator) DispatchByteBufferMethod(receiver Value, methodName string, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	buffer, ok := byteBufferReceiver(receiver)
	if !ok {
		return e.newError(node, "ByteBuffer method '%s' requires a ByteBuffer receiver", methodName)
	}

	normalized := ident.Normalize(methodName)

	if result, handled := e.byteBufferTypedAccessor(buffer, normalized, args, node, ctx); handled {
		return result
	}
	if result, handled := e.byteBufferShapeMember(buffer, normalized, args, node, ctx); handled {
		return result
	}
	if result, handled := e.byteBufferBulkMember(buffer, normalized, args, node, ctx); handled {
		return result
	}

	return e.newError(node, "method '%s' not found for type 'ByteBuffer'", methodName)
}

// byteBufferTypedAccessor handles the Get<T>/Set<T> family for the named
// integer and floating point widths. It reports whether it recognised the
// member.
func (e *Evaluator) byteBufferTypedAccessor(buffer *runtime.ByteBufferValue, normalized string, args []Value, node ast.Node, ctx *ExecutionContext) (Value, bool) {
	if suffix, isGetter := byteBufferAccessorSuffix(normalized, "get"); isGetter {
		if spec, known := byteBufferIntegerAccessor(suffix); known {
			return e.byteBufferGetInteger(buffer, spec, suffix, args, node, ctx), true
		}
		if byteBufferIsFloatAccessor(suffix) {
			return e.byteBufferGetFloat(buffer, suffix, args, node, ctx), true
		}
	}
	if suffix, isSetter := byteBufferAccessorSuffix(normalized, "set"); isSetter {
		if _, known := byteBufferIntegerAccessor(suffix); known {
			return e.byteBufferSetInteger(buffer, suffix, args, node, ctx), true
		}
		if byteBufferIsFloatAccessor(suffix) {
			return e.byteBufferSetFloat(buffer, suffix, args, node, ctx), true
		}
	}
	return nil, false
}

// byteBufferBulkMember handles the members that move whole runs of bytes:
// GetData, SetData, GetIntegers, the Assign family and Copy. It reports whether
// it recognised the member.
func (e *Evaluator) byteBufferBulkMember(buffer *runtime.ByteBufferValue, normalized string, args []Value, node ast.Node, ctx *ExecutionContext) (Value, bool) {
	switch normalized {
	case "getdata":
		index, okIndex := byteBufferIntArg(args, 0)
		size, okSize := byteBufferIntArg(args, 1)
		if !okIndex || !okSize {
			return e.newError(node, "ByteBuffer.GetData expects (index, size)"), true
		}
		data, err := buffer.GetDataAt(int(index), int(size))
		if err != nil {
			return e.raiseByteBufferError(node, err, ctx), true
		}
		return &runtime.StringValue{Value: data}, true

	case "setdata":
		return e.byteBufferSetData(buffer, args, node, ctx), true

	case "getintegers":
		return e.byteBufferGetIntegers(buffer, args, node, ctx), true

	case "assign":
		if len(args) != 1 {
			return e.newError(node, "ByteBuffer.Assign expects exactly 1 argument"), true
		}
		source, ok := byteBufferReceiver(args[0])
		if !ok {
			return e.newError(node, "ByteBuffer.Assign expects a ByteBuffer argument"), true
		}
		buffer.Assign(source)
		return e.nilValue(), true

	case "assigndatastring", "assignbase64", "assignjson", "assignhexstring":
		return e.byteBufferAssignString(buffer, normalized, args, node, ctx), true

	case "copy":
		return e.byteBufferCopy(buffer, args, node), true
	}
	return nil, false
}

// byteBufferCopy implements Copy, Copy(index) and Copy(index, count). An
// omitted count means "to the end"; the range is clamped to the buffer.
func (e *Evaluator) byteBufferCopy(buffer *runtime.ByteBufferValue, args []Value, node ast.Node) Value {
	if len(args) > 2 {
		return e.newError(node, "ByteBuffer.Copy expects 0 to 2 arguments, got %d", len(args))
	}
	var index int64
	if len(args) >= 1 {
		// An omitted index means 0, but a supplied argument that is not an
		// Integer is an error rather than a silent fallback to Copy(0).
		explicit, ok := byteBufferIntArg(args, 0)
		if !ok {
			return e.newError(node, "ByteBuffer.Copy expects Integer arguments")
		}
		index = explicit
	}
	count := int64(buffer.Length())
	if len(args) >= 2 {
		explicit, ok := byteBufferIntArg(args, 1)
		if !ok {
			return e.newError(node, "ByteBuffer.Copy expects Integer arguments")
		}
		count = explicit
	}
	return buffer.Copy(int(index), int(count))
}

// byteBufferShapeMember handles the members that describe or reshape the buffer
// as a whole: Length, Position, SetLength, SetPosition and the To* renderings.
// It reports whether it recognised the member.
func (e *Evaluator) byteBufferShapeMember(buffer *runtime.ByteBufferValue, normalized string, args []Value, node ast.Node, ctx *ExecutionContext) (Value, bool) {
	switch normalized {
	case "length":
		return &runtime.IntegerValue{Value: int64(buffer.Length())}, true

	case "position":
		return &runtime.IntegerValue{Value: int64(buffer.Position())}, true

	case "setlength":
		length, ok := byteBufferIntArg(args, 0)
		if !ok {
			return e.newError(node, "ByteBuffer.SetLength expects an Integer argument"), true
		}
		if length > runtime.MaxByteBufferLength {
			// Turn an unsatisfiable allocation into a catchable script error
			// rather than letting make() panic the host.
			return e.raiseByteBufferError(node, runtime.NewByteBufferLengthError(length), ctx), true
		}
		buffer.SetLength(int(length))
		return e.nilValue(), true

	case "setposition":
		position, ok := byteBufferIntArg(args, 0)
		if !ok {
			return e.newError(node, "ByteBuffer.SetPosition expects an Integer argument"), true
		}
		if err := buffer.SetPosition(int(position)); err != nil {
			return e.raiseByteBufferError(node, err, ctx), true
		}
		return e.nilValue(), true

	case "tojson":
		return &runtime.StringValue{Value: buffer.ToJSON()}, true
	case "todatastring":
		return &runtime.StringValue{Value: buffer.ToDataString()}, true
	case "tobase64":
		return &runtime.StringValue{Value: buffer.ToBase64()}, true
	case "tohexstring":
		return &runtime.StringValue{Value: buffer.ToHexString()}, true
	}
	return nil, false
}

// byteBufferAccessorSuffix splits a normalized member name into its Get/Set
// prefix and the accessor suffix (byte, int32, single, ...).
func byteBufferAccessorSuffix(normalized, prefix string) (string, bool) {
	if !strings.HasPrefix(normalized, prefix) {
		return "", false
	}
	return normalized[len(prefix):], true
}

// byteBufferIntegerAccessor reports the width and signedness of a named integer
// accessor suffix.
func byteBufferIntegerAccessor(suffix string) (runtime.ByteBufferIntSpec, bool) {
	return runtime.ByteBufferIntegerSpec(suffix)
}

// byteBufferIsFloatAccessor reports whether suffix names a float accessor.
func byteBufferIsFloatAccessor(suffix string) bool {
	return runtime.ByteBufferIsFloatAccessor(suffix)
}

// byteBufferGetInteger implements the Get<IntegerType> accessors in both the
// cursor and the indexed form.
func (e *Evaluator) byteBufferGetInteger(buffer *runtime.ByteBufferValue, spec runtime.ByteBufferIntSpec, suffix string, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	var (
		value int64
		err   error
	)
	if len(args) > 1 {
		// DWScript declares the getters as two overloads, () and (index), so a
		// third argument has no signature to bind to.
		return e.newError(node, "ByteBuffer.Get%s expects 0 or 1 arguments, got %d", suffix, len(args))
	}
	if len(args) == 0 {
		value, err = buffer.GetInt(spec.Size, spec.Signed)
	} else {
		index, ok := byteBufferIntArg(args, 0)
		if !ok {
			return e.newError(node, "ByteBuffer.Get%s expects an Integer index", suffix)
		}
		value, err = buffer.GetIntAt(int(index), spec.Size, spec.Signed)
	}
	if err != nil {
		return e.raiseByteBufferError(node, err, ctx)
	}
	return &runtime.IntegerValue{Value: value}
}

// byteBufferSetInteger implements the Set<IntegerType> accessors in both the
// cursor form (one argument) and the indexed form (index and value).
func (e *Evaluator) byteBufferSetInteger(buffer *runtime.ByteBufferValue, suffix string, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	var err error
	switch len(args) {
	case 1:
		value, ok := byteBufferIntArg(args, 0)
		if !ok {
			return e.newError(node, "ByteBuffer.Set%s expects an Integer value", suffix)
		}
		err = buffer.SetInt(suffix, value)
	case 2:
		index, okIndex := byteBufferIntArg(args, 0)
		value, okValue := byteBufferIntArg(args, 1)
		if !okIndex || !okValue {
			return e.newError(node, "ByteBuffer.Set%s expects Integer arguments", suffix)
		}
		err = buffer.SetIntAt(suffix, int(index), value)
	default:
		return e.newError(node, "ByteBuffer.Set%s expects 1 or 2 arguments, got %d", suffix, len(args))
	}
	if err != nil {
		return e.raiseByteBufferError(node, err, ctx)
	}
	return e.nilValue()
}

// byteBufferGetFloat implements GetSingle/GetDouble/GetExtended.
func (e *Evaluator) byteBufferGetFloat(buffer *runtime.ByteBufferValue, suffix string, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	var (
		value float64
		err   error
	)
	if len(args) > 1 {
		return e.newError(node, "ByteBuffer.Get%s expects 0 or 1 arguments, got %d", suffix, len(args))
	}
	if len(args) == 0 {
		value, err = buffer.GetFloat(suffix)
	} else {
		index, ok := byteBufferIntArg(args, 0)
		if !ok {
			return e.newError(node, "ByteBuffer.Get%s expects an Integer index", suffix)
		}
		value, err = buffer.GetFloatAt(suffix, int(index))
	}
	if err != nil {
		return e.raiseByteBufferError(node, err, ctx)
	}
	return &runtime.FloatValue{Value: value}
}

// byteBufferSetFloat implements SetSingle/SetDouble/SetExtended.
func (e *Evaluator) byteBufferSetFloat(buffer *runtime.ByteBufferValue, suffix string, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	var err error
	switch len(args) {
	case 1:
		value, ok := byteBufferFloatArg(args, 0)
		if !ok {
			return e.newError(node, "ByteBuffer.Set%s expects a Float value", suffix)
		}
		err = buffer.SetFloat(suffix, value)
	case 2:
		index, okIndex := byteBufferIntArg(args, 0)
		value, okValue := byteBufferFloatArg(args, 1)
		if !okIndex || !okValue {
			return e.newError(node, "ByteBuffer.Set%s expects (index, value)", suffix)
		}
		err = buffer.SetFloatAt(suffix, int(index), value)
	default:
		return e.newError(node, "ByteBuffer.Set%s expects 1 or 2 arguments, got %d", suffix, len(args))
	}
	if err != nil {
		return e.raiseByteBufferError(node, err, ctx)
	}
	return e.nilValue()
}

// byteBufferSetData implements SetData in both the cursor and indexed forms.
func (e *Evaluator) byteBufferSetData(buffer *runtime.ByteBufferValue, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	var err error
	switch len(args) {
	case 1:
		data, ok := byteBufferStringArg(args, 0)
		if !ok {
			return e.newError(node, "ByteBuffer.SetData expects a String value")
		}
		err = buffer.SetData(data)
	case 2:
		index, okIndex := byteBufferIntArg(args, 0)
		data, okData := byteBufferStringArg(args, 1)
		if !okIndex || !okData {
			return e.newError(node, "ByteBuffer.SetData expects (index, value)")
		}
		err = buffer.SetDataAt(int(index), data)
	default:
		return e.newError(node, "ByteBuffer.SetData expects 1 or 2 arguments, got %d", len(args))
	}
	if err != nil {
		return e.raiseByteBufferError(node, err, ctx)
	}
	return e.nilValue()
}

// byteBufferGetIntegers implements GetIntegers(index, count, size, signed).
func (e *Evaluator) byteBufferGetIntegers(buffer *runtime.ByteBufferValue, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	if len(args) != 4 {
		return e.newError(node, "ByteBuffer.GetIntegers expects 4 arguments, got %d", len(args))
	}
	index, okIndex := byteBufferIntArg(args, 0)
	count, okCount := byteBufferIntArg(args, 1)
	size, okSize := byteBufferIntArg(args, 2)
	signed, okSigned := byteBufferBoolArg(args, 3)
	if !okIndex || !okCount || !okSize || !okSigned {
		return e.newError(node, "ByteBuffer.GetIntegers expects (index, count, size, signed)")
	}
	values, err := buffer.GetIntegers(int(index), int(count), int(size), signed)
	if err != nil {
		return e.raiseByteBufferError(node, err, ctx)
	}
	elements := make([]Value, len(values))
	for i, value := range values {
		elements[i] = &runtime.IntegerValue{Value: value}
	}
	return &runtime.ArrayValue{
		ArrayType: types.NewDynamicArrayType(types.INTEGER),
		Elements:  elements,
	}
}

// byteBufferAssignString implements the AssignDataString / AssignBase64 /
// AssignJSON / AssignHexString family.
func (e *Evaluator) byteBufferAssignString(buffer *runtime.ByteBufferValue, normalized string, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	data, ok := byteBufferStringArg(args, 0)
	if !ok {
		return e.newError(node, "ByteBuffer.%s expects a String argument", normalized)
	}
	var err error
	switch normalized {
	case "assigndatastring":
		buffer.AssignDataString(data)
	case "assignbase64":
		err = buffer.AssignBase64(data)
	case "assignjson":
		err = buffer.AssignJSON(data)
	case "assignhexstring":
		err = buffer.AssignHexString(data)
	}
	if err != nil {
		return e.raiseByteBufferError(node, err, ctx)
	}
	return e.nilValue()
}
