package component

import (
	"gioui.org/f32"
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
		ShadowStyle: Shadow(unit.Dp(8), unit.Dp(8)),
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
	SurfaceStyle
}

// Menu constructs a menu with the provided state and a default Surface behind
// it.
func Menu(th *material.Theme, state *MenuState) MenuStyle {
	return MenuStyle{
		Theme:        th,
		MenuState:    state,
		SurfaceStyle: Card(th),
	}
}

// Layout renders the menu.
func (m MenuStyle) Layout(gtx C) D {
	m.OptionList.Axis = layout.Vertical
	return m.SurfaceStyle.Layout(gtx, func(gtx C) D {
		return m.OptionList.Layout(gtx, len(m.Options), func(gtx C, index int) D {
			return m.Options[index](gtx)
		})
	})
}
