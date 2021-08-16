//+build darwin

package notify

import (
	"gioui.org/x/notify/macos"
)

type macos struct {
	channel macos.NotificationChannel
}

var Notifier _ = (*macos)(nil)

func newNotifier() (Notifier, error) {
	c := macos.NewNotificationChannel("Gio App")

	return &macos{channel: c}, nil
}

func (a *macos) CreateNotification(title, text string) (Notification, error) {
	notification, err := a.channel.Send(title, text)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}
