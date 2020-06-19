package android

import (
	"fmt"
	"runtime"

	"gioui.org/app"
	"github.com/tailscale/tailscale-android/jni"
)

type NotificationChannel struct {
	javaObj jni.Object
}

func NewChannel(name, description string) (*NotificationChannel, error) {
	var channel jni.Object
	if err := jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		appCtx := jni.Object(app.AppContext())
		classLoader := jni.ClassLoaderFor(env, appCtx)
		notifyClass, err := jni.LoadClass(env, classLoader, "ht/sr/git/whereswaldon/niotify/NotificationHelper")
		if err != nil {
			return err
		}
		newChannelMethod := jni.GetStaticMethodID(env, notifyClass, "newChannel", "(Landroid/content/Context;)Landroid/app/NotificationChannel;")
		// jname := jni.Value(jni.JavaString(env, name))
		// jdescription := jni.Value(jni.JavaString(env, description))
		channel, err = jni.CallStaticObjectMethod(env, notifyClass, newChannelMethod, jni.Value(app.AppContext()))
		if err != nil {
			return err
		}
		channel = jni.NewGlobalRef(env, channel)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed creating notification channel: %w", err)
	}
	nc := &NotificationChannel{
		javaObj: channel,
	}
	runtime.SetFinalizer(nc, func(obj *NotificationChannel) {
		_ = jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
			jni.DeleteGlobalRef(env, obj.javaObj)
			return nil
		})
	})
	return nc, nil
}
