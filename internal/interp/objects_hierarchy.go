package interp

// bindClassConstantsToEnv adds all class constants to the current environment,
// allowing methods to access them directly without qualification.
func (i *Interpreter) bindClassConstantsToEnv(classInfo *ClassInfo) {
	for constName, constValue := range classInfo.Constants {
		i.Env().Define(constName, constValue)
	}
}
