// SPDX-License-Identifier: Unlicense OR MIT

package main

// A simple Gio program. See https://gioui.org for more information.

import (
	"image/color"
	"log"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget/material"

	"gioui.org/font/gofont"
	"git.sr.ht/~whereswaldon/niotify/android"
)

//go:generate javac -target 1.8 -source 1.8 -bootclasspath $ANDROID_HOME/platforms/android-29/android.jar ../android/NotificationHelper.java
//go:generate jar cf NotificationHelper.jar ../android/NotificationHelper.class
//go:generate rm ../android/NotificationHelper.class

func main() {
	go func() {
		w := app.NewWindow()
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		channel, err := android.NewChannel("hello", "description")
		if err != nil {
			log.Printf("channel creation failed: %v", err)
		}
		log.Println(channel)
	}()
	app.Main()
}

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	var ops op.Ops
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			l := material.H1(th, "Hello, Gio")
			maroon := color.RGBA{127, 0, 0, 255}
			l.Color = maroon
			l.Alignment = text.Middle
			l.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}
