package main

import (
	"image/color"
	"log"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"git.sr.ht/~whereswaldon/colorpicker"
)

func main() {
	go func() {
		w := app.NewWindow(app.Size(unit.Dp(200), unit.Dp(300)))
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
	}()
	app.Main()
}

type (
	C = layout.Context
	D = layout.Dimensions
)

var white = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	background := white
	picker := colorpicker.State{}
	picker.SetColor(color.RGBA{R: 255, G: 128, B: 75, A: 255})
	muxState := colorpicker.NewMuxState(
		[]colorpicker.MuxOption{
			{
				Label: "white",
				Value: &white,
			},
			{
				Label: "primary",
				Value: &th.Color.Primary,
			},
			{
				Label: "hint",
				Value: &th.Color.Hint,
			},
			{
				Label: "text",
				Value: &th.Color.Text,
			},
		}...)
	var ops op.Ops
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			if muxState.Changed() {
				background = *muxState.Color()
				log.Printf("mux changed")
			}
			if picker.Changed() {
				th.Color.Primary = picker.Color()
				log.Printf("picker changed")
			}
			layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return colorpicker.PickerStyle{Label: "Primary", Theme: th, State: &picker}.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return colorpicker.Mux(th, &muxState, "Display Right:").Layout(gtx)
						}),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							defer op.Push(gtx.Ops).Pop()
							paint.ColorOp{Color: background}.Add(gtx.Ops)
							paint.PaintOp{Rect: f32.Rect(0, 0, float32(gtx.Constraints.Max.X), float32(gtx.Constraints.Max.Y))}.Add(gtx.Ops)
							return D{}
						}),
					)
				}),
			)
			e.Frame(gtx.Ops)
		}
	}
}
