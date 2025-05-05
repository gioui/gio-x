package battery

import (
	"errors"
	"math"
	"syscall/js"
)

var _GetBattery = js.Global().Get("navigator").Get("getBattery")

func batteryLevel() (uint8, error) {
	value, err := do("level")
	if err != nil || !value.Truthy() {
		return 100, err
	}

	b := uint8(math.Ceil(value.Float() * 100))
	switch {
	case b > 100:
		return 100, nil
	case b < 0:
		return 0, nil
	default:
		return b, nil
	}
}

func isSavingBattery() (bool, error) {
	return false, ErrNotAvailableAPI
}

func isCharging() (bool, error) {
	value, err := do("charging")
	if err != nil || !value.Truthy() {
		return false, err
	}

	return value.Bool(), nil
}

func do(name string) (js.Value, error) {
	if !_GetBattery.Truthy() {
		return js.Value{}, ErrNotAvailableAPI
	}

	var (
		success, failure js.Func

		value = make(chan js.Value, 1)
		err   = make(chan error, 1)
	)

	success = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		success.Release()
		failure.Release()

		value <- args[0].Get(name)

		return nil
	})

	failure = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		success.Release()
		failure.Release()

		err <- errors.New("failure getting battery")

		return nil
	})

	go func() {
		js.Global().Get("navigator").Call("getBattery").Call("then", success, failure)
	}()

	select {
	case value := <-value:
		return value, nil
	case <-err:
		return js.Value{}, ErrNotAvailableAPI
	}
}
