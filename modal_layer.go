package materials

import (
	"time"

	"gioui.org/layout"
	"gioui.org/widget/material"
)

// ModalLayer is a widget drawn on top of the normal UI that can be populated
// by other material components with dismissble modal dialogs. For instance,
// the App Bar can render its overflow menu within the modal layer, and the
// modal navigation drawer is entirely within the modal layer.
type ModalLayer struct {
	VisibilityAnimation
	Scrim
	Widget func(gtx layout.Context, th *material.Theme, anim *VisibilityAnimation) layout.Dimensions
}

const defaultModalAnimationDuration = time.Millisecond * 250

// NewModal creates an initializes a modal layer.
func NewModal() *ModalLayer {
	m := ModalLayer{}
	m.VisibilityAnimation.State = Invisible
	m.VisibilityAnimation.Duration = defaultModalAnimationDuration
	m.Scrim.FinalAlpha = 82 //default
	return &m
}

// Layout renders the modal layer. Unless a modal widget has been triggered,
// this will do nothing.
func (m *ModalLayer) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if !m.Visible() {
		return D{}
	}
	if m.Scrim.Clicked() {
		m.Disappear(gtx.Now)
	}
	scrimDims := m.Scrim.Layout(gtx, th, &m.VisibilityAnimation)
	if m.Widget != nil {
		_ = m.Widget(gtx, th, &m.VisibilityAnimation)
	}
	return scrimDims
}
