package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/micahcourey/docpush/internal/mapper"
	"github.com/micahcourey/docpush/internal/publisher/confluence"
)

func scaffoldCmd() *cobra.Command {
	var (
		flagParentID  string
		flagOutputDir string
		flagDryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "scaffold",
		Short: "Generate markdown stubs from existing Confluence pages",
		Long: `Query the Confluence REST API for child pages under a parent page,
generate markdown stub files with docpush frontmatter (pageId, title),
and add entries to .docpush.yaml. Existing entries are skipped.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScaffold(flagParentID, flagOutputDir, flagDryRun)
		},
	}

	cmd.Flags().StringVar(&flagParentID, "parent-id", "", "Confluence parent page ID to inventory (required)")
	cmd.Flags().StringVar(&flagOutputDir, "output-dir", "docs/features", "directory for generated markdown stubs")
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "show what would be created without writing files")
	_ = cmd.MarkFlagRequired("parent-id")

	return cmd
}

func runScaffold(parentID, outputDir string, dryRun bool) error {
	// Load existing config
	cfg, err := mapper.LoadConfig(flagConfig)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("loading config: %w", err)
	}
	if cfg == nil {
		return fmt.Errorf("config file %s not found — run 'docpush init' first", flagConfig)
	}

	// Resolve target
	target := flagTarget
	if target == "" {
		if len(cfg.Targets) == 1 {
			for k := range cfg.Targets {
				target = k
			}
		} else {
			return fmt.Errorf("multiple targets configured — use --target to specify one")
		}
	}

	targetCfg, ok := cfg.Targets[target]
	if !ok {
		return fmt.Errorf("target %q not found in config", target)
	}

	// Build Confluence client
	url := targetCfg.URL
	if envURL := os.Getenv("CONFLUENCE_URL"); envURL != "" {
		url = envURL
	}
	pat := os.Getenv("CONFLUENCE_PAT")
	if pat == "" {
		return fmt.Errorf("CONFLUENCE_PAT environment variable is required")
	}
	client := confluence.NewClient(url, pat)

	ctx := context.Background()

	// Validate connection
	if err := client.Validate(ctx); err != nil {
		return fmt.Errorf("confluence connection: %w", err)
	}

	// Get child pages
	fmt.Printf("Fetching child pages of %s...\n", parentID)
	pages, err := client.GetChildPages(ctx, parentID)
	if err != nil {
		return fmt.Errorf("fetching child pages: %w", err)
	}
	fmt.Printf("Found %d child pages\n\n", len(pages))

	if len(pages) == 0 {
		fmt.Println("No child pages found.")
		return nil
	}

	// Build index of existing page entries
	existingPages := make(map[string]bool)
	if cfg.Pages != nil {
		for filePath, overrides := range cfg.Pages {
			if targetOverrides, ok := overrides[target]; ok {
				if pid, ok := targetOverrides["pageId"]; ok {
					existingPages[fmt.Sprintf("%v", pid)] = true
					_ = filePath // mark file path as known
				}
			}
		}
	}

	// Process each child page
	var created, skipped int
	newEntries := make(map[string]mapper.PageOverrides)

	for _, page := range pages {
		pageID := page.ID
		title := page.Title

		// Skip if already in config
		if existingPages[pageID] {
			fmt.Printf("  skip  %-40s (pageId %s already in config)\n", title, pageID)
			skipped++
			continue
		}

		// Generate file path
		slug := titleToSlug(title)
		filePath := filepath.Join(outputDir, slug+".md")

		// Check if file already exists
		if _, err := os.Stat(filePath); err == nil {
			fmt.Printf("  skip  %-40s (file %s already exists)\n", title, filePath)
			skipped++
			continue
		}

		fmt.Printf("  create %-40s → %s (pageId: %s)\n", title, filePath, pageID)

		if !dryRun {
			// Create stub markdown file
			stub := generateStub(title, target)
			dir := filepath.Dir(filePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", dir, err)
			}
			if err := os.WriteFile(filePath, []byte(stub), 0644); err != nil {
				return fmt.Errorf("writing %s: %w", filePath, err)
			}
		}

		// Track for config update
		newEntries[filePath] = mapper.PageOverrides{
			target: map[string]interface{}{
				"pageId": pageID,
			},
		}
		created++
	}

	// Update config with new entries
	if !dryRun && len(newEntries) > 0 {
		if cfg.Pages == nil {
			cfg.Pages = make(map[string]mapper.PageOverrides)
		}
		for filePath, overrides := range newEntries {
			cfg.Pages[filePath] = overrides
		}

		if err := mapper.WriteConfig(flagConfig, cfg); err != nil {
			return fmt.Errorf("updating config: %w", err)
		}
	}

	fmt.Printf("\nDone: %d created, %d skipped\n", created, skipped)
	if dryRun && created > 0 {
		fmt.Println("(dry run — no files written)")
	}
	return nil
}

// titleToSlug converts a Confluence page title to a kebab-case filename slug.
func titleToSlug(title string) string {
	s := strings.ToLower(title)
	// Replace common separators with hyphens
	s = strings.ReplaceAll(s, " & ", "-")
	s = strings.ReplaceAll(s, " / ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "&", "-")
	// Replace spaces and underscores
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	// Remove non-alphanumeric except hyphens
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	s = reg.ReplaceAllString(s, "")
	// Collapse multiple hyphens
	reg = regexp.MustCompile(`-{2,}`)
	s = reg.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// generateStub creates a markdown stub with docpush frontmatter.
func generateStub(title, target string) string {
	return fmt.Sprintf(`---
docpush:
  %s:
    title: "%s"
---

# %s

> **TODO**: Migrate and enhance this document from Confluence.
`, target, title, title)
}
