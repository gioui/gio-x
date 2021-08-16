//+build windows

package notify

import (
	"github.com/go-toast/toast"
)

type windows struct {
	// icon contains the path to an icon to use.
	// Ignored if empty.
	icon string
}

var _ Notifier = (*windows)(nil)

func newNotifier() (Notifier, error) {
	return &windows{}, nil
}

// CreateNotification pushes a notification to windows.
// Note; cancellation is not implemented.
func (m *windows) CreateNotification(title, text string) (Notification, error) {
	return noop{}, (&toast.Notification{
		AppID:   title,
		Title:   title,
		Message: text,
		Icon:    m.icon,
	}).Push()

}

// UseIcon configures an icon to use for notifications, specified as a filepath.
func (m *windows) UseIcon(path string) {
	m.icon = path
}
