// SPDX-License-Identifier: Unlicense OR MIT

package explorer

import (
	"gioui.org/io/event"
	"io"
	"strings"
	"syscall/js"
)

func listenEvents(_ event.Event) {}

type File struct {
	buffer                 js.Value
	isWritable, isReadable bool
	isClosed               bool
	name                   string
	index                  uint32

	callbackChan             chan js.Value
	successFunc, failureFunc js.Func
}

func (f *File) Read(b []byte) (n int, err error) {
	if f == nil || !f.isReadable || f.isClosed {
		return 0, io.ErrClosedPipe
	}

	if cap(f.callbackChan) == 0 {
		f.callbackChan = make(chan js.Value, 1)

		f.successFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			f.callbackChan <- args[0]
			return nil
		})
		f.failureFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			f.callbackChan <- js.Undefined()
			return nil
		})
	}

	go func() {
		fileSlice(f.index, f.index+uint32(len(b)), f.buffer, f.successFunc, f.failureFunc)
	}()

	buffer := <-f.callbackChan
	n32 := fileRead(buffer, b)
	if n32 == 0 {
		return 0, io.EOF
	}
	f.index += n32

	return int(n32), err
}

func (f *File) Write(b []byte) (n int, err error) {
	if f == nil || !f.isWritable || f.isClosed {
		return 0, io.ErrClosedPipe
	}
	if len(b) == 0 {
		return 0, nil
	}

	fileWrite(f.buffer, b)
	return len(b), err
}

// fileRead and fileWrite calls the JS function directly (without syscall/js to avoid double copying).
// The function is defined into explorer_js.s, which calls explorer_js.js.
func fileRead(value js.Value, b []byte) uint32
func fileWrite(value js.Value, b []byte)
func fileSlice(start, end uint32, value js.Value, success, failure js.Func)

func (f *File) Close() error {
	if f == nil || f.isClosed {
		return io.ErrClosedPipe
	}
	f.isClosed = true

	if f.isReadable {
		f.failureFunc.Release()
		f.successFunc.Release()
		return nil
	}
	return f.saveFile()
}

func openFile(extensions ...string) (io.ReadCloser, error) {
	// TODO: Replace with "File System Access API" when that becomes available on most browsers.
	// BUG: Not work on iOS/Safari.
	fileCallback := make(chan *File, 1)

	document := js.Global().Get("document")
	input := document.Call("createElement", "input")
	input.Call("addEventListener", "change", openCallback(fileCallback))
	input.Set("type", "file")
	if len(extensions) > 0 {
		input.Set("accept", strings.Join(extensions, ","))
	}
	document.Get("body").Call("appendChild", input)
	input.Call("click")

	file := <-fileCallback
	if file == nil {
		return nil, ErrUserDecline
	}

	return file, nil
}

func createFile(name string) (io.WriteCloser, error) {
	// TODO: Replace with "File System Access API" when that becomes available on most browsers.
	return &File{
		name:       name,
		buffer:     js.Global().Get("Uint8Array").New(),
		isWritable: true,
	}, nil
}

func (f *File) saveFile() error {
	config := js.Global().Get("Object").New()
	config.Set("type", "octet/stream")

	blob := js.Global().Get("Blob").New(
		js.Global().Get("Array").New().Call("concat", f.buffer),
		config,
	)

	document := js.Global().Get("document")
	anchor := document.Call("createElement", "a")
	anchor.Set("download", f.name)
	anchor.Set("href", js.Global().Get("URL").Call("createObjectURL", blob))
	document.Get("body").Call("appendChild", anchor)
	anchor.Call("click")

	return nil
}

func openCallback(c chan *File) js.Func {
	// There's no way to detect when the dialog is closed, so we can't re-use the callback.
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		files := args[0].Get("target").Get("files")
		if files.Length() <= 0 {
			c <- nil
			return nil
		}
		c <- &File{buffer: files.Index(0), isReadable: true}
		return nil
	})
}
