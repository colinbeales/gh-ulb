package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/colinbeales/gh-ulb/internal/api"
	"github.com/colinbeales/gh-ulb/internal/batch"
	"github.com/spf13/cobra"
)

var deleteTeamOrg string
var deleteTeamSlug string
var deleteTeamDryRun bool
var deleteTeamConfirm bool
var deleteTeamConcurrency int

var deleteTeamCmd = &cobra.Command{
	Use:   "delete-team",
	Short: "Delete per-user budgets for all members of an org-level GitHub Team",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		members, err := api.ListOrgTeamMembers(client, deleteTeamOrg, deleteTeamSlug)
		if err != nil {
			return fmt.Errorf("listing team members: %w", err)
		}

		usernames := make([]string, len(members))
		for i, m := range members {
			usernames[i] = m.Login
		}

		if !deleteTeamConfirm && !deleteTeamDryRun {
			fmt.Printf("Delete budgets for %d members of team %s/%s? [y/N]: ", len(usernames), deleteTeamOrg, deleteTeamSlug)
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

		_, err = batch.RunDelete(cmd.Context(), client, enterprise, usernames, deleteTeamConcurrency, deleteTeamDryRun, os.Stdout)
		return err
	},
}

func init() {
	deleteTeamCmd.Flags().StringVarP(&deleteTeamOrg, "org", "o", "", "Organization name (required)")
	deleteTeamCmd.Flags().StringVarP(&deleteTeamSlug, "team", "t", "", "Team slug (required)")
	deleteTeamCmd.Flags().BoolVar(&deleteTeamDryRun, "dry-run", false, "Preview changes without applying them")
	deleteTeamCmd.Flags().BoolVar(&deleteTeamConfirm, "confirm", false, "Skip confirmation prompt")
	deleteTeamCmd.Flags().IntVar(&deleteTeamConcurrency, "concurrency", 5, "Number of parallel API calls")
	if err := deleteTeamCmd.MarkFlagRequired("org"); err != nil {
		panic(err)
	}
	if err := deleteTeamCmd.MarkFlagRequired("team"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(deleteTeamCmd)
}
