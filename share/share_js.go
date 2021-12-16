// SPDX-License-Identifier: Unlicense OR MIT

package share

import (
	"gioui.org/app"
	"gioui.org/io/event"
	"syscall/js"
)

type share struct{}

func newShare(w *app.Window) *share {
	return new(share)
}

func (e *Share) listenEvents(_ event.Event) {
	// NO-OP
}

func (e *Share) shareShareable(shareable Shareable) error {
	obj := js.Global().Get("Object").New()
	switch s := shareable.(type) {
	case Text:
		obj.Set("text", s.Text)
		obj.Set("title", s.Title)
	case Website:
		obj.Set("text", s.Text)
		obj.Set("title", s.Title)
		obj.Set("url", s.Link)
	default:
		return ErrNotAvailableAction
	}

	navigator := js.Global().Get("navigator")
	if !navigator.Get("share").Truthy() {
		return ErrNotAvailable
	}

	if !navigator.Call("share", obj).Truthy() {
		return ErrNotAvailable
	}

	return nil
}
