package cmd

import (
	"fmt"
	"os"

	"github.com/colinbeales/gh-ulb/internal/api"
	"github.com/colinbeales/gh-ulb/internal/batch"
	ulbcsv "github.com/colinbeales/gh-ulb/internal/csv"
	"github.com/spf13/cobra"
)

var setCsvFile string
var setCsvAmount float64
var setCsvDryRun bool
var setCsvConcurrency int

var setCsvCmd = &cobra.Command{
	Use:   "set-csv",
	Short: "Set per-user budgets from a CSV file",
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(setCsvFile)
		if err != nil {
			return fmt.Errorf("opening CSV file: %w", err)
		}
		defer func() {
			if closeErr := f.Close(); closeErr != nil {
				_, _ = fmt.Fprintf(os.Stderr, "warning: failed to close CSV file: %v\n", closeErr)
			}
		}()

		csvEntries, err := ulbcsv.Parse(f)
		if err != nil {
			return fmt.Errorf("parsing CSV: %w", err)
		}

		var entries []batch.UserBudgetEntry
		for _, e := range csvEntries {
			a := setCsvAmount
			if e.HasAmount {
				a = e.Amount
			}
			if a == 0 {
				return fmt.Errorf("no amount for user %s: specify --amount or add an amount column to the CSV", e.Username)
			}
			entries = append(entries, batch.UserBudgetEntry{Username: e.Username, Amount: a})
		}

		client, err := api.NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		_, err = batch.Run(cmd.Context(), client, enterprise, entries, setCsvConcurrency, setCsvDryRun, os.Stdout)
		return err
	},
}

func init() {
	setCsvCmd.Flags().StringVarP(&setCsvFile, "file", "f", "", "Path to CSV file (required)")
	setCsvCmd.Flags().Float64VarP(&setCsvAmount, "amount", "a", 0, "Default budget dollar amount (used when CSV row has no amount)")
	setCsvCmd.Flags().BoolVar(&setCsvDryRun, "dry-run", false, "Preview changes without applying them")
	setCsvCmd.Flags().IntVar(&setCsvConcurrency, "concurrency", 5, "Number of parallel API calls")
	if err := setCsvCmd.MarkFlagRequired("file"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(setCsvCmd)
}
