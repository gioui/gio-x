// SPDX-License-Identifier: Unlicense OR MIT

package explorer

/*
#cgo LDFLAGS: -landroid

#include <jni.h>
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"unsafe"

	"gioui.org/app"
	"gioui.org/io/event"
	"git.wow.st/gmp/jni"
)

//go:generate javac -source 8 -target 8 -bootclasspath $ANDROID_HOME/platforms/android-36/android.jar -d $TEMP/explorer_explorer_android/classes explorer_android.java
//go:generate jar cf explorer_android.jar -C $TEMP/explorer_explorer_android/classes .

type explorer struct {
	window *app.Window
	view   uintptr

	libObject jni.Object
	libClass  jni.Class

	importFile jni.MethodID
	exportFile jni.MethodID

	result chan result
}

func newExplorer(w *app.Window) *explorer {
	return &explorer{window: w, result: make(chan result)}
}

// init will get all necessary MethodID (to future JNI calls) and get our Java library/class (which
// is defined on explorer_android.java file). The Java class doesn't retain information about the view,
// the view (GioView/GioActivity) is passed as argument for each importFile/exportFile function, so it
// can safely change between each call.
func (e *explorer) init(env jni.Env) error {
	if e.libObject != 0 && e.libClass != 0 {
		return nil // Already initialized
	}

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

	fileInfoClass, err := jni.LoadClass(env, jni.ClassLoaderFor(env, jni.Object(app.AppContext())), "org/gioui/x/explorer/explorer_android$FileInfo")
	displayNameId = jni.GetFieldID(env, fileInfoClass, "displayName", "Ljava/lang/String;")
	sizeId = jni.GetFieldID(env, fileInfoClass, "size", "J")

	return nil
}

func (e *Explorer) listenEvents(evt event.Event) {
	if evt, ok := evt.(app.AndroidViewEvent); ok {
		e.view = evt.View
	}
}

func (e *Explorer) exportFile(name string) (io.WriteCloser, error) {
	go e.window.Run(func() {
		err := jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
			if err := e.init(env); err != nil {
				return err
			}

			return jni.CallVoidMethod(env, e.libObject, e.explorer.exportFile,
				jni.Value(e.view),
				jni.Value(jni.JavaString(env, filepath.Base(name))),
				jni.Value(e.id),
			)
		})

		if err != nil {
			e.result <- result{error: err}
		}
	})

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
	go e.window.Run(func() {
		err := jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
			if err := e.init(env); err != nil {
				return err
			}

			return jni.CallVoidMethod(env, e.libObject, e.explorer.importFile,
				jni.Value(e.view),
				jni.Value(jni.JavaString(env, mimes)),
				jni.Value(e.id),
			)
		})

		if err != nil {
			e.result <- result{error: err}
		}
	})

	file := <-e.result
	if file.error != nil {
		return nil, file.error
	}
	return file.file.(io.ReadCloser), nil
}

func (e *Explorer) importFiles(_ ...string) ([]io.ReadCloser, error) {
	return nil, ErrNotAvailable
}

//export Java_org_gioui_x_explorer_explorer_1android_ImportCallback
func Java_org_gioui_x_explorer_explorer_1android_ImportCallback(env *C.JNIEnv, _ C.jclass, stream C.jobject, id C.jint, fileInfo C.jobject, err C.jstring) {
	fileCallback(env, stream, id, fileInfo, err)
}

//export Java_org_gioui_x_explorer_explorer_1android_ExportCallback
func Java_org_gioui_x_explorer_explorer_1android_ExportCallback(env *C.JNIEnv, _ C.jclass, stream C.jobject, id C.jint, fileInfo C.jobject, err C.jstring) {
	fileCallback(env, stream, id, fileInfo, err)
}

func fileCallback(env *C.JNIEnv, stream C.jobject, id C.jint, fileInfo C.jobject, err C.jstring) {
	var res result
	if v, ok := active.Load(int32(id)); ok {
		env := jni.EnvFor(uintptr(unsafe.Pointer(env)))
		if stream == 0 {
			res.error = ErrUserDecline
			if err != 0 {
				if err := jni.GoString(env, jni.String(uintptr(err))); len(err) > 0 {
					res.error = errors.New(err)
				}
			}
		} else {
			if unsafe.Pointer(fileInfo) == nil {
				res.file, res.error = newFile(env, "", 0, jni.NewGlobalRef(env, jni.Object(uintptr(stream))))
			} else {
				name := jni.GoString(env, jni.String(uintptr(jni.GetObjectField(env, jni.Object(uintptr(fileInfo)), displayNameId))))
				size := jni.GetLongField(env, jni.Object(uintptr(fileInfo)), sizeId)
				res.file, res.error = newFile(env, name, size, jni.NewGlobalRef(env, jni.Object(uintptr(stream))))
			}
		}
		v.(*explorer).result <- res
	}
}

var (
	displayNameId jni.FieldID
	sizeId        jni.FieldID
	_             io.ReadCloser  = (*File)(nil)
	_             io.WriteCloser = (*File)(nil)
)
