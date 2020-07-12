package materials

import (
	"image"
	"image/color"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type drawerState uint8

const (
	extended drawerState = iota
	extending
	retracted
	retracting
)

const (
	drawerAnimationDuration = time.Millisecond * 250
)

type NavItem struct {
	// Tag is an externally-provided identifier for the view
	// that this item should navigate to. It's value is
	// opaque to navigation elements.
	Tag  interface{}
	Name string
}

func (n NavItem) Layout(th *material.Theme, gtx layout.Context) layout.Dimensions {
	return layout.Inset{
		Top:    unit.Dp(4),
		Bottom: unit.Dp(4),
		Left:   unit.Dp(8),
		Right:  unit.Dp(8),
	}.Layout(gtx, func(gtx C) D {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
					return material.Label(th, unit.Dp(14), n.Name).Layout(gtx)
				})
			}),
		)
	})
}

// ModalNavDrawer implements the Material Design Modal Navigation Drawer
// described here: https://material.io/components/navigation-drawer
type ModalNavDrawer struct {
	*material.Theme

	Title    string
	Subtitle string
	Items    []NavItem

	scrim   widget.Clickable
	navList layout.List

	// animation state
	drawerState
	stateStarted time.Time
}

func (m *ModalNavDrawer) Layout(gtx layout.Context) layout.Dimensions {
	if m.scrim.Clicked() {
		m.drawerState = retracting
		m.stateStarted = gtx.Now
	}
	m.updateAnimationState(gtx)
	if m.drawerState == retracted {
		return layout.Dimensions{}
	}
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(m.layoutScrim),
		layout.Stacked(func(gtx C) D {
			if m.drawerState == retracting || m.drawerState == extending {
				defer op.Push(gtx.Ops).Pop()
				m.drawerTransform(gtx).Add(gtx.Ops)
				op.InvalidateOp{}.Add(gtx.Ops)
			}
			return m.layoutSheet(gtx)
		}),
	)
}

// updateAnimationState checks for drawer animations that have finished and
// updates the drawerState accordingly.
func (m *ModalNavDrawer) updateAnimationState(gtx layout.Context) {
	if m.drawerState == extended || m.drawerState == retracted {
		return
	}
	sinceStarted := gtx.Now.Sub(m.stateStarted)
	if sinceStarted > drawerAnimationDuration {
		if m.drawerState == retracting {
			m.drawerState = retracted
		} else if m.drawerState == extending {
			m.drawerState = extended
		}
		m.stateStarted = gtx.Now
	}
}

// drawerAnimationProgress returns the current animation progress as a value
// in the range [0,1)
func (m *ModalNavDrawer) drawerAnimationProgress(gtx layout.Context) float32 {
	sinceStarted := gtx.Now.Sub(m.stateStarted)
	return float32(sinceStarted.Milliseconds()) / float32(drawerAnimationDuration.Milliseconds())
}

// drawerTransform returns the TransformOp that should be used for the current
// animation frame.
func (m *ModalNavDrawer) drawerTransform(gtx layout.Context) op.TransformOp {
	progress := m.drawerAnimationProgress(gtx)
	if m.drawerState == retracting {
		progress *= -1
	} else if m.drawerState == extending {
		progress = -1 + progress
	}
	return op.Offset(f32.Point{X: progress * float32(m.sheetWidth(gtx))})
}

func (m *ModalNavDrawer) layoutScrim(gtx layout.Context) layout.Dimensions {
	defer op.Push(gtx.Ops).Pop()
	gtx.Constraints.Min = gtx.Constraints.Max
	paintRect(gtx, gtx.Constraints.Max, color.RGBA{A: 82})
	m.scrim.Layout(gtx)
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func (m ModalNavDrawer) sheetWidth(gtx layout.Context) int {
	scrimWidth := gtx.Px(unit.Dp(56))
	return gtx.Constraints.Max.X - scrimWidth
}

func (m *ModalNavDrawer) layoutSheet(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Max.X = m.sheetWidth(gtx)
	paintRect(gtx, gtx.Constraints.Max, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Inset{
				Left:   unit.Dp(16),
				Bottom: unit.Dp(18),
			}.Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						gtx.Constraints.Max.Y = gtx.Px(unit.Dp(36))
						gtx.Constraints.Min = gtx.Constraints.Max
						title := material.Label(m.Theme, unit.Dp(18), m.Title)
						title.Font.Weight = text.Bold
						return layout.SW.Layout(gtx, title.Layout)
					}),
					layout.Rigid(func(gtx C) D {
						gtx.Constraints.Max.Y = gtx.Px(unit.Dp(20))
						gtx.Constraints.Min = gtx.Constraints.Max
						return layout.SW.Layout(gtx, material.Label(m.Theme, unit.Dp(12), m.Subtitle).Layout)
					}),
				)
			})
		}),
		layout.Flexed(1, m.layoutNavList),
	)
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func (m *ModalNavDrawer) layoutNavList(gtx layout.Context) layout.Dimensions {
	m.navList.Axis = layout.Vertical
	return m.navList.Layout(gtx, len(m.Items), func(gtx C, index int) D {
		gtx.Constraints.Max.Y = gtx.Px(unit.Dp(48))
		gtx.Constraints.Min = gtx.Constraints.Max
		return m.Items[index].Layout(m.Theme, gtx)
	})
}

func (m *ModalNavDrawer) ToggleVisibility(when time.Time) {
	switch m.drawerState {
	case extending:
		fallthrough
	case extended:
		m.drawerState = retracting
	case retracting:
		fallthrough
	case retracted:
		m.drawerState = extending
	}
	m.stateStarted = when
}

func (m *ModalNavDrawer) CurrentNavDestiation() interface{} {
	return nil
}

func paintRect(gtx layout.Context, size image.Point, fill color.RGBA) {
	paint.ColorOp{Color: fill}.Add(gtx.Ops)
	paint.PaintOp{
		Rect: f32.Rectangle{
			Max: layout.FPt(size),
		},
	}.Add(gtx.Ops)
}
