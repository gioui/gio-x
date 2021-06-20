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
	"mime"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"
)

//go:generate javac -source 8 -target 8  -bootclasspath $ANDROID_HOME/platforms/android-30/android.jar -d $TEMP/explorer/classes explorer_android.java
//go:generate jar cf explorer_android.jar -C $TEMP/explorer/classes .

var (
	_Lib = "org/gioui/x/explorer/explorer_android"

	view  uintptr
	mutex = new(sync.Mutex)

	openFileCallback   = make(chan []byte, 1)
	createFileCallback = make(chan bool, 1)
)

func listenEvents(event event.Event) {
	if e, ok := event.(app.ViewEvent); ok {
		mutex.Lock()
		view = e.View
		mutex.Unlock()
	}
}

func openFile(extensions ...string) ([]byte, error) {
	mutex.Lock()
	defer mutex.Unlock()

	for i, ext := range extensions {
		extensions[i] = mime.TypeByExtension(ext)
	}

	open := func(env jni.Env) error {
		class, err := jni.LoadClass(env, jni.ClassLoaderFor(env, jni.Object(app.AppContext())), _Lib)
		if err != nil {
			return err
		}

		obj, err := jni.NewObject(env, class, jni.GetMethodID(env, class, "<init>", `()V`))
		if err != nil {
			return err
		}

		err = jni.CallVoidMethod(env, obj, jni.GetMethodID(env, class, "openFile", "(Landroid/view/View;Ljava/lang/String;)V"),
			jni.Value(view),
			jni.Value(jni.JavaString(env, strings.Join(extensions, ","))),
		)
		if err != nil {
			return err
		}

		return nil
	}

	if err := jni.Do(jni.JVMFor(app.JavaVM()), open); err != nil {
		return nil, err
	}

	content := <-openFileCallback
	if content == nil {
		return nil, ErrUserDecline
	}

	return content, nil
}

func createFile(content []byte, name string) error {
	mutex.Lock()
	defer mutex.Unlock()

	save := func(env jni.Env) error {
		class, err := jni.LoadClass(env, jni.ClassLoaderFor(env, jni.Object(app.AppContext())), _Lib)
		if err != nil {
			return err
		}

		obj, err := jni.NewObject(env, class, jni.GetMethodID(env, class, "<init>", `()V`))
		if err != nil {
			return err
		}

		err = jni.CallVoidMethod(env, obj, jni.GetMethodID(env, class, "createFile", "(Landroid/view/View;Ljava/lang/String;[B)V"),
			jni.Value(view),
			jni.Value(jni.JavaString(env, strings.TrimPrefix(strings.ToLower(filepath.Ext(name)), "."))),
			jni.Value(jni.NewByteArray(env, content)),
		)
		if err != nil {
			return err
		}

		return nil
	}

	if err := jni.Do(jni.JVMFor(app.JavaVM()), save); err != nil {
		return err
	}

	if ok := <-createFileCallback; !ok {
		return ErrUserDecline
	}

	return nil
}

//export Java_org_gioui_x_explorer_explorer_1android_OpenCallback
func Java_org_gioui_x_explorer_explorer_1android_OpenCallback(env *C.JNIEnv, _ C.jclass, b C.jobject) {
	if b == 0 {
		openFileCallback <- nil
	} else {
		openFileCallback <- jni.GetByteArrayElements(jni.EnvFor(uintptr(unsafe.Pointer(env))), jni.ByteArray(uintptr(b)))
	}
}

//export Java_org_gioui_x_explorer_explorer_1android_CreateCallback
func Java_org_gioui_x_explorer_explorer_1android_CreateCallback(env *C.JNIEnv, _ C.jclass, b C.jobject) {
	createFileCallback <- b == jni.TRUE
}
