//go:build (linux && !android) || freebsd || openbsd
// +build linux,!android freebsd openbsd

package browser

import (
	"os/exec"
)

func OpenUrl(url string) error {
	// xdg-open is part of the freedesktop.org suite and should be available on all distro
	cmd := exec.Command("xdg-open", url)
	return cmd.Run()
}
