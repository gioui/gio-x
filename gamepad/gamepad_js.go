package gamepad

import (
	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/system"
	"syscall/js"
)

// mappingButton corresponds to https://w3c.github.io/gamepad/#dom-gamepad-mapping:
var mappingButton = [...]int{
	buttonA,
	buttonB,
	buttonX,
	buttonY,
	buttonLB,
	buttonRB,
	buttonLT,
	buttonRT,
	buttonBack,
	buttonStart,
	buttonLeftThumb,
	buttonRightThumb,
	buttonUp,
	buttonDown,
	buttonLeft,
	buttonRight,
}

type gamepad struct{}

func newGamepad(_ *app.Window) *gamepad {
	return &gamepad{}
}

func (g *Gamepad) listenEvents(evt event.Event) {
	switch evt.(type) {
	case system.FrameEvent:
		g.getState()
	}
}

var (
	_Navigator = js.Global().Get("navigator")
)

func (g *Gamepad) getState() {
	gamepads := _Navigator.Get("getGamepads")
	if !gamepads.Truthy() {
		return
	}

	gamepads = _Navigator.Call("getGamepads")
	for player, controller := range g.Controllers {
		controller.updateState(gamepads.Index(player))
	}
}

func (controller *Controller) updateState(state js.Value) {
	if !state.Truthy() {
		controller.Connected = false
		controller.Changed = false
		return
	}

	packet := state.Get("timestamp").Float()
	if packet == controller.packet {
		controller.Changed = false
		return
	}

	controller.packet = packet
	controller.Connected = true
	controller.Changed = true

	// Buttons
	buttons := state.Get("buttons")
	for index, button := range mappingButton {
		btn := buttons.Index(index)
		force := 0.0
		if btn.Truthy() {
			force = btn.Get("value").Float()
		}
		controller.Buttons.setButtonForce(button, float32(force))
	}

	// Joysticks
	axes := state.Get("axes")
	controller.Joysticks.LeftThumb.X = float32(axes.Index(0).Float())
	controller.Joysticks.LeftThumb.Y = float32(axes.Index(1).Float())
	controller.Joysticks.RightThumb.X = float32(axes.Index(2).Float())
	controller.Joysticks.RightThumb.Y = float32(axes.Index(3).Float())
}
