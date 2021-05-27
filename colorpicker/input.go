package colorpicker

import (
	"gioui.org/layout"
	"image/color"
)

type ColorInput interface {
	Layout(gtx layout.Context) layout.Dimensions
	Changed() bool
	SetColor(col color.NRGBA)
	Color() color.NRGBA
}
