//+build linux,!android openbsd freebsd netbsd

package notify

import (
	"fmt"

	"github.com/esiqveland/notify"
	dbus "github.com/godbus/dbus/v5"
)

type linuxManager struct {
	notify.Notifier
}

var _ Manager = &linuxManager{}

func newManager() (Manager, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return Manager{}, fmt.Errorf("failed connecting to dbus: %w", err)
	}
	notifier, err := notify.New(conn)
	if err != nil {
		return Manager{}, fmt.Errorf("failed creating notifier: %w", err)
	}
	return &linuxManager{
		Notifier: notifier,
	}, nil
}

type linuxNotification struct {
	id uint32
	*linuxManager
}

var _ notificationInterface = linuxNotification{}

func (l *linuxManager) CreateNotification(title, text string) (Notification, error) {
	id, err := l.Notifier.SendNotification(notify.Notification{
		Summary: title,
		Body:    text,
	})
	if err != nil {
		return nil, err
	}
	return &linuxNotification{
		id:           id,
		linuxManager: l,
	}, nil
}

func (l linuxNotification) Cancel() error {
	_, err := l.linuxManager.CloseNotification(l.id)
	return err
}
