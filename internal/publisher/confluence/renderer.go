// Package confluence implements the Publisher interface for Confluence Data Center.
package confluence

import (
	"bytes"
	"fmt"
	"strings"

	east "github.com/yuin/goldmark/extension/ast"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

// RenderResult holds the rendered XHTML output and any local image paths found.
type RenderResult struct {
	XHTML      string
	LocalImages []string // relative paths to local images (e.g., "images/diagram.png")
}

// Render converts a goldmark AST to Confluence XHTML storage format.
func Render(source []byte, node ast.Node) (string, error) {
	result, err := RenderWithImages(source, node)
	if err != nil {
		return "", err
	}
	return result.XHTML, nil
}

// RenderWithImages converts a goldmark AST to Confluence XHTML storage format
// and also returns a list of local image paths referenced in the document.
func RenderWithImages(source []byte, node ast.Node) (*RenderResult, error) {
	var buf bytes.Buffer
	r := &renderer{source: source, buf: &buf}
	if err := r.render(node); err != nil {
		return nil, err
	}
	return &RenderResult{
		XHTML:      buf.String(),
		LocalImages: r.localImages,
	}, nil
}

type renderer struct {
	source      []byte
	buf         *bytes.Buffer
	localImages []string
}

func (r *renderer) render(node ast.Node) error {
	return ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		var err error
		switch v := n.(type) {
		case *ast.Document:
			// no-op wrapper
		case *ast.Heading:
			err = r.renderHeading(v, entering)
		case *ast.Paragraph:
			err = r.renderParagraph(v, entering)
		case *ast.TextBlock:
			// in tables, paragraphs may be TextBlocks
		case *ast.Text:
			err = r.renderText(v, entering)
		case *ast.String:
			if entering {
				r.buf.Write(v.Value)
			}
		case *ast.CodeSpan:
			err = r.renderCodeSpan(v, entering)
		case *ast.FencedCodeBlock:
			err = r.renderFencedCodeBlock(v, entering)
		case *ast.CodeBlock:
			err = r.renderCodeBlock(v, entering)
		case *ast.Blockquote:
			err = r.renderBlockquote(v, entering)
		case *ast.List:
			err = r.renderList(v, entering)
		case *ast.ListItem:
			err = r.renderListItem(v, entering)
		case *ast.ThematicBreak:
			if entering {
				r.buf.WriteString("<hr />")
			}
		case *ast.Link:
			err = r.renderLink(v, entering)
		case *ast.Image:
			err = r.renderImage(v, entering)
		case *ast.Emphasis:
			err = r.renderEmphasis(v, entering)
		case *ast.RawHTML:
			err = r.renderRawHTML(v, entering)
		case *ast.HTMLBlock:
			err = r.renderHTMLBlock(v, entering)
		case *ast.AutoLink:
			err = r.renderAutoLink(v, entering)

		// GFM extensions
		case *east.Table:
			err = r.renderTable(v, entering)
		case *east.TableHeader:
			err = r.renderTableHeader(v, entering)
		case *east.TableRow:
			err = r.renderTableRow(v, entering)
		case *east.TableCell:
			err = r.renderTableCell(v, entering)
		case *east.Strikethrough:
			err = r.renderStrikethrough(v, entering)
		case *east.TaskCheckBox:
			err = r.renderTaskCheckBox(v, entering)
		}
		return ast.WalkContinue, err
	})
}

func (r *renderer) renderHeading(n *ast.Heading, entering bool) error {
	if entering {
		fmt.Fprintf(r.buf, "<h%d>", n.Level)
	} else {
		fmt.Fprintf(r.buf, "</h%d>", n.Level)
	}
	return nil
}

func (r *renderer) renderParagraph(n *ast.Paragraph, entering bool) error {
	// Skip paragraph tags inside table cells and list items — Confluence doesn't want them
	if n.Parent() != nil {
		switch n.Parent().(type) {
		case *east.TableCell:
			return nil
		}
	}
	if entering {
		r.buf.WriteString("<p>")
	} else {
		r.buf.WriteString("</p>")
	}
	return nil
}

func (r *renderer) renderText(n *ast.Text, entering bool) error {
	if !entering {
		return nil
	}
	segment := n.Segment
	r.buf.Write(util.EscapeHTML(segment.Value(r.source)))
	if n.HardLineBreak() {
		r.buf.WriteString("<br />")
	} else if n.SoftLineBreak() {
		r.buf.WriteString("\n")
	}
	return nil
}

func (r *renderer) renderCodeSpan(_ *ast.CodeSpan, entering bool) error {
	if entering {
		r.buf.WriteString("<code>")
	} else {
		r.buf.WriteString("</code>")
	}
	return nil
}

func (r *renderer) renderFencedCodeBlock(n *ast.FencedCodeBlock, entering bool) error {
	if entering {
		language := string(n.Language(r.source))
		r.buf.WriteString(`<ac:structured-macro ac:name="code">`)
		if language != "" {
			fmt.Fprintf(r.buf, `<ac:parameter ac:name="language">%s</ac:parameter>`, escapeXML(language))
		}
		r.buf.WriteString(`<ac:plain-text-body><![CDATA[`)
		// Write all lines of the code block
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			r.buf.Write(line.Value(r.source))
		}
		r.buf.WriteString(`]]></ac:plain-text-body></ac:structured-macro>`)
	}
	return nil
}

func (r *renderer) renderCodeBlock(n *ast.CodeBlock, entering bool) error {
	if entering {
		r.buf.WriteString(`<ac:structured-macro ac:name="code"><ac:plain-text-body><![CDATA[`)
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			r.buf.Write(line.Value(r.source))
		}
		r.buf.WriteString(`]]></ac:plain-text-body></ac:structured-macro>`)
	}
	return nil
}

func (r *renderer) renderBlockquote(_ *ast.Blockquote, entering bool) error {
	if entering {
		r.buf.WriteString("<blockquote>")
	} else {
		r.buf.WriteString("</blockquote>")
	}
	return nil
}

func (r *renderer) renderList(n *ast.List, entering bool) error {
	tag := "ul"
	if n.IsOrdered() {
		tag = "ol"
	}
	if entering {
		fmt.Fprintf(r.buf, "<%s>", tag)
	} else {
		fmt.Fprintf(r.buf, "</%s>", tag)
	}
	return nil
}

func (r *renderer) renderListItem(_ *ast.ListItem, entering bool) error {
	if entering {
		r.buf.WriteString("<li>")
	} else {
		r.buf.WriteString("</li>")
	}
	return nil
}

func (r *renderer) renderLink(n *ast.Link, entering bool) error {
	if entering {
		fmt.Fprintf(r.buf, `<a href="%s"`, escapeXML(string(n.Destination)))
		if n.Title != nil {
			fmt.Fprintf(r.buf, ` title="%s"`, escapeXML(string(n.Title)))
		}
		r.buf.WriteString(">")
	} else {
		r.buf.WriteString("</a>")
	}
	return nil
}

func (r *renderer) renderImage(n *ast.Image, entering bool) error {
	if !entering {
		return nil
	}
	src := string(n.Destination)
	alt := nodeText(n, r.source)

	// Local images → Confluence attachment macro
	if !strings.HasPrefix(src, "http://") && !strings.HasPrefix(src, "https://") {
		filename := src
		if idx := strings.LastIndex(src, "/"); idx >= 0 {
			filename = src[idx+1:]
		}
		r.buf.WriteString(`<ac:image>`)
		fmt.Fprintf(r.buf, `<ri:attachment ri:filename="%s" />`, escapeXML(filename))
		r.buf.WriteString(`</ac:image>`)
		// Track local image path for attachment upload
		r.localImages = append(r.localImages, src)
	} else {
		// External images → regular img via Confluence image macro
		r.buf.WriteString(`<ac:image>`)
		fmt.Fprintf(r.buf, `<ri:url ri:value="%s" />`, escapeXML(src))
		r.buf.WriteString(`</ac:image>`)
	}
	_ = alt
	return nil
}

func (r *renderer) renderEmphasis(n *ast.Emphasis, entering bool) error {
	tag := "em"
	if n.Level == 2 {
		tag = "strong"
	}
	if entering {
		fmt.Fprintf(r.buf, "<%s>", tag)
	} else {
		fmt.Fprintf(r.buf, "</%s>", tag)
	}
	return nil
}

func (r *renderer) renderRawHTML(n *ast.RawHTML, entering bool) error {
	if !entering {
		return nil
	}
	for i := 0; i < n.Segments.Len(); i++ {
		seg := n.Segments.At(i)
		r.buf.Write(seg.Value(r.source))
	}
	return nil
}

func (r *renderer) renderHTMLBlock(n *ast.HTMLBlock, entering bool) error {
	if !entering {
		return nil
	}
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		r.buf.Write(line.Value(r.source))
	}
	return nil
}

func (r *renderer) renderAutoLink(n *ast.AutoLink, entering bool) error {
	if entering {
		url := string(n.URL(r.source))
		fmt.Fprintf(r.buf, `<a href="%s">`, escapeXML(url))
	} else {
		r.buf.WriteString("</a>")
	}
	return nil
}

// --- GFM Table rendering ---

func (r *renderer) renderTable(_ *east.Table, entering bool) error {
	if entering {
		r.buf.WriteString("<table><tbody>")
	} else {
		r.buf.WriteString("</tbody></table>")
	}
	return nil
}

func (r *renderer) renderTableHeader(_ *east.TableHeader, entering bool) error {
	if entering {
		r.buf.WriteString("<tr>")
	} else {
		r.buf.WriteString("</tr>")
	}
	return nil
}

func (r *renderer) renderTableRow(_ *east.TableRow, entering bool) error {
	if entering {
		r.buf.WriteString("<tr>")
	} else {
		r.buf.WriteString("</tr>")
	}
	return nil
}

func (r *renderer) renderTableCell(n *east.TableCell, entering bool) error {
	tag := "td"
	if n.Parent() != nil {
		if _, ok := n.Parent().(*east.TableHeader); ok {
			tag = "th"
		}
	}
	if entering {
		switch n.Alignment {
		case east.AlignLeft:
			fmt.Fprintf(r.buf, `<%s style="text-align: left">`, tag)
		case east.AlignCenter:
			fmt.Fprintf(r.buf, `<%s style="text-align: center">`, tag)
		case east.AlignRight:
			fmt.Fprintf(r.buf, `<%s style="text-align: right">`, tag)
		default:
			fmt.Fprintf(r.buf, "<%s>", tag)
		}
	} else {
		fmt.Fprintf(r.buf, "</%s>", tag)
	}
	return nil
}

func (r *renderer) renderStrikethrough(_ *east.Strikethrough, entering bool) error {
	if entering {
		r.buf.WriteString("<del>")
	} else {
		r.buf.WriteString("</del>")
	}
	return nil
}

func (r *renderer) renderTaskCheckBox(n *east.TaskCheckBox, entering bool) error {
	if !entering {
		return nil
	}
	if n.IsChecked {
		r.buf.WriteString(`<ac:task-status>complete</ac:task-status> `)
	} else {
		r.buf.WriteString(`<ac:task-status>incomplete</ac:task-status> `)
	}
	return nil
}

// --- Helpers ---

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

func nodeText(n ast.Node, source []byte) string {
	var buf bytes.Buffer
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			buf.Write(t.Segment.Value(source))
		}
	}
	return buf.String()
}
