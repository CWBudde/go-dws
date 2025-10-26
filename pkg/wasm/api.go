//go:build js && wasm

// Package wasm provides WebAssembly-specific functionality for go-dws,
// including JavaScript/Go interop and browser API bindings.
package wasm

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"syscall/js"
	"time"

	"github.com/cwbudde/go-dws/pkg/dwscript"
	"github.com/cwbudde/go-dws/pkg/platform/wasm"
)

// RegisterAPI registers the DWScript API with JavaScript.
// This makes the DWScript class available as window.DWScript.
func RegisterAPI() {
	// Create the DWScript constructor function
	constructor := js.FuncOf(newDWScriptInstance)
	js.Global().Set("DWScript", constructor)
}

// newDWScriptInstance creates a new DWScript instance for JavaScript.
// This is called when JavaScript does: new DWScript()
func newDWScriptInstance(this js.Value, args []js.Value) interface{} {
	// Recover from panics and return them as JavaScript errors
	defer func() {
		if r := recover(); r != nil {
			ConsoleError(fmt.Sprintf("Panic creating DWScript instance: %v", r))
		}
	}()

	// Create a Go DWScript engine
	engine, err := dwscript.New()
	if err != nil {
		return CreateErrorObject("InitializationError", err.Error(), nil)
	}

	// Create platform with output capture
	var outputBuffer bytes.Buffer
	wasmPlat := wasm.NewWASMPlatformWithIO(&outputBuffer)

	// Create callbacks system
	callbacks := NewCallbacks()

	// Store engine and platform in a context
	ctx := &Context{
		engine:       engine,
		platform:     wasmPlat.(*wasm.WASMPlatform),
		outputBuffer: &outputBuffer,
		callbacks:    callbacks,
		programs:     make(map[int]*dwscript.Program),
		funcRefs:     make([]js.Func, 0),
	}

	// Create JavaScript object with methods
	obj := js.Global().Get("Object").New()

	// Bind methods and store function references for cleanup
	ctx.bindMethod(obj, "init", initFunc)
	ctx.bindMethod(obj, "compile", compileFunc)
	ctx.bindMethod(obj, "run", runFunc)
	ctx.bindMethod(obj, "eval", evalFunc)
	ctx.bindMethod(obj, "on", onFunc)
	ctx.bindMethod(obj, "setFileSystem", setFileSystemFunc)
	ctx.bindMethod(obj, "version", versionFunc)
	ctx.bindMethod(obj, "dispose", disposeFunc)

	return obj
}

// Context holds the state for a DWScript instance.
type Context struct {
	engine       *dwscript.Engine
	platform     *wasm.WASMPlatform
	outputBuffer *bytes.Buffer
	callbacks    *Callbacks
	programs     map[int]*dwscript.Program
	nextID       int
	funcRefs     []js.Func // Store func references for proper cleanup
}

// bindMethod binds a method to the JavaScript object and tracks the function reference.
func (ctx *Context) bindMethod(obj js.Value, name string, fn func(*Context, []js.Value) interface{}) {
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Wrap in panic recovery
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				ConsoleError(fmt.Sprintf("Panic in %s: %v\n%s", name, r, stack))
			}
		}()
		return fn(ctx, args)
	})
	ctx.funcRefs = append(ctx.funcRefs, jsFunc)
	obj.Set(name, jsFunc)
}

// initFunc initializes the DWScript instance with options.
// JavaScript usage: await dws.init({ onOutput: fn, onError: fn, fs: customFS })
func initFunc(ctx *Context, args []js.Value) interface{} {
	// Options are optional
	if len(args) > 0 && args[0].Type() == js.TypeObject {
		options := args[0]

		// Set output callback if provided
		if onOutput := options.Get("onOutput"); onOutput.Type() == js.TypeFunction {
			ctx.callbacks.SetOutputCallback(onOutput)
		}

		// Set error callback if provided
		if onError := options.Get("onError"); onError.Type() == js.TypeFunction {
			ctx.callbacks.SetErrorCallback(onError)
		}

		// Set input callback if provided
		if onInput := options.Get("onInput"); onInput.Type() == js.TypeFunction {
			ctx.callbacks.SetInputCallback(onInput)
		}

		// Set custom filesystem if provided
		if fs := options.Get("fs"); !fs.IsNull() && !fs.IsUndefined() {
			// TODO: Implement custom filesystem integration
			ConsoleWarn("Custom filesystem not yet implemented")
		}
	}

	// Return a resolved promise for async compatibility
	promise := js.Global().Get("Promise")
	return promise.Call("resolve", js.Null())
}

// compileFunc compiles DWScript source code and returns a program ID.
// JavaScript usage: program = dws.compile(sourceCode)
func compileFunc(ctx *Context, args []js.Value) interface{} {
	if len(args) < 1 {
		return CreateErrorObject("ArgumentError", "compile requires 1 argument: source code", nil)
	}

	sourceCode := args[0].String()

	// Compile the program
	program, err := ctx.engine.Compile(sourceCode)
	if err != nil {
		return CreateErrorObject("CompileError", err.Error(), map[string]interface{}{
			"source": sourceCode,
		})
	}

	// Store program and assign ID
	ctx.nextID++
	programID := ctx.nextID
	ctx.programs[programID] = program

	// Return program object
	result := js.Global().Get("Object").New()
	result.Set("id", programID)
	result.Set("success", true)
	return result
}

// runFunc executes a previously compiled program.
// JavaScript usage: result = dws.run(program)
func runFunc(ctx *Context, args []js.Value) interface{} {
	if len(args) < 1 {
		return CreateErrorObject("ArgumentError", "run requires 1 argument: program object", nil)
	}

	programObj := args[0]
	if !programObj.Get("id").Truthy() {
		return CreateErrorObject("ArgumentError", "invalid program object", nil)
	}

	programID := programObj.Get("id").Int()

	program, exists := ctx.programs[programID]
	if !exists {
		return CreateErrorObject("ProgramError", fmt.Sprintf("program not found: %d", programID), nil)
	}

	// Clear output buffer
	ctx.outputBuffer.Reset()

	// Execute program
	startTime := time.Now()
	_, err := ctx.engine.Run(program)
	executionTime := time.Since(startTime).Milliseconds()

	// Emit output event if there's output
	output := ctx.outputBuffer.String()
	if output != "" && ctx.callbacks.HasOutputCallback() {
		ctx.callbacks.Output(output)
	}

	// Build result object
	resultObj := js.Global().Get("Object").New()
	resultObj.Set("success", err == nil)
	resultObj.Set("output", output)
	resultObj.Set("executionTime", executionTime)

	if err != nil {
		errObj := CreateErrorObject("RuntimeError", err.Error(), map[string]interface{}{
			"executionTime": executionTime,
		})
		resultObj.Set("error", errObj)

		// Emit error event
		if ctx.callbacks.HasErrorCallback() {
			ctx.callbacks.Error(err)
		}
	}

	return resultObj
}

// evalFunc compiles and runs DWScript code in one step.
// JavaScript usage: result = dws.eval(sourceCode)
func evalFunc(ctx *Context, args []js.Value) interface{} {
	if len(args) < 1 {
		return CreateErrorObject("ArgumentError", "eval requires 1 argument: source code", nil)
	}

	sourceCode := args[0].String()

	// Clear output buffer
	ctx.outputBuffer.Reset()

	// Compile and run
	startTime := time.Now()
	_, err := ctx.engine.Eval(sourceCode)
	executionTime := time.Since(startTime).Milliseconds()

	// Emit output event if there's output
	output := ctx.outputBuffer.String()
	if output != "" && ctx.callbacks.HasOutputCallback() {
		ctx.callbacks.Output(output)
	}

	// Build result object
	resultObj := js.Global().Get("Object").New()
	resultObj.Set("success", err == nil)
	resultObj.Set("output", output)
	resultObj.Set("executionTime", executionTime)

	if err != nil {
		errObj := CreateErrorObject("RuntimeError", err.Error(), map[string]interface{}{
			"source":        sourceCode,
			"executionTime": executionTime,
		})
		resultObj.Set("error", errObj)

		// Emit error event
		if ctx.callbacks.HasErrorCallback() {
			ctx.callbacks.Error(err)
		}
	}

	return resultObj
}

// onFunc registers an event listener.
// JavaScript usage: dws.on('output', (text) => {...})
func onFunc(ctx *Context, args []js.Value) interface{} {
	if len(args) < 2 {
		return CreateErrorObject("ArgumentError", "on requires 2 arguments: event name and callback", nil)
	}

	eventName := args[0].String()
	callback := args[1]

	if callback.Type() != js.TypeFunction {
		return CreateErrorObject("ArgumentError", "callback must be a function", nil)
	}

	switch eventName {
	case "output":
		ctx.callbacks.SetOutputCallback(callback)
	case "error":
		ctx.callbacks.SetErrorCallback(callback)
	case "input":
		ctx.callbacks.SetInputCallback(callback)
	default:
		return CreateErrorObject("ArgumentError", fmt.Sprintf("unknown event: %s", eventName), nil)
	}

	return js.Null()
}

// setFileSystemFunc sets a custom filesystem implementation.
// JavaScript usage: dws.setFileSystem(customFS)
func setFileSystemFunc(ctx *Context, args []js.Value) interface{} {
	if len(args) < 1 {
		return CreateErrorObject("ArgumentError", "setFileSystem requires 1 argument: filesystem object", nil)
	}

	// TODO: Implement custom filesystem integration
	// For now, just log a warning
	ConsoleWarn("Custom filesystem not yet implemented")

	return js.Null()
}

// versionFunc returns version information.
// JavaScript usage: version = dws.version()
func versionFunc(ctx *Context, args []js.Value) interface{} {
	result := js.Global().Get("Object").New()
	result.Set("version", "0.1.0")
	result.Set("build", "wasm")
	result.Set("platform", "javascript")
	return result
}

// disposeFunc cleans up resources and releases function references.
// JavaScript usage: dws.dispose()
func disposeFunc(ctx *Context, args []js.Value) interface{} {
	// Release all function references
	for _, fn := range ctx.funcRefs {
		fn.Release()
	}
	ctx.funcRefs = nil

	// Clear programs
	ctx.programs = nil

	// Clear callbacks
	ctx.callbacks.Clear()

	ConsoleLog("DWScript instance disposed")
	return js.Null()
}
