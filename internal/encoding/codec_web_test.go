package encoding

import "testing"

func TestURLEncodedEncode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"unreserved", "aZ09-._~", "aZ09-._~"},
		{"mixed", "url encoded/+\x01that é!", "url%20encoded%2F%2B%01that%20%C3%A9%21"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := URLEncodedEncode(tt.input); got != tt.want {
				t.Errorf("URLEncodedEncode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestURLEncodedDecode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"round trip", "url%20encoded%2F%2B%01that%20%C3%A9%21", "url encoded/+\x01that é!"},
		{"lowercase escapes", "%c3%a9%5d", "é]"},
		{"truncated escape", "%z", ""},
		{"invalid escape", "%zz", "�"},
		{"half invalid escape", "%1z", "�"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := URLEncodedDecode(tt.input); got != tt.want {
				t.Errorf("URLEncodedDecode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHTMLTextEncode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"plain", "hello", "hello"},
		{"ampersand", "&hello", "&amp;hello"},
		{"greater than", "hello>", "hello&gt;"},
		{"apostrophe and nbsp", "' ", "&#39;&nbsp;"},
		{"all specials", `<hello"world>here&`, "&lt;hello&quot;world&gt;here&amp;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HTMLTextEncode(tt.input); got != tt.want {
				t.Errorf("HTMLTextEncode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHTMLTextDecode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"round trip", "&lt;hello&quot;world&gt;here&amp;", `<hello"world>here&`},
		{"named and numeric", "&amp;&lt;&#43;&gt;&apos;&quot;", `&<+>'"`},
		{"hash variants", "&num;&#x00023;&#35;", "###"},
		{"unknown entities untouched", "&&&bug;&;", "&&&bug;&;"},
		{"tags stripped", `<tag a="'" b='"'>azerty</tag>`, "azerty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HTMLTextDecode(tt.input); got != tt.want {
				t.Errorf("HTMLTextDecode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHTMLAttribute_RoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		decoded string
		encoded string
	}{
		{"empty", "", ""},
		{"alphanumeric", "hello", "hello"},
		{"punctuation", "&/?", "&#38;&#47;&#63;"},
		{"percent", "%%aaaaaaaaa", "&#37;&#37;aaaaaaaaa"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HTMLAttributeEncode(tt.decoded)
			if got != tt.encoded {
				t.Errorf("HTMLAttributeEncode(%q) = %q, want %q", tt.decoded, got, tt.encoded)
			}
			if back := HTMLAttributeDecode(got); back != tt.decoded {
				t.Errorf("HTMLAttributeDecode(%q) = %q, want %q", got, back, tt.decoded)
			}
		})
	}
}
