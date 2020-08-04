package main

import (
	"flag"
	"image/color"
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

var barOnBottom bool

func main() {
	flag.BoolVar(&barOnBottom, "bottom-bar", false, "place the app bar on the bottom of the screen instead of the top")
	flag.Parse()
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
	modal := materials.NewModal()
	th := material.NewTheme(gofont.Collection())
	var ops op.Ops
	nav := materials.NewModalNav(th, modal, "Navigation Drawer", "This is an example.")
	bar := materials.NewAppBar(th, modal)
	bar.NavigationIcon = MenuIcon
	if barOnBottom {
		bar.Anchor = materials.Bottom
		nav.Anchor = materials.Bottom
	}

	var (
		heartBtn, plusBtn, exampleOverflowState widget.Clickable
		red, green, blue                        widget.Clickable
		contextBtn                              widget.Clickable
		favorited                               bool
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
						Name: "Favorite",
						Tag:  &heartBtn,
					},
					Layout: func(gtx layout.Context, bg, fg color.RGBA) layout.Dimensions {
						btn := materials.SimpleIconButton(th, &heartBtn, HeartIcon)
						btn.Background = bg
						if favorited {
							btn.Color = color.RGBA{R: 200, A: 255}
						} else {
							btn.Color = fg
						}
						return btn.Layout(gtx)
					},
				},
				materials.SimpleIconAction(th, &plusBtn, PlusIcon,
					materials.OverflowAction{
						Name: "Create",
						Tag:  &plusBtn,
					},
				),
			},
			Overflow: []materials.OverflowAction{
				{
					Name: "Example",
					Tag:  &exampleOverflowState,
				},
				{
					Name: "Red",
					Tag:  &red,
				},
				{
					Name: "Green",
					Tag:  &green,
				},
				{
					Name: "Blue",
					Tag:  &blue,
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
				materials.SimpleIconAction(th, &plusBtn, PlusIcon,
					materials.OverflowAction{
						Name: "Create",
						Tag:  &plusBtn,
					},
				),
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
				materials.SimpleIconAction(th, &heartBtn, HeartIcon,
					materials.OverflowAction{
						Name: "Favorite",
						Tag:  &heartBtn,
					},
				),
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
			if bar.OverflowActionClicked() {
				log.Printf("Overflow clicked: %v", bar.SelectedOverflowAction())
			}
			if heartBtn.Clicked() {
				favorited = !favorited
			}
			if contextBtn.Clicked() {
				bar.SetContextualActions(
					[]materials.AppBarAction{
						materials.SimpleIconAction(th, &red, HeartIcon,
							materials.OverflowAction{
								Name: "House",
								Tag:  &red,
							},
						),
					},
					[]materials.OverflowAction{
						{
							Name: "foo",
							Tag:  &blue,
						},
						{
							Name: "bar",
							Tag:  &green,
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
				content := layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return pages[nav.CurrentNavDestination().(int)].layout(gtx)
					})
				})
				bar := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return bar.Layout(gtx)
				})
				flex := layout.Flex{Axis: layout.Vertical}
				if barOnBottom {
					flex.Layout(gtx, content, bar)
				} else {
					flex.Layout(gtx, bar, content)
				}
				modal.Layout(gtx)
				return layout.Dimensions{Size: gtx.Constraints.Max}
			})
			e.Frame(gtx.Ops)
		}
	}
}
