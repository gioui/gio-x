package android

import (
	"fmt"

	"gioui.org/app"
	"github.com/tailscale/tailscale-android/jni"
)

type NotificationChannel struct {
}

func NewChannel(jvm uintptr, name, description string) (*NotificationChannel, error) {
	if err := jni.Do(jni.JVMFor(jvm), func(env jni.Env) error {
		classLoader := jni.ClassLoaderFor(env, jni.Object(app.AppContext()))
		notifyClass, err := jni.LoadClass(env, classLoader, "ht/sr/git/whereswaldon/niotify/NotificationHelper")
		if err != nil {
			return err
		}
		newChannelMethod := jni.GetStaticMethodID(env, notifyClass, "newChannel", "()V")
		jname := jni.Value(jni.JavaString(env, name))
		jdescription := jni.Value(jni.JavaString(env, description))
		return jni.CallStaticVoidMethod(env, notifyClass, newChannelMethod, jname, jdescription)
	}); err != nil {
		return nil, fmt.Errorf("failed creating notification channel: %w", err)
	}
	return &NotificationChannel{}, nil
}
