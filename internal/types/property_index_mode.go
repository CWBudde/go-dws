package types

// PropertyIndexMode is the declared passing mode of a property index.
type PropertyIndexMode uint8

const (
	// PropertyIndexValue passes a copy and is the zero-value legacy mode.
	PropertyIndexValue PropertyIndexMode = iota
	// PropertyIndexConst passes a read-only value.
	PropertyIndexConst
	// PropertyIndexVar passes live caller storage.
	PropertyIndexVar
)

// IndexMode returns the declared mode, defaulting legacy metadata to value mode.
func (p *PropertyInfo) IndexMode(index int) PropertyIndexMode {
	if index < len(p.IndexParamModes) {
		return p.IndexParamModes[index]
	}
	return PropertyIndexValue
}

// IndexMode returns the declared record index mode, defaulting legacy metadata to value.
func (p *RecordPropertyInfo) IndexMode(index int) PropertyIndexMode {
	if index >= 0 && index < len(p.IndexParamModes) {
		return p.IndexParamModes[index]
	}
	return PropertyIndexValue
}
