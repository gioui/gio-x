package component

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// SurfaceStyle defines the visual aspects of a material design surface
// with (optionally) rounded corners and a drop shadow.
type SurfaceStyle struct {
	*material.Theme
	// The CornerRadius and Elevation fields of the embedded shadow
	// style also define the corner radius and elevation of the card.
	ShadowStyle
}

func Card(th *material.Theme) SurfaceStyle {
	return SurfaceStyle{
		Theme:       th,
		ShadowStyle: Shadow(unit.Dp(4), unit.Dp(4)),
	}
}

func (c SurfaceStyle) Layout(gtx C, w layout.Widget) D {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			c.ShadowStyle.Layout(gtx)
			surface := clip.UniformRRect(f32.Rectangle{Max: layout.FPt(gtx.Constraints.Min)}, float32(gtx.Px(c.ShadowStyle.CornerRadius)))
			paint.FillShape(gtx.Ops, c.Theme.Bg, surface.Op(gtx.Ops))
			return D{Size: gtx.Constraints.Min}
		}),
		layout.Stacked(w),
	)
}

type MenuItemStyle struct {
	State      *gesture.Click
	HoverColor color.NRGBA
	layout.Inset
	material.LabelStyle
}

func MenuItem(th *material.Theme, state *gesture.Click, label string) MenuItemStyle {
	return MenuItemStyle{
		State: state,
		Inset: layout.Inset{
			Left:   unit.Dp(16),
			Right:  unit.Dp(16),
			Top:    unit.Dp(16),
			Bottom: unit.Dp(16),
		},
		LabelStyle: material.Body1(th, label),
		HoverColor: WithAlpha(th.ContrastBg, 0x30),
	}
}

func (m MenuItemStyle) Layout(gtx C) D {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			area := image.Rectangle{
				Max: gtx.Constraints.Min,
			}
			pointer.Rect(area).Add(gtx.Ops)
			m.State.Add(gtx.Ops)
			m.State.Events(gtx)
			if m.State.Hovered() {
				paint.FillShape(gtx.Ops, m.HoverColor, clip.Rect(area).Op())
			}
			return D{Size: area.Max}
		}),
		layout.Stacked(func(gtx C) D {
			return m.Inset.Layout(gtx, func(gtx C) D {
				return m.LabelStyle.Layout(gtx)
			})
		}),
	)
}

// MenuState holds the state of a menu material design component
// across frames.
type MenuState struct {
	OptionList layout.List
	Options    []func(gtx C) D
}

// MenuStyle defines the presentation of a material design menu component.
type MenuStyle struct {
	*MenuState
	*material.Theme
	// Inset applied around the rendered contents of the state's Options field.
	layout.Inset
	SurfaceStyle
}

// Menu constructs a menu with the provided state and a default Surface behind
// it.
func Menu(th *material.Theme, state *MenuState) MenuStyle {
	m := MenuStyle{
		Theme:        th,
		MenuState:    state,
		SurfaceStyle: Card(th),
		Inset: layout.Inset{
			Top:    unit.Dp(8),
			Bottom: unit.Dp(8),
		},
	}
	m.OptionList.Axis = layout.Vertical
	return m
}

// Layout renders the menu.
func (m MenuStyle) Layout(gtx C) D {
	return m.SurfaceStyle.Layout(gtx, func(gtx C) D {
		return m.Inset.Layout(gtx, func(gtx C) D {
			return m.OptionList.Layout(gtx, len(m.Options), func(gtx C, index int) D {
				return m.Options[index](gtx)
			})
		})
	})
}
