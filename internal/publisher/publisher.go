// Package publisher defines the target-agnostic Publisher interface.
// Each publishing target (Confluence, SharePoint, etc.) implements this interface.
package publisher

import (
	"context"
	"time"
)

// Publisher is the adapter interface that each target must implement.
type Publisher interface {
	// Publish converts content and creates/updates the page on the target.
	Publish(ctx context.Context, page *Page, opts PublishOpts) (*Result, error)
	// Status returns the sync state of a page.
	Status(ctx context.Context, page *Page) (*SyncStatus, error)
	// Validate tests the connection and credentials.
	Validate(ctx context.Context) error
}

// Page represents a document to be published.
type Page struct {
	LocalPath string         // relative path to the markdown file
	Title     string         // page title (from frontmatter or config)
	Body      []byte         // raw markdown content
	Metadata  map[string]any // target-specific config (pageId, parentId, etc.)
}

// PublishOpts controls publish behavior.
type PublishOpts struct {
	DryRun          bool
	CreateIfMissing bool
}

// Result describes the outcome of a Publish call.
type Result struct {
	Action  string // "created", "updated", "skipped"
	URL     string // target page URL
	Version int    // target version number
	PageID  string // target page ID (written back to config)
}

// SyncStatus describes the sync state of a page.
type SyncStatus struct {
	State      string // "in-sync", "local-ahead", "remote-edited", "missing"
	LocalHash  string
	RemoteHash string
	LastSynced time.Time
}
