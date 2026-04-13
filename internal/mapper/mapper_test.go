package mapper

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfig(t *testing.T) {
	data := []byte(`
targets:
  confluence:
    type: confluence
    url: https://my-domain.atlassian.net
    space: MYSPACE
    defaults:
      parentId: "987654321"
      labels:
        - feature-doc
        - auto-synced

pages:
  docs/features/mass-communications.md:
    confluence:
      pageId: "123456789"
  docs/features/file-exchange.md:
    confluence:
      pageId: null
`)

	cfg, err := ParseConfig(data)
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}
	if cfg.Targets["confluence"].Space != "MYSPACE" {
		t.Errorf("expected space MYSPACE, got %s", cfg.Targets["confluence"].Space)
	}
	if cfg.Targets["confluence"].Defaults.ParentID != "987654321" {
		t.Errorf("expected parentId 987654321, got %s", cfg.Targets["confluence"].Defaults.ParentID)
	}
	if len(cfg.Targets["confluence"].Defaults.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(cfg.Targets["confluence"].Defaults.Labels))
	}

	// Check page config
	mc := cfg.Pages["docs/features/mass-communications.md"]
	if mc == nil {
		t.Fatal("expected page config for mass-communications.md")
	}
	pageID := mc["confluence"]["pageId"]
	if pageID != "123456789" {
		t.Errorf("expected pageId 123456789, got %v", pageID)
	}
}

func TestGetPageMeta_ConfigOnly(t *testing.T) {
	cfg := &Config{
		Targets: map[string]TargetEntry{
			"confluence": {
				Space: "MYSPACE",
				Defaults: DefaultsEntry{
					ParentID: "100",
				},
			},
		},
		Pages: map[string]PageOverrides{
			"docs/test.md": {
				"confluence": {"pageId": "555"},
			},
		},
	}

	meta := GetPageMeta(cfg, "docs/test.md", "confluence", nil)
	if meta["pageId"] != "555" {
		t.Errorf("expected pageId 555, got %v", meta["pageId"])
	}
	if meta["parentId"] != "100" {
		t.Errorf("expected parentId 100, got %v", meta["parentId"])
	}
}

func TestGetPageMeta_FrontmatterOverrides(t *testing.T) {
	cfg := &Config{
		Targets: map[string]TargetEntry{
			"confluence": {Space: "MYSPACE"},
		},
		Pages: map[string]PageOverrides{
			"docs/test.md": {
				"confluence": {"pageId": "555"},
			},
		},
	}

	frontmatter := map[string]interface{}{
		"docpush": map[string]interface{}{
			"confluence": map[string]interface{}{
				"pageId": "999",
			},
		},
	}

	meta := GetPageMeta(cfg, "docs/test.md", "confluence", frontmatter)
	if meta["pageId"] != "999" {
		t.Errorf("expected frontmatter pageId 999, got %v", meta["pageId"])
	}
}

func TestWritePageID(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".docpush.yaml")

	initial := []byte(`targets:
  confluence:
    type: confluence
    url: https://example.com
    space: TEST
pages: {}
`)
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatalf("writing initial config: %v", err)
	}

	err := WritePageID(configPath, "docs/new.md", "confluence", "12345")
	if err != nil {
		t.Fatalf("WritePageID error: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig after write: %v", err)
	}

	pageID := cfg.Pages["docs/new.md"]["confluence"]["pageId"]
	if pageID != "12345" {
		t.Errorf("expected pageId 12345, got %v", pageID)
	}
}
