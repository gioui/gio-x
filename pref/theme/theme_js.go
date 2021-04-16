package theme

import (
	"syscall/js"
)

var (
	_MatchMedia = js.Global().Get("matchMedia")
)

func isDark() (bool, error) {
	return do("(prefers-color-scheme: dark)")
}

func isReducedMotion() (bool, error) {
	return do("(prefers-reduced-motion: reduce)")
}

func do(name string) (bool, error) {
	if !_MatchMedia.Truthy() {
		return false, ErrNotAvailableAPI
	}

	return _MatchMedia.Invoke(name).Get("matches").Bool(), nil
}
