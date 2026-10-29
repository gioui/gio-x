//go:build android
// +build android

package browser

/*
#cgo LDFLAGS: -landroid

#include <jni.h>
#include <stdlib.h>
*/

import (
	"gioui.org/app"
	"git.wow.st/gmp/jni"
)

//go:generate javac -source 8 -target 8  -bootclasspath $ANDROID_HOME/platforms/android-30/android.jar -d $TEMP/browser_browser_android/classes browser_android.java
//go:generate jar cf browser_android.jar -C $TEMP/browser_browser_android/classes .

func OpenUrl(url string) error {
	err := jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		class, err := jni.LoadClass(env, jni.ClassLoaderFor(env, jni.Object(app.AppContext())), "org/gioui/x/browser/browser_android")
		if err != nil {
			return err
		}

		methodId := jni.GetStaticMethodID(env, class, "openUrl", "(Landroid/content/Context;Ljava/lang/String;)V")
		err = jni.CallStaticVoidMethod(env, class, methodId, jni.Value(app.AppContext()), jni.Value(jni.JavaString(env, url)))
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
