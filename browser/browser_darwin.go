//go:build darwin
// +build darwin

package browser

import (
	"os/exec"
)

func OpenUrl(url string) error {
	cmd := exec.Command("open", url)
	return cmd.Run()
}
