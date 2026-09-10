package interp

import "testing"

// ============================================================================
// ByteBuffer: the built-in raw byte block.
//
// The subtle parts of the surface are its value/reference behaviour and the
// two arities every typed accessor carries. The FunctionsByteBuffer fixture
// corpus covers the accessors exhaustively; these tests pin the semantics that
// a reader of the implementation would most easily get wrong.
// ============================================================================

func TestByteBuffer_DeclaredVariableIsAutoInstantiated(t *testing.T) {
	runScriptTestWithSemantic(t, `
		var b : ByteBuffer;
		PrintLn(b.Length);
		PrintLn(b.ToJSON);
	`, "0\n[]")
}

func TestByteBuffer_AssignmentAliasesStorage(t *testing.T) {
	runScriptTestWithSemantic(t, `
		var a : ByteBuffer;
		a.SetLength(3);
		var b := new ByteBuffer;
		a := b;
		b.SetLength(2);
		PrintLn(a.Length);
	`, "2")
}

func TestByteBuffer_AssignMethodCopies(t *testing.T) {
	runScriptTestWithSemantic(t, `
		var a := ByteBuffer('hello');
		var b : ByteBuffer;
		b.Assign(a);
		a.SetLength(2);
		PrintLn(a.ToDataString);
		PrintLn(b.ToDataString);
	`, "he\nhello")
}

func TestByteBuffer_CursorAndIndexedAccessors(t *testing.T) {
	runScriptTestWithSemantic(t, `
		var b : ByteBuffer;
		b.SetLength(4);
		b.SetWord(258);
		PrintLn(b.Position);
		b.SetWord(2, 772);
		PrintLn(b.Position);
		PrintLn(b.ToJSON);
		PrintLn(b.GetWord(0));
		PrintLn(b.GetWord(2));
	`, "2\n2\n[2,1,4,3]\n258\n772")
}

func TestByteBuffer_StringCastKeepsLowByteOfEachUnit(t *testing.T) {
	runScriptTestWithSemantic(t, `
		PrintLn(ByteBuffer('hello').ToHexString);
		PrintLn(ByteBuffer(#$1234#$5678).ToHexString);
	`, "68656c6c6f\n3478")
}

func TestByteBuffer_RangeAndOverflowAreCatchable(t *testing.T) {
	runScriptTestWithSemantic(t, `
		var b : ByteBuffer;
		try
			b.GetByte(0);
		except
			on E : Exception do PrintLn(E.Message);
		end;
		b.SetLength(1);
		try
			b.SetByte(256);
		except
			on E : Exception do PrintLn(E.Message);
		end;
	`, "Out of range (index 0, size 1 for length 0) [line: 4, column: 6]\n"+
		"value 256 out of Byte range [line: 10, column: 6]")
}
