//go:build !js && !windows && !android
// +build !js,!windows,!android

package theme

func isDark() (bool, error) {
	return false, ErrNotAvailableAPI
}

func isReducedMotion() (bool, error) {
	return false, ErrNotAvailableAPI
}
