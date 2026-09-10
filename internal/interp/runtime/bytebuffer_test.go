package runtime

import (
	"math"
	"testing"
)

func TestByteBufferValue_SetLengthZeroFills(t *testing.T) {
	b := NewByteBufferValue()
	if b.Length() != 0 {
		t.Fatalf("fresh buffer length = %d, want 0", b.Length())
	}

	b.SetLength(4)
	if err := b.SetIntAt("int32", 0, 0x12345678); err != nil {
		t.Fatalf("SetIntAt: %v", err)
	}
	b.SetLength(0)
	b.SetLength(4)

	got, err := b.GetIntAt(0, 4, true)
	if err != nil {
		t.Fatalf("GetIntAt: %v", err)
	}
	if got != 0 {
		t.Fatalf("regrown buffer = %d, want 0 (must be zero filled)", got)
	}
}

func TestByteBufferValue_SetLengthPreservesPrefix(t *testing.T) {
	b := NewByteBufferValue()
	b.SetLength(2)
	if err := b.SetInt("byte", 6); err != nil {
		t.Fatalf("SetInt: %v", err)
	}
	if err := b.SetInt("byte", 9); err != nil {
		t.Fatalf("SetInt: %v", err)
	}

	tests := []struct {
		length int
		want   string
	}{
		{2, "[6,9]"},
		{1, "[6]"},
		{2, "[6,0]"},
		{0, "[]"},
	}
	for _, tt := range tests {
		b.SetLength(tt.length)
		if got := b.ToJSON(); got != tt.want {
			t.Errorf("SetLength(%d) -> ToJSON = %q, want %q", tt.length, got, tt.want)
		}
	}
}

func TestByteBufferValue_SetPositionRange(t *testing.T) {
	b := NewByteBufferValue()
	b.SetLength(1)

	tests := []struct {
		pos     int
		wantErr string
	}{
		{0, ""},
		{1, ""},
		{-1, "Position -1 out of range (length 1)"},
		{2, "Position 2 out of range (length 1)"},
	}
	for _, tt := range tests {
		err := b.SetPosition(tt.pos)
		switch {
		case tt.wantErr == "" && err != nil:
			t.Errorf("SetPosition(%d) = %v, want nil", tt.pos, err)
		case tt.wantErr != "" && (err == nil || err.Error() != tt.wantErr):
			t.Errorf("SetPosition(%d) = %v, want %q", tt.pos, err, tt.wantErr)
		}
	}
}

func TestByteBufferValue_IntegerOverflowDiagnostics(t *testing.T) {
	tests := []struct {
		suffix string
		value  int64
		want   string
	}{
		{"byte", -1, "value -1 out of Byte range"},
		{"byte", 256, "value 256 out of Byte range"},
		{"int8", -129, "value -129 out of Int8 range"},
		{"int8", 128, "value 128 out of Int8 range"},
		{"word", -1, "value -1 out of Word range"},
		{"word", 0x10000, "value 65536 out of Word range"},
		{"int16", -0x8001, "value -32769 out of Int16 range"},
		{"int16", 0x8000, "value 32768 out of Int16 range"},
		{"dword", -1, "value -1 out of DWord range"},
		{"dword", 0x100000000, "value 4294967296 out of DWord range"},
		{"int32", -0x80000001, "value -2147483649 out of Int32 range"},
		{"int32", 0x80000000, "value 2147483648 out of Int32 range"},
	}
	for _, tt := range tests {
		b := NewByteBufferValue()
		b.SetLength(8)
		err := b.SetInt(tt.suffix, tt.value)
		if err == nil || err.Error() != tt.want {
			t.Errorf("SetInt(%q, %d) = %v, want %q", tt.suffix, tt.value, err, tt.want)
		}
		if b.Position() != 0 {
			t.Errorf("SetInt(%q, %d) advanced position to %d after failure", tt.suffix, tt.value, b.Position())
		}
	}
}

func TestByteBufferValue_RangeDiagnostics(t *testing.T) {
	tests := []struct {
		name   string
		length int
		pos    int
		suffix string
		want   string
	}{
		{"byte on empty", 0, 0, "byte", "Out of range (index 0, size 1 for length 0)"},
		{"word past end", 3, 2, "word", "Out of range (index 2, size 2 for length 3)"},
		{"dword past end", 5, 4, "dword", "Out of range (index 4, size 4 for length 5)"},
		{"int64 past end", 9, 8, "int64", "Out of range (index 8, size 8 for length 9)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewByteBufferValue()
			b.SetLength(tt.length)
			if err := b.SetPosition(tt.pos); err != nil {
				t.Fatalf("SetPosition: %v", err)
			}
			err := b.SetInt(tt.suffix, 0)
			if err == nil || err.Error() != tt.want {
				t.Fatalf("SetInt = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestByteBufferValue_LittleEndianRoundTrip(t *testing.T) {
	tests := []struct {
		suffix   string
		value    int64
		wantJSON string
		signed   bool
	}{
		{"int8", -3, "[253]", true},
		{"int16", -3, "[253,255]", true},
		{"word", 34*256 + 12, "[12,34]", false},
		{"int32", -0x01020304, "[252,252,253,254]", true},
		{"dword", 0x01020304, "[4,3,2,1]", false},
		{"int64", 0x01020304050708, "[8,7,5,4,3,2,1,0]", true},
	}
	for _, tt := range tests {
		t.Run(tt.suffix, func(t *testing.T) {
			spec, ok := ByteBufferIntegerSpec(tt.suffix)
			if !ok {
				t.Fatalf("unknown suffix %q", tt.suffix)
			}
			b := NewByteBufferValue()
			b.SetLength(spec.Size)
			if err := b.SetInt(tt.suffix, tt.value); err != nil {
				t.Fatalf("SetInt: %v", err)
			}
			if got := b.ToJSON(); got != tt.wantJSON {
				t.Fatalf("ToJSON = %s, want %s", got, tt.wantJSON)
			}
			if b.Position() != spec.Size {
				t.Fatalf("position = %d, want %d", b.Position(), spec.Size)
			}
			got, err := b.GetIntAt(0, spec.Size, tt.signed)
			if err != nil {
				t.Fatalf("GetIntAt: %v", err)
			}
			if got != tt.value {
				t.Fatalf("round trip = %d, want %d", got, tt.value)
			}
		})
	}
}

func TestByteBufferValue_GetIntegers(t *testing.T) {
	b := NewByteBufferValue()
	b.SetLength(32)
	for i := 0; i < 32; i++ {
		if err := b.SetInt("byte", int64(128+i-16)); err != nil {
			t.Fatalf("SetInt: %v", err)
		}
	}

	// The index argument is bounds-checked only; elements always start at byte 0.
	unsigned, err := b.GetIntegers(1, 3, 2, false)
	if err != nil {
		t.Fatalf("GetIntegers: %v", err)
	}
	want := []int64{0x7170, 0x7372, 0x7574}
	for i := range want {
		if unsigned[i] != want[i] {
			t.Fatalf("unsigned[%d] = %#x, want %#x", i, unsigned[i], want[i])
		}
	}

	signed, err := b.GetIntegers(0, 32, 1, true)
	if err != nil {
		t.Fatalf("GetIntegers: %v", err)
	}
	if signed[0] != 0x70 || signed[16] != -128 || signed[31] != -113 {
		t.Fatalf("signed = %v, want 0x70 ... -128 ... -113", signed)
	}

	if _, err := b.GetIntegers(1, 16, 2, false); err == nil {
		t.Fatal("GetIntegers past the end should fail")
	}
}

func TestByteBufferValue_Floats(t *testing.T) {
	b := NewByteBufferValue()
	b.SetLength(8)
	if err := b.SetIntAt("dword", 0, 0x40490fd8); err != nil {
		t.Fatalf("SetIntAt: %v", err)
	}
	if err := b.SetFloatAt("single", 4, 3.141592); err != nil {
		t.Fatalf("SetFloatAt: %v", err)
	}
	first, err := b.GetFloatAt("single", 0)
	if err != nil {
		t.Fatalf("GetFloatAt: %v", err)
	}
	second, err := b.GetFloatAt("single", 4)
	if err != nil {
		t.Fatalf("GetFloatAt: %v", err)
	}
	if first != second {
		t.Fatalf("Single(3.141592) = %v, want the bit pattern 0x40490FD8 = %v", second, first)
	}

	if err := b.SetFloatAt("double", 0, 3.141592); err != nil {
		t.Fatalf("SetFloatAt: %v", err)
	}
	bits, err := b.GetIntAt(0, 8, true)
	if err != nil {
		t.Fatalf("GetIntAt: %v", err)
	}
	if uint64(bits) != math.Float64bits(3.141592) {
		t.Fatalf("Double bits = %#x, want %#x", uint64(bits), math.Float64bits(3.141592))
	}
}

func TestByteBufferValue_Extended(t *testing.T) {
	tests := []float64{0, 1, -1, 123456789.0, 3.141592653589793, -2.5e300, 1e-300}
	for _, want := range tests {
		b := NewByteBufferValue()
		b.SetLength(10)
		if err := b.SetFloatAt("extended", 0, want); err != nil {
			t.Fatalf("SetFloatAt(%v): %v", want, err)
		}
		got, err := b.GetFloatAt("extended", 0)
		if err != nil {
			t.Fatalf("GetFloatAt(%v): %v", want, err)
		}
		if got != want {
			t.Fatalf("extended round trip = %v, want %v", got, want)
		}
	}

	// A corrupted mantissa byte must shift the value by exactly its weight.
	b := NewByteBufferValue()
	b.SetLength(10)
	if err := b.SetFloatAt("extended", 0, 123456789.0); err != nil {
		t.Fatalf("SetFloatAt: %v", err)
	}
	if err := b.SetIntAt("byte", 3, 3); err != nil {
		t.Fatalf("SetIntAt: %v", err)
	}
	got, err := b.GetFloatAt("extended", 0)
	if err != nil {
		t.Fatalf("GetFloatAt: %v", err)
	}
	if math.Abs(got-123456789.000366) > 1e-6 {
		t.Fatalf("corrupted extended = %v, want ~123456789.000366", got)
	}
}

func TestByteBufferValue_DataStrings(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"ascii", "hello", "68656c6c6f"},
		{"empty", "", ""},
		{"digit", "0", "30"},
		{"low byte of utf16 units", "ሴ噸", "3478"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewByteBufferValue()
			b.AssignDataString(tt.in)
			if got := b.ToHexString(); got != tt.want {
				t.Fatalf("ToHexString = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestByteBufferValue_SetDataRange(t *testing.T) {
	b := NewByteBufferValue()
	b.SetLength(4)
	if err := b.SetData("ab"); err != nil {
		t.Fatalf("SetData: %v", err)
	}
	if err := b.SetData("cd"); err != nil {
		t.Fatalf("SetData: %v", err)
	}
	err := b.SetData("ef")
	if err == nil || err.Error() != "Out of range (index 4, size 2 for length 4)" {
		t.Fatalf("SetData past end = %v", err)
	}

	got, err := b.GetDataAt(0, 4)
	if err != nil || got != "abcd" {
		t.Fatalf("GetDataAt = %q, %v", got, err)
	}
	if _, err := b.GetDataAt(1, 4); err == nil {
		t.Fatal("GetDataAt past end should fail")
	}
	if err := b.SetDataAt(1, "xy"); err != nil {
		t.Fatalf("SetDataAt: %v", err)
	}
	if got, _ := b.GetDataAt(0, 3); got != "axy" {
		t.Fatalf("GetDataAt(0,3) = %q, want %q", got, "axy")
	}
}

func TestByteBufferValue_Encodings(t *testing.T) {
	b := NewByteBufferValue()
	b.AssignDataString("hello world")
	if got := b.ToBase64(); got != "aGVsbG8gd29ybGQ=" {
		t.Fatalf("ToBase64 = %q", got)
	}
	b.SetLength(0)
	if got := b.ToBase64(); got != "" {
		t.Fatalf("empty ToBase64 = %q", got)
	}
	b.SetLength(2)
	if got := b.ToBase64(); got != "AAA=" {
		t.Fatalf("zeroed ToBase64 = %q", got)
	}

	if err := b.AssignBase64("dGVzdGluZw=="); err != nil {
		t.Fatalf("AssignBase64: %v", err)
	}
	if got := b.ToDataString(); got != "testing" {
		t.Fatalf("AssignBase64 -> %q", got)
	}
	if err := b.AssignJSON("[48,49,50]"); err != nil {
		t.Fatalf("AssignJSON: %v", err)
	}
	if got := b.ToDataString(); got != "012" {
		t.Fatalf("AssignJSON -> %q", got)
	}
	if err := b.AssignHexString("39383736"); err != nil {
		t.Fatalf("AssignHexString: %v", err)
	}
	if got := b.ToDataString(); got != "9876" {
		t.Fatalf("AssignHexString -> %q", got)
	}
}

func TestByteBufferValue_AssignIsACopy(t *testing.T) {
	source := NewByteBufferValue()
	source.AssignDataString("hello")
	target := NewByteBufferValue()
	target.Assign(source)
	source.SetLength(2)

	if got := source.ToDataString(); got != "he" {
		t.Fatalf("source = %q, want %q", got, "he")
	}
	if got := target.ToDataString(); got != "hello" {
		t.Fatalf("Assign must copy, target = %q, want %q", got, "hello")
	}
}

func TestByteBufferValue_Copy(t *testing.T) {
	b := NewByteBufferValue()
	b.AssignDataString("hello world")

	tests := []struct {
		index int
		count int
		want  string
	}{
		{0, 11, "hello world"},
		{6, 11, "world"},
		{6, 10, "world"},
		{1, 2, "el"},
		{2, 0, ""},
	}
	for _, tt := range tests {
		if got := b.Copy(tt.index, tt.count).ToDataString(); got != tt.want {
			t.Errorf("Copy(%d, %d) = %q, want %q", tt.index, tt.count, got, tt.want)
		}
	}
}
