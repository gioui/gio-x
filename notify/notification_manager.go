// Package notify provides cross-platform notifications for Gio applications.
//
// 	   https://gioui.org
//
// Sending a notification is easy:
//
//     manager, _ := NewManager()
//     notification, _ := manager.CreateNotification("hello!", "I was sent from Gio!")
//     notification.Cancel()
package notify

// Manager provides methods for creating and managing notifications.
type Manager interface {
	CreateNotification(title, text string) (Notification, error)
}

// NewManager creates a new Manager tailored to the current operating system.
func NewManager() (Manager, error) {
	return newManager()
}

// Notification handle that can used to manipulate a platform notification,
// such as by cancelling it.
type Notification interface {
	// Cancel a notification.
	Cancel() error
}

// noop notification for convenience.
type noop struct{}

func (noop) Cancel() error {
	return nil
}
