// Package theme provides functions to retrieve user preferences related to theme and accessibility.
package theme

import (
	"errors"
)

// ErrNotAvailableAPI indicates that the current device doesn't support such function.
var ErrNotAvailableAPI = errors.New("pref: not available api")

// IsDarkMode returns "true" if the end-user prefers dark-mode theme.
func IsDarkMode() (bool, error) {
	return isDark()
}

// IsReducedMotion returns "true" if the end-user prefers reduced-motion/disabled-animations.
func IsReducedMotion() (bool, error) {
	return isReducedMotion()
}
