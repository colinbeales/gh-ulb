package batch

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/colinbeales/gh-ulb/internal/api"
)

func newDeleteTestClient(t *testing.T, ts *httptest.Server) *api.Client {
	t.Helper()
	c, err := api.NewTestClient(ts.URL)
	if err != nil {
		t.Fatalf("newDeleteTestClient: %v", err)
	}
	return c
}

func TestRunDelete_MatchedBudgetsAreDeleted(t *testing.T) {
	budgets := []api.Budget{
		{ID: "b1", BudgetScope: "user", User: "octocat"},
		{ID: "b2", BudgetScope: "user", User: "monalisa"},
	}
	var mu sync.Mutex
	var deletedIDs []string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/budgets") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"budgets":       budgets,
				"has_next_page": false,
			})
		} else if r.Method == http.MethodDelete {
			parts := strings.Split(r.URL.Path, "/")
			mu.Lock()
			deletedIDs = append(deletedIDs, parts[len(parts)-1])
			mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
		} else {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()

	client := newDeleteTestClient(t, ts)
	result, err := RunDelete(context.Background(), client, "my-enterprise", []string{"octocat", "monalisa"}, 2, false, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Deleted != 2 {
		t.Errorf("expected 2 deleted, got %d", result.Deleted)
	}
	if result.Skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", result.Skipped)
	}
	if result.Failed != 0 {
		t.Errorf("expected 0 failed, got %d", result.Failed)
	}
	if len(deletedIDs) != 2 {
		t.Errorf("expected 2 delete API calls, got %d: %v", len(deletedIDs), deletedIDs)
	}
}

func TestRunDelete_UsersWithoutBudgetsAreSkipped(t *testing.T) {
	budgets := []api.Budget{
		{ID: "b1", BudgetScope: "user", User: "octocat"},
	}
	var deletedIDs []string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/budgets") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"budgets":       budgets,
				"has_next_page": false,
			})
		} else if r.Method == http.MethodDelete {
			parts := strings.Split(r.URL.Path, "/")
			id := parts[len(parts)-1]
			deletedIDs = append(deletedIDs, id)
			if id != "b1" {
				t.Errorf("unexpected delete for budget ID %s (user without budget should be skipped)", id)
			}
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer ts.Close()

	client := newDeleteTestClient(t, ts)
	result, err := RunDelete(context.Background(), client, "my-enterprise", []string{"octocat", "no-budget-user"}, 1, false, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", result.Deleted)
	}
	if result.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", result.Skipped)
	}
	if len(deletedIDs) != 1 {
		t.Errorf("expected exactly 1 delete API call, got %d: %v", len(deletedIDs), deletedIDs)
	}
}

func TestRunDelete_DryRun(t *testing.T) {
	budgets := []api.Budget{
		{ID: "b1", BudgetScope: "user", User: "octocat"},
	}
	deleteCallCount := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/budgets") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"budgets":       budgets,
				"has_next_page": false,
			})
		} else if r.Method == http.MethodDelete {
			deleteCallCount++
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer ts.Close()

	client := newDeleteTestClient(t, ts)
	result, err := RunDelete(context.Background(), client, "my-enterprise", []string{"octocat", "no-budget-user"}, 1, true, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleteCallCount != 0 {
		t.Errorf("expected no delete API calls in dry-run, got %d", deleteCallCount)
	}
	if result.Deleted != 1 {
		t.Errorf("expected dry-run to count 1 deleted, got %d", result.Deleted)
	}
	if result.Skipped != 1 {
		t.Errorf("expected dry-run to count 1 skipped, got %d", result.Skipped)
	}
}

func TestRunDelete_OnlyDeletesUserScopedBudgets(t *testing.T) {
	budgets := []api.Budget{
		{ID: "b1", BudgetScope: "multi_user_customer"}, // universal — should not be deleted
		{ID: "b2", BudgetScope: "user", User: "octocat"},
	}
	var deletedIDs []string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/budgets") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"budgets":       budgets,
				"has_next_page": false,
			})
		} else if r.Method == http.MethodDelete {
			parts := strings.Split(r.URL.Path, "/")
			deletedIDs = append(deletedIDs, parts[len(parts)-1])
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer ts.Close()

	client := newDeleteTestClient(t, ts)
	result, err := RunDelete(context.Background(), client, "my-enterprise", []string{"octocat"}, 1, false, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", result.Deleted)
	}
	if len(deletedIDs) != 1 || deletedIDs[0] != "b2" {
		t.Errorf("expected only b2 to be deleted, got %v", deletedIDs)
	}
}
