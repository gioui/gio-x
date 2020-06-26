# haptic

Haptic feedback for Gio applications on Android

## Status

Experimental, but working. API is not stable, so use go modules to lock
to a particular version.

I'd like to also support iOS in the future, but currently only android
support is implemented.

On non-supported OSes, the API is the same but does nothing. This makes
it easier to write cross-platform code that depends on haptic.

## Usage

Until [this PR is merged](https://git.wow.st/gmp/jni/pulls/2), you'll need to add this to your `go.mod`:

```
replace git.wow.st/gmp/jni => git.wow.st/whereswaldon/jni v0.0.0-20200626194017-b74a17279b1f
```

Create a `Buzzer`:

```go
buzzer := haptic.NewBuzzer(window)

// check for problems:
select {
    case err := <- buzzer.Errors():
    // handle
    case event := <-window.Events():
    // normal gio stuff
}
```

Send a haptic buzz:

```go
if !buzzer.Buzz() {
    // Couldn't trigger a buzz without blocking. Handle however you like.
    // I recommend just retrying soon (maybe next frame).
}
```

When you're done:

```go
buzzer.Shutdown()
```

## Why is the API weird?

We can't interact with the JVM from the same OS thread that runs your Gio
event processing. Rather than accidentally allow you to deadlock by calling
these methods the wrong way, they're written to be safe to invoke from your
normal Gio layout code without deadlock. This means that all of the work
needs to occur on other goroutines.

## Contributing

Send questions, comments, and patches to [my public inbox](https://lists.sr.ht/~whereswaldon/public-inbox).

## License

Dual MIT/Unlicense
