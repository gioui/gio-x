// SPDX-License-Identifier: Unlicense OR MIT

// +build !android,!windows,!js

package explorer

import (
	"gioui.org/io/event"
	"io"
)

func listenEvents(_ event.Event) {}

func openFile(extensions ...string) (io.ReadCloser, error) {
	return nil, ErrNotAvailable
}

func createFile(content []byte, name string) (io.WriteCloser, error) {
	return nil, ErrNotAvailable
}
