package explorer

import (
	"gioui.org/app"
	"git.wow.st/gmp/jni"
	"os"
)

//go:generate javac -source 8 -target 8  -bootclasspath $ANDROID_HOME/platforms/android-30/android.jar -d $TEMP/explorer_file_android/classes dir_android.java
//go:generate jar cf dir_android.jar -C $TEMP/explorer_file_android/classes .

type Directory struct {
	createFile jni.MethodID
	libClass   jni.Class
	libObject  jni.Object

	url      jni.Object
	openFile jni.MethodID
}

func newDirectory(env jni.Env, url jni.Object) (*Directory, error) {
	f := &Directory{url: url}

	class, err := jni.LoadClass(env, jni.ClassLoaderFor(env, jni.Object(app.AppContext())), "org/gioui/x/explorer/dir_android")
	if err != nil {
		return nil, err
	}

	obj, err := jni.NewObject(env, class, jni.GetMethodID(env, class, "<init>", `()V`))
	if err != nil {
		return nil, err
	}

	// For some reason, using `f.stream` as argument for a constructor (`public file_android(Object j) {}`) doesn't work.
	if err := jni.CallVoidMethod(env, obj, jni.GetMethodID(env, class, "setHandle", `(Ljava/lang/Object;)V`), jni.Value(f.url)); err != nil {
		return nil, err
	}

	f.libObject = jni.NewGlobalRef(env, obj)
	f.libClass = jni.Class(jni.NewGlobalRef(env, jni.Object(class)))
	f.openFile = jni.GetMethodID(env, f.libClass, "readFile", "(Landroid/view/View;Ljava/lang/String;)Ljava/lang/Object;")
	f.createFile = jni.GetMethodID(env, f.libClass, "writeFile", "(Landroid/view/View;Ljava/lang/String;)Ljava/lang/Object;")

	return f, nil

}

func (d *Directory) ReadFile(name string) (file *File, err error) {
	if d == nil || d.libObject == 0 {
		return nil, os.ErrClosed
	}

	err = jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		nameJava := jni.JavaString(env, name)
		stream, err := jni.CallObjectMethod(env, d.libObject, d.openFile, jni.Value(_View), jni.Value(nameJava))
		if err != nil {
			return err
		}
		if stream == 0 {
			return os.ErrNotExist
		}

		file, err = newFile(env, stream)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return file, nil
}

/*
func (d *Directory) OpenDirectory(name string) (Directory, error) {
	// Implementation for opening a subdirectory in the directory
	return nil, nil
}

*/

func (d *Directory) WriteFile(name string) (file *File, err error) {
	if d == nil || d.libObject == 0 {
		return nil, os.ErrClosed
	}

	err = jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		nameJava := jni.JavaString(env, name)
		stream, err := jni.CallObjectMethod(env, d.libObject, d.createFile, jni.Value(_View), jni.Value(nameJava))
		if err != nil {
			return err
		}
		if stream == 0 {
			return os.ErrExist
		}

		file, err = newFile(env, stream)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return file, nil
}

/*
func (d *Directory) CreateDirectory(name string) (Directory, error) {
	// Implementation for creating a subdirectory in the directory
	return nil, nil
}

func (d *Directory) ListFiles() ([]File, error) {
	// Implementation for listing files in the directory
	return nil, nil
}

func (d *Directory) ListDirectories() ([]Directory, error) {
	// Implementation for listing subdirectories in the directory
	return nil, nil
}

*/
