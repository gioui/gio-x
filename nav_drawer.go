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

// renderNavItem holds both basic nav item state and the interaction
// state for that item.
type renderNavItem struct {
	*material.Theme
	NavItem
	hovering bool
	selected bool
	pressed  bool
}

func (n *renderNavItem) Layout(gtx layout.Context) layout.Dimensions {
	events := gtx.Events(n)
	for _, event := range events {
		switch event := event.(type) {
		case pointer.Event:
			switch event.Type {
			case pointer.Enter:
				n.hovering = true
			case pointer.Leave:
				n.hovering = false
				n.pressed = false
			case pointer.Press:
				n.pressed = true
			case pointer.Cancel:
				n.hovering = false
				n.pressed = false
			}
		}
	}
	defer op.Push(gtx.Ops).Pop()
	pointer.Rect(image.Rectangle{
		Max: gtx.Constraints.Max,
	}).Add(gtx.Ops)
	pointer.InputOp{
		Tag:   n,
		Types: pointer.Enter | pointer.Leave | pointer.Press | pointer.Release,
	}.Add(gtx.Ops)
	return layout.Inset{
		Top:    unit.Dp(4),
		Bottom: unit.Dp(4),
		Left:   unit.Dp(8),
		Right:  unit.Dp(8),
	}.Layout(gtx, func(gtx C) D {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Stack{}.Layout(gtx,
					layout.Expanded(n.layoutBackground),
					layout.Stacked(n.layoutContent),
				)
			}),
		)
	})
}

func (n *renderNavItem) layoutContent(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
		defer op.Push(gtx.Ops).Pop()
		macro := op.Record(gtx.Ops)
		label := material.Label(n.Theme, unit.Dp(14), n.Name)
		label.Font.Weight = text.Bold
		if n.selected {
			label.Color = n.Theme.Color.Primary
		}
		dimensions := label.Layout(gtx)
		labelOp := macro.Stop()
		top := (gtx.Constraints.Max.Y - dimensions.Size.Y) / 2
		op.Offset(f32.Point{Y: float32(top)}).Add(gtx.Ops)
		labelOp.Add(gtx.Ops)
		return layout.Dimensions{
			Size: gtx.Constraints.Max,
		}
	})
}

func (n *renderNavItem) layoutBackground(gtx layout.Context) layout.Dimensions {
	if !n.selected && !n.hovering {
		return layout.Dimensions{}
	}
	var fill color.RGBA
	if n.selected {
		fill = n.Theme.Color.Primary
	} else if n.hovering {
		fill = n.Theme.Color.Text
	}
	if n.selected && n.hovering {
		fill.A = 150
	} else {
		fill.A = 100
	}
	defer op.Push(gtx.Ops).Pop()
	rr := float32(gtx.Px(unit.Dp(4)))
	clip.RRect{
		Rect: f32.Rectangle{
			Max: layout.FPt(gtx.Constraints.Max),
		},
		NE: rr,
		SE: rr,
		NW: rr,
		SW: rr,
	}.Add(gtx.Ops)
	paintRect(gtx, gtx.Constraints.Max, fill)
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

// ModalNavDrawer implements the Material Design Modal Navigation Drawer
// described here: https://material.io/components/navigation-drawer
type ModalNavDrawer struct {
	*material.Theme

	Title    string
	Subtitle string

	selectedItem int
	items        []renderNavItem

	scrim   widget.Clickable
	navList layout.List

	// animation state
	drawerState
	stateStarted time.Time
}

func (m *ModalNavDrawer) AddNavItem(item NavItem) {
	m.items = append(m.items, renderNavItem{
		Theme:   m.Theme,
		NavItem: item,
	})
	if len(m.items) == 1 {
		m.items[0].selected = true
	}
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
	return m.navList.Layout(gtx, len(m.items), func(gtx C, index int) D {
		gtx.Constraints.Max.Y = gtx.Px(unit.Dp(48))
		gtx.Constraints.Min = gtx.Constraints.Max
		return m.items[index].Layout(gtx)
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
