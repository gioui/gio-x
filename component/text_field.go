package component

import (
	"image"
	"image/color"
	"strconv"
	"time"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
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
	// Alignment specifies where to anchor the text.
	Alignment layout.Alignment

	// Helper text to give additional context to a field.
	Helper string
	// CharLimit specifies the maximum number of characters the text input
	// will allow. Zero means "no limit".
	CharLimit uint
	// Prefix appears before the content of the text input.
	Prefix layout.Widget
	// Suffix appears after the content of the text input.
	Suffix layout.Widget

	// Animation state.
	state
	label  label
	border border
	helper helper
	anim   *Progress

	// errored tracks whether the input is in an errored state.
	// This is orthogonal to the other states: the input can be both errored
	// and inactive for example.
	errored bool
}

// Validator validates text and returns a string describing the error.
// Error is displayed as helper text.
type Validator = func(string) string

type label struct {
	TextSize float32
	Inset    layout.Inset
	Smallest layout.Dimensions
}

type border struct {
	Thickness float32
	Color     color.NRGBA
}

type helper struct {
	Color color.NRGBA
	Text  string
}

type state int

const (
	inactive state = iota
	hovered
	activated
	focused
)

// IsActive if input is in an active state (Active, Focused or Errored).
func (in TextField) IsActive() bool {
	return in.state >= activated
}

// IsErrored if input is in an errored state.
// Typically this is when the validator has returned an error message.
func (in *TextField) IsErrored() bool {
	return in.errored
}

// SetError puts the input into an errored state with the specified error text.
func (in *TextField) SetError(err string) {
	in.errored = true
	in.helper.Text = err
}

// ClearError clears any errored status.
func (in *TextField) ClearError() {
	in.errored = false
	in.helper.Text = in.Helper
}

// Clear the input text and reset any error status.
func (in *TextField) Clear() {
	in.Editor.SetText("")
	in.ClearError()
}

// TextTooLong returns whether the current editor text exceeds the set character
// limit.
func (in *TextField) TextTooLong() bool {
	return !(in.CharLimit == 0 || uint(len(in.Editor.Text())) < in.CharLimit)
}

func (in *TextField) Update(gtx C, th *material.Theme, hint string) {
	var disabled = gtx.Queue == nil
	for in.Hoverable.Clicked() {
		in.Editor.Focus()
	}
	in.state = inactive
	if in.Hoverable.Hovered() && !disabled {
		in.state = hovered
	}
	if in.Editor.Len() > 0 {
		in.state = activated
	}
	if in.Editor.Focused() && !disabled {
		in.state = focused
	}
	const (
		duration = time.Millisecond * 100
	)
	if in.anim == nil {
		in.anim = &Progress{}
	}
	if in.state == activated {
		in.anim.Start(gtx.Now, Forward, 0)
	}
	if in.state == focused && in.Editor.Len() == 0 && !in.anim.Started() {
		in.anim.Start(gtx.Now, Forward, duration)
	}
	if in.state == inactive && in.Editor.Len() == 0 && in.anim.Finished() {
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
		borderColor        = WithAlpha(th.Palette.Fg, 128)
		borderColorHovered = WithAlpha(th.Palette.Fg, 221)
		borderColorActive  = th.Palette.ContrastBg
		// TODO: derive from Theme.Error or Theme.Danger
		dangerColor = color.NRGBA{R: 200, A: 255}
		// Border thickness transitions.
		borderThickness       = float32(0.5)
		borderThicknessActive = float32(2.0)
	)
	in.label.TextSize = lerp(textSmall.V, textNormal.V, 1.0-in.anim.Progress())
	switch in.state {
	case inactive:
		in.border.Thickness = borderThickness
		in.border.Color = borderColor
		in.helper.Color = borderColor
	case hovered, activated:
		in.border.Thickness = borderThickness
		in.border.Color = borderColorHovered
		in.helper.Color = borderColorHovered
	case focused:
		in.border.Thickness = borderThicknessActive
		in.border.Color = borderColorActive
		in.helper.Color = borderColorHovered
	}
	if in.IsErrored() {
		in.border.Color = dangerColor
		in.helper.Color = dangerColor
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
	defer op.Save(gtx.Ops).Load()
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
	dims := layout.Flex{
		Axis: layout.Vertical,
	}.Layout(
		gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Stack{}.Layout(
				gtx,
				layout.Expanded(func(gtx C) D {
					var cornerRadius = unit.Dp(4)
					macro := op.Record(gtx.Ops)
					dims := widget.Border{
						Color:        in.border.Color,
						Width:        unit.Dp(in.border.Thickness),
						CornerRadius: cornerRadius,
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
						clips := []func(gtx C){
							func(gtx C) {
								clip.RRect{
									Rect: layout.FRect(image.Rectangle{
										Max: image.Point{
											X: gtx.Px(in.label.Inset.Left),
											Y: gtx.Constraints.Min.Y,
										},
									}),
									NW: float32(gtx.Px(cornerRadius)),
									SW: float32(gtx.Px(cornerRadius)),
								}.Add(gtx.Ops)
							},
							func(gtx C) {
								clip.Rect{
									Min: image.Point{
										X: gtx.Px(in.label.Inset.Left),
										Y: in.label.Smallest.Size.Y / 2,
									},
									Max: image.Point{
										X: gtx.Px(in.label.Inset.Left) + in.label.Smallest.Size.X,
										Y: gtx.Constraints.Min.Y,
									},
								}.Add(gtx.Ops)
							},
							func(gtx C) {
								clip.RRect{
									Rect: layout.FRect(image.Rectangle{
										Min: image.Point{
											X: gtx.Px(in.label.Inset.Left) + in.label.Smallest.Size.X,
										},
										Max: image.Point{
											X: gtx.Constraints.Max.X,
											Y: gtx.Constraints.Min.Y,
										},
									}),
									NE: float32(gtx.Px(cornerRadius)),
									SE: float32(gtx.Px(cornerRadius)),
								}.Add(gtx.Ops)
							},
						}
						for _, c := range clips {
							stack := op.Save(gtx.Ops)
							c(gtx)
							border.Add(gtx.Ops)
							stack.Load()
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
							return layout.Flex{
								Axis:      layout.Horizontal,
								Alignment: layout.Middle,
								Spacing: func() layout.Spacing {
									switch in.Alignment {
									case layout.Middle:
										return layout.SpaceSides
									case layout.End:
										return layout.SpaceStart
									default: // layout.Start and all others
										return layout.SpaceEnd
									}
								}(),
							}.Layout(
								gtx,
								layout.Rigid(func(gtx C) D {
									if in.IsActive() && in.Prefix != nil {
										return in.Prefix(gtx)
									}
									return D{}
								}),
								layout.Rigid(func(gtx C) D {
									return material.Editor(th, &in.Editor, "").Layout(gtx)
								}),
								layout.Rigid(func(gtx C) D {
									if in.IsActive() && in.Suffix != nil {
										return in.Suffix(gtx)
									}
									return D{}
								}),
							)
						},
					)
				}),
				layout.Expanded(func(gtx C) D {
					return in.Hoverable.Layout(gtx)
				}),
			)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
				Spacing:   layout.SpaceBetween,
			}.Layout(
				gtx,
				layout.Rigid(func(gtx C) D {
					if in.helper.Text == "" {
						return D{}
					}
					return layout.Inset{
						Top:  unit.Dp(4),
						Left: unit.Dp(10),
					}.Layout(
						gtx,
						func(gtx C) D {
							helper := material.Label(th, unit.Dp(12), in.helper.Text)
							helper.Color = in.helper.Color
							return helper.Layout(gtx)
						},
					)
				}),
				layout.Rigid(func(gtx C) D {
					if in.CharLimit == 0 {
						return D{}
					}
					return layout.Inset{
						Top:   unit.Dp(4),
						Right: unit.Dp(10),
					}.Layout(
						gtx,
						func(gtx C) D {
							count := material.Label(
								th,
								unit.Dp(12),
								strconv.Itoa(in.Editor.Len())+"/"+strconv.Itoa(int(in.CharLimit)),
							)
							count.Color = in.helper.Color
							return count.Layout(gtx)
						},
					)
				}),
			)
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
		stack := op.Save(gtx.Ops)
		pointer.PassOp{Pass: true}.Add(gtx.Ops)
		pointer.Rect(image.Rectangle{Max: gtx.Constraints.Min}).Add(gtx.Ops)
		h.Clickable.Layout(gtx)
		stack.Load()
	}
	h.update(gtx)
	{
		stack := op.Save(gtx.Ops)
		pointer.PassOp{Pass: true}.Add(gtx.Ops)
		pointer.Rect(image.Rectangle{Max: gtx.Constraints.Min}).Add(gtx.Ops)
		pointer.InputOp{
			Tag:   h,
			Types: pointer.Enter | pointer.Leave | pointer.Cancel,
		}.Add(gtx.Ops)
		stack.Load()
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
