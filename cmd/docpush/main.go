package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var (
	flagTarget string
	flagConfig string
	flagJSON   bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "docpush",
		Short:   "Markdown → Confluence automation",
		Long:    "docpush converts markdown documentation to Confluence storage format and publishes it.",
		Version: version,
	}

	rootCmd.PersistentFlags().StringVar(&flagTarget, "target", "", "publish target name (default: only configured target)")
	rootCmd.PersistentFlags().StringVar(&flagConfig, "config", ".docpush.yaml", "path to config file")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "output in JSON format")

	rootCmd.AddCommand(syncCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(linkCmd())
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(scaffoldCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
