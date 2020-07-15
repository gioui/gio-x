package main

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"

	"git.sr.ht/~whereswaldon/materials"
)

var MenuIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.NavigationMenu)
	return icon
}()

var HomeIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ActionHome)
	return icon
}()

var SettingsIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ActionSettings)
	return icon
}()

var OtherIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ActionHelp)
	return icon
}()

var HeartIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ActionFavorite)
	return icon
}()

var PlusIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ContentAdd)
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
			Icon: HomeIcon,
		},
		{
			Name: "Settings",
			Tag:  "settings",
			Icon: SettingsIcon,
		},
		{
			Name: "Elsewhere",
			Tag:  "elsewhere",
			Icon: OtherIcon,
		},
	} {
		nav.AddNavItem(item)
	}
	bar := materials.AppBar{
		Theme:          th,
		NavigationIcon: MenuIcon,
		Title:          "Title",
	}
	dests := map[interface{}]func(layout.Context) layout.Dimensions{
		"home": func(gtx layout.Context) layout.Dimensions {
			bar.Title = "Home"
			gtx.Constraints.Min.Y = 0
			return layout.Flex{
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.H3(th, "Home").Layout(gtx)
				}),
			)
		},
		"settings": func(gtx layout.Context) layout.Dimensions {
			bar.Title = "Settings"
			gtx.Constraints.Min.Y = 0
			return layout.Flex{
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.H3(th, "Settings").Layout(gtx)
				}),
			)
		},
		"elsewhere": func(gtx layout.Context) layout.Dimensions {
			bar.Title = "Elsewhere"
			gtx.Constraints.Min.Y = 0
			return layout.Flex{
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.H3(th, "Elsewhere").Layout(gtx)
				}),
			)
		},
	}
	var heartBtn, plusBtn widget.Clickable
	bar.SetActions([]materials.AppBarAction{
		materials.AppBarAction{
			Name:  "Favorite",
			Icon:  HeartIcon,
			State: &heartBtn,
		},
		materials.AppBarAction{
			Name:  "Create",
			Icon:  PlusIcon,
			State: &plusBtn,
		},
	})
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			if bar.NavigationClicked() {
				nav.ToggleVisibility(gtx.Now)
			}
			layout.Inset{
				Top:    e.Insets.Top,
				Bottom: e.Insets.Bottom,
				Left:   e.Insets.Left,
				Right:  e.Insets.Right,
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return bar.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return dests[nav.CurrentNavDestiation()](gtx)
						})
					}),
				)
				nav.Layout(gtx)
				return layout.Dimensions{Size: gtx.Constraints.Max}
			})
			e.Frame(gtx.Ops)
		}
	}
}
