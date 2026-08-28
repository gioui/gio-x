package gamepad

import (
	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/system"
	"golang.org/x/sys/windows"
	"math"
	"unsafe"
)

var (
	_XInput         = windows.NewLazySystemDLL("XInput1_4.dll")
	_XInputGetState = _XInput.NewProc("XInputGetState")
	_XInputEnable   = _XInput.NewProc("XInputEnable")
)

// mappingButton corresponds to https://docs.microsoft.com/en-us/windows/win32/api/xinput/ns-xinput-xinput_gamepad.
var mappingButton = map[uint16]int{
	uint16(0x0001): buttonUp,
	uint16(0x0002): buttonDown,
	uint16(0x0004): buttonLeft,
	uint16(0x0008): buttonRight,
	uint16(0x0010): buttonStart,
	uint16(0x0020): buttonBack,
	uint16(0x0040): buttonLeftThumb,
	uint16(0x0080): buttonRightThumb,
	uint16(0x0100): buttonLB,
	uint16(0x0200): buttonRB,
	// uint16(0x0400): Reserved,
	// uint16(0x0800): Reserved,
	uint16(0x1000): buttonA,
	uint16(0x2000): buttonB,
	uint16(0x4000): buttonX,
	uint16(0x8000): buttonY,
}

// _INPUT_STATE is the XINPUT_GAMEPAD.
// See https://docs.microsoft.com/en-us/windows/win32/api/xinput/ns-xinput-xinput_gamepad.
type _XINPUT_STATE struct {
	Packet  uint32
	Gamepad struct {
		Buttons                   uint16
		LeftTrigger, RightTrigger uint8
		LeftThumb, RightThumb     struct {
			X, Y int16
		}
	}
}

type gamepad struct {
	focused bool
}

func newGamepad(_ *app.Window) *gamepad {
	return &gamepad{}
}

func (g *Gamepad) listenEvents(evt event.Event) {
	switch evt := evt.(type) {
	case system.FrameEvent:
		if !g.focused {
			g.inputEnable(true)
		}
		g.getState()
	case system.StageEvent:
		g.inputEnable(evt.Stage == system.StageRunning)
	}
}

func (g *Gamepad) inputEnable(focus bool) {
	b := uint8(0)
	if focus {
		b = 1
	}
	if hr, _, _ := _XInputEnable.Call(uintptr(b)); hr != 0 {
		return
	}
	g.focused = focus
}

func (g *Gamepad) getState() {
	var state _XINPUT_STATE

	for player, controller := range g.Controllers {
		if hr, _, _ := _XInputGetState.Call(uintptr(uint32(player)), uintptr(unsafe.Pointer(&state))); hr != 0 {
			if controller.Connected {
				controller.updateState(false, _XINPUT_STATE{})
			}
			continue
		}

		controller.updateState(true, state)
	}
}

func (controller *Controller) updateState(connected bool, state _XINPUT_STATE) {
	if float64(state.Packet) == controller.packet {
		controller.Changed = false
		return
	}

	controller.packet = float64(state.Packet)
	controller.Connected = connected
	controller.Changed = true

	// Buttons
	for flag, button := range mappingButton {
		controller.Buttons.setButtonPressed(button, state.Gamepad.Buttons&flag == flag)
	}

	// Triggers
	controller.Buttons.LT = forceTrigger(state.Gamepad.LeftTrigger)
	controller.Buttons.RT = forceTrigger(state.Gamepad.RightTrigger)

	// Joysticks
	controller.Joysticks.LeftThumb.X = posJoystick(state.Gamepad.LeftThumb.X)
	controller.Joysticks.LeftThumb.Y = -posJoystick(state.Gamepad.LeftThumb.Y)
	controller.Joysticks.RightThumb.X = posJoystick(state.Gamepad.RightThumb.X)
	controller.Joysticks.RightThumb.Y = -posJoystick(state.Gamepad.RightThumb.Y)
}

func forceTrigger(trigger uint8) Button {
	if trigger == 0 {
		return Button{}
	}
	f := float32(trigger) / float32(math.MaxUint8)
	if f >= 1.0 {
		return Button{Force: 1.0, Pressed: true}
	}
	if f < 0 {
		return Button{Force: 0.0, Pressed: false}
	}
	return Button{Force: f, Pressed: true}
}

func posJoystick(pos int16) float32 {
	if pos == 0 {
		return 0
	}
	p := float32(pos) / float32(math.MaxInt16)
	if p > +1.0 {
		return 1.0
	}
	if p < -1.0 {
		return -1.0
	}
	return p
}
