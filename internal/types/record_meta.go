package types

// RecordMetaType is the type of a record declaration used as a value. Its
// identity is the canonical record declaration, not the shape of its fields.
// SourceType preserves an alias for helper lookup without changing that identity.
type RecordMetaType struct {
	RecordType *RecordType
	SourceType Type
}

// NewRecordMetaType creates the metatype of a record or an alias of that record.
func NewRecordMetaType(source Type) *RecordMetaType {
	record, ok := GetUnderlyingType(source).(*RecordType)
	if !ok || record == nil {
		return nil
	}
	return &RecordMetaType{RecordType: record, SourceType: source}
}

func (t *RecordMetaType) String() string { return "meta of " + t.RecordType.String() }

// TypeKind identifies record metatypes independently of record instance types.
func (t *RecordMetaType) TypeKind() string { return "RECORD_META" }

// Equals compares metatypes by their canonical record declaration identity.
func (t *RecordMetaType) Equals(other Type) bool {
	meta, ok := GetUnderlyingType(other).(*RecordMetaType)
	return ok && t.RecordType != nil && t.RecordType == meta.RecordType
}
