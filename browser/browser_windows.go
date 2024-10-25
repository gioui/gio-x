//go:build windows
// +build windows

package browser

import (
	"os/exec"
)

func OpenUrl(url string) error {
	cmd := exec.Command("cmd", "/c", "start", url)
	return cmd.Run()
}
