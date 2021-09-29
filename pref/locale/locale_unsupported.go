//go:build !js && !windows && !android && !linux
// +build !js,!windows,!android,!linux

package locale

func getLanguage() string {
	return ""
}
