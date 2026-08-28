package main

import (
	"fmt"
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"gioui.org/x/gamepad"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	w := app.NewWindow()

	gamepad := gamepad.NewGamepad(w)

	t := material.NewTheme(gofont.Collection())
	ops := new(op.Ops)
	go func() {
		for evt := range w.Events() {
			gamepad.ListenEvents(evt)
			switch evt := evt.(type) {
			case system.FrameEvent:
				gtx := layout.NewContext(ops, evt)
				material.Label(t, unit.Dp(14), fmt.Sprint(gamepad.Controllers[0])).Layout(gtx)

				op.InvalidateOp{}.Add(ops)
				evt.Frame(ops)
			}

		}

	}()

	app.Main()
}
