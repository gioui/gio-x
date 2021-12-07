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
- (void) createFile:(CFTypeRef)viewRef name:(char*)name;
- (void) openFile:(CFTypeRef)viewRef ext:(char*)ext;
@end

static CFTypeRef createExplorer(uint64_t mode) {
	explorer * expl = [[explorer alloc] init];
	expl.mode = mode;
	return (__bridge_retained CFTypeRef)expl;
}

static void createFile(CFTypeRef viewRef, CFTypeRef expl, char * name) {
	[(__bridge explorer *)expl createFile:viewRef name:name];
}

static void openFile(CFTypeRef viewRef, CFTypeRef expl, char * ext) {
	[(__bridge explorer *)expl openFile:viewRef ext:ext];
}

static CFTypeRef writeFileHandler(CFTypeRef u) {
	NSError *err = nil;
	NSFileHandle *handler = [NSFileHandle fileHandleForWritingToURL:(__bridge NSURL *)u error:&err];
	if (err != nil) {
		return 0;
	}
	return (__bridge_retained CFTypeRef)handler;
}

static CFTypeRef readFileHandler(CFTypeRef u) {
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

static void writeFile(CFTypeRef handler, uint8_t* b, uint64_t len) {
	NSData *data = [NSData dataWithBytes:b length:len];

	NSFileHandle * h = (__bridge NSFileHandle *)handler;
	[h writeData:data];
}

static uint64_t readFile(CFTypeRef handler, uint8_t* b, uint64_t len) {
	if (@available(iOS 14, *)) {
		NSError *err = nil;

		NSFileHandle * h = (__bridge NSFileHandle *)handler;
		NSData *data = [h readDataUpToLength:len error:&err];
		[data getBytes:b length:data.length];
		[h seekToOffset:data.length error:&err];

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
	"unsafe"
)

type File struct {
	url     C.CFTypeRef
	handler C.CFTypeRef
	closed  bool
}

func (f *File) Read(b []byte) (n int, err error) {
	if f.handler == 0 {
		if f.handler = C.readFileHandler(f.url); f.handler == 0 {
			panic("")
		}
	}

	buf := (*C.uint8_t)(unsafe.Pointer(&b[0]))

	winMutex.Lock()
	w := window
	winMutex.Unlock()

	var nc C.uint64_t
	w.Run(func() {
		nc = C.readFile(f.handler, buf, C.uint64_t(int64(len(b))))
	})

	return int(int64(nc)), nil
}

func (f *File) Write(b []byte) (n int, err error) {
	if f.handler == 0 {
		if f.handler = C.writeFileHandler(f.url); f.handler == 0 {
			panic("")
		}
	}

	buf := (*C.uint8_t)(unsafe.Pointer(&b[0]))

	winMutex.Lock()
	w := window
	winMutex.Unlock()

	w.Run(func() {
		C.writeFile(f.handler, buf, C.uint64_t(int64(len(b))))
	})

	return len(b), nil
}

func (f *File) Close() error {
	C.closeFile(f.handler, f.url)
	return nil
}

var (
	view C.CFTypeRef

	createFileCallback = make(chan *File, 1)
	openFileCallback   = make(chan *File, 1)
)

func listenEvents(event event.Event) {
	switch event := event.(type) {
	case app.ViewEvent:
		view = C.CFTypeRef(event.ViewController)
	}
}

func openFile(extensions ...string) (io.ReadCloser, error) {
	winMutex.Lock()
	w := window
	winMutex.Unlock()

	for i, ext := range extensions {
		extensions[i] = strings.TrimPrefix(ext, ".")
	}

	explorerRead := C.createExplorer(C.uint64_t(int64(1)))
	cextensions := C.CString(strings.Join(extensions, ","))
	w.Run(func() { C.openFile(view, explorerRead, cextensions) })

	resp := <-openFileCallback
	if resp == nil {
		return nil, ErrUserDecline
	}
	return resp, nil
}

func createFile(name string) (io.WriteCloser, error) {
	name = filepath.Join(os.TempDir(), name)

	f, err := os.Create(name)
	if err != nil {
		return nil, nil
	}
	f.Close()

	name = "file://" + name

	winMutex.Lock()
	w := window
	winMutex.Unlock()

	explorerWrite := C.createExplorer(C.uint64_t(int64(0)))
	w.Run(func() { C.createFile(view, explorerWrite, C.CString(name)) })

	resp := <-createFileCallback
	if resp == nil {
		return nil, ErrUserDecline
	}
	return resp, nil
}

//export openCallback
func openCallback(u C.CFTypeRef) {
	if u == 0 {
		openFileCallback <- nil
	} else {
		openFileCallback <- &File{url: u}
	}
}

//export createCallback
func createCallback(u C.CFTypeRef) {
	if u == 0 {
		createFileCallback <- nil
	} else {
		createFileCallback <- &File{url: u}
	}
}
