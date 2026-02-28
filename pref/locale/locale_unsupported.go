//go:build !js && !windows && !android && !linux && !darwin
// +build !js,!windows,!android,!linux,!darwin

package locale

func getLanguage() string {
	return ""
}
