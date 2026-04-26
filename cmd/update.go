package cmd

import (
	"fmt"
	"os"

	"github.com/colinbeales/gh-ulb/internal/api"
	"github.com/colinbeales/gh-ulb/internal/output"
	"github.com/spf13/cobra"
)

var updateBudgetID string
var updateAmount float64

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing budget by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		budget, err := api.UpdateBudget(client, enterprise, updateBudgetID, updateAmount)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := output.PrintBudget(os.Stdout, budget, outputJSON); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("Budget %s updated to $%.2f\n", updateBudgetID, updateAmount)
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVarP(&updateBudgetID, "budget-id", "b", "", "Budget ID (required)")
	updateCmd.Flags().Float64VarP(&updateAmount, "amount", "a", 0, "New dollar amount (required)")
	if err := updateCmd.MarkFlagRequired("budget-id"); err != nil {
		panic(err)
	}
	if err := updateCmd.MarkFlagRequired("amount"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(updateCmd)
}
