## niotify

[![GoDoc](https://godoc.org/git.sr.ht/~whereswaldon/niotify?status.svg)](https://godoc.org/git.sr.ht/~whereswaldon/niotify)

Cross platform notifications for [Gio](https://gioui.org) applications.

## Status

This repo is experimental, and does not have a stable interface. Currently niotify
only supports the following OSes:

- linux (x11/wayland doesn't matter so long as dbus is used for notifications)
- android
- macOS (support is preliminary; some code-signing-related issues)
- iOS (support is preliminary; some code-signing-related issues)

Contributions of support for other platforms are welcome! Send inquiries and patches
to [my public inbox](https://lists.sr.ht/~whereswaldon/public-inbox) for now.

### macOS/iOS Support

We've had mixed success with this platform, as sending notifications requires your
"app bundle" to authenticate with the system. This is only possible with a signed
app bundle. To see a way to get a signed app bundle for local devleopment see
`./example/hello_macos.go`, which uses `go generate` to create one. To run the
example on macOS, do the following:

```sh
cd ./example/
go generate -x
./example.app/Contents/MacOS/example
```

Depending on your settings, you may need to reveal the notification center in order
to actually see the notifications.

The above uses ad-hoc code signing, which doesn't work well for redistribution.
However, it *should* be possible to sign the bundle with a real apple signing
key to eliminate this problem.

Please report issues with this to [this mailing list](https://lists.sr.ht/~whereswaldon/public-inbox)!

## Use

See the package documentation of `./notification_manager.go` for usage information.

## Name

go => gio

notify => niotify
