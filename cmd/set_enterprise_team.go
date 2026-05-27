package cmd

import (
	"fmt"
	"os"

	"github.com/colinbeales/gh-ulb/internal/api"
	"github.com/colinbeales/gh-ulb/internal/batch"
	"github.com/spf13/cobra"
)

var setEnterpriseTeamSlug string
var setEnterpriseTeamAmount float64
var setEnterpriseTeamDryRun bool
var setEnterpriseTeamConcurrency int

var setEnterpriseTeamCmd = &cobra.Command{
	Use:   "set-enterprise-team",
	Short: "Set per-user budgets for all members of an Enterprise Team",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		members, err := api.ListEnterpriseTeamMembers(client, enterprise, setEnterpriseTeamSlug)
		if err != nil {
			return fmt.Errorf("listing enterprise team members: %w", err)
		}

		entries := make([]batch.UserBudgetEntry, len(members))
		for i, m := range members {
			entries[i] = batch.UserBudgetEntry{Username: m.Login, Amount: setEnterpriseTeamAmount}
		}

		_, err = batch.Run(cmd.Context(), client, enterprise, entries, setEnterpriseTeamConcurrency, setEnterpriseTeamDryRun, os.Stdout)
		return err
	},
}

func init() {
	setEnterpriseTeamCmd.Flags().StringVarP(&setEnterpriseTeamSlug, "team", "t", "", "Enterprise team slug (required)")
	setEnterpriseTeamCmd.Flags().Float64VarP(&setEnterpriseTeamAmount, "amount", "a", 0, "Budget dollar amount per user (required)")
	setEnterpriseTeamCmd.Flags().BoolVar(&setEnterpriseTeamDryRun, "dry-run", false, "Preview changes without applying them")
	setEnterpriseTeamCmd.Flags().IntVar(&setEnterpriseTeamConcurrency, "concurrency", 5, "Number of parallel API calls")
	if err := setEnterpriseTeamCmd.MarkFlagRequired("team"); err != nil {
		panic(err)
	}
	if err := setEnterpriseTeamCmd.MarkFlagRequired("amount"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(setEnterpriseTeamCmd)
}
