package interp

import (
	"math/rand"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/units"
)

func (i *Interpreter) exceptionValue() *runtime.ExceptionValue {
	exc, _ := i.ctx.Exception().(*runtime.ExceptionValue)
	return exc
}

func (i *Interpreter) setExceptionValue(exc *runtime.ExceptionValue) {
	i.ctx.SetException(exc)
}

func (i *Interpreter) clearException() {
	i.ctx.SetException(nil)
}

func (i *Interpreter) callStackTrace() errors.StackTrace {
	return i.ctx.CallStack()
}

func (i *Interpreter) unitRegistry() *units.UnitRegistry {
	return i.engineState.UnitRegistry
}

func (i *Interpreter) loadedUnits() []string {
	return i.engineState.LoadedUnits
}

func (i *Interpreter) addLoadedUnit(unitName string) {
	i.engineState.LoadedUnits = append(i.engineState.LoadedUnits, unitName)
}

func (i *Interpreter) initializedUnits() map[string]bool {
	return i.engineState.InitializedUnits
}

func (i *Interpreter) externalFunctions() *ExternalFunctionRegistry {
	registry, _ := i.engineState.ExternalFunctions.(*ExternalFunctionRegistry)
	return registry
}

func (i *Interpreter) sourceCode() string {
	return i.engineState.SourceCode
}

func (i *Interpreter) sourceFile() string {
	return i.engineState.SourceFile
}

func (i *Interpreter) randomSource() *rand.Rand {
	return i.engineState.Random
}

func (i *Interpreter) randomSeed() int64 {
	return i.engineState.RandomSeed
}

func (i *Interpreter) setRandomSeed(seed int64) {
	i.engineState.RandomSeed = seed
	source := rand.NewSource(seed)
	i.engineState.Random = rand.New(source)
}

func (i *Interpreter) refCountManager() runtime.RefCountManager {
	return i.engineState.RefCountManager
}
