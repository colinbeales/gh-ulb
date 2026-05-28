package cmd

import (
	"fmt"
	"os"

	"github.com/colinbeales/gh-ulb/internal/api"
	"github.com/colinbeales/gh-ulb/internal/output"
	"github.com/spf13/cobra"
)

var setUserUsername string
var setUserAmount float64

var setUserCmd = &cobra.Command{
	Use:   "set-user",
	Short: "Create a per-user budget override",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		sku, err := api.ResolveBudgetProductSku()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		params := api.CreateBudgetParams{
			BudgetAmount:        setUserAmount,
			BudgetScope:         "user",
			BudgetProductSku:    sku,
			BudgetType:          "BundlePricing",
			User:                setUserUsername,
			PreventFurtherUsage: true,
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
		fmt.Printf("Budget for user %s set to $%.2f\n", setUserUsername, setUserAmount)
		return nil
	},
}

func init() {
	setUserCmd.Flags().StringVarP(&setUserUsername, "user", "u", "", "GitHub username (required)")
	setUserCmd.Flags().Float64VarP(&setUserAmount, "amount", "a", 0, "Budget dollar amount (required)")
	if err := setUserCmd.MarkFlagRequired("user"); err != nil {
		panic(err)
	}
	if err := setUserCmd.MarkFlagRequired("amount"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(setUserCmd)
}
