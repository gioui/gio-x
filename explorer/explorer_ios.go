// SPDX-License-Identifier: Unlicense OR MIT

//go:build ios
// +build ios

package explorer

/*
#cgo CFLAGS: -Werror -xobjective-c -fmodules -fobjc-arc

#include <UIKit/UIKit.h>
#include <stdint.h>

// Defined on explorer_ios.m file (implements UIDocumentPickerDelegate).
@interface explorer_picker:NSObject<UIDocumentPickerDelegate>
@property (strong) UIDocumentPickerViewController * picker;
@property (strong) UIViewController * controller;
@property uint64_t mode;
@property uint32_t id;
@end

static const uint64_t IMPORT_MODE = 1;
static const uint64_t EXPORT_MODE = 2;

extern CFTypeRef createPicker(CFTypeRef controllerRef, int32_t id);
extern bool exportFile(CFTypeRef expl, char * name);
extern bool importFile(CFTypeRef expl, char * ext);

extern CFTypeRef fileWriteHandler(CFTypeRef u);
extern CFTypeRef fileReadHandler(CFTypeRef u);

extern bool fileWrite(CFTypeRef handler, uint8_t *b, uint64_t len);
extern uint64_t fileRead(CFTypeRef handler, uint8_t *b, uint64_t len);
extern void closeFile(CFTypeRef handler, CFTypeRef u);
*/
import "C"
import (
	"gioui.org/app"
	"gioui.org/io/event"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"
)

type explorer struct {
	window         *app.Window
	mutex          sync.Mutex
	controller C.CFTypeRef
	picker     C.CFTypeRef
	result     chan result
}

func newExplorer(w *app.Window) *explorer {
	return &explorer{window: w, result: make(chan result)}
}

func (e *Explorer) listenEvents(evt event.Event) {
	switch evt := evt.(type) {
	case app.ViewEvent:
		e.controller = C.CFTypeRef(evt.ViewController)
		e.explorer.picker = C.createPicker(e.controller, C.int32_t(e.id))
	}
}

func (e *Explorer) exportFile(name string) (io.WriteCloser, error) {
	name = filepath.Join(os.TempDir(), name)

	f, err := os.Create(name)
	if err != nil {
		return nil, nil
	}
	f.Close()

	name = "file://" + name

	go func() {
		e.window.Run(func() {
			if ok := bool(C.exportFile(e.explorer.picker, C.CString(name))); !ok {
				e.result <- result{error: ErrNotAvailable}
			}
		})
	}()

	file := <-e.result
	if file.error != nil {
		return nil, file.error
	}
	return file.file.(io.WriteCloser), nil
}

func (e *Explorer) importFile(extensions ...string) (io.ReadCloser, error) {
	for i, ext := range extensions {
		extensions[i] = strings.TrimPrefix(ext, ".")
	}

	cextensions := C.CString(strings.Join(extensions, ","))
	go func() {
		e.window.Run(func() {
			if ok := bool(C.importFile(e.explorer.picker, cextensions)); !ok {
				e.result <- result{error: ErrNotAvailable}
			}
		})
	}()

	file := <-e.result
	if file.error != nil {
		return nil, file.error
	}
	return file.file.(io.ReadCloser), nil
}

type FileReader struct {
	*explorer

	url     C.CFTypeRef
	handler C.CFTypeRef
	closed  bool
}

func newFileReader(e *explorer, url C.CFTypeRef) *FileReader {
	return &FileReader{explorer: e, url: url, handler: C.fileReadHandler(url)}
}

func (f *FileReader) Read(b []byte) (n int, err error) {
	if f.handler == 0 {
		return 0, io.ErrUnexpectedEOF
	}

	buf := (*C.uint8_t)(unsafe.Pointer(&b[0]))

	var nc int
	f.window.Run(func() {
		nc = int(int64(C.fileRead(f.handler, buf, C.uint64_t(uint64(len(b))))))
	})

	if nc == 0 {
		return nc, io.EOF
	}
	return nc, nil
}

func (f *FileReader) Close() error {
	C.closeFile(f.handler, f.url)
	return nil
}

type FileWriter struct {
	*explorer

	url     C.CFTypeRef
	handler C.CFTypeRef
	closed  bool
}

func newFileWriter(e *explorer, url C.CFTypeRef) *FileWriter {
	return &FileWriter{explorer: e, url: url, handler: C.fileWriteHandler(url)}
}

func (f *FileWriter) Write(b []byte) (n int, err error) {
	if f.handler == 0 {
		return 0, io.ErrUnexpectedEOF
	}

	buf := (*C.uint8_t)(unsafe.Pointer(&b[0]))

	f.window.Run(func() {
		C.fileWrite(f.handler, buf, C.uint64_t(int64(len(b))))
	})

	return len(b), nil
}

func (f *FileWriter) Close() error {
	C.closeFile(f.handler, f.url)
	return nil
}

//export importCallback
func importCallback(u C.CFTypeRef, id C.int32_t) {
	if v, ok := active.Load(int32(id)); ok {
		v := v.(*explorer)
		if u == 0 {
			v.result <- result{error: ErrUserDecline}
		} else {
			v.result <- result{file: newFileReader(v, u)}
		}
	}
}

//export exportCallback
func exportCallback(u C.CFTypeRef, id C.int32_t) {
	if v, ok := active.Load(int32(id)); ok {
		v := v.(*explorer)
		if u == 0 {
			v.result <- result{error: ErrUserDecline}
		} else {
			v.result <- result{file: newFileWriter(v, u)}
		}
	}
}
