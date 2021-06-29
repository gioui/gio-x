package scroll

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

// ScrollbarState holds the persistent state for an area that can
// display a scrollbar.
type ScrollbarState struct {
	track, indicator gesture.Click
	drag             gesture.Drag
	ScrollPosition
}

// ScrollTrackStyle configures the presentation of a track for a scroll area.
type ScrollTrackStyle struct {
	// Padding along the major and minor axis of the scrollbar's
	// track. This is used to keep the scrollbar from touching the
	// edges of the content area.
	MajorPadding, MinorPadding unit.Value
}

// ScrollIndicatorStyle configures the presentation of a scroll indicator.
type ScrollIndicatorStyle struct {
	// The smallest that the scroll indicator is allowed to be along
	// the major axis.
	MajorMinLen unit.Value
	// The width of the scroll indicator across the minor axis.
	MinorWidth unit.Value
	// The normal and hovered colors of the scroll indicator.
	Color, HoverColor color.NRGBA
	// The corner radius of the rectangular indicator. 0 will produce
	// square corners. 0.5*MinorWidth will produce perfectly round
	// corners.
	CornerRadius unit.Value
}

// ScrollPosition describes the position of a scrollable viewport atop
// a finite scrollable region.
type ScrollPosition struct {
	// VisibleStart is the start position of the viewport within the
	// scrollabe region represented as a fraction. It should be in the
	// range [0,1]
	VisibleStart float32
	// VisibleEnd is the end position of the viewport within the scrollable
	// region represented as a fraction. It should be in the range [0,1]
	VisibleEnd float32
}

// Shift moves the scroll position by up to the delta value, keeping the
// range described by VisibleStart and VisibleEnd within the range [0,1]
// and preserving the distance between VisibleStart and VisibleEnd.
func (s ScrollPosition) Shift(delta float32) ScrollPosition {
	interval := s.VisibleEnd - s.VisibleStart
	s.VisibleStart += delta
	s.VisibleEnd += delta

	// Keep the scrollbar within the track and prevent it from
	// distorting against the ends of the track.
	s.VisibleStart = clamp(s.VisibleStart)
	if s.VisibleStart == 0 {
		s.VisibleEnd = interval
	}
	s.VisibleEnd = clamp(s.VisibleEnd)
	if s.VisibleEnd == 1 {
		s.VisibleStart = 1 - interval
	}
	return s
}

// ScrollbarStyle configures the presentation of a scrollbar.
type ScrollbarStyle struct {
	Axis      layout.Axis
	State     *ScrollbarState
	Track     ScrollTrackStyle
	Indicator ScrollIndicatorStyle
}

// Scrollbar configures the presentation of a scrollbar using the provided
// theme, state, and positional information.
func Scrollbar(th *material.Theme, state *ScrollbarState, pos ScrollPosition) ScrollbarStyle {
	state.ScrollPosition = pos
	lightFg := th.Palette.Fg
	lightFg.A = 150
	darkFg := lightFg
	darkFg.A = 200

	return ScrollbarStyle{
		State: state,
		Track: ScrollTrackStyle{
			MajorPadding: unit.Dp(2),
			MinorPadding: unit.Dp(2),
		},
		Indicator: ScrollIndicatorStyle{
			MajorMinLen:  unit.Dp(8),
			MinorWidth:   unit.Dp(6),
			CornerRadius: unit.Dp(3),
			Color:        lightFg,
			HoverColor:   darkFg,
		},
	}
}

// Layout renders the scrollbar anchored to the appropriate edge of the
// provided context.
func (s ScrollbarStyle) Layout(gtx layout.Context) layout.Dimensions {
	anchoring := func() layout.Direction {
		if s.Axis == layout.Horizontal {
			return layout.S
		}
		return layout.E
	}()
	return anchoring.Layout(gtx, func(gtx C) D {
		// Convert constraints to an axis-independent form.
		gtx.Constraints.Max = s.Axis.Convert(gtx.Constraints.Max)
		gtx.Constraints.Min = s.Axis.Convert(gtx.Constraints.Min)
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		gtx.Constraints.Min.Y = gtx.Px(unit.Add(gtx.Metric, s.Indicator.MinorWidth, s.Track.MinorPadding))

		trackHeight := float32(gtx.Constraints.Max.X)
		delta := float32(0)

		// Now that we know the dimensions for the scrollbar track, process events
		// that may have modified the indicator position.

		// Jump to a click in the track.
		for _, event := range s.State.track.Events(gtx) {
			if event.Type != gesture.TypeClick ||
				event.Modifiers != key.Modifiers(0) ||
				event.NumClicks > 1 {
				continue
			}
			pos := s.Axis.Convert(image.Point{
				X: int(event.Position.X),
				Y: int(event.Position.Y),
			})
			normalizedPos := float32(pos.X) / trackHeight
			delta += normalizedPos - s.State.VisibleStart
		}

		// Offset to account for any drags.
		for _, event := range s.State.drag.Events(gtx.Metric, gtx, gesture.Axis(s.Axis)) {
			if event.Type != pointer.Drag {
				continue
			}
			dragOffset := s.Axis.FConvert(event.Position).X
			normalizedDragOffset := (dragOffset / trackHeight)
			delta += (normalizedDragOffset - s.State.VisibleStart) * .5

		}

		// Darken indicator if hovered.
		if _ = s.State.indicator.Events(gtx); s.State.indicator.Hovered() {
			s.Indicator.Color = s.Indicator.HoverColor
		}

		// Actually shift the list in response to drags or clicks.
		if delta != 0 {
			s.State.ScrollPosition = s.State.ScrollPosition.Shift(delta)
			op.InvalidateOp{}.Add(gtx.Ops)
		}

		return s.layout(gtx)
	})
}

// clamp ensures that the input value is within the range [0,1], and
// returns either the input value or the end of the range that the
// input is closest to.
func clamp(in float32) float32 {
	if in < 0 {
		return 0
	}
	if in > 1 {
		return 1
	}
	return in
}

// layout lays out the scroll track and indicator under the assumption
// that the current gtx is converted to (main,cross) coordinates for
// the scrollbar's axis.
func (s ScrollbarStyle) layout(gtx C) D {
	inset := layout.Inset{
		Top:    s.Track.MajorPadding,
		Bottom: s.Track.MajorPadding,
		Left:   s.Track.MinorPadding,
		Right:  s.Track.MinorPadding,
	}

	return inset.Layout(gtx, func(gtx C) D {
		// Lay out the clickable track underneath the scroll indicator.
		area := s.Axis.Convert(gtx.Constraints.Min)
		pointer.Rect(image.Rectangle{
			Max: area,
		}).Add(gtx.Ops)
		s.State.track.Add(gtx.Ops)

		// Compute the pixel size and position of the scroll indicator within
		// the track.
		trackLen := float32(gtx.Constraints.Min.X)
		viewStart := s.State.VisibleStart * trackLen
		viewEnd := s.State.VisibleEnd * trackLen
		indicatorLen := unit.Max(gtx.Metric, unit.Px(viewEnd-viewStart), s.Indicator.MajorMinLen)
		indicatorDims := s.Axis.Convert(image.Point{
			X: gtx.Px(indicatorLen),
			Y: gtx.Px(s.Indicator.MinorWidth),
		})
		indicatorDimsF := layout.FPt(indicatorDims)
		radius := float32(gtx.Px(s.Indicator.CornerRadius))

		// Lay out the indicator.
		offset := s.Axis.Convert(image.Pt(int(viewStart), 0))
		defer op.Save(gtx.Ops).Load()
		op.Offset(layout.FPt(offset)).Add(gtx.Ops)
		paint.FillShape(gtx.Ops, s.Indicator.Color, clip.RRect{
			Rect: f32.Rectangle{
				Max: indicatorDimsF,
			},
			SW: radius,
			NW: radius,
			NE: radius,
			SE: radius,
		}.Op(gtx.Ops))

		// Add the indicator pointer hit areas.
		pointer.Rect(image.Rectangle{Max: indicatorDims}).Add(gtx.Ops)
		s.State.drag.Add(gtx.Ops)
		s.State.indicator.Add(gtx.Ops)

		return D{Size: area}
	})
}

// ListState holds the persistent state for a layout.List that has a
// scrollbar attached.
type ListState struct {
	ScrollbarState
	layout.List
}

// ListStyle configures the presentation of a layout.List with a scrollbar.
type ListStyle struct {
	state *ListState
	ScrollbarStyle
}

// List constructs a ListStyle using the provided theme and state.
func List(th *material.Theme, state *ListState) ListStyle {
	return ListStyle{
		state:          state,
		ScrollbarStyle: Scrollbar(th, &state.ScrollbarState, state.ScrollPosition),
	}
}

// Layout renders the list and its scrollbar.
func (l ListStyle) Layout(gtx layout.Context, length int, w func(gtx layout.Context, index int) layout.Dimensions) layout.Dimensions {
	// Ensure that the scrolling axis is synchronized, using the layout.List
	// as the source of truth.
	l.ScrollbarStyle.Axis = l.state.List.Axis

	longAxisSum := 0
	var meanElementHeight float32

	// Lay out the list elements and track their dimensions.
	listDims := l.state.List.Layout(gtx, length, func(gtx C, index int) D {
		elementDims := w(gtx, index)
		longAxisSum += l.Axis.Convert(elementDims.Size).X
		return elementDims
	})

	listOffsetF := float32(l.state.List.Position.Offset)

	// Approximate the size of the scrollable content.
	visibleCount := float32(l.state.List.Position.Count)
	meanElementHeight = float32(longAxisSum) / visibleCount

	// Determine how much of the content is visible.
	lengthPx := meanElementHeight * float32(length)
	visiblePx := visibleCount*meanElementHeight - listOffsetF + float32(l.state.List.Position.OffsetLast)
	visibleFraction := visiblePx / lengthPx

	// Compute the location of the beginning of the viewport.
	viewportStart := (float32(l.state.List.Position.First)*meanElementHeight + listOffsetF) / lengthPx
	l.state.ScrollPosition.VisibleStart = viewportStart
	l.state.ScrollPosition.VisibleEnd = viewportStart + visibleFraction

	// Render the scrollbar.
	defer op.Save(gtx.Ops).Load()
	l.ScrollbarStyle.Layout(gtx)

	// Handle any changes to the scroll position as a result of user interaction.
	scrollPos := l.state.ScrollPosition
	totalPx := meanElementHeight * float32(length)
	offsetPx := (totalPx * scrollPos.VisibleStart)
	var offset float32
	if meanElementHeight > 0 {
		offset = offsetPx / meanElementHeight
	} else {
		offset = 0
	}
	l.state.List.Position.First = int(offset)
	l.state.List.Position.Offset = int((offset - float32(int(offset))) * meanElementHeight)

	return listDims
}
