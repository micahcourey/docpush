package converter

import (
	"testing"
)

func TestParse_BasicMarkdown(t *testing.T) {
	source := []byte("# Hello\n\nThis is a paragraph.\n")
	p := New()
	doc, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if doc.Node == nil {
		t.Fatal("Parse() returned nil Node")
	}
	if doc.Node.ChildCount() == 0 {
		t.Fatal("Parse() returned empty AST")
	}
}

func TestParse_Frontmatter(t *testing.T) {
	source := []byte(`---
title: Test Doc
docpush:
  confluence:
    space: MYSPACE
    pageId: 12345
---
# Hello

Content here.
`)
	p := New()
	doc, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if doc.Metadata == nil {
		t.Fatal("Parse() returned nil Metadata")
	}
	if doc.Metadata["title"] != "Test Doc" {
		t.Errorf("Metadata[title] = %v, want Test Doc", doc.Metadata["title"])
	}
	dp, ok := doc.Metadata["docpush"].(map[interface{}]interface{})
	if !ok {
		// goldmark-meta may return map[string]interface{} depending on version
		dp2, ok2 := doc.Metadata["docpush"].(map[string]interface{})
		if !ok2 {
			t.Fatalf("Metadata[docpush] unexpected type: %T", doc.Metadata["docpush"])
		}
		if dp2["confluence"] == nil {
			t.Fatal("Metadata[docpush][confluence] is nil")
		}
		return
	}
	if dp["confluence"] == nil {
		t.Fatal("Metadata[docpush][confluence] is nil")
	}
}

func TestParse_GFMTable(t *testing.T) {
	source := []byte(`| Col A | Col B |
|-------|-------|
| 1     | 2     |
`)
	p := New()
	doc, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if doc.Node.ChildCount() == 0 {
		t.Fatal("Parse() returned empty AST for GFM table")
	}
}
