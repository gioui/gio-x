# eventx

[![Go Reference](https://pkg.go.dev/badge/gioui.org/x/eventx.svg)](https://pkg.go.dev/gioui.org/x/eventx)

This package provides types to manage [Gio](https://gioui.org) events.

## State

This package has no stable API, and should always be locked to a particular commit with
go modules.

## Features

### Spy

A spy is an event processing tool that wraps a Gio event queue. Whenever its `Events`
method is invoked, it makes the same `Events` call on the wrapped queue, but makes
a copy of all events that it receives from the underlying queue. This copy can be
accessed using the `AllEvents()` method.

The primary use-case for this type is observing the raw event stream for a UI
component that consumes some-but-not-all of the relevant events. For instance,
you can extend the keyboard shortcuts understood by `material.Editor` by providing
a spied-upon `layout.Context` to it and then (after laying it out) checking the
events within the spy for the keystrokes of interest.

The Spy was conceived by ~eliasnaur in [this mailing list discussion](https://lists.sr.ht/~eliasnaur/gio-patches/patches/14507).

### EventGroup

This type is returned by the Spy, but can also be instantiated literally. It
functions as a simple standalone event queue tha responds to a single, specific
tag.

### CombinedQueue

CombinedQueue combines the output of two queues. This can be useful to join the
events of the "real" gio event queue with a fake one like an EventGroup.
