# haptic

[![Go Reference](https://pkg.go.dev/badge/gioui.org/x/haptic.svg)](https://pkg.go.dev/gioui.org/x/haptic)

Haptic feedback for Gio applications

## Status

Experimental, but working. API is not stable, so use go modules to lock
to a particular version.

On non-supported OSes, the API is the same but does nothing. This makes
it easier to write cross-platform code that depends on haptic.

## Why is the API weird?

We can't interact with the JVM from the same OS thread that runs your Gio
event processing. Rather than accidentally allow you to deadlock by calling
these methods the wrong way, they're written to be safe to invoke from your
normal Gio layout code without deadlock. This means that all of the work
needs to occur on other goroutines.
