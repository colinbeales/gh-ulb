package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var enterprise string
var outputJSON bool
var hostname string

var rootCmd = &cobra.Command{
	Use:     "ulb",
	Short:   "Manage Copilot User-Level Budgets",
	Version: "0.1.0",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if hostname != "" {
			if err := os.Setenv("GH_HOST", hostname); err != nil {
				return fmt.Errorf("setting GH_HOST: %w", err)
			}
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&enterprise, "enterprise", "e", "", "Enterprise slug (required)")
	rootCmd.PersistentFlags().BoolVar(&outputJSON, "json", false, "Output raw JSON")
	rootCmd.PersistentFlags().StringVar(&hostname, "hostname", "", "GitHub host (for GHE, e.g. <subdomain>.ghe.com)")
	if err := rootCmd.MarkPersistentFlagRequired("enterprise"); err != nil {
		panic(err)
	}
}
