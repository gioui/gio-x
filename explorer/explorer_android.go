// SPDX-License-Identifier: Unlicense OR MIT

package explorer

/*
#cgo LDFLAGS: -landroid

#include <jni.h>
#include <stdlib.h>
*/
import "C"
import (
	"gioui.org/app"
	"gioui.org/io/event"
	"git.wow.st/gmp/jni"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"
)

//go:generate javac -source 8 -target 8  -bootclasspath $ANDROID_HOME/platforms/android-30/android.jar -d $TEMP/explorer/classes explorer_android.java
//go:generate jar cf explorer_android.jar -C $TEMP/explorer/classes .

type explorer struct {
	window *app.Window
	mutex  sync.Mutex
	view   uintptr

	libObject jni.Object
	libClass  jni.Class

	importFile jni.MethodID
	exportFile jni.MethodID
	fileRead   jni.MethodID
	fileWrite  jni.MethodID
	fileClose  jni.MethodID

	result chan result
}

func newExplorer(w *app.Window) *explorer {
	return &explorer{window: w, result: make(chan result)}
}

func (e *explorer) initLib() {
	err := jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		class, err := jni.LoadClass(env, jni.ClassLoaderFor(env, jni.Object(app.AppContext())), "org/gioui/x/explorer/explorer_android")
		if err != nil {
			return err
		}

		obj, err := jni.NewObject(env, class, jni.GetMethodID(env, class, "<init>", `()V`))
		if err != nil {
			return err
		}

		e.libObject = jni.NewGlobalRef(env, obj)
		e.libClass = jni.Class(jni.NewGlobalRef(env, jni.Object(class)))
		e.importFile = jni.GetMethodID(env, e.libClass, "importFile", "(Landroid/view/View;Ljava/lang/String;I)V")
		e.exportFile = jni.GetMethodID(env, e.libClass, "exportFile", "(Landroid/view/View;Ljava/lang/String;I)V")
		e.fileRead = jni.GetMethodID(env, e.libClass, "fileRead", "(Ljava/io/InputStream;[B)I")
		e.fileWrite = jni.GetMethodID(env, e.libClass, "fileWrite", "(Ljava/io/OutputStream;[B)Z")
		e.fileClose = jni.GetMethodID(env, e.libClass, "fileClose", "(Ljava/io/Closeable;Ljava/io/Flushable;)Z")

		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (e *Explorer) listenEvents(evt event.Event) {
	if evt, ok := evt.(app.ViewEvent); ok {
		e.view = evt.View
		e.initLib()
	}
}

func (e *Explorer) exportFile(name string) (io.WriteCloser, error) {
	go func() {
		jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
			return jni.CallVoidMethod(env, e.libObject, e.explorer.exportFile,
				jni.Value(e.view),
				jni.Value(jni.JavaString(env, strings.TrimPrefix(strings.ToLower(filepath.Ext(name)), "."))),
				jni.Value(e.id),
			)
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
		extensions[i] = mime.TypeByExtension(ext)
	}

	mimes := strings.Join(extensions, ",")
	go func() {
		jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
			return jni.CallVoidMethod(env, e.libObject, e.explorer.importFile,
				jni.Value(e.view),
				jni.Value(jni.JavaString(env, mimes)),
				jni.Value(e.id),
			)
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

	obj             jni.Object
	sharedBuffer    jni.Object
	sharedBufferLen int
	isClosed        bool
}

func newFileReader(e *explorer, obj jni.Object) *FileReader {
	return &FileReader{explorer: e, obj: obj}
}

func (f *FileReader) Read(b []byte) (n int, err error) {
	if f == nil || f.isClosed {
		return 0, io.ErrClosedPipe
	}
	if len(b) == 0 {
		return 0, nil
	}

	if len(b) != f.sharedBufferLen {
		err := jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
			f.sharedBuffer = jni.Object(jni.NewGlobalRef(env, jni.Object(jni.NewByteArray(env, b))))
			return nil
		})
		if err != nil {
			return 0, err
		}
		f.sharedBufferLen = len(b)
	}

	err = jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		n32, err := jni.CallIntMethod(env, f.libObject, f.fileRead, jni.Value(f.obj), jni.Value(f.sharedBuffer))
		if err != nil {
			return err
		}
		n = int(n32)
		if n > 0 {
			n = copy(b, jni.GetByteArrayElements(env, jni.ByteArray(f.sharedBuffer))[:n])
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	if n == -1 {
		return 0, io.EOF
	}
	return n, err
}

func (f *FileReader) Close() error {
	if f == nil || f.isClosed {
		return io.ErrClosedPipe
	}
	f.isClosed = true

	return jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		ok, err := jni.CallBooleanMethod(env, f.libObject, f.fileClose, jni.Value(f.obj), 0)
		if err != nil {
			return err
		}

		jni.DeleteGlobalRef(env, f.sharedBuffer)
		jni.DeleteGlobalRef(env, f.obj)

		if !ok {
			return io.ErrClosedPipe
		}
		return nil
	})
}

type FileWriter struct {
	*explorer

	obj             jni.Object
	sharedBuffer    jni.Object
	sharedBufferLen int
	isClosed        bool
}

func newFileWriter(e *explorer, obj jni.Object) *FileWriter {
	return &FileWriter{explorer: e, obj: obj}
}

func (f *FileWriter) Write(b []byte) (n int, err error) {
	if f == nil || f.isClosed {
		return 0, io.ErrClosedPipe
	}
	if len(b) == 0 {
		return 0, nil
	}

	err = jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		ok, err := jni.CallBooleanMethod(env, f.libObject, f.fileWrite, jni.Value(f.obj), jni.Value(jni.NewByteArray(env, b)))
		if err != nil {
			return err
		}
		if ok {
			n = len(b)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return n, err
}

func (f *FileWriter) Close() error {
	if f == nil || f.isClosed {
		return io.ErrClosedPipe
	}
	f.isClosed = true

	return jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		ok, err := jni.CallBooleanMethod(env, f.libObject, f.fileClose, jni.Value(f.obj), jni.Value(f.obj))
		if err != nil {
			return err
		}

		jni.DeleteGlobalRef(env, f.obj)

		if !ok {
			return io.ErrClosedPipe
		}
		return nil
	})
}

//export Java_org_gioui_x_explorer_explorer_1android_ImportCallback
func Java_org_gioui_x_explorer_explorer_1android_ImportCallback(env *C.JNIEnv, _ C.jclass, b C.jobject, id C.jint) {
	if v, ok := active.Load(int32(id)); ok {
		v := v.(*explorer)
		if b == 0 {
			v.result <- result{error: ErrUserDecline}
		} else {
			v.result <- result{file: newFileReader(v, jni.NewGlobalRef(jni.EnvFor(uintptr(unsafe.Pointer(env))), jni.Object(uintptr(b))))}
		}
	}
}

//export Java_org_gioui_x_explorer_explorer_1android_ExportCallback
func Java_org_gioui_x_explorer_explorer_1android_ExportCallback(env *C.JNIEnv, _ C.jclass, b C.jobject, id C.jint) {
	if v, ok := active.Load(int32(id)); ok {
		v := v.(*explorer)
		if b == 0 {
			v.result <- result{error: ErrUserDecline}
		} else {
			v.result <- result{file: newFileWriter(v, jni.NewGlobalRef(jni.EnvFor(uintptr(unsafe.Pointer(env))), jni.Object(uintptr(b))))}
		}
	}
}

var (
	_ io.ReadCloser  = (*FileReader)(nil)
	_ io.WriteCloser = (*FileWriter)(nil)
)
