package materials

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
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// TextField implements the Material Design Text Field
// described here: https://material.io/components/text-fields
type TextField struct {
	// Editor contains the edit buffer.
	widget.Editor
	// Hoverable detects mouse hovers.
	Hoverable Hoverable

	// Animation state.
	label
	border
	anim *Progress
}

type label struct {
	TextSize float32
	Inset    layout.Inset
	Smallest layout.Dimensions
}

type border struct {
	Thickness float32
	Color     color.RGBA
}

func (in *TextField) Update(gtx C, th *material.Theme, hint string) {
	for in.Hoverable.Clicked() {
		in.Editor.Focus()
	}
	const (
		duration = time.Millisecond * 100
	)
	if in.anim == nil {
		in.anim = &Progress{}
	}
	if in.Editor.Focused() && !in.anim.Started() {
		in.anim.Start(gtx.Now, Forward, duration)
	}
	if !in.Editor.Focused() && in.Editor.Len() == 0 && in.anim.Finished() {
		in.anim.Start(gtx.Now, Reverse, duration)
	}
	if in.anim.Started() {
		op.InvalidateOp{}.Add(gtx.Ops)
	}
	in.anim.Update(gtx.Now)
	var (
		// Text size transitions.
		textNormal = th.TextSize
		textSmall  = th.TextSize.Scale(0.8)
		// Border color transitions.
		borderColor        = color.RGBA{A: 107}
		borderColorHovered = color.RGBA{A: 221}
		borderColorActive  = th.Color.Primary
		// Border thickness transitions.
		borderThickness       = float32(0.5)
		borderThicknessActive = float32(2.0)
	)
	in.label.TextSize = lerp(textSmall.V, textNormal.V, 1.0-in.anim.Progress())
	in.border.Thickness = borderThickness
	in.border.Color = borderColor
	if in.Hoverable.Hovered() {
		in.border.Color = borderColorHovered
	}
	if in.Editor.Focused() {
		in.border.Thickness = borderThicknessActive
		in.border.Color = borderColorActive
	}
	// Calculate the dimensions of the smallest label size and store the
	// result for use in clipping.
	// Hack: Reset min constraint to 0 to avoid min == max.
	gtx.Constraints.Min.X = 0
	macro := op.Record(gtx.Ops)
	in.label.Smallest = layout.Inset{
		Left:  unit.Dp(4),
		Right: unit.Dp(4),
	}.Layout(gtx, func(gtx C) D {
		return material.Label(th, textSmall, hint).Layout(gtx)
	})
	macro.Stop()
	labelTopInsetNormal := float32(in.label.Smallest.Size.Y) - float32(in.label.Smallest.Size.Y/4)
	labelTopInsetActive := (labelTopInsetNormal / 2 * -1) - in.border.Thickness
	in.label.Inset = layout.Inset{
		Top:  unit.Px(lerp(labelTopInsetNormal, labelTopInsetActive, in.anim.Progress())),
		Left: unit.Dp(10),
	}
}

func (in *TextField) Layout(gtx C, th *material.Theme, hint string) D {
	in.Update(gtx, th, hint)
	defer op.Push(gtx.Ops).Pop()
	// Offset accounts for label height, which sticks above the border dimensions.
	op.Offset(f32.Pt(0, float32(in.label.Smallest.Size.Y)/2)).Add(gtx.Ops)
	in.label.Inset.Layout(
		gtx,
		func(gtx C) D {
			return layout.Inset{
				Left:  unit.Dp(4),
				Right: unit.Dp(4),
			}.Layout(gtx, func(gtx C) D {
				label := material.Label(th, unit.Sp(in.label.TextSize), hint)
				label.Color = in.border.Color
				return label.Layout(gtx)
			})
		})
	dims := layout.Stack{}.Layout(
		gtx,
		layout.Expanded(func(gtx C) D {
			macro := op.Record(gtx.Ops)
			dims := widget.Border{
				Color:        in.border.Color,
				Width:        unit.Dp(in.border.Thickness),
				CornerRadius: unit.Dp(4),
			}.Layout(
				gtx,
				func(gtx C) D {
					return D{Size: image.Point{
						X: gtx.Constraints.Max.X,
						Y: gtx.Constraints.Min.Y,
					}}
				},
			)
			border := macro.Stop()
			if in.Editor.Focused() || in.Editor.Len() > 0 {
				// Clip a concave shape which ignores the label area.
				clips := []clip.Rect{
					{
						Max: image.Point{
							X: gtx.Px(in.label.Inset.Left),
							Y: gtx.Constraints.Min.Y,
						},
					},
					{
						Min: image.Point{
							X: gtx.Px(in.label.Inset.Left),
							Y: in.label.Smallest.Size.Y / 2,
						},
						Max: image.Point{
							X: gtx.Px(in.label.Inset.Left) + in.label.Smallest.Size.X,
							Y: gtx.Constraints.Min.Y,
						},
					},
					{
						Min: image.Point{
							X: gtx.Px(in.label.Inset.Left) + in.label.Smallest.Size.X,
						},
						Max: image.Point{
							X: gtx.Constraints.Max.X,
							Y: gtx.Constraints.Min.Y,
						},
					},
				}
				for _, c := range clips {
					stack := op.Push(gtx.Ops)
					c.Add(gtx.Ops)
					border.Add(gtx.Ops)
					stack.Pop()
				}
			} else {
				border.Add(gtx.Ops)
			}
			return dims
		}),
		layout.Stacked(func(gtx C) D {
			return layout.UniformInset(unit.Dp(12)).Layout(
				gtx,
				func(gtx C) D {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					return material.Editor(th, &in.Editor, "").Layout(gtx)
				},
			)
		}),
		layout.Expanded(func(gtx C) D {
			return in.Hoverable.Layout(gtx)
		}),
	)
	return D{
		Size: image.Point{
			X: dims.Size.X,
			Y: dims.Size.Y + in.label.Smallest.Size.Y/2,
		},
		Baseline: dims.Baseline,
	}
}

// interpolate linearly between two values based on progress.
//
// Progress is expected to be [0, 1]. Values greater than 1 will therefore be
// become a coeficient.
//
// For example, 2.5 is 250% progress.
func lerp(start, end, progress float32) float32 {
	return start + (end-start)*progress
}

// Hoverable tracks mouse hovers over some area.
type Hoverable struct {
	widget.Clickable
	hovered bool
}

// Hovered if mouse has entered the area.
func (h *Hoverable) Hovered() bool {
	return h.hovered
}

// Layout Hoverable according to min constraints.
func (h *Hoverable) Layout(gtx C) D {
	{
		stack := op.Push(gtx.Ops)
		pointer.PassOp{Pass: true}.Add(gtx.Ops)
		pointer.Rect(image.Rectangle{Max: gtx.Constraints.Min}).Add(gtx.Ops)
		h.Clickable.Layout(gtx)
		stack.Pop()
	}
	h.update(gtx)
	{
		stack := op.Push(gtx.Ops)
		pointer.PassOp{Pass: true}.Add(gtx.Ops)
		pointer.Rect(image.Rectangle{Max: gtx.Constraints.Min}).Add(gtx.Ops)
		pointer.InputOp{
			Tag:   h,
			Types: pointer.Enter | pointer.Leave | pointer.Cancel,
		}.Add(gtx.Ops)
		stack.Pop()
	}
	return D{Size: gtx.Constraints.Min}

}

func (h *Hoverable) update(gtx C) {
	for _, event := range gtx.Events(h) {
		if event, ok := event.(pointer.Event); ok {
			switch event.Type {
			case pointer.Enter:
				h.hovered = true
			case pointer.Leave, pointer.Cancel:
				h.hovered = false
			}
		}
	}
}

type Rect struct {
	Color color.RGBA
	Size  image.Point
	Radii float32
}

func (r Rect) Layout(gtx C) D {
	paint.FillShape(gtx.Ops, clip.UniformRRect(f32.Rectangle{Max: layout.FPt(r.Size)}, r.Radii).Op(gtx.Ops), r.Color)
	return layout.Dimensions{Size: image.Pt(int(r.Size.X), int(r.Size.Y))}

}
