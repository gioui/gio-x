//go:build ios && cgo
// +build ios,cgo

package haptic

//#cgo LDFLAGS: -framework AudioToolbox

/*
#cgo CFLAGS: -x objective-c -fno-objc-arc -fmodules
#pragma clang diagnostic ignored "-Wformat-security"
@import AudioToolbox;

SystemSoundID buzzID = 1520;

void buzz() {
    AudioServicesPlaySystemSound(buzzID);
}
*/
import "C"

import (
	"gioui.org/app"
)

// Buzzer provides methods to trigger haptic feedback. On OSes other than android,
// all methods are no-ops.
type Buzzer struct {
}

// Buzz attempts to trigger a haptic vibration without blocking. It returns whether
// or not it was successful. If it returns false, it is safe to retry. On unsupported
// platforms, it always returns true.
func (b *Buzzer) Buzz() bool {
	C.buzz()
	return true
}

// Update does nothing on platforms other than Android. See the documentation with
// GOOS=android for information on using this method correctly on that platform.
func (b *Buzzer) SetView(_ uintptr) {
}

// Shutdown stops the background event loop that interfaces with the JVM.
// Call this when you are done with a Buzzer to allow it to be garbage
// collected. Do not call this method more than per Buzzer.
func (b *Buzzer) Shutdown() {
}

// Errors returns a channel of errors from trying to interface with the JVM. This
// channel will close when Shutdown() is invoked.
func (b *Buzzer) Errors() <-chan error {
	return nil
}

// NewBuzzer constructs a buzzer.
func NewBuzzer(_ *app.Window) *Buzzer {
	return &Buzzer{}
}
