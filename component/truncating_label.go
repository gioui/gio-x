package component

import (
	"gioui.org/layout"
	"gioui.org/op"
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
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			return asLabel.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return truncationIndicator.Layout(gtx)
		}),
	)
}
