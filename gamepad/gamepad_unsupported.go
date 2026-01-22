//go:build !windows && !js && !darwin
// +build !windows,!js,!darwin

package gamepad

import (
	"gioui.org/app"
	"gioui.org/io/event"
)

type gamepad struct{}

func newGamepad(w *app.Window) *gamepad {
	return &gamepad{}
}

func (g *Gamepad) listenEvents(evt event.Event) {
	// NO-OP
}
