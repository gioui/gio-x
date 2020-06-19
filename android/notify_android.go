package android

import (
	"fmt"
	"sync"

	"gioui.org/app"
	"github.com/tailscale/tailscale-android/jni"
)

var (
	idlock             sync.Mutex
	nextNotificationID int32
)

func nextID() int32 {
	idlock.Lock()
	defer idlock.Unlock()
	id := nextNotificationID
	nextNotificationID++
	return id
}

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

type Notification struct {
	id int32
}

func (nc *NotificationChannel) Send(title, text string) (*Notification, error) {
	notificationID := nextID()
	if err := jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		appCtx := jni.Object(app.AppContext())
		classLoader := jni.ClassLoaderFor(env, appCtx)
		notifyClass, err := jni.LoadClass(env, classLoader, "ht/sr/git/whereswaldon/niotify/NotificationHelper")
		if err != nil {
			return err
		}
		newChannelMethod := jni.GetStaticMethodID(env, notifyClass, "sendNotification", "(Landroid/content/Context;Ljava/lang/String;ILjava/lang/String;Ljava/lang/String;)V")
		jtitle := jni.Value(jni.JavaString(env, title))
		jtext := jni.Value(jni.JavaString(env, text))
		jID := jni.Value(jni.JavaString(env, nc.id))
		err = jni.CallStaticVoidMethod(env, notifyClass, newChannelMethod, jni.Value(app.AppContext()), jID, jni.Value(notificationID), jtitle, jtext)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed creating notification channel: %w", err)
	}
	return &Notification{
		id: notificationID,
	}, nil
}
