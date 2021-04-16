package locale

import (
	"syscall/js"
)

var (
	_Navigator = js.Global().Get("navigator")
)

func getLanguage() string {
	if !_Navigator.Truthy() {
		return ""
	}

	return _Navigator.Get("language").String()
}
