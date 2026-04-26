package cmd

import (
	"fmt"
	"os"

	"github.com/colinbeales/gh-ulb/internal/api"
	"github.com/colinbeales/gh-ulb/internal/batch"
	"github.com/spf13/cobra"
)

var setTeamOrg string
var setTeamSlug string
var setTeamAmount float64
var setTeamDryRun bool
var setTeamConcurrency int

var setTeamCmd = &cobra.Command{
	Use:   "set-team",
	Short: "Set per-user budgets for all members of an org-level GitHub Team",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		members, err := api.ListOrgTeamMembers(client, setTeamOrg, setTeamSlug)
		if err != nil {
			return fmt.Errorf("listing team members: %w", err)
		}

		entries := make([]batch.UserBudgetEntry, len(members))
		for i, m := range members {
			entries[i] = batch.UserBudgetEntry{Username: m.Login, Amount: setTeamAmount}
		}

		_, err = batch.Run(cmd.Context(), client, enterprise, entries, setTeamConcurrency, setTeamDryRun, os.Stdout)
		return err
	},
}

func init() {
	setTeamCmd.Flags().StringVarP(&setTeamOrg, "org", "o", "", "Organization name (required)")
	setTeamCmd.Flags().StringVarP(&setTeamSlug, "team", "t", "", "Team slug (required)")
	setTeamCmd.Flags().Float64VarP(&setTeamAmount, "amount", "a", 0, "Budget dollar amount per user (required)")
	setTeamCmd.Flags().BoolVar(&setTeamDryRun, "dry-run", false, "Preview changes without applying them")
	setTeamCmd.Flags().IntVar(&setTeamConcurrency, "concurrency", 5, "Number of parallel API calls")
	if err := setTeamCmd.MarkFlagRequired("org"); err != nil {
		panic(err)
	}
	if err := setTeamCmd.MarkFlagRequired("team"); err != nil {
		panic(err)
	}
	if err := setTeamCmd.MarkFlagRequired("amount"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(setTeamCmd)
}
