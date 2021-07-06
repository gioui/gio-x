package colorpicker

import (
	"encoding/hex"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image/color"
	"strconv"
)

func NewHexEditor(th *material.Theme) *HexEditor {
	return &HexEditor{theme: th, hex: newHexField(th, widget.Editor{SingleLine: true})}
}

type HexEditor struct {
	theme *material.Theme
	color color.NRGBA
	hex   *hexField
}

func (e *HexEditor) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceAround}.Layout(gtx,
		layout.Rigid(material.Label(e.theme, unit.Sp(14), "Hex ").Layout),
		layout.Flexed(1, e.hex.Layout))
}

func (e *HexEditor) Color() color.NRGBA {
	return e.color
}

func (e *HexEditor) SetColor(col color.NRGBA) {
	e.color = col
	e.hex.SetHex([]byte{col.R, col.G, col.B})
}

func (e *HexEditor) Changed() bool {
	if !e.hex.Changed() {
		return false
	}
	b := e.hex.Hex()
	if len(b) < 3 {
		return false
	}
	e.color.R, e.color.G, e.color.B = b[0], b[1], b[2]
	return true
}

func NewRgbEditor(th *material.Theme) *RgbEditor {
	return &RgbEditor{theme: th,
		r: &byteField{editor: widget.Editor{SingleLine: true}},
		g: &byteField{editor: widget.Editor{SingleLine: true}},
		b: &byteField{editor: widget.Editor{SingleLine: true}}}
}

type RgbEditor struct {
	theme *material.Theme
	rgb   color.NRGBA
	r     *byteField
	g     *byteField
	b     *byteField
}

func (e *RgbEditor) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceAround, Alignment: layout.Baseline}.Layout(gtx,
		layout.Rigid(material.Label(e.theme, unit.Sp(14), "R ").Layout),
		layout.Flexed(1, material.Editor(e.theme, &e.r.editor, "").Layout),
		layout.Rigid(material.Label(e.theme, unit.Sp(14), "G ").Layout),
		layout.Flexed(1, material.Editor(e.theme, &e.g.editor, "").Layout),
		layout.Rigid(material.Label(e.theme, unit.Sp(14), "B ").Layout),
		layout.Flexed(1, material.Editor(e.theme, &e.b.editor, "").Layout))
}

func (e *RgbEditor) Changed() bool {
	changed := false
	if e.r.Changed() {
		changed = true
		e.rgb.R = e.r.Byte()
	}
	if e.g.Changed() {
		changed = true
		e.rgb.G = e.g.Byte()
	}
	if e.b.Changed() {
		changed = true
		e.rgb.B = e.b.Byte()
	}
	return changed
}

func (e *RgbEditor) SetColor(col color.NRGBA) {
	e.rgb = col
	e.r.SetByte(col.R)
	e.g.SetByte(col.G)
	e.b.SetByte(col.B)
}

func (e *RgbEditor) Color() color.NRGBA {
	return e.rgb
}

func NewHsvEditor(th *material.Theme) *HsvEditor {
	return &HsvEditor{theme: th,
		h: &degreeField{editor: widget.Editor{SingleLine: true}},
		s: &percentageField{editor: widget.Editor{SingleLine: true}},
		v: &percentageField{editor: widget.Editor{SingleLine: true}}}
}

type HsvEditor struct {
	theme *material.Theme
	hsv   HSVColor
	h     *degreeField
	s     *percentageField
	v     *percentageField
}

func (e *HsvEditor) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceAround, Alignment: layout.Baseline}.Layout(gtx,
		layout.Rigid(material.Label(e.theme, unit.Sp(14), "H ").Layout),
		layout.Flexed(1, material.Editor(e.theme, &e.h.editor, "").Layout),
		layout.Rigid(material.Label(e.theme, unit.Sp(14), "S ").Layout),
		layout.Flexed(1, material.Editor(e.theme, &e.s.editor, "").Layout),
		layout.Rigid(material.Label(e.theme, unit.Sp(14), "V ").Layout),
		layout.Flexed(1, material.Editor(e.theme, &e.v.editor, "").Layout))
}

func (e *HsvEditor) Color() color.NRGBA {
	return HsvToRgb(e.hsv)
}

func (e *HsvEditor) Changed() bool {
	changed := false
	if e.h.Changed() {
		changed = true
		e.hsv.H = float32(e.h.Degree())
	}
	if e.s.Changed() {
		changed = true
		e.hsv.S = float32(e.s.Percentage() / 100)
	}
	if e.v.Changed() {
		changed = true
		e.hsv.V = float32(e.v.Percentage() / 100)
	}
	return changed
}

func (e *HsvEditor) SetColor(col color.NRGBA) {
	hsv := RgbToHsv(col)
	e.hsv.H = hsv.H
	e.hsv.S = hsv.S
	e.hsv.V = hsv.V
	e.h.SetDegree(int(hsv.H))
	e.s.SetPercentage(int(hsv.S * 100))
	e.v.SetPercentage(int(hsv.V * 100))
}

func parseHex(s string) ([]byte, bool) {
	out, err := hex.DecodeString(s)
	if err != nil {
		return nil, false
	}
	if len(out) < 3 {
		return nil, false
	}
	return out, true
}

func parseByte(s string) (byte, bool) {
	i, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return 0, false
	}
	return byte(i), true
}

func parseDegree(s string) (int, bool) {
	i, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, false
	}
	return int(i), true
}

func parsePercentage(s string) (int, bool) {
	i, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, false
	}
	return int(i), true
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
