package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/micahcourey/docpush/internal/mapper"
)

func linkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link <file> <pageId>",
		Short: "Associate a file with a target page ID",
		Long:  "Write a file → pageId mapping into .docpush.yaml for the specified target.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLink(args[0], args[1])
		},
	}
	return cmd
}

func runLink(filePath, pageID string) error {
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

	if err := mapper.WritePageID(flagConfig, filePath, target, pageID); err != nil {
		return fmt.Errorf("writing page ID: %w", err)
	}

	fmt.Printf("Linked %s → %s (target: %s)\n", filePath, pageID, target)
	return nil
}
