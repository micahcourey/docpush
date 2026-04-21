package confluence

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestClient_GetPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-pat" {
			t.Errorf("expected Bearer auth, got: %s", r.Header.Get("Authorization"))
		}
		resp := PageResponse{
			ID:    "12345",
			Title: "Test Page",
		}
		resp.Version.Number = 3
		resp.Body.Storage.Value = "<p>Hello</p>"
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-pat")
	page, err := c.GetPage(context.Background(), "12345")
	if err != nil {
		t.Fatalf("GetPage error: %v", err)
	}
	if page.ID != "12345" {
		t.Errorf("expected ID 12345, got %s", page.ID)
	}
	if page.Title != "Test Page" {
		t.Errorf("expected title Test Page, got %s", page.Title)
	}
	if page.Version.Number != 3 {
		t.Errorf("expected version 3, got %d", page.Version.Number)
	}
}

func TestClient_CreatePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		resp := PageResponse{ID: "99999", Title: "New Page"}
		resp.Version.Number = 1
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-pat")
	page, err := c.CreatePage(context.Background(), "MYSPACE", "12345", "New Page", "<p>Content</p>")
	if err != nil {
		t.Fatalf("CreatePage error: %v", err)
	}
	if page.ID != "99999" {
		t.Errorf("expected ID 99999, got %s", page.ID)
	}
}

func TestClient_UpdatePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		resp := PageResponse{ID: "12345", Title: "Updated"}
		resp.Version.Number = 4
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-pat")
	page, err := c.UpdatePage(context.Background(), "12345", "Updated", "<p>New</p>", 4)
	if err != nil {
		t.Fatalf("UpdatePage error: %v", err)
	}
	if page.Version.Number != 4 {
		t.Errorf("expected version 4, got %d", page.Version.Number)
	}
}

func TestClient_Validate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results":[]}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-pat")
	err := c.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}
}

func TestClient_Validate_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "bad-pat")
	err := c.Validate(context.Background())
	if err == nil {
		t.Fatal("expected error for unauthorized, got nil")
	}
}

func TestClient_GetAttachments(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		resp := AttachmentListResponse{
			Results: []AttachmentResponse{
				{ID: "att1", Title: "diagram.png"},
				{ID: "att2", Title: "flow.svg"},
			},
			Size: 2,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-pat")
	atts, err := c.GetAttachments(context.Background(), "12345")
	if err != nil {
		t.Fatalf("GetAttachments error: %v", err)
	}
	if len(atts) != 2 {
		t.Fatalf("expected 2 attachments, got %d", len(atts))
	}
	if atts[0].Title != "diagram.png" {
		t.Errorf("expected diagram.png, got %s", atts[0].Title)
	}
}

func TestClient_UploadAttachment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.Header.Get("X-Atlassian-Token") != "nocheck" {
			t.Errorf("expected X-Atlassian-Token: nocheck")
		}
		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "multipart/form-data") {
			t.Errorf("expected multipart/form-data content type, got %s", ct)
		}
		resp := AttachmentListResponse{
			Results: []AttachmentResponse{
				{ID: "att-new", Title: "test.png"},
			},
			Size: 1,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	// Create a temp file to upload
	tmpFile, err := os.CreateTemp("", "docpush-test-*.png")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("fake png data"))
	tmpFile.Close()

	c := NewClient(srv.URL, "test-pat")
	att, err := c.UploadAttachment(context.Background(), "12345", tmpFile.Name())
	if err != nil {
		t.Fatalf("UploadAttachment error: %v", err)
	}
	if att.ID != "att-new" {
		t.Errorf("expected att-new, got %s", att.ID)
	}
}

func TestFileHash(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "docpush-hash-test-*")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("test content"))
	tmpFile.Close()

	hash1, err := FileHash(tmpFile.Name())
	if err != nil {
		t.Fatalf("FileHash error: %v", err)
	}
	if !strings.HasPrefix(hash1, "sha256:") {
		t.Errorf("expected sha256: prefix, got %s", hash1)
	}

	// Same content should produce same hash
	hash2, err := FileHash(tmpFile.Name())
	if err != nil {
		t.Fatalf("FileHash error: %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("expected same hash, got %s and %s", hash1, hash2)
	}
}
