package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/micahcourey/docpush/internal/converter"
	"github.com/micahcourey/docpush/internal/mapper"
	"github.com/micahcourey/docpush/internal/publisher"
	"github.com/micahcourey/docpush/internal/publisher/confluence"
)

func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show sync state per file",
		Long:  "Compare local file content against target page to show sync status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus()
		},
	}
	return cmd
}

func runStatus() error {
	cfg, err := mapper.LoadConfig(flagConfig)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

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

	ctx := context.Background()
	parser := converter.New()

	type statusRow struct {
		File       string `json:"file"`
		State      string `json:"state"`
		LocalHash  string `json:"localHash,omitempty"`
		RemoteHash string `json:"remoteHash,omitempty"`
		Error      string `json:"error,omitempty"`
	}
	var rows []statusRow

	for filePath := range cfg.Pages {
		body, err := os.ReadFile(filePath)
		if err != nil {
			rows = append(rows, statusRow{File: filePath, State: "error", Error: err.Error()})
			continue
		}

		doc, err := parser.Parse(body)
		if err != nil {
			rows = append(rows, statusRow{File: filePath, State: "error", Error: err.Error()})
			continue
		}

		meta := mapper.GetPageMeta(cfg, filePath, target, doc.Metadata)
		page := &publisher.Page{
			LocalPath: filePath,
			Body:      body,
			Metadata:  meta,
		}

		status, err := pub.Status(ctx, page)
		if err != nil {
			rows = append(rows, statusRow{File: filePath, State: "error", Error: err.Error()})
			continue
		}

		rows = append(rows, statusRow{
			File:       filePath,
			State:      status.State,
			LocalHash:  status.LocalHash,
			RemoteHash: status.RemoteHash,
		})
	}

	if flagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(rows)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "FILE\tSTATE\tLOCAL HASH\tREMOTE HASH\tERROR")
	for _, r := range rows {
		localShort := shortHash(r.LocalHash)
		remoteShort := shortHash(r.RemoteHash)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.File, r.State, localShort, remoteShort, r.Error)
	}
	w.Flush()
	return nil
}

func shortHash(h string) string {
	if len(h) > 12 {
		return h[:12]
	}
	return h
}
