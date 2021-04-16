// That package is used to reduce code duplication, when using JNI.
package xjni

import (
	"gioui.org/app"
	"git.wow.st/gmp/jni"
)

// DoInt invokes a static int method in the JVM and returns its results. lib is the path to the
// java class with the method, function is the name of the method, signature is the JNI
// string description of the method signature, and args allows providing parameters to
// the method.
func DoInt(lib string, function string, signature string, args ...jni.Value) (i int, err error) {
	err = jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		class, err := jni.LoadClass(env, jni.ClassLoaderFor(env, jni.Object(app.AppContext())), lib)
		if err != nil {
			return err
		}

		i, err = jni.CallStaticIntMethod(env, class, jni.GetStaticMethodID(env, class, function, signature), args...)
		if err != nil {
			return err
		}

		return nil
	})

	return i, err
}

// DoString invokes a static String method in the JVM and returns its results. lib is the path to the
// java class with the method, function is the name of the method, signature is the JNI
// string description of the method signature, and args allows providing parameters to
// the method.
func DoString(lib string, function string, signature string, args ...jni.Value) (s string, err error) {
	err = jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		class, err := jni.LoadClass(env, jni.ClassLoaderFor(env, jni.Object(app.AppContext())), lib)
		if err != nil {
			return err
		}

		o, err := jni.CallStaticObjectMethod(env, class, jni.GetStaticMethodID(env, class, function, signature), args...)
		if err != nil {
			return err
		}

		s = jni.GoString(env, jni.String(o))
		return nil
	})

	return s, err
}
