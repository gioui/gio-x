package component

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget/material"
)

// TruncatingLabelStyle is a type that forces a label to
// fit on one line and adds a truncation indicator symbol
// to the end of the line if the text has been truncated.
type TruncatingLabelStyle material.LabelStyle

// Layout renders the label into the provided context.
func (t TruncatingLabelStyle) Layout(gtx layout.Context) layout.Dimensions {
	originalMaxX := gtx.Constraints.Max.X
	gtx.Constraints.Max.X *= 2
	asLabel := material.LabelStyle(t)
	asLabel.MaxLines = 1
	macro := op.Record(gtx.Ops)
	dimensions := asLabel.Layout(gtx)
	labelOp := macro.Stop()
	if dimensions.Size.X <= originalMaxX {
		// No need to truncate
		labelOp.Add(gtx.Ops)
		return dimensions
	}
	gtx.Constraints.Max.X = originalMaxX
	truncationIndicator := asLabel
	truncationIndicator.Text = "…"
	macro = op.Record(gtx.Ops)
	gtx.Constraints.Min.X = 0
	indDims := truncationIndicator.Layout(gtx)
	indOp := macro.Stop()
	maxX := gtx.Constraints.Max.X - indDims.Size.X
	return layout.Flex{
		Alignment: layout.Baseline,
		Spacing:   layout.SpaceEnd,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			defer clip.Rect(image.Rectangle{
				Max: image.Point{
					Y: dimensions.Size.Y,
					X: maxX,
				},
			}).Push(gtx.Ops).Pop()
			labelOp.Add(gtx.Ops)
			dims := dimensions
			dims.Size.X = maxX
			return dims
		}),
		layout.Rigid(func(gtx C) D {
			indOp.Add(gtx.Ops)
			return indDims
		}),
	)
}
