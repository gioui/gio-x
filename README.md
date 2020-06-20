## niotify

Cross platform notifications for [Gio](https://gioui.org) applications.

## Status

This repo is experimental, and does not have a stable interface. Currently niotify
only supports the following OSes:

- linux (x11/wayland doesn't matter so long as dbus is used for notifications)
- android

Contributions of support for other platforms are welcome! Send inquiries and patches
to [my public inbox](https://lists.sr.ht/~whereswaldon/public-inbox) for now.

## Use

niotify requires a `replace` directive in your `go.mod` to add features to an underlying
JNI library. This should be temporary.

For now, add:

```
replace git.wow.st/gmp/jni => git.wow.st/whereswaldon/jni v0.0.0-20200620152723-b380472956a0
```

See the package documentation of `./notification_manager.go` for usage information.

## Name

go => gio

notify => niotify
