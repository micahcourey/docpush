package confluence

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/micahcourey/docpush/internal/publisher"
)

func TestAdapter_Publish_DryRun(t *testing.T) {
	cfg := TargetConfig{
		URL:   "https://example.com",
		Space: "TEST",
	}
	a := New(cfg)

	page := &publisher.Page{
		LocalPath: "docs/test.md",
		Title:     "Test",
		Body:      []byte("# Hello\n\nWorld.\n"),
		Metadata:  map[string]any{"pageId": "123"},
	}
	result, err := a.Publish(context.Background(), page, publisher.PublishOpts{DryRun: true})
	if err != nil {
		t.Fatalf("Publish dry-run error: %v", err)
	}
	if result.Action != "dry-run" {
		t.Errorf("expected action dry-run, got %s", result.Action)
	}
}

func TestAdapter_Publish_Update(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resp := PageResponse{ID: "123", Title: "Test"}
			resp.Version.Number = 2
			resp.Body.Storage.Value = "<p>Old content</p>"
			json.NewEncoder(w).Encode(resp)
		case http.MethodPut:
			resp := PageResponse{ID: "123", Title: "Test"}
			resp.Version.Number = 3
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer srv.Close()

	cfg := TargetConfig{URL: srv.URL, Space: "TEST"}
	a := New(cfg)
	// Override client to use test server
	a.client = NewClient(srv.URL, "test-pat")

	page := &publisher.Page{
		LocalPath: "docs/test.md",
		Title:     "Test",
		Body:      []byte("# New Content\n\nUpdated.\n"),
		Metadata:  map[string]any{"pageId": "123"},
	}
	result, err := a.Publish(context.Background(), page, publisher.PublishOpts{})
	if err != nil {
		t.Fatalf("Publish update error: %v", err)
	}
	if result.Action != "updated" {
		t.Errorf("expected action updated, got %s", result.Action)
	}
}

func TestAdapter_Publish_Create(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			resp := PageResponse{ID: "999", Title: "New Doc"}
			resp.Version.Number = 1
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer srv.Close()

	cfg := TargetConfig{
		URL:   srv.URL,
		Space: "TEST",
		Defaults: DefaultsConfig{
			ParentID: "100",
		},
	}
	a := New(cfg)
	a.client = NewClient(srv.URL, "test-pat")

	page := &publisher.Page{
		LocalPath: "docs/new.md",
		Title:     "New Doc",
		Body:      []byte("# Brand New\n\nContent.\n"),
		Metadata:  map[string]any{},
	}
	result, err := a.Publish(context.Background(), page, publisher.PublishOpts{CreateIfMissing: true})
	if err != nil {
		t.Fatalf("Publish create error: %v", err)
	}
	if result.Action != "created" {
		t.Errorf("expected action created, got %s", result.Action)
	}
	if result.PageID != "999" {
		t.Errorf("expected pageID 999, got %s", result.PageID)
	}
}

func TestAdapter_Publish_NoCreateWhenMissing(t *testing.T) {
	cfg := TargetConfig{URL: "https://example.com", Space: "TEST"}
	a := New(cfg)

	page := &publisher.Page{
		LocalPath: "docs/missing.md",
		Title:     "Missing",
		Body:      []byte("# Missing\n"),
		Metadata:  map[string]any{},
	}
	_, err := a.Publish(context.Background(), page, publisher.PublishOpts{CreateIfMissing: false})
	if err == nil {
		t.Fatal("expected error when pageId missing and CreateIfMissing=false")
	}
}

func TestAdapter_DryRunRender(t *testing.T) {
	cfg := TargetConfig{URL: "https://example.com", Space: "TEST"}
	a := New(cfg)

	xhtml, err := a.DryRunRender([]byte("# Hello\n\nParagraph.\n"))
	if err != nil {
		t.Fatalf("DryRunRender error: %v", err)
	}
	if xhtml == "" {
		t.Fatal("DryRunRender returned empty string")
	}
}

func TestAdapter_UploadImages_UpdatesExistingAttachment(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "diagram.png"), []byte("new png data"), 0o644); err != nil {
		t.Fatalf("writing image: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "new.png"), []byte("brand new"), 0o644); err != nil {
		t.Fatalf("writing image: %v", err)
	}

	var created, updated []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			// diagram.png already exists with a stale hash (e.g. uploaded before docpush)
			resp := AttachmentListResponse{Results: []AttachmentResponse{{ID: "att-1", Title: "diagram.png"}}, Size: 1}
			json.NewEncoder(w).Encode(resp)
		case strings.HasSuffix(r.URL.Path, "/data"):
			updated = append(updated, r.URL.Path)
			json.NewEncoder(w).Encode(AttachmentResponse{ID: "att-1", Title: "diagram.png"})
		default:
			created = append(created, r.URL.Path)
			json.NewEncoder(w).Encode(AttachmentListResponse{Results: []AttachmentResponse{{ID: "att-2", Title: "new.png"}}, Size: 1})
		}
	}))
	defer srv.Close()

	a := New(TargetConfig{URL: srv.URL, Space: "TEST"})
	a.client = NewClient(srv.URL, "test-pat")
	a.uploadImages(context.Background(), "123", filepath.Join(dir, "page.md"), []string{"diagram.png", "new.png"})

	if len(updated) != 1 || updated[0] != "/rest/api/content/123/child/attachment/att-1/data" {
		t.Errorf("expected existing attachment att-1 to be updated, got %v", updated)
	}
	if len(created) != 1 {
		t.Errorf("expected exactly one new attachment to be created, got %v", created)
	}
}
