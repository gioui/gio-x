package outlay

import (
	"fmt"
	"log"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"

	"git.sr.ht/~whereswaldon/sprig/anim"
)

type Fan struct {
	itemsCache        []cacheItem
	last              fanParams
	animatedLastFrame bool
	anim.Normal
}

type fanParams struct {
	arc    float32
	radius float32
	len    int
}

func (f fanParams) String() string {
	return fmt.Sprintf("arc: %v radus: %v len: %v", f.arc, f.radius, f.len)
}

type cacheItem struct {
	elevated bool
	op.CallOp
	layout.Dimensions
}

type FanItem struct {
	W       layout.Widget
	Elevate bool
}

func Item(evelvate bool, w layout.Widget) FanItem {
	return FanItem{
		W:       w,
		Elevate: evelvate,
	}
}

func (f *Fan) Layout(gtx layout.Context, items ...FanItem) layout.Dimensions {
	defer op.Push(gtx.Ops).Pop()
	op.Offset(f32.Point{
		X: float32(gtx.Constraints.Max.X / 2),
		Y: float32(gtx.Constraints.Max.Y / 2),
	}).Add(gtx.Ops)
	f.itemsCache = f.itemsCache[:0]
	maxWidth := 0
	for i := range items {
		item := items[i]
		macro := op.Record(gtx.Ops)
		dims := item.W(gtx)
		if dims.Size.X > maxWidth {
			maxWidth = dims.Size.X
		}
		f.itemsCache = append(f.itemsCache, cacheItem{
			CallOp:     macro.Stop(),
			Dimensions: dims,
			elevated:   item.Elevate,
		})
	}
	var current fanParams
	current.len = len(items)
	current.radius = float32(maxWidth * 2.0)
	var itemArcFraction float32
	if len(items) > 1 {
		itemArcFraction = float32(1) / float32(len(items)-1)
	} else {
		itemArcFraction = 1
	}
	current.arc = math.Pi / 2 * itemArcFraction

	var empty fanParams
	if f.last == empty {
		f.last = current
	} else if f.last != current {

		if !f.animatedLastFrame {
			f.Start(gtx.Now)
		}
		progress := f.Progress(gtx)
		if f.animatedLastFrame && progress >= 1 {
			f.last = current
		}
		f.animatedLastFrame = false
		if f.Animating(gtx) {
			f.animatedLastFrame = true
			op.InvalidateOp{}.Add(gtx.Ops)
		}
		current.arc = f.last.arc - (f.last.arc-current.arc)*progress
		current.radius = f.last.radius - (f.last.radius-current.radius)*progress
		log.Println(progress, current)
	}

	visible := f.itemsCache[:min(f.last.len, current.len)]
	for i := range visible {
		if !f.itemsCache[i].elevated {
			f.layoutItem(gtx, i, current)
		}
	}
	for i := range visible {
		if f.itemsCache[i].elevated {
			f.layoutItem(gtx, i, current)
		}
	}
	return layout.Dimensions{
		Size: gtx.Constraints.Max,
	}

}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (f *Fan) layoutItem(gtx layout.Context, index int, params fanParams) layout.Dimensions {
	defer op.Push(gtx.Ops).Pop()
	arc := params.arc
	radius := params.radius
	if len(f.itemsCache) > 1 {
		arc = arc*float32(index) + math.Pi/4
	} else {
		arc = math.Pi / 2
	}
	var transform f32.Affine2D
	transform = transform.Rotate(f32.Point{}, -math.Pi/2).
		Offset(f32.Pt(-radius, float32(f.itemsCache[index].Dimensions.Size.X/2))).
		Rotate(f32.Point{}, arc)
	op.Affine(transform).Add(gtx.Ops)
	f.itemsCache[index].Add(gtx.Ops)
	return layout.Dimensions{}
}
