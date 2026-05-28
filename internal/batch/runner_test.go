package batch

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/colinbeales/gh-ulb/internal/api"
)

func newTestAPIClient(t *testing.T, ts *httptest.Server) *api.Client {
	t.Helper()
	c, err := api.NewTestClient(ts.URL)
	if err != nil {
		t.Fatalf("newTestAPIClient: %v", err)
	}
	return c
}

func TestRun_DryRun(t *testing.T) {
	entries := []UserBudgetEntry{
		{Username: "octocat", Amount: 50.0},
		{Username: "monalisa", Amount: 25.0},
	}

	// No HTTP calls should be made in dry-run mode
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("unexpected HTTP call in dry-run mode")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := newTestAPIClient(t, ts)
	var buf bytes.Buffer
	result, err := Run(context.Background(), client, "my-enterprise", entries, 2, true, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(result.Results))
	}
	for _, r := range result.Results {
		if r.Action != actionDryRun {
			t.Errorf("expected action dry-run, got %v", r.Action)
		}
	}

	out := buf.String()
	if !strings.Contains(out, "[dry-run]") {
		t.Errorf("expected [dry-run] in output, got:\n%s", out)
	}
	if !strings.Contains(out, "octocat") {
		t.Errorf("expected octocat in output, got:\n%s", out)
	}
	if !strings.Contains(out, "$50.00") {
		t.Errorf("expected $50.00 in output, got:\n%s", out)
	}
}

func TestRun_AllCreated(t *testing.T) {
	t.Setenv(api.EnvUsePremiumRequests, "")
	var seenSKU string

	entries := []UserBudgetEntry{
		{Username: "octocat", Amount: 50.0},
		{Username: "monalisa", Amount: 25.0},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading request body: %v", err)
		}
		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("invalid json payload: %v", err)
		}
		if sku, ok := payload["budget_product_sku"].(string); ok {
			seenSKU = sku
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(api.Budget{ID: "b1", BudgetAmount: 50.0, BudgetScope: "user"})
	}))
	defer ts.Close()

	client := newTestAPIClient(t, ts)
	var buf bytes.Buffer
	result, err := Run(context.Background(), client, "my-enterprise", entries, 2, false, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Created != 2 {
		t.Errorf("expected 2 created, got %d", result.Created)
	}
	if result.Updated != 0 {
		t.Errorf("expected 0 updated, got %d", result.Updated)
	}
	if result.Failed != 0 {
		t.Errorf("expected 0 failed, got %d", result.Failed)
	}
	if seenSKU != api.BudgetProductSkuAICredits {
		t.Errorf("expected sku %q, got %q", api.BudgetProductSkuAICredits, seenSKU)
	}
}

func TestRun_AllCreated_PremiumOnlyViaEnv(t *testing.T) {
	t.Setenv(api.EnvUsePremiumRequests, "true")
	var seenSKU string

	entries := []UserBudgetEntry{{Username: "octocat", Amount: 50.0}}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("reading request body: %v", err)
		}
		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("invalid json payload: %v", err)
		}
		if sku, ok := payload["budget_product_sku"].(string); ok {
			seenSKU = sku
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(api.Budget{ID: "b1", BudgetAmount: 50.0, BudgetScope: "user"})
	}))
	defer ts.Close()

	client := newTestAPIClient(t, ts)
	var buf bytes.Buffer
	result, err := Run(context.Background(), client, "my-enterprise", entries, 1, false, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Created != 1 || result.Updated != 0 || result.Failed != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if seenSKU != api.BudgetProductSkuPremiumRequest {
		t.Errorf("expected sku %q, got %q", api.BudgetProductSkuPremiumRequest, seenSKU)
	}
}

func TestRun_AllUpdated(t *testing.T) {
	entries := []UserBudgetEntry{
		{Username: "octocat", Amount: 50.0},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			// Return 409 Conflict to trigger the update path
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(`{"message": "budget already exists"}`))
		case http.MethodGet:
			// Return existing budget for the list call
			json.NewEncoder(w).Encode(map[string]interface{}{
				"budgets": []api.Budget{
					{ID: "existing-b1", BudgetScope: "user", BudgetAmount: 10.0, User: "octocat"},
				},
			})
		case http.MethodPatch:
			json.NewEncoder(w).Encode(api.Budget{ID: "existing-b1", BudgetScope: "user", BudgetAmount: 50.0, User: "octocat"})
		default:
			t.Errorf("unexpected method %s", r.Method)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer ts.Close()

	client := newTestAPIClient(t, ts)
	var buf bytes.Buffer
	result, err := Run(context.Background(), client, "my-enterprise", entries, 1, false, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Updated != 1 {
		t.Errorf("expected 1 updated, got %d", result.Updated)
	}
	if result.Created != 0 {
		t.Errorf("expected 0 created, got %d", result.Created)
	}
	if result.Failed != 0 {
		t.Errorf("expected 0 failed, got %d (output: %s)", result.Failed, buf.String())
	}
}

func TestRun_Failed(t *testing.T) {
	entries := []UserBudgetEntry{
		{Username: "octocat", Amount: 50.0},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"message": "validation error"}`))
	}))
	defer ts.Close()

	client := newTestAPIClient(t, ts)
	var buf bytes.Buffer
	result, err := Run(context.Background(), client, "my-enterprise", entries, 1, false, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", result.Failed)
	}
	if result.Created != 0 {
		t.Errorf("expected 0 created, got %d", result.Created)
	}
}

func TestRun_SummaryLine(t *testing.T) {
	entries := []UserBudgetEntry{
		{Username: "octocat", Amount: 50.0},
		{Username: "monalisa", Amount: 25.0},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(api.Budget{ID: "b1", BudgetAmount: 50.0})
	}))
	defer ts.Close()

	client := newTestAPIClient(t, ts)
	var buf bytes.Buffer
	_, err := Run(context.Background(), client, "my-enterprise", entries, 2, false, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "out of 2 users") {
		t.Errorf("expected summary line with 'out of 2 users', got:\n%s", out)
	}
}

func TestRun_Concurrency(t *testing.T) {
	const numEntries = 10
	var concurrentCalls int64

	entries := make([]UserBudgetEntry, numEntries)
	for i := range entries {
		entries[i] = UserBudgetEntry{Username: "user", Amount: float64(i + 1)}
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&concurrentCalls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(api.Budget{ID: "b1", BudgetAmount: 10.0})
	}))
	defer ts.Close()

	client := newTestAPIClient(t, ts)
	var buf bytes.Buffer
	result, err := Run(context.Background(), client, "my-enterprise", entries, 3, false, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Created != numEntries {
		t.Errorf("expected %d created, got %d", numEntries, result.Created)
	}
}
