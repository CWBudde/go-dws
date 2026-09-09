package interp

import "github.com/cwbudde/go-dws/internal/interp/runtime"

// ValueKind identifies the runtime representation.
func (*RTTITypeInfoValue) ValueKind() runtime.ValueKind { return runtime.KindRTTITypeInfo }

// ValueKind identifies the runtime representation.
func (*ReferenceValue) ValueKind() runtime.ValueKind { return runtime.KindReference }

// ValueKind identifies the runtime representation.
func (*TypeCastValue) ValueKind() runtime.ValueKind { return runtime.KindTypeCast }

// ValueKind identifies the runtime representation.
func (*ExternalFunctionValue) ValueKind() runtime.ValueKind { return runtime.KindExternalFunction }

// ValueKind identifies the runtime representation.
func (*ErrorValue) ValueKind() runtime.ValueKind { return runtime.KindError }

// ValueKind identifies the runtime representation.
func (*ContractFailureError) ValueKind() runtime.ValueKind { return runtime.KindError }

// ValueKind identifies the runtime representation.
func (*RuntimeError) ValueKind() runtime.ValueKind { return runtime.KindError }
