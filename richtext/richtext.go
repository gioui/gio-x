package richtext

import (
	"image"
	"image/color"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"golang.org/x/image/math/fixed"
)

//TextObjects represents the whole richtext widget
type TextObjects []*TextObject

//TextObject is one of the objects in the richtext widget
type TextObject struct {
	Font      text.Font
	Size      unit.Value
	Color     color.NRGBA
	Content   string
	Clickable bool
	metadata  map[string]string

	origin     *TextObject
	textLayout text.Layout
	xOffset    int
	size       image.Point
	clicked    int
	hovered    bool
}

//Layout prints out the text at specified offset
func (to *TextObject) Layout(gtx layout.Context, s text.Shaper, off image.Point) layout.Dimensions {
	stack := op.Save(gtx.Ops)
	paint.ColorOp{Color: to.Color}.Add(gtx.Ops)
	op.Offset(layout.FPt(off)).Add(gtx.Ops)
	s.Shape(to.Font, fixed.I(gtx.Px(to.Size)), to.textLayout).Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	stack.Load()
	return layout.Dimensions{Size: to.size}
}

func (to *TextObject) Clicked() bool {
	if to.clicked <= 0 {
		return false
	}
	to.clicked--
	return true
}

func (to *TextObject) Hovered() bool {
	return to.hovered
}

func (to *TextObject) updateClicks(gtx layout.Context) {
	if !to.Clickable {
		return
	}
	for _, e := range gtx.Events(origin(to)) {
		switch p := e.(type) {
		case pointer.Event:
			switch p.Type {
			case pointer.Release:
				to.clicked++
			case pointer.Enter:
				to.hovered = true
			case pointer.Leave, pointer.Cancel:
				to.hovered = false
			}
		}
	}
}

func (to *TextObject) SetMetadata(key, value string) {
	if to.metadata == nil {
		to.metadata = make(map[string]string)
	}
	to.metadata[key] = value
}

func (to *TextObject) GetMetadata(key string) string {
	if to.metadata == nil {
		return ""
	}
	return to.metadata[key]
}

func (to *TextObject) DeepCopy() *TextObject {
	t := *to
	t.metadata = make(map[string]string)
	for k, v := range to.metadata {
		t.metadata[k] = v
	}
	return &t
}

//Layout prints out the rich text widget
func (tos TextObjects) Layout(gtx layout.Context, s text.Shaper) layout.Dimensions {
	var tosDims layout.Dimensions
	var lineDims image.Point
	oi := &objectIterator{objs: tos}
	//TextObjects (or it's parts) on a single line
	var lineObjects []*TextObject

	for to := oi.Next(); to != nil; to = oi.Next() {
		maxWidth := gtx.Constraints.Max.X - lineDims.X
		//layout the string using the maxWidthe remaining for the line
		lines := s.LayoutString(to.Font, fixed.I(gtx.Px(to.Size)), maxWidth, to.Content)
		//using just the first line
		first := lines[0]
		lineWidth := first.Width.Ceil()
		lineHeight := (first.Ascent + first.Descent).Ceil()

		//getting the used text in the line and calculating the dimensions
		to.textLayout = first.Layout
		to.xOffset = lineDims.X
		to.size = image.Point{X: lineWidth, Y: lineHeight}
		lineObjects = append(lineObjects, to)

		//calculating the X and the max of the Y of the line
		lineDims.X += lineWidth
		if lineDims.Y < lineHeight {
			lineDims.Y = lineHeight
		}
		//the width of the whole richtext object
		if tosDims.Size.X < lineDims.X {
			tosDims.Size.X = lineDims.X
		}

		//if we are going to break the line, or we are at the end of the richtext
		if len(lines) > 1 || oi.Empty() {
			//we print the line
			for _, obj := range lineObjects {
				off := image.Point{
					X: obj.xOffset,
					Y: tosDims.Size.Y + lineDims.Y,
				}
				obj.Layout(gtx, s, off)
				if !obj.Clickable {
					continue
				}
				to.updateClicks(gtx)
				stack := op.Save(gtx.Ops)
				clickableOffset := image.Point{X: obj.xOffset, Y: tosDims.Size.Y}
				op.Offset(layout.FPt(clickableOffset)).Add(gtx.Ops)
				pointer.Rect(image.Rectangle{Max: obj.size}).Add(gtx.Ops)
				pointer.InputOp{
					Tag:   origin(obj),
					Types: pointer.Release | pointer.Enter | pointer.Leave,
				}.Add(gtx.Ops)
				pointer.CursorNameOp{Name: pointer.CursorPointer}.Add(gtx.Ops)
				stack.Load()
			}
			//we printed these objectss, so we get rid of them
			lineObjects = nil
			tosDims.Size.Y += lineDims.Y
		}
		if len(lines) > 1 {
			//add the rest of the TextObject to be printed on a new line
			//this puts it to the front of the objectIterator
			oi.Add(&TextObject{
				origin:    origin(to),
				Font:      to.Font,
				Size:      to.Size,
				Color:     to.Color,
				Clickable: to.Clickable,
				Content:   to.Content[len(first.Layout.Text):],
			})
			lineDims.X = 0
			lineDims.Y = 0
		}
	}

	return tosDims
}

type objectIterator struct {
	objs []*TextObject
}

func (oi *objectIterator) Add(to *TextObject) {
	oi.objs = append([]*TextObject{to}, oi.objs...)
}

func (oi *objectIterator) Next() *TextObject {
	if len(oi.objs) == 0 {
		return nil
	}
	next := oi.objs[0]
	oi.objs = oi.objs[1:]
	return next
}

func (oi *objectIterator) Empty() bool {
	return len(oi.objs) == 0
}

func (tos TextObjects) Clicked() *TextObject {
	for _, to := range tos {
		if to.Clicked() {
			return to
		}
	}
	return nil
}

func (tos TextObjects) Hovered() *TextObject {
	for _, to := range tos {
		if to.Hovered() {
			return to
		}
	}
	return nil
}

func origin(to *TextObject) *TextObject {
	if to.origin == nil {
		return to
	}
	return to.origin
}
