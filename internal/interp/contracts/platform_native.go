//go:build !js && !wasm

package contracts

import (
	"github.com/cwbudde/go-dws/pkg/platform"
	"github.com/cwbudde/go-dws/pkg/platform/native"
)

// DefaultPlatform returns the platform an engine uses when the host installs
// none. Natively that is the real OS filesystem, console and clock.
//
// The selection is a build-tag pair rather than a runtime check because
// pkg/platform/native is itself `!js && !wasm` — it imports os facilities a
// WASM build cannot link.
func DefaultPlatform() platform.Platform {
	return native.NewNativePlatform()
}
