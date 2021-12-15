// SPDX-License-Identifier: Unlicense OR MIT

package explorer

import (
	"crypto/rand"
	"errors"
	"gioui.org/app"
	"gioui.org/io/event"
	"io"
	"math"
	"math/big"
	"runtime"
	"sync"
)

var (
	// ErrUserDecline is returned when the user doesn't select the file.
	ErrUserDecline = errors.New("user exit the file selector without selecting a file")

	// ErrNotAvailable is return when the current OS isn't supported.
	ErrNotAvailable = errors.New("current OS not supported")
)

type result struct {
	file  interface{}
	error error
}

type Explorer struct {
	id    int32
	mutex sync.Mutex

	// explorer holds OS-Specific content, it varies for each OS.
	*explorer
}

// active holds all explorer currently active, that may necessary for callback functions.
//
// Some OSes (Android, iOS, macOS) may call Golang exported functions as callback, but we need
// someway to link that callback with the respective explorer, in order to give them a response.
//
// In that case, a construction like `callback(..., id int32)` is used. Then, it's possible to get the explorer
// by lookup the active using the callback id.
//
// To avoid hold dead/unnecessary explorer, the active will be removed using `runtime.SetFinalizer` on the related
// Explorer.
var active = sync.Map{} // map[int32]*explorer

// NewExplorer creates a new Explorer for the given *app.Window.
// The given app.Window must be unique and you should call NewExplorer
// once per new app.Window.
//
// It's mandatory to use Explorer.ListenEvents on the same *app.Window.
func NewExplorer(w *app.Window) (e *Explorer) {
	idb, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt32))

	e = &Explorer{
		explorer: newExplorer(w),
		id:       int32(idb.Int64()),
	}

	active.Store(e.id, e.explorer)
	runtime.SetFinalizer(e, func(e *Explorer) { active.Delete(e.id) })

	return e
}

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
func (e *Explorer) ListenEvents(evt event.Event) {
	if e == nil {
		return
	}
	e.listenEvents(evt)
}

// Import shows the file selector, allowing the user to select a single file.
// Optionally, it's possible to define which file extensions is supported to
// be selected (such as `.jpg`, `.png`).
//
// Example: ReadFile(".jpg", ".png") will only accept the selection of files with
// .jpg or .png extensions.
//
// In some platforms the resulting `io.ReadCloser` is a `os.File`, but it's not
// a guarantee.
//
// It's a blocking call, you should call it on a separated goroutine. For most OSes, only one
// Import or Export, can happen at the same time, for each app.Window/Explorer.
func (e *Explorer) Import(extensions ...string) (io.ReadCloser, error) {
	if e == nil {
		return nil, ErrNotAvailable
	}

	if runtime.GOOS != "js" {
		e.mutex.Lock()
		defer e.mutex.Unlock()
	}

	return e.importFile(extensions...)
}

// Export opens the file selector, and writes the given content into
// some file, which the use can choose the location.
//
// It's important to close the `io.WriteCloser`. In some platforms the
// file will be saved only when the writer is closer.
//
// In some platforms the resulting `io.WriteCloser` is a `os.File`, but it's not
// a guarantee.
//
// It's a blocking call, you should call it on a separated goroutine. For most OSes, only one
// Import or Export, can happen at the same time, for each app.Window/Explorer.
func (e *Explorer) Export(name string) (io.WriteCloser, error) {
	if e == nil {
		return nil, ErrNotAvailable
	}

	if runtime.GOOS != "js" {
		e.mutex.Lock()
		defer e.mutex.Unlock()
	}

	return e.exportFile(name)
}

var (
	DefaultExplorer *Explorer
)

// ListenEventsWindow calls Explorer.ListenEvents on DefaultExplorer,
// and creates a new Explorer, if needed.
//
// Deprecated: Use NewExplorer instead.
func ListenEventsWindow(win *app.Window, event event.Event) {
	if DefaultExplorer == nil {
		DefaultExplorer = NewExplorer(win)
	}
	DefaultExplorer.ListenEvents(event)
}

// ReadFile calls Explorer.ImportFile on DefaultExplorer.
//
// Deprecated: Use NewExplorer and Explorer.Import instead.
func ReadFile(extensions ...string) (io.ReadCloser, error) {
	return DefaultExplorer.Import(extensions...)
}

// WriteFile calls Explorer.ExportFile on DefaultExplorer.
//
// Deprecated: Use NewExplorer and Explorer.Export instead.
func WriteFile(name string) (io.WriteCloser, error) {
	return DefaultExplorer.Export(name)
}
