// +build !android,!windows,!js

package explorer

import (
	"gioui.org/io/event"
)

func listenEvents(_ event.Event) {}

func openFile(extensions ...string) ([]byte, error) {
	return ErrNotAvailable
}

func createFile(content []byte, name string) error {
	return ErrNotAvailable
}
