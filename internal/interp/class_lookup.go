package interp

func (i *Interpreter) lookupRegisteredClassInfo(name string) *ClassInfo {
	if i == nil {
		return nil
	}
	if i.typeSystem != nil {
		if classInfoAny := i.typeSystem.LookupClass(name); classInfoAny != nil {
			if classInfo, ok := classInfoAny.(*ClassInfo); ok {
				return classInfo
			}
		}
	}
	return nil
}
