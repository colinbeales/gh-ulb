package output

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/colinbeales/gh-ulb/internal/api"
)

// PrintBudgets writes budgets to w in either table or JSON format.
func PrintBudgets(w io.Writer, budgets []api.Budget, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(budgets)
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tSCOPE\tUSER\tAMOUNT\tPREVENT OVERAGE"); err != nil {
		return err
	}
	for _, b := range budgets {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t$%.2f\t%v\n",
			b.ID, b.BudgetScope, b.User, b.BudgetAmount, b.PreventFurtherUsage); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// PrintBudget writes a single budget.
func PrintBudget(w io.Writer, b *api.Budget, asJSON bool) error {
	return PrintBudgets(w, []api.Budget{*b}, asJSON)
}
