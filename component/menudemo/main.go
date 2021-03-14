// SPDX-License-Identifier: Unlicense OR MIT

package main

// A simple Gio program. See https://gioui.org for more information.

import (
	"image"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
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

type RightClickArea struct {
	position f32.Point
	dims     D
	active   bool
}

func (r *RightClickArea) Layout(gtx C, w layout.Widget) D {
	pointer.Rect(image.Rectangle{Max: gtx.Constraints.Min}).Add(gtx.Ops)
	pointer.PassOp{Pass: true}.Add(gtx.Ops)
	pointer.InputOp{
		Tag:   r,
		Grab:  false,
		Types: pointer.Press | pointer.Release,
	}.Add(gtx.Ops)
	for _, e := range gtx.Events(r) {
		e, ok := e.(pointer.Event)
		if !ok {
			continue
		}
		if r.active {
			// Check whether we should dismiss menu.
			if e.Buttons.Contain(pointer.ButtonPrimary) {
				clickPos := e.Position.Sub(r.position)
				if !clickPos.In(f32.Rectangle{Max: layout.FPt(r.dims.Size)}) {
					r.Dismiss()
				}
			}
		}
		if e.Buttons.Contain(pointer.ButtonSecondary) {
			r.active = true
			r.position = e.Position
		}
	}
	dims := D{Size: gtx.Constraints.Min}

	if !r.active {
		return dims
	}

	for _, e := range gtx.Events(&r.active) {
		e, ok := e.(pointer.Event)
		if !ok {
			continue
		}
		if e.Type == pointer.Release {
			r.Dismiss()
		}
	}

	defer op.Save(gtx.Ops).Load()
	macro := op.Record(gtx.Ops)
	r.dims = w(gtx)
	call := macro.Stop()

	if int(r.position.X)+r.dims.Size.X > gtx.Constraints.Max.X {
		r.position.X = float32(gtx.Constraints.Max.X - r.dims.Size.X)
	}
	if int(r.position.Y)+r.dims.Size.Y > gtx.Constraints.Max.Y {
		r.position.Y = float32(gtx.Constraints.Max.Y - r.dims.Size.Y)
	}
	macro2 := op.Record(gtx.Ops)
	op.Offset(r.position).Add(gtx.Ops)
	call.Add(gtx.Ops)
	pointer.PassOp{Pass: true}.Add(gtx.Ops)
	pointer.Rect(image.Rectangle{Max: r.dims.Size}).Add(gtx.Ops)
	pointer.InputOp{
		Tag:   &r.active,
		Grab:  false,
		Types: pointer.Release,
	}.Add(gtx.Ops)
	call2 := macro2.Stop()
	op.Defer(gtx.Ops, call2)
	return dims
}

func (r *RightClickArea) Dismiss() {
	r.active = false
}

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	var (
		ops     op.Ops
		a, b, c widget.Clickable
	)
	inset := layout.UniformInset(unit.Dp(8))
	menu := component.MenuState{
		Options: []func(gtx layout.Context) layout.Dimensions{
			func(gtx layout.Context) layout.Dimensions {
				return material.Clickable(gtx, &a, func(gtx C) D {
					return inset.Layout(gtx, func(gtx C) D {
						return material.Body1(th, "Foo").Layout(gtx)
					})
				})
			},
			func(gtx layout.Context) layout.Dimensions {
				return material.Clickable(gtx, &b, func(gtx C) D {
					return inset.Layout(gtx, func(gtx C) D {
						return material.Body1(th, "Bar").Layout(gtx)
					})
				})
			},
			func(gtx layout.Context) layout.Dimensions {
				return material.Clickable(gtx, &c, func(gtx C) D {
					return inset.Layout(gtx, func(gtx C) D {
						return material.Body1(th, "Baz").Layout(gtx)
					})
				})
			},
		},
	}
	var shadows layout.List
	var area RightClickArea
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
			area.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min = image.Point{}
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
