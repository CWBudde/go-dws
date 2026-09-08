package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

type HelperInfo = runtime.MutableHelperInfo

var NewHelperInfo = runtime.NewMutableHelperInfo

// registerBuiltinHelper adapts the shared semantic catalog to runtime metadata.
func (i *Interpreter) registerBuiltinHelper(target string) {
	definition := types.NewBuiltinHelper(target)
	helper := NewHelperInfo(definition.Name, definition.TargetType, definition.IsRecordHelper)
	helper.Properties = definition.Properties
	helper.BuiltinMethods = definition.BuiltinMethods
	for name := range definition.BuiltinMethods {
		helper.Methods[name] = nil
	}
	i.typeSystem.RegisterHelper(target, helper)
}

func (i *Interpreter) initArrayHelpers() { i.registerBuiltinHelper("array") }

func (i *Interpreter) initIntrinsicHelpers() {
	for _, target := range []string{"Integer", "Float", "Boolean", "String", "array of String"} {
		i.registerBuiltinHelper(target)
	}
}

func (i *Interpreter) initEnumHelpers() { i.registerBuiltinHelper("enum") }
