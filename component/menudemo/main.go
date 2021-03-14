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
	"gioui.org/widget/material"
	"gioui.org/x/component"

	"gioui.org/font/gofont"
)

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
	var ops op.Ops
	inset := layout.UniformInset(unit.Dp(8))
	menu := component.MenuState{
		Options: []func(gtx layout.Context) layout.Dimensions{
			func(gtx layout.Context) layout.Dimensions {
				return inset.Layout(gtx, func(gtx C) D {
					return material.Body1(th, "Foo").Layout(gtx)
				})
			},
			func(gtx layout.Context) layout.Dimensions {
				return inset.Layout(gtx, func(gtx C) D {
					return material.Body1(th, "Bar").Layout(gtx)
				})
			},
			func(gtx layout.Context) layout.Dimensions {
				return inset.Layout(gtx, func(gtx C) D {
					return material.Body1(th, "Baz").Layout(gtx)
				})
			},
		},
	}
	var shadows layout.List
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			paint.Fill(gtx.Ops, color.NRGBA{R: 200, G: 200, B: 200, A: 255})
			layout.Center.Layout(gtx, func(gtx C) D {
				return component.Menu(th, &menu).Layout(gtx)
			})
			layout.N.Layout(gtx, func(gtx C) D {
				return shadows.Layout(gtx, 30, func(gtx C, index int) D {
					return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx C) D {
						gtx.Constraints.Min = image.Point{X: gtx.Px(unit.Dp(30)), Y: gtx.Px(unit.Dp(30))}
						return component.Shadow(unit.Dp(8), unit.Dp(float32(index))).Layout(gtx)
					})
				})
			})
			e.Frame(gtx.Ops)
		}
	}
}
