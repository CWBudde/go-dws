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

func (i *Interpreter) allRegisteredClassInfos() []*ClassInfo {
	if i == nil {
		return nil
	}

	if i.typeSystem != nil {
		all := i.typeSystem.AllClasses()
		result := make([]*ClassInfo, 0, len(all))
		for _, classInfoAny := range all {
			if classInfo, ok := classInfoAny.(*ClassInfo); ok && classInfo != nil {
				result = append(result, classInfo)
			}
		}
		return result
	}

	return nil
}
