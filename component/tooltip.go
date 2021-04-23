package component

import (
	"image"
	"image/color"
	"time"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
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

// TipArea holds the state information for displaying a tooltip.
type TipArea struct {
	VisibilityAnimation
	Appeared     time.Time
	HoverStarted time.Time
	Hovering     bool
	PressStarted time.Time
	Pressing     bool
	LongPressed  bool
	init         bool
}

const (
	tipAreaHoverDelay        = time.Millisecond * 500
	tipAreaLongPressDuration = time.Millisecond * 1500
	tipAreaFadeDuration      = time.Millisecond * 250
	longPressTheshold        = time.Millisecond * 500
)

// Layout renders the provided widget with the provided tooltip. The tooltip
// will be summoned if the widget is hovered or long-pressed.
func (t *TipArea) Layout(gtx C, tip Tooltip, w layout.Widget) D {
	if !t.init {
		t.init = true
		t.VisibilityAnimation.State = Invisible
		t.VisibilityAnimation.Duration = tipAreaFadeDuration
	}
	for _, e := range gtx.Events(t) {
		e, ok := e.(pointer.Event)
		if !ok {
			continue
		}
		switch e.Type {
		case pointer.Enter:
			t.Hovering = true
			t.HoverStarted = gtx.Now
		case pointer.Leave:
			t.VisibilityAnimation.Disappear(gtx.Now)
			t.Hovering = false
		case pointer.Press:
			t.Pressing = true
			t.PressStarted = gtx.Now
		case pointer.Release:
			t.Pressing = false
		case pointer.Cancel:
			t.Pressing = false
			t.Hovering = false
		}
	}
	if t.Hovering || t.Pressing || t.LongPressed {
		op.InvalidateOp{}.Add(gtx.Ops)
	}
	if t.Hovering && gtx.Now.Sub(t.HoverStarted) > tipAreaHoverDelay {
		t.VisibilityAnimation.Appear(gtx.Now)
		t.Appeared = gtx.Now
	}
	if t.Pressing && gtx.Now.Sub(t.PressStarted) > longPressTheshold {
		t.LongPressed = true
		t.VisibilityAnimation.Appear(gtx.Now)
		t.Appeared = gtx.Now
	}
	if t.LongPressed && gtx.Now.Sub(t.Appeared) > tipAreaLongPressDuration {
		t.VisibilityAnimation.Disappear(gtx.Now)
		t.LongPressed = false
	}
	return layout.Stack{}.Layout(gtx,
		layout.Stacked(w),
		layout.Expanded(func(gtx C) D {
			defer op.Save(gtx.Ops).Load()
			pointer.PassOp{Pass: true}.Add(gtx.Ops)
			pointer.Rect(image.Rectangle{Max: gtx.Constraints.Min}).Add(gtx.Ops)
			pointer.InputOp{
				Tag:   t,
				Types: pointer.Press | pointer.Release | pointer.Enter | pointer.Leave,
			}.Add(gtx.Ops)

			originalMin := gtx.Constraints.Min
			gtx.Constraints.Min = image.Point{}

			if t.Visible() {
				macro := op.Record(gtx.Ops)
				tip.Bg = Interpolate(color.NRGBA{}, tip.Bg, t.VisibilityAnimation.Revealed(gtx))
				dims := tip.Layout(gtx)
				call := macro.Stop()
				xOffset := float32((originalMin.X / 2) - (dims.Size.X / 2))
				yOffset := float32(originalMin.Y)
				macro = op.Record(gtx.Ops)
				op.Offset(f32.Pt(xOffset, yOffset)).Add(gtx.Ops)
				call.Add(gtx.Ops)
				call = macro.Stop()
				op.Defer(gtx.Ops, call)
			}
			return D{}
		}),
	)
}
