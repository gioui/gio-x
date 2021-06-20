package explorer

import (
	"errors"
	"gioui.org/io/event"
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

// ReadFile opens the file selector, allowing the user to select a single file.
// It will open the default folder. Optionally, the extensions to filter
// the file extension (such as `.jpg`, `.png`).
//
// Example: ReadFile(".jpg", ".png") will only accept the selection of files with
// .jpg or .png extensions.
func ReadFile(extensions ...string) ([]byte, error) {
	return openFile(extensions...)
}

// WriteFile opens the file selector, and writes the given content into
// some file, which the use can choose the location.
func WriteFile(content []byte, name string) error {
	return createFile(content, name)
}
