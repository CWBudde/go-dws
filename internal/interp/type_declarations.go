package interp

// lookupInterfaceInfo finds runtime interface metadata by its case-insensitive name.
func (i *Interpreter) lookupInterfaceInfo(name string) *InterfaceInfo {
	return i.LookupInterfaceInfo(name)
}

// LookupInterfaceInfo returns registered runtime interface metadata, or nil.
func (i *Interpreter) LookupInterfaceInfo(name string) *InterfaceInfo {
	return i.typeSystem.LookupInterface(name)
}
