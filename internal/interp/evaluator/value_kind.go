package evaluator

import "github.com/cwbudde/go-dws/internal/interp/runtime"

// ValueKind identifies the runtime representation.
func (*interp_TypeCastValue) ValueKind() runtime.ValueKind { return runtime.KindTypeCast }

// ValueKind identifies the runtime representation.
func (*LocalFunctionSet) ValueKind() runtime.ValueKind { return runtime.KindLocalFunctionSet }
