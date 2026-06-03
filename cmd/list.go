package cmd

import (
	"fmt"
	"os"

	"github.com/colinbeales/gh-ulb/internal/api"
	"github.com/colinbeales/gh-ulb/internal/output"
	"github.com/spf13/cobra"
)

var listUser string
var listBudgetTarget string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all budgets",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		budgetTarget := listBudgetTarget
		if budgetTarget == "" {
			budgetTarget, err = api.ResolveListBudgetTarget()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}

		opts := api.ListBudgetsOptions{
			User:         listUser,
			BudgetTarget: budgetTarget,
		}

		budgets, err := api.ListBudgets(client, enterprise, opts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := output.PrintBudgets(os.Stdout, budgets, outputJSON); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listUser, "user", "", "Filter to a specific user")
	listCmd.Flags().StringVar(&listBudgetTarget, "budget-target", "", "Filter by budget target (default: premium_requests; temporary API workaround)")
	rootCmd.AddCommand(listCmd)
}
