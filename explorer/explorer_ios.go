// SPDX-License-Identifier: Unlicense OR MIT

//go:build ios
// +build ios

package explorer

/*
#cgo CFLAGS: -Werror -xobjective-c -fmodules -fobjc-arc

#include <UIKit/UIKit.h>
#include <stdint.h>

@interface explorer_delegate:NSObject
@end

@interface explorer:NSObject<UIDocumentPickerDelegate>
@property uint64_t mode;
@property (strong) UIDocumentPickerViewController * picker;
@property (strong) UIViewController * controller;
@property uint32_t id;
- (bool) show:(char*)text;
- (bool) exportFile:(char*)name;
- (bool) importFile:(char*)ext;
@end

static const uint64_t IMPORT_MODE = 1;
static const uint64_t EXPORT_MODE = 2;

static CFTypeRef createExplorer(uint64_t mode, CFTypeRef controllerRef, int32_t id) {
	explorer * expl = [[explorer alloc] init];
	expl.mode = mode;
	expl.controller = (__bridge UIViewController *)controllerRef;
	expl.id = id;
	return (__bridge_retained CFTypeRef)expl;
}

static bool show(CFTypeRef expl, char * text) {
	return [(__bridge explorer *)expl show:text];
}

static CFTypeRef fileWriteHandler(CFTypeRef u) {
	NSError *err = nil;
	NSFileHandle *handler = [NSFileHandle fileHandleForWritingToURL:(__bridge NSURL *)u error:&err];
	if (err != nil) {
		return 0;
	}
	return (__bridge_retained CFTypeRef)handler;
}

static CFTypeRef fileReadHandler(CFTypeRef u) {
	NSError *err = nil;
	NSFileHandle *handler = [NSFileHandle fileHandleForReadingFromURL:(__bridge NSURL *)u error:&err];
	if (err != nil) {
		return 0;
	}
	return (__bridge_retained CFTypeRef)handler;
}

static void closeFile(CFTypeRef handler, CFTypeRef u) {
	[(__bridge NSURL *)u stopAccessingSecurityScopedResource];
	[(__bridge NSFileHandle *)handler closeFile];
}

static void fileWrite(CFTypeRef handler, uint8_t* b, uint64_t len) {
	NSData *data = [NSData dataWithBytes:b length:len];

	NSFileHandle * h = (__bridge NSFileHandle *)handler;
	[h writeData:data];
}

static uint64_t fileRead(CFTypeRef handler, uint8_t* b, uint64_t len) {
	if (@available(iOS 14, *)) {
		NSError *err = nil;

		NSFileHandle * h = (__bridge NSFileHandle *)handler;
		NSData *data = [h readDataUpToLength:len error:&err];
		[data getBytes:b length:data.length];

		return data.length;
	}
	return 0;
}
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
	window     *app.Window
	mutex      sync.Mutex
	controller C.CFTypeRef
	importFile C.CFTypeRef
	exportFile C.CFTypeRef
	result     chan result
}

func newExplorer(w *app.Window) *explorer {
	return &explorer{window: w, result: make(chan result)}
}

func (e *Explorer) listenEvents(evt event.Event) {
	switch evt := evt.(type) {
	case app.ViewEvent:
		e.controller = C.CFTypeRef(evt.ViewController)

		e.explorer.exportFile = C.createExplorer(C.EXPORT_MODE, e.controller, C.int32_t(e.id))
		e.explorer.importFile = C.createExplorer(C.IMPORT_MODE, e.controller, C.int32_t(e.id))
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
			if ok := bool(C.show(e.explorer.exportFile, C.CString(name))); !ok {
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
			if ok := bool(C.show(e.explorer.importFile, cextensions)); !ok {
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
		nc = int(int64(C.fileRead(f.handler, buf, C.uint64_t(int64(len(b))))))
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
