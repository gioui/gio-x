// Package battery provides functions to get the current status of the batteries.
//
// That is useful to change app behavior based on the battery level. You can reduce animations
// or requests when the battery Level is low or if the user IsSaving battery.
package battery

import (
	"errors"
)

var (
	// ErrNotAvailableAPI indicates that the current device/OS doesn't support such function.
	ErrNotAvailableAPI = errors.New("pref: not available api")

	// ErrNoSystemBattery indicates that the current device doesn't use batteries.
	//
	// Some APIs (like Android and JS) don't provide a mechanism to determine whether the machine uses batteries or not.
	// In such a case ErrNoSystemBattery will never be returned.
	ErrNoSystemBattery = errors.New("pref: device isn't battery-powered")
)

// Level returns the battery level as percent level, between 0~100.
func Level() (uint8, error) {
	return batteryLevel()
}

// IsSaving returns "true" if the end-user enables the battery saver on the device.
func IsSaving() (bool, error) {
	return isSavingBattery()
}

// IsCharging returns "true" if the device is charging.
// If the device doesn't rely on batteries it will be always true.
func IsCharging() (bool, error) {
	return isCharging()
}
