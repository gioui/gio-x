//+build linux,!android openbsd freebsd netbsd

package notify

import (
	"fmt"

	"github.com/esiqveland/notify"
	dbus "github.com/godbus/dbus/v5"
)

type linux struct {
	notify.Notifier
}

var _ Notifier = (*linux)(nil)

func newNotifier() (Notifier, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return Manager{}, fmt.Errorf("failed connecting to dbus: %w", err)
	}
	notifier, err := notify.New(conn)
	if err != nil {
		return Manager{}, fmt.Errorf("failed creating notifier: %w", err)
	}
	return &linux{
		Notifier: notifier,
	}, nil
}

type linuxNotification struct {
	id uint32
	*linux
}

var _ notificationInterface = linuxNotification{}

func (l *linux) CreateNotification(title, text string) (Notification, error) {
	id, err := l.Notifier.SendNotification(notify.Notification{
		Summary: title,
		Body:    text,
	})
	if err != nil {
		return nil, err
	}
	return &linuxNotification{
		id:    id,
		linux: l,
	}, nil
}

func (l linuxNotification) Cancel() error {
	_, err := l.linux.CloseNotification(l.id)
	return err
}

func init() {
	impl, _ = newNotifier()
}
