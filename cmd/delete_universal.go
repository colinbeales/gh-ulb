package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/colinbeales/gh-ulb/internal/api"
	"github.com/spf13/cobra"
)

var deleteUniversalConfirm bool

var deleteUniversalCmd = &cobra.Command{
	Use:   "delete-universal",
	Short: "Delete the universal (multi_user_customer) budget",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !deleteUniversalConfirm {
			fmt.Print("Delete the universal budget? [y/N]: ")
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

		deletedID, err := api.DeleteUniversalBudget(client, enterprise)
		if err != nil {
			if errors.Is(err, api.ErrUniversalBudgetNotFound) {
				fmt.Println("No universal budget found. Nothing to delete.")
				return nil
			}
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Printf("Universal budget %s deleted.\n", deletedID)
		return nil
	},
}

func init() {
	deleteUniversalCmd.Flags().BoolVar(&deleteUniversalConfirm, "confirm", false, "Skip confirmation prompt")
	rootCmd.AddCommand(deleteUniversalCmd)
}
