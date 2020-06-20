/*
Package niotify provides cross-platform notifications for Gio applications. https://gioui.org

It aims to eventually support all Gio target platforms, but currently only supports
android and linux.

Sending a notification is easy:

    // NOTE: error handling omitted
    // construct a manager
    manager, _ := NewManager()
    // send notification
    notification, _ := manager.CreateNotification("hello!", "I was sent from Gio!")

    // you can also cancel notifications
    notification.Cancel()
*/
package niotify

// Manager provides methods for creating and managing notifications.
type Manager struct {
	impl managerInterface
}

// NewManager creates a new Manager tailored to the current operating system.
func NewManager() (Manager, error) {
	return newManager()
}

// CreateNotification creates and sends a notification on the current platform.
// NOTE: it currently only supports Android and Linux. All other platforms are a
// no-op.
func (m Manager) CreateNotification(title, text string) (*Notification, error) {
	return m.impl.CreateNotification(title, text)
}

// managerInterface is the set of methods required for a cross-platform notification manager.
type managerInterface interface {
	CreateNotification(title, text string) (*Notification, error)
}

// notificationInterface is the set of methods required for a cross-platform notification.
type notificationInterface interface {
	Cancel() error
}

// Notification provides a cross-platform set of methods to manage a sent Notification.
type Notification struct {
	impl notificationInterface
}

// Cancel attempts to remove a previously-created notification from view on the current
// platform.
func (n *Notification) Cancel() error {
	return n.impl.Cancel()
}
