package colorpicker

import (
	"gioui.org/widget"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

var (
	toggleIcon = func() *widget.Icon {
		icon, err := widget.NewIcon(icons.NavigationUnfoldMore)
		if err != nil {
			panic(err)
		}
		return icon
	}()
)

func loadIcon(b []byte) *widget.Icon {
	icon, err := widget.NewIcon(b)
	if err != nil {
		panic(err)
	}
	return icon
}
