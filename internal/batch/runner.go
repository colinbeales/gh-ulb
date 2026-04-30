package batch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	ghapi "github.com/cli/go-gh/v2/pkg/api"

	"github.com/colinbeales/gh-ulb/internal/api"
)

type UserBudgetEntry struct {
	Username string
	Amount   float64
}

type userAction uint8

const (
	actionCreated userAction = iota + 1
	actionUpdated
	actionFailed
	actionDryRun
)

type UserResult struct {
	Err      error
	Username string
	Action   userAction
}

type BatchResult struct {
	Results []UserResult
	Created int
	Updated int
	Failed  int
}

// lockedWriter serialises concurrent writes to an underlying io.Writer.
type lockedWriter struct {
	w  io.Writer
	mu sync.Mutex
}

func (lw *lockedWriter) Write(p []byte) (n int, err error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	return lw.w.Write(p)
}

func Run(ctx context.Context, client *api.Client, enterprise string, entries []UserBudgetEntry, preventOverage bool, concurrency int, dryRun bool, w io.Writer) (BatchResult, error) {
	if dryRun {
		var result BatchResult
		for _, e := range entries {
			_, _ = fmt.Fprintf(w, "[dry-run] would set budget for %s to $%.2f\n", e.Username, e.Amount)
			result.Results = append(result.Results, UserResult{Username: e.Username, Action: actionDryRun})
		}
		return result, nil
	}

	lw := &lockedWriter{w: w}
	sem := make(chan struct{}, concurrency)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var result BatchResult

	for _, e := range entries {
		sem <- struct{}{}
		wg.Add(1)
		entry := e
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			ur := processUser(ctx, client, enterprise, entry, preventOverage, lw)
			mu.Lock()
			result.Results = append(result.Results, ur)
			switch ur.Action {
			case actionCreated:
				result.Created++
			case actionUpdated:
				result.Updated++
			case actionFailed:
				result.Failed++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()
	_, _ = fmt.Fprintf(w, "%d created, %d updated, %d failed out of %d users\n", result.Created, result.Updated, result.Failed, len(entries))
	return result, nil
}

func processUser(ctx context.Context, client *api.Client, enterprise string, entry UserBudgetEntry, preventOverage bool, w io.Writer) UserResult {
	params := api.CreateBudgetParams{
		BudgetScope:         "user",
		BudgetProductSku:    "premium_requests",
		BudgetType:          "BundlePricing",
		BudgetAmount:        entry.Amount,
		PreventFurtherUsage: preventOverage,
		User:                entry.Username,
		BudgetAlerting: api.BudgetAlerting{
			WillAlert:       false,
			AlertRecipients: []string{},
		},
	}

	const maxRetries = 3
	baseDelay := time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		_, err := api.CreateBudget(client, enterprise, params)
		if err == nil {
			_, _ = fmt.Fprintf(w, "✓ %s: budget set to $%.2f [created]\n", entry.Username, entry.Amount)
			return UserResult{Username: entry.Username, Action: actionCreated}
		}

		var httpErr *ghapi.HTTPError
		if errors.As(err, &httpErr) {
			switch httpErr.StatusCode {
			case 409:
				budgets, listErr := api.ListBudgets(client, enterprise, api.ListBudgetsOptions{
					User:         entry.Username,
					BudgetTarget: "premium_requests",
				})
				if listErr != nil {
					_, _ = fmt.Fprintf(w, "✗ %s: %v [failed]\n", entry.Username, listErr)
					return UserResult{Username: entry.Username, Action: actionFailed, Err: listErr}
				}
				var existing *api.Budget
				for i := range budgets {
					if budgets[i].BudgetScope == "user" {
						existing = &budgets[i]
						break
					}
				}
				if existing == nil {
					e2 := fmt.Errorf("conflict but no existing user-scope budget found")
					_, _ = fmt.Fprintf(w, "✗ %s: %v [failed]\n", entry.Username, e2)
					return UserResult{Username: entry.Username, Action: actionFailed, Err: e2}
				}
				_, updateErr := api.UpdateBudget(client, enterprise, existing.ID, entry.Amount, &preventOverage)
				if updateErr != nil {
					_, _ = fmt.Fprintf(w, "✗ %s: %v [failed]\n", entry.Username, updateErr)
					return UserResult{Username: entry.Username, Action: actionFailed, Err: updateErr}
				}
				_, _ = fmt.Fprintf(w, "✓ %s: budget updated to $%.2f [updated]\n", entry.Username, entry.Amount)
				return UserResult{Username: entry.Username, Action: actionUpdated}

			case 429, 500, 502, 503, 504:
				if attempt < maxRetries {
					jitter := time.Duration(rand.Intn(500)) * time.Millisecond
					delay := baseDelay*(1<<attempt) + jitter
					select {
					case <-ctx.Done():
						_, _ = fmt.Fprintf(w, "✗ %s: %v [failed]\n", entry.Username, ctx.Err())
						return UserResult{Username: entry.Username, Action: actionFailed, Err: ctx.Err()}
					case <-time.After(delay):
					}
					continue
				}
			}
		}

		_, _ = fmt.Fprintf(w, "✗ %s: %v [failed]\n", entry.Username, err)
		return UserResult{Username: entry.Username, Action: actionFailed, Err: err}
	}

	// exhausted retries
	_, _ = fmt.Fprintf(w, "✗ %s: max retries exceeded [failed]\n", entry.Username)
	return UserResult{Username: entry.Username, Action: actionFailed, Err: fmt.Errorf("max retries exceeded")}
}
