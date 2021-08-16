//+build android

package notify

import (
	"gioui.org/x/notify/android"
)

type android struct {
	channel *android.NotificationChannel
}

var _ Notifier = (*android)(nil)

func newNotifier() (Notifier, error) {
	channel, err := android.NewChannel(android.ImportanceDefault, "DEFAULT", "niotify", "background notifications")
	if err != nil {
		return Manager{}, err
	}
	return &android{
		channel: channel,
	}, nil
}

func (a *android) CreateNotification(title, text string) (Notification, error) {
	notification, err := a.channel.Send(title, text)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func init() {
	impl, _ = newNotifier()
}
