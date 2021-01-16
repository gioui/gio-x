## gio-x

[![Go Reference](https://pkg.go.dev/badge/gioui.org/x.svg)](https://pkg.go.dev/gioui.org/x)

This repository hosts `gioui.org/x`. Two kinds of package exist in this namespace. Some are extensions that will eventually be merged into `gioui.org`'s core repository once their APIs stabilize and their value to the community is proven. The rest are extensions to Gio that are not likely to be needed by every application and require new dependencies. These will likely never be merged to the core repository, but will be maintained here.

This table describes the current status of each package in `gioui.org/x`:

| Name        | Intended for core? | Non-core dependencies? | API Stability |
| ---         | ---                | ---                    | ---           |
| colorpicker | uncertain          | no                     | unstable      |
| component   | uncertain          | no                     | unstable      |
| eventx      | yes                | no                     | unstable      |
| haptic      | no                 | yes                    | unstable      |
| notify      | no                 | yes                    | unstable      |
| outlay      | yes                | no                     | unstable      |
| profiling   | uncertain          | no                     | unstable      |
| scroll      | yes                | no                     | unstable      |

## Contributing

Report bugs on the [gio issue tracker]() with the prefix `gio-x:` in your issue title.

Ask questions on the [gio discussion mailing list]().

Send patches on the [gio patches mailing list]() with the subject line prefix `[PATCH gio-x]`

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
