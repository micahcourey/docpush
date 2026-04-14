package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"gopkg.in/yaml.v3"
)

func initCmd() *cobra.Command {
	var flagGlob string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold .docpush.yaml from docs glob",
		Long:  "Discover markdown files matching a glob pattern and create an initial .docpush.yaml config file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(flagGlob)
		},
	}

	cmd.Flags().StringVar(&flagGlob, "glob", "docs/**/*.md", "glob pattern to discover markdown files")

	return cmd
}

func runInit(glob string) error {
	if _, err := os.Stat(flagConfig); err == nil {
		return fmt.Errorf("%s already exists — remove it first or edit manually", flagConfig)
	}

	// Discover markdown files
	matches, err := filepath.Glob(glob)
	if err != nil {
		return fmt.Errorf("glob error: %w", err)
	}

	// Also try recursive pattern if no matches
	if len(matches) == 0 {
		// Try with doublestar workaround
		parts := strings.Split(glob, "**")
		if len(parts) == 2 {
			base := strings.TrimRight(parts[0], "/")
			suffix := strings.TrimLeft(parts[1], "/")
			err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if !info.IsDir() && matchesSuffix(path, suffix) {
					matches = append(matches, path)
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("walking %s: %w", base, err)
			}
		}
	}

	// Build config structure
	pages := make(map[string]interface{})
	for _, m := range matches {
		pages[m] = map[string]interface{}{
			"confluence": map[string]interface{}{
				"pageId": nil,
			},
		}
	}

	config := map[string]interface{}{
		"targets": map[string]interface{}{
			"confluence": map[string]interface{}{
				"type":  "confluence",
				"url":   "${CONFLUENCE_URL}",
				"space": "MYSPACE",
				"defaults": map[string]interface{}{
					"parentId": "",
					"labels":   []string{"feature-doc", "auto-synced"},
				},
			},
		},
		"pages": pages,
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(flagConfig, data, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", flagConfig, err)
	}

	fmt.Printf("Created %s with %d page(s)\n", flagConfig, len(matches))
	if len(matches) > 0 {
		for _, m := range matches {
			fmt.Printf("  %s\n", m)
		}
	}
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Set CONFLUENCE_URL and update the space/parentId")
	fmt.Println("  2. Run: docpush link <file> <pageId> --target confluence")
	fmt.Println("  3. Run: docpush sync --dry-run --files <file>")
	return nil
}

func matchesSuffix(path, suffix string) bool {
	if suffix == "" {
		return true
	}
	suffix = strings.TrimLeft(suffix, "/")
	if strings.HasPrefix(suffix, "*.") {
		ext := suffix[1:] // e.g., ".md"
		return strings.HasSuffix(path, ext)
	}
	return strings.HasSuffix(path, suffix)
}
