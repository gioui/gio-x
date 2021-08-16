//+build !linux,!android,!openbsd,!freebsd,!dragonfly,!netbsd,!darwin,!windows

package notify

type unsupportedManager struct{}

func newManager() (Manager, error) {
	return unsupportedManager{}, nil
}

func (u unsupportedManager) CreateNotification(title, text string) (Notification, error) {
	return &noop{}, nil
}
