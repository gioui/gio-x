## notify

[![Go Reference](https://pkg.go.dev/badge/gioui.org/x/notify.svg)](https://pkg.go.dev/gioui.org/x/notify)

Cross platform notifications for [Gio](https://gioui.org) applications.

## Status

This repo is experimental, and does not have a stable interface. Currently niotify
only supports the following OSes:

- linux (x11/wayland doesn't matter so long as dbus is used for notifications)
- android
- macOS (support is preliminary; some code-signing-related issues)
- iOS (support is preliminary; some code-signing-related issues)

## Use

See the package documentation of `./notification_manager.go` for usage information.
