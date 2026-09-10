package encoding

import "fmt"

// This file is the single source of truth for DWScript's EncodingLib encoder
// classes. Both the semantic analyzer (which needs the class and method names)
// and the interpreter (which needs the implementations) build their
// registrations from EncoderClassSpecs, so the two surfaces cannot drift.

// EncoderMethodFunc is the shape every encoder method has: one String
// parameter, a String result, and a DWScript-worded error on invalid input.
type EncoderMethodFunc func(s string) (string, error)

// EncoderMethod describes one class method of an encoder class.
type EncoderMethod struct {
	// Fn implements the method.
	Fn EncoderMethodFunc
	// Name is the method name as written in scripts, e.g. "Encode".
	Name string
}

// EncoderClassSpec describes one encoder class of DWScript's EncodingLib.
type EncoderClassSpec struct {
	// Name is the class name as written in scripts, e.g. "Base64Encoder".
	Name string
	// Parent is the name of the base class, empty for the root "Encoder".
	Parent string
	// Methods are the class methods the class declares or overrides.
	Methods []EncoderMethod
	// IsAbstract marks the root class, whose methods must not be called.
	IsAbstract bool
}

// EncoderBaseClassName is the root of the encoder hierarchy; scripts use it as
// `class of Encoder` to pass encoders around.
const EncoderBaseClassName = "Encoder"

// EncoderClassSpecs returns the encoder classes of DWScript's EncodingLib in
// registration order: the abstract root first, then its subclasses.
//
// Every method takes a single String and returns a String. Encoders that work
// on bytes consume and produce byte strings (see the codec.go file comment).
func EncoderClassSpecs() []EncoderClassSpec {
	return []EncoderClassSpec{
		{
			Name:       EncoderBaseClassName,
			IsAbstract: true,
			Methods: []EncoderMethod{
				{Name: "Encode", Fn: abstractEncoderMethod("Encode")},
				{Name: "Decode", Fn: abstractEncoderMethod("Decode")},
			},
		},
		{
			Name:   "Base64Encoder",
			Parent: EncoderBaseClassName,
			Methods: []EncoderMethod{
				{Name: "Encode", Fn: infallible(Base64Encode)},
				{Name: "Decode", Fn: Base64Decode},
				{Name: "EncodeMIME", Fn: infallible(Base64EncodeMIME)},
			},
		},
		{
			Name:   "Base64URIEncoder",
			Parent: EncoderBaseClassName,
			Methods: []EncoderMethod{
				{Name: "Encode", Fn: infallible(Base64URIEncode)},
				{Name: "Decode", Fn: Base64URIDecode},
			},
		},
		{
			Name:   "Base32Encoder",
			Parent: EncoderBaseClassName,
			Methods: []EncoderMethod{
				{Name: "Encode", Fn: infallible(Base32Encode)},
				{Name: "Decode", Fn: Base32Decode},
			},
		},
		{
			Name:   "Base58Encoder",
			Parent: EncoderBaseClassName,
			Methods: []EncoderMethod{
				{Name: "Encode", Fn: infallible(Base58Encode)},
				{Name: "Decode", Fn: Base58Decode},
			},
		},
		{
			Name:   "HexadecimalEncoder",
			Parent: EncoderBaseClassName,
			Methods: []EncoderMethod{
				{Name: "Encode", Fn: infallible(HexadecimalEncode)},
				{Name: "Decode", Fn: HexadecimalDecode},
			},
		},
		{
			Name:   "UTF8Encoder",
			Parent: EncoderBaseClassName,
			Methods: []EncoderMethod{
				{Name: "Encode", Fn: infallible(UTF8Encode)},
				{Name: "Decode", Fn: infallible(UTF8Decode)},
			},
		},
		{
			Name:   "UTF16BigEndianEncoder",
			Parent: EncoderBaseClassName,
			Methods: []EncoderMethod{
				{Name: "Encode", Fn: infallible(func(s string) string { return UTF16Encode(s, true) })},
				{Name: "Decode", Fn: infallible(func(s string) string { return UTF16Decode(s, true) })},
			},
		},
		{
			Name:   "UTF16LittleEndianEncoder",
			Parent: EncoderBaseClassName,
			Methods: []EncoderMethod{
				{Name: "Encode", Fn: infallible(func(s string) string { return UTF16Encode(s, false) })},
				{Name: "Decode", Fn: infallible(func(s string) string { return UTF16Decode(s, false) })},
			},
		},
		{
			Name:   "URLEncodedEncoder",
			Parent: EncoderBaseClassName,
			Methods: []EncoderMethod{
				{Name: "Encode", Fn: infallible(URLEncodedEncode)},
				{Name: "Decode", Fn: infallible(URLEncodedDecode)},
			},
		},
		{
			Name:   "HTMLTextEncoder",
			Parent: EncoderBaseClassName,
			Methods: []EncoderMethod{
				{Name: "Encode", Fn: infallible(HTMLTextEncode)},
				{Name: "Decode", Fn: infallible(HTMLTextDecode)},
			},
		},
		{
			Name:   "HTMLAttributeEncoder",
			Parent: EncoderBaseClassName,
			Methods: []EncoderMethod{
				{Name: "Encode", Fn: infallible(HTMLAttributeEncode)},
				{Name: "Decode", Fn: infallible(HTMLAttributeDecode)},
			},
		},
	}
}

// infallible adapts a codec that cannot fail to EncoderMethodFunc.
func infallible(fn func(string) string) EncoderMethodFunc {
	return func(s string) (string, error) {
		return fn(s), nil
	}
}

// abstractEncoderMethod builds the body of an abstract root method, which
// raises rather than encoding anything.
func abstractEncoderMethod(name string) EncoderMethodFunc {
	return func(string) (string, error) {
		return "", fmt.Errorf("Abstract class method \"%s\" called in class %q", name, EncoderBaseClassName)
	}
}
