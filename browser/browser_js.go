//go:build js
// +build js

package browser

import (
	"syscall/js"
)

func OpenUrl(url string) error {
	js.Global().Call("open", url, "_blank", "")
	return nil
}
