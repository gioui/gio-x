// SPDX-License-Identifier: Unlicense OR MIT

//go:build !windows && !js && !ios
// +build !windows,!js,!ios

package share

import (
	"gioui.org/app"
	"gioui.org/io/event"
)

type share struct{}

func newShare(w *app.Window) *share {
	return new(share)
}

func (e *Share) listenEvents(_ event.Event) {

}

func (e *Share) shareShareable(shareable Shareable) error {
	return ErrNotAvailable
}
