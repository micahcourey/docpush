package confluence

import (
	"strings"
	"testing"

	"github.com/micahcourey/docpush/internal/converter"
)

func TestRender_Heading(t *testing.T) {
	src := []byte("# Hello World\n")
	p := converter.New()
	doc, err := p.Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	got, err := Render(doc.Source, doc.Node)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(got, "<h1>Hello World</h1>") {
		t.Errorf("expected <h1>Hello World</h1>, got: %s", got)
	}
}

func TestRender_Paragraph(t *testing.T) {
	src := []byte("Some text here.\n")
	p := converter.New()
	doc, err := p.Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	got, err := Render(doc.Source, doc.Node)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(got, "<p>Some text here.</p>") {
		t.Errorf("expected paragraph, got: %s", got)
	}
}

func TestRender_FencedCodeBlock(t *testing.T) {
	src := []byte("```typescript\nconst x = 1;\n```\n")
	p := converter.New()
	doc, err := p.Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	got, err := Render(doc.Source, doc.Node)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(got, `ac:name="code"`) {
		t.Errorf("expected code macro, got: %s", got)
	}
	if !strings.Contains(got, `ac:name="language">typescript`) {
		t.Errorf("expected language=typescript, got: %s", got)
	}
	if !strings.Contains(got, "const x = 1;") {
		t.Errorf("expected code content, got: %s", got)
	}
}

func TestRender_Table(t *testing.T) {
	src := []byte("| A | B |\n|---|---|\n| 1 | 2 |\n")
	p := converter.New()
	doc, err := p.Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	got, err := Render(doc.Source, doc.Node)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(got, "<table><tbody>") {
		t.Errorf("expected <table><tbody>, got: %s", got)
	}
	if !strings.Contains(got, "<th>") {
		t.Errorf("expected <th>, got: %s", got)
	}
	if !strings.Contains(got, "<td>") {
		t.Errorf("expected <td>, got: %s", got)
	}
}

func TestRender_Blockquote(t *testing.T) {
	src := []byte("> Important note\n")
	p := converter.New()
	doc, err := p.Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	got, err := Render(doc.Source, doc.Node)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(got, "<blockquote>") {
		t.Errorf("expected blockquote, got: %s", got)
	}
}

func TestRender_List(t *testing.T) {
	src := []byte("- Item 1\n- Item 2\n")
	p := converter.New()
	doc, err := p.Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	got, err := Render(doc.Source, doc.Node)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(got, "<ul>") {
		t.Errorf("expected <ul>, got: %s", got)
	}
	if !strings.Contains(got, "<li>") {
		t.Errorf("expected <li>, got: %s", got)
	}
}

func TestRender_OrderedList(t *testing.T) {
	src := []byte("1. First\n2. Second\n")
	p := converter.New()
	doc, err := p.Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	got, err := Render(doc.Source, doc.Node)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(got, "<ol>") {
		t.Errorf("expected <ol>, got: %s", got)
	}
}

func TestRender_Link(t *testing.T) {
	src := []byte("[Click here](https://example.com)\n")
	p := converter.New()
	doc, err := p.Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	got, err := Render(doc.Source, doc.Node)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(got, `href="https://example.com"`) {
		t.Errorf("expected link href, got: %s", got)
	}
}

func TestRender_Emphasis(t *testing.T) {
	src := []byte("*italic* and **bold**\n")
	p := converter.New()
	doc, err := p.Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	got, err := Render(doc.Source, doc.Node)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(got, "<em>italic</em>") {
		t.Errorf("expected <em>, got: %s", got)
	}
	if !strings.Contains(got, "<strong>bold</strong>") {
		t.Errorf("expected <strong>, got: %s", got)
	}
}

func TestRender_Image_Local(t *testing.T) {
	src := []byte("![alt](images/diagram.png)\n")
	p := converter.New()
	doc, err := p.Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	got, err := Render(doc.Source, doc.Node)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(got, `ri:filename="diagram.png"`) {
		t.Errorf("expected attachment macro, got: %s", got)
	}
}

func TestRenderWithImages_CollectsLocalPaths(t *testing.T) {
	src := []byte("![](images/arch.png)\n\nSome text.\n\n![](images/flow.svg)\n")
	p := converter.New()
	doc, err := p.Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	result, err := RenderWithImages(doc.Source, doc.Node)
	if err != nil {
		t.Fatalf("RenderWithImages error: %v", err)
	}
	if len(result.LocalImages) != 2 {
		t.Fatalf("expected 2 local images, got %d: %v", len(result.LocalImages), result.LocalImages)
	}
	if result.LocalImages[0] != "images/arch.png" {
		t.Errorf("expected images/arch.png, got %s", result.LocalImages[0])
	}
	if result.LocalImages[1] != "images/flow.svg" {
		t.Errorf("expected images/flow.svg, got %s", result.LocalImages[1])
	}
	// XHTML should still contain the attachment macros
	if !strings.Contains(result.XHTML, `ri:filename="arch.png"`) {
		t.Errorf("expected arch.png attachment macro in XHTML")
	}
}

func TestRenderWithImages_IgnoresExternalImages(t *testing.T) {
	src := []byte("![](https://example.com/logo.png)\n\n![](images/local.png)\n")
	p := converter.New()
	doc, err := p.Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	result, err := RenderWithImages(doc.Source, doc.Node)
	if err != nil {
		t.Fatalf("RenderWithImages error: %v", err)
	}
	if len(result.LocalImages) != 1 {
		t.Fatalf("expected 1 local image, got %d: %v", len(result.LocalImages), result.LocalImages)
	}
	if result.LocalImages[0] != "images/local.png" {
		t.Errorf("expected images/local.png, got %s", result.LocalImages[0])
	}
}

func TestRenderWithImages_NoImages(t *testing.T) {
	src := []byte("# Hello\n\nJust text.\n")
	p := converter.New()
	doc, err := p.Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	result, err := RenderWithImages(doc.Source, doc.Node)
	if err != nil {
		t.Fatalf("RenderWithImages error: %v", err)
	}
	if len(result.LocalImages) != 0 {
		t.Errorf("expected 0 local images, got %d", len(result.LocalImages))
	}
}
