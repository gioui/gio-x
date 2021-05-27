# Color Input

A Gio library for color input widgets

## API

This library lets one compose a color input widget from a number of wigets that implement the `ColorInput` interface.

* `Picker`: Choose a primary color and one of it tones.
* `NewAlphaSlider`: Set Alpha value of Color.
* `ColorSelection`: A dropdown that wraps any other color widget.
* `HexEditor`, `RgbEditor`, `HsvEditor`: Edit textual representations of a color.
* `Mux`: A special widgets that wraps multiple color widgets into on a horizontal view.
* `Toggle`: Toggle between different color widget, useful for the different textual widgets. 


Or implement the `ColorInput` to create your own widget. 

```go
type ColorInput interface {
	Layout(gtx layout.Context) layout.Dimensions
	Changed() bool
	SetColor(col color.NRGBA)
	Color() color.NRGBA
}
```

## Example

A Simple dropdown with a colorpicker, a alpha slider and a toggle between 3 textual color inputs.

```go
package main

import (
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/colorpicker"
	"image/color"
)

func main() {
	go func() {
		w := app.NewWindow(app.Size(unit.Dp(800), unit.Dp(700)))
		th := material.NewTheme(gofont.Collection())
		colorField := colorpicker.NewColorSelection(th, layout.SW,
			colorpicker.NewPicker(th),
			colorpicker.NewAlphaSlider(),
			colorpicker.NewToggle(&widget.Clickable{},
				colorpicker.NewHexEditor(th),
				colorpicker.NewRgbEditor(th),
				colorpicker.NewHsvEditor(th)))
		colorField.SetColor(color.NRGBA{G: 255, A: 255})
		var ops op.Ops
		for {
			e := <-w.Events()
			switch e := e.(type) {
			case system.FrameEvent:
				colorField.Changed()
				gtx := layout.NewContext(&ops, e)
				layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(material.H2(th, "Color Inputs").Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.UniformInset(unit.Dp(4)).Layout(gtx,
							func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
									layout.Rigid(material.Label(th, unit.Sp(16), " Color: ").Layout),
									layout.Rigid(colorField.Layout))
							})
					}))
				e.Frame(gtx.Ops)
			}
		}
	}()
	app.Main()
}
```

## TODO

* The tone of a color is not what is shown under the control of the picker.
  * Maybe we need to use the HSL colorspace instead of HSV.
  * Maybe the gradient can be drawn differently.
* Make the control of the tone and hue picker align to the middle of a mouse event.
* Fix issue with size of hue slider when the width in pixels is not a multiple of 6.
* Fix bug where changing text does not update the rainbow slider
* Make rainbow slider and tone window into their own component.
* Use the generic `color.Color` interface instead of the `color.NRGBA` struct to prevent subtle issues when going from HSV to RGB.
* Have `HSVColor` implement `color.Color`.

## Author

Werner Laurensse
