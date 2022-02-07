//go:build android
// +build android

/*
Package haptic provides access to haptic feedback methods on Android to gio
applications.
*/
package haptic

import (
	"unsafe"

	"gioui.org/app"
	"git.wow.st/gmp/jni"
)

// Buzzer provides methods to trigger haptic feedback
type Buzzer struct {
	// the latest view reference from an app.ViewEvent
	view uintptr
	// updated signals changes to the view field
	updated chan struct{}

	jvm               jni.JVM
	performFeedbackID jni.MethodID
	jvmOperations     chan func(env jni.Env, view jni.Object) error
	errors            chan error
}

// inJVM runs the provided closure within the JVM context associated with this buzzer.
// This method must not be invoked from the same goroutine/thread as the gio event
// processing code, so it's best not to invoke it directly and instead to submit
// closures to b.jvmOperations.
func (b *Buzzer) inJVM(req func(env jni.Env, view jni.Object) error) {
	for b.view == 0 {
		<-b.updated
	}
	if err := jni.Do(b.jvm, func(env jni.Env) error {
		view := *(*jni.Object)(unsafe.Pointer(&b.view))
		return req(env, view)
	}); err != nil {
		b.errors <- err
	}
}

// Buzz attempts to trigger a haptic vibration without blocking. It returns whether
// or not it was successful. If it returns false, it is safe to retry.
func (b *Buzzer) Buzz() bool {
	select {
	case b.jvmOperations <- func(env jni.Env, view jni.Object) error {
		_, err := jni.CallBooleanMethod(env, view, b.performFeedbackID, 0)
		return err
	}:
		return true
	default:
		return false
	}
}

// Shutdown stops the background event loop that interfaces with the JVM.
// Call this when you are done with a Buzzer to allow it to be garbage
// collected. Do not call this method more than once per Buzzer.
func (b *Buzzer) Shutdown() {
	close(b.jvmOperations)
	close(b.errors)
}

// Errors returns a channel of errors from trying to interface with the JVM. This
// channel will close when Shutdown() is invoked.
func (b *Buzzer) Errors() <-chan error {
	if b == nil {
		return nil
	}
	return b.errors
}

// SetView updates the buzzer's internal reference to the android view. This value
// should come from the View field of an app.ViewEvent, and this method should be
// invoked each time an app.ViewEvent is emitted.
func (b *Buzzer) SetView(view uintptr) {
	b.view = view
	// signal the state change if it isn't already being signaled.
	select {
	case b.updated <- struct{}{}:
	default:
	}
}

// NewBuzzer constructs a buzzer.
func NewBuzzer(_ *app.Window) *Buzzer {
	b := &Buzzer{
		updated:       make(chan struct{}, 1),
		jvm:           jni.JVMFor(app.JavaVM()),
		jvmOperations: make(chan func(env jni.Env, view jni.Object) error),
		errors:        make(chan error),
	}
	go func() {
		for op := range b.jvmOperations {
			b.inJVM(op)
		}
	}()
	b.jvmOperations <- func(env jni.Env, view jni.Object) error {
		viewClass := jni.GetObjectClass(env, view)
		b.performFeedbackID = jni.GetMethodID(env, viewClass, "performHapticFeedback", "(I)Z")
		methodID := jni.GetMethodID(env, viewClass, "setHapticFeedbackEnabled", "(Z)V")
		return jni.CallVoidMethod(env, view, methodID, jni.TRUE)
	}
	return b
}
