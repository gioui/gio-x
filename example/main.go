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

type Page struct {
	layout func(layout.Context) layout.Dimensions
	materials.NavItem
	Actions  []materials.AppBarAction
	Overflow []materials.OverflowAction
}

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	var ops op.Ops
	nav := materials.NewModalNav(th, "Navigation Drawer", "This is an example.")
	bar := materials.AppBar{Theme: th}
	bar.NavigationIcon = MenuIcon

	var (
		heartBtn, plusBtn, exampleOverflowState widget.Clickable
		red, green, blue                        widget.Clickable
		contextBtn                              widget.Clickable
	)

	pages := []Page{
		Page{
			NavItem: materials.NavItem{
				Name: "Home",
				Icon: HomeIcon,
			},
			layout: func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Alignment: layout.Middle,
					Axis:      layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.H3(th, "Home").Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Button(th, &contextBtn, "Context").Layout(gtx)
					}),
				)
			},
			Actions: []materials.AppBarAction{
				materials.AppBarAction{
					OverflowAction: materials.OverflowAction{
						Name:  "Favorite",
						State: &heartBtn,
					},
					Icon: HeartIcon,
				},
				materials.AppBarAction{
					OverflowAction: materials.OverflowAction{
						Name:  "Create",
						State: &plusBtn,
					},
					Icon: PlusIcon,
				},
			},
			Overflow: []materials.OverflowAction{
				{
					Name:  "Example",
					State: &exampleOverflowState,
				},
				{
					Name:  "Red",
					State: &red,
				},
				{
					Name:  "Green",
					State: &green,
				},
				{
					Name:  "Blue",
					State: &blue,
				},
			},
		},
		Page{
			NavItem: materials.NavItem{
				Name: "Settings",
				Icon: SettingsIcon,
			},
			layout: func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.H3(th, "Settings").Layout(gtx)
					}),
				)
			},
			Actions: []materials.AppBarAction{
				materials.AppBarAction{
					OverflowAction: materials.OverflowAction{
						Name:  "Create",
						State: &plusBtn,
					},
					Icon: PlusIcon,
				},
			},
		},
		Page{
			NavItem: materials.NavItem{
				Name: "Elsewhere",
				Icon: OtherIcon,
			},
			layout: func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.H3(th, "Elsewhere").Layout(gtx)
					}),
				)
			},
			Actions: []materials.AppBarAction{
				materials.AppBarAction{
					OverflowAction: materials.OverflowAction{
						Name:  "Favorite",
						State: &heartBtn,
					},
					Icon: HeartIcon,
				},
			},
		},
	}

	for i, page := range pages {
		page.NavItem.Tag = i
		nav.AddNavItem(page.NavItem)
	}
	{
		page := pages[nav.CurrentNavDestination().(int)]
		bar.Title = page.Name
		bar.SetActions(page.Actions, page.Overflow)
	}
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			if bar.NavigationClicked(gtx) {
				nav.ToggleVisibility(gtx.Now)
			}
			if green.Clicked() || red.Clicked() || blue.Clicked() || exampleOverflowState.Clicked() {
				bar.CloseOverflowMenu(gtx.Now)
			}
			if contextBtn.Clicked() {
				bar.SetContextualActions(
					[]materials.AppBarAction{
						{
							Icon: HomeIcon,
							OverflowAction: materials.OverflowAction{
								Name:  "House",
								State: &red,
							},
						},
					},
					[]materials.OverflowAction{
						{
							Name:  "foo",
							State: &blue,
						},
						{
							Name:  "bar",
							State: &green,
						},
					},
				)
				bar.ToggleContextual(gtx.Now, "Contextual Title")
			}
			if nav.NavDestinationChanged() {
				page := pages[nav.CurrentNavDestination().(int)]
				bar.Title = page.Name
				bar.SetActions(page.Actions, page.Overflow)
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
							return pages[nav.CurrentNavDestination().(int)].layout(gtx)
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
