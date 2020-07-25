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

Features:
- Animated drawer open/close.
- Swipe or touch scrim to close the drawer.
- Navigation items respond to hovering.
- Navigation selection is animated.
- Navigation item icons are optional.

Known issues:

- API targets a fairly static and simplistic menu. Sub-sections with dividers are not yet supported. An API-driven way to traverse the current menu options is also not yet supported. Contributions welcome!

### App Bar (Top)

The App Bar [specified here](https://material.io/components/app-bars-top) is mostly implemented by the type
`AppBar`. It looks like this:

Normal state:

![modal navigation drawer example screenshot](https://git.sr.ht/~whereswaldon/materials/blob/master/img/app-bar-top.png)

Contextual state:

![modal navigation drawer example screenshot](https://git.sr.ht/~whereswaldon/materials/blob/master/img/app-bar-top-contextual.png)

Features:
 - Action buttons and overflow menu contents can be changed easily.
 - Overflow button disappears when no items overflow.
 - Overflow menu can be dismissed by touching the scrim outside of it.
 - Action items disapper into overflow when screen is too narrow to fit them. This is animated.
 - Navigation button icon is customizable, and the button is not drawn if no icon is provided.
 - Contextual app bar can be triggered and dismissed programatically.

Known Issues:
 - Compact and prominent App Bars are not yet implemented.
 - Cannot currently be used as a bottom app bar, though this would not be a terribly
   difficult addition (patches welcome).

### Example

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
