package colorpicker

import (
	"encoding/hex"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"strconv"
)

func newHexField(theme *material.Theme, editor widget.Editor) *hexField {
	he := &hexField{theme: theme, editor: editor}
	return he
}

type hexField struct {
	theme   *material.Theme
	editor  widget.Editor
	value   []byte
	Invalid bool
}

func (ed *hexField) Changed() bool {
	s := ed.editor.Text()
	newValue, err := hex.DecodeString(ed.editor.Text())
	if len(s) != 0 && err != nil {
		return false
	}
	changed := !bytesEqual(newValue, ed.value)
	if changed {
		ed.value = newValue
	}
	return changed
}

func (ed *hexField) SetHex(v []byte) {
	ed.value = v
	ed.editor.SetText(hex.EncodeToString(v))
}

func (ed *hexField) Hex() []byte {
	return ed.value
}

func (ed *hexField) Layout(gtx layout.Context) layout.Dimensions {
	return material.Editor(ed.theme, &ed.editor, "").Layout(gtx)
}

type byteField struct {
	editor  widget.Editor
	Invalid bool
	old     byte
}

func (ed *byteField) Changed() bool {
	newText := ed.editor.Text()
	newByte, ok := parseByte(newText)
	if !ok {
		return false
	}
	changed := newByte != ed.old
	ed.old = newByte
	return changed
}

func (ed *byteField) SetByte(b byte) {
	ed.old = b
	ed.editor.SetText(strconv.FormatUint(uint64(b), 10))
}

func (ed *byteField) Byte() byte {
	return ed.old
}

type percentageField struct {
	editor  widget.Editor
	Invalid bool
	old     int
}

func (ed *percentageField) Changed() bool {
	newText := ed.editor.Text()
	newInt, ok := parsePercentage(newText)
	if !ok {
		return false
	}
	changed := newInt != ed.old
	ed.old = newInt
	return changed
}

// SetText sets editor content without marking the editor changed.
func (ed *percentageField) SetPercentage(p int) {
	ed.old = p
	ed.editor.SetText(strconv.Itoa(p))
}

func (ed *percentageField) Percentage() int {
	return ed.old
}

type degreeField struct {
	editor  widget.Editor
	Invalid bool
	old     int
}

func (ed *degreeField) Changed() bool {
	newText := ed.editor.Text()
	degree, ok := parseDegree(newText)
	if !ok {
		return false
	}
	changed := degree != ed.old
	ed.old = degree
	return changed
}

// SetText sets editor content without marking the editor changed.
func (ed *degreeField) SetDegree(d int) {
	ed.old = d
	ed.editor.SetText(strconv.Itoa(d))
}

func (ed *degreeField) Degree() int {
	return ed.old
}
