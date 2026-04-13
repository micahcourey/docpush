// Package diff provides content hash comparison for skip-on-no-change detection.
package diff

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// Hash returns a SHA-256 hex digest of the given content after normalization.
func Hash(content string) string {
	normalized := normalize(content)
	h := sha256.Sum256([]byte(normalized))
	return fmt.Sprintf("%x", h)
}

// Equal returns true if two content strings are semantically identical
// after whitespace normalization.
func Equal(a, b string) bool {
	return Hash(a) == Hash(b)
}

// normalize trims and collapses whitespace to enable robust comparison
// between locally rendered and remotely stored XHTML.
func normalize(s string) string {
	s = strings.TrimSpace(s)
	// Collapse runs of whitespace (newlines, tabs, spaces) to single space
	var prev rune
	var buf strings.Builder
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if prev != ' ' {
				buf.WriteRune(' ')
			}
			prev = ' '
		} else {
			buf.WriteRune(r)
			prev = r
		}
	}
	return buf.String()
}
