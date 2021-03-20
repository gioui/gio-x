// SPDX-License-Identifier: Unlicense OR MIT

package main

// A simple Gio program. See https://gioui.org for more information.

import (
	"image"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"golang.org/x/exp/shiny/materialdesign/icons"

	"gioui.org/font/gofont"
)

var SettingsIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ActionSettings)
	return icon
}()

var RotationIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.Action3DRotation)
	return icon
}()

var SomethingIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ActionAccountBox)
	return icon
}()

func main() {
	go func() {
		w := app.NewWindow()
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

type (
	C = layout.Context
	D = layout.Dimensions
)

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	var (
		ops     op.Ops
		a, b, c widget.Clickable
	)
	menu := component.MenuState{
		Options: []func(gtx layout.Context) layout.Dimensions{
			func(gtx C) D {
				m := component.MenuItem(th, &a, "Foobarbaz")
				m.Icon = SettingsIcon
				m.Hint = component.MenuHintText(th, "Hint")
				return m.Layout(gtx)
			},
			func(gtx C) D {
				m := component.MenuItem(th, &b, "Something")
				m.Icon = SomethingIcon
				m.Hint = component.MenuHintText(th, "Hin")
				return m.Layout(gtx)
			},
			component.SubheadingDivider(th, "subheading").Layout,
			func(gtx C) D {
				m := component.MenuItem(th, &c, "else")
				m.Icon = RotationIcon
				m.Hint = component.MenuHintText(th, "H")
				return m.Layout(gtx)
			},
		},
	}
	var shadows layout.List
	var areas []component.ContextArea
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			paint.Fill(gtx.Ops, color.NRGBA{R: 200, G: 200, B: 200, A: 255})
			gtx.Constraints = layout.Exact(gtx.Constraints.Max)
			if a.Clicked() {
				log.Println("A")
			}
			if b.Clicked() {
				log.Println("B")
			}
			if c.Clicked() {
				log.Println("C")
			}
			layout.N.Layout(gtx, func(gtx C) D {
				return shadows.Layout(gtx, 30, func(gtx C, index int) D {
					if len(areas) < index+1 {
						areas = append(areas, component.ContextArea{})
					}
					active := areas[index].Active()
					return layout.Stack{}.Layout(gtx,
						layout.Expanded(func(gtx C) D {
							return areas[index].Layout(gtx, func(gtx C) D {
								gtx.Constraints.Min = image.Point{}
								return component.Menu(th, &menu).Layout(gtx)
							})
						}),
						layout.Stacked(func(gtx C) D {
							return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx C) D {
								gtx.Constraints.Min = image.Point{X: gtx.Px(unit.Dp(30)), Y: gtx.Px(unit.Dp(30))}
								shadow := component.Shadow(unit.Dp(8), unit.Dp(float32(index)))
								if active {
									col := color.NRGBA{R: 255, A: 0x30}
									shadow.AmbientColor = col
									shadow.PenumbraColor = col
									shadow.UmbraColor = col
								}
								return shadow.Layout(gtx)
							})
						}),
					)
				})
			})
			e.Frame(gtx.Ops)
		}
	}
}
