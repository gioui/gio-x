// SPDX-License-Identifier: Unlicense OR MIT

package share

//go:generate javac -source 8 -target 8  -bootclasspath $ANDROID_HOME/platforms/android-30/android.jar -d $TEMP/explorer/classes share_android.java
//go:generate jar cf share_android.jar -C $TEMP/explorer/classes .

import "C"
import (
	"gioui.org/app"
	"gioui.org/io/event"
	"git.wow.st/gmp/jni"
)

type share struct {
	window *app.Window
	view   uintptr

	shareClass   jni.Class
	shareText    jni.MethodID
	shareWebsite jni.MethodID
}

func newShare(w *app.Window) *share {
	return &share{
		window: w,
	}
}

func (e *share) init() {
	jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		share, err := jni.LoadClass(env, jni.ClassLoaderFor(env, jni.Object(app.AppContext())), "org/gioui/x/share/share_android")
		if err != nil {
			return err
		}

		e.shareClass = jni.Class(jni.NewGlobalRef(env, jni.Object(share)))
		e.shareText = jni.GetStaticMethodID(env, share, "shareText", "(Landroid/view/View;Ljava/lang/String;Ljava/lang/String;)V")
		e.shareWebsite = jni.GetStaticMethodID(env, share, "shareWebsite", "(Landroid/view/View;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;)V")

		return nil
	})
}

func (e *Share) listenEvents(evt event.Event) {
	switch evt := evt.(type) {
	case app.ViewEvent:
		e.view = evt.View
		e.init()
	}
}

func (e *Share) shareShareable(shareable Shareable) error {
	if e == nil || e.shareClass == 0 {
		return ErrNotAvailable
	}

	switch s := shareable.(type) {
	case Text:
		e.window.Run(func() {
			jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
				title, text := jni.JavaString(env, s.Title), jni.JavaString(env, s.Text)
				return jni.CallStaticVoidMethod(env, e.shareClass, e.shareText, jni.Value(e.view), jni.Value(title), jni.Value(text))
			})
		})
	case Website:
		e.window.Run(func() {
			jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
				title, text, link := jni.JavaString(env, s.Title), jni.JavaString(env, s.Text), jni.JavaString(env, s.Link)
				return jni.CallStaticVoidMethod(env, e.shareClass, e.shareWebsite, jni.Value(e.view), jni.Value(title), jni.Value(text), jni.Value(link))
			})
		})
	default:
		return ErrNotAvailableAction
	}
	return nil
}
