// SPDX-License-Identifier: Unlicense OR MIT

package share

import (
	"crypto/rand"
	"errors"
	"gioui.org/app"
	"gioui.org/io/event"
	"math"
	"math/big"
	"runtime"
	"sync"
)

var (
	// ErrNotAvailable is return when the current OS isn't supported.
	ErrNotAvailable = errors.New("current OS not supported")

	// ErrNotAvailableAction is return when the current Shareable item isn't supported.
	ErrNotAvailableAction = errors.New("current shareable item not supported")
)

type Share struct {
	id    int32
	mutex sync.Mutex

	// share holds OS-Specific content, it varies for each OS.
	*share
}

// active holds all share currently active, that may necessary for callback functions.
//
// Some OSes (Android, iOS, macOS) may call Golang exported functions as callback, but we need
// someway to link that callback with the respective share, in order to give them a response.
//
// In that case, a construction like `callback(..., id int32)` is used. Then, it's possible to get the share
// by lookup the active using the callback id.
//
// To avoid hold dead/unnecessary share, the active will be removed using `runtime.SetFinalizer` on the related
// Share.
var active = sync.Map{} // map[int32]*share

// NewShare creates a new Share for the given *app.Window.
// The given app.Window must be unique, and you should call NewShare
// once per new app.Window.
//
// It's mandatory to use Share.ListenEvents on the same *app.Window.
func NewShare(w *app.Window) (e *Share) {
	idb, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt32))

	e = &Share{
		share: newShare(w),
		id:    int32(idb.Int64()),
	}

	active.Store(e.id, e.share)
	runtime.SetFinalizer(e, func(e *Share) { active.Delete(e.id) })

	return e
}

// ListenEvents must get all the events from Gio, in order to get the GioView. You must
// include that function where you listen for Gio events.
//
// Similar as:
//
// select {
// case e := <-window.Events():
// 		share.ListenEvents(e)
// 		switch e := e.(type) {
// 				(( ... your code ...  ))
// 		}
// }
func (e *Share) ListenEvents(evt event.Event) {
	if e == nil {
		return
	}
	e.listenEvents(evt)
}

type Shareable interface {
	ImplementsShareable()
}

// Text represents the text to be shared.
type Text struct {
	Title string
	Text  string
}

// Website represents the website/link to be shared.
type Website struct {
	Title string
	Text  string
	Link  string
}

func (w Text) ImplementsShareable()    {}
func (w Website) ImplementsShareable() {}

// Share will share the Shareable item, such as Text or Website.
func (e *Share) Share(shareable Shareable) error {
	if e == nil {
		return ErrNotAvailable
	}

	switch shareable.(type) {
	case Text:
	case Website:
	default:
		return ErrNotAvailableAction
	}

	return e.shareShareable(shareable)
}

// ShareText shows the share-dialog to share the given text.
func (e *Share) ShareText(title, text string) error {
	return e.Share(Text{Title: title, Text: text})
}

// ShareWebsite shows the share-dialog to share the given text and link (link have priority over text).
func (e *Share) ShareWebsite(title, text, link string) error {
	return e.Share(Website{Title: title, Text: text, Link: link})
}
