## gio-x

[![Go Reference](https://pkg.go.dev/badge/gioui.org/x.svg)](https://pkg.go.dev/gioui.org/x)

This repository hosts `gioui.org/x`. Two kinds of package exist in this namespace. Some are extensions that will eventually be merged into `gioui.org`'s core repository once their APIs stabilize and their value to the community is proven. The rest are extensions to Gio that are not likely to be needed by every application and require new dependencies. These will likely never be merged to the core repository, but will be maintained here.

This table describes the current status of each package in `gioui.org/x`:

| Name        | Purpose                                     | Intended for core? | Non-core dependencies? | API Stability |
| ----------- | ------------------------------------------- | ------------------ | ---------------------- | ------------- |
| colorpicker | Widgets for choosing colors                 | uncertain          | no                     | unstable      |
| component   | Material.io components                      | uncertain          | no                     | unstable      |
| eventx      | Event management tools                      | yes                | no                     | unstable      |
| haptic      | Haptic feedback for mobile devices          | no                 | yes                    | unstable      |
| notify      | Background notifications                    | no                 | yes                    | unstable      |
| outlay      | Extra layouts                               | yes                | no                     | unstable      |
| pref        | Query user/device preferences               | no                 | yes                    | unstable      |
| profiling   | Gio render performance recording tools      | uncertain          | no                     | unstable      |
| richtext    | Printing text objects with different styles | uncertain          | no                     | unstable      |

## Contributing

Report bugs on the [gio issue tracker](https://todo.sr.ht/~eliasnaur/gio) with the prefix `gio-x:` in your issue title.

Ask questions on the [gio discussion mailing list](https://lists.sr.ht/~eliasnaur/gio).

Send patches on the [gio patches mailing list](https://lists.sr.ht/~eliasnaur/gio-patches) with the subject line prefix `[PATCH gio-x]`

All patches should be Signed-off to indicate conformance with the [LICENSE](https://git.sr.ht/~whereswaldon/gio-x/tree/main/LICENSE) of this repo.

## Maintainers

This repository is primarily maintained by Chris Waldon.

## License

Dual MIT/Unlicense; same as Gio

## Support

If gio provides value to you, please consider supporting one or more of its developers and maintainers:

Elias Naur:
https://github.com/sponsors/eliasnaur

Chris Waldon:
https://github.com/sponsors/whereswaldon
https://liberapay.com/whereswaldon
