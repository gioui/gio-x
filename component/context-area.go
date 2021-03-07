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
	pointer.Rect(image.Rectangle{Max: gtx.Constraints.Min}).Add(gtx.Ops)
	pointer.PassOp{Pass: true}.Add(gtx.Ops)
	pointer.InputOp{
		Tag:   r,
		Grab:  false,
		Types: pointer.Press | pointer.Release,
	}.Add(gtx.Ops)
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
		if e.Buttons.Contain(pointer.ButtonSecondary) {
			r.active = true
			r.position = e.Position
		}
	}
	dims := D{Size: gtx.Constraints.Min}

	if !r.active {
		return dims
	}

	for _, e := range gtx.Events(&r.active) {
		e, ok := e.(pointer.Event)
		if !ok {
			continue
		}
		if e.Type == pointer.Release {
			r.Dismiss()
		}
	}

	defer op.Save(gtx.Ops).Load()
	macro := op.Record(gtx.Ops)
	r.dims = w(gtx)
	call := macro.Stop()

	if int(r.position.X)+r.dims.Size.X > gtx.Constraints.Max.X {
		r.position.X = float32(gtx.Constraints.Max.X - r.dims.Size.X)
	}
	if int(r.position.Y)+r.dims.Size.Y > gtx.Constraints.Max.Y {
		r.position.Y = float32(gtx.Constraints.Max.Y - r.dims.Size.Y)
	}
	macro2 := op.Record(gtx.Ops)
	op.Offset(r.position).Add(gtx.Ops)
	call.Add(gtx.Ops)
	pointer.PassOp{Pass: true}.Add(gtx.Ops)
	pointer.Rect(image.Rectangle{Min: image.Point{-1e6, -1e6}, Max: image.Point{1e6, 1e6}}).Add(gtx.Ops)
	pointer.InputOp{
		Tag:   &r.active,
		Grab:  false,
		Types: pointer.Release,
	}.Add(gtx.Ops)
	call2 := macro2.Stop()
	op.Defer(gtx.Ops, call2)
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
