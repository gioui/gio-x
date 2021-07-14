// SPDX-License-Identifier: Unlicense OR MIT

package component

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
)

// ContextArea is a region of the UI that responds to right-clicks
// with a contextual widget. The contextual widget is overlaid
// using an op.DeferOp.
type ContextArea struct {
	position f32.Point
	dims     D
	active   bool
}

// Layout renders the context area and -- if the area is activated by an
// appropriate gesture -- also the provided widget overlaid using an op.DeferOp.
func (r *ContextArea) Layout(gtx C, w layout.Widget) D {
	suppressionTag := &r.active
	dismissTag := &r.dims

	startedActive := r.active
	// Summon the contextual widget if the area recieved a secondary click.
	for _, e := range gtx.Events(r) {
		e, ok := e.(pointer.Event)
		if !ok {
			continue
		}
		if r.active {
			// Check whether we should dismiss menu.
			if e.Buttons.Contain(pointer.ButtonPrimary) {
				clickPos := e.Position.Sub(r.position)
				if !clickPos.In(f32.Rectangle{Max: layout.FPt(r.dims.Size)}) {
					r.Dismiss()
				}
			}
		}
		if e.Buttons.Contain(pointer.ButtonSecondary) && e.Type == pointer.Press {
			r.active = true
			r.position = e.Position
		}
	}

	// Dismiss the contextual widget if the user clicked outside of it.
	for _, e := range gtx.Events(suppressionTag) {
		e, ok := e.(pointer.Event)
		if !ok {
			continue
		}
		if e.Type == pointer.Press {
			r.Dismiss()
		}
	}
	// Dismiss the contextual widget if the user released a click within it.
	for _, e := range gtx.Events(dismissTag) {
		e, ok := e.(pointer.Event)
		if !ok {
			continue
		}
		if e.Type == pointer.Release {
			r.Dismiss()
		}
	}

	dims := D{Size: gtx.Constraints.Min}

	var contextual op.CallOp
	if r.active || startedActive {
		// Render if the layout started as active to ensure that widgets
		// within the contextual content get to update their state in reponse
		// to the event that dismissed the contextual widget.
		contextual = func() op.CallOp {
			defer op.Save(gtx.Ops).Load()
			macro := op.Record(gtx.Ops)
			r.dims = w(gtx)
			return macro.Stop()
		}()
	}

	if r.active {
		if int(r.position.X)+r.dims.Size.X > gtx.Constraints.Max.X {
			r.position.X = float32(gtx.Constraints.Max.X - r.dims.Size.X)
		}
		if int(r.position.Y)+r.dims.Size.Y > gtx.Constraints.Max.Y {
			r.position.Y = float32(gtx.Constraints.Max.Y - r.dims.Size.Y)
		}
		// Lay out a transparent scrim to block input to things beneath the
		// contextual widget.
		suppressionScrim := func() op.CallOp {
			defer op.Save(gtx.Ops).Load()
			macro2 := op.Record(gtx.Ops)
			pointer.PassOp{Pass: false}.Add(gtx.Ops)
			pointer.Rect(image.Rectangle{Min: image.Point{-1e6, -1e6}, Max: image.Point{1e6, 1e6}}).Add(gtx.Ops)
			pointer.InputOp{
				Tag:   suppressionTag,
				Grab:  false,
				Types: pointer.Press,
			}.Add(gtx.Ops)
			return macro2.Stop()
		}()
		op.Defer(gtx.Ops, suppressionScrim)

		// Lay out the contextual widget itself.
		macro := op.Record(gtx.Ops)
		op.Offset(r.position).Add(gtx.Ops)
		contextual.Add(gtx.Ops)

		// Lay out a scrim on top of the contextual widget to detect
		// completed interactions with it (that should dismiss it).
		saved := op.Save(gtx.Ops)
		pointer.PassOp{Pass: true}.Add(gtx.Ops)
		pointer.Rect(image.Rectangle{Max: r.dims.Size}).Add(gtx.Ops)
		pointer.InputOp{
			Tag:   dismissTag,
			Grab:  false,
			Types: pointer.Release,
		}.Add(gtx.Ops)

		saved.Load()
		contextual = macro.Stop()
		op.Defer(gtx.Ops, contextual)
	}

	// Capture pointer events in the contextual area.
	pointer.PassOp{Pass: true}.Add(gtx.Ops)
	pointer.Rect(image.Rectangle{Max: gtx.Constraints.Min}).Add(gtx.Ops)
	pointer.InputOp{
		Tag:   r,
		Grab:  false,
		Types: pointer.Press | pointer.Release,
	}.Add(gtx.Ops)

	return dims
}

// Dismiss sets the ContextArea to not be active.
func (r *ContextArea) Dismiss() {
	r.active = false
}

// Active returns whether the ContextArea is currently active (whether
// it is currently displaying overlaid content or not).
func (r ContextArea) Active() bool {
	return r.active
}
