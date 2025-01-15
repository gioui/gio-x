package outlay

import (
	"gioui.org/layout"
	"gioui.org/unit"
)

// Spacer spaces along both axis according to the value.
type Spacer unit.Dp

func (s Spacer) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Spacer{
		Height: unit.Dp(s),
		Width:  unit.Dp(s),
	}.Layout(gtx)
}

// Space returns a widget that spaces both axis by a size.
func Space(sz unit.Dp) layout.Widget {
	return Spacer(sz).Layout
}

// EmptyRigid is like space but returns a flex child directly.
func EmptyRigid(sz unit.Dp) layout.FlexChild {
	return layout.Rigid(Space(sz))
}

// EmptyRigidHorizontal is like EmptyRigid, but stretches only the horizontal space.
func EmptyRigidHorizontal(sz unit.Dp) layout.FlexChild {
	return layout.Rigid(layout.Spacer{Width: sz}.Layout)
}

// EmptyRigidVertical is like EmptyRigid, but stretches only the vertical space.
func EmptyRigidVertical(sz unit.Dp) layout.FlexChild {
	return layout.Rigid(layout.Spacer{Height: sz}.Layout)
}

// EmptyFlexed returns a flex that consumes its available space.
// Use this to push rigid widgets around.
func EmptyFlexed() layout.FlexChild {
	return layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: gtx.Constraints.Min}
	})
}
