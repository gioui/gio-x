//go:build android
// +build android

package notify

import (
	"gioui.org/x/notify/android"
)

type androidNotifier struct {
	channel *android.NotificationChannel
	// ongoing is if the notifications should be ongoing, meaning they cannot be removed.
	ongoing bool
}

var _ Notifier = (*androidNotifier)(nil)

func newNotifier() (Notifier, error) {
	channel, err := android.NewChannel(android.ImportanceDefault, "DEFAULT", "niotify", "background notifications")
	if err != nil {
		return nil, err
	}
	return &androidNotifier{
		channel: channel,
	}, nil
}

func (a *androidNotifier) CreateNotification(title, text string) (Notification, error) {
	notification, err := a.channel.Send(title, text, a.ongoing)
	if err != nil {
		return nil, err
	}
	return notification, nil
}

func (a *androidNotifier) SetOnGoing(ongoing bool) {
	a.ongoing = ongoing
}
