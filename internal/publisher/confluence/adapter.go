package confluence

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/micahcourey/docpush/internal/converter"
	"github.com/micahcourey/docpush/internal/diff"
	"github.com/micahcourey/docpush/internal/publisher"
)

// Adapter implements the publisher.Publisher interface for Confluence DC.
type Adapter struct {
	client *Client
	config TargetConfig
	parser *converter.Parser
}

// New creates a new Confluence adapter from the given config.
// The Confluence URL and PAT can be overridden via environment variables.
func New(cfg TargetConfig) *Adapter {
	url := cfg.URL
	if envURL := os.Getenv("CONFLUENCE_URL"); envURL != "" {
		url = envURL
	}
	pat := os.Getenv("CONFLUENCE_PAT")

	return &Adapter{
		client: NewClient(url, pat),
		config: cfg,
		parser: converter.New(),
	}
}

// Publish converts markdown to Confluence XHTML and creates/updates the page.
func (a *Adapter) Publish(ctx context.Context, page *publisher.Page, opts publisher.PublishOpts) (*publisher.Result, error) {
	// Parse markdown
	doc, err := a.parser.Parse(page.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing markdown for %s: %w", page.LocalPath, err)
	}

	// Render to Confluence XHTML
	xhtml, err := Render(doc.Source, doc.Node)
	if err != nil {
		return nil, fmt.Errorf("rendering XHTML for %s: %w", page.LocalPath, err)
	}

	if opts.DryRun {
		return &publisher.Result{
			Action: "dry-run",
			PageID: getStringMeta(page.Metadata, "pageId"),
		}, nil
	}

	pageID := getStringMeta(page.Metadata, "pageId")
	title := page.Title
	if title == "" {
		title = getStringMeta(page.Metadata, "title")
	}

	// Prepend source banner if sourceBaseUrl is configured
	xhtml = a.prependSourceBanner(xhtml, page.LocalPath)

	// Update existing page
	if pageID != "" {
		existing, err := a.client.GetPage(ctx, pageID)
		if err != nil {
			return nil, fmt.Errorf("fetching existing page %s: %w", pageID, err)
		}

		// Skip if content hasn't changed
		if diff.Equal(existing.Body.Storage.Value, xhtml) {
			return &publisher.Result{
				Action:  "skipped",
				PageID:  pageID,
				Version: existing.Version.Number,
				URL:     existing.Links.Base + existing.Links.WebUI,
			}, nil
		}

		if title == "" {
			title = existing.Title
		}
		updated, err := a.client.UpdatePage(ctx, pageID, title, xhtml, existing.Version.Number+1)
		if err != nil {
			return nil, fmt.Errorf("updating page %s: %w", pageID, err)
		}

		// Apply labels
		a.applyLabels(ctx, pageID)

		// Apply read-only restriction
		a.applyReadOnly(ctx, pageID)

		return &publisher.Result{
			Action:  "updated",
			PageID:  updated.ID,
			Version: updated.Version.Number,
			URL:     updated.Links.Base + updated.Links.WebUI,
		}, nil
	}

	// Create new page
	if !opts.CreateIfMissing {
		return nil, fmt.Errorf("page not found for %s and --create-if-missing not set", page.LocalPath)
	}

	space := a.config.Space
	parentID := getStringMeta(page.Metadata, "parentId")
	if parentID == "" {
		parentID = a.config.Defaults.ParentID
	}

	if title == "" {
		title = page.LocalPath
	}

	created, err := a.client.CreatePage(ctx, space, parentID, title, xhtml)
	if err != nil {
		return nil, fmt.Errorf("creating page for %s: %w", page.LocalPath, err)
	}

	// Apply labels
	a.applyLabels(ctx, created.ID)

	// Apply read-only restriction
	a.applyReadOnly(ctx, created.ID)

	return &publisher.Result{
		Action:  "created",
		PageID:  created.ID,
		Version: created.Version.Number,
		URL:     created.Links.Base + created.Links.WebUI,
	}, nil
}

// Status returns the sync state of a page.
func (a *Adapter) Status(ctx context.Context, page *publisher.Page) (*publisher.SyncStatus, error) {
	pageID := getStringMeta(page.Metadata, "pageId")
	if pageID == "" {
		return &publisher.SyncStatus{State: "missing"}, nil
	}

	existing, err := a.client.GetPage(ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("fetching page %s: %w", pageID, err)
	}

	// Parse and render local content
	doc, err := a.parser.Parse(page.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing markdown: %w", err)
	}
	xhtml, err := Render(doc.Source, doc.Node)
	if err != nil {
		return nil, fmt.Errorf("rendering XHTML: %w", err)
	}

	localHash := diff.Hash(xhtml)
	remoteHash := diff.Hash(existing.Body.Storage.Value)

	state := "in-sync"
	if localHash != remoteHash {
		state = "local-ahead"
	}

	return &publisher.SyncStatus{
		State:      state,
		LocalHash:  localHash,
		RemoteHash: remoteHash,
	}, nil
}

// Validate checks the Confluence connection and credentials.
func (a *Adapter) Validate(ctx context.Context) error {
	return a.client.Validate(ctx)
}

// DryRunRender parses markdown and returns the Confluence XHTML without publishing.
func (a *Adapter) DryRunRender(body []byte) (string, error) {
	doc, err := a.parser.Parse(body)
	if err != nil {
		return "", fmt.Errorf("parsing markdown: %w", err)
	}
	return Render(doc.Source, doc.Node)
}

func (a *Adapter) applyLabels(ctx context.Context, pageID string) {
	labels := a.config.Defaults.Labels
	if len(labels) > 0 {
		_ = a.client.AddLabels(ctx, pageID, labels)
	}
}

func (a *Adapter) applyReadOnly(ctx context.Context, pageID string) {
	if a.config.Defaults.ReadOnly {
		_ = a.client.SetReadOnly(ctx, pageID)
	}
}

// prependSourceBanner adds a Confluence info panel linking to the source file in GitHub.
func (a *Adapter) prependSourceBanner(xhtml, localPath string) string {
	baseURL := a.config.Defaults.SourceBaseURL
	if baseURL == "" {
		return xhtml
	}
	sourceURL := strings.TrimRight(baseURL, "/") + "/" + localPath
	banner := `<ac:structured-macro ac:name="info"><ac:rich-text-body>` +
		`<p>&#9888; This page is auto-published by <strong>docpush</strong> — do not edit directly in Confluence.</p>` +
		`<p>Source: <a href="` + sourceURL + `">` + sourceURL + `</a></p>` +
		`</ac:rich-text-body></ac:structured-macro>`
	return banner + xhtml
}

func getStringMeta(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		// Handle numeric pageId
		return fmt.Sprintf("%v", v)
	}
	return s
}
