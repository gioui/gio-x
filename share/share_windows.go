// SPDX-License-Identifier: Unlicense OR MIT

package share

import (
	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/x/share/internal"
	"github.com/go-ole/go-ole"
	"syscall"
)

type share struct {
	window *app.Window
	hwnd   uintptr

	shareable Shareable

	// The definition of those field lives at share_windows_idl.go:
	// It's  important to keep those values here to prevent the content to be freed
	// by GC, so it must live here "forever".
	_IDataTransferManagerInterop *internal.IDataTransferManagerInterop
	_IDataTransferManager        *internal.IDataTransferManager
	_ITypedEventHandler          *internal.ITypedEventHandler

	_IUriRuntimeClassFactory *internal.IUriRuntimeClassFactory
}

func newShare(w *app.Window) *share {
	return &share{window: w}
}

func (e *Share) init() {
	if err := ole.RoInitialize(1); err != nil {
		return
	}

	if err := internal.NewIDataTransferManagerInterop(&e._IDataTransferManagerInterop); err != nil {
		return
	}

	if err := internal.NewIUriRuntimeClassFactory(&e._IUriRuntimeClassFactory); err != nil {
		return
	}

	if e._IDataTransferManagerInterop == nil || e._IUriRuntimeClassFactory == nil {
		return
	}

	if err := e._IDataTransferManagerInterop.GetForWindow(e.hwnd, &e._IDataTransferManager); err != nil {
		return
	}

	callback := syscall.NewCallback(func(_, _ uintptr, args *internal.IDataRequestedEventArgs) int {
		e.mutex.Lock()
		defer e.mutex.Unlock()

		var dataRequest *internal.IDataRequest
		if err := args.GetRequest(&dataRequest); err != nil {
			return ole.E_FAIL
		}

		var dataPackage *internal.IDataPackage
		if err := dataRequest.GetData(&dataPackage); err != nil {
			return ole.E_FAIL
		}

		var dataPackage2 *internal.IDataPackage2
		if err := dataPackage.GetIDataPackage2(&dataPackage2); err != nil {
			return ole.E_FAIL
		}

		var dataProperty *internal.IDataPackagePropertySet
		if err := dataPackage.GetProperties(&dataProperty); err != nil {
			return ole.E_FAIL
		}

		switch s := e.shareable.(type) {
		case Text:
			dataProperty.SetTitle(s.Title)
			dataPackage.SetText(s.Text)
		case Website:
			dataProperty.SetTitle(s.Title)
			dataPackage.SetText(s.Text)

			var uri *internal.IUriRuntimeClass
			if err := e._IUriRuntimeClassFactory.CreateUri(s.Link, &uri); err != nil {
				return ole.S_OK
			}

			dataPackage2.SetWebLink(uri)
		}

		return ole.S_OK
	})

	e._ITypedEventHandler = internal.NewTypedEventHandler(internal.ITypedEventHandlerDataTransferManagerDataRequestedEventArgsGUID, callback)

	if err := e._IDataTransferManager.AddDataRequested(e._ITypedEventHandler); err != nil {
		return
	}
}

func (e *Share) listenEvents(evt event.Event) {
	switch evt := evt.(type) {
	case app.ViewEvent:
		e.hwnd = evt.HWND
		if e.hwnd != 0 {
			e.init()
		}
	}
}

func (e *Share) shareShareable(shareable Shareable) error {
	if e == nil || e._IDataTransferManagerInterop == nil {
		return ErrNotAvailable
	}

	// Mutex prevents changes of shareable data when Window is triggering the callback.
	e.mutex.Lock()
	e.shareable = shareable
	e.mutex.Unlock()

	if err := e._IDataTransferManagerInterop.ShowShareUIWindow(e.hwnd); err != nil {
		return err
	}
	return nil
}
