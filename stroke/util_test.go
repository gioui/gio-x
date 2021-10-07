// SPDX-License-Identifier: Unlicense OR MIT

package stroke

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"

	"golang.org/x/image/colornames"

	"gioui.org/gpu/headless"
	"gioui.org/internal/f32color"
	"gioui.org/op"
)

var (
	dumpImages = flag.Bool("saveimages", false, "save test images")
)

var (
	red         = f32color.RGBAToNRGBA(colornames.Red)
	black       = f32color.RGBAToNRGBA(colornames.Black)
	transparent = color.RGBA{}
)

func drawImage(t *testing.T, size int, ops *op.Ops, draw func(o *op.Ops)) (im *image.RGBA, err error) {
	sz := image.Point{X: size, Y: size}
	w := newWindow(t, sz.X, sz.Y)
	defer w.Release()
	draw(ops)
	if err := w.Frame(ops); err != nil {
		return nil, err
	}
	return w.Screenshot()
}

func run(t *testing.T, f func(o *op.Ops), c func(r result)) {
	// Draw a few times and check that it is correct each time, to
	// ensure any caching effects still generate the correct images.
	var img *image.RGBA
	var err error
	ops := new(op.Ops)
	for i := 0; i < 3; i++ {
		ops.Reset()
		img, err = drawImage(t, 128, ops, f)
		if err != nil {
			t.Error("error rendering:", err)
			return
		}
		// Check for a reference image and make sure it is identical.
		if !verifyRef(t, img, 0) {
			name := fmt.Sprintf("%s-%d-bad.png", t.Name(), i)
			saveImage(t, name, img)
		}
		c(result{t: t, img: img})
	}
}

func verifyRef(t *testing.T, img *image.RGBA, frame int) (ok bool) {
	// ensure identical to ref data
	path := filepath.Join("refs", t.Name()+".png")
	if frame != 0 {
		path = filepath.Join("refs", t.Name()+"_"+strconv.Itoa(frame)+".png")
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error("could not open ref:", err)
		return
	}
	r, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Error("could not decode ref:", err)
		return
	}
	if img.Bounds() != r.Bounds() {
		t.Errorf("reference image is %v, expected %v", r.Bounds(), img.Bounds())
		return false
	}
	var ref *image.RGBA
	switch r := r.(type) {
	case *image.RGBA:
		ref = r
	case *image.NRGBA:
		ref = image.NewRGBA(r.Bounds())
		bnd := r.Bounds()
		for x := bnd.Min.X; x < bnd.Max.X; x++ {
			for y := bnd.Min.Y; y < bnd.Max.Y; y++ {
				ref.SetRGBA(x, y, f32color.NRGBAToRGBA(r.NRGBAAt(x, y)))
			}
		}
	default:
		t.Fatalf("reference image is a %T, expected *image.NRGBA or *image.RGBA", r)
	}
	bnd := img.Bounds()
	for x := bnd.Min.X; x < bnd.Max.X; x++ {
		for y := bnd.Min.Y; y < bnd.Max.Y; y++ {
			exp := ref.RGBAAt(x, y)
			got := img.RGBAAt(x, y)
			if !colorsClose(exp, got) {
				t.Error("not equal to ref at", x, y, " ", got, exp)
				return false
			}
		}
	}
	return true
}

func colorsClose(c1, c2 color.RGBA) bool {
	const delta = 0.01 // magic value obtained from experimentation.
	return yiqEqApprox(c1, c2, delta)
}

// yiqEqApprox compares the colors of 2 pixels, in the NTSC YIQ color space,
// as described in:
//
//   Measuring perceived color difference using YIQ NTSC
//   transmission color space in mobile applications.
//   Yuriy Kotsarenko, Fernando Ramos.
//
// An electronic version is available at:
//
// - http://www.progmat.uaem.mx:8080/artVol2Num2/Articulo3Vol2Num2.pdf
func yiqEqApprox(c1, c2 color.RGBA, d2 float64) bool {
	const max = 35215.0 // difference between 2 maximally different pixels.

	var (
		r1 = float64(c1.R)
		g1 = float64(c1.G)
		b1 = float64(c1.B)

		r2 = float64(c2.R)
		g2 = float64(c2.G)
		b2 = float64(c2.B)

		y1 = r1*0.29889531 + g1*0.58662247 + b1*0.11448223
		i1 = r1*0.59597799 - g1*0.27417610 - b1*0.32180189
		q1 = r1*0.21147017 - g1*0.52261711 + b1*0.31114694

		y2 = r2*0.29889531 + g2*0.58662247 + b2*0.11448223
		i2 = r2*0.59597799 - g2*0.27417610 - b2*0.32180189
		q2 = r2*0.21147017 - g2*0.52261711 + b2*0.31114694

		y = y1 - y2
		i = i1 - i2
		q = q1 - q2

		diff = 0.5053*y*y + 0.299*i*i + 0.1957*q*q
	)
	return diff <= max*d2
}

func (r result) expect(x, y int, col color.RGBA) {
	r.t.Helper()
	if r.img == nil {
		return
	}
	c := r.img.RGBAAt(x, y)
	if !colorsClose(c, col) {
		r.t.Error("expected ", col, " at ", "(", x, ",", y, ") but got ", c)
	}
}

type result struct {
	t   *testing.T
	img *image.RGBA
}

func saveImage(t testing.TB, file string, img *image.RGBA) {
	if !*dumpImages {
		return
	}
	// Only NRGBA images are losslessly encoded by png.Encode.
	nrgba := image.NewNRGBA(img.Bounds())
	bnd := img.Bounds()
	for x := bnd.Min.X; x < bnd.Max.X; x++ {
		for y := bnd.Min.Y; y < bnd.Max.Y; y++ {
			nrgba.SetNRGBA(x, y, f32color.RGBAToNRGBA(img.RGBAAt(x, y)))
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, nrgba); err != nil {
		t.Error(err)
		return
	}
	if err := ioutil.WriteFile(file, buf.Bytes(), 0666); err != nil {
		t.Error(err)
		return
	}
}

func newWindow(t testing.TB, width, height int) *headless.Window {
	w, err := headless.NewWindow(width, height)
	if err != nil {
		t.Skipf("failed to create headless window, skipping: %v", err)
	}
	return w
}
