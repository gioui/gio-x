//go:build !js && !windows && !android && !linux
// +build !js,!windows,!android,!linux

package battery

func batteryLevel() (uint8, error) {
	return 100, ErrNotAvailableAPI
}

func isSavingBattery() (bool, error) {
	return false, ErrNotAvailableAPI
}

func isCharging() (bool, error) {
	return false, ErrNotAvailableAPI
}
