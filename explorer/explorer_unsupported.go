// SPDX-License-Identifier: Unlicense OR MIT

//go:build !windows
// +build !windows

package explorer

import (
	"gioui.org/app"
	"gioui.org/io/event"
	"io"
)

type explorer struct{}

func newExplorer(w *app.Window) *explorer {
	return new(explorer)
}

func (e *Explorer) listenEvents(_ event.Event) {}

func (e *Explorer) exportFile(_ string) (io.WriteCloser, error) {
	return nil, ErrNotAvailable
}

func (e *Explorer) importFile(_ ...string) (io.ReadCloser, error) {
	return nil, ErrNotAvailable
}
