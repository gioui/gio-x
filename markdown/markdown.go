// SPDX-License-Identifier: Unlicense OR MIT

/*
Package markdown transforms markdown text into gio richtext.
*/
package markdown

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"gioui.org/text"
	"gioui.org/widget/material"
	"gioui.org/x/richtext"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// gioNodeRenderer transforms AST nodes into gio's richtext types
type gioNodeRenderer struct {
	TextObjects []richtext.SpanStyle

	Current      richtext.SpanStyle
	Theme        *material.Theme
	OrderedList  bool
	OrderedIndex int
}

func newNodeRenderer() *gioNodeRenderer {
	return &gioNodeRenderer{}
}

func (g *gioNodeRenderer) CommitCurrent() {
	g.TextObjects = append(g.TextObjects, g.Current.DeepCopy())
}

func (g *gioNodeRenderer) UpdateCurrent(l material.LabelStyle) {
	g.Current.Font = l.Font
	g.Current.Color = l.Color
	g.Current.Size = l.TextSize
}

func (g *gioNodeRenderer) AppendNewline() {
	if len(g.TextObjects) < 1 {
		return
	}
	g.TextObjects[len(g.TextObjects)-1].Content += "\n"
}

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
		var l material.LabelStyle
		switch n.Level {
		case 1:
			l = material.H1(g.Theme, "")
		case 2:
			l = material.H2(g.Theme, "")
		case 3:
			l = material.H3(g.Theme, "")
		case 4:
			l = material.H4(g.Theme, "")
		case 5:
			l = material.H5(g.Theme, "")
		case 6:
			l = material.H6(g.Theme, "")
		}
		g.UpdateCurrent(l)
	} else {
		l := material.Body1(g.Theme, "")
		g.UpdateCurrent(l)
	}
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderBlockquote(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		g.EnsureSeparationFromPrevious()
		g.Current.Font.Variant = "Mono"
	} else {
		g.Current.Font.Variant = ""
	}
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)
	if entering {
		g.EnsureSeparationFromPrevious()
		g.Current.Font.Variant = "Mono"
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			g.Current.Content = string(line.Value(source))
			g.CommitCurrent()
		}
	} else {
		g.Current.Font.Variant = ""
	}
	return ast.WalkContinue, nil
}

func (g *gioNodeRenderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		g.EnsureSeparationFromPrevious()
		g.Current.Font.Variant = "Mono"
	} else {
		g.Current.Font.Variant = ""
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
		g.Current.Set(urlMetadataKey, url)
		g.Current.Color = g.Theme.ContrastBg
		g.Current.Content = url
		g.CommitCurrent()
	} else {
		g.Current.Set(urlMetadataKey, "")
		g.Current.Color = g.Theme.Fg
	}
	return ast.WalkContinue, nil
}
func (g *gioNodeRenderer) renderCodeSpan(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		g.Current.Font.Variant = "Mono"
	} else {
		g.Current.Font.Variant = ""
	}
	return ast.WalkContinue, nil
}
func (g *gioNodeRenderer) renderEmphasis(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Emphasis)

	if entering {
		if n.Level == 2 {
			g.Current.Font.Weight = text.Bold
		} else {
			g.Current.Font.Style = text.Italic
		}
	} else {
		g.Current.Font.Style = text.Regular
		g.Current.Font.Weight = text.Normal
	}
	return ast.WalkContinue, nil
}
func (g *gioNodeRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

const urlMetadataKey = "url"

func (g *gioNodeRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if entering {
		g.Current.Color = g.Theme.ContrastBg
		g.Current.Interactive = true
		g.Current.Set("url", string(n.Destination))
	} else {
		g.Current.Color = g.Theme.Fg
		g.Current.Interactive = false
		g.Current.Set("url", "")
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

func (g *gioNodeRenderer) Result() []richtext.SpanStyle {
	o := g.TextObjects
	g.TextObjects = nil
	return o
}

// Renderer can transform source markdown into Gio richtext.
type Renderer struct {
	md goldmark.Markdown
	nr *gioNodeRenderer
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

var urlExp = regexp.MustCompile(`(^|\s)([^[\s]\S*://\S+)`)

// Render transforms the provided src markdown into gio richtext using the
// fonts and styles defined by the given theme.
func (r *Renderer) Render(th *material.Theme, src []byte) ([]richtext.SpanStyle, error) {
	if bytes.Contains(src, []byte("://")) {
		src = urlExp.ReplaceAll(src, []byte("$1[$2]($2)"))
	}
	l := material.Body1(th, "")
	r.nr.Theme = th
	r.nr.UpdateCurrent(l)
	if err := r.md.Convert(src, ioutil.Discard); err != nil {
		return nil, err
	}
	return r.nr.Result(), nil
}
