//+build !linux,!android

package niotify

type unsupportedManager struct{}

func newManager() (Manager, error) {
	return Manager{unsupportedManager{}}, nil
}

func (u unsupportedManager) CreateNotification(title, text string) (*Notification, error) {
	return &Notification{unsupportedNotification{}}, nil
}

type unsupportedNotification struct{}

func (u unsupportedNotification) Cancel() error {
	return nil
}
