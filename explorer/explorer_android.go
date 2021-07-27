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

var (
	view  uintptr
	mutex = new(sync.Mutex)

	openFileCallback   = make(chan *File, 1)
	createFileCallback = make(chan *File, 1)
)

// listenEvents gets app.ViewEvent from Gio, and (re-)initialize our library.
func listenEvents(event event.Event) {
	if e, ok := event.(app.ViewEvent); ok {
		mutex.Lock()
		view = e.View
		initLib()
		mutex.Unlock()
	}
}

var (
	_LibObject jni.Object
	_LibClass  jni.Class

	_OpenFile   jni.MethodID
	_CreateFile jni.MethodID
	_FileRead   jni.MethodID
	_FileWrite  jni.MethodID
	_FileClose  jni.MethodID
)

func initLib() {
	err := jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		class, err := jni.LoadClass(env, jni.ClassLoaderFor(env, jni.Object(app.AppContext())), "org/gioui/x/explorer/explorer_android")
		if err != nil {
			return err
		}

		obj, err := jni.NewObject(env, class, jni.GetMethodID(env, class, "<init>", `()V`))
		if err != nil {
			return err
		}

		_LibObject = jni.NewGlobalRef(env, obj)
		_LibClass = jni.Class(jni.NewGlobalRef(env, jni.Object(class)))
		_OpenFile = jni.GetMethodID(env, _LibClass, "openFile", "(Landroid/view/View;Ljava/lang/String;)V")
		_CreateFile = jni.GetMethodID(env, _LibClass, "createFile", "(Landroid/view/View;Ljava/lang/String;)V")
		_FileRead = jni.GetMethodID(env, _LibClass, "fileRead", "(Ljava/io/InputStream;[B)I")
		_FileWrite = jni.GetMethodID(env, _LibClass, "fileWrite", "(Ljava/io/OutputStream;[B)Z")
		_FileClose = jni.GetMethodID(env, _LibClass, "fileClose", "(Ljava/io/Closeable;Ljava/io/Flushable;)Z")

		return nil
	})
	if err != nil {
		panic(err)
	}
}

type File struct {
	obj                    jni.Object
	sharedBuffer           jni.Object
	sharedBufferLen        int
	isWritable, isReadable bool
	isClosed               bool
}

func (f *File) Read(b []byte) (n int, err error) {
	if f == nil || !f.isReadable || f.isClosed {
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
		mutex.Lock()
		defer mutex.Unlock()

		n32, err := jni.CallIntMethod(env, _LibObject, _FileRead, jni.Value(f.obj), jni.Value(f.sharedBuffer))
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
		n, err = 0, io.EOF
	}
	return n, err
}

func (f *File) Write(b []byte) (n int, err error) {
	if f == nil || !f.isWritable || f.isClosed {
		return 0, io.ErrClosedPipe
	}
	if len(b) == 0 {
		return 0, nil
	}

	err = jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		mutex.Lock()
		defer mutex.Unlock()

		ok, err := jni.CallBooleanMethod(env, _LibObject, _FileWrite, jni.Value(f.obj), jni.Value(jni.NewByteArray(env, b)))
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

func (f *File) Close() error {
	if f == nil || f.isClosed {
		return io.ErrClosedPipe
	}
	f.isClosed = true

	return jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		mutex.Lock()
		defer mutex.Unlock()

		defer func() {
			if f.obj != 0 {
				jni.DeleteGlobalRef(env, f.obj)
			}
			if f.sharedBuffer != 0 {
				jni.DeleteGlobalRef(env, f.sharedBuffer)
			}
		}()

		flush := jni.Value(0)
		if f.isWritable {
			flush = jni.Value(f.obj)
		}

		ok, err := jni.CallBooleanMethod(env, _LibObject, _FileClose, jni.Value(f.obj), flush)
		if err != nil {
			return err
		}

		if !ok {
			return io.ErrClosedPipe
		}
		return nil
	})
}

func openFile(extensions ...string) (io.ReadCloser, error) {
	for i, ext := range extensions {
		extensions[i] = mime.TypeByExtension(ext)
	}

	err := jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		mutex.Lock()
		defer mutex.Unlock()

		return jni.CallVoidMethod(env, _LibObject, _OpenFile,
			jni.Value(view),
			jni.Value(jni.JavaString(env, strings.Join(extensions, ","))),
		)
	})

	if err != nil {
		return nil, err
	}

	file := <-openFileCallback
	if file == nil {
		return nil, ErrUserDecline
	}
	return file, nil
}

func createFile(name string) (io.WriteCloser, error) {
	err := jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		mutex.Lock()
		defer mutex.Unlock()

		return jni.CallVoidMethod(env, _LibObject, _CreateFile,
			jni.Value(view),
			jni.Value(jni.JavaString(env, strings.TrimPrefix(strings.ToLower(filepath.Ext(name)), "."))),
		)
	})
	if err != nil {
		return nil, err
	}

	file := <-createFileCallback
	if file == nil {
		return nil, ErrUserDecline
	}
	return file, nil
}

//export Java_org_gioui_x_explorer_explorer_1android_OpenCallback
func Java_org_gioui_x_explorer_explorer_1android_OpenCallback(env *C.JNIEnv, _ C.jclass, b C.jobject) {
	if b == 0 {
		openFileCallback <- nil
	} else {
		openFileCallback <- &File{
			obj:        jni.NewGlobalRef(jni.EnvFor(uintptr(unsafe.Pointer(env))), jni.Object(uintptr(b))),
			isReadable: true,
		}
	}
}

//export Java_org_gioui_x_explorer_explorer_1android_CreateCallback
func Java_org_gioui_x_explorer_explorer_1android_CreateCallback(env *C.JNIEnv, _ C.jclass, b C.jobject) {
	if b == 0 {
		createFileCallback <- nil
	} else {
		createFileCallback <- &File{
			obj:        jni.NewGlobalRef(jni.EnvFor(uintptr(unsafe.Pointer(env))), jni.Object(uintptr(b))),
			isWritable: true,
		}
	}
}
