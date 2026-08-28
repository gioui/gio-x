package gamepad

/*
#cgo CFLAGS: -Werror -xobjective-c -fmodules -fobjc-arc

#import <Foundation/Foundation.h>
#import <GameController/GameController.h>

static CFTypeRef getGamepads() {
	if (@available(iOS 15, macOS 12, *)) {
		NSArray<GCController *> * Controllers = [GCController controllers];
		return (CFTypeRef)CFBridgingRetain(Controllers);
	}
	return 0;
}

static CFTypeRef getState(CFTypeRef gamepads, int64_t player) {
	if (@available(iOS 15, macOS 12, *)) {
		NSArray<GCController *> * Controllers = (__bridge NSArray<GCController *> *)gamepads;
		if ([Controllers count] <= player) {
			return 0;
		}

		GCExtendedGamepad * Gamepad = [[Controllers objectAtIndex:player] extendedGamepad];
		if (Gamepad == nil) {
			return 0;
		}

		GCPhysicalInputProfile* Inputs = (GCPhysicalInputProfile*)Gamepad;
		return (CFTypeRef)CFBridgingRetain(Inputs);
	}
	return 0;
}

static double getLastEventFrom(CFTypeRef inputs) {
	if (@available(iOS 15, macOS 12, *)) {
		return (double)(((__bridge GCPhysicalInputProfile*)(inputs)).lastEventTimestamp);
	}
	return 0;
}

static NSString * getKeyName(GCPhysicalInputProfile * Inputs, void * button) {
	if (@available(iOS 15, macOS 12, *)) {
		NSString * name = *((__unsafe_unretained NSString **)(button));
		if ([Inputs hasRemappedElements] == false) {
			return name;
		}
		return [Inputs mappedElementAliasForPhysicalInputName:name];
	}
	return nil;
}

static float getButtonFrom(CFTypeRef inputs, void * button) {
	if (@available(iOS 15, macOS 12, *)) {
		GCPhysicalInputProfile * Inputs = ((__bridge GCPhysicalInputProfile*)(inputs));
		return Inputs.buttons[getKeyName(Inputs, button)].value;
	}
	return 0;
}

static void getAxesFrom(CFTypeRef inputs, void * button, void * x, void * y) {
	if (@available(iOS 15, macOS 12, *)) {
		GCPhysicalInputProfile * Inputs = ((__bridge GCPhysicalInputProfile*)(inputs));
		GCControllerDirectionPad * Pad = Inputs.dpads[getKeyName(Inputs, button)];

		*((float *)(x)) = Pad.xAxis.value;
		*((float *)(y)) = -Pad.yAxis.value;
	}
}
*/
import "C"
import (
	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/system"
	"unsafe"
)

var mappingButton = map[unsafe.Pointer]int{
	unsafe.Pointer(&C.GCInputButtonA):               buttonA,
	unsafe.Pointer(&C.GCInputButtonB):               buttonB,
	unsafe.Pointer(&C.GCInputButtonX):               buttonX,
	unsafe.Pointer(&C.GCInputButtonY):               buttonY,
	unsafe.Pointer(&C.GCInputLeftThumbstickButton):  buttonLeftThumb,
	unsafe.Pointer(&C.GCInputRightThumbstickButton): buttonRightThumb,
	unsafe.Pointer(&C.GCInputLeftShoulder):          buttonLB,
	unsafe.Pointer(&C.GCInputRightShoulder):         buttonRB,
	unsafe.Pointer(&C.GCInputLeftTrigger):           buttonLT,
	unsafe.Pointer(&C.GCInputRightTrigger):          buttonRT,
	unsafe.Pointer(&C.GCInputButtonMenu):            buttonStart,
	unsafe.Pointer(&C.GCInputButtonOptions):         buttonBack,
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

func (g *Gamepad) getState() {
	gamepads := C.getGamepads()
	defer C.CFRelease(gamepads)
	for player, controller := range g.Controllers {
		controller.updateState(C.getState(gamepads, C.int64_t(player)))
	}
}

func (controller *Controller) updateState(state C.CFTypeRef) {
	if state == 0 {
		controller.Connected = false
		controller.Changed = false
		return
	}
	defer C.CFRelease(state)

	packet := float64(C.getLastEventFrom(state))
	if controller.packet == packet {
		controller.Changed = false
		return
	}

	controller.packet = packet
	controller.Connected = true
	controller.Changed = true

	// Buttons
	for name, button := range mappingButton {
		controller.Buttons.setButtonForce(button, float32(C.getButtonFrom(state, name)))
	}

	// D-Pads
	var x, y float32
	C.getAxesFrom(state, unsafe.Pointer(&C.GCInputDirectionPad), unsafe.Pointer(&x), unsafe.Pointer(&y))
	controller.Buttons.setButtonPressed(buttonLeft, x < 0)
	controller.Buttons.setButtonPressed(buttonRight, x > 0)
	controller.Buttons.setButtonPressed(buttonUp, y < 0)
	controller.Buttons.setButtonPressed(buttonDown, y > 0)

	// Joysticks
	C.getAxesFrom(state, unsafe.Pointer(&C.GCInputLeftThumbstick),
		unsafe.Pointer(&controller.Joysticks.LeftThumb.X),
		unsafe.Pointer(&controller.Joysticks.LeftThumb.Y),
	)
	C.getAxesFrom(state, unsafe.Pointer(&C.GCInputRightThumbstick),
		unsafe.Pointer(&controller.Joysticks.RightThumb.X),
		unsafe.Pointer(&controller.Joysticks.RightThumb.Y),
	)
}
