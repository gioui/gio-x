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

// Tooltip implements a material design tool tip as defined at:
// https://material.io/components/tooltips#specs
type Tooltip struct {
	// Inset defines the interior padding of the Tooltip.
	layout.Inset
	// CornerRadius defines the corner radius of the RRect background.
	// of the tooltip.
	CornerRadius unit.Value
	// Text defines the content of the tooltip.
	Text material.LabelStyle
	// Bg defines the color of the RRect background.
	Bg color.NRGBA
}

// MobileTooltip constructs a tooltip suitable for use on mobile devices.
func MobileTooltip(th *material.Theme, text string) Tooltip {
	txt := material.Body1(th, text)
	txt.Color = th.Bg
	txt.TextSize = unit.Dp(16)
	return Tooltip{
		Inset: layout.Inset{
			Top:    unit.Dp(8),
			Bottom: unit.Dp(8),
			Left:   unit.Dp(16),
			Right:  unit.Dp(16),
		},
		Bg:           WithAlpha(th.Fg, 220),
		CornerRadius: unit.Dp(4),
		Text:         txt,
	}
}

// DesktopTooltip constructs a tooltip suitable for use on desktop devices.
func DesktopTooltip(th *material.Theme, text string) Tooltip {
	txt := material.Body2(th, text)
	txt.Color = th.Bg
	txt.TextSize = unit.Dp(12)
	return Tooltip{
		Inset: layout.Inset{
			Top:    unit.Dp(6),
			Bottom: unit.Dp(6),
			Left:   unit.Dp(8),
			Right:  unit.Dp(8),
		},
		Bg:           WithAlpha(th.Fg, 220),
		CornerRadius: unit.Dp(4),
		Text:         txt,
	}
}

// Layout renders the tooltip.
func (t Tooltip) Layout(gtx C) D {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			radius := float32(gtx.Px(t.CornerRadius))
			paint.FillShape(gtx.Ops, t.Bg, clip.RRect{
				Rect: f32.Rectangle{
					Max: layout.FPt(gtx.Constraints.Min),
				},
				NW: radius,
				NE: radius,
				SW: radius,
				SE: radius,
			}.Op(gtx.Ops))
			return D{}
		}),
		layout.Stacked(func(gtx C) D {
			return t.Inset.Layout(gtx, t.Text.Layout)
		}),
	)
}
