# pref [![Go Reference](https://pkg.go.dev/badge/gioui.org/x/pref.svg)](https://pkg.go.dev/gioui.org/x/pref)

-------------

Get the user preferences for your Gio app.

## What can it be used for?

The [Theme](https://pkg.go.dev/gioui.org/x/pref/theme) package provides `IsDarkMode`, which can be used to
change your palette in order to honor the user's preferences.

    // Check the preference:
    isDark, _ := theme.IsDarkMode() 
    
    // Change the Palette based on the preference:
	var palette material.Palette
	if isDark {
		palette.Bg = color.NRGBA{A: 255}                         // Black Background
		palette.Fg = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // White Foreground
	} else {
		palette.Bg = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // White Background 
		palette.Fg = color.NRGBA{A: 255}                         // Black Foreground
	}

The [Locale](https://pkg.go.dev/gioui.org/x/pref/locale) makes possible to match the user language preference, that
is important for multi-language apps. So, let your app speak the user's native language.

	// Your dictionary (in that case using x/text/catalog):
	cat := catalog.NewBuilder()
	cat.SetString(language.English, "Hello World", "Hello World")
	cat.SetString(language.Portuguese, "Hello World", "Olá Mundo")
	cat.SetString(language.Spanish, "Hello World", "Hola Mundo")
	cat.SetString(language.Czech, "Hello World", "Ahoj světe")
	cat.SetString(language.French, "Hello World", "Bonjour le monde")

	// Get the user preferences:
	userLanguage, _ := locale.Language()

	// Get the best match based on the preferred language:
	userLanguage, _, confidence := cat.Matcher().Match(userLanguage)
	if confidence <= language.Low {
		userLanguage = language.English // Switch to the default language, due to low confidence.
	}

	// Creates the printer with the user language:
	printer := message.NewPrinter(userLanguage, message.Catalog(cat))

	// Display the text based on the language:
	widget.Label{}.Layout(gtx,
		yourTheme.Shaper,
		text.Font{},
		unit.Dp(12),
		printer.Sprintf("Hello World"),
	)

## Status

Most of the features is supported across Android 6+, JS and Windows 10. It will return ErrAvailableAPI for any other
platform that isn't supported.

| Package      | OS | 
| ----------- | ----------- | 
| [Locale](https://pkg.go.dev/gioui.org/x/pref/locale)     |  Android 6+  <br> JS <br> Linux <br> Windows Vista+ | 
| [Theme](https://pkg.go.dev/gioui.org/x/pref/theme)   |  Android 4+ <br> JS <br> Windows 10+ | 
| [Battery](https://pkg.go.dev/gioui.org/x/pref/battery)   |  Android 6+ <br> JS (Chrome) <br> Windows Vista+| 
