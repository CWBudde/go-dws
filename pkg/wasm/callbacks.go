//go:build js && wasm

package wasm

import (
	"syscall/js"
)

// Callbacks manages JavaScript callbacks for Go code.
// This enables JavaScript → Go communication for events like output, errors, etc.
type Callbacks struct {
	onOutput js.Value
	onError  js.Value
	onInput  js.Value
}

// NewCallbacks creates a new Callbacks instance.
func NewCallbacks() *Callbacks {
	return &Callbacks{
		onOutput: js.Null(),
		onError:  js.Null(),
		onInput:  js.Null(),
	}
}

// SetOutputCallback sets the callback for program output.
// JavaScript usage: dws.onOutput = (text) => { ... }
func (c *Callbacks) SetOutputCallback(callback js.Value) {
	if callback.Type() == js.TypeFunction {
		c.onOutput = callback
	}
}

// SetErrorCallback sets the callback for errors.
// JavaScript usage: dws.onError = (error) => { ... }
func (c *Callbacks) SetErrorCallback(callback js.Value) {
	if callback.Type() == js.TypeFunction {
		c.onError = callback
	}
}

// SetInputCallback sets the callback for input requests.
// JavaScript usage: dws.onInput = (prompt) => { return userInput; }
func (c *Callbacks) SetInputCallback(callback js.Value) {
	if callback.Type() == js.TypeFunction {
		c.onInput = callback
	}
}

// Output invokes the output callback if set.
func (c *Callbacks) Output(text string) {
	if !c.onOutput.IsNull() && c.onOutput.Type() == js.TypeFunction {
		c.onOutput.Invoke(text)
	}
}

// Error invokes the error callback if set.
func (c *Callbacks) Error(err error) {
	if !c.onError.IsNull() && c.onError.Type() == js.TypeFunction {
		c.onError.Invoke(WrapError(err))
	}
}

// Input invokes the input callback if set and returns the result.
// Returns empty string if no callback is set.
func (c *Callbacks) Input(prompt string) string {
	if !c.onInput.IsNull() && c.onInput.Type() == js.TypeFunction {
		result := c.onInput.Invoke(prompt)
		if result.Type() == js.TypeString {
			return result.String()
		}
	}
	return ""
}

// HasOutputCallback returns true if an output callback is set.
func (c *Callbacks) HasOutputCallback() bool {
	return !c.onOutput.IsNull() && c.onOutput.Type() == js.TypeFunction
}

// HasErrorCallback returns true if an error callback is set.
func (c *Callbacks) HasErrorCallback() bool {
	return !c.onError.IsNull() && c.onError.Type() == js.TypeFunction
}

// HasInputCallback returns true if an input callback is set.
func (c *Callbacks) HasInputCallback() bool {
	return !c.onInput.IsNull() && c.onInput.Type() == js.TypeFunction
}

// Clear removes all callbacks.
func (c *Callbacks) Clear() {
	c.onOutput = js.Null()
	c.onError = js.Null()
	c.onInput = js.Null()
}
