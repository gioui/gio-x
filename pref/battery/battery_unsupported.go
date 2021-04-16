//+build !js,!windows,!android

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
