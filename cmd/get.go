package cmd

import (
	"fmt"
	"os"

	"github.com/colinbeales/gh-ulb/internal/api"
	"github.com/colinbeales/gh-ulb/internal/output"
	"github.com/spf13/cobra"
)

var getUser string

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the effective budget for a specific user",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		opts := api.ListBudgetsOptions{
			User: getUser,
		}

		budgets, err := api.ListBudgets(client, enterprise, opts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if len(budgets) == 0 {
			fmt.Printf("No budget found for user %s\n", getUser)
			return nil
		}

		if err := output.PrintBudgets(os.Stdout, budgets, outputJSON); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	getCmd.Flags().StringVarP(&getUser, "user", "u", "", "GitHub username (required)")
	if err := getCmd.MarkFlagRequired("user"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(getCmd)
}
