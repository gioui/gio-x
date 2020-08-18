package materials

import (
	"image"
	"image/color"
	"time"

	"gioui.org/f32"
	"gioui.org/gesture"
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

	// Icon, if set, renders the provided icon to the left of the
	// item's name. Material specifies that either all navigation
	// items should have an icon, or none should. As such, if this
	// field is nil, the Name will be aligned all the way to the
	// left. A mixture of icon and non-icon items will be misaligned.
	// Users should either set icons for all elements or none.
	Icon *widget.Icon
}

// renderNavItem holds both basic nav item state and the interaction
// state for that item.
type renderNavItem struct {
	*material.Theme
	NavItem
	hovering bool
	selected bool
	widget.Clickable
}

func (n *renderNavItem) Clicked() bool {
	return n.Clickable.Clicked()
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
			case pointer.Cancel:
				n.hovering = false
			}
		}
	}
	defer op.Push(gtx.Ops).Pop()
	pointer.PassOp{Pass: true}.Add(gtx.Ops)
	pointer.Rect(image.Rectangle{
		Max: gtx.Constraints.Max,
	}).Add(gtx.Ops)
	pointer.InputOp{
		Tag:   n,
		Types: pointer.Enter | pointer.Leave,
	}.Add(gtx.Ops)
	return layout.Inset{
		Top:    unit.Dp(4),
		Bottom: unit.Dp(4),
		Left:   unit.Dp(8),
		Right:  unit.Dp(8),
	}.Layout(gtx, func(gtx C) D {
		return material.Clickable(gtx, &n.Clickable, func(gtx C) D {
			return layout.Stack{}.Layout(gtx,
				layout.Expanded(n.layoutBackground),
				layout.Stacked(n.layoutContent),
			)
		})
	})
}

func (n *renderNavItem) layoutContent(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	contentColor := n.Theme.Color.Text
	if n.selected {
		contentColor = n.Theme.Color.Primary
	}
	return layout.Inset{
		Left:  unit.Dp(8),
		Right: unit.Dp(8),
	}.Layout(gtx, func(gtx C) D {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				if n.NavItem.Icon == nil {
					return layout.Dimensions{}
				}
				return layout.Inset{Right: unit.Dp(40)}.Layout(gtx,
					func(gtx C) D {
						n.NavItem.Icon.Color = contentColor
						return n.NavItem.Icon.Layout(gtx, unit.Dp(24))
					})
			}),
			layout.Rigid(func(gtx C) D {
				label := material.Label(n.Theme, unit.Dp(14), n.Name)
				label.Color = contentColor
				label.Font.Weight = text.Bold
				return layout.Center.Layout(gtx, TruncatingLabelStyle(label).Layout)
			}),
		)
	})
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
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

// ModalNavDrawer implements the Material Design Modal Navigation Drawer
// described here: https://material.io/components/navigation-drawer
type ModalNavDrawer struct {
	*material.Theme

	Title    string
	Subtitle string
	// MaxWidth constrains the maximum amount of horizontal screen real-estate
	// covered by the drawer. If the screen is narrower than this value, the
	// width will be inferred by reserving space for the scrim and using the
	// leftover area for the drawer. Values between 200 and 400 Dp are recommended.
	//
	// The default value used by NewModalNav is 400 Dp.
	MaxWidth unit.Value

	Modal *ModalLayer

	// Anchor indicates whether content in the nav drawer should be anchored to
	// the upper or lower edge of the drawer. This value should match the anchor
	// of an app bar if an app bar is used in conjunction with this nav drawer.
	Anchor VerticalAnchorPosition

	selectedItem    int
	selectedChanged bool // selected item changed during the last frame
	items           []renderNavItem

	navList layout.List
	drag    gesture.Drag

	// animation state
	dragging    bool
	dragStarted f32.Point
	dragOffset  float32
}

// NewModalNav configures a modal navigation drawer that will render itself into the provided ModalLayer
func NewModalNav(th *material.Theme, modal *ModalLayer, title, subtitle string) *ModalNavDrawer {
	m := &ModalNavDrawer{
		Theme:    th,
		Title:    title,
		Subtitle: subtitle,
		MaxWidth: unit.Dp(400),
		Modal:    modal,
	}
	return m
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

// ConfigureModal prepares the modal layer to draw this navigation drawer.
func (m *ModalNavDrawer) ConfigureModal() {
	m.Modal.Widget = func(gtx C, anim *VisibilityAnimation) D {
		m.selectedChanged = false
		m.updateDragState(gtx, anim)
		if !anim.Visible() {
			return layout.Dimensions{}
		}
		for _, event := range m.drag.Events(gtx.Metric, gtx.Queue, gesture.Horizontal) {
			switch event.Type {
			case pointer.Press:
				m.dragStarted = event.Position
				m.dragOffset = 0
				m.dragging = true
			case pointer.Drag:
				newOffset := m.dragStarted.X - event.Position.X
				if newOffset > m.dragOffset {
					m.dragOffset = newOffset
				}
			case pointer.Release:
				fallthrough
			case pointer.Cancel:
				m.dragging = false
			}
		}
		if m.dragOffset != 0 || anim.Animating() {
			defer op.Push(gtx.Ops).Pop()
			m.drawerTransform(gtx, anim).Add(gtx.Ops)
			op.InvalidateOp{}.Add(gtx.Ops)
		}
		return m.layoutSheet(gtx)
	}
}

// updateDragState checks for drawer animations that have finished and
// updates the drawerState accordingly.
func (m *ModalNavDrawer) updateDragState(gtx layout.Context, anim *VisibilityAnimation) {
	if m.dragOffset != 0 && !m.dragging && !anim.Animating() {
		if m.dragOffset < 2 {
			m.dragOffset = 0
		} else {
			m.dragOffset /= 2
		}
	} else if m.dragging && int(m.dragOffset) > gtx.Constraints.Max.X/10 {
		anim.Disappear(gtx.Now)
	}
}

// drawerTransform returns the TransformOp that should be used for the current
// animation frame.
func (m *ModalNavDrawer) drawerTransform(gtx layout.Context, anim *VisibilityAnimation) op.TransformOp {
	revealed := -1 + anim.Revealed(gtx)
	finalOffset := revealed*(float32(m.sheetWidth(gtx))) - m.dragOffset
	return op.Offset(f32.Point{X: finalOffset})
}

func (m ModalNavDrawer) sheetWidth(gtx layout.Context) int {
	scrimWidth := gtx.Px(unit.Dp(56))
	withScrim := gtx.Constraints.Max.X - scrimWidth
	max := gtx.Px(m.MaxWidth)
	return min(withScrim, max)
}

func (m *ModalNavDrawer) layoutSheet(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Max.X = m.sheetWidth(gtx)
	pointer.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Add(gtx.Ops)
	m.drag.Add(gtx.Ops)
	paintRect(gtx, gtx.Constraints.Max, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	spacing := layout.SpaceEnd
	if m.Anchor == Bottom {
		spacing = layout.SpaceStart
	}

	layout.Flex{
		Spacing: spacing,
		Axis:    layout.Vertical,
	}.Layout(gtx,
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
	gtx.Constraints.Min.Y = 0
	m.navList.Axis = layout.Vertical
	return m.navList.Layout(gtx, len(m.items), func(gtx C, index int) D {
		gtx.Constraints.Max.Y = gtx.Px(unit.Dp(48))
		gtx.Constraints.Min = gtx.Constraints.Max
		dimensions := m.items[index].Layout(gtx)
		if m.items[index].Clicked() {
			m.changeSelected(index)
			m.ToggleVisibility(gtx.Now)
			m.selectedChanged = true
		}
		return dimensions
	})
}

// ToggleVisibility changes the state of the nav drawer from retracted to
// extended or visa versa.
func (m *ModalNavDrawer) ToggleVisibility(when time.Time) {
	m.ConfigureModal()
	if !m.Modal.Visible() {
		m.Modal.Appear(when)
	} else {
		m.Modal.Disappear(when)
	}
}

func (m *ModalNavDrawer) changeSelected(newIndex int) {
	m.items[m.selectedItem].selected = false
	m.selectedItem = newIndex
	m.items[m.selectedItem].selected = true
}

// SetNavDestination changes the selected navigation item to the item with
// the provided tag. If the provided tag does not exist, it has no effect.
func (m *ModalNavDrawer) SetNavDestination(tag interface{}) {
	for i, item := range m.items {
		if item.Tag == tag {
			m.changeSelected(i)
			break
		}
	}
}

// CurrentNavDestination returns the tag of the navigation destination
// selected in the drawer.
func (m *ModalNavDrawer) CurrentNavDestination() interface{} {
	return m.items[m.selectedItem].Tag
}

// NavDestinationChanged returns whether the selected navigation destination
// has changed since the last frame.
func (m *ModalNavDrawer) NavDestinationChanged() bool {
	return m.selectedChanged
}

func paintRect(gtx layout.Context, size image.Point, fill color.RGBA) {
	paint.ColorOp{Color: fill}.Add(gtx.Ops)
	paint.PaintOp{
		Rect: f32.Rectangle{
			Max: layout.FPt(size),
		},
	}.Add(gtx.Ops)
}
