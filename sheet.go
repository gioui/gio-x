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
	"gioui.org/unit"
)

// Sheet implements an unanimated sheet's state.
//
// TODO(whereswaldon): animate the appearance of this type, possibly by
// wrapping it in anther type.
type Sheet struct {
	Background color.RGBA
}

// NewSheet returns a sheet with its background color initialized to white.
func NewSheet() Sheet {
	return Sheet{
		Background: color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
	}
}

// Layout renders the provided widget on a background. The background will use
// the maximum space available.
func (s Sheet) Layout(gtx layout.Context, widget layout.Widget) layout.Dimensions {
	// lay out background
	paintRect(gtx, gtx.Constraints.Max, s.Background)

	// lay out sheet contents
	return widget(gtx)
}

// ModalSheet implements the Modal Side Sheet component
// specified at https://material.io/components/sheets-side#modal-side-sheet
type ModalSheet struct {
	// MaxWidth constrains the maximum amount of horizontal screen real-estate
	// covered by the drawer. If the screen is narrower than this value, the
	// width will be inferred by reserving space for the scrim and using the
	// leftover area for the drawer. Values between 200 and 400 Dp are recommended.
	//
	// The default value used by NewModalNav is 400 Dp.
	MaxWidth unit.Value

	Modal *ModalLayer

	drag gesture.Drag

	// animation state
	dragging    bool
	dragStarted f32.Point
	dragOffset  float32

	// Sheet is the sheet upon which the contents of the modal sheet will be laid out.
	Sheet

	layout.Widget
}

// NewModalSheet creates a modal sheet that will render the provided
// widget within the provided modal layer when it is made visible.
func NewModalSheet(m *ModalLayer, widget layout.Widget) *ModalSheet {
	s := &ModalSheet{
		MaxWidth: unit.Dp(400),
		Modal:    m,
		Sheet:    NewSheet(),
		Widget:   widget,
	}
	return s
}

// updateDragState ensures that a partially-dragged sheet
// snaps back into place when released and otherwise chooses
// when the sheet has been dragged far enough to close.
func (s *ModalSheet) updateDragState(gtx layout.Context, anim *VisibilityAnimation) {
	if s.dragOffset != 0 && !s.dragging && !anim.Animating() {
		if s.dragOffset < 2 {
			s.dragOffset = 0
		} else {
			s.dragOffset /= 2
		}
	} else if s.dragging && int(s.dragOffset) > gtx.Constraints.Max.X/10 {
		anim.Disappear(gtx.Now)
	}
}

// ConfigureModal requests that the sheet prepare the associated
// ModalLayer to render itself (rather than another modal widget).
func (s *ModalSheet) ConfigureModal() {
	s.Modal.Widget = func(gtx C, anim *VisibilityAnimation) D {
		s.updateDragState(gtx, anim)
		if !anim.Visible() {
			return D{}
		}
		for _, event := range s.drag.Events(gtx.Metric, gtx.Queue, gesture.Horizontal) {
			switch event.Type {
			case pointer.Press:
				s.dragStarted = event.Position
				s.dragOffset = 0
				s.dragging = true
			case pointer.Drag:
				newOffset := s.dragStarted.X - event.Position.X
				if newOffset > s.dragOffset {
					s.dragOffset = newOffset
				}
			case pointer.Release:
				fallthrough
			case pointer.Cancel:
				s.dragging = false
			}
		}
		if s.dragOffset != 0 || anim.Animating() {
			defer op.Push(gtx.Ops).Pop()
			s.drawerTransform(gtx, anim).Add(gtx.Ops)
			op.InvalidateOp{}.Add(gtx.Ops)
		}
		gtx.Constraints.Max.X = s.sheetWidth(gtx)

		// lay out widget
		dims := s.Sheet.Layout(gtx, s.Widget)

		// listen for drag events
		pointer.PassOp{Pass: true}.Add(gtx.Ops)
		pointer.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Add(gtx.Ops)
		s.drag.Add(gtx.Ops)

		return dims
	}
}

// drawerTransform returns the current offset transformation
// of the sheet taking both drag and animation progress
// into account.
func (s ModalSheet) drawerTransform(gtx C, anim *VisibilityAnimation) op.TransformOp {
	revealed := -1 + anim.Revealed(gtx)
	finalOffset := revealed*(float32(s.sheetWidth(gtx))) - s.dragOffset
	return op.Offset(f32.Point{X: finalOffset})
}

// sheetWidth returns the width of the sheet taking both the dimensions
// of the modal layer and the MaxWidth field into account.
func (s ModalSheet) sheetWidth(gtx layout.Context) int {
	scrimWidth := gtx.Px(unit.Dp(56))
	withScrim := gtx.Constraints.Max.X - scrimWidth
	max := gtx.Px(s.MaxWidth)
	return min(withScrim, max)
}

// ToggleVisibility triggers the appearance or disappearance of the
// ModalSheet. It automatically calls ConfigureModal().
func (s *ModalSheet) ToggleVisibility(when time.Time) {
	s.ConfigureModal()
	if !s.Modal.Visible() {
		s.Modal.Appear(when)
	} else {
		s.Modal.Disappear(when)
	}
}
