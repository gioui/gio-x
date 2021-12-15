// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios
// +build darwin,!ios

package explorer

/*
#cgo CFLAGS: -Werror -xobjective-c -fmodules -fobjc-arc

@import AppKit;

@interface explorer_macos:NSObject
+ (void) exportFile:(CFTypeRef)viewRef name:(char*)name id:(int32_t)id;
+ (void) importFile:(CFTypeRef)viewRef ext:(char*)ext id:(int32_t)id;
@end

static void exportFile(CFTypeRef viewRef, char * name, int32_t id) {
	[explorer_macos exportFile:viewRef name:name id:id];
}

static void importFile(CFTypeRef viewRef, char * ext, int32_t id) {
	[explorer_macos importFile:viewRef ext:ext id:id];
}
*/
import "C"
import (
	"gioui.org/app"
	"gioui.org/io/event"
	"io"
	"net/url"
	"os"
	"strings"
)

type explorer struct {
	window *app.Window
	view   C.CFTypeRef
	result chan result
}

func newExplorer(w *app.Window) *explorer {
	return &explorer{window: w, result: make(chan result)}
}

func (e *Explorer) listenEvents(evt event.Event) {
	switch evt := evt.(type) {
	case app.ViewEvent:
		e.view = C.CFTypeRef(evt.View)
	}
}

func (e *Explorer) exportFile(name string) (io.WriteCloser, error) {
	cname := C.CString(name)
	e.window.Run(func() { C.exportFile(e.view, cname, C.int32_t(e.id)) })

	resp := <-e.result
	if resp.error != nil {
		return nil, resp.error
	}
	return resp.file.(io.WriteCloser), resp.error

}

func (e *Explorer) importFile(extensions ...string) (io.ReadCloser, error) {
	for i, ext := range extensions {
		extensions[i] = strings.TrimPrefix(ext, ".")
	}

	cextensions := C.CString(strings.Join(extensions, ","))
	e.window.Run(func() { C.importFile(e.view, cextensions, C.int32_t(e.id)) })

	resp := <-e.result
	if resp.error != nil {
		return nil, resp.error
	}
	return resp.file.(io.ReadCloser), resp.error
}

//export importCallback
func importCallback(u *C.char, id C.int32_t) {
	if v, ok := active.Load(int32(id)); ok {
		v := v.(*explorer)
		v.result <- newFile(u, os.Open)
	}
}

//export exportCallback
func exportCallback(u *C.char, id C.int32_t) {
	if v, ok := active.Load(int32(id)); ok {
		v := v.(*explorer)
		v.result <- newFile(u, os.Create)
	}
}

func newFile(u *C.char, action func(s string) (*os.File, error)) result {
	name := C.GoString(u)
	if name == "" {
		return result{error: ErrUserDecline, file: nil}
	}

	uri, err := url.Parse(name)
	if err != nil {
		return result{error: err, file: nil}
	}
	uri.Scheme = ""

	f, err := action(uri.String())
	return result{error: err, file: f}
}
