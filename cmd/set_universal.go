package cmd

import (
	"fmt"
	"os"

	"github.com/colinbeales/gh-ulb/internal/api"
	"github.com/colinbeales/gh-ulb/internal/output"
	"github.com/spf13/cobra"
)

var setUniversalAmount float64
var preventOverage bool

var setUniversalCmd = &cobra.Command{
	Use:   "set-universal",
	Short: "Create or update the universal (multi_user_customer) budget",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		params := api.CreateBudgetParams{
			BudgetAmount:        setUniversalAmount,
			BudgetScope:         "multi_user_customer",
			BudgetProductSku:    "premium_requests",
			BudgetType:          "BundlePricing",
			PreventFurtherUsage: preventOverage,
			BudgetAlerting:      api.BudgetAlerting{WillAlert: false, AlertRecipients: []string{}},
		}

		budget, err := api.CreateBudget(client, enterprise, params)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := output.PrintBudget(os.Stdout, budget, outputJSON); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("Universal budget set to $%.2f\n", setUniversalAmount)
		return nil
	},
}

func init() {
	setUniversalCmd.Flags().Float64VarP(&setUniversalAmount, "amount", "a", 0, "Budget dollar amount (required)")
	setUniversalCmd.Flags().BoolVar(&preventOverage, "prevent-overage", true, "Block usage when budget exhausted")
	if err := setUniversalCmd.MarkFlagRequired("amount"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(setUniversalCmd)
}
