package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/micahcourey/docpush/internal/converter"
	"github.com/micahcourey/docpush/internal/mapper"
	"github.com/micahcourey/docpush/internal/publisher"
	"github.com/micahcourey/docpush/internal/publisher/confluence"
)

func syncCmd() *cobra.Command {
	var (
		flagDryRun          bool
		flagAll             bool
		flagCreateIfMissing bool
		flagFiles           string
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Convert and publish markdown to target",
		Long:  "Parse markdown files, convert to Confluence storage format, and publish to the configured target.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(flagDryRun, flagAll, flagCreateIfMissing, flagFiles)
		},
	}

	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "preview converted output without publishing")
	cmd.Flags().BoolVar(&flagAll, "all", false, "sync all mapped files")
	cmd.Flags().BoolVar(&flagCreateIfMissing, "create-if-missing", false, "create page if it doesn't exist")
	cmd.Flags().StringVar(&flagFiles, "files", "", "comma or newline-separated list of files to sync")

	return cmd
}

func runSync(dryRun, all, createIfMissing bool, filesFlag string) error {
	// Load config
	cfg, err := mapper.LoadConfig(flagConfig)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
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

	// Build publisher
	var pub publisher.Publisher
	switch targetCfg.Type {
	case "confluence":
		pub = confluence.New(confluence.TargetConfig{
			Type:  targetCfg.Type,
			URL:   targetCfg.URL,
			Space: targetCfg.Space,
			Defaults: confluence.DefaultsConfig{
				ParentID: targetCfg.Defaults.ParentID,
				Labels:   targetCfg.Defaults.Labels,
			},
		})
	default:
		return fmt.Errorf("unsupported target type: %s", targetCfg.Type)
	}

	// Validate connection (skip for dry-run)
	ctx := context.Background()
	if !dryRun {
		if err := pub.Validate(ctx); err != nil {
			return fmt.Errorf("validating target %s: %w", target, err)
		}
	}

	// Resolve file list
	var files []string
	if all {
		for f := range cfg.Pages {
			files = append(files, f)
		}
	} else if filesFlag != "" {
		// Support both comma-separated and newline-separated
		for _, f := range strings.FieldsFunc(filesFlag, func(r rune) bool {
			return r == ',' || r == '\n'
		}) {
			f = strings.TrimSpace(f)
			if f != "" {
				files = append(files, f)
			}
		}
	} else {
		return fmt.Errorf("specify --files or --all")
	}

	if len(files) == 0 {
		fmt.Println("No files to sync.")
		return nil
	}

	parser := converter.New()
	var results []syncResult

	for _, filePath := range files {
		body, err := os.ReadFile(filePath)
		if err != nil {
			results = append(results, syncResult{File: filePath, Error: err.Error()})
			continue
		}

		// Parse frontmatter
		doc, err := parser.Parse(body)
		if err != nil {
			results = append(results, syncResult{File: filePath, Error: err.Error()})
			continue
		}

		// Merge metadata (config + frontmatter)
		meta := mapper.GetPageMeta(cfg, filePath, target, doc.Metadata)

		// Determine title
		title := ""
		if t, ok := meta["title"]; ok {
			title = fmt.Sprintf("%v", t)
		}

		page := &publisher.Page{
			LocalPath: filePath,
			Title:     title,
			Body:      body,
			Metadata:  meta,
		}

		opts := publisher.PublishOpts{
			DryRun:          dryRun,
			CreateIfMissing: createIfMissing,
		}

		if dryRun {
			// For dry-run with confluence adapter, show the rendered XHTML
			if ca, ok := pub.(*confluence.Adapter); ok {
				xhtml, err := ca.DryRunRender(body)
				if err != nil {
					results = append(results, syncResult{File: filePath, Error: err.Error()})
					continue
				}
				results = append(results, syncResult{
					File:   filePath,
					Action: "dry-run",
					Output: xhtml,
				})
				continue
			}
		}

		result, err := pub.Publish(ctx, page, opts)
		if err != nil {
			results = append(results, syncResult{File: filePath, Error: err.Error()})
			continue
		}

		sr := syncResult{
			File:    filePath,
			Action:  result.Action,
			PageID:  result.PageID,
			URL:     result.URL,
			Version: result.Version,
		}

		// Write back pageId for newly created pages
		if result.Action == "created" && result.PageID != "" {
			if writeErr := mapper.WritePageID(flagConfig, filePath, target, result.PageID); writeErr != nil {
				sr.Warning = fmt.Sprintf("created but failed to write pageId: %v", writeErr)
			}
		}

		results = append(results, sr)
	}

	// Output results
	if flagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "FILE\tACTION\tPAGE ID\tURL\tERROR")
	for _, r := range results {
		errStr := ""
		if r.Error != "" {
			errStr = r.Error
		}
		if r.Warning != "" {
			errStr = r.Warning
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.File, r.Action, r.PageID, r.URL, errStr)
	}
	w.Flush()

	// If dry-run, also print the rendered output
	if dryRun {
		for _, r := range results {
			if r.Output != "" {
				fmt.Printf("\n--- %s ---\n%s\n", r.File, r.Output)
			}
		}
	}

	// Return error if any files failed
	for _, r := range results {
		if r.Error != "" {
			return fmt.Errorf("some files failed to sync")
		}
	}

	return nil
}

type syncResult struct {
	File    string `json:"file"`
	Action  string `json:"action,omitempty"`
	PageID  string `json:"pageId,omitempty"`
	URL     string `json:"url,omitempty"`
	Version int    `json:"version,omitempty"`
	Error   string `json:"error,omitempty"`
	Warning string `json:"warning,omitempty"`
	Output  string `json:"output,omitempty"`
}
