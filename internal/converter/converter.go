// Package converter provides shared markdown parsing via goldmark.
package converter

import (
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// ParsedDoc holds the parsed document with its AST and metadata.
type ParsedDoc struct {
	// Source is the original markdown bytes.
	Source []byte
	// Node is the root AST node.
	Node ast.Node
	// Metadata is the extracted YAML frontmatter.
	Metadata map[string]interface{}
}

// Parser wraps a configured goldmark instance for parsing markdown.
type Parser struct {
	md goldmark.Markdown
}

// New creates a Parser with GFM extensions and frontmatter support.
func New() *Parser {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			meta.Meta,
		),
	)
	return &Parser{md: md}
}

// Parse parses markdown source and returns the AST node and frontmatter metadata.
func (p *Parser) Parse(source []byte) (*ParsedDoc, error) {
	ctx := parser.NewContext()
	reader := text.NewReader(source)
	doc := p.md.Parser().Parse(reader, parser.WithContext(ctx))

	metadata := meta.Get(ctx)

	return &ParsedDoc{
		Source:   source,
		Node:     doc,
		Metadata: metadata,
	}, nil
}
