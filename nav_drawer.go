package materials

import (
	"image"
	"image/color"
	"math"
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
	retracted drawerState = iota
	retracting
	extended
	extending
)

var (
	hoverOverlayAlpha     = 0.16 * 256
	focuseOverlayAlpha    = 0.48 * 256
	selectedOverlayAlpha  = 0.32 * 256
	activatedOverlayAlpha = 0.48 * 256
	pressedOverlayAlpha   = 0.48 * 256
	draggedOverlayAlpha   = 0.32 * 256
)

const (
	drawerAnimationDuration   = time.Millisecond * 250
	navPressAnimationDuration = time.Millisecond * 250
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
	hovering       bool
	selected       bool
	pressed        bool
	animatingPress bool
	widget.Press
	clicked bool
}

func (n *renderNavItem) updateAnimationState(gtx layout.Context) {
	if !n.animatingPress {
		return
	}
	sinceStarted := gtx.Now.Sub(n.Press.Start)
	if sinceStarted > navPressAnimationDuration {
		n.animatingPress = false
	}
}

func (n *renderNavItem) Clicked() bool {
	return n.clicked
}

func (n *renderNavItem) Layout(gtx layout.Context) layout.Dimensions {
	n.clicked = false
	n.updateAnimationState(gtx)
	events := gtx.Events(n)
	for _, event := range events {
		switch event := event.(type) {
		case pointer.Event:
			switch event.Type {
			case pointer.Enter:
				n.hovering = true
			case pointer.Leave:
				n.hovering = false
			case pointer.Press:
				n.pressed = true
				n.Press.Start = gtx.Now
				n.Press.Position = event.Position
				n.Press.Cancelled = false
				n.Press.End = time.Time{}
			case pointer.Cancel:
				n.hovering = false
				n.Press.Cancelled = true
			case pointer.Release:
				n.animatingPress = true
				n.Press.End = gtx.Now
				n.clicked = true
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
		if n.hovering {
			label.Color = n.Theme.Color.Text
		} else if n.selected {
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

func (n *renderNavItem) pressAnimationProgress(gtx layout.Context) float32 {
	sinceStarted := gtx.Now.Sub(n.Press.Start)
	return float32(sinceStarted.Milliseconds()) / float32(navPressAnimationDuration.Milliseconds())
}

func (n *renderNavItem) layoutBackground(gtx layout.Context) layout.Dimensions {
	if !n.selected && !n.hovering {
		return layout.Dimensions{}
	}
	var fill color.RGBA
	if n.hovering {
		fill = n.Theme.Color.Text
		fill.A = uint8(hoverOverlayAlpha)
	} else if n.selected {
		fill = n.Theme.Color.Primary
		fill.A = uint8(selectedOverlayAlpha)
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
	if n.pressed {
		n.drawInk(gtx)
	}
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

// adapted from https://git.sr.ht/~eliasnaur/gio/tree/773939fe1dd10b3ac5f937a7f9993045a91e23a7/widget/material/button.go#L189
func (n *renderNavItem) drawInk(gtx layout.Context) {
	// duration is the number of seconds for the
	// completed animation: expand while fading in, then
	// out.
	const (
		expandDuration = float32(0.5)
		fadeDuration   = float32(0.9)
	)

	c := n.Press

	now := gtx.Now

	t := float32(now.Sub(c.Start).Seconds())

	end := c.End
	if end.IsZero() {
		// If the press hasn't ended, don't fade-out.
		end = now
	}

	endt := float32(end.Sub(c.Start).Seconds())

	// Compute the fade-in/out position in [0;1].
	var alphat float32
	{
		var haste float32
		if c.Cancelled {
			// If the press was cancelled before the inkwell
			// was fully faded in, fast forward the animation
			// to match the fade-out.
			if h := 0.5 - endt/fadeDuration; h > 0 {
				haste = h
			}
		}
		// Fade in.
		half1 := t/fadeDuration + haste
		if half1 > 0.5 {
			half1 = 0.5
		}

		// Fade out.
		half2 := float32(now.Sub(end).Seconds())
		half2 /= fadeDuration
		half2 += haste
		if half2 > 0.5 {
			// Too old.
			return
		}

		alphat = half1 + half2
	}

	// Compute the expand position in [0;1].
	sizet := t
	if c.Cancelled {
		// Freeze expansion of cancelled presses.
		sizet = endt
	}
	sizet /= expandDuration

	// Animate only ended presses, and presses that are fading in.
	if !c.End.IsZero() || sizet <= 1.0 {
		op.InvalidateOp{}.Add(gtx.Ops)
	}

	if sizet > 1.0 {
		sizet = 1.0
	}

	if alphat > .5 {
		// Start fadeout after half the animation.
		alphat = 1.0 - alphat
	}
	// Twice the speed to attain fully faded in at 0.5.
	t2 := alphat * 2
	// Beziér ease-in curve.
	alphaBezier := t2 * t2 * (3.0 - 2.0*t2)
	sizeBezier := sizet * sizet * (3.0 - 2.0*sizet)
	size := float32(gtx.Constraints.Min.X)
	if h := float32(gtx.Constraints.Min.Y); h > size {
		size = h
	}
	// Cover the entire constraints min rectangle.
	size *= 2 * float32(math.Sqrt(2))
	// Apply curve values to size and color.
	size *= sizeBezier
	alpha := 0.7 * alphaBezier
	const col = 0.8
	ba, bc := byte(alpha*0xff), byte(alpha*col*0xff)
	defer op.Push(gtx.Ops).Pop()
	ink := paint.ColorOp{Color: color.RGBA{A: ba, R: bc, G: bc, B: bc}}
	ink.Add(gtx.Ops)
	rr := size * .5
	op.Offset(c.Position.Add(f32.Point{
		X: -rr,
		Y: -rr,
	})).Add(gtx.Ops)
	clip.RRect{
		Rect: f32.Rectangle{Max: f32.Point{
			X: float32(size),
			Y: float32(size),
		}},
		NE: rr, NW: rr, SE: rr, SW: rr,
	}.Add(gtx.Ops)
	paint.PaintOp{Rect: f32.Rectangle{Max: f32.Point{X: float32(size), Y: float32(size)}}}.Add(gtx.Ops)
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

// AddNavItem inserts a navigation target into the drawer. This should be
// invoked only from the layout thread to avoid nasty race conditions.
func (m *ModalNavDrawer) AddNavItem(item NavItem) {
	m.items = append(m.items, renderNavItem{
		Theme:   m.Theme,
		NavItem: item,
	})
	if len(m.items) == 1 {
		m.items[0].selected = true
	}
}

// Layout renders the nav drawer
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
	const finalAlpha = 82
	currentAlpha := uint8(finalAlpha)
	if m.drawerState == extending {
		currentAlpha = uint8(finalAlpha * m.drawerAnimationProgress(gtx))
	} else if m.drawerState == retracting {
		currentAlpha = finalAlpha - uint8(finalAlpha*m.drawerAnimationProgress(gtx))
	}
	paintRect(gtx, gtx.Constraints.Max, color.RGBA{A: currentAlpha})
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
		dimensions := m.items[index].Layout(gtx)
		if m.items[index].Clicked() {
			m.items[m.selectedItem].selected = false
			m.selectedItem = index
			m.items[m.selectedItem].selected = true
			m.ToggleVisibility(gtx.Now)
		}
		return dimensions
	})
}

// ToggleVisibility changes the state of the nav drawer from retracted to
// extended or visa versa.
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

// CurrentNavDestiation returns the tag of the navigation destination
// selected in the drawer.
func (m *ModalNavDrawer) CurrentNavDestiation() interface{} {
	return m.items[m.selectedItem].Tag
}

func paintRect(gtx layout.Context, size image.Point, fill color.RGBA) {
	paint.ColorOp{Color: fill}.Add(gtx.Ops)
	paint.PaintOp{
		Rect: f32.Rectangle{
			Max: layout.FPt(size),
		},
	}.Add(gtx.Ops)
}
