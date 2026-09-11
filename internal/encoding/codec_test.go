package encoding

import "testing"

func TestBase64Encode_Fixtures(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"tab separated", "alpha\tomega", "YWxwaGEJb21lZ2E="},
		{"one byte padding", "a", "YQ=="},
		{"two byte padding", "ab", "YWI="},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Base64Encode(tt.input); got != tt.want {
				t.Errorf("Base64Encode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBase64Decode_IgnoresWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"plain", "YWxwaGEJb21lZ2E=", "alpha\tomega"},
		{"embedded cr lf", "aGVs\rbG8=\n", "hello"},
		{"embedded spaces and tabs", "aGVs  \r\t bG8=\n  ", "hello"},
		{"missing padding", "aGVsbG8", "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Base64Decode(tt.input)
			if err != nil {
				t.Fatalf("Base64Decode(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("Base64Decode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBase64EncodeMIME_WrapsAt76(t *testing.T) {
	if got := Base64EncodeMIME(""); got != "" {
		t.Errorf("Base64EncodeMIME(%q) = %q, want empty", "", got)
	}

	short := repeat("hello", 5)
	if got := Base64EncodeMIME(short); got != "aGVsbG9oZWxsb2hlbGxvaGVsbG9oZWxsbw==" {
		t.Errorf("Base64EncodeMIME(short) = %q", got)
	}

	long := repeat("hello", 50)
	encoded := Base64EncodeMIME(long)
	lines := splitLines(encoded)
	if len(lines) != 5 {
		t.Fatalf("expected 5 MIME lines, got %d (%q)", len(lines), encoded)
	}
	for i, line := range lines[:len(lines)-1] {
		if len(line) != base64MIMELineWidth {
			t.Errorf("line %d has length %d, want %d", i, len(line), base64MIMELineWidth)
		}
	}
	roundTrip, err := Base64Decode(encoded)
	if err != nil {
		t.Fatalf("round trip failed: %v", err)
	}
	if roundTrip != long {
		t.Errorf("MIME round trip mismatch")
	}
}

func TestBase64URI_RoundTrip(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"!>>", "IT4-"},
		{"!>", "IT4"},
		{"a", "YQ"},
		{"1", "MQ"},
		{"123456789", "MTIzNDU2Nzg5"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Base64URIEncode(tt.input)
			if got != tt.want {
				t.Errorf("Base64URIEncode(%q) = %q, want %q", tt.input, got, tt.want)
			}
			back, err := Base64URIDecode(got)
			if err != nil {
				t.Fatalf("Base64URIDecode(%q) returned error: %v", got, err)
			}
			if back != tt.input {
				t.Errorf("Base64URIDecode(%q) = %q, want %q", got, back, tt.input)
			}
		})
	}
}

func TestBase32_RoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		decoded string
		encoded string
	}{
		{"empty", "", ""},
		{"hello world", "Hello World", "JBSWY3DPEBLW64TMMQ"},
		{"single zero byte", "\x00", "AA"},
		{"seven zero bytes", "\x00\x00\x00\x00\x00\x00\x00", "AAAAAAAAAAAA"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Base32Encode(tt.decoded); got != tt.encoded {
				t.Errorf("Base32Encode(%q) = %q, want %q", tt.decoded, got, tt.encoded)
			}
			got, err := Base32Decode(tt.encoded)
			if err != nil {
				t.Fatalf("Base32Decode(%q) returned error: %v", tt.encoded, err)
			}
			if got != tt.decoded {
				t.Errorf("Base32Decode(%q) = %q, want %q", tt.encoded, got, tt.decoded)
			}
		})
	}
}

func TestBase32Decode_ZeroAliasAndErrors(t *testing.T) {
	got, err := Base32Decode("000000")
	if err != nil {
		t.Fatalf("Base32Decode returned error: %v", err)
	}
	if hex := HexadecimalEncode(got); hex != "739ce7" {
		t.Errorf("Base32Decode(\"000000\") hex = %q, want %q", hex, "739ce7")
	}

	if _, err := Base32Decode("...."); err == nil {
		t.Fatal("expected an error for invalid characters")
	} else if err.Error() != "Invalid character (#46) in Base32" {
		t.Errorf("error = %q", err.Error())
	}
}

func TestBase58_RoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		decoded string
		encoded string
	}{
		{"empty", "", ""},
		{"hello world", "Hello World", "JxF12TrwUP45BMd"},
		{"single zero byte", "\x00", "1"},
		{"seven zero bytes", "\x00\x00\x00\x00\x00\x00\x00", "1111111"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Base58Encode(tt.decoded); got != tt.encoded {
				t.Errorf("Base58Encode(%q) = %q, want %q", tt.decoded, got, tt.encoded)
			}
			got, err := Base58Decode(tt.encoded)
			if err != nil {
				t.Fatalf("Base58Decode(%q) returned error: %v", tt.encoded, err)
			}
			if got != tt.decoded {
				t.Errorf("Base58Decode(%q) = %q, want %q", tt.encoded, got, tt.decoded)
			}
		})
	}
}

func TestBase58_BitcoinAddressVector(t *testing.T) {
	const encoded = "16UwLL9Risc3QfPqBUvKofHmBQ7wMtjvM"
	const hex = "00010966776006953d5567439e5e39f86a0d273beed61967f6"

	decoded, err := Base58Decode(encoded)
	if err != nil {
		t.Fatalf("Base58Decode returned error: %v", err)
	}
	if got := HexadecimalEncode(decoded); got != hex {
		t.Errorf("hex = %q, want %q", got, hex)
	}

	raw, err := HexadecimalDecode(hex)
	if err != nil {
		t.Fatalf("HexadecimalDecode returned error: %v", err)
	}
	if got := Base58Encode(raw); got != encoded {
		t.Errorf("Base58Encode = %q, want %q", got, encoded)
	}
}

func TestBase58Decode_InvalidCharacter(t *testing.T) {
	_, err := Base58Decode("....")
	if err == nil {
		t.Fatal("expected an error")
	}
	if err.Error() != "Non-base58 character" {
		t.Errorf("error = %q", err.Error())
	}
}

func TestHexadecimal_RoundTripAllBytes(t *testing.T) {
	for i := 0; i < 256; i++ {
		s := string(rune(i))
		encoded := HexadecimalEncode(s)
		decoded, err := HexadecimalDecode(encoded)
		if err != nil {
			t.Fatalf("byte %d: %v", i, err)
		}
		if decoded != s {
			t.Fatalf("byte %d did not round trip", i)
		}
	}
	if got := HexadecimalEncode("alpha\tomega"); got != "616c706861096f6d656761" {
		t.Errorf("HexadecimalEncode = %q", got)
	}
}

func TestHexadecimalDecode_Errors(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr string
	}{
		{"", "", ""},
		{"1", "", "Even hexadecimal character count expected"},
		{"z1", "", "Invalid hexadecimal character at index 1"},
		{"1z", "", "Invalid hexadecimal character at index 2"},
		{"Aaz1", "", "Invalid hexadecimal character at index 3"},
		{"Aa1z", "", "Invalid hexadecimal character at index 4"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := HexadecimalDecode(tt.input)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tt.want {
					t.Errorf("got %q, want %q", got, tt.want)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestUTF8_RoundTrip(t *testing.T) {
	const script = "éric"
	encoded := UTF8Encode(script)
	if encoded != "Ã©ric" {
		t.Errorf("UTF8Encode = %q", encoded)
	}
	if got := UTF8Decode(encoded); got != script {
		t.Errorf("UTF8Decode = %q, want %q", got, script)
	}
}

func TestUTF16_RoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		hex       string
		bigEndian bool
	}{
		{"ascii big endian", "Example", "004500780061006d0070006c0065", true},
		{"ascii little endian", "Example", "4500780061006d0070006c006500", false},
		{"latin1 big endian", "éric", "00e9007200690063", true},
		{"latin1 little endian", "éric", "e900720069006300", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := UTF16Encode(tt.input, tt.bigEndian)
			if got := HexadecimalEncode(encoded); got != tt.hex {
				t.Errorf("hex = %q, want %q", got, tt.hex)
			}
			if got := UTF16Decode(encoded, tt.bigEndian); got != tt.input {
				t.Errorf("UTF16Decode = %q, want %q", got, tt.input)
			}
		})
	}
}

// repeat concatenates s count times.
func repeat(s string, count int) string {
	out := ""
	for i := 0; i < count; i++ {
		out += s
	}
	return out
}

// splitLines splits on CRLF, the separator used by MIME output.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i+1 < len(s); i++ {
		if s[i] == '\r' && s[i+1] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 2
			i++
		}
	}
	return append(lines, s[start:])
}
