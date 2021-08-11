// SPDX-License-Identifier: Unlicense OR MIT

package explorer

import (
	"errors"
	"gioui.org/io/event"
	"io"
)

var (
	// ErrUserDecline is returned when the user doesn't select the file.
	ErrUserDecline = errors.New("user exit the file selector without selecting a file")

	// ErrNotAvailable is return when the current OS isn't supported.
	ErrNotAvailable = errors.New("current OS not supported")
)

// ListenEvents must get all the events from Gio, in order to get the GioView. You must
// include that function where you listen for Gio events.
//
// Similar as:
//
// select {
// case e := <-window.Events():
// 		explorer.ListenEvents(e)
// 		switch e := e.(type) {
// 				(( ... your code ...  ))
// 		}
// }
func ListenEvents(event event.Event) {
	listenEvents(event)
}

// ReadFile shows the file selector, allowing the user to select a single file.
// Optionally, it's possible to define which file extensions is supported to
// be selected (such as `.jpg`, `.png`).
//
// Example: ReadFile(".jpg", ".png") will only accept the selection of files with
// .jpg or .png extensions.
//
// In some platforms the resulting `io.ReadCloser` is a `os.File`, but it's not
// a guarantee.
func ReadFile(extensions ...string) (io.ReadCloser, error) {
	return openFile(extensions...)
}

// WriteFile opens the file selector, and writes the given content into
// some file, which the use can choose the location.
//
// It's important to close the `io.WriteCloser`. In some platforms the
// file will be saved only when the writer is closer.
//
// In some platforms the resulting `io.ReadCloser` is a `os.File`, but it's not
// a guarantee.
func WriteFile(name string) (io.WriteCloser, error) {
	return createFile(name)
}
