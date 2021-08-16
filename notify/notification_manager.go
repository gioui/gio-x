// Package notify provides cross-platform notifications for Gio applications.
//
// 	   https://gioui.org
//
// Sending a notification is easy:
//
//     notifier, _ := NewNotifier()
//     notification, _ := notifier.CreateNotification("hello!", "I was sent from Gio!")
//     notification.Cancel()
package notify

// Notifier provides methods for creating and managing notifications.
type Notifier interface {
	CreateNotification(title, text string) (Notification, error)
}

// IconNotifier is a notifier that can display an icon notification.
type IconNotifier interface {
	Notifier
	UseIcon(path string)
}

// NewNotifier creates a new Manager tailored to the current operating system.
func NewNotifier() (Notifier, error) {
	return newNotifier()
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
