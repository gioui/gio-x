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
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var (
	hoverOverlayAlpha    uint8 = 25
	selectedOverlayAlpha uint8 = 50
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
	*AlphaPalette
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
	gtx.Constraints.Min = gtx.Constraints.Max
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
	var fill color.NRGBA
	if n.hovering {
		fill = AlphaMultiply(n.Theme.Color.Text, n.AlphaPalette.Hover)
	} else if n.selected {
		fill = AlphaMultiply(n.Theme.Color.Primary, n.AlphaPalette.Selected)
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

// NavDrawer implements the Material Design Navigation Drawer
// described here: https://material.io/components/navigation-drawer
type NavDrawer struct {
	*material.Theme

	// Background (if set) will be passed down to the underlying sheet when the
	// drawer is rendered.
	Background *color.NRGBA
	AlphaPalette

	Title    string
	Subtitle string

	// Anchor indicates whether content in the nav drawer should be anchored to
	// the upper or lower edge of the drawer. This value should match the anchor
	// of an app bar if an app bar is used in conjunction with this nav drawer.
	Anchor VerticalAnchorPosition

	selectedItem    int
	selectedChanged bool // selected item changed during the last frame
	items           []renderNavItem

	navList layout.List
}

// NewNav configures a navigation drawer
func NewNav(th *material.Theme, title, subtitle string) NavDrawer {
	m := NavDrawer{
		Theme:    th,
		Title:    title,
		Subtitle: subtitle,
		AlphaPalette: AlphaPalette{
			Hover:    hoverOverlayAlpha,
			Selected: selectedOverlayAlpha,
		},
	}
	return m
}

// AddNavItem inserts a navigation target into the drawer. This should be
// invoked only from the layout thread to avoid nasty race conditions.
func (m *NavDrawer) AddNavItem(item NavItem) {
	m.items = append(m.items, renderNavItem{
		Theme:        m.Theme,
		NavItem:      item,
		AlphaPalette: &m.AlphaPalette,
	})
	if len(m.items) == 1 {
		m.items[0].selected = true
	}
}

func (m *NavDrawer) Layout(gtx layout.Context, anim *VisibilityAnimation) layout.Dimensions {
	sheet := NewSheet()
	if m.Background != nil {
		sheet.Background = *m.Background
	}
	return sheet.Layout(gtx, anim, func(gtx C) D {
		return m.LayoutContents(gtx, anim)
	})
}

func (m *NavDrawer) LayoutContents(gtx layout.Context, anim *VisibilityAnimation) layout.Dimensions {
	if !anim.Visible() {
		return D{}
	}
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
		layout.Flexed(1, func(gtx C) D {
			return m.layoutNavList(gtx, anim)
		}),
	)
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func (m *NavDrawer) layoutNavList(gtx layout.Context, anim *VisibilityAnimation) layout.Dimensions {
	m.selectedChanged = false
	gtx.Constraints.Min.Y = 0
	m.navList.Axis = layout.Vertical
	return m.navList.Layout(gtx, len(m.items), func(gtx C, index int) D {
		gtx.Constraints.Max.Y = gtx.Px(unit.Dp(48))
		gtx.Constraints.Min = gtx.Constraints.Max
		m.items[index].Theme = m.Theme
		dimensions := m.items[index].Layout(gtx)
		if m.items[index].Clicked() {
			m.changeSelected(index)
		}
		return dimensions
	})
}

func (m *NavDrawer) changeSelected(newIndex int) {
	if newIndex == m.selectedItem {
		return
	}
	m.items[m.selectedItem].selected = false
	m.selectedItem = newIndex
	m.items[m.selectedItem].selected = true
	m.selectedChanged = true
}

// SetNavDestination changes the selected navigation item to the item with
// the provided tag. If the provided tag does not exist, it has no effect.
func (m *NavDrawer) SetNavDestination(tag interface{}) {
	for i, item := range m.items {
		if item.Tag == tag {
			m.changeSelected(i)
			break
		}
	}
}

// CurrentNavDestination returns the tag of the navigation destination
// selected in the drawer.
func (m *NavDrawer) CurrentNavDestination() interface{} {
	return m.items[m.selectedItem].Tag
}

// NavDestinationChanged returns whether the selected navigation destination
// has changed since the last frame.
func (m *NavDrawer) NavDestinationChanged() bool {
	return m.selectedChanged
}

// ModalNavDrawer implements the Material Design Modal Navigation Drawer
// described here: https://material.io/components/navigation-drawer
type ModalNavDrawer struct {
	*NavDrawer
	sheet *ModalSheet
}

// NewModalNav configures a modal navigation drawer that will render itself into the provided ModalLayer
func NewModalNav(th *material.Theme, modal *ModalLayer, title, subtitle string) *ModalNavDrawer {
	nav := NewNav(th, title, subtitle)
	return ModalNavFrom(&nav, modal)
}

func ModalNavFrom(nav *NavDrawer, modal *ModalLayer) *ModalNavDrawer {
	m := &ModalNavDrawer{}
	modalSheet := NewModalSheet(modal)
	m.NavDrawer = nav
	m.sheet = modalSheet
	if nav.Background != nil {
		m.sheet.Sheet.Background = *nav.Background
	}
	return m
}

func (m *ModalNavDrawer) Layout() layout.Dimensions {
	m.sheet.LayoutModal(func(gtx C, anim *VisibilityAnimation) D {
		dims := m.NavDrawer.LayoutContents(gtx, anim)
		if m.selectedChanged {
			anim.Disappear(gtx.Now)
		}
		return dims
	})
	return D{}
}

func (m *ModalNavDrawer) ToggleVisibility(when time.Time) {
	m.Layout()
	m.sheet.ToggleVisibility(when)
}

func (m *ModalNavDrawer) Appear(when time.Time) {
	m.Layout()
	m.sheet.Appear(when)
}

func (m *ModalNavDrawer) Disappear(when time.Time) {
	m.Layout()
	m.sheet.Disappear(when)
}

func paintRect(gtx layout.Context, size image.Point, fill color.NRGBA) {
	Rect{
		Color: fill,
		Size:  size,
	}.Layout(gtx)
}
