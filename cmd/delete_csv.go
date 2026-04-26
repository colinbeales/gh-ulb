package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/colinbeales/gh-ulb/internal/api"
	"github.com/colinbeales/gh-ulb/internal/batch"
	ulbcsv "github.com/colinbeales/gh-ulb/internal/csv"
	"github.com/spf13/cobra"
)

var deleteCsvFile string
var deleteCsvDryRun bool
var deleteCsvConfirm bool
var deleteCsvConcurrency int

var deleteCsvCmd = &cobra.Command{
	Use:   "delete-csv",
	Short: "Delete per-user budgets for users listed in a CSV file",
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(deleteCsvFile)
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

		usernames := make([]string, 0, len(csvEntries))
		for _, e := range csvEntries {
			usernames = append(usernames, e.Username)
		}

		if !deleteCsvConfirm && !deleteCsvDryRun {
			fmt.Printf("Delete budgets for %d users from %s? [y/N]: ", len(usernames), deleteCsvFile)
			reader := bufio.NewReader(os.Stdin)
			var answer string
			answer, err = reader.ReadString('\n')
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if strings.TrimSpace(answer) != "y" && strings.TrimSpace(answer) != "Y" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		client, err := api.NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		_, err = batch.RunDelete(cmd.Context(), client, enterprise, usernames, deleteCsvConcurrency, deleteCsvDryRun, os.Stdout)
		return err
	},
}

func init() {
	deleteCsvCmd.Flags().StringVarP(&deleteCsvFile, "file", "f", "", "Path to CSV file (required)")
	deleteCsvCmd.Flags().BoolVar(&deleteCsvDryRun, "dry-run", false, "Preview changes without applying them")
	deleteCsvCmd.Flags().BoolVar(&deleteCsvConfirm, "confirm", false, "Skip confirmation prompt")
	deleteCsvCmd.Flags().IntVar(&deleteCsvConcurrency, "concurrency", 5, "Number of parallel API calls")
	if err := deleteCsvCmd.MarkFlagRequired("file"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(deleteCsvCmd)
}
