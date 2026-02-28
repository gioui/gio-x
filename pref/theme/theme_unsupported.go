//go:build !js && !windows && !android && !darwin
// +build !js,!windows,!android,!darwin

package theme

func isDark() (bool, error) {
	return false, ErrNotAvailableAPI
}

func isReducedMotion() (bool, error) {
	return false, ErrNotAvailableAPI
}
