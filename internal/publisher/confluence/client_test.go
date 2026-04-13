package confluence

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
