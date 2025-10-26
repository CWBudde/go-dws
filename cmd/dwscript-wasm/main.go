//go:build js && wasm

// Package main is the WebAssembly entry point for the DWScript interpreter.
// It exports the DWScript API to JavaScript and handles the WASM lifecycle.
//
// Build with:
//   GOOS=js GOARCH=wasm go build -o dwscript.wasm ./cmd/dwscript-wasm
//
// Usage from JavaScript:
//   <script src="wasm_exec.js"></script>
//   <script>
//     const go = new Go();
//     WebAssembly.instantiateStreaming(fetch("dwscript.wasm"), go.importObject)
//       .then((result) => {
//         go.run(result.instance);
//         // DWScript API is now available as window.DWScript
//       });
//   </script>
package main

import (
	"syscall/js"

	"github.com/cwbudde/go-dws/pkg/wasm"
)

func main() {
	// Set up a channel to keep the Go program running
	// WASM needs to stay alive to handle JavaScript calls
	done := make(chan struct{})

	// Register the DWScript API with JavaScript
	wasm.RegisterAPI()

	// Log that we're ready
	js.Global().Get("console").Call("log", "DWScript WASM module initialized")

	// Keep the program running
	// Without this, the Go program would exit and all exported functions would be lost
	<-done
}
