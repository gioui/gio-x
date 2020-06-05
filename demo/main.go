package main

import (
	"image/color"
	"log"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"git.sr.ht/~whereswaldon/colorpicker"
)

func main() {
	go func() {
		w := app.NewWindow(app.Size(unit.Dp(200), unit.Dp(100)))
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
	}()
	app.Main()
}

func loop(w *app.Window) error {
	gofont.Register()
	th := material.NewTheme()
	picker := colorpicker.State{}
	picker.SetColor(color.RGBA{R: 255, G: 128, B: 75, A: 255})
	var ops op.Ops
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e.Queue, e.Config, e.Size)
			colorpicker.PickerStyle{Label: "Example Color", Theme: th, State: &picker}.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}
