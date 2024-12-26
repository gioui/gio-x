// Package gamepad package allows any Gio application to listen for gamepad input,
// it's supported on Windows 10+, JS, iOS 15+, macOS 12+.
//
// That package was inspired by WebGamepad API (see https://w3c.github.io/gamepad/#gamepad-interface).
//
// You must include `op.InvalidateOp` in your main game loop, otherwise the state of the gamepad will
// not be updated.
package gamepad

import (
	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/event"
	"unsafe"
)

// Gamepad is the main struct and holds information about the state of all Controllers currently available.
// You must use Gamepad.ListenEvents to keep the state up-to-date.
type Gamepad struct {
	Controllers [4]*Controller

	// gamepad varies accordingly with the current OS.
	*gamepad
}

// NewGamepad creates a new Share for the given *app.Window.
// The given app.Window must be unique, and you should call NewGamepad
// once per new app.Window.
//
// It's mandatory to use Gamepad.ListenEvents on the same *app.Window.
func NewGamepad(w *app.Window) *Gamepad {
	return &Gamepad{
		Controllers: [4]*Controller{
			new(Controller),
			new(Controller),
			new(Controller),
			new(Controller),
		},
		gamepad: newGamepad(w),
	}
}

// ListenEvents must get all the events from Gio, in order to get the GioView. You must
// include that function where you listen for Gio events.
//
// Similar as:
//
// select {
// case e := <-window.Events():
// 		gamepad.ListenEvents(e)
// 		switch e := e.(type) {
// 				(( ... your code ...  ))
// 		}
// }
func (g *Gamepad) ListenEvents(evt event.Event) {
	g.listenEvents(evt)
}

// Controller is used to report what Buttons are currently pressed, and where is the position of the Joysticks
// and how much the Triggers are pressed.
type Controller struct {
	Joysticks Joysticks
	Buttons   Buttons

	Connected bool
	Changed   bool
	packet    float64
}

// Joysticks hold the information about the position of the joystick, the position are from -1.0 to 1.0, and
// 0.0 represents the center.
// The maximum and minimum values are:
//        [Y:-1.0]
// [X:-1.0]      [X:+1.0]
//        [Y:+1.0]
type Joysticks struct {
	LeftThumb, RightThumb f32.Point
}

// Buttons hold the information about the state of the buttons, it's based on XBOX Controller scheme.
// The buttons will be informed based on their physical position. Clicking "B" on Nintendo
// gamepad will be "A" since it correspond to same key-position.
//
// That struct must NOT change, or those change must reflect on all maps, which varies per each OS.
//
// Internally, Buttons will be interpreted as [...]Button.
type Buttons struct {
	A, B, Y, X            Button
	Left, Right, Up, Down Button
	LT, RT, LB, RB        Button
	LeftThumb, RightThumb Button
	Start, Back           Button
}

// Button reports if the button is pressed or not, and how much it's pressed (from 0.0 to 1.0 when fully pressed).
type Button struct {
	Pressed bool
	Force   float32
}

func (b *Buttons) setButtonPressed(button int, v bool) {
	bp := (*Button)(unsafe.Add(unsafe.Pointer(b), button))
	bp.Pressed = v
	if v {
		bp.Force = 1.0
	} else {
		bp.Force = 0.0
	}
}

func (b *Buttons) setButtonForce(button int, v float32) {
	bp := (*Button)(unsafe.Add(unsafe.Pointer(b), button))
	bp.Force = v
	bp.Pressed = v > 0
}

const (
	buttonA          = int(unsafe.Offsetof(Buttons{}.A))
	buttonB          = int(unsafe.Offsetof(Buttons{}.B))
	buttonY          = int(unsafe.Offsetof(Buttons{}.Y))
	buttonX          = int(unsafe.Offsetof(Buttons{}.X))
	buttonLeft       = int(unsafe.Offsetof(Buttons{}.Left))
	buttonRight      = int(unsafe.Offsetof(Buttons{}.Right))
	buttonUp         = int(unsafe.Offsetof(Buttons{}.Up))
	buttonDown       = int(unsafe.Offsetof(Buttons{}.Down))
	buttonLT         = int(unsafe.Offsetof(Buttons{}.LT))
	buttonRT         = int(unsafe.Offsetof(Buttons{}.RT))
	buttonLB         = int(unsafe.Offsetof(Buttons{}.LB))
	buttonRB         = int(unsafe.Offsetof(Buttons{}.RB))
	buttonLeftThumb  = int(unsafe.Offsetof(Buttons{}.LeftThumb))
	buttonRightThumb = int(unsafe.Offsetof(Buttons{}.RightThumb))
	buttonStart      = int(unsafe.Offsetof(Buttons{}.Start))
	buttonBack       = int(unsafe.Offsetof(Buttons{}.Back))
)
