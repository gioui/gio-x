package component

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// Scrim implments a clickable translucent overlay. It can animate appearing
// and disappearing as a fade-in, fade-out transition from zero opacity
// to a fixed maximum opacity.
type Scrim struct {
	// FinalAlpha is the final opacity of the scrim on a scale from 0 to 255.
	FinalAlpha uint8
	widget.Clickable
}

// Layout draws the scrim using the provided animation. If the animation indicates
// that the scrim is not visible, this is a no-op.
func (s *Scrim) Layout(gtx layout.Context, th *material.Theme, anim *VisibilityAnimation) layout.Dimensions {
	if !anim.Visible() {
		return layout.Dimensions{}
	}
	defer op.Push(gtx.Ops).Pop()
	gtx.Constraints.Min = gtx.Constraints.Max
	currentAlpha := s.FinalAlpha
	if anim.Animating() {
		revealed := anim.Revealed(gtx)
		currentAlpha = uint8(float32(s.FinalAlpha) * revealed)
	}
	color := th.Fg
	color.A = currentAlpha
	fill := WithAlpha(color, currentAlpha)
	paintRect(gtx, gtx.Constraints.Max, fill)
	s.Clickable.Layout(gtx)
	return layout.Dimensions{Size: gtx.Constraints.Max}
}
