package semantic

import (
	"github.com/cwbudde/go-dws/internal/encoding"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// registerBuiltinEncoderTypes registers DWScript's EncodingLib encoder classes
// (Encoder and its Base64Encoder, HexadecimalEncoder, ... subclasses) so that
// `Base64Encoder.Encode(s)`, `var e := Base64Encoder` and `class of Encoder`
// parameters type-check.
//
// Every encoder method is a virtual class function taking one String and
// returning a String; the implementations live in internal/interp/encoders.go
// and both sides are generated from encoding.EncoderClassSpecs.
func (a *Analyzer) registerBuiltinEncoderTypes() {
	objectClass := a.getClassType("TObject")

	registered := make(map[string]*types.ClassType)
	for _, spec := range encoding.EncoderClassSpecs() {
		parent := objectClass
		if spec.Parent != "" {
			if declared, ok := registered[ident.Normalize(spec.Parent)]; ok {
				parent = declared
			}
		}

		classType := types.NewClassType(spec.Name, parent)
		isRoot := spec.Parent == ""

		for _, method := range spec.Methods {
			normalized := ident.Normalize(method.Name)
			// A subclass method that the root also declares overrides it;
			// anything else (e.g. Base64Encoder.EncodeMIME) is introduced here.
			isOverride := !isRoot && parent != nil && parent.HasMethod(method.Name)

			classType.AddMethodOverload(method.Name, &types.MethodInfo{
				Signature: &types.FunctionType{
					Parameters: []types.Type{types.STRING},
					ReturnType: types.STRING,
				},
				IsClassMethod: true,
				IsVirtual:     isRoot,
				IsOverride:    isOverride,
				Visibility:    int(ast.VisibilityPublic),
			})
			classType.MethodVisibility[normalized] = int(ast.VisibilityPublic)
			classType.ClassMethodFlags[normalized] = true
			if isRoot {
				classType.VirtualMethods[normalized] = true
			}
			if isOverride {
				classType.OverrideMethods[normalized] = true
			}
		}

		registered[ident.Normalize(spec.Name)] = classType
		a.registerBuiltinType(spec.Name, classType)
	}
}
