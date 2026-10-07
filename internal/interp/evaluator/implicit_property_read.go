package evaluator

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// evalImplicitPropertyRead executes the selected descriptor directly. Generic
// member lookup would rediscover a descendant property and would suppress valid
// nested reads while another property accessor is active.
func (e *Evaluator) evalImplicitPropertyRead(binding *ast.ImplicitPropertyReadBinding, ctx *ExecutionContext) Value {
	read := binding.Read
	owner := e.typeSystem.LookupClass(binding.Owner)
	if owner == nil {
		return e.newError(read, "property owner '%s' not found", binding.Owner)
	}
	descriptor := owner.LookupProperty(read.Member.Value)
	if descriptor == nil {
		return e.newError(read, "property '%s' not found in '%s'", read.Member.Value, binding.Owner)
	}
	prop, ok := unwrapPropertyInfo(descriptor.Impl)
	if !ok {
		return e.newError(read, "invalid property info type")
	}
	receiver := e.Eval(read.Object, ctx)
	if isError(receiver) {
		return receiver
	}
	if prop.IsClassProperty {
		class := e.helperReceiverClassInfo(receiver)
		if class == nil {
			return e.newError(read, "class property requires a class receiver")
		}
		// Storage belongs to the declaring class. Getter methods retain the dynamic
		// metaclass so virtual dispatch still selects descendant overrides.
		if prop.ReadKind == types.PropAccessField {
			if storage, _ := owner.LookupClassVar(prop.ReadSpec); storage != nil {
				return e.evalClassPropertyRead(owner, prop, read, ctx)
			}
		}
		return e.evalClassPropertyRead(class, prop, read, ctx)
	}
	object, ok := receiver.(ObjectValue)
	if !ok {
		return e.newError(read, "property requires an object receiver")
	}
	if prop.ReadKind == types.PropAccessField {
		if value := getFieldWithStaticClass(object, prop.ReadSpec, binding.Owner); value != nil {
			return value
		}
	}
	return e.executePropertyRead(receiver, prop, read, ctx)
}
