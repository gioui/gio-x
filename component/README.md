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

### Navigation Drawer (static and modal)

The navigation drawer [specified here](https://material.io/components/navigation-drawer) is mostly implemented by the type
`NavDrawer`, and the modal variant can be created with a `ModalNavDrawer`. The modal variant looks like this:

![modal navigation drawer example screenshot](https://git.sr.ht/~whereswaldon/materials/blob/main/img/modal-nav.png)

Features:
- Animated drawer open/close.
- Navigation items respond to hovering.
- Navigation selection is animated.
- Navigation item icons are optional.
- Content can be anchored to the bottom of the drawer for pairing with a bottom app bar.

Modal features:
- Swipe or touch scrim to close the drawer.

Known issues:

- API targets a fairly static and simplistic menu. Sub-sections with dividers are not yet supported. An API-driven way to traverse the current menu options is also not yet supported. Contributions welcome!

### App Bar (Top and Bottom)

The App Bar [specified here](https://material.io/components/app-bars-top) is mostly implemented by the type
`AppBar`. It looks like this:

Normal state:

![modal navigation drawer example screenshot](https://git.sr.ht/~whereswaldon/materials/blob/main/img/app-bar-top.png)

Contextual state:

![modal navigation drawer example screenshot](https://git.sr.ht/~whereswaldon/materials/blob/main/img/app-bar-top-contextual.png)

Features:
 - Action buttons and overflow menu contents can be changed easily.
 - Overflow button disappears when no items overflow.
 - Overflow menu can be dismissed by touching the scrim outside of it.
 - Action items disapper into overflow when screen is too narrow to fit them. This is animated.
 - Navigation button icon is customizable, and the button is not drawn if no icon is provided.
 - Contextual app bar can be triggered and dismissed programatically.
 - Bar supports use as a top and bottom app bar (animates the overflow menu in the proper direction).

Known Issues:
 - Compact and prominent App Bars are not yet implemented.

### Side sheet (static and modal)

Side sheets ([specified here](https://material.io/components/sheets-side)) are implemented by the `Sheet` and `ModalSheet` types.

Features:
- Animated appear/disappear

Modal features:
- Swipe to close
- Touch scrim to close

Known Issues:
- Only sheets anchored on the left are currently supported (contributions welcome!)

### Text Fields

Text Fields ([specified here](https://material.io/components/text-fields)) are implemented by the `TextField` type.

Features:
- Animated label transition when selected
- Responds to hover events
- Exposes underlying gio editor

Known Issues:
- Icons, hint text, error text, prefix/suffix, and other features are not yet implemented.

## Example

Want to see it?

```
git clone https://git.sr.ht/~whereswaldon/materials
cd materials
go run ./example
```

You can also see the demo using a bottom bar with:
```
go run ./example/ -bottom-bar
```

## Contributing

Contributions to this collection are welcome! All contributions should adhere to
the [material design specifications](https://material.io/components) for the components that they implement.

You can send bug reports, feature requests, questions, and patches to [my public inbox](https://lists.sr.ht/~whereswaldon/public-inbox).

All patches should be Signed-off to indicate conformance with the [LICENSE](https://git.sr.ht/~whereswaldon/materials/tree/main/LICENSE) of this repo.

## License

Dual MIT/Unlicense; same as Gio
