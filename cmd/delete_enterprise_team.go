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

var deleteEnterpriseTeamSlug string
var deleteEnterpriseTeamDryRun bool
var deleteEnterpriseTeamConfirm bool
var deleteEnterpriseTeamConcurrency int

var deleteEnterpriseTeamCmd = &cobra.Command{
	Use:   "delete-enterprise-team",
	Short: "Delete per-user budgets for all members of an Enterprise Team",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		members, err := api.ListEnterpriseTeamMembers(client, enterprise, deleteEnterpriseTeamSlug)
		if err != nil {
			return fmt.Errorf("listing enterprise team members: %w", err)
		}

		usernames := make([]string, len(members))
		for i, m := range members {
			usernames[i] = m.Login
		}

		if !deleteEnterpriseTeamConfirm && !deleteEnterpriseTeamDryRun {
			fmt.Printf("Delete budgets for %d members of enterprise team %s? [y/N]: ", len(usernames), deleteEnterpriseTeamSlug)
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

		_, err = batch.RunDelete(cmd.Context(), client, enterprise, usernames, deleteEnterpriseTeamConcurrency, deleteEnterpriseTeamDryRun, os.Stdout)
		return err
	},
}

func init() {
	deleteEnterpriseTeamCmd.Flags().StringVarP(&deleteEnterpriseTeamSlug, "team", "t", "", "Enterprise team slug (required)")
	deleteEnterpriseTeamCmd.Flags().BoolVar(&deleteEnterpriseTeamDryRun, "dry-run", false, "Preview changes without applying them")
	deleteEnterpriseTeamCmd.Flags().BoolVar(&deleteEnterpriseTeamConfirm, "confirm", false, "Skip confirmation prompt")
	deleteEnterpriseTeamCmd.Flags().IntVar(&deleteEnterpriseTeamConcurrency, "concurrency", 5, "Number of parallel API calls")
	if err := deleteEnterpriseTeamCmd.MarkFlagRequired("team"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(deleteEnterpriseTeamCmd)
}
