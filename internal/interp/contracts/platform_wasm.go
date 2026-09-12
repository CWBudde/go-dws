//go:build js && wasm

package contracts

import (
	"github.com/cwbudde/go-dws/pkg/platform"
	wasmplatform "github.com/cwbudde/go-dws/pkg/platform/wasm"
)

// DefaultPlatform returns the platform an engine uses when the host installs
// none. Under WASM that is the in-memory virtual filesystem, which a host can
// replace wholesale through the JavaScript `init({fs})` bridge.
func DefaultPlatform() platform.Platform {
	return wasmplatform.NewWASMPlatform()
}
