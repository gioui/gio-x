package android

import (
	"fmt"
	"sync"

	"gioui.org/app"
	"github.com/tailscale/tailscale-android/jni"
)

const (
	helperClass = "ht/sr/git/whereswaldon/niotify/NotificationHelper"
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

// NotificationChannel represents a stream of notifications that an application
// provisions on android. Such streams can be selectively enabled and disabled
// by the user, and should be used for different purposes.
type NotificationChannel struct {
	id string
}

// NewChannel creates a new notification channel identified by the provided id
// and with the given user-visible name and description.
func NewChannel(id, name, description string) (*NotificationChannel, error) {
	if err := jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		appCtx := jni.Object(app.AppContext())
		classLoader := jni.ClassLoaderFor(env, appCtx)
		notifyClass, err := jni.LoadClass(env, classLoader, helperClass)
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

// Notification represents a notification that has been requested to be shown to the user.
// This type provides methods to cancel or update the contents of the notification.
type Notification struct {
	id int32
}

// Send creates a new Notification and requests that it be displayed on this channel.
func (nc *NotificationChannel) Send(title, text string) (*Notification, error) {
	notificationID := nextID()
	if err := jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		appCtx := jni.Object(app.AppContext())
		classLoader := jni.ClassLoaderFor(env, appCtx)
		notifyClass, err := jni.LoadClass(env, classLoader, helperClass)
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
		return nil, fmt.Errorf("failed sending notification: %w", err)
	}
	return &Notification{
		id: notificationID,
	}, nil
}

// Cancel removes a previously created notification from display.
func (n *Notification) Cancel() error {
	notificationID := n.id
	if err := jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		appCtx := jni.Object(app.AppContext())
		classLoader := jni.ClassLoaderFor(env, appCtx)
		notifyClass, err := jni.LoadClass(env, classLoader, helperClass)
		if err != nil {
			return err
		}
		newChannelMethod := jni.GetStaticMethodID(env, notifyClass, "cancelNotification", "(Landroid/content/Context;I)V")
		return jni.CallStaticVoidMethod(env, notifyClass, newChannelMethod, jni.Value(app.AppContext()), jni.Value(notificationID))
	}); err != nil {
		return fmt.Errorf("failed cancelling notification: %w", err)
	}
	return nil
}
