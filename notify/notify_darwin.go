//+build darwin

package notify

import (
	"gioui.org/x/notify/macos"
)

type macosManager struct {
	channel macos.NotificationChannel
}

func newManager() (Manager, error) {
	c := macos.NewNotificationChannel("Gio App")

	return &macosManager{channel: c}, nil
}

func (a *macosManager) CreateNotification(title, text string) (Notification, error) {
	notification, err := a.channel.Send(title, text)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}
