//+build windows

package notify

import (
	"github.com/go-toast/toast"
)

type windowsManager struct {
	// icon contains the path to an icon to use.
	// Ignored if empty.
	icon string
}

var _ Manager = &windowsManager{}

func newManager() (Manager, error) {
	return &windowsManager{}, nil
}

// CreateNotification pushes a notification to windows.
// Note; cancellation is not implemented.
func (m *windowsManager) CreateNotification(title, text string) (Notification, error) {
	return noop{}, (&toast.Notification{
		AppID:   title,
		Title:   title,
		Message: text,
		Icon:    m.icon,
	}).Push()

}

// Icon configures an icon to use for notifications, specified as a filepath.
func (m *windowsManager) Icon(path string) {
	m.icon = path
}
