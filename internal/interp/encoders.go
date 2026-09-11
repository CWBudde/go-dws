package interp

import (
	"github.com/cwbudde/go-dws/internal/encoding"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// registerBuiltinEncoders registers the runtime side of DWScript's EncodingLib
// encoder classes: the abstract Encoder root plus Base64Encoder,
// Base32Encoder, HexadecimalEncoder and friends.
//
// The classes carry no fields and no constructors; they exist purely as
// namespaces for virtual class methods whose bodies are Go functions
// (runtime.MethodMetadata.Native). Because class-method dispatch walks the
// hierarchy from the receiver's runtime class, an override in a subclass wins
// automatically, which is what makes `class of Encoder` parameters dispatch
// correctly.
//
// The class and method names come from encoding.EncoderClassSpecs, the same
// table the analyzer registers from (see internal/semantic/encoders.go).
func (i *Interpreter) registerBuiltinEncoders() {
	for _, spec := range encoding.EncoderClassSpecs() {
		classInfo := NewClassInfo(spec.Name)
		classInfo.IsAbstractFlag = spec.IsAbstract
		classInfo.IsExternalFlag = false

		parentName := spec.Parent
		if parentName == "" {
			parentName = "TObject"
		}
		if parent, ok := i.typeSystem.LookupClass(parentName).(*runtime.ClassInfo); ok && parent != nil {
			classInfo.Parent = parent
			classInfo.Type.Parent = parent.Type
			classInfo.Metadata.Parent = parent.Metadata
		}

		for _, method := range spec.Methods {
			classInfo.ClassMethods[ident.Normalize(method.Name)] = newEncoderClassMethod(method)
		}

		i.typeSystem.RegisterClassWithParent(spec.Name, classInfo, parentName)
	}
}

// newEncoderClassMethod builds the runtime metadata for one encoder class
// method: a class function taking a single String and returning a String,
// implemented by the spec's Go function.
func newEncoderClassMethod(method encoding.EncoderMethod) *runtime.MethodMetadata {
	fn := method.Fn
	return &runtime.MethodMetadata{
		Name:          method.Name,
		ReturnType:    types.STRING,
		IsClassMethod: true,
		IsVirtual:     true,
		Visibility:    runtime.VisibilityPublic,
		Parameters: []runtime.ParameterMetadata{
			{Name: "s", Type: types.STRING},
		},
		Native: func(_ runtime.IClassInfo, args []runtime.Value) (runtime.Value, error) {
			result, err := fn(encoderStringArgument(args))
			if err != nil {
				return nil, err
			}
			return &runtime.StringValue{Value: result}, nil
		},
	}
}

// encoderStringArgument reads the single String argument an encoder method
// takes. The evaluator has already checked the arity, and any other value kind
// (a Variant holding a string, for instance) is taken by its string form.
func encoderStringArgument(args []runtime.Value) string {
	if len(args) == 0 || args[0] == nil {
		return ""
	}
	if str, ok := args[0].(*runtime.StringValue); ok {
		return str.Value
	}
	return args[0].String()
}
