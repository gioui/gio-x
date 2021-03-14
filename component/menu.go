package component

import (
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// CardStyle defines the visual aspects of a material design surface
// with (optionally) rounded corners and a drop shadow.
type CardStyle struct {
	*material.Theme
	Radius    unit.Value
	Elevation unit.Value
	Shadow    color.NRGBA
}

func Card(th *material.Theme) CardStyle {
	return CardStyle{
		Theme:     th,
		Radius:    unit.Dp(0),
		Elevation: unit.Dp(8),
		Shadow:    color.NRGBA{A: 155},
	}
}

func (c CardStyle) Layout(gtx C, w layout.Widget) D {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			Shadow(c.Radius, c.Elevation).Layout(gtx)
			surface := clip.UniformRRect(f32.Rectangle{Max: layout.FPt(gtx.Constraints.Min)}, float32(gtx.Px(c.Radius)))
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
	layout.Inset
	CardStyle
}

func Menu(th *material.Theme, state *MenuState) MenuStyle {
	return MenuStyle{
		Theme:     th,
		MenuState: state,
		CardStyle: Card(th),
	}
}

func (m MenuStyle) Layout(gtx C) D {
	m.OptionList.Axis = layout.Vertical
	return m.CardStyle.Layout(gtx, func(gtx C) D {
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx C) D {
				return Rect{
					Color: m.Theme.Bg,
					Size:  gtx.Constraints.Min,
				}.Layout(gtx)
			}),
			layout.Stacked(func(gtx C) D {
				return m.OptionList.Layout(gtx, len(m.Options), func(gtx C, index int) D {
					return m.Options[index](gtx)
				})
			}),
		)
	})
}
