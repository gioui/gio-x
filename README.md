# materials

This package provides various material design components for [gio](https://gioui.org).

## State

This package has no stable API, and should always be locked to a particular commit with
go modules.

The included components attempt to conform to the [material design specifications](https://material.io/components/)
whenever possible, but they may not support unusual style tweaks or especially exotic
configurations.

## Implemented Components

The list of currently-Implemented components follows:

### Modal Navigation Drawer

The modal navigation drawer [specified here](https://material.io/components/navigation-drawer#modal-drawer) is mostly implemented by the type
`ModalNavDrawer`. It looks like this:

![modal navigation drawer example screenshot](https://git.sr.ht/~whereswaldon/materials/blob/master/img/modal-nav.png)

Known issues:

- Icon support is not yet implemented.

Want to see it?

```
git clone https://git.sr.ht/~whereswaldon/materials
cd materials
go run ./example
```

## Contributing

Contributions to this collection are welcome! All contributions should adhere to
the [material design specifications](https://material.io/components) for the components that they implement.

You can send bug reports, feature requests, questions, and patches to [my public inbox](https://lists.sr.ht/~whereswaldon/public-inbox).

All patches should be Signed-off to indicate conformance with the [LICENSE](https://git.sr.ht/~whereswaldon/materials/tree/master/LICENSE) of this repo.

## License

Dual MIT/Unlicense; same as Gio
