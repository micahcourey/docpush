package diff

import "testing"

func TestHash_Deterministic(t *testing.T) {
	a := Hash("<p>Hello World</p>")
	b := Hash("<p>Hello World</p>")
	if a != b {
		t.Errorf("expected identical hashes, got %s vs %s", a, b)
	}
}

func TestHash_Different(t *testing.T) {
	a := Hash("<p>Hello</p>")
	b := Hash("<p>World</p>")
	if a == b {
		t.Error("expected different hashes for different content")
	}
}

func TestEqual_WhitespaceNormalization(t *testing.T) {
	a := "<p>Hello  World</p>"
	b := "<p>Hello\n\tWorld</p>"
	if !Equal(a, b) {
		t.Error("expected Equal to normalize whitespace")
	}
}

func TestEqual_TrailingWhitespace(t *testing.T) {
	a := "<p>Hello</p>  \n"
	b := "<p>Hello</p>"
	if !Equal(a, b) {
		t.Error("expected Equal to trim trailing whitespace")
	}
}

func TestEqual_NotEqual(t *testing.T) {
	a := "<p>Hello</p>"
	b := "<p>Goodbye</p>"
	if Equal(a, b) {
		t.Error("expected not equal for different content")
	}
}
