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

type deleteAction uint8

const (
	actionDeleted deleteAction = iota + 1
	actionSkipped
	actionDeleteFailed
)

// DeleteResult holds the outcome for a single user during a bulk delete.
type DeleteResult struct {
	Err      error
	Username string
	Action   deleteAction
}

// DeleteBatchResult summarises the outcome of RunDelete.
type DeleteBatchResult struct {
	Results []DeleteResult
	Deleted int
	Skipped int
	Failed  int
}

// RunDelete bulk-deletes per-user budgets for the given usernames.
// It fetches all enterprise budgets in one paginated pass, builds a local
// username→budgetID map, then deletes concurrently with the same retry
// semantics as Run.
func RunDelete(ctx context.Context, client *api.Client, enterprise string, usernames []string, concurrency int, dryRun bool, w io.Writer) (DeleteBatchResult, error) {
	lw := &lockedWriter{w: w}

	// One paginated pass to fetch all budgets — O(ceil(B/10)) API calls.
	allBudgets, err := api.ListBudgets(client, enterprise, api.ListBudgetsOptions{})
	if err != nil {
		return DeleteBatchResult{}, fmt.Errorf("listing budgets: %w", err)
	}

	// Build username → budget ID map; only include user-scoped budgets.
	budgetByUser := make(map[string]string, len(allBudgets))
	for _, b := range allBudgets {
		if b.BudgetScope == "user" && b.User != "" {
			budgetByUser[b.User] = b.ID
		}
	}

	if dryRun {
		var result DeleteBatchResult
		for _, username := range usernames {
			if budgetID, ok := budgetByUser[username]; ok {
				_, _ = fmt.Fprintf(w, "[dry-run] would delete budget %s for %s\n", budgetID, username)
				result.Results = append(result.Results, DeleteResult{Username: username, Action: actionDeleted})
				result.Deleted++
			} else {
				_, _ = fmt.Fprintf(w, "[dry-run] no budget found for %s (skipped)\n", username)
				result.Results = append(result.Results, DeleteResult{Username: username, Action: actionSkipped})
				result.Skipped++
			}
		}
		return result, nil
	}

	sem := make(chan struct{}, concurrency)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var result DeleteBatchResult

	for _, u := range usernames {
		budgetID, ok := budgetByUser[u]
		if !ok {
			_, _ = fmt.Fprintf(lw, "- %s: no budget found (skipped)\n", u)
			mu.Lock()
			result.Results = append(result.Results, DeleteResult{Username: u, Action: actionSkipped})
			result.Skipped++
			mu.Unlock()
			continue
		}

		sem <- struct{}{}
		wg.Add(1)
		username := u
		bid := budgetID
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			dr := deleteUser(ctx, client, enterprise, username, bid, lw)
			mu.Lock()
			result.Results = append(result.Results, dr)
			switch dr.Action {
			case actionDeleted:
				result.Deleted++
			case actionDeleteFailed:
				result.Failed++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()
	_, _ = fmt.Fprintf(w, "%d deleted, %d skipped, %d failed out of %d users\n", result.Deleted, result.Skipped, result.Failed, len(usernames))
	return result, nil
}

func deleteUser(ctx context.Context, client *api.Client, enterprise, username, budgetID string, w io.Writer) DeleteResult {
	const maxRetries = 3
	baseDelay := time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := api.DeleteBudget(client, enterprise, budgetID)
		if err == nil {
			_, _ = fmt.Fprintf(w, "✓ %s: budget deleted\n", username)
			return DeleteResult{Username: username, Action: actionDeleted}
		}

		var httpErr *ghapi.HTTPError
		if errors.As(err, &httpErr) {
			switch httpErr.StatusCode {
			case 429, 500, 502, 503, 504:
				if attempt < maxRetries {
					jitter := time.Duration(rand.Intn(500)) * time.Millisecond
					delay := baseDelay*(1<<attempt) + jitter
					select {
					case <-ctx.Done():
						_, _ = fmt.Fprintf(w, "✗ %s: %v [failed]\n", username, ctx.Err())
						return DeleteResult{Username: username, Action: actionDeleteFailed, Err: ctx.Err()}
					case <-time.After(delay):
					}
					continue
				}
			}
		}

		_, _ = fmt.Fprintf(w, "✗ %s: %v [failed]\n", username, err)
		return DeleteResult{Username: username, Action: actionDeleteFailed, Err: err}
	}

	_, _ = fmt.Fprintf(w, "✗ %s: max retries exceeded [failed]\n", username)
	return DeleteResult{Username: username, Action: actionDeleteFailed, Err: fmt.Errorf("max retries exceeded")}
}
