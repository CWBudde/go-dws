package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func (e *Evaluator) initializeObjectFields(classInfo runtime.IClassInfo, obj *runtime.ObjectInstance, node ast.Node, ctx *ExecutionContext) Value {
	if classInfo == nil || obj == nil {
		return nil
	}

	for _, meta := range classMetadataHierarchy(classInfo.GetMetadata()) {
		for fieldName, fieldMeta := range meta.Fields {
			if fieldMeta == nil {
				continue
			}

			var fieldValue Value
			if fieldMeta.InitValue != nil {
				fieldValue = e.Eval(fieldMeta.InitValue, ctx)
				if isError(fieldValue) {
					return e.newError(node, "failed to initialize field '%s': %v", fieldName, fieldValue)
				}
			} else if fieldMeta.Type != nil {
				fieldValue = e.getZeroValueForType(fieldMeta.Type)
			} else {
				fieldValue = &runtime.NilValue{}
			}

			obj.SetField(fieldName, fieldValue)
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
