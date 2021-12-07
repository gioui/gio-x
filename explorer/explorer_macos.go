// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios
// +build darwin,!ios

package explorer

/*
#cgo CFLAGS: -Werror -xobjective-c -fmodules -fobjc-arc

@import AppKit;

@interface explorer_macos:NSObject
+ (void) createFile:(CFTypeRef)viewRef name:(char*)name;
+ (void) openFile:(CFTypeRef)viewRef ext:(char*)ext;
@end

static void createFile(CFTypeRef viewRef, char * name) {
	[explorer_macos createFile:viewRef name:name];
}

static void openFile(CFTypeRef viewRef, char * ext) {
	[explorer_macos openFile:viewRef ext:ext];
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

type file struct {
	error error
	file  *os.File
}

var (
	view C.CFTypeRef

	openFileCallback   = make(chan file, 1)
	createFileCallback = make(chan file, 1)
)

func listenEvents(event event.Event) {
	switch event := event.(type) {
	case app.ViewEvent:
		view = C.CFTypeRef(event.View)
	}
}

func openFile(extensions ...string) (io.ReadCloser, error) {
	winMutex.Lock()
	w := window
	winMutex.Unlock()

	for i, ext := range extensions {
		extensions[i] = strings.TrimPrefix(ext, ".")
	}

	cextensions := C.CString(strings.Join(extensions, ","))
	w.Run(func() { C.openFile(view, cextensions) })

	resp := <-openFileCallback
	return resp.file, resp.error
}

func createFile(name string) (io.WriteCloser, error) {
	winMutex.Lock()
	w := window
	winMutex.Unlock()

	cname := C.CString(name)
	w.Run(func() { C.createFile(view, cname) })

	resp := <-createFileCallback
	return resp.file, resp.error
}

func newFile(u *C.char, action func(s string) (*os.File, error)) file {
	name := C.GoString(u)
	if name == "" {
		return file{error: ErrUserDecline, file: nil}
	}

	uri, err := url.Parse(name)
	if err != nil {
		return file{error: err, file: nil}
	}
	uri.Scheme = ""

	f, err := action(uri.String())
	return file{error: err, file: f}
}

//export openCallback
func openCallback(u *C.char) {
	openFileCallback <- newFile(u, os.Open)
}

//export createCallback
func createCallback(u *C.char) {
	createFileCallback <- newFile(u, os.Create)
}
