package outlay

import (
	"gioui.org/layout"
	"gioui.org/unit"
)

// If shows a child widget if the boolean expression is true.
// Allows the caller to easily display content conditionally
// without needing to boilerplate the noop branch.
type If bool

func (i If) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	if !i {
		return layout.Dimensions{}
	}
	return w(gtx)
}

func (i If) Flexed(weight float32, w layout.Widget) layout.FlexChild {
	if i {
		return layout.Flexed(weight, w)
	}

	return layout.Rigid(func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} })
}

func (i If) Rigid(w layout.Widget) layout.FlexChild {
	if i {
		return layout.Rigid(w)
	}

	return layout.Rigid(func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} })
}

func (i If) EmptyRigid(sz unit.Dp) layout.FlexChild {
	if i {
		return EmptyRigid(sz)
	}

	return layout.Rigid(func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} })
}

func (i If) EmptyRigidHorizontal(sz unit.Dp) layout.FlexChild {
	if i {
		return EmptyRigidHorizontal(sz)
	}

	return layout.Rigid(func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} })
}

func (i If) EmptyRigidVertical(sz unit.Dp) layout.FlexChild {
	if i {
		return EmptyRigidVertical(sz)
	}

	return layout.Rigid(func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} })
}

func (i If) EmptyFlexed() layout.FlexChild {
	if i {
		return EmptyFlexed()
	}

	return layout.Rigid(func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} })
}

func (i If) Stacked(w layout.Widget) layout.StackChild {
	if i {
		return layout.Stacked(w)
	}

	return layout.Stacked(func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} })
}

func (i If) Expanded(w layout.Widget) layout.StackChild {
	if i {
		return layout.Expanded(w)
	}

	return layout.Expanded(func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} })
}
