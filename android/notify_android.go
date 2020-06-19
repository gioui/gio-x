package android

import (
	"fmt"

	"gioui.org/app"
	"github.com/tailscale/tailscale-android/jni"
)

type NotificationChannel struct {
	id string
}

func NewChannel(id, name, description string) (*NotificationChannel, error) {
	if err := jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		appCtx := jni.Object(app.AppContext())
		classLoader := jni.ClassLoaderFor(env, appCtx)
		notifyClass, err := jni.LoadClass(env, classLoader, "ht/sr/git/whereswaldon/niotify/NotificationHelper")
		if err != nil {
			return err
		}
		newChannelMethod := jni.GetStaticMethodID(env, notifyClass, "newChannel", "(Landroid/content/Context;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;)V")
		jname := jni.Value(jni.JavaString(env, name))
		jdescription := jni.Value(jni.JavaString(env, description))
		jID := jni.Value(jni.JavaString(env, id))
		err = jni.CallStaticVoidMethod(env, notifyClass, newChannelMethod, jni.Value(app.AppContext()), jID, jname, jdescription)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed creating notification channel: %w", err)
	}
	nc := &NotificationChannel{
		id: id,
	}
	return nc, nil
}
