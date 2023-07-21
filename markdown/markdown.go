// SPDX-License-Identifier: Unlicense OR MIT

/*
Package markdown transforms markdown text into gio richtext.
*/
package markdown

import (
	"bytes"
	"fmt"
	"image/color"
	"io/ioutil"
	"math"
	"regexp"
	"strings"

	"gioui.org/font"
	"gioui.org/unit"
	"gioui.org/x/richtext"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// Config defines settings used by the renderer.
type Config struct {
	DefaultFont   font.Font
	MonospaceFont font.Font
	// Defaults to 12 if unset.
	DefaultSize unit.Sp
	// If unset, each level will be 1.2 times larger than the previous.
	H1Size, H2Size, H3Size, H4Size, H5Size, H6Size unit.Sp
	// Defaults to black.
	DefaultColor color.NRGBA
	// Defaults to blue.
	InteractiveColor color.NRGBA
}

// gioNodeRenderer transforms AST nodes into gio's richtext types
type gioNodeRenderer struct {
	TextObjects []richtext.SpanStyle

	Config       Config
	Current      richtext.SpanStyle
	OrderedList  bool
	OrderedIndex int
}

func newNodeRenderer() *gioNodeRenderer {
	return &gioNodeRenderer{}
}

// CommitCurrent compies the state of the Current field and appends it to
// TextObjects. This finalizes the content and style of that section of text.
func (g *gioNodeRenderer) CommitCurrent() {
	g.TextObjects = append(g.TextObjects, g.Current.DeepCopy())
}

// UpdateCurrentSize edits only the size of the current text.
func (g *gioNodeRenderer) UpdateCurrentSize(sp unit.Sp) {
	g.Current.Size = sp
}

// UpdateCurrentColor edits only the color of the current text.
func (g *gioNodeRenderer) UpdateCurrentColor(c color.NRGBA) {
	g.Current.Color = c
}

// UpdateCurrentFont uses the provided font as a set of attributes to
// update. If any of those attributes are not their zero value, the
// current text's corresponding attribute will be updated to match.
// If the provided font is the zero value, the current font will be
// reset to the zero value as well.
func (g *gioNodeRenderer) UpdateCurrentFont(f font.Font) {
	reset := true
	if f.Style != 0 {
		reset = false
		g.Current.Font.Style = f.Style
	}
	if f.Typeface != "" {
		reset = false
		g.Current.Font.Typeface = f.Typeface
	}
	if f.Weight != 0 {
		reset = false
		g.Current.Font.Weight = f.Weight
	}
	if reset {
		g.Current.Font = f
	}
}

// AppendNewline ensures that there is a newline character at the end
// of the most-recently-generated TextObject.
func (g *gioNodeRenderer) AppendNewline() {
	if len(g.TextObjects) < 1 {
		return
	}
	g.TextObjects[len(g.TextObjects)-1].Content += "\n"
}

// EnsureSeparationFromPrevious ensures that next text object will be
// visually separated from the previous by a blank line. It achieves
// this by inserting a synthetic label containing only newlines if
// necessary.
func (g *gioNodeRenderer) EnsureSeparationFromPrevious() {
	if len(g.TextObjects) < 1 {
		return
	}
	last := g.TextObjects[len(g.TextObjects)-1]
	if !strings.HasSuffix(last.Content, "\n\n") {
		if strings.HasSuffix(last.Content, "\n") {
			g.Current.Content = "\n"
		} else {
			g.Current.Content = "\n\n"
		}
		g.CommitCurrent()
	}
}

func (g *gioNodeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// blocks
	//
	reg.Register(ast.KindDocument, g.renderDocument)
	reg.Register(ast.KindHeading, g.renderHeading)
	reg.Register(ast.KindBlockquote, g.renderBlockquote)
	reg.Register(ast.KindCodeBlock, g.renderCodeBlock)
	reg.Register(ast.KindFencedCodeBlock, g.renderFencedCodeBlock)
	reg.Register(ast.KindHTMLBlock, g.renderHTMLBlock)
	reg.Register(ast.KindList, g.renderList)
	reg.Register(ast.KindListItem, g.renderListItem)
	reg.Register(ast.KindParagraph, g.renderParagraph)
	reg.Register(ast.KindTextBlock, g.renderTextBlock)
	reg.Register(ast.KindThematicBreak, g.renderThematicBreak)
	//
	//	// inlines
	//
	reg.Register(ast.KindAutoLink, g.renderAutoLink)
	reg.Register(ast.KindCodeSpan, g.renderCodeSpan)
	reg.Register(ast.KindEmphasis, g.renderEmphasis)
	reg.Register(ast.KindImage, g.renderImage)
	reg.Register(ast.KindLink, g.renderLink)
	reg.Register(ast.KindRawHTML, g.renderRawHTML)
	reg.Register(ast.KindText, g.renderText)
	reg.Register(ast.KindString, g.renderString)
}

func (g *gioNodeRenderer) renderDocument(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {
		g.EnsureSeparationFromPrevious()
		var sp unit.Sp
		switch n.Level {
		case 1:
			sp = g.Config.H1Size
		case 2:
			sp = g.Config.H2Size
		case 3:
			sp = g.Config.H3Size
		case 4:
			sp = g.Config.H4Size
		case 5:
			sp = g.Config.H5Size
		case 6:
			sp = g.Config.H6Size
		}
		g.UpdateCurrentSize(sp)
	} else {
		g.UpdateCurrentSize(g.Config.DefaultSize)
	}
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderBlockquote(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		g.EnsureSeparationFromPrevious()
		g.Current.Font = g.Config.MonospaceFont
	} else {
		g.Current.Font = g.Config.DefaultFont
	}
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)
	if entering {
		g.EnsureSeparationFromPrevious()
		g.Current.Font = g.Config.MonospaceFont
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			g.Current.Content = string(line.Value(source))
			g.CommitCurrent()
		}
	} else {
		g.Current.Font = g.Config.DefaultFont
	}
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		g.EnsureSeparationFromPrevious()
		g.Current.Font = g.Config.MonospaceFont
	} else {
		g.Current.Font = g.Config.DefaultFont
	}
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.List)
	if entering {
		g.EnsureSeparationFromPrevious()
		g.OrderedList = n.IsOrdered()
		g.OrderedIndex = 1
	} else {
	}
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderListItem(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		if g.OrderedList {
			g.Current.Content = fmt.Sprintf(" %d. ", g.OrderedIndex)
			g.OrderedIndex++
		} else {
			g.Current.Content = " • "
		}
		g.CommitCurrent()
	} else if len(g.TextObjects) > 0 {
		g.AppendNewline()
	}

	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderParagraph(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		g.EnsureSeparationFromPrevious()
	}
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderTextBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderThematicBreak(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderAutoLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.AutoLink)
	if entering {
		url := string(n.URL(source))
		g.Current.Set(MetadataURL, url)
		g.Current.Color = g.Config.InteractiveColor
		g.Current.Content = url
		g.CommitCurrent()
	} else {
		g.Current.Set(MetadataURL, "")
		g.Current.Color = g.Config.DefaultColor
	}
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderCodeSpan(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		g.Current.Font = g.Config.MonospaceFont
	} else {
		g.Current.Font = g.Config.DefaultFont
	}
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderEmphasis(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Emphasis)

	if entering {
		if n.Level == 2 {
			g.Current.Font.Weight = font.Bold
		} else {
			g.Current.Font.Style = font.Italic
		}
	} else {
		if n.Level == 2 {
			g.Current.Font.Weight = font.Normal
		} else {
			g.Current.Font.Style = font.Regular
		}
	}
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

// MetadataURL is the metadata key that the parser will set for hyperlinks
// detected within the markdown.
const MetadataURL = "url"

func (g *gioNodeRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if entering {
		g.Current.Color = g.Config.InteractiveColor
		g.Current.Interactive = true
		g.Current.Set(MetadataURL, string(n.Destination))
	} else {
		g.Current.Color = g.Config.DefaultColor
		g.Current.Interactive = false
		g.Current.Set(MetadataURL, "")
	}
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderRawHTML(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Text)
	segment := n.Segment
	content := segment.Value(source)
	g.Current.Content = string(content)
	g.CommitCurrent()

	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderString(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.String)
	g.Current.Content = string(n.Value)
	g.CommitCurrent()
	return ast.WalkContinue, nil
}

// Result returns the accumulated text objects.
func (g *gioNodeRenderer) Result() []richtext.SpanStyle {
	o := g.TextObjects
	g.TextObjects = nil
	return o
}

// Renderer can transform source markdown into Gio richtext.
// Hyperlinks will result in text that has the URL set as span metadata
// with key MetadataURL.
type Renderer struct {
	md goldmark.Markdown
	nr *gioNodeRenderer
	// Config defines how the various markdown elements are presented.
	// If left as the zero value, sane defaults will be used.
	Config Config
}

// NewRenderer creates a ready-to-use markdown renderer.
func NewRenderer() *Renderer {
	nr := newNodeRenderer()
	md := goldmark.New(
		goldmark.WithRenderer(
			renderer.NewRenderer(
				renderer.WithNodeRenderers(
					util.PrioritizedValue{Value: nr, Priority: 0},
				),
			),
		),
	)
	return &Renderer{md: md, nr: nr}
}

// this regex matches a :// with one or more character that isn't whitespace
// a square bracket, or a parentheses on either side. It seems to reliably
// detect content that should be hyperlinked without actually matching
// markdown link syntax.
var urlExp = regexp.MustCompile(`(^|\s)([^([\s]+://[^)\]\s]+)`)

// Render transforms the provided src markdown into gio richtext using the
// fonts and styles defined by the given theme.
func (r *Renderer) Render(src []byte) ([]richtext.SpanStyle, error) {
	if bytes.Contains(src, []byte("://")) {
		src = urlExp.ReplaceAll(src, []byte("$1[$2]($2)"))
	}
	if r.Config.DefaultSize == 0 {
		r.Config.DefaultSize = 16
	}
	if r.Config.H6Size == 0 {
		r.Config.H6Size = unit.Sp(math.Round(1.2 * float64(r.Config.DefaultSize)))
	}
	if r.Config.H5Size == 0 {
		r.Config.H5Size = unit.Sp(math.Round(1.2 * float64(r.Config.H6Size)))
	}
	if r.Config.H4Size == 0 {
		r.Config.H4Size = unit.Sp(math.Round(1.2 * float64(r.Config.H5Size)))
	}
	if r.Config.H3Size == 0 {
		r.Config.H3Size = unit.Sp(math.Round(1.2 * float64(r.Config.H4Size)))
	}
	if r.Config.H2Size == 0 {
		r.Config.H2Size = unit.Sp(math.Round(1.2 * float64(r.Config.H3Size)))
	}
	if r.Config.H1Size == 0 {
		r.Config.H1Size = unit.Sp(math.Round(1.2 * float64(r.Config.H2Size)))
	}
	if r.Config.DefaultColor == (color.NRGBA{}) {
		r.Config.DefaultColor = color.NRGBA{A: 255}
	}
	if r.Config.MonospaceFont == (font.Font{}) {
		r.Config.MonospaceFont = font.Font{
			Typeface: "monospace",
			Weight:   r.Config.DefaultFont.Weight,
			Style:    r.Config.DefaultFont.Style,
		}
	}
	if r.Config.InteractiveColor == (color.NRGBA{}) {
		// Match the default material theme primary color.
		r.Config.InteractiveColor = color.NRGBA{R: 0x3f, G: 0x51, B: 0xb5, A: 255}
	}
	r.nr.Config = r.Config
	r.nr.UpdateCurrentColor(r.Config.DefaultColor)
	r.nr.UpdateCurrentFont(r.Config.DefaultFont)
	r.nr.UpdateCurrentSize(r.Config.DefaultSize)
	if err := r.md.Convert(src, ioutil.Discard); err != nil {
		return nil, err
	}
	return r.nr.Result(), nil
}
