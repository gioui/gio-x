// SPDX-License-Identifier: Unlicense OR MIT
package main

// A simple Gio program. See https://gioui.org for more information.

import (
	"log"
	"os"
	"strconv"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/scroll"

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

const oneThird float32 = 1.0 / 3.0

type (
	C = layout.Context
	D = layout.Dimensions
)

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	var (
		increaseBtn, decreaseBtn widget.Clickable
		horizontalSwitch         widget.Bool
		list                     scroll.ListState
		ops                      op.Ops
		length                   int = 32
		inset                        = layout.UniformInset(unit.Dp(4))
	)
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			if increaseBtn.Clicked() {
				length *= 2
			}
			if decreaseBtn.Clicked() {
				length /= 2
				if length < 1 {
					length = 1
				}
			}
			if !horizontalSwitch.Value {
				list.Axis = layout.Vertical
			} else {
				list.Axis = layout.Horizontal
			}

			layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
						layout.Flexed(oneThird, func(gtx C) D {
							return inset.Layout(gtx, material.Button(th, &increaseBtn, "Double list length").Layout)
						}),
						layout.Flexed(oneThird, func(gtx C) D {
							return layout.Center.Layout(gtx, func(gtx C) D {
								return material.Body1(th, "Current List Length: "+strconv.Itoa(length)).Layout(gtx)
							})
						}),
						layout.Flexed(oneThird, func(gtx C) D {
							return inset.Layout(gtx, material.Button(th, &decreaseBtn, "Halve list length").Layout)
						}),
						layout.Rigid(func(gtx C) D {
							return inset.Layout(gtx, func(gtx C) D {
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										return material.Body1(th, "Horizontal").Layout(gtx)
									}),
									layout.Rigid(func(gtx C) D {
										return inset.Layout(gtx, material.Switch(th, &horizontalSwitch).Layout)
									}),
								)
							})
						}),
					)
				}),
				layout.Flexed(1, func(gtx C) D {
					return scroll.List(th, &list).Layout(gtx, length, func(gtx C, index int) D {
						return layout.Center.Layout(gtx, material.H3(th, "List item #"+strconv.Itoa(index)).Layout)
					})
				}),
			)
			e.Frame(gtx.Ops)
		}
	}
}
