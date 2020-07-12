package main

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"git.sr.ht/~whereswaldon/materials"
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

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	var ops op.Ops
	nav := materials.ModalNavDrawer{
		Theme:    th,
		Title:    "Navigation Drawer",
		Subtitle: "This is an example.",
	}

	for _, item := range []materials.NavItem{
		{
			Name: "Home",
			Tag:  "home",
		},
		{
			Name: "Settings",
			Tag:  "settings",
		},
		{
			Name: "Elsewhere",
			Tag:  "elsewhere",
		},
	} {
		nav.AddNavItem(item)
	}
	var btn widget.Clickable
	dests := map[interface{}]func(layout.Context) layout.Dimensions{
		nil: func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, material.Button(th, &btn, "nav").Layout)
		},
	}
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			if btn.Clicked() {
				nav.ToggleVisibility(gtx.Now)
			}
			layout.Inset{
				Top:    e.Insets.Top,
				Bottom: e.Insets.Bottom,
				Left:   e.Insets.Left,
				Right:  e.Insets.Right,
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				dests[nav.CurrentNavDestiation()](gtx)
				nav.Layout(gtx)
				return layout.Dimensions{Size: gtx.Constraints.Max}
			})
			e.Frame(gtx.Ops)
		}
	}
}
