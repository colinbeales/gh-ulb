package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/colinbeales/gh-ulb/internal/api"
	"github.com/spf13/cobra"
)

var deleteBudgetID string
var deleteConfirm bool

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a budget by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !deleteConfirm {
			fmt.Printf("Delete budget %s? [y/N]: ", deleteBudgetID)
			reader := bufio.NewReader(os.Stdin)
			answer, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			answer = strings.TrimSpace(answer)
			if answer != "y" && answer != "Y" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := api.DeleteBudget(client, enterprise, deleteBudgetID); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Printf("Budget %s deleted.\n", deleteBudgetID)
		return nil
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&deleteBudgetID, "budget-id", "b", "", "Budget ID (required)")
	deleteCmd.Flags().BoolVar(&deleteConfirm, "confirm", false, "Skip confirmation prompt")
	if err := deleteCmd.MarkFlagRequired("budget-id"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(deleteCmd)
}
