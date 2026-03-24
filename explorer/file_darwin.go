package explorer

/*
#cgo CFLAGS: -Werror -xobjective-c -fmodules -fobjc-arc

#import <Foundation/Foundation.h>

@interface explorer_file:NSObject
@property NSFileHandle* handler;
@property NSError* err;
@property NSURL* url;
@end

extern CFTypeRef newFile(CFTypeRef url);
extern uint64_t fileRead(CFTypeRef file, uint8_t *b, uint64_t len);
extern uint64_t fileSeek(CFTypeRef file, uint64_t offset, int whence);
extern bool fileWrite(CFTypeRef file, uint8_t *b, uint64_t len);
extern bool fileClose(CFTypeRef file);
extern char* getError(CFTypeRef file);
extern const char* getURL(CFTypeRef url_ref);
extern unsigned long long getSize(CFTypeRef url_ref);

*/
import "C"
import (
	"errors"
	"io"
	"net/url"
	"path/filepath"
	"unsafe"
)

type File struct {
	file   C.CFTypeRef
	size   uint64
	url    string
	closed bool
}

func newFile(url C.CFTypeRef) (*File, error) {
	file := C.newFile(url)
	if err := getError(file); err != nil {
		return nil, err
	}

	cstr := C.getURL(url)
	urlStr := C.GoString(cstr)
	C.free(unsafe.Pointer(cstr))
	size := uint64(C.getSize(url))

	ret := &File{
		file: file,
		size: size,
		url:  urlStr,
	}
	return ret, nil
}

func (f *File) Read(b []byte) (n int, err error) {
	if f.file == 0 || f.closed {
		return 0, io.ErrClosedPipe
	}

	buf := (*C.uint8_t)(unsafe.Pointer(&b[0]))
	length := C.uint64_t(uint64(len(b)))

	if n = int(int64(C.fileRead(f.file, buf, length))); n == 0 {
		if err := getError(f.file); err != nil {
			return n, err
		}
		return n, io.EOF
	}
	return n, nil
}

func (f *File) Write(b []byte) (n int, err error) {
	if f.file == 0 || f.closed {
		return 0, io.ErrClosedPipe
	}

	buf := (*C.uint8_t)(unsafe.Pointer(&b[0]))
	length := C.uint64_t(int64(len(b)))

	if ok := bool(C.fileWrite(f.file, buf, length)); !ok {
		if err := getError(f.file); err != nil {
			return 0, err
		}
		return 0, errors.New("unknown error")
	}
	return len(b), nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f.file == 0 || f.closed {
		return 0, io.ErrClosedPipe
	}

	return int64(C.fileSeek(f.file, C.uint64_t(offset), C.int(whence))), nil
}

func (f *File) Name() string {
	return filepath.Base(f.URI())
}

func (f *File) Size() int64 { return int64(f.size) }

func (f *File) URI() string {
	parsed, err := url.Parse(f.url)
	if err != nil {
		return ""
	}

	return parsed.Path
}

func (f *File) Close() error {
	if ok := bool(C.fileClose(f.file)); !ok {
		return getError(f.file)
	}
	f.closed = true
	return nil
}

func getError(file C.CFTypeRef) error {
	// file will be 0 if the current device doesn't match with @available (i.e older than iOS 13).
	if file == 0 {
		return ErrNotAvailable
	}
	if err := C.GoString(C.getError(file)); len(err) > 0 {
		return errors.New(err)
	}
	return nil
}

// Exported function is required to create cgo header.
//
//export file_darwin
func file_darwin() {}

var (
	_ io.ReadWriteCloser = (*File)(nil)
	_ io.ReadCloser      = (*File)(nil)
	_ io.WriteCloser     = (*File)(nil)
)
