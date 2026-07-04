package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func (e *Evaluator) initializeObjectFields(classInfo runtime.IClassInfo, obj *runtime.ObjectInstance, node ast.Node, ctx *ExecutionContext) Value {
	if classInfo == nil || obj == nil {
		return nil
	}

	// Field initializers may reference class constants (e.g. FField := Value
	// where Value is a class const), so evaluate them in a scope with the
	// hierarchy's constants bound.
	ctx.PushEnv()
	defer ctx.PopEnv()
	e.bindClassConstantsForMethod(classInfo, ctx)

	for _, meta := range classMetadataHierarchy(classInfo.GetMetadata()) {
		for fieldName, fieldMeta := range meta.Fields {
			if fieldMeta == nil {
				continue
			}

			var fieldValue Value
			if fieldMeta.InitValue != nil {
				prevRecordTypeName := ctx.RecordTypeContext()
				prevRecordType := ctx.RecordTypeContextType()
				if recordType, ok := types.GetUnderlyingType(fieldMeta.Type).(*types.RecordType); ok {
					if recordType.Name != "" {
						ctx.SetRecordTypeContext(recordType.Name)
					} else {
						ctx.SetRecordTypeContextType(recordType)
					}
				}
				fieldValue = e.Eval(fieldMeta.InitValue, ctx)
				if prevRecordType != nil {
					ctx.SetRecordTypeContextType(prevRecordType)
				} else {
					ctx.SetRecordTypeContext(prevRecordTypeName)
				}
				if isError(fieldValue) {
					return e.newError(node, "failed to initialize field '%s': %v", fieldName, fieldValue)
				}
			} else if fieldMeta.Type != nil {
				fieldValue = e.getZeroValueForType(fieldMeta.Type)
			} else {
				fieldValue = &runtime.NilValue{}
			}

			// Initialize the slot owned by this declaring class so shadowed
			// fields (subclass redeclaring a parent field name) each get
			// their own storage.
			obj.SetFieldFromClass(fieldName, fieldValue, meta.Name)
		}
	}

	return nil
}

func classMetadataHierarchy(meta *runtime.ClassMetadata) []*runtime.ClassMetadata {
	if meta == nil {
		return nil
	}

	var chain []*runtime.ClassMetadata
	for current := meta; current != nil; current = current.Parent {
		chain = append(chain, current)
	}

	for left, right := 0, len(chain)-1; left < right; left, right = left+1, right-1 {
		chain[left], chain[right] = chain[right], chain[left]
	}

	return chain
}
