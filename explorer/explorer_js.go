package explorer

import (
	"gioui.org/io/event"
	"strings"
	"syscall/js"
)

func listenEvents(_ event.Event) {}

func openFile(extensions ...string) ([]byte, error) {
	// TODO: Replace with "File System Access API" when that becomes available on most browsers.
	// BUG: Not work on iOS/Safari.

	openFileCallback := make(chan []byte, 1)

	document := js.Global().Get("document")
	input := document.Call("createElement", "input")
	input.Call("addEventListener", "change", openCallback(openFileCallback))
	input.Set("type", "file")
	input.Get("classList").Call("add", "gio/x/explorer")
	if len(extensions) > 0 {
		input.Set("accept", strings.Join(extensions, ","))
	}

	document.Get("body").Call("appendChild", input)
	input.Call("click")

	content := <-openFileCallback
	if content == nil {
		return nil, ErrUserDecline
	}

	return content, nil
}

func createFile(content []byte, name string) error {
	// TODO: Replace with "File System Access API" when that becomes available on most browsers.

	array := js.Global().Get("Uint8Array").New(len(content))
	// TODO: Move to assembly/CallImport to avoid copy
	js.CopyBytesToJS(array, content)

	config := js.Global().Get("Object").New()
	config.Set("type", "octet/stream")

	data := js.Global().Get("Array").New()
	data = data.Call("concat", array)

	blob := js.Global().Get("Blob").New(data, config)

	document := js.Global().Get("document")
	anchor := document.Call("createElement", "a")
	anchor.Set("download", name)
	anchor.Set("href", js.Global().Get("URL").Call("createObjectURL", blob))
	anchor.Get("classList").Call("add", "gio/x/explorer")
	document.Get("body").Call("appendChild", anchor)
	anchor.Call("click")

	return nil
}

func openCallback(c chan []byte) js.Func {
	// There's no way to detect when the dialog is closed, so we can't re-use the callback.
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		files := args[0].Get("target").Get("files")
		if files.Length() <= 0 {
			c <- nil
			return nil
		}

		fileReader := js.Global().Get("FileReader").New()
		fileReader.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			array := js.Global().Get("Uint8Array").New(fileReader.Get("result"))
			result := make([]byte, array.Length())
			js.CopyBytesToGo(result, array)
			c <- result
			return nil
		}))
		fileReader.Call("readAsArrayBuffer", files.Index(0))

		return nil
	})
}
