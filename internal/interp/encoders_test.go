package interp

import (
	"strings"
	"testing"
)

// TestEncoderClasses_ClassMethods exercises the built-in EncodingLib encoder
// classes through the script surface, which is what makes the native class
// method hook (runtime.MethodMetadata.Native) reachable.
func TestEncoderClasses_ClassMethods(t *testing.T) {
	tests := []struct {
		name   string
		script string
		want   string
	}{
		{
			name:   "base64 round trip",
			script: `PrintLn(Base64Encoder.Encode('alpha'#9'omega'));`,
			want:   "YWxwaGEJb21lZ2E=",
		},
		{
			name:   "base64 decode ignores whitespace",
			script: `PrintLn(Base64Encoder.Decode('aGVs'#13'bG8='#10));`,
			want:   "hello",
		},
		{
			name:   "base64 uri is unpadded",
			script: `PrintLn(Base64URIEncoder.Encode('!>>'));`,
			want:   "IT4-",
		},
		{
			name:   "hexadecimal is lowercase",
			script: `PrintLn(HexadecimalEncoder.Encode('alpha'#9'omega'));`,
			want:   "616c706861096f6d656761",
		},
		{
			name:   "base32 has no padding",
			script: `PrintLn(Base32Encoder.Encode('Hello World'));`,
			want:   "JBSWY3DPEBLW64TMMQ",
		},
		{
			name:   "base58 keeps leading zeros",
			script: `PrintLn(Base58Encoder.Encode(#0#0#0#0#0#0#0));`,
			want:   "1111111",
		},
		{
			name:   "html text encodes specials",
			script: `PrintLn(HTMLTextEncoder.Encode('<hello"world>here&'));`,
			want:   "&lt;hello&quot;world&gt;here&amp;",
		},
		{
			name:   "html attribute encodes punctuation",
			script: `PrintLn(HTMLAttributeEncoder.Encode('&/?'));`,
			want:   "&#38;&#47;&#63;",
		},
		{
			name:   "url encoding escapes reserved characters",
			script: `PrintLn(URLEncodedEncoder.Encode('a b/c'));`,
			want:   "a%20b%2Fc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runScriptTestWithSemantic(t, tt.script, tt.want)
		})
	}
}

// TestEncoderClasses_AsValueAndMetaclass covers the two reasons the encoders
// are classes rather than a namespace: they can be stored in a variable, and
// they can be passed as `class of Encoder` with virtual dispatch.
func TestEncoderClasses_AsValueAndMetaclass(t *testing.T) {
	runScriptTestWithSemantic(t, `
		var encoder := Base64Encoder;
		PrintLn(encoder.Decode(encoder.Encode('hello')));
	`, "hello")

	runScriptTestWithSemantic(t, `
		procedure Show(e : class of Encoder; s : String);
		begin
			PrintLn(HexadecimalEncoder.Encode(e.Encode(s)));
		end;

		Show(UTF16BigEndianEncoder, 'Example');
		Show(UTF16LittleEndianEncoder, 'Example');
	`, "004500780061006d0070006c0065\n4500780061006d0070006c006500")
}

// TestEncoderClasses_MalformedInputRaises checks that a native method's error
// surfaces as a catchable DWScript exception carrying DWScript's wording, the
// enclosing routine name and the failing statement's position.
func TestEncoderClasses_MalformedInputRaises(t *testing.T) {
	tests := []struct {
		name   string
		script string
		want   string
	}{
		{
			name: "base32 invalid character",
			script: `
try
   PrintLn(Base32Encoder.Decode('....'));
except
   on E : Exception do PrintLn(E.Message);
end;`,
			want: "Invalid character (#46) in Base32 [line: 3, column: 4]",
		},
		{
			name: "base58 invalid character",
			script: `
try
   PrintLn(Base58Encoder.Decode('....'));
except
   on E : Exception do PrintLn(E.Message);
end;`,
			want: "Non-base58 character [line: 3, column: 4]",
		},
		{
			name: "hexadecimal odd length names the routine",
			script: `
procedure TryDecode(s : String);
begin
   try
      PrintLn(HexadecimalEncoder.Decode(s));
   except
      on E : Exception do PrintLn(E.Message);
   end;
end;
TryDecode('1');`,
			want: "Even hexadecimal character count expected in TryDecode [line: 5, column: 7]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutputAndSemantic(t, tt.script)
			if got := strings.TrimSpace(output); got != tt.want {
				t.Errorf("Output mismatch:\nExpected: %s\nGot:      %s", tt.want, got)
			}
		})
	}
}
