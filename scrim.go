package materials

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
)

type scrim struct {
	finalAlpha uint8
	*visibilityAnimation
	widget.Clickable
}

func (s *scrim) Layout(gtx layout.Context) layout.Dimensions {
	defer op.Push(gtx.Ops).Pop()
	gtx.Constraints.Min = gtx.Constraints.Max
	currentAlpha := s.finalAlpha
	revealed := s.visibilityAnimation.Revealed(gtx, actionAnimationDuration)
	currentAlpha = uint8(float32(s.finalAlpha) * revealed)
	paintRect(gtx, gtx.Constraints.Max, color.RGBA{A: currentAlpha})
	s.Clickable.Layout(gtx)
	return layout.Dimensions{Size: gtx.Constraints.Max}
}
